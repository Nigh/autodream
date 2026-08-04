package frame

import "time"

// PixelFormat describes the layout of Frame.Data.
type PixelFormat int

const (
	PixelFormatBGRA8 PixelFormat = iota
	PixelFormatRGBA8
)

// Frame is the single shared image source for the pipeline.
// Data is read-only by contract: callers must not mutate the slice contents.
type Frame struct {
	FrameID     uint64
	Timestamp   time.Time
	Width       int
	Height      int
	PixelFormat PixelFormat
	Data        []byte
}

// BytesPerPixel returns bytes per pixel for known formats; 0 if unknown.
func (f PixelFormat) BytesPerPixel() int {
	switch f {
	case PixelFormatBGRA8, PixelFormatRGBA8:
		return 4
	default:
		return 0
	}
}

// Valid reports whether dimensions and buffer length are consistent.
func (f *Frame) Valid() bool {
	if f == nil || f.Width <= 0 || f.Height <= 0 {
		return false
	}
	bpp := f.PixelFormat.BytesPerPixel()
	if bpp == 0 {
		return false
	}
	return len(f.Data) >= f.Width*f.Height*bpp
}
