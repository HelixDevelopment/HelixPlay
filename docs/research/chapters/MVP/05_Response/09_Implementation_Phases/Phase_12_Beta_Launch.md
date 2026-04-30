# Phase_12 — Beta Launch

> **Source dimensions:** [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md), [`../08_Operations/04_Observability_and_Events.md`](../08_Operations/04_Observability_and_Events.md), [`../06_Submodules/per-submodule/helix-otel-init.md`](../06_Submodules/per-submodule/helix-otel-init.md), [`../03_Architecture/15_Deployment_Topology.md`](../03_Architecture/15_Deployment_Topology.md) (C16), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P12.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P12.
> **Phase targets:** R-13 anti-bluff beta-readiness + R-12 (Challenges + Smoke) green for 30 consecutive days + Constitution §16 operator signoff.
> **Cross-links:** [`Phase_13_GA.md`](Phase_13_GA.md), [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_12 takes HelixPlay from **internally-shippable** to **closed-beta-customer-shippable**. The operator selects a controlled set of beta customers (recommendation: 3-10 friendly customers across at least 2 regions); they get production-equivalent deployments + dedicated support channels + measurable SLAs.

The Phase_12 work:

- **Beta customer onboarding:** Tenant provisioning per-customer; per-customer SLA negotiated + documented; per-customer success-criteria document.
- **Observability uplift:** Per-tenant Grafana dashboards; per-customer alert routing; per-customer cost report; per-customer Latency + VMAF + ViSQOL trend dashboards.
- **Support tooling:** Operator's support team has session-replay-on-demand, log-search across all services, distributed-trace correlation, billing-dispute lookup, GDPR-erasure-trigger workflow.
- **Feedback loop:** Per-customer feedback channel (operator-mediated; recommendation: dedicated Slack/Teams + monthly review meeting); feedback + bug + feature requests tracked + triaged + responded to within SLA.
- **Performance baselines:** Per-region p999 latency floors + VMAF + ViSQOL floors documented + monitored.
- **Capacity planning:** Per-customer capacity reserved; auto-scale tested + verified; per-region capacity headroom > 30%.
- **Disaster recovery drill:** Per-region DR drill (full region failover); RPO ≤ 1 minute, RTO ≤ 15 minutes verified.
- **30-day green window:** All Challenges + Smoke + Stress + Chaos + Security tests run nightly; 30 consecutive days of green required before Phase_12 signoff.

After Phase_12, the operator has **paid beta customers running in production** with documented SLAs + a clean 30-day green window. This is the final gate before GA in Phase_13.

---

## 2. Prerequisites

- Phase_11 complete + signed off — security hardening done + pentest re-test green.
- Per-region production capacity deployed (per O01 + O03).
- Beta customer commercial agreements signed (operator-side commercial work).
- Operator's support team trained on the support tooling.
- Operator's on-call rotation operational (24×7 coverage minimum).

---

## 3. Tasks Catalogue

| Task ID    | Task                                                                | Subtasks |
|------------|---------------------------------------------------------------------|---------:|
| P12.T01   | Beta customer onboarding workflow + per-customer SLA documents       | 6        |
| P12.T02   | Per-tenant Grafana dashboards (latency + VMAF + ViSQOL + cost)       | 5        |
| P12.T03   | Per-customer alert routing — PagerDuty + email + Slack/Teams        | 4        |
| P12.T04   | Operator support tooling — session replay + log search + trace corr | 5        |
| P12.T05   | Per-customer feedback channel + triage + SLA on response             | 4        |
| P12.T06   | Performance baselines — per-region p999 latency + VMAF + ViSQOL      | 5        |
| P12.T07   | Capacity planning — auto-scale + per-region headroom monitoring      | 5        |
| P12.T08   | Disaster recovery drill — per-region failover + RPO/RTO verification | 5        |
| P12.T09   | Backup + restore drill — per-tenant point-in-time restore           | 4        |
| P12.T10   | 30-day green window enforcement — Challenges + Smoke + Stress + Chaos | 4       |
| P12.T11   | Beta customer success metrics — usage + retention + NPS              | 4        |
| P12.T12   | Beta customer billing reconciliation — first invoice + dispute response | 4     |
| P12.T13   | Per-customer GDPR erasure drill (per-region jurisdictional)          | 4        |
| P12.T14   | Service status page — public-facing + per-component health           | 3        |
| P12.T15   | Beta-to-GA migration playbook                                        | 4        |
| P12.T16   | Phase_12 acceptance review                                            | 2        |

16 tasks; ~68 subtasks.

---

## 4. Task Details

### 4.1 P12.T01 — Beta customer onboarding

Per-customer onboarding workflow: kickoff → tenant provisioning → SLA agreement → dedicated support channel → first session.

**Subtasks:**

- T01.S01 — Kickoff meeting + commercial agreement signing.
- T01.S02 — Tenant provisioning per [Phase_10 P10.T07](Phase_10_Monetization_and_Auth.md#47-p10t07--tenant-provisioning--decommissioning).
- T01.S03 — Per-customer SLA document — latency p999 ≤ X ms, availability ≥ 99.9%, support response ≤ 4 hours.
- T01.S04 — Dedicated support channel (Slack Connect or Teams).
- T01.S05 — First-session smoke test with customer.
- T01.S06 — Customer success-criteria document — what does success look like in 30 / 60 / 90 days.

### 4.2 P12.T02 — Per-tenant Grafana dashboards

Per-tenant dashboards for: latency (p50/p95/p99/p999), VMAF score (per-session histogram), ViSQOL score, cost (per-day + month-to-date), error rate, session count.

**Subtasks:**

- T02.S01 — Latency dashboard (HDR histogram per-stage).
- T02.S02 — VMAF + ViSQOL dashboard (rolling 7-day p50/p99 trend).
- T02.S03 — Cost dashboard (gateway + storage + bandwidth breakdown).
- T02.S04 — Error rate dashboard (per-service + per-stage).
- T02.S05 — Session-count + concurrency dashboard.

### 4.3 P12.T03 — Per-customer alert routing

PagerDuty integration (operator-managed); per-customer escalation policy; email + Slack/Teams notification.

**Subtasks:**

- T03.S01 — PagerDuty service per-customer + per-severity escalation policy.
- T03.S02 — Email notification (operator-managed transactional email).
- T03.S03 — Slack/Teams notification (per-customer dedicated channel).
- T03.S04 — Alert SLA — CRITICAL ≤ 5 min, HIGH ≤ 15 min, MEDIUM ≤ 1 hour.

### 4.4 P12.T04 — Operator support tooling

Operator's support team gets:

- Session replay on demand (helix-record + DASH per Phase_09).
- Log search across all services (Loki).
- Distributed-trace correlation (Tempo + Jaeger).
- Billing-dispute lookup (helix-billing.AuditLog).
- GDPR-erasure-trigger workflow (helix-vault.EraseTenant).

**Subtasks:**

- T04.S01 — Session replay UI for support team.
- T04.S02 — Loki log-search UI with per-tenant scoping.
- T04.S03 — Tempo trace-search UI with span-tag filtering.
- T04.S04 — Billing-dispute lookup UI.
- T04.S05 — GDPR erasure trigger UI (with operator-mediated approval).

### 4.5 P12.T05 — Per-customer feedback channel

Per-customer dedicated Slack Connect or Teams channel. Feedback + bug + feature-request tracked in operator's ticketing system (Jira / Linear / GitHub Issues — operator's choice). Monthly review meeting.

**Subtasks:**

- T05.S01 — Per-customer Slack Connect / Teams channel.
- T05.S02 — Ticketing system integration.
- T05.S03 — SLA on response — feedback acknowledged ≤ 1 business day.
- T05.S04 — Monthly review meeting cadence.

### 4.6 P12.T06 — Performance baselines

Per-region p999 input-to-photons latency floor + VMAF + ViSQOL floor + jitter + packet-loss + reconnect-rate baselines documented + Prometheus-monitored.

**Subtasks:**

- T06.S01 — Per-region p999 latency floor (e.g., NA-East p999 ≤ 8 ms, EU-Central p999 ≤ 9 ms, AP-East p999 ≤ 12 ms).
- T06.S02 — VMAF floor — per-region p10 ≥ 90.
- T06.S03 — ViSQOL floor — per-region p10 ≥ 4.0.
- T06.S04 — Jitter / packet-loss / reconnect-rate baselines.
- T06.S05 — Prometheus alerts when any baseline breached.

### 4.7 P12.T07 — Capacity planning

Per-region capacity reserved (operator's commitment to beta customers). Auto-scale tested + verified end-to-end. Per-region capacity headroom > 30% (operator's safety margin).

**Subtasks:**

- T07.S01 — Per-region capacity commitment document.
- T07.S02 — Auto-scale (Kubernetes HPA + Karpenter) tested for per-customer surge.
- T07.S03 — Per-region capacity headroom Prometheus alert (< 30% → operator escalation).
- T07.S04 — Capacity-burn forecast (operator's CFO-visible monthly projection).
- T07.S05 — Per-customer capacity reservation enforcement (cgroup + ResourceQuota).

### 4.8 P12.T08 — Disaster recovery drill

Per-region failover drill: simulate full region outage; verify per-tenant data preservation + per-tenant session continuity (where possible) + operator's response time.

**Subtasks:**

- T08.S01 — DR drill scoping document.
- T08.S02 — Per-region failover execution (operator-managed).
- T08.S03 — RPO ≤ 1 minute verified (CockroachDB multi-region).
- T08.S04 — RTO ≤ 15 minutes verified (per-tenant session restoration).
- T08.S05 — DR drill report + remediation.

### 4.9 P12.T09 — Backup + restore drill

Per-tenant point-in-time restore — operator can restore a tenant's data to any point in the last 30 days.

**Subtasks:**

- T09.S01 — CockroachDB BACKUP / RESTORE per-tenant.
- T09.S02 — MinIO recording archive + restore.
- T09.S03 — helix-vault namespace backup + restore.
- T09.S04 — Per-tenant point-in-time restore drill.

### 4.10 P12.T10 — 30-day green window

All Challenges + Smoke + Stress + Chaos test suites run nightly. 30 consecutive days of green required before Phase_12 signoff. Any single failure resets the counter.

**Subtasks:**

- T10.S01 — Nightly test scheduler (operator-managed).
- T10.S02 — Per-test result archive (cosign-signed; queryable).
- T10.S03 — Counter-reset logic — any FAIL resets to day-0.
- T10.S04 — Counter-visible operator dashboard.

### 4.11 P12.T11 — Beta customer success metrics

Per-customer usage + retention + NPS (Net Promoter Score). Operator reviews monthly + before GA.

**Subtasks:**

- T11.S01 — Usage metrics — sessions, minutes, bandwidth, recordings.
- T11.S02 — Retention metrics — month-over-month customer continuation.
- T11.S03 — NPS survey — quarterly.
- T11.S04 — Customer health score (composite).

### 4.12 P12.T12 — Beta customer billing reconciliation

First invoice generation + dispute response. Per Phase_10 P10.T12 + P10.T14 — exercised for real on a real customer.

**Subtasks:**

- T12.S01 — First invoice generation + manual review.
- T12.S02 — Customer-side review + sign-off.
- T12.S03 — Dispute response (if any) — operator's billing team triage.
- T12.S04 — Reconciliation lessons-learned document.

### 4.13 P12.T13 — Per-customer GDPR erasure drill

For at least one EU-jurisdictional beta customer — operator runs a synthetic GDPR erasure request (with customer's consent) + verifies erasure executes within SLA + erasure certificate cosign-signed.

**Subtasks:**

- T13.S01 — Synthetic erasure request workflow.
- T13.S02 — Erasure execution + verification.
- T13.S03 — Cosign-signed certificate.
- T13.S04 — Operator's compliance officer sign-off.

### 4.14 P12.T14 — Service status page

Public-facing status page (recommendation: statuspage.io or Atlassian Statuspage or self-hosted). Per-component health (host agent / streaming pipeline / billing / auth / replay / ...). Per-region health.

**Subtasks:**

- T14.S01 — Status page tooling deployed.
- T14.S02 — Per-component health probes.
- T14.S03 — Incident posting workflow (operator-managed).

### 4.15 P12.T15 — Beta-to-GA migration playbook

Once Phase_12 signed off, the operator's GA-readiness gate is documented. The Beta-to-GA migration playbook covers: customer-communication, SLA-uplift, capacity-uplift, support-team scale-up, marketing-launch coordination.

**Subtasks:**

- T15.S01 — Customer communication — beta-end + GA-start dates.
- T15.S02 — SLA uplift — beta SLA → GA SLA differences.
- T15.S03 — Capacity uplift — operator's GA capacity plan.
- T15.S04 — Marketing-launch coordination — operator's commercial team.

### 4.16 P12.T16 — Phase_12 acceptance review

Operator + customer-success + on-call lead + commercial signoff per Constitution §16 + the §6 exit criteria.

---

## 5. Subtask Catalogue

68 subtasks across 16 tasks; bulk-imported.

---

## 6. Exit Criteria

- [ ] At least 3 beta customers onboarded across at least 2 regions.
- [ ] Per-customer SLA documents signed.
- [ ] Per-tenant Grafana dashboards live.
- [ ] Per-customer alert routing operational.
- [ ] Operator support tooling deployed + team trained.
- [ ] Per-customer feedback channel operational.
- [ ] Performance baselines documented + monitored.
- [ ] Capacity planning verified — auto-scale + headroom > 30%.
- [ ] DR drill executed — RPO ≤ 1 min, RTO ≤ 15 min.
- [ ] Backup + restore drill executed.
- [ ] 30 consecutive days of green Challenges + Smoke + Stress + Chaos suite.
- [ ] Beta customer success metrics tracked.
- [ ] First invoice + dispute response cycle executed.
- [ ] GDPR erasure drill executed.
- [ ] Service status page live.
- [ ] Beta-to-GA migration playbook ready.
- [ ] Operator + customer-success + on-call lead + commercial signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP12-01  | Beta customer churns mid-beta due to SLA breach                     | Per-customer SLA monitored + auto-escalation on breach.               |
| RP12-02  | 30-day green window keeps resetting due to flaky test               | Per-test flakiness budget (≤ 1% flake rate) + flaky-test quarantine.  |
| RP12-03  | DR drill reveals data loss > RPO                                    | CockroachDB multi-region replication audit + remediation pre-drill.   |
| RP12-04  | Beta customer cost over-runs operator's capacity plan               | Per-customer capacity reservation + alert when approaching reservation. |
| RP12-05  | Per-customer GDPR erasure reveals leftover artifacts                | Erasure-completeness probe per [helix-vault §4](../06_Submodules/per-submodule/helix-vault.md). |
| RP12-06  | Service status page mis-reports health (false-green during outage)  | Per-component active probes + cross-validate with synthetic monitoring. |

---

## 8. Cross-Family Dependencies

- C16 §6 (Deployment Topology) — architectural source.
- O04 (Observability) — Grafana + Loki + Tempo integration.
- O05 (Backup + DR) — backup + restore tooling.
- T11 (Challenges) — nightly scenario suite.
- T12 (Smoke) — nightly smoke suite.
- helix-otel-init for per-tenant metric scoping.

---

## 9. Acceptance Criteria

Constitution §16 signoff (operator + customer-success + on-call lead + commercial) + §6 exit criteria.

---

## 10. The Phase_12 Calendar

~ 12 weeks (8 weeks operator-side beta execution + 4 weeks 30-day-green-window padding). Operator-side capacity: 6 engineers (2× SRE for capacity + DR; 1× backend for support tooling; 1× security engineer for GDPR drill; 1× customer-success engineer for feedback loop; 1× ops engineer for runbook).

The Phase_12 calendar is **gated by the 30-day green window** — any single test failure during the window resets the counter to day-0. This is by design — we don't ship beta-to-GA until we have 30 consecutive days of evidence the system holds up.

---

## 11. Per-Phase Observability Catalogue

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `helix_beta_customer_active_count` | gauge | region | per-region beta count |
| `helix_beta_session_count` | counter | tenant | per-tenant session rate |
| `helix_beta_30day_green_window_days` | gauge | — | counter resets on FAIL |
| `helix_dr_drill_rpo_seconds` | gauge | region | ≤ 60 |
| `helix_dr_drill_rto_seconds` | gauge | region | ≤ 900 |
| `helix_capacity_headroom_percent` | gauge | region | ≥ 30 |
| `helix_per_tenant_sla_breach_total` | counter | tenant, sli | 0 |

### 11.2 Grafana dashboards

- **Per-customer Beta Health** — usage + SLA status + feedback ticket count per beta customer.
- **Capacity Planning** — per-region utilisation + headroom + auto-scale rate.
- **30-day Green Window Counter** — daily green-streak indicator.
- **DR Drill Status** — per-region last-drill date + RPO + RTO.

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| Beta customer SLA adherence | Per-customer SLA met | 100% | 30-day rolling |
| 30-day green window | Consecutive days of green Challenges + Smoke + Stress + Chaos | 30 days | continuous |
| DR drill RPO | Region-failover data-loss window | ≤ 1 minute | per-drill |
| DR drill RTO | Region-failover recovery time | ≤ 15 minutes | per-drill |
| Capacity headroom | Per-region utilisation budget | ≥ 30% headroom | continuous |
| Customer success ticket SLA | First-response time | ≤ 4 hours | per-ticket |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixOps/docs/runbook/phase12-beta-launch.md` covering beta customer onboarding, dedicated support channel setup, performance baseline establishment, capacity planning, DR drill execution, 30-day green window enforcement.

---

## 14. Implementation Considerations

### 14.1 30-day green window resets

Any single FAIL in any nightly suite resets the counter. Per RP12-02 mitigation: per-test flakiness budget (≤ 1%) + flaky-test quarantine.

### 14.2 Per-customer feedback channel

Dedicated Slack Connect or Teams per customer; monthly review meeting cadence. Operator's customer-success team owns triage.

### 14.3 DR drill scope

Per-region failover with full data preservation verification. RP12-03 mitigation: pre-drill replication audit + remediation.

### 14.4 Per-customer GDPR drill

EU-jurisdictional beta customer with synthetic erasure request (with consent). Per Phase_12 P12.T13.

---

## 15. Phase_12 Cost Estimation

Beta-stage operational cost: ~$2,000 / month / region for capacity headroom + DR drill resources. Per-customer support tooling: operator-side already deployed in Phase_05/Phase_06.

---

## 16. Cross-Mirror Parity Verification

Phase_12 closure verification per the [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern.

---

## 17. Anti-Bluff Verification

### 11.1 Sources resolved

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`Phase_11_Hardening_and_Security.md`](Phase_11_Hardening_and_Security.md) | 400+ | 2026-04-30 | predecessor                                      |
| [`../03_Architecture/15_Deployment_Topology.md`](../03_Architecture/15_Deployment_Topology.md) | 1,500+ | 2026-04-30 | C16 architectural source             |
| [`../08_Operations/04_Observability_and_Events.md`](../08_Operations/04_Observability_and_Events.md) | 300 | 2026-04-30 | O04 observability        |
| [`../06_Submodules/per-submodule/helix-otel-init.md`](../06_Submodules/per-submodule/helix-otel-init.md) | 280+ | 2026-04-30 | OTel primitive |

### 11.2 Forbidden patterns

Clean.

### 11.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_12 execution + operator + customer-success + on-call lead + commercial signoff.

End of `09_Implementation_Phases/Phase_12_Beta_Launch.md` — 2026-04-30.
