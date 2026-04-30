# T06 — Benchmarking

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.5; [`../06_Submodules/per-submodule/helix-bench.md`](../06_Submodules/per-submodule/helix-bench.md); [`../04_Latency/10_Latency_Testing_and_Validation.md`](../04_Latency/10_Latency_Testing_and_Validation.md) §6 (the source for change-point detection methodology).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row T06.
> **Chapter targets:** R-11 (every documented performance budget measured), R-13 (regressions caught before merge).
> **Cross-links:** [`08_Stress.md`](08_Stress.md), [`11_Challenges.md`](11_Challenges.md), [`../06_Submodules/per-submodule/helix-bench.md`](../06_Submodules/per-submodule/helix-bench.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Benchmarking is **test-type 5 of 10** in the [T01 §2 grid](01_Test_Matrix.md#2-the-29--10--4-cell-grid-the-canonical-lookup). Cadence: nightly + canary + pre-release. Per-PR runs only a **smoke benchmark** (5K samples instead of the full 10K) to keep PR wall-clock manageable; the full benchmark runs nightly.

Per the [helix-bench](../06_Submodules/per-submodule/helix-bench.md) descriptor + S01 §5.2, **`helix-bench`** is the harness; **per-workload** benchmark code lives in the workload submodule's own `tests/benchmarking/`. Benchmarking is the only test type with documented partial delegation per [S01 §5.2](../06_Submodules/01_Submodule_Catalog.md#52-the-two-delegation-exceptions).

The role: **catch performance regressions before they ship**. Insight #2 of the Latency family (codified in [C13 §6](../03_Architecture/12_Latency_Engineering_Overview.md)) elevates **p999** (not p50, not p99) as the only steady-state metric that matters for HelixPlay; T06's role is to enforce the p999 contract on every PR.

---

## 2. The Discipline — What Makes "Green" Green

A green Benchmarking row means:

1. **≥ 10K samples** per benchmark (Insight #2 minimum sample count). Smaller samples are statistical noise — Mann-Whitney U + KS need the population.
2. **p999 within ±5 % of the recorded baseline**. > 5 % regression blocks the merge; > 10 % improvement requires baseline-replacement PR per [S03 §6.1](../06_Submodules/03_Challenges_Submodule.md#61-recording).
3. **Zero hot-path allocations** (per `helix-allocator` enforcement; [S05 helix-allocator](../06_Submodules/per-submodule/helix-allocator.md)).
4. **Workload pinned to the same CPU set** (via `helix-rtos.PromoteCurrent`); migration count = 0 for pinned threads.
5. **Run-archive entry produced** (via `helix-bench.RunArchive` per [S04 §7](../06_Submodules/04_HelixQA_Integration.md#7-run-archive-and-audit-trail-immutable-queryable-signed)).

---

## 3. Tooling

- **`helix-bench`** (origin C24 §6) — the harness. HDR-histogram-backed sampling; benchstat-compatible reports; ChangePoint detection.
- **HDR Histogram** (`github.com/HdrHistogram/hdrhistogram-go`) — high-dynamic-range sampling.
- **benchstat** (`golang.org/x/perf/cmd/benchstat`) — significance testing across baselines.
- **Go's built-in `testing.B`** — entry point: `Benchmark*` functions in `_test.go`.
- **`helix-rtos.PromoteCurrent`** for CPU pinning; eliminates context-switch jitter.
- **`runtime/trace`** — for the per-stage span recording during benchmark runs.
- **`runtime.MemStats` + `runtime.NumGoroutine()`** — for leak detection at boundaries.

---

## 4. Per-Submodule Performance Budgets

Each submodule's [S05 §9.2 *Performance budget*](../06_Submodules/per-submodule/) table is the canonical source of truth for its expected p50/p99/p999. Excerpt of the most operationally-significant ones:

| Submodule           | Metric                                  | p50      | p99     | p999    |
|---------------------|-----------------------------------------|---------:|--------:|--------:|
| helix-r18-safeexec  | `SafeExec` overhead per invocation     | 4 µs     | 6 µs    | 10 µs   |
| helix-grpc-frame    | `Health()` round-trip (localhost)       | 200 µs   | 800 µs  | 2 ms    |
| helix-grpc-frame    | `FrameStream.Send()` enqueue           | 3 µs     | 8 µs    | 15 µs   |
| helix-shm           | `Pool.Acquire()` (warm)                 | 200 ns   | 1 µs    | 5 µs    |
| helix-iouring       | SQE submit (cgo-free, batched)          | 80 ns    | 150 ns  | 200 ns  |
| helix-lockfree      | `SPSCRing.Write()`                      | 4 ns     | 6 ns    | 8 ns    |
| helix-encoder       | NVENC encode 4K120 H.264               | 5 ms     | 7 ms    | 8 ms    |
| helix-pipeline      | End-to-end latency (capture→sink)      | 5 ms     | 7 ms    | 8 ms    |
| helix-transport     | SendVideo (kernel-bypass)               | 100 µs   | 150 µs  | 200 µs  |

The Benchmarking row asserts p999 ≤ baseline p999 × 1.05 for every line above. A regression beyond 5 % blocks the merge; the maintainer either fixes the regression or proposes a baseline-replacement PR with operator sign-off.

---

## 5. Anti-Pattern Catalogue

### 5.1 Reporting only mean / p50

```go
b.Run("encode", func(b *testing.B) {
    b.ReportMetric(meanLatency.Seconds(), "s/op")
})
```

Mean is dominated by the bulk of fast cases; p999 is the tail. HelixPlay's user-perceived latency lives in the tail. Always report p50 + p99 + p999 (helix-bench does this automatically).

### 5.2 Insufficient sample count

```go
for i := 0; i < 100; i++ {  // 100 samples — way too few
    measureOnce()
}
```

Below 10K samples, p999 is dominated by noise (1-in-1000 events need at least 1000 samples). Use `helix-bench.Harness.RunSteadyState(workload, 10000)`.

### 5.3 Cold-cache benchmarking

```go
func BenchmarkEncode(b *testing.B) {
    for i := 0; i < b.N; i++ {
        encode(input)  // first iterations cold; not steady state
    }
}
```

The first ≈ 1000 iterations are cold-cache + warm-up. Use `helix-bench`'s `WithWarmup(samples=1000)` to discard them.

### 5.4 Running on a noisy laptop

```bash
# Developer's MacBook with Chrome + Slack open
$ go test -bench=BenchmarkEncode ./...
```

Noise makes the p999 unreproducible. Benchmarks run on **dedicated CI runners** with helix-rtos pinning. Local `go test -bench` is for sanity, not for the official run-archive.

### 5.5 Storing baseline in code

```go
const expectedP999 = 8 * time.Millisecond  // hardcoded
if observed.P999 > expectedP999 {
    b.Fatal("regression")
}
```

Hardcoding the baseline in code couples the test to one moment in time. Use the run-archive baseline + benchstat-style significance testing.

---

## 6. CI Lane Invocation Pattern

```yaml
- name: Benchmarking smoke (per-PR)
  if: matrix.test-type == 'benchmarking-smoke'
  run: |
    helix-bench -samples=5000 -warmup=500 -workload=$WORKLOAD \
        -baseline=run-archive://baseline-2026-04.json \
        -output=bench-pr.json
    helix-bench compare bench-pr.json baseline-2026-04.json

- name: Benchmarking full (nightly)
  if: matrix.test-type == 'benchmarking' && github.event_name == 'schedule'
  run: |
    helix-bench -samples=10000 -warmup=1000 -workload=$WORKLOAD \
        -baseline=run-archive://baseline-2026-04.json \
        -output=bench-nightly.json
    helix-bench compare bench-nightly.json baseline-2026-04.json
    helixqa-upload bench-nightly.json
```

The smoke variant is per-PR; full is nightly. Both feed the run-archive (per [S04 §7](../06_Submodules/04_HelixQA_Integration.md#7-run-archive-and-audit-trail-immutable-queryable-signed)).

---

## 7. Change-Point Detection (Cross-Run)

Per [helix-bench §2.4](../06_Submodules/per-submodule/helix-bench.md#24-statistical-helpers), `DetectChangePoint(reports)` runs Mann-Whitney U + Kolmogorov-Smirnov over a sequence of run-archive entries to identify the run where a regression first appeared. Use case: a regression is detected on day N; bisect by replaying runs [N-7, N] and identifying the first non-baseline-parity run.

The detection has three thresholds:

- `p > 0.05` → no significant change (noise)
- `0.01 < p ≤ 0.05` → suspected change (warn, don't block)
- `p ≤ 0.01` → confirmed change (block + alert)

The `helixqa-changepoint-report` tool produces a human-readable summary; HelixQA's nightly job invokes it automatically.

---

## 8. Run-Archive Integration

Every nightly + canary + pre-release Benchmarking run produces a run-archive entry per [S04 §7](../06_Submodules/04_HelixQA_Integration.md#7-run-archive-and-audit-trail-immutable-queryable-signed):

```
run-archive/2026-04-30T02:00:00Z/
├── run.manifest.json
├── run.manifest.json.minisig
├── run.metadata.json     # cadence=nightly, submodule=helix-pipeline, workload=...
├── bench/
│   ├── histogram.hgrm    # HDR histogram serialised
│   ├── result.json       # benchstat-compatible
│   └── trace.bin         # runtime/trace event log
└── ci-reports/
    └── benchmarking.log
```

The signed manifest means the run is auditable forever; benchstat compare across runs is the operator's primary regression-tracking interface.

---

## 8a. Sample Benchmark Skeleton

The canonical per-submodule benchmark file shape:

```go
//go:build benchmark
package pipeline_test

import (
    "context"
    "testing"
    "time"

    "github.com/vasic-digital/helix-bench"
    "github.com/vasic-digital/helix-rtos"
)

func BenchmarkPipelineEndToEnd(b *testing.B) {
    ctx := context.Background()

    rt, err := rtos.PromoteCurrent(50)
    if err != nil {
        b.Fatalf("rtos promote: %v", err)
    }
    defer rt.Demote()

    pipe := setupPipeline(b)
    defer pipe.Close()

    h := bench.NewHarness("pipeline-end-to-end",
        bench.WithMinSamples(10000),
        bench.WithMaxDuration(60*time.Second),
        bench.WithWarmup(1000),
        bench.WithCPUPin([]int{4, 5, 6, 7}),
    )

    workload := func() {
        frame := generateSyntheticFrame()
        if err := pipe.Submit(frame); err != nil {
            b.Fatalf("submit: %v", err)
        }
    }

    result, err := h.RunSteadyState(workload, 10000)
    if err != nil {
        b.Fatalf("run: %v", err)
    }

    b.ReportMetric(float64(result.P50), "p50_ns/op")
    b.ReportMetric(float64(result.P99), "p99_ns/op")
    b.ReportMetric(float64(result.P999), "p999_ns/op")

    // Compare to baseline + fail on regression > 5 %
    baseline := bench.LoadBaseline("pipeline-end-to-end")
    verdict, _ := bench.CompareRuns(baseline, &bench.Report{Runs: []bench.Result{*result}})
    if verdict == bench.VerdictRegression {
        b.Fatalf("p999 regression: %v", result.P999)
    }
}
```

Properties:
- `helix-rtos.PromoteCurrent(50)` pins the goroutine to SCHED_FIFO + a fixed CPU set; eliminates context-switch jitter.
- 10K samples (Insight #2 minimum); 1K warmup discarded.
- `bench.WithCPUPin([4,5,6,7])` keeps the workload off the runner's housekeeping CPUs.
- `bench.CompareRuns` is the regression gate (Mann-Whitney U + KS).

## 8b. Baseline Replacement Procedure

Per [S03 §6.1](../06_Submodules/03_Challenges_Submodule.md#61-recording), baselines are operator-signed; replacement is non-delegable. The procedure:

1. Maintainer detects an intentional improvement (e.g. encoder upgrade reduces p999 by 12 %).
2. Maintainer opens a PR to update the baseline file at `vasic-digital/Challenges/baselines/<topology>/<scenario>/baseline.manifest.json`.
3. PR body documents:
   - The improvement (which submodule + which workload).
   - The verification method (run-archive entry IDs that demonstrated the improvement reproducibly).
   - The new baseline values + their derivation.
4. Operator reviews + signs the new baseline.minisig with the operator's minisign key.
5. Merge + the new baseline is canonical.

A baseline regression (worsening) cannot be "replaced" by this procedure — the regression must be fixed in code first. Constitution §15 *Amendment Procedure* applies if the operator believes the regression is acceptable for a documented operational reason.

## 9. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T06-A            | Per-PR smoke samples (5K) vs nightly full (10K) — operator can re-tune?                                       | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T06-B            | RDTSC vs CLOCK_MONOTONIC vs perf_event — canonical timer?                                                     | C24 §6 next revision                                |
| OQ-T06-C            | Cross-architecture comparability (amd64 vs arm64 baselines) — separate ranges or unified scoring?            | C24 §6 next revision                                |

---

## 10. References & Anti-Bluff Verification

### 10.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.5.
- [`../06_Submodules/per-submodule/helix-bench.md`](../06_Submodules/per-submodule/helix-bench.md).
- [`../04_Latency/10_Latency_Testing_and_Validation.md`](../04_Latency/10_Latency_Testing_and_Validation.md) §6 (change-point methodology).

### 10.2 External (web)

- benchstat: https://pkg.go.dev/golang.org/x/perf/cmd/benchstat (accessed 2026-04-30).
- HDR histogram: http://hdrhistogram.org/ (accessed 2026-04-30).
- Mann-Whitney U test: https://en.wikipedia.org/wiki/Mann%E2%80%93Whitney_U_test (accessed 2026-04-30).
- Kolmogorov-Smirnov test: https://en.wikipedia.org/wiki/Kolmogorov%E2%80%93Smirnov_test (accessed 2026-04-30).

### 10.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §9 open questions named with deferred resolution chapters.
- §5 anti-pattern catalogue codifies the operational failure modes Benchmarking exists to prevent.
- §7 change-point detection cross-references the C24 §6 source methodology + helix-bench §2.4 implementation.

### 10.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/06_Benchmarking.md` — 2026-04-30.
