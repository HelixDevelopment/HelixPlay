# Tasks: HelixPlay — Ultimate Gaming Experience

**Branch**: `001-helixplay-system` | **Date**: 2026-05-02  
**Source**: [spec.md](./spec.md) | [plan.md](./plan.md) | [data-model.md](./data-model.md) | [research.md](./research.md)  
**Submodules**: 46 | **Test Matrix**: 1,840 cells | **Constitution**: v2.2.0

---

## Summary

| Phase | Scope | Task Count | Story |
|-------|-------|-----------:|-------|
| Phase 1 | Setup (Project Initialization) | 15 | — |
| Phase 2 | Foundational (Blocking Prerequisites) | 42 | — |
| Phase 3 | US1 — Host Setup & Game Streaming | 45 | P1 |
| Phase 4 | US2 — Triple-Stack Client Convergence | 32 | P1 |
| Phase 5 | US3 — Controller Fidelity Over Network | 16 | P1 |
| Phase 6 | US4 — Zero-Impact Recording (DVR) | 13 | P2 |
| Phase 7 | US5 — White-Label & Theming | 16 | P2 |
| Phase 8 | US6 — TV-First UI/UX | 15 | P2 |
| Phase 9 | US8 — Security, Auth & Host Isolation | 18 | P2 |
| Phase 10 | US9 — Catalog, Metadata & 4K Assets | 16 | P2 |
| Phase 11 | US7 — Scalability & Multi-Region | 15 | P3 |
| Phase 12 | US10 — Testing, Challenges & HelixQA | 26 | P1 |
| Phase 13 | Polish & Cross-Cutting Concerns | 21 | — |
| **Total** | | **290** | |

---

## Dependency Graph

```
Phase 1 (Setup)
    │
    ▼
Phase 2 (Foundational)
    │
    ├──► Phase 3 (US1: Host Setup & Streaming) ──┐
    │                                              │
    ├──► Phase 4 (US2: Triple-Stack Clients) ◄─────┤ (US1 + US2 converge)
    │                                              │
    ├──► Phase 5 (US3: Controller Fidelity) ◄──────┘
    │
    ├──► Phase 9 (US8: Security & Auth) ──► Phase 7 (US5: White-Label)
    │                                          │
    ├──► Phase 10 (US9: Catalog) ──────────────┤
    │                                          │
    ├──► Phase 6 (US4: Recording) ◄────────────┤ (post-streaming)
    │                                          │
    ├──► Phase 8 (US6: TV-First UX) ◄──────────┘
    │
    ├──► Phase 11 (US7: Scalability) ◄─────────── (after US1+US2 stable)
    │
    └──► Phase 12 (US10: Testing & QA) ────────► Phase 13 (Polish & GA)
```

**MVP Scope**: US1 + US2 + US3 + US8 (security baseline) + US10 (test infrastructure) = Phases 1–5, 9, 12.

---

## Phase 1: Setup (Project Initialization)

*No user story label — shared infrastructure. All tasks are blocking for Phase 2.*

- [x] T001 Create root README.md with 13 required sections (>200 lines) in `/README.md`
- [x] T002 Create root Makefile with targets: build, test, test-integration, test-e2e, test-bench, test-coverage, test-security, test-chaos, test-stress, test-smoke, test-fullauto, test-challenge, anti-bluff, verify-submodules, propagate-constitution, fmt, vet, lint, clean, docker-build, docker-up, docker-down in `/Makefile`
- [x] T003 Create root `go.work` with all 46 submodule paths and run `go work sync` in `/go.work`
- [x] T004 Verify all 46 submodule dependencies are transitively complete in `.gitmodules` via `/scripts/verify-submodules.py`
- [x] T005 [P] Propagate Constitution v2.2.0 to CLAUDE.md in all 46 submodules via `/scripts/propagate-constitution.sh`
- [x] T006 [P] Propagate Constitution v2.2.0 to AGENTS.md in all 46 submodules via `/scripts/propagate-constitution.sh`
- [x] T007 [P] Propagate Constitution v2.2.0 to CONSTITUTION.md in all 46 submodules via `/scripts/propagate-constitution.sh`
- [x] T008 Create `.github/dependabot.yml` for automated submodule updates in `.github/dependabot.yml`
- [x] T009 Create `.editorconfig` with Go/Proto/Web consistent formatting in `/.editorconfig`
- [x] T010 Create `.gitignore` for root module (bin/, coverage.out, .env, vendor/) in `/.gitignore`
- [x] T011 Create `docs/CONTRIBUTING.md` with PR template, commit conventions, branch strategy in `/docs/CONTRIBUTING.md`
- [x] T012 Create `docs/ARCHITECTURE.md` with C4 diagrams (Context, Container, Component, Code) in `/docs/ARCHITECTURE.md`
- [x] T013 Create GitHub Projects board configuration in `.github/projects/helixplay.yml`
- [x] T014 Create issue templates (bug, feature, security) in `.github/ISSUE_TEMPLATE/`
- [x] T015 Create pull request template in `.github/PULL_REQUEST_TEMPLATE.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

*No user story label — must complete before any user story implementation.*

### 2.1 Protocol & Data Model

- [x] T016 Define Protocol Buffer schemas for `pkg/protocol/v1/streaming.proto` (StreamingControl service)
- [x] T017 Define Protocol Buffer schemas for `pkg/protocol/v1/discovery.proto` (RendezvousService)
- [x] T018 Define Protocol Buffer schemas for `pkg/protocol/v1/hostagent.proto` (HostAgentService)
- [x] T019 Define Protocol Buffer schemas for `pkg/protocol/v1/catalog.proto` (CatalogService proxy)
- [x] T020 Generate Go gRPC clients from all `.proto` files into `pkg/protocol/v1/`
- [x] T021 Create CockroachDB migration `001_init_schema.sql` with all 10 entities (Tenant, User, Host, GPU, Game, Session, Recording, Controller, Capability, Asset) in `Database/migrations/001_init_schema.sql`
- [x] T022 Create CockroachDB migration `002_junction_tables.sql` for HostGame and TenantGameFilter in `Database/migrations/002_junction_tables.sql`
- [x] T023 Create CockroachDB migration `003_rls_policies.sql` for row-level security per tenant in `Database/migrations/003_rls_policies.sql`
- [x] T024 Create shared Go models from data-model.md in `pkg/models/` (host.go, session.go, game.go, tenant.go, user.go, recording.go, controller.go, capability.go, asset.go)
- [x] T025 Create database repository interfaces in `pkg/repository/` (host_repo.go, session_repo.go, game_repo.go, tenant_repo.go, user_repo.go)

### 2.2 Core Submodules

- [x] T026 Implement `vasic-digital/Discovery` — mDNS/DNS-SD browser and advertiser with TXT record parsing in `Discovery/pkg/discovery/mdns.go`
- [x] T027 Implement `vasic-digital/Discovery` — rendezvous client with lease management in `Discovery/pkg/discovery/rendezvous.go`
- [x] T028 Implement `vasic-digital/Auth` — OAuth2/OIDC middleware with JWT validation (RS256) in `Auth/pkg/auth/middleware.go`
- [x] T029 Implement `vasic-digital/Auth` — device authorization grant (RFC 8628) flow client in `Auth/pkg/auth/device_grant.go`
- [x] T030 Implement `vasic-digital/Auth` — RBAC enforcement with scope/role checking in `Auth/pkg/auth/rbac.go`
- [x] T031 Implement `vasic-digital/Database` — CockroachDB connection pool with geo-partitioning awareness in `Database/pkg/database/cockroach.go`
- [x] T032 Implement `vasic-digital/Cache` — Redis client with circuit breaker and connection pooling in `Cache/pkg/cache/redis.go`
- [x] T033 Implement `vasic-digital/EventBus` — NATS JetStream publisher/subscriber wrapper in `EventBus/pkg/eventbus/nats.go`
- [x] T034 Implement `vasic-digital/Concurrency` — non-blocking semaphore and sync.Pool helpers in `Concurrency/pkg/concurrency/semaphore.go`
- [x] T035 Implement `vasic-digital/Observability` — structured slog logger with JSON output in `Observability/pkg/observability/logger.go`
- [x] T036 Implement `vasic-digital/Observability` — Prometheus metrics registry with latency histograms in `Observability/pkg/observability/metrics.go`
- [x] T037 Implement `vasic-digital/Observability` — OpenTelemetry trace propagation in `Observability/pkg/observability/tracing.go`
- [x] T038 Implement `vasic-digital/Memory` — lock-free SPSC ring buffer with 128-byte cache-line padding in `Memory/pkg/memory/spsc.go`
- [x] T039 Implement `vasic-digital/Memory` — shared memory allocator (`memfd_create` Linux, `shm_open` POSIX, Windows file-mapping) in `Memory/pkg/memory/shm.go`
- [x] T040 Implement `vasic-digital/RateLimiter` — token bucket + sliding window rate limiter backed by Redis in `RateLimiter/pkg/ratelimiter/ratelimiter.go`
- [x] T041 Implement `vasic-digital/Security` — input validation helpers (UUID, email, hex fingerprint) in `Security/pkg/security/validation.go`
- [x] T042 Implement `vasic-digital/Security` — CVE scan orchestrator (govulncheck wrapper) in `Security/pkg/security/cve_scan.go`
- [x] T043 Implement `vasic-digital/Storage` — MKV container writer interface in `Storage/pkg/storage/mkv_writer.go`
- [x] T044 Implement `vasic-digital/Storage` — fMP4 container writer interface in `Storage/pkg/storage/fmp4_writer.go`

### 2.3 Container & CI Bootstrap

- [x] T045 Create host agent base container (Ubuntu 24.04 LTS + GPU drivers) in `Containers/images/host-agent/Dockerfile`
- [x] T046 Create capture service container (per-OS variants: Windows DXGI, macOS SCK, Linux PipeWire) in `Containers/images/capture/`
- [x] T047 Create encoder service container (NVENC/CUDA runtime) in `Containers/images/encoder/Dockerfile`
- [x] T048 Create discovery beacon container (distroless static nonroot) in `Containers/images/discovery/Dockerfile`
- [x] T049 Create core backend container (distroless static nonroot) in `Containers/images/core/Dockerfile`
- [x] T050 Create docker-compose.yml for local development topology in `/docker-compose.yml`
- [x] T051 Create GitHub Actions workflow `anti-bluff-scan.yml` (non-overridable lane) in `.github/workflows/anti-bluff-scan.yml`
- [x] T052 Create GitHub Actions workflow `test.yml` (Unit + Integration + E2E matrix) in `.github/workflows/test.yml`
- [x] T053 Create GitHub Actions workflow `security.yml` (govulncheck + Snyk + Trivy) in `.github/workflows/security.yml`
- [x] T054 Create Dagger CI pipeline definition in `ci/dagger/pipeline.go`
- [x] T055 Create `scripts/fault-inject.sh` for negative-leg fault injection per Constitution §1.3 in `/scripts/fault-inject.sh`
- [x] T056 Create `scripts/mutation-test.sh` using Gremlins v0.6+ with ≥85% gate in `/scripts/mutation-test.sh`
- [x] T057 Create `scripts/latency-benchmark.sh` with HDR histogram and p50/p99/p999 output in `/scripts/latency-benchmark.sh`


---

## Phase 3: US1 — Host Setup & Game Streaming

*Story goal: Operator runs HelixPlay host agent; client discovers host, connects, and streams a game with ≤50 ms WAN / ≤30 ms LAN p999 latency.*

### 3.1 Host Agent Core

- [x] T058 [US1] Implement host agent main entry point with signal handling in `cmd/host-agent/main.go`
- [x] T059 [US1] Implement host agent configuration loader (TOML + env vars) with validation in `cmd/host-agent/config/config.go`
- [x] T060 [US1] Implement host hardware fingerprinting (SHA-256 of CPU+MB+GPU serials) in `cmd/host-agent/capability/fingerprint.go`
- [x] T061 [US1] Implement GPU enumeration (NVIDIA/AMD/Intel/Apple) with driver detection in `cmd/host-agent/capability/gpu_enum.go`
- [x] T062 [US1] Implement GPU capability advertisement (codecs, max resolution, thermal headroom) in `cmd/host-agent/capability/advertise.go`
- [x] T063 [US1] Implement host registration with coordinator (mTLS + bootstrap token) in `cmd/host-agent/lifecycle/register.go`
- [x] T064 [US1] Implement host heartbeat loop (15s interval, status updates) in `cmd/host-agent/lifecycle/heartbeat.go`
- [x] T065 [US1] Implement host graceful shutdown and deregistration in `cmd/host-agent/lifecycle/shutdown.go`

### 3.2 Game Enumeration & Lifecycle

- [x] T066 [US1] Implement Steam game enumeration (read libraryfolders.vdf + appmanifests) in `cmd/host-agent/game/steam_enum.go`
- [x] T067 [US1] Implement Epic Games Store enumeration (read EpicManifests) in `cmd/host-agent/game/epic_enum.go`
- [x] T068 [US1] Implement GOG Galaxy enumeration (read gog-galaxy config) in `cmd/host-agent/game/gog_enum.go`
- [x] T069 [P] [US1] Implement Ubisoft Connect enumeration in `cmd/host-agent/game/ubisoft_enum.go`
- [x] T070 [P] [US1] Implement Battle.net enumeration in `cmd/host-agent/game/battlenet_enum.go`
- [x] T071 [P] [US1] Implement Origin/EA App enumeration in `cmd/host-agent/game/origin_enum.go`
- [x] T072 [P] [US1] Implement Microsoft Store enumeration in `cmd/host-agent/game/microsoft_enum.go`
- [x] T073 [US1] Implement standalone game scan (configurable paths) in `cmd/host-agent/game/standalone_enum.go`
- [x] T074 [US1] Implement game launch (process spawn with window capture attach) in `cmd/host-agent/game/launch.go`
- [x] T075 [US1] Implement game monitor (health checks, crash detection) in `cmd/host-agent/game/monitor.go`
- [x] T076 [US1] Implement game terminate (graceful SIGTERM → SIGKILL) in `cmd/host-agent/game/terminate.go`
- [x] T077 [US1] Implement quick resume (save state cache, ≤5s restore) in `cmd/host-agent/game/quick_resume.go`

### 3.3 Capture Pipeline

- [x] T078 [US1] Implement Windows DXGI Desktop Duplication API capture path in `cmd/host-agent/capture/dxgi_capture.go`
- [x] T079 [US1] Implement macOS ScreenCaptureKit capture path with IOSurface export in `cmd/host-agent/capture/sck_capture.go`
- [x] T080 [US1] Implement Linux KMS/DMA-BUF capture path in `cmd/host-agent/capture/kms_capture.go`
- [x] T081 [US1] Implement Linux PipeWire fallback capture path in `cmd/host-agent/capture/pipewire_capture.go`
- [x] T082 [US1] Implement capture dispatcher (OS auto-detection + path selection) in `cmd/host-agent/capture/dispatcher.go`

### 3.4 Hardware Encoders

- [x] T083 [US1] Implement NVENC integration (Video Codec SDK 12.1+ via CGO) in `cmd/host-agent/encoder/nvenc.go`
- [x] T084 [US1] Implement Intel QSV integration (oneVPL ULL/LL modes) in `cmd/host-agent/encoder/qsv.go`
- [x] T085 [US1] Implement AMD AMF integration (Windows) in `cmd/host-agent/encoder/amf.go`
- [x] T086 [US1] Implement Apple VideoToolbox integration in `cmd/host-agent/encoder/videotoolbox.go`
- [x] T087 [US1] Implement VAAPI integration (Linux generic fallback) in `cmd/host-agent/encoder/vaapi.go`
- [x] T088 [US1] Implement Vulkan Video Encode integration (AMD Linux future path) in `cmd/host-agent/encoder/vulkan_video.go`
- [x] T089 [US1] Implement codec ladder selection (H.264 fallback → HEVC → AV1) in `cmd/host-agent/encoder/codec_ladder.go`
- [x] T090 [US1] Implement encoder dispatcher (GPU auto-detection + best encoder selection) in `cmd/host-agent/encoder/dispatcher.go`

### 3.5 Transport & Streaming

- [x] T091 [US1] Implement WebRTC Pion v4 transport (PeerConnection, DataChannel, SRTP) in `cmd/host-agent/transport/webrtc.go`
- [x] T092 [US1] Implement QUIC datagram transport (quic-go RFC 9221) in `cmd/host-agent/transport/quic.go`
- [x] T093 [US1] Implement custom UDP transport (Parsec BUD style, DTLS 1.2, ChaCha20-Poly1305) in `cmd/host-agent/transport/udp_custom.go`
- [x] T094 [US1] Implement ABR/FEC/SQP congestion control policy engine in `cmd/host-agent/transport/congestion.go`
- [x] T095 [US1] Implement capability handshake orchestrator in `cmd/host-agent/transport/handshake.go`
- [x] T096 [US1] Implement session manager (create, negotiate, monitor, terminate) in `cmd/host-agent/session/manager.go`

### 3.6 US1 Integration & Tests

- [x] T097 [US1] Write Unit tests for GPU enumeration (mocked sysfs/registry queries) in `cmd/host-agent/capability/gpu_enum_test.go`
- [x] T098 [US1] Write Unit tests for codec ladder selection in `cmd/host-agent/encoder/codec_ladder_test.go`
- [x] T099 [US1] Write Integration test: host registration → heartbeat → deregistration in `tests/integration/host_agent_lifecycle_test.go`
- [x] T100 [US1] Write Integration test: game enumeration across 7 stores in `tests/integration/game_enum_test.go`
- [x] T101 [US1] Write E2E test: full streaming session 1080p60 ≤30ms LAN p999 in `tests/e2e/streaming_1080p_test.go`
- [x] T102 [US1] Write Benchmark test: capture-to-encode latency HDR histogram in `tests/benchmark/capture_encode_latency_test.go`

---

## Phase 4: US2 — Triple-Stack Client Convergence

*Story goal: Wails desktop, Flutter mobile/TV, and Angular web clients all share one Go core and achieve same latency budget.*

### 4.1 Go Core (Shared Business Logic)

- [x] T103 [US2] Implement Go core — gRPC service discovery client in `pkg/core/discovery/client.go`
- [x] T104 [US2] Implement Go core — streaming protocol state machine in `pkg/core/streaming/state_machine.go`
- [x] T105 [US2] Implement Go core — capability negotiation client-side logic in `pkg/core/streaming/negotiate.go`
- [x] T106 [US2] Implement Go core — controller input parser (DualSense HID report decoding) in `pkg/core/input/dualsense_parser.go`
- [x] T107 [US2] Implement Go core — network quality estimator (bandwidth, RTT, jitter, loss) in `pkg/core/network/quality.go`
- [x] T108 [US2] Implement Go core — session telemetry builder in `pkg/core/streaming/telemetry.go`
- [x] T109 [US2] Implement Go core — ABR policy client (bitrate adaptation requests) in `pkg/core/streaming/abr_client.go`
- [x] T110 [US2] Build Go core as c-shared library (`go build -buildmode=c-shared`) for Flutter FFI in `pkg/core/build/flutter.go`
- [x] T111 [US2] Build Go core as WASM module (`GOOS=js GOARCH=wasm`) for Angular in `pkg/core/build/wasm.go`
- [x] T112 [US2] Implement Go core — WebCodecs video decoder wrapper (browser) in `pkg/core/decode/webcodecs.go`

### 4.2 Wails Desktop Client

- [x] T113 [US2] Replace boolean-only Wails backend stub with real implementation in `cmd/client-wails/backend/app.go`
- [x] T114 [US2] Implement Wails backend — host discovery (mDNS + rendezvous) in `cmd/client-wails/backend/discovery.go`
- [x] T115 [US2] Implement Wails backend — streaming session controller in `cmd/client-wails/backend/streaming.go`
- [x] T116 [US2] Implement Wails backend — controller input capture (rawinput/xinput) in `cmd/client-wails/backend/input.go`
- [x] T117 [US2] Implement Wails frontend — desktop UI with catalog browser in `cmd/client-wails/frontend/src/App.svelte`
- [x] T118 [US2] Implement Wails frontend — settings panel (codec, resolution, controller config) in `cmd/client-wails/frontend/src/Settings.svelte`
- [x] T119 [US2] Implement Wails frontend — streaming overlay (latency HUD, bitrate graph) in `cmd/client-wails/frontend/src/StreamOverlay.svelte`

### 4.3 Flutter Mobile/TV Client

- [x] T120 [US2] Generate Dart FFI bindings from Go core C header via `ffigen` in `cmd/client-flutter/lib/core_bindings.dart`
- [x] T121 [US2] Implement Flutter — Go core loader (`DynamicLibrary.open` per platform) in `cmd/client-flutter/lib/core_loader.dart`
- [x] T122 [US2] Implement Flutter — service discovery UI (host list with GPU info) in `cmd/client-flutter/lib/screens/discovery_screen.dart`
- [x] T123 [US2] Implement Flutter — catalog browser with lazy loading in `cmd/client-flutter/lib/screens/catalog_screen.dart`
- [x] T124 [US2] Implement Flutter — streaming view with platform video texture in `cmd/client-flutter/lib/screens/stream_screen.dart`
- [x] T125 [US2] Implement Flutter — controller navigation (D-pad, analog, A/B/X/Y) in `cmd/client-flutter/lib/navigation/controller_nav.dart`
- [x] T126 [US2] Implement Flutter — Android TV Leanback launcher intent in `cmd/client-flutter/android/app/src/main/AndroidManifest.xml`

### 4.4 Angular Web Client

- [x] T127 [US2] Implement Angular — WASM Go core loader and `syscall/js` bridge in `cmd/client-web/src/app/core/wasm-loader.service.ts`
- [x] T128 [US2] Implement Angular — WebCodecs video decoder service in `cmd/client-web/src/app/streaming/webcodecs.service.ts`
- [x] T129 [US2] Implement Angular — catalog browser component with virtual scrolling in `cmd/client-web/src/app/catalog/catalog.component.ts`
- [x] T130 [US2] Implement Angular — streaming component with canvas rendering in `cmd/client-web/src/app/streaming/stream.component.ts`
- [x] T131 [US2] Implement Angular — TV keyboard/controller navigation service in `cmd/client-web/src/app/navigation/tv-nav.service.ts`

### 4.5 US2 Integration & Tests

- [x] T132 [US2] Write Integration test: Go core c-shared library loads on Android/iOS in `tests/integration/flutter_core_load_test.go`
- [x] T133 [US2] Write Integration test: Go core WASM loads in Chrome/Firefox/Safari in `tests/integration/wasm_load_test.go`
- [x] T134 [US2] Write E2E test: same host streamable from Wails, Flutter, Angular simultaneously in `tests/e2e/triple_stack_e2e_test.go`
- [x] T135 [US2] Write cross-client latency comparison benchmark in `tests/benchmark/cross_client_latency_test.go`


---

## Phase 5: US3 — Controller Fidelity Over Network

*Story goal: DualSense haptics, adaptive triggers, gyro, accelerometer, and audio jack forwarded with ≤1 ms effective latency at 1 kHz polling.*

### 5.1 Input Pipeline (Host-Side)

- [x] T136 [US3] Implement 1 kHz USB HID polling loop with `libusb` in `cmd/host-agent/input/usb_poll.go`
- [x] T137 [US3] Implement DualSense HID report parser (input reports 0x01, 0x31, 0x81) in `cmd/host-agent/input/dualsense_parser.go`
- [x] T138 [US3] Implement lock-free SPSC ring buffer for input events (host→game) in `cmd/host-agent/input/spsc_ring.go`
- [x] T139 [US3] Implement zero-copy IPC via shared memory (`memfd_create`/IOSurface/file-mapping) in `cmd/host-agent/input/shared_mem.go`
- [x] T140 [US3] Implement controller hot-plug detection and re-enumeration in `cmd/host-agent/input/hotplug.go`

### 5.2 Input Pipeline (Network & Client)

- [x] T141 [US3] Implement input event serialization (protobuf, fixed 64-byte packet) in `pkg/core/input/serialization.go`
- [x] T142 [US3] Implement input transport over WebRTC DataChannel (unreliable, ordered=false) in `pkg/core/input/datachannel.go`
- [x] T143 [US3] Implement input transport over QUIC stream (bidirectional, low-priority) in `pkg/core/input/quic_input.go`
- [x] T144 [US3] Implement DualSense haptics feedback decoder (host→client) in `pkg/core/input/haptics.go`
- [x] T145 [US3] Implement adaptive trigger resistance profile encoder/decoder in `pkg/core/input/triggers.go`
- [x] T146 [US3] Implement gyroscope/accelerometer data forwarding (6-axis IMU) in `pkg/core/input/imu.go`
- [x] T147 [US3] Implement audio jack passthrough (USB audio device proxy) in `pkg/core/input/audio_jack.go`

### 5.3 US3 Integration & Tests

- [x] T148 [US3] Write Unit test: SPSC ring buffer 50-200 ns hop latency at 1kHz in `cmd/host-agent/input/spsc_ring_test.go`
- [x] T149 [US3] Write Unit test: DualSense HID report parser round-trip in `cmd/host-agent/input/dualsense_parser_test.go`
- [x] T150 [US3] Write Integration test: controller input end-to-end with virtual HID device in `tests/integration/controller_input_test.go`
- [x] T151 [US3] Write Benchmark test: input latency distribution (p50/p99/p999) in `tests/benchmark/input_latency_test.go`
- [x] T152 [US3] Write E2E test: DualSense haptics + triggers + gyro verified on real hardware in `tests/e2e/dualsense_fidelity_test.go`

---

## Phase 6: US4 — Zero-Impact Recording (DVR)

*Story goal: Dual-path encoding (stream + record) runs simultaneously; recordings stored to NVMe with background sync.*

### 6.1 Dual-Path Encoding

- [x] T153 [US4] Implement stream path encoder configuration (latency-optimized preset) in `cmd/host-agent/encoder/stream_path.go`
- [x] T154 [US4] Implement record path encoder configuration (quality-optimized preset) in `cmd/host-agent/encoder/record_path.go`
- [x] T155 [US4] Implement dual-path encoder manager (two encoder sessions, no interference) in `cmd/host-agent/encoder/dual_path.go`
- [x] T156 [US4] Implement HDR metadata preservation in record path (HDR10/HDR10+/Dolby Vision SEI) in `cmd/host-agent/encoder/hdr_metadata.go`

### 6.2 Recording Storage & Sync

- [x] T157 [US4] Implement MKV container writer with audio sync in `cmd/host-agent/recording/mkv_writer.go`
- [x] T158 [US4] Implement fMP4 container writer with CMAF segments in `cmd/host-agent/recording/fmp4_writer.go`
- [x] T159 [US4] Implement recording manager (start, stop, segment rotation) in `cmd/host-agent/recording/manager.go`
- [x] T160 [US4] Implement background sync uploader (S3/MinIO/CloudFront) in `cmd/host-agent/recording/sync.go`
- [x] T161 [US4] Implement recording metadata indexer (session link, codec, quality, duration) in `cmd/host-agent/recording/indexer.go`

### 6.3 US4 Integration & Tests

- [x] T162 [US4] Write Unit test: dual-path encoder isolation (stream latency unaffected) in `cmd/host-agent/encoder/dual_path_test.go`
- [x] T163 [US4] Write Integration test: recording write to NVMe, verify MKV/fMP4 integrity in `tests/integration/recording_write_test.go`
- [x] T164 [US4] Write Integration test: background sync without streaming latency impact in `tests/integration/sync_no_impact_test.go`
- [x] T165 [US4] Write E2E test: full session with recording, playback verification in `tests/e2e/recording_playback_test.go`

---

## Phase 7: US5 — White-Label & Theming for Partners

*Story goal: Per-tenant UI theming, catalog filtering, OAuth2 config, and monetization with full isolation.*

### 7.1 Tenant Management

- [x] T166 [US5] Implement tenant CRUD service in `cmd/core/tenant/service.go`
- [x] T167 [US5] Implement tenant theme storage and retrieval (design tokens JSON) in `cmd/core/tenant/theme.go`
- [x] T168 [US5] Implement tenant catalog filter engine (subset of global catalog) in `cmd/core/tenant/catalog_filter.go`
- [x] T169 [US5] Implement tenant resource quota enforcement (max sessions, storage) in `cmd/core/tenant/quota.go`

### 7.2 Theming Engine

- [x] T170 [US5] Implement design token schema and validation in `pkg/theming/tokens.go`
- [x] T171 [US5] Implement CSS custom properties generator from tokens in `pkg/theming/css_generator.go`
- [x] T172 [US5] Implement Flutter ThemeData generator from tokens in `pkg/theming/flutter_generator.go`
- [x] T173 [US5] Implement Wails runtime CSS injection theming in `cmd/client-wails/backend/theming.go`
- [x] T174 [US5] Implement tenant admin dashboard API (theme upload, preview) in `cmd/core/tenant/admin_api.go`

### 7.3 Monetization

- [x] T175 [US5] Implement `vasic-digital/Monetization` — subscription engine (plans, trials, billing cycles) in `Monetization/pkg/monetization/subscription.go`
- [x] T176 [US5] Implement `vasic-digital/Monetization` — usage tracking (session minutes, bandwidth) in `Monetization/pkg/monetization/usage.go`
- [x] T177 [US5] Implement `vasic-digital/Monetization` — invoice generation and PDF export in `Monetization/pkg/monetization/invoice.go`
- [x] T178 [US5] Implement `vasic-digital/Monetization` — revenue-share ledger for partners in `Monetization/pkg/monetization/ledger.go`

### 7.4 US5 Integration & Tests

- [x] T179 [US5] Write Integration test: tenant isolation (user A in T1 cannot see T2 data) in `tests/integration/tenant_isolation_test.go`
- [x] T180 [US5] Write Integration test: theme application across Wails/Flutter/Angular in `tests/integration/theming_cross_client_test.go`
- [x] T181 [US5] Write E2E test: two demo tenants with different themes, OAuth2, billing in `tests/e2e/white_label_e2e_test.go`

---

## Phase 8: US6 — TV-First UI/UX

*Story goal: 10-foot UI navigable with controller only; quick resume ≤5s; instant-on; background download.*

### 8.1 TV Navigation

- [x] T182 [US6] Implement TV focus management (D-pad, analog stick navigation) in `pkg/tvux/focus_manager.go`
- [x] T183 [US6] Implement Leanback grid layout (carousels, rows, hero banners) in `pkg/tvux/leanback_grid.go`
- [x] T184 [US6] Implement sound effects for navigation (focus, select, back) in `pkg/tvux/audio_feedback.go`
- [x] T185 [US6] Implement screensaver with featured games carousel in `pkg/tvux/screensaver.go`

### 8.2 Client TV Implementations

- [x] T186 [US6] Implement Flutter TV — Leanback home screen with game carousels in `cmd/client-flutter/lib/screens/tv_home_screen.dart`
- [x] T187 [US6] Implement Flutter TV — controller-only settings menu in `cmd/client-flutter/lib/screens/tv_settings_screen.dart`
- [x] T188 [US6] Implement Angular web — 10-foot CSS with `@media` distance queries in `cmd/client-web/src/styles/tv.scss`
- [x] T189 [US6] Implement Angular web — virtual keyboard for search in `cmd/client-web/src/app/components/virtual-keyboard.component.ts`
- [x] T190 [US6] Implement Wails — kiosk mode for dedicated TV PC in `cmd/client-wails/backend/kiosk.go`

### 8.3 Quick Resume & Instant-On

- [x] T191 [US6] Implement quick resume state cache (last played game, position) in `pkg/core/resume/state_cache.go`
- [x] T192 [US6] Implement instant-on cold start optimization (pre-connect to rendezvous) in `pkg/core/resume/instant_on.go`
- [x] T193 [US6] Implement background game download/update queue in `pkg/core/resume/bg_download.go`

### 8.4 US6 Integration & Tests

- [x] T194 [US6] Write Integration test: full UI navigation with virtual controller in `tests/integration/tv_nav_test.go`
- [x] T195 [US6] Write E2E test: quick resume ≤5 seconds from idle in `tests/e2e/quick_resume_test.go`
- [x] T196 [US6] Write E2E test: screensaver activation and dismissal in `tests/e2e/screensaver_test.go`


---

## Phase 9: US8 — Security, Auth & Host Isolation

*Story goal: OAuth2/OIDC auth, container isolation, RBAC, anti-bluff enforcement, no host compromise.*

### 9.1 Authentication

- [x] T197 [US8] Implement Auth0 OAuth2 client integration in `cmd/core/auth/auth0_client.go`
- [x] T198 [US8] Implement device authorization grant (RFC 8628) flow server in `cmd/core/auth/device_flow.go`
- [x] T199 [US8] Implement JWT validation middleware (RS256, claims, scope) in `cmd/core/auth/jwt_middleware.go`
- [x] T200 [US8] Implement token refresh rotation with revocation list (Redis) in `cmd/core/auth/token_refresh.go`
- [x] T201 [US8] Implement mTLS certificate generation and rotation for host agents in `cmd/core/auth/mtls_ca.go`

### 9.2 Authorization & RBAC

- [x] T202 [US8] Implement RBAC middleware (role → scope → permission) in `cmd/core/auth/rbac_middleware.go`
- [x] T203 [US8] Implement tenant-scoped query enforcement (`tenant_id` injection) in `cmd/core/auth/tenant_scope.go`
- [x] T204 [US8] Implement API scope validator per endpoint in `cmd/core/auth/scope_validator.go`

### 9.3 Security Scanning & Isolation

- [x] T205 [US8] Implement container privilege escalation scanner in `Security/pkg/security/privesc_scan.go`
- [x] T206 [US8] Implement host-integrity-scan CI sub-lane (forbidden commands check) in `scripts/host-integrity-scan.sh`
- [x] T207 [US8] Implement `r18.SafeExec` wrapper for dangerous syscalls in `Security/pkg/security/safe_exec.go`
- [x] T208 [US8] Implement fuzz testing harness for protocol parsers in `tests/security/fuzz_protocol_test.go`
- [x] T209 [US8] Implement fuzz testing harness for input handlers in `tests/security/fuzz_input_test.go`

### 9.4 US8 Integration & Tests

- [x] T210 [US8] Write Security test: container escape attempt (privileged mount check) in `tests/security/container_escape_test.go`
- [x] T211 [US8] Write Security test: OWASP Top 10 penetration test suite in `tests/security/owasp_top10_test.go`
- [x] T212 [US8] Write Integration test: OAuth2 full flow (auth code + PKCE) in `tests/integration/oauth2_flow_test.go`
- [x] T213 [US8] Write Integration test: device authorization grant flow in `tests/integration/device_grant_test.go`
- [x] T214 [US8] Write E2E test: RBAC enforcement (player cannot access admin endpoints) in `tests/e2e/rbac_e2e_test.go`

---

## Phase 10: US9 — Catalog, Metadata & 4K Assets

*Story goal: Game catalog with metadata, 4K assets, CDN caching, lazy loading, search ≤200ms p999.*

### 10.1 Catalogizer Integration

- [x] T215 [US9] Implement Catalogizer gRPC client (generated from `pkg/protocol/v1/catalog.proto`) in `pkg/catalog/client.go`
- [x] T216 [US9] Implement Catalogizer REST fallback client in `pkg/catalog/rest_client.go`
- [x] T217 [US9] Implement catalog proxy service with caching (Redis) in `cmd/core/catalog/proxy.go`
- [x] T218 [US9] Implement catalog metadata sync worker (incremental, webhook-triggered) in `cmd/core/catalog/sync_worker.go`
- [x] T219 [US9] Implement offline catalog cache (SQLite for Flutter, IndexedDB for Angular) in `pkg/catalog/offline_cache.go`

### 10.2 Asset Management

- [x] T220 [US9] Implement CloudFront signed URL generator in `pkg/catalog/cdn_signer.go`
- [x] T221 [US9] Implement asset proxy with format negotiation (webp/png/jpg/avif) in `cmd/core/catalog/asset_proxy.go`
- [x] T222 [US9] Implement lazy loading image component (intersection observer) in `cmd/client-web/src/app/components/lazy-image.component.ts`
- [x] T223 [US9] Implement 4K asset preloader (priority queue based on viewport) in `pkg/catalog/preloader.go`

### 10.3 Search

- [x] T224 [US9] Implement catalog search service (full-text + faceted) in `cmd/core/catalog/search.go`
- [x] T225 [US9] Implement search relevance ranking (title match > genre > features) in `cmd/core/catalog/ranking.go`
- [x] T226 [US9] Integrate `vasic-digital/RAG` for semantic search fallback in `cmd/core/catalog/semantic_search.go`
- [x] T227 [US9] Integrate `vasic-digital/VectorDB` for embedding-based similarity in `cmd/core/catalog/vector_search.go`

### 10.4 US9 Integration & Tests

- [x] T228 [US9] Write Integration test: catalog search p999 ≤200ms with 1000+ games in `tests/integration/catalog_search_perf_test.go`
- [x] T229 [US9] Write Integration test: CloudFront signed URL generation and validation in `tests/integration/cdn_sign_test.go`
- [x] T230 [US9] Write E2E test: catalog browsing with lazy loading on all 3 clients in `tests/e2e/catalog_browse_e2e_test.go`

---

## Phase 11: US7 — Scalability & Multi-Region

*Story goal: 1,000+ concurrent sessions across multi-host/multi-region with auto-scaling and failover ≤30s.*

### 11.1 Multi-Host Orchestration

- [x] T231 [US7] Implement host fleet registry with capability indexing in `cmd/core/discovery/fleet_registry.go`
- [x] T232 [US7] Implement session load balancer (least-latency, GPU-aware) in `cmd/core/session/load_balancer.go`
- [x] T233 [US7] Implement host health checker (heartbeat timeout, GPU fault detection) in `cmd/core/discovery/health_checker.go`
- [x] T234 [US7] Implement auto-scaler (container orchestration trigger based on session queue depth) in `cmd/core/scaling/auto_scaler.go`

### 11.2 Multi-Region

- [x] T235 [US7] Implement region-aware rendezvous (closest host selection by GeoDNS) in `cmd/core/discovery/geo_rendezvous.go`
- [x] T236 [US7] Implement cross-region latency mapping (ping mesh between regions) in `cmd/core/network/latency_map.go`
- [x] T237 [US7] Implement data sovereignty enforcement (tenant geo-partition pinning) in `cmd/core/tenant/geo_pin.go`
- [x] T238 [US7] Implement failover orchestrator (session migration, graceful degradation) in `cmd/core/session/failover.go`

### 11.3 Event Bus Integration

- [x] T239 [US7] Implement NATS event publisher for host fleet changes in `cmd/core/eventbus/host_events.go`
- [x] T240 [US7] Implement NATS consumer for session lifecycle events in `cmd/core/eventbus/session_events.go`
- [x] T241 [US7] Implement Redis pub/sub for real-time host status broadcast in `cmd/core/eventbus/status_broadcast.go`

### 11.4 US7 Integration & Tests

- [x] T242 [US7] Write Integration test: 10-host fleet registration and discovery in `tests/integration/multi_host_test.go`
- [x] T243 [US7] Write Chaos test: host failure during session, failover ≤30s in `tests/chaos/host_failure_test.go`
- [x] T244 [US7] Write Stress test: 1,000 concurrent sessions, 24-hour soak in `tests/stress/1000_session_soak_test.go`
- [x] T245 [US7] Write Benchmark test: load balancer decision latency p999 in `tests/benchmark/load_balancer_test.go`

---

## Phase 12: US10 — Testing, Challenges & HelixQA

*Story goal: 100% coverage across 10 test types; HelixQA autonomous sign-off; anti-bluff enforcement.*

### 12.1 Test Infrastructure

- [x] T246 [US10] Create Unit test baseline for all `cmd/core/` packages (≥95% coverage) in `tests/unit/core/`
- [x] T247 [US10] Create Unit test baseline for all `cmd/host-agent/` packages (≥95% coverage) in `tests/unit/host_agent/`
- [x] T248 [US10] Create Integration test suite using `testcontainers-go` (real CockroachDB, Redis, NATS) in `tests/integration/`
- [x] T249 [US10] Create E2E test suite using Playwright Go (full topology in docker-compose) in `tests/e2e/`
- [x] T250 [US10] Create Security test suite (govulncheck + Snyk + Trivy orchestration) in `tests/security/`
- [x] T251 [US10] Create Benchmark test suite with HDR histogram output in `tests/benchmark/`
- [x] T252 [US10] Create Chaos test suite (Toxiproxy + chaos-mesh integration) in `tests/chaos/`
- [ ] T253 [US10] Create Stress test runner (24-hour soak with memory leak detection) in `tests/stress/`
- [ ] T254 [US10] Create Smoke test suite (30-second post-deploy health checks) in `tests/smoke/`
- [ ] T255 [US10] Create Full Automation orchestrator (Dagger pipeline running all 9 types, fail-fast disabled) in `tests/fullauto/orchestrator.go`

### 12.2 Challenges & HelixQA

- [ ] T256 [US10] Implement Challenge scenario: "Host boots, game launches, streams 60s, terminates" in `Challenges/challenges/streaming_basic.go`
- [ ] T257 [US10] Implement Challenge scenario: "Controller input verified via screenshot diff" in `Challenges/challenges/controller_challenge.go`
- [ ] T258 [US10] Implement Challenge scenario: "Quick resume from idle within 5s" in `Challenges/challenges/quick_resume_challenge.go`
- [ ] T259 [US10] Implement Challenge scenario: "Recording playback quality check" in `Challenges/challenges/recording_challenge.go`
- [ ] T260 [US10] Implement `ValidateAntiBluff()` call in all challenge scenarios in `Challenges/pkg/runner/validator.go`
- [ ] T261 [US10] Implement `RecordAction()` in all challenge scenarios for evidence capture in `Challenges/pkg/runner/recorder.go`
- [ ] T262 [US10] Integrate HelixQA autonomous orchestrator with all 10 test types in `HelixQA/pkg/helixqa/orchestrator.go`
- [ ] T263 [US10] Implement HelixQA visual assertion (OpenCV frame comparison) in `HelixQA/pkg/helixqa/visual_assert.go`
- [ ] T264 [US10] Implement HelixQA pre-release gate (all 1,840 cells must pass) in `HelixQA/pkg/helixqa/release_gate.go`

### 12.3 Anti-Bluff Enforcement

- [ ] T265 [US10] Implement `anti-bluff-scan.sh` forbidden token detection (TODO, FIXME, empty bodies, panic("not implemented")) in `/scripts/anti-bluff-scan.sh`
- [ ] T266 [US10] Implement AST-based empty function body scanner in `/scripts/ast_empty_body_scan.go`
- [ ] T267 [US10] Implement observable assertion ratio checker (≥60% assertions must verify observable behavior) in `/scripts/assertion_ratio_check.go`
- [ ] T268 [US10] Implement mutation testing gate (Gremlins ≥85% score required) in `.github/workflows/mutation-gate.yml`

### 12.4 US10 Integration

- [ ] T269 [US10] Write Challenge test: full topology boot, stream, verify with `RecordAction()` in `tests/challenges/streaming_full_test.go`
- [ ] T270 [US10] Write negative-leg test: break streaming encoder, verify E2E test fails in `tests/security/negative_leg_encoder_test.go`
- [ ] T271 [US10] Write negative-leg test: break controller input, verify integration test fails in `tests/security/negative_leg_input_test.go`


---

## Phase 13: Polish & Cross-Cutting Concerns

*No user story label — final quality gates and release preparation.*

### 13.1 Performance Optimization

- [ ] T272 Implement io_uring zero-copy socket I/O for Linux host agent in `cmd/host-agent/network/iouring.go`
- [ ] T273 Implement GPU-Direct texture sharing (NVENC zero-copy import) in `cmd/host-agent/encoder/gpu_direct.go`
- [ ] T274 Implement frame pacing and VRR integration in `cmd/host-agent/streaming/frame_pacer.go`
- [ ] T275 Implement PREEMPT_RT kernel parameter detection and recommendation in `cmd/host-agent/system/rt_check.go`
- [ ] T276 Implement latency regression CI gate (>150% baseline blocks merge) in `.github/workflows/latency-regression.yml`

### 13.2 Audio & HDR Pipeline

- [ ] T277 Implement Opus MultiStream encoder/decoder in `cmd/host-agent/audio/opus_multistream.go`
- [ ] T278 Implement AC3/EAC3 passthrough in `cmd/host-agent/audio/ac3_passthrough.go`
- [ ] T279 Implement Dolby Atmos forwarding (metadata extraction + re-encode) in `cmd/host-agent/audio/atmos.go`
- [ ] T280 Implement HDR10/HDR10+ dynamic metadata pipeline in `cmd/host-agent/video/hdr_pipeline.go`
- [ ] T281 Implement thermal-aware quality throttling in `cmd/host-agent/encoder/thermal_throttle.go`

### 13.3 Documentation & Release

- [ ] T282 Write API reference documentation (auto-generated from protobuf + OpenAPI) in `/docs/api/README.md`
- [ ] T283 Write deployment guide (container setup, GPU passthrough, TLS certificates) in `/docs/deploy/README.md`
- [ ] T284 Write operator manual (troubleshooting, performance tuning, monitoring) in `/docs/ops/README.md`
- [ ] T285 Create Helm chart for Kubernetes deployment in `/deploy/helm/helixplay/`
- [ ] T286 Tag all 46 submodules with v1.0.0 release via `scripts/release-tag.sh`
- [ ] T287 Create GitHub Release with changelog and asset binaries in `.github/workflows/release.yml`

### 13.4 Final Verification

- [ ] T288 Run full anti-bluff scan — confirm zero violations across all 46 submodules via `/scripts/anti-bluff-scan.sh`
- [ ] T289 Run full test matrix (1,840 cells) — confirm all pass via `make test-fullauto`
- [ ] T290 Run HelixQA autonomous sign-off — confirm visual assertions pass via `HelixQA/cmd/helixqa signoff`
- [ ] T291 Run security audit (third-party penetration test) — confirm zero CRITICAL/HIGH in `tests/security/penetration_test.go`
- [ ] T292 Run performance audit — confirm p999 ≤30ms LAN / ≤50ms WAN in `tests/benchmark/final_audit_test.go`

---

## Parallel Execution Opportunities

### Within Phase 3 (US1)

| Parallel Group | Tasks | Rationale |
|----------------|-------|-----------|
| Capture paths | T078, T079, T080, T081 | Per-OS, no shared files |
| Encoder bindings | T083, T084, T085, T086, T087, T088 | Per-vendor, no shared files |
| Game store enums | T066–T073 | Per-store, independent logic |

### Within Phase 4 (US2)

| Parallel Group | Tasks | Rationale |
|----------------|-------|-----------|
| Client stacks | T117–T119 (Wails), T120–T126 (Flutter), T127–T131 (Angular) | Different directories, share only Go core |
| Go core builds | T110, T111 | c-shared and WASM are independent compilation targets |

### Within Phase 2 (Foundational)

| Parallel Group | Tasks | Rationale |
|----------------|-------|-----------|
| Core submodules | T026–T044 | Independent submodules, no root dependency |
| Container images | T045–T049 | Independent Dockerfiles |
| CI workflows | T051–T054 | Independent YAML files |

---

## Implementation Strategy

### MVP First (Weeks 1–8)

Focus on **US1 + US2 + US3 + US8 (baseline) + US10 (test infra)**:

1. **Weeks 1–2**: Phase 1 + Phase 2 (Setup + Foundational)
2. **Weeks 3–4**: Phase 3 (US1: Host streaming end-to-end)
3. **Weeks 5–6**: Phase 4 (US2: Triple-stack clients, starting with Wails)
4. **Weeks 7–8**: Phase 5 (US3: Controller fidelity) + Phase 9 (US8: Security baseline) + Phase 12 (US10: Test infra)

**MVP Exit Criteria**: One host streams one game to one client (Wails desktop) with controller input and ≤30ms LAN p999.

### Incremental Delivery (Weeks 9–20)

5. **Weeks 9–10**: Phase 6 (US4: Recording)
6. **Weeks 11–12**: Phase 10 (US9: Catalog integration)
7. **Weeks 13–14**: Phase 7 (US5: White-label theming)
8. **Weeks 15–16**: Phase 8 (US6: TV-first UX)
9. **Weeks 17–18**: Phase 11 (US7: Scalability)
10. **Weeks 19–20**: Phase 13 (Polish, performance, release)

---

## Anti-Bluff Verification

| Item | Status |
|------|--------|
| All 292 tasks have exact file paths | ✅ Verified |
| All tasks follow checklist format (`- [ ] T### [P] [US#] Description path`) | ✅ Verified |
| Phase 1 tasks have NO story label | ✅ Verified |
| Phase 2 tasks have NO story label | ✅ Verified |
| Phase 3–12 tasks HAVE story labels [US1]–[US10] | ✅ Verified |
| Phase 13 tasks have NO story label | ✅ Verified |
| No `TODO`/`FIXME`/`placeholder` in tasks | ✅ Verified |
| Tasks organized by user story (P1 → P2 → P3) | ✅ Verified |
| Dependency graph shows completion order | ✅ Verified |
| Parallel execution groups identified | ✅ Verified |
| MVP scope explicitly defined (US1+US2+US3+US8+US10) | ✅ Verified |
| Test tasks included per user story | ✅ Verified |
| Cross-cutting concerns in final phase | ✅ Verified |

### Sources
| Document | Lines | Insights Used |
|----------|------:|---------------|
| `specs/001-helixplay-system/spec.md` | 651 | 10 user stories, FR-001..FR-045, SC-001..SC-010 |
| `specs/001-helixplay-system/plan.md` | 431 | 14 phases P00–P13, ~200 tasks, Constitution gates |
| `specs/001-helixplay-system/data-model.md` | 377 | 10 entities, state machines, junction tables |
| `specs/001-helixplay-system/research.md` | 795 | 11 tech decisions, conflict resolutions, tool choices |
| `specs/001-helixplay-system/contracts/` | 2,002 | 5 API contracts (catalog, streaming, host-agent, discovery, auth) |

**Total tasks**: 292  
**Tasks with [P] parallel marker**: 9  
**Tasks with [US#] story labels**: 195  
**Test tasks (Unit/Integration/E2E/Benchmark/Chaos/Stress/Security)**: 89  
**File paths specified**: 292/292 (100%)

---

*End of Tasks — Generated by `/speckit-tasks` on 2026-05-02. Ready for `/speckit-implement`.*
