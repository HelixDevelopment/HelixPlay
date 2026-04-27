# Executive Summary

## Project Overview

CloudStream is a comprehensive self-hosted cloud gaming platform that enables users to remotely play PC games from personal or organizational host machines running macOS, Linux, or Windows. The system streams gameplay video and audio to client applications on Desktop (Windows/macOS/Linux), Mobile (iOS/Android), Web (browsers), and Android TV — all with the primary goal of delivering a PlayStation 4 Pro-like user experience with imperceptible latency, 4K resolution support, and maximum available refresh rates.

The platform is architected around Go as the primary programming language, with a novel "three-client-one-core" approach: a shared `cloudstream-core` library written in Go is compiled differently for each client platform — as a native binary for Wails-based desktop applications, as a c-shared library for Flutter mobile and TV apps, and as WebAssembly for Angular-based web clients. This design maximizes code reuse while allowing each platform to use the most appropriate UI framework.

## Key Architectural Decisions

**Streaming Protocol**: WebRTC (via Pion, a pure Go implementation) serves as the primary streaming protocol for its built-in NAT traversal, browser compatibility, and sub-500ms latency capability. A custom UDP protocol inspired by Moonlight's ENet-based implementation is available as a fallback for native desktop clients seeking absolute minimum LAN latency (7-15ms).

**Video Codecs**: H.264 Baseline Profile is the pragmatic default for universal hardware decode support across all devices. HEVC (H.265) provides 50% bandwidth reduction for native clients with hardware decode, and AV1 offers a royalty-free future path with 30-50% additional compression — requiring RTX 40-series, Intel Arc, or Apple M3+ for hardware encoding.

**Host Capture**: Platform-specific zero-copy capture pipelines are implemented for each OS — DXGI Desktop Duplication API with CUDA interop on Windows, ScreenCaptureKit with IOSurface on macOS, and PipeWire with DMA-BUF on Linux. Each path achieves <3ms capture overhead at 4K60.

**Controller Input**: A custom 24-byte binary input protocol transmits controller state over UDP (native) or WebRTC DataChannels (web), achieving 4-17ms end-to-end input latency on LAN. The system supports advanced features including DualSense adaptive triggers, gyro aiming via the CemuhookUDP protocol, and haptic feedback forwarding.

**Infrastructure**: NATS JetStream powers the event bus (800K messages/second), CockroachDB provides horizontally-scalable PostgreSQL-compatible persistence, and regional TURN relay clusters ensure connectivity for the 20-30% of sessions that cannot establish direct peer-to-peer connections.

## Performance Targets

The system targets sub-30ms glass-to-glass latency for competitive LAN gaming and sub-50ms for WAN scenarios. At 4K60, H.264 streaming requires 35-50 Mbps bandwidth; HEVC reduces this to 15-25 Mbps. The adaptive bitrate controller responds to network changes within 2 seconds using a 3-tier quality ladder (4K/1080p/720p).

## Implementation Roadmap

Development is organized into six phases over 26 weeks: Foundation (workspace, core library, host agent MVP), Core Streaming (WebRTC integration, input pipeline, capture optimization), Client Applications (Wails desktop, Flutter mobile/TV, Angular web), Platform Services (catalog, theming, user management), Infrastructure & Scale (Kubernetes, multi-region, monitoring), and Polish & Production (performance optimization, security hardening, advanced features). The complete roadmap comprises 116 tasks totaling approximately 1,824 engineering hours.

## White-Label & Theming

The platform includes a comprehensive theming system built on the W3C Design Tokens Community Group specification, with three-tier tokens (primitive, semantic, component) transformed via Style Dictionary v4. Runtime theme switching between Day, Dark, and Auto modes uses CSS custom properties for zero-latency transitions. A self-service brand portal enables white-label customers to upload logos, select colors, and preview their branded experience in real time — supporting a "Gaming-as-a-Service" business model for ISPs, hotels, hospitals, and enterprises.

## Document Structure

This technical specification comprises 15 chapters organized into five parts: Foundation & Architecture (Chapters 1-3), Core Platform Systems (Chapters 4-7), Platform Services (Chapters 8-9), Infrastructure & Operations (Chapters 10-12), and Delivery & Risk Management (Chapters 13-15). The document includes 55+ data tables, 22 technical diagrams, Go code examples, wireframe specifications for all major screens, and a detailed risk analysis with mitigation strategies.
