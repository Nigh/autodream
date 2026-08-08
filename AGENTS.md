# AutoDream — Agent Notes

Automation framework (not a game-specific script). Agents and humans keep this file in sync with the repo.

## Goals

- High-performance, modular, interface-first automation framework
- MVP target platform: Windows (native DXGI + input)
- Dev machine: Linux — core packages must build and test here
- Extensible capture / recognition / executor / pipeline; later AI vision

## Locked decisions

| Item | Value |
|------|--------|
| Module | `github.com/Nigh/autodream` |
| Public repo | `https://github.com/Nigh/autodream` |
| Dev branch | `dev` (feature branch → PR → `dev`, squash merge) |
| Capture MVP | Native DXGI (`capture/dxgi`, Windows) + HTTP (`capture/dxgihttp` → [Nigh/dxgi-capture](https://github.com/Nigh/dxgi-capture)) + `capture/fake` |
| Frame | Single shared read-only source; recognizers must not capture or mutate pixels |
| Recognition → World | Recognizers return `Result` with `world.Update`; Runtime applies; no action decisions in recognition |
| Errors | `errors.Join` / `Is` / `As`; no panic |
| Logging | `automation/log` (slog-backed) |

## Architecture (one page)

```
Capture → Frame → Recognizer(s) → World.Apply → Pipeline.Tick → Executor
```

**How to build a reliable app workflow:** [docs/workflow.md](docs/workflow.md) (humans + agents).

| Module | May | Must not |
|--------|-----|----------|
| capture | Produce Frame | Recognize / execute |
| recognition | Frame → Result | Capture / mouse / decide actions |
| world | Hold state | Capture / execute |
| pipeline | Read world, emit actions | Capture / read pixels |
| executor | Run actions | Recognize / mutate world |
| runtime | Schedule | Hard-code game logic |

## Package dependency direction

Leaves: `frame`, `action`, `log`, `world`  
→ `capture`, `recognition`, `executor`  
→ `pipeline` (world + executor + action)  
→ `runtime` (wires interfaces)  
Assembly only in `cmd/demo`. No cycles. Concrete impls in subpackages.

## Platform / test matrix

| Package | Linux build | Linux test |
|---------|-------------|------------|
| frame, action, log, world, config, pipeline, runtime | yes | yes |
| capture + fake, capture/dxgihttp | yes | yes (httptest for HTTP) |
| capture/x11 | linux tag | skip if `$DISPLAY` unset; live GetImage when set |
| capture/dxgi | windows tag only | skip on Linux |
| recognition/* | yes | yes (synthetic frames) |
| executor + fake | yes | yes |
| executor/mouse, keyboard, gamepad | windows (all three); linux mouse/keyboard via XTest (gamepad still windows-only) | no real input injection in CI |
| cmd/demo | yes | manual; default `-capture=x11` when `$DISPLAY` set, else `fake` |

```bash
go test ./...
```

## Implementation progress

- [x] Bootstrap — repo, `dev`, AGENTS.md, README
- [x] Phase A — interfaces + leaf types + fake trio
- [x] Phase B — Frame pool + `capture/dxgihttp` + `capture/dxgi` (windows, godesktopdup)
- [x] Phase C — `Runtime.Run` + pipeline Sequence/Selector/Condition/Wait/Repeat/ActionNode
- [x] Phase D — color / template / ocr (Engine inject) / llmvision (HTTP)
- [x] Phase E — Windows mouse / keyboard / gamepad (rumble) + portable `executor/mux`
- [x] Phase F — `cmd/demo` (Linux: fake/dxgihttp; Windows: dxgi + optional real executor)
- [x] Linux L1 — `capture/x11` (XGetImage / jezek/xgb; no CGO)
- [x] Linux L2 — XTest mouse/keyboard (`mouse_linux.go`, `keyboard_linux.go`; US/QWERTY keycodes)
- [x] Linux L3 — demo (`-capture=x11`, `-real-executor`) + workflow/README/AGENTS

## Phase notes

- `world.Update` is the mutation type; `recognition.Result` carries `[]world.Update`
- Runtime tick: Capture → Recognizers → World.Apply → Root.Tick
- OCR: inject `ocr.Engine`; LLM vision: HTTP JSON; Template: SAD (only alpha==255)
- Capture: Windows `dxgi`; Linux `x11` (X11/XWayland root); portable `dxgihttp` / `fake`
- Executors: Windows mouse/keyboard/gamepad; Linux mouse/keyboard via XTest (needs XTEST extension + `$DISPLAY`); `mux` + `fake` everywhere
- Gamepad MVP: XInput rumble only (`control=vibrate`); no Linux gamepad
- Linux keycodes are layout-approximate (pc105/US); no real input injection in default CI tests
- `capture/x11`: ponytail GetImage socket copy; upgrade = MIT-SHM
- Demo: `go run ./cmd/demo -capture=fake -duration=1s` (headless); `go run ./cmd/demo -capture=x11 -duration=1s` (Linux DISPLAY)
- Deps: `gopkg.in/yaml.v3`, `github.com/shinkar94/godesktopdup`, `golang.org/x/sys`, `github.com/jezek/xgb`

## Git workflow for agents

1. Branch from `dev`: `feat/phase-*` or `fix/*`
2. Implement + update this file
3. `go test ./...` green on Linux
4. Open PR targeting `dev` (squash)

## Hard constraints (ponytail)

- YAGNI; no abstractions not required by the plan
- No new dependency if stdlib / existing dep suffices
- Interfaces only across module boundaries; replaceable impls
- Do not regenerate the whole project in one dump — phase PRs
- Sync AGENTS.md on every functional PR
- When teaching or implementing app workflows, follow [docs/workflow.md](docs/workflow.md)
