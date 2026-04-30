# Phase_01 — Containers & CI

> **Source dimensions:** [`Phase_00_Foundation.md`](Phase_00_Foundation.md), [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md), [`../08_Operations/01_Container_CI_CD.md`](../08_Operations/01_Container_CI_CD.md), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P01.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P01.
> **Phase targets:** R-05 + R-06 + R-10 — first phase to ship code under `vasic-digital/Containers`.
> **Cross-links:** [`Phase_02_Core_Submodules.md`](Phase_02_Core_Submodules.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_01 ships **`vasic-digital/Containers` at v1.0.0** — the canonical builder + runtime images + per-submodule CI lane fragments + admission policies that every subsequent Phase consumes. The Phase delivers no `helix-*` submodule code (Phase_02 does that); its work is the **toolchain** the submodules will use.

After Phase_01, the 29 submodule scaffolds bootstrapped in Phase_00 can build images via `vasic-digital/Containers/builders/golang-builder@sha256:...`; the per-submodule CI workflows can include `vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml`; the deployment gate can verify cosign + SLSA L3 against published canonical builders.

Phase_01 is **infrastructure code**, not application code. The 11 organisational images (4 builders + 4 runtimes + the host-integrity-scan lane + Renovate config) are the deliverables.

---

## 2. Prerequisites

- Phase_00 complete + signed off.
- 4 CI runner pools online + healthy.
- Operator signing keys + Sigstore + minisign anchors published.
- Vault + observability stack reachable from the runners.

A failed Phase_00 prerequisite blocks Phase_01 — there is no shortcut.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks | Depends on  |
|------------|---------------------------------------------------------------|---------:|-------------|
| P01.T01   | Build canonical builder images (golang-builder + variants)    | 5        | P00.T05     |
| P01.T02   | Build canonical runtime images (4 distroless variants)        | 4        | P01.T01     |
| P01.T03   | Author per-submodule CI lane templates (29 lanes)             | 6        | P01.T01     |
| P01.T04   | Author the canonical workflow YAML fragments                  | 4        | P01.T03     |
| P01.T05   | Author CUE admission policies (R-18 hazards + nonroot + distroless) | 4    | P01.T01     |
| P01.T06   | Author cosign + SLSA L3 + SBOM scripts                        | 5        | P00.T03     |
| P01.T07   | Author 4-mirror multiarch-build + replication scripts          | 4        | P01.T01     |
| P01.T08   | Author the host-integrity-scan harness (deny-list + strace + auditd) | 5  | P01.T01 + P01.T05 |
| P01.T09   | Author Renovate configuration                                 | 3        | P01.T01     |
| P01.T10   | Configure four-mirror replication                              | 3        | P01.T07     |
| P01.T11   | Establish helix-allocator-vet + helix-test-matrix-lint        | 4        | P01.T03     |
| P01.T12   | Bring vasic-digital/Containers to v1.0.0 graduation           | 3        | all of T01–T11 |
| P01.T13   | Phase_01 acceptance review                                     | 2        | P01.T12     |

13 tasks, ~50 subtasks total. Each is `[P01.Tyy.Szz]`-tagged + ticket-mirrored to GitHub Projects + GitLab.

---

## 4. Task Details

### 4.1 P01.T01 — Canonical builder images

Per [O01 §4](../08_Operations/01_Container_CI_CD.md#4-builder-image-discipline):

- **golang-builder** — `FROM golang:1.24-bookworm` + pinned tooling.
- **golang-builder-cgo** — `FROM golang-builder` + gcc + linux-headers + libv4l-dev.
- **golang-builder-gpu** — `FROM golang-builder-cgo` + cuda-toolkit + nvenc-headers + Level-Zero.
- **golang-builder-windows** — `FROM mcr.microsoft.com/windows/servercore:ltsc2022` + Go.
- **golang-builder-darwin** — `FROM mcr.microsoft.com/dotnet/sdk:8.0` + osxcross.

Each is built reproducibly + cosign-signed + published to all four mirror registries.

### 4.2 P01.T02 — Canonical runtime images

Per [O01 §5](../08_Operations/01_Container_CI_CD.md#5-runtime-image-discipline-distroless--pinned-digests):

- **distroless-static** — `FROM gcr.io/distroless/static-debian12:nonroot@sha256:...`.
- **distroless-cc** — `FROM gcr.io/distroless/cc-debian12:nonroot@sha256:...`.
- **distroless-base** — `FROM gcr.io/distroless/base-debian12:nonroot@sha256:...`.
- **distroless-cuda** — `FROM gcr.io/distroless/cc-debian12:nonroot@sha256:...` + CUDA 12.6 runtime libs.

Digest pins are committed to `vasic-digital/Containers/builders/PINNED-DIGESTS.md`; Renovate watches upstream tags + opens PRs on drift.

### 4.3 P01.T03 — 29 per-submodule CI lanes

Per [S02 §3.3](../06_Submodules/02_Containers_Submodule.md#33-lane-directory-contract), each lane directory contains the `Dockerfile`, `ci.yml`, `tests/` symlink to host-integrity-scan, `bom-config.json`, `snyk-policy.json`, `release-train.yml`. The 29 lane directories are populated from canonical templates + per-submodule customisation.

### 4.4 P01.T04 — Canonical workflow YAML fragments

Per [O01 §3 + T10 §4](../07_Testing/10_Full_Automation.md#4-the-canonical-matrix-yaml):

- `go-submodule-1.x.yml` — the canonical workflow.
- `ten-test-types-matrix.yml` — the matrix orchestrator.
- `sbom-emission.yml` — cyclonedx + syft + cosign.
- `visibility-audit.yml` + `licence-audit.yml` + `dep-cycle-check.yml` + `release-train.yml`.

Fragments are versioned with `vasic-digital/Containers`'s SemVer; updates are a release-train event.

### 4.5 P01.T05 — CUE admission policies

Per [S02 §8](../06_Submodules/02_Containers_Submodule.md#8-container-runtime-hazards-vs-r-18-constitution-§115_3):

- `policies/distroless-required.cue`
- `policies/nonroot-required.cue`
- `policies/no-privileged.cue`
- `policies/no-pid-host.cue`
- `policies/no-network-host.cue`
- `policies/seccomp-profile-required.cue`

Each is enforced by Kyverno (k8s) or `vasic-digital/Containers/scripts/policy-enforce.sh` (plain Docker/Podman).

### 4.6 P01.T06 — Cosign + SLSA L3 + SBOM scripts

- `scripts/cosign-keyless-sign.sh`
- `scripts/slsa-provenance-generate.sh`
- `scripts/sbom-emit.sh`
- `scripts/sbom-verify.sh`
- `scripts/four-mirror-signature-parity.sh`

### 4.7 P01.T07 — Multi-arch + 4-mirror scripts

- `scripts/multiarch-build.sh` — buildx wrapper for amd64+arm64 + 4-mirror push.
- `scripts/four-mirror-replication-audit.sh` — nightly cross-mirror digest verification.
- `scripts/registry-credential-rotate.sh` — annual rotation procedure per [O01 §14b](../08_Operations/01_Container_CI_CD.md#14b-per-mirror-registry-credential-rotation).

### 4.8 P01.T08 — host-integrity-scan harness

Per [S02 §3.4](../06_Submodules/02_Containers_Submodule.md#34-the-host-integrity-scan-shared-lane):

- `lanes/host-integrity-scan/Dockerfile` (FROM golang-builder + strace + auditd).
- `lanes/host-integrity-scan/deny-list.txt` (canonical R-18 deny-list).
- `lanes/host-integrity-scan/strace-allowlist.txt`.
- `lanes/host-integrity-scan/audited-syscalls.txt`.
- `lanes/host-integrity-scan/tests/{deny-list-rejects,strace-coverage,auditd-zero-events}.sh`.

The deny-list is the **single source of truth** for R-18 enforcement across the fleet; updates require Constitution §15 amendment.

### 4.9 P01.T09 — Renovate configuration

Per [S01 §4.3.3](../06_Submodules/01_Submodule_Catalog.md#433-renovate--dependabot-lockstep) + [O01 §13](../08_Operations/01_Container_CI_CD.md#13-the-renovate-configuration):

- `vasic-digital/.github/renovate.json5` — org-level config.
- Per-submodule overrides as needed.
- Monday-mornings major-bumps + grouped minor/patch.

### 4.10 P01.T10 — 4-mirror replication

Configure per-mirror webhook to trigger replication audit on every `vasic-digital/Containers` push. Verify at the §9 exit criteria.

### 4.11 P01.T11 — helix-allocator-vet + helix-test-matrix-lint

These two tools are critical for subsequent Phases:

- `helix-allocator-vet` — go vet plugin enforcing hot-path zero-alloc per [helix-allocator descriptor](../06_Submodules/per-submodule/helix-allocator.md).
- `helix-test-matrix-lint` — verifies T01 §2 cell coverage.

Both ship as Phase_01 artefacts so Phase_02 submodules can consume them.

### 4.12 P01.T12 — vasic-digital/Containers v1.0.0 graduation

Per [S01 §9.2](../06_Submodules/01_Submodule_Catalog.md#92-the-v100-graduation-criteria), `vasic-digital/Containers` graduates when:

1. All 4 builder images + 4 runtime images at `v1.0.0` digests.
2. All 11 organisational artefacts published + signed.
3. All 6 CUE policies + 5 cosign/SLSA/SBOM scripts + 4 multiarch scripts + the host-integrity-scan harness shipped.
4. The Ten-test-type matrix runs green for the Containers repo's own tests (yes, the meta-test).
5. Operator sign-off.

### 4.13 P01.T13 — Phase_01 acceptance review

Operator runs the §6 exit criteria + signs off.

---

### 4.13 The Detailed Subtask Listing for P01.T01

```yaml
P01.T01.S01:
  title: "Build golang-builder@1.24-bookworm with pinned tooling"
  body: |
    Dockerfile builds FROM golang:1.24-bookworm
    Pre-installs: govulncheck@latest, snyk-cli ≥v1.1300, cyclonedx-gomod ≥v1.6.0,
    syft ≥v1.16.0, trivy ≥v0.55.0, cosign ≥v2.4.0, helix-bench-cli, helix-allocator-vet.
    Output: ghcr.io/vasic-digital/containers/golang-builder@sha256:...
    Acceptance: cosign verify succeeds; image runs `go version` returning 1.24.x.
P01.T01.S02:
  title: "Build golang-builder-cgo with libv4l-dev + linux-headers"
  body: |
    FROM golang-builder + apt-get install gcc libc6-dev linux-headers libv4l-dev libpipewire-0.3-dev
    For helix-iouring, helix-network, helix-codec, helix-encoder cgo paths.
    Output: ghcr.io/vasic-digital/containers/golang-builder-cgo@sha256:...
P01.T01.S03:
  title: "Build golang-builder-gpu with CUDA 12.6 toolkit + NVENC/CUDA headers"
  body: |
    FROM golang-builder-cgo + NVIDIA CUDA 12.6 toolkit + nvenc-headers + Level-Zero loader
    For helix-encoder NVENC, helix-gpu-direct CUDA, helix-thermal NVML, helix-pipeline GPU stages.
    Output: ghcr.io/vasic-digital/containers/golang-builder-gpu@sha256:...
P01.T01.S04:
  title: "Build golang-builder-windows with Windows Server Core base + Go cross"
  body: |
    For helix-capture's DXGI variant.
P01.T01.S05:
  title: "Build golang-builder-darwin with osxcross"
  body: |
    For helix-capture's Metal variant.
```

The 5 subtasks are sequential — depth-1 builders depend on golang-builder. Each subtask's acceptance is `cosign verify` + `docker run <image> <expected-tool> --version` + `helix-spdx-check ./...` on a probe Dockerfile.

### 4.14 The Per-Lane Verification Probe

For each of the 29 per-submodule lanes built in P01.T03, the operator verifies:

```bash
# For each lane:
LANE=shared-memory-1.x
cd vasic-digital/Containers/lanes/$LANE
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile -t test:$LANE .
docker run --rm test:$LANE --version  # if the lane bundles a probe binary
cosign verify ghcr.io/vasic-digital/containers/lanes/$LANE@${DIGEST}
```

Each lane's probe runs in CI on every push to vasic-digital/Containers; a probe failure blocks the merge.

### 4.15 The First Probe PR

Once Phase_01 deliverables are in place, the operator opens a probe PR in vasic-digital/Containers itself:

1. Make a no-op change to the `golang-builder` Dockerfile (e.g. update a comment).
2. Push to a feature branch.
3. Watch the PR's CI lane invoke the full Ten-test-type matrix on all four mirrors.
4. Verify the matrix returns 14 × 4 = 56 green checks.
5. Verify the deployment-gate four-signal quad is green.
6. Merge + verify the four-mirror replication audit confirms the new digest on all four registries.

A successful probe PR is the gate to closing Phase_01.

### 4.16 The Phase_01 Deliverable Manifest

| Artifact                                              | Type                | Location                                                                |
|-------------------------------------------------------|---------------------|-------------------------------------------------------------------------|
| 5 builder images                                      | Docker image        | ghcr.io/vasic-digital/containers/<name>@sha256:...                      |
| 4 runtime images                                       | Docker image        | ghcr.io/vasic-digital/containers/<name>@sha256:...                      |
| 29 per-submodule lane directories                      | Source tree          | vasic-digital/Containers/lanes/                                         |
| 7 workflow YAML fragments                              | YAML                 | vasic-digital/Containers/ci-fragments/                                  |
| 6 CUE admission policies                               | CUE                  | vasic-digital/Containers/policies/                                      |
| 5 cosign + SLSA + SBOM scripts                         | Bash                 | vasic-digital/Containers/scripts/                                       |
| 4 multiarch + replication scripts                       | Bash                 | vasic-digital/Containers/scripts/                                       |
| host-integrity-scan harness                             | Source + fixtures    | vasic-digital/Containers/lanes/host-integrity-scan/                     |
| Renovate config                                         | JSON5                | vasic-digital/.github/renovate.json5                                    |
| helix-allocator-vet binary                              | Go binary             | ghcr.io/vasic-digital/containers/tools/helix-allocator-vet@sha256:...    |
| helix-test-matrix-lint binary                           | Go binary             | ghcr.io/vasic-digital/containers/tools/helix-test-matrix-lint@sha256:...|

11 categories of artefact; ~100 individual artefacts in total.

## 5. Subtask Catalogue (Aggregate)

50 subtasks across 13 tasks. Bulk-imported via `provision-tracking.sh` per [O05 §9c](../08_Operations/05_Tracking_GitHub_GitLab.md#9c-the-operators-bulk-provisioning-one-time-workflow).

---

## 6. Exit Criteria

Phase_01 exits when:

- [ ] 4 builder images + 4 runtime images at `v1.0.0` digests.
- [ ] 29 per-submodule CI lane directories scaffolded with templates.
- [ ] 7 canonical workflow YAML fragments published.
- [ ] 6 CUE admission policies enforced.
- [ ] cosign + SLSA L3 + SBOM scripts working on a probe image.
- [ ] Multi-arch + 4-mirror push works for a probe image.
- [ ] host-integrity-scan harness rejects every deny-list entry on a probe.
- [ ] Renovate is open-PR-ing on a synthetic dependency change.
- [ ] 4-mirror replication audit returns zero divergence.
- [ ] helix-allocator-vet + helix-test-matrix-lint shipped + working.
- [ ] vasic-digital/Containers tagged `v1.0.0` on all four mirrors.
- [ ] Operator sign-off per Constitution §16.

12 conditions; all must be green.

---

## 7. Risk Register

| ID         | Risk                                                                                  | Mitigation                                                              |
|------------|----------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| RP01-01    | Distroless upstream digest drift mid-Phase causing reproducibility break                | §4.2 PINNED-DIGESTS file + Renovate watch.                               |
| RP01-02    | Cross-builder Windows / Darwin failing on operator hardware                            | T01.S04 + S05 are operator-environment-dependent; defer if not viable.   |
| RP01-03    | host-integrity-scan harness false-positives on a legitimate test fixture               | §4.8's strace-allowlist.txt is operator-tunable per [S02 §3.4](../06_Submodules/02_Containers_Submodule.md#34-the-host-integrity-scan-shared-lane). |
| RP01-04    | Sigstore Rekor outage during Phase_01 testing                                         | Per OQ-S02-B fail-closed; operator override path.                       |
| RP01-05    | CUE policy false-positive blocking legitimate container behaviour                      | §4.5 policies are reviewed pre-merge + tested against probe containers.|
| RP01-06    | 4-mirror replication delay on push                                                     | §4.10 audit catches > 5-min delay; alert P3.                            |

---

## 8. Cross-Family Dependencies

| Source                               | Reference                                                              |
|--------------------------------------|------------------------------------------------------------------------|
| Submodules family — S02 (Containers) | The architectural specification this Phase implements.                |
| Operations family — O01 (CI/CD)      | The operational integration this Phase enables.                       |
| Testing family — T05 / T10           | The Security + Full Automation lanes that consume Phase_01 outputs.    |

---

## 9. Acceptance Criteria

Operator signoff per Constitution §16 + the §6 exit criteria checklist.

---

## 10. The Phase_01 Calendar

~ 4 weeks. Operator-side capacity: 4 engineers (2× containers/CI engineers; 1× security engineer for cosign + SLSA; 1× ops engineer for runner deployment).

---

## 11. Per-Phase Observability Catalogue

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_containers_build_seconds` | histogram | submodule, target | per-target wall-time |
| `helix_containers_publish_seconds` | histogram | submodule, mirror | p99 ≤ 30 s |
| `helix_cosign_sign_total` | counter | submodule, key | per-build |
| `helix_slsa_provenance_emitted_total` | counter | submodule | 100% |
| `helix_sbom_emitted_total` | counter | submodule, generator | both cyclonedx-gomod + syft |
| `helix_runner_pool_size` | gauge | mirror, arch | per-arch capacity |
| `helix_runner_active_count` | gauge | mirror | < 90% saturation |
| `helix_pinned_digest_drift_total` | counter | submodule, base | 0 |

### 11.2 Grafana dashboards

- **Per-submodule Build Pipeline** — build wall-time + cosign sign + SBOM emission per submodule.
- **Per-mirror Runner Health** — runner pool capacity + saturation + per-arch breakdown.
- **Cosign + SLSA + SBOM Coverage** — 100% coverage audit per release.

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Container build wall-time | Per-submodule build time with cache-warm | p99 ≤ 5 min | per-PR |
| Multi-arch publish | amd64 + arm64 + arm64-v8 publication | 100% | per-release |
| Cosign + SLSA L3 coverage | Per-release attestation | 100% | per-release |
| Dual SBOM emission | cyclonedx-gomod + syft | 100% | per-release |
| Pinned digest integrity | Base image digests pinned + verified | 100% | per-build |
| Runner pool utilisation | Saturation per arch | < 90% | 7-day rolling |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixContainers/docs/runbook/phase01-cicd.md` covering Containers v1.0.0 release-train procedure, per-mirror runner deployment, cosign keyless + Fulcio short-lived cert workflow, SLSA L3 build-environment isolation, dual SBOM verification, pinned-digest update procedure.

---

## 14. Implementation Considerations

### 14.1 Cosign keyless vs key-based

Cosign keyless (Sigstore + Fulcio short-lived certs) is canonical default. Key-based mode for operator-air-gapped + government-jurisdiction deployments.

### 14.2 Dual SBOM (cyclonedx-gomod + syft)

Both SBOM generators run; outputs cross-validated. cyclonedx-gomod is Go-specific (more accurate for Go modules); syft is universal (covers OS-level deps). Per [O01 §6](../08_Operations/01_Container_CI_CD.md).

### 14.3 GOCACHEPROG vs Bazel remote-cache

GOCACHEPROG selected (operator-self-hosted, simpler than Bazel BES). Per [O01 §10](../08_Operations/01_Container_CI_CD.md). Cuts 29-lane PR from 3.9 h to ~25 min.

### 14.4 Multi-arch publishing

amd64 + arm64 + arm64-v8 (Apple Silicon). QEMU emulation for cross-arch builds; native arm64 runners preferred for build wall-time.

---

## 15. Phase_01 Cost Estimation

Per-mirror runner pool: 4 mirrors × ~$200 / month = ~$800 / month operator-side. GOCACHEPROG MinIO storage: ~$50 / month. Total Phase_01 operational cost: ~$850 / month for 29-lane CI across 4 mirrors.

---

## 16. Cross-Mirror Parity Verification

Phase_01 closure verification per the [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern. Per-mirror image registry parity verified via cosign verify across all 4 mirrors.

---

## 17. Anti-Bluff Verification

### 17.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_00_Foundation.md`](Phase_00_Foundation.md)                |    800 | 2026-04-30 | predecessor Phase                                |
| [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) |    627 | 2026-04-30 | architectural specification                     |
| [`../08_Operations/01_Container_CI_CD.md`](../08_Operations/01_Container_CI_CD.md) |    606 | 2026-04-30 | operational specification                       |
| [`../07_Testing/10_Full_Automation.md`](../07_Testing/10_Full_Automation.md) |    307 | 2026-04-30 | workflow YAML consumer                           |

### 17.2 Forbidden patterns

Clean.

### 17.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_01 execution + operator signoff.

End of `09_Implementation_Phases/Phase_01_Containers_and_CI.md` — 2026-04-30.
