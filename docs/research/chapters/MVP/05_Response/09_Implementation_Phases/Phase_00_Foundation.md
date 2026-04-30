# Phase_00 — Foundation

> **Source dimensions:** [`00_Phase_Index.md`](00_Phase_Index.md), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P00, [`../01_Constitution.md`](../01_Constitution.md) (entire — Phase_00 stands up the operational substrate that satisfies every R-NN), [`../06_Submodules/`](../06_Submodules/) (catalog), [`../07_Testing/`](../07_Testing/), [`../08_Operations/`](../08_Operations/).
> **Source line count:** chapter floor 800 lines per Master Plan §7.2 row P00 (largest Phase floor — Phase_00 is the load-bearing infrastructure milestone).
> **Phase targets:** All R-NN clauses (Phase_00 is the operator's "make the project tractable" milestone before any code-bearing phase begins).
> **Cross-links:** [`Phase_01_Containers_and_CI.md`](Phase_01_Containers_and_CI.md), [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md), [`../07_Testing/01_Test_Matrix.md`](../07_Testing/01_Test_Matrix.md), [`../08_Operations/`](../08_Operations/) (all 6 chapters).
> **Status:** Draft v1 specification (execution: pending operator scheduling).
> **Last updated:** 2026-04-30.

---

## Table of Contents

1. Phase Scope
2. Prerequisites
3. Tasks Catalogue
4. Task Details (T01..T12)
5. Subtask Catalogue
6. Exit Criteria
7. Risk Register
8. Cross-Family Dependencies
9. Acceptance Criteria
10. Anti-Bluff Verification

---

## 1. Phase Scope

Phase_00 *Foundation* is the operator-side **infrastructure-and-tooling** milestone. It is the **only** Phase with no submodule deliverables — its work is environmental: provisioning the four-mirror repository topology, setting up CI runner pools, deploying the foundational backing services (Vault, CockroachDB, NATS, Redis, the observability stack), provisioning the GitHub Projects + GitLab boards (W07), and standing up the operator's signing keys + Sigstore + Rekor relationship.

After Phase_00, every subsequent Phase has a working substrate — `vasic-digital/Containers` (Phase_01) can build images; `vasic-digital/<submodule>` (Phase_02) can be tagged + pushed; HelixQA (the orchestrator from S04 + T12) can begin operating its run-archive + alert routing.

Phase_00 is **not** code-bearing — no submodule's `v0.x.y` is shipped here. The submodule scaffolds are created (per [O06 §7](../08_Operations/06_Git_Topology_and_Push_Policy.md#7-submodule-repository-topology) submodule-bootstrap script), but their first actual code commits land in Phase_01 + Phase_02.

---

## 2. Prerequisites

Phase_00 has **no predecessor Phase**. It does have **operator prerequisites**:

- The operator has accounts on GitHub + GitLab + GitFlic + GitVerse.
- The operator has obtained the necessary org-level permissions on each (`HelixDevelopment` + `vasic-digital` org admin, plus billing-admin for paid plans).
- The operator has identified the LAN topology + the hosting providers for backing services (CockroachDB, Vault, NATS, etc.).
- The operator has a minisign key for SBOM signing (or generates one in T03).
- The operator has the budget for the §11 cost-line items (estimated $300–$500 / month in Phase_00 ramp; ramps higher in subsequent Phases).

These are out-of-band prerequisites; they are **not** Master-Plan-level work.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks | Depends on  |
|------------|---------------------------------------------------------------|---------:|-------------|
| P00.T01   | Provision 4-mirror org accounts + repos (HelixDevelopment + vasic-digital) | 4 | (none)        |
| P00.T02   | Provision CI runner pools (4 mirrors × N lanes)               | 4 | P00.T01     |
| P00.T03   | Generate operator signing keys (cosign + minisign) + Sigstore | 3 | (none)        |
| P00.T04   | Deploy Vault + load operator KEK                              | 4 | P00.T01     |
| P00.T05   | Deploy CockroachDB + NATS + Redis (containerised)             | 6 | P00.T01     |
| P00.T06   | Deploy observability stack (OTel + Prom + Grafana + Loki + Tempo) | 6 | P00.T05    |
| P00.T07   | Provision GitHub Projects + GitLab boards (W07)               | 5 | P00.T01     |
| P00.T08   | Bootstrap the 29 vasic-digital submodule scaffolds            | 3 | P00.T01     |
| P00.T09   | Create the operator runbook directory + initial runbooks      | 4 | (none)        |
| P00.T10   | Provision SonarQube CE + Snyk + Trivy infra                   | 5 | P00.T05     |
| P00.T11   | Configure four-mirror DNS / discovery infrastructure          | 3 | P00.T05     |
| P00.T12   | Phase_00 acceptance review                                     | 2 | all of T01–T11 |

Total: 12 tasks, 49 subtasks. Each is mirrored to GitHub Projects + GitLab per [O05](../08_Operations/05_Tracking_GitHub_GitLab.md).

---

## 4. Task Details

### 4.1 P00.T01 — Provision 4-mirror org accounts + repos

**Subtasks:**

- **P00.T01.S01**: Create `HelixDevelopment/HelixPlay` repository on GitHub + import the existing local repo (this synthesis programme's commits) — already done in Sessions 1–10.
- **P00.T01.S02**: Create the same on GitLab (`helixdevelopment1/HelixPlay`), GitFlic (`helixdevelopment/helixplay`), GitVerse (`helixdevelopment/HelixPlay`) — already done.
- **P00.T01.S03**: Create the `vasic-digital` org accounts on all four mirrors. Configure `Members cannot create private repositories` setting per [S01 §4.7](../06_Submodules/01_Submodule_Catalog.md#47-public-visibility-enforcement-across-four-mirrors).
- **P00.T01.S04**: Configure `origin` remote's composite-push for the repo-root via [O06 §10a setup script](../08_Operations/06_Git_Topology_and_Push_Policy.md#10a-the-gitconfig-operator-setup).

**Exit:** all four mirrors reachable; composite-push works for `git push origin main` from the operator's host.

### 4.2 P00.T02 — Provision CI runner pools

**Subtasks:**

- **P00.T02.S01**: GitHub Actions runner pool — accept default GitHub-hosted runners or provision self-hosted runners via Actions Runner Controller for higher throughput. 8 concurrent lanes per [T01 §14d](../07_Testing/01_Test_Matrix.md#14d-ci-concurrency-limits).
- **P00.T02.S02**: GitLab CI runner pool — operator's plan determines minute quota; 8 concurrent lanes target.
- **P00.T02.S03**: GitFlic CI runner pool — operator-provisioned VMs running `gitflic-runner`. 4 concurrent lanes target.
- **P00.T02.S04**: GitVerse CI runner pool — operator-provisioned VMs running `gitverse-runner`. 4 concurrent lanes target.

**Exit:** a probe push to each mirror produces a successful CI lane invocation; idle queue depth = 0; runners report healthy in their respective dashboards.

### 4.3 P00.T03 — Operator signing keys + Sigstore

**Subtasks:**

- **P00.T03.S01**: Generate cosign keyless OIDC trust anchor for each mirror. The Sigstore Rekor public log is the verification trust anchor — no operator-managed key for cosign-keyless.
- **P00.T03.S02**: Generate minisign keypair for SBOM signing. Public key published at `vasic-digital/.github/sbom-signing-pubkey.minisig` per [S01 §4.4](../06_Submodules/01_Submodule_Catalog.md#44-sbom-generation-per-submodule-cyclonedx-gomod--syft-dual-format).
- **P00.T03.S03**: Generate the deployment-gate override key (operator-only) per [T12 §10b](../07_Testing/12_HelixQA_Autonomous.md#10b-override-audit-trail-format). Stored in HashiCorp Vault.

**Exit:** the operator can sign + verify a probe artefact via cosign + minisign + the override key.

### 4.4 P00.T04 — Deploy Vault

**Subtasks:**

- **P00.T04.S01**: Deploy HashiCorp Vault Enterprise (or OSS with manual unseal) per [helix-vault descriptor §3.3](../06_Submodules/per-submodule/helix-vault.md#33-external-system).
- **P00.T04.S02**: Configure AppRole auth method + create per-submodule AppRole policies.
- **P00.T04.S03**: Generate + load the operator KEK for Phase_00. KEK rotation cadence is annual per [helix-vault §9.1](../06_Submodules/per-submodule/helix-vault.md#91-configuration-knobs).
- **P00.T04.S04**: Configure Vault audit log to syslog + SIEM forward.

**Exit:** Vault accessible from each CI runner pool via `helix-vault.New(...)`; operator KEK present + audit log streaming.

### 4.5 P00.T05 — Backing services

**Subtasks:**

- **P00.T05.S01**: Deploy CockroachDB single-node (Phase_00) → cluster (later Phases). CockroachDB ≥ v25.x, image-pinned per [T03 §7b](../07_Testing/03_Integration_Tests.md#7b-container-image-pin-audit).
- **P00.T05.S02**: Deploy NATS single-node → cluster (Phase_03 expansion).
- **P00.T05.S03**: Deploy Redis single-node.
- **P00.T05.S04**: Deploy Coturn (TURN server for ICE; consumed by helix-transport in Phase_04).
- **P00.T05.S05**: Deploy MinIO (S3-compatible object storage; consumed by helix-record in Phase_09 + general SBOM storage).
- **P00.T05.S06**: Configure mDNS for each service per [O03 §2](../08_Operations/03_Service_Discovery_and_Ports.md#2-mdns--dns-sd-service-discovery).

**Exit:** each backing service responds to its respective health probe; mDNS-discoverable from any LAN client.

### 4.6 P00.T06 — Observability stack

**Subtasks:**

- **P00.T06.S01**: Deploy OTel collector per [O04 §8](../08_Operations/04_Observability_and_Events.md#8-the-otel-collector-pipeline-configuration).
- **P00.T06.S02**: Deploy Prometheus + AlertManager.
- **P00.T06.S03**: Deploy Grafana + provision the four canonical dashboards per [O04 §9c](../08_Operations/04_Observability_and_Events.md#9c-the-operators-dashboard-provisioning-one-time-setup).
- **P00.T06.S04**: Deploy Loki + promtail.
- **P00.T06.S05**: Deploy Tempo.
- **P00.T06.S06**: Configure HelixQA alert-routing webhook → Slack + PagerDuty per [T12 §4](../07_Testing/12_HelixQA_Autonomous.md#4-alert-routing-playbook).

**Exit:** every backing service emits metrics + traces + logs to the stack; dashboards render real data.

### 4.7 P00.T07 — GitHub Projects + GitLab boards (W07)

**Subtasks:**

- **P00.T07.S01**: Create the GitHub Project at `HelixDevelopment/HelixPlay/projects` — board "Operations" with the column layout per [O05 §8](../08_Operations/05_Tracking_GitHub_GitLab.md#8-the-project-board-layout).
- **P00.T07.S02**: Create the GitLab equivalent at `helixdevelopment1/HelixPlay/-/boards`.
- **P00.T07.S03**: Configure the bidirectional webhook per [O05 §6](../08_Operations/05_Tracking_GitHub_GitLab.md#6-the-bidirectional-mirror-webhook).
- **P00.T07.S04**: Run the bulk-provisioning script per [O05 §9](../08_Operations/05_Tracking_GitHub_GitLab.md#9-the-operators-bulk-issue-creation-workflow) to import every of the 14 phases × tasks × subtasks.
- **P00.T07.S05**: Configure HelixQA's `tracking-mirror` service per [T12 §6](../07_Testing/12_HelixQA_Autonomous.md#6-tracking-mirror-to-github-projects--gitlab-r-17).

**Exit:** every of the 14 phases has a corresponding GitHub Project item + GitLab issue; webhook is operational; HelixQA's auto-creation of P1/P2/P3 alert tickets fires correctly.

### 4.8 P00.T08 — Bootstrap submodule scaffolds

**Subtasks:**

- **P00.T08.S01**: Run `submodule-bootstrap.sh` per [O06 §7](../08_Operations/06_Git_Topology_and_Push_Policy.md#7-submodule-repository-topology) for each of the 29 catalog rows. Creates the four-mirror repos + initial `go.mod` + LICENSE + README + `.snyk` + `.coverage-exemptions.yaml`.
- **P00.T08.S02**: Install branch-protection rules per [O06 §6](../08_Operations/06_Git_Topology_and_Push_Policy.md#6-branch-protection-rules) — `main` protected, ≥ 1 approval (≥ 2 for `helix-r18-safeexec`).
- **P00.T08.S03**: Create the per-submodule `tests/` directory tree per [T01 §7](../07_Testing/01_Test_Matrix.md#7-mock-policy-enforcement-r-12).

**Exit:** 29 vasic-digital repos exist on all four mirrors with consistent scaffolds + branch protection.

### 4.9 P00.T09 — Operator runbook directory

**Subtasks:**

- **P00.T09.S01**: Create `HelixDevelopment/HelixQA/docs/runbook/` with `oncall-rotation.md`, `p1-fleet-down.md`, `p2-mirror-divergence.md`, `p3-flaky-scenario.md`, plus per-failure-mode runbooks referenced throughout the synthesis programme.
- **P00.T09.S02**: Create `HelixDevelopment/HelixQA/docs/onboarding/oncall.md` per [T12 §10a](../07_Testing/12_HelixQA_Autonomous.md#10a-the-operator-onboarding-path).
- **P00.T09.S03**: Create the soak-postmortem template at `vasic-digital/.github/soak-postmortem-template.md` per [T08 §9a](../07_Testing/08_Stress.md#9a-soak-test-findings-pattern-postmortem-template).
- **P00.T09.S04**: Create the override-audit-log directory at `HelixDevelopment/HelixQA/audit/overrides/`.

**Exit:** runbooks committed; auditable.

### 4.10 P00.T10 — SonarQube + Snyk + Trivy

**Subtasks:**

- **P00.T10.S01**: Deploy SonarQube CE per [O02 §3](../08_Operations/02_Quality_Gates_SonarQube_Snyk.md#3-sonarqube-integration). Configure HelixPlay-Strict quality profile.
- **P00.T10.S02**: Bulk-create per-submodule SonarQube projects per [O02 §10a](../08_Operations/02_Quality_Gates_SonarQube_Snyk.md#10a-per-submodule-sonar-project-provisioning).
- **P00.T10.S03**: Configure Snyk org + integrate with GitHub + GitLab webhook.
- **P00.T10.S04**: Configure Trivy as a service (hosted or operator-deployed).
- **P00.T10.S05**: Configure HelixQA `snyk-event-router` per [O02 §10b](../08_Operations/02_Quality_Gates_SonarQube_Snyk.md#10b-snyk-webhook--slack-integration).

**Exit:** the all-three-pass gate works on a probe PR; SonarQube + Snyk + govulncheck + Trivy all reachable from each CI runner.

### 4.11 P00.T11 — DNS / discovery infrastructure

**Subtasks:**

- **P00.T11.S01**: For deployments without LAN-mDNS, deploy the discovery server per [O03 §9b](../08_Operations/03_Service_Discovery_and_Ports.md#9b-fallback-discovery-server-topology).
- **P00.T11.S02**: Configure DNS-over-HTTPS as alternative to mDNS for Russian-jurisdiction operators per [O03 §11.4](../08_Operations/03_Service_Discovery_and_Ports.md#114-the-mdns-vs-dns-over-https-trade-off).
- **P00.T11.S03**: Configure the operator's LAN firewall to allow UDP 5353 + the dynamic-port range.

**Exit:** mDNS-or-fallback discovery operational; an LAN client can resolve `_helix-host._grpc._tcp.local.`.

### 4.11a P00.T11 detailed — DNS / discovery configuration walkthrough

For LAN deployments (the canonical home-broadband + small-enterprise topology):

1. Configure the operator's local LAN switch / router to allow IPv4 multicast `224.0.0.251:5353` + IPv6 multicast `ff02::fb:5353` (mDNS).
2. Verify mDNS works from a probe client: `dig +short -p 5353 @224.0.0.251 _services._dns-sd._udp.local. PTR`.
3. Configure systemd-resolved or nsswitch.conf to route `*.local` queries via mDNS rather than upstream DNS.
4. For Windows-side participation: install Bonjour Print Services + verify `dns-sd` resolves the test record.

For deployments where mDNS is restricted (some corporate networks):

1. Deploy the discovery server per [O03 §9b](../08_Operations/03_Service_Discovery_and_Ports.md#9b-fallback-discovery-server-topology).
2. Configure each HelixPlay client with `HELIX_DISCOVERY_SERVER=https://discovery.helix.example.com:8090`.
3. Verify by `curl https://discovery.helix.example.com:8090/api/v1/services/_helix-host._grpc._tcp/`.

For Russian-jurisdiction operators preferring DoH:

1. Deploy a private DNS zone (e.g. `helix.example.ru`) hosted on Cloudflare DNS or operator-self-hosted CoreDNS.
2. Populate the zone with SRV records mirroring what mDNS would produce.
3. Configure each HelixPlay client with `HELIX_DOH_RESOLVER=https://1.1.1.1/dns-query` (or operator's preferred DoH server) + `HELIX_DOH_ZONE=helix.example.ru`.

The three modes are mutually exclusive at the deployment level; operator picks one + verifies via probe queries before unblocking Phase_01.

### 4.11b P00.T11 detailed — firewall configuration walkthrough

Operator's LAN firewall opens:

| Direction | Protocol | Port range          | Purpose                                                |
|-----------|----------|---------------------|--------------------------------------------------------|
| Inbound   | UDP      | 5353                | mDNS                                                    |
| Inbound   | TCP      | 8443, 8200, 26257, 4222, 6379, 3000, 9090, 4317, 4318 | Default backing-service ports |
| Inbound   | TCP/UDP  | 32768-60999         | Dynamic-port range per [O03 §3](../08_Operations/03_Service_Discovery_and_Ports.md#3-dynamic-port-assignment) |
| Outbound  | TCP      | 443                  | HTTPS (Sigstore Rekor, GitHub/GitLab API, Slack/PagerDuty webhooks) |
| Outbound  | TCP      | 22                   | SSH (operator administrative access only)              |

Tighter firewall postures pin narrower ranges via the per-submodule `HELIX_*_PORT_RANGE` env-vars.

### 4.12 P00.T12 — Phase_00 acceptance review

**Subtasks:**

- **P00.T12.S01**: Operator runs the §6 *Exit Criteria* checklist + verifies every condition.
- **P00.T12.S02**: Operator signs off on Phase_00 ticket per Constitution §16 *Acceptance*.

**Exit:** Phase_00 closed; Phase_01 unblocked.

---

### 4.13 The 29-Submodule Bootstrap Matrix (P00.T08 detail)

Each submodule's bootstrap creates a four-mirror repository with a consistent scaffold:

```
vasic-digital/<submodule>/
├── .github/                       (issue + PR templates)
├── .gitignore
├── .snyk
├── .coverage-exemptions.yaml
├── go.mod                          (Go 1.24 toolchain pin)
├── go.sum
├── LICENSE                         (MIT or Apache-2.0 per S01 §4.8)
├── README.md                       (badge + quickstart + cross-link to S05 descriptor)
├── CHANGELOG.md                    (initial v0.1.0 stub)
├── Dockerfile                      (multi-stage; builder + distroless runtime)
├── docker-compose.yml              (per-submodule integration-test fixture; if applicable)
├── cmd/                            (binary entry point if applicable)
├── internal/                       (private implementation)
├── pkg/<exported>/                 (public API surface per S05 §2)
├── tests/
│   ├── unit/main_test.go           (per T02 §7a TestMain skeleton)
│   ├── integration/
│   ├── e2e/
│   ├── security/
│   ├── benchmarking/
│   ├── chaos/
│   ├── stress/
│   ├── smoke/probe.sh              (per T09 §4)
│   ├── full-automation/
│   └── challenges/                 (or delegated.md for helix-shm + helix-bench)
└── migrations/                     (golang-migrate format if SQL-bearing)
```

The bootstrap script populates each file from canonical templates; per-submodule customisation happens in subsequent Phases (Phase_02 + onwards). The 29 submodule names (per [S01 §3.1](../06_Submodules/01_Submodule_Catalog.md#31-the-29-submodule-table)) are bootstrapped sequentially to avoid GitHub API rate-limiting.

### 4.14 Operational scripts catalogue (P00 introduces these)

Phase_00 ships the following operational scripts at `vasic-digital/.github/scripts/` + `vasic-digital/Containers/scripts/` + `HelixDevelopment/HelixQA/scripts/`:

| Script                                              | Purpose                                                              |
|-----------------------------------------------------|----------------------------------------------------------------------|
| `git-onboard.sh`                                    | Configures the four-mirror remotes + composite origin (per O06 §10a)|
| `submodule-bootstrap.sh`                            | Creates a new vasic-digital submodule with the §4.13 scaffold        |
| `provision-tracking.sh`                             | Bulk-imports the 14-phase × tasks × subtasks to GitHub + GitLab boards |
| `phase-00-tasks.yaml`                               | The §8b master subtask list                                           |
| `multiarch-build.sh`                                | docker buildx wrapper for amd64+arm64 + 4-mirror push                |
| `cosign-keyless-sign.sh`                            | Sigstore keyless signing wrapper                                      |
| `slsa-provenance-generate.sh`                       | SLSA L3 provenance generator                                          |
| `host-integrity-scan-run.sh`                        | strace + auditd harness wrapper                                       |
| `four-mirror-replication-audit.sh`                  | Nightly registry replication audit                                    |
| `four-mirror-challenges-parity.sh`                  | Nightly Challenges parity audit                                       |
| `helixqa-deploy.sh`                                 | Deployment-gate decision engine                                       |
| `helix-leak-regress`                                | Stress-test slope-detection tool                                      |
| `helix-test-select`                                 | Per-PR diff-driven test selector                                      |
| `helix-test-matrix-lint`                            | Validates T01 §2 matrix coverage                                      |
| `helix-mock-discipline`                             | Rejects mocks outside tests/unit/                                     |
| `helix-no-skip`                                     | Rejects t.Skip annotations                                             |
| `helix-integration-coverage`                        | Verifies every exported symbol has integration coverage              |
| `helix-e2e-journey-coverage`                        | Verifies every System Overview §3 step has E2E coverage              |
| `helix-vuln-policy`                                 | Reachability-vs-presence vuln decision engine                         |
| `helix-vuln-suppression-audit`                      | Nightly suppression-expiry audit                                      |
| `helix-spdx-check`                                  | SPDX header lint                                                      |
| `helix-toolchain-pin-audit`                         | Per-submodule toolchain pin verification                              |
| `helix-coverage-exemption-audit`                    | Coverage-exemption expiry audit                                       |
| `helix-image-pin-audit`                             | Container image digest pin verification                               |
| `helix-builder-pin-audit`                           | Builder image pin verification                                        |
| `helix-protection-compare`                          | Branch-protection drift audit across 4 mirrors                       |
| `helix-tracking-audit`                              | GitHub-Projects + GitLab issue-bi-direction audit                    |
| `helix-tracking-drift-audit`                        | Per-mirror tracking drift audit                                       |
| `helix-workflow-yaml-parity`                        | Per-mirror workflow YAML equivalence audit                           |
| `helix-chaos-coverage`                              | Per-submodule §9.3 failure-mode coverage audit                       |
| `helix-chaos-blast-audit`                           | Chaos blast-radius containment audit                                  |
| `helix-down-migration-audit`                        | up/down migration pairing audit                                       |
| `helix-allocator-vet`                               | Hot-path allocation enforcer                                          |
| `helix-empty-test-detector`                         | Detects assertion-free tests                                          |
| `helix-dep-cycle-check`                             | Dependency-graph DAG verification                                     |
| `helix-bulk-issue-create`                           | CSV-driven bulk issue creation                                        |
| `helix-otel-init`                                   | Per-submodule OTel SDK initialiser                                    |
| `gitflic-cli.go`                                    | GitFlic API wrapper                                                   |
| `gitverse-audit-manual.sh`                          | GitVerse manual visibility audit                                      |
| `discovery-server-cli`                              | Discovery server admin                                                |

40+ canonical scripts ship with Phase_00. Subsequent phases extend the catalogue but do not replace it.

### 4.15 The Operator's Communication Channels (P00 setup)

| Channel                         | Purpose                                                       |
|---------------------------------|---------------------------------------------------------------|
| Slack `#helixplay-p1`            | P1 alerts (per T12 §4.1)                                      |
| Slack `#helixplay-p2`            | P2 alerts                                                      |
| Slack `#helixplay-p3`            | P3 alerts                                                      |
| Slack `#helixplay-deploys`       | Every deployment-gate decision                                 |
| Slack `#helixplay-legal`         | Licence-policy violations                                      |
| Slack `#helixplay-general`       | Operator + maintainer chat                                     |
| Email — operator                 | Nightly digest                                                 |
| PagerDuty                        | P1 phone + SMS escalation                                      |
| GitHub Discussions               | Public RFC + maintainer Q&A                                    |

The channels are operator-provisioned during Phase_00 + appear in the per-alert payload links.

### 4.16 The Phase_00 Documentation Audit

Operator verifies the following docs exist + are committed before Phase_00 closes:

- [`/CLAUDE.md`](../../../../../CLAUDE.md) — repo-root navigation for Claude agents.
- [`/AGENTS.md`](../../../../../AGENTS.md) — non-Claude agent guidance.
- `vasic-digital/.github/CONTRIBUTING.md` — contributor onboarding for the submodules.
- `HelixDevelopment/HelixQA/README.md` — operator-facing HelixQA overview.
- Per-submodule README badges + quickstarts (29 files).

## 5. Subtask Catalogue (Aggregate)

The 12 tasks decompose into **49 subtasks** total. Each subtask is `[P00.Tyy.Szz]`-tagged + ticketed on GitHub Projects + GitLab. Operator review of each subtask happens via the §4 task-level exit criteria.

---

## 6. Exit Criteria

Phase_00 exits when:

- [ ] All 4 mirror repos at `vasic-digital/<name>` for the 29 submodules + the 4 organisational repos exist + are public + branch-protected.
- [ ] All 4 CI runner pools online + healthy.
- [ ] Operator signing keys generated + Sigstore + minisign trust anchors published.
- [ ] Vault deployed + operator KEK loaded + audit log streaming.
- [ ] CockroachDB + NATS + Redis + Coturn + MinIO deployed + health-probe-green.
- [ ] OTel + Prometheus + Grafana + Loki + Tempo deployed; the 4 canonical dashboards rendering real data.
- [ ] GitHub Projects + GitLab boards provisioned + bidirectional webhook operational + 14-phase × tasks × subtasks bulk-imported.
- [ ] 29 submodule scaffolds bootstrapped with consistent files + branch protection.
- [ ] Operator runbook directory populated with the §4.9 minimum set.
- [ ] SonarQube CE + Snyk + Trivy operational; all-three-pass gate works on a probe PR.
- [ ] mDNS / DoH discovery operational.
- [ ] Operator signoff per Constitution §16 on the Phase_00 ticket.

11 conditions; all must be green.

---

## 6a. The Phase_00 Decision-Point Catalogue

Phase_00 has decision points that affect every subsequent Phase. Operator must decide before unblocking Phase_01:

| Decision                         | Options                                              | Default                  | Latch-in cost  |
|----------------------------------|------------------------------------------------------|--------------------------|----------------|
| Vault Enterprise vs OSS          | Auto-Unseal vs Shamir-share manual                   | Operator-choice          | Mid (rotate)   |
| GOCACHEPROG remote cache         | Buildbarn (self-host) vs BuildBuddy (SaaS)          | BuildBuddy free tier     | Low (re-config)|
| Snyk plan                        | Open Source free vs paid                             | Free for OSS submodules  | Low            |
| GitHub plan                      | Free / Team / Enterprise                             | Team for HelixDevelopment | Low           |
| Discovery model                  | mDNS / discovery server / DoH                       | mDNS for LAN; DoH for RU | Mid            |
| Observability storage retention  | hot 30 / warm 1y / cold forever vs shorter           | Default per O04 §7       | Low            |
| Tag-signing requirement          | Required for `v1.0.0+` vs optional                  | Required (recommended)   | Low            |
| Per-submodule reviewer count     | 1 vs 2 (helix-r18-safeexec is always 2)              | 1 default; 2 for SPOF    | Mid            |

Each decision is recorded in the Phase_00 ticket's body for permanent audit + future Phase reference.

## 6b. The Phase_00 Operator-Capacity Estimate

Phase_00 demands focused operator + engineering capacity:

| Role                | FTE-weeks | Activities                                                    |
|---------------------|-----------|---------------------------------------------------------------|
| Operator (lead)     | 4         | All decision points + acceptance review + Constitution §16 sign-off |
| Platform engineer   | 6         | T02 (CI runners) + T05 (backing services) + T06 (observability) + T11 (DNS) |
| Security engineer   | 3         | T03 (signing keys) + T04 (Vault) + T10 (quality-gate infra)  |
| DevOps / SRE        | 2         | T08 (submodule scaffolds) + T07 (project boards)              |
| Documentation       | 1         | T09 (runbooks)                                                 |
| **Total**           | **16 FTE-weeks** | (operator can compress with parallel work) |

A single-operator deployment scales the timeline; a four-engineer team (typical operator engineering org) completes in ~ 4 calendar weeks.

## 7. Risk Register

| ID         | Risk                                                                              | Mitigation                                                                |
|------------|------------------------------------------------------------------------------------|----------------------------------------------------------------------------|
| RP00-01    | Operator delays GitHub-vs-GitLab visibility-policy alignment                      | §4.1 default policy + §4.7 visibility audit catches drift within 24 h.    |
| RP00-02    | CI runner pool not provisioned, blocking subsequent phases                        | §6 exit criteria block Phase_01 unblock.                                  |
| RP00-03    | Vault unseal key loss — operational disaster                                      | OSS Shamir-share or Enterprise Auto-Unseal; operator's choice + recorded. |
| RP00-04    | DNS / discovery misconfiguration — clients can't find services                    | §4.11 fallback discovery + DoH alternative.                               |
| RP00-05    | Sigstore Rekor outage during Phase_00 setup                                       | §3.1 cosign keyless waits for Rekor; if outage > 4 h, fall back to keyed. |
| RP00-06    | Tracking-mirror webhook drops messages, causing GitHub-GitLab desynchronisation   | [O05 §11.4 drift audit](../08_Operations/05_Tracking_GitHub_GitLab.md#114-the-per-mirror-tracking-drift-audit) reconciles nightly. |

---

## 8. Cross-Family Dependencies

| Dependency on family                                          | Source chapter                                         |
|----------------------------------------------------------------|--------------------------------------------------------|
| `06_Submodules/` — catalog + scaffold templates                | [S01 §3.1](../06_Submodules/01_Submodule_Catalog.md#31-the-29-submodule-table) |
| `06_Submodules/02_Containers_Submodule.md` — container topology | [S02 §3](../06_Submodules/02_Containers_Submodule.md#3-the-per-submodule-ci-lane-catalog) |
| `06_Submodules/04_HelixQA_Integration.md` — HelixQA scaffold    | [S04 entire](../06_Submodules/04_HelixQA_Integration.md) |
| `07_Testing/`                                                  | [T01 §13 lifecycle](../07_Testing/01_Test_Matrix.md#13-ci-lane-anatomy--from-pr-to-green) |
| `08_Operations/`                                               | every chapter; Phase_00 is the operationalisation       |
| Constitution                                                  | §16 Acceptance signoff                                  |

---

## 8a. Phase_00 Cost Estimate

Operator's monthly spend during Phase_00 (estimated):

| Line item                                              | Monthly cost |
|--------------------------------------------------------|--------------|
| GitHub Team plan (HelixDevelopment + vasic-digital)   | $4 / user × ~ 5 users = $20 |
| GitLab Premium                                         | $19 / user × ~ 5 = $95 |
| GitFlic + GitVerse                                     | operator-self-hosted (compute only) |
| GitHub Actions runners (free tier baseline)            | $0–$200 (overage on private repos) |
| GitLab CI minutes                                      | included in Premium |
| Vault Enterprise (operator-deployed)                   | $0 (OSS) — $1k+ / month (Enterprise) |
| CockroachDB single-node                                 | $30 / month (compute) |
| NATS + Redis + Coturn + MinIO                          | $50 / month combined  |
| Observability stack (OTel + Prom + Grafana + Loki)    | $80 / month (compute + storage) |
| SonarQube CE                                            | $30 / month (compute) |
| Snyk Open Source (free tier)                           | $0 (free for OSS) |
| Sigstore Rekor                                          | $0 (public service) |
| TURN server bandwidth                                   | $20–$200 / month depending on usage |
| **Phase_00 subtotal**                                   | **$300–$1,500 / month** |

Phase_00 ramps the spend; subsequent phases add (helix-encoder NVENC sessions cost nothing in software but each NVIDIA GPU SKU is $$$$). The Phase_00 baseline is the floor.

## 8b. The 49-Subtask Master List

For operator's bulk-import via [O05 §9c](../08_Operations/05_Tracking_GitHub_GitLab.md#9c-the-operators-bulk-provisioning-one-time-workflow):

```yaml
# tasks/phase-00.yaml
- id: P00.T01.S01
  title: "[P00.T01.S01] HelixDevelopment/HelixPlay GitHub repo provisioned"
  body: "Create the public GitHub repo with branch protection rules per O06 §6."
  labels: [phase-00, subtask, priority-critical]
- id: P00.T01.S02
  title: "[P00.T01.S02] Mirror to GitLab + GitFlic + GitVerse"
  body: "Composite-push topology per O06 §3."
  labels: [phase-00, subtask, priority-critical]
- id: P00.T01.S03
  title: "[P00.T01.S03] vasic-digital org accounts on 4 mirrors"
  body: "Org-level visibility policy per S01 §4.7."
  labels: [phase-00, subtask, priority-critical]
- id: P00.T01.S04
  title: "[P00.T01.S04] Configure origin composite-push for repo-root"
  body: "Per O06 §10a setup script."
  labels: [phase-00, subtask, priority-critical]
- id: P00.T02.S01
  title: "[P00.T02.S01] GitHub Actions runner pool — 8 lanes"
  ...
- id: P00.T02.S02
  title: "[P00.T02.S02] GitLab CI runner pool — 8 lanes"
  ...
- id: P00.T02.S03
  title: "[P00.T02.S03] GitFlic CI runner pool — 4 lanes"
  ...
- id: P00.T02.S04
  title: "[P00.T02.S04] GitVerse CI runner pool — 4 lanes"
  ...
# ... (49 entries total per the §3 task table)
```

The full YAML lives at `vasic-digital/.github/scripts/phase-00-tasks.yaml`. Operator runs the §9c bulk-import once at Phase_00 start.

## 8c. Phase_00 Calendar Estimate

| Week | Activity                                                                          |
|------|-----------------------------------------------------------------------------------|
| 1    | T01 + T02 (org accounts + CI runners)                                             |
| 1–2  | T03 + T04 (signing keys + Vault)                                                  |
| 2    | T05 + T06 (backing services + observability)                                      |
| 3    | T07 + T08 (project boards + submodule scaffolds)                                  |
| 3    | T09 + T10 (runbooks + quality-gate infra)                                         |
| 4    | T11 + T12 (DNS / discovery + acceptance review)                                   |

Operator-facing duration: ~ 4 weeks of focused work. Compressible to ~ 2 weeks if operator has dedicated engineering capacity.

## 8d. The Phase_00 Smoke Test

After all 12 tasks land, the operator runs a Phase_00 acceptance smoke test:

1. From a fresh client, clone `HelixDevelopment/HelixPlay` via composite origin.
2. Push a no-op commit; verify it lands on all four mirrors.
3. Open a probe PR; verify the CI pipeline runs all 14 (test-type, platform) cells across all four mirrors.
4. Trigger a HelixQA P3 alert (e.g. `helixqa-cli synth-alert --severity=p3`); verify Slack delivery + GitHub + GitLab ticket creation.
5. Trigger a synthetic deployment-gate decision; verify the four-signal quad evaluates correctly.
6. Run a cross-mirror parity audit; verify zero divergence.

If all six probe steps pass, Phase_00 is complete + the operator can sign off. If any fail, the operator triages + fixes before unblocking Phase_01.

## 8e. The Phase_00 Reference Operator Persona

The "operator" referenced throughout HelixPlay's documentation is a deployment-owning entity (organisation or individual). Phase_00 assumes the operator profile:

- **Technical capacity**: 4–8 engineers covering Go + Linux + Kubernetes / Docker + GitHub + GitLab admin + Vault admin + cosign + Sigstore.
- **Budget**: $300–$1,500 / month for Phase_00 ramp; will scale higher.
- **Compliance posture**: GDPR-aware (helix-vault EraseTenant); EU DSA Article 17-aware (helix-tenant catalog scoping); operator may add jurisdictional audit requirements via Constitution §15 amendments.
- **Hosting**: at minimum a single-host LAN deployment (canonical home/small-office); scaling to multi-region multi-host deployments in subsequent Phases.
- **Network**: LAN with mDNS-multicast or fallback discovery server.

Operators with smaller capacity may compress Phase_00 by accepting longer wall-clock; operators with larger capacity may parallelise. The Phase exit criteria are the same regardless.

## 8f. The Acceptance Sign-off Format

Per Constitution §16, operator sign-off on Phase_00 is recorded as a structured comment on the Phase_00 GitHub Project ticket:

```markdown
## Phase_00 Acceptance — signed off 2026-MM-DD

Operator: <operator-handle> (publickey: <fingerprint>)

Exit criteria status:
- [x] All 4 mirror repos for 29 + 4 organisational repos exist + are public + branch-protected.
- [x] All 4 CI runner pools online + healthy.
- [x] Operator signing keys generated; Sigstore + minisign trust anchors published.
- [x] Vault deployed; operator KEK loaded; audit log streaming.
- [x] CockroachDB + NATS + Redis + Coturn + MinIO deployed; health probes green.
- [x] OTel + Prom + Grafana + Loki + Tempo deployed; the 4 dashboards rendering real data.
- [x] GitHub Projects + GitLab boards provisioned; webhook operational; 14-phase × tasks bulk-imported.
- [x] 29 submodule scaffolds bootstrapped with consistent files + branch protection.
- [x] Operator runbook directory populated.
- [x] SonarQube CE + Snyk + Trivy operational; all-three-pass gate works on a probe PR.
- [x] mDNS / DoH discovery operational.

Phase_00 → Phase_01 transition: APPROVED.

Decision-point latch-in (per §6a):
- Vault: OSS Shamir-share (operator-managed)
- GOCACHEPROG: BuildBuddy SaaS free tier
- Snyk: Free for OSS submodules
- GitHub: Team plan (5 seats)
- Discovery: mDNS for LAN; DoH (Cloudflare 1.1.1.1) for Russian-jurisdiction operators
- Observability retention: hot 30 / warm 1y / cold forever
- Tag signing: required for v1.0.0+
- Per-submodule reviewer count: 1 default; 2 for helix-r18-safeexec

Audit-trail: see Phase_00 ticket comments for the discussion that led to each decision.
```

The structured comment is the **canonical record** — replicated across GitHub + GitLab tickets via the bidirectional webhook + permanently archived in HelixQA's run-archive (signed manifest).

## 8g. The Phase_00 → Phase_01 Handover

When Phase_00 closes, the next-phase handover document at `09_Implementation_Phases/handover/P00_to_P01.md` records:

- Outstanding `helix-*` open issues from Phase_00 (typically all closed at signoff but exceptions logged).
- Operator-tunable settings + their values (the §6a decision-point latch-in).
- Known good-state hashes (the canonical-state SHA of HelixPlay repo + the 29 submodules + the 4 organisational repos at the moment Phase_00 closed).
- Anything Phase_01 needs to know that wasn't obvious from the §4 task details.

Handover files are append-only across phases — Phase_01 → Phase_02 handover augments the chain, never overwrites.

## 9. Acceptance Criteria

Per Constitution §16, Phase_00's Definition of Done is operator-signed-off when all §6 exit conditions are green. The signoff is non-delegable — only the operator can certify Phase_00 complete; only Phase_00 complete unblocks Phase_01.

The signoff is recorded in the Phase_00 ticket on GitHub Projects + GitLab + the run-archive's compliance log.

---

## 9a. Operational Verification Probe Catalogue

Each Phase_00 task has explicit operator-verifiable probe commands. The full catalogue (49 subtasks × 1-3 probes each):

### 9a.1 P00.T01 verification probes

```bash
# S01: HelixDevelopment/HelixPlay GitHub repo provisioned
gh repo view HelixDevelopment/HelixPlay --json visibility,defaultBranchRef
# Expect: visibility=PUBLIC, defaultBranchRef.name=main

# S02: Mirror to GitLab + GitFlic + GitVerse
for r in github gitlab gitflic gitverse; do
    git ls-remote $r main | head -1
done
# Expect: 4 lines, all the same SHA

# S03: vasic-digital org accounts on 4 mirrors
gh api orgs/vasic-digital | jq .login
glab api groups/vasic-digital | jq .name
# Expect: "vasic-digital" on both

# S04: Configure origin composite-push
git remote -v | grep "^origin"
# Expect: 1 fetch + 4 push lines
```

### 9a.2 P00.T02 verification probes

```bash
# S01-S04: Each mirror's runner pool is online
gh api repos/HelixDevelopment/HelixPlay/actions/runners | jq '.runners[] | .status'
glab ci runners list
# (GitFlic + GitVerse via mirror-specific UI)
# Expect: status=online for ≥ 4 runners on Western, ≥ 2 on Russian
```

### 9a.3 P00.T03 verification probes

```bash
# Verify cosign keyless against a probe artifact
cosign sign --identity-token=$OIDC_TOKEN --rekor-url=https://rekor.sigstore.dev probe-image:latest
cosign verify probe-image:latest
# Expect: signature appears in Rekor public log; verify exits 0

# Verify minisign keypair
minisign -P  # show pubkey
minisign -V -m sample.txt -P sbom-signing-pubkey.minisig
# Expect: verify exits 0
```

### 9a.4 P00.T04 verification probes

```bash
# Verify Vault is reachable + unsealed
vault status -address=https://vault.helix.example.com:8200
# Expect: Sealed=false, HA Mode=active

# Verify operator KEK loaded
vault kv get -address=https://vault.helix.example.com:8200 secret/helix-operator-kek
# Expect: kek-id present
```

### 9a.5 P00.T05 verification probes

```bash
# CockroachDB
cockroach sql --insecure -e 'SELECT 1' --host=cockroach.helix.example.com:26257
# Expect: 1

# NATS
nats-bench -s nats://nats.helix.example.com:4222 -n 100 helix.probe
# Expect: 100 msgs delivered

# Redis
redis-cli -h redis.helix.example.com PING
# Expect: PONG

# MinIO
mc alias set helix https://minio.helix.example.com:9000 $ACCESS_KEY $SECRET_KEY
mc admin info helix
# Expect: status=online

# Coturn
turnutils_uclient -p 3478 turn.helix.example.com
# Expect: TURN allocation succeeds
```

### 9a.6 P00.T06 verification probes

```bash
# OTel collector
curl http://otel-collector.helix.example.com:13133
# Expect: 200 OK with health response

# Prometheus
curl http://prometheus.helix.example.com:9090/-/healthy
# Expect: 200 OK

# Grafana — verify the four canonical dashboards render
curl -H "Authorization: Bearer $GRAFANA_TOKEN" \
    http://grafana.helix.example.com:3000/api/dashboards/uid/helix-fleet-health
# Expect: dashboard JSON returns

# Loki + Tempo
curl http://loki.helix.example.com:3100/ready
curl http://tempo.helix.example.com:3200/ready
# Expect: ready
```

### 9a.7 P00.T07 verification probes

```bash
# GitHub Project exists + has 14-phase items
gh project list --owner HelixDevelopment | jq '.[] | select(.title=="Operations") | .id'
gh project item-list $PROJECT_ID --format=json | jq 'length'
# Expect: ≥ 14 items + an aggregate of ~ 200+ tasks/subtasks

# GitLab board exists
glab api projects/helixdevelopment1%2FHelixPlay/boards | jq '.[] | .name'
# Expect: "Operations"

# Webhook is operational
curl -X POST $WEBHOOK_URL -H 'Content-Type: application/json' -d '{"test": "probe"}'
# Expect: 200 OK + verify GitLab issue created within 10 s
```

### 9a.8 P00.T08 verification probes

```bash
# 29 submodule scaffolds exist on each mirror
for submodule in $(cat vasic-digital/.github/helix-submodules.txt); do
    for r in github gitlab gitflic gitverse; do
        # Probe each mirror
        echo -n "$r/$submodule: "
        case $r in
          github)   gh api repos/vasic-digital/$submodule | jq .private ;;
          gitlab)   glab api projects/vasic-digital%2F$submodule | jq .visibility ;;
          # gitflic / gitverse via their respective APIs
        esac
    done
done
# Expect: every of 29 × 4 = 116 entries shows public visibility
```

### 9a.9 P00.T09 + T10 + T11 + T12 verification probes

```bash
# Runbook directory
ls HelixDevelopment/HelixQA/docs/runbook/ | wc -l
# Expect: ≥ 10 files

# SonarQube CE reachable
curl https://sonarqube.helix.example.com/api/system/status | jq .status
# Expect: UP

# Snyk webhook configured
snyk monitor --org=helix-org
# Expect: snapshot uploaded successfully

# DNS / discovery operational
dig +short -p 5353 @224.0.0.251 _helix-host._grpc._tcp.local. PTR
# Expect: at least one PTR record returned

# Phase_00 acceptance review
helixqa-cli archive-show phase-00-completion
# Expect: signed acceptance entry
```

The 49 subtasks × 1-3 probes = ~ 100 probe commands. Operator runs each + records output in the Phase_00 ticket comment.

## 10. Anti-Bluff Verification

### 10.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../06_Submodules/`](../06_Submodules/) S01..S05                  | 12,419 | 2026-04-30 | catalog scaffolds + container + Challenges + HelixQA |
| [`../07_Testing/`](../07_Testing/) T01..T12                        |  4,533 | 2026-04-30 | test discipline                                  |
| [`../08_Operations/`](../08_Operations/) O01..O06                  |  2,315 | 2026-04-30 | operational machinery                            |
| [`../01_Constitution.md`](../01_Constitution.md)                  |    879 | 2026-04-30 | every R-NN                                       |

### 10.2 Forbidden patterns

Clean.

### 10.3 Cross-link integrity

Every `[X §Y](path)` reference resolves to a real heading anchor across the prior 6 families.

### 10.4 Phase_00 Closure Note

Phase_00's 12 tasks + 49 subtasks together stand up the operational substrate that every subsequent Phase consumes by reference. A successful Phase_00 means: every required infrastructure piece is healthy + audit-trail-traceable; every operator-tunable decision is recorded; every 29-submodule scaffold is in place + four-mirror replicated; the Quality-Gate + Observability + Tracking machinery is operational. Phase_01 *Containers and CI* unblocks immediately upon Phase_00's operator signoff; from that point onward, the synthesis programme transitions from documentation-grade to execution-grade.

### 10.5 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_00 execution + operator signoff.

End of `09_Implementation_Phases/Phase_00_Foundation.md` — 2026-04-30.
