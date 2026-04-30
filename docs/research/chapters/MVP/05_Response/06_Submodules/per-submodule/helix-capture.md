# `helix-capture` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-capture`                                                                                                        |
| **Origin chapter:section**  | [C28 §6](../../05_Video_Audio/03_Capture_Pipelines.md) — *Per-OS capture (DXGI / Metal / X11 / PipeWire)*             |
| **Public path (4 mirrors)** | `vasic-digital/helix-capture` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-shm`                                                                                  |
| **External Go deps**        | DXGI cgo (Windows), Metal cgo (macOS), X11/XRandR (Linux), PipeWire client (Linux Wayland)                              |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `os-capture-1.x` — builder `golang-builder-cgo` (+ Win/Mac variants), runtime `distroless-cc` (+ 3 variants per S02 §4)|
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/01_minimum_viable_session/scenarios/05_capture_per_os_at_session_start.scenario.yaml`                  |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2** (depends on helix-shm)                                                                                       |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-capture` is the **OS-level frame-grab primitive**. Origin: [C28 §6](../../05_Video_Audio/03_Capture_Pipelines.md). The submodule wraps each OS's lowest-latency capture API — DXGI Desktop Duplication on Windows, AVCaptureScreenInput / Metal Texture Sharing on macOS, X11 SHM + XRandR + DRM on Linux X11, and PipeWire screencast on Linux Wayland — and writes captured frames into `helix-shm` Pages without an intermediate copy.

The submodule was introduced because per-OS capture APIs drift independently and each requires distinct error handling. Consolidation gives one canonical Sunshine++ capture surface (per Insight #1 of the Architecture family) that all client-side variants reuse.

`★ Multi-runtime variant.` Per [S02 §4](../02_Containers_Submodule.md#4-multi-arch-image-strategy-linuxamd64--linuxarm64-default), `helix-capture` is the **only** submodule that ships four runtime variants: linux (X11 + Wayland combined), windows (Server Core + DXGI), darwin (Apple Silicon + amd64 with osxcross), and wayland (PipeWire-only). The four-variant fan-out reflects that capture APIs cannot share a single binary.

---

## 2. Public API Surface

### 2.1 The `Capturer` interface

```go
package capture

// Capturer is implemented per-OS; consumers depend on the
// interface rather than the concrete type.
type Capturer interface {
    Start(targetFPS int, sink Sink) error
    Stop() error
    Capabilities() Capabilities
    OS() OperatingSystem
}
```

### 2.2 Per-OS factories

```go
package capture

func NewDXGI(adapter int, output int, opts ...Option) (Capturer, error)
func NewMetal(displayID int, opts ...Option) (Capturer, error)
func NewX11SHM(display string, opts ...Option) (Capturer, error)
func NewWayland(opts ...Option) (Capturer, error)  // PipeWire screencast portal
func NewBest(opts ...Option) (Capturer, error)     // auto-pick
```

### 2.3 The `Sink` interface

```go
package capture

// Sink receives captured frames as helix-shm Pages.
type Sink interface {
    Submit(frame *shm.Page, meta FrameMeta) error
}

type FrameMeta struct {
    PresentTimeNs int64
    PixelFormat   PixelFormat   // NV12 default; some OS APIs return RGBA which gets converted
    Width, Height int
    Stride        int
    DamageRegions []Rect        // dirty-region hints, where supported
}
```

### 2.4 Configuration options

```go
package capture

type Option func(*config)

func WithTargetFPS(fps int) Option
func WithRegion(rect Rect) Option              // sub-display capture
func WithCursorOverlay(enabled bool) Option
func WithDamageHints() Option                  // ask the OS for dirty regions
```

### 2.5 Capability detection

```go
package capture

func DetectCapabilities() (Capabilities, error)
type Capabilities struct {
    OS                 OperatingSystem
    APIs               []API   // DXGI, Metal, X11, PipeWire
    DamageRegionsHint  bool
    HardwareEncoderShare bool  // DXGI desktop -> NVENC zero-copy without a CPU hop
    HDRCapture         bool    // DXGI 1.6 HDR
    MultiDisplay       bool
    LowLatencyHint     bool    // DXGI present_immediately, etc.
}
```

### 2.6 Statistics

```go
package capture

type CapturerStats struct {
    FramesCaptured uint64
    DroppedFrames  uint64
    ConvertCount   uint64    // pixel-format conversions performed
    AverageFPS     float64
    CursorOverlayCount uint64
}
```

### 2.7 Cursor + overlay helpers

```go
package capture

// CursorOverlay returns the OS-side cursor sprite + position; the
// consumer composites onto the captured frame for client-side
// visibility.
type CursorOverlay struct {
    Sprite      []byte    // RGBA cursor sprite
    Position    Point
    HotSpot     Point
    Visible     bool
}
func (c Capturer) CursorOverlay() (*CursorOverlay, error)
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`xrandr` / `wlr-randr` for display enumeration; `swaymsg`/`hyprctl` for compositor probing).
- `helix-shm` — captured frames land in `helix-shm` Pages.

### 3.2 External (Go)

- DXGI cgo bindings (Windows-only build).
- Metal cgo (macOS-only build via osxcross).
- X11 + libXrandr + libXShm cgo (Linux X11 variant).
- PipeWire client cgo (Linux Wayland variant).

### 3.3 External (system)

- Per-OS prerequisites:
  - **Windows**: DXGI 1.5+ (Windows 10+).
  - **macOS**: macOS 14+ for the screen-recording entitlement.
  - **Linux X11**: X11 + DAMAGE extension + SHM extension.
  - **Linux Wayland**: PipeWire ≥ 1.0 + xdg-desktop-portal screencast portal.

---

## 4. Container Build (S02 §3 lane: `os-capture-1.x`, four runtime variants)

Per [S02 §4](../02_Containers_Submodule.md#4-multi-arch-image-strategy-linuxamd64--linuxarm64-default) the lane builds four images:

- `vasic-digital/helix-capture-linux:<tag>` — X11 + Wayland combined; multi-arch (amd64 + arm64), distroless-cc.
- `vasic-digital/helix-capture-wayland:<tag>` — Wayland-only with PipeWire deps; multi-arch.
- `vasic-digital/helix-capture-windows:<tag>` — Windows Server Core base; amd64 only.
- `vasic-digital/helix-capture-darwin:<tag>` — multi-arch macOS; not distroless (Apple has no distroless equivalent).

Hardening: standard S02 §8.2 + Linux variants need `--device=/dev/dri/card0` (DRM access) + `--volume /run/user/1000/pipewire-0:/run/pipewire-0` (PipeWire socket).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Per-OS Capturer construction + Capabilities reporting with mock display handles.

### 5.2 Integration
Real Xvfb display (Linux); verify a frame emerges with expected dimensions.

### 5.3 E2E
helix-capture → helix-shm → helix-encoder → helix-pipeline.

### 5.4 Security
govulncheck + Snyk + Trivy. Per-OS-specific: Windows DXGI HMODULE injection rejection; Wayland portal-permission denial graceful handling.

### 5.5 Benchmarking
DXGI capture 4K120 p999 ≤ 4 ms; Metal capture 4K60 p999 ≤ 6 ms; X11SHM 1080p60 p999 ≤ 4 ms.

### 5.6 Chaos
Display-disconnect + display-mode-change events; verify graceful re-init.

### 5.7 Stress
24-hour 4K120 capture; zero leak; AverageFPS within 0.5 % of target.

### 5.8 Smoke
30-second post-deploy: capture 1 frame; verify dimensions match display.

### 5.9 Full Automation
§5.1–§5.8 across all four runtime variants.

### 5.10 Challenges
`01_minimum_viable_session/05_capture_per_os_at_session_start.scenario` — full session start with each OS variant (parallel runs); verify capture starts within 200 ms.

---

## 6. Challenges Entry-Point (S03 §4 row #20)

**Topology:** `01_minimum_viable_session`. **Scenario:** `05_capture_per_os_at_session_start.scenario.yaml`. **Why this scenario.** Capture init time directly affects the user-perceived "click play → see game" latency; the scenario verifies init stays within 200 ms across all four OS variants. **Baseline:** capture-start p99 ≤ 150 ms; first-frame latency p99 ≤ 200 ms; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-capture` imports `helix-r18-safeexec` for boundary subprocess invocations (display enumeration). Hot path is cgo to OS APIs. Linux variants device-passthrough (DRM) per S02 §8.4 carve-out shape.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on per-OS variant green CI; Windows + Linux paths reach `v1.0.0` first, macOS + Wayland following as Apple Silicon CI runners + portal stability mature.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                                                       | Purpose                                                                |
|------------------------------------|---------------|--------------------------------------------------------------------|------------------------------------------------------------------------|
| `HELIX_CAPTURE_OS`                 | `auto`        | `windows`/`darwin`/`linux-x11`/`linux-wayland`/`auto`              | Force OS-specific capturer.                                             |
| `HELIX_CAPTURE_DISPLAY_ID`         | `0`           | int                                                                 | Display index to capture.                                              |
| `HELIX_CAPTURE_TARGET_FPS`         | `120`         | int [30, 240]                                                       | Target capture frame rate.                                             |
| `HELIX_CAPTURE_DAMAGE_HINTS`       | `true`        | bool                                                                | Use OS damage-region hints to skip unchanged frames.                  |
| `HELIX_CAPTURE_CURSOR_OVERLAY`     | `true`        | bool                                                                | Composite cursor onto captured frames.                                |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| DXGI capture 4K @ 120 Hz              | 3 ms     | 3.5 ms  | 4 ms    | Desktop Duplication zero-copy.                                   |
| Metal capture 4K @ 60 Hz              | 4 ms     | 5 ms    | 6 ms    | Apple Silicon GPU shared texture.                                |
| X11 SHM 1080p @ 60 Hz                 | 3 ms     | 3.5 ms  | 4 ms    | XShmGetImage path.                                                |
| PipeWire 4K @ 60 Hz                   | 5 ms     | 7 ms    | 10 ms   | Portal-mediated; one extra copy.                                 |
| Capturer.Start init                    | 50 ms    | 100 ms  | 150 ms  | Cold init; warm reuse via Pool.                                   |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `capture: ErrPortalDenied`                    | Wayland user denied screencast portal                          | Surface to UI; require user re-grant.                                     |
| `capture: ErrDisplayDisconnected`             | HDMI cable removed, display sleep                              | Stop + Restart; use Capabilities.MultiDisplay for fallback.               |
| `capture: ErrDXGIAccessDenied`                | UAC / Secure Desktop active                                    | Document operator constraint; cannot capture across UAC prompts.          |
| `capture: ErrShmTooLarge`                     | Capture size exceeds /dev/shm capacity                         | Increase shm-size: `--shm-size=1g`.                                       |
| `capture: ErrPixelFormatUnsupported`          | OS API returns unexpected format                                | Submodule converts to NV12; metric the conversion frequency.              |

### 9.4 Migration from OBS / FFmpeg capture

A consumer migrating from `obs --capture --output ...` or `ffmpeg -f x11grab ...`:

1. Replace the subprocess (R-18 forbids long-running ffmpeg/obs subprocess) with `cap, _ := capture.NewBest()`.
2. Wire the Sink to `helix-shm.Pool.Acquire` / `helix-encoder.EncodeFrame`.
3. Add OTLP span around `Submit`; metric for `CapturerStats.DroppedFrames`.

The migration is documented in `docs/migration-from-obs-ffmpeg.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_capture_frames_captured_total`           | counter    | Frames captured, labelled `os`, `api`.                                       |
| `helix_capture_dropped_total`                   | counter    | Frames dropped due to backpressure or stale data.                           |
| `helix_capture_convert_total`                   | counter    | Pixel-format conversions performed.                                          |
| `helix_capture_average_fps`                     | gauge      | Observed average FPS.                                                        |
| `helix_capture_init_latency_seconds`            | histogram  | Capturer.Start latency.                                                      |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C28 §6](../../05_Video_Audio/03_Capture_Pipelines.md) — origin                            | Origin chapter; full Capturer interface.                                             |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates capture → encode handoff.                                      |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-capture-A        | Wayland portal stability — production-ready across compositor matrix?                                         | C28 §6 next revision                                |
| OQ-capture-B        | macOS screen-recording entitlement — managed at submodule or operator level?                                  | C28 §6 next revision                                |
| OQ-capture-C        | DXGI HDR capture — full DolbyVision passthrough or only HDR10?                                                 | C32 §6 + C28 §6 next revisions                      |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/03_Capture_Pipelines.md`](../../05_Video_Audio/03_Capture_Pipelines.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-shm.md`](helix-shm.md) | (this batch) | 2026-04-30 | direct dependency                              |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (per-OS Capturer + Capabilities).                                  |
| Integration    | Xvfb (Linux); per-OS smoke tests in CI for Windows + macOS variants.                          |
| E2E            | Full capture → shm → encoder → pipeline.                                                      |
| Security       | govulncheck + Snyk + Trivy + per-OS denial fuzzers.                                           |
| Benchmarking   | All §9.2 budgets met across the four variants.                                                |
| Chaos          | Display-disconnect + mode-change graceful re-init.                                            |
| Stress         | 24-hour 4K120; zero leak; AverageFPS ±0.5 %.                                                  |
| Smoke          | 30-second 1-frame capture + dimensions check.                                                 |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped (per-variant matrix axis).         |
| Challenges     | `01_minimum_viable_session/05_capture_per_os_at_session_start` baseline-parity.               |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-capture.md` — 2026-04-30.
