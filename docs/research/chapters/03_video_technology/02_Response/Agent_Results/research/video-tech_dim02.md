# Dimension 02: Hardware GPU Encoder Technology

## Executive Summary

This document provides a comprehensive analysis of hardware GPU video encoders across all major vendors: NVIDIA NVENC (8th/9th gen), Intel QuickSync/AV1, AMD AMF/RDNA4, Apple VideoToolbox, and Linux VAAPI. It covers encoder capabilities, presets, latency modes, session limits, runtime detection APIs, Go language integration patterns, and thermal throttling impact on encoder performance.

---

## 1. NVIDIA NVENC (9th Gen on RTX 50, 8th Gen on RTX 40)

### 1.1 Architecture and Capabilities by Generation

**9th Generation NVENC (Blackwell / RTX 50 Series)**

Claim: "The new ninth-generation NVENC encoder in Blackwell improves quality for AV1 and HEVC by 5% BD-BR PSNR, and adds support for 4:2:2 H.264 and HEVC encoding."[^113^]
Source: NVIDIA RTX Blackwell GPU Architecture Whitepaper
URL: https://images.nvidia.com/aem-dam/Solutions/geforce/blackwell/nvidia-rtx-blackwell-gpu-architecture.pdf
Date: 2025 (CES)
Excerpt: "The new ninth-generation NVENC encoder in Blackwell improves quality for AV1 and HEVC by 5% BD-BR PSNR, and adds support for 4:2:2 H.264 and HEVC encoding. There's also a new AV1 Ultra High Quality (UHQ) mode that takes additional time and provides an extra 5% improvement for the best quality possible."
Context: Official NVIDIA architecture documentation
Confidence: High

Claim: "The GeForce RTX 5090 GPU supports up to three encoders and two decoders, boosting export speeds by over 50% gen-over-gen, and an impressive 4x compared to the RTX 3090 GPU with a single encoder."[^113^]
Source: NVIDIA RTX Blackwell GPU Architecture Whitepaper
URL: https://images.nvidia.com/aem-dam/Solutions/geforce/blackwell/nvidia-rtx-blackwell-gpu-architecture.pdf
Date: 2025 (CES)
Excerpt: "The GeForce RTX 5090 GPU supports up to three encoders and two decoders, boosting export speeds by over 50% gen-over-gen"
Context: Marketing claims partially verified by independent testing (see below)
Confidence: Medium (NVIDIA marketing; independent verification mixed)

**NVENC Encoder Count by SKU (RTX 50 Series)**

| GPU | NVENC Encoders (9th Gen) | NVDEC Decoders (6th Gen) |
|-----|-------------------------|-------------------------|
| RTX 5050 / 5060 / 5060 Ti / 5070 | 1 | 1 |
| RTX 5070 Ti | 2 | 1 |
| RTX 5080 | 2 | 2 |
| RTX 5090 | 3 | 2 |

Source: Wikipedia / NVIDIA official specs[^44^]

### 1.2 AV1 UHQ Mode Quality Benchmarks

Claim: "Combining [9th gen improvements] with the new AV1 UHQ mode can yield up to 15% BD-BR PSNR improvements. The gains are even larger when using the VMAF metric from Netflix."[^113^]
Source: NVIDIA RTX Blackwell GPU Architecture Whitepaper
URL: https://images.nvidia.com/aem-dam/Solutions/geforce/blackwell/nvidia-rtx-blackwell-gpu-architecture.pdf
Date: 2025
Excerpt: "The chart below illustrates the generational improvements to the encoder for AV1, and how combining them with the new AV1 UHQ mode can yield up to 15% BD-BR PSNR improvements. The gains are even larger when using the VMAF metric from Netflix."
Context: BD-Rate savings over Ada: Natural Content +5% (AV1), +15% (AV1+UHQ); Gaming +4%/+10%; VMAF Natural +10%/+18%; VMAF Gaming +9%/+14%
Confidence: High (vendor-provided but detailed)

Claim: "With `-tune uhq`, the worst preset is better than the best without `-tune uhq`."[^195^]
Source: scottstuff.net - Benchmarking FFMPEG's H.265 Options
URL: https://scottstuff.net/posts/2025/03/17/benchmarking-ffmpeg-h265/
Date: 2025-03-17
Excerpt: "So, with `-tune uhq`, the worst preset is better than the best without `-tune uhq`."
Context: Independent benchmark showing `-tune uhq` dramatically improves file size efficiency at VMAF=95; p1 with uhq produced 6456kB vs p7 without at 8912kB
Confidence: High

### 1.3 Independent Verification of Performance Claims

Claim: "We were unable to reproduce the 4x (300%) performance improvement reported for the RTX 5090 compared to the 3090... we were able to verify that H.264 10-bit 4:2:0 and 10-bit 4:2:2 exports exceeded the 60% threshold."[^25^]
Source: Puget Systems Labs
URL: https://www.pugetsystems.com/labs/articles/verifying-nvidia-geforce-rtx-50-series-performance/
Date: 2025
Excerpt: "Ultimately, we were unable to reproduce the 4x (300%) performance improvement reported for the RTX 5090 compared to the 3090... H.264 10-bit 4:2:0 and 10-bit 4:2:2 exports exceeded the 60% threshold, with gains ranging from 75% to 109%"
Context: Independent testing using Premiere Pro and DaVinci Resolve; 5090 improvements vs 3090 ranged from 34-200% (not 300%); vs 4090 from 10-66% (not 60% consistently)
Confidence: High (independent testing)

### 1.4 Presets and Latency Modes

The NVENC SDK exposes **7 presets (P1-P7)** and **4 tuning info modes**:

| Preset | Speed | Quality | Description |
|--------|-------|---------|-------------|
| P1 | Fastest | Lowest | Maximum throughput |
| P2 | Faster | Lower | |
| P3 | Fast | Low | |
| P4 (default) | Medium | Medium | |
| P5 | Slow | Good | |
| P6 | Slower | Better | |
| P7 | Slowest | Best | Maximum quality |

**Tuning Info Modes:**[^59^]
- `hq` - High quality (default for latency-tolerant transcoding, archiving, OTT)
- `uhq` - Ultra high quality (Blackwell/RTX 50 only, for HEVC and AV1)
- `ll` - Low latency (for cloud gaming, streaming, video conferencing with CBR)
- `ull` - Ultra low latency (for strictly bandwidth-constrained channels)
- `lossless` - Lossless encoding

Source: NVENC Video Encoder API Programming Guide (SDK 13.0)[^59^]

Claim: "For each tuning info, seven presets from P1 (highest performance) to P7 (lowest performance) have been provided to control performance/quality trade off. Using these presets will automatically set all relevant encoding parameters."[^59^]
Source: NVENC Video Encoder API Programming Guide
URL: https://docs.nvidia.com/video-technologies/video-codec-sdk/13.0/nvenc-video-encoder-api-prog-guide/index.html
Date: 2025
Excerpt: "For each tuning info, seven presets from P1 (highest performance) to P7 (lowest performance) have been provided to control performance/quality trade off."
Context: 28 combinations total (7 presets x 4 tuning modes)
Confidence: High

**FFmpeg Integration:**[^194^]
```bash
# Example: HEVC NVENC encoding with P7 + UHQ
ffmpeg -i input.mp4 -c:v hevc_nvenc -preset p7 -tune uhq -rc vbr -cq:v 20 output.mp4

# Example: Low-latency streaming
ffmpeg -i input.mp4 -c:v h264_nvenc -preset p4 -tune ll -rc cbr -b:v 6M output.mp4

# Example: Ultra low-latency
ffmpeg -i input.mp4 -c:v h264_nvenc -preset p1 -tune ull -rc cbr output.mp4
```

### 1.5 Split Frame Encoding (Multi-NVENC)

Claim: "When Split frame encoding is enabled, each input frame is partitioned into horizontal strips which are encoded independently and simultaneously by separate NVENCs, usually resulting in increased encoding speed compared to single NVENC encoding."[^59^]
Source: NVENC Video Encoder API Programming Guide
URL: https://docs.nvidia.com/video-technologies/video-codec-sdk/13.0/nvenc-video-encoder-api-prog-guide/index.html
Date: 2025
Excerpt: "When Split frame encoding is enabled, each input frame is partitioned into horizontal strips which are encoded independently and simultaneously by separate NVENCs... Though the feature improves the encoding speed it degrades quality."
Context: Available only for HEVC and AV1; implicit trigger conditions: 2+ NVENCs, frame height >= 2112 (HEVC) or >= 2048 (AV1); explicit mode also available
Confidence: High

**Split Frame Encoding - Implicit Mode Triggers:**

| Tuning Info | P1 | P2 | P3 | P4 | P5 | P6 | P7 |
|-------------|----|----|----|----|----|----|----|
| High Quality | Yes | Yes | No | No | No | No | No |
| Low Latency | Yes | Yes | Yes | Yes | No | No | No |
| Ultra Low Latency | Yes | Yes | Yes | Yes | No | No | No |

Source: NVENC Video Encoder API Programming Guide (SDK 13.0)[^59^]

Claim: "DaVinci Resolve Studio 20 beta adds support for three-way split-frame encoding — a technique where an input frame is divided into three parts, each processed by a different NVENC encoder."[^117^]
Source: NVIDIA Blog
URL: https://blogs.nvidia.com/blog/studio-rtx-ai-garage-davinci-resolve-flux1-nim/
Date: 2025-04-10
Excerpt: "GeForce RTX 5090 Desktop and Laptop GPUs include three NVENC modules each, leading to significantly faster encoding speeds — more than 37% faster than the last generation."
Context: Split-frame encoding now available in consumer applications
Confidence: High

### 1.6 Concurrent Session Limits

Claim: "All NVENC-capable GPUs can encode up to 8 sessions before the driver starts to limit them."[^52^]
Source: StreamGuides.gg
URL: https://streamguides.gg/2024/01/nvenc-update-all-nvidia-geforce-cards-quietly-updated-to-8-encoding-sessions/
Date: 2024-01-30
Excerpt: "Nvidia released Game Ready Driver 551.23 on January 24... an increase to the number of possible NVENC encoder sessions, bringing the cap up to 8 simultaneous sessions."
Context: History: 2 sessions (originally) -> 3 (2020) -> 5 (March 2023) -> 8 (January 2024); Professional GPUs (Quadro/RTX PRO) have no artificial session limit
Confidence: High

**Session Limit Evolution:**

| Year | Consumer GeForce Limit | Driver/Change |
|------|----------------------|---------------|
| Pre-2020 | 2 sessions | Hard limit |
| 2020 | 3 sessions | Driver update |
| March 2023 | 5 sessions | Driver update[^122^] |
| January 2024 | 8 sessions | Game Ready Driver 551.23[^52^] |

Claim: "Nvidia has increased the number of concurrent NVENC encodes on consumer GPUs from three to five... Workstation-grade and data center-grade boards do not have any restrictions."[^122^]
Source: Tom's Hardware
URL: https://www.tomshardware.com/news/nvidia-increases-concurrent-nvenc-sessions-on-consumer-gpus
Date: 2023-03-24
Excerpt: "Consumer-oriented GeForce boards supported up to three simultaneous NVENC video encoding sessions while Nvidia's workstation and data center solutions... could support 11-17 concurrent NVENC sessions"
Context: Historical context for session limits
Confidence: High

### 1.7 Multi-Pass Rate Control

NVENC SDK 10.0+ supports multi-pass encoding for improved rate control:

| Mode | Description |
|------|-------------|
| none | 1-pass (fastest) |
| 2pass-quarter | First pass at quarter resolution |
| 2pass-full | First pass at full resolution (best quality) |

Source: NVEnc documentation / NVENC SDK[^60^]

### 1.8 4:2:2 10-bit Support (RTX 50 Exclusive)

Claim: "GeForce RTX 50 Series GPUs include 4:2:2 hardware support that can decode up to eight times the 4K 60 frames per second (fps) video sources per decoder."[^45^]
Source: NVIDIA Blog - RTX 50 Series announcement
URL: https://blogs.nvidia.com/blog/generative-ai-studio-ces-geforce-rtx-50-series/
Date: 2025-01-06
Excerpt: "For the first time in a consumer GeForce GPU, encoding and decoding video in the 4:2:2 color format for professional-grade higher color depth is supported."
Context: RTX 5080/5090 can import 5x 8K30 or 20x 4K30 streams; RTX PRO 6000: 10x 8K30 or 40x 4K30
Confidence: High

---

## 2. Intel QuickSync / AV1 Encoder

### 2.1 Intel Arc (Discrete) vs. Integrated QuickSync

**Intel Arc A-Series (Alchemist/Xe) Media Engine:**
- H.264: 8-bit 4:2:0 encode
- HEVC: 8-bit/10-bit 4:2:0, 10-bit 4:2:2 encode
- AV1: 8-bit/10-bit 4:2:0 encode
- VP9: Decode only (8/10-bit 4:2:0, 4:4:4, 12-bit)

Source: Intel official codec support matrix[^47^]

**Intel Arc B-Series (Battlemage/Xe2):**

Claim: "The B580 also supports dual media engines, each with an encoder and decoder, for up to two 8K 10-bit workloads."[^135^]
Source: Puget Systems - Intel Arc B580 Content Creation Review
URL: https://www.pugetsystems.com/labs/articles/intel-arc-b580-content-creation-review/
Date: 2024-12-12
Excerpt: "The B580 also supports dual media engines, each with an encoder and decoder, for up to two 8K 10-bit workloads."
Context: Second-gen Xe2 architecture; HEVC 8, 10, 12-bit (decode only), 4:2:0, 4:2:2 encode support; AV1 encode/decode
Confidence: High

Claim: "Intel's media engine has acceleration for HEVC 8, 10, and 12-bit (decode only) 4:2:0, 4:2:2, and 4:2:0, as well as AV1. The 10-bit 4:2:2 is of particular note, as currently, only Intel iGPUs and Alchemist cards support this common variant of H.265."[^135^]
Source: Puget Systems
URL: https://www.pugetsystems.com/labs/articles/intel-arc-b580-content-creation-review/
Date: 2024-12-12
Excerpt: "The 10-bit 4:2:2 is of particular note, as currently, only Intel iGPUs and Alchemist cards support this common variant of H.265."
Context: Intel's key differentiator for professional workflows
Confidence: High

**Intel Arc B580 Key Specs:**
- Architecture: Xe2 (Battlemage)
- Media Engines: 2 dual-format transcoders (MFX)
- Max encode: 8K60 10-bit HDR
- Codecs: AVC, HEVC (including 4:2:2 10-bit), AV1, VP9
- TDP: 190W (retail) / 160W (OEM)

Source: Multiple sources[^135^][^134^][^165^]

### 2.2 Ultra Low-Latency (ULL) Mode Performance

Claim: "The Intel encoder achieved the lowest latency of 5 frames (83 ms) for H.265/HEVC and AV1 with no significant additional RD penalty compared to its standard Low-Latency mode."[^10^]
Source: IEEE/arXiv - Evaluation of GPU Video Encoder for Low-Latency Real-Time 4K UHD Encoding
URL: https://arxiv.org/html/2511.18688v2
Date: 2025
Excerpt: "Ultra Low-Latency tuning proved most effective, with the Intel encoder achieving the lowest latency of 5 frames (83 ms) for H.265/HEVC and AV1 with no significant additional RD penalty"
Context: Intel provided best H.265/HEVC RD performance and lowest potential latency; non-standard structure may affect compatibility
Confidence: High (peer-reviewed IEEE study)

**End-to-End Latency Comparison (4K 60p, frames at 60fps):**

| Encoder | Normal Latency | Low Latency | Ultra Low Latency |
|---------|---------------|-------------|-------------------|
| Intel H.264 | ~10-12 frames | ~10-12 frames | ~8 frames |
| Intel HEVC | ~10-12 frames | ~10-12 frames | ~5 frames (83ms) |
| Intel AV1 | ~10-12 frames | ~10-12 frames | ~6 frames |
| NVIDIA | ~7 frames | ~7 frames | ~7 frames |
| AMD | ~6-9 frames | ~6-9 frames | ~6-9 frames |

Source: IEEE study[^10^]

### 2.3 Non-Standard B-Frame Compatibility Issue

Claim: "Analysis of the hardware encoder output revealed Normal Latency tuning used three B-frames. Low-Latency tuning disabled B-frames... Ultra Low-Latency... was analogous to the Zero Latency tune."[^10^]
Source: IEEE/arXiv
URL: https://arxiv.org/html/2511.18688v2
Date: 2025
Excerpt: "Intel provided the best H.265/HEVC RD performance and the lowest potential latency, though its use of non-standard structure may affect compatibility."
Context: Intel QuickSync API does not offer a direct latency tuning option; parameters follow OBS implementation
Confidence: High

### 2.4 Intel Arc Encode Session Limits

Claim: "There is no theoretical limit to the number of concurring videos encodes that can happen simultaneously, at least not from the driver and graphics unit perspective. Generally, Intel Arc can support 4K@60fps with four AV1 encoder instances working simultaneously."[^145^]
Source: Intel Support
URL: https://www.intel.com/content/www/us/en/support/articles/000093450/graphics/intel-arc-dedicated-graphics-family.html
Date: 2023-07-29
Excerpt: "There is no theoretical limit to the number of concurring videos encodes that can happen simultaneously, at least not from the driver and graphics unit perspective."
Context: Unlike NVIDIA, Intel does not artificially limit encode sessions on consumer hardware
Confidence: High

### 2.5 Intel Arc B580 vs. Competition - Quality

Claim: "In real-world testing, Intel's AV1 encoder provides better image quality than the H.264 encoder (NVIDIA's NVENC and AMD's VCN) at the same bitrate, especially when streaming at lower bitrates."[^153^]
Source: MiGoVi - Intel Arc B580 Review
URL: https://migovi.com/en/intel-arc-b580-limited-edition-review/
Date: 2025
Excerpt: "In real-world testing, Intel's AV1 encoder provides better image quality than the H.264 encoder (NVIDIA's NVENC and AMD's VCN) at the same bitrate, especially when streaming at lower bitrates."
Context: Intel's competitive advantage in AV1 quality-per-bit
Confidence: Medium (review site)

---

## 3. AMD AMF / RDNA4 Media Engine

### 3.1 RDNA4 Media Engine Architecture

Claim: "Dual media engines and an updated Radiance Display Engine are present on the GPU as well. The media engines support H.264, HEVC, and AV1 accelerated encode and decode, and offer better output quality and increased performance (up to 30% at 720p), at lower power."[^19^]
Source: Hot Hardware - AMD RDNA 4 Architecture Deep Dive
URL: https://hothardware.com/reviews/amd-rdna-4-architecture-deep-dive
Date: 2025-02-28
Excerpt: "Dual media engines and an updated Radiance Display Engine are present on the GPU as well. The media engines support H.264, HEVC, and AV1 accelerated encode and decode"
Context: RDNA4 / VCN 5.0 architecture details
Confidence: High

### 3.2 Quality Improvements

Claim: "There's a roughly 25% increase in H.264 low latency encode quality, and 11% improvement with HEVC."[^19^]
Source: Hot Hardware - AMD RDNA 4 Architecture Deep Dive
URL: https://hothardware.com/reviews/amd-rdna-4-architecture-deep-dive
Date: 2025-02-28
Excerpt: "There's a roughly 25% increase in H.264 low latency encode quality, and 11% improvement with HEVC. Basically, RDNA 4's new media engine can more quickly produce cleaner, less blocky output, from the same input data, thanks to upgrades to its motion estimation algo and multi-frame reference capabilities."
Context: AMD is also claiming a 50% performance uplift for AV1 and VP9 decode
Confidence: High (reported from AMD architecture briefing)

### 3.3 AV1 B-Frame Support

Claim: "With RDNA 4, AMD has quietly delivered one of its most important generational upgrades — not in raw gaming performance, but in media encoding. The new architecture introduces hardware-accelerated AV1 encoding with full B-frame support."[^42^]
Source: Kad8.com - AMD RDNA 4 Unlocks AV1 Encoding
URL: https://www.kad8.com/news/amd-rdna-4-graphics-card-has-av1-encoding-capability/
Date: 2024-12-31
Excerpt: "The addition of B-frames (Bi-predictive frames) is the defining capability of VCN 5.0. B-frames reference both previous and future frames during compression, allowing the encoder to dramatically reduce redundant data."
Context: Previously AMD used I/P-only AV1 encoding on RDNA3; now competitive with NVENC
Confidence: High

**RDNA4 AV1 B-Frame Impact:**

| Feature | RDNA 3 (VCN 4.0) | RDNA 4 (VCN 5.0) |
|---------|-----------------|-----------------|
| AV1 Encoding | I / P Frames | I / P / B Frames |
| Encoding Throughput | Baseline | ~2x Higher |
| AVC / HEVC Quality | Standard | ~25% Improvement |
| Latency Profile | General | Optimized for Low-Latency |

Source: Kad8 analysis[^42^]

### 3.4 No Session Limits

Claim: "No limit on number of sessions/encode streams"[^138^]
Source: AMD RX 9070 Series Partner Hub
URL: https://www.amd.com/content/dam/amd/en/documents/partner-hub/radeon/radeon-rx-9070-series-how-to-sell-non-competitive.pdf
Date: 2025
Excerpt: "Supporting H.264, HEVC, AVI; Up to 8K 80 FPS max encode/decode; No limit on number of sessions/encode streams"
Context: AMD's competitive response to NVIDIA's session limits
Confidence: High (official AMD partner document)

### 3.5 AMF SDK API - Usage Modes

The AMD Media Framework (AMF) SDK exposes six usage modes for the encoder:[^84^]

| Usage Mode | Latency | Quality Preset | Key Characteristics |
|------------|---------|---------------|-------------------|
| Transcoding | Low | Balanced | General purpose, 3-frame pipeline |
| Ultra Low Latency | Ultra-low | Speed | LOWLATENCY_MODE=true, HRD enforced |
| Low Latency | Low | Speed | For streaming |
| Webcam | Low | Speed | Optimized for webcam input |
| HQ | Normal | Quality | Pre-analysis enabled |
| HQLL | Low | Quality | High quality + low latency |

Source: AMF Video Encode API documentation[^84^]

**Key AMF Parameters:**
- `AMF_VIDEO_ENCODER_USAGE` - Selects usage mode
- `AMF_VIDEO_ENCODER_INSTANCE_INDEX` - Selects encoder engine (0 or 1 for dual-engine GPUs)
- `AMF_VIDEO_ENCODER_LOWLATENCY_MODE` - Enables low latency mode
- `AMF_VIDEO_ENCODER_QUALITY_PRESET` - Speed/Balanced/Quality
- `AMF_VIDEO_ENCODER_MAX_CONSECUTIVE_BPICTURES` - B-frame count (0-3)
- `AMF_VIDEO_ENCODER_RATE_CONTROL_METHOD` - CQP/VBR/CBR/QVBR

### 3.6 Real-World Encoding Performance

Claim: "HEVC (H.265) at Constant Quality (CQ) 27 runs at 300-700 fps, scoring ~90 VMAF; AV1 at CQ 70 runs just as fast, landing a solid 93 VMAF"[^158^]
Source: Medium - User Experience with Radeon 9070XT
URL: https://medium.com/@ariffinsetya/my-nas-was-dwindling-so-i-embraced-av1-and-hevc-with-a-radeon-9070xt-d1f27209d8db
Date: 2025-04-23
Excerpt: "HEVC (H.265) at Constant Quality (CQ) 27 runs at 300-700 fps, scoring ~90 VMAF; AV1 at CQ 70 runs just as fast, landing a solid 93 VMAF"
Context: Real-world user encoding library of media content; significant speedup vs CPU encoding
Confidence: Medium (user report)

### 3.7 AMD vs. Competition Quality Comparison

Claim: "For all the hype about AV1 encoding, in practice it really doesn't look or feel that different from HEVC... From an overall quality and performance perspective, Nvidia's latest Ada Lovelace NVENC hardware comes out as the winner with AV1 as the codec of choice."[^159^]
Source: Tom's Hardware - AMD GPUs Still Lag Behind Nvidia, Intel
URL: https://www.tomshardware.com/news/amd-intel-nvidia-video-encoding-performance-quality-tested
Date: 2023-03-09
Excerpt: "AMD's latest RDNA 3 encoder still falls behind. It can be reasonably fast at encoding, but even Nvidia's Pascal era hardware generally delivers superior results."
Context: Pre-RDNA4 testing; AMD has significantly improved with RDNA4 VCN 5.0
Confidence: Medium (pre-RDNA4 data; RDNA4 substantially improved)

---

## 4. Apple VideoToolbox / Media Engine

### 4.1 Media Engine Capabilities by Chip Generation

**M3 Series (October 2023):**
- M3/M3 Pro: 1 video decode engine, 1 video encode engine, 1 ProRes encode/decode engine, AV1 **decode**
- M3 Max: 1 video decode engine, **2** video encode engines, **2** ProRes encode/decode engines, AV1 decode

Claim: "New for M3 is the addition of AV1 decoding. This is on top of the existing functionality of dealing with H.264, HEVC, ProRes, and ProRes RAW footage."[^48^]
Source: AppleInsider
URL: https://forums.appleinsider.com/discussion/234231/m3-vs-m3-pro-vs-m3-max-specs-features-compared
Date: 2023-11-07
Excerpt: "New for M3 is the addition of AV1 decoding. This is on top of the existing functionality of dealing with H.264, HEVC, ProRes, and ProRes RAW footage."
Context: M3 did NOT include AV1 encoding - decode only
Confidence: High

**M4 Series (May 2024):**
- M4: 1 video decode engine, 1 video encode engine, 1 ProRes encode/decode engine, AV1 **decode**[^115^]
- M4 Pro: 1 video decode engine, 1 video encode engine, 1 ProRes encode/decode engine, AV1 decode[^118^]
- M4 Max: 1 video decode engine, **2** video encode engines, **2** ProRes encode/decode engines, AV1 decode[^50^]

Claim: "The media engine of the M4 supports multiple codecs, including H.264, HEVC, ProRes and now AV1."[^109^]
Source: Bitmovin - Apple AV1 Support
URL: https://bitmovin.com/blog/apple-av1-support/
Date: 2024
Excerpt: "The media engine of the M4 supports multiple codecs, including H.264, HEVC, ProRes and now AV1, making it the most advanced media processor ever in an iPad."
Context: Note - "AV1 support" here refers to decode only, not encode
Confidence: High

**M3 Ultra (Mac Studio):**
- 2 video decode engines
- **4** video encode engines
- **4** ProRes encode/decode engines
- AV1 decode

Source: Apple Mac Studio Technical Specifications[^116^]

### 4.2 AV1 Encode Status on Apple Silicon

**Important finding: Apple Silicon does NOT currently support AV1 hardware encoding.** The M3 and M4 series chips support AV1 hardware decoding only. There are no Apple chips with AV1 hardware encode capability as of 2025.

The M4 iPad Pro was the first Apple device to mention AV1 in its media engine, but this refers to decode capability. Apple's VideoToolbox framework supports software AV1 encoding, but not hardware-accelerated AV1 encoding.

### 4.3 Dolby Vision Profile 10 Support (AV1)

Claim: "During WWDC24, Apple shared a 'What's new in HTTP Live Streaming 2024' doc... they called out support for using Dolby Vision Profile 10, which is Dolby's 10-bit AV1 aware profile."[^109^]
Source: Bitmovin
URL: https://bitmovin.com/blog/apple-av1-support/
Date: 2024
Excerpt: "Apple now supports 3 different Dolby Vision profiles: 10, 10.1 and 10.4. Profile 10 is 'true' Dolby Vision."
Context: AV1 playback support through Dolby Vision Profile 10
Confidence: High

### 4.4 VideoToolbox API

The VideoToolbox framework provides `VTCompressionSession` for hardware-accelerated encoding. Key properties include:

- `kVTCompressionPropertyKey_MaxKeyFrameInterval` - GOP size control
- `kVTCompressionPropertyKey_AllowFrameReordering` - B-frame control
- `kVTCompressionPropertyKey_AverageBitRate` - Target bitrate
- `kVTCompressionPropertyKey_DataRateLimits` - VBV control
- `kVTCompressionPropertyKey_ProfileLevel` - Profile/level selection
- `kVTCompressionPropertyKey_H264EntropyMode` - CABAC/CAVLC

Source: Apple Developer Documentation / Stack Overflow[^83^]

---

## 5. Linux VAAPI (Video Acceleration API)

### 5.1 Vendor-Agnostic Architecture

Claim: "VA API comes close to [answering 'can I use the hardware encoder']. And with that I can use my (integrated) AMD GPU with hardware independent user space encoders."[^148^]
Source: lazy-evaluation.net - Hardware Independent Accelerated Video Processing in Linux
URL: https://blog.lazy-evaluation.net/posts/linux/linux-vaapi.html
Date: 2023-12-30
Excerpt: "Even though I can encode movies roughly 9 times faster (and with minimal CPU load) with VA API than with libx265, the resulting bitrate would not be as good"
Context: VAAPI provides cross-vendor abstraction; quality tradeoffs exist vs. software encoding
Confidence: High

### 5.2 Capability Detection with vainfo

The `vainfo` command queries supported profiles and entrypoints:

```bash
$ vainfo
VAEntrypointVLD      = hardware decode
VAEntrypointEncSlice = hardware encode
VAEntrypointVideoProc = video processing
```

**Example: Intel Alder Lake (iGPU)**[^141^]
```
# Encoding Capabilities:
JPEG    Baseline
H264    Main        High        ConstrainedBaseline
HEVC    Main        Main10      Main444     Main444_10
        SccMain     SccMain10   SccMain444  SccMain444_10
VP9     Profile0    Profile1    Profile2    Profile3
```

**Example: AMD Radeon Vega (integrated)**[^148^]
```
VAProfileH264ConstrainedBaseline : VAEntrypointEncSlice
VAProfileH264Main               : VAEntrypointEncSlice
VAProfileH264High               : VAEntrypointEncSlice
VAProfileHEVCMain                : VAEntrypointEncSlice
```

**Example: AMD RX Vega (discrete)**[^143^]
```bash
sudo /usr/lib/jellyfin-ffmpeg/vainfo --display drm --device /dev/dri/renderD128
# Shows: VAEntrypointVLD = decode, VAEntrypointEncSlice = encode
```

### 5.3 Intel vs. AMD VAAPI Quality Differences

From the IEEE study[^10^], Intel's QuickSync (via VAAPI on Linux) generally delivers better rate-distortion performance than AMD's VCN implementation:

- Intel provides the best H.265/HEVC RD performance among hardware encoders
- AMD delivers predictable latency but comparatively lower RD performance
- Quality presets have minimal impact on latency for all hardware encoders

### 5.4 GStreamer Integration

```bash
# VA-API H.264 encode
GST_VAAPI_ALL_DRIVERS=1 gst-launch-1.0 videotestsrc ! vaapih264enc ! mp4mux ! filesink location=output.mp4

# VA-API HEVC encode
GST_VAAPI_ALL_DRIVERS=1 gst-launch-1.0 videotestsrc ! vaapih265enc ! mp4mux ! filesink location=output.mp4
```

Source: RHEL hardware acceleration guide[^69^]

### 5.5 FFmpeg VAAPI Usage

```bash
# Encode with VAAPI (generic, works on Intel and AMD)
ffmpeg -hwaccel vaapi -vaapi_device /dev/dri/renderD128 -i input.mp4 -c:v h264_vaapi output.mp4

# HEVC encode with VAAPI
ffmpeg -hwaccel vaapi -vaapi_device /dev/dri/renderD128 -i input.mp4 -c:v hevc_vaapi output.mp4
```

---

## 6. Encoder Capability Detection at Runtime

### 6.1 NVIDIA - NVML API

NVIDIA Management Library (NVML) provides encoder-specific queries:

```c
// Get encoder utilization
nvmlReturn_t nvmlDeviceGetEncoderUtilization(nvmlDevice_t device, 
    unsigned int *utilization, unsigned int *samplingPeriodUs);

// Get encoder statistics
nvmlReturn_t nvmlDeviceGetEncoderStats(nvmlDevice_t device,
    unsigned int *sessionCount, unsigned int *averageFps, 
    unsigned int *averageLatency);

// Query active encoder sessions
nvmlReturn_t nvmlDeviceGetEncoderSessions(nvmlDevice_t device,
    unsigned int *sessionCount, nvmlEncoderSessionInfo_t *sessionInfos);

// Get decoder utilization
nvmlReturn_t nvmlDeviceGetDecoderUtilization(nvmlDevice_t device,
    unsigned int *utilization, unsigned int *samplingPeriodUs);

// Query throttle reasons
nvmlReturn_t nvmlDeviceGetCurrentClocksThrottleReasons(nvmlDevice_t device,
    unsigned long long *clocksThrottleReasons);
```

Source: NVML API Reference Guide[^132^][^136^]

**NVML Clock Throttle Reasons (for thermal detection):**
- `NVML_CLOCKS_THROTTLE_REASON_GPU_IDLE`
- `NVML_CLOCKS_THROTTLE_REASON_APPLICATIONS_CLOCK_SETTING`
- `NVML_CLOCKS_THROTTLE_REASON_SW_POWER_CAP` (power limit)
- `NVML_CLOCKS_THROTTLE_REASON_HW_SLOWDOWN` (thermal)
- `NVML_CLOCKS_THROTTLE_REASON_HW_THERMAL` (thermal shutdown)

### 6.2 Windows - DXGI Queries

On Windows, DXGI adapter queries can enumerate GPU capabilities. The `IDXGIAdapter` interface provides basic GPU information, while DirectX Video Acceleration (DXVA) APIs can expose encode/decode capabilities.

Chromium source code shows DXGI-based detection patterns for Media Foundation video encode acceleration:[^150^]

```cpp
// DXGI device manager for hardware-accelerated encode
auto dxgi_device_manager_ = ...;
// Query adapter capabilities through DXGI
D3D11CreateDevice(..., &adapter, ...);
```

### 6.3 Linux - sysfs and Direct Rendering Manager

**GPU detection via sysfs:**
```bash
# Enumerate GPUs
lspci | grep -E "VGA|3D"

# Check loaded drivers
lsmod | grep -E "i915|amdgpu|nouveau|nvidia"

# DRM render nodes
ls /dev/dri/
# Typical output: card0  card1  renderD128  renderD129

# Intel GPU info
cat /sys/class/drm/card0/device/vendor  # 0x8086 for Intel
cat /sys/class/drm/card0/device/device  # Device ID

# AMD GPU info  
cat /sys/class/drm/card1/device/vendor  # 0x1002 for AMD
```

### 6.4 macOS - Metal API

On macOS, the Metal framework provides `MTLDevice` for GPU capability queries:

```objc
// Query GPU capabilities through Metal
id<MTLDevice> device = MTLCreateSystemDefaultDevice();
NSString *name = device.name;
uint64_t dedicatedMemory = device.recommendedMaxWorkingSetSize;
BOOL supportsBCTextureCompression = device.supportsBCTextureCompression;
```

Note: VideoToolbox encoder capabilities are queried through `VTSessionCopySupportedPropertyDictionary()` rather than Metal directly.[^83^]

### 6.5 Vulkan Video - Cross-Platform Detection

Vulkan Video extensions provide a cross-platform capability query mechanism:

```c
// Check for Vulkan Video encode support
VkPhysicalDeviceVideoFormatPropertiesKHR videoFormatProps = {...};
vkGetPhysicalDeviceVideoFormatPropertiesKHR(physicalDevice, &videoFormatInfo, 
    &formatCount, videoFormatProps);
```

**Vulkan Video Extension Status:**[^87^][^88^]

| Extension | Status | Drivers Available |
|-----------|--------|-----------------|
| VK_KHR_video_decode_h264 | Final | NVIDIA, AMD, Intel |
| VK_KHR_video_decode_h265 | Final | NVIDIA, AMD, Intel |
| VK_KHR_video_decode_av1 | Final (1.3.277) | NVIDIA, AMD beta |
| VK_KHR_video_encode_h264 | Final (1.3.274) | NVIDIA, AMD beta |
| VK_KHR_video_encode_h265 | Final (1.3.274) | NVIDIA, AMD beta |
| VK_KHR_video_encode_av1 | In Development | - |

Source: Khronos Vulkan Video announcements[^87^][^88^]

---

## 7. Go Integration Patterns

### 7.1 FFmpeg CGO Bindings (Recommended)

Claim: "Real FFmpeg bindings for Go. Not a wrapper. Not a CLI tool. The actual libraries. Hardware acceleration included. Zero runtime dependencies."[^55^]
Source: ffmpeg-statigo (GitHub)
URL: https://github.com/linuxmatters/ffmpeg-statigo
Date: 2025
Excerpt: "Cross-platform, static FFmpeg libraries bundled directly into your Go binary. Hardware acceleration included. Zero runtime dependencies."
Context: Hard fork of csnewman/ffmpeg-go; FFmpeg 8.0.x; Go 1.24
Confidence: High (active project, well-documented)

**ffmpeg-statigo Supported Hardware Acceleration:**
- NVENC/NVDEC (NVIDIA)
- QuickSync (Intel)
- VA-API (Linux - Intel/AMD)
- VideoToolbox (macOS)
- Vulkan Video

```go
// Example: Using ffmpeg-statigo for hardware-accelerated encoding
// The library provides FFmpeg 8.0 bindings with hardware acceleration
```

### 7.2 Alternative: go-media FFmpeg Bindings

Source: mutablelogic/go-media[^56^]
URL: https://github.com/mutablelogic/go-media
Date: 2026 (active)
Excerpt: "Low-level CGO bindings for FFmpeg 8.0 (in sys/ffmpeg80). High-level Go API for media operations (in pkg/ffmpeg). Hardware acceleration support."
Context: Complete FFmpeg 8.0 bindings with hardware acceleration
Confidence: High

### 7.3 Direct NVENC SDK Integration via CGO

For direct NVENC SDK integration without FFmpeg:

1. **Link against NVENC libraries:**
   - Windows: `nvEncodeAPI.dll` (loaded at runtime via `LoadLibrary`)
   - Linux: `libnvidia-encode.so` (loaded at runtime via `dlopen`)

2. **C API pattern:**
   The NVENCODE API uses a C-API with function pointer tables:
   ```c
   // Client loads library and retrieves function pointer table
   NvEncodeAPIGetMaxSupportedVersion()
   NvEncodeAPICreateInstance()
   // Function pointer table provides access to encoder functions
   ```

3. **Go CGO approach:**
   ```go
   // #cgo LDFLAGS: -lnvidia-encode
   // #include <nvEncodeAPI.h>
   import "C"
   ```

Source: NVENC Video Encoder API Programming Guide[^157^]

### 7.4 FFmpeg Hardware Acceleration Flags for Go

When using FFmpeg from Go (via CGO bindings), key flags for hardware encoding:

**NVIDIA NVENC:**
```bash
-c:v h264_nvenc -preset p4 -tune ll -rc cbr -b:v 6M
-c:v hevc_nvenc -preset p7 -tune uhq -rc vbr -cq:v 20
-c:v av1_nvenc -preset p7 -tune hq -rc vbr
```

**Intel QuickSync:**
```bash
-c:v h264_qsv -preset medium
-c:v hevc_qsv -preset quality
-c:v av1_qsv -preset medium
```

**AMD AMF:**
```bash
-c:v h264_amf -quality balanced -rc cbr
-c:v hevc_amf -quality quality -rc vbr_latency
-c:v av1_amf -quality balanced -rc cbr
```

**VAAPI (Linux):**
```bash
-c:v h264_vaapi
-c:v hevc_vaapi
```

**Apple VideoToolbox:**
```bash
-c:v h264_videotoolbox
-c:v hevc_videotoolbox
```

### 7.5 CUDA Go Bindings

For projects requiring direct CUDA integration alongside encoding:

- **gocu** - Go bindings for CUDA Driver and Runtime APIs, cuBLAS, cuDNN[^140^]
- **gocudnn** - Go bindings for cuDNN[^144^]
- **cuda-go** - Golang and CUDA bindings[^146^]

These can be used with the NVENC SDK for frame processing before/after encode.

---

## 8. Thermal Throttling Impact on Encoder Performance

### 8.1 Thermal vs. Power Limit Throttling

Claim: "The thermal throttling lowers the performance by ~50%, which is still ~400MB/s, so it's more than fine."[^58^]
Source: Overclockers.com Forums
URL: https://www.overclockers.com/forums/threads/question-about-gpu-clock-speeds-affecting-video-encoding-time.806260/
Date: 2025-02-18
Excerpt: "The thermal throttling lowers the performance by ~50%, which is still ~400MB/s"
Context: Discussion of GPU clock speeds affecting encoding time; ~50% performance loss under thermal throttle
Confidence: Medium (forum discussion)

### 8.2 Encoder-Specific Thermal Impact

Claim: "What you're really experiencing is the thermal throttling because it's a laptop, and -any- use of the GPU will do that. If you're seeing a frame rate loss when the encoder is in use, it's simply that the weak GPU in the laptop was already at its thermal limit and using the hardware encoder just makes it hit that limit sooner."[^61^]
Source: Linus Tech Tips Forums
URL: https://linustechtips.com/topic/1573165-does-a-gpu-encoderdecoder-affect-gaming-performance/
Date: 2024-06-10
Excerpt: "What you're really experiencing is the thermal throttling because it's a laptop, and -any- use of the GPU will do that."
Context: NVENC/NVDEC hardware encoders are dedicated silicon blocks; minimal gaming performance impact unless thermal throttling
Confidence: Medium

### 8.3 Thermal Throttling Mechanisms

Claim: "Thermal throttling reduces clocks because temperatures are too high; power-limit throttling reduces clocks because the card is drawing too much power."[^142^]
Source: Quora (technical analysis)
URL: https://www.quora.com/How-do-thermal-throttling-and-power-limit-throttling-affect-a-graphics-cards-performance-and-what-can-be-done-to-prevent-these-issues
Date: 2025
Excerpt: "Thermal throttling reduces clocks because temperatures are too high; power-limit throttling reduces clocks because the card is drawing too much power."
Context: NVML `nvmlDeviceGetCurrentClocksThrottleReasons` can distinguish between thermal and power limit throttling
Confidence: High

**Impact on Encoder Performance:**

| Scenario | Clock Reduction | Encode Performance Impact |
|----------|----------------|--------------------------|
| Thermal throttling | 20-50% | Proportional reduction in encode FPS |
| Power limit throttling | 10-30% | Reduced sustained encode throughput |
| VRAM throttling | Variable | May cause encode failures |

### 8.4 Monitoring Encoder Thermal State

Using NVML for runtime thermal monitoring:

```c
// Query GPU temperature
nvmlReturn_t nvmlDeviceGetTemperature(nvmlDevice_t device, 
    nvmlTemperatureSensors_t sensorType, unsigned int *temp);

// Query thermal throttle reasons
nvmlReturn_t nvmlDeviceGetCurrentClocksThrottleReasons(nvmlDevice_t device,
    unsigned long long *clocksThrottleReasons);

// Query performance state
nvmlReturn_t nvmlDeviceGetPerformanceState(nvmlDevice_t device, 
    nvmlPstates_t *pState);
```

Source: NVML API Reference[^132^]

---

## 9. Cross-Vendor Comparison Matrix

### 9.1 Feature Comparison (Latest Generation)

| Feature | NVIDIA RTX 50 (NVENC 9) | Intel Arc B580 (Xe2) | AMD RX 9070 (VCN 5.0) | Apple M4 Max |
|---------|------------------------|---------------------|----------------------|-------------|
| H.264 Encode | Yes | Yes | Yes | Yes |
| HEVC Encode | Yes + 4:2:2 10-bit | Yes + 4:2:2 10-bit | Yes 10-bit | Yes |
| AV1 Encode | Yes + UHQ mode | Yes | Yes + B-frames | No |
| ProRes Encode | No | No | No | Yes |
| Max Resolution | 8K240 (3 NVENC) | 8K60 (dual MFX) | 8K80fps (dual VCN) | 8K |
| Max Encode Sessions | 8 (consumer) | No driver limit | No driver limit | N/A |
| Low-Latency Mode | ll, ull tuning | ULL mode | lowlatency, ultralowlatency | Via VideoToolbox |
| Split-Frame Encode | Yes (2/3-way) | No | No | No |
| Linux Support | NVML + proprietary | VAAPI (iHD) | VAAPI (radeonsi) / AMF | N/A (macOS only) |

### 9.2 Quality Comparison (4K VMAF at comparable bitrates)

Claim: "From an overall quality and performance perspective, Nvidia's latest Ada Lovelace NVENC hardware comes out as the winner with AV1 as the codec of choice. Right behind Nvidia... Intel's Arc GPUs are also great for streaming purposes."[^159^]
Source: Tom's Hardware
URL: https://www.tomshardware.com/news/amd-intel-nvidia-video-encoding-performance-quality-tested
Date: 2023-03-09
Excerpt: "From an overall quality and performance perspective, Nvidia's latest Ada Lovelace NVENC hardware comes out as the winner with AV1 as the codec of choice... Right behind Nvidia in terms of quality and performance... Intel's Arc GPUs"
Context: Pre-RDNA4 testing; AMD has closed the gap significantly with RDNA4
Confidence: Medium (RDNA4 substantially improved results)

### 9.3 FFmpeg Encoder Names

| Vendor | H.264 | HEVC | AV1 | Linux VAAPI |
|--------|-------|------|-----|-------------|
| NVIDIA | `h264_nvenc` | `hevc_nvenc` | `av1_nvenc` | N/A |
| Intel | `h264_qsv` | `hevc_qsv` | `av1_qsv` | `h264_vaapi` / `hevc_vaapi` |
| AMD | `h264_amf` | `hevc_amf` | `av1_amf` | `h264_vaapi` / `hevc_vaapi` |
| Apple | `h264_videotoolbox` | `hevc_videotoolbox` | N/A | N/A |

---

## 10. Key Implementation Notes

### 10.1 NVENC SDK 13.0 (Blackwell) Key Changes

- **AV1 UHQ mode**: Available on RTX 40 via software update (lower quality than Blackwell native)
- **4:2:2 encode**: First time on consumer GeForce (H.264, HEVC)
- **MV-HEVC**: For stereoscopic/3D video
- **Split-frame encoding**: 2-way and 3-way (RTX 5090 only)
- **Up to 8K240 fps**: With 4 NVENCs on Blackwell Pro GPUs

Source: NVIDIA Video Codec SDK[^164^]

### 10.2 Encoder Latency Recommendations by Use Case

| Use Case | Recommended Encoder | Preset/Tune | Expected Latency |
|----------|-------------------|-------------|-----------------|
| Cloud Gaming | NVENC P4 | `ll` + CBR | ~7 frames |
| Live Streaming | NVENC P4-P5 | `ll` + CBR | ~7 frames |
| Video Conferencing | Intel QSV | ULL mode | ~5-6 frames |
| Ultra-low-latency | NVENC P1 | `ull` + CBR | Minimal |
| Archival/Transcoding | NVENC P7 | `uhq` + VBR | Not critical |

### 10.3 Go Implementation Checklist

1. **Capability Detection**: Query `vainfo` (Linux), NVML (NVIDIA), or VideoToolbox (macOS) before initializing encoder
2. **FFmpeg Binding**: Use `ffmpeg-statigo` or `go-media` for static FFmpeg integration with hwaccel
3. **Preset Selection**: Default to P4 (NVENC), medium (QSV), balanced (AMF) for general use
4. **Thermal Monitoring**: Use NVML to detect throttling and adjust quality/bitrate dynamically
5. **Session Management**: Track active encode sessions (NVML `nvmlDeviceGetEncoderStats`) to stay within GPU limits
6. **Fallback Strategy**: Implement software encode fallback when hardware encoder unavailable or at capacity

---

## References

[^10^] IEEE/arXiv - Evaluation of GPU Video Encoder for Low-Latency Real-Time 4K UHD Encoding, 2025
[^19^] Hot Hardware - AMD RDNA 4 Architecture Deep Dive, 2025-02-28
[^25^] Puget Systems - Verifying NVIDIA GeForce RTX 50 Series Performance, 2025
[^42^] Kad8.com - AMD RDNA 4 Unlocks AV1 Encoding with B-Frame Support, 2024-12-31
[^44^] Wikipedia - GeForce RTX 50 series
[^45^] NVIDIA Blog - RTX 50 Series Creative Performance, 2025-01-06
[^47^] Intel - Video Codecs Supported by Intel Arc GPUs, 2024-03-22
[^48^] AppleInsider - M3 vs M3 Pro vs M3 Max Specs, 2023-11-07
[^50^] Apple - MacBook Pro 14-inch M4 Max Tech Specs
[^52^] StreamGuides.gg - NVENC 8 Encoding Sessions Update, 2024-01-30
[^55^] GitHub - ffmpeg-statigo, 2025
[^56^] GitHub - mutablelogic/go-media, 2026
[^58^] Overclockers.com Forums - GPU Clock Speeds Affecting Encoding, 2025
[^59^] NVIDIA - NVENC Video Encoder API Programming Guide (SDK 13.0)
[^60^] GitHub - NVEnc Options Documentation
[^61^] LTT Forums - GPU Encoder/Decoder Thermal Throttling, 2024
[^84^] GitHub - AMF Video Encode API Documentation
[^87^] Khronos - Vulkan Video AV1 Decode Extension, 2024-02-01
[^88^] Khronos - Vulkan Video H.264/H.265 Encode Extensions, 2023-12-19
[^109^] Bitmovin - Apple AV1 Support, 2024
[^113^] NVIDIA - RTX Blackwell GPU Architecture Whitepaper
[^115^] Apple - MacBook Pro 14-inch M4 Tech Specs
[^116^] Apple - Mac Studio (2025) Tech Specs
[^117^] NVIDIA Blog - DaVinci Resolve 20 + RTX 50, 2025-04-10
[^118^] Apple - MacBook Pro 14-inch M4 Pro Tech Specs
[^122^] Tom's Hardware - NVIDIA Increases NVENC Sessions, 2023-03-24
[^132^] NVIDIA - NVML API Reference (Device Queries)
[^135^] Puget Systems - Intel Arc B580 Content Creation Review, 2024-12-12
[^136^] NVIDIA - NVML API Reference Guide (PDF)
[^138^] AMD - RX 9070 Series Partner Hub
[^140^] GitHub - gocu (CUDA Go Bindings)
[^141^] Ubuntu MATE Community - vainfo AV1 detection
[^142^] Quora - Thermal vs Power Limit Throttling
[^143^] Jellyfin - AMD GPU Hardware Acceleration Guide
[^145^] Intel - Simultaneous AV1 Encodes on Arc GPUs
[^148^] lazy-evaluation.net - VAAPI on Linux
[^150^] Chromium - Media Foundation Video Encode Accelerator (DXGI)
[^152^] WCCFTech - Intel Arc B580 Battlemage Review
[^153^] MiGoVi - Intel Arc B580 Limited Edition Review
[^154^] NVIDIA - Video Codec SDK 13.0 Read Me
[^157^] NVIDIA - NVENC Video Encoder API Programming Guide
[^158^] Medium - Radeon 9070XT AV1/HEVC Encoding Experience
[^159^] Tom's Hardware - AMD GPUs Lag in Encoding Quality, 2023-03-09
[^163^] TechPowerUp - NVIDIA Claims AV1 Superiority, 2023-05-03
[^164^] NVIDIA - Video Codec SDK Official Page
[^165^] Intel - Arc B-Series Launch Newsroom
[^194^] GitHub Gist - ffmpeg -h encoder=hevc_nvenc
[^195^] scottstuff.net - Benchmarking FFMPEG H.265 Options, 2025-03-17
[^198^] CodeCalamity - RTX 5090 Video Encoding First Look, 2025-02-14
[^200^] Overclockers.com - GPU Clock Speeds Affecting Encoding
