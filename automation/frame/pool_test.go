package frame_test

import (
	"testing"

	"github.com/Nigh/autodream/automation/frame"
)

func TestPoolReuse(t *testing.T) {
	var p frame.Pool
	f1 := p.Acquire(8, 8, frame.PixelFormatBGRA8)
	ptr := &f1.Data[0]
	p.Release(f1)
	f2 := p.Acquire(8, 8, frame.PixelFormatBGRA8)
	if &f2.Data[0] != ptr {
		t.Fatal("expected buffer reuse")
	}
	if f2.FrameID == 0 || !f2.Valid() {
		t.Fatalf("bad frame %+v", f2)
	}
}
