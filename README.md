# AutoDream

Modular game/desktop **automation framework** in Go. Not a script for a single game — a library other projects extend.

## Status

Early development on branch [`dev`](https://github.com/Nigh/autodream/tree/dev). See [AGENTS.md](AGENTS.md) for architecture, test matrix, and phase progress.

## Requirements

- Go 1.24+
- Linux: core + fakes + HTTP capture client are testable
- Windows: native DXGI capture and input executors

## Quick test (Linux)

```bash
git clone https://github.com/Nigh/autodream.git
cd autodream
git checkout dev
go test ./...
go run ./cmd/demo -capture=fake -duration=1s
```

### Demo flags

| Flag | Default (Linux) | Notes |
|------|-----------------|-------|
| `-capture` | `fake` | `fake` \| `dxgihttp` \| `dxgi` (Windows) |
| `-dxgihttp-url` | `http://127.0.0.1:3000` | Nigh/dxgi-capture service |
| `-duration` | `3s` | `0` = until Ctrl-C |
| `-real-executor` | false | Windows OS input via mux |

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
