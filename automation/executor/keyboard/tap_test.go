//go:build windows || linux

package keyboard

import (
	"context"
	"testing"
	"time"
)

func TestSleepTapCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if err := sleepTap(ctx); err == nil {
		t.Fatal("expected ctx error")
	}
	if time.Since(start) >= tapHold {
		t.Fatal("cancelled sleep should not wait full hold")
	}
}

func TestTapHold(t *testing.T) {
	if tapHold != 50*time.Millisecond {
		t.Fatalf("tapHold=%v want 50ms", tapHold)
	}
}
