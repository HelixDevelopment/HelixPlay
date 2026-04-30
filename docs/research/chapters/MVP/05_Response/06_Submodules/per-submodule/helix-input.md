# `helix-input` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-input`                                                                                                          |
| **Origin chapter:section**  | [C21 §6](../../04_Latency/07_Controller_Input_Optimization.md) — *Controller polling + Reflex round-trip echo*         |
| **Public path (4 mirrors)** | `vasic-digital/helix-input` on GitHub + GitLab + GitFlic + GitVerse                                                    |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `golang.org/x/sys/unix`, `github.com/holoplot/go-evdev`                                                                  |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `controller-input-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                      |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/08_controller_input_low_latency/scenarios/03_reflex_round_trip_under_load.scenario.yaml`                |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-input` is the **server-side controller-input ingress** for HelixPlay's host agent. The submodule receives controller events from the client over the gRPC FrameStream (per [C06 §6](../../03_Architecture/05_RealTime_APIs.md)) plus the dedicated UDP fast-path, polls them at 240 Hz to 1 KHz depending on hardware, applies NVIDIA Reflex-style round-trip echo where supported, and injects them into the host OS's input subsystem (uinput on Linux, ViGEmBus on Windows, etc.). Origin: [C21 §6](../../04_Latency/07_Controller_Input_Optimization.md).

The submodule was introduced to consolidate per-platform input-injection logic. Earlier projects re-implemented uinput/ViGEmBus integration with subtle drift across Steam Deck game-mode, Compose-for-TV, and Wails desktop client paths.

---

## 2. Public API Surface

### 2.1 The `Receiver` type

```go
package input

// Receiver decodes incoming controller events from the gRPC stream
// or UDP fast-path and forwards them to a Sink.
type Receiver struct { /* ... */ }

func NewReceiver(sink Sink, opts ...Option) *Receiver
func (r *Receiver) HandleFrameMetadata(meta FrameMetadata) error
func (r *Receiver) Close() error
```

### 2.2 The `Sink` interface

```go
package input

// Sink is implemented per host-OS input subsystem.
type Sink interface {
    Submit(event Event) error
    Capabilities() SinkCapabilities
    Close() error
}

type Event struct {
    DeviceID  string
    Type      EventType   // Button, Axis, Trigger, Touchpad, IMU
    Code      uint16      // BTN_A, ABS_X, etc.
    Value     int32
    Timestamp time.Time
}
```

### 2.3 Per-platform sinks

```go
package input

// LinuxUinputSink writes events to /dev/uinput.
type LinuxUinputSink struct { /* ... */ }
func NewLinuxUinputSink() (*LinuxUinputSink, error)

// WindowsViGEmSink writes events to a ViGEmBus virtual gamepad.
type WindowsViGEmSink struct { /* ... */ }
func NewWindowsViGEmSink(dllPath string) (*WindowsViGEmSink, error)

// MacOSIOKitSink writes events via IOKit HID.
type MacOSIOKitSink struct { /* ... */ }
func NewMacOSIOKitSink() (*MacOSIOKitSink, error)

// NoopSink discards events (testing).
type NoopSink struct{}
func (NoopSink) Submit(Event) error { return nil }
```

### 2.4 The Reflex round-trip echo

```go
package input

// ReflexEchoer reads the Reflex-style mark token from FrameMetadata
// and echoes it back to the client immediately upon submission to
// the OS input subsystem. The client measures the round-trip and
// uses the measurement for latency-budget reconciliation.
type ReflexEchoer struct { /* ... */ }

func (re *ReflexEchoer) Echo(token uint64, t time.Time) error
```

### 2.5 Polling helpers

```go
package input

// Poller polls events at a configurable rate. Wraps a Receiver +
// Sink chain in a hot-path goroutine that should be promoted via
// helix-rtos.PromoteCurrent for jitter control.
type Poller struct { /* ... */ }

func NewPoller(receiver *Receiver, rateHz int) *Poller
func (p *Poller) Start(ctx context.Context) error
func (p *Poller) Stats() PollerStats
```

### 2.6 Capability detection

```go
package input

func KernelCaps() (Capabilities, error)
type Capabilities struct {
    UinputAvailable    bool
    ViGEmBusInstalled  bool   // Windows-only
    IOKitHIDAvailable  bool   // macOS-only
    EVDEVUserspace     bool   // Linux uinput w/o root
}
```

### 2.7 Statistics

```go
package input

type PollerStats struct {
    EventsReceived  uint64
    EventsSubmitted uint64
    DropsLate       uint64
    DropsInvalid    uint64
    EchoSent        uint64
    PollSkew        time.Duration  // EWMA of poll-tick deviation from target
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`lsusb` for HID device enumeration; `xinput list` for X11 device discovery).

### 3.2 External (Go)

- `golang.org/x/sys/unix` — Linux uinput ioctls.
- `github.com/holoplot/go-evdev` — evdev event constants.

### 3.3 External (system)

- `/dev/uinput` accessible (CAP_SYS_ADMIN on the host) on Linux for the LinuxUinputSink. The submodule does NOT require CAP_SYS_ADMIN inside the container; instead, the operator pre-creates a uinput device on the host and bind-mounts it.
- ViGEmBus driver installed for the WindowsViGEmSink.

---

## 4. Container Build (S02 §3 lane: `controller-input-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2; bind-mount `/dev/uinput` from the operator host (no in-container privilege escalation).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Event decode + sink-routing logic with `NoopSink`.

### 5.2 Integration
Real `/dev/uinput` (Linux) or ViGEmBus (Windows); verify event emerges as a valid HID event.

### 5.3 E2E
Client → gRPC FrameStream → helix-input → Sink → host event subsystem; verify a real game receives the input.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: malformed-event fuzzer (must reject without crashing).

### 5.5 Benchmarking
Receiver decode p999 ≤ 5 µs; Sink submit p999 ≤ 50 µs; echo round-trip p999 ≤ 100 µs.

### 5.6 Chaos
Inject network jitter via Toxiproxy; verify drops are categorised correctly (DropsLate vs DropsInvalid).

### 5.7 Stress
24-hour 240 Hz controller stream; zero leak; PollSkew stays below 100 µs EWMA.

### 5.8 Smoke
30-second post-deploy: send a test BTN_A event; verify it appears in `evtest /dev/input/eventN`.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`08_controller_input_low_latency/03_reflex_round_trip_under_load.scenario` — full topology with controller round-trip while host CPUs are loaded; verify p999 ≤ 8 ms end-to-end.

---

## 6. Challenges Entry-Point (S03 §4 row #13)

**Topology:** `08_controller_input_low_latency`. **Scenario:** `03_reflex_round_trip_under_load.scenario.yaml`. **Why this scenario.** Controller round-trip is the most sensitive latency in HelixPlay's user journey; the scenario exercises the input submodule's ingress + Reflex echo under load. **Baseline:** p999 round-trip ≤ 8 ms; DropsLate count = 0 over the scenario duration; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-input` imports `helix-r18-safeexec` for boundary subprocess invocations (HID enumeration). Hot path (event decode, sink submit, echo) is pure syscall + cgo. The submodule does NOT require CAP_SYS_ADMIN; the operator pre-provisions the uinput device.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the per-platform sink coverage; Linux uinput is the production-priority path, Windows/macOS sinks land later.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_INPUT_POLL_RATE_HZ`         | `240`         | int [60, 1000]          | Poller tick rate.                                                       |
| `HELIX_INPUT_SINK`                 | `auto`        | `linux-uinput` / `windows-vigem` / `macos-iokit` / `noop` / `auto` | Per-platform sink choice.                            |
| `HELIX_INPUT_REFLEX_ENABLED`       | `true`        | bool                    | Enable Reflex round-trip echo.                                         |
| `HELIX_INPUT_DEVICE_NAME`          | `helix-pad`   | string                  | Virtual device name reported to the host OS.                          |
| `HELIX_INPUT_AXES_DEADZONE`        | `2000`        | int [0, 32767]          | Stick deadzone in raw evdev units.                                     |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Receiver decode                       | 1 µs     | 3 µs    | 5 µs    | Protobuf decode + struct fill.                                    |
| LinuxUinputSink.Submit                | 15 µs    | 30 µs   | 50 µs   | uinput write + ioctl.                                            |
| WindowsViGEmSink.Submit               | 30 µs    | 80 µs   | 150 µs  | ViGEmBus IOCTL via DLL.                                          |
| ReflexEchoer.Echo                     | 30 µs    | 70 µs   | 100 µs  | gRPC reverse-stream send.                                         |
| Poll-tick skew (EWMA)                 | 20 µs    | 50 µs   | 100 µs  | Steady-state with helix-rtos pinning.                             |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `input: ErrUinputNotMounted`                  | `/dev/uinput` missing in container                             | Bind-mount from host: `--device /dev/uinput`.                              |
| `input: ErrViGEmBusNotInstalled`              | Windows host missing ViGEmBus driver                           | Operator installs ViGEmBus_1.22.0+ on the host.                           |
| `input: ErrEventTooOld`                        | Event timestamp older than discard threshold                   | Investigate network latency; check `JitterBuffer.DelayMS`.                |
| `input: ErrSinkRefused`                        | Sink hit a per-OS rate limit                                   | Reduce poll rate; investigate downstream consumer.                         |
| `input: ErrPollSkewExcessive`                  | Poll thread not maintaining tick rate                          | Verify helix-rtos promotion is active; check host CPU load.               |

### 9.4 Migration from inline gamepad handling

A consumer migrating from inline evdev / DirectInput / IOKit:

1. Replace platform-specific code with `input.NewReceiver(sink, opts...)` where `sink` is one of the per-OS sinks.
2. Wire the receiver into the gRPC FrameStream from `helix-grpc-frame`.
3. Add `helix-rtos.PromoteCurrent(50)` on the poller goroutine.
4. Add OTLP span around `Receiver.HandleFrameMetadata`; metric for `PollerStats.PollSkew` (alert: > 100 µs sustained).

The migration is documented in `docs/migration-from-inline-gamepad.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_input_events_received_total`             | counter    | Events received from the gRPC stream.                                        |
| `helix_input_events_submitted_total`            | counter    | Events submitted to the OS sink.                                             |
| `helix_input_drops_late_total`                  | counter    | Events dropped due to late timestamp.                                        |
| `helix_input_drops_invalid_total`               | counter    | Events dropped due to invalid framing.                                       |
| `helix_input_echo_sent_total`                   | counter    | Reflex echoes sent back to the client.                                       |
| `helix_input_poll_skew_seconds`                 | gauge      | EWMA of poll-tick skew.                                                      |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C21 §6](../../04_Latency/07_Controller_Input_Optimization.md) — origin                    | Origin chapter; full Receiver + Sink + ReflexEchoer + Poller API.                    |
| [C03 §6](../../03_Architecture/02_Controller_Input_Pipeline.md) — Architecture pipeline   | Server-side pipeline element.                                                        |
| [C12 §6](../../03_Architecture/11_TV_UX.md) — TV UX                                       | Client emits events that helix-input ingests; symmetric pair with helix-tv-input.   |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-input-A          | Steam Deck game-mode integration — SteamInput SDK overlap with this submodule (cf. helix-tv-input OQ-B)?      | C03 §6 next revision                                |
| OQ-input-B          | macOS IOKit sink — production-readiness; Apple Silicon-specific quirks?                                       | C21 §6 next revision                                |
| OQ-input-C          | Reflex round-trip echo — should it move to a dedicated UDP fast-path microservice?                            | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/07_Controller_Input_Optimization.md`](../../04_Latency/07_Controller_Input_Optimization.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #13                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Receiver + Sink + ReflexEchoer + Poller).                         |
| Integration    | Real /dev/uinput or ViGEmBus or IOKit; HID event verification at host.                       |
| E2E            | Full client → server → game input roundtrip.                                                  |
| Security       | govulncheck + Snyk + Trivy + malformed-event fuzzer.                                         |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Network-jitter injection; drop categorisation correct.                                        |
| Stress         | 24-hour 240 Hz; PollSkew ≤ 100 µs EWMA.                                                       |
| Smoke          | 30-second BTN_A injection + evtest verification.                                              |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `08_controller_input_low_latency/03_reflex_round_trip_under_load` baseline-parity.            |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-input.md` — 2026-04-30.
