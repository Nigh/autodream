package template

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/internal/imgutil"
	"github.com/Nigh/autodream/automation/recognition"
	"github.com/Nigh/autodream/automation/world"
)

// Config for template matching inside an ROI.
type Config struct {
	Name       string    `yaml:"name" json:"name"`
	ROI        frame.ROI `yaml:"roi" json:"roi"`
	Image      string    `yaml:"image" json:"image"` // path to template PNG/JPEG
	Threshold  float64   `yaml:"threshold" json:"threshold"` // min score 0..1; default 0.9
	WorldKey   string    `yaml:"world_key" json:"world_key"`
	WorldValue any       `yaml:"world_value" json:"world_value"`
}

// Recognizer finds the best SAD-based match of a template inside ROI.
type Recognizer struct {
	cfg    Config
	tmplW  int
	tmplH  int
	tmpl   []byte // BGRA
}

// New loads the template image from cfg.Image.
func New(cfg Config) (*Recognizer, error) {
	if cfg.Name == "" {
		cfg.Name = "template"
	}
	if cfg.WorldKey == "" {
		return nil, fmt.Errorf("template: empty world_key")
	}
	if cfg.Image == "" {
		return nil, fmt.Errorf("template: empty image path")
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 0.9
	}
	if cfg.WorldValue == nil {
		cfg.WorldValue = true
	}
	f, err := os.Open(cfg.Image)
	if err != nil {
		return nil, fmt.Errorf("template: open: %w", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("template: decode: %w", err)
	}
	w, h, data := imgutil.ImageToBGRA(img)
	return &Recognizer{cfg: cfg, tmplW: w, tmplH: h, tmpl: data}, nil
}

// NewFromBGRA is useful in tests without touching disk.
func NewFromBGRA(cfg Config, w, h int, bgra []byte) (*Recognizer, error) {
	if cfg.Name == "" {
		cfg.Name = "template"
	}
	if cfg.WorldKey == "" {
		return nil, fmt.Errorf("template: empty world_key")
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 0.9
	}
	if cfg.WorldValue == nil {
		cfg.WorldValue = true
	}
	if w <= 0 || h <= 0 || len(bgra) < w*h*4 {
		return nil, fmt.Errorf("template: bad template buffer")
	}
	cp := make([]byte, w*h*4)
	copy(cp, bgra)
	return &Recognizer{cfg: cfg, tmplW: w, tmplH: h, tmpl: cp}, nil
}

func (r *Recognizer) Name() string { return r.cfg.Name }

func (r *Recognizer) Recognize(ctx context.Context, f *frame.Frame) (recognition.Result, error) {
	if err := ctx.Err(); err != nil {
		return recognition.Result{}, err
	}
	if f == nil || !f.Valid() {
		return recognition.Result{}, fmt.Errorf("template: invalid frame")
	}
	v, err := frame.Sub(f, r.cfg.ROI)
	if err != nil {
		return recognition.Result{}, fmt.Errorf("template: %w", err)
	}
	if r.tmplW > v.ROI.W || r.tmplH > v.ROI.H {
		return recognition.Result{}, fmt.Errorf("template: template larger than ROI")
	}

	best := 0.0
	bestX, bestY := 0, 0
	maxDiff := float64(r.tmplW * r.tmplH * 3 * 255)
	for y := 0; y <= v.ROI.H-r.tmplH; y++ {
		for x := 0; x <= v.ROI.W-r.tmplW; x++ {
			if err := ctx.Err(); err != nil {
				return recognition.Result{}, err
			}
			var sad float64
			ti := 0
			for ty := 0; ty < r.tmplH; ty++ {
				for tx := 0; tx < r.tmplW; tx++ {
					px, err := v.At(x+tx, y+ty)
					if err != nil {
						return recognition.Result{}, err
					}
					sad += absDiff(px[0], r.tmpl[ti]) + absDiff(px[1], r.tmpl[ti+1]) + absDiff(px[2], r.tmpl[ti+2])
					ti += 4
				}
			}
			score := 1 - sad/maxDiff
			if score > best {
				best = score
				bestX, bestY = x, y
			}
		}
	}

	out := recognition.Result{
		FrameID:   f.FrameID,
		Timestamp: f.Timestamp,
		Findings: []recognition.Finding{{
			Name:  r.cfg.Name,
			Score: best,
			Extra: map[string]any{"x": bestX + v.ROI.X, "y": bestY + v.ROI.Y},
		}},
	}
	if best >= r.cfg.Threshold {
		out.Updates = []world.Update{{Key: r.cfg.WorldKey, Value: r.cfg.WorldValue}}
	} else {
		out.Updates = []world.Update{{Key: r.cfg.WorldKey, Value: false}}
	}
	return out, nil
}

func absDiff(a, b byte) float64 {
	if a > b {
		return float64(a - b)
	}
	return float64(b - a)
}
