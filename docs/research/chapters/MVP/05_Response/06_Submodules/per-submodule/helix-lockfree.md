# `helix-lockfree` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-lockfree`                                                                                                       |
| **Origin chapter:section**  | [C17 §3](../../04_Latency/03_LockFree_Data_Structures.md) — *SPSC ringbuffer + MPSC queue + atomic primitives*         |
| **Public path (4 mirrors)** | `vasic-digital/helix-lockfree` on GitHub + GitLab + GitFlic + GitVerse                                                 |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `golang.org/x/sys/cpu` (cache-line size detection)                                                                      |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `lockfree-1.x` — builder `golang-builder`, runtime `distroless-static`                                              |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/02_multi_session_single_host/scenarios/02_spsc_ringbuffer_under_concurrent_session.scenario.yaml`        |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-lockfree` provides **lock-free Single-Producer/Single-Consumer ringbuffers, Multi-Producer/Single-Consumer queues, and atomic primitive helpers** for HelixPlay's hot-path concurrency. Origin: [C17 §3](../../04_Latency/03_LockFree_Data_Structures.md). The submodule replaces locked-channel patterns in the per-frame critical path where Go channel overhead (≈ 100 ns per send) is unacceptable; lock-free SPSC reaches ≈ 5 ns per write.

The submodule was introduced to consolidate atomic-pointer-based ringbuffer logic that earlier projects re-implemented inline. R-04 mandates one canonical landing; the submodule's API enforces correct cache-line alignment + memory-fence usage that hand-rolled implementations frequently misuse.

---

## 2. Public API Surface

### 2.1 The `SPSCRing` type

```go
package lockfree

// SPSCRing is a Single-Producer / Single-Consumer ringbuffer.
// Producer goroutine calls Write; consumer calls Read. Capacity is
// power-of-2; the producer/consumer indices are 64-bit atomics on
// separate cache lines to avoid false sharing.
type SPSCRing[T any] struct { /* ... */ }

func NewSPSCRing[T any](capacity int) *SPSCRing[T]
func (r *SPSCRing[T]) Write(item T) bool   // false if full
func (r *SPSCRing[T]) Read() (T, bool)     // false if empty
func (r *SPSCRing[T]) Size() int           // approximate
```

### 2.2 The `MPSCQueue` type

```go
package lockfree

// MPSCQueue is a Multi-Producer / Single-Consumer unbounded
// queue. Producers contend on a CAS-based tail; the consumer
// reads sequentially. Suitable for many-producer fan-in.
type MPSCQueue[T any] struct { /* ... */ }

func NewMPSCQueue[T any]() *MPSCQueue[T]
func (q *MPSCQueue[T]) Enqueue(item T)
func (q *MPSCQueue[T]) Dequeue() (T, bool)
func (q *MPSCQueue[T]) Size() int           // approximate
```

### 2.3 Atomic primitive helpers

```go
package lockfree

// AlignedAtomicInt64 ensures the int64 occupies its own cache line.
type AlignedAtomicInt64 struct {
    _ [56]byte
    Value int64
    _ [56]byte
}

// AlignedAtomicPointer provides cache-line-isolated pointer atomics.
type AlignedAtomicPointer[T any] struct { /* ... */ }
```

### 2.4 The `WaitGroup` (lock-free variant)

```go
package lockfree

// WaitGroup is a lock-free variant of sync.WaitGroup with
// counter-only atomic ops; suitable for hot-path use where
// sync.WaitGroup's mutex contention matters.
type WaitGroup struct { /* ... */ }

func (wg *WaitGroup) Add(delta int)
func (wg *WaitGroup) Done()
func (wg *WaitGroup) Wait()
```

### 2.5 Tagged atomic pointers (ABA-safe)

```go
package lockfree

// TaggedPtr embeds a 16-bit tag in the unused bits of a 64-bit
// pointer (x86-64 / arm64 enforce a 48-bit virtual address space,
// leaving 16 high bits free). The tag increments on every CAS to
// detect ABA scenarios.
type TaggedPtr[T any] struct { /* ... */ }

func NewTaggedPtr[T any](initial *T) *TaggedPtr[T]
func (tp *TaggedPtr[T]) Load() (*T, uint16)
func (tp *TaggedPtr[T]) CompareAndSwap(old *T, oldTag uint16, new *T) bool
func (tp *TaggedPtr[T]) Store(value *T)
```

### 2.6 The `Backoff` helper

```go
package lockfree

// Backoff implements an adaptive spin/yield/sleep backoff for
// CAS-retry loops in the MPSC queue and TaggedPtr. Starts with
// PAUSE-spin, escalates to runtime.Gosched, then to time.Sleep.
type Backoff struct { /* ... */ }

func NewBackoff() *Backoff
func (b *Backoff) Spin()         // pause spin (≤ 16 iterations)
func (b *Backoff) Yield()        // runtime.Gosched
func (b *Backoff) Sleep()        // exponential time.Sleep
func (b *Backoff) Reset()
```

### 2.7 Cache-line constants

```go
package lockfree

// CacheLineSize is the architecture-specific cache-line size.
// Set at init time from golang.org/x/sys/cpu.
var CacheLineSize int    // 64 on amd64, 64 on arm64

// PadSize is the byte count needed to pad a struct field to the
// next cache line boundary.
func PadSize(currentSize int) int
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (querying `/proc/cpuinfo` cache-line size at startup as a fallback if `golang.org/x/sys/cpu` doesn't expose it).

### 3.2 External (Go)

- `golang.org/x/sys/cpu` — for `cpu.CacheLinePadSize` constant. Pinned to ≥ v0.20.0.

### 3.3 External (system)

None. Pure-Go.

---

## 4. Container Build (S02 §3 lane: `lockfree-1.x`)

**Builder:** `golang-builder` (no cgo). **Runtime:** `distroless-static`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
SPSC + MPSC concurrent stress with race detector (`-race`); verify no data races on every operation.

### 5.2 Integration
Multi-goroutine producer/consumer pairs at 1 K writes/s; verify ordering + zero-loss.

### 5.3 E2E
helix-lockfree → helix-pipeline frame queue end-to-end roundtrip.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: ABA-problem fuzzer (verify the MPSC-queue's tail-CAS is ABA-safe via tagged pointers).

### 5.5 Benchmarking
SPSC write p999 ≤ 8 ns; MPSC enqueue p999 ≤ 25 ns under 4 producers.

### 5.6 Chaos
GC-pause injection during enqueue; verify queue invariants hold across GC.

### 5.7 Stress
24-hour 1 M ops/s SPSC + MPSC; verify no growth, no leak, no race.

### 5.8 Smoke
30-second post-deploy: SPSC write/read 1 K items; verify FIFO order.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`02_multi_session_single_host/02_spsc_ringbuffer_under_concurrent_session.scenario` — concurrent sessions using SPSC rings; verify zero data races + zero loss under concurrent session boundary events.

---

## 6. Challenges Entry-Point (S03 §4 row #09)

**Topology:** `02_multi_session_single_host`. **Scenario:** `02_spsc_ringbuffer_under_concurrent_session.scenario.yaml`. **Why this scenario.** Concurrent-session boundary events (session start/end while another session is mid-frame) stress the SPSC ringbuffer's lifecycle invariants. **Baseline:** zero `runtime.deadlock`, zero data-race reports under `-race`, OTLP trace topology stable across the boundary.

---

## 7. R-18 Inheritance

`helix-lockfree` imports `helix-r18-safeexec` for boundary subprocess invocations (cpu-cache-line-size discovery fallback). Hot path (atomic ops) is pure Go.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the C17 §3 API freeze + 2 consecutive green Ten-test cycles, including a dedicated `-race` E2E run that no concurrent submodule can introduce a race.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_LOCKFREE_RING_CAPACITY`     | `1024`        | int (power of 2)        | Default SPSC ring capacity.                                            |
| `HELIX_LOCKFREE_PADDING_BYTES`     | `64`          | int (typically 64 or 128)| Cache-line padding override; auto-detected by default.               |
| `HELIX_LOCKFREE_DEBUG_RACE`        | `false`       | bool                    | Enable additional sanity checks at writer/reader boundaries (dev only).|

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `SPSCRing.Write()`                    | 4 ns     | 6 ns    | 8 ns    | One atomic-store + cache-line-aligned data write.                |
| `SPSCRing.Read()`                     | 4 ns     | 6 ns    | 8 ns    | Symmetric.                                                        |
| `MPSCQueue.Enqueue()` (4 producers)   | 15 ns    | 22 ns   | 25 ns   | One CAS in the common case; retry on contention.                 |
| `MPSCQueue.Dequeue()`                 | 5 ns     | 8 ns    | 12 ns   | Single-consumer; one atomic-load.                                |
| `WaitGroup.Add()`                     | 3 ns     | 5 ns    | 8 ns    | Atomic-add on the counter.                                       |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `lockfree: ErrCapacityNotPowerOf2`            | Caller passed non-power-of-2 to NewSPSCRing                    | Round up to next power-of-2; the submodule rejects at construction time.   |
| Race report under `-race`                     | Producer / consumer roles violated (e.g. two producers on SPSC)| Audit the goroutines using the ring; SPSC must have exactly one of each.  |
| Memory growth on MPSC                         | Consumer not draining as fast as producers enqueue            | Add backpressure on the producer side; investigate consumer slow-drain.   |
| Stalls on `WaitGroup.Wait()`                  | Add was not paired with Done                                   | Standard sync.WaitGroup audit pattern applies; emit a metric for `Add - Done` skew. |

### 9.4 Migration from `chan T`

A consumer migrating from a buffered `chan T` to `helix-lockfree.SPSCRing[T]`:

1. Identify the channel; verify it's truly SPSC (one goroutine writes, one reads).
2. Replace `ch := make(chan T, N)` with `ring := lockfree.NewSPSCRing[T](N)`.
3. Replace `ch <- item` with `if !ring.Write(item) { /* full handling */ }`.
4. Replace `item := <-ch` with `item, ok := ring.Read(); if !ok { /* empty handling */ }`.
5. Note: the unbuffered semantics of `chan T` (blocking sync) is NOT supported; consumers that depend on that semantic must keep the channel.
6. Add OTLP span around the producer/consumer loop; metric for `Size()` to detect backpressure.

The migration is documented in `docs/migration-from-chan.md`. Be aware: the `-race` discipline is stricter for lock-free code than for channel code.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_lockfree_spsc_writes_total`              | counter    | Total SPSC writes, labelled `result={ok, full}`.                             |
| `helix_lockfree_spsc_reads_total`               | counter    | Total SPSC reads, labelled `result={ok, empty}`.                             |
| `helix_lockfree_mpsc_enqueues_total`            | counter    | Total MPSC enqueues, labelled `cas_retries`.                                 |
| `helix_lockfree_ring_size`                      | gauge      | Current ring occupancy (per ring instance).                                  |
| `helix_lockfree_waitgroup_skew`                 | gauge      | Add - Done counter; non-zero indicates pending work.                          |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C17 §3](../../04_Latency/03_LockFree_Data_Structures.md) — origin                         | Origin chapter; full SPSC / MPSC / atomic API.                                       |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Inter-stage SPSC rings replace channel hops in the encode pipeline.                  |
| [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md) — Bench                  | Bench harness uses MPSC for sample collection from many goroutines.                 |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-lockfree-A       | MPMC variant — should it ship in the submodule or remain unsupported?                                         | C17 §3 next revision                                |
| OQ-lockfree-B       | Generics overhead — Go's monomorphisation cost on hot-path types; benchmark periodically.                      | C17 §3 next revision                                |
| OQ-lockfree-C       | NUMA-aware variant for multi-socket hosts — required for HelixPlay's MVP target deployments?                   | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/03_LockFree_Data_Structures.md`](../../04_Latency/03_LockFree_Data_Structures.md) §3 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #09                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage; mandatory `-race` flag for every test.                            |
| Integration    | Real concurrent goroutines exercising SPSC + MPSC at 1 K ops/s.                              |
| E2E            | helix-lockfree → helix-pipeline frame queue roundtrip.                                       |
| Security       | govulncheck + Snyk + Trivy + ABA-problem fuzzer.                                             |
| Benchmarking   | SPSC ≤ 8 ns p999; MPSC ≤ 25 ns p999 (4 producers).                                           |
| Chaos          | GC-pause injection; queue invariants hold.                                                    |
| Stress         | 24-hour 1 M ops/s; zero leak; zero race.                                                      |
| Smoke          | 30-second SPSC FIFO-order verification.                                                       |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `02_multi_session_single_host/02_spsc_ringbuffer_under_concurrent_session` baseline-parity.   |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-lockfree.md` — 2026-04-30.
