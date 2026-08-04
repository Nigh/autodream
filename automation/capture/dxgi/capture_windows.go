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
	w, h, err := c.dd.GetSize()
	if err != nil {
		return nil, fmt.Errorf("dxgi: size: %w", err)
	}
	f := c.pool.Acquire(w, h, frame.PixelFormatBGRA8)
	if err := c.dd.GetFrameBGRA(f.Data, c.timeout); err != nil {
		c.pool.Release(f)
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
