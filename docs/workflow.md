# Building Reliable Workflows

Audience: humans and agents implementing automation **on top of** AutoDream (not changing the framework core).

This guide describes how to assemble a dependable tick-based workflow using the existing packages. Code paths and types refer to module `github.com/Nigh/autodream`.

## Mental model (non-negotiable)

Every Runtime tick runs the same pipeline:

```
Capture → Frame → Recognizer(s) → World.Apply → Pipeline.Tick → Executor
```

| Layer | Writes | Reads | Must not |
|-------|--------|-------|----------|
| Capture | Frame | — | Recognize, click |
| Recognition | `world.Update` via `Result` | Frame (read-only) | Decide actions, capture again, mutate pixels |
| World | State keys | — | Capture, execute |
| Pipeline | Actions (via Executor) | World only | Read pixels / Frame |
| Executor | OS / device effects | Action | Recognize, mutate World |
| Runtime | Scheduling | All interfaces | Game-specific rules |

**Reliability rule #1:** Perception and decision are separated. If you click inside a Recognizer, the workflow is already wrong.

Reference wiring: [`cmd/demo/main.go`](../cmd/demo/main.go). Architecture facts: [`AGENTS.md`](../AGENTS.md).

---

## Step-by-step construction

### 1. Freeze a World contract

Before writing nodes, list every key the pipeline will read. Treat this as an API.

| Key | Type | Writer (recognizer) | Meaning when true / set |
|-----|------|---------------------|-------------------------|
| `login_btn_found` | bool | template | Login button visible |
| `login_btn_x` / `login_btn_y` | int | template Finding → custom update | Click target |
| `in_lobby` | bool | color / scene | Safe to start match |
| `error_toast` | string | ocr | Non-empty → abort / retry |

**Practices**

- Prefer stable, namespaced keys: `ui.login.btn`, `scene.lobby`.
- Always write both success and failure for booleans (`true` / `false`), so stale `true` cannot survive a missing UI. Built-in `color` / `template` already clear to `false` on miss.
- Document units (pixels are screen-absolute unless you say otherwise).
- Do not overload one key with unrelated meanings.

### 2. Choose Capture for the environment

| Environment | Capture | Notes |
|-------------|---------|--------|
| Unit / CI (headless) | `capture/fake` | Synthetic BGRA frames |
| Linux desktop (X11/XWayland) | `capture/x11` | Root GetImage → BGRA; needs `$DISPLAY` |
| Linux + remote DXGI service | `capture/dxgihttp` | PNG+HTTP; slower; fine for integration |
| Windows production | `capture/dxgi` | Shared BGRA frames; default for real runs |

One Capture per Runtime. Never call capture from a Recognizer.

### 3. Stack Recognizers (perception only)

Order them from cheap → expensive:

1. `recognition/color` — ROI average color  
2. `recognition/template` — SAD match (small templates, tight ROI)  
3. `recognition/ocr` — inject an `ocr.Engine`  
4. `recognition/llmvision` — HTTP vision; highest latency/cost  

**Practices**

- Constrain **ROI** tightly; full-screen template match is fragile and slow.
- Tune thresholds with recorded frames, not live guessing.
- Keep YAML/JSON configs under version control (`examples/configs/` style); load with `automation/config`.
- All recognizers share one Frame per tick — good. Do not copy the full image unless a backend requires it (LLM PNG encode is an intentional exception).
- Put coordinates needed for clicks into World (from `Finding.Extra` via a thin custom Recognizer or post-step in your app). Stock `ActionNode` does not auto-bind findings to clicks.

### 4. Encode decisions as a Pipeline tree

Available nodes (`automation/pipeline`):

| Node | Use when |
|------|----------|
| `Condition` | Gate on World predicates |
| `ActionNode` | Emit one `action.Action` |
| `Wait` | Debounce across ticks (`StatusRunning` until duration elapsed) |
| `Sequence` | Do A then B then C |
| `Selector` | Try A; on failure try B (fallback / mode switch) |
| `Repeat` | Retry or loop a child |

Statuses: `Success` | `Failure` | `Running`. Composites keep an internal index across ticks — a long workflow advances **over many frames**, not inside one blocking function.

**Skeleton: detect → act → settle**

```go
root := &pipeline.Sequence{Children: []pipeline.Node{
    &pipeline.Condition{Pred: func(w world.World) (bool, error) {
        v, ok := w.Get("login_btn_found")
        return ok && v == true, nil
    }},
    &pipeline.ActionNode{Action: action.Action{
        Kind: action.KindMouseClick, X: 960, Y: 540, Button: "left",
    }},
    &pipeline.Wait{Duration: 400 * time.Millisecond},
}}
```

**Skeleton: mode select**

```go
root := &pipeline.Selector{Children: []pipeline.Node{
    battleFlow, // Sequence: Condition(in_battle), …
    lobbyFlow,  // Sequence: Condition(in_lobby), …
    &pipeline.Wait{Duration: 100 * time.Millisecond}, // idle / nothing matched
}}
```

**Practices**

- Prefer shallow trees with named subtrees (`var loginFlow pipeline.Node = &pipeline.Sequence{...}`).
- After every irreversible action, `Wait` or re-`Condition` on a **new** World fact (e.g. `login_btn_found == false`) before continuing.
- Use `Selector` for mutually exclusive scenes; use `Sequence` for ordered steps inside a scene.
- On `Failure` of a `Sequence`, the composite resets its index — design retries with `Repeat` or an outer `Selector` that returns to idle, not with busy loops inside `Pred`.
- Never read `Frame` in a Node. Need new info → add/adjust a Recognizer.

### 5. Wire Executor safely

| Mode | Executor |
|------|----------|
| Tests / headless | `executor/fake` — assert recorded actions |
| Linux real input | `executor/mux` + `mouse` / `keyboard` (XTest; needs XTEST + `$DISPLAY`) |
| Windows real input | `executor/mux` + `mouse` / `keyboard` (/ `gamepad` rumble) |

**Practices**

- Develop with fake executor until the World+Pipeline story is correct.
- Gate real input behind an explicit flag (see demo `-real-executor`).
- Prefer `KindMouseClick` / `KindKeyTap` over raw down/up unless you need holds.
- Validate coordinates against the capture surface (X11 root / DXGI output; multi-monitor is MVP-limited).
- Native Wayland (non-XWayland) capture/input is out of scope for the current MVP.

### 6. Run under Runtime

```go
engine, err := runtime.New(runtime.Config{
    Capture:     cap,
    Recognizers: []recognition.Recognizer{...},
    World:       world.NewMemory(),
    Root:        root,
    Executor:    ex,
    Logger:      log.NewSlog(nil),
})
err = engine.Run(ctx, runtime.RunOptions{
    TickInterval: 50 * time.Millisecond,
    OnTickError: func(err error) bool {
        // return false to stop; true to continue
        return true
    },
})
```

**Practices**

- Cancel with `context` (signal / timeout); do not `panic`.
- Tick interval: fast enough for UI, slow enough for capture+recognizers (start ~30–50ms; measure).
- `OnTickError`: log and continue for transient capture misses; stop on repeated hard failures.
- One Runtime loop per automation process unless you fully understand Frame ownership.

---

## Reliability checklist

Use this before calling a workflow “done”:

1. **Contract** — World keys documented; types stable; falsey clears on miss.  
2. **No cross-layer cheating** — Recognizer does not click; Pipeline does not decode pixels.  
3. **ROI** — Every visual check has a minimal ROI.  
4. **Settle** — After actions, wait or wait-for-state, not fixed sleep-only hope.  
5. **Idempotency** — Re-entering a scene does not double-submit (Condition on “already done” or disable button via World).  
6. **Fallback** — `Selector` idle branch or explicit error Condition (`error_toast != ""`).  
7. **Test pyramid**  
   - Recognizer unit tests with synthetic / fixture frames (`capture/fake`, golden PNGs).  
   - Pipeline tests with pre-seeded `world.Memory` + `executor/fake` (no Capture).  
   - Runtime smoke with fake Capture (see `automation/runtime/run_test.go`).  
8. **Observability** — Logger on Runtime; Trace for per-recognizer update counts; Info for scene transitions (log inside custom Nodes or when World keys flip in your app).  
9. **Platform** — Linux CI green with fake/dxgihttp; Windows-only paths behind build tags / flags.  
10. **Failure policy** — `errors.Is` / `Join` at boundaries; no silent swallow of recognizer errors if they leave World stale.

---

## Anti-patterns

| Anti-pattern | Why it fails | Do instead |
|--------------|--------------|------------|
| Screenshot inside Recognizer | Breaks shared Frame, races, duplicates work | Use Runtime Capture only |
| `time.Sleep` + click in Recognizer | Couples perception to actuation; untestable | `ActionNode` + `Wait` / Condition |
| Pipeline reads `*frame.Frame` | Hidden dependency; breaks DIP | New Recognizer → World key |
| Global mutable flags outside World | Second state channel; lost on tick | World keys only |
| Full-screen template every tick | CPU + flaky | Tight ROI + cheaper color gate first |
| Fire-and-forget clicks with no verify | Desync on lag | Condition on post-state |
| One giant Sequence for whole game | Reset on any Failure loses place | Scene `Selector` + small Sequences |
| Real mouse in unit tests | Non-deterministic / dangerous | `executor/fake` |

---

## Example: login then enter lobby

Intent: find login button → click → wait until lobby marker appears → stop acting.

**World keys:** `login_btn_found`, `in_lobby`

**Recognizers:** template(login) → `login_btn_found`; color/template(lobby) → `in_lobby`

**Pipeline:**

```text
Selector
├─ Sequence  (login)
│  ├─ Condition(login_btn_found && !in_lobby)
│  ├─ ActionNode(MouseClick at known or World coords)
│  └─ Wait(300ms)
├─ Sequence  (done)
│  └─ Condition(in_lobby) → Success (optional Noop)
└─ Wait(50ms)            (idle poll)
```

Assembly lives in **your** `cmd/` or app package — same pattern as demo — not inside `automation/runtime`.

---

## Extending the framework (only when needed)

Stay in application code until one of these is true:

- A new **sensor** → implement `recognition.Recognizer`.  
- A new **actuator** → implement `executor.Executor` and register on `executor/mux`.  
- A new **control-flow** primitive → implement `pipeline.Node` (keep BT-compatible Status).  
- A new **frame source** → implement `capture.Capture`.

Do not add game logic to Runtime. Do not introduce package cycles (recognition must not import pipeline; pipeline must not import recognition).

---

## Agent-oriented summary

When implementing a workflow for a user:

1. Define World key table first.  
2. Add/configure Recognizers that only emit those keys.  
3. Build Pipeline from Condition / Action / Wait / Sequence / Selector / Repeat.  
4. Wire `runtime.Config` in a cmd/app; use fake Capture+Executor until green.  
5. Enable real capture/input with an explicit flag: Windows `dxgi` / Linux `x11` + `-real-executor`.  
6. Update this doc or `AGENTS.md` if you change contracts or add shared patterns.

Canonical demos:

```bash
go run ./cmd/demo -capture=fake -duration=1s   # headless / CI
go run ./cmd/demo -capture=x11 -duration=1s    # Linux with DISPLAY
```
