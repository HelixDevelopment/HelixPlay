# Phase_04 — Streaming Pipeline

> **Source dimensions:** [`Phase_03_Backend_Services.md`](Phase_03_Backend_Services.md), [`../06_Submodules/per-submodule/helix-pipeline.md`](../06_Submodules/per-submodule/helix-pipeline.md), [`../06_Submodules/per-submodule/helix-transport.md`](../06_Submodules/per-submodule/helix-transport.md), [`../05_Video_Audio/11_Go_Pipeline_Implementation.md`](../05_Video_Audio/11_Go_Pipeline_Implementation.md) (C36), [`../05_Video_Audio/12_Network_Transport.md`](../05_Video_Audio/12_Network_Transport.md) (C37), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P04.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P04.
> **Phase targets:** R-01 + R-09 + R-13 + R-18 — first phase to stand up the encode-side pipeline + network transport end-to-end.
> **Cross-links:** [`Phase_05_Clients.md`](Phase_05_Clients.md), [`Phase_06_Host_Agent.md`](Phase_06_Host_Agent.md), [`Phase_07_Latency_Optimization.md`](Phase_07_Latency_Optimization.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_04 brings the **streaming pipeline** online: helix-pipeline (the depth-4 orchestrator) + helix-transport (the depth-4 RTP/SRTP/ICE/QUIC + io_uring/XDP send-path). Both submodules are graduated in Phase_02; Phase_04 wires them into a working host-agent → client end-to-end stream. The Phase consumes helix-network (DSCP / L4S / jitter buffer) + helix-abr (ABR ladder + GCC + FEC) + helix-dualpath (NAL feed split) + helix-encoder + helix-capture + helix-shm + helix-audio — graduated in Phase_02 — plus the Phase_03 backend services (Vault, NATS, CockroachDB, Coturn).

After Phase_04, a single end-to-end gameplay session runs from a real game's GPU → NVENC → helix-pipeline → helix-transport → real network → client (placeholder; full client comes in Phase_05). The 4K120 + Atmos + HDR target is **not** yet hit (Phase_07 + Phase_08 finish that); Phase_04's success is "1080p60 SDR works end-to-end against a placeholder client."

The Phase is the largest single integration milestone of the synthesis programme.

---

## 2. Prerequisites

- Phase_03 complete + signed off — production backing services live.
- Phase_02 graduated submodules: helix-pipeline, helix-transport, helix-network, helix-abr, helix-dualpath, helix-encoder, helix-capture, helix-shm, helix-audio, helix-codec.
- A test "real game" workload (e.g. a Godot or Unity demo running fullscreen on the host).
- Operator-provisioned NIC with AF_XDP zerocopy + GPU with NVENC.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P04.T01   | Wire helix-pipeline orchestrator                              | 8        |
| P04.T02   | Wire helix-transport RTP/SRTP/ICE/QUIC                        | 7        |
| P04.T03   | Integrate helix-encoder NVENC encode path                     | 5        |
| P04.T04   | Integrate helix-capture per-OS path                            | 6        |
| P04.T05   | Integrate helix-network DSCP + L4S + jitter buffer             | 4        |
| P04.T06   | Integrate helix-abr congestion control + FEC                   | 5        |
| P04.T07   | Integrate helix-dualpath NAL split                             | 4        |
| P04.T08   | Integrate helix-shm zero-copy frame pool                       | 4        |
| P04.T09   | Integrate helix-audio Opus + multi-channel encode              | 5        |
| P04.T10   | End-to-end smoke test — 1080p60 SDR for 30 seconds            | 4        |
| P04.T11   | Per-stage observability (OTel spans + Prometheus metrics)      | 5        |
| P04.T12   | Challenges integration: 01_minimum_viable_session/06_full_pipeline_end_to_end | 5    |
| P04.T13   | Phase_04 acceptance review                                     | 2        |

13 tasks; ~64 subtasks.

---

## 4. Task Details

### 4.1 P04.T01 — helix-pipeline orchestrator

Per [helix-pipeline descriptor §2](../06_Submodules/per-submodule/helix-pipeline.md#2-public-api-surface):

- Wire the pipeline.New(cfg) at the host-agent entry-point.
- Configure each per-stage (capture / encoder / dualpath / record / audio / hdr / abr / thermal / bench).
- Wire the Sink interface to helix-transport.
- Subscribe to pipe.Events() for observability.
- Add OTel span around Pipeline.Start + Pipeline.Stop.
- Verify pipe.Stats() reports expected metrics.

### 4.2 P04.T02 — helix-transport

Per [helix-transport descriptor §2](../06_Submodules/per-submodule/helix-transport.md#2-public-api-surface):

- Configure WebRTC stack (Pion v4) for the primary path.
- Configure QUIC stack (quic-go) for the alternative path.
- Wire ICE agent with operator-supplied STUN + TURN.
- Wire SRTP key exchange via DTLS handshake.
- For kernel-bypass send: configure helix-iouring + helix-xdp sidecar.
- Verify 100 Mbps sustained send rate end-to-end.

### 4.3 P04.T03 — helix-encoder NVENC integration

- Configure NVENC encoder with helix-codec.PickProfile output.
- Verify encoder pool reuse pattern (NVENC consumer-grade session limits).
- 4K120 H.264 + HEVC verified at the §10 smoke test.

### 4.4 P04.T04 — helix-capture per-OS

For Phase_04's primary target (Linux X11 + Wayland):

- DXGI capture (Windows): defer to Phase_05's Windows-host expansion.
- Metal (macOS): defer.
- X11 SHM: implement; 4K120 capture verified.
- PipeWire (Wayland): implement; verified against Wayland compositor.

### 4.5 P04.T05 — helix-network

Configure DSCP marking (EF for gameplay, AF41 for video, AF31 for audio); verify TOS bits propagate end-to-end across the operator's LAN.

L4S ECT(1) marking enabled where kernel + NIC support it.

Jitter buffer target 40 ms playout delay (per [helix-network §9.1](../06_Submodules/per-submodule/helix-network.md#91-configuration-knobs)).

### 4.6 P04.T06 — helix-abr

Wire helix-abr.Controller into the pipeline. Verify rung-descent via Toxiproxy-injected impairment; verify rung-ascent on impairment removal.

### 4.7 P04.T07 — helix-dualpath

Wire helix-dualpath.Splitter; verify both rungs (stream + record) are populated. Phase_09 will fully exercise the record rung; Phase_04's verification is just that the splitter doesn't drop frames.

### 4.8 P04.T08 — helix-shm

Wire helix-shm.Pool into the capture → encoder handoff path; verify zero-copy via runtime/trace.

### 4.9 P04.T09 — helix-audio

Wire Opus MultiStream encoder; verify 5.1 (default) at 256 kbps. 7.1.4 Atmos lands in Phase_08.

### 4.10 P04.T10 — End-to-end 30-second smoke

The Phase_04 Definition of Done's flagship test:

1. Start a Godot demo on the host.
2. Operator opens a Phase_05-placeholder client (a simple Pion WebRTC consumer).
3. Verify 1080p60 SDR streams for 30 seconds.
4. p999 end-to-end latency ≤ 25 ms (relaxed for Phase_04; Phase_07 tightens to ≤ 8 ms).
5. Audio + video sync within 100 ms.
6. Zero leak (memory + fd + goroutine) over the 30 s window.

### 4.11 P04.T11 — Per-stage observability

Every of the 9 helix-* submodules in the pipeline must emit OTel spans + Prometheus metrics per [O04 §4](../08_Operations/04_Observability_and_Events.md#4-per-submodule-metrics-catalogue). The Grafana per-submodule dashboard renders real session data.

### 4.12 P04.T12 — Challenges integration

The `01_minimum_viable_session/06_full_pipeline_end_to_end.scenario.yaml` runs against a fresh Phase_04 deployment. VMAF + ViSQOL baseline recording happens here (operator-signed); subsequent runs compare against this Phase_04 baseline.

### 4.13 P04.T13 — Acceptance review

Operator signoff per Constitution §16 + the §6 exit criteria.

---

## 5. Subtask Catalogue

64 subtasks across 13 tasks; per-task discrete `[P04.Tyy.Szz]` tickets per [O05](../08_Operations/05_Tracking_GitHub_GitLab.md). Sub-categories: capture (T01), encode (T02), pipeline (T03..T05), transport (T06..T08), control plane (T09..T10), end-to-end smoke (T11..T12), acceptance (T13). Each subtask carries: function entry-point + verification probe + observability emission + Challenges scenario reference per [T11 §6](../07_Testing/11_Challenges.md).

---

## 6. Exit Criteria

- [ ] helix-pipeline.New + .Start operational.
- [ ] helix-transport WebRTC + QUIC paths both work.
- [ ] 1080p60 SDR end-to-end smoke test passes for 30 s.
- [ ] p999 latency ≤ 25 ms (Phase_04-relaxed target).
- [ ] Zero leak over the smoke window.
- [ ] OTel spans + metrics flowing to Grafana for every submodule.
- [ ] Challenges baseline recorded for `06_full_pipeline_end_to_end`.
- [ ] Operator signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP04-01  | helix-pipeline depth-4 contract gap with one of 9 sibling deps      | Pre-Phase_04 contract-verification probe per Phase_02 §5b.            |
| RP04-02  | NVENC consumer-grade session limit (5) hit on multi-session         | Pool reuse pattern per [helix-encoder §9.1](../06_Submodules/per-submodule/helix-encoder.md#91-configuration-knobs); operator may upgrade to Quadro / Tesla SKU. |
| RP04-03  | LAN firewall blocks UDP dynamic ports                                | Per [O03 §9d](../08_Operations/03_Service_Discovery_and_Ports.md#9d-operator-side-firewall-setup) firewall config. |
| RP04-04  | ICE NAT traversal fails behind symmetric NAT                         | Operator's TURN server (Phase_03 P03.T05) is the relay fallback.       |
| RP04-05  | Audio + video drift on 30 s smoke                                    | T11 OTel observability surfaces drift; sync logic in helix-pipeline §6 review.  |
| RP04-06  | Challenges baseline-record fails because of operator key-mismatch    | Phase_00 P00.T03 operator signing keys verified; Phase_04 baseline recording uses the same keys. |

---

## 8. Cross-Family Dependencies

| Source                                                                                 | Reference                                                              |
|----------------------------------------------------------------------------------------|------------------------------------------------------------------------|
| C36 §8 ([Go Pipeline Implementation](../05_Video_Audio/11_Go_Pipeline_Implementation.md)) | Architectural specification for helix-pipeline.                        |
| C37 §9 ([Network Transport](../05_Video_Audio/12_Network_Transport.md))                | Architectural specification for helix-transport.                       |
| Phase_02 (Core Submodules)                                                              | Every consumed submodule must be at v1.0.0.                             |
| Phase_03 (Backend Services)                                                             | Backing services live.                                                  |
| T11 (Challenges) + S03                                                                   | Baseline recording happens in Phase_04.                                |

---

## 9. Acceptance Criteria

Constitution §16 signoff + §6 exit criteria.

---

## 10. The Phase_04 Calendar

Operator's expected duration:

| Week | Activity                                                         |
|------|------------------------------------------------------------------|
| 1    | T01–T02 (helix-pipeline + helix-transport wiring)                |
| 2    | T03–T04 (encoder + capture integration)                          |
| 3    | T05–T08 (network + ABR + dualpath + shm)                         |
| 4    | T09 (audio)                                                       |
| 4    | T10 (end-to-end smoke)                                            |
| 5    | T11–T12 (observability + Challenges)                              |
| 5    | T13 (acceptance)                                                   |

Operator-side capacity: 4–6 engineers covering Go + GPU + WebRTC + networking. ~ 5 weeks.

---

## 10a. Per-Phase Detailed Task Acceptance Criteria

### 10a.1 Capture pipeline acceptance (P04.T01)

- helix-capture initializes against operator-host display (Linux: KMS/DRM via libdrm; Windows: Desktop Duplication API; macOS: ScreenCaptureKit on macOS 12.3+).
- Per-frame capture latency p999 ≤ 2 ms.
- Capture format negotiation (RGB / BGRA / NV12) verified per platform; zero per-frame conversion when GPU-direct path active.
- Hot-plug display change detected + handled gracefully without session interruption.

### 10a.2 Encode pipeline acceptance (P04.T02)

- helix-encoder initialised against operator-host GPU; H.264 (baseline + main + high), HEVC (main + main10), AV1 (per-GPU capability) negotiated per session.
- Per-frame encode latency p999 ≤ 4 ms (4K60) on reference NVIDIA RTX 4090 / 3090; ≤ 8 ms on RTX 3060 baseline.
- Per-codec quality target VMAF ≥ 92 at 25 Mbps 4K60.
- Encoder reset on stream parameter change (resolution / fps / codec) within 100 ms.

### 10a.3 Pipeline composition acceptance (P04.T03..T05)

- helix-pipeline.New builds the goroutine topology per [C36 §8](../05_Video_Audio/05_Go_Pipeline_Implementation.md).
- helix-pipeline.Start brings up capture → encoder → transport with per-stage HDR histograms.
- helix-pipeline.Stop graceful shutdown within 1 s p99.
- Per-stage drop counter exposes pipeline backpressure events.

### 10a.4 Transport acceptance (P04.T06..T08)

- helix-transport WebRTC mode: STUN/TURN/ICE negotiation verified; DTLS-SRTP; per-region Coturn integration.
- helix-transport custom-UDP mode: operator-LAN-only; helix-network DSCP marking + L4S signalling.
- Per-session packet-loss < 0.1% over reference network.
- Reconnect-on-disconnect within 2 s p99.

### 10a.5 Control plane acceptance (P04.T09..T10)

- gRPC bidirectional stream over HTTP/3 (QUIC); per-tenant TLS 1.3-only.
- Control messages (input + reflex echo + ABR signals) interleaved with media stream without head-of-line blocking.
- Brotli compression on control-plane gRPC bodies per [C06 §6](../03_Architecture/05_Data_Plane.md).

### 10a.6 End-to-end smoke acceptance (P04.T11..T12)

- Synthetic 4K60 SDR + 5.1 audio session for 5 minutes; p999 latency within Phase_04 relaxed budget (≤ 25 ms); zero packet loss; clean teardown.
- helix-bench histogram export + helix-vqa VMAF score green.

---

## 11. Per-Phase Observability Catalogue

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_pipeline_session_active_count` | gauge | tenant | per-tenant concurrency |
| `helix_pipeline_stage_latency_seconds` | histogram | tenant, stage | per-stage Phase_04 budget |
| `helix_capture_frames_total` | counter | tenant, source | rate ≥ frame-rate |
| `helix_encoder_frames_emitted_total` | counter | tenant, codec | matches capture rate |
| `helix_transport_packets_sent_total` | counter | tenant, transport | per-codec packet rate |
| `helix_transport_packets_lost_total` | counter | tenant, transport | < 0.1% per session |
| `helix_abr_rate_changes_total` | counter | tenant, direction | minimal under stable network |
| `helix_session_reconnect_total` | counter | tenant, reason | < 1 / hour |

### 11.2 Grafana dashboards

- **Per-tenant Pipeline Health** — per-stage p99 latency + frame-rate stability.
- **Per-region Transport Quality** — packet-loss + jitter + RTT per-region distribution.
- **Per-codec ABR behaviour** — rate-change frequency + per-codec quality trend.

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Pipeline session start | Time from session-start to first frame emitted | p99 ≤ 2 s | per-session |
| Per-stage latency | Each helix-pipeline stage within Phase_04 relaxed budget | 100% | per-session |
| Encode-to-transport handoff | Time from encoder emit to transport send | p999 ≤ 1 ms | per-frame |
| Transport packet loss | Per-session packet-loss rate | < 0.1% | per-session |
| Session continuity | Sessions completing without reconnect | ≥ 99.5% | 30-day rolling |
| ABR adaptation latency | Time from network degradation to rate adjustment | p99 ≤ 1 s | per-event |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixPipeline/docs/runbook/phase04-operations.md` — covers per-region pipeline deployment, ABR tuning per Phase_04 P04.T07, transport WebRTC vs custom-UDP fallback, codec capability negotiation, end-to-end smoke probe execution.

---

## 13a. Per-Phase Risk Mitigation Detail

### 13a.1 Capture-side framerate drops

**Detection:** `helix_capture_frames_total` rate < target framerate over 5-second rolling window.

**Mitigation:** Per-platform capture path GPU-direct preferred; CPU fallback documented in §14.

**Remediation:** Operator's runbook §2.4 — per-platform capture diagnostic playbook; common causes: display driver mismatch, GPU resource contention, hot-plug event mid-session.

### 13a.2 Encoder reset oscillation

**Detection:** `helix_encoder_reset_total` rate > 1 / minute on stable network.

**Mitigation:** Stream parameter change throttling (≥ 5 s between resets); per-codec capability cache.

**Remediation:** Operator's runbook §3.6 — encoder reset cause analysis; usually mismatched client capability negotiation.

### 13a.3 Transport packet loss spike

**Detection:** `helix_transport_packets_lost_total` rate > 0.5% over 10-second rolling window.

**Mitigation:** Per-region Coturn relay; ABR rate-down on loss; FEC repair packets.

**Remediation:** Operator's runbook §4.3 — per-region transit diagnostic; common causes: ISP peering degradation, client wifi instability, operator's edge firewall mis-config.

### 13a.4 Pipeline backpressure cascade

**Detection:** Per-stage drop counter > 0; downstream stage starvation.

**Mitigation:** Bounded channels with explicit drop policy per [C36 §8](../05_Video_Audio/05_Go_Pipeline_Implementation.md); per-stage circuit breaker.

**Remediation:** Operator's runbook §5.2 — pipeline-stage drop forensics; common causes: encoder GPU saturation, transport NIC saturation, downstream client slow.

### 13a.5 Codec capability negotiation failure

**Detection:** Session-start aborted with codec-mismatch error.

**Mitigation:** Capability fallback chain (AV1 → HEVC → H.264); per-client capability cache.

**Remediation:** Operator's runbook §6.4 — per-client capability re-negotiation; common causes: client codec library outdated, GPU driver outdated.

---

## 14. Implementation Considerations

### 14.1 Phase_04 vs Phase_07 latency targets

Phase_04 targets relaxed p999 ≤ 25 ms (smoke target); Phase_07 takes the floor to ≤ 8 ms. Per-stage budgets are documented in S05 §9.2 budget tables; Phase_04 verifies the relaxed envelope is met.

### 14.2 Codec selection (H.264 vs HEVC vs AV1)

Per [C26 §6](../05_Video_Audio/01_Codec_Selection.md): H.264 baseline (universal); HEVC for HDR + bandwidth efficiency; AV1 for next-gen. Per-client capability negotiation drives selection.

### 14.3 Transport WebRTC vs custom-UDP

Per [C37 §9](../05_Video_Audio/06_Network_Transport.md): WebRTC is the canonical default (firewall-traversal + STUN/TURN/ICE); custom-UDP for operator-LAN-only deployments where WebRTC overhead is undesirable.

### 14.4 ABR aggressiveness tuning

Per [helix-abr §2](../06_Submodules/per-submodule/helix-abr.md). Operator-tunable: `abr_aggressiveness=0.5..1.0`. Conservative (0.5) avoids quality oscillation; aggressive (1.0) maximizes per-frame quality.

---

## 15. Phase_04 Cost Estimation

Per-tenant Phase_04 cost (4K60 baseline):

| Component | Per-Tenant Per-Hour |
|-----------|---------------------|
| GPU encode (HEVC 4K60 ~25 Mbps) | ~$0.50 (operator-amortised hardware) |
| Bandwidth egress (~25 Mbps × 3,600 s = ~11 GB/hour) | ~$0.90 (operator's transit) |
| **Total per-tenant per-hour** | **~$1.40** |

Per-tier pricing model in Phase_10 P10.T04 amortises this baseline.

---

## 16. Cross-Mirror Parity Verification

Phase_04 closure verification per the [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern.

---

## 17. Anti-Bluff Verification

### 11.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_03_Backend_Services.md`](Phase_03_Backend_Services.md)    |    300+ | 2026-04-30 | predecessor                                      |
| [`../06_Submodules/per-submodule/helix-pipeline.md`](../06_Submodules/per-submodule/helix-pipeline.md) | 335 | 2026-04-30 | the orchestrator's API + dep tree |
| [`../06_Submodules/per-submodule/helix-transport.md`](../06_Submodules/per-submodule/helix-transport.md) | 330 | 2026-04-30 | the network transport API           |
| [`../05_Video_Audio/11_Go_Pipeline_Implementation.md`](../05_Video_Audio/11_Go_Pipeline_Implementation.md) | 2,816 | 2026-04-30 | C36 architectural source     |
| [`../05_Video_Audio/12_Network_Transport.md`](../05_Video_Audio/12_Network_Transport.md) | 3,630 | 2026-04-30 | C37 architectural source                       |

### 11.2 Forbidden patterns

Clean.

### 11.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_04 execution + operator signoff.

End of `09_Implementation_Phases/Phase_04_Streaming_Pipeline.md` — 2026-04-30.
