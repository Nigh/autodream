package pipeline

import (
	"context"
	"time"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
	"github.com/Nigh/autodream/automation/world"
)

// Condition succeeds when Pred returns true.
type Condition struct {
	Pred func(world.World) (bool, error)
}

func (c *Condition) Tick(ctx context.Context, w world.World, _ executor.Executor) (Status, error) {
	if err := ctx.Err(); err != nil {
		return StatusFailure, err
	}
	if c.Pred == nil {
		return StatusFailure, nil
	}
	ok, err := c.Pred(w)
	if err != nil {
		return StatusFailure, err
	}
	if ok {
		return StatusSuccess, nil
	}
	return StatusFailure, nil
}

// Wait succeeds after Duration has elapsed across ticks (Running until then).
type Wait struct {
	Duration time.Duration
	start    time.Time
}

func (w *Wait) Tick(ctx context.Context, _ world.World, _ executor.Executor) (Status, error) {
	if err := ctx.Err(); err != nil {
		return StatusFailure, err
	}
	now := time.Now()
	if w.start.IsZero() {
		w.start = now
	}
	if now.Sub(w.start) >= w.Duration {
		w.start = time.Time{}
		return StatusSuccess, nil
	}
	return StatusRunning, nil
}

// Repeat runs Child up to Times (0 = forever until child Failure or ctx done).
type Repeat struct {
	Child Node
	Times int
	n     int
}

func (r *Repeat) Tick(ctx context.Context, w world.World, ex executor.Executor) (Status, error) {
	if r.Child == nil {
		return StatusFailure, nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return StatusFailure, err
		}
		if r.Times > 0 && r.n >= r.Times {
			r.n = 0
			return StatusSuccess, nil
		}
		st, err := r.Child.Tick(ctx, w, ex)
		if err != nil {
			return StatusFailure, err
		}
		switch st {
		case StatusRunning:
			return StatusRunning, nil
		case StatusFailure:
			r.n = 0
			return StatusFailure, nil
		case StatusSuccess:
			r.n++
			if r.Times > 0 && r.n >= r.Times {
				r.n = 0
				return StatusSuccess, nil
			}
			if r.Times == 0 {
				// forever mode: one successful child tick per outer tick
				return StatusRunning, nil
			}
		}
	}
}

// ActionNode executes a single Action then succeeds.
type ActionNode struct {
	Action action.Action
}

func (a *ActionNode) Tick(ctx context.Context, _ world.World, ex executor.Executor) (Status, error) {
	if err := ctx.Err(); err != nil {
		return StatusFailure, err
	}
	if ex == nil {
		return StatusFailure, nil
	}
	if err := ex.Execute(ctx, a.Action); err != nil {
		return StatusFailure, err
	}
	return StatusSuccess, nil
}
