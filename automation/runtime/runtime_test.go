package runtime_test

import (
	"context"
	"testing"

	"github.com/Nigh/autodream/automation/capture/fake"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	"github.com/Nigh/autodream/automation/executor"
	"github.com/Nigh/autodream/automation/pipeline"
	"github.com/Nigh/autodream/automation/runtime"
	"github.com/Nigh/autodream/automation/world"
)

type successNode struct{}

func (successNode) Tick(context.Context, world.World, executor.Executor) (pipeline.Status, error) {
	return pipeline.StatusSuccess, nil
}

func TestNewRequiresDeps(t *testing.T) {
	if _, err := runtime.New(runtime.Config{}); err == nil {
		t.Fatal("expected error")
	}
	rt, err := runtime.New(runtime.Config{
		Capture:  fake.NewSolid(1, 1, 0, 0, 0, 255),
		World:    world.NewMemory(),
		Root:     successNode{},
		Executor: &exfake.Executor{},
	})
	if err != nil || rt == nil {
		t.Fatalf("rt=%v err=%v", rt, err)
	}
}
