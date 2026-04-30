# O01 — Container CI/CD Pipeline

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) (S02 — full chapter); [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §5 + §9; [`../07_Testing/01_Test_Matrix.md`](../07_Testing/01_Test_Matrix.md) §13 + §14a + §14b; [`../07_Testing/10_Full_Automation.md`](../07_Testing/10_Full_Automation.md); [`../01_Constitution.md`](../01_Constitution.md) §3 (R-05 + R-06), §11.5 (R-18).
> **Source line count:** chapter floor 600 lines per Master Plan §7.2 row O01.
> **Chapter targets:** R-05 (Containers), R-06 (every CI step in containers), R-10 (quality gates), R-18 (Operational Integrity).
> **Cross-links:** [`02_Quality_Gates_SonarQube_Snyk.md`](02_Quality_Gates_SonarQube_Snyk.md), [`04_Observability_and_Events.md`](04_Observability_and_Events.md), [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## Table of Contents

1. The Pipeline in Three Sentences
2. The Four-Mirror CI Topology
3. The Pipeline Stages — From PR to Deploy
4. Builder Image Discipline
5. Runtime Image Discipline (Distroless + Pinned Digests)
6. The Multi-Arch Build (linux/amd64 + linux/arm64)
7. Cosign Keyless Signing + SLSA L3 Provenance
8. SBOM Emission (cyclonedx-gomod + syft)
9. Local CI via `act` (R-06 "CI/CD is local, container-driven")
10. The GOCACHEPROG Remote Build Cache
11. Per-Mirror Registry Topology + Image Replication
12. Auto-Revert + Deployment Gate
13. The Renovate Configuration
14. Runner Pool Sizing & Cost Model
15. Open Questions
16. References & Anti-Bluff Verification

---

## 1. The Pipeline in Three Sentences

The HelixPlay container CI/CD pipeline runs on **four parallel mirror topologies** (GitHub Actions / GitLab CI / GitFlic CI / GitVerse CI), with every step executing **inside containers** built from the canonical builder images shipped under `vasic-digital/Containers`. Every PR triggers the [T01 §13 15-step lifecycle](../07_Testing/01_Test_Matrix.md#13-ci-lane-anatomy--from-pr-to-green) — container build → cosign sign → SLSA L3 provenance → ten-test-type matrix → required-checks gate → merge → composite-push to all four mirrors. The pipeline operationalises R-05 + R-06 (every service / build / test / scan in containers), R-10 (heavy quality + security scans), R-18 (no host-disruptive command can leak through any pipeline step).

---

## 2. The Four-Mirror CI Topology

Per [S04 §6.1](../06_Submodules/04_HelixQA_Integration.md#61-the-four-ci-runner-topologies), four parallel CI runner pools exist:

| Mirror     | CI runner system    | Runner location          | Concurrent lanes | Connectivity |
|------------|---------------------|--------------------------|-----------------:|--------------|
| GitHub     | GitHub Actions      | GitHub-hosted ubuntu-22.04 | 8              | Western      |
| GitLab     | GitLab CI           | GitLab.com shared runners | 8              | Western      |
| GitFlic    | GitFlic CI          | GitFlic-hosted runners    | 4              | Russian      |
| GitVerse   | GitVerse CI         | GitVerse-hosted runners    | 4              | Russian      |

Each pool runs the same workflow YAML (with dialect adaptation per [T10 §7c](../07_Testing/10_Full_Automation.md#7c-per-mirror-workflow-yaml-equivalence)). Parity between pools is verified nightly via `four-mirror-challenges-parity.sh` per [S03 §5.3](../06_Submodules/03_Challenges_Submodule.md#53-the-four-mirror-amplifier-challenges-edition).

The four-mirror amplifier means every PR's effective CI cost is **4×** a single-mirror pipeline; the §14 cost model accommodates this.

---

## 3. The Pipeline Stages — From PR to Deploy

### 3.1 Pre-flight (≤ 30 s)

1. Webhook fires on PR open / push.
2. `helix-test-select` reads the diff + identifies affected submodules.
3. The matrix is reduced to only the affected (submodule, test-type, mirror) cells.
4. Pinned-toolchain audit (`helix-toolchain-pin-audit`) confirms `go.mod` toolchain + Go-stdlib version match the fleet pin.

### 3.2 Container build (≤ 4 min warm; ≤ 12 min cold)

1. Pull the canonical builder image from the mirror's local registry (per [§11](#11-per-mirror-registry-topology--image-replication)).
2. Multi-stage Dockerfile per [S02 §10.1](../06_Submodules/02_Containers_Submodule.md#101-the-container-layer-cache):
   - Layer 1: dependency closure (cached unless go.sum changes).
   - Layer 2: source compile.
   - Layer 3: distroless runtime stage (immutable).
3. `docker buildx build --platform linux/amd64,linux/arm64 --provenance=mode=max --sbom=true`.
4. Push to all four mirror registries simultaneously.

### 3.3 Image signing + provenance (≤ 1 min)

Per [S02 §6](../06_Submodules/02_Containers_Submodule.md#6-image-signing--provenance-cosign-keyless--sigstore--slsa-l3):

1. `cosign sign --identity-token=$GITHUB_OIDC_TOKEN ghcr.io/vasic-digital/helix-${NAME}@${DIGEST}` — Sigstore-anchored, OIDC-bound.
2. `slsa-provenance-generate` produces SLSA L3 provenance.
3. `cosign attest --predicate=provenance.intoto.jsonl --type=slsaprovenance ${IMAGE}@${DIGEST}`.
4. cyclonedx-gomod emits the Go SBOM; syft emits the container-image SBOM.

### 3.4 Ten-test-type matrix (≤ 12 min warm)

The full T10 matrix (per [T10 §4](../07_Testing/10_Full_Automation.md#4-the-canonical-matrix-yaml)). The matrix is **fail-fast disabled**; every test-type runs even if a prior one fails — surfaces maximum issue-count per PR.

### 3.5 Required-checks gate (≤ 30 s)

The 14 (test-type, platform) × 4 (mirror) = 56 required checks must all be green. Branch-protection rules at all four mirrors enforce this.

### 3.6 Merge (≤ 30 s)

`gh pr merge --squash` (GitHub) / `glab mr merge --squash` (GitLab) etc. The merge fires a webhook → kicks off the post-merge nightly-cadence pre-flight on the next 02:00 Europe/Moscow tick.

### 3.7 Post-merge composite push (≤ 1 min)

`git push origin main` on GitHub triggers the composite push to all four mirrors per [Constitution §9](../01_Constitution.md#9-source-control--git-topology).

### 3.8 Optional deploy stage (canary + pre-release only)

Per [§12](#12-auto-revert--deployment-gate), the deploy stage is gated on the four-signal quad (cosign + Challenges + SBOM + visibility). Smoke probe + auto-revert close the loop.

---

## 4. Builder Image Discipline

Per [S02 §3.2](../06_Submodules/02_Containers_Submodule.md#32-the-four-builder--four-runtime-matrix), four canonical builder images:

| Builder image                      | Use case                                                                    |
|------------------------------------|------------------------------------------------------------------------------|
| `golang-builder`                   | Pure-Go submodules (no cgo): helix-r18-safeexec, helix-grpc-frame, etc.    |
| `golang-builder-cgo`               | cgo-linked: helix-iouring, helix-network, helix-codec, helix-encoder, etc. |
| `golang-builder-gpu`               | GPU + CUDA: helix-gpu-direct, helix-encoder, helix-thermal, helix-pipeline |
| `golang-builder-windows` + `golang-builder-darwin` | Cross-build for helix-capture variants                          |

Each builder image:

- `FROM golang:1.24-bookworm@sha256:...` (digest pinned).
- Pre-installs the §10 [T01 §10 pinned toolchain](../07_Testing/01_Test_Matrix.md#10-tooling-lockstep--pinned-versions-across-the-fleet) (govulncheck, snyk, cyclonedx-gomod, syft, trivy, cosign).
- Built reproducibly — same input → same digest.
- Published to all four mirror registries.
- Signed via cosign.

Drift between submodules using different builder versions is detected by `helix-builder-pin-audit` (Renovate-driven; PR opened on every builder release).

---

## 5. Runtime Image Discipline (Distroless + Pinned Digests)

Per [S02 §5](../06_Submodules/02_Containers_Submodule.md#5-base-image-discipline-distroless-pinned-digests-no-latest):

- Every runtime image is `FROM gcr.io/distroless/<variant>@sha256:<digest>` (digest pinned, never `:latest` or floating tag).
- Four variants: `distroless-static`, `distroless-cc`, `distroless-base`, `distroless-cuda`.
- UID 65532 (`nonroot`) — no privileged user.
- No shell, no package manager, no debugging tools.
- The `helix-image-pin-audit` Renovate config opens a PR on every distroless tag drift; operator reviews before merging.

The discipline reduces post-exploitation surface dramatically; lateral movement from a compromised process to a `bash` reverse shell is impossible because there is no `bash` in the image.

---

## 6. The Multi-Arch Build (linux/amd64 + linux/arm64)

Per [S02 §4](../06_Submodules/02_Containers_Submodule.md#4-multi-arch-image-strategy-linuxamd64--linuxarm64-default), every runtime image is published as a multi-arch manifest covering at minimum `linux/amd64` + `linux/arm64`.

```bash
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag ghcr.io/vasic-digital/helix-${NAME}:${TAG} \
    --tag registry.gitlab.com/vasic-digital/helix-${NAME}:${TAG} \
    --tag registry.gitflic.ru/vasic-digital/helix-${NAME}:${TAG} \
    --tag registry.gitverse.ru/vasic-digital/helix-${NAME}:${TAG} \
    --provenance=mode=max \
    --sbom=true \
    --push \
    -f vasic-digital/Containers/lanes/${LANE}/Dockerfile \
    .
```

A buildx run that pushes to three mirrors successfully but fails on the fourth fails the whole script (buildx exits non-zero on partial success); the operator must manually retry the failed mirror.

The `helix-capture` submodule is the documented exception per [S02 §4](../06_Submodules/02_Containers_Submodule.md#4-multi-arch-image-strategy-linuxamd64--linuxarm64-default) — it ships **four runtime variants** (linux / wayland / windows / darwin) because OS-specific capture APIs cannot share a single binary.

---

## 7. Cosign Keyless Signing + SLSA L3 Provenance

Per [S02 §6](../06_Submodules/02_Containers_Submodule.md#6-image-signing--provenance-cosign-keyless--sigstore--slsa-l3):

### 7.1 The signing flow

```bash
cosign sign \
    --identity-token=$GITHUB_OIDC_TOKEN \
    --rekor-url=https://rekor.sigstore.dev \
    ghcr.io/vasic-digital/helix-${NAME}@${DIGEST}

slsa-provenance-generate \
    --artifact-path=ghcr.io/vasic-digital/helix-${NAME}@${DIGEST} \
    --build-type=https://github.com/slsa-framework/slsa-github-generator/container@v1 \
    --output=provenance.intoto.jsonl

cosign attest \
    --predicate=provenance.intoto.jsonl \
    --type=slsaprovenance \
    ghcr.io/vasic-digital/helix-${NAME}@${DIGEST}
```

### 7.2 Operator-side verification

```bash
cosign verify \
    --certificate-identity-regexp='^https://github\.com/vasic-digital/.+/\.github/workflows/.+$' \
    --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
    ghcr.io/vasic-digital/helix-${NAME}@${DIGEST}

cosign verify-attestation \
    --type=slsaprovenance \
    --certificate-identity-regexp='^https://github\.com/vasic-digital/.+/\.github/workflows/.+$' \
    --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
    ghcr.io/vasic-digital/helix-${NAME}@${DIGEST}
```

The verify commands are baked into HelixQA's deployment-gate scripts; HelixQA refuses to deploy any image that fails either verification.

### 7.3 SLSA L3 substantiation

The four SLSA L3 conditions are met:

1. Source integrity — build is from a tagged commit (release-train tag).
2. Build integrity — build runs in a hardened ephemeral environment (GitHub Actions runners).
3. Provenance non-falsifiability — Sigstore keyless, OIDC-bound.
4. Isolated builds — each build runs in a fresh runner.

Recorded in `vasic-digital/Containers/SLSA-L3-CONFORMANCE.md` and audited annually.

---

## 8. SBOM Emission (cyclonedx-gomod + syft)

Per [S01 §4.4](../06_Submodules/01_Submodule_Catalog.md#44-sbom-generation-per-submodule-cyclonedx-gomod--syft-dual-format):

- `cyclonedx-gomod mod -licenses=true -output bom.cdx.json` — Go SBOM (CycloneDX 1.5 JSON), per-submodule release artefact.
- `syft packages docker:ghcr.io/vasic-digital/helix-${NAME}@${DIGEST} -o spdx-json=image.spdx.json` — container image SBOM (SPDX 2.3 JSON).

Both attach to the GitHub Release / GitLab Release page + the four-mirror registries. Both are signed via minisign (operator-rotated annually).

The dual-format is intentional: CycloneDX for the Go-aware view; SPDX for the OS-package view. Together they cover the supply-chain surface from code → dependencies → OS-base-image.

---

## 9. Local CI via `act` (R-06 "CI/CD is local, container-driven")

Per [S02 §9](../06_Submodules/02_Containers_Submodule.md#9-local-cicd-inside-containers-constitution-3-r-06):

```bash
# Run the per-PR matrix locally.
act push -j ten-test-types-matrix \
    --platform ubuntu-22.04=ghcr.io/vasic-digital/containers/golang-builder:1.24-bookworm \
    --workflows .github/workflows/ci.yml \
    --secret-file ~/.helix/local-ci-secrets.env \
    --reuse
```

The `--reuse` flag keeps the builder container alive between local runs (90 s cold → 10 s warm).

R-06 says "CI/CD is local, container-driven". `act` operationalises this — a developer's laptop runs the same containerised workflow as the cloud CI, against the same builder image, with the same workflow YAML. There is no "works on my laptop, breaks in CI" gap because the laptop *is* the CI.

---

## 10. The GOCACHEPROG Remote Build Cache

Per [S01 §4.6.2](../06_Submodules/01_Submodule_Catalog.md#462-gocacheprog-remote-build-cache):

- Go 1.24's `GOCACHEPROG` protocol enables a shared remote build cache.
- Production deployment uses Buildbarn (self-hosted) or BuildBuddy (SaaS).
- A 29-lane PR drops from 3.9 h cold-no-remote to ~ 25 min with remote cache enabled.
- The cache is hosted at a Russian-jurisdiction-compatible cloud provider with CDN-fronted endpoint reachable from all four mirror runners.

Operator decision (per OQ-S01-C in [S01 §10.2](../06_Submodules/01_Submodule_Catalog.md#102-remaining-deferred-to-future-families)) on Buildbarn vs BuildBuddy is documented in O02; the choice is operationally significant for ongoing CI cost.

---

## 11. Per-Mirror Registry Topology + Image Replication

Per [S02 §7](../06_Submodules/02_Containers_Submodule.md#7-container-registries-across-the-four-mirror-topology):

| Mirror     | Registry path                                       | Auth                                  | Visibility |
|------------|-----------------------------------------------------|---------------------------------------|------------|
| GitHub     | `ghcr.io/vasic-digital/<name>:<tag>`                | GitHub OIDC (workflow)                | public     |
| GitLab     | `registry.gitlab.com/vasic-digital/<name>:<tag>`    | GitLab Deploy Token                   | public     |
| GitFlic    | `registry.gitflic.ru/vasic-digital/<name>:<tag>`    | GitFlic API token                     | public     |
| GitVerse   | `registry.gitverse.ru/vasic-digital/<name>:<tag>`   | GitVerse PAT                          | public     |

Replication is **synchronous on push** (the buildx run pushes to all four). Drift between mirrors triggers `four-mirror-replication-audit.sh` nightly.

A Russian-jurisdiction operator pulling from a Russian-mirror registry sees the same image digest as a Western operator pulling from `ghcr.io`; the cosign signature is verifiable against either pull source.

---

## 12. Auto-Revert + Deployment Gate

Per [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad), the deployment-gate four-signal quad:

1. Cosign + SLSA L3 verify.
2. Challenges row green for the affected submodule.
3. Both SBOMs (cyclonedx + syft) emitted + attached to registry + signed.
4. Visibility audit passed (repository public on all four mirrors + licence policy clean).

Fail-closed posture: any sub-check error → `VerdictBlock`. Operator override via `helixqa-cli gate-override <digest> --reason "..."` (logged forever in run-archive + triggers P2 alert).

Smoke probe per [T09 §6](../07_Testing/09_Smoke.md#6-ci-lane-invocation-pattern) closes the loop:

```yaml
deploy-and-smoke:
  steps:
    - name: Deploy to canary
      run: helixqa-deploy --target=canary --image=$IMAGE_DIGEST
    - name: Smoke probe
      run: timeout 35s tests/smoke/probe.sh
    - name: Auto-revert on smoke fail
      if: failure()
      run: helixqa-deploy --target=canary --revert --reason="smoke-failure"
```

---

## 13. The Renovate Configuration

Per [S01 §4.3.3](../06_Submodules/01_Submodule_Catalog.md#433-renovate--dependabot-lockstep), a single Renovate config at `vasic-digital/.github/renovate.json5` opens per-submodule PRs on every direct-dependency bump. Excerpt:

```json5
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["config:base", "group:goCdkMonorepo"],
  "gomodTidy": true,
  "gomodUpdateImportPaths": true,
  "postUpdateOptions": ["gomodTidy", "gomodMassage"],
  "packageRules": [
    {
      "matchManagers": ["gomod"],
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "go-deps-minor-patch",
      "automerge": false
    },
    {
      "matchManagers": ["gomod"],
      "matchUpdateTypes": ["major"],
      "labels": ["dependency-major", "release-train"],
      "automerge": false,
      "schedule": ["before 6am on Monday"]
    },
    {
      "matchManagers": ["dockerfile"],
      "matchUpdateTypes": ["digest"],
      "groupName": "docker-base-images",
      "automerge": false
    }
  ]
}
```

Major-version bumps land only on Monday mornings (Europe/Moscow). Container-image digest bumps are grouped + opened for explicit operator review (no auto-merge — per [S02 §5](../06_Submodules/02_Containers_Submodule.md#5-base-image-discipline-distroless-pinned-digests-no-latest)).

---

## 14. Runner Pool Sizing & Cost Model

Per [T01 §14a](../07_Testing/01_Test_Matrix.md#14a-the-1160-cell-cost-model):

| Cadence       | Per-mirror wall-clock | Per-mirror compute | Cross-4-mirror compute |
|---------------|-----------------------|---------------------|-------------------------|
| Per-PR (1 submodule) | ~12 min (warm)  | 14 CPU-min          | 56 CPU-min              |
| Per-PR (29 submodules) | ~45 min       | 405 CPU-min         | 1,620 CPU-min            |
| Nightly       | ~2.5 h                 | 405 CPU-min         | 1,620 CPU-min            |
| Canary (24-h Stress dominates) | ~28 h        | ≈ 14 CPU-h (Stress) + ~ 7 CPU-h (rest) | ≈ 84 CPU-h |
| Pre-release   | ~46 h                  | ≈ 14 CPU-h (Stress) + ~ 145 CPU-h (30-day replay) | ≈ 636 CPU-h |

Operator's monthly compute budget should accommodate ~ 30 nightly + ~ 6 canary + ~ 2 pre-release = ~ 700 CPU-hours / month. With GOCACHEPROG remote cache, this drops by ≈ 40 %.

GitHub-hosted runners cost approximately $0.008 / CPU-min on `ubuntu-latest` — roughly $336 / month at the unrestricted-cache budget; ~ $200 / month with cache discipline. GitFlic + GitVerse are operator-provisioned (different cost model).

---

## 14a. Pipeline Shape per Cadence

The pipeline is one workflow file but the cadence dimension shapes which jobs run.

### 14a.1 Per-PR cadence shape

```yaml
on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  pre-flight:
    runs-on: ubuntu-22.04
    steps:
      - uses: actions/checkout@v4
      - id: select
        run: echo "submodules=$(./helix-test-select)" >> $GITHUB_OUTPUT
      - run: ./vasic-digital/Containers/scripts/toolchain-pin-audit.sh

  container-build:
    needs: pre-flight
    runs-on: ubuntu-22.04
    strategy:
      fail-fast: false
      matrix:
        platform: [linux/amd64, linux/arm64]
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - run: ./vasic-digital/Containers/scripts/multiarch-build.sh ${{ matrix.platform }}

  ten-test-types:
    needs: container-build
    # see T10 §4 canonical YAML
    ...

  required-checks:
    needs: ten-test-types
    runs-on: ubuntu-22.04
    steps:
      - run: ./vasic-digital/Containers/scripts/required-checks-gate.sh
```

### 14a.2 Nightly cadence shape (cron-driven)

```yaml
on:
  schedule:
    - cron: '0 23 * * *'   # 02:00 Europe/Moscow

jobs:
  full-matrix:
    # Same as per-PR but no test-select narrowing — full 1,160 cells.
    ...
  full-challenges-fan-out:
    # All 14 topologies × all primary scenarios
    ...
  visibility-audit:
    # gh + glab + GitFlic API + GitVerse manual
    ...
  licence-audit:
    # spdx-check + .snyk policy verification
    ...
  four-mirror-parity-audit:
    # four-mirror-challenges-parity.sh + four-mirror-replication-audit.sh
    ...
```

### 14a.3 Canary cadence shape (release-branch-driven)

```yaml
on:
  push:
    branches: ['release/v*']

jobs:
  canary-deploy:
    needs: full-matrix
    # Deploy to canary infra (operator-provisioned), then run T08 stress 1-h subset.
    ...
  full-14-topology-fan-out:
    # All 14 × all scenarios
    ...
```

### 14a.4 Pre-release cadence shape (tag-candidate-driven)

```yaml
on:
  push:
    tags: ['v*']

jobs:
  exhaustive-30-day-replay:
    # helixqa-replay across past 30 days of run-archive entries
    ...
  full-24h-stress:
    # the 24-hour soak (per T08 §6 + S03 11_stress_24h_steady_state topology)
    ...
  release-tag-publish:
    needs: [exhaustive-30-day-replay, full-24h-stress]
    if: success()
    # only publishes if every gate is green
    ...
```

The four-shape cadence catalogue is the operational source of truth; deviations require Constitution §15 amendment.

---

## 14b. Per-Mirror Registry-Credential Rotation

Per [S02 §6](../06_Submodules/02_Containers_Submodule.md#6-image-signing--provenance-cosign-keyless--sigstore--slsa-l3) + Constitution §11.5, registry credentials are rotated annually:

| Mirror     | Credential type          | Rotation cadence | Storage                              |
|------------|--------------------------|------------------|---------------------------------------|
| GitHub     | OIDC token (per-workflow)| n/a (ephemeral)  | n/a                                   |
| GitLab     | Deploy Token             | annual           | GitLab Vault, mounted as secret       |
| GitFlic    | API token                 | annual           | HashiCorp Vault, mounted as secret    |
| GitVerse   | PAT                       | annual           | HashiCorp Vault, mounted as secret    |

Rotation procedure (operator-only):

1. Issue new credential on the relevant mirror.
2. Update Vault path with the new credential.
3. Roll the workflow's environment variable to the new Vault path.
4. Verify with a probe push (a no-op container retag).
5. Revoke the old credential on the mirror.
6. Log the rotation in `vasic-digital/Containers/credential-rotation-log.md`.

A rotation that fails at step 5 (old credential lingers) is a P3 alert — the new credential is operational, but the old credential should be revoked promptly.

---

## 14c. The Build-Cache Hot-Cold Topology

The GOCACHEPROG remote cache (per §10) sits in front of GitHub's `actions/cache` action — a two-tier cache:

| Tier        | Location                                            | Lookup latency     | Hit rate (steady state) |
|-------------|-----------------------------------------------------|--------------------|--------------------------|
| Hot (L1)    | Runner-local `~/.cache/go-build` + `~/go/pkg/mod`  | < 1 ms             | ≈ 60 % across PRs        |
| Warm (L2)   | GitHub Actions cache (`actions/cache`, per-key)     | ≈ 100 ms           | ≈ 30 % across PRs        |
| Cold (L3)   | GOCACHEPROG remote cache (Buildbarn / BuildBuddy)   | ≈ 500 ms - 2 s    | ≈ 8 % across PRs         |
| Compile     | Go toolchain compiles from source                    | seconds-to-minutes | ≈ 2 %                    |

The hit-rate breakdown is nightly-measured; the §14 cost model assumes the steady-state distribution. A drop in L3 hit rate triggers a P3 alert (probable cache-corruption or auth issue).

---

## 14d. CI Concurrency Limits

Per-mirror runner pool concurrency limits the throughput:

| Mirror     | Concurrent jobs (effective) | Bottleneck                                 |
|------------|----------------------------:|--------------------------------------------|
| GitHub     | 8 lanes × 2 runners-per-lane = 16 | GitHub Actions plan limits                |
| GitLab     | 8 lanes × 1 runner-per-lane = 8 | GitLab.com shared-runners minute quota   |
| GitFlic    | 4 lanes × 1 runner = 4      | operator-provisioned runner pool size    |
| GitVerse   | 4 lanes × 1 runner = 4      | operator-provisioned runner pool size    |

A spike in PR rate beyond the concurrency limits queues the excess; the queue depth is monitored via Prometheus + alerts at queue-depth > 20 PRs (P3) / > 50 (P2). Operator can scale runner pools temporarily for release-cycle peaks.

---

## 14e. The Container Lane Reuse Pattern

Across the 29 submodules + 7 organisational images (4 builders + 4 runtimes — overlapping), only 11 container images are built. The submodule lanes (per [§4](#4-builder-image-discipline) + [§5](#5-runtime-image-discipline-distroless--pinned-digests)) consume these via `FROM ghcr.io/vasic-digital/containers/<image>@sha256:<digest>`.

| Image (organisational)            | Reuse count | Refresh cadence   |
|-----------------------------------|-------------|--------------------|
| `golang-builder:1.24-bookworm`    | ~ 12 lanes  | quarterly + on Go releases |
| `golang-builder-cgo`              | ~ 13 lanes  | same                |
| `golang-builder-gpu`              | ~ 6 lanes   | quarterly           |
| `golang-builder-windows`          | 1 lane (helix-capture) | annual    |
| `golang-builder-darwin`            | 1 lane (helix-capture) | annual    |
| `distroless-static`                | ~ 6 runtimes| Renovate-driven     |
| `distroless-cc`                    | ~ 14 runtimes| Renovate-driven    |
| `distroless-base`                  | ~ 1 runtime  | Renovate-driven     |
| `distroless-cuda`                  | ~ 6 runtimes| quarterly           |

Operator's CI/CD compute spend is dominated by the 29 per-submodule lanes; the 11 organisational images are amortised across the fleet.

## 15. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-O01-A            | Self-hosted Buildbarn vs SaaS BuildBuddy — operator's pick?                                                    | this chapter next revision after operator-side cost study |
| OQ-O01-B            | act parity with cloud-CI — full coverage or known gaps documented?                                             | T10 next revision                                   |
| OQ-O01-C            | Renovate cadence — once per dependency-update or batched weekly?                                              | this chapter next revision                          |
| OQ-O01-D            | Per-mirror runner pool sizing — operator-tunable or fixed?                                                     | this chapter next revision                          |

---

## 16. References & Anti-Bluff Verification

### 16.1 Internal

- [`00_Index.md`](00_Index.md), [`02_Quality_Gates_SonarQube_Snyk.md`](02_Quality_Gates_SonarQube_Snyk.md), [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md), [`06_Git_Topology_and_Push_Policy.md`](06_Git_Topology_and_Push_Policy.md).
- [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) — full chapter (S02).
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §5, §9.
- [`../07_Testing/01_Test_Matrix.md`](../07_Testing/01_Test_Matrix.md) §13, §14a, §14b.
- [`../07_Testing/10_Full_Automation.md`](../07_Testing/10_Full_Automation.md).
- [`../01_Constitution.md`](../01_Constitution.md) §3, §4, §7, §11.5.

### 16.2 External (web)

- GitHub Actions: https://docs.github.com/en/actions (accessed 2026-04-30).
- GitLab CI: https://docs.gitlab.com/ee/ci/ (accessed 2026-04-30).
- Sigstore cosign: https://docs.sigstore.dev/cosign/ (accessed 2026-04-30).
- SLSA framework v1.0: https://slsa.dev/spec/v1.0/ (accessed 2026-04-30).
- Distroless: https://github.com/GoogleContainerTools/distroless (accessed 2026-04-30).
- nektos/act: https://github.com/nektos/act (accessed 2026-04-30).
- Buildbarn: https://buildbarn.github.io/ (accessed 2026-04-30).
- BuildBuddy: https://www.buildbuddy.io/ (accessed 2026-04-30).
- Renovate: https://docs.renovatebot.com/ (accessed 2026-04-30).

### 16.3 Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) | 627    | 2026-04-30 | full S02 chapter — the architectural spec O01 deploys |
| [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) | 584    | 2026-04-30 | S04 §5 + §9 — cadence + deployment gate         |
| [`../07_Testing/01_Test_Matrix.md`](../07_Testing/01_Test_Matrix.md) | 648    | 2026-04-30 | T01 §13 lifecycle + §14a cost + §14b workflow |
| [`../07_Testing/10_Full_Automation.md`](../07_Testing/10_Full_Automation.md) | 307    | 2026-04-30 | T10 — workflow shape + composite actions       |

- Coverage: chapter exceeds the 600-line floor.
- Forbidden patterns: clean. Four §15 open questions named with deferred resolution chapters.
- The chapter operationalises S02's container architecture + S04's deployment gate + T10's workflow YAML at the actual-pipeline level (vs the architectural level S02/S04 specify).

### 16.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 as a single inline `Write` call (Master Plan §5.3 inline rule for the Operations family).
- Reviewed by: pending operator review.

End of `08_Operations/01_Container_CI_CD.md` — 2026-04-30.
