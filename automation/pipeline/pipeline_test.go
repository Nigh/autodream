package pipeline_test

import (
	"context"
	"testing"

	"github.com/Nigh/autodream/automation/action"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	"github.com/Nigh/autodream/automation/pipeline"
	"github.com/Nigh/autodream/automation/world"
)

func TestSequenceSelectorAction(t *testing.T) {
	w := world.NewMemory()
	w.Set("ready", true)
	ex := &exfake.Executor{}
	root := &pipeline.Sequence{
		Children: []pipeline.Node{
			&pipeline.Condition{Pred: func(w world.World) (bool, error) {
				v, ok := w.Get("ready")
				return ok && v.(bool), nil
			}},
			&pipeline.ActionNode{Action: action.Action{Kind: action.KindNoop}},
		},
	}
	st, err := root.Tick(context.Background(), w, ex)
	if err != nil || st != pipeline.StatusSuccess {
		t.Fatalf("st=%v err=%v", st, err)
	}
	if ex.Len() != 1 {
		t.Fatalf("actions=%d", ex.Len())
	}

	sel := &pipeline.Selector{Children: []pipeline.Node{
		&pipeline.Condition{Pred: func(world.World) (bool, error) { return false, nil }},
		&pipeline.ActionNode{Action: action.Action{Kind: action.KindKeyTap, Key: "a"}},
	}}
	st, err = sel.Tick(context.Background(), w, ex)
	if err != nil || st != pipeline.StatusSuccess || ex.Len() != 2 {
		t.Fatalf("sel st=%v err=%v n=%d", st, err, ex.Len())
	}
}
