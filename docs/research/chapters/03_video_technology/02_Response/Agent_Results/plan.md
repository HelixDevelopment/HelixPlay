# Plan: Comprehensive Video Technology Research for CloudStream Gaming Platform

## Context
This is Stage 5 of the CloudStream research pipeline. Previous stages established:
1. Initial system architecture (Go-based cloud gaming)
2. Core streaming protocols & codecs (WebRTC, H.264/HEVC/AV1)
3. Controller input forwarding system
4. Latency reduction research

Now we need deep research on VIDEO TECHNOLOGY for:
- Real-time streaming + simultaneous recording
- Ultra-high resolution codecs with zero latency impact
- Configurable storage backends
- Full multi-channel audio
- Hardware acceleration maximization
- Testing strategy with implementation details

## Stage 1 — Deep Research (Parallel Agents)
Load: `deep-research-swarm`

### Research Dimensions (parallel sub-agents):
1. **Video Codec Technology Researcher**
   - Real-time codecs: H.264, HEVC, AV1, VVC/H.266, VP9, VP10
   - Hardware encode acceleration: NVENC, QuickSync, AMF, VideoToolbox, VAAPI
   - Zero-copy capture-to-encode pipelines
   - Frame pacing, V-Sync bypass, HDR pipelines
   - Emerging: NVIDIA Frame Warp, DLSS 4, FSR 4

2. **Recording & Storage Pipeline Researcher**
   - Simultaneous streaming + recording architectures
   - GPU memory split encoding (stream + record)
   - Storage backends: SMB, NFS, FTP, WebDAV, S3, local, custom pipelines
   - Container formats: MP4, MKV, MOV, fragmented MP4 for live
   - Write-through vs write-back strategies

3. **Audio Technology Researcher**
   - Real-time audio codecs: Opus, AAC, AC3/Dolby Digital, E-AC3, Dolby Atmos, DTS:X
   - Multi-channel: PCM, Stereo, 5.1, 7.1, 7.1.4 (Atmos)
   - Audio passthrough, HDMI ARC/eARC, SPDIF
   - Hardware audio acceleration (GPU audio offload)

4. **Hardware Detection & Optimization Researcher**
   - GPU capability detection (NVML, sysfs, DirectX queries)
   - Thermal throttling detection and mitigation
   - Dynamic quality adjustment based on hardware load
   - OS-specific optimizations (Windows DWM, macOS compositor, Linux Wayland/X11)
   - Memory bandwidth optimization for 4K/8K

5. **Testing & Validation Researcher**
   - Frame loss detection (presented vs captured vs encoded)
   - Latency measurement techniques
   - Automated test suites for codec validation
   - Stress testing scenarios
   - Benchmarking methodologies

## Stage 2 — Report Writing
Load: `report-writing`

Using Stage 1 findings, produce a comprehensive research document covering:
- Executive summary
- Video codec deep-dive with benchmarks
- Recording architecture with storage backends
- Audio pipeline design
- Hardware acceleration and detection
- Testing strategy with code examples
- Implementation phases with Go code
- Diagrams and wireframes
- Risk analysis

## Stage 3 — Document Conversion
Load: `docx`

Convert final markdown to Word document for delivery.

## Output
- `/mnt/agents/output/video-technology-research.md` — Final markdown
- `/mnt/agents/output/video-technology-research.docx` — Word document
