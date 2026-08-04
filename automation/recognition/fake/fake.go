package fake

import (
	"context"
	"time"

	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/recognition"
	"github.com/Nigh/autodream/automation/world"
)

// Recognizer returns a fixed Result (or from fn) for tests.
type Recognizer struct {
	Label string
	Fn    func(context.Context, *frame.Frame) (recognition.Result, error)
	Fixed recognition.Result
}

func (r *Recognizer) Name() string {
	if r.Label != "" {
		return r.Label
	}
	return "fake"
}

func (r *Recognizer) Recognize(ctx context.Context, f *frame.Frame) (recognition.Result, error) {
	if err := ctx.Err(); err != nil {
		return recognition.Result{}, err
	}
	if r.Fn != nil {
		return r.Fn(ctx, f)
	}
	out := r.Fixed
	if f != nil {
		out.FrameID = f.FrameID
		if out.Timestamp.IsZero() {
			out.Timestamp = f.Timestamp
		}
	}
	if out.Timestamp.IsZero() {
		out.Timestamp = time.Now().UTC()
	}
	return out, nil
}

// StaticUpdates returns a recognizer that always emits the given updates.
func StaticUpdates(name string, updates ...world.Update) *Recognizer {
	return &Recognizer{
		Label: name,
		Fixed: recognition.Result{Updates: updates},
	}
}
