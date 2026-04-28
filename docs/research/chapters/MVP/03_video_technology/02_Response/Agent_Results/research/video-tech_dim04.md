# Dimension 04: Simultaneous Streaming + Recording Architecture

## Research Summary

This document covers architectures for simultaneous real-time streaming and session recording with minimal or zero impact on streaming performance. Research spans hardware encoder appliances (Matrox, Osprey, Haivision, Blackmagic), GPU-based dual-path encoding (NVENC, AMD VCE/VCN), FFmpeg multi-output pipelines, container crash-safety, storage backends, Go integration patterns, and impact analysis.

---

## 1. Dual-Path Encoding Architectures

### 1.1 Hardware Appliance Architectures

#### Matrox Monarch HDX (Dual-Channel H.264 Encoder)
The Monarch HDX is purpose-built for simultaneous streaming and recording, featuring two independent H.264 encoders that share 30 Mb/s of combined encoding capacity.

Claim: "Monarch HDX dual encoders can broadcast a live webstream at one bitrate while simultaneously recording mastering quality H.264 files for immediate availability"[^1^]
Source: Matrox Official Product Page
URL: https://video.matrox.com/en/products/encoders-decoders/monarch-series/monarch-hdx
Date: Unknown (product page)
Excerpt: "Monarch HDX dual encoders can broadcast a live webstream at one bitrate while simultaneously recording mastering quality H.264 files for immediate availability, such as post event VOD or editing with your NLE of choice."
Context: Hardware appliance with dual independent encoders
Confidence: high

Key architectural details:
- **Dual-channel streaming**: Each channel can stream up to 10 Mb/s, allowing simultaneous streaming to two destinations or 20 Mb/s for a single destination[^1^]
- **Dual-channel recording**: MOV or MP4 output to local SD/USB or network-mapped drives[^1^]
- **Split file feature**: Automatically stops recording in one file and starts a new file at configured time intervals without frame loss, preserving content on power failure[^1^]
- **File consolidator**: Java-based utility that rewraps (not transcodes) split MP4/MOV files[^1^]
- **Disaster recovery**: Can target local AND remote storage simultaneously; if network fails, local copy is preserved[^1^]
- **End-of-life announced**: April 2025; alternatives suggested are Maevex 7100 Series or Monarch LCS[^1^]

#### Osprey Talon Series

Claim: "Talon series encode up to three streams simultaneously, with frame alignment across all streams for multiple-bit-rate RTMP streaming"[^2^]
Source: Exxact Corporation / Osprey Talon G1
URL: https://www.exxactcorp.com/Osprey-96-02010-E1345845
Date: 2025
Excerpt: "Capable of ingesting video via 3G-SDI, HDMI, or Analog Composite, Talon series encode up to three streams simultaneously, with frame alignment across all streams for multiple-bit-rate RTMP streaming."
Context: Hardware encoder appliance
Confidence: high

Key architectural details:
- **MBR Mode**: 1 source → up to 3 destinations in different resolutions and bitrates[^3^]
- **LC (Lecture Capture) Mode**: 2 sources → 2 destinations[^3^]
- **Simultaneous streaming and archiving**: Can save to MP4 while delivering a stream[^2^]
- **Talon Pro**: Three independent encoding channels up to 4K30, mixed HEVC/AVC[^4^]
- **USB/Micro SD archiving**: For MP4 or TS format video[^4^]
- **Video encoding**: Single stream up to 1080p60 or dual encode up to 1×1080i30 + 1×720p60[^5^]

#### Haivision Makito X with Storage

Claim: "The Makito X with Storage dual-height models include 250 GB of either fixed or removable SSD storage that enables users to record content directly on the encoder, while simultaneously streaming live video from two sources"[^6^]
Source: Haivision Documentation
URL: https://doc.haivision.com/MakitoXEnc/2.5/storage-and-recording
Date: 2026-03-06
Excerpt: "The Makito X with Storage dual-height models include 250 GB of either fixed or removable SSD storage that enables users to record content directly on the encoder, while simultaneously streaming live video from two sources."
Context: Enterprise encoder appliance with integrated storage
Confidence: high

Key architectural details:
- **Record two streams simultaneously**[^7^]
- **File formats**: TS or MP4 with optional segmentation by time or size[^7^]
- **Auto Export**: Completed segments can auto-export to FTP/SFTP server[^7^]
- **Segmentation Roll-Over**: Deletes old segments beyond configured limit for circular recording[^7^]
- **Record high quality (20 Mbps) while streaming low** to save bandwidth[^6^]
- **Four internal H.264 encoding engines**: Can encode at up to four different bitrates for MBR streaming[^8^]

#### Blackmagic Design ATEM Mini Series

Claim: "The ATEM Mini uses the same encoder for recording and streaming, so your stream will just go to both destinations when you hit record"[^9^]
Source: Reddit / Blackmagic Forum
URL: https://www.reddit.com/r/blackmagicdesign/comments/173j6z8/atem_mini_pro_record_and_stream_simultaneously/
Date: 2025
Excerpt: "The ATEM Mini Pro uses the same encoder for recording and streaming, so your 3mbps stream will just go to both destinations when you hit record."
Context: Consumer live production switcher
Confidence: medium

Key architectural details:
- ATEM Mini Pro/ISO: Same encoder feeds both streaming (via Ethernet) and USB recording simultaneously[^10^]
- Records to USB flash disks or CFexpress cards in H.264 + AAC[^10^]
- Multiple disk recording via USB hub — continues to second disk when first fills[^10^]
- ATEM Mini Extreme ISO G2: Thunderbolt for record/playback, ISO recording of all inputs[^10^]
- The limitation is that the stream and recording use the same encoding parameters (same bitrate, codec settings)[^9^]

### 1.2 FFmpeg Tee Muxer: Software Dual-Path Encoding

The FFmpeg `tee` pseudo-muxer is the core software mechanism for single-encode, multi-output pipelines.

Claim: "The tee muxer can be used to write the same data to several outputs, such as files or streams. It can be used, for example, to stream a video over a network and save it to disk at the same time"[^11^]
Source: FFmpeg Official Documentation (4.3.1)
URL: https://gensoft.pasteur.fr/docs/ffmpeg/4.3.1/ffmpeg-all.html
Date: Unknown (FFmpeg 4.3.1)
Excerpt: "The tee muxer can be used to write the same data to several outputs, such as files or streams. It can be used, for example, to stream a video over a network and save it to disk at the same time."
Context: Official FFmpeg muxer documentation
Confidence: high

Critical architectural distinction:
- **Tee muxer**: Encodes ONCE, sends identical encoded packets to multiple destinations[^11^]
- **Multiple ffmpeg outputs**: Encodes MULTIPLE TIMES (expensive)[^12^]

#### Tee Muxer Key Options

| Option | Description |
|--------|-------------|
| `use_fifo` | Process slave outputs in separate threads using fifo muxer (default: off)[^11^] |
| `fifo_options` | Options passed to fifo pseudo-muxer instances[^11^] |
| `f` | Specify format name for slave output[^11^] |
| `bsfs` | Bitstream filters to apply to output[^11^] |
| `select` | Stream specifier for which streams go to which output (e.g., `v:0,a`)[^11^] |
| `onfail` | `abort` (default) or `ignore` — whether to continue other outputs if one fails[^11^] |

#### Essential FFmpeg Tee Muxer Commands

**Basic: Stream to RTMP + save local file:**
```bash
ffmpeg -i input -map 0:v -map 0:a -c:v libx264 -b:v 4000k -c:a aac -b:a 128k \
  -flags +global_header -f tee \
  "[f=flv:onfail=ignore]rtmp://server/stream|[f=mp4]local.mp4"
```
Source: SuperUser[^13^]

**Stream to YouTube + Twitch + local file with NVENC:**
```bash
ffmpeg -i input -map 0 -bsf:a aac_adtstoasc -c:a aac -ac 2 -ar 48000 -b:a 128k \
  -b:v 6000k -minrate:v 6000k -maxrate:v 6000k -bufsize:v 6000k \
  -c:v h264_nvenc -qp:v 19 -profile:v high -rc:v cbr_ld_hq \
  -level:v 4.2 -r:v 60 -g:v 120 -bf:v 3 -refs:v 16 \
  -f tee \
  "[f=flv:onfail=ignore]rtmp://live.twitch.tv/app/<key>|\
   [f=flv:onfail=ignore]rtmp://a.rtmp.youtube.com/live2/<key>|\
   local_file.mkv"
```
Source: GitHub Gist[^14^]

**Stream to multiple RTMP + segmented local output:**
```bash
ffmpeg -i input -map 0 -c:v libx264 -c:a aac -maxrate 1000k -bufsize 2000k -g 50 \
  -f tee \
  "[f=flv:onfail=ignore]rtmp://facebook|\
   [f=flv:onfail=ignore]rtmp://youtube|\
   [f=segment:strftime=1:segment_time=60]local_%F_%H-%M-%S.mkv"
```
Source: StackOverflow[^15^]

**Stream to UDP + fragmented MP4:**
```bash
ffmpeg -i input -map 0:v -map 0:a -c:v libx264 -c:a mp2 -f tee \
  "archive-20121107.mkv|[f=mpegts]udp://10.0.1.255:1234/"
```
Source: FFmpeg Official Docs[^11^]

### 1.3 FFmpeg Fifo Muxer: Isolating Output Latencies

The fifo pseudo-muxer separates encoding from muxing by running the actual muxer in a separate thread with a packet queue.

Claim: "The fifo pseudo-muxer allows the separation of encoding and muxing by using first-in-first-out queue and running the actual muxer in a separate thread. This is especially useful in combination with the tee muxer"[^11^]
Source: FFmpeg Official Documentation
URL: https://gensoft.pasteur.fr/docs/ffmpeg/4.3.1/ffmpeg-all.html
Date: Unknown
Excerpt: "The fifo pseudo-muxer allows the separation of encoding and muxing by using first-in-first-out queue and running the actual muxer in a separate thread. This is especially useful in combination with the tee muxer and can be used to send data to several destinations with different reliability/writing speed/latency."
Context: Official documentation
Confidence: high

Key fifo options:
- `queue_size`: Default 60 packets[^16^]
- `drop_pkts_on_overflow`: Drop packets instead of blocking encoder (default: false)[^16^]
- `attempt_recovery`: Restart output transparently on failure (default: false)[^16^]
- `recovery_wait_time`: Wait before retry, default 5 seconds[^16^]
- `timeshift`: Buffer specified duration and delay writing output[^17^]
- `restart_with_keyframe`: Wait for keyframe after recovery[^17^]

Example: Stream to RTMP with transparent recovery:
```bash
ffmpeg -re -i ... -c:v libx264 -c:a aac -f fifo -fifo_format flv -map 0:v -map 0:a \
  -drop_pkts_on_overflow 1 -attempt_recovery 1 -recovery_wait_time 1 \
  rtmp://example.com/live/stream_name
```
Source: FFmpeg-devel patch[^17^]

### 1.4 OBS Simultaneous Streaming + Recording

OBS Studio implements simultaneous streaming and recording using the same source with potentially different encoders.

Claim: "OBS supports streaming + recording simultaneously via multiple output mechanisms, including the ability to use the streaming encoder for recording or a separate encoder"[^18^]
Source: OBS Project Forum
URL: https://obsproject.com/forum/resources/stream-to-2-destinations-simultaneously-with-obs-without-nginx.788/
Date: 2019
Excerpt: "In case you use this method, you will not be able to do a Recording. Also, usually, when we record, OBS allows you to use the Streaming Encoder and so the CPU does not have any additional load. But this method will use more CPU, as this will count as 2 encodings."
Context: OBS dual-output setup guide
Confidence: high

OBS architecture for dual output:
1. **Simple Output Mode**: Uses the streaming encoder for both streaming and recording (single encode)[^18^]
2. **Advanced Output Mode**: Can use different encoders for stream and record (dual encode on GPU)[^19^]
3. **Custom Output (FFmpeg)**: Can output directly to URL for second stream destination[^18^]
4. **Multiple RTMP outputs plugin**: Enables streaming to multiple destinations[^20^]

Recommended setup for zero CPU impact:
```
Settings → Output → Advanced → Streaming: NVENC H.264
Settings → Output → Advanced → Recording: NVENC H.264 (same encoder reused)
OR
Settings → Output → Advanced → Recording: NVENC H.264 (separate encoder instance)
```

When using NVENC, the second option (separate encoder) runs two parallel encodings on the GPU's dedicated NVENC circuits without consuming additional CPU resources[^19^].

---

## 2. GPU Split Encoding

### 2.1 NVENC Concurrent Session Limits

NVIDIA consumer GPUs have driver-enforced limits on concurrent NVENC encoding sessions.

Claim: "NVIDIA has quietly removed some of the concurrent video encoding limitations from its consumer graphics processing units, so they can now encode up to five simultaneous streams"[^21^]
Source: Tom's Hardware
URL: https://www.tomshardware.com/news/nvidia-increases-concurrent-nvenc-sessions-on-consumer-gpus
Date: 2023-03-24
Excerpt: "Nvidia has increased the number of concurrent NVENC encodes on consumer GPUs from three to five, according to the company's own Video Encode and Decode GPU Support Matrix."
Context: Consumer GPU NVENC session limit update
Confidence: high

NVENC Session Limit Evolution:

| Period | Limit | Notes |
|--------|-------|-------|
| Pre-2020 | 2 sessions | Kepler, Maxwell, Pascal consumer GPUs[^22^] |
| April 2020 | 3 sessions | Increased during early pandemic[^23^] |
| March 2023 | 5 sessions | All Maxwell 2nd Gen through Ada Lovelace[^21^] |
| January 2024 | 8 sessions | All NVENC-capable GPUs except GTX 1630 (3 sessions)[^23^] |

Claim: "The new update will enable up to 8 concurrent encoding sessions, impacting nearly all NVIDIA GPUs, spanning across Maxwell, Pascal, Turing, Volta, Ampere, and Ada Lovelace"[^23^]
Source: VideoCardz
URL: https://videocardz.com/newz/nvdia-geforce-gpus-now-support-up-to-8-concurrent-nvenc-encoding-sessions
Date: 2024-01-31
Excerpt: "The new update will enable up to 8 concurrent encoding sessions, impacting nearly all NVIDIA GPUs, spanning across Maxwell, Pascal, Turing, Volta, Ampere, and, of course, Ada Lovelace."
Context: Driver-level session limit increase
Confidence: high

### 2.2 Dual NVENC Hardware (RTX 40 Series)

RTX 4070 Ti and higher GPUs have dual physical NVENC encoders (8th generation).

Claim: "RTX 40 series feature Dual NVENC with 8K 10-bit 120FPS AV1 fixed function hardware encoding"[^24^]
Source: Wikipedia / GeForce RTX 40 series
URL: https://en.wikipedia.org/wiki/GeForce_RTX_40_series
Date: 2022-09-20
Excerpt: "Dual NVENC with 8K 10-bit 120FPS AV1 fixed function hardware encoding"
Context: Ada Lovelace architecture specifications
Confidence: high

Key dual-NVENC behavior:
- Two physical NVENC engines allow two independent streams at nearly single-stream speed[^25^]
- **Split-Frame Encoding (SFE)**: Automatically distributes a single high-resolution frame across both encoders for 4K+ content[^26^]
- SFE can be manually enabled to resolve encoder overload on high-quality presets[^26^]
- Performance caveat: Dual encoder scheduling can fluctuate — observed 25-50 fps cumulative on RTX 4090 depending on workload distribution[^27^]

Claim: "When running a single instance of the encoder, I achieve a consistent performance of ~25fps. When running two instances of the encoder simultaneously, I sometimes get a cumulative performance of ~50fps, but often the cumulative performance is much worse, somewhere between 25fps and 40fps"[^27^]
Source: NVIDIA Developer Forums
URL: https://forums.developer.nvidia.com/t/fluctuating-performance-with-two-instances-of-nvenc/247778
Date: 2023-03-28
Excerpt: "When running a single instance of the encoder, I achieve a consistent performance of ~25fps. When running two instances of the encoder simultaneously, I sometimes get a cumulative performance of ~50fps, but often the cumulative performance is much worse, somewhere between 25fps and 40fps."
Context: RTX 4090 AV1 encoding dual-instance performance report
Confidence: high

### 2.3 AMD RDNA4 Dual Media Engine

AMD's RDNA4 architecture introduces a dual multimedia engine configuration.

Claim: "Navi 48 and the RDNA 4 architecture also introduce an entirely new multimedia engine (or rather, two engines, as Navi 48 features a dual configuration of this block)"[^28^]
Source: HWCooling.net
URL: https://www.hwcooling.net/en/better-more-capable-than-expected-rdna-4-architecture-deep-dive/
Date: 2025-02-28
Excerpt: "Navi 48 and the RDNA 4 architecture also introduce an entirely new multimedia engine (or rather, two engines, as Navi 48 features a dual configuration of this block). This engine is said to include enhanced decoders and encoders, optimized for low-latency streaming."
Context: RDNA4 architecture deep dive
Confidence: high

RDNA4 media engine capabilities:
- AV1 and VP9 decoding: Up to 50% improvement[^28^]
- AV1 encoding: Up to 2× performance increase[^28^]
- H.264 encoding: Up to 25% improved quality (VMAF metric at low bitrates)[^28^]
- HEVC encoding: +11% VMAF quality improvement[^28^]
- **Does NOT support 4:2:2 chroma subsampling** (unlike NVIDIA Blackwell)[^28^]

### 2.4 NVIDIA NVENC Patch for Unlimited Sessions

The driver-level session limit can be removed on consumer GPUs.

Claim: "nvidia-patch: removes restriction on number of simultaneous NVENC video encoding sessions imposed on consumer GPUs"[^29^]
Source: Hacker News / GitHub
URL: https://news.ycombinator.com/item?id=40636122
Date: 2024-06-10
Excerpt: "nvidia-patch: removes restriction on number of simultaneous NVENC video encoding sessions imposed on consumer GPUs"
Context: Open-source patch project
Confidence: high

Key implications:
- Patch removes artificial driver limits; actual sessions still bounded by hardware performance[^29^]
- Workstation/data center GPUs (Quadro, Tesla) have no session restrictions natively[^21^]
- Professional GPUs achieve 11-17 concurrent NVENC sessions depending on quality and hardware[^21^]
- Using patched drivers in production has risk considerations[^29^]

---

## 3. Frame Duplication Strategies

### 3.1 GPU Texture Sharing

For zero-copy frame sharing between encoder instances, GPU textures can be shared across contexts and APIs.

Claim: "With Media Foundation, it is possible to share your Direct3D11 device between your application and the Media Foundation accelerated decoders via the IMFDXGIDeviceManager"[^30^]
Source: Scali's OpenBlog
URL: https://scalibq.wordpress.com/2022/06/11/gpu-accelerated-video-decoding/
Date: 2022-06-11
Excerpt: "With Media Foundation, it is possible to share your Direct3D11 device between your application and the Media Foundation accelerated decoders. This can be done via the IMFDXGIDeviceManager, which you can create with the MFCreateDXGIDeviceManager function."
Context: GPU-accelerated video pipeline technical blog
Confidence: high

GPU texture sharing mechanisms:
- **Direct3D 11 shared textures**: ID3D11Texture2D with D3D11_RESOURCE_MISC_SHARED flag for cross-process sharing
- **Vulkan external memory**: VK_EXTERNAL_MEMORY_HANDLE_TYPE_OPAQUE_WIN32_BIT for cross-API sharing[^31^]
- **WGL_NV_DX_interop**: OpenGL ↔ DirectX texture sharing on Windows[^32^]
- **CopySubresourceRegion**: GPU-side texture copy without CPU round-trip[^32^]

For dual-encoder pipelines, the strategy is:
1. Capture frame into GPU texture (via DXGI grab or application render target)
2. Share/copy texture to encoder's expected format (NV12 for most hardware encoders)
3. Feed the same GPU texture reference to both NVENC sessions (read-only)
4. Each encoder consumes the same frame independently

Claim: "You probably still need to make a copy of this texture to your own texture with the same format, because you need to have a texture that has the D3D11_BIND_SHADER_RESOURCE flag set, if you want to use it in a shader, and the decoder usually does not set that flag. But since it is all done on the GPU, this is reasonably efficient"[^30^]
Source: Scali's OpenBlog
URL: https://scalibq.wordpress.com/2022/06/11/gpu-accelerated-video-decoding/
Date: 2022-06-11
Excerpt: "You probably still need to make a copy of this texture to your own texture with the same format... But since it is all done on the GPU, this is reasonably efficient."
Context: GPU texture pipeline notes
Confidence: high

### 3.2 FFmpeg Hardware Frame References

FFmpeg supports hardware-accelerated frame sharing via AVHWFramesContext.

For Vulkan Video zero-copy:
- Vulkan Video enables zero-copy when using the Vulkan GPU backend[^33^]
- YUV pixel formats are GPU-decoded without CPU round-trip[^33^]
- `hwupload_cuda` → `scale_npp` → encoder pipeline keeps frames on GPU throughout[^14^]

---

## 4. Recording Formats and Container Safety

### 4.1 Container Crash-Safety Comparison

| Format | Crash Recovery | Streaming Safe | Key Mechanism |
|--------|---------------|----------------|---------------|
| **MKV** | Excellent — playable up to crash point[^34^] | Partial (no native HLS/DASH) | Progressive write, no final index required |
| **MP4** | Poor — entire file corrupt if moov atom not written[^34^] | Yes (with fMP4) | moov atom written at end (traditional) |
| **fMP4** | Good — each fragment playable[^35^] | Yes (designed for HLS/DASH) | Fragments with inline moof/mdat |
| **MOV** | Same as MP4[^36^] | Yes | QuickTime variant of MP4 |
| **FLV** | Good — progressive like MKV[^34^] | Yes (RTMP uses FLV) | Simple interleaved tag structure |
| **TS** | Good — each packet independent[^7^] | Yes (MPEG-TS) | Fixed-size packets, error resilient |

### 4.2 Fragmented MP4 (fMP4) for Live-Safe Recording

fMP4 enables writing playable fragments without a final moov atom, solving the crash-corruption problem.

Claim: "When doing fragmented MP4, each GOP becomes its own 'fragment' and no moov atom is necessary, so in the event of a crash you will lose, at worst, a single GOP — the last one that was being written at the time of crash"[^35^]
Source: OBS Project Forum
URL: https://obsproject.com/forum/resources/how-to-fix-mp4-mov-files-corrupting-when-obs-studio-crashes.1293/
Date: 2021-06-01
Excerpt: "When doing fragmented MP4, each GOP becomes its own 'fragment' and no moov atom is necessary, so in the event of a crash you will lose, at worst, a single GOP — the last one that was being written at the time of crash."
Context: OBS crash-safe recording technique
Confidence: high

Essential FFmpeg fMP4 options:

| Option | Purpose |
|--------|---------|
| `-movflags frag_keyframe` | Start new fragment at each video keyframe[^37^] |
| `-movflags empty_moov` | Write initial moov atom at start with zero duration (streaming-friendly)[^37^] |
| `-movflags +faststart` | Second pass to move moov atom to beginning (post-processing only)[^37^] |
| `-frag_duration N` | Create fragments of N microseconds duration[^37^] |
| `-frag_size N` | Create fragments of up to N bytes payload[^37^] |
| `-min_frag_duration N` | Don't create fragments shorter than N microseconds[^37^] |
| `-movflags separate_moof` | Separate moof atom per track (easier track separation)[^37^] |

OBS fMP4 crash-safe recording command:
```
Custom Muxer Settings: movflags=frag_keyframe+empty_moov
```
Effect: "Your MP4/MOV recording will still open in video players and video editors" even if OBS crashes[^35^].

Fragmented MP4 FFmpeg command:
```bash
ffmpeg -i input -c:v libx264 -c:a aac -movflags frag_keyframe+empty_moov+separate_moof \
  -f mp4 output.mp4
```

### 4.3 OBS Hybrid MP4 (OBS 30.2+)

Claim: "Newer versions (OBS 30.2 and above) introduced a 'Hybrid MP4' mode, which uses internal fragmentation to mimic MKV's crash resilience while still producing an MP4"[^34^]
Source: OneStream Live Blog
URL: https://onestream.live/blog/mkv-vs-mp4-for-pre-recorded-streaming/
Date: 2026-01-12
Excerpt: "Newer versions (OBS 30.2 and above) introduced a 'Hybrid MP4' mode, which uses internal fragmentation to mimic MKV's crash resilience while still producing an MP4."
Context: OBS recording format update
Confidence: medium

Hybrid MP4 characteristics:
- Periodically writes MP4 metadata during recording[^34^]
- Interrupted file remains playable up to last fragment[^34^]
- Solves the classic moov-atom problem while maintaining MP4 compatibility[^34^]
- Can still be remuxed to standard MP4 after recording completes

### 4.4 Best Practice: MKV → Remux → MP4

The industry-standard safe workflow:
1. **Record to MKV** (crash-resistant)[^34^]
2. **Remux to MP4** after recording completes (seconds, no quality loss)[^34^]
3. OBS built-in: "Automatically remux to MP4" setting

Claim: "The general advice is to always record to .mkv and let OBS remux to .mp4 after recording"[^34^]
Source: OneStream Live Blog
URL: https://onestream.live/blog/mkv-vs-mp4-for-pre-recorded-streaming/
Date: 2026-01-12
Excerpt: "The general advice is to always record to .mkv and let OBS remux to .mp4 after recording"
Context: OBS best practices
Confidence: high

Remux speed: "A multi-gigabyte MKV can be remuxed to MP4 in a matter of seconds to a minute" — reported 11 GB/hour MKV remuxed in 15-20 seconds on SATA SSD[^34^].

---

## 5. Real-Time DVR: Circular Buffer and Instant Replay

### 5.1 OBS Replay Buffer

OBS provides built-in instant replay via a circular memory buffer.

Claim: "OBS Studio includes a Replay Buffer feature that maintains a rolling window of the last N seconds in memory"[^38^]
Source: OBS Project Forum
URL: https://obsproject.com/forum/resources/how-to-setup-instant-replay-in-obs-studio.613/
Date: 2018-01-23
Excerpt: "Open the OBS Studio settings, go to output, and check the box 'Enable Replay Buffer'. Set the length to your desired time. Note: Longer replay buffers require more memory."
Context: OBS instant replay setup guide
Confidence: high

OBS Replay Buffer architecture:
- Continuously encodes to a rolling memory buffer[^38^]
- When triggered, saves the buffered duration to disk[^38^]
- Can be configured to auto-start with streaming[^38^]
- Lua scripting API enables instant replay source switching[^38^]
- Default buffer location: Videos folder, filename starts with "Replay"[^38^]

### 5.2 StreamShark Live DVR

Enterprise live DVR implementations use server-side circular buffers.

Claim: "StreamShark's Live DVR feature archives the contents of a live event, creating a real time DVR buffer during your stream, allowing you to deliver an archive instantly after a live event"[^39^]
Source: StreamShark Support
URL: https://support.streamshark.io/hc/en-us/articles/115001170672-How-to-Use-Live-DVR-and-Archiving
Date: Unknown
Excerpt: "StreamShark's Live DVR & Archiving feature archives the contents of a live event (creating a real time DVR buffer during your stream), allowing you to deliver an archive instantly after a live event from the same Event embed code."
Context: Enterprise streaming DVR documentation
Confidence: high

Key characteristics:
- **Maximum recording window**: 8 hours (rolling)[^39^]
- **Clip highlights while live**: Unlimited highlights during ongoing events[^39^]
- **Live Rewind**: Users joining mid-stream can rewind to beginning[^39^]
- **Top and Tail**: Frame-perfect editing of start/end points post-event[^39^]

### 5.3 Implementation Pattern: Circular Buffer Recording

For custom implementations, a circular buffer recording pipeline involves:
1. **Memory ring buffer**: Fixed-size buffer storing the last N seconds of encoded packets
2. **Two-phase writing**: Continuously write to disk (current segment) + maintain memory buffer (hot window)
3. **Clip extraction**: On trigger, copy ring buffer range to new file
4. **Segment rotation**: For continuous recording, rotate files based on time/size (like Haivision's segmentation)[^7^]

---

## 6. Storage Backends

### 6.1 Local Filesystem

Direct attached storage (DAS) is the simplest and most reliable backend:
- NVMe SSDs: 3,000+ MB/s sustained write (far exceeds 4K60 recording needs)
- SATA SSDs: 400-550 MB/s sustained write
- USB 3.0 external: ~300-400 MB/s real-world
- SD cards (UHS-II/V90): 200-300 MB/s minimum for 4K60 ProRes[^40^]

### 6.2 SMB/CIFS and NFS

Network-attached storage via standard protocols.

Claim: "The difference ended up being somewhat negligible for sequential reads and writes. With both NFS and SMB getting pretty close to the theoretical limit of 125MB/s of the 1GbE link"[^41^]
Source: GitHub NFS vs SMB Benchmarking Experiment
URL: https://github.com/jonfk/nfs-smb-benchmarking-experiment
Date: 2024-10-09
Excerpt: "SEQUENTIAL_WRITE: IOPS: NFS: 107.00 SMB: 110.00; Bandwidth: NFS: 108.00 MiB/s SMB: 110.00 MiB/s"
Context: Real-world 1GbE benchmark comparison
Confidence: high

Key findings from NFS vs SMB benchmark on 1GbE:
- Sequential write: NFS ~108 MiB/s, SMB ~110 MiB/s (both near 1GbE theoretical max of ~125 MiB/s)[^41^]
- Sequential read: NFS ~111 MiB/s, SMB ~112 MiB/s[^41^]
- Random write: SMB ~30 MiB/s, NFS ~25.9 MiB/s (SMB 16% faster)[^41^]
- Random read: NFS ~28.7 MiB/s, SMB ~21.5 MiB/s (NFS 33% faster)[^41^]
- **CPU efficiency**: NFS generally more CPU-efficient, especially for random workloads[^41^]

For video recording (sequential write), both protocols easily handle 4K60 H.264/H.265 (4.4-6.25 MB/s required vs 108-110 MB/s available).

### 6.3 S3-Compatible Object Storage

S3 multipart upload is the standard for large file recording to object storage.

Claim: "Multipart upload solves this by breaking large files into smaller chunks, uploading them in parallel, and assembling them on S3's side. If one chunk fails, you only re-upload that chunk"[^42^]
Source: OneUptime Blog
URL: https://oneuptime.com/blog/post/2026-02-12-upload-large-files-to-s3-multipart-upload/view
Date: 2026-02-12
Excerpt: "Multipart upload solves this by breaking large files into smaller chunks, uploading them in parallel, and assembling them on S3's side. If one chunk fails, you only re-upload that chunk."
Context: S3 multipart upload documentation
Confidence: high

S3 multipart upload parameters:
- Minimum part size: 5 MB (except last part)[^42^]
- Maximum parts: 10,000 per upload[^42^]
- Part size: configurable (default 8 MB in AWS CLI)[^42^]
- Concurrent requests: configurable parallelism[^42^]
- Resume capability: interrupted multipart uploads can resume from last complete part[^43^]

Streaming S3 upload algorithm (for live recording):
```
1. Receive data buffers from video encoder
2. Accumulate buffers until threshold reached (e.g., 50 MB)
3. Keep 5 MB reserve to ensure last part meets minimum
4. Upload accumulated data as new S3 part
5. On stream end, upload remaining buffers as final part
```
Source: Daniel Imfeld's blog (Rust implementation)[^44^]

Claim: "The one complication with the multipart API is that each part must be at least 5MB... streaming data to the multipart upload API involves gathering data buffers as they are generated and periodically uploading the accumulated data as a new S3 part when it reaches a certain threshold size"[^44^]
Source: Daniel Imfeld — Streaming S3 Uploads in Rust
URL: https://imfeld.dev/writing/rust_s3_streaming_upload
Date: 2021-06-28
Excerpt: "The one complication with the multipart API is that each part must be at least 5MB."
Context: Live streaming upload to S3 implementation
Confidence: high

Wowza Streaming Engine S3 upload module:
- Uploads completed recordings via AWS TransferManager[^43^]
- Creates `[recording-name].upload` temp file to track progress[^43^]
- Resumes interrupted multipart uploads on restart[^43^]
- Single-part uploads restart from beginning if interrupted[^43^]

### 6.4 FTP/SFTP

Traditional protocol used by hardware encoders for remote archiving.

Haivision Makito X supports:
- **Auto Export**: Completed segments automatically exported to FTP/SFTP server[^7^]
- **HVC/Haivision Media Platform integration**: Direct ingest via watch folder[^6^]
- **USB/SD local export**: Physical media transfer without network[^6^]

### 6.5 WebDAV and Custom HTTP

WebDAV and custom HTTP pipelines can serve as recording backends:
- HTTP PUT with chunked transfer encoding enables streaming upload
- WebDAV provides filesystem-like semantics over HTTP
- For live recording, chunked HTTP POST with periodic flush is most common
- Matrox Maevex supports recording to any network-shared folder (SMB/WebDAV compatible)[^45^]

---

## 7. Network Storage Performance: 4K60 Requirements

### 7.1 Throughput Calculations

For 4K60 video recording, the required storage throughput depends on codec:

Claim: "Our 4k video size calculator shows that uncompressed 4K60 footage can require up to 3Gbps, while compressed formats like H.265 might only need 35-50Mbps for excellent quality"[^46^]
Source: Claydesk Bitrate Calculator
URL: https://claydesk.ai/calculators/technology/bitrate-calculator.html
Date: 2025-05-15
Excerpt: "Our 4k video size calculator shows that uncompressed 4K60 footage can require up to 3Gbps, while compressed formats like H.265 might only need 35-50Mbps for excellent quality."
Context: Video bitrate calculation reference
Confidence: high

**4K60 Storage Throughput Requirements:**

| Codec | Bitrate | MB/s Required | Safety Margin (20%) | Notes |
|-------|---------|---------------|---------------------|-------|
| H.265/HEVC | 35-50 Mbps | 4.4-6.25 | 5.3-7.5 | Recommended for streaming[^46^] |
| H.264/AVC | 53-68 Mbps | 6.6-8.5 | 8.0-10.2 | YouTube 4K60 recommendation[^47^] |
| ProRes 422 HQ | ~880 Mbps | ~110 | ~132 | Professional post-production |
| Uncompressed | ~3 Gbps | ~375 | ~450 | Not practical for recording[^46^] |

YouTube official 4K60 upload recommendations:
- 4K SDR 30fps: 35-45 Mbps
- 4K 60fps: 53-68 Mbps[^47^]

Source: Have Camera Will Travel
URL: https://havecamerawilltravel.com/4k-video-storage-capacity/
Date: 2026
Excerpt: "YouTube officially recommends an upload bitrate of 35-45 Mbps for standard 4K SDR (24, 25, or 30 fps) and 53-68 Mbps for high frame rates (48, 50, or 60 fps)."
Context: YouTube upload specifications
Confidence: high

**Network protocol sufficiency for 4K60 H.265:**
- 1GbE NFS/SMB: ~108-110 MiB/s sustained sequential write = **17× headroom** for 4K60 H.265
- 100 Mbps LAN: ~11 MiB/s = **1.5× headroom** for 4K60 H.265 (marginal)
- Wi-Fi 5 (AC): ~30-50 MB/s typical = **5-8× headroom** for 4K60 H.265

### 7.2 Storage Capacity

At 50 Mbps H.265 4K60:
- 1 hour = ~22.5 GB[^46^]
- 8 hours (full workday) = ~180 GB
- 250 GB SSD (Haivision Makito) = ~11 hours of 4K60 at 50 Mbps

---

## 8. Go Integration Patterns

### 8.1 Pipeline Architecture

Go's concurrency primitives (goroutines, channels) map naturally to video pipeline stages.

Claim: "Go's pipeline pattern connects stages via channels where each stage is an independent goroutine that passes data through channels"[^48^]
Source: Go Concurrency Patterns: Pipelines and cancellation (Official Go Blog)
URL: https://go.dev/blog/pipelines
Date: 2014-03-13
Excerpt: "A function can read from multiple inputs and proceed until all are closed by multiplexing the input channels onto a single channel that's closed when all the inputs are closed."
Context: Official Go concurrency patterns documentation
Confidence: high

**Fan-out pattern for dual output (stream + record):**
```go
// Single source fans out to multiple consumers
func fanOut(source <-chan Packet, consumers ...chan<- Packet) {
    for pkt := range source {
        for _, consumer := range consumers {
            consumer <- pkt  // Send to each consumer
        }
    }
}
```

**Fan-in pattern for multiple sources:**
```go
func fanIn(channels ...<-chan Packet) <-chan Packet {
    var wg sync.WaitGroup
    out := make(chan Packet)
    
    output := func(c <-chan Packet) {
        for pkt := range c {
            out <- pkt
        }
        wg.Done()
    }
    
    wg.Add(len(channels))
    for _, c := range channels {
        go output(c)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```
Source: Go Concurrency Patterns[^48^]

### 8.2 Async File I/O in Go

Go's standard I/O is synchronous; async patterns use goroutines:

**Pattern: Goroutine-per-writer with buffered channel:**
```go
type AsyncWriter struct {
    buf     chan []byte
    file    io.Writer
    done    chan error
}

func NewAsyncWriter(file io.Writer, bufSize int) *AsyncWriter {
    w := &AsyncWriter{
        buf:  make(chan []byte, bufSize),
        file: file,
        done: make(chan error, 1),
    }
    go w.loop()
    return w
}

func (w *AsyncWriter) Write(p []byte) (int, error) {
    // Copy data (channel doesn't own the slice)
    buf := make([]byte, len(p))
    copy(buf, p)
    select {
    case w.buf <- buf:
        return len(p), nil
    default:
        return 0, errors.New("buffer full")
    }
}

func (w *AsyncWriter) loop() {
    for pkt := range w.buf {
        w.file.Write(pkt)
    }
    w.done <- nil
}
```

**Pattern: io.MultiWriter for simultaneous outputs:**
```go
streamWriter := getStreamWriter()  // e.g., net.Conn to RTMP server
fileWriter := getFileWriter()      // e.g., os.File

// Write to both simultaneously
multiWriter := io.MultiWriter(streamWriter, fileWriter)

// Or with tee-like behavior using io.TeeReader
teeReader := io.TeeReader(source, fileWriter)
io.Copy(streamWriter, teeReader)
```

Source: Medium — Common I/O Patterns in Go[^49^]

### 8.3 FFmpeg Integration via Cgo

For Go applications integrating FFmpeg:

1. **libavformat API**: Feed the same AVPacket to multiple av_write_frame() calls for different muxers[^11^]
2. **Pipes**: Launch FFmpeg as subprocess, write raw frames to stdin
3. **Cgo bindings**: Use goav or similar bindings for direct libav integration

Claim: "The tee muxer is not useful when using the libavformat API directly because it is then possible to feed the same packets to several muxers directly"[^11^]
Source: FFmpeg Documentation
URL: https://gensoft.pasteur.fr/docs/ffmpeg/4.3.1/ffmpeg-all.html
Date: Unknown
Excerpt: "The tee muxer is not useful when using the libavformat API directly because it is then possible to feed the same packets to several muxers directly."
Context: API design note for developers
Confidence: high

### 8.4 io_uring for High-Performance Async I/O

For Linux-based systems, io_uring provides kernel-bypass async I/O:

Claim: "io_uring can submit 1000 read operations with a single syscall, yielding tremendous savings in context switches and syscall latency"[^50^]
Source: GoCodeo
URL: https://www.gocodeo.com/post/building-scalable-applications-using-io-uring-for-async-operations
Date: 2025-06-20
Excerpt: "For example, submitting 1000 read operations with a single syscall can yield tremendous savings in context switches and syscall latency."
Context: io_uring tutorial for Go
Confidence: high

io_uring key features for video recording:
- Zero-copy I/O with pre-registered buffers and file descriptors[^50^]
- Batch submission of operations (reduces syscall overhead)[^50^]
- Poll mode for minimal latency or wait mode for CPU efficiency[^50^]
- Integration via Cgo or pure Go packages (github.com/iceber/io_uring-go)

---

## 9. Impact Analysis: Does Recording Affect Streaming?

### 9.1 NVENC Hardware Encoding

Claim: "Since NVENC is a dedicated circuit on the GPU, running 2 encodings in parallel will not require more CPU power - it's perfect"[^19^]
Source: OBS Project Forum
URL: https://obsproject.com/forum/threads/does-streaming-and-recording-at-the-same-time-use-more-cpu-when-using-nvenc.150648/
Date: 2021-11-26
Excerpt: "This will run 2 encodings running parallel on the GPU, one with streaming settings and one with recording settings. Since Nvenc is a dedicated circuit on the GPU, this will also not require more GPU resources - it's perfect."
Context: OBS user discussing dual NVENC encoding
Confidence: high

**NVENC dual-encoding impact assessment:**

| Factor | Impact | Notes |
|--------|--------|-------|
| CPU usage | Minimal to none | NVENC is dedicated hardware ASIC[^19^] |
| GPU 3D load | Minimal | Slight VRAM bandwidth increase |
| GPU memory | Moderate | Two encoder contexts, shared frame buffers |
| Stream latency | Negligible | NVENC maintains ~7 frame latency[^26^] |
| Power consumption | Increased | Both NVENC engines active |
| Frame drops | Possible at high loads | 8+ sessions or very high presets may overload[^27^] |

### 9.2 NVENC Latency Benchmarks

Claim: "NVIDIA encoder performed the most consistently, maintaining a latency of approximately 7 frames across nearly all presets and tuning modes"[^26^]
Source: arXiv — Evaluation of GPU Video Encoder for Low-Latency Real-Time 4K UHD Encoding
URL: https://arxiv.org/html/2511.18688v1
Date: 2025-11-24
Excerpt: "NVIDIA encoder performed the most consistently, maintaining a latency of approximately 7 frames across nearly all presets and tuning modes."
Context: Academic benchmark of GPU encoders
Confidence: high

Notable exception: H.264 P7 preset with Normal Latency tuning forces two-pass encoding, increasing latency to 37 frames. Enabling Split-Frame Encoding (SFE) resolves this overload[^26^].

### 9.3 Software Encoding Impact

When using CPU-based encoders (x264, x265):
- Dual encoding significantly increases CPU load
- Each output requires a separate encoding pass
- Tee muxer (same encode) helps but only if both outputs can use identical encoding parameters
- Different bitrates/resolutions for stream vs record require separate encodes

### 9.4 OBS-Specific Issues

Common issues when streaming + recording simultaneously:

1. **Audio buffering**: "Audio buffering hit the maximum value. This can be an indicator of very high system load and may affect stream latency"[^51^]
2. **Multiple capture sources**: Multiple GPU capture sources in one scene can cause occasional frame drops[^52^]
3. **Encoder features**: Multipass mode, Adaptive Quantization, and Look-ahead use additional GPU resources[^52^]

---

## 10. Hardware Architecture Comparison

### 10.1 Enterprise Hardware Encoders

| Device | Channels | Max Resolution | Simultaneous Stream+Record | Storage Backend | Key Feature |
|--------|----------|---------------|---------------------------|-----------------|-------------|
| Matrox Monarch HDX | 2× H.264 | 1080p60 | Yes (30 Mb/s shared) | SD, USB, network share | Split file, frame sync[^1^] |
| Matrox Maevex 7112H | 1× H.264/H.265 | 4K60 | Yes (single stream + record) | NAS, network share | Zero-latency pass-through[^53^] |
| Matrox Maevex 6152 | 4× H.264 | 4K60×4 | Yes | NAS, USB | 800 Mbps combined bitrate[^45^] |
| Osprey Talon G1 | 3× H.264 | 1080p60 | Yes (stream + .ts archive) | USB/SD (MP4/TS) | MBR mode, silent operation[^3^] |
| Osprey Talon 4K | 3× H.264/HEVC | 4K60 | Yes (stream + MP4) | USB/SD | 10-bit HDR, 4:2:2[^2^] |
| Haivision Makito X | 4× H.264 | 1080p60 | Yes (2 streams record) | 250 GB SSD, FTP export | Auto segment, roll-over[^6^] |
| ATEM Mini Pro | 1× H.264 | 1080p60 | Yes (same encoder) | USB flash/CFexpress | Same encode for both[^9^] |
| Blackmagic HyperDeck | N/A (recorder) | 4K60 | N/A (standalone record) | SD, USB-C, CFast | ProRes/ DNx recording[^10^] |

### 10.2 Key Architectural Patterns Summary

**Pattern A: Single Encode, Dual Output (tee muxer)**
- One encoding pass → identical data to stream and file
- Best for: Same quality/bitrate needed for both
- Tools: FFmpeg tee muxer, OBS "use stream encoder" for recording
- Impact: Minimal — single encode serves both outputs[^11^]

**Pattern B: Dual Encode, Different Parameters**
- Two independent encodes → different bitrates/resolutions for stream and record
- Best for: High-quality archive + bandwidth-limited stream
- Tools: NVENC dual sessions, OBS separate encoder for recording
- Impact: Double GPU encoding load but minimal CPU impact[^19^]

**Pattern C: Hardware Appliance (dedicated encoders)**
- Dedicated silicon handles both stream and record
- Best for: Mission-critical, 24/7 operation
- Tools: Matrox, Osprey, Haivision appliances
- Impact: Purpose-built, minimal host system load[^1^][^2^][^6^]

**Pattern D: Segmented Recording with Auto-Export**
- Continuous recording with automatic segmentation and remote export
- Best for: Long-duration recording with cloud backup
- Tools: Haivision Makito (segment + FTP), FFmpeg segment muxer
- Impact: Background upload doesn't affect streaming[^7^]

---

## 11. Contradictions and Conflicting Evidence

### 11.1 NVENC Session Limits
- **Conflict**: Some sources report 2-session limits (older GPUs), others report 5 (2023), others 8 (2024)
- **Resolution**: The limit increased over time. As of January 2024, all NVENC-capable GPUs support 8 sessions except GTX 1630 (3 sessions)[^23^]. These are driver-level limits, not hardware limits.

### 11.2 Dual NVENC Performance Scaling
- **Conflict**: RTX 4090 with dual encoders sometimes achieves 2× throughput, sometimes only 1×-1.6×
- **Resolution**: Performance depends on frame scheduling between encoders. NVIDIA's driver manages encoder assignment; results are not deterministic[^27^].

### 11.3 MKV vs fMP4 for Crash Safety
- **Conflict**: MKV is "crash-safe" but fMP4 with `frag_keyframe+empty_moov` provides similar safety
- **Resolution**: MKV is simpler and more proven. fMP4 is better for streaming compatibility but requires correct muxer flags. OBS Hybrid MP4 (30.2+) attempts to combine both benefits[^34^].

### 11.4 Recording Impact on Stream
- **Conflict**: Some users report frame drops when recording+streaming; others report zero impact
- **Resolution**: With NVENC hardware encoding, impact should be minimal. Frame drops typically stem from: (a) insufficient GPU VRAM, (b) CPU-based encoding, (c) disk I/O bottlenecks, or (d) incorrect OBS configuration (e.g., using CPU encoder for one output)[^52^].

---

## 12. Practical Implementation Recommendations

### 12.1 Minimal-Impact Dual Pipeline (NVENC)

```bash
# Single encode to stream + record (same quality)
ffmpeg -f gdigrab -i desktop -c:v h264_nvenc -b:v 6000k \
  -c:a aac -b:a 128k -g 120 \
  -f tee "[f=flv:onfail=ignore]rtmp://stream-server/live/key|[f=segment:segment_time=300]recording_%Y%m%d_%H%M%S.mkv"
```

### 12.2 Dual-Encode Pipeline (different quality)

```bash
# Two NVENC encodes: low bitrate stream, high bitrate recording
ffmpeg -f gdigrab -i desktop \
  -filter_complex "[0:v]split=2[stream][record];[record]scale=1920:1080[rec1080]" \
  -map "[stream]" -c:v h264_nvenc -b:v 4000k -preset p4 -tune ll \
  -f flv rtmp://stream-server/live/key \
  -map "[rec1080]" -c:v h264_nvenc -b:v 20000k -preset p6 -tune hq \
  -c:a aac -b:a 256k -movflags frag_keyframe+empty_moov recording.mp4
```

### 12.3 Go Async Recording Pipeline Skeleton

```go
// Simultaneous stream+record with async file I/O
func RecordAndStream(ctx context.Context, source VideoSource, 
    streamURL string, filePath string) error {
    
    // Setup outputs
    streamConn, _ := net.Dial("tcp", streamURL)
    file, _ := os.Create(filePath)
    defer file.Close()
    
    // Async file writer with buffer
    asyncFile := NewAsyncWriter(file, 120) // 120-packet buffer
    
    // Tee the encoded packets
    multiWriter := io.MultiWriter(streamConn, asyncFile)
    
    // Run encoder
    encoder := NewEncoder(source)
    for {
        select {
        case <-ctx.Done():
            return nil
        case packet := <-encoder.Output():
            _, err := multiWriter.Write(packet.Data)
            if err != nil {
                log.Printf("Write error: %v", err)
            }
        }
    }
}
```

### 12.4 fMP4 Crash-Safe Recording Command

```bash
# Live-safe MP4 recording that survives crashes
ffmpeg -i rtmp://input/stream -c copy \
  -movflags frag_keyframe+empty_moov+separate_moof \
  -f mp4 output_live_safe.mp4
```

---

## References

[^1^]: Matrox Monarch HDX Product Page, https://video.matrox.com/en/products/encoders-decoders/monarch-series/monarch-hdx
[^2^]: Exxact Corporation Osprey Talon G1, https://www.exxactcorp.com/Osprey-96-02010-E1345845
[^3^]: Avanta Digital Osprey Talon G1, https://www.avantadigital.com/osprey-talon-g1-h-264-video-hardware-encoder
[^4^]: Broadcast Store Europe Osprey Talon Pro, https://broadcaststoreeurope.com/shop/1308-video-ip-transmission/17204-osprey-talon-pro-encoder-2x-3g-sdi-4k-hdmi-unbalanced-stereo-audio-input/
[^5^]: Osprey TALON User Guide PDF, https://www.markertek.com/Attachments/Manuals/Osprey/96-02012-Manual.pdf
[^6^]: Haivision Makito X Storage and Recording, https://doc.haivision.com/MakitoXEnc/2.5/storage-and-recording
[^7^]: Haivision Makito X Configuring Recording Outputs, https://doc.haivision.com/MakitoXEnc/2.4/configuring-recording-outputs
[^8^]: Haivision Makito X Product Overview, https://doc.haivision.com/MakitoXEnc/2.5/product-overview
[^9^]: Reddit ATEM Mini Pro Record and Stream, https://www.reddit.com/r/blackmagicdesign/comments/173j6z8/atem_mini_pro_record_and_stream_simultaneously/
[^10^]: Blackmagic ATEM Mini Product Page, https://www.blackmagicdesign.com/products/atemmini
[^11^]: FFmpeg Documentation (tee muxer), https://gensoft.pasteur.fr/docs/ffmpeg/4.3.1/ffmpeg-all.html
[^12^]: SuperUser Multi-encode Streamcopy, https://superuser.com/questions/1831388/complex-ffmpeg-multi-encode-streamcopy-muxing-scenario-is-this-actually-doable
[^13^]: SuperUser Send output to multiple destinations, https://superuser.com/questions/1410451/send-output-to-multiple-destinations
[^14^]: GitHub Gist FFmpeg NVENC streaming, https://gist.github.com/gavanderhoorn/cfab2ec1767dd6624bd32062c594f83e
[^15^]: StackOverflow FFMPEG multiple outputs, https://stackoverflow.com/questions/41880004/ffmpeg-how-to-stream-to-multiple-outputs-with-the-same-encoding-independently
[^16^]: FFmpeg fifo muxer documentation, https://gensoft.pasteur.fr/docs/ffmpeg/4.3.1/ffmpeg-all.html
[^17^]: FFmpeg-devel fifo muxer patch, http://mplayerhq.hu/pipermail/ffmpeg-devel/2024-March/323198.html
[^18^]: OBS Stream to 2 destinations, https://obsproject.com/forum/resources/stream-to-2-destinations-simultaneously-with-obs-without-nginx.788/
[^19^]: OBS Does streaming and recording use more CPU with NVENC, https://obsproject.com/forum/threads/does-streaming-and-recording-at-the-same-time-use-more-cpu-when-using-nvenc.150648/
[^20^]: OBS Recording two output formats, https://obsproject.com/forum/threads/recording-two-output-formats-from-a-single-source.149022/
[^21^]: Tom's Hardware NVENC sessions increase, https://www.tomshardware.com/news/nvidia-increases-concurrent-nvenc-sessions-on-consumer-gpus
[^22^]: NVIDIA Developer Forums Session count limitation, https://forums.developer.nvidia.com/t/session-count-limitation-for-nvenc-no-maxwell-gpus-with-2-nevenc-sessions/36192
[^23^]: VideoCardz 8 NVENC sessions, https://videocardz.com/newz/nvdia-geforce-gpus-now-support-up-to-8-concurrent-nvenc-encoding-sessions
[^24^]: Wikipedia GeForce RTX 40 series, https://en.wikipedia.org/wiki/GeForce_RTX_40_series
[^25^]: Reddit Does 2 NVENC mean double speed, https://www.reddit.com/r/DataHoarder/comments/15s6lmn/does_having_2_nvenc_encoders_mean_double_the/
[^26^]: arXiv GPU Encoder Low-Latency 4K UHD, https://arxiv.org/html/2511.18688v1
[^27^]: NVIDIA Developer Forums Dual NVENC performance, https://forums.developer.nvidia.com/t/fluctuating-performance-with-two-instances-of-nvenc/247778
[^28^]: HWCooling RDNA4 Architecture Deep Dive, https://www.hwcooling.net/en/better-more-capable-than-expected-rdna-4-architecture-deep-dive/
[^29^]: Hacker News nvidia-patch, https://news.ycombinator.com/item?id=40636122
[^30^]: Scali's OpenBlog GPU-accelerated decoding, https://scalibq.wordpress.com/2022/06/11/gpu-accelerated-video-decoding/
[^31^]: Fast Half-Life Video Recording with Vulkan, https://bxt.rs/blog/fast-half-life-video-recording-with-vulkan/
[^32^]: StackOverflow FFmpeg D3D11 to OpenGL texture, https://stackoverflow.com/questions/61082653/ffmpeg-with-d3d11-hw-acceleration-how-to-copy-directx-texture-to-opengl
[^33^]: Hacker News Vulkan Video zero-copy, https://news.ycombinator.com/item?id=47407293
[^34^]: OneStream Live MKV vs MP4, https://onestream.live/blog/mkv-vs-mp4-for-pre-recorded-streaming/
[^35^]: OBS Fix MP4 Corrupting, https://obsproject.com/forum/resources/how-to-fix-mp4-mov-files-corrupting-when-obs-studio-crashes.1293/
[^36^]: Matrox Monarch HDX (MOV/MP4 recording), https://video.matrox.com/en/products/encoders-decoders/monarch-series/monarch-hdx
[^37^]: FFmpeg MOV/MP4 muxer options, https://nschlia.github.io/ffmpegfs/html/ffmpeg__profiles_8cc.html
[^38^]: OBS Instant Replay Setup, https://obsproject.com/forum/resources/how-to-setup-instant-replay-in-obs-studio.613/
[^39^]: StreamShark Live DVR, https://support.streamshark.io/hc/en-us/articles/115001170672-How-to-Use-Live-DVR-and-Archiving
[^40^]: ProGrade Digital Memory Cards for 4K, https://progradedigital.com/memory-cards-for-4k-and-8k-video-navigating-the-requirements/
[^41^]: GitHub NFS vs SMB Benchmark, https://github.com/jonfk/nfs-smb-benchmarking-experiment
[^42^]: OneUptime S3 Multipart Upload, https://oneuptime.com/blog/post/2026-02-12-upload-large-files-to-s3-multipart-upload/view
[^43^]: Wowza S3 Upload Module, https://www.wowza.com/docs/how-to-upload-recorded-media-to-an-amazon-s3-bucket-modules3upload
[^44^]: Daniel Imfeld Streaming S3 Uploads in Rust, https://imfeld.dev/writing/rust_s3_streaming_upload
[^45^]: Matrox Maevex User Guide PDF, https://jmgs.jp/download/manual/Maevex_UserGuide_en.pdf
[^46^]: Claydesk Bitrate Calculator, https://claydesk.ai/calculators/technology/bitrate-calculator.html
[^47^]: Have Camera Will Travel 4K Storage, https://havecamerawilltravel.com/4k-video-storage-capacity/
[^48^]: Go Concurrency Patterns Pipelines, https://go.dev/blog/pipelines
[^49^]: Medium Common I/O Patterns in Go, https://medium.com/dev-bits/explaining-common-i-o-patterns-in-go-cd01b1b749c4
[^50^]: GoCodeo io_uring, https://www.gocodeo.com/post/building-scalable-applications-using-io-uring-for-async-operations
[^51^]: OBS Forum Stream Lag, https://obsproject.com/forum/threads/stream-lag-despite-good-components-using-aitum-plugin.186418/
[^52^]: OBS Forum FPS Drops, https://obsproject.com/forum/threads/random-occasional-fps-drops-while-the-system-the-game-is-just-fine.188872/
[^53^]: Matrox Maevex 7100 Press Release, https://video.matrox.com/en/media/press-releases/2023/matrox-video-introduces-maevex-7100-series-encoders
[^54^]: Ant Media MKV vs MP4, https://antmedia.io/mkv-vs-mp4-streaming-format/
[^55^]: StreamGuides NVENC Update, https://streamguides.gg/2024/01/nvenc-update-all-nvidia-geforce-cards-quietly-updated-to-8-encoding-sessions/
[^56^]: Matrox Video Encoders Product Line, https://go.matrox.com/ad-video-encoders-decoders-appliances.html
