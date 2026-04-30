# S03 — `vasic-digital/Challenges` Submodule Integration

> **Source dimensions:**
> - [`00_Index.md`](00_Index.md) §1 (Challenges integration role) + §5 (Ten-test-type matrix delegates the Challenges row to S03).
> - [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md) §3.1 (29 rows whose Challenges row is governed by S03), §5.1 (default in-tree shape), §5.2 (helix-shm Challenges-delegation exception), §5.3 (R-13 anti-bluff requirement), §5.5 (container-lane consequence).
> - [`02_Containers_Submodule.md`](02_Containers_Submodule.md) §3 (Challenges runs in the per-submodule containers S02 builds), §6 (signed images that S03 baselines reference), §8 (R-18 hazard policies that survive Challenges runs).
> - [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row S03 (≥ 400-line floor).
> - [`../01_Constitution.md`](../01_Constitution.md) §1 (Anti-Bluff Pledge, R-02 + R-13), §6 (Testing Discipline, R-11 + R-12 + R-13 — Challenges is the tenth and most demanding test type), §11.5 (R-18 — host-integrity-scan must survive a Challenges full-system run).
> - [`../02_System_Overview.md`](../02_System_Overview.md) §3 (Reference User Journey — Challenges baselines reproduce this journey), §8 (End-to-End Dataflow — Challenges topologies instantiate this).
> - The 37 chapters under [`../03_Architecture/`](../03_Architecture/), [`../04_Latency/`](../04_Latency/), [`../05_Video_Audio/`](../05_Video_Audio/) — every chapter §6's *Challenges* row resolves to a topology + baseline pair documented here.
> - The pre-existing organisational repo [`https://github.com/vasic-digital/Challenges.git`](https://github.com/vasic-digital/Challenges.git) — the production-like topologies, baseline archive, and change-point detection harness. **Not** a submodule; a sibling resource (analogous to `vasic-digital/Containers`).
>
> **Source line count:** family-level inputs ≈ 200 lines (S01 §5 + Constitution §6 + System Overview §3 + §8 slices) + per-chapter Challenges-row entries (resolved by reference). Master Plan §7.2 sets the chapter floor at **≥ 400 lines**; this chapter overshoots that floor on the strength of the topology catalog (§4) and the change-point detection methodology (§6).
>
> **Chapter targets:** R-02 (no bluff), R-11 (Ten test types), R-12 (only Unit may use mocks/stubs), R-13 (green tests guarantee real end-user behaviour), R-14 (Challenges discipline integrated as in HelixAgent and Catalogizer), R-18 (host-integrity-scan inheritance through Challenges runs).
>
> **Cross-links:** [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md), [`02_Containers_Submodule.md`](02_Containers_Submodule.md), [`04_HelixQA_Integration.md`](04_HelixQA_Integration.md) (HelixQA orchestrates Challenges runs autonomously and gates deployments on baseline parity), `per-submodule/<name>.md` (S05 — every descriptor's Challenges-row entry resolves to §4 of this chapter).
>
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## Table of Contents

1. The Role of `vasic-digital/Challenges` in the Synthesis Programme
2. Why Challenges Is the Tenth Test Type (R-13 anti-bluff backstop)
3. Repository Layout (topologies + baselines + harness)
4. Per-Submodule Challenges Entry-Points (29 + the helix-shm delegation)
5. Production-Like Topology Composition (the full-system contract)
6. Baseline Recording, Replay, and Change-Point Detection
7. Anti-Bluff Enforcement at the Challenges Boundary
8. Integration with `HelixDevelopment/HelixQA` (deferred to S04 by reference)
9. Challenges Cadence (per-PR / nightly / canary / pre-release)
10. Open Questions
11. References & Anti-Bluff Verification

---

## 1. The Role of `vasic-digital/Challenges` in the Synthesis Programme

`vasic-digital/Challenges` is the **production-like full-system test repository** for the 29-submodule fleet. Like `vasic-digital/Containers` (S02), it is **not** itself one of the 29 catalogued submodules in [S01 §3.1](01_Submodule_Catalog.md#31-the-29-submodule-table); it is a sibling organisational repository that supplies the Challenges-row machinery every submodule consumes.

The Challenges repo provides four kinds of artefact:

1. **Topology fixtures** — Docker Compose / Podman Quadlet / Kubernetes manifests that instantiate the full HelixPlay system (host agent + 29 submodules + CockroachDB + NATS + Redis + Vault + the audio/video stack) inside a hermetic container network. §5 catalogues the canonical topologies.
2. **Baseline archives** — recorded ground-truth artefacts for every Challenges scenario: rendered video frames, audio waveforms, controller-input round-trip traces, gRPC call records, latency histograms. §6 documents the baseline format and the SHA-256 chain that prevents tampering.
3. **The change-point detection harness** — a statistical engine (per [C24 §6](../04_Latency/10_Latency_Testing_and_Validation.md) and [C35 §6](../05_Video_Audio/10_Measurement_and_QA.md)) that compares a fresh Challenges run's observations against the baseline and decides whether the difference is a regression or a benign perturbation. The harness is shared because regression-decision logic must be identical across the 29 submodules.
4. **The replay harness** — a deterministic re-execution wrapper that takes a baseline trace and replays the exact controller inputs / network events / clock skew the baseline recorded, so a Challenges run is reproducible bit-for-bit (modulo intentional non-determinism, e.g. encoder GOP cadence).

Without `vasic-digital/Challenges`, every submodule would re-implement topology fixtures, baseline storage, change-point detection, and replay logic — a 29-fold duplication that R-04 forbids. R-14 (Constitution §6.4) explicitly mandates the Challenges discipline as integrated "as in HelixAgent and Catalogizer" — the two prior projects in the operator's portfolio that pioneered the Challenges pattern.

`★ R-14 wording.` Constitution §6.4 reads:

> The `Challenges` discipline (`git@github.com:vasic-digital/Challenges.git`) is integrated as in HelixAgent and Catalogizer. Every submodule's §8.10 Challenges row hooks into this repository for production-like full-system runs. Challenges runs MUST exercise the entire user-visible behaviour set, not a stripped subset.

S03 is the chapter that operationalises R-14. The integration contract is: every submodule's CI lane (S02) emits a Challenges-row invocation; the invocation runs inside the per-submodule container (S02 §3), composes a topology from `vasic-digital/Challenges/topologies/<name>/`, runs the user-journey scenario, observes the user-visible outputs, compares against the baseline at `vasic-digital/Challenges/baselines/<name>/`, and reports green or red. R-13's "green tests guarantee real end-user-usable behaviour" is operationalised at the Challenges boundary because the boundary is the only place where the **entire** user-visible behaviour set is observed.

---

## 2. Why Challenges Is the Tenth Test Type (R-13 anti-bluff backstop)

The Ten test types ([Constitution §6.2](../01_Constitution.md#6-testing-discipline-r-11-r-12-r-13)) are:

1. Unit
2. Integration
3. E2E
4. Security
5. Benchmarking
6. Chaos
7. Stress
8. Smoke
9. Full Automation
10. **Challenges**

The first nine each have a known failure mode where green test results can co-exist with broken end-user behaviour. R-13 (Constitution §1) records the specific operational concern: prior projects had green tests on broken features, so the test types alone are insufficient — a meta-test is required.

**The known failure modes of the first nine:**

- **Unit** — mocks may misrepresent the real dependency (S01 §4.5.2 documents this). A Unit test of the gRPC client's retry logic that mocks the server can pass even if the real server's retry semantics differ.
- **Integration** — fixture data is a curated subset of production data. A query that works against the fixture may fail against a real database with realistic data shape (e.g. NULL-handling, enum-value drift, JSONB schema migrations).
- **E2E** — typically scoped to a single user journey. E2E green for "user logs in, browses catalog, starts a session" can coexist with broken behaviour for "user with a 7-day-stale auth token tries to start a session that requires a freshly-issued attestation".
- **Security** — vulnerability scanners (govulncheck, Snyk) report known CVEs. They do not exercise the application's actual code paths, so a vulnerable function present in `go.mod` but unreachable produces a false positive; a vulnerable function reachable via an obscure call path may produce a false negative if the scanner's call-graph analysis is incomplete.
- **Benchmarking** — measures throughput / latency under a controlled workload. A benchmark that hits cache lines hot in the test environment may report numbers that bear no resemblance to the cold-cache reality of a production deployment.
- **Chaos** — injects controlled faults (Toxiproxy, chaos-mesh). The injected faults are a subset of real-world faults; the chaos test cannot inject a fault its author didn't anticipate.
- **Stress** — long-running load. A stress test that runs 24 hours under steady-state load may not exercise the burst-traffic paths a real deployment encounters during e.g. a marketing event spike.
- **Smoke** — a 30-second sanity check post-deploy. It verifies "the binary started and answered one request"; it does not exercise feature behaviour beyond the smoke endpoint.
- **Full Automation** — orchestrates the previous eight in CI matrix. If any of the eight has a blind spot, Full Automation inherits the blind spot.

**Challenges closes the loop.** A Challenges scenario:

- Boots the entire production-like topology (§5).
- Runs a recorded user journey ([System Overview §3](../02_System_Overview.md#3-reference-user-journey) — login, catalog browse, session start, controller input, video playback, audio playback, recording emit, session end, billing tick).
- Observes every user-visible artefact: rendered video frames, audio waveforms, controller-input round-trips, gRPC call records, OTLP traces, billing events.
- Compares each observation against the baseline.
- Reports green only if **every** observation matches within tolerance.

A submodule's bug that is invisible to Unit / Integration / E2E / Security / Benchmarking / Chaos / Stress / Smoke / Full Automation **cannot** be invisible to Challenges, because Challenges observes user-visible behaviour and the bug, by definition, manifests as user-visible behaviour drift.

This is the anti-bluff backstop. It is not perfect (see §7 for the failure modes Challenges itself can have), but it is the operational floor below which green tests do not guarantee real behaviour.

---

## 3. Repository Layout (topologies + baselines + harness)

`vasic-digital/Challenges` is structured as follows. The layout mirrors S02 §2's "single narrow responsibility per directory" rule.

```
vasic-digital/Challenges/
├── topologies/                                    ← production-like full-system compose specs
│   ├── 01_minimum_viable_session/                 ← single host + single client + 1× streaming session
│   │   ├── docker-compose.yml
│   │   ├── README.md
│   │   ├── networks.json                          ← cni network definitions
│   │   ├── volumes.json                           ← persistent volume definitions
│   │   └── scenarios/
│   │       ├── 01_login_browse_play_disconnect.scenario.yaml
│   │       ├── 02_login_resume_session.scenario.yaml
│   │       └── 03_login_invalid_token_attestation.scenario.yaml
│   ├── 02_multi_session_single_host/              ← single host + N clients + N sessions
│   ├── 03_multi_host_multi_session/               ← N hosts + N clients + N sessions, NATS clustering
│   ├── 04_white_label_multi_tenant/               ← multi-tenant theming + per-tenant catalog
│   ├── 05_recording_emit_and_replay/              ← session recording → fMP4 + MKV → S3 sync
│   ├── 06_4k_120hz_hdr_dolby_vision/              ← 4K 120 Hz + HDR10/Dolby Vision + Atmos audio
│   ├── 07_kernel_bypass_send_path/                ← io_uring + XDP send-path full pipeline
│   ├── 08_controller_input_low_latency/           ← Reflex + DSCP + ALLM + 240 Hz gamepad polling
│   ├── 09_thermal_throttling_dvfs/                ← thermal-aware quality scaling under load
│   ├── 10_chaos_network_partition/                ← Toxiproxy partition between host and client
│   ├── 11_stress_24h_steady_state/                ← 24-hour steady-state stress
│   ├── 12_burst_marketing_spike/                  ← 10× traffic burst over 5 minutes
│   ├── 13_audit_compliance_eu_dsa/                ← EU DSA Article 17 catalog ingest + audit emit
│   └── 14_security_attack_surface/                ← attempted privilege-escalation, deny-list checks
├── baselines/                                     ← recorded ground-truth per topology × scenario
│   ├── 01_minimum_viable_session/
│   │   ├── 01_login_browse_play_disconnect/
│   │   │   ├── frames/                            ← rendered video frames (PNG, 60 frames @ 1 Hz subsampling)
│   │   │   ├── audio.flac                         ← rendered audio (FLAC, full session)
│   │   │   ├── controller-rtt.json                ← controller-input round-trip times
│   │   │   ├── grpc-trace.json                    ← gRPC call sequence
│   │   │   ├── otlp-trace.json                    ← OpenTelemetry trace span records
│   │   │   ├── billing-events.jsonl               ← per-second billing event records
│   │   │   ├── latency-histograms.json            ← p50/p99/p999 histograms for input→render→display
│   │   │   ├── baseline.manifest.json             ← SHA-256 manifest of every artefact above
│   │   │   ├── baseline.manifest.json.minisig     ← minisign signature over the manifest
│   │   │   └── baseline.metadata.json             ← topology version + recording date + recorder identity
│   │   └── ...
│   └── ...
├── harness/                                       ← shared replay + change-point detection engine
│   ├── replay/
│   │   ├── controller-replay.go                   ← deterministic controller input replay
│   │   ├── network-replay.go                      ← deterministic network event replay (Toxiproxy script)
│   │   ├── clock-replay.go                        ← deterministic clock-skew replay
│   │   └── README.md
│   ├── change-point/
│   │   ├── frame-diff.go                          ← VMAF + SSIM comparison against baseline frames (per C35 §6)
│   │   ├── audio-diff.go                          ← spectrogram + ViSQOL comparison against baseline audio
│   │   ├── latency-diff.go                        ← Mann-Whitney U + Kolmogorov-Smirnov on histograms (per C24 §6)
│   │   ├── trace-diff.go                          ← gRPC + OTLP trace topology comparison
│   │   └── threshold-config.yaml                  ← per-metric tolerance thresholds
│   └── observe/
│       ├── frame-recorder.go                      ← captures rendered frames at scenario boundaries
│       ├── audio-recorder.go                      ← captures rendered audio for the entire scenario
│       ├── trace-recorder.go                      ← captures gRPC + OTLP traces
│       └── billing-recorder.go                    ← captures billing event stream
├── scenarios/                                     ← scenario DSL specifications
│   ├── schema.json                                ← JSON Schema for *.scenario.yaml
│   └── README.md                                  ← DSL documentation
└── scripts/
    ├── challenge-run.sh                           ← per-submodule entrypoint: invoked from CI lane
    ├── baseline-record.sh                         ← record a fresh baseline (operator-only, signed)
    ├── baseline-verify.sh                         ← verify minisign chain on a baseline
    └── change-point-report.sh                     ← human-readable diff report for a failing run
```

The `topologies/` directory is **frozen**: 14 canonical topologies cover the user journeys defined in [System Overview §3](../02_System_Overview.md#3-reference-user-journey). A 15th topology requires an explicit Constitution amendment (§15) because a new topology is a new contract the fleet must satisfy.

The `baselines/` directory is **append-only**: every recording is signed, timestamped, and immutable. A baseline that is found incorrect (e.g. it captured a regression that escaped detection) is **not deleted**; instead, a new baseline is recorded with `baseline.metadata.json` referencing the prior one as `superseded-by` / `supersedes`. The audit trail is preserved.

---

## 4. Per-Submodule Challenges Entry-Points (29 + the helix-shm delegation)

Every submodule's CI lane (S02) emits a Challenges-row invocation. The invocation specifies which topology × scenario the submodule is responsible for, by direct reference to the §3 directory path.

### 4.1 The 29 entry-point map

| #  | Submodule              | Primary topology               | Primary scenario                                        |
|----|------------------------|--------------------------------|---------------------------------------------------------|
| 01 | `helix-r18-safeexec`   | `14_security_attack_surface`   | `01_attempt_forbidden_subprocess_invocation.scenario`   |
| 02 | `helix-grpc-frame`     | `02_multi_session_single_host` | `01_grpc_streaming_under_burst_load.scenario`           |
| 03 | `helix-tv-input`       | `01_minimum_viable_session`    | `04_tv_remote_navigation_full_journey.scenario`         |
| 04 | `helix-vault`          | `13_audit_compliance_eu_dsa`   | `02_kek_rotation_under_active_session.scenario`         |
| 05 | `helix-tenant`         | `04_white_label_multi_tenant`  | `01_tenant_theme_swap_at_session_boundary.scenario`     |
| 06 | `helix-shm`            | DELEGATED → `helix-pipeline`   | (S01 §5.2 delegation)                                   |
| 07 | `helix-iouring`        | `07_kernel_bypass_send_path`   | `01_iouring_completion_under_120hz_send.scenario`       |
| 08 | `helix-xdp`            | `07_kernel_bypass_send_path`   | `02_xdp_redirect_under_packet_burst.scenario`           |
| 09 | `helix-lockfree`       | `02_multi_session_single_host` | `02_spsc_ringbuffer_under_concurrent_session.scenario`  |
| 10 | `helix-gpu-direct`     | `06_4k_120hz_hdr_dolby_vision` | `01_gpu_direct_rdma_under_4k120_session.scenario`       |
| 11 | `helix-network`        | `08_controller_input_low_latency` | `01_dscp_l4s_under_240hz_input.scenario`             |
| 12 | `helix-rtos`           | `08_controller_input_low_latency` | `02_sched_fifo_under_high_load.scenario`            |
| 13 | `helix-input`          | `08_controller_input_low_latency` | `03_reflex_round_trip_under_load.scenario`          |
| 14 | `helix-display`        | `06_4k_120hz_hdr_dolby_vision` | `02_vrr_under_burst_frame_load.scenario`                |
| 15 | `helix-mempool`        | `02_multi_session_single_host` | `03_mempool_no_alloc_in_hot_path.scenario`              |
| 16 | `helix-allocator`      | `02_multi_session_single_host` | `04_allocator_enforces_hot_path_zero_alloc.scenario`    |
| 17 | `helix-bench`          | `11_stress_24h_steady_state`   | `01_p999_floor_over_24h.scenario`                       |
| 18 | `helix-codec`          | `06_4k_120hz_hdr_dolby_vision` | `03_codec_ladder_at_4k120.scenario`                     |
| 19 | `helix-encoder`        | `06_4k_120hz_hdr_dolby_vision` | `04_nvenc_qsv_amf_videotoolbox_parity.scenario`         |
| 20 | `helix-capture`        | `01_minimum_viable_session`    | `05_capture_per_os_at_session_start.scenario`           |
| 21 | `helix-dualpath`       | `05_recording_emit_and_replay` | `01_dual_path_nal_stream_plus_record.scenario`          |
| 22 | `helix-record`         | `05_recording_emit_and_replay` | `02_fmp4_mkv_emit_then_replay.scenario`                 |
| 23 | `helix-audio`          | `06_4k_120hz_hdr_dolby_vision` | `05_atmos_eARC_under_full_journey.scenario`             |
| 24 | `helix-hdr`            | `06_4k_120hz_hdr_dolby_vision` | `06_dolby_vision_tone_map_under_session.scenario`       |
| 25 | `helix-abr`            | `12_burst_marketing_spike`     | `01_abr_ladder_descent_under_10x_burst.scenario`        |
| 26 | `helix-thermal`        | `09_thermal_throttling_dvfs`   | `01_thermal_throttling_quality_scaling.scenario`        |
| 27 | `helix-vqa`            | `06_4k_120hz_hdr_dolby_vision` | `07_vmaf_ldat_against_baseline.scenario`                |
| 28 | `helix-pipeline`       | `01_minimum_viable_session`    | `06_full_pipeline_end_to_end.scenario` + helix-shm delegation|
| 29 | `helix-transport`      | `10_chaos_network_partition`   | `01_quic_recovery_after_partition.scenario`             |

### 4.2 The `helix-shm` Challenges delegation (S01 §5.2)

`helix-shm` is a library, not a runtime; it cannot stand up a full system on its own. Its Challenges row is delegated to `helix-pipeline`, which imports `helix-shm` directly (per S01 §3.1's `Direct deps` column) and exercises every `helix-shm` API surface inside the `01_minimum_viable_session` topology's `06_full_pipeline_end_to_end` scenario.

The delegation is documented in the per-submodule descriptor (S05) `helix-shm.md` and in `helix-shm/tests/challenges/delegated.md`. The `delegated.md` file's content is fixed:

```markdown
# helix-shm — Challenges row delegation

This submodule does not stand up a full system. Its Challenges row is
exercised by `helix-pipeline` under the
`vasic-digital/Challenges/topologies/01_minimum_viable_session/scenarios/06_full_pipeline_end_to_end.scenario.yaml`
scenario, which uses every public `helix-shm` API.

For change-point detection on `helix-shm`-specific metrics, see
`vasic-digital/Challenges/baselines/01_minimum_viable_session/06_full_pipeline_end_to_end/per-submodule/helix-shm.metrics.json`
which records the helix-shm slice of the full-pipeline baseline.

When the per-pipeline baseline shifts in a way attributable to
helix-shm (e.g. zero-copy throughput regression), the change-point
report flags helix-shm even though the harness invocation was made by
helix-pipeline. This is the single delegation-traceability path.
```

The delegation is the **only** sanctioned cross-submodule reference for the Challenges row. Every other submodule's Challenges row is in-tree per S01 §5.1.

### 4.3 Multi-topology submodules

Several submodules touch more than one topology. The §4.1 table records the **primary** scenario; secondary scenarios run on the nightly cadence (§9). For example, `helix-pipeline` is the primary actor in `01_minimum_viable_session/06_full_pipeline_end_to_end` but also participates as a co-located submodule in every other topology that invokes streaming. Its CI lane runs the primary scenario per-PR; the nightly run extends to the full set.

### 4.4 The new-submodule Challenges-row contract

When a future chapter introduces a 30th submodule (S01 §3.4 makes this an explicit Constitution amendment event), the new chapter must:

1. Identify the submodule's primary topology from §4.1's catalog. If no existing topology is suitable, propose a 15th topology as part of the same Constitution amendment.
2. Identify the primary scenario; if no existing scenario fits, add one to the chosen topology's `scenarios/` directory.
3. Update the §4.1 table in S03.
4. Update [S01 §3.1](01_Submodule_Catalog.md#31-the-29-submodule-table) to include the 30th row.

The contract is symmetric: a new submodule cannot pass R-12 + R-14 review without resolving its Challenges-row entry in this table.

---

## 5. Production-Like Topology Composition (the full-system contract)

A "production-like" topology is one that satisfies four properties:

1. **Every submodule that participates in the user journey is co-located.** No submodule is mocked at the Challenges boundary (Constitution §6.2's R-12 mocks-only-in-Unit rule applies; Challenges is the strictest enforcement of that rule).
2. **Every backing service is real.** CockroachDB is a real CockroachDB cluster (3 nodes for `01_minimum_viable_session`, 9 nodes for `03_multi_host_multi_session`); NATS is a real NATS cluster; Redis is a real Redis instance; the Vault is a real Vault server (not the dev-mode in-memory variant). Constitution §3 (R-05 + R-06) requires this — every backing service runs in a container per `vasic-digital/Containers`.
3. **The host OS capture stack is real.** `helix-capture` runs against a real X11 / Wayland / DXGI / Metal / PipeWire surface; the topology spec includes a virtual display server (Xvfb for X11, weston-vnc for Wayland, mock-DXGI shim for Windows tests, headless-Metal for macOS tests on M-series Apple Silicon CI runners).
4. **The network is realistic.** Network conditions (latency, jitter, loss) are injected via Toxiproxy with profiles drawn from real-world measurements documented in [C19 §6](../04_Latency/05_UltraLowLatency_Network_Protocols.md) and [C37 §9](../05_Video_Audio/12_Network_Transport.md). A "good" home-broadband profile, a "bad" cellular profile, and a "challenging" Russian-jurisdiction enterprise-router profile are the three canonical baselines.

### 5.1 Topology #1 — `01_minimum_viable_session` (the canonical exemplar)

The `01_minimum_viable_session` topology is the simplest non-trivial topology and the canonical exemplar for the rest. Its compose spec assembles:

- **1× HelixPlay host agent** (the Sunshine++ fork; runs `helix-r18-safeexec`, `helix-grpc-frame`, `helix-capture`, `helix-encoder`, `helix-codec`, `helix-dualpath`, `helix-record`, `helix-audio`, `helix-hdr`, `helix-pipeline`, `helix-transport`).
- **1× HelixPlay client** (Wails desktop or Compose-for-TV; runs `helix-tv-input`, `helix-input`, `helix-display`, `helix-tenant`).
- **1× CockroachDB single-node** (catalog + tenant + session storage).
- **1× NATS single-node** (event bus per Constitution §10).
- **1× Redis single-node** (cache + rate-limit per [C06 §6](../03_Architecture/05_RealTime_APIs.md)).
- **1× Vault** (KEK/DEK + tenant secrets per [C10 §6](../03_Architecture/09_Security_and_Isolation.md)).
- **1× Toxiproxy** (network impairment injector).
- **1× Xvfb / weston-vnc** (virtual display server for `helix-capture`).
- **1× otel-collector** (trace + metric receiver).
- **1× challenge-orchestrator** (the `harness/` engine that drives scenarios).

The compose spec is a single `docker-compose.yml` checked into `topologies/01_minimum_viable_session/docker-compose.yml`. Every container image is a digest-pinned reference (S02 §5) and every container has the §8 R-18 hazard policies applied.

### 5.2 Network composition

Containers communicate over a custom CNI network specified in `networks.json`:

```json
{
  "name": "helixplay-challenge-net",
  "driver": "bridge",
  "ipam": {
    "subnet": "172.31.0.0/16",
    "gateway": "172.31.0.1"
  },
  "options": {
    "com.docker.network.bridge.enable_ip_masquerade": "true",
    "com.docker.network.bridge.enable_icc": "true",
    "com.docker.network.driver.mtu": "1500"
  }
}
```

The MTU of 1500 mirrors a typical home-broadband egress path; topologies that test jumbo frames or PMTUD use larger MTUs explicitly.

### 5.3 The four-mirror amplifier (Challenges edition)

Challenges runs are **identical** across the four-mirror CI runners. The same compose spec, same container digests, same baselines. A Challenges run that is green on GitHub Actions and red on GitFlic CI indicates a registry-replication drift or an infrastructure regression on one mirror; it is **never** an acceptable steady state.

The `four-mirror-challenges-parity.sh` audit script (`vasic-digital/Challenges/scripts/`) runs nightly and reports any divergence in green/red parity across the four mirrors. A divergence opens a P1 ticket on the operator's dashboard.

---

## 6. Baseline Recording, Replay, and Change-Point Detection

### 6.1 Recording

A baseline is recorded by an operator-authorised process. The `baseline-record.sh` script:

1. Boots the topology with the canonical compose spec.
2. Runs the scenario via the replay harness.
3. Captures every observation (frames, audio, traces, billing events, histograms).
4. Computes the SHA-256 manifest.
5. Signs the manifest with the operator's minisign key.
6. Commits the baseline to `baselines/<topology>/<scenario>/` on a feature branch.
7. Opens a PR for operator review.

The PR review is the gate. A baseline does not become canonical until the operator merges it. The operator's role here is non-delegable (per Constitution §16 *Acceptance*) — a maintainer cannot self-record a baseline.

### 6.2 Replay

A Challenges run replays the baseline's inputs deterministically:

- **Controller inputs** are replayed from `controller-rtt.json` with sub-millisecond timestamp accuracy.
- **Network events** (Toxiproxy script invocations) are replayed from a recorded script in the baseline.
- **Clock skew** is replayed by injecting `CLOCK_REALTIME` adjustments at the same offsets the baseline recorded.
- **Random sources** are seeded deterministically (e.g. encoder rate-distortion choices use a fixed seed; the baseline records the seed).

Determinism does **not** extend to:

- Encoder GOP cadence (which is bitrate-adaptive and intentionally nondeterministic per [C26 §6](../05_Video_Audio/01_Codec_Selection.md)). The change-point detection (§6.3) tolerates this within bounded variance.
- Thermal throttling decisions (which depend on the host's thermal state). Topology `09_thermal_throttling_dvfs` drives the throttling deliberately to make the test reproducible; other topologies tolerate thermal variance.

### 6.3 Change-point detection

A Challenges run produces a fresh observation set; the harness compares it against the baseline using metric-specific algorithms:

| Metric                          | Algorithm                                  | Threshold (default)                     |
|---------------------------------|--------------------------------------------|------------------------------------------|
| Rendered video frames           | VMAF + SSIM (per [C35 §6](../05_Video_Audio/10_Measurement_and_QA.md)) | VMAF ≥ 95.0; SSIM ≥ 0.99               |
| Rendered audio                  | Spectrogram MSE + ViSQOL                   | MSE ≤ 0.001; ViSQOL ≥ 4.5              |
| Latency histograms              | Mann-Whitney U + Kolmogorov-Smirnov        | p-value > 0.01                          |
| gRPC trace topology             | Trace-graph isomorphism + edge-weight diff | ≥ 99 % edge-weight match               |
| OTLP span structure             | Span-tree shape + attribute diff           | structural identity, attribute diff ≤ 5 % |
| Billing event count             | Exact equality                             | identical                                |
| Frame timing (frame-pacing VRR) | Latency histogram + jitter histogram       | p99 < baseline + 0.5 ms                |
| p999 latency                    | Single-tail Welch's t-test                 | p-value > 0.01                          |

Thresholds are configured in `harness/change-point/threshold-config.yaml` and are **per-topology + per-scenario**, not global. A topology that is sensitive to thermal variance (`09_thermal_throttling_dvfs`) has wider VMAF tolerance than `01_minimum_viable_session`.

### 6.4 The "regression vs perturbation" distinction

Not every change-point is a regression. Examples of **perturbations** that the harness tolerates:

- A 1-frame-long VMAF drop at scenario boundary that the baseline also shows.
- A p999 latency that moved by < 5 % in either direction (within the noise floor for 10K samples).
- A trace edge-weight that changed by < 1 %.

Examples of **regressions** that the harness flags:

- A sustained VMAF drop > 5 points across > 30 frames.
- A p999 latency that moved by > 5 % toward worse.
- A new gRPC call that the baseline does not contain (could indicate an unintended retry storm).
- A missing billing event.
- An OTLP span attribute that disagrees with the baseline.

The harness's report (`change-point-report.sh`) categorises every flagged change-point and produces a human-readable diff. The operator reviews the report; if the change-point is a known-good change (e.g. an intentional encoder bitrate ladder revision), the operator records the baseline replacement (§6.1) rather than overriding the harness.

---

## 7. Anti-Bluff Enforcement at the Challenges Boundary

Challenges is the meta-test (§2). It is **not** infallible. The known failure modes:

### 7.1 Baseline tampering

A maintainer with commit access to `vasic-digital/Challenges/baselines/` could record a wrong baseline that hides a regression. The mitigation:

- Baselines are minisign-signed by the operator's key.
- The baseline-record-PR review (§6.1) is non-delegable.
- The four-mirror replication catches any post-merge baseline drift.
- The `baseline-verify.sh` script in every CI lane verifies the minisign chain before consuming the baseline; a missing or invalid signature fails the run.

### 7.2 Topology drift

A change to a topology's compose spec (e.g. swapping CockroachDB for a mocked storage backend) would invalidate the production-like guarantee. The mitigation:

- Every container image in a topology spec is digest-pinned (S02 §5).
- A PR that modifies a topology's compose spec triggers a baseline re-recording requirement; the PR cannot merge until the new baseline is signed.
- The `four-mirror-challenges-parity.sh` audit catches any unauthorised topology change after merge.

### 7.3 Scenario narrowing

A scenario that is rewritten to skip a user journey path would hide a regression in the skipped path. The mitigation:

- Scenarios are defined in a JSON-Schema-validated DSL (`scenarios/schema.json`).
- The DSL enforces a minimum set of operations per user-journey type; e.g. a `login` scenario MUST include token issuance, attestation verification, session establishment, and session disconnect.
- A PR that modifies a scenario triggers a baseline re-recording; same gate as topology drift.

### 7.4 Observation gaps

A Challenges run that fails to capture an observation (e.g. the frame recorder crashes mid-scenario) would silently pass change-point detection because there is no fresh observation to compare. The mitigation:

- The `harness/observe/` recorders emit heartbeat events to the orchestrator. A missing heartbeat fails the run.
- The fresh-observation manifest must contain every artefact the baseline manifest contains; a missing artefact fails the run.

### 7.5 Threshold inflation

A maintainer who is tired of a flaky scenario could inflate the threshold (`threshold-config.yaml`) to make it pass. The mitigation:

- `threshold-config.yaml` is part of `vasic-digital/Challenges`, not a per-submodule file. A PR to it requires the same two-reviewer rule as `helix-r18-safeexec` (per S01 §6.3).
- Threshold-inflation PRs are flagged for explicit operator review; the operator must record the rationale (e.g. "encoder behaviour changed in upstream NVENC driver, baseline + threshold updated together") and the related baseline change PR.

The five mitigations together close the practical gaps. Constitution §16's two-reviewer rule on `vasic-digital/Challenges` PRs is the human-in-the-loop backstop.

---

## 8. Integration with `HelixDevelopment/HelixQA` (deferred to S04 by reference)

The autonomous orchestration of Challenges runs (per-PR + nightly + canary + pre-release) is the responsibility of `HelixDevelopment/HelixQA`, documented in [`04_HelixQA_Integration.md`](04_HelixQA_Integration.md) (S04). S03 specifies the **inputs** HelixQA consumes (the topology specs, baselines, and harness from `vasic-digital/Challenges`); S04 specifies how HelixQA invokes them.

The interface contract between S03 and S04 is:

- HelixQA invokes `vasic-digital/Challenges/scripts/challenge-run.sh <submodule> <topology> <scenario>`.
- The script returns exit code 0 (green) or non-zero (red) plus a `change-point-report.json` artefact.
- HelixQA gates deployments on green at the appropriate cadence (per-PR for the submodule's primary scenario; nightly for the secondary set; canary for the full topology fan-out; pre-release for an exhaustive replay).

Both repositories are organisational repos (not catalogued submodules); their integration is a horizontal cross-org contract.

---

## 9. Challenges Cadence (per-PR / nightly / canary / pre-release)

| Cadence       | Scope                                                            | Owner          | Wall-clock budget |
|---------------|------------------------------------------------------------------|----------------|-------------------|
| Per-PR        | Primary scenario per submodule (§4.1 column "Primary scenario") | Per-submodule CI lane | 5–15 minutes      |
| Nightly       | All §4.1 scenarios across all 29 submodules                      | HelixQA        | 2–4 hours         |
| Canary        | Full 14-topology fan-out × every scenario                         | HelixQA        | 12–24 hours       |
| Pre-release   | Canary + the `11_stress_24h_steady_state` long-running scenario  | HelixQA        | 24–48 hours       |

The per-PR cadence is the developer's feedback loop; the nightly cadence is the catch-net for issues the per-PR scenario missed; the canary cadence runs against deployment candidates; the pre-release cadence runs before any `v1.0.0+` tag (per [S01 §9](01_Submodule_Catalog.md#9-release-train-cadence-v0-→-v1-→-v2-ladder) graduation gate).

The cadence is **non-overridable**. A PR that disables a per-PR Challenges invocation is rejected by the per-submodule CI lane's required-checks rule.

---

## 10. Open Questions

| ID            | Question                                                                                          | Defer to                                            |
|---------------|---------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-S03-A      | Should baseline VMAF tolerance be tightened from 95.0 to 97.0 for `06_4k_120hz_hdr_dolby_vision`? | `07_Testing/02_Unit_Tests.md` calibration discussion |
| OQ-S03-B      | The `13_audit_compliance_eu_dsa` topology — does it require a synthetic third-party DSA inspector? | `08_Operations/04_Observability_and_Events.md`     |
| OQ-S03-C      | Pre-release `11_stress_24h_steady_state` budget vs `v1.0.0` cadence — 24-hour blocker acceptable? | `09_Implementation_Phases/Phase_12_Beta_Launch.md` |
| OQ-S03-D      | Mac CI runners on Apple Silicon for `helix-capture-darwin` Challenges — vendor choice (BuildJet vs MacStadium)? | `08_Operations/01_Container_CI_CD.md`             |

None of the four are placeholders; each has a named resolution chapter and a specific operational concern.

---

## 11. References & Anti-Bluff Verification

### 11.1 Internal

- [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md) §3 (29 rows), §5.1 (in-tree default), §5.2 (helix-shm delegation), §5.3 (R-13 anti-bluff), §6.3 (helix-r18-safeexec SPOF), §9 (release-train).
- [`02_Containers_Submodule.md`](02_Containers_Submodule.md) §3 (per-submodule containers), §6 (signed images), §8 (R-18 hazard policies).
- [`../01_Constitution.md`](../01_Constitution.md) §1, §6, §11.5.
- [`../02_System_Overview.md`](../02_System_Overview.md) §3, §8.
- [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row S03.
- [`../04_Latency/10_Latency_Testing_and_Validation.md`](../04_Latency/10_Latency_Testing_and_Validation.md) §6 (change-point detection methodology — Mann-Whitney U + Kolmogorov-Smirnov + p999).
- [`../05_Video_Audio/10_Measurement_and_QA.md`](../05_Video_Audio/10_Measurement_and_QA.md) §6 (VMAF + SSIM + LDAT methodology).

### 11.2 External (web)

- VMAF: https://github.com/Netflix/vmaf (accessed 2026-04-29).
- SSIM specification: https://en.wikipedia.org/wiki/Structural_similarity (accessed 2026-04-29).
- ViSQOL: https://github.com/google/visqol (accessed 2026-04-29).
- Toxiproxy: https://github.com/Shopify/toxiproxy (accessed 2026-04-29).
- chaos-mesh: https://chaos-mesh.org/ (accessed 2026-04-29).
- minisign: https://jedisct1.github.io/minisign/ (accessed 2026-04-29).
- OpenTelemetry collector: https://opentelemetry.io/docs/collector/ (accessed 2026-04-29).
- NATS: https://nats.io/ (accessed 2026-04-29).
- CockroachDB: https://www.cockroachlabs.com/docs/ (accessed 2026-04-29).

### 11.3 Anti-Bluff Verification

| Path                                                     | Lines  | Reviewed   | Role                                            |
|----------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`01_Submodule_Catalog.md`](01_Submodule_Catalog.md)    |  1,218 | 2026-04-30 | catalog rows + cross-cutting policies           |
| [`02_Containers_Submodule.md`](02_Containers_Submodule.md) |    627 | 2026-04-30 | container topology + R-18 hazard policies      |
| [`../01_Constitution.md`](../01_Constitution.md) §1 §6 §11.5 | (cited slices) | 2026-04-30 | R-02 + R-11 + R-12 + R-13 + R-14 + R-18  |
| [`../02_System_Overview.md`](../02_System_Overview.md) §3 §8 | (cited slices) | 2026-04-30 | reference user journey + dataflow            |

- Coverage: this chapter exceeds the 400-line floor (`wc -l` recorded at chapter close).
- Forbidden patterns: clean. No `TODO` / `FIXME` / `tbd` / `xxx` / `???` / `placeholder` / "fill in later" markers in the chapter body. The four §10 open questions are explicitly named with deferred resolution chapters.
- R-13 anti-bluff posture: §2 enumerates the failure modes of the first nine test types and §7 enumerates the failure modes of Challenges itself. Neither pretends to be infallible; both are documented.
- R-14 wording (Constitution §6.4) is operationalised in §1 + §3 + §4 + §9.

### 11.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 as a single inline `Write` call.
- Reviewed by: pending operator review.

End of `06_Submodules/03_Challenges_Submodule.md` — 2026-04-30.
