//go:build windows

package gamepad

import (
	"context"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
)

var (
	xinput            = windows.NewLazySystemDLL("xinput1_4.dll")
	procXInputSetState = xinput.NewProc("XInputSetState")
)

type xinputVibration struct {
	LeftMotorSpeed  uint16
	RightMotorSpeed uint16
}

// Executor drives XInput vibration / digital buttons via a simple state buffer.
//
// ponytail: vibration-focused MVP; full button injection needs a virtual pad driver later.
type Executor struct {
	index uint32
}

func New(index int) *Executor {
	if index < 0 {
		index = 0
	}
	return &Executor{index: uint32(index)}
}

func (e *Executor) Execute(ctx context.Context, a action.Action) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.Kind != action.KindGamepad {
		return fmt.Errorf("%w: %s", executor.ErrUnsupported, a.Kind)
	}
	idx := e.index
	if a.GamepadIndex > 0 {
		idx = uint32(a.GamepadIndex)
	}
	ctrl := strings.ToLower(a.Control)
	switch ctrl {
	case "vibrate", "rumble":
		left := uint16(clamp01(a.Value) * 65535)
		right := left
		vib := xinputVibration{LeftMotorSpeed: left, RightMotorSpeed: right}
		r, _, callErr := procXInputSetState.Call(uintptr(idx), uintptr(unsafe.Pointer(&vib)))
		if r != 0 {
			return fmt.Errorf("gamepad: XInputSetState: %w", callErr)
		}
		return nil
	default:
		return fmt.Errorf("gamepad: unsupported control %q (MVP supports vibrate)", a.Control)
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
