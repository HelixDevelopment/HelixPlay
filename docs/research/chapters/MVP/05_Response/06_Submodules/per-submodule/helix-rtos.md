# `helix-rtos` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-rtos`                                                                                                           |
| **Origin chapter:section**  | [C20 §3](../../04_Latency/06_RealTime_OS_and_Scheduling.md) — *SCHED_FIFO + cgroup pinning + isolated CPUs*          |
| **Public path (4 mirrors)** | `vasic-digital/helix-rtos` on GitHub + GitLab + GitFlic + GitVerse                                                     |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `golang.org/x/sys/unix`, `github.com/containerd/cgroups/v3`                                                              |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `sched-fifo-cgroup-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                     |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/08_controller_input_low_latency/scenarios/02_sched_fifo_under_high_load.scenario.yaml`                  |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-rtos` provides **SCHED_FIFO real-time scheduling, CPU pinning via cgroups v2, IRQ affinity helpers, and isolated-CPU detection** for HelixPlay's hot-path goroutines. Origin: [C20 §3](../../04_Latency/06_RealTime_OS_and_Scheduling.md). The submodule operationalises the priority-inversion-resistant scheduling policy that the gameplay round-trip depends on.

The submodule was introduced because Linux's default CFS scheduler can introduce 100s of µs of jitter under load, and that jitter is visible at the player's controller round-trip. SCHED_FIFO + isolated CPUs reduces the jitter to single-digit µs; the submodule encapsulates the per-thread `sched_setscheduler` + cgroup configuration + IRQ-affinity work.

---

## 2. Public API Surface

### 2.1 The `RTThread` type

```go
package rtos

// RTThread is a Go runtime OS thread pinned with SCHED_FIFO.
type RTThread struct { /* ... */ }

func PromoteCurrent(priority int) (*RTThread, error)
func (rt *RTThread) Demote() error
func (rt *RTThread) Priority() int
```

Promoting a Go goroutine to SCHED_FIFO requires `runtime.LockOSThread()` first; the submodule's `PromoteCurrent` does this internally and panics if the caller has already invoked `LockOSThread`.

### 2.2 The `CPUSet` type

```go
package rtos

// CPUSet binds a thread to a specific set of CPUs.
type CPUSet struct { /* ... */ }

func NewCPUSet(cpus ...int) *CPUSet
func (cs *CPUSet) Apply(thread *RTThread) error
func (cs *CPUSet) Cpus() []int
```

### 2.3 cgroup-v2 helpers

```go
package rtos

// CGroup represents a cgroup-v2 subgroup. The submodule's helpers
// configure cpu.weight, cpuset.cpus, memory.max, etc., to give
// HelixPlay's hot-path the resources it needs deterministically.
type CGroup struct { /* ... */ }

func NewCGroup(path string) (*CGroup, error)
func (cg *CGroup) SetCPUSet(cpus []int) error
func (cg *CGroup) SetCPUWeight(weight uint64) error
func (cg *CGroup) SetMemoryMax(bytes uint64) error
func (cg *CGroup) AttachThread(tid int) error
func (cg *CGroup) Close() error
```

### 2.4 IRQ-affinity helpers

```go
package rtos

// SetIRQAffinity pins an IRQ to a specific set of CPUs.
// Used to ensure NIC interrupts hit the same NUMA node + sibling
// CPU as the userspace consumer.
func SetIRQAffinity(irq int, cpus []int) error

// FindIRQsByDevice returns IRQ numbers associated with a device.
func FindIRQsByDevice(devName string) ([]int, error)
```

### 2.5 Capability detection

```go
package rtos

func KernelCaps() (Capabilities, error)
type Capabilities struct {
    SchedFIFOAvailable   bool   // CAP_SYS_NICE present
    CGroupV2Mounted      bool
    IsolatedCPUsPresent  bool   // /sys/devices/system/cpu/isolated nonempty
    PreemptRTKernel      bool   // PREEMPT_RT-patched kernel
}
```

### 2.6 Configuration options

```go
package rtos

type Option func(*config)

func WithPriority(priority int) Option
func WithIsolatedCPUsOnly() Option        // refuse to start if isolcpus= empty
func WithPreemptRTRequired() Option         // refuse if non-RT kernel
```

### 2.7 Scheduler statistics

```go
package rtos

type SchedulerStats struct {
    CPUsPinned       []int
    Priority         int
    InvoluntaryCSW   uint64    // involuntary context switches
    VoluntaryCSW     uint64    // voluntary context switches
    Migrations       uint64    // CPU migrations (should be 0 for pinned threads)
}

func (rt *RTThread) Stats() SchedulerStats
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (reading `/sys/devices/system/cpu/isolated`, `/proc/cmdline` at startup).

### 3.2 External (Go)

- `golang.org/x/sys/unix` — Linux syscalls (`SchedSetscheduler`, `SchedSetaffinity`).
- `github.com/containerd/cgroups/v3` — cgroup-v2 management. Pinned to ≥ v3.0.4.

### 3.3 External (system)

- Linux kernel ≥ 5.10 (cgroup-v2 unified hierarchy in production deployments).
- `CAP_SYS_NICE` capability for `sched_setscheduler` calls (Constitution §11.5.3 carve-out, similar shape to `helix-xdp`'s CAP_BPF — narrow capability, not `--privileged`).
- (Optional) PREEMPT_RT kernel for the strictest jitter requirements; the submodule works on stock kernels but documents the jitter delta.

---

## 4. Container Build (S02 §3 lane: `sched-fifo-cgroup-1.x`)

**Builder:** `golang-builder-cgo` (cgo for cgroups-v3 native helpers). **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + `--cap-add=SYS_NICE` (carve-out documented inline at C20 §3 + S02 §8.4 referenced precedent).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
RTThread promotion + demotion; CPUSet apply; cgroup configuration round-trips with mock fs.

### 5.2 Integration
Real container with `--cap-add=SYS_NICE`; promote a goroutine; verify `chrt -p` reports SCHED_FIFO.

### 5.3 E2E
helix-rtos → helix-pipeline hot-path goroutine → measure context-switch rate before / after promotion.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: attempt to escalate priority above the host's RLIMIT_RTPRIO (must fail).

### 5.5 Benchmarking
SCHED_FIFO promotion p999 ≤ 100 µs; CPU migration count under load (target: zero for pinned threads).

### 5.6 Chaos
Inject CPU contention; verify pinned thread maintains its CPU; involuntary context switches stay below 100/s.

### 5.7 Stress
24-hour pinned hot-path; zero migration; involuntary CSW stays below baseline + 10 %.

### 5.8 Smoke
30-second post-deploy: promote + verify `chrt -p` output.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`08_controller_input_low_latency/02_sched_fifo_under_high_load.scenario` — full topology with controller round-trip while host CPUs are 80 % loaded with non-HelixPlay workload; verify p999 round-trip stays within budget.

---

## 6. Challenges Entry-Point (S03 §4 row #12)

**Topology:** `08_controller_input_low_latency`. **Scenario:** `02_sched_fifo_under_high_load.scenario.yaml`. **Why this scenario.** SCHED_FIFO's value prop is realised under host load — the scenario demonstrates that HelixPlay's round-trip survives a noisy neighbour. **Baseline:** p999 controller round-trip ≤ 8 ms even with 80 % background load; involuntary CSW count baseline-parity.

---

## 7. R-18 Inheritance

`helix-rtos` imports `helix-r18-safeexec` for boundary subprocess invocations (kernel-feature detection). Hot path is pure syscall + cgroup-fs writes. CAP_SYS_NICE is a documented carve-out (S02 §8.4 precedent), narrowly scoped — does not relax R-18's command-level deny-list.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on operator-deployment kernel matrix; PREEMPT_RT-required deployments graduate independently from stock-kernel deployments.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_RTOS_PRIORITY`              | `50`          | int [1, 99]             | SCHED_FIFO priority (50 is mid-range; > 70 risks starvation of system processes). |
| `HELIX_RTOS_PINNED_CPUS`           | `auto`        | csv int / `auto`        | CPUs to pin to; `auto` reads /sys/devices/system/cpu/isolated.        |
| `HELIX_RTOS_REQUIRE_ISOLATED`      | `false`       | bool                    | Refuse to start if `isolcpus=` boot param missing.                    |
| `HELIX_RTOS_REQUIRE_PREEMPT_RT`    | `false`       | bool                    | Refuse to start on non-PREEMPT_RT kernel.                             |
| `HELIX_RTOS_CGROUP_PATH`           | `/sys/fs/cgroup/helix-hot` | filesystem path | Path to the dedicated cgroup; must exist and be writable.            |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `PromoteCurrent()`                    | 30 µs    | 60 µs   | 100 µs  | sched_setscheduler + sched_setaffinity.                          |
| `CPUSet.Apply()`                      | 10 µs    | 20 µs   | 30 µs   | Affinity-only.                                                   |
| `CGroup.AttachThread()`               | 50 µs    | 100 µs  | 200 µs  | cgroup-fs write.                                                 |
| Migration rate (pinned thread)        | 0/min    | 0/min   | 0/min   | Hard target; non-zero is a defect.                              |
| Involuntary CSW (pinned, idle host)   | < 10/s   | < 50/s  | < 100/s | Idle steady-state baseline.                                      |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `rtos: ErrCapSysNiceMissing`                  | Container missing CAP_SYS_NICE                                | Add `--cap-add=SYS_NICE`; verify with `capsh --print`.                    |
| `rtos: ErrRTPRIORLimitExceeded`               | Requested priority > RLIMIT_RTPRIO                             | Adjust `prlimit --rtprio=99` host-side; reduce priority.                  |
| `rtos: ErrIsolatedCPUsRequired`               | `WithIsolatedCPUsOnly` set but `isolcpus=` empty              | Add `isolcpus=4-7` to host kernel cmdline; reboot.                        |
| `rtos: ErrCGroupV1Detected`                   | Container running on cgroup-v1 host                            | Migrate host to unified-hierarchy cgroup-v2; document in operator runbook. |
| `rtos: ErrCPUOffline`                          | Pinned CPU is offline                                          | Verify `/sys/devices/system/cpu/cpu*/online`; pick a different CPU.       |

### 9.4 Migration from default scheduling

A consumer migrating from default Go scheduling:

1. Identify the hot-path goroutine (typically the encoder dispatcher or input poller).
2. Wrap it with `runtime.LockOSThread()` + `rtos.PromoteCurrent(50)` at goroutine start.
3. Add a `defer rt.Demote()` for cleanup.
4. Apply `CPUSet` if specific CPU pinning is required.
5. Add OTLP span around the promote/demote; metric for `Stats().Migrations` (alert: > 0).
6. Verify with `chrt -p $TID` post-deploy.

The migration is documented in `docs/migration-from-default-scheduling.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_rtos_promote_total`                      | counter    | Promotion attempts, labelled `result={ok, denied, error}`.                   |
| `helix_rtos_threads_promoted`                   | gauge      | Currently-promoted thread count.                                             |
| `helix_rtos_migration_total`                    | counter    | CPU migrations observed; alert if > 0 sustained.                             |
| `helix_rtos_involuntary_csw_total`              | counter    | Involuntary context switches per thread.                                     |
| `helix_rtos_voluntary_csw_total`                | counter    | Voluntary context switches per thread.                                       |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C20 §3](../../04_Latency/06_RealTime_OS_and_Scheduling.md) — origin                       | Origin chapter; full RT-thread + cgroup + IRQ API.                                   |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline's encoder + dispatcher goroutines run under SCHED_FIFO.                     |
| [C21 §6](../../04_Latency/07_Controller_Input_Optimization.md) — Input                    | Input poll loop runs under SCHED_FIFO with CPU pinning.                              |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-rtos-A           | PREEMPT_RT requirement — soft (default off) or hard (default on for v1.0.0)?                                  | C20 §3 next revision                                |
| OQ-rtos-B           | Containerised cgroup-v2 — interaction with operator's outer cgroup hierarchy (Kubernetes, Podman)?            | `08_Operations/01_Container_CI_CD.md`              |
| OQ-rtos-C           | IRQ-affinity helper scope — should the submodule own it, or should ops own it via `tuned-adm`?                | `08_Operations/04_Observability_and_Events.md`     |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/06_RealTime_OS_and_Scheduling.md`](../../04_Latency/06_RealTime_OS_and_Scheduling.md) §3 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #12                                 |
| [`../02_Containers_Submodule.md`](../02_Containers_Submodule.md) §8 |   627 | 2026-04-30 | CAP_SYS_NICE carve-out precedent               |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Promote/Demote + CPUSet + CGroup + IRQ paths).                    |
| Integration    | Real container with SYS_NICE; chrt verification.                                              |
| E2E            | helix-rtos → helix-pipeline hot-path; CSW rate before/after measurement.                      |
| Security       | govulncheck + Snyk + Trivy + RLIMIT_RTPRIO escalation rejection.                              |
| Benchmarking   | Promote p999 ≤ 100 µs; migration count = 0 for pinned threads.                                |
| Chaos          | CPU-contention injection; pinned thread holds CPU.                                            |
| Stress         | 24-hour pinned; zero migration; CSW within +10 % of baseline.                                 |
| Smoke          | 30-second promote + chrt verification.                                                        |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `08_controller_input_low_latency/02_sched_fifo_under_high_load` baseline-parity.              |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-rtos.md` — 2026-04-30.
