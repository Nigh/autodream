package imgutil

import "image"

// RGBAToBGRA copies src into dst (must be len >= len(src)). Channels swapped in place layout.
func RGBAToBGRA(dst, src []byte) {
	n := len(src) / 4 * 4
	if len(dst) < n {
		n = len(dst) / 4 * 4
	}
	for i := 0; i < n; i += 4 {
		dst[i+0] = src[i+2] // B
		dst[i+1] = src[i+1] // G
		dst[i+2] = src[i+0] // R
		dst[i+3] = src[i+3] // A
	}
}

// ImageToBGRA packs an image.Image into tightly packed BGRA bytes (row-major).
func ImageToBGRA(img image.Image) (w, h int, data []byte) {
	b := img.Bounds()
	w, h = b.Dx(), b.Dy()
	data = make([]byte, w*h*4)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			data[i+0] = byte(bl >> 8)
			data[i+1] = byte(g >> 8)
			data[i+2] = byte(r >> 8)
			data[i+3] = byte(a >> 8)
			i += 4
		}
	}
	return w, h, data
}
