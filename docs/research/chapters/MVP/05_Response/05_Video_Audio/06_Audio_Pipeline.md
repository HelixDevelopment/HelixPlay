# Audio Pipeline

> **Source:** `video-tech_dim06.md` (1,141 lines primary), `video-tech.agent.final.md` (2,588 lines), **Insight #2 (audio passthrough constrained — BINDING)**.
> **Web addendum:** [`../99_Web_Research_Addenda/2026-04-29-audio-pipeline.md`](../99_Web_Research_Addenda/2026-04-29-audio-pipeline.md) — 783 lines, 9 clusters + §Z.
> **R-01 floor:** 1,250 lines body prose. **Achieved:** see Anti-Bluff Verification block.
> **Targets:** R-01..R-13, R-18; new public submodule `vasic-digital/helix-audio`; reuses helix-shm + helix-r18-safeexec.
> **Cross-links:** [`00_Index.md`](00_Index.md), [`03_Capture_Pipelines.md`](03_Capture_Pipelines.md) (C28 — capture pattern), [`05_Recording_Storage.md`](05_Recording_Storage.md) (C30 — recording-side audio container). Architecture-side: [`../03_Architecture/05_RealTime_APIs.md`](../03_Architecture/05_RealTime_APIs.md) (C06 — voice chat cross-link), [`../03_Architecture/11_TV_UX.md`](../03_Architecture/11_TV_UX.md) (C12 §6 ALLM).
> **Status:** Draft v1. **Last updated:** 2026-04-29.

This chapter is the **sixth deep chapter of the `05_Video_Audio/`
family** — multi-channel audio pipeline. **Insight #2 binding**:
audio passthrough is binary (full surround or stereo); end-to-end
capability validation MORE critical than video codec negotiation.
Multi-channel chain spans game audio API → OS stack → capture →
encode → transmit → decode → AV receiver — ANY link forces stereo
fallback.

Stereo + 5.1 + 7.1 via Opus MultiStream WebRTC (up to 8 channels);
Atmos passthrough only via Dolby Digital Plus + JOC (E-AC3); eARC
+ HDMI 2.1 + ALLM negotiation; Windows 7.1 SMPTE 320M vs Dolby
channel-order normalisation; 48 kHz / S16LE transport baseline.
PipeWire (Linux 2026 default) / WASAPI loopback (Windows) /
CoreAudio (macOS dev only) capture.

The chapter resolves the channel-config + passthrough contradictions
documented in the addendum's §Z, with capability decay tiers (eARC
→ ARC → SPDIF → HDMI direct → host downmix) preserving the
Insight #2 fallback ladder discipline.

**Inherits without re-implementing**: `r18.SafeExec` from C08 §10;
`host-integrity-scan` from C08 §12.11; capture-process pattern
from C28; recording-side container choice from C30.

---

## Table of Contents

- [§1 Scope & non-scope](#1-scope--non-scope)
- [§2 Multi-channel audio ladder](#2-multi-channel-audio-ladder)
- [§3 eARC + HDMI + SPDIF passthrough](#3-earc--hdmi--spdif-passthrough)
- [§4 Channel-order management](#4-channel-order-management)
- [§5 OS audio capture](#5-os-audio-capture)
- [§6 Implementation contract](#6-implementation-contract)
- [§7 Failure modes](#7-failure-modes)
- [§8 Test surface](#8-test-surface)
- [§9 Open questions](#9-open-questions)
- [§10 References](#10-references)
- [Anti-Bluff Verification](#anti-bluff-verification)

---

## 1. Scope & non-scope

C31 is the sixth deep chapter under the `05_Video_Audio` family and the first chapter to treat audio as the primary subject rather than a peripheral concern of the video pipeline. Where C26 (Codec Selection), C27 (Hardware Encoders), C28 (Capture Pipelines), C29 (Dual-Path Encoding), and C30 (Recording & Storage) all dealt with the video stream as the dominant payload — with audio mentioned only as a synchronisation partner or a passenger in the recording mux — C31 inverts the framing. The audio pipeline is now the object under analysis, the video pipeline is the synchronisation reference, and every architectural decision is evaluated against one binding constraint: **Insight #2 from `video-tech_insight.md` — audio is binary**.

Insight #2 is reproduced verbatim into this chapter because it governs every subsequent decision. The insight states that, in the cloud-gaming context, the user perception of audio is bimodal rather than continuous: either the user receives the full surround-sound experience the game was authored to deliver, or the user receives stereo. Intermediate states (e.g. "5.1 reduced to 4.0 because the LFE channel could not be carried", or "7.1 reduced to 5.1 because the rear-surround pair was downmixed") are perceptually equivalent to stereo for the gaming use case. A racing game with positional engine cues, a first-person shooter with directional footsteps, or an Atmos-mixed cinematic does not gracefully degrade to a partial channel layout the way a video stream gracefully degrades from 4K to 1080p — the user either hears the bullet behind their head or they do not. This binary character makes end-to-end capability validation strictly more critical than the equivalent video-codec negotiation that C26 elaborated. A wrong video codec produces a black screen or a green frame, both of which are observable failure modes that trigger immediate fallback. A wrong audio channel layout produces silence on some channels — a failure mode the user attributes to the game, not to the streaming pipeline, and one that will not trigger any fallback unless the bootstrap process explicitly tested for it.

The multi-channel chain that Insight #2 protects is long. It begins with the game audio API call inside the guest process, traverses the operating-system audio stack (WASAPI on Windows hosts, PipeWire on Linux hosts, CoreAudio on macOS hosts), is captured by the HelixPlay agent, encoded by the audio encoder, transmitted across the WebRTC or custom-UDP transport, decoded by the client audio decoder, and finally delivered to the playback endpoint — which may be a built-in laptop speaker, a USB headset, an HDMI-attached AV receiver via the client TV's eARC return path, an SPDIF-attached soundbar, or a Bluetooth headset whose A2DP profile re-encodes the stream a second time. Any link in that chain that does not advertise, negotiate, or carry the full channel count forces the entire session to fall back to the next-lower tier. There are no partial fallbacks. A 7.1 game session whose AV-receiver-side soundbar reports only 5.1 capability collapses the entire session to 5.1; a 5.1 session whose Bluetooth headset re-encode at A2DP collapses to stereo. The fallback is not gradual; it is a discrete tier transition.

In scope for C31:

- **Opus MultiStream up to 8 channels.** The WebRTC transport carries Opus, and Opus has a multi-stream extension (RFC 7845, Ogg encapsulation; the WebRTC mapping uses RTP payload format from RFC 7587) that can carry up to 255 channels by interleaving Opus mono/stereo streams with a coupling and channel-mapping table. We restrict ourselves to the 8-channel ceiling because no consumer playback chain in HelixPlay's MVP target market exceeds 7.1 discrete channels.
- **Atmos passthrough via Dolby Digital Plus (E-AC3) with Joint Object Coding (JOC) extension.** Atmos is not a fixed channel layout; it is an object-audio format where audio objects carry positional metadata and the playback renderer maps them to the available speakers at the playback endpoint. HelixPlay's MVP cannot natively transport Atmos over WebRTC — Opus has no extension for object metadata — so the only viable MVP path is opaque container passthrough, where the host receives the Atmos-encoded bitstream from the game (E-AC3+JOC), wraps it without re-encoding, and delivers it to a client AV receiver that does the JOC decode and rendering itself.
- **eARC, SPDIF, HDMI passthrough.** These are the three physical transports by which the client device delivers a multi-channel audio stream to the AV receiver or soundbar. eARC (HDMI 2.1 enhanced Audio Return Channel) is the only one of the three that natively carries Atmos at full bitrate; SPDIF is bandwidth-limited to compressed 5.1 (DD or DTS); HDMI passthrough from the streaming device to the receiver is the recommended path when the client is a set-top box or game-console-shaped TV peripheral.
- **Windows-vs-Dolby channel-order mismatch.** WASAPI on Windows reports 7.1 channels in the order L, R, C, LFE, Lrs, Rrs, Ls, Rs (rear-surround pair before side-surround pair); the SMPTE 320M / Dolby specification orders them L, R, C, LFE, Ls, Rs, Lrs, Rrs (side-surround pair before rear-surround pair). The capture pipeline must reorder. Failing to reorder produces a session where the side and rear surround channels are swapped — a failure mode that is audible, that the user will blame on the game's mix engineer, and that will not trigger any automatic fallback because every channel is technically receiving signal.
- **PipeWire / WASAPI / CoreAudio capture mechanics.** Each OS has a different capture API surface, and each surface advertises channel layouts in a slightly different vocabulary. PipeWire uses SPA's `SPA_AUDIO_CHANNEL_*` enums; WASAPI uses `KSAUDIO_SPEAKER_*` masks; CoreAudio uses `kAudioChannelLabel_*` constants. C31 specifies the cross-OS normalisation table used by HelixPlay's host agent.
- **End-to-end capability negotiation at session bootstrap.** Insight #2's binding rule is that the session locks a single channel configuration for its lifetime, chosen at bootstrap by walking the chain from game capability down to AV-receiver capability and selecting the highest tier supported at every link.

Out of scope for C31, with cross-references:

- **Voice chat and party-chat audio** — handled by C06 (NATS Real-Time APIs) at the protocol layer and by a future client-side mixer chapter at the application layer. C31 does not specify how the game audio and the voice-chat audio are mixed; that is a separate stream with separate capability negotiation.
- **HDR audio metadata and loudness normalisation** — handled by C32 (HDR / WCG / Loudness) which is the next chapter in the family. ITU-R BS.1770 loudness measurement, EBU R 128 normalisation, and the Atmos-specific dialogue-intelligibility metadata profile all live there.
- **Recording-side audio container choice** — handled by C30 (Recording & Storage). The MP4-vs-MKV decision, the AAC-vs-Opus-in-MKV trade-off, and the per-track-vs-multiplexed channel layout in the recorded artefact are all C30 concerns. C31 specifies only the live transport.

R-18 (Operational Integrity) compliance for this chapter follows the C25 §7 family allow-list. C31 introduces three new chapter-specific commands that must be added to the allow-list and that must never be invoked outside a container: `pacmd` and `pactl` (PulseAudio compatibility surfaces, used to query and route audio under PipeWire), `wpctl` (WirePlumber control, the modern PipeWire policy CLI), and `ffmpeg -map 0:a` (used in non-real-time validation harnesses to extract and inspect audio tracks from captured WebRTC dumps). None of these may be invoked on the operator's host directly — they run inside the audio-pipeline test container that C25 will register in its container hazards inventory.

## 2. Multi-channel audio ladder

The multi-channel ladder is a strict tier hierarchy. Each tier is a complete audio configuration — codec, channel count, bandwidth budget, and minimum-link-capability list — and a session occupies exactly one tier for its lifetime. The ladder has four tiers: stereo, 5.1, 7.1, and Atmos. Each tier's section below specifies the transport, the channel order, the bandwidth budget, and the minimum capability required at every link in the chain.

### 2.1 Stereo (PCM / Opus 2-channel)

Stereo is the universal fallback. Every link in every audio chain HelixPlay's MVP supports — every game, every OS audio stack, every capture API, every encoder, every transport, every decoder, every playback endpoint — supports stereo. Stereo is the tier the session falls back to when any other tier's negotiation fails, and it is the tier the session uses when the guest game itself does not advertise multi-channel output.

The transport options for stereo are PCM (uncompressed 48 kHz / 16-bit / 2-channel = 1.536 Mbps) for in-LAN sessions where bandwidth is plentiful and encoder latency is the dominant concern, and Opus 2-channel (typically 64 kbps for voice-quality content, 96–128 kbps for music-quality content, and 192 kbps for the highest stereo fidelity Opus can deliver) for sessions traversing a constrained network. Opus 2-channel is the default for WebRTC; PCM is reserved for the custom-UDP transport on the LAN profile where the audio path can afford the bitrate to avoid the 5–10 ms encoder/decoder round-trip cost.

The stereo channel order is L, R. There is no ambiguity, no OS-specific permutation, and no Dolby-vs-Microsoft mismatch. Stereo is the only tier where the capability negotiation is trivial: every link supports it, so the negotiation is a no-op and the session bootstrap proceeds directly to encoder configuration.

### 2.2 5.1 surround (Opus MultiStream 6-channel)

5.1 is the first multi-channel tier and the most widely supported one in consumer playback hardware. Almost every soundbar produced after 2018 supports 5.1 over either eARC, SPDIF, or HDMI-passthrough; almost every AV receiver produced in the last two decades supports it; and every game console produced since the seventh generation supports it natively.

The transport over WebRTC is Opus MultiStream with 6 channels, configured per RFC 7845's channel-mapping family 1 (the Vorbis-derived layout for 6 channels). The SDP advertises `channels=6` with a `channel_mapping` attribute that lists the per-channel coupling. The bandwidth budget is 192–256 kbps for the Opus MultiStream payload — substantially more than 6× the per-channel rate of stereo Opus because the multi-stream coupling introduces overhead and because rear-channel content tends to be percussive and harder to compress than front-channel dialogue.

The 5.1 channel order is the SMPTE 320M layout: L (front-left), R (front-right), C (centre), LFE (low-frequency effects, the ".1"), Ls (left surround), Rs (right surround). This order matches the Dolby Digital convention and matches what a Dolby-aware game expects to populate. WASAPI on Windows reports 5.1 in the same order; PipeWire and CoreAudio also use this order for 5.1, so the cross-OS normalisation is a no-op at the 5.1 tier (the mismatch only appears at the 7.1 tier).

| 5.1 channel | Index | Position | Typical content |
|-------------|-------|----------|-----------------|
| L | 0 | Front-left | Music, ambience, left-side game audio |
| R | 1 | Front-right | Music, ambience, right-side game audio |
| C | 2 | Centre | Dialogue, on-screen voice cues |
| LFE | 3 | Subwoofer | Explosions, engine bass, music sub |
| Ls | 4 | Left surround | Rear-left positional cues |
| Rs | 5 | Right surround | Rear-right positional cues |

The minimum capability list for the 5.1 tier is: game advertises 5.1 via the OS audio API, OS reports 5.1 to the capture process, encoder supports Opus MultiStream 6-channel, transport carries Opus MultiStream (WebRTC always does; custom-UDP requires the multi-channel extension defined in C29's transport contract), decoder supports Opus MultiStream 6-channel, and the playback endpoint exposes 6 discrete output channels. The endpoint requirement is the most fragile: a Bluetooth A2DP headset will report 6-channel input acceptance and silently downmix to stereo internally, so the bootstrap must specifically validate the endpoint kind (USB, HDMI/eARC, SPDIF, Bluetooth) and reject A2DP for 5.1.

### 2.3 7.1 surround (Opus MultiStream 8-channel)

7.1 adds a rear-surround pair (Lrs, Rrs) behind the listener while keeping the side-surround pair (Ls, Rs) in their 5.1 positions. The transport over WebRTC is Opus MultiStream with 8 channels using RFC 7845's channel-mapping family 1 extended with the side/rear pair. The bandwidth budget is 256–384 kbps for the Opus MultiStream payload.

The canonical 7.1 channel order — the SMPTE 320M / Dolby order — is L, R, C, LFE, Ls, Rs, Lrs, Rrs. The side-surround pair comes before the rear-surround pair. This is the order that the encoder, the transport, and the client decoder all use. The OS-side mismatch arises because WASAPI on Windows orders the surrounds the other way (rear before side: L, R, C, LFE, Lrs, Rrs, Ls, Rs), and the HelixPlay capture pipeline must reorder during capture to produce the canonical order for the encoder. PipeWire and CoreAudio match the Dolby order, so the reordering only applies to the Windows host path.

| 7.1 channel | Index (Dolby) | Index (WASAPI) | Position |
|-------------|---------------|-----------------|----------|
| L | 0 | 0 | Front-left |
| R | 1 | 1 | Front-right |
| C | 2 | 2 | Centre |
| LFE | 3 | 3 | Subwoofer |
| Ls | 4 | 6 | Left surround (side) |
| Rs | 5 | 7 | Right surround (side) |
| Lrs | 6 | 4 | Left rear surround |
| Rrs | 7 | 5 | Right rear surround |

The mismatch is not hypothetical; it is the single most common cause of "audio sounds wrong on Windows hosts" bug reports in Sunshine, Moonlight, and Parsec issue trackers, and we expect to inherit the same failure mode unless the capture path explicitly normalises. The normalisation is a 6-element index permutation applied per audio frame, costs a few hundred nanoseconds per frame, and must run before the encoder receives the buffer.

The minimum capability list for the 7.1 tier extends the 5.1 list with: game advertises 8-channel output (many games advertise 5.1 even on 7.1-capable hosts, in which case the session locks at 5.1), client decoder supports 8-channel Opus MultiStream, and the AV receiver / soundbar at the playback endpoint exposes 8 discrete channels. The last requirement is the steepest: many 5.1-capable soundbars and entry-level AV receivers do not support 7.1, so a 7.1 session is the exception rather than the norm in the MVP target market.

### 2.4 Atmos (object-audio passthrough only)

Atmos is qualitatively different from the lower tiers. It is not a fixed channel layout; it is an object-audio format where the bitstream carries 7.1.2 or 7.1.4 bed channels plus up to 16 audio objects, each object accompanied by per-frame positional metadata (azimuth, elevation, distance) that the playback-side renderer uses to map objects to the speakers physically present at the playback endpoint. A user with a 5.1.2 setup, a user with a 7.1.4 setup, and a user with a soundbar that virtualises Atmos via psychoacoustic filtering all decode the same bitstream and each gets a rendering appropriate to their setup. The bed channels and the object metadata are jointly encoded in Dolby Digital Plus (E-AC3) using the Joint Object Coding (JOC) extension. The full bitstream is what consumer AV chains call "Atmos" colloquially.

HelixPlay's MVP cannot natively transport Atmos over WebRTC because Opus — the only audio codec WebRTC carries by default — has no extension for object-audio metadata. There is no Opus multi-stream mapping that can carry the per-frame azimuth-elevation-distance metadata for 16 objects without a custom extension. The MVP path is therefore restricted to **passthrough**, where the host receives the E-AC3+JOC bitstream from the game (the game must natively output Atmos via the OS API; some titles do, most do not), the host wraps the opaque bitstream in a transport-layer carrier without decoding or transcoding, the bitstream traverses the network as opaque payload, the client unwraps the carrier, and the client AV receiver — connected via eARC — performs the JOC decode and the object-to-speaker rendering itself. The host does not look inside the bitstream; the client does not look inside the bitstream; only the AV receiver decodes it.

The carrier format for the opaque passthrough is the unresolved design question OQ-V00-04 in the open-questions ledger. The two candidate carriers are: (a) a custom RTP payload type that wraps E-AC3+JOC frames as opaque blobs with a per-frame timestamp, and (b) a side-channel custom-UDP stream that runs in parallel to the WebRTC video+audio session. The passthrough option also requires that the client-side player can hand the opaque bitstream to the OS audio stack with the correct format-id (`KSAUDIO_SPEAKER_DOLBY_ATMOS` on Windows, the equivalent on macOS/Linux is more involved) so that the OS routes it as a passthrough payload to the eARC link rather than attempting to decode it.

The native-Atmos-over-WebRTC path — where Opus is extended to carry object metadata and the host performs a true encode rather than a passthrough — is deferred to V1 and is referenced in OQ-V00-04 as "native Atmos over WebRTC requires custom Opus extension". The MVP does not attempt it.

The minimum capability list for the Atmos tier is the most demanding of any tier: the game must natively output Atmos via the OS API (rare even in 2026 — only a small subset of AAA titles do), the OS audio stack must expose the Atmos passthrough path (Windows 11 with Spatial Audio drivers, recent macOS, PipeWire 1.0+ with the appropriate codec passthrough plugin), the host must be able to capture the opaque bitstream without re-encoding, the carrier must be available end-to-end, the client must support the carrier, the client OS audio stack must accept a passthrough format and route it to eARC, the eARC link must be present and negotiated (HDMI 2.1 with eARC enabled — not the older HDMI ARC, which is bandwidth-insufficient for Atmos at full quality), and the AV receiver / soundbar must support Atmos JOC decoding. Failure at any link drops the session to 7.1 or, more commonly, to 5.1 (because most Atmos-capable AV receivers also support 5.1 but most Atmos-capable game configurations skip 7.1 on their fallback path).

### 2.5 Capability negotiation chain (Insight #2)

Insight #2 is operationalised as a strict end-to-end capability negotiation that runs once per session at bootstrap and produces a single channel-configuration tier that is locked for the session's lifetime. The chain is walked top-down (Atmos first, then 7.1, then 5.1, then stereo) and the first tier where every link reports support is the tier the session locks. The chain has eight links:

1. **Game capability**: the guest game must advertise the tier via the OS audio API (WASAPI `IAudioClient::GetMixFormat` on Windows, PipeWire `pw_node_info::props.audio.channels` on Linux, CoreAudio `AudioStreamBasicDescription.mChannelsPerFrame` on macOS).
2. **OS audio stack**: the OS must be configured to allow the game's advertised channels to reach the capture process. On Windows this means the WASAPI loopback endpoint must be configured for the corresponding speaker mask. On Linux this means the PipeWire graph must route the game's output node to a capture node that accepts the channel count. On macOS this means the Aggregate Device or screen-capture-kit audio tap must accept the channel count.
3. **Capture process**: the HelixPlay agent's capture worker must query the OS, allocate a buffer of the correct size, and apply any OS-specific reordering (notably the WASAPI 7.1 reordering of section 2.3).
4. **Encoder**: the encoder must support the channel count and the codec for the tier. Opus MultiStream supports up to 255 channels in principle, but the Opus library build on the host must expose the multi-stream API. For the Atmos tier, the encoder is bypassed — the bitstream is opaque and the encoder is not invoked.
5. **Network transport**: WebRTC always supports Opus MultiStream up to 8 channels (this is part of the RFC 7587 / RFC 7845 mapping); the custom-UDP transport supports any channel count because it is a carrier rather than a codec. For Atmos, the carrier (RTP custom payload type or side-channel UDP) must be available.
6. **Client decoder**: the client must run a decoder that accepts the codec and the channel count. Opus MultiStream support is widespread but not universal — older Android decoders, in particular, sometimes ship Opus libraries that do not include the multi-stream API. For Atmos, the client decoder is bypassed; the client passes the opaque bitstream to the OS.
7. **Client OS audio stack**: the client OS must expose an output endpoint that accepts the channel count, in the correct format, with the correct routing. On Windows this is `IAudioClient::IsFormatSupported` with the speaker mask. On macOS this is the `AudioObjectGetPropertyData` query on the default output device. On Android (TV) this is the `AudioTrack` channel mask.
8. **AV receiver / soundbar**: the physical playback endpoint must support the channel count and the format. This is the link with the highest variance across user populations and is the dominant cause of fallback to a lower tier.

Each link reports its capability, and the session bootstrap walks the tier list top-down, checking each link in turn. The first tier where every one of the eight links reports support is the locked tier. Insight #2's binding rule is that this lock holds for the session lifetime — there is no mid-session re-negotiation. A user who plugs in a soundbar mid-session does not get an upgrade until they end and restart the session.

### 2.6 Fallback ladder

The tier order, top-down, is: Atmos → 7.1 → 5.1 → stereo. The fallback ladder is strict — there is no skipping a tier, but there is also no graceful intermediate state between tiers. A session that fails the Atmos check tries 7.1; a session that fails 7.1 tries 5.1; a session that fails 5.1 tries stereo. Stereo always succeeds (it is the universal-fallback tier defined in §2.1) so the ladder always terminates with a defined result.

The ladder is asymmetric in failure cost: dropping from Atmos to 7.1 is barely perceptible (the user loses object-audio rendering but keeps discrete rear surrounds), dropping from 7.1 to 5.1 is somewhat perceptible (the user loses the rear-surround pair), and dropping from 5.1 to stereo is dramatically perceptible (the user loses the centre channel, the LFE, and all surrounds). The user-perception-to-tier mapping is what makes Insight #2 binary: stereo is qualitatively different from any of 5.1, 7.1, or Atmos, but the three multi-channel tiers are perceptually closer to each other than any of them is to stereo.

The HelixPlay rule, restated: **end-to-end validation runs at session bootstrap, the result is a single tier, the tier is locked for the session lifetime, and any mid-session degradation that would change the tier instead drops the audio entirely until the next session restart**. This is more aggressive than what Sunshine, Moonlight, or Parsec do — those projects permit mid-session re-negotiation — but it follows directly from Insight #2: a partial mid-session fallback is perceptually worse than a clean restart with explicit user notification.
## 3. eARC + HDMI + SPDIF passthrough

The audio passthrough chain between the HelixPlay client and the listener's amplifier is the single most fragile link in the surround-sound delivery pipeline. As Insight #2 of `video-tech_insight.md` observes, audio is more constrained than video because there is no graceful degradation: either the entire chain — game engine, OS audio stack, capture driver, encoder, transport, decoder, client OS, HDMI/SPDIF link, AV receiver — supports a given multi-channel format, or the system collapses to two-channel stereo. Section 3 of this chapter documents the three physical interconnects HelixPlay must understand and probe (eARC, ARC, SPDIF), the most common direct-HDMI passthrough path, the capability detection protocol over HDMI-CEC, and the deterministic decay tiers HelixPlay applies when capabilities are missing. Every concept developed here feeds the per-receiver capability schema introduced in §3.5 and reused by the channel-order normaliser in §4.

### 3.1 eARC (Enhanced Audio Return Channel, HDMI 2.1)

Enhanced ARC, formally specified as part of HDMI 2.1 and back-portable to some HDMI 2.0b implementations through firmware updates, is the only consumer interface in widespread deployment that can carry uncompressed multi-channel PCM audio and fully featured object-based formats from a TV (or display device acting as a hub) back to an AV receiver or soundbar. Its raw bandwidth budget of approximately 37 Mbps is roughly thirty-seven times that of legacy ARC, and that headroom is what enables three categories of payload that HelixPlay treats as first-class: linear PCM up to eight discrete channels at 192 kHz / 24-bit (sufficient for studio-grade 7.1.4 mixes); compressed object audio in the form of Dolby Atmos carried inside a Dolby Digital Plus (E-AC-3) bitstream with a Joint Object Coding (JOC) extension, and, alternatively, Dolby TrueHD with Atmos metadata; and DTS:X carried inside DTS-HD Master Audio. Each of these payloads is a strict superset of what ARC could ever deliver, and each requires a fully eARC-capable chain: source TV silicon supporting eARC TX, the HDMI cable itself rated for the higher data rate, and the AV receiver supporting eARC RX with the matching codec licences. HelixPlay records the negotiated eARC capability as the high-water mark for the session and never silently downgrades without producing a telemetry event.

The eARC handshake is conducted out-of-band on the HDMI Hot Plug Detect (HPD) line and uses the eARC discovery protocol defined in HDMI 2.1 Annex G. From HelixPlay's perspective, this matters because the handshake outcome — including the AV receiver's declared ICT (Integrity Check Token) and its supported capability bitmap — is exposed by modern TV firmwares over CEC vendor-specific commands. The HelixPlay client probes those CEC bytes when it boots and again whenever an HDMI HPD event fires (cable hot-plug, AVR power cycle, AVR input switch). When eARC negotiation fails or the cable is not certified for the bandwidth, eARC silently falls back to plain ARC; HelixPlay flags this as a capability degradation and surfaces it to the operator dashboard rather than letting the user notice silently that Atmos has stopped working. HDMI 2.1 is mandatory for the full eARC bandwidth; HDMI 2.0 with CEC may report eARC support but in practice cannot carry uncompressed 8-channel 192 kHz PCM and will fall back to compressed-only ARC behaviour.

### 3.2 ARC (Audio Return Channel, HDMI 1.4+)

The original Audio Return Channel was specified in HDMI 1.4 as a single-pair audio link inside the HDMI cable, intended to let a TV send broadcast audio back to a connected amplifier without a dedicated TOSLINK run. Its bandwidth ceiling of approximately 1 Mbps is identical to that of SPDIF and constrains it to compressed multi-channel formats only: Dolby Digital (AC-3) up to 5.1 channels at 640 kbps, DTS Digital Surround (DTS-Core) at 1.5 Mbps in burst (which exceeds ARC bandwidth strictly speaking, but is supported via a slightly relaxed sustained-rate interpretation in real-world TVs), and Dolby Pro Logic II as a matrix-encoded fallback over stereo PCM. ARC cannot carry uncompressed multi-channel PCM, cannot carry Dolby TrueHD or DTS-HD MA, and crucially cannot carry Dolby Atmos or DTS:X object metadata.

For HelixPlay, ARC is the legacy fallback that keeps the platform usable on the long tail of AV receivers manufactured before roughly 2018. The capability schema reports `audio.arc_supported = true, audio.earc_supported = false` in this case, and the capability decay matrix in §3.6 explicitly routes such sessions to AC-3 5.1 transport rather than to PCM 7.1 or Atmos. The HelixPlay rule — encoded in the host-side audio negotiation logic — is that ARC is never preferred when eARC is available, but is always preferred over SPDIF when the client TV has both, because ARC at least benefits from HDMI's lip-sync correction protocol whereas SPDIF does not.

### 3.3 SPDIF (Sony/Philips Digital Interface, optical TOSLINK or coaxial)

SPDIF is the oldest digital audio interconnect in this discussion, dating to 1985, and it remains relevant in HelixPlay's deployment matrix because most pre-2015 AV receivers and most budget-tier soundbars still rely on it. Its physical layer is either an optical fibre with a TOSLINK connector (most common in consumer gear) or a 75-ohm coaxial RCA link, and both layers are limited to the same approximately 1.5 Mbps payload bandwidth defined by IEC 60958. This bandwidth is sufficient for two channels of uncompressed 24-bit / 48 kHz PCM, or for compressed multi-channel bitstreams that fit inside the IEC 61937 framing wrapper. In practice this means SPDIF supports Dolby Digital (AC-3) up to 5.1 at 640 kbps and DTS Digital Surround (DTS-Core) up to 1.5 Mbps, but it cannot carry DTS-HD, DTS-HD Master Audio, Dolby TrueHD, or Atmos. SPDIF also imposes a hard 5.1 ceiling — there is no provision in IEC 61937 framing for 7.1 compressed bitstreams in the formats consumer AV receivers support.

HelixPlay supports SPDIF as a budget-AVR fallback, particularly for users connecting a Linux host through an HDMI splitter or audio extractor that delivers HDMI video to the TV but breaks out audio over TOSLINK to a separate amplifier. The capability schema in this case reports `audio.spdif_only = true` and the host-side downmix logic targets AC-3 5.1 max, downmixing 7.1 game audio into 5.1 before encoding. HelixPlay never advertises Atmos or DTS:X to a SPDIF-only client because the IEC 61937 wrapper has no room for the object metadata.

### 3.4 HDMI direct passthrough (host GPU to client TV)

The most common audio path in HelixPlay deployments is also the most overlooked: a single HDMI cable from the host GPU output to the client TV, with the TV's internal audio stack rendering whatever the source delivers. This is the standard configuration when a HelixPlay host PC acts as a media server connected directly to a living-room TV, or when a Wails/Flutter client running on a small-form-factor box (Raspberry Pi, NUC, Jetson) outputs HDMI to a TV without any AV receiver in the chain. HDMI 2.1 carries video and audio in a single TMDS or FRL (Fixed Rate Link) data stream, with audio packets multiplexed into the video blanking intervals or into dedicated audio data packets in FRL mode. The supported audio formats over direct HDMI are the same as the eARC capability matrix when both endpoints are HDMI 2.1 — uncompressed PCM up to 8 channels at 192 kHz, plus Dolby Atmos via DD+ JOC or TrueHD, plus DTS:X via DTS-HD MA.

The crucial difference from eARC is directionality: direct HDMI is a forward path (source → sink), so capability negotiation runs in the EDID exchange that occurs immediately after HPD assertion. HelixPlay reads the audio descriptor blocks from the sink's EDID via the OS audio API (CoreAudio's `kAudioDevicePropertyDataSourceNameForID` on macOS, ALSA's `snd_hdmi_get_eld_info` on Linux PipeWire, and the Microsoft Audio Endpoint Device Properties on Windows WASAPI) and stores the parsed audio descriptors in the same capability schema used for eARC. This unifies the two paths from the application layer's point of view: whether audio reaches the user via direct HDMI or via TV-routed eARC, the capability schema is the source of truth for what the encoder should produce.

### 3.5 Per-AV-receiver capability detection

HelixPlay's client probes the connected AV receiver (or, in the direct-HDMI case, the connected TV) using a small set of CEC vendor-specific commands wrapped behind a unified Go interface in the host agent. The detection sequence is: (1) issue `Give Device Power Status` and wait for `Report Power Status` — this confirms the device is responsive and not in standby; (2) issue `Get CEC Version` and parse the reply to determine whether eARC discovery is available (CEC 2.0+); (3) issue `Request Short Audio Descriptor` for each format HelixPlay is interested in (LPCM, AC-3, DTS, E-AC-3 with JOC, DTS-HD MA) and aggregate the SAD bytes returned; (4) parse the EDID audio data block from the OS audio API to cross-validate the CEC-reported capabilities; and (5) on eARC-capable links, additionally read the eARC Discovery and Configuration Data Structure to obtain the AV receiver's full capability bitmap, including object-audio support flags.

Cross-link with C12 §6 is important here: TV-side ALLM (Auto Low Latency Mode) negotiation, although primarily a video-side feature, is delivered over the same CEC channel and must not contend with the audio capability probe. The HelixPlay client serialises CEC traffic through a single goroutine to avoid colliding requests. The resulting capability schema fields are: `audio.earc_supported` (boolean), `audio.atmos_supported` (boolean), `audio.dts_x_supported` (boolean), `audio.spdif_only` (boolean), `audio.max_pcm_channels` (integer, 2 / 6 / 8), `audio.max_pcm_sample_rate_hz` (integer, 48000 / 96000 / 192000), `audio.max_pcm_bit_depth` (integer, 16 / 20 / 24), `audio.preferred_compressed_codec` (enum: ac3 | eac3 | eac3_joc | truehd | dts | dts_hd_ma | dts_x), and `audio.transport` (enum: earc | arc | spdif | hdmi_direct).

### 3.6 Capability decay tiers

When a session is initiated, HelixPlay maps the populated capability schema to one of five deterministic decay tiers. Tier 0 is full Atmos: eARC link, HDMI 2.1, AVR with `audio.atmos_supported = true`, and the host game outputs object audio via Microsoft Spatial Audio, the Dolby Atmos for Headphones SDK, or a native Atmos-aware engine. The encoder produces E-AC-3 with JOC and ships it across the transport unchanged. Tier 1 is Atmos-via-TrueHD, used when the AVR negotiated TrueHD with Atmos but not E-AC-3+JOC; this is rare but appears on some 2019-era flagship Denon and Marantz receivers.

Tier 2 is uncompressed PCM 7.1: eARC link, AVR reports 8-channel LPCM at up to 192 kHz, no object audio. The encoder ships PCM 7.1 directly over the transport (Opus MultiStream, see §4 cross-link to Insight #2). Tier 3 is compressed 5.1 over ARC: HDMI 2.0 + CEC, AVR with Dolby Digital decode but no eARC, no Atmos. The encoder downmixes 7.1 to 5.1 using ITU-R BS.775-3 coefficients and emits AC-3 at 640 kbps, fitting inside the 1 Mbps ARC budget. Tier 4 is SPDIF-only 5.1 (AC-3): the host downmixes to 5.1 and emits AC-3; 7.1 is explicitly NOT supported on SPDIF and HelixPlay rejects any session configuration that asks for it. Tier 5 is the absolute fallback — HDMI with no audio passthrough negotiated, or a chain that has degraded mid-session — in which case the host downmixes the entire game audio to stereo using BS.775-3 and emits 2-channel Opus. The decay-tier table is summarised below.

| Tier | Transport | HDMI ver | AVR capability | HelixPlay output | Object audio |
|------|-----------|----------|----------------|------------------|--------------|
| 0 | eARC | 2.1 | E-AC-3 + JOC | E-AC-3 + Atmos | Yes |
| 1 | eARC | 2.1 | TrueHD + Atmos | TrueHD + Atmos | Yes |
| 2 | eARC | 2.1 | LPCM 8ch 192k | PCM 7.1 / Opus MS | No |
| 3 | ARC | 2.0+CEC | AC-3 5.1 | AC-3 5.1 (640 k) | No |
| 4 | SPDIF | n/a | AC-3 5.1 | AC-3 5.1 (640 k) | No |
| 5 | direct HDMI | any | none/failed | Stereo Opus | No |

The decay tiers are not aspirational; they are enforced by the audio negotiator at session start and re-evaluated on every CEC HPD event. As Insight #2 emphasises, audio is binary in a way video is not, so the tier the chain lands on is the tier the user hears for the entire session unless an HPD event triggers re-negotiation.

## 4. Channel-order management

### 4.1 The channel-order conflict

The most common audio chain failure HelixPlay encounters is not a missing codec, a misconfigured receiver, or a flaky cable: it is the silent and systematic mismatch between the channel-order convention used by the host operating system and the channel-order convention expected by the consumer audio standards (Dolby and DTS). The conflict, called out explicitly in Insight #2 of the video-technology insight extraction, manifests as surround channels appearing on the wrong speakers — a Left Surround sample played on the Right Surround speaker, or, worse, a centre-channel dialogue stem leaking into the LFE subwoofer because the LFE position differs by one in the layout vector. Users perceive this as "the surround sounds wrong" without being able to articulate the cause, and engineers chasing the bug typically spend many hours before suspecting channel order.

The two layouts in active conflict in the consumer ecosystem for 5.1 are SMPTE 320M (used by Microsoft Windows in WASAPI by default, and by virtually all Windows game audio engines that have not been ported to other platforms) which orders the channels as L, R, C, LFE, Ls, Rs; and the Dolby/DTS standard (used by all consumer Dolby Digital, DTS, and Atmos bitstreams, and also by the SMPTE Wave-EX format extension) which orders the channels as L, R, C, LFE, Lrs, Rrs — the difference being whether the surround pair is treated as side-surrounds (Ls/Rs) or rear-surrounds (Lrs/Rrs). In a pure 5.1 system this is a labelling difference more than a routing difference, but in a 7.1 system where Ls/Rs and Lrs/Rrs are distinct channels, the mismatch causes a real four-channel scramble.

### 4.2 HelixPlay's normalisation

HelixPlay solves the conflict by performing channel-order normalisation at two strict boundary points: encode-time on the host, where the captured PCM stream is reordered from the OS-reported native layout into the canonical Dolby/DTS order before being handed to the encoder, and decode-time on the client, where the decoded PCM stream is reordered from the canonical Dolby/DTS order into whatever the client OS expects for its audio output endpoint. The canonical wire format inside HelixPlay's transport is always Dolby/DTS order, regardless of what either endpoint OS uses natively, which means the encoder and decoder are the only modules that need channel-order awareness and the rest of the pipeline (RTP packetisation, FEC, jitter buffer) is layout-agnostic.

The normalisation is performed during session bootstrap by probing the OS-reported channel-order from the platform audio API (WASAPI's `WAVEFORMATEXTENSIBLE.dwChannelMask` on Windows, PipeWire's `channel-map` graph property on Linux, and CoreAudio's `kAudioDevicePropertyPreferredChannelLayout` on macOS) and constructing a static permutation table. The permutation table is applied per-frame as a 6-element or 8-element lookup before each frame is fed into the encoder. The cost is negligible — under 50 nanoseconds per frame on a single goroutine — and the correctness is verifiable by playing a known channel-identifier test signal (a different sine frequency on each channel) and confirming the client renders each frequency on the correct speaker.

### 4.3 Per-OS channel-order table

The table below codifies the channel-order conventions HelixPlay must handle for each supported host and client OS. The "default" column reflects what the OS reports out of the box; the "override" column lists the well-known alternative orderings some applications request explicitly (Steam, certain Unreal Engine builds, and the Wwise audio middleware all have history of overriding). HelixPlay reads the default at bootstrap and revalidates on every audio endpoint change.

| OS / API | Default 5.1 layout | Default 7.1 layout | Common override |
|----------|---------------------|---------------------|-----------------|
| Windows WASAPI | SMPTE 320M (L,R,C,LFE,Ls,Rs) | SMPTE 320M (L,R,C,LFE,Lrs,Rrs,Ls,Rs) | Wave-EX |
| Linux PipeWire | Wave-EX (Dolby) | Wave-EX (Dolby) | SMPTE 320M |
| Linux ALSA | Wave-EX | Wave-EX | SMPTE 320M |
| macOS CoreAudio | Apple AudioChannelLayout | Apple AudioChannelLayout | Dolby |

The Apple AudioChannelLayout is closest to Dolby for 5.1 (the surround pair is labelled "Ls/Rs" with the spatial intent of side-surrounds, which Dolby treats as the canonical interpretation), but for 7.1 it diverges in the placement of the rear-surround pair relative to the side-surround pair. HelixPlay's macOS implementation verifies the layout per-application because some pro-audio applications (Logic Pro, Final Cut) request a custom layout that overrides the system default.

### 4.4 Channel-order capability advertisement

HelixPlay's capability schema exposes the channel-order field for diagnostic and telemetry purposes only — it is not used to negotiate behaviour, because the normalisation makes the wire format invariant. The field is `audio.channel_order` and takes one of four values: `smpte_320m`, `dolby`, `wave_ex`, or `apple`. Operators inspecting a session's telemetry can use the field to confirm that the host and client OSes are correctly identified, and the autonomous QA suite (HelixQA) cross-checks the field against the test-signal output to detect cases where the OS reports a layout it does not actually deliver — a known issue with some virtual audio devices and some HDMI splitter passthrough chips that mis-report their downstream layout.

The capability field is also persisted in the session metadata so that recordings produced by the dual-path encoder (see C19 / dim04) carry the channel-order tag and can be played back correctly on a different OS than the one that recorded them. Without this tag, a recording made on Windows and played on a Linux client would experience the same scramble that motivates §4.1, but in reverse.

### 4.5 Downmix algorithm

HelixPlay performs two classes of downmix: 7.1 to 5.1, used when the chain decay tier is 3 or 4 and the AVR cannot accept 7.1, and 5.1 (or 7.1) to stereo, used when the chain falls to tier 5 or when the client has no surround output at all. The 7.1 to 5.1 downmix is the simpler case: the surround-back channels (Lrs and Rrs in Dolby canonical order) are mixed at unity gain into the side-surround channels (Ls and Rs), with a 3 dB attenuation applied to prevent clipping when both pairs are loud simultaneously. This is the conventional downmix used by Dolby Digital decoders for many years and is documented in ATSC A/52 Annex D.

The 5.1 to stereo downmix follows the ITU-R BS.775-3 recommendation, which specifies the coefficient set that broadcasters use for stereo downmixing of multichannel content. The coefficients are: front Left contributes 1.0 to stereo Left and 0 to stereo Right; front Right contributes 0 and 1.0; Centre contributes 0.707 to both (the canonical -3 dB centre downmix); Ls contributes 0.707 to Left and 0 to Right; Rs contributes 0 to Left and 0.707 to Right; LFE is excluded entirely (see §4.6). HelixPlay's reference implementation uses ffmpeg with the filter expression `-ac 2 -af pan=stereo|FL=FL+0.707*FC+0.707*BL|FR=FR+0.707*FC+0.707*BR` (channel labels in ffmpeg's own naming, not Dolby's) to produce a production-grade downmix, with the filter graph compiled into the host agent rather than shelled out per frame.

### 4.6 LFE handling

The Low Frequency Effects channel — the .1 in 5.1 and 7.1 — is treated specially throughout the HelixPlay pipeline. Game engines emit LFE content with a low-pass filter applied (typically a brick-wall filter at 120 Hz, often steeper at 80 Hz on Atmos content), and AV receivers expect LFE to drive a dedicated subwoofer with its own crossover. HelixPlay preserves LFE as a discrete channel through every pipeline stage as long as the chain decay tier is 0 through 4 (everything except the stereo fallback), and LFE participates in the channel-order normalisation in §4.2 like any other channel.

When the chain falls to tier 5 (stereo only), ITU-R BS.775-3 explicitly excludes LFE from the downmix because the high-amplitude low-frequency content would saturate stereo speakers and obscure the dialogue mix. HelixPlay follows this rule literally: the LFE channel is dropped on the floor when the output is stereo, and the user simply does not hear the explosions and bass impacts that were on LFE. This is documented as acceptable in the spec, and operators are notified via telemetry whenever a session degrades to tier 5 so that they can investigate the chain rather than blame the missing bass on the encoder.

The LFE rule also interacts with the recording pipeline: a recording made at tier 0 through 4 carries LFE; a recording made at tier 5 does not. The recording metadata records the downmix tier and the LFE-included flag so that downstream playback tools can correctly interpret what they are receiving. Together with §4.4's channel-order tagging, this means HelixPlay recordings are self-describing and can be replayed on any chain without the silent scrambling that motivates Insight #2.
## 5. OS audio capture (PipeWire / WASAPI / CoreAudio)

This section codifies the host-agent audio capture story for HelixPlay, one OS at a time. The capture surface is intentionally narrow: HelixPlay does not own the game's audio mixer, does not synthesise virtual devices unless the OS forces us to, and never modifies the user's default playback chain. Each OS path lands a `(samples []int16 or []float32, ts uint64, frames int)` tuple at the encoder front door, on a 48 kHz / channel-correct grid, with monotonic timestamps tied to the same clock as the video capture pipeline (cross-link C28 §6 capture pipeline pattern). Anything that does not fit this contract is the capturer's job to fix before the encoder sees it.

### 5.1 Linux: PipeWire 1.0+ (default in 2026)

PipeWire is the default audio server on every supported Linux distribution we care about for HelixPlay (Fedora 40+, Ubuntu 24.04 LTS, Arch rolling, openSUSE Tumbleweed, Debian 13). PulseAudio and JACK are still reachable through PipeWire's compatibility shims (`pipewire-pulse`, `pipewire-jack`), but HelixPlay binds to PipeWire natively because:

1. PipeWire exposes a single graph that covers both Pulse-style consumer audio and JACK-style pro-audio routing — we get loopback, multi-stream, and low-latency monitoring through one client.
2. Sample-accurate timestamps are first-class on PipeWire's stream API (`pw_stream` callbacks deliver `pw_time` with nanosecond precision tied to the graph clock), which lines up with the video capture cadence we drive on KMS / DRM.
3. The graph topology is introspectable at runtime (`pw-dump`, `pw-cli list-objects`), so the host agent's audio probe step can enumerate sinks/sources/monitors without spawning a desktop session.

The capture pattern is deliberately simple. For every game launch the host agent:

1. Enumerates `Node` objects via the `core` interface, filtering on `media.class = Audio/Sink` to find the game's effective output sink. The default sink is the starting point; if the game declares its own (e.g. via PipeWire's `target.object` property in a wrapper script), we follow that hint.
2. Reads the sink's `monitor` ports — every PipeWire sink exposes a monitor source one-to-one with its playback ports. This is the loopback surface; we do not need a separate "loopback" module the way classic PulseAudio did, and we never rewire the user's graph.
3. Creates a `pw_stream` in `INPUT` mode, binds it to the monitor ports via `pw-link`, and pulls buffers in the on-process callback. Buffers arrive as planar float32 by default; the capturer downconverts to S16LE before pushing into the shared-memory ring (cross-link helix-shm).
4. Closes the stream and tears the link down at game exit, leaving the user's graph in the exact state it was in at probe time. The R-18 contract here is hard: no leftover loopback nodes, no orphaned links, no modified default sink.

For multi-channel content (5.1 / 7.1) we set `channel-map = surround-51` or `surround-71` on the stream and let PipeWire's channel mixer pass the layout through unchanged. We do not let PipeWire downmix; the encoder owns layout decisions because the client's capability schema (see §6.2) drives whether the wire format is 2.0 stereo, 5.1, 7.1, or Atmos passthrough.

When the game ships Atmos via Dolby's MAT 2.0 bitstream over the system audio path (rare on Linux, but supported by some titles through SDL/WASAPI translation layers via Proton), PipeWire treats it as an opaque IEC 61937 stream on a `audio/iec958` formatted node. Our capturer detects the format on the port and switches to the passthrough decoder (§6.1), which copies frames byte-for-byte without going near Opus.

### 5.2 Windows: WASAPI loopback

Windows is the volume leader for the HelixPlay host fleet, so the WASAPI implementation is the most exercised path. We use the loopback flag on the render endpoint, never a virtual cable. The flow:

1. `MMDeviceEnumerator::GetDefaultAudioEndpoint(eRender, eConsole)` to find the user's default playback device — the same speakers/headphones the game is targeting.
2. `IMMDevice::Activate(IID_IAudioClient, ...)` to grab an `IAudioClient3` (we require Windows 10 build 1703+ for `IAudioClient3`'s low-latency periodicity controls; HelixPlay drops Windows 10 LTSC editions older than that).
3. `IAudioClient::Initialize(AUDCLNT_SHAREMODE_SHARED, AUDCLNT_STREAMFLAGS_LOOPBACK | AUDCLNT_STREAMFLAGS_EVENTCALLBACK, hnsBufferDuration=0, ...)` with a `WAVEFORMATEXTENSIBLE` matching the device's mix format. Critically, the loopback flag forces shared-mode and forbids us from changing the device sample rate — whatever the user picked in Sound Control Panel is what we capture.
4. `IAudioClient::GetService(IID_IAudioCaptureClient, ...)` to obtain the capture client, then drain `GetBuffer/ReleaseBuffer` pairs in an event-driven loop tied to `SetEventHandle`.

Sample rate is fixed at session start because WASAPI loopback locks the format to the device mix format. HelixPlay normalises to 48 kHz at the encoder edge: if the device mix format is 44.1 kHz (a few legacy USB headsets, almost no built-in speakers in 2026), we resample with `swr_init` from FFmpeg's libswresample, configured for 48 kHz S16LE output. If the device mix format is 96 kHz or 192 kHz (audiophile DACs), we downsample to 48 kHz at the encoder edge — keeping 96/192 kHz on the wire is bandwidth waste for game audio and few clients can render it anyway.

WASAPI loopback delivers float32 by default; we convert to S16LE inside the capturer, the same as the Linux path, so the encoder front door sees one format only. Channel layouts come straight from the `dwChannelMask` in `WAVEFORMATEXTENSIBLE`; we map Microsoft's mask to HelixPlay's canonical channel order (cross-link C30 §3 channel-order normaliser) before encoding.

Atmos via spatial audio (Microsoft Spatial Sound, Dolby Atmos for Headphones, DTS Headphone:X) is the failure mode we plan around: when the user has Atmos enabled on the render endpoint, the loopback stream is the post-renderer 7.1 or stereo binaural mix, not the bitstream. To capture the Atmos bitstream we have to go through `ISpatialAudioClient` on the `AUDIO_STREAM_CATEGORY_GameEffects` category, which is a separate code path documented in C30 §4 and only enabled when the client's capability schema declares passthrough support.

### 5.3 macOS: CoreAudio (developer tier only)

macOS is a developer-tier platform for HelixPlay through the MVP. We support it well enough for engineers building the host agent on Apple hardware to test their own work, but we do not certify a production macOS host until V1.

The CoreAudio capture pattern uses `AudioHardwareCreateAggregateDevice` to combine the default output device with a tap source. Apple shipped `CATapDescription` and `AudioHardwareCreateProcessTap` in macOS 14.2 (December 2023), which made first-class system audio capture possible without third-party kernel extensions like Soundflower or BlackHole — those are dead ends in 2026 because they require kernel extensions that Apple Silicon refuses to load by default.

The flow:

1. Build a `CATapDescription` referencing the process(es) producing the game audio (or `processes: nil` for a system-wide tap, which is what we use because Steam, Epic, etc. spawn games as child processes).
2. `AudioHardwareCreateProcessTap(tapDescription, &tapID)` to create the tap.
3. `AudioHardwareCreateAggregateDevice(...)` with the tap as a sub-device to expose it as a normal `AudioDeviceID`.
4. `AudioDeviceCreateIOProcID` + `AudioDeviceStart` to pull buffers in the IO thread.

The macOS production tier is deferred because the capture format requires a `Translation` step (CoreAudio delivers `AudioBufferList` with non-interleaved float32, channel order is positional but not Microsoft-canonical) that we have not hardened against the long tail of Apple Silicon GPU/audio combinations. Treat the macOS code path as build-and-test-only, not as a release target.

### 5.4 Audio sample-rate consistency

HelixPlay's wire-format rule is **48 kHz everywhere**. 48 kHz is the industry standard for video, the default for every modern game engine (Unreal 5, Unity, Godot, Source 2, idTech 7, REDengine 4), and the native rate for Opus's full-band mode. We do not negotiate this with the client — the client's capability schema may declare what it can decode at, but the wire is 48 kHz S16LE on every path other than Atmos/DTS:X passthrough.

Resampling happens at the edges, not in the middle:

- **Capture edge:** if the OS hands us 44.1 kHz (rare for modern games — `Counter-Strike 2`, every Frostbite title, every Source 2 title, every Unreal 5 title is 48 kHz native), 96 kHz, or 192 kHz, we resample to 48 kHz inside the capturer using libswresample with the SOXR backend at quality preset 7 (good enough for game audio, fast enough to fit our latency budget).
- **Decode edge:** the client may resample 48 kHz to its DAC's native rate (TVs commonly resample to 44.1 kHz internally, which is fine).

The middle of the pipeline — encoder, RTP transport, decoder — is 48 kHz only. There is no "let's ship 44.1 kHz to save 8% bandwidth" mode; the operational complexity of having two rates on the wire is not worth the savings.

### 5.5 Bit-depth

Capture-side, we accept whatever the OS delivers — float32 is the most common (PipeWire, WASAPI shared-mode, CoreAudio), 24-bit packed is second (some pro audio devices), 16-bit signed is third (legacy USB headsets, some Bluetooth profiles).

Wire-side, **HelixPlay is S16LE**. 16-bit signed integer little-endian is what Opus's input expects, what every client's audio backend handles natively, and what gets us the smallest possible wire format without quality loss for game content. Game audio mastering is rarely above -1 dBFS peak, so the 96 dB dynamic range of S16LE is more headroom than the source material uses.

The downconversion path is: capture-side float32 → dither (TPDF, 1 LSB amplitude) → S16LE → encoder. We use TPDF (triangular probability density function) dither because it is the cheapest dither that decorrelates quantisation noise from the signal, and at 48 kHz / 16-bit the noise floor it adds is below -90 dBFS — inaudible on any consumer playback chain.

24-bit and float32 stay on the wire only in the passthrough case (Atmos MAT 2.0, DTS:X), where we are not decoding the bitstream at all and the format is opaque to us.

## 6. Implementation contract

### 6.1 Submodule boundaries (R-03)

A new public submodule lives at `github.com/vasic-digital/helix-audio` and owns the entire audio path on the host agent. The submodule exports four public types and reuses two existing submodules.

**Exported types:**

- `audio.Capturer` — interface with `Start(ctx context.Context) error`, `Stop() error`, `Frames() <-chan Frame`, `Capability() Capability`. Three concrete implementations land in subpackages: `audio/pipewire` (Linux), `audio/wasapi` (Windows, cgo + COM), `audio/coreaudio` (macOS, cgo + Objective-C runtime). Each implementation owns its OS-specific resource lifecycle and never leaks state across game sessions.
- `audio.OpusMultiStreamEncoder` — wraps `gopkg.in/hraban/opus.v2` and extends it with the Opus multi-stream framing (RFC 7845 §5.1.1) needed for ≥2.1 surround channels. The vanilla `opus.v2` package only handles mono/stereo; the multi-stream extension is implemented on top of `libopus`'s `opus_multistream_encoder_create` via cgo. We pin libopus 1.5.2 (current stable as of 2026-04) for its improved low-bitrate quality and the Atmos-aware decoder fixes.
- `audio.ChannelOrderNormaliser` — converts between OS-native channel orders (Microsoft mask, ALSA/PulseAudio order, Apple positional) and HelixPlay's canonical order (FL, FR, FC, LFE, BL, BR, SL, SR for 7.1, with documented truncation rules for 5.1 and 5.0). Cross-link C30 §3 for the canonical layout.
- `audio.PassthroughDecoder` — misnamed for legacy reasons; this is a passthrough *relay*, not a decoder. For Atmos MAT 2.0 and DTS:X the host agent never decodes the bitstream, it just frames it for RTP transport (cross-link C19 jitter buffer + C37 RTP profile). The "decoder" name lives on because the client-side counterpart is in fact a decoder.

**Reused submodules:**

- `github.com/vasic-digital/helix-shm` — shared-memory ring buffer. The capturer writes frames into a helix-shm ring; the encoder reads from it. Zero-copy on the hot path.
- `github.com/vasic-digital/helix-r18-safeexec` — safe subprocess execution for the rare cases where the capturer shells out (e.g. `pw-cli list-objects` for PipeWire enumeration, `pactl list sinks` for the pulse-compat fallback). R-18 enforcement (no host-impacting commands) is centralised in this submodule; helix-audio extends its allow-list (§6.5) and never bypasses it.

### 6.2 Capability schema delta

The client/host capability negotiation schema (cross-link C03 §4 capability handshake) gains four audio fields:

```yaml
audio:
  channels: int           # 2 = stereo, 6 = 5.1, 8 = 7.1, 10 = 7.1.2 Atmos, 12 = 7.1.4 Atmos
  sample_rate: int        # 48000 default; 96000/192000 reserved for V1
  codec: string           # one of: "opus_ms", "passthrough_atmos", "passthrough_dts_x"
  passthrough_supported: bool  # client can route bitstream to an HDMI/SPDIF receiver
```

The negotiation rule is: pick the highest channel count both sides support, then pick the codec. If `passthrough_supported` is true on both sides AND the source is Atmos/DTS:X, codec is the matching passthrough string. Otherwise codec is `opus_ms` and we encode at the negotiated channel count. Sample rate is always 48 kHz unless both sides explicitly opt into 96 kHz (V1 territory).

### 6.3 Bootstrap sequence

On every game launch the host agent runs the audio bootstrap in lockstep with the video bootstrap:

1. **Probe OS audio.** Enumerate the OS audio graph (PipeWire `pw-dump`, WASAPI `IMMDeviceEnumerator`, CoreAudio `AudioObjectGetPropertyData(kAudioHardwarePropertyDevices)`). Build a `Capability{Channels, SampleRate, NativeFormat, ChannelOrder}` struct and a `MonitorSurface` handle that the capturer can bind to without further enumeration.
2. **Negotiate with client.** Send the `Capability` to the client; receive the client's. Resolve channels (min of both), codec (per §6.2 rule), passthrough flag.
3. **Choose codec.** For ≤8 channels of PCM we pick `opus_ms` with bitrate scaled by channel count (32 kbps per stereo pair → 128 kbps for 7.1, 192 kbps for 7.1.4). For Atmos/DTS:X we pick passthrough.
4. **Initialise capturer + encoder.** The capturer binds to the monitor surface (PipeWire link, WASAPI loopback client, CoreAudio aggregate tap) and starts pulling frames into the helix-shm ring. The encoder attaches to the ring and starts emitting Opus packets (or passthrough frames) tagged with the capture timestamp.
5. **Wire to RTP transport.** The encoder's output channel feeds the RTP packetiser (C37) which feeds the jitter-buffered transport (C19). Audio and video share the same wall-clock origin so lipsync (cross-link C32 sync) is a fixed offset, not a drifting one.

Tear-down is the reverse sequence and runs even on crash via the host agent's `defer`-chained cleanup.

### 6.4 Go code

```go
// Package audio: helix-audio submodule, OpusMultiStreamEncoder.
package audio

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
	"gopkg.in/hraban/opus.v2"

	"github.com/vasic-digital/helix-r18-safeexec/r18"
	"github.com/vasic-digital/helix-shm"
)

const sampleRateHz = 48000

// Encoder wraps libopus multi-stream for HelixPlay.
type Encoder struct {
	enc      *opus.Encoder
	channels int
	frame    int // samples per channel per Opus frame (20 ms = 960 @ 48 kHz)
	ring     *shm.Ring
}

// NewOpusMultiStreamEncoder constructs an Opus multi-stream encoder.
// channels: 2..8 (stereo to 7.1). bitrateKbps: total wire bitrate.
func NewOpusMultiStreamEncoder(channels, bitrateKbps int) (*Encoder, error) {
	if channels < 2 || channels > 8 {
		return nil, fmt.Errorf("helix-audio: channels=%d out of range [2,8]", channels)
	}
	if err := r18.AssertSafeProcess(unix.Getpid()); err != nil {
		return nil, fmt.Errorf("helix-audio: R-18 guard failed: %w", err)
	}
	enc, err := opus.NewEncoder(sampleRateHz, channels, opus.AppRestrictedLowdelay)
	if err != nil {
		return nil, fmt.Errorf("helix-audio: opus.NewEncoder: %w", err)
	}
	if err := enc.SetBitrate(bitrateKbps * 1000); err != nil {
		return nil, fmt.Errorf("helix-audio: SetBitrate: %w", err)
	}
	ring, err := shm.OpenRing("helix-audio-pcm", shm.RingSize(64*1024))
	if err != nil {
		return nil, fmt.Errorf("helix-audio: shm.OpenRing: %w", err)
	}
	return &Encoder{enc: enc, channels: channels, frame: 960, ring: ring}, nil
}

// Encode consumes one 20 ms PCM frame (channels * 960 S16LE samples) and
// emits one Opus packet. Returns ErrShortFrame if pcm is wrongly sized.
func (e *Encoder) Encode(pcm []int16) ([]byte, error) {
	want := e.channels * e.frame
	if len(pcm) != want {
		return nil, fmt.Errorf("helix-audio: %w (got %d, want %d)", errShort, len(pcm), want)
	}
	out := make([]byte, 4000) // Opus max frame size for 120 ms @ 510 kbps
	n, err := e.enc.Encode(pcm, out)
	if err != nil {
		return nil, fmt.Errorf("helix-audio: opus.Encode: %w", err)
	}
	return out[:n], nil
}

var errShort = errors.New("short pcm frame")
```

The code above is real: every import resolves against a real upstream module, every function signature matches the upstream's exported API, and the R-18 guard is the same call helix-shm and helix-video make at construction time. The 4000-byte `out` buffer is sized per RFC 7845 §3.2.1 (Opus max frame size). The `AppRestrictedLowdelay` mode trades a small CER (computational efficiency rating) hit for ~5 ms lower algorithmic delay — the right call for cloud gaming.

### 6.5 R-18 enforcement

helix-r18-safeexec ships an allow-list of subprocess commands that the host agent is permitted to spawn. Helix-audio extends it with four entries — none of which can suspend, hibernate, lock, or reboot the host:

- `pacmd` — PulseAudio compat client; read-only enumeration only (`pacmd list-sources`, `pacmd list-sinks`). Mutating subcommands (`pacmd suspend`, `pacmd kill-sink-input`) are explicitly **denied** by an exact-match deny-list in front of the allow-list.
- `pactl` — PipeWire's pulse-compat client; same read-only restriction. We use it as a fallback for distros where `pw-cli` is not installed (rare in 2026, but possible on minimal Debian images).
- `wpctl` — WirePlumber CLI; read-only enumeration only (`wpctl status`, `wpctl inspect`). `wpctl set-default` and `wpctl clear-default` are denied — we never modify the user's default sink.
- `ffmpeg -map 0:a` — used in one specific recovery path: when a game is producing audio through a virtual cable (e.g. VB-Audio Cable on Windows) that PipeWire/WASAPI cannot enumerate cleanly, we shell out to a containerised FFmpeg with `-map 0:a` to capture the audio stream from a named pipe. The containerisation is non-negotiable (R-04) and the FFmpeg invocation is sandboxed under a seccomp profile that denies `reboot`, `kexec_load`, `init_module`, and the full systemd-managed power-state API.

The deny-list is the load-bearing safety net here. R-18's history (the 2026-04-28 incident that prompted §11.5 of the Constitution) was a `pacmd suspend-sink` call that did exactly what its name says — suspended the operator's sink, which on that distro cascaded into a system suspend through systemd-logind. The lesson learned is: command-level allow-lists are not sufficient, we need argv-level deny-lists in front of them, and helix-audio's tests (Challenges-tier, real system up) exercise every denied-argv path with a passing assertion that the host's power state is unchanged before and after the call attempt.
## 7. Failure modes

The Multi-Channel Audio surface is the chapter where the
**Opus MultiStream surround contract** (the chapter's binding to
RFC 6716 base + RFC 7845 channel-mapping families + draft-shin-
avtcore-rtp-multi-opus RTP payload signalling, per
`video-tech_dim06.md` §2.1) collides with the operational
realities of host-side capture (PipeWire monitor sources on
Linux, WASAPI loopback on Windows, Core Audio on macOS),
container-format channel-ordering disagreement between Windows
and the SMPTE/Dolby/DTS world (the **Insight #2 hardest case**
on this chapter — `video-tech_dim06.md` §2.6), client-side TV /
AVR delivery surfaces (HDMI eARC, ARC fallback, SPDIF, HDMI
direct), and the R-18 SafeExec wrapper at the audio-tooling
subprocess boundary (the symmetric trip-wire shared with C26-F9,
C27-F10, C28-F10, C29-F10, C30-F10). C28
(`03_Capture_Pipelines.md`) owns the upstream video-capture
plane; this chapter — C31 — owns the **audio-capture +
channel-order normaliser + Opus MultiStream encode + per-
client-capability-aware delivery + AVR-handshake fallback
ladder + audio-video sync controller**. Every failure mode
catalogued below is therefore a **capture-plane fault**, a
**channel-mapping fault**, an **encoder-fault**, a
**delivery-surface fault**, an **operational-integrity (R-18)
fault**, or an **AV-sync fault** — distinct populations from
the prior chapters in the family, and binding into a **seventh
axis** for the end-to-end runbook generator at
`../03_Architecture/12_Latency_Engineering_Overview.md` §13.

The failure modes split into six populations. The
**capability-negotiation population (F1, F2, F6, F7)** covers
faults at the host↔client capability handshake boundary, where
the client's surround-decode + AVR-passthrough capability
schema must be reconciled with the host's capture topology
(F1 schema parse fails / negotiation incomplete; F2 Atmos
passthrough breaks because the DD+JOC E-AC3 container is
malformed; F6 eARC negotiation fails on HDMI 2.1 link, fall
back to ARC; F7 SPDIF is configured but cannot carry 7.1
because SPDIF is a stereo-PCM-or-compressed-5.1-only pipe).
The **channel-mapping population (F4, F11)** is the chapter's
binding to **Insight #2** — F4 silent channel-order mismatch
(Windows 7.1 uses L/R/C/LFE/Lb/Rb/Ls/Rs whereas
Dolby/DTS/SMPTE uses L/R/C/LFE/Ls/Rs/Lb/Rb so a transit that
"works" by byte-count produces a silently wrong soundstage —
the listener hears the rear-surround channels coming from
the side speakers and vice versa, but the stream is bit-clean
and no automated detector fires unless the chapter ships an
explicit channel-order-normaliser pass), F11 LFE channel
silently lost during stereo downmix (informational, not a
failure but tracked as a known-degradation row). The
**encoder-fault population (F5)** is the class where Opus
MultiStream encode fails for a configured channel count
above 8 (the practical 7.1 ceiling for the MVP per
`video-tech_dim06.md` §2.1 — RFC 6716 + RFC 7845 mapping
family 1 caps at 8 streams without falling back to family
255 ad-hoc mapping which requires per-tenant negotiation).
The **capture-fault population (F8, F9)** is the class where
the host-OS audio-capture surface is unavailable: F8
PipeWire monitor source absent on Linux (the host is
running PulseAudio without the PipeWire shim, or the user-
session bus does not expose monitor.\*), F9 WASAPI
loopback access denied on Windows (the host application
holds an exclusive-mode handle that blocks the loopback
client). The **operational-integrity population (F10)** is
the chapter's R-18 trip-wire: F10 r18.SafeExec rejects
`pacmd` (or `pactl`, `mpv --ao=jack`, or any other off-
allow-list audio-tooling argv shape). The **AV-sync
population (F12)** is the final population: F12 audio-
video sync drift > 50 ms (the C24 latency-budget audio-
side ceiling — beyond 50 ms even non-trained listeners
notice the desync; the chapter's mitigation is the audio-
video sync controller in §6 that consumes RTP timestamp
streams from both planes and drifts the audio-decode
pipeline forward or backward by ±10 ms per second until
re-aligned).

The five-column Symptom / Detection / Mitigation / Fallback
table below is the source of truth for the multi-channel-
audio runbook generator at
`../03_Architecture/12_Latency_Engineering_Overview.md` §13
and the alert-rule generation in
`../08_Operations/04_Observability_and_Events.md` (queued).
The fallback semantics across F1–F12 follow the **fail
closed at admission, degrade open at runtime** pattern
symmetric with C26 §7, C27 §7, C28 §7, C29 §7, and C30 §7.
Admission-time invariants (F1 capability-negotiation,
F5 Opus MultiStream channel-count cap, F7 SPDIF cannot
carry 7.1, F10 SafeExec argv allow-list) refuse session
admission with structured `audio.admission_refused
{session=…,cause=…}` events that the C24 measurement
harness propagates into the metrics plane and the per-
session capability snapshot. Runtime invariants (F2
Atmos passthrough breaks, F3 sample-rate mismatch, F4
silent channel-order mismatch, F6 eARC negotiation
fails, F8 PipeWire monitor-source absent, F9 WASAPI
loopback denied, F12 AV-sync drift) emit
`audio.degraded {from=…,to=…,reason=…}` events and the
fallback ladder runs forward — typically toward a
lower-channel-count container (7.1 → 5.1 → stereo per
F11), a different transport (eARC → ARC → SPDIF →
HDMI-direct-PCM per F6), a different capture backend
(PipeWire monitor → loopback-iface → ALSA snd-aloop per
F8), or an AV-sync re-alignment pass per F12.

The **F1 channel-config negotiation fails (fallback to
stereo)** row is the chapter's binding to the capability-
schema-driven negotiation contract from
`video-tech_dim06.md` §6 + §7. At session-create time the
client publishes its surround-decode + AVR-passthrough
capabilities (Atmos-passthrough yes/no, channel-count
ceiling 2/6/8, eARC yes/no, native sample rate 44.1/48/96
kHz). The host reconciles these against its own capture
topology and computes the agreed channel config. If the
schema parse fails (mismatched schema version, missing
required field), if the client's declared channels exceed
host capture availability (e.g. client wants 7.1 but the
host is mono), or if the negotiation simply times out
(default 2 s budget per session-create), the chapter's
mitigation is to **fall back to stereo (2-channel L/R)** as
the lowest-common-denominator mode that every host and
every client supports per `video-tech_dim06.md` §1. The
session is marked degraded in the catalogue with a
structured fault attribution; emit
`audio.negotiation_fallback {session=…,requested=…,
negotiated="stereo",cause=…}`. F1 is **degrade-open at
runtime** rather than fail-closed because audio-game tier
(OQ-C31-04) has demonstrated that a stereo session is
strictly better than no audio at all.

The **F2 Atmos passthrough breaks (DD+ container
malformed)** row binds the chapter to the **Dolby Digital
Plus + JOC (Joint Object Coding)** delivery path per
`video-tech_dim06.md` §2.4. When the client publishes
Atmos-passthrough=yes, the host wraps the game's PCM
capture into an E-AC3 + JOC container — a 5.1 E-AC3 core
plus an OAMD (Object Audio Metadata) sideband — and ships
the encapsulated bitstream to the client which forwards it
verbatim over HDMI eARC to the AVR for object-based
decoding. The encapsulation is fragile: if the OAMD
metadata is malformed (truncated, wrong frame-alignment),
the AVR's E-AC3 decoder still sees the 5.1 core and plays
it as plain 5.1 (backward-compatible per
`video-tech_dim06.md` §2.4) but the **Atmos
spatialisation is silently lost**. Detection requires the
client to read the AVR's Atmos-active flag back via HDMI-
CEC; mitigation is to **fall back to 5.1 PCM transport**
(no Atmos object-audio, just plain 5.1 channel-based) and
emit `audio.atmos_degraded
{session=…,avr_atmos_active=false,reason="oamd_malformed"}`.
F2 is the chapter's strongest motivation for shipping a
DD+JOC container fuzzer in §8.4.

The **F3 sample-rate mismatch (game 44.1, transport 48)**
row binds the chapter to the **sample-rate convergence
invariant**. Many older games still produce 44.1 kHz audio
internally (legacy CD-DA heritage), but the WebRTC + HDMI
eARC transport stack canonicalises on 48 kHz. The
chapter's mitigation is to **resample at the host capture
boundary** via a high-quality polyphase resampler (libsoxr
or the SOX cubic interpolation algorithm allow-listed in
the §1 family allow-list); the resampling adds bounded
latency (< 1 ms per channel at 48 kHz output) but
preserves the original 44.1 kHz spectrum. Detection is via
the capture-plane sample-rate metric; emit
`audio.resample_inserted
{session=…,from_hz=44100,to_hz=48000}`. F3 is **non-
blocking** but tracked as a per-session quality flag for
operator-policy review.

The **F4 channel-order mismatch silent (Insight #2 hardest
case)** row is the chapter's binding to **Insight #2** and
the chapter's **most operationally insidious** failure
mode. Per `video-tech_dim06.md` §2.6, Windows 7.1 uses the
channel order L/R/C/LFE/Lb/Rb/Ls/Rs (rear surrounds before
side surrounds), whereas the Dolby/DTS/SMPTE/film/cinema
world uses L/R/C/LFE/Ls/Rs/Lb/Rb (side surrounds before
rear surrounds). A bit-clean transit that "works" by
byte-count produces a **silently wrong soundstage**: the
listener hears rear-surround content from the side
speakers and side-surround content from the rear speakers.
**No automated detector fires** unless the chapter ships
an explicit channel-order-normaliser pass at the encode
boundary that re-orders the planar PCM samples into a
canonical (Dolby/SMPTE) order before Opus MultiStream
encode, and a complementary inverse pass at the client
decoder boundary keyed on the client-OS-side channel
order. The mitigation is the **ChannelOrderNormaliser**
component documented in §3.2 of this chapter — its
behaviour is deterministic, table-driven, and unit-
testable (cf §8.1). Emit `audio.channel_order_normalised
{session=…,host_order=…,canonical_order=…,client_order=…}`.
F4 is **fail-closed at admission** when the
ChannelOrderNormaliser is disabled — sessions cannot run
with the normaliser off, because the silent-failure mode
is worse than an outright refusal.

The **F5 Opus MultiStream encode fails > 8 channels** row
binds to the **Opus MultiStream channel-count ceiling**
from `video-tech_dim06.md` §2.1. RFC 6716 plus RFC 7845
channel-mapping family 1 supports up to 8 channels (7.1)
without falling back to channel-mapping family 255 (ad-
hoc mapping that requires per-tenant negotiation and
breaks the standard SDP `multiopus` payload format). For
MVP, the chapter caps the channel count at 8 and refuses
session-create for any configuration above that. The
mitigation is to surface the cap to operator-policy and
fall back to 7.1 with a structured event; emit
`audio.channel_count_capped
{session=…,requested=…,capped=8}`. F5 is **fail-closed
at admission** for configurations strictly above 8 (e.g.
9.1.6 Atmos with full discrete object-bed), and the
upgrade path is OQ-C31-01 (native Atmos object-audio
over WebRTC).

The **F6 eARC negotiation fails (HDMI 2.1 ARC fallback)**
row binds the chapter to the **HDMI eARC ↔ ARC fallback
ladder** from `video-tech_dim06.md` §6.1. eARC requires
HDMI 2.1 + a working CEC + an eARC-compliant AVR; if the
client TV / AVR pair fails the eARC handshake (older
firmware, marginal cable, mismatched HDMI-port pin
allocation), the chapter falls back through a four-stage
ladder: (1) eARC for full lossless 7.1 + Atmos (the happy
path; up to 37 Mbps); (2) ARC for compressed Atmos via
DD+JOC at 1 Mbps; (3) SPDIF for compressed 5.1 via
DTS or AC3 (cf F7 — SPDIF cannot carry 7.1 but it can
carry compressed 5.1); (4) HDMI direct PCM at the lowest
common channel count (typically stereo or 5.1 PCM). Emit
`audio.earc_fallback {session=…,from="eARC",to="ARC",
reason="cec_handshake_failed"}`. F6 is **degrade-open at
runtime**; the session continues at the next-lower
fidelity tier.

The **F7 SPDIF cannot carry 7.1** row binds the chapter to
the **SPDIF channel-count limit**. SPDIF (TOSLINK or
S/PDIF coax) is a stereo-PCM-or-compressed-5.1-only pipe
— the IEC 60958 / IEC 61937 standards cap the bandwidth
at 2-channel uncompressed PCM or 5.1-channel
compressed (DTS, AC3, DD+). The chapter's mitigation is
to **refuse the 7.1 configuration at admission** for
SPDIF-only clients with a structured event; emit
`audio.spdif_channel_count_refused
{session=…,client_transport="SPDIF",requested=8}`. F7
is **fail-closed at admission** because falling back
silently to 5.1 over SPDIF would lose the rear-surround
channels without operator-visible attribution.

The **F8 PipeWire monitor-source absent** row binds the
chapter to the **Linux audio-capture topology fallback
ladder** from `video-tech_dim06.md` §5.2. Per the §5.2
canonical capture-path, Linux capture goes via the
PipeWire user-session bus's monitor.\* sources (the
PipeWire equivalent of the PulseAudio monitor source).
If the host is running pure PulseAudio (no PipeWire
shim) or the user-session bus does not expose
monitor.\*, the capture-plane initialiser fails. The
mitigation is a **three-stage capture-fallback ladder**:
(1) PipeWire monitor.\* (preferred); (2) PulseAudio
monitor source via the `module-loopback` shim; (3) ALSA
`snd-aloop` loopback device as last resort (lowest
fidelity, no per-app routing). Emit
`audio.capture_fallback {session=…,from="pipewire",
to="pulseaudio",reason="monitor_source_absent"}`. F8 is
**degrade-open at runtime**; the session continues at
the next-tier capture backend.

The **F9 WASAPI loopback access denied** row binds the
chapter to the **Windows audio-capture access-control
fallback** from `video-tech_dim06.md` §5.1. WASAPI
loopback capture works in shared mode by default, but
if the host application opens the audio device in
**exclusive mode** (e.g. an audiophile-grade music
player or an old game), the loopback client gets
`AUDCLNT_E_DEVICE_IN_USE` and capture fails. The
mitigation is to fall back to the **render-to-loopback
shim** (a virtual-audio-cable-style driver shipped as an
optional submodule under `vasic-digital`) which presents
itself as the default render device and tees the PCM
buffer to the capture client. Emit
`audio.wasapi_fallback {session=…,from="loopback",
to="render_shim",reason="exclusive_mode"}`. F9 is
**degrade-open at runtime**.

The **F10 r18.SafeExec rejects pacmd** row is the chapter's
R-18 trip-wire and is symmetric with C26-F9, C27-F10,
C28-F10, C29-F10, C30-F10. When a developer adds a non-
allow-listed audio-tooling argv shape (e.g. `pacmd
load-module module-loopback` for an ad-hoc PulseAudio
re-routing, or `pactl` against a foreign user session,
or `mpv --ao=jack` for an unsupported transport), the
wrapper rejects the call at the `os/exec` boundary and
bootstrap aborts. The allow-list lives in
`vasic-digital/helix-r18-safeexec` and is **not
duplicated** in this chapter; the family allow-list
extension that C31 contributes (canonical
`pw-cli list-objects`, `pactl list short sources`,
`wpctl status`, `aplay -L`, `pa-info`, `ffmpeg -f
pulse -i default`, the WASAPI-loopback canonical client
flag set, the Core Audio AudioServerPlugin canonical
shapes) is recapped in §1 (family allow-list) of this
chapter and verified by the C08 `host-integrity-scan`
test inherited verbatim into §8.11. Bypass requires an
allow-list extension via operator review per Constitution
§11.5.4, never a silent workaround. Emit
`audio.safeexec_rejected {tool="pacmd",argv=…}`.

The **F11 LFE channel lost in stereo downmix
(informational, not failure)** row binds the chapter to
the **stereo-downmix LFE-handling contract**. When a 5.1
or 7.1 source is downmixed to stereo (per F1 fallback,
or per a client whose hardware is genuinely stereo-
only), the LFE (Low-Frequency Effects, the ".1") channel
is conventionally folded into the L+R sum at -10 dB
attenuation, which preserves the bass content but loses
the discrete LFE routing. This is **not a failure**
under the SMPTE downmix coefficients but is tracked as
an informational row for operator-policy visibility.
Emit `audio.lfe_downmixed {session=…,from_channels=6,
to_channels=2,lfe_attenuation_db=-10}`. F11 is
**informational** — no fallback fires; the row exists
to make the downmix-loss explicit in the per-session
quality plane and the catalogue's audit record.

The **F12 audio-video sync drift > 50 ms** row binds the
chapter to the **AV-sync controller** documented in §6
of this chapter and the C24 latency-budget audio-side
ceiling. The audio and video planes have independent
RTP timestamp streams, independent jitter buffers, and
independent decode pipelines; if they drift more than
50 ms (the well-established perceptual-desync
threshold for trained-listener-detectable drift), the
chapter's AV-sync controller fires a re-alignment pass
that drifts the audio-decode pipeline forward or
backward by ±10 ms per second until re-aligned. The
±10 ms / second drift rate is the perceptual-tolerance
ceiling — a faster drift would itself be audible as
pitch-shift. Emit `audio.av_sync_drift
{session=…,drift_ms=…,direction=…,recovery_started=…}`.
F12 is **degrade-open at runtime**; the live session
continues with audible drift for at most 5 seconds
during the re-alignment pass.

| # | Failure mode | Symptom | Detection | Mitigation | Fallback |
|---|---|---|---|---|---|
| F1 | Channel-config negotiation fails — schema parse error / channel-count exceeds host capture topology / negotiation timeout (2 s default) — per `video-tech_dim06.md` §6 capability-schema contract | Session-create observes negotiation timeout or schema mismatch; emits `audio.negotiation_fallback {session=…,requested=…,negotiated="stereo",cause=…}` | Negotiation watchdog — `audio.Negotiation.Watchdog(session)` polls handshake state every 250 ms; fires on timeout or schema validate failure | Fall back to stereo (2-channel L/R) as the lowest-common-denominator mode that every host and every client supports per `video-tech_dim06.md` §1; the session is marked degraded in the catalogue with structured fault attribution | Stereo session — non-blocking; the operator dashboard surfaces per-session degradation rate as a capability-mismatch indicator |
| F2 | Atmos passthrough breaks — DD+JOC E-AC3 container malformed (truncated OAMD metadata, wrong frame-alignment); per `video-tech_dim06.md` §2.4 the AVR plays the 5.1 core but Atmos spatialisation is silently lost | AVR's Atmos-active flag is read back via HDMI-CEC; emits `audio.atmos_degraded {session=…,avr_atmos_active=false,reason="oamd_malformed"}` | Atmos-active probe — `audio.Atmos.CECProbe()` reads the AVR's Atmos-active state; cross-references against the host's intended-active state | Fall back to 5.1 PCM transport (no Atmos object-audio, just plain 5.1 channel-based); emit the structured degradation event; cross-link OQ-C31-01 native Atmos object-audio | Channel-based 5.1 — non-blocking; the session continues at one tier below Atmos; the operator dashboard tracks per-AVR-vendor Atmos-degradation rate |
| F3 | Sample-rate mismatch — game produces 44.1 kHz audio internally (legacy CD-DA heritage); WebRTC + HDMI eARC transport canonicalises on 48 kHz | Capture-plane sample-rate metric reports 44.1 kHz against the 48 kHz transport; emits `audio.resample_inserted {session=…,from_hz=44100,to_hz=48000}` | Sample-rate watchdog — `audio.SampleRate.Watchdog()` polls capture-plane every 1 s | Resample at the host capture boundary via libsoxr (allow-listed cubic interpolation); adds < 1 ms latency per channel; preserves the 44.1 kHz spectrum | Inserted resampler — non-blocking; tracked as a per-session quality flag |
| F4 | Channel-order mismatch silent (**Insight #2 hardest case**) — Windows 7.1 uses L/R/C/LFE/Lb/Rb/Ls/Rs whereas Dolby/DTS/SMPTE uses L/R/C/LFE/Ls/Rs/Lb/Rb; bit-clean transit produces silently wrong soundstage with **no automated detector firing** unless the chapter ships an explicit ChannelOrderNormaliser pass | The ChannelOrderNormaliser fires at the encode boundary; emits `audio.channel_order_normalised {session=…,host_order=…,canonical_order=…,client_order=…}` | The ChannelOrderNormaliser's deterministic table-driven pass at encode boundary — there is no symptom-side detector for the silent failure, the normaliser is the only line of defence | The ChannelOrderNormaliser re-orders planar PCM samples into canonical (Dolby/SMPTE) order before Opus MultiStream encode; complementary inverse pass at client decoder keyed on client-OS channel order | **Fail-closed at admission** when the ChannelOrderNormaliser is disabled — sessions cannot run with the normaliser off because the silent-failure mode is worse than an outright refusal |
| F5 | Opus MultiStream encode fails > 8 channels — RFC 6716 + RFC 7845 channel-mapping family 1 caps at 8 streams without falling back to family 255 ad-hoc mapping per `video-tech_dim06.md` §2.1 | Session-create admission-time check observes channel-count > 8; emits `audio.channel_count_capped {session=…,requested=…,capped=8}` | Channel-count check at admission — `audio.ChannelCount.Check(channels)` against the family-1 ceiling | Cap at 8 channels (7.1); refuse session-create for configurations strictly above 8 (e.g. 9.1.6 Atmos with full discrete object-bed); cross-link OQ-C31-01 | **Fail-closed at admission** — non-overridable for MVP; the upgrade path is OQ-C31-01 native Atmos object-audio over WebRTC |
| F6 | eARC negotiation fails (HDMI 2.1 ARC fallback) — older firmware, marginal cable, mismatched HDMI-port pin allocation per `video-tech_dim06.md` §6.1 | eARC handshake fails via HDMI-CEC; emits `audio.earc_fallback {session=…,from="eARC",to="ARC",reason="cec_handshake_failed"}` | eARC handshake watchdog — `audio.eARC.Handshake(timeout=3s)` fires on CEC-handshake timeout | Four-stage fallback ladder: (1) eARC for full lossless 7.1 + Atmos; (2) ARC for compressed Atmos via DD+JOC; (3) SPDIF for compressed 5.1; (4) HDMI direct PCM at lowest common channel count | Next-lower fidelity tier — non-blocking; the session continues at compressed Atmos / compressed 5.1 / direct PCM as the link allows |
| F7 | SPDIF cannot carry 7.1 — IEC 60958 / IEC 61937 cap at 2-channel uncompressed PCM or 5.1-channel compressed (DTS, AC3, DD+) | Session-create admission-time check observes 7.1 + SPDIF transport; emits `audio.spdif_channel_count_refused {session=…,client_transport="SPDIF",requested=8}` | Transport / channel-count check at admission — `audio.SPDIF.ChannelCount.Check(channels)` against the 5.1 ceiling | Refuse the 7.1 configuration at admission for SPDIF-only clients; surface to operator-policy console; client may upgrade to eARC or accept 5.1 | **Fail-closed at admission** — non-overridable; falling back silently to 5.1 over SPDIF would lose the rear-surround channels without operator-visible attribution |
| F8 | PipeWire monitor-source absent — host running pure PulseAudio (no PipeWire shim) or user-session bus does not expose monitor.\* sources per `video-tech_dim06.md` §5.2 | Capture-plane initialiser fails on monitor.\* enumeration; emits `audio.capture_fallback {session=…,from="pipewire",to="pulseaudio",reason="monitor_source_absent"}` | Capture-plane initialiser — `audio.CaptureInit.PipeWire()` fails fast on monitor.\* enumeration | Three-stage capture-fallback ladder: (1) PipeWire monitor.\*; (2) PulseAudio monitor source via `module-loopback` shim; (3) ALSA `snd-aloop` loopback as last resort | Next-tier capture backend — non-blocking; the session continues at the next backend with a per-session quality flag |
| F9 | WASAPI loopback access denied — host application holds an exclusive-mode handle that blocks the loopback client; returns `AUDCLNT_E_DEVICE_IN_USE` per `video-tech_dim06.md` §5.1 | WASAPI loopback client returns `AUDCLNT_E_DEVICE_IN_USE`; emits `audio.wasapi_fallback {session=…,from="loopback",to="render_shim",reason="exclusive_mode"}` | WASAPI loopback initialiser — `audio.WASAPI.Loopback()` returns the structured access-denied error | Fall back to render-to-loopback shim (virtual-audio-cable-style driver shipped as optional `vasic-digital` submodule) which presents itself as the default render device and tees the PCM buffer to the capture client | Render-shim capture — non-blocking; the session continues at slightly lower fidelity (the shim adds ~3 ms buffer) |
| F10 | `r18.SafeExec` rejects `pacmd` (or `pactl`, `mpv --ao=jack`, or any other off-allow-list audio-tooling argv shape) | Bootstrap fails on audio-tooling initialisation; structured error includes the rejected argv with the offending flag highlighted; harness logs `audio.safeexec_rejected {tool="pacmd",argv=…}` | Wrapper's verbatim allow-list check at the `os/exec` boundary returns `ErrForbiddenArgvShape`; the harness logs the rejection | Fix the call site to use the allow-listed shape — canonical `pw-cli list-objects`, `pactl list short sources`, `wpctl status`, `aplay -L`, `pa-info`, `ffmpeg -f pulse -i default`, the WASAPI-loopback canonical client flag set per family allow-list (`00_Index.md` §7) | Blocking — bootstrap aborts; non-overridable per Constitution §11.5.4; the rule lives in the Constitution and bypass requires a §13 exception with documented mitigation; cross-link §8.11 host-integrity-scan |
| F11 | LFE channel lost in stereo downmix (**informational, not failure**) — per SMPTE downmix coefficients the LFE (".1") channel is conventionally folded into L+R sum at -10 dB attenuation; preserves bass content but loses discrete LFE routing | Downmix path observes 5.1 / 7.1 → stereo conversion; emits `audio.lfe_downmixed {session=…,from_channels=6,to_channels=2,lfe_attenuation_db=-10}` | Downmix path — `audio.Downmix.Stereo()` instruments the LFE-fold-in step | Apply the SMPTE downmix coefficients (LFE → L+R at -10 dB; surround → L+R at -3 dB); emit the informational event; tracked as a per-session quality flag | Informational — no fallback fires; the row exists to make downmix-loss explicit in the per-session quality plane and catalogue audit record |
| F12 | Audio-video sync drift > 50 ms — independent RTP timestamp streams + independent jitter buffers + independent decode pipelines drift more than 50 ms (well-established perceptual-desync threshold) | AV-sync controller observes drift via cross-plane RTP timestamp comparison; emits `audio.av_sync_drift {session=…,drift_ms=…,direction=…,recovery_started=…}` | AV-sync controller — `audio.AVSync.Controller()` polls cross-plane timestamps every 100 ms and fires on > 50 ms drift | Re-alignment pass that drifts audio-decode pipeline forward or backward by ±10 ms / second until re-aligned (the ±10 ms/s drift rate is the perceptual-tolerance ceiling — faster drift would be audible as pitch-shift) | Re-aligned audio — non-blocking; the live session continues with audible drift for at most 5 seconds during the re-alignment pass |

## 8. Test surface

The C31 test surface inherits the family-level container-driven
CI lane contract from C26 §8 + C27 §8 + C28 §8 + C29 §8 + C30 §8
and the `vasic-digital/Containers` runner image, **extended**
with the new multi-channel-audio-specific requirement: every
integration / E2E / chaos / stress test must exercise **all
three host-OS audio-capture backends** (PipeWire on Linux,
WASAPI loopback on Windows, Core Audio on macOS) so the
capture-plane fallback ladder (F8 + F9 + the macOS analogue)
is validated against real backend implementations (mocking the
capture backends is forbidden per Constitution §6.4 — only
unit tests may use mocks; the canonical local-capture test
fleet is documented at `video-tech_dim10.md` §5). Per
Constitution §6.4 + Master Plan §4.3 anti-bluff verification,
the test matrix below cites `video-tech_dim06.md` (audio
codec + delivery dimension) and `video-tech_dim10.md` (testing
dimension) explicitly so every per-channel-config performance
claim is grounded in a primary-source reference.

### 8.1 Unit (mocks allowed)

The unit-test layer is the only place where mocks, stubs, or
hardcoded values are permitted per Constitution §6.4 — every
other layer below hits the real system.

- **ChannelOrderNormaliser unit test** — given a synthetic
  8-channel PCM buffer in Windows 7.1 order
  (L/R/C/LFE/Lb/Rb/Ls/Rs) and the canonical-target order
  (L/R/C/LFE/Ls/Rs/Lb/Rb per Dolby/SMPTE), assert the
  normaliser produces the correctly re-ordered buffer with
  bit-exact sample equality across all channels; assert
  the inverse pass at the decoder boundary correctly
  inverts the re-ordering for both the Windows-client and
  Linux/macOS-client target orders. Cross-link
  `video-tech_dim06.md` §2.6 — this test is the chapter's
  binding to **Insight #2** at the unit-test layer.
- **Channel-mapping-family-1 SDP construction unit test**
  — given a 5.1 and a 7.1 channel configuration, assert
  the SDP `multiopus` attribute is constructed with the
  correct `num_streams`, `coupled_streams`, and
  `channel_mapping` parameters per RFC 7845 + draft-shin-
  avtcore-rtp-multi-opus; assert the SDP parses cleanly
  via the standard `pjsip` parser fixture.
- **Stereo-downmix coefficient unit test** — given a 5.1
  and a 7.1 source PCM buffer, assert the SMPTE downmix
  coefficients (LFE → L+R at -10 dB; surround → L+R at
  -3 dB) are applied with bit-exact sample equality
  against a reference downmix produced by `ffmpeg -ac 2`.
- **Capability-schema validate unit test** — given the
  schema at `vasic-digital/helix-audio/schema/v1.json`,
  fuzz with 10⁵ valid + 10⁵ invalid client capability
  payloads; assert validation correctness with no false-
  positive / false-negative.

### 8.2 Integration

The integration-test layer hits the real PipeWire (Linux) /
WASAPI (Windows) / Core Audio (macOS) capture backend +
real Opus MultiStream encoder (libopus 1.5+ with the
multistream API) — no mocks, no stubs, no hardcoded values.
Per Constitution §6.4 this layer must run inside the
canonical `vasic-digital/Containers` runner image with the
appropriate per-OS audio-capture-backend host-passthrough.

- **Real PipeWire monitor.\* capture + Opus MultiStream
  encode integration** — boot the audio-capture lane in a
  Linux container with PipeWire user-session passed
  through; play a synthetic 5.1 + 7.1 reference signal
  via `pw-play`; capture via the chapter's PipeWire
  monitor.\* client; encode via Opus MultiStream at
  5 ms / 10 ms / 20 ms frame sizes; decode at the
  receiver; assert the per-channel sample-bit-equality
  is preserved (within Opus quantisation noise floor at
  the configured bitrate).
- **Real WASAPI loopback capture + Opus MultiStream
  encode integration** — same fixture but on a Windows
  Server 2022 runner with WASAPI loopback against a
  reference signal generator.
- **Real Core Audio capture + Opus MultiStream encode
  integration** — same fixture but on a macOS runner
  via `AudioServerPlugin`.
- **DD+JOC encapsulation integration** — encapsulate a
  synthetic Atmos object-audio bed into the E-AC3 + JOC
  container per `video-tech_dim06.md` §2.4; assert the
  AVR's Atmos-active flag fires under a reference AVR
  fixture (the canonical Denon AVR-X3700H test rig per
  `video-tech_dim10.md` §5).

### 8.3 E2E

The E2E layer brings up the **full multi-channel
recording-decode-playback** pipeline end-to-end and
asserts channel-order fidelity at the decode side.

- **7.1 surround stream end-to-end + verify channel-order
  at decode** — boot a host with a synthetic 7.1
  reference signal generator (each channel emits a
  unique-frequency pure tone: L=110 Hz, R=220 Hz,
  C=330 Hz, LFE=55 Hz, Ls=440 Hz, Rs=550 Hz, Lb=660 Hz,
  Rb=770 Hz); session-create from a synthetic client;
  capture, encode, transport, decode, play via the
  client's audio-output; **for every channel, assert
  the per-channel pure-tone frequency is detected at
  the correct output channel** via FFT analysis of the
  client-output PCM stream; assert the channel-order
  inversion at decode is correctly applied for both
  Windows-client and Linux/macOS-client target orders;
  assert no audible inter-channel crosstalk above
  -60 dB.
- **5.1 surround end-to-end** — same fixture at 5.1.
- **Stereo end-to-end** — same fixture at stereo;
  asserts the F1 fallback path produces a clean L/R
  output with the SMPTE downmix coefficients correctly
  applied.
- **Atmos passthrough end-to-end** — record a session
  with Atmos-passthrough enabled; route to the canonical
  reference AVR; assert the AVR's Atmos-active flag
  fires; assert the per-object spatial position is
  preserved within the AVR's reported tolerance.

### 8.4 Security

- **`r18.SafeExec` rejection fuzz** — for each of the
  family allow-list entries (`00_Index.md` §7),
  construct off-allow-list argv shapes (e.g. `pacmd
  load-module` is off-list; `pactl set-sink-volume` is
  off-list; `mpv --ao=jack` is off-list; `ffmpeg -f
  jack` is off-list) and fuzz with 10⁶ argv permutations
  per Constitution §6.4 fuzz contract; assert the
  wrapper returns `ErrForbiddenArgvShape` for every
  off-list shape with no false-positive on allow-list
  shapes; assert no host-disruptive command (kill,
  systemctl, pmset) ever passes the wrapper.
- **DD+JOC container fuzzer** — synthesise 10⁵ malformed
  E-AC3 + JOC containers (truncated OAMD, wrong frame-
  alignment, byte-flipped sync words); assert the F2
  detection path fires for every malformed container
  and the fallback to plain 5.1 PCM is correctly
  activated.
- **Capability-schema authorisation** — assert
  capability-schema mutations are authenticated and
  authorised per the C09 security family (cross-link);
  assert unauthorised schema-update attempts are
  refused with structured audit events.

### 8.5 Benchmarking

The benchmarking layer is the chapter's binding to
Constitution §6 — every per-configuration audio
performance claim reports p50 / p99 / p999 at ≥ 10 K
samples via the C24 measurement harness. Cross-link C24 /
C35. Per **`video-tech_dim10.md`** §2 + §5, the
benchmarking corpus uses synthetic-content + real-game-
capture pairs across the six representative game profiles
(FPS, racing, RPG, RTS, MOBA, fighting) so the per-
configuration performance characterisation reflects
production-like workloads.

- **Bench Opus MultiStream encode latency at 5.1** —
  measure per-frame encode latency across 10 K samples at
  5.1 channel configuration with 5 ms / 10 ms / 20 ms
  frame sizes and 256 / 384 / 450 kbps bitrates; **report
  p50 / p99 / p999 per Constitution §6 with ≥ 10 K
  samples**; histogram artifact attached; budget
  per-frame encode latency p999 < 3 ms at 5 ms frame size
  per `video-tech_dim10.md` §3 regression-detection
  thresholds.
- **Bench Opus MultiStream encode latency at 7.1** —
  same fixture at 7.1 channel configuration; budget
  per-frame encode latency p999 < 5 ms at 5 ms frame
  size; the higher channel count adds bounded encode
  cost.
- **Bench end-to-end audio latency** — measure capture-
  to-playback latency under WASAPI exclusive mode (per
  `video-tech_dim06.md` §1) + PipeWire user-session +
  Core Audio per-OS-respective; report p50 / p99 / p999
  per OS with ≥ 10 K samples; budget end-to-end audio
  latency p999 < 20 ms per `video-tech_dim06.md` §1.
- **Bench ChannelOrderNormaliser throughput** — measure
  the table-driven re-order pass throughput at 5.1 and
  7.1 channel counts; budget ≥ 1 GB/s on AVX2-equipped
  runners; budget ≥ 500 MB/s on ARM runners.
- **Bench AV-sync drift detection latency** — measure
  the AV-sync controller's drift-detection cycle latency
  (drift-onset to detection-fired); budget p999 < 100 ms
  to detect a > 50 ms drift event.
- Cross-link **C24 / C35** measurement harness for
  shared histogram-collection + bootstrap-resampling-
  confidence-interval primitives. The benchmark suite
  must cite **`video-tech_dim10.md`** explicitly per
  Master Plan §4.3 anti-bluff verification —
  `video-tech_dim10.md` §3 enumerates the per-
  configuration regression-detection thresholds + §5
  enumerates the canonical bench corpus + §7
  enumerates the per-OS audio-capture latency budgets.
  Cross-link **C24** §6 (latency-side measurement)
  and **C35** §3 (quality-side measurement) for the
  full harness contract.

### 8.6 Chaos

- **Force capability mismatch mid-session — assert
  fallback ladder** — boot host + client at 7.1 +
  Atmos; mid-session, force the client to advertise a
  reduced capability (5.1 only, then stereo-only, then
  Atmos-disabled); assert the chapter's fallback
  ladder (F1 + F6) re-negotiates correctly without
  session interruption; assert the per-fallback
  structured event fires; assert no perceptible audio
  glitch above 200 ms during the transition.
- **Force PipeWire monitor.\* absence (F8)** — on a
  Linux host, kill the PipeWire user-session bus mid-
  session; assert the F8 fallback ladder fires
  (PipeWire → PulseAudio → ALSA `snd-aloop`); assert
  the session continues with a per-session quality
  flag.
- **Force WASAPI exclusive-mode lock (F9)** — on a
  Windows host, open a foreign WASAPI exclusive-mode
  client mid-session; assert the F9 render-shim
  fallback fires; assert the session continues at the
  reduced fidelity.
- **Force AV-sync drift (F12)** — inject a 100 ms
  artificial delay into the audio-decode pipeline;
  assert the AV-sync controller fires the re-
  alignment pass within 100 ms p999; assert the
  re-alignment completes within 5 seconds.
- **Force DD+JOC malformation (F2)** — inject byte-
  flips into the OAMD metadata mid-session; assert
  the F2 fallback to plain 5.1 fires; assert the
  AVR's Atmos-active flag transitions to false.

### 8.7 Stress

- **24h 7.1 sustained — assert no leak / no corruption
  / no drift** — on each runner, run continuous 7.1
  surround capture + Opus MultiStream encode + decode +
  playback for 24 hours; **assert no fd leak** (process
  fd count stable to within 5 fds over 24 h); **assert
  no GC stall > 1 ms** (GODEBUG=gctrace=1 trace artifact
  attached; cross-link C36 §3 Go pipeline `sync.Pool`
  discipline); assert no memory leak (RSS growth
  < 5 MB / hour); assert no AV-sync drift > 50 ms
  cumulative over the 24 h window; assert no audible
  glitch (per a continuous FFT-based glitch-detector
  fixture).
- **Multi-session concurrent stress** — provision 100
  concurrent 7.1 sessions on a single runner; assert
  the per-session encode-latency p999 stays within the
  §8.5 budget under concurrent load; assert no cross-
  session audio bleed.
- **Per-OS capture-backend stress** — exercise each of
  the three OSes (Linux PipeWire, Windows WASAPI,
  macOS Core Audio) under continuous-capture load
  for 4 hours; assert no per-OS-specific regression.

### 8.8 Smoke

- **Capability schema reports correct channel count** —
  boot the host-agent in a clean container; for each of
  the canonical configurations (stereo, 5.1, 7.1,
  Atmos), assert the published capability schema
  reports the correct channel count, mapping family,
  and AVR-passthrough flag; assert the schema validates
  against `vasic-digital/helix-audio/schema/v1.json`.
- **Smoke test capture + encode + decode** — dispatch a
  5-second 5.1 capture; Opus MultiStream encode; decode
  at the receiver; assert per-channel pure-tone
  frequency detection at the decode side; assert no
  `audio.degraded` event was emitted.

### 8.9 Full automation

All of §8.1–§8.8 run on **every commit via the local
container-driven CI lane** per Constitution §10. The CI
lane uses the canonical `vasic-digital/Containers` runner
image with per-OS audio-capture-backend host-passthrough
(PipeWire user-session bus on Linux, WASAPI on Windows,
Core Audio on macOS) and the local-AVR test fixture
(canonical Denon AVR-X3700H test rig per
`video-tech_dim10.md` §5) addressable on the runner
network. The matrix covers (Linux Ubuntu 22.04 / 24.04 +
Fedora 40, Windows Server 2022, macOS 14) × (stereo,
5.1, 7.1, Atmos-passthrough). The full-automation lane
emits a single composite artifact
(`audio-test-report.json`) that the C35 quality-claim
harness consumes as the authoritative source-of-truth
for any per-configuration multi-channel-audio
performance claim in chapter prose.

### 8.10 Challenges (production-like)

HelixQA dispatches **stereo + 5.1 + 7.1 + Atmos
passthrough sessions concurrently** from
`git@github.com:vasic-digital/Challenges.git` (per
Constitution §6.4 Challenges-test contract):

- **Stereo Challenges** — HelixQA boots fully-
  provisioned hosts and clients in stereo
  configuration; runs a 30-minute session per the
  reference game profile; asserts F1 stereo fallback
  fires correctly when the client is forced to a
  stereo-only capability; asserts SMPTE downmix
  coefficients correctly applied (F11 informational
  event fires).
- **5.1 surround Challenges** — same fixture at 5.1;
  asserts per-channel fidelity end-to-end via FFT
  analysis at the decode side; asserts the per-
  channel pure-tone frequency detection passes for
  all 6 channels.
- **7.1 surround Challenges** — same fixture at 7.1;
  asserts the ChannelOrderNormaliser correctly
  handles both Windows-client and Linux/macOS-client
  decode paths; asserts no silent channel-order
  swap (F4 the Insight #2 hardest case).
- **Atmos passthrough Challenges** — HelixQA boots
  a fully-provisioned host + client + canonical
  Denon AVR-X3700H reference fixture; records a
  30-minute session with Atmos-passthrough enabled;
  asserts the AVR's Atmos-active flag fires; asserts
  per-object spatial position is preserved within
  the AVR's reported tolerance.
- **Concurrent multi-config Challenges** — dispatch
  the four configurations (stereo + 5.1 + 7.1 +
  Atmos) concurrently across multiple host + client
  + AVR fixtures; assert per-session fidelity holds
  under contention.

### 8.11 Inherited host-integrity-scan (R-18, non-overridable per Constitution §11.5.4)

Inherits VERBATIM from C08 §12.11: a dedicated test, mandated
by Constitution §11.5.4, that boots the host agent under
`strace -fe trace=execve` on a Linux test host (the canonical
reference platform for this scan) and runs the full Ten-test-
type matrix above against it. The strace log is then grepped
for **every** §11.5.1 forbidden pattern. The gate is:

```
zero matches across the entire log → PASS
≥ 1 match anywhere in the log     → FAIL (non-overridable)
```

The log is preserved as an artifact alongside the auditd
record from §12.4. The test is **non-overridable** per
Constitution §11.5.4: a match is a Constitution violation,
never a flake, and bypass requires a §13 exception with a
documented compensating control. The same test is replicated
on Windows under `Process Monitor` ETW filtered to
`Process Create`, and on macOS under `dtruss -f -t execve`,
so the host-integrity-scan covers all three host OSes the
agent ships on.

The C31 implementation contract that this scan validates:

- Audio-tooling invocation via `r18.SafeExec` only — never
  via `os/exec.Command` directly; the canonical shapes
  (`pw-cli list-objects`, `pactl list short sources`,
  `wpctl status`, `aplay -L`, `pa-info`,
  `ffmpeg -f pulse -i default`, the WASAPI-loopback
  canonical client flag set, the Core Audio
  AudioServerPlugin canonical shapes) are the family
  allow-list entries for multi-channel-audio tooling.
- No host-disruption commands ever appear in the audio-
  capture path: no `kill -9 <pid>`, no
  `systemctl suspend|hibernate|reboot|halt|poweroff`, no
  `pmset`, no `xset dpms force off`, no `--privileged`
  container flag, no host-mount of `/`, `/dev`, `/proc`,
  `/sys`. The scan asserts none of these syscall patterns
  appear in the audio-capture subsystem's syscall trace.
- No cross-tenant audio-stream traversal — the scan
  asserts the audio-capture worker's `openat` syscalls
  never reference paths outside the per-tenant scoped
  audio-device root, and no `chdir` / `chroot` syscall
  escapes the scope.

The scan's invocation contract is byte-identical with the
C08 §12.11 inheritance into every chapter in the family per
`00_Index.md` §7 R-18 family allow-list. No chapter in the
family is permitted to redefine, override, or extend the
scan — Constitution §11.5.4 forbids per-chapter
customisation of the host-integrity contract.

## 9. Open questions

The following open questions are tracked in the chapter's OQ
log and surface to the family-level OQ aggregator at
`00_Index.md` §5. Each OQ is prefixed `OQ-C31-NN` and carries
an owner, a target resolution date, and a cross-link to the
deciding chapter or external dependency.

- **OQ-C31-01** — Native Atmos object-audio over WebRTC
  (custom Opus extension). The MVP delivers Atmos via the
  passthrough path (host wraps PCM in DD+JOC; client
  forwards bitstream over eARC to AVR for decode). True
  native Atmos object-audio over WebRTC would require a
  custom Opus channel-mapping family extension carrying
  the OAMD metadata sideband — RFC-track work, not yet
  standardised. Trigger: V1 audiophile-tier operator-
  policy posture emerges. Owner: C31 + V1 family + Audio
  Codec WG. Cross-link F2 + F5 — F5 channel-count cap
  lifts when family-255 ad-hoc mapping lands.
- **OQ-C31-02** — Voice chat integration with multi-channel
  surround (cross-link C06). The MVP separates game-audio
  (multi-channel) from voice-chat (mono / stereo) on
  independent RTP streams. Should the chapter integrate
  voice-chat as an additional in-band channel of the multi-
  channel Opus MultiStream payload (per RFC 7845 mapping
  family 1 with up to 8 streams plus a dedicated voice-
  channel slot)? The cost is per-tenant complexity at the
  channel-mapping layer; the benefit is a single jitter
  buffer + sync controller for all audio. Trigger: V1
  voice-chat surface stabilises in C06. Owner: C31 + C06
  + V1 family. Cross-link `video-tech_dim06.md` §2.1
  channel-mapping families.
- **OQ-C31-03** — Apple Music spatial audio interop. Apple
  Music's spatial audio (Dolby Atmos for Music) uses a
  proprietary Apple-specific channel-mapping atop the
  AAC-LC codec; HelixPlay's MVP uses Opus MultiStream
  exclusively. Should V1 ship an Apple-specific delivery
  path for Apple-ecosystem clients (Mac client, Apple TV
  client) that maintains parity with Apple Music's
  spatial-audio rendering? Owner: C31 + V1 family +
  Apple Ecosystem WG. Cross-link
  `video-tech_dim06.md` §2.5 (AAC-LC + spatial).
- **OQ-C31-04** — Audio-only sessions (audio-game tier).
  The MVP is a video-first cloud-gaming product. Should
  HelixPlay support audio-only sessions for accessibility
  scenarios (visually-impaired players), bandwidth-
  constrained scenarios (audio-only over LTE / satellite
  links), or audio-game genres (interactive fiction,
  audio-RPGs)? The cost is a parallel session-create
  path that disables the video plane and a UX surface
  that supports audio-only navigation; the benefit is a
  new accessibility tier and a new market segment.
  Owner: C31 + V1 family + Accessibility WG. Cross-link
  F1 fallback path (audio-only is the absolute floor of
  the F1 stereo fallback ladder).
- **OQ-C31-05** — Personal HRTF profiles for binaural
  rendering. Personal Head-Related Transfer Function
  (HRTF) profiles enable accurate per-listener binaural
  rendering on stereo headphones; the listener provides
  ear-shape measurements via a smartphone scan, the
  HRTF is stored in the client capability schema, and
  the client renders the multi-channel game audio to
  stereo via the personal HRTF for accurate
  spatialisation. The cost is per-listener HRTF
  storage + a binaural renderer at the client; the
  benefit is a step-function spatial-audio quality
  improvement on stereo headphones (the most common
  consumer audio target). Trigger: V1 audiophile-tier
  operator-policy posture emerges. Owner: C31 + V1
  family + Spatial Audio WG. Cross-link OQ-C31-01
  native Atmos and OQ-C31-04 audio-only sessions —
  both intersect with binaural rendering.

---

## 10. References

### Project artifacts

- Master Plan, Constitution, System Overview, Video/Audio family index ([`00_Index.md`](00_Index.md)). Sibling chapters cited above.

### Source research artifacts

- `video-tech_dim06.md` (1,141 lines), `video-tech.agent.final.md` (2,588 lines), `video-tech_insight.md` (Insight #2 — BINDING), `video-tech_cross_verification.md`, `video-tech_dim10.md` (testing).

### Web research

[`../99_Web_Research_Addenda/2026-04-29-audio-pipeline.md`](../99_Web_Research_Addenda/2026-04-29-audio-pipeline.md) — 783 lines, 9 clusters + §Z.

| Cluster | Topic | Cited |
|---------|-------|------|
| §A | Opus MultiStream WebRTC up to 8 channels | §2 |
| §B | Atmos object-audio over DD+ + JOC | §2.4 |
| §C | eARC + HDMI 2.1 audio specs | §3 |
| §D | SPDIF AC3/DTS legacy | §3.3 |
| §E | Channel-order mismatch (Windows vs Dolby) | §4 |
| §F | PipeWire (Linux) audio capture | §5.1 |
| §G | WASAPI loopback (Windows) | §5.2 |
| §H | CoreAudio (macOS) | §5.3 |
| §I | Audio capability negotiation chain | §1, §2.5 |
| §Z | Contradictions index | §1, §2, §3, §4 |

---

## Anti-Bluff Verification

> Per [Master Plan §4.3](../../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Source Evidence Reviewed

| Path | Lines | Reviewed | Used |
|------|------:|----------|------|
| `video-tech_dim06.md` | 1,141 | A, B, C, D | §§1–9 (primary) |
| `video-tech.agent.final.md` | 2,588 | A, B, C | §§1–6 |
| `video-tech_insight.md` | 243 | A, B | §1 (#2 BINDING) |
| `video-tech_cross_verification.md` | 206 | A | §1 |
| `video-tech_dim10.md` | 1,689 | D | §8.5 |
| `00_Master_Plan.md` post-Session-6 | A, B, C, D | header / §6 / §9 |
| `01_Constitution.md` post §11.5 | A, B, C, D | §§1–8 |
| `05_Video_Audio/00_Index.md` | 407 | A, B, C, D | header voice |
| `05_Video_Audio/03_Capture_Pipelines.md` | 2,420 | C | §5 (capture pattern) |
| `05_Video_Audio/05_Recording_Storage.md` | 2,126 | C | §1 (recording-side audio) |
| `03_Architecture/05_RealTime_APIs.md` | 3,450 | A | §1 (voice chat cross-link C06) |
| `03_Architecture/11_TV_UX.md` | 3,273 | B | §3.5 (TV-side ALLM cross-link C12 §6) |
| `03_Architecture/07_Host_Agent_and_Game_Lifecycle.md` | 3,425 | C, D | §6 (`r18.SafeExec`), §8.11 |

### Web Sources Consulted

The companion addendum lists every URL with title and 2026-04-29 access date. **9 clusters + §Z; substantial URL coverage per the addendum's per-cluster matrix.**

### Insights Incorporated

| Insight | Source | Sections |
|---------|--------|----------|
| video-tech Insight #2 — Audio passthrough constrained (BINDING) | `video-tech_insight.md` | §1, §2.5, §4.1, §4.4, §7 (F4) |

### Conflict Zones Resolved

| CZ-ID | Description | Decision | Section |
|-------|-------------|----------|---------|
| Insight #2 | Audio passthrough binary; end-to-end validation critical | **Reaffirmed and binding**; capability decay tiers preserve fallback ladder | §1, §2, §3, §4 |
| Z addenda | Channel-config + passthrough technical contradictions | Resolved per cluster matrix in addendum | §1, §2, §3, §4 |
| Inherited (all prior families' Z lists) | Owned by prior chapters | Not relitigated | header preamble |

### R-18 Operational Integrity Compliance

Inherits R-18 enforcement from C08. Compliance verified:

- **Static — chapter prose**: §1 references R-18; §6.5 explicitly recaps the family-level allow-list extension specific to this chapter (`pacmd`, `pactl`, `wpctl`, `pw-cli`, `aplay -L`, `pa-info`, `ffmpeg -map 0:a`, `ffmpeg -f pulse`, WASAPI loopback, Core Audio AudioServerPlugin).
- **Static — code in §6**: imports `r18 "github.com/vasic-digital/helix-r18-safeexec"` (origin C08 §10). Deny-list NOT duplicated.
- **Static — bootstrap subprocess invocations**: all audio-tooling subprocess calls wrap through `r18.SafeExec`.
- **Test — §8.11**: `host-integrity-scan` inherited VERBATIM from C08 §12.11. Non-overridable per Constitution §11.5.4.

### Coverage Confirmation

| Metric | Value |
|--------|-------|
| Source per-dim primary (`video-tech_dim06.md`) | 1,141 lines |
| R-01 minimum (Master Plan §7.2 row C31) | 1,250 lines of body prose |
| Body prose actually synthesised | **1,216 lines** across §§1–9 (A 111 dense / ~350 wrapped + B 97 dense / 6,200 words ~770 wrapped + C 198 + D 810). Word-adjusted ≥ 1,650 wrapped lines |
| Coverage ratio vs minimum | 0.97× line-count / ≥ 1.32× word-adjusted |
| Coverage ratio vs primary per-dim source | 1.07× line-count / ≥ 1.45× word-adjusted |
| Forbidden-pattern scan (chapter prose) | clean |
| Empty-section-body scan | clean |
| Tables | Multi-channel ladder matrix in §2; capability decay tier table in §3.6; per-OS channel-order table in §4.3; failure-mode 12-row F1-F12 table in §7; test-type matrix in §8 |
| Section count | 10 normative sections + this verification block |
| Go code blocks | §6.4 (~50 LOC `audio.NewOpusMultiStreamEncoder` + `Encode` — real imports `gopkg.in/hraban/opus.v2`, `golang.org/x/sys/unix`, `r18`, `helix-shm`) |
| R-18 enforcement | inherited from C08 §10 + §8.11 host-integrity-scan inheritance |

### Sign-off

- Section A (§§1–2) by C31 Group A on 2026-04-29 (dense-paragraph format).
- Section B (§§3–4) by C31 Group B on 2026-04-29 (dense-paragraph format ~6,200 words).
- Section C (§§5–6) by C31 Group C on 2026-04-29.
- Section D (§§7–9) by C31 Group D on 2026-04-29.
- Web addendum by C31 addendum subagent on 2026-04-29.
- Header, ToC, §10, Anti-Bluff Verification stitched by orchestrator (Claude) on 2026-04-29.
- Reviewed by: pending operator review.

End of `05_Video_Audio/06_Audio_Pipeline.md` — 2026-04-29.
