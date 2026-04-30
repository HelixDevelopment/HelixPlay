# T08 — Stress Tests

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.7; [`../06_Submodules/per-submodule/helix-bench.md`](../06_Submodules/per-submodule/helix-bench.md) (24-hour soak harness); [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §5 (`11_stress_24h_steady_state` topology).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row T08.
> **Chapter targets:** R-11 (24-hour soak; zero leak), R-13.
> **Cross-links:** [`06_Benchmarking.md`](06_Benchmarking.md), [`11_Challenges.md`](11_Challenges.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Stress is **test-type 7 of 10**. Cadence: canary (24-hour subset) + pre-release (full 24-hour). Per-PR + nightly do **not** run Stress — its 24-hour wall-clock is incompatible with those cadences. Stress is the canary's defining cost (per [T01 §14a.3](01_Test_Matrix.md#14a-the-1160-cell-cost-model)).

Stress's role: **prove the system doesn't leak resources or degrade under sustained peak load**. Where Benchmarking measures p999 latency over 10K samples, Stress runs at peak workload for 24 hours (≈ 10⁸ samples at 1K req/s) and asserts no resource grows linearly with time. The two test types are complementary: Benchmarking catches *current* regressions; Stress catches *latent* leaks.

---

## 2. The Discipline — What Makes "Green" Green

A green Stress row means:

1. **24-hour run at peak documented rate** without crash, hang, or error escalation.
2. **Zero memory growth** — `runtime.MemStats` heap-in-use at the 24-h mark ≤ 1.05× the 1-h mark (5 % drift tolerance for cache fill-up).
3. **Zero fd leak** — `/proc/<pid>/fd` count at 24-h ≤ 1-h count.
4. **Zero goroutine leak** — `runtime.NumGoroutine()` at 24-h ≤ 1-h count.
5. **No ABR descent** — bitrate sustained at the workload's declared peak (otherwise the test is measuring the descent, not the steady state).
6. **OTLP traces continuous** — no trace gap > 60 seconds (gaps indicate the trace exporter died, even if the process kept going).

The test invocation samples the four leak-detector metrics every 1 minute over the 24-hour window. Linear regression on each series detects growth; a slope > 0 with `p ≤ 0.05` (statistically significant growth) fails the test.

---

## 3. Tooling

- **`helix-bench`** in 24-h mode: `helix-bench -duration=24h -workload=$WORKLOAD -interval=1m`.
- **`runtime/pprof`** — heap profiles every 4 hours for postmortem analysis.
- **`runtime/trace`** — start-of-day + end-of-day trace recordings for execution-trace diff.
- **`/proc/<pid>/{fd,status,smaps}`** — kernel-side resource counts.
- **`vasic-digital/Challenges/topologies/11_stress_24h_steady_state/docker-compose.yml`** — the canonical 24-h stress topology.
- **`helix-leak-regress`** — Mann-Whitney + linear-regression across the 1-min samples; fail if `p ≤ 0.05` slope > 0.

---

## 4. Per-Submodule Stress Targets

Each submodule's [S05 §11 Per-test-type coverage targets](../06_Submodules/per-submodule/) Stress row commits to:

- **`helix-r18-safeexec`** — 24-hour 10K invocations/s of allowed commands; zero memory growth, zero fd leak, zero goroutine leak.
- **`helix-shm`** — 24-hour 60 fps × 4K page allocate/release; zero fragmentation; PeakInUse stable.
- **`helix-iouring`** — 24-hour 60 Hz × 1024 SQE batches; zero fd leak; CQE-overflow counter stable.
- **`helix-pipeline`** — 24-hour pipeline run; per-stage histograms stable; end-to-end latency p999 within budget at 24 h same as 1 h.
- **`helix-record`** — 24-hour recording session; storage upload backlog drains; zero local-disk runaway.
- **`helix-vault`** — 24-hour 1 K encrypt+decrypt operations/s; zero leak.

---

## 5. Anti-Pattern Catalogue

### 5.1 24-hour test that uses up too much disk

A test that records 24 h of 4K video will fill 100s of GB of disk. Either:
- Use `helix-record` with `KEEP_LOCAL_HOURS=1` so the local disk stays bounded, OR
- Stream to a `dev/null`-equivalent sink via `record.NewRecorder(record.FormatfMP4, "dev-null:///")` for stress purposes.

### 5.2 Stress that depends on calendar time

```go
deadline := time.Now().Add(24 * time.Hour)
for time.Now().Before(deadline) {
    // ...
}
```

Wall-clock time can drift (NTP adjustments). Use `time.NewTicker(1*time.Minute)` + a counter for the 24-h horizon. Better: use `helix-bench`'s `-duration` flag — it normalises this.

### 5.3 Stress that doesn't sample resources

```go
// only at end-of-test
finalStats := runtime.ReadMemStats(...)
if finalStats.HeapInUse > 1*GiB { /* fail */ }
```

A single end-of-test sample says nothing about the trajectory. The growth from 0 → 1 GiB can happen in the first minute (large initial cache) or steadily (a leak); the test must distinguish. Sample every 1 minute.

### 5.4 Stress that runs against shared infrastructure

A 24-hour run on a CI runner that's also serving other PRs creates noise + resource contention. Stress runs on **dedicated** runners (operator-provisioned per [S04 §6](../06_Submodules/04_HelixQA_Integration.md#6-multi-mirror-orchestration-across-the-four-mirror-ci-topology)).

### 5.5 Stress that resumes from a checkpoint

```go
if checkpoint exists:
    skipFirstNHours()
```

A 24-hour soak from a 12-hour checkpoint says nothing about the first 12 hours. Always start fresh.

---

## 6. CI Lane Invocation Pattern

```yaml
- name: Stress (canary subset — 1-hour run)
  if: matrix.test-type == 'stress' && github.event_name == 'workflow_dispatch'
  timeout-minutes: 75   # 1-h test + 15-min margin for setup/teardown
  steps:
    - name: Start 11_stress_24h_steady_state topology
      run: |
        cd vasic-digital/Challenges/topologies/11_stress_24h_steady_state
        docker compose up -d --wait
    - name: Run 1-hour stress
      run: |
        helix-bench -duration=1h -interval=1m -workload=$WORKLOAD \
            -leak-check=true \
            -output=stress-canary.json
        helix-leak-regress stress-canary.json
    - name: Upload to run-archive
      run: helixqa-upload stress-canary.json

- name: Stress (pre-release — full 24-hour)
  if: matrix.test-type == 'stress' && github.ref_type == 'tag'
  timeout-minutes: 1500  # 25-h test
  steps:
    - name: Run full 24-hour stress
      run: |
        helix-bench -duration=24h -interval=1m -workload=$WORKLOAD \
            -leak-check=true \
            -pprof-interval=4h \
            -trace-bookend=true \
            -output=stress-prerelease.json
        helix-leak-regress stress-prerelease.json
```

The canary 1-h subset is the **gate** for canary cadence; the pre-release 24-h is the **gate** for `v1.0.0+` tag publication.

---

## 7. Resource-Leak Detection Methodology

The four leak metrics + their detection thresholds:

| Metric                          | Sampling             | Pass criterion                                                |
|---------------------------------|----------------------|----------------------------------------------------------------|
| Heap in-use (`MemStats.HeapInuse`) | every 1 min       | linear-regression slope's `p > 0.05` OR slope ≤ 0.5 % per hour |
| File descriptors (`/proc/<pid>/fd`) | every 1 min      | regression slope's `p > 0.05` (zero-tolerance for fd growth) |
| Goroutines (`runtime.NumGoroutine`) | every 1 min     | regression slope's `p > 0.05` OR slope ≤ 1 goroutine per hour |
| RSS (`/proc/<pid>/status`)        | every 1 min       | regression slope's `p > 0.05` OR slope ≤ 1 % per hour        |

The 1 % / 0.5 % per-hour tolerances accommodate a slowly-warming cache (typical for LRU eviction-driven caches that take hours to reach steady state). Strict zero-growth would fail false-positively; the regressional p-value test distinguishes growth from noise.

---

## 8. Pprof + Trace Postmortem

Stress runs emit `runtime/pprof` heap profiles every 4 hours (so the 24-h run has 6 profiles) + a `runtime/trace` recording at start-of-day and end-of-day. Post-test analysis:

- `pprof -base profile-04h.heap profile-24h.heap` — 20-hour heap diff.
- `go tool trace trace-eod.bin` — execution-trace summary.

The artifacts ship with the run-archive entry (per [S04 §7](../06_Submodules/04_HelixQA_Integration.md#7-run-archive-and-audit-trail-immutable-queryable-signed)) so postmortem on a Stress red is fully reproducible.

---

## 8a. The 24-Hour Soak Schedule

The pre-release 24-hour soak is the defining wall-clock cost of the pre-release cadence. The schedule:

| Hour | Action                                                                                       |
|------|----------------------------------------------------------------------------------------------|
| 0    | Topology brought up. helix-rtos-promoted hot-path goroutine started.                          |
| 0–1  | Cache fill-up window — slope detection allows up to 1 % heap growth per hour.                |
| 1    | First leak-detection sample. Set the linear-regression baseline.                             |
| 1–24 | Sample every 1 minute (1380 samples per metric). p999 latency tracked continuously.          |
| 4    | First runtime/pprof heap profile recorded.                                                   |
| 8    | Second pprof + runtime/trace bookmark.                                                       |
| 12   | Third pprof.                                                                                  |
| 16   | Fourth pprof.                                                                                 |
| 20   | Fifth pprof.                                                                                  |
| 24   | Final pprof + end-of-day trace recording. Final leak-detection regression.                   |
| 24+  | Tear-down + run-archive upload + helix-leak-regress final report.                            |

The 1-h cache-fill-up exemption ensures benign LRU warm-up doesn't fail the regression. Slope detection's linear-regression p-value test handles statistical noise distinct from genuine growth.

## 8b. Per-Workload Stress Targets

The §11_stress_24h_steady_state topology runs each submodule's documented peak workload. Specific peak rates (per the per-submodule §9.2 *Performance budget* tables):

| Submodule         | Peak workload (24-h sustained)                            |
|-------------------|-----------------------------------------------------------|
| helix-r18-safeexec | 10 K SafeExec invocations/sec                            |
| helix-shm         | 60 fps × 4K NV12 (≈ 720 MiB/s page allocate+release)    |
| helix-iouring     | 60 Hz × 1024 SQE batches                                  |
| helix-lockfree    | 1 M SPSC ops/sec                                          |
| helix-encoder     | 4K @ 120 fps NVENC continuous                             |
| helix-pipeline    | full pipeline at 4K120 + 7.1 audio                        |
| helix-record      | 4K120 fMP4 + S3 sync continuous                           |
| helix-vault       | 1 K encrypt+decrypt/sec                                   |
| helix-transport   | 10 Gbps RDMA send sustained                               |

The peaks are operator-tunable per [OQ-T08-A](#9-open-questions); the defaults are the documented MVP targets.

## 9. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T08-A            | Per-submodule peak rate — operator-tunable or fixed?                                                          | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T08-B            | Stress runner provisioning — dedicated long-running VMs or auto-spawned per run?                              | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T08-C            | Slope-detection p-value threshold (0.05 vs stricter)?                                                          | T08 next revision                                   |

---

## 9a. Soak-Test Findings Pattern (postmortem template)

When a 24-hour soak fails, the postmortem follows a fixed template stored at `vasic-digital/.github/soak-postmortem-template.md`:

```markdown
# Soak Postmortem — <run-id>

## Submodule + Workload
- Submodule: <name>
- Workload: <peak-rate-config>
- Run ID: <YYYY-MM-DDTHH:MM:SSZ>

## Detection
- Hour: <H>
- Metric: <heap | fd | goroutine | rss>
- Slope: <bytes-per-hour or count-per-hour> with p-value <p>
- Threshold breached: <yes/no>

## Root Cause
- Identified via: <pprof diff h4 vs h24 | trace inspection | code audit>
- Description: <free-form>

## Fix
- PR: <url>
- Code change: <free-form>
- Verification: <new-soak-run-id confirming the fix>

## Why the Slope-Detection Worked / Didn't
- Statistical analysis of the regression that caught the leak.
- (If the leak escaped: gap analysis + corrective action on detection.)

## Action Items
- [ ] Add a regression unit test that exercises the leak path.
- [ ] Update the per-submodule §9.3 *Common errors* table with the new failure mode.
- [ ] Update §4 *Per-Submodule Stress Targets* if the workload assumptions changed.
```

Postmortems are committed under `vasic-digital/.github/postmortems/soak/` with append-only retention.

## 9b. The Stress-vs-Bench Boundary

A common confusion: when does measurement belong in T06 (Benchmarking) and when in T08 (Stress)? The boundary:

- **T06 measures latency under steady state.** 10K samples; 1-h max wall-clock; p999 latency vs baseline.
- **T08 measures resource trajectory under sustained load.** 24-hour wall-clock; samples every 1 min; slope detection on heap/fd/goroutine/rss.

A test that captures p999 latency at hour 0 and hour 24 is using **both** — the bench harness records the latency, and the stress harness records the resource trajectory. They share the same workload generator (`helix-bench` is the harness for both); they differ in *what they assert* about the workload.

If a regression appears at hour 12 of a soak (p999 spikes), it's an **emergent** property — likely a slow-leak that hits a tipping point. T08's slope detection catches the leak preceding the spike; T06 catches the spike itself. Both reports go to the same run-archive entry.

## 10. References & Anti-Bluff Verification

### 10.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.7.
- [`../06_Submodules/per-submodule/helix-bench.md`](../06_Submodules/per-submodule/helix-bench.md).
- [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §5 (canonical 24-h topology).

### 10.2 External (web)

- runtime/pprof: https://pkg.go.dev/runtime/pprof (accessed 2026-04-30).
- runtime/trace: https://pkg.go.dev/runtime/trace (accessed 2026-04-30).

### 10.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §9 open questions named with deferred resolution chapters.
- §5 anti-pattern catalogue + §7 detection methodology operationalise the regression-vs-noise distinction.

### 10.4 Stress-vs-Bench Boundary cross-reference

The §9b boundary clarifies the role split between T06 and T08; the same boundary appears in [`06_Benchmarking.md`](06_Benchmarking.md) §1 (where Benchmarking explicitly disclaims the 24-h soak scope as Stress's territory). Together the two chapters operationalise R-13: *current* regressions caught by Bench's p999 gate; *latent* leaks caught by Stress's slope-detection gate.

A submodule with a pristine Bench p999 but a failing Stress slope is shipping a leak that hasn't yet hit the tipping point; the operator typically discovers this at the canary cadence's 1-h subset run + escalates to fix-before-pre-release.

The 1-h canary subset is intentional triage — it catches obvious leaks (heap doubling per hour) cheaply before committing to the full 24-hour pre-release run. A pristine 1-h canary does **not** guarantee a pristine 24-h pre-release; the slope-detection p-value test gets stronger with more samples, so a slow leak that's statistical noise at hour 1 may become significant at hour 24. This is why the pre-release cadence runs the full 24-hour soak — it's the only cadence with statistical power to catch slow leaks.

### 10.5 Cross-Reference Catalogue

The Stress chapter's content references — for the operator's grep convenience:

- 24-hour topology: [`vasic-digital/Challenges/topologies/11_stress_24h_steady_state/`](../06_Submodules/03_Challenges_Submodule.md#3-repository-layout-topologies--baselines--harness)
- Bench harness API: [`helix-bench.RunSteadyState`](../06_Submodules/per-submodule/helix-bench.md#21-the-harness-type)
- Per-submodule peak rates: each submodule's [S05 §9.2 *Performance budget*](../06_Submodules/per-submodule/) table
- Postmortem template: `vasic-digital/.github/soak-postmortem-template.md` (per §9a)
- Slope-detection algorithm: `helix-leak-regress` per [`helix-bench` §2.4](../06_Submodules/per-submodule/helix-bench.md#24-statistical-helpers)

### 10.6 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/08_Stress.md` — 2026-04-30.
