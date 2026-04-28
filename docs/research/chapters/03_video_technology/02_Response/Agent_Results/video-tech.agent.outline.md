# Comprehensive Video Technology Research for CloudStream Gaming Platform

## Executive Summary
### Key Findings
#### Hardware GPU encoders achieve 83-117ms E2E latency at 4K60, 10-20x faster than software encoding, with Intel ULL leading at 83ms (5 frames) and NVENC providing most consistent ~117ms (7 frames)
#### Zero-copy GPU pipelines reduce latency from 200-500ms to 10-30ms by eliminating CPU-GPU memory roundtrips via DXGI shared textures (Windows), IOSurface (macOS), and DMA-BUF (Linux)
#### Simultaneous streaming + recording is viable via dual NVENC sessions with "local-first + background sync" storage pattern, but thermal management is the hidden bottleneck (83°C throttling reduces throughput 25-30%)
#### 3-tier codec strategy: H.264 (universal fallback, 98.2% coverage), HEVC (standard tier, best compression), AV1 (premium tier, 40-55% bandwidth savings) — skip VVC for real-time until 2028+
#### Multi-channel audio passthrough has binary capability threshold (full surround vs stereo collapse), making endpoint chain validation more critical than video codec negotiation
#### SQP congestion control outperforms WebRTC GCC by 2-3x under TCP competition; combined with custom UDP transport, sub-10ms LAN streaming is achievable

## 1. Video Codec Architecture for Real-Time Gaming (~5000 words, 6 tables, 2 diagrams)
### 1.1 Codec Landscape Overview
#### 1.1.1 Six codecs evaluated: H.264/AVC, HEVC/H.265, AV1, VP9, VVC/H.266, and emerging JPEG XS — each mapped to latency, compression efficiency, and hardware decode coverage
#### 1.1.2 Hardware decode coverage comparison table: H.264 at 98.2%, HEVC at 85%, AV1 at ~25%, VVC at <1% (no browser support) — coverage determines fallback chain design
#### 1.1.3 WebRTC codec mandate (RFC 7742): H.264 is the only universally guaranteed codec; AV1 optional (rtp-av1-25), HEVC negotiated via rtcp-fb
### 1.2 Encoder Latency Benchmarks
#### 1.2.1 IEEE 2025 peer-reviewed benchmark results at 4K60: Intel ULL = 83ms (5 frames), NVENC = 117ms (7 frames), AMD = 100-150ms (6-9 frames), software (x264/x265/SVT-AV1) = 683-1500ms (41-90+ frames)
#### 1.2.2 Latency breakdown per pipeline stage: capture (1-3ms), encode (5-117ms), packetize (0.5-2ms), transmit (variable), decode (5-20ms), display (30-100ms) — encode is the controllable bottleneck
#### 1.2.3 Frame structure impact: intra-only vs inter-frame, B-frame trade-offs, GOP size selection for gaming (1-2 frames optimal), and why bidirectional B-frames add unacceptable latency
#### 1.2.4 Intel non-standard B-frame compatibility risk: unidirectional B-frames in ULL mode may affect decoder compatibility despite `-bf 0` flag (CZ-1 resolution)
### 1.3 Codec Selection Strategy
#### 1.3.1 Three-tier codec negotiation design: Tier 1 H.264 (universal fallback), Tier 2 HEVC (standard efficiency), Tier 3 AV1 (premium efficiency), with WebRTC SDP offer/answer implementation
#### 1.3.2 AV1 latency penalty is vendor-dependent: NVENC adds 2-3 frames, Intel ULL adds 1 frame, AMD RDNA4 shows no penalty — implementation must measure per-GPU
#### 1.3.3 VVC/H.266 analysis: 50% bitrate reduction over HEVC but 8-10x encoding complexity, zero browser support, no consumer hardware encode — recommendation: skip for real-time, monitor for passive recording only
#### 1.3.4 Emerging codecs: JPEG XS TDC profile (20:1 compression, 1ms encode, limited hardware); custom GPU compute codecs (PyroWave 0.13ms at 200+ Mbps) — position as R&D tracks, not production
### 1.4 Go Implementation: Codec Pipeline
#### 1.4.1 FFmpeg hardware codec initialization in Go: Cgo vs CLI wrapping trade-off, `ffmpeg-go` library usage, encoder context lifecycle management
#### 1.4.2 Codec negotiation via Pion WebRTC: creating `RTCRtpCodecParameters` for H.264/HEVC/AV1, SDP offer construction with codec preference ordering, `SetCodecPreferences` API
#### 1.4.3 Encoder configuration: Go structs for `CodecConfig` (codec, preset, bitrate, gop, profile, tier), JSON/YAML serialization, hot-reload support
#### 1.4.4 Codec capability registry: GPU vendor → supported codecs → profiles → levels → max resolution/framerate lookup table with automatic population at startup

## 2. Hardware GPU Acceleration & Encoder Technology (~4500 words, 5 tables, 1 diagram)
### 2.1 NVIDIA NVENC Deep-Dive
#### 2.1.1 Architecture: dedicated NVDEC/NVENC ASICs on GPU die, independent of CUDA cores, 9th generation on RTX 50 series with 4:2:2 10-bit HEVC/AV1 support
#### 2.1.2 RTX 50 improvements: ~5% quality gain, 60% faster encoding than RTX 4090, dual NVENC on 4070 Ti+ models, 5-8 concurrent sessions with recent drivers
#### 2.1.3 Latency consistency: NVENC maintains ~7 frames across all presets P1-P7, making it the most predictable encoder — critical for SLA guarantees
#### 2.1.4 Consumer vs professional: GeForce cards session-limited (2-8) vs RTX A-series/Quadro unlimited — deployment implications for multi-stream hosts
### 2.2 Intel QuickSync Video (QSV)
#### 2.2.1 Architecture: dedicated Media Fixed Function hardware on CPU die (integrated) and Arc GPUs (discrete), ULL mode achieving 83ms (5 frames) — lowest of all vendors
#### 2.2.2 ULL mode trade-offs: non-standard B-frame structure (CZ-1), encoder selection at stream start improves reliability (Sunshine lesson)
#### 2.2.3 No session limits: Intel QSV does not impose concurrent encode session caps — ideal for multi-session cloud gaming hosts
#### 2.2.4 Ultra Low-Latency tuning provides significant improvement over Low-Latency (10-12 frames → 8 frames for H.264), contrary to LL-tune-neglect claims (CZ-3 resolution)
### 2.3 AMD AMF / RDNA4 Media Engine
#### 2.3.1 RDNA4 dual media engines: 25% H.264 low-latency quality improvement, 11% HEVC improvement, AV1 B-frame support, 50% AV1 decode uplift
#### 2.3.2 No session limits: AMD does not artificially cap concurrent encode sessions — cost-effective multi-session deployment
#### 2.3.3 AMF SDK integration: `AMFVideoEncoderVCE_AVC`, `AMFVideoEncoder_HEVC`, `AMFVideoEncoder_AV1` component selection via Go CGO bindings
#### 2.3.4 VideoToolbox (Apple): M1-M4 Media Engine capabilities, M3+ AV1 encode support, ProRes hardware, platform-specific Go implementation via CGO
### 2.4 Encoder Capability Detection
#### 2.4.1 NVIDIA: `go-nvml` library for driver version, VRAM, temperature, encoder utilization, session count — NVML bindings for Go
#### 2.4.2 Intel/AMD: `gopsutil` + `sysfs` on Linux, DXGI adapter queries on Windows, IOKit on macOS for GPU enumeration and capability detection
#### 2.4.3 FFmpeg encoder enumeration: runtime `-encoders` query filtered by `h264_nvenc`, `hevc_qsv`, `av1_amf` — building supported codec list dynamically
#### 2.4.4 Encoder capability registry: Go struct `EncoderProfile` with Vendor, Codec, MaxResolution, MaxFPS, MaxSessions, Supports10Bit, SupportsHDR fields — populated at host startup
### 2.5 Thermal Management & Dynamic Quality
#### 2.5.1 Thermal throttling at 83°C reduces encoder throughput 25-30%: detection via NVML temperature polling, thermal margin calculation
#### 2.5.2 Dynamic quality adjustment ladder: preset escalation P7→P1, resolution 4K→1080p→720p, bitrate reduction 25-50%, codec fallback HEVC→H.264
#### 2.5.3 Proactive thermal management: reduce quality BEFORE throttling occurs, thermal prediction model based on GPU load trajectory
#### 2.5.4 Session routing: route recording-intensive sessions to thermally advantaged hosts, liquid cooling recommendations for recording-enabled deployments

## 3. Zero-Copy Capture Pipeline Architecture (~4500 words, 4 tables, 3 diagrams)
### 3.1 Windows: DXGI Desktop Duplication
#### 3.1.1 `IDXGIOutputDuplication` API: acquires shared GPU texture handles (`DXGI_OUTDUPL_FRAME_INFO`), returns dirty rectangles for incremental updates
#### 3.1.2 Zero-copy path: DXGI texture → `NV_ENC_REGISTER_RESOURCE` with `NV_ENC_INPUT_RESOURCE_TYPE_DIRECTX` → NVENC encode surface without CPU roundtrip
#### 3.1.3 DWM interaction: Desktop Window Manager adds 1-3 frame latency; `DWMWA_CLOAK` and fullscreen exclusive mode reduce compositor overhead
#### 3.1.4 Windows.Graphics.Capture (Win10 1903+): modern API with better HDR support, returned as `IDirect3D11Texture2D` — preferred for new implementations
#### 3.1.5 Go integration: `syscall` + `unsafe` for COM interface calls, `github.com/go-ole/go-ole` for DirectX interop, dedicated goroutine for acquire-frame loop
### 3.2 macOS: ScreenCaptureKit
#### 3.2.1 `SCStream` API: captures to `IOSurface` or `CVPixelBuffer`, supports HDR content (`kCGColorSpaceExtendedLinearDisplayP3`), low-latency stream configuration
#### 3.2.2 Zero-copy path: `IOSurface` → `CVPixelBuffer` → VideoToolbox `VTCompressionSession` with `kCVPixelBufferIOSurfacePropertiesKey`
#### 3.2.3 macOS compositor: Core Animation adds ~1 frame latency; `CGDisplayStream` deprecated in favor of ScreenCaptureKit (macOS 12.3+)
#### 3.2.4 Go integration: CGO bridge to Objective-C ScreenCaptureKit, `github.com/progrium/macdriver` or custom CGO bindings
### 3.3 Linux: PipeWire & DMA-BUF
#### 3.3.1 PipeWire screen capture: `pw_stream` with `SPA_PARAM_FORMAT_mediaType` video, DMA-BUF fd export for zero-copy
#### 3.3.2 DMA-BUF → EGLImage → VAAPI: `eglCreateImageKHR` with `EGL_LINUX_DMA_BUF_EXT` → `vaCreateSurfaces` with `VASurfaceAttribExternalBuffers`
#### 3.3.3 Wayland vs X11: `xdg-desktop-portal-wlr` (wlroots) for Wayland, `XShm` fallback for X11, X11 capture adds 1-2 frames vs Wayland
#### 3.3.4 Go integration: D-Bus bindings for xdg-desktop-portal, `github.com/rajveermalviya/go-wayland` or direct `syscall` for PipeWire
### 3.4 Frame Pacing & Synchronization
#### 3.4.1 V-Sync bypass strategies: NVIDIA Fast Sync, AMD Enhanced Sync, tearing-aware pacing — reducing capture-to-display pipeline by 1-2 frames
#### 3.4.2 Frame pacing algorithm: token bucket rate limiter at target FPS (16.67ms @ 60Hz), paced sender with RTCP feedback, adaptive pacing based on network conditions
#### 3.4.3 Frame deduplication: duplicate frame detection for static content, skip identical frames to save bandwidth, resume on scene change
#### 3.4.4 Timestamp management: `QueryPerformanceCounter` (Windows), `CMClock` (macOS), `CLOCK_MONOTONIC` (Linux) — nanosecond-precision frame timing for A/V sync
### 3.5 Go Pipeline Architecture
#### 3.5.1 Goroutine stage design: capture → preprocess (scale/color) → encode → packetize → transmit, each stage as independent goroutine with channel communication
#### 3.5.2 `sync.Pool` for frame buffers: `[]byte` pool with 33MB entries (4K NV12), eliminates GC pressure, benchmark target: 200M+ writes/sec via ring buffer channels
#### 3.5.3 Backpressure handling: buffered channels with capacity 1-3 frames, drop policy for stale frames (keep newest), non-blocking sends with select
#### 3.5.4 Memory layout: NV12 preferred over RGB for encode input (50% size reduction), GPU-native format avoids CPU conversion

## 4. Simultaneous Streaming + Recording Architecture (~4000 words, 4 tables, 2 diagrams)
### 4.1 Dual-Path Encoding Design
#### 4.1.1 Architecture: single capture source → fork to StreamEncoder + RecordEncoder, both run simultaneously on separate NVENC sessions or codec instances
#### 4.1.2 GPU session allocation: Stream = primary codec (HEVC/AV1), Record = secondary codec (H.264 for compatibility), session budgeting per GPU vendor limits
#### 4.1.3 FFmpeg tee muxer approach: `-f tee "[f=flv]rtmp://...|[f=matroska]/path/out.mkv"` — single encode, multiple outputs, but limits codec/format independence
#### 4.1.4 Independent encode approach: separate `EncoderConfig` for stream (low-latency, high compression) vs record (compatibility, editing-friendly) — preferred for quality
### 4.2 Frame Duplication Strategies
#### 4.2.1 GPU memory fork: capture texture duplicated to two encoder surfaces via `ID3D11DeviceContext::CopyResource`, ~0.1ms copy time on GPU
#### 4.2.2 Shared texture with dual consumers: both encoders read from same GPU texture (read-only), zero copy overhead, requires encoder API support
#### 4.2.3 CPU-side duplication fallback: when GPU memory fork unavailable, `sync.Pool`-buffered CPU copy at ~1-3ms per 4K frame — acceptable for recording path
#### 4.2.4 Audio duplication: separate goroutine captures loopback audio, duplicates to stream mixer + record muxer via channel fan-out
### 4.3 Recording Container Formats
#### 4.3.1 MKV (Matroska): progressive writing, crash-safe (playable up to crash point), supports all codecs and multi-channel audio, `libmatroska` via FFmpeg
#### 4.3.2 Fragmented MP4 (fMP4): `movflags=frag_keyframe+empty_moov+default_base_moof`, live-safe with playable fragments, better for network streaming
#### 4.3.3 Container selection logic: local recording → MKV (crash-safe); network upload → fMP4 (streamable); editing workflow → MOV (ProRes if available)
#### 4.3.4 Audio track embedding: AAC/Opus for stream-safe, PCM/AC3 passthrough for archive quality, multi-track audio support
### 4.4 Real-Time DVR & Instant Replay
#### 4.4.1 Circular buffer design: ring buffer of encoded frames in memory (30-min capacity at 1080p30 ~9GB), overwrite oldest on full, instant clip extraction
#### 4.4.2 Clip extraction: user-defined start/end timestamps → demux from circular buffer → remux to output container without re-encode
#### 4.4.3 Background finalize: async moov atom writing for MP4, header finalization for MKV, concurrent with ongoing recording
#### 4.4.4 Go implementation: `container/ring` or custom ring buffer with `sync.RWMutex`, clip server HTTP endpoint for playback while recording continues

## 5. Storage Backend & Pipeline Design (~4000 words, 5 tables, 1 diagram)
### 5.1 Storage Backend Implementations
#### 5.1.1 SMB/CIFS 3.1.1: `github.com/hirochachacha/go-smb2` library, multi-channel support, encryption, Go `io.Writer` interface adapter, connection pooling
#### 5.1.2 NFS v4.2: `github.com/vmware/go-nfs-client` or `github.com/willscott/go-nfs`, pNFS for parallel access, file delegation for write caching
#### 5.1.3 FTP/FTPS/SFTP: `github.com/jlaffaye/ftp` for FTP, `github.com/pkg/sftp` for SFTP, Go `io.WriteCloser` wrapper, TLS upgrade for FTPS
#### 5.1.4 WebDAV: `github.com/studio-b12/gowebdav`, HTTP-based, firewall-friendly, digest authentication support
#### 5.1.5 S3-compatible: AWS SDK v2 `github.com/aws/aws-sdk-go-v2/service/s3`, multipart upload, MinIO/Backblaze B2 compatibility
#### 5.1.6 Custom pipeline: HTTP POST/PUT endpoint with configurable headers, auth, retry logic — webhook-style delivery
### 5.2 Write Strategy & Resilience
#### 5.2.1 "Local-first + background sync" pattern: always record to local NVMe SSD first, background uploader copies to configured backends — the game-save model (Insight 4)
#### 5.2.2 Async write pipeline: Go channel-based producer-consumer, buffered writer with 8MB flush threshold, non-blocking backend writes
#### 5.2.3 Network interruption handling: exponential backoff retry (100ms → 1s → 10s), resume from last successful offset, automatic reconnection
#### 5.2.4 Write resilience: `io.Writer` wrapper with checksum validation, partial file recovery, transaction log for incomplete uploads
### 5.3 Bandwidth & Performance
#### 5.3.1 Bandwidth requirements table: 720p30 (3-5 Mbps), 1080p60 (10-15 Mbps), 4K60 H.264 (35-50 Mbps), 4K60 HEVC (20-30 Mbps), 4K120 (65+ Mbps), plus audio overhead (1-2 Mbps for 5.1)
#### 5.3.2 Storage performance targets: local NVMe write >3GB/s (6x headroom), 1GbE network storage ~110 MiB/s (17x headroom for 4K60 HEVC)
#### 5.3.3 Bandwidth throttling: configurable upload rate limit to prevent saturating network, token bucket rate limiter per backend
#### 5.3.4 Multi-backend fan-out: write to multiple backends simultaneously (e.g., local + SMB + S3), independent failure domains
### 5.4 Go Implementation: Storage Pipeline
#### 5.4.1 `StorageBackend` interface: `Write(p []byte) (n int, err error)`, `Close() error`, `HealthCheck() error` — uniform interface for all backends
#### 5.4.2 Backend factory: `NewStorageBackend(config BackendConfig) (StorageBackend, error)` — type dispatch based on config.Protocol ("smb", "nfs", "ftp", "webdav", "s3", "custom")
#### 5.4.3 Upload manager: goroutine pool for parallel uploads, `sync.WaitGroup` for multi-backend coordination, progress tracking with callback
#### 5.4.4 Encryption at rest: AES-256-GCM per-file encryption with key derived from session ID, transparent encrypt/decrypt via `cipher.StreamWriter`

## 6. Audio Pipeline Technology (~4000 words, 5 tables, 2 diagrams)
### 6.1 Audio Codecs & Formats
#### 6.1.1 Opus: IETF standard (RFC 6716), 5ms frame sizes, <20ms end-to-end latency, 96-128 kbps stereo, 192-256 kbps 5.1, 256-450 kbps 7.1 — primary real-time codec
#### 6.1.2 AAC-LC/HE-AAC: ubiquitous device support, 64-128 kbps stereo, ADTS framing for RTP, fallback when Opus unavailable
#### 6.1.3 AC3 (Dolby Digital): 5.1 at 448 kbps, SPDIF passthrough compatible, widely supported by AV receivers
#### 6.1.4 E-AC3 (Dolby Digital Plus): up to 15.1 channels, Atmos carrier (DD+ with Dolby MAT 2.0), HDMI ARC compatible
#### 6.1.5 Lossless formats: PCM (uncompressed reference), Dolby TrueHD (MLP-based, Atmos in TrueHD), DTS-HD MA — for recording/archive only
#### 6.1.6 Dolby Atmos & DTS:X: object-based 3D audio, 7.1.4+ channel layouts, eARC required for uncompressed Atmos transport
### 6.2 Multi-Channel Audio Architecture
#### 6.2.1 Channel layout matrix: Mono, Stereo, 2.1, 5.1, 7.1, 7.1.4 (Atmos height), 22.2 — with SMPTE channel ordering standard
#### 6.2.2 The passthrough binary threshold (Insight 2): audio either works in full surround or collapses to stereo — NO graceful degradation like video resolution scaling
#### 6.2.3 Hardware endpoint chain: Game Audio API → OS Audio Stack → Capture → Encode → Transmit → Decode → AV Receiver — ANY link can force stereo fallback
#### 6.2.4 Windows channel order quirk: Windows 7.1 uses different channel ordering than Dolby/DTS/SMPTE standard — requires remapping for passthrough
### 6.3 Audio Capture & Hardware Interfaces
#### 6.3.1 Windows audio capture: WASAPI shared mode (lower latency) vs exclusive mode (bit-accurate), loopback capture via `AUDCLNT_STREAMFLAGS_LOOPBACK`, endpoint detection via `IMMDeviceEnumerator`
#### 6.3.2 macOS audio: Core Audio HAL, `AudioObjectPropertySelector` for device enumeration, `BlackHole`/`Loopback` for virtual loopback capture
#### 6.3.3 Linux audio: PulseAudio `module-loopback` or PipeWire `loopback` node, ALSA `snd-aloop` kernel module, JACK for professional low-latency
#### 6.3.4 Hardware interfaces: HDMI eARC (uncompressed Atmos + 4K video), SPDIF/TOSLINK (compressed 5.1 only), DisplayPort audio, USB audio class 2.0
### 6.4 Go Audio Implementation
#### 6.4.1 Opus encoding in Go: `github.com/pion/opus` or `gopkg.in/hraban/opus.v2` bindings, MultiStream API for >2 channels, encoder configuration (bitrate, complexity, signal type)
#### 6.4.2 Audio pipeline: capture goroutine → channel split (stream + record) → Opus encoder (stream) → PCM/AAC muxer (record), with `sync.Pool` for audio buffers
#### 6.4.3 Hardware capability detection: enumerate audio endpoints, query channel count, sample rate, bit depth, format support via platform APIs wrapped in Go interfaces
#### 6.4.4 A/V synchronization: shared timestamp reference (same `CLOCK_MONOTONIC` source), audio drift correction via sample rate adjustment, ITU-R BS.1359 threshold (45ms audio lead max)

## 7. HDR & Color Space Management (~3000 words, 3 tables, 1 diagram)
### 7.1 HDR Format Comparison
#### 7.1.1 HDR10 (static metadata): base layer, free, GPU encode supported (HEVC Main 10, AV1 10-bit), limited scene-by-scene optimization
#### 7.1.2 HDR10+ (dynamic metadata): royalty-free, live encoder supported via SEI messages, backward-compatible HDR10 fallback — recommended for cloud gaming (HC-8)
#### 7.1.3 Dolby Vision: proprietary, $2.5K/year licensing, profile 5/8/9, NO consumer GPU real-time encoder support — avoid for open-source platform
#### 7.1.4 HLG: inherent SDR backward compatibility, no metadata needed, broadcast-friendly, limited gaming content support
### 7.2 HDR Pipeline Implementation
#### 7.2.1 HDR capture: Windows DXGI `R16G16B16A16_FLOAT` + scRGB, macOS `MTLPixelFormat.rgba16Float` + `extendedLinearDisplayP3`, Linux Vulkan HDR DMA-BUF with color metadata
#### 7.2.2 Tone mapping: host-side (before encode, fixed pipeline) vs client-side (after decode, display-aware); algorithms: Hable, ACES, BT.2390 via libplacebo
#### 7.2.3 HDR over WebRTC: HEVC Main 10 + SEI for HDR10/HDR10+ metadata, AV1 10-bit profile with metadata OBU, no native HDR signaling in WebRTC spec
#### 7.2.4 SDR fallback: automatic tone mapping for non-HDR clients, preserving color accuracy via perceptual color space conversion

## 8. Adaptive Bitrate & Congestion Control (~3500 words, 3 tables, 2 diagrams)
### 8.1 Gaming-Specific ABR
#### 8.1.1 Why VoD ABR fails for gaming: segment-based (2-10s) adds unacceptable latency, pre-encoded ladders inflexible, 10-30s buffer vs gaming's 1-3 frame buffer
#### 8.1.2 Frame-level adaptation: encoder reconfigured per-frame (not per-segment), compression params adjusted first, frame rate second, resolution last resort
#### 8.1.3 3-tier quality ladder: 4K Ultra (35-50 Mbps), 1080p High (15-25 Mbps), 720p Standard (10-15 Mbps) — with dynamic tier selection based on bandwidth
#### 8.1.4 AFR (Adaptive Frame Rate): reduce to 30fps when bandwidth constrained (halves bitrate), 7.4x tail queuing delay reduction vs fixed 60fps
### 8.2 Congestion Control Algorithms
#### 8.2.1 Google GCC default: delay-based, underperforms vs TCP Cubic (96% bitrate decrease under competition) — NOT recommended for gaming
#### 8.2.2 SQP (Scalable Quality Protocol): 2-3x higher bandwidth than GCC under TCP competition, Google Research + ACM COMSNETS 2025 verified (HC-11)
#### 8.2.3 BBR v2: model-based, 21% bitrate decrease under TCP competition, better than GCC but worse than SQP
#### 8.2.4 Pudica (Tencent START): 3.1x avg frame delay reduction, 5.7x on WiFi, proprietary — research insight only
#### 8.2.5 Camel (250M users): +70.8% 1080P resolution ratio via deep reinforcement learning — benchmark reference
### 8.3 Network Resilience
#### 8.3.1 FEC strategy by RTT: RTT<30ms NACK-only, RTT 30-80ms light FEC (10-20%) + RTX, RTT>80ms heavy FEC (20-25%) primary
#### 8.3.2 Reed-Solomon implementation: `klauspost/reedsolomon` Go library, >15GB/s per core with SIMD (AVX2/NEON), constant 100% recovery vs XOR FEC
#### 8.3.3 Jitter buffer: 1-3 frame adaptive buffer (16.7-50ms at 60fps), underflow handling (repeat frame), overflow handling (drop oldest)
#### 8.3.4 NACK-based repair: `nack` module in Pion WebRTC, configurable max retransmissions, RTX SSRC for retransmission packets

## 9. Hardware Detection & Dynamic Optimization (~3000 words, 3 tables, 1 diagram)
### 9.1 GPU Capability Detection
#### 9.1.1 NVIDIA detection: `github.com/NVIDIA/go-nvml` — GPU name, VRAM, driver version, temperature, utilization, encoder count, power draw
#### 9.1.2 AMD detection: `go-rocm-smi` or sysfs (`/sys/class/drm/card*/device/`) for GPU info, `amdgpu` driver exposes hwmon for temperature
#### 9.1.3 Intel detection: `intel_gpu_top` or sysfs, Arc GPUs via `intel-gpu-tools`, integrated via `/sys/class/drm/card*/`
#### 9.1.4 Apple detection: `IOKit` framework via CGO, `Metal` device enumeration, M-series chip detection via `sysctl hw.model`
### 9.2 System Capability Profiling
#### 9.2.1 CPU detection: `gopsutil/cpu` for cores/threads/clock speed, `golang.org/x/sys/cpu` for feature flags (AVX2, AVX-512)
#### 9.2.2 Memory detection: available RAM, swap status — minimum 8GB for 1080p streaming, 16GB for 4K recording
#### 9.2.3 Storage detection: disk speed benchmark (fio-style), available space check — NVMe required for 4K recording, SATA SSD minimum for 1080p
#### 9.2.4 Network detection: interface speed (1GbE minimum for 4K), WiFi vs Ethernet detection, latency probe to nearest router
### 9.3 Dynamic Quality Controller
#### 9.3.1 Quality controller state machine: Normal → Degraded (thermal warning) → Minimal (throttling active) → Recovery (cooling detected)
#### 9.3.2 Preset escalation: GPU temperature 70°C+ → reduce encode preset one step; 80°C+ → reduce resolution; 83°C+ → emergency H.264 fallback
#### 9.3.3 Adaptive bitrate: bandwidth estimator (SQP-based) drives bitrate target, frame drop rate drives resolution target, thermal drives preset target
#### 9.3.4 Go implementation: `QualityController` struct with goroutine for monitoring loop, channel-based config updates, thread-safe state transitions via `sync.RWMutex`

## 10. Testing & Validation Framework (~4000 words, 4 tables, 2 diagrams)
### 10.1 Frame Integrity Testing
#### 10.1.1 Frame counter injection: incrementing counter embedded in each frame during capture, validated at each pipeline stage (capture → encode → packetize → decode → render)
#### 10.1.2 Frame loss detection: sequence gap detection at receiver, `frames_presented` vs `frames_captured` vs `frames_encoded` vs `frames_decoded` metrics
#### 10.1.3 Recording validation: MD5 checksum per frame, container integrity verification (`ffmpeg -v error -i file.mkv -f null -`), frame count matching
#### 10.1.4 Zero-frame-loss guarantee: recording path validated by frame counter continuity check, no gaps allowed (vs streaming which tolerates <0.1% loss)
### 10.2 Latency Measurement
#### 10.2.1 Glass-to-glass measurement: LED + photodiode method (0.5ms precision), timecode source + photographic comparison (IEEE 2025 methodology)
#### 10.2.2 Stage-by-stage profiling: capture timestamp, encode start/end, packetize complete, network send, network receive, decode complete, display present
#### 10.2.3 Automated latency probes: synthetic frame with known timestamp, measure round-trip time through full pipeline, statistical aggregation (p50, p95, p99)
#### 10.2.4 PresentMon integration: Intel PresentMon for Windows frame timing, ETW tracing for pipeline stage breakdown
### 10.3 Video Quality Testing
#### 10.3.1 Objective metrics: VMAF (perceptual gold standard), SSIM (structural), PSNR (pixel-level) — comparison table with use cases
#### 10.3.2 Artifact detection: blocking (macroblock boundaries), banding (color gradients), ringing (edge halos) — automated via OpenCV analysis
#### 10.3.3 Codec conformance: H.264/AVC Annex B compliance, HEVC VPS/SPS/PPS validation, AV1 OBU structure verification — FFmpeg `ffprobe` + `h264bitstream`
#### 10.3.4 Long-term stability: 8+ hour sustained encoding test, memory leak detection (`go test -race`), thermal stress validation
### 10.4 Automated Test Framework
#### 10.4.1 Unit test hierarchy: encoder core 90%+, muxer 85%+, config 95%+, storage backend 80%+, pipeline integration 75%+
#### 10.4.2 Go testing patterns: table-driven tests with `t.Run`, `testify/assert` for validation, `testify/require` for fatal conditions, benchmark with `b.ReportAllocs()`
#### 10.4.3 Network resilience testing: Linux `tc/netem` for packet loss (1-10%), jitter (±50ms), bandwidth throttling (10-100 Mbps), reordering
#### 10.4.4 CI/CD integration: GitHub Actions multi-platform matrix (Windows, macOS, Linux), FFmpeg version matrix, GPU type matrix, benchmark regression detection
#### 10.4.5 100% coverage strategy: unit → integration → e2e → load → soak testing hierarchy, `go test -coverprofile`, target 90%+ overall with 100% on critical paths

## 11. Network Transport & Packet Optimization (~3000 words, 3 tables, 1 diagram)
### 11.1 WebRTC Transport Internals
#### 11.1.1 RTP packetization: H.264 NAL unit fragmentation (STAP-A, FU-A), HEVC VPS/SPS/PPS prefix in each keyframe, AV1 OBU encapsulation
#### 11.1.2 Pion WebRTC configuration: `SettingEngine` for custom MTU (1200 bytes), `SetSRTPProtectionProfile`, single-port multiplexing, DataChannels for controller input
#### 11.1.3 RTCP feedback: Receiver Reports (RR), Transport Wide CC (TWCC), Picture Loss Indication (PLI), Full Intra Request (FIR) — enable for adaptive quality
#### 11.1.4 ICE optimization: host preference, STUN binding timeout tuning, TURN relay selection by RTT (coturn/eturnal), ~20-30% sessions require relay
### 11.2 Custom UDP Transport (LAN Optimization)
#### 11.2.1 Parsec BUD-style protocol: raw UDP with application-layer reliability, 7ms LAN latency (vs WebRTC 15-20ms), DTLS 1.2 per-packet encryption (~0.5ms overhead)
#### 11.2.2 Dual transport strategy: custom UDP for native clients on LAN, WebRTC for browser/WAN — best of both worlds (Insight 7)
#### 11.2.3 Packet structure: 16-byte header (sequence, timestamp, flags, payload type) + encrypted payload, ~0.5ms parse time
#### 11.2.4 SQP integration: frame-coupled paced packet trains, bandwidth estimation via inter-arrival time, TCP-friendly rate control
### 11.3 Packet Optimization
#### 11.3.1 MTU selection: 1200 bytes default (VPN-safe), 1361 bytes for Cloudflare One compatibility, path MTU discovery via probe packets
#### 11.3.2 Packet pacing: token bucket at target bitrate, frame-coupled sending (distribute packet sends across frame interval), prevent bursty transmission
#### 11.3.3 DSCP marking: `IP_TOS` socket option with DSCP 46 (EF) for gaming traffic, effective on home networks, typically stripped by ISPs
#### 11.3.4 Multi-path: MPQUIC evaluation for WiFi+cellular bonding, packet scheduling across paths by RTT

## 12. Go Implementation Architecture & Integration (~4000 words, 3 tables, 2 diagrams)
### 12.1 Package Structure
#### 12.1.1 `internal/streaming` package: encoder, capture, pipeline, muxer sub-packages — integrates with protocol, controller, session service from previous stages
#### 12.1.2 `internal/recording` package: record controller, storage backends, upload manager — depends on streaming capture output
#### 12.1.3 `internal/audio` package: capture, encode, multi-channel, passthrough — separate goroutine pipeline from video
#### 12.1.4 `internal/hardware` package: GPU detection, thermal monitoring, capability registry — shared service for all packages
### 12.2 Integration with Existing Architecture
#### 12.2.1 Protocol package integration: streaming registers RTP payload types with protocol layer, session service negotiates codec via SDP
#### 12.2.2 Controller input integration: DataChannels for controller input run parallel to video stream, separate goroutine, no head-of-line blocking
#### 12.2.3 Session service integration: host agent advertises encoder capabilities on startup, session service routes based on codec + thermal + GPU availability
#### 12.2.4 Catalog integration: game metadata includes recommended codec, resolution, HDR flag; pre-validated against host capabilities
### 12.3 Memory Management
#### 12.3.1 `sync.Pool` strategy: 33MB frame buffers (4K NV12) pooled, 4KB audio buffers pooled, encoder output buffers pooled — eliminates GC pressure on hot path
#### 12.3.2 Memory budget: 4K streaming ~200MB (GPU), 4K recording +100MB, audio ~10MB, total ~310MB per session — 3 concurrent sessions per 16GB system feasible
#### 12.3.3 Mmap for large buffers: `syscall.Mmap` for circular buffer backing, `unsafe` for zero-copy access patterns
#### 12.3.4 Race safety: all shared state via channels (share by communicating), `sync.RWMutex` for capability registry, `atomic` for counters
### 12.4 Build & Deployment
#### 12.4.1 Cross-compilation: `CGO_ENABLED=1` with mingw-w64 (Windows), osxcross (macOS), native GCC (Linux); static linking where possible
#### 12.4.2 Build tags: `//go:build windows` for DXGI, `//go:build darwin` for ScreenCaptureKit, `//go:build linux` for PipeWire — platform-specific capture implementations
#### 12.4.3 FFmpeg dependency: static FFmpeg build bundled with binary, version pinning for reproducibility, `ffmpeg -version` check at startup
#### 12.4.4 Deployment: single binary with embedded static assets, Docker optional, systemd service template for Linux hosts

# References
## video-tech.agent.outline.md
- **Type**: Report outline
- **Description**: This outline file
- **Path**: /mnt/agents/output/video-tech.agent.outline.md

## Research Artifacts
- **Type**: Deep research dimension files
- **Description**: 12 dimension research files, cross-verification, and insight extraction
- **Path**: /mnt/agents/output/research/video-tech_dim01.md through video-tech_dim12.md, video-tech_cross_verification.md, video-tech_insight.md

## Requirements Analysis
- **Type**: Structured requirements
- **Description**: Explicit and implicit requirements extracted from user request
- **Path**: /mnt/agents/output/video-tech_requirements.md

## Artifact Synthesis
- **Type**: Research synthesis
- **Description**: Synthesis of all research artifacts with key findings and recommendations
- **Path**: /mnt/agents/output/video-tech_artifact_synthesis.md
