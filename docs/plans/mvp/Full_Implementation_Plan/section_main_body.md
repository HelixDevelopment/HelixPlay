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
