## 13. Implementation Phases & Task Breakdown

The preceding chapters defined the target state for the CloudStream platform: a Sunshine++ host agent streaming via Pion WebRTC v4 to Wails (desktop), Flutter (mobile/TV), and Angular (web) clients, backed by NATS JetStream, CockroachDB, and a design-token theming engine. This chapter decomposes the 26-week delivery target into six phases with overlapping boundaries, 116 individual tasks, explicit assignee roles, hour estimates, dependency chains, and phase-gate exit criteria. The plan assumes an 8-10 person team in two-week sprints.

### 13.1 Phase 1: Foundation (Weeks 1-4)

The foundation phase establishes the engineering toolchain, shared core libraries, and a minimally viable host agent. The phase is organized into four workstreams: workspace infrastructure, shared libraries, host agent MVP, and continuous integration.

#### 13.1.1 Go Workspace Setup

The platform is structured as a Go multi-module workspace with six modules: `cmd/gateway`, `cmd/catalog`, `cmd/session`, `cmd/relay`, `pkg/core`, and `cmd/hostagent`. Week 1 begins with `go.work` initialization, module boundaries, and import restrictions. The CI pipeline runs on GitHub Actions with three workflows: `lint` (golangci-lint v1.61 with `gosec`, `staticcheck`, and `errname`), `test` (race-detector-enabled with 70% coverage threshold for core, 60% for services), and `build` (cross-compilation for linux/amd64, linux/arm64, windows/amd64, darwin/arm64). Protobuf code generation uses `buf` v1.45 with `connect-go` and `openapiv2` plugins, executing on every pull request to keep API definitions and generated code synchronized.

#### 13.1.2 Shared Core Library

The `pkg/core` module provides three sub-packages. The `protocol` package abstracts WebRTC and UDP transport behind a `StreamTransport` interface with methods for `CreatePeerConnection()`, `AddVideoTrack()`, `CreateDataChannel(label)`, and `Close()`. This abstraction resolves the conflict between WebRTC (standards-compliant, built-in NAT traversal) and custom UDP (lower latency, 7 ms LAN demonstrated by Parsec's BUD protocol) by supporting both behind a common interface [^1^]. The `controller` package defines HID abstraction interfaces with platform-specific implementations via build tags (`sdl2`, `android`, `browser`). The `catalog` package implements typed API clients for IGDB's Apicalypse query language, SteamGridDB, and the Steam Web API, normalizing responses into the 28-field schema from Chapter 8.

#### 13.1.3 Host Agent MVP

The host agent MVP follows the Sunshine++ pattern: fork Sunshine v0.23.1 (1000+ GitHub stars, actively maintained) and wrap its capture-to-stream pipeline with a management REST API [^75^]. Sunshine solves cross-platform capture (DXGI DDA on Windows, ScreenCaptureKit on macOS, PipeWire/KMS on Linux), hardware-accelerated encoding (NVENC, AMF, QuickSync, VAAPI), and RTP packetization. The management layer adds four endpoints: `POST /api/v1/sessions` (launch and stream), `GET /api/v1/sessions/{id}` (status), `POST /api/v1/sessions/{id}/pause` (PS4-style home button), and `DELETE /api/v1/sessions/{id}` (graceful termination with 30-second escalation). Game launching uses the Steam URL protocol (`steam://rungameid/{appid}`). The MVP delivers a single-binary validated against three test games (Unreal, Unity, proprietary engines).

#### 13.1.4 Phase 1 Task Breakdown

Table 13.1 enumerates the 18 tasks with estimated effort, assigned role, dependencies, and deliverables.

| Task ID | Description | Assignee | Hours | Dependencies | Deliverable |
|---------|-------------|----------|-------|--------------|-------------|
| P1-T01 | Initialize Go workspace with 6-module layout | Backend Lead | 8 | — | `go.work`, `go.mod` files |
| P1-T02 | Configure GitHub Actions CI (lint, test, build) | DevOps | 16 | P1-T01 | Workflow YAML files |
| P1-T03 | golangci-lint config with security linters | Backend Lead | 6 | P1-T01 | `.golangci.yml` |
| P1-T04 | Protobuf code generation (buf, connect-go) | Backend Lead | 10 | P1-T01 | Generated proto Go files |
| P1-T05 | `protocol` package: StreamTransport interface | Backend Dev 1 | 20 | P1-T04 | `pkg/core/protocol/` |
| P1-T06 | WebRTC transport adapter (Pion v4) | Backend Dev 1 | 24 | P1-T05 | WebRTC transport impl |
| P1-T07 | UDP transport adapter (Parsec BUD-inspired) | Backend Dev 2 | 20 | P1-T05 | UDP transport impl |
| P1-T08 | Controller HID abstraction interfaces | Backend Lead | 8 | — | Interface definitions |
| P1-T09 | `controller` package with SDL2 capture | Backend Dev 1 | 24 | P1-T08 | `pkg/core/controller/` |
| P1-T10 | Catalog API client architecture design | Backend Lead | 6 | — | Design document |
| P1-T11 | IGDB Apicalypse client with normalization | Backend Dev 2 | 20 | P1-T10 | IGDB client package |
| P1-T12 | SteamGridDB 4K asset client | Backend Dev 2 | 12 | P1-T10 | Asset download client |
| P1-T13 | Fork Sunshine v0.23.1, establish clean build | Backend Dev 1 | 16 | — | Sunshine fork, compiles |
| P1-T14 | REST API wrapper (4 session endpoints) | Backend Dev 1 | 24 | P1-T13 | Session management API |
| P1-T15 | Steam URL protocol game launcher | Backend Dev 2 | 12 | P1-T14 | `steam://rungameid/` launcher |
| P1-T16 | Capture-to-stream MVP pipeline integration | Backend Dev 1 | 20 | P1-T13, P1-T06 | End-to-end stream test |
| P1-T17 | Unit tests for core packages (>70% coverage) | Backend Dev 2 | 16 | P1-T05, P1-T09 | Coverage report |
| P1-T18 | Integration test: host agent + client stub | QA Engineer | 16 | P1-T16, P1-T17 | Test report |

The 18 tasks sum to 268 hours (~67 hours/week across the backend team). The critical path is P1-T01 → P1-T04 → P1-T05 → P1-T06 → P1-T16 → P1-T18. Task P1-T07 (UDP transport) can slip to Week 5 without downstream impact since WebRTC is the default transport for all client platforms.

### 13.2 Phase 2: Core Streaming (Weeks 5-8)

Phase 2 integrates the streaming pipeline: video tracks from captured frames, controller input over DataChannels, and virtual controller injection on the host.

#### 13.2.1 WebRTC Streaming

Pion WebRTC v4 provides the PeerConnection and track primitives. Video tracks use a custom `SampleProvider` reading from Sunshine's capture thread into a `TrackLocalStaticSample` via a lock-free triple-buffered ring buffer — the same "capture-on-frame-available" pattern that contributes to Sunshine's 12.6–26.7% latency advantage [^15^]. The controller DataChannel uses `Ordered: false` and `MaxRetransmits: 0` for unreliable delivery of analog data, while a separate reliable channel carries discrete button events, avoiding head-of-line blocking [^17^].

#### 13.2.2 Controller Input Pipeline

Capture uses SDL2 on desktop (`SDL_GameController` with hot-plug detection), Android `InputDevice` with `SOURCE_GAMEPAD` filtering on mobile, and the W3C Gamepad API for web. All sources normalize into a 32-byte binary packet: 2 bytes button mask, 4 bytes per analog stick, 1 byte per trigger, 6 bytes accelerometer, 6 bytes gyroscope, and 5 bytes touchpad — compact enough for 120 Hz transmission at ~38 kbps while capturing the full DualSense feature set.

#### 13.2.3 Host-Side Input Injection

Windows uses ViGEmBus for virtual Xbox 360/DualShock 4 controllers (signed, open-source, though formally retired as of February 2024 — a successor is being evaluated) [^18^]. Linux uses the native `uinput` subsystem. macOS uses `foohid` or newer `DriverKit` equivalents. All three expose a common `VirtualController` Go interface with `CreateDevice()`, `SendReport(state)`, and `CloseDevice()` methods.

#### 13.2.4 Phase 2 Task Breakdown

| Task ID | Description | Assignee | Hours | Dependencies | Deliverable |
|---------|-------------|----------|-------|--------------|-------------|
| P2-T01 | Pion PeerConnection + Sunshine capture integration | Backend Dev 1 | 28 | P1-T16 | Video track from frames |
| P2-T02 | Triple-buffered lock-free frame ring | Backend Dev 1 | 16 | P2-T01 | Frame buffer impl |
| P2-T03 | DataChannel (unreliable) for analog input | Backend Dev 1 | 12 | P2-T01 | Unordered DataChannel |
| P2-T04 | DataChannel (reliable) for critical events | Backend Dev 1 | 8 | P2-T01 | Reliable event channel |
| P2-T05 | Binary input serialization (32-byte packet) | Backend Dev 2 | 16 | P1-T09 | `MarshalBinary()` |
| P2-T06 | SDL2 gamepad capture (desktop) | Backend Dev 2 | 20 | P2-T05 | SDL2 reader |
| P2-T07 | Browser Gamepad API capture (web) | Client Dev 1 | 16 | P2-T05 | JS Gamepad bridge |
| P2-T08 | Android InputDevice capture (mobile) | Client Dev 2 | 20 | P2-T05 | Android gamepad reader |
| P2-T09 | UDP input transmission (fallback) | Backend Dev 2 | 16 | P1-T07, P2-T05 | UDP sender/receiver |
| P2-T10 | Input transmission via WebRTC DataChannel | Backend Dev 1 | 12 | P2-T03, P2-T05 | DataChannel input path |
| P2-T11 | Windows ViGEmBus virtual controller | Backend Dev 1 | 24 | P2-T05 | `VirtualController` (Win) |
| P2-T12 | Linux uinput virtual controller | Backend Dev 2 | 20 | P2-T05 | `VirtualController` (Linux) |
| P2-T13 | macOS foohid virtual controller | Backend Dev 2 | 20 | P2-T05 | `VirtualController` (macOS) |
| P2-T14 | Host-side input packet receiver/parser | Backend Dev 1 | 12 | P2-T11, P2-T12 | Packet handler |
| P2-T15 | Raw HID access (gyro, touchpad) | Backend Dev 2 | 16 | P2-T06 | HID report parser |
| P2-T16 | Per-game controller profile system | Backend Dev 2 | 16 | P2-T06 | Profile storage |
| P2-T17 | End-to-end latency measurement harness | Backend Dev 1 | 12 | P2-T02, P2-T14 | LED-trigger timing tool |
| P2-T18 | LAN streaming: <30 ms validation | QA Engineer | 16 | P2-T17 | Latency report |
| P2-T19 | Multi-game compatibility (5+ titles) | QA Engineer | 20 | P2-T11, P2-T14 | Compatibility matrix |
| P2-T20 | Anti-cheat audit (EAC, BattlEye, Vanguard) | QA Engineer | 16 | P2-T19 | Anti-cheat report |
| P2-T21 | Streaming protocol specification doc | Backend Lead | 12 | P2-T10 | Protocol spec |
| P2-T22 | Encode latency + throughput benchmarking | Backend Dev 1 | 12 | P2-T02 | Benchmark report |

Phase 2 totals 332 hours. The critical path is P1-T16 → P2-T01 → P2-T02 → P2-T17 → P2-T18, with the sub-30 ms LAN validation as the phase exit gate. The three platform injection tracks (P2-T11, P2-T12, P2-T13) execute concurrently with Windows prioritized.

### 13.3 Phase 3: Client Applications (Weeks 9-14)

Phase 3 delivers four client applications sharing the Go core library via three compilation targets: C-shared for Flutter FFI, native binary for Wails IPC, and WASM for the web.

#### 13.3.1 Desktop Client

The desktop client uses Wails v2 with an Angular frontend. Wails v2's in-memory IPC avoids serialization overhead of HTTP or WebSocket bridges, making it optimal for latency-sensitive streaming [^26^]. Gamepad integration delegates directly to the Go core's SDL2 module with no FFI boundary. Fullscreen streaming switches the WebView to borderless mode, rendering the WebRTC video element at the DOM root with a PS4-style home button gesture (hold guide button 1 second) for overlay access.

#### 13.3.2 Mobile Client

The Flutter application uses Go FFI (`buildmode=c-shared`) linking the shared core. The Go core exports 15 C functions: `CSInitialize()`, `CSConnect()`, `CSGetVideoFrame()`, `CSSendInput()`, `CSDisconnect()`, `CSCleanup()`, and 9 others. Flutter `MethodChannel` handles platform-specific audio routing; `EventChannel` handles gamepad hot-plug. Touch controls overlay as a Flutter `Stack` widget above the video stream when no physical gamepad is connected.

#### 13.3.3 Web Client

The Angular SPA uses TinyGo-compiled WASM for controller serialization, catalog API calls, and session management, while JavaScript handles WebRTC peer connection and video rendering. The JS-Go bridge passes controller state as `ArrayBuffer` objects into WASM, which returns binary packets for DataChannel transmission. PWA configuration includes a dynamically generated Web App Manifest per tenant, service worker caching, and `standalone` display mode.

#### 13.3.4 Android TV Client

The TV client shares the Flutter + Go FFI foundation with distinct UI for 10-foot interaction. Compose for TV (v1.0 stable) replaces deprecated Leanback [^24^]. D-Pad navigation uses `ImmersiveList` for horizontal shelves and `Carousel` for hero banners with focus restoration. Voice search integrates via Android `SearchManager`. Overscan handling applies 5% safe-area padding via `WindowInsets`. The TV and mobile clients share ~70% of Dart/Go code.

#### 13.3.5 Phase 3 Task Breakdown

| Task ID | Description | Assignee | Hours | Dependencies | Deliverable |
|---------|-------------|----------|-------|--------------|-------------|
| P3-T01 | Wails v2 project scaffold + Angular frontend | Client Dev 1 | 16 | P1-T01 | Desktop skeleton |
| P3-T02 | Angular catalog UI (shelves, cards, hero) | Client Dev 1 | 32 | P3-T01 | Catalog browsing UI |
| P3-T03 | Desktop gamepad via Go core SDL2 | Client Dev 1 | 20 | P3-T01, P2-T06 | Gamepad input |
| P3-T04 | Fullscreen streaming view with WebRTC | Client Dev 1 | 24 | P3-T02, P2-T01 | Stream player |
| P3-T05 | Desktop settings, profile, login UI | Client Dev 1 | 16 | P3-T02 | User screens |
| P3-T06 | Flutter project scaffold with Go FFI | Client Dev 2 | 16 | P1-T01 | Mobile skeleton |
| P3-T07 | Design Go C API (15 functions) for FFI | Backend Lead | 12 | P2-T05 | C header file |
| P3-T08 | Go core c-shared build (all platforms) | Backend Dev 2 | 16 | P3-T07 | `.so`, `.dylib`, `.dll` |
| P3-T09 | Flutter catalog UI (mobile layout) | Client Dev 2 | 28 | P3-T06 | Mobile catalog UI |
| P3-T10 | Flutter Go FFI streaming integration | Client Dev 2 | 24 | P3-T08, P2-T01 | Mobile streaming |
| P3-T11 | Touch controls overlay | Client Dev 2 | 20 | P3-T10 | Virtual gamepad overlay |
| P3-T12 | Android/iOS gamepad via Go core | Client Dev 2 | 20 | P3-T10 | Platform gamepad |
| P3-T13 | Angular SPA project scaffold | Client Dev 1 | 12 | P1-T01 | Web skeleton |
| P3-T14 | Web catalog UI (responsive, PWA-ready) | Client Dev 1 | 28 | P3-T13 | Web catalog UI |
| P3-T15 | TinyGo WASM build of Go core | Backend Dev 2 | 16 | P3-T07 | `.wasm` module |
| P3-T16 | JS-Go bridge for WebRTC + controller | Client Dev 1 | 20 | P3-T15 | WASM integration layer |
| P3-T17 | Browser WebRTC peer connection | Client Dev 1 | 16 | P3-T16 | Web streaming |
| P3-T18 | PWA (service worker, manifest) | Client Dev 1 | 12 | P3-T14 | PWA features |
| P3-T19 | Compose for TV project scaffold | Client Dev 2 | 12 | P3-T06 | TV skeleton |
| P3-T20 | TV catalog UI with D-Pad navigation | Client Dev 2 | 24 | P3-T19 | TV browsing UI |
| P3-T21 | Voice search integration | Client Dev 2 | 12 | P3-T20 | Android voice search |
| P3-T22 | Overscan and focus handling | Client Dev 2 | 10 | P3-T20 | TV-safe layout |
| P3-T23 | TV Go FFI streaming integration | Client Dev 2 | 16 | P3-T08, P3-T20 | TV streaming |
| P3-T24 | Cross-platform UI component library | UI/UX Designer | 24 | P3-T02, P3-T09 | Design system |
| P3-T25 | Accessibility audit (WCAG 2.1 AA) | QA Engineer | 16 | P3-T04, P3-T14 | A11y report |
| P3-T26 | E2E test: catalog → stream → input | QA Engineer | 20 | P3-T04, P3-T10 | E2E test suite |
| P3-T27 | Client performance profiling | QA Engineer | 12 | P3-T26 | Performance report |
| P3-T28 | Beta build preparation (all 4 platforms) | DevOps | 16 | P3-T26 | Signed beta builds |

Phase 3 totals 456 hours, the largest allocation, reflecting four-platform delivery. The critical path runs through P3-T07 → P3-T08 → P3-T10 and P3-T23 (mobile and TV share the FFI dependency). Desktop and web proceed in parallel once the streaming interface stabilizes in Week 9.

### 13.4 Phase 4: Platform Services (Weeks 15-18)

Phase 4 adds the catalog service, theming engine, and user management system that transform the platform from a streaming tool into a full product.

#### 13.4.1 Game Catalog

The Catalog Service (Go/gRPC) aggregates IGDB metadata, SteamGridDB 4K artwork, and Steam Web API library data. IGDB's free tier allows 10,000 requests/month, sufficient for ~3,300 games/day with 3 calls per game [^276^]. SteamGridDB supplies 4K hero banners at 3840x1240 [^275^]. Search uses SQLite FTS5 for embedded clients and Meilisearch for server deployments. The sorting/filtering UI supports genre, release date, platform, player count, and rating, with filter state in URL query parameters for shareable views.

#### 13.4.2 Theming Engine

The three-tier token system (primitive → semantic → component) uses Style Dictionary v4 to transform token sources into CSS custom properties (web/desktop), Dart `ThemeData` (Flutter), and XML resources (Compose for TV) [^604^]. The Dynamic Theme API generates complete palettes from a single brand color via Material Design 3's HCT algorithm [^612^]. Runtime theme switching on web uses CSS custom properties for sub-millisecond changes versus hundreds of milliseconds with CSS-in-JS [^606^].

#### 13.4.3 User Management

OAuth2/OIDC federation uses Auth0 for rapid development with a migration path to Keycloak. Browser/desktop clients use Authorization Code + PKCE; TV clients use the Device Authorization Grant (RFC 8628) for input-constrained devices [^2^]. Profiles store display names, avatars, themes, and controller configs. Favorites pin games to a "My Games" shelf. Recently played tracks the last 20 games. Playtime tracking aggregates session durations from host heartbeats in CockroachDB.

#### 13.4.4 Phase 4 Task Breakdown

| Task ID | Description | Assignee | Hours | Dependencies | Deliverable |
|---------|-------------|----------|-------|--------------|-------------|
| P4-T01 | Catalog Service gRPC API definition | Backend Lead | 10 | P1-T04 | Proto files |
| P4-T02 | IGDB data ingestion pipeline | Backend Dev 1 | 20 | P4-T01 | Ingestion worker |
| P4-T03 | SteamGridDB 4K asset downloader | Backend Dev 1 | 16 | P4-T01 | Asset sync worker |
| P4-T04 | Canonical schema normalization | Backend Dev 1 | 14 | P4-T02 | Normalized records |
| P4-T05 | FTS5 search index implementation | Backend Dev 2 | 20 | P4-T04 | Search with ranking |
| P4-T06 | Sorting/filtering API endpoints | Backend Dev 2 | 16 | P4-T04 | Filtered query API |
| P4-T07 | CDN integration for 4K assets | Backend Dev 1 | 12 | P4-T03 | CDN image URLs |
| P4-T08 | Style Dictionary v4 token pipeline | Client Dev 1 | 20 | P3-T14 | Token build system |
| P4-T09 | Material Design 3 Dynamic Theme API | Backend Dev 2 | 18 | P4-T08 | Theme endpoint |
| P4-T10 | CSS custom properties runtime switching | Client Dev 1 | 12 | P4-T08 | Sub-ms theme switch |
| P4-T11 | White-label configuration API | Backend Dev 2 | 16 | P4-T09 | Tenant config CRUD |
| P4-T12 | Self-service brand portal UI | Client Dev 1 | 24 | P4-T11 | Admin portal |
| P4-T13 | OAuth2/OIDC integration (Auth0) | Backend Dev 1 | 20 | P4-T01 | Auth flow impl |
| P4-T14 | Device Authorization Grant (RFC 8628) | Backend Dev 1 | 14 | P4-T13 | TV auth flow |
| P4-T15 | User profile + favorites system | Backend Dev 2 | 16 | P4-T13 | Profile API |
| P4-T16 | Recently played + playtime tracking | Backend Dev 2 | 14 | P4-T15 | Activity tracking |
| P4-T17 | Catalog E2E integration test | QA Engineer | 16 | P4-T05, P4-T06 | E2E test report |
| P4-T18 | Theming integration test across clients | QA Engineer | 14 | P4-T10 | Cross-client theme test |

Phase 4 totals 272 hours. The critical path is P4-T01 → P4-T02 → P4-T04 → P4-T05 → P4-T17. Phase 4 begins in Week 15, overlapping with Phase 3's final week — this allows the Catalog Service API (P4-T01) to proceed while client beta builds (P3-T28) are finalized.

### 13.5 Phase 5: Infrastructure & Scale (Weeks 19-22)

Phase 5 transitions from development to production-ready deployment with containerization, multi-region topology, and observability.

#### 13.5.1 Containerization

All services use multi-stage Docker builds (`golang:1.23-alpine` builder, `distroless` runtime). The Host Agent supports GPU passthrough with a bare-metal option for anti-cheat compatibility. Kubernetes manifests include Deployments, Services, Ingress, and HPA with CPU (70%) and memory (80%) triggers. Helm charts parameterize per-environment configuration.

#### 13.5.2 Multi-Region

CockroachDB multi-region uses `REGIONAL BY TABLE` / `REGIONAL BY ROW` / `GLOBAL` locality controls with follower reads for up to 8× latency improvement [^538^]. Initial deployment spans three regions (US-East, US-West, EU-West). The game catalog is `GLOBAL`; session data is `REGIONAL BY ROW`. Regional TURN clusters prevent inter-region media relay. CDN pull-through caching serves 4K assets with 7-day TTL for artwork and 1-hour for user content.

#### 13.5.3 Monitoring

Prometheus instruments RED metrics (rate, errors, duration) at the service level, GPU utilization/temperature at the host level, and packet loss/bitrate/RTT per PeerConnection. Grafana dashboards show the four golden signals with per-service drill-down. Jaeger traces requests across gRPC and NATS boundaries. Alerting defines three tiers: `warning` (>1% error rate, 5 min), `critical` (service down >30s, GPU >83°C), and `page` (region failure).

#### 13.5.4 Phase 5 Task Breakdown

| Task ID | Description | Assignee | Hours | Dependencies | Deliverable |
|---------|-------------|----------|-------|--------------|-------------|
| P5-T01 | Docker multi-stage builds for all services | DevOps | 20 | P4-T17 | Docker images |
| P5-T02 | Kubernetes manifests (deploy, service, HPA) | DevOps | 24 | P5-T01 | K8s YAML |
| P5-T03 | Helm charts for environment-specific deploys | DevOps | 16 | P5-T02 | Helm packages |
| P5-T04 | CockroachDB multi-region cluster setup | DevOps | 20 | P5-T02 | 3-region DB |
| P5-T05 | `REGIONAL BY ROW` for session data | Backend Dev 1 | 12 | P5-T04 | Schema optimization |
| P5-T06 | Regional TURN server clusters | DevOps | 16 | P5-T04 | Per-region TURN |
| P5-T07 | CDN configuration for 4K asset caching | DevOps | 12 | P4-T07 | Pull-through cache |
| P5-T08 | Prometheus metrics instrumentation | Backend Dev 1 | 16 | P5-T02 | Metrics endpoints |
| P5-T09 | Grafana dashboards (four golden signals) | DevOps | 14 | P5-T08 | Dashboard JSON |
| P5-T10 | Jaeger distributed tracing integration | Backend Dev 1 | 14 | P5-T02 | Trace propagation |
| P5-T11 | Alerting rules (warning/critical/page) | DevOps | 12 | P5-T09 | AlertManager config |
| P5-T12 | Load testing: 500 concurrent streams | QA Engineer | 20 | P5-T03 | Load test report |
| P5-T13 | Chaos testing (node failure, partition) | QA Engineer | 16 | P5-T12 | Resilience report |
| P5-T14 | Production deployment runbook | DevOps | 12 | P5-T13 | Runbook document |

Phase 5 totals 224 hours. The critical path is P5-T01 → P5-T02 → P5-T03 → P5-T12 → P5-T13 → P5-T14. CockroachDB setup (P5-T04) carries high risk; a PostgreSQL read-replica contingency is maintained.

### 13.6 Phase 6: Polish & Production (Weeks 23-26)

The final phase optimizes performance, hardens security, and adds advanced features. Scope here is explicitly prioritized for reduction if earlier phases slip.

#### 13.6.1 Performance Optimization

Latency profiling uses the LED-trigger methodology validated by Moonlight (photodiode + microcontroller, sub-millisecond accuracy) [^14^]. Targets per pipeline stage from Chapter 12: input polling at 0.5 ms (1000 Hz), network uplink at 5-10 ms (edge nodes within 50 km), video encode at 2.0 ms (NVENC), display at 4-8 ms (120 Hz VRR). Encoder tuning selects H.264 Baseline (no B-frames) for LAN and HEVC ULL for WAN 4K, based on NVENC's 5.8 ms median [^5^]. DSCP EF (value 46) marks gaming traffic; BBR replaces CUBIC to minimize bufferbloat [^6^]. FEC at 25% overhead achieves 99.5% packet loss recovery [^6^].

#### 13.6.2 Security Hardening

Penetration testing covers OWASP Top 10 with focus on WSS signaling (no ws:// downgrade), JWT RS256 validation (15-minute expiry), and input sanitization (schema validation, 120 Hz rate limiting for gamepad). mTLS between services uses cert-manager CA Issuer with 90-day rotation [^24^]. Rate limiting validation tests: 100 req/min per IP (REST), 10/min per user (sessions), 3/min per user (streams). Audit logs verify all session lifecycle and authentication events are retained for 90 days.

#### 13.6.3 Advanced Features

HDR streaming reads 10-bit frames via DXGI `DuplicateOutput1` and negotiates HDR10 metadata. High refresh rate at 120/240 Hz requires VRR (G-SYNC/FreeSync) client displays. Gyro aiming uses CemuhookUDP for motion data transmission. Adaptive triggers (DualSense) extend the binary protocol. Game suspension uses `NtSuspendProcess` (Windows), `cgroup.freeze` (Linux), `SIGSTOP` (macOS), documented as "best effort" due to anti-cheat compatibility concerns.

#### 13.6.4 Phase 6 Task Breakdown

| Task ID | Description | Assignee | Hours | Dependencies | Deliverable |
|---------|-------------|----------|-------|--------------|-------------|
| P6-T01 | LED-trigger latency profiling harness | Backend Dev 1 | 16 | P2-T17 | Profiling tool |
| P6-T02 | End-to-end latency optimization (<30ms LAN) | Backend Dev 1 | 24 | P6-T01 | Optimized pipeline |
| P6-T03 | Encoder tuning: NVENC H.264/HEVC profiles | Backend Dev 1 | 16 | P6-T02 | Encoder configs |
| P6-T04 | DSCP QoS + BBR congestion control | Backend Dev 2 | 14 | P5-T06 | Network optimization |
| P6-T05 | FEC (25% overhead, XOR parity) | Backend Dev 2 | 20 | P2-T01 | FEC encoder/decoder |
| P6-T06 | Penetration testing (external firm) | QA Engineer | 40 | P5-T14 | Pentest report |
| P6-T07 | mTLS verification audit | DevOps | 12 | P5-T03 | mTLS compliance |
| P6-T08 | Rate limiting validation under load | QA Engineer | 12 | P5-T12 | Rate limit tests |
| P6-T09 | Audit log review and retention validation | Backend Dev 2 | 10 | P4-T13 | Audit report |
| P6-T10 | HDR streaming (10-bit, HDR10 metadata) | Backend Dev 1 | 20 | P6-T03 | HDR stream path |
| P6-T11 | 120Hz/240Hz high refresh rate | Backend Dev 1 | 16 | P6-T02 | High-FPS streaming |
| P6-T12 | Gyro aiming (CemuhookUDP) | Backend Dev 2 | 18 | P2-T15 | Gyro pipeline |
| P6-T13 | Adaptive triggers (DualSense extension) | Backend Dev 2 | 16 | P2-T15 | Trigger support |
| P6-T14 | Game suspension (NtSuspendProcess/cgroups) | Backend Dev 1 | 18 | P6-T02 | Quick resume |
| P6-T15 | Final QA regression (all platforms) | QA Engineer | 24 | P6-T06 | Regression report |
| P6-T16 | GA release preparation | Product Manager | 16 | P6-T15 | Release notes |

Phase 6 totals 272 hours. Penetration testing (P6-T06) is the longest single task at 40 hours. Advanced features (P6-T10 through P6-T14) are drop-if-needed, deferrable to post-GA without impacting core functionality.

### 13.7 Project Timeline & Dependencies

#### 13.7.1 Gantt Chart Overview

Figure 13.1 presents the 26-week timeline with phase bars, milestone markers, and the critical path.

![26-Week Implementation Timeline](fig13_gantt_timeline.png)
*Figure 13.1 — 26-week Gantt chart showing six implementation phases, six milestones (M1-M6), critical path, and team composition. Phase overlaps at boundaries compress the timeline by ~4 weeks versus strictly sequential execution.*

The chart illustrates three structural decisions. First, phases overlap at boundaries: Phase 2 begins in Week 5 while Phase 1 completes integration tests in Week 4, a two-week overlap pattern repeated at every boundary. Second, the critical path traces Foundation → Host Agent MVP → WebRTC Integration → Latency Optimization → GA Release; any delay here impacts final delivery. Third, Phase 3 is the longest at six weeks, but its parallel workstreams (desktop, mobile, web, TV) are independent after the Go C API design (P3-T07) completes in Week 9.

Six milestones define exit criteria. M1 (Week 4): Core library compiles with >70% coverage; host streams a test game. M2 (Week 8): LAN latency <30 ms with controller input functional. M3 (Week 14): All four clients browse catalog and initiate streams. M4 (Week 18): Catalog, theming, and user management feature-complete. M5 (Week 22): 500 concurrent streams, chaos tests pass. M6 (Week 26): GA release, all acceptance criteria met.

#### 13.7.2 Team Composition

The team comprises 8-10 people: 3 backend engineers, 2 client engineers, 1 DevOps engineer, 1 QA engineer, 1 UI/UX designer, and 1 product manager. Total capacity is ~320 hours/week (8 × 40), though effective engineering capacity accounts for meetings and code review — the plan budgets 180-228 hours per phase. Backend carries the highest load in Phases 1, 2, and 6; client peaks in Phase 3; DevOps concentrates in Phase 5; QA distributes across Phases 2-6.

#### 13.7.3 Phase Summary

| Phase | Weeks | Duration | Tasks | Est. Hours | Key Deliverable | Exit Criteria |
|-------|-------|----------|-------|-----------|-----------------|---------------|
| 1: Foundation | 1-4 | 4 weeks | 18 | 268 | Core library + host agent MVP | Tests >70%; streams test game |
| 2: Core Streaming | 5-8 | 4 weeks | 22 | 332 | Bidirectional streaming + input | LAN <30 ms; 5+ games |
| 3: Client Apps | 9-14 | 6 weeks | 28 | 456 | 4 functional clients | Catalog → stream → input |
| 4: Platform Services | 15-18 | 4 weeks | 18 | 272 | Catalog, theming, users | Search <50 ms; OAuth2 flow |
| 5: Infrastructure | 19-22 | 4 weeks | 14 | 224 | Containerized multi-region | 500 streams; chaos pass |
| 6: Polish | 23-26 | 4 weeks | 16 | 272 | GA-ready platform | Pentest clean; regression pass |
| **Total** | **1-26** | **26 weeks** | **116** | **1824** | **CloudStream GA** | **M1-M6 met** |

The aggregate 1,824 hours over 26 weeks represents ~70 hours/week of planned work, within team capacity. The highest risk-adjusted effort is Phase 3 (456 hours). The highest technical risk tasks are P2-T01 (Pion-Sunshine integration), P5-T04 (CockroachDB multi-region), and P6-T06 (penetration testing), each with contingency plans: raw UDP fallback, PostgreSQL read replicas, and internal security review respectively. The 26-week timeline is achievable with overlapping phases and parallel workstreams, contingent on Phase 1 delivering core library and protocol abstractions on schedule.
