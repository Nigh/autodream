package dxgi

import (
	"math"
	"strings"
)

func validSize(w, h int) bool { return w > 0 && h > 0 }

// pickSize chooses dims when GetSize may return 0 (first call / DPI switch).
// Prefer a fresh positive GetSize; else last good cache; else Win32 fallback.
func pickSize(getW, getH, cachedW, cachedH, fallbackW, fallbackH int) (int, int, bool) {
	if validSize(getW, getH) {
		return getW, getH, true
	}
	if validSize(cachedW, cachedH) {
		return cachedW, cachedH, true
	}
	if validSize(fallbackW, fallbackH) {
		return fallbackW, fallbackH, true
	}
	return 0, 0, false
}

// logicalToPhysical converts DesktopCoordinates size to DDA texture pixels.
// GetSize/GetBounds follow process DPI awareness (logical); GetFrameBGRA needs physical.
func logicalToPhysical(w, h int, dpi uint32) (int, int) {
	if !validSize(w, h) || dpi <= 96 {
		return w, h
	}
	scale := float64(dpi) / 96
	return int(math.Round(float64(w) * scale)), int(math.Round(float64(h) * scale))
}

func isBufferTooSmall(err error) bool {
	return err != nil && strings.Contains(err.Error(), "buffer too small")
}
