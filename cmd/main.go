package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/googlecal"
	"github.com/AbdeltwabMF/gomeet/internal/localcal"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	f, err := configs.OpenLog()
	if err != nil {
		return err
	}
	defer f.Close()
	configs.InitLogger(f)

	cfg, err := configs.LoadConfig()
	if err != nil {
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
		slog.Warn("Google Calendar credentials not found, only the local calendar is monitored")
	}

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	slog.Info("Shutdown complete")

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
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			return
		}

		slog.Error("google calendar monitor failed, retrying", slog.Any("error", err), slog.Duration("retry_in", delay))

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		if delay *= 2; delay > maxDelay {
			delay = maxDelay
		}
	}
}
