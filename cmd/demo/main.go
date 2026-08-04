package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Nigh/autodream/automation/action"
	"github.com/Nigh/autodream/automation/capture"
	capfake "github.com/Nigh/autodream/automation/capture/fake"
	"github.com/Nigh/autodream/automation/capture/dxgihttp"
	"github.com/Nigh/autodream/automation/executor"
	exfake "github.com/Nigh/autodream/automation/executor/fake"
	"github.com/Nigh/autodream/automation/frame"
	autolog "github.com/Nigh/autodream/automation/log"
	"github.com/Nigh/autodream/automation/pipeline"
	"github.com/Nigh/autodream/automation/recognition"
	"github.com/Nigh/autodream/automation/recognition/color"
	rt "github.com/Nigh/autodream/automation/runtime"
	"github.com/Nigh/autodream/automation/world"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "demo: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	captureKind := flag.String("capture", defaultCapture(), "fake | dxgihttp | dxgi")
	httpURL := flag.String("dxgihttp-url", "http://127.0.0.1:3000", "base URL for dxgihttp")
	displayID := flag.Int("display", 0, "display id for dxgi/dxgihttp")
	duration := flag.Duration("duration", 3*time.Second, "run duration (0 = until signal)")
	tick := flag.Duration("tick", 50*time.Millisecond, "runtime tick interval")
	useRealExec := flag.Bool("real-executor", false, "use OS input executors (Windows only)")
	flag.Parse()

	logger := autolog.NewSlog(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cap, err := openCapture(*captureKind, *httpURL, *displayID, logger)
	if err != nil {
		return err
	}
	ex, err := openExecutor(*useRealExec)
	if err != nil {
		return err
	}

	w := world.NewMemory()
	colorRec, err := color.New(color.Config{
		Name: "solid_reddish",
		B:    0, G: 0, R: 180,
		Threshold: 80,
		WorldKey:  "red_present",
		WorldValue: true,
		ROI: frame.ROI{}, // full frame
	})
	if err != nil {
		return err
	}

	root := &pipeline.Sequence{Children: []pipeline.Node{
		&pipeline.Condition{Pred: func(w world.World) (bool, error) {
			v, ok := w.Get("red_present")
			return ok && v == true, nil
		}},
		&pipeline.ActionNode{Action: action.Action{Kind: action.KindNoop}},
	}}

	engine, err := rt.New(rt.Config{
		Capture:     cap,
		Recognizers: []recognition.Recognizer{colorRec},
		World:       w,
		Root:        root,
		Executor:    ex,
		Logger:      logger,
	})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *duration)
		defer cancel()
	}

	logger.Info("demo start", "goos", runtime.GOOS, "capture", *captureKind)
	err = engine.Run(ctx, rt.RunOptions{TickInterval: *tick})
	if err != nil && ctx.Err() == nil {
		return err
	}
	logger.Info("demo done", "red_present", snapshotBool(w, "red_present"), "actions", actionCount(ex))
	return nil
}

func defaultCapture() string {
	if runtime.GOOS == "windows" {
		return "dxgi"
	}
	return "fake"
}

func openCapture(kind, httpURL string, displayID int, logger autolog.Logger) (capture.Capture, error) {
	switch kind {
	case "fake":
		// reddish frame so color recognizer succeeds
		return capfake.NewSolid(64, 64, 0, 0, 200, 255), nil
	case "dxgihttp":
		return dxgihttp.New(dxgihttp.Config{
			BaseURL:   httpURL,
			DisplayID: displayID,
			Logger:    logger,
		})
	case "dxgi":
		return openDXGICapture(displayID, logger)
	default:
		return nil, fmt.Errorf("unknown capture %q", kind)
	}
}

func snapshotBool(w world.World, key string) bool {
	v, ok := w.Get(key)
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

func actionCount(ex executor.Executor) int {
	if f, ok := ex.(*exfake.Executor); ok {
		return f.Len()
	}
	return -1
}
