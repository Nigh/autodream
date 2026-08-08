//go:build windows

package dxgi

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shinkar94/godesktopdup"

	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/log"
)

// Config for native DXGI Desktop Duplication.
type Config struct {
	OutputIndex   uint
	Timeout       time.Duration // AcquireFrame timeout; default 100ms
	CaptureCursor bool
	Logger        log.Logger
}

// Capture implements capture.Capture via DXGI Desktop Duplication.
type Capture struct {
	cfg     Config
	log     log.Logger
	timeout uint
	pool    frame.Pool

	started atomic.Bool
	mu      sync.Mutex
	dd      *dda.DesktopDuplication
	current *frame.Frame
	// ponytail: pix* = buffer size for GetFrameBGRA; logic* tracks last GetSize to detect DPI/mode change.
	pixW, pixH     int
	logicW, logicH int
}

// New creates a DXGI capture for the given output index.
func New(cfg Config) (*Capture, error) {
	lg := cfg.Logger
	if lg == nil {
		lg = log.Nop()
	}
	to := cfg.Timeout
	if to <= 0 {
		to = 100 * time.Millisecond
	}
	return &Capture{
		cfg:     cfg,
		log:     lg,
		timeout: uint(to.Milliseconds()),
	}, nil
}

func (c *Capture) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dd != nil {
		c.started.Store(true)
		return nil
	}
	dd, err := dda.New(c.cfg.OutputIndex)
	if err != nil {
		return fmt.Errorf("dxgi: open output %d: %w", c.cfg.OutputIndex, err)
	}
	dd.SetCaptureCursor(c.cfg.CaptureCursor)
	c.dd = dd
	c.started.Store(true)
	c.log.Info("dxgi started", "output", c.cfg.OutputIndex)
	return nil
}

func (c *Capture) Stop() error {
	c.started.Store(false)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current != nil {
		c.pool.Release(c.current)
		c.current = nil
	}
	if c.dd != nil {
		c.dd.Release()
		c.dd = nil
	}
	c.pixW, c.pixH = 0, 0
	c.logicW, c.logicH = 0, 0
	return nil
}

func (c *Capture) CurrentFrame() (*frame.Frame, error) {
	if !c.started.Load() {
		return nil, capture.ErrNotStarted
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dd == nil {
		return nil, capture.ErrNotStarted
	}
	w, h, err := c.framePixels()
	if err != nil {
		return nil, err
	}
	f, err := c.grab(w, h)
	if isBufferTooSmall(err) {
		// GetSize was logical; GetFrameBGRA needs physical. Retry once.
		if w2, h2, ok := c.physicalFromDPI(w, h); ok {
			f, err = c.grab(w2, h2)
			if err == nil {
				w, h = w2, h2
				c.pixW, c.pixH = w2, h2
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("dxgi: frame: %w", err)
	}
	old := c.current
	c.current = f
	if old != nil {
		c.pool.Release(old)
	}
	c.log.Trace("dxgi frame", "id", f.FrameID, "w", w, "h", h)
	return f, nil
}

func (c *Capture) framePixels() (int, int, error) {
	w, h, err := c.dd.GetSize()
	if err != nil {
		return 0, 0, fmt.Errorf("dxgi: size: %w", err)
	}
	if validSize(w, h) {
		if c.logicW != w || c.logicH != h {
			// DPI / mode change: drop physical override; may bump again on buffer-too-small.
			c.logicW, c.logicH = w, h
			c.pixW, c.pixH = w, h
		} else if !validSize(c.pixW, c.pixH) {
			c.pixW, c.pixH = w, h
		}
		return c.pixW, c.pixH, nil
	}
	// GetSize can return 0 on first call or during DPI switch.
	fbW, fbH := 0, 0
	if fw, fh, ok := displaySize(c.cfg.OutputIndex); ok {
		fbW, fbH = fw, fh
	}
	outW, outH, ok := pickSize(w, h, c.pixW, c.pixH, fbW, fbH)
	if !ok {
		return 0, 0, fmt.Errorf("dxgi: size unavailable (GetSize %dx%d)", w, h)
	}
	c.pixW, c.pixH = outW, outH
	return outW, outH, nil
}

func (c *Capture) physicalFromDPI(logicalW, logicalH int) (int, int, bool) {
	left, top, right, bottom, err := c.dd.GetBounds()
	if err != nil {
		return 0, 0, false
	}
	w, h := logicalToPhysical(logicalW, logicalH, monitorDPI(left, top, right, bottom))
	if w == logicalW && h == logicalH {
		return 0, 0, false
	}
	return w, h, true
}

func (c *Capture) grab(w, h int) (*frame.Frame, error) {
	if !validSize(w, h) {
		return nil, fmt.Errorf("invalid size %dx%d", w, h)
	}
	f := c.pool.Acquire(w, h, frame.PixelFormatBGRA8)
	if err := c.dd.GetFrameBGRA(f.Data, c.timeout); err != nil {
		c.pool.Release(f)
		return nil, err
	}
	return f, nil
}
