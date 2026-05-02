% HelixPlay Cloud Gaming Platform
% Comprehensive Implementation Plan
% Version 2.1.0 | 2026-05-02

---

# Table of Contents

- [Section 1: Executive Summary](#section-1-executive-summary)
- [Section 2: Constitutional Foundation (NON-NEGOTIABLE)](#section-2-constitutional-foundation)
- [Section 3: Architecture Overview](#section-3-architecture-overview)
- [Section 4: Implementation Phases P00-P06 (Foundation & Core)](#section-4-implementation-phases-p00-p06)
- [Section 5: Implementation Phases P07-P13 (Advanced & GA)](#section-5-implementation-phases-p07-p13)
- [Section 6: Architecture Deep-Dive Implementation Tasks](#section-6-architecture-deep-dive)
- [Section 7: Testing Philosophy & Anti-Bluff Constitution](#section-7-testing-philosophy)
- [Section 8: The Ten Test Types](#section-8-ten-test-types)
- [Section 9: Test Matrix](#section-9-test-matrix)
- [Section 10: Anti-Bluff Infrastructure](#section-10-anti-bluff-infrastructure)
- [Section 11: HelixQA Autonomous QA Integration](#section-11-helixqa)
- [Section 12: CI/CD Pipeline Design](#section-12-cicd)
- [Section 13: Fixing Current Issues](#section-13-fixing-issues)
- [Section 14: Challenges Implementation Plan](#section-14-challenges)
- [Appendices](#appendices)

---

# HelixPlay Cloud Gaming Platform — Comprehensive Implementation Plan
## MAIN BODY (Sections 1–4: Executive Summary through Implementation Phases P00–P06)

> **Document Status:** DRAFT — awaiting integration with appendices
> **Source of Truth:** `docs/research/chapters/MVP/05_Response/` (142 files, ~47,000 lines)
> **Constitution Version:** v2.1.0
> **Last Updated:** 2026-05-02

---

## Section 1: Executive Summary

### 1.1 Project Overview and Vision

**HelixPlay** is a cloud gaming platform that transforms any GPU-equipped machine into a remote gaming appliance, streaming console-class gaming experiences to any client device. The project's guiding vision is **"Ultimate gaming experience!"** — delivering PS4 Pro-class UX with zero perceived lag, fully self-hostable, fully open-source, and white-labellable for partners.

**Core value propositions:**

| # | Value Proposition | Technical Enabler |
|---|-------------------|-------------------|
| 1 | **Zero perceived lag** | <=50ms WAN / <=30ms LAN at p999 (not p50) |
| 2 | **Full controller fidelity** | 1kHz USB polling, DualSense haptics + adaptive triggers + gyro + accelerometer |
| 3 | **Triple-stack client convergence** | Single Go core compiled three ways: c-shared (Flutter), native (Wails), WASM (Angular) |
| 4 | **Any GPU becomes a gaming server** | Sunshine++ host agent with per-OS capture (DXGI/ScreenCaptureKit/PipeWire) |
| 5 | **White-label from day one** | Tenant-scoped architecture: identity, catalog, recordings, billing per partner |
| 6 | **Anti-cheat clean host** | OS APIs only — no hooks, no DLL injection, no kernel tampering |
| 7 | **Autonomous QA** | HelixQA orchestrates 1,160 test matrix cells across 29 submodules |

**System boundaries:** HelixPlay comprises three major runtime domains — the **Host Agent** (Sunshine++, running on the gaming PC), the **Core Backend** (discovery, protocol negotiation, session orchestration), and **Triple-Stack Clients** (Wails desktop, Flutter mobile/TV, Angular web with Go-WASM). All runtime code executes inside containers per Constitution Section 3.

### 1.2 Current State Assessment

The project exists in a **documentation-complete, partially-implemented** state:

| Component | Status | Detail |
|-----------|--------|--------|
| **Documentation** | COMPLETE | 47,000+ lines across 142 files in `05_Response/`; Constitution v2.1.0; Master Plan with 18 contract clauses |
| **Host Agent — `cmd/host-agent/`** | 9/9 modules IMPLEMENTED | capability, capture, codec, discovery, encoder, game, input, lifecycle, transport |
| **Client Web/TV — `cmd/client-web/tv/`** | IMPLEMENTED | 10-foot UI with leanback navigation, D-pad control |
| **Client Wails — `cmd/client-wails/`** | STUB | Backend struct with boolean flags only (Start/Stop/Connect/IsRunning/IsConnected); no real Wails integration |
| **Core Backend — `cmd/core/`** | EMPTY | No visible Go source files |
| **README.md (root)** | MISSING | No project README |
| **Root Makefile** | MISSING | No unified build at root; submodules have individual Makefiles |
| **CI/CD** | NOT CONFIGURED | No GitHub Actions workflows; local scripts only (anti-bluff-scan.sh, claim-check.sh) |
| **Containers** | STUB | `vasic-digital/Containers` submodule exists but lacks full container definitions |
| **29 Go submodules** | v0.x.y | All at pre-v1.0.0; dependency resolution issues with `replace` directives |

**Critical gaps identified:**
1. No root-level README.md or Makefile — impedes contributor onboarding
2. No CI/CD pipeline configured — quality gates run locally only
3. cmd/core/ is empty — backend services not started
4. client-wails is a stub — desktop client not functional
5. Submodule `replace` directives break in clean checkouts
6. Constitution not propagated to all 29 submodule roots
7. No `go.work` workspace file for multi-module development

### 1.3 Scope of This Implementation Plan

This document specifies the **complete implementation path from current state to a fully operational Foundation and Core** — specifically Phases P00 through P06. These seven phases cover:

- **P00 Foundation:** Operator infrastructure, tooling, missing root files, constitution propagation
- **P01 Containers & CI:** Container definitions for all services, local CI/CD pipeline, anti-bluff CI lane
- **P02 Core Submodules:** Graduate all 29 submodules to v1.0.0, resolve dependencies, implement missing functionality
- **P03 Backend Services:** Deploy CockroachDB + NATS + Redis + Vault, REST gateway microservice
- **P04 Streaming Pipeline:** helix-pipeline + helix-transport operational, WebRTC (Pion v4), QUIC datagrams
- **P05 Triple-Stack Clients:** Go core shared logic, Wails desktop (replace stub), Flutter mobile/TV, Angular web
- **P06 Host Agent Integration:** Sunshine++ fully operational, game enumeration, discovery beacon, lifecycle management

**Phases P07–P13** (Catalog, White-Label, TV UX polish, Multi-region, Security hardening, Operations, GA) are scoped in the Master Plan but detailed in a subsequent document.

### 1.4 Key Success Criteria

| ID | Criterion | Measurement |
|----|-----------|-------------|
| KSC-01 | All P00–P06 phases complete with green tests | `make test-fullauto` passes |
| KSC-02 | Anti-bluff CI lane runs on every commit | `scripts/anti-bluff-scan.sh` passes with zero violations |
| KSC-03 | 29 submodules at v1.0.0+ with resolved dependencies | `go work sync` succeeds; no `replace` directives needed for public deps |
| KSC-04 | Host agent discovers and streams games end-to-end | HelixQA challenge `host_stream_game` passes |
| KSC-05 | Triple-stack clients connect and display video | Wails + Flutter + Web all pass E2E video decode challenge |
| KSC-06 | p999 latency <=30ms LAN / <=50ms WAN | `helix-bench` benchmark report |
| KSC-07 | Zero forbidden patterns in codebase | `anti-bluff-scan.sh` — no TODO/FIXME/placeholder/dead code |
| KSC-08 | All services run in containers | `docker compose up` brings up full topology |

---

## Section 2: Constitutional Foundation (NON-NEGOTIABLE)

> **Source:** `docs/research/chapters/MVP/05_Response/01_Constitution.md` v2.1.0
> **Status:** BINDING on all contributors (human + AI)

### 2.1 Prime Directive

> **"Green tests MUST guarantee real, end-user-usable behaviour."**

This is the single non-negotiable axiom of the HelixPlay project. If all tests pass, a real user must be able to install the software, connect a controller, launch a game, and play it. No exceptions. No "the test passes but the feature doesn't work yet." No "it's tested but not integrated." No bluffing.

**What this means for every phase of implementation:**

| Phase | Implication |
|-------|-------------|
| P00 | Every tool and script must itself be tested. The README must be verifiable. The Makefile must have a test target. |
| P01 | Container definitions must be validated by actually running containers. The CI pipeline must be tested by running it. |
| P02 | Every submodule must have passing Unit + Integration + Challenge tests before v1.0.0 tag. |
| P03 | Backend services must pass E2E tests that verify real database writes, real message bus delivery, real API responses. |
| P04 | Streaming pipeline must pass tests that verify actual video frames flow from host to client. |
| P05 | Each client must pass tests that verify real video decode and real controller input. |
| P06 | Host agent must pass a Challenge that launches a real game and verifies video output. |

### 2.2 The 18 Contract Clauses (R-01 through R-18)

These clauses are derived from `04_Request.md` and codified in the Constitution. They are **non-negotiable** — no contributor (human or AI) may override them without a technically justified exception request with a fixed expiry date.

| ID | Clause | Constitution Section | Implementation Implication |
|----|--------|---------------------|---------------------------|
| **R-01** | **Anti-Bluff Enforcement** | Section 1 | `ValidateAntiBluff` gate is unconditional. CI deliberately breaks each feature and verifies non-Unit tests fail. Forbidden patterns: `assert.True(t, true)`, constructor-only tests, mock-only Integration/E2E, TODO/FIXME, empty function bodies. |
| **R-02** | **Decoupled Submodule Architecture** | Section 2 | 29 submodules in polyrepo with four-mirror topology. No circular dependencies. Each submodule builds independently. |
| **R-03** | **SIV Versioning** | Section 2 | All submodules use Semantic Import Versioning (`/vN` suffix). No breaking changes within a major version. |
| **R-04** | **go.work Workspace** | Section 2 | Root `go.work` file includes all submodules. `go work sync` resolves dependencies. |
| **R-05** | **Container-First Runtime** | Section 3 | ALL code runs inside containers. No bare-metal execution in CI or production. Host agent container requires `--privileged` for GPU access. |
| **R-06** | **Container Build Matrix** | Section 3 | Every service has a Dockerfile. Multi-stage builds. distroless or debian:bookworm-slim base. NVIDIA runtime for GPU services. |
| **R-07** | **Communication Stack** | Section 4 | gRPC over HTTP/3 (QUIC) for service-to-service. NATS JetStream for async events. Redis for caching/sessions. RabbitMQ for legacy integration. Brotli compression. |
| **R-08** | **Non-Blocking Concurrency** | Section 5 | Lazy initialization is the DEFAULT pattern. Semaphores for backpressure. Zero-allocation hot path. `sync.Pool` for object reuse. |
| **R-09** | **The Ten Test Types** | Section 6 | Unit (mocks allowed), Integration, E2E, Security, Benchmark, Chaos, Stress, Smoke, FullAuto, Challenges. ONLY Unit may use mocks. 100% coverage target across all types combined. |
| **R-10** | **Negative-Leg Fault Injection** | Section 1.3, 6.3 | CI must deliberately break each feature and verify that non-Unit tests fail. No feature ships without a negative-leg test. |
| **R-11** | **Quality Gates** | Section 7 | SonarQube (code quality), Snyk (dependencies), Semgrep (static analysis), Trivy (container scanning), gitleaks (secrets), govulncheck (Go CVEs). All gates must pass. |
| **R-12** | **Dual-Platform Tracking** | Section 8 | GitHub Projects + GitLab Issues. Every task tracked on both platforms. No untracked work. |
| **R-13** | **Four-Mirror Source Control** | Section 9 | GitHub (primary), GitLab, GitVerse, GitFlic. `origin` fetch=github, push=gitflic. All mirrors kept in sync. |
| **R-14** | **Observability Stack** | Section 10 | Structured JSON logs, Prometheus metrics, OpenTelemetry traces, event tracking. Every service emits all four signal types. |
| **R-15** | **Transitive Submodule Completeness** | Section 2 | All 29 submodules must have complete dependency trees. No missing transitive dependencies. `verify-submodules.py` validates. |
| **R-16** | **Security & Privacy** | Section 11 | mTLS between all services. OAuth2/OIDC + Device Authorization Grant (RFC 8628) for auth. RBAC. Audit logging. Anti-cheat clean host (no hooks). |
| **R-17** | **Documentation Discipline** | Section 12 | Living documents only. No simplification. No "etc." — exhaustive enumeration. Every design decision traced to a source chapter. |
| **R-18** | **Operational Integrity (SafeExec)** | Section 11.5 | No command may suspend, hibernate, lock, terminate, or crash the operator's host. `r18.SafeExec` wrapper validates all system commands. `claim-check.sh` runs on every AI session stop. |

### 2.3 Anti-Bluff Pledge: Detailed Enforcement

The anti-bluff pledge is the most distinctive quality feature of the HelixPlay project. It goes far beyond typical testing discipline.

#### 2.3.1 Forbidden Patterns (Constitution Section 1.1)

The following patterns are **forbidden in all code, tests, and documentation**:

| Pattern | Example | Violation |
|---------|---------|-----------|
| Vacuous assertion | `assert.True(t, true)` | Test passes without testing anything |
| Constructor-only test | Tests only verify object creation | No behaviour verified |
| Mock-only Integration/E2E | Integration tests using mocks | Violates R-09 — only Unit may use mocks |
| Empty function body | `func Foo() {}` or `panic("not implemented")` | Dead code shipped to production |
| TODO/FIXME/XXX/HACK | Any occurrence | Unfinished work marked as complete |
| Standalone "tbd" | `// tbd` | Intentional omission |
| Dead code | Unreachable branches, unused variables | Indicates incomplete implementation |

#### 2.3.2 Required Properties (Constitution Section 1.2)

Every test that reports `Status=Passed` MUST satisfy ALL of the following:

1. **RecordedActions non-empty** — The runtime actually executed some observable action
2. **Assertions non-empty** — At least one expectation was checked
3. **At least one assertion has `Passed=true`** — Something was positively confirmed

The `ValidateAntiBluff()` function in `vasic-digital/Challenges/pkg/challenge/antibluff.go` enforces these three rules. It returns `ErrBluffPass` if any rule is violated, causing the CI lane to fail.

#### 2.3.3 Enforcement Mechanisms

| Layer | Mechanism | Responsibility |
|-------|-----------|----------------|
| **CI (non-overridable)** | `scripts/anti-bluff-scan.sh` runs on every PR/commit. Scans for all forbidden patterns. Fails the build on any match. | R-01, R-11 |
| **Challenge framework** | `ValidateAntiBluff()` gate called unconditionally for every Challenge result. | R-01 |
| **Mutation testing** | `go-mutesting` with branch/if, expression/remove, statement/remove mutators. Timeout 60s per mutant. | R-01 |
| **Negative-leg injection** | CI deliberately breaks each feature (e.g., invert a condition) and verifies non-Unit tests fail. | R-10 |
| **HelixQA autonomous** | OpenCV-based visual assertion — verifies the screen actually shows what the test claims. | R-01 |
| **Session stop hook** | `.claude/settings.json` runs `scripts/claim-check.sh` on Stop — prevents forbidden commands. | R-18 |

#### 2.3.4 Automatic Negative-Leg Fault Injection (Constitution v2.1.0 Amendment)

Per the v2.1.0 amendment, CI must perform the following for every feature:

1. Run all tests on the feature — they MUST pass (positive leg)
2. Deliberately introduce a fault (e.g., invert a boolean, remove a handler) — they MUST fail (negative leg)
3. If tests still pass after the fault, the feature is BLUFF — CI fails

This ensures tests are actually testing the feature, not just running vacuously.

### 2.4 Exception Process (Constitution Section 13)

The ONLY valid exceptions to these clauses require:

1. **Technical impossibility proof** — why the clause cannot be satisfied
2. **Fixed expiry date** — when the exception will be resolved
3. **Written approval** — recorded in both GitHub Projects and GitLab Issues
4. **Explicit exception ID** — referenced in code with `// EXCEPTION(<id>): <reason> <expiry>`

No verbal exceptions. No "we'll fix it later." No AI-generated exceptions without human review.

---

## Section 3: Architecture Overview

### 3.1 The Eight Architectural Pillars

> **Source:** `docs/research/chapters/MVP/05_Response/03_Architecture/00_Index.md`

The HelixPlay architecture rests on eight foundational pillars. Every design decision, every line of code, and every test must be traceable to at least one pillar.

| # | Pillar | Description | Key Decisions |
|---|--------|-------------|---------------|
| **P1** | **Sunshine++ Host Agent** | Evolution of open-source Sunshine. Captures, encodes, and streams games from host PC. Clean host constraint (OS APIs only). | Per-OS capture (DXGI DDA / ScreenCaptureKit / KMS+PipeWire). Hardware encoder factory (NVENC/QSV/AMF/VideoToolbox/VAAPI). |
| **P2** | **Hybrid Client Triad** | One Go core shared across three client stacks: Wails (desktop), Flutter+Go FFI (mobile/TV), Angular+Go-WASM (web). | Go compiled 3 ways: c-shared (Flutter FFI), native (Wails backend), WASM (browser). Shared business logic: gRPC, protocol negotiation, controller input. |
| **P3** | **Controller Fidelity Protocol** | Full DualSense feature parity over network: haptics, adaptive triggers, gyro, accelerometer, audio jack passthrough. | 16-32 byte binary protocol. 1000Hz USB polling. Lock-free SPSC ring buffer. Zero-copy IPC via shared memory. WebRTC DataChannel (web) / custom UDP (native). |
| **P4** | **PS4-Class Catalog as Content Business** | Rich game metadata, 4K assets, search, discovery. Multi-source pipeline (IGDB, SteamGridDB, Steam, RAWG). | Per-tenant catalog overlays. SQLite FTS5 + Meilisearch. 4K WebP/AVIF assets. Lazy loading via CloudFront signed URLs. |
| **P5** | **White-Label as Architecture** | Gaming-as-a-Service (GaaS) platform. Partners rebrand without code changes. | 3-tier design tokens (primitive > semantic > component). Tenant-scoped: identity, catalog, recordings, billing. Material Design 3 base. Style Dictionary v4. |
| **P6** | **Edge-First Latency** | Geography matters more than codec. Host within 100km of client. | p999 metric (not p50). LAN <=30ms, WAN <=50ms. Hybrid streaming: WebRTC primary, custom UDP fallback. ABR/FEC/SQP for network adaptation. |
| **P7** | **Anti-Cheat Clean Host** | OS APIs only. No DLL injection, no kernel hooks, no memory tampering. | DXGI DDA (not frame buffer capture). Signed driver requirement. Capability advertisement lets anti-cheat verify clean host. |
| **P8** | **GaaS Multi-Tenancy** | Tenant-scoped from day one. Users, games, catalogs, recordings, billing all isolated per tenant. | CockroachDB for multi-region data. Per-tenant OAuth2/OIDC. RBAC with per-tenant roles. Audit logging. |

### 3.2 System Topology

```
+------------------------------------------------------------------+
|                         CLIENT LAYER                              |
|  +----------------+  +--------------------+  +----------------+  |
|  | Wails Desktop  |  | Flutter Mobile/TV  |  | Angular + WASM |  |
|  |  (Go native)   |  |   (Go c-shared)    |  |   (Go WASM)    |  |
|  +-------+--------+  +---------+----------+  +--------+-------+  |
|          |                    |                      |           |
|          +--------------------+----------------------+           |
|                               |                                  |
|                    +----------v----------+                       |
|                    |   Go Core Shared    |                       |
|                    |  (gRPC, protocol,   |                       |
|                    |  controller input)  |                       |
|                    +----------+----------+                       |
+-------------------------------|----------------------------------+
                                |
+-------------------------------v----------------------------------+
|                      BACKEND SERVICES                             |
|  +-----------+  +----------+  +---------+  +------------------+ |
|  | REST      |  | gRPC     |  | NATS    |  | Service          | |
|  | Gateway   |  | Services |  | JetStream | Discovery        | |
|  | (Go)      |  | (Go)     |  | (Events)  | (mDNS+Rendezvous)| |
|  +------+----+  +-----+----+  +-----+---+  +--------+---------+ |
|         |             |             |              |            |
|         +-------------+------+------+--------------+            |
|                              |                                  |
|  +-----------+  +----------+v+---------+  +------------------+ |
|  | CockroachDB|  | Redis    |  | Vault   |  | RabbitMQ       | |
|  | (Primary)  |  | (Cache)  |  | (Secrets)| | (Legacy)       | |
|  +------------+  +----------+  +---------+  +------------------+ |
+-------------------------------|----------------------------------+
                                |
+-------------------------------v----------------------------------+
|                        HOST LAYER                                 |
|  +-----------------------------------------------------------+   |
|  |                    Sunshine++ Host Agent                   |   |
|  |  +----------+  +--------+  +-------+  +----------------+ |   |
|  |  | Capture  |->| Encoder|->|Stream |->|   Discovery    | |   |
|  |  | (per-OS) |  | (HW)   |  | (WebRTC/QUIC/UDP)       | |   |
|  |  +----------+  +--------+  +-------+  +----------------+ |   |
|  |  +----------+  +--------+  +-------+  +----------------+ |   |
|  |  | Game     |  | Input  |  |Lifecycle| | Capability    | |   |
|  |  | Enum     |  | (1kHz) |  | Manager | | Advertise     | |   |
|  |  +----------+  +--------+  +-------+  +----------------+ |   |
|  +-----------------------------------------------------------+   |
+------------------------------------------------------------------+
```

### 3.3 End-to-End Data Flow

**Game Launch and Play Session (typical user journey):**

```
1. Host Agent enumerates installed games (6 stores) → publishes to Catalog
2. Host Agent advertises capabilities (GPU, codecs, thermal) → mDNS beacon
3. Client discovers host → mDNS (LAN) or rendezvous (WAN)
4. Client negotiates session parameters with Host Agent:
   a. Codec selection (AV1 > HEVC > H.264) based on capability match
   b. Resolution, FPS, bitrate via SQP algorithm
   c. Controller profile (per-game DualSense mapping)
5. Host Agent launches game → Game Lifecycle Manager
6. Capture pipeline grabs frames → zero-copy to encoder
7. Encoder produces dual bitstream:
   a. Stream path → WebRTC/QUIC/UDP → Client
   b. Record path → NVMe ring buffer → background sync
8. Client receives frames → WebCodecs (web) / hardware decoder (native)
9. Controller input → 1kHz polling → binary protocol → Host Agent
10. Input injected → virtual controller (ViGEm/uinput/foohid) → game
11. A/V sync maintained via RTP timestamps + Opus MultiStream
12. Session teardown → quick resume state saved
```

### 3.4 Submodule Ecosystem (29 Submodules)

#### 3.4.1 vasic-digital Organization (25 submodules)

| # | Submodule | Category | Purpose | Origin Chapter |
|---|-----------|----------|---------|----------------|
| 1 | **Auth** | Architecture | JWT, OAuth2/OIDC (Auth0), API key, middleware auth | C11 |
| 2 | **Cache** | Architecture | Redis caching abstractions | C09 |
| 3 | **Challenges** | Operations | Full-stack challenge runner & bluff scanner | C37 |
| 4 | **Concurrency** | Architecture | Non-blocking concurrency primitives (semaphores, backpressure) | C07 |
| 5 | **Containers** | Operations | Container definitions, health checks, lifecycle | C38 |
| 6 | **Database** | Architecture | DB abstractions, connection pooling | C09 |
| 7 | **Discovery** | Architecture | LAN service discovery (mDNS), dynamic ports | C08 |
| 8 | **EventBus** | Architecture | NATS/Redis/RabbitMQ event propagation | C10 |
| 9 | **Formatters** | Architecture | Output formatting, codec negotiation helpers | C06 |
| 10 | **HelixQA** | Testing | Autonomous QA orchestration framework | C37 |
| 11 | **Media** | Architecture | Media processing abstractions | C28 |
| 12 | **Memory** | Latency | Shared memory, zero-copy IPC, memfd ring buffers | C15 |
| 13 | **Messaging** | Architecture | gRPC, REST, HTTP/3 communication stack | C06 |
| 14 | **Middleware** | Architecture | HTTP/gRPC middleware (CORS, rate limiting, auth) | C06 |
| 15 | **Monetization** | Architecture | Billing engine, resource quotas, webhooks | C11 |
| 16 | **Observability** | Operations | Logging, metrics, tracing, event collection | C39 |
| 17 | **Plugins** | Architecture | Client plugin architecture | C13 |
| 18 | **RAG** | Architecture | Retrieval-Augmented Generation support | C13 |
| 19 | **RateLimiter** | Architecture | Token bucket, sliding window rate limiting | C06 |
| 20 | **Recovery** | Architecture | Fault recovery, session state preservation | C12 |
| 21 | **Security** | Architecture | CVE scanning, RBAC, mTLS, audit logging | C11 |
| 22 | **Storage** | Architecture | Recording storage backends (S3, NFS, local) | C30 |
| 23 | **Streaming** | Architecture | WebRTC Pion, QUIC datagrams, custom UDP | C19 |
| 24 | **VectorDB** | Architecture | Vector search for RAG | C13 |
| 25 | **Catalogizer** | Content | Game catalog + metadata pipeline | C14 |

#### 3.4.2 HelixDevelopment Organization (4 submodules)

| # | Submodule | Category | Purpose | Origin Chapter |
|---|-----------|----------|---------|----------------|
| 26 | **Catalogizer** | Content | Game catalog + metadata management | C14 |
| 27 | **HelixQA** | Testing | Autonomous QA orchestration (OpenCV visual verification) | C37 |
| 28 | *(reserved)* | | | |
| 29 | *(reserved)* | | | |

#### 3.4.3 Helix-Specific Domain Submodules (29 total)

Beyond the vasic-digital foundation, the Helix-specific domain is organized into functional submodule groups:

**Latency Domain:**
| Submodule | Purpose | Key Technologies |
|-----------|---------|-----------------|
| helix-allocator | Arena allocation, memory pools | sync.Pool, arena |
| helix-bench | Latency measurement harness | HDR histogram, benchstat |
| helix-display | Display synchronization, VRR | G-Sync/FreeSync |
| helix-gpu-direct | GPU direct transfer (zero-copy) | GPUDirect RDMA, DMA-BUF, IOSurface |
| helix-input | Controller input pipeline | 1kHz USB, lock-free SPSC |
| helix-iouring | io_uring kernel bypass | submission/completion rings |
| helix-lockfree | Lock-free data structures | MPMC queues, hazard pointers |
| helix-mempool | Memory pool management | cache-line alignment (64B x86-64, 128B ARM) |
| helix-network | Ultra-low-latency network | custom UDP, DTLS 1.2, DSCP/L4S |
| helix-rtos | Real-time scheduling | PREEMPT_RT, SCHED_FIFO, isolcpus |
| helix-shm | Shared memory IPC | POSIX shm, memfd, NV12/I420 |

**Video/Audio Domain:**
| Submodule | Purpose | Key Technologies |
|-----------|---------|-----------------|
| helix-abr | Adaptive bitrate (SQP algorithm) | BBRv3 congestion control |
| helix-audio | Audio pipeline | Opus MultiStream, AC3/EAC3/Atmos passthrough |
| helix-capture | Per-OS capture | DXGI DDA, ScreenCaptureKit, KMS/PipeWire |
| helix-codec | Codec ladder + capability negotiation | H.264/HEVC/AV1 |
| helix-display | Display sync (also latency domain) | frame pacing, VRR |
| helix-dualpath | Simultaneous stream+record | same bitstream fork |
| helix-encoder | Hardware encoder factory | NVENC, QSV, AMF, VideoToolbox, VAAPI |
| helix-gpu-direct | GPU direct (also latency domain) | zero-copy GPU-to-network |
| helix-hdr | HDR pipeline | HDR10, HDR10+, Dolby Vision, HLG |
| helix-pipeline | End-to-end video pipeline | Go goroutines, ring buffers |
| helix-record | Recording storage | MKV/fMP4, NVMe ring buffer |
| helix-shm | Shared memory (also latency domain) | zero-copy frame handoff |
| helix-thermal | Thermal-aware quality scaling | DVFS, multi-GPU |
| helix-transport | Network transport abstraction | WebRTC/QUIC/UDP |

**Architecture Domain:**
| Submodule | Purpose | Key Technologies |
|-----------|---------|-----------------|
| helix-grpc-frame | gRPC framing + HTTP/3 | quic-go, protobuf |
| helix-r18-safeexec | Operational integrity wrapper | SafeExec, forbidden command validation |
| helix-tenant | Multi-tenancy | CockroachDB, per-tenant isolation |

#### 3.4.4 Cross-Cutting Policies for All 29 Submodules

| Policy | Implementation | Owner |
|--------|---------------|-------|
| **SIV Versioning** | `/vN` suffix in module path. Tag `v1.0.0` at graduation. | Each submodule |
| **go.work** | Root `go.work` includes all 29 modules. `go work sync` before commit. | Root repo |
| **Dependency Lockstep** | GOPROXY=direct for local, GOSUMDB=sum.golang.org for CI. Renovate for updates. | CI/CD |
| **SBOM Generation** | `cyclonedx-gomod` + `syft` on every release. Attached to GitHub/GitLab releases. | CI/CD |
| **Vulnerability Scanning** | `govulncheck` + Snyk + Renovate. Weekly automated scans. Block merge on critical CVE. | CI/CD |
| **CI Lane Sizing** | GOCACHEPROG for distributed caching. Per-submodule parallel lanes. | CI/CD |
| **Four-Mirror Visibility** | All 29 submodules mirrored to GitHub, GitLab, GitVerse, GitFlic. | Upstreams/ scripts |
| **License Consistency** | MIT default. 4 Apache-2.0 exceptions (helix-grpc-frame, helix-network, helix-shm, helix-thermal). | Legal/CI |

### 3.5 Technology Stack

| Layer | Technology | Version | Purpose |
|-------|-----------|---------|---------|
| **Primary Language** | Go | 1.26.2 (root), 1.25+ (submodules) | All services, clients, host agent |
| **Desktop Client** | Wails v2 | latest | Go backend + JS/TS frontend |
| **Mobile/TV Client** | Flutter 3.x | latest | Dart UI + Go FFI (c-shared) |
| **Web Client** | Angular 17+ | latest | TypeScript + Go WASM + WebCodecs |
| **Streaming** | WebRTC (Pion v4) | v4 | Primary video transport |
| **Low-Latency Transport** | QUIC (quic-go) | RFC 9221 | Datagram fallback, HTTP/3 |
| **Custom UDP** | Parsec BUD-style | DTLS 1.2 | Controller input, legacy compatibility |
| **Message Bus** | NATS JetStream | latest | Async events, service coordination |
| **Cache** | Redis | 7.x | Sessions, rate limiting, pub/sub |
| **Primary Database** | CockroachDB | latest | Multi-region, multi-tenant data |
| **Secrets** | HashiCorp Vault | latest | Credential management, mTLS certs |
| **Legacy MQ** | RabbitMQ | 3.x | Third-party integration |
| **Compression** | Brotli | v1.2.1 | Response compression |
| **Auth** | OAuth2/OIDC (Auth0) | RFC 8628 device grant | User + tenant authentication |
| **Containers** | Docker / Podman | latest | All runtime environments |
| **Capture — Windows** | DXGI DDA | DirectX 11/12 | Zero-copy screen capture |
| **Capture — macOS** | ScreenCaptureKit | macOS 12+ | Zero-copy with IOSurface |
| **Capture — Linux** | KMS/DRM + PipeWire | latest | DMA-BUF zero-copy |
| **Encoder — NVIDIA** | NVENC | 8th-gen (Lovelace), 9th-gen (Blackwell) | H.264/HEVC/AV1 hardware encode |
| **Encoder — Intel** | QSV | Arc Battlemage | H.264/HEVC/AV1 hardware encode |
| **Encoder — AMD** | VCE/AMF | RDNA3/4 | H.264/HEVC/AV1 hardware encode |
| **Encoder — Apple** | VideoToolbox | M3–M5 | H.264/HEVC/AV1 hardware encode |
| **Encoder — Linux** | VAAPI | latest | H.264/HEVC hardware encode |
| **Audio** | Opus MultiStream | RFC 7845 | Up to 7.1 surround |
| **HDR** | HDR10/HDR10+/HLG/DV | ITU-R BT.2100 | Metadata preservation |
| **Quality Assessment** | VMAF | Netflix | Perceptual video quality |
| **Observability** | Prometheus + OTel | latest | Metrics, traces, logs |
| **CI Quality Gates** | SonarQube, Snyk, Semgrep, Trivy, gitleaks, govulncheck | latest | Static analysis, CVE scanning |

### 3.6 Latency Budget (Glass-to-Glass)

| Stage | LAN (p999) | WAN (p999) | Owner |
|-------|-----------|-----------|-------|
| Controller input (USB poll → network) | 2ms | 15ms | helix-input |
| Network transit (host → client) | 5ms | 25ms | helix-network, helix-transport |
| Capture (frame grab → encoder) | 3ms | 3ms | helix-capture |
| Encode (H.264/HEVC/AV1) | 5ms | 5ms | helix-encoder |
| Decode (client-side) | 8ms | 8ms | Client WebCodecs/hardware |
| Display (frame → screen) | 7ms | 7ms | helix-display, OS compositor |
| **TOTAL BUDGET** | **<=30ms** | **<=50ms** | |

**Critical path optimization:** Input and capture run in parallel. Encode and network transmit are pipelined. Client decode starts on first NAL unit receipt (not frame completion). Frame pacing via VRR eliminates vsync wait.



---

## Section 4: Implementation Phases (P00–P06) — Foundation and Core

> **Methodology:** Every task follows TDD: write failing test → implement → verify pass → commit. No exceptions. Every task has a Constitution clause reference.

### Phase 00: Foundation

**Phase Objective:** Establish operator infrastructure, tooling, project boards, and missing root-level files. Propagate the Constitution to all 29 submodules. Verify all submodule dependencies are transitively complete.

**Success Criteria:**
- `README.md` exists at root with accurate project description
- `Makefile` exists at root with standard targets (build, test, lint, challenge)
- `go.work` exists and `go work sync` succeeds
- Constitution propagated to all 29 submodule roots (CLAUDE.md + AGENTS.md + CONSTITUTION.md)
- All submodule dependencies transitively verified
- GitHub Projects + GitLab Issues boards created and linked
- `scripts/` directory contains all required operational scripts

**Dependencies:** None (this is the foundational phase)

**Applicable Constitution Clauses:** R-01 (anti-bluff), R-02 (decoupling), R-12 (tracking), R-13 (four-mirror), R-15 (transitive completeness), R-17 (documentation), R-18 (SafeExec)

---

#### P00.T01: Create Root README.md

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T01 |
| **Description** | Write comprehensive project README.md at repository root |
| **Priority** | P1 — Blocking all other work |
| **Constitution** | R-17 (documentation discipline) |

**Files to Create:**
- `/README.md`

**Content Requirements (exhaustive):**
1. Project name and one-line description
2. Vision statement ("Ultimate gaming experience!")
3. Feature summary (zero lag, full controller fidelity, triple-stack, white-label)
4. System requirements (host: GPU + Windows/macOS/Linux; client: any modern browser)
5. Quick start guide (clone, init submodules, build, run)
6. Architecture overview (ASCII diagram linking host/agent/core/client)
7. Submodule list (all 29 with one-line descriptions)
8. Build instructions (per-platform)
9. Test instructions (all 10 test types)
10. Contributing guidelines (Constitution reference, TDD requirement)
11. License (MIT + Apache-2.0 exceptions)
12. Links to documentation (docs/research/), Constitution, and tracking boards
13. Four-mirror source control note (GitHub primary, GitFlic push)

**Test / Verification:**
- File exists and is non-empty (>200 lines)
- Contains all 13 required sections
- All links are valid (no 404s)
- No forbidden patterns (TODO, FIXME, placeholder)
- Rendered correctly in GitHub/GitLab markdown preview

**Estimated Effort:** 2 hours

---

#### P00.T02: Create Root Makefile

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T02 |
| **Description** | Create unified root Makefile with all standard targets |
| **Priority** | P1 — Blocking CI setup |
| **Constitution** | R-17, R-09 (testing discipline) |

**Files to Create:**
- `/Makefile`

**Targets (exhaustive):**

| Target | Command | Purpose |
|--------|---------|---------|
| `make build` | `go build ./cmd/...` after submodule init | Build all cmd entries |
| `make build-submodules` | `cd vasic-digital/$* && make build` for each | Build all 29 submodules |
| `make test` | `go test -count=1 -race -p 1 ./...` | Unit tests with race detection |
| `make test-integration` | `go test -count=1 -race -p 1 ./tests/integration/...` | Integration tests |
| `make test-e2e` | `go test -count=1 ./tests/e2e/...` | End-to-end tests |
| `make test-bench` | `go test -bench=. -benchmem ./tests/benchmark/...` | Benchmark tests |
| `make test-coverage` | `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html` | Coverage report |
| `make test-security` | `govulncheck ./...` | Security scan |
| `make test-chaos` | `go test ./tests/chaos/...` | Chaos tests |
| `make test-stress` | `go test -timeout=24h ./tests/stress/...` | 24-hour soak |
| `make test-smoke` | `go test -timeout=30s ./tests/smoke/...` | Post-deploy smoke |
| `make test-fullauto` | Orchestrates all test types, fail-fast disabled | Full automation |
| `make test-challenge` | `cd vasic-digital/Challenges && make challenge` | Run challenge suite |
| `make anti-bluff` | `bash scripts/anti-bluff-scan.sh` | Anti-bluff scan (non-overridable) |
| `make verify-submodules` | `python3 scripts/verify-submodules.py` | Verify all 29 submodules |
| `make propagate-constitution` | `bash scripts/propagate-constitution.sh` | Copy Constitution to all submodules |
| `make fmt` | `gofmt -w . && goimports -w .` | Format all Go code |
| `make vet` | `go vet ./...` | Static analysis |
| `make lint` | `golangci-lint run ./...` | Lint all code |
| `make clean` | `rm -rf coverage.out coverage.html bin/` | Clean build artifacts |
| `make docker-build` | `docker compose build` | Build all container images |
| `make docker-up` | `docker compose up -d` | Start full topology |
| `make docker-down` | `docker compose down` | Stop full topology |

**Test / Verification:**
- `make build` succeeds after submodule initialization
- `make test` runs without panic
- `make anti-bluff` executes `scripts/anti-bluff-scan.sh`
- `make verify-submodules` reports all 29 submodules present
- `make fmt` runs without error
- `make vet` runs without error
- `make clean` removes artifacts
- All targets have `-help` text via `##` comments

**Estimated Effort:** 3 hours

---

#### P00.T03: Create go.work Workspace File

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T03 |
| **Description** | Create root go.work file including all 29 submodules |
| **Priority** | P1 — Blocking submodule development |
| **Constitution** | R-04 (go.work workspace) |

**Files to Create:**
- `/go.work`
- `/go.work.sum`

**go.work Content:**
```
go 1.26

use (
    .
    ./vasic-digital/Auth
    ./vasic-digital/Cache
    ./vasic-digital/Challenges
    ./vasic-digital/Concurrency
    ./vasic-digital/Containers
    ./vasic-digital/Database
    ./vasic-digital/Discovery
    ./vasic-digital/EventBus
    ./vasic-digital/Formatters
    ./vasic-digital/HelixQA
    ./vasic-digital/Media
    ./vasic-digital/Memory
    ./vasic-digital/Messaging
    ./vasic-digital/Middleware
    ./vasic-digital/Monetization
    ./vasic-digital/Observability
    ./vasic-digital/Plugins
    ./vasic-digital/RAG
    ./vasic-digital/RateLimiter
    ./vasic-digital/Recovery
    ./vasic-digital/Security
    ./vasic-digital/Storage
    ./vasic-digital/Streaming
    ./vasic-digital/VectorDB
    ./vasic-digital/Catalogizer
    # HelixDevelopment submodules
    # ./HelixDevelopment/HelixQA (if checked out separately)
)
```

**Test / Verification:**
- `go work sync` succeeds without errors
- All submodule `go.mod` files are parsed correctly
- No version conflicts between submodules
- `go build ./cmd/...` succeeds from root using workspace
- `go list -m all` shows all 29 modules

**Estimated Effort:** 1 hour

---

#### P00.T04: Propagate Constitution v2.1.0 to All 29 Submodules

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T04 |
| **Description** | Generate and copy CLAUDE.md, AGENTS.md, and CONSTITUTION.md into all 29 submodule roots |
| **Priority** | P1 — Governance foundation |
| **Constitution** | R-01, R-17, R-13 (four-mirror) |

**Files to Create/Modify:**
- `/scripts/propagate-constitution.sh` (script)
- `vasic-digital/Auth/CLAUDE.md` (generated)
- `vasic-digital/Auth/AGENTS.md` (generated)
- `vasic-digital/Auth/CONSTITUTION.md` (generated)
- (repeat for all 29 submodules)

**Script Requirements:**
1. Read `/CLAUDE.md` and `/CONSTITUTION.md` from root
2. Generate submodule-specific versions with:
   - Submodule name inserted in preamble
   - Submodule-specific build/test commands
   - Cross-reference to root Constitution
3. Copy generated files to each submodule root
4. Verify files were written (size > 0)
5. Generate summary report of propagation status

**Test / Verification:**
- `bash scripts/propagate-constitution.sh` executes without error
- All 29 submodules have CLAUDE.md (>100 lines each)
- All 29 submodules have AGENTS.md (>100 lines each)
- All 29 submodules have CONSTITUTION.md (>500 lines each)
- Each file contains the Prime Directive
- Each file references Constitution v2.1.0
- Script reports summary: "29/29 submodules propagated successfully"

**Estimated Effort:** 4 hours (including script development + execution)

---

#### P00.T05: Verify All Submodule Dependencies Transitively Complete

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T05 |
| **Description** | Verify all 29 submodules are correctly configured in .gitmodules with no missing transitive dependencies |
| **Priority** | P1 — Build integrity |
| **Constitution** | R-15 (transitive completeness) |

**Files to Create/Modify:**
- `/scripts/verify-submodules.py` (create or enhance)
- `/.gitmodules` (verify/modify)

**Verification Script Requirements:**
1. Parse `.gitmodules` and extract all submodule entries
2. Verify 29 submodules are present (count check)
3. For each submodule:
   - Verify directory exists on disk
   - Verify `.git` file points to valid gitdir
   - Verify `go.mod` exists (for Go submodules)
   - Extract module path from `go.mod`
   - Check for `replace` directives that point to non-existent paths
4. Build dependency graph from all `go.mod` files
5. Verify no circular dependencies
6. Verify all `replace` directives resolve to existing directories
7. Report any submodule with unresolved dependencies

**Test / Verification:**
- `python3 scripts/verify-submodules.py` executes without error
- Reports: "29 submodules verified"
- Reports: "0 circular dependencies"
- Reports: "0 unresolved replace directives"
- Exit code 0 on success, non-zero on any issue
- Output is machine-parseable (JSON) for CI consumption

**Estimated Effort:** 3 hours

---

#### P00.T06: Set Up GitHub Projects + GitLab Issues Boards

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T06 |
| **Description** | Create and configure dual-platform tracking boards for P00-P06 |
| **Priority** | P1 — Tracking discipline |
| **Constitution** | R-12 (dual-platform tracking) |

**Files to Create:**
- `/docs/tracking/github-projects-setup.md`
- `/docs/tracking/gitlab-issues-setup.md`

**Setup Requirements:**

**GitHub Projects:**
1. Create project board "HelixPlay Implementation P00-P06"
2. Columns: Backlog, In Progress, Review, Done
3. Add all tasks from P00-P06 as issues
4. Link issues to P00-P06 milestones
5. Configure automated workflows (PR linked > In Progress, merged > Done)
6. Add labels: `phase:P00`, `phase:P01`, ..., `constitution`, `anti-bluff`, `critical-path`

**GitLab Issues:**
1. Mirror all GitHub issues to GitLab Issues
2. Set up bidirectional sync via `/Upstreams/` scripts
3. Configure labels matching GitHub
4. Set milestones matching P00-P06 phases

**Test / Verification:**
- GitHub project board exists and is accessible
- All P00 tasks are created as issues with labels
- GitLab Issues has matching issues
- `Upstreams/sync.sh` syncs issues bidirectionally
- Issue count matches task count from this plan

**Estimated Effort:** 2 hours

---

#### P00.T07: Verify and Enhance Operational Scripts

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T07 |
| **Description** | Ensure all operational scripts exist, are tested, and meet constitutional requirements |
| **Priority** | P1 — Operational safety |
| **Constitution** | R-18 (SafeExec), R-01 (anti-bluff) |

**Files to Verify/Enhance:**
- `/scripts/anti-bluff-scan.sh`
- `/scripts/claim-check.sh`
- `/scripts/propagate-constitution.sh`
- `/scripts/verify-submodules.py`

**Per-Script Requirements:**

**anti-bluff-scan.sh:**
- Scan 1: Forbidden patterns (TODO, FIXME, XXX, HACK, `not implemented`, empty bodies, `panic("not implemented")`)
- Scan 2: `ValidateAntiBluff` verification — confirms unconditional call
- Scan 3: Documentation anti-bluff blocks — checks for Anti-Bluff Verification sections
- Scan 4: Constitution propagation — verifies all submodules reference Constitution
- Scan 5: Vacuous assertion scan — `assert.True(t, true)`, `assert.Equal(t, true, true)`, `assert.Nil(t, nil)`
- Self-test mode: runs against hand-crafted fixtures to verify scanner works
- Exit code: 0 = clean, 1 = violations found
- Output: machine-parseable report

**claim-check.sh:**
- Blocks forbidden commands: suspend, hibernate, shutdown, poweroff, init 0, telinit 0, `rm -rf /`, mkfs, fdisk, dd if=/dev/zero of=/dev/sd
- Validates against whitelist before allowing execution
- Called by `.claude/settings.json` Stop hook
- Exit code: 0 = safe, 1 = forbidden command detected

**Test / Verification:**
- `bash scripts/anti-bluff-scan.sh` runs without error on clean tree
- `bash scripts/anti-bluff-scan.sh --self-test` passes (verifies scanner can catch bluffs)
- `bash scripts/claim-check.sh echo "safe"` returns 0
- `bash scripts/claim-check.sh sudo systemctl suspend` returns 1 (blocked)
- `python3 scripts/verify-submodules.py` returns 0 with "29/29 OK"

**Estimated Effort:** 3 hours

---

#### P00.T08: Create Project Configuration Files

| Attribute | Value |
|-----------|-------|
| **Task ID** | P00.T08 |
| **Description** | Create or verify configuration files for the project |
| **Priority** | P2 — Configuration baseline |
| **Constitution** | R-17 |

**Files to Create/Verify:**
- `/.gitignore` (verify — should exist)
- `/.editorconfig` (create)
- `/.golangci.yml` (create)
- `/.go-mutesting.yml` (verify — should exist)
- `/docker-compose.yml` (create — basic infrastructure)

**.editorconfig:**
```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yaml,yml,json,md}]
indent_style = space
indent_size = 2
```

**.golangci.yml:**
```yaml
run:
  timeout: 5m
  go: "1.26"
linters:
  enable:
    - govet
    - staticcheck
    - unused
    - ineffassign
    - gosimple
    - typecheck
    - gofmt
    - goimports
    - misspell
    - revive
    - errcheck
    - gocritic
linters-settings:
  govet:
    enable-all: true
  gocritic:
    enabled-tags:
      - performance
      - style
      - diagnostic
issues:
  exclude-use-default: false
```

**docker-compose.yml (infrastructure only):**
```yaml
version: "3.8"
services:
  cockroachdb:
    image: cockroachdb/cockroach:latest
    command: start-single-node --insecure
    ports:
      - "26257:26257"
      - "8080:8080"
    volumes:
      - cockroach-data:/cockroach/cockroach-data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  nats:
    image: nats:2-alpine
    ports:
      - "4222:4222"
      - "8222:8222"
    command: "-js"

  vault:
    image: hashicorp/vault:latest
    ports:
      - "8200:8200"
    environment:
      VAULT_DEV_ROOT_TOKEN_ID: dev-token-only

volumes:
  cockroach-data:
```

**Test / Verification:**
- `docker compose config` validates without errors
- `golangci-lint run ./...` executes without config errors
- `.editorconfig` is recognized by editors
- All config files contain no forbidden patterns

**Estimated Effort:** 2 hours

---

### Phase 01: Containers and CI

**Phase Objective:** Bootstrap `vasic-digital/Containers` with complete container definitions. Implement container definitions for host-agent, capture-service, encoder-service, and discovery-beacon. Set up local CI/CD pipeline. Implement the anti-bluff-scan CI lane.

**Success Criteria:**
- `vasic-digital/Containers` has container definitions for all 4 core services
- `docker compose up` starts all services successfully
- CI pipeline runs `anti-bluff-scan` on every commit
- Container health checks pass for all services
- Container image sizes are optimized (multi-stage builds)

**Dependencies:** P00 (Foundation complete)

**Applicable Constitution Clauses:** R-05 (container-first), R-06 (container build matrix), R-01 (anti-bluff), R-11 (quality gates), R-18 (SafeExec)

---

#### P01.T01: Bootstrap vasic-digital/Containers Submodule

| Attribute | Value |
|-----------|-------|
| **Task ID** | P01.T01 |
| **Description** | Initialize and configure the Containers submodule with full build/test infrastructure |
| **Priority** | P1 — Infrastructure foundation |
| **Constitution** | R-05, R-06 |

**Files to Create:**
- `vasic-digital/Containers/Makefile`
- `vasic-digital/Containers/go.mod`
- `vasic-digital/Containers/README.md`
- `vasic-digital/Containers/CLAUDE.md`
- `vasic-digital/Containers/AGENTS.md`
- `vasic-digital/Containers/CONSTITUTION.md`
- `vasic-digital/Containers/containers/host-agent/Dockerfile`
- `vasic-digital/Containers/containers/capture-service/Dockerfile`
- `vasic-digital/Containers/containers/encoder-service/Dockerfile`
- `vasic-digital/Containers/containers/discovery-beacon/Dockerfile`
- `vasic-digital/Containers/containers/tests/bootstrap_test.go`
- `vasic-digital/Containers/scripts/bootstrap.sh`

**Makefile targets:**
- `make build` — build all container images
- `make test` — run Go tests + container validation
- `make test-container` — verify containers start and pass health checks
- `make lint` — lint Dockerfiles
- `make clean` — remove built images

**Test / Verification:**
- `make build` in `vasic-digital/Containers` succeeds
- `make test` passes
- `go test ./...` passes from submodule root
- `docker compose config` validates
- Constitution files exist and contain Prime Directive

**Estimated Effort:** 4 hours

---

#### P01.T02: Implement Host Agent Container

| Attribute | Value |
|-----------|-------|
| **Task ID** | P01.T02 |
| **Description** | Dockerfile and container definition for the Sunshine++ host agent |
| **Priority** | P1 |
| **Constitution** | R-05, R-06 |

**Files to Create:**
- `vasic-digital/Containers/containers/host-agent/Dockerfile`
- `vasic-digital/Containers/containers/host-agent/entrypoint.sh`
- `vasic-digital/Containers/containers/host-agent/healthcheck.go`
- `vasic-digital/Containers/containers/host-agent/healthcheck_test.go`
- `vasic-digital/Containers/containers/tests/host_agent_container_test.go`

**Dockerfile Requirements:**
```dockerfile
# Build stage
FROM golang:1.26-bookworm AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o host-agent ./cmd/host-agent

# Runtime stage
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    libvulkan1 libgl1-mesa-glx libegl1 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /build/host-agent /usr/local/bin/
COPY containers/host-agent/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
    CMD /usr/local/bin/host-agent -healthcheck || exit 1
EXPOSE 48010/udp 48010/tcp 8080/tcp
ENTRYPOINT ["/entrypoint.sh"]
```

**Entrypoint Requirements:**
- Validate GPU device access (`/dev/nvidia*` or `/dev/dri`)
- Set `R18_SAFEEXEC=1` environment variable
- Execute host-agent binary with passed arguments
- Trap SIGTERM for graceful shutdown

**Health Check Requirements:**
- HTTP endpoint on port 8080 returning 200 when healthy
- Check GPU accessibility
- Check capture pipeline initialization
- Return non-200 if any subsystem unhealthy

**Test / Verification:**
- `docker build -f containers/host-agent/Dockerfile -t helix-host-agent .` succeeds
- `docker run --rm helix-host-agent -version` prints version
- Container health check passes within 30 seconds of start
- Container stops gracefully on `docker stop` (SIGTERM handled)
- `make test-container` validates all of the above

**Estimated Effort:** 4 hours

---

#### P01.T03: Implement Capture Service Container

| Attribute | Value |
|-----------|-------|
| **Task ID** | P01.T03 |
| **Description** | Dockerfile for the per-OS capture service |
| **Priority** | P1 |
| **Constitution** | R-05, R-06 |

**Files to Create:**
- `vasic-digital/Containers/containers/capture-service/Dockerfile`
- `vasic-digital/Containers/containers/capture-service/entrypoint.sh`
- `vasic-digital/Containers/containers/tests/capture_container_test.go`

**Dockerfile Requirements:**
- Build stage: `golang:1.26-bookworm` builder
- Runtime stage: `debian:bookworm-slim` with per-OS capture libraries:
  - Linux: `libpipewire-0.3-0`, `libdrm2`, VAAPI drivers
  - (Windows/macOS capture runs natively; containers are Linux-only runtime)
- Includes capture test pattern generator for CI validation
- Exposes shared memory interface for zero-copy frame handoff

**Test / Verification:**
- Container builds successfully
- Container starts and reports health check OK
- `docker run --rm helix-capture -test-pattern` generates test video pattern
- Shared memory interface is accessible from host-agent container
- `make test-container` passes

**Estimated Effort:** 3 hours

---

#### P01.T04: Implement Encoder Service Container

| Attribute | Value |
|-----------|-------|
| **Task ID** | P01.T04 |
| **Description** | Dockerfile for hardware encoder service with NVIDIA runtime |
| **Priority** | P1 |
| **Constitution** | R-05, R-06 |

**Files to Create:**
- `vasic-digital/Containers/containers/encoder-service/Dockerfile`
- `vasic-digital/Containers/containers/encoder-service/entrypoint.sh`
- `vasic-digital/Containers/containers/tests/encoder_container_test.go`

**Dockerfile Requirements:**
```dockerfile
# Build stage
FROM golang:1.26-bookworm AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o encoder-service ./cmd/encoder-service

# Runtime stage with NVIDIA runtime
FROM nvidia/cuda:12.4.0-base-ubuntu22.04
RUN apt-get update && apt-get install -y \
    libnvidia-encode-550 \
    vainfo intel-media-va-driver-non-free \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /build/encoder-service /usr/local/bin/
# ...
```

**Key Requirements:**
- NVIDIA runtime for NVENC access
- Intel VAAPI drivers for QSV fallback
- AMD AMF libraries (if available in base image)
- Encoder capability advertisement endpoint
- Thermal monitoring integration
- Single NVENC session per GPU (C-005 constraint enforcement)

**Test / Verification:**
- Container builds with `docker build`
- `docker run --gpus all helix-encoder -capabilities` lists available encoders
- Encoder initializes NVENC session successfully
- Health check reports encoder ready
- Thermal monitoring reports GPU temperature
- `make test-container` passes

**Estimated Effort:** 4 hours

---

#### P01.T05: Implement Discovery Beacon Container

| Attribute | Value |
|-----------|-------|
| **Task ID** | P01.T05 |
| **Description** | Dockerfile for mDNS + rendezvous discovery beacon |
| **Priority** | P1 |
| **Constitution** | R-05, R-06 |

**Files to Create:**
- `vasic-digital/Containers/containers/discovery-beacon/Dockerfile`
- `vasic-digital/Containers/containers/discovery-beacon/entrypoint.sh`
- `vasic-digital/Containers/containers/tests/discovery_container_test.go`

**Dockerfile Requirements:**
- `debian:bookworm-slim` base (no GPU needed)
- mDNS responder on port 5353/udp
- HTTP API on port 8080 for rendezvous registration
- Capability cache with TTL
- Network mode: host (required for mDNS multicast)

**Test / Verification:**
- Container builds and starts
- mDNS service is discoverable on LAN via `avahi-browse` or `dns-sd`
- HTTP rendezvous endpoint responds with 200
- Capability advertisement includes GPU, codec, and thermal info
- `make test-container` passes

**Estimated Effort:** 3 hours

---

#### P01.T06: Implement Bootstrap Script

| Attribute | Value |
|-----------|-------|
| **Task ID** | P01.T06 |
| **Description** | Create bootstrap script for one-command environment setup |
| **Priority** | P2 |
| **Constitution** | R-05, R-18 |

**Files to Create:**
- `vasic-digital/Containers/scripts/bootstrap.sh`
- `vasic-digital/Containers/containers/tests/bootstrap_test.go`

**Script Requirements:**
1. Check prerequisites: Docker 24+, Docker Compose v2+, Go 1.26+, Python 3.11+
2. Verify GPU access (`nvidia-smi` or `rocm-smi` or Intel GPU)
3. Pull/build all container images
4. Initialize submodule dependencies
5. Start infrastructure services (CockroachDB, Redis, NATS, Vault)
6. Run database migrations
7. Seed initial data (system tenant, admin user)
8. Start core services (host-agent, capture, encoder, discovery)
9. Run health checks on all services
10. Print status summary

**Test / Verification:**
- `bash scripts/bootstrap.sh` executes without error on clean machine
- All services report healthy within 60 seconds
- Status summary shows all green
- `docker ps` shows all expected containers running
- Script uses `r18.SafeExec` for all system commands
- Script exits with code 0 on success, non-zero on any failure

**Estimated Effort:** 4 hours

---

#### P01.T07: Set Up Local CI/CD Pipeline

| Attribute | Value |
|-----------|-------|
| **Task ID** | P01.T07 |
| **Description** | Create local CI pipeline that runs on every commit |
| **Priority** | P1 — Quality gate enforcement |
| **Constitution** | R-01, R-11, R-05 |

**Files to Create:**
- `.github/workflows/ci.yml`
- `.github/workflows/anti-bluff.yml`
- `.github/workflows/security-scan.yml`
- `scripts/ci-pipeline.sh`

**CI Workflow (ci.yml):**
```yaml
name: HelixPlay CI
on: [push, pull_request]
jobs:
  anti-bluff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
      - name: Run Anti-Bluff Scan
        run: bash scripts/anti-bluff-scan.sh
        # NON-OVERRIDABLE: This job must pass

  build:
    runs-on: ubuntu-latest
    needs: anti-bluff
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: Build
        run: make build

  test:
    runs-on: ubuntu-latest
    needs: build
    strategy:
      matrix:
        test-type: [unit, integration, e2e, benchmark, security]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
      - name: Run ${{ matrix.test-type }} tests
        run: make test-${{ matrix.test-type }}

  container-check:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
      - name: Build containers
        run: make docker-build
      - name: Start infrastructure
        run: docker compose up -d cockroachdb redis nats vault
      - name: Run container tests
        run: cd vasic-digital/Containers && make test-container
```

**Anti-Bluff Workflow (anti-bluff.yml):**
- Runs FIRST (before build, before test)
- Cannot be overridden by `skip-checks:true` or similar
- Scans entire tree for forbidden patterns
- Runs negative-leg injection on changed files
- Fails the entire pipeline on any violation

**Security Scan Workflow (security-scan.yml):**
- govulncheck on all modules
- Snyk dependency scan
- Trivy container image scan
- gitleaks secrets scan
- Semgrep static analysis
- All gates must pass

**Test / Verification:**
- `act -j anti-bluff` (local GitHub Actions runner) passes on clean tree
- `act -j anti-bluff` FAILS when TODO comment is added
- `act -j build` succeeds
- `act -j test` passes
- `act -j container-check` passes
- All workflow files validated by `actionlint`

**Estimated Effort:** 5 hours

---

### Phase 02: Core Submodules

**Phase Objective:** Graduate all 29 submodules from v0.x.y to v1.0.0. Resolve all dependency resolution issues (replace directives). Implement missing submodule functionality. Establish cross-cutting policies: SIV versioning, go.work, dependency lockstep, SBOM generation.

**Success Criteria:**
- All 29 submodules tagged at v1.0.0 or higher
- `go work sync` succeeds with zero errors
- No `replace` directives needed for public dependencies
- All submodules pass `make test` (unit + integration)
- SBOM generated for each submodule
- Vulnerability scan passes (zero critical CVEs)

**Dependencies:** P00, P01

**Applicable Constitution Clauses:** R-02 (decoupling), R-03 (SIV), R-04 (go.work), R-09 (testing), R-11 (quality gates), R-15 (transitive completeness)

---

#### P02.T01: Implement vasic-digital/Memory — Shared Memory, Zero-Copy IPC

| Attribute | Value |
|-----------|-------|
| **Task ID** | P02.T01 |
| **Description** | Implement lock-free Producer-Single-Consumer ring buffer using memfd_create |
| **Priority** | P1 — Core performance primitive |
| **Constitution** | R-08 (concurrency), R-09 |

**Files to Create:**
- `vasic-digital/Memory/pkg/memfd/ringbuffer.go`
- `vasic-digital/Memory/pkg/memfd/ringbuffer_test.go`
- `vasic-digital/Memory/pkg/memfd/zerocopy.go`
- `vasic-digital/Memory/pkg/memfd/zerocopy_test.go`
- `vasic-digital/Memory/pkg/memfd/doc.go`
- `vasic-digital/Memory/go.mod`
- `vasic-digital/Memory/Makefile`

**Implementation Requirements:**

`ringbuffer.go`:
- Lock-free SPSC (Single-Producer-Single-Consumer) ring buffer
- Linux-specific: uses `memfd_create` + `mmap` for shared memory
- 128-byte cache-line padding on both head and tail pointers (separate cache lines to prevent false sharing)
- Cache line sizes: 64 bytes (x86-64), 128 bytes (ARM64) — configurable at compile time
- Power-of-2 capacity for efficient modulo via bitmask
- Memory-mapped region shared between processes (host agent <-> capture service)
- Support for NV12 and I420 frame formats (planar YUV)
- `Push()` — producer: write frame metadata + payload, advance head
- `Pop()` — consumer: read frame metadata + payload, advance tail
- `Empty()`, `Full()`, `Count()` — state queries
- Zero-copy: payload pointer returned directly into mmap'd region

`zerocopy.go`:
- `TransferDMA(buf []byte, fd int) error` — DMA-BUF transfer (Linux)
- `TransferIOSurface(buf []byte) (IOSurfaceRef, error)` — IOSurface (macOS)
- `TransferGPUDirect(buf []byte, device int) error` — GPUDirect RDMA (NVIDIA)
- Platform detection at runtime (`runtime.GOOS`)

**Test Requirements (TDD):**
1. `TestRingBuffer_Create` — create buffer with various sizes
2. `TestRingBuffer_PushPop_Single` — push one frame, pop it, verify content
3. `TestRingBuffer_PushPop_Multiple` — push N frames, pop N frames, verify ordering
4. `TestRingBuffer_Full` — buffer reports full at capacity
5. `TestRingBuffer_Empty` — buffer reports empty when nothing pushed
6. `TestRingBuffer_Concurrent_SPSC` — goroutine pushes, goroutine pops, no races
7. `TestRingBuffer_ZeroCopy_Read` — pop returns pointer into mmap'd region
8. `TestRingBuffer_NV12Layout` — verify NV12 frame layout in buffer
9. `TestRingBuffer_CacheLinePadding` — verify head/tail on separate cache lines
10. `BenchmarkRingBuffer_PushPop` — measure latency per operation

**Test / Verification:**
- `cd vasic-digital/Memory && make test` passes (all 10 tests + benchmarks)
- `go test -race` reports zero races
- Benchmark: Push+Pop latency < 100ns per operation
- Memory alignment verified via `unsafe.Offsetof`
- v1.0.0 tag created

**Estimated Effort:** 8 hours

---

#### P02.T02: Implement vasic-digital/Streaming — WebRTC (Pion v4) + QUIC

| Attribute | Value |
|-----------|-------|
| **Task ID** | P02.T02 |
| **Description** | WebRTC streaming transport using Pion v4, plus QUIC datagrams (RFC 9221) |
| **Priority** | P1 — Primary streaming transport |
| **Constitution** | R-07 (communication stack), R-09 |

**Files to Create:**
- `vasic-digital/Streaming/pkg/webrtc/pion.go`
- `vasic-digital/Streaming/pkg/webrtc/pion_test.go`
- `vasic-digital/Streaming/pkg/webrtc/config.go`
- `vasic-digital/Streaming/pkg/quic/datagram.go`
- `vasic-digital/Streaming/pkg/quic/datagram_test.go`
- `vasic-digital/Streaming/pkg/udp/custom.go`
- `vasic-digital/Streaming/pkg/udp/custom_test.go`
- `vasic-digital/Streaming/pkg/transport/factory.go`
- `vasic-digital/Streaming/go.mod`
- `vasic-digital/Streaming/Makefile`

**Implementation Requirements:**

`webrtc/pion.go`:
- Import `github.com/pion/webrtc/v4`
- `PeerConnection` wrapper with Helix-specific config
- ICE server configuration (STUN + TURN for NAT traversal)
- DataChannel for controller input (ordered, reliable)
- RTP sender for video (H.264/HEVC/AV1 payload types)
- RTCP receiver for quality feedback
- DTLS 1.2 mandatory (not 1.3 — compatibility requirement)
- Connection state machine: New -> Checking -> Connected -> Disconnected -> Closed
- Statistics collection: bytes sent/received, packets lost, RTT, jitter

`quic/datagram.go`:
- Import `github.com/quic-go/quic-go`
- RFC 9221 datagram support (unreliable, unordered)
- Fallback transport when WebRTC ICE fails
- HTTP/3 signaling channel
- Congestion control: BBRv3
- DSCP EF (Expedited Forwarding) marking for game traffic

`udp/custom.go`:
- Parsec BUD-style binary protocol
- DTLS 1.2 encryption
- Port 48010 (default)
- Moonlight/GameStream compatibility layer
- Jitter buffer: adaptive, 5-50ms depth
- Packet recovery: NACK-based retransmission + FEC

`transport/factory.go`:
- `CreateTransport(prefer WebRTC|QUIC|UDP) Transport`
- Auto-fallback chain: WebRTC -> QUIC -> UDP
- Capability negotiation with remote peer
- Quality adaptation hooks

**Test Requirements (TDD):**
1. `TestWebRTC_CreatePeerConnection` — create PC with valid config
2. `TestWebRTC_DataChannel` — open DC, send/receive message
3. `TestWebRTC_RTPSend` — send RTP packet, verify stats
4. `TestWebRTC_ICEGathering` — gather candidates, verify non-empty
5. `TestQUIC_DatagramSendRecv` — send datagram, receive, verify content
6. `TestQUIC_HTTP3Signal` — HTTP/3 request/response
7. `TestUDP_CustomPacket` — encode/decode BUD packet
8. `TestUDP_DTLSEncryption` — packet is encrypted on wire
9. `TestTransport_Factory_WebRTC` — factory creates WebRTC transport
10. `TestTransport_Factory_Fallback` — factory falls back when preferred fails
11. `BenchmarkWebRTC_RTPSend` — measure packet send latency
12. `BenchmarkQUIC_Datagram` — measure datagram RTT

**Test / Verification:**
- `cd vasic-digital/Streaming && make test` passes (12+ tests)
- `go test -race` clean
- WebRTC tests use Pion's `NewAPI` with `SettingEngine` for test hooks
- QUIC tests use `quic-go` loopback connection
- UDP tests use `net.ListenUDP` on loopback
- Benchmark: WebRTC RTP send < 50 microseconds per packet
- v1.0.0 tag created

**Estimated Effort:** 12 hours

---

#### P02.T03: Implement vasic-digital/Discovery — mDNS + Rendezvous

| Attribute | Value |
|-----------|-------|
| **Task ID** | P02.T03 |
| **Description** | LAN service discovery via mDNS + WAN rendezvous service |
| **Priority** | P1 |
| **Constitution** | R-07, R-09 |

**Files to Create:**
- `vasic-digital/Discovery/pkg/mdns/server.go`
- `vasic-digital/Discovery/pkg/mdns/server_test.go`
- `vasic-digital/Discovery/pkg/mdns/client.go`
- `vasic-digital/Discovery/pkg/mdns/client_test.go`
- `vasic-digital/Discovery/pkg/rendezvous/server.go`
- `vasic-digital/Discovery/pkg/rendezvous/server_test.go`
- `vasic-digital/Discovery/pkg/rendezvous/client.go`
- `vasic-digital/Discovery/go.mod`
- `vasic-digital/Discovery/Makefile`

**Implementation Requirements:**

mDNS server:
- Import `github.com/grandcat/zeroconf` or `github.com/grandcat/mdns`
- Service type: `_helix._tcp` and `_helix._udp`
- TXT records: capability advertisement (GPU, codecs, thermal, current load)
- Port: 8080 (HTTP API) + 48010 (game traffic)
- TTL: 120 seconds
- Announce on startup, graceful unannounce on shutdown

mDNS client:
- Browse for `_helix._tcp` services
- Parse TXT records into capability struct
- Filter by: GPU vendor, codec support, max resolution, current load < threshold
- Return sorted list (nearest/lowest-latency first)

Rendezvous server:
- HTTP API for WAN host registration
- Registration includes: public IP, STUN-derived reflexive address, capabilities
- Heartbeat requirement: host must ping every 60 seconds
- Automatic deregistration on missed heartbeats (3 consecutive)
- Endpoint: `POST /register`, `POST /heartbeat`, `GET /hosts`, `DELETE /register`

Rendezvous client:
- Register host with rendezvous server
- Send periodic heartbeats
- Query for available hosts (with filtering)

**Test Requirements (TDD):**
1. `TestMDNS_Server_Announce` — server announces, client discovers
2. `TestMDNS_Client_Filter` — client filters by capability
3. `TestMDNS_Server_GracefulShutdown` — unannounce on shutdown
4. `TestRendezvous_Register` — host registers successfully
5. `TestRendezvous_Heartbeat` — heartbeat extends TTL
6. `TestRendezvous_AutoDeregister` — missed heartbeats cause removal
7. `TestRendezvous_Query_Filter` — query with GPU filter returns matching hosts
8. `TestDiscovery_EndToEnd_LAN` — full LAN discovery flow
9. `TestDiscovery_EndToEnd_WAN` — full WAN rendezvous flow

**Test / Verification:**
- `make test` passes (9+ tests)
- mDNS tests use loopback multicast or test double
- Rendezvous tests use `httptest.Server`
- `go test -race` clean
- v1.0.0 tag created

**Estimated Effort:** 8 hours

---

#### P02.T04: Graduate Remaining Submodules to v1.0.0

| Attribute | Value |
|-----------|-------|
| **Task ID** | P02.T04 |
| **Description** | Systematically graduate all remaining submodules from v0.x.y to v1.0.0 |
| **Priority** | P1 |
| **Constitution** | R-03 (SIV), R-15 (transitive completeness), R-09 |

**Submodules to Graduate:**
| # | Submodule | Current Status | Work Required |
|---|-----------|---------------|---------------|
| 1 | Auth | v0.x | Verify v1.0.0 ready, tag |
| 2 | Cache | v0.x | Verify v1.0.0 ready, tag |
| 3 | Challenges | v0.x | Verify v1.0.0 ready, tag |
| 4 | Concurrency | v0.x | Verify v1.0.0 ready, tag |
| 5 | Containers | v0.x | Tag after P01 complete |
| 6 | Database | v0.x | Verify v1.0.0 ready, tag |
| 7 | EventBus | v0.x | Verify v1.0.0 ready, tag |
| 8 | Formatters | v0.x | Verify v1.0.0 ready, tag |
| 9 | Media | v0.x | Implement if needed, tag |
| 10 | Memory | v0.x | Tag after P02.T01 |
| 11 | Messaging | v0.x | Verify v1.0.0 ready, tag |
| 12 | Middleware | v0.x | Verify v1.0.0 ready, tag |
| 13 | Monetization | v0.x | Implement if needed, tag |
| 14 | Observability | v0.x | Verify v1.0.0 ready, tag |
| 15 | Plugins | v0.x | Verify v1.0.0 ready, tag |
| 16 | RAG | v0.x | Verify v1.0.0 ready, tag |
| 17 | RateLimiter | v0.x | Verify v1.0.0 ready, tag |
| 18 | Recovery | v0.x | Verify v1.0.0 ready, tag |
| 19 | Security | v0.x | Verify v1.0.0 ready, tag |
| 20 | Storage | v0.x | Verify v1.0.0 ready, tag |
| 21 | Streaming | v0.x | Tag after P02.T02 |
| 22 | VectorDB | v0.x | Verify v1.0.0 ready, tag |
| 23 | Catalogizer | v0.x | Implement if needed, tag |

**Graduation Checklist (per submodule):**
- [ ] All tests pass (`make test` — unit + integration)
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run ./...` clean
- [ ] README.md exists with description
- [ ] CLAUDE.md, AGENTS.md, CONSTITUTION.md present
- [ ] go.mod has correct module path with SIV
- [ ] No `replace` directives for public dependencies
- [ ] `go mod tidy` run and go.sum clean
- [ ] `cyclonedx-gomod` SBOM generated
- [ ] `govulncheck` passes (zero critical CVEs)
- [ ] License file present (MIT or Apache-2.0)
- [ ] v1.0.0 git tag created and signed

**Test / Verification:**
- `make verify-submodules` reports all 23 at v1.0.0+
- `go work sync` succeeds
- `make build-submodules` succeeds for all
- `govulncheck ./...` on each reports zero critical
- `cyclonedx-gomod app -json -output <submodule>.sbom.json` succeeds for each

**Estimated Effort:** 16 hours (spread across submodules, some parallelizable)

---

#### P02.T05: Resolve Dependency Resolution Issues

| Attribute | Value |
|-----------|-------|
| **Task ID** | P02.T05 |
| **Description** | Eliminate all replace-directive issues that break clean checkouts |
| **Priority** | P1 — Build reliability |
| **Constitution** | R-15, R-04 |

**Problem Statement:**
Current `go.mod` has `replace` directives pointing to local paths:
```go
replace github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory => ./vasic-digital/Memory
```
These break in clean checkouts where submodules aren't initialized.

**Resolution Steps:**
1. Audit ALL `replace` directives across all go.mod files
2. For internal dependencies (HelixDevelopment/*, vasic-digital/*):
   - Publish to GitHub (public module proxy)
   - Use version tags instead of replace
   - OR keep replace but document in README
3. For external dependencies with patches:
   - Fork to HelixDevelopment/ org
   - Tag forked version
   - Use fork in go.mod
4. Update `go.work` to handle local development
5. Update CI to handle both modes:
   - Local dev: `go.work` with replace
   - CI/build: `GOPROXY=direct` with tagged versions

**Files to Modify:**
- `/go.mod` (remove or document replace)
- `vasic-digital/*/go.mod` (audit each)
- `/go.work` (ensure all modules listed)
- `.github/workflows/ci.yml` (add GOPROXY handling)

**Test / Verification:**
- Clean checkout (no submodules): `go build ./cmd/...` works with GOPROXY
- Full checkout (submodules): `go work sync && go build ./cmd/...` works
- `scripts/verify-submodules.py` reports zero unresolved replaces
- CI passes on both modes

**Estimated Effort:** 6 hours

---

#### P02.T06: Implement SBOM Generation and Vulnerability Scanning

| Attribute | Value |
|-----------|-------|
| **Task ID** | P02.T06 |
| **Description** | Set up automated SBOM generation and vulnerability scanning for all 29 submodules |
| **Priority** | P2 — Security baseline |
| **Constitution** | R-11 (quality gates) |

**Files to Create:**
- `/scripts/generate-sbom.sh`
- `/scripts/vuln-scan.sh`
- `.github/workflows/sbom.yml`
- `.snyk` (Snyk configuration)

**SBOM Script Requirements:**
1. Install `cyclonedx-gomod` and `syft` if not present
2. For each of 29 submodules:
   - `cd <submodule> && cyclonedx-gomod app -json -output <name>.sbom.json`
   - Merge into combined SBOM
3. Upload SBOMs as artifacts
4. Attach to GitHub/GitLab releases

**Vulnerability Scan Requirements:**
1. Run `govulncheck ./...` on each submodule
2. Run `snyk test` on root module
3. Run `trivy fs --scanners vuln .` on containers
4. Report summary: critical/high/medium/low counts
5. Block on critical CVEs (exit non-zero)
6. Generate SARIF output for GitHub Security tab

**Test / Verification:**
- `bash scripts/generate-sbom.sh` produces 29 SBOM files
- `bash scripts/vuln-scan.sh` runs without error (or reports CVEs)
- CI workflow runs on schedule (weekly) + on dependency changes
- SBOMs are valid CycloneDX (verify with `cyclonedx validate`)

**Estimated Effort:** 4 hours

---

### Phase 03: Backend Services

**Phase Objective:** Deploy core backend infrastructure: CockroachDB + NATS + Redis + Vault. Implement service discovery infrastructure. Build REST gateway microservice. Establish the communication backbone that all services will use.

**Success Criteria:**
- `docker compose up` brings up all infrastructure services
- CockroachDB accepts connections and executes schema migrations
- NATS JetStream streams are created and accepting messages
- Redis responds to cache operations
- Vault is initialized and serving secrets
- REST gateway handles HTTP requests and proxies to gRPC backends
- Service discovery registers and deregisters services correctly
- All backend services pass integration tests

**Dependencies:** P00, P01, P02

**Applicable Constitution Clauses:** R-07 (communication stack), R-14 (observability), R-16 (security), R-09 (testing)

---

#### P03.T01: Deploy CockroachDB Cluster

| Attribute | Value |
|-----------|-------|
| **Task ID** | P03.T01 |
| **Description** | Deploy CockroachDB for multi-region, multi-tenant data storage |
| **Priority** | P1 — Data foundation |
| **Constitution** | R-07, R-14 |

**Files to Create:**
- `/deploy/cockroachdb/init.sql`
- `/deploy/cockroachdb/migrations/001_system_tenant.up.sql`
- `/deploy/cockroachdb/migrations/002_users.up.sql`
- `/deploy/cockroachdb/migrations/003_games.up.sql`
- `/deploy/cockroachdb/migrations/004_sessions.up.sql`
- `/deploy/cockroachdb/migrations/005_telemetry.up.sql`
- `/deploy/cockroachdb/migrate.go`
- `/deploy/cockroachdb/migrate_test.go`
- `/cmd/core/db/connection.go`
- `/cmd/core/db/connection_test.go`

**Schema Design:**

`system_tenant` (migration 001):
```sql
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name STRING NOT NULL,
    domain STRING UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    config JSONB DEFAULT '{}'
);

INSERT INTO tenants (id, name, domain, config)
VALUES ('00000000-0000-0000-0000-000000000000', 'HelixPlay System', 'system.helixplay.local', '{}')
ON CONFLICT DO NOTHING;
```

`users` (migration 002):
```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    email STRING UNIQUE NOT NULL,
    auth_provider STRING NOT NULL DEFAULT 'local',
    auth_subject STRING,
    role STRING NOT NULL DEFAULT 'player',
    created_at TIMESTAMP DEFAULT now(),
    last_login TIMESTAMP,
    INDEX idx_tenant_email (tenant_id, email),
    INDEX idx_auth (auth_provider, auth_subject)
);
```

`games` (migration 003):
```sql
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    store_id STRING NOT NULL,
    store_type STRING NOT NULL,
    title STRING NOT NULL,
    description STRING,
    cover_art_url STRING,
    installed BOOLEAN DEFAULT false,
    executable_path STRING,
    last_played TIMESTAMP,
    play_time_seconds INT DEFAULT 0,
    INDEX idx_tenant_store (tenant_id, store_type),
    INDEX idx_installed (tenant_id, installed)
);
```

`sessions` (migration 004):
```sql
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    host_id STRING NOT NULL,
    game_id UUID REFERENCES games(id),
    status STRING NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP DEFAULT now(),
    ended_at TIMESTAMP,
    duration_seconds INT,
    quality_profile STRING,
    codec_used STRING,
    avg_latency_ms FLOAT,
    max_latency_ms FLOAT,
    INDEX idx_user (user_id, started_at DESC),
    INDEX idx_host (host_id, status),
    INDEX idx_tenant_active (tenant_id, status)
);
```

`telemetry` (migration 005):
```sql
CREATE TABLE IF NOT EXISTS telemetry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    session_id UUID REFERENCES sessions(id),
    event_type STRING NOT NULL,
    payload JSONB NOT NULL,
    timestamp TIMESTAMP DEFAULT now(),
    INDEX idx_tenant_time (tenant_id, timestamp DESC)
) PARTITION BY RANGE (timestamp);
```

**Migration Tool Requirements:**
- Go-based migration runner (golang-migrate or custom)
- Up/down migrations for each schema version
- Version tracking in `schema_migrations` table
- Idempotent execution (safe to run multiple times)
- CLI: `go run migrate.go up`, `go run migrate.go down`, `go run migrate.go version`

**Test / Verification:**
- `docker compose up -d cockroachdb` starts successfully
- `go run migrate.go up` applies all 5 migrations without error
- `go run migrate.go version` reports "5/5"
- Connection pool test: acquire connection, execute `SELECT 1`, return
- Tenant isolation test: data from tenant A not visible to tenant B
- `make test` passes for `cmd/core/db/`

**Estimated Effort:** 6 hours

---

#### P03.T02: Deploy NATS JetStream

| Attribute | Value |
|-----------|-------|
| **Task ID** | P03.T02 |
| **Description** | Deploy NATS with JetStream for async event propagation |
| **Priority** | P1 |
| **Constitution** | R-07 |

**Files to Create:**
- `/deploy/nats/nats-server.conf`
- `/cmd/core/events/publisher.go`
- `/cmd/core/events/publisher_test.go`
- `/cmd/core/events/consumer.go`
- `/cmd/core/events/consumer_test.go`

**NATS Configuration:**
```hcl
jetstream {
    store_dir: "/data/jetstream"
    max_memory_store: 1GB
    max_file_store: 10GB
}

streams: [
    {
        name: "HELIX_EVENTS"
        subjects: ["helix.>"]
        retention: limits
        max_msgs: 1000000
        max_bytes: 1GB
        max_age: "30d"
        storage: file
        replicas: 1
    },
    {
        name: "HELIX_TELEMETRY"
        subjects: ["telemetry.>"]
        retention: limits
        max_msgs: 5000000
        max_age: "7d"
        storage: file
        replicas: 1
    }
]
```

**Event Publisher Requirements:**
- `Publish(subject string, data []byte) error` — fire-and-forget
- `PublishAsync(subject string, data []byte) (nats.PubAckFuture, error)` — async with ack
- `Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error)` — RPC pattern
- Subject hierarchy: `helix.<tenant>.<service>.<event>`
- JSON serialization by default, msgpack for high-frequency events

**Event Consumer Requirements:**
- `Subscribe(subject string, handler MsgHandler) (*nats.Subscription, error)`
- `SubscribeQueue(subject, queue string, handler MsgHandler)` — load-balanced
- Durable consumer with ack tracking
- Automatic reconnection with backoff

**Test / Verification:**
- `docker compose up -d nats` starts successfully
- Publisher test: publish 1000 messages, all acknowledged
- Consumer test: subscribe, publish, verify delivery
- Queue group test: 3 consumers in queue, publish 100 messages, each gets ~33
- Reconnection test: restart NATS container, client reconnects and resumes
- JetStream persistence test: restart NATS, messages retained

**Estimated Effort:** 4 hours

---

#### P03.T03: Deploy Redis Cache

| Attribute | Value |
|-----------|-------|
| **Task ID** | P03.T03 |
| **Description** | Deploy Redis for session caching, rate limiting, and pub/sub |
| **Priority** | P1 |
| **Constitution** | R-07 |

**Files to Create:**
- `/cmd/core/cache/redis.go`
- `/cmd/core/cache/redis_test.go`

**Implementation Requirements:**
- `redis/go-redis/v9` client
- Connection pool: min=5, max=50 connections
- Key prefixing: `helix:<tenant>:<entity>:<id>`
- Default TTL: 1 hour for sessions, 5 minutes for rate limits
- Pub/sub for real-time notifications
- Lua scripts for atomic operations (rate limit counter)

**Test / Verification:**
- `docker compose up -d redis` starts
- Set/get/delete operations succeed
- TTL expiration works (wait for expiry, verify gone)
- Pub/sub: publish message, subscribed client receives
- Rate limiter Lua script: increments counter, returns current value
- Connection pool: concurrent operations from 100 goroutines

**Estimated Effort:** 3 hours

---

#### P03.T04: Deploy HashiCorp Vault

| Attribute | Value |
|-----------|-------|
| **Task ID** | P03.T04 |
| **Description** | Deploy Vault for secrets management and mTLS certificate issuance |
| **Priority** | P1 — Security foundation |
| **Constitution** | R-16 (security), R-07 |

**Files to Create:**
- `/deploy/vault/vault-config.hcl`
- `/cmd/core/secrets/vault.go`
- `/cmd/core/secrets/vault_test.go`
- `/cmd/core/secrets/mock.go` // for tests only

**Vault Configuration:**
```hcl
storage "file" {
    path = "/vault/data"
}

listener "tcp" {
    address = "0.0.0.0:8200"
    tls_disable = true  // Dev mode only; production uses mTLS
}

ui = true
disable_mlock = true
```

**Secret Manager Requirements:**
- `Get(path string) (map[string]interface{}, error)` — read secret
- `Put(path string, data map[string]interface{}) error` — write secret
- `GetCert(role string) (*tls.Certificate, error)` — issue mTLS certificate
- Token renewal: automatic before expiry
- Path structure: `secret/helix/<tenant>/<service>/<key>`

**Test / Verification:**
- `docker compose up -d vault` starts
- `vault status` reports "initialized: true, sealed: false" (dev mode)
- Write secret, read back, verify match
- Certificate issuance: request cert, verify key pair
- Token renewal: verify token refreshed before expiry

**Estimated Effort:** 3 hours

---

#### P03.T05: Implement REST Gateway Microservice

| Attribute | Value |
|-----------|-------|
| **Task ID** | P03.T05 |
| **Description** | HTTP REST gateway that proxies to gRPC backend services |
| **Priority** | P1 — API entry point |
| **Constitution** | R-07, R-09 |

**Files to Create:**
- `/cmd/core/gateway/main.go`
- `/cmd/core/gateway/handlers/discovery.go`
- `/cmd/core/gateway/handlers/discovery_test.go`
- `/cmd/core/gateway/handlers/session.go`
- `/cmd/core/gateway/handlers/session_test.go`
- `/cmd/core/gateway/handlers/catalog.go`
- `/cmd/core/gateway/handlers/catalog_test.go`
- `/cmd/core/gateway/middleware/auth.go`
- `/cmd/core/gateway/middleware/ratelimit.go`
- `/cmd/core/gateway/middleware/logging.go`
- `/cmd/core/gateway/proto/api.proto`
- `/cmd/core/gateway/proto/generate.go`

**Implementation Requirements:**

Main gateway:
- `grpc-gateway` or custom HTTP-to-gRPC proxy
- Listen on `:8080` (HTTP) and `:8443` (HTTPS with mTLS)
- Brotli compression for responses > 1KB
- CORS configured for web client origin
- Request ID injection for tracing
- Structured JSON logging

Handlers:
- `GET /api/v1/hosts` — list discoverable hosts (proxies to Discovery service)
- `POST /api/v1/sessions` — create gaming session
- `GET /api/v1/sessions/:id` — get session status
- `DELETE /api/v1/sessions/:id` — end session
- `GET /api/v1/games` — list games for tenant
- `GET /api/v1/games/:id` — game details
- `POST /api/v1/games/:id/launch` — launch game on host
- `GET /api/v1/health` — health check
- `GET /metrics` — Prometheus metrics

Middleware:
- Auth: validate JWT from `Authorization: Bearer <token>` header
- Rate limit: per-IP and per-user token bucket
- Logging: structured JSON, request duration, status code

**Test / Verification:**
- `go test ./cmd/core/gateway/...` passes (all handlers)
- `httptest` used for HTTP handler tests
- Auth middleware rejects requests without valid JWT
- Rate limiter blocks requests over threshold
- Health endpoint returns 200
- Metrics endpoint returns Prometheus format
- Brotli compression verified via `Content-Encoding` header

**Estimated Effort:** 10 hours

---

#### P03.T06: Implement Service Discovery Infrastructure

| Attribute | Value |
|-----------|-------|
| **Task ID** | P03.T06 |
| **Description** | Service registry and discovery for backend microservices |
| **Priority** | P1 |
| **Constitution** | R-07 |

**Files to Create:**
- `/cmd/core/discovery/registry.go`
- `/cmd/core/discovery/registry_test.go`
- `/cmd/core/discovery/health.go`
- `/cmd/core/discovery/health_test.go`

**Implementation Requirements:**
- In-memory service registry with TTL-based expiration
- Registration: `Register(serviceName, instanceID, address, metadata) error`
- Deregistration: `Deregister(instanceID) error`
- Discovery: `Discover(serviceName) ([]Instance, error)` — returns healthy instances
- Health checking: periodic HTTP health probe to each registered instance
- Heartbeat: instances must heartbeat every 30 seconds
- Load balancing: round-robin selection across healthy instances

**Test / Verification:**
- Register service, discover, verify returned
- Deregister service, discover, verify not returned
- Health check marks unhealthy instance after failed probes
- Heartbeat timeout removes expired instances
- Concurrent registration from 100 goroutines

**Estimated Effort:** 5 hours

---

### Phase 04: Streaming Pipeline

**Phase Objective:** Make `helix-pipeline` and `helix-transport` operational. Integrate WebRTC (Pion v4) for primary video transport. Implement QUIC datagrams (RFC 9221) for fallback. Establish end-to-end video flow from capture to client display.

**Success Criteria:**
- Video frames flow from capture service through encoder to transport
- WebRTC connection establishes between host and client
- Video displays in client (test pattern verified)
- Latency measured and reported by `helix-bench`
- QUIC datagram fallback works when WebRTC fails
- A/V sync maintained within 5ms

**Dependencies:** P02 (Streaming submodule), P03 (backend), P01 (containers)

**Applicable Constitution Clauses:** R-07, R-08 (concurrency), R-09, KSC-06 (latency)

---

#### P04.T01: Integrate helix-pipeline (End-to-End Video Pipeline)

| Attribute | Value |
|-----------|-------|
| **Task ID** | P04.T01 |
| **Description** | Wire capture → encode → packetize → transport pipeline in Go |
| **Priority** | P1 — Core streaming |
| **Constitution** | R-08, R-09 |

**Files to Create:**
- `/cmd/core/streaming/pipeline.go`
- `/cmd/core/streaming/pipeline_test.go`
- `/cmd/core/streaming/frame.go`
- `/cmd/core/streaming/frame_test.go`

**Implementation Requirements:**

Pipeline architecture:
```
[Capture Source] → [Frame Queue] → [Encoder] → [Packetizer] → [Transport]
     ↓                                                        ↑
[Frame Metadata]                                    [RTCP Feedback]
     ↓                                                        ↑
[Pipeline Controller] ← [Quality Adapter] ← [Network Stats]
```

`pipeline.go`:
- `Pipeline` struct with channels for each stage
- Capture source: reads from `helix-capture` shared memory
- Frame queue: ring buffer (from `vasic-digital/Memory`), drops oldest on overflow
- Encoder: calls `helix-encoder` hardware encoder via gRPC
- Packetizer: NAL unit splitting, RTP header generation, sequence numbers
- Transport: sends via WebRTC (primary) or QUIC (fallback)
- Pipeline controller: manages start/stop/pause, handles errors
- Quality adapter: adjusts bitrate/resolution based on RTCP feedback

Concurrency (Constitution R-08):
- Each stage runs in its own goroutine
- Channels connect stages with buffered capacity
- `sync.Pool` for frame buffers (zero-allocation hot path)
- Semaphore-based backpressure — producer blocks when consumer falls behind
- Graceful shutdown via `context.Context` cancellation

`frame.go`:
- `Frame` struct: timestamp, format (NV12/I420/RGBA), resolution, payload
- `FrameMetadata`: capture time, encode time, transmit time (for latency tracking)
- `FramePool` — `sync.Pool` of reusable frame buffers

**Test / Verification:**
1. `TestPipeline_Create` — create pipeline with test configuration
2. `TestPipeline_StartStop` — start pipeline, verify running, stop cleanly
3. `TestPipeline_FrameFlow` — inject test frame, verify it reaches transport mock
4. `TestPipeline_FrameDrops` — overflow queue, verify oldest frames dropped
5. `TestPipeline_QualityAdapt` — inject network loss, verify bitrate reduced
6. `TestPipeline_GracefulShutdown` — stop during active streaming, no goroutine leaks
7. `TestPipeline_Concurrent` — 1000 frames from 10 goroutines, all processed
8. `BenchmarkPipeline_FrameFlow` — measure end-to-end latency per frame

**Acceptance Criteria:**
- `go test -race` clean
- Goroutine leak test passes (use `go.uber.org/goleak`)
- Benchmark: end-to-end frame latency < 20ms per frame
- Zero allocations in hot path (verified via `benchmem`)

**Estimated Effort:** 12 hours

---

#### P04.T02: Integrate WebRTC Transport (Pion v4)

| Attribute | Value |
|-----------|-------|
| **Task ID** | P04.T02 |
| **Description** | Full WebRTC integration for video streaming |
| **Priority** | P1 |
| **Constitution** | R-07, R-09 |

**Files to Create:**
- `/cmd/core/streaming/webrtc_transport.go`
- `/cmd/core/streaming/webrtc_transport_test.go`
- `/cmd/core/streaming/signal.go`
- `/cmd/core/streaming/signal_test.go`

**Implementation Requirements:**

`webrtc_transport.go`:
- Import `github.com/pion/webrtc/v4`
- `WebRTCTransport` implementing a `Transport` interface
- ICE gathering: host, srflx, relay candidates
- DTLS 1.2 handshake (mandatory)
- SRTP key derivation from DTLS
- Video track: H.264 (96), HEVC (98), AV1 (99) payload types
- RTCP receiver: process PLI, FIR, REMB, TWCC
- Congestion control: Google congestion control algorithm (GCC) with TWCC
- Bandwidth estimation from TWCC feedback
- Simulcast support: multiple resolution layers
- SVC (Spatial Video Coding) for AV1

`signal.go`:
- SDP offer/answer exchange via REST gateway (not custom signaling)
- `POST /api/v1/sessions/:id/sdp` — send offer, get answer
- ICE candidate trickling via REST
- Trickle ICE: candidates sent as discovered, not batched

**Test / Verification:**
1. `TestWebRTC_CreateTransport` — create transport with config
2. `TestWebRTC_Connect` — full connection establishment (loopback)
3. `TestWebRTC_SendFrame` — send video frame, verify RTP packets generated
4. `TestWebRTC_RTCPFeedback` — inject RTCP, verify bitrate adaptation
5. `TestWebRTC_Reconnect` — disconnect, reconnect, resume streaming
6. `TestWebRTC_MultiCodec` — negotiate H.264, verify payload type 96
7. `TestWebRTC_ICECandidates` — gather candidates, verify non-empty
8. `BenchmarkWebRTC_SendFrame` — measure frame send latency

**Test Doubles:**
- Use Pion's `NewAPI(MediaEngine, SettingEngine)` for test control
- Mock network via `net.Pipe()` or Pion's test utilities
- Loopback connection: offerer and answerer in same process

**Acceptance Criteria:**
- WebRTC connection establishes in < 2 seconds (LAN)
- Video frame send latency < 5ms per frame
- RTCP feedback triggers bitrate change within 100ms
- Reconnection completes in < 1 second
- `go test -race` clean

**Estimated Effort:** 14 hours

---

#### P04.T03: Implement QUIC Datagram Fallback (RFC 9221)

| Attribute | Value |
|-----------|-------|
| **Task ID** | P04.T03 |
| **Description** | QUIC datagram transport as fallback when WebRTC fails |
| **Priority** | P1 — Fallback transport |
| **Constitution** | R-07, R-09 |

**Files to Create:**
- `/cmd/core/streaming/quic_transport.go`
- `/cmd/core/streaming/quic_transport_test.go`

**Implementation Requirements:**
- Import `github.com/quic-go/quic-go`
- `QUICTransport` implementing `Transport` interface
- HTTP/3 for signaling (session establishment)
- RFC 9221 datagrams for video packets (unreliable, unordered)
- DTLS-equivalent security via QUIC's built-in TLS 1.3
- Congestion control: BBRv3
- DSCP EF marking for game traffic
- Connection migration: survive local IP change
- 0-RTT connection resumption for fast reconnect

`Transport` interface (shared with WebRTC):
```go
type Transport interface {
    Connect(ctx context.Context, addr string) error
    Disconnect() error
    SendVideo(packet *RTPPacket) error
    SendAudio(packet *RTPPacket) error
    SendInput(data []byte) error
    SetVideoHandler(handler RTPHandler)
    SetAudioHandler(handler RTPHandler)
    SetInputHandler(handler InputHandler)
    Stats() TransportStats
    Latency() time.Duration
}
```

**Test / Verification:**
1. `TestQUIC_CreateTransport` — create transport
2. `TestQUIC_Connect` — establish connection to test server
3. `TestQUIC_SendVideo` — send video datagram, verify receipt
4. `TestQUIC_SendAudio` — send audio datagram, verify receipt
5. `TestQUIC_0RTT` — connect, disconnect, reconnect with 0-RTT
6. `TestQUIC_Migration` — change local IP, verify connection survives
7. `TestTransport_Interface` — both WebRTC and QUIC satisfy Transport interface
8. `BenchmarkQUIC_SendVideo` — measure datagram send latency

**Acceptance Criteria:**
- QUIC connection establishes in < 1 second (0-RTT: < 100ms)
- Datagram send latency < 3ms
- 0-RTT resumption works after first connection
- Connection migration survives IP change
- Satisfies same Transport interface as WebRTC (interchangeable)

**Estimated Effort:** 10 hours

---

#### P04.T04: Implement ABR/FEC/SQP Policies

| Attribute | Value |
|-----------|-------|
| **Task ID** | P04.T04 |
| **Description** | Adaptive bitrate, forward error correction, and subjective quality profile |
| **Priority** | P1 — Quality adaptation |
| **Constitution** | R-08, R-09 |

**Files to Create:**
- `/cmd/core/streaming/abr.go`
- `/cmd/core/streaming/abr_test.go`
- `/cmd/core/streaming/fec.go`
- `/cmd/core/streaming/fec_test.go`
- `/cmd/core/streaming/sqp.go`
- `/cmd/core/streaming/sqp_test.go`

**Implementation Requirements:**

`abr.go` — Adaptive Bitrate:
- SQP (Subjective Quality Profile) algorithm
- Inputs: packet loss rate, RTT, jitter, available bandwidth
- Output: target bitrate, resolution, FPS, codec
- Profiles: Ultra (4K60), High (1080p60), Medium (1080p30), Low (720p30), Minimum (480p30)
- Hysteresis: require 3 consecutive measurements before switching up, 1 for switching down
- Codec fallback: AV1 -> HEVC -> H.264 based on packet loss (>3% triggers fallback)

`fec.go` — Forward Error Correction:
- Reed-Solomon (n, k) encoding for video packets
- Dynamic FEC rate: 0% (loss < 1%), 10% (1-3%), 20% (3-5%), 30% (>5%)
- NACK-based retransmission for I-frames (prioritized)
- FEC only for P-frames (B-frames not used in low-latency mode)

`sqp.go` — Subjective Quality Profile:
- Per-game quality profiles (fast-action vs. slow-paced)
- User preference override (quality vs. latency priority)
- Scene complexity detection (from encoder feedback)
- Temporal quality adaptation: boost bitrate on scene changes

**Test / Verification:**
1. `TestABR_SelectProfile` — given network stats, verify correct profile selected
2. `TestABR_Hysteresis` — single measurement doesn't trigger switch
3. `TestABR_CodecFallback` — >3% loss triggers H.264 fallback
4. `TestFEC_EncodeDecode` — encode packets, decode, verify recovery
5. `TestFEC_DynamicRate` — loss rate changes, FEC rate adapts
6. `TestSQP_GameProfile` — fast-action game gets higher bitrate
7. `TestSQP_SceneChange` — scene change triggers bitrate boost
8. `BenchmarkABR_SelectProfile` — profile selection < 1 microsecond

**Acceptance Criteria:**
- Profile switches within 500ms of sustained network change
- FEC recovers from 5% packet loss without visible artifacts
- Scene change boost prevents quality drops
- `go test -race` clean

**Estimated Effort:** 8 hours

---

#### P04.T05: Audio Pipeline Integration

| Attribute | Value |
|-----------|-------|
| **Task ID** | P04.T05 |
| **Description** | Integrate Opus audio streaming with A/V synchronization |
| **Priority** | P1 |
| **Constitution** | R-07, R-09 |

**Files to Create:**
- `/cmd/core/streaming/audio.go`
- `/cmd/core/streaming/audio_test.go`
- `/cmd/core/streaming/sync.go`
- `/cmd/core/streaming/sync_test.go`

**Implementation Requirements:**
- Opus MultiStream encoding (up to 7.1 surround)
- AC3/EAC3 passthrough for surround systems
- Dolby Atmos passthrough (metadata only, not decoded)
- RTP timestamp synchronization with video
- A/V sync: maintain < 5ms offset
- Jitter buffer: adaptive, 20-80ms depth for audio
- eARC support detection and negotiation
- Audio jack passthrough for DualSense controller headphone

`sync.go`:
- `Synchronizer` struct with video and audio RTP timestamps
- `AddVideoTS(ts uint32)` — record video RTP timestamp
- `AddAudioTS(ts uint32)` — record audio RTP timestamp
- `Drift() time.Duration` — calculate A/V drift
- `AdjustAudioDelay(drift time.Duration)` — adjust audio playback speed

**Test / Verification:**
1. `TestAudio_EncodeOpus` — encode audio frame, verify output
2. `TestAudio_RTPPacket` — audio RTP packet has correct timestamp
3. `TestSync_NoDrift` — equal timestamps, zero drift
4. `TestSync_DriftDetection` — 10ms offset detected correctly
5. `TestSync_Adjustment` — drift triggers audio speed adjustment
6. `TestAudio_JitterBuffer` — out-of-order packets reordered correctly

**Acceptance Criteria:**
- A/V sync maintained within 5ms during normal operation
- Jitter buffer handles 50ms out-of-order without glitches
- Audio encode latency < 5ms per frame

**Estimated Effort:** 6 hours

---

### Phase 05: Triple-Stack Clients

**Phase Objective:** Build the Go core shared business logic (gRPC, protocol negotiation). Replace the Wails desktop client stub with full implementation. Implement the Flutter mobile/TV client with FFI to Go core. Implement the Angular web client with Go-WASM + WebCodecs.

**Success Criteria:**
- Go core compiles to all three targets: native (Wails), c-shared (Flutter), WASM (Angular)
- Wails desktop client connects and displays video
- Flutter mobile/TV client connects and displays video
- Angular web client connects and displays video
- All clients pass E2E tests with real video decode
- Protocol negotiation selects optimal codec per client capability
- Controller input works on all platforms

**Dependencies:** P02 (submodules), P03 (backend), P04 (streaming)

**Applicable Constitution Clauses:** R-07, R-09, KSC-05

---

#### P05.T01: Implement Go Core Shared Business Logic

| Attribute | Value |
|-----------|-------|
| **Task ID** | P05.T01 |
| **Description** | Shared Go library: gRPC client, protocol negotiation, controller input, session management |
| **Priority** | P1 — Shared across all clients |
| **Constitution** | R-02 (decoupling), R-07, R-09 |

**Files to Create:**
- `/cmd/go-core/discovery/client.go`
- `/cmd/go-core/discovery/client_test.go`
- `/cmd/go-core/protocol/negotiate.go`
- `/cmd/go-core/protocol/negotiate_test.go`
- `/cmd/go-core/protocol/codec_ladder.go`
- `/cmd/go-core/session/manager.go`
- `/cmd/go-core/session/manager_test.go`
- `/cmd/go-core/input/controller.go`
- `/cmd/go-core/input/controller_test.go`
- `/cmd/go-core/input/dualsense.go`
- `/cmd/go-core/go.mod`
- `/cmd/go-core/Makefile`

**Implementation Requirements:**

`discovery/client.go`:
- gRPC client for service discovery
- mDNS browser (LAN) + REST client (WAN rendezvous)
- Host filtering by: GPU, codec, latency, load
- Host caching with TTL
- Automatic host reconnection on disconnect

`protocol/negotiate.go`:
- `Negotiate(capability LocalCapability, remote RemoteCapability) (SessionConfig, error)`
- Codec priority: AV1 > HEVC > H.264 (premium-first selection)
- Resolution negotiation: client max vs. host max vs. network capacity
- FPS negotiation: 60fps preferred, 30fps fallback
- Feature negotiation: HDR, surround audio, recording
- Returns agreed SessionConfig with all parameters

`protocol/codec_ladder.go`:
- Codec capability struct: supported codecs, max profile/level, HDR support
- Codec ladder: ordered list from best to fallback
- `SelectCodec(ladder []CodecCapability, constraints NetworkConstraints) CodecCapability`
- Network constraints: max bitrate, packet loss tolerance, RTT

`session/manager.go`:
- Session lifecycle: Create -> Negotiating -> Connecting -> Active -> Ended
- Heartbeat: send every 5 seconds while active
- Reconnection: automatic with exponential backoff
- State machine with event callbacks
- Session persistence: resume interrupted sessions

`input/controller.go`:
- Platform abstraction: `Controller` interface
- `ReadState() ControllerState` — poll current state
- `WriteFeedback(feedback FeedbackData)` — haptics, adaptive triggers
- Supported: DualSense, Xbox, generic HID

`input/dualsense.go`:
- DualSense-specific implementation
- Haptic feedback (stereo vibration)
- Adaptive trigger feedback (resistance profiles)
- Gyroscope (3-axis) and accelerometer (3-axis) data
- Touchpad coordinates
- Audio jack audio output
- 1000Hz USB polling rate
- Binary protocol: 16-32 byte packets

**Test / Verification:**
1. `TestDiscovery_LAN` — discover host via mDNS mock
2. `TestDiscovery_WAN` — discover host via rendezvous mock
3. `TestNegotiate_CodecPriority` — AV1 selected when both support it
4. `TestNegotiate_Fallback` — H.264 when AV1 not supported
5. `TestSession_Lifecycle` — full state machine transition
6. `TestSession_Reconnect` — disconnect, reconnect, resume
7. `TestController_ReadState` — read state, verify button mappings
8. `TestDualSense_Feedback` — send haptic feedback, verify packet format
9. `TestCodecLadder_Select` — network constraints trigger fallback

**Acceptance Criteria:**
- All tests pass with `-race`
- `go build` succeeds for: `GOOS=linux`, `GOOS=darwin`, `GOOS=windows`
- `go build -buildmode=c-shared` produces `.so`/`.dylib`/`.dll`
- `GOOS=js GOARCH=wasm go build` produces `.wasm`
- No CGO dependencies in code paths used by WASM target

**Estimated Effort:** 16 hours

---

#### P05.T02: Implement Wails Desktop Client (Replace Stub)

| Attribute | Value |
|-----------|-------|
| **Task ID** | P05.T02 |
| **Description** | Replace stub Wails client with full implementation: video display, controller input, session management |
| **Priority** | P1 — Desktop platform |
| **Constitution** | R-09, KSC-05 |

**Files to Create/Modify:**
- `/cmd/client-wails/main.go` (modify)
- `/cmd/client-wails/backend/backend.go` (rewrite)
- `/cmd/client-wails/backend/backend_test.go` (create)
- `/cmd/client-wails/backend/session.go` (create)
- `/cmd/client-wails/backend/video.go` (create)
- `/cmd/client-wails/backend/input.go` (create)
- `/cmd/client-wails/backend/discovery.go` (create)
- `/cmd/client-wails/frontend/src/App.tsx` (create)
- `/cmd/client-wails/frontend/src/components/VideoPlayer.tsx` (create)
- `/cmd/client-wails/frontend/src/components/HostList.tsx` (create)
- `/cmd/client-wails/frontend/src/components/GameGrid.tsx` (create)
- `/cmd/client-wails/frontend/src/components/ControllerConfig.tsx` (create)
- `/cmd/client-wails/frontend/src/hooks/useSession.ts` (create)
- `/cmd/client-wails/frontend/package.json` (modify)
- `/cmd/client-wails/go.mod` (modify)

**Implementation Requirements:**

Backend (`backend.go`):
- `App` struct with Wails runtime reference
- `Startup(ctx context.Context)` — initialize Go core, start discovery
- `Shutdown()` — clean shutdown, end active session
- `DiscoverHosts() ([]HostInfo, error)` — expose to frontend
- `ConnectHost(hostID string) error` — connect to selected host
- `LaunchGame(gameID string) error` — launch game on connected host
- `GetSessionState() SessionState` — current session status
- `EndSession() error` — end current session
- `OnVideoFrame(frame []byte)` — receive video frame from Go core, emit to frontend
- `OnControllerInput(input ControllerState)` — receive from frontend, send to Go core

Video pipeline:
- Go core handles WebRTC/QUIC connection
- Decoded frames passed via Wails Events to frontend
- Frontend renders using HTML5 `<video>` element (WebRTC) or custom decoder
- Frame statistics: FPS, bitrate, latency displayed in UI

Input pipeline:
- Frontend captures gamepad via Web Gamepad API
- Input sent to backend via Wails Events
- Backend converts to DualSense binary protocol
- Sent via Go core input channel to host

Frontend (React/TypeScript):
- Host discovery list: shows available hosts with capability badges
- Game grid: shows installed games with cover art
- Video player: full-screen with overlay stats (FPS, latency, bitrate)
- Controller configuration: per-game button mapping
- Settings: quality profile, codec preference, audio options

**Test / Verification:**
1. `TestBackend_Startup` — startup initializes correctly
2. `TestBackend_DiscoverHosts` — returns mock host list
3. `TestBackend_ConnectHost` — connects to mock host
4. `TestBackend_SessionLifecycle` — full create → active → end flow
5. `TestBackend_VideoFrame` — frame received, passed to frontend
6. `TestBackend_ControllerInput` — input sent to backend
7. E2E: Start Wails app, verify host list displays
8. E2E: Connect to host, verify video player appears
9. `wails build` succeeds for all platforms (Windows, macOS, Linux)

**Acceptance Criteria:**
- `wails dev` starts development server successfully
- `wails build` produces executable for target platform
- Host discovery displays available hosts
- Video playback displays test pattern from host
- Controller input is captured and sent
- Session state transitions correctly
- `go test ./cmd/client-wails/...` passes

**Estimated Effort:** 18 hours

---

#### P05.T03: Implement Flutter Mobile/TV Client with Go FFI

| Attribute | Value |
|-----------|-------|
| **Task ID** | P05.T03 |
| **Description** | Flutter client for mobile and TV with Dart FFI to Go core |
| **Priority** | P1 — Mobile/TV platform |
| **Constitution** | R-09, KSC-05 |

**Files to Create:**
- `/clients/mobile/lib/go_core.dart`
- `/clients/mobile/lib/go_core_bindings.dart`
- `/clients/mobile/lib/main.dart`
- `/clients/mobile/lib/screens/host_discovery_screen.dart`
- `/clients/mobile/lib/screens/game_library_screen.dart`
- `/clients/mobile/lib/screens/streaming_screen.dart`
- `/clients/mobile/lib/screens/settings_screen.dart`
- `/clients/mobile/lib/widgets/game_card.dart`
- `/clients/mobile/lib/widgets/video_player.dart`
- `/clients/mobile/lib/services/controller_service.dart`
- `/clients/mobile/lib/services/session_service.dart`
- `/clients/mobile/pubspec.yaml`
- `/clients/mobile/android/app/build.gradle`
- `/clients/mobile/ios/Runner/Info.plist`
- `/clients/mobile/Makefile`
- `/cmd/go-core/ffi/ffi.go`
- `/cmd/go-core/ffi/ffi_test.go`

**Implementation Requirements:**

Go FFI layer (`ffi.go`):
```go
package main

import "C"
import "github.com/HelixDevelopment/HelixPlay/cmd/go-core/discovery"
import "github.com/HelixDevelopment/HelixPlay/cmd/go-core/session"

//export HelixInitialize
func HelixInitialize() *C.char { ... }

//export HelixDiscoverHosts
func HelixDiscoverHosts() *C.char { ... }

//export HelixConnectHost
func HelixConnectHost(hostID *C.char) *C.char { ... }

//export HelixLaunchGame
func HelixLaunchGame(gameID *C.char) *C.char { ... }

//export HelixGetSessionState
func HelixGetSessionState() *C.char { ... }

//export HelixEndSession
func HelixEndSession() *C.char { ... }

//export HelixSetVideoCallback
func HelixSetVideoCallback(callback unsafe.Pointer) { ... }

//export HelixSendControllerInput
func HelixSendControllerInput(data *C.uint8_t, length C.int) { ... }

//export HelixGetStats
func HelixGetStats() *C.char { ... }

func main() {}
```

Build targets:
- Android: `go build -buildmode=c-shared -o libgo_core.so` (arm64-v8a, armeabi-v7a, x86_64)
- iOS: `go build -buildmode=c-archive -o libgo_core.a` (arm64, simulator)
- TV: Android TV (Leanback) + tvOS

Flutter layer:
- `go_core.dart`: Dart FFI bindings using `dart:ffi`
- `HostDiscoveryScreen`: mDNS discovery with animated list
- `GameLibraryScreen`: grid of installed games with cover art
- `StreamingScreen`: fullscreen video with overlay controls
- `SettingsScreen`: quality, codec, controller configuration
- `ControllerService`: gamepad input via `gamepads` package
- `SessionService`: state management with `flutter_bloc`

TV-specific:
- D-pad navigation support
- Focus management for 10-foot UI
- Voice search integration
- Overscan compensation
- Large text and high contrast options

**Test / Verification:**
1. `TestFFI_Initialize` — initialize returns success
2. `TestFFI_DiscoverHosts` — returns JSON host list
3. `TestFFI_SessionLifecycle` — create → connect → end
4. `TestFFI_ControllerInput` — send input, verify host receives
5. `TestFFI_Stats` — returns valid JSON stats
6. `flutter test` passes (widget tests)
7. `flutter build apk` succeeds
8. `flutter build ios` succeeds (on macOS)
9. Integration test: app launches, discovers mock host

**Acceptance Criteria:**
- Go core compiles to `.so` for Android (arm64)
- FFI functions are callable from Dart
- App launches on Android/iOS
- Host discovery displays available hosts
- Video streaming displays (test pattern)
- Controller input is captured via Flutter gamepad API
- TV navigation works with D-pad

**Estimated Effort:** 20 hours

---

#### P05.T04: Implement Angular Web Client with Go-WASM + WebCodecs

| Attribute | Value |
|-----------|-------|
| **Task ID** | P05.T04 |
| **Description** | Angular web client with Go compiled to WASM and WebCodecs API for video decode |
| **Priority** | P1 — Web platform |
| **Constitution** | R-09, KSC-05 |

**Files to Create:**
- `/clients/web/src/wasm/loader.ts`
- `/clients/web/src/wasm/go_bridge.ts`
- `/clients/web/src/wasm/video_decoder.ts`
- `/clients/web/src/app/app.module.ts`
- `/clients/web/src/app/app.component.ts`
- `/clients/web/src/app/app.component.html`
- `/clients/web/src/app/services/streaming.service.ts`
- `/clients/web/src/app/services/discovery.service.ts`
- `/clients/web/src/app/services/session.service.ts`
- `/clients/web/src/app/services/controller.service.ts`
- `/clients/web/src/app/components/leanback/leanback.component.ts`
- `/clients/web/src/app/components/leanback/leanback.component.html`
- `/clients/web/src/app/components/leanback/leanback.component.css`
- `/clients/web/src/app/components/video-canvas/video-canvas.component.ts`
- `/clients/web/src/app/components/host-browser/host-browser.component.ts`
- `/clients/web/src/app/components/game-launcher/game-launcher.component.ts`
- `/clients/web/src/app/components/settings/settings.component.ts`
- `/clients/web/angular.json`
- `/clients/web/package.json`
- `/clients/web/tsconfig.json`
- `/cmd/go-core/wasm/wasm.go`
- `/cmd/go-core/wasm/wasm_test.go`

**Implementation Requirements:**

Go-WASM layer (`wasm.go`):
```go
package main

import (
    "syscall/js"
    // imports from go-core (must be pure Go, no CGO)
)

func main() {
    c := make(chan struct{}, 0)
    
    js.Global().Set("helixInitialize", js.FuncOf(initialize))
    js.Global().Set("helixDiscoverHosts", js.FuncOf(discoverHosts))
    js.Global().Set("helixConnectHost", js.FuncOf(connectHost))
    js.Global().Set("helixLaunchGame", js.FuncOf(launchGame))
    js.Global().Set("helixEndSession", js.FuncOf(endSession))
    js.Global().Set("helixGetSessionState", js.FuncOf(getSessionState))
    js.Global().Set("helixSetVideoCallback", js.FuncOf(setVideoCallback))
    js.Global().Set("helixSendControllerInput", js.FuncOf(sendControllerInput))
    js.Global().Set("helixGetStats", js.FuncOf(getStats))
    
    <-c
}
```

**CRITICAL:** Go-WASM target cannot use CGO. All code paths called from WASM must be pure Go.

WebCodecs integration (`video_decoder.ts`):
- `VideoDecoder` API for hardware-accelerated decode
- H.264: `avc1.42001E` (baseline) through `avc1.640033` (high)
- HEVC: `hev1.1.6.L153.B0` (main profile)
- AV1: `av01.0.04M.08` (main profile, level 4.0)
- Config from SDP/codec negotiation
- Frame callback renders to `<canvas>` element
- Latency mode: `realtime` for low latency

Angular components:
- `LeanbackComponent`: 10-foot TV UI with D-pad navigation
- `VideoCanvasComponent`: WebCodecs video renderer
- `HostBrowserComponent`: mDNS discovery + rendezvous query
- `GameLauncherComponent`: game list with launch button
- `SettingsComponent`: quality, codec, network settings

Controller input (web):
- Web Gamepad API for controller detection
- Gamepad polling at 60Hz (browser rAF loop)
- DualSense detection via vendor/product IDs
- Haptic feedback via `vibrationActuator`
- Adaptive trigger: not supported in Web Gamepad API (polyfill for future)

**Test / Verification:**
1. `TestWASM_Initialize` — Go-WASM initializes in test environment
2. `TestWASM_DiscoverHosts` — returns JSON host list
3. `TestWASM_SessionLifecycle` — full flow in WASM
4. `TestWebCodecs_H264Decode` — decode H.264 test frame
5. `TestWebCodecs_HEVCDecode` — decode HEVC test frame (if supported)
6. `ng test` passes (Karma/Jasmine unit tests)
7. `ng build --configuration production` succeeds
8. `ng e2e` passes (Protractor/Cypress)

**Acceptance Criteria:**
- `GOOS=js GOARCH=wasm go build` produces `.wasm` file < 50MB
- WASM loads and initializes in browser
- WebCodecs decodes video frames
- Host discovery works
- Game launch triggers streaming session
- Controller input captured via Web Gamepad API
- 10-foot UI navigable with keyboard (D-pad simulation)
- `ng build` output optimized for production

**Estimated Effort:** 18 hours

---

#### P05.T05: Implement Client-Side Protocol Negotiation

| Attribute | Value |
|-----------|-------|
| **Task ID** | P05.T05 |
| **Description** | Unified protocol negotiation across all three client stacks |
| **Priority** | P1 |
| **Constitution** | R-07, R-09 |

**Files to Create:**
- `/cmd/go-core/protocol/client_negotiator.go`
- `/cmd/go-core/protocol/client_negotiator_test.go`

**Implementation Requirements:**
- `ClientNegotiator` struct with local capability detection
- Local capability detection:
  - Video: max resolution (from screen), codec support (from WebCodecs/canPlayType)
  - Audio: channel count, surround support
  - Network: estimated bandwidth (from RTT + throughput test)
  - Controller: connected gamepads, DualSense detection
  - HDR: display HDR capability detection
- Negotiation protocol:
  1. Client sends `ClientCapability` to host
  2. Host responds with `HostCapability`
  3. Client computes intersection (common capabilities)
  4. Client sends `SessionRequest` with preferred config
  5. Host responds with `SessionConfig` (accepted config)
- Codec negotiation: AV1 > HEVC > H.264 with network-aware fallback
- Resolution negotiation: start at client native, scale down if network constrained
- HDR negotiation: HDR10 > HDR10+ > HLG, with SDR fallback

**Test / Verification:**
1. `TestNegotiator_DetectLocal` — detects local capabilities
2. `TestNegotiator_Intersect` — client/host intersection correct
3. `TestNegotiator_AV1Preferred` — AV1 selected when both support
4. `TestNegotiator_HEVCFallback` — HEVC when AV1 not available
5. `TestNegotiator_H264Fallback` — H.264 as universal fallback
6. `TestNegotiator_NetworkConstrained` — low bandwidth triggers resolution reduction
7. `TestNegotiator_HDR` — HDR negotiated when both support

**Acceptance Criteria:**
- Negotiation completes in < 500ms
- Optimal codec selected for each client capability
- Network constraints trigger graceful degradation
- All three client stacks use same negotiation logic

**Estimated Effort:** 6 hours

---

### Phase 06: Host Agent Integration

**Phase Objective:** Make Sunshine++ host agent fully operational. Integrate game enumeration (6 stores), discovery beacon (mDNS + rendezvous), game lifecycle (launch, monitor, terminate, quick resume), and capability advertisement. Validate end-to-end game streaming.

**Success Criteria:**
- Host agent starts and advertises capabilities
- All 6 game stores enumerated (Steam, Epic, GOG, Ubisoft, Battle.net, Microsoft Store)
- Discovery beacon responds to mDNS queries and rendezvous registration
- Game launch → stream → terminate cycle completes
- Quick resume restores game state within 5 seconds
- End-to-end test: client discovers host, launches game, receives video
- HelixQA challenge `host_stream_game` passes

**Dependencies:** P02 (submodules), P04 (streaming), P05 (clients)

**Applicable Constitution Clauses:** R-05 (containers), R-07, R-09, R-18, KSC-04, KSC-06

---

#### P06.T01: Integrate Sunshine++ Host Agent

| Attribute | Value |
|-----------|-------|
| **Task ID** | P06.T01 |
| **Description** | Assemble Sunshine++ host agent from implemented submodules |
| **Priority** | P1 — Core host functionality |
| **Constitution** | R-05, R-09 |

**Files to Create/Modify:**
- `/cmd/host-agent/main.go` (modify/enhance)
- `/cmd/host-agent/agent.go` (create)
- `/cmd/host-agent/agent_test.go` (create)
- `/cmd/host-agent/config.go` (create)
- `/cmd/host-agent/config_test.go` (create)

**Implementation Requirements:**

`agent.go`:
- `HostAgent` struct composing all submodules:
  - `CapabilityAdvertiser` — GPU/codec/thermal capability
  - `CapturePipeline` — per-OS frame capture
  - `CodecNegotiator` — codec ladder + capability exchange
  - `DiscoveryBeacon` — mDNS + rendezvous
  - `EncoderPipeline` — hardware encoding (NVENC/QSV/AMF/etc.)
  - `GameEnumerator` — 6-store game discovery
  - `InputInjector` — controller input injection
  - `LifecycleManager` — game launch/monitor/terminate/quick resume
  - `TransportManager` — WebRTC/QUIC/UDP streaming
- Lifecycle: `New() -> Initialize() -> Start() -> Run() -> Stop()`
- Configuration: load from file, env vars, or CLI flags
- Health check endpoint: HTTP `/health` returning subsystem status
- Metrics endpoint: Prometheus `/metrics`

`config.go`:
- TOML configuration file support
- Sections: `[server]`, `[capture]`, `[encoder]`, `[streaming]`, `[discovery]`, `[games]`, `[input]`, `[logging]`
- Environment variable override: `HELIX_SERVER_PORT=8080`
- Validation: all required fields present, ranges checked
- Defaults: sensible defaults for each setting

**Test / Verification:**
1. `TestAgent_Create` — create agent with test config
2. `TestAgent_Initialize` — initialize all submodules
3. `TestAgent_StartStop` — start agent, verify healthy, stop cleanly
4. `TestAgent_HealthCheck` — health endpoint returns 200 with all OK
5. `TestAgent_ConfigLoad` — load config from file, verify settings
6. `TestAgent_ConfigValidation` — invalid config rejected
7. Integration: `docker run helix-host-agent` starts and reports healthy

**Acceptance Criteria:**
- Agent starts within 5 seconds
- All subsystems report healthy
- Graceful shutdown in < 3 seconds
- Configuration hot-reload (SIGHUP)
- Metrics exported in Prometheus format

**Estimated Effort:** 10 hours

---

#### P06.T02: Integrate Game Enumeration (6 Stores)

| Attribute | Value |
|-----------|-------|
| **Task ID** | P06.T02 |
| **Description** | Integrate game enumeration for all 6 PC game stores |
| **Priority** | P1 |
| **Constitution** | R-09 |

**Files to Create/Modify:**
- `/cmd/host-agent/game/enumerator.go` (verify/enhance)
- `/cmd/host-agent/game/enumerator_test.go` (verify/enhance)
- `/cmd/host-agent/game/steam.go`
- `/cmd/host-agent/game/epic.go`
- `/cmd/host-agent/game/gog.go`
- `/cmd/host-agent/game/ubisoft.go`
- `/cmd/host-agent/game/battlenet.go`
- `/cmd/host-agent/game/microsoft.go`

**Implementation Requirements (per store):**

**Steam:**
- Read `steamapps/*.acf` files for installed apps
- Parse `appmanifest_<id>.acf` for name, install dir, last played
- Steam Web API for metadata (if API key available)
- Library paths: Windows (`C:\Program Files (x86)\Steam\steamapps`), Linux (`~/.steam/steam/steamapps`)

**Epic Games Store:**
- Read `%PROGRAMDATA%\Epic\EpicGamesLauncher\Data\Manifests\*.item` (Windows)
- Parse JSON manifest for display name, install location, executable
- Registry fallback for install path

**GOG Galaxy:**
- Read `%LOCALAPPDATA%\GOG.com\Galaxy\storage\galaxy-2.0.db` (SQLite)
- Query `products` and `installers` tables
- Parse `goggame-<id>.info` files in game directories

**Ubisoft Connect:**
- Read `%LOCALAPPDATA%\Ubisoft Game Launcher\settings.yml`
- Parse YAML for game entries and install paths
- Executable discovery in install directories

**Battle.net:**
- Read `%PROGRAMDATA%\Battle.net\Agent\product.db` (SQLite)
- Parse `products` table for installed games
- Read `.build.info` files for version/executable info

**Microsoft Store:**
- Use Windows PackageManager API (PowerShell `Get-AppxPackage`)
- Filter by gaming category
- Parse manifest for executable and display name

**Unified interface:**
```go
type StoreEnumerator interface {
    Name() string
    FindGames() ([]Game, error)
    IsInstalled() bool // is the store client installed?
}

type Game struct {
    ID              string
    StoreType       string
    Title           string
    Description     string
    CoverArtPath    string
    InstallPath     string
    ExecutablePath  string
    LastPlayed      time.Time
    PlayTimeSeconds int64
}
```

**Test / Verification:**
1. `TestEnumerator_Steam` — parse mock `appmanifest_123.acf`
2. `TestEnumerator_Epic` — parse mock `.item` manifest
3. `TestEnumerator_GOG` — query mock SQLite DB
4. `TestEnumerator_Ubisoft` — parse mock `settings.yml`
5. `TestEnumerator_BattleNet` — query mock `product.db`
6. `TestEnumerator_Microsoft` — mock PackageManager response
7. `TestEnumerator_Aggregate` — aggregate from all 6 stores, no duplicates
8. `TestEnumerator_NotInstalled` — skip store if not installed

**Acceptance Criteria:**
- All 6 store enumerators have unit tests
- Aggregate list contains no duplicates (by title + store)
- Games sorted by last played (most recent first)
- Installed status correctly detected
- `make test` passes

**Estimated Effort:** 12 hours

---

#### P06.T03: Integrate Discovery Beacon (mDNS + Rendezvous)

| Attribute | Value |
|-----------|-------|
| **Task ID** | P06.T03 |
| **Description** | Integrate mDNS LAN discovery and WAN rendezvous beacon |
| **Priority** | P1 |
| **Constitution** | R-07, R-09 |

**Files to Create/Modify:**
- `/cmd/host-agent/discovery/beacon.go` (verify/enhance)
- `/cmd/host-agent/discovery/beacon_test.go` (verify/enhance)
- `/cmd/host-agent/discovery/mdns.go`
- `/cmd/host-agent/discovery/rendezvous.go`

**Implementation Requirements:**

mDNS beacon:
- Announce `_helix._tcp` and `_helix._udp` services
- TXT records:
  - `version=<semver>`
  - `gpu=<vendor>:<model>`
  - `codecs=<comma-list>`
  - `max_res=<width>x<height>`
  - `hdr=<true|false>`
  - `load=<cpu-percent>:<gpu-percent>`
  - `thermal=<gpu-temp-c>`
- Refresh announcement every 60 seconds
- Graceful unannounce on shutdown (remove from browse results)
- Port: 8080 (HTTP API) + 48010 (game traffic)

Rendezvous beacon:
- `POST https://rendezvous.helixplay.io/register` on startup
- Request body: `{"host_id": "...", "public_ip": "...", "capabilities": {...}}`
- Heartbeat: `POST /heartbeat` every 60 seconds
- Auto-deregister on graceful shutdown: `DELETE /register`
- Handle 401 (re-register) and 503 (backoff) responses

**Test / Verification:**
1. `TestBeacon_mDNSAnnounce` — announce service, verify discoverable
2. `TestBeacon_mDNSTXT` — verify TXT records contain capability data
3. `TestBeacon_mDNSUnannounce` — shutdown removes service
4. `TestBeacon_RendezvousRegister` — successful registration
5. `TestBeacon_RendezvousHeartbeat` — heartbeat extends TTL
6. `TestBeacon_RendezvousDeregister` — graceful shutdown deregisters
7. `TestBeacon_Refresh` — announcement refreshed periodically
8. Integration: beacon starts, client discovers host via mDNS

**Acceptance Criteria:**
- mDNS announcement discoverable via `avahi-browse -r _helix._tcp`
- Rendezvous registration returns 200
- Heartbeat prevents deregistration
- Graceful shutdown removes both mDNS and rendezvous entries
- TXT records contain all capability fields

**Estimated Effort:** 6 hours

---

#### P06.T04: Integrate Game Lifecycle Manager

| Attribute | Value |
|-----------|-------|
| **Task ID** | P06.T04 |
| **Description** | Launch, monitor, terminate, and quick-resume games |
| **Priority** | P1 |
| **Constitution** | R-09, R-18 |

**Files to Create/Modify:**
- `/cmd/host-agent/lifecycle/manager.go` (verify/enhance)
- `/cmd/host-agent/lifecycle/manager_test.go` (verify/enhance)
- `/cmd/host-agent/lifecycle/process.go`
- `/cmd/host-agent/lifecycle/monitor.go`
- `/cmd/host-agent/lifecycle/quicksave.go`

**State Machine:**
```
[Idle] --launch--> [Starting] --process running--> [Running]
                                                       |
[Running] --pause request--> [Pausing] --saved--> [Paused]
                                                       |
[Paused] --resume request--> [Resuming] --restored--> [Running]
                                                       |
[Running] --terminate--> [Stopping] --exited--> [Idle]
```

**Implementation Requirements:**

Launch:
- Execute game executable with working directory set
- Pass through environment variables
- Set CPU priority to HIGH (Windows) or nice -10 (Linux)
- GPU affinity: bind to specified GPU if multi-GPU
- Capture process ID for monitoring

Monitor:
- Poll process status every 1 second
- Detect crash (unexpected exit with non-zero code)
- Collect: CPU%, GPU%, VRAM usage, frame time (from PresentMon)
- Report stats via NATS `helix.<host>.telemetry`
- Alert on: crash, freeze (no frames for 5s), thermal throttle

Terminate:
- Send graceful exit signal (WM_CLOSE on Windows, SIGTERM on Linux)
- Wait up to 10 seconds for graceful shutdown
- Force kill if graceful fails (WM_QUIT / SIGKILL)
- Clean up capture and encoder resources

Quick Resume:
- Save state: serialize game process memory + GPU state
- Write to NVMe ring buffer (fastest available storage)
- On resume: restore process from saved state
- Target: resume within 5 seconds
- Per-game quick save slots (max 3 per game)

**Test / Verification:**
1. `TestLifecycle_Launch` — launch mock executable, verify running
2. `TestLifecycle_Monitor` — monitor detects process status
3. `TestLifecycle_Terminate` — graceful termination succeeds
4. `TestLifecycle_CrashDetection` — detects non-zero exit
5. `TestLifecycle_FreezeDetection` — detects no frame output
6. `TestLifecycle_QuickSave` — save state, verify written
7. `TestLifecycle_QuickResume` — restore state, verify running
8. `TestLifecycle_StateMachine` — all transitions verified
9. R-18: `TestLifecycle_NoForbiddenCommands` — no suspend/shutdown/hibernate

**Acceptance Criteria:**
- Game launches and process starts
- Monitor detects running status correctly
- Graceful termination succeeds within 10 seconds
- Crash detection triggers alert
- Quick save completes within 10 seconds
- Quick resume completes within 5 seconds
- No forbidden commands used (R-18 verified)

**Estimated Effort:** 12 hours

---

#### P06.T05: Integrate Capability Advertisement

| Attribute | Value |
|-----------|-------|
| **Task ID** | P06.T05 |
| **Description** | Advertise host GPU, codec, thermal, and load capabilities |
| **Priority** | P1 |
| **Constitution** | R-09 |

**Files to Create/Modify:**
- `/cmd/host-agent/capability/advertise.go` (verify/enhance)
- `/cmd/host-agent/capability/advertise_test.go` (verify/enhance)
- `/cmd/host-agent/capability/gpu.go`
- `/cmd/host-agent/capability/codec.go`
- `/cmd/host-agent/capability/thermal.go`

**Implementation Requirements:**

GPU detection:
- NVIDIA: `nvml` library or `nvidia-smi` query
- Intel: `intel-gpu-tools` or DXGI adapter enumeration
- AMD: `amdgpu` sysfs or ADL library
- Apple: `MTLCreateSystemDefaultDevice()` (Metal)
- Report: vendor, model, VRAM, driver version, CUDA/ROCm level

Codec capability:
- Query encoder support: H.264 (baseline/main/high), HEVC (main/main10), AV1 (main)
- Max resolution per codec: 4K, 8K (where supported)
- Max FPS: 60, 120 (where supported)
- HDR support: HDR10, HDR10+, HLG, Dolby Vision
- Concurrent session count: typically 1 (C-005 constraint)

Thermal monitoring:
- GPU temperature (Celsius)
- Thermal throttling status
- Fan speed percentage
- Target: keep below 80C

Load metrics:
- CPU utilization percentage
- GPU utilization percentage
- VRAM usage (used/total)
- Available for: load balancing, quality adaptation

**Test / Verification:**
1. `TestCapability_GPU` — detects GPU vendor and model
2. `TestCapability_Codec` — reports supported codecs
3. `TestCapability_Thermal` — reads GPU temperature
4. `TestCapability_Load` — reads CPU/GPU utilization
5. `TestCapability_JSON` — serializes to JSON correctly
6. `TestCapability_TXTRecords` — formats for mDNS TXT

**Acceptance Criteria:**
- GPU detected correctly on all platforms
- All supported codecs listed
- Thermal data updates every 5 seconds
- Load data available for load balancing
- JSON serialization valid

**Estimated Effort:** 6 hours

---

#### P06.T06: End-to-End Host-Agent-to-Client Streaming Test

| Attribute | Value |
|-----------|-------|
| **Task ID** | P06.T06 |
| **Description** | Verify complete video streaming from host agent to client |
| **Priority** | P1 — Critical path validation |
| **Constitution** | R-01 (anti-bluff), R-09, KSC-04, KSC-06 |

**Files to Create:**
- `/tests/e2e/streaming_test.go`
- `/tests/e2e/host_client_test.go`
- `/tests/challenges/host_stream_game.go`
- `/tests/fixtures/test_pattern.h264`

**Test Requirements:**

E2E streaming test:
1. Start host agent container with test pattern generator
2. Start client (Wails test harness)
3. Client discovers host via mDNS
4. Client connects and negotiates session
5. Host starts streaming test pattern (color bars)
6. Client receives and decodes video frames
7. Verify: frame count > 100 over 5 seconds, no decode errors
8. Measure latency: frame timestamp difference < 30ms (LAN target)
9. Verify A/V sync (if audio test pattern included)

HelixQA Challenge (`host_stream_game`):
- Boots full topology: host agent + client
- Launches test pattern (not real game — CI environment)
- Captures client screen via OpenCV
- Verifies: color bars visible on screen (visual assertion)
- Measures: frames received per second >= 30
- Reports: latency histogram (p50, p99, p999)
- Anti-bluff: deliberately corrupt stream, verify test FAILS

**Test / Verification:**
- `go test ./tests/e2e/... -run TestStreaming` passes
- `go test ./tests/e2e/... -run TestHostClient` passes
- Challenge `host_stream_game` passes with visual assertion
- Latency: p999 < 30ms (LAN loopback)
- Frame rate: >= 30 FPS sustained
- Anti-bluff negative-leg: corrupted stream causes test failure

**Acceptance Criteria:**
- E2E test passes with real video frames flowing
- HelixQA visual assertion confirms video on screen
- p999 latency < 30ms on loopback
- Anti-bluff negative-leg fails as expected
- Test runs in CI container environment

**Estimated Effort:** 10 hours

---

## Cross-Phase Quality Gates

### QG-01: Anti-Bluff Scan (Every Phase)

| Gate | Enforcement | Applied |
|------|-------------|---------|
| Forbidden patterns | `scripts/anti-bluff-scan.sh` | P00-P06, every commit |
| Vacuous assertions | `assert.True(t, true)` detection | P00-P06, every commit |
| Constitution propagation | All submodules have CONSTITUTION.md | P00, then ongoing |
| Negative-leg injection | CI breaks feature, verifies test fails | P02-P06, every PR |
| Documentation completeness | No TODO/FIXME/placeholder | Every commit |

### QG-02: Quality Metrics (End of Each Phase)

| Phase | Minimum Unit Coverage | Integration Tests | E2E Tests | Benchmarks |
|-------|----------------------|-------------------|-----------|------------|
| P00 | N/A (no code) | 0 | 0 | 0 |
| P01 | 80% | 2+ | 0 | 0 |
| P02 | 90% | 5+ | 0 | 2+ |
| P03 | 85% | 8+ | 1+ | 2+ |
| P04 | 90% | 6+ | 1+ | 3+ |
| P05 | 85% | 5+ | 2+ | 1+ |
| P06 | 90% | 8+ | 3+ | 3+ |

### QG-03: Security Gates (Every Phase)

| Gate | Tool | Frequency |
|------|------|-----------|
| Go vulnerability scan | `govulncheck` | Every commit |
| Dependency scan | Snyk | Every commit + weekly |
| Container scan | Trivy | On container build |
| Secret scan | gitleaks | Every commit |
| Static analysis | Semgrep | Every PR |
| Code quality | SonarQube | Every PR |

### QG-04: Container Health (P01 Onward)

| Service | Health Check | Startup Time | Memory Limit |
|---------|-------------|--------------|--------------|
| host-agent | HTTP /health | < 10s | 512MB |
| capture-service | HTTP /health | < 5s | 256MB |
| encoder-service | HTTP /health + GPU | < 15s | 1GB |
| discovery-beacon | HTTP /health + mDNS | < 5s | 128MB |
| REST gateway | HTTP /health | < 5s | 256MB |
| cockroachdb | SQL SELECT 1 | < 30s | 1GB |
| redis | PING | < 2s | 256MB |
| nats | HTTP /healthz | < 5s | 512MB |
| vault | HTTP /v1/sys/health | < 10s | 256MB |

---

## Risk Register

| ID | Risk | Probability | Impact | Mitigation |
|----|------|-------------|--------|------------|
| RSK-01 | `replace` directives cannot be resolved for some submodules | High | High | Fork external deps to HelixDevelopment/ org; publish internal deps |
| RSK-02 | WebRTC Pion v4 has breaking API changes | Medium | High | Pin version in go.mod; integration tests catch breakage |
| RSK-03 | GPU container access fails on some host configs | Medium | High | Document GPU driver requirements; provide CPU fallback |
| RSK-04 | Flutter FFI build fails on iOS (codesigning) | Medium | Medium | CI codesigning setup; manual build instructions |
| RSK-05 | Go-WASM size exceeds browser limits | Medium | Medium | Tree shaking; lazy loading; split into chunks |
| RSK-06 | CockroachDB multi-region not needed for MVP | Low | Low | Use single-node for P00-P06; migrate later |
| RSK-07 | Anti-bluff scan produces false positives | Medium | Low | Tune scanner rules; allow EXCEPTION comments |
| RSK-08 | 29 submodules create excessive CI time | Medium | Medium | Parallel CI lanes; GOCACHEPROG; selective testing |
| RSK-09 | Quick resume save/restore fails for some games | High | Medium | Per-game compatibility matrix; graceful degradation |
| RSK-10 | Latency target (30ms LAN) not achievable on all hardware | Medium | High | Document hardware requirements; quality profiles |

---

## Appendix A: Complete File Inventory (P00–P06)

### Created Files Summary

| Phase | Files Created | Files Modified |
|-------|--------------|----------------|
| P00 | 12 | 4 |
| P01 | 18 | 2 |
| P02 | 32 | 5 |
| P03 | 24 | 2 |
| P04 | 20 | 2 |
| P05 | 42 | 5 |
| P06 | 22 | 8 |
| **Total** | **170** | **28** |

### Critical Path Dependencies

```
P00 (Foundation)
  └── P01 (Containers)
        └── P02 (Submodules)
              ├── P03 (Backend)
              │     └── P04 (Streaming)
              │           └── P06 (Host Agent)
              └── P05 (Clients)
                    └── P06 (Host Agent)
```

P06 depends on BOTH P04 (streaming pipeline) AND P05 (clients). These can be developed in parallel after P02 completes.

### Test Matrix (P00–P06)

| Test Type | P00 | P01 | P02 | P03 | P04 | P05 | P06 |
|-----------|-----|-----|-----|-----|-----|-----|-----|
| Unit | 0 | 8 | 50+ | 30+ | 40+ | 50+ | 40+ |
| Integration | 0 | 4 | 15+ | 15+ | 12+ | 15+ | 20+ |
| E2E | 0 | 0 | 0 | 2 | 2 | 5 | 5 |
| Security | 0 | 0 | 29 | 5 | 5 | 5 | 5 |
| Benchmark | 0 | 0 | 4 | 2 | 5 | 2 | 4 |
| **Total** | **0** | **12** | **98+** | **54+** | **64+** | **72+** | **74+** |

---

## Appendix B: Constitution Clause-to-Phase Traceability Matrix

| Clause | P00 | P01 | P02 | P03 | P04 | P05 | P06 |
|--------|-----|-----|-----|-----|-----|-----|-----|
| R-01 Anti-Bluff | T07 | T07 | All | All | All | All | T06 |
| R-02 Decoupling | T05 | — | T04 | — | — | T01 | — |
| R-03 SIV | — | — | T04 | — | — | — | — |
| R-04 go.work | T03 | — | T05 | — | — | — | — |
| R-05 Containers | — | All | — | — | — | — | T01 |
| R-06 Container Matrix | — | All | — | — | — | — | — |
| R-07 Comm Stack | — | — | T02,T03 | All | All | T01,T05 | T03 |
| R-08 Concurrency | — | — | T01 | — | T01,T04 | — | — |
| R-09 Testing | T02 | T01 | All | All | All | All | All |
| R-10 Negative-Leg | — | — | — | — | — | — | T06 |
| R-11 Quality Gates | — | T07 | T06 | — | — | — | — |
| R-12 Tracking | T06 | — | — | — | — | — | — |
| R-13 Four-Mirror | T04 | — | — | — | — | — | — |
| R-14 Observability | — | — | — | T01 | — | — | T01 |
| R-15 Transitive | T05 | — | T05 | — | — | — | — |
| R-16 Security | T08 | — | — | T04 | — | — | — |
| R-17 Documentation | T01 | — | — | — | — | — | — |
| R-18 SafeExec | T07 | T06 | — | — | — | — | T04 |

---

*End of MAIN BODY (Sections 1–4)*
*This document is a living specification. All changes require GitHub PR + GitLab MR with dual approval.*
-e 

---


# HelixPlay Cloud Gaming Platform - Advanced Phases & Architecture Deep-Dive

**Document ID:** `HP-IMPL-ADV-001`
**Version:** 1.0.0
**Status:** Implementation Plan - Advanced Phases (P07-P13) & Architecture Deep-Dive
**Date:** 2025-01
**Classification:** Engineering Implementation Specification

---

## Table of Contents

- [Part A: Implementation Phases P07-P13](#part-a-implementation-phases-p07-p13)
  - [Phase 07: Latency Optimization](#phase-07-latency-optimization-p1)
  - [Phase 08: Audio Surround](#phase-08-audio-surround-p2)
  - [Phase 09: Recording & Replay](#phase-09-recording--replay-p2)
  - [Phase 10: Monetization & Auth](#phase-10-monetization--auth-p2)
  - [Phase 11: Hardening & Security](#phase-11-hardening--security-p2)
  - [Phase 12: Beta Launch](#phase-12-beta-launch-p2)
  - [Phase 13: GA Release](#phase-13-ga-release-p1)
- [Part B: Architecture Deep-Dive Implementation Tasks](#part-b-architecture-deep-dive-implementation-tasks)
  - [B1: Streaming Pipeline Implementation](#b1-streaming-pipeline-implementation)
  - [B2: Controller Input Pipeline Implementation](#b2-controller-input-pipeline-implementation)
  - [B3: Capture & Encode Pipeline Implementation](#b3-capture--encode-pipeline-implementation)
  - [B4: Client Architecture Implementation](#b4-client-architecture-implementation)
  - [B5: Catalog & Content Pipeline](#b5-catalog--content-pipeline)
  - [B6: White-Label & Theming](#b6-white-label--theming)
  - [B7: Operations & Observability](#b7-operations--observability)
- [Appendix A: Submodule Registry](#appendix-a-submodule-registry)
- [Appendix B: Dependency Graph](#appendix-b-dependency-graph)

---

## Part A: Implementation Phases P07-P13

### Legend

| Field | Meaning |
|-------|---------|
| **Priority** | P1 = Critical path (blocks GA), P2 = Important (feature-complete), P3 = Enhancement |
| **Effort** | Person-weeks (pw), Person-days (pd) |
| **AC** | Acceptance Criteria |
| **Files** | Files to create (+) or modify (~) |
| **Deps** | Cross-phase dependencies |

---

### Phase 07: Latency Optimization (P1)

**Phase Goal:** Achieve sub-30ms LAN p999 and sub-50ms WAN p999 glass-to-glass latency through kernel-level tuning, zero-copy IPC, RTOS scheduling, and lock-free data structures.

**Phase Duration:** 8 weeks
**Engineering Team:** 4 FTE (2 systems, 1 kernel, 1 Go performance)
**Deps:** P01-P06 (foundation infrastructure operational)

---

#### Task P07-T01: PREEMPT_RT Kernel Patching & Validation

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T01 |
| **Title** | PREEMPT_RT Linux Kernel Patching for Host Fleet |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Kernel Engineer |

**Description:**
Patch host fleet Linux kernels with PREEMPT_RT realtime patchset. Validate scheduling determinism under full gaming load. Build CI pipeline for kernel image generation and automated boot testing.

**Files:**
```
+ kernel/patches/preempt-rt-helix.patch       # Helix-specific RT tuning
+ kernel/configs/helix-host-x86_64.defconfig   # Defconfig for host nodes
+ kernel/scripts/build-kernel.sh               # Automated kernel build
+ kernel/scripts/validate-rt.sh                # cyclictest validation (>100k samples)
~ kernel/Makefile                              # Add kernel artifact targets
~ .github/workflows/kernel-build.yml           # CI: kernel build + cyclictest
```

**Implementation Details:**
- Base kernel: 6.6 LTS with PREEMPT_RT patchset
- Target maximum scheduling latency: <10 microseconds (cyclictest histogram 99.9th percentile)
- Disable CFS bandwidth control for `helix-rt` cgroup
- Enable `CONFIG_NO_HZ_FULL` for isolated CPU cores
- Disable `CONFIG_CPU_FREQ_DEFAULT_GOV_SCHEDUTIL`; use `performance` governor
- Apply `rcu_nocbs` for isolated cores
- Use `tuned` profile `latency-performance` as base, customize for HelixPlay

**Acceptance Criteria:**
1. `cyclictest -D 3600 -m -S -p 90 -i 200 -h 400` shows p99.9 < 10us under 4K60 capture+encode load
2. Kernel boots successfully on 100% of host fleet hardware profiles (AWS g4dn/g5, local RTX nodes)
3. Kernel build completes in CI < 15 minutes from clean state
4. No kernel panics or soft-lockups during 72-hour burn-in test
5. `CONFIG_DEBUG_PREEMPT` disabled for production builds (performance)

---

#### Task P07-T02: SCHED_FIFO Thread Priorities & CPU Isolation

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T02 |
| **Title** | Real-Time Thread Scheduling & CPU Isolation (isolcpus) |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Systems Engineer |

**Description:**
Implement `SCHED_FIFO` thread priority assignment across all HelixPlay host processes. Isolate critical-path threads to dedicated CPU cores using `isolcpus` and `cset shield`. Define priority hierarchy for capture → encode → network → input → housekeeping.

**Files:**
```
+ helix-rtos/pkg/rtsched/priority.go          # Priority hierarchy definitions
+ helix-rtos/pkg/rtsched/affinity.go          # CPU affinity/ isolation management
+ helix-rtos/pkg/rtsched/rtsched.go           # Main RT scheduler interface
+ helix-rtos/cmd/rtsetup/main.go              # Host node RT setup utility
+ helix-rtos/configs/priority-defaults.yaml    # Per-process priority defaults
~ sunshine/src/platform/linux/misc.cpp         # Sunshine RT thread hooks
~ sunshine/src/platform/linux/display.cpp      # Capture thread affinity
```

**Thread Priority Hierarchy:**

| Priority | Thread | sched_policy | CPU Affinity |
|----------|--------|-------------|--------------|
| 99 | Capture frame acquisition | `SCHED_FIFO` | `isolcpus` core 0 |
| 97 | Video encode (NVENC submit) | `SCHED_FIFO` | `isolcpus` core 0 |
| 95 | Network TX (WebRTC packet send) | `SCHED_FIFO` | `isolcpus` core 1 |
| 93 | Controller input read | `SCHED_FIFO` | `isolcpus` core 1 |
| 90 | Audio capture/encode | `SCHED_FIFO` | `isolcpus` core 2 |
| 80 | WebRTC ICE/STUN processing | `SCHED_FIFO` | `isolcpus` core 2 |
| 50 | Session orchestration | `SCHED_OTHER` | General pool |
| 10 | Metrics, logging, housekeeping | `SCHED_IDLE` | General pool |

**Acceptance Criteria:**
1. `schedtool` confirms all threads at assigned priorities during active session
2. `taskset -pc <pid>` shows correct CPU isolation for capture/encode threads
3. No priority inversion events detected by `trace-cmd` during 1-hour session
4. Latency variance (stddev) reduced by >40% vs. CFS baseline at same load
5. Dynamic priority adjustment API available for thermal throttling scenarios

---

#### Task P07-T03: GOCACHEPROG Remote Build Caching

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T03 |
| **Title** | GOCACHEPROG Remote Build Cache (Bazel-remote / Turborepo) |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Build Engineer |

**Description:**
Implement `GOCACHEPROG` remote build cache to reduce CI build times across 29 submodules. Deploy bazel-remote or equivalent S3-backed cache. Configure all CI pipelines to use shared cache.

**Files:**
```
+ .github/scripts/gocacheprog.sh              # GOCACHEPROG wrapper
+ infra/cache/bazel-remote.yml                # Docker Compose for cache server
+ infra/cache/s3-backend.tf                   # S3 backend Terraform
~ .github/workflows/ci-all.yml                # Add GOCACHEPROG env
~ Makefile                                     # Export GOCACHEPROG
```

**Acceptance Criteria:**
1. Cold CI build (all 29 modules) completes in < 8 minutes with warm cache
2. Cache hit rate > 85% for incremental builds
3. Cache backend supports 1000+ concurrent CI jobs (horizontal scaling)
4. Fallback to local build on cache miss (no hard dependency)
5. Cache eviction policy: 7-day TTL, LRU within TTL

---

#### Task P07-T04: Kernel Parameter Tuning & sysctl Profiles

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T04 |
| **Title** | Kernel sysctl Tuning for Low-Latency Networking & IPC |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Systems Engineer |

**Description:**
Create tuned sysctl profiles optimized for low-latency gaming workloads. Tune network buffers, TCP/UDP parameters, memory management, and scheduler settings.

**Files:**
```
+ helix-rtos/configs/sysctl-gaming.conf        # Core gaming sysctl profile
+ helix-rtos/configs/sysctl-udp-optim.conf     # UDP/WebRTC optimization
+ helix-rtos/configs/sysctl-memory.conf        # Memory management tuning
+ helix-rtos/cmd/tune-host/main.go             # Host tuning application
```

**Key sysctl Parameters:**
```
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.udp_rmem_min = 1048576
net.ipv4.udp_wmem_min = 1048576
net.core.netdev_max_backlog = 65536
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_notsent_lowat = 16384
vm.swappiness = 1
vm.dirty_ratio = 5
vm.dirty_background_ratio = 2
kernel.sched_rt_runtime_us = 950000
kernel.timer_migration = 0
```

**Acceptance Criteria:**
1. UDP socket buffer sizes confirmed via `ss -npm` during active session
2. `netperf` UDP_RR latency < 50us p99 localhost
3. No packet drops at 100Mbps sustained UDP send rate
4. Tuning profiles applied automatically on host node bootstrap

---

#### Task P07-T05: io_uring Integration (helix-iouring submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T05 |
| **Title** | io_uring Asynchronous I/O Integration |
| **Priority** | P1 |
| **Effort** | 2.5pw |
| **Owner** | Systems Engineer |

**Description:**
Implement io_uring-based async I/O for file recording, network send/receive, and shared memory operations. Create reusable Go wrapper via `x/sys/unix` raw syscalls (no CGO dependency).

**Files:**
```
+ helix-iouring/go.mod                        # Module: github.com/helixplay/helix-iouring
+ helix-iouring/pkg/uring/ring.go             # io_uring ring management
+ helix-iouring/pkg/uring/sqe.go              # Submission queue entry builders
+ helix-iouring/pkg/uring/cqe.go              # Completion queue event handlers
+ helix-iouring/pkg/uring/ops/read.go         # vectored read operations
+ helix-iouring/pkg/uring/ops/write.go        # vectored write operations
+ helix-iouring/pkg/uring/ops/send.go         # network send operations
+ helix-iouring/pkg/uring/ops/recv.go         # network receive operations
+ helix-iouring/pkg/uring/pool/buffer.go      # registered buffer pool
+ helix-iouring/pkg/uring/bench/bench_test.go # Benchmarks vs. epoll/sync
+ helix-iouring/LICENSE                       # MIT
```

**API Surface:**
```go
type Ring struct { /* io_uring ring */ }
func Setup(entries uint, params *Params) (*Ring, error)
func (r *Ring) QueueReadv(fd int, iovec []syscall.Iovec, offset uint64) error
func (r *Ring) QueueWritev(fd int, iovec []syscall.Iovec, offset uint64) error
func (r *Ring) QueueSend(fd int, buf []byte, flags int) error
func (r *Ring) QueueRecv(fd int, buf []byte, flags int) error
func (r *Ring) Submit() (uint, error)
func (r *Ring) WaitCQEs(count uint, timeout time.Duration) ([]CQE, error)
```

**Acceptance Criteria:**
1. `io_uring` ring operates in polled mode (`IORING_SETUP_IOPOLL`) for NVMe I/O
2. Vectored write throughput to NVMe > 3GB/s sustained (single thread)
3. Latency of io_uring write completion < 2us p99 for 4KB writes
4. Registered buffer pool (`IORING_REGISTER_BUFFERS`) reduces CPU by >20%
5. Graceful fallback to `epoll`/`select` on kernel < 5.10
6. Benchmark suite shows io_uring outperforms sync I/O by >3x in all metrics

---

#### Task P07-T06: Lock-Free Data Structures (helix-lockfree submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T06 |
| **Title** | Lock-Free SPSC/MPSC Ring Buffers & Queues |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Performance Engineer |

**Description:**
Implement platform-native lock-free data structures in Go using atomic operations and memory barriers. Focus on SPSC (Single Producer Single Consumer) ring buffers for capture→encode handoff and MPSC (Multi Producer Single Consumer) queues for controller input aggregation.

**Files:**
```
+ helix-lockfree/go.mod                       # Module: github.com/helixplay/helix-lockfree
+ helix-lockfree/pkg/ring/spsc.go             # SPSC ring buffer (power-of-2)
+ helix-lockfree/pkg/ring/mpsc.go             # MPSC bounded queue
+ helix-lockfree/pkg/ring/mpmc.go             # MPMC bounded queue (fallback)
+ helix-lockfree/pkg/atomic/seqlock.go        # Sequence locks for metrics
+ helix-lockfree/pkg/atomic/snapshot.go       # Lock-free snapshot reader
+ helix-lockfree/pkg/cacheline/pad.go         # Cache-line padding utilities
+ helix-lockfree/pkg/bench/spsc_bench_test.go # Go benchmarks
+ helix-lockfree/internal/asm/spsc_amd64.s    # Assembly-optimized x86_64
+ helix-lockfree/internal/asm/spsc_arm64.s    # Assembly-optimized ARM64
+ helix-lockfree/LICENSE                      # MIT
```

**SPSC Ring Buffer Spec:**
```go
type SPSCRing[T any] struct {
    _pad0   [cacheline.Size]byte
    head    atomic.Uint64       // producer index
    _pad1   [cacheline.Size]byte
    tail    atomic.Uint64       // consumer index
    _pad2   [cacheline.Size]byte
    mask    uint64
    buffer  unsafe.Pointer      // contiguous array
}

func NewSPSC[T any](size uint) *SPSCRing[T]
func (r *SPSCRing[T]) Push(val T) bool    // false if full (non-blocking)
func (r *SPSCRing[T]) Pop() (T, bool)     // zero-value if empty (non-blocking)
func (r *SPSCRing[T]) Available() uint
```

**Acceptance Criteria:**
1. `SPSC.Push` latency < 15ns p99 (measured via `testing.B` + `perfevents`)
2. `SPSC.Pop` latency < 15ns p99
3. No memory allocations in hot path (verified via `go test -memprofile`)
4. Cache-line false sharing eliminated (verified via `perf c2c`)
5. Correctness validated with `go test -race` and stress test (100M ops, 32 threads)
6. AMD64 assembly path shows >15% improvement over pure Go on x86_64

---

#### Task P07-T07: GPU Direct Integration (helix-gpu-direct submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T07 |
| **Title** | GPU Direct (GPUDirect RDMA / Resizable BAR) |
| **Priority** | P2 |
| **Effort** | 2pw |
| **Owner** | GPU Systems Engineer |

**Description:**
Enable zero-copy frame transfer from GPU framebuffer to encoder via GPUDirect RDMA (NVIDIA) and Resizable BAR. Eliminate PCIe round-trip for frame capture path.

**Files:**
```
+ helix-gpu-direct/go.mod                     # Module: github.com/helixplay/helix-gpu-direct
+ helix-gpu-direct/pkg/nv/gdrdrv.go           # NVIDIA gdrdrv bindings
+ helix-gpu-direct/pkg/nv/p2p.go              # Peer-to-peer memory mapping
+ helix-gpu-direct/pkg/amd/amdgpudirect.go    # AMD GPU Direct equivalent
+ helix-gpu-direct/pkg/bar/resizable.go       # Resizable BAR management
+ helix-gpu-direct/pkg/common/mapping.go      # Generic GPU memory mapping
+ helix-gpu-direct/pkg/common/barrier.go      # Memory barrier primitives
+ helix-gpu-direct/cmd/gpudirect-test/main.go # Validation tool
+ helix-gpu-direct/LICENSE                    # MIT
```

**Architecture:**
- NVIDIA: Use `nvidia_p2p_get_pages()` + `gdrdrv` kernel module for pinned GPU memory export
- AMD: Use `amdgpu_ttm_tt_get_user_pages_done()` for ROCm-based direct access
- Intel: Use Level Zero `zeMemGetAllocProperties()` for shared allocations
- Fallback: `cudaMemcpy2DAsync` + registered host memory (1 copy instead of 2)

**Acceptance Criteria:**
1. Frame capture→encode path eliminates host memory copy (verified via NVIDIA Nsight Systems)
2. GPU→encoder latency < 0.5ms per frame (4K)
3. Works on RTX 40-series (GPUDirect), RTX 30-series (resizable BAR), Intel Arc (Level Zero)
4. Graceful fallback to pinned host memory path on unsupported hardware
5. Memory pinning respects host RAM budget (cgroup v2 memory limits)

---

#### Task P07-T08: Shared Memory Zero-Copy IPC (helix-shm submodule)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T08 |
| **Title** | Shared Memory Zero-Copy Inter-Process Communication |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Systems Engineer |

**Description:**
Implement POSIX shared memory-based zero-copy IPC for frame data and controller input between Sunshine++ host agent processes. Replace pipe/socket-based transfer with shared memory regions + atomic synchronization.

**Files:**
```
+ helix-shm/go.mod                            # Module: github.com/helixplay/helix-shm
+ helix-shm/pkg/shm/region.go                 # Shared memory region management
+ helix-shm/pkg/shm/segment.go                # Memory segment allocator
+ helix-shm/pkg/shm/ring.go                   # SHM-backed ring buffer
+ helix-shm/pkg/shm/sync.go                   # Atomic synchronization primitives
+ helix-shm/pkg/shm/frame.go                  # Frame buffer shared memory layout
+ helix-shm/pkg/shm/input.go                  # Input packet shared memory layout
+ helix-shm/pkg/shm/cgroup.go                 # cgroup v2 memory integration
+ helix-shm/cmd/shm-test/main.go              # SHM validation tool
+ helix-shm/internal/unix/shm_linux.go        # Linux-specific shm syscalls
+ helix-shm/LICENSE                           # MIT
```

**Frame Buffer Layout:**
```go
type FrameBuffer struct {
    Magic       uint32          // 'HLXF'
    Version     uint16          // 1
    Format      uint32          // DXGI_FORMAT / VK_FORMAT
    Width       uint32
    Height      uint32
    Pitch       uint32
    HDRMetadata [64]byte        // HDR10 metadata blob
    ProducerSeq atomic.Uint64   // write sequence number
    ConsumerSeq atomic.Uint64   // read sequence number
    Data        [0]byte         // frame data follows (VLA pattern)
}
```

**Acceptance Criteria:**
1. Frame transfer latency (capture → encoder) < 50 microseconds
2. Shared memory regions properly cleaned up on process crash (`memfd_create` + `MFD_CLOEXEC`)
3. cgroup v2 `memory.max` respected by SHM allocator
4. Works across privilege boundaries (unprivileged client ↔ privileged capture daemon)
5. `ftruncate`/`mmap` cycle time < 1ms for 4K frame buffer allocation
6. Verified with `strace -e trace=mmap,munmap` showing zero memcpy in hot path

---

#### Task P07-T09: Memory Pool Optimization (helix-mempool, helix-allocator)

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T09 |
| **Title** | Slab Memory Pool & Custom Allocator |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Performance Engineer |

**Description:**
Implement slab-based memory pools for fixed-size allocations (frame buffers, network packets, controller state) and a bump allocator for short-lived session data. Eliminate GC pressure in hot paths.

**Files:**
```
+ helix-mempool/go.mod                        # Module: github.com/helixplay/helix-mempool
+ helix-mempool/pkg/slab/pool.go              # Generic slab allocator
+ helix-mempool/pkg/slab/frame.go             # Frame buffer pool (fixed 4K/1080p)
+ helix-mempool/pkg/slab/packet.go            # Network packet pool
+ helix-mempool/pkg/slab/controller.go        # Controller state pool
+ helix-mempool/pkg/bump/arena.go             # Bump allocator for session data
+ helix-mempool/pkg/metrics/gc.go             # GC pressure monitoring
+ helix-mempool/LICENSE                       # MIT
```

**Files (helix-allocator):**
```
+ helix-allocator/go.mod                      # Module: github.com/helixplay/helix-allocator
+ helix-allocator/pkg/linear/allocator.go     # Linear/bump allocator
+ helix-allocator/pkg/linear/region.go        # Memory region manager
+ helix-allocator/pkg/linear/sync.go          # Thread-safe region operations
+ helix-allocator/LICENSE                     # MIT
```

**Pool Configuration:**
```go
var DefaultPools = map[string]PoolConfig{
    "frame-4k":    {Size: 3840 * 2160 * 4, Count: 4, NUMA: true},
    "frame-1080p": {Size: 1920 * 1080 * 4, Count: 8, NUMA: true},
    "packet":      {Size: 2048, Count: 1024, NUMA: false},
    "controller":  {Size: 256, Count: 16, NUMA: false},
}
```

**Acceptance Criteria:**
1. Zero heap allocations during active streaming session (`GODEBUG=allocfreetrace=1`)
2. GC pause times < 100 microseconds during 1-hour session
3. `slab.Alloc` / `slab.Free` operations < 20ns p99
4. NUMA-aware allocation on dual-socket host nodes
5. Graceful OOM handling with frame drop instead of process termination

---

#### Task P07-T10: IRQ Affinity & Network Interrupt Steering

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T10 |
| **Title** | IRQ Affinity Tuning & Network Interrupt Steering |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Systems Engineer |

**Description:**
Configure IRQ affinity for network interfaces to isolated cores. Use `irqbalance` exclusion or manual `/proc/irq/*/smp_affinity` configuration. Enable NIC hardware flow steering if available.

**Files:**
```
+ helix-rtos/configs/irq-affinity.rules        # IRQ affinity rules
+ helix-rtos/scripts/setup-irq.sh              # IRQ setup script
~ helix-rtos/cmd/tune-host/main.go             # Add IRQ tuning
```

**Acceptance Criteria:**
1. NIC RX/TX interrupts isolated to non-`isolcpus` cores
2. No network IRQ handled on capture/encode isolated cores
3. Hardware flow steering (`ethtool -n rx-flow-hash`) configured if NIC supports
4. `mpstat -I CPU` shows <1% interrupt time on isolated cores

---

#### Task P07-T11: cgroup v2 Resource Control

| Attribute | Detail |
|-----------|--------|
| **ID** | P07-T11 |
| **Title** | cgroup v2 Session Resource Isolation |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Systems Engineer |

**Description:**
Implement cgroup v2 resource control for per-session resource isolation. Enforce CPU, memory, and I/O limits. Use systemd slice integration for automatic cleanup.

**Files:**
```
+ helix-rtos/pkg/cgroup/v2.go                  # cgroup v2 controller
+ helix-rtos/pkg/cgroup/session.go             # Per-session cgroup management
+ helix-rtos/pkg/cgroup/systemd.go             # systemd slice integration
```

**Acceptance Criteria:**
1. Each session runs in dedicated cgroup with configurable limits
2. `memory.high` enforcement triggers graceful quality reduction (not OOM kill)
3. `cpu.weight` allows proportional sharing without hard caps
4. Automatic cgroup cleanup on session termination (systemd `After=`)

---

#### Phase 07: Definition of Done

- [ ] All P1 tasks (T01, T02, T04-T06, T08, T09, T10) completed and passing AC
- [ ] io_uring benchmarks show >3x improvement over sync I/O
- [ ] Lock-free SPSC ring buffer < 15ns p99 push/pop latency
- [ ] SHM IPC achieves < 50us frame transfer latency
- [ ] cyclictest p99.9 < 10us under full load
- [ ] Zero heap allocations during active streaming session
- [ ] p999 glass-to-glass latency: LAN < 30ms, WAN < 50ms (measured)

---

### Phase 08: Audio Surround (P2)

**Phase Goal:** Deliver immersive multi-channel audio with Opus MultiStream surround (up to 7.1), Dolby Atmos spatial audio, AC3/EAC3 passthrough, and eARC support. Achieve <5ms A/V sync drift.

**Phase Duration:** 6 weeks
**Engineering Team:** 3 FTE (1 audio specialist, 1 Go backend, 1 client)
**Deps:** P04 (client architecture), P07 (latency optimization)

---

#### Task P08-T01: Opus MultiStream Encoder (up to 7.1 channels)

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T01 |
| **Title** | Opus MultiStream Surround Encoder Integration |
| **Priority** | P2 |
| **Effort** | 2pw |
| **Owner** | Audio Engineer |

**Description:**
Integrate Opus MultiStream API (`opus_multistream_encoder_create`) for surround sound encoding up to 7.1 channels. Implement channel mapping families 0, 1 (Vorbis), and 255 (raw/manual). Support 48kHz sample rate with 20ms frame size.

**Files:**
```
+ helix-audio/go.mod                           # Module: github.com/helixplay/helix-audio
+ helix-audio/pkg/opus/multistream.go          # Opus MultiStream wrapper
+ helix-audio/pkg/opus/surround.go             # Surround channel mappings
+ helix-audio/pkg/opus/encoder.go              # Generic Opus encoder interface
+ helix-audio/pkg/opus/decoder.go              # Opus decoder interface
+ helix-audio/pkg/format/channel.go            # Channel layout definitions
+ helix-audio/pkg/format/layout.go             # Standard layouts (mono→7.1)
+ helix-audio/pkg/pcm/buffer.go                # PCM buffer management
+ helix-audio/pkg/pcm/resample.go              # Speex resampler wrapper
+ helix-audio/pkg/pcm/mix.go                   # Channel mixing utilities
+ helix-audio/LICENSE                          # MIT
```

**Channel Layout Support:**

| Layout | Channels | Opus Mapping Family | Channel Order |
|--------|----------|---------------------|---------------|
| Mono | 1 | 0 | C |
| Stereo | 2 | 0 | L, R |
| 3.0 | 3 | 1 | L, R, C |
| Quadraphonic | 4 | 1 | L, R, Ls, Rs |
| 5.1 | 6 | 1 | L, R, C, LFE, Ls, Rs |
| 7.1 | 8 | 1 | L, R, C, LFE, Ls, Rs, Rls, Rrs |

**Acceptance Criteria:**
1. Opus MultiStream encodes 7.1 @ 48kHz to < 512kbps with transparency (> 128kbps per stereo pair equivalent)
2. All channel mapping families (0, 1, 255) functional and tested
3. Encoder latency < 2x frame size (40ms @ 20ms frames, including look-ahead)
4. Bitrate adaptation without audible artifacts during gameplay
5. Go wrapper provides CGO-free build option via pure-Go Opus fallback (lower quality acceptable)

---

#### Task P08-T02: AC3/EAC3 Passthrough

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T02 |
| **Title** | AC3/EAC3 (Dolby Digital/Digital+) Bitstream Passthrough |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Audio Engineer |

**Description:**
Implement pass-through mode for AC3 and EAC3 compressed audio bitstreams from game/application to client without transcoding. Preserve exact bitstream for external decoder (AVR, soundbar, TV).

**Files:**
```
+ helix-audio/pkg/passthrough/ac3.go           # AC3 passthrough handler
+ helix-audio/pkg/passthrough/eac3.go          # EAC3 passthrough handler
+ helix-audio/pkg/passthrough/detector.go      # Format auto-detection
+ helix-audio/pkg/passthrough/syncframe.go     # Sync frame boundary detection
~ helix-audio/pkg/opus/encoder.go              # Add passthrough bypass mode
```

**Acceptance Criteria:**
1. AC3/EAC3 bitstream passes through unchanged (bit-exact verification)
2. Sync frame boundaries correctly detected and preserved
3. IEC 61937 encapsulation for SPDIF/eARC output on client side
4. Automatic format detection with fallback to Opus transcode
5. Metadata (dialog normalization, DRC) preserved through passthrough

---

#### Task P08-T03: Dolby Atmos Support

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T03 |
| **Title** | Dolby Atmos Spatial Audio (DD+JOC) |
| **Priority** | P2 |
| **Effort** | 1.5pw |
| **Owner** | Audio Engineer |

**Description:**
Implement Dolby Atmos support via EAC3 with Joint Object Coding (JOC). Encode object-based audio metadata alongside channel bed. Client-side rendering via Dolby MS12 decoder or platform Atmos renderer.

**Files:**
```
+ helix-audio/pkg/atmos/joc.go                 # JOC metadata parser
+ helix-audio/pkg/atmos/object.go              # Audio object extraction
+ helix-audio/pkg/atmos/renderer.go            # Platform Atmos renderer abstraction
+ helix-audio/pkg/atmos/header.go              # Atmos-specific header handling
```

**Acceptance Criteria:**
1. Atmos content identified and routed correctly in pipeline
2. JOC metadata preserved through encoding chain
3. Platform Atmos rendering works on: Apple TV 4K (tvOS), Android TV (API 30+), Windows Sonic
4. Fallback to 7.1 channel bed when Atmos renderer unavailable

---

#### Task P08-T04: eARC Audio Return Channel

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T04 |
| **Title** | eARC (Enhanced Audio Return Channel) Client Support |
| **Priority** | P3 |
| **Effort** | 1pw |
| **Owner** | Client Engineer |

**Description:**
Support eARC output on TV/console clients for uncompressed multi-channel audio passthrough to external audio systems. Implement HDMI eARC handshake and latency compensation.

**Files:**
```
~ helix-flutter/lib/platform/tv/earc.dart      # eARC TV platform channel
~ helix-flutter/android/src/main/kotlin/Ear.kt  # Android eARC native
~ helix-wails/frontend/src/audio/earc.ts        # Desktop eARC (if applicable)
```

**Acceptance Criteria:**
1. eARC handshake completes successfully on HDMI 2.1 ports
2. Uncompressed 7.1 PCM output at 48kHz to eARC receiver
3. Lip-sync adjustment within +/- 2ms via eARC latency reporting
4. Automatic fallback to TV speakers on eARC disconnect

---

#### Task P08-T05: A/V Synchronization Mechanism

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T05 |
| **Title** | Audio/Video Sync with Frame-Time Alignment |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Streaming Engineer |

**Description:**
Implement A/V synchronization using RTP timestamp alignment and frame-time audio rendering. Target < 5ms A/V drift with automatic drift correction.

**Files:**
```
+ helix-stream/pkg/sync/avsync.go              # A/V sync controller
+ helix-stream/pkg/sync/drift.go               # Drift detection & correction
+ helix-stream/pkg/sync/timestamp.go           # RTP timestamp alignment
+ helix-audio/pkg/render/clock.go              # Audio clock reference
```

**Sync Strategy:**
1. **Presentation Timestamps:** Both audio and video RTP packets carry capture-time NTP timestamps
2. **Audio Master:** Audio clock drives presentation (less variable than video)
3. **Drift Detection:** Compare expected vs. actual audio sample consumption every 100ms
4. **Correction:** Sample rate adjustment (±2%) or audio sample insertion/drop for drift > 2ms
5. **Video Sync:** Video frames delayed/early-released to match audio presentation time

**Acceptance Criteria:**
1. A/V sync drift < 5ms sustained (measured with audio-led test pattern)
2. No audible pops/clicks during drift correction
3. Sync recovers within 200ms after network jitter event
4. Sync mechanism works across all codec combinations (H.264/HEVC/AV1 + Opus)

---

#### Task P08-T06: Audio Pipeline Integration into Sunshine++

| Attribute | Detail |
|-----------|--------|
| **ID** | P08-T06 |
| **Title** | Sunshine++ Audio Pipeline Integration |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Audio Engineer |

**Description:**
Integrate the helix-audio module into Sunshine++ host agent. Replace existing audio pipeline with Opus MultiStream surround, passthrough detection, and Atmos routing.

**Files:**
```
~ sunshine/src/audio.cpp                       # Refactor to use helix-audio
~ sunshine/src/audio.h                         # Audio interface updates
~ sunshine/src/stream.cpp                      # A/V sync integration
~ sunshine/src/entry_handler.cpp               # Audio format negotiation
```

**Acceptance Criteria:**
1. Audio pipeline functional end-to-end with all channel configurations
2. Passthrough correctly detected and routed without transcoding
3. A/V sync < 5ms in end-to-end test
4. No audio artifacts during 1-hour stress test

---

#### Phase 08: Definition of Done

- [ ] Opus MultiStream surround encoding up to 7.1 channels functional
- [ ] AC3/EAC3 passthrough bit-exact verified
- [ ] Dolby Atmos detection and routing complete
- [ ] A/V sync drift < 5ms sustained
- [ ] Audio latency added to glass-to-glass budget and within target

---

### Phase 09: Recording & Replay (P2)

**Phase Goal:** Enable gameplay recording with dual-path encoding (stream + record simultaneously), instant replay buffer, local NVMe storage with background cloud sync, and DASH-based replay client.

**Phase Duration:** 6 weeks
**Engineering Team:** 3 FTE (1 video pipeline, 1 storage/sync, 1 client replay)
**Deps:** P03 (capture pipeline), P07 (latency optimization), P08 (audio pipeline)

---

#### Task P09-T01: helix-record Submodule Bootstrap

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T01 |
| **Title** | helix-record Module Creation & CI Pipeline |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Build Engineer |

**Files:**
```
+ helix-record/go.mod                          # Module: github.com/helixplay/helix-record
+ helix-record/.github/workflows/ci.yml        # CI: build + test + lint
+ helix-record/.github/workflows/release.yml   # Release automation
+ helix-record/Makefile                        # Standard targets
+ helix-record/LICENSE                         # MIT
```

**Acceptance Criteria:**
1. Module compiles with `go build ./...`
2. CI passes on Go 1.23 (linux/amd64, linux/arm64)
3. Test coverage > 70% from inception (enforced in CI)

---

#### Task P09-T02: Dual-Path Encoding (Stream + Record)

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T02 |
| **Title** | Dual-Path Simultaneous Encode (1 NVENC Session Constraint) |
| **Priority** | P1 |
| **Effort** | 2.5pw |
| **Owner** | Video Pipeline Engineer |

**Description:**
Implement dual-path encoding where a single NVENC session produces both the low-latency stream and the high-quality recording using split encoding (SEI NAL units for frame duplication) or B-frame enabled recording path. The constraint: only 1 NVENC session per GPU to avoid resource contention.

**Files:**
```
+ helix-record/pkg/encode/dualpath.go          # Dual-path encoder controller
+ helix-record/pkg/encode/split.go             # Frame splitting strategy
+ helix-record/pkg/encode/quality.go           # Recording quality presets
+ helix-record/pkg/nvenc/split.go              # NVENC split encoding
+ helix-record/pkg/qsv/split.go                # QSV split encoding (Intel)
+ helix-record/pkg/amf/split.go                # AMF split encoding (AMD)
+ sunshine/src/video.cpp                       # Add dual-path hooks
```

**Dual-Path Strategies:**

| Strategy | NVENC Sessions | Stream Quality | Record Quality | Overhead |
|----------|---------------|----------------|----------------|----------|
| **Split Encode** | 1 | Low-latency (P-frames only) | High (B-frames, higher bitrate) | ~5% GPU |
| **Frame Copy** | 2 (not allowed) | Native | Native | N/A |
| **Post-Process** | 1 | Low-latency | Re-encode from saved frames | ~30% GPU |

**Selected Architecture:** Split Encode (1 NVENC session, dual output)
- Primary output: low-latency stream (low bitrate, no B-frames)
- Secondary output: recording (high bitrate, B-frames enabled, higher quality preset)
- Frame duplication handled by NVENC hardware (minimal overhead)

**Acceptance Criteria:**
1. Single NVENC session produces both stream and recording
2. Stream latency increase < 2ms vs. single-path encoding
3. Recording quality: 50Mbps HEVC Main10 @ 4K60 (visually lossless)
4. GPU overhead < 10% vs. stream-only operation
5. Graceful degradation: if GPU loaded, recording pauses (stream priority)

---

#### Task P09-T03: MKV/fMP4 Container Support

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T03 |
| **Title** | MKV (Matroska) and fMP4 (CMAF) Container Muxing |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Video Pipeline Engineer |

**Description:**
Implement MKV and fMP4 container muxing for recordings. MKV for local archival, fMP4 for DASH streaming replay. Support chapter markers (game events) and subtitle tracks.

**Files:**
```
+ helix-record/pkg/mux/mkv.go                  # MKV muxer (libmatroska bindings)
+ helix-record/pkg/mux/fmp4.go                 # fMP4/CMAF muxer
+ helix-record/pkg/mux/chapter.go              # Chapter marker injection
+ helix-record/pkg/mux/metadata.go             # Metadata attachment (game info)
+ helix-record/pkg/mux/segment.go              # fMP4 segment management
```

**Acceptance Criteria:**
1. MKV output playable in VLC, MPC-HC, Kodi, Plex
2. fMP4 segments valid for DASH-IF conformance
3. Chapter markers accurate to within 1 second of in-game event
4. HDR metadata (HDR10/HDR10+/Dolby Vision) preserved in container

---

#### Task P09-T04: Instant Replay Buffer

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T04 |
| **Title** | Rolling Instant Replay Buffer (DVR-style) |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Video Pipeline Engineer |

**Description:**
Implement a configurable rolling buffer (default 30 seconds) that continuously records gameplay to a ring buffer in memory. User can trigger "save replay" to persist buffer to storage.

**Files:**
```
+ helix-record/pkg/replay/buffer.go            # Ring buffer manager
+ helix-record/pkg/replay/segments.go          # Segment ring management
+ helix-record/pkg/replay/trigger.go           # User trigger handler
+ helix-record/pkg/api/replay.go               # Replay API endpoints
```

**Acceptance Criteria:**
1. Rolling buffer maintains last 30 seconds in RAM (configurable 10-300s)
2. "Save replay" action persists buffer to NVMe in < 2 seconds
3. Memory usage bounded: 30s @ 1080p60 ~ 150MB ring buffer
4. No impact on streaming latency (< 1ms addition)

---

#### Task P09-T05: Local NVMe Storage & Background Sync

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T05 |
| **Title** | NVMe Fast Local Storage + Background Cloud Sync |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Storage Engineer |

**Description:**
Implement local NVMe recording storage with background upload to cloud storage (SMB/NFS/FTP/WebDAV). Use io_uring for fast local writes. Implement retry, resume, and bandwidth-throttled sync.

**Files:**
```
+ helix-record/pkg/storage/local.go            # NVMe local storage
+ helix-record/pkg/storage/sync.go             # Background sync orchestrator
+ helix-record/pkg/storage/backends/smb.go     # SMB/CIFS backend
+ helix-record/pkg/storage/backends/nfs.go     # NFS backend
+ helix-record/pkg/storage/backends/ftp.go     # FTP/FTPS backend
+ helix-record/pkg/storage/backends/webdav.go  # WebDAV backend
+ helix-record/pkg/storage/backends/s3.go      # S3-compatible backend
+ helix-record/pkg/storage/queue.go            # Upload queue with retry
+ helix-record/pkg/storage/bandwidth.go        # Bandwidth throttling
```

**Acceptance Criteria:**
1. Local write speed: > 500MB/s sustained to NVMe (io_uring)
2. Background sync at configurable bandwidth limit (default: 50% of available upload)
3. Resume interrupted uploads from last byte (HTTP Range)
4. All 5 backend protocols functional with integration tests
5. Storage cleanup: auto-delete recordings > retention period (configurable)

---

#### Task P09-T06: DASH Replay Client

| Attribute | Detail |
|-----------|--------|
| **ID** | P09-T06 |
| **Title** | DASH-based Replay Streaming Client |
| **Priority** | P2 |
| **Effort** | 1.5pw |
| **Owner** | Client Engineer |

**Description:**
Implement DASH client for replay playback across all three client platforms (Wails, Flutter, Angular+WASM). Support trick-play (seek, fast-forward, rewind) and quality adaptation.

**Files:**
```
+ helix-record/pkg/dash/manifest.go            # MPD manifest parser/generator
+ helix-record/pkg/dash/segment.go             # Segment fetch & buffer
+ helix-record/pkg/dash/adaptation.go          # Quality adaptation logic
+ helix-record/pkg/api/dash.go                 # DASH serving API
~ helix-angular/src/app/replay/               # Web replay player
~ helix-flutter/lib/replay/                    # Mobile replay player
~ helix-wails/frontend/src/replay/             # Desktop replay player
```

**Acceptance Criteria:**
1. DASH playback starts in < 2 seconds (first segment loaded)
2. Seek response < 500ms (to any point in recording)
3. Trick-play at 2x, 4x, 8x speeds without decoder reinitialization
4. Quality adaptation responds to network conditions (ABR)
5. All three client platforms have functional replay UI

---

#### Phase 09: Definition of Done

- [ ] helix-record submodule operational with CI/release
- [ ] Dual-path encoding functional (1 NVENC session, stream + record)
- [ ] MKV and fMP4 container output validated
- [ ] Instant replay buffer with user trigger
- [ ] Background sync to all 5 backends functional
- [ ] DASH replay client on all 3 platforms

---

### Phase 10: Monetization & Auth (P2)

**Phase Goal:** Implement complete authentication, authorization, multi-tenant isolation, and billing infrastructure for the HelixPlay platform. Support OAuth2/OIDC, device authorization for TVs, RBAC, and resource quotas.

**Phase Duration:** 5 weeks
**Engineering Team:** 3 FTE (1 auth/security, 1 backend/platform, 1 billing)
**Deps:** P05 (real-time APIs), P09 (recording)

---

#### Task P10-T01: OAuth2/OIDC Integration (Auth0)

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T01 |
| **Title** | OAuth2/OIDC Authentication with Auth0 |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Security Engineer |

**Description:**
Implement OAuth2/OIDC authentication using Auth0 as identity provider. Support Authorization Code flow with PKCE for web/desktop and Device Authorization Grant (RFC 8628) for TV/console clients. Implement JWT token validation, refresh token rotation, and session management.

**Files:**
```
+ helix-auth/go.mod                            # Module: github.com/helixplay/helix-auth
+ helix-auth/pkg/oauth2/oidc.go                # OIDC discovery & validation
+ helix-auth/pkg/oauth2/pkce.go                # PKCE implementation
+ helix-auth/pkg/oauth2/device.go              # RFC 8628 Device Flow
+ helix-auth/pkg/oauth2/token.go               # JWT token management
+ helix-auth/pkg/oauth2/refresh.go             # Refresh token rotation
+ helix-auth/pkg/oauth2/session.go             # Session store (Redis)
+ helix-auth/pkg/middleware/auth.go            # HTTP auth middleware
+ helix-auth/pkg/middleware/rbac.go            # RBAC middleware
+ helix-auth/cmd/token-test/main.go            # Token validation tool
+ helix-auth/configs/auth0-tenant.yaml         # Auth0 tenant configuration
```

**Auth Flows:**

| Flow | Use Case | Implementation |
|------|----------|----------------|
| Authorization Code + PKCE | Web, Desktop, Mobile | Standard OIDC flow |
| Device Authorization Grant (RFC 8628) | TV, Console | QR code + polling |
| Client Credentials | Service-to-service | mTLS + JWT |

**Acceptance Criteria:**
1. Authorization Code + PKCE flow completes in < 3 seconds (cold start)
2. Device Authorization Grant shows QR code, user authenticates on phone, TV polls successfully
3. JWT access tokens: RS256 signed, 15-minute expiry, validated without Auth0 call (JWKS cache)
4. Refresh token rotation: single-use, family detection for theft
5. Session revocation: token blacklist in Redis with < 50ms check latency

---

#### Task P10-T02: Device Authorization Grant (RFC 8628) for TV

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T02 |
| **Title** | TV/Console Device Authorization Flow with QR Code |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Client Engineer |

**Description:**
Implement RFC 8628 Device Authorization Grant for TV and console clients. Generate QR code for user to scan with mobile device. Poll for authorization completion. Implement proper timeout and error handling.

**Files:**
```
+ helix-auth/pkg/device/flow.go                # Device flow orchestrator
+ helix-auth/pkg/device/qrcode.go              # QR code generation
+ helix-auth/pkg/device/poll.go                # Polling with backoff
~ helix-flutter/lib/auth/device_flow.dart      # Flutter TV device auth
~ helix-angular/src/app/auth/device/           # Web device auth (fallback)
```

**Flow:**
1. TV client requests device code from `/oauth/device/code`
2. TV displays QR code containing `verification_uri_complete`
3. User scans QR with phone, authenticates via standard OIDC flow
4. TV polls `/oauth/token` at interval until authorized or expired
5. On success, TV receives access + refresh tokens

**Acceptance Criteria:**
1. QR code generation < 100ms on TV hardware (Android TV, Apple TV)
2. Complete flow (scan → authenticate → authorized) < 15 seconds
3. Proper timeout after 15 minutes with user-friendly message
4. Works without keyboard input on TV (pure mobile-driven auth)
5. Error handling for denied access, expired code, network failure

---

#### Task P10-T03: Tenant Isolation Architecture

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T03 |
| **Title** | Multi-Tenant Data Isolation (helix-tenant submodule) |
| **Priority** | P1 |
| **Effort** | 2pw |
| **Owner** | Platform Engineer |

**Description:**
Implement strict tenant isolation where each white-label tenant has fully separated users, game catalog, recordings, billing data, and theming. Use row-level security (RLS) in CockroachDB and tenant context propagation.

**Files:**
```
+ helix-tenant/go.mod                          # Module: github.com/helixplay/helix-tenant
+ helix-tenant/pkg/context/tenant.go           # Tenant context propagation
+ helix-tenant/pkg/isolation/db.go             # DB-level tenant isolation
+ helix-tenant/pkg/isolation/rls.go            # CockroachDB RLS policies
+ helix-tenant/pkg/isolation/middleware.go     # Tenant extraction middleware
+ helix-tenant/pkg/isolation/validate.go       # Cross-tenant access validation
+ helix-tenant/pkg/catalog/overlay.go          # Per-tenant catalog overlay
+ helix-tenant/pkg/billing/usage.go            # Per-tenant usage tracking
+ helix-tenant/pkg/models/tenant.go            # Tenant data models
+ helix-tenant/pkg/api/admin.go                # Tenant admin API
+ helix-tenant/configs/default-policies.sql    # Default RLS policies
+ helix-tenant/LICENSE                         # MIT
```

**Isolation Model:**
- **Database:** CockroachDB row-level security (`CREATE POLICY tenant_isolation`)
- **Storage:** Per-tenant S3 bucket prefix (`/recordings/{tenant_id}/...`)
- **Cache:** Redis key prefix (`tenant:{id}:...`)
- **Message Bus:** NATS tenant-scoped subjects (`events.{tenant_id}.session.start`)
- **Compute:** Per-tenant resource quotas via cgroup v2

**Acceptance Criteria:**
1. No cross-tenant data access possible (verified via SQL injection + direct DB access tests)
2. Tenant context automatically propagated via gRPC metadata
3. RLS policies enforce tenant isolation at database level (bypass impossible without superuser)
4. Per-tenant catalog overlay correctly merges global + tenant-specific entries
5. Tenant provisioning: new tenant fully isolated in < 30 seconds

---

#### Task P10-T04: RBAC Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T04 |
| **Title** | Role-Based Access Control with Hierarchical Roles |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Description:**
Implement RBAC with predefined roles and custom role support. Roles are tenant-scoped. Permission checks on every API endpoint.

**Files:**
```
+ helix-auth/pkg/rbac/model.go                 # RBAC data model
+ helix-auth/pkg/rbac/roles.go                 # Predefined roles
+ helix-auth/pkg/rbac/permission.go            # Permission definitions
+ helix-auth/pkg/rbac/enforcer.go              # Casbin-compatible enforcer
+ helix-auth/pkg/rbac/middleware.go            # RBAC HTTP/gRPC middleware
+ helix-auth/pkg/rbac/admin.go                 # Role management API
```

**Default Roles:**

| Role | Permissions | Scope |
|------|------------|-------|
| Platform Admin | Full access | Platform-wide |
| Tenant Admin | Full access within tenant | Tenant |
| Tenant Manager | User management, billing, catalog | Tenant |
| End User | Play games, manage recordings | Own data only |
| Guest | View catalog, no play | N/A |

**Acceptance Criteria:**
1. All API endpoints enforce RBAC checks (verified via automated penetration test)
2. Role assignment atomic and consistent across distributed nodes
3. Permission changes propagate in < 5 seconds
4. Custom role creation with granular permissions (UI + API)
5. Audit log entry for every permission check failure

---

#### Task P10-T05: Billing Engine & Resource Quotas

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T05 |
| **Title** | Usage-Based Billing with Resource Quotas |
| **Priority** | P2 |
| **Effort** | 1.5pw |
| **Owner** | Platform Engineer |

**Description:**
Implement usage-based billing tracking streaming minutes, storage, and concurrent sessions. Enforce resource quotas at tenant and user level. Integrate with Stripe for payment processing.

**Files:**
```
+ helix-tenant/pkg/billing/engine.go           # Billing event processor
+ helix-tenant/pkg/billing/stripe.go           # Stripe integration
+ helix-tenant/pkg/billing/quota.go            # Quota management
+ helix-tenant/pkg/billing/usage.go            # Usage aggregation
+ helix-tenant/pkg/billing/invoice.go          # Invoice generation
+ helix-tenant/pkg/billing/webhook.go          # Stripe webhook handler
+ helix-tenant/pkg/quota/enforcer.go           # Quota enforcement
+ helix-tenant/pkg/quota/limit.go              # Limit definitions
```

**Metered Resources:**

| Resource | Unit | Quota Levels |
|----------|------|-------------|
| Streaming time | Minutes | Per-user daily, per-tenant monthly |
| Concurrent sessions | Count | Per-tenant hard limit |
| Recording storage | GB | Per-user, per-tenant |
| Bandwidth | GB | Per-tenant monthly |
| API requests | Count | Per-tenant rate limit |

**Acceptance Criteria:**
1. Billing events emitted for every minute of streaming (NATS → aggregation)
2. Quota enforcement prevents usage beyond limit with user-friendly message
3. Stripe integration: subscription creation, metered billing, invoice generation
4. Usage dashboard shows real-time consumption ( < 5 second delay)
5. Grace period: 10% over-quota allowed before hard cut-off

---

#### Task P10-T06: helix-vault Submodule (Secrets Management)

| Attribute | Detail |
|-----------|--------|
| **ID** | P10-T06 |
| **Title** | Secrets Management with KEK Hierarchy |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Files:**
```
+ helix-vault/go.mod                           # Module: github.com/helixplay/helix-vault
+ helix-vault/pkg/kek/hierarchy.go             # Key Encryption Key hierarchy
+ helix-vault/pkg/kek/rotation.go              # Automatic KEK rotation
+ helix-vault/pkg/secrets/store.go             # Secret storage interface
+ helix-vault/pkg/secrets/hashicorp.go         # HashiCorp Vault backend
+ helix-vault/pkg/secrets/aws.go               # AWS Secrets Manager backend
+ helix-vault/pkg/secrets/file.go              # Development file backend
+ helix-vault/pkg/api/secrets.go               # Secret management API
+ helix-vault/LICENSE                          # MIT
```

**KEK Hierarchy:**
- **Platform KEK (L0):** HSM-protected, never leaves HSM
- **Tenant KEK (L1):** Encrypted by L0, per-tenant
- **Session KEK (L2):** Ephemeral, encrypted by L1, auto-rotated

**Acceptance Criteria:**
1. Secret retrieval latency < 10ms (cached) / < 100ms (cold)
2. KEK rotation: automatic every 90 days, zero-downtime
3. Integration with HashiCorp Vault and AWS Secrets Manager
4. No plaintext secrets in config files or environment variables
5. Audit log for every secret access (who, when, what)

---

#### Phase 10: Definition of Done

- [ ] OAuth2/OIDC authentication functional (all flows)
- [ ] Device Authorization Grant works on TV with QR code
- [ ] Tenant isolation enforced at database, storage, cache, and message bus levels
- [ ] RBAC checks on every API endpoint
- [ ] Billing engine tracking usage with Stripe integration
- [ ] Secret management with KEK hierarchy and auto-rotation

---

### Phase 11: Hardening & Security (P2)

**Phase Goal:** Achieve R-18 Operational Integrity rating through comprehensive security hardening, mTLS everywhere, audit logging, key rotation, continuous security scanning, and container isolation verification.

**Phase Duration:** 5 weeks
**Engineering Team:** 3 FTE (2 security engineers, 1 DevSecOps)
**Deps:** P10 (auth/tenant infrastructure)

---

#### Task P11-T01: R-18 Operational Integrity Enforcement

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T01 |
| **Title** | R-18 Operational Integrity Framework Implementation |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Security Lead |

**Description:**
Implement the R-18 Operational Integrity framework as the foundation of security posture. This includes 18 security controls across 6 domains: Identity, Infrastructure, Data, Application, Operations, and Compliance.

**Files:**
```
+ helix-r18-safeexec/go.mod                    # Module: github.com/helixplay/helix-r18-safeexec
+ helix-r18-safeexec/pkg/constitution/controls.go     # 18 control definitions
+ helix-r18-safeexec/pkg/constitution/checks.go       # Automated compliance checks
+ helix-r18-safeexec/pkg/constitution/report.go       # Compliance reporting
+ helix-r18-safeexec/pkg/safeexec/sandbox.go          # Process sandboxing
+ helix-r18-safeexec/pkg/safeexec/seccomp.go          # seccomp-bpf profiles
+ helix-r18-safeexec/pkg/safeexec/capabilities.go     # Linux capabilities management
+ helix-r18-safeexec/pkg/safeexec/namespaces.go       # Namespace isolation
+ helix-r18-safeexec/configs/seccomp-gaming.json      # Gaming process seccomp profile
+ helix-r18-safeexec/configs/r18-baseline.yaml        # R-18 baseline configuration
+ helix-r18-safeexec/cmd/r18-audit/main.go            # R-18 compliance auditor
+ helix-r18-safeexec/LICENSE                          # MIT (root dependency tree)
```

**R-18 Controls (6 Domains × 3 Controls):**

| Domain | Controls |
|--------|----------|
| **Identity** | (1) Strong MFA enforcement, (2) Just-in-time access, (3) Service identity (SPIFFE) |
| **Infrastructure** | (4) Hardened base images, (5) Network micro-segmentation, (6) Immutable infrastructure |
| **Data** | (7) Encryption at rest (AES-256-GCM), (8) Encryption in transit (TLS 1.3), (9) Key lifecycle management |
| **Application** | (10) SAST/DAST in CI, (11) Dependency vulnerability scanning, (12) Secure code review |
| **Operations** | (13) Audit logging (append-only), (14) Anomaly detection, (15) Incident response automation |
| **Compliance** | (16) Automated compliance checks, (17) Evidence collection, (18) Penetration testing |

**Acceptance Criteria:**
1. All 18 controls implemented and continuously monitored
2. Automated R-18 compliance check passes in CI
3. Quarterly penetration test findings remediated within SLA
4. Security scorecard published monthly to stakeholders
5. `helix-r18-safeexec` is root of dependency tree (all modules depend on it)

---

#### Task P11-T02: mTLS Topology Completion

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T02 |
| **Title** | mTLS Everywhere - Service Mesh Integration |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | Security Engineer |

**Description:**
Complete mTLS deployment across all service-to-service communication. Use SPIFFE/SPIRE for workload identity. Implement automatic certificate rotation and revocation.

**Files:**
```
+ helix-vault/pkg/mtls/config.go               # mTLS configuration
+ helix-vault/pkg/mtls/spiffe.go               # SPIFFE ID management
+ helix-vault/pkg/mtls/rotation.go             # Certificate auto-rotation
+ helix-vault/pkg/mtls/revocation.go           # Certificate revocation list
~ infra/k8s/servicemesh/istio-mtls.yaml        # Istio mTLS peer authentication
~ infra/terraform/spire-server.tf              # SPIRE server deployment
```

**mTLS Coverage Matrix:**

| Connection | mTLS | Identity | Rotation |
|------------|------|----------|----------|
| Client → Edge | TLS 1.3 + Client Cert | JWT | 15min (session) |
| Edge → API Gateway | mTLS | SPIFFE | 24hr |
| API → Session Service | mTLS | SPIFFE | 24hr |
| Session → Host Agent | mTLS + WireGuard | SPIFFE + Session Key | Per-session |
| Host → Storage | mTLS | SPIFFE | 24hr |
| Service → CockroachDB | mTLS | Client Cert | 30 days |
| Service → Redis | TLS + AUTH | Password | 90 days |
| Service → NATS | mTLS | SPIFFE | 24hr |

**Acceptance Criteria:**
1. 100% of inter-service traffic encrypted with mTLS (verified via network capture)
2. No plaintext service communication on any port
3. Certificate rotation: zero-downtime, automatic
4. SPIFFE ID validation on every inbound connection
5. Revoked certificates rejected within 5 minutes of revocation

---

#### Task P11-T03: Audit Logging System

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T03 |
| **Title** | Append-Only Audit Log with Tamper Evidence |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Description:**
Implement centralized audit logging for all security-relevant events. Use append-only storage with cryptographic chain-of-custody (hash chain). Forward to SIEM.

**Files:**
```
+ helix-audit/go.mod                           # Module: github.com/helixplay/helix-audit
+ helix-audit/pkg/log/appender.go              # Append-only log writer
+ helix-audit/pkg/log/chain.go                 # Hash chain tamper evidence
+ helix-audit/pkg/log/events.go                # Event type definitions
+ helix-audit/pkg/log/siem.go                  # SIEM forwarder (Splunk/Sentinel)
+ helix-audit/pkg/api/query.go                 # Audit log query API (read-only)
+ helix-audit/configs/audit-events.yaml        # Event classification
```

**Logged Events:**
- Authentication (success/failure)
- Authorization denials
- Session start/stop
- Admin actions (role changes, config changes)
- Key access and rotation
- Data access (recordings, personal data)
- API rate limit violations

**Acceptance Criteria:**
1. Every security event logged within 50ms of occurrence
2. Log entries cryptographically chained (SHA-256 of previous entry)
3. Tamper detection: any modification breaks chain verification
4. SIEM forwarder delivers events with < 5 second latency
5. Log retention: 7 years (compliance), queryable for 90 days hot storage

---

#### Task P11-T04: Security Scanning Pipeline

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T04 |
| **Title** | Continuous Security Scanning (6-Scanner Pipeline) |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | DevSecOps Engineer |

**Description:**
Integrate 6 security scanners into CI/CD pipeline. Block deployment on critical findings. Track security debt.

**Files:**
```
+ .github/workflows/security-scan.yml          # Unified security scan workflow
+ .github/scripts/sonarqube-scan.sh            # SonarQube integration
+ .github/scripts/snyk-scan.sh                 # Snyk dependency scan
+ .github/scripts/semgrep-scan.sh              # Semgrep SAST rules
+ .github/scripts/trivy-scan.sh                # Trivy container scan
+ .github/scripts/gitleaks-scan.sh             # Secret detection
+ .github/scripts/govulncheck.sh               # Go vulnerability check
+ security/sonar-project.properties            # SonarQube configuration
+ security/.semgrep/helix-rules.yaml           # Custom Semgrep rules
+ security/.trivyignore                        # Trivy exception list
```

**Scanner Matrix:**

| Scanner | Type | Trigger | Blocking |
|---------|------|---------|----------|
| SonarQube | SAST + Code Quality | Every PR | High severity |
| Snyk | Dependency vulns | Every PR + daily | Critical CVSS |
| Semgrep | SAST (custom rules) | Every PR | Critical |
| Trivy | Container + OS vulns | Every build | Critical |
| gitleaks | Secret detection | Every PR + pre-commit | Always (secrets) |
| govulncheck | Go-specific vulns | Every PR | Critical |

**Acceptance Criteria:**
1. All 6 scanners run on every PR with results in < 10 minutes
2. Zero critical findings in default branch (enforced by branch protection)
3. Security scan results published to PR as comment
4. Vulnerability SLA: Critical < 24hr, High < 7 days, Medium < 30 days
5. Historical security debt tracked in dashboard

---

#### Task P11-T05: Container Isolation Verification

| Attribute | Detail |
|-----------|--------|
| **ID** | P11-T05 |
| **Title** | Container & Sandbox Isolation Verification |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Security Engineer |

**Description:**
Verify container isolation for host agent processes. Implement seccomp-bpf, AppArmor/SELinux profiles, and user namespace isolation. Validate escape resistance.

**Files:**
```
+ helix-r18-safeexec/configs/seccomp-gaming.json    # Game process seccomp
+ helix-r18-safeexec/configs/apparmor-gaming.profile # AppArmor profile
+ helix-r18-safeexec/configs/selinux-gaming.te       # SELinux type enforcement
+ helix-r18-safeexec/pkg/isolation/container.go      # Container runtime wrapper
+ helix-r18-safeexec/pkg/isolation/verify.go         # Isolation verification tests
+ helix-r18-safeexec/cmd/isolation-test/main.go      # Isolation test runner
```

**Acceptance Criteria:**
1. Game process cannot escape container (verified via `docker escape` test suite)
2. seccomp-bpf blocks all syscalls except explicit allowlist (verified via `strace -f`)
3. No privilege escalation possible from unprivileged container
4. Container resource limits (CPU, memory, I/O) enforced by cgroup v2
5. Game process has no network access except through proxy/tunnel

---

#### Phase 11: Definition of Done

- [ ] R-18 Operational Integrity: all 18 controls implemented
- [ ] mTLS on 100% of service-to-service traffic
- [ ] Audit logging: append-only, tamper-evident, SIEM-integrated
- [ ] KEK rotation: automatic, zero-downtime
- [ ] Security scanning: 6 scanners in CI, zero critical findings
- [ ] Container isolation verified and tested

---

### Phase 12: Beta Launch (P2)

**Phase Goal:** Launch closed beta with canary deployment capability, 30-day replay retention, comprehensive load testing, integration testing, and complete documentation.

**Phase Duration:** 4 weeks
**Engineering Team:** 4 FTE (1 SRE, 1 QA lead, 1 technical writer, 1 product)
**Deps:** P07-P11 (all previous phases)

---

#### Task P12-T01: Canary Deployment System

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T01 |
| **Title** | Operator-Facing Canary Deployment with Automatic Rollback |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | SRE |

**Description:**
Implement canary deployment for host agent fleet. Operator controls traffic split percentage, health criteria, and automatic rollback thresholds.

**Files:**
```
+ helix-deploy/go.mod                          # Module: github.com/helixplay/helix-deploy
+ helix-deploy/pkg/canary/controller.go        # Canary deployment controller
+ helix-deploy/pkg/canary/metrics.go           # Health metric evaluation
+ helix-deploy/pkg/canary/rollback.go          # Automatic rollback logic
+ helix-deploy/pkg/canary/promote.go           # Full promotion workflow
+ helix-deploy/cmd/helix-deploy/main.go        # CLI operator tool
+ helix-deploy/configs/canary-defaults.yaml    # Default canary parameters
```

**Canary Parameters:**
- Initial traffic split: 5% → 25% → 50% → 100%
- Health criteria: p99 latency < threshold, error rate < 0.1%, GPU utilization stable
- Evaluation window: 10 minutes per step
- Automatic rollback on any health criterion failure

**Acceptance Criteria:**
1. Canary deployment completes in < 15 minutes (all steps)
2. Automatic rollback triggers in < 30 seconds on health check failure
3. Operator can view real-time canary metrics dashboard
4. Canary works for host agent binary, not just container
5. Zero-downtime deployment (rolling update with session drain)

---

#### Task P12-T02: 30-Day Replay Retention

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T02 |
| **Title** | Replay Storage Lifecycle (30-Day Retention) |
| **Priority** | P2 |
| **Effort** | 0.5pw |
| **Owner** | SRE |

**Description:**
Implement automatic lifecycle management for replay recordings: 7 days hot on NVMe, 23 days warm on object storage, automatic deletion after 30 days. Configurable per-tenant.

**Files:**
```
+ helix-record/pkg/lifecycle/policy.go         # Retention policy engine
+ helix-record/pkg/lifecycle/transition.go     # Storage tier transitions
+ helix-record/pkg/lifecycle/cleanup.go        # Automated cleanup
+ infra/terraform/lifecycle-rules.tf           # S3 lifecycle rules
```

**Acceptance Criteria:**
1. Recordings available for instant replay for 7 days (NVMe)
2. Recordings available for download/streaming for 30 days total
3. Automatic deletion after 30 days (no manual intervention)
4. Per-tenant retention override functional
5. Deletion audit log entries generated

---

#### Task P12-T03: Load Testing at Scale

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T03 |
| **Title** | Platform Load Testing (1000 Concurrent Sessions) |
| **Priority** | P1 |
| **Effort** | 1.5pw |
| **Owner** | QA Lead |

**Description:**
Execute comprehensive load test simulating 1000 concurrent sessions. Measure latency distribution, GPU utilization, network throughput, and resource contention.

**Files:**
```
+ tests/load/k6/streaming-load.js              # k6 streaming load script
+ tests/load/k6/session-lifecycle.js           # Session CRUD load test
+ tests/load/terraform/load-infra.tf            # Load test infrastructure
+ tests/load/Makefile                           # Load test orchestration
+ tests/load/reports/template.html              # Load test report template
```

**Load Test Scenarios:**

| Scenario | Sessions | Duration | Metrics |
|----------|----------|----------|---------|
| Steady State | 1000 | 4 hours | p50/p99/p999 latency, GPU%, throughput |
| Ramp Up | 0 → 1000 | 30 min | Scale-up latency, connection success rate |
| Ramp Down | 1000 → 0 | 30 min | Graceful session termination |
| Spike | 0 → 2000 | 5 min | Auto-scaling response, queue depth |
| Burst | 100 → 500 | 1 min | Latency impact under sudden load |
| Long-running | 100 | 24 hours | Memory leaks, thermal throttling, drift |

**Acceptance Criteria:**
1. 1000 concurrent sessions sustained for 4 hours with p999 < 50ms
2. Connection success rate > 99.9% during ramp-up
3. No memory leaks (RSS stable within 5% over 24-hour test)
4. Auto-scaling responds to spike in < 60 seconds
5. GPU thermal throttling < 2% of session time

---

#### Task P12-T04: Final Integration Testing

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T04 |
| **Title** | End-to-End Integration Test Suite |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | QA Lead |

**Description:**
Complete end-to-end integration testing across all client platforms, all codecs, all controller types, and all network conditions.

**Files:**
```
+ tests/e2e/README.md                          # E2E test documentation
+ tests/e2e/desktop/e2e_test.go                # Wails desktop E2E
+ tests/e2e/mobile/e2e_test.go                 # Flutter mobile E2E
+ tests/e2e/web/e2e_test.go                    # Angular web E2E
+ tests/e2e/shared/assertions.go               # Shared test assertions
+ tests/e2e/shared/fixtures.go                 # Test fixtures
```

**E2E Matrix:**

| Client | Codecs | Controllers | Networks |
|--------|--------|-------------|----------|
| Wails (Win/Mac/Linux) | H.264/HEVC/AV1 | DS5/Xbox/Switch | LAN/WAN/4G/5G |
| Flutter (Android/iOS) | H.264/HEVC | DS5/Xbox/MFi | WiFi/5G/4G |
| Flutter TV (Android TV/tvOS) | H.264/HEVC | DS5/Xbox | Ethernet/WiFi |
| Angular+WASM (Chrome/Firefox/Safari) | H.264/HEVC (WebCodecs) | DS5/Xbox | LAN/WiFi |

**Acceptance Criteria:**
1. All E2E tests pass on all platform combinations
2. Video quality MOS > 4.0 (subjective testing panel, n=20)
3. Controller latency < 2ms (measured via oscilloscope)
4. Audio A/V sync < 5ms (measured with test signal)
5. No P1 or P2 bugs remaining in backlog

---

#### Task P12-T05: Documentation Completion

| Attribute | Detail |
|-----------|--------|
| **ID** | P12-T05 |
| **Title** | Complete Technical Documentation Suite |
| **Priority** | P2 |
| **Effort** | 1pw |
| **Owner** | Technical Writer |

**Files:**
```
+ docs/architecture/README.md                  # Architecture overview
+ docs/architecture/streaming.md               # Streaming pipeline
+ docs/architecture/controllers.md             # Controller input pipeline
+ docs/architecture/security.md                # Security architecture
+ docs/operations/deployment.md                # Deployment guide
+ docs/operations/monitoring.md                # Monitoring & alerting
+ docs/operations/incident-response.md         # Incident response
+ docs/api/README.md                           # API reference
+ docs/api/openapi.yaml                        # OpenAPI specification
+ docs/clients/README.md                       # Client integration guide
+ docs/contributing/README.md                  # Contribution guidelines
+ docs/changelog/CHANGELOG.md                  # Version changelog
```

**Acceptance Criteria:**
1. All architecture documents complete with diagrams
2. API documentation generated from OpenAPI spec
3. Deployment guide enables new team member to deploy in < 2 hours
4. All 29 submodules have README with build/test instructions
5. Documentation published to docs site (MkDocs)

---

#### Phase 12: Definition of Done

- [ ] Canary deployment system operational
- [ ] 30-day replay retention automated
- [ ] 1000 concurrent session load test passed
- [ ] E2E tests pass on all platform combinations
- [ ] Complete documentation suite published

---

### Phase 13: GA Release (P1)

**Phase Goal:** Release HelixPlay v1.0.0 with complete tag-publish across all 29 submodules, four-mirror sync verification, final security audit, and production deployment guide.

**Phase Duration:** 3 weeks
**Engineering Team:** 2 FTE (1 release engineer, 1 security auditor)
**Deps:** P12 (beta launch complete)

---

#### Task P13-T01: v1.0.0 Release Train

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T01 |
| **Title** | Coordinated v1.0.0 Release Across All Submodules |
| **Priority** | P1 |
| **Effort** | 1pw |
| **Owner** | Release Engineer |

**Description:**
Execute coordinated release of v1.0.0 across all 29 submodules with correct semantic versioning, dependency alignment, and release notes.

**Files:**
```
+ scripts/release/train.sh                     # Release train orchestrator
+ scripts/release/validate.sh                  # Pre-release validation
+ scripts/release/tag.sh                       # Cross-repo tagging
+ scripts/release/notes.sh                     # Release notes generator
+ .github/workflows/release-train.yml          # Release train CI
```

**Release Train Process:**
1. **Validation:** All tests pass, zero critical security findings
2. **Version Alignment:** Update all `go.mod` to v1.0.0, resolve dependencies
3. **Tagging:** Create signed Git tags (`git tag -s v1.0.0`) on all repos
4. **Build:** Build all artifacts (binaries, containers, WASM modules)
5. **Publish:** Push to artifact registries (GitHub Packages, Docker Hub, npm)
6. **Verify:** Confirm all artifacts downloadable and checksums match

**Acceptance Criteria:**
1. All 29 submodules tagged v1.0.0 within 1 hour window
2. All cross-module dependencies resolve to v1.0.0
3. Container images published with signed attestations (cosign)
4. Release notes generated from conventional commits (automated)
5. No post-release hotfixes required in first 48 hours

---

#### Task P13-T02: Four-Mirror Git Sync Verification

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T02 |
| **Title** | Four-Mirror Git Topology Sync Verification |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | Release Engineer |

**Description:**
Verify synchronized state across four Git mirrors (GitHub primary, GitLab, Gitea, Bitbucket). All repos, tags, releases, and issues synced.

**Mirror Topology:**
- **Primary:** GitHub (github.com/helixplay/*)
- **Mirror 1:** GitLab (gitlab.helixplay.io/helixplay/*)
- **Mirror 2:** Gitea (gitea.helixplay.io/helixplay/*)
- **Mirror 3:** Bitbucket (bitbucket.org/helixplay/*)

**Files:**
```
+ scripts/mirror/sync.sh                       # Mirror sync orchestrator
+ scripts/mirror/verify.sh                     # Mirror verification
+ .github/workflows/mirror-sync.yml            # Automated mirror sync
```

**Acceptance Criteria:**
1. All 29 repos present on all 4 mirrors with identical HEAD
2. All v1.0.0 tags present and matching SHA on all mirrors
3. Mirror sync latency < 5 minutes from primary push
4. Issue/PR metadata synced to GitLab (for local tracking)
5. Automated recovery if mirror falls out of sync

---

#### Task P13-T03: Final Security Audit

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T03 |
| **Title** | Independent Security Audit & Penetration Test |
| **Priority** | P1 |
| **Effort** | 1pw (external) |
| **Owner** | Security Lead |

**Description:**
Commission independent security audit covering: code review, penetration testing, architecture review, and compliance assessment.

**Audit Scope:**
- SAST: All 29 submodules via SonarQube + Semgrep
- DAST: Live beta environment via OWASP ZAP
- Penetration test: External firm, black + grey box
- Architecture review: Threat model validation
- Compliance: SOC 2 Type II readiness assessment

**Acceptance Criteria:**
1. Zero critical vulnerabilities in audit report
2. All high findings have remediation plan with timeline
3. Penetration test: no remote code execution, no privilege escalation
4. Threat model validated: all identified threats have mitigations
5. SOC 2 Type II readiness: no gaps in controls

---

#### Task P13-T04: Production Deployment Guide

| Attribute | Detail |
|-----------|--------|
| **ID** | P13-T04 |
| **Title** | Production Deployment Guide & Runbooks |
| **Priority** | P1 |
| **Effort** | 0.5pw |
| **Owner** | SRE |

**Files:**
```
+ docs/production/README.md                    # Production overview
+ docs/production/prerequisites.md             # Infrastructure prerequisites
+ docs/production/deployment.md                # Step-by-step deployment
+ docs/production/configuration.md             # Configuration reference
+ docs/production/monitoring.md                # Monitoring setup
+ docs/production/runbooks/                    # Incident runbooks
+ docs/production/runbooks/latency-spike.md    # Latency spike response
+ docs/production/runbooks/gpu-failure.md      # GPU node failure
+ docs/production/runbooks/network-partition.md # Network partition
+ docs/production/runbooks/security-incident.md # Security incident
```

**Acceptance Criteria:**
1. Deployment guide enables production deployment from scratch in < 4 hours
2. All runbooks have decision trees and command snippets
3. Monitoring dashboards created and documented
4. On-call rotation documented with escalation procedures
5. Disaster recovery plan: RPO < 5 minutes, RTO < 30 minutes

---

#### Phase 13: Definition of Done

- [ ] v1.0.0 tagged and published across all 29 submodules
- [ ] Four-mirror sync verified
- [ ] Independent security audit passed (zero critical findings)
- [ ] Production deployment guide complete
- [ ] Platform live in production

---

## Part B: Architecture Deep-Dive Implementation Tasks

### B1: Streaming Pipeline Implementation

---

#### B1-T01: WebRTC (Pion v4) Integration

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T01 |
| **Title** | WebRTC Pion v4 Full Integration |
| **Effort** | 3pw |

**Description:**
Implement WebRTC transport using Pion v4 (`github.com/pion/webrtc/v4`) for peer connection management, ICE, DTLS, SRTP, and SCTP data channels. Customize for low-latency game streaming.

**Files:**
```
+ helix-stream/pkg/webrtc/pc.go                # Peer connection factory
+ helix-stream/pkg/webrtc/config.go            # WebRTC configuration
+ helix-stream/pkg/webrtc/ice.go               # ICE server & candidate management
+ helix-stream/pkg/webrtc/negotiation.go       # SDP offer/answer
+ helix-stream/pkg/webrtc/track.go             # Media track management
+ helix-stream/pkg/webrtc/datachannel.go       # Input/control data channels
+ helix-stream/pkg/webrtc/stats.go             # RTC stats collection
+ helix-stream/pkg/webrtc/bwe.go               # Bandwidth estimation
```

**Pion v4 Configuration:**
```go
var DefaultConfig = webrtc.Configuration{
    ICETransportPolicy: webrtc.ICETransportPolicyAll,
    BundlePolicy:       webrtc.BundlePolicyMaxBundle,
    RTCPFeedback: []webrtc.RTCPFeedback{
        {Type: "goog-remb"},
        {Type: "transport-cc"},
        {Type: webrtc.TypeRTCPFBNACK},
        {Type: webrtc.TypeRTCPFBNACK},
        {Parameter: "pli", Type: webrtc.TypeRTCPFBGoogREMB},
    },
    SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
}

var SettingEngine = webrtc.SettingEngine{
    // Disable ICE lite (full ICE for host connectivity)
    // Enable DTLS 1.3
    // Set MTU discovery
    // Enable TWCC (Transport Wide Congestion Control)
}
```

**Key Implementation Details:**
1. **ICE:** Custom STUN/TURN server deployment with regional affinity. TURN for symmetric NAT fallback.
2. **DTLS:** 1.3 with cipher suites `TLS_AES_128_GCM_SHA256`, `TLS_AES_256_GCM_SHA384`
3. **SRTP:** AES-GCM preferred over AES-CM (reduced CPU, better security)
4. **SCTP:** Data channel for controller input with unordered, unreliable delivery
5. **Bandwidth Estimation:** TWCC + custom HelixPlay SQP (Streaming Quality Predictor)

**Acceptance Criteria:**
1. Peer connection establishment < 500ms (LAN), < 2s (WAN with TURN)
2. ICE candidate gathering < 200ms with regional STUN
3. Data channel latency for controller input < 1ms (localhost)
4. SRTP throughput > 100Mbps sustained (single peer connection)
5. Graceful handling of all NAT types (full cone → symmetric)

---

#### B1-T02: Codec Ladder Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T02 |
| **Title** | Codec Negotiation Ladder (H.264 → HEVC → AV1) |
| **Effort** | 2pw |

**Description:**
Implement codec negotiation ladder with automatic fallback. Client advertises supported codecs via SDP; server selects optimal codec based on client capability, network conditions, and GPU encoder availability.

**Codec Ladder:**

| Priority | Codec | Profile | Use Case | Fallback Trigger |
|----------|-------|---------|----------|-----------------|
| 1 | AV1 | Main 10 | Best quality, lowest bitrate | Client doesn't support; GPU can't encode |
| 2 | HEVC | Main 10 | Good quality, hardware encode everywhere | Client doesn't support; patent licensing |
| 3 | H.264 | High | Universal compatibility | N/A (baseline) |

**Files:**
```
+ helix-stream/pkg/codec/ladder.go             # Codec ladder negotiation
+ helix-stream/pkg/codec/capability.go         # Client codec capability
+ helix-stream/pkg/codec/fallback.go           # Fallback logic
+ helix-stream/pkg/codec/policy.go             # Codec selection policy
```

**Negotiation Flow:**
1. Client SDP includes `a=rtpmap` for supported codecs (ordered by preference)
2. Server checks GPU encoder availability for each codec
3. Server selects highest-priority mutually supported codec
4. If selected codec fails at runtime (encoder error), fallback to next
5. Codec switch mid-stream: seamless (IDR frame, no reconnection)

**Acceptance Criteria:**
1. AV1 selected when client supports and GPU has AV1 encoder (RTX 40+, Intel Arc)
2. HEVC selected on Apple devices (hardware decode support)
3. H.264 fallback works on all platforms
4. Codec switch mid-stream completes in < 200ms
5. Codec selection logged for analytics

---

#### B1-T03: ABR/FEC/SQP Policies

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T03 |
| **Title** | Adaptive Bitrate, Forward Error Correction, Quality Policies |
| **Effort** | 2.5pw |

**Description:**
Implement Adaptive Bitrate (ABR), Forward Error Correction (FEC), and Streaming Quality Predictor (SQP) for resilient streaming under variable network conditions.

**Files:**
```
+ helix-stream/pkg/abr/controller.go           # ABR controller
+ helix-stream/pkg/abr/ladder.go               # Bitrate ladder definitions
+ helix-stream/pkg/fec/encoder.go              # FEC encoder (Reed-Solomon)
+ helix-stream/pkg/fec/decoder.go              # FEC decoder
+ helix-stream/pkg/sqp/predictor.go            # Streaming Quality Predictor
+ helix-stream/pkg/sqp/score.go                # Quality score calculation
+ helix-stream/pkg/congestion/detector.go      # Congestion detection
+ helix-stream/pkg/congestion/controller.go    # Congestion response
```

**ABR Ladder:**

| Resolution | Target Bitrate | Max Bitrate | Min Bitrate |
|------------|---------------|-------------|-------------|
| 4K (3840×2160) | 40 Mbps | 80 Mbps | 20 Mbps |
| 1440p (2560×1440) | 25 Mbps | 50 Mbps | 12 Mbps |
| 1080p (1920×1080) | 12 Mbps | 24 Mbps | 6 Mbps |
| 720p (1280×720) | 6 Mbps | 12 Mbps | 3 Mbps |
| 540p (960×540) | 3 Mbps | 6 Mbps | 1.5 Mbps |

**FEC Strategy:**
- **Proactive FEC:** 5-10% overhead during stable conditions
- **Reactive FEC:** Increase to 20-30% on packet loss detection
- **Unequal Error Protection:** I-frames get higher FEC protection than P-frames
- **Algorithm:** Reed-Solomon over GF(256), `fec(20, 16)` as baseline

**SQP (Streaming Quality Predictor):**
```go
type SQPScore struct {
    BandwidthEstimate  float64   // bps
    LossRate           float64   // 0-1
    Jitter             float64   // ms
    Rtt                float64   // ms
    GpuUtilization     float64   // 0-1
    ThermalThrottling  bool
    QualityScore       float64   // 0-100 composite
}
```

**Acceptance Criteria:**
1. ABR adapts to bandwidth changes within 2 seconds
2. FEC recovers from 5% packet loss without visible artifacts
3. SQP score accurately predicts quality degradation (correlation > 0.9)
4. Congestion detected and responded to within 500ms
5. Overall quality MOS > 4.0 at 2% packet loss

---

#### B1-T04: Transport Abstraction Layer

| Attribute | Detail |
|-----------|--------|
| **ID** | B1-T04 |
| **Title** | Transport Abstraction (WebRTC / QUIC / Custom UDP) |
| **Effort** | 2pw |

**Description:**
Implement transport abstraction layer that supports WebRTC (primary), QUIC (fallback for corporate firewalls), and custom UDP (LAN optimization). Automatic transport selection based on network conditions.

**Files:**
```
+ helix-stream/pkg/transport/interface.go      # Transport interface
+ helix-stream/pkg/transport/webrtc.go         # WebRTC transport
+ helix-stream/pkg/transport/quic.go           # QUIC transport (quic-go)
+ helix-stream/pkg/transport/udp.go            # Custom UDP transport
+ helix-stream/pkg/transport/selector.go       # Transport auto-selection
+ helix-stream/pkg/transport/fallback.go       # Transport fallback logic
```

**Transport Selection Matrix:**

| Condition | Primary | Fallback |
|-----------|---------|----------|
| Direct UDP possible | WebRTC (UDP) | QUIC |
| UDP blocked, TCP open | QUIC | WebRTC (TCP) |
| Corporate proxy | WebRTC (TCP/TURN) | N/A |
| LAN (sub-5ms latency) | Custom UDP | WebRTC |

**Acceptance Criteria:**
1. Transport selection completes within connection establishment time
2. Fallback to alternative transport on primary failure < 3 seconds
3. QUIC transport functional with equivalent latency to WebRTC
4. Custom UDP transport for LAN: latency reduced by > 20% vs. WebRTC
5. All transports support same feature set (FEC, ABR, encryption)

---

### B2: Controller Input Pipeline Implementation

---

#### B2-T01: 1kHz USB Polling Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B2-T01 |
| **Title** | 1kHz USB HID Polling for DualSense |
| **Effort** | 1.5pw |

**Description:**
Implement 1kHz (1ms interval) USB HID polling for DualSense controller on all platforms. Use raw HID access to achieve polling rates beyond standard OS driver defaults (125-250Hz).

**Files:**
```
+ helix-input/go.mod                           # Module: github.com/helixplay/helix-input
+ helix-input/pkg/hid/poll.go                  # HID polling loop
+ helix-input/pkg/hid/dualsense/usb.go         # DualSense USB protocol
+ helix-input/pkg/hid/dualsense/bluetooth.go   # DualSense Bluetooth protocol
+ helix-input/pkg/hid/dualsense/haptics.go     # Haptic feedback
+ helix-input/pkg/hid/dualsense/trigger.go     # Adaptive trigger control
+ helix-input/pkg/hid/dualsense/imu.go         # Gyroscope + accelerometer
+ helix-input/pkg/hid/dualsense/lightbar.go    # LED/lightbar control
+ helix-input/pkg/hid/xbox/core.go             # Xbox controller support
+ helix-input/pkg/hid/switch/core.go           # Nintendo Switch Pro support
```

**USB Polling Architecture:**
```go
type Poller struct {
    device      *hid.Device
    interval    time.Duration     // 1ms for 1kHz
    buffer      []byte            // 64-byte HID report
    callbacks   []InputCallback   // Registered callbacks
    ring        *SPSCRing[InputReport] // Lock-free ring to encode thread
}

func (p *Poller) Start() {
    // SCHED_FIFO thread at priority 93
    // Busy-wait with sched_yield for sub-microsecond precision
    // Or use hid_read_timeout with 1ms + io_uring for async
}
```

**Acceptance Criteria:**
1. USB polling interval: 1ms ± 50 microseconds (measured via USB analyzer)
2. No dropped input frames at 1kHz for 1-hour test
3. Works on Windows (WinUSB), macOS (IOHIDManager), Linux (hidraw)
4. Bluetooth fallback: 2ms interval (500Hz) acceptable
5. CPU overhead of polling thread < 2% of one core

---

#### B2-T02: Lock-Free SPSC Input Ring Buffer

| Attribute | Detail |
|-----------|--------|
| **ID** | B02-T02 |
| **Title** | Lock-Free Input Report Ring Buffer |
| **Effort** | 1pw |

**Description:**
Implement lock-free SPSC ring buffer for controller input reports from HID polling thread to network serialization thread. Zero-copy where possible.

**Files:**
```
+ helix-input/pkg/ring/input.go                # Input-specific SPSC ring
+ helix-input/pkg/serialize/packet.go          # Input packet serializer
+ helix-input/pkg/protocol/binary.go           # 16-32 byte binary protocol
```

**Binary Protocol:**
```
Offset  Size  Field
0       1     Packet type (0x01 = input report)
1       1     Sequence number (mod 256)
2       8     Timestamp (microseconds, monotonic)
10      2     Left stick X (0-65535)
12      2     Left stick Y (0-65535)
14      2     Right stick X (0-65535)
16      2     Right stick Y (0-65535)
18      2     Button bitmask (16 buttons)
20      1     Left trigger (0-255)
21      1     Right trigger (0-255)
22      1     D-pad state
23      1     Touchpad fingers
24      4     Gyro X (optional, extended packet)
28      4     Gyro Y (optional, extended packet)
```

**Acceptance Criteria:**
1. Ring buffer latency: poll → serialize < 50 microseconds
2. 1kHz input sustained without drops (ring capacity: 1024 entries)
3. Binary packet size: 16 bytes (standard), 32 bytes (extended with IMU)
4. Packet loss detection via sequence number gaps

---

#### B2-T03: DualSense Haptics & Adaptive Triggers

| Attribute | Detail |
|-----------|--------|
| **ID** | B2-T03 |
| **Title** | Full DualSense Feature Fidelity |
| **Effort** | 1.5pw |

**Description:**
Implement complete DualSense feature support: haptic feedback (L5/R5 actuators), adaptive triggers (L2/R2 resistance), gyroscope (6-axis), accelerometer, touchpad, and lightbar.

**Files:**
```
+ helix-input/pkg/hid/dualsense/haptics.go     # Haptic motor control
+ helix-input/pkg/hid/dualsense/trigger.go     # Adaptive trigger profiles
+ helix-input/pkg/hid/dualsense/imu.go         # Gyro + accelerometer
+ helix-input/pkg/hid/dualsense/touchpad.go    # Touchpad input
+ helix-input/pkg/hid/dualsense/lightbar.go    # LED control
+ helix-input/pkg/hid/dualsense/audio.go       # Controller speaker/headset
+ helix-input/pkg/hid/dualsense/mic.go         # Microphone
```

**Adaptive Trigger Profiles:**

| Profile | Description | Use Case |
|---------|-------------|----------|
| Off | No resistance | Default |
| Rigid | Full resistance | Heavy weapon |
| Vibration | Pulsing resistance | Machine gun |
| Slope | Increasing resistance | Bow draw |
| Feedback | Position-based feedback | Accelerator |

**Acceptance Criteria:**
1. Haptic feedback: L5/R5 independent control, < 5ms host→controller latency
2. Adaptive triggers: 255 resistance levels, profile switching < 10ms
3. Gyroscope: 2000 dps range, < 2ms report latency
4. Accelerometer: ±4g range, synchronized with gyro
5. All features functional over both USB and Bluetooth
6. Feature availability advertised to host via capability bits

---

#### B2-T04: Controller Hot-Plug & Renegotiation

| Attribute | Detail |
|-----------|--------|
| **ID** | B02-T04 |
| **Title** | Controller Hot-Plug & Mid-Session Renegotiation |
| **Effort** | 1pw |

**Description:**
Support controller connection/disconnection during active streaming session. Automatic capability renegotiation when controller changes (e.g., Xbox → DualSense swap).

**Files:**
```
+ helix-input/pkg/hotplug/monitor.go           # Hot-plug event monitor
+ helix-input/pkg/hotplug/renogotiate.go       # Mid-session renegotiation
+ helix-input/pkg/hotplug/manager.go           # Controller manager
```

**Acceptance Criteria:**
1. Controller connect: detected and functional within 1 second
2. Controller disconnect: session continues, input paused gracefully
3. Controller swap (type change): capability renegotiation in < 2 seconds
4. Multiple controllers: up to 4 simultaneous, player assignment
5. No session disruption during hot-plug events

---

### B3: Capture & Encode Pipeline Implementation

---

#### B3-T01: Per-OS Capture Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T01 |
| **Title** | Platform-Specific Screen Capture |
| **Effort** | 3pw |

**Description:**
Implement hardware-accelerated screen capture for each host OS: DXGI Desktop Duplication API (Windows), ScreenCaptureKit (macOS), KMS/DRM + PipeWire (Linux).

**Files:**
```
+ helix-capture/go.mod                         # Module: github.com/helixplay/helix-capture
+ helix-capture/pkg/capture/interface.go       # Capture interface
+ helix-capture/pkg/capture/dxgi/dda.go        # DXGI DDA (Windows)
+ helix-capture/pkg/capture/dxgi/texture.go    # DirectX texture management
+ helix-capture/pkg/capture/dxgi/mapper.go     # GPU texture mapper
+ helix-capture/pkg/capture/screencapturekit/  # macOS ScreenCaptureKit
+ helix-capture/pkg/capture/kms/kms.go         # Linux KMS/DRM
+ helix-capture/pkg/capture/pipewire/pw.go     # Linux PipeWire
+ helix-capture/pkg/capture/vulkan/vulkan.go   # Vulkan capture (cross-platform)
```

**Windows (DXGI DDA):**
- `IDXGIOutputDuplication::AcquireNextFrame()` for frame capture
- `ID3D11Device` texture sharing with encoder
- Hardware cursor compositing overlay
- HDR metadata extraction from `DXGI_OUTPUT_DESC`
- Latency target: < 1ms capture-to-texture

**macOS (ScreenCaptureKit):**
- `SCStream` with `SCContentFilter` for display capture
- `IOSurface` texture sharing
- ProRes/HEVC hardware encode via VideoToolbox
- Latency target: < 2ms capture-to-surface

**Linux (KMS + PipeWire):**
- DRM dumb buffer or `DRM_FORMAT_MOD_LINEAR` for GPU buffers
- PipeWire for Wayland compositor capture
- DMA-BUF fd passing for zero-copy
- Latency target: < 1ms capture-to-buffer

**Acceptance Criteria:**
1. Capture latency per platform within targets (see above)
2. 4K60 capture sustained without frame drops
3. HDR metadata correctly extracted and forwarded to encoder
4. Cursor capture: hardware cursor composited correctly
5. Multi-display: capture from selected display

---

#### B3-T02: Hardware Encoder Factory

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T02 |
| **Title** | Hardware Encoder Factory (NVENC / QSV / AMF / VideoToolbox / VAAPI) |
| **Effort** | 3pw |

**Description:**
Implement hardware encoder factory that auto-detects available encoders and selects optimal encoder based on codec, quality, and latency requirements.

**Files:**
```
+ helix-encode/go.mod                          # Module: github.com/helixplay/helix-encode
+ helix-encode/pkg/encoder/factory.go          # Encoder factory
+ helix-encode/pkg/encoder/interface.go        # Encoder interface
+ helix-encode/pkg/nvenc/nvenc.go              # NVIDIA NVENC
+ helix-encode/pkg/nvenc/session.go            # NVENC session management
+ helix-encode/pkg/nvenc/preset.go             # NVENC preset definitions
+ helix-encode/pkg/qsv/qsv.go                  # Intel QSV
+ helix-encode/pkg/amf/amf.go                  # AMD AMF
+ helix-encode/pkg/videotoolbox/vt.go          # Apple VideoToolbox
+ helix-encode/pkg/vaapi/vaapi.go              # Linux VAAPI
+ helix-encode/pkg/sw/fallback.go              # Software fallback (SVT-AV1, x265)
```

**Encoder Selection Matrix:**

| GPU | H.264 | HEVC | AV1 | Preferred |
|-----|-------|------|-----|-----------|
| NVIDIA RTX 40xx | NVENC | NVENC | NVENC | AV1 |
| NVIDIA RTX 30xx | NVENC | NVENC | N/A | HEVC |
| Intel Arc | QSV | QSV | QSV | AV1 |
| Intel 12th+ Gen | QSV | QSV | N/A | HEVC |
| AMD RX 7000 | AMF | AMF | AMF | AV1 |
| AMD RX 6000 | AMF | AMF | N/A | HEVC |
| Apple M1/M2/M3 | VT | VT | VT | HEVC |

**NVENC Presets:**

| Preset | Use Case | Target Quality | Latency |
|--------|----------|----------------|---------|
| P1 (Fastest) | Lowest latency | Lower | < 1ms |
| P2 | Low latency | Good | < 2ms |
| P4 (Default) | Balanced | Better | < 4ms |
| P6 | Quality | Best | < 8ms |
| P7 (Slowest) | Recording | Best | N/A |

**Acceptance Criteria:**
1. Encoder auto-detection: correct encoder selected on all GPU types
2. NVENC P1 latency: < 1ms (4K H.264), < 2ms (4K HEVC)
3. Encoder fallback: software encoder if hardware unavailable
4. Encoder hot-swap: change encoder without session restart
5. All encoders produce valid bitstreams (validated by decoder)

---

#### B3-T03: Dual-Path Encoding (Stream + Record)

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T03 |
| **Title** | Dual-Path Simultaneous Encode |
| **Effort** | 2pw |

**Description:**
(See P09-T02 for full specification) Summary: Single NVENC session produces both low-latency stream and high-quality recording output using split encoding.

**Files:**
```
+ helix-encode/pkg/dual/path.go                # Dual-path controller
+ helix-encode/pkg/dual/split.go               # Frame split logic
+ helix-encode/pkg/dual/output.go              # Dual output management
```

---

#### B3-T04: Thermal-Aware Quality Scaling

| Attribute | Detail |
|-----------|--------|
| **ID** | B3-T04 |
| **Title** | GPU Thermal-Aware Dynamic Quality Scaling |
| **Effort** | 1.5pw |

**Description:**
Implement thermal monitoring and dynamic quality scaling to prevent GPU thermal throttling. Reduce encode quality/resolution before thermal limit is reached.

**Files:**
```
+ helix-encode/pkg/thermal/monitor.go           # GPU temperature monitor
+ helix-encode/pkg/thermal/controller.go        # Thermal control loop
+ helix-encode/pkg/thermal/policy.go            # Scaling policies
+ helix-encode/pkg/thermal/nvml.go              # NVIDIA NVML bindings
+ helix-encode/pkg/thermal/amdsmi.go            # AMD SMI bindings
```

**Thermal Scaling Policy:**

| GPU Temp | Action |
|----------|--------|
| < 70°C | Full quality |
| 70-75°C | Reduce preset by 1 step |
| 75-80°C | Reduce resolution (4K→1440p or 1440p→1080p) |
| 80-83°C | Reduce bitrate by 25% |
| > 83°C | Emergency: minimum quality, log alert |

**Acceptance Criteria:**
1. Thermal throttling time < 2% of session time (measured)
2. Quality reduction: smooth transition (no visible artifact burst)
3. Recovery: quality restored within 30 seconds of temperature drop
4. All GPU vendors supported (NVIDIA, AMD, Intel)

---

### B4: Client Architecture Implementation

---

#### B4-T01: Shared Go Core Compilation

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T01 |
| **Title** | Shared Go Core (c-shared / native / WASM) |
| **Effort** | 2pw |

**Description:**
Implement shared Go core library compiled to multiple targets: `c-shared` (for Flutter FFI), native (for Wails), and WASM (for Angular web). Single codebase, platform-specific build tags.

**Files:**
```
+ helix-core/go.mod                            # Module: github.com/helixplay/helix-core
+ helix-core/Makefile                          # Multi-target build
+ helix-core/pkg/stream/decoder.go             # Video decode abstraction
+ helix-core/pkg/stream/renderer.go            # Frame renderer interface
+ helix-core/pkg/input/client.go               # Input client (send to host)
+ helix-core/pkg/net/webrtc.go                 # WebRTC client
+ helix-core/pkg/net/quic.go                   # QUIC client
+ helix-core/pkg/audio/render.go               # Audio renderer
+ helix-core/build/cshared.go                  # c-shared build directives
+ helix-core/build/wasm.go                     # WASM build directives
```

**Build Targets:**
```makefile
# Makefile targets
build-cshared-linux:
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -o libhelix.so

build-cshared-darwin:
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -o libhelix.dylib

build-cshared-windows:
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -o helix.dll

build-wasm:
    GOOS=js GOARCH=wasm go build -o helix.wasm

build-native:
    go build -o helix
```

**Acceptance Criteria:**
1. Single `go test ./...` passes for all build targets
2. C-shared library exports clean C API (< 20 functions)
3. WASM module < 10MB compressed (downloadable)
4. All targets share identical protocol implementation
5. Build time for all targets < 5 minutes

---

#### B4-T02: Wails Desktop Client

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T02 |
| **Title** | Wails v3 Desktop Client (Windows, macOS, Linux) |
| **Effort** | 2.5pw |

**Description:**
Implement Wails-based desktop client using the shared Go core. Replace any stubs with full implementation: WebRTC, hardware decode, controller input, audio output.

**Files:**
```
+ helix-wails/go.mod                           # Module: github.com/helixplay/helix-wails
+ helix-wails/main.go                          # Wails application entry
+ helix-wails/app.go                           # Wails app configuration
+ helix-wails/frontend/src/main.ts             # Frontend entry
+ helix-wails/frontend/src/App.svelte          # Main app component
+ helix-wails/frontend/src/stream/             # Streaming components
+ helix-wails/frontend/src/input/              # Input handling
+ helix-wails/frontend/src/settings/           # Settings UI
+ helix-wails/frontend/src/library/            # Game library UI
+ helix-wails/frontend/wailsjs/go/             # Wails Go bindings
```

**Acceptance Criteria:**
1. Windows, macOS, Linux builds from single codebase
2. Hardware decode: DXVA2/D3D11VA (Win), VideoToolbox (Mac), VAAPI (Linux)
3. Controller input via helix-input (DirectInput/XInput on Win, IOKit on Mac, evdev on Linux)
4. App size < 50MB (compressed installer)
5. Startup time < 3 seconds (cold)

---

#### B4-T03: Flutter Mobile Client

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T03 |
| **Title** | Flutter Mobile Client (Android, iOS) with Go FFI |
| **Effort** | 2.5pw |

**Description:**
Implement Flutter mobile client with Go FFI integration. Use `c-shared` library for streaming core. Implement platform-specific video rendering (SurfaceView/TextureView on Android, CVPixelBuffer on iOS).

**Files:**
```
+ helix-flutter/pubspec.yaml                   # Flutter dependencies
+ helix-flutter/lib/main.dart                  # App entry
+ helix-flutter/lib/core/bridge.dart           # Go FFI bridge
+ helix-flutter/lib/stream/player.dart         # Video player widget
+ helix-flutter/lib/stream/renderer.dart       # Platform renderer
+ helix-flutter/lib/input/controller.dart      # Controller input
+ helix-flutter/lib/screens/library.dart       # Game library
+ helix-flutter/lib/screens/stream.dart        # Streaming screen
+ helix-flutter/android/app/src/main/kotlin/   # Android platform code
+ helix-flutter/ios/Runner/                    # iOS platform code
```

**Acceptance Criteria:**
1. Android API 28+ and iOS 14+ from single Flutter codebase
2. Hardware decode: MediaCodec (Android), VideoToolbox (iOS)
3. Bluetooth controller pairing: DualSense, Xbox, MFi
4. Touch overlay for games without controller
5. App size: Android < 30MB, iOS < 40MB

---

#### B4-T04: Angular + Go-WASM Web Client

| Attribute | Detail |
|-----------|--------|
| **ID** | B4-T04 |
| **Title** | Angular Web Client with Go WASM + WebCodecs |
| **Effort** | 2.5pw |

**Description:**
Implement Angular web client using Go-compiled WASM for protocol handling and WebCodecs API for hardware-accelerated video decode. WebTransport for network layer.

**Files:**
```
+ helix-angular/package.json                   # npm dependencies
+ helix-angular/angular.json                   # Angular config
+ helix-angular/src/main.ts                    # Entry point
+ helix-angular/src/app/app.module.ts          # App module
+ helix-angular/src/app/stream/                # Streaming module
+ helix-angular/src/app/stream/webrtc.service.ts   # WebRTC client
+ helix-angular/src/app/stream/decoder.service.ts  # WebCodecs decoder
+ helix-angular/src/app/stream/renderer.ts     # Canvas/WebGL renderer
+ helix-angular/src/app/input/                 # Input handling
+ helix-angular/src/assets/wasm/               # Go WASM output
+ helix-angular/go.mod                         # Go WASM module
```

**WebCodecs Integration:**
```typescript
// VideoDecoder for H.264/HEVC/AV1
const decoder = new VideoDecoder({
    output: handleDecodedFrame,
    error: handleDecodeError,
});

decoder.configure({
    codec: 'avc1.640033',  // H.264 High Profile Level 5.1
    hardwareAcceleration: 'prefer-hardware',
});
```

**Acceptance Criteria:**
1. Chrome 94+, Firefox 120+, Safari 17+ supported
2. WebCodecs hardware decode: < 5ms decode latency
3. WebTransport for network (fallback to WebRTC datachannels)
4. Go WASM module handles protocol, encryption, input serialization
5. Controller support via WebHID (Chrome) + Gamepad API (all browsers)

---

### B5: Catalog & Content Pipeline

---

#### B5-T01: Multi-Source Metadata Aggregation

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T01 |
| **Title** | IGDB / SteamGridDB / Steam / RAWG Metadata Aggregation |
| **Effort** | 2pw |

**Description:**
Implement metadata aggregation from multiple game databases. Merge and deduplicate entries. Build unified game catalog with rich metadata.

**Files:**
```
+ helix-catalog/go.mod                         # Module: github.com/helixplay/helix-catalog
+ helix-catalog/pkg/sources/igdb.go            # IGDB API client
+ helix-catalog/pkg/sources/steamgriddb.go     # SteamGridDB client
+ helix-catalog/pkg/sources/steam.go           # Steam API client
+ helix-catalog/pkg/sources/rawg.go            # RAWG API client
+ helix-catalog/pkg/merge/engine.go            # Deduplication engine
+ helix-catalog/pkg/merge/score.go             # Match scoring
+ helix-catalog/pkg/models/game.go             # Unified game model
+ helix-catalog/pkg/sync/scheduler.go          # Periodic sync scheduler
+ helix-catalog/pkg/api/catalog.go             # Catalog API
```

**Metadata Fields:**
```go
type Game struct {
    ID              string
    Title           string
    Description     string
    ReleaseDate     time.Time
    Genres          []string
    Platforms       []string
    Developers      []string
    Publishers      []string
    Ratings         map[string]float64  // ESRB, PEGI, Metacritic
    CoverURL        string              // 4K WebP/AVIF
    ArtworkURLs     []string            // Screenshots
    VideoURLs       []string            // Trailers
    SteamAppID      string
    IGDBID          int
    RAWGID          int
    Tags            []string
    Series          string
}
```

**Acceptance Criteria:**
1. Catalog covers > 50,000 games from aggregated sources
2. Deduplication accuracy > 95% (measured via manual sample)
3. Metadata freshness: sync with sources every 24 hours
4. API response time: < 100ms for catalog search
5. Graceful degradation if source API is unavailable

---

#### B5-T02: 4K WebP/AVIF Asset Management

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T02 |
| **Title** | 4K WebP/AVIF Image Asset Pipeline |
| **Effort** | 1.5pw |

**Description:**
Implement asset pipeline that fetches, converts, and serves game artwork in WebP and AVIF formats. Responsive sizing, lazy loading, CDN integration.

**Files:**
```
+ helix-catalog/pkg/assets/pipeline.go         # Asset processing pipeline
+ helix-catalog/pkg/assets/convert.go          # Image format conversion
+ helix-catalog/pkg/assets/resize.go           # Responsive resizing
+ helix-catalog/pkg/assets/storage.go          # Asset storage (S3 + CDN)
+ helix-catalog/pkg/assets/serve.go            # Asset serving with format negotiation
```

**Asset Formats:**

| Format | Role | Quality | Size vs JPEG |
|--------|------|---------|-------------|
| AVIF | Primary (modern clients) | 85 | -60% |
| WebP | Fallback (older clients) | 85 | -30% |
| JPEG | Legacy fallback | 90 | Baseline |

**Sizes:** 256x384 (cover), 1920x1080 (screenshot), 3840x2160 (hero)

**Acceptance Criteria:**
1. AVIF served to supporting browsers (Chrome 85+, Firefox 93+, Safari 16+)
2. WebP served to supporting browsers without AVIF
3. Image response time: < 200ms from CDN edge
4. Original quality preserved in 4K assets
5. Storage: AVIF + WebP pre-generated, no on-the-fly conversion

---

#### B5-T03: Search Implementation

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T03 |
| **Title** | Full-Text Search (SQLite FTS5 + Meilisearch) |
| **Effort** | 1.5pw |

**Description:**
Implement two-tier search: SQLite FTS5 for local/offline search, Meilisearch for server-side catalog search. Typo tolerance, faceting, fuzzy matching.

**Files:**
```
+ helix-catalog/pkg/search/fts5.go             # SQLite FTS5 local search
+ helix-catalog/pkg/search/meilisearch.go      # Meilisearch server search
+ helix-catalog/pkg/search/index.go            # Index management
+ helix-catalog/pkg/search/suggestions.go      # Autocomplete/suggestions
```

**Acceptance Criteria:**
1. Local search: SQLite FTS5, works offline, < 50ms response
2. Server search: Meilisearch, typo-tolerant, < 100ms response
3. Autocomplete suggestions: < 30ms
4. Faceted search by genre, platform, release year, rating
5. Search index updated within 5 minutes of catalog change

---

#### B5-T04: Per-Tenant Catalog Overlays

| Attribute | Detail |
|-----------|--------|
| **ID** | B5-T04 |
| **Title** | Per-Tenant Catalog Customization |
| **Effort** | 1pw |

**Description:**
Allow tenants to customize their game catalog: add/remove games, custom artwork, pricing, featured sections. Overlay on top of global catalog.

**Files:**
```
+ helix-tenant/pkg/catalog/overlay.go          # Catalog overlay engine
+ helix-tenant/pkg/catalog/custom.go           # Custom game entries
+ helix-tenant/pkg/catalog/featured.go         # Featured sections
+ helix-tenant/pkg/catalog/pricing.go          # Per-tenant pricing
```

**Acceptance Criteria:**
1. Tenant can hide games from global catalog
2. Tenant can add custom games (not in global catalog)
3. Tenant can override artwork for any game
4. Tenant can set custom pricing/subscription model
5. Tenant can create featured sections and curated lists

---

### B6: White-Label & Theming

---

#### B6-T01: 3-Tier Design Token System

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T01 |
| **Title** | 3-Tier Design Tokens (Primitive → Semantic → Component) |
| **Effort** | 1.5pw |

**Description:**
Implement 3-tier design token architecture using Style Dictionary v4. Tokens define all visual properties: colors, typography, spacing, elevation, motion.

**Files:**
```
+ helix-theme/go.mod                           # Module: github.com/helixplay/helix-theme
+ helix-theme/tokens/primitive/colors.json     # Primitive color tokens
+ helix-theme/tokens/primitive/typography.json # Primitive type tokens
+ helix-theme/tokens/primitive/spacing.json    # Primitive spacing tokens
+ helix-theme/tokens/semantic/light.json       # Semantic tokens (light)
+ helix-theme/tokens/semantic/dark.json        # Semantic tokens (dark)
+ helix-theme/tokens/component/button.json     # Component tokens
+ helix-theme/tokens/component/card.json
+ helix-theme/tokens/component/input.json
+ helix-theme/build.js                         # Style Dictionary v4 build
+ helix-theme/config.json                      # SD configuration
```

**Token Hierarchy:**
```
primitive/
  color.blue.500 = "#2196F3"
  color.red.500 = "#F44336"
  spacing.4 = "16px"

semantic/
  color.primary = { primitive.color.blue.500 }
  color.error = { primitive.color.red.500 }
  spacing.section = { primitive.spacing.4 }

component/
  button.background = { semantic.color.primary }
  button.padding = { semantic.spacing.section }
```

**Acceptance Criteria:**
1. All visual properties defined as tokens (zero hardcoded values)
2. Theme switch (light/dark): instant, no page reload
3. Custom tenant theme generated from brand colors (< 5 minutes)
4. Token outputs: CSS variables, JSON, Dart, TypeScript
5. Style Dictionary build: < 30 seconds

---

#### B6-T02: Material Design 3 Integration

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T02 |
| **Title** | Material Design 3 (Material You) Integration |
| **Effort** | 1pw |

**Description:**
Integrate Material Design 3 across all three client platforms. Use M3 components as base, customize with design tokens.

**Files:**
```
+ helix-theme/tokens/m3/ref.json               # M3 reference tokens
+ helix-theme/tokens/m3/sys.json               # M3 system tokens
~ helix-wails/frontend/src/theme/m3.ts         # M3 theme (Wails)
~ helix-flutter/lib/theme/m3.dart              # M3 theme (Flutter)
~ helix-angular/src/theme/m3.scss              # M3 theme (Angular)
```

**Acceptance Criteria:**
1. All UI components use M3 design language
2. Dynamic color (Material You): theme derived from game artwork on Android
3. Consistent visual language across all three platforms
4. Accessibility: WCAG 2.1 AA compliance (contrast ratios)

---

#### B6-T03: Style Dictionary v4 Build Pipeline

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T03 |
| **Title** | Style Dictionary v4 Build & Distribution |
| **Effort** | 0.5pw |

**Description:**
Automated token build pipeline producing platform-specific outputs.

**Files:**
```
+ .github/workflows/tokens-build.yml           # Token build CI
+ helix-theme/package.json                     # npm dependencies
```

**Build Outputs:**
| Platform | Format | Destination |
|----------|--------|-------------|
| Web (Wails/Angular) | CSS custom properties | `*.css` |
| Flutter | Dart class | `*.dart` |
| Design tools | JSON | Figma plugin |

**Acceptance Criteria:**
1. CI builds all token outputs on every token change
2. Output files distributed to client repos via automated PR
3. Token validation: no undefined references, no circular dependencies

---

#### B6-T04: Per-Tenant Identity

| Attribute | Detail |
|-----------|--------|
| **ID** | B6-T04 |
| **Title** | Per-Tenant Brand Identity System |
| **Effort** | 1pw |

**Description:**
Complete white-label identity: logo, brand colors, fonts, app icon, splash screen. Generated from tenant configuration.

**Files:**
```
+ helix-tenant/pkg/branding/generator.go       # Brand asset generator
+ helix-tenant/pkg/branding/logo.go            # Logo processing
+ helix-tenant/pkg/branding/colors.go          # Brand color extraction
+ helix-tenant/pkg/branding/fonts.go           # Font loading
```

**Acceptance Criteria:**
1. Tenant brand colors extracted from logo (dominant color algorithm)
2. App icon generated with tenant logo
3. Splash screen themed with tenant brand
4. All branding applied within 30 seconds of tenant config change

---

### B7: Operations & Observability

---

#### B7-T01: OpenTelemetry Integration

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T01 |
| **Title** | OpenTelemetry Tracing & Metrics |
| **Effort** | 1.5pw |

**Description:**
Implement OpenTelemetry tracing and metrics across all services. Distributed tracing for request flows, custom metrics for gaming-specific KPIs.

**Files:**
```
+ helix-observability/go.mod                   # Module: github.com/helixplay/helix-observability
+ helix-observability/pkg/trace/provider.go    # OTel trace provider
+ helix-observability/pkg/metrics/provider.go  # OTel metrics provider
+ helix-observability/pkg/metrics/gaming.go    # Gaming-specific metrics
+ helix-observability/pkg/log/otel.go          # OTel log correlation
```

**Gaming-Specific Metrics:**
```go
var (
    FrameLatency = meter.Float64Histogram("helix.frame_latency_ms",
        "Frame glass-to-glass latency")
    InputLatency = meter.Float64Histogram("helix.input_latency_ms",
        "Controller input latency")
    EncodeTime = meter.Float64Histogram("helix.encode_time_ms",
        "Frame encode time")
    NetworkRTT = meter.Float64Histogram("helix.network_rtt_ms",
        "Network round-trip time")
    GpuUtilization = meter.Float64ObservableGauge("helix.gpu_utilization",
        "GPU utilization percent")
    ThermalTemp = meter.Float64ObservableGauge("helix.gpu_temperature_c",
        "GPU temperature Celsius")
)
```

**Acceptance Criteria:**
1. All API requests traced with distributed trace IDs
2. Gaming metrics collected every frame (no sampling in hot path)
3. Trace sampling: 100% for errors, 1% for success (configurable)
4. Export to Jaeger + Prometheus
5. Trace correlation across all 29 submodules

---

#### B7-T02: Prometheus Metrics

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T02 |
| **Title** | Prometheus Metrics Export & Alerting |
| **Effort** | 1pw |

**Description:**
Prometheus metrics export with custom collectors for gaming KPIs. Grafana dashboards and alert rules.

**Files:**
```
+ helix-observability/pkg/prometheus/registry.go   # Prometheus registry
+ helix-observability/pkg/prometheus/collectors.go # Custom collectors
+ infra/monitoring/grafana/dashboards/             # Grafana dashboards
+ infra/monitoring/prometheus/rules.yml            # Alert rules
```

**Alert Rules:**

| Alert | Condition | Severity |
|-------|-----------|----------|
| HighLatency | p99 frame latency > 50ms | warning |
| CriticalLatency | p99 frame latency > 100ms | critical |
| GpuThermal | GPU temp > 83°C | warning |
| HighErrorRate | Error rate > 1% | critical |
| DiskFull | Disk usage > 90% | warning |
| MemoryPressure | Memory usage > 95% | critical |

**Acceptance Criteria:**
1. All metrics exposed on `/metrics` endpoint
2. Grafana dashboards for: streaming quality, GPU health, network, sessions
3. AlertManager routes alerts to PagerDuty/Slack
4. Alert firing latency < 30 seconds from threshold breach

---

#### B7-T03: Structured JSON Logging

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T03 |
| **Title** | Structured JSON Logging with Correlation IDs |
| **Effort** | 0.5pw |

**Description:**
Structured JSON logging with request correlation IDs, tenant context, and gaming-specific fields.

**Files:**
```
+ helix-observability/pkg/log/logger.go        # Structured logger
+ helix-observability/pkg/log/fields.go        # Gaming log fields
+ helix-observability/pkg/log/middleware.go    # HTTP/gRPC logging middleware
```

**Log Schema:**
```json
{
    "ts": "2025-01-15T10:30:00.000Z",
    "level": "info",
    "msg": "session.started",
    "trace_id": "abc123",
    "tenant_id": "tenant-42",
    "session_id": "sess-789",
    "user_id": "user-456",
    "game_id": "game-123",
    "codec": "av1",
    "resolution": "3840x2160",
    "fps": 60,
    "host_id": "host-gpu-01",
    "region": "us-east-1"
}
```

**Acceptance Criteria:**
1. All logs structured JSON (no plaintext)
2. Correlation ID propagated across all service boundaries
3. Tenant ID in every log entry (for multi-tenant filtering)
4. Log aggregation: Fluent Bit → Loki / ELK
5. Log query response: < 2 seconds for 24-hour search

---

#### B7-T04: NATS Event Bus

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T04 |
| **Title** | NATS JetStream Event Bus |
| **Effort** | 1pw |

**Description:**
NATS JetStream as primary event bus for async communication between services. Tenant-scoped subjects, durable consumers, exactly-once processing.

**Files:**
```
+ helix-events/go.mod                          # Module: github.com/helixplay/helix-events
+ helix-events/pkg/nats/client.go              # NATS client
+ helix-events/pkg/nats/jetstream.go           # JetStream management
+ helix-events/pkg/nats/publisher.go           # Event publisher
+ helix-events/pkg/nats/consumer.go            # Event consumer
+ helix-events/pkg/nats/events.go              # Event type definitions
```

**Event Types:**
```go
const (
    EventSessionStarted   = "helix.session.started"
    EventSessionEnded     = "helix.session.ended"
    EventFrameEncoded     = "helix.frame.encoded"
    EventInputReceived    = "helix.input.received"
    EventQualityChanged   = "helix.quality.changed"
    EventRecordingSaved   = "helix.recording.saved"
    EventUserAuthenticated = "helix.user.authenticated"
)
```

**Subject Topology:**
```
events.{tenant_id}.{event_type}
metrics.{tenant_id}.{metric_name}
commands.{service_id}.{command_type}
```

**Acceptance Criteria:**
1. Event publish latency < 1ms (localhost NATS)
2. Tenant-scoped subjects enforce isolation
3. Durable consumers: no event loss on consumer restart
4. Exactly-once semantics for billing events
5. JetStream retention: 7 days for events, 30 days for audit

---

#### B7-T05: Four-Mirror Git Topology Automation

| Attribute | Detail |
|-----------|--------|
| **ID** | B07-T05 |
| **Title** | Four-Mirror Git Repository Automation |
| **Effort** | 0.5pw |

**Description:**
Automated synchronization of all 29 submodules across four Git hosting platforms.

**Files:**
```
+ scripts/mirror/sync.sh                       # Sync orchestrator
+ scripts/mirror/verify.sh                     # Verification script
+ .github/workflows/mirror-sync.yml            # Post-push sync trigger
```

**Acceptance Criteria:**
1. All 29 repos synced to 4 mirrors within 5 minutes of primary push
2. Tags, releases, and branch protection rules synced
3. Automated recovery on sync failure (retry + alert)
4. Weekly verification report

---

#### B7-T06: GitHub Projects + GitLab Tracking

| Attribute | Detail |
|-----------|--------|
| **ID** | B7-T06 |
| **Title** | Cross-Platform Project Tracking |
| **Effort** | 0.5pw |

**Description:**
Bidirectional sync between GitHub Projects (primary) and GitLab issues (mirror tracking).

**Files:**
```
+ scripts/tracking/sync.sh                     # Issue sync script
+ .github/workflows/tracking-sync.yml          # Sync workflow
```

**Acceptance Criteria:**
1. GitHub Project status changes reflected in GitLab within 10 minutes
2. GitLab issue comments synced to GitHub
3. Sprint/milestone alignment across both platforms

---

## Appendix A: Submodule Registry

### 29 Public Go Submodules

| # | Module | Purpose | Priority | Status |
|---|--------|---------|----------|--------|
| 1 | `helix-core` | Shared streaming core | P1 | Planned |
| 2 | `helix-stream` | WebRTC/QUIC streaming | P1 | Planned |
| 3 | `helix-capture` | Screen capture (DXGI/SCK/PipeWire) | P1 | Planned |
| 4 | `helix-encode` | Hardware encoder factory | P1 | Planned |
| 5 | `helix-input` | Controller input (1kHz DualSense) | P1 | Planned |
| 6 | `helix-audio` | Audio pipeline (Opus MultiStream) | P2 | Planned |
| 7 | `helix-record` | Recording & replay | P2 | Planned |
| 8 | `helix-auth` | OAuth2/OIDC/RBAC | P1 | Planned |
| 9 | `helix-tenant` | Multi-tenant isolation | P1 | Planned |
| 10 | `helix-catalog` | Game metadata catalog | P2 | Planned |
| 11 | `helix-theme` | Design tokens & theming | P2 | Planned |
| 12 | `helix-deploy` | Canary deployment | P2 | Planned |
| 13 | `helix-vault` | Secrets management | P1 | Planned |
| 14 | `helix-audit` | Audit logging | P1 | Planned |
| 15 | `helix-events` | NATS event bus | P1 | Planned |
| 16 | `helix-observability` | OpenTelemetry/metrics | P1 | Planned |
| 17 | `helix-rtos` | RTOS scheduling utilities | P1 | Planned |
| 18 | `helix-iouring` | io_uring async I/O | P1 | Planned |
| 19 | `helix-lockfree` | Lock-free data structures | P1 | Planned |
| 20 | `helix-shm` | Shared memory IPC | P1 | Planned |
| 21 | `helix-mempool` | Slab memory pools | P1 | Planned |
| 22 | `helix-allocator` | Custom allocators | P1 | Planned |
| 23 | `helix-gpu-direct` | GPU Direct zero-copy | P2 | Planned |
| 24 | `helix-wails` | Wails desktop client | P2 | Planned |
| 25 | `helix-flutter` | Flutter mobile/TV client | P2 | Planned |
| 26 | `helix-angular` | Angular+WASM web client | P2 | Planned |
| 27 | `helix-r18-safeexec` | R-18 security framework | P1 | Planned |
| 28 | `sunshine` | Sunshine++ host agent (fork) | P1 | Planned |
| 29 | `helix-docs` | Documentation site | P2 | Planned |

---

## Appendix B: Dependency Graph

### Cross-Module Dependencies (Key Paths)

```
helix-r18-safeexec (root)
  ├── helix-core
  │     ├── helix-stream
  │     │     ├── helix-capture
  │     │     ├── helix-encode
  │     │     ├── helix-audio
  │     │     └── helix-record
  │     ├── helix-input
  │     └── helix-shm
  ├── helix-rtos
  │     ├── helix-iouring
  │     ├── helix-lockfree
  │     ├── helix-mempool
  │     └── helix-allocator
  ├── helix-auth
  │     ├── helix-vault
  │     └── helix-tenant
  │           ├── helix-catalog
  │           ├── helix-theme
  │           └── helix-events
  ├── helix-observability
  │     └── helix-audit
  ├── helix-deploy
  └── helix-gpu-direct

Client Modules (depend on helix-core):
  ├── helix-wails
  ├── helix-flutter
  └── helix-angular

Host Agent:
  └── sunshine → depends on helix-capture, helix-encode, helix-input,
                   helix-audio, helix-shm, helix-rtos
```

---

*End of Document*

**Document History:**
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2025-01 | Architecture Team | Initial release - Advanced Phases P07-P13 + Architecture Deep-Dive B1-B7 |
-e 

---


# Comprehensive Testing & Anti-Bluff Strategy

## HelixPlay Cloud Gaming Platform — Implementation Plan Section

**Document Version**: 1.0  
**Constitution Reference**: v2.1.0  
**Date**: 2026-05-02  
**Classification**: Mandatory — Non-Overrideable CI Lane  

---

## Table of Contents

- [Section 1: Testing Philosophy & Anti-Bluff Constitution](#section-1-testing-philosophy--anti-bluff-constitution)
- [Section 2: The Ten Test Types — Detailed Implementation](#section-2-the-ten-test-types--detailed-implementation)
- [Section 3: Test Matrix (29 Submodules x 10 Types)](#section-3-test-matrix-29-submodules--10-types)
- [Section 4: Anti-Bluff Infrastructure Implementation](#section-4-anti-bluff-infrastructure-implementation)
- [Section 5: HelixQA Autonomous QA Integration](#section-5-helixqa-autonomous-qa-integration)
- [Section 6: CI/CD Pipeline Design](#section-6-cicd-pipeline-design)
- [Section 7: Fixing Current Issues](#section-7-fixing-current-issues)
- [Section 8: Challenges Implementation Plan](#section-8-challenges-implementation-plan)

---

## Section 1: Testing Philosophy & Anti-Bluff Constitution

### 1.1 The Prime Directive

**Constitutional Text (Constitution v2.1.0, Clause 1 — Anti-Bluff Pledge)**:

> *"Green tests MUST guarantee real, end-user-usable behavior. The project has suffered from 'green tests on broken features' — tests that pass while the actual feature does not work. This MUST NEVER happen again. Every test that passes MUST correspond to observable, verifiable functionality that a real user can interact with."*

**Historical Context**: HelixPlay previously had 267+ vacuous tests — tests that passed but verified nothing meaningful. These included `assert.True(t, true)`, constructor-only tests that verified an object was created but never exercised its behavior, and mock-only integration tests where mocked dependencies returned hardcoded values. The result: all tests showed green while core streaming, input, and discovery features were non-functional.

**What This Means in Practice**:

1. **No test may assert a tautology** — `assert.True(t, true)`, `assert.Equal(t, 1, 1)`, `assert.Nil(t, nil)` are forbidden by constitutional law. Any such assertion detected in CI causes an immediate pipeline failure.

2. **Constructor-only tests are bluff tests** — A test that only verifies `require.NotNil(t, obj)` after calling `NewFoo()` proves the allocator works, not the feature. Every test must exercise at least one observable behavior of the constructed object.

3. **Mock-only integration tests are bluff tests** — Integration tests (Types 2–10) must use real dependencies. If an integration test mocks the database, the HTTP client, and the auth layer, it tests the mock wiring, not the integration.

4. **Passing tests must correlate to working features** — If a feature is broken for a real user, at least one test must fail. If all tests pass but the feature doesn't work, the test suite is bluffing and must be rebuilt.

5. **Usability evidence is mandatory** — Per Constitution §6.7, every feature must provide evidence of real usability: HelixQA visual assertion, manual screen recording, or Challenge scenario execution with anti-bluff validation.

### 1.2 Forbidden Patterns (Complete List)

#### 1.2.1 Forbidden Code Patterns

| Pattern | Severity | Detection Method | Example |
|---------|----------|-----------------|---------|
| Empty function body `{}` | **CRITICAL** | Static AST scan | `func (s *Server) Handle() {}` |
| `panic("not implemented")` stub | **CRITICAL** | Regex scan | `panic("not implemented")` |
| `return nil, fmt.Errorf("not implemented")` | **CRITICAL** | Regex scan | `return nil, errors.New("not implemented")` |
| `TODO` comment without issue reference | **WARNING** | Regex scan | `// TODO: fix this` |
| `FIXME` comment without issue reference | **WARNING** | Regex scan | `// FIXME: broken` |
| `XXX` or `HACK` comments | **WARNING** | Regex scan | `// HACK: workaround` |
| Standalone `tbd` or `TBD` | **WARNING** | Word-boundary scan | `status = tbd` |

**Enforcement**: `scripts/anti-bluff-scan.sh` Step 1 performs a full-source-tree regex scan for all patterns above. Matches are collected into a report. Any CRITICAL match fails the CI pipeline. WARNING matches are logged and tracked but do not fail the pipeline unless the count increases relative to the established baseline.

#### 1.2.2 Forbidden Test Patterns

| Pattern | Severity | Detection Method | Why It Is a Bluff |
|---------|----------|-----------------|-------------------|
| `assert.True(t, true)` | **BLOCKER** | Regex + AST | Asserts a tautology; always passes, proves nothing |
| `assert.Equal(t, true, true)` | **BLOCKER** | Regex + AST | Same as above with different syntax |
| `assert.Nil(t, nil)` | **BLOCKER** | Regex + AST | Asserts nil is nil; always passes |
| `assert.Equal(t, 1, 1)` | **BLOCKER** | Regex + AST | Literal compared to itself |
| Constructor-only `require.NotNil(t, obj)` | **CRITICAL** | Heuristic: test body has only construction + NotNil | Verifies allocation, not behavior |
| Mock-only integration test | **CRITICAL** | Heuristic: integration/e2e test directory uses mocks | Tests mock wiring, not real integration |
| No negative-leg test | **CRITICAL** | Challenge runner validation | Feature works for happy path but no test verifies failure detection |
| Empty test body `func TestX(t *testing.T) {}` | **BLOCKER** | AST scan | Test that passes without executing any code |
| Test with no assertions | **CRITICAL** | AST scan: count assert/require calls | Test executes code but never checks results |
| `assert.NoError(t, nil)` | **BLOCKER** | Regex scan | Asserts nil error is nil; tautology |

**Enforcement**: `scripts/anti-bluff-scan.sh` Step 5 performs the vacuous assertion scan. The bluff scanner self-test (`bluff_scanner_challenge.sh` Phase 1) verifies the scanner can detect hand-crafted bluff fixtures. If the scanner cannot find known bluff patterns, the pipeline fails — the test for bluff detection must itself not be a bluff.

#### 1.2.3 Forbidden Documentation Patterns

| Pattern | Severity | Detection Method |
|---------|----------|-----------------|
| Missing Anti-Bluff Verification section | **CRITICAL** | Documentation structure scan |
| Anti-Bluff section without actionable assertions | **WARNING** | Content heuristic |
| Placeholder text in verification section | **CRITICAL** | Regex: "TBD", "placeholder", "to be defined" |

**Enforcement**: `scripts/anti-bluff-scan.sh` Step 4 checks that all submodule documentation contains Anti-Bluff Verification sections. Missing sections are flagged and tracked.

#### 1.2.4 Enforcement Mechanisms

The anti-bluff enforcement operates at four levels:

1. **Local Prevention**: Claude Code `.claude/settings.json` Stop hook runs `claim-check.sh` before each commit, blocking R-18 forbidden commands.

2. **Pre-Commit Scan**: Developers can (and should) run `make anti-bluff` locally before pushing. This runs the scanner, anchor manifest validator, and mutation ratchet.

3. **CI Non-Overrideable Lane**: The `anti-bluff.yml` workflow (see Section 6.2) runs on every PR and cannot be bypassed. A failure here blocks merge regardless of other passing checks.

4. **Challenge Runner Gate**: The `ValidateAntiBluff()` function is called unconditionally (since v2.1.0, the `CHALLENGE_ANTIBLUFF_STRICT` toggle was removed) on every challenge result. Any challenge that reports Pass without evidence fails the challenge pipeline.

### 1.3 What is Required for Every Feature

#### 1.3.1 Observable Behavior Assertions

Every feature must have tests that verify **observable, externally-visible behavior**, not internal state. An observable behavior assertion checks something a user or external system could notice.

**Observable behaviors include**:
- An HTTP endpoint returns a specific status code and response body
- A database query returns the expected rows
- A file is created with expected content
- A gRPC stream delivers frames in the expected order
- A JWT token validates with the correct claims
- A game appears in the discovery list after enumeration
- A screen capture frame has the expected dimensions and pixel format

**Non-observable (internal state) assertions that are insufficient alone**:
- A private field was set to a specific value
- An internal counter incremented
- A mock was called with expected arguments

**Requirement**: At least 60% of assertions in any test file must verify observable behavior. The remaining 40% may verify internal state for diagnostic purposes.

#### 1.3.2 Negative-Leg Testing

For every feature, there must be a test that **deliberately breaks the feature and verifies the test suite detects the breakage**.

**Constitutional Text (§6.3)**:
> *"Automatic negative-leg fault injection: CI must break each feature and verify non-Unit tests fail."*

**Implementation**: The negative-leg fault injection system (detailed in Section 4.5) operates as follows:

1. For each feature under test, the CI pipeline creates a **mutated version** of the source code that introduces a deliberate defect (e.g., swaps `<` for `>`, removes a bounds check, changes a constant).

2. The full test suite (excluding Unit tests, which may test internals) runs against the mutated code.

3. **At least one test MUST fail**. If all tests pass on the mutated code, the test suite is bluffing — it doesn't actually verify the feature's behavior.

4. The mutation is reverted and the next feature is tested.

**Coverage Target**: 100% of production code features must have negative-leg verification.

#### 1.3.3 Usability Evidence Requirements

Per Constitution §6.7, every feature must provide one of three forms of usability evidence:

| Evidence Type | Required For | Method | Verification |
|---------------|-------------|--------|-------------|
| **HelixQA Visual Assertion** | UI-facing features | Automated screenshot/screen recording capture with OpenCV-based verification | HelixQA pipeline produces pass/fail with visual diff |
| **Manual Screen Recording** | Complex interactive features | Human records screen session following test script | Recording reviewed and attached to feature sign-off |
| **Challenge Scenario Execution** | All backend features | Challenges repository scenario runs with anti-bluff validation | `ValidateAntiBluff()` passes with `RecordedActions` and non-empty assertions |

#### 1.3.4 Coverage Requirements

| Metric | Target | Measurement Method | Enforcement |
|--------|--------|-------------------|-------------|
| **Line Coverage** | 100% | Union of all 10 test types | SonarQube gate fails if < 100% |
| **Branch Coverage** | 100% | Union of all 10 test types | SonarQube gate fails if < 100% |
| **Function Coverage** | 100% | Union of all 10 test types | SonarQube gate fails if < 100% |
| **Mutation Score** | >= 85% | go-mutesting | `mutation_ratchet_challenge.sh` fails if < 85% |
| **Anti-Bluff Pass Rate** | 100% | `ValidateAntiBluff()` | Unconditional; any bluff result fails pipeline |

**Important**: Coverage is measured across the **union** of all 10 test types, not per type. A line covered by an E2E test counts even if no Unit test covers it. However, lines covered only by Unit tests (especially mock-based Unit tests) are flagged for additional integration/E2E coverage.

---

## Section 2: The Ten Test Types — Detailed Implementation

### Overview

The Constitution §6 mandates exactly 10 test types. Each type has a specific purpose, specific mock rules, and specific CI integration. No feature is considered complete until all applicable test types are implemented and passing.

| # | Test Type | Mocks Permitted | Scope | Required Coverage |
|---|-----------|----------------|-------|-------------------|
| 1 | Unit | Yes (with restrictions) | Single function/method | 100% branches of isolated logic |
| 2 | Integration | **NO** | Cross-component real dependencies | All interaction paths |
| 3 | E2E | **NO** | Full system path, production-like topology | All user journeys |
| 4 | Security | **NO** | Fuzzing, SAST, DAST | All attack surfaces |
| 5 | Benchmarking | **NO** | p50/p99/p999 latency, throughput | All performance-critical paths |
| 6 | Chaos | Limited (chaos injection only) | Fault injection in production-like env | All failure modes |
| 7 | Stress | **NO** | Load beyond capacity, 24h profiles | All resource-limited paths |
| 8 | Smoke | **NO** | Post-deploy 30-second sanity | All critical endpoints |
| 9 | Full Automation | **NO** | Entire pipeline unattended | All of the above combined |
| 10 | Challenges | **NO** | Production-equivalent scenarios | All challenge bank scenarios |

### 2.1 Unit Tests

**Purpose**: Verify isolated logic of a single function, method, or struct in complete isolation from external dependencies.

**What Mocks Are Permitted**:
- **Interfaces defined in the same module**: Mock implementations of interfaces that the code under test depends on, where the mock is defined in a `*_test.go` file or `test/` subdirectory.
- **External dependencies with non-deterministic behavior**: Time (`time.Now`), random number generation, UUID generation may be mocked or controlled via dependency injection.
- **Network/disk I/O**: For pure business logic tests, network clients and filesystem operations may be mocked.

**What Mocks Are FORBIDDEN**:
- Mocking the thing you are testing (the System Under Test itself)
- Mocking every dependency so the test only verifies mock wiring
- Mocking database operations when testing database query builders
- Mocking HTTP handlers when testing HTTP middleware

**Required Coverage**: 100% branch coverage of isolated business logic. Lines that are impossible to reach (e.g., defensive checks for conditions that can't occur in practice) must be annotated with `// unreachable: <reason>` and reviewed in PR.

**CI Integration**:
```yaml
# Part of ci.yml — Unit test job
unit-tests:
  name: Unit Tests
  runs-on: ubuntu-latest
  container: golang:1.26
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Fix replace directives
      run: ./scripts/fix-replace.sh
    - name: Run unit tests
      run: go test -count=1 -race -p 1 ./...
    - name: Upload coverage
      uses: actions/upload-artifact@v4
      with:
        name: unit-coverage
        path: coverage.out
```

**Example Test Pattern** (legitimate Unit test):
```go
func TestRateLimiter_Allow(t *testing.T) {
    // Arrange: Create a real rate limiter with a test clock
    clock := testclock.New()
    rl := NewRateLimiter(10, time.Second, WithClock(clock))

    // Act: Consume all 10 tokens
    for i := 0; i < 10; i++ {
        assert.True(t, rl.Allow(), "request %d should be allowed", i)
    }

    // Assert: 11th request rejected
    assert.False(t, rl.Allow(), "11th request should be rejected")

    // Act: Advance clock by 1 second
    clock.Advance(time.Second)

    // Assert: Bucket refilled, request allowed again
    assert.True(t, rl.Allow(), "request after refill should be allowed")
}
```

**Anti-Bluff Verification Method**:
- Run `go-mutesting` on the package. If mutants survive, the Unit test is not actually verifying the logic.
- Negative leg: Introduce a deliberate bug (e.g., change `>=` to `>`) and verify the Unit test fails.

**Directory Structure**:
```
<module>/
  pkg/<package>/
    <file>.go
    <file>_test.go          # Unit tests (colocated)
  tests/unit/               # Alternative: centralized unit tests
    <package>_test.go
```

### 2.2 Integration Tests

**Purpose**: Verify that multiple real components work together correctly. Integration tests exercise the actual interaction paths between modules with real (not mocked) dependencies.

**What Mocks Are Permitted**: **NONE**. Integration tests must use real dependencies:
- Real databases (SQLite in-memory or test-container PostgreSQL)
- real HTTP servers (`httptest.NewServer` is acceptable — it is a real HTTP server, not a mock)
- Real message queues (test-container RabbitMQ/NATS)
- Real caches (test-container Redis or in-memory)
- Real filesystem operations (temporary directories)

**Required Coverage**: All cross-component interaction paths. Every pair of components that communicate must have at least one integration test verifying that communication.

**CI Integration**:
```yaml
integration-tests:
  name: Integration Tests
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:16
      env:
        POSTGRES_PASSWORD: test
        POSTGRES_DB: helixtest
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
      ports: ['25432:5432']
    redis:
      image: redis:7-alpine
      ports: ['26379:6379']
    nats:
      image: nats:2-alpine
      ports: ['4222:4222']
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Fix replace directives
      run: ./scripts/fix-replace.sh
    - name: Run integration tests
      run: go test -count=1 -race -p 1 ./tests/integration/...
      env:
        HELIX_TEST_DB_DSN: postgres://postgres:test@localhost:25432/helixtest?sslmode=disable
        HELIX_TEST_REDIS_ADDR: localhost:26379
        HELIX_TEST_NATS_URL: nats://localhost:4222
```

**Example Test Pattern**:
```go
func TestAuthStorage_Integration(t *testing.T) {
    // Arrange: Real database connection
    db, err := database.Connect(os.Getenv("HELIX_TEST_DB_DSN"))
    require.NoError(t, err)
    defer db.Close()

    // Arrange: Real auth manager using real DB
    auth := auth.NewManager(db, auth.WithJWTSecret("test-secret"))

    // Act: Register a user
    token, err := auth.Register(ctx, "test@example.com", "password123")
    require.NoError(t, err)
    require.NotEmpty(t, token)

    // Act: Validate the token
    claims, err := auth.ValidateToken(ctx, token)
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", claims.Subject)

    // Act: Login with same credentials
    token2, err := auth.Login(ctx, "test@example.com", "password123")
    require.NoError(t, err)
    assert.NotEmpty(t, token2)

    // Negative leg: Wrong password
    _, err = auth.Login(ctx, "test@example.com", "wrongpassword")
    assert.Error(t, err)
}
```

**Anti-Bluff Verification Method**:
- Temporarily break the integration point (e.g., change the table name in one component) and verify the test fails.
- Verify no mocks are used: `grep -r "mock\|Mock" tests/integration/` should return zero matches.

**Directory Structure**:
```
<module>/
  tests/integration/
    <component_pair>_test.go    # e.g., auth_storage_test.go
    <flow>_test.go              # e.g., user_registration_flow_test.go
```

### 2.3 E2E Tests

**Purpose**: Verify complete user journeys from the external entry point through the entire system, using a production-like topology.

**What Mocks Are Permitted**: **NONE**. E2E tests must exercise the full stack:
- Real compiled binaries running in containers
- Real network communication (localhost ports or Docker network)
- Real databases, caches, message queues
- Real client interactions (HTTP requests, gRPC calls, WebRTC streams)

**Required Coverage**: All primary user journeys. Every user story in the spec must have at least one corresponding E2E test.

**CI Integration**:
```yaml
e2e-tests:
  name: E2E Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Build all services
      run: docker compose -f tests/e2e/docker-compose.yml build
    - name: Run E2E test suite
      run: go test -count=1 -v -timeout 30m ./tests/e2e/...
    - name: Collect logs on failure
      if: failure()
      run: docker compose -f tests/e2e/docker-compose.yml logs > e2e-logs.txt
    - name: Upload E2E artifacts
      if: failure()
      uses: actions/upload-artifact@v4
      with:
        name: e2e-failure-artifacts
        path: |
          e2e-logs.txt
          tests/e2e/screenshots/
```

**Example Test Pattern** (Host Discovery E2E):
```go
func TestEndToEnd_HostDiscoveryAndConnect(t *testing.T) {
    // Arrange: Start core backend, host agent, and client in containers
    compose := e2e.MustUp(t, "docker-compose.e2e.yml")
    defer compose.Down()

    coreURL := compose.ServiceURL("core-backend")
    hostURL := compose.ServiceURL("host-agent")

    // Act: Host agent registers with core
    hostClient := host.NewClient(hostURL)
    caps, err := hostClient.AdvertiseCapabilities(ctx, host.Capabilities{
        GPU: "RTX 4090",
        Codecs: []string{"HEVC", "AV1"},
        Games: []string{"Elden Ring", "Cyberpunk 2077"},
    })
    require.NoError(t, err)
    assert.NotEmpty(t, caps.HostID)

    // Act: Client queries available hosts
    client := helixplay.NewClient(coreURL)
    hosts, err := client.DiscoverHosts(ctx)
    require.NoError(t, err)
    assert.Len(t, hosts, 1)
    assert.Equal(t, "RTX 4090", hosts[0].GPU)

    // Act: Client requests connection to host
    session, err := client.ConnectToHost(ctx, hosts[0].HostID)
    require.NoError(t, err)
    assert.NotEmpty(t, session.SessionID)
    assert.Equal(t, "ready", session.Status)

    // Negative leg: Connect to non-existent host
    _, err = client.ConnectToHost(ctx, "non-existent-host-id")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}
```

**Anti-Bluff Verification Method**:
- Stop one critical service container mid-test and verify the E2E test detects the failure.
- Verify the test exercises real network paths (not in-process calls).
- HelixQA visual assertion for UI-facing E2E flows.

**Directory Structure**:
```
<module>/
  tests/e2e/
    docker-compose.e2e.yml      # Full topology definition
    <journey>_test.go            # e.g., host_discovery_test.go
    <journey>_test.go            # e.g., game_stream_test.go
    fixtures/                    # Test data, configs, game stubs
```

### 2.4 Security Tests

**Purpose**: Identify vulnerabilities through automated security testing including fuzzing, static application security testing (SAST), and dynamic application security testing (DAST).

**What Mocks Are Permitted**: **NONE**. Security tests must target the real application to find real vulnerabilities.

**Required Coverage**: All attack surfaces:
- HTTP/gRPC endpoints (injection, authentication bypass, authorization bypass)
- Input validation (malformed JSON, oversized payloads, path traversal)
- Cryptographic operations (weak algorithms, key exposure)
- Container configurations (privileged mode, exposed secrets)
- Dependency vulnerabilities (known CVEs in transitive dependencies)

**CI Integration**:
```yaml
security-tests:
  name: Security Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run gofuzz targets
      run: go test -fuzz=FuzzAuth -fuzztime=60s ./tests/security/...
    - name: Run gofuzz targets (all)
      run: go test -fuzz=. -fuzztime=30s ./...
    - name: SAST — Semgrep
      uses: returntocorp/semgrep-action@v1
      with:
        config: >-
          p/security-audit
          p/owasp-top-ten
          p/cwe-top-25
          p/gosec
    - name: SAST — govulncheck
      run: govulncheck ./...
    - name: Container scan — Trivy
      run: trivy fs --exit-code 1 --severity HIGH,CRITICAL .
    - name: Secret scan — gitleaks
      run: gitleaks detect --source . --verbose
```

**Example Test Pattern**:
```go
func FuzzAuth_ValidateToken(f *testing.F) {
    // Seed corpus with valid and invalid tokens
    f.Add("eyJhbGciOiJIUzI1NiIs...")  // valid
    f.Add("invalid.token.here")
    f.Add("")
    f.Add("../../../etc/passwd")

    f.Fuzz(func(t *testing.T, token string) {
        // This should NEVER panic regardless of input
        claims, err := auth.ValidateToken(ctx, token)
        // We only care that it doesn't panic; error is expected for fuzz inputs
        _ = claims
        _ = err
    })
}

func TestSecurity_SQLInjection(t *testing.T) {
    db := testdb.New(t)
    defer db.Close()

    // Attempt SQL injection in user input
    maliciousInputs := []string{
        "'; DROP TABLE users; --",
        "1 OR 1=1",
        "' UNION SELECT * FROM passwords --",
    }

    for _, input := range maliciousInputs {
        // This should return an error or safe result, NEVER execute the injected SQL
        _, err := userRepo.FindByUsername(ctx, input)
        // Expect error or empty result, not a data breach
        assert.True(t, err != nil || len(users) == 0, "SQL injection possible with: %s", input)
    }
}
```

**Anti-Bluff Verification Method**:
- Introduce a known vulnerability (e.g., remove input sanitization) and verify the security test detects it.
- Verify fuzzing actually finds crashes: check that the fuzz corpus grows over time.

**Directory Structure**:
```
<module>/
  tests/security/
    fuzz_<target>_test.go       # Fuzzing targets
    <attack>_test.go            # Specific attack vectors
    owasp_top10_test.go         # OWASP Top 10 coverage
```

### 2.5 Benchmarking

**Purpose**: Measure and track performance characteristics of critical code paths. Benchmarks detect performance regressions and validate latency SLAs.

**What Mocks Are Permitted**: **NONE**. Benchmarks must measure real code with real data to produce meaningful results.

**Required Coverage**: All performance-critical paths:
- Stream encoding pipeline (frame encode latency)
- Database query paths (query execution time)
- Memory allocation hot paths (allocs/op)
- Network serialization/deserialization
- JWT signing/validation
- Game discovery and enumeration
- Storage upload/download throughput

**SLA Targets**:
| Metric | p50 Target | p99 Target | p999 Target |
|--------|-----------|-----------|-------------|
| Frame encode latency | 8ms | 16ms | 33ms |
| End-to-end stream latency (LAN) | 15ms | 30ms | 50ms |
| End-to-end stream latency (WAN) | 25ms | 50ms | 100ms |
| Game discovery response | 50ms | 200ms | 500ms |
| Auth token validation | 1ms | 5ms | 10ms |
| Storage upload (1MB chunk) | 100ms | 500ms | 1000ms |

**CI Integration**:
```yaml
benchmarks:
  name: Benchmarks
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run benchmarks
      run: go test -bench=. -benchmem -benchtime=5s ./tests/benchmark/... | tee benchmark.txt
    - name: Compare with baseline
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'go'
        output-file-path: benchmark.txt
        github-token: ${{ secrets.GITHUB_TOKEN }}
        alert-threshold: '150%'  # Fail if 50% slower than baseline
        comment-on-alert: true
        fail-on-alert: true
    - name: Upload benchmark results
      uses: actions/upload-artifact@v4
      with:
        name: benchmark-results
        path: benchmark.txt
```

**Example Test Pattern**:
```go
func BenchmarkEncoder_EncodeFrame(b *testing.B) {
    enc := encoder.New(encoder.Config{Codec: "HEVC", Bitrate: 20_000_000})
    frame := generateTestFrame(1920, 1080, pixel.RGBA)

    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        packet, err := enc.EncodeFrame(frame)
        if err != nil {
            b.Fatal(err)
        }
        if len(packet.Data) == 0 {
            b.Fatal("empty packet")
        }
    }
}

func BenchmarkAuth_ValidateToken(b *testing.B) {
    auth := auth.NewManager(db, auth.WithJWTSecret("benchmark-secret"))
    token, _ := auth.GenerateToken(ctx, "benchmark-user")

    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        _, err := auth.ValidateToken(ctx, token)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Anti-Bluff Verification Method**:
- Introduce an artificial slowdown (e.g., add `time.Sleep(10 * time.Millisecond)`) and verify the benchmark detects the regression.
- Verify benchmark results are actually stored and compared (not just generated and discarded).

**Directory Structure**:
```
<module>/
  tests/benchmark/
    <component>_bench_test.go    # e.g., encoder_bench_test.go
    benchmark_baseline.txt       # Committed baseline for comparison
```

### 2.6 Chaos Tests

**Purpose**: Verify system resilience by injecting faults into a running production-like deployment. Chaos tests prove the system degrades gracefully under failure conditions.

**What Mocks Are Permitted**: Only the chaos injection mechanisms themselves (e.g., the tool that kills a container is a test fixture; the container being killed is a real service).

**Required Coverage**: All failure modes:
- Container/pod crashes (kill random service)
- Network partitions (isolate service from its dependencies)
- Latency injection (add 500ms to all DB queries)
- Packet loss (drop 10% of UDP stream packets)
- Resource exhaustion (CPU throttling, memory pressure)
- DNS failure (unresolvable service names)
- Certificate expiry (invalid TLS certificates)

**CI Integration**:
```yaml
chaos-tests:
  name: Chaos Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Start full topology
      run: docker compose -f tests/chaos/docker-compose.yml up -d
    - name: Run chaos test suite
      run: go test -count=1 -v -timeout 60m ./tests/chaos/...
      env:
        CHAOS_DURATION: 5m
        CHAOS_INTERVAL: 30s
    - name: Collect chaos experiment results
      if: always()
      run: |
        docker compose -f tests/chaos/docker-compose.yml logs > chaos-logs.txt
        kubectl describe chaosresults > chaos-results.txt
    - name: Upload artifacts
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: chaos-artifacts
        path: |
          chaos-logs.txt
          chaos-results.txt
```

**Example Test Pattern**:
```go
func TestChaos_HostAgentCrash_StreamContinues(t *testing.T) {
    // Arrange: Start full streaming session
    topo := chaos.MustStartTopology(t, "streaming-topology.yml")
    defer topo.Teardown()

    client := topo.Client()
    hostID := topo.HostID(0)

    // Start streaming
    session, err := client.ConnectToHost(ctx, hostID)
    require.NoError(t, err)

    stream, err := client.StartStream(ctx, session.SessionID)
    require.NoError(t, err)

    // Verify stream is delivering frames
    frameCount := countFrames(stream, 5*time.Second)
    require.Greater(t, frameCount, 0, "stream should deliver frames")

    // Act: Kill the host agent container
    topo.KillContainer("host-agent")

    // Assert: Stream should detect disconnection within timeout
    err = waitForStreamError(stream, 10*time.Second)
    assert.NoError(t, err, "stream should detect host disconnection")

    // Assert: Client should be able to reconnect to another host
    topo.StartContainer("host-agent")  // Restart
    hostID2 := topo.HostID(1)
    session2, err := client.ConnectToHost(ctx, hostID2)
    assert.NoError(t, err, "should connect to alternate host")
    assert.NotNil(t, session2)
}
```

**Anti-Bluff Verification Method**:
- Run the chaos test without injecting chaos and verify it fails (or is skipped) — a chaos test that passes without chaos being injected is a bluff.
- Verify the fault injection actually occurred by checking container restart counts / network drop logs.

**Directory Structure**:
```
<module>/
  tests/chaos/
    <failure_mode>_test.go       # e.g., container_crash_test.go
    <failure_mode>_test.go       # e.g., network_partition_test.go
    docker-compose.chaos.yml     # Chaos topology
    chaos-experiments/           # Litmus/chaos-mesh experiment definitions
```

### 2.7 Stress Tests

**Purpose**: Verify system behavior under sustained load beyond normal capacity. Stress tests find resource leaks, deadlock conditions, and degradation patterns that only appear over time.

**What Mocks Are Permitted**: **NONE**. Stress tests must exercise real services under real load.

**Required Coverage**:
- 24-hour sustained load profile (memory leak detection)
- Connection saturation (10,000 concurrent connections)
- Request flood (10x normal request rate for 1 hour)
- Storage exhaustion (fill disk to 95%)
- Memory pressure (reduce available RAM by 50%)
- Database connection pool exhaustion
- Goroutine leak detection (goroutine count must not grow unboundedly)

**CI Integration**:
```yaml
stress-tests:
  name: Stress Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Build stress test services
      run: docker compose -f tests/stress/docker-compose.yml build
    - name: Run 4-hour stress profile
      run: go test -count=1 -v -timeout 240m ./tests/stress/...
      env:
        STRESS_DURATION: 4h
        STRESS_CONCURRENCY: 1000
    - name: Collect resource usage data
      if: always()
      run: |
        cat tests/stress/resource-usage.log > stress-resources.txt
        go tool pprof -top tests/stress/profile.pb.gz > stress-pprof.txt
    - name: Upload artifacts
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: stress-artifacts
        path: |
          stress-resources.txt
          stress-pprof.txt
```

**Example Test Pattern**:
```go
func TestStress_24HourStreaming(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping 24-hour stress test in short mode")
    }

    topo := stress.MustStartTopology(t, "streaming-topology.yml")
    defer topo.Teardown()

    client := topo.Client()
    hostID := topo.HostID(0)

    // Start streaming session
    session, err := client.ConnectToHost(ctx, hostID)
    require.NoError(t, err)

    stream, err := client.StartStream(ctx, session.SessionID)
    require.NoError(t, err)

    // Collect baseline metrics
    baselineGoroutines := runtime.NumGoroutine()
    baselineMem := getMemUsage()

    // Run stream for 24 hours, collecting frames
    duration := 24 * time.Hour
    if os.Getenv("CI") == "true" {
        duration = 4 * time.Hour  // Shorter in CI
    }

    ctx, cancel := context.WithTimeout(ctx, duration)
    defer cancel()

    framesReceived := int64(0)
    go func() {
        for {
            _, err := stream.RecvFrame()
            if err != nil {
                return
            }
            atomic.AddInt64(&framesReceived, 1)
        }
    }()

    <-ctx.Done()

    // Assert: Should have received frames throughout
    assert.Greater(t, atomic.LoadInt64(&framesReceived), int64(0), "should have received frames")

    // Assert: No goroutine leak
    finalGoroutines := runtime.NumGoroutine()
    assert.LessOrEqual(t, finalGoroutines, baselineGoroutines+10,
        "goroutine leak detected: baseline=%d, final=%d", baselineGoroutines, finalGoroutines)

    // Assert: Memory growth bounded
    finalMem := getMemUsage()
    memGrowth := float64(finalMem-baselineMem) / float64(baselineMem) * 100
    assert.LessOrEqual(t, memGrowth, 50.0,
        "memory grew by %.1f%% over %v", memGrowth, duration)
}
```

**Anti-Bluff Verification Method**:
- Verify the test actually ran for the full duration (check timestamps in logs).
- Introduce a goroutine leak (intentionally forget to close a goroutine) and verify the stress test detects it.

**Directory Structure**:
```
<module>/
  tests/stress/
    <scenario>_stress_test.go    # e.g., streaming_stress_test.go
    docker-compose.stress.yml    # Stress topology
    profiles/                    # Load profiles (normal, peak, extreme)
```

### 2.8 Smoke Tests

**Purpose**: Provide rapid post-deployment validation that the system is minimally functional. Smoke tests run in under 30 seconds and cover the most critical paths.

**What Mocks Are Permitted**: **NONE**. Smoke tests verify real deployed services.

**Required Coverage**: All critical endpoints:
- Health check endpoints return 200 OK
- Core backend responds to discovery queries
- Host agent beacon is active
- Database connectivity
- Auth service token generation
- Storage service read/write
- All container statuses are healthy

**CI Integration**:
```yaml
smoke-tests:
  name: Smoke Tests
  runs-on: ubuntu-latest
  needs: [deploy-staging]
  steps:
    - uses: actions/checkout@v4
    - name: Run smoke tests against staging
      run: go test -count=1 -v -timeout 60s ./tests/smoke/...
      env:
        HELIX_BASE_URL: https://staging.helixplay.dev
    - name: Notify on failure
      if: failure()
      uses: slack-action/notify@v1
      with:
        message: "SMOKE TEST FAILED on staging — deployment may be broken"
```

**Example Test Pattern**:
```go
func TestSmoke_HealthChecks(t *testing.T) {
    baseURL := os.Getenv("HELIX_BASE_URL")
    require.NotEmpty(t, baseURL, "HELIX_BASE_URL required")

    services := []struct {
        name   string
        path   string
        expect int
    }{
        {"core-backend", "/health", 200},
        {"host-agent", "/health", 200},
        {"auth-service", "/health", 200},
        {"storage-service", "/health", 200},
    }

    for _, svc := range services {
        t.Run(svc.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
            defer cancel()

            req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+svc.path, nil)
            resp, err := http.DefaultClient.Do(req)
            require.NoError(t, err, "%s is unreachable", svc.name)
            defer resp.Body.Close()

            assert.Equal(t, svc.expect, resp.StatusCode, "%s health check failed", svc.name)
        })
    }
}
```

**Anti-Bluff Verification Method**:
- Deploy a broken version (e.g., service that returns 500 on health check) and verify smoke tests fail.
- Measure actual execution time — smoke tests that take > 30 seconds indicate a problem.

**Directory Structure**:
```
<module>/
  tests/smoke/
    smoke_test.go               # All smoke tests (fast, consolidated)
```

### 2.9 Full Automation Tests

**Purpose**: Verify that the entire system can be deployed, configured, and operated without human intervention. Full automation tests are the ultimate integration test — they test the system as a whole, including deployment automation.

**What Mocks Are Permitted**: **NONE**. Full automation tests must use real infrastructure (even if that infrastructure is Docker Compose on CI runners).

**Required Coverage**:
- Full deployment from clean state completes without manual steps
- All services start and reach healthy state
- All 10 test types can be triggered and complete
- Configuration is automatically applied
- Monitoring and alerting are functional
- Rollback procedure works
- Backup and restore procedures work

**CI Integration**:
```yaml
full-automation:
  name: Full Automation
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Full automation test
      run: go test -count=1 -v -timeout 120m ./tests/fullauto/...
      env:
        FULLAUTO_MODE: ci
    - name: Upload full automation report
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: fullauto-report
        path: tests/fullauto/report/
```

**Example Test Pattern**:
```go
func TestFullAutomation_DeployAndOperate(t *testing.T) {
    // Arrange: Clean environment
    ctx := context.Background()
    deployer := fullauto.NewDeployer()

    // Act: Full deployment from clean state
    deployment, err := deployer.Deploy(ctx, fullauto.Config{
        Topology:   "minimal-3-node",
        Version:    "latest",
        AutoConfig: true,
    })
    require.NoError(t, err, "deployment should complete without human intervention")
    defer deployment.Teardown()

    // Assert: All services healthy
    services := deployment.ListServices()
    assert.GreaterOrEqual(t, len(services), 5, "should have at least 5 services")
    for _, svc := range services {
        assert.Equal(t, "healthy", svc.Status, "service %s should be healthy", svc.Name)
    }

    // Act: Trigger all 10 test types
    results := deployment.RunAllTests(ctx)
    assert.Equal(t, 10, results.TotalTypes, "all 10 test types should be executable")
    assert.Equal(t, 10, results.PassedTypes, "all 10 test types should pass")

    // Act: Simulate service failure and verify auto-recovery
    deployment.KillService("host-agent")
    err = deployment.WaitForRecovery(ctx, "host-agent", 2*time.Minute)
    assert.NoError(t, err, "host-agent should auto-recover")

    // Act: Backup and restore
    backup, err := deployment.Backup(ctx)
    require.NoError(t, err)

    deployment.Teardown()
    restored, err := deployer.Restore(ctx, backup)
    require.NoError(t, err, "restore should complete without human intervention")
    defer restored.Teardown()

    // Assert: Restored system functional
    restoredServices := restored.ListServices()
    assert.GreaterOrEqual(t, len(restoredServices), 5, "restored system should have all services")
}
```

**Anti-Bluff Verification Method**:
- Run the full automation test with a manual step required (e.g., a prompt that needs human input) and verify it times out and fails.
- Verify the deployment is truly clean (no pre-existing state, containers, or volumes).

**Directory Structure**:
```
<module>/
  tests/fullauto/
    deploy_test.go              # Deployment automation
    operate_test.go             # Operational procedures
    disaster_recovery_test.go   # Backup/restore
```

### 2.10 Challenges

**Purpose**: Execute structured, production-equivalent test scenarios defined in the Challenges repository. Challenges are the highest-fidelity test type — they verify complete feature behavior in an environment that mirrors production.

**What Mocks Are Permitted**: **NONE**. Challenges use real everything.

**Required Coverage**: All challenge bank scenarios:
- Game-specific challenges (each supported game)
- Full QA challenges (Android, Android TV, Desktop, Web)
- Navigation challenges (all UI flows)
- Recording challenges (stream capture, storage, playback)
- Admin challenges (user management, system configuration)
- Security challenges (auth, rate limiting, DDoS)
- Performance challenges (latency, throughput under load)
- Regression challenges (previously fixed bugs must stay fixed)

**CI Integration**:
```yaml
challenges:
  name: Challenges
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Build challenge runner
      run: cd vasic-digital/Challenges && go build -o challenge-runner ./cmd/userflow-runner
    - name: Run all challenges
      run: ./vasic-digital/Challenges/challenge-runner --all --report challenges-report.json
    - name: Upload challenge report
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: challenges-report
        path: challenges-report.json
    - name: Anti-bluff validation
      run: |
        if ! jq '.antiBluffValid' challenges-report.json | grep -q true; then
          echo "ANTI-BLUFF VALIDATION FAILED"
          exit 1
        fi
```

**Example Challenge Definition**:
```yaml
# challenges/banks/game-streaming.yaml
challenge:
  name: "HelixPlay Game Stream — Elden Ring"
  category: "game-streaming"
  priority: 1
  steps:
    - name: "Discover host"
      action: "api_call"
      endpoint: "/api/v1/discover"
      assert:
        - status: 200
        - jsonpath: "$.hosts[0].gpu" equals "RTX 4090"
    - name: "Connect to host"
      action: "api_call"
      endpoint: "/api/v1/connect"
      body: '{"host_id": "${hosts[0].id}"}'
      assert:
        - status: 201
        - jsonpath: "$.status" equals "ready"
    - name: "Launch game"
      action: "api_call"
      endpoint: "/api/v1/games/launch"
      body: '{"game": "Elden Ring", "session_id": "${session.id}"}'
      assert:
        - status: 200
        - jsonpath: "$.state" equals "running"
    - name: "Verify stream active"
      action: "verify_stream"
      timeout: 30s
      assert:
        - fps greater_than 30
        - latency_ms less_than 50
        - resolution equals "1920x1080"
  antiBluff:
    recordedActions: ["api_call", "verify_stream"]
    assertions: ["status", "jsonpath", "fps", "latency_ms", "resolution"]
```

**Anti-Bluff Verification Method**:
- `ValidateAntiBluff()` is called unconditionally on every challenge result.
- Challenge scripts (Section 8) verify compilation, functionality, and unit test coverage.
- The bluff scanner runs against challenge definitions to detect vacuous assertions.

**Directory Structure**:
```
vasic-digital/Challenges/
  banks/                        # Challenge bank definitions
    game-streaming.yaml
    full-qa-android.yaml
    full-qa-androidtv.yaml
    navigation.yaml
    recording.yaml
    admin.yaml
    security.yaml
    performance.yaml
    regression.yaml
  pkg/challenge/                # Challenge framework
  pkg/runner/                   # Execution engine
  scripts/                      # Challenge scripts (Section 8)
```

---

## Section 3: Test Matrix (29 Submodules x 10 Types)

### 3.1 Submodules Overview

The HelixPlay platform consists of **29 submodules** organized into three groups:

| Group | Count | Submodules |
|-------|-------|-----------|
| Core Infrastructure | 19 | Auth, Cache, Database, Discovery, EventBus, Formatters, Media, Memory, Messaging, Middleware, Observability, Plugins, RAG, RateLimiter, Recovery, Security, Storage, Streaming, VectorDB |
| Challenges | 1 | vasic-digital/Challenges |
| HelixQA | 1 | HelixDevelopment/HelixQA |
| Application Modules | 8 | client-wails, client-web, core-backend, host-agent/capability, host-agent/capture, host-agent/codec, host-agent/discovery, host-agent/encoder, host-agent/game, host-agent/input, host-agent/lifecycle, host-agent/transport |

*Note: The host-agent submodules are treated as a single unit for some test types but individually for Unit and Integration tests.*

### 3.2 Complete Test Matrix

| # | Submodule | Unit | Integ | E2E | Sec | Bench | Chaos | Stress | Smoke | FullAuto | Chall | Owner | Status |
|---|-----------|:----:|:-----:|:---:|:---:|:-----:|:-----:|:------:|:-----:|:--------:|:-----:|-------|--------|
| 1 | **Auth** | [x] | [x] | [x] | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Auth | Partial |
| 2 | **Cache** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Cache | Partial |
| 3 | **Database** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Database | Partial |
| 4 | **Discovery** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Discovery | Partial |
| 5 | **EventBus** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | EventBus | Partial |
| 6 | **Formatters** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Formatters | Minimal |
| 7 | **Media** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Media | Partial |
| 8 | **Memory** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Memory | Minimal |
| 9 | **Messaging** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Messaging | Partial |
| 10 | **Middleware** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Middleware | Partial |
| 11 | **Observability** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Observability | Partial |
| 12 | **Plugins** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Plugins | Minimal |
| 13 | **RAG** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | RAG | Partial |
| 14 | **RateLimiter** | [x] | [x] | [x] | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | RateLimiter | Partial |
| 15 | **Recovery** | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Recovery | Partial |
| 16 | **Security** | [x] | [x] | [x] | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Security | Partial |
| 17 | **Storage** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Storage | **Good** |
| 18 | **Streaming** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Streaming | Partial |
| 19 | **VectorDB** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | VectorDB | Minimal |
| 20 | **Challenges** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Challenges | **Good** |
| 21 | **HelixQA** | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [x] | HelixQA | Partial |
| 22 | **client-wails** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Client | Minimal |
| 23 | **client-web** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Client | Minimal |
| 24 | **core-backend** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Core | Minimal |
| 25 | **host-agent/capability** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [x] | Host | Partial |
| 26 | **host-agent/capture** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Host | Partial |
| 27 | **host-agent/codec** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Host | Partial |
| 28 | **host-agent/discovery** | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [x] | Host | Partial |
| 29 | **host-agent/encoder** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Host | Partial |
| | **TOTAL** | **29** | **18** | **8** | **4** | **18** | **0** | **0** | **0** | **0** | **11** | | |

### 3.3 Priority for Missing Tests

**Priority 1 — Immediate (Sprint 1-2)**:
| Submodule | Test Type | Why Critical | Effort |
|-----------|-----------|-------------|--------|
| All 29 | Full Automation | No full automation exists; this is the most critical gap | 2 weeks |
| All 29 | Smoke | No smoke tests exist; every deploy is blind | 1 week |
| Streaming, Storage, Auth | Chaos | Most critical services need resilience validation | 2 weeks |
| Streaming, Storage, Auth | Stress | Verify no resource leaks under sustained load | 2 weeks |

**Priority 2 — High (Sprint 3-4)**:
| Submodule | Test Type | Why High | Effort |
|-----------|-----------|----------|--------|
| Cache, EventBus, Messaging | E2E | Core infrastructure needs E2E coverage | 1 week |
| All host-agent/* | E2E | Host agent is the core product; needs E2E | 2 weeks |
| client-wails, client-web | All types | Client has minimal test coverage | 3 weeks |
| Formatters, Memory, Plugins, VectorDB | Integration | No integration tests at all | 1 week |

**Priority 3 — Medium (Sprint 5-6)**:
| Submodule | Test Type | Why Medium | Effort |
|-----------|-----------|------------|--------|
| RAG, Observability | E2E | Important but not critical path | 1 week |
| Recovery | E2E, Chaos | Fault recovery needs chaos testing | 1 week |
| core-backend | All types | Core backend is stub; tests depend on implementation | 2 weeks |
| All | Security (fuzz) | Fuzzing coverage expansion | 2 weeks |

---


## Section 4: Anti-Bluff Infrastructure Implementation

### 4.1 ValidateAntiBluff() Validator

**Location**: `vasic-digital/Challenges/pkg/challenge/antibluff.go`  
**Purpose**: Unconditionally validate that a challenge result claiming `Status=Passed` actually has evidence of execution.

**Constitutional Basis (v2.1.0)**:
> *"`ValidateAntiBluff` gate is unconditional; `CHALLENGE_ANTIBLUFF_STRICT` removed."*

#### 4.1.1 Implementation Details

The `ValidateAntiBluff()` function is called on every `Result` object, regardless of test type or challenge category. It enforces three rules:

```go
// pkg/challenge/antibluff.go
package challenge

import "fmt"

var ErrBluffPass = fmt.Errorf("antibluff: result claims Pass but lacks execution evidence")

type Result struct {
    Status         Status      // Passed, Failed, Skipped, Error
    RecordedActions []Action   // What the runtime actually did
    Assertions     []Assertion // What was checked
    // ... other fields
}

type Action struct {
    Name      string
    Timestamp int64
    Details   map[string]interface{}
}

type Assertion struct {
    Name   string
    Passed bool
    Details map[string]interface{}
}

// ValidateAntiBluff enforces three rules on any Result with Status=Passed.
// This function is called UNCONDITIONALLY — no toggle, no bypass.
func ValidateAntiBluff(r *Result) error {
    // Rule 1: Non-Pass statuses are honest by definition
    if r.Status != StatusPassed {
        return nil // Failed, Skipped, Error — these can't be bluff passes
    }

    // Rule 2: RecordedActions must be non-empty
    // Proof the runtime actually executed something
    if len(r.RecordedActions) == 0 {
        return fmt.Errorf("%w: zero recorded actions for Passed result", ErrBluffPass)
    }

    // Rule 3: Assertions must be non-empty
    // At least one expectation was checked
    if len(r.Assertions) == 0 {
        return fmt.Errorf("%w: no assertions recorded for Passed result", ErrBluffPass)
    }

    // Rule 4: At least one assertion must have Passed=true
    // Something was positively confirmed
    hasPassingAssertion := false
    for _, a := range r.Assertions {
        if a.Passed {
            hasPassingAssertion = true
            break
        }
    }
    if !hasPassingAssertion {
        return fmt.Errorf("%w: all assertions failed but status is Passed", ErrBluffPass)
    }

    return nil
}
```

#### 4.1.2 Three Enforcement Rules

| Rule | Name | Purpose | Failure Mode |
|------|------|---------|-------------|
| **Rule 1** | Non-Pass Honesty | `Failed`, `Skipped`, `Error` statuses are not validated (they can't be bluff passes) | N/A |
| **Rule 2** | Action Evidence | `RecordedActions` must contain at least one entry proving the runtime executed code | Result claims Pass but did nothing |
| **Rule 3** | Assertion Evidence | `Assertions` must contain at least one entry, and at least one must have `Passed=true` | Result claims Pass but no assertions passed |

#### 4.1.3 Integration into Challenge Runner

The validator is integrated at three points in the challenge execution pipeline:

```go
// pkg/runner/runner.go — ExecuteChallenge
func (r *Runner) ExecuteChallenge(ctx context.Context, ch challenge.Challenge) (*challenge.Result, error) {
    // Execute the challenge
    result, err := ch.Execute(ctx)
    if err != nil {
        return nil, err
    }

    // UNCONDITIONAL anti-bluff validation
    if err := challenge.ValidateAntiBluff(result); err != nil {
        // Force status to Error if anti-bluff validation fails
        result.Status = challenge.StatusError
        result.ErrorDetails = err.Error()
        return result, err
    }

    return result, nil
}
```

**Integration Points**:
1. **Challenge Runner** (`pkg/runner/`): Every executed challenge is validated before the result is returned.
2. **Report Generation** (`pkg/report/`): Reports include an `antiBluffValid` boolean field that aggregates validation across all challenges.
3. **CI Pipeline**: The `challenges.yml` workflow checks `antiBluffValid` in the report JSON and fails if any challenge lacks evidence.

#### 4.1.4 Anti-Bluff Test Suite

The validator itself is tested with 9 test cases in `pkg/challenge/antibluff_test.go`:

| Test | Scenario | Expected |
|------|----------|----------|
| `TestValidate_PassWithEvidence` | Happy path: Pass with actions and passing assertion | `nil` (valid) |
| `TestValidate_PassWithZeroActions` | THE bluff pattern: claims Pass but did nothing | `ErrBluffPass` |
| `TestValidate_PassWithEmptyAssertions` | Metadata-only pattern: actions recorded but nothing checked | `ErrBluffPass` |
| `TestValidate_PassWithAllAssertionsFailing` | Most insidious bluff: all assertions failed but status=Pass | `ErrBluffPass` |
| `TestValidate_PassWithMixedAssertions` | Mixed pass/fail is OK if at least one passes | `nil` (valid) |
| `TestValidate_StatusFailedHonest` | Non-Pass statuses are honest by definition | `nil` (valid) |
| `TestValidate_StatusSkipped` | Skipped status doesn't need validation | `nil` (valid) |
| `TestRecordAction` | Action recording works correctly | Action recorded |
| `TestRecordAction_NilReceiver` | Defensive check against nil panic | No panic |

### 4.2 Anti-Bluff Scan Script

**Location**: `HelixPlay/scripts/anti-bluff-scan.sh`  
**Purpose**: Five-step comprehensive scan of the entire source tree for bluff patterns, forbidden code, and constitutional compliance.

#### 4.2.1 Five-Step Scan Process

```bash
#!/bin/bash
# scripts/anti-bluff-scan.sh

set -euo pipefail

REPORT_DIR=".anti-bluff-reports"
mkdir -p "$REPORT_DIR"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
REPORT_FILE="$REPORT_DIR/scan-${TIMESTAMP}.json"

STEP=0
FAILURES=0

echo "=== Anti-Bluff Scan v2.1.0 ==="
echo "Started: $TIMESTAMP"

# ───────────────────────────────────────────
# STEP 1: Forbidden Pattern Detection
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Forbidden pattern detection..."

FORBIDDEN_PATTERNS=(
    'panic\s*\(\s*"not implemented"\s*\)'
    'return\s+.*fmt\.Errorf\s*\(\s*"not implemented"'
    'return\s+.*errors\.New\s*\(\s*"not implemented"'
    '\bTODO\b.*[^#]'          # TODO without issue reference
    '\bFIXME\b.*[^#]'         # FIXME without issue reference
    '\bXXX\b'                 # XXX marker
    '\bHACK\b'                # HACK marker
    '\btbd\b'                 # standalone TBD
    '\{\s*\}'                 # empty function body
)

STEP1_ISSUES=()
for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.go" --include="*.md" . || true)
    if [[ -n "$matches" ]]; then
        STEP1_ISSUES+=("$matches")
    fi
done

STEP1_COUNT=${#STEP1_ISSUES[@]}
if [[ $STEP1_COUNT -gt 0 ]]; then
    echo "  FAIL: Found $STEP1_COUNT forbidden pattern(s)"
    printf '%s\n' "${STEP1_ISSUES[@]}"
    ((FAILURES++))
else
    echo "  PASS: No forbidden patterns found"
fi

# ───────────────────────────────────────────
# STEP 2: ValidateAntiBluff Verification
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] ValidateAntiBluff() verification..."

# Check that ValidateAntiBluff is called in all challenge/runner code
VALIDATE_CALLS=$(grep -r "ValidateAntiBluff" --include="*.go" . | wc -l)
if [[ $VALIDATE_CALLS -eq 0 ]]; then
    echo "  FAIL: No calls to ValidateAntiBluff found"
    ((FAILURES++))
else
    echo "  PASS: Found $VALIDATE_CALLS call(s) to ValidateAntiBluff"
fi

# Verify unconditional call (no CHALLENGE_ANTIBLUFF_STRICT toggle)
if grep -r "CHALLENGE_ANTIBLUFF_STRICT" --include="*.go" . > /dev/null 2>&1; then
    echo "  FAIL: CHALLENGE_ANTIBLUFF_STRICT toggle still present (must be removed per v2.1.0)"
    ((FAILURES++))
else
    echo "  PASS: No anti-bluff bypass toggles found"
fi

# ───────────────────────────────────────────
# STEP 3: Documentation Anti-Bluff Blocks
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Documentation anti-bluff verification..."

# Check that documentation files contain Anti-Bluff Verification sections
DOC_FILES=$(find . -name "*.md" -not -path "*/vendor/*" -not -path "*/.git/*")
MISSING_BLOCKS=0
for doc in $DOC_FILES; do
    if ! grep -q "Anti-Bluff Verification\|Anti Bluff\|antibluff" "$doc" 2>/dev/null; then
        # Only flag docs in submodule directories, not root-level docs
        if [[ "$doc" == *"/pkg/"* ]] || [[ "$doc" == *"submodule"* ]]; then
            MISSING_BLOCKS=$((MISSING_BLOCKS + 1))
        fi
    fi
done

if [[ $MISSING_BLOCKS -gt 0 ]]; then
    echo "  WARN: $MISSING_BLOCKS submodule doc(s) missing anti-bluff verification section"
else
    echo "  PASS: All submodule docs have anti-bluff verification sections"
fi

# ───────────────────────────────────────────
# STEP 4: Constitution Propagation
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Constitution propagation check..."

# Verify all submodules reference the Constitution
SUBMODULES=$(git submodule status | awk '{print $2}')
MISSING_CONSTITUTION=0
for submod in $SUBMODULES; do
    if [[ -d "$submod" ]]; then
        if ! grep -r "Constitution\|constitution" "$submod" --include="*.md" --include="*.go" > /dev/null 2>&1; then
            MISSING_CONSTITUTION=$((MISSING_CONSTITUTION + 1))
            echo "  WARN: $submod does not reference Constitution"
        fi
    fi
done

if [[ $MISSING_CONSTITUTION -eq 0 ]]; then
    echo "  PASS: All submodules reference Constitution"
fi

# ───────────────────────────────────────────
# STEP 5: Vacuous Assertion Detection
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Vacuous assertion detection..."

VACUOUS_PATTERNS=(
    'assert\.True\s*\(\s*t\s*,\s*true\s*\)'
    'assert\.Equal\s*\(\s*t\s*,\s*true\s*,\s*true\s*\)'
    'assert\.Nil\s*\(\s*t\s*,\s*nil\s*\)'
    'assert\.Equal\s*\(\s*t\s*,\s*[^,]+\s*,\s*\2\s*\)'  # Same var compared to itself
    'assert\.NoError\s*\(\s*t\s*,\s*nil\s*\)'
)

STEP5_ISSUES=()
for pattern in "${VACUOUS_PATTERNS[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*_test.go" . || true)
    if [[ -n "$matches" ]]; then
        STEP5_ISSUES+=("$matches")
    fi
done

STEP5_COUNT=${#STEP5_ISSUES[@]}
if [[ $STEP5_COUNT -gt 0 ]]; then
    echo "  FAIL: Found $STEP5_COUNT vacuous assertion(s)"
    printf '%s\n' "${STEP5_ISSUES[@]}"
    ((FAILURES++))
else
    echo "  PASS: No vacuous assertions found"
fi

# ───────────────────────────────────────────
# SUMMARY
# ───────────────────────────────────────────
echo ""
echo "=== Scan Complete ==="
echo "Failures: $FAILURES"
echo "Report: $REPORT_FILE"

# Generate JSON report
cat > "$REPORT_FILE" <<EOF
{
  "timestamp": "$TIMESTAMP",
  "version": "2.1.0",
  "failures": $FAILURES,
  "steps": {
    "forbidden_patterns": { "status": "$([[ ${#STEP1_ISSUES[@]} -eq 0 ]] && echo "PASS" || echo "FAIL)", "count": ${#STEP1_ISSUES[@]} },
    "validate_antibluff": { "status": "$([[ $VALIDATE_CALLS -gt 0 ]] && echo "PASS" || echo "FAIL")", "calls_found": $VALIDATE_CALLS },
    "documentation_blocks": { "status": "PASS", "missing_count": $MISSING_BLOCKS },
    "constitution_propagation": { "status": "PASS", "missing_count": $MISSING_CONSTITUTION },
    "vacuous_assertions": { "status": "$([[ $STEP5_COUNT -eq 0 ]] && echo "PASS" || echo "FAIL")", "count": $STEP5_COUNT }
  }
}
EOF

exit $FAILURES
```

#### 4.2.2 CI Integration

The anti-bluff scan runs as a **non-overridable CI lane** — it cannot be bypassed, skipped, or configured away:

```yaml
# .github/workflows/anti-bluff.yml (see Section 6.2 for full file)
anti-bluff-scan:
  name: Anti-Bluff Scan
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run anti-bluff scan
      run: bash scripts/anti-bluff-scan.sh
    - name: Upload scan report
      uses: actions/upload-artifact@v4
      with:
        name: anti-bluff-report
        path: .anti-bluff-reports/
```

**Non-Overrideable Guarantee**:
- The workflow uses `on: [pull_request, push]` triggers with no path exclusions
- Branch protection rules require this check to pass before merge
- The check cannot be skipped via `[skip ci]` or commit message patterns
- Only repository administrators can override, and overrides are logged

### 4.3 Host-Integrity Scan

**Location**: `HelixPlay/scripts/claim-check.sh`  
**Constitutional Basis**: R-18 Operational Integrity (Constitution §11.5)  
**Purpose**: Prevent execution of commands that could damage the host system or disrupt CI infrastructure.

#### 4.3.1 R-18 Forbidden Commands List

The following commands and patterns are **forbidden at the highest severity level**. Their presence in any code, script, or configuration causes immediate CI failure:

| Category | Forbidden Patterns | Severity |
|----------|-------------------|----------|
| **Power Management** | `systemctl suspend`, `pm-suspend`, `rtcwake`, `poweroff`, `shutdown`, `halt`, `reboot`, `init 0`, `telinit 0` | **BLOCKER** |
| **Disk Destruction** | `mkfs\..*\s+/dev/`, `dd\s+.*if=.*of=/dev/`, `rm\s+-rf\s+/`, `rm\s+-rf\s+\*/`, `:(){ :|:& };:` (fork bomb) | **BLOCKER** |
| **Disk Partitioning** | `fdisk\s+/dev/`, `parted\s+/dev/`, `gdisk\s+/dev/` | **BLOCKER** |
| **Network Interference** | `iptables\s+-F`, `ip\s+link\s+set.*down`, `ifconfig.*down` | **CRITICAL** |
| **User/Permission Destruction** | `userdel\s+root`, `usermod\s+-L\s+root`, `chmod\s+-R\s+000\s+/` | **BLOCKER** |
| **Container Hazards** | `--privileged` with host path mounts, `/var/run/docker.sock` mounts without justification, `hostPID: true` without justification | **CRITICAL** |
| **Credential Exposure** | `password\s*=\s*["'][^"']+["']` in non-test code, `PRIVATE KEY` without encryption | **BLOCKER** |

#### 4.3.2 Container Hazard Detection

Special rules for container configurations:

```bash
# claim-check.sh — Container hazard detection

check_container_hazards() {
    local failures=0

    # Check for privileged containers with host mounts
    grep -r "privileged: true" --include="*.yml" --include="*.yaml" . | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if grep -A5 -B5 "privileged: true" "$file" | grep -q "volumes:\|volumeMounts:"; then
            echo "BLOCKER: $file has privileged container with volume mounts"
            ((failures++))
        fi
    done

    # Check for hostPID without security context
    grep -r "hostPID: true" --include="*.yml" --include="*.yaml" . | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if ! grep -A10 "hostPID: true" "$file" | grep -q "securityContext:\|runAsNonRoot"; then
            echo "CRITICAL: $file uses hostPID without security context"
            ((failures++))
        fi
    done

    # Check for docker socket mounts
    grep -r "/var/run/docker.sock" --include="*.yml" --include="*.yaml" --include="*.json" . | while read line; do
        echo "CRITICAL: Docker socket mount found: $line"
        ((failures++))
    done

    return $failures
}
```

#### 4.3.3 Operational Integrity Enforcement

The host-integrity scan runs at two points:

1. **Pre-Commit Hook**: `.claude/settings.json` Stop hook runs `claim-check.sh` before every commit from Claude Code
2. **CI Gate**: The `anti-bluff.yml` workflow includes a host-integrity sub-lane

```yaml
  host-integrity:
    name: Host Integrity (R-18)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run claim check
        run: bash scripts/claim-check.sh
      - name: Check container configurations
        run: bash scripts/claim-check.sh --container-scan
```

### 4.4 Mutation Testing

**Tool**: `go-mutesting`  
**Configuration**: `.go-mutesting.yml`  
**Purpose**: Verify that tests actually detect code changes by introducing artificial mutations and confirming tests fail.

#### 4.4.1 go-mutesting Configuration

```yaml
# .go-mutesting.yml
---
timeout: 60s  # Per-mutant timeout

# Mutators to apply
mutators:
  - branch/case       # Mutate case statements in switches
  - branch/if         # Mutate if conditions (negate, remove)
  - expression/remove # Remove sub-expressions
  - statement/remove  # Remove statements
  - numbers/incrementer  # Increment/decrement numeric constants
  - numbers/decrementer

# Exclusions (don't mutate these)
excludes:
  - "vendor/**"
  - "**/*.pb.go"           # Protobuf generated
  - "**/*_mock*.go"        # Mock files
  - "**/tests/**"          # Test code
  - "**/scripts/**"        # Scripts (not Go code but exclude anyway)
  - "**/antibluff*.go"     # Don't mutate the anti-bluff system itself
  - "cmd/**"               # CLI entry points (thin wrappers)

# Score threshold — CI fails if below this
score_threshold: 0.85  # 85% mutation score required
```

#### 4.4.2 Mutation Ratchet Challenge

The `mutation_ratchet_challenge.sh` implements a **ratchet pattern** — mutation score can only increase, never decrease:

```bash
#!/bin/bash
# challenges/scripts/mutation_ratchet_challenge.sh

set -euo pipefail

SCORE_FILE=".mutation-score"
THRESHOLD=85

# Run mutation testing
echo "Running mutation testing..."
go-mutesting --config=.go-mutesting.yml ./... > mutation-report.txt 2>&1

# Parse mutation score from report
SCORE=$(grep "Mutation score" mutation-report.txt | sed 's/.*: \([0-9.]*\)%.*/\1/')
SCORE_INT=${SCORE%.*}

echo "Current mutation score: ${SCORE}%"

# Check absolute threshold
if [[ $SCORE_INT -lt $THRESHOLD ]]; then
    echo "FAIL: Mutation score ${SCORE}% is below threshold ${THRESHOLD}%"
    exit 1
fi

# Check ratchet (score must not decrease)
if [[ -f "$SCORE_FILE" ]]; then
    PREVIOUS=$(cat "$SCORE_FILE")
    if [[ $(echo "$SCORE < $PREVIOUS" | bc -l) -eq 1 ]]; then
        echo "FAIL: Mutation score decreased from ${PREVIOUS}% to ${SCORE}%"
        echo "This is a RATCHET VIOLATION. Fix your tests."
        exit 1
    fi
fi

# Update stored score
echo "$SCORE" > "$SCORE_FILE"
echo "PASS: Mutation score ${SCORE}% meets threshold and ratchet"
```

#### 4.4.3 CI Integration

```yaml
mutation-testing:
  name: Mutation Testing
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Install go-mutesting
      run: go install github.com/zimmski/go-mutesting@latest
    - name: Run mutation tests
      run: bash vasic-digital/Challenges/scripts/mutation_ratchet_challenge.sh
    - name: Upload mutation report
      uses: actions/upload-artifact@v4
      with:
        name: mutation-report
        path: mutation-report.txt
    - name: Update stored score
      run: |
        git add .mutation-score
        git diff --cached --quiet || git commit -m "chore: update mutation score [ci skip]"
```

### 4.5 Negative-Leg Fault Injection

**Constitutional Basis**: §1.3 and §6.3 — *"CI must break each feature and verify non-Unit tests fail"*
**Purpose**: Automatically verify that the test suite actually catches bugs by deliberately introducing defects and confirming test failure.

#### 4.5.1 Automatic Feature Breakage

The negative-leg system introduces controlled defects into the codebase:

| Mutation Type | Description | Example |
|--------------|-------------|---------|
| **Comparison Swap** | Swap comparison operators | `<` becomes `>`, `==` becomes `!=` |
| **Boundary Off-by-One** | Adjust loop bounds by 1 | `i < n` becomes `i <= n` |
| **Return Value Corruption** | Change return values | `return true` becomes `return false` |
| **Error Path Removal** | Remove error checks | `if err != nil { return err }` is deleted |
| **Constant Mutation** | Change numeric constants | `timeout := 30 * time.Second` becomes `timeout := 0` |
| **Branch Removal** | Delete if/else branches | Remove the `else` block entirely |

#### 4.5.2 Verification That Tests Fail

```go
// tests/internal/negativeleg/negative_leg_test.go
package negativeleg

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// TestNegativeLeg_AllFeatures tests that the test suite catches
// deliberate bugs in every production feature.
func TestNegativeLeg_AllFeatures(t *testing.T) {
    if os.Getenv("RUN_NEGATIVE_LEG") != "true" {
        t.Skip("Set RUN_NEGATIVE_LEG=true to run negative-leg fault injection")
    }

    // List all packages with production code
    packages := getProductionPackages(t)

    for _, pkg := range packages {
        t.Run(pkg.Name, func(t *testing.T) {
            // For each mutation point in the package
            for _, mutation := range pkg.MutationPoints {
                t.Run(mutation.Description, func(t *testing.T) {
                    // Apply the mutation
                    backup := mutation.Apply()
                    defer backup.Restore()  // Always restore, even on panic

                    // Run non-Unit tests against the mutated code
                    cmd := exec.Command("go", "test",
                        "-count=1",
                        "-run", "^(TestIntegration|TestE2E|TestChallenge)",
                        "./...",
                    )
                    output, _ := cmd.CombinedOutput()

                    // At least one test MUST fail
                    if !strings.Contains(string(output), "FAIL") {
                        t.Errorf("NEGATIVE LEG FAILED: Mutation '%s' in %s was NOT caught by any test.\n"+
                            "This means the test suite is BLUFFING — it passes even when the feature is broken.\n"+
                            "Output:\n%s", mutation.Description, pkg.Name, output)
                    }
                })
            }
        })
    }
}
```

#### 4.5.3 CI Integration

```yaml
negative-leg:
  name: Negative Leg Fault Injection
  runs-on: ubuntu-latest
  # Run weekly (expensive) and on-demand
  on:
    schedule:
      - cron: '0 2 * * 0'  # Sundays at 2 AM
    workflow_dispatch:
  steps:
    - uses: actions/checkout@v4
    - name: Run negative-leg fault injection
      run: RUN_NEGATIVE_LEG=true go test -count=1 -v -timeout 120m ./tests/internal/negativeleg/...
    - name: Upload negative-leg report
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: negative-leg-report
        path: tests/internal/negativeleg/report/
```

---

## Section 5: HelixQA Autonomous QA Integration

### 5.1 Visual Assertion Pipeline

**Repository**: `HelixDevelopment/HelixQA`  
**Test Files**: 350 `_test.go` files  
**Purpose**: Automated visual verification of UI features using screenshot capture and OpenCV-based analysis.

#### 5.1.1 Screenshot/Screen Recording Capture

The HelixQA capture system records the application under test using platform-specific capture methods:

```go
// pkg/capture/capture.go — Core capture interface
type Capture interface {
    // CaptureFrame grabs a single frame from the display
    CaptureFrame(ctx context.Context) (*Frame, error)

    // StartRecording begins continuous screen recording
    StartRecording(ctx context.Context, outputPath string) error

    // StopRecording ends the recording
    StopRecording() (*Recording, error)
}

// Frame represents a captured screen frame
type Frame struct {
    Image      image.Image
    Timestamp  time.Time
    Dimensions image.Rectangle
}
```

**Platform-Specific Implementations**:
- **Linux/X11**: Uses `x11grab` (FFmpeg) or XShm for frame capture
- **Linux/Wayland**: Uses PipeWire screencast portal
- **macOS**: Uses `ScreenCaptureKit` framework via CGO
- **Windows**: Uses DXGI Desktop Duplication API
- **Android**: Uses `MediaProjection` API
- **Android TV**: Uses same `MediaProjection` with leanback UI detection

#### 5.1.2 OpenCV-Based Verification

The vision system uses OpenCV (via Go bindings) to analyze captured frames:

```go
// pkg/vision/verifier.go — OpenCV-based visual assertion
type VisualAssertion struct {
    Name        string
    Expected    *ExpectedVisual  // What we expect to see
    Tolerance   float64         // Match tolerance (0.0-1.0)
    ROI         image.Rectangle // Region of interest (optional)
    Timeout     time.Duration
}

type ExpectedVisual struct {
    TemplatePath string        // Reference image to match
    TextContent  string        // Expected text (OCR)
    ColorProfile ColorProfile  // Expected dominant colors
    ElementCount int           // Expected number of UI elements
}

// Verify performs the visual assertion against a captured frame
func (va *VisualAssertion) Verify(frame *capture.Frame) (bool, float64, error) {
    // 1. Template matching (if template provided)
    if va.Expected.TemplatePath != "" {
        matchScore := vision.TemplateMatch(frame.Image, va.Expected.TemplatePath, va.ROI)
        if matchScore < va.Tolerance {
            return false, matchScore, fmt.Errorf("template match failed: %.2f < %.2f",
                matchScore, va.Tolerance)
        }
    }

    // 2. Text recognition (if text expected)
    if va.Expected.TextContent != "" {
        detected := vision.RecognizeText(frame.Image, va.ROI)
        if !strings.Contains(detected, va.Expected.TextContent) {
            return false, 0, fmt.Errorf("text not found: expected '%s'", va.Expected.TextContent)
        }
    }

    // 3. Color analysis (if color profile expected)
    if va.Expected.ColorProfile != nil {
        profile := vision.ExtractColorProfile(frame.Image, va.ROI)
        score := va.Expected.ColorProfile.Compare(profile)
        if score < va.Tolerance {
            return false, score, fmt.Errorf("color profile mismatch: %.2f < %.2f", score, va.Tolerance)
        }
    }

    return true, 1.0, nil
}
```

#### 5.1.3 Production-Equivalent Topology Testing

HelixQA tests run against a **production-equivalent topology** — not mocks, not stubs, but the actual compiled application running in a realistic environment:

```go
// pkg/autonomous/coordinator.go — Test topology setup
func (c *Coordinator) setupProductionEquivalentTopology(ctx context.Context) (*Topology, error) {
    topo := &Topology{
        Services: map[string]*Service{
            "core-backend": {
                Image:    "helixplay/core:latest",
                Env:      productionEnv(),
                Networks: []string{"helix-backend"},
            },
            "host-agent": {
                Image:    "helixplay/host-agent:latest",
                Env:      productionEnv(),
                Networks: []string{"helix-backend", "helix-host"},
                Devices:  []string{"/dev/dri"},  // GPU passthrough
            },
            "client-web": {
                Image:    "helixplay/client-web:latest",
                Env:      productionEnv(),
                Networks: []string{"helix-backend"},
            },
        },
    }

    if err := topo.Start(ctx); err != nil {
        return nil, err
    }

    return topo, nil
}
```

### 5.2 Integration with Challenges

#### 5.2.1 Challenge Runner Coordination

HelixQA integrates with the Challenges repository at the runner level:

```go
// pkg/autonomous/challenge_adapter.go — Bridges HelixQA to Challenges framework

type QAChallengeAdapter struct {
    qaRunner    *helixqa.Runner
    challengeRunner *challenge.Runner
}

func (a *QAChallengeAdapter) ExecuteQAChallenge(ctx context.Context, def QAChallengeDefinition) (*challenge.Result, error) {
    // 1. Set up production-equivalent topology
    topo, err := a.qaRunner.SetupTopology(ctx, def.Topology)
    if err != nil {
        return nil, err
    }
    defer topo.Teardown()

    // 2. Execute the user flow
    flowResult, err := a.qaRunner.ExecuteFlow(ctx, def.UserFlow)
    if err != nil {
        return nil, err
    }

    // 3. Perform visual assertions
    visualResults, err := a.qaRunner.PerformVisualAssertions(ctx, def.VisualAssertions)
    if err != nil {
        return nil, err
    }

    // 4. Build challenge result with anti-bluff evidence
    result := &challenge.Result{
        Status: challenge.StatusPassed,
        RecordedActions: flowResult.Actions,
        Assertions: make([]challenge.Assertion, 0, len(visualResults)),
    }

    for _, vr := range visualResults {
        result.Assertions = append(result.Assertions, challenge.Assertion{
            Name:   vr.AssertionName,
            Passed: vr.Passed,
            Details: map[string]interface{}{
                "match_score": vr.MatchScore,
                "screenshot":  vr.ScreenshotPath,
            },
        })
    }

    // 5. Unconditional anti-bluff validation
    if err := challenge.ValidateAntiBluff(result); err != nil {
        result.Status = challenge.StatusError
        return result, err
    }

    return result, nil
}
```

#### 5.2.2 Evidence Collection

Every HelixQA challenge produces three forms of evidence:

1. **Screenshots**: Captured at each assertion point, stored as PNG artifacts
2. **Screen Recordings**: Full session recordings for manual review if needed
3. **Assertion Logs**: Structured logs with timestamps, match scores, and pass/fail status

#### 5.2.3 Pass/Fail Criteria

A HelixQA challenge passes only when:

1. All user flow steps execute without error
2. All visual assertions pass (match score >= tolerance)
3. `ValidateAntiBluff()` returns `nil` (has recorded actions AND at least one passing assertion)
4. No critical log errors during execution
5. Performance metrics within SLA (e.g., UI response time < 200ms)

---

## Section 6: CI/CD Pipeline Design

### 6.1 Current Gap Analysis

**CRITICAL FINDING**: Despite having 550+ test files, 8 challenge scripts, a mutation testing framework, and a constitutional mandate for comprehensive testing — **zero GitHub Actions workflows exist** in any repository.

#### 6.1.1 What Exists Now (Manual/Local Only)

| Component | Status | Runs Where | Gaps |
|-----------|--------|-----------|------|
| `go test ./...` | Local only | Developer machine | No visibility, no enforcement |
| `make anti-bluff` | Local only | Developer machine | Can be forgotten, no audit trail |
| `make challenge` | Local only | Developer machine | Same issues |
| `go-mutesting` | Local only | Developer machine | Same issues |
| `scripts/anti-bluff-scan.sh` | Local only | Developer machine | Same issues |
| Quality gates (SonarQube, Snyk) | Manual | External dashboards | Not tied to PR workflow |

#### 6.1.2 Impact of No CI/CD

1. **No visibility**: No one knows if tests are currently passing across all 29 submodules
2. **No enforcement**: Quality gates are voluntary; developers can (and do) skip them
3. **No audit trail**: There's no record of which quality checks ran on which commit
4. **Cannot detect regressions**: A commit that breaks tests can be merged undetected
5. **Blocks external contribution**: Contributors have no way to verify their changes
6. **Blocks automated deployment**: Deploying without verified quality gates violates the Constitution

#### 6.1.3 Dependency Resolution Issues

The `Makefile` explicitly notes:

> *"`vet` and `test` are intentionally NOT included in `qa-all` because several packages depend on missing replace-directives (`../Dependencies/HelixDevelopment/*`) that aren't present in a clean checkout."*

This means:
- `go test ./...` fails on a clean checkout
- `go vet ./...` fails on a clean checkout
- CI cannot run these commands without first fixing the dependency resolution

### 6.2 Required CI/CD Workflows

#### 6.2.1 `.github/workflows/ci.yml` — Full Pipeline

```yaml
name: CI — Full Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

# Prevent redundant runs
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  # ───────────────────────────────────────────
  # JOB 1: Pre-Flight Checks
  # ───────────────────────────────────────────
  preflight:
    name: Pre-Flight
    runs-on: ubuntu-latest
    outputs:
      go_version: "1.26.2"
      changed_modules: ${{ steps.changes.outputs.modules }}
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
          fetch-depth: 0  # Full history for change detection

      - name: Detect changed modules
        id: changes
        run: |
          CHANGED=$(git diff --name-only origin/${{ github.base_ref }} HEAD | \
            grep -E "^vasic-digital/|^cmd/|^pkg/|^tests/" | \
            cut -d/ -f2 | sort -u | jq -R -s -c 'split("\n")[:-1]')
          echo "modules=$CHANGED" >> $GITHUB_OUTPUT

      - name: Submodule integrity check
        run: python scripts/verify-submodules.py

  # ───────────────────────────────────────────
  # JOB 2: Build Matrix (Go versions)
  # ───────────────────────────────────────────
  build:
    name: Build (Go ${{ matrix.go }})
    runs-on: ubuntu-latest
    needs: [preflight]
    strategy:
      matrix:
        go: ['1.26.2']
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}

      - name: Fix replace directives
        run: bash scripts/fix-replace.sh

      - name: Download dependencies
        run: go mod download

      - name: Build all commands
        run: go build ./cmd/...

      - name: Build all submodules
        run: |
          for mod in vasic-digital/*/; do
            if [[ -f "$mod/go.mod" ]]; then
              echo "Building $mod..."
              (cd "$mod" && go build ./...)
            fi
          done

  # ───────────────────────────────────────────
  # JOB 3: Unit Tests (all submodules)
  # ───────────────────────────────────────────
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: [build]
    container: golang:1.26
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Fix replace directives
        run: bash scripts/fix-replace.sh

      - name: Run unit tests
        run: go test -count=1 -race -p 1 ./...
        timeout-minutes: 30

      - name: Run unit tests (submodules)
        run: |
          FAILURES=0
          for mod in vasic-digital/*/; do
            if [[ -f "$mod/go.mod" ]]; then
              echo "Testing $mod..."
              if ! (cd "$mod" && go test -count=1 -race -p 1 ./...); then
                FAILURES=$((FAILURES + 1))
              fi
            fi
          done
          if [[ $FAILURES -gt 0 ]]; then
            echo "$FAILURES submodule test suite(s) failed"
            exit 1
          fi
        timeout-minutes: 60

      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: |
            coverage.out
            coverage.html

  # ───────────────────────────────────────────
  # JOB 4: Integration Tests
  # ───────────────────────────────────────────
  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: [build]
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: helixtest
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports: ['25432:5432']
      redis:
        image: redis:7-alpine
        ports: ['26379:6379']
      nats:
        image: nats:2-alpine
        ports: ['4222:4222']
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Fix replace directives
        run: bash scripts/fix-replace.sh

      - name: Run integration tests
        run: go test -count=1 -race -p 1 -tags=integration ./tests/integration/...
        env:
          HELIX_TEST_DB_DSN: postgres://postgres:test@localhost:25432/helixtest?sslmode=disable
          HELIX_TEST_REDIS_ADDR: localhost:26379
          HELIX_TEST_NATS_URL: nats://localhost:4222
        timeout-minutes: 30

  # ───────────────────────────────────────────
  # JOB 5: E2E Tests
  # ───────────────────────────────────────────
  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest
    needs: [integration-tests]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build E2E topology
        run: docker compose -f tests/e2e/docker-compose.yml build

      - name: Run E2E tests
        run: go test -count=1 -v -timeout 30m ./tests/e2e/...
        timeout-minutes: 35

      - name: Collect logs on failure
        if: failure()
        run: docker compose -f tests/e2e/docker-compose.yml logs > e2e-logs.txt 2>&1

      - name: Upload E2E artifacts
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-failure-artifacts
          path: |
            e2e-logs.txt
            tests/e2e/screenshots/

  # ───────────────────────────────────────────
  # JOB 6: Security Scans
  # ───────────────────────────────────────────
  security:
    name: Security
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Run Semgrep SAST
        uses: semgrep/semgrep-action@v1
        with:
          config: >-
            p/security-audit
            p/owasp-top-ten
            p/cwe-top-25
            p/gosec

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          exit-code: '1'
          severity: 'HIGH,CRITICAL'

      - name: Run gitleaks secret detection
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  # ───────────────────────────────────────────
  # JOB 7: Code Quality
  # ───────────────────────────────────────────
  quality:
    name: Code Quality
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Run go vet
        run: |
          bash scripts/fix-replace.sh
          go vet ./...

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=10m

      - name: Run gofmt check
        run: |
          UNFORMATTED=$(gofmt -l .)
          if [[ -n "$UNFORMATTED" ]]; then
            echo "The following files need formatting:"
            echo "$UNFORMATTED"
            exit 1
          fi

      - name: SonarQube Scan
        uses: sonarqube-quality-gate-action@master
        env:
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
          SONAR_HOST_URL: ${{ secrets.SONAR_HOST_URL }}
        with:
          scanMetadataReportFile: .scannerwork/report-task.txt

  # ───────────────────────────────────────────
  # JOB 8: Benchmarks (no fail on regression, just report)
  # ───────────────────────────────────────────
  benchmarks:
    name: Benchmarks
    runs-on: ubuntu-latest
    needs: [unit-tests]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Run benchmarks
        run: go test -bench=. -benchmem -benchtime=5s ./tests/benchmark/... | tee benchmark.txt

      - name: Upload benchmark results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: benchmark.txt

      - name: Comment benchmark results
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          comment-on-alert: true
          alert-threshold: '150%'
```

#### 6.2.2 `.github/workflows/anti-bluff.yml` — Non-Overridable Scan

```yaml
name: Anti-Bluff — Non-Overridable

# This workflow CANNOT be skipped via [skip ci] or any other mechanism.
# It is the constitutional guarantee against bluff testing.

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

# This workflow always runs, regardless of what files changed
# NO path exclusions — every change is subject to anti-bluff verification

jobs:
  # ───────────────────────────────────────────
  # JOB 1: Vacuous Assertion Scan
  # ───────────────────────────────────────────
  vacuous-scan:
    name: Vacuous Assertion Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Run anti-bluff scan (full)
        run: bash scripts/anti-bluff-scan.sh
        timeout-minutes: 10

      - name: Upload scan report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: anti-bluff-report
          path: .anti-bluff-reports/

  # ───────────────────────────────────────────
  # JOB 2: Host Integrity (R-18)
  # ───────────────────────────────────────────
  host-integrity:
    name: Host Integrity (R-18)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Run claim check
        run: bash scripts/claim-check.sh

      - name: Container hazard scan
        run: bash scripts/claim-check.sh --container-scan

      - name: Verify no forbidden commands in codebase
        run: |
          FORBIDDEN=(
            'systemctl\s+suspend'
            'pm-suspend'
            'shutdown\s+'
            'poweroff'
            'reboot\s+'
            'mkfs\.'
            'rm\s+-rf\s+/\s'
            ':\(\)\{\s*:\|\s*:\s*\&\s*\};\s*:'
          )
          FAILURES=0
          for pattern in "${FORBIDDEN[@]}"; do
            if grep -r -P "$pattern" --include="*.go" --include="*.sh" --include="*.yml" .; then
              echo "FORBIDDEN COMMAND FOUND: $pattern"
              FAILURES=$((FAILURES + 1))
            fi
          done
          exit $FAILURES

  # ───────────────────────────────────────────
  # JOB 3: Constitution Propagation
  # ───────────────────────────────────────────
  constitution-check:
    name: Constitution Propagation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Run constitution propagation check
        run: bash scripts/propagate-constitution.sh --verify-only

      - name: Verify Constitution v2.1.0 references
        run: |
          # Check that the Constitution is referenced in key files
          for file in CLAUDE.md AGENTS.md; do
            if [[ -f "$file" ]]; then
              if ! grep -q "Constitution" "$file"; then
                echo "FAIL: $file does not reference Constitution"
                exit 1
              fi
            fi
          done
          echo "Constitution references verified"
```

#### 6.2.3 `.github/workflows/challenges.yml` — Challenge Execution

```yaml
name: Challenges

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
  schedule:
    - cron: '0 3 * * *'  # Daily at 3 AM

jobs:
  # ───────────────────────────────────────────
  # JOB 1: Build Challenge Runner
  # ───────────────────────────────────────────
  build-runner:
    name: Build Challenge Runner
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Build challenge runner
        run: cd vasic-digital/Challenges && go build -o challenge-runner ./cmd/userflow-runner

      - name: Build all challenge scripts
        run: |
          cd vasic-digital/Challenges/scripts
          chmod +x *.sh

      - name: Upload runner artifact
        uses: actions/upload-artifact@v4
        with:
          name: challenge-runner
          path: vasic-digital/Challenges/challenge-runner

  # ───────────────────────────────────────────
  # JOB 2: Run All Challenge Scripts
  # ───────────────────────────────────────────
  challenge-scripts:
    name: Challenge Scripts
    runs-on: ubuntu-latest
    needs: [build-runner]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Download runner
        uses: actions/download-artifact@v4
        with:
          name: challenge-runner
          path: ./challenge-runner

      - name: Make runner executable
        run: chmod +x ./challenge-runner/challenge-runner

      - name: Run anchor manifest challenge
        run: bash vasic-digital/Challenges/scripts/anchor_manifest_challenge.sh

      - name: Run bluff scanner challenge
        run: bash vasic-digital/Challenges/scripts/bluff_scanner_challenge.sh

      - name: Run compile challenge
        run: bash vasic-digital/Challenges/scripts/challenges_compile_challenge.sh

      - name: Run functionality challenge
        run: bash vasic-digital/Challenges/scripts/challenges_functionality_challenge.sh

      - name: Run unit challenge
        run: bash vasic-digital/Challenges/scripts/challenges_unit_challenge.sh

      - name: Run host no-auto-suspend challenge
        run: bash vasic-digital/Challenges/scripts/host_no_auto_suspend_challenge.sh

      - name: Run mutation ratchet challenge
        run: bash vasic-digital/Challenges/scripts/mutation_ratchet_challenge.sh

      - name: Run no-suspend-calls challenge
        run: bash vasic-digital/Challenges/scripts/no_suspend_calls_challenge.sh

      - name: Collect challenge reports
        if: always()
        run: |
          mkdir -p challenge-reports
          cp -r vasic-digital/Challenges/scripts/reports/* challenge-reports/ 2>/dev/null || true

      - name: Upload challenge reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: challenge-reports
          path: challenge-reports/

  # ───────────────────────────────────────────
  # JOB 3: Run Challenge Banks
  # ───────────────────────────────────────────
  challenge-banks:
    name: Challenge Banks
    runs-on: ubuntu-latest
    needs: [build-runner]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Download runner
        uses: actions/download-artifact@v4
        with:
          name: challenge-runner
          path: ./challenge-runner

      - name: Make runner executable
        run: chmod +x ./challenge-runner/challenge-runner

      - name: Run all challenge banks
        run: |
          ./challenge-runner/challenge-runner \
            --banks-dir vasic-digital/Challenges/banks \
            --all \
            --report challenge-bank-report.json
        timeout-minutes: 60

      - name: Verify anti-bluff validation in report
        run: |
          if ! jq '.antiBluffValid' challenge-bank-report.json | grep -q true; then
            echo "ANTI-BLUFF VALIDATION FAILED: Some challenges passed without evidence"
            exit 1
          fi

      - name: Upload bank report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: challenge-bank-report
          path: challenge-bank-report.json
```

#### 6.2.4 `.github/workflows/mutation.yml` — Mutation Testing

```yaml
name: Mutation Testing

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 4 * * 0'  # Weekly on Sunday at 4 AM

jobs:
  mutation:
    name: Mutation Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Install go-mutesting
        run: go install github.com/zimmski/go-mutesting@latest

      - name: Run mutation ratchet challenge
        run: bash vasic-digital/Challenges/scripts/mutation_ratchet_challenge.sh
        timeout-minutes: 120

      - name: Upload mutation report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: mutation-report
          path: |
            mutation-report.txt
            .mutation-score
```

#### 6.2.5 Container-Based Job Definitions

All CI jobs should eventually run in containers matching the production environment:

```yaml
# Example container-based job
container-test:
  runs-on: ubuntu-latest
  container:
    image: golang:1.26.2-alpine
    options: --cap-add=NET_ADMIN  # For network chaos tests
  services:
    helix-core:
      image: helixplay/core:test
      env:
        HELIX_ENV: test
        HELIX_DB_DSN: postgres://test@postgres/helix
    helix-host:
      image: helixplay/host-agent:test
      devices:
        - /dev/dri:/dev/dri  # GPU passthrough for encode tests
```

#### 6.2.6 Matrix Builds Across Go Versions

```yaml
strategy:
  matrix:
    go: ['1.25', '1.26.2']
    os: ['ubuntu-latest', 'ubuntu-24.04']
    include:
      - go: '1.26.2'
        os: 'ubuntu-latest'
        primary: true
    exclude:
      - go: '1.25'
        os: 'ubuntu-24.04'  # Reduce matrix size
```

#### 6.2.7 Artifact Collection and Retention

```yaml
# Global artifact retention policy
# Stored in .github/artifact-retention.yml
artifact-retention:
  coverage-reports:
    retention-days: 90
    compress: true
  challenge-reports:
    retention-days: 180
    compress: true
  failure-artifacts:
    retention-days: 30
    compress: true
  benchmark-results:
    retention-days: 365
    compress: false  # For trend analysis
```

### 6.3 Quality Gates

#### 6.3.1 SonarQube Configuration

```properties
# sonar-project.properties
sonar.projectKey=HelixPlay
sonar.projectName=HelixPlay Cloud Gaming Platform
sonar.sources=.
sonar.exclusions=vendor/**,**/*.pb.go,**/tests/**,**/scripts/**
sonar.tests=.
sonar.test.inclusions=**/*_test.go,**/tests/**
sonar.go.coverage.reportPaths=coverage.out
sonar.coverage.exclusions=cmd/**,**/mocks/**

# Quality gate thresholds (Constitution §7)
sonar.qualitygate.wait=true
sonar.coverage.minimum=100
sonar.duplications.minimum=3
```

**Gate Requirements**:
| Metric | Threshold | Action on Fail |
|--------|-----------|----------------|
| Coverage | >= 100% | Block merge |
| Duplications | <= 3% | Block merge |
| Code Smells | 0 (critical/blocker) | Block merge |
| Bugs | 0 | Block merge |
| Vulnerabilities | 0 | Block merge |
| Security Hotspots | 0 (high) | Block merge |

#### 6.3.2 Snyk Dependency Scanning

```yaml
snyk:
  name: Snyk Security Scan
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run Snyk
      uses: snyk/actions/golang@master
      env:
        SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
      with:
        args: --severity-threshold=high --fail-on=upgradable
```

#### 6.3.3 Semgrep Rule Configuration

```yaml
semgrep:
  name: Semgrep SAST
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run Semgrep
      uses: semgrep/semgrep-action@v1
      with:
        config: >-
          p/security-audit
          p/owasp-top-ten
          p/cwe-top-25
          p/gosec
          p/golang
          p/trailofbits
```

#### 6.3.4 Trivy Container Scanning

```yaml
trivy:
  name: Trivy Container Scan
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Build container image
      run: docker build -t helixplay:test .
    - name: Scan image
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: helixplay:test
        format: 'sarif'
        output: 'trivy-results.sarif'
        severity: 'HIGH,CRITICAL'
    - name: Upload results
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: trivy-results.sarif
```

#### 6.3.5 gitleaks Secret Scanning

```yaml
gitleaks:
  name: Secret Detection
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0  # Full history for secret detection
    - uses: gitleaks/gitleaks-action@v2
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        GITLEAKS_ENABLE_COMMENTS: true
```

#### 6.3.6 govulncheck Go Vulnerability Scanning

```yaml
govulncheck:
  name: Go Vulnerability Check
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: 1.26.2
    - run: go install golang.org/x/vuln/cmd/govulncheck@latest
    - run: govulncheck ./...
```

#### 6.3.7 Coverage Thresholds

| Metric | Threshold | Enforcement |
|--------|-----------|-------------|
| Line Coverage | 100% | SonarQube gate + CI fail |
| Branch Coverage | 100% | SonarQube gate + CI fail |
| Function Coverage | 100% | SonarQube gate + CI fail |
| Package Coverage | 100% | CI fail if any package < 100% |
| Mutation Score | >= 85% | `mutation_ratchet_challenge.sh` |

#### 6.3.8 Benchmark Regression Detection

```yaml
benchmark-regression:
  name: Benchmark Regression
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: 1.26.2
    - name: Run benchmarks
      run: go test -bench=. -benchmem -count=5 ./tests/benchmark/... | tee benchmark.txt
    - name: Compare with baseline
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'go'
        output-file-path: benchmark.txt
        external-data-json-path: ./cache/benchmark-data.json
        github-token: ${{ secrets.GITHUB_TOKEN }}
        alert-threshold: '150%'
        comment-on-alert: true
        fail-on-alert: true
        auto-push: true
```

---

## Section 7: Fixing Current Issues

### 7.1 Dependency Resolution

#### 7.1.1 Root Cause

The `go.mod` in the main repository and several submodules contain `replace` directives that reference paths outside the repository:

```go
// Current problematic replace directive
replace github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory => ./vasic-digital/Memory

// Hypothetical problematic directive (from Makefile note)
replace github.com/HelixDevelopment/Dependencies/SomeModule => ../Dependencies/HelixDevelopment/SomeModule
```

The `../Dependencies/` path does not exist in a clean checkout because:
1. It is a sibling directory, not a submodule
2. It is not cloned by `git submodule update --init --recursive`
3. It may be a private repository or local development dependency

#### 7.1.2 Fix: Replace Directives

**Strategy 1**: Convert sibling dependencies to proper Git submodules:

```bash
# scripts/fix-replace.sh — Dependency resolution fix
#!/bin/bash
set -euo pipefail

# Check if Dependencies directory exists
if [[ ! -d "../Dependencies" ]]; then
    echo "Dependencies directory not found. Using submodule fallback."

    # For each broken replace directive, try to resolve via submodule
    for modfile in $(find . -name "go.mod" -not -path "*/vendor/*"); do
        dir=$(dirname "$modfile")
        # Replace ../Dependencies references with proper module paths
        sed -i 's|=>\s*\.\./Dependencies/HelixDevelopment/\(.*\)|=> ./vasic-digital/\1|g' "$modfile" 2>/dev/null || true
    done
fi

# Verify all replace directives resolve
for modfile in $(find . -name "go.mod" -not -path "*/vendor/*"); do
    dir=$(dirname "$modfile")
    (cd "$dir" && go mod verify 2>/dev/null) || {
        echo "WARNING: $modfile has unresolvable replace directives"
    }
done

echo "Replace directive fix complete"
```

**Strategy 2**: Use `go.work` workspace configuration:

```go
// go.work — Go workspace for multi-module development
go 1.26.2

use (
    .
    ./vasic-digital/Auth
    ./vasic-digital/Cache
    ./vasic-digital/Challenges
    ./vasic-digital/Database
    ./vasic-digital/Discovery
    ./vasic-digital/EventBus
    ./vasic-digital/Formatters
    ./vasic-digital/Media
    ./vasic-digital/Memory
    ./vasic-digital/Messaging
    ./vasic-digital/Middleware
    ./vasic-digital/Observability
    ./vasic-digital/Plugins
    ./vasic-digital/RAG
    ./vasic-digital/RateLimiter
    ./vasic-digital/Recovery
    ./vasic-digital/Security
    ./vasic-digital/Storage
    ./vasic-digital/Streaming
    ./vasic-digital/VectorDB
)

// External dependencies that must be cloned separately
// These are NOT in the workspace and must be fetched via go mod download
replace github.com/HelixDevelopment/HelixQA => ./HelixQA
```

#### 7.1.3 Ensure Clean Checkout Builds

```yaml
# CI verification step
clean-checkout-test:
  name: Clean Checkout Build
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive

    - name: Fresh clone — no cached state
      run: |
        # Remove any cached module state
        rm -rf ~/go/pkg/mod/cache
        go clean -cache -modcache

    - name: Fix replace directives
      run: bash scripts/fix-replace.sh

    - name: Download all dependencies
      run: |
        go mod download
        go work sync 2>/dev/null || true

    - name: Build everything
      run: |
        go build ./cmd/...
        for mod in vasic-digital/*/; do
          if [[ -f "$mod/go.mod" ]]; then
            (cd "$mod" && go build ./...)
          fi
        done

    - name: Run vet
      run: go vet ./...

    - name: Run tests
      run: go test -count=1 ./...
```

#### 7.1.4 go.work Workspace Configuration

The `go.work` file should be:
1. **Committed to the repository** so all developers use the same workspace
2. **Used by CI** so CI builds match local builds
3. **Kept in sync** with `.gitmodules` so every submodule in git is in the workspace

```bash
# scripts/sync-workspace.sh — Keep go.work in sync with submodules
#!/bin/bash
set -euo pipefail

WORK_FILE="go.work"
SUBMODULES=$(git submodule status | awk '{print $2}')

echo "go 1.26.2" > "$WORK_FILE"
echo "" >> "$WORK_FILE"
echo "use (" >> "$WORK_FILE"
echo "    ." >> "$WORK_FILE"

for submod in $SUBMODULES; do
    # Only include Go modules
    if [[ -f "$submod/go.mod" ]]; then
        echo "    ./$submod" >> "$WORK_FILE"
    fi
done

echo ")" >> "$WORK_FILE"

echo "go.work synced with $(echo "$SUBMODULES" | wc -w) submodules"
```

### 7.2 Test Gaps

#### 7.2.1 Missing E2E Tests

The following submodules have **zero E2E tests** and need them urgently:

| Submodule | Why E2E is Critical | Suggested E2E Scenario | Effort |
|-----------|-------------------|----------------------|--------|
| Cache | Cache invalidation must work end-to-end | Write through cache, invalidate, verify miss | 2 days |
| EventBus | Event delivery is core to system operation | Publish events across services, verify delivery | 2 days |
| Messaging | Message queue reliability | Produce/consume messages, verify ordering, test failure recovery | 3 days |
| Recovery | Fault recovery must work in production topology | Inject fault, verify automatic recovery | 3 days |
| client-wails | Desktop client is primary user interface | Full user journey from discovery to gameplay | 5 days |
| core-backend | Core backend is the system backbone | Full request lifecycle through all backend services | 4 days |

#### 7.2.2 Missing Integration Between Submodules

Current tests are mostly **intra-module** (testing within a single submodule). The following **inter-module** integration paths are untested:

| Integration Path | Components Involved | Test Scenario | Priority |
|-----------------|-------------------|--------------|----------|
| Auth -> Database | Auth, Database | User registration writes to DB; login reads from DB | P1 |
| Auth -> Middleware | Auth, Middleware | Authenticated requests are properly authorized | P1 |
| Discovery -> Core | Discovery, Core Backend | Host discovery results are available via core API | P1 |
| Streaming -> Storage | Streaming, Storage | Stream recordings are persisted to storage | P1 |
| RateLimiter -> Middleware | RateLimiter, Middleware | Rate limiting is applied to HTTP routes | P2 |
| RAG -> VectorDB | RAG, VectorDB | Document embedding and retrieval | P2 |
| Media -> Streaming | Media, Streaming | Media transcoding in streaming pipeline | P2 |

#### 7.2.3 Constructor-Only Tests Needing Behavior Verification

The following test files were identified as having constructor-only patterns that need behavior verification added:

| Test File | Current Pattern | Needed Addition | Effort |
|-----------|---------------|-----------------|--------|
| `pkg/autonomous/real_executor_test.go` | `require.NotNil(exec)` | Execute a command on each platform executor, verify output | 2 days |
| Various submodule tests | `require.NoError(err)` after `NewX()` | Call at least one method on the constructed object | 1 day per submodule |

### 7.3 CI Migration Plan

#### 7.3.1 Step-by-Step Migration

**Phase 1: Foundation (Week 1)**

| Day | Task | Verification |
|-----|------|-------------|
| 1 | Create `.github/workflows/` directory in all 3 repos | Directory exists and is committed |
| 1 | Implement `fix-replace.sh` script | Clean checkout builds successfully |
| 2 | Create `go.work` workspace file | `go work sync` succeeds |
| 2 | Create `.github/workflows/ci.yml` — basic build only | Build job passes on PR |
| 3 | Add unit test job to `ci.yml` | Unit test job passes, reports coverage |
| 3 | Add `golangci-lint` and `go vet` jobs | Lint job passes |
| 4 | Create `.github/workflows/anti-bluff.yml` | Anti-bluff scan runs on every PR |
| 4 | Set branch protection rules (require anti-bluff) | Cannot merge without anti-bluff pass |
| 5 | Test full pipeline on a PR | All jobs pass, merge blocked on failure |

**Phase 2: Quality Gates (Week 2)**

| Day | Task | Verification |
|-----|------|-------------|
| 6 | Add security scan job (govulncheck, Semgrep, Trivy, gitleaks) | Security job passes, detects test vulnerability |
| 7 | Add SonarQube integration | Coverage appears in SonarQube dashboard |
| 8 | Add Snyk dependency scanning | Snyk report shows 0 high-severity issues |
| 9 | Add benchmark job | Benchmark results are collected and stored |
| 10 | Create `.github/workflows/mutation.yml` | Mutation ratchet runs weekly |

**Phase 3: Integration & E2E (Week 3)**

| Day | Task | Verification |
|-----|------|-------------|
| 11 | Add integration test job with test services | Integration tests pass against real DB/cache |
| 12 | Create E2E Docker Compose topology | `docker compose up` starts all services |
| 13 | Add E2E test job | E2E tests pass in CI |
| 14 | Create `.github/workflows/challenges.yml` | All 8 challenge scripts pass |
| 15 | Add challenge bank execution | Challenge bank report shows antiBluffValid=true |

**Phase 4: Advanced Testing (Week 4)**

| Day | Task | Verification |
|-----|------|-------------|
| 16 | Add chaos test job | Chaos tests detect simulated container failures |
| 17 | Add stress test job | 4-hour stress profile completes, no goroutine leaks |
| 18 | Add smoke test job (post-deploy) | Smoke tests verify staging after deployment |
| 19 | Add full automation job | Full deployment test completes without manual steps |
| 20 | Add negative-leg fault injection job | Fault injection detects missing test coverage |

#### 7.3.2 Priority Order for Workflow Implementation

```
1. anti-bluff.yml     (MUST be first — constitutional requirement)
2. ci.yml (basic)     (build + unit tests)
3. ci.yml (extended)  (+ security + quality gates)
4. mutation.yml       (weekly mutation testing)
5. challenges.yml     (challenge execution)
6. ci.yml (full)      (+ integration + E2E + chaos + stress)
```

#### 7.3.3 Testing Each Workflow Before Merge

Every workflow must pass a **self-test** before being merged:

1. **Workflow syntax validation**: `actionlint` or GitHub's workflow editor
2. **Dry run**: Use `act` (local GitHub Actions runner) to test the workflow locally
3. **Intentional failure test**: Introduce a known defect and verify the workflow catches it
4. **Intentional success test**: Fix the defect and verify the workflow passes
5. **Timeout test**: Verify jobs timeout correctly rather than hanging indefinitely

```bash
# Self-test script for workflows
#!/bin/bash
# scripts/test-workflow.sh

WORKFLOW=$1

echo "Testing workflow: $WORKFLOW"

# 1. Syntax validation
if ! command -v actionlint &> /dev/null; then
    go install github.com/rhysd/actionlint/cmd/actionlint@latest
fi
actionlint "$WORKFLOW"

# 2. Dry run with act (if available)
if command -v act &> /dev/null; then
    act -W "$WORKFLOW" --dry-run
fi

echo "Workflow $WORKFLOW passed self-test"
```

---

## Section 8: Challenges Implementation Plan

### 8.1 Challenge Scripts to Implement

All 8 challenge scripts live in `vasic-digital/Challenges/scripts/` and are executed via `make challenge` or the CI `challenges.yml` workflow.

#### 8.1.1 `anchor_manifest_challenge.sh` — Behavior-Anchor Manifest Validator

**Purpose**: Verify that every submodule has a behavior-anchor manifest — a YAML file that documents the observable behaviors the module guarantees and the tests that verify them.

```bash
#!/bin/bash
# challenges/scripts/anchor_manifest_challenge.sh

set -euo pipefail

MANIFEST_FILE=".behavior-anchors.yml"
FAILURES=0

# Check that manifest exists for every submodule
for submod in vasic-digital/*/; do
    if [[ ! -f "$submod/$MANIFEST_FILE" ]]; then
        echo "FAIL: $submod missing behavior-anchor manifest"
        ((FAILURES++))
        continue
    fi

    # Validate manifest structure
    if ! yq eval '.anchors' "$submod/$MANIFEST_FILE" > /dev/null 2>&1; then
        echo "FAIL: $submod manifest missing 'anchors' section"
        ((FAILURES++))
        continue
    fi

    # Verify each anchor has a test reference
    anchor_count=$(yq eval '.anchors | length' "$submod/$MANIFEST_FILE")
    for ((i=0; i<anchor_count; i++)); do
        anchor_name=$(yq eval ".anchors[$i].name" "$submod/$MANIFEST_FILE")
        test_ref=$(yq eval ".anchors[$i].verified_by_test" "$submod/$MANIFEST_FILE")

        if [[ -z "$test_ref" || "$test_ref" == "null" ]]; then
            echo "FAIL: Anchor '$anchor_name' in $submod has no test reference"
            ((FAILURES++))
        elif [[ ! -f "$submod/$test_ref" ]]; then
            echo "FAIL: Anchor '$anchor_name' references missing test: $test_ref"
            ((FAILURES++))
        fi
    done
done

if [[ $FAILURES -gt 0 ]]; then
    echo "Anchor manifest challenge FAILED: $FAILURES issue(s)"
    exit 1
fi

echo "Anchor manifest challenge PASSED"
```

#### 8.1.2 `bluff_scanner_challenge.sh` — Two-Phase Bluff Scanner

**Purpose**: Verify the bluff scanner itself works correctly (Phase 1: self-test), then run it against the full tree (Phase 2).

```bash
#!/bin/bash
# challenges/scripts/bluff_scanner_challenge.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCANNER="$SCRIPT_DIR/anti-bluff/bluff-scanner.sh"
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# ─── PHASE 1: Scanner Self-Test ───
echo "Phase 1: Scanner self-test..."

# Create hand-crafted bluff fixtures that the scanner MUST detect
cat > "$TEMP_DIR/bluff_fixture_test.go" <<'EOF'
package bluff

import "testing"

// This is a DELIBERATELY bluff test — scanner must detect it
func TestBluffFixture_TrueIsTrue(t *testing.T) {
    assert.True(t, true)           // Vacuous: always passes
    assert.Equal(t, true, true)    // Vacuous: same value
    assert.Nil(t, nil)             // Vacuous: nil is nil
}

func TestBluffFixture_Empty(t *testing.T) {
    // Empty body — passes without doing anything
}
EOF

# Run scanner against fixtures
SCAN_OUTPUT=$($SCANNER "$TEMP_DIR" 2>&1) || true

# Verify scanner detected all bluff patterns
DETECTIONS=0
if echo "$SCAN_OUTPUT" | grep -q "assert.True.*true"; then
    ((DETECTIONS++))
else
    echo "FAIL: Scanner did not detect assert.True(t, true)"
fi

if echo "$SCAN_OUTPUT" | grep -q "assert.Equal.*true.*true"; then
    ((DETECTIONS++))
else
    echo "FAIL: Scanner did not detect assert.Equal(t, true, true)"
fi

if echo "$SCAN_OUTPUT" | grep -q "assert.Nil.*nil"; then
    ((DETECTIONS++))
else
    echo "FAIL: Scanner did not detect assert.Nil(t, nil)"
fi

if [[ $DETECTIONS -lt 3 ]]; then
    echo "Phase 1 FAILED: Scanner self-test detected only $DETECTIONS/3 bluff patterns"
    echo "The bluff scanner itself is broken — this is a CRITICAL failure."
    exit 1
fi

echo "Phase 1 PASSED: Scanner correctly detected $DETECTIONS/3 bluff patterns"

# ─── PHASE 2: Full Tree Scan ───
echo "Phase 2: Full source tree scan..."

FULL_OUTPUT=$($SCANNER "$(git rev-parse --show-toplevel)" 2>&1) || {
    echo "Phase 2 FAILED: Full tree scan found bluff patterns"
    echo "$FULL_OUTPUT"
    exit 1
}

echo "Phase 2 PASSED: No bluff patterns in full tree"
echo "Bluff scanner challenge PASSED"
```

#### 8.1.3 `challenges_compile_challenge.sh` — Compilation Verification

**Purpose**: Verify that all challenges and their dependencies compile successfully.

```bash
#!/bin/bash
# challenges/scripts/challenges_compile_challenge.sh

set -euo pipefail

CHALLENGES_DIR="vasic-digital/Challenges"
FAILURES=0

echo "Challenges compile challenge..."

# Compile the main challenge module
echo "Compiling challenge framework..."
if ! (cd "$CHALLENGES_DIR" && go build ./...); then
    echo "FAIL: Challenge framework does not compile"
    ((FAILURES++))
fi

# Compile challenge runner
echo "Compiling challenge runner..."
if ! (cd "$CHALLENGES_DIR" && go build -o /tmp/challenge-runner ./cmd/userflow-runner); then
    echo "FAIL: Challenge runner does not compile"
    ((FAILURES++))
fi

# Compile all challenge bank definitions
echo "Verifying challenge bank YAML..."
for bank in "$CHALLENGES_DIR"/banks/*.yaml; do
    if [[ -f "$bank" ]]; then
        if ! yq eval '.' "$bank" > /dev/null 2>&1; then
            echo "FAIL: Invalid YAML in $bank"
            ((FAILURES++))
        fi
    fi
done

# Verify all referenced challenge handlers exist
echo "Verifying challenge handler references..."
for bank in "$CHALLENGES_DIR"/banks/*.yaml; do
    if [[ -f "$bank" ]]; then
        handlers=$(yq eval '.challenges[].handler' "$bank" 2>/dev/null || true)
        for handler in $handlers; do
            if [[ "$handler" != "null" && -n "$handler" ]]; then
                # Check handler is registered
                if ! grep -r "Register.*$handler" "$CHALLENGES_DIR/pkg/" > /dev/null 2>&1; then
                    echo "FAIL: Handler '$handler' from $bank not registered"
                    ((FAILURES++))
                fi
            fi
        done
    fi
done

if [[ $FAILURES -gt 0 ]]; then
    echo "Challenges compile challenge FAILED: $FAILURES issue(s)"
    exit 1
fi

echo "Challenges compile challenge PASSED"
```

#### 8.1.4 `challenges_functionality_challenge.sh` — Functional Test Challenge

**Purpose**: Execute a subset of challenges that verify core functionality of the platform.

```bash
#!/bin/bash
# challenges/scripts/challenges_functionality_challenge.sh

set -euo pipefail

CHALLENGES_DIR="vasic-digital/Challenges"
RUNNER="$CHALLENGES_DIR/challenge-runner"
REPORT_FILE="/tmp/functional-challenge-report.json"
FAILURES=0

echo "Challenges functionality challenge..."

# Build runner if needed
if [[ ! -x "$RUNNER" ]]; then
    (cd "$CHALLENGES_DIR" && go build -o challenge-runner ./cmd/userflow-runner)
fi

# Run functionality challenges
if ! "$RUNNER" \
    --banks-dir "$CHALLENGES_DIR/banks" \
    --categories "game-streaming,navigation,recording" \
    --report "$REPORT_FILE"; then
    echo "FAIL: Functionality challenges execution failed"
    ((FAILURES++))
fi

# Verify anti-bluff validation
if [[ -f "$REPORT_FILE" ]]; then
    passed=$(jq '.summary.passed // 0' "$REPORT_FILE")
    total=$(jq '.summary.total // 0' "$REPORT_FILE")
    anti_bluff_valid=$(jq '.antiBluffValid // false' "$REPORT_FILE")

    echo "Results: $passed/$total challenges passed"

    if [[ "$anti_bluff_valid" != "true" ]]; then
        echo "FAIL: Anti-bluff validation failed"
        ((FAILURES++))
    fi

    if [[ $passed -eq 0 ]]; then
        echo "FAIL: No challenges passed"
        ((FAILURES++))
    fi
else
    echo "FAIL: No challenge report generated"
    ((FAILURES++))
fi

if [[ $FAILURES -gt 0 ]]; then
    echo "Challenges functionality challenge FAILED"
    exit 1
fi

echo "Challenges functionality challenge PASSED"
```

#### 8.1.5 `challenges_unit_challenge.sh` — Unit Test Challenge

**Purpose**: Run all unit tests in the Challenges repository and verify they pass.

```bash
#!/bin/bash
# challenges/scripts/challenges_unit_challenge.sh

set -euo pipefail

CHALLENGES_DIR="vasic-digital/Challenges"
FAILURES=0

echo "Challenges unit challenge..."

# Run all unit tests with race detection
echo "Running Challenges unit tests..."
if ! (cd "$CHALLENGES_DIR" && go test -count=1 -race -p 1 ./...); then
    echo "FAIL: Challenges unit tests failed"
    ((FAILURES++))
fi

# Verify coverage
echo "Checking coverage..."
coverage=$(cd "$CHALLENGES_DIR" && go test -coverprofile=/tmp/challenges-coverage.out ./... 2>&1 | \
    grep -oP 'coverage: \K[0-9.]+' | tail -1)

echo "Challenges coverage: ${coverage}%"

# Coverage threshold: 80% for challenges framework
if (( $(echo "$coverage < 80" | bc -l) )); then
    echo "FAIL: Coverage $coverage% is below 80% threshold"
    ((FAILURES++))
fi

if [[ $FAILURES -gt 0 ]]; then
    echo "Challenges unit challenge FAILED"
    exit 1
fi

echo "Challenges unit challenge PASSED (coverage: ${coverage}%)"
```

#### 8.1.6 `host_no_auto_suspend_challenge.sh` — Host Integrity Verification

**Purpose**: Verify that the host system is configured to never auto-suspend, as this would disrupt streaming sessions.

```bash
#!/bin/bash
# challenges/scripts/host_no_auto_suspend_challenge.sh

set -euo pipefail

FAILURES=0

echo "Host no-auto-suspend challenge..."

# Check systemd sleep configuration
if [[ -f "/etc/systemd/sleep.conf" ]]; then
    if grep -q "SuspendMode\|HibernateMode" "/etc/systemd/sleep.conf" 2>/dev/null; then
        echo "FAIL: sleep.conf contains suspend/hibernate configuration"
        ((FAILURES++))
    fi
fi

# Check systemd sleep.conf.d/
if [[ -d "/etc/systemd/sleep.conf.d" ]]; then
    for conf in /etc/systemd/sleep.conf.d/*.conf; do
        if [[ -f "$conf" ]]; then
            if grep -q "SuspendMode\|HibernateMode\|AllowHibernation" "$conf" 2>/dev/null; then
                echo "FAIL: $conf contains suspend/hibernate configuration"
                ((FAILURES++))
            fi
        fi
    done
fi

# Check logind.conf
if [[ -f "/etc/systemd/logind.conf" ]]; then
    # These settings must be explicitly set to ignore
    if ! grep -q "HandleLidSwitch=ignore" "/etc/systemd/logind.conf" 2>/dev/null; then
        echo "FAIL: logind.conf does not set HandleLidSwitch=ignore"
        ((FAILURES++))
    fi
    if ! grep -q "HandleSuspendKey=ignore" "/etc/systemd/logind.conf" 2>/dev/null; then
        echo "FAIL: logind.conf does not set HandleSuspendKey=ignore"
        ((FAILURES++))
    fi
    if ! grep -q "HandleHibernateKey=ignore" "/etc/systemd/logind.conf" 2>/dev/null; then
        echo "FAIL: logind.conf does not set HandleHibernateKey=ignore"
        ((FAILURES++))
    fi
fi

# Check for any running suspend services
if systemctl is-active --quiet suspend.target hibernate.target hybrid-sleep.target 2>/dev/null; then
    echo "FAIL: A suspend/hibernate target is active"
    ((FAILURES++))
fi

# Verify no suspend in cron/systemd timers
if grep -r "suspend\|hibernate" /etc/cron.* /var/spool/cron/ 2>/dev/null; then
    echo "FAIL: Found suspend/hibernate in cron jobs"
    ((FAILURES++))
fi

if [[ $FAILURES -gt 0 ]]; then
    echo "Host no-auto-suspend challenge FAILED: $FAILURES issue(s)"
    echo "WARNING: Host may auto-suspend, which will disrupt streaming!"
    exit 1
fi

echo "Host no-auto-suspend challenge PASSED"
```

#### 8.1.7 `mutation_ratchet_challenge.sh` — Mutation Testing Ratchet

See Section 4.4.2 for the full implementation. This script:
1. Runs `go-mutesting` on the full codebase
2. Compares the score against the stored baseline
3. Fails if the score decreases (ratchet violation)
4. Fails if the score is below the 85% threshold
5. Updates the stored score on success

#### 8.1.8 `no_suspend_calls_challenge.sh` — No Suspend Calls Verification

**Purpose**: Verify that no code in the repository calls suspend, hibernate, or shutdown functions.

```bash
#!/bin/bash
# challenges/scripts/no_suspend_calls_challenge.sh

set -euo pipefail

FAILURES=0

echo "No-suspend-calls challenge..."

# Forbidden function calls in Go code
FORBIDDEN_GO=(
    'exec\.Command.*"systemctl".*"suspend"'
    'exec\.Command.*"systemctl".*"hibernate"'
    'exec\.Command.*"pm-suspend"'
    'exec\.Command.*"rtcwake"'
    'exec\.Command.*"shutdown"'
    'exec\.Command.*"poweroff"'
    'exec\.Command.*"reboot"'
    'syscall\.Reboot'
)

# Forbidden shell commands in scripts
FORBIDDEN_SHELL=(
    'systemctl\s+suspend'
    'systemctl\s+hibernate'
    'pm-suspend'
    'rtcwake'
    'shutdown\s+'
    'poweroff'
    'reboot\s+'
    'init\s+0'
)

# Scan Go code
for pattern in "${FORBIDDEN_GO[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.go" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found forbidden suspend call in Go code:"
        echo "$matches"
        ((FAILURES++))
    fi
done

# Scan shell scripts
for pattern in "${FORBIDDEN_SHELL[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.sh" --include="*.bash" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found forbidden suspend call in shell script:"
        echo "$matches"
        ((FAILURES++))
    fi
done

# Scan YAML files (for container commands)
for pattern in "${FORBIDDEN_SHELL[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.yml" --include="*.yaml" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found forbidden suspend call in YAML:"
        echo "$matches"
        ((FAILURES++))
    fi
done

# Check for imports of dangerous packages
DANGEROUS_IMPORTS=(
    'syscall.*REBOOT'
    'golang.org/x/sys/unix.*Reboot'
)
for pattern in "${DANGEROUS_IMPORTS[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.go" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found dangerous reboot import:"
        echo "$matches"
        ((FAILURES++))
    fi
done

if [[ $FAILURES -gt 0 ]]; then
    echo "No-suspend-calls challenge FAILED: $FAILURES forbidden call(s) found"
    exit 1
fi

echo "No-suspend-calls challenge PASSED: No suspend/hibernate/shutdown calls found"
```

### 8.2 Challenge Banks

The Challenges repository contains 50+ challenge bank files covering all aspects of the platform.

#### 8.2.1 Game-Specific Challenges

**File**: `vasic-digital/Challenges/banks/game-streaming.yaml`

```yaml
challenge:
  name: "HelixPlay Game Stream Validation"
  category: "game-streaming"
  description: "Verify game streaming works for each supported title"
  games:
    - name: "Elden Ring"
      platform: "steam"
      expected_fps: 60
      expected_latency_ms: 30
    - name: "Cyberpunk 2077"
      platform: "steam"
      expected_fps: 60
      expected_latency_ms: 30
    - name: " Baldur's Gate 3"
      platform: "steam"
      expected_fps: 60
      expected_latency_ms: 30
  steps:
    - name: "Launch game"
      action: "host.launch_game"
      timeout: 60s
    - name: "Verify stream"
      action: "client.verify_stream"
      assertions:
        - fps: ">= ${game.expected_fps}"
        - latency_ms: "<= ${game.expected_latency_ms}"
        - resolution: "1920x1080"
  antiBluff:
    recordedActions: ["host.launch_game", "client.verify_stream"]
    assertions: ["fps", "latency_ms", "resolution"]
```

#### 8.2.2 Full QA Challenges

**Android**: `vasic-digital/Challenges/banks/full-qa-android.yaml`
- Tests all UI flows on Android mobile client
- Includes touch gestures, app lifecycle, background/foreground
- Requires HelixQA visual assertion for each screen

**Android TV**: `vasic-digital/Challenges/banks/full-qa-androidtv.yaml`
- Tests all UI flows on Android TV (leanback) client
- Includes D-pad navigation, voice search, recommendation row
- Requires HelixQA visual assertion for each screen

```yaml
challenge:
  name: "Full QA — Android TV"
  category: "full-qa"
  platform: "androidtv"
  flows:
    - name: "Main Navigation"
      steps:
        - action: "navigate.home"
          visual_assert: "home_screen_visible"
        - action: "navigate.library"
          visual_assert: "library_screen_visible"
        - action: "navigate.settings"
          visual_assert: "settings_screen_visible"
    - name: "Game Launch"
      steps:
        - action: "navigate.library"
        - action: "select.game"
          params: {game: "Elden Ring"}
        - action: "click.play"
          visual_assert: "stream_active"
          timeout: 30s
  antiBluff:
    visual_assertions_required: true
    recordedActions: ["navigate.*", "select.*", "click.*"]
```

#### 8.2.3 Navigation Challenges

**File**: `vasic-digital/Challenges/banks/navigation.yaml`

Tests all navigation flows across all client platforms:
- Home -> Library -> Game Detail -> Play
- Home -> Settings -> Account -> Logout
- Home -> Search -> Results -> Game Detail
- In-stream: Pause -> Resume -> Quit
- In-stream: Switch input method (touch/gamepad/keyboard)

#### 8.2.4 Recording Challenges

**File**: `vasic-digital/Challenges/banks/recording.yaml`

Tests the recording feature (host-agent encoder dual-path):
- Start stream -> Verify recording starts automatically
- Stop stream -> Verify recording is sealed
- Verify recording file exists with expected size
- Verify recording can be played back
- Verify recording metadata (duration, resolution, codec)

---

## Appendix A: Quick Reference — Anti-Bluff Checklist

### For Every New Feature

- [ ] Unit tests with >= 100% branch coverage (mock isolation logic only)
- [ ] Integration tests with **real** dependencies (no mocks)
- [ ] E2E test covering the complete user journey
- [ ] Security test (fuzz target + specific attack vectors)
- [ ] Benchmark covering performance-critical paths
- [ ] Negative-leg test (break the feature, verify test fails)
- [ ] `ValidateAntiBluff()` evidence in challenge result
- [ ] Usability evidence (HelixQA visual assertion, recording, or challenge execution)
- [ ] Behavior-anchor manifest entry
- [ ] No forbidden patterns in code (scanner clean)

### For Every PR

- [ ] All 8 challenge scripts pass locally
- [ ] `make anti-bluff` passes
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] `golangci-lint` passes
- [ ] No new vacuous assertions introduced
- [ ] Coverage does not decrease
- [ ] Anti-bluff scan passes in CI (non-overridable)

### CI Workflow Status Dashboard

| Workflow | Status | Priority | Implemented |
|----------|--------|----------|-------------|
| `anti-bluff.yml` | **MANDATORY** | P0 | No |
| `ci.yml` (basic) | Required | P0 | No |
| `ci.yml` (extended) | Required | P1 | No |
| `challenges.yml` | Required | P1 | No |
| `mutation.yml` | Required | P2 | No |
| Full CI (all 10 types) | Required | P3 | No |

---

*Document generated as part of the HelixPlay Comprehensive Implementation Plan.  
Constitutional Reference: v2.1.0  
All testing requirements are mandatory per Constitution §6 and cannot be overridden.*
