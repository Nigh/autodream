package mux

import (
	"context"
	"fmt"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
)

// Mux routes actions to the first registered executor that accepts the kind.
type Mux struct {
	routes map[action.Kind]executor.Executor
}

// New creates an empty mux.
func New() *Mux {
	return &Mux{routes: make(map[action.Kind]executor.Executor)}
}

// Handle registers ex for kind.
func (m *Mux) Handle(kind action.Kind, ex executor.Executor) {
	m.routes[kind] = ex
}

func (m *Mux) Execute(ctx context.Context, a action.Action) error {
	if a.Kind == action.KindNoop {
		return nil
	}
	ex, ok := m.routes[a.Kind]
	if !ok || ex == nil {
		return fmt.Errorf("%w: %s", executor.ErrUnsupported, a.Kind)
	}
	return ex.Execute(ctx, a)
}
