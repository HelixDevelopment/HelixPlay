# T03 — Integration Tests

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.2, §7 (Mock policy enforcement), §10 (Tooling lockstep); [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §5; [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3 (per-submodule containers).
> **Source line count:** ≈ 1k+ lines from family inputs; chapter floor is 300 lines per Master Plan §7.2 row T03.
> **Chapter targets:** R-11 (every public exported symbol exercised), R-12 (no mocks).
> **Cross-links:** [`02_Unit_Tests.md`](02_Unit_Tests.md), [`04_E2E_Tests.md`](04_E2E_Tests.md), [`07_Chaos.md`](07_Chaos.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Integration is **test-type 2 of 10** in the [T01 §2 grid](01_Test_Matrix.md#2-the-29--10--4-cell-grid-the-canonical-lookup). It is the **first** test type where R-12 forbids mocks: every dependency must be the real thing — a real database, a real Redis, a real co-located submodule binary, a real network, a real audio device (or a virtual one that's not a mock).

Integration's role: **dependency-glue correctness**. A Unit test mocks `helix-vault`'s HTTP client and verifies the calling code handles every HTTP-status response correctly. An Integration test boots a real HashiCorp Vault container and verifies the calling code's HTTP client actually makes the requests it claims to make and parses the responses Vault actually returns. These differ when the mock misrepresents the real dependency — the most common source of past anti-bluff failures.

Integration runs at every cadence (per-PR + nightly + canary + pre-release).

---

## 2. The Discipline — What Makes "Green" Green

A green Integration row means:

1. **Every public exported symbol is exercised at least once with a real dependency**, verified by `helix-integration-coverage` lint that walks the godoc and matches against test invocations.
2. **No `*mock*` import** under `tests/integration/` (verified by `helix-mock-discipline`).
3. **All assertions pass** with the same no-skip / no-flake-tolerance rules as Unit (T02 §2).
4. **Real backing services up** for the duration of the test — verified by per-test `setup` + `teardown` hooks that boot containers via `docker-compose` or `podman-compose` and tear them down cleanly.
5. **Deterministic** — same input → same output under the same dependency-version set. Pinning is enforced via [S01 §4.3 dependency lockstep](../06_Submodules/01_Submodule_Catalog.md#43-dependency-lockstep--gosum-goproxy-gosumdb-renovate).

---

## 3. Tooling

- Go's `testing` package + `//go:build integration` build tag (so `go test ./...` without `-tags=integration` skips Integration; `go test -tags=integration ./tests/integration/...` runs them).
- `github.com/testcontainers/testcontainers-go` — programmatic container lifecycle for Go integration tests; pinned to ≥ v0.34.0.
- `docker-compose.yml` per-test fixture under `tests/integration/<feature>/docker-compose.yml`.
- Real binaries — Postgres / CockroachDB / Redis / NATS / Vault / Coturn / etc., pinned by digest per [S02 §5](../06_Submodules/02_Containers_Submodule.md#5-base-image-discipline-distroless-pinned-digests-no-latest).

The `helix-integration-coverage` lint is custom; it parses the submodule's godoc, extracts every `// public` symbol, and asserts at least one test invokes it.

---

## 4. Per-Submodule Expectations

Each submodule's [S05 §11 Per-test-type coverage targets](../06_Submodules/per-submodule/) entry for Integration commits to the dependency surface:

- **`helix-vault`** — real HashiCorp Vault dev-server container; KEK rotation; lazy DEK re-wrap; tenant-isolation tests.
- **`helix-shm`** — real `/dev/shm` allocation; cross-process page sharing via fd-passing over Unix sockets (two real processes, not mocked).
- **`helix-iouring`** — real ring against real `/tmp` file + real UDP socket; batched submit + completion reap.
- **`helix-grpc-frame`** — real Connect server + real Connect client over `127.0.0.1:0` (random port); WebSocket adapter against `coder/websocket`'s test fixtures.
- **`helix-codec`** — real libavutil validation against Profile parameters.
- **`helix-encoder`** — real vendor encoder (NVENC / QSV / AMF / VideoToolbox per host capability); 100-frame synthetic NV12 input; verify NAL units parse with FFmpeg.
- **`helix-pipeline`** — real per-stage submodule integration; verify Pipeline.Start brings every stage up.

The full per-submodule list is in T01 §2's grid; every `O` cell at column "Int" expects the discipline above.

---

## 5. Anti-Pattern Catalogue

### 5.1 Mock at the Integration boundary

```go
// tests/integration/vault_test.go
//go:build integration
package vault

import "github.com/stretchr/testify/mock"  // forbidden under integration tag

func TestVaultIntegration(t *testing.T) {
    mockVault := NewMockVault()  // mock under the integration tag
    ...
}
```

Forbidden by R-12 + helix-mock-discipline lint. Use `testcontainers-go` to start a real Vault instead.

### 5.2 Hardcoded production endpoint

```go
client, _ := vault.New("https://prod-vault.example.com:8200", roleID, secretID)
```

A test that talks to production is a defect. Use a containerised dev instance.

### 5.3 Test that depends on test-execution order

```go
func TestStep1(t *testing.T) { /* creates global state */ }
func TestStep2(t *testing.T) { /* expects global state from Step1 */ }
```

Tests must be independent. Use `t.Cleanup` + per-test fixtures.

### 5.4 Container-not-ready race

```go
container.Start(ctx)
client := vault.New(container.Address(), ...)  // Vault may not be listening yet
client.Encrypt(...)  // races against startup
```

Use `container.WaitFor(ctx, container.ReadyHook())` (testcontainers-go's wait strategies) before connecting.

### 5.5 Skipping under CI

```go
func TestSomething(t *testing.T) {
    if os.Getenv("CI") != "" {
        t.Skip("flaky on CI")
    }
    ...
}
```

A test that's "too flaky for CI but works locally" is a defect that fails CI in a different way; investigate the flake. Per Constitution §13, an exception requires operator approval.

---

## 6. CI Lane Invocation Pattern

```yaml
- name: Integration tests
  if: matrix.test-type == 'integration'
  services:
    docker:
      image: docker:dind
  steps:
    - name: Start dependencies
      run: |
        cd tests/integration
        docker compose up -d --wait
    - name: Run integration tests
      run: |
        go test -tags=integration -timeout=10m ./tests/integration/...
    - name: Tear down
      if: always()
      run: |
        cd tests/integration
        docker compose down -v
    - name: Integration-coverage lint
      run: helix-integration-coverage ./...
```

The `--wait` flag on `docker compose up -d --wait` blocks until each service's healthcheck reports healthy — eliminates the §5.4 race. The `if: always()` on teardown ensures a failed test still releases container resources.

---

## 7. Container-Up Patterns

### 7.1 testcontainers-go

```go
//go:build integration
package vault_integration_test

import (
    "context"
    "testing"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestVaultRoundTrip(t *testing.T) {
    ctx := context.Background()
    req := testcontainers.ContainerRequest{
        Image:        "hashicorp/vault:1.18",
        ExposedPorts: []string{"8200/tcp"},
        Env: map[string]string{
            "VAULT_DEV_ROOT_TOKEN_ID":     "test-token",
            "VAULT_DEV_LISTEN_ADDRESS":    "0.0.0.0:8200",
        },
        WaitingFor: wait.ForHTTP("/v1/sys/health").WithPort("8200/tcp"),
    }
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("Vault container start: %v", err)
    }
    t.Cleanup(func() { _ = container.Terminate(ctx) })

    addr, _ := container.Endpoint(ctx, "")
    v, err := vault.New("http://"+addr, "test-token")
    if err != nil {
        t.Fatalf("vault.New: %v", err)
    }
    // ... real round-trip tests ...
}
```

Properties:
- Real Vault binary; real HTTP traffic.
- Wait strategy ensures the container is ready before the test connects.
- `t.Cleanup` releases the container regardless of test outcome.
- No mocks anywhere.

### 7.2 docker-compose.yml fallback

For multi-service fixtures (e.g. Vault + Redis + Coturn together), a `tests/integration/docker-compose.yml` is more declarative. The CI lane invocation pattern in §6 brings it up + tears it down.

---

## 7a. Wait-Strategy Matrix per Service

testcontainers-go's `wait.For*` strategies must match the service's actual readiness signal. Mismatched strategies are the most common cause of §5.4 races. The canonical mapping:

| Service                  | Wait strategy                                                              |
|--------------------------|----------------------------------------------------------------------------|
| HashiCorp Vault          | `wait.ForHTTP("/v1/sys/health").WithPort("8200/tcp")`                       |
| CockroachDB single-node  | `wait.ForSQL("26257/tcp", "postgres", connStr)` after `cockroach init`      |
| NATS                      | `wait.ForLog("Server is ready")`                                            |
| Redis                     | `wait.ForLog("Ready to accept connections tcp")`                            |
| Coturn (TURN)             | `wait.ForListeningPort("3478/udp")`                                          |
| MinIO (S3-compat)         | `wait.ForHTTP("/minio/health/ready").WithPort("9000/tcp")`                  |
| OpenTelemetry collector   | `wait.ForListeningPort("4317/tcp")` (OTLP gRPC)                             |
| Toxiproxy (impairment)    | `wait.ForHTTP("/version").WithPort("8474/tcp")`                              |
| Xvfb (virtual X11)        | `wait.ForListeningPort("6000/tcp")`                                          |

A submodule that boots a service not listed above must add it to the `vasic-digital/.github/integration-wait-strategies.md` registry + propose a CR for this matrix to be updated.

## 7b. Container-Image Pin Audit

Per [S02 §5](../06_Submodules/02_Containers_Submodule.md#5-base-image-discipline-distroless-pinned-digests-no-latest), every Integration test's container reference is a digest pin (no floating tags). Excerpt of the canonical pin set as of 2026-04-30:

```yaml
# tests/integration/.image-pins.yaml
hashicorp_vault:    "hashicorp/vault@sha256:7e8...redacted...4d2"
cockroachdb:         "cockroachdb/cockroach@sha256:9a3...redacted...b1f"
redis:               "redis@sha256:c4d...redacted...8e7"
nats:                "nats@sha256:f1a...redacted...9c3"
minio:               "minio/minio@sha256:8b2...redacted...d4a"
otel_collector:      "otel/opentelemetry-collector@sha256:5e9...redacted...7f1"
toxiproxy:           "shopify/toxiproxy@sha256:3c8...redacted...6a4"
```

`helix-image-pin-audit` runs nightly; a digest drift triggers a Renovate PR for explicit operator review (Constitution §15 *Amendment Procedure* applies because base-image drift is operationally equivalent to dependency drift).

## 7c. Cross-Mirror Image Cache Strategy

Integration tests pull images on every fresh CI runner. A four-mirror runner topology means the same image is pulled four times (once per mirror) unless cached. Per [S02 §10](../06_Submodules/02_Containers_Submodule.md#10-container-ci-lane-cost-model--cache-strategy):

- Each mirror's CI runner pool has a **registry-side mirror cache**: GitHub runners pull from `ghcr.io/vasic-digital/...`; GitLab runners pull from `registry.gitlab.com/vasic-digital/...`; GitFlic runners from `registry.gitflic.ru/...`; GitVerse runners from `registry.gitverse.ru/...`.
- Per [S02 §4 multiarch publishing](../06_Submodules/02_Containers_Submodule.md#4-multi-arch-image-strategy-linuxamd64--linuxarm64-default), every release pushes the same image (and same digest) to all four registries, so a CI runner's local mirror cache always serves the right digest.
- The fallback cross-mirror pull (e.g. GitFlic runner pulling from `ghcr.io` when its local registry hasn't replicated yet) is logged + alerted via `four-mirror-replication-audit.sh`; this is rare and indicates a replication delay rather than a structural problem.

The strategy means Integration tests on each mirror complete in similar wall-clock; mirror divergence in test duration triggers a P3 alert per [S04 §8](../06_Submodules/04_HelixQA_Integration.md#8-alert-routing-and-on-call-rotation).

## 7d. The Test-Database Migration Discipline

Integration tests against CockroachDB / Postgres MUST run database migrations against a fresh database per test session. Fixed-schema or shared-database fixtures lead to test-execution-order coupling (per §5.3). The pattern:

```go
//go:build integration
package storage_integration_test

func TestMain(m *testing.M) {
    ctx := context.Background()
    container := startCockroach(ctx)              // testcontainers-go
    defer container.Terminate(ctx)

    db := connectAndMigrate(ctx, container)       // applies migrations from migrations/*.sql
    code := m.Run()
    container.Terminate(ctx)
    os.Exit(code)
}

func connectAndMigrate(ctx context.Context, c testcontainers.Container) *sql.DB {
    addr, _ := c.Endpoint(ctx, "")
    db, _ := sql.Open("postgres", "postgres://root@"+addr+"?sslmode=disable")
    migrate.Up(db, "migrations/")  // helix-storage's migration tool
    return db
}
```

The helper `helix-storage-migrate` is a thin wrapper over `golang-migrate/migrate` (pinned to ≥ v4.18.0). Migrations are committed under `<submodule>/migrations/<NNN>_<name>.up.sql` + `.down.sql`; the `.down.sql` is verified by `helix-down-migration-audit` to exist for every up-migration so rollback paths are testable.

## 8. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T03-A            | testcontainers-go vs docker-compose — operator preference for new submodules?                                 | `08_Operations/01_Container_CI_CD.md`              |
| OQ-T03-B            | Integration-coverage lint — godoc-driven or explicit annotation-driven?                                        | T10 Full Automation chapter                         |
| OQ-T03-C            | Per-mirror Integration parity — are container images replicated to all four mirrors fast enough?              | `08_Operations/01_Container_CI_CD.md`              |

---

## 9. References & Anti-Bluff Verification

### 9.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.2, §7, §10.
- [`../06_Submodules/02_Containers_Submodule.md`](../06_Submodules/02_Containers_Submodule.md) §3, §5.
- [`../01_Constitution.md`](../01_Constitution.md) §6.2.

### 9.2 External (web)

- testcontainers-go: https://golang.testcontainers.org/ (accessed 2026-04-30).
- Docker Compose: https://docs.docker.com/compose/ (accessed 2026-04-30).
- Go test build tags: https://pkg.go.dev/cmd/go#hdr-Build_constraints (accessed 2026-04-30).

### 9.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §8 open questions named with deferred resolution chapters.
- §5 anti-pattern catalogue + §6 CI lane pattern operationalise R-12 (no mocks at this layer).

### 9.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/03_Integration_Tests.md` — 2026-04-30.
