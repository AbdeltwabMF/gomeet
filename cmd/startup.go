package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/googlecal"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

func setupStartup(stdin io.Reader, stdout io.Writer, assumeYes, verbose bool) error {
	if !platform.SupportsStartup() {
		return errors.New("automatic startup installation is currently supported only on Windows")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	stdin = bufio.NewReader(stdin)

	if configs.HasCredentials() {
		if err := googlecal.CheckCredentials(); err != nil {
			return fmt.Errorf("google credentials are invalid: %w (run 'gomeet --doctor' for help)", err)
		}
		if configs.HasToken() {
			if err := googlecal.CheckToken(); err != nil {
				return fmt.Errorf("google authorization token is invalid: %w (delete %s and try again)", err, configs.TokenFile)
			}
		} else {
			approved, err := getConsent(stdin, stdout, assumeYes, "Google Calendar needs browser authorization. Authorize now? [y/N] ")
			if err != nil {
				return err
			}
			if !approved {
				return writeOutput(stdout, "Startup installation canceled; Google authorization was not completed.\n")
			}
			if err := googlecal.Authorize(ctx); err != nil {
				return fmt.Errorf("authorize Google Calendar: %w", err)
			}
			if err := writeOutput(stdout, "Google Calendar authorization completed.\n"); err != nil {
				return err
			}
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}
	arguments := "none"
	if verbose {
		arguments = "--verbose"
	}
	if err := writeOutput(stdout, `Startup plan:
  Task:       GoMeet
  Trigger:    Current user signs in (30-second delay)
  Process:    %s
  Arguments:  %s
  Window:     Hidden
  Recovery:   Restart up to 3 times; prevent duplicate instances
`, executable, arguments); err != nil {
		return err
	}

	approved, err := getConsent(stdin, stdout, assumeYes, "Install this scheduled task? [y/N] ")
	if err != nil {
		return err
	}
	if !approved {
		return writeOutput(stdout, "Startup installation canceled.\n")
	}
	if err := platform.InstallStartup(ctx, executable, verbose); err != nil {
		return err
	}

	return writeOutput(stdout, "GoMeet will start automatically after you sign in.\n")
}

func removeStartup(stdin io.Reader, stdout io.Writer, assumeYes bool) error {
	if !platform.SupportsStartup() {
		return errors.New("automatic startup removal is currently supported only on Windows")
	}
	approved, err := getConsent(bufio.NewReader(stdin), stdout, assumeYes, "Remove the GoMeet scheduled task? [y/N] ")
	if err != nil {
		return err
	}
	if !approved {
		return writeOutput(stdout, "Startup removal canceled.\n")
	}
	if err := platform.UninstallStartup(context.Background()); err != nil {
		return err
	}

	return writeOutput(stdout, "GoMeet automatic startup was removed.\n")
}

func getConsent(stdin io.Reader, stdout io.Writer, assumeYes bool, prompt string) (bool, error) {
	if assumeYes {
		return true, nil
	}
	if err := writeOutput(stdout, "%s", prompt); err != nil {
		return false, err
	}

	reader, ok := stdin.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(stdin)
	}
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer = strings.ToLower(strings.TrimSpace(answer))

	return answer == "y" || answer == "yes", nil
}
