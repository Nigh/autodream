package pipeline

import (
	"context"

	"github.com/Nigh/autodream/automation/executor"
	"github.com/Nigh/autodream/automation/world"
)

// Status is the tick result of a Node. Compatible with future behavior-tree semantics.
type Status int

const (
	StatusSuccess Status = iota
	StatusFailure
	StatusRunning
)

// Node is a composable pipeline unit. Implementations decide using World only (no pixels).
type Node interface {
	Tick(ctx context.Context, w world.World, ex executor.Executor) (Status, error)
}
