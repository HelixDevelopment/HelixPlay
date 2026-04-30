# T04 — End-to-End (E2E) Tests

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.3; [`../02_System_Overview.md`](../02_System_Overview.md) §3 (Reference User Journey); [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md); [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §5 (production-like topology composition).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row T04.
> **Chapter targets:** R-11 (every reference user journey covered), R-12 (no mocks), R-13 (close the gap that Unit + Integration leave).
> **Cross-links:** [`03_Integration_Tests.md`](03_Integration_Tests.md), [`11_Challenges.md`](11_Challenges.md), [`12_HelixQA_Autonomous.md`](12_HelixQA_Autonomous.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

E2E is **test-type 3 of 10** in the [T01 §2 grid](01_Test_Matrix.md#2-the-29--10--4-cell-grid-the-canonical-lookup). Where Integration verifies that *one submodule* talks correctly to its dependencies, E2E verifies that *a complete user journey* works end-to-end across the host agent + clients + backing services.

E2E sits between Integration (one submodule + deps) and Challenges (full system + recorded baseline + change-point detection). The boundary:

- **Integration** — boots ≤ 5 services and verifies one submodule's exported API; pass/fail by assertion.
- **E2E** — boots the full topology (host agent + 1 client + CockroachDB + NATS + Redis + Vault + capture stack + encoder + transport) and verifies one user journey end-to-end; pass/fail by user-visible behaviour.
- **Challenges** — same topology as E2E but with **recorded baselines** + **change-point detection** instead of binary pass/fail; the meta-test that closes R-13.

E2E runs at every cadence; the *primary* user journey runs per-PR, the *full secondary set* runs nightly.

---

## 2. The Discipline — What Makes "Green" Green

A green E2E row means:

1. **Every reference user journey from [System Overview §3](../02_System_Overview.md#3-reference-user-journey) is covered** by at least one E2E test.
2. **No mocks anywhere in the topology** (R-12). Every backing service is real; every co-located submodule is real.
3. **The journey reaches its terminal state without intervention** — login → catalog browse → session start → controller input → playback → session end → billing tick — without the test author manually nudging state.
4. **All observed user-visible artefacts match expectations** — rendered frames are non-empty; audio is present; controller inputs propagated to the host process; billing event emitted.
5. **Topology teardown is clean** — no leaked container, no leaked goroutine, no leaked file descriptor across test boundaries.

---

## 3. Tooling

- Go's `testing` package + `//go:build e2e` build tag.
- `vasic-digital/Containers/lanes/e2e/docker-compose.yml` — the canonical full-topology compose spec (mirrors [S03 §5.1](../06_Submodules/03_Challenges_Submodule.md#51-topology-1--01_minimum_viable_session-the-canonical-exemplar) at the test-fixture layer).
- `helix-e2e-journey-coverage` lint — parses the user journey definition + asserts every step is reachable from at least one E2E test.
- `helix-test-select` — selects which E2E suite to run on per-PR cadence (primary scenario only) vs nightly (full set).

E2E tests do **not** invoke `vasic-digital/Challenges` directly — that's T11's role. E2E uses `assertEqual`-style binary assertions rather than baseline-parity. This is the boundary: if a test wants baseline parity (e.g. VMAF ≥ 95), it belongs in T11.

---

## 4. The Reference User Journey (System Overview §3)

The canonical journey ratified at [System Overview §3](../02_System_Overview.md#3-reference-user-journey) and reproduced here for E2E coverage:

1. Cold-start — host agent + clients up.
2. User login — OAuth, attestation token issuance, Vault DEK acquisition.
3. Catalog browse — tenant-scoped catalog query via gRPC.
4. Title selection — DRM/licence check.
5. Session start — capture init, encoder cold start, transport ICE/QUIC handshake.
6. Streaming begin — first frame rendered on client within 200 ms of session-start.
7. Controller input — round-trip echo within 8 ms p999.
8. Sustained playback — 60-second steady state.
9. Audio synchronisation — A/V sync within 40 ms p999.
10. Session pause / resume — server-side state survives.
11. Session end — graceful teardown; billing event emitted.
12. Session disconnect — client-side teardown; resources released.

Each step has at least one E2E test under `tests/e2e/<step-name>_test.go`. The `helix-e2e-journey-coverage` lint enforces this.

---

## 5. Per-Submodule Expectations

Each submodule's [S05 §5 Test matrix entry](../06_Submodules/per-submodule/) commits to its E2E surface:

- **`helix-r18-safeexec`** — E2E invokes `r18.SafeExec(ctx, "true")` and `r18.SafeExec(ctx, "pm-suspend")` against a real container; asserts the deny-list rejection fires.
- **`helix-grpc-frame`** — E2E starts a real `helix-grpc-frame` server with a real client; exchanges real `FrameMetadata` messages at 60 fps for 60 seconds; asserts no dropped frame, OTLP traces present.
- **`helix-pipeline`** — E2E executes the full reference user journey; pass = every journey step passes its sub-assertion.
- **`helix-record`** — E2E records a 1-second clip via fMP4; verifies file exists, parses with FFmpeg, decodes back to byte-exact match (modulo encoder non-determinism).

The 29 submodules' E2E rows together exercise every public API on the journey path.

---

## 6. Anti-Pattern Catalogue

### 6.1 E2E that mocks "just one service"

```go
// tests/e2e/login_test.go
//go:build e2e

func TestLogin(t *testing.T) {
    realVault := testcontainers.Start(...)
    fakeRateLimiter := NewMockRateLimiter()  // forbidden
    ...
}
```

R-12 forbids mocks at this layer. If the rate limiter is hard to start as a real service, that's a deployment defect — fix the deployment.

### 6.2 E2E that depends on external network

```go
client := http.Get("https://api.production.example.com/...")
```

E2E tests must be **hermetic**. Production APIs are external dependencies that flake; they're forbidden. Boot a stub-server container if needed (a stub-server is **not** a mock — it's a real HTTP server returning canned responses; the client doesn't know the difference).

### 6.3 E2E with hardcoded port

```go
client := vault.New("http://localhost:8200", ...)
```

Port-conflict races break under parallel test execution. Use `container.Endpoint(ctx, "")` to get a dynamically-assigned port.

### 6.4 E2E that doesn't tear down

```go
func TestSession(t *testing.T) {
    container, _ := testcontainers.Start(...)
    // no t.Cleanup or container.Terminate
    ...
}
```

Resource leak. The next test in the suite gets contaminated. Always `t.Cleanup`.

### 6.5 E2E with broad `assert.NoError(err)` and nothing else

```go
err := client.StartSession(ctx, req)
assert.NoError(t, err)
// no further assertions
```

NoError says "the function returned" — not "the system did the right thing." Add output assertions: did the session ID come back? Did the billing event fire? Did the OTLP trace span appear?

### 6.6 E2E that retries on flake

```go
for i := 0; i < 3; i++ {
    if err := test(); err == nil {
        return
    }
}
t.Fatal("flaked")
```

Retrying hides flakes. Investigate the root cause; an E2E test that flakes 1-of-3 has a real bug that will show up in production.

---

## 7. CI Lane Invocation Pattern

```yaml
- name: E2E primary scenario
  if: matrix.test-type == 'e2e' && github.event_name == 'pull_request'
  steps:
    - name: Start full topology
      run: |
        cd tests/e2e
        docker compose -f topology-minimum.yml up -d --wait
    - name: Run E2E primary
      run: |
        go test -tags=e2e -run=TestPrimaryUserJourney -timeout=15m ./tests/e2e/...
    - name: Tear down
      if: always()
      run: |
        cd tests/e2e
        docker compose -f topology-minimum.yml down -v

- name: E2E full secondary set
  if: matrix.test-type == 'e2e' && github.event_name == 'schedule'
  steps:
    - name: Start full topology
      run: docker compose -f topology-full.yml up -d --wait
    - name: Run E2E full set
      run: go test -tags=e2e -timeout=30m ./tests/e2e/...
    - name: Tear down
      if: always()
      run: docker compose -f topology-full.yml down -v
```

Per-PR runs the primary scenario only (~ 3 min); nightly runs the full set (~ 25 min). The selection is driven by `github.event_name`.

---

## 8. The Cross-Family Boundary — E2E vs Challenges

E2E and Challenges share a topology (the canonical full-system compose). They differ in **observation strategy**:

| Property                  | E2E                                  | Challenges                                |
|---------------------------|--------------------------------------|-------------------------------------------|
| Topology                  | Full system                          | Full system                               |
| Pass/fail                 | Binary assertion                     | Baseline parity (per S03 §6.3)            |
| Owner                     | Per-submodule                        | HelixQA + S03                             |
| Cadence                   | Per-PR (primary) + nightly (full)    | Nightly + canary + pre-release           |
| Recorded baseline?        | No                                   | Yes (signed manifest)                     |
| Change-point detection?   | No                                   | Yes (Mann-Whitney U + KS)                 |
| Scope                     | One journey                          | Many topologies × many scenarios          |

A submodule's E2E suite catches obvious regressions (the test errored or asserted something wrong). Its Challenges suite catches subtle drift (VMAF dropped from 95 to 90, Mann-Whitney U flagged the latency histogram shifted). Together they cover the gap that Unit + Integration leave per R-13.

---

## 8a. The Canonical Topology Compose

The full-system compose for E2E lives at `vasic-digital/Containers/lanes/e2e/topology-minimum.yml`. Excerpt:

```yaml
# tests/e2e topology-minimum.yml
version: "3.9"
networks:
  e2e:
    driver: bridge
    ipam:
      config:
        - subnet: 172.30.0.0/16
services:
  cockroach:
    image: cockroachdb/cockroach@sha256:9a3...
    command: start-single-node --insecure
    healthcheck:
      test: ["CMD", "/cockroach/cockroach", "sql", "--insecure", "-e", "SELECT 1"]
      interval: 5s
      retries: 12
    networks: [e2e]
  vault:
    image: hashicorp/vault@sha256:7e8...
    cap_add: [IPC_LOCK]
    environment:
      VAULT_DEV_ROOT_TOKEN_ID: e2e-root
    healthcheck:
      test: ["CMD", "vault", "status"]
      interval: 3s
      retries: 10
    networks: [e2e]
  nats:
    image: nats@sha256:f1a...
    networks: [e2e]
  redis:
    image: redis@sha256:c4d...
    networks: [e2e]
  helix-host-agent:
    image: ghcr.io/vasic-digital/helix-pipeline:test
    depends_on:
      cockroach: { condition: service_healthy }
      vault: { condition: service_healthy }
      nats: { condition: service_started }
      redis: { condition: service_started }
    networks: [e2e]
    cap_add: [SYS_NICE, BPF]   # per S02 §8.4 carve-outs
    devices:
      - /dev/dri:/dev/dri        # GPU device passthrough
  helix-client:
    image: ghcr.io/vasic-digital/helix-client:test
    depends_on:
      helix-host-agent: { condition: service_started }
    networks: [e2e]
  xvfb:
    image: ghcr.io/vasic-digital/containers/xvfb:test
    networks: [e2e]
```

The compose is the source of truth; T11 Challenges adds Toxiproxy + the recording layer + the harness orchestrator on top.

## 8b. Per-Step Test File Layout

The 12 user-journey steps from §4 each map to one test file:

```
tests/e2e/
├── topology-minimum.yml
├── 01_cold_start_test.go
├── 02_login_test.go
├── 03_catalog_browse_test.go
├── 04_title_selection_test.go
├── 05_session_start_test.go
├── 06_streaming_begin_test.go
├── 07_controller_input_test.go
├── 08_sustained_playback_test.go
├── 09_audio_synchronisation_test.go
├── 10_pause_resume_test.go
├── 11_session_end_test.go
└── 12_disconnect_test.go
```

`helix-e2e-journey-coverage` parses [System Overview §3](../02_System_Overview.md#3-reference-user-journey) and asserts a corresponding test file + at least one test function per step.

## 9. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T04-A            | Should E2E + T11 share the same compose spec, or should T11 own its own?                                       | T11 Challenges chapter                              |
| OQ-T04-B            | Per-PR E2E budget — keep at primary-only, or extend to a 5-minute "extended primary" set?                     | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T04-C            | Stub servers vs. real upstream — when is a stub-server acceptable?                                             | Constitution §13 *Exceptions*                       |

---

## 10. References & Anti-Bluff Verification

### 10.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.3.
- [`../02_System_Overview.md`](../02_System_Overview.md) §3.
- [`../06_Submodules/03_Challenges_Submodule.md`](../06_Submodules/03_Challenges_Submodule.md) §5.

### 10.2 External (web)

- Docker Compose `--wait`: https://docs.docker.com/compose/reference/up/ (accessed 2026-04-30).
- Go testing build tags: https://pkg.go.dev/cmd/go#hdr-Build_constraints (accessed 2026-04-30).

### 10.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §9 open questions named with deferred resolution chapters.
- §6 anti-pattern catalogue codifies the failure modes that E2E (vs Integration) catches.
- §8 cross-family boundary explicitly documents what E2E does *not* do (no baseline parity, no change-point detection); Challenges (T11) does that.

### 10.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/04_E2E_Tests.md` — 2026-04-30.
