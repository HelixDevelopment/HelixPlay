# HelixPlay Repository - Complete Discovery Report

**Repository**: https://github.com/HelixDevelopment/HelixPlay
**Analyzed**: 2026-05-02
**Commits**: 126 (main branch)
**Contributors**: 2 (milos85vasic, claude)
**Stars/Forks**: 0/0 (private/public dev phase)

---

## 1. TOP-LEVEL FILES AND DIRECTORIES

### Root Files (9 total)
| File | Description | Significance |
|------|-------------|-------------|
| `AGENTS.md` | AI agent configuration for non-Claude tools | **CRITICAL** - Prime directive, Constitution refs, build commands |
| `CLAUDE.md` | Claude Code (claude.ai/code) guidance | **CRITICAL** - Same content as AGENTS.md but for Claude, v2.1.0 |
| `go.mod` | Go module definition (Go 1.26.2) | **CRITICAL** - Module graph with submodules, Brotli, UUID, testify |
| `go.sum` | Go module checksums | **CRITICAL** - Dependency verification |
| `config.json` | Application configuration | Server, DB (PostgreSQL), auth, catalog, storage, logging, proxy |
| `.gitignore` | Git ignore patterns | Standard exclusions |
| `.gitmodules` | Git submodule definitions | **24 submodules** (see Section 6) |
| `README.md` | **NOT FOUND** | No README at root level |

### Root Directories (12 total)
| Directory | Description |
|-----------|-------------|
| `cmd/` | Application entry points (4 sub-projects) |
| `docs/` | Research documentation, superpowers plans |
| `scripts/` | Build/automation scripts (4 scripts) |
| `specs/` | Feature specification (001-helixplay-system) |
| `vasic-digital/` | Catalogizer submodule mount point |
| `.claude/` | Claude Code IDE settings (settings.json with Stop hook) |
| `.opencode/` | OpenCode agent configuration (command/) |
| `.specify/` | Feature metadata, extensions, templates, workflows |
| `Upstreams/` | Multi-remote sync scripts (GitHub, GitLab, GitFlic, GitVerse) |

### Git Submodules (24 total - see Section 6 for full list)
Auth, Cache, Challenges, Concurrency, Containers, Database, Discovery, EventBus, Formatters, HelixQA, Media, Memory, Messaging, Middleware, Observability, Plugins, RAG, RateLimiter, Recovery, Security, Storage, Streaming, VectorDB, Catalogizer

---

## 2. CRITICAL FILES - FULL SUMMARIES

### 2.1 CLAUDE.md (v2.1.0) - AI Assistant Guidelines
**Location**: Root, 144 lines, 11.3 KB
**Source of Truth**: `docs/research/chapters/MVP/05_Response/01_Constitution.md` v2.1.0

**Key Content**:
- **NON-NEGOTIABLE PRIME DIRECTIVE**: Green tests MUST guarantee real end-user-usable behavior. No bluffing allowed.
- **Constitution v2.1.0 Amendments** (2026-05-01):
  1. Anti-bluff tests mandatory - forbids `assert.True(t, true)`, constructor-only tests, mock-only integration/E2E
  2. Usability evidence mandatory per §6.7 - HelixQA visual assertion, manual recording, or Challenge evidence
  3. Automatic negative-leg fault injection per §1.3/§6.3 - CI must break each feature and verify non-Unit tests fail
  4. `ValidateAntiBluff` gate unconditional; `CHALLENGE_ANTIBLUFF_STRICT` removed
  5. Container verifier `execCommand()` executes real commands (no-op eliminated)
- **Synthesis Programme**: `docs/research/chapters/MVP/05_Response/00_Master_Plan.md`
- **System Overview**: `docs/research/chapters/MVP/05_Response/02_System_Overview.md`
- **Repository State**: Both specs/research and active implementation. `cmd/` has Wails client, web client, core backend stubs, host agent stubs. 22+ Git submodules with own build systems.
- **User Mandate 2026-04-30**: DRY, KISS, Top 10 principles. Lazy initialization default. 100% coverage across 10 test types.
- **Build/Test Commands**: No root Makefile. Submodules use standardized Makefile. `go build ./cmd/...` from root.
- **Mandatory Constraints** (from `04_Request.md`): Anti-bluff, decoupling, 100% coverage (10 test types), container runtime, concurrency, quality gates (SonarQube, Snyk), tracking via GitHub Projects/GitLab
- **Git Topology**: 4 remotes (github, gitlab, gitverse, gitflic) - origin fetch=github, push=gitflic
- **Claude Code Config**: `.claude/settings.json` with Stop hook running `bash scripts/claim-check.sh`

### 2.2 AGENTS.md - Non-Claude Agent Configuration
**Location**: Root
**Content**: Same as CLAUDE.md but tailored for Codex, Cursor, Aider, Copilot, Cline, etc.

### 2.3 go.mod - Module Dependencies
**Go Version**: 1.26.2

```go
module github.com/HelixDevelopment/HelixPlay

require (
    github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory v0.0.0-00010101000000-000000000000
    github.com/andybalholm/brotli v1.2.1
    github.com/google/uuid v1.6.0
    github.com/stretchr/testify v1.11.1
)

replace github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory => ./vasic-digital/Memory
```

### 2.4 config.json - Application Configuration
**71 lines, 1.52 KB**

Configuration sections:
- **server**: localhost:28080, HTTPS enabled, 900s timeouts
- **database**: PostgreSQL on port 25432, db name "catalogizer" (also SQLite fallback)
- **auth**: JWT with 24h expiration
- **catalog**: 100 item default page size, 3 concurrent scans, 1MB chunks, 5GB max archive
- **storage**: Local/S3-compatible
- **logging**: JSON format to stdout
- **proxy**: Optional HTTP proxy

### 2.5 .claude/settings.json
```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "enabledPlugins": {
    "opsera-devsecops@claude-plugins-official": false
  },
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash scripts/claim-check.sh",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

---

## 3. SOURCE CODE STRUCTURE

### 3.1 cmd/ Directory (4 Application Entry Points)

```
cmd/
  client-wails/     # Wails Desktop Client
    backend/
      backend.go    # Backend struct with Start/Stop/Connect/IsRunning/IsConnected
    frontend/
      package.json  # Frontend dependencies
    main.go         # Entry point (creates backend, starts, stops)
    main_test.go    # 3 tests: Start, Stop, Connect

  client-web/
    tv/             # 10-foot TV UI - Leanback navigation, D-pad control

  core/             # Core Backend (discovery, protocol)

  host-agent/       # Sunshine-style Host Agent
    capability/     # GPU/codec capability advertisement
    capture/        # Screen capture (DX11, ScreenCaptureKit, PipeWire)
    codec/          # Codec ladder + capability negotiation
    discovery/      # mDNS Discovery Beacon
    encoder/        # Dual-path encoding (stream + record)
    game/           # Game enumeration (Steam/Epic/GOG/standalone)
    input/          # DualSense haptics, adaptive triggers, gyro, accel
    lifecycle/      # Game lifecycle manager (launch, terminate, quick resume)
    transport/      # ABR/FEC/SQP policies, frame pacing + VRR
```

### 3.2 Tech Stack

| Layer | Technology |
|-------|------------|
| **Primary Language** | Go 1.26.2 (root), Go 1.25+ (submodules) |
| **Desktop Client** | Wails (Go backend + JS/TS frontend) |
| **Web Client** | Go HTTP handlers (leanback/TV mode) |
| **Mobile Client** | Flutter + Go FFI (planned) |
| **Network** | gRPC preferred, REST + middleware microservices, HTTP/3 (QUIC/Cronet), WebRTC (Pion v4), UDP, NATS |
| **Databases** | CockroachDB (primary), PostgreSQL, Redis, RabbitMQ |
| **Compression** | Brotli |
| **Auth** | JWT (golang-jwt/jwt/v5), OAuth/OAuth2, API keys |
| **Containers** | Docker, Podman, Kubernetes |
| **QA** | Custom helixqa binary with OpenCV visual verification |
| **Codecs** | H.264 (fallback), HEVC (standard), AV1 (premium) |
| **Observability** | Prometheus client libraries |

### 3.3 Architecture Pattern
- **Microservices** with gRPC preferred
- **Git submodules** for reusable infrastructure (DRY principle)
- **Container-first** - everything runs in containers
- **Triple-stack convergence** - one Go core shared across Wails, Flutter, and Angular clients
- **Sunshine-style host agent** with capture, encode, input forwarding
- **Lazy initialization** (Constitution §5.2) - default pattern
- **Non-blocking concurrency** with semaphores/backpressure

---

## 4. EXISTING IMPLEMENTATION

### 4.1 Implemented Modules

| Module | Status | Description |
|--------|--------|-------------|
| **client-wails** | STUB | Backend struct with Start/Stop/Connect/IsRunning/IsConnected. Tests exist but are basic. |
| **host-agent/capability** | IMPLEMENTED | Game enumeration + capability advertisement |
| **host-agent/capture** | IMPLEMENTED | Screen capture (Windows DX11, Darwin ScreenCaptureKit, Linux PipeWire) |
| **host-agent/codec** | IMPLEMENTED | Codec ladder + capability negotiation (Task 3.3) |
| **host-agent/discovery** | IMPLEMENTED | Discovery Beacon with mDNS announcement (Task 2.2) |
| **host-agent/encoder** | IMPLEMENTED | Dual-path encoding (stream + record) (Task 3.4) |
| **host-agent/game** | IMPLEMENTED | Game enumeration (Steam/Epic/GOG/standalone) |
| **host-agent/input** | IMPLEMENTED | DualSense haptics, adaptive triggers, gyro, accel (Task 5.1) |
| **host-agent/lifecycle** | IMPLEMENTED | Game lifecycle manager (launch, terminate, quick resume) |
| **host-agent/transport** | IMPLEMENTED | ABR/FEC/SQP policies, frame pacing + VRR (Task 4.4) |
| **client-web/tv** | IMPLEMENTED | 10-foot TV UI with leanback navigation and D-pad control |

### 4.2 What's Missing/Incomplete

| Area | Issue |
|------|-------|
| **cmd/core/** | No visible Go source files found - likely STUB/empty |
| **cmd/client-wails** | Backend is a stub (just sets boolean flags). No actual Wails integration |
| **cmd/client-wails/frontend** | Only package.json, no actual frontend code visible |
| **README.md** | **DOES NOT EXIST** at root level |
| **Root Makefile** | Does not exist - no unified build at root |
| **Constitution file** | Not at root; located deep in docs hierarchy |
| **C09/C10 chapters** | Noted as missing in docs structure |
| **V1 phase** | Only placeholder directory exists |

### 4.3 Documentation Structure (Extensive)

```
docs/
  research/chapters/MVP/
    01_base/              # Architecture research (overall, Go backend, Wails/Flutter/Angular, WebRTC, NATS, CockroachDB)
      01_Request.md
      02_response/        # Agent-generated research
    02_latency/           # Zero-latency communication research
      01_Request.md
      02_Response/
    03_video_technology/  # Capture, codec, encode, recording, audio
      Request.md
      02_Response/
    04_Request.md         # **AUTHORITATIVE MVP BRIEF**
    05_Response/          # **Canonical synthesized docs**
      00_Master_Plan.md
      01_Constitution.md      # v2.1.0 - **SOURCE OF TRUTH**
      02_System_Overview.md   # Navigation hub
      03_Architecture/
      04_Latency/
      05_Video_Audio/
      06_Submodules/
      07_Testing/
      08_Operations/
      09_Implementation_Phases/
      99_Web_Research_Addenda/
  superpowers/
    plans/                # Superpowers planning docs
```

### 4.4 Feature Specification
- `specs/001-helixplay-system/spec.md` - Comprehensive system spec with 3 P1 user stories:
  1. Host Setup & Game Streaming
  2. Triple-Stack Client Convergence
  3. Controller Fidelity Over Network

---

## 5. BUILD/DEPLOYMENT CONFIGURATION

### 5.1 Root Module
- **No root-level Makefile**
- Build: `go build ./cmd/...` after submodules initialized
- No README.md

### 5.2 Scripts (4 scripts in scripts/)
| Script | Purpose |
|--------|---------|
| `anti-bluff-scan.sh` | CI lane that scans for bluff patterns (TODO, FIXME, assert.True(t,true), empty bodies, etc.) |
| `claim-check.sh` | Stop hook for R-18 Operational Integrity - checks forbidden commands (suspend, hibernate, shutdown, rm -rf /, etc.) |
| `propagate-constitution.sh` | Propagates Constitution to all submodules |
| `verify-submodules.py` | Verifies all submodules are correctly configured |

### 5.3 Submodule Build Pattern (Standardized)
Every vasic-digital submodule has identical Makefile targets:
```bash
make build              # go build ./...
make test               # go test -count=1 -race -p 1 ./...
make test-integration   # go test -count=1 -race -p 1 ./tests/integration/...
make test-bench         # go test -bench=. -benchmem ./tests/benchmark/...
make test-coverage      # coverage.out + coverage.html
make fmt                # gofmt -w . && goimports -w .
make vet                # go vet ./...
make lint               # golangci-lint run ./...
make challenge          # run challenge scripts
```

### 5.4 CI/CD
- **Container-driven** CI/CD (local, not cloud)
- Anti-bluff scan runs as non-overridable CI lane
- Tests MUST run inside containers per Constitution §3
- SonarQube + Snyk quality gates

### 5.5 Git Remotes (4-way sync)
```
github      git@github.com:HelixDevelopment/HelixPlay.git
gitlab      git@gitlab.com:helixdevelopment1/HelixPlay.git
gitverse    git@gitverse.ru:helixdevelopment/HelixPlay.git
gitflic     git@gitflic.ru:helixdevelopment/helixplay.git
origin      fetch=github, push=gitflic
```

---

## 6. ALL SUBMODULES (24 total)

### vasic-digital Organization (23 submodules)
| # | Name | Purpose | Commit |
|---|------|---------|--------|
| 1 | **Auth** | JWT, OAuth, API key, middleware auth | 2b5c28f |
| 2 | **Cache** | Caching abstractions | ad0fb83 |
| 3 | **Challenges** | Full-stack challenge runner & bluff scanner | e691cdf |
| 4 | **Concurrency** | Non-blocking concurrency primitives | 66b9a25 |
| 5 | **Containers** | Container orchestration, health checks, lifecycle | ecc0cd4 |
| 6 | **Database** | DB abstractions | 570a342 |
| 7 | **Discovery** | LAN service discovery, dynamic ports | 0e3616b |
| 8 | **EventBus** | Event propagation | 5605c9a |
| 9 | **Formatters** | Output formatting | 5d8db0f |
| 10 | **Media** | Media processing | d5172af |
| 11 | **Memory** | Memory management utilities | 1f4cac7 |
| 12 | **Messaging** | Message queue abstractions | 1bb506d |
| 13 | **Middleware** | HTTP/gRPC middleware | 60c95de |
| 14 | **Observability** | Metrics, logging, tracing | d4bd34e |
| 15 | **Plugins** | Plugin system | cc9479b |
| 16 | **RAG** | Retrieval-Augmented Generation support | 28c98dd |
| 17 | **RateLimiter** | Rate limiting | 8ae01d5 |
| 18 | **Recovery** | Fault recovery | 6d93011 |
| 19 | **Security** | Security utilities | 6ab452d |
| 20 | **Storage** | Storage abstractions | 077b1c3 |
| 21 | **Streaming** | Streaming protocols | 01d2126 |
| 22 | **VectorDB** | Vector database support | 09647d3 |
| 23 | **Catalogizer** | Asset catalogization | (in vasic-digital/) |

### HelixDevelopment Organization (1 submodule)
| # | Name | Purpose | Commit |
|---|------|---------|--------|
| 24 | **HelixQA** | Autonomous QA orchestration framework (OpenCV visual verification) | 8a5db17 |

---

## 7. PROJECT CORE PURPOSE AND VALUE PROPOSITION

### What is HelixPlay?
**HelixPlay is a cloud gaming platform** that turns any GPU-equipped machine into a remote gaming appliance, streaming console-class experience to any client device.

### Key Value Propositions:
1. **"Ultimate gaming experience!"** - Project tagline
2. **Any GPU machine becomes a gaming server** - Self-hostable, no cloud dependency
3. **Triple-stack client convergence** - One Go core shared across Wails (desktop), Flutter (mobile/TV), Angular (web)
4. **Zero perceived lag** - <=50ms WAN / <=30ms LAN p999 latency
5. **Full controller fidelity** - DualSense haptics, adaptive triggers, gyro, accelerometer, audio jack passthrough at 1kHz polling
6. **Fully open-source** - All code public under HelixDevelopment/vasic-digital organizations
7. **White-labellable** - Partners can rebrand
8. **Container-first** - Everything containerized, Kubernetes-ready

### Target Architecture:
- **Host Agent**: Sunshine-style capture/encode/stream running on gaming PC
- **Core Backend**: gRPC-based discovery, protocol negotiation
- **Clients**: Desktop (Wails), TV (Go HTTP leanback), Mobile (Flutter), Web (Angular)
- **Network**: WebRTC Pion v4, QUIC datagrams, custom UDP with DTLS 1.2
- **Codecs**: H.264 fallback, HEVC standard, AV1 premium - negotiated per-session
- **Input**: 1kHz USB polling, lock-free SPSC ring buffer, zero-copy IPC

### Current State (May 2026):
- **Documentation**: Extensively complete (MVP research chapters + synthesized 05_Response docs)
- **Host Agent**: 9 sub-modules implemented (capture, codec, discovery, encoder, game, input, lifecycle, transport, capability)
- **Client-Wails**: Stub implementation with basic backend structure
- **Client-Web/TV**: Implemented with leanback navigation
- **Core Backend**: Minimal/not visible
- **Constitution**: v2.1.0 fully documented with anti-bluff enforcement
- **Quality System**: 10 test types mandated, anti-bluff CI scanning, HelixQA integration

---

## 8. SIGNIFICANCE RATINGS SUMMARY

| Item | Rating | Notes |
|------|--------|-------|
| CLAUDE.md | 10/10 | Core AI guidance, prime directive, Constitution refs |
| AGENTS.md | 10/10 | Same content for non-Claude agents |
| Constitution (01_Constitution.md) | 10/10 | Source of truth, v2.1.0, 18 clauses |
| go.mod | 9/10 | Module graph, Go 1.26.2, minimal deps |
| config.json | 8/10 | Full app config, reveals architecture intent |
| .gitmodules | 8/10 | 24 submodules, full dependency map |
| Host Agent modules | 9/10 | 9 implemented sub-modules, most advanced area |
| Client-Wails | 5/10 | Stub only, needs significant work |
| Client-Web/TV | 6/10 | TV UI implemented |
| Core Backend | 3/10 | Not visible/empty |
| anti-bluff-scan.sh | 9/10 | Critical CI enforcement script |
| docs/ | 10/10 | Extensive research documentation |
| specs/ | 8/10 | Feature specification with user stories |
| README.md | 0/10 | **MISSING** |
| Root Makefile | 0/10 | **MISSING** |
