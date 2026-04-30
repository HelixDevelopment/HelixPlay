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

46 subtasks across 11 tasks; per-task discrete `[P03.Tyy.Szz]` tickets per [O05](../08_Operations/05_Tracking_GitHub_GitLab.md). Sub-categories: CockroachDB (T01..T03), NATS JetStream (T04..T05), Redis Sentinel (T06), Vault HA (T07..T08), Coturn (T09), service mesh + DNS (T10..T11). Per-subtask: helm-chart deployment + post-deploy verification probe + observability emission + backup + restore drill reference per §4d.

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

## 10. The Phase_03 Calendar

~ 4 weeks. Operator-side capacity: 4 engineers (2× SRE for CockroachDB + NATS deployment; 1× security engineer for Vault production; 1× ops engineer for Coturn + service mesh).

---

## 10a. Per-Phase Detailed Task Acceptance Criteria

### 10a.1 CockroachDB cluster acceptance (P03.T01..T03)

- 3-node minimum (5-node recommended) cluster operational across operator's region(s); per-node 16 vCPU + 64 GB RAM + NVMe SSD.
- TLS-only client connections (root + per-tenant SQL users); cluster-internal mTLS.
- Multi-region replication enabled (per Phase_12 P12.T08 RPO ≤ 1 minute target).
- Per-tenant database namespace with row-level security policies.
- Backup to MinIO every 4 hours; retention 30 days; restore drill quarterly per [helix-vault §4](../06_Submodules/per-submodule/helix-vault.md).
- Schema migrations via cockroach-sql migration tool; per-PR review before apply.

### 10a.2 NATS JetStream acceptance (P03.T04..T05)

- 3-node JetStream cluster; replication factor 3 on critical streams.
- Per-stream retention policy operator-tunable (default: 7 days).
- Per-tenant subject hierarchy: `helix.<tenant>.<service>.<event-type>`.
- TLS-only client connections; per-service NKey authentication.
- Streams for billing events (`helix.billing.events.<tenant>`), session lifecycle (`helix.sessions.<tenant>`), audit log (`helix.audit.<tenant>`).

### 10a.3 Redis Sentinel acceptance (P03.T06)

- 3-node Sentinel quorum + 1 master + 2 replicas per tenant tier.
- Automatic failover within 10 s on master failure.
- Per-tenant database isolation (Redis logical DB per tenant).
- Read-only replicas for high-volume read paths (session caching, rate-limit counters).
- TLS-only client connections.

### 10a.4 Vault HA acceptance (P03.T07..T08)

- 3-node Vault HA cluster with auto-unseal via cloud-KMS or per-region HSM.
- Per-tenant namespace isolation per [helix-vault §3](../06_Submodules/per-submodule/helix-vault.md).
- KEK rotation policy active (annual cadence).
- Per-tenant DEK lifecycle automated.
- Audit log to operator's SIEM.

### 10a.5 Coturn acceptance (P03.T09)

- Per-region Coturn cluster (2-node minimum) behind operator's edge firewall.
- STUN + TURN-over-UDP + TURN-over-TCP supported.
- Per-tenant relay credentials rotated via Vault.
- Bandwidth budget per-tenant enforceable.

### 10a.6 Service mesh + DNS acceptance (P03.T10..T11)

- mDNS/DNS-SD on operator's LAN per [O03 §2](../08_Operations/03_Service_Discovery_and_Ports.md).
- DoH alternative for jurisdiction-restricted operators (Russian-jurisdiction path).
- Cilium NetworkPolicy default-deny baseline.
- Per-service Prometheus scrape configuration via DNS-SD.

---

## 11. Per-Phase Observability Catalogue

### 11.1 Prometheus metrics

| Metric | Type | Labels | SLO Target |
|--------|------|--------|------------|
| `cockroach_sql_query_seconds` | histogram | tenant | p99 ≤ 50 ms |
| `cockroach_replication_lag_seconds` | gauge | range, region | p99 ≤ 1 s |
| `nats_jetstream_messages_total` | counter | stream, tenant | per-stream rate |
| `nats_jetstream_consumer_pending` | gauge | stream, consumer | < 1000 steady |
| `redis_connected_clients` | gauge | tenant | per-tenant connection budget |
| `redis_replication_lag_seconds` | gauge | replica | p99 ≤ 100 ms |
| `vault_secret_access_total` | counter | tenant, mount | per-secret audit |
| `vault_token_lifecycle_total` | counter | tenant, op | rotation cadence |
| `coturn_active_relays` | gauge | region | per-tenant TURN |

### 11.2 Grafana dashboards

- **Per-region CockroachDB Health** — SQL latency / replication lag / per-range health.
- **Per-tenant NATS JetStream** — stream depth + consumer lag + per-tenant retention.
- **Per-tenant Redis** — connection count + replication lag + per-tenant memory.
- **Per-tenant Vault** — secret-access audit + token lifecycle + KV-v2 versioning.
- **Per-region Coturn** — active relays + bandwidth + per-tenant relay budget.

---

## 12. Per-Phase SLI / SLO Definitions

| SLI | Definition | SLO Target | Window |
|-----|------------|------------|--------|
| CockroachDB query latency | SQL query p99 latency | ≤ 50 ms | 7-day rolling |
| CockroachDB replication | Multi-region replication lag p99 | ≤ 1 s | 7-day rolling |
| NATS JetStream durability | Messages delivered with at-least-once | 100% | per-message |
| Redis availability | Per-tenant Redis Sentinel availability | ≥ 99.95% | 30-day rolling |
| Vault availability | Per-tenant secret retrieval | ≥ 99.99% | 30-day rolling |
| Coturn relay availability | Per-region TURN relay | ≥ 99.9% | 30-day rolling |
| Backup + restore drill | Per-quarter drill executed | 100% | quarterly |

---

## 13. Per-Phase Operator Runbook

`HelixDevelopment/HelixOps/docs/runbook/phase03-backend-services.md` covering CockroachDB cluster bootstrap, NATS JetStream stream provisioning, Redis Sentinel failover drill, Vault HA + KEK rotation, Coturn STUN/TURN configuration, per-region service-mesh deployment, backup + restore drill procedure.

---

## 13a. Per-Phase Risk Mitigation Detail

Each Risk Register entry from §7 elaborated with concrete monitoring + remediation procedures:

### 13a.1 RP03-01 — CockroachDB cluster split-brain

**Detection:** Prometheus alert on `cockroach_distsender_rangelookups_total` divergence across nodes; alarm if range-lookup count differs by > 10% across nodes within 1-minute window.

**Mitigation:** 5-node minimum (vs 3-node default) for production; per-node clock skew NTP-enforced; etcd-style quorum-based consensus.

**Remediation:** Operator's runbook §3.2 — manual quorum repair via cockroach debug tooling; if repair fails, escalate to multi-region replication promotion.

### 13a.2 RP03-02 — NATS JetStream durability loss

**Detection:** `nats_jetstream_replicas_count` < configured replication-factor.

**Mitigation:** Replication factor 3 on critical streams; cross-AZ replica placement; per-stream operator-tunable storage limits.

**Remediation:** Operator's runbook §4.5 — replica replacement procedure; per-stream re-replication triggered automatically.

### 13a.3 RP03-03 — Redis Sentinel failover delay

**Detection:** Failover-trigger event observed but new master not elected within 10 s.

**Mitigation:** Per-region Sentinel quorum; per-tier Redis tier separation (operator's mission-critical paths use dedicated Redis cluster).

**Remediation:** Operator's runbook §5.3 — Sentinel re-quorum procedure; manual master promotion if quorum unrecoverable.

### 13a.4 RP03-04 — Vault auto-unseal failure

**Detection:** Vault sealed-state observed > 30 s post-restart.

**Mitigation:** Per-region cloud-KMS or HSM redundancy; per-cluster Shamir secret-sharing fallback for air-gapped operators.

**Remediation:** Operator's runbook §6.2 — manual unseal via Shamir keys held by 3-of-5 operator key custodians.

### 13a.5 RP03-05 — Coturn DDoS amplification

**Detection:** Per-relay bandwidth gauge spike to > 10× rolling baseline.

**Mitigation:** Per-tenant relay credentials rotated 90 days; per-relay bandwidth budget; rate-limit on STUN/TURN handshakes.

**Remediation:** Operator's runbook §7.4 — abusive-tenant credential revocation + IP block-list update.

---

## 14. Implementation Considerations

### 14.1 CockroachDB multi-region vs single-region

Multi-region for ≥ 99.99% availability + RPO ≤ 1 minute (per Phase_12 P12.T08). Single-region acceptable for early-stage operators with ≤ 99.9% SLA tier.

### 14.2 NATS JetStream vs Kafka

NATS JetStream selected per [C06](../03_Architecture/05_Data_Plane.md) — operator-self-hosted, simpler ops, sufficient throughput for per-tenant event streams. Kafka considered + rejected for ops complexity.

### 14.3 Vault HA topology

Vault HA (3-node cluster) with auto-unseal via cloud-KMS or per-region HSM. Per-tenant namespace isolation per [helix-vault §3](../06_Submodules/per-submodule/helix-vault.md).

### 14.4 Coturn deployment placement

Coturn placed inside operator's network per Phase_11 RP11-12 (DDoS scrubbing). Per-region Coturn cluster behind operator's edge firewall.

---

## 15. Phase_03 Cost Estimation

Per-region monthly cost (3-node clusters):

| Component | Per-Region Per-Month |
|-----------|----------------------|
| CockroachDB 3-node (16 vCPU + 64 GB RAM each) | ~$1,200 |
| NATS JetStream 3-node (4 vCPU + 16 GB RAM each) | ~$240 |
| Redis Sentinel 3-node (2 vCPU + 8 GB RAM each) | ~$120 |
| Vault HA 3-node (2 vCPU + 8 GB RAM each) | ~$120 |
| Coturn 2-node + bandwidth | ~$200 |
| **Total per-region per-month** | **~$1,880** |

Per-region scaling cost amortised across all tenants in that region.

---

## 16. Cross-Mirror Parity Verification

Phase_03 closure verification per the [Phase_09 §16](Phase_09_Recording_and_Replay.md#16-cross-mirror-parity-verification) pattern.

---

## 17. Anti-Bluff Verification

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
