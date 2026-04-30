# `helix-pipeline` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-pipeline`                                                                                                       |
| **Origin chapter:section**  | [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — *Goroutine topology + cgo coordination*             |
| **Public path (4 mirrors)** | `vasic-digital/helix-pipeline` on GitHub + GitLab + GitFlic + GitVerse                                                 |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-shm`, `helix-lockfree`, `helix-mempool`, `helix-allocator`, `helix-bench`, `helix-encoder`, `helix-capture`, `helix-dualpath` |
| **External Go deps**        | `golang.org/x/sync/errgroup`, `runtime/trace`                                                                            |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `pipeline-cgo-1.x` — builder `golang-builder-gpu`, runtime `distroless-cuda`                                         |
| **Test matrix (S01 §5)**    | Ten / inline (also owns helix-shm's delegated Challenges row per S01 §5.2)                                             |
| **Challenges entry (S03 §4)** | `topologies/01_minimum_viable_session/scenarios/06_full_pipeline_end_to_end.scenario.yaml`                          |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **4** (deepest in the catalog; imports 9 sibling submodules)                                                       |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-pipeline` is the **encode-side pipeline orchestrator** that ties together capture, encoder, dual-path split, recording, audio, HDR, ABR, thermal, and bench instrumentation into a single coordinated goroutine topology. Origin: [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md). It is the **deepest** submodule in the catalog (depth 4) — every Latency-family + Video/Audio-family submodule it composes hangs off `helix-r18-safeexec` at depth 0.

The submodule was introduced because composing 9 sibling submodules requires careful goroutine lifecycle management — context cancellation, error propagation, backpressure handoff between SPSC rings, allocation-free hot paths via `helix-allocator`. R-04 mandates one canonical landing.

`★ Per S01 §5.2`, `helix-pipeline` additionally owns `helix-shm`'s delegated Challenges row — its `06_full_pipeline_end_to_end.scenario.yaml` exercises every public `helix-shm` API.

---

## 2. Public API Surface

### 2.1 The `Pipeline` type

```go
package pipeline

// Pipeline is the encode-side orchestration unit. It owns capture +
// encoder + dual-path + recording + audio + thermal + abr + hdr +
// bench goroutines and coordinates them via context cancellation.
type Pipeline struct { /* ... */ }

func New(cfg *Config) (*Pipeline, error)
func (p *Pipeline) Start(ctx context.Context) error
func (p *Pipeline) Stop() error
func (p *Pipeline) Stats() PipelineStats
```

### 2.2 The `Config` type

```go
package pipeline

type Config struct {
    Capture   *capture.Config
    Encoder   *encoder.Config
    DualPath  *dualpath.Config
    Record    *record.Config
    Audio     *audio.Config
    HDR       *hdr.Config
    ABR       *abr.Config
    Thermal   *thermal.Config
    Bench     *bench.Config
    
    StreamSink Sink   // helix-transport adapter
}
```

### 2.3 The `Sink` interface

```go
package pipeline

// Sink receives encoded NAL units + audio packets ready for transport.
type Sink interface {
    SendVideo(rung dualpath.Rung, nalUnits [][]byte, meta FrameMeta) error
    SendAudio(packet []byte, meta AudioMeta) error
}
```

### 2.4 Pipeline events (observability)

```go
package pipeline

type Event struct {
    Type       EventType
    Timestamp  time.Time
    Stage      Stage     // Capture, Encode, DualPath, Record, Audio, Transport
    Detail     string
}

type EventType int
const (
    EventStarted EventType = iota
    EventStopped
    EventFrameDropped
    EventEncodeError
    EventBackpressure
    EventThermalScale
)

func (p *Pipeline) Events() <-chan Event
```

### 2.5 Pipeline statistics

```go
package pipeline

type PipelineStats struct {
    StartTime     time.Time
    FramesProcessed uint64
    AudioPacketsProcessed uint64
    DropsCapture  uint64
    DropsEncode   uint64
    DropsTransport uint64
    AverageEndToEndLatencyMS float64
    PerStageLatencyHistogram map[Stage]*hdrhistogram.Histogram
}
```

### 2.6 Configuration options

```go
package pipeline

type Option func(*config)

func WithRTSScheduling() Option              // PromoteCurrent on hot-path goroutines
func WithBudget(maxLatencyMs int) Option    // hard end-to-end latency cap
func WithEventChannelSize(n int) Option
func WithBenchEvery(samples int) Option
```

### 2.7 The pipeline graph

```go
package pipeline

// PipelineGraph returns a DOT representation of the goroutine
// topology for runtime visualisation.
func (p *Pipeline) PipelineGraph() string
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

Per S01 §3.1's row #28, the deepest dependency list in the catalog: **9 sibling submodules**:

- `helix-r18-safeexec`
- `helix-shm`
- `helix-lockfree`
- `helix-mempool`
- `helix-allocator`
- `helix-bench`
- `helix-encoder`
- `helix-capture`
- `helix-dualpath`

Plus implicit per-stage references through capture/encoder/dualpath into helix-record, helix-audio, helix-hdr, helix-abr, helix-thermal.

### 3.2 External (Go)

- `golang.org/x/sync/errgroup` — goroutine error propagation.
- `runtime/trace` — for the allocator + change-point harness instrumentation.

### 3.3 External (system)

GPU + capture device per the per-stage dependencies.

---

## 4. Container Build (S02 §3 lane: `pipeline-cgo-1.x`)

**Builder:** `golang-builder-gpu`. **Runtime:** `distroless-cuda`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + GPU device-passthrough + (Linux) DRM device for capture.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Pipeline lifecycle + event-channel correctness.

### 5.2 Integration
Real per-stage submodule integration; verify Pipeline.Start brings every stage up.

### 5.3 E2E
Full pipeline: capture → encode → dual-path → record + transport → client decode.

### 5.4 Security
govulncheck + Snyk + Trivy.

### 5.5 Benchmarking
End-to-end latency p999 ≤ 8 ms (per Insight #2 + C13 §6 budget).

### 5.6 Chaos
Inject per-stage failure (capture device disconnect; encoder OOM); verify Pipeline.Stop terminates cleanly.

### 5.7 Stress
24-hour pipeline run; zero leak; PerStageLatencyHistogram stable.

### 5.8 Smoke
30-second post-deploy: Pipeline.Start + Stop; verify event log shows clean lifecycle.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`01_minimum_viable_session/06_full_pipeline_end_to_end.scenario` — full reference user journey; verify VMAF ≥ 95 + p999 latency ≤ 8 ms + zero leak. **Also serves as the helix-shm delegated Challenges row** (S01 §5.2 / S03 §4.2).

---

## 6. Challenges Entry-Point (S03 §4 row #28 + helix-shm delegation)

**Topology:** `01_minimum_viable_session`. **Scenario:** `06_full_pipeline_end_to_end.scenario.yaml`. **Why this scenario.** This is the canonical pipeline reference: every public API of every depth-1+ submodule is exercised. **Baseline:** VMAF ≥ 95; ViSQOL ≥ 4.5; p999 latency ≤ 8 ms; zero hot-path allocations (verified via runtime/trace). **Per-submodule slices** of the baseline (e.g. `helix-shm` zero-copy throughput) are recorded at `vasic-digital/Challenges/baselines/01_minimum_viable_session/06_full_pipeline_end_to_end/per-submodule/<name>.metrics.json`.

---

## 7. R-18 Inheritance

`helix-pipeline` imports `helix-r18-safeexec` for boundary subprocess invocations + transitively through every imported sibling. Hot path is pure-Go orchestration + cgo through encoder/capture/HDR.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on **every** sibling reaching `v1.0.0` first; `helix-pipeline` is the last Video/Audio-family submodule to graduate.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_PIPELINE_BUDGET_MS`         | `8`           | int [1, 100]            | Hard end-to-end latency cap; budget exceeded triggers EventBackpressure. |
| `HELIX_PIPELINE_EVENT_BUFFER`      | `1024`        | int [16, 16384]         | Event-channel buffer size.                                              |
| `HELIX_PIPELINE_BENCH_EVERY`       | `120`         | int [1, 1000000]        | Bench sampling cadence; 120 = once per second @ 120 fps.              |
| `HELIX_PIPELINE_RTS_PROMOTION`     | `true`        | bool                    | Promote hot-path goroutines via helix-rtos.                            |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| End-to-end latency (capture→sink)     | 5 ms     | 7 ms    | 8 ms    | Per Insight #2; zero allocations on hot path.                    |
| Pipeline.Start to first frame         | 200 ms   | 300 ms  | 500 ms  | Capture init + encoder cold start dominate.                      |
| Pipeline.Stop                         | 50 ms    | 100 ms  | 200 ms  | All-stage drain.                                                  |
| Memory steady state (4K120, single session) | 200 MiB | — | — | Mostly helix-shm pools + encoder state.                         |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `pipeline: ErrStageStartFailed`               | One sibling submodule failed Start                            | Inspect Event log; per-stage error handlers surface root cause.           |
| `pipeline: ErrBudgetExceeded`                 | End-to-end latency > budget                                   | Inspect PerStageLatencyHistogram; helix-thermal may have throttled.       |
| `pipeline: ErrTransportSinkRejected`          | helix-transport applied backpressure                          | helix-abr should descend; investigate if backpressure persists.           |
| `pipeline: ErrHotPathAllocation`              | helix-allocator detected an allocation                        | Audit; offending function identified by stack trace.                      |

### 9.4 Migration from monolithic encode-and-send

A consumer migrating from inline goroutine spaghetti:

1. Identify the existing pipeline's stages + their backpressure points.
2. Construct `pipeline.New(&pipeline.Config{...})`.
3. Replace existing `go captureLoop()` / `go encodeLoop()` etc. with `pipe.Start(ctx)`.
4. Subscribe to `pipe.Events()` for observability.
5. Remove the bespoke pipeline; the submodule owns the topology.

The migration is documented in `docs/migration-from-spaghetti.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_pipeline_frames_processed_total`         | counter    | Frames through the pipeline.                                                 |
| `helix_pipeline_drops_total`                    | counter    | Drops, labelled `stage`.                                                      |
| `helix_pipeline_end_to_end_latency_seconds`     | histogram  | End-to-end latency.                                                          |
| `helix_pipeline_per_stage_latency_seconds`      | histogram  | Per-stage latency, labelled `stage`.                                          |
| `helix_pipeline_active_sessions`                | gauge      | Concurrent active pipelines.                                                |
| `helix_pipeline_budget_exceeded_total`          | counter    | EventBudgetExceeded fires.                                                   |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — origin                   | Origin chapter; full Pipeline + Config + Events API.                                 |
| HelixPlay host agent (`HelixDevelopment/HelixAgent`)                                      | Production deployment; the host agent's session lifecycle owns Pipeline.Start/Stop. |
| HelixQA (S04)                                                                              | Reads Pipeline.Stats() for Challenges-row baselines.                                |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-pipeline-A       | Multi-session pipeline pool — should the submodule own pool semantics or delegate to host agent?              | C36 §8 next revision                                |
| OQ-pipeline-B       | Goroutine count cap — operator-tunable per session?                                                            | `08_Operations/01_Container_CI_CD.md`              |
| OQ-pipeline-C       | Per-stage circuit-breaker — should the submodule expose a CircuitBreaker hook?                                | C36 §8 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/11_Go_Pipeline_Implementation.md`](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) §8 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §5.2 §6 §7 | 1,218 | 2026-04-30 | catalog row #28; depth-4; helix-shm delegation |
| All 9 dependency descriptors in this batch                                                                                                              | (this batch) | 2026-04-30 | direct deps                                    |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Pipeline lifecycle + Event channel).                              |
| Integration    | All 9 sibling submodules integrated; Pipeline.Start brings each up.                          |
| E2E            | Full reference user journey: capture→encode→sink.                                             |
| Security       | govulncheck + Snyk + Trivy.                                                                   |
| Benchmarking   | End-to-end p999 ≤ 8 ms; zero hot-path allocations via helix-allocator.                       |
| Chaos          | Per-stage failure injection; clean Pipeline.Stop.                                             |
| Stress         | 24-hour pipeline; histograms stable.                                                          |
| Smoke          | 30-second Start + Stop; clean event log.                                                      |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `01_minimum_viable_session/06_full_pipeline_end_to_end` baseline-parity (also covers helix-shm delegation). |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-pipeline.md` — 2026-04-30.
