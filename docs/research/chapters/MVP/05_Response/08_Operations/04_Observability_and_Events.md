# O04 — Observability & Event-Bus Operations

> **Source dimensions:** [`00_Index.md`](00_Index.md); [`../01_Constitution.md`](../01_Constitution.md) §10 (R-08 Observability); [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §3 + §8.
> **Source line count:** chapter floor 300 lines per Master Plan §7.2 row O04.
> **Chapter targets:** R-08 (heavy use of events + observability for real-time propagation), R-13.
> **Cross-links:** [`03_Service_Discovery_and_Ports.md`](03_Service_Discovery_and_Ports.md), [`02_Quality_Gates_SonarQube_Snyk.md`](02_Quality_Gates_SonarQube_Snyk.md), [`05_Tracking_GitHub_GitLab.md`](05_Tracking_GitHub_GitLab.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. The Observability Stack in Three Sentences

HelixPlay's observability stack is **OpenTelemetry-end-to-end** for traces + metrics + logs, with **NATS** as the event bus, **Prometheus** as the metric scraper, **Grafana** as the visualisation surface, and **Loki** as the log-aggregator. Every state change in every submodule emits an OTLP span; every metric defined in the per-submodule [S05 §9.5 *Observability metrics catalog*](../06_Submodules/per-submodule/) is scraped by Prometheus; every dashboard at [S04 §3 + T12 §9c](../07_Testing/12_HelixQA_Autonomous.md#9c-helixqa-dashboard-tour) is sourced from this stack. R-08 demands real-time propagation of state changes — the OTel + NATS combination is the operationalisation.

---

## 2. The Stack Components

| Component             | Role                                       | Pinned version              |
|-----------------------|--------------------------------------------|------------------------------|
| OpenTelemetry collector | OTLP gRPC + OTLP HTTP receiver           | otelcol-contrib ≥ v0.110     |
| Prometheus             | Metric scraping + storage                  | ≥ v2.55                      |
| Grafana                | Dashboard rendering                        | ≥ v11.4                      |
| Loki                    | Log aggregation                            | ≥ v3.3                       |
| Tempo                   | Trace storage backend                      | ≥ v2.7                       |
| NATS                    | Event bus                                  | ≥ v2.10                      |
| nats-cli                | Operator CLI                               | ≥ v0.1.5                     |
| promtail                | Log shipper from container stdout to Loki | ≥ v3.3                       |

All shipped as containers (per R-05 + R-06) under `vasic-digital/Containers/lanes/observability-1.x/`. Operator deploys via a single `docker compose up -d` from `vasic-digital/Containers/lanes/observability-1.x/docker-compose.yml`.

---

## 3. The OTLP Trace Pipeline

Every submodule's OTel SDK auto-instruments + ships traces:

```
HelixPlay submodule (Go)
  ↓ OTLP gRPC (4317)
otel-collector
  ↓ OTLP gRPC export
Tempo (trace storage)
  ↑ Grafana queries via Tempo datasource
```

The OTel SDK is initialised once per submodule via `helix-otel-init` (a thin wrapper shipped under `vasic-digital/.github/`):

```go
import (
    "context"
    helixotel "github.com/vasic-digital/helix-otel-init"
)

func main() {
    shutdown, err := helixotel.Init(context.Background(), helixotel.Config{
        ServiceName:    "helix-pipeline",
        ServiceVersion: BuildVersion,
        ExporterEndpoint: os.Getenv("HELIX_OTLP_ENDPOINT"),
    })
    if err != nil { /* fatal */ }
    defer shutdown(context.Background())
    // ...
}
```

The wrapper sets up:
- Trace exporter (OTLP gRPC).
- Metric exporter (OTLP gRPC).
- Resource attributes (`service.name`, `service.version`, `service.instance.id`).
- Sampler — head-based at 100 % for non-prod, ratio-based at 1 % for prod (operator-configurable).

---

## 4. Per-Submodule Metrics Catalogue

Each submodule's [S05 §9.5 *Observability metrics catalog*](../06_Submodules/per-submodule/) entry is the canonical metric definition. Aggregated across the 29 submodules, the metric inventory is ≈ 200 distinct metric names. Examples:

- `helix_pipeline_frames_processed_total` (counter)
- `helix_pipeline_end_to_end_latency_seconds` (histogram)
- `helix_shm_pool_size` (gauge)
- `helix_vault_kek_rotation_events_total` (counter)
- `helix_thermal_temperature_celsius` (gauge)

Prometheus scrapes via the Prometheus → otel-collector pull-mode receiver every 15 s. Grafana queries via PromQL.

---

## 5. The NATS Event Bus

NATS carries **state-change events** (vs OTel's request-tracing). Subjects:

| Subject pattern                              | Producer                           | Consumer                                   |
|----------------------------------------------|------------------------------------|--------------------------------------------|
| `helix.session.{tenant}.{session}.started`   | helix-host-agent                   | helix-record + helix-bench + HelixQA       |
| `helix.session.{tenant}.{session}.ended`     | helix-host-agent                   | helix-record + billing service              |
| `helix.deployment.{cadence}.gate-decision`   | HelixQA                            | dashboards + on-call paging                 |
| `helix.alert.{severity}.{alert-type}`         | HelixQA                            | Slack + PagerDuty                          |
| `helix.tenant.{tenant}.theme.changed`        | helix-tenant                       | client cache invalidation                  |
| `helix.thermal.{device}.scaling-changed`    | helix-thermal                      | helix-pipeline                              |

Subjects are **hierarchical**; consumers subscribe with wildcards (`helix.session.>` for all sessions; `helix.alert.p1.*` for all P1 alerts).

NATS retention: events retained 7 days (operator-tunable); replay via JetStream where the consumer needs durable subscription (e.g. billing-service consuming `helix.session.>.ended` for late-arriving events).

---

## 6. Grafana Dashboards

Per [T12 §9c](../07_Testing/12_HelixQA_Autonomous.md#9c-helixqa-dashboard-tour), four canonical dashboards:

| Dashboard            | Primary panels                                                                  |
|----------------------|---------------------------------------------------------------------------------|
| **Fleet Health**     | 29-submodule rollup; per-submodule green/red over the last 7 days.             |
| **Per-Submodule Deep-Dive** | Per-submodule cell heatmap (cadence × test-type × mirror) over the last 30 days; click into any cell for the run-archive entry. |
| **Deployment Gate**  | Real-time gate decisions; VerdictDeploy / VerdictBlock / VerdictError ratios. |
| **Change-Point Trends** | VMAF / ViSQOL / latency p999 / billing-event-count trends over the last 30 days. |
| **On-Call**          | Open alerts; rotation roster; SLA timers.                                       |

Dashboards are committed under `HelixDevelopment/HelixQA/dashboards/` as JSON exports + Grafanan provisioning. The operator's Grafana instance imports them via `grafana-provisioning/dashboards/`.

---

## 7. Log Aggregation via Loki + promtail

Every container's stdout/stderr is shipped to Loki via promtail:

```yaml
# vasic-digital/Containers/lanes/observability-1.x/promtail.yml
clients:
  - url: http://loki:3100/loki/api/v1/push
scrape_configs:
  - job_name: containers
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
    pipeline_stages:
      - docker: {}
      - labels:
          tenant_id:   # from log field
          submodule:   # from log field
```

Loki stores 30 days hot + 1 year warm + forever cold (operator-tunable per Constitution §16 *Acceptance* compliance retention).

---

## 8. The OTel Collector Pipeline Configuration

```yaml
# vasic-digital/Containers/lanes/observability-1.x/otelcol-config.yml
receivers:
  otlp:
    protocols:
      grpc: { endpoint: 0.0.0.0:4317 }
      http: { endpoint: 0.0.0.0:4318 }

processors:
  batch: { timeout: 5s }
  memory_limiter:
    check_interval: 5s
    limit_percentage: 75
    spike_limit_percentage: 25
  resource:
    attributes:
      - key: deployment.environment
        value: ${HELIX_ENV}   # prod / canary / dev
        action: insert

exporters:
  otlp/tempo:
    endpoint: tempo:4317
    tls: { insecure: true }
  prometheusremotewrite:
    endpoint: http://prometheus:9090/api/v1/write
  loki:
    endpoint: http://loki:3100/loki/api/v1/push

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [otlp/tempo]
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [prometheusremotewrite]
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [loki]
```

Operator-configurable via env vars + secrets per Constitution §11.5 (no hardcoded credentials; Vault-mounted).

---

## 9. Alerting via Prometheus AlertManager + NATS

Prometheus AlertManager evaluates rules + fires alerts via the NATS bus per [O04 §5](#5-the-nats-event-bus). The fan-out per [T12 §4 alert-routing playbook](../07_Testing/12_HelixQA_Autonomous.md#4-alert-routing-playbook):

- **P1**: phone + SMS + Slack #helixplay-p1.
- **P2**: Slack #helixplay-p2 + email.
- **P3**: Slack #helixplay-p3.

AlertManager → NATS publishers operationalise the §4 playbook. The runbook URL is part of every alert payload.

---

## 9a. The Sampling Decision Tree

OTel head-based sampling is operationally tunable per environment:

| Environment       | Sampler                       | Rate          | Tail-sampling? |
|-------------------|-------------------------------|---------------|-----------------|
| Local development | AlwaysOn                      | 100 %         | no              |
| CI                | AlwaysOn                      | 100 %         | no              |
| Canary            | TraceIDRatioBased             | 10 %          | tail-error 100 % |
| Production        | TraceIDRatioBased             | 1 %           | tail-error 100 % |

Tail-sampling at 100 % for errors means every traced operation that ends in an error span gets retained even if its head-sampling decision was "drop". The combination keeps storage cost bounded while preserving every interesting (i.e. failed) trace.

Configured via `OTEL_TRACES_SAMPLER` + `OTEL_TRACES_SAMPLER_ARG` env-vars per Constitution §10.

## 9b. The Per-Submodule Metric Export Wrapper

To minimise per-submodule boilerplate, every submodule consumes `helix-otel-init` (per [§3](#3-the-otlp-trace-pipeline)) which wires up the metric exporter. Each metric is registered once via:

```go
import "go.opentelemetry.io/otel/metric"

var (
    framesProcessedCounter = otel.Meter("helix-pipeline").Int64Counter("helix_pipeline_frames_processed_total")
    endToEndLatency        = otel.Meter("helix-pipeline").Float64Histogram("helix_pipeline_end_to_end_latency_seconds")
)
```

The metrics emit to the OTel collector via OTLP gRPC every 10 s (operator-tunable via `OTEL_METRIC_EXPORT_INTERVAL`). Prometheus scrapes the collector via `prometheus_remote_write` receiver.

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-O04-A            | OpenTelemetry sampling — head-based at 1 % vs tail-based at 100 % errors-only?                                | this chapter next revision                          |
| OQ-O04-B            | Loki log retention — 1 year warm vs 7 days hot only?                                                          | this chapter next revision                          |
| OQ-O04-C            | NATS JetStream adoption — for which event subjects beyond billing?                                            | C06 §6 next revision                                |

---

## 9c. The Operator's Dashboard Provisioning One-Time Setup

Operator runs once at Phase_00:

```bash
# Apply Grafana dashboard provisioning.
docker compose -f vasic-digital/Containers/lanes/observability-1.x/docker-compose.yml exec grafana \
    grafana-cli admin reset-admin-password $GRAFANA_ADMIN_PASSWORD

# Provision data sources + dashboards.
cd HelixDevelopment/HelixQA/dashboards
./provision.sh \
    --grafana-url=http://grafana:3000 \
    --admin-token=$GRAFANA_ADMIN_TOKEN
```

The `provision.sh` script:
1. Creates Tempo data source pointing at `http://tempo:3200`.
2. Creates Prometheus data source pointing at `http://prometheus:9090`.
3. Creates Loki data source pointing at `http://loki:3100`.
4. Imports the four canonical dashboards as JSON.
5. Creates the operator + read-only user accounts.

## 9d. The Multi-Tenant Dashboard Filtering

Per [helix-tenant descriptor §9.6](../06_Submodules/per-submodule/helix-tenant.md#96-consumer-matrix), every metric carries a `tenant_id` label. The Grafana dashboards default to the operator's view (all tenants); per-tenant operators see a tenant-scoped variant via the `__tenant_id` template variable. Cross-tenant data leakage is forbidden — tenant operators cannot see other tenants' metrics.

## 11. References & Anti-Bluff Verification

### 11.1 Internal

- [`../06_Submodules/04_HelixQA_Integration.md`](../06_Submodules/04_HelixQA_Integration.md) §3, §8.
- [`../07_Testing/12_HelixQA_Autonomous.md`](../07_Testing/12_HelixQA_Autonomous.md) §4, §9c.
- [`../06_Submodules/per-submodule/`](../06_Submodules/per-submodule/) §9.5 of every descriptor — the per-submodule metrics catalogue.
- [`../01_Constitution.md`](../01_Constitution.md) §10.

### 11.2 External (web)

- OpenTelemetry: https://opentelemetry.io/docs/ (accessed 2026-04-30).
- Prometheus: https://prometheus.io/docs/ (accessed 2026-04-30).
- Grafana: https://grafana.com/docs/grafana/latest/ (accessed 2026-04-30).
- Loki: https://grafana.com/docs/loki/latest/ (accessed 2026-04-30).
- Tempo: https://grafana.com/docs/tempo/latest/ (accessed 2026-04-30).
- NATS: https://nats.io/ (accessed 2026-04-30).

### 11.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.
- §4 metric inventory + §5 NATS subject inventory operationalise R-08 at the per-submodule level (cross-references S05 §9.5 + the 6 NATS subject patterns).

### 11.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `08_Operations/04_Observability_and_Events.md` — 2026-04-30.
