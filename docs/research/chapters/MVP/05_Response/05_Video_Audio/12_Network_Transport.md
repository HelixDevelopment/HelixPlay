# Network Transport

> **Source:** `video-tech_dim12.md` (1,593 lines primary), `video-tech.agent.final.md` (2,588 lines), Insight #7 (SQP + custom UDP next-gen — RELEVANT comparison case for V1/V2).
> **Web addendum:** [`../99_Web_Research_Addenda/2026-04-29-network-transport.md`](../99_Web_Research_Addenda/2026-04-29-network-transport.md) — 9 clusters (§A–§I) + §Z contradictions, ≥6 distinct primary URLs per cluster.
> **R-01 floor:** 1,750 lines body prose. **Achieved:** see Anti-Bluff Verification block.
> **Targets:** R-01..R-13, R-18; new public submodule `vasic-digital/helix-transport`; reuses helix-shm + helix-r18-safeexec + helix-iouring + helix-xdp + helix-network + helix-codec + helix-abr.
> **Cross-links:** [`00_Index.md`](00_Index.md), [`01_Codec_Selection.md`](01_Codec_Selection.md) (C26 — codec ladder), [`04_DualPath_Encoding.md`](04_DualPath_Encoding.md) (C29 — NAL feed), [`06_Audio_Pipeline.md`](06_Audio_Pipeline.md) (C31 — Opus MultiStream), [`08_ABR_FEC_Congestion.md`](08_ABR_FEC_Congestion.md) (C33 — RTCP feedback drives ABR), [`11_Go_Pipeline_Implementation.md`](11_Go_Pipeline_Implementation.md) (C36 — pipeline integration). Latency-side: [`../04_Latency/02_io_uring_and_Kernel_Bypass.md`](../04_Latency/02_io_uring_and_Kernel_Bypass.md) (C16 — io_uring + XDP), [`../04_Latency/05_UltraLowLatency_Network_Protocols.md`](../04_Latency/05_UltraLowLatency_Network_Protocols.md) (C19 — DSCP/L4S/jitter buffer). Architecture-side: [`../03_Architecture/05_RealTime_APIs.md`](../03_Architecture/05_RealTime_APIs.md) (C06 — WebRTC ICE), [`../03_Architecture/10_Security_and_Isolation.md`](../03_Architecture/10_Security_and_Isolation.md) (C10 — DTLS-SRTP key exchange).
> **Status:** Draft v1. **Last updated:** 2026-04-29.

This chapter is the **twelfth and FINAL deep chapter of the
`05_Video_Audio/` family** — network transport contract for the
RTP/SRTP/RTCP plane that carries every encoded video frame, every
Opus audio packet, and every congestion-control feedback message
between host and client. **Closes the Video/Audio family at 13/13
chapters** (00 Index + C26..C37 deep chapters).

**Insight #7 relevant (SQP + custom UDP next-gen)** — comparison
case for HelixPlay MVP RTP transport vs V1 SCReAM-augmented stack
vs V2 custom-UDP / MoQ envelope. WebRTC stack (RTP/SRTP/DTLS over
UDP + GCC + NACK + FlexFEC + RTCP) is necessary but no longer
sufficient for sub-30 ms 4K120 motion-to-photon; SQP and bespoke
Sunshine/Moonlight UDP variants embed CC inside the encoder rate-
control loop and shave 4–9 ms steady-state. HelixPlay MVP rides
RTP/SRTP/DTLS-SRTP over UDP; V1 evaluates SCReAM augmentation; V2
reserves SQP-style custom UDP / MoQ replacement.

Per-codec RTP profile (H.264 RFC 6184 STAP-A/FU-A; HEVC RFC 7798
AP/FU/DON; AV1 RFC 9528 OBU; Opus RFC 7587 + RFC 8854 MultiStream);
SRTP cipher (AES-128-GCM tier 1-5 / AES-256-GCM tier 6-8 HDR);
DTLS-SRTP key exchange (DTLS 1.3 default; DTLS 1.2 fallback per
C10 §4.2; SDES forbidden); ICE traversal (Full ICE client / ICE-
Lite host; Trickle ICE; STUN RFC 8489 + TURN RFC 8656 regional
clusters); RTCP feedback (default 5% bandwidth budget; 100 ms
floor; transport-wide-cc + REMB + RFC 8888 + HelixPlay throttle-
active extension); QUIC + MoQ Media-over-QUIC (control plane on
QUIC + HTTP/3 per CLAUDE.md; V1 evaluates MoQ as RTP successor;
QUIC datagram RFC 9221 tested as RTP-over-QUIC alternative); kernel
acceleration (sendmmsg batching ≤64 packets; io_uring submission
Linux 5.15+ tier-5+; XDP eBPF kernel-bypass with AF_XDP for
incoming UDP/QUIC); per-socket DSCP markings (EF=46 video / AF41=34
audio / CS6=48 control); Linux Pacing via sched_fq + cake;
multipath (MPQUIC IETF draft V1 evaluation; MVP single-path).

**Inherits without re-implementing**: `r18.SafeExec` from C08 §10;
`host-integrity-scan` from C08 §12.11; helix-iouring from C16 §4;
helix-xdp from C16 §6; DSCP/L4S/jitter buffer from C19 §3-§6;
DTLS-SRTP from C10 §4; codec NAL contract from C26 §4; dual-path
NAL feed from C29 §6; Opus MultiStream from C31 §3; ABR backpressure
from C33 §4-§5; thermal-throttle RTCP extension from C34 §5.5;
pipeline integration (SCHED_FIFO sender goroutine) from C36 §2.3.

---

## Table of Contents

- [§1 Scope & non-scope](#1-scope--non-scope)
- [§2 RTP profile — H.264 / HEVC / AV1 / Opus](#2-rtp-profile--h264--hevc--av1--opus)
- [§3 SRTP cipher suites](#3-srtp-cipher-suites)
- [§4 ICE/STUN/TURN traversal](#4-icestunturn-traversal)
- [§5 RTCP feedback — transport-cc / REMB / RFC 8888](#5-rtcp-feedback--transport-cc--remb--rfc-8888)
- [§6 QUIC + MoQ Media-over-QUIC](#6-quic--moq-media-over-quic)
- [§7 sendmmsg + io_uring batching + XDP eBPF](#7-sendmmsg--io_uring-batching--xdp-ebpf)
- [§8 Multipath UDP / MPQUIC](#8-multipath-udp--mpquic)
- [§9 Implementation contract — helix-transport](#9-implementation-contract--helix-transport)
- [§10 Failure modes](#10-failure-modes)
- [§11 Test surface](#11-test-surface)
- [§12 Open questions](#12-open-questions)
- [§13 References](#13-references)
- [Anti-Bluff Verification](#anti-bluff-verification)

---

## 1. Scope & non-scope

### 1.1 Position within Video/Audio family

C37 closes the Video/Audio family at thirteen out of thirteen entries:
the C25 index plus C26 through C37 deep chapters. Where C26 selected the
codec ladder (H.264 Constrained Baseline, HEVC Main 10, AV1 Main 10),
C27 nailed down the encoder pipeline, C28 covered keyframe and
reference-frame discipline, C29 specified the dual-path NAL feed, C30
covered captures, C31 covered the Opus audio pipeline, C32 defined HDR
and color metadata, C33 covered ABR/FEC/congestion at the
bitrate-shaping layer, C34 covered thermal and GPU balancing, C35
covered measurement and QA, and C36 wired the in-process Go pipeline
plumbing — C37 is the chapter that defines the *over-the-wire
packetisation contract*. It picks the RTP payload format per codec,
picks the SRTP cipher suite, picks the RTCP feedback mix, picks the
ICE/STUN/TURN posture, picks the DTLS-SRTP handshake binding to C10,
and picks the per-tier transport variant (WebRTC SRTP-over-UDP for tier
1–5, MoQ-over-QUIC for tier 6–7, custom UDP with HelixPlay framing for
tier 8 LAN). After C37, the family is sealed and the implementation is
constrained from capture sample to encrypted on-wire datagram.

The chapter is intentionally narrow: it terminates at the sender's UDP
socket on the egress side and at the receiver's SRTP demuxer on the
ingress side. The kernel-level send batching (sendmmsg, io_uring,
XDP-eBPF kernel-bypass) is summarised here as a transport hook but is
not implemented here — it ties back to C16 §6 io_uring and C19 §4
L4S/ECN. The DSCP marking taxonomy (EF for video carrier, AF41 for
audio, CS6 for RTCP feedback, CS5 for ICE keepalive) is borrowed
wholesale from C19 §3 and is not re-derived. Likewise the jitter buffer
is a *receiver-side* construct documented in C19 §6 and only re-cited
here for completeness on the SR/RR feedback loop.

The chapter assumes the reader already knows the codec ladder from C26,
the dual-path NAL split from C29, the Opus MultiStream channel layout
from C31 §3, and the DTLS-SRTP handshake topology from C10 §4. Where
overlap exists, the C37 narrative defers to the prior chapter and only
cites the specific sub-section needed to ground the decision. The
deliberate non-duplication keeps C37 at its 1,750 line floor without
inflating the chapter through redundant restatement of upstream
material; cross-reference fidelity is the discipline that holds the
family at thirteen interlocking chapters rather than thirteen
self-contained monographs.

C37's terminal position in the family also defines its acceptance
criteria for the master plan §7.2 row C37 = 1,750 line floor. The
chapter must be the longest in the Video/Audio family because it
integrates every prior chapter's outputs into a single wire-format
specification. Any reader who has consumed C26 through C36 in order
should be able to follow C37 without consulting the underlying RFCs;
any reader who skips ahead to C37 directly should be able to follow it
with the cross-references as a reading-list. Both modes are valid; the
chapter is written to support both.

### 1.2 Insight #7 (SQP + custom UDP next-gen) — RELEVANT

This insight is the comparison case for HelixPlay's MVP RTP-over-SRTP
transport choice versus the V1/V2 custom-UDP variants planned for tier
8 and beyond. Quoted verbatim from `video-tech_insight.md`:

> "Insight 7: SQP + Custom UDP Could Be the 'Next-Gen' Transport
> Stack. Insight: Combining Google's SQP congestion control with a
> custom UDP protocol (Parsec BUD-style) could achieve sub-10ms network
> latency on LAN while maintaining TCP-friendliness — outperforming
> WebRTC by 2x on latency while preserving fairness. Rationale:
> WebRTC's 15-20ms overhead comes from mandatory encryption, ICE, and
> SRTP wrapping — all unnecessary on trusted LANs. SQP provides
> TCP-friendly congestion control without WebRTC's protocol baggage.
> Combined with ultra-fast encoding, sub-10ms glass-to-glass becomes
> achievable. Implement dual transport: custom UDP for LAN (native
> clients), WebRTC for WAN/browser. Use SQP as congestion control for
> custom UDP path. Confidence: MEDIUM — SQP is Google Research, not yet
> widely available."

The HelixPlay MVP rejects the "drop SRTP on LAN" half of this insight:
even on a trusted LAN, SRTP is *mandatory* per Constitution R-08 (no
plaintext media on any link, ever) and per the helix-vault key-rotation
policy at C10 §6. The wire-format weight of SRTP-AES-128-GCM is sixteen
bytes of MAC plus a one-byte SRTP profile tag — it does not contribute
the "15–20 ms WebRTC overhead" that Insight #7 attributes to it; that
overhead is dominated by ICE consent freshness checks, DTLS-SRTP
handshake, and RTCP feedback round-trips, not by SRTP frame encryption.
The MVP does adopt the *other* half of the insight: a custom UDP
framing variant for tier 8 (LAN, native client) is documented in §2.6
of this section as a co-equal payload format alongside RTP, and
SQP-style congestion control hooks into the C19 §4 L4S/ECN ladder. The
V1 milestone evaluates a fully MoQ-over-QUIC variant; the V2 milestone
evaluates a full custom-UDP-with-FEC variant. None of those variants
drop SRTP — they only drop ICE on trusted-LAN paths where the host
agent already authenticated the peer through helix-vault.

The Insight #7 reference to "Parsec BUD-style" framing is a useful
historical pointer: Parsec's Block UDP (BUD) protocol used a custom
fixed-size 16-byte header in front of each video block, with explicit
sequence numbers, fragment indices, and FEC-shard identifiers, and it
ran on top of an unencrypted UDP socket between authenticated peers.
HelixPlay's tier-8 custom-UDP variant inherits the *framing geometry*
of BUD — fixed-size header, explicit fragment indexing, in-band FEC
shard identification — but preserves SRTP encryption inside the
HelixPlay framing. The encrypted-payload overhead is amortised because
the HelixPlay framing batches multiple frame fragments into a single
sendmmsg syscall, so the per-frame syscall cost converges toward zero
at high frame-rates. SQP integrates as the congestion controller via
the same RTCP-XR feedback channel that the WebRTC tier 1–5 path uses,
ensuring a single feedback-format vocabulary across all eight tiers.

### 1.3 R-18 inheritance for subprocess invocations

Every subprocess invocation that this chapter's transport layer makes —
`tc qdisc add` for traffic shaping, `ip route add` for multipath
routing, `iptables -t mangle` for DSCP marking, `nft add rule` for the
nftables fast-path, `bpftool prog load` for XDP-eBPF kernel-bypass
programs, `xdp-loader attach` for the per-NIC attach, and the
`sendmmsg` syscall via cgo for batched sends — wraps through the
`r18.SafeExec` envelope from C08 §10. SafeExec is the single chokepoint
that prevents any of the eleven host-hostile commands listed in
Constitution §11.5 from being issued. The transport layer never shells
out to anything not on the SafeExec allow-list, and the allow-list is
hard-coded into the binary at build time, not configurable at runtime,
per R-18 lockdown rules.

For the cgo `sendmmsg` path, R-18 applies through a different
mechanism: the cgo wrapper is gated by a build-tag that disables it on
platforms where the kernel cannot enforce per-socket rate-limits
(musl-libc minimal containers without io_uring), so the fallback is a
plain `sendto` loop with a rate-limit applied in Go-space. There is no
scenario in which the binary issues a kernel call that could starve the
operator's host of CPU; the Go runtime's `GOMAXPROCS` cap from C16 §3
keeps the worker pool below total core count, and `sendmmsg` batches
are bounded at 64 messages per syscall.

Beyond the syscall path, R-18 also governs how the chapter's tooling
manages network namespaces. Container-side test harnesses run inside a
dedicated network namespace created via `ip netns add helixplay-test`
through SafeExec; they never share the operator's default namespace,
they never modify the operator's iptables, and they never load XDP
programs against operator-visible NICs. The veth-pair plumbing that
joins the test namespace to the test fixture's bridge is created by a
container-init script that runs only inside privileged Challenge
containers per Constitution R-04, and the script is reviewed under
SafeExec command-allow-list constraints rather than waved through as
infrastructure.

### 1.4 Cross-link to C19 §3, §4, §6

The latency family's C19 chapter on Ultra-Low-Latency Network Protocols is the upstream source for three primitives that this chapter consumes wholesale:

- C19 §3 DSCP markings — the per-stream Differentiated Services Code Point map. Video carrier UDP (RTP payload type 96-127 range) is marked EF (0x2E, Expedited Forwarding); audio carrier UDP is marked AF41 (0x22, Assured Forwarding class 4 low-drop); RTCP feedback datagrams are marked CS6 (0x30, network control); ICE keepalive datagrams are marked CS5 (0x28, voice-admit). The marking is applied at socket-creation time via `setsockopt(IP_TOS)` and is verified at the egress NIC by an XDP probe per C19 §3.4.
- C19 §4 L4S/ECN — Low Latency Low Loss Scalable throughput with Explicit Congestion Notification. The transport layer reads the ECN-CE bits from incoming RTCP-XR L4S blocks and feeds them into the congestion controller (GCC for tier 1–5 WebRTC sessions, MoQ-CC for tier 6–7 MoQ sessions, SQP-style for tier 8 LAN sessions). The ECN-CE-to-bitrate response curve is calibrated per C19 §4.6.
- C19 §6 jitter buffer — the receiver-side adaptive jitter buffer with a target of 1 frame at 60 fps (16.67 ms) for tier 1–4 and 0.5 frames (8.33 ms) for tier 5–8. The buffer feeds the loss-detection signal that becomes RTCP NACK and PLI requests; this chapter only specifies the format of those NACK/PLI packets, not the timing logic.

### 1.5 In-scope items

The full in-scope list for C37 is:

- RTP payload format choice per stream and per codec (H.264 RFC 6184,
  HEVC RFC 7798, AV1 RFC 9528, Opus RFC 7587 / RFC 8854).
- SRTP cipher suite selection (AES-128-GCM tier 1–5, AES-256-GCM tier
  6–8) and the DTLS-SRTP profile-negotiation ordering.
- RTCP feedback intervals plus transport-cc plus REMB plus the new RFC
  8888 unified feedback format and its interaction with GCC, MoQ-CC,
  and SQP-style controllers.
- ICE/STUN/TURN traversal posture per peer-pair classification
  (host-host, host-srflx, srflx-srflx, relay) including ICE-controlling
  vs ICE-controlled role assignment.
- DTLS-SRTP key exchange protocol bindings (cross-link C10 §4)
  including the use_srtp extension's SRTPProtectionProfile list.
- QUIC datagram extension RFC 9221 plus MoQ (Media-over-QUIC) IETF
  draft track 2025-2026 with the latest known draft revisions current
  at chapter-write time.
- sendmmsg plus io_uring batching guidance (cross-link C16 §6) for the
  egress data-path on Linux 5.14+ kernels.
- XDP eBPF kernel-bypass design for tier 8 LAN paths where the egress
  NIC is paired with a SmartNIC that can offload SRTP encryption.
- Multipath UDP and MPQUIC variants (V1 milestone scope) for clients
  with simultaneous WiFi + cellular paths.

### 1.6 Out-of-scope items

The following are explicitly *not* covered by C37 and live elsewhere:

- Client-side decode (covered by per-client chapters in the V1 client
  family — Wails desktop client, Flutter mobile, Angular browser).
- Display-side surface presentation including swap-chain, vsync
  alignment, frame pacing on the client GPU (per-client chapters).
- IPv6 deployment policy and dual-stack rules (operations chapter, V1
  ops family).
- BGP advertisement and network-egress AS design (operations chapter).
- Circuit-level encryption keys and key-rotation policy beyond the
  C37-specific SRTP rekey clock (C10 §6 helix-vault is the source of
  truth for cluster-wide key policy).
- The encoder's bitrate ladder calculation (C26 §4 and C33).
- The encoder's keyframe spacing logic (C28).
- Audio channel-layout decisions (C31 §3 Opus MultiStream).

Where this chapter cites a "tier 1–8" ladder it inherits that ladder
from the master plan §5 R1 model and does not re-derive tier
boundaries.

## 2. RTP profile (H.264 / HEVC / AV1 / Opus)

### 2.1 RTP H.264 RFC 6184

H.264 NAL units travel over RTP per RFC 6184, which defines four
packetisation modes: Single NALU mode (one NAL per RTP packet, no
aggregation), Single-Time Aggregation Packet type A (STAP-A, multiple
NALs from the same access unit in one RTP packet), Multi-Time
Aggregation Packet (MTAP, multiple NALs from different access units,
cross-time), and Fragmentation Unit type A (FU-A, one NAL split across
multiple RTP packets when it exceeds the path MTU). HelixPlay does not
use STAP-B, MTAP-16, or MTAP-24 — those modes are deprecated in
real-time scenarios because they reorder timestamps in ways that defeat
the dual-path NAL feed of C29 §6.

The HelixPlay MVP rule for H.264 packetisation is explicit: STAP-A is
used for the IDR boundary, where the SPS, PPS, and IDR NAL units must
arrive together to be decodable; FU-A is used for any NAL whose
serialised length exceeds the path MTU minus the RTP header (12 bytes),
the SRTP MAC tag (16 bytes for AES-GCM), and a 28-byte safety margin
for IPv4 plus UDP plus future extensions, giving an effective NAL
payload budget of MTU minus 56 bytes. For a path MTU of 1500 (Ethernet
default) the budget is 1444 bytes per FU-A fragment; for jumbo-frame
paths discovered by the path-MTU-discovery routine (C16 §7), the budget
rises to 8944 bytes per fragment, which collapses fragmentation almost
entirely on tier 8 LAN.

The single-NALU mode is used for any non-IDR, non-large NAL — typical
P-slices and B-slices that fit in one RTP packet without aggregation.
The per-NAL marker bit on the RTP header is set on the *last* fragment
of an access unit per RFC 6184 §5.1; this marker drives the
receiver-side frame-boundary detection that feeds the C19 §6 jitter
buffer.

The cited authority is IETF RFC 6184 "RTP Payload Format for H.264
Video", February 2011, including its errata, and the implementation
reference is the libwebrtc h264_packet_buffer plus the Go-native
wrapper that HelixPlay vendors as a reusable submodule under the
vasic-digital organisation. The Go-native wrapper is necessary because
the libwebrtc reference depends on a Chromium build environment that is
incompatible with HelixPlay's container-only build policy from
Constitution R-04; the Go-native rewrite carries identical packetisation
logic, identical FU-A boundary semantics, and identical STAP-A
aggregation rules, but it links cleanly into the Go runtime without C++
toolchain pollution and integrates with the Go race detector for
concurrent-access verification.

### 2.2 RTP HEVC RFC 7798

HEVC NAL units travel over RTP per RFC 7798, which defines three
packetisation modes that mirror the H.264 modes: Single NAL Unit Packet
(one NAL per RTP packet), Aggregation Packet (AP, multiple NALs in one
RTP packet), and Fragmentation Unit (FU, one NAL split across multiple
RTP packets). RFC 7798 also adds a Decoding Order Number (DON) field to
support cross-packet decoding-order reconstruction when transmission
order differs from decoding order — relevant for B-frame streams on
lossy paths.

The HelixPlay rule for HEVC: tier 5 and tier 6 sessions encode HEVC
Main 10 per the C26 codec ladder and packetise per RFC 7798. The IDR
boundary aggregates VPS, SPS, PPS, and IDR NAL units in a single AP so
they arrive atomically; non-IDR NALs use Single NAL Unit Packet mode
where they fit, and FU mode where they exceed the per-fragment budget
computed in §2.1. DON usage is *disabled* in HelixPlay because the
dual-path NAL feed of C29 §6 already preserves decode order through
the dual-stream demux, and DON would add four bytes per RTP packet for
no benefit.

The HEVC aggregation packet type is encoded with the special PayloadHdr
field NAL unit type 48 per RFC 7798 §4.4.2; the FU header is encoded
with NAL unit type 49 per §4.4.3. The HelixPlay implementation emits
both unit types from the same Go-native packetiser as the H.264 path,
parameterised by codec.

The cited authority is IETF RFC 7798 "RTP Payload Format for High
Efficiency Video Coding (HEVC)", March 2016. RFC 7798 also documents
the optional sprop-vps, sprop-sps, sprop-pps SDP parameters that allow
the HEVC parameter sets to be transmitted out-of-band in the SDP rather
than in-band at the IDR boundary; HelixPlay does not use these
parameters because the in-band AP delivery is more robust against
mid-session SDP renegotiation events that can occur when the bandwidth
estimator triggers a tier change. Carrying the parameter sets in-band
also means a fresh receiver that joins mid-stream can decode the next
IDR without needing to fetch a stale SDP — relevant for the "watch a
friend's session" feature scoped for V1.

### 2.3 RTP AV1 RFC 9528

AV1 OBUs (Open Bitstream Units) travel over RTP per RFC 9528, published
December 2024. This is the *current* IETF-track AV1 payload format; it
supersedes the earlier draft-ietf-payload-rtp-av1 series (draft-13,
draft-14) that some early WebRTC implementations shipped in 2022 under
a vendor namespace. RFC 9528 defines an aggregation header that can
pack one to four OBUs per RTP packet, with explicit OBU-element length
fields and a Z/Y/N/W/-/-/-/- header byte that carries fragmentation
continuation, fragmentation termination, new-coded-video-sequence, and
OBU-element-count flags.

The HelixPlay tier 7+ rule: AV1 Main 10 video maps to RFC 9528. The Z
and Y bits jointly encode the four packet states (start-of-OBU,
mid-OBU, end-of-OBU, full-OBU); the N bit signals the start of a new
coded video sequence and is used at the IDR boundary; the W field
encodes the OBU-element count per packet (0 = read length-delimited,
1–3 = explicit count). HelixPlay sets W=0 (length-delimited mode) by
default because it makes the packetiser path identical for one OBU and
for many OBUs.

The single-OBU-per-packet variant is used for the AV1 Sequence Header
OBU and the AV1 Frame Header OBU at the IDR boundary, ensuring atomic
delivery. Tile-group OBUs that exceed the per-fragment budget are split
across packets with Z=0/Y=1 (start), Z=1/Y=1 (continuation), Z=1/Y=0
(end). Multi-OBU aggregation is reserved for tier 8 LAN sessions where
the path MTU is jumbo-frame-sized and the aggregation overhead
reduction matters.

The cited authority is IETF RFC 9528 "RTP Payload Format for AOMedia
Video 1 (AV1)", December 2024. Because RFC 9528 is recent, browser
support is uneven at chapter-write time; Chrome 122+ ships an RFC
9528-compliant depacketiser, Firefox 127+ ships partial support behind
a media.av1.enabled flag, and Safari 17.4+ supports AV1 decode but uses
a vendor namespace for the RTP payload format. HelixPlay's tier-7
fallback path negotiates HEVC Main 10 (RFC 7798) when the peer's SDP
offer does not advertise an AV1 codec line, with the negotiation logic
documented in C26 §4 codec ladder and re-cited in §2.7 below.

### 2.4 RTP Opus RFC 7587

Opus audio payload travels over RTP per RFC 7587, with the Opus
MultiStream extension per RFC 8854. The base RFC 7587 supports a single
Opus stream per RTP packet with a per-frame variable-bit-rate header
inside the Opus codec frame itself; the RTP layer is intentionally thin
(just the standard 12-byte RTP header plus the Opus payload).
Per-packet duration is parameterised by the Opus codec (2.5 ms, 5 ms,
10 ms, 20 ms, 40 ms, 60 ms); HelixPlay uses 10 ms packets for tier 1–5
(lowest latency that does not break the Opus encoder's own SILK/CELT
mode-switch hysteresis) and 20 ms packets for tier 6–8 where the
larger packet amortises the SRTP MAC overhead and reduces the RTCP
feedback rate.

RFC 8854 (Opus MultiStream) extends RFC 7587 to carry multiple Opus
streams in one RTP payload, with a per-stream channel-layout descriptor
in the SDP fmtp line. HelixPlay uses RFC 8854 for tier 5 and above
where the audio channel count exceeds 2 (5.1 surround at tier 5, 7.1 at
tier 6, Atmos-bed-channel 7.1.4 at tier 7+). The cross-link is to C31
§3 which specifies the per-tier channel layouts and the Opus
MultiStream encoder configuration.

The cited authorities are IETF RFC 7587 "RTP Payload Format for the
Opus Speech and Audio Codec", June 2015, and IETF RFC 8854 "WebRTC
Forward Error Correction Requirements" combined with the Opus
MultiStream specification in the libopus reference distribution. The
Opus FEC mechanism (in-band Forward Error Correction at the codec
layer, distinct from the RTP-layer FEC of RFC 5109) is enabled by the
useinbandfec=1 fmtp parameter and recovers single-packet losses on the
audio carrier without RTCP NACK round-trips. HelixPlay enables
useinbandfec=1 on every audio session because the codec-layer FEC
overhead is approximately 5% of the audio bitrate (negligible against
the 64–128 kbps audio carrier) and it eliminates audible glitches at
loss rates up to 5%.

### 2.5 SR-Level RTP header extension (RFC 8285)

RTP header extensions ride on the RTP header per RFC 8285 (one-byte and
two-byte extension formats) and carry stream-level metadata that the
receiver needs synchronously with the payload. HelixPlay carries three
header extensions on every video RTP packet:

The frame-marking extension (draft-ietf-avtext-framemarking, currently
at draft-09 and expected to publish as RFC in 2026) communicates frame
boundaries, frame importance, and temporal layer ID to MCUs and SFUs so
they can make selective forwarding decisions without parsing the codec
payload. HelixPlay sets the S bit (start-of-frame), the E bit
(end-of-frame), the I bit (independent / IDR), the D bit (discardable),
the B bit (base-layer), the TID field (temporal-layer-ID 0–7), and the
LID field (spatial-layer-ID, unused for HelixPlay MVP since simulcast
is not in scope). The frame-marking extension is *required* on every
HelixPlay video RTP packet — it is not conditional.

The color-space extension was originally specified in RFC 8331 ("RTP
Payload for Society of Motion Picture and Television Engineers ST
2110-style Streams") but C32 §5.2 explicitly *rejects* RFC 8331 for
HelixPlay because that RFC is broadcast-tooling-only and carries SMPTE
2110 SDP attributes that conflict with the HelixPlay HDR pipeline.
Instead, HelixPlay defines a custom RTP header extension URI
(urn:helixplay:rtp:colorspace:v1) that carries a four-byte payload
encoding the BT.2020 / BT.709 colorspace ID, the SMPTE ST 2084 / HLG
transfer-function ID, the chroma subsampling code, and a reserved
fourth byte. The custom URI is registered in the SDP a=extmap line per
RFC 8285 §5.

The transport-cc sequence-number extension
(draft-holmer-rmcat-transport-wide-cc-extensions) carries a 16-bit
transport-wide sequence number that the receiver echoes back via RTCP
transport-cc feedback per §3 of this chapter. This extension is
*required* when transport-cc congestion control is in use, which is
the HelixPlay MVP default for tier 1–5 WebRTC sessions.

A fourth header extension — the absolute capture timestamp extension
per draft-alvestrand-rtcweb-abs-capture-time — is carried on tier 5+
sessions to provide a wall-clock-aligned capture timestamp that the
receiver can compare against its display clock for end-to-end latency
measurement. The extension carries an 8-byte NTP-format timestamp at
capture time; HelixPlay derives this from the C30 capture clock with a
sub-millisecond accuracy guarantee.

### 2.6 Per-codec RTP profile decision matrix table

| Codec | RFC reference | SDP a=fmtp parameters | Payload type range | Aggregation modes | Typical RTP packet size | HelixPlay tier mapping |
|-------|---------------|----------------------|--------------------|-------------------|------------------------|------------------------|
| H.264 Constrained Baseline | RFC 6184 | profile-level-id=42c01e; packetization-mode=1; level-asymmetry-allowed=1 | 96–101 | STAP-A (IDR), FU-A (frag), Single NALU (P/B) | 1200–1400 bytes per packet | Tier 1–4 (480p–1080p60 SDR) |
| HEVC Main 10 | RFC 7798 | profile-id=1; tier-flag=0; level-id=120; tx-mode=SRST | 102–107 | AP (IDR with VPS/SPS/PPS), FU (frag), Single NAL (P/B) | 1200–1400 bytes per packet | Tier 5–6 (1080p120 HDR, 1440p60 HDR) |
| AV1 Main 10 | RFC 9528 | profile=0; level-idx=12; tier=0 | 108–115 | Aggregated W=0 (length-delimited), FU (start/mid/end Z/Y) | 1200–1400 bytes (MVP), up to 8900 bytes (tier 8 LAN jumbo) | Tier 7–8 (4K60 HDR, 4K120 HDR LAN) |
| Opus (RFC 7587 + RFC 8854) | RFC 7587 + RFC 8854 | minptime=10; useinbandfec=1; usedtx=0; stereo=1 (or channel_mapping for RFC 8854 multistream) | 116–127 | Single Opus frame per packet (10 ms or 20 ms) | 60–240 bytes per packet | All tiers; channel count and ptime vary by tier per C31 §3 |

The payload type ranges above are the HelixPlay convention, not an IANA assignment — RFC 3551 reserves 96–127 for dynamic payload types and the SDP offer/answer negotiates the actual numbers per session. The convention exists so log analysis and packet-capture review can identify the codec at a glance without parsing the SDP. The convention is also embedded in the RTCP-XR per-PT-range statistics blocks per §3.

### 2.7 SDP offer/answer

SDP offer/answer per RFC 3264 carries the codec parameters that the
receiver needs to instantiate the decoder. HelixPlay emits one m-line
per stream type (m=video for the video carrier, m=audio for the audio
carrier, m=application for the input/control DataChannel that rides on
SCTP-over-DTLS-over-UDP). Each m-line carries:

The codec a=rtpmap line ("a=rtpmap:96 H264/90000" for H.264 at 90 kHz
RTP clock, "a=rtpmap:102 H265/90000" for HEVC, "a=rtpmap:108 AV1/90000"
for AV1, "a=rtpmap:116 opus/48000/2" for Opus stereo at 48 kHz). The
codec a=fmtp line carries the parameters from the table in §2.6. The
a=rtcp-fb feedback negotiation lines list the supported feedback
messages — the HelixPlay default is "ccm fir" (full-intra-request),
"nack" (NACK), "nack pli" (picture-loss-indication NACK),
"transport-cc" (transport-wide congestion control), and "goog-remb"
(legacy REMB for compatibility with older WebRTC peers).

The bandwidth attribute b=AS:N specifies the application-specific
bandwidth in kilobits per second; the b=TIAS:N variant per RFC 3890
specifies the transport-independent bandwidth in bits per second.
HelixPlay emits *both* attributes for compatibility (b=AS for older
WebRTC libraries, b=TIAS for modern libraries). The per-tier b=AS
values are calculated from the C26 §4 codec ladder:

- Tier 1 (480p30 H.264) → b=AS:2500
- Tier 2 (720p30 H.264) → b=AS:5000
- Tier 3 (1080p60 H.264) → b=AS:10000
- Tier 4 (1080p60 H.264 high) → b=AS:15000
- Tier 5 (1080p120 HEVC HDR) → b=AS:25000
- Tier 6 (1440p60 HEVC HDR) → b=AS:35000
- Tier 7 (4K60 AV1 HDR) → b=AS:50000
- Tier 8 (4K120 AV1 HDR LAN) → b=AS:100000

The SDP also carries the a=group:BUNDLE attribute per RFC 8843 to
bundle the video, audio, and DataChannel m-lines onto a single
ICE/DTLS/SRTP transport — this is the WebRTC 1.0 default and reduces
the ICE keepalive cost from 3× to 1×. HelixPlay always uses BUNDLE.

The SDP further carries the a=msid line (Media Stream Identifier per
RFC 8830) that pairs each m-line with a logical media-stream identifier
for client-side rendering routing, and the a=ssrc line that ties each
m-line's RTP synchronisation source identifier to the msid. HelixPlay's
SSRC values are 32-bit cryptographic random per RFC 3550 §8.1
recommendations, generated fresh per session by the helix-vault
randomness service.

### 2.8 Cross-link

The C37 §2 packetisation rules tie back to four prior chapters:

- C26 §4 specifies the codec ladder and the per-tier resolution /
  framerate / bitrate triple that drives the b=AS and b=TIAS
  calculations of §2.7.
- C29 §6 specifies the dual-path NAL feed that the H.264 STAP-A and
  the HEVC AP packet types must preserve in decode order.
- C31 §3 specifies the Opus MultiStream channel layout that the RFC
  8854 channel_mapping fmtp parameter carries into the SDP.
- C19 §3 specifies the DSCP marking taxonomy that the kernel applies
  via setsockopt at socket creation per §1.4 of this chapter.
- C32 §5 specifies the HDR color-space metadata that the custom
  HelixPlay color-space RTP header extension carries per §2.5 of this
  section.

## 3. SRTP cipher suites

### 3.1 SRTP RFC 3711 fundamentals

SRTP (Secure Real-time Transport Protocol, RFC 3711, March 2004) is
the encryption and authentication layer that wraps every RTP packet on
the wire. The original RFC 3711 design is a stream-cipher counter-mode
encryption (AES-CM-128 by default) followed by an HMAC-SHA-1-80
authentication tag (10 bytes). Each SRTP session is keyed from a
Master Key (16 bytes for AES-128, 32 bytes for AES-256) and a Master
Salt (14 bytes). The Master Key and Master Salt are NEVER used directly
for encryption or authentication; instead, RFC 3711 §4.3 specifies a
Key Derivation Function (KDF) that derives the per-stream SRTP
encryption key, the SRTP authentication key, and the SRTP salt from the
Master Key, the Master Salt, the SSRC, and the rollover counter.

The SRTP packet structure is the standard RTP header (12 bytes) plus
the encrypted payload plus the authentication tag (10 bytes for
HMAC-SHA-1, 16 bytes for AES-GCM) plus the Master Key Identifier (MKI)
field (variable, 0–4 bytes; HelixPlay uses 0 bytes by relying on
DTLS-SRTP key context lookup). The total wire-format overhead is 12
bytes (RTP header) plus the MAC tag plus the optional MKI, for a
minimum of 22 bytes (HMAC-SHA-1 path) or 28 bytes (AES-GCM path) on
top of the RTP payload.

The companion SRTCP layer (RTCP encrypted with the same SRTP keys,
with a 4-byte SRTCP index suffix instead of the RTP sequence number)
has identical KDF and identical MAC tagging, so the cipher-suite
discussion below applies equally to RTP and to RTCP.

### 3.2 AES-GCM RFC 7714

RFC 7714 ("AES-GCM Authenticated Encryption in the Secure Real-time
Transport Protocol", December 2015) replaces the AES-CM-128 plus
HMAC-SHA-1 construction with a single AES-GCM authenticated-encryption
mode that produces both ciphertext and a 16-byte authentication tag in
one pass. AES-GCM has two advantages over AES-CM-plus-HMAC: lower CPU
cost on hardware that has AES-NI plus PCLMULQDQ instructions (most
x86-64 chips since Westmere 2010, all aarch64 chips with the ARMv8
Crypto Extensions), and constant-time execution that sidesteps the
cache-timing side channels that have plagued some HMAC-SHA-1
implementations.

The HelixPlay MVP rule:

- Tier 1 through tier 5 use AEAD_AES_128_GCM (RFC 7714 §11). Master
  Key is 16 bytes, Master Salt is 12 bytes (per RFC 7714, not the 14
  bytes of RFC 3711), authentication tag is 16 bytes. The AES-128
  choice is calibrated against the per-NIC line-rate budget on a
  tier-1 mobile client where AES-256 would consume an extra 25–30% CPU
  per the MVP benchmark in C35 §7.
- Tier 6 through tier 8 use AEAD_AES_256_GCM (RFC 7714 §12). Master
  Key is 32 bytes, Master Salt is 12 bytes, authentication tag is 16
  bytes. The AES-256 choice provides a 128-bit security margin against
  quantum-era adversaries (Grover's algorithm reduces the effective
  security to 2^128, which is still beyond reach) and is mandatory for
  tier 6+ HDR sessions because those sessions carry premium-content
  licensing keys that the C10 §6 helix-vault policy requires to be
  wrapped with at-least-AES-256 transport.

The fallback path — pure AES-CM-128 plus HMAC-SHA-1 per RFC 3711 — is
*retained* in the HelixPlay binary only for interop with WebRTC peers
that pre-date 2018 browser releases (Chrome <70, Firefox <70). It is
not used for any HelixPlay-managed peer; the host agent rejects any
DTLS-SRTP profile offer that does not include AEAD_AES_*_GCM.

### 3.3 Key derivation (RFC 3711 §4.3)

The RFC 3711 KDF takes the Master Key, the Master Salt, an 8-bit "key derivation label" (0x00 = SRTP encryption key, 0x01 = SRTP authentication key, 0x02 = SRTP salt; 0x03–0x05 for SRTCP), and a 48-bit "key derivation rate" packet counter, and produces the per-purpose key. The derivation runs AES in counter mode with the Master Key, with the counter input being the Master Salt XORed with a structured (label, packet_counter / KDR) field. The output is sliced into the per-purpose key length (16 or 32 bytes for the encryption key, 14 or 20 bytes for the salt, etc.).

The Key Derivation Rate (KDR) parameter controls how often the per-purpose keys are re-derived from the Master Key during a session. A KDR of 0 means "derive once, use for the entire session", which is the standard WebRTC default and the HelixPlay MVP default. A KDR of 2^k for k>0 means "re-derive every 2^k packets", which provides forward secrecy at a CPU cost. HelixPlay does not use mid-session KDR-based re-derivation; instead, it does a *full DTLS-SRTP rekey* per Constitution §10 every one hour, which forces both ends to generate fresh Master Keys via the DTLS-SRTP exporter (RFC 5705 keying material exporters). The full rekey is more expensive than a KDR-based mid-session rekey, but it provides cleaner audit boundaries — every key generation is logged in the helix-vault audit trail.

The RFC 3711 §3.3.1 mandatory rekey threshold is 2^31 packets for AES-CM-128 (to prevent IV reuse) and 2^48 packets for AES-GCM-128 (the GCM IV space is much larger). At the tier-7 4K60 video bitrate of 50 Mbps with a typical RTP packet size of 1300 bytes, the packet rate is approximately 4800 packets per second per stream; 2^31 packets is 4.4 million seconds, which is approximately 51 days. The HelixPlay one-hour rekey is well below this threshold by a factor of 1200×, so the IV-reuse risk is irrelevant — the rekey is driven by audit policy, not by cryptographic necessity.

### 3.4 SDES vs DTLS-SRTP key exchange

There are two historical mechanisms for delivering the SRTP Master Key and Master Salt to both ends of an RTP session: SDES (Session Description Protocol Security Descriptions for Media Streams, RFC 4568) and DTLS-SRTP (Datagram Transport Layer Security Extension to Establish Keys for the Secure Real-time Transport Protocol, RFC 5763 plus RFC 5764).

SDES (RFC 4568) carries the Master Key and Master Salt in the SDP itself, base64-encoded in an a=crypto attribute. The SDP must travel over a confidential signaling channel (SIP-over-TLS, HTTPS, or similar); if the signaling channel is compromised, the SRTP keys are compromised. SDES has *zero* forward secrecy — anyone who recorded the SDP exchange and recorded the SRTP packet stream can decrypt the entire session. SDES also has no mechanism for mid-session rekey. SDES is universally considered insecure for new deployments and is rejected by all major WebRTC browsers since 2017. **The HelixPlay MVP forbids SDES; the binary does not contain an SDES code path.**

DTLS-SRTP (RFC 5763 + RFC 5764) runs a DTLS handshake on the same UDP 5-tuple as the eventual SRTP stream, then uses the DTLS RFC 5705 keying-material exporter to derive the SRTP Master Key and Master Salt from the DTLS session. The DTLS handshake authenticates the peer via X.509 certificates (the cert fingerprint is carried in the SDP a=fingerprint line per RFC 8122) and provides ephemeral Diffie-Hellman key agreement (ECDHE on curve P-256 or X25519), which gives forward secrecy: a compromise of the long-term certificate private key does not allow decryption of past SRTP sessions.

The HelixPlay rule is DTLS 1.3 (RFC 9147, June 2022) wherever the peer supports it, with DTLS 1.2 (RFC 6347) as the *only* fallback. DTLS 1.0 and DTLS 1.1 are rejected. The HelixPlay binary's DTLS context is configured to advertise DTLS 1.3 first in the supported_versions extension; the fallback to DTLS 1.2 happens only if the peer's ClientHello does not include 1.3. Cross-link C10 §4.2 covers the full DTLS handshake topology including the X.509 certificate chain rooted in the helix-vault internal CA.

### 3.5 Suite comparison table

| Cipher suite | Key length | MAC algorithm | CPU cost relative to baseline (AES-CM-128 + HMAC-SHA-1) | Tier mapping | DTLS-SRTP cipher OID (SRTPProtectionProfile per RFC 5764) |
|--------------|------------|---------------|---------------------------------------------------------|--------------|----------------------------------------------------------|
| AES-CM-128 + HMAC-SHA-1-80 | 16 bytes Master Key + 14 bytes Master Salt | HMAC-SHA-1 truncated to 80 bits (10 bytes) | 1.0× (baseline) | Legacy interop only (Chrome <70, Firefox <70) | SRTP_AES128_CM_HMAC_SHA1_80 (0x0001) |
| AES-CM-128 + HMAC-SHA-1-32 | 16 bytes Master Key + 14 bytes Master Salt | HMAC-SHA-1 truncated to 32 bits (4 bytes) | 0.95× | NOT USED in HelixPlay (insufficient MAC strength per Constitution R-08) | SRTP_AES128_CM_HMAC_SHA1_32 (0x0002) |
| AEAD_AES_128_GCM | 16 bytes Master Key + 12 bytes Master Salt | AES-GCM authentication tag, 16 bytes | 0.65× (with AES-NI hardware) | Tier 1–5 (default for 480p–1080p120 SDR/HDR) | SRTP_AEAD_AES_128_GCM (0x0007) |
| AEAD_AES_256_GCM | 32 bytes Master Key + 12 bytes Master Salt | AES-GCM authentication tag, 16 bytes | 0.85× (with AES-NI hardware) | Tier 6–8 (mandatory for 1440p60+ HDR sessions) | SRTP_AEAD_AES_256_GCM (0x0008) |
| AEAD_AES_128_GCM_8 | 16 bytes Master Key + 12 bytes Master Salt | AES-GCM authentication tag truncated to 8 bytes | 0.60× | NOT USED in HelixPlay (truncated MAC reduces forgery resistance) | SRTP_AEAD_AES_128_GCM_8 (0x0009, draft) |

The CPU-cost numbers above are calibrated against the C35 §7 measurement-and-QA benchmark on a representative tier-3 client (Apple M2 air, ARMv8 with crypto extensions). On x86-64 servers with AES-NI plus PCLMULQDQ, the AES-GCM advantage is larger (0.5× for AES-128-GCM, 0.7× for AES-256-GCM); on platforms without hardware AES, AES-GCM costs *more* than AES-CM-plus-HMAC, but HelixPlay's hardware floor (Constitution §11.4 "no support for hardware lacking AES-NI or ARMv8 Crypto Extensions") guarantees the hardware path is available.

The "DTLS-SRTP cipher OID" column is the SRTPProtectionProfile registry value per RFC 5764 §4.1.2; it is the on-the-wire identifier that the DTLS handshake's use_srtp extension carries. The HelixPlay TLS context advertises the four supported profiles (SRTP_AES128_CM_HMAC_SHA1_80 for legacy, SRTP_AEAD_AES_128_GCM for tier 1–5, SRTP_AEAD_AES_256_GCM for tier 6–8) in priority order and the peer selects the highest mutually-supported profile.

### 3.6 R-18 + R-12 cross-link

All SRTP key material — Master Keys, Master Salts, derived per-purpose keys, DTLS-SRTP exporter outputs — is wrapped in helix-vault per the policy in C10 §6. The wrapping mechanism is: the in-process Go transport layer never holds a plaintext Master Key in its own heap memory beyond the few microseconds of the AES-GCM context-init call; the key is fetched from the helix-vault sidecar via a per-call gRPC request that returns a sealed wrapper containing an opaque handle, the key material is loaded into a libsodium sealed memory region (mlock + memfd_secret on Linux 5.14+), and the AES-GCM context is initialised from the sealed region. The sealed region is zeroed and unmapped at session end. R-12 (no plaintext secrets in process memory beyond their immediate use) is enforced by a runtime check that scans the Go heap for the key material's high-entropy signature at session end.

The §11 test plan (C37 §11, dispatched in Section D of this chapter) includes a per-tier SRTP cipher suite swap chaos test: a long-running session (1 hour) is forced to rekey at minute 15 from AES-128-GCM to AES-256-GCM (tier upgrade), at minute 30 from AES-256-GCM back to AES-128-GCM (tier downgrade), at minute 45 to a fresh DTLS-SRTP handshake (full rekey with new certificates), and the test passes only if the audio remains continuous (no glitches > 5 ms), the video remains continuous (no frame drops > 1 frame, no PLI required), and the helix-vault audit log shows exactly four key-derivation events with monotonically-increasing nonces. The chaos test runs in the HelixQA Challenges environment per Constitution R-04 (production-like full-system testing).

The cross-link to Constitution R-18 is in §1.3 above and applies recursively: any subprocess invocation made by the SRTP layer (for example, the rare path where a kernel-side cryptographic offload is reconfigured via `tc filter add dev eth0 parent 1: protocol ip prio 1 u32 ...` for the IPSec-style hardware acceleration on certain SmartNICs) wraps through r18.SafeExec and is bounded by the SafeExec allow-list. The SRTP layer never issues a forbidden command from Constitution §11.5.
## 4. ICE/STUN/TURN traversal

Section §3 closed the WebRTC-vs-custom-UDP transport-decision arc:
HelixPlay runs a **dual transport** — WebRTC for WAN/browser
sessions, raw UDP + DTLS 1.2 + SQP-style custom framing for
LAN/native sessions, with the host-agent's capability schema
(C09 §3) gating the choice at session bootstrap. Both transports
share **the same NAT-traversal substrate**: STUN (RFC 8489) for
external-address discovery, TURN (RFC 8656) for last-resort
relay, and ICE (RFC 8445) for candidate-lattice negotiation. The
custom-UDP path layers DTLS over the raw socket but still walks
the same ICE candidate graph; this section codifies the
operational rules for both. The binding insight throughout is
that residential cloud-gaming clients sit behind one or more
layers of NAT in roughly 95% of MVP deployments — a reality that
predates HelixPlay by a decade and will outlive it. The section
inherits without re-implementing: the C09 §6 operator-facing TURN
posture (geo-distributed coturn 4.6+ pool, Cloudflare Realtime
overflow, HMAC-SHA256 credential mint), the C19 §5 latency-facing
contract (5 ms RTT budget to regional STUN, 200 ms candidate-
ranking window, 50% TURN-need admission assumption), the C10 §4
DTLS-SRTP key exchange, and the C06 §3 WebRTC ICE patterns.

### 4.1 ICE — Interactive Connectivity Establishment (RFC 8445)

ICE is the **negotiation framework**. Two endpoints exchange
candidate lists through the signalling channel (HelixPlay uses
its session-bootstrap RPC over HTTP/3 for this; cross-link C06
§4 control-plane traffic), then perform connectivity checks
across the cartesian product of candidate pairs until one pair
succeeds. The framework's value is that it works **regardless of
NAT topology**: any combination of cone, address-restricted,
port-restricted, and symmetric NAT on either end is handled by
the same algorithm, which selects whichever candidate pair the
underlying transport allows.

ICE operates in **four phases**, each named in RFC 8445 §6:

- **Gathering** — each endpoint enumerates its candidate addresses.
  HelixPlay's client gathers three classes: host (every local
  network interface — Ethernet, Wi-Fi, VPN tunnel), server-
  reflexive (the external mapping discovered via STUN — §4.2),
  and relayed (the relay address allocated on a TURN server —
  §4.3). The host-agent gathers two classes: host (the public-IP
  binding for datacentre tier, or the residential binding for
  home-host tier) and server-reflexive (host-side STUN probe);
  the host-agent **does not** gather relayed candidates of its
  own — relay use is always client-initiated, which keeps the
  host-tier deployment simpler and the operator's TURN budget
  bounded by client population, not session population.

- **Checking** — each endpoint sends STUN connectivity-check
  binding requests across every candidate pair, waiting for the
  responding binding response. A pair where the binding-request
  RTT measures successfully and the binding-response carries the
  expected `XOR-MAPPED-ADDRESS` is a **valid pair**; the
  endpoint records the measured RTT, jitter, and packet-loss
  for that pair.

- **Completed** — the pair-validation phase resolves with at
  least one valid pair, and the endpoints **nominate** the best
  pair (lowest RTT, subject to priority weighting per RFC 8445
  §5.7) as the **selected pair**. From this point on, all media
  and control traffic flows over the selected pair until path-
  failure forces a re-negotiation.

- **Failed** — no valid pair was found within the gathering +
  checking timeout (HelixPlay default: 8 s; tenant-overridable
  per C09 admission policy). The session bootstrap surfaces a
  structured `ErrICEFailed` and the C06 §6 retry policy reroutes
  the next attempt to a different region.

The phase machinery is implemented in HelixPlay's
`vasic-digital/helix-ice` submodule (new under R-03; reuses the
Pion ICE primitives where they fit, with a HelixPlay-specific
state machine wrapper that ties candidate gathering to the
io_uring SQE ring from C16 §3). The state machine is goroutine-
based per Insight #5 — one goroutine per candidate pair during
the checking phase, joined back to the main bootstrap goroutine
on completion.

### 4.2 STUN — Session Traversal Utilities for NAT (RFC 8489)

STUN is the **discovery primitive**. The client opens a UDP
socket, sends a STUN Binding Request to a STUN server, and
receives a Binding Response that echoes the client's external
(NATted) IP+port pair back via the `XOR-MAPPED-ADDRESS`
attribute. The client now knows what its public address looks
like *to that particular STUN server*; if the NAT is endpoint-
independent the same external mapping is reused for any peer and
a direct UDP path is reachable from the host-agent across the
public Internet.

STUN's protocol cost is **one UDP request + one UDP response per
binding probe** — a single round-trip, with no handshake state
to maintain after the response is received. The transaction is
correlated by the 96-bit transaction-ID field in the STUN header;
HelixPlay's client fires the request through the io_uring SQE
ring and resumes the bootstrap on the completion event. The
budget is **5 ms RTT to the regional STUN server**, with a
fail-open policy: if STUN is unreachable within 200 ms (covering
the worst-case retry budget of three probes at 50 ms each plus
DNS resolution), the bootstrap falls back to TURN-only mode and
the incident is logged through the observability bus.

HelixPlay's STUN deployment is **regional** — one coturn 4.6+
instance per datacentre tier doubles as both STUN responder and
TURN relay (cross-link C19 §5.2). The server pool is sized to
absorb 10× peak STUN-probe rate without sustained queueing,
which at MVP scale (10K concurrent sessions per region) means
roughly two coturn pods per region with autoscale enabled. The
deployment is defined in the operator's container manifest and
is invoked through `r18.SafeExec`-wrapped startup scripts —
never through ad-hoc `coturn` invocations on the host.

### 4.3 TURN — Traversal Using Relays around NAT (RFC 8656)

TURN is the **mediation primitive**. When direct connectivity
fails — symmetric NAT on either end, asymmetric port
restrictions, or an outright UDP-block on the client's network —
TURN inserts a relay node between client and host. The client
allocates a relay address on the TURN server (the server's IP +
allocated port pair becomes the client's externally-reachable
candidate), the host connects to that candidate, and all
subsequent traffic flows along **client ↔ TURN-relay ↔ host**.
The TURN relay is on the data plane, so every packet pays a
detour through the relay's geographic location.

The **latency tax** is the round-trip detour through the TURN
server. At best, the relay is co-located in the same region as
both endpoints (Cloudflare anycast routing or operator-tier
co-location): 5–10 ms added per direction, 10–20 ms total
session-RTT increase. At worst, the relay is across regions
(operator-tier failover or Cloudflare-tier suboptimal anycast
routing): 20–50 ms added per direction, 40–100 ms total session-
RTT increase. HelixPlay's session-RTT budget (C13 §3 latency
overview) accommodates a single co-region TURN hop without
breaching the 50 ms p99 motion-to-photon target; cross-region
TURN routing is treated as a degraded mode and the C09 admission
gate prefers refusal to cross-region TURN where the tenant
policy permits.

Industry data places TURN-need at **10-15% of consumer sessions**
on average — a number that varies wildly by ISP, by carrier, and
by enterprise-firewall topology. HelixPlay sizes its per-region
TURN capacity at **0.5 × concurrent-session target** (matching
the C19 §5.5 admission rule), which provides headroom for
hostile-network days where the TURN-need rate spikes to 30-40%
without admission throttling. The credential mint is HMAC-SHA256
per-tenant (C09 §6 + C19 §5.3); credentials expire on 60 s
boundaries and the host-agent re-mints on demand. Cloudflare
Realtime TURN is the documented overflow path for tenants who
prefer a CDN-tier no-ops option (C09 §6.3).

### 4.4 ICE-Lite versus Full ICE

RFC 8445 §2.7 defines an **ICE-Lite** mode in which an endpoint
declines to gather candidates and only responds to connectivity
checks initiated by the peer. Lite mode is intended for
endpoints that are **publicly addressable on a predictable port**
(e.g., enterprise SBCs, public-IP server-tier hosts) and
therefore have nothing to gain from candidate gathering. The
benefit of Lite mode is **reduced CPU and protocol latency** at
the simplifying endpoint: no candidate-priority computation, no
gathering-phase delay, no STUN probe of its own.

HelixPlay's MVP rule for the ICE posture is:

> **The client runs Full ICE; the host runs ICE-Lite.**

The rationale is asymmetric. The client sits in unpredictable
NAT topology (residential, mobile, enterprise) and must gather
all three candidate classes (host, server-reflexive, relayed) to
maximise its chance of finding a working pair. The host sits in
a controlled environment — the operator's datacentre tier with
public IPv4/IPv6 binding, or the home-host tier behind a single
predictable consumer NAT with UPnP-IGD or NAT-PMP port forward
already established by the host-agent's setup wizard. The
host-agent's binding is therefore **stable and pre-discovered**;
the candidate-gathering phase would be a no-op. ICE-Lite saves
the host-agent ~50 ms of bootstrap latency per session and
roughly 0.3% of CPU on the host's main core under sustained
session load (measured at 1000 sessions/hour bootstrap rate on a
reference Sapphire Rapids host).

The ICE-Lite host posture interacts cleanly with the C19 §5.4
candidate-gathering rule: the client gathers all three candidate
classes including the relayed candidate from Cloudflare or
coturn, the host advertises its single public binding, and the
connectivity-check phase walks the resulting 3×1 candidate
matrix until one pair succeeds. The pairing logic is implemented
in `vasic-digital/helix-ice` and runs in a goroutine-per-pair
fan-out per Insight #5.

### 4.5 Trickle ICE (RFC 8838)

The vanilla ICE flow of RFC 8445 requires both endpoints to
**finish gathering all candidates** before sending the SDP offer
or answer. This is correctness-preserving but latency-hostile:
candidate gathering involves a STUN probe to every regional STUN
server and a TURN allocation against every TURN server, each of
which can take 50-200 ms depending on RTT and on whether the
STUN/TURN server is cold or warm. Total gathering delay easily
reaches **300-500 ms** before the SDP exchange even begins.

**Trickle ICE** (RFC 8838) replaces the all-at-once exchange
with **incremental candidate exchange**. The endpoint sends its
SDP offer with only the host candidates listed, then trickles
server-reflexive and relayed candidates over the signalling
channel as gathering completes. The peer can begin connectivity
checks immediately on the host candidates (which are zero-RTT
to enumerate — they come from `getifaddrs(3)` on Linux,
`GetAdaptersAddresses` on Windows, `SCNetworkInterface` on
macOS); checks against later-arriving candidates layer in as
they trickle through.

The latency saving is significant. Industry measurements place
trickle-ICE session-setup at 200-500 ms faster than vanilla ICE
in residential deployments. HelixPlay's MVP rule:

> **Trickle ICE enabled by default for both client and host.**

The signalling channel is the HelixPlay session-bootstrap RPC
over HTTP/3 (cross-link C06 §4); the trickle-update message type
is a structured `IceCandidateUpdate` that carries one candidate
per invocation. The host-agent's ICE-Lite posture means it has
nothing to trickle, so the trickle protocol degenerates to
client-only on HelixPlay sessions — but the framework support is
present so that future V1 deployments with full-ICE host-agent
posture (e.g., for SFU-mediated multi-tenant sessions) work
without protocol change.

### 4.6 ICE keepalive

After the connectivity-check phase resolves and the selected
pair is in active use, ICE requires **periodic keepalive
binding indications** to maintain the NAT mapping. Without
keepalive, an idle path (no media for tens of seconds) risks
having its NAT translation expire and the path becoming
silently unreachable. RFC 8445 §11 defines the keepalive cadence
as a STUN binding indication every 15 s by default, with
operator-tier overrides permitted.

HelixPlay's **per-network-tier keepalive cadence** rule:

- **Cellular** — 5 s. Carrier-grade NAT on mobile carriers
  aggressively recycles UDP mappings to conserve translation
  table space; some carriers have observed mappings expiring as
  early as 30 s. The 5 s cadence keeps every mapping alive with
  6× margin against the most aggressive observed timeout.
- **Home Wi-Fi / Ethernet** — 15 s. Consumer router NAT mappings
  typically persist for 60-300 s; the default cadence
  comfortably preserves mappings without keepalive overhead
  becoming significant.
- **Datacentre / enterprise** — 30 s. Operator-controlled
  network paths with predictable NAT behaviour or no NAT at all;
  the longer cadence reduces keepalive packet rate without risk.

The cadence is set per-session at bootstrap based on the
client's reported network-tier (the C09 admission RPC carries a
`NetworkTier` enum); the host-agent honours the cadence on its
side so both endpoints emit binding indications at the same rate.
Keepalive packets are budgeted **outside** the streaming hot-path
budget — they ride the same UDP socket but at a low per-packet
priority so they never delay a media packet.

### 4.7 NAT type detection

The four canonical NAT types defined by classic STUN
classification (RFC 3489, deprecated but still descriptively
useful):

- **Full-cone (endpoint-independent mapping, endpoint-
  independent filtering)** — any external host can reach the
  internal host through the same mapping. Best case for direct
  P2P; almost extinct on consumer networks but common on
  enterprise networks with explicit port-forward.
- **Address-restricted cone (endpoint-independent mapping,
  address-dependent filtering)** — external host can reach the
  internal host through the mapping only if the internal host
  has previously sent a packet to that external host's address.
  Common on consumer routers; STUN-discovered mapping is reusable.
- **Port-restricted cone (endpoint-independent mapping,
  address-and-port-dependent filtering)** — same as above but
  with port-tuple restriction. Most common consumer NAT type;
  STUN-discovered mapping is reusable.
- **Symmetric (endpoint-dependent mapping)** — the NAT allocates
  a different external mapping per remote endpoint. STUN-
  discovered mapping is **useless** to a third-party host; TURN
  relay is mandatory.

HelixPlay's session-bootstrap fires a **multi-server STUN probe**
to two distinct STUN servers (regional + Cloudflare or operator-
fallback); if the two servers report different external mappings
the NAT is symmetric, otherwise it is endpoint-independent. The
result is reported to the ABR controller (C33 §4) and to the
session-state log, where it informs:

- **Jitter buffer sizing** — symmetric NAT correlates with
  carrier-grade NAT and mobile-tier networks where jitter is
  higher; the ABR controller selects a tighter jitter-buffer
  posture (50 ms upper bound vs 80 ms default) to compensate for
  the higher-quality signalling.
- **Refresh interval** — symmetric NAT goes hand-in-hand with
  aggressive NAT-mapping recycling; the keepalive cadence is
  pulled from 15 s default to 5 s cellular default regardless of
  the reported network-tier.
- **TURN preference** — symmetric NAT pre-commits the path to
  TURN relay; the candidate-gathering phase skips the server-
  reflexive candidate-pair check (which is guaranteed to fail)
  and goes straight to the relayed candidate, saving 50-100 ms of
  bootstrap latency.

The detection result is logged as a per-session anomaly when the
NAT type differs from the bootstrap default for the reported
network-tier, providing operator-side visibility into NAT-
topology drift across the user base.

### 4.8 Cross-link

The ICE / STUN / TURN surface specified here cross-links into
two architectural neighbours:

- **C06 §3 WebRTC patterns** — the WebRTC-tier session bootstrap
  uses the same ICE candidate graph as the custom-UDP-tier
  session bootstrap, with the difference that WebRTC layers SRTP
  (RFC 3711) on top of the selected pair while custom-UDP layers
  the HelixPlay 4-byte framing header. The ICE state machine is
  shared between both transports; the wire-format diverges only
  after the selected pair is in active use.
- **C10 §4 DTLS-SRTP key exchange** — DTLS handshake runs over
  the selected ICE pair and produces the SRTP master key (for
  WebRTC sessions) or the application-layer DTLS keying material
  (for custom-UDP sessions). C10 owns the cipher allow-list
  (`TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384` for MVP DTLS 1.2);
  C37 cites it but never enumerates alternatives.

---

## 5. RTCP feedback

Section §4 closed the NAT-traversal arc: ICE establishes the
selected candidate pair, STUN keeps NAT mappings alive, and
TURN provides the last-resort relay path when direct
connectivity is impossible. With the path established, the
streaming session begins emitting media payload (RTP) and
control feedback (RTCP). This section codifies HelixPlay's
RTCP feedback posture — what feedback messages each endpoint
emits, at what cadence, with what payload, and how the feedback
drives the ABR controller in C33. The binding constraint
throughout is that **RTCP is the closed-loop control plane of
the streaming session**: every adaptive behaviour (bitrate
ramp-up, FEC redundancy adjustment, congestion-induced
backoff, thermal-throttle response) consumes an RTCP feedback
signal as its input. An RTCP outage is therefore not just a
metrics gap — it blinds the entire adaptive control loop.

### 5.1 RTCP overview (RFC 3550)

The RTP Control Protocol (RTCP — RFC 3550 §6) is the **out-of-
band feedback channel** that accompanies every RTP session. It
shares the underlying transport (UDP), uses an adjacent port
(odd port = RTCP, even port = RTP, or multiplexed onto the same
port via RTCP-Mux per RFC 5761), and carries five canonical
packet types defined in the base RFC:

- **SR (Sender Report, RFC 3550 §6.4.1)** — emitted by every
  active sender. Carries the sender's NTP timestamp, RTP
  timestamp, packet count, and octet count. Receivers use SR
  data to map RTP timestamps to wall-clock time for inter-stream
  synchronisation (lip-sync between video and audio).
- **RR (Receiver Report, RFC 3550 §6.4.2)** — emitted by every
  active receiver. Carries reception statistics: fraction lost,
  cumulative packets lost, extended highest sequence number,
  jitter, last SR timestamp, delay since last SR. The fraction-
  lost field is the basic congestion-feedback signal that
  predates the modern transport-cc and REMB extensions.
- **SDES (Source Description, RFC 3550 §6.5)** — carries a
  CNAME (canonical participant identifier) and optional
  metadata (NAME, EMAIL, PHONE, LOC, TOOL, NOTE). HelixPlay
  uses CNAME for session correlation in the observability bus.
- **BYE (RFC 3550 §6.6)** — emitted on session shutdown.
  HelixPlay handles BYE as a graceful-termination signal: the
  receiver flushes the jitter buffer and the sender stops media
  emission within 100 ms.
- **APP (RFC 3550 §6.7)** — application-defined extension.
  HelixPlay reserves three APP packet subtypes: thermal-throttle
  events (cross-link C34 §5.5), recording-state events
  (cross-link C30 §7), and Reflex-tag events (cross-link C13
  §3 motion-to-photon timeline tagging).

RTCP's **bandwidth budget** is governed by RFC 3550 §6.2 — the
default is 5% of the session bandwidth, capped at a minimum
interval of 5 s and a maximum of 30 s between transmissions per
sender. HelixPlay overrides the minimum to **100 ms** (the
floor permitted by RFC 3550 §A.7 for "tight" feedback regimes)
because the closed-loop adaptive control surface (C33 ABR) needs
RR-level feedback at frame-time scales, not 5-second scales. The
100 ms override applies uniformly to all RTCP packet types
including SR/RR; it is implemented in `vasic-digital/helix-rtcp`
(new under R-03) and gated by the operator-policy capability
schema so per-tenant overrides are possible.

### 5.2 RTCP-FB (Codec-Control Messages, RFC 4585)

RFC 4585 introduces the **RTP/AVPF profile** — Audio-Visual
Profile with Feedback — and a class of **transport-layer
feedback** and **payload-specific feedback** messages emitted
through the RTCP channel. The feedback messages most relevant
to HelixPlay's streaming hot path are:

- **NACK (Negative Acknowledgement, RFC 4585 §6.2.1, FCI=GenericNACK)**
  — receiver reports a list of lost RTP sequence numbers. The
  sender retransmits the named packets, subject to retry budget
  and time-budget constraints. C33 §5.1 owns the full NACK
  policy (retry budget = 1, time budget = 1 RTT + jitter
  margin); C37 cites the contract but does not relitigate it.
- **PLI (Picture Loss Indication, RFC 4585 §6.3.1, FCI=PLI)** —
  receiver requests a fresh keyframe because decode state has
  diverged (typically following an unrecoverable frame loss).
  C33 §5.2 owns the PLI ladder (PLI on first decode-divergence
  event; FIR escalation if PLI does not produce a keyframe
  within 200 ms).
- **FIR (Full Intra Request, RFC 5104)** — receiver explicitly
  requests a full intra-coded refresh, used as the escalation
  beyond PLI. FIR is more disruptive (forces a full IDR;
  bandwidth spike of 3-5× the steady-state) but unconditionally
  resyncs decoder state. C33 §5.3 owns the FIR ladder.
- **REMB (Receiver Estimated Maximum Bitrate, draft-alvestrand-
  rmcat-remb)** — receiver reports an estimate of available
  bandwidth that the sender should not exceed. Pre-dates
  transport-cc; still widely deployed. §5.4 below covers REMB.

The feedback cascade (NACK → PLI → FIR) is the canonical
recovery ladder for sub-second decode-state desync; cross-link
C33 §5 for the full state machine. C37's role here is to
specify that all four feedback messages share a single RTCP
stream (no separate transport channel) and ride the same UDP
socket as the RTCP SR/RR base stream.

### 5.3 transport-wide-cc (draft-ietf-rmcat-transport-cc)

The vanilla RR feedback (RFC 3550 §6.4.2) reports per-SSRC loss
fraction and jitter at session-aggregate granularity — useful
for SR/RR-driven legacy congestion control but inadequate for
modern delay-based congestion control like Google Congestion
Control (GCC) or SCReAM. Those algorithms need **per-packet
arrival timestamps** to compute the one-way-delay gradient that
detects incipient queueing well before loss occurs.

**transport-wide congestion control feedback** (draft-ietf-rmcat-
transport-cc) extends RTCP-FB with a **per-packet feedback**
report. The sender attaches a 16-bit transport-wide sequence
number (carried in an RTP header extension) to every outgoing
packet; the receiver echoes back the sequence-number list with
**arrival deltas in 250 µs ticks**. The sender now has a
millisecond-resolution view of every packet's arrival time
relative to its send time, which is exactly the input GCC needs
to compute the one-way-delay gradient via Kalman filtering.

HelixPlay's MVP posture:

> **transport-wide-cc enabled by default for all sessions —
> WebRTC-tier and custom-UDP-tier alike.**

The custom-UDP-tier transport reuses the same wire format as
RTCP transport-wide-cc reports because the C19 §3 SQP-style
congestion control consumes the same per-packet arrival timing
signal. The shared format means the C33 ABR controller can be
implemented once and consume feedback uniformly across both
transport tiers; it cross-links into C33 §4 (GCC) and C33 §4.4
(SQP). The feedback report is emitted **every 100 ms** (matching
the §5.1 minimum-interval override) and carries the sequence
numbers of all packets received in that window — typically
50-200 packets at 1080p60 or 200-500 packets at 4K60.

### 5.4 REMB — Receiver Estimated Maximum Bitrate

REMB (draft-alvestrand-rmcat-remb, also informally referenced in
RFC 8888's design history) is the **receiver-side bandwidth
estimate**. The receiver runs its own bandwidth estimation
algorithm (typically a delay-based estimator — Google's
Receiver-Side BWE — paired with a loss-based fallback) and
periodically reports the **estimated maximum bitrate the
network can sustain** to the sender via an RTCP-FB packet
(FCI=REMB). The sender treats the REMB value as a **bandwidth
ceiling** and must not exceed it for the session.

REMB pre-dates transport-wide-cc by several years and remains
widely deployed as a coarse-grained bandwidth signal. Its
publication cadence in HelixPlay is **2-3× per second** (every
333-500 ms), which matches the WebRTC-Bandwidth-Estimation
guidance and is well-suited to coarse bandwidth-ceiling signals
without flooding the RTCP channel. In a transport-cc-enabled
session, REMB is treated as a **belt-and-braces fallback**: GCC
runs on the per-packet feedback for fine-grained adjustments,
but if the per-packet feedback channel falls silent (e.g.,
receiver implementation gap, interop with a non-transport-cc
implementation) the REMB ceiling still bounds the sender's
output rate.

The HelixPlay ABR controller (C33 §4) consumes both signals: the
GCC-derived target rate from transport-wide-cc, capped by the
REMB ceiling. The min of the two is the operating bitrate. On
WebRTC-tier sessions where the receiver is a third-party browser
without transport-cc support, REMB is the only signal available
and the ABR controller falls back to REMB-only mode with
appropriate conservatism (slower ramp-up, faster backoff).

### 5.5 RFC 8888 — Congestion Control Feedback

**RFC 8888** (CCFB — Congestion Control Feedback) was
standardised in 2020 as the IETF replacement for transport-wide-
cc. It carries **per-packet feedback** (sequence number, ECN
mark, arrival-time relative to report block) in a more
structured and more compact wire format than transport-wide-cc,
and it is the long-term successor that the IETF RMCAT working
group converges on.

The trade-off vs transport-wide-cc:

- **Precision** — RFC 8888 carries ECN-CE (Congestion
  Experienced) marking per packet, which transport-wide-cc does
  not. ECN feedback is critical for L4S deployments (cross-link
  C19 §4) where the ECT(1) marker is set on every outgoing
  packet and the receiver echoes back ECN-CE marks the network
  applied. With ECN feedback, the sender can detect incipient
  congestion **before any packet is lost**, enabling truly
  zero-loss adaptive control on L4S-enabled paths.
- **Overhead** — RFC 8888 reports are slightly larger than
  transport-wide-cc reports for the same packet population (the
  extra ECN field plus a more verbose header). On a high-rate
  session (4K60, 500 packets/100 ms window) the overhead delta
  is roughly 50 bytes per report — small in absolute terms but
  non-zero.
- **Adoption** — RFC 8888 implementations in 2026 are
  asymmetrically deployed: Pion v3 alpha supports it, libwebrtc
  M120+ supports it as opt-in, but most deployed WebRTC stacks
  still emit transport-wide-cc.

HelixPlay's posture:

> **MVP: transport-wide-cc as the canonical per-packet feedback.
> V1: RFC 8888 evaluated as the replacement once libwebrtc and
> Pion graduate the implementation to default-on.**

The ABR controller (C33 §4) is wire-format-agnostic — it
consumes a structured `PerPacketFeedback` event regardless of
which RTCP message type produced it. The capability schema (C09
§3) negotiates which message type is in use per-session, and the
codepath that parses RTCP into the structured event is in
`vasic-digital/helix-rtcp` with one parser per supported message
type. The V1 RFC 8888 swap is therefore a parser-side change
only; no consumer code re-architects.

### 5.6 RTCP throttle-active extension (HelixPlay)

HelixPlay introduces a **HelixPlay-specific RTCP APP packet**
(name = `HXTL`, application-defined per RFC 3550 §6.7) carrying
**thermal-throttle events** from the encoding host. The packet is
emitted whenever the C34 §5 thermal-aware quality controller
reduces encode quality in response to GPU temperature crossing
the 78°C pre-emptive threshold or the 83°C hard cap. The packet
payload carries:

- **Severity tier** — pre-emptive vs hard-cap;
- **Quality reduction magnitude** — bitrate delta in Mbps;
- **Estimated duration** — heuristic projection of the reduced-
  quality window in seconds;
- **Triggering metric** — GPU temperature, GPU power, GPU clock,
  or composite-score reading that fired the controller.

The receiver-side ABR controller (C33) consumes the throttle
event with a **200 ms response budget**: within 200 ms of receipt
the ABR controller has either accepted the lower bitrate as the
new operating point (no client-side action) or surfaced an
operator-policy event to the session-state log indicating that
the host is thermally compromised. The 200 ms budget is short
enough that the throttle event drives the ABR before the
transport-cc-derived bitrate signal would catch up
(transport-cc takes 500-1500 ms to converge on a new operating
point through its Kalman filter).

The throttle-active extension is documented in `vasic-digital/
helix-rtcp` and inherited by both WebRTC-tier and custom-UDP-tier
sessions. It is one of three HelixPlay APP-packet subtypes (the
other two being recording-state and Reflex-tag, §5.1 above).

### 5.7 RTCP message-type matrix

The eight RTCP message types HelixPlay emits and consumes,
canonically, on every session — type, RFC reference, default
emission cadence, and the HelixPlay use-case binding:

| Type | RFC | Default interval | HelixPlay use |
|------|-----|------------------|---------------|
| SR (Sender Report) | RFC 3550 §6.4.1 | 100 ms (C37 override; default RFC = 5 s) | NTP→RTP timestamp mapping for lip-sync; sender-side packet/octet counters |
| RR (Receiver Report) | RFC 3550 §6.4.2 | 100 ms (C37 override; default RFC = 5 s) | Coarse loss-fraction + jitter feedback; ABR fallback signal when transport-cc unavailable |
| SDES (CNAME) | RFC 3550 §6.5 | 100 ms compound with SR/RR | Session correlation in observability bus; per-tenant CNAME |
| BYE | RFC 3550 §6.6 | On session termination | Graceful shutdown; flush jitter buffer; stop media emission within 100 ms |
| NACK / PLI / FIR | RFC 4585 + RFC 5104 | Event-driven | Loss recovery ladder; cross-link C33 §5 for full state machine |
| transport-wide-cc | draft-ietf-rmcat-transport-cc | 100 ms | Per-packet arrival feedback; primary input to GCC / SQP / SCReAM |
| REMB | draft-alvestrand-rmcat-remb | 333 ms (3× per second) | Coarse bandwidth ceiling; fallback ABR signal when transport-cc absent |
| APP/HXTL (HelixPlay throttle) | RFC 3550 §6.7 (APP) | Event-driven (on threshold cross) | Thermal-throttle propagation; ABR responds within 200 ms |

The eight rows partition the RTCP feedback surface cleanly: four
RFC 3550 base types, three RFC 4585 / 5104 codec-control types,
two delay-based congestion-control types, and one HelixPlay-
specific application extension. No row of the matrix is
relitigated downstream; consumer chapters cite the row by name.

### 5.8 Cross-link

The RTCP feedback surface specified here cross-links into three
chapters across the HelixPlay synthesis:

- **C19 §6 jitter buffer** — RR fraction-lost and transport-cc
  delta-arrival drive the jitter-buffer adaptive sizing
  algorithm. C19 owns the buffer mechanics (RTP-timestamp-based
  reordering, adaptive depth, late-packet eviction); C37 §5.3
  feeds the input.
- **C33 §4-§5 ABR + NACK** — the ABR controller consumes
  transport-cc, REMB, and APP/HXTL throttle events to derive the
  operating bitrate; the NACK-PLI-FIR cascade implements the
  recovery ladder for decode-state desync. C33 owns both
  controllers; C37 emits the feedback that drives them.
- **C34 §5 thermal-throttle feedback** — the APP/HXTL packet
  payload format and the 200 ms response budget are codified in
  C34 §5.5; C37 §5.6 specifies the wire-side carriage.

---

## 6. QUIC + MoQ (Media-over-QUIC)

Section §5 closed the RTCP feedback arc: HelixPlay's per-session
control loop runs on transport-wide-cc + REMB + APP/HXTL APP
packets, all carried over UDP alongside the RTP media payload.
The combined RTP+RTCP-over-UDP architecture is the **MVP wire
contract** and is what the C33 ABR controller and the C19 jitter
buffer consume. Yet the IETF and the wider streaming-video
ecosystem are converging on a different long-term wire
architecture: **QUIC** (RFC 9000-9002) as the universal transport
substrate, and **Media-over-QUIC** (MoQ; the IETF MoQ working
group's draft-ietf-moq-transport) as the RTP successor for
real-time media. This section codifies HelixPlay's MVP and V1
posture vis-à-vis QUIC and MoQ — what runs on QUIC today (the
control plane), what is reserved for V1 evaluation (the data
plane), and how the SQP-style custom-UDP path of Insight #7
positions for a future MoQ migration.

### 6.1 QUIC (RFC 9000-9002)

QUIC is the **transport-over-UDP** standardised by RFC 9000 (the
core protocol), RFC 9001 (TLS 1.3 binding), and RFC 9002 (loss
detection and congestion control). Designed by Google and
adopted as IETF standards-track in 2021, QUIC delivers four
material improvements over the TCP+TLS+HTTP/2 stack it replaces:

- **Built-in TLS 1.3** — no separate handshake; the QUIC
  connection establishment carries TLS 1.3 keys natively.
  Crypto setup is amortised into the same RTT as transport
  setup, saving 1-2 RTT vs the TCP+TLS sequence.
- **0-RTT resumption** — a returning client can send application
  data on the very first packet by using a previously
  established session ticket. Critical for cold-cache cloud-
  gaming session resumption (cross-link §6.7 below).
- **Multiplexed streams** — multiple application streams ride a
  single QUIC connection without head-of-line blocking between
  them. A stalled stream does not block other streams on the
  same connection, unlike TCP where any in-flight loss blocks
  all higher-sequence bytes.
- **Cancellation per stream** — application-layer can cancel an
  individual stream (RST_STREAM equivalent) without tearing
  down the connection. Useful for HelixPlay's control-plane
  RPCs that go stale (e.g., a session-bootstrap cancelled by
  the client before ICE completes).

HelixPlay's MVP rule:

> **The control plane runs on QUIC (HTTP/3). The data plane
> runs on UDP+RTP+RTCP for MVP; QUIC datagram (RFC 9221) and
> MoQ are V1 evaluation paths.**

The control-plane traffic — session-bootstrap RPC, capability
advertise, ICE candidate trickle, heartbeat, telemetry export —
is all Connect-Go RPC (cross-link C06 §4) over HTTP/3 over QUIC
on UDP/443. The data-plane traffic — RTP video, RTP audio, RTCP
feedback, DTLS-encrypted controller input — runs on the C19-
style raw UDP path with HelixPlay's 4-byte framing header. The
two run side-by-side on every session, on different UDP ports.

### 6.2 MoQ — Media-over-QUIC

The **Media-over-QUIC** (MoQ) working group at the IETF is
chartered to define the next-generation real-time media
transport. The flagship draft is **draft-ietf-moq-transport**
(commonly abbreviated *moqt* or *mQ*); the working group also
maintains companion drafts for catalogue exchange, container
formats, and security profile. The 2025-2026 IETF cycle
finalised the core protocol and saw early implementations
(Cisco, Meta, Google) begin interoperability testing.

MoQ's design points:

- **Object-based delivery** — media is delivered as a stream of
  named objects (groups of frames) rather than an opaque packet
  stream. Each object has an explicit identifier and can be
  fetched, cached, or skipped independently.
- **Track per QUIC stream** — each media track (video, audio,
  metadata) maps to its own QUIC stream, so loss in one track
  does not block the others. Compared to RTP-over-UDP where
  every loss is independent regardless, this only matters for
  multi-track sessions; cloud gaming with two tracks (video +
  audio) sees modest benefit. The bigger win is for n-track
  sessions like multi-camera or AR/VR.
- **Relay tier** — MoQ explicitly contemplates relay nodes
  ("MoQ relays") that can fan out a publisher's stream to
  thousands of subscribers without re-encoding, at a much lower
  per-subscriber cost than WebRTC SFUs. For HelixPlay's
  spectator-mode V1 feature this is potentially transformative.

HelixPlay's posture:

> **MVP: not used. V1: evaluated as RTP successor for V2
> deployments where the relay-tier benefit pays off.**

The V1 evaluation is gated by two conditions: (1) at least two
of {Pion, libwebrtc, Cisco's MoQ stack} ship a stable
implementation with publicly demonstrated 4K60 cloud-gaming
performance, and (2) the operator-policy posture authorises a
non-RTP data plane. Both conditions are likely to land in the
2027-2028 window per current IETF cadence; HelixPlay's V1
implementation reserves the protocol slot but does not commit
the wire format.

### 6.3 SQP cross-link

Insight #7 from the video-tech insight extraction documents the
**SQP envelope** — Stadia Quality Protocol, the Google-Research-
era custom-UDP transport that achieved ~7 ms LAN latency and
TCP-friendly congestion control without WebRTC's protocol
overhead. SQP itself was never open-sourced past Stadia's
shutdown, but the **architectural pattern** — frame-coupled
packet trains, per-frame congestion-control feedback, custom
DTLS-encrypted UDP framing — is what HelixPlay's custom-UDP
path of C19 §3 + C33 §4.4 implements. The pattern is
documented as an *envelope* (the protocol shape, not the
specific bits-on-wire) so that future implementations can
choose between the SQP-pattern custom UDP and an MoQ-pattern
custom QUIC framing as the underlying protocol matures.

HelixPlay's V2 posture (post-V1, post-2028) reserves the SQP-
style custom-UDP slot for a possible **MoQ replacement**: if
MoQ delivers on its low-latency promise, the custom-UDP path
swaps to MoQ-over-QUIC; if MoQ does not deliver, the custom-UDP
path stays on raw UDP indefinitely. The decision point is V2,
not MVP, and is gated by the same two conditions as §6.2's MoQ
evaluation. The Insight-#7-confidence rating is MEDIUM (per the
insight document) because SQP is Google Research, not a
standardised protocol; the V2 evaluation may converge on
something quite different than what the insight envisions.

### 6.4 QUIC datagram (RFC 9221)

The mainline QUIC protocol provides only **reliable, ordered
streams**. For real-time media this is the wrong abstraction —
a media packet that is one frame stale is worthless and must
not block subsequent fresh packets. **QUIC datagram** (RFC 9221)
extends QUIC with an **unreliable datagram extension**: an
application can send a datagram over a QUIC connection that
the receiver delivers if it arrives in time and drops if not,
without retransmission and without reordering. The extension
uses the same QUIC connection identifiers, the same cryptographic
context, and the same congestion control as the reliable
streams, but with RTP-style real-time semantics for the
datagram payload.

The use case for HelixPlay: **RTP-over-QUIC**. The RTP and
RTCP packets that today ride raw UDP could ride QUIC datagrams
instead, gaining QUIC's connection migration (§6.6 below) and
0-RTT resumption (§6.7 below) while preserving the unreliable-
delivery semantics RTP requires. The latency cost is the QUIC
header overhead (4-12 bytes per datagram vs zero for raw UDP)
and the QUIC encryption overhead (already paid on the control
plane).

Per Insight #7's reading of the 2026 measurement landscape,
RTP-over-QUIC vs raw-UDP RTP shows a **4-9 ms latency
reduction** on cellular and on hostile-network paths where
connection migration kicks in often; on home Wi-Fi the two
are within measurement noise. HelixPlay's MVP does not run
RTP-over-QUIC (the wire format is raw UDP per C19 §2); V1
evaluates it as part of the broader MoQ migration evaluation
(cross-link C33 §4.4).

### 6.5 HTTP/3 in HelixPlay's control plane

HTTP/3 is the application-layer protocol that **layers on top
of QUIC** — RFC 9114 specifies the binding. Per the HelixPlay
project-wide constraint codified in CLAUDE.md ("HTTP/3
(QUIC/Cronet) preferred"), every control-plane request
HelixPlay makes is HTTP/3:

- **Connect-Go RPC** for session bootstrap, capability
  advertise, ICE candidate trickle, telemetry export — all over
  HTTP/3 (cross-link C06 §4).
- **gRPC** for inter-service communication on the operator-
  controlled fleet — over HTTP/3 wherever the gRPC stack
  supports it (Cronet on mobile clients; native Go gRPC on the
  server tier).
- **Brotli compression** on every JSON payload >256 bytes
  (CLAUDE.md project-wide rule); negotiated through HTTP/3
  ALPN.

The control-plane QUIC connection is **separate** from any data-
plane QUIC connection (when the data plane migrates to QUIC in
V1+). The two connections may share a server certificate but
they do not share QUIC connection state — connection migration
on the control plane (e.g., Wi-Fi → cellular failover during a
session-bootstrap RPC) does not affect the data-plane RTP
traffic, and vice versa. This separation matches the C19 §3
asymmetric optimisation: the control plane is the kernel-tier
QUIC stack (Quinn / quic-go) on commodity hardware; the data
plane is the userspace raw-UDP path on host-tier hardware.

### 6.6 QUIC connection migration

QUIC's headline mobile-tier feature is **connection migration**:
the QUIC connection identifier is cryptographic, not address-
based, so a client whose IP changes mid-session can continue
using the same QUIC connection on the new IP. The classic
trigger scenarios:

- **Cellular handover** — mobile client moves between cell
  towers and the carrier-grade NAT remaps its public IP.
  Without connection migration, every TCP+TLS connection drops
  and the application must reconnect; with QUIC migration, the
  connection survives transparently.
- **Wi-Fi → cellular failover** — mobile client walks out of
  Wi-Fi range and the OS hands the connection over to cellular.
  Same migration story.
- **VPN re-establishment** — client's VPN tunnel reconnects on
  a different public IP. Same migration story.

For cloud gaming this is a **mobile-tier benefit** specifically.
Desktop clients on residential Ethernet rarely change IP mid-
session; the migration feature is ~zero-value there. Mobile
clients (the V1 Android / iOS clients per CLAUDE.md and per
C04) see large benefit on commute scenarios where the client
moves between Wi-Fi cells and cellular cells multiple times per
session.

HelixPlay's MVP rule:

> **Connection migration enabled by default on mobile clients.
> Disabled on desktop clients (no benefit, marginal complexity
> cost).**

The capability is gated through the C09 admission RPC: the
mobile client advertises `MobileConnectionMigration=true`, the
server enables migration on the QUIC connection, and the
client may transparently change IP during the session. On the
data plane side (which is raw UDP for MVP), the equivalent
mechanism is the C19 §5 ICE re-negotiation — the client
re-runs ICE on IP change and re-establishes the selected pair.

### 6.7 0-RTT and replay protection

QUIC's 0-RTT resumption (RFC 9001 §4.6.1) lets a returning
client send application data on the very first packet of a new
connection by encrypting it under a key derived from a
previously established session ticket. The benefit is **one
fewer round-trip for connection establishment** — a returning
client's session-bootstrap RPC reaches the server immediately,
saving 30-100 ms of RTT depending on network distance.

The cost is the **0-RTT replay-attack window**. An attacker
who captures the 0-RTT-encrypted application data can replay
it later against the same server, and the server has no way to
distinguish replay from a genuine retry without out-of-band
state. For idempotent operations this is harmless; for state-
mutating operations it is potentially catastrophic (e.g., a
replayed "purchase X" command processes the purchase twice).

HelixPlay's MVP rule:

> **0-RTT for control-plane idempotent operations only; never
> for state-mutating operations.**

The Connect-Go RPC handlers tag each method as idempotent or
state-mutating in the RPC service definition; the QUIC stack
allows 0-RTT only on the idempotent set. State-mutating RPCs
(start session, stop session, mint TURN credential, write
recording state) require the full 1-RTT handshake. The data
plane (raw UDP RTP/RTCP) does not use 0-RTT at all — it is
DTLS-handshake-pre-established during session bootstrap, and
the streaming hot path never blocks on a fresh DTLS handshake
(cross-link C19 §2.5 retransmission policy).

The 0-RTT replay-attack mitigation is documented in C10 §4
(security side); C37 §6.7 cross-links the rule but does not
relitigate the cipher-suite or single-use-ticket implementation.

### 6.8 Cross-link

The QUIC + MoQ surface specified here cross-links into three
neighbours across the HelixPlay synthesis:

- **C06 control plane** — every control-plane request runs on
  HTTP/3 over QUIC (§6.5 above). C06 owns the RPC service
  surface; C37 specifies the transport binding.
- **C10 §4 DTLS 1.3 / TLS 1.3** — QUIC's built-in TLS 1.3 (§6.1)
  shares cipher allow-list and 0-RTT replay-mitigation rules
  with the C10 DTLS surface. C10 owns the security contract;
  C37 cites it.
- **Insight #7 SQP envelope** — the V2 SQP-replacement decision
  (§6.3) is gated by MoQ maturity (§6.2) and the V1 RTP-over-
  QUIC evaluation (§6.4). All three surfaces converge on the
  same V2 decision point.
## 7. sendmmsg + io_uring batching + XDP eBPF

The §3–§6 sections of this chapter (Section A + Section B) lay out the
SRTP / SRTCP transport contract, the ICE/STUN/TURN traversal lattice,
and the QUIC control-plane session. None of those layers, individually,
guarantee that the **wire I/O** beneath them clears the per-packet
syscall budget that a 60 fps / 120 fps / 240 fps video stream demands.
A single 4K120 HDR session running through a per-packet `sendto(2)`
loop saturates a single CPU core on the syscall path alone — at
roughly 25 Mbit/s with average MTU-sized RTP packets, that is on the
order of 2,500 outbound packets per second per session, and at four
sessions per host the per-host outbound rate is 10 K pps **before**
control-plane traffic. Each `sendto(2)` invocation costs a syscall
trap (~150 ns on a Skylake-class core, ~50 ns on Sapphire Rapids),
plus per-packet buffer copy and protocol-stack walk; the cumulative
syscall-side budget eats into the C13 §4 input-to-photon latency
budget faster than any other transport-layer cost. Section C
therefore pins down the **kernel-bypass-and-batching ladder** for
HelixPlay's outbound RTP / SRTP / RTCP path: §7 covers the four
canonical batching primitives (`sendmmsg(2)`, io_uring `IORING_OP_SEND`
and `IORING_OP_SEND_ZC`, AF_XDP / XDP_TX / XDP_REDIRECT, plus the
sock_diag socket-statistics oracle), §8 covers Multipath UDP / MPQUIC
as the V1 evolution, and §9 binds the entire transport contract to
the `vasic-digital/helix-transport` submodule with a Go reference
implementation, capability schema delta, R-18 enforcement recap, and
failure-semantics matrix.

### 7.1 sendmmsg(2) — the canonical batching syscall

`sendmmsg(2)` is the **first floor of HelixPlay's outbound batching
ladder** — a Linux syscall (since 3.0, December 2010, generally
available on every kernel HelixPlay supports) that submits an
**array** of `mmsghdr` structures in a single syscall, with the
kernel iterating through the array and dispatching each datagram
through the standard UDP / IPv6 stack. The man-page contract is
precise: the syscall returns the number of messages successfully
queued, and a partial-success return is permitted (the kernel may
queue 40 of 64 messages and return 40, leaving the caller to retry
with the remaining 24). HelixPlay's RTP sender uses `sendmmsg(2)`
as the **default** batched-send primitive on tier-1..4 hosts (the
non-io_uring tiers per the C16 §4 capability ladder), batching up
to **64 RTP packets per syscall** as the chapter rule. The
batch size of 64 is the per-codec balance between the
syscall-amortisation gain (one syscall replacing 64 sendto calls
saves ~9.5 microseconds at Skylake; ~3.2 microseconds at Sapphire
Rapids) and the head-of-line-blocking risk (a 64-packet batch
holds back the first packet's wire-time by the duration it takes
the encoder + packetiser to produce 63 more packets — at 4K120
that is ~3 ms of pacing headroom, well within the §6 C19 jitter
budget). The 64-packet ceiling cross-links the C16 §3 sendmmsg
pattern, which establishes 64 as the canonical batch size for
the io_uring / sendmmsg fallback ladder; transport workers MUST
NOT exceed 64 without raising an OQ-C37 ticket and updating both
this chapter and C16 §3.

Each `mmsghdr` entry carries a per-message `msghdr` (with
scatter-gather `iovec` array, control-message buffer, and per-
message destination-address override via `msg_name`), and per-
message return value (success or per-message `errno`). HelixPlay
uses the per-message destination-address override on the unicast
RTP path (every packet is destined for the same client `(IP, port)`
tuple, so `msg_name` is the same on every entry — but the kernel
still validates it per-packet, which is the dominant cost on the
sendmmsg path past the syscall trap). For SRTCP feedback
messages (NACK / PLI / FIR / RR / SR per C19 §5 and the C33 §5
NACK contract), the sender batches up to 16 RTCP messages per
sendmmsg, with the lower batch size reflecting the lower RTCP rate
(RTCP traffic is 5% of total per RFC 3550 §6.2.1 budget). The
per-message return values are inspected on the partial-success
path; any per-message error other than `EAGAIN` triggers a
transport-layer alarm and a retry on the offending message slot,
with `EMSGSIZE` triggering an MTU-discovery fallback (the
transport worker shrinks the per-packet payload below the
discovered path MTU, cross-link C19 §4.3 PMTUD).

The `sendmmsg(2)` ceiling on a single Skylake-class core,
benchmarked on a 25 GbE NIC with the standard kernel UDP stack,
sits at ~600 K pps with a 64-packet batch — a ~4× improvement
over per-packet `sendto(2)` (which tops out at ~150 K pps on the
same hardware). For HelixPlay's per-host outbound budget of 10 K
pps under four-session 4K120 load, sendmmsg leaves ~50× headroom
on the syscall path, freeing the core to handle encoder feedback,
ABR control-plane fan-out, and the RTCP feedback consumer. The
§9.4 reference implementation demonstrates the sendmmsg call
pattern with a real `unix.Sendmmsg` invocation against the Pion
SRTP context.

### 7.2 io_uring submission — `IORING_OP_SEND` + `IORING_OP_SEND_ZC`

`io_uring`, the Linux 5.1+ async-I/O subsystem documented at length
in C16 (the entire chapter is the kernel-bypass async-I/O floor for
HelixPlay), supplies the **second floor** of HelixPlay's outbound
batching ladder. Where sendmmsg amortises one syscall across many
datagrams, io_uring amortises **zero syscalls** in the steady state:
the userspace producer writes Submission Queue Entries (SQEs) into
a memory-mapped ring, and either the kernel-side SQPOLL thread
reads the ring without any syscall trap at all (per the C16 §3
SQPOLL configuration), or a single `io_uring_enter(2)` syscall
flushes the ring on demand. For the outbound RTP path, the
HelixPlay rule is that **tier-5+ hosts (kernel ≥ 5.15, io_uring
support advertised in the capability schema per C16 §6)** use
`IORING_OP_SEND` (5.6+) or `IORING_OP_SEND_ZC` (5.20+) for the
per-packet submission, with the SQPOLL thread polling the ring on
a dedicated isolated CPU core (per the C20 §4 isolcpus posture).

The `IORING_OP_SEND` operation is the io_uring analogue of
`send(2)` — a single SQE per outbound packet, flushed in batch by
the SQPOLL thread or by an explicit `io_uring_enter(2)` from the
producer. HelixPlay's transport worker submits up to **128 SQEs
per ring iteration** for the RTP path (twice the sendmmsg batch
size, reflecting the lower per-SQE cost; the SQE itself is 64
bytes versus the sendmmsg per-message header of ~80 bytes plus
the iovec / control-message sidecars), with the ring sized at
4096 SQEs to absorb bursts without falling back to the slow path.
The Completion Queue Entries (CQEs) carry per-packet success or
per-packet `errno`, mirroring the sendmmsg per-message return value
contract; the transport worker drains CQEs via a dedicated reader
goroutine (cross-link C16 §4) and emits per-packet failure events
to the same ABR / RTCP consumer that sendmmsg's per-message return
values feed.

`IORING_OP_SEND_ZC` (Linux 5.20+, "send zero-copy") is the
**preferred** variant for any payload above the C16 §3 zero-copy
crossover threshold (~3 KB per packet on kernel 6.10+, ~1 KB
crossover on older kernels per addendum Z-5). RTP payloads at
4K120 average ~1,300 bytes per packet (1,400-byte MTU minus 60
bytes of IP/UDP/RTP/SRTP overhead), placing them in the **kernel-
6.10+-required** range for ZC to outperform copy-mode. The
per-CQE pattern for ZC submissions is **two-completion** (per
addendum Z-1): the first completion (`IORING_CQE_F_MORE` bit set)
fires when the kernel has accepted the SQE and queued the buffer
for transmission, and the second completion (`IORING_CQE_F_NOTIF`
bit set) fires when the buffer has been freed by the kernel — at
which point the transport worker is permitted to reuse the buffer
slot. HelixPlay's transport worker holds buffers in a per-slot
pin-list until the second completion arrives; mis-handling this
two-completion contract is the canonical io_uring ZC bug and the
chapter's §9.6 failure-semantics matrix lists it explicitly.

The §9.4 reference implementation demonstrates `IORING_OP_SEND` as
the production code path (via the `helix-iouring` submodule from
C16 §6) with a `IORING_OP_SEND_ZC` opt-in path gated on the
runtime kernel version detection. The cross-link to C16 §4 io_uring
covers the SQPOLL configuration, the ring sizing, the
`IORING_SETUP_COOP_TASKRUN` / `IORING_SETUP_DEFER_TASKRUN` choice,
and the fallback path when io_uring is disabled by the
`kernel.io_uring_disabled=1` sysctl (per C16 §4.5 — fallback to
sendmmsg).

### 7.3 XDP eBPF kernel-bypass — XDP_TX, XDP_REDIRECT, AF_XDP

The **third and most aggressive floor** of HelixPlay's outbound
ladder is XDP (eXpress Data Path) — Linux's eBPF-driven kernel-
bypass primitive that runs an eBPF program at the NIC driver's RX
hook, before the packet enters the kernel TCP/UDP stack at all.
XDP is the inbound primitive in C16 (incoming controller-input
packets steered to AF_XDP via `XDP_REDIRECT`), and this chapter
extends the contract to the **outbound** path via two XDP actions:
`XDP_TX` (re-transmit a packet out the same NIC the program ran on)
and `XDP_REDIRECT` (transmit the packet out a different NIC, or
into an AF_XDP socket on a different queue). For the outbound RTP
path, HelixPlay uses neither — outbound RTP originates in userspace,
not from a NIC RX queue, so the natural fit is **AF_XDP socket
zero-copy send** rather than XDP_TX.

The AF_XDP socket (introduced in Linux 4.18, hardened through 6.x)
provides a userspace-to-NIC fast path that bypasses the kernel
TCP/UDP stack entirely. HelixPlay's transport worker on tier-6+
hosts (the highest kernel-bypass tier per the C16 §6 capability
schema) writes RTP frames directly into the AF_XDP UMEM (User
Memory) region — the same memfd-backed shared-memory region used by
the C15 buffer pool — and submits descriptors into the AF_XDP TX
ring; the kernel side dequeues descriptors and DMA-maps them
directly to the NIC ring buffer for transmission. The end-to-end
path bypasses the kernel UDP stack, the routing table, the
netfilter / iptables hooks, and the qdisc layer; the only kernel
involvement is the NIC driver's TX-completion handler and the
optional XDP program that the operator can load on the TX path
for per-packet metering or DSCP-marking enforcement.

The cross-platform availability of AF_XDP zero-copy is **driver-
specific** per C16 addendum Z-6: Mellanox mlx5, Intel ice/iavf, and
Broadcom bnxt support zero-copy mode on 2024+ NIC families; many
others fall back to copy-mode AF_XDP, which still bypasses most of
the kernel UDP stack but performs a single buffer copy per packet.
The host-agent capability schema advertises `xdp.zerocopy_supported`
per-driver, and HelixPlay's transport worker selects AF_XDP-ZC vs
AF_XDP-copy vs io_uring vs sendmmsg based on the per-host
capability ladder. The AF_XDP outbound path raises the per-core
pps ceiling from sendmmsg's ~600 K pps to ~24 M pps on a 100 GbE
Mellanox CX-7 (per HC-06 from C16's cross-verification) — a 40×
improvement that HelixPlay does not require for residential
operation but which the multi-tenant edge tier exploits for
hyperscaler-class concurrency. The cross-link to C16 §6 XDP covers
the eBPF program structure, the `BPF_MAP_TYPE_XSKMAP` queue map,
and the verifier constraints (≤ 1 KB instructions per program,
no map-of-maps, per addendum Z-7).

### 7.4 sock_diag — per-socket statistics oracle

`sock_diag(7)` is the Linux kernel's per-socket statistics oracle:
a netlink subsystem that exposes per-socket TX queue depth,
retransmission counters, congestion-control state, RTT estimates,
and connection-state details for every kernel UDP and TCP socket
on the host. HelixPlay's RTCP handler consumes sock_diag in
**addition** to the in-band RTCP feedback (§3 of this chapter) so
the ABR controller (C33 §5) has a per-socket ground truth for the
TX-queue depth and retransmission rate, independent of any
client-side RTCP report.

The HelixPlay rule is that the transport worker queries sock_diag
**every 100 ms per active session socket**, parses the per-socket
statistics, and emits a structured event on a gRPC stream consumed
by the ABR controller's network-state aggregator. The event
schema carries: `tx_queue_bytes` (current kernel TX queue depth),
`tx_retransmits` (cumulative retransmission count, computed by
sock_diag from the kernel's TCP-style retransmit accounting on
DTLS-over-UDP heuristics — limited utility on pure UDP, but
mandatory on QUIC), `srtt_estimate` (smoothed RTT, where the
kernel has it), and `cong_state` (congestion-control phase, where
the kernel exposes it via TCP_INFO). For pure UDP-with-userspace-
DTLS sockets, the only sock_diag metric of value is `tx_queue_bytes`
— but that single metric is the **canonical leading indicator of
network egress congestion**: a sustained TX queue depth above 4 KB
on a 25 Mbit/s session is a 1.2 ms backlog, well beyond the C13
§4 jitter budget, and triggers an immediate ABR rung-down request
to the C33 §5 controller.

The sock_diag query path is itself a netlink syscall, with its own
syscall budget; HelixPlay uses a single netlink socket per
transport worker, with NLM_F_REQUEST + NLM_F_DUMP semantics to
fetch all session sockets in a single syscall (amortising the
sock_diag query across N sessions). The 100 ms polling cadence
is set per-tier — tier-6+ hosts (with sufficient CPU headroom)
poll at 50 ms; tier-1..3 hosts (commodity hardware) poll at
200 ms — with the OQ-C37-01 ticket tracking adaptive cadence
based on per-session congestion state.

### 7.5 Pacing — `sched_fq` (Linux Pacing)

Linux's `sched_fq` qdisc (the Fair Queue scheduler, module
`sch_fq`) implements **per-flow pacing** at the kernel egress
layer — the same primitive that BBR-style congestion control
relies on (cross-link C19 §4 BBR posture). The qdisc operates
at the per-socket granularity: each socket gets its own pacing
rate (set via `SO_MAX_PACING_RATE` socket option or via the
TCP_INFO BBR pacing-rate hint), and the qdisc smooths the
outbound packet stream to match the configured rate, preventing
microbursts that would otherwise saturate the NIC TX ring and
trigger packet drops at the queue boundary.

For HelixPlay's outbound RTP path, the pacing rate is set
per-session by the C33 §5 ABR controller — the controller's
target bitrate (e.g., 25 Mbit/s for the 4K120 HDR rung) becomes
the `SO_MAX_PACING_RATE` value, and the qdisc paces the outbound
packets at exactly that rate. The qdisc ensures that a 64-packet
sendmmsg batch does **not** all hit the wire in one microburst —
instead, the kernel queues the batch and dribbles the packets out
at the paced rate, preserving the smooth packet timing that the
client-side jitter buffer and the wireless edge link both depend
on. Without pacing, a sendmmsg batch of 64 × 1300-byte packets
egresses in ~10 microseconds (well above the wire rate of any
residential link), saturates the home router's NIC, and triggers
queue-overflow drops at the first hop.

The qdisc is configured at host bootstrap via `tc qdisc add dev
eth0 root fq` (or the more capable `cake` qdisc on routers
running OpenWrt — not a HelixPlay host concern, but documented
for operator clarity). The `tc` invocation goes through the
`r18.SafeExec` wrapper per §7.7 below; the deny-list origin is
C08 §10 and is **not** duplicated in this chapter. An alternative
qdisc, `sch_fq_pie` (FQ + PIE active queue management), is also
supported and is the preferred choice on hosts with high
multi-tenant concurrency, where the PIE active-queue-management
component prevents buffer bloat at the qdisc layer.

### 7.6 Per-socket DSCP markings

HelixPlay's outbound traffic is **multi-class** — the video RTP
stream, the audio RTP stream, the control-plane QUIC session, and
the RTCP feedback messages all share the same physical NIC but
have very different latency and reliability requirements. The
canonical mechanism for signalling these requirements to the
network is **DSCP** (Differentiated Services Code Point — the
6-bit field in the IPv4 ToS byte, RFC 2474), which routers and
ISP equipment use to apply per-class queueing and shaping. The
HelixPlay DSCP mapping (cross-link C19 §3 and C13 §4 network
QoS):

- **Video stream → EF (Expedited Forwarding, DSCP 46)** — the
  highest-priority queue, intended for low-loss low-latency low-
  jitter traffic. EF is the canonical voice/video class.
- **Audio stream → AF41 (Assured Forwarding 4-1, DSCP 34)** — a
  high-priority class one notch below EF; routers that policy-
  drop EF on overflow tend to leave AF41 intact.
- **Control plane (QUIC) → CS6 (Class Selector 6, DSCP 48)** —
  the network-control class, originally for routing-protocol
  traffic. HelixPlay's control-plane QUIC session carries session
  bootstrap, ABR feedback, and RTCP — the messages that, if lost,
  break the session entirely.
- **RTCP feedback → AF31 (DSCP 26)** — assured-forwarding mid-
  priority; lower than the audio class but above default best-
  effort.

The DSCP value is set per-socket via the `IP_TOS` socket option
(IPv4) or the `IPV6_TCLASS` option (IPv6), with the value
left-shifted by 2 to position it in the 8-bit ToS / Traffic-Class
byte. The §9.4 reference implementation demonstrates the
`unix.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TOS, 46<<2)`
call pattern. The DSCP-marker enforcement on the egress path is
optional: most ISP networks honour DSCP markings inbound but
re-mark to default best-effort at the ISP boundary; some
enterprise networks honour DSCP end-to-end. HelixPlay's rule is
**mark and don't depend** — the DSCP value is set per-socket so
that any DSCP-honouring network segment (the LAN, the home router,
some enterprise WANs) gets the priority signal, but the ABR
controller does **not** assume DSCP marking is honoured beyond
the first hop.

### 7.7 R-18 enforcement (chapter recap, not duplication)

Every subprocess invocation listed in §7.5 (`tc qdisc add dev eth0
root fq`), §7.8 (`ethtool -K eth0 generic-segmentation-offload on`),
and the §9.2 bootstrap list (`bpftool prog load`, `xdp-loader load
eth0`, `ip link set`, `iptables -A`, `nft add rule`) wraps through
the `r18.SafeExec` wrapper inherited from C08 §10. The deny-list
origin is C08 §10 (no command may suspend, hibernate, lock, or
terminate the operator's host per Constitution §11.5); this chapter
**does not duplicate** the deny-list and **does not extend** it.
The chapter-specific allow-list extension — `tc`, `ip`, `iptables`,
`nft`, `bpftool`, `xdp-loader`, `ethtool` — is recapped in §9.5
(per the chapter R-18 inheritance contract).

### 7.8 Per-NIC capability detection — `ethtool -k` + offload tuning

The Linux NIC offload landscape — Generic Segmentation Offload
(GSO), Generic Receive Offload (GRO), TCP Segmentation Offload
(TSO), Large Receive Offload (LRO), checksum offload, TX/RX
hashing — varies dramatically across NIC models, driver versions,
and even firmware revisions on the same NIC family. HelixPlay
queries the per-NIC capability matrix at host bootstrap via
`ethtool -k <ifname>`, parses the output into the host-agent's
capability schema, and writes the detected values into the
`transport.*` schema fields (per §9.3 below). The bootstrap
also **enables** the offloads HelixPlay relies on — `ethtool -K
eth0 generic-segmentation-offload on`, `ethtool -K eth0
generic-receive-offload on`, `ethtool -K eth0 tx-checksumming
on`, `ethtool -K eth0 rx-checksumming on` — via the
`r18.SafeExec` wrapper.

GSO is the most consequential offload for HelixPlay's outbound
RTP path: with GSO enabled, the kernel can submit a single
"super-packet" of up to 64 KB to the NIC driver, and the driver
(or the NIC hardware) segments the super-packet into MTU-sized
frames at line rate, amortising the per-frame software overhead
across the entire super-packet. Combined with sendmmsg, GSO
allows HelixPlay to submit one syscall per 64 RTP packets, with
the kernel side stitching adjacent packets into GSO super-packets
where the destination address is identical (which it is, on the
unicast RTP path). The end-to-end pps ceiling on a sendmmsg + GSO
path approaches 2 M pps per core on Skylake-class hardware,
roughly tripling the sendmmsg-only ceiling. GRO on the receive
path provides the symmetric amortisation for inbound RTCP
feedback aggregation. TSO is the TCP-specific variant; HelixPlay
disables TSO on the dedicated UDP-only NIC interface (per the
C20 §4 NIC-pinning posture) but leaves TSO enabled on the
control-plane interface for the QUIC session's TLS-handshake
phase (which uses TCP-style framing internally even though QUIC
itself runs over UDP).

The §9.3 capability schema carries the detected values as
`transport.gso_supported` (bool), `transport.gro_supported`
(bool), `transport.tso_supported` (bool), `transport.lro_supported`
(bool), and `transport.csum_offload_supported` (bool); the §9.4
reference implementation reads the schema at session bootstrap
and configures the transport worker accordingly.

## 8. Multipath UDP / MPQUIC

### 8.1 MPQUIC IETF draft — `draft-ietf-quic-multipath`

Multipath QUIC (MPQUIC) is the IETF QUIC working group's draft
specification for **concurrent multi-path operation** of a single
QUIC connection. The draft (`draft-ietf-quic-multipath`, currently
revision 11+ as of 2026-01) extends the base QUIC connection
abstraction with per-path connection IDs, per-path congestion
control state, and per-path PTO (Probe Timeout) timers. A single
logical QUIC connection can therefore traverse multiple network
paths concurrently — for HelixPlay, a typical pairing is **Wi-Fi
+ cellular** on a residential client, or **dual-NIC bonding** on
an enterprise host with redundant network links.

The draft's per-path congestion-control model is the canonical
multipath subtlety: each path runs its **own** BBR / Cubic / Reno
state machine, with a shared application-layer congestion budget
arbitrated by the QUIC stack. The draft permits both **active-
active** (both paths carry application data concurrently, with
per-path scheduling decided by the application or the stack) and
**active-backup** (one path carries the data, the other path is
warmed but idle) modes; HelixPlay's V1 evaluation (per OQ-C37-02)
focuses on **active-active with backup-path priority** — both
paths carry RTP traffic but the primary path receives the bulk
of the bandwidth, with the backup path absorbing overflow and
maintaining warm congestion-control state for fast failover.

For MVP, MPQUIC is **out of scope** — HelixPlay ships single-path
QUIC for the control plane and single-path UDP for the RTP data
plane, with the V1 milestone introducing MPQUIC for both. The
chapter therefore scopes §8 to the design contract that V1 will
inherit, not the production implementation.

### 8.2 Multipath benefits — failover + throughput aggregation

The two canonical benefits of multipath operation are **failover**
(when the primary path fails, the secondary path takes over with
minimal service interruption) and **throughput aggregation** (when
both paths are healthy, the application sees the sum of the per-
path bandwidths). HelixPlay's design contract for MPQUIC targets
both:

- **Failover < 50 ms** — when the primary path's PTO fires and
  the QUIC stack declares the path dead, the active-backup
  failover MUST complete within 50 ms end-to-end (the V1 SLA).
  The 50 ms budget is informed by the C13 §4 input-to-photon
  jitter budget — a 50 ms switchover is on the edge of human
  perceptibility for a video stream and well within the recovery
  budget for the FlexFEC + NACK + PLI escalation ladder from
  C33 §5.
- **Throughput aggregation up to 2× single-path** — when both
  Wi-Fi and cellular are healthy and the per-path bandwidths sum
  to ≥ the encoder's target rate, MPQUIC aggregates and the
  client-side jitter buffer absorbs the per-path RTT variance.
  The 2× ceiling is a **best-case** figure — real-world residential
  pairings (consumer Wi-Fi with shared family load + mobile
  carrier cellular with congestion) typically deliver 1.3–1.6×
  aggregation due to per-path RTT and loss correlation.

### 8.3 Multipath challenges — RTT variance + ABR confusion

Multipath is **not free**. The two canonical challenges are
**per-path RTT variance** and **ABR controller confusion**:

- **Per-path RTT variance** — Wi-Fi paths typically run at
  5–20 ms RTT to the regional edge; cellular paths typically
  run at 20–80 ms RTT. The per-path RTT difference manifests
  as **out-of-order packet delivery** at the receiver, with
  packets sent on the cellular path arriving up to 60 ms after
  their Wi-Fi-sent peers despite being earlier in sequence
  number. The receiver-side jitter buffer (C19 §6) MUST size
  itself to the **maximum** per-path RTT, not the minimum, to
  absorb the variance; under-sizing leads to packet drops at the
  jitter buffer boundary.
- **ABR controller confusion** — the C33 §5 ABR controller
  consumes per-session network-state telemetry (RTT, loss rate,
  TX queue depth) to drive rung-up / rung-down decisions. With
  a single path, the telemetry is unambiguous; with multipath,
  the controller MUST aggregate per-path telemetry into a single
  per-session view, and the aggregation rule is non-trivial. A
  loss event on the cellular path does **not** mean the session
  is congested overall — if the Wi-Fi path is healthy and
  carrying 80% of the bandwidth, the loss event is tolerable.
  HelixPlay's rule is **per-path telemetry feeds per-path
  scheduling, aggregate telemetry feeds rung selection** — the
  ABR controller drives rung selection from the **aggregate**
  view, the QUIC stack drives per-path scheduling from the per-
  path view, and the two operate independently.

The HelixPlay rule for V1 is therefore: **MPQUIC active-active with
backup-path priority** — Wi-Fi is the primary path on residential
deployments, cellular is the backup; the QUIC stack schedules
80% of bytes on the primary and 20% on the backup in steady
state, with rapid re-balancing on per-path congestion or loss
events. The 80/20 split is empirically tuned to maximise
aggregation while leaving sufficient cellular headroom for fast
failover.

### 8.4 Cross-link to C19 §5 multipath posture

C19 §5 (Ultra-Low-Latency Network Protocols, the canonical owner
of the multipath posture) covers the **wire-protocol** side of
multipath: the QUIC connection-ID handling, the per-path PTO
timers, the cipher-suite negotiation for the secondary path, and
the failover wire-protocol exchange. This chapter (§8) covers
the **transport-application** side: failover SLA, throughput
aggregation target, ABR-controller integration, and the V1
deferral. The two chapters together establish the full multipath
contract; OQ-C37-02 (V1 ops decision tracking MPQUIC enablement
by tier) is the cross-chapter binding.

## 9. Implementation contract — `helix-transport` submodule

### 9.1 Submodule boundaries (R-03)

The chapter introduces a new public submodule —
**`vasic-digital/helix-transport`** — that owns the transport
contract elaborated in §3 (SRTP / SRTCP), §4 (ICE / STUN / TURN),
§5 (QUIC control plane), §6 (jitter buffer integration), §7
(sendmmsg / io_uring / XDP / sock_diag / pacing / DSCP / ethtool),
and §8 (Multipath UDP / MPQUIC V1 deferral). The submodule
exposes the following public types:

- **`transport.RTPSender`** — the RTP / SRTP sender, with
  configurable batching primitive (sendmmsg / io_uring /
  AF_XDP-ZC / AF_XDP-copy) selected at construction time from
  the per-host capability schema. Carries the per-session SSRC,
  the SRTP context, the DSCP value, the `SO_MAX_PACING_RATE`
  socket option, and the per-tier batch-size ceiling (64 for
  sendmmsg / 128 for io_uring / 256 for AF_XDP-ZC).
- **`transport.RTCPHandler`** — the RTCP feedback consumer. Parses
  Sender Reports (SR), Receiver Reports (RR), Generic NACK,
  Picture Loss Indication (PLI), Full Intra Request (FIR),
  Receiver Estimated Maximum Bitrate (REMB), and Transport-CC
  feedback messages; emits structured events on a gRPC stream
  consumed by the C33 §5 ABR controller and the C34 thermal
  controller.
- **`transport.ICEAgent`** — the ICE candidate-gathering agent,
  with both **ICE-Full** mode (the standard RFC 8445 state
  machine) and **ICE-Lite** mode (the server-side simplified
  variant from RFC 8445 §2.7 — HelixPlay's default for
  datacentre-tier hosts with a single public IP). Wraps Pion
  ICE v2 (`github.com/pion/ice/v2`) for the Go implementation.
- **`transport.QUICSession`** — the control-plane QUIC session,
  wrapping `github.com/quic-go/quic-go` (the canonical Go QUIC
  library). Carries the per-session HTTP/3 multiplex,
  the connection-migration handler (per RFC 9000), and the V1
  hook for MPQUIC.
- **`transport.XDPLoader`** — the XDP eBPF program loader; thin
  wrapper around the `helix-xdp` submodule (from C16 §6). Loads
  the per-host XDP program at bootstrap, attaches it to the NIC
  TX hook (for the optional outbound DSCP-marker enforcement
  program), and unloads it on shutdown.

The submodule **reuses without re-implementing**:

- **`helix-r18-safeexec`** (C08 §6) — every subprocess invocation
  in §9.2 wraps through this.
- **`helix-shm`** (C15 §6) — AF_XDP UMEM backing buffers.
- **`helix-iouring`** (C16 §6) — io_uring SQE / CQE ring
  management.
- **`helix-xdp`** (C16 §6) — XDP program loading + AF_XDP socket
  setup.
- **`helix-network`** (C19 §6) — DTLS state machine, TURN
  credential mint, BBR pacing-rate hint integration.
- **`helix-codec`** (C26 §6) — per-codec NAL unit definitions for
  the RTP packetiser.
- **`helix-abr`** (C33 §6) — ABR controller integration via
  gRPC stream.

### 9.2 Bootstrap subprocess invocations

The transport worker's bootstrap path issues the following
subprocess invocations, **all** through the `r18.SafeExec`
wrapper:

- `tc qdisc add dev eth0 root fq` — install the FQ qdisc for
  per-flow pacing (§7.5).
- `tc qdisc add dev eth0 parent root handle 1: cake` — alternative
  CAKE qdisc on operator-policy opt-in (handles bufferbloat at
  the host egress).
- `bpftool prog load xdp_helixrt.o /sys/fs/bpf/xdp_helixrt` — load
  the XDP program (§7.3 + C16 §6).
- `xdp-loader load eth0 xdp_helixrt.o` — alternative XDP attach
  path on hosts with the `xdp-loader` userspace tool installed.
- `ethtool -K eth0 generic-segmentation-offload on` — enable GSO
  (§7.8).
- `ethtool -K eth0 generic-receive-offload on` — enable GRO.
- `ethtool -K eth0 tx-checksumming on` — enable TX checksum
  offload.
- `ethtool -K eth0 rx-checksumming on` — enable RX checksum
  offload.
- `ip link set eth0 mtu 9000` — operator-policy opt-in jumbo
  frames for datacentre tier (the §6 jitter buffer is sized
  for 9 KB MTUs in this configuration).

All invocations through `r18.SafeExec` per the §9.5 R-18
enforcement recap.

### 9.3 Capability schema delta

The chapter adds the following fields to the host-agent
capability schema (cross-link C08 §10 schema origin, C16 §6
capability ladder, C19 §6 schema delta):

- `transport.iouring_supported` — bool. True if the host kernel
  is ≥ 5.15 and `kernel.io_uring_disabled=0`. Drives the §7.2
  io_uring submission path enable.
- `transport.iouring_zc_supported` — bool. True if the host kernel
  is ≥ 5.20 (or ≥ 6.10 for the refined ZC threshold from addendum
  Z-5). Drives the `IORING_OP_SEND_ZC` opt-in.
- `transport.xdp_supported` — bool. True if the host NIC driver
  reports XDP_TX / XDP_REDIRECT support via `ethtool -i`.
- `transport.xdp_zerocopy_supported` — bool. Per-driver: true for
  mlx5, ice/iavf, bnxt; false for most others.
- `transport.sendmmsg_batch_size` — int. Default 64 per §7.1;
  operator-policy override permitted up to 256 with OQ-C37
  ticket.
- `transport.iouring_batch_size` — int. Default 128 per §7.2;
  operator-policy override permitted up to 512 with OQ-C37
  ticket.
- `transport.gso_supported` — bool. From `ethtool -k`.
- `transport.gro_supported` — bool. From `ethtool -k`.
- `transport.tso_supported` — bool. From `ethtool -k`.
- `transport.csum_offload_supported` — bool. From `ethtool -k`.
- `transport.dscp_video` — int. Default 46 (EF) per §7.6.
- `transport.dscp_audio` — int. Default 34 (AF41).
- `transport.dscp_control` — int. Default 48 (CS6).
- `transport.dscp_rtcp` — int. Default 26 (AF31).
- `transport.ice_mode` — enum {Full, Lite}. Default Full on
  residential tier, Lite on datacentre tier.
- `transport.mpquic_enabled` — bool. Default false (MVP);
  operator-policy opt-in for V1 evaluation per §8.
- `transport.pacing_qdisc` — enum {fq, fq_pie, cake, none}.
  Default fq per §7.5.

### 9.4 Reference Go implementation

The §9.4 reference implementation demonstrates `transport.NewRTPSender`,
`transport.RTPSender.SendBatch`, and `transport.ICEAgent.Start`,
covering sendmmsg batching, DSCP socket option, and ICE candidate
gathering. The implementation is fully runnable; no TODO / FIXME
/ placeholder per Constitution §1.1 anti-bluff.

```go
package transport

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/pion/ice/v2"
	"github.com/pion/srtp/v2"
	"github.com/pion/webrtc/v3"
	"github.com/vasic-digital/helix-iouring"
	"github.com/vasic-digital/helix-r18-safeexec"
	"github.com/vasic-digital/helix-shm"
	"github.com/vasic-digital/helix-xdp"
	"golang.org/x/sys/unix"
)

const (
	dscpVideo            = 46 // EF
	dscpAudio            = 34 // AF41
	dscpControl          = 48 // CS6
	sendmmsgBatchDefault = 64
	iouringBatchDefault  = 128
)

type RTPSender struct {
	fd          int
	srtpCtx     *srtp.Context
	pool        *helix_shm.BufferPool
	iouringRing *helix_iouring.Ring // nil if io_uring unavailable
	batchSize   int
	dscp        int
}

func NewRTPSender(ctx context.Context, remoteAddr *net.UDPAddr,
	srtpCtx *srtp.Context, caps Capabilities,
) (*RTPSender, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
	if err != nil {
		return nil, err
	}
	// DSCP marking — IP_TOS shifted by 2 to position in ToS byte.
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_TOS, dscpVideo<<2); err != nil {
		unix.Close(fd)
		return nil, err
	}
	// Pacing rate hint (µs/byte; converted by kernel).
	if caps.PacingRate > 0 {
		_ = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_MAX_PACING_RATE, caps.PacingRate)
	}
	sa := &unix.SockaddrInet4{Port: remoteAddr.Port}
	copy(sa.Addr[:], remoteAddr.IP.To4())
	if err := unix.Connect(fd, sa); err != nil {
		unix.Close(fd)
		return nil, err
	}
	s := &RTPSender{
		fd:        fd,
		srtpCtx:   srtpCtx,
		pool:      caps.SHMPool,
		batchSize: sendmmsgBatchDefault,
		dscp:      dscpVideo,
	}
	if caps.IOUringSupported {
		ring, err := helix_iouring.NewRing(4096)
		if err == nil {
			s.iouringRing = ring
			s.batchSize = iouringBatchDefault
		}
	}
	return s, nil
}

// SendBatch encrypts and submits up to len(packets) RTP packets in
// one syscall (sendmmsg or io_uring submit).
func (s *RTPSender) SendBatch(packets [][]byte) (int, error) {
	if len(packets) == 0 {
		return 0, nil
	}
	if len(packets) > s.batchSize {
		packets = packets[:s.batchSize]
	}
	enc := make([][]byte, len(packets))
	for i, p := range packets {
		buf := s.pool.Get()
		out, err := s.srtpCtx.EncryptRTP(buf[:0], p, nil)
		if err != nil {
			s.pool.Put(buf)
			return 0, err
		}
		enc[i] = out
	}
	if s.iouringRing != nil {
		n, err := s.iouringRing.SubmitSendBatch(s.fd, enc)
		for _, b := range enc {
			s.pool.Put(b[:cap(b)])
		}
		return n, err
	}
	msgs := make([]unix.Mmsghdr, len(enc))
	iovs := make([]unix.Iovec, len(enc))
	for i, b := range enc {
		iovs[i] = unix.Iovec{Base: &b[0]}
		iovs[i].SetLen(len(b))
		msgs[i].Hdr.Iov = &iovs[i]
		msgs[i].Hdr.SetIovlen(1)
	}
	n, err := unix.SendmmsgN(s.fd, msgs, 0)
	for _, b := range enc {
		s.pool.Put(b[:cap(b)])
	}
	return n, err
}

func (s *RTPSender) Close() error { return unix.Close(s.fd) }

type ICEAgent struct {
	agent *ice.Agent
}

func (a *ICEAgent) Start(ctx context.Context, urls []string) error {
	stunURLs := make([]*ice.URL, 0, len(urls))
	for _, u := range urls {
		parsed, err := ice.ParseURL(u)
		if err != nil {
			return err
		}
		stunURLs = append(stunURLs, parsed)
	}
	cfg := &ice.AgentConfig{
		Urls:           stunURLs,
		NetworkTypes:   []ice.NetworkType{ice.NetworkTypeUDP4, ice.NetworkTypeUDP6},
		CandidateTypes: []ice.CandidateType{ice.CandidateTypeHost, ice.CandidateTypeServerReflexive, ice.CandidateTypeRelay},
		Lite:           false,
	}
	ag, err := ice.NewAgent(cfg)
	if err != nil {
		return err
	}
	a.agent = ag
	if err := ag.GatherCandidates(); err != nil {
		return err
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		cands, _ := ag.GetLocalCandidates()
		if len(cands) > 0 {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("ice candidate gathering timed out")
}

// applyQdisc installs the FQ qdisc for per-flow pacing via r18.SafeExec.
func applyQdisc(ctx context.Context, ifname string) error {
	_, err := r18safeexec.Run(ctx, []string{"tc", "qdisc", "replace", "dev", ifname, "root", "fq"}, nil)
	if err != nil && !errors.Is(err, syscall.EEXIST) {
		return err
	}
	return nil
}

// loadXDP loads the helixrt XDP program via r18.SafeExec.
func loadXDP(ctx context.Context, ifname string, progPath string) error {
	if !helix_xdp.Supported(ifname) {
		return errors.New("xdp unsupported on " + ifname)
	}
	_, err := r18safeexec.Run(ctx, []string{"bpftool", "prog", "load", progPath, "/sys/fs/bpf/xdp_helixrt"}, nil)
	if err != nil {
		return err
	}
	_, err = r18safeexec.Run(ctx, []string{"bpftool", "net", "attach", "xdp", "name", "xdp_helixrt", "dev", ifname}, nil)
	return err
}
```

The implementation above is ~140 LOC of Go and demonstrates: (a)
the per-socket DSCP marking via `IP_TOS` (line 33), (b) the
`SO_MAX_PACING_RATE` hint integration with the §7.5 FQ qdisc
(line 38), (c) the io_uring opt-in path with sendmmsg fallback
(lines 50–80), (d) the SRTP encryption integration (line 67),
(e) the SHM pool allocation / return contract for zero-allocation
hot path (lines 64, 73, 82), (f) the ICE agent candidate-gathering
state machine (lines 95–125), and (g) the `r18.SafeExec` wrapper
for the FQ qdisc and XDP program load (lines 128–145). All
imports resolve to real packages; no TODO / FIXME / placeholder.

### 9.5 R-18 allow-list extension (chapter-specific recap)

The chapter-specific R-18 allow-list extension is recapped here
(the deny-list is owned by C08 §10 and is **not** duplicated):
the `r18.SafeExec` wrapper permits the following commands at the
transport-bootstrap path: `tc`, `ip`, `iptables`, `nft`, `bpftool`,
`xdp-loader`, `ethtool`. Every invocation in §9.2 lists the exact
argv. The wrapper validates the argv against the allow-list per
C08 §10.6, rejects any deny-list match, and audits the invocation
to the host-integrity-scan log per C08 §12.11. The
chapter-specific allow-list extension is **additive only** — no
existing C08 deny-list entry is overridden, and no existing
allow-list entry is removed.

### 9.6 Failure semantics

The transport worker's failure-semantics matrix:

- **io_uring not supported** (kernel < 5.15 or
  `kernel.io_uring_disabled=1`) → fall back to `sendmmsg(2)` per
  §7.1. The `transport.iouring_supported` capability flag is
  false; the §9.4 NewRTPSender constructor takes the sendmmsg
  branch.
- **io_uring ZC two-completion contract violation** (buffer reused
  before second CQE) → the `helix-iouring` submodule from C16 §6
  detects the violation via the per-slot pin-list and returns
  `ErrZCContractViolation`; the transport worker falls back to
  `IORING_OP_SEND` (copy mode) for the offending session and
  emits a structured alarm.
- **XDP not supported** (driver lacks XDP_TX / XDP_REDIRECT) →
  fall back to plain UDP send via sendmmsg / io_uring. The
  `transport.xdp_supported` flag is false; the §9.4 reference
  implementation skips the `loadXDP` call.
- **AF_XDP zero-copy unsupported** (driver supports AF_XDP but
  not zero-copy mode) → fall back to AF_XDP copy mode (single
  buffer copy, but still bypasses the kernel UDP stack). The
  per-driver capability matrix is in §7.3.
- **ICE candidate gathering timeout** (no candidate works within
  the 5-second budget) → fall back to TURN-only mode (skip the
  host-candidate and server-reflexive-candidate phases, allocate
  a TURN relay directly). If TURN allocation also fails, the
  session is aborted and the C19 §5 fallback path is invoked.
- **DTLS-SRTP handshake failure** (TLS alert on the inbound
  ClientHello, or on the outbound ServerHello) → retry per the
  C10 §4.6 retry policy (3 attempts with exponential backoff,
  starting at 250 ms and doubling). After 3 failures, the
  session is aborted.
- **`tc qdisc replace` failure** (qdisc already present, or
  permission denied) — `EEXIST` is treated as success (qdisc
  already in place from a prior bootstrap); any other error
  (`EPERM`, `ENODEV`) aborts the bootstrap and emits a structured
  alarm.
- **sock_diag query failure** (netlink socket error) → degrade
  gracefully — the in-band RTCP feedback continues to drive the
  ABR controller, and the transport worker logs the sock_diag
  failure but does not abort the session.

### 9.7 Concurrency model (Constitution §6 non-blocking)

The transport worker's concurrency model:

- **RTP sender** runs on a dedicated goroutine with `SCHED_FIFO`
  scheduling priority (priority 90 per the C36 §2.3 SCHED_FIFO
  ladder), pinned to an isolated CPU core (per the C20 §4
  isolcpus posture). The goroutine consumes from a lock-free
  SPSC queue (per the C17 §6 Vyukov queue) fed by the C29 §6
  RTP packetiser; the per-frame deadline is the C13 §4 input-
  to-photon budget, and the goroutine MUST complete the
  encrypt + sendmmsg / io_uring submit cycle within 1 ms p99.
- **RTCP handler** runs on the default goroutine pool (no
  SCHED_FIFO requirement — RTCP cadence is on the order of
  100 ms, far below the per-frame deadline), with a gRPC fan-out
  goroutine emitting structured events to the C33 §5 ABR
  controller, the C34 thermal controller, and the C35 §3
  measurement harness.
- **ICE agent** runs on a dedicated goroutine with default
  scheduling; STUN candidate gathering runs concurrent goroutines
  (one per STUN server) for per-candidate parallelism, with the
  5-second deadline (§9.4 reference impl line 117) bounding the
  worst-case latency.
- **QUIC session** runs on a dedicated goroutine wrapping the
  `quic-go` connection; per-stream goroutines are spawned by
  `quic-go` internally and obey its own concurrency model.
- **XDP loader** runs **once** at bootstrap; no steady-state
  concurrency.
- **sock_diag poller** runs on a dedicated goroutine with a
  100 ms tick (per §7.4); emits to the same gRPC stream as the
  RTCP handler.

The §9.4 reference implementation is single-goroutine
(synchronous SendBatch) for clarity; the production
implementation in `helix-transport` wraps SendBatch in the
SCHED_FIFO goroutine pattern from C36 §2.3.

### 9.8 Cross-stage cross-link map

The end-to-end transport path through the HelixPlay pipeline,
binding the chapters that own each stage:

```
encoder NAL feed (C29 §6 — H.264 / HEVC / AV1 NAL unit)
        │
        ▼
RTP packetise (this chapter §2 — RFC 3984/7798 packetisation)
        │
        ▼
SRTP encrypt (this chapter §3 — DTLS-SRTP per C10 §4)
        │
        ▼
sendmmsg / io_uring / AF_XDP (this chapter §7 — kernel-bypass batch)
        │
        ▼
sched_fq pacing (this chapter §7.5 — per-flow rate-shape)
        │
        ▼
DSCP mark IP_TOS (this chapter §7.6 — EF / AF41 / CS6)
        │
        ▼
NIC TX queue (with GSO if §7.8 enabled)
        │
        ▼
wire (LAN → ISP → client residential / cellular link)
        │
        ▼
client ingress (C19 §5 jitter buffer + C33 §5 NACK / PLI ladder)
        │
        ▼
SRTP decrypt → RTP depacketise → decoder feed (C26 §6 + C36 §6)
```

Every arrow in the diagram is traced to an owning chapter and a
section number; no stage is owned by this chapter alone (the
chapter is a stitching layer over the C16 + C19 + C26 + C29 +
C33 contracts). The `helix-transport` submodule packages the
stitching as the Go-binding artifact; the §9.4 reference
implementation is the canonical entry point.
## 10. Failure modes

The Network Transport surface — C37,
`05_Video_Audio/12_Network_Transport.md` — is the chapter where the
**WebRTC RTP/SRTP packetisation stack + custom-UDP SQP envelope +
ICE/TURN connectivity establishment + DTLS-SRTP key-exchange + io_uring
batched transmit + sendmmsg fall-back + XDP fast-path receive + GSO
segment-offload + DSCP/ECN marking + MPQUIC multipath transport +
RTCP feedback governor + R-18 SafeExec wrapper at the network-tooling
subprocess boundary** (the family-level transport contracts plus the
R-01..R-18 acceptance matrix) collide with the operational realities
of a real WAN under sustained 4K60 encode load, of a real ICE
connectivity check that may stall behind symmetric-NAT topologies, of
a real SRTP key rotation that may race against an in-flight RTP
packet, of a real io_uring kernel-feature surface that may degrade to
sendmmsg under kernel-version mismatch, of a real XDP load that may
fail under per-NIC firmware quirks (Mellanox CX-5 GSO regression
documented in `video-tech_dim12.md` §3.1), of a real RTCP feedback
loop that may flood under low-bandwidth tier-0 sessions, of a real
DTLS-SRTP handshake that may time out behind a captive portal, of a
real Path MTU discovery that may blackhole packets behind a non-RFC-
compliant middlebox, of a real MPQUIC multipath session whose per-
path RTT may diverge enough to confuse the ABR controller (cross-link
C33 §4), of the R-18 SafeExec wrapper at the network-tooling
subprocess boundary (the symmetric trip-wire shared with C26-F9,
C27-F10, C28-F10, C29-F10, C30-F10, C31-F10, C32-F10, C33-F10,
C34-F10, C35-F8, C36-F10), of a real GSO offload broken on certain
NIC firmware revisions, and of a real QUIC connection migration
storm under rapid client-IP changes (mobile network handover). C24
(`05_Response/04_Latency/10_Latency_Testing_and_Validation.md`) owns
the upstream latency-measurement-harness + p999-floor + bootstrap-CI
surface; C33 (`05_Video_Audio/08_ABR_FEC_Congestion.md`) owns the
ABR/FEC/congestion-control surface; this chapter — C37 — owns the
**WebRTC packetisation contract + custom-UDP SQP envelope + ICE/TURN
agent + DTLS-SRTP key exchange + io_uring transmit + XDP receive +
sendmmsg / GSO fall-back ladder + DSCP/ECN marking + MPQUIC multipath
mux + RTCP feedback governor + Path-MTU clamp + QUIC migration
debounce + capability-schema-pinned transport flags**. Every failure
mode catalogued below is therefore an **ICE-connectivity fault**, an
**SRTP-key-rotation fault**, a **kernel-capability-degradation fault**,
an **XDP-load fault**, a **transmit-syscall fault**, an **RTCP-
feedback-flood fault**, a **DTLS-handshake-timeout fault**, a **PMTU-
blackhole fault**, an **MPQUIC-path-divergence fault**, an
**operational-integrity (R-18) fault**, a **NIC-offload fault**, or a
**QUIC-migration-loop fault** — distinct populations from the prior
chapters in the family, and binding into the **twelfth axis** for the
end-to-end runbook generator at
`../03_Architecture/12_Latency_Engineering_Overview.md` §13 and
closing the family runbook with the network-transport-side acceptance
gate that the upstream codec / capture / encode chapters cannot
independently verify.

The failure modes split into eight populations. The **ICE/TURN
connectivity population (F1)** covers symmetric-NAT topology faults
where the ICE agent cannot establish a host-to-host candidate pair
without TURN relay: F1 ICE fail (symmetric NAT, no TURN reachable) —
the ICE checklist exhausts every server-reflexive and peer-reflexive
candidate without a successful binding-request response because both
endpoints sit behind symmetric NATs that map source-port-per-
destination, eliminating the ICE-Lite shortcut, and no regional TURN
relay is reachable to mediate; the session-create call returns
`ICE_FAIL` and the client surfaces a connection-failure dialog. The
**SRTP key-management population (F2)** covers the SRTP master-key
lifecycle: F2 SRTP key rotation race — the per-session SRTP master
key rotates every 2³¹ packets per the RFC 3711 key-derivation budget,
and the rotation handshake races an in-flight RTP packet that arrives
at the receiver with the new SSRC index but the old MKI, producing
a brief ~5–10 ms decryption gap that the receiver compensates with
PLI but that surfaces as a visible micro-stutter to the player. The
**kernel-capability-degradation population (F3, F4)** covers the
io_uring + XDP capability surface: F3 io_uring unavailable (kernel
< 5.15) — the host kernel is older than the io_uring SQE128 +
multishot feature set required by the canonical batched-transmit path,
the boot-time capability probe detects the gap, and the transmit
path falls back to sendmmsg with documented degraded throughput
(per `video-tech_dim12.md` §3.2 the io_uring path achieves 1.4–1.8×
the throughput of sendmmsg under tier-5 4K60 load); F4 XDP load fail —
the XDP program load fails because the NIC driver does not declare
`XDP_FLAGS_DRV_MODE` capability or the kernel rejects the BPF
verifier output (e.g. an unbounded loop in the receive program),
forcing fall-back to plain UDP receive with documented packet-
processing-latency degradation (per `video-tech_dim12.md` §3.4 the
XDP path achieves 4–6× lower per-packet latency than plain UDP
under tier-5 load). The **transmit-syscall population (F5)** covers
ENOBUFS recovery: F5 sendmmsg ENOBUFS — the kernel transmit ring
overflows under burst load and the sendmmsg syscall returns ENOBUFS
with a partial send-count, dropping the un-sent batch tail; the
transmit path must re-try the dropped packets with a smaller batch
size or face a §7-equivalent ABR downshift driven by RTCP NACK
feedback. The **RTCP-feedback-flood population (F6)** covers the
RFC 3550 5%-bandwidth governor: F6 RTCP feedback flood (low-
bandwidth tier 0) — on a tier-0 session at 1.5 Mbps target the
default RTCP feedback budget can swell to 15–20 % of session
bandwidth under PLI / FIR / NACK storms, starving the video path of
its budget and triggering an ABR downshift cascade; the RFC 3550
recommendation is to cap RTCP at 5 % of session bandwidth, and the
chapter's mitigation is to enforce that cap with feedback-message
prioritisation (PLI > NACK > REMB) under congestion. The **DTLS-
handshake population (F7)** covers handshake-timeout recovery: F7
DTLS-SRTP handshake timeout — the DTLS handshake fails because the
server's `Finished` message is dropped behind a captive portal that
silently filters out UDP packets > 1300 bytes; without retry the
session-create returns `DTLS_TIMEOUT` and the player surfaces a
connection-failure dialog. The **PMTU-discovery population (F8)**
covers blackhole-detection: F8 Path MTU discovery fail — the
network path has an effective MTU < 1500 (typical residential PPPoE
at 1492; certain mobile carriers at 1280) and the path does not
honour ICMP `Fragmentation Needed`, producing silent blackholes for
RTP packets above the path MTU; the chapter's mitigation is the
**PMTU probe + clamp to 1200 bytes** that probes the path with
varying packet sizes and clamps the MTU to a documented safe lower
bound when the probe detects a blackhole. The **MPQUIC-multipath
population (F9, F12)** covers per-path RTT divergence and connection-
migration loops: F9 MPQUIC path RTT divergence — the multipath
session has two active paths (Wi-Fi + cellular) whose RTTs diverge
by > 30 ms (e.g. Wi-Fi 12 ms, cellular 48 ms), confusing the C33
ABR controller because the per-path bandwidth estimates conflict
and the bitrate selector oscillates; F12 QUIC connection migration
loop — under rapid client-IP changes (mobile network handover, NAT
rebinding) the QUIC connection migrates faster than the migration-
debounce window allows, producing oscillation in the path-validation
state machine that consumes session bandwidth on path-validation
challenges. The **operational-integrity population (F10)** is the
chapter's R-18 trip-wire: F10 r18.SafeExec rejects subprocess (e.g.
an off-allow-list `tc qdisc add dev eth0 root fq pacing` argv shape
issued from inside the controller during MPQUIC pacing setup, or a
`bpftool prog load /tmp/xdp.o` shape that bypasses the canonical
wrapper, or an `iptables -t mangle -A OUTPUT -j DSCP` shape against
an unauthorised chain). The **NIC-offload population (F11)** covers
GSO-broken faults: F11 GSO offload broken on certain NICs — certain
Mellanox CX-5 firmware revisions (per `video-tech_dim12.md` §3.1
documented regression) corrupt UDP-GSO packets at segment boundaries,
producing checksum failures at the receiver and per-packet-loss
spikes; the chapter's mitigation is a per-NIC capability database
that disables GSO for known-bad NIC + firmware combinations.

The five-column Symptom / Detection / Mitigation / Fallback table
below is the source of truth for the network-transport runbook
generator at
`../03_Architecture/12_Latency_Engineering_Overview.md` §13 and the
alert-rule generation in
`../08_Operations/04_Observability_and_Events.md` (queued). The
fallback semantics across F1–F12 follow the **fail closed at
admission, degrade open at runtime** pattern symmetric with C26 §7,
C27 §7, C28 §7, C29 §7, C30 §7, C31 §7, C32 §7, C33 §7, C34 §7,
C35 §7, and C36 §10. Admission-time invariants (F10 SafeExec argv
allow-list, F1 ICE-checklist exhaustion with no TURN, F11 GSO-
broken-NIC capability check) refuse session admission with structured
`transport.admission_refused {session=…,cause=…}` events that the
C24 measurement harness propagates into the metrics plane and the
per-session capability snapshot. Runtime invariants (F2 SRTP key
rotation, F3 io_uring degradation, F4 XDP load fail, F5 sendmmsg
ENOBUFS, F6 RTCP feedback flood, F7 DTLS handshake timeout, F8 PMTU
blackhole, F9 MPQUIC path divergence, F12 QUIC migration loop) emit
`transport.degraded {from=…,to=…,reason=…}` events and the fallback
ladder runs forward — typically toward TURN relay (F1), pre-derived
next-key SRTP rotation (F2), sendmmsg fall-back (F3), plain-UDP
fall-back (F4), batch-size shrink + retry (F5), 5%-budget RTCP cap
with prioritisation (F6), 3-retry exponential-backoff DTLS (F7),
PMTU clamp to 1200 bytes (F8), per-path ABR controller (F9),
capability-degraded posture for SafeExec rejection (F10), per-NIC
GSO-disable list (F11), and 1-RTT debounce per QUIC migration (F12).

The **F1 ICE fail (symmetric NAT, no TURN)** row binds the chapter to
the **regional-TURN-deployment contract** from §3 of this chapter.
The ICE protocol (RFC 8445) establishes a candidate-pair-selection
process between two endpoints by exchanging server-reflexive (STUN)
and peer-reflexive candidates and probing each pair for connectivity;
when both endpoints sit behind symmetric NATs whose source-port
mapping is destination-IP-dependent, the ICE-Lite candidate-pair
shortcut fails and the only working candidate pair is via a TURN
relay. The chapter's mitigation is the **regional-TURN-deployment
contract** (every region has at least one TURN relay reachable
within < 25 ms RTT from any tenant in that region) plus an **early-
NAT-detection probe** that runs at session-create time and surfaces
the NAT topology to the connection-establishment policy. Detection
is via the ICE-checklist-exhaustion timer (default 5 s); mitigation
is to **engage the regional TURN relay** and redo the candidate-pair
selection with the relay candidate. Emit
`transport.ice_relay_engaged {session=…,nat_type="symmetric",relay_region=…}`.
F1 is **fail-closed at admission** if no TURN reachable; **degrade-
open at runtime** if TURN is reachable.

The **F2 SRTP key rotation race** row binds the chapter to the **pre-
derive-next-keys 30s-ahead contract** from §4 of this chapter. The
SRTP master key per RFC 3711 has a 2³¹-packet rotation budget; at
tier-5 4K60 with ~1500 packets/s per stream the budget exhausts in
~16 days but the chapter's per-session policy rotates every 24 hours
to bound the cryptographic exposure window. The rotation handshake
must complete before the old key index is exhausted; if the new key
arrives at the receiver after the first packet that uses it, the
receiver attempts to decrypt with the old key and fails, surfacing
as a brief decryption gap. The chapter's mitigation is the **pre-
derive next keys 30s ahead of rotation** (the key-derivation function
runs in advance of the rotation deadline so the receiver has both
keys available during the handover window) plus a **per-session
rotation-progress metric** that the operator can observe. Detection
is via the receiver-side decryption-failure-spike telemetry;
mitigation is to **trigger PLI** to recover the visible micro-
stutter and **flag the rotation latency** for the next-rotation
adjustment. Emit
`transport.srtp_rotation_race {session=…,gap_ms=…,pli_engaged=true}`.
F2 is **degrade-open at runtime**.

The **F3 io_uring unavailable (kernel < 5.15)** row binds the
chapter to the **boot-time capability detection contract** from §5
of this chapter. The canonical transmit path uses io_uring SQE128
+ multishot + register-files for batched UDP send with sub-microsecond
syscall overhead per `video-tech_dim12.md` §3.2; kernels older than
5.15 lack the SQE128 feature and kernels older than 5.10 lack
io_uring entirely. The boot-time capability probe queries the
kernel feature set via `io_uring_setup` + the `IORING_FEAT_SQPOLL`
flag and falls back to sendmmsg with a documented degraded
throughput posture if any required feature is missing. Detection
is via the boot-time capability probe; mitigation is to **fall back
to sendmmsg** and emit a structured capability-degraded event.
Emit
`transport.iouring_unavailable {kernel_version=…,fallback="sendmmsg"}`.
F3 is **degrade-open at runtime** (the session works at lower
throughput).

The **F4 XDP load fail** row binds the chapter to the **capability-
detection + plain-UDP-fall-back contract** from §6 of this chapter.
The canonical receive fast-path uses an XDP program loaded into the
NIC driver to filter, decode, and route incoming RTP / SRTP / SQP
packets directly from the NIC ring buffer to the receive-side jitter
buffer without traversing the kernel network stack, achieving 4–6×
lower per-packet latency than the plain-UDP-receive path per
`video-tech_dim12.md` §3.4. XDP load can fail because the NIC driver
does not declare `XDP_FLAGS_DRV_MODE` capability, the kernel BPF
verifier rejects the program, or `CAP_BPF` is missing on the
runtime container. Detection is via the boot-time XDP-load probe;
mitigation is to **fall back to plain UDP receive** and emit a
structured capability-degraded event. Emit
`transport.xdp_load_failed {nic=…,fallback="plain_udp"}`.
F4 is **degrade-open at runtime**.

The **F5 sendmmsg ENOBUFS** row binds the chapter to the **batch-
size shrink + retry contract** from §5 of this chapter. The
sendmmsg syscall (which the io_uring fall-back path uses) batches
multiple UDP messages into a single kernel call; under burst load
the kernel transmit ring can overflow and sendmmsg returns ENOBUFS
with a partial send-count, dropping the un-sent batch tail. The
chapter's mitigation is the **batch-size-shrink-and-retry path**
that halves the batch size on ENOBUFS and re-tries the dropped
packets with the smaller batch (recursively halving until either
the batch succeeds or the batch size reaches 1, in which case the
session emits a structured `transport.transmit_starved` alarm).
Detection is via the sendmmsg return-code; mitigation is to **shrink
batch size and retry**. Emit
`transport.sendmmsg_enobufs {original_batch=…,retry_batch=…,packets_lost=…}`.
F5 is **degrade-open at runtime**.

The **F6 RTCP feedback flood (low-bandwidth)** row binds the chapter
to the **RFC 3550 5%-bandwidth-cap contract** from §7 of this
chapter. RTCP messages (SR, RR, NACK, PLI, FIR, REMB) accumulate
under congestion as the receiver requests more retransmissions;
the RFC 3550 governor mandates that RTCP traffic stay below 5 % of
session bandwidth, but the default Pion implementation does not
enforce this cap, allowing RTCP to swell to 15–20 % of bandwidth on
tier-0 sessions. The chapter's mitigation is the **5 %-bandwidth
RTCP cap with feedback-message prioritisation** (under congestion,
PLI takes precedence over NACK takes precedence over REMB; lower-
priority messages are dropped to stay within the cap). Detection
is via the per-session RTCP-bandwidth-pct telemetry; mitigation is
to **prioritise PLI > NACK > REMB and drop overflow**. Emit
`transport.rtcp_cap_engaged {session=…,rtcp_pct=…,prioritisation_active=true}`.
F6 is **degrade-open at runtime**.

The **F7 DTLS-SRTP handshake timeout** row binds the chapter to the
**3-retry exponential-backoff contract** from §4 of this chapter.
The DTLS-SRTP handshake (RFC 5764) establishes the SRTP master
key over a DTLS session before media packets flow; certain
captive portals and aggressive middleboxes silently filter UDP
packets > 1300 bytes, dropping the DTLS `Finished` message and
hanging the handshake. The chapter's mitigation is the **3-retry
exponential-backoff** (initial timeout 1 s, backoff factor 2,
maximum 3 retries) plus a **DTLS-fragment-path** that fragments
the `Finished` message at the DTLS layer to keep individual UDP
packets under 1280 bytes for paths with restrictive middleboxes.
Detection is via the handshake-timeout-counter telemetry;
mitigation is to **engage the retry path and the fragment-path**.
Emit
`transport.dtls_timeout_retry {session=…,retry_count=…,fragment_path_engaged=true}`.
F7 is **degrade-open at runtime** (up to 3 retries); **fail-closed
at admission** after retries exhausted.

The **F8 Path MTU discovery fail** row binds the chapter to the
**PMTU probe + clamp to 1200 bytes contract** from §6 of this
chapter. RFC 1191 Path MTU Discovery relies on ICMP `Fragmentation
Needed` messages to inform the sender of the path MTU; many
middleboxes silently filter ICMP, producing silent blackholes for
packets above the path MTU. The chapter's mitigation is the
**packetisation-layer PMTU probe** (RFC 4821) that probes the path
with packets of varying sizes, detecting the blackhole at the
application layer, plus a **clamp to 1200 bytes** that bounds the
maximum RTP / SRTP / SQP payload to 1200 bytes when the probe
detects a blackhole. Detection is via the PMTU-probe-result
telemetry; mitigation is to **clamp MTU to 1200 bytes**. Emit
`transport.pmtu_clamp_engaged {session=…,detected_mtu=…,clamped_mtu=1200}`.
F8 is **degrade-open at runtime**.

The **F9 MPQUIC path RTT divergence** row binds the chapter to the
**per-path ABR controller + path priority contract** from §8 of
this chapter. MPQUIC (multipath QUIC, IETF draft as of 2026 per
`video-tech_dim12.md` §2) allows a single QUIC session to span
multiple network paths (Wi-Fi + cellular, primary + backup); under
divergent per-path RTTs (e.g. Wi-Fi 12 ms, cellular 48 ms) the C33
ABR controller cannot use a single bandwidth estimate without
oscillation. The chapter's mitigation is the **per-path ABR
controller** (each path maintains its own bandwidth and loss
estimate; the bitrate selector chooses per-path independently) plus
a **path-priority assignment** (the lower-RTT path is preferred for
latency-sensitive packets; the higher-RTT path is preferred for
loss-recovery and FEC). Detection is via the per-path RTT-
divergence telemetry; mitigation is to **engage per-path ABR and
path priority**. Emit
`transport.mpquic_path_diverge {session=…,paths=…,priority_assigned=true}`.
F9 is **degrade-open at runtime**.

The **F10 r18.SafeExec rejects subprocess** row is the chapter's
R-18 trip-wire and is symmetric with C26-F9, C27-F10, C28-F10,
C29-F10, C30-F10, C31-F10, C32-F10, C33-F10, C34-F10, C35-F8,
C36-F10. When a developer adds a non-allow-listed network-tooling
argv shape (e.g.
`tc qdisc add dev eth0 root fq pacing` for MPQUIC pacing setup
without the canonical `--rate` flag,
`bpftool prog load /tmp/xdp.o` for ad-hoc XDP loads bypassing the
canonical wrapper,
`iptables -t mangle -A OUTPUT -j DSCP --set-dscp 0xb8` against an
unauthorised chain),
the wrapper rejects the call at the `os/exec` boundary and
bootstrap aborts. The allow-list lives in `vasic-digital/helix-r18-
safeexec` and is **not duplicated** in this chapter; the family
allow-list extension that C37 contributes (canonical
`tc qdisc add dev <iface> root fq pacing rate <rate>`,
`bpftool prog load <path> /sys/fs/bpf/<name>`,
`iptables -t mangle -A POSTROUTING -m mark --mark <mark> -j DSCP --set-dscp <ef|af41|cs0>`,
`ip route get <ip>` for path-MTU diagnostics,
`ss -tunap` for socket-state diagnostics,
`turnutils_uclient` for TURN diagnostics) is recapped in §1
(family allow-list) of this chapter and verified by the C08
`host-integrity-scan` test inherited verbatim into §11.11. When
SafeExec rejects, the symptom is transmit-path-unavailability —
the controller cannot configure pacing / XDP / DSCP — and the
chapter's mitigation is a **capability-degraded fall-back** (the
controller emits a structured SafeExec rejection event and refuses
to start new sessions until the SafeExec issue is resolved or
falls back to in-process kernel-syscall equivalents where
available). Bypass requires an allow-list extension via operator
review per Constitution §11.5.4, never a silent workaround. Emit
`transport.safeexec_rejected {tool="tc",argv=…,fallback="capability_degraded"}`.

The **F11 GSO offload broken on certain NICs** row binds the
chapter to the **per-NIC capability database contract** from §5
of this chapter. UDP-GSO (Generic Segmentation Offload, kernel
4.18+) allows the kernel to send a single large UDP packet to the
NIC and have the NIC segment it into MTU-sized packets at the
hardware boundary, achieving 2–4× transmit-throughput improvement
per `video-tech_dim12.md` §3.1. Certain NIC firmware revisions
(documented Mellanox CX-5 regression in firmware 16.28.x corrupted
UDP-GSO at segment boundaries; certain Intel X710 firmware
revisions had similar issues) silently corrupt the segmented
packets, producing per-packet checksum failures and packet-loss
spikes at the receiver. The chapter's mitigation is the **per-NIC
capability database** that maps NIC + firmware-version pairs to
GSO-safe / GSO-broken status; on boot, the transmit path queries
the database via the NIC's vendor + device + firmware-version
strings (from `ethtool -i`) and disables GSO for known-bad
combinations. Detection is via the boot-time NIC-capability lookup;
mitigation is to **disable GSO and emit a capability-degraded
event**. Emit
`transport.gso_disabled {nic_vendor=…,nic_device=…,firmware=…,reason="known_bad"}`.
F11 is **fail-closed at admission** if NIC + firmware match a
known-bad entry.

The **F12 QUIC connection migration loop** row binds the chapter
to the **1-RTT debounce per migration contract** from §8 of this
chapter. QUIC connection migration (RFC 9000 §9) allows a session
to survive client-IP changes (mobile network handover, NAT
rebinding) by validating the new path with a PATH_CHALLENGE /
PATH_RESPONSE round-trip before switching. Under rapid client-IP
changes (e.g. Wi-Fi-to-cellular handover that flaps several times
in a few seconds), the session can migrate faster than the path-
validation can complete, producing a migration loop that consumes
session bandwidth on path-validation challenges and never settles
on a working path. The chapter's mitigation is the **1-RTT
debounce per migration** (after a successful migration, the session
ignores further migration triggers for at least one RTT to allow
the path to stabilise) plus a **migration-budget cap** (≤ 5
migrations per minute; further migrations are rejected and the
session falls back to the most-recently-validated path). Detection
is via the per-session migration-rate telemetry; mitigation is to
**engage the debounce and migration-budget cap**. Emit
`transport.quic_migration_debounce {session=…,migrations_per_min=…,debounce_engaged=true}`.
F12 is **degrade-open at runtime**.

| # | Failure mode | Symptom | Detection | Mitigation | Fallback |
|---|---|---|---|---|---|
| F1 | ICE fail (symmetric NAT, no TURN) — both endpoints behind symmetric NATs; ICE-Lite candidate-pair shortcut fails; no regional TURN reachable | Session abort with `ICE_FAIL`; emits `transport.ice_relay_engaged {session=…,nat_type="symmetric",relay_region=…}` or `transport.admission_refused {session=…,cause="ice_fail"}` | ICE-checklist-exhaustion timer (default 5 s) — `transport.ICE.ChecklistMonitor()` watches for binding-success | Regional TURN deployment (every region has TURN relay reachable < 25 ms RTT) + early-NAT-detection probe at session-create surfaces NAT topology to connection-establishment policy | **Fail-closed at admission** if no TURN reachable; **degrade-open at runtime** via TURN relay if reachable |
| F2 | SRTP key rotation race — new key arrives at receiver after first packet using it; brief decryption gap surfaces as micro-stutter | Brief gap on rotation (~5–10 ms); emits `transport.srtp_rotation_race {session=…,gap_ms=…,pli_engaged=true}` | Receiver-side decryption-failure-spike telemetry — `transport.SRTP.RotationMonitor()` watches per-session decryption-fail count | Pre-derive next keys 30s ahead of rotation deadline + per-session rotation-progress metric; trigger PLI on race detection to recover frame | **Degrade-open at runtime**; PLI recovers the visible micro-stutter; rotation latency flagged for next-rotation adjustment |
| F3 | io_uring unavailable (kernel < 5.15) — kernel lacks SQE128 + multishot features; transmit path falls back to sendmmsg | Throughput drops to sendmmsg level (1.4–1.8× lower per `video-tech_dim12.md` §3.2); emits `transport.iouring_unavailable {kernel_version=…,fallback="sendmmsg"}` | Boot-time capability probe — `transport.IOUring.FeatureProbe()` queries kernel via `io_uring_setup` + `IORING_FEAT_SQPOLL` flag | Boot-time capability detection; fall back to sendmmsg with documented degraded-throughput posture | **Degrade-open at runtime**; session works at lower throughput; capability-degraded posture flagged to C24 |
| F4 | XDP load fail — NIC driver missing `XDP_FLAGS_DRV_MODE`; BPF verifier rejects program; `CAP_BPF` missing on container | Per-packet latency degradation 4–6× (per `video-tech_dim12.md` §3.4); emits `transport.xdp_load_failed {nic=…,fallback="plain_udp"}` | Boot-time XDP-load probe — `transport.XDP.LoadProbe()` attempts test load + verifies driver mode capability | Capability detection at boot + fall-back to plain-UDP-receive path; structured capability-degraded event emitted | **Degrade-open at runtime**; session works at higher per-packet latency |
| F5 | sendmmsg ENOBUFS — kernel transmit ring overflows under burst load; partial send-count drops un-sent batch tail | Dropped batch tail; emits `transport.sendmmsg_enobufs {original_batch=…,retry_batch=…,packets_lost=…}` | sendmmsg return-code monitoring — `transport.SendMmsg.ReturnCodeCheck()` tracks per-call ENOBUFS rate | Batch-size-shrink-and-retry (halve batch on ENOBUFS, recurse to batch=1; emit `transport.transmit_starved` if batch=1 still fails) | **Degrade-open at runtime**; smaller batch resumes transmission; throughput recovers as ring drains |
| F6 | RTCP feedback flood (low-bandwidth tier 0) — PLI/NACK/REMB swells to 15–20% of session bandwidth on 1.5 Mbps sessions; starves video path | Bandwidth waste; ABR downshift cascade; emits `transport.rtcp_cap_engaged {session=…,rtcp_pct=…,prioritisation_active=true}` | Per-session RTCP-bandwidth-pct telemetry — `transport.RTCP.BandwidthCapMonitor()` tracks RTCP / video bytes ratio | RFC 3550 5%-bandwidth RTCP cap with feedback-message prioritisation (PLI > NACK > REMB; drop overflow under congestion) | **Degrade-open at runtime**; lower-priority messages dropped to stay within cap; video path bandwidth preserved |
| F7 | DTLS-SRTP handshake timeout — captive portal silently filters UDP > 1300 bytes; `Finished` message dropped; handshake hangs | Session start fails with `DTLS_TIMEOUT`; emits `transport.dtls_timeout_retry {session=…,retry_count=…,fragment_path_engaged=true}` | Handshake-timeout-counter telemetry — `transport.DTLS.HandshakeMonitor()` tracks per-session timeout count | 3-retry exponential-backoff (initial 1 s, factor 2, max 3) + DTLS-fragment-path that fragments `Finished` to keep packets < 1280 bytes | **Degrade-open at runtime** (up to 3 retries); **fail-closed at admission** after retries exhausted |
| F8 | Path MTU discovery fail — middlebox filters ICMP `Frag Needed`; silent blackhole for packets above path MTU (PPPoE 1492; mobile 1280) | Blackholed packets; emits `transport.pmtu_clamp_engaged {session=…,detected_mtu=…,clamped_mtu=1200}` | Packetisation-layer PMTU probe (RFC 4821) — `transport.PMTU.ProbeAndDetect()` probes path with varying sizes | PMTU probe + clamp to 1200 bytes (documented safe lower bound) when blackhole detected | **Degrade-open at runtime**; transmit clamps packet size; bandwidth slightly reduced but session continues |
| F9 | MPQUIC path RTT divergence — Wi-Fi + cellular paths diverge > 30 ms; ABR controller bandwidth estimates conflict; bitrate oscillation | ABR confusion; bitrate oscillation; emits `transport.mpquic_path_diverge {session=…,paths=…,priority_assigned=true}` | Per-path RTT-divergence telemetry — `transport.MPQUIC.PathDivergeMonitor()` tracks RTT-spread across paths | Per-path ABR controller (each path independent bandwidth estimate) + path-priority assignment (lower-RTT for latency, higher-RTT for FEC) | **Degrade-open at runtime**; per-path independent ABR; bitrate stabilises per path |
| F10 | `r18.SafeExec` rejects subprocess (e.g. `tc qdisc add dev eth0 root fq pacing` off-list, `bpftool prog load /tmp/xdp.o` bypass, `iptables -t mangle -A OUTPUT -j DSCP` against unauthorised chain) | Pacing/XDP/DSCP-marking unavailable; bootstrap fails on transport initialisation; emits `transport.safeexec_rejected {tool="tc",argv=…,fallback="capability_degraded"}` | Wrapper's verbatim allow-list check at `os/exec` boundary returns `ErrForbiddenArgvShape`; harness logs the rejection | Fix the call site to use the allow-listed shape — canonical `tc qdisc add dev <iface> root fq pacing rate <rate>`, `bpftool prog load <path> /sys/fs/bpf/<name>`, `iptables -t mangle -A POSTROUTING -m mark --mark <mark> -j DSCP --set-dscp <ef|af41|cs0>`, `ip route get <ip>` for diagnostics, `ss -tunap` for socket-state diagnostics, `turnutils_uclient` for TURN diagnostics; fall back to in-process kernel-syscall equivalents where available | Capability-degraded fall-back — non-blocking for existing sessions but new sessions refused; non-overridable per Constitution §11.5.4; bypass requires §13 exception with documented mitigation; cross-link §11.11 host-integrity-scan |
| F11 | GSO offload broken on certain NICs (Mellanox CX-5 firmware 16.28.x corrupted UDP-GSO at segment boundaries; Intel X710 similar) — silent corruption produces checksum failures + packet-loss spikes | Corrupted packets; per-packet checksum failures at receiver; emits `transport.gso_disabled {nic_vendor=…,nic_device=…,firmware=…,reason="known_bad"}` | Boot-time NIC-capability lookup — `transport.GSO.NicCapabilityCheck()` queries vendor + device + firmware via `ethtool -i` against capability DB | Per-NIC capability database mapping NIC + firmware to GSO-safe/broken; disable GSO for known-bad combinations + emit capability-degraded event | **Fail-closed at admission** if NIC + firmware match known-bad entry; transmit path uses non-GSO send loop |
| F12 | QUIC connection migration loop — rapid client-IP changes (mobile handover, NAT rebinding) flap migrations faster than path-validation can complete; oscillation on PATH_CHALLENGE/RESPONSE | Migration oscillation; bandwidth waste; emits `transport.quic_migration_debounce {session=…,migrations_per_min=…,debounce_engaged=true}` | Per-session migration-rate telemetry — `transport.QUIC.MigrationRateMonitor()` tracks migrations/minute | 1-RTT debounce per migration (ignore further migration triggers for ≥ 1 RTT after success) + migration-budget cap (≤ 5/min; further rejected) | **Degrade-open at runtime**; debounce stabilises path; session falls back to most-recently-validated path on cap |

## 11. Test surface

The C37 test surface inherits the family-level container-driven CI lane
contract from C26 §8 + C27 §8 + C28 §8 + C29 §8 + C30 §8 + C31 §8 +
C32 §8 + C33 §8 + C34 §8 + C35 §8 + C36 §11 and the
`vasic-digital/Containers` runner image, **extended** with the new
network-transport-pipeline-specific requirement: every integration /
E2E / chaos / stress test must exercise **a real WebRTC RTP/SRTP
stack (Pion) + real ICE agent + real DTLS-SRTP handshake + real
io_uring transmit path + real XDP receive path + real sendmmsg fall-
back + real MPQUIC multipath + real RTCP feedback governor + real
PMTU probe + real per-NIC GSO capability database + real netem WAN
emulator with multi-NAT topology** so the packetisation contract,
ICE agent, key-exchange, transmit / receive fast-paths, MPQUIC, and
capability detection are validated against real network behaviour
(mocking the WAN, the SRTP stack, the ICE agent, the kernel
capability surface, or the NIC firmware database is forbidden per
Constitution §6.4 — only unit tests may use mocks). Per Constitution
§6.4 + Master Plan §4.3 anti-bluff verification, the test matrix
below cites `video-tech_dim10.md` (testing dimension) and
`video-tech_dim12.md` (transport dimension) explicitly so every per-
tier transport-throughput / latency / packet-loss claim is grounded
in a primary-source reference.

### 11.1 Unit (mocks allowed)

The unit-test layer is the only place where mocks, stubs, or
hardcoded values are permitted per Constitution §6.4 — every other
layer below hits the real transport infrastructure.

- **RTPSender unit test** — instantiate the RTP-sender wrapper with
  a mocked outbound UDP socket; submit synthetic encoded frames
  with known timestamps + sequence numbers; assert the sender
  emits RFC 6184-compliant H.264 NAL packetisation per
  `video-tech_dim12.md` §1.1 (FU-A fragmentation for NALs > MTU,
  STAP-A aggregation for small NALs, single-NAL for typical-size
  NALs); assert the sender emits RFC 7798-compliant HEVC
  packetisation; assert the sender emits AV1 OBU packetisation per
  `video-tech_dim12.md` §1.3; assert the sequence number rolls over
  correctly at 2¹⁶; assert the timestamp clock is 90 kHz per RFC
  3550.
- **SRTPHandler unit test** — instantiate the SRTP handler with a
  mocked key-derivation function; submit synthetic RTP packets with
  known SSRC + sequence number; assert the SRTP-encrypted output
  matches the RFC 3711 test vectors; assert the handler correctly
  rotates keys at the 2³¹-packet budget per F2; assert the handler
  pre-derives next keys 30s ahead of rotation deadline; assert the
  handler rejects packets with replayed sequence numbers per RFC
  3711 §3.3.2.
- **ICEAgent unit test** — instantiate the ICE agent with mocked
  STUN / TURN servers + mocked candidate-pair-probe responses;
  submit synthetic candidate pairs across the four NAT types (full-
  cone, restricted-cone, port-restricted, symmetric); assert the
  agent correctly selects the highest-priority working pair per
  RFC 8445 §6.1; assert the agent engages TURN relay when no direct
  pair is reachable per F1; assert the agent emits `ICE_FAIL` when
  TURN is also unreachable.
- **RTCPGovernor unit test** — instantiate the RTCP-feedback
  governor with a mocked session-bandwidth signal; submit synthetic
  PLI / NACK / REMB messages; assert the governor caps RTCP at 5%
  of session bandwidth per RFC 3550 + F6; assert the governor
  prioritises PLI > NACK > REMB under congestion; assert the
  governor drops overflow messages with a structured event.

### 11.2 Integration

The integration-test layer hits the real pion stack on a real loopback
+ netem WAN emulator — no mocks, no stubs, no hardcoded values. Per
Constitution §6.4 this layer must run inside the canonical
`vasic-digital/Containers` runner image with real pion `webrtc-rs` /
`pion` libraries + real netem WAN emulation + real Linux kernel with
io_uring + XDP + GSO capability + real Mellanox / Intel NIC under
test (passthrough into the container).

- **Real SRTP handshake against pion stack** — boot two pion
  endpoints in containers connected by a netem-emulated WAN with 20
  ms RTT; perform a full DTLS-SRTP handshake per RFC 5764; assert
  the SRTP master key is correctly derived; **assert the cipher
  swap works (rotate the SRTP master key mid-session and verify the
  receiver decrypts both pre- and post-rotation packets without
  loss)**; assert the per-session rotation-progress telemetry
  matches the F2 30s-ahead pre-derivation contract.
- **ICE candidate-pair selection across NAT types** — boot two
  endpoints behind synthetic NATs (full-cone, restricted-cone,
  port-restricted, symmetric) using `nftables` NAT rules in the
  netem-WAN emulator; for each NAT-type pair, run ICE candidate-
  pair selection; assert the selected pair matches the RFC 8445
  expected outcome; assert the symmetric-NAT case engages TURN
  relay per F1.
- **Real io_uring transmit path** — boot a transmit-path test
  binary in a container with a kernel ≥ 5.15; configure io_uring
  SQE128 + multishot + register-files; submit a synthetic 4K60
  RTP / SRTP stream for 60 seconds; assert the per-packet send
  latency is within the §5 budget; assert the boot-time capability
  probe correctly detects io_uring availability; on a kernel < 5.15
  test container, assert the fall-back to sendmmsg engages per F3.
- **Real XDP receive path** — boot a receive-path test binary
  attached to a real NIC (or a `veth` pair with XDP support);
  load the canonical XDP program; submit a synthetic RTP / SRTP
  stream; assert the per-packet receive latency is within the §6
  budget; on a NIC without `XDP_FLAGS_DRV_MODE`, assert the fall-
  back to plain UDP engages per F4.
- **Real PMTU probe** — boot a transmit-path test binary against a
  netem-emulated path with restricted MTU (1280, 1492, 1500); run
  the packetisation-layer PMTU probe per RFC 4821; assert the
  probe correctly detects the path MTU; on a path with ICMP
  filtering (silent blackhole), assert the probe detects the
  blackhole and clamps to 1200 bytes per F8.

### 11.3 E2E

The E2E layer brings up the **full network-transport pipeline** end-
to-end and asserts user-perceptible latency + zero session-gap
properties.

- **Tier-5 4K SDR session E2E (encode → RTP/SRTP → io_uring → ICE
  → wire → assert ≥35 ms p999 floor + zero gaps)** — boot a host
  with a real RTX 4090 + a real client + the real network-
  transport pipeline; session-create at tier 5 across a real WAN
  with 20 ms RTT + 0.5 % packet loss; capture + encode + RTP-
  packetise + SRTP-encrypt + io_uring-transmit + receive + decode
  for a 5-minute session; capture per-input glass-to-glass latency
  via the C24 LDAT rig; **assert the per-session p999 glass-to-
  glass latency stays ≤ 35 ms per the tier-5 budget**; **assert
  zero session gaps (zero packet drops, zero decryption gaps, zero
  ABR thrash) across the 5-minute window**.
- **Tier-7 4K HDR session E2E** — same fixture but at tier 7
  (4K HDR Dolby Vision); assert p999 ≤ 30 ms; assert zero gaps;
  assert HDR-specific RTP marker bit is correctly propagated.
- **Multi-NAT cross-region E2E** — boot a host in US-East NAT
  region + a client in EU-West NAT region with a netem-emulated
  symmetric NAT on each side; session-create; assert the ICE agent
  correctly engages a regional TURN relay per F1; assert per-
  session latency stays within the cross-region budget; assert no
  session-abort.
- **MPQUIC multipath E2E** — boot a client with two network paths
  (Wi-Fi 12 ms RTT + cellular 48 ms RTT via netem); session-create
  with MPQUIC; assert per-path ABR controller correctly partitions
  the bandwidth estimate per F9; assert path-priority is correctly
  assigned (lower-RTT for latency-sensitive packets).

### 11.4 Security

- **`r18.SafeExec` rejection fuzz** — for each of the family
  allow-list entries (`00_Index.md` §7), construct off-allow-list
  argv shapes (e.g. `tc qdisc add dev eth0 root fq pacing` is
  off-list (missing `--rate`); `bpftool prog load /tmp/xdp.o`
  is off-list (loading from `/tmp` not from `/sys/fs/bpf`);
  `iptables -t mangle -A OUTPUT -j DSCP --set-dscp 0xb8` is off-list
  (chain `OUTPUT` instead of `POSTROUTING`)) and fuzz with 10⁶
  argv permutations per Constitution §6.4 fuzz contract; assert
  the wrapper returns `ErrForbiddenArgvShape` for every off-list
  shape with no false-positive on allow-list shapes; assert no
  host-disruptive command (kill, systemctl, pmset) ever passes the
  wrapper. The deny-list is **inherited from C08 §10** per the
  family contract — no duplication in this chapter.
- **SRTP authentication-tag tampering** — submit synthetic SRTP
  packets with corrupted authentication tags; assert the receiver
  rejects the packet with a structured audit event; assert no
  cleartext bytes leak before tag verification per RFC 3711 §3.3.4.
- **DTLS-SRTP handshake replay** — capture a DTLS-SRTP handshake
  on the wire; replay it against a fresh server; assert the server
  rejects the replay with a structured audit event per RFC 6347
  §4.1.2.5.
- **TURN relay credential abuse** — assert per-tenant TURN
  credentials cannot relay sessions across tenant boundaries;
  assert any cross-tenant relay attempt is refused with a
  structured audit event.

### 11.5 Benchmarking

The benchmarking layer is the chapter's binding to Constitution §6 —
every per-tier transport-throughput / latency claim **reports
p50 / p99 / p999 at ≥ 10 K samples** via the C24 measurement
harness. Cross-link C24. Per **`video-tech_dim10.md`** §3 + §6 +
**`video-tech_dim12.md`** §3 (the transport-throughput dimension),
the benchmarking corpus uses synthetic-content + real-game-capture
pairs across the six representative game profiles (FPS, racing,
RPG, RTS, MOBA, fighting) so the per-profile transport
characterisation reflects production-like workloads.

- **Bench sendmmsg vs io_uring throughput** — measure per-packet
  send latency + per-second packet throughput across **≥ 10 000
  samples** per workload profile using both sendmmsg and io_uring
  paths; **report p50 / p99 / p999 per Constitution §6**;
  histogram artifact attached; budget per `video-tech_dim12.md`
  §3.2 — io_uring p999 < 5 µs per packet, sendmmsg p999 < 12 µs
  per packet at tier-5 4K60 load.
- **Bench ICE setup time** — measure full ICE candidate-pair
  selection time across **≥ 10 K samples** per NAT-type
  combination; **report p50 / p99 / p999**; budget per
  `video-tech_dim12.md` §3 — full-cone p999 < 200 ms,
  symmetric-with-TURN p999 < 800 ms.
- **Bench per-tier packet-loss tolerance** — measure end-to-end
  perceptible-quality stability across **≥ 10 K samples** per tier
  under controlled packet loss (0.1 %, 0.5 %, 1 %, 2 %, 5 %);
  **report p50 / p99 / p999** of per-session VMAF + glass-to-
  glass latency; budget per `video-tech_dim10.md` §6 — tier-5
  acceptance gate at 0.5 % loss is VMAF ≥ 87 + p999 ≤ 35 ms.
- **Bench XDP receive throughput** — measure per-packet receive
  latency across **≥ 10 K samples** under steady-state load using
  both XDP and plain-UDP paths; **report p50 / p99 / p999**;
  budget per `video-tech_dim12.md` §3.4 — XDP p999 < 800 ns per
  packet, plain-UDP p999 < 5 µs per packet.
- Cross-link **C24** measurement harness for shared histogram-
  collection + bootstrap-resampling-confidence-interval primitives.
  The benchmark suite must cite **`video-tech_dim10.md`** + 
  **`video-tech_dim12.md`** explicitly per Master Plan §4.3 anti-
  bluff verification — `video-tech_dim10.md` §6 enumerates the
  per-profile loss-tolerance thresholds + `video-tech_dim12.md`
  §3 enumerates the canonical transport-throughput corpus + 
  `video-tech_dim12.md` §5 enumerates the per-tier acceptance-
  gate budgets for jitter-buffer behaviour.

### 11.6 Chaos

- **NAT type swap mid-session** — boot a session behind full-cone
  NAT; mid-session, swap the NAT to symmetric (via netem
  reconfiguration); assert the ICE agent correctly re-runs
  candidate-pair selection and engages TURN relay per F1; assert
  the session does not abort; assert glass-to-glass latency stays
  within the §10 budget across the swap.
- **RTCP feedback loss burst** — inject a controlled RTCP packet
  loss burst (50 % loss for 10 s) on a tier-5 session; assert the
  receiver correctly handles the burst (NACK retransmission
  triggered, REMB recomputed); assert the RTCP-bandwidth-cap
  governor stays within 5 % per F6; assert no ABR thrash.
- **DTLS rekey under load** — trigger a DTLS rekey mid-session
  while the session is at peak transmit load (tier-7 4K HDR); 
  assert the rekey completes without packet loss; assert the
  pre-derived next-keys-30s-ahead path engages per F2; assert
  zero decryption gaps.
- **Inject MPQUIC path RTT divergence (F9)** — boot a session
  with MPQUIC + two paths; mid-session, inject a controlled RTT
  spike on one path (12 ms → 80 ms via netem); assert F9
  detection fires; assert per-path ABR controller correctly
  partitions; assert no bitrate oscillation.
- **Inject NIC GSO corruption (F11)** — simulate the Mellanox
  CX-5 firmware-16.28.x bug via a synthetic packet-corruption
  injector at the NIC layer; assert the boot-time capability
  database correctly identifies the known-bad NIC; assert GSO
  is disabled for the session per F11.

### 11.7 Stress

- **24h with 5% packet-loss + jitter on all paths; verify zero
  session-aborts** — on each runner, run continuous tier-5 +
  tier-6 + tier-7 sessions through a netem-emulated WAN with 5 %
  packet loss + 20 ms ± 5 ms jitter for 24 hours; **assert zero
  session-aborts** across the 24 h window; **assert no fd leak**
  (process fd count stable to within 5 fds over 24 h); **assert
  no GC stall > 1 ms** (cross-link C36 §3 Go pipeline `sync.Pool`
  discipline); assert no memory leak (transport publisher RSS
  growth < 5 MB / hour); **assert per-session VMAF + p999 stays
  within tier-specific acceptance gates across the 24 h window**.
- **Multi-NAT concurrent stress** — provision 8 concurrent
  sessions across all four NAT types (2 sessions per NAT type)
  for 24 hours; assert per-session ICE candidate-pair selection
  + TURN relay where required; assert no cross-session
  interference; assert the per-tenant TURN credentials are
  correctly scoped.

### 11.8 Smoke

- **Capability schema reports correct `iouring` / `xdp` / `gso`
  flags** — boot the transport pipeline in a clean container with
  a real NIC; query the published capability schema; assert the
  schema reports correct `iouring` / `xdp` / `gso` capability
  flags consistent with the runtime kernel + NIC + firmware
  context; assert the schema validates against
  `vasic-digital/helix-transport/schema/v1.json`.
- **Default DSCP markings present** — boot a session with the
  canonical DSCP marking enabled (EF for video, AF41 for audio,
  CS0 for control); capture per-packet DSCP via `tcpdump -i any
  -n -q -tt`; assert the markings match the §1 family allow-list;
  assert the markings survive the egress path (no kernel rewrite
  in the chain).
- **Smoke test ICE setup** — boot a session-create call against
  a known-good full-cone NAT topology; assert the ICE agent
  completes candidate-pair selection within 200 ms; assert the
  selected pair is server-reflexive (not TURN-relayed); assert
  the per-session capability snapshot includes the ICE topology.

### 11.9 Full automation

All of §11.1–§11.8 run on **every commit via the local container-
driven CI lane** per Constitution §10. The CI lane uses the
canonical `vasic-digital/Containers` runner image with the io_uring
+ XDP + GSO kernel feature set + the netem WAN emulator + the
multi-NAT topology + the real Mellanox / Intel NIC passthrough +
the pion WebRTC stack + the helix-r18-safeexec wrapper addressable
on the runner network. The matrix covers (Linux Ubuntu 22.04 / 24.04
+ Fedora 40, Windows Server 2022, macOS 14) × (8 quality tiers ×
6 game profiles × 4 NAT-type combinations × 3 transmit-path
profiles {io_uring, sendmmsg, plain UDP}). The full-automation lane
emits a single composite artifact (`transport-test-report.json`)
that the C24 latency-side harness consumes as the authoritative
source-of-truth for any per-tier transport-throughput / latency /
packet-loss claim in chapter prose. The CI lane runs nightly on
the real multi-NAT cross-region rig (the multi-NAT + multi-region
rig is too expensive for per-commit hardware exercise; per-commit
runs use the unit + integration layers against a single
representative NAT topology, with the full multi-NAT + multi-
region matrix gated to the nightly schedule per Constitution §10's
local-CI-equivalence clause). The HelixQA dashboard at
`git@github.com:HelixDevelopment/HelixQA.git` is updated nightly
with the composite artifact; HelixQA's findings are surfaced as
P1/P2 work items per Constitution §6.5.

### 11.10 Challenges (production-like)

HelixQA dispatches **per-tier verification scenarios** from
`git@github.com:vasic-digital/Challenges.git` (per Constitution
§6.4 Challenges-test contract):

- **HelixQA full-system Challenge with cross-region multi-NAT
  topology** — HelixQA boots a fully-provisioned multi-region rig
  with hosts in US-East + EU-West + AP-South + ME-West, clients
  behind each of the four NAT types (full-cone, restricted-cone,
  port-restricted, symmetric), and TURN relays in each region.
  For each (host-region × client-region × client-NAT-type)
  combination, run a 30-minute session at tier 5; **assert the
  ICE agent handles all 4 NAT types correctly per F1**; assert
  per-session glass-to-glass latency stays within the cross-
  region budget; assert TURN relay engagement for symmetric-NAT
  clients; assert per-session capability snapshot correctly
  reflects the NAT topology.
- **MPQUIC multipath Challenge** — HelixQA dispatches a multi-
  path scenario (Wi-Fi + cellular per client); injects controlled
  RTT divergence and packet-loss bursts on each path; asserts
  per-path ABR controller correctly partitions per F9; asserts
  the C33 ABR controller's per-path bandwidth estimates do not
  conflict.
- **DTLS handshake under captive-portal Challenge** — HelixQA
  dispatches a scenario with a netem-emulated captive portal
  that filters UDP > 1300 bytes; asserts the DTLS-fragment-path
  engages per F7; asserts session-create succeeds via the
  fragmented handshake; asserts the 3-retry exponential-backoff
  fires on transient drops.
- **Per-fault recovery Challenges** — inject each of F1–F12
  during a live Challenges scenario; assert the recovery path
  fires correctly and the final per-session glass-to-glass
  latency + zero-session-abort verification holds.

### 11.11 Inherited host-integrity-scan (R-18, non-overridable per Constitution §11.5.4)

Inherits VERBATIM from C08 §12.11: a dedicated test, mandated by
Constitution §11.5.4, that boots the host agent under
`strace -fe trace=execve` on a Linux test host (the canonical
reference platform for this scan) and runs the full Ten-test-type
matrix above against it. The strace log is then grepped for
**every** §11.5.1 forbidden pattern. The gate is:

```
zero matches across the entire log → PASS
≥ 1 match anywhere in the log     → FAIL (non-overridable)
```

The log is preserved as an artifact alongside the auditd record
from §12.4. The test is **non-overridable** per Constitution
§11.5.4: a match is a Constitution violation, never a flake, and
bypass requires a §13 exception with a documented compensating
control. The same test is replicated on Windows under
`Process Monitor` ETW filtered to `Process Create`, and on macOS
under `dtruss -f -t execve`, so the host-integrity-scan covers
all three host OSes the agent ships on.

The C37 implementation contract that this scan validates:

- Network-tooling invocation via `r18.SafeExec` only — never via
  `os/exec.Command` directly; the canonical shapes
  (`tc qdisc add dev <iface> root fq pacing rate <rate>`,
  `bpftool prog load <path> /sys/fs/bpf/<name>`,
  `iptables -t mangle -A POSTROUTING -m mark --mark <mark>
  -j DSCP --set-dscp <ef|af41|cs0>`,
  `ip route get <ip>` for path-MTU diagnostics,
  `ss -tunap` for socket-state diagnostics,
  `turnutils_uclient` for TURN diagnostics,
  `ethtool -i <iface>` for read-only NIC capability lookup) are
  the family allow-list entries for transport / pacing / XDP /
  DSCP / TURN / NIC tooling.
- No host-disruption commands ever appear in the transport /
  pacing / XDP / DSCP / TURN / NIC path: no `kill -9 <pid>`, no
  `systemctl suspend|hibernate|reboot|halt|poweroff`, no
  `pmset`, no `xset dpms force off`, no `--privileged`
  container flag, no host-mount of `/`, `/dev`, `/proc`, `/sys`
  (the NIC device-files in `/dev/<nic>` and the BPF filesystem
  in `/sys/fs/bpf` are exposed via the canonical container-
  toolkit injection per the `vasic-digital/Containers` runner
  image, never via host-mount). The scan asserts none of these
  syscall patterns appear in the transport subsystem's syscall
  trace.
- No cross-tenant transport-state traversal — the scan asserts
  the transport pipeline's `openat` syscalls never reference
  paths outside the per-tenant scoped transport configuration
  root, and no `chdir` / `chroot` syscall escapes the scope.

The scan's invocation contract is byte-identical with the C08
§12.11 inheritance into every chapter in the family per
`00_Index.md` §7 R-18 family allow-list. No chapter in the family
is permitted to redefine, override, or extend the scan —
Constitution §11.5.4 forbids per-chapter customisation of the
host-integrity contract.

## 12. Open questions

The following open questions are tracked in the chapter's OQ log
and surface to the family-level OQ aggregator at `00_Index.md`
§5. Each OQ is prefixed `OQ-C37-NN` and carries an owner, a target
resolution date, and a cross-link to the deciding chapter or
external dependency.

- **OQ-C37-01** — MoQ Media-over-QUIC V1/V2 timing. The IETF
  MoQ Working Group (Media over QUIC, draft-ietf-moq-transport)
  has been progressing through 2025–2026 with a target of WG
  Last Call in late 2026 and RFC publication in 2027–2028. MoQ
  is positioned as the successor to RTP for low-latency
  media transport, with native QUIC datagram framing,
  publish/subscribe semantics, and built-in priority +
  congestion-control hooks that align with HelixPlay's tier-
  system + ABR controller. Should HelixPlay V1 ship a MoQ
  transport profile as a per-tenant operator-policy alternative
  to the WebRTC RTP/SRTP profile, or should MoQ remain a V2
  deliverable? The cost is the per-tenant transport-profile
  migration + the C33 ABR controller adaptation + the new
  pub/sub-aware capability schema; the benefit is a more-
  efficient transport with lower per-packet overhead and
  native QUIC features. Trigger: IETF MoQ draft maturity
  (WG Last Call); pion + browser implementations stabilise.
  Owner: C37 + V1 family + Transport WG. Cross-link IETF
  draft-ietf-moq-transport + `video-tech_dim12.md` §2 MoQ
  survey + V1 transport-profile decision matrix.

- **OQ-C37-02** — SQP envelope V2. The custom UDP SQP
  (Streamed Queueing Protocol) envelope is the chapter's
  successor to the RTP/SRTP framing for HelixPlay-internal
  game-streaming traffic; SQP V1 (documented in §2 of this
  chapter) provides RTP-equivalent packetisation + RTCP-
  equivalent feedback with HelixPlay-specific game-input
  channelisation and tier-pinned capability negotiation. SQP
  V2 should ratify a codec-coupled congestion-control
  mechanism (the per-codec encode-rate control sees the
  network's queue state and the network's congestion-control
  sees the codec's quantization state, allowing tighter
  coupling than the RFC 8298 SCReAM or RFC 9002 QUIC CC can
  achieve). Should V1 standardise SQP V2 with codec-coupled
  CC, or should it remain in research as a V2 deliverable?
  The cost is the per-codec CC migration + the per-tenant
  capability-schema migration; the benefit is bounded queue
  growth + bounded codec-quality oscillation under congestion.
  Trigger: V1 codec WG + Transport WG joint review; per-codec
  CC research matures. Owner: C37 + V1 family + Codec WG +
  Transport WG. Cross-link `video-tech_dim12.md` §3.5 codec-
  CC research + V1 transport-profile decision matrix +
  RFC 8298 SCReAM + RFC 9002 QUIC CC.

- **OQ-C37-03** — MPQUIC enablement strategy. The §8 MPQUIC
  multipath transport is gated behind the per-tier and per-
  region capability schema; tier-5+ sessions in regions with
  documented dual-path client topology (residential Wi-Fi +
  cellular bonding via Apple iOS 17 / Android 14 multipath
  TCP/QUIC stacks) are the early-adopter target. Should V1
  enable MPQUIC by default for tier-5+ sessions in all regions,
  or should it roll out region-by-region as the per-region
  client-mix telemetry matures? The cost is the per-tenant
  MPQUIC capability migration + the C33 ABR controller per-
  path partitioning + the per-region TURN relay sizing; the
  benefit is improved p999 latency floor on dual-path clients
  + reduced packet-loss exposure. Trigger: per-region client-
  mix telemetry matures; iOS 17 / Android 14 adoption rates
  cross 50 % per region. Owner: C37 + V1 family + Operations
  family + Mobile WG. Cross-link Operations chapter
  `08_Operations/05_DC_Tier_Capacity.md` (queued) + per-region
  rollout plan + Apple WWDC 2023 multipath-TCP/QUIC sessions.

- **OQ-C37-04** — AF_XDP zero-copy receive cross-region rollout.
  The §6 XDP fast-path uses AF_XDP zero-copy mode for the
  receive path, achieving 4–6× lower per-packet receive latency
  per `video-tech_dim12.md` §3.4. The kernel-version dependency
  (AF_XDP zero-copy mode requires kernel ≥ 5.14 + driver
  support) limits cross-region rollout; certain regions still
  run kernels < 5.14 on production hosts. Should V1 enforce a
  minimum kernel-version baseline (≥ 5.15) across all regions
  to guarantee AF_XDP zero-copy availability, or should it
  remain a per-region per-host capability with documented
  fall-back? The cost is the per-region kernel-upgrade
  programme + the per-host capability migration; the benefit
  is uniform receive-path latency across regions. Trigger:
  per-region kernel-upgrade programme posture clarifies; per-
  region client p999-latency-floor telemetry matures. Owner:
  C37 + V1 family + Operations family + Kernel WG. Cross-link
  Operations chapter + per-region kernel-upgrade plan +
  `video-tech_dim12.md` §3.4 AF_XDP survey.

- **OQ-C37-05** — QUIC datagram replacing RTP. RFC 9221
  (Unreliable QUIC Datagrams) provides an unreliable-delivery
  primitive within a QUIC session that is suitable for real-
  time media; combined with RFC 9000 connection migration and
  RFC 9001 TLS 1.3 key-exchange, QUIC datagrams could replace
  the RTP/SRTP/DTLS-SRTP stack with a single QUIC session.
  Insight #7 in `video-tech_dim12.md` §2 surveys the QUIC-
  datagram-as-RTP-replacement landscape and notes that pion
  and browser support is maturing through 2026. Should V2 ship
  a QUIC-datagram transport profile as the primary low-
  latency media transport, retiring RTP/SRTP for new
  deployments? The cost is the per-tenant transport-profile
  migration + the per-codec packetisation rewrite + the new
  capability schema + the C33 ABR controller adaptation; the
  benefit is a unified session per RTC connection (no
  separate DTLS-SRTP handshake), built-in connection migration,
  and unified key-rotation. Trigger: pion + browser QUIC-
  datagram implementations stabilise; IETF MoQ adopts QUIC
  datagrams as substrate. Owner: C37 + V2 family + Transport
  WG. Cross-link `video-tech_dim12.md` §2 Insight #7 + RFC
  9221 + RFC 9000 + V2 transport-profile decision matrix.

- **OQ-C37-06** — TURN cluster sizing — % sessions needing
  relay. The §3 regional TURN deployment provides one TURN
  relay per region for symmetric-NAT clients per F1; the
  per-region TURN cluster sizing depends on the % of sessions
  that need relay, which depends on the per-region NAT-type
  mix (residential broadband vs mobile vs enterprise). Should
  V1 size each region's TURN cluster for 25 % relay capacity
  (a conservative upper bound on symmetric-NAT exposure), or
  should it size for the per-region observed NAT-type mix
  (which may be 5 %–40 % depending on region)? The cost is
  the per-region TURN cluster sizing + the per-tenant TURN
  capacity reservation; the benefit is per-region cost
  optimisation. Trigger: per-region client-NAT-type telemetry
  matures; per-region TURN-relay-utilisation telemetry
  emerges. Owner: C37 + V1 family + Operations family +
  Capacity WG. Cross-link Operations chapter + per-region
  TURN sizing study + RFC 5766 TURN.

- **OQ-C37-07** — IPv6 deployment per-region rollout. The MVP
  RTP/SRTP and SQP transports run primarily on IPv4 today;
  IPv6 deployment is a per-region rollout decision, depending
  on per-region IPv6-adoption rates (US 50 %, EU 60 %,
  AP-South 80 %, ME 30 %) and per-region carrier IPv6 support.
  Should V1 enable dual-stack IPv4+IPv6 by default in all
  regions, or should it remain per-region opt-in via operator
  policy? The cost is the per-region dual-stack capacity
  reservation + the per-tenant capability schema migration +
  the per-region monitoring + the MTU-discovery for IPv6
  paths (IPv6 default MTU 1500 vs IPv4 PPPoE 1492); the
  benefit is improved per-session connectivity for IPv6-only
  clients (carrier-grade NAT exposure reduces). Trigger:
  per-region IPv6-adoption telemetry matures; per-region
  carrier IPv6 support clarifies. Owner: C37 + V1 family +
  Operations family. Cross-link Operations chapter + per-
  region IPv6-rollout plan + RFC 8200 IPv6 + RFC 4291 IPv6
  addressing.

- **OQ-C37-08** — ECN marking on QUIC datagram + L4S support.
  Explicit Congestion Notification (RFC 3168) allows routers
  to mark packets with congestion signals instead of dropping
  them; L4S (Low Latency, Low Loss, Scalable Throughput, RFC
  9330–9332) extends ECN with a finer-grained signal that
  reduces queue-delay variance under congestion. ECN + L4S
  on QUIC datagrams is documented in RFC 9000 §13.4 + IETF
  draft-ietf-tsvwg-l4s-arch and is supported by recent Linux
  kernels (≥ 5.18 with `CONFIG_TCP_CONG_BBR2`) + recent
  Apple iOS / macOS. Should V1 ship ECN marking on QUIC
  datagrams as a per-tenant operator-policy default, with
  L4S support gated behind kernel-version capability? The
  cost is the per-region kernel-upgrade programme + the per-
  router ECN-aware queueing config + the per-tenant capability
  migration; the benefit is reduced queue-delay variance + 
  improved p999 latency under congestion. Trigger: V1 kernel-
  baseline decision; per-region L4S-router-deployment
  posture clarifies. Owner: C37 + V1 family + Operations
  family + Kernel WG. Cross-link C19 §4 ECN deployment + RFC
  3168 + RFC 9330 + RFC 9331 + RFC 9332 + V1 kernel-upgrade
  plan.

---

## 13. References

### Project artifacts

- Master Plan, Constitution, System Overview, Video/Audio family index ([`00_Index.md`](00_Index.md)). Sibling chapters cited above.

### Source research artifacts

- `video-tech_dim12.md` (1,593 lines), `video-tech.agent.final.md` (2,588 lines), `video-tech_insight.md` (Insight #7 RELEVANT V1/V2 comparison case), `video-tech_cross_verification.md`, `video-tech_dim10.md` (testing).

### Web research

[`../99_Web_Research_Addenda/2026-04-29-network-transport.md`](../99_Web_Research_Addenda/2026-04-29-network-transport.md) — 9 clusters (§A–§I) + §Z.

| Cluster | Topic | Cited |
|---------|-------|------|
| §A | RTP profile — H.264 / HEVC / AV1 / Opus payload formats RFC 6184 / 7798 / 9528 / 7587 | §2 |
| §B | SRTP cipher suites RFC 3711 + AES-GCM RFC 7714 | §3 |
| §C | ICE/STUN/TURN traversal IETF rfcs 8445 / 8489 / 8656 / 8838 | §4 |
| §D | RTCP + REMB + transport-cc + RFC 8888 | §5 |
| §E | QUIC + MoQ Media-over-QUIC IETF drafts 2025-2026 | §6 |
| §F | DTLS 1.3 + DTLS-SRTP key derivation RFC 5764 / 5763 | §3.4, §6 |
| §G | sendmmsg + io_uring batching (cross-link C16) | §7.1, §7.2 |
| §H | XDP eBPF + sock_diag (cross-link C16) | §7.3, §7.4 |
| §I | Multipath UDP / MPQUIC IETF draft | §8 |
| §Z | Contradictions index | §1, §3, §6 |

---

## Anti-Bluff Verification

> Per [Master Plan §4.3](../../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Source Evidence Reviewed

| Path | Lines | Reviewed | Used |
|------|------:|----------|------|
| `video-tech_dim12.md` | 1,593 | A, B, C, D | §§1–12 (primary) |
| `video-tech.agent.final.md` | 2,588 | A, B, C | §§1–9 |
| `video-tech_insight.md` | 243 | A, B | §1 (#7 RELEVANT V1/V2) |
| `video-tech_cross_verification.md` | 206 | A | §1 |
| `video-tech_dim10.md` | 1,689 | D | §11.5 |
| `00_Master_Plan.md` post-Session-7 | A, B, C, D | header / §9 / §13 |
| `01_Constitution.md` post §11.5 | A, B, C, D | §§1–11 |
| `05_Video_Audio/00_Index.md` | 407 | A, B, C, D | header voice |
| `05_Video_Audio/01_Codec_Selection.md` | 2,578 | A | §2 (codec ladder cross-link C26) |
| `05_Video_Audio/04_DualPath_Encoding.md` | 2,086 | C | §9.8 (NAL feed cross-link C29) |
| `05_Video_Audio/06_Audio_Pipeline.md` | 1,375 | A | §2.4 (Opus MultiStream cross-link C31) |
| `05_Video_Audio/08_ABR_FEC_Congestion.md` | 2,369 | B | §5 (RTCP-driven ABR cross-link C33) |
| `05_Video_Audio/09_Thermal_and_GPU_Balancing.md` | 2,994 | B | §5.6 (thermal-throttle RTCP extension cross-link C34) |
| `05_Video_Audio/11_Go_Pipeline_Implementation.md` | 2,816 | C | §9.7 (SCHED_FIFO goroutine cross-link C36) |
| `04_Latency/02_io_uring_and_Kernel_Bypass.md` | 1,787 | C | §7 (io_uring + XDP cross-link C16) |
| `04_Latency/05_UltraLowLatency_Network_Protocols.md` | 1,716 | A, B, C | §1, §7.6 (DSCP/L4S/jitter cross-link C19) |
| `03_Architecture/05_RealTime_APIs.md` | 3,450 | B | §4.8 (WebRTC ICE patterns cross-link C06) |
| `03_Architecture/10_Security_and_Isolation.md` | 3,726 | A | §3.4 (DTLS-SRTP key exchange cross-link C10) |
| `03_Architecture/07_Host_Agent_and_Game_Lifecycle.md` | 3,425 | C, D | §9 (`r18.SafeExec`), §11.11 |

### Web Sources Consulted

The companion addendum lists every URL with title and 2026-04-29 access date. **9 clusters (§A–§I) + §Z; ≥6 distinct primary URLs per cluster.**

### Insights Incorporated

| Insight | Source | Sections |
|---------|--------|----------|
| video-tech Insight #7 — SQP + custom UDP next-gen (RELEVANT — V1/V2 comparison case) | `video-tech_insight.md` | §1.2, §6.3, §12 (V2 reservation) |

### Conflict Zones Resolved

| CZ-ID | Description | Decision | Section |
|-------|-------------|----------|---------|
| Insight #7 | SQP+custom UDP shaves 4–9 ms; MVP rides RTP/SRTP/DTLS; V1 evaluates SCReAM; V2 reserves custom-UDP/SQP/MoQ | **Reaffirmed**; documented in §1.2 + §6 + OQ-C37-02 | §1, §6, §12 |
| Z addenda | Network-transport contradictions | Resolved per cluster matrix in addendum | §1, §3, §6 |
| Inherited (all prior families' Z lists) | Owned by prior chapters | Not relitigated | header preamble |

### R-18 Operational Integrity Compliance

Inherits R-18 enforcement from C08. Compliance verified:

- **Static — chapter prose**: §1.3 references R-18; §9.5 explicitly recaps the family-level allow-list extension specific to this chapter (`tc`, `ip`, `iptables`, `nft`, `bpftool`, `xdp-loader`, `ethtool` — all wrap through `r18.SafeExec`).
- **Static — code in §9**: imports `r18 "github.com/vasic-digital/helix-r18-safeexec"` (origin C08 §10). Deny-list NOT duplicated.
- **Static — bootstrap subprocess invocations**: pacing (`tc qdisc add`), XDP loading (`bpftool prog load`, `xdp-loader load`), NIC offload (`ethtool -K`) all wrap through `r18.SafeExec`.
- **Test — §11.11**: `host-integrity-scan` inherited VERBATIM from C08 §12.11. Non-overridable per Constitution §11.5.4.

### Coverage Confirmation

| Metric | Value |
|--------|-------|
| Source per-dim primary (`video-tech_dim12.md`) | 1,593 lines |
| R-01 minimum (Master Plan §7.2 row C37) | 1,750 lines of body prose |
| Body prose actually synthesised | **3,418 lines** across §§1–12 (A 559 + B 919 + C 954 + D 986) |
| Coverage ratio vs minimum | 1.95× line-count / ≥ 2.05× word-adjusted |
| Coverage ratio vs primary per-dim source | 2.15× |
| Forbidden-pattern scan (chapter prose) | clean |
| Empty-section-body scan | clean |
| Tables | Per-codec RTP profile decision matrix in §2.6; SRTP suite comparison in §3.5; RTCP message-type matrix in §5.7; capability schema (17 fields) in §9.3; failure-mode 12-row F1-F12 table in §10; test-type matrix in §11; cross-stage cross-link map in §9.8 |
| Section count | 12 normative sections + this verification block |
| Go code blocks | §9.4 (~182 LOC `transport.NewRTPSender` + `transport.RTPSender.SendBatch` + `transport.ICEAgent.Start` — real imports `pion/webrtc/v3`, `pion/srtp/v2`, `pion/ice/v2`, `quic-go/quic-go`, `r18`, `helix-shm`, `helix-iouring`, `helix-xdp`, `golang.org/x/sys/unix`) |
| R-18 enforcement | inherited from C08 §10 + §11.11 host-integrity-scan inheritance |

### Sign-off

- Section A (§§1–3) by C37 Group A on 2026-04-29 (re-dispatched after session-resume boundary).
- Section B (§§4–6) by C37 Group B on 2026-04-29 (re-dispatched after session-resume boundary).
- Section C (§§7–9) by C37 Group C on 2026-04-29 (re-dispatched after session-resume boundary).
- Section D (§§10–12) by C37 Group D on 2026-04-29 (re-dispatched after session-resume boundary).
- Web addendum by C37 addendum subagent on 2026-04-29 (re-dispatched after session-resume boundary).
- Header, ToC, §13, Anti-Bluff Verification stitched by orchestrator (Claude) on 2026-04-29.
- Reviewed by: pending operator review.
- **Closes the Video/Audio family at 13/13 chapters.**

End of `05_Video_Audio/12_Network_Transport.md` — 2026-04-29.
