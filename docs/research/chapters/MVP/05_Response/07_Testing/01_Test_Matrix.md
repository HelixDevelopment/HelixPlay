# T01 — The Master Test Matrix

> **Source dimensions:**
> - [`00_Index.md`](00_Index.md) — family index (draft v1, 2026-04-30, 177 lines)
> - [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §3.1 (29 rows), §5 (per-submodule Ten / inline + 2 delegation exceptions)
> - [`../06_Submodules/per-submodule/`](../06_Submodules/per-submodule/) — 29 descriptors, each carrying §5 and §11 *Per-test-type coverage targets*
> - [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3 (per-submodule CI lanes), §5 (test matrix container-bound rule), §8 (R-18 hazard policies that survive every test invocation)
> - [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §4 (29 + helix-shm-delegation Challenges entry-point map)
> - [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §5 (per-cadence orchestration), §9 (deployment gate)
> - [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row T01 (≥ 600-line floor)
> - [`../01_Constitution.md`](../01_Constitution.md) §1 (Anti-Bluff Pledge), §6 (Testing Discipline R-11 + R-12 + R-13), §7 (Quality Gates R-10), §11.5 (R-18 Operational Integrity), §16 (Acceptance)
>
> **Source line count:** family inputs ≈ 1k+ lines from S01..S04 + 12,419 lines from `06_Submodules/` consumed by reference rather than duplicated. The §7.2 chapter floor is 600 lines; this chapter exceeds the floor on the strength of the cell-by-cell matrix definition.
>
> **Chapter targets:** R-02 + R-11 + R-12 + R-13 + R-14 + R-17 + R-18.
>
> **Cross-links:** [`02_Unit_Tests.md`](02_Unit_Tests.md) (T02), [`03_Integration_Tests.md`](03_Integration_Tests.md) (T03), [`04_E2E_Tests.md`](04_E2E_Tests.md) (T04), [`05_Security_Tests.md`](05_Security_Tests.md) (T05), [`06_Benchmarking.md`](06_Benchmarking.md) (T06), [`07_Chaos.md`](07_Chaos.md) (T07), [`08_Stress.md`](08_Stress.md) (T08), [`09_Smoke.md`](09_Smoke.md) (T09), [`10_Full_Automation.md`](10_Full_Automation.md) (T10), [`11_Challenges.md`](11_Challenges.md) (T11), [`12_HelixQA_Autonomous.md`](12_HelixQA_Autonomous.md) (T12).
>
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## Table of Contents

1. The Matrix in Five Sentences
2. The 29 × 10 × 4 Cell Grid (the canonical lookup)
3. Cell Ownership Rules
4. The Two Documented Delegation Exceptions
5. Cadence × Test-Type Cross-Tab
6. Coverage Targets per R-11
7. Mock Policy Enforcement (R-12)
8. Anti-Bluff Guarantees (R-13)
9. Four-Mirror Parity Contract
10. Tooling Lockstep — Pinned Versions Across the Fleet
11. R-18 Inheritance Across the Test Matrix
12. Test Selection — Which Tests Run on Which Cadence
13. CI Lane Anatomy — From PR to Green
14. The "No-Mock-In-Production-Path" Audit
15. Open Questions
16. References & Anti-Bluff Verification

---

## 1. The Matrix in Five Sentences

The Master Test Matrix is **29 submodules × 10 test types × 4 CI-runner mirrors = 1,160 cells**. Each cell answers four questions: *who owns it* (the submodule itself or a sibling via documented delegation), *what cadence triggers it* (per-PR / nightly / canary / pre-release), *what tooling executes it* (govulncheck, Snyk, VMAF, Toxiproxy, etc.), and *what gate it satisfies* (≥ 95 % statement coverage; reach-criteria; baseline-parity). The matrix is the operationalisation of [Constitution §6](../01_Constitution.md#6-testing-discipline-r-11-r-12-r-13) (R-11 + R-12 + R-13) and the contract every submodule's [S05 §5](../06_Submodules/per-submodule/) test-matrix entry resolves to. **Only Unit may use mocks; the other nine MUST hit a real production-like system per R-12** — the all-three-pass deployment gate at [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad) refuses to deploy any image whose Challenges row is anything but green. Past incidents had green tests on broken features, which R-13 forbids; the matrix is the primary mitigation.

---

## 2. The 29 × 10 × 4 Cell Grid (the canonical lookup)

The grid is rendered here in compressed form: rows are the 29 submodules; columns are the 10 test types; each cell is **`O` (in-tree) / `D-<sibling>` (delegated) / `T` (terminal-only — applies to the 9 non-mock test types' container-bound nature)**. The mirror axis is implicit: every cell runs on all four mirrors with parity enforced (§9).

### 2.1 The 29 × 10 ownership grid

| Submodule (S01 §3.1)     | Unit | Int | E2E | Sec | Bench | Chaos | Stress | Smoke | FA | Chal             |
|--------------------------|:----:|:---:|:---:|:---:|:-----:|:-----:|:------:|:-----:|:--:|:----------------:|
| 01 helix-r18-safeexec    |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 02 helix-grpc-frame      |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 03 helix-tv-input        |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 04 helix-vault           |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 05 helix-tenant          |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 06 helix-shm             |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O | **D-pipeline** *(§4.1)* |
| 07 helix-iouring         |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 08 helix-xdp             |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 09 helix-lockfree        |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 10 helix-gpu-direct      |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 11 helix-network         |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 12 helix-rtos            |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 13 helix-input           |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 14 helix-display         |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 15 helix-mempool         |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 16 helix-allocator       |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 17 helix-bench           |  O   |  O  |  O  |  O  | **D-workload** *(§4.2)* | O | O | O | O |        O         |
| 18 helix-codec           |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 19 helix-encoder         |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 20 helix-capture         |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 21 helix-dualpath        |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 22 helix-record          |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 23 helix-audio           |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 24 helix-hdr             |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 25 helix-abr             |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 26 helix-thermal         |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 27 helix-vqa             |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |
| 28 helix-pipeline        |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |  O *(also owns helix-shm delegation)* |
| 29 helix-transport       |  O   |  O  |  O  |  O  |   O   |   O   |   O    |   O   |  O |        O         |

**Cell count audit**: 29 × 10 = 290 cells. Of these, **288 are `O` (in-tree)** and **2 are documented delegations** — `helix-shm.Challenges → helix-pipeline` and `helix-bench.Benchmarking → workload-owner`. The matrix has **zero unowned cells**; every test-type slot has a named owner.

### 2.2 The mirror axis

Every `O` and `D-*` cell additionally carries the mirror axis: `{github, gitlab, gitflic, gitverse}`. The expectation is **identical green / red verdict on every mirror**; divergence is documented at [S04 §6.3](../06_Submodules/04_HelixQA_Integration.md#63-the-reverse-a-russian-mirror-green-western-red) and triggers a P2 alert.

### 2.3 Total cell count

29 × 10 × 4 = **1,160 cells**. At per-PR cadence (smallest scope), a PR touching one submodule fires 1 × 10 × 4 = 40 cells (one per test type per mirror); a PR touching all 29 submodules fires the full 1,160. The CI cost model in [S01 §4.6.3](../06_Submodules/01_Submodule_Catalog.md#463-per-pr-cost-budget) sizes for the worst case.

---

## 3. Cell Ownership Rules

### 3.1 Default — `O` (in-tree)

A cell labelled `O` means the submodule's own repository hosts the test code under `tests/<test-type>/`. The CI lane (per [S02 §3](../06_Submodules/02_Containers_Submodule.md#3-the-per-submodule-ci-lane-catalog)) invokes the test row inside the per-submodule container. This is the default for **288 of 290** cells.

### 3.2 Delegation — `D-<sibling>`

A cell labelled `D-<sibling>` means the test row is *not* hosted in the submodule's own repository; the named sibling submodule's tests cover this row indirectly. The full set of delegations is exactly **two** cells:

- `helix-shm.Challenges → helix-pipeline` ([S01 §5.2](../06_Submodules/01_Submodule_Catalog.md#52-the-two-delegation-exceptions); [S03 §4.2](../06_Submodules/03_Challenges_Submodule.md#42-the-helix-shm-challenges-delegation-s01-§52))
- `helix-bench.Benchmarking → workload-owner` ([S01 §5.2](../06_Submodules/01_Submodule_Catalog.md#52-the-two-delegation-exceptions))

Both are **explicitly documented** in S01 + S03 + the per-submodule descriptors. Adding a 3rd delegation requires Constitution amendment per §15.

### 3.3 The "no unowned cell" invariant

A cell with neither `O` nor `D-*` is a defect. The CI lane's required-checks include `helix-test-matrix-lint` (a custom lint defined in `vasic-digital/Containers/ci-fragments/`) that loads this T01 §2 matrix and asserts every (submodule, test-type) pair has either `O` or a recognised `D-*` value. A PR that introduces a new submodule without filling its row in this matrix is rejected.

---

## 4. The Two Documented Delegation Exceptions

### 4.1 `helix-shm.Challenges → helix-pipeline`

`helix-shm` is a library — page allocation + mmap + frame plane accessors. It has no observable user-visible behaviour by itself; observable behaviour appears only when consumers (capture, encoder, transport) write and read its pages. A Challenges-class test, by definition, requires a full system to be running, which `helix-shm` cannot stand up alone.

The delegation contract is fixed in `helix-shm/tests/challenges/delegated.md`:

> This submodule does not stand up a full system. Its Challenges row is exercised by `helix-pipeline` under the `vasic-digital/Challenges/topologies/01_minimum_viable_session/scenarios/06_full_pipeline_end_to_end.scenario.yaml` scenario, which uses every public `helix-shm` API. For change-point detection on `helix-shm`-specific metrics, see `vasic-digital/Challenges/baselines/01_minimum_viable_session/06_full_pipeline_end_to_end/per-submodule/helix-shm.metrics.json` which records the helix-shm slice of the full-pipeline baseline.

The per-submodule baseline file means **`helix-shm` regressions are still attributed to `helix-shm`** even though the test was driven by `helix-pipeline`. The accountability survives the delegation.

### 4.2 `helix-bench.Benchmarking → workload-owner`

`helix-bench` is the harness, not the workload. It samples wall-clock; it produces HDR histograms; it does benchstat comparisons. But the *thing being benchmarked* is something else — `helix-shm`'s zero-copy throughput, `helix-iouring`'s I/O completion rate, `helix-transport`'s send-path latency.

The benchmarking row at `helix-bench`'s tier verifies the **harness itself** is correct (does it sample at the requested rate; does it compute p999 correctly; does it survive a 10K-sample run). The workload-specific benchmarks live in the workload submodule's own `tests/benchmarking/` directory.

This is **delegation-with-cross-references**, not full delegation. The audit tooling treats it as such.

---

## 5. Cadence × Test-Type Cross-Tab

Per [S04 §5](../06_Submodules/04_HelixQA_Integration.md#5-per-cadence-orchestration-per-pr--nightly--canary--pre-release), the four cadences map to the 10 test types:

| Test type            | Per-PR | Nightly | Canary | Pre-release |
|----------------------|:------:|:-------:|:------:|:-----------:|
| 1 Unit               |   ✓    |    ✓    |   ✓    |      ✓      |
| 2 Integration        |   ✓    |    ✓    |   ✓    |      ✓      |
| 3 E2E (primary scenario) |   ✓    |    ✓    |   ✓    |      ✓      |
| 4 E2E (full secondary set) |   —    |    ✓    |   ✓    |      ✓      |
| 5 Security           |   ✓    |    ✓    |   ✓    |      ✓      |
| 6 Benchmarking (smoke) |   —    |    ✓    |   ✓    |      ✓      |
| 7 Chaos (per-submodule fault) |   —    |    ✓    |   ✓    |      ✓      |
| 8 Stress (24-h subset) |   —    |    —    |   ✓    |      ✓      |
| 9 Stress (full 24-h)  |   —    |    —    |   —    |      ✓      |
| 10 Smoke (post-deploy) |   —    |    —    |   ✓    |      ✓      |
| 11 Full Automation   |   ✓    |    ✓    |   ✓    |      ✓      |
| 12 Challenges (per-submodule primary) |   ✓    |    ✓    |   ✓    |      ✓      |
| 13 Challenges (all 14 topologies × all scenarios) |   —    |    —    |   ✓    |      ✓      |
| 14 Challenges (30-day exhaustive replay) |   —    |    —    |   —    |      ✓      |

Reading the table: a per-PR run includes Unit + Integration + primary E2E + Security + Full Automation + per-submodule primary Challenges scenario. Nightly adds full E2E + Benchmarking + Chaos. Canary adds 24-h Stress subset + Smoke + full Challenges fan-out. Pre-release adds the full 24-h Stress + 30-day exhaustive Challenges replay.

The cadence is **non-overridable**. A PR that disables a per-PR row is rejected by the per-submodule CI lane's required-checks rule. The list of required checks lives in [S04 §5.1](../06_Submodules/04_HelixQA_Integration.md#51-per-pr-cadence) and is enforced by GitHub branch protection, GitLab Protected Branches, GitFlic + GitVerse equivalents.

---

## 6. Coverage Targets per R-11

R-11 (Constitution §6.1) demands **100 % coverage** per submodule. The breakdown:

### 6.1 Unit (≥ 95 % statement coverage)

Measured by `go test -coverprofile=coverage.out -covermode=atomic ./...`. The 5 % gap is reserved for:

- Unreachable error branches (e.g. `panic("unreachable: switch case exhausted")` after an exhaustive switch).
- Compiler-generated synthetic functions (init, empty interface methods).
- Auto-generated code (protobuf-generated stubs).

A PR whose Unit coverage falls below 95 % is rejected. The auto-generated-code allowance is documented per submodule in `<submodule>/.coverage-exemptions.yaml` and audited at PR review.

### 6.2 Integration (every public exported symbol exercised)

The lint `helix-integration-coverage` walks every exported symbol in the submodule's API surface (per its `<submodule>/README.md` API section + the godoc) and verifies that at least one Integration test invokes it. Any unexercised export is a defect.

### 6.3 E2E (every reference user journey from System Overview §3)

[System Overview §3](../02_System_Overview.md#3-reference-user-journey) defines the canonical user journey. Every step in the journey has at least one E2E test. The journey-coverage lint `helix-e2e-journey-coverage` parses the journey definition and verifies every step is reachable from at least one test fixture.

### 6.4 Security (zero high-severity findings)

govulncheck + Snyk + Trivy all return exit code 0; high-severity findings (CVSS ≥ 7.0) block the merge. Medium and low findings are reported but do not block. Licence-policy violations (per [S01 §4.5.2](../06_Submodules/01_Submodule_Catalog.md#451-govulncheck--the-go-aware-call-graph-scanner)) block immediately.

### 6.5 Benchmarking (p999 budget compliance)

Every submodule's S05 §9.2 *Performance budget* table defines its p50/p99/p999 targets. The Benchmarking row runs the workload via `helix-bench` and asserts p999 within ±5 % of the recorded baseline. A regression > 5 % blocks the merge; an improvement > 10 % requires baseline-replacement PR per [S03 §6.1](../06_Submodules/03_Challenges_Submodule.md#61-recording).

### 6.6 Chaos (every documented failure mode injected)

Every submodule's S05 §9.3 *Common errors* table is the chaos-injection target list. The chaos test fires each error condition (e.g. "Toxiproxy partition", "GPU memory pressure") and verifies the submodule surfaces the error per the documented remediation. Failure to surface is a defect.

### 6.7 Stress (24-hour soak; zero leak)

The stress test runs the workload at peak documented rate for 24 hours. Leak detection: every 1-minute boundary, the runtime captures `runtime.MemStats` + `runtime.NumGoroutine() + /proc/<pid>/fd` count + open file descriptors. Linear growth on any metric over the 24-hour window is a defect.

### 6.8 Smoke (30-second post-deploy)

After deployment, a 30-second probe verifies the binary started, answered one request, and didn't crash. Smoke failures auto-revert the deployment per [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad).

### 6.9 Full Automation (orchestrates 1–8 with fail-fast disabled)

The Full Automation test orchestrator is `vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml`. Fail-fast is disabled (`fail-fast: false`) so a Unit failure does not skip the Integration / E2E / Security rows — every row runs and reports independently. This surfaces the maximum issue count per PR; the maintainer fixes everything in one cycle rather than discover-fix-rediscover-fix sequentially.

### 6.10 Challenges (cross-fleet baseline parity)

The Challenges row's coverage criterion is **baseline parity** per [S03 §6.3](../06_Submodules/03_Challenges_Submodule.md#63-change-point-detection): every metric in the recorded baseline must match the fresh observation within tolerance. Tolerances are **per-topology + per-scenario**, not global, and are configured in `vasic-digital/Challenges/harness/change-point/threshold-config.yaml`.

---

## 7. Mock Policy Enforcement (R-12)

R-12 (Constitution §6.2) reads:

> Only Unit tests may use mocks/stubs/hardcoded values. Every other test type must drive a real, production-like system with all containers running.

Enforcement is at three layers:

### 7.1 Layer 1 — Source-tree rule

Per the canonical S05 §5 ten-test-type structure, the directory layout fixes the rule:

```
<submodule>/tests/
├── unit/         ← mocks allowed (R-12)
├── integration/  ← real DB, real Redis, real co-located submodule binaries
├── e2e/          ← full container-up stack via vasic-digital/Containers
├── security/     ← govulncheck + Snyk + custom fuzz
├── benchmarking/ ← Go benchstat + vasic-digital/helix-bench
├── chaos/        ← Toxiproxy + chaos-mesh (kubernetes manifests)
├── stress/       ← long-running load tests, 24-hour profile
├── smoke/        ← post-deploy 30-second sanity runs
├── full-automation/ ← invokes 1..9 above in CI matrix
└── challenges/   ← vasic-digital/Challenges integration (S03)
```

Each subdirectory's tests have their own discipline. A test under `unit/` may import a mock package; a test under `integration/` (or any of `e2e/`, `chaos/`, etc.) MUST NOT.

### 7.2 Layer 2 — Static analyser

The `helix-mock-discipline` lint (a custom Go vet plugin shipped under `vasic-digital/.github/`) scans test files and rejects any import of `github.com/golang/mock/gomock`, `github.com/stretchr/testify/mock`, or any package matching `*mock*` from a non-`unit/` directory. The lint runs on every PR.

### 7.3 Layer 3 — Container-bound execution

Per [S02 §5](../06_Submodules/02_Containers_Submodule.md#55-the-container-lane-consequence), every non-Unit test runs **inside the per-submodule container**. The container brings up the real backing services (CockroachDB, NATS, Redis, etc.) per the per-submodule CI lane's compose spec. A test that pretends to talk to a mocked database while running in the container would still hit the real container's localhost; the only way to use a mock would be to inject one *into* the container, which the layer-1 directory rule already forbids.

The three layers together close the loop. The mock policy is not aspirational; it is enforced.

---

## 8. Anti-Bluff Guarantees (R-13)

R-13 demands that green tests guarantee real, end-user-usable behaviour. The matrix's anti-bluff guarantees:

### 8.1 Per-test-type guarantee

| Test type      | What "green" means                                                      |
|----------------|-------------------------------------------------------------------------|
| Unit           | API contract holds (modulo mock-misrepresentation risk).                |
| Integration    | Real dependencies behave as expected (closes Unit's mock-risk gap).     |
| E2E            | At least one user journey works.                                         |
| Security       | No known vulnerability in the dependency closure (modulo unreachability). |
| Benchmarking   | p999 latency / throughput within budget.                                |
| Chaos          | Documented failures recoverable.                                         |
| Stress         | No resource leak over 24 h.                                              |
| Smoke          | Binary started.                                                          |
| Full Automation| All of 1–8 above are simultaneously green.                              |
| **Challenges** | **The full system reproduces the recorded baseline** — the anti-bluff backstop. |

### 8.2 Why Challenges is the backstop

The first 9 test types each have a known failure mode (per [S03 §2](../06_Submodules/03_Challenges_Submodule.md#2-why-challenges-is-the-tenth-test-type-r-13-anti-bluff-backstop)). Challenges closes the loop because it boots the entire production-like topology and observes user-visible artefacts (rendered video frames, audio waveforms, controller round-trips, gRPC trace topology, billing events). A submodule's bug invisible to types 1–9 cannot be invisible to Challenges, because the bug, by definition, manifests as user-visible behaviour drift.

### 8.3 Challenges' own failure modes (and mitigations)

Per [S03 §7](../06_Submodules/03_Challenges_Submodule.md#7-anti-bluff-enforcement-at-the-challenges-boundary):

| Failure mode              | Mitigation                                                            |
|---------------------------|-----------------------------------------------------------------------|
| Baseline tampering        | minisign-signed manifests + non-delegable operator review on baselines |
| Topology drift            | digest-pinned compose spec + four-mirror replication audit            |
| Scenario narrowing        | DSL with JSON-Schema validation + minimum-step enforcement            |
| Observation gaps          | heartbeat manifests + recorder-required artefacts                     |
| Threshold inflation       | two-reviewer rule on `threshold-config.yaml`                          |

The five mitigations together close the practical gaps. R-13 is not aspirational; it is operationally enforced.

---

## 9. Four-Mirror Parity Contract

Every cell of the matrix runs on all four CI runner topologies — GitHub Actions, GitLab CI, GitFlic CI, GitVerse CI. The parity contract:

- Same scenario / scenario / Challenge runs.
- Same observation produced (within S03 §6.4 perturbation tolerance).
- Same change-point verdict.

Divergence triggers an alert per [S04 §6.3](../06_Submodules/04_HelixQA_Integration.md#63-the-reverse-a-russian-mirror-green-western-red):

| Divergence pattern             | Severity | Alert                                        |
|--------------------------------|----------|----------------------------------------------|
| 1 of 4 mirrors red             | P3       | flaky scenario; probably runner regression   |
| 2 of 4 mirrors red             | P2       | real regression candidate                    |
| 3+ of 4 mirrors red            | P1       | fleet-down candidate                         |
| Russian mirrors green / Western red | P1   | sanctions-related runtime difference; investigate |

The `four-mirror-challenges-parity.sh` audit script runs nightly and produces a parity report.

---

## 10. Tooling Lockstep — Pinned Versions Across the Fleet

Per [S01 §4.3](../06_Submodules/01_Submodule_Catalog.md#43-dependency-lockstep--gosum-goproxy-gosumdb-renovate), every submodule pins the same toolchain. The pin-set as of 2026-04-30:

| Tool                     | Pinned version              | Purpose                                                |
|--------------------------|-----------------------------|--------------------------------------------------------|
| Go toolchain             | `1.24.x` (latest patch)     | base                                                    |
| `govulncheck`            | latest from `vuln.go.dev`   | reachability-aware vuln scan                            |
| Snyk                     | `snyk-cli ≥ v1.1300`        | quality gate (Constitution §7)                         |
| Trivy                    | `≥ v0.55.0`                 | container CVE scan                                      |
| `golangci-lint`          | `≥ v1.62.0`                 | aggregate Go linters                                   |
| `cyclonedx-gomod`        | `≥ v1.6.0`                  | Go SBOM                                                 |
| `syft`                   | `≥ v1.16.0`                 | container SBOM                                          |
| `cosign`                 | `≥ v2.4.0`                  | Sigstore keyless signing                                |
| `helix-bench`            | matched repo tag            | benchmark harness                                       |
| `helix-allocator-vet`    | matched repo tag            | hot-path allocation enforcer                           |
| `Toxiproxy`              | `≥ v2.10.0`                 | network impairment                                      |
| `chaos-mesh`             | `≥ v2.7.0`                  | Kubernetes chaos                                       |
| VMAF                     | `libvmaf ≥ v3.0.0`          | video QA                                                |
| ViSQOL                   | upstream stable             | audio QA                                                |
| nektos/act               | `≥ v0.2.65`                 | local CI runner                                        |

A drift between a submodule's pin and the fleet pin is detected by the `helix-toolchain-pin-audit` Renovate config and blocked at merge.

---

## 11. R-18 Inheritance Across the Test Matrix

Every cell that runs inside a container inherits the [S01 §7 R-18 ladder](../06_Submodules/01_Submodule_Catalog.md#7-the-r-18-inheritance-ladder-canonical-from-helix-r18-safeexec-outward):

1. **Chapter prose**: every chapter §6 documents the submodule's R-18 surface.
2. **Static deny-list**: `helix-r18-safeexec`'s private constant.
3. **Runtime `SafeExec` wrapper**: every test fixture invocation goes through it.
4. **Ripgrep CI lane**: rejects any direct `exec.Command` in test code.
5. **Host-integrity-scan strace + auditd**: verifies no forbidden syscall fires during the test run.

The five-layer enforcement is non-overridable per [Constitution §11.5.4](../01_Constitution.md#115-operational-integrity-r-18). A test fixture that calls `exec.Command("pm-suspend", "now")` directly (bypassing SafeExec) is rejected at layer 4. A fixture that calls `r18.SafeExec(ctx, "shutdown", "-h", "now")` is rejected at layer 2 (deny-list). Every test invocation runs under auditd in the container, and a forbidden syscall observed at layer 5 fails the CI lane regardless of test result.

---

## 12. Test Selection — Which Tests Run on Which Cadence

The matrix at §5 specifies which test types run at which cadence. Within a cadence, the **selection** of which submodules run is determined by the PR's diff:

- **PR touches one submodule**: only that submodule's row of the matrix runs at per-PR cadence.
- **PR touches `vasic-digital/.github`**: every submodule's per-PR row runs (the change is fleet-wide).
- **Nightly cadence**: every submodule's full row runs regardless of PR activity.
- **Canary cadence**: same as nightly + 24-h Stress + full Challenges fan-out.
- **Pre-release cadence**: same as canary + 30-day exhaustive Challenges replay.

The `helix-test-select` tool reads `git diff` and computes the affected submodule set. Its output is consumed by the GitHub Actions / GitLab CI workflow YAML to skip irrelevant submodule rows on per-PR runs (the only cadence where selection matters; the others run everything).

---

## 13. CI Lane Anatomy — From PR to Green

A typical PR lifecycle for one submodule:

1. **PR opened** → GitHub webhook fires.
2. **`helix-test-select`** identifies the affected submodule (one row of the matrix).
3. **Per-submodule CI lane** (S02 §3) starts on each of 4 mirrors.
4. **Container build** (S02 §5): builder image pulled; binary built with `CGO_ENABLED=0` where possible; runtime image assembled.
5. **Cosign sign + SLSA L3 provenance + dual SBOM** (S02 §6).
6. **Test row 1 (Unit)**: ≥ 95 % coverage assertion.
7. **Test row 2 (Integration)**: real dependencies up; exported-symbol coverage verified.
8. **Test row 3 (E2E)**: primary user journey replayed.
9. **Test row 4 (Security)**: govulncheck + Snyk + Trivy + spdx-check.
10. **Test row 5 (Full Automation)**: orchestrate 1–4 + Challenges-row primary scenario.
11. **Test row 10 (Challenges primary)**: invokes `vasic-digital/Challenges/scripts/challenge-run.sh`.
12. **Four-mirror parity check**: `four-mirror-challenges-parity.sh`.
13. **Required-checks gate**: all of 6–12 must be green; merge button unlocks.
14. **Merge** → push to all four mirrors via `git push origin main` (composite push).
15. **Nightly cadence** picks up the new HEAD next 02:00 Europe/Moscow.

The full lifecycle takes 5–15 minutes warm-cache per [S04 §5.1](../06_Submodules/04_HelixQA_Integration.md#51-per-pr-cadence) on amd64.

---

## 14. The "No-Mock-In-Production-Path" Audit

A monthly audit verifies the production binary contains no mock packages. The audit:

1. Pulls every released container image's syft SBOM.
2. Greps for `*mock*` package names.
3. Verifies no match.

A match would be a defect — the build pipeline accidentally compiled in a test fixture. The audit's output is a P2 alert per [S04 §8](../06_Submodules/04_HelixQA_Integration.md#8-alert-routing-and-on-call-rotation).

The audit runs nightly, but the action threshold is monthly (one match per month is investigated; sustained matches escalate to P1).

---

## 14a. The 1,160-Cell Cost Model

Each cadence's wall-clock + compute cost across the 1,160 cells (29 submodules × 10 test types × 4 mirrors). Numbers are per-mirror; per-cadence totals run in parallel across the four mirror pools.

### 14a.1 Per-PR cadence cost

A PR touching one submodule fires 1 × 10 × 4 = 40 cells. With cache warm:

| Stage                      | Wall-clock | Compute (CPU-min) |
|----------------------------|-----------:|-------------------:|
| Container build (cache hit)|   30 s     |   0.5             |
| Unit (≥ 95 % coverage)     |   30 s     |   0.5             |
| Integration (real deps)    |   2 min    |   2               |
| E2E (primary scenario)     |   3 min    |   3               |
| Security (gov+snyk+trivy)  |   1 min    |   1               |
| Smoke                      |   30 s     |   0.5             |
| Full Automation orchestrator| 30 s     |   0.5             |
| Challenges primary         |   3 min    |   3               |
| Cosign + SLSA + SBOM       |   1 min    |   1               |
| Four-mirror parity check   |   30 s     |   0.5 (×4 mirrors) |
| **Total (warm cache)**     | **~12 min**| **~14 / mirror**   |

A PR touching all 29 submodules: 12 min × 29 ≈ 5.8 hours **per mirror** if run sequentially; with the per-runner parallelism budgeted at 8 lanes per mirror, ~ 45 min wall-clock on each mirror, in parallel across the four mirrors → ~ 45 min total wall-clock.

Cold cache (first PR after a Renovate bump) costs ≈ 4× warm-cache. The GOCACHEPROG remote build cache (per [S01 §4.6.2](../06_Submodules/01_Submodule_Catalog.md#462-gocacheprog-remote-build-cache)) reduces cold-cache penalty to ~ 2× by serving the dependency closure from a shared cache.

### 14a.2 Nightly cadence cost

Full matrix without per-PR selection. 1,160 cells. Per-mirror total: ≈ 2.5 h with full parallelism. Across 4 mirrors in parallel: ≈ 2.5 h wall-clock. Compute cost: ≈ 14 CPU-min × 29 submodules × 4 mirrors ≈ 1,624 CPU-min ≈ 27 CPU-hours.

### 14a.3 Canary cadence cost

Nightly + 24-h Stress + full 14-topology Challenges fan-out. Stress is the dominant cost (1 × 24-h run per submodule × 29 submodules = 696 GPU-hours sequentially — but the 24-h soak is per *session*, not per *submodule*, so the actual cost is 1 × 24-h × 4 mirrors = 96 GPU-hours). Full Challenges fan-out adds ≈ 2 h. **Canary total: ≈ 28 h wall-clock per mirror; ~ 28 h across 4 mirrors in parallel.**

### 14a.4 Pre-release cadence cost

Canary + 30-day exhaustive Challenges replay. The 30-day replay set is the dominant cost: every Challenges run from the past 30 days (≈ 200 runs per submodule × 29 submodules = 5,800 runs) replayed against the new tag candidate. Each replay is ≈ 5 min, so the replay set is ≈ 480 h sequentially. With 4-mirror × 8-lane parallelism, ≈ 15 h wall-clock + the canary's 28 h = **~ 43 h wall-clock pre-release**. This matches the [S04 §5.4 24–48 h budget](../06_Submodules/04_HelixQA_Integration.md#54-pre-release-cadence).

---

## 14b. CI Workflow Anatomy

The canonical workflow YAML lives in `vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml` and is consumed by every per-submodule CI lane. Excerpt of the workflow shape (full file in the Containers repo):

```yaml
# vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml
on:
  pull_request:
    types: [opened, synchronize, reopened]
  schedule:
    - cron: '0 23 * * *'   # 02:00 Europe/Moscow nightly

jobs:
  test-matrix:
    name: Ten-test-type matrix
    runs-on: ubuntu-22.04
    strategy:
      fail-fast: false      # Per §6.9 — surface all failures
      matrix:
        platform: [linux/amd64, linux/arm64]
        test-type:
          - unit
          - integration
          - e2e
          - security
          - benchmarking-smoke
          - chaos
          - smoke
          - full-automation
          - challenges-primary
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: go.sum
      - name: Cache build artefacts
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      - name: helix-toolchain-pin-audit
        run: ./vasic-digital/Containers/scripts/toolchain-pin-audit.sh
      - name: Run test row
        run: ./vasic-digital/Containers/scripts/test-row.sh ${{ matrix.test-type }} ${{ matrix.platform }}
      - name: Upload coverage / change-point report
        uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.test-type }}-${{ matrix.platform }}-report
          path: tests/${{ matrix.test-type }}/report.json
      - name: Upload to HelixQA run-archive
        if: github.event_name == 'schedule' || github.event_name == 'push'
        run: ./vasic-digital/Containers/scripts/helixqa-upload.sh
```

The required-checks gate at GitHub Settings → Branches → main matches each `test-type` matrix entry; a missing or red entry blocks the merge.

---

## 14c. Test Fixture Conventions

Every submodule's tests follow the conventions below. The conventions are **enforced** by `helix-test-conventions-lint`, a Go vet plugin shipped under `vasic-digital/.github/`.

### 14c.1 Naming

- Unit test files: `*_test.go` co-located with the source (Go convention).
- Integration test files: `*_integration_test.go` under `tests/integration/` with `//go:build integration`.
- E2E test files: `*_e2e_test.go` under `tests/e2e/` with `//go:build e2e`.
- Security: `tests/security/<scanner>.json` (config) + `tests/security/<test>_test.go`.
- Benchmarking: `Benchmark*` Go functions, run via `go test -bench=.`
- Chaos: `tests/chaos/<failure-mode>_test.go` with `//go:build chaos`.
- Stress: `tests/stress/<workload>_test.go` with `//go:build stress`.
- Smoke: `tests/smoke/probe.sh` (shell, not Go).
- Full Automation: `tests/full-automation/orchestrator_test.go`.
- Challenges: delegated to `vasic-digital/Challenges` per S03; the per-submodule `tests/challenges/delegated.md` records the delegation contract.

### 14c.2 Table-driven tests

Go's table-driven idiom is preferred for Unit + Integration tests:

```go
func TestEncrypt(t *testing.T) {
    cases := []struct {
        name    string
        plain   []byte
        wantErr error
    }{
        {"empty", nil, ErrEmptyInput},
        {"valid", []byte("hello"), nil},
        // ...
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            _, err := Encrypt(tc.plain)
            if !errors.Is(err, tc.wantErr) {
                t.Fatalf("got %v, want %v", err, tc.wantErr)
            }
        })
    }
}
```

### 14c.3 Golden-file approach

Where the output is large (e.g. NAL byte streams, JSON payloads), tests compare against **golden files** committed under `tests/<type>/testdata/`. The lint `helix-golden-file-discipline` verifies:

- Every `testdata/*.golden` file has a corresponding test that loads it.
- A `-update` flag regenerates goldens (so contributors don't hand-edit).
- Goldens are versioned with the test code; golden drift across PRs is flagged for explicit review.

### 14c.4 Test fixtures vs production fixtures

Test data lives under `tests/<type>/testdata/`. Production reference data (e.g. canonical configs, default themes) lives under `internal/testdata-shared/`. The split prevents test fixtures from leaking into the production binary (per the §14 audit).

---

## 14d. Test Artifact Retention

| Artifact                              | Retention         | Storage                                                   |
|---------------------------------------|-------------------|-----------------------------------------------------------|
| Per-PR coverage report                | 90 days           | GitHub / GitLab Actions artifact storage                  |
| Per-PR change-point report            | 90 days           | Same as coverage                                          |
| Nightly run-archive entry             | forever           | HelixQA run-archive (S04 §7) with hot/warm/cold tiering   |
| Canary run-archive                    | forever           | Same as nightly                                            |
| Pre-release run-archive               | forever           | Same; pre-release entries flagged for compliance retention |
| Stress soak metric streams            | 1 year            | Object storage; rotated to cold tier at 30 days           |
| Test fixtures (testdata/)             | versioned with code | Git history                                              |

The "forever" retention on nightly + canary + pre-release is a Constitution §16 *Acceptance* requirement — operator audits reference run-archive entries as evidence, and immutability is non-negotiable.

---

## 15. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T01-A            | helix-test-matrix-lint implementation — Go vet plugin or stand-alone tool?                                    | T10 Full Automation chapter                         |
| OQ-T01-B            | Coverage exemption review cadence — quarterly or per-release?                                                 | `08_Operations/04_Observability_and_Events.md`     |
| OQ-T01-C            | Chaos test orchestrator preference — chaos-mesh (k8s) vs Toxiproxy (process-only)?                            | T07 Chaos chapter                                  |
| OQ-T01-D            | Pre-release 30-day Challenges replay budget — operator can reduce for emergency hotfixes?                    | `09_Implementation_Phases/Phase_12_Beta_Launch.md` |

---

## 16. References & Anti-Bluff Verification

### 16.1 Internal

- [`00_Index.md`](00_Index.md) — family index.
- [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §3, §4, §5, §7.
- [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3, §5, §8.
- [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §2, §4, §6, §7.
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §5, §6.3, §8, §9.
- [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row T01.
- [`../01_Constitution.md`](../01_Constitution.md) §1, §6, §7, §11.5.
- [`../02_System_Overview.md`](../02_System_Overview.md) §3 (Reference User Journey).

### 16.2 External (web)

- Go testing package: https://pkg.go.dev/testing (accessed 2026-04-30).
- govulncheck: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck (accessed 2026-04-30).
- Snyk Open Source for Go: https://docs.snyk.io/scan-with-snyk/snyk-open-source/snyk-open-source-supported-languages-and-package-managers/snyk-open-source-for-go (accessed 2026-04-30).
- Toxiproxy: https://github.com/Shopify/toxiproxy (accessed 2026-04-30).
- chaos-mesh: https://chaos-mesh.org/ (accessed 2026-04-30).
- HDR histogram: http://hdrhistogram.org/ (accessed 2026-04-30).
- benchstat: https://pkg.go.dev/golang.org/x/perf/cmd/benchstat (accessed 2026-04-30).

### 16.3 Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`00_Index.md`](00_Index.md)                                       |    177 | 2026-04-30 | family index                                    |
| [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) | 1,218 | 2026-04-30 | catalog rows + delegation exceptions  |
| [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) |   627 | 2026-04-30 | containers + R-18 hazards               |
| [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) |   517 | 2026-04-30 | Challenges contract                     |
| [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) |   584 | 2026-04-30 | cadence + deployment gate              |
| 29× per-submodule descriptors                                                                                                                          |  9,245 | 2026-04-30 | §5 cell ownership                              |
| [`../01_Constitution.md`](../01_Constitution.md) §1 §6 §7 §11.5     |    879 | 2026-04-30 | R-02/R-11/R-12/R-13/R-14/R-18                  |

- Coverage: chapter exceeds the 600-line floor.
- Forbidden patterns: clean. The four §15 open questions are explicitly named with deferred resolution chapters.
- Cell-count audit: 29 × 10 = 290 cells; 288 `O` + 2 `D-*` (helix-shm.Challenges → helix-pipeline; helix-bench.Benchmarking → workload). Zero unowned cells.
- R-12 mock-policy three-layer enforcement (§7.1–§7.3) operationalises the rule rather than relying on prose.
- R-13 anti-bluff: §8 enumerates each test type's known failure mode and identifies Challenges as the backstop with explicit failure-mode mitigations from S03 §7.
- R-18 inheritance: §11 chains every test invocation through the five-layer enforcement.

### 16.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 as a single inline `Write` call (Master Plan §5.3 inline rule for the Testing family).
- Reviewed by: pending operator review.

End of `07_Testing/01_Test_Matrix.md` — 2026-04-30.
