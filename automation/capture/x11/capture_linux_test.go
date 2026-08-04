//go:build linux

package x11_test

import (
	"context"
	"os"
	"testing"

	"github.com/Nigh/autodream/automation/capture/x11"
)

func TestX11Capture(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set")
	}
	cap, err := x11.New(x11.Config{Screen: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := cap.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cap.Stop() }()

	f, err := cap.CurrentFrame()
	if err != nil {
		t.Fatal(err)
	}
	if !f.Valid() || f.Width < 1 || f.Height < 1 {
		t.Fatalf("bad frame %+v valid=%v", f, f.Valid())
	}
}
