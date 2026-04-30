# Phase_03 — Backend Services

> **Source dimensions:** [`Phase_02_Core_Submodules.md`](Phase_02_Core_Submodules.md), [`../06_Submodules/per-submodule/helix-vault.md`](../06_Submodules/per-submodule/helix-vault.md), [`../06_Submodules/per-submodule/helix-tenant.md`](../06_Submodules/per-submodule/helix-tenant.md), [`../06_Submodules/per-submodule/helix-grpc-frame.md`](../06_Submodules/per-submodule/helix-grpc-frame.md), [`../03_Architecture/05_RealTime_APIs.md`](../03_Architecture/05_RealTime_APIs.md) (C06), [`../03_Architecture/08_Scalability_and_MultiRegion.md`](../03_Architecture/08_Scalability_and_MultiRegion.md) (C09), [`../00_Master_Plan.md`](../00_Master_Plan.md) §7.2 row P03.
> **Source line count:** chapter floor 500 lines per Master Plan §7.2 row P03.
> **Phase targets:** R-05 + R-06 + R-07 + R-08 — the production-tier backend infrastructure goes live.
> **Cross-links:** [`Phase_04_Streaming_Pipeline.md`](Phase_04_Streaming_Pipeline.md), [`Phase_10_Monetization_and_Auth.md`](Phase_10_Monetization_and_Auth.md).
> **Status:** Draft v1 specification.
> **Last updated:** 2026-04-30.

---

## 1. Phase Scope

Phase_03 stands up **production-tier backend services**: CockroachDB cluster (3-node minimum, scaling to 9-node multi-region), NATS clustering, Redis Sentinel, Vault Enterprise (or OSS Shamir-share at production scale), Coturn TURN server, and the supporting service-discovery infrastructure. The Phase consumes the v1.0.0-graduated **helix-vault**, **helix-tenant**, and **helix-grpc-frame** from Phase_02 to wire authoritative storage + tenant management + control-plane RPC.

After Phase_03, the system has **persistent state** for the first time. Sessions can be created + tracked across operator restarts; tenant configurations persist; OAuth tokens get vaulted; per-session billing events flow through NATS. Phase_04 (streaming pipeline) requires this substrate.

---

## 2. Prerequisites

- Phase_02 complete + signed off — helix-vault, helix-tenant, helix-grpc-frame at v1.0.0.
- Phase_00's single-node backing services must be migrated to production-tier.
- Operator decision-points latched: Vault Enterprise vs OSS Shamir-share; CockroachDB single-region vs multi-region; NATS standalone vs JetStream cluster.

---

## 3. Tasks Catalogue

| Task ID    | Task                                                          | Subtasks |
|------------|---------------------------------------------------------------|---------:|
| P03.T01   | CockroachDB 3-node cluster (single-region)                    | 6        |
| P03.T02   | NATS clustering with JetStream                                | 5        |
| P03.T03   | Redis Sentinel (master + 2 replicas)                          | 4        |
| P03.T04   | Vault production deployment (HA + auto-unseal or Shamir)      | 5        |
| P03.T05   | Coturn TURN server with relay                                 | 4        |
| P03.T06   | Backup + restore procedures for each backing service          | 6        |
| P03.T07   | Service-discovery integration (mDNS or DoH per Phase_00 latch)| 4        |
| P03.T08   | Backing-service health probes integrated with HelixQA         | 3        |
| P03.T09   | Per-tenant catalog provisioning workflow                      | 4        |
| P03.T10   | Migration of Phase_00's single-node CockroachDB → cluster    | 3        |
| P03.T11   | Phase_03 acceptance review                                     | 2        |

11 tasks, ~46 subtasks.

---

## 4. Task Details

### 4.1 P03.T01 — CockroachDB 3-node cluster

Production deployment per [helix-vault descriptor §4](../06_Submodules/per-submodule/helix-vault.md) + [C09 §6 multi-region scalability](../03_Architecture/08_Scalability_and_MultiRegion.md) ratified at the host-agent's session-storage layer:

- 3 nodes minimum for quorum; 5 nodes recommended for production.
- TLS certificates via cert-manager (operator-provisioned).
- Backup to MinIO via `cockroach backup` daily incremental + weekly full.
- Per-tenant table-per-tenant pattern (per [helix-tenant descriptor §3.1](../06_Submodules/per-submodule/helix-tenant.md#31-the-default--o-in-tree)).

### 4.2 P03.T02 — NATS clustering with JetStream

JetStream enabled for billing events (per [O04 §5 NATS subjects](../08_Operations/04_Observability_and_Events.md#5-the-nats-event-bus)):

- 3-node NATS cluster.
- JetStream stream for `helix.session.*.ended` + `billing.*` (durable retention).
- Consumer groups per microservice.

### 4.3 P03.T03 — Redis Sentinel

Master + 2 replicas + Sentinel quorum-based failover. Per-tenant rate-limit keys via `helix-tenant.FeatureFlags`.

### 4.4 P03.T04 — Vault production

Per [helix-vault descriptor §3.3](../06_Submodules/per-submodule/helix-vault.md#33-external-system) + Phase_00 decision-point latch. HA mode + storage backend (Consul or operator-chosen).

### 4.5 P03.T05 — Coturn

TURN server for ICE NAT traversal in Phase_04's helix-transport. Static credentials via Vault.

### 4.6 P03.T06 — Backup + restore

Per Constitution §11 + §16 *Acceptance* — restorable to any of the past 30 days for compliance audit. Backup procedures + restore drills committed to runbook directory.

### 4.7 P03.T07–T11

Service discovery + health probes + tenant provisioning + migration + acceptance review. Each follows the Phase_00 task pattern adapted to backing-service operations.

---

## 4a. Detailed CockroachDB Cluster Provisioning

Per the per-tenant table-per-tenant pattern + multi-region readiness:

```yaml
P03.T01.S01:
  title: "[P03.T01.S01] Provision 3-node CockroachDB cluster"
  body: |
    docker compose -f vasic-digital/Containers/lanes/cockroach-cluster-1.x/docker-compose.yml up -d
    Cluster initialised with `cockroach init`; verify with `cockroach node status`.
P03.T01.S02:
  title: "[P03.T01.S02] TLS cert provisioning + cert-manager integration"
  body: |
    Operator-provisioned cert-manager + per-node certificates.
    `cockroach cert create-node`/create-client per cluster member + per-tenant client.
P03.T01.S03:
  title: "[P03.T01.S03] Per-tenant database + table provisioning"
  body: |
    For each operator tenant: `CREATE DATABASE tenant_<id>; GRANT ALL ON DATABASE tenant_<id> TO helix_<id>;`
    Per-table creation per Constitution §11 tenant-isolation pattern.
P03.T01.S04:
  title: "[P03.T01.S04] Daily incremental + weekly full backup to MinIO"
  body: |
    `cockroach backup INCREMENTAL FROM ...` cron; tested restore at the §6 exit criteria.
P03.T01.S05:
  title: "[P03.T01.S05] Configure cluster observability"
  body: |
    Prometheus scrape config + Grafana dashboard import per O04 §6.
P03.T01.S06:
  title: "[P03.T01.S06] Cluster failover drill"
  body: |
    Stop one node; verify cluster continues + leader re-election; restart node + verify rejoin.
```

## 4b. Detailed NATS JetStream Setup

```yaml
P03.T02.S01: "[P03.T02.S01] 3-node NATS cluster with JetStream"
P03.T02.S02: "[P03.T02.S02] Subjects + JetStream streams per O04 §5"
P03.T02.S03: "[P03.T02.S03] Consumer groups for each subscriber"
P03.T02.S04: "[P03.T02.S04] Cluster failover drill"
P03.T02.S05: "[P03.T02.S05] OTel + Prometheus integration"
```

## 4c. Detailed Vault Production Deployment

Per [helix-vault descriptor §3.3 + §8](../06_Submodules/per-submodule/helix-vault.md):

```yaml
P03.T04.S01: "[P03.T04.S01] HA-mode Vault with 3-node Consul or Raft storage"
P03.T04.S02: "[P03.T04.S02] Auto-Unseal (Enterprise) or Shamir-share (OSS) per Phase_00 latch"
P03.T04.S03: "[P03.T04.S03] Per-submodule AppRole policy creation"
P03.T04.S04: "[P03.T04.S04] KEK rotation drill — under-active-session test"
P03.T04.S05: "[P03.T04.S05] Audit log integration with SIEM forwarding"
```

## 4d. The Backup + Restore Drill Procedure

For each backing service, operator runs:

1. **Backup**: trigger a fresh backup; verify the backup file SHA-256 matches the manifest.
2. **Restore probe**: bring up a sandbox cluster; restore the backup; verify the restored data matches the source.
3. **Document RPO + RTO**: recovery-point-objective (data-loss window) + recovery-time-objective (downtime window). Per Constitution §11 production-tier RPO ≤ 24 h + RTO ≤ 4 h.

The drill produces a signed run-archive entry; operator signs off the drill before Phase_03 closes.

## 5. Subtask Catalogue

46 subtasks across 11 tasks; bulk-imported.

---

## 6. Exit Criteria

- [ ] CockroachDB 3-node cluster operational + backup + restore drilled.
- [ ] NATS JetStream cluster operational + topics configured.
- [ ] Redis Sentinel operational with verified failover.
- [ ] Vault HA-mode operational with verified KEK rotation under load.
- [ ] Coturn TURN server reachable from external clients.
- [ ] All 5 backing services emit metrics/traces/logs to the observability stack.
- [ ] Service-discovery operational (mDNS or DoH per Phase_00 latch).
- [ ] HelixQA continuous-mode probes the 5 services healthily for ≥ 7 consecutive days.
- [ ] Operator signoff per Constitution §16.

---

## 7. Risk Register

| ID       | Risk                                                                | Mitigation                                                            |
|----------|----------------------------------------------------------------------|------------------------------------------------------------------------|
| RP03-01  | CockroachDB cluster split-brain                                       | 3-node quorum + automatic leader election; tested via Toxiproxy partition. |
| RP03-02  | Vault unseal-key loss                                                | Phase_00 decision-point latch — Auto-Unseal or Shamir-share documented. |
| RP03-03  | NATS JetStream stream corruption                                     | Mirroring across 3 nodes + Phase_03's restore-drill task.             |
| RP03-04  | Migration from Phase_00 single-node CockroachDB drops data           | T10 migration procedure tested in dry-run before live cutover.         |
| RP03-05  | Coturn under DDoS                                                    | Operator-side rate-limiting + monitoring.                              |

---

## 8. Cross-Family Dependencies

- helix-vault, helix-tenant, helix-grpc-frame (Phase_02 graduates).
- C06 §6 RealTime APIs + C09 §6 Scalability + C10 §6 Security architecturally specify the patterns.
- O01-O06 (Operations) operationalises the deployment.

---

## 9. Acceptance Criteria

Constitution §16 signoff + §6 exit criteria.

---

## 10. Anti-Bluff Verification

### 10.1 Sources resolved

| Path                                                              | Lines  | Role                                            |
|-------------------------------------------------------------------|-------:|-------------------------------------------------|
| [`Phase_02_Core_Submodules.md`](Phase_02_Core_Submodules.md)      |    500 | predecessor — helix-vault/tenant/grpc-frame ready |
| [`../06_Submodules/per-submodule/helix-vault.md`](../06_Submodules/per-submodule/helix-vault.md) |    307 | Vault wrapper                |
| [`../06_Submodules/per-submodule/helix-tenant.md`](../06_Submodules/per-submodule/helix-tenant.md) |    310 | tenant config              |
| [`../03_Architecture/05_RealTime_APIs.md`](../03_Architecture/05_RealTime_APIs.md) | 3,450 | C06 architectural source                |
| [`../03_Architecture/08_Scalability_and_MultiRegion.md`](../03_Architecture/08_Scalability_and_MultiRegion.md) | 3,425 | C09 architectural source |

### 10.2 Forbidden patterns

Clean.

### 10.3 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 (specification only).
- Pending: Phase_03 execution + operator signoff.

End of `09_Implementation_Phases/Phase_03_Backend_Services.md` — 2026-04-30.
