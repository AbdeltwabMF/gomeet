package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/googlecal"
	"github.com/AbdeltwabMF/gomeet/internal/localcal"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

type options struct {
	background bool
	doctor     bool
	verbose    bool
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "gomeet:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	opts, help, err := parseOptions(args, stderr)
	if err != nil || help {
		return err
	}
	if opts.doctor && opts.background {
		return errors.New("--doctor and --background cannot be used together")
	}
	if opts.doctor {
		return runDoctor(stdout)
	}

	configPath, created, err := configs.EnsureConfig()
	if err != nil {
		return err
	}
	if created {
		if err := writeOutput(stdout, "Created an empty configuration at %s\n", configPath); err != nil {
			return err
		}
		if err := writeOutput(stdout, "Add meetings to the events array, then run gomeet again. Run 'gomeet --doctor' for field details and system checks.\n"); err != nil {
			return err
		}
		return nil
	}

	cfg, err := configs.LoadConfig()
	if err != nil {
		return fmt.Errorf("load %s: %w (run 'gomeet --doctor' for help)", configPath, err)
	}

	if opts.background {
		logPath, err := configs.LogPath()
		if err != nil {
			return err
		}
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("find executable: %w", err)
		}
		childArgs := []string{}
		if opts.verbose {
			childArgs = append(childArgs, "--verbose")
		}
		pid, err := platform.StartBackground(executable, childArgs)
		if err != nil {
			return err
		}
		return writeOutput(stdout, "gomeet started in the background (PID %d). Logs: %s\n", pid, logPath)
	}

	return monitor(cfg, opts.verbose, stdout, stderr)
}

func parseOptions(args []string, stderr io.Writer) (options, bool, error) {
	var opts options
	var usageErr error
	flags := flag.NewFlagSet("gomeet", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		for _, line := range []string{
			"Usage: gomeet [--background] [--verbose] [--doctor]",
			"  --background  detach and continue running in the background",
			"  --doctor      check configuration and system requirements",
			"  --verbose     include informational and debug log messages",
		} {
			if usageErr == nil {
				usageErr = writeOutput(stderr, "%s\n", line)
			}
		}
	}
	flags.BoolVar(&opts.background, "background", false, "run in the background")
	flags.BoolVar(&opts.doctor, "doctor", false, "check configuration and requirements")
	flags.BoolVar(&opts.verbose, "verbose", false, "enable verbose logging")
	flags.BoolVar(&opts.verbose, "v", false, "enable verbose logging")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, true, usageErr
		}
		return opts, false, errors.Join(err, usageErr)
	}
	if flags.NArg() != 0 {
		return opts, false, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

	return opts, false, nil
}

func writeOutput(w io.Writer, format string, args ...any) error {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func monitor(cfg *configs.Config, verbose bool, stdout, stderr io.Writer) (retErr error) {
	f, err := configs.OpenLog()
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close log: %w", err))
		}
	}()
	configs.InitLogger(io.MultiWriter(f, stderr), verbose)

	if err := writeOutput(stdout, "gomeet is running in the foreground; press Ctrl+C to stop.\n"); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error { return localcal.Monitor(ctx, cfg) })

	if configs.HasCredentials() {
		// The Google monitor runs under its own supervision, so that a
		// failure there does not stop the local calendar.
		go monitorGoogle(ctx, cfg)
	} else {
		slog.Warn("Google Calendar credentials not found; only the local calendar is monitored")
	}

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	slog.Info("shutdown complete")
	return nil
}

// monitorGoogle runs the Google Calendar monitor and retries it when it
// fails, so that a temporary problem does not disable Google reminders.
func monitorGoogle(ctx context.Context, cfg *configs.Config) {
	const (
		minDelay = 15 * time.Second
		maxDelay = 5 * time.Minute
	)

	delay := minDelay
	for {
		err := googlecal.Monitor(ctx, cfg)
		if ctx.Err() != nil || err == nil {
			return
		}

		slog.Error("google calendar monitor failed, retrying", slog.Any("error", err), slog.Duration("retry_in", delay))
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		if delay *= 2; delay > maxDelay {
			delay = maxDelay
		}
	}
}
