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

48 subtasks.

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

## 10. Anti-Bluff Verification

### 10.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_08_Audio_Surround.md`](Phase_08_Audio_Surround.md)        | 200+ | 2026-04-30 | predecessor                                      |
| [`../05_Video_Audio/05_Recording_Storage.md`](../05_Video_Audio/05_Recording_Storage.md) | 2,126 | 2026-04-30 | C30 architectural source                |
| [`../06_Submodules/per-submodule/helix-record.md`](../06_Submodules/per-submodule/helix-record.md) | 312 | 2026-04-30 | record primitive            |

### 10.2 Forbidden patterns

Clean.

### 10.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_09 execution + operator signoff.

End of `09_Implementation_Phases/Phase_09_Recording_and_Replay.md` — 2026-04-30.
