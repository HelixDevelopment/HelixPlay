# HelixPlay API Reference

> **Version**: v0.1.0 (MVP)  
> **Generated**: 2026-05-02  
> **Source**: `pkg/protocol/v1/` (protobuf + OpenAPI)

---

## Overview

HelixPlay exposes a unified API layer consisting of:

1. **gRPC services** for internal microservice communication
2. **REST HTTP endpoints** for client-facing operations
3. **WebSocket streams** for real-time game telemetry

All APIs are versioned under `/api/v1/`.

---

## gRPC Services

### Catalog Service

`pkg/protocol/v1/catalog/catalog.proto`

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `ListGames` | `ListGamesRequest` | `ListGamesResponse` | List games available to a tenant |
| `GetGame` | `GetGameRequest` | `Game` | Get detailed game metadata |
| `SearchGames` | `SearchGamesRequest` | `SearchGamesResponse` | Full-text search across catalog |

### Discovery Service

`pkg/protocol/v1/discovery/discovery.proto`

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `AdvertiseCapability` | `Capability` | `Ack` | Host advertises streaming capability |
| `QueryEndpoints` | `QueryRequest` | `EndpointList` | Client queries available hosts |

### Streaming Service

`pkg/protocol/v1/streaming/streaming.proto`

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `StartSession` | `SessionRequest` | `Session` | Initiate a game streaming session |
| `Heartbeat` | `Ping` | `Pong` | Keep-alive and latency probe |
| `TerminateSession` | `TerminateRequest` | `Ack` | End streaming session |

### Host Agent Service

`pkg/protocol/v1/hostagent/hostagent.proto`

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `EnumerateGames` | `EnumRequest` | `GameList` | List installed games on host |
| `LaunchGame` | `LaunchRequest` | `LaunchStatus` | Start a game process |
| `GetTelemetry` | `TelemetryRequest` | `TelemetryStream` | Real-time host metrics |

---

## REST Endpoints

### Health & Readiness

```
GET /health
GET /ready
```

Returns service health status JSON.

### Tenant Management

```
GET    /api/v1/tenants
POST   /api/v1/tenants
GET    /api/v1/tenants/{id}
PUT    /api/v1/tenants/{id}
DELETE /api/v1/tenants/{id}
```

### Host Management

```
GET    /api/v1/hosts
POST   /api/v1/hosts
GET    /api/v1/hosts/{id}
PUT    /api/v1/hosts/{id}
DELETE /api/v1/hosts/{id}
```

### Session Management

```
GET    /api/v1/sessions
POST   /api/v1/sessions
GET    /api/v1/sessions/{id}
DELETE /api/v1/sessions/{id}
```

### Game Catalog

```
GET /api/v1/games
GET /api/v1/games/{id}
```

### User Management

```
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{id}
PUT    /api/v1/users/{id}
DELETE /api/v1/users/{id}
```

---

## WebSocket Streams

### Real-time Telemetry

```
WS /ws/v1/telemetry/{session_id}
```

Bi-directional WebSocket for:
- Frame delivery metadata
- Input event forwarding
- Latency probes (RTT measurement)

### Controller Input

```
WS /ws/v1/input/{session_id}
```

Low-latency input stream for controller state.

---

## Authentication

All API requests require one of:
- **JWT Bearer token** in `Authorization` header
- **API key** in `X-API-Key` header
- **OAuth2 access token** (for third-party integrations)

### Token Scopes

| Scope | Access |
|-------|--------|
| `read:games` | List and search games |
| `write:sessions` | Create and manage sessions |
| `admin:hosts` | Full host management |
| `admin:tenants` | Tenant provisioning |

---

## Error Codes

| HTTP | gRPC | Meaning |
|------|------|---------|
| 200 | OK | Success |
| 400 | INVALID_ARGUMENT | Bad request parameters |
| 401 | UNAUTHENTICATED | Missing or invalid credentials |
| 403 | PERMISSION_DENIED | Insufficient scope |
| 404 | NOT_FOUND | Resource does not exist |
| 409 | ALREADY_EXISTS | Resource conflict |
| 429 | RESOURCE_EXHAUSTED | Rate limit exceeded |
| 500 | INTERNAL | Server error |
| 503 | UNAVAILABLE | Service temporarily unavailable |

---

## Generating from Protobuf

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go bindings
protoc --go_out=. --go-grpc_out=. pkg/protocol/v1/**/*.proto
```

---

## Anti-Bluff Verification

### Source Evidence
- `pkg/protocol/v1/*/ *.proto` — 4 service definitions, reviewed 2026-05-02.
- `cmd/core/api/gateway.go` — REST routing implementation.
- `docs/api/README.md` — this document.

### Coverage Confirmation
- All 4 gRPC services have `.proto` definitions.
- All REST endpoints listed above have corresponding handler stubs.
- Authentication middleware enforces scopes per `cmd/core/auth/middleware.go`.
