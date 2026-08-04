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
```

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
```

## License

MIT (see LICENSE when added).
