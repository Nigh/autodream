//go:build linux

package mouse

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgb/xtest"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
)

// Executor drives the pointer via the XTEST extension.
type Executor struct {
	once sync.Once
	init error
	conn *xgb.Conn
	root xproto.Window
}

// New returns a Linux XTest mouse executor (connects lazily on first Execute).
func New() *Executor { return &Executor{} }

func (e *Executor) ensure() error {
	e.once.Do(func() {
		display := os.Getenv("DISPLAY")
		if display == "" {
			e.init = fmt.Errorf("mouse: DISPLAY not set")
			return
		}
		conn, err := xgb.NewConnDisplay(display)
		if err != nil {
			e.init = fmt.Errorf("mouse: connect: %w", err)
			return
		}
		if err := xtest.Init(conn); err != nil {
			conn.Close()
			e.init = fmt.Errorf("mouse: XTEST: %w", err)
			return
		}
		setup := xproto.Setup(conn)
		if len(setup.Roots) == 0 {
			conn.Close()
			e.init = fmt.Errorf("mouse: no X screens")
			return
		}
		e.conn = conn
		e.root = setup.Roots[0].Root
	})
	return e.init
}

func (e *Executor) Execute(ctx context.Context, a action.Action) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := e.ensure(); err != nil {
		return err
	}
	switch a.Kind {
	case action.KindMouseMove:
		return e.motion(a.X, a.Y)
	case action.KindMouseClick:
		if err := e.motion(a.X, a.Y); err != nil {
			return err
		}
		btn, err := buttonDetail(a.Button)
		if err != nil {
			return err
		}
		if err := e.button(btn, true); err != nil {
			return err
		}
		return e.button(btn, false)
	case action.KindMouseButton:
		btn, err := buttonDetail(a.Button)
		if err != nil {
			return err
		}
		return e.button(btn, a.Down)
	default:
		return fmt.Errorf("%w: %s", executor.ErrUnsupported, a.Kind)
	}
}

func (e *Executor) motion(x, y int) error {
	// detail=0 → absolute coordinates relative to root (XTEST).
	err := xtest.FakeInputChecked(e.conn, xproto.MotionNotify, 0, 0, e.root, int16(x), int16(y), 0).Check()
	if err != nil {
		return fmt.Errorf("mouse: motion: %w", err)
	}
	return nil
}

func (e *Executor) button(detail byte, down bool) error {
	typ := byte(xproto.ButtonRelease)
	if down {
		typ = xproto.ButtonPress
	}
	err := xtest.FakeInputChecked(e.conn, typ, detail, 0, e.root, 0, 0, 0).Check()
	if err != nil {
		return fmt.Errorf("mouse: button: %w", err)
	}
	return nil
}

func buttonDetail(btn string) (byte, error) {
	switch strings.ToLower(btn) {
	case "", "left":
		return 1, nil
	case "middle":
		return 2, nil
	case "right":
		return 3, nil
	default:
		return 0, fmt.Errorf("mouse: unknown button %q", btn)
	}
}
