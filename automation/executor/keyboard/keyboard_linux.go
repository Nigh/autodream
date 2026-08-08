//go:build linux

package keyboard

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgb/xtest"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
)

// Executor sends key events via the XTEST extension.
type Executor struct {
	once sync.Once
	init error
	conn *xgb.Conn
	root xproto.Window
}

// New returns a Linux XTest keyboard executor (connects lazily on first Execute).
func New() *Executor { return &Executor{} }

func (e *Executor) ensure() error {
	e.once.Do(func() {
		display := os.Getenv("DISPLAY")
		if display == "" {
			e.init = fmt.Errorf("keyboard: DISPLAY not set")
			return
		}
		conn, err := xgb.NewConnDisplay(display)
		if err != nil {
			e.init = fmt.Errorf("keyboard: connect: %w", err)
			return
		}
		if err := xtest.Init(conn); err != nil {
			conn.Close()
			e.init = fmt.Errorf("keyboard: XTEST: %w", err)
			return
		}
		setup := xproto.Setup(conn)
		if len(setup.Roots) == 0 {
			conn.Close()
			e.init = fmt.Errorf("keyboard: no X screens")
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
	code, err := LookupKeycode(a.Key)
	if err != nil {
		return err
	}
	switch a.Kind {
	case action.KindKeyTap:
		if err := e.key(code, true); err != nil {
			return err
		}
		err := sleepTap(ctx)
		if upErr := e.key(code, false); upErr != nil {
			return upErr
		}
		return err

	case action.KindKey:
		return e.key(code, a.Down)
	default:
		return fmt.Errorf("%w: %s", executor.ErrUnsupported, a.Kind)
	}
}

func (e *Executor) key(code byte, down bool) error {
	typ := byte(xproto.KeyRelease)
	if down {
		typ = xproto.KeyPress
	}
	err := xtest.FakeInputChecked(e.conn, typ, code, 0, e.root, 0, 0, 0).Check()
	if err != nil {
		return fmt.Errorf("keyboard: key: %w", err)
	}
	return nil
}
