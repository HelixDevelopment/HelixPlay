# `helix-hdr` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-hdr`                                                                                                            |
| **Origin chapter:section**  | [C32 §6](../../05_Video_Audio/07_HDR_and_Color.md) — *PQ/HLG transfer + HDR10/10+/Dolby Vision + tone-mapping*        |
| **Public path (4 mirrors)** | `vasic-digital/helix-hdr` on GitHub + GitLab + GitFlic + GitVerse                                                      |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-codec`                                                                                |
| **External Go deps**        | libavutil cgo (colour-space conversion), Vulkan compute cgo (tone-mapping)                                              |
| **Licence (S01 §4.8)**      | **Apache-2.0** (HDR10+/Dolby Vision tone-map patents)                                                                   |
| **Container CI lane (S02 §3)** | `hdr-tone-map-1.x` — builder `golang-builder-gpu`, runtime `distroless-cuda`                                        |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/06_4k_120hz_hdr_dolby_vision/scenarios/06_dolby_vision_tone_map_under_session.scenario.yaml`            |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2** (depends on helix-codec)                                                                                     |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-hdr` is the **HDR transfer-function and tone-mapping** primitive. Origin: [C32 §6](../../05_Video_Audio/07_HDR_and_Color.md). It tags encoded streams with the correct PQ (SMPTE ST 2084) or HLG (Rec. 2100) transfer-function metadata, packages HDR10 / HDR10+ dynamic metadata SEI / Dolby Vision RPU into the NAL emission, and provides GPU-accelerated tone-mapping for clients that lack native HDR display.

The submodule was introduced because HDR metadata handling drifts across H.264 / HEVC / AV1 codecs and across NVIDIA / AMD / Intel hardware decoders. R-04 mandates one canonical landing.

`★ Apache-2.0 rationale.` HDR10+ and Dolby Vision involve patentable tone-mapping algorithms (Dolby holds the DV patent suite; HDR10+ is co-managed by Samsung + 20th Century + others). Apache-2.0 §3 patent grant is the appropriate licence per S01 §4.8.2.

---

## 2. Public API Surface

### 2.1 The `Tagger` type

```go
package hdr

// Tagger annotates encoded NAL units with HDR transfer-function
// + colour-primaries metadata SEIs.
type Tagger struct { /* ... */ }

func NewTagger(transfer TransferFunction, primaries Primaries, opts ...Option) *Tagger
func (t *Tagger) Tag(nalUnits [][]byte, dynamicMeta DynamicMeta) ([][]byte, error)
```

### 2.2 Transfer-function + primaries

```go
package hdr

type TransferFunction int
const (
    TransferSDR TransferFunction = iota
    TransferPQ                    // SMPTE ST 2084
    TransferHLG                   // ARIB STD-B67 / Rec. 2100
)

type Primaries int
const (
    PrimariesBT709  Primaries = iota   // SDR
    PrimariesBT2020                     // HDR
    PrimariesP3                         // DCI-P3
)
```

### 2.3 HDR mode container

```go
package hdr

type Mode int
const (
    ModeSDR Mode = iota
    ModeHDR10
    ModeHDR10Plus
    ModeDolbyVision
)

type DynamicMeta struct {
    Mode           Mode
    HDR10MaxCLL    int     // Maximum Content Light Level (cd/m²)
    HDR10MaxFALL   int     // Maximum Frame-Average Light Level
    HDR10PlusST2094_40 []byte  // SMPTE ST 2094-40 dynamic-metadata SEI
    DolbyVisionRPU []byte  // Dolby Vision Reference Processing Unit metadata
}
```

### 2.4 Tone-mapping (client-side)

```go
package hdr

// ToneMapper performs PQ→SDR or HDR→sRGB tone-mapping using GPU
// compute shaders. Suitable for clients on SDR displays.
type ToneMapper struct { /* ... */ }

func NewToneMapper(input, output ColorSpace, opts ...Option) (*ToneMapper, error)
func (tm *ToneMapper) Process(input *Frame) (*Frame, error)
func (tm *ToneMapper) Close() error
```

### 2.5 Configuration options

```go
package hdr

type Option func(*config)

func WithDynamicMetadata(mode Mode) Option
func WithToneMapper(strategy ToneMapStrategy) Option

type ToneMapStrategy int
const (
    ToneMapHable ToneMapStrategy = iota  // Hable / Uncharted-2 curve
    ToneMapReinhard
    ToneMapBT2390                         // ITU-R BT.2390 reference
)
```

### 2.6 Capability detection

```go
package hdr

func DetectCapabilities() (Capabilities, error)
type Capabilities struct {
    HDR10Encoder         bool
    HDR10PlusEncoder     bool
    DolbyVisionEncoder   bool
    PQTransferSupported  bool
    HLGTransferSupported bool
    BT2020WideGamut      bool
    GPUToneMappingAvailable bool
}
```

### 2.7 Statistics

```go
package hdr

type HDRStats struct {
    FramesTagged      uint64
    DynamicMetaApplied uint64
    ToneMapsPerformed uint64
    AverageMaxCLL     int
    AverageMaxFALL    int
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations.
- `helix-codec` — reads HDRMode from Profile.

### 3.2 External (Go)

- libavutil cgo — colour-space matrix coefficients.
- Vulkan compute cgo — GPU tone-mapping shader execution.

### 3.3 External (system)

- GPU with Vulkan compute support (NVIDIA / AMD / Intel) for tone-mapping.

---

## 4. Container Build (S02 §3 lane: `hdr-tone-map-1.x`)

**Builder:** `golang-builder-gpu`. **Runtime:** `distroless-cuda`. **Multi-arch:** `linux/amd64` (Vulkan); arm64 deferred. **Hardening:** standard S02 §8.2 + device-passthrough for GPU per S02 §8.4.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
SEI byte-correctness against reference HDR10 / HDR10+ samples; Dolby Vision RPU pass-through.

### 5.2 Integration
helix-codec → helix-hdr Tagger; verify SEI emerges in NAL output.

### 5.3 E2E
Tagged stream → decoder → display verification (HDR-aware monitor required for true coverage).

### 5.4 Security
govulncheck + Snyk + Trivy; malformed-DV-RPU fuzzer.

### 5.5 Benchmarking
Tagger.Tag p999 ≤ 100 µs; ToneMapper.Process 4K frame p999 ≤ 4 ms (GPU compute).

### 5.6 Chaos
Inject malformed metadata; verify SEI is dropped without breaking the video stream.

### 5.7 Stress
24-hour 4K HDR session; zero leak; HDRStats stable.

### 5.8 Smoke
30-second post-deploy: tag a frame; verify SEI present in output.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`06_4k_120hz_hdr_dolby_vision/06_dolby_vision_tone_map_under_session.scenario` — full DV session with on-the-fly tone-mapping for SDR fallback; verify VMAF ≥ 90 on tone-mapped output.

---

## 6. Challenges Entry-Point (S03 §4 row #24)

**Topology:** `06_4k_120hz_hdr_dolby_vision`. **Scenario:** `06_dolby_vision_tone_map_under_session.scenario.yaml`. **Why this scenario.** Dolby Vision + tone-mapping is the highest-fidelity HDR path; the scenario verifies metadata round-trip + perceptual quality. **Baseline:** SEI byte-exact against baseline; tone-mapped VMAF ≥ 90 vs DV-native reference.

---

## 7. R-18 Inheritance

`helix-hdr` imports `helix-r18-safeexec` for boundary subprocess invocations. Hot path is cgo (libavutil + Vulkan).

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on per-mode Challenges parity (HDR10 + HDR10+ + DV) and Apache-2.0 SPDX header compliance.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                          | Purpose                                                                |
|------------------------------------|---------------|---------------------------------------|------------------------------------------------------------------------|
| `HELIX_HDR_MODE`                   | `auto`        | `sdr`/`hdr10`/`hdr10plus`/`dolby_vision`/`auto` | Force HDR mode or auto-detect.                              |
| `HELIX_HDR_TONE_MAP_STRATEGY`      | `bt2390`      | `hable`/`reinhard`/`bt2390`           | Tone-mapping strategy.                                                  |
| `HELIX_HDR_TARGET_NITS`            | `100`         | int [50, 1000]                        | SDR target peak luminance.                                              |
| `HELIX_HDR_VULKAN_DEVICE`          | `0`           | int                                   | Vulkan device index for tone-mapping.                                  |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Tagger.Tag (HDR10 SEI)                | 30 µs    | 70 µs   | 100 µs  | Single SEI append.                                                |
| Tagger.Tag (HDR10+ ST2094-40)         | 100 µs   | 200 µs  | 400 µs  | Larger metadata block.                                            |
| Tagger.Tag (Dolby Vision RPU)         | 200 µs   | 400 µs  | 600 µs  | RPU is up to 2 KiB per frame.                                    |
| ToneMapper.Process 4K frame           | 2 ms     | 3 ms    | 4 ms    | GPU compute; bandwidth-bound.                                    |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `hdr: ErrModeUnsupported`                     | Encoder + display combination doesn't support requested mode  | Auto-fall-back via `HELIX_HDR_MODE=auto`.                                 |
| `hdr: ErrDVRPUInvalid`                        | Dolby Vision RPU fails parse                                  | Audit input; verify Dolby Vision authoring chain.                         |
| `hdr: ErrToneMapVulkanInit`                   | Vulkan init failed                                             | Verify GPU driver; check `--device=/dev/dri/...` passthrough.             |
| `hdr: ErrSEIDropped`                          | NAL container couldn't fit the SEI                             | Audit codec config; some constrained profiles reject large SEIs.          |

### 9.4 Migration from inline HDR metadata

A consumer migrating from inline x265-side HDR options:

1. Replace `--master-display ...` x265 flags with `hdr.NewTagger(transfer, primaries)`.
2. After encoder produces NAL output, run `tagger.Tag(nals, dynamicMeta)`.
3. Forward tagged NAL to `helix-dualpath`.
4. For SDR-fallback clients, allocate `hdr.NewToneMapper(...)` on the client-side and run before display.
5. Add OTLP span; metric for `HDRStats.DynamicMetaApplied`.

The migration is documented in `docs/migration-from-inline-hdr.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_hdr_frames_tagged_total`                 | counter    | Frames tagged, labelled `mode`.                                              |
| `helix_hdr_dynamic_meta_total`                  | counter    | Dynamic metadata SEIs applied.                                               |
| `helix_hdr_tone_maps_total`                     | counter    | Tone-map operations.                                                         |
| `helix_hdr_max_cll`                             | gauge      | Last observed MaxCLL.                                                        |
| `helix_hdr_max_fall`                            | gauge      | Last observed MaxFALL.                                                       |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C32 §6](../../05_Video_Audio/07_HDR_and_Color.md) — origin                                | Origin chapter; full Tagger + ToneMapper + Mode + DynamicMeta.                       |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates HDR tagging downstream of encode.                              |
| [C22 §6](../../04_Latency/08_Frame_Pacing_and_VRR.md) — Display                            | Display path coordinates DV low-latency negotiation.                                 |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-hdr-A            | Dolby Vision Profile 8.4 (single-layer DV with HDR10 fallback) — production-supported?                         | C32 §6 next revision                                |
| OQ-hdr-B            | HDR10+ ST 2094-40 generation — submodule-side or relies on encoder ASIC?                                       | C32 §6 next revision                                |
| OQ-hdr-C            | Tone-mapping strategy default — operator-tunable per tenant?                                                  | C11 §6 + C32 §6 next revisions                      |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/07_HDR_and_Color.md`](../../05_Video_Audio/07_HDR_and_Color.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-codec.md`](helix-codec.md) | (this batch) | 2026-04-30 | direct dependency                              |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §4.8 | 1,218 | 2026-04-30 | catalog row #24, Apache-2.0 rationale  |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Tagger + ToneMapper + DynamicMeta).                               |
| Integration    | helix-codec → Tagger SEI emission round-trip.                                                 |
| E2E            | Tagged stream → decoder → display.                                                            |
| Security       | govulncheck + Snyk + Trivy + malformed-DV-RPU fuzzer.                                        |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Malformed metadata; SEI dropped cleanly.                                                      |
| Stress         | 24-hour 4K HDR session; zero leak.                                                            |
| Smoke          | 30-second tag + SEI verification.                                                             |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `06_4k_120hz_hdr_dolby_vision/06_dolby_vision_tone_map_under_session` baseline-parity.        |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-hdr.md` — 2026-04-30.
