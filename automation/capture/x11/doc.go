//go:build !linux

package x11

// X11 root-window capture is Linux-only (see capture_linux.go).
// On other platforms use capture/fake, capture/dxgihttp, or capture/dxgi.
