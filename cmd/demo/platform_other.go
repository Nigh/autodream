//go:build !windows

package main

import (
	"fmt"

	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/executor"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	autolog "github.com/Nigh/autodream/automation/log"
)

func openDXGICapture(int, autolog.Logger) (capture.Capture, error) {
	return nil, fmt.Errorf("demo: dxgi capture requires Windows; use -capture=fake or -capture=dxgihttp")
}

func openExecutor(real bool) (executor.Executor, error) {
	if real {
		return nil, fmt.Errorf("demo: -real-executor requires Windows")
	}
	return &exfake.Executor{}, nil
}
