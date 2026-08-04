//go:build linux

package x11

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"

	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/log"
)

// Config for X11 root-window capture.
//
// ponytail: XGetImage copies over the X socket each frame; upgrade = MIT-SHM.
type Config struct {
	// Display is the X display string (e.g. ":0"). Empty uses $DISPLAY.
	Display string
	// Screen is the screen index in the X setup (0 = default).
	Screen int
	Logger log.Logger
}

// Capture implements capture.Capture via xproto.GetImage on the root window.
type Capture struct {
	cfg     Config
	log     log.Logger
	pool    frame.Pool
	started atomic.Bool

	mu      sync.Mutex
	conn    *xgb.Conn
	root    xproto.Window
	width   int
	height  int
	current *frame.Frame
}

// New builds an X11 capture (not connected until Start).
func New(cfg Config) (*Capture, error) {
	if cfg.Screen < 0 {
		return nil, fmt.Errorf("x11: negative screen index")
	}
	lg := cfg.Logger
	if lg == nil {
		lg = log.Nop()
	}
	return &Capture{cfg: cfg, log: lg}, nil
}

func (c *Capture) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	display := c.cfg.Display
	if display == "" {
		display = os.Getenv("DISPLAY")
	}
	if display == "" {
		return fmt.Errorf("x11: DISPLAY not set")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.started.Store(true)
		return nil
	}

	conn, err := xgb.NewConnDisplay(display)
	if err != nil {
		return fmt.Errorf("x11: connect %q: %w", display, err)
	}
	setup := xproto.Setup(conn)
	if c.cfg.Screen >= len(setup.Roots) {
		conn.Close()
		return fmt.Errorf("x11: screen %d out of range (have %d)", c.cfg.Screen, len(setup.Roots))
	}
	screen := setup.Roots[c.cfg.Screen]
	c.conn = conn
	c.root = screen.Root
	c.width = int(screen.WidthInPixels)
	c.height = int(screen.HeightInPixels)
	c.started.Store(true)
	c.log.Info("x11 capture started", "display", display, "screen", c.cfg.Screen, "w", c.width, "h", c.height)
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
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.log.Info("x11 capture stopped")
	return nil
}

func (c *Capture) CurrentFrame() (*frame.Frame, error) {
	if !c.started.Load() {
		return nil, capture.ErrNotStarted
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil, capture.ErrNotStarted
	}

	reply, err := xproto.GetImage(
		c.conn,
		xproto.ImageFormatZPixmap,
		xproto.Drawable(c.root),
		0, 0,
		uint16(c.width), uint16(c.height),
		0xffffffff,
	).Reply()
	if err != nil {
		return nil, fmt.Errorf("x11: GetImage: %w", err)
	}
	need := c.width * c.height * 4
	if len(reply.Data) < need {
		return nil, fmt.Errorf("x11: short image data %d < %d", len(reply.Data), need)
	}

	f := c.pool.Acquire(c.width, c.height, frame.PixelFormatBGRA8)
	// X11 ZPixmap for 24/32-bit TrueColor is typically BGRA/BGRx in native byte order on little-endian.
	copy(f.Data, reply.Data[:need])
	old := c.current
	c.current = f
	if old != nil {
		c.pool.Release(old)
	}
	c.log.Trace("x11 frame", "id", f.FrameID, "w", c.width, "h", c.height)
	return f, nil
}
