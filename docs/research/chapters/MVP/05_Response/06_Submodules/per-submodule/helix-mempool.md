# `helix-mempool` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-mempool`                                                                                                        |
| **Origin chapter:section**  | [C23 §3](../../04_Latency/09_Memory_and_Cache_Optimization.md) — *Memory pool + arena allocator for hot-path objects*  |
| **Public path (4 mirrors)** | `vasic-digital/helix-mempool` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `golang.org/x/sys/cpu` (cache-line size)                                                                                |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `mempool-arena-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                         |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/02_multi_session_single_host/scenarios/03_mempool_no_alloc_in_hot_path.scenario.yaml`                    |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-mempool` provides **typed memory pools and arena allocators** for HelixPlay's hot-path objects. Origin: [C23 §3](../../04_Latency/09_Memory_and_Cache_Optimization.md). The submodule's role is to keep the hot path **allocation-free**: every per-frame and per-input-event object is acquired from a pre-allocated pool rather than `make`'d, eliminating GC pressure that would introduce 100s of µs of pause on the gameplay round-trip.

The submodule pairs with `helix-allocator` (which enforces "no make / no new in hot path") to form the allocation-free discipline mandated by Insight #4 of the Latency family ([C13 §6](../../03_Architecture/12_Latency_Engineering_Overview.md)).

---

## 2. Public API Surface

### 2.1 The `Pool[T]` type

```go
package mempool

// Pool is a typed sync.Pool-style allocator with HelixPlay-specific
// cache-line padding and pre-allocation. Acquire and Release are
// allocation-free in the steady state.
type Pool[T any] struct { /* ... */ }

func NewPool[T any](capacity int, factory func() *T, opts ...Option) *Pool[T]
func (p *Pool[T]) Acquire() *T
func (p *Pool[T]) Release(v *T)
func (p *Pool[T]) Stats() PoolStats
func (p *Pool[T]) Close()
```

### 2.2 The `Arena` type

```go
package mempool

// Arena is a bump allocator for short-lived heterogeneous objects.
// All allocations within an Arena are freed together via Reset.
type Arena struct { /* ... */ }

func NewArena(initialSize int, opts ...ArenaOption) *Arena
func ArenaAlloc[T any](a *Arena) *T
func ArenaAllocSlice[T any](a *Arena, n int) []T
func (a *Arena) Reset()
func (a *Arena) Bytes() int
func (a *Arena) Close()
```

### 2.3 The `ByteBufferPool`

```go
package mempool

// ByteBufferPool is a specialised pool for variable-sized byte
// slices, organised in size classes (powers of two from 1 KiB to
// 16 MiB).
type ByteBufferPool struct { /* ... */ }

func NewByteBufferPool(opts ...Option) *ByteBufferPool
func (p *ByteBufferPool) Acquire(size int) []byte
func (p *ByteBufferPool) Release(buf []byte)
```

### 2.4 The `Slab` type

```go
package mempool

// Slab is a fixed-size-block slab allocator; suitable for
// homogeneous blocks where Pool's value-type semantics impose
// overhead.
type Slab struct { /* ... */ }

func NewSlab(blockSize int, capacity int) *Slab
func (s *Slab) Allocate() unsafe.Pointer
func (s *Slab) Free(ptr unsafe.Pointer)
```

### 2.5 Pool options

```go
package mempool

type Option func(*config)

func WithPreAllocation(n int) Option        // pre-fill at construction
func WithCacheLinePadding() Option           // pad each entry to a cache line
func WithReleaseHook(fn func(any)) Option    // hook to reset state
```

### 2.6 Statistics

```go
package mempool

type PoolStats struct {
    Capacity       int
    InUse          int
    AcquireTotal   uint64
    ReleaseTotal   uint64
    GCAllocations  uint64    // pool growth events; alert if non-zero in steady state
    PeakInUse      int
}
```

### 2.7 The `MetricRecorder` interface

```go
package mempool

// MetricRecorder is implemented by the consumer's metrics layer; the
// pool emits acquire/release/peak counts via this interface.
type MetricRecorder interface {
    RecordAcquire(poolName string)
    RecordRelease(poolName string)
    RecordGCAllocation(poolName string)
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (querying `/proc/sys/vm/overcommit_memory`).

### 3.2 External (Go)

- `golang.org/x/sys/cpu` — cache-line size detection.

### 3.3 External (system)

None at runtime; pure-Go.

---

## 4. Container Build (S02 §3 lane: `mempool-arena-1.x`)

**Builder:** `golang-builder-cgo` (cgo for the slab unsafe.Pointer interactions). **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Pool + Arena + ByteBufferPool + Slab acquire/release; allocation-counter assertions via `testing.AllocsPerRun`.

### 5.2 Integration
Real concurrent goroutines hammering Pool[Frame] at 60 fps × 1000 frames; verify zero GC growth.

### 5.3 E2E
helix-mempool → helix-pipeline frame buffer lifecycle.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: ABI-fuzzer for Slab unsafe.Pointer paths.

### 5.5 Benchmarking
Pool.Acquire (warm) p999 ≤ 50 ns; Arena.Alloc p999 ≤ 30 ns; ByteBufferPool size-class lookup p999 ≤ 100 ns.

### 5.6 Chaos
Inject GC pause; verify pool invariants hold across GC.

### 5.7 Stress
24-hour 60 fps × 100 K acquire/release cycles; zero GC allocations after warmup; PeakInUse stable.

### 5.8 Smoke
30-second post-deploy: acquire + release 1 K Frame objects; verify zero allocations beyond pre-allocation.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`02_multi_session_single_host/03_mempool_no_alloc_in_hot_path.scenario` — concurrent sessions with `runtime/trace` enabled; verify zero hot-path allocations during the steady-state phase.

---

## 6. Challenges Entry-Point (S03 §4 row #15)

**Topology:** `02_multi_session_single_host`. **Scenario:** `03_mempool_no_alloc_in_hot_path.scenario.yaml`. **Why this scenario.** The allocation-free hot-path is the strictest invariant in the Latency family; the Challenges scenario verifies it under realistic concurrent-session load. **Baseline:** zero hot-path allocations (verified via `runtime/trace`); GCAllocations counter remains zero.

---

## 7. R-18 Inheritance

`helix-mempool` imports `helix-r18-safeexec` for boundary subprocess invocations (kernel-feature detection). Hot path is pure Go (Pool/Arena) or pure unsafe.Pointer (Slab); no subprocess.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the C23 §3 API freeze + 2 consecutive green Ten-test cycles, including a runtime/trace-based hot-path-allocation-zero invariant.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_MEMPOOL_PREALLOC`           | `64`          | int [4, 16384]          | Default pre-allocation count per Pool.                                  |
| `HELIX_MEMPOOL_CACHE_LINE_PAD`     | `true`        | bool                    | Enable cache-line padding in pools.                                    |
| `HELIX_MEMPOOL_TRACK_PEAK`         | `true`        | bool                    | Track PeakInUse for capacity tuning.                                  |
| `HELIX_MEMPOOL_DEBUG_LEAK_SCAN`    | `false`       | bool                    | Track Acquire/Release pairing; report leaks at Close (dev only).      |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Pool.Acquire (warm)                   | 15 ns    | 30 ns   | 50 ns   | sync.Pool-style fast path; one atomic-load.                      |
| Pool.Acquire (cold)                   | 200 ns   | 500 ns  | 1 µs    | Triggers factory function.                                        |
| Pool.Release                          | 15 ns    | 30 ns   | 50 ns   | One atomic-store.                                                 |
| Arena.Alloc                           | 10 ns    | 20 ns   | 30 ns   | Bump pointer; no atomic.                                          |
| ByteBufferPool.Acquire                | 30 ns    | 80 ns   | 100 ns  | Size-class lookup + Pool.Acquire.                                |
| Memory per Pool[Frame] (cap=64)       | varies   | —       | —       | Depends on Frame size; typical 64 × 16 KiB = 1 MiB.              |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| Non-zero `GCAllocations` in steady state      | Pool capacity exceeded; growth triggered                       | Raise `HELIX_MEMPOOL_PREALLOC` or capacity argument.                       |
| Acquire returns nil unexpectedly              | Factory returned nil                                           | Audit factory function; ensure non-nil return.                             |
| Memory growth despite Release                 | `T` contains pointer fields not zeroed on Release             | Add `WithReleaseHook` to clear pointer fields; prevents reachable garbage. |
| Slab Free on wrong allocator                  | Cross-allocator Free                                           | Audit the consumer; the Slab uses ABI bookkeeping, mismatched Free panics. |
| Race detected in Pool                         | Concurrent Acquire/Release without proper synchronisation     | Pool is safe for concurrent use; race indicates a bug — audit the access pattern. |

### 9.4 Migration from `make`/`new` per-frame

A consumer migrating from `frame := &Frame{...}` per call:

1. Identify the hot-path types (`Frame`, `RTPPacket`, `OTLPSpan`, etc.).
2. Define a factory: `func newFrame() *Frame { return &Frame{...} }`.
3. Construct a Pool: `pool := mempool.NewPool[Frame](64, newFrame)`.
4. Replace `frame := &Frame{...}` with `frame := pool.Acquire()`.
5. Replace `// frame goes out of scope` with `defer pool.Release(frame)`.
6. Add a `runtime/trace` test verifying zero allocations.

The migration is documented in `docs/migration-from-make-per-frame.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_mempool_acquire_total`                   | counter    | Acquires, labelled `pool_name`, `result={ok, growth_required}`.              |
| `helix_mempool_release_total`                   | counter    | Releases.                                                                    |
| `helix_mempool_in_use`                          | gauge      | Current InUse per pool.                                                      |
| `helix_mempool_peak_in_use`                     | gauge      | All-time PeakInUse per pool.                                                 |
| `helix_mempool_gc_allocations_total`            | counter    | Pool growth events; alert: > 0 in steady state.                              |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C23 §3](../../04_Latency/09_Memory_and_Cache_Optimization.md) — origin                    | Origin chapter; full Pool / Arena / ByteBufferPool / Slab API.                       |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Per-frame Frame and RTP packet pools.                                                |
| `helix-allocator` (descriptor sibling)                                                    | helix-allocator imports helix-mempool to enforce hot-path zero-alloc.                |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-mempool-A        | sync.Pool-style victim cache vs strict capacity — which is the better default?                                | C23 §3 next revision                                |
| OQ-mempool-B        | NUMA-aware pools — needed for HelixPlay's multi-socket targets?                                                | `08_Operations/04_Observability_and_Events.md`     |
| OQ-mempool-C        | Pool size auto-tuning based on PeakInUse trends — automated or operator-driven?                                | C23 §3 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/09_Memory_and_Cache_Optimization.md`](../../04_Latency/09_Memory_and_Cache_Optimization.md) §3 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #15                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage; mandatory `testing.AllocsPerRun` assertions on hot paths.         |
| Integration    | Concurrent goroutines hammering Pool[Frame] at 60 fps; zero GC growth.                       |
| E2E            | helix-mempool → helix-pipeline frame buffer lifecycle.                                        |
| Security       | govulncheck + Snyk + Trivy + ABI fuzzer for Slab.                                            |
| Benchmarking   | Pool.Acquire ≤ 50 ns p999; Arena.Alloc ≤ 30 ns p999.                                          |
| Chaos          | GC-pause injection; pool invariants hold.                                                     |
| Stress         | 24-hour cycle; zero GC allocations.                                                           |
| Smoke          | 30-second 1K acquire/release; zero allocations.                                               |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `02_multi_session_single_host/03_mempool_no_alloc_in_hot_path` baseline-parity.               |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-mempool.md` — 2026-04-30.
