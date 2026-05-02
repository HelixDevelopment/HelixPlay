# HelixPlay Cloud Gaming Platform - Advanced Phases & Architecture Deep-Dive

**Document ID:** `HP-IMPL-ADV-001`
**Version:** 1.0.0
**Status:** Implementation Plan - Advanced Phases (P07-P13) & Architecture Deep-Dive
**Date:** 2025-01
**Classification:** Engineering Implementation Specification

---

## Table of Contents

- [Part A: Implementation Phases P07-P13](#part-a-implementation-phases-p07-p13)
  - [Phase 07: Latency Optimization](#phase-07-latency-optimization-p1)
  - [Phase 08: Audio Surround](#phase-08-audio-surround-p2)
  - [Phase 09: Recording & Replay](#phase-09-recording--replay-p2)
  - [Phase 10: Monetization & Auth](#phase-10-monetization--auth-p2)
  - [Phase 11: Hardening & Security](#phase-11-hardening--security-p2)
  - [Phase 12: Beta Launch](#phase-12-beta-launch-p2)
  - [Phase 13: GA Release](#phase-13-ga-release-p1)
- [Part B: Architecture Deep-Dive Implementation Tasks](#part-b-architecture-deep-dive-implementation-tasks)
  - [B1: Streaming Pipeline Implementation](#b1-streaming-pipeline-implementation)
  - [B2: Controller Input Pipeline Implementation](#b2-controller-input-pipeline-implementation)
  - [B3: Capture & Encode Pipeline Implementation](#b3-capture--encode-pipeline-implementation)
  - [B4: Client Architecture Implementation](#b4-client-architecture-implementation)
  - [B5: Catalog & Content Pipeline](#b5-catalog--content-pipeline)
  - [B6: White-Label & Theming](#b6-white-label--theming)
  - [B7: Operations & Observability](#b7-operations--observability)
- [Appendix A: Submodule Registry](#appendix-a-submodule-registry)
- [Appendix B: Dependency Graph](#appendix-b-dependency-graph)

---

## Part A: Implementation Phases P07-P13

### Legend

| Field | Meaning |
|-------|---------|
| **Priority** | P1 = Critical path (blocks GA), P2 = Important (feature-complete), P3 = Enhancement |
| **Effort** | Person-weeks (pw), Person-days (pd) |
| **AC** | Acceptance Criteria |
| **Files** | Files to create (+) or modify (~) |
| **Deps** | Cross-phase dependencies |

---

### Phase 07: Latency Optimization (P1)

**Phase Goal:** Achieve sub-30ms LAN p999 and sub-50ms WAN p999 glass-to-glass latency through kernel-level tuning, zero-copy IPC, RTOS scheduling, and lock-free data structures.

**Phase Duration:** 8 weeks
**Engineering Team:** 4 FTE (2 systems, 1 kernel, 1 Go performance)
**Deps:** P01-P06 (foundation infrastructure operational)

---

#### Task P07-T01: PREEMPT_RT Kernel Patching & Validation

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T01 |
| **Title** | PREEMPT_RT Linux Kernel Patching for Host Fleet |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Kernel Engineer |

**Description:**
Patch host fleet Linux kernels with PREEMPT_RT realtime patchset. Validate scheduling determinism under full gaming load. Build CI pipeline for kernel image generation and automated boot testing.

**Files:**
```
+ kernel/patches/preempt-rt-helix.patch       # Helix-specific RT tuning
+ kernel/configs/helix-host-x86_64.defconfig   # Defconfig for host nodes
+ kernel/scripts/build-kernel.sh               # Automated kernel build
+ kernel/scripts/validate-rt.sh                # cyclictest validation (>100k samples)
~ kernel/Makefile                              # Add kernel artifact targets
~ .github/workflows/kernel-build.yml           # CI: kernel build + cyclictest
```

**Implementation Details:**
- Base kernel: 6.6 LTS with PREEMPT_RT patchset
- Target maximum scheduling latency: <10 microseconds (cyclictest histogram 99.9th percentile)
- Disable CFS bandwidth control for `helix-rt` cgroup
- Enable `CONFIG_NO_HZ_FULL` for isolated CPU cores
- Disable `CONFIG_CPU_FREQ_DEFAULT_GOV_SCHEDUTIL`; use `performance` governor
- Apply `rcu_nocbs` for isolated cores
- Use `tuned` profile `latency-performance` as base, customize for HelixPlay

**Acceptance Criteria:**
1. `cyclictest -D 3600 -m -S -p 90 -i 200 -h 400` shows p99.9 < 10us under 4K60 capture+encode load
2. Kernel boots successfully on 100% of host fleet hardware profiles (AWS g4dn/g5, local RTX nodes)
3. Kernel build completes in CI < 15 minutes from clean state
4. No kernel panics or soft-lockups during 72-hour burn-in test
5. `CONFIG_DEBUG_PREEMPT` disabled for production builds (performance)

---

#### Task P07-T02: SCHED_FIFO Thread Priorities & CPU Isolation

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T02 |
| **Title** | Real-Time Thread Scheduling & CPU Isolation (isolcpus) |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Systems Engineer |

**Description:**
Implement `SCHED_FIFO` thread priority assignment across all HelixPlay host processes. Isolate critical-path threads to dedicated CPU cores using `isolcpus` and `cset shield`. Define priority hierarchy for capture → encode → network → input → housekeeping.

**Files:**
```
+ helix-rtos/pkg/rtsched/priority.go          # Priority hierarchy definitions
+ helix-rtos/pkg/rtsched/affinity.go          # CPU affinity/ isolation management
+ helix-rtos/pkg/rtsched/rtsched.go           # Main RT scheduler interface
+ helix-rtos/cmd/rtsetup/main.go              # Host node RT setup utility
+ helix-rtos/configs/priority-defaults.yaml    # Per-process priority defaults
~ sunshine/src/platform/linux/misc.cpp         # Sunshine RT thread hooks
~ sunshine/src/platform/linux/display.cpp      # Capture thread affinity
```

**Thread Priority Hierarchy:**

| Priority | Thread | sched_policy | CPU Affinity |
|----------|--------|-------------|--------------|
| 99 | Capture frame acquisition | `SCHED_FIFO` | `isolcpus` core 0 |
| 97 | Video encode (NVENC submit) | `SCHED_FIFO` | `isolcpus` core 0 |
| 95 | Network TX (WebRTC packet send) | `SCHED_FIFO` | `isolcpus` core 1 |
| 93 | Controller input read | `SCHED_FIFO` | `isolcpus` core 1 |
| 90 | Audio capture/encode | `SCHED_FIFO` | `isolcpus` core 2 |
| 80 | WebRTC ICE/STUN processing | `SCHED_FIFO` | `isolcpus` core 2 |
| 50 | Session orchestration | `SCHED_OTHER` | General pool |
| 10 | Metrics, logging, housekeeping | `SCHED_IDLE` | General pool |

**Acceptance Criteria:**
1. `schedtool` confirms all threads at assigned priorities during active session
2. `taskset -pc <pid>` shows correct CPU isolation for capture/encode threads
3. No priority inversion events detected by `trace-cmd` during 1-hour session
4. Latency variance (stddev) reduced by >40% vs. CFS baseline at same load
5. Dynamic priority adjustment API available for thermal throttling scenarios

---

#### Task P07-T03: GOCACHEPROG Remote Build Caching

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T03 |
| **Title** | GOCACHEPROG Remote Build Cache (Bazel-remote / Turborepo) |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Build Engineer |

**Description:**
Implement `GOCACHEPROG` remote build cache to reduce CI build times across 29 submodules. Deploy bazel-remote or equivalent S3-backed cache. Configure all CI pipelines to use shared cache.

**Files:**
```
+ .github/scripts/gocacheprog.sh              # GOCACHEPROG wrapper
+ infra/cache/bazel-remote.yml                # Docker Compose for cache server
+ infra/cache/s3-backend.tf                   # S3 backend Terraform
~ .github/workflows/ci-all.yml                # Add GOCACHEPROG env
~ Makefile                                     # Export GOCACHEPROG
```

**Acceptance Criteria:**
1. Cold CI build (all 29 modules) completes in < 8 minutes with warm cache
2. Cache hit rate > 85% for incremental builds
3. Cache backend supports 1000+ concurrent CI jobs (horizontal scaling)
4. Fallback to local build on cache miss (no hard dependency)
5. Cache eviction policy: 7-day TTL, LRU within TTL

---

#### Task P07-T04: Kernel Parameter Tuning & sysctl Profiles

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T04 |
| **Title** | Kernel sysctl Tuning for Low-Latency Networking & IPC |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Systems Engineer |

**Description:**
Create tuned sysctl profiles optimized for low-latency gaming workloads. Tune network buffers, TCP/UDP parameters, memory management, and scheduler settings.

**Files:**
```
+ helix-rtos/configs/sysctl-gaming.conf        # Core gaming sysctl profile
+ helix-rtos/configs/sysctl-udp-optim.conf     # UDP/WebRTC optimization
+ helix-rtos/configs/sysctl-memory.conf        # Memory management tuning
+ helix-rtos/cmd/tune-host/main.go             # Host tuning application
```

**Key sysctl Parameters:**
```
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.udp_rmem_min = 1048576
net.ipv4.udp_wmem_min = 1048576
net.core.netdev_max_backlog = 65536
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_notsent_lowat = 16384
vm.swappiness = 1
vm.dirty_ratio = 5
vm.dirty_background_ratio = 2
kernel.sched_rt_runtime_us = 950000
kernel.timer_migration = 0
```

**Acceptance Criteria:**
1. UDP socket buffer sizes confirmed via `ss -npm` during active session
2. `netperf` UDP_RR latency < 50us p99 localhost
3. No packet drops at 100Mbps sustained UDP send rate
4. Tuning profiles applied automatically on host node bootstrap

---

#### Task P07-T05: io_uring Integration (helix-iouring submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T05 |
| **Title** | io_uring Asynchronous I/O Integration |
| **Priority** | P1 |
| **Effort** | 2.5pw |
| **Owner** | Systems Engineer |

**Description:**
Implement io_uring-based async I/O for file recording, network send/receive, and shared memory operations. Create reusable Go wrapper via `x/sys/unix` raw syscalls (no CGO dependency).

**Files:**
```
+ helix-iouring/go.mod                        # Module: github.com/helixplay/helix-iouring
+ helix-iouring/pkg/uring/ring.go             # io_uring ring management
+ helix-iouring/pkg/uring/sqe.go              # Submission queue entry builders
+ helix-iouring/pkg/uring/cqe.go              # Completion queue event handlers
+ helix-iouring/pkg/uring/ops/read.go         # vectored read operations
+ helix-iouring/pkg/uring/ops/write.go        # vectored write operations
+ helix-iouring/pkg/uring/ops/send.go         # network send operations
+ helix-iouring/pkg/uring/ops/recv.go         # network receive operations
+ helix-iouring/pkg/uring/pool/buffer.go      # registered buffer pool
+ helix-iouring/pkg/uring/bench/bench_test.go # Benchmarks vs. epoll/sync
+ helix-iouring/LICENSE                       # MIT
```

**API Surface:**
```go
type Ring struct { /* io_uring ring */ }
func Setup(entries uint, params *Params) (*Ring, error)
func (r *Ring) QueueReadv(fd int, iovec []syscall.Iovec, offset uint64) error
func (r *Ring) QueueWritev(fd int, iovec []syscall.Iovec, offset uint64) error
func (r *Ring) QueueSend(fd int, buf []byte, flags int) error
func (r *Ring) QueueRecv(fd int, buf []byte, flags int) error
func (r *Ring) Submit() (uint, error)
func (r *Ring) WaitCQEs(count uint, timeout time.Duration) ([]CQE, error)
```

**Acceptance Criteria:**
1. `io_uring` ring operates in polled mode (`IORING_SETUP_IOPOLL`) for NVMe I/O
2. Vectored write throughput to NVMe > 3GB/s sustained (single thread)
3. Latency of io_uring write completion < 2us p99 for 4KB writes
4. Registered buffer pool (`IORING_REGISTER_BUFFERS`) reduces CPU by >20%
5. Graceful fallback to `epoll`/`select` on kernel < 5.10
6. Benchmark suite shows io_uring outperforms sync I/O by >3x in all metrics

---

#### Task P07-T06: Lock-Free Data Structures (helix-lockfree submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T06 |
| **Title** | Lock-Free SPSC/MPSC Ring Buffers & Queues |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Performance Engineer |

**Description:**
Implement platform-native lock-free data structures in Go using atomic operations and memory barriers. Focus on SPSC (Single Producer Single Consumer) ring buffers for capture→encode handoff and MPSC (Multi Producer Single Consumer) queues for controller input aggregation.

**Files:**
```
+ helix-lockfree/go.mod                       # Module: github.com/helixplay/helix-lockfree
+ helix-lockfree/pkg/ring/spsc.go             # SPSC ring buffer (power-of-2)
+ helix-lockfree/pkg/ring/mpsc.go             # MPSC bounded queue
+ helix-lockfree/pkg/ring/mpmc.go             # MPMC bounded queue (fallback)
+ helix-lockfree/pkg/atomic/seqlock.go        # Sequence locks for metrics
+ helix-lockfree/pkg/atomic/snapshot.go       # Lock-free snapshot reader
+ helix-lockfree/pkg/cacheline/pad.go         # Cache-line padding utilities
+ helix-lockfree/pkg/bench/spsc_bench_test.go # Go benchmarks
+ helix-lockfree/internal/asm/spsc_amd64.s    # Assembly-optimized x86_64
+ helix-lockfree/internal/asm/spsc_arm64.s    # Assembly-optimized ARM64
+ helix-lockfree/LICENSE                      # MIT
```

**SPSC Ring Buffer Spec:**
```go
type SPSCRing[T any] struct {
    _pad0   [cacheline.Size]byte
    head    atomic.Uint64       // producer index
    _pad1   [cacheline.Size]byte
    tail    atomic.Uint64       // consumer index
    _pad2   [cacheline.Size]byte
    mask    uint64
    buffer  unsafe.Pointer      // contiguous array
}

func NewSPSC[T any](size uint) *SPSCRing[T]
func (r *SPSCRing[T]) Push(val T) bool    // false if full (non-blocking)
func (r *SPSCRing[T]) Pop() (T, bool)     // zero-value if empty (non-blocking)
func (r *SPSCRing[T]) Available() uint
```

**Acceptance Criteria:**
1. `SPSC.Push` latency < 15ns p99 (measured via `testing.B` + `perfevents`)
2. `SPSC.Pop` latency < 15ns p99
3. No memory allocations in hot path (verified via `go test -memprofile`)
4. Cache-line false sharing eliminated (verified via `perf c2c`)
5. Correctness validated with `go test -race` and stress test (100M ops, 32 threads)
6. AMD64 assembly path shows >15% improvement over pure Go on x86_64

---

#### Task P07-T07: GPU Direct Integration (helix-gpu-direct submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T07 |
| **Title** | GPU Direct (GPUDirect RDMA / Resizable BAR) |
| **Priority** | P2 |
| **Effort** | 2pw |
| **Owner** | GPU Systems Engineer |

**Description:**
Enable zero-copy frame transfer from GPU framebuffer to encoder via GPUDirect RDMA (NVIDIA) and Resizable BAR. Eliminate PCIe round-trip for frame capture path.

**Files:**
```
+ helix-gpu-direct/go.mod                     # Module: github.com/helixplay/helix-gpu-direct
+ helix-gpu-direct/pkg/nv/gdrdrv.go           # NVIDIA gdrdrv bindings
+ helix-gpu-direct/pkg/nv/p2p.go              # Peer-to-peer memory mapping
+ helix-gpu-direct/pkg/amd/amdgpudirect.go    # AMD GPU Direct equivalent
+ helix-gpu-direct/pkg/bar/resizable.go       # Resizable BAR management
+ helix-gpu-direct/pkg/common/mapping.go      # Generic GPU memory mapping
+ helix-gpu-direct/pkg/common/barrier.go      # Memory barrier primitives
+ helix-gpu-direct/cmd/gpudirect-test/main.go # Validation tool
+ helix-gpu-direct/LICENSE                    # MIT
```

**Architecture:**
- NVIDIA: Use `nvidia_p2p_get_pages()` + `gdrdrv` kernel module for pinned GPU memory export
- AMD: Use `amdgpu_ttm_tt_get_user_pages_done()` for ROCm-based direct access
- Intel: Use Level Zero `zeMemGetAllocProperties()` for shared allocations
- Fallback: `cudaMemcpy2DAsync` + registered host memory (1 copy instead of 2)

**Acceptance Criteria:**
1. Frame capture→encode path eliminates host memory copy (verified via NVIDIA Nsight Systems)
2. GPU→encoder latency < 0.5ms per frame (4K)
3. Works on RTX 40-series (GPUDirect), RTX 30-series (resizable BAR), Intel Arc (Level Zero)
4. Graceful fallback to pinned host memory path on unsupported hardware
5. Memory pinning respects host RAM budget (cgroup v2 memory limits)

---

#### Task P07-T08: Shared Memory Zero-Copy IPC (helix-shm submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T08 |
| **Title** | Shared Memory Zero-Copy Inter-Process Communication |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Systems Engineer |

**Description:**
Implement POSIX shared memory-based zero-copy IPC for frame data and controller input between Sunshine++ host agent processes. Replace pipe/socket-based transfer with shared memory regions + atomic synchronization.

**Files:**
```
+ helix-shm/go.mod                            # Module: github.com/helixplay/helix-shm
+ helix-shm/pkg/shm/region.go                 # Shared memory region management
+ helix-shm/pkg/shm/segment.go                # Memory segment allocator
+ helix-shm/pkg/shm/ring.go                   # SHM-backed ring buffer
+ helix-shm/pkg/shm/sync.go                   # Atomic synchronization primitives
+ helix-shm/pkg/shm/frame.go                  # Frame buffer shared memory layout
+ helix-shm/pkg/shm/input.go                  # Input packet shared memory layout
+ helix-shm/pkg/shm/cgroup.go                 # cgroup v2 memory integration
+ helix-shm/cmd/shm-test/main.go              # SHM validation tool
+ helix-shm/internal/unix/shm_linux.go        # Linux-specific shm syscalls
+ helix-shm/LICENSE                           # MIT
```

**Frame Buffer Layout:**
```go
type FrameBuffer struct {
    Magic       uint32          // 'HLXF'
    Version     uint16          // 1
    Format      uint32          // DXGI_FORMAT / VK_FORMAT
    Width       uint32
    Height      uint32
    Pitch       uint32
    HDRMetadata [64]byte        // HDR10 metadata blob
    ProducerSeq atomic.Uint64   // write sequence number
    ConsumerSeq atomic.Uint64   // read sequence number
    Data        [0]byte         // frame data follows (VLA pattern)
}
```

**Acceptance Criteria:**
1. Frame transfer latency (capture → encoder) < 50 microseconds
2. Shared memory regions properly cleaned up on process crash (`memfd_create` + `MFD_CLOEXEC`)
3. cgroup v2 `memory.max` respected by SHM allocator
4. Works across privilege boundaries (unprivileged client ↔ privileged capture daemon)
5. `ftruncate`/`mmap` cycle time < 1ms for 4K frame buffer allocation
6. Verified with `strace -e trace=mmap,munmap` showing zero memcpy in hot path

---

#### Task P07-T09: Memory Pool Optimization (helix-mempool, helix-allocator)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T09 |
| **Title** | Slab Memory Pool & Custom Allocator |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Performance Engineer |

**Description:**
Implement slab-based memory pools for fixed-size allocations (frame buffers, network packets, controller state) and a bump allocator for short-lived session data. Eliminate GC pressure in hot paths.

**Files:**
```
+ helix-mempool/go.mod                        # Module: github.com/helixplay/helix-mempool
+ helix-mempool/pkg/slab/pool.go              # Generic slab allocator
+ helix-mempool/pkg/slab/frame.go             # Frame buffer pool (fixed 4K/1080p)
+ helix-mempool/pkg/slab/packet.go            # Network packet pool
+ helix-mempool/pkg/slab/controller.go        # Controller state pool
+ helix-mempool/pkg/bump/arena.go             # Bump allocator for session data
+ helix-mempool/pkg/metrics/gc.go             # GC pressure monitoring
+ helix-mempool/LICENSE                       # MIT
```

**Files (helix-allocator):**
```
+ helix-allocator/go.mod                      # Module: github.com/helixplay/helix-allocator
+ helix-allocator/pkg/linear/allocator.go     # Linear/bump allocator
+ helix-allocator/pkg/linear/region.go        # Memory region manager
+ helix-allocator/pkg/linear/sync.go          # Thread-safe region operations
+ helix-allocator/LICENSE                     # MIT
```

**Pool Configuration:**
```go
var DefaultPools = map[string]PoolConfig{
    "frame-4k":    {Size: 3840 * 2160 * 4, Count: 4, NUMA: true},
    "frame-1080p": {Size: 1920 * 1080 * 4, Count: 8, NUMA: true},
    "packet":      {Size: 2048, Count: 1024, NUMA: false},
    "controller":  {Size: 256, Count: 16, NUMA: false},
}
```

**Acceptance Criteria:**
1. Zero heap allocations during active streaming session (`GODEBUG=allocfreetrace=1`)
2. GC pause times < 100 microseconds during 1-hour session
3. `slab.Alloc` / `slab.Free` operations < 20ns p99
4. NUMA-aware allocation on dual-socket host nodes
5. Graceful OOM handling with frame drop instead of process termination

---

#### Task P07-T10: IRQ Affinity & Network Interrupt Steering

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T10 |
| **Title** | IRQ Affinity Tuning & Network Interrupt Steering |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Systems Engineer |

**Description:**
Configure IRQ affinity for network interfaces to isolated cores. Use `irqbalance` exclusion or manual `/proc/irq/*/smp_affinity` configuration. Enable NIC hardware flow steering if available.

**Files:**
```
+ helix-rtos/configs/irq-affinity.rules        # IRQ affinity rules
+ helix-rtos/scripts/setup-irq.sh              # IRQ setup script
~ helix-rtos/cmd/tune-host/main.go             # Add IRQ tuning
```

**Acceptance Criteria:**
1. NIC RX/TX interrupts isolated to non-`isolcpus` cores
2. No network IRQ handled on capture/encode isolated cores
3. Hardware flow steering (`ethtool -n rx-flow-hash`) configured if NIC supports
4. `mpstat -I CPU` shows <1% interrupt time on isolated cores

---

#### Task P07-T11: cgroup v2 Resource Control

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T11 |
| **Title** | cgroup v2 Session Resource Isolation |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Systems Engineer |

**Description:**
Implement cgroup v2 resource control for per-session resource isolation. Enforce CPU, memory, and I/O limits. Use systemd slice integration for automatic cleanup.

**Files:**
```
+ helix-rtos/pkg/cgroup/v2.go                  # cgroup v2 controller
+ helix-rtos/pkg/cgroup/session.go             # Per-session cgroup management
+ helix-rtos/pkg/cgroup/systemd.go             # systemd slice integration
```

**Acceptance Criteria:**
1. Each session runs in dedicated cgroup with configurable limits
2. `memory.high` enforcement triggers graceful quality reduction (not OOM kill)
3. `cpu.weight` allows proportional sharing without hard caps
4. Automatic cgroup cleanup on session termination (systemd `After=`)

---

#### Phase 07: Definition of Done

- [ ] All P1 tasks (T01, T02, T04-T06, T08, T09, T10) completed and passing AC
- [ ] io_uring benchmarks show >3x improvement over sync I/O
- [ ] Lock-free SPSC ring buffer < 15ns p99 push/pop latency
- [ ] SHM IPC achieves < 50us frame transfer latency
- [ ] cyclictest p99.9 < 10us under full load
- [ ] Zero heap allocations during active streaming session
- [ ] p999 glass-to-glass latency: LAN < 30ms, WAN < 50ms (measured)

---

### Phase 08: Audio Surround (P2)

**Phase Goal:** Deliver immersive multi-channel audio with Opus MultiStream surround (up to 7.1), Dolby Atmos spatial audio, AC3/EAC3 passthrough, and eARC support. Achieve <5ms A/V sync drift.

**Phase Duration:** 6 weeks
**Engineering Team:** 3 FTE (1 audio specialist, 1 Go backend, 1 client)
**Deps:** P04 (client architecture), P07 (latency optimization)

---

#### Task P08-T01: Opus MultiStream Encoder (up to 7.1 channels)

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T01 |
| **Title** | Opus MultiStream Surround Encoder Integration |
| **Priority** | P2 |
| **Effort** | 2pw |
| **Owner** | Audio Engineer |

**Description:**
Integrate Opus MultiStream API (`opus_multistream_encoder_create`) for surround sound encoding up to 7.1 channels. Implement channel mapping families 0, 1 (Vorbis), and 255 (raw/manual). Support 48kHz sample rate with 20ms frame size.

**Files:**
```
+ helix-audio/go.mod                           # Module: github.com/helixplay/helix-audio
+ helix-audio/pkg/opus/multistream.go          # Opus MultiStream wrapper
+ helix-audio/pkg/opus/surround.go             # Surround channel mappings
+ helix-audio/pkg/opus/encoder.go              # Generic Opus encoder interface
+ helix-audio/pkg/opus/decoder.go              # Opus decoder interface
+ helix-audio/pkg/format/channel.go            # Channel layout definitions
+ helix-audio/pkg/format/layout.go             # Standard layouts (mono→7.1)
+ helix-audio/pkg/pcm/buffer.go                # PCM buffer management
+ helix-audio/pkg/pcm/resample.go              # Speex resampler wrapper
+ helix-audio/pkg/pcm/mix.go                   # Channel mixing utilities
+ helix-audio/LICENSE                          # MIT
```

**Channel Layout Support:**

| Layout | Channels | Opus Mapping Family | Channel Order |
|--------|----------|---------------------|---------------|
| Mono | 1 | 0 | C |
| Stereo | 2 | 0 | L, R |
| 3.0 | 3 | 1 | L, R, C |
| Quadraphonic | 4 | 1 | L, R, Ls, Rs |
| 5.1 | 6 | 1 | L, R, C, LFE, Ls, Rs |
| 7.1 | 8 | 1 | L, R, C, LFE, Ls, Rs, Rls, Rrs |

**Acceptance Criteria:**
1. Opus MultiStream encodes 7.1 @ 48kHz to < 512kbps with transparency (> 128kbps per stereo pair equivalent)
2. All channel mapping families (0, 1, 255) functional and tested
3. Encoder latency < 2x frame size (40ms @ 20ms frames, including look-ahead)
4. Bitrate adaptation without audible artifacts during gameplay
5. Go wrapper provides CGO-free build option via pure-Go Opus fallback (lower quality acceptable)

---

#### Task P08-T02: AC3/EAC3 Passthrough

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T02 |
| **Title** | AC3/EAC3 (Dolby Digital/Digital+) Bitstream Passthrough |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Audio Engineer |

**Description:**
Implement pass-through mode for AC3 and EAC3 compressed audio bitstreams from game/application to client without transcoding. Preserve exact bitstream for external decoder (AVR, soundbar, TV).

**Files:**
```
+ helix-audio/pkg/passthrough/ac3.go           # AC3 passthrough handler
+ helix-audio/pkg/passthrough/eac3.go          # EAC3 passthrough handler
+ helix-audio/pkg/passthrough/detector.go      # Format auto-detection
+ helix-audio/pkg/passthrough/syncframe.go     # Sync frame boundary detection
~ helix-audio/pkg/opus/encoder.go              # Add passthrough bypass mode
```

**Acceptance Criteria:**
1. AC3/EAC3 bitstream passes through unchanged (bit-exact verification)
2. Sync frame boundaries correctly detected and preserved
3. IEC 61937 encapsulation for SPDIF/eARC output on client side
4. Automatic format detection with fallback to Opus transcode
5. Metadata (dialog normalization, DRC) preserved through passthrough

---

#### Task P08-T03: Dolby Atmos Support

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T03 |
| **Title** | Dolby Atmos Spatial Audio (DD+JOC) |
| **Priority** | P2 |
| **Effort** | 1.5pw |
| **Owner** | Audio Engineer |

**Description:**
Implement Dolby Atmos support via EAC3 with Joint Object Coding (JOC). Encode object-based audio metadata alongside channel bed. Client-side rendering via Dolby MS12 decoder or platform Atmos renderer.

**Files:**
```
+ helix-audio/pkg/atmos/joc.go                 # JOC metadata parser
+ helix-audio/pkg/atmos/object.go              # Audio object extraction
+ helix-audio/pkg/atmos/renderer.go            # Platform Atmos renderer abstraction
+ helix-audio/pkg/atmos/header.go              # Atmos-specific header handling
```

**Acceptance Criteria:**
1. Atmos content identified and routed correctly in pipeline
2. JOC metadata preserved through encoding chain
3. Platform Atmos rendering works on: Apple TV 4K (tvOS), Android TV (API 30+), Windows Sonic
4. Fallback to 7.1 channel bed when Atmos renderer unavailable

---

#### Task P08-T04: eARC Audio Return Channel

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T04 |
| **Title** | eARC (Enhanced Audio Return Channel) Client Support |
| **Priority** | P3 |
| **Effort** | 1pw |
| **Owner** | Client Engineer |

**Description:**
Support eARC output on TV/console clients for uncompressed multi-channel audio passthrough to external audio systems. Implement HDMI eARC handshake and latency compensation.

**Files:**
```
~ helix-flutter/lib/platform/tv/earc.dart      # eARC TV platform channel
~ helix-flutter/android/src/main/kotlin/Ear.kt  # Android eARC native
~ helix-wails/frontend/src/audio/earc.ts        # Desktop eARC (if applicable)
```

**Acceptance Criteria:**
1. eARC handshake completes successfully on HDMI 2.1 ports
2. Uncompressed 7.1 PCM output at 48kHz to eARC receiver
3. Lip-sync adjustment within +/- 2ms via eARC latency reporting
4. Automatic fallback to TV speakers on eARC disconnect

---

#### Task P08-T05: A/V Synchronization Mechanism

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T05 |
| **Title** | Audio/Video Sync with Frame-Time Alignment |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Streaming Engineer |

**Description:**
Implement A/V synchronization using RTP timestamp alignment and frame-time audio rendering. Target < 5ms A/V drift with automatic drift correction.

**Files:**
```
+ helix-stream/pkg/sync/avsync.go              # A/V sync controller
+ helix-stream/pkg/sync/drift.go               # Drift detection & correction
+ helix-stream/pkg/sync/timestamp.go           # RTP timestamp alignment
+ helix-audio/pkg/render/clock.go              # Audio clock reference
```

**Sync Strategy:**
1. **Presentation Timestamps:** Both audio and video RTP packets carry capture-time NTP timestamps
2. **Audio Master:** Audio clock drives presentation (less variable than video)
3. **Drift Detection:** Compare expected vs. actual audio sample consumption every 100ms
4. **Correction:** Sample rate adjustment (±2%) or audio sample insertion/drop for drift > 2ms
5. **Video Sync:** Video frames delayed/early-released to match audio presentation time

**Acceptance Criteria:**
1. A/V sync drift < 5ms sustained (measured with audio-led test pattern)
2. No audible pops/clicks during drift correction
3. Sync recovers within 200ms after network jitter event
4. Sync mechanism works across all codec combinations (H.264/HEVC/AV1 + Opus)

---

#### Task P08-T06: Audio Pipeline Integration into Sunshine++

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T06 |
| **Title** | Sunshine++ Audio Pipeline Integration |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Audio Engineer |

**Description:**
Integrate the helix-audio module into Sunshine++ host agent. Replace existing audio pipeline with Opus MultiStream surround, passthrough detection, and Atmos routing.

**Files:**
```
~ sunshine/src/audio.cpp                       # Refactor to use helix-audio
~ sunshine/src/audio.h                         # Audio interface updates
~ sunshine/src/stream.cpp                      # A/V sync integration
~ sunshine/src/entry_handler.cpp               # Audio format negotiation
```

**Acceptance Criteria:**
1. Audio pipeline functional end-to-end with all channel configurations
2. Passthrough correctly detected and routed without transcoding
3. A/V sync < 5ms in end-to-end test
4. No audio artifacts during 1-hour stress test

---

#### Phase 08: Definition of Done

- [ ] Opus MultiStream surround encoding up to 7.1 channels functional
- [ ] AC3/EAC3 passthrough bit-exact verified
- [ ] Dolby Atmos detection and routing complete
- [ ] A/V sync drift < 5ms sustained
- [ ] Audio latency added to glass-to-glass budget and within target

---

### Phase 09: Recording & Replay (P2)

**Phase Goal:** Enable gameplay recording with dual-path encoding (stream + record simultaneously), instant replay buffer, local NVMe storage with background cloud sync, and DASH-based replay client.

**Phase Duration:** 6 weeks
**Engineering Team:** 3 FTE (1 video pipeline, 1 storage/sync, 1 client replay)
**Deps:** P03 (capture pipeline), P07 (latency optimization), P08 (audio pipeline)

---

#### Task P09-T01: helix-record Submodule Bootstrap

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T01 |
| **Title** | helix-record Module Creation & CI Pipeline |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Build Engineer |

**Files:**
```
+ helix-record/go.mod                          # Module: github.com/helixplay/helix-record
+ helix-record/.github/workflows/ci.yml        # CI: build + test + lint
+ helix-record/.github/workflows/release.yml   # Release automation
+ helix-record/Makefile                        # Standard targets
+ helix-record/LICENSE                         # MIT
```

**Acceptance Criteria:**
1. Module compiles with `go build ./...`
2. CI passes on Go 1.23 (linux/amd64, linux/arm64)
3. Test coverage > 70% from inception (enforced in CI)

---

#### Task P09-T02: Dual-Path Encoding (Stream + Record)

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T02 |
| **Title** | Dual-Path Simultaneous Encode (1 NVENC Session Constraint) |
| **Priority** | P1 |
| **Effort** | 2.5pw |
| **Owner** | Video Pipeline Engineer |

**Description:**
Implement dual-path encoding where a single NVENC session produces both the low-latency stream and the high-quality recording using split encoding (SEI NAL units for frame duplication) or B-frame enabled recording path. The constraint: only 1 NVENC session per GPU to avoid resource contention.

**Files:**
```
+ helix-record/pkg/encode/dualpath.go          # Dual-path encoder controller
+ helix-record/pkg/encode/split.go             # Frame splitting strategy
+ helix-record/pkg/encode/quality.go           # Recording quality presets
+ helix-record/pkg/nvenc/split.go              # NVENC split encoding
+ helix-record/pkg/qsv/split.go                # QSV split encoding (Intel)
+ helix-record/pkg/amf/split.go                # AMF split encoding (AMD)
+ sunshine/src/video.cpp                       # Add dual-path hooks
```

**Dual-Path Strategies:**

| Strategy | NVENC Sessions | Stream Quality | Record Quality | Overhead |
|----------|---------------|----------------|----------------|----------|
| **Split Encode** | 1 | Low-latency (P-frames only) | High (B-frames, higher bitrate) | ~5% GPU |
| **Frame Copy** | 2 (not allowed) | Native | Native | N/A |
| **Post-Process** | 1 | Low-latency | Re-encode from saved frames | ~30% GPU |

**Selected Architecture:** Split Encode (1 NVENC session, dual output)
- Primary output: low-latency stream (low bitrate, no B-frames)
- Secondary output: recording (high bitrate, B-frames enabled, higher quality preset)
- Frame duplication handled by NVENC hardware (minimal overhead)

**Acceptance Criteria:**
1. Single NVENC session produces both stream and recording
2. Stream latency increase < 2ms vs. single-path encoding
3. Recording quality: 50Mbps HEVC Main10 @ 4K60 (visually lossless)
4. GPU overhead < 10% vs. stream-only operation
5. Graceful degradation: if GPU loaded, recording pauses (stream priority)

---

#### Task P09-T03: MKV/fMP4 Container Support

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T03 |
| **Title** | MKV (Matroska) and fMP4 (CMAF) Container Muxing |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Video Pipeline Engineer |

**Description:**
Implement MKV and fMP4 container muxing for recordings. MKV for local archival, fMP4 for DASH streaming replay. Support chapter markers (game events) and subtitle tracks.

**Files:**
```
+ helix-record/pkg/mux/mkv.go                  # MKV muxer (libmatroska bindings)
+ helix-record/pkg/mux/fmp4.go                 # fMP4/CMAF muxer
+ helix-record/pkg/mux/chapter.go              # Chapter marker injection
+ helix-record/pkg/mux/metadata.go             # Metadata attachment (game info)
+ helix-record/pkg/mux/segment.go              # fMP4 segment management
```

**Acceptance Criteria:**
1. MKV output playable in VLC, MPC-HC, Kodi, Plex
2. fMP4 segments valid for DASH-IF conformance
3. Chapter markers accurate to within 1 second of in-game event
4. HDR metadata (HDR10/HDR10+/Dolby Vision) preserved in container

---

#### Task P09-T04: Instant Replay Buffer

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T04 |
| **Title** | Rolling Instant Replay Buffer (DVR-style) |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Video Pipeline Engineer |

**Description:**
Implement a configurable rolling buffer (default 30 seconds) that continuously records gameplay to a ring buffer in memory. User can trigger "save replay" to persist buffer to storage.

**Files:**
```
+ helix-record/pkg/replay/buffer.go            # Ring buffer manager
+ helix-record/pkg/replay/segments.go          # Segment ring management
+ helix-record/pkg/replay/trigger.go           # User trigger handler
+ helix-record/pkg/api/replay.go               # Replay API endpoints
```

**Acceptance Criteria:**
1. Rolling buffer maintains last 30 seconds in RAM (configurable 10-300s)
2. "Save replay" action persists buffer to NVMe in < 2 seconds
3. Memory usage bounded: 30s @ 1080p60 ~ 150MB ring buffer
4. No impact on streaming latency (< 1ms addition)

---

#### Task P09-T05: Local NVMe Storage & Background Sync

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T05 |
| **Title** | NVMe Fast Local Storage + Background Cloud Sync |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Storage Engineer |

**Description:**
Implement local NVMe recording storage with background upload to cloud storage (SMB/NFS/FTP/WebDAV). Use io_uring for fast local writes. Implement retry, resume, and bandwidth-throttled sync.

**Files:**
```
+ helix-record/pkg/storage/local.go            # NVMe local storage
+ helix-record/pkg/storage/sync.go             # Background sync orchestrator
+ helix-record/pkg/storage/backends/smb.go     # SMB/CIFS backend
+ helix-record/pkg/storage/backends/nfs.go     # NFS backend
+ helix-record/pkg/storage/backends/ftp.go     # FTP/FTPS backend
+ helix-record/pkg/storage/backends/webdav.go  # WebDAV backend
+ helix-record/pkg/storage/backends/s3.go      # S3-compatible backend
+ helix-record/pkg/storage/queue.go            # Upload queue with retry
+ helix-record/pkg/storage/bandwidth.go        # Bandwidth throttling
```

**Acceptance Criteria:**
1. Local write speed: > 500MB/s sustained to NVMe (io_uring)
2. Background sync at configurable bandwidth limit (default: 50% of available upload)
3. Resume interrupted uploads from last byte (HTTP Range)
4. All 5 backend protocols functional with integration tests
5. Storage cleanup: auto-delete recordings > retention period (configurable)

---

#### Task P09-T06: DASH Replay Client

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T06 |
| **Title** | DASH-based Replay Streaming Client |
| **Priority** | P2 |
| **Effort** | 1.5pw |
| **Owner** | Client Engineer |

**Description:**
Implement DASH client for replay playback across all three client platforms (Wails, Flutter, Angular+WASM). Support trick-play (seek, fast-forward, rewind) and quality adaptation.

**Files:**
```
+ helix-record/pkg/dash/manifest.go            # MPD manifest parser/generator
+ helix-record/pkg/dash/segment.go             # Segment fetch & buffer
+ helix-record/pkg/dash/adaptation.go          # Quality adaptation logic
+ helix-record/pkg/api/dash.go                 # DASH serving API
~ helix-angular/src/app/replay/               # Web replay player
~ helix-flutter/lib/replay/                    # Mobile replay player
~ helix-wails/frontend/src/replay/             # Desktop replay player
```

**Acceptance Criteria:**
1. DASH playback starts in < 2 seconds (first segment loaded)
2. Seek response < 500ms (to any point in recording)
3. Trick-play at 2x, 4x, 8x speeds without decoder reinitialization
4. Quality adaptation responds to network conditions (ABR)
5. All three client platforms have functional replay UI

---

#### Phase 09: Definition of Done

- [ ] helix-record submodule operational with CI/release
- [ ] Dual-path encoding functional (1 NVENC session, stream + record)
- [ ] MKV and fMP4 container output validated
- [ ] Instant replay buffer with user trigger
- [ ] Background sync to all 5 backends functional
- [ ] DASH replay client on all 3 platforms

---

### Phase 10: Monetization & Auth (P2)

**Phase Goal:** Implement complete authentication, authorization, multi-tenant isolation, and billing infrastructure for the HelixPlay platform. Support OAuth2/OIDC, device authorization for TVs, RBAC, and resource quotas.

**Phase Duration:** 5 weeks
**Engineering Team:** 3 FTE (1 auth/security, 1 backend/platform, 1 billing)
**Deps:** P05 (real-time APIs), P09 (recording)

---

#### Task P10-T01: OAuth2/OIDC Integration (Auth0)

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T01 |
| **Title** | OAuth2/OIDC Authentication with Auth0 |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Security Engineer |

**Description:**
Implement OAuth2/OIDC authentication using Auth0 as identity provider. Support Authorization Code flow with PKCE for web/desktop and Device Authorization Grant (RFC 8628) for TV/console clients. Implement JWT token validation, refresh token rotation, and session management.

**Files:**
```
+ helix-auth/go.mod                            # Module: github.com/helixplay/helix-auth
+ helix-auth/pkg/oauth2/oidc.go                # OIDC discovery & validation
+ helix-auth/pkg/oauth2/pkce.go                # PKCE implementation
+ helix-auth/pkg/oauth2/device.go              # RFC 8628 Device Flow
+ helix-auth/pkg/oauth2/token.go               # JWT token management
+ helix-auth/pkg/oauth2/refresh.go             # Refresh token rotation
+ helix-auth/pkg/oauth2/session.go             # Session store (Redis)
+ helix-auth/pkg/middleware/auth.go            # HTTP auth middleware
+ helix-auth/pkg/middleware/rbac.go            # RBAC middleware
+ helix-auth/cmd/token-test/main.go            # Token validation tool
+ helix-auth/configs/auth0-tenant.yaml         # Auth0 tenant configuration
```

**Auth Flows:**

| Flow | Use Case | Implementation |
|------|----------|----------------|
| Authorization Code + PKCE | Web, Desktop, Mobile | Standard OIDC flow |
| Device Authorization Grant (RFC 8628) | TV, Console | QR code + polling |
| Client Credentials | Service-to-service | mTLS + JWT |

**Acceptance Criteria:**
1. Authorization Code + PKCE flow completes in < 3 seconds (cold start)
2. Device Authorization Grant shows QR code, user authenticates on phone, TV polls successfully
3. JWT access tokens: RS256 signed, 15-minute expiry, validated without Auth0 call (JWKS cache)
4. Refresh token rotation: single-use, family detection for theft
5. Session revocation: token blacklist in Redis with < 50ms check latency

---

#### Task P10-T02: Device Authorization Grant (RFC 8628) for TV

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T02 |
| **Title** | TV/Console Device Authorization Flow with QR Code |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Client Engineer |

**Description:**
Implement RFC 8628 Device Authorization Grant for TV and console clients. Generate QR code for user to scan with mobile device. Poll for authorization completion. Implement proper timeout and error handling.

**Files:**
```
+ helix-auth/pkg/device/flow.go                # Device flow orchestrator
+ helix-auth/pkg/device/qrcode.go              # QR code generation
+ helix-auth/pkg/device/poll.go                # Polling with backoff
~ helix-flutter/lib/auth/device_flow.dart      # Flutter TV device auth
~ helix-angular/src/app/auth/device/           # Web device auth (fallback)
```

**Flow:**
1. TV client requests device code from `/oauth/device/code`
2. TV displays QR code containing `verification_uri_complete`
3. User scans QR with phone, authenticates via standard OIDC flow
4. TV polls `/oauth/token` at interval until authorized or expired
5. On success, TV receives access + refresh tokens

**Acceptance Criteria:**
1. QR code generation < 100ms on TV hardware (Android TV, Apple TV)
2. Complete flow (scan → authenticate → authorized) < 15 seconds
3. Proper timeout after 15 minutes with user-friendly message
4. Works without keyboard input on TV (pure mobile-driven auth)
5. Error handling for denied access, expired code, network failure

---

#### Task P10-T03: Tenant Isolation Architecture

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T03 |
| **Title** | Multi-Tenant Data Isolation (helix-tenant submodule) |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Platform Engineer |

**Description:**
Implement strict tenant isolation where each white-label tenant has fully separated users, game catalog, recordings, billing data, and theming. Use row-level security (RLS) in CockroachDB and tenant context propagation.

**Files:**
```
+ helix-tenant/go.mod                          # Module: github.com/helixplay/helix-tenant
+ helix-tenant/pkg/context/tenant.go           # Tenant context propagation
+ helix-tenant/pkg/isolation/db.go             # DB-level tenant isolation
+ helix-tenant/pkg/isolation/rls.go            # CockroachDB RLS policies
+ helix-tenant/pkg/isolation/middleware.go     # Tenant extraction middleware
+ helix-tenant/pkg/isolation/validate.go       # Cross-tenant access validation
+ helix-tenant/pkg/catalog/overlay.go          # Per-tenant catalog overlay
+ helix-tenant/pkg/billing/usage.go            # Per-tenant usage tracking
+ helix-tenant/pkg/models/tenant.go            # Tenant data models
+ helix-tenant/pkg/api/admin.go                # Tenant admin API
+ helix-tenant/configs/default-policies.sql    # Default RLS policies
+ helix-tenant/LICENSE                         # MIT
```

**Isolation Model:**
- **Database:** CockroachDB row-level security (`CREATE POLICY tenant_isolation`)
- **Storage:** Per-tenant S3 bucket prefix (`/recordings/{tenant_id}/...`)
- **Cache:** Redis key prefix (`tenant:{id}:...`)
- **Message Bus:** NATS tenant-scoped subjects (`events.{tenant_id}.session.start`)
- **Compute:** Per-tenant resource quotas via cgroup v2

**Acceptance Criteria:**
1. No cross-tenant data access possible (verified via SQL injection + direct DB access tests)
2. Tenant context automatically propagated via gRPC metadata
3. RLS policies enforce tenant isolation at database level (bypass impossible without superuser)
4. Per-tenant catalog overlay correctly merges global + tenant-specific entries
5. Tenant provisioning: new tenant fully isolated in < 30 seconds

---

#### Task P10-T04: RBAC Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T04 |
| **Title** | Role-Based Access Control with Hierarchical Roles |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Description:**
Implement RBAC with predefined roles and custom role support. Roles are tenant-scoped. Permission checks on every API endpoint.

**Files:**
```
+ helix-auth/pkg/rbac/model.go                 # RBAC data model
+ helix-auth/pkg/rbac/roles.go                 # Predefined roles
+ helix-auth/pkg/rbac/permission.go            # Permission definitions
+ helix-auth/pkg/rbac/enforcer.go              # Casbin-compatible enforcer
+ helix-auth/pkg/rbac/middleware.go            # RBAC HTTP/gRPC middleware
+ helix-auth/pkg/rbac/admin.go                 # Role management API
```

**Default Roles:**

| Role | Permissions | Scope |
|------|------------|-------|
| Platform Admin | Full access | Platform-wide |
| Tenant Admin | Full access within tenant | Tenant |
| Tenant Manager | User management, billing, catalog | Tenant |
| End User | Play games, manage recordings | Own data only |
| Guest | View catalog, no play | N/A |

**Acceptance Criteria:**
1. All API endpoints enforce RBAC checks (verified via automated penetration test)
2. Role assignment atomic and consistent across distributed nodes
3. Permission changes propagate in < 5 seconds
4. Custom role creation with granular permissions (UI + API)
5. Audit log entry for every permission check failure

---

#### Task P10-T05: Billing Engine & Resource Quotas

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T05 |
| **Title** | Usage-Based Billing with Resource Quotas |
| **Priority** | P2 |
| **Effort** | 1.5pw |
| **Owner** | Platform Engineer |

**Description:**
Implement usage-based billing tracking streaming minutes, storage, and concurrent sessions. Enforce resource quotas at tenant and user level. Integrate with Stripe for payment processing.

**Files:**
```
+ helix-tenant/pkg/billing/engine.go           # Billing event processor
+ helix-tenant/pkg/billing/stripe.go           # Stripe integration
+ helix-tenant/pkg/billing/quota.go            # Quota management
+ helix-tenant/pkg/billing/usage.go            # Usage aggregation
+ helix-tenant/pkg/billing/invoice.go          # Invoice generation
+ helix-tenant/pkg/billing/webhook.go          # Stripe webhook handler
+ helix-tenant/pkg/quota/enforcer.go           # Quota enforcement
+ helix-tenant/pkg/quota/limit.go              # Limit definitions
```

**Metered Resources:**

| Resource | Unit | Quota Levels |
|----------|------|-------------|
| Streaming time | Minutes | Per-user daily, per-tenant monthly |
| Concurrent sessions | Count | Per-tenant hard limit |
| Recording storage | GB | Per-user, per-tenant |
| Bandwidth | GB | Per-tenant monthly |
| API requests | Count | Per-tenant rate limit |

**Acceptance Criteria:**
1. Billing events emitted for every minute of streaming (NATS → aggregation)
2. Quota enforcement prevents usage beyond limit with user-friendly message
3. Stripe integration: subscription creation, metered billing, invoice generation
4. Usage dashboard shows real-time consumption ( < 5 second delay)
5. Grace period: 10% over-quota allowed before hard cut-off

---

#### Task P10-T06: helix-vault Submodule (Secrets Management)

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T06 |
| **Title** | Secrets Management with KEK Hierarchy |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Files:**
```
+ helix-vault/go.mod                           # Module: github.com/helixplay/helix-vault
+ helix-vault/pkg/kek/hierarchy.go             # Key Encryption Key hierarchy
+ helix-vault/pkg/kek/rotation.go              # Automatic KEK rotation
+ helix-vault/pkg/secrets/store.go             # Secret storage interface
+ helix-vault/pkg/secrets/hashicorp.go         # HashiCorp Vault backend
+ helix-vault/pkg/secrets/aws.go               # AWS Secrets Manager backend
+ helix-vault/pkg/secrets/file.go              # Development file backend
+ helix-vault/pkg/api/secrets.go               # Secret management API
+ helix-vault/LICENSE                          # MIT
```

**KEK Hierarchy:**
- **Platform KEK (L0):** HSM-protected, never leaves HSM
- **Tenant KEK (L1):** Encrypted by L0, per-tenant
- **Session KEK (L2):** Ephemeral, encrypted by L1, auto-rotated

**Acceptance Criteria:**
1. Secret retrieval latency < 10ms (cached) / < 100ms (cold)
2. KEK rotation: automatic every 90 days, zero-downtime
3. Integration with HashiCorp Vault and AWS Secrets Manager
4. No plaintext secrets in config files or environment variables
5. Audit log for every secret access (who, when, what)

---

#### Phase 10: Definition of Done

- [ ] OAuth2/OIDC authentication functional (all flows)
- [ ] Device Authorization Grant works on TV with QR code
- [ ] Tenant isolation enforced at database, storage, cache, and message bus levels
- [ ] RBAC checks on every API endpoint
- [ ] Billing engine tracking usage with Stripe integration
- [ ] Secret management with KEK hierarchy and auto-rotation

---

### Phase 11: Hardening & Security (P2)

**Phase Goal:** Achieve R-18 Operational Integrity rating through comprehensive security hardening, mTLS everywhere, audit logging, key rotation, continuous security scanning, and container isolation verification.

**Phase Duration:** 5 weeks
**Engineering Team:** 3 FTE (2 security engineers, 1 DevSecOps)
**Deps:** P10 (auth/tenant infrastructure)

---

#### Task P11-T01: R-18 Operational Integrity Enforcement

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T01 |
| **Title** | R-18 Operational Integrity Framework Implementation |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Security Lead |

**Description:**
Implement the R-18 Operational Integrity framework as the foundation of security posture. This includes 18 security controls across 6 domains: Identity, Infrastructure, Data, Application, Operations, and Compliance.

**Files:**
```
+ helix-r18-safeexec/go.mod                    # Module: github.com/helixplay/helix-r18-safeexec
+ helix-r18-safeexec/pkg/constitution/controls.go     # 18 control definitions
+ helix-r18-safeexec/pkg/constitution/checks.go       # Automated compliance checks
+ helix-r18-safeexec/pkg/constitution/report.go       # Compliance reporting
+ helix-r18-safeexec/pkg/safeexec/sandbox.go          # Process sandboxing
+ helix-r18-safeexec/pkg/safeexec/seccomp.go          # seccomp-bpf profiles
+ helix-r18-safeexec/pkg/safeexec/capabilities.go     # Linux capabilities management
+ helix-r18-safeexec/pkg/safeexec/namespaces.go       # Namespace isolation
+ helix-r18-safeexec/configs/seccomp-gaming.json      # Gaming process seccomp profile
+ helix-r18-safeexec/configs/r18-baseline.yaml        # R-18 baseline configuration
+ helix-r18-safeexec/cmd/r18-audit/main.go            # R-18 compliance auditor
+ helix-r18-safeexec/LICENSE                          # MIT (root dependency tree)
```

**R-18 Controls (6 Domains × 3 Controls):**

| Domain | Controls |
|--------|----------|
| **Identity** | (1) Strong MFA enforcement, (2) Just-in-time access, (3) Service identity (SPIFFE) |
| **Infrastructure** | (4) Hardened base images, (5) Network micro-segmentation, (6) Immutable infrastructure |
| **Data** | (7) Encryption at rest (AES-256-GCM), (8) Encryption in transit (TLS 1.3), (9) Key lifecycle management |
| **Application** | (10) SAST/DAST in CI, (11) Dependency vulnerability scanning, (12) Secure code review |
| **Operations** | (13) Audit logging (append-only), (14) Anomaly detection, (15) Incident response automation |
| **Compliance** | (16) Automated compliance checks, (17) Evidence collection, (18) Penetration testing |

**Acceptance Criteria:**
1. All 18 controls implemented and continuously monitored
2. Automated R-18 compliance check passes in CI
3. Quarterly penetration test findings remediated within SLA
4. Security scorecard published monthly to stakeholders
5. `helix-r18-safeexec` is root of dependency tree (all modules depend on it)

---

#### Task P11-T02: mTLS Topology Completion

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T02 |
| **Title** | mTLS Everywhere - Service Mesh Integration |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Security Engineer |

**Description:**
Complete mTLS deployment across all service-to-service communication. Use SPIFFE/SPIRE for workload identity. Implement automatic certificate rotation and revocation.

**Files:**
```
+ helix-vault/pkg/mtls/config.go               # mTLS configuration
+ helix-vault/pkg/mtls/spiffe.go               # SPIFFE ID management
+ helix-vault/pkg/mtls/rotation.go             # Certificate auto-rotation
+ helix-vault/pkg/mtls/revocation.go           # Certificate revocation list
~ infra/k8s/servicemesh/istio-mtls.yaml        # Istio mTLS peer authentication
~ infra/terraform/spire-server.tf              # SPIRE server deployment
```

**mTLS Coverage Matrix:**

| Connection | mTLS | Identity | Rotation |
|------------|------|----------|----------|
| Client → Edge | TLS 1.3 + Client Cert | JWT | 15min (session) |
| Edge → API Gateway | mTLS | SPIFFE | 24hr |
| API → Session Service | mTLS | SPIFFE | 24hr |
| Session → Host Agent | mTLS + WireGuard | SPIFFE + Session Key | Per-session |
| Host → Storage | mTLS | SPIFFE | 24hr |
| Service → CockroachDB | mTLS | Client Cert | 30 days |
| Service → Redis | TLS + AUTH | Password | 90 days |
| Service → NATS | mTLS | SPIFFE | 24hr |

**Acceptance Criteria:**
1. 100% of inter-service traffic encrypted with mTLS (verified via network capture)
2. No plaintext service communication on any port
3. Certificate rotation: zero-downtime, automatic
4. SPIFFE ID validation on every inbound connection
5. Revoked certificates rejected within 5 minutes of revocation

---

#### Task P11-T03: Audit Logging System

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T03 |
| **Title** | Append-Only Audit Log with Tamper Evidence |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Description:**
Implement centralized audit logging for all security-relevant events. Use append-only storage with cryptographic chain-of-custody (hash chain). Forward to SIEM.

**Files:**
```
+ helix-audit/go.mod                           # Module: github.com/helixplay/helix-audit
+ helix-audit/pkg/log/appender.go              # Append-only log writer
+ helix-audit/pkg/log/chain.go                 # Hash chain tamper evidence
+ helix-audit/pkg/log/events.go                # Event type definitions
+ helix-audit/pkg/log/siem.go                  # SIEM forwarder (Splunk/Sentinel)
+ helix-audit/pkg/api/query.go                 # Audit log query API (read-only)
+ helix-audit/configs/audit-events.yaml        # Event classification
```

**Logged Events:**
- Authentication (success/failure)
- Authorization denials
- Session start/stop
- Admin actions (role changes, config changes)
- Key access and rotation
- Data access (recordings, personal data)
- API rate limit violations

**Acceptance Criteria:**
1. Every security event logged within 50ms of occurrence
2. Log entries cryptographically chained (SHA-256 of previous entry)
3. Tamper detection: any modification breaks chain verification
4. SIEM forwarder delivers events with < 5 second latency
5. Log retention: 7 years (compliance), queryable for 90 days hot storage

---

#### Task P11-T04: Security Scanning Pipeline

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T04 |
| **Title** | Continuous Security Scanning (6-Scanner Pipeline) |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | DevSecOps Engineer |

**Description:**
Integrate 6 security scanners into CI/CD pipeline. Block deployment on critical findings. Track security debt.

**Files:**
```
+ .github/workflows/security-scan.yml          # Unified security scan workflow
+ .github/scripts/sonarqube-scan.sh            # SonarQube integration
+ .github/scripts/snyk-scan.sh                 # Snyk dependency scan
+ .github/scripts/semgrep-scan.sh              # Semgrep SAST rules
+ .github/scripts/trivy-scan.sh                # Trivy container scan
+ .github/scripts/gitleaks-scan.sh             # Secret detection
+ .github/scripts/govulncheck.sh               # Go vulnerability check
+ security/sonar-project.properties            # SonarQube configuration
+ security/.semgrep/helix-rules.yaml           # Custom Semgrep rules
+ security/.trivyignore                        # Trivy exception list
```

**Scanner Matrix:**

| Scanner | Type | Trigger | Blocking |
|---------|------|---------|----------|
| SonarQube | SAST + Code Quality | Every PR | High severity |
| Snyk | Dependency vulns | Every PR + daily | Critical CVSS |
| Semgrep | SAST (custom rules) | Every PR | Critical |
| Trivy | Container + OS vulns | Every build | Critical |
| gitleaks | Secret detection | Every PR + pre-commit | Always (secrets) |
| govulncheck | Go-specific vulns | Every PR | Critical |

**Acceptance Criteria:**
1. All 6 scanners run on every PR with results in < 10 minutes
2. Zero critical findings in default branch (enforced by branch protection)
3. Security scan results published to PR as comment
4. Vulnerability SLA: Critical < 24hr, High < 7 days, Medium < 30 days
5. Historical security debt tracked in dashboard

---

#### Task P11-T05: Container Isolation Verification

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T05 |
| **Title** | Container & Sandbox Isolation Verification |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Description:**
Verify container isolation for host agent processes. Implement seccomp-bpf, AppArmor/SELinux profiles, and user namespace isolation. Validate escape resistance.

**Files:**
```
+ helix-r18-safeexec/configs/seccomp-gaming.json    # Game process seccomp
+ helix-r18-safeexec/configs/apparmor-gaming.profile # AppArmor profile
+ helix-r18-safeexec/configs/selinux-gaming.te       # SELinux type enforcement
+ helix-r18-safeexec/pkg/isolation/container.go      # Container runtime wrapper
+ helix-r18-safeexec/pkg/isolation/verify.go         # Isolation verification tests
+ helix-r18-safeexec/cmd/isolation-test/main.go      # Isolation test runner
```

**Acceptance Criteria:**
1. Game process cannot escape container (verified via `docker escape` test suite)
2. seccomp-bpf blocks all syscalls except explicit allowlist (verified via `strace -f`)
3. No privilege escalation possible from unprivileged container
4. Container resource limits (CPU, memory, I/O) enforced by cgroup v2
5. Game process has no network access except through proxy/tunnel

---

#### Phase 11: Definition of Done

- [ ] R-18 Operational Integrity: all 18 controls implemented
- [ ] mTLS on 100% of service-to-service traffic
- [ ] Audit logging: append-only, tamper-evident, SIEM-integrated
- [ ] KEK rotation: automatic, zero-downtime
- [ ] Security scanning: 6 scanners in CI, zero critical findings
- [ ] Container isolation verified and tested

---

### Phase 12: Beta Launch (P2)

**Phase Goal:** Launch closed beta with canary deployment capability, 30-day replay retention, comprehensive load testing, integration testing, and complete documentation.

**Phase Duration:** 4 weeks
**Engineering Team:** 4 FTE (1 SRE, 1 QA lead, 1 technical writer, 1 product)
**Deps:** P07-P11 (all previous phases)

---

#### Task P12-T01: Canary Deployment System

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T01 |
| **Title** | Operator-Facing Canary Deployment with Automatic Rollback |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | SRE |

**Description:**
Implement canary deployment for host agent fleet. Operator controls traffic split percentage, health criteria, and automatic rollback thresholds.

**Files:**
```
+ helix-deploy/go.mod                          # Module: github.com/helixplay/helix-deploy
+ helix-deploy/pkg/canary/controller.go        # Canary deployment controller
+ helix-deploy/pkg/canary/metrics.go           # Health metric evaluation
+ helix-deploy/pkg/canary/rollback.go          # Automatic rollback logic
+ helix-deploy/pkg/canary/promote.go           # Full promotion workflow
+ helix-deploy/cmd/helix-deploy/main.go        # CLI operator tool
+ helix-deploy/configs/canary-defaults.yaml    # Default canary parameters
```

**Canary Parameters:**
- Initial traffic split: 5% → 25% → 50% → 100%
- Health criteria: p99 latency < threshold, error rate < 0.1%, GPU utilization stable
- Evaluation window: 10 minutes per step
- Automatic rollback on any health criterion failure

**Acceptance Criteria:**
1. Canary deployment completes in < 15 minutes (all steps)
2. Automatic rollback triggers in < 30 seconds on health check failure
3. Operator can view real-time canary metrics dashboard
4. Canary works for host agent binary, not just container
5. Zero-downtime deployment (rolling update with session drain)

---

#### Task P12-T02: 30-Day Replay Retention

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T02 |
| **Title** | Replay Storage Lifecycle (30-Day Retention) |
| **Priority** | P2 |
| **Effort** | 0.5pw |
| **Owner** | SRE |

**Description:**
Implement automatic lifecycle management for replay recordings: 7 days hot on NVMe, 23 days warm on object storage, automatic deletion after 30 days. Configurable per-tenant.

**Files:**
```
+ helix-record/pkg/lifecycle/policy.go         # Retention policy engine
+ helix-record/pkg/lifecycle/transition.go     # Storage tier transitions
+ helix-record/pkg/lifecycle/cleanup.go        # Automated cleanup
+ infra/terraform/lifecycle-rules.tf           # S3 lifecycle rules
```

**Acceptance Criteria:**
1. Recordings available for instant replay for 7 days (NVMe)
2. Recordings available for download/streaming for 30 days total
3. Automatic deletion after 30 days (no manual intervention)
4. Per-tenant retention override functional
5. Deletion audit log entries generated

---

#### Task P12-T03: Load Testing at Scale

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T03 |
| **Title** | Platform Load Testing (1000 Concurrent Sessions) |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | QA Lead |

**Description:**
Execute comprehensive load test simulating 1000 concurrent sessions. Measure latency distribution, GPU utilization, network throughput, and resource contention.

**Files:**
```
+ tests/load/k6/streaming-load.js              # k6 streaming load script
+ tests/load/k6/session-lifecycle.js           # Session CRUD load test
+ tests/load/terraform/load-infra.tf            # Load test infrastructure
+ tests/load/Makefile                           # Load test orchestration
+ tests/load/reports/template.html              # Load test report template
```

**Load Test Scenarios:**

| Scenario | Sessions | Duration | Metrics |
|----------|----------|----------|---------|
| Steady State | 1000 | 4 hours | p50/p99/p999 latency, GPU%, throughput |
| Ramp Up | 0 → 1000 | 30 min | Scale-up latency, connection success rate |
| Ramp Down | 1000 → 0 | 30 min | Graceful session termination |
| Spike | 0 → 2000 | 5 min | Auto-scaling response, queue depth |
| Burst | 100 → 500 | 1 min | Latency impact under sudden load |
| Long-running | 100 | 24 hours | Memory leaks, thermal throttling, drift |

**Acceptance Criteria:**
1. 1000 concurrent sessions sustained for 4 hours with p999 < 50ms
2. Connection success rate > 99.9% during ramp-up
3. No memory leaks (RSS stable within 5% over 24-hour test)
4. Auto-scaling responds to spike in < 60 seconds
5. GPU thermal throttling < 2% of session time

---

#### Task P12-T04: Final Integration Testing

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T04 |
| **Title** | End-to-End Integration Test Suite |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | QA Lead |

**Description:**
Complete end-to-end integration testing across all client platforms, all codecs, all controller types, and all network conditions.

**Files:**
```
+ tests/e2e/README.md                          # E2E test documentation
+ tests/e2e/desktop/e2e_test.go                # Wails desktop E2E
+ tests/e2e/mobile/e2e_test.go                 # Flutter mobile E2E
+ tests/e2e/web/e2e_test.go                    # Angular web E2E
+ tests/e2e/shared/assertions.go               # Shared test assertions
+ tests/e2e/shared/fixtures.go                 # Test fixtures
```

**E2E Matrix:**

| Client | Codecs | Controllers | Networks |
|--------|--------|-------------|----------|
| Wails (Win/Mac/Linux) | H.264/HEVC/AV1 | DS5/Xbox/Switch | LAN/WAN/4G/5G |
| Flutter (Android/iOS) | H.264/HEVC | DS5/Xbox/MFi | WiFi/5G/4G |
| Flutter TV (Android TV/tvOS) | H.264/HEVC | DS5/Xbox | Ethernet/WiFi |
| Angular+WASM (Chrome/Firefox/Safari) | H.264/HEVC (WebCodecs) | DS5/Xbox | LAN/WiFi |

**Acceptance Criteria:**
1. All E2E tests pass on all platform combinations
2. Video quality MOS > 4.0 (subjective testing panel, n=20)
3. Controller latency < 2ms (measured via oscilloscope)
4. Audio A/V sync < 5ms (measured with test signal)
5. No P1 or P2 bugs remaining in backlog

---

#### Task P12-T05: Documentation Completion

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T05 |
| **Title** | Complete Technical Documentation Suite |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Technical Writer |

**Files:**
```
+ docs/architecture/README.md                  # Architecture overview
+ docs/architecture/streaming.md               # Streaming pipeline
+ docs/architecture/controllers.md             # Controller input pipeline
+ docs/architecture/security.md                # Security architecture
+ docs/operations/deployment.md                # Deployment guide
+ docs/operations/monitoring.md                # Monitoring & alerting
+ docs/operations/incident-response.md         # Incident response
+ docs/api/README.md                           # API reference
+ docs/api/openapi.yaml                        # OpenAPI specification
+ docs/clients/README.md                       # Client integration guide
+ docs/contributing/README.md                  # Contribution guidelines
+ docs/changelog/CHANGELOG.md                  # Version changelog
```

**Acceptance Criteria:**
1. All architecture documents complete with diagrams
2. API documentation generated from OpenAPI spec
3. Deployment guide enables new team member to deploy in < 2 hours
4. All 29 submodules have README with build/test instructions
5. Documentation published to docs site (MkDocs)

---

#### Phase 12: Definition of Done

- [ ] Canary deployment system operational
- [ ] 30-day replay retention automated
- [ ] 1000 concurrent session load test passed
- [ ] E2E tests pass on all platform combinations
- [ ] Complete documentation suite published

---

### Phase 13: GA Release (P1)

**Phase Goal:** Release HelixPlay v1.0.0 with complete tag-publish across all 29 submodules, four-mirror sync verification, final security audit, and production deployment guide.

**Phase Duration:** 3 weeks
**Engineering Team:** 2 FTE (1 release engineer, 1 security auditor)
**Deps:** P12 (beta launch complete)

---

#### Task P13-T01: v1.0.0 Release Train

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T01 |
| **Title** | Coordinated v1.0.0 Release Across All Submodules |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Release Engineer |

**Description:**
Execute coordinated release of v1.0.0 across all 29 submodules with correct semantic versioning, dependency alignment, and release notes.

**Files:**
```
+ scripts/release/train.sh                     # Release train orchestrator
+ scripts/release/validate.sh                  # Pre-release validation
+ scripts/release/tag.sh                       # Cross-repo tagging
+ scripts/release/notes.sh                     # Release notes generator
+ .github/workflows/release-train.yml          # Release train CI
```

**Release Train Process:**
1. **Validation:** All tests pass, zero critical security findings
2. **Version Alignment:** Update all `go.mod` to v1.0.0, resolve dependencies
3. **Tagging:** Create signed Git tags (`git tag -s v1.0.0`) on all repos
4. **Build:** Build all artifacts (binaries, containers, WASM modules)
5. **Publish:** Push to artifact registries (GitHub Packages, Docker Hub, npm)
6. **Verify:** Confirm all artifacts downloadable and checksums match

**Acceptance Criteria:**
1. All 29 submodules tagged v1.0.0 within 1 hour window
2. All cross-module dependencies resolve to v1.0.0
3. Container images published with signed attestations (cosign)
4. Release notes generated from conventional commits (automated)
5. No post-release hotfixes required in first 48 hours

---

#### Task P13-T02: Four-Mirror Git Sync Verification

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T02 |
| **Title** | Four-Mirror Git Topology Sync Verification |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Release Engineer |

**Description:**
Verify synchronized state across four Git mirrors (GitHub primary, GitLab, Gitea, Bitbucket). All repos, tags, releases, and issues synced.

**Mirror Topology:**
- **Primary:** GitHub (github.com/helixplay/*)
- **Mirror 1:** GitLab (gitlab.helixplay.io/helixplay/*)
- **Mirror 2:** Gitea (gitea.helixplay.io/helixplay/*)
- **Mirror 3:** Bitbucket (bitbucket.org/helixplay/*)

**Files:**
```
+ scripts/mirror/sync.sh                       # Mirror sync orchestrator
+ scripts/mirror/verify.sh                     # Mirror verification
+ .github/workflows/mirror-sync.yml            # Automated mirror sync
```

**Acceptance Criteria:**
1. All 29 repos present on all 4 mirrors with identical HEAD
2. All v1.0.0 tags present and matching SHA on all mirrors
3. Mirror sync latency < 5 minutes from primary push
4. Issue/PR metadata synced to GitLab (for local tracking)
5. Automated recovery if mirror falls out of sync

---

#### Task P13-T03: Final Security Audit

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T03 |
| **Title** | Independent Security Audit & Penetration Test |
| **Priority** | P1 |
| **Effort** | 1pw (external) |
| **Owner** | Security Lead |

**Description:**
Commission independent security audit covering: code review, penetration testing, architecture review, and compliance assessment.

**Audit Scope:**
- SAST: All 29 submodules via SonarQube + Semgrep
- DAST: Live beta environment via OWASP ZAP
- Penetration test: External firm, black + grey box
- Architecture review: Threat model validation
- Compliance: SOC 2 Type II readiness assessment

**Acceptance Criteria:**
1. Zero critical vulnerabilities in audit report
2. All high findings have remediation plan with timeline
3. Penetration test: no remote code execution, no privilege escalation
4. Threat model validated: all identified threats have mitigations
5. SOC 2 Type II readiness: no gaps in controls

---

#### Task P13-T04: Production Deployment Guide

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T04 |
| **Title** | Production Deployment Guide & Runbooks |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | SRE |

**Files:**
```
+ docs/production/README.md                    # Production overview
+ docs/production/prerequisites.md             # Infrastructure prerequisites
+ docs/production/deployment.md                # Step-by-step deployment
+ docs/production/configuration.md             # Configuration reference
+ docs/production/monitoring.md                # Monitoring setup
+ docs/production/runbooks/                    # Incident runbooks
+ docs/production/runbooks/latency-spike.md    # Latency spike response
+ docs/production/runbooks/gpu-failure.md      # GPU node failure
+ docs/production/runbooks/network-partition.md # Network partition
+ docs/production/runbooks/security-incident.md # Security incident
```

**Acceptance Criteria:**
1. Deployment guide enables production deployment from scratch in < 4 hours
2. All runbooks have decision trees and command snippets
3. Monitoring dashboards created and documented
4. On-call rotation documented with escalation procedures
5. Disaster recovery plan: RPO < 5 minutes, RTO < 30 minutes

---

#### Phase 13: Definition of Done

- [ ] v1.0.0 tagged and published across all 29 submodules
- [ ] Four-mirror sync verified
- [ ] Independent security audit passed (zero critical findings)
- [ ] Production deployment guide complete
- [ ] Platform live in production

---

## Part B: Architecture Deep-Dive Implementation Tasks

### B1: Streaming Pipeline Implementation

---

#### B1-T01: WebRTC (Pion v4) Integration

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T01 |
| **Title** | WebRTC Pion v4 Full Integration |
| **Effort** | 3pw |

**Description:**
Implement WebRTC transport using Pion v4 (`github.com/pion/webrtc/v4`) for peer connection management, ICE, DTLS, SRTP, and SCTP data channels. Customize for low-latency game streaming.

**Files:**
```
+ helix-stream/pkg/webrtc/pc.go                # Peer connection factory
+ helix-stream/pkg/webrtc/config.go            # WebRTC configuration
+ helix-stream/pkg/webrtc/ice.go               # ICE server & candidate management
+ helix-stream/pkg/webrtc/negotiation.go       # SDP offer/answer
+ helix-stream/pkg/webrtc/track.go             # Media track management
+ helix-stream/pkg/webrtc/datachannel.go       # Input/control data channels
+ helix-stream/pkg/webrtc/stats.go             # RTC stats collection
+ helix-stream/pkg/webrtc/bwe.go               # Bandwidth estimation
```

**Pion v4 Configuration:**
```go
var DefaultConfig = webrtc.Configuration{
    ICETransportPolicy: webrtc.ICETransportPolicyAll,
    BundlePolicy:       webrtc.BundlePolicyMaxBundle,
    RTCPFeedback: []webrtc.RTCPFeedback{
        {Type: "goog-remb"},
        {Type: "transport-cc"},
        {Type: webrtc.TypeRTCPFBNACK},
        {Type: webrtc.TypeRTCPFBNACK},
        {Parameter: "pli", Type: webrtc.TypeRTCPFBGoogREMB},
    },
    SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
}

var SettingEngine = webrtc.SettingEngine{
    // Disable ICE lite (full ICE for host connectivity)
    // Enable DTLS 1.3
    // Set MTU discovery
    // Enable TWCC (Transport Wide Congestion Control)
}
```

**Key Implementation Details:**
1. **ICE:** Custom STUN/TURN server deployment with regional affinity. TURN for symmetric NAT fallback.
2. **DTLS:** 1.3 with cipher suites `TLS_AES_128_GCM_SHA256`, `TLS_AES_256_GCM_SHA384`
3. **SRTP:** AES-GCM preferred over AES-CM (reduced CPU, better security)
4. **SCTP:** Data channel for controller input with unordered, unreliable delivery
5. **Bandwidth Estimation:** TWCC + custom HelixPlay SQP (Streaming Quality Predictor)

**Acceptance Criteria:**
1. Peer connection establishment < 500ms (LAN), < 2s (WAN with TURN)
2. ICE candidate gathering < 200ms with regional STUN
3. Data channel latency for controller input < 1ms (localhost)
4. SRTP throughput > 100Mbps sustained (single peer connection)
5. Graceful handling of all NAT types (full cone → symmetric)

---

#### B1-T02: Codec Ladder Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T02 |
| **Title** | Codec Negotiation Ladder (H.264 → HEVC → AV1) |
| **Effort** | 2pw |

**Description:**
Implement codec negotiation ladder with automatic fallback. Client advertises supported codecs via SDP; server selects optimal codec based on client capability, network conditions, and GPU encoder availability.

**Codec Ladder:**

| Priority | Codec | Profile | Use Case | Fallback Trigger |
|----------|-------|---------|----------|-----------------|
| 1 | AV1 | Main 10 | Best quality, lowest bitrate | Client doesn't support; GPU can't encode |
| 2 | HEVC | Main 10 | Good quality, hardware encode everywhere | Client doesn't support; patent licensing |
| 3 | H.264 | High | Universal compatibility | N/A (baseline) |

**Files:**
```
+ helix-stream/pkg/codec/ladder.go             # Codec ladder negotiation
+ helix-stream/pkg/codec/capability.go         # Client codec capability
+ helix-stream/pkg/codec/fallback.go           # Fallback logic
+ helix-stream/pkg/codec/policy.go             # Codec selection policy
```

**Negotiation Flow:**
1. Client SDP includes `a=rtpmap` for supported codecs (ordered by preference)
2. Server checks GPU encoder availability for each codec
3. Server selects highest-priority mutually supported codec
4. If selected codec fails at runtime (encoder error), fallback to next
5. Codec switch mid-stream: seamless (IDR frame, no reconnection)

**Acceptance Criteria:**
1. AV1 selected when client supports and GPU has AV1 encoder (RTX 40+, Intel Arc)
2. HEVC selected on Apple devices (hardware decode support)
3. H.264 fallback works on all platforms
4. Codec switch mid-stream completes in < 200ms
5. Codec selection logged for analytics

---

#### B1-T03: ABR/FEC/SQP Policies

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T03 |
| **Title** | Adaptive Bitrate, Forward Error Correction, Quality Policies |
| **Effort** | 2.5pw |

**Description:**
Implement Adaptive Bitrate (ABR), Forward Error Correction (FEC), and Streaming Quality Predictor (SQP) for resilient streaming under variable network conditions.

**Files:**
```
+ helix-stream/pkg/abr/controller.go           # ABR controller
+ helix-stream/pkg/abr/ladder.go               # Bitrate ladder definitions
+ helix-stream/pkg/fec/encoder.go              # FEC encoder (Reed-Solomon)
+ helix-stream/pkg/fec/decoder.go              # FEC decoder
+ helix-stream/pkg/sqp/predictor.go            # Streaming Quality Predictor
+ helix-stream/pkg/sqp/score.go                # Quality score calculation
+ helix-stream/pkg/congestion/detector.go      # Congestion detection
+ helix-stream/pkg/congestion/controller.go    # Congestion response
```

**ABR Ladder:**

| Resolution | Target Bitrate | Max Bitrate | Min Bitrate |
|------------|---------------|-------------|-------------|
| 4K (3840×2160) | 40 Mbps | 80 Mbps | 20 Mbps |
| 1440p (2560×1440) | 25 Mbps | 50 Mbps | 12 Mbps |
| 1080p (1920×1080) | 12 Mbps | 24 Mbps | 6 Mbps |
| 720p (1280×720) | 6 Mbps | 12 Mbps | 3 Mbps |
| 540p (960×540) | 3 Mbps | 6 Mbps | 1.5 Mbps |

**FEC Strategy:**
- **Proactive FEC:** 5-10% overhead during stable conditions
- **Reactive FEC:** Increase to 20-30% on packet loss detection
- **Unequal Error Protection:** I-frames get higher FEC protection than P-frames
- **Algorithm:** Reed-Solomon over GF(256), `fec(20, 16)` as baseline

**SQP (Streaming Quality Predictor):**
```go
type SQPScore struct {
    BandwidthEstimate  float64   // bps
    LossRate           float64   // 0-1
    Jitter             float64   // ms
    Rtt                float64   // ms
    GpuUtilization     float64   // 0-1
    ThermalThrottling  bool
    QualityScore       float64   // 0-100 composite
}
```

**Acceptance Criteria:**
1. ABR adapts to bandwidth changes within 2 seconds
2. FEC recovers from 5% packet loss without visible artifacts
3. SQP score accurately predicts quality degradation (correlation > 0.9)
4. Congestion detected and responded to within 500ms
5. Overall quality MOS > 4.0 at 2% packet loss

---

#### B1-T04: Transport Abstraction Layer

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T04 |
| **Title** | Transport Abstraction (WebRTC / QUIC / Custom UDP) |
| **Effort** | 2pw |

**Description:**
Implement transport abstraction layer that supports WebRTC (primary), QUIC (fallback for corporate firewalls), and custom UDP (LAN optimization). Automatic transport selection based on network conditions.

**Files:**
```
+ helix-stream/pkg/transport/interface.go      # Transport interface
+ helix-stream/pkg/transport/webrtc.go         # WebRTC transport
+ helix-stream/pkg/transport/quic.go           # QUIC transport (quic-go)
+ helix-stream/pkg/transport/udp.go            # Custom UDP transport
+ helix-stream/pkg/transport/selector.go       # Transport auto-selection
+ helix-stream/pkg/transport/fallback.go       # Transport fallback logic
```

**Transport Selection Matrix:**

| Condition | Primary | Fallback |
|-----------|---------|----------|
| Direct UDP possible | WebRTC (UDP) | QUIC |
| UDP blocked, TCP open | QUIC | WebRTC (TCP) |
| Corporate proxy | WebRTC (TCP/TURN) | N/A |
| LAN (sub-5ms latency) | Custom UDP | WebRTC |

**Acceptance Criteria:**
1. Transport selection completes within connection establishment time
2. Fallback to alternative transport on primary failure < 3 seconds
3. QUIC transport functional with equivalent latency to WebRTC
4. Custom UDP transport for LAN: latency reduced by > 20% vs. WebRTC
5. All transports support same feature set (FEC, ABR, encryption)

---

### B2: Controller Input Pipeline Implementation

---

#### B2-T01: 1kHz USB Polling Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B2-T01 |
| **Title** | 1kHz USB HID Polling for DualSense |
| **Effort** | 1.5pw |

**Description:**
Implement 1kHz (1ms interval) USB HID polling for DualSense controller on all platforms. Use raw HID access to achieve polling rates beyond standard OS driver defaults (125-250Hz).

**Files:**
```
+ helix-input/go.mod                           # Module: github.com/helixplay/helix-input
+ helix-input/pkg/hid/poll.go                  # HID polling loop
+ helix-input/pkg/hid/dualsense/usb.go         # DualSense USB protocol
+ helix-input/pkg/hid/dualsense/bluetooth.go   # DualSense Bluetooth protocol
+ helix-input/pkg/hid/dualsense/haptics.go     # Haptic feedback
+ helix-input/pkg/hid/dualsense/trigger.go     # Adaptive trigger control
+ helix-input/pkg/hid/dualsense/imu.go         # Gyroscope + accelerometer
+ helix-input/pkg/hid/dualsense/lightbar.go    # LED/lightbar control
+ helix-input/pkg/hid/xbox/core.go             # Xbox controller support
+ helix-input/pkg/hid/switch/core.go           # Nintendo Switch Pro support
```

**USB Polling Architecture:**
```go
type Poller struct {
    device      *hid.Device
    interval    time.Duration     // 1ms for 1kHz
    buffer      []byte            // 64-byte HID report
    callbacks   []InputCallback   // Registered callbacks
    ring        *SPSCRing[InputReport] // Lock-free ring to encode thread
}

func (p *Poller) Start() {
    // SCHED_FIFO thread at priority 93
    // Busy-wait with sched_yield for sub-microsecond precision
    // Or use hid_read_timeout with 1ms + io_uring for async
}
```

**Acceptance Criteria:**
1. USB polling interval: 1ms ± 50 microseconds (measured via USB analyzer)
2. No dropped input frames at 1kHz for 1-hour test
3. Works on Windows (WinUSB), macOS (IOHIDManager), Linux (hidraw)
4. Bluetooth fallback: 2ms interval (500Hz) acceptable
5. CPU overhead of polling thread < 2% of one core

---

#### B2-T02: Lock-Free SPSC Input Ring Buffer

| Attribute | Detail |
|-----------|--------|
| **ID** | B02-T02 |
| **Title** | Lock-Free Input Report Ring Buffer |
| **Effort** | 1pw |

**Description:**
Implement lock-free SPSC ring buffer for controller input reports from HID polling thread to network serialization thread. Zero-copy where possible.

**Files:**
```
+ helix-input/pkg/ring/input.go                # Input-specific SPSC ring
+ helix-input/pkg/serialize/packet.go          # Input packet serializer
+ helix-input/pkg/protocol/binary.go           # 16-32 byte binary protocol
```

**Binary Protocol:**
```
Offset  Size  Field
0       1     Packet type (0x01 = input report)
1       1     Sequence number (mod 256)
2       8     Timestamp (microseconds, monotonic)
10      2     Left stick X (0-65535)
12      2     Left stick Y (0-65535)
14      2     Right stick X (0-65535)
16      2     Right stick Y (0-65535)
18      2     Button bitmask (16 buttons)
20      1     Left trigger (0-255)
21      1     Right trigger (0-255)
22      1     D-pad state
23      1     Touchpad fingers
24      4     Gyro X (optional, extended packet)
28      4     Gyro Y (optional, extended packet)
```

**Acceptance Criteria:**
1. Ring buffer latency: poll → serialize < 50 microseconds
2. 1kHz input sustained without drops (ring capacity: 1024 entries)
3. Binary packet size: 16 bytes (standard), 32 bytes (extended with IMU)
4. Packet loss detection via sequence number gaps

---

#### B2-T03: DualSense Haptics & Adaptive Triggers

| Attribute | Detail |
|-----------|--------|
| **ID** | B2-T03 |
| **Title** | Full DualSense Feature Fidelity |
| **Effort** | 1.5pw |

**Description:**
Implement complete DualSense feature support: haptic feedback (L5/R5 actuators), adaptive triggers (L2/R2 resistance), gyroscope (6-axis), accelerometer, touchpad, and lightbar.

**Files:**
```
+ helix-input/pkg/hid/dualsense/haptics.go     # Haptic motor control
+ helix-input/pkg/hid/dualsense/trigger.go     # Adaptive trigger profiles
+ helix-input/pkg/hid/dualsense/imu.go         # Gyro + accelerometer
+ helix-input/pkg/hid/dualsense/touchpad.go    # Touchpad input
+ helix-input/pkg/hid/dualsense/lightbar.go    # LED control
+ helix-input/pkg/hid/dualsense/audio.go       # Controller speaker/headset
+ helix-input/pkg/hid/dualsense/mic.go         # Microphone
```

**Adaptive Trigger Profiles:**

| Profile | Description | Use Case |
|---------|-------------|----------|
| Off | No resistance | Default |
| Rigid | Full resistance | Heavy weapon |
| Vibration | Pulsing resistance | Machine gun |
| Slope | Increasing resistance | Bow draw |
| Feedback | Position-based feedback | Accelerator |

**Acceptance Criteria:**
1. Haptic feedback: L5/R5 independent control, < 5ms host→controller latency
2. Adaptive triggers: 255 resistance levels, profile switching < 10ms
3. Gyroscope: 2000 dps range, < 2ms report latency
4. Accelerometer: ±4g range, synchronized with gyro
5. All features functional over both USB and Bluetooth
6. Feature availability advertised to host via capability bits

---

#### B2-T04: Controller Hot-Plug & Renegotiation

| Attribute | Detail |
|-----------|--------|
| **ID** | B02-T04 |
| **Title** | Controller Hot-Plug & Mid-Session Renegotiation |
| **Effort** | 1pw |

**Description:**
Support controller connection/disconnection during active streaming session. Automatic capability renegotiation when controller changes (e.g., Xbox → DualSense swap).

**Files:**
```
+ helix-input/pkg/hotplug/monitor.go           # Hot-plug event monitor
+ helix-input/pkg/hotplug/renogotiate.go       # Mid-session renegotiation
+ helix-input/pkg/hotplug/manager.go           # Controller manager
```

**Acceptance Criteria:**
1. Controller connect: detected and functional within 1 second
2. Controller disconnect: session continues, input paused gracefully
3. Controller swap (type change): capability renegotiation in < 2 seconds
4. Multiple controllers: up to 4 simultaneous, player assignment
5. No session disruption during hot-plug events

---

### B3: Capture & Encode Pipeline Implementation

---

#### B3-T01: Per-OS Capture Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T01 |
| **Title** | Platform-Specific Screen Capture |
| **Effort** | 3pw |

**Description:**
Implement hardware-accelerated screen capture for each host OS: DXGI Desktop Duplication API (Windows), ScreenCaptureKit (macOS), KMS/DRM + PipeWire (Linux).

**Files:**
```
+ helix-capture/go.mod                         # Module: github.com/helixplay/helix-capture
+ helix-capture/pkg/capture/interface.go       # Capture interface
+ helix-capture/pkg/capture/dxgi/dda.go        # DXGI DDA (Windows)
+ helix-capture/pkg/capture/dxgi/texture.go    # DirectX texture management
+ helix-capture/pkg/capture/dxgi/mapper.go     # GPU texture mapper
+ helix-capture/pkg/capture/screencapturekit/  # macOS ScreenCaptureKit
+ helix-capture/pkg/capture/kms/kms.go         # Linux KMS/DRM
+ helix-capture/pkg/capture/pipewire/pw.go     # Linux PipeWire
+ helix-capture/pkg/capture/vulkan/vulkan.go   # Vulkan capture (cross-platform)
```

**Windows (DXGI DDA):**
- `IDXGIOutputDuplication::AcquireNextFrame()` for frame capture
- `ID3D11Device` texture sharing with encoder
- Hardware cursor compositing overlay
- HDR metadata extraction from `DXGI_OUTPUT_DESC`
- Latency target: < 1ms capture-to-texture

**macOS (ScreenCaptureKit):**
- `SCStream` with `SCContentFilter` for display capture
- `IOSurface` texture sharing
- ProRes/HEVC hardware encode via VideoToolbox
- Latency target: < 2ms capture-to-surface

**Linux (KMS + PipeWire):**
- DRM dumb buffer or `DRM_FORMAT_MOD_LINEAR` for GPU buffers
- PipeWire for Wayland compositor capture
- DMA-BUF fd passing for zero-copy
- Latency target: < 1ms capture-to-buffer

**Acceptance Criteria:**
1. Capture latency per platform within targets (see above)
2. 4K60 capture sustained without frame drops
3. HDR metadata correctly extracted and forwarded to encoder
4. Cursor capture: hardware cursor composited correctly
5. Multi-display: capture from selected display

---

#### B3-T02: Hardware Encoder Factory

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T02 |
| **Title** | Hardware Encoder Factory (NVENC / QSV / AMF / VideoToolbox / VAAPI) |
| **Effort** | 3pw |

**Description:**
Implement hardware encoder factory that auto-detects available encoders and selects optimal encoder based on codec, quality, and latency requirements.

**Files:**
```
+ helix-encode/go.mod                          # Module: github.com/helixplay/helix-encode
+ helix-encode/pkg/encoder/factory.go          # Encoder factory
+ helix-encode/pkg/encoder/interface.go        # Encoder interface
+ helix-encode/pkg/nvenc/nvenc.go              # NVIDIA NVENC
+ helix-encode/pkg/nvenc/session.go            # NVENC session management
+ helix-encode/pkg/nvenc/preset.go             # NVENC preset definitions
+ helix-encode/pkg/qsv/qsv.go                  # Intel QSV
+ helix-encode/pkg/amf/amf.go                  # AMD AMF
+ helix-encode/pkg/videotoolbox/vt.go          # Apple VideoToolbox
+ helix-encode/pkg/vaapi/vaapi.go              # Linux VAAPI
+ helix-encode/pkg/sw/fallback.go              # Software fallback (SVT-AV1, x265)
```

**Encoder Selection Matrix:**

| GPU | H.264 | HEVC | AV1 | Preferred |
|-----|-------|------|-----|-----------|
| NVIDIA RTX 40xx | NVENC | NVENC | NVENC | AV1 |
| NVIDIA RTX 30xx | NVENC | NVENC | N/A | HEVC |
| Intel Arc | QSV | QSV | QSV | AV1 |
| Intel 12th+ Gen | QSV | QSV | N/A | HEVC |
| AMD RX 7000 | AMF | AMF | AMF | AV1 |
| AMD RX 6000 | AMF | AMF | N/A | HEVC |
| Apple M1/M2/M3 | VT | VT | VT | HEVC |

**NVENC Presets:**

| Preset | Use Case | Target Quality | Latency |
|--------|----------|----------------|---------|
| P1 (Fastest) | Lowest latency | Lower | < 1ms |
| P2 | Low latency | Good | < 2ms |
| P4 (Default) | Balanced | Better | < 4ms |
| P6 | Quality | Best | < 8ms |
| P7 (Slowest) | Recording | Best | N/A |

**Acceptance Criteria:**
1. Encoder auto-detection: correct encoder selected on all GPU types
2. NVENC P1 latency: < 1ms (4K H.264), < 2ms (4K HEVC)
3. Encoder fallback: software encoder if hardware unavailable
4. Encoder hot-swap: change encoder without session restart
5. All encoders produce valid bitstreams (validated by decoder)

---

#### B3-T03: Dual-Path Encoding (Stream + Record)

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T03 |
| **Title** | Dual-Path Simultaneous Encode |
| **Effort** | 2pw |

**Description:**
(See P09-T02 for full specification) Summary: Single NVENC session produces both low-latency stream and high-quality recording output using split encoding.

**Files:**
```
+ helix-encode/pkg/dual/path.go                # Dual-path controller
+ helix-encode/pkg/dual/split.go               # Frame split logic
+ helix-encode/pkg/dual/output.go              # Dual output management
```

---

#### B3-T04: Thermal-Aware Quality Scaling

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T04 |
| **Title** | GPU Thermal-Aware Dynamic Quality Scaling |
| **Effort** | 1.5pw |

**Description:**
Implement thermal monitoring and dynamic quality scaling to prevent GPU thermal throttling. Reduce encode quality/resolution before thermal limit is reached.

**Files:**
```
+ helix-encode/pkg/thermal/monitor.go           # GPU temperature monitor
+ helix-encode/pkg/thermal/controller.go        # Thermal control loop
+ helix-encode/pkg/thermal/policy.go            # Scaling policies
+ helix-encode/pkg/thermal/nvml.go              # NVIDIA NVML bindings
+ helix-encode/pkg/thermal/amdsmi.go            # AMD SMI bindings
```

**Thermal Scaling Policy:**

| GPU Temp | Action |
|----------|--------|
| < 70°C | Full quality |
| 70-75°C | Reduce preset by 1 step |
| 75-80°C | Reduce resolution (4K→1440p or 1440p→1080p) |
| 80-83°C | Reduce bitrate by 25% |
| > 83°C | Emergency: minimum quality, log alert |

**Acceptance Criteria:**
1. Thermal throttling time < 2% of session time (measured)
2. Quality reduction: smooth transition (no visible artifact burst)
3. Recovery: quality restored within 30 seconds of temperature drop
4. All GPU vendors supported (NVIDIA, AMD, Intel)

---

### B4: Client Architecture Implementation

---

#### B4-T01: Shared Go Core Compilation

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T01 |
| **Title** | Shared Go Core (c-shared / native / WASM) |
| **Effort** | 2pw |

**Description:**
Implement shared Go core library compiled to multiple targets: `c-shared` (for Flutter FFI), native (for Wails), and WASM (for Angular web). Single codebase, platform-specific build tags.

**Files:**
```
+ helix-core/go.mod                            # Module: github.com/helixplay/helix-core
+ helix-core/Makefile                          # Multi-target build
+ helix-core/pkg/stream/decoder.go             # Video decode abstraction
+ helix-core/pkg/stream/renderer.go            # Frame renderer interface
+ helix-core/pkg/input/client.go               # Input client (send to host)
+ helix-core/pkg/net/webrtc.go                 # WebRTC client
+ helix-core/pkg/net/quic.go                   # QUIC client
+ helix-core/pkg/audio/render.go               # Audio renderer
+ helix-core/build/cshared.go                  # c-shared build directives
+ helix-core/build/wasm.go                     # WASM build directives
```

**Build Targets:**
```makefile
# Makefile targets
build-cshared-linux:
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -o libhelix.so

build-cshared-darwin:
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -o libhelix.dylib

build-cshared-windows:
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -o helix.dll

build-wasm:
    GOOS=js GOARCH=wasm go build -o helix.wasm

build-native:
    go build -o helix
```

**Acceptance Criteria:**
1. Single `go test ./...` passes for all build targets
2. C-shared library exports clean C API (< 20 functions)
3. WASM module < 10MB compressed (downloadable)
4. All targets share identical protocol implementation
5. Build time for all targets < 5 minutes

---

#### B4-T02: Wails Desktop Client

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T02 |
| **Title** | Wails v3 Desktop Client (Windows, macOS, Linux) |
| **Effort** | 2.5pw |

**Description:**
Implement Wails-based desktop client using the shared Go core. Replace any stubs with full implementation: WebRTC, hardware decode, controller input, audio output.

**Files:**
```
+ helix-wails/go.mod                           # Module: github.com/helixplay/helix-wails
+ helix-wails/main.go                          # Wails application entry
+ helix-wails/app.go                           # Wails app configuration
+ helix-wails/frontend/src/main.ts             # Frontend entry
+ helix-wails/frontend/src/App.svelte          # Main app component
+ helix-wails/frontend/src/stream/             # Streaming components
+ helix-wails/frontend/src/input/              # Input handling
+ helix-wails/frontend/src/settings/           # Settings UI
+ helix-wails/frontend/src/library/            # Game library UI
+ helix-wails/frontend/wailsjs/go/             # Wails Go bindings
```

**Acceptance Criteria:**
1. Windows, macOS, Linux builds from single codebase
2. Hardware decode: DXVA2/D3D11VA (Win), VideoToolbox (Mac), VAAPI (Linux)
3. Controller input via helix-input (DirectInput/XInput on Win, IOKit on Mac, evdev on Linux)
4. App size < 50MB (compressed installer)
5. Startup time < 3 seconds (cold)

---

#### B4-T03: Flutter Mobile Client

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T03 |
| **Title** | Flutter Mobile Client (Android, iOS) with Go FFI |
| **Effort** | 2.5pw |

**Description:**
Implement Flutter mobile client with Go FFI integration. Use `c-shared` library for streaming core. Implement platform-specific video rendering (SurfaceView/TextureView on Android, CVPixelBuffer on iOS).

**Files:**
```
+ helix-flutter/pubspec.yaml                   # Flutter dependencies
+ helix-flutter/lib/main.dart                  # App entry
+ helix-flutter/lib/core/bridge.dart           # Go FFI bridge
+ helix-flutter/lib/stream/player.dart         # Video player widget
+ helix-flutter/lib/stream/renderer.dart       # Platform renderer
+ helix-flutter/lib/input/controller.dart      # Controller input
+ helix-flutter/lib/screens/library.dart       # Game library
+ helix-flutter/lib/screens/stream.dart        # Streaming screen
+ helix-flutter/android/app/src/main/kotlin/   # Android platform code
+ helix-flutter/ios/Runner/                    # iOS platform code
```

**Acceptance Criteria:**
1. Android API 28+ and iOS 14+ from single Flutter codebase
2. Hardware decode: MediaCodec (Android), VideoToolbox (iOS)
3. Bluetooth controller pairing: DualSense, Xbox, MFi
4. Touch overlay for games without controller
5. App size: Android < 30MB, iOS < 40MB

---

#### B4-T04: Angular + Go-WASM Web Client

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T04 |
| **Title** | Angular Web Client with Go WASM + WebCodecs |
| **Effort** | 2.5pw |

**Description:**
Implement Angular web client using Go-compiled WASM for protocol handling and WebCodecs API for hardware-accelerated video decode. WebTransport for network layer.

**Files:**
```
+ helix-angular/package.json                   # npm dependencies
+ helix-angular/angular.json                   # Angular config
+ helix-angular/src/main.ts                    # Entry point
+ helix-angular/src/app/app.module.ts          # App module
+ helix-angular/src/app/stream/                # Streaming module
+ helix-angular/src/app/stream/webrtc.service.ts   # WebRTC client
+ helix-angular/src/app/stream/decoder.service.ts  # WebCodecs decoder
+ helix-angular/src/app/stream/renderer.ts     # Canvas/WebGL renderer
+ helix-angular/src/app/input/                 # Input handling
+ helix-angular/src/assets/wasm/               # Go WASM output
+ helix-angular/go.mod                         # Go WASM module
```

**WebCodecs Integration:**
```typescript
// VideoDecoder for H.264/HEVC/AV1
const decoder = new VideoDecoder({
    output: handleDecodedFrame,
    error: handleDecodeError,
});

decoder.configure({
    codec: 'avc1.640033',  // H.264 High Profile Level 5.1
    hardwareAcceleration: 'prefer-hardware',
});
```

**Acceptance Criteria:**
1. Chrome 94+, Firefox 120+, Safari 17+ supported
2. WebCodecs hardware decode: < 5ms decode latency
3. WebTransport for network (fallback to WebRTC datachannels)
4. Go WASM module handles protocol, encryption, input serialization
5. Controller support via WebHID (Chrome) + Gamepad API (all browsers)

---

### B5: Catalog & Content Pipeline

---

#### B5-T01: Multi-Source Metadata Aggregation

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T01 |
| **Title** | IGDB / SteamGridDB / Steam / RAWG Metadata Aggregation |
| **Effort** | 2pw |

**Description:**
Implement metadata aggregation from multiple game databases. Merge and deduplicate entries. Build unified game catalog with rich metadata.

**Files:**
```
+ helix-catalog/go.mod                         # Module: github.com/helixplay/helix-catalog
+ helix-catalog/pkg/sources/igdb.go            # IGDB API client
+ helix-catalog/pkg/sources/steamgriddb.go     # SteamGridDB client
+ helix-catalog/pkg/sources/steam.go           # Steam API client
+ helix-catalog/pkg/sources/rawg.go            # RAWG API client
+ helix-catalog/pkg/merge/engine.go            # Deduplication engine
+ helix-catalog/pkg/merge/score.go             # Match scoring
+ helix-catalog/pkg/models/game.go             # Unified game model
+ helix-catalog/pkg/sync/scheduler.go          # Periodic sync scheduler
+ helix-catalog/pkg/api/catalog.go             # Catalog API
```

**Metadata Fields:**
```go
type Game struct {
    ID              string
    Title           string
    Description     string
    ReleaseDate     time.Time
    Genres          []string
    Platforms       []string
    Developers      []string
    Publishers      []string
    Ratings         map[string]float64  // ESRB, PEGI, Metacritic
    CoverURL        string              // 4K WebP/AVIF
    ArtworkURLs     []string            // Screenshots
    VideoURLs       []string            // Trailers
    SteamAppID      string
    IGDBID          int
    RAWGID          int
    Tags            []string
    Series          string
}
```

**Acceptance Criteria:**
1. Catalog covers > 50,000 games from aggregated sources
2. Deduplication accuracy > 95% (measured via manual sample)
3. Metadata freshness: sync with sources every 24 hours
4. API response time: < 100ms for catalog search
5. Graceful degradation if source API is unavailable

---

#### B5-T02: 4K WebP/AVIF Asset Management

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T02 |
| **Title** | 4K WebP/AVIF Image Asset Pipeline |
| **Effort** | 1.5pw |

**Description:**
Implement asset pipeline that fetches, converts, and serves game artwork in WebP and AVIF formats. Responsive sizing, lazy loading, CDN integration.

**Files:**
```
+ helix-catalog/pkg/assets/pipeline.go         # Asset processing pipeline
+ helix-catalog/pkg/assets/convert.go          # Image format conversion
+ helix-catalog/pkg/assets/resize.go           # Responsive resizing
+ helix-catalog/pkg/assets/storage.go          # Asset storage (S3 + CDN)
+ helix-catalog/pkg/assets/serve.go            # Asset serving with format negotiation
```

**Asset Formats:**

| Format | Role | Quality | Size vs JPEG |
|--------|------|---------|-------------|
| AVIF | Primary (modern clients) | 85 | -60% |
| WebP | Fallback (older clients) | 85 | -30% |
| JPEG | Legacy fallback | 90 | Baseline |

**Sizes:** 256x384 (cover), 1920x1080 (screenshot), 3840x2160 (hero)

**Acceptance Criteria:**
1. AVIF served to supporting browsers (Chrome 85+, Firefox 93+, Safari 16+)
2. WebP served to supporting browsers without AVIF
3. Image response time: < 200ms from CDN edge
4. Original quality preserved in 4K assets
5. Storage: AVIF + WebP pre-generated, no on-the-fly conversion

---

#### B5-T03: Search Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T03 |
| **Title** | Full-Text Search (SQLite FTS5 + Meilisearch) |
| **Effort** | 1.5pw |

**Description:**
Implement two-tier search: SQLite FTS5 for local/offline search, Meilisearch for server-side catalog search. Typo tolerance, faceting, fuzzy matching.

**Files:**
```
+ helix-catalog/pkg/search/fts5.go             # SQLite FTS5 local search
+ helix-catalog/pkg/search/meilisearch.go      # Meilisearch server search
+ helix-catalog/pkg/search/index.go            # Index management
+ helix-catalog/pkg/search/suggestions.go      # Autocomplete/suggestions
```

**Acceptance Criteria:**
1. Local search: SQLite FTS5, works offline, < 50ms response
2. Server search: Meilisearch, typo-tolerant, < 100ms response
3. Autocomplete suggestions: < 30ms
4. Faceted search by genre, platform, release year, rating
5. Search index updated within 5 minutes of catalog change

---

#### B5-T04: Per-Tenant Catalog Overlays

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T04 |
| **Title** | Per-Tenant Catalog Customization |
| **Effort** | 1pw |

**Description:**
Allow tenants to customize their game catalog: add/remove games, custom artwork, pricing, featured sections. Overlay on top of global catalog.

**Files:**
```
+ helix-tenant/pkg/catalog/overlay.go          # Catalog overlay engine
+ helix-tenant/pkg/catalog/custom.go           # Custom game entries
+ helix-tenant/pkg/catalog/featured.go         # Featured sections
+ helix-tenant/pkg/catalog/pricing.go          # Per-tenant pricing
```

**Acceptance Criteria:**
1. Tenant can hide games from global catalog
2. Tenant can add custom games (not in global catalog)
3. Tenant can override artwork for any game
4. Tenant can set custom pricing/subscription model
5. Tenant can create featured sections and curated lists

---

### B6: White-Label & Theming

---

#### B6-T01: 3-Tier Design Token System

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T01 |
| **Title** | 3-Tier Design Tokens (Primitive → Semantic → Component) |
| **Effort** | 1.5pw |

**Description:**
Implement 3-tier design token architecture using Style Dictionary v4. Tokens define all visual properties: colors, typography, spacing, elevation, motion.

**Files:**
```
+ helix-theme/go.mod                           # Module: github.com/helixplay/helix-theme
+ helix-theme/tokens/primitive/colors.json     # Primitive color tokens
+ helix-theme/tokens/primitive/typography.json # Primitive type tokens
+ helix-theme/tokens/primitive/spacing.json    # Primitive spacing tokens
+ helix-theme/tokens/semantic/light.json       # Semantic tokens (light)
+ helix-theme/tokens/semantic/dark.json        # Semantic tokens (dark)
+ helix-theme/tokens/component/button.json     # Component tokens
+ helix-theme/tokens/component/card.json
+ helix-theme/tokens/component/input.json
+ helix-theme/build.js                         # Style Dictionary v4 build
+ helix-theme/config.json                      # SD configuration
```

**Token Hierarchy:**
```
primitive/
  color.blue.500 = "#2196F3"
  color.red.500 = "#F44336"
  spacing.4 = "16px"

semantic/
  color.primary = { primitive.color.blue.500 }
  color.error = { primitive.color.red.500 }
  spacing.section = { primitive.spacing.4 }

component/
  button.background = { semantic.color.primary }
  button.padding = { semantic.spacing.section }
```

**Acceptance Criteria:**
1. All visual properties defined as tokens (zero hardcoded values)
2. Theme switch (light/dark): instant, no page reload
3. Custom tenant theme generated from brand colors (< 5 minutes)
4. Token outputs: CSS variables, JSON, Dart, TypeScript
5. Style Dictionary build: < 30 seconds

---

#### B6-T02: Material Design 3 Integration

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T02 |
| **Title** | Material Design 3 (Material You) Integration |
| **Effort** | 1pw |

**Description:**
Integrate Material Design 3 across all three client platforms. Use M3 components as base, customize with design tokens.

**Files:**
```
+ helix-theme/tokens/m3/ref.json               # M3 reference tokens
+ helix-theme/tokens/m3/sys.json               # M3 system tokens
~ helix-wails/frontend/src/theme/m3.ts         # M3 theme (Wails)
~ helix-flutter/lib/theme/m3.dart              # M3 theme (Flutter)
~ helix-angular/src/theme/m3.scss              # M3 theme (Angular)
```

**Acceptance Criteria:**
1. All UI components use M3 design language
2. Dynamic color (Material You): theme derived from game artwork on Android
3. Consistent visual language across all three platforms
4. Accessibility: WCAG 2.1 AA compliance (contrast ratios)

---

#### B6-T03: Style Dictionary v4 Build Pipeline

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T03 |
| **Title** | Style Dictionary v4 Build & Distribution |
| **Effort** | 0.5pw |

**Description:**
Automated token build pipeline producing platform-specific outputs.

**Files:**
```
+ .github/workflows/tokens-build.yml           # Token build CI
+ helix-theme/package.json                     # npm dependencies
```

**Build Outputs:**
| Platform | Format | Destination |
|----------|--------|-------------|
| Web (Wails/Angular) | CSS custom properties | `*.css` |
| Flutter | Dart class | `*.dart` |
| Design tools | JSON | Figma plugin |

**Acceptance Criteria:**
1. CI builds all token outputs on every token change
2. Output files distributed to client repos via automated PR
3. Token validation: no undefined references, no circular dependencies

---

#### B6-T04: Per-Tenant Identity

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T04 |
| **Title** | Per-Tenant Brand Identity System |
| **Effort** | 1pw |

**Description:**
Complete white-label identity: logo, brand colors, fonts, app icon, splash screen. Generated from tenant configuration.

**Files:**
```
+ helix-tenant/pkg/branding/generator.go       # Brand asset generator
+ helix-tenant/pkg/branding/logo.go            # Logo processing
+ helix-tenant/pkg/branding/colors.go          # Brand color extraction
+ helix-tenant/pkg/branding/fonts.go           # Font loading
```

**Acceptance Criteria:**
1. Tenant brand colors extracted from logo (dominant color algorithm)
2. App icon generated with tenant logo
3. Splash screen themed with tenant brand
4. All branding applied within 30 seconds of tenant config change

---

### B7: Operations & Observability

---

#### B7-T01: OpenTelemetry Integration

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T01 |
| **Title** | OpenTelemetry Tracing & Metrics |
| **Effort** | 1.5pw |

**Description:**
Implement OpenTelemetry tracing and metrics across all services. Distributed tracing for request flows, custom metrics for gaming-specific KPIs.

**Files:**
```
+ helix-observability/go.mod                   # Module: github.com/helixplay/helix-observability
+ helix-observability/pkg/trace/provider.go    # OTel trace provider
+ helix-observability/pkg/metrics/provider.go  # OTel metrics provider
+ helix-observability/pkg/metrics/gaming.go    # Gaming-specific metrics
+ helix-observability/pkg/log/otel.go          # OTel log correlation
```

**Gaming-Specific Metrics:**
```go
var (
    FrameLatency = meter.Float64Histogram("helix.frame_latency_ms",
        "Frame glass-to-glass latency")
    InputLatency = meter.Float64Histogram("helix.input_latency_ms",
        "Controller input latency")
    EncodeTime = meter.Float64Histogram("helix.encode_time_ms",
        "Frame encode time")
    NetworkRTT = meter.Float64Histogram("helix.network_rtt_ms",
        "Network round-trip time")
    GpuUtilization = meter.Float64ObservableGauge("helix.gpu_utilization",
        "GPU utilization percent")
    ThermalTemp = meter.Float64ObservableGauge("helix.gpu_temperature_c",
        "GPU temperature Celsius")
)
```

**Acceptance Criteria:**
1. All API requests traced with distributed trace IDs
2. Gaming metrics collected every frame (no sampling in hot path)
3. Trace sampling: 100% for errors, 1% for success (configurable)
4. Export to Jaeger + Prometheus
5. Trace correlation across all 29 submodules

---

#### B7-T02: Prometheus Metrics

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T02 |
| **Title** | Prometheus Metrics Export & Alerting |
| **Effort** | 1pw |

**Description:**
Prometheus metrics export with custom collectors for gaming KPIs. Grafana dashboards and alert rules.

**Files:**
```
+ helix-observability/pkg/prometheus/registry.go   # Prometheus registry
+ helix-observability/pkg/prometheus/collectors.go # Custom collectors
+ infra/monitoring/grafana/dashboards/             # Grafana dashboards
+ infra/monitoring/prometheus/rules.yml            # Alert rules
```

**Alert Rules:**

| Alert | Condition | Severity |
|-------|-----------|----------|
| HighLatency | p99 frame latency > 50ms | warning |
| CriticalLatency | p99 frame latency > 100ms | critical |
| GpuThermal | GPU temp > 83°C | warning |
| HighErrorRate | Error rate > 1% | critical |
| DiskFull | Disk usage > 90% | warning |
| MemoryPressure | Memory usage > 95% | critical |

**Acceptance Criteria:**
1. All metrics exposed on `/metrics` endpoint
2. Grafana dashboards for: streaming quality, GPU health, network, sessions
3. AlertManager routes alerts to PagerDuty/Slack
4. Alert firing latency < 30 seconds from threshold breach

---

#### B7-T03: Structured JSON Logging

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T03 |
| **Title** | Structured JSON Logging with Correlation IDs |
| **Effort** | 0.5pw |

**Description:**
Structured JSON logging with request correlation IDs, tenant context, and gaming-specific fields.

**Files:**
```
+ helix-observability/pkg/log/logger.go        # Structured logger
+ helix-observability/pkg/log/fields.go        # Gaming log fields
+ helix-observability/pkg/log/middleware.go    # HTTP/gRPC logging middleware
```

**Log Schema:**
```json
{
    "ts": "2025-01-15T10:30:00.000Z",
    "level": "info",
    "msg": "session.started",
    "trace_id": "abc123",
    "tenant_id": "tenant-42",
    "session_id": "sess-789",
    "user_id": "user-456",
    "game_id": "game-123",
    "codec": "av1",
    "resolution": "3840x2160",
    "fps": 60,
    "host_id": "host-gpu-01",
    "region": "us-east-1"
}
```

**Acceptance Criteria:**
1. All logs structured JSON (no plaintext)
2. Correlation ID propagated across all service boundaries
3. Tenant ID in every log entry (for multi-tenant filtering)
4. Log aggregation: Fluent Bit → Loki / ELK
5. Log query response: < 2 seconds for 24-hour search

---

#### B7-T04: NATS Event Bus

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T04 |
| **Title** | NATS JetStream Event Bus |
| **Effort** | 1pw |

**Description:**
NATS JetStream as primary event bus for async communication between services. Tenant-scoped subjects, durable consumers, exactly-once processing.

**Files:**
```
+ helix-events/go.mod                          # Module: github.com/helixplay/helix-events
+ helix-events/pkg/nats/client.go              # NATS client
+ helix-events/pkg/nats/jetstream.go           # JetStream management
+ helix-events/pkg/nats/publisher.go           # Event publisher
+ helix-events/pkg/nats/consumer.go            # Event consumer
+ helix-events/pkg/nats/events.go              # Event type definitions
```

**Event Types:**
```go
const (
    EventSessionStarted   = "helix.session.started"
    EventSessionEnded     = "helix.session.ended"
    EventFrameEncoded     = "helix.frame.encoded"
    EventInputReceived    = "helix.input.received"
    EventQualityChanged   = "helix.quality.changed"
    EventRecordingSaved   = "helix.recording.saved"
    EventUserAuthenticated = "helix.user.authenticated"
)
```

**Subject Topology:**
```
events.{tenant_id}.{event_type}
metrics.{tenant_id}.{metric_name}
commands.{service_id}.{command_type}
```

**Acceptance Criteria:**
1. Event publish latency < 1ms (localhost NATS)
2. Tenant-scoped subjects enforce isolation
3. Durable consumers: no event loss on consumer restart
4. Exactly-once semantics for billing events
5. JetStream retention: 7 days for events, 30 days for audit

---

#### B7-T05: Four-Mirror Git Topology Automation

| Attribute | Detail |
|-----------|--------|
| **ID** | B07-T05 |
| **Title** | Four-Mirror Git Repository Automation |
| **Effort** | 0.5pw |

**Description:**
Automated synchronization of all 29 submodules across four Git hosting platforms.

**Files:**
```
+ scripts/mirror/sync.sh                       # Sync orchestrator
+ scripts/mirror/verify.sh                     # Verification script
+ .github/workflows/mirror-sync.yml            # Post-push sync trigger
```

**Acceptance Criteria:**
1. All 29 repos synced to 4 mirrors within 5 minutes of primary push
2. Tags, releases, and branch protection rules synced
3. Automated recovery on sync failure (retry + alert)
4. Weekly verification report

---

#### B7-T06: GitHub Projects + GitLab Tracking

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T06 |
| **Title** | Cross-Platform Project Tracking |
| **Effort** | 0.5pw |

**Description:**
Bidirectional sync between GitHub Projects (primary) and GitLab issues (mirror tracking).

**Files:**
```
+ scripts/tracking/sync.sh                     # Issue sync script
+ .github/workflows/tracking-sync.yml          # Sync workflow
```

**Acceptance Criteria:**
1. GitHub Project status changes reflected in GitLab within 10 minutes
2. GitLab issue comments synced to GitHub
3. Sprint/milestone alignment across both platforms

---

## Appendix A: Submodule Registry

### 29 Public Go Submodules

| # | Module | Purpose | Priority | Status |
|---|--------|---------|----------|--------|
| 1 | `helix-core` | Shared streaming core | P1 | Planned |
| 2 | `helix-stream` | WebRTC/QUIC streaming | P1 | Planned |
| 3 | `helix-capture` | Screen capture (DXGI/SCK/PipeWire) | P1 | Planned |
| 4 | `helix-encode` | Hardware encoder factory | P1 | Planned |
| 5 | `helix-input` | Controller input (1kHz DualSense) | P1 | Planned |
| 6 | `helix-audio` | Audio pipeline (Opus MultiStream) | P2 | Planned |
| 7 | `helix-record` | Recording & replay | P2 | Planned |
| 8 | `helix-auth` | OAuth2/OIDC/RBAC | P1 | Planned |
| 9 | `helix-tenant` | Multi-tenant isolation | P1 | Planned |
| 10 | `helix-catalog` | Game metadata catalog | P2 | Planned |
| 11 | `helix-theme` | Design tokens & theming | P2 | Planned |
| 12 | `helix-deploy` | Canary deployment | P2 | Planned |
| 13 | `helix-vault` | Secrets management | P1 | Planned |
| 14 | `helix-audit` | Audit logging | P1 | Planned |
| 15 | `helix-events` | NATS event bus | P1 | Planned |
| 16 | `helix-observability` | OpenTelemetry/metrics | P1 | Planned |
| 17 | `helix-rtos` | RTOS scheduling utilities | P1 | Planned |
| 18 | `helix-iouring` | io_uring async I/O | P1 | Planned |
| 19 | `helix-lockfree` | Lock-free data structures | P1 | Planned |
| 20 | `helix-shm` | Shared memory IPC | P1 | Planned |
| 21 | `helix-mempool` | Slab memory pools | P1 | Planned |
| 22 | `helix-allocator` | Custom allocators | P1 | Planned |
| 23 | `helix-gpu-direct` | GPU Direct zero-copy | P2 | Planned |
| 24 | `helix-wails` | Wails desktop client | P2 | Planned |
| 25 | `helix-flutter` | Flutter mobile/TV client | P2 | Planned |
| 26 | `helix-angular` | Angular+WASM web client | P2 | Planned |
| 27 | `helix-r18-safeexec` | R-18 security framework | P1 | Planned |
| 28 | `sunshine` | Sunshine++ host agent (fork) | P1 | Planned |
| 29 | `helix-docs` | Documentation site | P2 | Planned |

---

## Appendix B: Dependency Graph

### Cross-Module Dependencies (Key Paths)

```
helix-r18-safeexec (root)
  ├── helix-core
  │     ├── helix-stream
  │     │     ├── helix-capture
  │     │     ├── helix-encode
  │     │     ├── helix-audio
  │     │     └── helix-record
  │     ├── helix-input
  │     └── helix-shm
  ├── helix-rtos
  │     ├── helix-iouring
  │     ├── helix-lockfree
  │     ├── helix-mempool
  │     └── helix-allocator
  ├── helix-auth
  │     ├── helix-vault
  │     └── helix-tenant
  │           ├── helix-catalog
  │           ├── helix-theme
  │           └── helix-events
  ├── helix-observability
  │     └── helix-audit
  ├── helix-deploy
  └── helix-gpu-direct

Client Modules (depend on helix-core):
  ├── helix-wails
  ├── helix-flutter
  └── helix-angular

Host Agent:
  └── sunshine → depends on helix-capture, helix-encode, helix-input,
                   helix-audio, helix-shm, helix-rtos
```

---

*End of Document*

**Document History:**
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2025-01 | Architecture Team | Initial release - Advanced Phases P07-P13 + Architecture Deep-Dive B1-B7 |
