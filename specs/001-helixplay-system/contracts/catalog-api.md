# Contract: Catalogizer API

**Feature Branch**: `001-helixplay-system`  
**Created**: 2026-05-02  
**Status**: Draft — Phase 1 Design Artifact  
**Decoupling**: Catalogizer (`HelixDevelopment/Catalogizer`) is a **standalone submodule**. HelixPlay consumes this API with no direct code coupling.  
**Source**: `specs/001-helixplay-system/spec.md` §User Story 9, FR-026/FR-027, C-002, DEP-024

---

## 1. Protocol

- **Primary**: gRPC (`catalogizer.v1.CatalogService`) with Protocol Buffers v3
- **Fallback**: REST JSON over HTTP/2 (auto-generated via gRPC-Gateway or manual mapping)
- **Transport security**: mTLS between services; JWT Bearer for client-facing endpoints
- **Compression**: Brotli for REST; native gRPC compression (gzip or zstd) for gRPC

---

## 2. gRPC Service Definition

```protobuf
syntax = "proto3";
package catalogizer.v1;

option go_package = "github.com/HelixDevelopment/Catalogizer/pkg/api/v1;catalogizerv1";

service CatalogService {
  // Game metadata
  rpc ListGames(ListGamesRequest) returns (ListGamesResponse);
  rpc GetGame(GetGameRequest) returns (Game);
  rpc SearchGames(SearchGamesRequest) returns (SearchGamesResponse);
  rpc GetGameAssets(GetGameAssetsRequest) returns (GameAssets);

  // Tenant-scoped catalog management
  rpc GetTenantCatalog(GetTenantCatalogRequest) returns (TenantCatalog);
  rpc UpdateTenantCatalogFilter(UpdateTenantCatalogFilterRequest) returns (TenantCatalog);

  // Health
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## 3. Endpoints & Methods

### 3.1 `ListGames`

Paginated list of games with filtering.

**gRPC**: `catalogizer.v1.CatalogService/ListGames`  
**REST**: `GET /v1/games`

#### Request (`ListGamesRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tenant_id` | string (UUID) | YES | Valid UUID v5 |
| `page_size` | int32 | NO | 1–100, default 20 |
| `page_token` | string | NO | Opaque cursor from previous response |
| `filter` | GameFilter | NO | See below |
| `order_by` | string | NO | `title`, `release_date`, `popularity`, `last_played`; prefix `-` for DESC |

**`GameFilter`**:

| Field | Type | Description |
|-------|------|-------------|
| `store` | repeated string | `steam`, `epic`, `gog`, `ubisoft`, `battlenet`, `origin`, `microsoft`, `standalone` |
| `genre` | repeated string | Controlled vocabulary |
| `hdr` | bool | Filter by HDR support |
| `dualsense` | bool | Filter by DualSense support |
| `installed_only` | bool | TRUE = only games with `is_installed = true` |
| `search_query` | string | Full-text search across title, developer, publisher |

#### Response (`ListGamesResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `games` | repeated Game | Max `page_size` items |
| `next_page_token` | string | Empty if last page |
| `total_size` | int64 | Total matching count (approximate if >10 000) |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `INVALID_TENANT` | `INVALID_ARGUMENT` | 400 | `tenant_id` malformed or unknown |
| `FILTER_REJECTED` | `INVALID_ARGUMENT` | 400 | Filter violates tenant `catalog_filter` rules |
| `RATE_LIMITED` | `RESOURCE_EXHAUSTED` | 429 | See Rate Limits |

---

### 3.2 `GetGame`

Retrieve single game metadata.

**gRPC**: `catalogizer.v1.CatalogService/GetGame`  
**REST**: `GET /v1/games/{game_id}`

#### Request (`GetGameRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tenant_id` | string (UUID) | YES | Valid UUID v5 |
| `game_id` | string (UUID) | YES | Valid UUID v5 |
| `include_assets` | bool | NO | Include presigned asset URLs |

#### Response (`Game`)

Full game message (see schema in §5).

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `GAME_NOT_FOUND` | `NOT_FOUND` | 404 | Game does not exist or not in tenant catalog |
| `TENANT_MISMATCH` | `PERMISSION_DENIED` | 403 | Game belongs to different tenant |

---

### 3.3 `SearchGames`

Full-text search with relevance ranking.

**gRPC**: `catalogizer.v1.CatalogService/SearchGames`  
**REST**: `GET /v1/games:search`

#### Request (`SearchGamesRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tenant_id` | string (UUID) | YES | Valid UUID v5 |
| `query` | string | YES | 1–128 UTF-8 characters, trimmed |
| `page_size` | int32 | NO | 1–50, default 10 |
| `page_token` | string | NO | Opaque cursor |
| `facets` | repeated string | NO | `genre`, `store`, `hdr`, `year` |

#### Response (`SearchGamesResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `results` | repeated Game | Ranked by relevance score |
| `facets` | map<string, Facet> | Aggregation buckets per requested facet |
| `next_page_token` | string | — |
| `total_size` | int64 | — |
| `latency_ms` | int32 | Server-side query latency (for SLO monitoring) |

**SLA**: p999 ≤ 200 ms (SC-010).

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `QUERY_EMPTY` | `INVALID_ARGUMENT` | 400 | `query` empty or whitespace only |
| `QUERY_TOO_LONG` | `INVALID_ARGUMENT` | 400 | `query` > 128 chars |
| `SEARCH_UNAVAILABLE` | `UNAVAILABLE` | 503 | VectorDB or RAG backend down |

---

### 3.4 `GetGameAssets`

Retrieve CDN-signed URLs for 4K assets.

**gRPC**: `catalogizer.v1.CatalogService/GetGameAssets`  
**REST**: `GET /v1/games/{game_id}/assets`

#### Request (`GetGameAssetsRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tenant_id` | string (UUID) | YES | Valid UUID v5 |
| `game_id` | string (UUID) | YES | Valid UUID v5 |
| `asset_types` | repeated string | NO | `cover`, `screenshot`, `video`, `icon`, `banner` |
| `max_resolution` | string | NO | `1080p`, `1440p`, `4k`, `8k` |
| `format` | string | NO | `webp`, `png`, `jpg`, `avif` |

#### Response (`GameAssets`)

| Field | Type | Description |
|-------|------|-------------|
| `game_id` | string | — |
| `assets` | repeated Asset | Signed URLs with expiry |
| `cdn_ttl_seconds` | int32 | Validity of signed URLs |

**`Asset`**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `cover`, `screenshot`, etc. |
| `url` | string | HTTPS CloudFront signed URL |
| `width` | int32 | Pixel width |
| `height` | int32 | Pixel height |
| `size_bytes` | int64 | — |
| `format` | string | `webp`, `png`, etc. |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `ASSET_NOT_FOUND` | `NOT_FOUND` | 404 | Game exists but requested assets missing |
| `CDN_SIGNING_FAILED` | `INTERNAL` | 500 | CloudFront signing error |

---

### 3.5 `GetTenantCatalog`

Retrieve tenant's effective catalog configuration.

**gRPC**: `catalogizer.v1.CatalogService/GetTenantCatalog`  
**REST**: `GET /v1/tenants/{tenant_id}/catalog`

#### Request (`GetTenantCatalogRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tenant_id` | string (UUID) | YES | Valid UUID v5 |

#### Response (`TenantCatalog`)

| Field | Type | Description |
|-------|------|-------------|
| `tenant_id` | string | — |
| `filter` | CatalogFilter | Effective merged filter |
| `total_games` | int64 | Count of games visible to tenant |
| `last_sync_at` | timestamp | Last metadata sync from store integrations |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `TENANT_NOT_FOUND` | `NOT_FOUND` | 404 | Tenant unknown |

---

### 3.6 `UpdateTenantCatalogFilter`

Admin-only: update tenant catalog filter.

**gRPC**: `catalogizer.v1.CatalogService/UpdateTenantCatalogFilter`  
**REST**: `PATCH /v1/tenants/{tenant_id}/catalog`

#### Auth Requirements

- Scope: `catalog:admin` or `tenant_admin` role
- RBAC: `tenant_id` in JWT `tenant_ids` claim, or `super_admin`

#### Request (`UpdateTenantCatalogFilterRequest`)

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `tenant_id` | string (UUID) | YES | Valid UUID v5 |
| `filter` | CatalogFilter | YES | Must be subset of global catalog |

#### Error Codes

| Code | gRPC Status | HTTP | When |
|------|-------------|------|------|
| `FORBIDDEN` | `PERMISSION_DENIED` | 403 | Insufficient RBAC |
| `FILTER_TOO_BROAD` | `INVALID_ARGUMENT` | 400 | Filter would expose unauthorized games |

---

### 3.7 `Health`

Liveness/readiness probe.

**gRPC**: `catalogizer.v1.CatalogService/Health`  
**REST**: `GET /v1/health`

#### Request (`HealthRequest`)

| Field | Type | Required |
|-------|------|----------|
| `service` | string | NO | `catalog`, `search`, `cdn`, `all` |

#### Response (`HealthResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | `healthy`, `degraded`, `unhealthy` |
| `checks` | repeated HealthCheck | Per-dependency status |

---

## 4. Rate Limits

| Endpoint | Rate Limit | Burst | Scope |
|----------|------------|-------|-------|
| `ListGames` | 100 req/min | 20 | Per `tenant_id` + client IP |
| `GetGame` | 300 req/min | 50 | Per `tenant_id` + client IP |
| `SearchGames` | 60 req/min | 10 | Per `tenant_id` + client IP |
| `GetGameAssets` | 200 req/min | 30 | Per `tenant_id` + client IP |
| `GetTenantCatalog` | 30 req/min | 5 | Per `tenant_id` |
| `UpdateTenantCatalogFilter` | 10 req/min | 2 | Per `tenant_id` + user ID |

Enforced by `vasic-digital/RateLimiter` (token bucket + sliding window).

---

## 5. Message Schemas

### 5.1 `Game`

```protobuf
message Game {
  string game_id = 1;
  string tenant_id = 2;
  string store = 3;
  string store_app_id = 4;
  string title = 5;
  string slug = 6;
  string description = 7;
  string developer = 8;
  string publisher = 9;
  string release_date = 10; // ISO-8601 date
  repeated string genres = 11;
  repeated string tags = 12;
  string cover_art_url = 13;
  repeated string screenshot_urls = 14;
  repeated string video_urls = 15;
  SystemRequirements min_requirements = 16;
  SystemRequirements rec_requirements = 17;
  bool supports_hdr = 18;
  bool supports_atmos = 19;
  bool supports_dualsense = 20;
  repeated string supported_controllers = 21;
  bool is_installed = 22;
  int64 playtime_seconds = 23;
  string last_played_at = 24; // ISO-8601 timestamp
  map<string, string> metadata = 25;
}

message SystemRequirements {
  string gpu = 1;
  int32 ram_gb = 2;
  string os = 3;
  string storage_gb = 4;
}
```

### 5.2 `CatalogFilter`

```protobuf
message CatalogFilter {
  repeated string allowed_stores = 1;
  repeated string allowed_genres = 2;
  repeated string excluded_tags = 3;
  int32 rating_max = 4; // ESRB/PEGI age rating max
  bool require_hdr = 5;
  bool require_dualsense = 6;
}
```

---

## 6. Auth Requirements

| Endpoint | Auth Method | Required Scopes | RBAC |
|----------|-------------|-----------------|------|
| `ListGames` | JWT Bearer | `catalog:read` | Any authenticated user in tenant |
| `GetGame` | JWT Bearer | `catalog:read` | Any authenticated user in tenant |
| `SearchGames` | JWT Bearer | `catalog:read` | Any authenticated user in tenant |
| `GetGameAssets` | JWT Bearer | `catalog:read` | Any authenticated user in tenant |
| `GetTenantCatalog` | JWT Bearer | `catalog:read` | `tenant_admin` or `super_admin` |
| `UpdateTenantCatalogFilter` | JWT Bearer | `catalog:admin` | `tenant_admin` or `super_admin` |
| `Health` | None | — | Public |

JWT validation delegated to `vasic-digital/Auth` submodule.

---

## 7. Error Response Schema (REST)

All REST errors return JSON:

```json
{
  "error": {
    "code": "GAME_NOT_FOUND",
    "message": "Game '550e8400-e29b-41d4-a716-446655440000' not found in tenant catalog",
    "details": {
      "tenant_id": "...",
      "game_id": "..."
    },
    "request_id": "req_abc123",
    "retry_after_seconds": 0
  }
}
```

---

## 8. Anti-Bluff Verification

### Sources
| Source | Lines | Insights Used |
|--------|-------|---------------|
| `specs/001-helixplay-system/spec.md` | 651 | User Story 9 (Catalog), FR-026/FR-027, C-002 (CloudFront), DEP-024 (Catalogizer decoupled) |
| `AGENTS.md` | 286 | gRPC preferred, REST as separate microservice, Brotli compression, JWT auth |

### Conflict Resolution
- Catalogizer remains **decoupled** per spec: this contract is the only integration surface. HelixPlay does not import Catalogizer packages directly; it uses generated gRPC client stubs or REST.
- CloudFront signed URLs (C-002) implemented in `GetGameAssets` response.
- Rate limits scoped per `tenant_id` to prevent noisy-neighbor issues in multi-tenant deployments.

### Verification
- ✅ All 7 methods documented with request/response schemas.
- ✅ Error codes enumerated per method.
- ✅ Rate limits specified per endpoint.
- ✅ Auth requirements mapped to scopes and RBAC.
- ✅ REST fallback paths provided for every gRPC method.
- ✅ No `TODO`/`FIXME`/`placeholder` present.
