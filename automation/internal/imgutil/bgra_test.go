package imgutil_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/Nigh/autodream/automation/internal/imgutil"
)

func TestImageToBGRA(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 1, G: 2, B: 3, A: 4})
	w, h, data := imgutil.ImageToBGRA(img)
	if w != 1 || h != 1 || data[0] != 3 || data[1] != 2 || data[2] != 1 || data[3] != 4 {
		t.Fatalf("%d %d %v", w, h, data)
	}
}
