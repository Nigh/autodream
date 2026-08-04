package world

import (
	"errors"
	"fmt"
	"maps"
	"sync"
)

// Update is the authoritative mutation unit applied to World.
type Update struct {
	Key   string
	Value any
}

// World holds automation state shared between recognition results and pipeline decisions.
type World interface {
	Get(key string) (any, bool)
	Set(key string, v any)
	Apply(updates []Update) error
	Snapshot() map[string]any
}

// Memory is a mutex-guarded in-process World.
type Memory struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewMemory creates an empty World.
func NewMemory() *Memory {
	return &Memory{data: make(map[string]any)}
}

func (m *Memory) Get(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[key]
	return v, ok
}

func (m *Memory) Set(key string, v any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = v
}

func (m *Memory) Apply(updates []Update) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var errs []error
	for _, u := range updates {
		if u.Key == "" {
			errs = append(errs, fmt.Errorf("world: empty update key"))
			continue
		}
		m.data[u.Key] = u.Value
	}
	return errors.Join(errs...)
}

func (m *Memory) Snapshot() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return maps.Clone(m.data)
}
