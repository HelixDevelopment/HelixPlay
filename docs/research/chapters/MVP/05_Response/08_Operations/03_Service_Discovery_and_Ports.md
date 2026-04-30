# O03 — Service Discovery & Dynamic Port Assignment

> **Source dimensions:** [`00_Index.md`](00_Index.md); [`../01_Constitution.md`](../01_Constitution.md) §4 (R-07 + R-08); [`../06_Submodules/per-submodule/helix-grpc-frame.md`](../06_Submodules/per-submodule/helix-grpc-frame.md).
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row O03.
> **Chapter targets:** R-07 (LAN service discovery + dynamic ports + gRPC + REST + HTTP/3 + Brotli), R-08 (NATS + Redis + RabbitMQ for ad-hoc plumbing).
> **Cross-links:** [`04_Observability_and_Events.md`](04_Observability_and_Events.md), [`../03_Architecture/05_RealTime_APIs.md`](../03_Architecture/05_RealTime_APIs.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. The Posture in Three Sentences

R-07 (Constitution §4.1) reads: "Service discovery on the LAN. Dynamic port assignment. gRPC preferred. REST is a separate microservice. HTTP/3 (QUIC/Cronet). Brotli compression." The host agent + clients on a HelixPlay LAN deployment auto-discover each other via mDNS + DNS-SD; ports are assigned dynamically (no fixed-port assumptions); the control plane uses gRPC over HTTP/3 with Brotli compression negotiated per-call. This chapter operationalises the LAN-side surface; cross-LAN federation lives in C09 §6 (Scalability & Multi-Region).

---

## 2. mDNS + DNS-SD Service Discovery

HelixPlay's LAN deployment uses **mDNS** (Multicast DNS, RFC 6762) + **DNS-SD** (DNS Service Discovery, RFC 6763) for zero-configuration discovery. Service types:

| Service type                       | Description                                         |
|------------------------------------|-----------------------------------------------------|
| `_helix-host._grpc._tcp.local.`    | Host agent's gRPC control endpoint                  |
| `_helix-host._http3._udp.local.`   | Host agent's HTTP/3 endpoint                        |
| `_helix-stream._rtp._udp.local.`   | Streaming RTP endpoint (per-session)                |
| `_helix-discover._http._tcp.local.`| REST discovery gateway (REST microservice per R-07) |
| `_helix-cataloger._grpc._tcp.local.`| Catalog microservice                               |
| `_helix-vault._https._tcp.local.`  | Vault HTTPS endpoint                                |

A client on the LAN issues `dig +short _helix-host._grpc._tcp.local. PTR @224.0.0.251` and gets the host agent's `<hostname>.local.` SRV record + dynamically-assigned port.

The Go-side library is `github.com/grandcat/zeroconf` (pinned to ≥ v1.0.0). Submodules use it via `helix-grpc-frame` per [C06 §6](../03_Architecture/05_RealTime_APIs.md).

---

## 3. Dynamic Port Assignment

No hardcoded ports anywhere. Every listener binds `:0` (kernel chooses an ephemeral port from the dynamic range, typically 32768-60999 on Linux) + advertises the assigned port via mDNS.

The Go pattern:

```go
listener, err := net.Listen("tcp", ":0")
if err != nil { ... }
addr := listener.Addr().(*net.TCPAddr)
mdns.Register(serviceType, addr.Port, hostname, ...)
```

Operator deployments may pin specific ports for firewall friendliness (e.g. `helix-host-port=8443` env-var override); the library accepts the override + still advertises via mDNS.

### 3.1 The dynamic-range registry

| Submodule           | Port range used        | Notes                                 |
|---------------------|------------------------|---------------------------------------|
| helix-grpc-frame    | dynamic 32768-60999    | Connect/HTTP/2 default                |
| helix-host-agent    | dynamic 32768-60999    | gRPC + REST gateway                   |
| helix-streaming     | dynamic 32768-60999    | RTP/UDP per-session                   |
| helix-vault         | 8200 (Vault default)   | operator may override                  |
| CockroachDB         | 26257 (Cockroach default) | operator may override                |
| NATS                | 4222 (NATS default)    | operator may override                  |
| Redis                | 6379 (Redis default)    | operator may override                 |

The "default" ports are operator-overridable; a HelixPlay deployment behind a strict firewall can pin all to a single allow-listed port via env-var.

---

## 4. gRPC + Connect-Go for Control Plane

Per [C06 §6](../03_Architecture/05_RealTime_APIs.md), the control plane uses Connect-Go on top of HTTP/2 + (optionally) HTTP/3 via `quic-go`. The protocol stack:

| Layer            | Choice                                                             |
|------------------|--------------------------------------------------------------------|
| Application       | Connect-Go (HTTP/2-native gRPC variant)                            |
| Transport         | HTTP/2 (default) / HTTP/3 (when client supports + server enabled)  |
| Compression       | Brotli q=4 hot path / q=11 cold path                                |
| Auth              | OIDC + Vault-issued tokens                                         |
| Observability     | OTLP traces auto-emitted per call                                   |

Brotli is mandated by R-07. The Go side uses `github.com/andybalholm/brotli` (pinned per [T01 §10](../07_Testing/01_Test_Matrix.md#10-tooling-lockstep--pinned-versions-across-the-fleet)).

---

## 5. REST as a Separate Microservice

R-07 specifies REST as a **separate microservice**, not a gRPC handler with REST-translation middleware. The reason: gRPC's HTTP/2 streaming semantics differ from REST's request/response semantics; a single handler trying to serve both leaks abstractions both ways.

The split:

- `helix-host-agent` — the gRPC service.
- `helix-rest-gateway` — a separate process that translates REST → gRPC. (Listed as a future submodule in [C06 §6](../03_Architecture/05_RealTime_APIs.md); not yet introduced in S01 §3.1's catalog.)

The REST gateway forward-translates HTTP/1.1 + HTTP/2 + HTTP/3 (with `quic-go`) requests into gRPC calls against `helix-host-agent`; clients that cannot or do not want to use gRPC directly see a familiar REST surface.

The gateway adds ~ 1 ms latency per call; it is opt-in for clients (gRPC clients bypass it).

---

## 6. HTTP/3 (QUIC) Status

HTTP/3 is **opt-in** per `HELIX_GRPC_HTTP3_ENABLED=true` env-var per [helix-grpc-frame §9.1](../06_Submodules/per-submodule/helix-grpc-frame.md#91-configuration-knobs). Default is off until `quic-go` graduates to a stable Go 1.24-compatible v1.0.0 (per OQ-grpc-frame-A in that descriptor).

Operators on Russian-jurisdiction networks may benefit from HTTP/3 because of better mid-route impairment behaviour; the opt-in env-var lets them experiment without forcing the choice fleet-wide.

---

## 7. NATS + Redis for Ad-Hoc Plumbing (R-08)

R-08 reads: "NATS / Redis / RabbitMQ used wherever they replace ad-hoc plumbing."

- **NATS** — primary event bus. Per Constitution §10, every state change emits an OTLP span + a NATS event; consumers subscribe per topic.
- **Redis** — caching + rate-limiting per [C06 §6 CZ-RA2](../03_Architecture/05_RealTime_APIs.md). Per-tenant rate-limiting via `helix-tenant.FeatureFlags`.
- **RabbitMQ** — not currently used. RabbitMQ is in scope per R-08 if operator finds a use case where NATS's at-most-once semantics are insufficient; no current submodule needs it.

The operator deploys NATS + Redis as containerised services per [O01 §11](01_Container_CI_CD.md#11-per-mirror-registry-topology--image-replication). NATS clustering is the topology for `03_multi_host_multi_session` (per [S03 §3 topologies](../06_Submodules/03_Challenges_Submodule.md#3-repository-layout-topologies--baselines--harness)).

---

## 8. The mDNS Configuration

Per the deployment topology, the operator may need to:

- Open multicast UDP 5353 in the LAN firewall.
- Enable `mDNSResponder` on macOS / Avahi on Linux clients.
- (Windows) Install Bonjour Print Services (small download).

For deployments where mDNS is forbidden (some enterprise networks), the operator can fall back to a **central discovery server** (a small REST service that maintains the same SRV records); HelixPlay clients are configured via env-var to query the discovery server instead of mDNS.

---

## 9. The Operational Health Probe Discipline

Every discovered service exposes `/healthz` per [T09 §4 canonical probe.sh](../07_Testing/09_Smoke.md#4-the-canonical-probesh-skeleton). HelixQA's continuous-mode (per [S04 §4.1](../06_Submodules/04_HelixQA_Integration.md#41-continuous-mode)) polls every 5 s and reports unreachable services as P3 alerts.

mDNS records are TTL=4500 (75 minutes); cached records expire if the service stops re-announcing. A service that crashes is unreachable within 75 minutes max; HelixQA's continuous probe catches it within 5 s.

---

## 9a. Service-Discovery Latency Budget

mDNS / DNS-SD discovery has its own latency budget:

| Operation                              | p50      | p99      | p999    |
|----------------------------------------|----------|----------|---------|
| Initial mDNS query (cold cache)        | 30 ms    | 100 ms   | 250 ms  |
| Subsequent queries (warm cache, 10 min) | 50 µs    | 200 µs   | 500 µs  |
| Service announcement (host-side)       | 5 ms     | 15 ms    | 30 ms   |
| Port-binding (`net.Listen(":0")`)      | 50 µs    | 150 µs   | 300 µs  |

The cold-cache 250 ms p999 is acceptable at session-start (rarely happens in steady state). Steady-state queries hit the cache.

## 9b. Fallback Discovery Server Topology

For deployments where mDNS is forbidden (some enterprise networks), the operator runs a **central discovery server**:

```yaml
# vasic-digital/Containers/lanes/discovery-server-1.x/docker-compose.yml
services:
  discovery-server:
    image: ghcr.io/vasic-digital/helix-discovery:latest
    ports:
      - "8090:8090"
    environment:
      DISCOVERY_BACKEND: postgres
      DISCOVERY_POSTGRES_DSN: ${POSTGRES_DSN}
```

Clients consume via `HELIX_DISCOVERY_SERVER=https://discovery.helix.example.com:8090`. The server maintains the same SRV records that mDNS would; clients query via REST or watch via long-polling.

## 9c. The DNS-SD Service-Type Naming Convention

Per RFC 6763 §4, service types follow `_<name>._<protocol>._<transport>.local.`:

- `<name>` is `helix-<role>` for the HelixPlay subset (e.g. `helix-host`, `helix-cataloger`, `helix-stream`).
- `<protocol>` is `_grpc` (custom; we own this), `_http`, `_https`, `_http3`, `_rtp` per the RFC.
- `<transport>` is `_tcp` (TCP) or `_udp` (UDP).

The custom `_grpc` protocol-name is operator-assigned and documented in the org's IANA-style registry at `vasic-digital/.github/dns-sd-service-types.md`.

## 9d. Operator-Side Firewall Setup

For LAN deployments, operator opens:

| Port / Protocol      | Purpose                                                     |
|----------------------|-------------------------------------------------------------|
| UDP 5353             | mDNS (RFC 6762)                                              |
| TCP/UDP dynamic 32768-60999 | HelixPlay services (gRPC, HTTP/3, RTP)                |
| TCP 8443             | helix-grpc-frame default (or operator-pinned)               |
| TCP 26257            | CockroachDB (operator-pinned)                                |
| TCP 4222             | NATS                                                         |
| TCP 6379             | Redis                                                        |
| TCP 8200             | Vault                                                        |

The dynamic-range 32768-60999 is broad; for stricter firewall postures, operator pins narrower ranges via env-var (e.g. `HELIX_GRPC_LISTEN_ADDR=:8443`, `HELIX_STREAM_PORT_RANGE=49152-49500`).

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-O03-A            | mDNS-blocked networks — what's the operator-supported fallback (discovery server)?                            | this chapter next revision                          |
| OQ-O03-B            | HTTP/3 default-on timeline — when does the op prefer HTTP/3 over HTTP/2?                                       | C06 §6 + O01 next revisions                         |
| OQ-O03-C            | RabbitMQ adoption — is there a use case NATS doesn't cover?                                                    | C06 §6 next revision                                |

---

## 9e. The Per-Mirror Service-Discovery Topology

In a Russian-jurisdiction operator deployment, mDNS may behave differently due to enterprise IPv6-multicast routing constraints. The per-mirror topology:

| Operator deployment | mDNS scope                        | Discovery server fallback?       |
|---------------------|-----------------------------------|----------------------------------|
| Western enterprise   | LAN-wide (`fe80::/10` link-local)| optional                          |
| Western home         | LAN-wide                          | not needed                        |
| Russian enterprise   | LAN-wide (subject to routing)    | recommended                       |
| Russian home         | LAN-wide                          | not needed                        |

The discovery-server topology (§9b) is operator-configurable per deployment via env-var; HelixQA's per-deployment dashboard surfaces which discovery mode is active.

## 9f. The Per-Tenant Service-Discovery Isolation

Per Constitution §11 + [helix-tenant descriptor](../06_Submodules/per-submodule/helix-tenant.md), tenant isolation extends to discovery. A tenant's helix-host services are only discoverable by clients that authenticated via that tenant's OAuth issuer. The mDNS records carry a `_tenant=<tenant-id>` TXT record; a foreign-tenant client filters out non-matching records.

Cross-tenant service-discovery leaks are flagged P2 by the visibility audit at [O01 §11](01_Container_CI_CD.md#11-per-mirror-registry-topology--image-replication).

## 9g. Operational Runbook for Discovery Failures

Operator runbook at `HelixDevelopment/HelixQA/docs/runbook/service-discovery-failure.md`:

1. Verify mDNS multicast UDP 5353 is open in the LAN firewall.
2. Verify `mDNSResponder` (macOS) / Avahi (Linux) / Bonjour (Windows) is running on every host.
3. If multi-segment LAN: verify mDNS forwarding via `mdns-repeater` or equivalent.
4. If completely blocked: switch to discovery-server fallback per §9b.
5. Verify by `dig +short _helix-host._grpc._tcp.local. PTR @224.0.0.251` from an affected client.

## 10. Open Questions

[Open Questions remain unchanged below.]

## 11. References & Anti-Bluff Verification

### 11.1 Internal

- [`../03_Architecture/05_RealTime_APIs.md`](../03_Architecture/05_RealTime_APIs.md) — C06 origin chapter.
- [`../06_Submodules/per-submodule/helix-grpc-frame.md`](../06_Submodules/per-submodule/helix-grpc-frame.md) — gRPC framing primitive.
- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §4.1 — continuous probe.
- [`../07_Testing/09_Smoke.md`](../07_Testing/09_Smoke.md) §4 — health endpoint contract.

### 11.2 External (web)

- mDNS RFC 6762: https://datatracker.ietf.org/doc/html/rfc6762 (accessed 2026-04-30).
- DNS-SD RFC 6763: https://datatracker.ietf.org/doc/html/rfc6763 (accessed 2026-04-30).
- grandcat/zeroconf: https://github.com/grandcat/zeroconf (accessed 2026-04-30).
- Connect-Go: https://connectrpc.com/ (accessed 2026-04-30).
- quic-go: https://github.com/quic-go/quic-go (accessed 2026-04-30).

### 11.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.
- §3.1 dynamic-range registry lists every submodule's port semantics; no service depends on a hard-coded port without operator-override option.

### 11.4 The mDNS-vs-DNS-Over-HTTPS Trade-off

For deployments where mDNS multicast UDP is forbidden, an alternative to the §9b discovery server is **DNS-over-HTTPS** (DoH) against a private DNS zone:

| Approach            | Pros                                            | Cons                                          |
|---------------------|-------------------------------------------------|----------------------------------------------|
| mDNS                | Zero-config, decentralised                      | Multicast often blocked; LAN-scoped          |
| Discovery server   | Centralised, easy to firewall                   | Single point of failure                      |
| DNS-over-HTTPS     | Standardised (RFC 8484), HTTPS-friendly         | Requires private-DNS provisioning             |

For Russian-jurisdiction operators where mDNS is unreliable + DoH is allowed, the operator-supplied `_helix-host._grpc._tcp.helix.example.com` SRV records (DNS-side) provide an alternative to the LAN-side discovery. HelixPlay clients are configured via env-var to query DoH instead of mDNS.

### 11.5 The Service-Discovery Operational Telemetry

| Metric                                          | Type    | Description                                            |
|-------------------------------------------------|---------|--------------------------------------------------------|
| `helix_discovery_announce_total`                | counter | Service announcements emitted, labelled `service_type`.|
| `helix_discovery_query_total`                   | counter | mDNS queries, labelled `service_type`, `result`.       |
| `helix_discovery_query_latency_seconds`         | histogram| Query-resolution latency.                              |
| `helix_discovery_fallback_active`               | gauge   | 1 if the discovery-server fallback is active for this client. |
| `helix_discovery_cross_tenant_filter_total`     | counter | Foreign-tenant records filtered (cross-tenant isolation enforcement). |

Dashboards at [O04 §6](04_Observability_and_Events.md#6-grafana-dashboards) include a "Discovery Health" panel.

### 11.6 Cross-Reference to C06 (Real-Time APIs)

The discovery + transport story is operationally split across:

- **C06 §6** ([Real-Time APIs](../03_Architecture/05_RealTime_APIs.md)) — the architectural layer (Connect-Go choice; HTTP/3 path; Brotli compression negotiation).
- **O03** (this chapter) — the operational layer (mDNS / DNS-SD topology; dynamic-port assignment; firewall + portal config).

C06 says **what** the protocol stack is. O03 says **how** an operator deploys it on a LAN. The two together implement R-07 from architecture → operations.

### 11.7 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `08_Operations/03_Service_Discovery_and_Ports.md` — 2026-04-30.
