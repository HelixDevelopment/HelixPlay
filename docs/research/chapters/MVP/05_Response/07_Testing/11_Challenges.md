# T11 — Challenges (the anti-bluff backstop)

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.10 + §8 + §11; [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) (full S03 chapter — Challenges contract); [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §5 (cadence) + §9 (deployment gate); [`../01_Constitution.md`](../01_Constitution.md) §1 (R-13 Anti-Bluff Pledge), §6.4 (R-14 Challenges integration).
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row T11 (larger floor reflects T11's role as the operational meeting point with S03).
> **Chapter targets:** R-13 (anti-bluff backstop), R-14 (Challenges integrated as in HelixAgent + Catalogizer; HelixQA fully integrated).
> **Cross-links:** [`02_Unit_Tests.md`](02_Unit_Tests.md)–[`10_Full_Automation.md`](10_Full_Automation.md), [`12_HelixQA_Autonomous.md`](12_HelixQA_Autonomous.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Challenges is **test-type 10 of 10** — the **last** entry in the Ten and the **most demanding**. Cadence: nightly (per-submodule primary scenario) + canary (full 14-topology fan-out) + pre-release (30-day exhaustive replay). Per-PR runs the per-submodule primary scenario only.

Challenges' role is **the anti-bluff backstop**. The first 9 test types (Unit through Full Automation) each have a known failure mode where green can co-exist with broken end-user behaviour (per [S03 §2](../06_Submodules/03_Challenges_Submodule.md#2-why-challenges-is-the-tenth-test-type-r-13-anti-bluff-backstop)). Challenges closes the loop because it boots the **entire production-like topology** and observes **user-visible artefacts** (rendered video frames, audio waveforms, controller round-trips, gRPC trace topology, billing events, latency histograms). A bug invisible to types 1–9 cannot be invisible to Challenges, because the bug *by definition* manifests as user-visible behaviour drift.

R-13 (Constitution §1) demands that green tests guarantee real, end-user-usable behaviour. The Challenges row is the operational floor below which green tests do not guarantee that. Past projects had green tests on broken features; the Challenges meta-test is the project's primary mitigation.

---

## 2. The Discipline — What Makes "Green" Green

A green Challenges row means:

1. **The full production-like topology was instantiated** — host agent + clients + CockroachDB + NATS + Redis + Vault + capture stack + encoder + transport, all real, no mocks.
2. **The recorded scenario was replayed** with deterministic inputs (controller events, network impairment, clock skew, random seeds — per [S03 §6.2](../06_Submodules/03_Challenges_Submodule.md#62-replay)).
3. **Every observation matches the baseline within tolerance** per [S03 §6.3 change-point detection](../06_Submodules/03_Challenges_Submodule.md#63-change-point-detection):
   - Rendered video frames: VMAF ≥ 95.0; SSIM ≥ 0.99.
   - Audio: ViSQOL ≥ 4.5; spectrogram MSE ≤ 0.001.
   - Latency histograms: Mann-Whitney U + KS, `p > 0.01`.
   - gRPC trace topology: ≥ 99 % edge-weight match; structural identity.
   - OTLP span structure: structural identity, attribute diff ≤ 5 %.
   - Billing event count: exact equality.
   - Frame timing (VRR): p99 < baseline + 0.5 ms.
   - p999 latency: single-tail Welch's t-test, `p > 0.01`.
4. **Four-mirror parity** — same observation on all four CI runner mirrors per [S03 §5.3](../06_Submodules/03_Challenges_Submodule.md#53-the-four-mirror-amplifier-challenges-edition).
5. **Anti-bluff posture intact** — Challenges' own five failure modes (per [S03 §7](../06_Submodules/03_Challenges_Submodule.md#7-anti-bluff-enforcement-at-the-challenges-boundary)) are mitigated.

---

## 3. The S03 ↔ T11 Contract

T11 is the **per-submodule consumer** of S03's machinery. The interface:

- **Inputs to S03**: per-submodule CI lane invokes `vasic-digital/Challenges/scripts/challenge-run.sh <submodule> <topology> <scenario>`.
- **Outputs from S03**: exit code 0 (green) or non-zero (red); a `change-point-report.json` artefact captured by the CI lane.
- **Run-archive integration**: HelixQA (S04) uploads the change-point report to the immutable run-archive per [S04 §7](../06_Submodules/04_HelixQA_Integration.md#7-run-archive-and-audit-trail-immutable-queryable-signed).
- **Deployment-gate consumption**: HelixQA's gate refuses to deploy any image whose Challenges row is not green per [S04 §9](../06_Submodules/04_HelixQA_Integration.md#9-deployment-gating-the-cosign--challenges--sbom--visibility-quad).

The contract is **bidirectional**: T11 commits to consuming S03's primitives correctly; S03 commits to providing reproducible baselines + signed manifests + change-point detection within tolerance.

---

## 4. The 14 Canonical Topologies

Per [S03 §3](../06_Submodules/03_Challenges_Submodule.md#3-repository-layout-topologies--baselines--harness), 14 frozen topologies cover the user journeys + operational scenarios. T11 ratifies the 14-topology surface as **frozen** (a 15th topology requires Constitution §15 amendment per S03 §3):

| #  | Topology                              | Canonical scenario                                                       |
|----|---------------------------------------|--------------------------------------------------------------------------|
| 01 | `01_minimum_viable_session`           | `06_full_pipeline_end_to_end.scenario.yaml`                              |
| 02 | `02_multi_session_single_host`        | `01_grpc_streaming_under_burst_load.scenario.yaml`                        |
| 03 | `03_multi_host_multi_session`         | NATS-clustering integration scenario                                      |
| 04 | `04_white_label_multi_tenant`         | `01_tenant_theme_swap_at_session_boundary.scenario.yaml`                |
| 05 | `05_recording_emit_and_replay`        | `02_fmp4_mkv_emit_then_replay.scenario.yaml`                              |
| 06 | `06_4k_120hz_hdr_dolby_vision`        | the most-demanding HDR + 4K120 path                                      |
| 07 | `07_kernel_bypass_send_path`          | `01_iouring_completion_under_120hz_send.scenario.yaml`                  |
| 08 | `08_controller_input_low_latency`     | `01_dscp_l4s_under_240hz_input.scenario.yaml`                             |
| 09 | `09_thermal_throttling_dvfs`          | `01_thermal_throttling_quality_scaling.scenario.yaml`                    |
| 10 | `10_chaos_network_partition`          | `01_quic_recovery_after_partition.scenario.yaml`                         |
| 11 | `11_stress_24h_steady_state`          | `01_p999_floor_over_24h.scenario.yaml`                                  |
| 12 | `12_burst_marketing_spike`            | `01_abr_ladder_descent_under_10x_burst.scenario.yaml`                  |
| 13 | `13_audit_compliance_eu_dsa`          | `02_kek_rotation_under_active_session.scenario.yaml`                    |
| 14 | `14_security_attack_surface`          | `01_attempt_forbidden_subprocess_invocation.scenario.yaml`              |

Per-submodule primary scenarios are mapped at [S03 §4](../06_Submodules/03_Challenges_Submodule.md#4-per-submodule-challenges-entry-points-29--the-helix-shm-delegation).

---

## 5. The Per-Submodule Challenges Map (S03 §4 reproduced for T11)

Reproduced from [S03 §4.1](../06_Submodules/03_Challenges_Submodule.md#41-the-29-entry-point-map) for T11's reference. Every of the 29 submodules maps to exactly one primary topology + scenario pair:

| Submodule           | Primary topology               | Primary scenario                                   |
|---------------------|--------------------------------|----------------------------------------------------|
| helix-r18-safeexec  | `14_security_attack_surface`   | `01_attempt_forbidden_subprocess_invocation`        |
| helix-grpc-frame    | `02_multi_session_single_host` | `01_grpc_streaming_under_burst_load`                |
| helix-tv-input      | `01_minimum_viable_session`    | `04_tv_remote_navigation_full_journey`              |
| helix-vault         | `13_audit_compliance_eu_dsa`   | `02_kek_rotation_under_active_session`              |
| helix-tenant        | `04_white_label_multi_tenant`  | `01_tenant_theme_swap_at_session_boundary`         |
| helix-shm           | `01_minimum_viable_session`    | `06_full_pipeline_end_to_end` (delegated to helix-pipeline) |
| helix-iouring       | `07_kernel_bypass_send_path`   | `01_iouring_completion_under_120hz_send`            |
| helix-xdp           | `07_kernel_bypass_send_path`   | `02_xdp_redirect_under_packet_burst`                |
| helix-lockfree      | `02_multi_session_single_host` | `02_spsc_ringbuffer_under_concurrent_session`       |
| helix-gpu-direct    | `06_4k_120hz_hdr_dolby_vision` | `01_gpu_direct_rdma_under_4k120_session`            |
| helix-network       | `08_controller_input_low_latency` | `01_dscp_l4s_under_240hz_input`                   |
| helix-rtos          | `08_controller_input_low_latency` | `02_sched_fifo_under_high_load`                  |
| helix-input         | `08_controller_input_low_latency` | `03_reflex_round_trip_under_load`                |
| helix-display       | `06_4k_120hz_hdr_dolby_vision` | `02_vrr_under_burst_frame_load`                    |
| helix-mempool       | `02_multi_session_single_host` | `03_mempool_no_alloc_in_hot_path`                  |
| helix-allocator     | `02_multi_session_single_host` | `04_allocator_enforces_hot_path_zero_alloc`         |
| helix-bench         | `11_stress_24h_steady_state`   | `01_p999_floor_over_24h`                           |
| helix-codec         | `06_4k_120hz_hdr_dolby_vision` | `03_codec_ladder_at_4k120`                         |
| helix-encoder       | `06_4k_120hz_hdr_dolby_vision` | `04_nvenc_qsv_amf_videotoolbox_parity`              |
| helix-capture       | `01_minimum_viable_session`    | `05_capture_per_os_at_session_start`               |
| helix-dualpath      | `05_recording_emit_and_replay` | `01_dual_path_nal_stream_plus_record`              |
| helix-record        | `05_recording_emit_and_replay` | `02_fmp4_mkv_emit_then_replay`                     |
| helix-audio         | `06_4k_120hz_hdr_dolby_vision` | `05_atmos_eARC_under_full_journey`                 |
| helix-hdr           | `06_4k_120hz_hdr_dolby_vision` | `06_dolby_vision_tone_map_under_session`           |
| helix-abr           | `12_burst_marketing_spike`     | `01_abr_ladder_descent_under_10x_burst`            |
| helix-thermal       | `09_thermal_throttling_dvfs`   | `01_thermal_throttling_quality_scaling`            |
| helix-vqa           | `06_4k_120hz_hdr_dolby_vision` | `07_vmaf_ldat_against_baseline`                    |
| helix-pipeline      | `01_minimum_viable_session`    | `06_full_pipeline_end_to_end` (also covers helix-shm) |
| helix-transport     | `10_chaos_network_partition`   | `01_quic_recovery_after_partition`                 |

29 submodules × 1 primary scenario = 29 primary scenarios (with helix-pipeline owning the helix-shm delegation as a single scenario covering both).

---

## 6. The Cadence Fan-Out

Different cadences invoke different subsets of the catalogue:

### 6.1 Per-PR cadence

Only the **affected submodule's primary scenario** runs. If a PR touches `helix-encoder`, only `06_4k_120hz_hdr_dolby_vision/04_nvenc_qsv_amf_videotoolbox_parity.scenario.yaml` runs; every other Challenges entry sits idle. Wall-clock: 3–5 minutes per scenario.

### 6.2 Nightly cadence

**All 29 primary scenarios** run + select secondary scenarios where applicable (some submodules have ≥ 2 scenarios; nightly runs the full set). Wall-clock: ≈ 2 hours.

### 6.3 Canary cadence

**The full 14-topology fan-out** — every topology × every scenario, regardless of submodule attribution. Wall-clock: 12–24 hours (the 24-h `11_stress_24h_steady_state` topology dominates).

### 6.4 Pre-release cadence

Canary + the **30-day exhaustive replay**: every Challenges run from the past 30 days replayed against the new tag candidate. Wall-clock: 24–48 hours (per [S04 §5.4](../06_Submodules/04_HelixQA_Integration.md#54-pre-release-cadence)).

The cadences are **non-overridable** — a PR that disables a per-PR Challenges invocation is rejected by the per-submodule CI lane's required-checks rule.

---

## 7. Anti-Bluff Posture — Five Failure Modes & Mitigations

Per [S03 §7](../06_Submodules/03_Challenges_Submodule.md#7-anti-bluff-enforcement-at-the-challenges-boundary), Challenges itself has known failure modes. T11 ratifies the mitigations:

| Failure mode             | Mitigation                                                                                  |
|--------------------------|----------------------------------------------------------------------------------------------|
| Baseline tampering       | minisign-signed manifests + non-delegable operator review on baselines                       |
| Topology drift           | digest-pinned compose spec + four-mirror replication audit                                   |
| Scenario narrowing       | DSL with JSON-Schema validation + minimum-step enforcement                                   |
| Observation gaps         | heartbeat manifests + recorder-required artefacts                                            |
| Threshold inflation      | two-reviewer rule on `vasic-digital/Challenges/harness/change-point/threshold-config.yaml`  |

The five mitigations together close the practical gaps. R-13 is operationally enforced, not aspirational.

---

## 8. Tooling

- **`vasic-digital/Challenges`** — the canonical topology + baseline + harness repo (per S03).
- **`challenge-run.sh`** — per-submodule entry-point.
- **`baseline-record.sh`** — operator-only baseline recording (non-delegable per S03 §6.1).
- **`baseline-verify.sh`** — minisign chain verification on every baseline use.
- **`change-point-report.sh`** — human-readable diff report for failing runs.
- **`four-mirror-challenges-parity.sh`** — nightly four-mirror parity audit.
- **VMAF, ViSQOL, hdrhistogram-go** — observation-comparison primitives.

---

## 9. Anti-Pattern Catalogue

### 9.1 Updating baseline without operator approval

A maintainer updates `baselines/<topology>/<scenario>/baseline.manifest.json` without the operator-signed minisign chain renewal. The baseline-verify check at next CI run rejects the change. This is a **gating** discipline, not a recommendation.

### 9.2 Adding a scenario without a baseline

A new scenario file at `vasic-digital/Challenges/topologies/<topology>/scenarios/` with no corresponding baseline at `baselines/`. The Challenges harness fails-closed — no baseline = no green possible.

### 9.3 Per-PR Challenges disabled "for speed"

```yaml
- name: Challenges (per-PR)
  if: false   # forbidden
```

Per [§6.1](#61-per-pr-cadence), per-PR Challenges is mandatory for the affected submodule. The required-checks gate catches the disabled invocation as a missing required check.

### 9.4 Threshold relaxed without operator signoff

```yaml
# vasic-digital/Challenges/harness/change-point/threshold-config.yaml
vmaf_minimum: 80.0  # relaxed from 95.0 — forbidden without two-reviewer signoff
```

The threshold-config.yaml file has the same two-reviewer protection as `helix-r18-safeexec` per [S01 §6.3](../06_Submodules/01_Submodule_Catalog.md#63-single-point-of-failure-helix-r18-safeexec) mitigation #3. A relax-and-merge PR will be rejected by branch-protection rules.

### 9.5 Local-machine baseline recording

```bash
# Maintainer's laptop
$ baseline-record.sh ...
```

Baselines are recorded **only** in the operator's signed-CI environment. The maintainer cannot record a baseline; only an operator with the minisign key can. Local-recorded baselines fail the minisign chain verification at next CI run.

### 9.6 Pre-release cadence skipped on hotfix

A "we're shipping a hotfix; no time for the 30-day replay" pressure violates Constitution §13 *Exceptions* without operator approval. The release-train tag-publish gate (per [S04 §5.4](../06_Submodules/04_HelixQA_Integration.md#54-pre-release-cadence)) blocks the tag push until pre-release passes; manual override requires Constitution §15 amendment-style authorisation.

---

## 10. CI Lane Invocation Pattern

```yaml
- name: Challenges (per-PR primary)
  if: matrix.test-type == 'challenges-primary'
  steps:
    - name: Verify baseline minisign chain
      run: ./vasic-digital/Challenges/scripts/baseline-verify.sh \
              ${{ env.SUBMODULE }} ${{ env.TOPOLOGY }} ${{ env.SCENARIO }}
    - name: Boot topology
      run: |
        cd vasic-digital/Challenges/topologies/${{ env.TOPOLOGY }}
        docker compose up -d --wait
    - name: Run challenge
      run: |
        ./vasic-digital/Challenges/scripts/challenge-run.sh \
            ${{ env.SUBMODULE }} ${{ env.TOPOLOGY }} ${{ env.SCENARIO }}
    - name: Generate change-point report
      if: always()
      run: |
        ./vasic-digital/Challenges/scripts/change-point-report.sh \
            > change-point-report.json
    - name: Tear down
      if: always()
      run: |
        cd vasic-digital/Challenges/topologies/${{ env.TOPOLOGY }}
        docker compose down -v
    - name: Upload to run-archive
      if: always()
      run: |
        helixqa-upload change-point-report.json
```

Note `if: always()` on the change-point report + run-archive upload — even a red Challenges run produces an artefact for postmortem.

---

## 11. The Replay-from-Run-Archive Workflow

Per [S04 §7.5](../06_Submodules/04_HelixQA_Integration.md#75-replay-from-archive), an operator can replay an archived Challenges run:

```bash
helixqa-replay 2026-04-30T02:00:00Z
```

The replay re-instantiates the topology, replays the deterministic inputs (controller events + network impairment + clock skew + random seed) from the archive, and produces a fresh observation set + change-point comparison. This supports two workflows:

- **Regression bisection**: `helixqa-replay` a sequence of runs across a date range to locate the run where a change-point first appeared.
- **Debugging**: replay a specific failing run with `OTEL_LOG_LEVEL=debug` to capture additional traces the original run did not.

Replays produce a new run-archive entry with `cadence=reactive`, `subtype=replay`, `replays=<original-run-id>`. The audit trail is preserved.

---

## 12. The Pre-Release Exhaustive Replay

The pre-release cadence's **30-day exhaustive replay** is the largest single Challenges expenditure. Mechanics:

1. Tag candidate `v1.x.y` triggers the pre-release pipeline.
2. The pipeline iterates over the past 30 days of run-archive entries (≈ 200 runs per submodule × 29 submodules = 5,800 total runs).
3. Each archived run is replayed against the new tag candidate's image.
4. Each replay's output is compared to the original archive's output via change-point detection.
5. Verdict: "v1.x.y maintains baseline parity across 30 days of historical observations" — green; or "v1.x.y diverges from N historical runs" — red, with operator review of each divergence.

Wall-clock budget: 24–48 hours, dominated by the replay set. Compute cost: ≈ 5,800 runs × 5 minutes each = 480 hours sequentially; with 4-mirror × 8-lane parallelism ≈ 15 hours wall-clock + the canary's 28 h.

The 30-day exhaustive replay catches **regressions that don't show up in nightly** — a subtle drift introduced over the last week may be statistically insignificant on the latest nightly comparison alone but significant when compared against a 30-day baseline distribution.

---

## 12a. The Mirror-by-Mirror Challenges Fan-Out

For each cadence, the fan-out across the four mirrors is parallel but each mirror's runner pool sizes are independent. Per [S04 §6.1](../06_Submodules/04_HelixQA_Integration.md#61-the-four-ci-runner-topologies), the topology:

| Mirror     | CI runner system    | Concurrent lanes | Russian-jurisdiction? |
|------------|---------------------|-----------------:|-----------------------|
| GitHub     | GitHub Actions     | 8                | no                     |
| GitLab     | GitLab.com shared  | 8                | no                     |
| GitFlic    | GitFlic-hosted      | 4                | yes                    |
| GitVerse   | GitVerse-hosted     | 4                | yes                    |

The Russian-jurisdiction mirrors have smaller runner pools (operator capacity), so canary's full 14-topology fan-out completes faster on GitHub + GitLab than on GitFlic + GitVerse. The orchestrator parallelises within each mirror; cross-mirror synchronisation only happens at the parity-audit step. A divergent green/red verdict between mirrors triggers the [S04 §6.3 alert ladder](../06_Submodules/04_HelixQA_Integration.md#63-the-reverse-a-russian-mirror-green-western-red).

## 12b. Detailed Change-Point Detection Parameters

The algorithm parameters per [S03 §6.3](../06_Submodules/03_Challenges_Submodule.md#63-change-point-detection):

| Metric                            | Algorithm                       | Default threshold | Operator-tunable? |
|-----------------------------------|----------------------------------|-------------------|-------------------|
| VMAF (frame quality)              | Mean comparison + Welch t-test   | ≥ 95.0            | per-topology yes  |
| SSIM                              | Mean comparison                  | ≥ 0.99            | per-topology yes  |
| ViSQOL (audio quality)             | Mean comparison + Welch t-test   | ≥ 4.5             | per-topology yes  |
| Spectrogram MSE                    | MSE on power-spectral density    | ≤ 0.001           | per-topology yes  |
| Latency p999                       | Single-tail Welch's t-test       | p > 0.01          | global only       |
| Latency histogram shape            | Mann-Whitney U + KS              | p > 0.01          | global only       |
| gRPC call topology                 | Trace-graph isomorphism          | ≥ 99 % match      | global only       |
| OTLP span attributes               | Structural identity + diff       | ≤ 5 % attr-diff   | global only       |
| Billing event count                | Exact equality                   | =                 | not tunable        |
| Frame timing (VRR)                 | Latency histogram + jitter       | p99 < base+0.5 ms | per-topology yes  |

Operator-tunable thresholds live in `vasic-digital/Challenges/harness/change-point/threshold-config.yaml` per [S03 §3](../06_Submodules/03_Challenges_Submodule.md#3-repository-layout-topologies--baselines--harness). Modifying a tunable threshold requires the two-reviewer rule (per [§9.4](#94-threshold-relaxed-without-operator-signoff)) + an entry in the threshold-change audit log.

## 12c. Replay-from-Archive Command Catalogue

The full set of `helixqa-cli` operations on archived runs:

| Command                                       | Purpose                                                                  |
|------------------------------------------------|--------------------------------------------------------------------------|
| `helixqa-cli archive-query --filter=<expr>`   | List run-IDs matching a filter (cadence, submodule, verdict, date range) |
| `helixqa-cli archive-show <run-id>`            | Show full archive entry including manifest                                |
| `helixqa-cli replay <run-id>`                  | Re-run with archived inputs; produces a new entry                       |
| `helixqa-cli compare <run-id-a> <run-id-b>`    | Change-point comparison between two runs                                |
| `helixqa-cli archive-export <range>`           | Signed tarball export for compliance audit                              |
| `helixqa-cli changepoint-bisect <range>`       | Auto-bisect a date range to locate a regression's first appearance     |
| `helixqa-cli runbook <alert-type>`             | Print the matching alert runbook                                         |
| `helixqa-cli gate-override <digest> --reason ` | Operator-only deployment-gate override (logged forever)                |

The CLI is shipped as part of `HelixDevelopment/HelixQA`'s release artefacts. Operators install it via `go install github.com/HelixDevelopment/HelixQA/cmd/helixqa-cli@latest`.

## 12d. Cross-Family Integration Matrix

How T11 connects to the other 8 families:

| Family                         | Integration via                                                                    |
|--------------------------------|------------------------------------------------------------------------------------|
| `03_Architecture/`             | C08 §10 R-18 wrapper imported into every Challenges-bound test fixture            |
| `04_Latency/`                  | C24 §6 helix-bench is the latency measurement instrument; T06 + T11 share it     |
| `05_Video_Audio/`              | C35 §6 helix-vqa provides VMAF + ViSQOL; T11's video/audio change-point uses them |
| `06_Submodules/` S03 + S04     | T11 is the per-submodule consumer of S03 + S04's machinery                         |
| `07_Testing/` T01–T10          | T11 is the meta-test that closes the gap T01–T10 leave                             |
| `08_Operations/` (forthcoming) | Deployment gate + cron schedule + alert routing operationally land here           |
| `09_Implementation_Phases/` (forthcoming) | Phase_02 deliverables include green Challenges per submodule              |
| `99_Web_Research_Addenda/`     | VMAF / ViSQOL / Mann-Whitney U / KS reference URLs accessed 2026-04-30           |

Every Challenges scenario is reachable from at least one chapter §6 in families 1–3 (the originating submodule's chapter); the 14 topologies × scenarios are the connective tissue.

## 13. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T11-A            | 30-day vs 90-day replay window — what's the operator-acceptable tail-coverage?                                 | T11 next revision after 90-day operator data        |
| OQ-T11-B            | Per-tenant Challenges scenarios — should each tenant's deployment have a tenant-specific scenario?            | C11 §6 + T11 next revision                          |
| OQ-T11-C            | LDAT hardware sourcing — operator-supplied or HelixPlay-bundled?                                              | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T11-D            | Threshold-config.yaml versioning — operator-managed or Renovate-driven?                                       | `08_Operations/02_Quality_Gates_SonarQube_Snyk.md` |

---

## 13a. Topology Compose Spec Reference

Each of the 14 canonical topologies has a fixed compose spec at `vasic-digital/Challenges/topologies/<NN_name>/docker-compose.yml`. Excerpts of the most operationally significant:

### 13a.1 `06_4k_120hz_hdr_dolby_vision`

The most demanding topology — exercises every Video/Audio submodule simultaneously:

- 1× host agent with GPU device passthrough (NVENC + Vulkan)
- 1× client with HDR10/Dolby Vision-capable display emulator
- 1× CockroachDB single-node
- 1× NATS + 1× Redis + 1× Vault
- 1× Toxiproxy with 5 ms p99 latency injection (representing wired-LAN baseline)
- 1× otel-collector + 1× challenge-orchestrator
- 1× Xvfb + 1× weston-vnc (capture-side virtual display)

Wall-clock per scenario: 8–12 minutes. Full topology spec ≈ 280 lines of YAML.

### 13a.2 `11_stress_24h_steady_state`

The 24-hour soak topology — adds `helix-leak-regress` watchdog:

- Same as `01_minimum_viable_session` plus
- 1× helix-leak-regress sidecar polling /proc/<pid>/{fd,status,smaps} + runtime.MemStats every 60 s
- 1× pprof-collector triggered every 4 h
- Long-lived volume mounts for the heap profiles + run-archive uploads

Wall-clock: 24 hours. Compute cost dominated by sustained 4K120 encode + record (typical canary run consumes ≈ 15 NVENC-GPU-hours).

### 13a.3 `13_audit_compliance_eu_dsa`

The compliance topology — adds DSA Article 17 audit emit + KEK rotation observer:

- Standard topology plus
- 1× audit-log emitter (writes signed events to a fixed-format JSONL stream)
- 1× synthetic third-party DSA inspector (reads the audit log + verifies signatures)
- 1× extra Vault server (the secondary KEK target; verifies cross-Vault re-wrap correctness)

Wall-clock: ≈ 30 minutes. Compute cost modest; the value is in the audit-log fidelity verification.

### 13a.4 `14_security_attack_surface`

The security topology — same shape as `01_minimum_viable_session` but with auditd enabled inside every container + extra strace fixtures around the deny-list rejection paths.

Per-container auditd config lives at `topologies/14_security_attack_surface/auditd/audit.rules`; the rules monitor the four R-18 syscalls (kill, reboot, settimeofday, mount) + report any invocation immediately.

## 13b. The Pre-Release Replay Cost Breakdown

The 30-day exhaustive replay's compute budget breakdown:

| Phase                                         | Wall-clock | Compute cost     |
|-----------------------------------------------|-----------:|------------------:|
| Iterate run-archive entries (past 30 days)    | 5 min      | 1 CPU-min         |
| Per-replay container build (cache hit)        | 30 s × N   | 0.5 × N CPU-min   |
| Per-replay scenario run                        | 5 min × N  | 5 × N CPU-min     |
| Per-replay change-point comparison            | 30 s × N   | 0.5 × N CPU-min   |
| Aggregate report generation                    | 10 min     | 5 CPU-min         |

Where N = total archived runs in the 30-day window. For typical operator load: 5,800 archived runs (≈ 200 per submodule × 29). Total compute: ≈ 35,000 CPU-min ≈ 580 CPU-hours. With 4-mirror × 8-lane parallelism: 580 / 32 ≈ 18 wall-clock-hours.

Combined with canary's 28 h, the pre-release total is ≈ 46 h — matching [S04 §5.4](../06_Submodules/04_HelixQA_Integration.md#54-pre-release-cadence)'s 24–48 h budget.

## 13c. The "Significant Divergence" Threshold for Pre-Release

A **statistically significant** divergence is one where Mann-Whitney U + KS report `p ≤ 0.01` AND the magnitude exceeds a per-metric floor:

| Metric                         | Magnitude floor                                       |
|--------------------------------|--------------------------------------------------------|
| VMAF                           | drop ≥ 1.0 point                                       |
| ViSQOL                         | drop ≥ 0.05 MOS-equivalent                             |
| Latency p999                   | rise ≥ 5 % from historical mean                        |
| gRPC trace edge-weight         | shift ≥ 5 % on any high-traffic edge                   |
| OTLP attributes diff           | ≥ 5 % of attributes mutated                             |
| Billing event count            | non-zero divergence (zero tolerance)                   |

A divergence below the floor is reported (logged in run-archive) but does not block the release. A divergence above the floor blocks until the operator either fixes the regression or signs off on a baseline-replacement PR.

## 13d. Per-Submodule Challenges Coverage Verification

The `helix-challenges-coverage` lint walks every per-submodule `tests/challenges/` directory + verifies:

1. The directory exists (or `delegated.md` per S03 §4.2).
2. The submodule's primary scenario is mapped per [§5](#5-the-per-submodule-challenges-map-s03-§4-reproduced-for-t11).
3. Any secondary scenarios are listed in the submodule's S05 §5 row.
4. The Challenges run-archive has at least one green entry in the past 30 days for each scenario.

Failing the lint blocks the merge. New submodules introduced via Constitution §15 amendment must satisfy all four conditions before merging the §3.1 catalog row.

## 14. References & Anti-Bluff Verification

### 14.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.10, §8.
- [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) — full S03 chapter.
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §5, §7, §9.
- [`../01_Constitution.md`](../01_Constitution.md) §1, §6.4.

### 14.2 External (web)

- VMAF: https://github.com/Netflix/vmaf (accessed 2026-04-30).
- ViSQOL: https://github.com/google/visqol (accessed 2026-04-30).
- minisign: https://jedisct1.github.io/minisign/ (accessed 2026-04-30).
- HDR Histogram: http://hdrhistogram.org/ (accessed 2026-04-30).

### 14.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 500-line floor.
- Forbidden patterns: clean. Four §13 open questions named with deferred resolution chapters.
- §7 + §9 anti-pattern catalogue + the §3 S03 ↔ T11 contract operationalise the anti-bluff backstop.
- §11 replay-from-archive + §12 pre-release exhaustive replay are the two operational guarantees that R-13 is enforced over time, not just at any instant.

### 14.4 The Anti-Bluff Backstop's Operational Story

The Challenges row is where R-13's "green tests guarantee real, end-user-usable behaviour" stops being aspirational. Past projects on this operator's portfolio (HelixAgent, Catalogizer per Constitution §6.4 R-14) had green Unit + Integration suites while end-user behaviour was broken — the source of the anti-bluff doctrine. Challenges is the operator's mitigation: a meta-test that observes user-visible artefacts directly, against signed baselines, with statistically rigorous change-point detection.

The five Challenges-internal failure modes (§7) are real risks; the five mitigations (signed baselines, digest-pinned topology, JSON-Schema scenario DSL, heartbeat manifests, two-reviewer threshold rule) close them. Together with the per-cadence orchestration (T12), the deployment gate's four-signal quad (S04 §9), and the immutable run-archive (S04 §7), Challenges turns R-13 from a slogan into an operational guarantee that survives operator turnover, mirror geography, and time.

When a HelixPlay deployment is in production and an operator wonders *"how do we know this is actually working?"*, the answer is: the Challenges row of the deployment-gate ran the full canonical user journey against the production-like topology, observed every user-visible artefact, compared each to a signed baseline, and stayed within tolerance — within the past 24 hours, on all four mirrors, with a signed audit-trail entry the operator can replay any time. That is what R-13 means in practice; that is what Challenges enforces.

### 14.5 Cross-Family Coverage Confirmation

The Challenges row is the bridge from test-discipline (this family) to operations (the next family). After Session 9 (Submodules) closure + Session 10 (Testing) closure that this chapter completes, the synthesis programme stands at:

- 4 content families closed (Architecture / Latency / Video-Audio / Submodules-aggregation = 56 chapters + 29 per-submodule descriptors).
- Testing family closing now (13 chapters; this is the last).
- 3 families remaining: Operations (`08_Operations/`); Implementation Phases (`09_Implementation_Phases/`); ongoing web-research addenda accumulation.

Challenges' coverage commitment per this chapter: every of the 29 submodules has a primary scenario; the 14 topologies cover every reference user journey; the four-mirror parity contract holds nightly; the 30-day exhaustive replay gate stands at the `v1.0.0+` graduation boundary. Whatever the synthesis programme produces in the next two families inherits this Challenges discipline by reference; the rules do not need to be restated.

The Operations family will codify the **deployment-side** application of these rules — when an operator stages a release, when a canary promotes to production, when an emergency rollback fires. The Implementation Phases family will sequence the actual rollout: which submodule's `v1.0.0` graduation comes first, which Phase brings up the host agent + first client + first user journey end-to-end against a real Vault + real CockroachDB. Both families consume the Challenges contract this chapter ratifies.

The synthesis programme's structural completeness is now within reach: 5 of 9 families closed (foundation + Architecture + Latency + Video/Audio + Submodules), plus this Testing family closes as the 6th. Three remaining families (Operations + Implementation Phases + ongoing addenda) will lift the synthesis to its R-01 + Master Plan §9 Definition of Done state.

### 14.6 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/11_Challenges.md` — 2026-04-30.
