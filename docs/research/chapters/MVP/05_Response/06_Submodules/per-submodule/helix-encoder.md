# `helix-encoder` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-encoder`                                                                                                        |
| **Origin chapter:section**  | [C27 §6](../../05_Video_Audio/02_Hardware_Encoders.md) — *Vendor encoder wrappers (NVENC / QSV / AMF / VideoToolbox / V4L2)* |
| **Public path (4 mirrors)** | `vasic-digital/helix-encoder` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-codec`                                                                                |
| **External Go deps**        | NVIDIA NVENC headers, Intel QSV (libmfx), AMD AMF, Apple VideoToolbox, Linux V4L2 — all via cgo                          |
| **Licence (S01 §4.8)**      | **Apache-2.0** (encoder hardware-acceleration patents)                                                                  |
| **Container CI lane (S02 §3)** | `vendor-encoder-1.x` — builder `golang-builder-gpu`, runtime `distroless-cuda`                                      |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/06_4k_120hz_hdr_dolby_vision/scenarios/04_nvenc_qsv_amf_videotoolbox_parity.scenario.yaml`              |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2** (depends on helix-codec)                                                                                     |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-encoder` exposes a **vendor-agnostic encoder API** that wraps the five hardware encoder ecosystems HelixPlay targets: NVIDIA NVENC, Intel Quick Sync (QSV), AMD AMF, Apple VideoToolbox, and Linux V4L2 stateful encoders. Origin: [C27 §6](../../05_Video_Audio/02_Hardware_Encoders.md). The submodule consumes a `helix-codec.Config` and emits H.264 / HEVC / AV1 NAL units; the consumer (typically `helix-pipeline`) is unaware of which vendor is doing the work.

The submodule was introduced because per-vendor encoder integration drifts in subtle ways (NVENC's `nvEncEncodePicture` signature changed in driver 535+; AMF's session lifecycle differs from QSV's; VideoToolbox's frame-pacing assumptions break under burst load). Consolidation prevents per-pipeline drift and gives a single place to patch vendor-specific behaviours.

`★ Apache-2.0 rationale.` Codec hardware-acceleration code may practice patentable algorithms (NVENC + AMF + QSV all carry their own patent footprint; AV1 hardware-encoder support specifically is covered under AOMedia's patent commitments). Apache-2.0 §3 patent grant gives downstream consumers explicit permission per [S01 §4.8.2](../01_Submodule_Catalog.md#482-why-apache-20-specifically-for-those-four).

---

## 2. Public API Surface

### 2.1 The `Encoder` interface

```go
package encoder

// Encoder is implemented per-vendor; consumers depend on the
// interface, not the concrete type.
type Encoder interface {
    Configure(cfg *codec.Config) error
    EncodeFrame(input *Frame) (*EncodedFrame, error)
    Flush() error
    Close() error
    Vendor() Vendor
    Capabilities() Capabilities
}
```

### 2.2 Vendor-specific factories

```go
package encoder

func NewNVENC(opts ...Option) (Encoder, error)
func NewQSV(opts ...Option) (Encoder, error)
func NewAMF(opts ...Option) (Encoder, error)
func NewVideoToolbox(opts ...Option) (Encoder, error)
func NewV4L2(devicePath string, opts ...Option) (Encoder, error)
func NewBest(cfg *codec.Config) (Encoder, error)  // auto-pick by capability
```

### 2.3 The `Frame` and `EncodedFrame` types

```go
package encoder

type Frame struct {
    Plane         []byte           // NV12 / I420 plane data
    PixelFormat   PixelFormat
    Width, Height int
    Stride        int
    PresentTimeNs int64            // input PTS
}

type EncodedFrame struct {
    NALUnits  [][]byte    // emitted NAL units (H.264) or OBUs (AV1)
    KeyFrame  bool
    PTS       int64       // output PTS
    QP        int         // chosen quantisation parameter
    Bitrate   int         // observed output bitrate (last-N-frame EWMA)
}
```

### 2.4 Vendor capabilities

```go
package encoder

func DetectVendors() ([]Vendor, error)
type Vendor int
const (
    VendorNVENC Vendor = iota
    VendorQSV
    VendorAMF
    VendorVideoToolbox
    VendorV4L2
)

type Capabilities struct {
    H264, HEVC, AV1 bool
    Max4K120        bool
    Max8K60         bool
    HDR10, HDR10Plus, DolbyVision bool
    LowLatencyMode  bool   // NVENC's "ultra-low-latency"; QSV's "TargetUsage=7"
}
```

### 2.5 Configuration options

```go
package encoder

type Option func(*config)

func WithDeviceID(id int) Option
func WithGOPSize(n int) Option
func WithMaxBitrate(bps int) Option
func WithLowLatencyMode() Option
func WithLookahead(frames int) Option
```

### 2.6 Statistics

```go
package encoder

type EncoderStats struct {
    FramesEncoded uint64
    KeyFrames     uint64
    DroppedFrames uint64
    AverageQP     float64
    BitrateBps    int
    LastError     error
}
```

### 2.7 The `Pool` of encoder instances

```go
package encoder

// Pool reuses encoder instances across sessions; particularly
// relevant for NVENC where session creation is expensive.
type Pool struct { /* ... */ }
func NewPool(factory func() (Encoder, error), capacity int) *Pool
func (p *Pool) Acquire() (Encoder, error)
func (p *Pool) Release(e Encoder)
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`nvidia-smi -q` / `vainfo` for capability detection).
- `helix-codec` — consumes `codec.Config` for encoder configuration.

### 3.2 External (Go)

- NVIDIA NVENC headers (cgo): pinned to NVIDIA driver ≥ 560.
- Intel libmfx (QSV): pinned to ≥ v25.0.0.
- AMD AMF: pinned to AMF SDK ≥ 1.4.36.
- Apple VideoToolbox: macOS ≥ 14.0 system framework.
- Linux V4L2: kernel ≥ 5.4 v4l2-stateful-encoder API.

### 3.3 External (system)

- GPU drivers (NVIDIA / AMD / Intel) installed at the operator host level.
- Container device-passthrough per S02 §8.4 carve-out: `--device=/dev/nvidia*` for NVENC; `--device=/dev/dri/renderD128` for QSV + AMF on Linux; macOS encoder runs only on a Mac host (CI runs on Apple Silicon CI runners).

---

## 4. Container Build (S02 §3 lane: `vendor-encoder-1.x`)

**Builder:** `golang-builder-gpu`. **Runtime:** `distroless-cuda`. **Multi-arch:** `linux/amd64` (GPU drivers); `linux/arm64` for Tegra-class deployments (defer per OQ-S02-C). **Hardening:** standard S02 §8.2 + device-passthrough per S02 §8.4 carve-out (NOT `--privileged`).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Per-vendor configuration round-trips with mock device handles.

### 5.2 Integration
Real vendor encoder; encode 100 frames of synthetic NV12 input; verify NAL units parse with FFmpeg.

### 5.3 E2E
helix-encoder → helix-pipeline → helix-transport.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: malformed-Frame fuzzer.

### 5.5 Benchmarking
NVENC encode 4K120 p999 ≤ 8 ms / frame; QSV encode 4K60 p999 ≤ 12 ms / frame.

### 5.6 Chaos
Inject GPU memory pressure; verify encoder surfaces OOM cleanly without crashing.

### 5.7 Stress
24-hour 4K120 encode; zero session leak; AverageQP stable.

### 5.8 Smoke
30-second post-deploy: encode 1 frame; verify NAL unit emerges.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`06_4k_120hz_hdr_dolby_vision/04_nvenc_qsv_amf_videotoolbox_parity.scenario` — same input, all four (or available) vendors; verify VMAF parity ≥ 95.0 across vendors.

---

## 6. Challenges Entry-Point (S03 §4 row #19)

**Topology:** `06_4k_120hz_hdr_dolby_vision`. **Scenario:** `04_nvenc_qsv_amf_videotoolbox_parity.scenario.yaml`. **Why this scenario.** Cross-vendor parity is the strongest signal that the submodule's abstraction is leak-free. **Baseline:** VMAF ≥ 95.0 between any two vendors on the same input; bitrate within ±5 % of target across all four.

---

## 7. R-18 Inheritance

`helix-encoder` imports `helix-r18-safeexec` for boundary subprocess invocations (vendor diagnostic tools). Hot path is cgo to vendor SDKs. Device-passthrough carve-out (S02 §8.4) is the canonical R-18 §11.5.3 named exception.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on cross-vendor parity (Challenges scenario), per-vendor SDK stability (NVENC, QSV, AMF all release annually), and Apache-2.0 SPDX header compliance.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                                                                  | Purpose                                                                |
|------------------------------------|---------------|------------------------------------------------------------------------------|------------------------------------------------------------------------|
| `HELIX_ENCODER_VENDOR`             | `auto`        | `nvenc`/`qsv`/`amf`/`videotoolbox`/`v4l2`/`auto`                             | Vendor; `auto` picks by capability.                                    |
| `HELIX_ENCODER_DEVICE_ID`          | `0`           | int                                                                          | GPU device index.                                                       |
| `HELIX_ENCODER_LOW_LATENCY`        | `true`        | bool                                                                         | Enable vendor-specific low-latency mode.                              |
| `HELIX_ENCODER_LOOKAHEAD_FRAMES`   | `0`           | int [0, 16]                                                                  | Lookahead for rate control; 0 = disabled (lowest latency).            |
| `HELIX_ENCODER_POOL_CAPACITY`      | `4`           | int [1, 16]                                                                  | Encoder-instance pool capacity.                                        |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| NVENC encode 4K120 H.264             | 5 ms     | 7 ms    | 8 ms    | Latest GeForce / Quadro hardware.                                |
| NVENC encode 4K120 HEVC              | 5 ms     | 7 ms    | 8 ms    | Same hardware; HEVC similar latency.                             |
| QSV encode 4K60 HEVC                 | 8 ms     | 11 ms   | 12 ms   | Iris Xe / Arc.                                                    |
| AMF encode 4K60 HEVC                 | 7 ms     | 10 ms   | 12 ms   | RDNA 3 hardware.                                                  |
| Encoder pool acquire (warm)           | 5 µs     | 12 µs   | 20 µs   | Reuses existing session.                                          |
| Encoder pool acquire (cold)           | 50 ms    | 100 ms  | 200 ms  | NVENC session creation; one-time per pool grow.                  |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `encoder: ErrVendorNotAvailable`              | Device passthrough missing                                     | Add `--device=/dev/nvidia*` etc.                                          |
| `encoder: ErrSessionLimitReached`             | NVENC concurrent-session limit (consumer-grade caps at 5)      | Pool re-use; document operator's GPU SKU.                                  |
| `encoder: ErrUnsupportedProfile`              | Codec / resolution combination not in vendor capabilities      | Fall back to a supported profile via codec.PickProfile.                   |
| `encoder: ErrDriverMismatch`                  | Compiled against newer NVENC headers than runtime              | Rebuild against the runtime driver's headers.                              |
| `encoder: ErrFrameDropped`                    | Encoder fell behind; backpressure to consumer not applied     | Investigate input rate; ensure the consumer applies backpressure.          |

### 9.4 Migration from inline FFmpeg encoders

A consumer migrating from `ffmpeg -c:v h264_nvenc ...` invocations:

1. Replace the FFmpeg subprocess (R-18 forbids the long-running ffmpeg subprocess anyway) with `enc, _ := encoder.NewBest(cfg)`.
2. Pump frames: `enc.EncodeFrame(frame)`.
3. Forward NAL units to `helix-transport` via `helix-pipeline`.
4. Add OTLP span around `EncodeFrame`; metric for `EncoderStats.AverageQP`.

The migration is documented in `docs/migration-from-ffmpeg-subprocess.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_encoder_frames_encoded_total`            | counter    | Frames encoded, labelled `vendor`, `codec`.                                  |
| `helix_encoder_keyframes_total`                 | counter    | Keyframes emitted.                                                           |
| `helix_encoder_dropped_total`                   | counter    | Frames dropped due to backpressure.                                          |
| `helix_encoder_qp_average`                      | gauge      | Average QP per session.                                                      |
| `helix_encoder_bitrate_bps`                     | gauge      | Observed bitrate.                                                            |
| `helix_encoder_session_count`                   | gauge      | Active encoder sessions.                                                     |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C27 §6](../../05_Video_Audio/02_Hardware_Encoders.md) — origin                            | Origin chapter; full Encoder interface.                                              |
| [C29 §6](../../05_Video_Audio/04_DualPath_Encoding.md) — Dual-Path                        | helix-dualpath consumes EncodedFrame for stream + record split.                     |
| [C30 §6](../../05_Video_Audio/05_Recording_Storage.md) — Recording                        | helix-record consumes EncodedFrame for fMP4/MKV emission.                           |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates encoder lifecycle.                                             |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-encoder-A        | NVENC consumer-card session limit — bypass via patched driver acceptable for production?                      | C27 §6 next revision                                |
| OQ-encoder-B        | macOS Apple Silicon encoder integration — VideoToolbox stable for streaming workloads?                        | C27 §6 next revision                                |
| OQ-encoder-C        | V4L2 stateful encoder — RPi5 / Jetson coverage?                                                                | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/02_Hardware_Encoders.md`](../../05_Video_Audio/02_Hardware_Encoders.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-codec.md`](helix-codec.md) | (this batch) | 2026-04-30 | direct dependency                              |
| [`../02_Containers_Submodule.md`](../02_Containers_Submodule.md) §8.4 |   627 | 2026-04-30 | device-passthrough carve-out                   |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (per-vendor configure + EncodeFrame).                              |
| Integration    | Real per-vendor encoder; FFmpeg-parseable NAL output.                                         |
| E2E            | helix-encoder → helix-pipeline → helix-transport.                                             |
| Security       | govulncheck + Snyk + Trivy + Frame fuzzer.                                                   |
| Benchmarking   | NVENC 4K120 H.264 / HEVC ≤ 8 ms p999.                                                        |
| Chaos          | GPU memory pressure; OOM surfaced cleanly.                                                    |
| Stress         | 24-hour 4K120; zero leak; QP stable.                                                          |
| Smoke          | 30-second 1-frame encode + NAL verification.                                                  |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `06_4k_120hz_hdr_dolby_vision/04_nvenc_qsv_amf_videotoolbox_parity` baseline-parity.          |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-encoder.md` — 2026-04-30.
