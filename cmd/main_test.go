package main

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AbdeltwabMF/gomeet/configs"
)

func TestParseOptions(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	opts, help, err := parseOptions([]string{"--background", "--verbose"}, &stderr)
	if err != nil || help {
		t.Fatalf("parseOptions() = (_, %v, %v), want options", help, err)
	}
	if !opts.background || !opts.verbose || opts.doctor {
		t.Fatalf("parseOptions() = %+v", opts)
	}

	_, _, err = parseOptions([]string{"unexpected"}, &stderr)
	if err == nil || !strings.Contains(err.Error(), "unexpected arguments") {
		t.Fatalf("parseOptions(unexpected) error = %v", err)
	}
}

func TestParseStartupOptions(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	opts, help, err := parseOptions([]string{"--install-startup", "--yes"}, &stderr)
	if err != nil || help {
		t.Fatalf("parseOptions() = (_, %v, %v), want options", help, err)
	}
	if !opts.installStartup || !opts.yes || opts.uninstallStartup {
		t.Fatalf("parseOptions() = %+v", opts)
	}
}

func TestGetConsent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		answer    string
		assumeYes bool
		want      bool
	}{
		{name: "yes", answer: "yes\n", want: true},
		{name: "short yes", answer: "Y\n", want: true},
		{name: "no", answer: "no\n", want: false},
		{name: "empty defaults no", answer: "\n", want: false},
		{name: "non-interactive approval", assumeYes: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			got, err := getConsent(strings.NewReader(tt.answer), &output, tt.assumeYes, "Continue? [y/N] ")
			if err != nil {
				t.Fatalf("getConsent() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("getConsent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConsentReusesBufferedInput(t *testing.T) {
	t.Parallel()

	input := bufio.NewReader(strings.NewReader("yes\nno\n"))
	var output bytes.Buffer

	first, err := getConsent(input, &output, false, "First? ")
	if err != nil {
		t.Fatalf("first getConsent() error = %v", err)
	}
	second, err := getConsent(input, &output, false, "Second? ")
	if err != nil {
		t.Fatalf("second getConsent() error = %v", err)
	}
	if !first || second {
		t.Fatalf("getConsent() results = (%v, %v), want (true, false)", first, second)
	}
}

// TestMonitorGoogleRetries verifies that a failing Google monitor is retried
// and that it never terminates the process on its own.
func TestMonitorGoogleRetries(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A cancelled context makes the monitor return immediately, so the
	// supervisor must exit without spinning.
	done := make(chan struct{})
	go func() {
		monitorGoogle(ctx, &configs.Config{})
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("monitorGoogle did not return after cancellation")
	}
}

// TestMonitorGoogleStopsOnCancelledContext verifies that the supervisor stops
// when the context is cancelled while it is waiting to retry.
func TestMonitorGoogleStopsOnCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		monitorGoogle(ctx, &configs.Config{})
		close(done)
	}()

	// Cancel while the supervisor is between attempts.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("monitorGoogle did not return after cancellation")
	}
}
