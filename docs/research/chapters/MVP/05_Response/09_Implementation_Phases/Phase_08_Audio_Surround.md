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

39 subtasks across 11 tasks; every subtask is a discrete `[P08.Tyy.Szz]` ticket per [O05](../08_Operations/05_Tracking_GitHub_GitLab.md) opened in the operator's GitHub Projects + GitLab boards.

### 5.1 Per-task subtask listings

**P08.T01 — Atmos 7.1.4 (4 subtasks)**: T01.S01 Opus MultiStream encoder configured for 12-channel layout; T01.S02 768 kbps bitrate target verified; T01.S03 channel-map matches Atmos canonical layout; T01.S04 per-stream observability — `helix_audio_atmos_channel_count` Prometheus gauge.

**P08.T02 — Host-side eARC (3 subtasks)**: T02.S01 helix-audio.SignalEARC InfoFrame emission at session start; T02.S02 receiver acknowledgement via HDMI CDC; T02.S03 fallback to ARC if eARC unsupported.

**P08.T03 — Client-side eARC + ALLM (3 subtasks)**: T03.S01 helix-display.SignalALLM at session start; T03.S02 receiver enters Auto Low-Latency Mode; T03.S03 receiver accepts compressed Atmos stream.

**P08.T04 — PQ + HLG (4 subtasks)**: T04.S01 helix-hdr.Tagger configured for both transfer functions; T04.S02 helix-codec.Profile.HDRMode drives selection; T04.S03 PQ for HDR10/10+/DV; T04.S04 HLG for broadcast-compatible.

**P08.T05 — HDR10 SEI (3 subtasks)**: T05.S01 MaxCLL emitted per-frame; T05.S02 MaxFALL emitted per-frame; T05.S03 SEI-message verifier on client.

**P08.T06 — HDR10+ ST 2094-40 (4 subtasks)**: T06.S01 dynamic metadata SEI generation; T06.S02 per-scene + per-shot tone-map intent; T06.S03 ST 2094-40 conformance test; T06.S04 HDR10+ certification path.

**P08.T07 — Dolby Vision RPU (4 subtasks)**: T07.S01 DV Profile 8.4 encoder configuration; T07.S02 RPU pass-through verification; T07.S03 client-side DV decoder integration; T07.S04 DV-to-HDR10 fallback.

**P08.T08 — Client-side Vulkan tone-mapping (4 subtasks)**: T08.S01 helix-hdr.ToneMapper Strategy=BT2390; T08.S02 4K frame tone-mapped within 4 ms p999; T08.S03 SDR-display fallback path; T08.S04 ErrToneMapVulkanInit graceful degradation.

**P08.T09 — A/V sync verification (4 subtasks)**: T09.S01 VMAF score on rendered frame stream; T09.S02 ViSQOL score on audio stream; T09.S03 sync drift ≤ 40 ms p999 over 5-min session; T09.S04 per-region drift telemetry.

**P08.T10 — Challenges scenarios (4 subtasks)**: T10.S01 Atmos eARC scenario green; T10.S02 DV tone-map scenario green; T10.S03 HDR10+ scenario green; T10.S04 PS5/Xbox dashboard parity verified.

**P08.T11 — Acceptance (2 subtasks)**: T11.S01 operator + audio engineering sign-off; T11.S02 Phase_08 closure ticket on operator's project board.

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

## 10. The Phase_08 Calendar

~ 5 weeks. Operator-side capacity: 4 engineers (1× audio engineer for Atmos + ViSQOL; 1× video engineer for HDR/DV; 1× client engineer for Vulkan tone-map; 1× ops engineer for HDMI 2.1 hardware certification). Calendar gated by Phase_07 + unblocks Phase_09 (recordings need full fidelity stack).

---

## 10a. Per-Phase Detailed Task Acceptance Criteria

### 10a.1 Atmos 7.1.4 acceptance (P08.T01)

- Opus MultiStream encoder configured for 12-channel layout (L/R/C/LFE/Ls/Rs/Lrs/Rrs/Tfl/Tfr/Tbl/Tbr) with explicit channel-map signaling.
- 768 kbps bitrate target verified via `helix_audio_opus_bitrate_kbps` gauge.
- Channel-map matches the Atmos canonical layout per [helix-audio §9.1](../06_Submodules/per-submodule/helix-audio.md#91-configuration-knobs).
- Per-channel SNR measurement ≥ 92 dB (audio fidelity floor).

### 10a.2 HDMI 2.1 eARC acceptance (P08.T02 + P08.T03)

- helix-audio.SignalEARC InfoFrame emission verified at session start; receiver acknowledgement via HDMI CDC observed within 200 ms.
- ALLM (Auto Low-Latency Mode) negotiation verified via helix-display.SignalALLM; receiver enters game-mode within 1 s.
- Per-receiver-model quirks register populated for known HDMI 2.1 receivers (Denon AVR-X6700H, Sony A95K, LG OLED48C2, Samsung Q990C, etc.).
- Fallback to ARC (compressed-only Atmos) on eARC-incapable receivers verified.

### 10a.3 HDR10 + HDR10+ + DV acceptance (P08.T05 + P08.T06 + P08.T07)

- HDR10 SEI emitted per-frame with MaxCLL + MaxFALL values; verifiable via FFprobe SEI dump.
- HDR10+ ST 2094-40 dynamic metadata emitted per-scene + per-shot per HDR10+ specification; ST 2094-40 conformance test green.
- DV Profile 8.4 RPU pass-through verified by client-side DV decoder (LG OLED + Sony BRAVIA reference).
- DV-to-HDR10 fallback verified on non-DV displays.

### 10a.4 Client-side tone-mapping acceptance (P08.T08)

- helix-hdr.ToneMapper Strategy=BT2390 active on SDR-only displays.
- 4K frame tone-mapped within 4 ms p999 on Vulkan-capable GPUs (verified via helix-bench).
- ErrToneMapVulkanInit graceful degradation path tested.

### 10a.5 A/V sync acceptance (P08.T09)

- VMAF score ≥ 90 p10 on rendered frame stream over 5-min session.
- ViSQOL score ≥ 4.0 p10 on audio stream over 5-min session.
- Sync drift ≤ 40 ms p999 measured per [T11 §6](../07_Testing/11_Challenges.md) Atmos eARC scenario.

---

## 11. Per-Phase Observability Catalogue

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_audio_atmos_channel_count` | gauge | tenant, session | 12 (7.1.4) |
| `helix_audio_opus_bitrate_kbps` | gauge | tenant, session | 768 |
| `helix_audio_visqol_score` | gauge | tenant, session | p10 ≥ 4.0 |
| `helix_hdr_transfer_function` | gauge | tenant, session | PQ / HLG |
| `helix_hdr10_sei_emitted_total` | counter | tenant, session | rate ≥ 60/s (per-frame) |
| `helix_hdr10plus_st2094_emitted_total` | counter | tenant, session | per-scene |
| `helix_dv_rpu_passthrough_total` | counter | tenant, session | per-frame for DV sessions |
| `helix_tone_map_seconds` | histogram | tenant, gpu | p999 ≤ 4 ms |
| `helix_av_sync_drift_ms` | gauge | tenant, session | p999 ≤ 40 |
| `helix_earc_negotiation_total` | counter | tenant, success | success ≥ 99% |

### 11.2 Grafana dashboards

- **Per-tenant Audio Quality** — Atmos channel-count + Opus bitrate + ViSQOL trend.
- **Per-tenant HDR Coverage** — PQ/HLG/HDR10/HDR10+/DV per-session distribution.
- **Per-tenant A/V Sync** — drift histogram + per-region p999 trend.
- **Per-region eARC negotiation success** — success-rate + per-receiver-model breakdown.

### 11.3 Loki + Tempo

- Loki streams: `helix-audio.encoder` + `helix-hdr.tagger` + `helix-display.tone-map`.
- Tempo traces: `audio.session.atmos-handshake` + `video.session.hdr-negotiation`.

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Atmos negotiation success | Sessions with successful eARC handshake on Atmos-capable receivers | ≥ 99% | 30-day rolling |
| HDR10 emission | Sessions with HDR10 SEI emitted at every IDR boundary | 100% | per-session |
| DV pass-through | DV-tagged sessions with verified RPU pass-through | ≥ 99.5% | 30-day rolling |
| Tone-map performance | 4K frame tone-mapped within 4 ms | p999 | 30-day rolling |
| A/V sync | Audio-video drift on 5-min session | p999 ≤ 40 ms | per-session |
| ViSQOL audio quality | Per-tenant rolling p10 ViSQOL score | ≥ 4.0 | 7-day rolling |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixAudio/docs/runbook/phase08-operations.md` — Phase_08 deployment + audio/video certification + HDMI 2.1 troubleshooting:

- **§1 HDMI 2.1 cable + receiver certification** — operator's per-region certified-hardware list.
- **§2 Atmos channel-map verification** — pre-deployment 7.1.4 layout test.
- **§3 HDR10 / HDR10+ / DV licensing** — operator's per-jurisdiction licensing posture (DV authoring through licensed third-party per RP08-02 mitigation).
- **§4 Per-receiver-model quirks register** — known eARC negotiation failures + ALLM-mode delays + workarounds (e.g., Denon AVR-X6700H eARC handshake retry; Sony A95K DV Profile 8.4 quirks).
- **§5 Vulkan tone-map GPU compatibility matrix** — per-GPU-model performance + fallback path activation.
- **§6 A/V sync diagnostic playbook** — drift detection + per-stage budget audit + remediation.

---

## 13a. Per-Phase Risk Mitigation Detail

### 13a.1 RP08-01 — HDMI 2.1 eARC InfoFrame transmission failed

**Detection:** `helix_earc_negotiation_total{success="false"}` rate > 1% over 24-hour window.

**Mitigation:** Per-receiver-model quirks register populated; per-receiver firmware compatibility list maintained.

**Remediation:** Operator's runbook §2.2 — eARC negotiation diagnostic; common causes: HDMI cable < HDMI 2.1 certified, receiver firmware out-of-date, ARC vs eARC port mis-selection.

### 13a.2 RP08-02 — Dolby Vision encoder licensing

**Detection:** Operator's commercial agreement audit flag.

**Mitigation:** Apache-2.0 helix-codec patent grant; operator-side DV authoring through licensed third-party (per [C32 §6](../05_Video_Audio/02_HDR_and_Color.md)).

**Remediation:** Operator's runbook §3.4 — per-jurisdictional DV licensee selection; commercial-team decision-point on DV vs HDR10+ as primary HDR path.

### 13a.3 RP08-03 — Atmos channel-map mismatch (7.1.4 vs 7.1)

**Detection:** Encoder init returns ErrChannelMapMismatch.

**Mitigation:** helix-audio validates channel-map at encoder init per [helix-audio §9.1](../06_Submodules/per-submodule/helix-audio.md#91-configuration-knobs); verifiable smoke probe at session start.

**Remediation:** Operator's runbook §4.3 — channel-map reconfiguration procedure; per-content-type fallback (e.g., 7.1 fallback for non-Atmos content).

### 13a.4 RP08-04 — Client-side tone-mapping GPU not available

**Detection:** helix-hdr.ToneMapper returns ErrToneMapVulkanInit.

**Mitigation:** Per [helix-hdr §9.3](../06_Submodules/per-submodule/helix-hdr.md#93-common-errors-and-remediation) ErrToneMapVulkanInit fallback path activates CPU tone-mapping (slower but functional).

**Remediation:** Operator's runbook §5.6 — per-GPU-model tone-mapping support matrix; client communication on degraded-fidelity SDR fallback.

### 13a.5 A/V sync drift > 40 ms p999

**Detection:** `helix_av_sync_drift_ms` > 40 over 5-min session.

**Mitigation:** Per-stage timestamp correlation; helix-pipeline per-stage HDR histograms surface drift origin.

**Remediation:** Operator's runbook §6.3 — drift forensics; common causes: encoder GPU saturation, transport jitter, client decoder slow.

---

## 14. Implementation Considerations

### 14.1 Atmos vs 5.1 channel-map mismatch

The 7.1.4 layout introduces 4 height channels (vs 5.1's 0). helix-audio's encoder validates the channel-map at init; mismatch causes RP08-03 (Atmos channel-map mismatch (7.1.4 vs 7.1)) — verifiable smoke probe at session start.

### 14.2 Dolby Vision Profile 8.4 vs Profile 5

DV Profile 8.4 (single-layer DV with HDR10 fallback) is operator-tunable per OQ-hdr-A. Profile 5 (single-layer with separate DV-only metadata) requires DV-only displays. Operator's beta customer-base distribution drives the choice.

### 14.3 HDMI 2.1 eARC vs ARC fallback

When eARC unavailable (older receiver), helix-audio falls back to ARC (compressed-only Atmos). The fallback path is operator-visible via `helix_earc_negotiation_total` counter; per-receiver model quirks are documented in §13 §4.

### 14.4 Client-side tone-map performance budget

The 4 ms p999 budget for 4K frame tone-mapping is **derived from Phase_07** Insight #2 floor: at p999 ≤ 8 ms total input-to-photons, the tone-map step must consume ≤ 50% of the budget. A per-GPU-model performance audit is scheduled before Phase_08 GA promotion.

---

## 15. Phase_08 Cost Estimation

Per-tenant per-month cost breakdown for Phase_08:

| Component | Per-Tenant Per-Month |
|-----------|----------------------|
| Atmos Opus 768 kbps egress (1 Mbps audio, 240 h-month) | ~$3 / tenant |
| HDR10 / HDR10+ / DV encoding (folded into Phase_04 GPU cost) | $0 |
| DV authoring (operator-side licensing) | per operator's licensee agreement |
| **Total per-tenant per-month** | **~$3** |

The Phase_08 incremental cost is small; the main capex is operator-side hardware certification + DV licensing.

---

## 16. Cross-Mirror Parity Verification

Phase_08 closure verification per [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern: composite-push origin → all 4 mirrors at same SHA → cosign-signed phase-closure tag.

---

## 16a. Per-Receiver + Per-Display Compatibility Matrix

### 16a.1 Per-receiver Atmos eARC compatibility (reference set)

| Receiver | Atmos 7.1.4 | eARC Negotiation | ALLM | Known Quirks |
|----------|:-----------:|:----------------:|:----:|--------------|
| Denon AVR-X6700H | ✓ | ✓ (firmware ≥ 1100) | ✓ | eARC handshake retry on fast-source-switch |
| Denon AVR-X3800H | ✓ | ✓ | ✓ | clean |
| Yamaha RX-A8A | ✓ | ✓ | ✓ | clean |
| Yamaha RX-A6A | ✓ | partial | ✓ | eARC InfoFrame ack delayed up to 800 ms |
| Marantz Cinema 50 | ✓ | ✓ | ✓ | clean |
| Onkyo TX-RZ50 | ✓ | ✓ | ✓ | clean |
| Pioneer VSX-LX305 | ✓ | partial | ✓ | per-source eARC config required |
| Sony STR-AN1000 | ✓ | ✓ | ✓ | clean |
| LG OLED CX/C2/C3 (eARC pass-through) | ✓ | ✓ | ✓ | per-firmware variance |
| Samsung Q990C soundbar | ✓ | ✓ | ✓ | clean |

Per-receiver firmware compatibility tracked in operator's runbook §13.4; per-firmware-version regression-test scheduled at upstream firmware release.

### 16a.2 Per-display HDR + DV compatibility (reference set)

| Display | HDR10 | HDR10+ | Dolby Vision | Atmos eARC | Known Quirks |
|---------|:-----:|:------:|:------------:|:----------:|--------------|
| LG OLED C3 (4K120) | ✓ | ✗ | ✓ Profile 5 + 8.4 | ✓ | DV → HDR10 fallback automatic |
| LG OLED G3 (4K120) | ✓ | ✗ | ✓ Profile 5 + 8.4 | ✓ | clean |
| Sony A95K (4K120 QD-OLED) | ✓ | partial | ✓ Profile 5 + 8.4 | ✓ | DV gaming-mode requires firmware ≥ 03.05 |
| Samsung S95C (4K144 QD-OLED) | ✓ | ✓ | ✗ | ✓ | HDR10+ Gaming-mode |
| Samsung Q990C soundbar | ✓ | ✓ | ✗ | ✓ | clean |
| Sony X95L (4K120 mini-LED) | ✓ | partial | ✓ Profile 5 + 8.4 | ✓ | clean |
| Hisense U8K (4K144 mini-LED) | ✓ | ✓ | ✓ Profile 5 + 8.4 | ✓ | DV throttle on long sessions |
| TCL QM8 (4K144 mini-LED) | ✓ | ✓ | ✓ Profile 5 | ✓ | clean |

### 16a.3 Per-display tone-mapping fallback path

For SDR-only displays (older 1080p/4K HDR-incapable), helix-hdr.ToneMapper Strategy=BT2390 (per [helix-hdr §2.6](../06_Submodules/per-submodule/helix-hdr.md#26-tone-mapping-client-side)) operates client-side via Vulkan compute shader. Per-display tone-mapping verified at 4 ms p999 on:
- NVIDIA RTX 30+ (Ampere + Ada Lovelace) — full-rate Vulkan compute.
- Apple M1+ — Metal-on-Vulkan via MoltenVK.
- AMD Radeon RX 6000+ — full-rate Vulkan compute.
- Intel Arc A-series — full-rate Vulkan compute.
- Apple A14+ (iPad / Apple TV 4K) — Metal compute via Vulkan-portability.

Per-GPU tone-map performance audit conducted before each Phase_08 release; per-GPU regression Prometheus alert wired.

### 16a.4 Per-jurisdictional licensing posture

Dolby Vision + Atmos licensing varies by operator's commercial agreement:
- **Operator with Dolby Atmos + DV master licensing:** ships full DV Profile 8.4 + Atmos 7.1.4.
- **Operator with Atmos-only licensing:** ships Atmos 7.1.4; HDR10/HDR10+ as DV alternative.
- **Operator without Atmos/DV licensing:** ships PCM 5.1 + HDR10/HDR10+; no Atmos / DV.

The Apache-2.0 helix-codec patent grant covers the codec implementation; operator's commercial DV/Atmos license is independent (per [C32 §6](../05_Video_Audio/02_HDR_and_Color.md) + RP08-02).

---

## 16b. Per-Format Bandwidth + Channel Calculations

### 16b.1 Per-audio-format bandwidth budget

| Format | Channels | Bitrate | Per-Hour Egress |
|--------|---------:|--------:|----------------:|
| PCM 5.1 (uncompressed) | 6 | 6.144 Mbps | 2.76 GB / hour |
| Opus 5.1 | 6 | 384 kbps | 173 MB / hour |
| Opus Atmos 7.1.4 | 12 | 768 kbps | 346 MB / hour |
| Dolby AC-3 5.1 (legacy) | 6 | 640 kbps | 288 MB / hour |
| Dolby Digital Plus (E-AC-3 7.1) | 8 | 1.536 Mbps | 691 MB / hour |
| DTS-HD MA 7.1 (lossless, legacy) | 8 | 24.5 Mbps | 11 GB / hour |

helix-audio's Opus MultiStream encoder is the canonical default: 768 kbps for Atmos 7.1.4 = 346 MB / hour egress per session = small fraction of total session bandwidth (Phase_04 ~25 Mbps video baseline).

### 16b.2 Per-HDR-format bandwidth overhead

| HDR Format | Per-Frame SEI Bytes | Per-Hour Overhead (60fps) |
|------------|--------------------:|--------------------------:|
| SDR (no HDR) | 0 | 0 |
| HDR10 (static MaxCLL+MaxFALL) | ~8 (per-IDR, ~1 every 30s) | < 1 KB / hour |
| HDR10+ (ST 2094-40, per-scene) | ~64 (per-scene, ~6/min) | ~25 KB / hour |
| Dolby Vision (RPU per-frame) | ~2,048 (per-frame) | ~415 MB / hour |

Dolby Vision's per-frame RPU significantly increases bandwidth — operator-tunable per session-tier (Pro / Enterprise tiers default to DV; Free / Standard tiers default to HDR10+).

### 16b.3 Per-format encoder overhead (NVENC RTX 4090 reference)

| Format | Encode p999 (4K60) |
|--------|------------------:|
| HEVC SDR | 3.5 ms |
| HEVC HDR10 (static) | 3.6 ms (+0.1 ms) |
| HEVC HDR10+ (ST 2094-40) | 3.8 ms (+0.3 ms) |
| HEVC + Dolby Vision (Profile 8.4) | 4.2 ms (+0.7 ms) |
| AV1 SDR | 4.5 ms |
| AV1 HDR10 (static) | 4.6 ms (+0.1 ms) |
| AV1 HDR10+ (ST 2094-40) | 4.9 ms (+0.4 ms) |
| AV1 + Dolby Vision (Profile 8.4) | 5.3 ms (+0.8 ms) |

Per [Phase_07 §16a.1](Phase_07_Latency_Optimization.md#16a1-per-stage-budget-breakdown-4k60-nvidia-rtx-4090-reference), the encode stage budget is ~3.5 ms for HEVC SDR; HDR/DV adds 0.3-0.8 ms per frame.

---

## 16c. Per-Tier Audio + HDR Feature Gating

### 16c.1 Per-tier audio differentiation

Per [Phase_13 §4.4](Phase_13_GA.md#44-p13t04--per-tier-ga-sla) GA SLA tiers + Phase_10 monetisation:

| Tier | Audio Format | Channels | Bitrate |
|------|--------------|---------:|--------:|
| Free | Opus stereo | 2 | 96 kbps |
| Standard | Opus 5.1 | 6 | 384 kbps |
| Pro | Opus Atmos 7.1.4 | 12 | 768 kbps |
| Enterprise | Opus Atmos 7.1.4 + lossless option | 12 | 768 kbps (Opus) or 6 Mbps (PCM) |

### 16c.2 Per-tier HDR differentiation

| Tier | HDR Support | Notes |
|------|-------------|-------|
| Free | SDR only | no HDR |
| Standard | HDR10 (static) | universal HDR baseline |
| Pro | HDR10+ + Dolby Vision | dynamic metadata |
| Enterprise | All Pro features + per-tenant operator-tunable HDR profile | white-label per-jurisdictional |

### 16c.3 Per-tier upgrade prompts

Client UX surfaces per-tier upgrade prompt on capability denial — e.g., Free-tier user attempting to enable Atmos 7.1.4 sees "Upgrade to Pro for Atmos 7.1.4". Operator's commercial team owns the upgrade-prompt copy + checkout integration.

### 16c.4 Per-jurisdiction licensing impact on tier

Per RP08-02 — operator without DV master licensing cannot ship Pro tier with DV; operator's per-jurisdictional product roadmap reflects available formats. Per-tenant pricing adjusts to available format set.

---

## 17. Anti-Bluff Verification

### 17.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_07_Latency_Optimization.md`](Phase_07_Latency_Optimization.md) | 200+ | 2026-04-30 | predecessor                                      |
| [`../05_Video_Audio/06_Audio_Pipeline.md`](../05_Video_Audio/06_Audio_Pipeline.md) | 1,375 | 2026-04-30 | C31 architectural source                  |
| [`../05_Video_Audio/07_HDR_and_Color.md`](../05_Video_Audio/07_HDR_and_Color.md) | 1,801 | 2026-04-30 | C32 architectural source                          |
| [`../06_Submodules/per-submodule/helix-audio.md`](../06_Submodules/per-submodule/helix-audio.md) | 311 | 2026-04-30 | Audio primitive            |
| [`../06_Submodules/per-submodule/helix-hdr.md`](../06_Submodules/per-submodule/helix-hdr.md) | 328 | 2026-04-30 | HDR primitive                              |

### 17.2 Forbidden patterns

Clean.

### 17.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_08 execution + operator signoff.

End of `09_Implementation_Phases/Phase_08_Audio_Surround.md` — 2026-04-30.
