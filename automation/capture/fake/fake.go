package fake

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/frame"
)

// Capture serves a fixed or sequenced set of frames for tests.
type Capture struct {
	mu      sync.Mutex
	frames  []*frame.Frame
	idx     int
	started atomic.Bool
	loop    bool
}

// New returns a fake capture. If loop is true, CurrentFrame cycles the slice.
func New(frames []*frame.Frame, loop bool) *Capture {
	cp := make([]*frame.Frame, len(frames))
	copy(cp, frames)
	return &Capture{frames: cp, loop: loop}
}

// NewSolid builds a single BGRA frame filled with one color.
func NewSolid(w, h int, b, g, r, a byte) *Capture {
	data := make([]byte, w*h*4)
	for i := 0; i < len(data); i += 4 {
		data[i] = b
		data[i+1] = g
		data[i+2] = r
		data[i+3] = a
	}
	f := &frame.Frame{
		FrameID:     1,
		Timestamp:   time.Now().UTC(),
		Width:       w,
		Height:      h,
		PixelFormat: frame.PixelFormatBGRA8,
		Data:        data,
	}
	return New([]*frame.Frame{f}, true)
}

func (c *Capture) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.started.Store(true)
	return nil
}

func (c *Capture) Stop() error {
	c.started.Store(false)
	return nil
}

func (c *Capture) CurrentFrame() (*frame.Frame, error) {
	if !c.started.Load() {
		return nil, capture.ErrNotStarted
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.frames) == 0 {
		return nil, capture.ErrStopped
	}
	f := c.frames[c.idx]
	if c.loop {
		c.idx = (c.idx + 1) % len(c.frames)
	} else if c.idx < len(c.frames)-1 {
		c.idx++
	}
	return f, nil
}
