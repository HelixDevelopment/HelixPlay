# Phase_10 — Monetization & Auth

> **Source dimensions:** [`Phase_09_Recording_and_Replay.md`](Phase_09_Recording_and_Replay.md), [`../06_Submodules/per-submodule/helix-vault.md`](../06_Submodules/per-submodule/helix-vault.md), [`../06_Submodules/per-submodule/helix-tenant.md`](../06_Submodules/per-submodule/helix-tenant.md), [`../06_Submodules/per-submodule/helix-billing.md`](../06_Submodules/per-submodule/helix-billing.md), [`../03_Architecture/09_Tenant_and_Auth.md`](../03_Architecture/09_Tenant_and_Auth.md) (C10), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P10.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P10.
> **Phase targets:** R-13 anti-bluff billing audits + GDPR (Constitution §11) + per-tenant key isolation (Constitution §11.5).
> **Cross-links:** [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md), [`Phase_09_Recording_and_Replay.md`](Phase_09_Recording_and_Replay.md), [`Phase_06_Host_Agent.md`](Phase_06_Host_Agent.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_10 ships the **commercial layer** that turns a working operator-deployed HelixPlay stack into a billable + revenue-bearing product:

- **Authentication:** OAuth 2.1 + OIDC via the operator-pinned IdP (Keycloak default; ADFS/Okta/Azure AD pluggable). MFA enforced for all operator + admin tiers.
- **Authorization:** RBAC + per-tenant policy engine via helix-tenant. Per-game launcher entitlement check.
- **Billing:** Per-seat + per-minute + per-bandwidth SKU options. Stripe + Adyen + per-region jurisdictional gateway abstraction.
- **Tenant lifecycle:** Provisioning + decommissioning + GDPR erasure-on-demand + cross-tenant migration.
- **Per-seat licensing:** Concurrent-session counting + revocation + grace periods + offline tokens for operator-LAN-isolated sessions.
- **Audit trail:** Every billing-relevant event lands in helix-billing's tamper-evident audit log (cosign-signed per-event), per Constitution §11.5 R-13.
- **Anti-fraud:** Per-IP / per-device / per-payment-method velocity + spike detection. Operator-tunable rate limits.

After Phase_10, an operator can **charge real money** for HelixPlay sessions + survive an external audit on the resulting revenue events. The operator's CFO + compliance officer have the dashboards + audit log + reconciliation reports they need.

---

## 2. Prerequisites

- Phase_09 complete + signed off — recording + replay operational so per-session billing has the source-of-truth artifact.
- helix-vault + helix-tenant + helix-billing at v1.0.0.
- Operator-provisioned commercial agreements: payment-gateway merchant account, IdP tenant, jurisdictional tax registration (per [C10 §11](../03_Architecture/09_Tenant_and_Auth.md)).
- helix-r18-safeexec wraps every billing-disrupting subprocess (per Constitution §11.5).

---

## 3. Tasks Catalogue

| Task ID    | Task                                                                | Subtasks |
|------------|---------------------------------------------------------------------|---------:|
| P10.T01   | OAuth 2.1 + OIDC integration (Keycloak default + pluggable adapter) | 5        |
| P10.T02   | MFA enforcement on operator + admin tiers                            | 4        |
| P10.T03   | RBAC + per-tenant policy engine — helix-tenant.PolicyEngine wired   | 5        |
| P10.T04   | helix-billing — Stripe + Adyen + per-region adapters                 | 6        |
| P10.T05   | Per-seat licensing + concurrent-session counting + grace periods     | 5        |
| P10.T06   | Per-minute + per-bandwidth metering — Prometheus → helix-billing      | 4        |
| P10.T07   | Tenant provisioning + decommissioning workflows                      | 5        |
| P10.T08   | GDPR erasure-on-demand SLA (30 days) — verified end-to-end           | 5        |
| P10.T09   | Cross-tenant migration — preserves recording history + entitlements  | 4        |
| P10.T10   | Anti-fraud velocity + spike detection — operator-tunable rules       | 4        |
| P10.T11   | Audit log (cosign-signed per-event tamper-evident chain)              | 4        |
| P10.T12   | Reconciliation reports — daily + monthly + tax-period boundaries     | 4        |
| P10.T13   | Per-game launcher entitlement check — Steam + GOG + Epic             | 4        |
| P10.T14   | Refund + chargeback + dispute workflows                              | 4        |
| P10.T15   | Operator runbook — billing operations + incident response             | 3        |
| P10.T16   | End-to-end smoke — provision tenant → charge → refund → erase       | 5        |
| P10.T17   | Phase_10 acceptance review                                            | 2        |

17 tasks; ~73 subtasks.

---

## 4. Task Details

### 4.1 P10.T01 — OAuth 2.1 + OIDC

helix-tenant.AuthService configured for OAuth 2.1 + OIDC per [C10 §6.1](../03_Architecture/09_Tenant_and_Auth.md). Keycloak is the canonical default; the AuthService.Provider interface accepts ADFS / Okta / Azure AD / Google Workspace adapters. The operator pins one provider per tenant via the tenant's bootstrap config.

**Subtasks:**

- T01.S01 — Wire Keycloak realm bootstrap (per-tenant realm with operator-managed admin credentials).
- T01.S02 — OIDC discovery endpoint exposed at `https://auth.helix.<operator-domain>/realms/<tenant>/.well-known/openid-configuration`.
- T01.S03 — PKCE-required for public clients (Phase_05 Wails + Compose-for-TV + Steam Deck).
- T01.S04 — Refresh token rotation per OAuth 2.1 (refresh tokens are single-use; replay attempts revoke the chain).
- T01.S05 — Provider-adapter conformance Challenges scenarios per [helix-tenant §6](../06_Submodules/per-submodule/helix-tenant.md).

### 4.2 P10.T02 — MFA enforcement

Constitution §11 mandates MFA for all operator + admin tier access. helix-tenant.MFAService supports TOTP (canonical), WebAuthn (FIDO2 + passkeys), and SMS as last-resort fallback. The operator can disable SMS per tenant via the tenant's `allow_sms_mfa: false` config.

**Subtasks:**

- T02.S01 — TOTP enrollment flow (QR code + recovery codes — recovery codes are one-time-use + tracked in audit log).
- T02.S02 — WebAuthn enrollment + authentication flow (passkey + per-device hardware token).
- T02.S03 — SMS fallback (operator-disable-able; warning emitted per regulatory framework — Russia, Brazil, India default-off because of carrier-billing fraud history).
- T02.S04 — MFA bypass audit trail — every bypass is logged + sent to operator's SIEM (Constitution §11.5 R-13).

### 4.3 P10.T03 — RBAC + per-tenant policy engine

helix-tenant.PolicyEngine evaluates per-request authorization decisions per [C10 §6.2](../03_Architecture/09_Tenant_and_Auth.md). Five canonical roles: `tenant.admin`, `tenant.billing`, `tenant.operator`, `tenant.user`, `tenant.guest`. Per-tenant custom roles supported via Rego policies; the operator's compliance officer signs off on custom policies before they go live.

**Subtasks:**

- T03.S01 — Five canonical roles defined + RBAC matrix exported to operator's compliance officer.
- T03.S02 — Custom-role Rego policy evaluator wired (OPA — Open Policy Agent).
- T03.S03 — Per-request audit log entry (per Constitution §11.5).
- T03.S04 — Cross-tenant access denial — verifiable smoke probe (a tenant.admin in tenant A cannot access tenant B's resources).
- T03.S05 — Time-of-day + IP-range + device-attestation policy primitives.

### 4.4 P10.T04 — helix-billing payment gateway adapters

helix-billing's PaymentGateway interface accepts adapter implementations for jurisdictional gateways:

| Region   | Default Gateway | Fallback Gateway | Currency |
|----------|-----------------|------------------|----------|
| EU       | Adyen           | Stripe           | EUR      |
| North America | Stripe     | Adyen            | USD/CAD  |
| Russia   | YooKassa        | (none)           | RUB      |
| Brazil   | Adyen (Brazil)  | Stripe Brazil    | BRL      |
| India    | Razorpay        | (none)           | INR      |
| China    | Alipay/WeChat Pay | (none)         | CNY      |

**Subtasks:**

- T04.S01 — Stripe adapter (Stripe Treasury + Stripe Connect for marketplace operators).
- T04.S02 — Adyen adapter (Adyen Marketplace + Adyen Dispute Manager).
- T04.S03 — YooKassa adapter (Russia-jurisdiction).
- T04.S04 — Razorpay adapter (India-jurisdiction).
- T04.S05 — Alipay + WeChat Pay adapter (China-jurisdiction).
- T04.S06 — Adapter conformance Challenges scenarios per [helix-billing §6](../06_Submodules/per-submodule/helix-billing.md).

### 4.5 P10.T05 — Per-seat licensing

helix-tenant.LicenseService enforces concurrent-session limits. Three licensing modes:

- **Per-seat:** Operator buys N seats; up to N concurrent sessions allowed; (N+1)-th session is queued or rejected per the tenant's `over_seat_policy`.
- **Per-minute:** Sessions billed by elapsed wall-clock minute; minute boundaries flushed to helix-billing every 60 s.
- **Per-bandwidth:** Sessions billed by aggregate egress bandwidth (per RFC-3917 IPFIX records).

**Subtasks:**

- T05.S01 — Concurrent-session counter — Redis-backed, per-tenant, atomic increment/decrement.
- T05.S02 — Over-seat policy primitives (`reject`, `queue`, `auto_upgrade`).
- T05.S03 — Grace periods (5-minute reconnect window for transient network failure — session counter doesn't decrement).
- T05.S04 — Offline tokens for operator-LAN-isolated deployments (per [helix-tenant §9.4](../06_Submodules/per-submodule/helix-tenant.md)).
- T05.S05 — Revocation — operator can revoke a session in-flight; client receives a graceful disconnect + post-disconnect billing reconciliation.

### 4.6 P10.T06 — Per-minute + per-bandwidth metering

Prometheus scrapes per-session metrics from helix-pipeline + helix-transport; helix-billing's MeterIngester aggregates per-minute + per-bandwidth deltas into the per-tenant billing event stream.

**Subtasks:**

- T06.S01 — Prometheus session counters + bandwidth histograms.
- T06.S02 — MeterIngester batch-aggregates per minute + per hour.
- T06.S03 — Late-arriving metrics handling (tolerates ≤ 5-minute clock skew; beyond that, escalates to operator).
- T06.S04 — Per-tenant billing event emission (Kafka topic `helix.billing.events.<tenant>`).

### 4.7 P10.T07 — Tenant provisioning + decommissioning

helix-tenant.LifecycleService — automated workflows for tenant creation + decommissioning. Per [C10 §6.3](../03_Architecture/09_Tenant_and_Auth.md).

**Subtasks:**

- T07.S01 — Tenant creation: bootstrap Keycloak realm + Vault namespace + CockroachDB schema + helix-billing customer record.
- T07.S02 — Tenant decommissioning: graceful 30-day grace period + final-invoice generation + GDPR erasure trigger.
- T07.S03 — Tenant configuration export/import — JSON Schema-validated.
- T07.S04 — Tenant cloning — operator can clone a tenant's config to bootstrap a new one (with ID/secret rotation).
- T07.S05 — Per-tenant resource quotas (CPU + GPU + bandwidth + storage) — enforced via cgroup + Cilium NetworkPolicy.

### 4.8 P10.T08 — GDPR erasure-on-demand

Per Constitution §11.5 + helix-vault.EraseTenant. SLA: ≤ 30 days from request to verified erasure (per GDPR Article 17). Erasure scope: tenant's recordings (DEK shred), tenant's billing artifacts (cryptographic deletion via encryption-at-rest), tenant's auth credentials (Keycloak realm deletion), tenant's metrics + logs (helix-otel-init retention policy).

**Subtasks:**

- T08.S01 — Erasure request workflow (tenant.admin or external regulator request; operator-mediated).
- T08.S02 — Erasure execution — atomic across all sub-systems + rollback-protection.
- T08.S03 — Erasure audit certificate — cosign-signed; given to the requesting party as legal proof of compliance.
- T08.S04 — Per-tenant DEK shred (renders ciphertext unrecoverable).
- T08.S05 — Erasure SLA monitoring — Prometheus alert if any pending erasure exceeds 25 days.

### 4.9 P10.T09 — Cross-tenant migration

A tenant may migrate from operator A to operator B (e.g., M&A scenario). helix-tenant.MigrationService preserves recording history + entitlements + billing audit log.

**Subtasks:**

- T09.S01 — Migration export bundle (recordings + entitlements + audit log + tenant config).
- T09.S02 — Cosign-signed migration bundle — destination operator verifies signature before import.
- T09.S03 — Migration import (tenant arrives at destination operator with full history).
- T09.S04 — Per-region jurisdictional checks (e.g., EU tenant cannot migrate to a US operator without GDPR DPA in place — operator-mediated).

### 4.10 P10.T10 — Anti-fraud velocity + spike detection

helix-billing.FraudDetector evaluates per-IP / per-device / per-payment-method velocity. Operator-tunable rules per [helix-billing §10](../06_Submodules/per-submodule/helix-billing.md).

**Subtasks:**

- T10.S01 — Velocity rules (e.g., "same payment method, ≥ 5 sessions in 1 hour, different IPs ≥ 3" → flag + manual review).
- T10.S02 — Spike detection — anomaly detection on per-tenant cost; alert operator if 3σ deviation from rolling baseline.
- T10.S03 — Manual review queue — operator's fraud team triages flagged sessions.
- T10.S04 — Block-list integration — operator-managed IP / device / payment-method block-list with cosign-signed entry.

### 4.11 P10.T11 — Audit log (tamper-evident)

helix-billing.AuditLog — every billing-relevant event is cosign-signed per-event + chained (each event references the prior event's hash). Per Constitution §11.5 R-13.

**Subtasks:**

- T11.S01 — Per-event signing (cosign + private key in helix-vault).
- T11.S02 — Hash-chained event sequence (Merkle root published daily + signed by operator).
- T11.S03 — Verification tooling — auditor's CLI verifies any event's chain-of-custody.
- T11.S04 — Tamper detection — Prometheus alert on chain break.

### 4.12 P10.T12 — Reconciliation reports

helix-billing.ReconciliationService produces daily + monthly + tax-period boundary reports. Per [helix-billing §11](../06_Submodules/per-submodule/helix-billing.md).

**Subtasks:**

- T12.S01 — Daily reconciliation — gateway-side vs. helix-billing-side event counts + amounts + tax.
- T12.S02 — Monthly reconciliation — per-tenant invoice generation.
- T12.S03 — Tax-period boundary — quarterly + annual tax-jurisdiction-aware reports.
- T12.S04 — Discrepancy alert — Prometheus alert if any reconciliation reveals > 0.1% mismatch.

### 4.13 P10.T13 — Per-game launcher entitlement check

helix-launcher (Steam, GOG, Epic, Battle.net) — per [C08 §9](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md). Each launcher exposes a per-user game-ownership API; helix-tenant queries the launcher before letting a session launch a given game.

**Subtasks:**

- T13.S01 — Steam OAuth + ownership-API integration.
- T13.S02 — GOG GalaxyAPI integration.
- T13.S03 — Epic Online Services integration.
- T13.S04 — Battle.net authentication + entitlement.

### 4.14 P10.T14 — Refund + chargeback + dispute workflows

helix-billing.DisputeService — per-gateway dispute APIs + operator-mediated workflow. Disputes go through:

1. Customer-initiated chargeback or refund-request.
2. Operator's billing team triage (per `tenant.billing` role).
3. Evidence packet generation (recording + session metadata + audit log entries).
4. Gateway-side response submission.
5. Resolution + audit log entry.

**Subtasks:**

- T14.S01 — Stripe Dispute Manager integration.
- T14.S02 — Adyen Dispute Manager integration.
- T14.S03 — Evidence packet generation (recording snippet + session metadata + audit log entries).
- T14.S04 — Resolution audit log entry (cosign-signed per-event).

### 4.15 P10.T15 — Operator runbook

`HelixDevelopment/HelixBilling/docs/runbook/billing-operations.md` — covers billing onboarding, dispute response, refund workflow, anti-fraud tuning, GDPR erasure response.

**Subtasks:**

- T15.S01 — Onboarding flow — tenant signup → realm bootstrap → first invoice.
- T15.S02 — Dispute response — chargeback escalation playbook.
- T15.S03 — GDPR erasure response — request-to-certificate workflow.

### 4.16 P10.T16 — End-to-end smoke

Provision a test tenant → charge a USD 1.00 + EUR 1.00 + RUB 100 transaction → trigger a refund → trigger erasure → verify all artifacts deleted. 30-minute end-to-end smoke.

**Subtasks:**

- T16.S01 — Test tenant provisioning (per Stripe + Adyen + YooKassa test-mode credentials).
- T16.S02 — Charge transaction — verify webhook + event chain.
- T16.S03 — Refund transaction — verify reverse event + reconciliation entry.
- T16.S04 — Erasure trigger — verify all artifacts gone within 5 minutes (test tenant; production SLA is 30 days).
- T16.S05 — Audit log integrity — cosign-verify every event.

### 4.17 P10.T17 — Phase_10 acceptance review

Operator + compliance officer + CFO signoff per Constitution §16 + the §6 exit criteria.

---

## 5. Subtask Catalogue

73 subtasks across 17 tasks; bulk-imported.

---

## 6. Exit Criteria

- [ ] OAuth 2.1 + OIDC operational with Keycloak + at least one alternative provider verified.
- [ ] MFA enforced on operator + admin tiers.
- [ ] RBAC + per-tenant policy engine wired.
- [ ] At least 3 payment-gateway adapters operational (Stripe + Adyen + one regional).
- [ ] Per-seat + per-minute + per-bandwidth metering working.
- [ ] Tenant provisioning + decommissioning automated.
- [ ] GDPR erasure-on-demand verified end-to-end.
- [ ] Cross-tenant migration verified.
- [ ] Anti-fraud velocity + spike detection live.
- [ ] Audit log tamper-evident chain verified.
- [ ] Daily + monthly reconciliation passes.
- [ ] At least 1 game launcher entitlement check operational.
- [ ] Refund + chargeback + dispute workflows tested.
- [ ] End-to-end smoke (provision → charge → refund → erase) passes.
- [ ] Operator + compliance officer + CFO signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP10-01  | Payment-gateway downtime breaks per-tenant billing                  | Multi-gateway fallback (Stripe ↔ Adyen) per region.                   |
| RP10-02  | GDPR erasure SLA exceeded (> 30 days)                               | Prometheus alert at 25 days + automated escalation.                   |
| RP10-03  | Audit log tamper detected (chain break)                             | Cosign verification + auto-quarantine of tenant + operator escalation. |
| RP10-04  | OAuth refresh-token replay attack                                   | OAuth 2.1 single-use refresh tokens + chain-revocation on replay.      |
| RP10-05  | MFA bypass audit log tampered                                       | Tamper-evident chain (RP10-03 mitigation applies).                    |
| RP10-06  | Anti-fraud false-positive blocks legitimate operator                | Manual review queue + SLA on triage (≤ 4 hours).                      |
| RP10-07  | Per-game launcher entitlement API rate-limited                      | Per-tenant cache + exponential backoff per [helix-tenant §9.5](../06_Submodules/per-submodule/helix-tenant.md). |
| RP10-08  | Cross-tenant migration loses recording history                      | Cosign-signed bundle + checksum verification at import.               |

---

## 8. Cross-Family Dependencies

- C10 §6 (Tenant & Auth) — architecturally specifies the patterns.
- helix-tenant + helix-billing + helix-vault at v1.0.0 from Phase_02.
- helix-r18-safeexec wraps every billing-disrupting subprocess.
- helix-otel-init for billing-event observability.
- helix-launcher submodules for per-launcher entitlement.

---

## 9. Acceptance Criteria

Constitution §16 signoff (operator + compliance officer + CFO) + §6 exit criteria.

---

## 10. The Phase_10 Calendar

~ 6 weeks. Operator-side capacity: 4 engineers (2× backend Go for helix-billing + helix-tenant; 1× security engineer for OAuth + MFA + audit chain; 1× ops engineer for runbook + reconciliation).

The Phase_10 calendar slots align with the operator's commercial agreement signing — payment gateway merchant accounts, IdP tenant agreements, and jurisdictional tax registration must be in place before the engineering work has anything to integrate against.

---

## 11. Anti-Bluff Verification

### 11.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_09_Recording_and_Replay.md`](Phase_09_Recording_and_Replay.md) | 167+ | 2026-04-30 | predecessor                                      |
| [`../03_Architecture/09_Tenant_and_Auth.md`](../03_Architecture/09_Tenant_and_Auth.md) | 1,800+ | 2026-04-30 | C10 architectural source             |
| [`../06_Submodules/per-submodule/helix-tenant.md`](../06_Submodules/per-submodule/helix-tenant.md) | 320+ | 2026-04-30 | tenant primitive  |
| [`../06_Submodules/per-submodule/helix-billing.md`](../06_Submodules/per-submodule/helix-billing.md) | 320+ | 2026-04-30 | billing primitive |
| [`../06_Submodules/per-submodule/helix-vault.md`](../06_Submodules/per-submodule/helix-vault.md) | 290+ | 2026-04-30 | secrets primitive   |

### 11.2 Forbidden patterns

Clean.

### 11.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_10 execution + operator + compliance officer + CFO signoff.

End of `09_Implementation_Phases/Phase_10_Monetization_and_Auth.md` — 2026-04-30.
