package capture

import (
	"context"
	"errors"

	"github.com/Nigh/autodream/automation/frame"
)

// ErrNotStarted is returned when CurrentFrame is called before Start.
var ErrNotStarted = errors.New("capture: not started")

// ErrStopped is returned when the capture has been stopped.
var ErrStopped = errors.New("capture: stopped")

// Capture produces shared read-only Frames. Implementations must not run recognition.
type Capture interface {
	Start(ctx context.Context) error
	Stop() error
	CurrentFrame() (*frame.Frame, error)
}
