package ocr

import (
	"context"
	"fmt"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition"
	"github.com/Nigh/autodream/automation/world"
)

// Engine extracts text from a frame view. Inject for tests / real backends (tesseract, cloud, …).
type Engine interface {
	ReadText(ctx context.Context, v frame.View, language string) (string, error)
}

// Config for OCR recognition.
type Config struct {
	Name     string    `yaml:"name" json:"name"`
	ROI      frame.ROI `yaml:"roi" json:"roi"`
	Language string    `yaml:"language" json:"language"`
	WorldKey string    `yaml:"world_key" json:"world_key"`
}

// Recognizer runs an injected OCR Engine over an ROI.
type Recognizer struct {
	cfg    Config
	engine Engine
}

// New requires a non-nil Engine.
func New(cfg Config, engine Engine) (*Recognizer, error) {
	if cfg.Name == "" {
		cfg.Name = "ocr"
	}
	if cfg.WorldKey == "" {
		return nil, fmt.Errorf("ocr: empty world_key")
	}
	if engine == nil {
		return nil, fmt.Errorf("ocr: nil engine")
	}
	if cfg.Language == "" {
		cfg.Language = "eng"
	}
	return &Recognizer{cfg: cfg, engine: engine}, nil
}

func (r *Recognizer) Name() string { return r.cfg.Name }

func (r *Recognizer) Recognize(ctx context.Context, f *frame.Frame) (recognition.Result, error) {
	if err := ctx.Err(); err != nil {
		return recognition.Result{}, err
	}
	if f == nil || !f.Valid() {
		return recognition.Result{}, fmt.Errorf("ocr: invalid frame")
	}
	v, err := frame.Sub(f, r.cfg.ROI)
	if err != nil {
		return recognition.Result{}, fmt.Errorf("ocr: %w", err)
	}
	text, err := r.engine.ReadText(ctx, v, r.cfg.Language)
	if err != nil {
		return recognition.Result{}, fmt.Errorf("ocr: read: %w", err)
	}
	return recognition.Result{
		FrameID:   f.FrameID,
		Timestamp: f.Timestamp,
		Updates:   []world.Update{{Key: r.cfg.WorldKey, Value: text}},
		Findings:  []recognition.Finding{{Name: r.cfg.Name, Extra: map[string]any{"text": text}}},
	}, nil
}

// StaticEngine returns fixed text (tests / stubs).
type StaticEngine struct{ Text string }

func (s StaticEngine) ReadText(context.Context, frame.View, string) (string, error) {
	return s.Text, nil
}
