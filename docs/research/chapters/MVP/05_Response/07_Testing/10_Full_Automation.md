# T10 — Full Automation

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.9 + §14b (CI workflow YAML); [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3.4 (host-integrity-scan inheritance).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row T10.
> **Chapter targets:** R-11, R-12, R-13.
> **Cross-links:** [`02_Unit_Tests.md`](02_Unit_Tests.md) through [`09_Smoke.md`](09_Smoke.md), [`12_HelixQA_Autonomous.md`](12_HelixQA_Autonomous.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Full Automation is **test-type 9 of 10**. Cadence: every cadence (per-PR + nightly + canary + pre-release). Its role is **orchestration** — it does not introduce new test logic; it invokes test types 1–8 in CI matrix order, with `fail-fast: false` so every failure is surfaced rather than the first-failure short-circuit hiding others.

Full Automation's purpose: **maximise issue-count surfaced per PR**. A maintainer who fixes a Unit failure and resubmits should not discover an Integration failure, fix it, then a Security failure, fix it. The Full Automation row exposes all of them in one CI cycle.

---

## 2. The Discipline — What Makes "Green" Green

A green Full Automation row means:

1. **All of Unit + Integration + E2E + Security + Benchmarking-smoke + Chaos + Smoke** are green simultaneously.
2. **No row was skipped** — `fail-fast: false` is mandatory in the workflow YAML; the lint rejects any matrix that uses `fail-fast: true`.
3. **Every row's report is uploaded** as a CI artefact (per [T01 §14d](01_Test_Matrix.md#14d-test-artifact-retention) retention policy).
4. **The aggregate report links back to every row** — operator can click into any failed row from the workflow's summary page.

If any of T02–T09 is red, Full Automation is red. If all are green, Full Automation is green. This is composition, not invention.

---

## 3. Tooling

- **GitHub Actions matrix** with `fail-fast: false` — primary orchestrator.
- **GitLab CI parent-child pipelines** — the GitLab equivalent.
- **GitFlic CI workflow** + **GitVerse CI workflow** — the four-mirror parallel runs.
- **`vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml`** — the canonical workflow YAML template (excerpt at [T01 §14b](01_Test_Matrix.md#14b-ci-workflow-anatomy)).

The four-mirror amplifier means the Full Automation row runs *four times in parallel*, once per mirror. Per [T01 §9](01_Test_Matrix.md#9-four-mirror-parity-contract), divergent verdicts trigger alerts.

---

## 4. The Canonical Matrix YAML

The full workflow under `vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml`. Excerpt covering the T10-orchestration shape:

```yaml
on:
  pull_request:
    types: [opened, synchronize, reopened]
  schedule:
    - cron: '0 23 * * *'   # 02:00 Europe/Moscow nightly

jobs:
  full-automation:
    runs-on: ubuntu-22.04
    strategy:
      fail-fast: false      # Mandatory per §2.
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
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go + cache
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: go.sum
      - name: helix-toolchain-pin-audit
        run: ./vasic-digital/Containers/scripts/toolchain-pin-audit.sh
      - name: Run test row
        run: ./vasic-digital/Containers/scripts/test-row.sh ${{ matrix.test-type }} ${{ matrix.platform }}
      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.test-type }}-${{ matrix.platform }}-report
          path: tests/${{ matrix.test-type }}/report.json
      - name: Upload to HelixQA run-archive
        if: always() && (github.event_name == 'schedule' || github.event_name == 'push')
        run: ./vasic-digital/Containers/scripts/helixqa-upload.sh

  aggregate-summary:
    needs: full-automation
    if: always()
    runs-on: ubuntu-22.04
    steps:
      - name: Download all reports
        uses: actions/download-artifact@v4
      - name: Build summary
        run: |
          ./vasic-digital/Containers/scripts/full-automation-summary.sh > summary.md
      - name: Comment on PR with summary
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const summary = fs.readFileSync('summary.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: summary
            });
```

The `aggregate-summary` job runs `if: always()` so the PR comment posts even when one or more rows fail; the maintainer sees the full picture.

---

## 5. Required-Checks Configuration

Per [S04 §5.1](../06_Submodules/04_HelixQA_Integration.md#51-per-pr-cadence), the required-checks list at GitHub Settings → Branches → main matches each `(test-type, platform)` matrix entry. Excerpt:

- `Full Automation / unit (linux/amd64)`
- `Full Automation / unit (linux/arm64)`
- `Full Automation / integration (linux/amd64)`
- `Full Automation / integration (linux/arm64)`
- `Full Automation / e2e (linux/amd64)`
- `Full Automation / e2e (linux/arm64)`
- `Full Automation / security (linux/amd64)`
- `Full Automation / security (linux/arm64)`
- `Full Automation / benchmarking-smoke (linux/amd64)`
- `Full Automation / benchmarking-smoke (linux/arm64)`
- `Full Automation / chaos (linux/amd64)` (nightly+ only)
- `Full Automation / chaos (linux/arm64)` (nightly+ only)
- `Full Automation / smoke (linux/amd64)` (post-deploy only)

A missing required check (e.g. unit failed to start) blocks the merge as if it had reported red. This is the "no skipping" rule operationalised — the absence of a check is a red signal.

The same required-checks list lives at GitLab Settings → Repository → Protected Branches, GitFlic equivalent, and GitVerse manual policy (per [S01 §4.7.4](../06_Submodules/01_Submodule_Catalog.md#474-gitverse) GitVerse-API gap).

---

## 6. Anti-Pattern Catalogue

### 6.1 `fail-fast: true`

```yaml
strategy:
  fail-fast: true   # forbidden
  matrix: ...
```

The first failed cell short-circuits the rest. Maintainer fixes that cell, resubmits, discovers cell #2 failed; repeats. Wastes CI cycles + maintainer attention. **Always `fail-fast: false`**.

### 6.2 Missing `if: always()` on report upload

```yaml
- name: Upload report
  uses: actions/upload-artifact@v4
  # missing: if: always()
```

Without `if: always()`, a failed test step skips the artefact upload. The operator inspecting the failure has no report to look at. Always upload artefacts unconditionally.

### 6.3 Workflow YAML drift between mirrors

GitHub Actions YAML at `.github/workflows/ci.yml`; GitLab equivalent at `.gitlab-ci.yml`; GitFlic at `.gitflic-ci.yml`; GitVerse at `gitverse-ci.yaml`. Drift between them creates parity violations. The lint `helix-workflow-yaml-parity` parses all four + asserts equivalence (modulo platform-specific syntax).

### 6.4 Per-submodule deviations

A submodule that ships a custom workflow YAML (instead of consuming `vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml`) escapes the canonical discipline. The lint `helix-ci-fragment-usage` rejects per-submodule YAML that doesn't include the canonical fragment.

### 6.5 Disabling matrix entries via if

```yaml
- name: Run E2E
  if: github.event_name != 'pull_request'  # disables E2E on every PR
  run: ...
```

This silently disables the row. If the intent is "skip E2E on per-PR cadence", document it via the Cadence cross-tab (T01 §5) + use the matrix's `cadence` axis, not an inline `if`.

---

## 7. The Required-Checks Drift Audit

Required-checks lists drift over time. The audit:

1. Fetch the current required-checks list from GitHub + GitLab + GitFlic + GitVerse.
2. Compare against the canonical list in `vasic-digital/.github/required-checks-canonical.yaml`.
3. Open a PR if drift detected; operator approves restoration.

Runs nightly. Discrepancies are P3 (informational); a check missing from a mirror's list is a P2 (bypass risk).

---

## 7a. Aggregate-Summary Format

The PR comment / GitLab discussion / GitFlic-comment posted by the `aggregate-summary` job uses a fixed Markdown shape so maintainers can scan quickly:

```markdown
## Full Automation Summary — PR #1234

| Test type            | linux/amd64 | linux/arm64 | Reports                                        |
|----------------------|:-----------:|:-----------:|------------------------------------------------|
| Unit                 | ✅ green     | ✅ green     | [report](artifact://unit-amd64), [report](artifact://unit-arm64)|
| Integration          | ✅ green     | ✅ green     | …                                                |
| E2E                  | ❌ red       | ✅ green     | [report](…)                                       |
| Security             | ✅ green     | ✅ green     | …                                                |
| Benchmarking smoke   | ⚠️ regress   | ✅ green     | [report](…) p999=12ms (baseline 8ms; +50 %)     |
| Chaos                | ⏸️ skipped   | ⏸️ skipped   | (per-PR cadence does not run chaos)              |
| Smoke                | ⏸️ skipped   | ⏸️ skipped   | (post-deploy only)                                |

**Verdict**: ❌ red — E2E (linux/amd64) failed; Benchmarking regressed (linux/amd64).

Required-checks gate: 2 of 14 red. Merge blocked.
```

The single-table format is parseable by both maintainers and automated tooling (e.g. a Slack-bot summarising open PRs).

## 7b. Aggregate-Summary Sources

The `full-automation-summary.sh` script aggregates from:

- `tests/<type>/report.json` artefacts uploaded per-row.
- The `helix-bench compare` output for benchmarking-smoke.
- The `helix-leak-regress` output for stress (canary+).
- The `helix-vuln-policy` output for security.

Each source's verdict (✅ / ⚠️ / ❌ / ⏸️) is mapped per the script's verdict table. A row with no report (e.g. the test-row.sh script crashed before producing report.json) is mapped to ❌ red — absence is treated as failure (per [§5 required-checks](#5-required-checks-configuration)).

## 8. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T10-A            | Workflow-YAML drift detection — pre-merge lint vs nightly audit?                                              | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T10-B            | Per-submodule customisation policy — operator-approved exceptions allowed?                                    | Constitution §13 *Exceptions*                       |
| OQ-T10-C            | Aggregate-summary GitHub-comment vs GitLab-discussion vs GitFlic-PR-comment — unify or per-mirror?            | `08_Operations/05_Tracking_GitHub_GitLab.md`       |

---

## 7c. Per-Mirror Workflow YAML Equivalence

Each mirror's CI system has its own YAML dialect:

- **GitHub Actions** — `.github/workflows/ci.yml`, `on: { pull_request, schedule }` syntax.
- **GitLab CI** — `.gitlab-ci.yml`, `rules:` clauses, parent-child via `trigger:`.
- **GitFlic CI** — `.gitflic-ci.yml`, similar shape to GitHub Actions but with GitFlic-specific marketplace actions.
- **GitVerse CI** — `gitverse-ci.yaml`, currently lacks the `matrix:` strategy at parity with GitHub; documented gap per [S01 §4.7.4](../06_Submodules/01_Submodule_Catalog.md#474-gitverse).

The lint `helix-workflow-yaml-parity` parses all four + asserts logical equivalence (modulo dialect). The matrix axes (test-type, platform), the workflow steps' execution semantics, and the required-checks coverage must match across mirrors. Drift triggers a P3 alert; the operator either fixes the drift or amends the canonical fragment to encode the divergence intentionally.

## 7d. The Composite-Step Pattern

To minimise YAML duplication across the four mirror dialects, common steps are encoded as **composite actions** (GitHub) / **`include:`d job templates** (GitLab) / equivalent abstractions on GitFlic + GitVerse. Excerpt of the canonical composite action:

```yaml
# vasic-digital/Containers/.github/actions/run-test-row/action.yml
name: Run Test Row
description: Runs a single test type with cache + report upload + helixqa upload.
inputs:
  test-type:
    required: true
    description: One of unit/integration/e2e/security/...
  platform:
    required: true
runs:
  using: composite
  steps:
    - run: ./vasic-digital/Containers/scripts/test-row.sh ${{ inputs.test-type }} ${{ inputs.platform }}
      shell: bash
    - if: always()
      uses: actions/upload-artifact@v4
      with:
        name: ${{ inputs.test-type }}-${{ inputs.platform }}-report
        path: tests/${{ inputs.test-type }}/report.json
```

Submodules consume the action via `uses: vasic-digital/.github/actions/run-test-row@main`. The action is versioned with the org-level `.github` repo; updates roll through Renovate.

## 8. Open Questions

### 9.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.9, §14b.
- [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3.4.
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §5.1.

### 9.2 External (web)

- GitHub Actions matrix: https://docs.github.com/en/actions/using-jobs/using-a-matrix-for-your-jobs (accessed 2026-04-30).
- GitLab parent-child pipelines: https://docs.gitlab.com/ee/ci/pipelines/parent_child_pipelines.html (accessed 2026-04-30).

### 9.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §8 open questions named with deferred resolution chapters.
- §6 anti-pattern catalogue + §7 drift audit operationalise the "every required check enforced everywhere" invariant.

### 9.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/10_Full_Automation.md` — 2026-04-30.
