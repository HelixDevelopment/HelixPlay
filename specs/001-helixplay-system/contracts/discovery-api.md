# Contract: Discovery API

**Feature Branch**: `001-helixplay-system`  
**Created**: 2026-05-02  
**Status**: Draft — Phase 1 Design Artifact  
**Decoupling**: Discovery primitives in `vasic-digital/Discovery`; this contract defines HelixPlay-specific usage.  
**Source**: `specs/001-helixplay-system/spec.md` §User Story 1, FR-022, FR-023, DEP-007

---

## 1. Protocol Overview

HelixPlay uses two discovery modes:

| Mode | Protocol | Scope | Use Case |
|------|----------|-------|----------|
| LAN | mDNS (RFC 6762) + DNS-SD (RFC 6763) | Same subnet | Home networks, local LAN parties |
| WAN | Rendezvous service (gRPC/HTTP) | Internet | Remote friends, multi-region fleet |

### 1.1 mDNS / DNS-SD (LAN)

- **Service type**: `_helixplay._tcp.local.` and `_helixplay._udp.local.`
- **Port**: Dynamic per host (R-07); advertised in SRV record
- **TXT record**: Host capability summary, host_id, tenant_id (if assigned)
- **Security**: mTLS certificate fingerprint in TXT record for initial trust

### 1.2 Rendezvous Service (WAN)

- **Protocol**: gRPC over HTTPS/2 (HTTP/3 QUIC preferred)
- **Auth**: JWT Bearer for client queries; mTLS for host registration
- **Storage**: Redis (hot) + CockroachDB (persistent) for host registry
- **TTL**: Host entries expire after 3× heartbeat interval (default 45 s)

---

## 2. mDNS / DNS-SD Contract

### 2.1 Service Advertisement (Host Agent)

```
Service Instance: HelixPlay-Host-<host_id>._helixplay._tcp.local.
  Host: <hostname>.local.
  Port: <dynamic_port>
  TXT:
    - host_id=<uuid>
    - tenant_id=<uuid> or "personal"
    - version=1.0.0
    - os=linux|windows|darwin
    - gpu_count=1
    - gpu_0=nvidia_rtx4090
    - codecs=h264,hevc,av1
    - transports=webrtc,quic,udp
    - max_res=3840x2160
    - max_hz=120
    - hdr=true
    - status=online|busy
    - cert_fingerprint=<SHA-256 of mTLS cert>
```

### 2.2 Service Discovery (Client)

1. Client browses `_helixplay._tcp.local.` via system mDNS resolver.
2. For each discovered instance:
   - Verify `cert_fingerprint` against pinned cert or CA trust store.
   - Check `status` != `busy` (C-005: 1 session per GPU).
   - Check `tenant_id` matches user's tenant (or `personal` for direct pairing).
3. Client displays discovered hosts in UI with GPU model and status.

### 2.3 Dynamic Port Assignment (R-07)

- Host Agent requests an ephemeral port from OS (`bind(0)`).
- Port is advertised in mDNS SRV record.
- If port collision detected, Host Agent re-binds and re-advertises.
- Port range: 40000–65535 (avoids well-known ports).

---

## 3. Rendezvous Service gRPC API

```protobuf
syntax = "proto3";
package discovery.v1;

option go_package = "github.com/HelixDevelopment/HelixPlay/pkg/discovery/v1;discoveryv1";

service RendezvousService {
  // Host registration / heartbeat
  rpc RegisterHost(RendezvousHost) returns (RegisterHostResponse);
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  rpc DeregisterHost(DeregisterHostRequest) returns (Empty);

  // Client queries
  rpc ListHosts(ListHostsRequest) returns (ListHostsResponse);
  rpc GetHost(GetHostRequest) returns (RendezvousHost);
  rpc WatchHosts(WatchHostsRequest) returns (stream HostEvent);

  // Health
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## 4. Endpoints & Methods

### 4.1 `RegisterHost`

Host registers with rendezvous service on startup.

**gRPC**: `discovery.v1.RendezvousService/RegisterHost`  
**REST**: `POST /v1/rendezvous/hosts`

#### Auth

- mTLS client certificate (host identity)
- Host JWT token (issued at `HostAgentService.RegisterHost`)

#### Request (`RendezvousHost`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `host_id` | string (UUID) | YES | Valid UUID v7 |
| `tenant_id` | string (UUID) | NO | NULL = personal |
| `public_endpoint` | string | YES | `host:port` or IPv4/IPv6 address |
| `local_endpoint` | string | YES | LAN IP for direct connections |
| `os` | string | YES | `linux`, `windows`, `darwin` |
| `version` | string | YES | SemVer |
| `capabilities` | CapabilitiesSnapshot | YES | Current GPU/codec/transport caps |
| `cert_fingerprint` | string (64 hex) | YES | SHA-256 of mTLS cert |

#### Response (`RegisterHostResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `lease_id` | string | Opaque lease token |
| `heartbeat_interval_sec` | int32 | Expected heartbeat frequency (default 15) |
| `expires_at` | string | ISO-8601; host entry deleted if no heartbeat |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `HOST_ALREADY_REGISTERED` | `ALREADY_EXISTS` | 409 | Host ID already has active lease |
| `INVALID_CERT` | `UNAUTHENTICATED` | 401 | mTLS cert does not match known fingerprint |
| `INVALID_CAPABILITIES` | `INVALID_ARGUMENT` | 400 | Capability schema validation failed |

---

### 4.2 `Heartbeat`

Host sends periodic heartbeat to refresh lease.

**gRPC**: `discovery.v1.RendezvousService/Heartbeat`  
**REST**: `PUT /v1/rendezvous/hosts/{host_id}/heartbeat`

#### Request (`HeartbeatRequest`)

| Field | Type | Required |
|-------|------|----------|
| `host_id` | string (UUID) | YES |
| `lease_id` | string | YES |
| `status` | string | YES | `online`, `busy`, `maintenance`, `thermal_throttled` |
| `active_session_count` | int32 | NO |
| `timestamp_ns` | int64 | YES |

#### Response (`HeartbeatResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `lease_extended` | bool | TRUE if heartbeat accepted |
| `expires_at` | string | New expiration timestamp |
| `action` | string | `none`, `re_register`, `update_caps` |

---

### 4.3 `DeregisterHost`

Graceful host removal.

**gRPC**: `discovery.v1.RendezvousService/DeregisterHost`  
**REST**: `DELETE /v1/rendezvous/hosts/{host_id}`

#### Request (`DeregisterHostRequest`)

| Field | Type | Required |
|-------|------|----------|
| `host_id` | string (UUID) | YES |
| `lease_id` | string | YES |
| `reason` | string | NO | `shutdown`, `maintenance`, `error` |

---

### 4.4 `ListHosts`

Client queries available hosts.

**gRPC**: `discovery.v1.RendezvousService/ListHosts`  
**REST**: `GET /v1/rendezvous/hosts`

#### Auth

- JWT Bearer (user token)
- RBAC: user sees hosts in their tenant + personal hosts they own

#### Request (`ListHostsRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tenant_id` | string (UUID) | NO | Filter by tenant; omit for personal hosts |
| `region` | string | NO | IATA region code |
| `status` | string | NO | `online`, `busy`, `all` |
| `capability_filter` | CapabilityFilter | NO | GPU model, codec, transport requirements |
| `page_size` | int32 | NO | 1–100, default 20 |
| `page_token` | string | NO | Opaque cursor |

#### Response (`ListHostsResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `hosts` | repeated RendezvousHost | Current host list |
| `next_page_token` | string | — |
| `total_size` | int64 | — |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `TENANT_ACCESS_DENIED` | `PERMISSION_DENIED` | 403 | User not in requested tenant |

---

### 4.5 `GetHost`

Get single host details.

**gRPC**: `discovery.v1.RendezvousService/GetHost`  
**REST**: `GET /v1/rendezvous/hosts/{host_id}`

#### Request (`GetHostRequest`)

| Field | Type | Required |
|-------|------|----------|
| `host_id` | string (UUID) | YES |

---

### 4.6 `WatchHosts`

Server-streaming real-time host list updates.

**gRPC**: `discovery.v1.RendezvousService/WatchHosts`  
**REST**: Server-Sent Events `GET /v1/rendezvous/hosts:watch`

#### Request (`WatchHostsRequest`)

| Field | Type | Required |
|-------|------|----------|
| `tenant_id` | string (UUID) | NO |
| `region` | string | NO |
| `capability_filter` | CapabilityFilter | NO |

#### Response Stream (`HostEvent`)

| Field | Type | Description |
|-------|------|-------------|
| `event_type` | string | `added`, `updated`, `removed` |
| `host` | RendezvousHost | Full host snapshot |
| `timestamp_ns` | int64 | — |

---

## 5. Rendezvous Host Schema

```protobuf
message RendezvousHost {
  string host_id = 1;
  string tenant_id = 2;
  string public_endpoint = 3;   // Internet-reachable address
  string local_endpoint = 4;    // LAN address (for direct connection)
  string os = 5;
  string version = 6;
  string status = 7;            // online, busy, maintenance, thermal_throttled
  CapabilitiesSnapshot capabilities = 8;
  string cert_fingerprint = 9;
  string region = 10;
  int32 active_session_count = 11;
  string last_seen_at = 12;     // ISO-8601
  string registered_at = 13;    // ISO-8601
}
```

---

## 6. Rate Limits

| Endpoint | Rate Limit | Burst | Scope |
|----------|------------|-------|-------|
| `RegisterHost` | 5/min | 2 | Per host fingerprint |
| `Heartbeat` | 4/min | 1 | Per `host_id` |
| `DeregisterHost` | 10/min | 3 | Per `host_id` |
| `ListHosts` | 60/min | 10 | Per user |
| `GetHost` | 120/min | 20 | Per user |
| `WatchHosts` | 1 connection | — | Per user |

---

## 7. Auth Requirements

| Endpoint | Auth | Scopes | RBAC |
|----------|------|--------|------|
| `RegisterHost` | mTLS + host JWT | `host:register` | Host self-only |
| `Heartbeat` | mTLS + host JWT | `host:heartbeat` | Host self-only |
| `DeregisterHost` | mTLS + host JWT | `host:register` | Host self-only |
| `ListHosts` | JWT Bearer | `catalog:read` | Tenant-scoped |
| `GetHost` | JWT Bearer | `catalog:read` | Tenant-scoped |
| `WatchHosts` | JWT Bearer | `catalog:read` | Tenant-scoped |
| `Health` | None | — | Public |

---

## 8. NAT Traversal

For WAN connections where host is behind NAT:

1. Host registers `public_endpoint` via STUN (RFC 8489) to discover public IP:port.
2. If symmetric NAT detected, host opens relay via TURN (RFC 8656) on rendezvous service.
3. Client receives both `public_endpoint` (P2P attempt) and `relay_endpoint` (fallback).
4. WebRTC ICE handles this automatically; QUIC uses rendezvous-coordinated hole punching.

---

## 9. Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `HOST_ALREADY_REGISTERED` | `ALREADY_EXISTS` | 409 | Duplicate registration |
| `INVALID_CERT` | `UNAUTHENTICATED` | 401 | mTLS mismatch |
| `INVALID_CAPABILITIES` | `INVALID_ARGUMENT` | 400 | Capability schema error |
| `LEASE_EXPIRED` | `UNAUTHENTICATED` | 401 | Heartbeat missed, lease lost |
| `TENANT_ACCESS_DENIED` | `PERMISSION_DENIED` | 403 | User cannot query tenant hosts |
| `HOST_NOT_FOUND` | `NOT_FOUND` | 404 | Host ID unknown in registry |

---

## 10. Anti-Bluff Verification

### Sources
| Source | Lines | Insights Used |
|--------|-------|---------------|
| `specs/001-helixplay-system/spec.md` | 651 | User Story 1, FR-022, FR-023, R-07 dynamic ports, DEP-007 |
| `AGENTS.md` | 286 | mDNS, rendezvous, NATS/Redis, gRPC preferred |

### Conflict Resolution
- **LAN vs WAN**: mDNS for same-subnet (no server dependency); rendezvous for cross-network. Both return identical `RendezvousHost` schema.
- **Dynamic ports (R-07)**: Host Agent binds ephemeral port and advertises via mDNS TXT + rendezvous `public_endpoint`. Port range 40000–65535 avoids conflicts.
- **NAT traversal**: STUN/TURN integrated into rendezvous; WebRTC ICE handles most cases automatically.

### Verification
- ✅ mDNS/DNS-SD TXT record schema fully specified.
- ✅ Rendezvous gRPC API with 6 methods documented.
- ✅ REST fallback paths provided.
- ✅ Rate limits, auth requirements, and RBAC enumerated.
- ✅ NAT traversal strategy (STUN/TURN/ICE) documented.
- ✅ No `TODO`/`FIXME`/`placeholder` present.
