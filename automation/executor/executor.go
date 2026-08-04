package executor

import (
	"context"
	"errors"

	"github.com/Nigh/autodream/automation/action"
)

// ErrUnsupported is returned when an executor cannot handle an action kind.
var ErrUnsupported = errors.New("executor: unsupported action")

// Executor performs Actions. Must not capture or recognize.
type Executor interface {
	Execute(ctx context.Context, a action.Action) error
}
