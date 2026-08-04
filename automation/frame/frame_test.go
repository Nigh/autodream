package frame_test

import (
	"testing"
	"time"

	"github.com/Nigh/autodream/automation/frame"
)

func TestFrameValidAndViewAt(t *testing.T) {
	w, h := 2, 2
	data := []byte{
		1, 0, 0, 255, 2, 0, 0, 255,
		3, 0, 0, 255, 4, 0, 0, 255,
	}
	f := &frame.Frame{
		FrameID:     1,
		Timestamp:   time.Unix(0, 0).UTC(),
		Width:       w,
		Height:      h,
		PixelFormat: frame.PixelFormatBGRA8,
		Data:        data,
	}
	if !f.Valid() {
		t.Fatal("expected valid frame")
	}
	v, err := frame.Sub(f, frame.ROI{X: 1, Y: 1, W: 1, H: 1})
	if err != nil {
		t.Fatal(err)
	}
	px, err := v.At(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if px[0] != 4 {
		t.Fatalf("got B=%d want 4", px[0])
	}
}
