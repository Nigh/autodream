package dxgihttp_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nigh/autodream/automation/capture/dxgihttp"
	"github.com/Nigh/autodream/automation/frame"
)

func TestHTTPCaptureFull(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{B: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	pngBytes := buf.Bytes()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/capture/full" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	}))
	defer srv.Close()

	cap, err := dxgihttp.New(dxgihttp.Config{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err := cap.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	f, err := cap.CurrentFrame()
	if err != nil {
		t.Fatal(err)
	}
	if !f.Valid() || f.Width != 2 || f.Height != 1 || f.PixelFormat != frame.PixelFormatBGRA8 {
		t.Fatalf("bad frame meta: %+v", f)
	}
	// pixel0 was red → BGRA B=0 G=0 R=255
	if f.Data[2] != 255 || f.Data[0] != 0 {
		t.Fatalf("pixel0 BGRA=%v", f.Data[:4])
	}
	_ = cap.Stop()
}
