# Web Research Addendum — Streaming Protocols & Codecs

> **Topic:** Streaming protocols, codecs, hardware encoders, congestion control circa April 2026.
> **Owning chapter:** [`../03_Architecture/01_Streaming_Protocols_and_Codecs.md`](../03_Architecture/01_Streaming_Protocols_and_Codecs.md) (C02).
> **Compiled by:** subagent (C02).
> **Date:** 2026-04-28.
> **Status:** Append-only. Subsequent edits to the chapter that need new web evidence MUST add a separate dated addendum.

This addendum captures the web sources cited heavily by the chapter
above. Each entry includes the resolved URL, an access date, and a
short extracted summary that the chapter relies on. The chapter's
`## Anti-Bluff Verification` block lists every source that resolves
here.

---

## A. Pion WebRTC v4 status (Apr 2026)

| Aspect | Finding | Source |
|--------|---------|--------|
| Active major version | v4 line; latest in 2025 was v4.2.x with v4.2.4 adding DTLS option upgrades and `ICECandidatePoolSize` support | [Releases · pion/webrtc](https://github.com/pion/webrtc/releases), accessed 2026-04-28 |
| ICE-Lite mode | `SettingEngine.SetLite(true)` puts the agent into ICE-Lite; recommended for publicly-routable servers behind a fixed IP, paired with NAT 1:1 mappings or the new `SetICEAddressRewriteRules` (replaces deprecated NAT 1:1 API) | [Pion Discussion #1983 — How to use the pion ice-lite](https://github.com/pion/webrtc/discussions/1983), accessed 2026-04-28 |
| ICE renomination | Added in v4.2.0; toggled via `SettingEngine`, removes the candidate-pair re-selection penalty when network changes | [Release v4.2.0 · pion/webrtc](https://github.com/pion/webrtc/releases/tag/v4.2.0), accessed 2026-04-28 |
| FlexFEC | Stable since v4.2.0 — explicit "production ready" wording in release notes | [Release v4.2.0 · pion/webrtc](https://github.com/pion/webrtc/releases/tag/v4.2.0), accessed 2026-04-28 |
| Build / portability | Pure Go; no CGO; supports `js/wasm`; supported Go ≥ 1.22 per `go.mod` | [pkg.go.dev/github.com/pion/webrtc/v4](https://pkg.go.dev/github.com/pion/webrtc/v4), accessed 2026-04-28 |

The chapter relies on (a) ICE-Lite for the host agent and edge relays
where the WAN address is stable, (b) FlexFEC's production status to
justify dropping the custom XOR-FEC fallback, and (c) ICE renomination
for fast handover when a client roams between Wi-Fi and mobile data.

---

## B. AV1 hardware encode landscape (Apr 2026)

| Vendor / generation | AV1 encode? | AV1 decode? | Notes | Source |
|---|---|---|---|---|
| NVIDIA RTX 40 (Ada) | Yes (since 2022) | Yes | Dual NVENC on RTX 4070 Ti+ enables Split Frame Encoding (SFE) for 4K60/8K30 | [transcodely AV1 in 2026](https://www.transcodely.com/blog/av1-in-2026), accessed 2026-04-28 |
| NVIDIA RTX 50 (Blackwell) | Yes | Yes | Dual NVENC standard; lower power than Ada at the same throughput; SFE available across the line | [transcodely AV1 in 2026](https://www.transcodely.com/blog/av1-in-2026), accessed 2026-04-28 |
| Intel Arc Alchemist | Yes | Yes | Tiger Lake / Alchemist were the first consumer AV1 encoders | [Apple AV1 Support 2026 (videoconverterfactory)](https://www.videoconverterfactory.com/multimedia-solution/apple-av1.html), accessed 2026-04-28 |
| Intel Arc Battlemage | Yes | Yes | Improved RD performance vs Alchemist; Intel ULL preset still hits 5 frames @ 60 fps | [transcodely AV1 in 2026](https://www.transcodely.com/blog/av1-in-2026), accessed 2026-04-28 |
| AMD RDNA 3 (RX 7000) | Yes | Yes | First AMD generation with AV1 encode | [Apple AV1 Support 2026 (videoconverterfactory)](https://www.videoconverterfactory.com/multimedia-solution/apple-av1.html), accessed 2026-04-28 |
| AMD RDNA 4 (RX 9000) | Yes (mid/high), **No on Navi 44** | Yes | Navi 44 entry-level cards explicitly omit hardware encoders, restricting them for streaming workloads | [Igor's Lab — AMD RDNA 4: AV1 coding and limitations for entry-level GPUs](https://www.igorslab.de/en/amd-rdna-4-av1-coding-and-restrictions-for-beginner-gpus/), accessed 2026-04-28 |
| Apple M3 / M4 (standard) | No (decode only on M3, decode + iPad Pro M4) | M3+ | Standard M3 and standard M4 still do AV1 in software for encode | [Bitmovin — Apple AV1 Support](https://bitmovin.com/blog/apple-av1-support/), accessed 2026-04-28 |
| Apple M4 Ultra / M5 Pro / M5 Max | Yes | Yes | First Apple Silicon parts with AV1 hardware encode; M5 Pro/Max landed March 2026 | [transcodely AV1 in 2026](https://www.transcodely.com/blog/av1-in-2026), accessed 2026-04-28 |
| Apple A17+ (iPhone 15 Pro+) | n/a (decode only) | Yes | iOS Safari AV1 only on these devices — narrows the addressable mobile-web audience | [Bitmovin — Apple AV1 Support](https://bitmovin.com/blog/apple-av1-support/), accessed 2026-04-28 |

The chapter uses this matrix to (a) keep H.264 as guaranteed fallback,
(b) negotiate AV1 only when both sides advertise hardware support per
this table, and (c) flag Navi 44 / standard M3 / standard M4 as known
"AV1-capable in marketing, software-only in practice" platforms that
require the negotiation path to fall back to HEVC.

---

## C. QUIC datagrams (RFC 9221) vs WebRTC for game streaming

| Source | Headline finding | Access date |
|--------|------------------|-------------|
| [RFC 9221 — An Unreliable Datagram Extension to QUIC](https://datatracker.ietf.org/doc/html/rfc9221) | DATAGRAM frames negotiated at handshake via `max_datagram_frame_size`; not retransmitted by QUIC; sender controls flow with priority and pacing; receiver MAY drop on backpressure | 2026-04-28 |
| [arXiv — Streaming Remote rendering services: QUIC vs WebRTC](https://arxiv.org/html/2505.22132v1) | RTP-over-QUIC reduces session-establishment latency by ~90 ms vs WebRTC in 5G testbed (no SDP, no ICE); steady-state latency comparable but QUIC connection-migration helps mobile roams | 2026-04-28 |
| [quic-go datagrams documentation](https://quic-go.net/docs/quic/datagrams/) | Production-ready Go implementation; supports unreliable datagrams; integrates with congestion controller of choice (Cubic, BBR) | 2026-04-28 |

The chapter uses these to (a) keep WebRTC as the default but treat
QUIC datagrams as a deliberate Phase-2 alternative for the native
client transport, (b) document `max_datagram_frame_size` negotiation
as the migration trigger, (c) note the connection-migration property
as the headline benefit for mobile clients roaming Wi-Fi → 5G mid-
session.

---

## D. SQP — Google's Scalable Quality Protocol

| Source | Headline finding | Access date |
|--------|------------------|-------------|
| [Google Research — SQP: Congestion Control for Low-Latency Interactive Video Streaming](https://research.google/pubs/sqp-congestion-control-for-low-latency-interactive-video-streaming/) | SQP samples bandwidth via paced packet trains coupled to frames; adaptive one-way delay; deployed against Stadia/AR services | 2026-04-28 |
| [arXiv 2207.11857 — SQP paper](https://arxiv.org/abs/2207.11857) | 2–3× higher bandwidth than WebRTC GCC under TCP competition; 27% / 15% more sessions hitting "high bitrate + low delay" on LTE / Wi-Fi vs Copa in real A/B tests | 2026-04-28 |

No new follow-up paper from Google was found in 2025 or 2026; the
2022 paper remains the canonical reference. The chapter therefore
treats SQP as a Phase-2 experimental addition, not a Phase-1 default.

---

## E. Sunshine / Moonlight 2026 status

| Source | Headline finding | Access date |
|--------|------------------|-------------|
| [Mustafa.net — Moonlight + Sunshine 2026](https://mustafa.net/2026/02/22/moonlight-sunshine-free-open-source-remote-gaming-stream-2026/) | Sunshine continues to support NVENC, AMF, QSV, VAAPI; works without an NVIDIA GPU on the host; AV1 encoding requires a supported encoder on the host (RTX 40+, Intel Arc, AMD 7000+) | 2026-04-28 |
| [GamingOnLinux — Sunshine HDR for Linux/Steam Deck](https://www.gamingonlinux.com/2024/06/moonlight-pc-for-game-streaming-via-sunshine-gets-hdr-support-for-linux-steam-deck/) | Sunshine added Linux HDR via PipeWire + KMS metadata in mid-2024; Moonlight clients on Steam Deck consume it | 2026-04-28 |
| [LizardByte Sunshine docs — Getting Started](https://docs.lizardbyte.dev/projects/sunshine/latest/md_docs_2getting__started.html) | Server lists supported codecs per platform, including AV1 path on Linux via VAAPI | 2026-04-28 |
| [Moonlight TV — AV1 streaming slow/laggy issue #386](https://github.com/mariotaku/moonlight-tv/issues/386) | Recent Moonlight builds have known AV1 decode-pacing issues on some embedded clients — practical reason HelixPlay does not advertise AV1 to TVs without explicit capability proof | 2026-04-28 |

The chapter cites these to anchor the "Sunshine++" pillar: HelixPlay
inherits Sunshine's mature capture/encode/stream core, while the
Moonlight protocol stays as a compatibility surface for users who
already have Moonlight-only clients.

---

## F. NVIDIA Reflex / Frame Warp (Reflex 2)

| Source | Headline finding | Access date |
|--------|------------------|-------------|
| [NVIDIA GeForce News — Reflex 2 With Frame Warp](https://www.nvidia.com/en-us/geforce/news/reflex-2-even-lower-latency-gameplay-with-frame-warp/) | Frame Warp re-projects the rendered frame using the latest mouse input just before scan-out; THE FINALS at 4K on RTX 5070 went 56 ms → 27 ms (Reflex) → 14 ms (Reflex 2) | 2026-04-28 |
| [Tom's Guide — Reflex 2 in The Finals](https://www.tomsguide.com/computing/i-just-used-nvidia-reflex-2-playing-the-finals-heres-what-the-latency-drop-actually-feels-like) | Independent test confirms ~75% E2E latency reduction in supported title | 2026-04-28 |
| [NVIDIA — How do I enable Low Latency Reflex streaming mode on GeForce NOW Ultimate](https://nvidia.custhelp.com/app/answers/detail/a_id/5402/) | Reflex is integrated into GeForce NOW's cloud streaming pipeline at 360/240/120/60 Hz tiers | 2026-04-28 |

The chapter draws the boundary at Reflex availability: when the host
has a supported GPU and the game has the Reflex SDK integrated, the
host agent advertises a `reflex=true` capability and the client UI
exposes the per-game tier; otherwise the host falls back to the
software-frame-pacing path.

---

## G. Index of citations (for the chapter's References section)

The chapter's `§12 References` section consolidates the URLs above
with the source-research artifacts. This addendum is the dated
authoritative copy for everything web-sourced; the chapter cross-
references this file by relative path, e.g. `[Web addendum 2026-04-28
§B]`.

---

## Sign-off

Compiled-by: subagent (C02) on 2026-04-28.
Reviewed-by: pending orchestrator review.
End of addendum 2026-04-28-streaming-protocols-and-codecs.
