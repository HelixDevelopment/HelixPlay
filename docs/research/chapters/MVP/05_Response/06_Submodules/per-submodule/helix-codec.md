# `helix-codec` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-codec`                                                                                                          |
| **Origin chapter:section**  | [C26 §6](../../05_Video_Audio/01_Codec_Selection.md) — *Codec ladder + GOP cadence + bitrate ranges*                   |
| **Public path (4 mirrors)** | `vasic-digital/helix-codec` on GitHub + GitLab + GitFlic + GitVerse                                                    |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | None at runtime (codec selection is pure config + algorithm); cgo bindings to libavutil for parameter validation only |
| **Licence (S01 §4.8)**      | **Apache-2.0** (codec-related patentable algorithms)                                                                    |
| **Container CI lane (S02 §3)** | `codec-ladder-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                          |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/06_4k_120hz_hdr_dolby_vision/scenarios/03_codec_ladder_at_4k120.scenario.yaml`                          |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-codec` is the **codec selection + GOP cadence + bitrate ladder configuration** primitive. Origin: [C26 §6](../../05_Video_Audio/01_Codec_Selection.md). The submodule is **declarative** — it does not encode bytes itself (that's `helix-encoder`'s job). It exposes the canonical codec configurations: H.264, HEVC, AV1, JPEG-XS, plus the experimental PyroWave path; the GOP cadence per profile (60-frame, 120-frame, ultra-low-latency 1-frame); the bitrate ranges per resolution / framerate combination per the C26 §6 ABR ladder.

The submodule was introduced because earlier projects re-implemented codec config inline with subtle drift (different default GOP, different bitrate floors). R-04 mandates one canonical landing.

`★ Apache-2.0 rationale.` Per [S01 §4.8.2](../01_Submodule_Catalog.md#482-why-apache-20-specifically-for-those-four), codec-related code may practice patentable algorithms (H.264 / HEVC patent pools, AV1 AOMedia patent commitments). Apache-2.0's §3 patent grant gives downstream consumers explicit permission to incorporate the submodule into their own product without an additional licence negotiation.

---

## 2. Public API Surface

### 2.1 The `Codec` enum and `Profile` struct

```go
package codec

type Codec int
const (
    CodecH264 Codec = iota
    CodecHEVC
    CodecAV1
    CodecJPEGXS
    CodecPyroWave  // experimental
)

type Profile struct {
    Codec       Codec
    Resolution  Resolution    // 1080p, 1440p, 4K, 8K
    Framerate   int           // 60, 90, 120, 144, 240
    GOPSize     int           // frames per GOP; 1 for ultra-low-latency
    BitrateBps  int           // target bitrate
    HDRMode     HDRMode       // SDR, HDR10, HDR10+, DolbyVision
    EntropyMode EntropyMode   // CABAC for H.264; CABAC variants for HEVC
}
```

### 2.2 The canonical ladder

```go
package codec

// CanonicalLadder returns the C26 §6-mandated 8-tier ladder per
// resolution. ABR descents step down through these.
func CanonicalLadder(resolution Resolution) []Profile

// PickProfile selects a profile from the ladder given current
// network conditions (RTT, observed bandwidth, loss rate).
func PickProfile(ladder []Profile, conditions NetworkConditions) Profile
```

### 2.3 The `Config` builder

```go
package codec

// Config builds a codec configuration that helix-encoder consumes.
type Config struct {
    Profile        Profile
    KeyframeInterval int
    BFrames        int
    LookaheadFrames int
    RateControl    RateControlMode  // CBR, VBR, CRF, CQP
    AdaptiveQuant  AdaptiveQuantMode
}

func DefaultConfig(profile Profile) *Config
func (c *Config) Validate() error  // calls libavutil for sanity
```

### 2.4 GOP cadence helpers

```go
package codec

// UltraLowLatencyGOP returns a GOP=1 (intra-only) configuration
// suitable for the lowest-latency mode. Trade-off: bitrate roughly
// 4× a typical 60-frame GOP.
func UltraLowLatencyGOP(profile Profile) *Config

// LowLatencyGOP returns the canonical GOP=60 config with no B-frames.
func LowLatencyGOP(profile Profile) *Config

// HighEfficiencyGOP returns a GOP=120 config with B-frames; lower
// bitrate but higher first-frame latency.
func HighEfficiencyGOP(profile Profile) *Config
```

### 2.5 Capability detection

```go
package codec

func DetectCapabilities() (Capabilities, error)
type Capabilities struct {
    H264  bool
    HEVC  bool
    AV1   bool
    JPEGXS bool
    PyroWaveSupported bool
    HDR10Support bool
    HDR10PlusSupport bool
    DolbyVisionLowLatency bool
}
```

### 2.6 Configuration options

```go
package codec

type Option func(*config)

func WithMaxBitrate(bps int) Option
func WithMinBitrate(bps int) Option
func WithRateControl(mode RateControlMode) Option
```

### 2.7 Statistics

```go
package codec

type ConfigStats struct {
    ProfilesPicked   uint64
    LadderTransitions uint64    // up + down rung changes
    RateControlMode  RateControlMode
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`ffprobe -show_streams` for capability detection during init).

### 3.2 External (Go)

- libavutil cgo bindings (validation only; no encoding): pinned to FFmpeg ≥ 7.0 headers.

### 3.3 External (system)

None at runtime; the submodule is config-only.

---

## 4. Container Build (S02 §3 lane: `codec-ladder-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Profile validation; CanonicalLadder shape correctness; PickProfile decision matrix.

### 5.2 Integration
Real libavutil validation against Profile parameters.

### 5.3 E2E
helix-codec → helix-encoder configuration round-trip.

### 5.4 Security
govulncheck + Snyk + Trivy.

### 5.5 Benchmarking
PickProfile p999 ≤ 1 µs; Validate p999 ≤ 100 µs.

### 5.6 Chaos
Inject malformed NetworkConditions; verify PickProfile returns a safe default.

### 5.7 Stress
24-hour 1 K profile-switches/s; verify no leak in the ladder transitions counter.

### 5.8 Smoke
30-second post-deploy: build a canonical profile + validate.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`06_4k_120hz_hdr_dolby_vision/03_codec_ladder_at_4k120.scenario` — 4K120 HDR session; verify ladder selection stays at the 4K120 rung under steady-state and descends correctly under network impairment.

---

## 6. Challenges Entry-Point (S03 §4 row #18)

**Topology:** `06_4k_120hz_hdr_dolby_vision`. **Scenario:** `03_codec_ladder_at_4k120.scenario.yaml`. **Why this scenario.** The 4K120 HDR profile is the most demanding ladder rung; the scenario verifies the codec config stays valid + descend / ascend transitions are bit-exact. **Baseline:** ladder transitions match recorded sequence; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-codec` imports `helix-r18-safeexec` for boundary subprocess invocations (ffprobe diagnostic). Hot path is pure-Go config + cgo libavutil validation.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the C26 §6 API freeze + ladder shape stability across release cycles. Apache-2.0 SPDX header on every source file is enforced by the spdx-check lint per S01 §4.8.6.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                        | Purpose                                                                |
|------------------------------------|---------------|-------------------------------------|------------------------------------------------------------------------|
| `HELIX_CODEC_DEFAULT`              | `auto`        | `h264`/`hevc`/`av1`/`jpegxs`/`auto` | Default codec; auto picks best supported.                              |
| `HELIX_CODEC_GOP_DEFAULT`          | `low-latency` | `ultra-low`/`low-latency`/`high-eff`| Default GOP cadence.                                                    |
| `HELIX_CODEC_RATE_CONTROL`         | `CBR`         | `CBR`/`VBR`/`CRF`/`CQP`             | Default rate control mode.                                              |
| `HELIX_CODEC_LADDER_VERSION`       | `c26-v1`      | string                              | Pin to a specific ladder version for reproducibility.                  |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `PickProfile()`                       | 200 ns   | 500 ns  | 1 µs    | O(8) ladder traversal.                                            |
| `Validate()`                          | 30 µs    | 70 µs   | 100 µs  | libavutil cgo round-trip.                                        |
| `CanonicalLadder()`                   | 100 ns   | 300 ns  | 500 ns  | Static lookup.                                                   |
| Ladder transition memory              | 64 B     | 64 B    | 64 B    | Constant; no allocation.                                          |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `codec: ErrUnsupportedCodec`                  | Hardware lacks the codec                                       | Fall back via `HELIX_CODEC_DEFAULT=auto`.                                  |
| `codec: ErrInvalidProfile`                    | Bitrate / framerate / resolution combination out of range     | Use CanonicalLadder to constrain choices.                                  |
| `codec: ErrLibavutilMismatch`                 | FFmpeg headers differ from runtime libavutil                  | Rebuild against current FFmpeg version; pin FFmpeg version in container. |
| `codec: ErrHDRNotSupported`                   | Codec does not support requested HDR mode                     | Pick a different codec (HEVC + AV1 support HDR10+; AV1 supports DV).      |

### 9.4 Migration from inline codec configs

A consumer migrating from inline encoder parameters:

1. Replace `Encoder.SetBitrate(...)` etc. with `cfg := codec.DefaultConfig(profile); encoder.Configure(cfg)`.
2. Replace ABR-descent logic with `codec.PickProfile(ladder, conditions)`.
3. Add `cfg.Validate()` at startup; surface errors clearly.
4. Add OTLP span around PickProfile; metric for `LadderTransitions`.

The migration is documented in `docs/migration-from-inline-codec.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_codec_profiles_picked_total`             | counter    | PickProfile invocations, labelled `codec`, `resolution`, `rung`.             |
| `helix_codec_ladder_transitions_total`          | counter    | Up + down transitions, labelled `direction`, `from_rung`, `to_rung`.        |
| `helix_codec_validate_failures_total`           | counter    | Validate() failures, labelled `error_kind`.                                  |
| `helix_codec_capability_detected`               | gauge      | Per-codec capability (1 if available, 0 if not).                            |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C26 §6](../../05_Video_Audio/01_Codec_Selection.md) — origin                              | Origin chapter; full Codec / Profile / Config API.                                   |
| [C27 §6](../../05_Video_Audio/02_Hardware_Encoders.md) — Encoder                          | helix-encoder consumes Profile + Config.                                            |
| [C32 §6](../../05_Video_Audio/07_HDR_and_Color.md) — HDR                                  | HDR pipeline reads HDRMode from Profile.                                            |
| [C33 §6](../../05_Video_Audio/08_ABR_FEC_Congestion.md) — ABR                             | ABR uses CanonicalLadder + PickProfile for rung descent.                            |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-codec-A          | PyroWave production-readiness — graduate from experimental in v1.0.0 or wait?                                 | C26 §6 next revision                                |
| OQ-codec-B          | JPEG-XS — niche broadcast codec; should it remain in the ladder?                                              | C26 §6 next revision                                |
| OQ-codec-C          | AV1 patent-pool monitoring — ongoing AOMedia patent-commitments tracking?                                      | `08_Operations/02_Quality_Gates_SonarQube_Snyk.md` |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/01_Codec_Selection.md`](../../05_Video_Audio/01_Codec_Selection.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §4.8 | 1,218 | 2026-04-30 | catalog row #18, Apache-2.0 rationale  |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Profile + Config + ladder).                                       |
| Integration    | libavutil validation round-trip.                                                              |
| E2E            | helix-codec → helix-encoder configuration round-trip.                                         |
| Security       | govulncheck + Snyk + Trivy.                                                                   |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Malformed NetworkConditions; safe default returned.                                           |
| Stress         | 24-hour 1 K profile-switches/s; zero leak.                                                    |
| Smoke          | 30-second canonical profile + validate.                                                       |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `06_4k_120hz_hdr_dolby_vision/03_codec_ladder_at_4k120` baseline-parity.                      |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-codec.md` — 2026-04-30.
