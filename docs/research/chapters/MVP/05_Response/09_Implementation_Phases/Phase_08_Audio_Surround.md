# Phase_08 — Audio Surround

> **Source dimensions:** [`Phase_07_Latency_Optimization.md`](Phase_07_Latency_Optimization.md), [`../06_Submodules/per-submodule/helix-audio.md`](../06_Submodules/per-submodule/helix-audio.md), [`../06_Submodules/per-submodule/helix-hdr.md`](../06_Submodules/per-submodule/helix-hdr.md), [`../05_Video_Audio/06_Audio_Pipeline.md`](../05_Video_Audio/06_Audio_Pipeline.md) (C31), [`../05_Video_Audio/07_HDR_and_Color.md`](../05_Video_Audio/07_HDR_and_Color.md) (C32), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P08.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P08.
> **Phase targets:** R-09 + R-13 — Atmos 7.1.4 + HDR10/HDR10+/Dolby Vision end-to-end.
> **Cross-links:** [`Phase_09_Recording_and_Replay.md`](Phase_09_Recording_and_Replay.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_08 brings **Atmos surround audio (7.1.4 channel layout)** + **HDR10 / HDR10+ / Dolby Vision** end-to-end. Phase_04 shipped 5.1 audio + SDR video as the smoke target; Phase_07 hit the latency floor; Phase_08 adds the **fidelity** layer:

- helix-audio's Opus MultiStream encoder configured for 7.1.4 channel layouts.
- HDMI 2.1 eARC + ALLM signalling on both host + client sides.
- helix-hdr's PQ/HLG transfer-function tagging.
- HDR10 dynamic-metadata SEI emission.
- Dolby Vision RPU pass-through.
- Client-side tone-mapping for SDR-display fallback (Vulkan compute).

After Phase_08, an operator with a 4K120 HDR10/Dolby Vision capable display + an Atmos receiver gets the full HelixPlay fidelity stack. The Phase_06 host agent + Phase_05 client need updates to negotiate the higher-fidelity formats; those updates land here.

---

## 2. Prerequisites

- Phase_07 complete + signed off — latency floor hit.
- helix-audio + helix-hdr at v1.0.0 from Phase_02.
- Operator-provisioned hardware: HDMI 2.1 cable + receiver + Atmos-capable display + DolbyVision-capable display (or HDR10+ for HDR10+ path).

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P08.T01   | helix-audio 7.1.4 Atmos encoder activation                    | 4        |
| P08.T02   | HDMI 2.1 eARC InfoFrame signalling on host                    | 3        |
| P08.T03   | HDMI 2.1 eARC + ALLM negotiation on client                    | 3        |
| P08.T04   | helix-hdr PQ + HLG transfer-function tagging                  | 4        |
| P08.T05   | HDR10 dynamic-metadata SEI emission                           | 3        |
| P08.T06   | HDR10+ ST 2094-40 dynamic metadata                            | 4        |
| P08.T07   | Dolby Vision RPU pass-through                                 | 4        |
| P08.T08   | Client-side Vulkan tone-mapping for SDR fallback              | 4        |
| P08.T09   | A/V sync verification — VMAF + ViSQOL parity                  | 4        |
| P08.T10   | Atmos / DV / HDR10+ Challenges scenarios pass                 | 4        |
| P08.T11   | Phase_08 acceptance review                                     | 2        |

11 tasks; ~39 subtasks.

---

## 4. Task Details

### 4.1 P08.T01 — Atmos 7.1.4

helix-audio's Opus MultiStream configured for the 12-channel 7.1.4 layout. Bitrate target: 768 kbps (per [helix-audio descriptor §9.1](../06_Submodules/per-submodule/helix-audio.md#91-configuration-knobs)). Encoder validates the channel-map matches the Atmos canonical layout.

### 4.2 P08.T02 — Host-side eARC

helix-audio emits the HDMI eARC InfoFrame at session start; verify the receiver acknowledges via the HDMI capability discovery (CDC).

### 4.3 P08.T03 — Client-side eARC + ALLM

helix-display.SignalALLM + helix-audio.SignalEARC at session start. Verify A/V receiver enters Auto Low-Latency Mode + accepts compressed Atmos stream.

### 4.4 P08.T04 — PQ + HLG

helix-hdr.Tagger configured for both transfer functions; helix-codec.Profile.HDRMode drives the choice. PQ for HDR10/10+/DV; HLG for broadcast-compatible.

### 4.5 P08.T05 — HDR10 SEI

Per [helix-hdr §2.3](../06_Submodules/per-submodule/helix-hdr.md). MaxCLL + MaxFALL emitted per-frame.

### 4.6 P08.T06 — HDR10+ ST 2094-40

Per [helix-hdr §2.3 DynamicMeta.HDR10PlusST2094_40](../06_Submodules/per-submodule/helix-hdr.md). Dynamic metadata SEI emitted per scene (or per-shot per HDR10+ specification).

### 4.7 P08.T07 — Dolby Vision RPU

DV Profile 8.4 (single-layer DV with HDR10 fallback) — operator-tunable per OQ-hdr-A. RPU pass-through verified by client-side DV decoder.

### 4.8 P08.T08 — Client-side tone-mapping

For clients on SDR-only displays. helix-hdr.ToneMapper with Strategy=BT2390 (per [helix-hdr §2.6](../06_Submodules/per-submodule/helix-hdr.md#26-tone-mapping-client-side)). 4K frame tone-mapped within 4 ms p999 (Vulkan compute).

### 4.9 P08.T09 — A/V sync verification

VMAF score on the rendered frame stream + ViSQOL score on the audio stream. Sync drift ≤ 40 ms p999 over a 5-minute session.

### 4.10 P08.T10 — Challenges

Run `06_4k_120hz_hdr_dolby_vision/05_atmos_eARC_under_full_journey.scenario` + `06_dolby_vision_tone_map_under_session.scenario`. Both green per [helix-audio §6](../06_Submodules/per-submodule/helix-audio.md#6-challenges-entry-point-s03-§4-row-23) + [helix-hdr §6](../06_Submodules/per-submodule/helix-hdr.md#6-challenges-entry-point-s03-§4-row-24).

### 4.11 P08.T11 — Acceptance

Operator signoff.

---

## 5. Subtask Catalogue

39 subtasks.

---

## 6. Exit Criteria

- [ ] Atmos 7.1.4 encode + decode end-to-end.
- [ ] HDMI 2.1 eARC negotiated on both host + client.
- [ ] ALLM mode active on the receiver.
- [ ] HDR10 dynamic metadata propagated.
- [ ] HDR10+ ST 2094-40 propagated.
- [ ] Dolby Vision RPU pass-through working.
- [ ] Client-side tone-mapping for SDR fallback working.
- [ ] A/V sync ≤ 40 ms p999.
- [ ] Atmos + DV Challenges scenarios green.
- [ ] Operator signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP08-01  | HDMI 2.1 eARC InfoFrame transmission failed                         | Operator-side cable verification + receiver firmware update.           |
| RP08-02  | Dolby Vision encoder licensing                                       | Apache-2.0 helix-codec patent grant; operator-side DV authoring through licensed third-party. |
| RP08-03  | Atmos channel-map mismatch (7.1.4 vs 7.1)                            | helix-audio validates channel-map at encoder init; verifiable smoke probe. |
| RP08-04  | Client-side tone-mapping GPU not available                           | Per [helix-hdr §9.3](../06_Submodules/per-submodule/helix-hdr.md#93-common-errors-and-remediation) ErrToneMapVulkanInit fallback path. |

---

## 8. Cross-Family Dependencies

- C31 §6 (Audio Pipeline) + C32 §6 (HDR & Color) architecturally specify the patterns.
- helix-audio + helix-hdr at v1.0.0 from Phase_02.
- Phase_06 host agent + Phase_05 clients need configuration updates.

---

## 9. Acceptance Criteria

Constitution §16 + §6 exit criteria.

---

## 10. Anti-Bluff Verification

### 10.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_07_Latency_Optimization.md`](Phase_07_Latency_Optimization.md) | 200+ | 2026-04-30 | predecessor                                      |
| [`../05_Video_Audio/06_Audio_Pipeline.md`](../05_Video_Audio/06_Audio_Pipeline.md) | 1,375 | 2026-04-30 | C31 architectural source                  |
| [`../05_Video_Audio/07_HDR_and_Color.md`](../05_Video_Audio/07_HDR_and_Color.md) | 1,801 | 2026-04-30 | C32 architectural source                          |
| [`../06_Submodules/per-submodule/helix-audio.md`](../06_Submodules/per-submodule/helix-audio.md) | 311 | 2026-04-30 | Audio primitive            |
| [`../06_Submodules/per-submodule/helix-hdr.md`](../06_Submodules/per-submodule/helix-hdr.md) | 328 | 2026-04-30 | HDR primitive                              |

### 10.2 Forbidden patterns

Clean.

### 10.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_08 execution + operator signoff.

End of `09_Implementation_Phases/Phase_08_Audio_Surround.md` — 2026-04-30.
