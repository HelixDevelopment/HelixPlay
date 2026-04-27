# Phase 2: Dimension Decomposition — Cloud Gaming System

## Research Context (Phase 1 Key Findings)
- Cloud gaming uses client-server with video encoding on host, streaming to thin clients. WebRTC is dominant for sub-500ms latency.
- Moonlight + Sunshine is the leading open-source self-hosted combo; Parsec is proprietary but easier.
- H.264 remains the pragmatic choice for live streaming (WebRTC mandatory, universal hardware decode). H.265/AV1 offer bandwidth savings but with tradeoffs.
- Go ecosystem has Pion WebRTC (pure Go, no CGO, cross-platform including WASM). Fyne (pure Go UI, desktop+mobile). Wails (Go + web frontend, desktop only). Gio (immediate mode, all platforms but complex).
- HID controller forwarding over network is sparsely documented — requires deep investigation.

## Dimensions (12 Total)

### Dim 01 — Low-Latency Video Streaming Protocols & Codecs
**Angle**: Technical/protocol perspective — how to achieve 4K@120Hz streaming with minimal glass-to-glass latency.
**Scope**: WebRTC vs custom UDP/RTP protocols; Moonlight/NVIDIA GameStream protocol; codec selection (H.264/HEVC/AV1) for real-time gaming; hardware encoder integration (NVENC, QuickSync, AMF); frame pacing, V-sync handling, HDR; adaptive bitrate; SRT/RTMP comparisons.
**Overlap**: Dim 03 (capture), Dim 05 (Go WebRTC), Dim 12 (latency optimization).

### Dim 02 — Cross-Platform Controller Input Capture & Forwarding
**Angle**: Hardware/peripheral perspective — how to capture and forward controller input from any client device to remote host with <1ms effective latency.
**Scope**: HID APIs (Windows Raw Input, Linux evdev, macOS IOKit); Bluetooth HID profiles; USB OTG on Android/iOS; SDL2 GameController abstraction; input serialization (binary protocol); network transmission (UDP vs WebRTC DataChannels); host-side input injection (SendInput, XTest, CGEvent); rumble/haptic/gyro/accelerometer forwarding; latency masking (client-side prediction).
**Overlap**: Dim 04 (client frameworks), Dim 07 (host agent), Dim 12 (latency).

### Dim 03 — Host OS Game Capture Technologies
**Angle**: OS-specific host perspective — how to capture game video on Windows, macOS, and Linux.
**Scope**: Windows (DXGI Desktop Duplication, NVIDIA Capture SDK, Windows.Graphics.Capture, OBS hook approach); macOS (CoreDisplay, ScreenCaptureKit, IOSurface); Linux (X11 SHM, PipeWire, KMS/DRM); Vulkan/OpenGL/DirectX/Metal capture methods; fullscreen vs windowed capture; HDR surface capture; multi-monitor handling; capture-to-encoder zero-copy pipeline.
**Overlap**: Dim 01 (streaming), Dim 07 (host agent), Dim 12 (latency).

### Dim 04 — Go Ecosystem for Cross-Platform Client Development
**Angle**: Client development perspective — evaluating all Go-based approaches for Desktop, Mobile, Web, and TV clients.
**Scope**: Fyne (desktop+mobile, Material Design, limitations); Wails v2 (desktop only, web frontend); Gio (all platforms, immediate mode, WASM); Go Mobile (gomobile bind/pkg); Go WASM (TinyGo vs standard); Tauri with Go backend; Flutter FFI to Go shared library; Kotlin Multiplatform + Go c-shared; Angular + Go WASM; Performance benchmarks; accessibility support; TV/D-Pad navigation support in each framework.
**Overlap**: Dim 05 (APIs), Dim 10 (theming), Dim 11 (TV UI).

### Dim 05 — Real-Time Communication APIs in Go
**Angle**: Backend API architecture perspective — designing REST, SSE, WebSocket, and WebRTC APIs in Go for gaming.
**Scope**: Gin/Fiber/Echo for REST; Gorilla WebSocket vs native; SSE implementation patterns; gRPC/Connect for internal services; Pion WebRTC for media streaming; NATS/NATS JetStream for event bus; Protocol Buffers schema design; API versioning; rate limiting; load balancing WebSocket connections; message brokers comparison (NATS, Redis Pub/Sub, RabbitMQ, Kafka).
**Overlap**: Dim 01 (streaming), Dim 08 (scalability), Dim 07 (host agent).

### Dim 06 — Game Catalog, Metadata & 4K Asset Management
**Angle**: Content management perspective — building a game library experience comparable to PS4/PS5.
**Scope**: IGDB API integration; SteamGridDB for 4K covers and artwork; Steam/EGS/GOG library import; game metadata schema; 4K image asset pipeline (WebP/AVIF optimization); local caching strategy; CDN for assets; search indexing (Elasticsearch/Meilisearch); sorting/filtering architecture; user ratings & playtime tracking; game categories/genres.
**Overlap**: Dim 10 (white-label), Dim 07 (host catalog).

### Dim 07 — Host Agent Architecture & Game Lifecycle Management
**Angle**: Host runtime perspective — managing game processes safely on remote hosts.
**Scope**: Per-OS host agent design; game process spawning (Steam URL protocol, executable launch); process monitoring (alive, suspended, crashed); graceful shutdown (sending WM_CLOSE vs SIGTERM); save game synchronization; per-game controller profile mapping (like Steam Input); session state machine; host capability advertisement (GPU, encoder, supported codecs); game streaming session negotiation.
**Overlap**: Dim 02 (input), Dim 03 (capture), Dim 09 (security).

### Dim 08 — Scalability, Load Balancing & Multi-Region Infrastructure
**Angle**: DevOps/infrastructure perspective — scaling from single host to many hosts across regions.
**Scope**: Host discovery mechanisms (mDNS, registry service); load balancing algorithms (least-loaded, nearest, GPU-aware); Kubernetes/Docker for host agents; auto-scaling host pools; relay server architecture for NAT traversal; edge computing placement; CDN integration for assets; database clustering (CockroachDB/TiDB/YugabyteDB vs PostgreSQL); horizontal scaling of API servers; metrics and monitoring (Prometheus/Grafana).
**Overlap**: Dim 05 (APIs), Dim 09 (security).

### Dim 09 — Security, Authentication & Host Isolation
**Angle**: Security perspective — protecting hosts, streams, and users.
**Scope**: OAuth2/OIDC authentication; JWT token design; WebRTC DTLS/SRTP encryption; host agent sandboxing; input validation; stream access control; anti-cheat considerations (game ban risks with virtualization); host isolation (VM vs container vs bare metal); certificate management (mTLS between components); rate limiting; DDoS protection; audit logging.
**Overlap**: Dim 07 (host agent), Dim 08 (infrastructure).

### Dim 10 — White-Label, Theming & Customization Architecture
**Angle**: Business/customization perspective — enabling branding and theme customization.
**Scope**: Theme engine design (JSON/YAML theme tokens); runtime theme switching; day/dark/auto modes; color scheme generation; font system; logo/branding asset injection; layout configuration (grid density, list vs card view); per-tenant configuration storage; theme compilation/packaging; client-side theme caching; accessibility (contrast ratios, WCAG compliance).
**Overlap**: Dim 04 (client frameworks), Dim 06 (catalog UI), Dim 11 (TV UI).

### Dim 11 — TV-First UI/UX & Living Room Experience
**Angle**: TV/living room perspective — optimizing for 10-foot experience on Android TV.
**Scope**: Android TV Leanback library; D-Pad/remote navigation; voice search integration; overscan safe areas; 4K UI rendering on TV; controller-driven UI (no touch); PS4/Xbox dashboard paradigms; shelf/carousel layouts; game trailer auto-play; quick resume UI; HDMI-CEC integration; picture-in-picture considerations; performance on low-end TV SoCs.
**Overlap**: Dim 04 (client frameworks), Dim 10 (theming), Dim 06 (catalog).

### Dim 12 — Performance Optimization & End-to-End Latency Engineering
**Angle**: Systems engineering perspective — achieving "zero lag" experience.
**Scope**: End-to-end latency budget breakdown (capture, encode, network, decode, display); frame time analysis; network QoS (DSCP marking); UDP vs TCP tradeoffs; forward error correction (FEC); jitter buffers; client-side frame interpolation; input batching vs immediate send; display refresh rate matching (VRR, G-Sync, FreeSync); 120Hz/144Hz streaming feasibility; bandwidth requirements for 4K60/4K120; latency measurement methodology.
**Overlap**: All other dimensions.
