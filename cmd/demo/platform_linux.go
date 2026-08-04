//go:build linux

package main

import (
	"fmt"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/capture/x11"
	"github.com/Nigh/autodream/automation/executor"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	"github.com/Nigh/autodream/automation/executor/keyboard"
	"github.com/Nigh/autodream/automation/executor/mouse"
	"github.com/Nigh/autodream/automation/executor/mux"
	autolog "github.com/Nigh/autodream/automation/log"
)

func openDXGICapture(int, autolog.Logger) (capture.Capture, error) {
	return nil, fmt.Errorf("demo: dxgi capture requires Windows; use -capture=x11, fake, or dxgihttp")
}

func openX11Capture(screen int, logger autolog.Logger) (capture.Capture, error) {
	return x11.New(x11.Config{Screen: screen, Logger: logger})
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
