# HelixPlay Architecture

## C4 Model Diagrams

### Level 1: System Context

```mermaid
C4Context
    title System Context — HelixPlay Cloud Gaming

    Person(player, "Player", "Wants to stream games from their PC to any device")
    Person(partner, "Partner", "ISP/Hotel/Hospital reselling white-label gaming")

    System(helixplay, "HelixPlay", "Cloud gaming platform")
    System_Ext(auth0, "Auth0", "OAuth2/OIDC identity provider")
    System_Ext(cdn, "CloudFront + S3", "4K asset delivery")
    System_Ext(catalogizer, "Catalogizer", "Game catalog & metadata API")

    Rel(player, helixplay, "Streams games via", "WebRTC/QUIC/UDP")
    Rel(partner, helixplay, "Configures white-label tenant via", "HTTPS/gRPC")
    Rel(helixplay, auth0, "Authenticates users via", "OAuth2/OIDC")
    Rel(helixplay, cdn, "Serves 4K assets via", "HTTPS")
    Rel(helixplay, catalogizer, "Fetches catalog via", "gRPC/REST")
```

### Level 2: Container Diagram

```mermaid
C4Container
    title Container Diagram — HelixPlay

    Person(player, "Player")

    Container(wails, "Wails Desktop", "Go + Svelte", "Desktop client")
    Container(flutter, "Flutter Mobile/TV", "Dart + Go FFI", "Mobile/TV client")
    Container(angular, "Angular Web", "TypeScript + Go WASM", "Browser client")

    Container(core, "Core Backend", "Go 1.26.2", "Session, tenant, catalog proxy")
    Container(host_agent, "Host Agent", "Go 1.26.2 + CGO", "Capture, encode, transport")
    Container(db, "CockroachDB", "CockroachDB 24.2", "Primary database")
    Container(redis, "Redis", "Redis 7", "Cache & sessions")
    Container(nats, "NATS JetStream", "NATS 2.10", "Event bus")

    Rel(player, wails, "Uses")
    Rel(player, flutter, "Uses")
    Rel(player, angular, "Uses")
    Rel(wails, core, "gRPC/REST", "mTLS + JWT")
    Rel(flutter, core, "gRPC/REST", "mTLS + JWT")
    Rel(angular, core, "REST/WebSocket", "JWT")
    Rel(core, db, "SQL", "TLS")
    Rel(core, redis, "Redis Protocol")
    Rel(core, nats, "NATS Protocol")
    Rel(wails, host_agent, "WebRTC/QUIC", "mTLS")
    Rel(flutter, host_agent, "WebRTC/QUIC", "mTLS")
```

### Level 3: Component Diagram (Host Agent)

```mermaid
C4Component
    title Component Diagram — Host Agent

    Container_Boundary(host, "Host Agent") {
        Component(capture, "Capture Service", "Go + OS APIs", "DXGI/SCK/PipeWire capture")
        Component(encoder, "Encoder Service", "Go + CGO", "NVENC/QSV/AMF/VT/VAAPI")
        Component(transport, "Transport Service", "Go", "WebRTC/QUIC/UDP")
        Component(input, "Input Pipeline", "Go + libusb", "1kHz USB polling, DualSense")
        Component(session, "Session Manager", "Go", "Lifecycle, negotiation, monitoring")
        Component(discovery, "Discovery Beacon", "Go", "mDNS + rendezvous")
    }

    Rel(capture, encoder, "Raw frames", "Zero-copy texture")
    Rel(encoder, transport, "Encoded bitstream", "RTP/QUIC datagram")
    Rel(input, transport, "Input events", "DataChannel/QUIC stream")
    Rel(session, capture, "Controls")
    Rel(session, encoder, "Configures")
    Rel(session, transport, "Manages")
    Rel(discovery, session, "Registers")
```

## Data Flow

```
Controller Input → USB HID → SPSC Ring Buffer → gRPC/QUIC → Host Agent
                                                          ↓
Game Frame → Capture (DXGI/SCK/PipeWire) → Encoder (NVENC/QSV/AMF) →
                                                          ↓
Transport (WebRTC SRTP / QUIC datagram / UDP DTLS) → Network →
                                                          ↓
Client Decode (WebCodecs/native) → Display (OpenGL/Metal/DirectX)
```

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go 1.26.2 | Green Tea GC, reduced CGO overhead, SIMD |
| Desktop | Wails v2 | Stable, battle-tested, native OS integration |
| Mobile/TV | Flutter 3.29+ | Impeller, Go FFI, Android TV support |
| Web | Angular 17+ | Signals, standalone components, Go WASM |
| Database | CockroachDB 24.2 | Serializable default, geo-partitioning |
| Streaming | WebRTC Pion v4 | Pure Go, cross-compilable, ICE/STUN/TURN |
| Events | NATS JetStream | At-least-once, geo-replication, consumer groups |

## Performance Budgets

| Stage | LAN | WAN | Measurement |
|-------|-----|-----|-------------|
| Controller input | 2ms | 15ms | `hid-bpf` trace |
| Network transit | 5ms | 25ms | QUIC RTT sample |
| Capture | 3ms | 3ms | GPU timestamp |
| Encode | 5ms | 5ms | Encoder API timestamp |
| Decode | 8ms | 8ms | WebCodecs callback |
| Display | 7ms | 7ms | `presentmon` |
| **Total p999** | **≤30ms** | **≤50ms** | HDR Histogram |

## Security Model

- mTLS between all services
- JWT (RS256) with short expiry + refresh rotation
- RBAC: player → host_admin → tenant_admin → super_admin
- Row-level security in CockroachDB per tenant
- Container isolation (rootless Podman in prod)
- Anti-bluff CI lane (non-overridable)

---

*Architecture v1.0.0 — 2026-05-02*
