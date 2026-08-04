package llmvision_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"context"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition/llmvision"
)

func TestLLMVisionHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"text": "ok"})
	}))
	defer srv.Close()

	data := make([]byte, 4*4*4)
	for i := 3; i < len(data); i += 4 {
		data[i] = 255
	}
	f := &frame.Frame{
		FrameID: 1, Timestamp: time.Now().UTC(),
		Width: 4, Height: 4, PixelFormat: frame.PixelFormatBGRA8, Data: data,
	}
	rec, err := llmvision.New(llmvision.Config{
		Prompt: "describe", Endpoint: srv.URL, WorldKey: "llm.text",
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rec.Recognize(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}
	if res.Updates[0].Value != "ok" {
		t.Fatalf("%+v", res.Updates)
	}
}
