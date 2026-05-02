# Implementation Plan: HelixPlay — Ultimate Gaming Experience

**Branch**: `001-helixplay-system` | **Date**: 2026-05-02 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/001-helixplay-system/spec.md`  
**Constitution**: v2.2.0 | **Submodules**: 46 | **Test Matrix**: 1,840 cells

---

## Summary

HelixPlay transforms any GPU-equipped machine into a remote gaming appliance, streaming console-class experiences to any client device. This plan covers the full implementation from foundation (P00) through general availability (P13), with 14 phases, ~200 fine-grained tasks, and enforcement of the Anti-Bluff Pledge (Constitution §1) at every gate.

**Primary approach:** TDD for all code; container-native for all runtime; decoupled submodule architecture; triple-stack client convergence (Wails/Flutter/Angular sharing one Go core); autonomous QA via HelixQA + Challenges.

---

## Technical Context

**Language/Version**: Go 1.26.2 (root), 1.25+ (submodules)  
**Primary Dependencies**: Wails v2, Flutter 3.x+, Angular 17+, WebRTC Pion v4, quic-go, NATS, Redis, RabbitMQ, CockroachDB, Prometheus, OpenTelemetry, Auth0  
**Storage**: CockroachDB (primary), Redis (cache/sessions), S3 + CloudFront (assets), NVMe (recordings)  
**Testing**: `go test` (all 10 types), `go-mutesting` (mutation ≥85%), `golangci-lint`, `govulncheck`, HelixQA autonomous  
**Target Platform**: Linux 6.3+ / Windows 10+ / macOS 14+ (host); iOS 15+ / Android 10+ / tvOS 15+ / any modern browser (client)  
**Project Type**: Cloud gaming platform — host agent + core backend + triple-stack clients + white-label SaaS  
**Performance Goals**: ≤30ms LAN p999 / ≤50ms WAN p999 glass-to-glass; 1kHz controller polling; 4K HDR @ 120fps  
**Constraints**: All services/DBs/builds/tests in containers; no `TODO`/`FIXME`/placeholders; mocks only in Unit tests; mutation score ≥85%; observable assertion ratio ≥60%  
**Scale/Scope**: 1,000+ concurrent sessions; 100+ white-label tenants; 46 submodules; 1,840 test matrix cells

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Justification |
|------|--------|---------------|
| R-01 Anti-bluff | ✅ PASS | Spec mandates observable behaviour assertions, mutation score ≥85%, usability evidence per feature. |
| R-02 No forbidden patterns | ✅ PASS | `anti-bluff-scan.sh` CI lane configured; forbidden patterns table in spec §1.1.1. |
| R-03 SIV Versioning | ✅ PASS | FR-045 mandates `/vN` suffix for all public submodules. |
| R-04 go.work | ⚠️ PARTIAL | FR-044 mandates root `go.work`; currently missing — tracked as P00.T03. |
| R-05 Containers only | ✅ PASS | FR-021, ASM-008 mandate container runtime for all services. |
| R-06 CI/CD | ⚠️ PARTIAL | Task 15.2 defines CI/CD; currently no GitHub Actions — tracked as P01. |
| R-07 Dynamic ports | ✅ PASS | FR-022, Discovery Beacon design covers mDNS + rendezvous. |
| R-08 Messaging | ✅ PASS | NATS/Redis/RabbitMQ selection documented (ASM-009). |
| R-09 Non-blocking | ✅ PASS | FR-020 mandates lazy init + semaphores; SPSC ring buffer design covers hot path. |
| R-10 Quality gates | ⚠️ PARTIAL | SonarQube/Snyk defined but not deployed — tracked as P11. |
| R-11 Ten test types | ✅ PASS | All 10 types defined in spec §Testing & Quality; test matrix 1,840 cells. |
| R-12 Mocks only in Unit | ✅ PASS | FR-030 explicitly restricts mocks to Unit; 9 files flagged for remediation in audit. |
| R-13 Anti-bluff enforcement | ✅ PASS | `anti-bluff-scan` non-overridable; `ValidateAntiBluff` unconditional. |
| R-14 Observability | ✅ PASS | Four-signal stack defined: slog, Prometheus, OpenTelemetry, NATS events. |
| R-15 Transitive deps | ✅ PASS | 46 submodules verified in `.gitmodules`; Constitution propagated to all. |
| R-16 Security | ✅ PASS | OAuth2/OIDC, mTLS, RBAC, container isolation all defined. |
| R-17 Documentation | ✅ PASS | Living docs mandate; 36,815 lines of research absorbed; no drift >30 days. |
| R-18 Operational Integrity | ✅ PASS | Forbidden commands list, container hazards inventory, `r18.SafeExec` wrapper defined. |

**Re-evaluation after Phase 1:** All gates remain valid. P00.T03 (go.work) and P01 (CI/CD) are planned work, not violations.

---

## Project Structure

### Source Code (repository root)

```text
cmd/
├── host-agent/          # Sunshine++ host agent (capture, encode, input, transport)
│   ├── capability/      # GPU capability advertisement
│   ├── capture/         # Per-OS capture (DXGI, ScreenCaptureKit, PipeWire)
│   ├── codec/           # Codec negotiation and hardware encoder bindings
│   ├── discovery/       # mDNS + rendezvous beacon
│   ├── encoder/         # NVENC/QSV/AMF/VideoToolbox/VAAPI integration
│   ├── game/            # Game enumeration (Steam/Epic/GOG/etc.)
│   ├── input/           # 1kHz USB polling, DualSense forwarding
│   ├── lifecycle/       # Launch, monitor, terminate, quick resume
│   └── transport/       # WebRTC/QUIC/UDP session management
├── core/                # Backend services (currently EMPTY — P03)
│   ├── api/             # gRPC/REST gateway
│   ├── session/         # Session orchestration
│   ├── tenant/          # Multi-tenant management
│   └── discovery/       # Host fleet registry
├── client-wails/        # Desktop client (Go backend + web frontend)
│   ├── backend/         # Go business logic (shared core)
│   └── frontend/        # HTML/JS/CSS UI
├── client-web/          # Web/TV client
│   ├── tv/              # 10-foot UI with leanback navigation
│   └── wasm/            # Go core compiled to WebAssembly
└── client-flutter/      # Mobile/TV client (FFI to Go core)
    ├── lib/             # Dart UI layer
    └── go_core/         # c-shared Go library

pkg/
├── protocol/            # Shared protocol definitions (protobuf)
├── streaming/           # Shared streaming abstractions
└── catalog/             # Catalog client (decoupled from Catalogizer)

tests/
├── unit/                # Unit tests (mocks allowed)
├── integration/         # Integration tests (real deps, containers)
├── e2e/                 # End-to-end tests (full topology)
├── security/            # govulncheck, Snyk, Trivy, fuzz
├── benchmark/           # p999 latency, HDR histogram
├── chaos/               # Toxiproxy, chaos-mesh
├── stress/              # 24-hour soak
├── smoke/               # 30-second post-deploy
├── fullauto/            # Orchestrates all above
└── challenges/          # vasic-digital/Challenges integration

scripts/
├── anti-bluff-scan.sh   # Non-overridable CI lane
├── propagate-constitution.sh
├── verify-submodules.py
└── claim-check.sh       # Session stop hook

.submodules/ (46 total)
├── vasic-digital/       # Reusable infrastructure
│   ├── Auth, Cache, Challenges, Concurrency, Containers, Database
│   ├── Discovery, EventBus, Formatters, Media, Memory, Messaging
│   ├── Middleware, Observability, Plugins, RAG, RateLimiter
│   ├── Recovery, Security, Storage, Streaming, VectorDB
│   ├── Assets, Auth-Context-React, Catalogizer-API-Client-TS
│   ├── Collection-Manager-React, Config, Dashboard-Analytics-React
│   ├── DocProcessor, Entities, Filesystem, Lazy, LLMOrchestrator
│   ├── LLMProvider, Media-Browser-React, Media-Player-React
│   ├── Media-Types-TS, ReplayBuffer, ScreenDiff, TrainingCollector
│   ├── UI-Components-React, VisionEngine, VisualRegression
│   ├── WebSocket-Client-TS, Watcher
│   └── Catalogizer (decoupled standalone)
└── HelixDevelopment/
    └── HelixQA          # Autonomous QA orchestration
```

---

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 46 submodules | Complete ecosystem coverage; each submodule is independently reusable | Fewer submodules would force unrelated code together, violating decoupling (R-03) |
| Triple-stack clients | Market reach requires desktop + mobile + web | Single client would exclude platforms; native per-platform would duplicate business logic |
| Dual-path encoding | Stream latency + record quality are mutually exclusive optimizations | Single encoder cannot simultaneously achieve ≤5ms encode and archival quality |
| 10 test types | Anti-bluff mandate requires real-system validation at every layer | Fewer types would allow mock-only "green tests on broken features" |
| White-label multi-tenancy | B2B revenue model requires partner isolation | Single-tenant would prevent partner resale, blocking business model |

---

## Implementation Phases

### Phase P00: Foundation

**Objective:** Root infrastructure, Constitution propagation, submodule verification.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P00.T01 | Create root README.md (13 sections, >200 lines) | P1 | R-17 | 2h |
| P00.T02 | Create root Makefile (all 10 test types + build + lint) | P1 | R-09 | 3h |
| P00.T03 | Create root go.work + go work sync | P1 | R-04 | 2h |
| P00.T04 | Verify all 46 submodule dependencies transitively complete | P1 | R-15 | 1h |
| P00.T05 | Propagate Constitution v2.2.0 to all 46 submodules | P1 | R-15 | 2h |
| P00.T06 | Create GitHub Projects + GitLab Issues boards | P2 | R-12 | 1h |
| P00.T07 | Verify `scripts/` directory completeness | P2 | R-17 | 1h |

**Exit Criteria:** `make build` succeeds from root; `go work sync` clean; all 46 submodules have CLAUDE.md + AGENTS.md + CONSTITUTION.md.

---

### Phase P01: Containers & CI

**Objective:** Container-native runtime and CI/CD pipeline.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P01.T01 | Bootstrap `vasic-digital/Containers` — host agent container definition | P1 | R-05 | 4h |
| P01.T02 | Bootstrap capture service container (per-OS variants) | P1 | R-05 | 4h |
| P01.T03 | Bootstrap encoder service container (GPU passthrough) | P1 | R-05 | 4h |
| P01.T04 | Bootstrap discovery beacon container | P1 | R-05 | 2h |
| P01.T05 | Create GitHub Actions workflow — `anti-bluff-scan` (non-overridable) | P1 | R-13 | 3h |
| P01.T06 | Create GitHub Actions workflow — test matrix (Unit + Integration + E2E) | P1 | R-11 | 3h |
| P01.T07 | Create GitHub Actions workflow — security scan (govulncheck + Snyk + Trivy) | P1 | R-10 | 2h |
| P01.T08 | Create GitHub Actions workflow — mutation testing (go-mutesting ≥85%) | P1 | §6.4 | 3h |
| P01.T09 | Configure four-mirror push (GitHub fetch, GitFlic push) | P2 | R-13 | 1h |

**Exit Criteria:** `docker compose up -d` boots full topology; CI passes on `main` push.

---

### Phase P02: Core Submodules

**Objective:** Implement foundational reusable submodules.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P02.T01 | `vasic-digital/Memory` — shared memory, `memfd_create`, lock-free SPSC | P1 | R-09 | 8h |
| P02.T02 | `vasic-digital/Discovery` — mDNS + rendezvous service | P1 | R-07 | 6h |
| P02.T03 | `vasic-digital/Auth` — OAuth2/OIDC middleware, JWT validation | P1 | R-16 | 6h |
| P02.T04 | `vasic-digital/Database` — CockroachDB abstraction, migrations | P1 | R-04 | 4h |
| P02.T05 | `vasic-digital/Cache` — Redis client with circuit breaker | P1 | R-08 | 4h |
| P02.T06 | `vasic-digital/EventBus` — NATS pub/sub wrapper | P1 | R-08 | 4h |
| P02.T07 | `vasic-digital/Concurrency` — non-blocking primitives, semaphores | P1 | R-09 | 4h |
| P02.T08 | `vasic-digital/Observability` — slog, Prometheus, OpenTelemetry | P1 | R-14 | 6h |
| P02.T09 | `vasic-digital/Security` — RBAC, input validation, CVE scan | P1 | R-16 | 4h |

**Exit Criteria:** Each submodule passes `make test` (Unit + Integration) with ≥95% coverage.

---

### Phase P03: Backend Services (`cmd/core/`)

**Objective:** Implement core backend — currently empty.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P03.T01 | Session service — create, negotiate, monitor, terminate sessions | P1 | FR-001 | 12h |
| P03.T02 | Host registry — fleet management, capability advertisement | P1 | FR-022 | 8h |
| P03.T03 | Tenant service — CRUD, theming, isolation | P1 | FR-024 | 8h |
| P03.T04 | User service — OAuth2 integration, RBAC enforcement | P1 | FR-017 | 6h |
| P03.T05 | Catalog proxy — decoupled Catalogizer API client | P1 | FR-026 | 6h |
| P03.T06 | gRPC gateway — public API with mTLS | P1 | FR-018 | 6h |
| P03.T07 | REST gateway — separate microservice for web clients | P1 | FR-018 | 4h |
| P03.T08 | Load balancer — session distribution across hosts | P2 | FR-023 | 6h |

**Exit Criteria:** `make test-e2e` passes with real containers; p999 latency ≤50ms in benchmark.

---

### Phase P04: Streaming Pipeline

**Objective:** Capture, encode, transport — the heart of the system.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P04.T01 | Per-OS capture: DXGI (Windows) | P1 | FR-005 | 12h |
| P04.T02 | Per-OS capture: ScreenCaptureKit (macOS) | P1 | FR-005 | 10h |
| P04.T03 | Per-OS capture: PipeWire/KMS (Linux) | P1 | FR-005 | 10h |
| P04.T04 | Hardware encoder: NVENC integration | P1 | FR-010 | 12h |
| P04.T05 | Hardware encoder: QSV integration | P1 | FR-010 | 8h |
| P04.T06 | Hardware encoder: AMF integration | P1 | FR-010 | 8h |
| P04.T07 | Hardware encoder: VideoToolbox integration | P1 | FR-010 | 8h |
| P04.T08 | Hardware encoder: VAAPI integration | P1 | FR-010 | 6h |
| P04.T09 | Codec ladder: H.264 fallback | P1 | FR-002 | 6h |
| P04.T10 | Codec ladder: HEVC Main 10 | P1 | FR-002 | 6h |
| P04.T11 | Codec ladder: AV1 Main | P1 | FR-002 | 8h |
| P04.T12 | WebRTC Pion v4 transport | P1 | FR-003 | 12h |
| P04.T13 | QUIC datagram transport | P1 | FR-003 | 8h |
| P04.T14 | Custom UDP transport (Parsec BUD style) | P1 | FR-003 | 8h |
| P04.T15 | ABR/FEC/SQP congestion control | P1 | FR-003 | 10h |
| P04.T16 | Capability handshake (codec + transport negotiation) | P1 | FR-004 | 6h |

**Exit Criteria:** End-to-end streaming test passes at 1080p60 with ≤30ms LAN p999.

---

### Phase P05: Triple-Stack Clients

**Objective:** Wails desktop, Flutter mobile/TV, Angular web.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P05.T01 | Go core — shared business logic, gRPC client, protocol negotiation | P1 | FR-015 | 12h |
| P05.T02 | Wails backend — real implementation (replace stub booleans) | P1 | FR-015 | 8h |
| P05.T03 | Wails frontend — desktop UI, native OS integration | P1 | FR-015 | 10h |
| P05.T04 | Flutter FFI — c-shared Go core binding | P1 | FR-015 | 10h |
| P05.T05 | Flutter UI — mobile navigation, controller support | P1 | FR-015 | 10h |
| P05.T06 | Angular web — WASM compilation, WebCodecs decode | P1 | FR-015 | 10h |
| P05.T07 | Client auto-detection — codec/transport selection | P1 | FR-004 | 4h |
| P05.T08 | Cross-client E2E test suite | P1 | SC-003 | 6h |

**Exit Criteria:** All three clients connect to host and stream with same latency budget.

---

### Phase P06: Host Agent Integration

**Objective:** Game enumeration, lifecycle, quick resume.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P06.T01 | Game enumeration — Steam/Epic/GOG store APIs | P1 | FR-001 | 8h |
| P06.T02 | Game enumeration — Ubisoft/Battle.net/Origin/Microsoft | P1 | FR-001 | 8h |
| P06.T03 | Capability metadata publication | P1 | FR-001 | 4h |
| P06.T04 | Game launch — process spawn, window capture attach | P1 | FR-001 | 6h |
| P06.T05 | Game monitor — health checks, crash detection | P1 | FR-001 | 4h |
| P06.T06 | Game terminate — graceful shutdown, state preservation | P1 | FR-001 | 4h |
| P06.T07 | Quick resume — save state caching, ≤5s restore | P1 | SC-007 | 8h |
| P06.T08 | Controller hot-plug — detection, renegotiation | P1 | FR-009 | 6h |

**Exit Criteria:** Host agent discovers and streams real games end-to-end (KSC-04).

---

### Phase P07: Latency Optimization

**Objective:** Achieve ≤30ms LAN / ≤50ms WAN p999.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P07.T01 | Shared memory + zero-copy IPC (`memfd_create`, SPSC) | P1 | FR-007 | 8h |
| P07.T02 | io_uring kernel bypass | P1 | FR-007 | 8h |
| P07.T03 | GPU-Direct + zero-copy texture sharing | P1 | FR-007 | 8h |
| P07.T04 | Real-time OS scheduling (PREEMPT_RT, CPU isolation) | P1 | FR-007 | 6h |
| P07.T05 | Frame pacing + VRR integration | P1 | FR-007 | 6h |
| P07.T06 | p50/p99/p999 measurement (HDR histogram, ≥10K samples) | P1 | SC-001 | 4h |
| P07.T07 | Latency regression CI gate (>150% baseline blocks merge) | P1 | §19.2 | 4h |

**Exit Criteria:** Benchmark reports p999 ≤30ms LAN / ≤50ms WAN over ≥10K samples.

---

### Phase P08: Audio Surround

**Objective:** Multi-channel audio pipeline.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P08.T01 | Opus MultiStream encoder/decoder | P2 | FR-013 | 6h |
| P08.T02 | AC3/EAC3 passthrough | P2 | FR-013 | 4h |
| P08.T03 | Dolby Atmos forwarding | P2 | FR-013 | 6h |
| P08.T04 | eARC interaction | P2 | FR-013 | 4h |

**Exit Criteria:** Audio latency ≤1ms effective; surround formats verified on reference hardware.

---

### Phase P09: Recording & Replay

**Objective:** DVR for gaming PC — dual-path encoding.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P09.T01 | Dual-path encoder — stream path (latency) + record path (quality) | P2 | FR-011 | 8h |
| P09.T02 | MKV container writer | P2 | FR-014 | 4h |
| P09.T03 | fMP4 container writer | P2 | FR-014 | 4h |
| P09.T04 | Background sync — upload to user storage | P2 | FR-014 | 6h |
| P09.T05 | HDR metadata preservation in recordings | P2 | FR-012 | 4h |

**Exit Criteria:** Recording plays back with correct quality; sync doesn't impact streaming latency.

---

### Phase P10: Monetization & Auth

**Objective:** White-label billing and partner onboarding.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P10.T01 | `vasic-digital/Monetization` — subscription engine | P2 | FR-024 | 10h |
| P10.T02 | `vasic-digital/Monetization` — usage tracking | P2 | FR-024 | 6h |
| P10.T03 | `vasic-digital/Monetization` — invoice generation | P2 | FR-024 | 6h |
| P10.T04 | `vasic-digital/Monetization` — revenue-share ledger | P2 | FR-024 | 6h |
| P10.T05 | Partner onboarding API — tenant provisioning | P2 | FR-024 | 6h |
| P10.T06 | Per-tenant OAuth2 config (Auth0 actions) | P2 | FR-017 | 4h |

**Exit Criteria:** Two demo tenants bill independently; revenue reports accurate.

---

### Phase P11: Hardening & Security

**Objective:** Penetration tests, chaos engineering, fault injection.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P11.T01 | Container isolation verification — privilege escalation attempts | P2 | FR-034 | 6h |
| P11.T02 | Penetration test — OWASP Top 10 | P2 | R-16 | 8h |
| P11.T03 | Fuzz testing — protocol parsers, input handlers | P2 | R-11 | 6h |
| P11.T04 | Chaos engineering — Toxiproxy network partitions | P2 | R-11 | 4h |
| P11.T05 | Chaos engineering — chaos-mesh pod failures | P2 | R-11 | 4h |
| P11.T06 | SonarQube deployment + quality gate | P2 | R-10 | 4h |
| P11.T07 | Snyk + Trivy CI integration | P2 | R-10 | 2h |

**Exit Criteria:** All security tests pass; no CRITICAL or HIGH vulnerabilities unremediated.

---

### Phase P12: Beta Launch

**Objective:** Closed beta with 100 users, 10 hosts, 2 partners.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P12.T01 | Beta signup system — invite codes, waitlist | P2 | R-12 | 4h |
| P12.T02 | Telemetry — anonymized latency, error reporting | P2 | R-14 | 4h |
| P12.T03 | Feedback loop — in-app bug reporting | P2 | R-12 | 4h |
| P12.T04 | Load test — 100 concurrent sessions | P2 | SC-008 | 6h |
| P12.T05 | Partner onboarding — 2 pilot white-label tenants | P2 | SC-006 | 8h |

**Exit Criteria:** 100 beta users with ≥90% session success rate; partner tenants operational.

---

### Phase P13: GA Release

**Objective:** Public launch with full feature set.

| Task ID | Description | Priority | Constitution | Effort |
|---------|-------------|----------|--------------|--------|
| P13.T01 | Performance audit — all p999 targets verified | P1 | SC-001 | 4h |
| P13.T02 | Security audit — third-party penetration test | P1 | R-16 | 8h |
| P13.T03 | Documentation — API reference, deployment guide | P1 | R-17 | 6h |
| P13.T04 | HelixQA autonomous sign-off | P1 | SC-004 | 4h |
| P13.T05 | Anti-bluff final scan — zero violations | P1 | R-13 | 2h |
| P13.T06 | Release tags — all 46 submodules at v1.0.0+ | P1 | R-03 | 2h |

**Exit Criteria:** All Key Success Criteria (KSC-01 through KSC-08) satisfied with evidence.

---

## Phase Reference: Existing Detailed Plans

For task-level implementation details beyond the summaries above, refer to:

- `docs/plans/mvp/Full_Implementation_Plan/helixplay_implementation_plan.md` (10,482 lines) — P00-P13 full task breakdowns
- `docs/plans/mvp/Full_Implementation_Plan/section_testing_strategy.md` (3,857 lines) — Test matrix, anti-bluff infrastructure, HelixQA integration
- `docs/plans/mvp/Full_Implementation_Plan/section_advanced_phases.md` (3,198 lines) — P07-P13 advanced topics
- `docs/plans/mvp/Full_Implementation_Plan/section_main_body.md` (3,391 lines) — Architecture deep-dive

---

## Artifact Checklist

| Artifact | Status | Path |
|----------|--------|------|
| plan.md | ✅ This file | `specs/001-helixplay-system/plan.md` |
| research.md | ✅ Complete | `specs/001-helixplay-system/research.md` |
| data-model.md | ✅ Complete | `specs/001-helixplay-system/data-model.md` |
| quickstart.md | ✅ Complete | `specs/001-helixplay-system/quickstart.md` |
| contracts/ | ✅ Complete | `specs/001-helixplay-system/contracts/` |
| tasks.md | ⏳ Next: `/speckit-tasks` | `specs/001-helixplay-system/tasks.md` |

---

*End of Implementation Plan — Phase 0 (Research) and Phase 1 (Design) complete.*
