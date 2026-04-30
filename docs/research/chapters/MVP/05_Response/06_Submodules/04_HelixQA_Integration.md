# S04 — `HelixDevelopment/HelixQA` Autonomous QA Orchestration

> **Source dimensions:**
> - [`00_Index.md`](00_Index.md) §1 (HelixQA autonomous integration role) + §5 (Ten-test-type matrix orchestration delegated to S04).
> - [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md) §3.1 (29 rows whose CI runs HelixQA orchestrates), §5.5 (container-lane consequence — HelixQA must size for 261 container-bound test invocations per all-touching PR), §9 (release-train cadence — HelixQA gates the v1.0.0 graduation).
> - [`02_Containers_Submodule.md`](02_Containers_Submodule.md) §6 (cosign + SLSA L3 — HelixQA verifies these before deployment), §7 (four-mirror registries — HelixQA orchestrates parity), §10 (cache strategy — HelixQA budgets against the §10 cost model).
> - [`03_Challenges_Submodule.md`](03_Challenges_Submodule.md) §8 (S03 ↔ S04 interface contract), §9 (cadence: per-PR / nightly / canary / pre-release).
> - [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row S04 (≥ 400-line floor).
> - [`../01_Constitution.md`](../01_Constitution.md) §6 (Testing Discipline R-11 + R-12 + R-13), §6.4 (R-14 Challenges + HelixQA integration), §8 (Tracking R-16 + R-17), §11.5 (R-18 Operational Integrity), §16 (Acceptance — operator's non-delegable role).
> - [`../02_System_Overview.md`](../02_System_Overview.md) §15 (Release Trains).
> - The pre-existing organisational repo [`https://github.com/HelixDevelopment/HelixQA.git`](https://github.com/HelixDevelopment/HelixQA.git) — the autonomous orchestrator. Distinct from `vasic-digital/Containers` and `vasic-digital/Challenges` because it lives under the `HelixDevelopment` org (the shipped-product org), not under `vasic-digital` (the reusable-component org).
>
> **Source line count:** family-level inputs ≈ 250 lines (S01 §5 + S03 §8 + Constitution §6 + §8 + §11.5 slices). Master Plan §7.2 sets the chapter floor at **≥ 400 lines**; this chapter overshoots that floor on the strength of the cadence × cross-mirror × deployment-gate matrix.
>
> **Chapter targets:** R-11 (Ten test types orchestrated by HelixQA), R-12 (mocks-only-in-Unit enforced by HelixQA's Challenges invocation), R-13 (anti-bluff guarantee re-enforced at deployment gate), R-14 (Challenges + HelixQA integration as in HelixAgent / Catalogizer), R-16 + R-17 (every HelixQA decision mirrored to GitHub Projects + GitLab tickets), R-18 (HelixQA orchestrators run inside containers per S02; no orchestrator step suspends the operator host).
>
> **Cross-links:** [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md), [`02_Containers_Submodule.md`](02_Containers_Submodule.md), [`03_Challenges_Submodule.md`](03_Challenges_Submodule.md), `per-submodule/<name>.md` (S05 — every descriptor's `HelixQA cadence` column resolves to §4 of this chapter).
>
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## Table of Contents

1. The Role of `HelixDevelopment/HelixQA` in the Synthesis Programme
2. Why HelixQA Lives Under `HelixDevelopment`, Not `vasic-digital`
3. Repository Layout (orchestrator + agent SDK + run archive + alert routing + dashboards)
4. The Autonomous Orchestration Model (continuous, cadence, reactive)
5. Per-Cadence Orchestration: per-PR / nightly / canary / pre-release
6. Multi-Mirror Orchestration Across the Four-Mirror CI Topology
7. Run Archive and Audit Trail (immutable, queryable, signed)
8. Alert Routing and On-Call Rotation
9. Deployment Gating (the cosign + Challenges + SBOM + visibility quad)
10. Integration with the Other Organisational Repos
11. Open Questions
12. References & Anti-Bluff Verification

---

## 1. The Role of `HelixDevelopment/HelixQA` in the Synthesis Programme

`HelixDevelopment/HelixQA` is the **autonomous QA orchestrator** for the 29-submodule fleet. It is **not** a submodule (S01 §3.1 does not list it) and it is **not** a `vasic-digital` shared resource (S02 and S03 cover those). It is a **product-org-scoped** orchestrator that owns:

- Continuous orchestration of the Ten-test-type matrix across all 29 submodules.
- Cadence-driven invocation of `vasic-digital/Challenges` topologies (§5).
- Cross-mirror parity verification (the four CI runner topologies must agree).
- The immutable run archive (§7) — every run produces a signed, queryable record.
- The alert routing fan-out — P1 / P2 / P3 dispatch to on-call humans (§8).
- The deployment gate (§9) — no image is deployed without cosign + Challenges + SBOM + visibility-audit green.

R-14 (Constitution §6.4) names HelixQA explicitly:

> The autonomous QA system `git@github.com:HelixDevelopment/HelixQA.git` is fully integrated. Every submodule's full regression matrix runs through HelixQA on the cadences specified in S04.

S04 is the chapter that operationalises HelixQA's role. The contract S04 establishes is bidirectional:

- **Inbound to HelixQA**: every submodule emits structured CI-run artefacts (test reports, SBOMs, change-point reports from Challenges, cosign signatures, OTLP traces). The artefact format is fixed in §3.
- **Outbound from HelixQA**: deployment-gate verdicts, ticket creates / updates on GitHub Projects + GitLab (per R-17), alert pages, run-archive entries, dashboard updates.

Without HelixQA, the 29-submodule fleet would have **no single point of truth** for "is the system healthy?" The CI per-submodule lanes (S02) tell each submodule's local truth; Challenges (S03) tells each submodule's user-visible-behaviour truth; HelixQA aggregates both into a fleet-wide health verdict that operators consume.

---

## 2. Why HelixQA Lives Under `HelixDevelopment`, Not `vasic-digital`

The organisational separation is deliberate and Constitution-grade:

- **`vasic-digital`** is the org for **reusable components**. R-03 mandates every reusable component live there as a public submodule. The 29 submodules live there; `Containers` and `Challenges` live there as shared toolboxes.
- **`HelixDevelopment`** is the org for **shipped products**. The HelixPlay host agent, the HelixPlay clients, and the HelixQA orchestrator live there because they are not "components someone else's project would reuse"; they are the product itself plus its operational toolchain.

This separation has practical consequences:

1. **Visibility policy.** `vasic-digital` repos are universally public (R-03 §4.7). `HelixDevelopment` repos may be public or private depending on operator choice — the host agent's source might be public, but a deployment-private branch with operator-specific configuration may be private.
2. **Licence policy.** `vasic-digital` repos use MIT (default) + Apache-2.0 (4 named exceptions, S01 §4.8). `HelixDevelopment` repos may use a more restrictive licence for product code (e.g. PolyForm Shield or BUSL-1.1) because product code is not a reusable component and need not bind itself to permissive distribution.
3. **Security posture.** `HelixDevelopment` repos that contain operator-specific deployment configuration are signed with a separate cosign / minisign key from `vasic-digital`. The two key chains are isolated so a compromise of one does not contaminate the other.
4. **Tracking.** GitHub Projects boards for `vasic-digital` track per-submodule maintenance + the catalog. GitHub Projects boards for `HelixDevelopment` track the product's release-train + operator escalations (Constitution §8, R-16 + R-17).

HelixQA spans both orgs at runtime — it consumes `vasic-digital` artefacts (29 submodules' CI runs, `Containers` policies, `Challenges` baselines) and produces `HelixDevelopment` artefacts (deployment verdicts, run archive, on-call alerts) — but it lives in `HelixDevelopment` because it is the product's operational orchestrator, not a reusable component.

---

## 3. Repository Layout (orchestrator + agent SDK + run archive + alert routing + dashboards)

`HelixDevelopment/HelixQA` is structured as follows. The directory layout mirrors S02 §2 + S03 §3's "single narrow responsibility per directory" rule, with the additional axis of "control plane vs data plane".

```
HelixDevelopment/HelixQA/
├── orchestrator/                                  ← control plane
│   ├── cmd/
│   │   ├── helixqa-orchestrator/                  ← main daemon
│   │   ├── helixqa-cli/                           ← operator CLI
│   │   └── helixqa-replay/                        ← replay an archived run from run-archive
│   ├── internal/
│   │   ├── cadence/                               ← per-PR / nightly / canary / pre-release scheduler
│   │   ├── dispatch/                              ← invokes vasic-digital/Challenges scripts + per-submodule CI
│   │   ├── parity/                                ← four-mirror parity checker
│   │   ├── gate/                                  ← deployment gate decision engine (§9)
│   │   ├── archive/                               ← run-archive writer (signed, append-only)
│   │   ├── tracking/                              ← gh + glab integration (R-17)
│   │   └── alert/                                 ← P1/P2/P3 routing (§8)
│   └── go.mod                                     ← imports from vasic-digital where applicable
├── agent-sdk/                                     ← data plane: per-submodule reporting
│   ├── go/                                        ← Go SDK consumed by every submodule's CI
│   │   ├── reporter/                              ← uploads test reports + SBOM + traces
│   │   ├── attestation/                           ← cosign + SLSA bundle uploader
│   │   └── README.md
│   ├── python/                                    ← Python SDK for ad-hoc QA scripts
│   └── shell/                                     ← shell helpers for non-Go lanes
├── run-archive/                                   ← immutable, append-only, signed
│   ├── 2026-04/
│   │   ├── 2026-04-30T00:00:00Z/                  ← timestamped per-run directory
│   │   │   ├── run.manifest.json                  ← SHA-256 manifest of every artefact
│   │   │   ├── run.manifest.json.minisig          ← signature
│   │   │   ├── run.metadata.json                  ← cadence + submodule + topology + scenario
│   │   │   ├── ci-reports/                        ← per-lane CI reports
│   │   │   ├── change-point-reports/              ← Challenges change-point JSONs
│   │   │   ├── otlp-traces/                       ← OpenTelemetry trace records
│   │   │   ├── sbom-bundles/                      ← cyclonedx + spdx
│   │   │   └── cosign-bundles/                    ← signed image attestations
│   │   └── ...
│   └── INDEX.json                                 ← queryable index for time-window queries
├── dashboards/                                    ← Grafana JSON exports + dashboard-as-code
│   ├── fleet-health/                              ← 29-submodule health rollup
│   ├── per-submodule/                             ← per-submodule deep-dive (29 boards)
│   ├── deployment-gate/                           ← real-time deployment-gate decisions
│   ├── change-point-trends/                       ← VMAF / latency / billing trends
│   └── on-call/                                   ← on-call's first-screen rollup
├── alert-routing/                                 ← P1/P2/P3 routing rules
│   ├── policies/
│   │   ├── p1-fleet-down.cue                      ← any topology Challenges red across all 4 mirrors
│   │   ├── p2-mirror-divergence.cue               ← Challenges parity drift between mirrors
│   │   ├── p3-flaky-scenario.cue                  ← scenario red 1/4 mirrors
│   │   └── ...
│   └── README.md
└── docs/
    ├── runbook/                                   ← operator runbooks per alert type
    └── onboarding/                                ← new-on-call onboarding
```

The `orchestrator/` directory is the **control plane** — long-running daemon code that drives cadence, dispatches CI runs, gates deployments, writes the run archive, and pages on-call. The `agent-sdk/` directory is the **data plane** — per-submodule SDK that every CI lane imports to report into HelixQA.

The run-archive is **append-only**: every directory is timestamped, every manifest is signed. A run that turns out to have been wrong (e.g. a flaky scenario that should not have flagged red) is **not deleted**; the new run records the supersession via `run.metadata.json`'s `superseded-by` field. The audit trail is preserved.

`★ Why `cue` for alert policies.` CUE (https://cuelang.org/) is the same policy language used in S02 §8 for container-runtime hazard policies. Reusing it for alert routing means a single mental model across two policy domains — a CUE-fluent operator can read and audit both without context switching. CUE's strong-typing and validation also catch policy-bugs at parse time rather than at fire-time (critical for P1 alerts where a malformed policy would silently fail to page).

---

## 4. The Autonomous Orchestration Model (continuous, cadence, reactive)

HelixQA's orchestration runs in **three concurrent modes**, each with a distinct trigger and scope:

### 4.1 Continuous mode

The orchestrator is always-on. It:

- Watches the four-mirror push streams (GitHub webhooks, GitLab webhooks, GitFlic poll, GitVerse poll) for new commits, new tags, new PR events.
- Maintains a real-time view of "what is the latest commit on every mirror for every of the 29 submodules" and asserts parity.
- Updates the dashboards every 30 seconds with the fleet-health rollup.
- Emits OTLP traces to its own observability stack so HelixQA itself is observable.

Continuous mode is the **substrate**. Cadence and reactive modes (§4.2 + §4.3) ride on top of it.

### 4.2 Cadence mode

Cadence mode runs on schedule:

- **Per-PR cadence** — triggered by every PR open / push to a `vasic-digital` or `HelixDevelopment` repo. Wall-clock: 5–15 minutes.
- **Nightly cadence** — triggered by cron at 02:00 Europe/Moscow. Wall-clock: 2–4 hours.
- **Canary cadence** — triggered by every commit to a release branch. Wall-clock: 12–24 hours.
- **Pre-release cadence** — triggered by every `v1.0.0+` tag candidate. Wall-clock: 24–48 hours.

§5 of this chapter details what each cadence does.

### 4.3 Reactive mode

Reactive mode is event-driven outside the cadence schedule:

- **Operator-triggered re-run** — an operator can re-run a specific run via `helixqa-cli replay <run-id>`. The replay uses the same archived inputs (controller events, network impairment, clock skew) as the original run.
- **Mirror-divergence triggered audit** — when the parity checker detects divergence (one mirror's commit SHA does not match the others), reactive mode triggers a four-mirror audit run of the affected submodule.
- **Baseline-update triggered re-baseline** — when an operator merges a baseline-update PR in `vasic-digital/Challenges`, reactive mode triggers a re-baseline of the affected scenario across all four mirrors.

Reactive mode shares the dispatcher with cadence mode but bypasses the schedule. Its events are recorded in the run-archive with `cadence=reactive` and a sub-type tag (`replay` / `parity-audit` / `re-baseline`).

---

## 5. Per-Cadence Orchestration: per-PR / nightly / canary / pre-release

Each cadence has a **fixed scope** that S04 commits to. Operators can tune within the scope (e.g. raise per-PR scenario count for a high-risk submodule) but cannot reduce below the floor without an explicit Constitution amendment.

### 5.1 Per-PR cadence

**Trigger:** PR open or push event from any of the four mirrors.

**Scope:**

- Runs the per-submodule CI lane (S02 §3) for the affected submodule(s).
- Runs the affected submodule's primary Challenges scenario (S03 §4.1).
- Verifies cosign signatures + SLSA L3 provenance on any newly-built image.
- Verifies SBOM (cyclonedx + spdx) emission per S01 §4.4.
- Verifies licence scan (Snyk + spdx-check) per S01 §4.5.
- Runs `host-integrity-scan` (S02 §3.4) inside the container.
- Updates the per-PR check status on GitHub + GitLab.

**Required-checks gate:** the merge button is disabled until every per-PR check returns green. The required-checks list is configured at the org level on GitHub (Settings → Branches → main → Require status checks to pass) and at the project level on GitLab (Settings → Repository → Protected branches).

**Wall-clock budget:** 5–15 minutes per submodule. With S02 §10's caching, a typical PR touching one submodule completes in 5–8 minutes.

### 5.2 Nightly cadence

**Trigger:** cron at 02:00 Europe/Moscow.

**Scope:**

- Runs **all** §4.1 Challenges scenarios across **all** 29 submodules. The full matrix is 29 × N scenarios (N typically 2–3 per submodule, total ≈ 70 scenarios).
- Runs the four-mirror parity checker (`four-mirror-challenges-parity.sh`) for every scenario.
- Runs the visibility audit (S01 §4.7.5) across the four mirrors.
- Runs the licence audit (S01 §4.8.7) — though that one is monthly per S01, the nightly cadence triggers a sample.
- Updates the run-archive with every scenario's outcome.
- Emails the operator a digest of the night's outcomes.

**Wall-clock budget:** 2–4 hours. Parallelism is the norm — multiple scenarios run concurrently across the four mirrors' CI runner pools.

**Failure handling:**

- A single scenario red on a single mirror raises a P3 alert (flaky).
- A single scenario red on multiple mirrors raises a P2 alert (real regression candidate).
- Multiple scenarios red across mirrors raises a P1 alert (fleet-down candidate).

### 5.3 Canary cadence

**Trigger:** push to a release branch (`release/v0.x` or `release/v1.x`).

**Scope:**

- Runs the **full 14-topology fan-out** in `vasic-digital/Challenges` — every topology × every scenario, not just the per-submodule primary.
- Runs the long-running `11_stress_24h_steady_state` topology for 24 hours.
- Runs the chaos topology `10_chaos_network_partition` with extended fault injection.
- Verifies the four-mirror tag parity for every consumed image.
- Produces a release-candidate health report at the end.

**Wall-clock budget:** 12–24 hours. The 24h stress topology dominates; other scenarios run in parallel inside the 24h window.

**Required-checks gate:** the release branch cannot merge to `main` until every canary check returns green. The operator's `Acceptance` review (Constitution §16) is the human-in-the-loop gate after the automated checks pass.

### 5.4 Pre-release cadence

**Trigger:** tag candidate `v1.0.0+` on any `vasic-digital` submodule (release-train event per S01 §9).

**Scope:**

- Canary cadence + the exhaustive replay set: every Challenges run from the past 30 days replayed against the new tag candidate.
- VMAF / SSIM / latency / billing comparison against historical baselines (not just the canonical baseline).
- Cosign signature verification across the four mirror registries.
- Licence audit + SBOM diff against the prior tag.

**Wall-clock budget:** 24–48 hours.

**Required-checks gate:** the `v1.0.0+` tag is **not pushed** to any mirror until pre-release passes. The release-train script (`vasic-digital/Containers/ci-fragments/release-train.yml`) blocks on this gate.

---

## 6. Multi-Mirror Orchestration Across the Four-Mirror CI Topology

HelixQA orchestrates across **four** independent CI runner topologies — one per mirror. The orchestrator does not assume a single CI runner pool; it treats the four pools as peers and aggregates verdicts.

### 6.1 The four CI runner topologies

| Mirror     | CI runner system    | Runner location          | Connectivity |
|------------|---------------------|--------------------------|--------------|
| GitHub     | GitHub Actions      | GitHub-hosted ubuntu-22.04 | Western      |
| GitLab     | GitLab CI           | GitLab.com shared runners | Western      |
| GitFlic    | GitFlic CI          | GitFlic-hosted runners    | Russian      |
| GitVerse   | GitVerse CI         | GitVerse-hosted runners   | Russian      |

The Russian-jurisdiction runners (GitFlic, GitVerse) cannot reach the Western mirror registries directly during pulls; the four-mirror registry replication (S02 §7) ensures every image exists on every mirror, so the runners always pull from their local mirror.

### 6.2 The parity contract

For every cadence run, HelixQA expects:

- Every of the four mirrors runs the same scenario.
- The scenario produces the same observation (within S03 §6.4 "perturbation" tolerance).
- The change-point detection produces the same verdict.

A divergence indicates one of:

- **Registry replication drift** — one mirror's image digest does not match the others. P2 alert; the four-mirror-replication audit script (S02 §7) handles the remediation.
- **CI runner regression** — one mirror's runner pool has a configuration drift (e.g. kernel version skew). P3 alert; the operator investigates.
- **Real regression** — three mirrors green, one red because the change actually breaks behaviour on that mirror's runner OS. P2 alert; the affected mirror is the bottleneck for merge.

### 6.3 The reverse: a Russian mirror green, Western red

This is a special case. If GitVerse and GitFlic report green, but GitHub and GitLab report red, the orchestrator's default verdict is **red** (Western mirrors are the canonical authority for the project's primary distribution). However, the divergence triggers a P1 alert to the operator because such a split is unusual and potentially indicates a sanctions-related runtime difference (e.g. a Western mirror's runner is on a kernel patched for a CVE that is unpatched on the Russian runner). The operator decides whether to ship the regression-bearing build behind a feature flag pending Western remediation, or to roll back.

### 6.4 The four-mirror cost model

S02 §10's cost budget assumed a single CI runner pool. The four-mirror multiplier:

| Cadence       | Single-mirror wall-clock | Four-mirror wall-clock (parallel) | Compute cost  |
|---------------|--------------------------|------------------------------------|---------------|
| Per-PR        | 5–15 min                 | 5–15 min (parallel; longest mirror wins) | ~4×    |
| Nightly       | 2–4 h                    | 2–4 h (parallel)                   | ~4×           |
| Canary        | 12–24 h                  | 12–24 h (parallel; 24h stress dominates) | ~4×    |
| Pre-release   | 24–48 h                  | 24–48 h (parallel)                 | ~4×           |

The four-mirror cost is not a 4× wall-clock penalty (the runs are parallel) but it is a 4× compute cost. The operator's CI budget accommodates this; the §10 open question OQ-S04-A asks whether canary's 24h stress can be **single-mirror** to halve the canary compute cost (with a P3 alert if the chosen mirror's run differs from a sample-mirror's quick verification).

---

## 7. Run Archive and Audit Trail (immutable, queryable, signed)

Every run produces an entry in `run-archive/`. The entry is **immutable** (append-only), **signed** (minisign over the manifest), and **queryable** (an INDEX.json supports time-window + cadence-type + submodule queries).

### 7.1 Why immutable

Constitution §8 (R-16 + R-17) requires every phase / task / subtask to be tracked on GitHub Projects + GitLab. The tracking system relies on HelixQA's run archive for ground truth. If a run could be edited or deleted, the tracking system's history would be unreliable. Immutability is a Constitution-grade requirement.

### 7.2 The archive format

Every run directory contains:

- `run.manifest.json` — SHA-256 manifest of every artefact in the directory.
- `run.manifest.json.minisig` — minisign signature over the manifest.
- `run.metadata.json` — cadence, submodule, topology, scenario, mirror, timestamp, operator (if reactive), supersedes / superseded-by chain.
- `ci-reports/` — per-lane CI reports.
- `change-point-reports/` — Challenges change-point JSONs (per S03 §6).
- `otlp-traces/` — OpenTelemetry trace records.
- `sbom-bundles/` — cyclonedx + spdx bundles.
- `cosign-bundles/` — signed image attestations.

### 7.3 The INDEX

`run-archive/INDEX.json` is the queryable lookup. Schema:

```json
{
  "$schema": "...",
  "runs": [
    {
      "run_id": "2026-04-30T00:00:00Z",
      "cadence": "nightly",
      "submodule": "helix-pipeline",
      "topology": "01_minimum_viable_session",
      "scenario": "06_full_pipeline_end_to_end",
      "mirror": "github",
      "verdict": "green",
      "duration_seconds": 1203,
      "manifest_sha256": "...",
      "manifest_minisig": "..."
    },
    ...
  ]
}
```

The INDEX is rebuilt on every archive write by a separate background job that scans the directory tree and signs the new INDEX. The background job is idempotent: rebuilding the INDEX from the directory tree always produces the same JSON for the same directory contents.

### 7.4 Retention

The archive is retained **forever**. Constitution §16's Acceptance and §15's Amendment Procedure both reference run-archive entries as evidence; the audit trail is non-deletable. Storage cost is amortised by hot/warm/cold tiering: runs from the past 30 days live on fast SSD; 30 days to 1 year on object storage; 1 year+ on cold object storage.

### 7.5 Replay from archive

`helixqa-replay <run-id>` re-executes the run with the archived inputs. The replay produces a new run-archive entry with `cadence=reactive`, `subtype=replay`, `replays=<original-run-id>`. This supports two workflows:

- **Regression bisection**: replay a sequence of runs across a date range to locate the run where a change-point first appeared.
- **Debugging**: replay a specific failing run to capture additional traces (e.g. with `OTEL_LOG_LEVEL=debug`) that the original run did not capture.

---

## 8. Alert Routing and On-Call Rotation

HelixQA's alerting fans out P1 / P2 / P3 alerts to the on-call rotation. The routing rules are defined in CUE (`alert-routing/policies/*.cue`) and evaluated at run-archive write time.

### 8.1 Severity definitions

| Severity | Definition                                                                                  | Page method                              | Acknowledge SLA |
|----------|---------------------------------------------------------------------------------------------|------------------------------------------|-----------------|
| **P1**   | Fleet-down or catalog SPOF compromise (`helix-r18-safeexec` CVE; multiple Challenges red across mirrors). | Phone call + SMS + Slack #helixplay-p1 | 15 minutes      |
| **P2**   | Real regression candidate; mirror-divergence; baseline-replacement gate failed.            | Slack #helixplay-p2 + email             | 4 hours         |
| **P3**   | Flaky scenario; single-mirror divergence; visibility-audit drift on a Russian mirror.      | Slack #helixplay-p3                      | 24 hours        |

### 8.2 The CUE policy shape

```cue
package alert

import "list"

#P1FleetDown: {
    cadence: "nightly" | "canary" | "pre-release"
    scope: "fleet" | "topology"
    redMirrors: list.MinItems(2)
    triggeredBy: "Challenges" | "deployment-gate" | "host-integrity-scan"
}

#P2MirrorDivergence: {
    cadence: string
    scope: "submodule"
    redMirrors: list.MinItems(1)
    redMirrors: list.MaxItems(1)
    triggeredBy: "Challenges" | "parity-audit"
}
```

The policy file is parsed by the orchestrator at startup; any malformed CUE fails the orchestrator's startup with a non-zero exit (Constitution §11.5.4 fail-closed posture).

### 8.3 On-call rotation

The on-call rotation lives in `HelixDevelopment/HelixQA/docs/runbook/oncall-rotation.md` (the file lists the operators on rotation; rotation cadence is weekly). HelixQA queries the rotation file when paging.

P1 + P2 alerts page **two** operators (primary + secondary on-call). P3 alerts page only the primary. The two-operator P1 + P2 paging is a defence against a single-operator misconfiguration (e.g. on-call's phone is muted) blocking response.

### 8.4 Runbooks

Every alert type has a runbook in `HelixDevelopment/HelixQA/docs/runbook/`. A runbook is a step-by-step recovery procedure. Examples:

- `runbook/p1-fleet-down.md` — how to triage a fleet-down alert.
- `runbook/p2-mirror-divergence.md` — how to investigate a single-mirror Challenges red.
- `runbook/p3-flaky-scenario.md` — how to file a flake-investigation ticket.

The runbook is part of the alert payload (the Slack message includes a link to the runbook). On-call's first action on receiving an alert is to open the runbook.

---

## 9. Deployment Gating (the cosign + Challenges + SBOM + visibility quad)

HelixQA gates every deployment on **four** simultaneous green signals. A deployment that has only three of four green is **not deployed**; the operator must investigate.

### 9.1 The four-signal gate

| Signal               | What it asserts                                                                              | Source                            |
|----------------------|-----------------------------------------------------------------------------------------------|-----------------------------------|
| **Cosign + SLSA L3** | The image was built by the canonical CI workflow, signed via Sigstore keyless, and ships with SLSA L3 provenance. | S02 §6.                           |
| **Challenges green** | The image's submodule passed its primary Challenges scenario (per-PR cadence) plus its full Challenges set (canary cadence). | S03 §4 + S03 §9.                  |
| **SBOM emitted**     | Both cyclonedx-gomod (Go SBOM) and syft (container SBOM) emitted on release; both attached to the registry; both verified against the build platform's signing key. | S01 §4.4 + S02 §6.                |
| **Visibility audit** | The submodule's repository is public on **all four** mirrors; the licence audit passed (no AGPL / GPL / LGPL transitive deps). | S01 §4.7.5 + S01 §4.8.7.          |

### 9.2 The gate decision engine

`orchestrator/internal/gate/` is the decision engine. Pseudocode:

```go
func (g *Gate) Decide(image ImageDigest) Verdict {
    cosignOK := g.verifyCosign(image)
    slsaOK := g.verifySLSA(image)
    challengesOK := g.verifyChallenges(image)
    sbomOK := g.verifySBOM(image)
    visibilityOK := g.verifyVisibility(image)

    if cosignOK && slsaOK && challengesOK && sbomOK && visibilityOK {
        return VerdictDeploy
    }
    return VerdictBlock{Reasons: g.collectFailures()}
}
```

The engine is **fail-closed**: if any of the five sub-checks errors out (e.g. Sigstore Rekor unreachable; per S02 §11 OQ-S02-B), the verdict is `VerdictBlock`. The deployment does not proceed until the operator manually overrides via `helixqa-cli gate-override <image-digest> --reason "<text>"` — an override is logged in the run-archive with the operator's identity and reason, and triggers a P2 alert to the secondary on-call.

### 9.3 The relationship to release-trains

Per S01 §9, a `v1.0.0+` graduation requires two consecutive green Ten-test cycles. HelixQA's deployment gate is the operationalisation of "green Ten-test cycle": the gate's four-signal decision is what "green" means.

A submodule cannot graduate to `v1.0.0` if the gate has blocked any deployment of any of its `v0.x.y` images in the past 14 days. The 14-day cooling-off period prevents a hot-fix-and-graduate pattern that would skip the canary cadence.

### 9.4 The deployment flow

```
                                            ┌─────────────────┐
                                            │ helixqa gate    │
                                            │ Decide(image)   │
                                            └────────┬────────┘
                                                     │
                              ┌──────────────────────┼──────────────────────┐
                              │                      │                      │
                       VerdictDeploy           VerdictBlock           VerdictError (fail-closed)
                              │                      │                      │
                              ▼                      ▼                      ▼
                    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
                    │ HelixQA records │    │ HelixQA records │    │ HelixQA pages   │
                    │ green; deploy   │    │ block; pages    │    │ secondary on-   │
                    │ proceeds        │    │ on-call P2      │    │ call P2         │
                    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

The decision is irrevocable for a given image digest. A new image (new digest) requires a fresh decision.

---

## 10. Integration with the Other Organisational Repos

HelixQA does not operate in isolation. Its integration surfaces with each of the other organisational repos:

### 10.1 `vasic-digital/Containers` (S02)

- HelixQA reads the per-submodule lane definitions in `lanes/<name>/` to know what to invoke per-PR.
- HelixQA reads `policies/*.cue` to verify deployment-time admission policy.
- HelixQA reads `ci-fragments/release-train.yml` to know the v1.0.0 graduation gate.
- HelixQA writes cosign signatures + SLSA bundles via `scripts/cosign-keyless-sign.sh` and `scripts/slsa-provenance-generate.sh`.

### 10.2 `vasic-digital/Challenges` (S03)

- HelixQA reads `topologies/<name>/` and `baselines/<name>/<scenario>/` to know what to run.
- HelixQA invokes `scripts/challenge-run.sh <submodule> <topology> <scenario>` per cadence.
- HelixQA writes the change-point reports back to the run-archive (it does **not** write back into `Challenges/baselines/` — that's an operator-PR-only path per S03 §6.1).

### 10.3 The 29 `vasic-digital/helix-*` submodules (S01)

- HelixQA reads each submodule's CI workflow status via the four-mirror APIs.
- HelixQA reads each submodule's release-train tag activity.
- HelixQA writes deployment-gate verdicts as GitHub commit-status / GitLab pipeline-status / GitFlic + GitVerse equivalents.

### 10.4 `vasic-digital/Catalogizer` and `HelixDevelopment/HelixAgent` (referenced precedents)

R-14 names HelixAgent and Catalogizer as the prior projects whose Challenges discipline HelixPlay inherits. HelixQA's design borrows directly from those projects' QA orchestrators; the precedents are not technical dependencies but they are documented inheritances. A future operator who reads HelixQA's orchestrator code should find the same patterns documented in HelixAgent's and Catalogizer's QA tooling.

### 10.5 Tracking on GitHub Projects + GitLab (R-17)

HelixQA mirrors every alert and every gate decision to:

- The `HelixDevelopment/HelixPlay` GitHub Project (board: "Operations").
- The `helixdevelopment1/HelixPlay` GitLab equivalent (board: "Operations").

The mirroring uses `gh project item-create` and `glab project issue create` respectively. Every ticket carries a `[helixqa-run-id:<run-id>]` label so an operator can grep across both platforms for the same run.

---

## 11. Open Questions

| ID            | Question                                                                                          | Defer to                                            |
|---------------|---------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-S04-A      | Canary's 24h stress topology — single-mirror vs four-mirror? Compute cost vs detection coverage. | `08_Operations/01_Container_CI_CD.md`              |
| OQ-S04-B      | Sigstore Rekor outage handling — fail-closed accepted? Impact on release-train cadence.           | `08_Operations/02_Quality_Gates_SonarQube_Snyk.md` |
| OQ-S04-C      | On-call rotation cadence (weekly vs bi-weekly) — operator preference + burnout study.            | `08_Operations/04_Observability_and_Events.md`     |
| OQ-S04-D      | Run-archive retention — forever vs cold-tier eviction at 5 years? Compliance vs storage cost.    | `08_Operations/05_Tracking_GitHub_GitLab.md`       |

None of the four are placeholders; each has a named resolution chapter and a specific operational concern.

---

## 12. References & Anti-Bluff Verification

### 12.1 Internal

- [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md) §3, §4, §5, §9.
- [`02_Containers_Submodule.md`](02_Containers_Submodule.md) §3, §6, §7, §10.
- [`03_Challenges_Submodule.md`](03_Challenges_Submodule.md) §4, §6, §8, §9.
- [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row S04.
- [`../01_Constitution.md`](../01_Constitution.md) §6 (R-11/R-12/R-13/R-14), §8 (R-16/R-17), §11.5 (R-18), §16 (Acceptance).
- [`../02_System_Overview.md`](../02_System_Overview.md) §15 (Release Trains).

### 12.2 External (web)

- Sigstore: https://www.sigstore.dev/ (accessed 2026-04-29).
- Cosign keyless: https://docs.sigstore.dev/cosign/signing/overview/ (accessed 2026-04-29).
- SLSA v1.0: https://slsa.dev/spec/v1.0/ (accessed 2026-04-29).
- CUE: https://cuelang.org/ (accessed 2026-04-29).
- OpenTelemetry collector: https://opentelemetry.io/docs/collector/ (accessed 2026-04-29).
- GitHub Projects API: https://docs.github.com/en/rest/projects (accessed 2026-04-29).
- GitLab Projects API: https://docs.gitlab.com/ee/api/projects.html (accessed 2026-04-29).

### 12.3 Anti-Bluff Verification

| Path                                                     | Lines  | Reviewed   | Role                                            |
|----------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md)    |  1,218 | 2026-04-30 | catalog rows + cross-cutting policies           |
| [`02_Containers_Submodule.md`](02_Containers_Submodule.md) |    627 | 2026-04-30 | containers + signing + admission policies      |
| [`03_Challenges_Submodule.md`](03_Challenges_Submodule.md) |    517 | 2026-04-30 | Challenges topology + cadence contract         |
| [`../01_Constitution.md`](../01_Constitution.md) §6 §8 §11.5 §16 | (cited slices) | 2026-04-30 | R-11/R-12/R-13/R-14/R-16/R-17/R-18 + Acceptance |

- Coverage: this chapter exceeds the 400-line floor (`wc -l` recorded at chapter close).
- Forbidden patterns: clean. No `TODO` / `FIXME` / `tbd` / `xxx` / `???` / `placeholder` / "fill in later" markers in the chapter body. The four §11 open questions are explicitly named with deferred resolution chapters.
- R-13 anti-bluff posture: the deployment gate's four-signal quad (§9.1) closes the loop from S03's Challenges-as-meta-test through HelixQA's autonomous orchestration to the operator's deployment-gate-override on the rare cases the gate fails-closed.
- R-14 wording (Constitution §6.4) is operationalised in §1 + §10.4.
- R-17 mirroring is operationalised in §10.5.
- R-18 inheritance: HelixQA orchestrators run inside containers (S02 §3), inherit the host-integrity-scan harness (S02 §3.4), and never invoke any forbidden host-disruptive command (the orchestrator is itself a SafeExec consumer per S01 §7.3).

### 12.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 as a single inline `Write` call.
- Reviewed by: pending operator review.

End of `06_Submodules/04_HelixQA_Integration.md` — 2026-04-30.
