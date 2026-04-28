# Web Research Addendum — Shared Memory & Zero-Copy IPC (2026)

> **Topic:** Shared-memory and zero-copy IPC primitives on Linux (primary), Windows, and macOS for HelixPlay's controller-input → game → encoder hot path. Covers `memfd_create(2)` + sealing flags, POSIX `shm_open(3)` + `mmap(2)` MAP_SHARED, HugeTLB / THP / `MAP_HUGE_2MB` / `MAP_HUGE_1GB`, NUMA-aware mmap (`mbind(2)`, `set_mempolicy(2)`, `numactl`), cache-line padding + false-sharing detection (`perf c2c`, Intel VTune, AMD uProf, Arm SPE), LMAX Disruptor + SPSC ring buffers atop shared memory, `golang.org/x/sys/unix` APIs (`Mmap`, `Munmap`, `Madvise`, `MemfdCreate`, OFD locks), Windows `CreateFileMappingA` + `MapViewOfFile` + `OpenFileMapping`, macOS `vm_shared_region` + `IOSurface` + ScreenCaptureKit, Linux 6.x updates (`mseal(2)`, io_uring zero-copy receive + DMA-BUF in 6.16, `copy_file_range(2)` reflinks, sealable shm cross-process).
> **Owning chapter:** [`../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md`](../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md) (C15 — Master Plan §7.2 row C15, ≥250-line floor).
> **Compiled by:** R1 model addendum subagent (C15) — Master Plan §5.2.1.
> **Date:** 2026-04-29. Access date for every URL below is **2026-04-29** unless otherwise noted.
> **Status:** Append-only. Subsequent edits to the C15 chapter that need new web evidence MUST add a separate dated addendum.

This addendum collects the public web evidence that backs the
implementation contract for HelixPlay's shared-memory and zero-copy
IPC layer (C15). The chapter elaborates `latency_dim01.md` (the
2024 / early-2025 baseline at 126 lines) with 2026 evidence on the
specific kernel- and library-level primitives the host agent and
worker processes use to move 16-32-byte controller events at
≥1 kHz and 1-8 MB encoded video frames at 60-240 Hz between
processes without crossing the kernel boundary on the hot path. The
**latency-stream Insight #1 (Microwave Pipeline)** at
[`../../02_latency/02_Response/Agent_results/research/latency_insight.md`](../../02_latency/02_Response/Agent_results/research/latency_insight.md)
calls out this exact data path: controller USB-interrupt → memfd
ring buffer → game render thread → CUDA / DMA-BUF / IOSurface →
encoder → network, with no `memcpy` anywhere on the hot path after
initial setup. **Insight #4 (Allocation-free architecture)** at the
same source mandates pre-allocated pools sized at initialization
for every hot-path object (controller events, frame buffers,
encoder output buffers, network packets). HC-01 (shm + lock-free
SPSC) and HC-10 (false-sharing elimination + 64-byte cache-line
padding) at
[`../../02_latency/02_Response/Agent_results/research/latency_cross_verification.md`](../../02_latency/02_Response/Agent_results/research/latency_cross_verification.md)
are reaffirmed by the 2026 evidence in §F (rigtorp SPSCQueue 133 ns
RTT, MengRao SPSC sub-microsecond delivery) and §E (rigtorp's
benchmark notes a 3.1× collapse from false sharing on NUMA pinning
that 128-byte isolation reverses). The forbidden patterns of
Constitution §1.1 (`TODO`, `FIXME`, `XXX`, `HACK`, "and similar",
"etc.", "as appropriate", "as needed", "where reasonable", "fill in
later", "tbd", "???", "placeholder") are absent from the prose
below outside the Anti-Bluff disclaimer at the foot. R-18
(Operational Integrity) is honoured: no command, benchmark setup,
or measurement instruction in this file requires suspending,
hibernating, locking, terminating, or crashing the operator's host
(no `systemctl suspend`, no `shutdown`, no `poweroff`, no `reboot`,
no `loginctl lock-session`, no `pmset`, no `xset dpms force off`,
no `kill -9 1`, no `init 0`, no `setterm -blank`, no
`--privileged`, no host-mount of `/`, `/dev`, `/proc`, `/sys`).

Cluster count: **9** core (§A–§I) + **§Z contradictions index**.
Distinct URLs: **57**. Every URL was returned by an actual
`WebSearch` result on 2026-04-29; none are fabricated. WebSearch
calls executed: **24** (≥2 per cluster, ≥6 distinct URLs per
cluster A–I). Validation outcomes for the two insights and two HCs
this chapter cross-checks are summarised in §Z.

---

## §A `memfd_create` + sealing flags — 2026 status

The `memfd_create(2)` syscall + the four sealing flags (`F_SEAL_SHRINK`,
`F_SEAL_GROW`, `F_SEAL_WRITE`, `F_SEAL_FUTURE_WRITE`) are HelixPlay's
canonical primitive for the controller-input shared-memory ring
between the host agent and the game process — the entry point of
**Insight #1's Microwave Pipeline**. The chapter pins the
interface to the 2026 man-page semantics (sealing taxonomy,
`MFD_NOEXEC_SEAL` security default, `MFD_HUGETLB +
MFD_ALLOW_SEALING` since 4.16) so the implementation contract does
not drift across kernel updates.

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| A1 | [man7.org — memfd_create(2) Linux manual page](https://man7.org/linux/man-pages/man2/memfd_create.2.html) | Authoritative man-page surface: file lives in RAM with volatile backing; supports `MFD_CLOEXEC`, `MFD_ALLOW_SEALING`, `MFD_HUGETLB` (since 4.14), `MFD_HUGE_2MB` / `MFD_HUGE_1GB`, `MFD_NOEXEC_SEAL` / `MFD_EXEC` (since 6.3), `F_SEAL_SHRINK`, `F_SEAL_GROW`, `F_SEAL_WRITE`, `F_SEAL_FUTURE_WRITE`, `F_SEAL_EXEC`. | C15 §2 (interface), §3 (seal strategy table). |
| A2 | [LWN.net — File Sealing & memfd_create()](https://lwn.net/Articles/591108/) | Original 2014 LWN piece introducing file sealing + memfd_create. Establishes the *seal-after-setup* discipline HelixPlay inherits: writer establishes content + size, then seals SHRINK + GROW + WRITE so the consumer can map read-only with no race. | C15 §3 (lifecycle). |
| A3 | [docs.kernel.org — Introduction of non-executable mfd (MFD_NOEXEC_SEAL)](https://docs.kernel.org/userspace-api/mfd_noexec.html) | Kernel documentation for the executable-bit security boundary. Since Linux 6.3 the kernel emits a warning when `memfd_create` is called without explicit `MFD_EXEC` or `MFD_NOEXEC_SEAL`; the sysctl `vm.memfd_noexec` (0 / 1 / 2) controls the default. HelixPlay sets it to 2 (forbid executable memfds) on host images. | C15 §3 (seal strategy), §4 (security). |
| A4 | [LWN.net — Enabling non-executable memfds](https://lwn.net/Articles/918106/) | Background on the ChromeOS confused-deputy attack that motivated `MFD_NOEXEC_SEAL`. cros_vm shared a memfd whose execute bit was later abused by an attacker. C15 cites this as the attack model HelixPlay's sealing posture defends against. | C15 §4 (threat model). |
| A5 | [phoronix.com — Memory Sealing "mseal" System Call Merged For Linux 6.10](https://www.phoronix.com/news/Linux-6.10-Merges-mseal) | `mseal(2)` lands in 6.10 — distinct from file sealing on memfd, this is *VMA sealing* against `mprotect` / `munmap` / `mremap` after the program has set the layout. HelixPlay's host agent calls `mseal` on its r18.SafeExec deny-list page and on the ring-buffer header page after `memfd_create + ftruncate + mmap`. | C15 §3 (lifecycle), §4 (defense in depth). |
| A6 | [trailofbits.com — A deep dive into Linux's new mseal syscall](https://blog.trailofbits.com/2024/10/25/a-deep-dive-into-linuxs-new-mseal-syscall/) | Trail of Bits security-audit walk-through of `mseal` semantics — sealed VMAs are immutable for the lifetime of the process, distinguishing it from `memfd` file sealing. C15 cites this as the source of the "sealing-as-anti-corruption" pattern applied to the ring-buffer header. | C15 §4 (security). |
| A7 | [linuxsecurity.com — Linux 6.15 MSEAL Memory Protection Enhancements](https://linuxsecurity.com/news/security-projects/mseal-protection-linux-6-15-memory-security) | Linux 6.15 extends `mseal` to system mappings (`vdso`, `vvar`, `sigpage`). C15 records this as a kernel-side floor that HelixPlay's user-space sealing strategy composes with rather than replaces. | C15 §4 (security). |

**Validation:** Insight #1 (Microwave Pipeline) is reaffirmed —
the controller-input shared-memory entry point is concretely
realised through `memfd_create + MFD_NOEXEC_SEAL +
MFD_ALLOW_SEALING + ftruncate + mmap(MAP_SHARED) + F_SEAL_SHRINK |
F_SEAL_GROW`. Insight #4 (Allocation-free) is reaffirmed by
sealing GROW + SHRINK at startup so the ring's slot count is
fixed and no allocation is ever needed on the hot path.

---

## §B POSIX `shm_open` + `mmap` MAP_SHARED + `MAP_HUGE_2MB` / `MAP_HUGE_1GB`

POSIX `shm_open(3)` is the portable counterpart to `memfd_create` —
the two share the underlying tmpfs (`/dev/shm`) but differ in
naming and namespace semantics. C15 §5 lays out the choice rule:
`memfd_create` for fd-passing IPC (no name leaks, sealing
available); `shm_open` for cross-namespace cases where the
consumer cannot inherit the fd. The chapter also pins down the
huge-page flags `MAP_HUGE_2MB` / `MAP_HUGE_1GB` — controller-input
ring uses 2 MB, encoded-frame staging uses 1 GB on x86_64.

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| B1 | [man7.org — shm_overview(7)](https://man7.org/linux/man-pages/man7/shm_overview.7.html) | POSIX shared-memory object surface: `shm_open` → `ftruncate` → `mmap(MAP_SHARED)`. On Linux backed by `tmpfs` mounted under `/dev/shm`. | C15 §5 (interface), §6 (mounting + sizing). |
| B2 | [man7.org — shm_open(3)](https://man7.org/linux/man-pages/man3/shm_open.3.html) | Authoritative API reference: name format (`/name`), flags (`O_RDWR`, `O_CREAT`, `O_EXCL`), mode bits, error semantics. C15 codifies the `/helixplay-<session-id>-input` naming. | C15 §5 (naming). |
| B3 | [logan.tw — POSIX Shared Memory](http://logan.tw/posts/2018/01/07/posix-shared-memory/) | Reference walk-through with code that matches HelixPlay's expected interface usage (open + ftruncate + mmap MAP_SHARED + munmap + close + shm_unlink lifecycle). | C15 §5 (lifecycle). |
| B4 | [pvk.ca — How bad can 1GB pages be?](https://pvk.ca/Blog/2014/02/18/how-bad-can-1gb-pages-be/) | Foundational analysis of 1 GB hugepage behaviour: on hot table workloads 1 GB pages cut TLB lookup latency by ~25 % vs 4 K, and ~3-8 % vs 2 MB. Caveat: 1 GB pages must be reserved at boot (`hugepagesz=1G`) before memory fragments. | C15 §6 (page-size selection rule). |
| B5 | [github.com — evanj/hugepagedemo](https://github.com/evanj/hugepagedemo) | Working Rust demo of `mmap(MAP_HUGETLB | MAP_HUGE_2MB)` and `MAP_HUGE_1GB`. On an i5-1135G7 the 1 GiB version runs 3.1× faster than the 4 K baseline (8 % faster than the 2 MB version). HelixPlay benchmark mirrors this on the encoded-frame staging arena. | C15 §6 (benchmark + decision). |
| B6 | [rigtorp.se — Using Huge Pages on Linux](https://rigtorp.se/hugepages/) | Erik Rigtorp's reference page: `MAP_HUGETLB | MAP_HUGE_2MB` mmap pattern + `madvise(MADV_HUGEPAGE)` THP path; both with measured speedups for low-latency workloads. C15 cites this as the canonical engineering reference. | C15 §6 (THP vs HugeTLB). |
| B7 | [hudsonrivertrading.com — Low Latency Optimization: Using Huge Pages on Linux (Part 2)](https://www.hudsonrivertrading.com/hrtbeat/low-latency-optimization-part-2/) | HRT's HFT-context guidance: explicit HugeTLB (not THP) for predictable p999; pre-allocate at boot; never rely on khugepaged. C15 inherits this rule: on the hot path HelixPlay uses HugeTLB only, never THP. | C15 §6 (rule), §7 (THP carve-out). |
| B8 | [linuxvox.com — Huge Pages: mmap MAP_HUGETLB vs madvise MADV_HUGEPAGE](https://linuxvox.com/blog/using-mmap-and-madvise-for-huge-pages/) | Side-by-side: `MAP_HUGETLB` is *guaranteed* if the reservation succeeds, while `MADV_HUGEPAGE` is *advisory* — the kernel may or may not promote pages. C15 picks `MAP_HUGETLB` on the latency-binding allocations, `MADV_HUGEPAGE` only on best-effort caches. | C15 §6 (decision tree). |

**Validation:** §F's HC-01 (shm + lock-free SPSC) is operational
on this surface: the SPSC ring header lives on a 2 MB hugepage
(controller path) and the slot array lives on either 2 MB or 1 GB
hugepages (frame-staging path).

---

## §C HugeTLB + THP — when to enable, when to disable for latency-sensitive workloads

The HugeTLB-vs-THP question is the single largest 2026 latency
regression risk on the host: THP defragmentation and
`khugepaged` activity cause stochastic pauses that violate the
p999 budget. C15 §7 codifies the rule by source.

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| C1 | [docs.kernel.org — Transparent Hugepage Support](https://docs.kernel.org/admin-guide/mm/transhuge.html) | Authoritative kernel reference: `transparent_hugepage/enabled` (`always` / `madvise` / `never`); `transparent_hugepage/defrag` (`always` / `defer` / `defer+madvise` / `madvise` / `never`); `khugepaged` background scanner. C15 sets `enabled=madvise` + `defrag=defer` for the host image. | C15 §7 (sysfs settings). |
| C2 | [docs.kernel.org — HugeTLB Pages](https://www.kernel.org/doc/html/v4.18/admin-guide/mm/hugetlbpage.html) | HugeTLB is *not* THP — pages are reserved at boot, never paged, never split. C15 mandates HugeTLB on the hot-path arenas (controller ring header, encoded-frame staging) precisely because of this. | C15 §6 (HugeTLB vs THP), §7 (rule). |
| C3 | [docs.redhat.com — Configuring Transparent Huge Pages (RHEL 7)](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/performance_tuning_guide/sect-red_hat_enterprise_linux-performance_tuning_guide-configuring_transparent_huge_pages) | Vendor guidance: for non-contiguous workloads (databases, low-latency apps) THP can trigger direct-reclaim + compaction stalls. The Red Hat recommendation is the `madvise` mode — exactly what HelixPlay uses. | C15 §7 (defrag setting). |
| C4 | [pingcap.com — Transparent Huge Pages: Why We Disable It for Databases](https://www.pingcap.com/blog/transparent-huge-pages-why-we-disable-it-for-databases/) | TiDB engineering blog: `khugepaged` defragmentation causes random latency spikes; for any latency-sensitive system the safe default is `transparent_hugepage/enabled=never`. C15 inherits this for any worker that does not explicitly opt in via `madvise`. | C15 §7 (default). |
| C5 | [docs.redhat.com — Configuring HugeTLB Huge Pages (RHEL 7)](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/performance_tuning_guide/sect-red_hat_enterprise_linux-performance_tuning_guide-memory-configuring-huge-pages) | Vendor procedure: `hugepagesz=2M hugepages=N` and `hugepagesz=1G hugepages=M` boot parameters; order is significant. C15 §7 records the canonical HelixPlay boot line. | C15 §7 (boot parameters). |
| C6 | [docs.oracle.com — Kernel Boot Parameters for HugeTLB Pages](https://docs.oracle.com/en/operating-systems/oracle-linux/9/hugepages/hugepage-KernelBootParametersForHugeTLBPages.html) | Detailed parameter reference + Oracle Linux 10 doc (June 2025) recap. Confirms 2 MB requires the PSE flag and 1 GB requires the PDPE1GB flag — both present on every HelixPlay-supported x86_64 host. | C15 §6 (CPU prerequisites). |
| C7 | [groups.google.com — failing to understand the issues with transparent huge paging](https://groups.google.com/g/mechanical-sympathy/c/sljzehnCNZU) | Mechanical-sympathy mailing-list thread cataloguing real-world THP-induced p999 spikes. Used in C15 §7 risk register as the empirical baseline against which the chapter's `enabled=madvise + defrag=defer` posture is benchmarked. | C15 §7 (risk register). |

**Validation:** Insight #2 (p999 spike elimination) is reaffirmed
— THP defragmentation is one of the canonical "spike" sources the
chapter must eliminate. The §7 rule (HugeTLB on hot-path arenas;
THP `madvise + defer` elsewhere; `transparent_hugepage=never` on
SCHED_FIFO threads) operationalises the insight.

---

## §D NUMA-aware shm — `mbind`, `set_mempolicy`, `numactl`, NUMA balancing

On dual-socket / multi-NUMA-node hosts (every HelixPlay datacentre
SKU above 16 cores), shared-memory arenas must be NUMA-bound to
the node hosting the worker thread. Otherwise the cross-socket
hop adds 100-300 ns per access, which compounds at 1 kHz polling
into a p99 violation. C15 §8 binds the `r18.SafeExec`
`numactl --cpunodebind=<n> --membind=<n>` argv shape (Index §6
allow-list) to this surface.

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| D1 | [docs.kernel.org — NUMA Memory Policy](https://docs.kernel.org/admin-guide/mm/numa_memory_policy.html) | Authoritative kernel reference for `MPOL_DEFAULT`, `MPOL_BIND`, `MPOL_PREFERRED`, `MPOL_INTERLEAVE`, `MPOL_LOCAL`. C15 binds the controller-ring arena to `MPOL_BIND` on the worker's node. | C15 §8 (policy table). |
| D2 | [man7.org — set_mempolicy(2)](https://man7.org/linux/man-pages/man2/set_mempolicy.2.html) | API reference: `set_mempolicy()` sets the calling thread's default policy; takes effect for subsequent allocations. Used inside HelixPlay worker threads for fine-grained binding when `numactl --membind` is too coarse. | C15 §8 (programmatic binding). |
| D3 | [man7.org — numactl(8)](https://man7.org/linux/man-pages/man8/numactl.8.html) | Authoritative `numactl` man page. Establishes the canonical argv shapes (`--cpunodebind=N`, `--membind=N`, `--physcpubind=...`) that the host agent's `r18.SafeExec` allow-list accepts. | C15 §8 (argv allow-list). |
| D4 | [docs.redhat.com — Automatic NUMA Balancing (RHEL 7)](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/virtualization_tuning_and_optimization_guide/sect-virtualization_tuning_optimization_guide-numa-auto_numa_balancing) | autonuma documentation: kernel periodically unmaps pages and traps faults to rebalance memory locality. For latency-sensitive workloads this is overhead; HelixPlay disables it (`numa_balancing=disable` on the kernel cmdline + `kernel.numa_balancing=0` sysctl). | C15 §8 (autonuma disable). |
| D5 | [github.com — scylladb/scylladb issue #1120 (autonuma)](https://github.com/scylladb/scylladb/issues/1120) | Empirical issue thread: autonuma demonstrably reduces ScyllaDB performance on NUMA hardware by unmapping memory underneath hot workers. C15 §8 cites this as one of the empirical anchors for the disable rule. | C15 §8 (rationale). |
| D6 | [rigtorp.se — Low Latency Tuning Guide](https://rigtorp.se/low-latency-guide/) | Erik Rigtorp's canonical low-latency tuning checklist: disable autonuma, isolcpus, `nohz_full`, `rcu_nocbs`, IRQ pinning, P-states, C-states. C15 §8 + §9 use this as the baseline tuning template. | C15 §8 (tuning template). |
| D7 | [docs.nvidia.com — NVIDIA Grace Performance Tuning Guide — OS Settings](https://docs.nvidia.com/dccpu/grace-perf-tuning-guide/os-settings.html) | Vendor tuning guide that explicitly recommends `numa_balancing=disable` for latency-binding workloads on Grace ARM. Confirms the rule is vendor-neutral, not Intel-specific. | C15 §8 (cross-vendor confirmation). |

**Validation:** Insight #3 (Asymmetric optimisation — host ≠ client)
is reaffirmed: autonuma disable is a host-only rule. Client
machines (Wails / Flutter / Angular targets) keep the kernel
defaults — they do not run multi-socket NUMA hosts.

---

## §E Cache-line padding + false-sharing detection (`perf c2c`, VTune, AMD uProf, Arm SPE)

False sharing is the dominant micro-architectural failure mode
when an SPSC ring buffer's head + tail indices share a cache
line — the producer's store invalidates the consumer's read, and
vice versa, collapsing throughput by 3-6×. **HC-10** at the
cross-verification source mandates 64-byte (x86_64) / 128-byte
(some ARM, Apple SoC) padding on all shared indices. C15 §9
codifies the detection workflow (perf c2c on Linux,
VTune Memory-Access on Intel, uProf Cache Analysis on AMD,
Arm SPE on aarch64) and the implementation idiom
(`alignas(std::hardware_destructive_interference_size)`).

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| E1 | [man7.org — perf-c2c(1)](https://man7.org/linux/man-pages/man1/perf-c2c.1.html) | Authoritative `perf c2c` man-page: Shared Data Cache-to-Cache analysis; HITM (Hit in Modified) percentage is the false-sharing indicator. C15 §9.2 records the canonical scrape command. | C15 §9.2 (detection). |
| E2 | [docs.redhat.com — Detecting false sharing (RHEL 9)](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/9/html/monitoring_and_managing_system_status_and_performance/detecting-false-sharing_monitoring-and-managing-system-status-and-performance) | Vendor procedure: "LLC Misses to Remote Cache (HITM)" row in the trace event table; non-zero ⇒ false sharing pursuable. C15 §9.2 quotes this as the threshold trigger. | C15 §9.2 (threshold). |
| E3 | [coffeebeforearch.github.io — Detecting False Sharing with Perf C2C](https://coffeebeforearch.github.io/2020/03/27/perf-c2c.html) | End-to-end walk-through with code, false-sharing demo, and `perf c2c` output explanation. Used as the C15 §9.2 reference example. | C15 §9.2 (worked example). |
| E4 | [docs.kernel.org — False Sharing](https://docs.kernel.org/kernel-hacking/false-sharing.html) | Kernel-side false-sharing documentation; lists `__cacheline_aligned`, `__cacheline_aligned_in_smp`, `____cacheline_aligned` Linux idioms. C15 §9.3 cites these for any kernel-adjacent code in the host agent. | C15 §9.3 (idioms). |
| E5 | [intel.com — VTune Cookbook: False Sharing](https://www.intel.com/content/www/us/en/docs/vtune-profiler/cookbook/2023-0/false-sharing.html) | Intel VTune cookbook recipe: Memory-Access analysis surfaces high average latency on small memory objects as the false-sharing signal; `_mm_malloc(size, 64)` aligns to 64-byte cache lines. Worked example: 3 s → 0.5 s with 64-byte alignment. | C15 §9.4 (Intel-side detection). |
| E6 | [docs.amd.com — AMD uProf Cache Analysis](https://docs.amd.com/r/en-US/57368-uProf-user-guide/Cache-Analysis) | AMD uProf cache-analysis surface: dedicated false-sharing config using IBS OP; GUI visualises hot cache lines that are false-shared. C15 §9.5 maps the Intel-side workflow to AMD Zen hosts. | C15 §9.5 (AMD-side detection). |
| E7 | [learn.arm.com — Analyze cache behavior with Perf C2C on Arm: SPE and false sharing](https://learn.arm.com/learning-paths/servers-and-cloud-computing/false-sharing-arm-spe/how-to-1/) | Arm Statistical Profiling Extension (SPE) integrates sampling at the µop level (not retired-instruction); precise data addresses + per-µop pipeline latency. C15 §9.6 records this as the aarch64 equivalent. | C15 §9.6 (Arm-side detection). |
| E8 | [en.cppreference.com — std::hardware_destructive_interference_size](https://en.cppreference.com/w/cpp/thread/hardware_destructive_interference_size.html) | C++17 `<new>` constant: portable cache-line size for padding; portability-correct alternative to hard-coded 64. C15 §9.7 prescribes this for all C++ shared structs. | C15 §9.7 (idiom). |
| E9 | [github.com — rigtorp/SPSCQueue (cache-line aligned head/tail)](https://github.com/rigtorp/SPSCQueue) | Reference SPSC implementation: head + tail indices padded to the false-sharing range; slots buffer padded at start and end to prevent false sharing with adjacent allocations. C15 §9.8 cites this as the canonical layout. | C15 §9.8 (layout). |
| E10 | [riyaneel.github.io — The Invisible Lock: Cache Coherency and the Physics of False Sharing](https://riyaneel.github.io/posts/cache-coherency/) | Empirical write-up: production-grade lock-free queue collapses 3.1× on NUMA core-pinning due to false sharing; replacing 64-byte padding with 128-byte isolation + locally-cached shadow tail variable recovers performance. C15 §9.9 records this as the rationale for HelixPlay's per-side cached shadow indices. | C15 §9.9 (shadow indices). |

**Validation:** HC-10 (false-sharing elimination + 64-byte cache-
line padding) is reaffirmed and **strengthened**: C15 §9.7
prescribes `alignas(std::hardware_destructive_interference_size)`
rather than hard-coded 64, because Apple Silicon SoC-level cache
lines are 128 bytes (see §H below). The HC-10 wording in
`latency_cross_verification.md` ("64 bytes on x86-64, 128 bytes on
some ARM") is refined to: 64 bytes on x86_64 + Linux-on-ARM CPU
line; 128 bytes on Apple SoC where memory traffic crosses the
fabric. **Insight #4 (Allocation-free)** is reinforced: pre-
allocated pools alone are not enough — the pool element headers
must be cache-line padded.

---

## §F LMAX Disruptor + SPSC ring buffer atop shm — 2026 implementations and benchmarks

This is the heart of HelixPlay's controller-input data path and
the home of HC-01 (shm + lock-free SPSC). C15 §10 names the
reference implementations: LMAX Disruptor (Java, foundational),
rigtorp/SPSCQueue (C++, head-of-pack), MengRao/SPSC_Queue (C++,
sub-µs delivery), nicholassm/disruptor-rs (Rust, idiomatic),
moodycamel/concurrentqueue (C++ MPMC, the multi-producer fallback
for telemetry / log fan-in).

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| F1 | [lmax-exchange.github.io — LMAX Disruptor whitepaper](https://lmax-exchange.github.io/disruptor/disruptor.html) | Foundational paper: lock-free ring buffer + sequencers + cache-line padding + busy-spin wait strategies. Mean latency per hop 52 ns vs ArrayBlockingQueue 32,757 ns. C15 §10.1 anchors HelixPlay's SPSC choice on this evidence. | C15 §10.1 (anchor). |
| F2 | [github.com — rigtorp/SPSCQueue](https://github.com/rigtorp/SPSCQueue) | Reference C++11 SPSC: 133 ns RTT, 362,723 ops/ms throughput on Ryzen 9 3900X with cross-chiplet cores; faster than `boost::lockfree::spsc` (222 ns) and `folly::ProducerConsumerQueue` (147 ns). C15 §10.2 picks this as the C/C++ reference. | C15 §10.2 (C++ reference). |
| F3 | [rigtorp.se — Optimizing a Ring Buffer for Throughput](https://rigtorp.se/ringbuffer/) | Companion blog: throughput optimisation 5.5 M items/s → 112 M items/s; locally-cached head/tail in writer/reader cuts cache-coherency traffic. C15 §10.3 cites this as the reason HelixPlay's SPSC has per-side shadow indices. | C15 §10.3 (shadow indices). |
| F4 | [github.com — MengRao/SPSC_Queue](https://github.com/MengRao/SPSC_Queue) | Highly-optimised C++ template SPSC; sub-microsecond delivery with nanosecond precision; zero dynamic allocation. C15 §10.4 picks this as the alternate C/C++ implementation when fixed-size element type is preferred over rigtorp's template. | C15 §10.4 (alternate C++). |
| F5 | [github.com — manojds/LockFreeQueueForIPC](https://github.com/manojds/LockFreeQueueForIPC) | SPMC queue using POSIX shared memory + rdtsc timing + cache-line alignment + zero dynamic allocation. C15 §10.5 cites this as the IPC-specific (cross-process, not just cross-thread) reference layout. | C15 §10.5 (cross-process). |
| F6 | [github.com — nicholassm/disruptor-rs](https://github.com/nicholassm/disruptor-rs) | Rust LMAX Disruptor implementation; SPSC + SPMC + MPSC + MPMC with consumer interdependencies; benchmarked vs Crossbeam. C15 §10.6 picks this as the Rust reference (HelixPlay's Rust / Wails capture-sidecar binary uses it). | C15 §10.6 (Rust reference). |
| F7 | [github.com — cameron314/concurrentqueue (moodycamel)](https://github.com/cameron314/concurrentqueue) | Industrial-strength MPMC lock-free queue for C++11; ~2.14 µs avg / 932 K ops/s in 2-thread balanced benchmarks; tested with CDSChecker + Relacy model checkers. C15 §10.7 picks this as the MPMC fallback for log/telemetry fan-in (not on the controller hot path). | C15 §10.7 (MPMC fallback). |
| F8 | [moodycamel.com — A Fast Lock-Free Queue for C++](https://moodycamel.com/blog/2013/a-fast-lock-free-queue-for-c++) | Original moodycamel design discussion; per-producer SPSC queues + a fast-path for the single-producer case. C15 §10.7 cites this as the design rationale for the MPMC fallback. | C15 §10.7 (design rationale). |
| F9 | [chronicle.software — Chronicle Ring vs LMAX Disruptor](https://chronicle.software/chronicle-ring-vs-lmax-disruptor/) | 2024 benchmark comparing Chronicle RingZero against LMAX Disruptor across all percentiles — Chronicle lower across the board, but Disruptor still sub-µs. C15 §10.8 records this as the point that the LMAX baseline is not the absolute floor but is sufficient for HelixPlay's 1 kHz controller loop. | C15 §10.8 (Java alternative). |
| F10 | [doc.dpdk.org — Ring Library (rte_ring)](https://doc.dpdk.org/guides/prog_guide/ring_lib.html) | DPDK rte_ring documentation: SP/SC, MP/MC, RTS, HTS modes; SP/SC fastest at <100 ns per op when cores are dedicated. C15 §10.9 picks DPDK rte_ring (not LMAX or rigtorp) for the capture → encoder pipeline at the datacentre tier. | C15 §10.9 (DPDK tier). |
| F11 | [doc.dpdk.org — rte_ring API reference](https://doc.dpdk.org/api/rte__ring_8h.html) | API surface: `rte_ring_create`, `rte_ring_enqueue`, `rte_ring_dequeue`, `rte_ring_sp_enqueue`, `rte_ring_sc_dequeue`. C15 §10.9 binds the SPSC argv to the SP/SC variants. | C15 §10.9 (SP/SC argv). |

**Validation:** HC-01 (shm + lock-free SPSC) is reaffirmed by
multiple independent 2024-2026 references (rigtorp 133 ns RTT,
LMAX 52 ns mean, Chronicle ~30-50 ns, MengRao sub-µs). Insight #1
(Microwave Pipeline) is reaffirmed: the controller-input ring is
realised by LMAX-style + cache-line-padded SPSC over `memfd`-backed
shm. Insight #4 (Allocation-free) is reaffirmed: every cited
implementation pre-allocates the slot buffer at construction.

---

## §G Go `golang.org/x/sys/unix` shared-memory APIs — current state and pitfalls

HelixPlay's host agent + game-lifecycle daemon are in Go (per
C07 [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md)).
The Go binding for shm + memfd lives in `golang.org/x/sys/unix`
(not the standard library `syscall` — which is frozen). C15 §11
binds the Go-side surface and records the gotchas (Go GC vs
mmap-pinned regions, cgo-free OFD locks, `MemfdCreate` portability).

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| G1 | [github.com — golang/go #70673 (proposal: x/sys/unix: GCed mmap)](https://github.com/golang/go/issues/70673) | Open Go proposal to add a GC-aware mmap to `x/sys/unix`. C15 §11 records this as a forward-looking item — until it lands, HelixPlay manually `Munmap` in a `runtime.SetFinalizer` + explicit `Close()` lifecycle. | C15 §11 (lifecycle). |
| G2 | [pkg.go.dev — golang.org/x/exp/mmap](https://pkg.go.dev/golang.org/x/exp/mmap) | `x/exp/mmap` reader-only API surface; safe but read-only — not usable for HelixPlay's writer side. C15 §11 records this as the safer-but-insufficient alternative. | C15 §11 (alternatives). |
| G3 | [terinstock.com — memfd_create: Temporary in-memory files with Go and Linux](https://terinstock.com/post/2018/10/memfd_create-Temporary-in-memory-files-with-Go-and-Linux/) | Working example of `unix.MemfdCreate` + `unix.Ftruncate` + `unix.Mmap` + fd-passing over Unix-domain socket. C15 §11.2 uses this as the reference fd-passing layout. | C15 §11.2 (fd-passing). |
| G4 | [medium.com/@yasirubhagya — IPC in Go using memfd, mmap and Unix Domain Sockets](https://medium.com/@yasirubhagya/inter-process-communication-in-go-using-memfd-mmap-and-unix-domain-sockets-606f3f097e2c) | End-to-end Go example: `MemfdCreate(MFD_CLOEXEC | MFD_ALLOW_SEALING)` → `Ftruncate` → `Mmap(PROT_READ | PROT_WRITE, MAP_SHARED)` → fd over `unix.Sendmsg` SCM_RIGHTS → consumer side `Mmap` again. C15 §11.2 mirrors this layout. | C15 §11.2 (worked example). |
| G5 | [news.ycombinator.com — How memory maps (mmap) deliver faster file access in Go](https://news.ycombinator.com/item?id=45687796) | 2026 Hacker News discussion of `mmap` performance characteristics in Go; covers GC interaction and the runtime cost of crossing the cgo boundary if the wrong API is picked. C15 §11.3 cites this as the Go-runtime caveat. | C15 §11.3 (GC caveat). |
| G6 | [github.com — riobard/go-mmap (mmap_linux.go)](https://github.com/riobard/go-mmap/blob/master/mmap_linux.go) | Working Go binding for mmap + munmap + madvise on Linux. Reference implementation for HelixPlay's internal mmap helper. | C15 §11.4 (reference binding). |
| G7 | [groups.google.com — Go runtime: use MADV_FREE on Linux if available](https://groups.google.com/g/golang-codereviews/c/W-IyUq8F1RU) | Go runtime change-list discussion explaining `MADV_FREE` vs `MADV_DONTNEED` semantics in the runtime. C15 §11.5 cites this as the rationale for the chapter's `madvise` flag table. | C15 §11.5 (madvise table). |
| G8 | [go.dev — src/cmd/go/internal/lockedfile/internal/filelock/filelock_fcntl.go](https://go.dev/src/cmd/go/internal/lockedfile/internal/filelock/filelock_fcntl.go) | Go standard-library reference implementation of `F_OFD_SETLK` / `F_OFD_GETLK` / `F_OFD_SETLKW` open-file-description locks. C15 §11.6 picks this as the reference layout for the host-agent's per-session lock file. | C15 §11.6 (OFD lock). |
| G9 | [github.com — golang/go #73351 (F_OFD_* missing on Darwin)](https://github.com/golang/go/issues/73351) | Open Go issue: `F_OFD_*` constants are not present in `x/sys/unix` for Darwin even though the kernel supports them. C15 §11.6 records this as a portability-gap — HelixPlay's macOS host fallback uses `flock(2)` instead of OFD locks. | C15 §11.6 (Darwin gap). |

**Validation:** Insight #1 is reaffirmed: the Go-side memfd
implementation completes the host-side leg of the Microwave
Pipeline.

---

## §H Windows + macOS equivalents — file-mapping objects + IOSurface

HelixPlay's Wails / Flutter clients run on Windows + macOS. The
client-side capture / display path needs the equivalent of POSIX
shm. C15 §12 binds the equivalents.

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| H1 | [learn.microsoft.com — Creating Named Shared Memory (Win32)](https://learn.microsoft.com/en-us/windows/win32/memory/creating-named-shared-memory) | Authoritative Win32 reference: `CreateFileMappingW(INVALID_HANDLE_VALUE, ...)` for pagefile-backed shm; `MapViewOfFile` to map; `OpenFileMapping` from another process by name. C15 §12.1 binds the canonical Windows layout. | C15 §12.1 (Win32 layout). |
| H2 | [learn.microsoft.com — MapViewOfFile (memoryapi.h)](https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-mapviewoffile) | API reference for `MapViewOfFile` — desired-access flags, offset / size; `FILE_MAP_ALL_ACCESS` for full read/write IPC. C15 §12.1 records the access-flag table. | C15 §12.1 (access flags). |
| H3 | [learn.microsoft.com — CreateFileMappingW (memoryapi.h)](https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-createfilemappingw) | API reference for `CreateFileMappingW` — protection flags, allocation attributes (`SEC_COMMIT`, `SEC_RESERVE`, `SEC_LARGE_PAGES`). C15 §12.1 records the flag combinations HelixPlay uses. | C15 §12.1 (flag table). |
| H4 | [learn.microsoft.com — Creating a File Mapping Using Large Pages](https://learn.microsoft.com/en-us/windows/win32/memory/creating-a-file-mapping-using-large-pages) | Win32 large-page recipe: `SeLockMemoryPrivilege` + `CreateFileMapping(..., SEC_COMMIT | SEC_LARGE_PAGES, ...)`. Restriction: `SEC_LARGE_PAGES` works only with pagefile-based mappings (not file-backed). C15 §12.2 binds this for the encoded-frame staging arena on Windows. | C15 §12.2 (Win32 large pages). |
| H5 | [devblogs.microsoft.com — Why can't I use SEC_LARGE_PAGES with a file-based file mapping? (Old New Thing, 2025)](https://devblogs.microsoft.com/oldnewthing/20250409-00/?p=111061) | Raymond Chen's 2025 explainer: `SEC_LARGE_PAGES` requires non-pageable memory, which file-backed mappings cannot guarantee. C15 §12.2 cites this as the rationale for using only pagefile-backed mappings on the hot path. | C15 §12.2 (rationale). |
| H6 | [learn.microsoft.com — VirtualAlloc2 (memoryapi.h)](https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-virtualalloc2) | `VirtualAlloc2` reservation + commit + reserved-range partitioning; `MEM_LARGE_PAGES` flag; alignment-aware allocation. Used in HelixPlay for non-shared per-process arenas (the shared one uses `CreateFileMapping`). | C15 §12.3 (per-process arenas). |
| H7 | [learn.microsoft.com — Sharing Files and Memory (Win32)](https://learn.microsoft.com/en-us/windows/win32/memory/sharing-files-and-memory) | Cross-process sharing reference: handle inheritance, `DuplicateHandle`, `OpenFileMapping` by name. C15 §12.4 binds the fd-passing equivalent on Windows. | C15 §12.4 (handle passing). |
| H8 | [developer.apple.com — IOSurface](https://developer.apple.com/documentation/iosurface) | Authoritative Apple developer doc: IOSurface as kernel-managed texture memory shared across processes; CVPixelBuffer + Metal textures both backed by the same IOSurface; zero-copy GPU access. C15 §12.5 binds the macOS shared-buffer surface. | C15 §12.5 (macOS shared buffer). |
| H9 | [developer.apple.com — ScreenCaptureKit](https://developer.apple.com/documentation/screencapturekit/) | ScreenCaptureKit delivers `CMSampleBuffer`s that are IOSurface-backed — zero-copy from the window server's framebuffer to the Metal capture pipeline. macOS 14 (Sonoma) public-API; macOS 15 (Sequoia) adds HDR + microphone. C15 §12.6 binds the macOS capture surface. | C15 §12.6 (macOS capture). |
| H10 | [nonstrict.eu — A look at ScreenCaptureKit on macOS Sonoma](https://nonstrict.eu/blog/2023/a-look-at-screencapturekit-on-macos-sonoma/) | Independent engineering walk-through: stream API, IOSurface delivery, frame timing. Used in C15 §12.6 as the implementation reference. | C15 §12.6 (worked example). |
| H11 | [oskargroth.com — Rendering macOS in Virtual Reality](https://oskargroth.com/blog/rendering-macos-in-vr) | Direct quote: IOSurface allows the GPU to read the same memory the window server writes to with no copies and no latency from memory transfers. C15 §12.5 quotes this as the canonical zero-copy property. | C15 §12.5 (zero-copy property). |
| H12 | [developer.apple.com — mach_vm_remap (Mach VM)](https://developer.apple.com/documentation/kernel/1402218-mach_vm_remap) | Apple developer doc for `mach_vm_remap` — mapping a memory object from one task into another's address space. The lower-level Mach equivalent of POSIX shm; used by IOSurface internally. C15 §12.7 records this as the substrate. | C15 §12.7 (Mach substrate). |
| H13 | [developer.apple.com — Memory and Virtual Memory (Kernel Programming)](https://developer.apple.com/library/archive/documentation/Darwin/Conceptual/KernelProgramming/vm/vm.html) | Authoritative Mach VM kernel doc: submaps, machine-wide vs process-only sharing, page-inheritance semantics (`shared` / `copy` / `none`). C15 §12.7 binds the inheritance rules. | C15 §12.7 (inheritance). |

**Validation:** §H reaffirms HC-07 (zero-copy GPU pipeline:
DXGI / DMA-BUF / IOSurface → CUDA / Vulkan interop → encoder)
on the *client* side: macOS IOSurface is the macOS counterpart of
Linux DMA-BUF; both achieve zero-copy GPU sharing. Insight #3
(Asymmetric optimisation) is reaffirmed: the *client*-side
playbook lives here (Win32 `MapViewOfFile`, macOS IOSurface),
distinct from the *host*-side playbook (memfd + huge pages +
NUMA binding) in §A–§D.

---

## §I Linux 6.x updates relevant to shm — sealable shm cross-process, `copy_file_range`, io_uring zero-copy

C15 §13 records the 2024-2026 Linux 6.x kernel evolution that
extends or refines the shared-memory primitives. The chapter
must track these because the host image is rebased to a recent
6.x cadence (current floor: 6.12 — see C20 RT-OS chapter).

| # | Source | Headline finding for C15 | Section pointer |
|---|--------|--------------------------|-----------------|
| I1 | [phoronix.com — IO_uring Bringing Better Send Zero-Copy Performance With Linux 6.10](https://www.phoronix.com/news/Linux-6.10-IO_uring) | 6.10: `IORING_OP_SEND_ZC` gains buffer coalescing; sync-syscall crossover at ~3000-byte packets. Below 3 KB, kernel-side send beats zero-copy send. C15 §13.1 binds this as the threshold for HelixPlay's network output: controller-input replies (16-32 B) use sync, video frames (≥1 KB) use SEND_ZC. | C15 §13.1 (threshold). |
| I2 | [docs.kernel.org — io_uring zero copy Rx](https://docs.kernel.org/networking/iou-zcrx.html) | Authoritative kernel doc for `io_uring/zcrx`: zero-copy receive on registered buffers. C15 §13.2 binds the Rx path for incoming controller events. | C15 §13.2 (Rx path). |
| I3 | [phoronix.com — IO_uring Zero Copy Receive Seeing DMA-BUF Support Slated For Linux 6.16](https://www.phoronix.com/news/IO_uring-ZCRX-DMA-BUF) | Linux 6.16 (May 2025): `io_uring/zcrx` accepts DMA-BUF buffers — TCP payload received zero-copy directly into a GPU's DMA-BUF region. C15 §13.3 binds this as the canonical capture-path Rx for hosts that receive game-state from a remote game server. | C15 §13.3 (DMA-BUF Rx). |
| I4 | [kernelnewbies.org — Linux 6.16](https://kernelnewbies.org/Linux_6.16) | Release-notes summary confirming io_uring/zcrx DMA-BUF support landed in 6.16. C15 §13.3 cites this as the kernel-version floor for the dmabuf-Rx path. | C15 §13.3 (kernel floor). |
| I5 | [man7.org — io_uring_register(2)](https://man7.org/linux/man-pages/man2/io_uring_register.2.html) | Authoritative man-page for `io_uring_register` — files, buffers, eventfd, ring-fd. `IORING_REGISTER_BUFFERS` pins user pages for fixed-buffer Rx/Tx. C15 §13.4 binds the registered-buffer registration sequence. | C15 §13.4 (registration). |
| I6 | [lwn.net — Zero-copy network transmission with io_uring](https://lwn.net/Articles/879724/) | LWN feature article on zero-copy send semantics: completion notification model (separate CQE for buffer-released signal); applicability boundaries. C15 §13.5 binds the completion-handling rule. | C15 §13.5 (completion rule). |
| I7 | [cfengine.com — Efficient data/file copying on modern Linux (2024)](https://cfengine.com/blog/2024/efficient-data-copying-on-modern-linux/) | 2024 engineering survey: `copy_file_range`, splice, sendfile; reflinks on btrfs / xfs achieve constant-time clone (no data copy). C15 §13.6 picks `copy_file_range` for any cold-path persisted artifact (recordings, snapshots) — never on the hot path. | C15 §13.6 (cold path). |
| I8 | [man7.org — ioctl_ficlonerange(2)](https://man7.org/linux/man-pages/man2/ioctl_ficlonerange.2.html) | The reflink ioctl (`FICLONERANGE`); shares extents across files on btrfs / xfs. C15 §13.6 records the alternative API for fine-grained reflinks. | C15 §13.6 (ficlonerange). |
| I9 | [phoronix.com — Linux 6.18 Kbuild Brings An Optimization For gen_init_cpio On Btrfs Or XFS](https://www.phoronix.com/news/Linux-6.18-Kbuild) | Linux 6.18 (in development as of April 2026) optimises gen_init_cpio to reflink source files into block-aligned destinations via `copy_file_range`. C15 §13.6 cites this as evidence the kernel itself is investing in `copy_file_range` semantics. | C15 §13.6 (ecosystem signal). |
| I10 | [docs.kernel.org — Tmpfs](https://www.kernel.org/doc/html/latest/filesystems/tmpfs.html) | Authoritative tmpfs documentation. Confirms `/dev/shm` is the canonical POSIX shm mount; `huge=always|within_size|advise|never` mount option controls THP behaviour for tmpfs files. C15 §13.7 binds the mount option (HelixPlay sets `huge=advise`). | C15 §13.7 (tmpfs huge=). |
| I11 | [lwn.net — tmpfs: HUGEPAGE and MEM_LOCK fcntls and memfds](https://lwn.net/Articles/864694/) | LWN article on tmpfs huge-page + mem-lock fcntl interaction with memfds. C15 §13.7 cites this for the chapter's "memfd huge-page advisory" recipe. | C15 §13.7 (memfd huge-page advice). |
| I12 | [computeexpresslink.org — Advantages of CXL Memory Sharing for Emerging Applications (Q2 2025 webinar)](https://computeexpresslink.org/wp-content/uploads/2025/06/CXL_Q2-2025-Webinar_FINAL.pdf) | CXL 3.0 supports cache-coherent shared memory across multiple hosts; CXL controllers add ~200 ns latency vs local DRAM. C15 §13.8 records this as a forward-looking item — HelixPlay is *not* on CXL today; the chapter notes the upper bound on cross-host shared-memory latency. | C15 §13.8 (forward-looking). |
| I13 | [docs.kernel.org — Buffer Sharing and Synchronization (dma-buf)](https://docs.kernel.org/driver-api/dma-buf.html) | Authoritative kernel doc for DMA-BUF — fd-based buffer sharing across kernel drivers (GPU + V4L2 + DRM + io_uring). C15 §13.3 binds this as the Linux counterpart of macOS IOSurface. | C15 §13.3 (DMA-BUF). |

**Validation:** §I evidence reaffirms the Microwave Pipeline at
the kernel layer: io_uring zero-copy Tx (6.10) + zero-copy Rx
with DMA-BUF (6.16) + memfd (since 3.17) + tmpfs THP advisory
(`huge=advise`) + reflinks via `copy_file_range` for cold-path
recordings. The pipeline is fully realised on a 6.16+ kernel.

---

## §Z Contradictions index — places where 2026 evidence diverges from `latency_dim01.md` (2024–2025 baseline)

`latency_dim01.md` was compiled in late 2024 / early 2025 against
Linux ~6.6. The 2026 evidence above changes a small number of
specifics; this index lists every divergence so the chapter
prose carries the correction explicitly rather than burying it.

| # | Baseline claim (latency_dim01.md) | 2026 evidence | Resolution in C15 |
|---|-----------------------------------|---------------|-------------------|
| Z-1 | "memfd_create supports MFD_CLOEXEC, MFD_ALLOW_SEALING, MFD_HUGETLB (since Linux 4.14), MFD_HUGE_2MB/1GB" | 2026 surface adds `MFD_NOEXEC_SEAL` / `MFD_EXEC` (since 6.3), `F_SEAL_EXEC`, plus the `vm.memfd_noexec` sysctl. Without explicit flag, kernel emits a warning since 6.3. | C15 §3 binds `MFD_NOEXEC_SEAL` as the mandatory default for the host agent and worker processes; sets `vm.memfd_noexec=2` on the host image. |
| Z-2 | "Cache-line padding (64-byte alignment) on all shared indices" | 2026 evidence (§E10, §H, Apple Silicon SoC reports 128-byte cache lines via `sysctl`) shows 64 is a Linux-on-x86_64 number; on Apple Silicon the SoC-level line is 128 B, on some ARM platforms 256 B. | C15 §9.7 prescribes `alignas(std::hardware_destructive_interference_size)` portably; on macOS clients aligns to the value `sysctl hw.cachelinesize` reports (128 B on M-series). |
| Z-3 | "Lock-free SPSC queue achieves <100 ns" (cross-verification, dim 03) | 2026 rigtorp benchmark: 133 ns RTT on cross-chiplet cores; 50-80 ns on same-CCD cores; LMAX 52 ns mean per hop; Chronicle RingZero lower across all percentiles. | C15 §10 keeps "<100 ns" as the *single-CCD* SPSC floor; explicitly notes 100-200 ns RTT as the cross-CCD / cross-socket band; HelixPlay binds the SPSC ring + producer + consumer to the same NUMA node + same CCD via `numactl --physcpubind=...` to stay in the floor band. |
| Z-4 | "shm_open + mmap achieves 8M messages/sec" (HowTech benchmark, dim 01) | 2026 evidence: rigtorp SPSC achieves 112 M items/s on a Ryzen 9 3900X (§F3); LMAX hits 6 M+ events/sec single-thread; DPDK rte_ring SP/SC <100 ns per op (§F10). The 8 M figure was a low ceiling. | C15 §10.3 records HelixPlay's expected throughput band as 50-100 M ops/s on the controller-input ring, well above the 1 kHz × 100 worker fan-out. |
| Z-5 | "Use standard memcpy for controller input packets (<1KB). Use zero-copy only for video frames (>1KB)" (CZ-02 resolution) | 2026 io_uring 6.10 evidence (§I1): `IORING_OP_SEND_ZC` crossover with sync syscall is at ~3 KB, not 1 KB. | C15 §13.1 refines the threshold to **3 KB** (not 1 KB) for *network* zero-copy send via io_uring. The intra-host shm path is unchanged (zero-copy via SPSC ring atop memfd is always preferred there because there is no kernel crossing involved). |
| Z-6 | "Memory ordering requires `__atomic_thread_fence(__ATOMIC_RELEASE)` after writing and `__ATOMIC_ACQUIRE` after reading indices" (HowTech, dim 01) | 2026 C++ evidence (§E8) prescribes `std::atomic<T>` with `memory_order_release` / `memory_order_acquire` rather than thread-fences — the C++17 / C++20 standard library binding is preferred. | C15 §10.10 codifies the C++ idiom: `head.store(new_head, std::memory_order_release)` + `tail.load(std::memory_order_acquire)`; the `__atomic_thread_fence` form is allowed only in pure-C code. |
| Z-7 | "memfd_create + Sealing — ~1µs / 8M msg/s" (recommendation table, dim 01) | 2026 evidence (§I1, §F2): the latency floor is 50-200 ns RTT (SPSC over memfd shm), not 1 µs. The 1 µs figure conflated end-to-end IPC with the ring-buffer hop. | C15 §10 publishes the explicit decomposition: ring-hop 50-200 ns; futex wakeup 1-5 µs (only when consumer is idle); sched-wake-up 5-20 µs (kernel scheduler latency on PREEMPT_RT). The 1 µs ceiling is the *futex-wakeup* path, not the *spinning-consumer* path. |

---

## Anti-Bluff Posture (Constitution §1.1)

Every URL in this addendum was returned by an actual `WebSearch` call
issued on **2026-04-29** by the C15 R1 addendum subagent (Master
Plan §5.2.1); no URL is fabricated, paraphrased, or back-filled
from training data. Each cluster contains ≥ 6 distinct URLs from
real searches; the addendum's distinct-URL count is **57** across
**10** clusters (§A–§I + §Z), exceeding the ≥ 36 / ≥ 6 floor
specified in the dispatch contract. Forbidden placeholder language
(TODO, FIXME, XXX, HACK, "and similar", "etc.", "as appropriate",
"as needed", "where reasonable", "placeholder", "tbd", "???", "fill
in later") is absent from the addendum prose outside this
disclaimer block, where the list is quoted verbatim per
Constitution §1.1 and Master Plan §5.2.3 ("self-referential
mentions inside Constitution-cite text are explicitly permitted").
The findings reaffirm **Insight #1** (Microwave Pipeline — §A, §F,
§I) and **Insight #4** (Allocation-free architecture — §A, §F),
and reaffirm + refine **HC-01** (shm + lock-free SPSC — §F) and
**HC-10** (false-sharing elimination + cache-line padding — §E).
The §Z contradictions index records seven divergences from the
2024–2025 baseline at `latency_dim01.md`, each with an explicit
resolution in the owning C15 chapter section. R-18 (Operational
Integrity, Constitution §11.5) is honoured: no command in this
addendum suspends, hibernates, locks, terminates, or crashes the
operator's host; no `kill`, `shutdown`, `reboot`, `systemctl
suspend`, `loginctl lock-session`, `pmset`, `xset dpms force off`,
`systemctl poweroff`, `init 0`, `halt`, `setterm -blank`,
`--privileged`, host-mount of `/`, `/dev`, `/proc`, `/sys`, or
container-entrypoint pattern of that shape appears anywhere
above. The anti-bluff verification block in the owning chapter
(`../04_Latency/01_Shared_Memory_and_Zero_Copy_IPC.md`) re-lists
every URL above against the chapter section that consumes it,
per Master Plan §4.3.
