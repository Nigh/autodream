//go:build windows

package keyboard

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf16"

	"golang.org/x/sys/windows"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
)

var (
	user32         = windows.NewLazySystemDLL("user32.dll")
	procKeybdEvent = user32.NewProc("keybd_event")
)

const (
	keyeventfKeyup = 0x0002
)

// Executor sends keyboard events via user32 keybd_event.
type Executor struct{}

func New() *Executor { return &Executor{} }

func (e *Executor) Execute(ctx context.Context, a action.Action) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	vk, err := virtualKey(a.Key)
	if err != nil {
		return err
	}
	switch a.Kind {
	case action.KindKeyTap:
		procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
		procKeybdEvent.Call(uintptr(vk), 0, keyeventfKeyup, 0)
		return nil
	case action.KindKey:
		flag := uintptr(0)
		if !a.Down {
			flag = keyeventfKeyup
		}
		procKeybdEvent.Call(uintptr(vk), 0, flag, 0)
		return nil
	default:
		return fmt.Errorf("%w: %s", executor.ErrUnsupported, a.Kind)
	}
}

func virtualKey(key string) (byte, error) {
	k := strings.TrimSpace(strings.ToLower(key))
	if k == "" {
		return 0, fmt.Errorf("keyboard: empty key")
	}
	if m, ok := namedKeys[k]; ok {
		return m, nil
	}
	r := []rune(k)
	if len(r) == 1 {
		u := utf16.Encode(r)
		// Map ASCII letters/digits to VK codes (same as ASCII for A-Z 0-9)
		c := byte(u[0])
		if c >= 'a' && c <= 'z' {
			return c - 32, nil
		}
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') {
			return c, nil
		}
	}
	return 0, fmt.Errorf("keyboard: unsupported key %q", key)
}

var namedKeys = map[string]byte{
	"enter": 0x0D, "return": 0x0D, "esc": 0x1B, "escape": 0x1B,
	"tab": 0x09, "space": 0x20, "backspace": 0x08,
	"shift": 0x10, "ctrl": 0x11, "control": 0x11, "alt": 0x12,
	"left": 0x25, "up": 0x26, "right": 0x27, "down": 0x28,
	"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73,
	"f5": 0x74, "f6": 0x75, "f7": 0x76, "f8": 0x77,
	"f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
}
