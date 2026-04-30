# `helix-vqa` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-vqa`                                                                                                            |
| **Origin chapter:section**  | [C35 §6](../../05_Video_Audio/10_Measurement_and_QA.md) — *VMAF + LDAT + change-point measurement harness*            |
| **Public path (4 mirrors)** | `vasic-digital/helix-vqa` on GitHub + GitLab + GitFlic + GitVerse                                                      |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-bench`                                                                                |
| **External Go deps**        | libvmaf cgo, ViSQOL cgo, scikit-learn-go (change-point), `golang.org/x/sys/unix`                                        |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `vqa-vmaf-ldat-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                          |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/06_4k_120hz_hdr_dolby_vision/scenarios/07_vmaf_ldat_against_baseline.scenario.yaml`                     |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2** (depends on helix-bench)                                                                                     |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-vqa` is the **video / audio quality measurement harness**. Origin: [C35 §6](../../05_Video_Audio/10_Measurement_and_QA.md). It runs VMAF (Netflix Video Multi-Method Assessment Fusion) for video quality, ViSQOL + spectrogram MSE for audio quality, LDAT-style hardware-anchored latency probes, and SciPy-equivalent change-point detection (Mann-Whitney U + Kolmogorov-Smirnov per [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md) + [C35 §6](../../05_Video_Audio/10_Measurement_and_QA.md)).

The submodule was introduced because video QA logic is the natural complement to `helix-bench`'s latency QA — both are measurement instruments, both feed `vasic-digital/Challenges` baselines + change-point detection per [S03 §6.3](../03_Challenges_Submodule.md#63-change-point-detection).

---

## 2. Public API Surface

### 2.1 The `VMAF` runner

```go
package vqa

// VMAF computes the per-frame VMAF score between a reference and
// a distorted video stream.
type VMAF struct { /* ... */ }

func NewVMAF(modelPath string, opts ...Option) (*VMAF, error)
func (v *VMAF) Score(referenceFrames, distortedFrames []Frame) ([]float64, error)
func (v *VMAF) ScoreStreaming(reference, distorted FrameSource) (<-chan float64, error)
```

### 2.2 The `ViSQOL` runner (audio)

```go
package vqa

// ViSQOL computes audio MOS-equivalent score.
type ViSQOL struct { /* ... */ }
func NewViSQOL(opts ...Option) (*ViSQOL, error)
func (v *ViSQOL) Score(reference, distorted []float32, sampleRate int) (float64, error)
```

### 2.3 The `LDAT` latency probe

```go
package vqa

// LDAT (Latency Display Analysis Tool) measures end-to-end latency
// using a hardware-anchored visual probe (HelixPlay-specific Spec).
type LDAT struct { /* ... */ }
func NewLDAT(probeColor color.RGBA) (*LDAT, error)
func (l *LDAT) Measure(captureDevice string, durationS int) (LDATResult, error)
type LDATResult struct {
    SamplesCollected int
    P50Ms, P99Ms, P999Ms float64
}
```

### 2.4 Change-point detection

```go
package vqa

// DetectChangePoints runs Mann-Whitney U + Kolmogorov-Smirnov on
// a time-series of measurements (per C35 §6).
func DetectChangePoints(series []float64, sensitivityP float64) ([]int, error)
```

### 2.5 Configuration options

```go
package vqa

type Option func(*config)

func WithVMAFModel(path string) Option        // NEG vs default
func WithViSQOLMode(mode string) Option       // audio vs speech
func WithChangePointThreshold(p float64) Option
```

### 2.6 Statistics

```go
package vqa

type VQAStats struct {
    VMAFRunsTotal       uint64
    ViSQOLRunsTotal     uint64
    LDATSamplesCollected uint64
    ChangePointsDetected uint64
    AverageVMAF          float64
    AverageViSQOL        float64
}
```

### 2.7 Frame + FrameSource types

```go
package vqa

type Frame struct {
    YUV    [][]byte
    Width  int
    Height int
}

type FrameSource interface {
    Next() (*Frame, error)
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (LDAT capture-device probing).
- `helix-bench` — VQA results piped to bench's HDR histogram for trend analysis.

### 3.2 External (Go)

- libvmaf cgo (Netflix's VMAF; pinned to ≥ v3.0.0).
- ViSQOL cgo (Google's ViSQOL).
- scikit-learn-go (for Mann-Whitney U + KS).

### 3.3 External (system)

- LDAT-compatible capture device (NVIDIA Reflex Latency-and-Display Analysis Tool spec).

---

## 4. Container Build (S02 §3 lane: `vqa-vmaf-ldat-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
VMAF score computation against synthetic frames; ChangePoint detection on known-shifted series.

### 5.2 Integration
Real reference + distorted clips; verify VMAF ≥ 90 baseline.

### 5.3 E2E
helix-vqa → run-archive → HelixQA change-point report.

### 5.4 Security
govulncheck + Snyk + Trivy.

### 5.5 Benchmarking
VMAF score 4K 60-frame clip p999 ≤ 4 s; ViSQOL 30-second audio p999 ≤ 2 s.

### 5.6 Chaos
Truncated video clip; verify VMAF surfaces error cleanly.

### 5.7 Stress
24-hour bench-coupled run; zero leak; histogram trends stable.

### 5.8 Smoke
30-second post-deploy: run a known-pair VMAF; verify score within ±0.5 of expected.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`06_4k_120hz_hdr_dolby_vision/07_vmaf_ldat_against_baseline.scenario` — full session VMAF + LDAT against archived baseline.

---

## 6. Challenges Entry-Point (S03 §4 row #27)

**Topology:** `06_4k_120hz_hdr_dolby_vision`. **Scenario:** `07_vmaf_ldat_against_baseline.scenario.yaml`. **Why this scenario.** VMAF + LDAT together cover the end-to-end perceptual + latency signal that operators care about. **Baseline:** VMAF ≥ 95 vs reference; LDAT p999 ≤ 8 ms; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-vqa` imports `helix-r18-safeexec` for boundary subprocess invocations (capture-device probing). Hot path is cgo (libvmaf + ViSQOL).

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on libvmaf + ViSQOL ABI stability + LDAT spec finalisation.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default                          | Range / type                | Purpose                                                                |
|------------------------------------|----------------------------------|-----------------------------|------------------------------------------------------------------------|
| `HELIX_VQA_VMAF_MODEL`             | `vmaf_v0.6.1neg.json`            | filesystem path             | VMAF model path.                                                        |
| `HELIX_VQA_VISQOL_MODE`            | `audio`                          | `audio`/`speech`            | ViSQOL mode.                                                            |
| `HELIX_VQA_CHANGEPOINT_P`          | `0.01`                           | float [0.001, 0.1]          | Change-point detection p-value threshold.                              |
| `HELIX_VQA_LDAT_DEVICE`            | (required for LDAT)              | string                      | Capture-device path / identifier.                                      |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| VMAF score 1080p 60-frame             | 1.5 s    | 2.5 s   | 3 s     | Single-thread; libvmaf reference.                                |
| VMAF score 4K 60-frame                | 2.5 s    | 3.5 s   | 4 s     | Single-thread.                                                    |
| ViSQOL 30-second audio                | 1 s      | 1.5 s   | 2 s     | 48 kHz audio scoring.                                            |
| ChangePoint 1000-element series       | 50 ms    | 100 ms  | 200 ms  | Mann-Whitney + KS combined.                                       |
| LDAT measurement (60 s window)        | 60 s     | —       | —       | Wall-clock equals window; samples ≥ 600.                        |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `vqa: ErrModelNotFound`                       | VMAF model file path invalid                                   | Verify `HELIX_VQA_VMAF_MODEL`; ship default in container.                 |
| `vqa: ErrFrameMismatch`                       | Reference and distorted differ in shape                       | Resize / crop pre-processing required at consumer side.                    |
| `vqa: ErrLDATDeviceMissing`                   | Capture device not present                                     | Verify operator-supplied LDAT device; document in runbook.                |
| `vqa: ErrChangePointNoSignal`                 | Series too short for detection                                | Increase sample count; minimum ~50 elements.                               |

### 9.4 Migration from FFmpeg vmaf-runs

A consumer migrating from `ffmpeg -lavfi libvmaf=...`:

1. Replace the FFmpeg subprocess (R-18 forbids long-running ffmpeg) with `vmaf, _ := vqa.NewVMAF(modelPath)`.
2. Stream frames in: `scores, _ := vmaf.Score(refFrames, distFrames)`.
3. Pipe results to `helix-bench` for histogram + change-point analysis.

The migration is documented in `docs/migration-from-ffmpeg-vmaf.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_vqa_vmaf_runs_total`                     | counter    | VMAF score runs.                                                             |
| `helix_vqa_visqol_runs_total`                   | counter    | ViSQOL score runs.                                                           |
| `helix_vqa_ldat_samples_total`                  | counter    | LDAT samples collected.                                                      |
| `helix_vqa_changepoints_detected_total`         | counter    | Change-points detected.                                                      |
| `helix_vqa_average_vmaf`                        | gauge      | Latest average VMAF.                                                         |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C35 §6](../../05_Video_Audio/10_Measurement_and_QA.md) — origin                           | Origin chapter; full VMAF + ViSQOL + LDAT + ChangePoint API.                         |
| [`helix-bench.md`](helix-bench.md)                                                        | helix-bench → helix-vqa run-archive integration.                                     |
| HelixQA (S04)                                                                              | Reads VQA results from run-archive for change-point reports.                        |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-vqa-A            | VMAF NEG vs default model — operator preference?                                                              | C35 §6 next revision                                |
| OQ-vqa-B            | LDAT device sourcing — operator-supplied or HelixPlay-bundled hardware?                                       | `08_Operations/01_Container_CI_CD.md`              |
| OQ-vqa-C            | Real-time VMAF (per-frame streaming) — feasibility on amd64 vs arm64?                                          | C35 §6 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/10_Measurement_and_QA.md`](../../05_Video_Audio/10_Measurement_and_QA.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-bench.md`](helix-bench.md) | (this batch) | 2026-04-30 | direct dependency                              |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (VMAF + ViSQOL + ChangePoint).                                     |
| Integration    | Reference + distorted clip VMAF baseline.                                                    |
| E2E            | helix-vqa → run-archive → HelixQA change-point report.                                       |
| Security       | govulncheck + Snyk + Trivy.                                                                   |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Truncated clip; clean error surfacing.                                                        |
| Stress         | 24-hour bench-coupled; histograms stable.                                                     |
| Smoke          | 30-second known-pair VMAF score.                                                              |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `06_4k_120hz_hdr_dolby_vision/07_vmaf_ldat_against_baseline` baseline-parity.                 |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-vqa.md` — 2026-04-30.
