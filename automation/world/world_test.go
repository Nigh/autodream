package world_test

import (
	"testing"

	"github.com/Nigh/autodream/automation/world"
)

func TestMemoryApply(t *testing.T) {
	w := world.NewMemory()
	err := w.Apply([]world.Update{
		{Key: "hp", Value: 100},
		{Key: "", Value: 1},
		{Key: "mp", Value: 50},
	})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
	if v, ok := w.Get("hp"); !ok || v.(int) != 100 {
		t.Fatalf("hp=%v ok=%v", v, ok)
	}
	if v, ok := w.Get("mp"); !ok || v.(int) != 50 {
		t.Fatalf("mp=%v ok=%v", v, ok)
	}
	snap := w.Snapshot()
	snap["hp"] = 0
	if v, _ := w.Get("hp"); v.(int) != 100 {
		t.Fatal("snapshot must be a clone")
	}
}
