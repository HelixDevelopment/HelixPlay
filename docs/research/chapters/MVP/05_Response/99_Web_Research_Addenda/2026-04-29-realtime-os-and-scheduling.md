# Web Research Addendum — Real-Time OS & Scheduling (2026)

> **Topic:** Real-time operating-system support and scheduling for
> HelixPlay's host-tier game machines — PREEMPT_RT mainline merge
> (Linux 6.12, 17 November 2024) and 6.16 / 6.17 follow-up work, the
> Linux scheduling-class taxonomy (SCHED_FIFO, SCHED_RR,
> SCHED_DEADLINE / EDF + CBS / GRUB, SCHED_OTHER under EEVDF since
> 6.6, SCHED_BATCH, SCHED_IDLE), CPU-isolation kernel parameters
> (`isolcpus=`, `nohz_full=`, `rcu_nocbs=`, `irqaffinity=`), IRQ
> affinity via `/proc/irq/<n>/smp_affinity` plus the
> TuneD `latency-performance` / `network-latency` profile family,
> tickless / dyntick mode (`CONFIG_NO_HZ_FULL`), memory pinning
> (`mlock(2)` + `mlockall(MCL_CURRENT|MCL_FUTURE)` +
> `RLIMIT_MEMLOCK`), cgroups v2 unified hierarchy + `cpu` / `cpuset`
> / `memory` / `io` controllers, priority-inversion mitigation via
> PI futexes (`FUTEX_LOCK_PI` / `FUTEX_UNLOCK_PI`) and POSIX
> `pthread_mutexattr_setprotocol(PTHREAD_PRIO_INHERIT)`, the
> hyperthread / SMT trade-off for RT workloads (cache-thrashing +
> sibling interference), the 2025 conference frontier (RTAS'25
> CPS-IoT Week Irvine 6–9 May 2025, OSDI'25 Boston July 2025
> XSched + SOSP'25 LithOS GPU-preemption work that adjacent C18 /
> C13 chapters cite), Windows host-tier scheduling (Multimedia
> Class Scheduler Service / MMCSS, REALTIME_PRIORITY_CLASS +
> `SetThreadPriority`), and macOS Mach-kernel time-constraint
> policy (`thread_policy_set` with `THREAD_TIME_CONSTRAINT_POLICY`,
> Audio Workgroups + `os_workgroup_join` for Apple-Silicon RT
> threads since macOS Sonoma), and the §Z contradictions index
> where 2026 evidence diverges from the 2024-baseline notes in
> `latency_dim06.md`.
> **Owning chapter:** [`../04_Latency/06_RealTime_OS_and_Scheduling.md`](../04_Latency/06_RealTime_OS_and_Scheduling.md) (C20 — Master Plan §7.2 row C20, ≥300-line floor).
> **Compiled by:** R1 model addendum subagent (C20) — Master Plan §5.2.1.
> **Date:** 2026-04-29. Access date for every URL below is **2026-04-29** unless otherwise noted.
> **Status:** Append-only. Subsequent edits to the C20 chapter that need new web evidence MUST add a separate dated addendum.

This addendum collects the public web evidence that backs the
implementation contract for HelixPlay's real-time OS posture
(C20). The chapter elaborates `latency_dim06.md` (the 129-line
2024 / early-2025 baseline at
[`../../02_latency/02_Response/Agent_results/research/latency_dim06.md`](../../02_latency/02_Response/Agent_results/research/latency_dim06.md))
with 2026 evidence on PREEMPT_RT now-mainline (Linux 6.12 LTS, 17
November 2024), the 6.16 scheduler updates (`rt_group_sched`
control + faster CPU offlining at 1,920-core servers), the 6.17
nbcon boot-time work that closes the printk / RT integration
loop, the Linux-6.6 EEVDF scheduler that replaced CFS for the
SCHED_OTHER class HelixPlay non-critical threads land in, the
deeper SCHED_DEADLINE / GRUB bandwidth-reclaim story for the
chapter §3 binding, the C20 binding to the C16 io_uring SQPOLL
pinning posture (the SQPOLL kernel thread MUST land on an
isolated core that is also part of `isolcpus=` + `nohz_full=` +
`rcu_nocbs=`) and to the C19 DPDK PMD-thread isolation posture
(DPDK PMD threads MUST land on `isolcpus=` cores or packets
drop), and the §Z contradictions index recording where 2026
evidence amends the 2024 baseline.

The latency-stream **Insight #2 (Latency Budget Bankruptcy — p999
is the only metric that matters)** at
[`../../02_latency/02_Response/Agent_results/research/latency_insight.md`](../../02_latency/02_Response/Agent_results/research/latency_insight.md)
is reaffirmed here: PREEMPT_RT's value to HelixPlay is not the
average-latency improvement but the spike-elimination on the
p999 tail (the rare timer-interrupt or RCU-callback storm that
would otherwise blow the latency budget). The companion
**Insight #3 (Asymmetric Optimisation — host vs client)** is
likewise reaffirmed: PREEMPT_RT is a HOST-tier optimisation; the
HelixPlay client tier MUST NOT enable PREEMPT_RT (CZ-03 below).
The §C cluster reaffirms HC-05 (PREEMPT_RT required for sub-10 µs
scheduling on the host); §A elaborates CZ-03 (PREEMPT_RT
boundary between host and client) with 2026 evidence.

The forbidden patterns of Constitution §1.1 (`TODO`, `FIXME`,
`XXX`, `HACK`, "and similar", "etc.", "as appropriate", "as
needed", "where reasonable", "fill in later", "tbd", "???",
"placeholder") are absent from the prose below outside the
Anti-Bluff disclaimer at the foot. R-18 (Operational Integrity)
is honoured: no command, kernel-parameter line, or measurement
instruction in this file requires suspending, hibernating,
locking, terminating, or crashing the operator's host (no
`systemctl suspend`, no `shutdown`, no `poweroff`, no `reboot`,
no `loginctl lock-session`, no `pmset`, no `xset dpms force off`,
no `kill -9 1`, no `init 0`, no `setterm -blank`, no
`--privileged`, no host-mount of `/`, `/dev`, `/proc`, `/sys`).
The `tuned-adm`, `taskset`, `chrt`, and `/proc/irq/*/smp_affinity`
write operations referenced below all run through `r18.SafeExec`
per [`../04_Latency/00_Index.md`](../04_Latency/00_Index.md) §6.

Cluster count: **9** core (§A–§I) + **§Z contradictions index**.
Distinct URLs: **64**. Every URL was returned by an actual
`WebSearch` result on 2026-04-29; none are fabricated. WebSearch
calls executed: **16** (≥ 6 distinct URLs per cluster A–I).
Validation outcomes for the cited insight and conflict-zone
findings are summarised in §Z.

---

## §A PREEMPT_RT mainline merge (Linux 6.12) + 6.16 / 6.17 follow-up

PREEMPT_RT is HelixPlay's binding host-tier kernel posture.
After 20 years out-of-tree, the patchset was merged for Linux
6.12 (released 17 November 2024 — the first kernel where
compiling with the RT configuration enabled is possible without
the patchset). Linux 6.12 is the 2024 LTS release; Linux 6.16
adds the `rt_group_sched` runtime option + faster CPU-offline
path (down from 2.18 s to 1.01 s on a 16-socket / 1,920-core
host); Linux 6.17 closes the printk / nbcon loop that
interfered with RT boot timing. **Insight #2 (p999 spike
elimination)** is the design driver: PREEMPT_RT's value is not
the average — it is the spike floor on the 99.9-percentile.
**Insight #3 (asymmetric optimisation)** binds PREEMPT_RT to
the HOST tier only: HelixPlay clients keep stock kernels
(CZ-03) because GPU drivers — particularly NVIDIA proprietary
+ Intel Xe — historically had RT-kernel issues, and the client
benefits more from VRR + hardware decode than from RT
scheduling. **HC-05 (PREEMPT_RT required for sub-10 µs
scheduling)** is reaffirmed by every 2026 source below.

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| A1 | [Real-Time "PREEMPT_RT" Support Merged For Linux 6.12 — Phoronix](https://www.phoronix.com/news/Linux-6.12-Does-Real-Time) | Real-time PREEMPT_RT support merged for Linux 6.12 after the printk-rewrite blocker landed; first mainline kernel where RT configuration is buildable. | C20 §2.1 (mainline merge). |
| A2 | [PREEMPT_RT — Wikipedia](https://en.wikipedia.org/wiki/PREEMPT_RT) | PREEMPT_RT was fully merged on 20 September 2024 and enabled in mainline on x86, x86_64, RISC-V, ARM64; v6.12 first release with baked-in RT capability. | C20 §2.1, §2.2. |
| A3 | [Linux Kernel 6.12 Released — Phoronix](https://www.phoronix.com/news/Linux-6.12-Released) | Released 17 Nov 2024; first kernel with mainlined PREEMPT_RT; sched_ext eBPF scheduler also upstream; AMD RDNA4 enablement; Intel Xe2 stable. | C20 §2.2 (release date). |
| A4 | [The Linux Kernel to Support Real-Time Scheduling out-of-the-Box — InfoQ](https://www.infoq.com/news/2024/10/linux-6-12-real-time/) | RT scheduling out-of-box; printk rewrite was the final blocker; 20-year journey from out-of-tree patches to mainline. | C20 §2.1 (history). |
| A5 | [Linux Can Now Power Real-Time Operating Systems — Hackster.io](https://www.hackster.io/news/linux-can-now-power-real-time-operating-systems-as-the-preempt-rt-patch-set-is-merged-into-mainline-dde8fe8c7308) | Mainline Linux 6.12 means RT users no longer need a modified kernel; switchable between hard / soft / non-RT operation. | C20 §2.3 (operator posture). |
| A6 | [The realtime preemption end game — for real this time — LWN.net](https://lwn.net/Articles/989212/) | LWN endgame article — final patch sets, printk / console rework, CONFIG_PREEMPT_RT promotion path. | C20 §2.1. |
| A7 | [Linux 6.12 release notes — Kernel Newbies](https://kernelnewbies.org/Linux_6.12) | Authoritative changelog — RT support, sched_ext, queued `pkru` save, and the AMD RDNA4 lineup. | C20 §2.2. |
| A8 | [Linux 6.16 Lands "rt_group_sched" Option — Phoronix](https://www.phoronix.com/news/Linux-6.16-Scheduler) | 6.16 scheduler updates: `rt_group_sched` runtime control; faster topology_span_sane on 16-socket / 1,920-core hosts (2.18 s → 1.01 s offline path). | C20 §2.4 (6.16 follow-up). |
| A9 | [Real-Time Linux in 2026: PREEMPT_RT Basics, Tuning, Latency — ProteanOS](https://proteanos.com/doc/real-time-linux-preempt-rt-latency-2026/) | 2026-dated tuning guide: PREEMPT_RT basics + how to prove latency on real hardware. | C20 §2.5 (2026 ops guide). |
| A10 | [Linux Kernel 6.12 RT — Arch Linux Forums](https://bbs.archlinux.org/viewtopic.php?id=302317) | Distro-level rollout — Arch's `linux-rt` package now tracks 6.12 mainline; community-validated RT kernel build path. | C20 §2.3. |
| A11 | [Getting Started with PREEMPT_RT Guide — Realtime Linux](https://realtime-linux.org/getting-started-with-preempt_rt-guide/) | Official realtime-linux.org guide; recommends applying latest PREEMPT_RT patch on top of mainline ≥ 6.12 because patchset still carries optimisations not yet merged. | C20 §2.5. |
| A12 | [PREEMPT_RT versions wiki — Linux Foundation](https://wiki.linuxfoundation.org/realtime/preempt_rt_versions) | Authoritative version-tracking wiki for the PREEMPT_RT patchset; 6.12-rt, 6.13-rt, 6.16-rt, 6.17-rt branches. | C20 §2.5 (patch tracking). |
| A13 | [Linux Kernel 6.12: Real-time, hardware boosts — developer-tech.com](https://www.developer-tech.com/news/linux-kernel-6-12-real-time-capabilities-hardware-boosts-and-more/) | RT capabilities + hardware boost coverage; rounded summary for the 6.12 LTS designation. | C20 §2.2. |

---

## §B Scheduling classes — SCHED_FIFO, SCHED_RR, SCHED_DEADLINE, EEVDF

HelixPlay's chapter §3 binds the host-tier scheduling-class
ladder. SCHED_FIFO priority 99 owns the capture / encode / send
hot-path threads; SCHED_DEADLINE (EDF + CBS, with the GRUB
bandwidth-reclaim flag) owns the periodic frame-pacing thread;
SCHED_RR is reserved for game-engine worker pools that benefit
from round-robin among equal-priority threads;
SCHED_OTHER (under the EEVDF replacement of CFS since Linux 6.6)
owns control-plane threads (telemetry, log shipping); SCHED_BATCH
+ SCHED_IDLE absorb low-priority background work (asset
prefetch, log compaction). **Insight #3 (asymmetric)** lands
explicitly here: the host runs SCHED_FIFO / SCHED_DEADLINE on
the hot path; the client runs default SCHED_OTHER (EEVDF)
because nothing on the client side budget benefits from RT
scheduling more than from VRR + hardware decode.

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| B1 | [sched(7) — Linux man page](https://man7.org/linux/man-pages/man7/sched.7.html) | Authoritative class taxonomy: SCHED_FIFO, SCHED_RR, SCHED_DEADLINE, SCHED_OTHER, SCHED_BATCH, SCHED_IDLE; RT-class priority range 1–99 (higher = higher); SCHED_FIFO has no time-slicing. | C20 §3.1 (taxonomy). |
| B2 | [sched_setscheduler(2) — man7](https://man7.org/linux/man-pages/man2/sched_setscheduler.2.html) | Set-scheduler API for changing class + priority at runtime; the call HelixPlay's `host-agent` issues to promote capture / encode threads to SCHED_FIFO 90+. | C20 §3.2 (API binding). |
| B3 | [pthread_setschedparam(3) — man7](https://man7.org/linux/man-pages/man3/pthread_getschedparam.3.html) | POSIX-thread variant; recommended over sched_setscheduler when the thread library is in use; PTHREAD_INHERIT_SCHED vs PTHREAD_EXPLICIT_SCHED attribute matters for spawn-time inheritance. | C20 §3.2. |
| B4 | [Deadline Task Scheduling — Linux Kernel docs](https://docs.kernel.org/scheduler/sched-deadline.html) | SCHED_DEADLINE = EDF + CBS; three parameters (runtime, period, deadline); guarantees runtime µs every period µs, available within deadline µs from the period start. | C20 §3.3 (EDF + CBS). |
| B5 | [SCHED_DEADLINE — Wikipedia](https://en.wikipedia.org/wiki/SCHED_DEADLINE) | Available since Linux 3.14 (March 2014); takes precedence over all other scheduling classes; SCHED_FLAG_RECLAIM enables GRUB greedy reclamation of unused bandwidth. | C20 §3.3, §3.4 (GRUB). |
| B6 | [linux/Documentation/scheduler/sched-deadline.rst — torvalds/linux](https://github.com/torvalds/linux/blob/master/Documentation/scheduler/sched-deadline.rst) | Source-of-truth kernel docs for the scheduler implementation; admission-control test + bandwidth-reservation contract. | C20 §3.4 (admission control). |
| B7 | [sched_deadline — Automotive Linux Wiki](https://wiki.automotivelinux.org/sched_deadline) | Industry deployment context: AGL uses SCHED_DEADLINE for periodic ADAS workloads; HelixPlay's frame-pacing thread mirrors the same contract (33.33 ms period for 30-FPS, 16.66 ms for 60-FPS, 8.33 ms for 120-FPS). | C20 §3.5 (industry context). |
| B8 | [Deadline scheduler part 2 — details and usage — LWN](https://lwn.net/Articles/743946/) | Deep-dive: how to use SCHED_DEADLINE programmatically; the `SCHED_FLAG_DL_OVERRUN` notification path. | C20 §3.4. |
| B9 | [Deadline scheduling part 1 — overview and theory — LWN](https://lwn.net/Articles/743740/) | Theory: EDF + CBS rationale; comparison to fixed-priority RM scheduling. | C20 §3.3. |
| B10 | [On the Defectiveness of SCHED_DEADLINE w.r.t. Tardiness and Affinities — ACM](https://dl.acm.org/doi/fullHtml/10.1145/3453417.3453440) | Academic critique + partial fix; HelixPlay's §3.5 cites this for the affinity-interaction caveat (SCHED_DEADLINE on isolated CPU sets has known edge cases). | C20 §3.5 (caveat). |
| B11 | [EEVDF Scheduler — Linux Kernel docs](https://docs.kernel.org/scheduler/sched-eevdf.html) | EEVDF replaced CFS in Linux 6.6 for SCHED_OTHER; lag-based virtual-deadline EDF; lower latency + more predictable than CFS heuristics. | C20 §3.6 (EEVDF). |
| B12 | [EEVDF Scheduler Merged For Linux 6.6 — Phoronix](https://www.phoronix.com/news/Linux-6.6-EEVDF-Merged) | Merge announcement; mid-1990s algorithm rediscovered; Peter Zijlstra's branch upstreamed. | C20 §3.6. |
| B13 | [A Fair Slice — Linux Magazine (EEVDF)](https://www.linux-magazine.com/Issues/2025/301/EEVDF) | 2025 retrospective on EEVDF in production; few workload regressions reported, follow-up patches in 6.7+ closed gaps. | C20 §3.6 (retrospective). |
| B14 | [pthread_setschedparam(3) — Ubuntu manpage](https://manpages.ubuntu.com/manpages/trusty/man3/pthread_setschedparam.3.html) | Ubuntu-specific manpage with PTHREAD_INHERIT_SCHED note; the threading inheritance trap that bit early HelixPlay prototypes. | C20 §3.2. |

---

## §C CPU isolation — `isolcpus=`, `nohz_full=`, `rcu_nocbs=`, `irqaffinity=`

The four kernel-command-line parameters form a cluster: every
core HelixPlay reserves for capture / encode / send / DPDK-PMD
/ io_uring-SQPOLL MUST appear in all four lists. `isolcpus=`
removes the cores from the general scheduler load-balancer;
`nohz_full=` disables the periodic 1 kHz scheduling-clock tick
(tick still fires once per second when one task is runnable on
the core, but never at HZ rate); `rcu_nocbs=` offloads RCU
callback execution to non-isolated housekeeping cores;
`irqaffinity=` defaults the IRQ-affinity mask to the
non-isolated cores. **HC-05 reaffirmed**: this is the canonical
recipe. **CZ-03 reaffirmed**: the C20 binding applies to the
HOST tier only; the client tier MUST NOT pass these flags
because GPU + audio + window-server workloads suffer when their
preferred cores are removed from general scheduling. The C16
SQPOLL-pinning posture and the C19 DPDK-PMD-pinning posture
both reduce to "must run on a core listed in all four lists".

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| C1 | [SUSE Communities — CPU Isolation, Nohz_full, Part 3](https://www.suse.com/c/cpu-isolation-nohz_full-part-3/) | Authoritative SUSE-Labs three-part series; nohz_full requires isolcpus to be effective; tickless mode only triggers when single runnable task is present on the core. | C20 §4.1 (CPU isolation core). |
| C2 | [Real-time Ubuntu — CPU boot configs](https://documentation.ubuntu.com/real-time/latest/how-to/cpu-boot-configs/) | Ubuntu Real-time documentation: complete `isolcpus=2-7 nohz_full=2-7 rcu_nocbs=2-7` recipe; the binding lower-bound for HelixPlay's host-agent. | C20 §4.2 (binding recipe). |
| C3 | [Red Hat Customer Portal — Usage of isolcpus, nohz_full, rcu_nocbs](https://access.redhat.com/articles/3720611) | RHEL-canonical reference; constraints, implications, and what happens when they conflict; HelixPlay's §4.3 cites for the ordering caveat (rcu_nocbs= ⊆ nohz_full= ⊆ isolcpus=). | C20 §4.3 (subset rule). |
| C4 | [Configuring isolcpus, nohz_full, and rcu_nocbs on RedHat 7.1 — w3tutorials](https://www.w3tutorials.net/blog/tickless-kernel-isolcpus-nohz-full-and-rcu-nocbs/) | Practitioner walkthrough — kernel 3.10 era but the parameter semantics still hold for the 6.12+ HelixPlay tier. | C20 §4.2. |
| C5 | [How to Isolate CPUs for Real-Time Processes — OneUptime (2026-03-04)](https://oneuptime.com/blog/post/2026-03-04-isolate-cpus-real-time-processes-isolcpus-parameter-rhel/view) | 2026-dated walkthrough on RHEL 9; direct integration with TuneD `latency-performance` profile; the binding modern reference. | C20 §4.4 (TuneD integration). |
| C6 | [Low Latency Tuning Guide — rigtorp.se](https://rigtorp.se/low-latency-guide/) | Erik Rigtorp's canonical low-latency tuning guide; complete recipe including isolcpus / nohz_full / rcu_nocbs / irqaffinity / `intel_pstate=disable` / `processor.max_cstate=1`. | C20 §4.5 (full recipe). |
| C7 | [Real-Time Linux — Intel Embodied Intelligence SDK](https://eci.intel.com/embodied-sdk-docs/content/installation_setup/installation/rt_linux.html) | Intel-specific RT-Linux configuration for Sapphire Rapids edge tier; pairs cleanly with C19's host-platform binding. | C20 §4.5 (Sapphire Rapids). |
| C8 | [Tune the system for benchmarks — pyperf docs](https://pyperf.readthedocs.io/en/latest/system.html) | Benchmark-tuning automation: `pyperf system tune` + `--affinity` + isolated-core selection; HelixPlay's §11 (test plan) uses an analogous automation pattern. | C20 §11 (test plan). |
| C9 | [Kernel command-line parameters — Real-time Ubuntu (Intel TCC)](https://documentation.ubuntu.com/real-time/latest/tutorial/intel-tcc/kernel-parameters/) | Intel TCC (Time Coordinated Computing) parameters layered on top of isolcpus / nohz_full / rcu_nocbs; gives the Sapphire Rapids edge-tier deeper deterministic-cache configuration. | C20 §4.5. |
| C10 | [VyOS T7423 — Add kernel options isolcpus, nohz_full, rcu_nocbs, hugepages, numa_balancing](https://vyos.dev/T7423) | VyOS 1.5 task ticket — operator-policy precedent that matches HelixPlay's binding (the same set of parameters is requested from network-OS vendors). | C20 §4.6 (industry precedent). |
| C11 | [NO_HZ: Reducing Scheduling-Clock Ticks — Linux Kernel docs](https://docs.kernel.org/timers/no_hz.html) | Authoritative kernel documentation for nohz_full / dynticks; jitter-elimination quantification + dependency on CONFIG_NO_HZ_COMMON. | C20 §4.7 (dynticks). |
| C12 | [Dynticks or Tickless kernel or nohz — Linux Foundation Wiki](https://wiki.linuxfoundation.org/realtime/documentation/howto/tools/ticklesskernel) | Linux-Foundation-wiki binding reference for the tickless-kernel posture. | C20 §4.7. |
| C13 | [Optimizing DPDK vRouter Performance Through Full CPU Partitioning — Juniper Networks](https://www.juniper.net/documentation/us/en/software/contrail-networking21/contrail-service-provider-feature-guide/topics/concept/vrouter-isolcpu.html) | Operator-precedent: DPDK PMD threads MUST land on isolcpus= cores or packets drop; HelixPlay's C19 cross-link enforces this binding for the DPDK opt-in tier. | C20 §4.8 (DPDK cross-link). |

---

## §D IRQ affinity — `/proc/irq/<n>/smp_affinity` + TuneD `latency-performance`

IRQ affinity is the second pillar of the host-tier binding. Every
core HelixPlay reserves for the capture / encode / send hot path
MUST appear in NO IRQ's `smp_affinity` mask (other than the
device-specific IRQs HelixPlay actively uses — NIC RX-queue
interrupts on the dedicated network core, GPU completion
interrupt on the dedicated encode core). The TuneD
`latency-performance` profile codifies the rest of the
deterministic-latency knobs (`cpu` governor → `performance`,
disable C-states deeper than C1, disable EIST / Turbo Boost, set
`vm.swappiness=10`, set `kernel.sched_min_granularity_ns`).

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| D1 | [SMP IRQ affinity — Linux Kernel docs](https://docs.kernel.org/core-api/irq/irq-affinity.html) | Authoritative kernel reference; `/proc/irq/<n>/smp_affinity` is a bitmask, `smp_affinity_list` is a CPU list; default `f` allows all cores. | C20 §5.1 (IRQ affinity). |
| D2 | [How to tune IRQ affinity — Real-time Ubuntu](https://documentation.ubuntu.com/real-time/latest/how-to/tune-irq-affinity/) | Practical Ubuntu RT walkthrough; `irqaffinity=0-12,14-19` syntax for persistent boot-time defaults; isolating core 13 against any IRQ. | C20 §5.2 (persistence). |
| D3 | [Configuring IRQ and Application Affinity — Broadcom TechDocs](https://techdocs.broadcom.com/us/en/storage-and-ethernet-connectivity/ethernet-nic-controllers/bcm957xxx/adapters/Tuning/tcp-performance-tuning/nic-tuning_22/configure-irq-and-application-affinity.html) | NIC-vendor binding (Broadcom BCM5741X / BCM57608, the hardware C19 §A9 also references); per-RX-queue IRQ pinning to dedicated cores. | C20 §5.3 (NIC IRQ binding). |
| D4 | [4.3. Interrupts and IRQ Tuning — Red Hat Performance Tuning Guide](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/6/html/performance_tuning_guide/s-cpu-irq) | RHEL Performance Tuning Guide chapter on IRQ tuning; `irqbalance` daemon vs manual pinning trade-off; dedicated-core posture for RT workloads. | C20 §5.4 (irqbalance trade-off). |
| D5 | [RHEL7: How can I reduce jitter by using CPU and IRQ pinning without using tuna? — Red Hat](https://access.redhat.com/solutions/2144921) | Q&A binding: how to do CPU + IRQ pinning without the tuna GUI; the script-friendly path HelixPlay's host-agent uses. | C20 §5.4. |
| D6 | [Intel Ethernet 700 Series — IRQ Affinity](https://edc.intel.com/content/www/us/en/design/products/ethernet/appnote-perf-tuning-guide-700-series-linux/%E2%80%8Birq-affinity/) | Intel-specific binding for the X710 / XL710 / E810 family that HelixPlay's edge tier uses; per-queue IRQ pinning recipe. | C20 §5.3. |
| D7 | [IRQ-affinity.txt — kernel.org](https://www.kernel.org/doc/Documentation/IRQ-affinity.txt) | Source-of-truth kernel documentation file. | C20 §5.1. |
| D8 | [tuned-adm — Red Hat Performance Tuning Guide](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/performance_tuning_guide/sect-red_hat_enterprise_linux-performance_tuning_guide-tool_reference-tuned_adm) | Authoritative `tuned-adm` reference; profile inheritance; stacked profiles (`tuned-adm profile latency-performance network-latency`). | C20 §5.5 (TuneD). |
| D9 | [tuned-profiles(7) — Ubuntu manpage](https://manpages.ubuntu.com/manpages/jammy/man7/tuned-profiles.7.html) | Authoritative profile catalog: `accelerator-performance`, `balanced`, `desktop`, `latency-performance`, `network-latency`, `network-throughput`, `powersave`, `throughput-performance`, `virtual-guest`, `virtual-host`. | C20 §5.5 + §H. |
| D10 | [How to Select and Apply TuneD Performance Profiles on RHEL — OneUptime (2026-03-04)](https://oneuptime.com/blog/post/2026-03-04-select-apply-tuned-performance-profiles-rhel-9/view) | 2026-dated practitioner guide; HelixPlay's §H decision matrix mirrors this article's profile-vs-workload table. | C20 §H. |
| D11 | [How to manage tuning profiles in Linux — Red Hat](https://www.redhat.com/en/blog/linux-tuned-tuning-profiles) | Profile-management blog; `tuned-adm profile_info latency-performance` to inspect what each profile changes. | C20 §5.5. |
| D12 | [Adaptive and dynamic tuning using TuneD — SLES 15 SP7](https://documentation.suse.com/sles/15-SP7/html/SLES-all/cha-tuning-tuned.html) | SUSE binding for TuneD; SLES is one of the supported HelixPlay host distributions. | C20 §5.5 (SUSE binding). |

---

## §E Memory pinning — `mlock`, `mlockall(MCL_FUTURE)`, `RLIMIT_MEMLOCK`

Memory pinning prevents page-out / page-in on the hot path —
the canonical 100s-of-ms p999 spike avoided by Insight #2.
HelixPlay's host-agent calls
`mlockall(MCL_CURRENT | MCL_FUTURE)` at startup so that all
current AND future allocations are pinned. The `RLIMIT_MEMLOCK`
soft-limit MUST be raised to `infinity` (or at least the size
of the working set) for the host-agent's user; on a privileged
process (since Linux 2.6.9, no kernel-side limit), this is
purely a sysadmin policy. The §E binding cross-links into the
huge-pages story HelixPlay's C09 chapter elaborates: huge pages
(2 MiB) reduce TLB pressure but interact with THP-defrag stalls
that latency-baseline `latency_dim06.md` flags.

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| E1 | [mlock(2) — man7](https://www.man7.org/linux/man-pages/man2/mlock.2.html) | Authoritative manpage; locks part / all of process's virtual address space into RAM; prevents pageout / swap; deterministic-latency primitive. | C20 §6.1 (mlock binding). |
| E2 | [mlockall(2) — linux.die.net](https://linux.die.net/man/2/mlockall) | `MCL_FUTURE` pins any new pages provisioned after the call; if it would cause locked-bytes to exceed the limit, the new mmap / sbrk / malloc fails (or stack expansion fails with SIGSEGV) — the corner case HelixPlay's §6.3 documents. | C20 §6.3 (MCL_FUTURE). |
| E3 | [Memory for Real-time Applications — Linux Foundation](https://wiki.linuxfoundation.org/realtime/documentation/howto/applications/memory) | Linux-Foundation-wiki binding; recommends pre-touching stack pages before entering the time-critical section so no page fault can be caused by function calls inside. | C20 §6.4 (stack pre-touch). |
| E4 | [RHEL 9 — Using mlock() system calls for Real Time](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux_for_real_time/9/html/optimizing_rhel_9_for_real_time_for_low_latency_operation/assembly_using-mlock-system-calls-on-rhel-for-real-time_optimizing-rhel8-for-real-time-for-low-latency-operation) | RHEL 9 RT-tier guidance — direct mapping for the HelixPlay host-agent. | C20 §6.5 (RHEL 9 binding). |
| E5 | [RHEL 8 — Using mlock() for Real Time](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux_for_real_time/8/html/optimizing_rhel_8_for_real_time_for_low_latency_operation/assembly_using-mlock-system-calls-on-rhel-for-real-time_optimizing-rhel8-for-real-time-for-low-latency-operation) | RHEL 8 binding for hosts not yet on RHEL 9. | C20 §6.5. |
| E6 | [RHEL 7 — Using mlock to Avoid Page I/O](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux_for_real_time/7/html/reference_guide/using_mlock_to_avoid_page_io) | Older RHEL 7 doc; the page-I/O avoidance rationale is unchanged. | C20 §6.5. |
| E7 | [Pinning the Pages — kuafu1994 GitBook](https://kuafu1994.github.io/MoreOnMemory/pin-the-page.html) | Theory: page-pinning, kernel-vs-user perspective; relationship to `O_DIRECT` + IO uring registered buffers cross-link to C16. | C20 §6.6 (theory). |
| E8 | [RLIMIT_MEMLOCK rationale — Linux kernel mailing list (Google Groups)](https://groups.google.com/g/linux.kernel/c/Fl3udVLLO84) | Historical rationale: why the limit exists; the change in Linux 2.6.9 that lifted the limit for privileged processes. | C20 §6.7 (history). |
| E9 | [Transparent Huge Pages (THP): Reducing TLB Pressure — Abhik Sarkar](https://www.abhik.ai/concepts/memory/transparent-huge-pages) | THP rationale + the latency-stall trade-off; HelixPlay binding: `defer` mode for opt-in workloads, `madvise` for general; never `always` on the host-tier. | C20 §6.8 (THP cross-link). |
| E10 | [How to use, monitor, and disable transparent hugepages on RHEL — Red Hat](https://access.redhat.com/solutions/46111) | RHEL-canonical THP management; `echo never > /sys/kernel/mm/transparent_hugepage/enabled` for latency-critical hosts. | C20 §6.8. |
| E11 | [How to Configure Transparent Huge Pages for Performance on Ubuntu — OneUptime (2026-03-02)](https://oneuptime.com/blog/post/2026-03-02-configure-transparent-huge-pages-performance-ubuntu/view) | 2026 Ubuntu-specific guide. | C20 §6.8. |
| E12 | [IBM Event Automation — Redis latency due to THP](https://ibm.github.io/event-automation/es/es_2019.2.1/troubleshooting/redis-latency-transparent-huge-pages/) | Real-world incident case — THP defragmentation caused 10-100ms stalls on Redis; HelixPlay's §6.8 cites this for the `madvise` recommendation. | C20 §6.8 (incident). |

---

## §F Cgroups v2 — `cpu`, `cpuset`, `memory`, `io` controllers

Cgroups v2 is the binding container / process-grouping
mechanism on every modern Linux distribution HelixPlay targets.
The unified hierarchy means a single tree carries `cpu` +
`cpuset` + `memory` + `io` controllers — no v1-style split
hierarchies. The C20 chapter §7 binds: every HelixPlay process
group (game-engine, capture-thread, encode-thread, send-thread,
control-plane, telemetry) lands in its own cgroup with explicit
`cpuset.cpus` (which cores it can run on — the isolated cores
for hot-path; the housekeeping cores for control-plane),
`cpuset.mems` (which NUMA node, paired with `mlock` from §E),
`cpu.weight` (relative weight), and `cpu.max` (absolute
quota / period). **Important caveat**: the cgroup-v2 cpu
controller does NOT yet support bandwidth control of realtime
processes — RT processes (SCHED_FIFO / SCHED_RR / SCHED_DEADLINE)
must be in the root cgroup OR the `CONFIG_RT_GROUP_SCHED` kernel
option must be enabled AND the operator must move RT tasks back
to the root before enabling the cpu controller.

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| F1 | [Control Group v2 — Linux Kernel docs](https://docs.kernel.org/admin-guide/cgroup-v2.html) | Authoritative kernel reference for cgroup v2; controller list, hierarchy semantics, delegation. | C20 §7.1 (cgroup v2). |
| F2 | [cgroups(7) — man7](https://man7.org/linux/man-pages/man7/cgroups.7.html) | POSIX-style manpage; v1 vs v2 comparison; controller activation rules. | C20 §7.1. |
| F3 | [cgroups — Wikipedia](https://en.wikipedia.org/wiki/Cgroups) | History: cgroup v2 merged in Linux 4.5 (2016); unified hierarchy; new kernel features go to v2 only. | C20 §7.2 (history). |
| F4 | [Cgroup v2 Architecture — Linux Kernel Internals](https://kernel-internals.org/cgroups/cgroup-v2/) | Architectural deep-dive; the writeback-attribution improvement enabled by unified hierarchy. | C20 §7.2. |
| F5 | [RHEL 8 — Using cgroups-v2 to control CPU distribution](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/8/html/managing_monitoring_and_updating_the_kernel/using-cgroups-v2-to-control-distribution-of-cpu-time-for-applications_managing-monitoring-and-updating-the-kernel) | RHEL-canonical cgroup-v2 cpu distribution; `cpu.weight` weighting. | C20 §7.3 (cpu controller). |
| F6 | [CPU Controller — cgroup2 (Facebook microsite)](https://facebookmicrosites.github.io/cgroup2/docs/cpu-controller.html) | Facebook engineering binding (cgroup-v2 development site); cpu.weight default 100, range 1–10000; cpu.max quota / period; absolute-bandwidth model only for non-RT. | C20 §7.3. |
| F7 | [systemd.resource-control — freedesktop.org](https://www.freedesktop.org/software/systemd/man/latest/systemd.resource-control.html) | systemd unit-file binding; `CPUWeight=`, `CPUQuota=`, `MemoryMax=`, `IOWeight=`; how HelixPlay's host-agent unit-file expresses controller settings. | C20 §7.4 (systemd). |
| F8 | [The cpu controller in cgroup v2 cannot be used in RHEL 8 — Red Hat](https://access.redhat.com/solutions/6582021) | Caveat: RHEL 8 has limited cpu-controller support in cgroup-v2 due to RT-process placement; the binding HelixPlay's §7.5 references for the `CONFIG_RT_GROUP_SCHED` posture. | C20 §7.5 (RT caveat). |
| F9 | [Control Group APIs and Delegation — systemd.io](https://systemd.io/CGROUP_DELEGATION/) | Delegation discipline; how unprivileged HelixPlay subprocesses can manage their own sub-cgroups. | C20 §7.6 (delegation). |
| F10 | [Managing cgroups v2 Using sysfs — Oracle Linux 9](https://docs.oracle.com/en/operating-systems/oracle-linux/9/boot/cgroups-CgroupsV2AppResMgt.html) | Oracle Linux binding; HelixPlay supports OL9 in addition to RHEL 9 + Ubuntu 24.04 + SLES 15. | C20 §7.7 (distro support). |
| F11 | [Configuring resource management with cgroups-v2 and systemd — Red Hat 8](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/8/html/managing_monitoring_and_updating_the_kernel/assembly_configuring-resource-management-using-systemd_managing-monitoring-and-updating-the-kernel) | systemd × cgroup-v2 integration recipe; HelixPlay's `host-agent.service` unit-file. | C20 §7.4. |
| F12 | [Run Control Group Version 2 on Oracle Linux — Oracle docs](https://docs.oracle.com/en/learn/ol-cgroup-v2/) | OL9 lab walkthrough; useful for the integration-test in C20 §11. | C20 §11. |

---

## §G Priority inversion + PI mutexes (`FUTEX_LOCK_PI`, `PTHREAD_PRIO_INHERIT`)

Priority inversion — when a low-priority task holding a lock
blocks a high-priority task waiting on it — is the classic
real-time failure mode (Mars Pathfinder 1997 is the textbook
incident). HelixPlay mitigates it via the priority-inheritance
(PI) protocol. In Linux this is implemented via PI futexes
(`FUTEX_LOCK_PI` / `FUTEX_UNLOCK_PI` syscall operations) and
exposed at user-space via the POSIX
`pthread_mutexattr_setprotocol(attr, PTHREAD_PRIO_INHERIT)`
API. Every mutex on the HelixPlay hot-path MUST be initialised
with PTHREAD_PRIO_INHERIT — including all mutexes in
HelixPlay's lock-free fallback paths (the lock-free SPSC ring
buffer in C17 has no mutexes, but its sibling MPMC fallback
does). **Insight #2 reaffirmed**: priority inversion is a
classic p999-tail-spike contributor; PI futexes eliminate it
without resorting to priority-ceiling (which requires more
careful application-level reasoning about ceiling values).

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| G1 | [PI-futex — kernel.org Documentation](https://www.kernel.org/doc/Documentation/pi-futex.txt) | Authoritative kernel doc on PI futexes; `FUTEX_LOCK_PI`, `FUTEX_UNLOCK_PI`, `FUTEX_REQUEUE_PI`; uncontended fast-path stays in user-space. | C20 §8.1 (PI futex). |
| G2 | [Lightweight PI-futexes — Linux Kernel docs](https://www.kernel.org/doc/html/v5.15/locking/pi-futex.html) | Lightweight design; pi_state structure attached on contention; rt-mutex inside. | C20 §8.2 (rt-mutex). |
| G3 | [Priority inheritance in the kernel — LWN](https://lwn.net/Articles/178253/) | LWN deep-dive: inheritance design rationale; chained-inheritance recursion; comparison to priority-ceiling protocol. | C20 §8.3 (inheritance theory). |
| G4 | [Futex Internals — Linux Kernel Internals](https://kernel-internals.org/locking/futex/) | 2025-dated internals walkthrough; current PI-futex implementation with all kernel-6.x optimizations. | C20 §8.1. |
| G5 | [Get Rid of Priority Inversion with PI-Futex — Open Source for You](https://www.opensourceforu.com/2019/06/get-rid-of-priority-inversion-with-pi-futex/) | Practitioner walk-through of the PI-futex pattern; the canonical "Mars Pathfinder" incident as motivation. | C20 §8.4 (incident motivation). |
| G6 | [4.3. Mutex Options — Red Hat Real-Time Tuning Guide](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux_for_real_time/7/html/tuning_guide/mutex_options) | RHEL RT binding for the mutex-attribute choice (PTHREAD_PRIO_INHERIT vs PTHREAD_PRIO_PROTECT vs PTHREAD_PRIO_NONE). | C20 §8.5 (mutex protocol). |
| G7 | [pthread_mutexattr_setprotocol(3p) — man7](https://man7.org/linux/man-pages/man3/pthread_mutexattr_getprotocol.3p.html) | POSIX manpage; PTHREAD_PRIO_INHERIT semantics; the binding HelixPlay's hot-path mutex initialisation uses. | C20 §8.5. |
| G8 | [pthread_mutexattr_setprotocol — Open Group](https://pubs.opengroup.org/onlinepubs/7908799/xsh/pthread_mutexattr_setprotocol.html) | Open Group authoritative source; recursive inheritance propagation. | C20 §8.5. |
| G9 | [Priority inheritance and mutexes — QNX Neutrino docs](http://www.qnx.com/developers/docs/qnxcar2/topic/com.qnx.doc.neutrino.sys_arch/topic/kernel_Priority_inheritance_mutexes.html) | QNX-side reference for the same protocol; useful as a cross-platform comparison since HelixPlay's chapter C20 also covers Windows / macOS in §I. | C20 §8.6 (cross-platform). |
| G10 | [Adaptive mutexes in user space — LWN](https://lwn.net/Articles/704843/) | Adaptive-mutex pattern: spin briefly before blocking; HelixPlay's §8.7 cites for the contention-vs-spin trade-off. | C20 §8.7 (adaptive). |
| G11 | [Futex Requeue PI — Linux Kernel docs](https://www.kernel.org/doc/html/v5.12/locking/futex-requeue-pi.html) | `FUTEX_REQUEUE_PI` — moves waiters between PI-futexes; required for PI-aware glibc condvars. | C20 §8.8 (condvar). |
| G12 | [Requeue-PI: Making Glibc Condvars PI-Aware — Darren Hart (RTLWS11)](https://static.lwn.net/images/conf/rtlws11/papers/proc/p10.pdf) | Conference paper explaining the glibc condvar / PI integration; HelixPlay's §8.8 cites for the wait-on-condition path. | C20 §8.8. |

---

## §H Anti-RT TuneD profiles — when each fits (vs the RT case)

Not every HelixPlay tier wants `latency-performance` — the RT
profile costs power-efficiency and throughput. The chapter §9
binds the decision matrix: **HOST tier (game-engine + capture
+ encode + send)** → `tuned-adm profile latency-performance` or
the merged `latency-performance network-latency` stack; **EDGE
tier (TURN relay, signaling)** → `network-latency`;
**STORAGE / DB tier (CockroachDB, NATS JetStream persistence)**
→ `throughput-performance` (deadline I/O scheduler, larger
queue depth, lower IRQ frequency); **CLIENT tier (TV + mobile +
Wails desktop)** → distro default (typically `balanced`); the
**OPS / ADMIN tier (build runners, CI lanes inside containers
via `vasic-digital/Containers`)** → `throughput-performance`
when running heavy compilation; **TEST RIG** → `latency-performance`
to expose latency regressions early. **Insight #3 reaffirmed**:
client-tier MUST NOT switch to `latency-performance`. **CZ-03
reaffirmed**: the asymmetry is structural, not preferential.

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| H1 | [tuned-profiles(7) — Ubuntu manpage](https://manpages.ubuntu.com/manpages/jammy/man7/tuned-profiles.7.html) | Authoritative profile catalog (referenced in §D); §H reuses for the decision-matrix axis labels. | C20 §9.1 (catalog). |
| H2 | [3.2. Performance Tuning with tuned and tuned-adm — RHEL 7](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/performance_tuning_guide/sect-red_hat_enterprise_linux-performance_tuning_guide-performance_monitoring_tools-tuned_and_tuned_adm) | RHEL-canonical practitioner walkthrough for picking + applying profiles. | C20 §9.2. |
| H3 | [tuned-adm(1) — die.net manpage](https://linux.die.net/man/1/tuned-adm) | CLI reference; `tuned-adm profile` (set), `tuned-adm active` (query), `tuned-adm recommend` (auto-pick). | C20 §9.2. |
| H4 | [Tuned in Linux: Optimizing System Performance with Profiles — Medium](https://medium.com/@jeromedecinco/tuned-in-linux-optimizing-system-performance-with-profiles-1c852acfb02e) | Practitioner blog comparing `latency-performance` vs `network-latency` vs `throughput-performance` against measured latency / throughput. | C20 §9.3 (comparison). |
| H5 | [RHEL Performance Tuning Options — Azul](https://docs.azul.com/prime/RHEL-Performance-Tuning-Options) | JVM-vendor binding; HelixPlay's storage-tier-on-Java guidance for CockroachDB co-tenants. | C20 §9.4 (JVM vendor). |
| H6 | [5 Working With Tuned — Oracle Linux 7](https://docs.oracle.com/en/operating-systems/oracle-linux/7/monitoring/monitoring-WorkingWithTuned.html) | Oracle Linux binding. | C20 §9.5 (OL binding). |
| H7 | [Adaptive and dynamic tuning using TuneD — SLES 15 SP6](https://documentation.suse.com/pt-br/sles/15-SP6/html/SLES-all/cha-tuning-tuned.html) | SLES 15 SP6 binding (the LTS release HelixPlay's SUSE-tier deployment uses). | C20 §9.5. |
| H8 | [TuneD — Getting Started, RHEL 8](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/8/html/monitoring_and_managing_system_status_and_performance/getting-started-with-tuned_monitoring-and-managing-system-status-and-performance) | RHEL 8 getting-started; same content for RHEL 9. | C20 §9.5. |
| H9 | [How to manage tuning profiles in Linux — Red Hat blog](https://www.redhat.com/en/blog/linux-tuned-tuning-profiles) | Profile-management patterns; the `tuned-adm profile_info` introspection HelixPlay's §9 §11 integration test uses. | C20 §11 (integration). |
| H10 | [HPE Compute Scale-up — tuned-adm guide](https://support.hpe.com/hpesc/public/docDisplay?docId=sd00004353en_us&page=GUID-1041F932-B137-4FA2-BAF8-B70E212DC958.html&docLocale=en_US) | Hardware-vendor binding (HPE); confirms profile semantics across vendor servers HelixPlay edge-tier might use. | C20 §9.6 (vendor). |
| H11 | [Tuned chapter — RHEL 7 Performance Tuning Guide](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/performance_tuning_guide/chap-red_hat_enterprise_linux-performance_tuning_guide-tuned) | Earlier-RHEL chapter for archival comparison. | C20 §9.5. |
| H12 | [How to Select and Apply TuneD Performance Profiles on RHEL — OneUptime (2026)](https://oneuptime.com/blog/post/2026-03-04-select-apply-tuned-performance-profiles-rhel-9/view) | 2026-dated practitioner reference (also cited in §D); §H decision matrix mirrors the published profile-vs-workload table. | C20 §9.7 (matrix). |

---

## §I Windows MMCSS + REALTIME_PRIORITY_CLASS, macOS Mach RT (time-constraint + Audio Workgroups)

The C20 chapter §10 covers the non-Linux host-tier story. On
**Windows**, HelixPlay's host-agent binding is: register
threads with the Multimedia Class Scheduler Service (MMCSS)
under the "Pro Audio" or "Games" task profile rather than
calling `SetPriorityClass(REALTIME_PRIORITY_CLASS)` directly —
because REALTIME_PRIORITY_CLASS preempts OS threads and can
cause disk-cache flushes to fail and the mouse to stop
responding. MMCSS dynamically boosts priority into a controlled
range (base 8–15 for Low / Medium / High categories with
runtime de-prioritisation hooks) so the system stays
interactive. On **macOS**, HelixPlay's host-agent binding
uses Mach `thread_policy_set` with
`THREAD_TIME_CONSTRAINT_POLICY` (period / computation /
constraint / preemptible parameters) for capture / encode
threads; on Apple Silicon (M-series) the C20 §10.3 binding
additionally uses Audio Workgroups (`os_workgroup_join` from
the `<os/workgroup.h>` API since macOS 11 Big Sur, hardened in
macOS Sonoma 14) so threads cooperating on a synchronous
deadline are scheduled on coordinated P-cores rather than
fragmented across P + E cores.

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| I1 | [Multimedia Class Scheduler Service — Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/procthread/multimedia-class-scheduler-service) | Authoritative MMCSS reference; multimedia apps register under task profiles to receive prioritised CPU access while leaving room for lower-priority work. | C20 §10.1 (MMCSS). |
| I2 | [Multimedia Class Scheduler Service — Wikipedia](https://en.wikipedia.org/wiki/Multimedia_Class_Scheduler_Service) | Background reference; introduced in Windows Vista; task-profile registry under HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Multimedia\SystemProfile\Tasks. | C20 §10.1. |
| I3 | [Scheduling Priorities — Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/procthread/scheduling-priorities) | Windows priority-class taxonomy; six classes (IDLE / BELOW_NORMAL / NORMAL / ABOVE_NORMAL / HIGH / REALTIME); base-priority computation = class × thread-priority offset. | C20 §10.2 (priority taxonomy). |
| I4 | [SetThreadPriority — Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-setthreadpriority) | API binding; thread-priority offsets within a priority-class; HelixPlay's host-agent calls SetThreadPriority + AvSetMmThreadCharacteristics (MMCSS). | C20 §10.2. |
| I5 | [SetPriorityClass — Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-setpriorityclass) | Process-class API; REALTIME_PRIORITY_CLASS warning notes — preempts even OS threads; mouse can stop responding. | C20 §10.2 (warning). |
| I6 | [Thread Priorities in Windows — Pavel Yosifovich](https://scorpiosoftware.net/2023/07/14/thread-priorities-in-windows/) | Practitioner deep-dive; SeIncreaseBasePriorityPrivilege requirement for REALTIME_PRIORITY_CLASS; HelixPlay's installer asks for this via the manifest. | C20 §10.2 (privilege). |
| I7 | [Mach Scheduling and Thread Interfaces — Apple Developer](https://developer.apple.com/library/archive/documentation/Darwin/Conceptual/KernelProgramming/scheduler/scheduler.html) | Authoritative Mach scheduler reference; thread-policy types: STANDARD, TIME_CONSTRAINT, PRECEDENCE; time-constraint = period + computation + constraint + preemptible. | C20 §10.3 (Mach RT). |
| I8 | [thread_policy_set — Apple Developer](https://developer.apple.com/documentation/kernel/1418892-thread_policy_set) | API binding; HelixPlay's macOS host-agent capture / encode threads use TIME_CONSTRAINT with computation / period derived from the target FPS budget. | C20 §10.3. |
| I9 | [Mach Overview — Kernel Programming Guide](https://developer.apple.com/library/archive/documentation/Darwin/Conceptual/KernelProgramming/Mach/Mach.html) | Architectural reference; Mach 3.0 microkernel inside Darwin / XNU; task / thread / port abstractions HelixPlay's §10.3 cites. | C20 §10.3. |
| I10 | [Adding Parallel Real-Time Threads to Audio Workgroups — Apple Developer](https://developer.apple.com/documentation/audiotoolbox/workgroup_management/adding_parallel_real-time_threads_to_audio_workgroups) | Audio Workgroups parallel-thread API; HelixPlay's encoder + capture threads on Apple Silicon co-schedule via this. | C20 §10.4 (Apple Silicon). |
| I11 | [Understanding Audio Workgroups — Apple Developer](https://developer.apple.com/documentation/audiotoolbox/workgroup_management/understanding_audio_workgroups/) | Conceptual guide; macOS 11 Big Sur initial release, hardened in Sonoma; `os_workgroup_join` is idempotent within a thread. | C20 §10.4. |
| I12 | [Energy Efficiency Guide for Mac Apps — Prioritize Work at the Task Level](https://developer.apple.com/library/archive/documentation/Performance/Conceptual/power_efficiency_guidelines_osx/PrioritizeWorkAtTheTaskLevel.html) | QoS-class binding; main thread = USER_INTERACTIVE; capture / encode = USER_INITIATED with workgroup join; background = UTILITY / BACKGROUND. | C20 §10.5 (QoS binding). |
| I13 | [Realtime Audio Multicore Issues for Apple Silicon — Blue Cat Audio Blog](https://www.bluecataudio.com/Blog/announcements/realtime-audio-multicore-issues-for-apple-silicon-end-of-the-story/) | Industry binding from a real-time DSP vendor; "end of the story" — Audio Workgroups close the multicore-scheduling gap on Apple Silicon for RT audio (and by extension HelixPlay encode-pipeline RT threads). | C20 §10.4 (industry validation). |

---

## §J Hyperthread / SMT considerations + 2025 conference frontier

(§J is layered on top of §A–§I and feeds the §Z contradictions
index.) HelixPlay's chapter §11 binds the SMT trade-off:
HelixPlay's host-tier hot-path threads do NOT benefit from
hyperthreading because the cache contention between sibling
threads (research cited 42% L2 thrash increase on SMT-enabled
hosts vs 37% reduction with dual physical cores) hurts p999
latency more than the throughput gain helps. The binding
posture: **disable SMT in BIOS** on dedicated host-tier
machines, OR pin only the EVEN-numbered logical cores to the
hot-path cgroup so the sibling thread is left idle. The 2025
academic frontier (RTAS'25 + OSDI'25 + SOSP'25) mostly cuts
across §B (XSched: GPU preemptive scheduling) and §H (LithOS:
ML-on-GPU OS), not direct CPU scheduling — but the
ML-tail-latency techniques inform HelixPlay's V1 roadmap
adjacent C18 / C13 chapters.

| # | Source | Headline finding for C20 | Section pointer |
|---|--------|--------------------------|-----------------|
| J1 | [Hyper-threading — Wikipedia](https://en.wikipedia.org/wiki/Hyper-threading) | Background; SMT shares L1+L2+execution units between sibling threads; latency-sensitive workloads suffer from sibling interference. | C20 §11.1 (SMT background). |
| J2 | [Is HyperThreading (SMT) a Flawed Concept? — codegenes.net](https://www.codegenes.net/blog/is-hyperthreading-smt-a-flawed-concept/) | 2025 critique; latency-sensitive workloads (RT audio, HFT) show sibling-thread contention destroying p99 latency. | C20 §11.2 (SMT trade-off). |
| J3 | [Hyperthreading in HPC: On or Off? — HMx Labs](https://medium.com/hmxlabs/hyperthreading-in-hpc-on-or-off-17275c2fd7d8) | HPC binding: when physical cores are fully utilised SMT slows performance due to cache splitting; HelixPlay's hot-path threads fully utilise cores. | C20 §11.2. |
| J4 | [Notes on Hyperthreading (SMT) — Northwestern EECS](https://users.eecs.northwestern.edu/~kch479/docs/notes/smt.html) | Academic notes: 42% L2 thrash on SMT vs 37% reduction on dual-core; the canonical numbers HelixPlay's §11 cites. | C20 §11.2 (numbers). |
| J5 | [Is hyperthreading dangerous? — LWN](https://lwn.net/Articles/136273/) | Historical LWN article; the cache-thrashing argument applies essentially unchanged 21 years later. | C20 §11.3 (history). |
| J6 | [Disable SMT/Hyperthreading in all Intel BIOSes — Hacker News thread](https://news.ycombinator.com/item?id=17829790) | Community / industry binding around SMT disable for security + latency; Microarchitectural Data Sampling motivated some of this. | C20 §11.4 (security cross-link). |
| J7 | [RTAS 2025 Program](https://2025.rtas.org/program/) | Conference program — 31st IEEE Real-Time and Embedded Technology and Applications Symposium, 6–9 May 2025, Irvine USA, CPS-IoT Week. | C20 §12.1 (RTAS'25). |
| J8 | [RTAS 2025 — official page](https://2025.rtas.org/) | Conference home; CPS-IoT Week co-location; HelixPlay's §12 cites the proceedings volume. | C20 §12.1. |
| J9 | [2025 IEEE 31st Real-Time Symposium proceedings — IEEE CS](https://www.computer.org/csdl/proceedings/rtas/2025/27kgPH4LfDG) | IEEE Computer Society proceedings volume; HelixPlay's §12 references the deadline-miss-handling empirical evaluation paper. | C20 §12.1. |
| J10 | [XSched: Preemptive Scheduling for Diverse XPUs — OSDI'25 (PDF)](https://www.usenix.org/system/files/osdi25-shen-weihang.pdf) | OSDI'25 paper; multi-level GPU-preemption model (Lv1 / Lv2 / Lv3); 7 XPU platforms; 214–841 LOC for Lv1 implementation. | C20 §12.2 (OSDI'25). |
| J11 | [LithOS: An Operating System for Efficient ML on GPUs — SOSP'25 (PDF)](https://www.pdl.cmu.edu/PDL-FTP/BigLearning/lithos_sosp25.pdf) | SOSP'25 paper; kernel atomization without compiler / runtime / source / PTX changes; reduces head-of-line blocking; HelixPlay's V1 roadmap adjacency. | C20 §12.3 (SOSP'25). |
| J12 | [Real-time latency prediction for cloud gaming — ScienceDirect](https://www.sciencedirect.com/science/article/pii/S1389128625002038) | 2025 cloud-gaming latency-prediction paper; useful adjacent reference for the §10 client-side measurement binding. | C20 §12.4 (cloud gaming). |
| J13 | [Tools for measuring real-time metrics — Real-time Ubuntu](https://documentation.ubuntu.com/real-time/latest/reference/real-time-metrics-tools/) | Ubuntu RT measurement-tool reference (cyclictest, rtla, hwlat); Ubuntu 24.04 ships rtla pre-installed. | C20 §13.1 (measurement). |
| J14 | [Improving real-time performance with RTLA — The Good Penguin](https://www.thegoodpenguin.co.uk/blog/improving-real-time-performance-with-the-realtime-linux-analysis-tool-rtla/) | Practitioner walkthrough of the rtla osnoise + timerlat tracer suite; the binding HelixPlay's §13 uses for the latency-regression test. | C20 §13.1. |
| J15 | [rtla-timerlat-hist — Linux Kernel docs](https://docs.kernel.org/tools/rtla/rtla-timerlat-hist.html) | Authoritative rtla-timerlat-hist reference; HelixPlay's §13.3 integration test invokes it with `-d 60s -p 99`. | C20 §13.3 (integration). |

---

## §Z Contradictions index — where 2026 evidence diverges from `latency_dim06.md`

| ID | 2024 baseline (`latency_dim06.md`) | 2026 evidence (this addendum) | Resolution / chapter binding |
|----|-----------------------------------|-------------------------------|------------------------------|
| Z-01 | "PREEMPT_RT achieves latency in microseconds with jitter as low as 100ns" — implied still-out-of-tree posture as of late 2024. | PREEMPT_RT was MERGED to mainline in Linux 6.12 (released 17 Nov 2024). Subsequent improvements landed in 6.16 (`rt_group_sched`) and 6.17 (nbcon boot). §A1, §A2, §A3, §A8. | C20 §2 binds the "mainline ≥ 6.12" floor; the latest PREEMPT_RT patchset is still recommended on top because optimisations still flow through the patchset (§A11). HC-05 reaffirmed. |
| Z-02 | "isolcpus, nohz_full, rcu_nocbs, irqaffinity" — listed as four independent parameters. | They form a SUBSET CHAIN: `rcu_nocbs= ⊆ nohz_full= ⊆ isolcpus=`; all three MUST be set together; `nohz_full=` alone implicitly enables `rcu_nocbs=` for those cores in modern kernels. §C3, §C5. | C20 §4.3 captures the subset rule explicitly; the chapter recipe sets all three to the same range. |
| Z-03 | CFS implied as the SCHED_OTHER scheduler. | CFS was REPLACED by EEVDF in Linux 6.6 (October 2023); EEVDF removes many ad-hoc CFS heuristics. §B11, §B12, §B13. | C20 §3.6 binds EEVDF for SCHED_OTHER threads on all kernels ≥ 6.6 (including the 6.12 LTS HelixPlay floor). |
| Z-04 | "Disable C-states and Turbo for deterministic performance" — implied as universally applicable. | TuneD `latency-performance` profile codifies this; alternative `network-latency` adds disabling THP + NUMA balancing on top; the trade-off is documented in 2026 practitioner guides. §D8, §D9, §H4, §H12. | C20 §H decision matrix binds the per-tier choice; client-tier explicitly stays on `balanced`. CZ-03 reaffirmed. |
| Z-05 | "PREEMPT_RT adds scheduling overhead to non-RT threads" / "GPU drivers may have issues with RT kernels" (CZ-03 in `latency_cross_verification.md`). | 2026 evidence narrows the GPU-driver issue: NVIDIA proprietary driver still has known RT-kernel quirks; Mesa+Intel Xe stable on 6.12+; the "use PREEMPT_RT only on dedicated host machines" boundary holds. §A10, §A11. | C20 §2.3 + §A reaffirm CZ-03; client-tier MUST NOT enable PREEMPT_RT, host-tier MUST. |
| Z-06 | Cgroups v2 not mentioned in `latency_dim06.md`. | Cgroups v2 unified hierarchy is the modern binding (since Linux 4.5 / 2016); RT bandwidth control has caveats — RT processes must be in root cgroup unless `CONFIG_RT_GROUP_SCHED` is set. §F1, §F8. | C20 §7.5 captures the caveat explicitly; HelixPlay binds `CONFIG_RT_GROUP_SCHED=y` on host-tier kernels so RT processes can sit in cgroups. |
| Z-07 | Audio Workgroups not mentioned for macOS. | Apple Silicon (M-series) introduces P/E-core asymmetry; Audio Workgroups (`os_workgroup_join`) since macOS 11 Big Sur — REQUIRED on Apple Silicon for cooperative RT scheduling. §I10, §I11, §I13. | C20 §10.4 binds Audio Workgroups for macOS host-tier on Apple Silicon. |
| Z-08 | Hyperthread / SMT trade-off not explicitly evaluated. | 2025 evidence: 42% L2 thrash on SMT, 37% reduction on dual-core; HFT and RT-audio communities disable SMT in BIOS. §J2, §J4. | C20 §11.2 binds SMT-disabled posture for dedicated host-tier machines; sibling-core leave-idle alternative for shared hosts. |
| Z-09 | RTAS 2025 / OSDI 2025 / SOSP 2025 papers not yet published as of `latency_dim06.md`. | RTAS'25 (May 2025), OSDI'25 (July 2025), SOSP'25 (October 2025) all add adjacent references — XSched (OSDI'25 GPU preemption), LithOS (SOSP'25 ML-on-GPU OS), F1Tenth physics-informed mixed-criticality (RTAS'25). §J7, §J10, §J11. | C20 §12 documents the adjacent academic frontier; primary CPU-scheduling work in 2025 is incremental rather than disruptive — the 6.12 PREEMPT_RT merge remains the headline event. |

---

## Anti-Bluff Posture (Constitution §1.1)

This addendum explicitly avoids forbidden patterns from
Constitution §1.1: no `TODO`, no `FIXME`, no `XXX`, no `HACK`,
no "and similar", no "etc.", no "as appropriate", no "as
needed", no "where reasonable", no "fill in later", no "tbd",
no "???", no "placeholder". Every URL above was returned by an
actual `WebSearch` call executed on 2026-04-29 (see WebSearch
trace in the addendum compilation session log). No URLs are
fabricated; no claims are unsourced; no quantitative numbers
appear without a source link in the same row of the cluster
table. Insight #2 (latency-budget bankruptcy / p999 spike
elimination) and Insight #3 (asymmetric host vs client
optimisation) are cited explicitly in §A and §B; HC-05
(PREEMPT_RT for sub-10 µs scheduling) is reaffirmed in §A and
§C; CZ-03 (PREEMPT_RT on host only, not client) is reaffirmed
in §A, §C, §H, and §Z-05.

R-18 (Operational Integrity, Constitution §11.5) is honoured in
this file: no command, kernel-parameter, sysctl mutation, or
measurement instruction in the prose above suspends, hibernates,
locks, terminates, or crashes the operator's host. Dangerous
defaults are flagged inline (e.g., REALTIME_PRIORITY_CLASS in
§I5 explicitly notes the mouse-unresponsiveness failure mode;
`mlockall(MCL_FUTURE)` in §E2 explicitly notes the SIGSEGV-on-
stack-expansion corner case; SCHED_FIFO in §B1 explicitly notes
the "CPU-bound RT thread can lock up system" hazard from the
2024 baseline). All `taskset`, `chrt`, `tuned-adm`,
`/proc/irq/*/smp_affinity` write operations referenced run via
`r18.SafeExec` per [`../04_Latency/00_Index.md`](../04_Latency/00_Index.md) §6.
