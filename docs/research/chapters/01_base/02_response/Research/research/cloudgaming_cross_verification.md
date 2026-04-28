# Cross-Verification Report: Cloud Gaming System Architecture

## Methodology
Cross-verification analyzed 12 dimension research files (684KB, 970 sections, 1196+ citations) for overlapping claims, conflicting evidence, and confidence classification.

---

## High Confidence Findings (Confirmed by ≥2 dimensions from independent sources)

### HC-01: WebRTC with Pion is the Optimal Streaming Stack for Go
- **Confirmed by**: Dim01 (Streaming Protocols), Dim05 (Go APIs), Dim12 (Latency)
- **Evidence**: Pion WebRTC v4 is pure Go with no CGO, supports all platforms including WASM [^17^]. Sub-500ms glass-to-glass achievable [^104^]. DataChannels support unreliable/unordered mode for game input [^201^]. Parsec's BUD shows custom UDP can achieve 7ms LAN but WebRTC wins on ecosystem and NAT traversal [^81^].
- **Synthesis**: Use Pion WebRTC for streaming with unreliable DataChannels for input. Fallback to UDP for native desktop clients seeking minimal latency.

### HC-02: H.264 Baseline as Default, AV1 for Bandwidth-Constrained Scenarios
- **Confirmed by**: Dim01 (Codecs), Dim03 (Capture/Encode), Dim12 (Bandwidth)
- **Evidence**: H.264 has 98% browser/device support, mandatory in WebRTC spec [^2^]. NVENC achieves ~5.8ms median encode latency [^5^]. AV1 delivers 40-55% bandwidth savings but requires RTX 40+/Intel Arc/M3+ for hardware encode [^55^]. H.265 has complex 3-pool licensing [^3^].
- **Synthesis**: Multi-codec strategy: H.264 Baseline for universal compatibility, HEVC for native clients with hardware decode, AV1 as forward-looking option.

### HC-03: No Single Go Framework Covers All Client Platforms
- **Confirmed by**: Dim04 (Go Ecosystem), Dim11 (TV UI)
- **Evidence**: Fyne (desktop+mobile, no TV/D-Pad, no accessibility) [^37^]. Wails (desktop only, excellent) [^26^]. Gio (all platforms but complex) [^19^]. Flutter has full accessibility + TV support [^24^].
- **Synthesis**: Hybrid client architecture is mandatory: Wails (desktop) + Flutter+Go FFI (mobile/TV) + Angular+Go WASM (web).

### HC-04: Sunshine Architecture as Host Agent Reference
- **Confirmed by**: Dim01 (Moonlight/Sunshine), Dim03 (Capture), Dim07 (Host Agent)
- **Evidence**: Sunshine is single-binary host agent with capture, encode, streaming [^75^]. Uses DXGI DDA (Win), VideoToolbox (macOS), KMS/VA-API (Linux). Open-source, actively maintained, supports NVENC/AMF/QuickSync/VAAPI.
- **Synthesis**: Fork/adapt Sunshine architecture for host agent. Multi-process variant (like Moonshine) for session isolation.

### HC-05: Platform-Specific Capture APIs Are Mature
- **Confirmed by**: Dim03 (Capture), Dim07 (Host Agent), Dim12 (Latency)
- **Evidence**: Windows: DXGI DDA with dirty rectangles, HDR via DuplicateOutput1 [^137^][^162^]. macOS: ScreenCaptureKit + IOSurface zero-copy [^ScreenCaptureKit^]. Linux: DMA-BUF + PipeWire/KMS [^DMABUF^].
- **Synthesis**: Implement per-OS capture module with unified frame output interface. Zero-copy to encoder is critical.

### HC-06: Input Pipeline Architecture — UDP/DataChannels + Virtual Controller
- **Confirmed by**: Dim02 (Controllers), Dim01 (WebRTC DataChannels), Dim12 (Latency Budget)
- **Evidence**: Custom binary serialization (16-32 bytes) fastest. WebRTC DataChannels in unreliable mode avoid HOL blocking. Windows: ViGEmBus (virtual controller, retired — successor needed). Linux: uinput native. macOS: foohid kext.
- **Synthesis**: Implement binary input protocol over UDP for native clients, WebRTC DataChannels for web. Per-OS virtual controller driver on host.

### HC-07: OAuth2/OIDC + JWT + DTLS-SRTP for Security
- **Confirmed by**: Dim05 (Go APIs), Dim09 (Security)
- **Evidence**: OAuth2/OIDC standard for gaming devices. Device Authorization Grant (RFC 8628) for TV/console input-constrained devices [^RFC8628^]. DTLS-SRTP provides mandatory AES-128 encryption for WebRTC [^RFC5764^].
- **Synthesis**: Auth0/Keycloak/Authentik for identity. Short-lived JWT access tokens + refresh rotation. mTLS between internal services.

### HC-08: Design Token Architecture for White-Label
- **Confirmed by**: Dim06 (Catalog UI), Dim10 (Theming)
- **Evidence**: CSS custom properties outperform CSS-in-JS for runtime switching [^606^]. Material Design 3 tonal palette generates complete themes from single source color [^612^]. Style Dictionary v4 is industry standard [^604^].
- **Synthesis**: 3-tier token system (primitive → semantic → component) with CSS custom properties. Runtime theme API with per-tenant config.

### HC-09: Sub-50ms LAN Latency Achievable
- **Confirmed by**: Dim01 (Protocols), Dim12 (Latency Engineering)
- **Evidence**: Parsec achieves 4-8ms at 240Hz LAN [^1^]. Moonlight optimized at 15.7ms. Sunshine 12.6-26.7% lower than alternatives [^5^]. Total budget: ~33-110ms typical, <30ms competitive with optimization.
- **Synthesis**: Target <30ms LAN, <50ms WAN for competitive gaming. Hardware encode/decode + edge nodes + DSCP QoS.

### HC-10: Anti-Cheat is a Major Architectural Constraint
- **Confirmed by**: Dim07 (Host Agent), Dim09 (Security)
- **Evidence**: EAC/BattlEye/Vanguard are kernel-level and may block capture hooks [Dim07]. VM-per-session provides best isolation [Dim09]. Sunshine/Parsec are generally compatible but some games may block.
- **Synthesis**: Document anti-catch compatibility. Offer VM-per-session for highest compatibility. Monitor anti-cheat evolution.

---

## Medium Confidence Findings (Single authoritative source)

### MC-01: Flutter + Go FFI for Mobile/TV Clients
- **Source**: Dim04
- **Evidence**: Flutter has best accessibility, TV/D-Pad support, and Go FFI via c-shared compilation. Full platform coverage from single Dart codebase.
- **Risk**: FFI adds complexity; Go c-shared has mobile packaging challenges.

### MC-02: NATS JetStream as Event Bus
- **Source**: Dim05
- **Evidence**: Sub-ms core latency, 800K msg/s throughput. Superior to Redis/RabbitMQ for gaming event patterns.
- **Risk**: Operational complexity of NATS cluster vs simpler Redis.

### MC-03: CockroachDB for Multi-Region Database
- **Source**: Dim08
- **Evidence**: Follower reads provide 8x latency improvement. PostgreSQL-compatible.
- **Risk**: Operational complexity; PostgreSQL with read replicas may suffice for smaller scale.

### MC-04: Intel QuickSync Lowest Encoder Latency
- **Source**: Dim01, Dim12
- **Evidence**: 5 frames (83ms) in ULL mode for HEVC/AV1. NVIDIA NVENC ~7 frames.
- **Risk**: Quality tradeoffs at ULL preset; AMD ecosystem support weaker.

### MC-05: Compose for TV as Android TV Framework
- **Source**: Dim11
- **Evidence**: Leanback officially deprecated. Compose for TV at stable 1.0 with focus APIs.
- **Risk**: Newer framework, smaller community than Leanback.

---

## Conflict Zones

### CZ-01: WebRTC vs Custom UDP Protocol
- **Dim01**: Parsec BUD achieves 7ms LAN — custom UDP outperforms WebRTC for pure latency
- **Dim05**: WebRTC/Pion offers standards compliance, built-in NAT traversal, browser support
- **Dim12**: WebRTC adds ~10-20ms overhead vs raw UDP
- **Resolution**: Use hybrid — WebRTC for web clients and as default; custom UDP available for native desktop clients. Implement both behind abstraction layer.

### CZ-02: Fyne Viability for Mobile
- **Dim04**: Rates Fyne as having "significant accessibility gaps and no TV/D-Pad support" — effectively excludes it for inclusive/TV use
- **Phase 1**: Fyne community reports positive mobile experiences
- **Resolution**: Exclude Fyne from TV/mobile clients. Consider only for desktop if Wails is not suitable.

### CZ-03: Game Suspension Feasibility
- **Dim07**: Windows NtSuspendProcess works; Linux cgroups freeze works; macOS SIGSTOP works
- **Dim07**: Anti-cheat may flag suspension as tampering; CRIU doesn't preserve GPU state
- **Resolution**: Mark as Phase 2 feature. Implement graceful game-switching (return to catalog, game continues running background) as Phase 1.

### CZ-04: Bluetooth Controller Latency
- **Dim02**: Bluetooth HOGP polls at 125Hz (8ms) — too slow for competitive play
- **User requirement**: Wireless controller support via Bluetooth
- **Resolution**: Support both Bluetooth (convenience) and USB/2.4GHz dongle (competitive). Document latency tradeoffs.

### CZ-05: Bare Metal vs Cloud for Hosts
- **Dim08**: Bare metal 45-90% cheaper; cloud offers flexibility
- **Dim09**: VM-per-session requires hardware virtualization; cloud GPU instances have limited GPU passthrough
- **Resolution**: Hybrid — bare metal GPU servers as primary, cloud for burst/failover.

---

## Confidence Summary

| Tier | Count | Coverage |
|------|-------|----------|
| High Confidence | 10 findings | Core architecture decisions |
| Medium Confidence | 5 findings | Technology selections |
| Conflict Zone | 5 items | Require architectural decisions |
| Low Confidence | 0 | None identified |

**Overall Assessment**: The research provides strong, multi-source support for all major architectural decisions. Conflict zones are primarily optimization tradeoffs rather than fundamental disagreements. No critical unresolved conflicts.
