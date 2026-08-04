package frame

import "fmt"

// ROI is a rectangle in pixel coordinates (origin top-left).
type ROI struct {
	X, Y, W, H int
}

// View shares Frame.Data without copying the full image.
type View struct {
	Frame *Frame
	ROI   ROI
}

// Sub returns a view into f. Empty ROI means the full frame.
func Sub(f *Frame, roi ROI) (View, error) {
	if f == nil {
		return View{}, fmt.Errorf("frame: nil frame")
	}
	if roi.W == 0 && roi.H == 0 && roi.X == 0 && roi.Y == 0 {
		roi = ROI{X: 0, Y: 0, W: f.Width, H: f.Height}
	}
	if roi.X < 0 || roi.Y < 0 || roi.W <= 0 || roi.H <= 0 {
		return View{}, fmt.Errorf("frame: invalid ROI %+v", roi)
	}
	if roi.X+roi.W > f.Width || roi.Y+roi.H > f.Height {
		return View{}, fmt.Errorf("frame: ROI %+v out of bounds %dx%d", roi, f.Width, f.Height)
	}
	return View{Frame: f, ROI: roi}, nil
}

// At returns a pointer into the shared buffer for pixel (x,y) within the view
// (view-local coordinates). The returned slice length is BytesPerPixel.
// Callers must treat it as read-only.
func (v View) At(x, y int) ([]byte, error) {
	if v.Frame == nil {
		return nil, fmt.Errorf("frame: nil view frame")
	}
	if x < 0 || y < 0 || x >= v.ROI.W || y >= v.ROI.H {
		return nil, fmt.Errorf("frame: pixel (%d,%d) outside view %dx%d", x, y, v.ROI.W, v.ROI.H)
	}
	bpp := v.Frame.PixelFormat.BytesPerPixel()
	if bpp == 0 {
		return nil, fmt.Errorf("frame: unknown pixel format %d", v.Frame.PixelFormat)
	}
	absX := v.ROI.X + x
	absY := v.ROI.Y + y
	off := (absY*v.Frame.Width + absX) * bpp
	end := off + bpp
	if end > len(v.Frame.Data) {
		return nil, fmt.Errorf("frame: buffer underrun at (%d,%d)", absX, absY)
	}
	return v.Frame.Data[off:end:end], nil
}
