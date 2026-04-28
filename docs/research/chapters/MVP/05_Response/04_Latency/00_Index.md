# Latency Engineering — Chapter Index

> **Family scope.** This is the index for HelixPlay's `04_Latency/`
> chapter family — the deep, OS-/hardware-/performance-engineering
> elaboration of latency primitives that the
> Architecture-side overview at
> [`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md)
> identifies as the binding architectural floor. Where C13 is the
> *what* and *why* of latency engineering for the cloud-gaming pipeline,
> this family is the *how* — kernel bypass, lock-free IPC, GPU-Direct
> data paths, real-time scheduling, controller polling at 1000 Hz,
> frame pacing + VRR, allocator + cache hierarchy, and the
> measurement / validation harness that proves the rest works.
>
> **Source stream:** Stream 2 (`docs/research/chapters/MVP/02_latency/`)
> — 10 latency dimensions plus a long-form synthesis at
> [`docs/research/chapters/MVP/02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md`](../../02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md).
>
> **Status:** index landed; C15..C24 (10 chapters) queued under R1
> section-stitched dispatch (Master Plan §5).
>
> **Last updated:** 2026-04-29.

---

## 1. Why this family exists

The Architecture chapter family (C01..C13) ends with C13
[`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md)
— a 3,816-line overview that breaks the end-to-end latency budget
into 11 layers and forward-links each to a deeper chapter under
this family. C13 is the architectural table-of-contents; the
chapters indexed here (C15..C24) are where the implementation
contract lives.

Every chapter in this family inherits, without re-implementing
(Constitution §2 DRY):

- The `r18.SafeExec` wrapper from
  [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md)
  §10 — used for any subprocess invocation
  (`tc qdisc`, `setcap`, `chrt`, `taskset`, `numactl`, `irqbalance`,
  `presentmon`, etc.).
- The `host-integrity-scan` test from C08 §12.11 — non-overridable
  per Constitution §11.5.4.
- The Constitution-§6 mandate that **every latency claim reports
  p50 / p99 / p999 at ≥ 10 K samples**.

The family is the canonical home for the four **latency-stream
conflict zones** (CZ-01..CZ-04 from
[`../../02_latency/02_Response/Agent_results/research/latency_cross_verification.md`](../../02_latency/02_Response/Agent_results/research/latency_cross_verification.md)):

- **CZ-01** io_uring vs DPDK for video streaming — owned by C19
  (`05_UltraLowLatency_Network_Protocols.md`).
- **CZ-02** Zero-copy overhead for small packets — owned by C16
  (`02_io_uring_and_Kernel_Bypass.md`).
- **CZ-03** PREEMPT_RT vs standard kernel — owned by C20
  (`06_RealTime_OS_and_Scheduling.md`).
- **CZ-04** 1000 Hz USB polling vs power consumption — owned by C21
  (`07_Controller_Input_Optimization.md`).

The five **latency-stream insights**
([`../../02_latency/02_Response/Agent_results/research/latency_insight.md`](../../02_latency/02_Response/Agent_results/research/latency_insight.md))
are distributed across the family:

| # | Insight | Primary chapter | Cross-cited |
|---|---------|-----------------|-------------|
| 1 | **Microwave Pipeline** — unified zero-copy controller→GPU→encoder→network path | C18 (GPU-Direct) | C15 (shared mem), C16 (io_uring), C17 (lock-free) |
| 2 | **p999 only metric** — spike elimination, not average optimisation | C24 (testing/validation) | every chapter ("report p50/p99/p999 ≥10K samples" rule) |
| 3 | **Asymmetric optimisation** — host ≠ client playbooks | C20 (host RT-OS) + C22 (client frame pacing) | every chapter |
| 4 | **Allocation-free architecture** — pre-allocated pools on hot path | C23 (memory/cache) | C17 (lock-free), C21 (controller) |
| 5 | **Conservative prediction paradox** — analog yes, discrete no | C21 (controller) | C22 (frame pacing — Frame Warp limits) |

The ten **high-confidence cross-verified findings** (HC-01..HC-10
from `latency_cross_verification.md`) distribute as: HC-01 (shm + lock-free SPSC) → C15+C17;
HC-02 (io_uring async I/O) → C16; HC-03 (Reflex + Frame Warp) → C18+C22;
HC-04 (1000 Hz USB polling) → C21; HC-05 (PREEMPT_RT scheduling) → C20;
HC-06 (DPDK / XDP / io_uring trade-off) → C19; HC-07 (zero-copy GPU
pipeline) → C18; HC-08 (p99/p999 + ≥10 K samples) → C24;
HC-09 (VRR < 1 ms display-side cost) → C22 (and C13 §8 already);
HC-10 (false-sharing elimination + 64-byte cache-line padding) → C23+C17.

---

## 2. Chapter map

The chapter map mirrors the dimension decomposition in
[`../../02_latency/02_Response/Agent_results/research/latency_dim_decomposition.md`](../../02_latency/02_Response/Agent_results/research/latency_dim_decomposition.md):

| Chapter | File | Source dim | Line floor (R-01) | Status |
|---------|------|------------|------------------:|:------:|
| C14 | [`00_Index.md`](00_Index.md) (this file) | overview | 300 | landed |
| C15 | `01_Shared_Memory_and_Zero_Copy_IPC.md` | latency dim01 | 250 | queued |
| C16 | `02_io_uring_and_Kernel_Bypass.md` | latency dim02 | 250 | queued |
| C17 | `03_LockFree_Data_Structures.md` | latency dim03 | 250 | queued |
| C18 | `04_GPU_Direct_and_Hardware_Pipelines.md` | latency dim04 | 300 | queued |
| C19 | `05_UltraLowLatency_Network_Protocols.md` | latency dim05 | 250 | queued |
| C20 | `06_RealTime_OS_and_Scheduling.md` | latency dim06 | 300 | queued |
| C21 | `07_Controller_Input_Optimization.md` | latency dim07 | 250 | queued |
| C22 | `08_Frame_Pacing_and_VRR.md` | latency dim08 | 250 | queued |
| C23 | `09_Memory_and_Cache_Optimization.md` | latency dim09 | 250 | queued |
| C24 | `10_Latency_Testing_and_Validation.md` | latency dim10 | 300 | queued |

The line floors are calibrated against the per-dim source files
(91–136 lines each, total ≈ 1,148 lines across `latency_dim01..10.md`)
— much smaller than the cloud-gaming dimension files (1,000+ lines
each) that backed the Architecture family. Every chapter still
targets ≥ 2× its floor in body prose under R1 dispatch, and chapter
prose elaborates the dim file with the long-form synthesis at
[`../../02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md`](../../02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md)
plus a per-chapter `99_Web_Research_Addenda/` web-evidence file.

---

## 3. Cross-references to C13 sections

C13 is the architectural overview; the deep chapters elaborate
specific sections of C13:

| C13 section | Maps to | Content scope |
|-------------|---------|---------------|
| C13 §3 frame-time + frame pacing | C22 (`08_Frame_Pacing_and_VRR.md`) | Frame Warp / VRR / G-Sync / FreeSync / ALLM / 120-240 Hz; client decode + scanout floor |
| C13 §4 DSCP / WMM / L4S / QoS | C19 (`05_UltraLowLatency_Network_Protocols.md`) | DPDK, eBPF/XDP, io_uring + NAPI, RoCE, raw UDP, kernel bypass |
| C13 §5 UDP vs TCP | C19 | (same chapter — owned end-to-end) |
| C13 §6 FEC + jitter buffer | C22 + C19 | jitter buffer on client side; FEC schedules in C19 |
| C13 §7 client-side frame interpolation | C22 | DLSS 4.5 / FSR 4.1 / XeSS 3.0; Frame Warp |
| C13 §8 VRR / G-Sync / FreeSync / ALLM | C22 | (primary chapter) |
| C13 §9 120/144/240 Hz feasibility | C22 | (primary chapter) |
| C13 §10 measurement methodology | C24 (`10_Latency_Testing_and_Validation.md`) | LDAT / OSRTT / PresentMon 2.2 / Reflex SDK / G-SYNC 10 K-sample p99 |
| C13 §11 bandwidth requirements | C19 | (codec budgets + MEC RTT cross-link to C09 §5) |

The hardware-pipeline + IPC layer (C15 / C16 / C17 / C18) and the
RT-OS + scheduling layer (C20) and the memory/cache layer (C23)
are *new* in this family — they extend below the architectural
abstraction line that C13 holds.

---

## 4. Forward-links to other families

- **Operations** — [`../08_Operations/04_Observability_and_Events.md`](../08_Operations/04_Observability_and_Events.md) consumes the per-frame histogram pipeline produced under HC-08 + C24.
- **Testing** — [`../07_Testing/02_Benchmarking_Tests.md`](../07_Testing/02_Benchmarking_Tests.md) imports the C24 measurement harness as the canonical Benchmarking-test surface for latency claims.
- **Implementation Phases** — [`../09_Implementation_Phases/Phase_12_Latency_Tuning.md`](../09_Implementation_Phases/Phase_12_Latency_Tuning.md) is the primary implementation phase for the family; it pulls C15..C24 into a phased rollout (Phase 12.1 IPC, 12.2 io_uring, 12.3 lock-free, 12.4 GPU-Direct, 12.5 ULL networking, 12.6 RT-OS, 12.7 controller, 12.8 frame pacing, 12.9 memory/cache, 12.10 validation harness).

---

## 5. Anti-Bluff posture (R-13)

Every chapter in this family carries an `## Anti-Bluff Verification`
block per Master Plan §4.3. The blocks list:

- The exact source files reviewed (path + line count + reviewer + date + sections used).
- The web-research addendum URLs with cluster table.
- The latency-stream insights and HCs incorporated.
- The conflict zones resolved (own + inherited).
- R-18 compliance evidence (chapter prose + code + tests).
- Coverage confirmation (line counts vs floors, forbidden-pattern scan, table inventory, code-block inventory).
- Sign-off rows per section subagent + orchestrator.

Forbidden-pattern scan: `TODO`, `FIXME`, `XXX`, `HACK`, `tbd`,
`???`, `placeholder`, "and similar", "etc.", "as appropriate",
"as needed", "where reasonable", "fill in later". Self-referential
mentions inside Constitution-cite text are explicitly permitted by
Constitution §1.1 and Master Plan §5.2.3.

---

## 6. R-18 Operational Integrity inheritance

Every chapter in this family inherits R-18 enforcement from C08
(Constitution §2 DRY):

- **Static — chapter prose**: §1 of each chapter references R-18; §11.5 forbidden-command list is recapped where needed (host RT-OS chapter C20 has the longest recap because `chrt`, `taskset`, `numactl`, `cgconfig`, etc. are the closest legitimate-but-near-forbidden commands).
- **Static — code**: imports `r18.SafeExec` from `vasic-digital/helix-r18-safeexec` (origin C08 §10). The deny-list is **not duplicated** anywhere in this family.
- **Test — `host-integrity-scan`**: inherited from C08 §12.11 verbatim into every chapter's §12 Test surface.

No `systemctl suspend|hibernate|poweroff|reboot|halt`,
`loginctl lock-session`, `pmset`, `xset dpms force off`, `kill -9 1`,
`init 0`, `setterm -blank`, `--privileged`, host-mount of
`/`, `/dev`, `/proc`, `/sys` appears anywhere in this family —
the few legitimate-near-forbidden commands (`chrt`, `taskset`,
`numactl`, `cgconfig`, `tc qdisc`, `ip link set`, `ethtool`)
all run through `r18.SafeExec`.

---

End of `04_Latency/00_Index.md` — 2026-04-29.
