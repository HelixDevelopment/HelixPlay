# Comprehensive Research Synthesis: Video Technology for Cloud Gaming Platform

## Document Metadata
- **Synthesis Date**: July 2025
- **Source Dimensions**: 12 research dimensions + cross-verification + insights
- **Total Research Files**: 14 artifacts
- **Sources Consulted**: 300+ primary sources (IEEE/ACM papers, RFCs, vendor documentation, open-source projects)

---

## 1. KEY TECHNICAL FINDINGS (Grouped by Theme)

### Theme A: Codec Architecture & Selection

**Finding A1: H.264 remains the strategic cornerstone despite being technically inferior**
- 98.2% device decode coverage vs AV1 at ~25% (dim01)
- Only mandatory WebRTC codec per RFC 7742 (dim01, dim11, dim12)
- Most consistent latency across all GPU presets (dim02)
- Modern GPUs encode with near-zero overhead (dim02)
- **Insight**: H.264 is strategically optimal, not technically superior (insight_03)

**Finding A2: AV1 is the undisputed next-gen codec for near-term investment**
- 40-55% bandwidth savings over H.264 (dim08)
- Hardware encode available on RTX 40+, Arc, RDNA3+, M3+ (dim02)
- AV1 hardware adoption expected majority by 2028 (dim01)
- VVC's 8-10x encoding complexity makes it non-viable for real-time before 2028-2030 (insight_08)
- Adds 2-3 frames encode latency vs HEVC on NVENC (CZ-2: vendor-dependent, AMD shows no penalty)

**Finding A3: HEVC serves as the standard tier codec**
- HEVC Main 10 profile is the most mature for 10-bit HDR encoding (dim07)
- NVENC HEVC = ~7 frames (117ms) at 4K60 consistent latency (dim02)
- Available on all major GPU vendors (dim02)
- HDR10+ dynamic metadata supported via SEI (dim07)

**Finding A4: Codec selection should follow a 3-tier strategy**
- **Tier 1 (Universal Fallback)**: H.264 — 100% compatibility
- **Tier 2 (Standard)**: HEVC — best compression with broad support
- **Tier 3 (Premium)**: AV1 — best efficiency, limited hardware decode
- Multi-codec negotiation with H.264 as guaranteed fallback (insight_03)

### Theme B: Hardware GPU Encoding

**Finding B1: Intel ULL achieves lowest latency at 83ms (5 frames)**
- Intel Quick Sync Video ULL mode: 5 frames for HEVC/AV1 at 4K60 (dim02, cross-verification HC-2)
- Non-standard B-frame structure may affect decoder compatibility (CZ-1)
- No concurrent encoding session limits (dim09)

**Finding B2: NVIDIA NVENC provides most consistent latency at ~117ms (7 frames)**
- ~7 frames across nearly all presets and codecs (dim02, HC-3)
- 7 presets (P1 fastest / P7 best quality) (dim09)
- Consumer GPUs: 2-8 concurrent sessions (evolved: 2→3→5→8 over time) (CZ-4)
- RTX 50 9th Gen: 4:2:2 10-bit, ~5% quality gain, 60% faster than RTX 4090 (HC-14)
- **Thermal critical**: 83°C throttling reduces throughput 25-30% (dim09, insight_01)

**Finding B3: AMD RDNA4 brings major quality improvements**
- 25% H.264 low-latency quality improvement, 11% HEVC improvement (HC-13)
- AV1 B-frame support, dual media engines (dim02)
- No session limits — concurrency-first choice (insight_09)

**Finding B4: GPU vendor selection should be topology-driven**
- **Intel QSV**: latency-first, single-host, no session limits (dim02, insight_09)
- **NVIDIA NVENC**: scale/reliability-first, most consistent, best tooling (dim02)
- **AMD RDNA4**: cost/concurrency-first, no session limits (dim02)

**Finding B5: Zero-copy GPU pipelines reduce latency 3-10x**
- Eliminating CPU-GPU memory copies reduces per-frame latency by 1-3ms per copy (HC-4)
- Pipeline from 200-500ms → 10-30ms with zero-copy (dim03, dim04)
- Confirmed by NVIDIA Jetson, Sunshine/Moonlight, NETINT case studies (dim03)

### Theme C: Simultaneous Streaming + Recording

**Finding C1: Dual-path encoding is viable with hardware encoders**
- NVENC supports 2+ sessions minimum for stream+record (HC-5)
- OBS demonstrates this for millions of users daily (HC-10)
- FFmpeg tee muxer enables single-encode/multi-output (dim04)
- MKV is best container for crash-safe recording (HC-6)

**Finding C2: Recording storage should mirror game save systems**
- "Local buffer + background sync" model is most reliable (insight_04)
- Always record to local NVMe SSD first (dim05)
- Background uploader with retry, resume, bandwidth limiting (dim05)
- Network storage is DESTINATION, not recording target (insight_04)

**Finding C3: Thermal wall is the hidden bottleneck for dual-path**
- Dual encoding increases GPU power draw by 15-25W (insight_01)
- Thermal throttling at 83°C reduces throughput 25-30% (dim09)
- Proactive thermal-aware quality reduction BEFORE throttling (insight_01)

### Theme D: Adaptive Bitrate & Congestion Control

**Finding D1: Frame-level adaptation outperforms segment-based ABR for gaming**
- Traditional HLS/DASH segment-based ABR: 2-5 second latency penalty (dim08)
- Cloud gaming requires sub-100ms E2E latency (dim08)
- Salsify: 4.6x delay reduction, 60% SSIM improvement over WebRTC (dim08)
- Adaptation priority: compression params → frame rate → resolution (last resort)

**Finding D2: SQP congestion control outperforms GCC by 2-3x**
- SQP: 2-3x higher bandwidth than GoogCC when competing with TCP (HC-11)
- Google's AR streaming platform production A/B testing (dim08)
- Pudica (Tencent START): 3.1x avg frame delay reduction, 5.7x on WiFi (dim08)
- Camel (250M users): +70.8% 1080P resolution ratio (dim08)

**Finding D3: Gaming-specific ABR differs fundamentally from VoD**
- VoD ABR: segment-level (2-10s), 10-30s buffer (dim08)
- Gaming ABR: frame-level (16.7ms @ 60fps), 1-3 frame buffer (dim08)
- Encoder reconfigured per-frame, not pre-encoded ladder (dim08)
- AFR (Adaptive Frame Rate): 7.4x tail queuing delay reduction (dim08)

**Finding D4: FEC strategy must be network-condition dependent**
- RTT < 30ms: NACK/RTX only (dim12)
- RTT 30-80ms: Light FEC (10-20%) + RTX (dim12)
- RTT > 80ms: Heavy FEC (20-25%) primary (dim12)
- Reed-Solomon: constant 100% recovery vs XOR FEC bounded by O(k²/2^k) (dim12)
- klauspost/reedsolomon: >15GB/s per core with SIMD (dim12)

### Theme E: Audio Pipeline

**Finding E1: Opus is the optimal real-time audio codec**
- 5ms frame sizes, <20ms end-to-end audio latency achievable (HC-7)
- Supports up to 255 channels via MultiStream (dim06)
- Royalty-free, IETF standard (RFC 6716), mandatory in WebRTC (dim06)
- Recommended: 96-128 kbps stereo, 192-256 kbps 5.1, 256-450 kbps 7.1 (dim06)

**Finding E2: Audio passthrough is more constrained than video**
- Audio has binary capability threshold (full surround vs stereo collapse) (insight_02)
- eARC is ONLY consumer interface for uncompressed multi-channel + Atmos (dim06)
- SPDIF limited to compressed 5.1 (AC3/DTS) (dim06)
- Windows 7.1 channel order differs from Dolby/DTS/SMPTE standard (dim06)
- No native Atmos transport in WebRTC (dim12)

**Finding E3: Go audio libraries are viable**
- `oto`: low-level, cross-platform, no Cgo on Windows/macOS (dim06)
- `malgo`: miniaudio bindings, multi-backend, multi-channel support (dim06)
- `gopxl/beep`: high-level audio processing with Streamer interface (dim06)

### Theme F: HDR & Color Space

**Finding F1: HDR10+ is recommended over Dolby Vision for cloud gaming**
- HDR10+: royalty-free, live encoder support, GPU-accelerated (HC-8)
- Dolby Vision: $2.5K/year licensing, no consumer GPU encoder support (dim07)
- HEVC Main 10 / AV1 10-bit provide viable 10-bit encoding (dim07)

**Finding F2: SDR fallback remains critical**
- HLG: inherent SDR backward compatibility, no metadata needed (dim07)
- HDR10+: good fallback via HDR10 base (dim07)
- Client capability negotiation via WebRTC color space RTP extension (dim07)
- Tone mapping: Hable, ACES, BT.2390 available via libplacebo (dim07)

**Finding F3: HDR capture requires platform-specific handling**
- Windows: DXGI R16G16B16A16_FLOAT + scRGB (dim07)
- macOS: MTLPixelFormat.rgba16Float + extendedLinearDisplayP3 (dim07)
- Linux: Vulkan DMA-BUF import with color metadata (dim07)

### Theme G: Network Transport & Packet Optimization

**Finding G1: WebRTC protocol overhead is 15-20ms vs raw UDP**
- DTLS/SRTP/ICE adds 70-110 bytes per-packet overhead (dim12)
- Parsec BUD achieves 7ms LAN latency (vs WebRTC's 15-20ms) (dim12)
- **Insight**: SQP + custom UDP could achieve sub-10ms on LAN (insight_07)

**Finding G2: MTU considerations are critical**
- WebRTC default MTU: 1200 bytes payload (avoids VPN fragmentation) (dim12)
- Cloudflare One requires minimum 1361 bytes (dim12)
- Effective payload after all headers: ~1146 bytes at 1200 MTU (dim12)

**Finding G3: DSCP marking can help on local networks**
- DSCP 46 (EF) for gaming traffic prioritization (dim12)
- Typically stripped by ISPs — benefits mainly on home network (dim12)
- Xbox uses DSCP 46 on outbound UDP packets (dim12)

### Theme H: Hardware Detection & Dynamic Optimization

**Finding H1: Go has robust GPU monitoring libraries**
- `go-nvml`: NVIDIA GPU detection, temperature, encoder utilization (dim09)
- `go-rocm-smi`: AMD GPU monitoring (dim09)
- `gopsutil`: cross-platform CPU, memory, disk info (dim09)
- `golang.org/x/sys/cpu`: runtime CPU feature detection (dim09)

**Finding H2: Dynamic quality adjustment is essential**
- Preset escalation: P7 → P1 as thermal increases (dim09)
- Resolution reduction: 4K → 1080p → 720p (dim09)
- Bitrate reduction: 25-50% when constrained (dim09)
- Codec fallback: HEVC → H.264 (lower encoding complexity) (dim09)

### Theme I: Testing & Validation

**Finding I1: Comprehensive testing framework identified**
- Frame integrity: counter injection, timestamp validation, sequence gaps (dim10)
- Latency: LED+photodiode G2G (0.5ms precision), PresentMon, slow-mo (dim10)
- Quality: VMAF > SSIM > PSNR for perceptual accuracy (dim10)
- A/V sync: ITU-R BS.1359 threshold (45ms audio lead) (dim10)
- Network resilience: tc/netem for packet loss, jitter, bandwidth (dim10)

**Finding I2: Go testing patterns validated**
- Table-driven tests with `t.Run` for isolation (dim10)
- testify `assert` vs `require` distinction (dim10)
- Benchmark tests with `b.ReportAllocs()` (dim10)
- Race detection in CI (`go test -race`) (dim10)
- Coverage targets: encoder core 90%+, muxer 85%+, config 95%+ (dim10)

### Theme J: Cross-Platform Go Implementation

**Finding J1: Go's goroutine model maps perfectly to video pipelines**
- Each pipeline stage (capture → encode → transmit) = goroutine (insight_05)
- Channels provide lock-free frame passing (dim11)
- `sync.Pool` for `[]byte` frame buffers eliminates GC pressure (dim11)
- Ring buffer: 200M+ writes/sec, ~5ns/op (dim11)

**Finding J2: CGO vs Pure Go trade-offs are well-understood**
- CGO call overhead: ~40ns single core, ~4ns with 16 cores (dim11)
- Go 1.26: CGO overhead reduced ~30% (dim11)
- Screen capture requires CGO (DXGI, ScreenCaptureKit, PipeWire) (dim11)
- WebRTC signaling: Pure Go (Pion) — no CGO needed (dim11)

**Finding J3: Recommended Go stack validated**
- WebRTC: Pion v4 (pure Go, production-ready) (HC-12, dim11)
- Screen capture: Platform CGO + FFmpeg (dim11)
- Video encode: FFmpeg CLI or go-astiav (CGO) (dim11)
- Frame pipeline: Go channels + sync.Pool (dim11)
- Controller input: Pion DataChannels (unreliable/unordered mode) (dim11)

---

## 2. DATA POINTS & BENCHMARKS THAT MUST APPEAR

### Latency Benchmarks (Critical for Report)

| Encoder | Codec | Latency (frames) | Latency (ms @ 60fps) | Source |
|---------|-------|-----------------|---------------------|--------|
| Intel ULL | HEVC/AV1 | 5 frames | 83ms | IEEE 2025 (HC-2) |
| Intel ULL | H.264 | 8 frames | 133ms | IEEE 2025 |
| NVIDIA NVENC | All | ~7 frames | ~117ms | IEEE 2025 (HC-3) |
| AMD RDNA3+ | H.264 | 9 frames | 150ms | IEEE 2025 |
| AMD RDNA3+ | HEVC | 6 frames | 100ms | IEEE 2025 |
| Software (x264) | H.264 | 41-90+ frames | 683-1500ms | IEEE 2025 |

### Quality Tier Bandwidth Requirements

| Resolution | FPS | Minimum Bandwidth | Service Reference |
|-----------|-----|-------------------|-------------------|
| 720p | 60 | 10-15 Mbps | GeForce NOW Entry (dim08) |
| 1080p | 60 | 20-28 Mbps | GeForce NOW Standard / Stadia (dim08) |
| 1440p | 60 | 35 Mbps | Xbox Cloud Ultimate (dim08) |
| 4K | 60 | 35-45 Mbps | GeForce NOW Ultimate (dim08) |
| 4K | 120 | 65 Mbps | GeForce NOW Ultimate (dim08) |

### GPU Encoder Session Limits

| GPU Generation | Concurrent Sessions | Notes |
|---------------|-------------------|-------|
| RTX 20/30 series | 2 | Original limit (dim09) |
| RTX 40 series | 3 | Baseline (dim09) |
| RTX 4070 Ti+ | Dual physical NVENCs | Effectively double (dim09) |
| Post-2024 drivers | 5-8 | Expanded limits (CZ-4) |
| Intel QSV | No limit | As many as resources allow (dim09) |
| AMD VCN | No limit | As many as resources allow (dim09) |

### Audio Latency Budget

| Stage | Target | Optimization |
|-------|--------|-------------|
| Audio capture | 1-2ms | WASAPI exclusive / PipeWire quantum=128 (dim06) |
| Opus encoding | 2.5-5ms | 5ms frames, complexity=10 (dim06) |
| Network transmission | 5-10ms | UDP, jitter buffer=5ms (dim06) |
| Opus decoding | 1-2ms | Optimized decoder (dim06) |
| Audio render | 2-5ms | WASAPI exclusive / low-quantum PipeWire (dim06) |
| **Total** | **<20ms** | Aggressive optimization (dim06) |

### Protocol Overhead Comparison

| Protocol | Latency | Transport | Gaming Suitability |
|----------|---------|-----------|-------------------|
| Parsec BUD | ~7ms LAN | UDP+DTLS | Best for gaming (dim12) |
| Moonlight | ~10-20ms | UDP+RTP | Good (dim12) |
| WebRTC | 15-20ms overhead | UDP+SRTP | Standard (HC-9) |
| RIST | Configurable | UDP+RTP | Broadcast (dim12) |

---

## 3. ARCHITECTURE PATTERNS IDENTIFIED

### Pattern 1: Goroutine Pipeline Stages
- Each processing stage = independent goroutine
- Channels provide typed, synchronized communication
- `sync.Pool` for frame buffer recycling
- Buffered channels (capacity 1-3 frames) for backpressure
- Ring buffer for single-writer multi-reader fan-out (200M+ writes/sec)

### Pattern 2: Three-Client-One-Core
- Shared Go Core: protocol, session, catalog, streaming client, theme
- Native UI (Desktop): Go + CGO for capture/encode
- Mobile: c-shared library with JNI/Swift bridge
- Browser: TinyGo WASM + JS shim

### Pattern 3: Thermal-Aware Quality Controller
- Continuous GPU monitoring via NVML/ROCm SMI
- Proactive quality reduction BEFORE throttling
- Preset escalation, resolution scaling, bitrate reduction
- Session routing to thermally advantaged hosts

### Pattern 4: Local-First Recording with Background Sync
- Record to local NVMe SSD first (crash-safe)
- Circular buffer for instant replay (30-min rolling)
- Background uploader: retry, resume, bandwidth limiting
- Network storage as destination, not recording target

### Pattern 5: Frame-Coupled Congestion Control
- Bandwidth estimation tied to frame boundaries
- TWCC feedback at 50-100ms intervals
- Bitrate adjustment at 50-100ms (faster than GCC's 1-5s)
- Frame rate adjustment (AFR) as secondary mechanism

### Pattern 6: Platform Abstraction via Build Tags
```
capture_windows.go  // DXGI + CGO
capture_linux.go    // PipeWire + Portal
capture_darwin.go   // ScreenCaptureKit
```
- Interface-based design: `Capturer` interface with platform implementations
- Clean separation of platform-specific code

### Pattern 7: Dual Transport Architecture
- Custom UDP (Parsec BUD-style) for LAN: sub-10ms
- WebRTC for WAN/browser: universal compatibility
- SQP congestion control for custom UDP path
- DTLS 1.2 per-packet encryption for LAN path

### Pattern 8: 3-Tier Codec Negotiation
1. **Premium tier**: AV1 — best compression, limited decode support
2. **Standard tier**: HEVC — good compression, broad support
3. **Universal fallback**: H.264 — 100% compatibility
- Dynamic codec selection based on client capability + network conditions

### Pattern 9: Zero-Copy GPU Pipeline
```
GPU Render → GPU Capture → GPU Encode → GPU Mux → Network
     (no CPU round-trips for frame data)
```
- Eliminate CPU-GPU memory copies entirely
- Use GPU memory-mapped buffers
- NVENC/Quick Sync/VCN direct from GPU framebuffer

---

## 4. TECHNOLOGY RECOMMENDATIONS WITH CONFIDENCE LEVELS

### HIGH Confidence (Confirmed by multiple independent sources)

| Technology | Recommendation | Confidence | Evidence |
|-----------|---------------|------------|----------|
| **H.264 as fallback codec** | Mandatory universal fallback | HIGH | 98.2% decode coverage, WebRTC mandatory (HC-3) |
| **AV1 as premium codec** | Primary near-term next-gen investment | HIGH | 40-55% bandwidth savings, hardware available (HC-8) |
| **Skip VVC for real-time** | Not viable before 2028-2030 | HIGH | 8-10x complexity, no browser support (insight_08) |
| **Intel ULL for lowest latency** | Best for competitive gaming | HIGH | 83ms at 4K60 (HC-2) |
| **NVIDIA NVENC for consistency** | Best for multi-host cloud | HIGH | Most consistent across presets/codecs (HC-3) |
| **Opus for audio** | Only real-time audio codec | HIGH | <20ms achievable, 255 channels (HC-7) |
| **HDR10+ over Dolby Vision** | Royalty-free, GPU supported | HIGH | No licensing, live encoder support (HC-8) |
| **MKV for recording** | Crash-safe progressive writing | HIGH | Universally acknowledged (HC-6) |
| **Pion WebRTC** | Production-ready pure-Go | HIGH | 13k+ stars, no CGO (HC-12) |
| **SQP congestion control** | 2-3x better than GCC | HIGH | Google Research + ACM verified (HC-11) |
| **Zero-copy GPU pipeline** | 3-10x latency reduction | HIGH | Multiple independent implementations (HC-4) |
| **Local-first recording** | Mirror game save pattern | HIGH | Network storage as destination (insight_04) |
| **sync.Pool for frame buffers** | Eliminate GC pressure | HIGH | 2-5x throughput, zero allocations (dim11) |
| **go-astiav for FFmpeg CGO** | Most maintained FFmpeg binding | HIGH | Active, FFmpeg n8.0 compatible (dim11) |
| **go-nvml for NVIDIA monitoring** | Official NVIDIA Go bindings | HIGH | Runtime dynamic loading (dim09) |

### MEDIUM Confidence (Single authoritative source or emerging)

| Technology | Recommendation | Confidence | Evidence |
|-----------|---------------|------------|----------|
| **SQP + custom UDP for LAN** | Sub-10ms transport stack | MEDIUM | SQP is Google Research, limited availability (insight_07) |
| **HDR10+ Advanced for gaming** | Reduced cloud gaming latency | MEDIUM | Vendor claims without technical details (MC-2) |
| **AF_XDP kernel bypass** | 2.6M pps in Go | MEDIUM | Requires root, kernel BPF, significant complexity (MC-5) |
| **JPEG XS TDC** | Gaming-optimized low-latency | MEDIUM | Emerging standard, limited hardware (MC-6) |
| **PyroWave GPU compute codec** | 0.13ms encode | MEDIUM | Single developer, intra-only at 200+ Mbps (MC-1) |
| **FEC at 25% overhead** | 99.5% packet recovery | MEDIUM | Theoretically sound, loss-pattern dependent (MC-4) |

### LOW Confidence (Weak sourcing or speculative)

| Technology | Assessment | Confidence | Concern |
|-----------|-----------|------------|---------|
| **Custom GPU compute replacing ASICs** | Not viable short-term | LOW | Ignores power efficiency, ecosystem (LC-1) |
| **VVC by 2028** | Unlikely for real-time | LOW | 8-10x complexity barrier (LC-2) |
| **CGO overhead negligible** | Actually matters at scale | LOW | Compounds with high-frequency calls (LC-3) |

---

## 5. CRITICAL IMPLEMENTATION DETAILS FOR GO

### 5.1 Pipeline Goroutine Architecture
```go
// Stage 1: Capture (platform-specific CGO)
// Stage 2: Encode (FFmpeg CLI or CGO)
// Stage 3: Transmit (WebRTC via Pion)
// Stage 4: RTCP feedback handler
// Stage 5: Input event processor (DataChannels)

// Use buffered channels with capacity 1-3 frames for backpressure
// Use sync.Pool for []byte frame buffers
// Use ring buffer (golang-cz/ringbuf) for single-writer multi-reader
```

### 5.2 Memory Management Strategy
- **`sync.Pool`**: Pre-allocate frame buffers by resolution
  - 1080p RGBA: 1920×1080×4 = ~8MB
  - 1080p YUV420P: 1920×1080×1.5 = ~3MB
  - 4K RGBA: 3840×2160×4 = ~33MB
- **Zero-allocation hot paths**: Use pool for all frame data
- **GC tuning**: `GOGC=100` default; consider `GOGC=50` for lower latency
- **mmap for large files**: 2-6x faster than system calls (dim11)

### 5.3 CGO Integration Points
| Component | Approach | CGO Required | Rationale |
|-----------|----------|-------------|-----------|
| WebRTC signaling | Pion (pure Go) | No | No CGO needed |
| Screen capture | Platform shim | Yes | DXGI/ScreenCaptureKit/PipeWire |
| Video encoding | FFmpeg CLI or go-astiav | Optional | Hardware encoder access |
| Audio I/O | malgo or oto | Optional | Low-level audio access |
| GPU monitoring | go-nvml/go-rocm-smi | Yes | Vendor SDK access |

### 5.4 Cross-Compilation Strategy
- Pure Go components: `GOOS=windows GOARCH=amd64 go build` (trivial)
- CGO components: Docker-based cross-compilation with toolchains (dim11)
- musl-static for Linux targets (most reliable) (dim11)
- c-archive for iOS, c-shared for Android (dim11)

### 5.5 WebRTC Configuration for Gaming
```javascript
// Minimum playout delay
receiver.jitterBufferTarget = 0;

// Codec parameters
params.degradationPreference = "maintain-framerate";

// Playout delay RTP header extension
// min_delay = max_delay = 0 for gaming
```

### 5.6 Critical Go Libraries
| Library | Purpose | Version | Notes |
|---------|---------|---------|-------|
| `github.com/pion/webrtc/v4` | WebRTC | v4 | Pure Go, production-ready |
| `github.com/u2takey/ffmpeg-go` | FFmpeg CLI wrapper | latest | No CGO |
| `github.com/asticode/go-astiav` | FFmpeg CGO bindings | latest | FFmpeg n8.0 |
| `github.com/NVIDIA/go-nvml` | NVIDIA GPU monitoring | latest | Dynamic loading |
| `github.com/gen2brain/malgo` | Audio I/O | latest | miniaudio bindings |
| `github.com/hajimehoshi/oto` | Simple audio output | v2 | No Cgo on Win/Mac |
| `github.com/golang-cz/ringbuf` | Lock-free ring buffer | latest | 200M+ writes/sec |
| `github.com/shirou/gopsutil` | System info | v4 | Cross-platform |

### 5.7 Platform-Specific Capture
| Platform | API | CGO | Latency |
|----------|-----|-----|---------|
| Windows | DXGI Desktop Duplication | Yes | ~1ms |
| macOS | ScreenCaptureKit | Yes | ~3ms |
| Linux (Wayland) | PipeWire + xdg-portal | Yes | ~5ms |
| Linux (X11) | XShm | No | ~5ms |

### 5.8 Thermal Management Loop
```
every 5 seconds:
  for each GPU:
    temp = nvml.GetTemperature()
    throttle = nvml.GetThrottleReasons()
    
    if temp >= slowdown_threshold:
      reduce_preset()        // P7 → P5 → P3 → P1
      reduce_bitrate(25%)   // 8Mbps → 6Mbps
    
    if temp >= critical_threshold:
      reduce_resolution()    // 4K → 1080p
      reduce_fps()           // 60 → 30
```

---

## 6. TABLES AND DIAGRAMS TO INCLUDE IN FINAL REPORT

### Must-Include Tables

1. **Latency Comparison Table** (Section 2) — All encoder latencies at 4K60
2. **Quality Tier Bandwidth Requirements** (Section 2) — Service reference data
3. **GPU Encoder Session Limits** (Section 2) — Per-generation limits
4. **Audio Latency Budget** (Section 2) — Stage-by-stage breakdown
5. **Protocol Overhead Comparison** (Section 2) — Parsec vs Moonlight vs WebRTC
6. **High Confidence Recommendations** (Section 4) — 15 entries with evidence
7. **Go Library Reference** (Section 5.6) — Critical dependencies
8. **Platform Capture Options** (Section 5.7) — Latency by platform
9. **Codec 3-Tier Strategy** — H.264 / HEVC / AV1 with use cases
10. **Congestion Control Comparison** — GCC vs SQP vs Pudica vs Camel
11. **FEC Decision Matrix** — RTT-based strategy
12. **HDR Format Comparison** — HDR10 vs HDR10+ vs Dolby Vision vs HLG
13. **CGO vs Pure Go Trade-offs** — Build complexity comparison
14. **Confidence Tier Summary** — HC/MC/LC/CZ counts and percentages

### Must-Include Diagrams

1. **System Architecture Diagram** — Three-Client-One-Core pattern
   - Shared Go Core → Native UI / Mobile (c-shared) / Browser (WASM)

2. **Video Pipeline Flow** — Goroutine stage diagram
   - Capture → Encode → Transmit with channel connections

3. **Dual-Path Encoding** — Stream + Record simultaneously
   - Single encode → tee muxer → stream output + local recording

4. **Zero-Copy GPU Pipeline** — Memory flow diagram
   - GPU Render → GPU Capture → GPU Encode → Network (no CPU)

5. **Thermal Management State Machine** — Quality reduction transitions
   - Normal → Warning → Critical states with actions

6. **Congestion Control Architecture** — Frame-coupled CC diagram
   - TWCC feedback → bandwidth estimator → encoder reconfiguration

7. **Audio Pipeline Architecture** — End-to-end flow
   - Game Audio → Capture → Opus Encode → Network → Decode → Output

8. **Recording Storage Pattern** — Local-first + background sync
   - Local NVMe → circular buffer → background uploader → NAS

9. **Network Stack Comparison** — Protocol layer diagrams
   - WebRTC (DTLS/SRTP/RTP) vs Parsec BUD vs Custom UDP

10. **Codec Selection Decision Tree** — Dynamic negotiation flow
    - Check H.264 support → check HEVC → check AV1 → fallback chain

---

## 7. CONTENT PRIORITY RANKING

### P0: Critical — Must Appear in Executive Summary / Architecture Document

| # | Content | Source Dimensions | Rationale |
|---|---------|-----------------|-----------|
| 1 | H.264 as strategic fallback (98.2% coverage) | dim01, dim02, dim08, insight_03 | Universal compatibility |
| 2 | AV1 as near-term premium codec | dim01, dim02, dim08, insight_08 | 40-55% bandwidth savings |
| 3 | Intel ULL = 83ms, NVENC = 117ms latency | dim02, HC-2, HC-3 | Key differentiator |
| 4 | Zero-copy GPU pipeline = 3-10x latency reduction | dim03, dim04, HC-4 | Architecture-critical |
| 5 | Go goroutine pipeline architecture | dim11, insight_05 | Implementation pattern |
| 6 | Pion WebRTC (pure Go, production-ready) | dim11, dim12, HC-12 | Core technology |
| 7 | Thermal wall as hidden bottleneck | dim09, insight_01 | Reliability-critical |
| 8 | Local-first recording storage | dim05, insight_04 | Differentiating feature |
| 9 | SQP 2-3x outperforms GCC | dim08, HC-11 | Network performance |
| 10 | MKV for crash-safe recording | dim04, HC-6 | Data integrity |

### P1: Important — Must Appear in Technical Design Document

| # | Content | Source Dimensions | Rationale |
|---|---------|-----------------|-----------|
| 11 | HDR10+ over Dolby Vision | dim07, HC-8 | Licensing + GPU support |
| 12 | Opus audio <20ms E2E latency | dim06, HC-7 | Audio quality |
| 13 | Audio passthrough constraints | dim06, insight_02 | Multi-channel complexity |
| 14 | GPU vendor topology-driven selection | dim02, insight_09 | Deployment guidance |
| 15 | Frame-level ABR vs segment-based | dim08 | Gaming-specific design |
| 16 | Dual transport (LAN UDP + WAN WebRTC) | dim12, insight_07 | Latency optimization |
| 17 | go-astiav for FFmpeg CGO | dim11 | Video encoding integration |
| 18 | sync.Pool for frame buffers | dim11 | Go performance |
| 19 | Platform capture abstraction | dim11 | Cross-platform design |
| 20 | NVENC session limits (2-8) | dim09, HC-5 | Capacity planning |

### P2: Supporting — Appear in Detailed Implementation Guides

| # | Content | Source Dimensions | Rationale |
|---|---------|-----------------|-----------|
| 21 | VVC hardware gap timeline | dim01, insight_08 | Future planning |
| 22 | AFR (Adaptive Frame Rate) | dim08 | Queue management |
| 23 | Pudica (Tencent START) results | dim08 | Production validation |
| 24 | Reed-Solomon FEC performance | dim12 | Network resilience |
| 25 | MTU 1200 default for WebRTC | dim12 | Network optimization |
| 26 | DSCP 46 for gaming traffic | dim12 | QoS marking |
| 27 | HLG for SDR fallback | dim07 | HDR compatibility |
| 28 | VMAF for quality testing | dim10 | Validation |
| 29 | Table-driven Go tests | dim10 | Code quality |
| 30 | AF_XDP kernel bypass option | dim12 | Ultra-low latency path |

### P3: Reference — Appendix or Deep-Dive Documentation

| # | Content | Source Dimensions | Rationale |
|---|---------|-----------------|-----------|
| 31 | HEVC VPS/SPS/PPS NAL overhead | dim12 | Bitstream detail |
| 32 | AV1 OBU packetization rules | dim12 | Protocol detail |
| 33 | RTCP XR metrics | dim08 | Monitoring |
| 34 | WiFi 7 MLO for gaming | dim08 | Future networking |
| 35 | JPEG XS TDC profile | dim01 | Emerging codec |
| 36 | PyroWave GPU compute codec | dim01, MC-1 | Research direction |
| 37 | SCReAMv2 standardization | dim08 | Standards tracking |
| 38 | RIST protocol for broadcast | dim12 | Alternative transport |
| 39 | GPU hot-plug detection | dim09 | Advanced management |
| 40 | WebAssembly limitations | dim11 | Browser client constraints |

---

## 8. CROSS-DIMENSIONAL INSIGHT SUMMARY

From the insight analysis (10 cross-dimensional insights):

| # | Insight | Confidence | Key Dimensions |
|---|---------|-----------|----------------|
| 1 | **Thermal wall is hidden bottleneck** for dual-path encoding | HIGH | 02, 03, 04, 09 |
| 2 | **Audio passthrough more constrained than video** | HIGH | 06, 07, 12 |
| 3 | **H.264 strategically optimal** despite technical inferiority | HIGH | 01, 02, 08 |
| 4 | **Recording storage should mirror game save systems** | HIGH | 03, 04, 05 |
| 5 | **Go goroutines map perfectly to video pipeline stages** | HIGH | 03, 05, 11 |
| 6 | **Display pipeline is largest unaddressed latency source** (30-100ms) | HIGH | 01, 03, 07, 08 |
| 7 | **SQP + custom UDP = next-gen sub-10ms transport** | MEDIUM | 01, 08, 12 |
| 8 | **VVC hardware gap makes AV1 correct near-term investment** | HIGH | 01, 02, 08 |
| 9 | **GPU vendor selection should be topology-driven** | HIGH | 02, 09 |
| 10 | **Recording feature differentiates from all competitors** | HIGH | 04, 05, 10 |

## 9. CONFLICT ZONES & RESOLUTIONS

From cross-verification analysis (6 conflicts identified, all resolved):

| Conflict | Resolution | Status |
|----------|-----------|--------|
| Intel non-standard B-frames | Accepted trade-off — best latency, requires decoder validation | RESOLVED |
| AV1 latency penalty varies (1-3 frames) | Vendor-dependent — NVENC largest, AMD none | RESOLVED |
| LL vs ULL tuning value | ULL provides dramatically more benefit than LL | PARTIALLY RESOLVED |
| NVENC session limit exact numbers | Temporal evolution — 2→3→5→8 over time | TEMPORALLY RESOLVED |
| Software encoding viability for 4K60 | Possible on 16+ core CPUs but 5-10x worse latency | CONDITIONALLY RESOLVED |
| Recording impact on streaming | Minimal IF thermals managed | CONDITIONALLY RESOLVED |

**No genuine irreconcilable contradictions were found.**

## 10. STATISTICS & CONFIDENCE SUMMARY

| Tier | Count | Percentage |
|------|-------|------------|
| High Confidence (HC) | 15 | 57% |
| Medium Confidence (MC) | 6 | 23% |
| Low Confidence (LC) | 3 | 11% |
| Conflict Zones (CZ) | 6 | 23% |

- Total findings cross-verified: 26
- Total research dimensions: 12
- Total sources consulted: 300+
- Insights derived: 10 (all actionable)

---

*Synthesis compiled from all 14 research artifacts. All citations reference the original dimension files for traceability. Key claims include inline source references from the original research.*
