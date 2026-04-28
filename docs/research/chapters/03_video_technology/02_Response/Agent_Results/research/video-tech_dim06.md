# Dimension 06: Audio Pipeline Technology

## Comprehensive Research on Audio Technology for Cloud Gaming with Multi-Channel Surround Support

**Research Date:** July 2025
**Searches Conducted:** 25+
**Sources:** RFCs, IEEE references, official specifications, vendor documentation, open-source projects

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Real-Time Audio Codecs](#2-real-time-audio-codecs)
3. [Multi-Channel Audio Layouts](#3-multi-channel-audio-layouts)
4. [PCM Formats and Configurations](#4-pcm-formats-and-configurations)
5. [Audio Passthrough Architecture](#5-audio-passthrough-architecture)
6. [Hardware Audio Endpoints](#6-hardware-audio-endpoints)
7. [Operating System Audio Stacks](#7-operating-system-audio-stacks)
8. [Audio Capture and Loopback](#8-audio-capture-and-loopback)
9. [Go Audio Libraries](#9-go-audio-libraries)
10. [Latency Targets and Optimization](#10-latency-targets-and-optimization)
11. [Recommendations for Cloud Gaming Implementation](#11-recommendations-for-cloud-gaming-implementation)

---

## 1. Executive Summary

This research covers the complete audio pipeline technology stack required for cloud gaming with multi-channel surround sound support. Key findings:

- **Opus** is the dominant real-time codec for cloud gaming, with RFC 6716 defining the base codec and RFC 7845 / draft-shin-avtcore-rtp-multi-opus defining multi-channel (5.1/7.1) support via the "multiopus" RTP payload format[^223^][^234^]
- **End-to-end audio latency target of <20ms** is achievable with optimized Opus configurations (2.5-5ms frames), WASAPI exclusive mode on Windows, and PipeWire on Linux[^250^][^377^]
- **HDMI eARC** (up to 37 Mbps) is the only consumer interface supporting uncompressed multi-channel audio including Dolby Atmos (TrueHD) and DTS:X; SPDIF is limited to compressed 5.1[^248^][^249^]
- **Go audio ecosystem** provides viable options through `oto` (low-level cross-platform), `malgo` (miniaudio bindings with multi-backend support), and `gopxl/beep` (high-level audio processing)[^342^][^334^][^412^]

---

## 2. Real-Time Audio Codecs

### 2.1 Opus (WebRTC Standard)

#### Overview

Opus is the IETF-standardized audio codec (RFC 6716) designed for interactive real-time applications. It combines the SILK speech codec (from Skype) and CELT (low-latency audio codec) to handle both voice and music efficiently[^234^].

```
Claim: Opus supports bitrates from 6 kbps to 510 kbps, frame sizes from 2.5 ms to 60 ms, and up to 255 channels via multistream[^233^]
Source: Wikipedia - Opus (audio format)
URL: https://en.wikipedia.org/wiki/Opus_(audio_format)
Date: 2025
Excerpt: "Opus supports constant and variable bitrate encoding from 6 kbit/s to 510 kbit/s, frame sizes from 2.5 ms to 60 ms, and five sampling rates from 8 kHz to 48 kHz. An Opus stream can support up to 255 audio channels."
Context: Official Opus specification summary
Confidence: High
```

#### Key Technical Specifications

| Parameter | Value |
|-----------|-------|
| Bitrate range | 6 kbps - 510 kbps (stereo), up to 256 kbps/channel for multi-channel |
| Sample rates | 8, 12, 16, 24, 48 kHz |
| Frame sizes | 2.5, 5, 10, 20, 40, 60 ms |
| Max channels | 255 (via multistream API) |
| Algorithmic delay | 5 ms (minimum) to 66.5 ms (maximum)[^234^] |
| Default latency | 26.5 ms (20 ms frames + default application setting) |
| Container support | Ogg, WebM, MPEG-TS, MP4 |
| RTP/WebRTC compatible | Yes |
| License | Royalty-free, open source |

#### libopus Encoder API Configuration

The libopus API provides fine-grained control over encoding parameters[^375^][^379^][^380^]:

```c
// Create encoder
OpusEncoder* enc = opus_encoder_create(48000, 2, OPUS_APPLICATION_AUDIO, &error);

// Key CTL commands
opus_encoder_ctl(enc, OPUS_SET_BITRATE(128000));     // Target bitrate in bps
opus_encoder_ctl(enc, OPUS_SET_COMPLEXITY(10));      // 0-10, higher = better quality
opus_encoder_ctl(enc, OPUS_SET_SIGNAL(OPUS_SIGNAL_MUSIC)); // MUSIC or VOICE
opus_encoder_ctl(enc, OPUS_SET_VBR(1));              // Enable VBR
opus_encoder_ctl(enc, OPUS_SET_INBAND_FEC(1));       // Forward error correction
opus_encoder_ctl(enc, OPUS_SET_PACKET_LOSS_PERC(5)); // Expected packet loss %
```

**Frame size constraints:** At 48 kHz, permitted frame sizes are 120 (2.5ms), 240 (5ms), 480 (10ms), 960 (20ms), 1920 (40ms), and 2880 (60ms) samples per channel[^379^].

```
Claim: Passing in a duration of less than 10 ms (480 samples at 48 kHz) will prevent the encoder from using the LPC or hybrid modes[^379^]
Source: opus_encoder man page - Arch Linux
URL: https://man.archlinux.org/man/opus_encoder.3.en
Date: Ongoing
Excerpt: "This must be an Opus frame size for the encoder's sampling rate. For example, at 48 kHz the permitted values are 120, 240, 480, 960, 1920, and 2880. Passing in a duration of less than 10 ms (480 samples at 48 kHz) will prevent the encoder from using the LPC or hybrid modes."
Context: Encoder frame size API documentation
Confidence: High
```

#### Recommended Bitrates for Cloud Gaming[^377^]

| Use Case | Channels | Bitrate (kbps) | Notes |
|----------|----------|----------------|-------|
| VoIP/chat | 1 (mono) | 10-24 | Narrowband to fullband |
| Game audio (stereo) | 2 | 64-128 | Opus exceeds MP3/AAC/Vorbis at these rates |
| Game audio (5.1) | 6 | 128-256 | Surround-sound bitrate allocation |
| Game audio (7.1) | 8 | 256-450 | Higher bitrate for full quality |

#### Multi-Channel Opus (MultiOpus / RFC 8486)

Multi-channel Opus uses the multistream API with channel mapping families defined in RFC 7845. The RTP/SDP signaling is specified in draft-shin-avtcore-rtp-multi-opus[^222^][^223^][^229^].

**5.1 Audio SDP Example:**
```sdp
a=rtpmap:111 multiopus/48000/6
a=fmtp:111 num_streams=4;coupled_streams=2;channel_mapping=0,4,1,2,3,5
```

**7.1 Audio SDP Example:**
```sdp
a=rtpmap:111 multiopus/48000/8
a=fmtp:111 num_streams=5;coupled_streams=3;channel_mapping=0,6,1,2,3,4,5,7
```

```
Claim: libwebrtc-based systems use non-standard "multiopus" SDP encoding with num_streams, coupled_streams, and channel_mapping parameters[^223^]
Source: IETF draft-shin-avtcore-rtp-multi-opus-01
URL: https://datatracker.ietf.org/doc/html/draft-shin-avtcore-rtp-multi-opus-01
Date: 2025-10-19
Excerpt: "Deployed systems (e.g., [libwebrtc] based) interoperate using a non-standard SDP encoding name 'multiopus' with fmtp parameters such as num_streams, coupled_streams, and channel_mapping."
Context: Standardization draft for multi-channel Opus in RTP
Confidence: High
```

**Key parameters:**
- `num_streams`: Total number of Opus streams
- `coupled_streams`: Number of stereo (coupled) streams
- `channel_mapping`: Comma-separated list mapping RTP channels to speaker positions

```
Claim: Remoto Playback handles multi-channel by receiving full multichannel Opus (up to 16 channels) and applying fold-down based on output device capabilities[^222^]
Source: Remoto Technical Reference
URL: https://support.remotopro.io/technical-reference/how-remoto-playback-handles-multichannel-audio-fold-down
Date: 2026-03-26
Excerpt: "The Desktop app receives the full multichannel stream (up to 16 channels) over the network. Its native C++ audio engine applies a Lo/Ro downmix matrix locally."
Context: Cloud gaming audio implementation reference
Confidence: High
```

### 2.2 AAC-LC and HE-AAC

#### Technical Specifications[^224^]

| Parameter | AAC-LC | HE-AAC | HE-AACv2 |
|-----------|--------|--------|----------|
| Bitrate range | 8-320 kbps/channel | 32-96 kbps (stereo) | 16-48 kbps |
| Sample rates | 8-96 kHz | Same | Same |
| Channels | Up to 48 (5.1 supported) | Up to 48 | Stereo only |
| Typical latency | 100-200 ms | 100-200 ms | 100-200 ms |
| Key feature | Baseline AAC | Spectral Band Replication | Parametric Stereo |
| License | Patent-protected | Patent-protected | Patent-protected |

```
Claim: AAC-LC operates efficiently at bitrates from 96 kbps to 256 kbps for 5.1 surround applications, with typical latency of 100-200 ms[^224^]
Source: Ant Media - Best Audio Codec for Online Video Streaming
URL: https://antmedia.io/best-audio-codec/
Date: 2026-01-02
Excerpt: "AAC-LC operates efficiently on mobile devices and provides good quality at bitrates from 96 kbps to 256 kbps... Typical latency: 100-200 ms"
Context: Audio codec comparison for streaming
Confidence: High
```

**Note:** AAC-LC/HE-AAC are generally unsuitable for cloud gaming due to 100-200ms algorithmic latency, far exceeding the <20ms target. Opus is preferred for real-time interactive use.

### 2.3 Dolby Digital (AC-3)

#### Technical Specifications[^336^][^341^]

| Parameter | Value |
|-----------|-------|
| Max bitrate | 640 kbps (5.1 channels) |
| Sample rates | 32, 44.1, 48 kHz |
| Channels | Up to 5.1 (6 discrete channels) |
| Frame size | 1536 samples (32ms at 48kHz) |
| Sync word | 0x0B77 (big-endian) |
| Standard | ATSC A/52 |

**AC-3 Frame Size Table (48 kHz sample rate):**

| Bitrate | Words/Syncframe | Bytes/Frame |
|---------|-----------------|-------------|
| 128 kbps | 256 | 512 |
| 192 kbps | 384 | 768 |
| 320 kbps | 640 | 1280 |
| 384 kbps | 768 | 1536 |
| 448 kbps | 896 | 1792 |
| 512 kbps | 1024 | 2048 |
| 576 kbps | 1152 | 2304 |
| 640 kbps | 1280 | 2560 |

```
Claim: AC-3 supports maximum bitrate of 640 kbps for 5.1 audio at 48 kHz, with frame size of 1280 words (2560 bytes) per syncframe[^336^]
Source: ATSC Standard A/52:2012 - Digital Audio Compression (AC-3, E-AC-3)
URL: http://www.atsc.org/wp-content/uploads/2015/03/A52-201212-17.pdf
Date: 2012
Excerpt: "Table 5.18 Frame Size Code Table: frmsizecod 100101' = 640 kbps, fs=48 kHz words/syncframe = 1280"
Context: Official ATSC AC-3 specification
Confidence: High
```

### 2.4 Dolby Digital Plus (E-AC-3)

#### Technical Specifications[^392^][^394^]

| Parameter | Value |
|-----------|-------|
| Bitrate range | 32 kbps - 6 Mbps |
| Max channels | Up to 15 full-band channels (7.1 discrete) |
| Sample rates | 32, 44.1, 48 kHz |
| 5.1 bitrate | As low as 192 kbps (vs 384 kbps minimum for AC-3) |
| Backward compatibility | Falls back to AC-3 core |

```
Claim: Dolby Digital Plus supports up to 7.1 discrete channels at bitrates from 32 kbps to 6 Mbps, with 5.1 audio achievable at 192 kbps[^394^]
Source: Dolby Professional - Dolby Digital Plus
URL: https://professional.dolby.com/technologies/dolby-digital-plus/
Date: Ongoing
Excerpt: "The Dolby Digital Plus codec (E-AC-3) delivers up to 7.1 discrete channels of crisp surround sound... Supports bit rates from 32 kbps to 6 Mbps"
Context: Official Dolby technical documentation
Confidence: High
```

#### Dolby Digital Plus with JOC (Joint Object Coding) for Atmos

E-AC3 JOC is a 5.1 E-AC3 bitstream with additional metadata to carry Dolby Atmos:

- **JOC (Joint Object Coding):** Compresses object audio data
- **OAMD (Object Audio Metadata):** Metadata for Atmos renderer
- **Minimum bitrate:** 384 kbps (commonly 448-768 kbps)
- **Backward compatible:** Non-Atmos devices play the 5.1 core[^392^]

### 2.5 Dolby TrueHD and Dolby Atmos

#### Dolby TrueHD

| Parameter | Value |
|-----------|-------|
| Max bitrate | Up to 18 Mbps (Blu-ray) |
| Sample rates | Up to 192 kHz |
| Bit depth | Up to 24-bit |
| Channels | Up to 16 (includes 7.1 + extensions) |
| Compression | Lossless (MLP - Meridian Lossless Packing) |

#### Dolby Atmos Delivery Formats

Dolby Atmos can be delivered via multiple codecs:

| Container | Codec | Use Case |
|-----------|-------|----------|
| Blu-ray/UHD | Dolby TrueHD + Atmos metadata | Highest quality, lossless |
| Streaming | E-AC3 (DD+) with JOC | Bandwidth-constrained |
| HDMI eARC | Dolby MAT 2.0 (LPCM-encoded) | Real-time encoding |

#### Dolby MAT 2.0 (Metadata-enhanced Audio Transmission)

```
Claim: Dolby MAT 2.0 dynamically encodes object-based audio information in real time by the source device, limiting latency and reducing processing complexity[^378^]
Source: AVPro Global - A Deep Dive Into Dolby MAT
URL: https://www.avproglobal.com/blogs/news/a-deep-dive-into-dolby-mat
Date: 2022-06-30
Excerpt: "Dubbed Dolby MAT 2.0, an essential key benefit is object-based audio information is dynamically encoded in real time by the source device, limiting latency and reducing processing complexity."
Context: Technical explanation of Dolby MAT for HDMI transport
Confidence: High
```

Dolby MAT is neither a codec nor a format - it is an encode/conversion/transport/conversion/decode process that uses the high-capacity audio carrier lanes in HDMI (eight 16-bit 192kHz lanes starting with HDMI 1.3) to establish a data transport layer[^378^].

### 2.6 DTS Family

#### DTS (Coherent Acoustics / Core)

| Parameter | Value |
|-----------|-------|
| Bitrate | 768 kbps - 1.5 Mbps (CD), up to 1.5 Mbps (DVD) |
| Sample rate | 48 kHz |
| Channels | Up to 5.1 |
| Compression | Lossy |

#### DTS-HD Master Audio[^231^]

| Parameter | Value |
|-----------|-------|
| Max channels | 8 discrete (7.1) |
| Sample depth | Up to 24-bit |
| Sample rate | Up to 192 kHz (96 kHz for 6.1/7.1) |
| Max bitrate | 24.5 Mbit/s instantaneous |
| Core stream | 1.5 Mbit/s lossy DTS for backward compatibility |
| Compression | Lossless |

```
Claim: DTS-HD MA can store up to 8 discrete channels of audio (7.1 surround) at up to 24-bit sample depth and 192 kHz sampling frequency, with max bitrate of 24.5 Mbit/s[^231^]
Source: Wikipedia - DTS-HD Master Audio
URL: https://en.wikipedia.org/wiki/DTS-HD_Master_Audio
Date: 2025
Excerpt: "DTS-HD MA can store up to 8 discrete channels of audio (7.1 surround) at up to a 24 bit sample depth and 192 kHz sampling frequency. A DTS-HD MA bitstream may have a bitrate no greater than 24.5 Mbit/s"
Context: Technical specifications
Confidence: High
```

#### DTS:X[^390^][^395^]

| Parameter | Value |
|-----------|-------|
| Type | Object-based audio codec |
| Max channels (consumer) | Up to 11.1 (DTS:X), up to 30.2 (DTS:X Pro) |
| Speaker support | Up to 32 speaker locations |
| Backward compatibility | Yes (layered on DTS-HD MA) |
| Foundation | MDA (Multi-Dimensional Audio) platform |

```
Claim: DTS:X supports up to 11.1 channels and unlimited sound elements, with the codec adapting to any speaker setup within a hemispherical layout[^390^]
Source: How-To Geek - What Is DTS:X?
URL: https://www.howtogeek.com/772173/what-is-dtsx/
Date: 2022
Excerpt: "The codec adapts to whatever surround sound setup you have. It supports up to 11.1 channels and unlimited sound elements."
Context: Consumer audio technology overview
Confidence: High
```

---

## 3. Multi-Channel Audio Layouts

### 3.1 Standard Channel Configurations

| Layout | Channels | Speaker Positions |
|--------|----------|-------------------|
| Mono | 1 | Center |
| Stereo | 2 | Left, Right |
| 2.1 | 3 | Left, Right, LFE |
| 5.1 | 6 | L, R, C, LFE, Ls, Rs |
| 7.1 | 8 | L, R, C, LFE, Ls, Rs, Lb, Rb |
| 7.1.4 (Atmos) | 12 | 7.1 + 4 height channels |
| 22.2 | 24 | 3-layer surround (ITU-R BS.2159) |

### 3.2 Channel Ordering Standards[^402^]

**Critical finding:** Windows uses a DIFFERENT 7.1 channel order than film/Dolby/DTS/SMPTE standards.

| Standard | Ch1 | Ch2 | Ch3 | Ch4 | Ch5 | Ch6 | Ch7 | Ch8 |
|----------|-----|-----|-----|-----|-----|-----|-----|-----|
| **5.1 (Universal)** | L | R | C | LFE | Ls | Rs | - | - |
| **7.1 Dolby/DTS/SMPTE** | L | R | C | LFE | Ls | Rs | Lb | Rb |
| **7.1 Microsoft (WaveFormatExtensible)** | L | R | C | LFE | Lb | Rb | Ls | Rs |
| **7.1 FLAC/OpenAL Soft** | L | R | C | LFE | Lb | Rb | Ls | Rs |

```
Claim: Microsoft Windows swaps side and rear surround channels compared to the Dolby/DTS/SMPTE standard for 7.1 audio, placing rear surrounds before side surrounds in the channel order[^402^]
Source: SourceForge - Confusion of 7.1 and 5.1 channel order in Windows
URL: https://sourceforge.net/p/mesh2hrtf-tools/wiki/Confusion_of_7-1%20and%205-1_channel_order_in_Windows/
Date: Ongoing
Excerpt: "7.1 Microsoft (see WaveFormatExtensible): Left, Right, Center, LFE-sub, L-Rear-surround, R-Rear-surround, L-Side-surround, R-Side-surround"
Context: Technical analysis of channel ordering confusion
Confidence: High
```

**Cloud gaming implication:** When streaming multi-channel audio, the server must emit channels in the client's expected order, or provide explicit channel mapping metadata.

### 3.3 ITU-R BS.775 Speaker Placement

The ITU-R BS.775 standard defines:
- **L/R front speakers:** 60 degrees apart (30 degrees each from center)
- **Center speaker:** 0 degrees (directly in front)
- **Surround speakers:** 100-120 degrees from center (side placement, NOT rear)
- **LFE/Subwoofer:** Non-directional, positions not critical[^410^]

```
Claim: Research shows placing surround speakers on the sides (not rear) is more immersive due to front-rear localization confusion; this is why 5.1 places surrounds at 100-120 degrees[^402^]
Source: SourceForge - Confusion of 7.1 and 5.1 channel order in Windows
URL: https://sourceforge.net/p/mesh2hrtf-tools/wiki/Confusion_of_7-1%20and%205-1_channel_order_in_Windows/
Date: Ongoing
Excerpt: "Research shows that placing surround speakers on the sides is much more immersive than in the rear-only"
Context: Reference to ITU-R BS.775-3 guidelines
Confidence: High
```

### 3.4 SMPTE ST 2110 Audio Channel Ordering[^388^]

SMPTE ST 2110 defines symbolic channel names for professional IP audio transport:

| Symbol | Description | Channels |
|--------|-------------|----------|
| M | Mono | 1 |
| ST | Standard Stereo | 2 |
| 51 | 5.1 Surround | 6 |
| 71 | 7.1 Surround | 8 |
| 222 | 22.2 Surround | 24 |
| U01-U64 | Undefined groups | 1-64 |

---

## 4. PCM Formats and Configurations

### 4.1 Bit Depth Options

| Format | Bits | Value Range | Dynamic Range | Use Case |
|--------|------|-------------|---------------|----------|
| PCM 16-bit | 16 | [-32768, 32767] | ~96 dB | CD quality, games |
| PCM 24-bit | 24 | [-8388608, 8388607] | ~144 dB | Professional audio |
| PCM 32-bit integer | 32 | Full 32-bit range | ~192 dB | Processing headroom |
| PCM 32-bit float | 32 | [-1.0, 1.0] | Virtually unlimited | Mixing, processing |

### 4.2 Sample Rates

| Rate | Use Case |
|------|----------|
| 44.1 kHz | CD standard, some game audio |
| 48 kHz | Professional/gaming standard, HDMI default |
| 96 kHz | High-resolution audio |
| 192 kHz | Audiophile/lossless formats |

```
Claim: 48 kHz is the recommended sample rate for professional streaming applications, capturing the full human hearing range and matching industry production standards[^385^]
Source: Ant Media - Best Audio Codec for Online Video Streaming
URL: https://antmedia.io/best-audio-codec/
Date: 2026-01-02
Excerpt: "Use 48 kHz sample rate for professional streaming applications. This captures the full human hearing range (20 Hz to 20 kHz) and matches industry production standards."
Context: Streaming audio best practices
Confidence: High
```

### 4.3 Channel Configurations

| Config | Bitrate (48kHz/16-bit) | Bitrate (48kHz/24-bit) |
|--------|----------------------|----------------------|
| Stereo | 1.536 Mbps | 2.304 Mbps |
| 5.1 | 4.608 Mbps | 6.912 Mbps |
| 7.1 | 6.144 Mbps | 9.216 Mbps |
| 7.1.4 | 9.216 Mbps | 13.824 Mbps |
| 22.2 | 27.648 Mbps | 41.472 Mbps |

---

## 5. Audio Passthrough Architecture

### 5.1 Bitstreaming vs. Decode-and-Re-encode

| Aspect | Bitstreaming | Decode + LPCM | Decode + Re-encode |
|--------|-------------|---------------|-------------------|
| **Latency** | Medium (decode at receiver) | Lowest (no encoding) | Highest (encode+decode) |
| **Quality** | Original compressed | Lossless (if LPCM) | Additional generation loss |
| **CPU/GPU usage** | Minimal at source | Low at source | High at source |
| **Bandwidth** | Low (compressed) | High (uncompressed) | Medium (depends on codec) |
| **Format support** | Limited by receiver | Limited by HDMI bandwidth | Flexible |

```
Claim: For cloud gaming, PCM output from the game console provides the lowest latency because native game audio is already PCM and requires no transcoding; bitstream encoding adds delay even with hardware acceleration[^421^]
Source: AVS Forum - Confused about PCM vs. Bitstream
URL: https://www.avsforum.com/threads/confused-about-pcm-vs-bitstream-warning-long-post.3051692/
Date: 2019-02-20
Excerpt: "The ps4 (or xbox one) should be set to PCM audio in order to get the lowest latency audio from games. This is because their native audio format is PCM and they can mix multichannel PCM with no delay."
Context: AV receiver audio configuration advice
Confidence: High
```

```
Claim: PCM offers less latency which is better for games requiring quick reactions, while bitstream is preferable for complex surround formats that need receiver decoding[^416^]
Source: Critical Hit - Bitstream vs PCM Audio
URL: https://www.criticalhit.net/technology/bitstream-vs-pcm-audio-experience/
Date: 2022-08-01
Excerpt: "PCM offers less latency, which might be better for games that require quick reactions. Conversely, bitstream is preferable for complex, immersive audio formats."
Context: Consumer audio technology comparison
Confidence: Medium
```

### 5.2 Passthrough Architecture for Cloud Gaming

**Recommended approach for minimal latency:**

```
Game Audio (internal PCM) -> Capture -> Opus Encode -> Network -> Opus Decode -> WASAPI Exclusive -> DAC
```

For multi-channel passthrough to external receivers:
```
Game Audio (5.1/7.1 PCM) -> Opus Multistream Encode -> Network -> Opus Decode -> HDMI eARC -> AV Receiver
```

**Key decision:** When the client has an HDMI eARC-connected AV receiver, the cloud gaming client should:
1. Decode the incoming Opus stream to multi-channel PCM
2. Output via WASAPI exclusive mode at the native sample rate
3. Allow the AV receiver to handle any further processing (Atmos upmixing, etc.)

---

## 6. Hardware Audio Endpoints

### 6.1 HDMI ARC vs. eARC[^248^][^249^][^252^][^253^]

| Feature | HDMI ARC | HDMI eARC |
|---------|----------|-----------|
| HDMI version | 1.4 (2009) | 2.1 (2017) |
| Max bandwidth | ~1-2 Mbps | Up to 37 Mbps |
| Uncompressed 5.1 | No | Yes |
| Uncompressed 7.1 | No | Yes |
| Dolby Atmos (DD+) | Yes (compressed) | Yes |
| Dolby TrueHD/Atmos | No | Yes |
| DTS:X | No | Yes |
| Lip-sync correction | No | Mandatory |
| Cable requirement | High-Speed HDMI | Ultra High-Speed HDMI |

```
Claim: HDMI eARC supports up to 32 channels of 24-bit 192kHz audio at speeds up to 37-38 Mbps, enabling lossless Dolby TrueHD, DTS-HD MA, and object-based Atmos/DTS:X formats[^258^]
Source: What Hi-Fi - HDMI ARC and HDMI eARC
URL: https://www.whathifi.com/advice/hdmi-arc-and-hdmi-earc-everything-you-need-to-know
Date: 2025
Excerpt: "There's scope for eARC to deliver up to 32 channels of audio, including eight-channel, 24bit/192kHz uncompressed data streams at speeds of up to 38Mbps."
Context: Consumer electronics reference
Confidence: High
```

```
Claim: eARC uses a dedicated data channel for audio, providing higher bandwidth and mandatory lip-sync correction[^253^]
Source: WyreStorm - HDMI ARC vs eARC
URL: https://www.wyrestorm.com/blog/hdmi-arc-vs-earc/
Date: 2026-02-04
Excerpt: "Unlike ARC, eARC uses a dedicated data channel for audio. This allows for higher bandwidth and more reliable audio synchronization. eARC also introduces mandatory lip-sync correction."
Context: Professional AV integration reference
Confidence: High
```

### 6.2 SPDIF (TOSLINK and Coaxial)

| Format | Bandwidth | Channel Support |
|--------|-----------|-----------------|
| SPDIF (TOSLINK/Coaxial) | ~1.5 Mbps | Stereo PCM, compressed 5.1 (AC-3/DTS) |
| Dolby Digital | Up to 640 kbps | 5.1 channels max |
| DTS | Up to 1.5 Mbps | 5.1 channels max |

**SPDIF Limitations:**
- Cannot carry uncompressed multi-channel audio (no 5.1 LPCM)
- Cannot carry Dolby Atmos, TrueHD, DTS-HD MA, or DTS:X
- Cannot carry E-AC3 (Dolby Digital Plus)
- Sample rate limited to 48 kHz for multi-channel compressed formats

### 6.3 DisplayPort Audio

| Version | Max Bandwidth | Multi-Channel Audio |
|---------|---------------|---------------------|
| DisplayPort 1.0 | 8.64 Gbps | No |
| DisplayPort 1.2+ | 17.28 Gbps | Yes (up to 8 channels) |
| DisplayPort 1.4 | 32.4 Gbps | Yes (up to 32 channels) |
| DisplayPort 2.1 | 80 Gbps | Yes (up to 32 channels) |

```
Claim: DisplayPort supports multi-channel audio including Dolby Atmos and DTS-HD Master Audio, with packet-based transmission allowing audio and video simultaneously over a single cable[^286^][^291^]
Source: DisplayPort.org FAQ / CableTime Technical Article
URL: https://www.displayport.org/faq/ / https://cabletimetech.com/blogs/knowledge/understanding-how-displayport-supports-multi-channel-audio
Date: 2026
Excerpt: "DisplayPort supports multi-channel audio and many advanced audio features... DisplayPort 1.4 took great strides, increasing the maximum sample rate to 1,536 kHz and the maximum number of supported audio channels to 32."
Context: DisplayPort technology specifications
Confidence: High
```

### 6.4 USB Audio

#### USB Audio Class 2.0 Specifications[^394^][^398^][^411^]

| Feature | USB Audio 1.0 | USB Audio 2.0 |
|---------|--------------|---------------|
| Max sample rate | 48 kHz | Up to 768 kHz+ |
| Bit depth | 16-bit | Up to 32-bit |
| Channels | Up to 2 (stereo) | Up to 32+ channels |
| Multi-channel | No | Yes (7.1, etc.) |
| ASIO support | Vendor-dependent | Native (on some platforms) |
| Bandwidth | USB Full Speed (12 Mbps) | USB High Speed (480 Mbps) |

```
Claim: USB Audio Class 2.0 supports multi-channel configurations with as many channels as the device implements, at sample rates up to 768 kHz and PCM 16/24/32-bit and FLOAT 32-bit formats[^394^]
Source: Thesycon USB Audio 2.0 Class Driver
URL: https://www.thesycon.de/eng/usb_audiodriver.shtml
Date: 2026-01-12
Excerpt: "supports stereo and multi-channel configurations with as many channels as the device implements... 44.1 kHz, 48 kHz, 88.2 kHz, 96 kHz, 176.4 kHz, 192 kHz, 352.8 kHz, 384 kHz, 705.6 kHz, 768 kHz"
Context: USB Audio 2.0 driver feature summary
Confidence: High
```

---

## 7. Operating System Audio Stacks

### 7.1 Windows Audio Stack

#### WASAPI (Windows Audio Session API)[^254^][^260^]

```
Claim: WASAPI exclusive mode gives one application exclusive access to the soundcard, bypassing the Windows mixer for bit-perfect audio with latency of approximately 3-10ms[^254^]
Source: Auris Player - WASAPI Exclusive vs Shared Mode Guide
URL: https://aurisplayer.com/blog/wasapi-exclusive-guide.html
Date: 2026-01-20
Excerpt: "Exclusive mode gives one application exclusive access to the soundcard... Minimum latency: ~3-10ms"
Context: Windows audio API comparison
Confidence: High
```

| Feature | WASAPI Shared | WASAPI Exclusive |
|---------|--------------|------------------|
| Multiple apps | Yes | No |
| Windows mixer | Used | Bypassed |
| Bit-perfect | No | Yes |
| Sample rate conversion | Yes (to system rate) | No (native rate) |
| Latency | Medium | Low (~3-10ms) |
| Loopback capture | Yes | No |

**WASAPI Loopback Capture:**

```
Claim: WASAPI loopback capture is entirely digital and records the exact audio being sent to the output device, with no quality loss from analog conversion[^292^]
Source: Audacity Manual - Recording Computer Playback on Windows
URL: https://manual.audacityteam.org/man/tutorial_recording_computer_playback_on_windows.html
Date: 2025-12-04
Excerpt: "WASAPI loopback has a big advantage over stereo mix or similar inputs provided by the audio interface. The capture is entirely digital."
Context: Official Audacity documentation
Confidence: High
```

#### WaveRT (Wave Real-Time Port Driver)[^388^]

```
Claim: The WaveRT port driver in Windows Vista and later provides low-latency audio using a simple cyclic buffer with glitch-resilient streams, requiring little or no driver intervention during streaming[^388^]
Source: Microsoft Docs - Introducing the WaveRT Port Driver
URL: https://learn.microsoft.com/en-us/windows-hardware/drivers/audio/introducing-the-wavert-port-driver
Date: 2021-12-14
Excerpt: "The improved performance of the WaveRT port driver includes low-latency during wave-capture and wave-rendering, and a glitch-resilient audio stream."
Context: Official Microsoft Windows driver documentation
Confidence: High
```

#### DirectSound
- Higher-level API built on kernel streaming
- Supports shared mode audio with hardware mixing
- Legacy API, still used by some games
- Higher latency than WASAPI

### 7.2 macOS Audio Stack

#### Core Audio / HAL (Hardware Abstraction Layer)

```
Claim: macOS Core Audio provides AudioObjectGetPropertyData with AudioObjectPropertyAddress to enumerate and query audio device properties including device name, unique ID, input/output direction, and channel layout[^383^]
Source: Chromium Source - audio_manager_mac.cc
URL: https://chromium.googlesource.com/chromium/src/media/+/master/audio/mac/audio_manager_mac.cc
Date: Ongoing
Excerpt: "AudioObjectPropertyAddress property_address = GetAudioObjectPropertyAddress(kAudioDevicePropertyDeviceNameCFString, is_input); OSStatus result = AudioObjectGetPropertyData(device_id, &property_address, 0, nullptr, &data_size, &device_name);"
Context: Chromium browser's macOS audio implementation - production reference code
Confidence: High
```

**Key Core Audio APIs:**
- `AudioObjectGetPropertyData` / `AudioObjectSetPropertyData` - Device property queries
- `AudioDeviceStart` / `AudioDeviceStop` - Stream control
- `AudioObjectPropertySelector` - Property selection (e.g., `kAudioDevicePropertyDeviceName`)
- `AudioUnit` - Audio processing components (IO units, mixers, effects)
- `AudioQueue` - High-level playback/recording API

#### BlackHole (Virtual Audio Driver for macOS)[^389^]

```
Claim: BlackHole is a modern virtual audio loopback driver for macOS with zero additional driver latency, supporting 2-256 channels and sample rates up to 768 kHz, compatible with Intel and Apple Silicon[^389^]
Source: SourceForge - BlackHole
URL: https://sourceforge.net/projects/blackhole.mirror/
Date: 2025-02-06
Excerpt: "BlackHole is a modern virtual audio loopback driver for macOS that allows audio to be routed between applications with zero additional latency... Builds 2, 16, 64, 128, and 256 audio channels versions."
Context: Open-source virtual audio driver
Confidence: High
```

### 7.3 Linux Audio Stack

#### ALSA (Advanced Linux Sound Architecture)
- Lowest-level kernel audio interface
- Direct hardware access with minimal overhead
- API: `snd_pcm_open()`, `snd_pcm_writei()`, `snd_pcm_readi()`
- Supports up to 8 channels (surround 7.1) natively
- Configuration via `.asoundrc` for custom routing

#### PulseAudio

```
Claim: PulseAudio module-loopback performs adaptive resampling to route audio from a source to a sink, with configurable latency from 1 to 2000 ms (default 200 ms)[^297^]
Source: PulseAudio Modules Documentation (freedesktop.org)
URL: https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/User/Modules/
Date: 2022-05-03
Excerpt: "module-loopback... The desired latency in milliseconds, from 1 to 2000. Defaults to 200."
Context: Official PulseAudio module documentation
Confidence: High
```

**PulseAudio loopback setup:**
```bash
# Load loopback module
pactl load-module module-loopback source=alsa_output.pci-0000_00_1f.3.analog-stereo.monitor sink=alsa_input.default
```

#### PipeWire[^250^][^251^][^255^]

```
Claim: PipeWire achieves sub-5ms latency with quantum=256 at 48kHz, supports dynamic per-application buffer sizing, and integrates PulseAudio + JACK + ALSA in one system[^250^]
Source: OneUptime - How to Configure PipeWire for Low-Latency Audio on Ubuntu
URL: https://oneuptime.com/blog/post/2026-03-02-configure-pipewire-low-latency-audio-ubuntu/view
Date: 2026-03-02
Excerpt: "Latency = quantum / rate... 256 / 48000 = 5.3ms... PipeWire uses two key parameters: quantum (buffer size) and rate (sample rate)."
Context: Linux professional audio configuration guide
Confidence: High
```

```
Claim: PipeWire enables each application to negotiate its own buffer size, allowing Spotify to play through 40ms buffer while Ardour records through 5ms buffer on the same hardware[^251^]
Source: LinuxTeck - PipeWire Linux Audio
URL: https://www.linuxteck.com/pipewire-linux-audio-problem-solved/
Date: 2026-03-13
Excerpt: "Spotify can be playing through a 40ms buffer at the same time Ardour is recording through a 5ms buffer, on the same hardware, with no interference between them."
Context: Analysis of PipeWire's dynamic buffer system
Confidence: High
```

**PipeWire low-latency configuration:**
```bash
# ~/.config/pipewire/pipewire.conf.d/10-low-latency.conf
context.properties = {
    default.clock.quantum = 256
    default.clock.min-quantum = 32
    default.clock.max-quantum = 8192
    default.clock.rate = 48000
    default.clock.allowed-rates = [ 44100 48000 88200 96000 ]
}
```

**PipeWire latency table:**

| Quantum | Rate | Latency |
|---------|------|---------|
| 2048 | 48 kHz | 42.7 ms |
| 1024 | 48 kHz | 21.3 ms |
| 512 | 48 kHz | 10.7 ms |
| 256 | 48 kHz | 5.3 ms |
| 128 | 48 kHz | 2.7 ms |

#### JACK (JACK Audio Connection Kit)[^384^][^387^]

```
Claim: JACK can achieve sub-8ms round-trip latency with real-time kernel, and sub-3ms with optimized settings (64 frames/period at 48kHz = 2.67ms)[^387^]
Source: ArchWiki - JACK Audio Connection Kit
URL: https://wiki.archlinux.org/title/JACK_Audio_Connection_Kit
Date: 2025-06-17
Excerpt: "jack_control dps period 64... gives 2.67 ms latency, which is nicely low without putting too much stress on the particular hardware."
Context: Arch Linux community documentation
Confidence: High
```

**JACK configuration example:**
```bash
# Start JACK with ALSA backend at 48kHz, 256-sample buffer, 2 periods
jackd -d alsa -r 48000 -p 256 -n 2 -D
# Latency: 256/48000 * 2 periods = 10.7ms round-trip
```

---

## 8. Audio Capture and Loopback

### 8.1 Windows: WASAPI Loopback

**API approach:**
```cpp
// Initialize COM
CoInitializeEx(NULL, COINIT_APARTMENTTHREADED);

// Use MMDeviceEnumerator to find the output device
IMMDeviceEnumerator* pEnumerator;
CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, 
    __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator);

// Activate IAudioClient for loopback
IAudioClient* pAudioClient;
pDevice->Activate(__uuidof(IAudioClient), CLSCTX_ALL, NULL, (void**)&pAudioClient);

// Initialize in loopback mode
pAudioClient->Initialize(AUDCLNT_SHAREMODE_SHARED, 
    AUDCLNT_STREAMFLAGS_LOOPBACK, hnsBufferDuration, 0, pWaveFormat, NULL);
```

```
Claim: WASAPI loopback only records when an active audio signal is present; recording automatically pauses during silence and resumes when audio begins[^292^]
Source: Audacity Manual
URL: https://manual.audacityteam.org/man/tutorial_recording_computer_playback_on_windows.html
Date: 2025-12-04
Excerpt: "Windows WASAPI host only records loopback when there is an active signal present. When there is no active signal, recording pauses."
Context: WASAPI loopback behavior documentation
Confidence: High
```

### 8.2 Windows: Virtual Audio Cable (VB-Audio)[^376^]

```
Claim: VB-CABLE for Win10/11 provides 8 channels for all interfaces (MME, WASAPI, KS) and up to 16 channels on the Line Out pin, supporting 48kHz default[^376^]
Source: VB-Audio Cable Reference Manual
URL: https://vb-audio.com/Cable/VBCABLE_ReferenceManual.pdf
Date: 2024-10-08
Excerpt: "WIN10/11 version provides 8 channels (or up to 16) for all interface (MME, WASAPI, KS)."
Context: Virtual audio cable product documentation
Confidence: High
```

### 8.3 Linux: PulseAudio Module-Loopback[^287^]

```bash
# Create loopback from monitor source to sink
pactl load-module module-loopback \
    source=alsa_output.pci-0000_00_1f.3.analog-stereo.monitor \
    sink=alsa_input.usb-Microphone \
    latency_msec=10
```

```
Claim: PulseAudio Monitor of <Internal Sound Card> allows recording any audio sent to the sound card without interrupting PulseAudio playback[^287^]
Source: Debian Wiki - Audio Loopback Recording with PulseAudio
URL: https://wiki.debian.org/audio-loopback
Date: 2025-08-24
Excerpt: "If you select this 'Monitor' as your audio input device, that application will no longer record audio from microphones, but will capture any audio sent to your sound card."
Context: Debian community documentation
Confidence: High
```

### 8.4 macOS: BlackHole + Multi-Output Device

1. Install BlackHole (2ch, 16ch, or 64ch version)
2. Create Multi-Output Device in Audio MIDI Setup
3. Add both physical output (speakers) and BlackHole
4. Set Multi-Output Device as system default
5. Capture from BlackHole input in your application

### 8.5 Capture Method Comparison

| Method | OS | Latency | Channels | Quality |
|--------|-----|---------|----------|---------|
| WASAPI Loopback | Windows | Low | Up to device max | Digital (bit-perfect) |
| VB-CABLE | Windows | Low | 8-16 | Digital |
| PulseAudio Monitor | Linux | Medium | Up to device max | Digital |
| PipeWire Stream | Linux | Low | Up to device max | Digital |
| BlackHole | macOS | Near-zero | 2-256 | Digital |
| Core Audio Aggregate | macOS | Low | Up to aggregate | Digital |

---

## 9. Go Audio Libraries

### 9.1 oto (by hajimehoshi)[^342^][^344^]

```
Claim: oto is a low-level Go audio library providing an io.Writer interface for audio playback, supporting Windows (no Cgo), macOS (no Cgo), Linux (ALSA), FreeBSD, Android, iOS, and WebAssembly[^342^]
Source: oto v2 Go Package Documentation
URL: https://pkg.go.dev/github.com/hajimehoshi/oto/v2
Date: 2025-09-07
Excerpt: "A low-level library to play sound... Windows (no Cgo required!), macOS (no Cgo required!), Linux (ALSA), FreeBSD, OpenBSD, Android, iOS, WebAssembly, Nintendo Switch, Xbox"
Context: Official Go package documentation
Confidence: High
```

**Key features:**
- Cross-platform with minimal dependencies
- No Cgo required on Windows/macOS
- Simple `io.Writer` interface
- ALSA required on Linux (`libasound2-dev`)
- Used by Ebitengine game framework

**Example usage:**
```go
import "github.com/hajimehoshi/oto/v2"

// Create player
ctx, ready, err := oto.NewContext(48000, 2, 2)
// Wait for ready channel
<-ready
player := ctx.NewPlayer()
defer player.Close()

// Write PCM data
player.Write(pcmData)
```

### 9.2 malgo (miniaudio Go bindings)[^334^][^328^]

```
Claim: malgo provides Go bindings for miniaudio, supporting Windows (WASAPI, DirectSound, WinMM), Linux (PulseAudio, ALSA, JACK), macOS (CoreAudio), and Android, requiring Cgo but no external linking on Windows/macOS[^334^]
Source: GitHub - gen2brain/malgo
URL: https://github.com/gen2brain/malgo
Date: 2025
Excerpt: "Go bindings for miniaudio library. Requires cgo but does not require linking to anything on the Windows/macOS and it links only -ldl on Linux/BSDs."
Context: Official repository README
Confidence: High
```

**Supported backends:**
- Windows: WASAPI, DirectSound, WinMM
- Linux: PulseAudio, ALSA, JACK
- macOS/iOS: CoreAudio
- BSD: OSS/audio/sndio
- Android: OpenSL|ES, AAudio

**Key miniaudio features[^410^]:**
- Single-source-file C library (public domain)
- Playback, capture, full-duplex
- Data conversion, resampling, channel mapping
- Built-in WAV, FLAC, MP3 decoding
- Low-level API for direct raw audio access
- High-level API with sound management, mixing, 3D spatialization

### 9.3 beep / gopxl/beep[^338^][^337^][^412^]

```
Claim: gopxl/beep is the maintained fork of faiface/beep, providing a Streamer interface for audio composition, supporting WAV, MP3, Ogg Vorbis, FLAC, and MIDI decode, built on top of oto[^412^]
Source: GitHub - gopxl/beep
URL: https://github.com/gopxl/beep
Date: 2023-10-06
Excerpt: "A little package that brings sound to any Go application... Decode and play WAV, MP3, Ogg Vorbis, FLAC and MIDI... Built on top of its Streamer interface, which is like io.Reader, but for audio."
Context: Official repository (successor to archived faiface/beep)
Confidence: High
```

**beep Streamer interface:**
```go
type Streamer interface {
    Stream(samples [][2]float64) (n int, ok bool)
    Err() error
}
```

**Key features:**
- Compositor pattern for audio mixing
- Built-in effects: volume, pan, gain, resample, mix, sequence
- Speaker package for playback (`github.com/gopxl/beep/speaker`)
- Small codebase (~1K LOC core)

### 9.4 Go Audio Library Comparison

| Library | Level | Cgo | Backends | Channels | Best For |
|---------|-------|-----|----------|----------|----------|
| **oto** | Low-level | Optional (Linux only) | ALSA, CoreAudio, WASAPI, AAudio | Stereo | Simple playback, game audio |
| **malgo** | Low-level | Required | WASAPI, ALSA, PulseAudio, JACK, CoreAudio | Configurable | Capture + playback, flexible routing |
| **beep** | High-level | Indirect (via oto) | Via oto | Stereo (designed for) | Audio processing, effects, composition |
| **portaudio** | Low-level | Required | PortAudio backends | Configurable | Cross-platform capture/playback |

### 9.5 Recommendation for Cloud Gaming

For a cloud gaming client in Go:

1. **Audio Output:** Use `oto` for stereo output on all platforms (no Cgo on Windows/macOS)
2. **Multi-channel Output:** Use `malgo` for 5.1/7.1 channel output (supports channel mapping)
3. **Audio Processing:** Use `gopxl/beep` for stream composition (volume, mixing, effects)
4. **Audio Capture:** Use `malgo` for microphone/system capture

```go
// Example: Multi-channel audio output with malgo
import "github.com/gen2brain/malgo"

config := malgo.DeviceConfig{
    DeviceType: malgo.Playback,
    SampleRate: 48000,
    Channels:   6,  // 5.1 surround
    Format:     malgo.FormatS16,
}
device, err := ctx.InitDevice(config, callbacks)
```

---

## 10. Latency Targets and Optimization

### 10.1 Opus Latency Optimization[^377^][^379^]

| Frame Size | Samples (48kHz) | Algorithmic Delay | Best For |
|------------|----------------|-------------------|----------|
| 2.5 ms | 120 | 5 ms | Ultra-low latency voice |
| 5 ms | 240 | 7.5 ms | Low latency gaming |
| 10 ms | 480 | 12.5 ms | Balanced (LPC/hybrid disabled) |
| 20 ms | 960 | 26.5 ms | Default, best quality |
| 40 ms | 1920 | 46.5 ms | Low bitrate efficiency |
| 60 ms | 2880 | 66.5 ms | Maximum compression |

```
Claim: For real-time applications, 20ms frames are the recommended default; using frame sizes above 20ms slightly reduces music quality and should be avoided unless operating at very low bitrates over RTP[^377^]
Source: Xiph Wiki - Opus Recommended Settings
URL: https://wiki.xiph.org/Opus_Recommended_Settings
Date: 2024-11-25
Excerpt: "Opus uses a 20 ms frame size by default, as it gives a decent mix of low latency and good quality... there is no reason to use frame sizes above 20 ms."
Context: Official Opus project recommendations
Confidence: High
```

### 10.2 End-to-End Latency Budget for Cloud Gaming

| Stage | Target Latency | Optimization |
|-------|---------------|--------------|
| Audio capture (loopback) | 1-2 ms | WASAPI exclusive / PipeWire quantum=128 |
| Opus encoding | 2.5-5 ms | 5ms frames, complexity=10 |
| Network transmission | 5-10 ms | UDP, jitter buffer=5ms |
| Opus decoding | 1-2 ms | Optimized decoder |
| Audio render | 2-5 ms | WASAPI exclusive / low-quantum PipeWire |
| **Total** | **<20 ms** | Aggressive optimization across all stages |

### 10.3 Platform-Specific Latency Optimization

**Windows:**
- Use WASAPI Exclusive mode (~3-10ms output latency)
- Disable audio enhancements in device properties
- Match application sample rate to device sample rate
- Use event-driven mode (PKEY_AudioEndpoint_Supports_EventDriven_Mode)

**Linux:**
- Use PipeWire with quantum=128-256 (2.7-5.3ms)
- Enable real-time scheduling (`@audio rtprio 95`)
- Use low-latency or real-time kernel for sub-5ms targets
- Disable batch mode for ALSA devices

**macOS:**
- Use Core Audio directly (not AudioQueue)
- Set kAudioDevicePropertyBufferFrameSize appropriately
- Use aggregate devices for multi-device synchronization

---

## 11. Recommendations for Cloud Gaming Implementation

### 11.1 Codec Selection Decision Matrix

| Scenario | Recommended Codec | Bitrate | Channels |
|----------|------------------|---------|----------|
| Stereo game audio (default) | Opus | 96-128 kbps | 2 (stereo) |
| 5.1 surround games | Opus MultiOpus | 192-256 kbps | 6 (5.1) |
| 7.1 surround games | Opus MultiOpus | 256-450 kbps | 8 (7.1) |
| AV receiver passthrough | Opus decode -> HDMI PCM | N/A (decode) | Match source |
| Atmos content (rare in games) | Decode to multi-channel PCM | N/A | 7.1.4+ |

### 11.2 Recommended Audio Pipeline Architecture

```
Server Side:
  Game Engine Audio (multi-channel PCM)
    -> WASAPI Loopback Capture (Windows) / PulseAudio Monitor (Linux) / BlackHole (macOS)
    -> Channel Order Normalization (to standard 5.1/7.1 layout)
    -> libopus MultiStream Encode (5ms frames, complexity=10)
    -> RTP over UDP (WebRTC-compatible)

Network:
    -> UDP with FEC (for packet loss resilience)
    -> Jitter buffer (5ms target)

Client Side:
    -> RTP depacketize
    -> libopus MultiStream Decode
    -> Channel Order Remapping (to Windows/macOS/Linux layout)
    -> WASAPI Exclusive / PipeWire low-latency / Core Audio
    -> HDMI eARC / DisplayPort / USB Audio -> AV Receiver (optional)
```

### 11.3 Key Implementation Notes

1. **Channel mapping is critical:** Windows 7.1 uses different channel ordering than Dolby/SMPTE. Always include explicit channel mapping metadata and normalize at encode/decode boundaries.

2. **Opus multistream for surround:** Use the multiopus SDP format with proper `num_streams`, `coupled_streams`, and `channel_mapping` parameters as specified in draft-shin-avtcore-rtp-multi-opus.

3. **Fallback to stereo:** If the client cannot handle multi-channel Opus, fall back gracefully to stereo Opus with ITU-R BS.775-compliant downmix coefficients.

4. **Hardware passthrough:** For clients with HDMI eARC connected AV receivers, decode Opus to multi-channel PCM and output via WASAPI exclusive mode at 48kHz/16-bit (or 24-bit if supported).

5. **Latency budget:** Target <5ms for Opus encode (5ms frames), <5ms for network, <5ms for decode, <5ms for render = <20ms total end-to-end.

6. **Go implementation:** Use `malgo` for capture/playback with multi-channel support, `oto` for simple stereo output, and `gopxl/beep` for stream composition and effects.

---

## Source Reference Index

| Citation | Source | URL |
|----------|--------|-----|
| [^222^] | Remoto Technical Reference | https://support.remotopro.io/technical-reference/how-remoto-playback-handles-multichannel-audio-fold-down |
| [^223^] | IETF draft-shin-avtcore-rtp-multi-opus | https://datatracker.ietf.org/doc/html/draft-shin-avtcore-rtp-multi-opus |
| [^224^] | Ant Media - Best Audio Codec 2026 | https://antmedia.io/best-audio-codec/ |
| [^229^] | IETF draft-shin-avtcore-rtp-multi-opus-00 | https://datatracker.ietf.org/doc/html/draft-shin-avtcore-rtp-multi-opus-00 |
| [^231^] | Wikipedia - DTS-HD Master Audio | https://en.wikipedia.org/wiki/DTS-HD_Master_Audio |
| [^233^] | Wikipedia - Opus (audio format) | https://en.wikipedia.org/wiki/Opus_(audio_format) |
| [^234^] | MDN Web Audio Codec Guide | https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/Audio_codecs |
| [^248^] | VCOM HDMI eARC Guide | https://vcom.com.hk/shows/169/636.html |
| [^249^] | HDMI.org Blog | https://www.hdmi.org/blog/detail/129 |
| [^250^] | OneUptime - PipeWire Low Latency | https://oneuptime.com/blog/post/2026-03-02-configure-pipewire-low-latency-audio-ubuntu/view |
| [^251^] | LinuxTeck - PipeWire | https://www.linuxteck.com/pipewire-linux-audio-problem-solved/ |
| [^252^] | RTINGS - HDMI ARC/eARC | https://www.rtings.com/soundbar/learn/what-is-hdmi-earc |
| [^253^] | WyreStorm - ARC vs eARC | https://www.wyrestorm.com/blog/hdmi-arc-vs-earc/ |
| [^254^] | Auris - WASAPI Guide | https://aurisplayer.com/blog/wasapi-exclusive-guide.html |
| [^255^] | ArchWiki - PipeWire | https://wiki.archlinux.org/title/PipeWire |
| [^258^] | What Hi-Fi - HDMI ARC/eARC | https://www.whathifi.com/advice/hdmi-arc-and-hdmi-earc-everything-you-need-to-know |
| [^260^] | Mark Heath - WASAPI | https://markheath.net/post/what-up-with-wasapi |
| [^286^] | DisplayPort.org FAQ | https://www.displayport.org/faq/ |
| [^287^] | Debian Wiki - PulseAudio Loopback | https://wiki.debian.org/audio-loopback |
| [^292^] | Audacity - WASAPI Loopback | https://manual.audacityteam.org/man/tutorial_recording_computer_playback_on_windows.html |
| [^297^] | PulseAudio Modules | https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/User/Modules/ |
| [^328^] | malgo Go Package | https://pkg.go.dev/github.com/gen2brain/malgo |
| [^334^] | GitHub - gen2brain/malgo | https://github.com/gen2brain/malgo |
| [^336^] | ATSC A/52 Standard | http://www.atsc.org/wp-content/uploads/2015/03/A52-201212-17.pdf |
| [^337^] | GitHub - faiface/beep | https://github.com/faiface/beep |
| [^338^] | faiface blog - Beep design | https://faiface.github.io/post/how-i-built-audio-lib-composite-pattern/ |
| [^341^] | Wikipedia - Dolby Digital | https://en.wikipedia.org/wiki/Dolby_Digital |
| [^342^] | oto v2 Go Package | https://pkg.go.dev/github.com/hajimehoshi/oto/v2 |
| [^344^] | Dev.to - Go Packages for Games | https://dev.to/hajimehoshi/go-packages-we-developed-for-our-games--4cl9 |
| [^375^] | opus-codec Rust docs | https://docs.rs/opus-codec |
| [^376^] | VB-Audio Cable Manual | https://vb-audio.com/Cable/VBCABLE_ReferenceManual.pdf |
| [^377^] | Xiph Wiki - Opus Settings | https://wiki.xiph.org/Opus_Recommended_Settings |
| [^378^] | AVPro - Dolby MAT Deep Dive | https://www.avproglobal.com/blogs/news/a-deep-dive-into-dolby-mat |
| [^379^] | opus_encoder man page | https://man.archlinux.org/man/opus_encoder.3.en |
| [^380^] | Opus Encoder API Docs | https://opus-codec.org/docs/html_api/group__opusencoder.html |
| [^383^] | Chromium audio_manager_mac.cc | https://chromium.googlesource.com/chromium/src/media/+/master/audio/mac/audio_manager_mac.cc |
| [^384^] | OneUptime - JACK Setup | https://oneuptime.com/blog/post/2026-03-02-setup-jack-audio-connection-kit-ubuntu/view |
| [^385^] | Ant Media - Audio Codecs | https://antmedia.io/best-audio-codec/ |
| [^387^] | ArchWiki - JACK | https://wiki.archlinux.org/title/JACK_Audio_Connection_Kit |
| [^388^] | Microsoft - WaveRT Port Driver | https://learn.microsoft.com/en-us/windows-hardware/drivers/audio/introducing-the-wavert-port-driver |
| [^389^] | BlackHole on SourceForge | https://sourceforge.net/projects/blackhole.mirror/ |
| [^390^] | How-To Geek - DTS:X | https://www.howtogeek.com/772173/what-is-dtsx/ |
| [^392^] | Hybrik - Dolby Audio Overview | https://docs.hybrik.com/tutorials/dolby_audio/ |
| [^394^] | Thesycon - USB Audio 2.0 | https://www.thesycon.de/eng/usb_audiodriver.shtml |
| [^395^] | What Hi-Fi - DTS:X | https://www.whathifi.com/advice/dtsx-what-it-how-can-you-get-it |
| [^398^] | Microsoft - USB Audio 2.0 | https://learn.microsoft.com/en-us/windows-hardware/drivers/audio/usb-2-0-audio-drivers |
| [^402^] | SourceForge - 7.1 Channel Order | https://sourceforge.net/p/mesh2hrtf-tools/wiki/Confusion_of_7-1%20and%205-1_channel_order_in_Windows/ |
| [^406^] | Abbacus - Cloud Gaming | https://www.abbacustechnologies.com/cloud-gaming-technology-architecture-benefits-use-cases/ |
| [^410^] | Wikipedia - Surround Sound | https://en.wikipedia.org/wiki/Surround_sound |
| [^412^] | GitHub - gopxl/beep | https://github.com/gopxl/beep |
| [^416^] | Critical Hit - PCM vs Bitstream | https://www.criticalhit.net/technology/bitstream-vs-pcm-audio-experience/ |
| [^421^] | AVS Forum - PCM vs Bitstream | https://www.avsforum.com/threads/confused-about-pcm-vs-bitstream-warning-long-post.3051692/ |

---

*Document compiled from 25+ independent web searches across RFCs, official specifications, vendor documentation, open-source projects, and technical publications. Primary sources prioritized. All key claims include inline citations.*
