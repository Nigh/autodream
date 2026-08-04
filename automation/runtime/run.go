package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nigh/autodream/automation/world"
)

// RunOptions controls the scheduler loop.
type RunOptions struct {
	// TickInterval is the minimum time between ticks. Zero means as-fast-as-possible with yield.
	TickInterval time.Duration
	// OnTickError if set is called for non-fatal tick errors; if it returns false, Run stops.
	OnTickError func(error) bool
}

// Run starts Capture and loops until ctx is cancelled.
func (r *Runtime) Run(ctx context.Context, opt RunOptions) error {
	if err := r.cfg.Capture.Start(ctx); err != nil {
		return fmt.Errorf("runtime: capture start: %w", err)
	}
	defer func() {
		_ = r.cfg.Capture.Stop()
	}()

	r.cfg.Logger.Info("runtime started")
	defer r.cfg.Logger.Info("runtime stopped")

	var ticker *time.Ticker
	if opt.TickInterval > 0 {
		ticker = time.NewTicker(opt.TickInterval)
		defer ticker.Stop()
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.tick(ctx); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			r.cfg.Logger.Error("tick failed", "err", err)
			if opt.OnTickError != nil && !opt.OnTickError(err) {
				return err
			}
		}
		if ticker != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		} else {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(time.Millisecond) // ponytail: avoid busy spin; upgrade = adaptive backoff
			}
		}
	}
}

func (r *Runtime) tick(ctx context.Context) error {
	f, err := r.cfg.Capture.CurrentFrame()
	if err != nil {
		return fmt.Errorf("runtime: frame: %w", err)
	}

	var updates []world.Update
	var errs []error
	for _, rec := range r.cfg.Recognizers {
		if rec == nil {
			continue
		}
		res, err := rec.Recognize(ctx, f)
		if err != nil {
			errs = append(errs, fmt.Errorf("runtime: recognizer %s: %w", rec.Name(), err))
			continue
		}
		updates = append(updates, res.Updates...)
		r.cfg.Logger.Trace("recognized", "name", rec.Name(), "updates", len(res.Updates))
	}
	if len(updates) > 0 {
		if err := r.cfg.World.Apply(updates); err != nil {
			errs = append(errs, fmt.Errorf("runtime: world apply: %w", err))
		}
	}

	st, err := r.cfg.Root.Tick(ctx, r.cfg.World, r.cfg.Executor)
	if err != nil {
		errs = append(errs, fmt.Errorf("runtime: pipeline: %w", err))
	} else {
		r.cfg.Logger.Trace("pipeline", "status", st)
	}
	return errors.Join(errs...)
}
