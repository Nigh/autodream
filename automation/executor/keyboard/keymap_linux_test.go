//go:build linux

package keyboard_test

import (
	"testing"

	"github.com/Nigh/autodream/automation/executor/keyboard"
)

func TestLookupKeycode(t *testing.T) {
	cases := map[string]byte{
		"a": 38, "A": 38, "0": 19, "enter": 36, "esc": 9, "f1": 67,
	}
	for k, want := range cases {
		got, err := keyboard.LookupKeycode(k)
		if err != nil || got != want {
			t.Fatalf("%q: got %d err=%v want %d", k, got, err, want)
		}
	}
	if _, err := keyboard.LookupKeycode(""); err == nil {
		t.Fatal("expected error")
	}
	if _, err := keyboard.LookupKeycode("not-a-key"); err == nil {
		t.Fatal("expected error")
	}
}
