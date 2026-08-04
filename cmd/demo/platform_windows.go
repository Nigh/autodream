//go:build windows

package main

import (
	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/capture/dxgi"
	"github.com/Nigh/autodream/automation/executor"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	"github.com/Nigh/autodream/automation/executor/keyboard"
	"github.com/Nigh/autodream/automation/executor/mouse"
	"github.com/Nigh/autodream/automation/executor/mux"
	autolog "github.com/Nigh/autodream/automation/log"
)

func openDXGICapture(displayID int, logger autolog.Logger) (capture.Capture, error) {
	return dxgi.New(dxgi.Config{
		OutputIndex: uint(displayID),
		Logger:      logger,
	})
}

func openExecutor(real bool) (executor.Executor, error) {
	if !real {
		return &exfake.Executor{}, nil
	}
	m := mux.New()
	ms := mouse.New()
	kb := keyboard.New()
	m.Handle(action.KindMouseMove, ms)
	m.Handle(action.KindMouseClick, ms)
	m.Handle(action.KindMouseButton, ms)
	m.Handle(action.KindKey, kb)
	m.Handle(action.KindKeyTap, kb)
	return m, nil
}
