//go:build linux

package keyboard

import (
	"fmt"
	"strings"
)

// LookupKeycode maps a key name to an X11 keycode (US/QWERTY-oriented MVP).
func LookupKeycode(key string) (byte, error) {
	k := strings.TrimSpace(strings.ToLower(key))
	if k == "" {
		return 0, fmt.Errorf("keyboard: empty key")
	}
	if code, ok := namedKeycodes[k]; ok {
		return code, nil
	}
	if len(k) == 1 {
		c := k[0]
		if c >= 'a' && c <= 'z' {
			return letterKeycodes[c-'a'], nil
		}
		if c >= '0' && c <= '9' {
			return digitKeycodes[c-'0'], nil
		}
	}
	return 0, fmt.Errorf("keyboard: unsupported key %q", key)
}

// X11 keycodes for a typical evdev/pc105 layout (not keysyms).
var letterKeycodes = [26]byte{
	// a b c d e f g h i j k l m n o p q r s t u v w x y z
	38, 56, 54, 40, 26, 41, 42, 43, 31, 44, 45, 46, 58,
	57, 32, 33, 24, 27, 39, 28, 30, 55, 25, 53, 29, 52,
}

var digitKeycodes = [10]byte{
	// 0 1 2 3 4 5 6 7 8 9
	19, 10, 11, 12, 13, 14, 15, 16, 17, 18,
}

var namedKeycodes = map[string]byte{
	"enter": 36, "return": 36, "esc": 9, "escape": 9,
	"tab": 23, "space": 65, "backspace": 22,
	"shift": 50, "ctrl": 37, "control": 37, "alt": 64,
	"left": 113, "up": 111, "right": 114, "down": 116,
	"f1": 67, "f2": 68, "f3": 69, "f4": 70,
	"f5": 71, "f6": 72, "f7": 73, "f8": 74,
	"f9": 75, "f10": 76, "f11": 95, "f12": 96,
}
