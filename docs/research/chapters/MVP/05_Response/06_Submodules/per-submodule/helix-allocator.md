# `helix-allocator` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-allocator`                                                                                                      |
| **Origin chapter:section**  | [C23 §6](../../04_Latency/09_Memory_and_Cache_Optimization.md) — *Allocation-free hot-path enforcer*                  |
| **Public path (4 mirrors)** | `vasic-digital/helix-allocator` on GitHub + GitLab + GitFlic + GitVerse                                                |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-mempool`                                                                              |
| **External Go deps**        | `runtime/debug`, `runtime/trace`, `golang.org/x/tools/go/analysis` (build-time)                                         |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `allocator-enforcer-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                    |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/02_multi_session_single_host/scenarios/04_allocator_enforces_hot_path_zero_alloc.scenario.yaml`         |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-allocator` is the **hot-path allocation-free enforcer** that complements `helix-mempool`. Origin: [C23 §6](../../04_Latency/09_Memory_and_Cache_Optimization.md). The submodule provides:

- **A static analyser** (Go `go/analysis` plugin) that scans hot-path functions for `make` / `new` / map-literal / append-on-nil patterns and fails CI if found.
- **A runtime guard** that wraps the hot-path goroutine with `runtime.SetFinalizer`-based escape analysis and reports any unintended GC allocation.
- **Build-time tags** (`//helix:nohotalloc`) consumed by both the analyser and runtime guard.

The submodule was introduced because `helix-mempool` alone does not prevent regressions — a future PR can re-introduce inline `make` calls. The allocator enforcer is the active discipline that keeps the hot path allocation-free across the project's lifetime. Insight #4 of the Latency family (codified in [C13 §6](../../03_Architecture/12_Latency_Engineering_Overview.md)) demands this guarantee.

---

## 2. Public API Surface

### 2.1 The `HotPath` annotation marker

```go
package allocator

// HotPath is a no-op runtime marker that the analyser recognises;
// any function starting with HotPath() is checked for forbidden
// allocations.
//
//   func encodeFrame(f *Frame) {
//       allocator.HotPath()
//       // ... must not contain make / new / map-literal / append-on-nil
//   }
func HotPath()
```

### 2.2 The `EnforceNoAlloc` runtime guard

```go
package allocator

// EnforceNoAlloc starts a runtime allocation tracker on the
// caller's goroutine. Any GC allocation triggers a panic (in
// Strict mode) or a metric emission (in Report mode).
type Mode int
const (
    ModeReport Mode = iota
    ModeStrict
)

func EnforceNoAlloc(mode Mode) Cancel
type Cancel func()
```

### 2.3 The static analyser entry point

```go
package allocator

// Analyser is the Go static analyser plugin. Registered with
// `go vet` via:
//   //go:build allocator
//   package allocator
//   var _ = analysis.Analyser{...}
//
// CI-side invocation:
//   go vet -vettool=$(which helix-allocator-vet) ./...
func Analyser() *analysis.Analyser
```

### 2.4 Build-tag helpers

```go
package allocator

// HotAllocAllowed signals (via build tag) that the function may
// allocate. Consumers should use //helix:nohotalloc to opt OUT
// individual functions; HotAllocAllowed is the inverse for cases
// where the caller knows what they're doing.
const HotAllocAllowed = "//helix:hotalloc-allowed"
const NoHotAlloc      = "//helix:nohotalloc"
```

### 2.5 Runtime allocation reporter

```go
package allocator

type AllocReport struct {
    GoroutineID  uint64
    AllocsBefore uint64
    AllocsAfter  uint64
    Diff         int64
    Stack        []byte
    Timestamp    time.Time
}

func ReportChannel() <-chan AllocReport
```

### 2.6 Statistics

```go
package allocator

type EnforcerStats struct {
    HotPathFunctions    int      // total annotated functions in this binary
    ViolationsAtRuntime uint64   // ModeReport / ModeStrict violations
    BuildTimeRejections int      // CI failures (recorded historically)
}
```

### 2.7 The build-time CLI tool

```go
package allocator

// CLI command: helix-allocator-vet
//
// Usage:
//   helix-allocator-vet [flags] [packages]
//
// Flags:
//   -strict    Fail on any violation (default).
//   -report    Print violations but exit 0 (for inventory).
//   -emit-trace Emit runtime/trace event on violation (dev mode).
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`go env GOROOT` for analyser tool path).
- `helix-mempool` — every consumer of `helix-allocator` is expected to use `helix-mempool` for actual allocation.

### 3.2 External (Go)

- `runtime/debug`, `runtime/trace` — for the runtime guard.
- `golang.org/x/tools/go/analysis` (build-time) — analyser framework.

### 3.3 External (system)

None at runtime; pure-Go.

---

## 4. Container Build (S02 §3 lane: `allocator-enforcer-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2.

The CI lane additionally builds the `helix-allocator-vet` binary as a release artefact attached to every per-submodule release; this binary is consumed by every other submodule's `go vet` step.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Analyser correctly detects `make`/`new`/map-literal/append-on-nil in synthetic ASTs.

### 5.2 Integration
Run analyser on a real consumer (e.g. `helix-pipeline` test fixtures); verify expected violations / non-violations.

### 5.3 E2E
helix-allocator-vet → real submodule's source tree → CI pass/fail.

### 5.4 Security
govulncheck + Snyk + Trivy.

### 5.5 Benchmarking
Analyser walk speed: ≥ 100 K LOC / second on amd64.

### 5.6 Chaos
Inject GC pause; verify the runtime guard surfaces correctly without amplifying the pause.

### 5.7 Stress
24-hour ModeReport on a hot-path goroutine; zero false positives.

### 5.8 Smoke
30-second post-deploy: vet a known-clean submodule; expect exit 0.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`02_multi_session_single_host/04_allocator_enforces_hot_path_zero_alloc.scenario` — full system with ModeStrict on every hot-path goroutine; verify zero strict-mode panics over the scenario duration.

---

## 6. Challenges Entry-Point (S03 §4 row #16)

**Topology:** `02_multi_session_single_host`. **Scenario:** `04_allocator_enforces_hot_path_zero_alloc.scenario.yaml`. **Why this scenario.** ModeStrict is the strongest discipline; the scenario verifies the entire fleet stays allocation-free under realistic load. **Baseline:** zero ViolationsAtRuntime; OTLP trace shows EnforceNoAlloc spans wrapping every hot-path goroutine.

---

## 7. R-18 Inheritance

`helix-allocator` imports `helix-r18-safeexec` for the `helix-allocator-vet` CLI's `go env` invocation. Hot path is pure Go.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the analyser ruleset stability — every new rule risks false positives, so the graduation gate requires zero-false-positives on the canonical fleet.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_ALLOCATOR_MODE`             | `report`      | `report` / `strict`     | Runtime guard mode.                                                     |
| `HELIX_ALLOCATOR_TRACE_ENABLED`    | `false`       | bool                    | Emit runtime/trace events on violation.                                |
| `HELIX_ALLOCATOR_VET_PARALLEL`     | `auto`        | int / `auto`            | Parallel analyser workers; `auto` = NumCPU.                            |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Analyser walk per file                | 5 ms     | 15 ms   | 30 ms   | Typical 1 K-LOC Go file.                                         |
| EnforceNoAlloc setup overhead         | 5 µs     | 15 µs   | 30 µs   | Once per goroutine.                                              |
| Runtime allocation tracker overhead   | 2 ns     | 5 ns    | 10 ns   | Per allocation; ModeReport.                                      |
| ModeStrict panic-overhead (violation) | (unmeasured) | — | —     | Panic terminates the goroutine; intentional.                     |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `allocator: ErrAnalyserViolation`             | Hot-path function contains forbidden allocation                | Replace with `helix-mempool` Pool acquire/release.                         |
| `allocator: ErrFalsePositive`                 | Analyser flagged a benign pattern                              | Audit the pattern; if benign, file an issue + add `//helix:hotalloc-allowed`.|
| `allocator: ErrModeStrict panicked`           | Runtime ModeStrict violation                                   | Capture the stack from the panic; trace back to the offending allocation site. |

### 9.4 Migration to allocator discipline

A consumer adopting `helix-allocator`:

1. Mark the hot-path entry function with `allocator.HotPath()` as the first call.
2. Run `helix-allocator-vet ./...` locally; fix all violations.
3. Add `helix-allocator-vet` to the per-submodule CI lane (S02 §3.3 contract).
4. Wrap the hot-path goroutine startup with `cancel := allocator.EnforceNoAlloc(allocator.ModeReport); defer cancel()`.
5. After 2 release cycles of zero ModeReport violations, escalate to ModeStrict.

The migration is documented in `docs/migration-to-allocator-discipline.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_allocator_violations_runtime_total`      | counter    | Runtime guard violations, labelled `function`, `mode`.                       |
| `helix_allocator_vet_runs_total`                | counter    | Analyser runs, labelled `result={pass, fail}`.                               |
| `helix_allocator_vet_files_scanned`             | counter    | Files analysed.                                                              |
| `helix_allocator_hotpath_count`                 | gauge      | Annotated hot-path functions in this binary.                                |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C23 §6](../../04_Latency/09_Memory_and_Cache_Optimization.md) — origin                    | Origin chapter; full enforcer API.                                                   |
| [C13 §6](../../03_Architecture/12_Latency_Engineering_Overview.md) — Latency overview      | Insight #4 codification.                                                             |
| **Every** depth-2+ submodule with a hot path                                              | Adopts the discipline via per-submodule CI lane.                                     |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-allocator-A      | False-positive whitelist — should it live in this submodule or in each consumer's `.allocator.yaml`?           | C23 §6 next revision                                |
| OQ-allocator-B      | ModeStrict adoption timeline — fleet-wide v1.0.0 gate or per-submodule choice?                                | `09_Implementation_Phases/Phase_02_Core_Submodules.md` |
| OQ-allocator-C      | Generics overhead detection — does the analyser need to flag boxing patterns?                                  | C23 §6 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/09_Memory_and_Cache_Optimization.md`](../../04_Latency/09_Memory_and_Cache_Optimization.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-mempool.md`](helix-mempool.md) | (this batch) | 2026-04-30 | direct dependency                              |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage; analyser detects every documented pattern.                        |
| Integration    | helix-allocator-vet on real consumer source.                                                  |
| E2E            | CI lane reject-on-violation works.                                                            |
| Security       | govulncheck + Snyk + Trivy.                                                                   |
| Benchmarking   | Analyser ≥ 100 K LOC/s.                                                                        |
| Chaos          | GC-pause injection; runtime guard surfaces cleanly.                                           |
| Stress         | 24-hour ModeReport; zero false positives.                                                     |
| Smoke          | 30-second clean-vet exit-0.                                                                   |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `02_multi_session_single_host/04_allocator_enforces_hot_path_zero_alloc` baseline-parity.     |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-allocator.md` — 2026-04-30.
