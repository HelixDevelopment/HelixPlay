# Contract: Host Agent API

**Feature Branch**: `001-helixplay-system`  
**Created**: 2026-05-02  
**Status**: Draft — Phase 1 Design Artifact  
**Source**: `specs/001-helixplay-system/spec.md` §User Stories 1, 4, 7, FR-005, FR-010..FR-014, C-005

---

## 1. Protocol

- **Primary**: gRPC (`hostagent.v1.HostAgentService`) over mTLS
- **Fallback**: REST JSON over HTTPS (for simple admin/dashboard operations)
- **Streaming**: gRPC streaming for session events, telemetry, and logs
- **Event bus**: NATS for cross-host fleet events; Redis for host-local pub/sub

---

## 2. gRPC Service Definition

```protobuf
syntax = "proto3";
package hostagent.v1;

option go_package = "github.com/HelixDevelopment/HelixPlay/pkg/hostagent/v1;hostagentv1";

service HostAgentService {
  // Lifecycle
  rpc RegisterHost(RegisterHostRequest) returns (Host);
  rpc UpdateHostStatus(UpdateHostStatusRequest) returns (Host);
  rpc DeregisterHost(DeregisterHostRequest) returns (Empty);

  // Game management
  rpc ListInstalledGames(ListInstalledGamesRequest) returns (ListInstalledGamesResponse);
  rpc LaunchGame(LaunchGameRequest) returns (LaunchGameResponse);
  rpc TerminateGame(TerminateGameRequest) returns (TerminateGameResponse);
  rpc GetGameStatus(GetGameStatusRequest) returns (GameStatus);

  // Session management
  rpc CreateSession(CreateSessionRequest) returns (Session);
  rpc GetSession(GetSessionRequest) returns (Session);
  rpc ListSessions(ListSessionsRequest) returns (ListSessionsResponse);
  rpc TerminateSession(TerminateSessionRequest) returns (Session);

  // Capability
  rpc AdvertiseCapabilities(AdvertiseCapabilitiesRequest) returns (CapabilityAck);
  rpc GetCapabilities(GetCapabilitiesRequest) returns (CapabilitiesSnapshot);

  // Recording
  rpc StartRecording(StartRecordingRequest) returns (Recording);
  rpc StopRecording(StopRecordingRequest) returns (Recording);
  rpc ListRecordings(ListRecordingsRequest) returns (ListRecordingsResponse);

  // Telemetry / monitoring
  rpc StreamHostTelemetry(StreamHostTelemetryRequest) returns (stream TelemetryFrame);
  rpc GetHostMetrics(GetHostMetricsRequest) returns (HostMetrics);

  // Health
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## 3. Endpoints & Methods

### 3.1 `RegisterHost`

Initial registration of a new gaming PC with the coordinator.

**gRPC**: `hostagent.v1.HostAgentService/RegisterHost`  
**REST**: `POST /v1/hosts`

#### Auth

- mTLS client certificate identifies host hardware
- Bootstrap token (short-lived, single-use) from container bootstrapper

#### Request (`RegisterHostRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `fingerprint` | string (64 hex) | YES | SHA-256 of hardware IDs |
| `hostname` | string | YES | `^[a-zA-Z0-9][a-zA-Z0-9_-]{1,126}[a-zA-Z0-9]$` |
| `os` | string | YES | `linux`, `windows`, `darwin` |
| `os_version` | string | YES | SemVer |
| `agent_version` | string | YES | SemVer |
| `tenant_id` | string (UUID) | NO | NULL = personal host |
| `bootstrap_token` | string | YES | JWT, 15-minute expiry |
| `region` | string | NO | IATA region code |

#### Response (`Host`)

Returns full `Host` entity (see `data-model.md` §2.1) with assigned `host_id`.

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `FINGERPRINT_CONFLICT` | `ALREADY_EXISTS` | 409 | Host with this fingerprint already registered |
| `INVALID_BOOTSTRAP_TOKEN` | `UNAUTHENTICATED` | 401 | Token expired or signature invalid |
| `INVALID_HOSTNAME` | `INVALID_ARGUMENT` | 400 | Hostname pattern violation |
| `TENANT_NOT_FOUND` | `NOT_FOUND` | 404 | `tenant_id` does not exist |

---

### 3.2 `UpdateHostStatus`

Heartbeat and status update from host agent.

**gRPC**: `hostagent.v1.HostAgentService/UpdateHostStatus`  
**REST**: `PATCH /v1/hosts/{host_id}/status`

#### Auth

- mTLS client certificate
- JWT host token (issued at registration, rotated every 24 h)

#### Request (`UpdateHostStatusRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `host_id` | string (UUID) | YES | Valid UUID v7 |
| `status` | string | YES | `online`, `busy`, `maintenance`, `thermal_throttled` |
| `gpu_states` | repeated GPUState | NO | Per-GPU thermal/load state |
| `active_session_count` | int32 | NO | ≥ 0 |
| `timestamp_ns` | int64 | YES | Monotonic clock |

#### Response (`Host`)

Updated `Host` entity.

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `HOST_NOT_FOUND` | `NOT_FOUND` | 404 | `host_id` unknown |
| `STATUS_TRANSITION_INVALID` | `FAILED_PRECONDITION` | 400 | Illegal state transition |

---

### 3.3 `ListInstalledGames`

Enumerate games discovered by Host Agent on this host.

**gRPC**: `hostagent.v1.HostAgentService/ListInstalledGames`  
**REST**: `GET /v1/hosts/{host_id}/games`

#### Auth

- JWT Bearer (user token)
- RBAC: user must have `catalog:read` scope; host must be in user's tenant or personal

#### Request (`ListInstalledGamesRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `host_id` | string (UUID) | YES | Valid UUID v7 |
| `store` | string | NO | Filter by store |
| `include_metadata` | bool | NO | Fetch Catalogizer metadata |

#### Response (`ListInstalledGamesResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `games` | repeated Game | Installed games on this host |
| `last_scan_at` | string | ISO-8601 timestamp of last Host Agent scan |
| `scan_in_progress` | bool | TRUE if enumeration currently running |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `HOST_NOT_FOUND` | `NOT_FOUND` | 404 | — |
| `SCAN_IN_PROGRESS` | `UNAVAILABLE` | 503 | Results stale; retry in 30 s |

---

### 3.4 `LaunchGame`

Start a game process on the host.

**gRPC**: `hostagent.v1.HostAgentService/LaunchGame`  
**REST**: `POST /v1/hosts/{host_id}/games/{game_id}/launch`

#### Auth

- JWT Bearer (user token)
- RBAC: `game:launch` scope; user must own game or have subscription

#### Request (`LaunchGameRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `host_id` | string (UUID) | YES | Valid UUID v7 |
| `game_id` | string (UUID) | YES | Valid UUID v5 |
| `launch_args` | string | NO | Overrides default args; validated against allowlist |
| `env_vars` | map<string, string> | NO | Max 32 entries, keys validated |
| `priority` | string | NO | `normal`, `high`, `realtime` (OS scheduling hint) |

#### Response (`LaunchGameResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `process_id` | int64 | OS process ID |
| `status` | string | `launching`, `running`, `failed` |
| `session_ready` | bool | TRUE if streaming session can be created |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `GAME_NOT_INSTALLED` | `NOT_FOUND` | 404 | Game not found on host |
| `HOST_BUSY` | `RESOURCE_EXHAUSTED` | 503 | GPU already in session (C-005) |
| `LAUNCH_FAILED` | `INTERNAL` | 500 | Process spawn error |
| `ANTICHEAT_CONFLICT` | `FAILED_PRECONDITION` | 400 | Anti-cheat blocks injection |

---

### 3.5 `CreateSession`

Create a streaming session after game launch or for desktop streaming.

**gRPC**: `hostagent.v1.HostAgentService/CreateSession`  
**REST**: `POST /v1/hosts/{host_id}/sessions`

#### Auth

- JWT Bearer (user token)
- mTLS (client cert for device binding)

#### Request (`CreateSessionRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `host_id` | string (UUID) | YES | Valid UUID v7 |
| `client_id` | string (UUID) | YES | Valid UUID v7 |
| `user_id` | string (UUID) | YES | Valid UUID v5 |
| `game_id` | string (UUID) | NO | NULL = desktop stream |
| `preferred_codec` | string | NO | `h264`, `hevc`, `av1` |
| `preferred_transport` | string | NO | `webrtc`, `quic`, `udp_dtls` |
| `resolution` | Resolution | NO | `{w, h}` |
| `refresh_rate` | int32 | NO | Hz |
| `recording_enabled` | bool | NO | FALSE |

#### Response (`Session`)

Returns `Session` entity (see `data-model.md` §2.4) with `status = negotiating`.

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `HOST_NOT_FOUND` | `NOT_FOUND` | 404 | — |
| `HOST_BUSY` | `RESOURCE_EXHAUSTED` | 503 | GPU in use |
| `GAME_NOT_RUNNING` | `FAILED_PRECONDITION` | 400 | Game requested but not launched |
| `CLIENT_UNAUTHORIZED` | `PERMISSION_DENIED` | 403 | Client does not belong to user |
| `TENANT_QUOTA_EXCEEDED` | `RESOURCE_EXHAUSTED` | 429 | Tenant concurrent session limit |

---

### 3.6 `TerminateSession`

Gracefully end a streaming session.

**gRPC**: `hostagent.v1.HostAgentService/TerminateSession`  
**REST**: `DELETE /v1/hosts/{host_id}/sessions/{session_id}`

#### Auth

- JWT Bearer (user or admin token)
- Users may terminate their own sessions; `host_admin` may terminate any on their host

#### Request (`TerminateSessionRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `host_id` | string (UUID) | YES | Valid UUID v7 |
| `session_id` | string (UUID) | YES | Valid UUID v7 |
| `reason` | string | NO | `user_request`, `host_shutdown`, `timeout`, `error` |
| `force` | bool | NO | TRUE = SIGKILL game process |

#### Response (`Session`)

Updated `Session` with `status = terminating` or `ended`.

---

### 3.7 `AdvertiseCapabilities`

Host pushes current GPU/codec/transport capabilities.

**gRPC**: `hostagent.v1.HostAgentService/AdvertiseCapabilities`  
**REST**: `PUT /v1/hosts/{host_id}/capabilities`

#### Request (`AdvertiseCapabilitiesRequest`)

| Field | Type | Required |
|-------|------|----------|
| `host_id` | string (UUID) | YES |
| `capabilities` | CapabilitiesSnapshot | YES |

#### Response (`CapabilityAck`)

| Field | Type | Description |
|-------|------|-------------|
| `capability_id` | string (UUID) | Assigned capability snapshot ID |
| `expires_at` | string | ISO-8601 timestamp (now + TTL) |

---

### 3.8 `StartRecording` / `StopRecording`

Control dual-path recording.

**gRPC**: `hostagent.v1.HostAgentService/StartRecording` / `StopRecording`  
**REST**: `POST/DELETE /v1/hosts/{host_id}/sessions/{session_id}/recording`

#### Auth

- JWT Bearer with `recording:write` scope
- User quota check: `storage_used_gb + estimated_size ≤ storage_quota_gb`

#### Request (`StartRecordingRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `host_id` | string (UUID) | YES | — |
| `session_id` | string (UUID) | YES | — |
| `quality_preset` | string | NO | `ultra`, `high`, `medium`, `low`; default `high` |
| `container_format` | string | NO | `mkv`, `fmp4`; default `mkv` |

#### Response (`Recording`)

Returns `Recording` entity (see `data-model.md` §2.8) with `status = capturing`.

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `SESSION_NOT_STREAMING` | `FAILED_PRECONDITION` | 400 | Session not in `active` status |
| `STORAGE_QUOTA_EXCEEDED` | `RESOURCE_EXHAUSTED` | 429 | User storage full |
| `RECORDING_ALREADY_ACTIVE` | `ALREADY_EXISTS` | 409 | Recording exists for this session |
| `ENCODER_BUSY` | `RESOURCE_EXHAUSTED` | 503 | Record path encoder unavailable |

---

### 3.9 `StreamHostTelemetry`

Server-streaming telemetry from host to coordinator or authorized client.

**gRPC**: `hostagent.v1.HostAgentService/StreamHostTelemetry`  
**REST**: Server-Sent Events `GET /v1/hosts/{host_id}/telemetry`

#### Request (`StreamHostTelemetryRequest`)

| Field | Type | Required |
|-------|------|----------|
| `host_id` | string (UUID) | YES |
| `session_id` | string (UUID) | NO | Filter to specific session |
| `metric_types` | repeated string | NO | `gpu`, `thermal`, `network`, `encoder`, `input` |
| `sample_interval_ms` | int32 | NO | Default 1000, min 100 |

#### Response Stream (`TelemetryFrame`)

| Field | Type | Description |
|-------|------|-------------|
| `timestamp_ns` | int64 | — |
| `gpu_metrics` | repeated GPUMetric | `{index, utilization_pct, temp_c, vram_used_mb, vram_total_mb}` |
| `thermal_metrics` | ThermalMetric | `{cpu_temp_c, chassis_temp_c, fan_rpm[]}` |
| `encoder_metrics` | EncoderMetric | `{codec, bitrate_kbps, fps, dropped_frames}` |
| `network_metrics` | NetworkMetric | `{bytes_sent, bytes_received, retransmits}` |

---

### 3.10 `GetHostMetrics`

Point-in-time host metrics.

**gRPC**: `hostagent.v1.HostAgentService/GetHostMetrics`  
**REST**: `GET /v1/hosts/{host_id}/metrics`

---

## 4. Rate Limits

| Endpoint | Rate Limit | Burst | Scope |
|----------|------------|-------|-------|
| `RegisterHost` | 5/min | 2 | Per IP |
| `UpdateHostStatus` (heartbeat) | 4/min | 1 | Per `host_id` |
| `ListInstalledGames` | 60/min | 10 | Per `host_id` + user |
| `LaunchGame` | 10/min | 3 | Per `host_id` + user |
| `CreateSession` | 10/min | 3 | Per `host_id` + user |
| `TerminateSession` | 20/min | 5 | Per `host_id` + user |
| `StartRecording` | 5/min | 2 | Per `host_id` + user |
| `StreamHostTelemetry` | 1 connection | — | Per `host_id` |

---

## 5. Auth Requirements Summary

| Endpoint | Auth | Scopes | RBAC |
|----------|------|--------|------|
| `RegisterHost` | mTLS + bootstrap JWT | `host:register` | One-time token from bootstrapper |
| `UpdateHostStatus` | mTLS + host JWT | `host:heartbeat` | Host self-only |
| `ListInstalledGames` | JWT Bearer | `catalog:read` | User in tenant or personal host owner |
| `LaunchGame` | JWT Bearer | `game:launch` | Game owner or subscriber |
| `CreateSession` | JWT + mTLS | `stream:negotiate` | User + device binding |
| `TerminateSession` | JWT Bearer | `stream:terminate` | Self or `host_admin` |
| `AdvertiseCapabilities` | mTLS + host JWT | `host:advertise` | Host self-only |
| `StartRecording` | JWT Bearer | `recording:write` | Session owner |
| `StreamHostTelemetry` | JWT Bearer | `host:read` | `host_admin` or `super_admin` |
| `Health` | None | — | Public |

---

## 6. Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `FINGERPRINT_CONFLICT` | `ALREADY_EXISTS` | 409 | Duplicate host registration |
| `INVALID_BOOTSTRAP_TOKEN` | `UNAUTHENTICATED` | 401 | Bootstrap token invalid |
| `HOST_NOT_FOUND` | `NOT_FOUND` | 404 | Host ID unknown |
| `HOST_BUSY` | `RESOURCE_EXHAUSTED` | 503 | GPU in active session (C-005) |
| `STATUS_TRANSITION_INVALID` | `FAILED_PRECONDITION` | 400 | Illegal state machine step |
| `GAME_NOT_INSTALLED` | `NOT_FOUND` | 404 | Game not on host |
| `GAME_NOT_RUNNING` | `FAILED_PRECONDITION` | 400 | Launch before session creation |
| `LAUNCH_FAILED` | `INTERNAL` | 500 | OS process error |
| `ANTICHEAT_CONFLICT` | `FAILED_PRECONDITION` | 400 | Anti-cheat blocks |
| `CLIENT_UNAUTHORIZED` | `PERMISSION_DENIED` | 403 | Device not owned by user |
| `TENANT_QUOTA_EXCEEDED` | `RESOURCE_EXHAUSTED` | 429 | Tenant limits |
| `SESSION_NOT_STREAMING` | `FAILED_PRECONDITION` | 400 | Session inactive |
| `STORAGE_QUOTA_EXCEEDED` | `RESOURCE_EXHAUSTED` | 429 | User storage full |
| `RECORDING_ALREADY_ACTIVE` | `ALREADY_EXISTS` | 409 | Duplicate recording |
| `ENCODER_BUSY` | `RESOURCE_EXHAUSTED` | 503 | Record path unavailable |

---

## 7. Anti-Bluff Verification

### Sources
| Source | Lines | Insights Used |
|--------|-------|---------------|
| `specs/001-helixplay-system/spec.md` | 651 | User Stories 1, 4, 7; FR-005, FR-010..FR-014; C-005 (1 session/GPU); edge cases (thermal, multi-GPU, anticheat) |
| `AGENTS.md` | 286 | Host Agent scope, container bootstrap, mTLS, NATS/Redis events |

### Conflict Resolution
- **1 session per GPU (C-005)**: Enforced at `CreateSession` and `LaunchGame` via `HOST_BUSY`.
- **Anti-cheat**: Documented as `ANTICHEAT_CONFLICT` error; Host Agent runs in user space and avoids kernel hooks that trigger EAC/BattlEye.
- **Dual-path recording**: `StartRecording` controls record path independently of stream path.

### Verification
- ✅ All 12 gRPC methods documented with request/response schemas.
- ✅ REST fallback paths provided.
- ✅ Auth requirements mapped to scopes and RBAC per method.
- ✅ Rate limits and error codes enumerated.
- ✅ State transitions referenced from `data-model.md`.
- ✅ No `TODO`/`FIXME`/`placeholder` present.
