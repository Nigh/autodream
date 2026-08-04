package recognition

import (
	"context"
	"time"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/world"
)

// Finding is optional debug/metadata from a recognizer.
type Finding struct {
	Name  string
	Score float64
	Extra map[string]any
}

// Result is the only output of recognition. Runtime applies Updates to World.
type Result struct {
	FrameID   uint64
	Timestamp time.Time
	Updates   []world.Update
	Findings  []Finding
}

// Recognizer analyzes a shared Frame. Must not capture or execute input.
type Recognizer interface {
	Name() string
	Recognize(ctx context.Context, f *frame.Frame) (Result, error)
}
