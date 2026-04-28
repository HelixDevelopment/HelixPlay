# HelixPlay System Overview & Vision

> **Slogan:** "Ultimate gaming experience!"
>
> **Source dimensions:** all three MVP streams at the meta level
> (`docs/research/chapters/MVP/01_base`, `02_latency`, `03_video_technology`).
> **Cross-links:** [Master Plan](00_Master_Plan.md) · [Constitution](01_Constitution.md) ·
> [Architecture index](03_Architecture/00_Index.md) ·
> [Latency index](04_Latency/00_Index.md) ·
> [Video/Audio index](05_Video_Audio/00_Index.md).
> **Status:** Draft v1 — 2026-04-28.
> **Audience:** product, architecture, engineering leads. Newcomers
> should read this document first, then the Constitution, then dive
> into the chapter most relevant to their role.

---

## Table of Contents
1. Vision & Slogan
2. Problem Statement
3. Reference User Journey
4. System Boundaries
5. Topology at a Glance
6. Client Matrix
7. Host Matrix
8. End-to-End Dataflow
9. Latency Budget Snapshot
10. Codec & Transport Posture
11. Catalog & Content Story
12. White-Label Posture
13. Tenancy & Identity
14. Test Posture
15. Release Trains
16. Glossary
17. Anti-Bluff Verification

---

## 1. Vision & Slogan

HelixPlay turns any household machine with a powerful GPU into a remote
**gaming appliance** that streams a console-class experience to any
client device a player owns — desktop, phone, tablet, browser tab, or
TV — with **PS4 Pro–class UX** and **zero perceived lag**, while
remaining **fully self-hostable**, **fully open**, and **white-labellable**
for partners.

The slogan — *"Ultimate gaming experience!"* — is the brief in five
words: best-in-class video, best-in-class audio, best-in-class controller
fidelity, best-in-class catalog, all running on consumer hardware the
player already owns or rents directly, with no proprietary middleman.

The product narrative pulls from three commercial precedents and
one architectural one:

- **Sony PlayStation Plus Premium / PS Now** — the UX bar for catalog
  navigation, quick resume, instant join, controller fidelity. PS4/PS5
  patterns are the default visual language (cloudgaming Insight #6 —
  legal safe harbour).
- **NVIDIA GeForce NOW / Microsoft xCloud** — the latency bar for
  cloud streaming. Sub-50 ms WAN, sub-30 ms LAN, p999 not p50 (latency
  Insight #2).
- **Parsec / Moonlight + Sunshine** — the architectural precedent
  for self-hosted streaming. HelixPlay's host agent is a "Sunshine++"
  evolution (cloudgaming Insight #1).
- **Steam Big Picture / Apple TV / Android TV Leanback** — the TV-first
  10-foot UX bar.

What HelixPlay adds beyond those precedents:

- **Triple-stack client convergence** — Wails (desktop), Flutter+Go FFI
  (mobile/TV), Angular+Go WASM (web), all sharing one Go core
  (cloudgaming Insight #3).
- **Full controller-feature fidelity over network**: DualSense haptics,
  adaptive triggers, gyro, accelerometer, audio jack on the controller
  — all forwarded with ≤1 ms effective latency. This is the hidden
  differentiator (cloudgaming Insight #2).
- **Zero-impact recording** ("DVR for your gaming PC") — the host
  records every session locally to NVMe with background sync to
  user-controlled storage (video-tech Insight #10).
- **White-label–first** — every aspect of the client UI, the catalog
  metadata, and the brand surface is themable per tenant, enabling
  Gaming-as-a-Service for ISPs, hotels, hospitals, and venues
  (cloudgaming Insight #8).
- **Constitutional anti-bluff posture** — see [Constitution §1](01_Constitution.md#1-the-anti-bluff-pledge-r-02-r-13).
  Quality is structural, not a vibe.

---

## 2. Problem Statement

A modern gaming PC is **expensive and stationary**. A modern player
is **mobile**. The two facts pull in opposite directions, and the
existing solutions for closing the gap each carry a tax:

| Solution                 | Tax                                                                                  |
|--------------------------|--------------------------------------------------------------------------------------|
| Cloud gaming (GFN, xCloud)| Subscription cost; library limited to vendor catalog; no own-game streaming.         |
| In-home streaming (Steam Link, Moonlight) | LAN-only by default; client device limited; weak white-label story; partial controller forwarding. |
| Carrying the laptop      | Battery life, bag space, thermal limits, fan noise.                                  |
| Console-only             | Closed ecosystem; no PC-only titles; no mod scene; no dev tools.                      |

HelixPlay's wager is that the **streaming-from-your-own-host** model
becomes acceptable when:

1. Latency is **p999-bounded** (latency Insight #2), not just
   p50-bounded.
2. The **controller "feel"** is faithful (cloudgaming Insight #2).
3. The **UX** matches a console (cloudgaming Insight #6).
4. The setup is **frictionless** for non-technical users — and
   **self-hostable** for technical ones, so the same product works
   for an ISP and a power user.
5. The **business surface** (catalog, theming, monetization, tenancy)
   is general enough that partners can resell it.

This is the gap HelixPlay closes.

---

## 3. Reference User Journey

The "golden path" the system MUST deliver flawlessly. Every other
flow is derived from this one.

### 3.1 First-time setup (host)

1. Operator runs the **Containers** bootstrapper on the gaming PC.
2. Bootstrapper installs (containerised) the **Host Agent**, the
   **Capture Service**, the **Encoder Service**, and the
   **Discovery Beacon**.
3. Host Agent enumerates installed games (Steam, Epic, GOG, standalone)
   and publishes capability metadata (GPU model, supported codecs,
   max resolution, refresh rate, NVENC session count, thermal headroom).
4. Discovery Beacon advertises the host on the LAN via mDNS and to the
   user's account on the rendezvous service.

### 3.2 First-time setup (client)

1. Player installs the HelixPlay client (Wails desktop, Flutter mobile,
   Compose-for-TV TV app, or opens the Angular web client).
2. Player signs in via OAuth2/OIDC. Input-constrained devices use
   **Device Authorization Grant (RFC 8628)**.
3. Player pairs a controller — wired USB (preferred), 2.4 GHz dongle,
   or Bluetooth (with documented latency tradeoff).
4. Client discovers hosts on the LAN (via mDNS) and the operator's
   personal cloud roster (via the rendezvous service).

### 3.3 Play (the "PS4 moment")

1. Player opens the HelixPlay client.
2. Landing screen renders **the catalog** — horizontal hero shelf,
   horizontally-scrolled rows of game cards with 4K covers and
   screenshots, "Continue Playing" row pinned at top.
3. Player picks a game card.
4. Client → Rendezvous → Host: "I want to play `<game-id>` on
   `<host-id>`."
5. Host Agent **launches the game**, captures the first rendered
   frame within the latency budget (§9), encodes it, transmits it.
6. Client decodes and renders. Controller input is forwarded over
   WebRTC DataChannel (web) or custom UDP (native).
7. Pressing **Home** returns the player to the catalog. The running
   game is **suspended-or-kept-running** per its compatibility class
   (some games tolerate suspension; others must keep running in the
   background — see `03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`).
8. Player picks another game; the previous one closes safely or
   is replaced. **No data corruption, no crashed save files.**
9. Optionally: the session was recorded; player retrieves the recording
   from their NAS or local SSD (video-tech Insight #4).

### 3.4 The "10-foot moment" (TV)

Same flow, but:

- D-pad / remote / controller navigation, no touch.
- 10-foot typography and overscan-safe layout.
- Voice search (Android TV).
- HDR-aware UI (HDR10+ when display supports it; tone-mapped fallback otherwise).
- Quick-resume tile on the launcher home.

### 3.5 The white-label moment

- An ISP partner deploys HelixPlay under their brand "ISP-Play."
- Theme tokens, logos, fonts, hero artwork, even category labels
  swap at runtime per tenant. The user never sees the HelixPlay name.
- Catalog is filtered to titles licensed for the tenant's audience.
- Identity provider is the ISP's own OIDC.

---

## 4. System Boundaries

What is **inside** HelixPlay:

- **Host Agent** (per-OS) — game lifecycle, capture, encode, stream,
  controller injection.
- **Backend services** — rendezvous, catalog, identity, theming,
  recording metadata, telemetry, billing-meter (later phases).
- **Client apps** — Desktop (Wails), Mobile (Flutter+Go FFI), Web
  (Angular+Go WASM), TV (Compose for TV, Flutter for TV fallback).
- **Shared Go core** — controller protocol, streaming protocol state
  machine, auth client, catalog client. Compiled three ways
  (c-shared, native binary, WASM) per cloudgaming Insight #3.
- **Containers** submodule — every image we publish.
- **Challenges** submodule — production-equivalent E2E tests.
- **HelixQA** — autonomous QA system.
- **Observability stack** — OpenTelemetry collectors, log/metric/trace
  storage, NATS event bus.

What is **outside**:

- The games themselves (we launch, capture, and forward — we don't
  ship game binaries).
- The metadata sources of record (IGDB, SteamGridDB, Steam, Epic,
  GOG) — we cache and remix, we don't author.
- Anti-cheat vendors (we accommodate; we do not negotiate).
- The display device (we render to whatever it is; latency floor
  imposed by the display is acknowledged — video-tech Insight #6).
- The player's home network beyond the LAN we discover on.

---

## 5. Topology at a Glance

```
                   ┌─────────────────────────────────┐
                   │         Operator Tenant         │
                   │  (HelixPlay-hosted or partner)  │
                   └───────────────┬─────────────────┘
                                   │
                  mTLS / gRPC / HTTP3 + Brotli
                                   │
            ┌──────────────────────┴──────────────────────┐
            │           Backend Services Plane            │
            │  rendezvous · catalog · identity · theme    │
            │  recording-meta · telemetry · billing       │
            │           (Go, gRPC, NATS bus)              │
            └─────────┬────────────────────────┬──────────┘
                      │                        │
   service discovery                 events (NATS JetStream)
                      │                        │
        ┌─────────────┴─────────────┐      ┌───┴────────────────┐
        │     Host Agents (×N)      │      │ HelixQA + Challenges│
        │  Win / macOS / Linux       │      │  autonomous QA     │
        │  capture → encode → stream │      └────────────────────┘
        └─────────────┬─────────────┘
                      │ WebRTC / custom UDP
            ┌─────────┴─────────────┐
            │      Clients          │
            │ Wails · Flutter · Web │
            │ · Compose-for-TV      │
            └───────────────────────┘
```

Detailed C4 contexts and containers live in
`03_Architecture/00_Index.md` and the per-section files.

---

## 6. Client Matrix

| Surface  | Framework                        | Go integration              | Why                                                                                                          |
|----------|----------------------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------|
| Desktop  | Wails v2                         | Native Go core (in-proc)    | Best Go-native desktop story; web frontend permits a single design system shared with the web client.        |
| Mobile (iOS/Android) | Flutter (Dart)        | `c-shared` Go core via FFI  | Best accessibility + TV/D-pad story; Dart drives UI, Go drives protocol logic.                              |
| TV (Android TV) | Compose for TV (Kotlin), with Flutter as fallback | Same `c-shared` Go core | Compose for TV is the official Google direction (Leanback deprecated). Flutter retained as fallback for low-end SoCs. |
| TV (Apple TV) | SwiftUI on tvOS                  | Same `c-shared` Go core     | tvOS-only path; controlled-input UX, AirPlay-aware.                                                          |
| Web      | Angular                          | Go core compiled to WASM    | Browser ubiquity; share design system with desktop.                                                          |

**Common Go core responsibilities**:

- Controller binary protocol (encode/decode, sequencing, packet pacing).
- Streaming session state machine (offer/answer, ICE, fallback paths).
- Auth client (OAuth2/OIDC with Device Authorization Grant).
- Catalog client (gRPC + cache).
- Observability (OpenTelemetry exports, metric counters).

The C-API surface that crosses the FFI boundary is documented in
`03_Architecture/04_Go_Client_Ecosystem.md`. Ground rule: every API
that is exposed via FFI is also exposed via WASM with the same shape.

---

## 7. Host Matrix

| OS            | Capture API                                  | Encoder        | Virtual controller | Notes                                                                                                       |
|---------------|----------------------------------------------|----------------|--------------------|-------------------------------------------------------------------------------------------------------------|
| Windows 10/11 | DXGI Desktop Duplication (DDA), Windows.Graphics.Capture | NVENC / QSV / AMF | ViGEmBus successor (signed driver) | Anti-cheat constraint forces "official" capture only (Constitution §11.3). HDR via `DuplicateOutput1`.       |
| macOS 13+     | ScreenCaptureKit + IOSurface                  | VideoToolbox   | foohid kext / DriverKit | Apple Silicon path is privileged; ScreenCaptureKit zero-copy is mandatory.                                  |
| Linux         | KMS/DRM, PipeWire, DMA-BUF                    | VAAPI / NVENC / AMF | uinput              | PREEMPT_RT optional but recommended on dedicated hosts (latency CZ-03).                                     |

Host agent topology choices, OS-specific issues, and capability
advertisement are detailed in
`03_Architecture/03_Host_OS_Capture.md` and
`03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`.

---

## 8. End-to-End Dataflow

The following is the canonical dataflow for the **steady-state** of a
running session. Setup, NAT traversal, ICE candidate exchange, codec
negotiation, and quality re-negotiation are detailed in their own
chapters.

```
Player input
   │
   │ controller (USB 1000Hz / 2.4GHz / BT)
   ▼
Client: HID adapter ──► binary packet (16–32 B)
                              │
                              │ unreliable / unordered
                              ▼
                   WebRTC DataChannel  ◀── web client
                   custom UDP w/ DTLS  ◀── native client
                              │
                              │ LAN / WAN / mDNS-discovered host
                              ▼
Host Agent: virtual controller injection (ViGEm/uinput/foohid)
                              │
                              ▼
                   Game process consumes input
                              │
                              ▼
                   GPU renders frame
                              │
                              ▼
Capture: DXGI DDA / ScreenCaptureKit / DMA-BUF (zero-copy)
                              │
                              ▼
Encoder: NVENC / QSV / AMF / VideoToolbox / VAAPI
            (H.264 universal, HEVC/AV1 capable peers)
                              │
                              ▼
Packetizer: RTP / custom UDP, FEC + retransmit window
                              │
                              ▼
Transport: WebRTC (DTLS+SRTP) or custom UDP (DTLS 1.2 per packet)
                              │
                              ▼
Client decoder: hardware decode → IOSurface/EGLImage/SurfaceTexture
                              │
                              ▼
Renderer: VRR/G-Sync/FreeSync, frame pacing, HDR tone mapping
                              │
                              ▼
Display: OS scanout
```

Audio runs as a parallel pipeline on its own RTP stream (Opus
MultiStream up to 7.1, with passthrough for AC3/EAC3/Atmos when
the chain supports it — video-tech Insight #2). A/V sync is handled
at the client decoder per
`05_Video_Audio/06_Audio_Pipeline.md`.

Recording, when enabled, **forks at the encoder**: the same encoded
bitstream is written to a local NVMe ring buffer in MKV/fMP4 with
periodic background sync to the user-configured network target
(SMB/NFS/FTP/WebDAV). See
`05_Video_Audio/04_DualPath_Encoding.md` and
`05_Video_Audio/05_Recording_Storage.md`.

---

## 9. Latency Budget Snapshot

The detailed budget lives in
`04_Latency/00_Index.md` and per-stage files. The headline is:

| Stage                          | LAN target (p999) | WAN target (p999) | Notes                                                                                                  |
|--------------------------------|------------------:|------------------:|--------------------------------------------------------------------------------------------------------|
| Controller poll (1000 Hz USB)  |               1ms |               1ms | Constitution §6.1 enforces. BT falls back to 8 ms with documented tradeoff.                            |
| Input → host inject            |              <1ms |              <2ms | binary packet over UDP; host-side virtual controller.                                                  |
| Game render                    |             5–8ms |             5–8ms | Game-dependent; we cannot tune further than NVIDIA Reflex / Frame Warp can.                           |
| Capture                        |              <1ms |              <1ms | Zero-copy mandatory.                                                                                   |
| Encode                         |             5–8ms |             5–8ms | Intel ULL (5 frames @ ULL) is the floor; NVENC ~7 frames; AMD 6–9.                                    |
| Network transport              |             1–3ms |             5–25ms| Edge placement matters far more than codec optimisation (cloudgaming Insight #7).                      |
| Decode                         |             4–6ms |             4–6ms | Hardware decode mandatory.                                                                             |
| Display scanout                |             4–8ms |             4–8ms | VRR; "game mode" on TVs adds 30–100 ms unless explicitly disabled (video-tech Insight #6).            |
| **Total budget (glass-to-glass)** | **20–35 ms** | **35–60 ms**     | Aspirational floor: 30 ms LAN, 50 ms WAN.                                                              |

p999 is the metric, not p50 (latency Insight #2). The latency chapter
defines instrumentation, the testing chapter defines validation,
the operations chapter defines the dashboards.

---

## 10. Codec & Transport Posture

| Layer       | Default               | Capable upgrade                         | Universal fallback              |
|-------------|-----------------------|------------------------------------------|---------------------------------|
| Video codec | H.264 Main/High       | HEVC (native peers), AV1 (RTX 40+/Arc/M3+) | H.264 Baseline (always)         |
| Audio codec | Opus (MultiStream)    | EAC3/AC3 passthrough where supported     | Opus stereo                     |
| Transport (web) | WebRTC SRTP+DTLS  | n/a                                      | n/a                             |
| Transport (native) | Custom UDP (Parsec BUD–style) + DTLS 1.2 | SQP congestion control on top      | WebRTC SRTP+DTLS                |
| Compression | Brotli                | n/a                                      | gzip                            |

Rationale for the "H.264 default" choice — see video-tech Insight #3
(strategically optimal despite being technically inferior). H.264 is
the universal baseline; HEVC/AV1 are negotiated upgrades when both
peers support them and the bandwidth/quality math favours the upgrade.

---

## 11. Catalog & Content Story

The catalog is a **content business** wrapped around a thin technical
veneer (cloudgaming Insight #4). Treat it as such:

- **Sources**: IGDB (primary, paid Pro tier), SteamGridDB (community
  artwork), Steam API (where licensed), RAWG (fallback). Per-tenant
  filtering to comply with regional licensing.
- **4K assets**: WebP/AVIF, multi-resolution variants, lazy-loaded.
- **Search**: SQLite FTS5 on-device cache for "Continue Playing,"
  Meilisearch (or equivalent) on backend for global search.
- **User-contributed artwork**: yes (mirroring SteamGridDB's
  community model) — moderated.
- **Per-tenant catalogs**: isolated database schemas; per-tenant
  metadata overlays.
- **Continue Playing pinning**: persisted per user, per tenant, per
  client surface (TV vs phone may show different "next up" tiles).

Detailed schema, ingestion pipeline, refresh policy, rights handling
in `03_Architecture/06_Catalog_and_Assets.md`.

---

## 12. White-Label Posture

White-label is **not a skin**. It is a first-class architectural
property (cloudgaming Insight #8 — enables Gaming-as-a-Service).

- **Theme tokens**: 3-tier (primitive → semantic → component) per
  the Material Design 3 tonal-palette pattern. Style Dictionary v4
  is the build tool. CSS custom properties for runtime swaps.
- **Brand surface**: logo, name, colours, hero artwork, fonts,
  legal copy, support links — all per tenant.
- **Layout configuration**: grid density, list-vs-card, default
  filters — per tenant, per surface.
- **Identity provider**: per tenant (their OIDC, our OIDC, or
  social).
- **Catalog overlay**: per tenant licensing filters.
- **Recording defaults**: per tenant policy (off by default for
  hospitality, on for content creator tenants).

Detailed design system in
`03_Architecture/10_WhiteLabel_and_Theming.md`.

---

## 13. Tenancy & Identity

- **Multi-tenant from day one**. No "single tenant" mode that breaks
  later.
- **Tenant boundary**: database schema per tenant; storage prefixes
  per tenant; per-tenant rate limits.
- **Identity**: each tenant brings its own OIDC issuer; fall back to
  HelixPlay's hosted identity for self-hosted operators.
- **Authentication**: short-lived access tokens (≤15 min), rotation
  refresh tokens, mTLS between services.
- **Authorization**: RBAC at the tenant level; per-user policies
  managed in the identity service.
- **Privacy**: input streams treated as personal data
  (Constitution §11.4).
- **Audit logging**: tenant-scoped; integrates with the central
  observability bus.

Schema, sequence diagrams, and threat model in
`03_Architecture/09_Security_and_Isolation.md`.

---

## 14. Test Posture

Reproduced for visibility — full detail in
`07_Testing/00_Index.md`.

Every submodule, every file, ten test types
(Constitution §6.1). Unit tests **may** mock; everything else
**must not**. The Challenges submodule and HelixQA together prove
end-user usability before any release. Latency tests report
p50/p99/p999. Coverage gate is 100% across the union of test
types — not unit-only coverage. Anti-bluff CI scan is non-overridable
(Constitution §1.3).

---

## 15. Release Trains

The implementation phases in `09_Implementation_Phases/` map to
release trains:

| Train | Phases | What ships                                                                                  |
|-------|--------|---------------------------------------------------------------------------------------------|
| Foundation | P00, P01 | Containers submodule extension, local CI/CD, Constitution propagation across submodules. |
| Core platform | P02, P03 | vasic-digital submodules; backend services with gRPC + NATS.                            |
| Streaming MVP | P04 | Host Agent + Capture + Encode + Stream; web client minimal player.                         |
| Client family | P05 | Wails, Flutter, Compose-for-TV, Angular WASM all wired to the same Go core.                |
| Host depth | P06 | Game lifecycle, save sync, anti-cheat compatibility surface.                                |
| Latency depth | P07 | 1000 Hz polling, PREEMPT_RT recipe, NVIDIA Reflex/Frame Warp integration.                  |
| Audio depth | P08 | Surround sound chain, Opus MultiStream, passthrough.                                          |
| Recording depth | P09 | Local NVMe + background sync; circular buffer for instant replay.                           |
| Monetization | P10 | Tenancy billing meter, per-tenant identity, Device Authorization Grant.                      |
| Hardening | P11 | Security review, scan blockers, Challenges full pass, HelixQA full pass.                       |
| Beta | P12 | Closed beta; observability burn-in.                                                              |
| GA | P13 | General availability across all surfaces, all platforms, all tenants supported.                 |

No phase is allowed to ship until the previous phase's tickets are
**closed with evidence** per Constitution §8.4.

---

## 16. Glossary

- **Anti-bluff** — Constitution §1; a property of artifacts and tests.
- **Challenges** — production-equivalent E2E test suite. Constitution
  §6.1 #10. Source: `git@github.com:vasic-digital/Challenges.git`.
- **Clean host** — a host machine that uses only OS-provided capture
  APIs; no hooks, no DLL injection. cloudgaming Insight #5.
- **Cold path** — code paths not traversed per controller event or
  per frame; opposite of hot path.
- **DDA** — DXGI Desktop Duplication API (Windows native screen
  capture).
- **Device Authorization Grant** — RFC 8628; OAuth2 flow for
  input-constrained devices.
- **DualSense feature parity** — full forwarding of haptics,
  adaptive triggers, gyro, accelerometer, mic, headphone jack,
  touchpad, lightbar.
- **eARC** — enhanced Audio Return Channel; the only consumer
  interface that carries uncompressed multi-channel + Atmos.
- **FEC** — Forward Error Correction; reduces retransmissions.
- **Frame Warp** — NVIDIA Reflex 2 feature; rewarp the rendered
  frame at the last millisecond using the latest input.
- **GaaS** — Gaming-as-a-Service; the white-label business model
  enabled by the platform. cloudgaming Insight #8.
- **Glass-to-glass** — input-event glass to displayed-frame glass;
  the only latency that matters to the player.
- **GPUDirect RDMA** — NVIDIA tech for GPU-to-network without
  CPU touching the data.
- **HelixQA** — autonomous QA system at
  `git@github.com:HelixDevelopment/HelixQA.git`. Constitution §6.5.
- **Hot path** — every-controller-event or every-frame path.
  Constitution §14.
- **Host agent** — the per-OS service that lives on the gaming PC
  and turns the GPU's output into a stream.
- **Hybrid client** — the architectural choice forced by
  cloudgaming HC-03; one Go core, three UI front ends.
- **mTLS** — mutual TLS; service-to-service authentication.
- **NVENC / QSV / AMF** — NVIDIA / Intel / AMD hardware video
  encoders.
- **p999** — 99.9th-percentile latency. The only metric that
  matters for competitive play. latency Insight #2.
- **Quick resume** — the PS5/Xbox feature where switching games
  preserves the previous game's state. Phase 2; in MVP we
  implement graceful game-switching.
- **Rendezvous** — backend service that pairs clients with hosts
  outside the LAN.
- **Sunshine++** — the architectural pattern of forking Sunshine's
  capture/encode/stream and adding management features as a layer.
  cloudgaming Insight #1.
- **Tenant** — an operator deploying HelixPlay under their own brand;
  isolated DB schema, identity, and catalog.
- **The Ten** — the ten test types in Constitution §6.1.
- **Tone mapping** — adapting HDR content to SDR or to a different
  HDR format on the client.
- **VRR** — Variable Refresh Rate (G-Sync / FreeSync).
- **WebRTC DataChannel** — the unreliable, unordered transport used
  for controller input on web clients.
- **White-label** — running HelixPlay under a tenant's brand.

---

## 17. Anti-Bluff Verification

### Source Evidence Reviewed

- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/04_Request.md`
  — 99 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/01_Request.md`
  — first 80 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim_decomposition.md`
  — 70 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_insight.md`
  — 156 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_cross_verification.md`
  — 130 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_dim_decomposition.md`
  — 33 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_insight.md`
  — 100 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_cross_verification.md`
  — 101 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md`
  — 243 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/03_video_technology/02_Response/Agent_Results/research/video-tech_cross_verification.md`
  — 206 lines, reviewed 2026-04-28.

### Web Sources Consulted
- None at the System Overview level. The overview is derived from the
  three streams' synthesis files; deep web sourcing happens in the
  technical chapters.

### Insights Incorporated
- cloudgaming Insight #1 (Sunshine++) → §1, §15 (host agent path).
- cloudgaming Insight #2 (Controller as differentiator) → §1, §3.3.
- cloudgaming Insight #3 (Three clients, one Go core) → §1, §6.
- cloudgaming Insight #4 (Catalog as content business) → §11.
- cloudgaming Insight #5 (Anti-cheat clean host) → §7 (Notes column),
  Constitution §11.3.
- cloudgaming Insight #6 (PS4 UX as legal safe harbor) → §1, §3.3.
- cloudgaming Insight #7 (Edge > codec for latency) → §9.
- cloudgaming Insight #8 (White-label = GaaS) → §1, §12.
- latency Insight #1 (Microwave pipeline) → §8 dataflow.
- latency Insight #2 (p999 only metric) → §9, §14, Constitution §6.1.
- latency Insight #4 (Allocation-free hot path) → Constitution §5.4.
- video-tech Insight #2 (Audio passthrough constraint) → §8 (audio path).
- video-tech Insight #3 (H.264 sweet spot paradox) → §10.
- video-tech Insight #4 (Recording = save pattern) → §3.3 #9, §8 (recording).
- video-tech Insight #6 (Display pipeline floor) → §9 (display row note).
- video-tech Insight #10 (Recording differentiator) → §1.

### Conflict Zones Resolved
| CZ-ID | Conflict | Decision | Rationale |
|-------|----------|----------|-----------|
| cloudgaming CZ-01 | WebRTC vs custom UDP | Hybrid: WebRTC default + web; custom UDP for native | §10 documents both; native clients get the latency savings, web stays standards-compliant. |
| cloudgaming CZ-04 | Bluetooth controller latency | Both supported with documented tradeoff | §6 row + Constitution §6.1; BT permitted, USB/2.4 GHz preferred for competitive. |
| latency CZ-03 | PREEMPT_RT for hosts only | Hosts yes, clients no | §7 Notes column; PREEMPT_RT is host-side only. |
| video-tech CZ-1 | Intel non-standard B-frames | Accepted tradeoff | §10 default + capable upgrade story; client validation in test matrix. |
| video-tech CZ-6 | Recording impact on streaming | Conditionally resolved on thermals | §3.3 #9 + reference to dual-path encoding chapter that handles thermal-aware quality. |

### Coverage Confirmation
- This is an overview document; the line-floor in R-01 applies to the
  programme aggregate, not this single file. The overview's role is
  to be the navigable index of vision and posture; depth lives in
  the per-chapter files queued in
  [Master Plan §7.2](00_Master_Plan.md#72-queued).

### Sign-off
Executed by: Claude (orchestrator session 1)
Reviewed by: pending operator review
Date: 2026-04-28

End of System Overview v1 — 2026-04-28.
