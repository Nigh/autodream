package dxgi

import (
	"errors"
	"fmt"
	"testing"
)

func TestLogicalToPhysical(t *testing.T) {
	w, h := logicalToPhysical(1920, 1080, 96)
	if w != 1920 || h != 1080 {
		t.Fatalf("96dpi: got %dx%d", w, h)
	}
	w, h = logicalToPhysical(1280, 720, 144) // 150%
	if w != 1920 || h != 1080 {
		t.Fatalf("144dpi: got %dx%d want 1920x1080", w, h)
	}
	w, h = logicalToPhysical(1280, 720, 192) // 200%
	if w != 2560 || h != 1440 {
		t.Fatalf("192dpi: got %dx%d want 2560x1440", w, h)
	}
}

func TestIsBufferTooSmall(t *testing.T) {
	if !isBufferTooSmall(errors.New("buffer too small")) {
		t.Fatal("plain")
	}
	if !isBufferTooSmall(fmt.Errorf("dxgi: frame: %w", errors.New("buffer too small"))) {
		t.Fatal("wrapped")
	}
	if isBufferTooSmall(errors.New("no image yet")) {
		t.Fatal("other error")
	}
}

func TestPickSize(t *testing.T) {
	w, h, ok := pickSize(1920, 1080, 0, 0, 0, 0)
	if !ok || w != 1920 || h != 1080 {
		t.Fatalf("fresh GetSize: %dx%d ok=%v", w, h, ok)
	}
	// GetSize 0 → keep cache
	w, h, ok = pickSize(0, 0, 2560, 1440, 800, 600)
	if !ok || w != 2560 || h != 1440 {
		t.Fatalf("cache: %dx%d ok=%v", w, h, ok)
	}
	// GetSize 0, no cache → Win32 fallback
	w, h, ok = pickSize(0, 0, 0, 0, 1920, 1080)
	if !ok || w != 1920 || h != 1080 {
		t.Fatalf("fallback: %dx%d ok=%v", w, h, ok)
	}
	_, _, ok = pickSize(0, 0, 0, 0, 0, 0)
	if ok {
		t.Fatal("expected unavailable")
	}
}
