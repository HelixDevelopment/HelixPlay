# `helix-vault` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-vault`                                                                                                          |
| **Origin chapter:section**  | [C10 §6](../../03_Architecture/09_Security_and_Isolation.md) — *Vault wrapper for KEK/DEK + GDPR right-to-erasure*    |
| **Public path (4 mirrors)** | `vasic-digital/helix-vault` on GitHub + GitLab + GitFlic + GitVerse                                                    |
| **Direct deps (vasic-digital)** | **(none — wraps HashiCorp Vault library, no subprocess)**                                                          |
| **External Go deps**        | `github.com/hashicorp/vault/api` (HashiCorp Vault Go client), `github.com/hashicorp/vault-plugin-secrets-kv/v2`         |
| **Licence (S01 §4.8)**      | **Apache-2.0** (one of four named exceptions to the MIT default — cryptographic key-handling submodule, patent-grant matters) |
| **Container CI lane (S02 §3)** | `vault-wrapper-1.x` — builder `golang-builder`, runtime `distroless-static`                                          |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/13_audit_compliance_eu_dsa/scenarios/02_kek_rotation_under_active_session.scenario.yaml`                |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **0** (no `vasic-digital` deps; one of three depth-0 submodules)                                                  |
| **R-18 SafeExec abstention**| **Yes — does NOT import `helix-r18-safeexec`.** Rationale below in §3.1.                                                |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-vault` is the **HashiCorp Vault wrapper** for HelixPlay's secret-management surface. It exposes:

- **KEK / DEK envelope encryption** — Key-Encryption-Key + Data-Encryption-Key two-tier model where the KEK never leaves the Vault server and the DEK is rotated per-tenant per-session.
- **GDPR right-to-erasure** — per [C10 §6](../../03_Architecture/09_Security_and_Isolation.md), the wrapper exposes a `EraseTenant(ctx, tenantID)` API that securely shreds the tenant's DEK, rendering the encrypted data unrecoverable. This is the operationalisation of GDPR Article 17.
- **Per-tenant secrets storage** — OAuth tokens, API keys, signing keys, all stored in tenant-scoped Vault paths.
- **Annual KEK rotation** — automated rotation per Constitution §11.5 *Security & Privacy* with the rotation cadence enforced by HelixQA's deployment gate.

The submodule was introduced in C10 §6 as the consolidation of secret-management logic that earlier projects (HelixAgent, Catalogizer) re-implemented locally. R-04 mandates reuse over duplication.

`★ Why Apache-2.0 specifically.` Per [S01 §4.8.2](../01_Submodule_Catalog.md#482-why-apache-20-specifically-for-those-four), `helix-vault` carries the Apache-2.0 licence (not the MIT default) because it implements cryptographic key-handling primitives (AES-GCM, AES-NI selectors, BoringSSL-derived constructions) that may practice patentable parallel-algorithm techniques. The Apache-2.0 §3 patent grant is therefore operationally significant — a downstream consumer integrating `helix-vault` into their own product needs the patent-grant comfort to do so without an additional licence negotiation. The MIT licence does not grant patents and would leave a downstream consumer to navigate the underlying patent landscape separately.

---

## 2. Public API Surface

### 2.1 The `Vault` client

```go
package vault

// Vault is the HelixPlay-specific wrapper around the HashiCorp
// Vault Go client. It enforces tenant-scoped paths, KEK/DEK
// envelope encryption, and the GDPR right-to-erasure API.
type Vault struct { /* ... */ }

// New creates a Vault client bound to the given Vault server
// address and AppRole-issued token (per C10 §6's auth model).
func New(addr, roleID, secretID string) (*Vault, error)

// Encrypt returns ciphertext encrypted with a freshly-rotated
// per-tenant DEK. The DEK itself is wrapped under the KEK and
// stored in Vault; the returned ciphertext can be decrypted only
// by a Vault that has the matching KEK loaded.
func (v *Vault) Encrypt(ctx context.Context, tenantID string, plaintext []byte) (ciphertext []byte, err error)

// Decrypt is the inverse of Encrypt. Returns ErrTenantErased if
// the tenant's DEK has been shredded by EraseTenant.
func (v *Vault) Decrypt(ctx context.Context, tenantID string, ciphertext []byte) (plaintext []byte, err error)

// EraseTenant shreds the tenant's DEK, rendering all ciphertext
// encrypted under it unrecoverable. This is the GDPR Article 17
// "right to erasure" implementation.
func (v *Vault) EraseTenant(ctx context.Context, tenantID string) error
```

### 2.2 Secret storage helpers

```go
package vault

// PutSecret stores a secret under a tenant-scoped path. The path
// is automatically prefixed with tenant/<tenantID>/.
func (v *Vault) PutSecret(ctx context.Context, tenantID, path string, value []byte) error

// GetSecret retrieves a secret. Returns ErrNotFound for missing
// keys, ErrTenantErased for erased tenants.
func (v *Vault) GetSecret(ctx context.Context, tenantID, path string) ([]byte, error)
```

### 2.3 The KEK rotation API

```go
package vault

// RotateKEK rotates the Key-Encryption-Key. New DEKs are wrapped
// under the new KEK; existing DEKs are re-wrapped lazily on next
// access (the KEK never decrypts plaintext directly, only DEKs).
//
// Rotation cadence: annually per Constitution §11.5; enforced by
// HelixQA's deployment gate (S04 §9).
func (v *Vault) RotateKEK(ctx context.Context) error
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital` — none. Why no `helix-r18-safeexec`?

`helix-vault` does **not** invoke subprocesses. Its only system-side interactions are HTTP calls to the HashiCorp Vault server via the `github.com/hashicorp/vault/api` library. Subprocess invocations are out of the submodule's surface; consequently it does not need the R-18 enforcement primitive that `helix-r18-safeexec` provides.

The abstention is documented in [S01 §7.1](../01_Submodule_Catalog.md#71-the-ladder)'s "ABSTAIN (no SafeExec — they wrap libraries, not subprocesses)" group. Per [S01 §3.3](../01_Submodule_Catalog.md#33-the-none-dependency-rows), `helix-vault` is one of three submodules with no `vasic-digital` dependencies (the others being `helix-r18-safeexec` itself and `helix-tenant`).

The abstention is not a relaxation of R-18; it is a reflection of the submodule's actual surface. Constitution §11.5.4 (R-18) regulates **subprocess invocations**; a submodule that invokes no subprocesses is vacuously compliant. The host-integrity-scan harness (S02 §3.4) verifies this empirically — an attempted forbidden invocation in a `helix-vault` test would fail because the submodule has no `os/exec` import in its source tree (per the layer-4 ripgrep CI lint, S01 §7.3).

### 3.2 External (Go)

- `github.com/hashicorp/vault/api` — official HashiCorp Vault Go client.
- `github.com/hashicorp/vault-plugin-secrets-kv/v2` — KV v2 secret-engine support.
- `crypto/aes`, `crypto/cipher`, `crypto/rand` — Go standard library cryptography.

### 3.3 External (system)

A running HashiCorp Vault server reachable via HTTPS. Production deployments use Vault Enterprise's Auto-Unseal feature (per Constitution §11.5); MVP development uses `vault server -dev` in a local container per S02 §3.

---

## 4. Container Build (S02 §3 lane: `vault-wrapper-1.x`)

**Builder:** `golang-builder`. **Runtime:** `distroless-static`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** `--user 65532:65532`, `--read-only`, seccomp default-deny, `--cap-drop=ALL`, `--memory=512m`, `--cpus=0.5`.

The container does **not** carry any cryptographic material at rest — keys live in Vault, not in the container image. The image is fully reproducible across the four mirrors.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
- KEK / DEK envelope round-trips with mock Vault HTTP transport.
- `EraseTenant` semantics on synthetic tenant inventories.
- Path-prefix enforcement (no escape from `tenant/<tenantID>/`).

### 5.2 Integration
- Real Vault server (containerised dev-mode); KEK rotation; DEK lazy re-wrap; cross-tenant isolation.

### 5.3 E2E
- HelixPlay host agent → `helix-vault` → real Vault server → encrypt + decrypt round-trip.

### 5.4 Security
- govulncheck + Snyk + Trivy.
- Custom test: attempted cross-tenant access (must fail with `ErrUnauthorized`).
- KEK-rotation-while-decrypting test: must not lose data; must not return stale plaintext.

### 5.5 Benchmarking
- Encrypt p999 ≤ 5 ms (HTTP RTT to Vault dominates); decrypt p999 ≤ 5 ms.

### 5.6 Chaos
- Toxiproxy partition between `helix-vault` and Vault server; verify the wrapper surfaces the partition cleanly.

### 5.7 Stress
- 24-hour run at 1 K encrypt+decrypt operations per second; verify no leak.

### 5.8 Smoke
- 30-second post-deploy: PutSecret + GetSecret round-trip.

### 5.9 Full Automation
- §5.1–§5.8.

### 5.10 Challenges
- `13_audit_compliance_eu_dsa/02_kek_rotation_under_active_session.scenario` — KEK rotation while a session is actively encrypting / decrypting; verify no session interruption, no plaintext leak, OTLP trace shows the rotation event.

---

## 6. Challenges Entry-Point (S03 §4 row #04)

**Topology:** `13_audit_compliance_eu_dsa`. **Scenario:** `02_kek_rotation_under_active_session.scenario.yaml`. **Why this scenario.** KEK rotation under load is the highest-risk operation `helix-vault` performs; the scenario specifically exercises the lazy DEK re-wrap path under active encryption traffic. **Baseline:** zero session interruptions, OTLP trace topology with the rotation event, p999 latency staying within 1.5× the steady-state baseline during the rotation window.

---

## 7. R-18 Inheritance

`helix-vault` **abstains** from R-18 SafeExec inheritance because it invokes no subprocesses (§3.1). The five-layer enforcement model (S01 §7.3) still applies at layers 1 (chapter prose — C10 §6 documents the no-subprocess invariant), 4 (ripgrep CI lint — verifies no `os/exec` import is added in future PRs), and 5 (host-integrity-scan — verifies no forbidden syscall fires during test runs). Layers 2 and 3 are not applicable to this submodule because they specifically govern subprocess invocations.

A future PR that introduces a subprocess invocation would fail the layer-4 lint and would require explicit acknowledgement in the PR description that R-18 SafeExec inheritance is now required (i.e. the submodule must add `helix-r18-safeexec` as a dependency).

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation gated on KEK rotation cadence being green for two consecutive cycles in HelixQA's deployment-gate run-archive. Apache-2.0 licence requires SPDX header on every source file; the per-submodule CI lane's `spdx-check` lint enforces this.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default                  | Range / type             | Purpose                                                                |
|------------------------------------|--------------------------|--------------------------|------------------------------------------------------------------------|
| `HELIX_VAULT_ADDR`                 | `https://vault:8200`     | URL                      | Vault server address; HTTPS only in production.                        |
| `HELIX_VAULT_ROLE_ID`              | (required)               | UUID                     | AppRole role identifier; provisioned by ops.                          |
| `HELIX_VAULT_SECRET_ID_FILE`       | `/run/secrets/vault-sid` | filesystem path          | AppRole secret-id; expected as a Docker secret mount.                 |
| `HELIX_VAULT_TENANT_PATH_PREFIX`   | `tenant/`                | string                   | Tenant-scoped path prefix; never overrideable per-tenant.             |
| `HELIX_VAULT_KEK_ROTATION_DAYS`    | `365`                    | int [30, 730]            | KEK rotation cadence; default annual per Constitution §11.5.          |
| `HELIX_VAULT_DEK_CACHE_TTL_S`      | `300`                    | int [60, 3600]           | In-process DEK cache TTL; balances Vault load against latency.        |
| `HELIX_VAULT_TLS_CA_FILE`          | `/etc/helix/tls/ca.pem`  | filesystem path          | CA bundle for verifying the Vault server's cert.                       |
| `HELIX_VAULT_AUDIT_LOG_DEST`       | `stdout`                 | `stdout` / `syslog` / URL| Audit-log destination; production uses syslog + SIEM forwarding.      |

### 9.2 Performance budget

| Metric                                | p50     | p99     | p999    | Notes                                                                |
|---------------------------------------|---------|---------|---------|----------------------------------------------------------------------|
| `Encrypt()` (warm DEK cache)          | 200 µs  | 500 µs  | 1 ms    | AES-GCM in-process; no HTTP round-trip.                              |
| `Encrypt()` (cold DEK cache)          | 4 ms    | 8 ms    | 15 ms   | One HTTP RTT to Vault to fetch + unwrap DEK.                         |
| `Decrypt()` (warm)                    | 200 µs  | 500 µs  | 1 ms    | Same as Encrypt warm-path.                                            |
| `EraseTenant()`                       | 50 ms   | 200 ms  | 500 ms  | Multiple Vault writes (DEK shred + audit-log emit).                  |
| `RotateKEK()` (full fleet)            | 30 s    | 120 s   | 300 s   | Lazy DEK re-wrap; running sessions tolerate the rotation window.    |
| `PutSecret()` / `GetSecret()`         | 4 ms    | 8 ms    | 15 ms   | One HTTP RTT to Vault.                                                |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                 |
|-----------------------------------------------|----------------------------------------------------------------|-----------------------------------------------------------------------------|
| `vault: ErrTenantErased`                      | Caller tried to access a tenant that has been GDPR-erased     | Surface to the user as "data has been deleted"; do **not** attempt fallback recovery. |
| `vault: ErrUnauthorized`                       | AppRole secret-id expired or rotated                          | Re-provision via Vault admin; verify clock skew < 60 s.                    |
| `vault: ErrVaultSealed`                       | Vault is sealed (post-restart or after a manual seal command) | Operator unseals via Shamir shares (OSS) or auto-unseal (Enterprise).      |
| `vault: ErrConnectionRefused`                 | Network partition between client and Vault                     | Investigate via Toxiproxy / network dashboards; client retries with exponential backoff. |
| `vault: ErrKEKVersionMismatch`                | DEK was wrapped under a KEK version no longer in Vault        | Re-encrypt the affected ciphertext; emit P2 alert (data may be stale).    |

### 9.4 Migration from inline crypto

A consumer migrating from inline AES-GCM (with locally-stored keys) to `helix-vault`:

1. Inventory existing key material; document each key's purpose and tenant scope.
2. Provision a Vault server (or an existing operator's Vault) and configure the AppRole auth method.
3. Migrate keys to Vault under the `tenant/<tenantID>/keys/` path; the migration tool `helixctl vault-migrate` automates this with operator approval per key.
4. Replace inline `crypto/aes` calls with `helix-vault.Encrypt` / `helix-vault.Decrypt`.
5. Remove the local key material (per Constitution §11.5 — "no key material at rest in containers").
6. Wire `EraseTenant` into the GDPR right-to-erasure handler.
7. Schedule the first KEK rotation; HelixQA's deployment gate enforces annual cadence thereafter.

The migration is documented in `docs/migration-from-inline-crypto.md` in the submodule's repo.

### 9.5 Observability metrics catalog

Per Constitution §10 and §11.5 *Security & Privacy*, the submodule emits the following Prometheus metrics. Every metric carries the `tenant_id` and `submodule="helix-vault"` labels at minimum. **No metric carries plaintext payloads or KEK/DEK material** — that would defeat the security posture this submodule exists to enforce.

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_vault_encrypt_total`                     | counter    | Encrypt invocations, labelled `result={ok, err}`, `cache={hit, miss}`.       |
| `helix_vault_decrypt_total`                     | counter    | Decrypt invocations, same labels.                                            |
| `helix_vault_encrypt_latency_seconds`           | histogram  | Encrypt latency; buckets at 200 µs, 1 ms, 5 ms, 20 ms.                       |
| `helix_vault_decrypt_latency_seconds`           | histogram  | Decrypt latency; same bucket shape.                                          |
| `helix_vault_dek_cache_size`                    | gauge      | Current DEK cache occupancy.                                                 |
| `helix_vault_dek_cache_evictions_total`         | counter    | Cumulative DEK cache evictions.                                              |
| `helix_vault_kek_rotation_events_total`         | counter    | KEK rotation events, labelled `phase={start, complete, error}`.              |
| `helix_vault_erase_tenant_events_total`         | counter    | GDPR erasure events.                                                         |
| `helix_vault_secret_op_total`                   | counter    | PutSecret / GetSecret invocations, labelled `op`, `result`.                  |
| `helix_vault_auth_failures_total`               | counter    | AppRole auth failures (label `reason={expired, refused, network}`); P1 if rate spikes. |

Recommended Grafana dashboard panels: cache hit-rate, p999 encrypt latency, KEK rotation lag (time since last rotation; alert threshold ≥ 1 year per `HELIX_VAULT_KEK_ROTATION_DAYS`), auth failure rate (alert: any non-zero rate over 1-min window is suspicious).

### 9.6 Consumer matrix (chapters that import this submodule)

`helix-vault` is consumed by the following chapters' security + privacy integration surfaces. The submodule is the canonical landing for any KEK/DEK or tenant-secret operation.

| Consumer chapter:section                                                                   | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C09 §6](../../03_Architecture/08_Scalability_and_MultiRegion.md) — Scalability            | Cross-region KEK consistency via Vault Enterprise's replication.                     |
| [C10 §6](../../03_Architecture/09_Security_and_Isolation.md) — Security (origin)           | Origin chapter; full Vault wrapper API + GDPR Article 17 implementation.             |
| [C08 §6](../../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) — Host Agent           | Per-session DEK acquisition for in-flight payload encryption.                        |
| [C11 §6](../../03_Architecture/10_WhiteLabel_and_Theming.md) — White-Label                 | Per-tenant secret storage (OAuth tokens, signing keys, tenant-private API keys).     |

Future chapters (T01..T02 testing — security test row, O01..O02 operations — Vault deployment topology, P00..P13 implementation phases) will add additional integration rows; the matrix is updated when those chapters land.

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-vault-A          | Vault Enterprise Auto-Unseal vs OSS Shamir-share — operator deployment matrix?                                 | `08_Operations/01_Container_CI_CD.md`              |
| OQ-vault-B          | KEK rotation cadence — annual fixed or operator-configurable?                                                  | C10 §6 next revision                                |
| OQ-vault-C          | GDPR Article 17 audit trail — what record format for the erasure event?                                        | C10 §6 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../03_Architecture/09_Security_and_Isolation.md`](../../03_Architecture/09_Security_and_Isolation.md) §6 | (slice) | 2026-04-30 | origin chapter; KEK/DEK + GDPR              |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §4.8 §7 | 1,218 | 2026-04-30 | catalog row #04, Apache-2.0 rationale, R-18 abstention |
| [`../../01_Constitution.md`](../../01_Constitution.md) §11.5      |    879 | 2026-04-30 | R-18 + Security & Privacy normative wording     |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target (this submodule)                                                              |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (encrypt/decrypt round-trips + cross-tenant isolation paths).       |
| Integration    | Real containerised Vault dev-server; KEK rotation; lazy DEK re-wrap; tenant-isolation tests. |
| E2E            | Host agent → helix-vault → real Vault encrypt+decrypt round-trip with realistic payloads.    |
| Security       | govulncheck + Snyk + Trivy + KEK-rotation-while-decrypting + cross-tenant access denial.     |
| Benchmarking   | Encrypt/Decrypt p999 ≤ §9.2 budget; rotation full-fleet ≤ 5 min p999.                        |
| Chaos          | Toxiproxy partition between client and Vault; verify clean error surfacing + retry logic.   |
| Stress         | 24-hour run at 1 K ops/s; zero memory leak; zero auth-token leak.                            |
| Smoke          | 30-second post-deploy: PutSecret + GetSecret round-trip verifies Vault connectivity.        |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `13_audit_compliance_eu_dsa/02_kek_rotation_under_active_session` baseline-parity.            |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-vault.md` — 2026-04-30.
