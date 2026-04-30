# `helix-tv-input` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-tv-input`                                                                                                       |
| **Origin chapter:section**  | [C12 §6](../../03_Architecture/11_TV_UX.md) — *TV-first remote / D-pad input dispatch + 64 dp focus targeting*          |
| **Public path (4 mirrors)** | `vasic-digital/helix-tv-input` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `gioui.org/io/event`, `golang.org/x/mobile/event/key`, Android-TV JNI bridge (`gomobile`)                              |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `android-tv-input-1.x` — builder `golang-builder`, runtime `distroless-static`                                      |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/01_minimum_viable_session/scenarios/04_tv_remote_navigation_full_journey.scenario.yaml`                  |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-tv-input` is the **TV-first input dispatch primitive** for HelixPlay's living-room clients. It abstracts the keycode + focus-traversal model that Compose-for-TV (the primary Android-TV path per C12 §6 MC-05 closure) and Wails desktop clients consume, providing a unified Go-side API for D-pad / remote-control event handling. The submodule was introduced in C12 §6 specifically to consolidate input-dispatch logic across Compose-for-TV (Android), Wails Linux (gioui-driven), and the Steam Deck game-mode client.

The submodule's design is rooted in C12 §6's WCAG 2.2 SC 2.5.8 conformance work — focus targets must be ≥ 64 dp on the longer side, and the dispatch logic must guarantee that focus traversal never lands on an off-screen or zero-sized element. The submodule encodes those rules in its public API rather than leaving them to be re-implemented per client.

`★ Why a Go submodule and not a Kotlin / Swift library.` The Compose-for-TV path lives in Kotlin, and the Steam Deck path lives in Go. The Wails desktop path lives in Go with a JS frontend. A Kotlin-only library would force the Steam Deck and Wails paths to re-implement the focus-traversal rules; a Go submodule plus a thin Kotlin JNI shim (built via `gomobile bind` from this submodule) gives both paths the same tested implementation. C12 §6 documents the JNI shim's stability surface.

---

## 2. Public API Surface

### 2.1 Event types

```go
package tvinput

// Event is the unified TV-input event. Every Compose-for-TV /
// Wails / Steam Deck input source produces an Event; downstream
// consumers handle the unified type rather than the source-
// specific raw event.
type Event struct {
    Kind      EventKind     // KeyDown, KeyUp, FocusEnter, FocusExit, MediaButton
    KeyCode   KeyCode       // DpadUp, DpadDown, DpadLeft, DpadRight, Center, Back, Home, MediaPlay, MediaPause, ...
    SourceID  string        // "android-tv-remote-1", "wails-keyboard", "steam-deck-button-A"
    Timestamp time.Time
    Pressure  float32       // 0.0–1.0 for analog triggers; 1.0 for digital buttons
}
```

### 2.2 The `Dispatcher` type

```go
package tvinput

// Dispatcher routes events to focused widgets and enforces the
// 64-dp focus-target rule. It maintains a focus tree and a focus
// pointer; events are dispatched to the focused widget unless
// they are global navigation keys (Back, Home).
type Dispatcher struct { /* ... */ }

func NewDispatcher(rootFocus FocusTree) *Dispatcher
func (d *Dispatcher) Dispatch(ev Event) Result
func (d *Dispatcher) FocusedID() string
func (d *Dispatcher) MoveFocus(direction Direction) Result
```

### 2.3 The `FocusTree` interface

```go
package tvinput

// FocusTree is the client's focus topology. The dispatcher uses
// it to compute the next focus target on D-pad navigation.
type FocusTree interface {
    Children() []FocusNode
    Focusable() bool
    Bounds() Rect            // in device-independent pixels (dp)
    ID() string
}

type FocusNode interface {
    FocusTree
    Parent() FocusTree
}
```

### 2.4 The `Bounds` validator

```go
package tvinput

// ValidateFocusTree walks the FocusTree and returns errors for any
// node that violates the 64-dp rule (Bounds().Width < 64 ||
// Bounds().Height < 64) AND has Focusable() == true. This is
// invoked at FocusTree mount time; CI fails the test row if any
// validator error is returned for a known FocusTree.
func ValidateFocusTree(root FocusTree) []FocusError
```

The validator is the operationalisation of [C12 §6](../../03_Architecture/11_TV_UX.md)'s WCAG 2.2 SC 2.5.8 conformance — a focus target smaller than 64 dp on the longer side fails WCAG and would fail the EU Accessibility Act 2025 audit.

### 2.5 The Compose-for-TV JNI shim

```go
package tvinput

// AndroidEventFromJNI converts a raw Android KeyEvent (passed via
// gomobile bind) into the unified Event. The Kotlin side calls
// this on every key event before forwarding to the Dispatcher.
func AndroidEventFromJNI(keyCode int, action int, eventTime int64, source int) Event
```

The JNI shim is generated by `gomobile bind`; the Kotlin side imports the resulting AAR and calls `TvInput.androidEventFromJNI(...)` from its `Activity.onKeyDown`.

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (e.g. `getprop ro.product.model` on Android-TV via `helixctl`-forwarding for diagnostic logging).

### 3.2 External (Go)

- `gioui.org/io/event` — Wails desktop event types.
- `golang.org/x/mobile/event/key` — Android key event types.
- `gomobile` (build-time) — for the JNI shim generation.

### 3.3 External (system)

- Compose-for-TV's `androidx.tv.material3` 1.0 GA + 1.1.0-rc01 (consumer-side; this submodule does not embed Compose, but its API is shaped to match Compose's `Modifier.focusable` semantics per C12 §6).

---

## 4. Container Build (S02 §3 lane: `android-tv-input-1.x`)

**Builder:** `golang-builder` (no cgo; the JNI shim is built separately via `gomobile bind`, not in the container). **Runtime:** `distroless-static`. **Multi-arch:** `linux/amd64` + `linux/arm64`. The Android AAR is built in a separate Android SDK container (S02 §3.2 referenced as the cross-builder variant) but published alongside the main artefact.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
- Every key code → unified Event mapping.
- Focus-tree traversal (every direction, every cycle case).
- `ValidateFocusTree` on synthetic trees with 64-dp violations.

### 5.2 Integration
- Real Android-TV emulator (Android Studio's TV emulator profile) running the JNI shim against a real Compose-for-TV widget tree.
- Real Wails desktop with a real gioui event loop.

### 5.3 E2E
- Full client startup → focus tree mount → 100-key navigation sequence → expected focus path.

### 5.4 Security
- govulncheck + Snyk + Trivy. The submodule has no native code paths beyond the JNI shim, so the security surface is small.

### 5.5 Benchmarking
- Dispatch latency p999 ≤ 100 µs.

### 5.6 Chaos
- Random key event injection at 1 KHz; verify no dropped events.

### 5.7 Stress
- 24-hour run with 60 fps focus animations + continuous D-pad navigation; verify no focus-pointer drift.

### 5.8 Smoke
- 30-second post-deploy: load FocusTree, invoke `MoveFocus(DirectionDown)` 5 times, verify focus pointer changed.

### 5.9 Full Automation
- §5.1–§5.8 in CI matrix.

### 5.10 Challenges
- `01_minimum_viable_session/04_tv_remote_navigation_full_journey.scenario` — full user journey with TV remote: power on, navigate to Browse, select a title, press Play, control playback, return to Browse.

---

## 6. Challenges Entry-Point (S03 §4 row #03)

**Topology:** `01_minimum_viable_session`. **Scenario:** `04_tv_remote_navigation_full_journey.scenario.yaml`. **Why this scenario.** The TV-first user journey is the canonical reference user flow ([System Overview §3](../../02_System_Overview.md#3-reference-user-journey)); the scenario exercises every public API of `helix-tv-input` end-to-end. **Baseline:** focus path SHA-256 (deterministic given the recorded input), per-event dispatch latency histogram, OTLP trace topology.

---

## 7. R-18 Inheritance

`helix-tv-input` imports `helix-r18-safeexec` for diagnostic-only subprocess invocations (querying device properties on Android-TV). The hot path (focus dispatch) does not invoke subprocesses; the R-18 surface is minimal.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation gated on the C12 §6 API freeze and Compose-for-TV `androidx.tv.material3` 1.1.0 going GA upstream (currently 1.1.0-rc01 per C12 §6 MC-05 closure note).

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                          | Default      | Range / type                | Purpose                                                                |
|----------------------------------|--------------|-----------------------------|------------------------------------------------------------------------|
| `HELIX_TV_INPUT_FOCUS_RING_COLOR`| `#FFAA00`    | hex RGB                     | Default focus-ring colour; tenants override via `helix-tenant`.       |
| `HELIX_TV_INPUT_REPEAT_DELAY_MS` | `500`        | int [100, 2000]             | D-pad repeat first-delay (long-press detection threshold).            |
| `HELIX_TV_INPUT_REPEAT_RATE_MS`  | `60`         | int [16, 250]               | D-pad repeat rate after first-delay; 60 ms = ~16 events/s.            |
| `HELIX_TV_INPUT_DEBUG_OVERLAY`   | `false`      | bool                        | Render an overlay showing focus path + dispatch timing (dev mode).    |
| `HELIX_TV_INPUT_VOICE_ENABLED`   | `false`      | bool                        | Opt-in to voice-input event surfacing (off until OQ-C resolved).      |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `Dispatch(Event)` end-to-end          | 30 µs    | 80 µs   | 100 µs  | Lock-free focus pointer; widget callback excluded.               |
| `MoveFocus(direction)` decision       | 15 µs    | 40 µs   | 70 µs   | Walks focus tree; depth ≤ 6 in canonical TV UIs.                |
| `ValidateFocusTree` (mount-time)      | 50 ms    | 200 ms  | 400 ms  | Once per FocusTree mount; not in the hot path.                  |
| Event-to-render dispatch              | 80 µs    | 250 µs  | 500 µs  | Bounded by Compose-for-TV / gioui frame deadline.               |
| Memory per FocusTree node             | 256 B    | 512 B   | 1 KiB   | Steady-state; transient bursts during reflow tolerated.          |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                  |
|-----------------------------------------------|----------------------------------------------------------------|------------------------------------------------------------------------------|
| `FocusError: target smaller than 64 dp`       | Focusable widget violates WCAG 2.2 SC 2.5.8                    | Resize the widget; if intentional (decorative), set `Focusable() = false`.   |
| `FocusError: orphan node in focus tree`       | Node has no parent reachable from root                         | Re-attach to root; debug overlay reveals the orphan's coordinates.          |
| `Dispatch: no focused widget`                 | Initial focus not set on FocusTree mount                       | Call `Dispatcher.MoveFocus(DirectionInitial)` after `NewDispatcher`.         |
| `JNI: AndroidEventFromJNI invalid keyCode`    | Android KeyEvent unmapped                                       | Add the keyCode to `tvinput/keymap.go`; emit a metric for unmapped events.  |
| `Dispatch: focus-ring jitter`                 | Theme animations interfering with focus-pointer translation    | Disable per-tenant animation; check `helix-tenant.Theme.Animations`.        |

### 9.4 Migration from per-platform input dispatch

A consumer migrating from inline keyEvent handling to `helix-tv-input`:

1. Replace the per-platform `KeyEvent` listener with a unified `tvinput.Dispatcher.Dispatch(ev)` call.
2. Define the FocusTree once; derive from the existing widget hierarchy.
3. Run `tvinput.ValidateFocusTree(root)` at mount; fix any 64-dp violations before shipping.
4. For Compose-for-TV: replace `Modifier.onKeyEvent { ... }` chains with the SDK's `TvInput.dispatch(...)` Kotlin shim.
5. For Wails / Steam Deck: replace `gioui` event-loop key handlers with `Dispatcher.Dispatch`.
6. Migrate per-platform unit tests; the unified Event type means tests can be written once and run on every client surface.

The migration is documented in `docs/migration-from-per-platform-input.md` in the submodule's repo. Steam Deck specifics (per OQ-tv-input-B above) are deferred until C03 §6 next revision settles the SteamInput overlap.

### 9.5 Observability metrics catalog

Per Constitution §10, the submodule emits the following Prometheus metrics. Every metric carries the `tenant_id`, `client_id`, and `submodule="helix-tv-input"` labels at minimum.

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_tv_input_dispatch_total`                 | counter    | Events dispatched, labelled `key_code`, `kind`, `result={handled, ignored}`.|
| `helix_tv_input_dispatch_latency_seconds`       | histogram  | Dispatch end-to-end latency; buckets at 30 µs, 100 µs, 300 µs, 1 ms.         |
| `helix_tv_input_focus_moves_total`              | counter    | Focus traversal events, labelled `direction`.                                |
| `helix_tv_input_focus_violations_total`         | counter    | 64-dp / orphan-node violations detected at mount.                            |
| `helix_tv_input_jni_invalid_keycode_total`      | counter    | Android JNI shim received an unmapped keyCode.                               |
| `helix_tv_input_repeat_events_total`            | counter    | D-pad repeat events emitted.                                                 |
| `helix_tv_input_voice_events_total`             | counter    | Voice events surfaced (only if `HELIX_TV_INPUT_VOICE_ENABLED=true`).         |

Recommended Grafana dashboard panels: dispatch p99 latency, focus violations rate (alert threshold: any non-zero rate is a defect), JNI invalid-keycode rate (alert: spikes indicate an Android-OEM-specific keyCode that needs mapping).

### 9.6 Consumer matrix (chapters that import this submodule)

`helix-tv-input` is consumed by the following chapters' client-side integration surfaces. The submodule is exclusively client-side; no server-side consumer exists.

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C04 §6](../../03_Architecture/04_Go_Client_Ecosystem.md) — Go Client Ecosystem            | Wails desktop / Steam Deck client integrate via the Go API.                          |
| [C12 §6](../../03_Architecture/11_TV_UX.md) — TV UX (origin)                               | Origin chapter; Compose-for-TV consumes the Kotlin JNI shim AAR.                     |
| [C03 §6](../../03_Architecture/02_Controller_Input_Pipeline.md) — Controller Input         | Cross-references for the Steam Deck's button-event surfacing (OQ-tv-input-B).       |
| [C11 §6](../../03_Architecture/10_WhiteLabel_and_Theming.md) — White-Label                 | Theme-driven focus-ring colour via `helix-tenant.Theme`.                            |

Future chapters (T01..T02 testing, P00..P13 implementation phases) will add the per-platform integration rows; the matrix is updated when those land.

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-tv-input-A       | Fire TV VSK launcher integration — direct or launcher-mediated only (per C12 Z-1 closure)?                     | C12 §6 next revision                                |
| OQ-tv-input-B       | Steam Deck game-mode input — does the SteamInput SDK overlap the JNI shim?                                     | C03 §6 next revision                                |
| OQ-tv-input-C       | Voice-input integration (e.g. Google Assistant) — surface as a separate Event.Kind or consume as MediaButton?  | C12 §6 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../03_Architecture/11_TV_UX.md`](../../03_Architecture/11_TV_UX.md) §6 | (slice) | 2026-04-30 | origin chapter; full API surface; WCAG 2.2 SC 2.5.8 |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #03                                 |
| [`helix-r18-safeexec.md`](helix-r18-safeexec.md)                   | (this batch) | 2026-04-30 | dependency                              |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target (this submodule)                                                              |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (every keyCode mapping + focus-traversal direction).               |
| Integration    | Real Android-TV emulator + real Wails gioui event loop both exercised.                       |
| E2E            | Full focus-tree mount → 100-key navigation sequence with deterministic expected path.        |
| Security       | govulncheck + Snyk + Trivy; zero high findings on every run.                                  |
| Benchmarking   | Dispatch p999 ≤ §9.2 budget; FocusTree validation < 400 ms on typical trees.                 |
| Chaos          | Random-key injection at 1 KHz; zero dropped events.                                           |
| Stress         | 24-hour run with 60 fps focus animations; zero focus-pointer drift.                          |
| Smoke          | 30-second post-deploy: load FocusTree + 5 MoveFocus calls verify pointer changed.           |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `01_minimum_viable_session/04_tv_remote_navigation_full_journey` baseline-parity.             |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-tv-input.md` — 2026-04-30.
