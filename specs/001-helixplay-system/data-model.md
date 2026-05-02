# Data Model: HelixPlay System

**Date:** 2026-05-02  
**Branch:** `001-helixplay-system`  
**Source:** spec.md §Requirements + research.md

---

## Entity Overview

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  Tenant │────<│  User   │────<│ Session │────>│  Host   │
└────┬────┘     └────┬────┘     └────┬────┘     └────┬────┘
     │               │               │               │
     │         ┌─────┴─────┐         │               │
     │         │ Controller│         │               │
     │         └───────────┘         │               │
     │                               │               │
     └───────────────┐   ┌───────────┘               │
                     ▼   ▼                           │
               ┌─────────────┐                 ┌─────┴─────┐
               │ Catalog/Game│                 │ Capability│
               └─────────────┘                 └───────────┘
                     │                               │
                     ▼                               ▼
               ┌─────────────┐                 ┌─────────────┐
               │  Recording  │                 │   Asset     │
               └─────────────┘                 └─────────────┘
```

---

## 1. Tenant

A white-label partner (ISP, hotel, hospital, venue) that resells HelixPlay under their own brand.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique tenant identifier |
| `slug` | string | unique, lowercase, regex `^[a-z0-9-]+$` | URL-safe identifier |
| `name` | string | max 128 chars | Display name |
| `theme_primary_color` | string | hex `#RRGGBB` | Primary brand color |
| `theme_secondary_color` | string | hex `#RRGGBB` | Secondary brand color |
| `theme_logo_url` | string | URL, max 2048 chars | CDN URL to logo asset |
| `oauth2_provider` | string | enum: `auth0`, `custom` | Identity provider config |
| `oauth2_config` | JSONB | nullable | Provider-specific settings (client_id, domain, etc.) |
| `catalog_filter` | JSONB | nullable | Allowed game genres, ratings, platforms |
| `monetization_model` | string | enum: `subscription`, `usage`, `revenue_share` | Billing model |
| `resource_quota_max_sessions` | int | ≥1 | Max concurrent sessions |
| `resource_quota_storage_gb` | int | ≥0 | Max recording storage per user |
| `created_at` | timestamp | auto-gen | Creation time |
| `updated_at` | timestamp | auto-update | Last modification |

**Relationships:**
- One Tenant → Many Users (`tenant_id` FK on User)
- One Tenant → Many CatalogFilters (row-level security on Game visibility)

**State Machine:** N/A (Tenant is persistent; soft-delete only)

---

## 2. User

A player with OAuth2/OIDC identity, scoped to a Tenant.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique user identifier |
| `tenant_id` | UUID | FK → Tenant.id, cascade delete | Tenant scope |
| `email` | string | unique per tenant, max 256 chars | Primary email |
| `display_name` | string | max 64 chars | Public name |
| `roles` | string[] | enum: `player`, `admin`, `partner_admin` | RBAC roles |
| `oauth2_subject` | string | max 256 chars | External identity provider subject |
| `oauth2_provider` | string | max 64 chars | Provider name (e.g., `auth0`) |
| `preferred_controller_type` | string | enum: `dualsense`, `xbox`, `keyboard_mouse` | Default input device |
| `storage_quota_used_gb` | float | ≥0, default 0 | Current storage usage |
| `created_at` | timestamp | auto-gen | Creation time |
| `last_login_at` | timestamp | nullable | Last successful login |

**Relationships:**
- One User → Many Sessions
- One User → Many Recordings

**Validation:**
- `email` must match RFC 5322
- `roles` must contain at least `player`
- `oauth2_subject` + `oauth2_provider` composite unique per tenant

---

## 3. Host

A gaming PC running the HelixPlay Host Agent.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique host identifier |
| `name` | string | max 128 chars | Human-readable host name |
| `hardware_id` | string | unique, max 256 chars | Hardware fingerprint (hash of CPU+MB+GPU serials) |
| `os` | string | enum: `windows`, `macos`, `linux` | Host operating system |
| `os_version` | string | max 64 chars | OS version string |
| `last_seen_at` | timestamp | nullable | Last heartbeat timestamp |
| `status` | string | enum: `online`, `offline`, `streaming`, `maintenance` | Current status |
| `lan_ip` | string | nullable, IPv4/IPv6 | Local network address |
| `wan_endpoint` | string | nullable | Public rendezvous endpoint |
| `discovery_port` | int | 1024-65535 | mDNS/rendezvous port (dynamic per R-07) |
| `thermal_throttle` | bool | default false | Thermal throttling active |
| `created_at` | timestamp | auto-gen | Registration time |

**Relationships:**
- One Host → Many Sessions
- One Host → Many GPUs (1:N, typical 1-4 GPUs)
- One Host → Many Games (installed games)

**State Machine:**
```
offline → online (heartbeat received)
online → streaming (session started)
streaming → online (session ended)
online → offline (heartbeat timeout >30s)
online → maintenance (operator sets mode)
maintenance → online (operator clears mode)
```

---

## 4. GPU (Capability Advertisement)

Per-GPU capabilities advertised by the Host Agent.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique GPU identifier |
| `host_id` | UUID | FK → Host.id, cascade delete | Parent host |
| `vendor` | string | enum: `nvidia`, `intel`, `amd`, `apple` | GPU vendor |
| `model` | string | max 128 chars | GPU model name (e.g., `RTX 4090`) |
| `driver_version` | string | max 64 chars | Driver version |
| `vram_gb` | int | ≥0 | Video RAM in GB |
| `encoder` | string | enum: `nvenc`, `qsv`, `amf`, `videotoolbox`, `vaapi` | Hardware encoder API |
| `codecs_supported` | string[] | enum values: `h264`, `hevc`, `av1` | Supported encode codecs |
| `max_resolution` | string | enum: `1080p`, `1440p`, `4k`, `8k` | Max encode resolution |
| `max_refresh_hz` | int | ≥30 | Max refresh rate |
| `nvenc_sessions` | int | ≥1, nullable | Available NVENC sessions (NVIDIA only) |
| `thermal_c` | int | nullable | Current temperature in Celsius |
| `session_count` | int | default 0, ≥0 | Active sessions on this GPU |
| `updated_at` | timestamp | auto-update | Last capability refresh |

**Validation:**
- `nvenc_sessions` required only when `vendor = nvidia`
- `codecs_supported` must contain at least `h264`
- `session_count` must be ≤ `nvenc_sessions` (or 1 if null, per C-005)

---

## 5. Game

A game entry in the catalog, normalized across all store integrations.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Canonical HelixPlay game ID |
| `store_id` | string | max 256 chars | Store-specific identifier (Steam appid, etc.) |
| `store_type` | string | enum: `steam`, `epic`, `gog`, `ubisoft`, `battlenet`, `origin`, `microsoft`, `standalone` | Source store |
| `title` | string | max 256 chars | Game title |
| `description` | text | max 4096 chars | Short description |
| `cover_art_url` | string | URL, max 2048 chars | 4K cover art CDN URL |
| `screenshot_urls` | string[] | max 20 items | Screenshot CDN URLs |
| `video_urls` | string[] | max 5 items | Trailer CDN URLs |
| `genres` | string[] | max 10 items | Genre tags |
| `release_date` | date | nullable | Original release date |
| `rating_esrb` | string | enum: `E`, `E10`, `T`, `M`, `AO` | ESRB rating |
| `min_gpu` | string | nullable | Minimum GPU requirement |
| `rec_gpu` | string | nullable | Recommended GPU requirement |
| `hdr_support` | bool | default false | HDR10/HDR10+ support |
| `atmos_support` | bool | default false | Dolby Atmos support |
| `dualsense_support` | bool | default false | DualSense haptics/triggers |
| `multiplayer` | bool | default false | Multiplayer capable |
| `installed_path` | string | nullable, max 1024 chars | Local install directory (host-scoped) |
| `exe_path` | string | nullable, max 1024 chars | Executable path relative to installed_path |
| `updated_at` | timestamp | auto-update | Last metadata sync |

**Relationships:**
- Many Games → Many Hosts (via HostGame junction table)
- Many Games → Many Tenants (via CatalogFilter)

**Uniqueness:**
- Composite unique: (`store_type`, `store_id`) — same game from different stores is deduplicated
- Canonical `id` is generated from hash of (`store_type`, `store_id`)

---

## 6. Session

An active streaming session between a Client and a Host.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique session identifier |
| `user_id` | UUID | FK → User.id | Session owner |
| `host_id` | UUID | FK → Host.id | Host running the game |
| `game_id` | UUID | FK → Game.id | Game being streamed |
| `gpu_id` | UUID | FK → GPU.id | GPU handling encode |
| `status` | string | enum: `connecting`, `negotiating`, `streaming`, `paused`, `reconnecting`, `disconnected`, `ended` | Session state |
| `codec` | string | enum: `h264`, `hevc`, `av1` | Selected codec |
| `transport` | string | enum: `webrtc`, `quic`, `udp` | Selected transport |
| `resolution` | string | enum: `720p`, `1080p`, `1440p`, `4k` | Stream resolution |
| `refresh_hz` | int | enum: 30, 60, 120, 144 | Stream refresh rate |
| `latency_p50_ms` | float | nullable | p50 glass-to-glass latency |
| `latency_p99_ms` | float | nullable | p99 glass-to-glass latency |
| `latency_p999_ms` | float | nullable | p999 glass-to-glass latency |
| `bandwidth_mbps` | float | nullable | Estimated bandwidth |
| `packet_loss_pct` | float | 0-100, nullable | Packet loss percentage |
| `controller_type` | string | enum: `dualsense`, `xbox`, `keyboard_mouse` | Active controller |
| `recording_enabled` | bool | default false | DVR recording active |
| `recording_path` | string | nullable | Local recording file path |
| `started_at` | timestamp | auto-gen | Session start time |
| `ended_at` | timestamp | nullable | Session end time |
| `last_activity_at` | timestamp | auto-update | Last input/video frame |

**State Machine:**
```
connecting → negotiating (ICE/capability handshake complete)
negotiating → streaming (codec/transport selected, first frame sent)
streaming → paused (user pauses)
paused → streaming (user resumes)
streaming → reconnecting (network disruption detected)
reconnecting → streaming (reconnection successful, ≤30s)
reconnecting → disconnected (reconnection failed, >30s)
streaming → ended (user quits or host stops)
disconnected → ended (cleanup after grace period)
```

**Validation:**
- `ended_at` ≥ `started_at`
- `latency_p999_ms` ≥ `latency_p99_ms` ≥ `latency_p50_ms`
- `status = ended` requires `ended_at` not null
- `recording_path` required only when `recording_enabled = true`

---

## 7. Recording

A DVR capture of a streaming session.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique recording identifier |
| `session_id` | UUID | FK → Session.id | Source session |
| `user_id` | UUID | FK → User.id | Recording owner |
| `format` | string | enum: `mkv`, `fmp4` | Container format |
| `codec` | string | enum: `h264`, `hevc`, `av1` | Video codec |
| `quality_preset` | string | enum: `ultra`, `high`, `medium` | Encode quality |
| `file_size_gb` | float | ≥0 | File size in GB |
| `duration_sec` | int | ≥0 | Duration in seconds |
| `local_path` | string | max 2048 chars | NVMe storage path |
| `cloud_sync_status` | string | enum: `pending`, `syncing`, `complete`, `failed` | Background sync status |
| `cloud_url` | string | nullable, URL | Cloud storage URL after sync |
| `hdr_metadata` | JSONB | nullable | HDR10/HDR10+/Dolby Vision SEI |
| `created_at` | timestamp | auto-gen | Creation time |

**Relationships:**
- One Session → One Recording (1:1, optional)
- One User → Many Recordings

---

## 8. Controller

Input device state forwarded from client to host.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique controller identifier |
| `session_id` | UUID | FK → Session.id | Active session |
| `device_type` | string | enum: `dualsense`, `xbox`, `generic` | Controller type |
| `connection_type` | string | enum: `usb`, `dongle_2_4ghz`, `bluetooth` | Physical connection |
| `haptics_enabled` | bool | default false | Haptic feedback active |
| `adaptive_trigger_enabled` | bool | default false | Adaptive trigger active |
| `gyro_enabled` | bool | default false | Gyroscope forwarding active |
| `audio_jack_enabled` | bool | default false | Audio jack passthrough active |
| `poll_rate_hz` | int | default 1000, ≥125 | USB polling rate |
| `latency_us` | int | nullable | Measured input latency in microseconds |
| `connected_at` | timestamp | auto-gen | Connection time |
| `disconnected_at` | timestamp | nullable | Disconnection time |

**Validation:**
- `poll_rate_hz` must be 125, 250, 500, or 1000
- `latency_us` ≤ 1000 (1ms target)
- `disconnected_at` ≥ `connected_at`

---

## 9. Capability (Handshake)

Codec/transport/GPU capability advertisement exchanged during session negotiation.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique capability record |
| `session_id` | UUID | FK → Session.id | Negotiation session |
| `side` | string | enum: `host`, `client` | Which side advertised |
| `codecs_offered` | string[] | enum: `h264`, `hevc`, `av1` | Codecs supported |
| `transports_offered` | string[] | enum: `webrtc`, `quic`, `udp` | Transports supported |
| `resolutions_offered` | string[] | enum: `720p`, `1080p`, `1440p`, `4k` | Resolutions supported |
| `selected_codec` | string | nullable | Mutually selected codec |
| `selected_transport` | string | nullable | Mutually selected transport |
| `selected_resolution` | string | nullable | Mutually selected resolution |
| `negotiated_at` | timestamp | nullable | Handshake completion time |

**Validation:**
- `selected_codec` must be in intersection of host and client `codecs_offered`
- `selected_transport` must be in intersection of host and client `transports_offered`
- `negotiated_at` required when any `selected_*` field is set

---

## 10. Asset

4K media asset managed by the catalog system.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK, auto-gen | Unique asset identifier |
| `game_id` | UUID | FK → Game.id | Parent game |
| `asset_type` | string | enum: `cover_art`, `screenshot`, `video_trailer`, `icon` | Asset category |
| `cdn_url` | string | URL, max 2048 chars | CloudFront signed URL |
| `cdn_expiry` | timestamp | nullable | URL expiration |
| `width_px` | int | ≥0 | Image width |
| `height_px` | int | ≥0 | Image height |
| `file_size_mb` | float | ≥0 | Asset size |
| `format` | string | enum: `jpg`, `png`, `webp`, `mp4`, `webm` | File format |
| `created_at` | timestamp | auto-gen | Upload time |

---

## Junction Tables

### HostGame
Maps installed games to hosts.

| Field | Type | Constraints |
|-------|------|-------------|
| `host_id` | UUID | FK → Host.id |
| `game_id` | UUID | FK → Game.id |
| `installed_at` | timestamp | auto-gen |
| `last_played_at` | timestamp | nullable |
| `play_count` | int | default 0, ≥0 |

**PK:** (`host_id`, `game_id`)

### TenantGameFilter
Maps catalog visibility to tenants.

| Field | Type | Constraints |
|-------|------|-------------|
| `tenant_id` | UUID | FK → Tenant.id |
| `game_id` | UUID | FK → Game.id |
| `visible` | bool | default true |
| `custom_price_cents` | int | nullable, ≥0 |

**PK:** (`tenant_id`, `game_id`)

---

## State Transition Summary

| Entity | States | Transitions | Triggers |
|--------|--------|-------------|----------|
| Host | `offline`, `online`, `streaming`, `maintenance` | 6 | Heartbeat, session events, operator action |
| Session | `connecting`, `negotiating`, `streaming`, `paused`, `reconnecting`, `disconnected`, `ended` | 9 | Network, user input, host events |
| Recording | `pending`, `syncing`, `complete`, `failed` | 4 | Background sync job status |
| GPU | implicit via `session_count` | N/A | Session start/end |

---

*End of Data Model*
