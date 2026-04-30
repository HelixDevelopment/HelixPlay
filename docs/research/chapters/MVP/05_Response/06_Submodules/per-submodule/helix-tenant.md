# `helix-tenant` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-tenant`                                                                                                         |
| **Origin chapter:section**  | [C11 §6](../../03_Architecture/10_WhiteLabel_and_Theming.md) — *White-label theming + per-tenant configuration*       |
| **Public path (4 mirrors)** | `vasic-digital/helix-tenant` on GitHub + GitLab + GitFlic + GitVerse                                                   |
| **Direct deps (vasic-digital)** | **(none — pure-Go theming engine)**                                                                                |
| **External Go deps**        | `gopkg.in/yaml.v3`, `github.com/BurntSushi/toml`, `text/template`, `html/template`, `golang.org/x/text/language`        |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `tenant-theming-1.x` — builder `golang-builder`, runtime `distroless-static`                                        |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/04_white_label_multi_tenant/scenarios/01_tenant_theme_swap_at_session_boundary.scenario.yaml`           |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **0** (no `vasic-digital` deps; one of three depth-0 submodules)                                                  |
| **R-18 SafeExec abstention**| **Yes — does NOT import `helix-r18-safeexec`.** Pure-Go theming engine; no subprocess.                                  |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-tenant` is the **white-label theming + per-tenant configuration engine** for HelixPlay. It loads tenant-specific theme bundles (colours, typography, logos, layout overrides), applies them to client + server-rendered surfaces, manages per-tenant feature flags, and tracks tenant-scoped i18n bundles. The submodule was introduced in [C11 §6](../../03_Architecture/10_WhiteLabel_and_Theming.md) as the consolidation of theming-engine logic that earlier projects implemented inline in their handlers.

The white-label posture is operationally significant: HelixPlay's revenue model includes per-tenant licensing, and a tenant's brand identity (colours, logo, custom domain) must be settable without code changes. The submodule's API is therefore stable across many tenants — adding a tenant is a configuration change, not a code change.

`★ Why a Go submodule and not a JSON-driven config layer.` C11 §6 explores three design alternatives: (1) JSON config consumed by clients only (rejected because server-side renders need theme tokens too), (2) Go template engine consumed by both (chosen — provides type-safety + compile-time error detection), (3) WASM-driven theming (rejected because the developer ergonomics regression is steep). The Go template engine plus the YAML / TOML loaders gives type-safety at code-review time, full template flexibility at render time, and trivial integration with both the Go-rendered server side and the Wails / Compose-for-TV clients (via the `gomobile bind`-generated AAR).

---

## 2. Public API Surface

### 2.1 The `Tenant` type

```go
package tenant

// Tenant is a HelixPlay tenant: a white-label customer with their
// own theme, feature flags, locale set, and per-tenant catalog
// scope.
type Tenant struct {
    ID         string
    Name       string
    Theme      *Theme
    Features   FeatureFlags
    Locales    []language.Tag
    CatalogScope CatalogScopeRef
    Custom     map[string]any  // tenant-specific extension fields, validated against schema
}
```

### 2.2 The `Theme` type

```go
package tenant

// Theme is a tenant's visual identity: design tokens, layout
// overrides, asset references.
type Theme struct {
    DesignTokens DesignTokens   // colours, typography, spacing, radius
    Logo         AssetRef       // SVG or PNG with multiple resolutions
    Layout       LayoutOverrides // grid templates, list densities
    Animations   AnimationOverrides
}

func (t *Theme) RenderHTML(name string, data any) (template.HTML, error)
func (t *Theme) ApplyToClient(client ClientHandle) error  // for Compose-for-TV / Wails
```

### 2.3 The `Loader` interface

```go
package tenant

// Loader resolves a tenant from a domain name, an OAuth token, or
// an explicit tenant ID. Cached aggressively per Constitution §10
// observability + per-tenant rate-limit semantics.
type Loader interface {
    LoadByID(ctx context.Context, id string) (*Tenant, error)
    LoadByDomain(ctx context.Context, domain string) (*Tenant, error)
    Reload(ctx context.Context, id string) error  // for hot-reload during dev / on config change
}
```

### 2.4 The `FeatureFlags` evaluator

```go
package tenant

// FeatureFlags is per-tenant feature configuration.
type FeatureFlags map[string]any

func (f FeatureFlags) Bool(key string, defaultVal bool) bool
func (f FeatureFlags) Int(key string, defaultVal int) int
func (f FeatureFlags) String(key string, defaultVal string) string
```

The flags are deliberately untyped at the API surface (just `map[string]any`) because feature-flag values are highly heterogeneous (booleans, sample-rates, enabled-region lists). Type-safe accessors are layered on top.

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital` — none. Why no `helix-r18-safeexec`?

`helix-tenant` is a **pure-Go theming engine**. It loads bundles from disk (or from S3 / object storage in production), parses YAML / TOML, and renders Go templates. It does not invoke subprocesses. Per [S01 §3.3](../01_Submodule_Catalog.md#33-the-none-dependency-rows) and [S01 §7.1](../01_Submodule_Catalog.md#71-the-ladder)'s "ABSTAIN" group, this submodule has no `vasic-digital` dependency.

The abstention is not a relaxation of R-18; the submodule has no subprocess surface to regulate. The five-layer R-18 enforcement still applies at layers 1, 4, 5 (chapter prose, ripgrep CI lint, host-integrity-scan).

### 3.2 External (Go)

- `gopkg.in/yaml.v3` — YAML parser for theme bundles.
- `github.com/BurntSushi/toml` — TOML parser for alternative theme format.
- `text/template`, `html/template` — Go standard library template engines.
- `golang.org/x/text/language` — locale handling (BCP 47 tags, fallback chains).
- `github.com/golang/protobuf/proto` — protobuf reflection for theme-bundle schema validation.

### 3.3 External (system)

A bundle store (filesystem path, S3 bucket, or object-storage equivalent). Production deployments use S3-compatible storage per Constitution §3 (R-05 — Containerised Runtime; the bundle store is itself containerised when running on Russian-jurisdiction operators that cannot use AWS S3 directly).

---

## 4. Container Build (S02 §3 lane: `tenant-theming-1.x`)

**Builder:** `golang-builder` (no cgo). **Runtime:** `distroless-static`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2.

The container ships with **no embedded theme bundles**; production deployments mount the bundle store as a read-only volume or fetch at startup. This keeps the image generic across tenants — one image, many tenants.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
- YAML / TOML parsing for valid + invalid bundles.
- Template rendering with edge-case data.
- Locale fallback chain (e.g. `sr-Latn-RS` → `sr-Latn` → `sr` → `en`).
- Feature-flag accessors with type mismatches.

### 5.2 Integration
- Real bundle store (containerised MinIO for S3-compat); `LoadByDomain` + `LoadByID` round-trip.

### 5.3 E2E
- Full client startup → tenant load → theme apply → render a page → SHA-256 of rendered HTML matches expected.

### 5.4 Security
- govulncheck + Snyk + Trivy.
- Custom: attempted theme-bundle escape (e.g. `{{exec "rm -rf /"}}` injected into a template — must fail at template parse time because exec is not a registered template function).
- SSTI (server-side template injection) fuzzer.

### 5.5 Benchmarking
- Tenant load p999 ≤ 100 µs (cached); ≤ 50 ms (cold from S3).

### 5.6 Chaos
- S3 partition during tenant load; verify graceful fallback to last-known-good cache.

### 5.7 Stress
- 24-hour run with 100 concurrent tenant loads / second; verify no cache thrashing.

### 5.8 Smoke
- 30-second post-deploy: load a known tenant; render a known template; verify SHA-256.

### 5.9 Full Automation
- §5.1–§5.8.

### 5.10 Challenges
- `04_white_label_multi_tenant/01_tenant_theme_swap_at_session_boundary.scenario` — active session with tenant A; theme swap to tenant B at session boundary; verify no visual artefact, no auth-token bleeding, no theme-cache corruption.

---

## 6. Challenges Entry-Point (S03 §4 row #05)

**Topology:** `04_white_label_multi_tenant`. **Scenario:** `01_tenant_theme_swap_at_session_boundary.scenario.yaml`. **Why this scenario.** Tenant theme-swap is the highest-risk multi-tenant operation; the scenario specifically exercises the cross-tenant isolation guarantee. **Baseline:** SHA-256 of rendered frames, OTLP trace topology, zero cross-tenant cache hits.

---

## 7. R-18 Inheritance

`helix-tenant` **abstains** from R-18 SafeExec inheritance because it invokes no subprocesses. Same posture as `helix-vault` (§7 of that descriptor). The R-18 layer-4 ripgrep lint is the active guard against future regressions; a PR that introduces an `os/exec` import would fail CI.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation gated on the C11 §6 API freeze + Constitution §10 *Documentation Discipline* sign-off (the theme-bundle schema must be documented in JSON Schema + every supported design token must be enumerated).

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default                          | Range / type                        | Purpose                                                                |
|------------------------------------|----------------------------------|-------------------------------------|------------------------------------------------------------------------|
| `HELIX_TENANT_BUNDLE_STORE`        | `s3://helix-tenant-bundles`      | `s3://...` / `file://...` / `mem://`| Bundle storage backend.                                                |
| `HELIX_TENANT_BUNDLE_FORMAT`       | `yaml`                           | `yaml` / `toml`                     | Default bundle format on write; both supported on read.                |
| `HELIX_TENANT_CACHE_TTL_S`         | `300`                            | int [10, 3600]                      | In-process tenant-cache TTL.                                            |
| `HELIX_TENANT_CACHE_MAX_ENTRIES`   | `10000`                          | int [100, 100000]                   | LRU upper bound; tenant eviction beyond this point.                   |
| `HELIX_TENANT_HOT_RELOAD_ENABLED`  | `true`                           | bool                                | Watch S3 / filesystem for bundle changes; reload affected tenants.    |
| `HELIX_TENANT_FALLBACK_LOCALE`     | `en`                             | BCP 47 tag                          | Locale used if a tenant's `Locales` list is empty.                    |
| `HELIX_TENANT_DEFAULT_THEME`       | `default`                        | tenant ID                           | Theme used if a tenant references an unresolved theme.                |
| `HELIX_TENANT_VALIDATE_AT_LOAD`    | `true`                           | bool                                | Run JSON Schema validation on every bundle load (off only for dev).   |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `LoadByID()` (cache hit)              | 5 µs     | 20 µs   | 50 µs   | LRU lookup; no I/O.                                              |
| `LoadByID()` (cache miss, S3)         | 30 ms    | 80 ms   | 200 ms  | One S3 GetObject + parse + validate.                             |
| `LoadByDomain()` (cache hit)          | 8 µs     | 25 µs   | 60 µs   | Domain-to-ID lookup + LRU.                                       |
| `Theme.RenderHTML()`                  | 80 µs    | 250 µs  | 600 µs  | text/template render of typical page; depends on data complexity. |
| `ValidateFocusTree`-style schema check| 200 µs   | 500 µs  | 1 ms    | At bundle-load time; not in the hot path.                        |
| Memory per cached tenant              | 8 KiB    | 16 KiB  | 32 KiB  | Theme tokens + feature flags + locale set.                       |

### 9.3 Common errors and remediation

| Error                                     | Cause                                                  | Remediation                                                                 |
|-------------------------------------------|--------------------------------------------------------|-----------------------------------------------------------------------------|
| `tenant: ErrNotFound`                     | TenantID or domain not registered                       | Verify the spelling; check the tenant-registry admin tool.                |
| `tenant: ErrSchemaViolation`              | Bundle violates the Theme JSON Schema                  | Reformat the bundle; the error includes the schema path that failed.       |
| `tenant: ErrTemplateExec`                 | Template render failed (e.g. nil data field)            | Add the missing field to data; check the template's expected shape.        |
| `tenant: ErrCacheStampede`                | Multiple concurrent loads on the same cold tenant      | The submodule's per-tenant single-flight singleton dedup avoids this; an error here indicates a bug. |
| `tenant: ErrLocaleUnsupported`            | Tenant's Locales list contains a tag with no fallback  | Add a fallback in `golang.org/x/text/language.Matcher`; default to `en`.   |

### 9.4 Migration from inline theming

A consumer migrating from inline theme constants to `helix-tenant`:

1. Inventory existing theme constants; group by purpose (colours, typography, logo, layout).
2. Express the constants as a YAML bundle conforming to the `Theme` JSON Schema (`docs/theme-schema.json`).
3. Upload the bundle to the configured bundle store under a new tenant ID.
4. Replace inline constant references with `tenant.Theme.DesignTokens.<token>` calls.
5. For server-rendered surfaces: replace `ParseTemplate` calls with `tenant.Theme.RenderHTML`.
6. For client-side surfaces: forward `tenant.Theme` over gRPC to the client and re-apply via `ApplyToClient`.
7. Remove the inline constants once the migration completes; the SPDX licence audit (S01 §4.8.7) flags any dangling theme constants outside the submodule.

The migration is documented in `docs/migration-from-inline-theming.md` in the submodule's repo.

### 9.5 Observability metrics catalog

Per Constitution §10 *Observability*, the submodule emits the following Prometheus metrics. Every metric carries the `tenant_id` and `submodule="helix-tenant"` labels at minimum.

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_tenant_load_total`                       | counter    | Total tenant loads, labelled `result={hit, miss, error}`.                   |
| `helix_tenant_load_latency_seconds`             | histogram  | Load latency distribution; buckets at 100 µs, 1 ms, 10 ms, 100 ms, 1 s.     |
| `helix_tenant_cache_size`                       | gauge      | Current LRU cache occupancy (entries).                                      |
| `helix_tenant_cache_evictions_total`            | counter    | Cumulative cache evictions.                                                 |
| `helix_tenant_render_total`                     | counter    | Template renders, labelled `template_name`, `result={ok, error}`.            |
| `helix_tenant_render_latency_seconds`           | histogram  | Render latency distribution.                                                 |
| `helix_tenant_schema_violations_total`          | counter    | Bundle JSON-Schema violations on load.                                       |
| `helix_tenant_hot_reload_events_total`          | counter    | Hot-reload events, labelled `source={s3, file, manual}`.                    |
| `helix_tenant_locale_fallback_total`            | counter    | Locale-resolution fallbacks; high values indicate i18n-bundle gaps.         |

Recommended Grafana dashboard panels: cache hit-rate (derived from `*_load_total{result="hit"} / *_load_total`), p99 render latency (from the histogram), schema-violation rate (alert threshold: > 0.1 % of loads).

### 9.6 Consumer matrix (chapters that import this submodule)

`helix-tenant` is consumed by the following chapters' implementation surfaces. Each entry resolves the consumer's specific integration to a chapter:section anchor.

| Consumer chapter:section                                                                   | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C06 §6](../../03_Architecture/05_RealTime_APIs.md) — Real-Time APIs                        | Server-side per-tenant rate limiting via `tenant.FeatureFlags`.                      |
| [C07 §6](../../03_Architecture/06_Catalog_and_Assets.md) — Catalog                         | Per-tenant catalog scoping via `tenant.CatalogScope`.                                |
| [C08 §6](../../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) — Host Agent           | Tenant-scoped session metadata in the host-agent's gRPC handlers.                    |
| [C10 §6](../../03_Architecture/09_Security_and_Isolation.md) — Security                    | Tenant-scoped Vault path prefix (`tenant/<tenantID>/`).                              |
| [C11 §6](../../03_Architecture/10_WhiteLabel_and_Theming.md) — White-Label (origin)        | Origin chapter; full theming engine API surface.                                     |
| [C12 §6](../../03_Architecture/11_TV_UX.md) — TV UX                                        | Compose-for-TV theme application via `Theme.ApplyToClient`.                          |

Future chapters (T01..T02 testing, O01..O02 operations, P00..P13 implementation phases) will add additional integration rows; the matrix is updated when those chapters land.

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-tenant-A         | Bundle storage backend — S3-compatible only or also local-filesystem for air-gapped operators?                | C11 §6 next revision                                |
| OQ-tenant-B         | Theme-bundle schema versioning — semver in the bundle itself, or in the loader configuration?                  | C11 §6 next revision                                |
| OQ-tenant-C         | Hot-reload trigger — file-watch (inotify) only, or also S3 event-notification?                                 | `08_Operations/04_Observability_and_Events.md`     |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../03_Architecture/10_WhiteLabel_and_Theming.md`](../../03_Architecture/10_WhiteLabel_and_Theming.md) §6 | (slice) | 2026-04-30 | origin chapter; theming engine API + design tokens |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7 | 1,218 | 2026-04-30 | catalog row #05, R-18 abstention                |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target (this submodule)                                                              |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (parser + template + locale-fallback paths).                       |
| Integration    | All bundle-store backends (S3-compat MinIO, file://, mem://) covered by at least one test.   |
| E2E            | Full client-startup → tenant-load → render path tested end-to-end with two distinct tenants. |
| Security       | govulncheck + Snyk + custom SSTI fuzzer; zero high findings on every run.                    |
| Benchmarking   | p999 ≤ §9.2 budget; cache hit-rate ≥ 95 % under steady-state load.                           |
| Chaos          | S3 partition + clock skew injection; verify graceful fallback.                               |
| Stress         | 24-hour run at 100 loads/s; zero memory growth; zero fd leak.                                |
| Smoke          | 30-second post-deploy verification with known tenant SHA-256.                                |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `04_white_label_multi_tenant/01_tenant_theme_swap_at_session_boundary` baseline-parity.       |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-tenant.md` — 2026-04-30.
