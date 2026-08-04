package llmvision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"time"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition"
	"github.com/Nigh/autodream/automation/world"
)

// Config for vision-LLM recognition.
type Config struct {
	Name     string    `yaml:"name" json:"name"`
	ROI      frame.ROI `yaml:"roi" json:"roi"`
	Prompt   string    `yaml:"prompt" json:"prompt"`
	Endpoint string    `yaml:"endpoint" json:"endpoint"` // POST JSON {prompt,image_base64_png} → {text} or {updates}
	WorldKey string    `yaml:"world_key" json:"world_key"`
	Client   *http.Client
}

type requestBody struct {
	Prompt        string `json:"prompt"`
	ImageBase64PNG string `json:"image_base64_png"`
}

type responseBody struct {
	Text    string         `json:"text"`
	Updates []world.Update `json:"updates"`
}

// Recognizer crops ROI, PNG-encodes, and POSTs to a vision endpoint.
//
// ponytail: full ROI PNG encode per call; upgrade = shared encoder pool / native multimodal SDK.
type Recognizer struct {
	cfg    Config
	client *http.Client
}

// New builds an LLM vision recognizer.
func New(cfg Config) (*Recognizer, error) {
	if cfg.Name == "" {
		cfg.Name = "llmvision"
	}
	if cfg.Prompt == "" {
		return nil, fmt.Errorf("llmvision: empty prompt")
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("llmvision: empty endpoint")
	}
	if cfg.WorldKey == "" {
		cfg.WorldKey = "llm.text"
	}
	c := cfg.Client
	if c == nil {
		c = &http.Client{Timeout: 60 * time.Second}
	}
	return &Recognizer{cfg: cfg, client: c}, nil
}

func (r *Recognizer) Name() string { return r.cfg.Name }

func (r *Recognizer) Recognize(ctx context.Context, f *frame.Frame) (recognition.Result, error) {
	if err := ctx.Err(); err != nil {
		return recognition.Result{}, err
	}
	if f == nil || !f.Valid() {
		return recognition.Result{}, fmt.Errorf("llmvision: invalid frame")
	}
	v, err := frame.Sub(f, r.cfg.ROI)
	if err != nil {
		return recognition.Result{}, fmt.Errorf("llmvision: %w", err)
	}
	b64, err := viewPNGBase64(v)
	if err != nil {
		return recognition.Result{}, err
	}
	body, err := json.Marshal(requestBody{Prompt: r.cfg.Prompt, ImageBase64PNG: b64})
	if err != nil {
		return recognition.Result{}, fmt.Errorf("llmvision: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return recognition.Result{}, fmt.Errorf("llmvision: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := r.client.Do(req)
	if err != nil {
		return recognition.Result{}, fmt.Errorf("llmvision: post: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return recognition.Result{}, fmt.Errorf("llmvision: read: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return recognition.Result{}, fmt.Errorf("llmvision: status %d: %s", res.StatusCode, raw)
	}
	var parsed responseBody
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return recognition.Result{}, fmt.Errorf("llmvision: decode: %w", err)
	}
	updates := parsed.Updates
	if len(updates) == 0 {
		updates = []world.Update{{Key: r.cfg.WorldKey, Value: parsed.Text}}
	}
	return recognition.Result{
		FrameID:   f.FrameID,
		Timestamp: f.Timestamp,
		Updates:   updates,
		Findings:  []recognition.Finding{{Name: r.cfg.Name, Extra: map[string]any{"text": parsed.Text}}},
	}, nil
}

func viewPNGBase64(v frame.View) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, v.ROI.W, v.ROI.H))
	for y := 0; y < v.ROI.H; y++ {
		for x := 0; x < v.ROI.W; x++ {
			px, err := v.At(x, y)
			if err != nil {
				return "", err
			}
			// BGRA → RGBA
			off := (y*v.ROI.W + x) * 4
			img.Pix[off+0] = px[2]
			img.Pix[off+1] = px[1]
			img.Pix[off+2] = px[0]
			img.Pix[off+3] = px[3]
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("llmvision: png: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
