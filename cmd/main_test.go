package main

import (
	"context"
	"testing"
	"time"

	"github.com/AbdeltwabMF/gomeet/configs"
)

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
