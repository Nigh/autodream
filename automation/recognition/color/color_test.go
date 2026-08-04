package color_test

import (
	"context"
	"testing"
	"time"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition/color"
)

func solid(w, h int, b, g, r, a byte) *frame.Frame {
	data := make([]byte, w*h*4)
	for i := 0; i < len(data); i += 4 {
		data[i], data[i+1], data[i+2], data[i+3] = b, g, r, a
	}
	return &frame.Frame{
		FrameID: 1, Timestamp: time.Now().UTC(),
		Width: w, Height: h, PixelFormat: frame.PixelFormatBGRA8, Data: data,
	}
}

func TestColorMatch(t *testing.T) {
	rec, err := color.New(color.Config{
		B: 10, G: 20, R: 30, Threshold: 5, WorldKey: "match",
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rec.Recognize(context.Background(), solid(4, 4, 10, 20, 30, 255))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Updates) != 1 || res.Updates[0].Value != true {
		t.Fatalf("%+v", res.Updates)
	}
	res, err = rec.Recognize(context.Background(), solid(4, 4, 200, 200, 200, 255))
	if err != nil {
		t.Fatal(err)
	}
	if res.Updates[0].Value != false {
		t.Fatalf("want false got %+v", res.Updates)
	}
}
