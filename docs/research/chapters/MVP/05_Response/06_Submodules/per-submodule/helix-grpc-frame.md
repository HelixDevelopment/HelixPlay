# `helix-grpc-frame` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-grpc-frame`                                                                                                     |
| **Origin chapter:section**  | [C06 §6](../../03_Architecture/05_RealTime_APIs.md) — *Real-Time gRPC streaming for control-plane + frame metadata*    |
| **Public path (4 mirrors)** | `vasic-digital/helix-grpc-frame` on GitHub + GitLab + GitFlic + GitVerse                                                |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `google.golang.org/grpc`, `google.golang.org/protobuf`, `connectrpc.com/connect`, `coder/websocket`, `golang.org/x/net/http2` |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `grpc-framing-1.x` — builder `golang-builder`, runtime `distroless-static`                                          |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/02_multi_session_single_host/scenarios/01_grpc_streaming_under_burst_load.scenario.yaml`                |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision (similar public OSS exists outside `vasic-digital`, e.g. `connectrpc.com/connect`, but those are upstream dependencies, not org-internal duplicates). |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-grpc-frame` is the **real-time gRPC framing primitive** for HelixPlay's control plane. The submodule wraps `connectrpc.com/connect` (per [C06 §6](../../03_Architecture/05_RealTime_APIs.md)'s CZ-RA1 resolution — Connect-Go over Gorilla gRPC) and exports HelixPlay-specific framing helpers that are reused across the host agent, the clients, and the host-control microservice. The framing covers (a) bidirectional streaming for session-control RPCs, (b) per-frame metadata side-channel that rides alongside the WebRTC media stream, (c) HTTP/3 transport via `golang.org/x/net/http2` + `quic-go` once HTTP/3 is GA in the Go ecosystem, and (d) `coder/websocket` adapter for browser clients that cannot terminate gRPC directly.

The submodule was introduced in C06 §6 specifically to consolidate the "we do gRPC three different ways across the codebase" duplication that earlier projects (HelixAgent, Catalogizer) accumulated. R-04 mandates reuse over duplication, and S01 §3.1 names this submodule as the canonical landing for any gRPC-framing logic.

`★ Why Connect-Go and not native gRPC.` C06 §6 documents the four-way decision: Connect-Go's HTTP/2 + JSON + Protobuf trifecta beats Gorilla's WebSocket-only model on debuggability (a `curl` against a Connect endpoint works without protoc-gen-grpc-web) and beats native gRPC on browser support (Connect transparently downgrades to HTTP/1.1 + length-delimited streaming when HTTP/2 is unavailable). HelixPlay's four-mirror topology means some clients land on Russian-jurisdiction networks where HTTP/2 quirks have surfaced; Connect's degradation path matters operationally.

---

## 2. Public API Surface

The full API is documented in C06 §6. This section reproduces the high-level shape.

### 2.1 The `FrameStream` bidirectional stream type

```go
package grpcframe

// FrameStream is a bidirectional Connect stream specialised for
// HelixPlay's per-frame metadata side-channel. It carries the
// presentation timestamp, the encoder's QP target, the dual-rung
// NAL type (per helix-dualpath C29 §6), and the controller-input
// round-trip echo for the corresponding frame.
type FrameStream struct {
    // (private fields)
}

func NewFrameStream(ctx context.Context, conn *connect.BidiStream[FrameMetadata, FrameAck]) *FrameStream
func (s *FrameStream) Send(meta FrameMetadata) error
func (s *FrameStream) Recv() (FrameAck, error)
func (s *FrameStream) Close() error
```

### 2.2 The `ControlClient` interface

```go
package grpcframe

// ControlClient is the high-level client for HelixPlay's control
// plane. It wraps a Connect client and exposes session-lifecycle
// methods plus the per-frame metadata stream.
type ControlClient interface {
    StartSession(ctx context.Context, req *StartSessionRequest) (*StartSessionResponse, error)
    EndSession(ctx context.Context, req *EndSessionRequest) (*EndSessionResponse, error)
    Frames(ctx context.Context) (*FrameStream, error)
    Health(ctx context.Context) (*HealthResponse, error)
}

func NewControlClient(addr string, opts ...ClientOption) (ControlClient, error)
```

### 2.3 The `Server` registration helper

```go
package grpcframe

// Server wraps a connect.NewServer with HelixPlay-specific
// middleware: OTLP tracing, R-18 wrapper compliance check,
// per-tenant rate limiting (delegated to helix-tenant), Brotli
// compression negotiation.
type Server struct { /* ... */ }

func NewServer(impl ControlServer, opts ...ServerOption) (*Server, error)
func (s *Server) Mux() *http.ServeMux
```

### 2.4 The `WebSocketAdapter` for browser clients

```go
package grpcframe

// WebSocketAdapter wraps coder/websocket so a browser client can
// reach the gRPC streaming endpoint without protoc-gen-grpc-web.
// The adapter terminates the WebSocket frame and re-emits the
// payload as a Connect bidirectional stream.
type WebSocketAdapter struct { /* ... */ }

func NewWebSocketAdapter(server *Server) http.Handler
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for the rare subprocess invocation (e.g. `helixctl --version` health-check forwarding). The grpc framing layer does not invoke subprocesses in the hot path; the dependency is for the boundary-case forwarding only.

### 3.2 External (Go)

- `google.golang.org/grpc` — gRPC reference implementation (used for the protobuf encoder + the gRPC-native interop fallback).
- `google.golang.org/protobuf` — protobuf runtime.
- `connectrpc.com/connect` — primary RPC layer (per C06 §6 CZ-RA1).
- `coder/websocket` — WebSocket library for browser-client adapter (per C06 §6 CZ-RA2; chosen over Gorilla for its smaller surface and better timeout handling).
- `golang.org/x/net/http2` — HTTP/2 stack for streaming.
- `github.com/quic-go/quic-go` — HTTP/3 transport (per C06 §6 CZ-RA3; pinned to a release that has shipped a stable Go 1.24-compatible build).
- `github.com/andybalholm/brotli` — Brotli compression (per Constitution §4 *Communication Stack*).
- `go.opentelemetry.io/otel` + `go.opentelemetry.io/otel/sdk` + `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` — OTLP tracing.

#### Version-pinning rationale

Each direct dependency carries a specific version-floor the submodule depends on:

- `connectrpc.com/connect ≥ v1.16.0` — required for the bidirectional-stream cancellation semantics fixed in v1.16; earlier versions leak a goroutine on client cancel.
- `coder/websocket ≥ v1.8.13` — required for the `WriteContext`-aware close-frame handling that prevents the WebSocket-1006 spurious-close bug seen in earlier releases.
- `quic-go ≥ v0.51.0` — required for the Go 1.24 cgo-free build (earlier versions required cgo for AES-GCM hardware acceleration; Go 1.24's stdlib AES is now competitive).
- `golang.org/x/net ≥ v0.40.0` — security floor; pre-v0.40 has CVE-2024-45337 in HTTP/2 stream-priority handling.
- `andybalholm/brotli` — pinned to upstream's latest tag; the project follows semver loosely, so Renovate opens a PR per release for explicit operator review.

#### Transitive risk surface

The dependency closure includes ≈ 80 transitive modules; SBOM emission (S01 §4.4) tracks them all. The largest risk surface is `golang.org/x/net` (CVEs cluster around HTTP/2 stream-priority); `golang.org/x/crypto` is second (TLS cipher-suite drift); `connectrpc.com/connect` is third (relatively young — first GA was 2023, so its API is still hardening). Renovate scans these on a 4-hour cadence; high-severity CVE PRs are auto-assigned to the on-call.

### 3.3 External (C / system)

None at runtime. The build is pure-Go.

#### Build-time tooling

- `protoc ≥ v25.3` — generates Go protobuf bindings from `.proto` files. The `vasic-digital/Containers/builders/golang-builder` image ships a pinned protoc; per-PR builds use this rather than the developer's host-installed version.
- `protoc-gen-go ≥ v1.34.0` — Go binding generator.
- `protoc-gen-connect-go ≥ v1.16.0` — Connect-specific RPC generator.
- `buf ≥ v1.32.0` — Protobuf lint + breaking-change detection (per Constitution §12 *Documentation Discipline* — the `.proto` files are the documented API contract, and `buf breaking` runs on every PR to detect accidental API breakage).

---

## 4. Container Build (S02 §3 lane: `grpc-framing-1.x`)

**Builder:** `golang-builder` (no cgo). **Runtime:** `distroless-static`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Signing chain:** cosign + SLSA L3 + cyclonedx + syft per S02 §6. **Hardening:** `--user 65532:65532`, `--read-only`, seccomp default-deny, `--cap-drop=ALL`, `--memory=1g`, `--cpus=1.0` (gRPC server is more memory-hungry than the safeexec wrapper).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit (mocks allowed)

- Encoder / decoder round-trips for every protobuf message type.
- Connect client / server pairs with mock transport.
- Brotli compression negotiation.
- WebSocket adapter framing logic.

### 5.2 Integration (no mocks)

- Real Connect server + real Connect client over `127.0.0.1:0` (random port).
- Real WebSocket adapter against `coder/websocket`'s test fixtures.
- Real OTLP collector receiving traces.

### 5.3 E2E (no mocks)

End-to-end against a real `helix-grpc-frame` server with a real client, exchanging real `FrameMetadata` messages at 60 fps for 60 seconds.

### 5.4 Security

govulncheck (zero high findings expected; periodic findings in `golang.org/x/net/http2` patched on Renovate cadence). Snyk (zero high). Trivy (zero on distroless-static).

### 5.5 Benchmarking

Round-trip latency p999 ≤ 1 ms on `127.0.0.1`; throughput ≥ 60 K frames/s on amd64; 30 K frames/s on arm64.

### 5.6 Chaos

Toxiproxy-injected packet loss (1 %, 5 %, 10 %); the framing layer must surface the loss via OTLP traces and (for FrameStream) emit a NACK that triggers `helix-dualpath`'s rung descent.

### 5.7 Stress

24-hour run at 60 fps × 1,000 concurrent FrameStreams; verify no goroutine leak, no fd leak, no memory growth above 100 MiB.

### 5.8 Smoke

30-second post-deploy: Health() RPC against the running service; verify HTTP/2 negotiation succeeded and the response carries the expected build SHA.

### 5.9 Full Automation

§5.1–§5.8 in CI matrix order.

### 5.10 Challenges

`02_multi_session_single_host/01_grpc_streaming_under_burst_load.scenario` — single host with N concurrent FrameStreams under burst load (50 → 500 streams in 10 seconds). Baseline records p999 latency, OTLP trace topology, and the rate at which Brotli compression chose `q=11` vs `q=4`.

---

## 6. Challenges Entry-Point (S03 §4 row #02)

**Topology:** `02_multi_session_single_host`. **Scenario:** `01_grpc_streaming_under_burst_load.scenario.yaml`. **Why this scenario.** gRPC framing's failure mode at scale is connection-pool exhaustion; the burst-load scenario specifically targets the connection-pool sizing in the framing layer's HTTP/2 transport. **Baseline:** p999 latency under burst (≤ 50 ms), Brotli compression ratio (≥ 60 %), connection-pool growth curve (linear up to N=500, then plateau).

**Change-point detection:** Mann-Whitney U on latency histograms (p > 0.01 to pass), exact equality on connection-pool growth shape, ≥ 95 % Brotli compression ratio match.

---

## 7. R-18 Inheritance

`helix-grpc-frame` imports `helix-r18-safeexec` for its subprocess invocations (boundary cases — `helixctl --version` forwarding, `git rev-parse HEAD` for build-SHA reporting). The hot path (gRPC stream handling) does **not** invoke subprocesses; the R-18 surface is therefore minimal and well-bounded.

Per [S01 §7.3](../01_Submodule_Catalog.md#73-the-five-layer-enforcement-per-c08-§10--§115) layer 4 (ripgrep CI lint), the per-PR check `rg -nE 'os/exec|exec\.Command|exec\.CommandContext'` runs against this submodule's source tree and verifies every match is wrapped in `r18.SafeExec`. As of 2026-04-30, the only matches are in `internal/healthcheck/version.go` and `internal/build/sha.go`, both correctly wrapped.

---

## 8. Release-Train Cadence (S01 §9)

Currently `v0.x.y`. Graduation criteria same as `helix-r18-safeexec`'s §8.2 (S01 §9.2). The connect-go API stability is upstream-dependent; HelixPlay's `v1.0.0` graduation tracks Connect-Go's own API stability (Connect-Go is at `v1.x.y` upstream; HelixPlay's adapter can graduate independently once HelixPlay's own surface is frozen).

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                      | Default                | Range / type                        | Purpose                                                                 |
|------------------------------|------------------------|-------------------------------------|-------------------------------------------------------------------------|
| `HELIX_GRPC_LISTEN_ADDR`     | `:8443`                | `<host>:<port>`                     | Bind address for the Connect server.                                    |
| `HELIX_GRPC_MAX_STREAMS`     | `1024`                 | int [16, 65536]                     | Max concurrent FrameStream connections per server instance.             |
| `HELIX_GRPC_BROTLI_Q`        | `4`                    | int [0, 11]                         | Brotli compression level for hot-path responses (q=4 is C06 §6 default).|
| `HELIX_GRPC_KEEPALIVE_S`     | `30`                   | int seconds                         | HTTP/2 keepalive interval; set to `0` to disable.                       |
| `HELIX_GRPC_OTLP_ENDPOINT`   | `localhost:4317`       | `<host>:<port>`                     | OTLP collector endpoint per Constitution §10.                           |
| `HELIX_GRPC_HTTP3_ENABLED`   | `false`                | bool                                | Opt-in to HTTP/3 transport via `quic-go` (off by default until OQ-A resolved). |
| `HELIX_GRPC_TLS_CERT_FILE`   | (required)             | filesystem path                     | TLS cert; container expects mount at `/etc/helix/tls/cert.pem`.         |
| `HELIX_GRPC_TLS_KEY_FILE`    | (required)             | filesystem path                     | TLS private key.                                                        |
| `HELIX_GRPC_TENANT_LOADER`   | `helix-tenant`         | submodule name                      | Tenant resolver implementation; references row #05.                     |

### 9.2 Performance budget (`v0.x.y`)

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `Health()` round-trip                 | 200 µs   | 800 µs  | 2 ms    | Localhost; on-net to the host agent.                             |
| `FrameStream.Send()` enqueue          | 3 µs     | 8 µs    | 15 µs   | Lock-free path; consumer drains in goroutine.                    |
| `FrameStream.Recv()` blocking         | 50 µs    | 200 µs  | 500 µs  | Includes one HTTP/2 frame round-trip.                            |
| Connection-pool growth (50 → 500)     | 8 s      | 12 s    | 18 s    | Linear ramp; budget allows for HTTP/2 multiplexing reservation.  |
| Brotli compression (q=4) ratio        | ≥ 60 %   | ≥ 55 %  | ≥ 50 %  | On a typical FrameMetadata payload (≈ 800 bytes plain).         |
| Memory per FrameStream                | 4 KiB    | 8 KiB   | 16 KiB  | Steady-state; transient bursts during reconfig are tolerated.    |

The benchmark suite (`tests/benchmarking/`) records these metrics on every PR and the run-archive (S04 §7) tracks the trend. A regression > 5 % on any p999 column triggers a P3 alert per S04 §8.

### 9.3 Common errors and remediation

| Error                                 | Cause                                              | Remediation                                                           |
|---------------------------------------|----------------------------------------------------|------------------------------------------------------------------------|
| `connect: code = ResourceExhausted`   | `HELIX_GRPC_MAX_STREAMS` ceiling hit               | Raise the env var; horizontal-scale the server instances.             |
| `connect: code = DeadlineExceeded`    | Client deadline shorter than server-side processing | Lengthen client deadline; investigate slow handlers via OTLP traces.  |
| `connect: code = Unauthenticated`     | TLS cert expired or AppRole auth refused           | Renew cert via Vault; rotate AppRole secret; verify clock skew < 60 s.|
| `WebSocket close code 1006`           | Browser-side connection lost without close frame   | Client reconnects automatically; monitor frequency in dashboard.     |
| `quic-go: TLS handshake failed`       | HTTP/3 enabled but TLS 1.3 ALPN missing            | Disable HTTP/3 (`HELIX_GRPC_HTTP3_ENABLED=false`) until OQ-A closes. |

### 9.4 Migration from inline gRPC handlers (pre-`helix-grpc-frame`)

A consumer migrating from inline `google.golang.org/grpc.NewServer()` to `helix-grpc-frame.NewServer()`:

1. Replace `grpc.NewServer(...)` with `grpcframe.NewServer(impl, opts...)`.
2. Register the implementation via `RegisterControlServer` (auto-generated from the `.proto` file).
3. Drop hand-rolled OTLP middleware — `grpcframe.NewServer` registers it automatically.
4. Drop hand-rolled rate-limiting — `grpcframe.NewServer` delegates to `helix-tenant`.
5. Replace `ctx, cancel := context.WithDeadline(...)` deadlines on the server side; the framing layer surfaces deadlines via per-RPC `connect.Spec().Timeout` reads.
6. Update CI to invoke `subprocess-wrapper-1.x` lane fixtures rather than the prior bespoke fixtures.

The migration is documented in `docs/migration-from-inline-grpc.md` in the submodule's repo.

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-grpc-frame-A     | HTTP/3 stability — when does `quic-go` graduate to a Go 1.24-compatible `v1.0.0` upstream?                     | C06 §6 next revision                                |
| OQ-grpc-frame-B     | WebSocket adapter — should it use Connect-Go's native HTTP/1.1 fallback instead?                               | C06 §6 next revision                                |
| OQ-grpc-frame-C     | Brotli compression `q` level — fixed `q=4` for hot path, `q=11` for cold catalog responses?                    | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../03_Architecture/05_RealTime_APIs.md`](../../03_Architecture/05_RealTime_APIs.md) §6 | (slice) | 2026-04-30 | origin chapter; full API surface              |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #02, R-18 inheritance              |
| [`../02_Containers_Submodule.md`](../02_Containers_Submodule.md) §3 |   627 | 2026-04-30 | container lane                                  |
| [`../03_Challenges_Submodule.md`](../03_Challenges_Submodule.md) §4 |   517 | 2026-04-30 | scenario                                        |
| [`helix-r18-safeexec.md`](helix-r18-safeexec.md)                   | (this batch) | 2026-04-30 | dependency descriptor                  |

Forbidden patterns: clean. The three §9 open questions are explicitly named with deferred resolution chapters.

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-grpc-frame.md` — 2026-04-30.
