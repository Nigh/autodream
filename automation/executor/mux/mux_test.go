package mux_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/executor"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	"github.com/Nigh/autodream/automation/executor/mux"
)

func TestMuxRoute(t *testing.T) {
	m := mux.New()
	mouse := &exfake.Executor{}
	m.Handle(action.KindMouseClick, mouse)
	if err := m.Execute(context.Background(), action.Action{Kind: action.KindMouseClick, X: 1}); err != nil {
		t.Fatal(err)
	}
	if mouse.Len() != 1 {
		t.Fatal(mouse.Len())
	}
	err := m.Execute(context.Background(), action.Action{Kind: action.KindKeyTap})
	if !errors.Is(err, executor.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}
