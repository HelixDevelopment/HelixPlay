# `helix-bench` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-bench`                                                                                                          |
| **Origin chapter:section**  | [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md) — *Benchmark harness with ≥ 10K samples + p999*       |
| **Public path (4 mirrors)** | `vasic-digital/helix-bench` on GitHub + GitLab + GitFlic + GitVerse                                                    |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-shm`, `helix-iouring`                                                                  |
| **External Go deps**        | `golang.org/x/perf/benchstat`, `github.com/HdrHistogram/hdrhistogram-go`, `golang.org/x/sys/unix`                       |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `bench-harness-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                          |
| **Test matrix (S01 §5)**    | Ten / inline + Benchmarking row delegates per-workload to the workload's own submodule                                  |
| **Challenges entry (S03 §4)** | `topologies/11_stress_24h_steady_state/scenarios/01_p999_floor_over_24h.scenario.yaml`                              |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2** (depends on helix-shm + helix-iouring)                                                                       |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-bench` is the **canonical benchmark harness** for the 29-submodule fleet. Origin: [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md). The submodule provides:

- **HDR-histogram-backed sampling** with the ≥ 10 K-sample minimum mandated by Constitution §6 + Insight #2 of the Latency family.
- **p999-aware reporting** (the only steady-state metric per Insight #2 — p99 is too lenient for HelixPlay's tail-sensitive workload).
- **Cross-fleet result comparison** via `benchstat` for run-archive analysis.
- **Workload generators** for the canonical workloads (frame encoding, packet send, controller round-trip).

The submodule was introduced because earlier projects measured latency inconsistently — different bin sizes, different sample counts, different percentile definitions. R-04 mandates one canonical landing for the bench harness.

---

## 2. Public API Surface

### 2.1 The `Harness` type

```go
package bench

// Harness runs a workload, collects samples into an HDR histogram,
// and reports percentiles + benchstat-compatible output.
type Harness struct { /* ... */ }

func NewHarness(name string, opts ...Option) *Harness
func (h *Harness) RunOnce(workload func()) error
func (h *Harness) RunSteadyState(workload func(), samples int) (*Result, error)
func (h *Harness) Report() *Report
```

### 2.2 The `Result` and `Report` types

```go
package bench

type Result struct {
    Samples    int
    Histogram  *hdrhistogram.Histogram
    P50, P99, P999 time.Duration
    Min, Max, Mean time.Duration
    StdDev      time.Duration
}

type Report struct {
    Runs      []Result
    Workload  string
    Timestamp time.Time
}
```

### 2.3 Workload generators

```go
package bench

// FrameEncodeWorkload generates synthetic NV12 frames at the
// requested rate and submits to the supplied encoder function.
func FrameEncodeWorkload(rate int, encoder func([]byte) error) func()

// PacketSendWorkload generates synthetic UDP packets and sends via
// the supplied function (typically helix-transport.Send).
func PacketSendWorkload(rate int, send func([]byte) error) func()

// ControllerRoundTripWorkload generates synthetic input events and
// measures the round-trip echo latency.
func ControllerRoundTripWorkload(rate int, sendRecv func() error) func()

// ZeroCopyThroughputWorkload uses helix-shm pages + helix-iouring
// for measurement without involving the GC.
func ZeroCopyThroughputWorkload(pageSize int, rate int, channel func([]byte) error) func()
```

### 2.4 Statistical helpers

```go
package bench

// CompareRuns uses benchstat to detect significant differences
// between two Reports. Returns a verdict: improvement, regression,
// or no significant change.
func CompareRuns(baseline, current *Report) (Verdict, error)
type Verdict int
const (
    VerdictImprovement Verdict = iota
    VerdictNoSignificantChange
    VerdictRegression
)

// ChangePoint detection on a sequence of Reports (per [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md))
// using Mann-Whitney U + Kolmogorov-Smirnov.
func DetectChangePoint(reports []*Report) ([]int, error)
```

### 2.5 The `RunArchive` writer

```go
package bench

// RunArchive writes a Report to HelixQA's run-archive (S04 §7) so
// trends can be tracked over time.
type RunArchive struct { /* ... */ }

func NewRunArchive(path string, opts ...ArchiveOption) (*RunArchive, error)
func (ra *RunArchive) Append(report *Report) error
func (ra *RunArchive) LoadAll() ([]*Report, error)
```

### 2.6 Configuration options

```go
package bench

type Option func(*config)

func WithMinSamples(n int) Option            // ≥ 10K per Insight #2
func WithMaxDuration(d time.Duration) Option
func WithCPUPin(cpus []int) Option           // delegates to helix-rtos
func WithWarmup(samples int) Option
func WithHDRPercision(precision int) Option  // hdrhistogram precision
```

### 2.7 Harness statistics (operational, not benchmark output)

```go
package bench

type HarnessStats struct {
    SamplesCollected uint64
    WarmupDiscarded  uint64
    GCAllocsObserved uint64    // alert: > 0 indicates harness-side instrumentation noise
    PinnedCPUs       []int
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (querying CPU model, RDTSC capability, perf-event setup).
- `helix-shm` — for the `ZeroCopyThroughputWorkload`'s page allocation.
- `helix-iouring` — for the same workload's batched submit/reap path.

### 3.2 External (Go)

- `golang.org/x/perf/benchstat` — significance-testing.
- `github.com/HdrHistogram/hdrhistogram-go` — high-dynamic-range histograms.
- `golang.org/x/sys/unix` — for RDTSC + perf_event_open.

### 3.3 External (system)

- `/proc/cpuinfo` for CPU model (informational).
- `perf_event_open` syscall (kernel ≥ 4.4) for precise wall-clock measurement.

---

## 4. Container Build (S02 §3 lane: `bench-harness-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2.

---

## 5. Test Matrix (S01 §5: Ten / inline + delegation note)

Per [S01 §5.2](../01_Submodule_Catalog.md#52-the-two-delegation-exceptions), `helix-bench` is the harness, not the workload. Its **own** Benchmarking row tests the harness itself — does it sample at the requested rate, does it compute p99 / p999 correctly, does it survive a 10K-sample run. Workload-specific benchmarks live in the workload submodule's own `tests/benchmarking/` directory.

### 5.1 Unit
HDR histogram math; benchstat invocation; ChangePoint algorithm correctness.

### 5.2 Integration
Real CPU pinning via helix-rtos; verify samples are CPU-correlated.

### 5.3 E2E
helix-bench → helix-shm + helix-iouring zero-copy throughput workload roundtrip.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: malformed-Result-input fuzzer.

### 5.5 Benchmarking (the harness's own — meta)
- Sample collection p999 ≤ 100 ns (measurement overhead).
- 10 K samples in ≤ 1 second wall-clock.
- HDR histogram serialise/deserialise round-trip.

### 5.6 Chaos
Inject GC pause; verify HDR histogram correctly attributes the pause to the right sample window.

### 5.7 Stress
24-hour bench harness running synthetic workload; verify zero leak; HDR histograms stable.

### 5.8 Smoke
30-second post-deploy: run a 10 K-sample workload; report p999.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`11_stress_24h_steady_state/01_p999_floor_over_24h.scenario` — full system steady-state for 24 h; verify p999 floors documented in §9.2 hold across the entire run.

---

## 6. Challenges Entry-Point (S03 §4 row #17)

**Topology:** `11_stress_24h_steady_state`. **Scenario:** `01_p999_floor_over_24h.scenario.yaml`. **Why this scenario.** The 24-hour steady-state run is the canonical pre-release gate (S04 §5.4); helix-bench is the measurement instrument that decides whether the run passes. **Baseline:** every measured submodule's p999 stays within ±5 % of the recorded baseline; HDR histograms reproducible bit-for-bit (modulo random-seeded workload).

---

## 7. R-18 Inheritance

`helix-bench` imports `helix-r18-safeexec` for boundary subprocess invocations (CPU model query, perf_event setup). Hot path is pure Go + cgo via the perf_event syscall.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the C24 §6 API freeze + 2 consecutive green Ten-test cycles, including a meta-bench that the harness's own measurement overhead stays within budget.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_BENCH_MIN_SAMPLES`          | `10000`       | int [1000, 1000000]     | Minimum samples per RunSteadyState.                                     |
| `HELIX_BENCH_MAX_DURATION_S`       | `60`          | int [1, 86400]          | Max wall-clock per RunSteadyState.                                     |
| `HELIX_BENCH_WARMUP_SAMPLES`       | `1000`        | int [0, 100000]         | Discarded warmup samples.                                              |
| `HELIX_BENCH_HDR_PRECISION`        | `3`           | int [1, 5]              | hdrhistogram significant-digits precision.                              |
| `HELIX_BENCH_PIN_CPU`              | `auto`        | csv int / `auto`        | CPU pinning (delegates to helix-rtos).                                  |

### 9.2 Performance budget (the harness itself)

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Sample collection overhead            | 30 ns    | 70 ns   | 100 ns  | Per-sample HDR histogram update.                                 |
| 10 K-sample run wall-clock            | 0.5 s    | 0.9 s   | 1 s     | Excluding workload time.                                         |
| benchstat compare (50 runs)           | 5 ms     | 15 ms   | 30 ms   | Significance test.                                                |
| ChangePoint detection (50 reports)    | 30 ms    | 80 ms   | 150 ms  | Mann-Whitney U + KS over the sequence.                          |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `bench: ErrInsufficientSamples`               | < 10 K samples collected in the configured wall-clock         | Raise `HELIX_BENCH_MAX_DURATION_S`; investigate workload throttling.       |
| `bench: ErrPerfEventOpenDenied`               | kernel.perf_event_paranoid > 1                                 | Operator: `sysctl kernel.perf_event_paranoid=1`; or fall back to monotonic clock. |
| `bench: ErrHDRSerialisationDrift`             | hdrhistogram lib version mismatch between writer and reader   | Pin the dependency; bump lockstep across the fleet.                       |
| `bench: ErrCPUPinFailed`                      | helix-rtos returned ErrCapSysNiceMissing                       | Add CAP_SYS_NICE per helix-rtos §4.                                       |
| `bench: ErrChangePointAmbiguous`              | Detection algorithm returned no clear change-point             | Inspect the report manually; algorithm documented at C24 §6.              |

### 9.4 Migration from `testing.B` benchmarks

A consumer migrating from Go's `testing.B`:

1. Identify benchmarks whose p999 matters (per Insight #2; `testing.B` reports mean which is too lenient).
2. Replace `b.RunParallel(func(pb *testing.PB) { for pb.Next() { workload() } })` with `harness.RunSteadyState(workload, 10000)`.
3. Convert benchmark output to a `Report` and append to the run-archive.
4. Wire the per-PR CI lane to call `helix-bench-cli compare baseline.json current.json` and fail on `VerdictRegression`.

The migration is documented in `docs/migration-from-testing-b.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_bench_samples_collected_total`           | counter    | Samples collected, labelled `workload_name`.                                 |
| `helix_bench_runs_total`                        | counter    | Bench runs, labelled `workload_name`, `verdict`.                             |
| `helix_bench_p999_seconds`                      | gauge      | Latest p999 per workload.                                                    |
| `helix_bench_change_points_total`               | counter    | Change-points detected; labelled `workload_name`.                            |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md) — origin                   | Origin chapter; full Harness + Result + Report API.                                  |
| **Every** submodule's Benchmarking test row                                               | Harness owner; per-workload code lives in the workload submodule.                    |
| HelixQA (S04)                                                                              | Reads helix-bench Reports from run-archive for change-point detection.              |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-bench-A          | RDTSC vs CLOCK_MONOTONIC vs perf_event — which timer is canonical for HelixPlay measurements?                 | C24 §6 next revision                                |
| OQ-bench-B          | Cross-architecture comparability — amd64 vs arm64 baselines need separate ranges or unified scoring?          | C24 §6 next revision                                |
| OQ-bench-C          | benchstat's significance threshold — default 5 % or tunable per workload?                                      | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/10_Latency_Testing_and_Validation.md`](../../04_Latency/10_Latency_Testing_and_Validation.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-shm.md`](helix-shm.md) | (this batch) | 2026-04-30 | direct dependency                              |
| [`helix-iouring.md`](helix-iouring.md) | (this batch) | 2026-04-30 | direct dependency                      |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (HDR + benchstat + ChangePoint paths).                             |
| Integration    | Real CPU pinning + perf_event_open.                                                          |
| E2E            | helix-shm + helix-iouring zero-copy workload measurement.                                     |
| Security       | govulncheck + Snyk + Trivy + Result fuzzer.                                                  |
| Benchmarking   | Harness overhead ≤ 100 ns p999 per sample; 10K samples in ≤ 1 s.                              |
| Chaos          | GC-pause attribution to correct sample window.                                                |
| Stress         | 24-hour synthetic workload; zero leak; histograms stable.                                     |
| Smoke          | 30-second 10 K-sample run + p999 report.                                                      |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `11_stress_24h_steady_state/01_p999_floor_over_24h` baseline-parity.                          |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-bench.md` — 2026-04-30.
