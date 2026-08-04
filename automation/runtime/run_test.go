package runtime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nigh/autodream/automation/action"
	capfake "github.com/Nigh/autodream/automation/capture/fake"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	"github.com/Nigh/autodream/automation/pipeline"
	"github.com/Nigh/autodream/automation/recognition"
	recfake "github.com/Nigh/autodream/automation/recognition/fake"
	"github.com/Nigh/autodream/automation/runtime"
	"github.com/Nigh/autodream/automation/world"
)

func TestRunFakePipeline(t *testing.T) {
	w := world.NewMemory()
	ex := &exfake.Executor{}
	root := &pipeline.Sequence{
		Children: []pipeline.Node{
			&pipeline.Condition{Pred: func(w world.World) (bool, error) {
				v, ok := w.Get("seen")
				return ok && v.(bool), nil
			}},
			&pipeline.ActionNode{Action: action.Action{Kind: action.KindMouseClick, X: 1, Y: 2, Button: "left"}},
		},
	}
	rt, err := runtime.New(runtime.Config{
		Capture: capfake.NewSolid(2, 2, 0, 0, 0, 255),
		Recognizers: []recognition.Recognizer{
			recfake.StaticUpdates("mark", world.Update{Key: "seen", Value: true}),
		},
		World:    w,
		Root:     root,
		Executor: ex,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- rt.Run(ctx, runtime.RunOptions{TickInterval: 5 * time.Millisecond})
	}()

	deadline := time.After(2 * time.Second)
	for {
		if ex.Len() >= 1 {
			cancel()
			break
		}
		select {
		case <-deadline:
			cancel()
			t.Fatal("timeout waiting for action")
		case <-time.After(5 * time.Millisecond):
		}
	}
	err = <-done
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("run: %v", err)
	}
	if v, ok := w.Get("seen"); !ok || v != true {
		t.Fatalf("world seen=%v ok=%v", v, ok)
	}
	if ex.Len() < 1 {
		t.Fatal("expected at least one executed action")
	}
}
