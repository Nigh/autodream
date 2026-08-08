package template_test

import (
	"context"
	"testing"
	"time"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition/template"
)

func TestTemplateMatch(t *testing.T) {
	w, h := 8, 8
	data := make([]byte, w*h*4)
	// paint a 2x2 red block at (3,3)
	for y := 3; y < 5; y++ {
		for x := 3; x < 5; x++ {
			i := (y*w + x) * 4
			data[i+2] = 255
			data[i+3] = 255
		}
	}
	f := &frame.Frame{
		FrameID: 1, Timestamp: time.Now().UTC(),
		Width: w, Height: h, PixelFormat: frame.PixelFormatBGRA8, Data: data,
	}
	tmpl := []byte{
		0, 0, 255, 255, 0, 0, 255, 255,
		0, 0, 255, 255, 0, 0, 255, 255,
	}
	rec, err := template.NewFromBGRA(template.Config{WorldKey: "hit", Threshold: 0.95}, 2, 2, tmpl)
	if err != nil {
		t.Fatal(err)
	}
	res, err := rec.Recognize(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}
	if res.Updates[0].Value != true {
		t.Fatalf("%+v", res)
	}
	if res.Findings[0].Extra["x"] != 3 || res.Findings[0].Extra["y"] != 3 {
		t.Fatalf("pos %+v", res.Findings[0].Extra)
	}
}

func TestTemplateSkipTransparent(t *testing.T) {
	w, h := 4, 4
	data := make([]byte, w*h*4)
	for i := 0; i < len(data); i += 4 {
		data[i] = 255 // blue background
		data[i+3] = 255
	}
	// red pixel at (2,2)
	i := (2*w + 2) * 4
	data[i+0], data[i+1], data[i+2], data[i+3] = 0, 0, 255, 255

	// 3x3: only center fully opaque red; neighbors black with alpha 0 / 128
	// (would spoil SAD against blue if semi-transparent were counted)
	tmpl := make([]byte, 3*3*4)
	ci := (1*3 + 1) * 4
	tmpl[ci+2], tmpl[ci+3] = 255, 255
	si := (1*3 + 0) * 4
	tmpl[si+3] = 128 // semi-transparent black

	rec, err := template.NewFromBGRA(template.Config{WorldKey: "hit", Threshold: 0.95}, 3, 3, tmpl)
	if err != nil {
		t.Fatal(err)
	}
	res, err := rec.Recognize(context.Background(), &frame.Frame{
		FrameID: 1, Timestamp: time.Now().UTC(),
		Width: w, Height: h, PixelFormat: frame.PixelFormatBGRA8, Data: data,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updates[0].Value != true {
		t.Fatalf("want hit, got %+v", res)
	}
	if res.Findings[0].Extra["x"] != 1 || res.Findings[0].Extra["y"] != 1 {
		t.Fatalf("pos %+v", res.Findings[0].Extra)
	}
}
