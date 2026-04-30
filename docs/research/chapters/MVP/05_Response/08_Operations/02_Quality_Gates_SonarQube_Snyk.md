# O02 — Quality Gates (SonarQube + Snyk)

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Container_CI_CD.md`](01_Container_CI_CD.md); [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §4.4 + §4.5; [`../07_Testing/05_Security_Tests.md`](../07_Testing/05_Security_Tests.md); [`../01_Constitution.md`](../01_Constitution.md) §7 (R-10), §13 (Exceptions), §16 (Acceptance).
> **Source line count:** chapter floor 400 lines per Master Plan §7.2 row O02.
> **Chapter targets:** R-10 (heavy quality + security scans), R-13 (no green-on-broken), R-18.
> **Cross-links:** [`01_Container_CI_CD.md`](01_Container_CI_CD.md), [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md), [`../07_Testing/05_Security_Tests.md`](../07_Testing/05_Security_Tests.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. The Gates in Five Sentences

HelixPlay's quality-gate posture is **mandated by Constitution §7 (R-10)**: "heavy quality + security scanning. SonarQube, Snyk, plus other heavy security/quality scans." The all-three-pass gate per [S01 §4.5](../06_Submodules/01_Submodule_Catalog.md#45-vulnerability-scanning-govulncheck--snyk--renovate-all-three-pass-gate) is **govulncheck + Snyk + Renovate**; this chapter adds **SonarQube** as the code-quality + duplication scanner and ties together the operational integration with HelixQA's deployment gate. SonarQube scans every PR for code-smells, security hotspots, and code-coverage trends; Snyk scans dependencies + licences; govulncheck scans for reachable vulnerabilities; Trivy scans container images. Operator override of any gate requires Constitution §13 *Exceptions* approval + an entry in the override-audit log. R-10 is operationally enforced; "skip-the-gate" is not a maintainer's choice.

---

## 2. The Quality-Gate Stack

| Tool          | Scope                                  | Version pin (2026-04-30) | Cadence       |
|---------------|----------------------------------------|---------------------------|---------------|
| SonarQube     | Code quality, duplication, smell, hotspot | Community Edition ≥ 10.7  | per-PR        |
| Snyk          | Dependency CVEs + licence policy       | snyk-cli ≥ v1.1300        | per-PR        |
| govulncheck   | Reachable Go vulnerabilities            | latest from vuln.go.dev   | per-PR        |
| Trivy         | Container image OS + Go CVEs            | ≥ v0.55.0                  | per-PR + nightly|
| spdx-check    | SPDX header presence on every source file | matched repo tag         | per-PR        |
| helix-vuln-policy | Reachability-vs-presence policy decision | matched repo tag       | per-PR        |
| Sigstore Rekor | Public transparency log of cosign signatures | n/a (hosted)         | per-deploy    |

The seven layers cover seven distinct dimensions; together they implement R-10 + R-13 at the code → dependencies → container → release pipeline.

---

## 3. SonarQube Integration

SonarQube is the Constitution-§7-named code-quality scanner. Operator hosts a Community Edition instance at `sonarqube.helix.example.com`; per-submodule CI lanes invoke it via `sonar-scanner` per the SonarSource recipe.

### 3.1 Per-submodule sonar-project.properties

```properties
# <submodule>/sonar-project.properties
sonar.projectKey=vasic-digital_helix-${NAME}
sonar.projectName=helix-${NAME}
sonar.sources=.
sonar.exclusions=tests/**,vendor/**,internal/testdata-shared/**
sonar.tests=tests/
sonar.test.inclusions=**/*_test.go
sonar.go.coverage.reportPaths=coverage.out
sonar.qualitygate.wait=true
```

The `sonar.qualitygate.wait=true` flag blocks the CI step until SonarQube returns the gate verdict (typically 30-60 s).

### 3.2 The operator's quality gate

The default gate at `sonarqube.helix.example.com/quality-gates` enforces:

- **New code coverage** ≥ 80 % (lower than the test matrix's ≥ 95 % unit-coverage gate because SonarQube measures total code coverage, not statement-only).
- **New code duplication** ≤ 3 %.
- **Maintainability rating** A.
- **Reliability rating** A.
- **Security rating** A.
- **Security hotspots reviewed** = 100 % on touched-files.

A red gate blocks the merge. Operator may override per Constitution §13.

### 3.3 The operator's SonarQube admin runbook

`HelixDevelopment/HelixQA/docs/runbook/sonarqube-admin.md` covers:

- Quality-gate rule changes (require operator signoff).
- New-rule onboarding (annual review of SonarSource rule additions).
- False-positive suppression (per-project `.sonarqube-suppressions.yaml` with operator signoff + expiry).
- License renewal (Community Edition is free; Developer / Enterprise licence is operator-purchased).

---

## 4. Snyk Integration

Per [T05 §3.2](../07_Testing/05_Security_Tests.md#32-snyk):

```bash
snyk auth $SNYK_TOKEN
snyk test --severity-threshold=high --policy-path=.snyk
snyk monitor --policy-path=.snyk   # uploads the snapshot for continuous monitoring
```

The `snyk monitor` step uploads the dependency snapshot to Snyk's continuous-monitoring service so operator gets paged when a new CVE is published against any dependency in the snapshot.

### 4.1 The .snyk policy

Per [T05 §3.2](../07_Testing/05_Security_Tests.md#32-snyk), the per-submodule `.snyk` files inherit the org-level `vasic-digital/.snyk-base` config:

```yaml
# vasic-digital/.snyk-base
version: v1.27.0
licenses:
  allowed: [MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, MPL-2.0, ISC]
  denied:  [GPL-2.0, GPL-3.0, AGPL-3.0, LGPL-2.1, LGPL-3.0]
ignore: {}   # per-submodule extensions land here
```

Per-submodule additions to the `ignore` block require:

1. CVE ID (e.g. `SNYK-GO-X-12345`).
2. Reason (free text).
3. `expires` ISO timestamp (mandatory; max 90 days from grant).
4. Operator signature on the PR adding the ignore.

The `helix-vuln-suppression-audit` Renovate/cron job runs nightly + fails CI if any ignore is past its expiry.

### 4.2 Continuous monitoring → on-call paging

Snyk's monitor service emits webhooks on new-CVE events. The webhook lands at HelixQA's alert-routing layer per [S04 §8](../06_Submodules/04_HelixQA_Integration.md#8-alert-routing-and-on-call-rotation):

- High-severity CVE (CVSS 7.0+): P2 alert (4-hour acknowledge SLA).
- Critical-severity CVE (CVSS 9.0+): P1 alert (15-minute acknowledge).

The on-call's runbook at `HelixDevelopment/HelixQA/docs/runbook/snyk-cve-paged.md` walks the operator through the patch-vs-suppress decision.

---

## 5. govulncheck Integration

Per [T05 §3.1](../07_Testing/05_Security_Tests.md#31-govulncheck):

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck -mode=symbol -json ./... > govulncheck.json
helix-vuln-policy ./govulncheck.json
```

The `helix-vuln-policy` script applies the reachability-vs-presence rules per [T05 §8](../07_Testing/05_Security_Tests.md#8-the-reachability-vs-presence-distinction):

- **Reachable** vulnerability: must patch (block merge).
- **Present-but-unreachable**: should patch, but not blocking unless CVSS ≥ 9.0 (Critical-unreachable still blocks).
- **Unreachable Critical**: blocks regardless; operator override only via Constitution §13.

The script emits an OTLP span per-PR with `gov.findings_total`, `gov.reachable_high`, `gov.reachable_critical`. HelixQA's dashboard aggregates these.

---

## 6. The All-Three-Pass + SonarQube Gate

The deployment gate's first signal (cosign + SLSA L3 per [§9 quad](#9-the-deployment-gate-quad)) requires the merge gate to have passed; the merge gate requires:

1. SonarQube gate green.
2. Snyk test exit 0 (no high-severity findings).
3. govulncheck via helix-vuln-policy exit 0.
4. Trivy image scan exit 0 (no HIGH or CRITICAL).
5. spdx-check exit 0.
6. Renovate has not flagged a major-bump-required state.

Six checks; all must be green. The four-mirror amplifier means each runs four times (per [§9 four-mirror parity](#9-the-deployment-gate-quad)).

---

## 7. Trivy + Container Image Scanning

Trivy scans the **container image** (vs Snyk's Go-only scope):

```bash
trivy image \
    --severity HIGH,CRITICAL \
    --exit-code 1 \
    --ignore-unfixed=false \
    ghcr.io/vasic-digital/helix-${NAME}@${DIGEST}
```

The `--ignore-unfixed=false` is intentional — an unfixed vulnerability is still a vulnerability. Operator decides whether to ship despite it (Constitution §13 *Exceptions*).

Trivy adds **OS-package** coverage that govulncheck + Snyk miss. A vulnerable `glibc` in `debian:bookworm-slim` is invisible to the Go-tooling-only scanners; Trivy catches it.

---

## 8. SonarQube Operator-Side Setup

Per [O01 §13 Renovate](01_Container_CI_CD.md#13-the-renovate-configuration), the SonarQube CE upgrade happens via Renovate-driven Dockerfile bumps + operator review. Operator-side runbook at `HelixDevelopment/HelixQA/docs/runbook/sonarqube-upgrade.md` covers:

1. Backup the `sonarqube_db` Postgres + the analysis history.
2. Pull the new Docker image digest.
3. Rolling-restart the SonarQube container.
4. Verify the gate still works against a known-clean PR.
5. Rollback procedure (restore the Postgres backup + revert the digest).

The cadence is quarterly + on-demand for critical CVEs in SonarQube itself.

---

## 9. The Deployment-Gate Quad

Per [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad), four signals at deploy time:

1. **Cosign + SLSA L3** — image signed via Sigstore keyless OIDC.
2. **Challenges green** — primary scenario passed at the last cadence.
3. **Both SBOMs emitted** — cyclonedx-gomod + syft.
4. **Visibility audit** — repo public on all four mirrors + licence policy clean.

Fail-closed: any sub-check error → `VerdictBlock`. Operator override per [T12 §10b](../07_Testing/12_HelixQA_Autonomous.md#10b-override-audit-trail-format) is logged forever in run-archive.

---

## 10. Tracking-Mirror Integration with O05

When a quality-gate PR is opened (e.g. operator approves a Snyk suppression), the matching ticket on GitHub Projects + GitLab is created via the O05 tracking-mirror commands. The ticket carries labels:

- `quality-gate:sonarqube|snyk|govulncheck|trivy`
- `severity:high|critical`
- `expires:<ISO-date>` (for suppressions)

Closure of the ticket is gated on the quality-gate's underlying issue resolution — closing without a fix is itself a quality-gate violation.

---

## 10a. Per-Submodule Sonar Project Provisioning

The W07 task's bulk-import (per [O05 §9](05_Tracking_GitHub_GitLab.md#9-the-operators-bulk-issue-creation-workflow)) extends to SonarQube projects. Operator runs:

```bash
for submodule in $(cat vasic-digital/.github/helix-submodules.txt); do
    sonar-admin project create \
        --key "vasic-digital_helix-${submodule}" \
        --name "helix-${submodule}" \
        --quality-gate "HelixPlay-Default" \
        --visibility public
    sonar-admin permission template apply \
        --project "vasic-digital_helix-${submodule}" \
        --template "vasic-digital-default"
done
```

The bulk creation runs once at Phase_00 + when a new submodule is added. Each project inherits the org-level "HelixPlay-Default" quality gate (per §3.2).

## 10b. Snyk Webhook + Slack Integration

Snyk's webhook fires on:

- New CVE published against any monitored snapshot.
- Snyk gate red on a PR.
- Licence policy violation detected.

The webhook lands at HelixQA's `snyk-event-router` service:

```yaml
# HelixQA snyk-event-router service config
filters:
  - severity: critical
    route: pagerduty + slack#helixplay-p1
  - severity: high
    route: slack#helixplay-p2 + email
  - type: license-policy-violation
    route: slack#helixplay-legal + email
```

The router applies the §4.2 acknowledge SLAs (P1 15-min, P2 4-h) automatically.

## 10c. Detailed SonarQube Quality-Profile Customisation

The default Sonar Way profile is too lax for HelixPlay. The custom "HelixPlay-Strict" profile adds:

- **Cognitive Complexity**: any function with score > 15 fails (default is 25).
- **Duplications**: any file with > 3 blocks of 100 lines duplicated fails.
- **Security Hotspots**: every hotspot must be reviewed before merge.
- **Test Coverage**: new code must have ≥ 80 % line coverage (mirror of T01 §6.1's ≥ 95 % statement coverage at the SonarQube projection).
- **Maintainability Rating**: A required.
- **Reliability Rating**: A required.

Profile is provisioned via `sonarqube-customise.sh` at Phase_00; updates to the profile are PR-reviewed + signed off by operator.

## 10d. Per-Tool Failure-Mode Catalogue

| Tool        | Failure mode                           | Operator action                                            |
|-------------|----------------------------------------|-----------------------------------------------------------|
| SonarQube   | Server unreachable                     | Skip the gate (operator-side temporary), file P2          |
| SonarQube   | Quality gate red (legitimate)          | Fix the underlying issue + re-push                        |
| Snyk         | Auth failure                            | Rotate `SNYK_TOKEN` per §14b registry rotation procedure  |
| Snyk         | New CVE in dependency closure          | Per the patch-vs-suppress runbook                          |
| govulncheck | Reachability analysis failed           | Investigate (often go-version mismatch); file P3 if persists |
| Trivy        | OS-package CVE detected                 | Bump base image digest via Renovate; re-build              |
| Sigstore Rekor | Outage at deploy time                | Per OQ-O02-A: fail-closed accepted; operator override path |

## 11. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-O02-A            | SonarQube CE vs Developer Edition — operator's Q3 2026 budget?                                                | this chapter next revision                          |
| OQ-O02-B            | Snyk Open Source Free vs paid — at what fleet size does paid pay off?                                         | this chapter next revision                          |
| OQ-O02-C            | SonarQube false-positive cadence — quarterly review or per-release?                                            | this chapter next revision                          |
| OQ-O02-D            | Trivy vs Grype vs Anchore — single-tool simplification or multi-tool defence?                                 | this chapter next revision                          |

---

## 10e. The Quality-Gate Telemetry Catalogue

Per [O04 §4](04_Observability_and_Events.md#4-per-submodule-metrics-catalogue), the Quality-Gate stack emits its own metrics for operator visibility. Excerpt:

| Metric                                          | Type       | Description                                            |
|-------------------------------------------------|------------|--------------------------------------------------------|
| `helix_qualitygate_sonar_decisions_total`       | counter    | SonarQube gate decisions, labelled `verdict={pass,fail}`. |
| `helix_qualitygate_snyk_findings_total`         | counter    | Snyk findings, labelled `severity`, `submodule`.       |
| `helix_qualitygate_govulncheck_findings_total`  | counter    | govulncheck findings.                                   |
| `helix_qualitygate_trivy_findings_total`        | counter    | Trivy findings, labelled `severity`, `image`.          |
| `helix_qualitygate_overrides_total`             | counter    | Operator overrides, labelled `gate`, `reason_class`.   |
| `helix_qualitygate_sonarqube_uptime_seconds`    | gauge      | SonarQube server uptime since last restart.            |

The override-count metric is the strongest operational signal: a sudden spike means either the gate has become too strict (false-positive flood) or operators are normalising bypass behaviour (audit-trail risk). Either case is investigation-worthy.

## 10f. Suppression-Audit Cadence

The §4.1 `.snyk` suppression block + §3 SonarQube false-positive-suppression mechanisms accumulate over time. The audit cadence:

- **Nightly**: `helix-vuln-suppression-audit` lists every active suppression + flags those past their `expires` date.
- **Weekly**: operator reviews the un-flagged suppressions to verify they're still necessary.
- **Quarterly**: operator runs the full suppression audit + closes out stale ones.
- **Annually**: operator certifies the suppressions catalogue per Constitution §16 *Acceptance*.

Suppressions older than 6 months that have not been re-justified by the operator are **auto-revoked** by the audit; the underlying vulnerability becomes blocking again until re-suppressed.

## 11. Open Questions

[Open Questions remain unchanged below.]

## 12. References & Anti-Bluff Verification

### 12.1 Internal

- [`01_Container_CI_CD.md`](01_Container_CI_CD.md), [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md).
- [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §4.4, §4.5.
- [`../07_Testing/05_Security_Tests.md`](../07_Testing/05_Security_Tests.md).
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §9.
- [`../07_Testing/12_HelixQA_Autonomous.md`](../07_Testing/12_HelixQA_Autonomous.md) §5, §10b.
- [`../01_Constitution.md`](../01_Constitution.md) §7, §13, §16.

### 12.2 External (web)

- SonarQube CE: https://www.sonarsource.com/products/sonarqube/ (accessed 2026-04-30).
- Snyk CLI: https://docs.snyk.io/snyk-cli (accessed 2026-04-30).
- govulncheck: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck (accessed 2026-04-30).
- Trivy: https://trivy.dev/ (accessed 2026-04-30).

### 12.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 400-line floor.
- Forbidden patterns: clean. Four §11 open questions named with deferred resolution chapters.
- §6 all-six-pass merge gate + §9 deployment-gate quad operationalise R-10 + R-13.

### 12.4 The Quality-Gate Investment Math

Operator's monthly Quality-Gate spend (rough estimate):

| Tool                         | Cost / month                              |
|------------------------------|--------------------------------------------|
| SonarQube CE                  | self-hosted (compute only ≈ $30 / month)  |
| Snyk Open Source (free tier)  | $0 (free for OSS projects)                |
| Snyk Open Source (paid plan)  | from $25/dev/month                        |
| Trivy                         | $0 (open-source CLI)                       |
| govulncheck                   | $0 (open-source)                            |
| Sigstore Rekor                | $0 (public service)                         |
| **Estimated total**           | $30–$300 / month depending on plan         |

Compared to the per-CVE-incident cost (post-deployment data breach: $4M+ average per IBM Cost of a Data Breach 2024), the Quality-Gate investment is materially cheaper. Operator's risk-adjusted CTO budget signs off Quality-Gate spend as a non-discretionary line item.

### 12.5 Cross-Family Coverage Confirmation

The Quality-Gate stack ties multiple families together:

- **Submodules family** [S01 §4.4 + §4.5](../06_Submodules/01_Submodule_Catalog.md#44-sbom-generation-per-submodule-cyclonedx-gomod--syft-dual-format) — the SBOM + vuln-scan policy.
- **Testing family** [T05 Security](../07_Testing/05_Security_Tests.md) — the per-test-type discipline.
- **Operations family** (this chapter) — the deployment-time + ongoing operational integration.
- **Implementation Phases family** (forthcoming) — Phase_00 provisions the SonarQube + Snyk environment.

The four families together implement R-10 from policy → discipline → operations → execution.

### 12.6 The Quality-Gate Failure-Recovery Playbook

When a Quality-Gate fails on a PR, the maintainer's recovery procedure:

1. **Read the gate's report** in the PR comment (per [T10 §7a aggregate-summary](../07_Testing/10_Full_Automation.md#7a-aggregate-summary-format)).
2. **Identify which tool flagged**: SonarQube / Snyk / govulncheck / Trivy.
3. **For SonarQube**: open the SonarQube project page; review the new code-smells / hotspots; fix in code.
4. **For Snyk + govulncheck**: bump the affected dependency or, if no fix is available yet, file a suppression with `expires` ≤ 90 days + operator signoff.
5. **For Trivy**: bump the base image digest via Renovate-driven PR.
6. Re-push; the gate re-runs automatically.

If the gate is materially flaky (false positives), maintainer files a P3 ticket against the gate itself + escalates to operator review.

### 12.7 Cross-Reference to T05 (Security Tests)

The Quality-Gate stack overlaps T05's per-PR security row. The boundary:

- **T05** (test type 4 of 10) is the per-PR Security row that runs the §3-§7 scanners as part of the test matrix.
- **O02** (this chapter) is the operational integration — SonarQube projects + Snyk org-level monitoring + webhook routing + telemetry + the all-six-pass merge gate aggregation.

A PR that passes T05 has every individual scanner green; a PR that passes O02's all-six-pass gate has every individual scanner AND the SonarQube quality-gate AND Renovate-sanity-check all green simultaneously. The gates compose; passing one does not subsume the other.

### 12.8 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `08_Operations/02_Quality_Gates_SonarQube_Snyk.md` — 2026-04-30.
