package fake

import (
	"context"
	"sync"

	"github.com/Nigh/autodream/automation/action"
)

// Executor records actions for assertions.
type Executor struct {
	mu      sync.Mutex
	Actions []action.Action
	Err     error
}

func (e *Executor) Execute(ctx context.Context, a action.Action) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Actions = append(e.Actions, a)
	return e.Err
}

// Len returns how many actions were recorded.
func (e *Executor) Len() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.Actions)
}
