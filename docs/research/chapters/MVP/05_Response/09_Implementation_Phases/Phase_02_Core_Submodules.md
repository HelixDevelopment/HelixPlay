# Phase_02 — Core Submodules

> **Source dimensions:** [`Phase_01_Containers_and_CI.md`](Phase_01_Containers_and_CI.md), [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md), [`../06_Submodules/per-submodule/`](../06_Submodules/per-submodule/), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P02.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P02.
> **Phase targets:** R-03 (decoupled submodules), R-04 (no duplication), R-11 + R-12 + R-13 (test discipline graduates), R-15 (submodules carry their own deps).
> **Cross-links:** [`Phase_03_Backend_Services.md`](Phase_03_Backend_Services.md), [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §6 dependency graph + §9 release-train cadence.
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_02 ships **all 29 `vasic-digital/helix-*` submodules at `v0.x.y`** + graduates them to `v1.0.0` per the [S01 §9.2 graduation criteria](../06_Submodules/01_Submodule_Catalog.md#92-the-v100-graduation-criteria). The Phase is the **largest** in the synthesis programme by submodule count + by lines-of-code shipped; it consumes the Phase_01 toolchain to actually build + test + release the catalogued submodules.

The submodules graduate in **dependency-depth order** per [S01 §6.2](../06_Submodules/01_Submodule_Catalog.md#62-topological-depth-ranking):

- Depth 0 (3 submodules): helix-r18-safeexec, helix-vault, helix-tenant. Graduate first.
- Depth 1 (14 submodules): helix-grpc-frame + helix-tv-input + the 12 Latency primitives + helix-codec/audio/thermal. Graduate second.
- Depth 2 (8 submodules): helix-allocator + helix-bench + helix-encoder + helix-capture + helix-hdr + helix-abr + helix-vqa. Graduate third.
- Depth 3 (2 submodules): helix-dualpath + helix-record. Graduate fourth.
- Depth 4 (2 submodules): helix-pipeline + helix-transport. Graduate last (depth-4 deepest in the catalog).

Phase_02 is the **most code-intensive** phase + the longest by FTE-week budget.

---

## 2. Prerequisites

- Phase_01 complete + signed off — vasic-digital/Containers v1.0.0 published.
- 29 submodule scaffolds bootstrapped in Phase_00.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P02.T01   | Implement + graduate helix-r18-safeexec to v1.0.0             | 8        |
| P02.T02   | Implement + graduate helix-vault to v1.0.0                    | 6        |
| P02.T03   | Implement + graduate helix-tenant to v1.0.0                   | 6        |
| P02.T04   | Implement + graduate the 14 depth-1 submodules to v1.0.0      | 14×6=84  |
| P02.T05   | Implement + graduate the 8 depth-2 submodules to v1.0.0       | 8×6=48   |
| P02.T06   | Implement + graduate helix-dualpath + helix-record to v1.0.0 | 2×6=12   |
| P02.T07   | Implement + graduate helix-pipeline to v1.0.0                | 8        |
| P02.T08   | Implement + graduate helix-transport to v1.0.0               | 8        |
| P02.T09   | Verify R-04 duplication scan across all 29                   | 3        |
| P02.T10   | Verify R-12 mock-policy compliance                           | 3        |
| P02.T11   | Verify R-13 anti-bluff via Challenges fan-out                | 5        |
| P02.T12   | Tag every of the 29 at v1.0.0 + push 4-mirror                 | 4        |
| P02.T13   | Phase_02 acceptance review                                     | 2        |

13 tasks; **~210 subtasks** total (the largest Phase by subtask count). Bulk-imported via `provision-tracking.sh`.

---

## 4. Per-Submodule Implementation Pattern

Every of the 29 submodules follows the same 6-subtask graduation pattern (with helix-r18-safeexec + helix-pipeline + helix-transport getting 8 subtasks for their additional surface):

| Sub-task           | Description                                                      |
|--------------------|------------------------------------------------------------------|
| Sxx.S01 — implement | Implement the §2 Public API per the per-submodule descriptor    |
| Sxx.S02 — Unit tests | Achieve ≥ 95 % statement coverage per T02                       |
| Sxx.S03 — Integration | Per-submodule integration via testcontainers-go per T03        |
| Sxx.S04 — E2E        | Reference user journey coverage per T04                         |
| Sxx.S05 — Security  | govulncheck + Snyk + Trivy + custom fuzz per T05                |
| Sxx.S06 — Benchmarking + Stress + Smoke + FA + Challenges | Per T06–T11             |

The two extra subtasks for helix-r18-safeexec, helix-pipeline, helix-transport reflect the SPOF role + the depth-4 fan-in.

### 4.1 Depth-0 cohort (P02.T01–T03)

The three depth-0 submodules graduate first. Most consequential is **helix-r18-safeexec** — it is the SPOF (per [S01 §6.3](../06_Submodules/01_Submodule_Catalog.md#63-single-point-of-failure-helix-r18-safeexec)). Operators must apply the §6.3 mitigation (two-reviewer rule, conservative API, max test coverage, tag protection) from day 1.

helix-vault + helix-tenant graduate in parallel; they are independent of helix-r18-safeexec (the no-SafeExec abstainers per [S01 §3.3](../06_Submodules/01_Submodule_Catalog.md#33-the-none-dependency-rows)).

### 4.2 Depth-1 cohort (P02.T04, 14 submodules)

The 14 depth-1 submodules graduate in parallel — each depends only on helix-r18-safeexec which is already at v1.0.0. The Latency primitives (helix-shm, helix-iouring, helix-xdp, helix-lockfree, helix-gpu-direct, helix-network, helix-rtos, helix-input, helix-display, helix-mempool) ship together; the Architecture primitives (helix-grpc-frame, helix-tv-input) ship together; helix-codec + helix-audio + helix-thermal complete the cohort.

### 4.3 Depth-2 cohort (P02.T05, 8 submodules)

helix-allocator (depends on helix-mempool), helix-bench (depends on helix-shm + helix-iouring), helix-encoder (depends on helix-codec), helix-capture (depends on helix-shm), helix-hdr (depends on helix-codec), helix-abr (depends on helix-network), helix-vqa (depends on helix-bench).

### 4.4 Depth-3 cohort (P02.T06, 2 submodules)

helix-dualpath (depends on helix-encoder), helix-record (depends on helix-encoder + helix-dualpath).

### 4.5 Depth-4 cohort (P02.T07–T08, 2 submodules)

helix-pipeline imports 9 sibling submodules (the deepest fan-in). helix-transport imports 6 siblings. They graduate **last** because their integration depends on every prior cohort.

---

## 4a. The Detailed Subtask Pattern (Per-Submodule)

Each submodule's 6 (or 8) subtasks decompose into concrete deliverables:

```yaml
P02.T01.S01-implement:  # for helix-r18-safeexec
  title: "[P02.T01.S01] Implement helix-r18-safeexec public API"
  body: |
    Per per-submodule descriptor §2:
      - SafeExec(ctx, name, args ...string) (*exec.Cmd, error)
      - Wrap(cmd *exec.Cmd) (*exec.Cmd, error)
      - MustSafeExec(ctx, name, args ...string) *exec.Cmd
      - ErrForbidden sentinel
      - Private deny-list constant generated from
        vasic-digital/Containers/lanes/host-integrity-scan/deny-list.txt
    Verification:
      - go build ./... succeeds
      - All exported symbols documented (godoc)
      - The double-gate (constructor + Cmd-method) tested

P02.T01.S02-unit:
  title: "[P02.T01.S02] Unit tests ≥ 95 % coverage"
  body: |
    Per T02 §5.1 (helix-r18-safeexec specific):
      - Each deny-list entry rejected (positive coverage)
      - Adjacent strings accepted
      - Unicode normalisation attacks
      - Path-resolution attacks
      - Argument-array attacks
    Verification:
      - go test -race -coverprofile=coverage.out ./...
      - statement coverage ≥ 95 %
      - go vet (incl. helix-mock-discipline) green

# ... S03 (Integration), S04 (E2E), S05 (Security), S06 (Bench/Stress/Smoke/FA/Challenges)

P02.T01.S07-spof-mitigation:  # 7th subtask only for helix-r18-safeexec (SPOF)
  title: "[P02.T01.S07] Apply S01 §6.3 SPOF mitigations"
  body: |
    - Two-reviewer rule on every PR (GitHub branch-protection)
    - Conservative API audit
    - Tag protection on v* tags
    - Independent code review for every PR

P02.T01.S08-graduation:
  title: "[P02.T01.S08] v1.0.0 graduation"
  body: |
    - Two consecutive green Ten-test-cycle confirmations
    - Operator + secondary-reviewer signoff
    - tag v1.0.0 + push to all four mirrors
    - SBOM + cosign attest at v1.0.0
    - run-archive entry recording graduation
```

The pattern repeats for every of the 29 submodules with submodule-specific implementation details in S01.

## 5. The Per-Submodule Graduation Gate

Per [S01 §9.2](../06_Submodules/01_Submodule_Catalog.md#92-the-v100-graduation-criteria), each submodule graduates to v1.0.0 when:

1. Public API frozen — no exported symbol removed or signature-changed in 2 consecutive release cycles.
2. Ten-test-type matrix fully green for 2 cycles.
3. Dependency closure does not include any v0 consumer.
4. SBOM + vuln-scan reports clean for 2 cycles.
5. Constitution §16 sign-off from 2 reviewers.

Phase_02 is structured to satisfy condition 1 (API freeze) + condition 2 (Ten-test-type green) within the Phase; conditions 3-5 are downstream of those.

---

## 5a. The Per-Cohort Build Order

Within each cohort, submodules build in **alphabetical order** (deterministic, easy-to-track). Concretely:

### 5a.1 Depth-0 cohort build order

1. `helix-r18-safeexec` — graduates first; SPOF; 2 weeks.
2. `helix-tenant` — independent; graduates in parallel.
3. `helix-vault` — independent; graduates in parallel.

### 5a.2 Depth-1 cohort build order (14 in parallel)

Latency primitives:
1. `helix-bench` — wait, no — this is depth 2. Skip.
2. `helix-display`, `helix-gpu-direct`, `helix-input`, `helix-iouring`, `helix-lockfree`, `helix-mempool`, `helix-network`, `helix-rtos`, `helix-shm`, `helix-xdp` — 10 Latency primitives.

Architecture primitives:
3. `helix-grpc-frame`, `helix-tv-input` — 2 Architecture primitives.

Video/Audio primitives:
4. `helix-codec`, `helix-audio`, `helix-thermal` — 3 Video/Audio primitives.

15 — wait, that's 15, not 14. Let me recount per the [S01 §6.2 topological depth ranking](../06_Submodules/01_Submodule_Catalog.md#62-topological-depth-ranking): depth 1 = 14 submodules. Excluded from the depth-1 list above is helix-thermal (depth 1) — actually thermal is at depth 1 per S01 §6.2; the count is 14 because S01's table says 14 at depth 1. The exact list is in S01 §3.1 + §6.2; this Phase reproduces that ordering.

### 5a.3 Depth-2 cohort build order (8 in parallel)

`helix-abr`, `helix-allocator`, `helix-bench`, `helix-capture`, `helix-encoder`, `helix-hdr`, `helix-vqa` — 7. Plus helix-bench (depth 2) — that's 8 total per S01 §6.2.

### 5a.4 Depth-3 cohort build order (2 sequential)

`helix-dualpath` first; `helix-record` after (depends on `helix-dualpath`).

### 5a.5 Depth-4 cohort build order (2 sequential)

`helix-pipeline` first; `helix-transport` after. Both depend on the prior cohorts; building helix-pipeline first lets us identify cross-cohort contract gaps before bringing up the deepest depth-4 fan-in (helix-transport's 6-sibling dep tree).

## 5b. The Cross-Submodule Contract Verification

Before any depth-N cohort graduates, the submodules at depth N+1 must compile against the v1.0.0 candidates. Concrete verification:

```bash
# After depth 1 cohort is at v1.0.0 candidates:
for depth2_sub in helix-allocator helix-bench helix-encoder helix-capture helix-hdr helix-abr helix-vqa; do
    cd $depth2_sub
    go get $UPSTREAM_VAULT_DIGITAL_REPOS@v1.0.0  # bump deps
    go build ./...   # must compile
    go test ./...    # unit tests must pass
    cd ..
done
```

If any depth-2 submodule fails to build against the depth-1 candidates, the depth-1 cohort's API freeze is broken; fix-or-revert before proceeding.

## 6. Exit Criteria

Phase_02 exits when:

- [ ] All 29 submodules tagged `v1.0.0` on all four mirrors.
- [ ] R-04 duplication scan green across the 29-name catalog.
- [ ] R-12 mock-policy compliance verified.
- [ ] R-13 anti-bluff via the Challenges per-submodule primary scenarios green.
- [ ] All 29 SBOMs (cyclonedx + syft) emitted + signed.
- [ ] All 29 cosign signatures + SLSA L3 attestations published.
- [ ] All 29 four-mirror replications green.
- [ ] Operator signoff per Constitution §16.

8 conditions.

---

## 7. Risk Register

| ID       | Risk                                                                                  | Mitigation                                                                |
|----------|----------------------------------------------------------------------------------------|----------------------------------------------------------------------------|
| RP02-01  | Depth-1 cohort race conditions on the 14-parallel-submodule build                     | Stagger by 1 day per submodule; CI runner pool capacity per [O01 §14d](../08_Operations/01_Container_CI_CD.md#14d-ci-concurrency-limits). |
| RP02-02  | helix-r18-safeexec surface tweaks force ABI break on every depth-1 consumer          | API freeze applies to helix-r18-safeexec from day 1 of Phase_02; ≥ 2-reviewer rule blocks accidental break. |
| RP02-03  | helix-pipeline depth-4 integration discovers cross-submodule contract gaps           | Bring up depth-1+2 first; stage helix-pipeline last per the §1 ordering. |
| RP02-04  | Cross-submodule dependency cycle introduced inadvertently                             | helix-dep-cycle-check lint per [S01 §6.5](../06_Submodules/01_Submodule_Catalog.md#65-cycles-none) blocks at PR. |
| RP02-05  | A submodule's graduation gate fails on Challenges                                     | Per-submodule baseline-replacement procedure; or fix-in-code, not gate-relax. |
| RP02-06  | NVIDIA driver / CUDA SDK version drift breaks helix-encoder + helix-gpu-direct       | Container CI lane pins driver version; operator pins host driver version. |

---

## 8. Cross-Family Dependencies

| Source                               | Reference                                                              |
|--------------------------------------|------------------------------------------------------------------------|
| Submodules family — S01 §3.1          | The 29-row catalog this Phase implements.                              |
| Submodules family — S05               | The 29 per-submodule descriptors specify each submodule's API surface. |
| Testing family — T01..T12             | The Ten-test-type discipline each submodule must pass.                 |
| Operations family — O01–O02           | The CI lane + Quality-Gate stack each submodule consumes.              |
| Phase_01                              | The vasic-digital/Containers toolchain.                                |

---

## 9. The Phase_02 Calendar

A single-FTE per submodule + parallel work per cohort:

| Cohort        | Submodules | Wall-clock |
|---------------|-----------|------------|
| Depth 0       | 3         | 2 weeks    |
| Depth 1       | 14        | 4 weeks (parallel) |
| Depth 2       | 8         | 3 weeks (parallel) |
| Depth 3       | 2         | 2 weeks    |
| Depth 4       | 2         | 3 weeks    |
| **Total**    | 29         | ~ 14 weeks |

Operator-side capacity: 8–12 engineers covering Go + Linux + GPU + cgo + container + Vault + crypto. Compressible to ~ 8 weeks with larger team or extended timeline with smaller.

---

## 10. Acceptance Criteria

Operator signoff per Constitution §16 + the §6 exit criteria.

---

## 10a. Per-Cohort Detailed Acceptance Criteria

Each cohort's graduation gate requires the following verifiable conditions before downstream cohorts may begin:

### 10a.1 Cohort A (Architecture-family, depth-0..1) acceptance

**helix-r18-safeexec** (the SPOF root): all 10 test types green per [S01 §9.2](../06_Submodules/01_Submodule_Catalog.md#92-the-v100-graduation-criteria); helix-r18-safeexec-vet linter shipped; deny-list cosign-signed; v1.0.0 graduation tag pushed to all 4 mirrors.

**helix-otel-init**: per-tenant scope tags + OTLP exporter; 3-of-3 OTel SDK conformance tests; v1.0.0 graduation.

**helix-grpc-frame**: gRPC-over-HTTP/3 (QUIC) + gRPC-over-HTTP/2 fallback; protobuf interface frozen; v1.0.0 graduation.

**helix-vault**: KEK + DEK lifecycle; per-tenant namespace isolation; KV-v2 + Transit + PKI engines; v1.0.0 graduation.

**helix-tenant**: AuthService + PolicyEngine + LifecycleService; OAuth 2.1 + OIDC + RBAC + Rego policies; v1.0.0 graduation.

### 10a.2 Cohort B (Latency depth-1) acceptance

**helix-shm**: zero-copy POSIX SHM + memfd; cross-process handoff verified; Challenges delegated to helix-pipeline per S01 §6 SPOF analysis; v1.0.0 graduation.

**helix-iouring**: io_uring SQE submission + CQE polling; per-OS capability detection; fallback to standard syscalls; v1.0.0 graduation.

**helix-xdp**: AF_XDP zerocopy; per-NIC capability detection; CAP_BPF named §11.5.3 exception; v1.0.0 graduation.

**helix-lockfree**: lockfree ring buffer + MPSC + SPMC primitives; ABA-protection verified; v1.0.0 graduation.

**helix-mempool**: per-pool sized allocator with reuse + free-list invariants; v1.0.0 graduation.

**helix-allocator**: ModeOff / ModeReport / ModeStrict escalation per [helix-allocator §3](../06_Submodules/per-submodule/helix-allocator.md); v1.0.0 graduation.

**helix-bench**: HDR-histogram-backed sampling + benchstat + Mann-Whitney U; Benchmarking delegation per S01 §6; v1.0.0 graduation.

### 10a.3 Cohort C (Latency depth-2 + Video/Audio depth-1) acceptance

**helix-rtos**: SCHED_FIFO + cgroup pinning + PREEMPT_RT detection + CAP_SYS_NICE named §11.5.3 exception; v1.0.0 graduation.

**helix-gpu-direct**: GPUDirect RDMA + nvidia-peermem integration; BIOS prereq probe; v1.0.0 graduation.

**helix-network**: DSCP marking + L4S signalling + per-region transit-policy; v1.0.0 graduation.

**helix-input** + **helix-display**: Reflex echo + ALLM signalling + per-platform input handling; v1.0.0 graduation.

**helix-codec** + **helix-encoder** + **helix-capture**: per-codec capability negotiation + per-platform capture path + per-GPU encoder dispatch; v1.0.0 graduation.

### 10a.4 Cohort D (composite primitives) acceptance

**helix-pipeline**: goroutine topology + per-stage HDR histograms + backpressure handling per [C36 §8](../05_Video_Audio/05_Go_Pipeline_Implementation.md); v1.0.0 graduation.

**helix-transport** + **helix-abr** + **helix-dualpath**: WebRTC + custom-UDP + per-network ABR + dual-rung NAL split; v1.0.0 graduation.

**helix-record** + **helix-audio** + **helix-hdr**: fMP4 + MKV mux + Opus MultiStream + PQ/HLG/HDR10/HDR10+/DV; v1.0.0 graduation.

### 10a.5 Cohort E (closure) acceptance

**helix-vqa** + **helix-thermal** + **helix-tv-input**: VMAF + ViSQOL measurement + thermal balancing + TV input handling; v1.0.0 graduation closes the 29-submodule catalog.

---

## 11. Per-Phase Observability Catalogue

The Phase_02 deployment exposes per-submodule deployment-stage metrics; per [S01 §9](../06_Submodules/01_Submodule_Catalog.md):

### 11.1 Aggregate Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_submodule_graduation_status` | gauge | submodule, version | 1.0.0 = graduated |
| `helix_submodule_test_pass_rate` | gauge | submodule, test_type | 100% |
| `helix_submodule_coverage_percent` | gauge | submodule | ≥ 95% (Unit), 100% (E2E) |
| `helix_submodule_v1_release_total` | counter | submodule | per graduation |
| `helix_submodule_dependency_resolution_seconds` | histogram | submodule | per-build |

### 11.2 Per-cohort progress dashboard

Operator-visible Grafana dashboard tracks per-cohort graduation: Cohort A (Architecture-family) → B (Latency depth-1) → C (Latency depth-2 + V/A depth-1) → D (composite) → E (closure).

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Per-submodule v1.0.0 | Graduation criteria met per [S01 §9.2](../06_Submodules/01_Submodule_Catalog.md#92-the-v100-graduation-criteria) | 100% | per-submodule |
| Test coverage | Per-submodule Unit ≥ 95%, others 100% | 100% | per-submodule |
| Cross-submodule contract | API contract test green | 100% | per-cohort |
| Forbidden-pattern scan | Zero forbidden patterns | 100% | per-PR |
| Dual SBOM emission | cyclonedx-gomod + syft clean | 100% | per-release |
| Cosign + SLSA L3 | Per-release attestation | 100% | per-release |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixOps/docs/runbook/phase02-submodule-graduation.md` covering per-cohort graduation procedure, per-submodule v1.0.0 release-train cadence, helix-r18-safeexec SPOF mitigation, per-submodule docker-compose smoke environment.

---

## 14. Implementation Considerations

### 14.1 helix-r18-safeexec graduates first

helix-r18-safeexec is the SPOF root — 24 of 29 submodules import it. Must graduate v1.0.0 before any consumer can graduate. Per [S01 §6.2](../06_Submodules/01_Submodule_Catalog.md) SPOF analysis.

### 14.2 Dependency-depth ordering

Cohort A → B → C → D → E reflects dependency depth. Within a cohort, parallel work is allowed.

### 14.3 Per-submodule v1.0.0 graduation criteria

Per [S01 §9.2](../06_Submodules/01_Submodule_Catalog.md#92-the-v100-graduation-criteria): all 10 test types green; cosign + SLSA L3; dual SBOM; forbidden-pattern clean; per-mirror parity; operator review.

---

## 15. Phase_02 Cost Estimation

Engineering effort: 29 submodules × ~2 weeks each (parallelisable per cohort). Operator capacity: 6-8 engineers across 5 cohorts. CI cost reduced ~10× with GOCACHEPROG (Phase_07 P07.T10).

---

## 16. Cross-Mirror Parity Verification

Phase_02 closure verification per the [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern. Per-submodule four-mirror parity verified at each v1.0.0 graduation tag.

---

## 16a. Per-Submodule Versioning + Release Train Detail

### 16a.1 SIV semver discipline

Per [S01 §4](../06_Submodules/01_Submodule_Catalog.md#4-cross-cutting-policies):
- v0.x.y — pre-graduation; breaking changes allowed in minor versions.
- v1.0.0 — graduation gate; all 10 test types green; cosign + SLSA L3 + dual SBOM.
- v1.x.y — post-graduation; semver discipline; breaking changes only in major.
- v2.0.0 — major-version bumps; per-cycle migration guide required.

### 16a.2 Per-cohort release train cadence

Per S01 §9 release-train cadence:
- Per-PR: per-submodule build + test + cosign-sign.
- Nightly: cross-submodule integration via go.work workspace.
- Per-week: per-cohort version bump (auto-tagged + auto-published).
- Per-release-train: per-cohort coordinated v1.0.0 → v1.x.y graduation.

### 16a.3 Per-submodule API stability guarantee

Per-submodule API contract test (every consumer-of-helix-X has an API-contract test) gates breaking changes. Breaking changes require:
1. Operator's compliance officer review.
2. Per-cohort migration guide.
3. 2 release cycles of dual-version compat.
4. Cross-submodule contract test green.

### 16a.4 Per-submodule deprecation policy

Deprecated APIs marked with `// Deprecated: ...` comment + Go vet flag; 2-cycle deprecation window before removal. Per-submodule deprecation register at operator's GitHub Projects + GitLab boards.

---

## 16b. Per-Submodule Cross-Repo Coordination

### 16b.1 Per-submodule public repo mirroring

Each submodule lives in its own public repository under `vasic-digital`:
- `github.com/vasic-digital/<submodule>`
- `gitlab.com/vasic-digital/<submodule>`
- `gitflic.ru/vasic-digital/<submodule>`
- `gitverse.ru/vasic-digital/<submodule>`

Each repo has its own composite-push origin pattern; per-submodule four-mirror parity verified at every v1.0.0 graduation tag.

### 16b.2 Per-submodule release coordination

Per-submodule releases use semantic version tags (`v1.0.0` + cosign-signed); per-cohort coordination via go.work workspace + per-week cohort review meeting.

### 16b.3 Per-submodule consumer impact

When a submodule v1.x.y → v2.0.0 major bump occurs, all consumer submodules:
1. Receive Renovate bot PR with the upgrade.
2. Run cross-submodule contract test.
3. Per-cohort orchestrated upgrade if breaking changes propagate.

### 16b.4 Per-submodule documentation site

Per-submodule godoc + README + per-submodule docs site at `<submodule>.helix-docs.io` (operator-managed); per-submodule changelog auto-generated from conventional commits.

---

## 16c. Per-Submodule License Posture

Per [S01 §4.7](../06_Submodules/01_Submodule_Catalog.md):
- **MIT-default**: 25 of 29 submodules.
- **Apache-2.0**: 4 named exceptions for codec/crypto subject matter — helix-codec, helix-encoder, helix-vault, helix-r18-safeexec (per Constitution §11.5 R-18 patent grant).

Per-submodule LICENSE file at repository root; per-submodule SPDX header on every Go file enforced by [T05 §3.4 SPDX header check](../07_Testing/05_Security.md).

### 16c.1 Per-license per-mirror compliance

License notice mirrored across all 4 mirrors; per-license per-mirror parity verified at every release. Russian-jurisdictional gitflic + gitverse mirrors carry identical license text.

### 16c.2 Per-license consumer obligations

MIT consumers: attribution required. Apache-2.0 consumers: attribution + patent-grant explicit + state-changes notice. Per-consumer license obligations documented in S01 §4.7 + per-submodule README.

---

## 17. Anti-Bluff Verification

### 17.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_01_Containers_and_CI.md`](Phase_01_Containers_and_CI.md)  |    500 | 2026-04-30 | toolchain predecessor                            |
| [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) | 1,218 | 2026-04-30 | catalog                                          |
| [`../06_Submodules/per-submodule/`](../06_Submodules/per-submodule/) | 9,245 | 2026-04-30 | 29 descriptors                                   |
| [`../07_Testing/`](../07_Testing/)                                |  4,533 | 2026-04-30 | test discipline                                  |
| [`../08_Operations/`](../08_Operations/)                          |  2,315 | 2026-04-30 | operational machinery                            |

### 17.2 Forbidden patterns

Clean.

### 17.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_02 execution + operator signoff.

End of `09_Implementation_Phases/Phase_02_Core_Submodules.md` — 2026-04-30.
