# AutoDream

Modular game/desktop **automation framework** in Go. Not a script for a single game — a library other projects extend.

## Status

Early development on branch [`dev`](https://github.com/Nigh/autodream/tree/dev). See [AGENTS.md](AGENTS.md) for architecture, test matrix, and phase progress.

**Workflow authoring (best practices):** [docs/workflow.md](docs/workflow.md)

## Requirements

- Go 1.24+
- Linux: core + fakes + `capture/x11` (X11/XWayland) + XTest mouse/keyboard
- Windows: native DXGI capture and input executors

## Quick test (Linux)

```bash
git clone https://github.com/Nigh/autodream.git
cd autodream
git checkout dev
go test ./...
go run ./cmd/demo -capture=fake -duration=1s
# with a local X display:
go run ./cmd/demo -capture=x11 -duration=1s
```

### Demo flags

| Flag | Default | Notes |
|------|---------|-------|
| `-capture` | Linux: `x11` if `$DISPLAY` else `fake`; Windows: `dxgi` | `fake` \| `dxgihttp` \| `dxgi` \| `x11` |
| `-dxgihttp-url` | `http://127.0.0.1:3000` | Nigh/dxgi-capture service |
| `-display` | `0` | DXGI output / X11 screen index |
| `-duration` | `3s` | `0` = until Ctrl-C |
| `-real-executor` | false | Windows user32 or Linux XTest via mux (manual only) |

## Layout

```
automation/
  capture/       # Frame sources (interface + impls)
  recognition/   # Frame → Result
  executor/      # Actions
  pipeline/      # Sequence / Selector / …
  world/         # Shared state
  runtime/       # Tick loop
  frame/ action/ log/ config/
cmd/demo/        # Wiring example
examples/configs/
AGENTS.md        # Architecture + progress (keep in sync)
```

## License

MIT
