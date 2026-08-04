package pipeline

import (
	"context"

	"github.com/Nigh/autodream/automation/executor"
	"github.com/Nigh/autodream/automation/world"
)

// Sequence ticks children in order; fails on first Failure; Running short-circuits.
type Sequence struct {
	Children []Node
	index    int
}

func (s *Sequence) Tick(ctx context.Context, w world.World, ex executor.Executor) (Status, error) {
	for s.index < len(s.Children) {
		if err := ctx.Err(); err != nil {
			return StatusFailure, err
		}
		st, err := s.Children[s.index].Tick(ctx, w, ex)
		if err != nil {
			return StatusFailure, err
		}
		switch st {
		case StatusRunning:
			return StatusRunning, nil
		case StatusFailure:
			s.index = 0
			return StatusFailure, nil
		case StatusSuccess:
			s.index++
		}
	}
	s.index = 0
	return StatusSuccess, nil
}

// Selector ticks children until one succeeds.
type Selector struct {
	Children []Node
	index    int
}

func (s *Selector) Tick(ctx context.Context, w world.World, ex executor.Executor) (Status, error) {
	for s.index < len(s.Children) {
		if err := ctx.Err(); err != nil {
			return StatusFailure, err
		}
		st, err := s.Children[s.index].Tick(ctx, w, ex)
		if err != nil {
			return StatusFailure, err
		}
		switch st {
		case StatusRunning:
			return StatusRunning, nil
		case StatusSuccess:
			s.index = 0
			return StatusSuccess, nil
		case StatusFailure:
			s.index++
		}
	}
	s.index = 0
	return StatusFailure, nil
}
