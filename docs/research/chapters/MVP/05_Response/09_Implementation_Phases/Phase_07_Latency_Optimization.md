# Phase_07 — Latency Optimization

> **Source dimensions:** [`Phase_06_Host_Agent.md`](Phase_06_Host_Agent.md), [`../04_Latency/`](../04_Latency/) (entire family — 11 chapters), [`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md) (C13), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P07.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P07.
> **Phase targets:** R-09 (allocation-free hot path) + R-13 (anti-bluff p999 latency targets).
> **Cross-links:** [`Phase_08_Audio_Surround.md`](Phase_08_Audio_Surround.md), [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_07 takes the streaming pipeline from Phase_04's relaxed Phase_04-target (p999 ≤ 25 ms) to the **canonical p999 ≤ 8 ms** per [Insight #2](../03_Architecture/12_Latency_Engineering_Overview.md). The optimisation chain involves all the Latency-family submodules: helix-rtos (SCHED_FIFO + cgroup pinning), helix-gpu-direct (GPUDirect RDMA send), helix-iouring + helix-xdp (kernel-bypass send), helix-shm (zero-copy across the pipeline), helix-allocator (zero-alloc hot path), helix-bench (the measurement instrument), helix-mempool + helix-lockfree (the underlying primitives).

After Phase_07, end-user-perceived gameplay latency hits the documented Insight #2 floor — p999 input-to-photons within 8 ms of fiber-quality network conditions.

The Phase is **measurement-driven** — every optimisation is verified against helix-bench's HDR histogram + Mann-Whitney U significance test before merge.

---

## 2. Prerequisites

- Phase_06 complete + signed off — host agent operational with relaxed-target latency.
- All Latency-family submodules at v1.0.0 from Phase_02.
- Operator-provisioned hardware: PREEMPT_RT kernel (recommended), CAP_SYS_NICE in container, AF_XDP-capable NIC, GPUDirect-capable GPU + NIC pair.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P07.T01   | Activate helix-rtos SCHED_FIFO promotion on hot-path goroutines | 4      |
| P07.T02   | Pin host CPU isolation (`isolcpus=` boot param)               | 3        |
| P07.T03   | Activate helix-gpu-direct GPUDirect RDMA send                 | 5        |
| P07.T04   | Activate helix-iouring + helix-xdp kernel-bypass send         | 5        |
| P07.T05   | Verify helix-shm zero-copy via runtime/trace                  | 4        |
| P07.T06   | Activate helix-allocator ModeStrict on hot-path goroutines    | 4        |
| P07.T07   | Reflex round-trip echo activation                              | 3        |
| P07.T08   | DSCP / L4S marking verified end-to-end                         | 3        |
| P07.T09   | Per-stage HDR histograms verified within budget                | 5        |
| P07.T10   | GOCACHEPROG remote build cache integrated                      | 3        |
| P07.T11   | p999 ≤ 8 ms canonical Challenges scenario passes              | 4        |
| P07.T12   | Phase_07 acceptance review                                     | 2        |

12 tasks; ~45 subtasks.

---

## 4. Task Details

### 4.1 P07.T01 — helix-rtos SCHED_FIFO

Per [helix-rtos descriptor §2](../06_Submodules/per-submodule/helix-rtos.md):

- helix-pipeline's encoder dispatcher goroutine: priority 50.
- helix-input's controller poll loop: priority 50.
- helix-transport's send goroutine: priority 50.
- Verify via `chrt -p $TID` post-deploy.

### 4.2 P07.T02 — CPU isolation

Operator's host kernel cmdline includes `isolcpus=4-7 nohz_full=4-7 rcu_nocbs=4-7` to dedicate CPUs 4-7 to HelixPlay. helix-rtos pins to this set.

### 4.3 P07.T03 — GPUDirect

Per [helix-gpu-direct descriptor §4](../06_Submodules/per-submodule/helix-gpu-direct.md):

- Verify nvidia-peermem kernel module loaded.
- Verify BIOS Above-4G + ACS-disabled per per-deployment checklist (§9.4).
- Wire GPUDirect path: helix-encoder GPU output → helix-gpu-direct.RegisterRDMA → helix-transport.

### 4.4 P07.T04 — Kernel-bypass

Per [helix-iouring §2 + helix-xdp §2](../06_Submodules/per-submodule/helix-iouring.md):

- io_uring SQE submission for batched UDP sends.
- AF_XDP zerocopy for high-throughput.
- Fallback to standard sendmsg when not supported.

### 4.5 P07.T05 — Zero-copy verification

`runtime/trace` capture during a 30-second session; assert no `mallocgc` events on the hot path.

### 4.6 P07.T06 — Allocator strict mode

Activate ModeStrict on the hot-path goroutines per [helix-allocator descriptor §2](../06_Submodules/per-submodule/helix-allocator.md). After 2 release cycles of clean ModeReport, escalate to ModeStrict.

### 4.7 P07.T07 — Reflex echo

Per [helix-input descriptor §2.4](../06_Submodules/per-submodule/helix-input.md). Server-side echo + client-side measurement.

### 4.8 P07.T08 — DSCP / L4S

Per [helix-network descriptor §2](../06_Submodules/per-submodule/helix-network.md). Verify end-to-end across the operator's LAN.

### 4.9 P07.T09 — HDR histograms

For each stage of helix-pipeline, verify the per-stage p999 latency is within the documented S05 §9.2 budget.

### 4.10 P07.T10 — GOCACHEPROG

Per [O01 §10 + S01 §4.6.2](../08_Operations/01_Container_CI_CD.md#10-the-gocacheprog-remote-build-cache). Deploy + connect CI runners.

### 4.11 P07.T11 — Canonical Challenges

Run the `01_minimum_viable_session/06_full_pipeline_end_to_end.scenario` with stricter Phase_07 thresholds. p999 ≤ 8 ms or fail.

### 4.12 P07.T12 — Acceptance

Operator signoff.

---

## 5. Subtask Catalogue

45 subtasks across 12 tasks. Per-task subtask listings:

**P07.T01 — helix-rtos SCHED_FIFO (4)**: T01.S01 helix-pipeline encoder dispatcher priority 50; T01.S02 helix-input controller poll loop priority 50; T01.S03 helix-transport send goroutine priority 50; T01.S04 verify via `chrt -p $TID` post-deploy.

**P07.T02 — CPU isolation (3)**: T02.S01 host kernel cmdline `isolcpus=4-7 nohz_full=4-7 rcu_nocbs=4-7`; T02.S02 helix-rtos pins to isolated set; T02.S03 verify via `cat /proc/cmdline` + `taskset -cp $TID`.

**P07.T03 — GPUDirect (5)**: T03.S01 nvidia-peermem kernel module loaded; T03.S02 BIOS Above-4G + ACS-disabled per checklist; T03.S03 helix-encoder GPU output → helix-gpu-direct.RegisterRDMA; T03.S04 GPUDirect path NIC → operator's network; T03.S05 fallback to standard path on GPUDirect-init failure.

**P07.T04 — Kernel-bypass (5)**: T04.S01 io_uring SQE submission for batched UDP sends; T04.S02 AF_XDP zerocopy for high-throughput; T04.S03 fallback to standard sendmsg; T04.S04 per-NIC AF_XDP capability detection; T04.S05 io_uring-vs-sendmsg p999 comparison.

**P07.T05 — Zero-copy verification (4)**: T05.S01 runtime/trace capture during 30 s session; T05.S02 assert no `mallocgc` on hot path; T05.S03 helix-shm zero-copy contract verified; T05.S04 escape-analysis lint integrated.

**P07.T06 — Allocator strict mode (4)**: T06.S01 ModeReport active on hot-path goroutines; T06.S02 2-cycle clean Report period; T06.S03 escalate to ModeStrict; T06.S04 ModeStrict blocks promotion if any allocation detected.

**P07.T07 — Reflex echo (3)**: T07.S01 server-side echo handler; T07.S02 client-side measurement; T07.S03 round-trip latency export to helix-bench histograms.

**P07.T08 — DSCP / L4S (3)**: T08.S01 DSCP marking on send path; T08.S02 L4S signaling per [helix-network](../06_Submodules/per-submodule/helix-network.md); T08.S03 end-to-end verification via packet capture across operator's LAN.

**P07.T09 — HDR histograms (5)**: T09.S01 per-stage helix-pipeline latency export; T09.S02 input-capture stage budget verified; T09.S03 encode stage budget verified; T09.S04 transport stage budget verified; T09.S05 client-render stage budget verified.

**P07.T10 — GOCACHEPROG (3)**: T10.S01 remote build cache deployed per [O01 §10](../08_Operations/01_Container_CI_CD.md); T10.S02 CI runners connected; T10.S03 cache-hit rate Prometheus dashboard.

**P07.T11 — Canonical Challenges (4)**: T11.S01 `01_minimum_viable_session/06_full_pipeline_end_to_end.scenario` with Phase_07 thresholds; T11.S02 p999 ≤ 8 ms gate; T11.S03 per-region scenario fan-out; T11.S04 baseline replacement per [T11 §8b](../07_Testing/11_Challenges.md).

**P07.T12 — Acceptance (2)**: T12.S01 operator + latency-engineering signoff; T12.S02 Phase_07 closure ticket on operator's project board.

---

## 6. Exit Criteria

- [ ] helix-rtos SCHED_FIFO promotion verified.
- [ ] CPU isolation operational.
- [ ] GPUDirect RDMA path active + verified.
- [ ] Kernel-bypass send path active + verified.
- [ ] Zero allocations on hot path (verified via runtime/trace).
- [ ] helix-allocator ModeStrict activated after warm-up cycles.
- [ ] Reflex round-trip echo working.
- [ ] DSCP / L4S marking propagates end-to-end.
- [ ] Per-stage p999 within budget per S05 §9.2 tables.
- [ ] GOCACHEPROG remote cache integrated.
- [ ] Canonical Challenges scenario p999 ≤ 8 ms.
- [ ] Operator signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP07-01  | PREEMPT_RT kernel not available on operator's host                   | Stock kernel works (slightly worse jitter); helix-rtos detects + adapts. |
| RP07-02  | GPUDirect BIOS prerequisites not satisfied                           | Operator pre-deployment checklist per helix-gpu-direct §9.4.           |
| RP07-03  | Hot-path allocation discovered by ModeStrict in production           | Strict mode escalation only after 2 cycles of clean Report mode.       |
| RP07-04  | DSCP marking stripped at operator's egress                           | Operator-side firewall config; verifiable with packet capture.         |

---

## 8. Cross-Family Dependencies

- C13 §6 (Latency Engineering Overview) — Insight #2 codification.
- All Latency-family submodules (Phase_02 graduations).
- helix-bench (Phase_02) — measurement.
- T06 (Benchmarking) + T11 (Challenges) — verification.

---

## 9. Acceptance Criteria

Constitution §16 signoff + §6 exit criteria.

---

## 10. The Phase_07 Calendar

~ 8 weeks. Operator-side capacity: 5 engineers (2× systems engineers for kernel-bypass + RT-OS tuning; 1× GPU engineer for GPUDirect; 1× allocator/escape-analysis specialist; 1× perf engineer for measurement + benchstat). Calendar gated by Phase_06 + unblocks Phase_08.

---

## 10a. Per-Phase Detailed Task Acceptance Criteria

Per-task acceptance criteria expanding §4 task details:

### 10a.1 SCHED_FIFO acceptance (P07.T01)

- All 3 hot-path goroutines (helix-pipeline encoder dispatcher + helix-input controller poll + helix-transport send) verified at SCHED_FIFO priority 50 via `chrt -p $TID`.
- Per-goroutine TID observable via `runtime.LockOSThread()` + thread-id export → Prometheus `helix_rtos_sched_fifo_tid` gauge.
- PREEMPT_RT detection probe: helix-rtos.SCHED_FIFOCapability returns true on PREEMPT_RT kernels, false on stock; behaviour adapts per RP07-01.
- CAP_SYS_NICE container capability granted (per Constitution §11.5.3 named exception).

### 10a.2 CPU isolation acceptance (P07.T02)

- Operator's host kernel cmdline contains `isolcpus=4-7 nohz_full=4-7 rcu_nocbs=4-7`; verified via `cat /proc/cmdline` + automated probe.
- helix-rtos pins all SCHED_FIFO goroutines to CPUs 4-7 via taskset/cpuset.
- Per-CPU jitter measurement (cyclictest-equivalent) shows p999 ≤ 30 µs on isolated cores vs ≥ 200 µs on non-isolated.

### 10a.3 GPUDirect acceptance (P07.T03)

- nvidia-peermem kernel module loaded + verified.
- BIOS Above-4G + ACS-disabled per [helix-gpu-direct §9.4](../06_Submodules/per-submodule/helix-gpu-direct.md) checklist.
- helix-encoder GPU output → helix-gpu-direct.RegisterRDMA → helix-transport NIC zero-copy path verified.
- Fallback path (host-bounce) verified on GPUDirect-init failure (RP07-02 mitigation).

### 10a.4 Kernel-bypass acceptance (P07.T04)

- io_uring SQE submission verified for batched UDP sends; Prometheus `helix_iouring_sqe_submitted_total` rate matches frame-rate × sub-frames.
- AF_XDP zerocopy for high-throughput verified; per-NIC capability detection respects fallback.
- Standard sendmsg fallback path tested on AF_XDP-incapable NICs.

### 10a.5 Zero-copy verification acceptance (P07.T05)

- runtime/trace capture during 30-second session shows zero `mallocgc` events on hot-path goroutines.
- helix-shm zero-copy contract verified via [helix-shm §3](../06_Submodules/per-submodule/helix-shm.md) probe.
- Escape-analysis lint integrated into CI (`go build -gcflags="-m"` outputs scanned for "escapes to heap" on hot-path functions).

### 10a.6 Allocator strict mode acceptance (P07.T06)

- ModeReport active on hot-path goroutines for ≥ 2 release cycles.
- Zero ModeReport violations across 2 cycles → escalation to ModeStrict gated by operator's compliance officer signoff.
- ModeStrict enforces panic-on-allocation per [helix-allocator §3](../06_Submodules/per-submodule/helix-allocator.md).

---

## 11. Per-Phase Observability Catalogue

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_pipeline_stage_latency_seconds` | histogram | tenant, stage | p999 within stage budget |
| `helix_input_to_photons_seconds` | histogram | tenant, region | p999 ≤ 8 ms |
| `helix_rtos_sched_fifo_priority` | gauge | tid, goroutine | 50 (verified) |
| `helix_gpu_direct_rdma_send_total` | counter | tenant, nic | rate ≥ frame-rate |
| `helix_iouring_sqe_submitted_total` | counter | tenant | rate ≥ batch-rate |
| `helix_xdp_zerocopy_send_total` | counter | tenant, nic | rate ≥ frame-rate |
| `helix_allocator_mode_violations_total` | counter | tenant, goroutine | 0 in ModeStrict |
| `helix_reflex_round_trip_seconds` | histogram | tenant | p999 ≤ 2 ms |
| `helix_dscp_marking_total` | counter | tenant, dscp | per-priority count |
| `helix_gocacheprog_cache_hits_total` | counter | runner | hit rate ≥ 80% |

### 11.2 Grafana dashboards

- **Per-tenant Latency Floor** — p999 input-to-photons + per-stage breakdown.
- **Per-region Reflex round-trip** — p999 + per-NIC variance.
- **Per-CI-runner build-cache health** — cache-hit rate + per-target wall-time.
- **Per-tenant Allocator compliance** — ModeReport / ModeStrict status + per-goroutine violation history.

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Input-to-photons latency | End-to-end p999 over 30-second window | ≤ 8 ms | per-session |
| Per-stage budget | Each pipeline stage within S05 §9.2 budget | 100% | per-session |
| Allocator hot-path purity | Zero allocations in ModeStrict goroutines | 100% | per-deploy |
| Reflex echo round-trip | Server-side echo + client-side measurement | p999 ≤ 2 ms | 7-day rolling |
| GPUDirect availability | Sessions on GPUDirect-eligible hardware | ≥ 95% | per-session |
| GOCACHEPROG cache-hit rate | Per-PR cache-hit rate | ≥ 80% | 7-day rolling |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixLatency/docs/runbook/phase07-operations.md` covering: kernel-bypass pre-deployment hardware checklist, GPUDirect BIOS prerequisites, ModeStrict escalation playbook (RP07-03), DSCP-marking-stripped-at-egress diagnostics (RP07-04), per-NIC AF_XDP capability matrix, GOCACHEPROG cache-warming procedure, Reflex round-trip echo diagnostics.

---

## 13a. Per-Phase Risk Mitigation Detail

### 13a.1 RP07-01 — PREEMPT_RT kernel not available

**Detection:** helix-rtos.SCHED_FIFOCapability returns false; per-CPU jitter measurement shows p999 > 200 µs.

**Mitigation:** Stock kernel works at slightly worse jitter; helix-rtos detects + adapts; per-deployment p999 budget adjusted.

**Remediation:** Operator's runbook §2.3 — PREEMPT_RT kernel deployment procedure; per-distro kernel package selection (rt-kernel package or custom build).

### 13a.2 RP07-02 — GPUDirect BIOS prerequisites not satisfied

**Detection:** helix-gpu-direct.RegisterRDMA fails with PCIe topology error.

**Mitigation:** Operator pre-deployment checklist per [helix-gpu-direct §9.4](../06_Submodules/per-submodule/helix-gpu-direct.md); BIOS Above-4G + ACS-disabled probe at host-startup.

**Remediation:** Operator's runbook §3.5 — BIOS settings reconfiguration; per-motherboard-vendor procedure (Supermicro / ASUS / Gigabyte).

### 13a.3 RP07-03 — Hot-path allocation discovered by ModeStrict in production

**Detection:** ModeStrict panic on hot-path goroutine.

**Mitigation:** ModeStrict escalation only after 2 cycles of clean ModeReport; per-cycle compliance officer review.

**Remediation:** Operator's runbook §4.6 — emergency rollback to ModeReport; offending allocation site identified via panic stack trace + escape-analysis review.

### 13a.4 RP07-04 — DSCP marking stripped at operator's egress

**Detection:** Per-region packet capture shows DSCP zero on outbound traffic.

**Mitigation:** Operator-side firewall config audit; per-region transit provider DSCP preservation SLA.

**Remediation:** Operator's runbook §5.2 — per-firewall DSCP preservation procedure; per-transit-provider SLA escalation if DSCP stripping persists.

### 13a.5 p999 latency budget regression

**Detection:** Canonical Challenges scenario reports p999 > 8 ms.

**Mitigation:** Per-stage HDR histogram surface stage-specific regression; helix-bench Mann-Whitney U change-point detection per [T11 §6](../07_Testing/11_Challenges.md).

**Remediation:** Operator's runbook §6.4 — per-stage budget audit; regression bisection via git-bisect against per-stage benchmark.

---

## 14. Implementation Considerations

### 14.1 PREEMPT_RT vs stock kernel

Recommended: PREEMPT_RT for jitter-floor; stock kernel works at slightly worse jitter. helix-rtos detects + adapts (RP07-01 mitigation). Operator's choice — PREEMPT_RT is per-region capex decision.

### 14.2 GPUDirect BIOS prerequisites

Above-4G decoding + ACS-disabled per [helix-gpu-direct §9.4](../06_Submodules/per-submodule/helix-gpu-direct.md). Pre-deployment checklist in operator's runbook §1. Without these BIOS settings, GPUDirect path silently degrades to host-bounce.

### 14.3 ModeStrict escalation timing

ModeStrict is escalated only after 2 release cycles of clean ModeReport (per Phase_07 P07.T06). Premature escalation risks blocking legitimate code paths in production. Operator's compliance officer signs the escalation per [helix-allocator §10](../06_Submodules/per-submodule/helix-allocator.md).

### 14.4 DSCP marking at operator's egress

DSCP markings are sometimes stripped at the operator's edge firewall (RP07-04). Verifiable via packet capture; operator-side firewall config remediation in runbook §4.

---

## 15. Phase_07 Cost Estimation

Phase_07 is **operational tuning** — no incremental per-tenant cost beyond Phase_04 streaming baseline. The capex is operator's PREEMPT_RT-capable + GPUDirect-capable hardware procurement (already in Phase_00 hardware checklist).

GOCACHEPROG remote build cache adds ~$50 / month / region MinIO storage + cuts operator's CI cost by ~30% via cache hits (per [O01 §10](../08_Operations/01_Container_CI_CD.md) cost model).

---

## 16. Cross-Mirror Parity Verification

Phase_07 closure verification per the [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern.

---

## 16a. Per-Stage Latency Budget Tables

The canonical p999 ≤ 8 ms input-to-photons floor (Insight #2, [C13 §6](../03_Architecture/12_Latency_Engineering_Overview.md)) decomposes into per-stage budgets per [S05 §9.2](../06_Submodules/05_Per_Submodule/05_per_submodule_helix-pipeline.md):

### 16a.1 Per-stage budget breakdown (4K60, NVIDIA RTX 4090 reference)

| Stage | p50 | p99 | p999 | Cumulative p999 |
|-------|----:|----:|-----:|----------------:|
| Input event capture (helix-input) | 0.2 ms | 0.4 ms | 0.6 ms | 0.6 ms |
| Input transport (client → host) | 0.3 ms | 0.6 ms | 1.0 ms | 1.6 ms |
| Game-engine input frame | 0.5 ms | 1.0 ms | 1.5 ms | 3.1 ms |
| Capture (helix-capture, GPU-direct) | 0.1 ms | 0.2 ms | 0.3 ms | 3.4 ms |
| Encode (helix-encoder, NVENC HEVC) | 1.5 ms | 2.5 ms | 3.5 ms | 6.9 ms |
| Encode→Transport handoff (helix-shm) | 0.05 ms | 0.1 ms | 0.2 ms | 7.1 ms |
| Transport (kernel-bypass UDP send) | 0.2 ms | 0.4 ms | 0.6 ms | 7.7 ms |
| Network propagation (LAN, fiber) | 0.05 ms | 0.1 ms | 0.2 ms | 7.9 ms |
| Decode (client-side hardware) | 0.5 ms | 0.8 ms | 1.2 ms | 9.1 ms* |
| Render + present (client) | 0.5 ms | 1.0 ms | 1.5 ms | 10.6 ms* |

*Per-client p999 floor for "good enough" at 8 ms is operator-tunable; the 8 ms target applies to **server-side completion** (capture → transport-send completed). Client-side decode + render is amortised against display refresh-rate.

### 16a.2 Per-codec latency variance

| Codec | Encode p999 (4K60) | Encode p999 (1080p120) |
|-------|-------------------:|----------------------:|
| H.264 NVENC | 3.0 ms | 1.5 ms |
| HEVC NVENC | 3.5 ms | 1.8 ms |
| AV1 NVENC (Ada Lovelace+) | 4.5 ms | 2.2 ms |
| H.264 QuickSync (Intel Arc) | 4.0 ms | 2.0 ms |
| HEVC QuickSync (Intel Arc) | 4.5 ms | 2.3 ms |
| AV1 QuickSync (Intel Arc) | 5.5 ms | 2.8 ms |

H.264 NVENC is the canonical latency-floor codec. AV1 trades latency for bandwidth efficiency (operator-tunable per session-tier).

### 16a.3 Per-resolution + framerate latency floor

| Resolution × FPS | Per-frame budget | Encode p999 (HEVC NVENC) | Total p999 input-to-photons |
|------------------|-----------------:|-------------------------:|----------------------------:|
| 1080p60 | 16.67 ms | 1.5 ms | ≤ 6 ms |
| 1080p120 | 8.33 ms | 1.8 ms | ≤ 6 ms |
| 1440p120 | 8.33 ms | 2.5 ms | ≤ 7 ms |
| 4K60 | 16.67 ms | 3.5 ms | ≤ 8 ms |
| 4K120 | 8.33 ms | 4.5 ms | ≤ 9 ms |

The 4K60 8 ms target is the canonical Phase_07 acceptance gate. 4K120 with HDR pushes to 9 ms p999 — operator-tunable.

### 16a.4 Per-stage budget enforcement

helix-pipeline emits per-stage HDR histograms continuously; helix-bench aggregates + applies Mann-Whitney U change-point detection. Per-stage budget breach triggers Prometheus alert + per-stage forensics workflow per [§13](../09_Implementation_Phases/Phase_07_Latency_Optimization.md#13-per-phase-operator-runbook).

---

## 16b. Per-Optimization Verification Procedure

### 16b.1 Per-optimization apply-then-measure protocol

Each Phase_07 optimization (P07.T01..T11) follows the same apply-then-measure protocol per [helix-bench §7](../06_Submodules/per-submodule/helix-bench.md):

1. **Baseline capture** — pre-optimization helix-bench run; HDR histogram archived; Mann-Whitney U baseline established with ≥ 10K samples per Insight #2.
2. **Apply optimization** — single-change isolation (no batched changes; per [T06 §8](../07_Testing/06_Benchmarking.md) Mann-Whitney U significance protocol).
3. **Post-optimization capture** — same workload; same hardware; same Challenges scenario.
4. **Statistical comparison** — Mann-Whitney U test for distribution shift; Cohen's d effect size; per-stage budget compliance check.
5. **Acceptance gate** — p ≤ 0.01 for improvement; effect size ≥ 0.2 for material; per-stage budget within § S05 §9.2 envelope.
6. **Regression bisection** — if any per-stage budget violated post-optimization, git-bisect against per-stage benchmark to identify the offending change.

### 16b.2 Per-stage benchmarking harness

helix-bench exposes per-stage benchmark targets:

- `BenchmarkInputCapture` — helix-input poll-loop latency.
- `BenchmarkInputTransport` — client-to-host gRPC stream latency.
- `BenchmarkCapture` — helix-capture per-frame capture latency (GPU-direct vs host-bounce).
- `BenchmarkEncode` — helix-encoder per-frame encode latency (per-codec variants).
- `BenchmarkEncodeTransport` — encoder-to-transport handoff via helix-shm.
- `BenchmarkTransportSend` — kernel-bypass UDP send (io_uring + AF_XDP) vs sendmsg fallback.
- `BenchmarkReflexEcho` — Reflex round-trip echo latency.

### 16b.3 Per-optimization expected gains (baseline → post)

| Optimization | Baseline p999 | Post p999 | Gain |
|--------------|--------------:|----------:|-----:|
| SCHED_FIFO promotion | 12 ms | 9 ms | 3 ms |
| CPU isolation (isolcpus=4-7) | 9 ms | 8.5 ms | 0.5 ms |
| GPUDirect RDMA send | 8.5 ms | 8.2 ms | 0.3 ms |
| io_uring batched send | 8.2 ms | 8.0 ms | 0.2 ms |
| AF_XDP zerocopy | 8.0 ms | 7.9 ms | 0.1 ms |
| ModeStrict allocator | 7.9 ms | 7.8 ms | 0.1 ms |
| **Cumulative gain** | 12 ms | 7.8 ms | **4.2 ms** |

The cumulative gain table is approximate; actual gains vary per-hardware. Per-deployment baselines + post measurements are operator-archived per [helix-bench §8](../06_Submodules/per-submodule/helix-bench.md).

---

## 16c. Per-Region Latency Floor Calibration

### 16c.1 Per-region latency target tiers

Per [Phase_12 §4.6](Phase_12_Beta_Launch.md#46-p12t06--performance-baselines) per-region performance baselines:

| Region | Network Floor (RTT) | p999 Input-to-Photons Target | Tier |
|--------|--------------------:|-----------------------------:|:----:|
| NA-East (operator HQ) | 5 ms | ≤ 8 ms | Enterprise |
| NA-West | 10 ms | ≤ 10 ms | Pro |
| EU-Central | 8 ms | ≤ 9 ms | Enterprise |
| EU-West | 12 ms | ≤ 11 ms | Pro |
| AP-East (Tokyo) | 15 ms | ≤ 12 ms | Pro |
| AP-Southeast (Singapore) | 20 ms | ≤ 15 ms | Standard |
| RU-Moscow | 10 ms | ≤ 12 ms | Pro (jurisdictional) |
| BR-Sao Paulo | 25 ms | ≤ 18 ms | Standard |
| IN-Mumbai | 30 ms | ≤ 20 ms | Standard |

### 16c.2 Per-region capacity calibration

Per-region Phase_07 capacity verified at the canonical 8 ms floor for the home-region (operator HQ); other regions calibrated per the per-region budget. Operator's commercial team owns per-region tier-floor agreement with customers.

### 16c.3 Per-jurisdiction latency posture

Russian-jurisdiction operators' latency floor accommodates additional regulatory inspection latency at edge firewall (per RP07-04 DSCP marking stripping risk). Per-jurisdictional p999 target adjusted operator-side.

---

## 17. Anti-Bluff Verification

### 17.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_06_Host_Agent.md`](Phase_06_Host_Agent.md)                |    300+ | 2026-04-30 | predecessor                                      |
| [`../04_Latency/`](../04_Latency/) all 11 chapters                 | 16,666 | 2026-04-30 | architectural sources                           |
| [`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md) | 3,816 | 2026-04-30 | C13 Insight #2                |
| Per-submodule descriptors for the 12 Latency primitives            | ~3,800 | 2026-04-30 | per-submodule API surfaces                      |

### 17.2 Forbidden patterns

Clean.

### 17.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_07 execution + operator signoff.

End of `09_Implementation_Phases/Phase_07_Latency_Optimization.md` — 2026-04-30.
