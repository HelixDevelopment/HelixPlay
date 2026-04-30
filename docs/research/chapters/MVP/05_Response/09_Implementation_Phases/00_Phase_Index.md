# 09_Implementation_Phases/ — Phase Index

> **Status:** Draft v1.
> **Last updated:** 2026-04-30.
> **Purpose:** Navigation hub for the *Implementation Phases* family — the seventh family in `05_Response/`, the executable plan that turns the synthesis programme into actual code shipped under the 29 submodules + 4 organisational repos.
> **Targets (R-XX):** R-16 (fine-grained phases / tasks / subtasks), R-17 (mirrored to GitHub Projects + GitLab via `gh` + `glab`).
> **Cross-links:** [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row block PNN; [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §9 (release-train cadence); [`../07_Testing/`](../07_Testing/) (the test discipline each phase must satisfy); [`../08_Operations/`](../08_Operations/) (the operational machinery each phase deploys).

---

## 1. Position in the Synthesis Programme

The Implementation Phases family is the **seventh of nine** in `05_Response/` and the **fourth aggregation** family. Where the prior aggregation families (Submodules + Testing + Operations) ratify the *standing rules*, this family ratifies the **execution plan**: which submodules graduate first, which infrastructure stands up first, which milestones unlock subsequent work.

| # | Family               | Role                  | Status                                |
|---|----------------------|----------------------|---------------------------------------|
| 1 | `03_Architecture/`   | content              | **closed**                            |
| 2 | `04_Latency/`        | content              | **closed**                            |
| 3 | `05_Video_Audio/`    | content              | **closed**                            |
| 4 | `06_Submodules/`     | aggregation          | **closed**                            |
| 5 | `07_Testing/`        | aggregation          | **closed**                            |
| 6 | `08_Operations/`     | aggregation          | **closed**                            |
| 7 | `09_Implementation_Phases/` | **execution plan**| this family — **in progress**         |
| 8 | `99_Web_Research_Addenda/` | living research | append-only                            |
| 9 | `00–02` foundation   | governance           | **closed**                            |

The family does **not** introduce new submodules / test types / operational layers — those are frozen by the prior families. What it adds:

1. **A 14-phase execution plan** (Phase_00 Foundation through Phase_13 GA) sequencing the actual code-and-deploy work.
2. **Per-phase task / subtask breakdowns** — fine-grained enough to be `[Pxx.Tyy.Szz]` ticket-mirrored to GitHub Projects + GitLab via [O05](../08_Operations/05_Tracking_GitHub_GitLab.md).
3. **Phase-exit criteria** — measurable conditions that must be met before the next phase can begin.
4. **Risk register entries** specific to each phase.

---

## 2. Family Chapter List

Per [Master Plan §3](../00_Master_Plan.md#3-output-structure) + §7.2:

| ID  | Chapter                                          | Floor (lines) | Status     |
|-----|--------------------------------------------------|--------------:|------------|
| —   | [`00_Phase_Index.md`](00_Phase_Index.md) — this file | navigation     | **draft**  |
| P00 | [`Phase_00_Foundation.md`](Phase_00_Foundation.md)   | 800           | pending    |
| P01 | [`Phase_01_Containers_and_CI.md`](Phase_01_Containers_and_CI.md) | 500       | pending    |
| P02 | [`Phase_02_Core_Submodules.md`](Phase_02_Core_Submodules.md)   | 500       | pending    |
| P03 | [`Phase_03_Backend_Services.md`](Phase_03_Backend_Services.md) | 500       | pending    |
| P04 | [`Phase_04_Streaming_Pipeline.md`](Phase_04_Streaming_Pipeline.md) | 500    | pending    |
| P05 | [`Phase_05_Clients.md`](Phase_05_Clients.md)              | 500       | pending    |
| P06 | [`Phase_06_Host_Agent.md`](Phase_06_Host_Agent.md)        | 500       | pending    |
| P07 | [`Phase_07_Latency_Optimization.md`](Phase_07_Latency_Optimization.md) | 500 | pending    |
| P08 | [`Phase_08_Audio_Surround.md`](Phase_08_Audio_Surround.md) | 500       | pending    |
| P09 | [`Phase_09_Recording_and_Replay.md`](Phase_09_Recording_and_Replay.md) | 500 | pending    |
| P10 | [`Phase_10_Monetization_and_Auth.md`](Phase_10_Monetization_and_Auth.md) | 500 | pending |
| P11 | [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md) | 500 | pending |
| P12 | [`Phase_12_Beta_Launch.md`](Phase_12_Beta_Launch.md)      | 500       | pending    |
| P13 | [`Phase_13_GA.md`](Phase_13_GA.md)                        | 500       | pending    |

---

## 3. The 14-Phase Execution Sequence at a Glance

Per Master Plan §3 output structure + the dependency chain explicit in the per-submodule [S01 §6](../06_Submodules/01_Submodule_Catalog.md#6-dependency-graph--single-point-of-failure-analysis) topology:

| Phase | Theme                                         | Primary deliverable                                        | Predecessor |
|-------|-----------------------------------------------|------------------------------------------------------------|-------------|
| P00   | Foundation                                    | Operator infra + tooling + project boards + Vault         | (none)      |
| P01   | Containers & CI                                | `vasic-digital/Containers` v1.0.0 + four-mirror runners   | P00         |
| P02   | Core Submodules                                | 29 submodules at v0.x.y → graduating                      | P01         |
| P03   | Backend Services                                | CockroachDB + NATS + Redis + Vault deployed              | P02 partial |
| P04   | Streaming Pipeline                              | helix-pipeline + helix-transport up                       | P02 + P03   |
| P05   | Clients                                         | Wails desktop + Compose-for-TV + Steam Deck client       | P04         |
| P06   | Host Agent                                      | Sunshine++ host agent up                                  | P04         |
| P07   | Latency Optimization                            | helix-rtos + GOCACHEPROG + tuning                         | P04 + P06   |
| P08   | Audio Surround                                  | Atmos + 7.1.4 + eARC                                      | P05 + P06   |
| P09   | Recording + Replay                              | helix-record + DASH replay client                         | P04 + P05   |
| P10   | Monetization + Auth                              | OAuth + tenant + billing                                   | P03         |
| P11   | Hardening + Security                             | R-18 enforcement + audit log + KEK rotation              | P02 + P10   |
| P12   | Beta Launch                                     | Operator-facing canary + 30-day exhaustive replay        | P01–P11    |
| P13   | GA                                              | v1.0.0 release-train tag-publish across the fleet         | P12         |

The dependency graph is documented per-phase in each chapter's §1 *Prerequisites*.

---

## 4. The Implementation-Phase ↔ Submodule Catalog Mapping

Each phase delivers + tests a subset of the 29 submodules. The mapping (high level):

| Phase | Submodules delivered (v1.0.0 graduation gate at end-of-phase) |
|-------|---------------------------------------------------------------|
| P02   | helix-r18-safeexec (the SPOF root, first to graduate)         |
| P02   | helix-shm + helix-iouring + helix-xdp + helix-lockfree + helix-mempool + helix-allocator + helix-bench (depth-1/2 Latency primitives) |
| P02   | helix-codec + helix-encoder + helix-capture (depth-1/2 Video primitives) |
| P03   | helix-vault + helix-tenant + helix-grpc-frame                 |
| P04   | helix-network + helix-abr + helix-dualpath + helix-pipeline + helix-transport |
| P05   | helix-tv-input + helix-input + helix-display                  |
| P06   | helix-rtos + helix-gpu-direct                                 |
| P08   | helix-audio + helix-hdr                                        |
| P09   | helix-record + helix-vqa                                      |
| P11   | helix-thermal (final hardening submodule)                     |

The 29 submodules are delivered across phases P02..P11; Phases P00 + P01 + P12 + P13 are infrastructure-only.

---

## 5. The Phase-Exit Criteria Pattern

Every phase chapter's §6 *Exit Criteria* enumerates measurable conditions:

- All deliverables landed + signed.
- Per-submodule `v1.0.0` graduation gates met (per [S01 §9.2](../06_Submodules/01_Submodule_Catalog.md#92-the-v100-graduation-criteria)) for in-phase submodules.
- All Challenges scenarios green for the in-phase submodules.
- All Operations-family runbooks for the in-phase deployments green.
- Operator sign-off on the phase's Definition of Done.

A phase that doesn't meet exit criteria does not unblock the next phase. The Implementation Phases family is the only family with **strict serial dependency** (vs the parallel-friendly content + aggregation families).

---

## 6. The Implementation-Phases Operating Discipline

Each phase chapter is the **specification**; the actual execution happens under per-submodule + per-organisational-repo PRs. The discipline:

1. **Open the phase ticket** on GitHub Projects + GitLab board (per [O05 §4 + §5](../08_Operations/05_Tracking_GitHub_GitLab.md#4-the-gh-invocation-pattern)).
2. **Open per-task tickets** with the `[Pxx.Tyy]` prefix.
3. **Open per-subtask tickets** with the `[Pxx.Tyy.Szz]` prefix.
4. **Per-PR work** lands the subtasks; per-task PRs aggregate; per-phase PRs close once all tasks complete.
5. **Phase Definition of Done** is signed off by the operator on the phase ticket; closing the ticket marks the phase complete.

The discipline is mandated by Constitution §16 *Acceptance* — non-delegable operator review of phase-exit conditions.

---

## 7. The Anti-Bluff Commitment

Every phase chapter ends with §10 *Anti-Bluff Verification* per [Master Plan §4.3](../00_Master_Plan.md#43-anti-bluff-verification-block-r-13). The chapter explicitly:

- Lists every source it consumes.
- Confirms zero forbidden-pattern matches.
- Confirms cross-link integrity.
- Records the operator review status.

R-02 + R-13 are non-negotiable; a phase chapter with an unfilled task or a TODO marker is rejected.

---

## 8. Anti-Bluff Verification (family-level)

### Sources resolved

| Path                                                              | Lines  | Role                                            |
|-------------------------------------------------------------------|-------:|-------------------------------------------------|
| [`../00_Master_Plan.md`](../00_Master_Plan.md) §3, §7.2            |    2k+ | family chapter list (P00..Phase_13_GA)         |
| [`../06_Submodules/`](../06_Submodules/) S01..S05                  | 12,419 | submodule catalog + dependency graph            |
| [`../07_Testing/`](../07_Testing/) T01..T12                        |  4,533 | test discipline                                  |
| [`../08_Operations/`](../08_Operations/) O01..O06                  |  2,315 | operational machinery                            |

### Forbidden patterns

Index prose scanned: clean.

### Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Pending: Phase_00..Phase_13_GA chapters, each with their own Anti-Bluff block.
- Reviewed by: pending operator review.

End of `09_Implementation_Phases/00_Phase_Index.md` — 2026-04-30.
