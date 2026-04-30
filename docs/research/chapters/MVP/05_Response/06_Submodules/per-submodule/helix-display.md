# `helix-display` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-display`                                                                                                        |
| **Origin chapter:section**  | [C22 §6](../../04_Latency/08_Frame_Pacing_and_VRR.md) — *Frame pacing + VRR (G-SYNC, FreeSync, HDMI VRR, ALLM)*       |
| **Public path (4 mirrors)** | `vasic-digital/helix-display` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `golang.org/x/sys/unix`, vendor-specific cgo bindings (NVIDIA NVAPI / AMD ADL / Intel Display API)                    |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `frame-pacing-vrr-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                      |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/06_4k_120hz_hdr_dolby_vision/scenarios/02_vrr_under_burst_frame_load.scenario.yaml`                     |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-display` is the **client-side frame-pacing + VRR negotiation primitive** for HelixPlay. It coordinates the receiver-side decode + present pipeline with the display's variable-refresh-rate (VRR) capability — G-SYNC, FreeSync Premium, HDMI 2.1 VRR, and ALLM (Auto Low-Latency Mode). Origin: [C22 §6](../../04_Latency/08_Frame_Pacing_and_VRR.md). The submodule's role is to ensure the present time of each frame matches the encoder's intent rather than the local frame-pacer's preference.

The submodule was introduced because frame-pacing logic spans NVIDIA / AMD / Intel / Steam Deck / Apple Silicon vendors with different APIs; consolidation reduces 5-platform drift.

---

## 2. Public API Surface

### 2.1 The `Pacer` type

```go
package display

// Pacer schedules frame presentation. Wraps the platform-specific
// vsync / VRR negotiation in a unified Go-side API.
type Pacer struct { /* ... */ }

func NewPacer(opts ...Option) (*Pacer, error)
func (p *Pacer) Present(frame Frame, presentTime time.Time) error
func (p *Pacer) RefreshRate() float64           // current Hz
func (p *Pacer) VRRActive() bool
func (p *Pacer) Close() error
```

### 2.2 VRR negotiation

```go
package display

// NegotiateVRR attempts to put the display into VRR mode using
// vendor-specific APIs (NVAPI / ADL / DRM, depending on the OS +
// GPU vendor).
func NegotiateVRR(monitor MonitorID, opts ...VRROption) (VRRSession, error)

type VRRSession struct { /* ... */ }
func (s *VRRSession) MinHz() float64
func (s *VRRSession) MaxHz() float64
func (s *VRRSession) ALLMActive() bool
func (s *VRRSession) Close() error
```

### 2.3 The `Frame` type

```go
package display

type Frame struct {
    Pixels       []byte         // decoded YUV or RGBA buffer
    Format       PixelFormat    // FormatNV12, FormatYV12, FormatRGBA8
    Width        int
    Height       int
    Stride       int
    PresentTime  time.Time      // intended present time
    EncodeID     uint64         // correlated with encoder's frame ID for OTLP tracing
}
```

### 2.4 ALLM (Auto Low-Latency Mode)

```go
package display

// EnableALLM tells the display to use its lowest-latency
// post-processing path. Required for HDMI 2.1; signalled via the
// HDMI InfoFrame.
func EnableALLM(monitor MonitorID) error

// DisableALLM restores the display's default post-processing.
func DisableALLM(monitor MonitorID) error
```

### 2.5 Capability detection

```go
package display

func DetectCapabilities() (Capabilities, error)
type Capabilities struct {
    GSyncSupported       bool
    FreeSyncSupported    bool
    HDMIVRRSupported     bool
    ALLMSupported        bool
    DolbyVisionLowLatency bool
    EDIDExtensionRead    bool   // EDID extension block readable
}
```

### 2.6 Configuration options

```go
package display

type Option func(*config)

func WithMonitor(id MonitorID) Option
func WithMinHz(hz float64) Option         // VRR min-Hz override
func WithMaxHz(hz float64) Option
func WithLowFramerateCompensation() Option // LFC for FreeSync
func WithFrameDoubling() Option            // for sub-min-Hz frames
```

### 2.7 Pacer statistics

```go
package display

type PacerStats struct {
    FramesPresented   uint64
    DroppedLate       uint64    // missed present deadline
    DroppedDup        uint64    // duplicate present (frame-doubling)
    PresentJitter     time.Duration  // EWMA of present-time deviation
    VRRSwitchEvents   uint64
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`xrandr` / `wlr-randr` for monitor enumeration; `nvidia-settings -q` for VRR status).

### 3.2 External (Go)

- `golang.org/x/sys/unix` — Linux DRM ioctls.
- Vendor cgo bindings:
  - NVIDIA NVAPI (Windows / Linux) for G-SYNC.
  - AMD ADL (Windows) / DRM (Linux) for FreeSync.
  - Intel Display API for Arc + integrated GPUs.

### 3.3 External (system)

- DisplayPort 1.2a+ or HDMI 2.1 link; the submodule reads EDID extension blocks to verify VRR support.
- For wayland clients: a compositor with `zwp_presentation_time_v1` protocol (KDE Plasma 5.27+, GNOME 45+, sway 1.9+).

---

## 4. Container Build (S02 §3 lane: `frame-pacing-vrr-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + bind-mount `/dev/dri` for DRM access (per-device passthrough, not `--privileged`).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Pacer scheduling logic with synthetic vsync clock.

### 5.2 Integration
Real DRM ioctl on a Linux host with a known monitor; verify VRR negotiation succeeds.

### 5.3 E2E
Decoded frame → Pacer → real display; verify present-time accuracy via PresentMon-equivalent.

### 5.4 Security
govulncheck + Snyk + Trivy.

### 5.5 Benchmarking
Pacer schedule decision p999 ≤ 50 µs; present-time jitter EWMA ≤ 100 µs.

### 5.6 Chaos
Inject vsync jitter; verify the pacer absorbs short-burst jitter without dropping frames.

### 5.7 Stress
24-hour 4K120 frame stream; zero leak; jitter EWMA stable.

### 5.8 Smoke
30-second post-deploy: read EDID; report VRR support; report current refresh rate.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`06_4k_120hz_hdr_dolby_vision/02_vrr_under_burst_frame_load.scenario` — VRR active during burst frame-rate variations (60–144 Hz oscillation); verify no judder, no tear, no dropped frames.

---

## 6. Challenges Entry-Point (S03 §4 row #14)

**Topology:** `06_4k_120hz_hdr_dolby_vision`. **Scenario:** `02_vrr_under_burst_frame_load.scenario.yaml`. **Why this scenario.** VRR's value prop is realised under variable-rate workloads; the burst scenario stresses the pacer's ability to track encoder cadence. **Baseline:** zero observed judder (per-frame present-time deviation < 1 ms p999); frame-doubling rate ≤ 0.1 %; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-display` imports `helix-r18-safeexec` for boundary subprocess invocations (xrandr / nvidia-settings diagnostics). Hot path (DRM ioctls + cgo) is pure syscall + cgo. Device-passthrough for `/dev/dri` follows S02 §8.4 carve-out shape.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on per-vendor API stability; G-SYNC / FreeSync paths reach `v1.0.0` first, with HDMI 2.1 VRR following as TV-target deployments come online.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_DISPLAY_MONITOR`            | `auto`        | string                  | Monitor identifier; `auto` picks the primary display.                  |
| `HELIX_DISPLAY_VRR_MIN_HZ`         | `48`          | float [24, 144]         | VRR minimum range.                                                      |
| `HELIX_DISPLAY_VRR_MAX_HZ`         | `120`         | float [48, 240]         | VRR maximum range.                                                      |
| `HELIX_DISPLAY_LFC_ENABLED`        | `true`        | bool                    | Low-Framerate Compensation: double frames when below MinHz.             |
| `HELIX_DISPLAY_ALLM_ENABLED`       | `true`        | bool                    | Auto Low-Latency Mode (HDMI 2.1).                                       |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Pacer schedule decision               | 10 µs    | 30 µs   | 50 µs   | Deadline check + cgo call to vendor API.                         |
| VRR negotiation                       | 100 ms   | 300 ms  | 500 ms  | One-time at session start.                                       |
| ALLM enable                           | 50 ms    | 100 ms  | 200 ms  | HDMI InfoFrame transmission.                                     |
| Present jitter (EWMA)                 | 30 µs    | 70 µs   | 100 µs  | Steady-state with healthy VRR link.                              |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `display: ErrVRRNotSupported`                 | Display does not advertise VRR via EDID                       | Verify DisplayPort 1.2a+ / HDMI 2.1; check EDID extension blocks.         |
| `display: ErrEDIDUnreadable`                  | DRM-master not held; EDID extension blocked                   | Run with `--device=/dev/dri/card0`; check udev permissions.               |
| `display: ErrALLMNotSignalled`                | HDMI InfoFrame transmission failed                            | Verify HDMI 2.1 cable; check display firmware.                            |
| `display: ErrLFCRequired`                      | Frame rate dropped below MinHz; LFC disabled                  | Enable `HELIX_DISPLAY_LFC_ENABLED` or raise floor frame rate.            |
| `display: ErrPaceMisalignment`                | Pacer's clock drifted from compositor's vsync                 | Verify wayland presentation-time-v1 protocol presence.                    |

### 9.4 Migration from default present path

A consumer migrating from default vsync-locked present:

1. Replace `eglSwapBuffers` / `Present(0, ...)` with `pacer.Present(frame, presentTime)`.
2. Compute `presentTime` from encoder's intended present time (carried over via `helix-grpc-frame.FrameMetadata`).
3. Negotiate VRR at session start: `vrr, _ := display.NegotiateVRR(monitor)`.
4. Enable ALLM where supported: `display.EnableALLM(monitor)`.
5. Add OTLP span around `Present`; metric for `PacerStats.PresentJitter` (alert: > 100 µs EWMA).

The migration is documented in `docs/migration-from-default-present.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_display_frames_presented_total`          | counter    | Frames presented to the display.                                             |
| `helix_display_dropped_late_total`              | counter    | Frames dropped due to missed present deadline.                              |
| `helix_display_dropped_dup_total`               | counter    | Frames duplicated for LFC.                                                   |
| `helix_display_present_jitter_seconds`          | gauge      | EWMA of present-time jitter.                                                 |
| `helix_display_vrr_active`                      | gauge      | 1 if VRR active, 0 if not.                                                   |
| `helix_display_allm_active`                     | gauge      | 1 if ALLM active.                                                            |
| `helix_display_refresh_rate_hz`                 | gauge      | Current display refresh rate.                                                |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C22 §6](../../04_Latency/08_Frame_Pacing_and_VRR.md) — origin                             | Origin chapter; full Pacer / VRR / ALLM API.                                         |
| [C04 §6](../../03_Architecture/04_Go_Client_Ecosystem.md) — Go Client Ecosystem            | Wails / Compose-for-TV / Steam Deck client present path.                            |
| [C32 §6](../../05_Video_Audio/07_HDR_and_Color.md) — HDR                                  | HDR present path coordinates with Dolby Vision low-latency negotiation.              |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-display-A        | Wayland presentation-time-v1 — required for V1 graduation or optional?                                        | C22 §6 next revision                                |
| OQ-display-B        | Apple Silicon ProMotion — does the submodule own integration or delegate to Compose Multiplatform?           | C04 §6 + C22 §6 next revision                       |
| OQ-display-C        | LFC implementation — frame-doubling vs frame-tripling fallback for sub-30 Hz?                                  | C22 §6 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/08_Frame_Pacing_and_VRR.md`](../../04_Latency/08_Frame_Pacing_and_VRR.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #14                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Pacer + VRR + ALLM + Capabilities).                               |
| Integration    | Real DRM ioctl; VRR negotiation success.                                                      |
| E2E            | Decoded frame → Pacer → display; PresentMon-style verification.                              |
| Security       | govulncheck + Snyk + Trivy.                                                                   |
| Benchmarking   | All §9.2 budgets met; jitter ≤ 100 µs EWMA.                                                   |
| Chaos          | Vsync-jitter injection; pacer absorbs short bursts.                                           |
| Stress         | 24-hour 4K120; jitter stable.                                                                 |
| Smoke          | 30-second EDID + refresh-rate report.                                                         |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `06_4k_120hz_hdr_dolby_vision/02_vrr_under_burst_frame_load` baseline-parity.                 |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-display.md` — 2026-04-30.
