# Phase_09 — Recording & Replay

> **Source dimensions:** [`Phase_08_Audio_Surround.md`](Phase_08_Audio_Surround.md), [`../06_Submodules/per-submodule/helix-record.md`](../06_Submodules/per-submodule/helix-record.md), [`../06_Submodules/per-submodule/helix-vqa.md`](../06_Submodules/per-submodule/helix-vqa.md), [`../05_Video_Audio/04_DualPath_Encoding.md`](../05_Video_Audio/04_DualPath_Encoding.md) (C29), [`../05_Video_Audio/05_Recording_Storage.md`](../05_Video_Audio/05_Recording_Storage.md) (C30), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P09.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P09.
> **Phase targets:** R-13 + GDPR (recording-side data classification per Constitution §11).
> **Cross-links:** [`Phase_10_Monetization_and_Auth.md`](Phase_10_Monetization_and_Auth.md), [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_09 ships the **recording + replay** stack: helix-dualpath splits the encoded NAL feed into a stream rung (low-latency) + a record rung (recording-quality); helix-record muxes the record rung into fMP4 + MKV containers; the recorded sessions are stored in MinIO (S3-compat) + replayable via DASH. helix-vqa attaches VMAF + ViSQOL measurement to the recording for change-point detection.

After Phase_09, every gameplay session is **post-game replayable** — operator's revenue-significant feature. A user can scrub through a 60-minute match recording on the web client + on the Wails desktop client.

---

## 2. Prerequisites

- Phase_04 (streaming pipeline) + Phase_05 (clients) + Phase_08 (audio surround + HDR) complete.
- helix-record + helix-vqa + helix-dualpath at v1.0.0.
- MinIO S3-compatible storage operational (Phase_03).

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P09.T01   | helix-dualpath stream + record rung activation                | 4        |
| P09.T02   | helix-record fMP4 + MKV mux pipeline                          | 5        |
| P09.T03   | helix-record S3 sync to MinIO with retry logic                | 4        |
| P09.T04   | At-rest encryption via Vault-managed AES-GCM keys            | 4        |
| P09.T05   | DASH manifest generation + segment-by-segment publishing      | 4        |
| P09.T06   | Web replay client (HLS / DASH player)                          | 5        |
| P09.T07   | helix-vqa VMAF + ViSQOL on archived recordings                | 4        |
| P09.T08   | Replay-ahead + replay-behind scrubbing                        | 4        |
| P09.T09   | GDPR retention policy + per-tenant deletion workflow          | 5        |
| P09.T10   | Operator runbook for recording storage management              | 3        |
| P09.T11   | End-to-end smoke — record + replay 5-minute session            | 4        |
| P09.T12   | Phase_09 acceptance review                                     | 2        |

12 tasks; ~48 subtasks.

---

## 4. Task Details

### 4.1 P09.T01 — Dual-rung activation

helix-dualpath.Splitter wired between encoder and the two sinks (transport + record). Per [helix-dualpath descriptor §2](../06_Submodules/per-submodule/helix-dualpath.md). Verify both rungs receive their NAL slices.

### 4.2 P09.T02 — Mux pipeline

helix-record.Recorder configured for fMP4 (DASH-friendly) + MKV (archival). Segment duration per [O01 §13](../08_Operations/01_Container_CI_CD.md): 4-second segments + 30-second IDR.

### 4.3 P09.T03 — S3 sync

Per [helix-record descriptor §2.3](../06_Submodules/per-submodule/helix-record.md). Async upload via goroutine pool; local-buffer queue for partition tolerance.

### 4.4 P09.T04 — At-rest encryption

Vault-managed AES-GCM key per recording. Operator's KEK rotation (annual per [helix-vault](../06_Submodules/per-submodule/helix-vault.md)) re-wraps the recording's DEK on rotation.

### 4.5 P09.T05 — DASH manifest

DASH-CMAF manifest generation per [helix-record §10](../06_Submodules/per-submodule/helix-record.md). Per-segment cosign signatures for tamper-evident replay.

### 4.6 P09.T06 — Web replay client

A Web Component player based on dash.js v4 + cosign verification. Hosted at `replay.helix.example.com/<session-id>`. Per-tenant access control via OAuth.

### 4.7 P09.T07 — VMAF + ViSQOL on archive

helix-vqa runs against archived recordings on a nightly batch. Stores per-recording VMAF score + ViSQOL audio score in CockroachDB.

### 4.8 P09.T08 — Replay scrubbing

DASH random-access seeking with helix-record.Replay backend; ≤ 200 ms seek latency.

### 4.9 P09.T09 — GDPR retention

Per Constitution §11.5 + helix-vault EraseTenant. Per-tenant retention policy (default: 90 days; operator-tunable). On tenant erasure, all the tenant's recordings' DEKs are shredded (renders ciphertext unrecoverable).

### 4.10 P09.T10 — Storage management runbook

Operator's runbook for monitoring storage growth + cold-tier eviction + compliance audit.

### 4.11 P09.T11 — End-to-end smoke

Record a 5-minute session → upload to MinIO → trigger DASH manifest publication → open web replay → scrub through.

### 4.12 P09.T12 — Acceptance

Operator signoff.

---

## 5. Subtask Catalogue

The 48 subtasks (per the §3 task table) enumerated below — every subtask is a discrete `[P09.Tyy.Szz]` ticket per [O05](../08_Operations/05_Tracking_GitHub_GitLab.md) opened in the operator's GitHub Projects + GitLab boards.

### 5.1 Per-task subtask listings

**P09.T01 — Dual-rung activation (4 subtasks)**

- T01.S01 — helix-dualpath.Splitter wired between encoder and the two sinks (transport + record).
- T01.S02 — Stream rung target verified — helix-transport accepts NAL slices with stream-rung profile.
- T01.S03 — Record rung target verified — helix-record accepts NAL slices with record-rung profile (higher fidelity).
- T01.S04 — Per-rung observability — Prometheus per-rung NAL count + bytes counter exposed.

**P09.T02 — Mux pipeline (5 subtasks)**

- T02.S01 — fMP4 mux configured with 4-second segment duration + 30-second IDR cadence.
- T02.S02 — MKV mux configured for archival format (lossless container).
- T02.S03 — Per-mux backend NAL→muxer goroutine pool; backpressure surfaced via channel depth metric.
- T02.S04 — Mux-failure handling — partial segment recovered + cosign-signed at next-segment boundary.
- T02.S05 — Mux Challenges scenario per [helix-record §6](../06_Submodules/per-submodule/helix-record.md).

**P09.T03 — S3 sync (4 subtasks)**

- T03.S01 — Async S3 upload via goroutine pool sized to per-tenant bandwidth budget.
- T03.S02 — Local-buffer queue (disk-backed) for partition tolerance — survives MinIO outage up to 4 hours.
- T03.S03 — Retry + exponential backoff per [helix-record §2.3](../06_Submodules/per-submodule/helix-record.md).
- T03.S04 — Per-tenant S3 credentials rotation hooks (per Phase_11 secret rotation).

**P09.T04 — At-rest encryption (4 subtasks)**

- T04.S01 — Vault-managed AES-GCM key per recording.
- T04.S02 — Per-tenant DEK isolation (no cross-tenant key reuse).
- T04.S03 — Lazy DEK re-wrap on annual KEK rotation per [helix-vault §2.4](../06_Submodules/per-submodule/helix-vault.md).
- T04.S04 — Encrypted-at-rest verification probe — write recording, read back, verify ciphertext + decrypt.

**P09.T05 — DASH manifest (4 subtasks)**

- T05.S01 — DASH-CMAF manifest generation per segment-by-segment publication cadence.
- T05.S02 — Per-segment cosign signature (tamper-evident replay).
- T05.S03 — Manifest publication to MinIO — `replay.helix.<operator>/<session-id>/manifest.mpd`.
- T05.S04 — Manifest TTL + per-tenant retention enforcement.

**P09.T06 — Web replay client (5 subtasks)**

- T06.S01 — Web Component player based on dash.js v4.
- T06.S02 — Cosign verification of every segment before render.
- T06.S03 — Per-tenant access control via OAuth (per Phase_10 P10.T01).
- T06.S04 — Hosted at `replay.helix.<operator-domain>/<session-id>` + CSP-hardened.
- T06.S05 — Offline-mode fallback (operator-LAN-isolated tenant).

**P09.T07 — VMAF + ViSQOL on archive (4 subtasks)**

- T07.S01 — Nightly batch job — helix-vqa runs against all completed recordings of the prior 24 hours.
- T07.S02 — Per-recording VMAF score stored in CockroachDB `recording_quality.vmaf` column.
- T07.S03 — Per-recording ViSQOL audio score stored in `recording_quality.visqol` column.
- T07.S04 — Quality-trend Grafana dashboard — operator-visible per-tenant rolling p10 / p50 / p99.

**P09.T08 — Replay scrubbing (4 subtasks)**

- T08.S01 — DASH random-access seeking with helix-record.Replay backend.
- T08.S02 — ≤ 200 ms seek latency p99 — verifiable via helix-bench scenario.
- T08.S03 — Frame-accurate seek — IDR-anchored.
- T08.S04 — Audio + video sync verified at every seek (≤ 40 ms drift per Phase_08).

**P09.T09 — GDPR retention (5 subtasks)**

- T09.S01 — Per-tenant retention policy default — 90 days; operator-tunable per tenant.
- T09.S02 — Per-tenant erasure trigger via helix-vault.EraseTenant.
- T09.S03 — DEK shred renders ciphertext unrecoverable.
- T09.S04 — Erasure certificate cosign-signed + delivered to tenant.
- T09.S05 — SLA monitoring — Prometheus alert if erasure exceeds 25 days.

**P09.T10 — Storage management runbook (3 subtasks)**

- T10.S01 — Storage growth monitoring — operator-visible per-region MinIO capacity.
- T10.S02 — Cold-tier eviction — per-recording lifecycle policy (≥ 30 days → cold tier).
- T10.S03 — Compliance audit workflow — per-tenant + per-region GDPR + tax-jurisdiction reports.

**P09.T11 — End-to-end smoke (4 subtasks)**

- T11.S01 — Record a 5-minute synthetic gameplay session.
- T11.S02 — Verify upload to MinIO + DASH manifest publication.
- T11.S03 — Open web replay + scrub through.
- T11.S04 — Verify VMAF + ViSQOL scores recorded.

**P09.T12 — Acceptance (2 subtasks)**

- T12.S01 — Operator + compliance officer signoff.
- T12.S02 — Phase_09 closure ticket on operator's project board.

---

## 6. Exit Criteria

- [ ] Dual-rung split operational on every session.
- [ ] fMP4 + MKV recording produced + uploaded to MinIO.
- [ ] At-rest encryption verified via key-rotation drill.
- [ ] DASH manifest published + verifiable.
- [ ] Web replay client streams the recording.
- [ ] VMAF + ViSQOL scoring on nightly batch.
- [ ] Replay scrubbing ≤ 200 ms seek.
- [ ] GDPR per-tenant deletion verified.
- [ ] End-to-end smoke passes.
- [ ] Operator signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP09-01  | MinIO storage cost runaway                                            | Per-tenant retention policy + cold-tier eviction.                     |
| RP09-02  | DASH segment cosign signature mismatch                                | Per-segment signing + verify; mismatch rejects replay.                |
| RP09-03  | At-rest encryption key-rotation breaks replay of older recordings     | Lazy DEK re-wrap per [helix-vault](../06_Submodules/per-submodule/helix-vault.md) §2.4 — recordings remain decryptable. |
| RP09-04  | GDPR erasure-on-demand SLA exceeded                                   | Erasure procedure + SLA documented; auto-triggered DEK shred. |

---

## 8. Cross-Family Dependencies

- C29 §6 (Dual-Path Encoding) + C30 §6 (Recording Storage) architectural sources.
- helix-dualpath, helix-record, helix-vqa at v1.0.0.
- helix-vault for at-rest encryption.

---

## 9. Acceptance Criteria

Constitution §16 + §6 exit criteria.

---

## 10. The Phase_09 Calendar

~ 6 weeks. Operator-side capacity: 4 engineers (1× backend Go for helix-record + helix-dualpath integration; 1× backend Go for helix-vqa nightly batch; 1× frontend / web for the replay client + DASH player; 1× ops engineer for MinIO + storage runbook).

The Phase_09 calendar is **gated by Phase_08** (audio + HDR fidelity must be complete before recordings have the full fidelity stack to capture) and unblocks Phase_12 beta launch (recordings are operator's revenue-significant feature; no recording = no beta-customer onboarding).

---

## 11. Per-Phase Observability Catalogue

The Phase_09 deployment exposes the following Prometheus metrics + Grafana dashboards + Loki log streams + Tempo traces, consumed by [O04 §6](../08_Operations/04_Observability_and_Events.md):

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_dualpath_nal_split_total` | counter | tenant, rung | rate ≥ 60/s per active session |
| `helix_record_segments_written_total` | counter | tenant, format | one per 4 s + IDR boundary |
| `helix_record_s3_upload_seconds` | histogram | tenant, region | p99 ≤ 2 s per 4-s segment |
| `helix_record_local_buffer_bytes` | gauge | tenant | < 10 % capacity steady-state |
| `helix_dash_manifest_segments_total` | counter | tenant, session | matches segments-written |
| `helix_dash_segment_signing_seconds` | histogram | tenant | p99 ≤ 50 ms |
| `helix_replay_seek_latency_seconds` | histogram | tenant | p99 ≤ 200 ms |
| `helix_vqa_vmaf_score` | gauge | tenant, session | per-tenant rolling p10 ≥ 90 |
| `helix_vqa_visqol_score` | gauge | tenant, session | per-tenant rolling p10 ≥ 4.0 |
| `helix_vault_dek_rewrap_seconds` | histogram | tenant | annual KEK rotation envelope |
| `helix_gdpr_erasure_pending_days` | gauge | tenant | < 25 (alert at 25; SLA 30) |

### 11.2 Grafana dashboards

- **Per-tenant Recording Health** — segments-written rate / S3 upload latency / local buffer depth / VMAF + ViSQOL trend.
- **Per-region Storage Capacity** — MinIO capacity / cold-tier migration rate / per-tenant retention compliance.
- **Per-tenant GDPR Compliance** — pending-erasure count + days-elapsed / SLA-breach alerts.
- **Per-session Replay Quality** — seek latency / segment cosign-verify failure rate / per-segment availability.

### 11.3 Loki log streams

- `helix-record.recorder` — per-segment write events.
- `helix-record.uploader` — S3 upload events + retry events.
- `helix-vqa.scorer` — per-recording quality-score events.
- `helix-vault.dek-rewrap` — KEK rotation events.

### 11.4 Tempo traces

- `record.session.lifecycle` — per-session span tree from session-start through final-segment + DASH manifest publication.
- `replay.session.scrub` — per-scrub-event span tree (request → manifest lookup → segment fetch → cosign verify → render).

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Recording availability | Sessions with ≥ 99% segments uploaded to MinIO | ≥ 99.9% | 30-day rolling |
| Recording latency | Time from session-end to DASH manifest publication | p99 ≤ 30 s | 30-day rolling |
| Replay availability | Recordings retrievable via web replay | ≥ 99.95% | 30-day rolling |
| Replay seek latency | Time from scrub-event to first frame rendered | p99 ≤ 200 ms | 30-day rolling |
| VMAF quality floor | Per-tenant rolling p10 VMAF score | ≥ 90 | 7-day rolling |
| ViSQOL quality floor | Per-tenant rolling p10 ViSQOL score | ≥ 4.0 | 7-day rolling |
| GDPR erasure SLA | Time from erasure request to certificate delivery | p100 ≤ 30 days | per-request |
| At-rest encryption coverage | Recordings with verified Vault-managed DEK | 100% | per-recording |

SLO breaches trigger Prometheus alerts routed via [O04 §4](../08_Operations/04_Observability_and_Events.md) per-tier alert routing (CRITICAL → PagerDuty + on-call escalation; HIGH → operator-billing-team Slack channel).

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixRecord/docs/runbook/phase09-operations.md` — operator's runbook covering Phase_09 deployment, ongoing operations, incident response. Sections:

- **§1 Deployment** — helix-record + helix-dualpath + helix-vqa container deployment per [O01](../08_Operations/01_Container_CI_CD.md).
- **§2 MinIO bootstrapping** — per-region bucket + per-tenant prefix + lifecycle policy + encryption-at-rest configuration.
- **§3 At-rest encryption** — Vault namespace setup + KEK rotation cadence + DEK lifecycle.
- **§4 Per-tenant retention** — operator-tunable retention policy + cold-tier migration + audit reports.
- **§5 GDPR erasure** — request-to-certificate workflow + DEK shred procedure + cosign-signed certificate generation.
- **§6 Incident response — recording loss** — segment-loss detection + reconstruction-from-stream-rung fallback + operator escalation.
- **§7 Incident response — replay outage** — DASH manifest validation + segment availability check + cosign chain integrity verification.
- **§8 Storage capacity management** — per-region capacity monitoring + cold-tier eviction tuning + capacity expansion procedure.
- **§9 Cosign key rotation** — per-segment signing key lifecycle + Sigstore-Fulcio short-lived cert workflow.
- **§10 Quality regression response** — VMAF / ViSQOL score regression detection + operator escalation + per-tenant communication.

---

## 14. Implementation Considerations

### 14.1 Segment-duration vs replay-seek-latency trade-off

The 4-second segment + 30-second IDR cadence (per [O01 §10](../08_Operations/01_Container_CI_CD.md)) is operator-tunable but the trade-off space is constrained:

- **Shorter segments** (e.g., 2s) → finer scrubbing granularity but ~2× the manifest size + ~2× the cosign-sign overhead.
- **Longer segments** (e.g., 8s) → coarser scrubbing + bigger initial-buffer at replay-start, but ~50% manifest + sign overhead.

The 4-second canonical setting was derived from [helix-record §10](../06_Submodules/per-submodule/helix-record.md) calibration runs; operator's commercial team can override per-tier (e.g., Enterprise tier gets 2s for finer scrubbing).

### 14.2 Per-tenant retention vs storage cost

The default 90-day retention drives MinIO storage budgeting. Per-tenant 4K60 session at ~25 Mbps → ~280 MB / minute → ~17 GB / hour → ~4 TB / 240-hour-month. At 90-day retention, 30 GB / hour-of-operator-use accumulates. Operator's commercial-tier pricing must amortize storage against per-tier rate.

### 14.3 At-rest encryption performance overhead

AES-GCM 256-bit at the MinIO ingestion path adds ~3% CPU overhead per [helix-record §9.3](../06_Submodules/per-submodule/helix-record.md). The CPU cost is negligible relative to the encode + segment + sign chain; the genuine cost is the per-recording DEK lifecycle (Vault round-trip + per-recording cryptographic operation).

### 14.4 Dual-rung activation correctness

helix-dualpath's NAL split must be **exact** — every NAL slice must land in both rungs (no drops). Per [helix-dualpath §2.2](../06_Submodules/per-submodule/helix-dualpath.md):

- The Splitter uses a single read-cursor across the encoder output → emits to both sinks.
- A drop detected in either rung surfaces a `helix_dualpath_split_drop_total` Prometheus counter increment + operator alert.
- The Phase_09 P09.T01.S04 observability subtask explicitly wires this counter.

### 14.5 Cosign-segment-signing correctness

Per-segment cosign signature verifies tamper-evident replay (per Phase_11 P11.T15). The sign overhead is ~50 ms p99 per segment (per §11.1 SLO). At 4-second segments, that's 1.25% of segment-window budget — acceptable. The verify-side at the web replay player is symmetric.

---

## 15. Phase_09 Cost Estimation

Per-tenant per-month cost breakdown for Phase_09 (operator-side; per region):

| Component | Unit Cost | Per-Tenant Per-Month |
|-----------|-----------|----------------------|
| MinIO storage (4 TB / 240 h × 90 d retention) | ~$23 / TB-month | ~$92 / tenant @ 240 h-month |
| MinIO egress (replay traffic, ~100 GB / month) | ~$80 / TB | ~$8 / tenant |
| Vault DEK operations (per-recording) | ~$0 (operator-self-hosted) | $0 |
| helix-record CPU (~3% encoder overhead) | folded into Phase_04 | $0 |
| helix-vqa nightly batch (per-recording, ~30 s GPU-second) | ~$0.001 / GPU-second | ~$0.50 / tenant @ 240 h-month |
| Cosign signing (per-segment) | folded into helix-record CPU | $0 |
| **Total per-tenant per-month** | — | **~$100** |

Operator's per-tenant rate must amortize this baseline + Phase_03 (backend services) + Phase_04 (streaming) costs.

---

## 16. Cross-Mirror Parity Verification

Per-phase commit cadence pushes to the four-mirror topology (github + gitlab + gitflic + gitverse) via the composite-push origin. Phase_09 closure verification:

- `git ls-remote github main` + `git ls-remote gitlab main` + `git ls-remote gitflic main` + `git ls-remote gitverse main` → all four SHAs equal.
- Phase_09 closure tag `phase-09-recording-replay-complete` cosign-signed + pushed to all four mirrors.
- Per-phase board mirror — operator's GitHub Projects + GitLab boards reflect every `[P09.Tyy.Szz]` ticket.

---

## 17. Anti-Bluff Verification

### 17.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_08_Audio_Surround.md`](Phase_08_Audio_Surround.md)        | 200+ | 2026-04-30 | predecessor                                      |
| [`../05_Video_Audio/05_Recording_Storage.md`](../05_Video_Audio/05_Recording_Storage.md) | 2,126 | 2026-04-30 | C30 architectural source                |
| [`../06_Submodules/per-submodule/helix-record.md`](../06_Submodules/per-submodule/helix-record.md) | 312 | 2026-04-30 | record primitive            |

### 17.2 Forbidden patterns

Clean.

### 17.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_09 execution + operator signoff.

End of `09_Implementation_Phases/Phase_09_Recording_and_Replay.md` — 2026-04-30.
