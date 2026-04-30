# 07_Testing/ — Family Index

> **Status:** Draft v1.
> **Last updated:** 2026-04-30.
> **Purpose:** Navigation hub for the *Testing & QA* family — the fifth aggregation family in the synthesis programme.
> **Targets (R-XX):** R-11 (Ten test types per submodule, 100 % coverage), R-12 (only Unit may use mocks/stubs/hardcoded values; the other nine MUST drive a real production-like system), R-13 (anti-bluff — green tests guarantee real end-user-usable behaviour), R-14 (Challenges discipline integrated as in HelixAgent + Catalogizer; HelixQA fully integrated).
> **Cross-links:** [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row block T0X; [`../01_Constitution.md`](../01_Constitution.md) §1 (Anti-Bluff Pledge), §6 (Testing Discipline), §7 (Quality Gates), §11.5 (R-18 Operational Integrity); [`../06_Submodules/`](../06_Submodules/) — the 29-submodule catalog this family tests; [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) (S03 — the Challenges row that this family's test matrix terminates in); [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) (S04 — the autonomous orchestrator that drives every cadence).

---

## 1. Position in the synthesis programme

The Testing family is the **fifth of nine** families in `05_Response/` and the **second aggregation** family (the first being [`06_Submodules/`](../06_Submodules/)). Whereas the Submodules family ratifies the *catalog* and *cross-cutting policies*, the Testing family ratifies the *test discipline* — the matrix every submodule must satisfy, the per-test-type playbook, and the meta-test that closes the anti-bluff loop.

| # | Family               | Role          | Status                                                       |
|---|----------------------|---------------|--------------------------------------------------------------|
| 1 | `03_Architecture/`   | content       | 13 chapters, ≈ 36,000 lines, **closed**                      |
| 2 | `04_Latency/`        | content       | 11 chapters, 16,666 lines, **closed**                        |
| 3 | `05_Video_Audio/`    | content       | 13 chapters, 30,802 lines, **closed**                        |
| 4 | `06_Submodules/`     | aggregation   | 4 + 29 chapters, 12,419 lines, **closed**                    |
| 5 | `07_Testing/`        | aggregation   | this family — **in progress**                                |
| 6 | `08_Operations/`     | aggregation   | follows Testing — pending                                     |
| 7 | `09_Implementation_Phases/` | execution plan | follows Testing + Operations — pending                |
| 8 | `99_Web_Research_Addenda/` | living research | append-only                                              |
| 9 | `00–02` foundation   | governance    | **closed**                                                    |

The family does **not** introduce new submodules. Each submodule's API surface, R-18 wrapper, and per-test-type matrix are defined in [`../06_Submodules/per-submodule/<name>.md`](../06_Submodules/per-submodule/) §5. What this family adds:

1. **The canonical Test Matrix** (T01) — one matrix that maps every of the 29 submodules × 10 test types × 4 mirror-CI runners = 1,160 cells, with each cell's row defining ownership (in-tree vs delegated), cadence trigger (per-PR / nightly / canary / pre-release), tooling (govulncheck, Snyk, VMAF, Toxiproxy, etc.), and coverage gate (≥ 95 % unit; reach-criteria for E2E; no-mocks-allowed for nine of ten).
2. **Per-test-type playbooks** (T02..T12) — one chapter per test type with the canonical setup, fixtures, anti-pattern catalogue, and CI-lane invocation pattern.

---

## 2. Family chapter list

Per [Master Plan §3](../00_Master_Plan.md#3-output-structure):

| ID  | Chapter                                                                                 | Floor (lines) | Status     |
|-----|-----------------------------------------------------------------------------------------|--------------:|------------|
| —   | [`00_Index.md`](00_Index.md) — this file                                                  | navigation    | **draft**  |
| T01 | [`01_Test_Matrix.md`](01_Test_Matrix.md) — the 29 × 10 × 4 matrix                        | 600           | pending    |
| T02 | [`02_Unit_Tests.md`](02_Unit_Tests.md) — Unit (mocks allowed)                            | 300           | pending    |
| T03 | [`03_Integration_Tests.md`](03_Integration_Tests.md) — Integration (no mocks)            | 300           | pending    |
| T04 | [`04_E2E_Tests.md`](04_E2E_Tests.md) — End-to-end                                        | 300           | pending    |
| T05 | [`05_Security_Tests.md`](05_Security_Tests.md) — govulncheck + Snyk + Trivy + fuzz       | 300           | pending    |
| T06 | [`06_Benchmarking.md`](06_Benchmarking.md) — HDR histogram + p999 + benchstat            | 300           | pending    |
| T07 | [`07_Chaos.md`](07_Chaos.md) — Toxiproxy + chaos-mesh + fault injection                  | 300           | pending    |
| T08 | [`08_Stress.md`](08_Stress.md) — 24-hour steady-state + soak                             | 300           | pending    |
| T09 | [`09_Smoke.md`](09_Smoke.md) — 30-second post-deploy sanity                              | 300           | pending    |
| T10 | [`10_Full_Automation.md`](10_Full_Automation.md) — orchestrating §1..§9 in CI matrix     | 300           | pending    |
| T11 | [`11_Challenges.md`](11_Challenges.md) — meta-test consuming `vasic-digital/Challenges`  | 500           | pending    |
| T12 | [`12_HelixQA_Autonomous.md`](12_HelixQA_Autonomous.md) — HelixQA orchestration playbook | 400           | pending    |

T11 (Challenges) and T12 (HelixQA) carry slightly larger floors because they are the *operational meeting points* with [S03](../06_Submodules/03_Challenges_Submodule.md) and [S04](../06_Submodules/04_HelixQA_Integration.md) respectively.

---

## 3. The Ten test types — at a glance

Per [Constitution §6.2](../01_Constitution.md#6-testing-discipline-r-11-r-12-r-13):

| #   | Type           | Mocks?           | Hits real system?      | Cadence         | Owner            | Anti-bluff role                          |
|-----|----------------|------------------|-------------------------|-----------------|------------------|-------------------------------------------|
| 1   | Unit           | **YES** (R-12)   | partial                 | per-PR          | submodule        | API contract                              |
| 2   | Integration    | NO               | yes (real deps)        | per-PR          | submodule        | dependency-glue correctness               |
| 3   | E2E            | NO               | yes (full topology)    | per-PR + nightly| submodule        | user-journey shape                        |
| 4   | Security       | NO               | yes (live scanners)    | per-PR + monthly| submodule        | CVE + reachability                        |
| 5   | Benchmarking   | NO               | yes (real workload)    | nightly         | submodule        | latency p999 / throughput regression      |
| 6   | Chaos          | NO               | yes (with injected fault)| nightly       | submodule        | fault tolerance                           |
| 7   | Stress         | NO               | yes (24-hour soak)     | canary           | submodule        | resource leaks                            |
| 8   | Smoke          | NO               | yes (post-deploy)      | post-deploy      | submodule        | basic liveness                            |
| 9   | Full Automation| inherits         | inherits                | per-PR + nightly| submodule        | orchestrates 1..8                         |
| 10  | **Challenges** | **NO**           | **yes (full system)**  | nightly + canary + pre-release | HelixQA + S03 | **the anti-bluff backstop** |

Only Unit may use mocks / stubs / hardcoded values. R-13 makes this rule operationally significant — past projects had green Unit + green Integration that hid broken end-user behaviour because Unit's mocks misrepresented the real dependency. Challenges is the meta-test that closes the loop because it boots the entire production-like topology and observes user-visible behaviour.

---

## 4. Coverage targets (R-11)

R-11 demands **100 % coverage** per submodule. The breakdown per test type:

- **Unit**: ≥ 95 % statement coverage measured by `go test -coverprofile`. The 5 % gap accommodates unreachable error branches (e.g. `panic(unreachable)` after exhaustive switch-cases) — auditable, not aspirational.
- **Integration**: every public exported symbol exercised at least once with a real dependency.
- **E2E**: every reference user journey from [System Overview §3](../02_System_Overview.md#3-reference-user-journey) covered.
- **Security**: zero high-severity findings on every PR.
- **Benchmarking**: every submodule's documented performance budget (per its S05 §9.2 table) measured + p999 within tolerance.
- **Chaos**: every documented failure mode (per S05 §9.3 *Common errors* table) injected + recovery verified.
- **Stress**: 24-hour soak at the workload's documented peak rate; zero memory growth, zero fd leak, zero goroutine leak.
- **Smoke**: 30-second post-deploy invocation verifying the binary started and answered one request.
- **Full Automation**: orchestrates 1–8 with fail-fast disabled to surface maximum issues per PR.
- **Challenges**: cross-fleet baseline parity (per [S03 §6.3 thresholds](../06_Submodules/03_Challenges_Submodule.md#63-change-point-detection)).

The matrix below lists which test types of which submodules are owned in-tree vs delegated; T01 §2 freezes the 29 × 10 cell ownership.

---

## 5. Tooling lockstep across the fleet

Per [S01 §4](../06_Submodules/01_Submodule_Catalog.md#4-cross-cutting-policies--eight-surfaces-with-no-chapter-§6-home), the toolchain is the same across every submodule:

- **Unit + Integration + E2E**: Go's `testing` package + `github.com/stretchr/testify` + `github.com/onsi/gomega` (selected per submodule taste; no project-wide preference because the Go-stdlib path already covers most cases).
- **Security**: `govulncheck`, Snyk, Trivy (container), custom fuzzers via `go-fuzz` or stdlib fuzz.
- **Benchmarking**: [`helix-bench`](../06_Submodules/per-submodule/helix-bench.md) harness; `benchstat` for significance.
- **Chaos**: Toxiproxy (network); chaos-mesh (Kubernetes-orchestrated); custom syscall-rejection harness for the [`host-integrity-scan`](../06_Submodules/02_Containers_Submodule.md#34-the-host-integrity-scan-shared-lane) row.
- **Stress**: same harness as Benchmarking, run for 24 hours.
- **Smoke**: per-submodule shell script invoked by HelixQA's deployment-gate post-deploy probe (S04 §9).
- **Full Automation**: GitHub Actions / GitLab CI workflow YAML (in `vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml`).
- **Challenges**: `vasic-digital/Challenges` topology-driven — see [S03](../06_Submodules/03_Challenges_Submodule.md).

The four-mirror amplifier (per [S01 §4.7](../06_Submodules/01_Submodule_Catalog.md#47-public-visibility-enforcement-across-four-mirrors)) means every test type runs four times — once per mirror — with parity audit nightly. A Russian-mirror green / Western-mirror red split (or vice versa) is documented in [S04 §6.3](../06_Submodules/04_HelixQA_Integration.md#63-the-reverse-a-russian-mirror-green-western-red).

---

## 6. R-18 inheritance for the test matrix

Every test type that runs inside a container inherits the [S01 §7 R-18 ladder](../06_Submodules/01_Submodule_Catalog.md#7-the-r-18-inheritance-ladder-canonical-from-helix-r18-safeexec-outward) plus the [S02 §3.4 host-integrity-scan harness](../06_Submodules/02_Containers_Submodule.md#34-the-host-integrity-scan-shared-lane). The five-layer enforcement model applies to every test invocation:

1. Chapter prose documents the test row's expected R-18 surface.
2. Static deny-list — `helix-r18-safeexec`'s private constant.
3. Runtime `SafeExec` wrapper — every `os/exec` invocation in test fixtures.
4. Ripgrep CI lane — verifies no direct `exec.Command` slipped into test code.
5. Host-integrity-scan strace + auditd — verifies no forbidden syscall fires during the test run.

Test infrastructure that violates layer 4 (e.g. a test fixture that calls `exec.Command("pm-suspend", "...")` even via SafeExec — which would still be rejected by layer 2) is rejected at PR review. The discipline is non-negotiable per [Constitution §11.5.4](../01_Constitution.md#115-operational-integrity-r-18).

---

## 7. Cadence orchestration (S04 §5)

The 10 test types map to four cadences per [S04 §5](../06_Submodules/04_HelixQA_Integration.md#5-per-cadence-orchestration-per-pr--nightly--canary--pre-release):

| Cadence       | Test types invoked                                     | Wall-clock (29-submodule fleet)|
|---------------|--------------------------------------------------------|-------------------------------|
| Per-PR        | 1 (Unit) + 2 (Integration) + 3 (E2E primary) + 4 (Security per-submodule) + 9 (Smoke if post-deploy) | 5–15 min   |
| Nightly       | All of per-PR + 5 (Benchmarking) + 6 (Chaos) + 10 (Challenges per-submodule)                          | 2–4 h       |
| Canary        | Nightly + 7 (Stress 24-h subset) + 10 (full 14-topology Challenges fan-out)                           | 12–24 h     |
| Pre-release   | Canary + 7 (full 24-h Stress) + 10 (30-day exhaustive Challenges replay)                              | 24–48 h     |

The cadence is **non-overridable**. A PR that disables a per-PR test invocation is rejected by the per-submodule CI lane's required-checks rule.

---

## 8. Family-level cross-links

- **Architecture family**: [`../03_Architecture/`](../03_Architecture/) — every chapter §6 documents its submodule's R-18 surface, which the test matrix verifies at layer 5.
- **Latency family**: [`../04_Latency/`](../04_Latency/) — `helix-bench` originates here (C24 §6); the Benchmarking row (T06) is structurally a pull-through to that submodule.
- **Video/Audio family**: [`../05_Video_Audio/`](../05_Video_Audio/) — `helix-vqa` originates here (C35 §6); the Challenges video-quality scoring (per T11) consumes its VMAF + ViSQOL primitives.
- **Submodules family**: [`../06_Submodules/`](../06_Submodules/) — the 29-submodule catalog this family tests; every per-submodule §5 row is a cell in T01's matrix.
- **Operations family** (forthcoming): [`../08_Operations/`](../08_Operations/) — the deployment-gate operationalises T11's Challenges + T12's HelixQA cadences.
- **Implementation Phases** (forthcoming): [`../09_Implementation_Phases/`](../09_Implementation_Phases/) — Phase_02_Core_Submodules deliverables include the per-submodule Ten-test-type matrix; the Testing family is its specification.

---

## 9. Anti-Bluff Verification (family-level)

Per [Master Plan §4.3](../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Sources resolved in this index

| Path                                                             | Lines  | Role                              |
|------------------------------------------------------------------|-------:|-----------------------------------|
| [`../00_Master_Plan.md`](../00_Master_Plan.md) §3, §7.2           |    1k+ | family chapter list (T01..T12)    |
| [`../01_Constitution.md`](../01_Constitution.md) §1 + §6 + §7 + §11.5 |    879 | R-02 + R-11 + R-12 + R-13 + R-18 |
| [`../06_Submodules/`](../06_Submodules/) all chapters             | 12,419 | per-submodule §5 cell ownership  |

### Forbidden patterns

Index prose scanned: clean. No `TODO` / `FIXME` / `tbd` / `xxx` / `???` / `placeholder` / "fill in later" markers. The structural placeholder for T02–T12 chapters in §2 is a queued-work indicator (each row's `Status: pending`), not a content placeholder.

### Sign-off

- Index drafted by orchestrator (Claude) on 2026-04-30.
- Pending: T01–T12 chapters, each with their own Anti-Bluff block.
- Reviewed by: pending operator review.

End of `07_Testing/00_Index.md` — 2026-04-30.
