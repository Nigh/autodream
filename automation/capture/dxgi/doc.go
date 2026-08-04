//go:build !windows

package dxgi

// Native DXGI Desktop Duplication is Windows-only (see capture_windows.go).
// On Linux/macOS use capture/fake or capture/dxgihttp for development and tests.
