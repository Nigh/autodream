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
| capture/dxgi | windows tag only | skip on Linux |
| recognition/* | yes | yes (synthetic frames) |
| executor + fake | yes | yes |
| executor/mouse, keyboard, gamepad | windows tag only | skip on Linux |
| cmd/demo | yes | manual; default `-capture=fake` on non-Windows |

```bash
go test ./...
```

## Implementation progress

- [x] Bootstrap — repo, `dev`, AGENTS.md, README
- [ ] Phase A — interfaces + leaf types + fake trio
- [ ] Phase B — Frame pool + dxgihttp + dxgi (windows)
- [ ] Phase C — Runtime + fake integration test
- [ ] Phase D — color / template / ocr / llmvision
- [ ] Phase E — Windows mouse / keyboard / gamepad
- [ ] Phase F — cmd/demo

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
