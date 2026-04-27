# CloudStream Gaming Platform — Comprehensive Technical Specification & Implementation Plan

## Executive Summary
### Platform Vision & Scope
#### Cloud gaming market projected to exceed $8B by 2027[^4^], with self-hosted solutions (Sunshine, Moonlight, Parsec) gaining traction among technical users
#### CloudStream enables remote gameplay from personal/corporate host machines (macOS/Linux/Windows) to any device with PS4 Pro-like UX and zero-lag streaming
#### Primary language Go with Angular/Tauri/Flutter/KMP client options; targets LAN <30ms and WAN <50ms glass-to-glass latency
### Architecture at a Glance
#### Four-layer topology: Client Apps → API Gateway/Streaming Relay → Host Agents → Data Layer; WebRTC/Pion for streaming, NATS for events, CockroachDB for persistence
#### Hybrid client strategy: Wails (desktop) + Flutter+Go FFI (mobile/TV) + Angular+Go WASM (web); shared Go core library compiled three ways
#### "Sunshine++" host agent pattern: fork Sunshine's proven capture/encode pipeline, add enterprise session/lifecycle management layer
### Key Performance Targets
#### Table: Performance KPIs — 4K@60Hz LAN <30ms, WAN <50ms; input forwarding <5ms; catalog API <100ms p99; supports 10K concurrent streams per region
### Document Guide
#### Report organized into 15 chapters across 5 parts: Foundation, Core Platform, Platform Services, Infrastructure, Delivery & Risk
#### 6-phase implementation roadmap with 110+ fine-grained tasks from MVP to production scale

## 1. System Architecture Overview (~3500 words, 3 tables, 2 diagrams)
### 1.1 Architecture Principles & Design Philosophy
#### 1.1.1 Latency-first design: every decision optimizes for minimal glass-to-glass delay; microseconds matter at scale
#### 1.1.2 Event-driven microservices with Go: goroutines + channels for concurrent I/O, Protocol Buffers for serialization, NATS JetStream for event bus
#### 1.1.3 State management strategy: host state in CockroachDB, session state in Redis, streaming state in-memory; eventual consistency acceptable for catalog, strong consistency for sessions
#### 1.1.4 Three-client-one-core pattern: shared Go business logic compiled as c-shared (Flutter), native binary (Wails IPC), and WASM (web)
### 1.2 System Topology
#### 1.2.1 C4 Context diagram: Users → CloudStream Platform → Host Machines → Game Data Sources (IGDB, SteamGridDB)
#### 1.2.2 C4 Container diagram: API Gateway (Go/Fiber), Streaming Relay (Go/Pion WebRTC), Host Agent (Go/Sunshine++), Catalog Service (Go/gRPC), Auth Service (Go/OAuth2), Web UI (Angular), Mobile/TV App (Flutter), Desktop App (Wails)
#### 1.2.3 Component interaction: REST for catalog/management, WebSocket for session control, SSE for events, WebRTC for media streaming, DataChannels for controller input
#### 1.2.4 Table: Component responsibility matrix — 8 core components with their technology, protocol, and scaling strategy
### 1.3 Host Agent Architecture
#### 1.3.1 Single-binary design inspired by Sunshine: capture module + encode module + stream server + HTTP API in one process
#### 1.3.2 Session state machine: IDLE → CONNECTING → NEGOTIATING → STREAMING ↔ PAUSED → TERMINATING; per-session goroutine isolation
#### 1.3.3 Host capability advertisement: GPU model via NVML, encoder support via FFmpeg enumeration, max resolution, HDR support; JSON broadcast via mDNS and registry API
#### 1.3.4 Health monitoring: heartbeat every 5s to registry, GPU temperature throttling detection, automatic session recovery on capture failure
### 1.4 Streaming Pipeline Architecture
#### 1.4.1 Five-stage pipeline: Game Render → Capture (DXGI/ScreenCaptureKit/PipeWire) → Encode (NVENC/QuickSync/AMF/VAAPI) → Packetize (RTP/WebRTC) → Transmit (UDP/ICE)
#### 1.4.2 Zero-copy pipeline: DXGI shared handle → CUDA interop → NVENC (Windows); IOSurface → CVPixelBuffer → VideoToolbox (macOS); DMA-BUF → EGLImage → VAAPI (Linux)
#### 1.4.3 Adaptive bitrate control loop: receiver reports → bandwidth estimation → encoder reconfiguration → gradual quality transition; target 2-second reaction time
### 1.5 Data Flow & Communication Patterns
#### 1.5.1 Control plane vs data plane separation: REST/WebSocket for control (low frequency, reliable), WebRTC/DataChannels for data (high frequency, latency-sensitive)
#### 1.5.2 Event streaming architecture: NATS JetStream topics for host.status, session.events, controller.input, stream.metrics; consumer groups per service
#### 1.5.3 Table: Message types and their transport protocol — 12 message categories with protocol rationale
### 1.6 Deployment Topology
#### 1.6.1 Single-host (home user): all services on one machine, mDNS discovery, direct P2P WebRTC
#### 1.6.2 Multi-host LAN: dedicated API/registry server, host agents on gaming machines, local TURN relay
#### 1.6.3 Cloud-hosted: Kubernetes-orchestrated microservices, regional TURN clusters, CDN for assets, CockroachDB multi-region

## 2. Go Language Technology Stack Selection (~3000 words, 4 tables, 1 diagram)
### 2.1 Go Ecosystem for Backend Services
#### 2.1.1 Fiber vs Gin vs Echo: Fiber leads raw throughput at 89K req/s, Gin offers best ecosystem balance, Echo provides middleware richness; Gin selected for API Gateway
#### 2.1.2 Pion WebRTC v4: pure Go, no CGO, 100% cross-platform including WASM; full PeerConnection API, DataChannels, ICE, Simulcast, SVC support
#### 2.1.3 NATS JetStream: sub-ms core latency, 800K msg/s throughput, Go-native client; selected as primary event bus over Redis Pub/Sub and RabbitMQ
#### 2.1.4 CockroachDB: PostgreSQL-compatible, horizontal scaling, follower reads for 8x latency improvement; native Go pgx driver with connection pooling
### 2.2 Cross-Platform Client Development in Go
#### 2.2.1 Table: Framework capability matrix — Fyne vs Wails vs Gio vs Go Mobile vs Go WASM across 12 criteria (desktop, mobile, web, TV, accessibility, performance, maturity)
#### 2.2.2 Wails v2 for Desktop: Go backend + Angular web frontend, native OS webviews (WebView2/WebKit), ~15MB binaries, <0.5s startup, in-memory IPC; desktop-only limitation accepted
#### 2.2.3 Flutter + Go FFI for Mobile/TV: Go compiled as c-shared library, Dart FFI for synchronous calls, full platform coverage (iOS/Android/Android TV), best accessibility support
#### 2.2.4 Angular + Go WASM for Web: standard full-stack web architecture, Go business logic compiled to TinyGo WASM (~500KB), REST/gRPC-Web APIs; works on any browser including smart TV browsers
#### 2.2.5 Table: Client platform → framework mapping with Go core compilation target and IPC mechanism
### 2.3 Shared Go Core Library Design
#### 2.3.1 Module structure: `cloudstream-core` Go module with packages for protocol, controller, session, catalog-api, streaming-client, theme-engine
#### 2.3.2 C API boundary design: `cloudstream.h` header with opaque types, 20 exported functions for session lifecycle, input forwarding, and catalog queries
#### 2.3.3 WASM JS shim: `cloudstream.js` wrapper exposing async/await APIs over the Go WASM runtime, bridging WebRTC and browser APIs
#### 2.3.4 Build pipeline: Makefile targets for `build-desktop` (native), `build-mobile` (c-shared), `build-web` (TinyGo WASM); CI matrix across all targets
### 2.4 Development Tooling & DevOps
#### 2.4.1 Go workspace: `go.work` with 6 modules (core, gateway, host-agent, catalog-service, relay-server, web-ui); unified versioning
#### 2.4.2 Testing: `go test` for unit, `testcontainers-go` for integration, `k6` for load, `PresentMon` for capture latency validation
#### 2.4.3 CI/CD: GitHub Actions with build/test matrix (Win/Mac/Linux + Android/iOS/WASM), Docker image builds, Helm chart packaging

## 3. Video Streaming Protocols & Codecs (~3500 words, 3 tables, 2 diagrams)
### 3.1 Protocol Selection: WebRTC with Custom UDP Fallback
#### 3.1.1 WebRTC as primary: standards-based, built-in NAT traversal (ICE/STUN/TURN), browser-native, DTLS-SRTP encryption mandatory; Pion implementation for Go
#### 3.1.2 Custom UDP for native desktop: UDP+ENet-like protocol (inspired by Moonlight) for lowest possible latency on LAN; 7ms achievable vs WebRTC's 15-20ms overhead
#### 3.1.3 Protocol abstraction layer: `Streamer` interface with `WebRTCStreamer` and `UDPStreamer` implementations; negotiation during session establishment
#### 3.1.4 Table: WebRTC vs Custom UDP comparison across 10 criteria (latency, NAT traversal, browser support, encryption, complexity)
### 3.2 Codec Strategy: Multi-Codec with H.264 Default
#### 3.2.1 H.264 Baseline Profile: universal hardware decode (98% device coverage), mandatory WebRTC support, lowest encode complexity; 35-50 Mbps for 4K60
#### 3.2.2 HEVC (H.265): 50% bandwidth reduction vs H.264, excellent 4K quality at 15-25 Mbps; limited browser support, complex 3-pool patent licensing
#### 3.2.3 AV1: royalty-free, 30-50% better than HEVC; RTX 40+/Intel Arc/M3+ required for hardware encode; 10-18 Mbps for 4K60; forward-looking option
#### 3.2.4 Table: Codec comparison matrix — H.264 vs HEVC vs AV1 across 8 dimensions (compatibility, bandwidth, quality, latency, licensing, hardware support, battery, future-proofing)
#### 3.2.5 Codec negotiation: SDP offer/answer with fallback chain (AV1 → HEVC → H.264); client capability detection at session start
### 3.3 Hardware Encoder Integration
#### 3.3.1 NVIDIA NVENC: most consistent latency (~7 frames all presets), excellent quality at P4/P5 presets; NVENC SDK via FFmpeg or direct CUDA interop
#### 3.3.2 Intel QuickSync: lowest latency at 5 frames (83ms) in Ultra Low Latency mode; best for integrated graphics scenarios
#### 3.3.3 AMD AMF: predictable 6-9 frame latency; lower rate-distortion performance but sufficient for mid-range
#### 3.3.4 Apple Media Engine: dedicated H.264/HEVC/ProRes hardware on Apple Silicon; VideoToolbox framework integration
#### 3.3.5 Linux VAAPI: vendor-agnostic, supports Intel/AMD via Mesa drivers; quality varies by implementation
### 3.4 Frame Pacing, V-Sync, and HDR
#### 3.4.1 V-Sync bypass strategy: disable host V-Sync, use frame limiter at stream FPS to minimize encode-start latency; Fast Sync/Enhanced Sync alternatives
#### 3.4.2 Frame pacing algorithm: capture immediately on frame availability, encode in parallel with next frame render, transmit via paced sending
#### 3.4.3 HDR pass-through: Windows DXGI R16G16B16A16_FLOAT scRGB capture, macOS EDR IOSurface, Linux HDR DMA-BUF; HLG and PQ tone mapping pipeline
### 3.5 Adaptive Bitrate for Gaming
#### 3.5.1 Why gaming ABR differs from VoD: sudden motion requires proactive bitrate increase, static scenes can tolerate lower bitrate, reaction time must be <2 seconds
#### 3.5.2 SQP (Scalable Quality Protocol) vs GCC: SQP achieves 2-3x higher bandwidth than GCC under TCP competition; Camel addresses frame-level burst undershooting
#### 3.5.3 Implementation: receiver-side bandwidth estimation from RTCP receiver reports, sender-side encoder reconfiguration, 3-tier quality ladder (4K/1080p/720p)

## 4. Controller Input Forwarding System (~3000 words, 3 tables, 2 diagrams)
### 4.1 Input Capture Architecture
#### 4.1.1 Cross-platform HID abstraction: `ControllerDevice` interface with OS-specific implementations (Windows Raw Input, Linux evdev/hidraw, macOS IOKit, Android InputManager)
#### 4.1.2 SDL2 GameController as fallback: automatic button/axis mapping for 200+ controllers, but limited to basic input; raw HID for advanced features
#### 4.1.3 Bluetooth HID support: HOGP (HID over GATT Profile) for BLE gamepads, BR/EDR for PS4/Xbox controllers; pairing managed by OS, enumerated via standard HID APIs
#### 4.1.4 USB OTG: direct wired connection preferred for competitive play (1ms polling at 1000Hz vs 8ms for Bluetooth 125Hz)
### 4.2 Input Serialization & Network Protocol
#### 4.2.1 Binary protocol design: 16-32 byte packets with controller ID, timestamp, button mask, 6 analog axes (left stick XY, right stick XY, L2/R2 triggers), 2 gyro axes
#### 4.2.2 Transport: UDP for native clients (lowest overhead), WebRTC DataChannels in unreliable/unordered mode for web clients; maxRetransmits=0, ordered=false
#### 4.2.3 Input batching strategy: critical inputs (button presses) sent immediately, analog updates batched at 1-2ms intervals; 1000Hz effective update rate on wired USB
#### 4.2.4 CemuhookUDP protocol: port 26730, standard for gyro/motion forwarding; implement server-side for gyroscope/accelerometer data relay
### 4.3 Host-Side Input Injection
#### 4.3.1 Windows: ViGEmBus virtual gamepad driver (successor to ScpVBus) for XInput/DirectInput emulation; SendInput for keyboard/mouse
#### 4.3.2 Linux: uinput kernel module for creating virtual input devices; evdev event injection; native, no external drivers needed
#### 4.3.3 macOS: foohid kext for virtual HID devices; CGEventPost for keyboard/mouse injection; requires disabling SIP for kext installation
#### 4.3.4 Table: Per-OS injection methods with capabilities, limitations, and setup requirements
### 4.4 Advanced Controller Features
#### 4.4.1 Rumble/haptic forwarding: dual-motor vibration commands forwarded from host to client, played on physical controller; 15ms+ WAN delay makes precise haptics challenging
#### 4.4.2 DualSense adaptive triggers: require raw HID output reports over USB; Bluetooth does not support haptics on PC; custom protocol extension for trigger resistance
#### 4.4.3 Gyro aiming: CemuhookUDP standard at ~1000Hz IMU polling; maps to virtual mouse or right-stick offset; transformative for FPS games
#### 4.4.4 Touchpad/gyro mouse emulation: map DualSense touchpad to virtual mouse on host; configurable sensitivity and dead zones
### 4.5 Input Latency Optimization
#### 4.5.1 End-to-end input pipeline latency: capture (0.5-1ms) → serialize (0.05ms) → network (1-10ms LAN) → inject (1-3ms) → game process (1-2ms) = ~4-17ms total
#### 4.5.2 Client-side prediction: local echo for UI navigation (menu cursor), not for gameplay actions; masks up to 30ms network latency
#### 4.5.3 Table: Input latency budget breakdown per stage with optimization targets and measurement methodology

## 5. Host Game Capture Technologies (~3000 words, 2 tables, 2 diagrams)
### 5.1 Windows Capture: DXGI Desktop Duplication API
#### 5.1.1 DXGI DDA as gold standard: hardware-accelerated capture with dirty rectangle tracking, event-driven frame acquisition via AcquireNextFrame
#### 5.1.2 HDR capture: DuplicateOutput1() with R16G16B16A16_FLOAT scRGB format; tone mapping pipeline for SDR clients
#### 5.1.3 Zero-copy to NVENC: DXGI shared handle → CUDA interop → encoder; no CPU-RAM roundtrip
#### 5.1.4 Limitations: cannot capture fullscreen exclusive DirectX without driver workaround, max 4 concurrent duplication sessions per GPU, ACCESS_LOST on mode changes
### 5.2 macOS Capture: ScreenCaptureKit + IOSurface
#### 5.2.1 ScreenCaptureKit (macOS 12.3+): modern API replacing CoreDisplay, provides IOSurface-backed frames for zero-copy GPU texture access
#### 5.2.2 IOSurface: kernel-managed shared texture memory across processes; zero-copy to VideoToolbox encoder
#### 5.2.3 Apple Silicon unified memory: eliminates GPU paging entirely, capture has negligible overhead
#### 5.2.4 Limitations: user consent required (first capture shows permission dialog), content captured indicator in menu bar
### 5.3 Linux Capture: PipeWire + DMA-BUF
#### 5.3.1 PipeWire with xdg-desktop-portal: most compatible across Wayland compositors, standard DBus API
#### 5.3.2 DMA-BUF + KMS/DRM: lowest-latency capture path (~2.5% CPU at 4K60); EGL_EXT_image_dma_buf_import for zero-copy to encoder
#### 5.3.3 wlroots protocols: wlr-export-dmabuf and wlr-screencopy for efficient capture on wlroots compositors
#### 5.3.4 Fallback: X11 XShm for legacy setups; massive performance overhead due to GPU→RAM→GPU copies
### 5.4 Capture-to-Encoder Pipeline Integration
#### 5.4.1 Unified frame interface: `CapturedFrame` struct with pixel buffer handle, format (NV12/P010/I420), resolution, timestamp, HDR metadata
#### 5.4.2 Platform-specific implementations: DXGI shared handle (Win), IOSurface ID (macOS), DMA-BUF fd (Linux); all convert to encoder-compatible format
#### 5.4.3 Performance targets: capture overhead <3ms per frame, <5% GPU utilization at 4K60; measured via GPUView and custom profiling
#### 5.4.4 Table: Per-OS capture technology summary with API, zero-copy path, HDR support, and limitations

## 6. Client Application Architecture (~3500 words, 4 tables, 2 diagrams)
### 6.1 Shared Go Core Library
#### 6.1.1 Module design: `cloudstream-core` with 6 packages — `protocol` (streaming negotiation), `controller` (input abstraction), `session` (session state machine), `catalog` (API client), `theme` (design token engine), `platform` (OS abstraction)
#### 6.1.2 C API for Flutter FFI: 20 exported functions including `CS_CreateSession`, `CS_SendInput`, `CS_GetFrame`, `CS_SetTheme`, `CS_GetCatalog`; C-string and handle-based
#### 6.1.3 WASM exports for web: `js/wasm` package with `StartSession`, `ProcessInput`, `RenderFrame` mapped to JavaScript async functions via syscall/js
#### 6.1.4 Platform abstraction layer: `Platform` interface with OS-specific implementations for file storage, notifications, gamepad enumeration, deep linking
### 6.2 Desktop Client (Wails + Angular)
#### 6.2.1 Architecture: Go backend process running cloudstream-core + Wails v2 runtime, Angular frontend in embedded WebView2/WebKit
#### 6.2.2 IPC mechanism: Wails Events for pub/sub (theme changes, session state), Wails Bindings for request/response (catalog API calls, session control)
#### 6.2.3 Window management: fullscreen game view, overlay HUD (Home button, settings, connection stats), catalog browser window
#### 6.2.4 Gamepad integration: Angular subscribes to `gamepadconnected`/`gamepaddisconnected` events via Wails binding to Go controller package
### 6.3 Mobile Client (Flutter + Go FFI)
#### 6.3.1 Architecture: Flutter UI layer + Go c-shared library loaded via `dart:ffi`; single codebase for iOS and Android
#### 6.3.2 Gamepad support: Flutter `gamepads` package for Bluetooth/USB HID enumeration; Go core handles input serialization and network transmission
#### 6.3.3 Touch controls: on-screen virtual gamepad overlay when no physical controller connected; customizable layout and opacity
#### 6.3.4 Android TV variant: Compose for TV Flutter embedding, D-Pad navigation, Leanback launcher integration, 10-foot UI optimization
### 6.4 Web Client (Angular + Go WASM)
#### 6.4.1 Architecture: Angular SPA with Go WASM module (~500KB TinyGo) for protocol handling; browser WebRTC APIs for video reception
#### 6.4.2 WebRTC integration: JS `RTCPeerConnection` for video/audio reception, Go WASM manages DataChannels for controller input via js/wasm bridge
#### 6.4.3 Gamepad API: browser `navigator.getGamepads()` polled at 60Hz; input processed by Go WASM, transmitted via WebRTC DataChannel
#### 6.4.4 PWA capabilities: service worker for offline catalog caching, installable on mobile home screens, fullscreen mode for gaming
### 6.5 TV-First UI Design
#### 6.5.1 10-foot UI principles: all elements readable from 3 meters, D-Pad/remote-only navigation, no hover states, high contrast
#### 6.5.2 Navigation model: horizontal shelf/carousel pattern (PS4/Xbox style), L1/R1 or trigger for shelf switching, directional pad for item selection
#### 6.5.3 Overscan handling: 5% safe margins (48dp sides, 27dp top/bottom) on all screens; background images can bleed to edges
#### 6.5.4 Performance for TV SoCs: 280MB memory budget on 1GB devices, aggressive image caching, RecyclerView/Flutter ListView for large catalogs
### 6.6 Table: Client Platform Comparison
#### 6.6.1 Platform capability matrix: 4 platforms (Desktop, Mobile, Web, TV) × 10 features (streaming quality, controller support, offline catalog, theme switching, HDR, etc.)

## 7. API Design & Communication Layer (~3000 words, 4 tables, 2 code examples)
### 7.1 REST API Design
#### 7.1.1 Base conventions: URL path versioning (`/api/v1/`), JSON request/response, OAuth2 Bearer token auth, rate-limited via uber-go/ratelimit
#### 7.1.2 Endpoints: `GET /api/v1/games` (catalog list), `GET /api/v1/games/{id}` (game details), `POST /api/v1/sessions` (create session), `GET /api/v1/sessions/{id}` (session status), `GET /api/v1/hosts` (list hosts), `POST /api/v1/hosts/{id}/pair` (pair new host)
#### 7.1.3 Protobuf schema: `cloudstream.catalog.v1`, `cloudstream.session.v1`, `cloudstream.host.v1` packages; generated Go + TypeScript/Dart bindings
#### 7.1.4 Table: Complete REST endpoint reference with method, path, request/response schema, auth requirement, and rate limit
### 7.2 WebSocket API for Session Control
#### 7.2.1 Protocol: JSON messages over WebSocket, authenticated via JWT in `Authorization` header at connection time
#### 7.2.2 Message types: `session.create`, `session.join`, `session.leave`, `stream.start`, `stream.pause`, `stream.resume`, `input.configure`, `host.command`
#### 7.2.3 Scaling: Redis Pub/Sub for cross-server WebSocket message distribution; sticky sessions via cookie-based routing
#### 7.2.4 Code example: Go WebSocket handler with goroutine-per-connection pattern and graceful shutdown
### 7.3 Server-Sent Events (SSE) for Real-Time Updates
#### 7.3.1 Event types: `host.online`, `host.offline`, `game.installed`, `game.updated`, `session.started`, `session.ended`, `achievement.unlocked`
#### 7.3.2 Implementation: Go `http.Flusher` with `Content-Type: text/event-stream`, reconnection handling via `Last-Event-ID` header
#### 7.3.3 Client consumption: Angular `EventSource`, Flutter `sse_client` package, native `NSURLSessionEventSource` on iOS
### 7.4 WebRTC Signaling Protocol
#### 7.4.1 Signaling flow: WebSocket-based SDP offer/answer exchange with ICE candidate trickling; Go `pion/webrtc` v4 PeerConnection management
#### 7.4.2 TURN server integration: coturn or pion/turn with ephemeral HMAC credentials (time-limited, auto-generated); required for 15-20% of sessions behind symmetric NAT
#### 7.4.3 ICE candidate handling: host and client gather candidates → exchange via WebSocket → connectivity check → select best path (host < srflx < relay)
#### 7.4.4 Code example: Pion PeerConnection setup with DataChannel configuration for game input
### 7.5 gRPC for Internal Services
#### 7.5.1 Connect RPC (buf.build/connect) for browser-facing services: supports gRPC, gRPC-Web, and REST/JSON from single handler
#### 7.5.2 Internal service mesh: mTLS between all Go services, service discovery via Consul, load balancing via client-side round-robin

## 8. Game Catalog & Metadata System (~3000 words, 3 tables, 1 diagram)
### 8.1 Metadata Sources & Integration
#### 8.1.1 IGDB API: free tier 10K requests/month, Pro $99/50K; covers 200K+ games with titles, descriptions, genres, ratings, release dates, platforms, screenshots; daily data dumps
#### 8.1.2 SteamGridDB: best 4K artwork source (heroes at 3840x1240, covers at 600x900); WebP/PNG support, static + animated; max 50 results per request
#### 8.1.3 Steam Web API: 100K requests/day free; `GetOwnedGames` for library import, `appdetails` for metadata (screenshots, trailers, genres)
#### 8.1.4 Fallback sources: RAWG (500K games, ML recommendations), GOG Galaxy API (200 req/hour/IP), MobyGames, Giant Bomb
### 8.2 Game Metadata Schema
#### 8.2.1 Core schema: 25 fields covering external IDs (IGDB, Steam, etc.), title, description, media (cover/hero/logo/icon/screenshots/trailers), release info, classification, ratings, technical specs, ownership
#### 8.2.2 Per-game settings: controller profiles (JSON bindings, sensitivity, dead zones), graphics settings (resolution, bitrate, codec preference), cloud save location mapping
#### 8.2.3 Table: Complete metadata schema with field name, type, source, and example value
### 8.3 4K Asset Pipeline
#### 8.3.1 Image optimization: AVIF at 50-80% quality for primary delivery, WebP at 75-85% as fallback; format negotiation via Accept header or CDN
#### 8.3.2 Responsive images: `srcset` with width descriptors (480w, 720w, 1080w, 1440w, 2160w); lazy loading with Intersection Observer
#### 8.3.3 CDN integration: Cloudflare Images or ImageKit for automatic format conversion, resizing, and global edge caching
#### 8.3.4 Local caching: Ristretto in-memory cache (hot assets) + disk LRU (image files) + SQLite metadata index; total cache budget 500MB on mobile, 2GB on desktop
### 8.4 Search, Sorting & Filtering
#### 8.4.1 Search engine: SQLite FTS5 for MVP (zero dependencies, BM25 ranking); Meilisearch for production (<50ms typo-tolerant search, 6-8x storage overhead)
#### 8.4.2 Sorting options: recently played, alphabetical, release date, genre, rating, playtime, date added; B-tree indexes on all sort fields
#### 8.4.3 Filtering: faceted search with pre-computed counts; filters by genre, platform, multiplayer, controller support, source (Steam/EGS/GOG), year, rating
#### 8.4.4 Game collections: favorites, wishlist, "recently played", custom user collections; stored per-user in PostgreSQL

## 9. Theming & White-Label Customization System (~3000 words, 2 tables, 2 diagrams)
### 9.1 Design Token Architecture
#### 9.1.1 Three-tier token system: primitive tokens (raw values: #3B82F6), semantic tokens (color-primary, bg-surface), component tokens (button-bg-primary, card-border-radius)
#### 9.1.2 W3C DTCG specification: JSON token files following Design Tokens Community Group format; Style Dictionary v4 for cross-platform transformation
#### 9.1.3 Platform outputs: CSS custom properties (web), Dart theme classes (Flutter), JSON theme config (Wails/Angular), Android XML (TV)
### 9.2 Runtime Theme Engine
#### 9.2.1 CSS custom properties for web: `--cs-primary`, `--cs-bg-surface`, etc.; runtime switching by updating `data-theme` attribute on `<html>` element
#### 9.2.2 Day/Dark/Auto modes: 3-layer preference (user choice > system `prefers-color-scheme` > default dark); inline `<head>` script prevents flash-of-wrong-theme
#### 9.2.3 Material Design 3 tonal palette: algorithmic color generation from single source color using HCT color space; produces 13 tonal steps for primary, secondary, tertiary palettes
#### 9.2.4 Dynamic theme API: `POST /api/v1/themes` to upload brand colors, logo SVG, font preferences; runtime fetch and application without restart
### 9.3 White-Label Configuration
#### 9.3.1 Per-tenant configuration: database-stored brand settings (name, colors, logo URL, fonts, layout density); fetched on app startup and cached locally
#### 9.3.2 Self-service brand portal: web UI for white-label customers to upload logos, pick colors, preview theme changes in real-time
#### 9.3.3 Asset injection: SVG logos via dynamic component, splash screens via PWA manifest generation, favicon generation from brand color
#### 9.3.4 Table: White-label configuration parameters with type, default value, and scope (global vs per-tenant)
### 9.4 Accessibility in Theming
#### 9.4.1 WCAG 2.1 AA compliance: 4.5:1 contrast ratio for normal text, 3:1 for large text and UI components; contrast checking in token generation pipeline
#### 9.4.2 Focus indicators: 2px+ outline with offset, `:focus-visible` for keyboard-only indication; never rely on color alone for state
#### 9.4.3 Reduced motion: `prefers-reduced-motion` media query disables parallax, crossfade transitions; instant state changes for accessibility

## 10. Scalability & Infrastructure Design (~3000 words, 3 tables, 2 diagrams)
### 10.1 Host Discovery & Registration
#### 10.1.1 Local discovery: mDNS/Bonjour for same-network host detection; host broadcasts `_cloudstream._tcp` service with capability metadata
#### 10.1.2 Cloud registry: REST API for host registration with JWT authentication; heartbeat every 5s via WebSocket ping; automatic unregistration after 3 missed heartbeats
#### 10.1.3 Host capability advertisement: JSON document with GPU model, VRAM, encoder support, installed games, max resolution, HDR capability
### 10.2 Load Balancing & Session Scheduling
#### 10.2.1 Composite scoring algorithm: 60% network latency (measured via STUN), 25% GPU utilization (NVML query), 15% active sessions; lowest score wins
#### 10.2.2 Geographic affinity: anycast DNS for regional routing; clients connect to nearest relay server, which selects optimal host within region
#### 10.2.3 GPU-aware scheduling: track per-GPU VRAM and utilization; prevent oversubscription; migrate sessions on GPU failure
#### 10.2.4 Table: Load balancing algorithm parameters with weights, data sources, and update frequency
### 10.3 Relay & TURN Infrastructure
#### 10.3.1 TURN server: coturn or pion/turn with ephemeral HMAC credentials; UDP relay primary, TCP fallback for restrictive firewalls
#### 10.3.2 Relay topology: regional relay clusters co-located with host pools; relay bandwidth is primary scaling cost (~25 Mbps per 4K stream)
#### 10.3.3 P2P vs relayed: attempt direct (host reflexive) first, relay only when necessary; 80% of LAN sessions are direct, 20-30% of WAN need relay
### 10.4 Database & Caching Architecture
#### 10.4.1 Primary database: CockroachDB multi-region with follower reads; game metadata, user profiles, session history, host registry
#### 10.4.2 Cache layers: Valkey (open-source Redis fork) for session state and hot catalog data; local LRU on clients for 4K assets
#### 10.4.3 Table: Data store selection matrix — 8 data categories (session state, user data, game metadata, etc.) with store choice and rationale
### 10.5 Monitoring & Observability
#### 10.5.1 Metrics: Prometheus with custom collectors for streaming latency, encode time, frame drops, input lag, host GPU utilization
#### 10.5.2 Tracing: Jaeger with OpenTelemetry Go SDK; trace spans across all microservices for end-to-end latency analysis
#### 10.5.3 Alerting: Grafana alerts for p99 latency >100ms, host GPU temp >85C, error rate >1%, relay bandwidth >80% capacity

## 11. Security Architecture (~2500 words, 2 tables, 1 diagram)
### 11.1 Authentication & Authorization
#### 11.1.1 OAuth2/OIDC: Auth0/Keycloak/Authentik identity provider; Authorization Code flow for web/desktop, Device Authorization Grant (RFC 8628) for TV/console
#### 11.1.2 JWT design: short-lived access tokens (15 min) + long-lived refresh tokens (7 days) stored in httpOnly cookies; RS256 signature with key rotation
#### 11.1.3 Role-based access: `user` (play games, browse catalog), `admin` (manage hosts, users), `host` (host agent service account), `whitelabel` (brand configuration)
### 11.2 WebRTC Security
#### 11.2.1 DTLS-SRTP: mandatory AES-128-CM-HMAC-SHA1 encryption for all media; no opt-out possible per WebRTC spec
#### 11.2.2 Signaling security: WebSocket over TLS (WSS) with JWT auth; signaling channel is the primary attack surface
#### 11.2.3 ICE privacy: mDNS candidates prevent local IP leakage; TURN server restricts relay IPs to registered users only
### 11.3 Host Agent Security
#### 11.3.1 Host sandboxing: host agent runs as unprivileged user; game processes launched in separate job object (Windows) or cgroup (Linux)
#### 11.3.2 Anti-cheat compatibility: use only OS-provided capture APIs (DXGI DDA, ScreenCaptureKit, PipeWire) — no hooks or injections that trigger EAC/BattlEye/Vanguard
#### 11.3.3 Input sanitization: validate all controller input packets (range checks, dead zone application) before injection; prevent malformed HID reports
### 11.4 Infrastructure Security
#### 11.4.1 mTLS: all inter-service communication encrypted with mutual TLS; cert-manager for automatic certificate rotation in Kubernetes
#### 11.4.2 Rate limiting: token bucket per IP (100 req/min for REST), per user (10 session creates/min), per host (50 connections/min)
#### 11.4.3 Audit logging: all authentication events, session creation/termination, host pairing, admin actions logged to immutable store with 90-day retention

## 12. Performance Optimization & Latency Engineering (~3000 words, 2 tables, 2 diagrams)
### 12.1 End-to-End Latency Budget
#### 12.1.1 Complete pipeline: input polling (~1ms) → serialization (~0.1ms) → network uplink (~5-30ms) → host processing (~1-3ms) → game render (~8-16ms) → capture (~1-3ms) → encode (~2-5ms) → network downlink (~5-30ms) → decode (~2-5ms) → display (~4-16ms)
#### 12.1.2 Optimization targets: <30ms competitive (LAN), <50ms acceptable (WAN); requires optimization of every stage simultaneously
#### 12.1.3 Table: Latency budget per stage with current baseline, target, and optimization technique
### 12.2 Network Optimization
#### 12.2.1 DSCP QoS: EF (46) for game controller input, AF41 (34) for video stream; router-level traffic prioritization
#### 12.2.2 Forward Error Correction: XOR-based parity at 20-25% overhead achieves 99.5% packet recovery; Reed-Solomon for higher redundancy scenarios
#### 12.2.3 BBR congestion control: replaces CUBIC for streaming connections; better throughput on high-BDP links, faster recovery from loss
### 12.3 Frame Pipeline Optimization
#### 12.3.1 NVIDIA Reflex integration: Reflex 2.0 with Frame Warp reduces latency up to 75% by warping rendered frame based on latest input; host-side requirement
#### 12.3.2 Frame pacing algorithm: capture-on-frame-available pattern eliminates vsync wait; encode pipeline runs parallel with game render thread
#### 12.3.3 Jitter buffer: adaptive 20-50ms base on client, scaling to network conditions; resize threshold at 10% to avoid oscillation
### 12.4 High Refresh Rate & 4K
#### 12.4.1 Bandwidth requirements: H.264 at 4K60 = 35-50 Mbps, 4K120 = 70-100 Mbps; HEVC halves these; AV1 provides further 30% reduction
#### 12.4.2 Display sync: G-Sync/FreeSync/VRR on client for tear-free display; Cloud G-SYNC requires >60Hz VRR-capable display
#### 12.4.3 Parsec achieves 4-8ms at 240Hz LAN [^1^]; Moonlight 15.7ms optimized; target matching Parsec-level performance for competitive scenarios

## 13. Implementation Phases & Task Breakdown (~4000 words, 6 tables, 1 diagram)
### 13.1 Phase 1: Foundation (Weeks 1-4)
#### 13.1.1 Go workspace setup: initialize 6-module workspace, CI pipeline, code generation for Protobuf, linting (golangci-lint), testing framework
#### 13.1.2 Shared core library: `protocol` package with WebRTC/UDP abstraction, `controller` package with HID interfaces, `catalog` package with API client
#### 13.1.3 Host agent MVP: fork Sunshine, add REST API wrapper, implement game launch via Steam URL protocol, basic capture-to-stream pipeline
#### 13.1.4 Table: Phase 1 task breakdown — 18 tasks with assignee, estimated hours, dependencies, and deliverables
### 13.2 Phase 2: Core Streaming (Weeks 5-8)
#### 13.2.1 WebRTC streaming: Pion PeerConnection integration, video track creation from captured frames, DataChannel for controller input
#### 13.2.2 Controller input pipeline: implement capture (SDL2 + raw HID), serialization (binary protocol), network transmission (UDP + WebRTC DataChannel)
#### 13.2.3 Host-side injection: Windows ViGEmBus integration, Linux uinput implementation, macOS foohid support
#### 13.2.4 Table: Phase 2 task breakdown — 22 tasks with technical details
### 13.3 Phase 3: Client Applications (Weeks 9-14)
#### 13.3.1 Desktop client: Wails v2 project, Angular catalog UI, gamepad integration, fullscreen streaming view
#### 13.3.2 Mobile client: Flutter project, Go FFI bindings, Android/iOS gamepad support, touch controls overlay
#### 13.3.3 Web client: Angular SPA, Go WASM integration, browser WebRTC, PWA configuration
#### 13.3.4 Android TV: Compose for TV UI, D-Pad navigation, voice search, overscan handling
#### 13.3.5 Table: Phase 3 task breakdown — 28 tasks across 4 client platforms
### 13.4 Phase 4: Platform Services (Weeks 15-18)
#### 13.4.1 Game catalog: IGDB + SteamGridDB integration, 4K asset pipeline, search (FTS5), sorting/filtering UI
#### 13.4.2 Theming engine: design token system, CSS custom properties, runtime theme switching, white-label configuration API
#### 13.4.3 User management: OAuth2 integration, profiles, favorites, recently played, playtime tracking
#### 13.4.4 Table: Phase 4 task breakdown — 18 tasks
### 13.5 Phase 5: Infrastructure & Scale (Weeks 19-22)
#### 13.5.1 Containerization: Docker images for all services, Kubernetes manifests, Helm charts
#### 13.5.2 Multi-region: CockroachDB multi-region setup, regional TURN clusters, CDN configuration for 4K assets
#### 13.5.3 Monitoring: Prometheus metrics, Grafana dashboards, Jaeger tracing, alerting rules
#### 13.5.4 Table: Phase 5 task breakdown — 14 tasks
### 13.6 Phase 6: Polish & Production (Weeks 23-26)
#### 13.6.1 Performance optimization: end-to-end latency profiling, encoder tuning, network QoS implementation, FEC
#### 13.6.2 Security hardening: penetration testing, mTLS verification, rate limiting validation, audit log review
#### 13.6.3 Advanced features: HDR streaming, high refresh rate (120Hz+), gyro aiming, adaptive triggers, game suspension
#### 13.6.4 Table: Phase 6 task breakdown — 16 tasks
### 13.7 Project Timeline & Dependencies
#### 13.7.1 Gantt chart overview: 26-week timeline with phase overlaps, critical path identification, milestone definitions
#### 13.7.2 Team composition: 8-10 person team (3 backend, 2 client, 1 DevOps, 1 QA, 1 UI/UX, 1 product)
#### 13.7.3 Table: Phase summary with duration, deliverables, and exit criteria

## 14. Wireframes & UI/UX Specifications (~3000 words, 6 wireframe descriptions, 2 tables)
### 14.1 Design System Overview
#### 14.1.1 Layout grid: 12-column responsive grid, 24px gutters, 1200px max content width on desktop; TV uses 5% overscan safe area
#### 14.1.2 Typography: Inter font family (desktop/web), Roboto (Android/TV), San Francisco (iOS); 6-level scale from caption (12px) to hero (48px)
#### 14.1.3 Spacing: 8px base unit, 11 tokens (0, 4, 8, 12, 16, 24, 32, 48, 64, 96, 128); density modes: compact/comfortable/spacious
#### 14.1.4 Motion: 150ms standard transition, 300ms emphasis, decelerate easing; `prefers-reduced-motion` support
### 14.2 Landing Screen Wireframe
#### 14.2.1 PS4 Pro-inspired layout: full-screen background blur with ambient game art, horizontal shelves for "Continue Playing", "Recently Added", "Favorites", "All Games"
#### 14.2.2 Navigation: top bar with user avatar, settings gear, search icon; left sidebar with category shortcuts on desktop; D-Pad/arrow navigation on TV
#### 14.2.3 Game card: 2:3 aspect ratio cover art, title overlay at bottom, hover/focus state with scale 1.05x and shadow elevation; 4K cover display
#### 14.2.4 Empty state: animated gamepad illustration, "Connect a host to start" CTA, setup wizard trigger
### 14.3 Game Library Screen Wireframe
#### 14.3.1 Three view modes: grid (default, 4-6 columns), list (compact with details), cover flow (3D carousel); toggle via top-right segmented control
#### 14.3.2 Filter sidebar: genre checkboxes, platform toggles, year range slider, rating stars, multiplayer toggle; collapsible on mobile
#### 14.3.3 Sort options: dropdown with recently played, alphabetical, release date, rating, playtime; ascending/descending toggle
#### 14.3.4 Search: real-time search bar with debounced 300ms query, instant results with game covers, empty state with suggestions
### 14.4 Game Detail Screen Wireframe
#### 14.4.1 Hero section: full-width hero artwork (3840x1240), game title, developer/publisher, release year, genre tags, rating, "Play" primary button
#### 14.4.2 Media gallery: horizontal scrollable screenshot carousel (4K thumbnails), auto-playing muted trailer, lightbox on click
#### 14.4.3 Details panel: description, system requirements (host-side), supported controllers, playtime statistics, last played date
#### 14.4.4 Related games: "Similar Titles" shelf at bottom based on genre/platform matching
### 14.5 In-Game Streaming Screen Wireframe
#### 14.5.1 Full-screen video: game stream fills entire display, no window chrome; aspect ratio maintained with letterboxing if needed
#### 14.5.2 HUD overlay: press Home/guide button to reveal translucent overlay with menu (Resume, Settings, Quit, Switch Game); 300ms fade-in
#### 14.5.3 Connection stats (optional): small corner overlay showing latency, bitrate, FPS, packet loss; toggle in settings
#### 14.5.4 Safe quit flow: "Quit Game?" confirmation modal, auto-save detection, graceful shutdown progress indicator
### 14.6 Settings Screens Wireframe
#### 14.6.1 General settings: theme selection (Day/Dark/Auto), language, audio output, notification preferences
#### 14.6.2 Streaming settings: resolution (720p/1080p/1440p/4K), FPS limit (30/60/120/Auto), bitrate (manual or auto), codec preference, HDR toggle
#### 14.6.3 Controller settings: detected controllers list, mapping configuration per game, sensitivity sliders, dead zone adjustment, vibration toggle
#### 14.6.4 Host management: paired hosts list, connection status, GPU info, test connection button, unpair option
### 14.7 White-Label Configuration Screen
#### 14.7.1 Brand portal: preview pane on left, configuration form on right; real-time theme preview
#### 14.7.2 Upload areas: logo SVG upload with preview, favicon generation, splash screen image; drag-and-drop with validation
#### 14.7.3 Color picker: primary/secondary/accent color wheels with WCAG contrast preview; generated tonal palette display
#### 14.7.4 Export: theme JSON download, CSS custom properties file, integration guide for developers

## 15. Risk Analysis & Mitigation Strategies (~2500 words, 2 tables)
### 15.1 Technical Risks
#### 15.1.1 Anti-cheat blocking: RISK — kernel-level anti-cheat (Vanguard, EAC) may block screen capture or virtual controllers; MITIGATION — use official OS capture APIs only, maintain anti-cheat compatibility matrix, engage vendors for commercial whitelisting
#### 15.1.2 Cross-platform controller latency: RISK — Bluetooth polling at 125Hz adds 8ms vs 1ms USB; MITIGATION — recommend wired/2.4GHz for competitive play, implement input prediction, document latency tradeoffs
#### 15.1.3 macOS virtual controller limitation: RISK — foohid kext requires SIP disable, reducing security; MITIGATION — document clearly, provide CGEventPost fallback for basic input, explore DriverKit alternative
### 15.2 Operational Risks
#### 15.2.1 Bandwidth costs at scale: RISK — 4K streaming at 50 Mbps × 10K users = 500 Gbps; $135K-$233K/month bare metal; MITIGATION — HEVC/AV1 to reduce bandwidth, edge caching, tiered quality based on subscription
#### 15.2.2 Host machine availability: RISK — consumer-grade hosts not designed for 24/7 operation, GPU failure, thermal throttling; MITIGATION — health monitoring, automatic failover, thermal management, redundant hosts
#### 15.2.3 Content rights for metadata: RISK — IGDB/Steam API terms may restrict commercial redistribution of game data; MITIGATION — review ToS carefully, implement multi-source aggregation, cache with attribution
### 15.3 Business Risks
#### 15.3.1 Market competition: RISK — established players (GeForce NOW, Xbox Cloud Gaming) have massive resources; MITIGATION — differentiate on self-hosted/private cloud model, controller feature richness, white-label capability
#### 15.3.2 Technology obsolescence: RISK — WebRTC evolution, new codecs (AV2 expected 2027), OS API changes; MITIGATION — abstraction layers for codec and protocol, active open-source community engagement, quarterly technology reviews
#### 15.3.3 Regulatory: RISK — data privacy laws (GDPR, CCPA), accessibility requirements (ADA, EAA); MITIGATION — privacy-by-design, WCAG 2.1 AA compliance built in, data residency options
### 15.4 Risk Matrix
#### 15.4.1 Table: Risk severity × probability matrix for all 10 identified risks with mitigation owner, timeline, and residual risk level

# References
## Research Dimension Files
- **Type**: Deep research outputs
- **Description**: 12 dimension research files, cross-verification, and insight extraction
- **Path**: /mnt/agents/output/research/cloudgaming_dim01.md through cloudgaming_dim12.md, cloudgaming_cross_verification.md, cloudgaming_insight.md

## Requirements Analysis
- **Type**: Structured requirements extraction
- **Description**: 160 explicit and implicit requirements categorized by priority and type
- **Path**: /mnt/agents/output/cloudgaming_requirements.md

## Structure Design
- **Type**: Report structure specification
- **Description**: Chapter hierarchy with word counts and required elements
- **Path**: /mnt/agents/output/cloudgaming_structure.md

## Outline References
- **Type**: Raw citation collection
- **Description**: Sources collected during outline design
- **Path**: /mnt/agents/output/cloudgaming_outline_references_raw.md
