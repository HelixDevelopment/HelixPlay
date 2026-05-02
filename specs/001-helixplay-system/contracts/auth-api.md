# Contract: Auth API

**Feature Branch**: `001-helixplay-system`  
**Created**: 2026-05-02  
**Status**: Draft — Phase 1 Design Artifact  
**Decoupling**: Authentication delegated to `vasic-digital/Auth` submodule. HelixPlay consumes this contract.  
**Source**: `specs/001-helixplay-system/spec.md` §User Stories 5, 8, FR-017, FR-024, FR-025, C-001, DEP-003

---

## 1. Protocol

- **Primary**: OAuth2/OIDC over HTTPS (REST JSON)
- **Token format**: JWT (JWS RS256), `golang-jwt/jwt/v5`
- **Identity Provider**: Auth0 (C-001) or Supabase with Auth0 integration
- **Device authorization**: RFC 8628 for input-constrained devices (TV, console)
- **Service-to-service**: mTLS + short-lived JWT service tokens

---

## 2. OAuth2 Flows

### 2.1 Authorization Code + PKCE (Web/Desktop/Mobile)

```
Client (Wails/Flutter/Angular)
  │
  │── (1) GET /authorize?response_type=code&client_id=...&redirect_uri=...&code_challenge=... ──► Auth0
  │                                                                                               │
  │◄── (2) 302 redirect to Auth0 login page ──────────────────────────────────────────────────────│
  │                                                                                               │
  │── (3) User authenticates ─────────────────────────────────────────────────────────────────────►
  │                                                                                               │
  │◄── (4) 302 redirect to redirect_uri?code=...&state=... ───────────────────────────────────────│
  │
  │── (5) POST /oauth/token {grant_type=authorization_code, code, code_verifier, redirect_uri} ──► Auth0
  │                                                                                               │
  │◄── (6) 200 {access_token, refresh_token, id_token, expires_in} ───────────────────────────────│
  │
  │── (7) access_token → HelixPlay API (Authorization: Bearer ...) ──► HelixPlay Services
```

### 2.2 Device Authorization Grant (RFC 8628) — TV/Constrained Devices

```
TV Client
  │
  │── (1) POST /oauth/device/code {client_id, scope, audience} ──► Auth0
  │
  │◄── (2) 200 {device_code, user_code, verification_uri, expires_in, interval}
  │
  │── (3) Display user_code + QR code to verification_uri_complete
  │
  │── (4) Poll POST /oauth/token {grant_type=urn:ietf:params:oauth:grant-type:device_code, device_code}
  │       ► Auth0
  │
  │◄── (5a) 200 {access_token, refresh_token, id_token}  (when user completes auth)
  │◄── (5b) 428 {error: authorization_pending}            (keep polling)
  │◄── (5c) 428 {error: slow_down}                        (increase interval)
  │◄── (5d) 428 {error: expired_token}                    (flow aborted)
```

### 2.3 Client Credentials (Service-to-Service)

```
Host Agent / Microservice
  │
  │── (1) POST /oauth/token {grant_type=client_credentials, client_id, client_secret, audience=helixplay-api}
  │       ► Auth0
  │
  │◄── (2) 200 {access_token, expires_in}
  │
  │── (3) access_token → HelixPlay gRPC metadata
```

---

## 3. REST Endpoints

All endpoints are implemented by `vasic-digital/Auth` and proxied/validated by HelixPlay services.

### 3.1 `POST /v1/auth/token`

Token exchange and refresh.

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `grant_type` | string | YES | `authorization_code`, `refresh_token`, `client_credentials`, `urn:ietf:params:oauth:grant-type:device_code` |
| `client_id` | string | YES | Auth0 application client ID |
| `code` | string | NO | Authorization code (for `authorization_code`) |
| `code_verifier` | string | NO | PKCE verifier (for `authorization_code`) |
| `redirect_uri` | string | NO | Must match registration |
| `refresh_token` | string | NO | For `refresh_token` grant |
| `device_code` | string | NO | For device authorization grant |

#### Response (`TokenResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `access_token` | string | JWT, RS256 signed |
| `refresh_token` | string | Opaque, rotatable |
| `id_token` | string | OIDC JWT with user claims |
| `token_type` | string | `Bearer` |
| `expires_in` | int32 | Seconds until access_token expiry |
| `scope` | string | Granted scopes space-separated |

#### Error Codes

| Code | HTTP | When |
|------|------|------|
| `invalid_request` | 400 | Missing required parameter |
| `invalid_client` | 401 | Client authentication failed |
| `invalid_grant` | 403 | Code expired or reused |
| `unauthorized_client` | 403 | Grant type not allowed for client |
| `unsupported_grant_type` | 400 | Unknown `grant_type` |
| `authorization_pending` | 428 | Device flow: user not yet authorized |
| `slow_down` | 428 | Device flow: polling too fast |
| `expired_token` | 428 | Device flow: `device_code` expired |

---

### 3.2 `POST /v1/auth/device/code`

Initiate device authorization flow.

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `client_id` | string | YES | Auth0 application client ID |
| `scope` | string | NO | Space-separated scopes |
| `audience` | string | YES | `helixplay-api` |

#### Response (`DeviceCodeResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `device_code` | string | Opaque, used in token polling |
| `user_code` | string | Short alphanumeric, displayed to user |
| `verification_uri` | string | URL where user enters `user_code` |
| `verification_uri_complete` | string | `verification_uri?user_code=...` |
| `expires_in` | int32 | Seconds until `device_code` expiry |
| `interval` | int32 | Minimum polling interval in seconds |

---

### 3.3 `POST /v1/auth/logout`

Revoke tokens and end session.

#### Auth

- JWT Bearer (any valid access token)

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `refresh_token` | string | YES | Token to revoke |

#### Response

`204 No Content` on success.

---

### 3.4 `GET /v1/auth/userinfo`

Retrieve current user profile.

#### Auth

- JWT Bearer

#### Response (`UserInfo`)

| Field | Type | Description |
|-------|------|-------------|
| `sub` | string | Auth0 user ID (maps to `oauth_sub` in data model) |
| `tenant_id` | string | HelixPlay tenant UUID |
| `email` | string | Verified email |
| `email_verified` | bool | — |
| `name` | string | Display name |
| `picture` | string | Avatar URL |
| `roles` | string[] | `player`, `host_admin`, `tenant_admin`, `super_admin` |
| `permissions` | string[] | Fine-grained API permissions |

---

### 3.5 `POST /v1/auth/introspect`

Token introspection (RFC 7662) for resource servers.

#### Auth

- Basic Auth (client_id:client_secret) or mTLS

#### Request

| Field | Type | Required |
|-------|------|----------|
| `token` | string | YES |
| `token_type_hint` | string | NO | `access_token` or `refresh_token` |

#### Response (`IntrospectionResponse`)

| Field | Type | Description |
|-------|------|-------------|
| `active` | bool | FALSE if revoked or expired |
| `scope` | string | — |
| `client_id` | string | — |
| `username` | string | — |
| `token_type` | string | `Bearer` |
| `exp` | int64 | Unix timestamp |
| `iat` | int64 | Unix timestamp |
| `sub` | string | User ID |
| `tenant_id` | string | HelixPlay tenant |

---

## 4. JWT Claims Schema

### 4.1 Access Token

```json
{
  "sub": "auth0|abc123",
  "iss": "https://helixplay.auth0.com/",
  "aud": "helixplay-api",
  "iat": 1714645200,
  "exp": 1714648800,
  "scope": "catalog:read stream:negotiate recording:write",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "roles": ["player", "host_admin"],
  "permissions": ["game:launch", "session:terminate"],
  "client_id": "xyz789",
  "jti": "unique-token-id"
}
```

### 4.2 ID Token (OIDC)

```json
{
  "sub": "auth0|abc123",
  "iss": "https://helixplay.auth0.com/",
  "aud": "helixplay-client-id",
  "iat": 1714645200,
  "exp": 1714648800,
  "email": "player@example.com",
  "email_verified": true,
  "name": "PlayerOne",
  "picture": "https://cdn.example.com/avatars/abc123.png",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 4.3 Validation Rules

| Claim | Rule | Enforcement |
|-------|------|-------------|
| `iss` | Must match allowed issuers list (`https://helixplay.auth0.com/`) | `vasic-digital/Auth` middleware |
| `aud` | Must include `helixplay-api` | Middleware |
| `exp` | Must be > now() | Middleware |
| `iat` | Must be ≤ now() + 60 s clock skew | Middleware |
| `scope` | Space-separated, validated per endpoint | RBAC middleware |
| `tenant_id` | Must match resource tenant (for tenant-scoped endpoints) | Middleware |
| `jti` | Checked against revocation list (Redis) | Middleware |

---

## 5. RBAC Model

### 5.1 Roles

| Role | Description | Inherits |
|------|-------------|----------|
| `player` | Standard user | — |
| `host_admin` | Can manage own hosts | `player` |
| `tenant_admin` | Can manage tenant config, users, catalog filter | `host_admin` |
| `super_admin` | Platform-wide access | `tenant_admin` |

### 5.2 Scopes

| Scope | Description | Required Role |
|-------|-------------|---------------|
| `catalog:read` | Browse game catalog | `player` |
| `catalog:admin` | Modify tenant catalog filter | `tenant_admin` |
| `game:launch` | Launch games on host | `player` |
| `stream:negotiate` | Create streaming sessions | `player` |
| `stream:terminate` | End sessions | `player` |
| `recording:write` | Start/stop recordings | `player` |
| `recording:read` | View own recordings | `player` |
| `recording:admin` | View tenant recordings | `tenant_admin` |
| `host:register` | Register new host | `host_admin` |
| `host:heartbeat` | Host agent self-report | Host token only |
| `host:read` | View host telemetry | `host_admin` |
| `host:admin` | Manage any host in tenant | `tenant_admin` |
| `user:read` | View user profiles | `player` (self) / `tenant_admin` (any in tenant) |
| `user:admin` | Manage users | `tenant_admin` |

---

## 6. Rate Limits

| Endpoint | Rate Limit | Burst | Scope |
|----------|------------|-------|-------|
| `POST /v1/auth/token` | 30/min | 5 | Per IP |
| `POST /v1/auth/device/code` | 10/min | 2 | Per IP |
| `GET /v1/auth/userinfo` | 120/min | 20 | Per `sub` |
| `POST /v1/auth/introspect` | 300/min | 50 | Per client_id |
| `POST /v1/auth/logout` | 30/min | 5 | Per `sub` |

Token refresh rate: unlimited for valid refresh tokens, but refresh token rotation limits to 1 concurrent refresh per token family.

---

## 7. Auth Requirements Summary

| Endpoint | Auth | Notes |
|----------|------|-------|
| `POST /v1/auth/token` | Public | Rate-limited by IP |
| `POST /v1/auth/device/code` | Public | Rate-limited by IP |
| `POST /v1/auth/logout` | JWT Bearer | Revokes caller's tokens |
| `GET /v1/auth/userinfo` | JWT Bearer | Returns caller's profile |
| `POST /v1/auth/introspect` | Basic Auth or mTLS | For resource servers |

---

## 8. Multi-Tenant OAuth2 Configuration

Per-tenant Auth0 configuration (stored in `Tenant.oauth_config`):

```json
{
  "domain": "partner-example.auth0.com",
  "client_id": "abc123",
  "audience": "helixplay-api",
  "connection": "Username-Password-Authentication",
  "social_connections": ["google-oauth2", "steam"],
  "enterprise_connections": ["samlp"],
  "branding": {
    "logo_url": "https://cdn.example.com/logo.png",
    "primary_color": "#FF6600"
  }
}
```

HelixPlay API validates JWT `iss` against registered tenant domains. Unauthorized issuer → `401 invalid_issuer`.

---

## 9. Error Response Schema

```json
{
  "error": "invalid_grant",
  "error_description": "The provided authorization grant is invalid, expired, or revoked.",
  "error_uri": "https://docs.helixplay.dev/errors/invalid_grant",
  "request_id": "req_xyz789"
}
```

---

## 10. Anti-Bluff Verification

### Sources
| Source | Lines | Insights Used |
|--------|-------|---------------|
| `specs/001-helixplay-system/spec.md` | 651 | User Stories 5, 8; FR-017, FR-024, FR-025; C-001 (Auth0); DEP-003 |
| `AGENTS.md` | 286 | JWT (`golang-jwt/jwt/v5`), OAuth2/OIDC, RBAC, mTLS |

### Conflict Resolution
- **Auth0 vs custom IdP**: Auth0 chosen per C-001; custom OIDC supported via `Tenant.oauth_config` fallback.
- **Device authorization grant**: Explicitly required for TV clients (RFC 8628); documented with full poll sequence.
- **Service-to-service auth**: mTLS preferred over shared secrets; `client_credentials` documented for bootstrap scenarios.

### Verification
- ✅ All 3 OAuth2 flows documented: Authorization Code + PKCE, Device Grant, Client Credentials.
- ✅ JWT claims schema with validation rules.
- ✅ RBAC roles and scopes enumerated.
- ✅ Rate limits per endpoint specified.
- ✅ Multi-tenant Auth0 configuration schema provided.
- ✅ No `TODO`/`FIXME`/`placeholder` present.
