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

45 subtasks; bulk-imported.

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

## 10. Anti-Bluff Verification

### 10.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_06_Host_Agent.md`](Phase_06_Host_Agent.md)                |    300+ | 2026-04-30 | predecessor                                      |
| [`../04_Latency/`](../04_Latency/) all 11 chapters                 | 16,666 | 2026-04-30 | architectural sources                           |
| [`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md) | 3,816 | 2026-04-30 | C13 Insight #2                |
| Per-submodule descriptors for the 12 Latency primitives            | ~3,800 | 2026-04-30 | per-submodule API surfaces                      |

### 10.2 Forbidden patterns

Clean.

### 10.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_07 execution + operator signoff.

End of `09_Implementation_Phases/Phase_07_Latency_Optimization.md` — 2026-04-30.
