//go:build windows

package mouse

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sys/windows"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
)

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	procSetCursorPos  = user32.NewProc("SetCursorPos")
	procMouseEvent    = user32.NewProc("mouse_event")
)

const (
	mouseeventfLeftDown   = 0x0002
	mouseeventfLeftUp     = 0x0004
	mouseeventfRightDown  = 0x0008
	mouseeventfRightUp    = 0x0010
	mouseeventfMiddleDown = 0x0020
	mouseeventfMiddleUp   = 0x0040
)

// Executor drives the system cursor via user32.
type Executor struct{}

func New() *Executor { return &Executor{} }

func (e *Executor) Execute(ctx context.Context, a action.Action) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	switch a.Kind {
	case action.KindMouseMove:
		r, _, err := procSetCursorPos.Call(uintptr(a.X), uintptr(a.Y))
		if r == 0 {
			return fmt.Errorf("mouse: SetCursorPos: %w", err)
		}
		return nil
	case action.KindMouseClick:
		if err := e.move(a.X, a.Y); err != nil {
			return err
		}
		down, up, err := buttonFlags(a.Button)
		if err != nil {
			return err
		}
		procMouseEvent.Call(uintptr(down), 0, 0, 0, 0)
		procMouseEvent.Call(uintptr(up), 0, 0, 0, 0)
		return nil
	case action.KindMouseButton:
		down, up, err := buttonFlags(a.Button)
		if err != nil {
			return err
		}
		flag := up
		if a.Down {
			flag = down
		}
		procMouseEvent.Call(uintptr(flag), 0, 0, 0, 0)
		return nil
	default:
		return fmt.Errorf("%w: %s", executor.ErrUnsupported, a.Kind)
	}
}

func (e *Executor) move(x, y int) error {
	r, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if r == 0 {
		return fmt.Errorf("mouse: SetCursorPos: %w", err)
	}
	return nil
}

func buttonFlags(btn string) (down, up uintptr, err error) {
	switch strings.ToLower(btn) {
	case "", "left":
		return mouseeventfLeftDown, mouseeventfLeftUp, nil
	case "right":
		return mouseeventfRightDown, mouseeventfRightUp, nil
	case "middle":
		return mouseeventfMiddleDown, mouseeventfMiddleUp, nil
	default:
		return 0, 0, fmt.Errorf("mouse: unknown button %q", btn)
	}
}
