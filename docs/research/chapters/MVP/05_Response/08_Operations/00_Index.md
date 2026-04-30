# 08_Operations/ — Family Index

> **Status:** Draft v1.
> **Last updated:** 2026-04-30.
> **Purpose:** Navigation hub for the *Operations* family — the sixth aggregation family + the operational meeting point with the deployment topology.
> **Targets (R-XX):** R-05 (Containers), R-06 (every service / build / test / scan inside containers), R-07 (gRPC + HTTP/3 + Brotli + service-discovery), R-08 (NATS / Redis / RabbitMQ + observability), R-10 (heavy quality + security scans), R-16 (fine-grained phases / tasks / subtasks), R-17 (mirror to GitHub Projects + GitLab), R-18 (Operational Integrity).
> **Cross-links:** [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row block O0X; [`../01_Constitution.md`](../01_Constitution.md) §3 (Containerised Runtime), §4 (Communication Stack), §7 (Quality Gates), §8 (Tracking), §9 (Source Control & Git Topology), §10 (Observability), §11.5 (R-18); [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) (S02 — operationalised here at the deployment-pipeline level); [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) (S04 — operationalised here at the deployment-gate level); [`../07_Testing/12_HelixQA_Autonomous.md`](../07_Testing/12_HelixQA_Autonomous.md) (T12 — Operations consumes T12's playbook for the actual cron + alert-routing wiring).

---

## 1. Position in the synthesis programme

The Operations family is the **sixth of nine** in `05_Response/` and the **third aggregation** family (after [`06_Submodules/`](../06_Submodules/) catalog and [`07_Testing/`](../07_Testing/) discipline). Its role: ratify the **deployment-time** and **production-runtime** operational contracts that the prior families specify only at the architectural / catalog / test-discipline level.

| # | Family               | Role          | Status                                            |
|---|----------------------|---------------|---------------------------------------------------|
| 1 | `03_Architecture/`   | content       | 13 chapters, **closed**                            |
| 2 | `04_Latency/`        | content       | 11 chapters, **closed**                            |
| 3 | `05_Video_Audio/`    | content       | 13 chapters, **closed**                            |
| 4 | `06_Submodules/`     | aggregation   | 4 + 29 chapters, **closed**                        |
| 5 | `07_Testing/`        | aggregation   | 13 chapters, **closed**                            |
| 6 | `08_Operations/`     | aggregation   | this family — **in progress**                     |
| 7 | `09_Implementation_Phases/` | execution plan | follows Operations — pending                  |
| 8 | `99_Web_Research_Addenda/` | living | append-only                                       |
| 9 | `00–02` foundation   | governance    | **closed**                                         |

What this family **does not** introduce:
- New submodules (the catalog at [S01 §3.1](../06_Submodules/01_Submodule_Catalog.md#31-the-29-submodule-table) is frozen).
- New test types (the Ten at [T01 §2](../07_Testing/01_Test_Matrix.md#2-the-29--10--4-cell-grid-the-canonical-lookup) are frozen).
- New Constitution clauses (R-01..R-18 are frozen).

What it **adds**:
1. **The container CI/CD pipeline** (O01) — the actual GitHub Actions / GitLab CI / GitFlic / GitVerse pipeline shape that consumes [S02 §3](../06_Submodules/02_Containers_Submodule.md#3-the-per-submodule-ci-lane-catalog) lanes + ships images per [S02 §4–§7](../06_Submodules/02_Containers_Submodule.md#4-multi-arch-image-strategy-linuxamd64--linuxarm64-default).
2. **The quality-gate operations** (O02) — SonarQube + Snyk integration the deployment gate consumes per [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad).
3. **Service discovery + dynamic-port assignment** (O03) — the LAN service-discovery mechanism per [Constitution §4](../01_Constitution.md#4-communication-stack-r-07-r-08) (R-07).
4. **Observability + event-bus operations** (O04) — NATS + Redis + Prometheus + Grafana + OTel topology per [Constitution §10](../01_Constitution.md#10-observability-r-08).
5. **Tracking integration** (O05) — `gh project` + `glab project` invocations operationalising R-17 (the W07 task in Master Plan §7.2).
6. **Git topology + push policy** (O06) — operationalises [Constitution §9](../01_Constitution.md#9-source-control--git-topology) (the four-mirror composite-push pattern).

---

## 2. Family chapter list

Per [Master Plan §3](../00_Master_Plan.md#3-output-structure):

| ID  | Chapter                                                                                  | Floor (lines) | Status     |
|-----|------------------------------------------------------------------------------------------|--------------:|------------|
| —   | [`00_Index.md`](00_Index.md) — this file                                                  | navigation    | **draft**  |
| O01 | [`01_Container_CI_CD.md`](01_Container_CI_CD.md) — container CI/CD pipeline              | 600           | pending    |
| O02 | [`02_Quality_Gates_SonarQube_Snyk.md`](02_Quality_Gates_SonarQube_Snyk.md) — quality gates | 400           | pending    |
| O03 | [`03_Service_Discovery_and_Ports.md`](03_Service_Discovery_and_Ports.md) — service discovery | 300         | pending    |
| O04 | [`04_Observability_and_Events.md`](04_Observability_and_Events.md) — observability + events | 300          | pending    |
| O05 | [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md) — R-17 ticket mirror       | 300           | pending    |
| O06 | [`06_Git_Topology_and_Push_Policy.md`](06_Git_Topology_and_Push_Policy.md) — git topology | 300           | pending    |

---

## 3. Anti-Bluff Verification

Per [Master Plan §4.3](../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Sources resolved

| Path                                                              | Lines  | Role                                            |
|-------------------------------------------------------------------|-------:|-------------------------------------------------|
| [`../00_Master_Plan.md`](../00_Master_Plan.md) §3, §7.2            |    1k+ | family chapter list (O01..O06)                 |
| [`../01_Constitution.md`](../01_Constitution.md) §3 + §4 + §7..§11 |    879 | R-05/06/07/08/10/16/17/18                       |
| [`../06_Submodules/`](../06_Submodules/) S01..S04                  | 12,419 | the per-submodule + container + Challenges + HelixQA contracts this family operationalises |
| [`../07_Testing/`](../07_Testing/) T01..T12                        |  4,533 | the test discipline this family deploys         |

### Forbidden patterns

Index prose scanned: clean.

### Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Pending: O01–O06 chapters, each with their own Anti-Bluff block.
- Reviewed by: pending operator review.

End of `08_Operations/00_Index.md` — 2026-04-30.
