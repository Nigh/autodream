package runtime

import (
	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/executor"
	"github.com/Nigh/autodream/automation/log"
	"github.com/Nigh/autodream/automation/pipeline"
	"github.com/Nigh/autodream/automation/recognition"
	"github.com/Nigh/autodream/automation/world"
)

// Config wires interface dependencies for the automation loop.
// Run is implemented in a later phase.
type Config struct {
	Capture     capture.Capture
	Recognizers []recognition.Recognizer
	World       world.World
	Root        pipeline.Node
	Executor    executor.Executor
	Logger      log.Logger
}

// Runtime schedules Capture → Recognize → World → Pipeline → Executor.
type Runtime struct {
	cfg Config
}

// New validates required deps and returns a Runtime.
func New(cfg Config) (*Runtime, error) {
	if cfg.Logger == nil {
		cfg.Logger = log.Nop()
	}
	if cfg.Capture == nil {
		return nil, errMissing("capture")
	}
	if cfg.World == nil {
		return nil, errMissing("world")
	}
	if cfg.Root == nil {
		return nil, errMissing("pipeline root")
	}
	if cfg.Executor == nil {
		return nil, errMissing("executor")
	}
	return &Runtime{cfg: cfg}, nil
}
