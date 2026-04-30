# `helix-shm` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-shm`                                                                                                            |
| **Origin chapter:section**  | [C15 §6](../../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md) — *Shared-memory NV12 / I420 frame pages*             |
| **Public path (4 mirrors)** | `vasic-digital/helix-shm` on GitHub + GitLab + GitFlic + GitVerse                                                       |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `golang.org/x/sys/unix` (mmap, mlock, memfd_create), `golang.org/x/sync/singleflight`                                  |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `shared-memory-1.x` — builder `golang-builder`, runtime `distroless-base` (needs shm support in /dev/shm)           |
| **Test matrix (S01 §5)**    | Ten / inline + Challenges row **delegated to `helix-pipeline`** (S01 §5.2 exception)                                    |
| **Challenges entry (S03 §4)** | Delegated; see [S03 §4.2](../03_Challenges_Submodule.md#42-the-helix-shm-challenges-delegation-s01-§52)               |
| **HelixQA cadence (S04 §5)**| Per-PR (own primary scenario in helix-pipeline) + nightly + canary + pre-release                                       |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-shm` is the **shared-memory NV12 / I420 frame page allocator** for HelixPlay's zero-copy capture-to-encode pipeline. The submodule wraps `mmap` + `memfd_create` + `mlock` to allocate page-aligned frame buffers that the host OS capture stack ([C28 §6](../../05_Video_Audio/03_Capture_Pipelines.md)) writes into and the encoder ([C27 §6](../../05_Video_Audio/02_Hardware_Encoders.md)) reads from without an intermediate copy.

The submodule was introduced in [C15 §6](../../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md) as the consolidation of NV12 / I420 buffer-management logic that earlier projects (HelixAgent, Catalogizer) re-implemented per-pipeline. R-04 mandates reuse over duplication; the per-pipeline implementations had drifted in subtle ways (alignment assumptions, lock-page lifecycle) that this submodule's canonical implementation eliminates.

`★ Why a library, not a runtime.` `helix-shm` is intentionally a library — it allocates pages, maps them, manages their lifecycle, and provides a typed Go-side handle. It does not run as a separate process. The Challenges row delegation to `helix-pipeline` (S01 §5.2) reflects this: `helix-shm` cannot stand up a full system on its own because it has no behaviour observable to a user — observable behaviour appears only when the pages are written and read by other components.

---

## 2. Public API Surface

### 2.1 The `Page` type

```go
package shm

// Page is a page-aligned shared-memory region suitable for NV12
// or I420 frame data. Backed by an anonymous file (memfd_create
// where available, /dev/shm fallback) plus mmap with MAP_POPULATE.
// Optionally mlock'd to prevent eviction under memory pressure.
type Page struct {
    Data []byte    // mapped memory; len() is exactly Size().
    Fd   int       // file descriptor for the backing memfd; -1 after Close.
    Size int       // page-aligned size, ≥ requested size.
}

func New(size int, opts ...Option) (*Page, error)
func (p *Page) Close() error
func (p *Page) Mlock() error
func (p *Page) Munlock() error
```

### 2.2 The `Pool` type

```go
package shm

// Pool is a reusable pool of Pages of a fixed size. Allocates
// lazily; reuses pages on Release. Bounded; blocks acquirers when
// the pool is exhausted (with context cancellation support).
type Pool struct { /* ... */ }

func NewPool(pageSize int, capacity int, opts ...PoolOption) *Pool
func (p *Pool) Acquire(ctx context.Context) (*Page, error)
func (p *Pool) Release(page *Page)
func (p *Pool) Close() error
func (p *Pool) Stats() PoolStats
```

### 2.3 The `Frame` typed-handle helper

```go
package shm

// Frame is a Page interpreted as an NV12 or I420 frame. Provides
// strided accessors for the Y, U, V planes without copying.
type Frame struct {
    Page    *Page
    Format  PixelFormat   // FormatNV12, FormatI420
    Width   int
    Height  int
    Stride  int
}

func NewFrame(page *Page, format PixelFormat, width, height int) *Frame
func (f *Frame) PlaneY() []byte
func (f *Frame) PlaneU() []byte
func (f *Frame) PlaneV() []byte
func (f *Frame) Stride() int
```

### 2.4 The `Region` type for sub-frame access

```go
package shm

// Region describes a sub-rectangle of a Frame. Used by the encoder
// when only a slice of the frame is being processed (e.g. parallel
// region-of-interest encoding per C29 §6 dual-path).
type Region struct {
    Frame  *Frame
    Rect   Rect           // x, y, width, height in pixels
}

func (r *Region) PlaneY() []byte
func (r *Region) PlaneU() []byte
func (r *Region) PlaneV() []byte
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (querying `/proc/sys/kernel/shmmax`, `/proc/sys/kernel/shmall` at startup for capacity sizing).

### 3.2 External (Go)

- `golang.org/x/sys/unix` — `unix.Mmap`, `unix.Mlock`, `unix.MemfdCreate`, `unix.Madvise`. Pinned to ≥ v0.20.0 for the MemfdCreate signature stability.
- `golang.org/x/sync/singleflight` — deduplicates concurrent same-size pool growth requests.

### 3.3 External (system)

- A Linux kernel ≥ 5.4 (for `MAP_POPULATE` + `memfd_create`).
- For non-Linux platforms (macOS Metal capture variant, Windows DXGI variant), the submodule falls back to anonymous mmap without memfd_create; the fallback is documented in `docs/non-linux-fallback.md` of the submodule.

---

## 4. Container Build (S02 §3 lane: `shared-memory-1.x`)

**Builder:** `golang-builder`. **Runtime:** `distroless-base` (needs `/dev/shm` support; static doesn't). **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + `--shm-size=512m` + `--ipc=private` (NOT `--ipc=host` per Constitution §11.5.3 deny-list).

---

## 5. Test Matrix (S01 §5: Ten / inline, Challenges delegated)

### 5.1 Unit
- Page allocation across sizes (4 KiB to 16 MiB).
- Pool concurrent acquire/release under context cancellation.
- Frame plane accessor offsets for NV12 + I420 at 1080p, 4K, 8K.

### 5.2 Integration
- Real `/dev/shm` allocation; verify `mlock` succeeds and `munlock` releases.
- Cross-process page sharing via fd passing over Unix sockets.

### 5.3 E2E
- `helix-shm` page → `helix-capture` write → `helix-encoder` read round-trip with byte-exact verification.

### 5.4 Security
- govulncheck + Snyk + Trivy; verify `mmap` flags are not user-controllable.
- Custom: attempt a `MAP_FIXED` invocation (must be rejected by the API).

### 5.5 Benchmarking
- Page allocation p999 ≤ 50 µs; mlock p999 ≤ 200 µs.
- Pool acquire (warm) p999 ≤ 5 µs.

### 5.6 Chaos
- Memory-pressure injection (`echo 1 > /proc/sys/vm/drop_caches`); verify mlock'd pages survive.

### 5.7 Stress
- 24-hour run with 60 fps × 8 K (4× 1080p) page allocate/release cycle; verify no fragmentation, no leak.

### 5.8 Smoke
- 30-second post-deploy: allocate one 4K NV12 frame, verify size + alignment.

### 5.9 Full Automation
- §5.1–§5.8 in CI matrix.

### 5.10 Challenges
- **Delegated to `helix-pipeline`** per S01 §5.2. The delegated test exercises `helix-shm` end-to-end inside `01_minimum_viable_session/06_full_pipeline_end_to_end.scenario`. Per-submodule attribution is recorded at `vasic-digital/Challenges/baselines/01_minimum_viable_session/06_full_pipeline_end_to_end/per-submodule/helix-shm.metrics.json`.

---

## 6. Challenges Entry-Point (S03 §4 row #06, delegated)

**Topology:** `01_minimum_viable_session` (via `helix-pipeline` delegation). **Scenario:** `06_full_pipeline_end_to_end`. **Why delegated.** A library with no observable behaviour cannot pass a Challenges row in isolation; the delegation contract is documented in `helix-shm/tests/challenges/delegated.md` per S03 §4.2.

---

## 7. R-18 Inheritance

`helix-shm` imports `helix-r18-safeexec` for boundary subprocess invocations only. The hot path (page allocation, mmap, mlock) does not invoke subprocesses; `mmap` itself is a syscall, not a subprocess.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation gated on the C15 §6 API freeze + 2 consecutive green Ten-test cycles via the helix-pipeline delegation.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default          | Range / type                | Purpose                                                                |
|------------------------------------|------------------|-----------------------------|------------------------------------------------------------------------|
| `HELIX_SHM_BACKING`                | `memfd`          | `memfd` / `tmpfs`           | Backing-store choice; memfd preferred on Linux ≥ 5.4.                  |
| `HELIX_SHM_MLOCK_ENABLED`          | `true`           | bool                        | Mlock pages by default; disable only for memory-constrained operators. |
| `HELIX_SHM_POOL_CAPACITY`          | `64`             | int [4, 1024]               | Default pool capacity per pixel-format/resolution.                     |
| `HELIX_SHM_PAGE_SIZE_HINT`         | `2097152` (2 MiB)| int                         | Default page size; adjusts to NV12-4K (≈ 12 MiB) at runtime.           |
| `HELIX_SHM_MADVISE_HUGE`           | `true`           | bool                        | Use `MADV_HUGEPAGE` to opt into transparent huge pages.                |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `New(size)` allocation                | 20 µs    | 40 µs   | 50 µs   | memfd_create + mmap + (optional) mlock.                           |
| `Pool.Acquire()` (warm)               | 200 ns   | 1 µs    | 5 µs    | Lock-free SPSC ring; one CAS in the common case.                 |
| `Pool.Acquire()` (cold)               | 30 µs    | 60 µs   | 100 µs  | Triggers a `New()`.                                               |
| `Mlock()`                             | 80 µs    | 150 µs  | 200 µs  | Linux kernel page-locking syscall.                               |
| Frame plane accessor (no copy)        | 5 ns     | 15 ns   | 30 ns   | Pure pointer arithmetic.                                          |
| Memory per pool (NV12 1080p, cap=64)  | 192 MiB  | 192 MiB | 192 MiB | 64 × 3 MiB; constant.                                             |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                  |
|-----------------------------------------------|----------------------------------------------------------------|------------------------------------------------------------------------------|
| `shm: ErrPoolExhausted`                       | All capacity-N pages in use; new acquirer blocked              | Increase `HELIX_SHM_POOL_CAPACITY`; investigate slow consumer.               |
| `shm: ErrMlockFailed`                         | RLIMIT_MEMLOCK too low                                         | Raise via `prlimit --memlock=unlimited` (operator-side, container privilege). |
| `shm: ErrMmapFailed: ENOMEM`                  | Host out of virtual address space                              | Reduce pool capacity; horizontal-scale.                                      |
| `shm: ErrUnsupportedKernel`                   | Linux < 5.4 detected                                           | Upgrade host kernel; falls back to `tmpfs` backing automatically with warning. |
| `shm: ErrInvalidPlaneOffset`                  | Frame stride does not match width × bytes-per-pixel            | Caller bug; report Frame parameters and trace to allocation site.            |

### 9.4 Migration from per-pipeline NV12 buffers

A consumer migrating from inline `make([]byte, ...)` NV12 buffers:

1. Replace `make([]byte, width*height*3/2)` with `shm.NewPool(...)` allocation at startup.
2. Replace `buf := make(...)` per-frame with `page, _ := pool.Acquire(ctx); frame := shm.NewFrame(page, ...)`.
3. Pass the `*Frame` (not the `[]byte`) through the pipeline; downstream consumers use the typed accessors.
4. Replace `runtime.GC()`-friendly buffers with `pool.Release(page)` on the consumer side.
5. Add the `pool.Stats()` metric to the OTLP trace per Constitution §10.

The migration is documented in `docs/migration-from-inline-nv12.md`.

### 9.5 Observability metrics catalog

Per Constitution §10:

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_shm_page_alloc_total`                    | counter    | Page allocations, labelled `result={ok, pool_exhausted, mmap_failed}`.       |
| `helix_shm_page_alloc_latency_seconds`          | histogram  | Allocation latency; buckets at 1 µs, 10 µs, 100 µs, 1 ms.                    |
| `helix_shm_pool_size`                           | gauge      | Current pool occupancy.                                                      |
| `helix_shm_pool_capacity`                       | gauge      | Pool configured capacity.                                                    |
| `helix_shm_mlock_failures_total`                | counter    | Mlock failures, labelled `errno`.                                            |
| `helix_shm_huge_pages_used`                     | gauge      | Pages backed by transparent huge pages (when MADV_HUGEPAGE granted).         |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C15 §6](../../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md) — origin                  | Origin chapter; full Page / Pool / Frame / Region API surface.                       |
| [C28 §6](../../05_Video_Audio/03_Capture_Pipelines.md) — Capture                          | Captured frames written into `helix-shm` Pages.                                      |
| [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md) — Bench                  | Benchmark harness allocates Pages for throughput tests.                             |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates Acquire/Release lifecycle across capture/encode stages.        |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport reads Pages for zero-copy send via io_uring + XDP.                        |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-shm-A            | macOS Metal capture variant — should it use IOSurface instead of mmap?                                         | C15 §6 + C28 §6 next revision                       |
| OQ-shm-B            | Windows DXGI variant — `CreateFileMapping` vs `VirtualAlloc` for the equivalent?                               | C15 §6 + C28 §6 next revision                       |
| OQ-shm-C            | Memory-pressure handling — pool eviction policy when host approaches OOM?                                      | `08_Operations/04_Observability_and_Events.md`     |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md`](../../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §5.2 §7 | 1,218 | 2026-04-30 | catalog row #06, delegation exception      |
| [`../03_Challenges_Submodule.md`](../03_Challenges_Submodule.md) §4.2 |   517 | 2026-04-30 | delegation contract                       |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target (this submodule)                                                              |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (allocation + pool + frame plane accessors).                       |
| Integration    | Real /dev/shm + memfd_create + mlock all covered with at least one test.                     |
| E2E            | helix-shm → helix-capture → helix-encoder roundtrip with byte-exact verification.            |
| Security       | govulncheck + Snyk + Trivy + custom MAP_FIXED rejection test.                                |
| Benchmarking   | p999 ≤ §9.2 budget; pool acquire (warm) ≤ 5 µs.                                              |
| Chaos          | Memory-pressure injection; mlock'd pages survive verification.                              |
| Stress         | 24-hour 60 fps × 4K cycle; zero fragmentation; zero leak.                                   |
| Smoke          | 30-second post-deploy: 4K NV12 page allocation + size verification.                          |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | Delegated to helix-pipeline; baseline-parity at the per-submodule slice.                      |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-shm.md` — 2026-04-30.
