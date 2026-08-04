package fake_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/capture/fake"
)

func TestFakeCapture(t *testing.T) {
	c := fake.NewSolid(4, 4, 10, 20, 30, 255)
	if _, err := c.CurrentFrame(); !errors.Is(err, capture.ErrNotStarted) {
		t.Fatalf("want ErrNotStarted got %v", err)
	}
	if err := c.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	f, err := c.CurrentFrame()
	if err != nil {
		t.Fatal(err)
	}
	if !f.Valid() || f.Width != 4 || f.Data[0] != 10 {
		t.Fatalf("bad frame: %+v", f)
	}
	_ = c.Stop()
}
