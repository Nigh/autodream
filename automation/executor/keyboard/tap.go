//go:build windows || linux

package keyboard

import (
	"context"
	"time"
)

// Hold between key-down and key-up for KindKeyTap so targets register the key.
const tapHold = 50 * time.Millisecond

func sleepTap(ctx context.Context) error {
	t := time.NewTimer(tapHold)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
