# `helix-iouring` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-iouring`                                                                                                        |
| **Origin chapter:section**  | [C16 §4](../../04_Latency/02_io_uring_and_Kernel_Bypass.md) — *io_uring async I/O wrapper*                            |
| **Public path (4 mirrors)** | `vasic-digital/helix-iouring` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `github.com/iceber/iouring-go`, `golang.org/x/sys/unix`                                                                 |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `io-uring-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                              |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/07_kernel_bypass_send_path/scenarios/01_iouring_completion_under_120hz_send.scenario.yaml`              |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision (upstream OSS exists, none under vasic-digital).                                       |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-iouring` is the **io_uring async-I/O wrapper** for HelixPlay's kernel-bypass send path. The submodule wraps `iouring-go` to expose a Go-idiomatic, allocation-free API for batched read/write/sendmsg/recvmsg/splice operations over file descriptors and sockets. The hot-path use case: `helix-transport` ([C37 §9](../../05_Video_Audio/12_Network_Transport.md)) submits batched UDP sends via SQE rings without per-send syscalls.

The submodule was introduced in [C16 §4](../../04_Latency/02_io_uring_and_Kernel_Bypass.md) as the consolidation of io_uring-Go bindings. Direct use of `iouring-go` works but bleeds implementation-detail across the codebase; the submodule's API normalises ring-creation, capability-detection, and batched completion-reaping into a stable shape.

---

## 2. Public API Surface

### 2.1 The `Ring` type

```go
package iouring

// Ring is a configured io_uring submission/completion ring pair.
type Ring struct { /* ... */ }

func New(entries int, opts ...Option) (*Ring, error)
func (r *Ring) Close() error
func (r *Ring) Submit() (n int, err error)
func (r *Ring) WaitCompletions(ctx context.Context, max int) ([]Completion, error)
```

### 2.2 Submission helpers

```go
package iouring

func (r *Ring) PrepRead(fd int, buf []byte, offset uint64) (*Request, error)
func (r *Ring) PrepWrite(fd int, buf []byte, offset uint64) (*Request, error)
func (r *Ring) PrepSendmsg(fd int, msg *unix.Msghdr, flags int) (*Request, error)
func (r *Ring) PrepRecvmsg(fd int, msg *unix.Msghdr, flags int) (*Request, error)
func (r *Ring) PrepSplice(fdIn, fdOut int, count int, flags int) (*Request, error)
```

### 2.3 Capability detection

```go
package iouring

func KernelCaps() (Capabilities, error)
type Capabilities struct {
    OpSendmsg     bool   // SQE_CMD_OP_SENDMSG (5.1+)
    OpSendZerocopy bool  // SQE_CMD_OP_SEND_ZC (6.0+)
    OpSplice       bool
    OpFutex        bool
}
```

The capability detection lets the submodule fall back to `unix.Sendmsg` when running on kernels lacking specific opcodes; consumers query `KernelCaps()` once at startup.

### 2.4 Configuration `Option` types

```go
package iouring

type Option func(*config)

// WithSQPoll enables the kernel SQ-polling thread; reduces syscalls
// at the cost of a busy CPU.
func WithSQPoll(idleMs int) Option

// WithRegisteredFiles pre-registers a set of fds; bypasses fd lookup
// on each SQE.
func WithRegisteredFiles(fds []int) Option

// WithIOPoll enables polled-mode I/O; useful for NVMe direct I/O.
func WithIOPoll() Option

// WithSingleIssuer optimises for a single submitter goroutine; the
// kernel skips internal locking on the submission path.
func WithSingleIssuer() Option
```

### 2.5 The `Completion` struct

```go
package iouring

type Completion struct {
    UserData uint64    // opaque; consumers correlate by setting this on Prep*.
    Result   int32     // syscall return value (negative for errno).
    Flags    uint32    // see CQE_F_* constants below.
}

const (
    CQE_F_BUFFER     = 1 << 0  // result contains the buffer ID for buffer-pool reads.
    CQE_F_MORE       = 1 << 1  // more completions for this submission to follow.
    CQE_F_SOCK_NONEMPTY = 1 << 2
    CQE_F_NOTIF      = 1 << 3  // zero-copy notification.
)
```

### 2.6 The `Request` chain helper

```go
package iouring

// Request is a queued SQE. The caller can chain: SetLink for IOSQE_IO_LINK
// (sequential), or set the IOSQE_IO_HARDLINK flag for unconditional chaining.
type Request struct { /* ... */ }

func (r *Request) SetLink() *Request          // mark next as IO_LINK
func (r *Request) SetHardLink() *Request      // mark next as IO_HARDLINK
func (r *Request) SetUserData(u uint64) *Request
func (r *Request) Cancel() error              // SQE-side cancel
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (querying `/proc/sys/kernel/io_uring_disabled`).

### 3.2 External (Go)

- `github.com/iceber/iouring-go` — primary cgo-free io_uring binding for Go. Pinned to ≥ v0.0.0-20240310 for IORING_OP_SEND_ZC support.
- `golang.org/x/sys/unix` — Linux syscall surface.

### 3.3 External (system)

- Linux kernel ≥ 5.6 (basic io_uring) or ≥ 6.0 (zero-copy send). Capability detection (§2.3) gracefully degrades on older kernels.

---

## 4. Container Build (S02 §3 lane: `io-uring-1.x`)

**Builder:** `golang-builder-cgo` (cgo enabled for select transitive ops). **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + seccomp profile must allow `io_uring_setup`, `io_uring_register`, `io_uring_enter` syscalls (specifically allow-listed in `policies/seccomp-allowlist.json`).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Ring lifecycle + SQE/CQE round-trips with synthetic file descriptors (pipe pair).

### 5.2 Integration
Real ring against real /tmp file + UDP socket; verify batched submit and completion reap.

### 5.3 E2E
helix-iouring → helix-transport → real UDP send loopback round-trip.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: attempt to register an out-of-process fd (must be rejected by the kernel; submodule surfaces error cleanly).

### 5.5 Benchmarking
SQE submit p999 ≤ 200 ns (cgo-free path); CQE reap throughput ≥ 1 M completions/s on amd64.

### 5.6 Chaos
Inject completion-ring overflow; verify backpressure surfacing.

### 5.7 Stress
24-hour 60 Hz × 1024 SQE batches; zero fd leak, zero memory growth.

### 5.8 Smoke
30-second post-deploy: write 1 KiB to /tmp via io_uring; verify file content.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`07_kernel_bypass_send_path/01_iouring_completion_under_120hz_send.scenario` — full pipeline at 120 Hz with io_uring-batched send path.

---

## 6. Challenges Entry-Point (S03 §4 row #07)

**Topology:** `07_kernel_bypass_send_path`. **Scenario:** `01_iouring_completion_under_120hz_send.scenario.yaml`. **Why this scenario.** io_uring's completion-reaping under sustained 120-Hz send rate is the production-like workload that exercises the submodule's batched-submit + reap paths simultaneously. **Baseline:** completion-reap latency p999 ≤ 50 µs; zero dropped completions over the scenario duration; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-iouring` imports `helix-r18-safeexec` for boundary subprocess invocations (kernel-feature detection). The hot path (SQE submit, CQE reap) is pure syscall, no subprocess.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on Linux kernel availability across operator deployments — the `v1.0.0` API freeze can land independently, but the SEND_ZC opcode use-cases require kernel 6.0+, which not every operator has. The capability-detection fallback (§2.3) keeps the submodule usable across the kernel matrix.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_IOURING_RING_ENTRIES`       | `1024`        | int [16, 32768]         | SQE/CQE ring size (must be power of 2).                                |
| `HELIX_IOURING_SQ_POLL`            | `false`       | bool                    | Enable kernel SQ-polling thread (reduces syscalls; costs a CPU).      |
| `HELIX_IOURING_SQ_POLL_IDLE_MS`    | `100`         | int [1, 10000]          | SQ-poll idle timeout before kernel thread sleeps.                     |
| `HELIX_IOURING_USE_ZEROCOPY`       | `auto`        | `auto` / `true` / `false`| Use IORING_OP_SEND_ZC when kernel ≥ 6.0; `auto` detects.             |
| `HELIX_IOURING_REGISTER_FILES`     | `true`        | bool                    | Use IORING_REGISTER_FILES for hot fds (avoids per-op fd lookup).      |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| SQE submit (cgo-free, batched)        | 80 ns    | 150 ns  | 200 ns  | One memory-mapped ring write + zero syscalls per SQE.            |
| `Submit()` (kernel transition)        | 600 ns   | 1.5 µs  | 3 µs    | Single io_uring_enter syscall regardless of batch size.          |
| `WaitCompletions()` reap rate         | 1 M/s    | —       | —       | Steady state on amd64; SQ-poll mode reaches 2 M/s.               |
| Memory per Ring (entries=1024)        | 256 KiB  | —       | —       | SQ array + CQ array + SQE pool, page-aligned.                    |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                      | Remediation                                                                |
|-----------------------------------------------|------------------------------------------------------------|----------------------------------------------------------------------------|
| `iouring: ErrUnsupportedKernel`               | Kernel < 5.6 detected                                      | Upgrade kernel; submodule falls back to `unix.Read`/`Write` paths.        |
| `iouring: ErrRingFull`                        | SQE submission rate exceeded ring capacity                  | Increase `HELIX_IOURING_RING_ENTRIES`; investigate consumer slow-drain.   |
| `iouring: ErrCQEOverflow`                     | Completion-ring overflow (kernel dropped CQEs)              | Reap completions more frequently; increase ring entries.                  |
| `iouring: ErrPermissionDenied`                | `io_uring_disabled` sysctl set or seccomp denial            | Verify container seccomp policy; check sysctl on host.                    |
| `iouring: ErrSendZCNotSupported`              | Kernel < 6.0; SEND_ZC opcode unavailable                    | Set `HELIX_IOURING_USE_ZEROCOPY=false` or upgrade kernel.                |

### 9.4 Migration from `unix.Read`/`Write`

A consumer migrating from blocking `unix.Read`/`Write`:

1. At startup: `ring, err := iouring.New(1024)`; query capabilities via `iouring.KernelCaps()`.
2. Replace each `unix.Read(fd, buf)` with `ring.PrepRead(fd, buf, 0)`; queue in batch.
3. Replace each `unix.Write(fd, buf)` with `ring.PrepWrite(fd, buf, 0)`.
4. After preparing N requests, `ring.Submit()`.
5. Reap completions: `comps, _ := ring.WaitCompletions(ctx, N)`; correlate by user_data.
6. Add OTLP span around `Submit()` + `WaitCompletions()` calls.

The migration is documented in `docs/migration-from-blocking-io.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_iouring_sqe_submitted_total`             | counter    | SQEs submitted, labelled `op` (read, write, sendmsg, recvmsg, splice).      |
| `helix_iouring_cqe_reaped_total`                | counter    | CQEs reaped, labelled `op`, `result={ok, err}`.                              |
| `helix_iouring_submit_latency_seconds`          | histogram  | Submit (kernel transition) latency.                                          |
| `helix_iouring_ring_full_total`                 | counter    | Times the SQE ring rejected a submission (full).                            |
| `helix_iouring_cqe_overflow_total`              | counter    | Kernel-side CQE overflow events.                                             |
| `helix_iouring_zerocopy_used_total`             | counter    | SEND_ZC opcode invocations (only when kernel ≥ 6.0).                        |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C16 §4](../../04_Latency/02_io_uring_and_Kernel_Bypass.md) — origin                       | Origin chapter; full Ring + Request + Capabilities API.                              |
| [C24 §6](../../04_Latency/10_Latency_Testing_and_Validation.md) — Bench                  | Benchmark harness uses io_uring for high-throughput workload generation.            |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport's UDP send path batches via io_uring.                                      |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-iouring-A        | SQ-poll mode — default off (CPU cost) or default on (latency benefit)?                                         | C16 §4 next revision                                |
| OQ-iouring-B        | iouring-go vs go-uring vs custom binding — re-evaluate annually?                                               | C16 §4 next revision                                |
| OQ-iouring-C        | Per-CPU rings vs single shared ring — trade-off for HelixPlay's typical 4-CPU host?                            | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/02_io_uring_and_Kernel_Bypass.md`](../../04_Latency/02_io_uring_and_Kernel_Bypass.md) §4 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #07                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Ring lifecycle + every Prep* helper).                              |
| Integration    | Real /tmp file + real UDP socket exercises submit / reap.                                     |
| E2E            | helix-iouring → helix-transport UDP loopback round-trip.                                      |
| Security       | govulncheck + Snyk + Trivy; cross-process fd registration rejection test.                    |
| Benchmarking   | SQE submit p999 ≤ 200 ns; reap throughput ≥ 1 M/s amd64.                                     |
| Chaos          | Completion-ring overflow injection; backpressure surfaces cleanly.                            |
| Stress         | 24-hour 60 Hz × 1024 SQE batches; zero leak.                                                  |
| Smoke          | 30-second post-deploy: 1 KiB write via io_uring + content verification.                      |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `07_kernel_bypass_send_path/01_iouring_completion_under_120hz_send` baseline-parity.          |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-iouring.md` — 2026-04-30.
