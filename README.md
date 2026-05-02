# HelixPlay — Ultimate Gaming Experience

> **Transform any GPU-equipped machine into a remote gaming appliance.** Stream console-class experiences to any device with ≤30ms LAN / ≤50ms WAN p999 latency. Fully self-hostable, fully open, white-labellable for partners.

[![Constitution](https://img.shields.io/badge/Constitution-v2.2.0-blue)](docs/research/chapters/MVP/05_Response/01_Constitution.md)
[![Go Version](https://img.shields.io/badge/Go-1.26.2-blue)](go.mod)
[![Submodules](https://img.shields.io/badge/Submodules-46-blue)](.gitmodules)
[![Test Matrix](https://img.shields.io/badge/Tests-1%2C840%20cells-green)](specs/001-helixplay-system/spec.md)

---

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Architecture](#architecture)
4. [Technology Stack](#technology-stack)
5. [Project Structure](#project-structure)
6. [Build & Test](#build--test)
7. [Containers](#containers)
8. [Testing Strategy](#testing-strategy)
9. [Security](#security)
10. [Contributing](#contributing)
11. [License](#license)
12. [Acknowledgments](#acknowledgments)
13. [Support](#support)

---

## 1. Overview

HelixPlay is a cloud-gaming platform built as a Go-centric monorepo with 46 Git submodules. It turns any gaming PC into a streaming host, delivering PS4 Pro-class UX to desktop, mobile, TV, and browser clients.

**Key differentiators:**
- **Triple-stack client convergence**: Wails (desktop) + Flutter (mobile/TV) + Angular (web) share one Go core
- **Decoupled architecture**: 46 reusable submodules under `vasic-digital` and `HelixDevelopment`
- **Anti-bluff enforcement**: Green tests guarantee real end-user-usable behavior (Constitution §1)
- **White-label SaaS**: Per-tenant theming, catalog filtering, OAuth2, and billing
- **Container-native**: Every service, DB, build, test, and scan runs inside containers

---

## 2. Quick Start

### Prerequisites
- Go 1.26.2+
- Docker / Podman
- NVIDIA/AMD/Intel GPU with latest drivers (host)
- DualSense controller (for full experience)

### Clone
```bash
git clone --recurse-submodules git@github.com:HelixDevelopment/HelixPlay.git
cd HelixPlay
```

### Build
```bash
make build
```

### Test (all 10 types)
```bash
make test-fullauto
```

### Run Local Topology
```bash
make docker-up
```

---

## 3. Architecture

### Three Runtime Domains

```
┌─────────────────┐     gRPC/REST      ┌─────────────────┐
│  Triple-Stack   │◄──────────────────►│  Core Backend   │
│    Clients      │   WebRTC/QUIC/UDP  │   (cmd/core/)   │
│  (Wails/Flutter/│◄──────────────────►│                 │
│   Angular+WASM) │                    │  Session/Tenant/│
└─────────────────┘                    │  Catalog/Auth   │
         ▲                             └────────┬────────┘
         │                                      │
         │         mDNS / Rendezvous            │
         └──────────────────────────────────────┘
                          │
                    ┌─────┴─────┐
                    │ Host Agent │
                    │(cmd/host- │
                    │  agent/)   │
                    │ Capture /  │
                    │ Encode /   │
                    │ Transport  │
                    └────────────┘
```

### Data Flow
```
Capture → Encode → Packetize → Transmit → Decode → Render
  (DXGI/   (NVENC/   (RTP/      (WebRTC/   (WebCodecs/ (OpenGL/
   SCK/     QSV/      QUIC/      UDP)       native)     Metal/
   PipeWire) AMF/VT)   datagram)
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for C4 diagrams.

---

## 4. Technology Stack

| Layer | Technology |
|-------|-----------|
| **Language** | Go 1.26.2 (root), 1.25+ (submodules) |
| **Desktop Client** | Wails v2 (Go + embedded webview) |
| **Mobile/TV Client** | Flutter 3.29+ (FFI to Go core) |
| **Web Client** | Angular 17+ (Go compiled to WASM) |
| **Streaming** | WebRTC Pion v4, QUIC/quic-go, custom UDP |
| **Database** | CockroachDB v24.2+ (primary), Redis (cache) |
| **Messaging** | NATS JetStream, RabbitMQ (billing) |
| **Observability** | slog, Prometheus, OpenTelemetry |
| **Auth** | OAuth2/OIDC via Auth0, mTLS between services |
| **Containers** | Docker (dev), Podman rootless (prod) |

---

## 5. Project Structure

```
HelixPlay/
├── cmd/
│   ├── host-agent/       # Sunshine++ host agent
│   ├── core/             # Backend services (session, tenant, catalog)
│   ├── client-wails/     # Desktop client
│   ├── client-web/       # Web/TV client
│   └── client-flutter/   # Mobile/TV client
├── pkg/
│   ├── protocol/         # Protobuf schemas
│   ├── models/           # Shared Go models
│   ├── repository/       # DB repository interfaces
│   ├── core/             # Shared Go core (3 compilation targets)
│   ├── catalog/          # Catalogizer API client
│   └── streaming/        # Streaming abstractions
├── tests/
│   ├── unit/             # Mocks allowed
│   ├── integration/      # Real deps, containers
│   ├── e2e/              # Full topology
│   ├── security/         # govulncheck, fuzz, pen-test
│   ├── benchmark/        # p999 HDR histograms
│   ├── chaos/            # Toxiproxy, chaos-mesh
│   ├── stress/           # 24-hour soak
│   ├── smoke/            # 30-second post-deploy
│   ├── fullauto/         # Orchestrates all above
│   └── challenges/       # vasic-digital/Challenges
├── scripts/
│   ├── anti-bluff-scan.sh      # Non-overridable CI lane
│   ├── fault-inject.sh         # Negative-leg fault injection
│   ├── mutation-test.sh        # Gremlins ≥85% gate
│   ├── latency-benchmark.sh    # HDR histogram output
│   ├── propagate-constitution.sh
│   └── verify-submodules.py
├── specs/001-helixplay-system/  # Feature specification
│   ├── spec.md
│   ├── plan.md
│   ├── tasks.md
│   ├── data-model.md
│   ├── research.md
│   └── contracts/
└── <46 submodules>/       # Reusable infrastructure
```

---

## 6. Build & Test

```bash
# Build everything
make build

# Run all 10 test types
make test-fullauto

# Individual test types
make test-unit          # Unit tests (mocks allowed)
make test-integration   # Integration tests (real deps)
make test-e2e           # End-to-end tests
make test-security      # govulncheck + Snyk + Trivy
make test-bench         # Benchmarks (p50/p99/p999)
make test-chaos         # Chaos engineering
make test-stress        # 24-hour soak
make test-smoke         # 30-second post-deploy
make test-challenge     # Challenge scenarios

# Anti-bluff scan (non-overridable)
make anti-bluff

# Format, vet, lint
make fmt vet lint

# Coverage report
make test-coverage
```

---

## 7. Containers

All services run in containers via `vasic-digital/Containers`:

```bash
# Build service images
make docker-build

# Start full local topology
make docker-up

# Stop topology
make docker-down
```

Base images:
- **Distroless** (`gcr.io/distroless/static:nonroot`) — Go services
- **Alpine** — Debug sidecars, CI builders
- **Ubuntu 24.04 LTS** — GPU-enabled host agent

---

## 8. Testing Strategy

**1,840 test matrix cells**: 46 submodules × 10 test types × 4 CI runners.

| Type | Mocks | Frequency | Tools |
|------|-------|-----------|-------|
| Unit | ✅ Yes | Per PR | go test, testify |
| Integration | ❌ No | Per PR + Nightly | testcontainers-go |
| E2E | ❌ No | Per PR + Nightly | Playwright Go, HelixQA |
| Security | N/A | Per PR + Monthly | govulncheck, Snyk, Trivy, fuzz |
| Benchmark | ❌ No | Nightly | hdrhistogram-go, benchstat |
| Chaos | ❌ No | Nightly | Toxiproxy, chaos-mesh |
| Stress | ❌ No | Canary | 24-hour soak runner |
| Smoke | ❌ No | Post-deploy | helm test, curl |
| Full Automation | ❌ No | Per PR + Nightly | Dagger Go SDK |
| Challenges | ❌ No | Nightly | vasic-digital/Challenges |

**Anti-bluff enforcement**:
- `anti-bluff-scan.sh` runs on every commit (non-overridable)
- Mutation testing ≥85% (Gremlins)
- Observable assertion ratio ≥60%
- Negative-leg fault injection per feature
- HelixQA autonomous pre-release sign-off

---

## 9. Security

- **OAuth2/OIDC** with Auth0 (device authorization grant RFC 8628 for TV)
- **mTLS** between all services
- **RBAC** with scope/role validation
- **Container isolation** — no privileged mounts
- **Row-level security** in CockroachDB per tenant
- **CVE scanning** — govulncheck, Snyk, Trivy
- **Fuzz testing** — protocol parsers and input handlers
- **Operational Integrity (R-18)** — no command may suspend/hibernate/lock/terminate host

See [Security Policy](.github/SECURITY.md).

---

## 10. Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for:
- PR template and commit conventions
- Branch strategy (`main`, `001-helixplay-system` feature branch)
- Code review requirements
- Anti-bluff pledge

**Quick checklist before PR:**
- [ ] `make anti-bluff` passes
- [ ] `make test-unit` passes with ≥95% coverage
- [ ] No `TODO`/`FIXME`/`placeholder` in code
- [ ] Tests verify observable behavior (not just nil checks)
- [ ] Documentation updated if interfaces changed

---

## 11. License

[License TBD — pending legal review]

---

## 12. Acknowledgments

- Constitution v2.2.0: [docs/research/chapters/MVP/05_Response/01_Constitution.md](docs/research/chapters/MVP/05_Response/01_Constitution.md)
- System Overview: [docs/research/chapters/MVP/05_Response/02_System_Overview.md](docs/research/chapters/MVP/05_Response/02_System_Overview.md)
- 36,815 lines of research across three MVP streams

---

## 13. Support

- **Issues**: [GitHub Issues](https://github.com/HelixDevelopment/HelixPlay/issues)
- **Discussions**: [GitHub Discussions](https://github.com/HelixDevelopment/HelixPlay/discussions)
- **Email**: support@helixplay.dev

---

*HelixPlay — "Ultimate gaming experience!"*
