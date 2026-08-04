package color

import (
	"context"
	"fmt"
	"math"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition"
	"github.com/Nigh/autodream/automation/world"
)

// Config is YAML/JSON friendly color match settings.
type Config struct {
	Name      string    `yaml:"name" json:"name"`
	ROI       frame.ROI `yaml:"roi" json:"roi"`
	B byte `yaml:"b" json:"b"`
	G byte `yaml:"g" json:"g"`
	R byte `yaml:"r" json:"r"`
	A byte `yaml:"a" json:"a"` // reserved; unused in distance
	Threshold  float64 `yaml:"threshold" json:"threshold"` // max mean L1 distance 0..255; default 32
	WorldKey   string  `yaml:"world_key" json:"world_key"`
	WorldValue any     `yaml:"world_value" json:"world_value"`
}

// Recognizer matches average ROI color against a target.
type Recognizer struct {
	cfg Config
}

// New builds a ColorRecognizer.
func New(cfg Config) (*Recognizer, error) {
	if cfg.Name == "" {
		cfg.Name = "color"
	}
	if cfg.WorldKey == "" {
		return nil, fmt.Errorf("color: empty world_key")
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 32
	}
	if cfg.WorldValue == nil {
		cfg.WorldValue = true
	}
	return &Recognizer{cfg: cfg}, nil
}

func (r *Recognizer) Name() string { return r.cfg.Name }

func (r *Recognizer) Recognize(ctx context.Context, f *frame.Frame) (recognition.Result, error) {
	if err := ctx.Err(); err != nil {
		return recognition.Result{}, err
	}
	if f == nil || !f.Valid() {
		return recognition.Result{}, fmt.Errorf("color: invalid frame")
	}
	v, err := frame.Sub(f, r.cfg.ROI)
	if err != nil {
		return recognition.Result{}, fmt.Errorf("color: %w", err)
	}
	var sumB, sumG, sumR float64
	n := v.ROI.W * v.ROI.H
	for y := 0; y < v.ROI.H; y++ {
		for x := 0; x < v.ROI.W; x++ {
			px, err := v.At(x, y)
			if err != nil {
				return recognition.Result{}, err
			}
			sumB += float64(px[0])
			sumG += float64(px[1])
			sumR += float64(px[2])
		}
	}
	nf := float64(n)
	meanB, meanG, meanR := sumB/nf, sumG/nf, sumR/nf
	dist := (math.Abs(meanB-float64(r.cfg.B)) + math.Abs(meanG-float64(r.cfg.G)) + math.Abs(meanR-float64(r.cfg.R))) / 3
	out := recognition.Result{
		FrameID:   f.FrameID,
		Timestamp: f.Timestamp,
		Findings: []recognition.Finding{{
			Name:  r.cfg.Name,
			Score: 1 - dist/255,
			Extra: map[string]any{"distance": dist},
		}},
	}
	if dist <= r.cfg.Threshold {
		out.Updates = []world.Update{{Key: r.cfg.WorldKey, Value: r.cfg.WorldValue}}
	} else {
		out.Updates = []world.Update{{Key: r.cfg.WorldKey, Value: false}}
	}
	return out, nil
}
