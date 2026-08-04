package ocr_test

import (
	"context"
	"testing"
	"time"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition/ocr"
)

func TestOCRStatic(t *testing.T) {
	data := make([]byte, 4)
	data[3] = 255
	f := &frame.Frame{
		FrameID: 1, Timestamp: time.Now().UTC(),
		Width: 1, Height: 1, PixelFormat: frame.PixelFormatBGRA8, Data: data,
	}
	rec, err := ocr.New(ocr.Config{WorldKey: "text"}, ocr.StaticEngine{Text: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rec.Recognize(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}
	if res.Updates[0].Value != "hello" {
		t.Fatalf("%+v", res.Updates)
	}
}
