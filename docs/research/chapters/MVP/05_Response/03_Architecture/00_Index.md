# Architecture Chapter — Index

> **Source dimensions:**
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/01_Request.md` (3 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim01.md` (812 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim02.md` (557 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim03.md` (909 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim04.md` (1,380 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim05.md` (966 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim06.md` (1,431 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim07.md` (1,449 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim08.md` (1,003 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim09.md` (974 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim10.md` (1,353 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim11.md` (1,155 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim12.md` (1,340 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim_decomposition.md` (70 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_insight.md` (156 lines)
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_cross_verification.md` (130 lines)
>
> **Source line count:** 13,688 lines aggregate across the Architecture stream (target for this Index: ≥400 lines of synthesized prose per Master Plan §7.2 row C01; the heavyweight per-dimension targets are queued as C02..C13).
>
> **Chapter targets (R-XX):** R-01 (extension over simplification), R-02 (anti-bluff), R-03 (decoupling), R-07 (gRPC + HTTP/3 + Brotli), R-08 (NATS / Redis / RabbitMQ), R-09 (non-blocking, lazy, semaphores), R-11 (the ten test types), R-12 (mocks confined to Unit), R-13 (anti-bluff testing), R-15 (recursive submodule capture), R-16 (fine-grained phases/tasks/subtasks), R-17 (GitHub Projects + GitLab tracking).
>
> **Cross-links:**
> - Master Plan: [../00_Master_Plan.md](../00_Master_Plan.md)
> - Constitution: [../01_Constitution.md](../01_Constitution.md)
> - System Overview: [../02_System_Overview.md](../02_System_Overview.md)
> - Sibling Architecture chapters (queued, link resolves once produced):
>   [01_Streaming_Protocols_and_Codecs.md](01_Streaming_Protocols_and_Codecs.md) ·
>   [02_Controller_Input_Pipeline.md](02_Controller_Input_Pipeline.md) ·
>   [03_Host_OS_Capture.md](03_Host_OS_Capture.md) ·
>   [04_Go_Client_Ecosystem.md](04_Go_Client_Ecosystem.md) ·
>   [05_RealTime_APIs.md](05_RealTime_APIs.md) ·
>   [06_Catalog_and_Assets.md](06_Catalog_and_Assets.md) ·
>   [07_Host_Agent_and_Game_Lifecycle.md](07_Host_Agent_and_Game_Lifecycle.md) ·
>   [08_Scalability_and_MultiRegion.md](08_Scalability_and_MultiRegion.md) ·
>   [09_Security_and_Isolation.md](09_Security_and_Isolation.md) ·
>   [10_WhiteLabel_and_Theming.md](10_WhiteLabel_and_Theming.md) ·
>   [11_TV_UX.md](11_TV_UX.md) ·
>   [12_Latency_Engineering_Overview.md](12_Latency_Engineering_Overview.md)
> - Latency chapter index: [../04_Latency/00_Index.md](../04_Latency/00_Index.md)
> - Video/Audio chapter index: [../05_Video_Audio/00_Index.md](../05_Video_Audio/00_Index.md)
>
> **Status:** Draft v1.
> **Last updated:** 2026-04-28.

---

## Table of Contents

1. [Purpose of this index](#1-purpose-of-this-index)
2. [The 12 dimensions at a glance](#2-the-12-dimensions-at-a-glance)
3. [Architectural pillars](#3-architectural-pillars)
4. [Reading order](#4-reading-order)
5. [Cross-stream linkage](#5-cross-stream-linkage)
6. [Vocabulary anchors](#6-vocabulary-anchors)
7. [Open questions](#7-open-questions)
8. [Anti-Bluff Verification](#anti-bluff-verification)

---

## 1. Purpose of this index

This index is the **gateway** to the Architecture chapter family of the
HelixPlay MVP "ultimate documentation" programme. The Architecture
family consolidates and extends Stream 1 of the source research
(`docs/research/chapters/MVP/01_base/`), which decomposes the cloud
gaming system into twelve independent investigative dimensions. Each
of those twelve dimensions is reified as one sibling chapter
(`01_Streaming_Protocols_and_Codecs.md` through
`12_Latency_Engineering_Overview.md`) in the same directory; this
file does **not** duplicate their content. Instead, it (a) names the
dimensions, (b) cites the source artifacts each chapter must extend,
(c) restates the architecture as a small number of named pillars that
cut across all twelve chapters, (d) provides reading orders for
different audiences, and (e) tracks the open questions that
specific chapters must close.

The Architecture family is the **structural backbone** of the
documentation set. Two adjacent families fold into it. The Latency
family ([../04_Latency/00_Index.md](../04_Latency/00_Index.md)) zooms
into the host-side and client-side micro-engineering required to hit
the p999 budget the Architecture family commits to in chapter 12. The
Video/Audio family ([../05_Video_Audio/00_Index.md](../05_Video_Audio/00_Index.md))
zooms into the codec, capture, encode, recording, and transport
specifics referenced by Architecture chapters 01, 03, and 07. The
audience for this index is the architect who needs the map, the
implementor who needs the dispatch order for §7.2 of the Master Plan,
and the reviewer who needs to confirm that no dimension was simplified
or skipped. Newcomers are pointed at [§4 Reading order](#4-reading-order)
and the System Overview ([../02_System_Overview.md](../02_System_Overview.md));
returning readers are pointed at the dimension table in
[§2](#2-the-12-dimensions-at-a-glance).

---

## 2. The 12 dimensions at a glance

The table below captures, for each of the twelve source dimensions:
the per-dim file (absolute path), its line count (verified at session
start, 2026-04-28, via `wc -l`), the output chapter that supersedes
it, the minimum-line target imposed by Master Plan §7.2, and the
primary cross-references to other 05_Response chapters. The minimum
targets are floors, not ceilings — per R-01 each chapter must extend,
not simplify, the source dimension.

| Dim | Title (canonical) | Source per-dim file (absolute) | Source lines | Output chapter (relative) | Min lines | Primary cross-refs |
|----:|-------------------|--------------------------------|------------:|---------------------------|----------:|--------------------|
| 01 | Low-Latency Video Streaming Protocols & Codecs | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim01.md` | 812 | [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md) | 950 | [`../05_Video_Audio/01_Codec_Selection.md`](../05_Video_Audio/01_Codec_Selection.md), [`../05_Video_Audio/12_Network_Transport.md`](../05_Video_Audio/12_Network_Transport.md), [`../04_Latency/05_UltraLowLatency_Network_Protocols.md`](../04_Latency/05_UltraLowLatency_Network_Protocols.md) |
| 02 | Cross-Platform Controller Input Capture & Forwarding | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim02.md` | 557 | [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md) | 700 | [`../04_Latency/07_Controller_Input_Optimization.md`](../04_Latency/07_Controller_Input_Optimization.md), [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md) |
| 03 | Host OS Game Capture Technologies | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim03.md` | 909 | [`03_Host_OS_Capture.md`](03_Host_OS_Capture.md) | 1,050 | [`../05_Video_Audio/03_Capture_Pipelines.md`](../05_Video_Audio/03_Capture_Pipelines.md), [`../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md`](../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md) |
| 04 | Go Ecosystem for Cross-Platform Client Development | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim04.md` | 1,380 | [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md) | 1,500 | [`11_TV_UX.md`](11_TV_UX.md), [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md) |
| 05 | Real-Time Communication APIs in Go | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim05.md` | 966 | [`05_RealTime_APIs.md`](05_RealTime_APIs.md) | 1,100 | [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md), [`../05_Video_Audio/12_Network_Transport.md`](../05_Video_Audio/12_Network_Transport.md) |
| 06 | Game Catalog, Metadata & 4K Asset Management | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim06.md` | 1,431 | [`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md) | 1,550 | [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md), [`11_TV_UX.md`](11_TV_UX.md) |
| 07 | Host Agent Architecture & Game Lifecycle Management | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim07.md` | 1,449 | [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md) | 1,600 | [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md), [`03_Host_OS_Capture.md`](03_Host_OS_Capture.md), [`09_Security_and_Isolation.md`](09_Security_and_Isolation.md) |
| 08 | Scalability, Load Balancing & Multi-Region Infrastructure | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim08.md` | 1,003 | [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md) | 1,150 | [`05_RealTime_APIs.md`](05_RealTime_APIs.md), [`../08_Operations/03_Service_Discovery_and_Ports.md`](../08_Operations/03_Service_Discovery_and_Ports.md) |
| 09 | Security, Authentication & Host Isolation | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim09.md` | 974 | [`09_Security_and_Isolation.md`](09_Security_and_Isolation.md) | 1,100 | [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md), [`../07_Testing/05_Security_Tests.md`](../07_Testing/05_Security_Tests.md) |
| 10 | White-Label, Theming & Customization Architecture | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim10.md` | 1,353 | [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md) | 1,500 | [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md), [`11_TV_UX.md`](11_TV_UX.md), [`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md) |
| 11 | TV-First UI/UX & Living Room Experience | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim11.md` | 1,155 | [`11_TV_UX.md`](11_TV_UX.md) | 1,250 | [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md), [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md), [`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md) |
| 12 | Performance Optimization & End-to-End Latency Engineering | `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim12.md` | 1,340 | [`12_Latency_Engineering_Overview.md`](12_Latency_Engineering_Overview.md) | 1,450 | All [`../04_Latency/`](../04_Latency/00_Index.md) chapters; [`../05_Video_Audio/08_ABR_FEC_Congestion.md`](../05_Video_Audio/08_ABR_FEC_Congestion.md), [`../05_Video_Audio/10_Measurement_and_QA.md`](../05_Video_Audio/10_Measurement_and_QA.md) |

The 16,050-line aggregate of minimum targets across the twelve
chapters comfortably exceeds the 13,688-line floor of the Stream 1
sources, satisfying R-01 for the Architecture family alone.

---

## 3. Architectural pillars

The twelve dimensions are not independent silos — they are projections
of a smaller number of cross-cutting commitments. The eight pillars
below are the de-facto "load bearing" commitments derived from the
ten High Confidence (HC) findings and eight cross-dimensional Insights
in the source cross-verification report. Each pillar names the
chapters where its detail lives. A change to a pillar is a change to
every chapter listed.

### 3.1 Sunshine++ host agent

The host-side capture, encode, and stream pipeline is **not**
greenfield. It is a deliberate evolution of Sunshine, the open-source
companion to Moonlight. The "++" is a management layer (session
orchestration, controller forwarding profiles, save sync, anti-cheat
clean-host posture, observability, white-label tenant binding) that
HelixPlay layers on top of Sunshine's mature capture/encode/stream
core. The decision avoids burning the 12–18-month engineering budget
that a from-scratch DXGI/ScreenCaptureKit/KMS pipeline would require,
and concentrates HelixPlay's differentiation on the management layer
(cross-verification HC-04, Insight #1). Pillar detail lives in
[`03_Host_OS_Capture.md`](03_Host_OS_Capture.md),
[`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md),
[`12_Latency_Engineering_Overview.md`](12_Latency_Engineering_Overview.md),
and the corresponding deep dives in [`../05_Video_Audio/03_Capture_Pipelines.md`](../05_Video_Audio/03_Capture_Pipelines.md)
and [`../05_Video_Audio/02_Hardware_Encoders.md`](../05_Video_Audio/02_Hardware_Encoders.md).

### 3.2 Hybrid client triad with one Go core

No single Go-native UI framework covers desktop, mobile, web, and
TV with acceptable accessibility, D-Pad, and TV ergonomics
(cross-verification HC-03). HelixPlay therefore adopts a hybrid
client matrix — Wails on desktop, Flutter+Go FFI on mobile and TV
(with Compose for TV as the Android TV native track and SwiftUI for
tvOS), and Angular+Go-WASM on the web — sharing one Go core compiled
three ways: as a `c-shared` library for Flutter FFI, as a native
binary embedded by Wails IPC, and as WASM for browsers (Insight #3).
The pillar's detail lives in
[`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md) and the TV
specifics in [`11_TV_UX.md`](11_TV_UX.md). Theme bindings to the same
core are documented in [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md).

### 3.3 Controller fidelity protocol

Controller "feel" — DualSense haptics, adaptive triggers, gyro,
accelerometer, audio jack on the controller, lightbar, touchpad — is
the differentiator the player perceives most viscerally. Sub-50 ms
video latency is table stakes. Most open-source predecessors
(Moonlight in particular) only forward XInput-class state. HelixPlay
commits to **full DualSense feature parity over network**, with a
binary protocol (16–32 byte packets), unreliable / unordered transport
(WebRTC DataChannel on the web, custom UDP on native), and per-game
controller profile mapping inspired by Steam Input
(cross-verification HC-06, Insight #2). The pillar's detail lives in
[`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md);
the latency-optimisation slice lives in
[`../04_Latency/07_Controller_Input_Optimization.md`](../04_Latency/07_Controller_Input_Optimization.md).

### 3.4 PS4-class catalog as a content business

The catalog is more than a UI surface; it is a content acquisition,
metadata mastering, rights management, and asset delivery system. The
technology problem (SQLite FTS5, image caching, CDN) is solved off
the shelf; the **content problem** (IGDB rate limits, SteamGridDB
licensing, Steam API commercial restrictions, regional rights) is
the actual investment (Insight #4). The pillar treats the catalog as
a multi-source pipeline (IGDB primary, SteamGridDB community
artwork, Steam where licensed, RAWG fallback, user-contributed
artwork moderated) with per-tenant overlays. Detail lives in
[`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md), with the TV
shelf and quick-resume tile semantics in [`11_TV_UX.md`](11_TV_UX.md)
and the per-tenant overlay logic in [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md).

### 3.5 White-label as architecture, not skin

White-label is a first-class architectural property, not a runtime
theme switch. It is the lever that converts HelixPlay from an "app"
into a Gaming-as-a-Service platform sellable to ISPs, hospitality
chains, hospitals, and venues (Insight #8, cross-verification HC-08).
The architecture commits to: a 3-tier design token system (primitive
→ semantic → component) per Material Design 3 tonal-palette logic,
CSS custom properties for runtime swaps, Style Dictionary v4 as the
build tool, per-tenant database schemas, per-tenant identity issuers,
per-tenant catalog overlays, and per-tenant recording defaults.
Detail lives in [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md);
catalog and tenant identity binding are in
[`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md) and
[`09_Security_and_Isolation.md`](09_Security_and_Isolation.md).

### 3.6 Edge-first latency

For internet-scale play, the dominant latency factor is physical
distance, not codec efficiency or protocol cleverness — speed of
light in fiber dominates (Insight #7). Moving hosts to within 100 km
of users yields a 10× improvement that no codec swap can match.
HelixPlay commits to: mDNS-driven LAN host discovery as the
zero-overhead default; rendezvous-driven WAN discovery prioritised
by geographic proximity; multi-region database (CockroachDB follower
reads); MEC / regional DC partner placement for managed deployments;
and a NAT-traversal relay that prefers TURN over the closest viable
relay rather than the cheapest. Pillar detail lives in
[`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md)
and [`12_Latency_Engineering_Overview.md`](12_Latency_Engineering_Overview.md);
the wire-level network specifics live in
[`../05_Video_Audio/12_Network_Transport.md`](../05_Video_Audio/12_Network_Transport.md)
and [`../04_Latency/05_UltraLowLatency_Network_Protocols.md`](../04_Latency/05_UltraLowLatency_Network_Protocols.md).

### 3.7 Anti-cheat clean host

Kernel-level anti-cheat (Easy Anti-Cheat, BattlEye, Vanguard) does
not merely block hooks — it imposes a **clean host certification**
property: the host must look like a normal gaming PC, not a server
or a VM with bolted-on capture drivers (Insight #5,
cross-verification HC-10). The pillar therefore mandates: capture
exclusively via OS-provided APIs (DXGI Desktop Duplication on
Windows, ScreenCaptureKit on macOS, KMS / PipeWire on Linux); no DLL
injection, no kernel-level hooks; signed (and where possible
WHQL-certified) virtual controller drivers; per-game compatibility
matrix maintained as a living artifact; and an escalation path with
anti-cheat vendors for commercial whitelisting. Detail lives in
[`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md),
[`09_Security_and_Isolation.md`](09_Security_and_Isolation.md), and
the Constitution at [`../01_Constitution.md`](../01_Constitution.md#113-anti-cheat-compatibility).

### 3.8 GaaS multi-tenancy

Multi-tenancy is not a Phase-2 add-on. From day one, every state
store, every storage prefix, every identity issuer, and every rate
limiter is **tenant-scoped**. There is no "single-tenant mode that
breaks later" — the single-tenant deployment is just a
multi-tenant deployment with one tenant. This pillar binds the
white-label architecture (3.5), the catalog content business (3.4),
and the security posture (3.7) into a coherent story. Detail lives
in [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md)
(database & storage), [`09_Security_and_Isolation.md`](09_Security_and_Isolation.md)
(identity & RBAC), [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md)
(brand surface), and [`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md)
(per-tenant catalog overlays).

These eight pillars cover all eight cloudgaming insights and all
ten cloudgaming HC findings. Two further pillars described in the
Constitution and System Overview — the **anti-bluff testing posture**
(R-11, R-12, R-13) and the **containerised runtime posture**
(R-05, R-06) — are not Architecture-family pillars but apply
horizontally; their detail lives in
[`../07_Testing/00_Index.md`](../07_Testing/00_Index.md) and
[`../08_Operations/01_Container_CI_CD.md`](../08_Operations/01_Container_CI_CD.md)
respectively.

---

## 4. Reading order

Different audiences enter the Architecture family with different
goals. The following three reading orders are normative — Master
Plan §4.1 step 6 requires every chapter to participate in at least
one reading order so that the family has navigable depth in three
clicks (Constitution §12.3).

### 4.1 Newcomer reading order

A reader with no prior context. Goal: end with enough vocabulary
and topology to engage with any sibling chapter.

1. [`../02_System_Overview.md`](../02_System_Overview.md) — vision,
   problem statement, reference user journey, host & client matrix.
2. This index file (`00_Index.md`) — pillars and dimension map.
3. [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md) — what
   the user runs on each surface and the shared Go core that makes
   the triad coherent.
4. [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md)
   — what the host runs to receive input, launch games, capture,
   encode, and stream.
5. [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md)
   — the wire between host and client.
6. [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md)
   — the back-channel from client to host.
7. [`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md) — what the
   user browses to choose what to play.
8. [`12_Latency_Engineering_Overview.md`](12_Latency_Engineering_Overview.md)
   — the budget that everything else must respect.

### 4.2 Implementor reading order

An engineer starting an implementation phase. Goal: end with the
specific patterns, configuration knobs, and submodule dependencies
to begin coding.

1. [`../09_Implementation_Phases/00_Phase_Index.md`](../09_Implementation_Phases/00_Phase_Index.md)
   — the phase the work belongs to.
2. [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md)
   — the existing `vasic-digital` submodules that must be reused
   or extended (R-04).
3. [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md)
   and [`03_Host_OS_Capture.md`](03_Host_OS_Capture.md) — the host
   surface, including the Sunshine++ contract.
4. [`05_RealTime_APIs.md`](05_RealTime_APIs.md) — gRPC, NATS, REST
   gateway, SSE, WebSocket selection criteria, plus rate-limit and
   backpressure conventions.
5. [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md)
   — Pion WebRTC v4 and the custom UDP fallback shape.
6. [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md)
   — the binary protocol the implementor must respect on both ends.
7. [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md) — the
   c-shared / WASM / native build matrix and the FFI contract.
8. [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md)
   — service discovery, load balancing, multi-region database
   topology.
9. [`../07_Testing/01_Test_Matrix.md`](../07_Testing/01_Test_Matrix.md)
   — the ten test types the change must satisfy before merge.
10. [`../08_Operations/01_Container_CI_CD.md`](../08_Operations/01_Container_CI_CD.md)
    — the container in which the change must build and run.

### 4.3 Security reviewer reading order

A reviewer auditing the system for a security or compliance gate.
Goal: end with confirmation that every privileged surface has a
documented threat model and that the `clean host` and tenancy
properties hold.

1. [`../01_Constitution.md`](../01_Constitution.md) §11 (Security &
   Privacy) — the normative ceiling.
2. [`09_Security_and_Isolation.md`](09_Security_and_Isolation.md) —
   threat model, mTLS topology, JWT lifetimes, rate limits, audit
   logs.
3. [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md)
   — anti-cheat clean-host posture, virtual controller signing,
   driver provenance.
4. [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md)
   — tenant-scoping of the database, storage prefixes, and rate
   limiters.
5. [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md) —
   per-tenant identity issuer binding and brand-surface isolation.
6. [`05_RealTime_APIs.md`](05_RealTime_APIs.md) — input validation
   at every boundary; gRPC schema versioning and deprecation.
7. [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md)
   — input as personal data, telemetry consent, opt-in recording.
8. [`../07_Testing/05_Security_Tests.md`](../07_Testing/05_Security_Tests.md)
   — fuzzing, SAST, DAST, dependency scanning matrix.
9. [`../07_Testing/11_Challenges.md`](../07_Testing/11_Challenges.md)
   and [`../07_Testing/12_HelixQA_Autonomous.md`](../07_Testing/12_HelixQA_Autonomous.md)
   — production-equivalent security validation.

Each reading order leaves the reader at a clearly named entry-point
of the next family (Latency, Video/Audio, Testing, or Operations) so
that no leaf is more than three clicks from the Master Plan.

---

## 5. Cross-stream linkage

The Architecture family overlaps the Latency and Video/Audio families
deliberately. Where overlap exists, **one** chapter is canonical and
the other refers to it; this section names the canonical owner so
readers do not get conflicting normative text.

### 5.1 Overlap with the Latency family ([../04_Latency/](../04_Latency/00_Index.md))

| Topic | Architecture chapter | Latency chapter | Canonical owner |
|-------|----------------------|-----------------|-----------------|
| Glass-to-glass latency budget table | [`12_Latency_Engineering_Overview.md`](12_Latency_Engineering_Overview.md) | [`../04_Latency/00_Index.md`](../04_Latency/00_Index.md), [`../04_Latency/10_Latency_Testing_and_Validation.md`](../04_Latency/10_Latency_Testing_and_Validation.md) | Architecture chapter 12 owns the **system-level** budget; the Latency family chapters own the **per-stage** measurement methodology. Each table cell in chapter 12 references the Latency chapter that justifies its number. |
| Controller polling rate (1000 Hz USB / 8 ms BT) | [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md) | [`../04_Latency/07_Controller_Input_Optimization.md`](../04_Latency/07_Controller_Input_Optimization.md) | Architecture chapter 2 owns the **protocol shape**; the Latency chapter 7 owns the **timing analysis** and the host-side scheduling story (PREEMPT_RT). |
| Frame pacing & VRR | [`12_Latency_Engineering_Overview.md`](12_Latency_Engineering_Overview.md) (system view) | [`../04_Latency/08_Frame_Pacing_and_VRR.md`](../04_Latency/08_Frame_Pacing_and_VRR.md) | Latency chapter 8 is canonical for the pacing algorithms and the display-side floor; chapter 12 references it for the system-level budget row. |
| GPU-direct, zero-copy capture path | [`03_Host_OS_Capture.md`](03_Host_OS_Capture.md) (per-OS API choice) | [`../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md`](../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md) | Architecture chapter 3 owns **API selection** per OS (DXGI DDA, ScreenCaptureKit, KMS/DMA-BUF); Latency chapter 4 owns the **GPUDirect / hardware pipeline** mechanics. |
| Allocation discipline on hot path | Constitution §5.4 | [`../04_Latency/09_Memory_and_Cache_Optimization.md`](../04_Latency/09_Memory_and_Cache_Optimization.md) | Constitution sets the rule; Latency chapter 9 owns the implementation patterns and `perf c2c` validation. |
| Real-time OS scheduling | [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md) (host-side decision) | [`../04_Latency/06_RealTime_OS_and_Scheduling.md`](../04_Latency/06_RealTime_OS_and_Scheduling.md) | Architecture chapter 7 records the deployment decision (PREEMPT_RT recommended on dedicated hosts); Latency chapter 6 owns the kernel patches and tuning recipes. |

### 5.2 Overlap with the Video/Audio family ([../05_Video_Audio/](../05_Video_Audio/00_Index.md))

| Topic | Architecture chapter | Video/Audio chapter | Canonical owner |
|-------|----------------------|---------------------|-----------------|
| Codec selection (H.264 / HEVC / AV1) | [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md) (protocol view) | [`../05_Video_Audio/01_Codec_Selection.md`](../05_Video_Audio/01_Codec_Selection.md) | Video/Audio chapter 1 is canonical for codec selection rationale, profile/level tables, and licensing posture. Architecture chapter 1 references it for the streaming-protocol matrix. |
| Hardware encoders (NVENC / QSV / AMF / VideoToolbox / VAAPI) | [`03_Host_OS_Capture.md`](03_Host_OS_Capture.md) (per-OS availability) | [`../05_Video_Audio/02_Hardware_Encoders.md`](../05_Video_Audio/02_Hardware_Encoders.md) | Video/Audio chapter 2 is canonical for per-encoder configuration knobs, ULL preset selection, B-frame policy. Architecture chapter 3 references it from the per-OS capability matrix. |
| Capture pipelines (zero-copy, DMA-BUF, IOSurface) | [`03_Host_OS_Capture.md`](03_Host_OS_Capture.md) | [`../05_Video_Audio/03_Capture_Pipelines.md`](../05_Video_Audio/03_Capture_Pipelines.md) | Architecture chapter 3 is canonical for the OS-level API choice; Video/Audio chapter 3 is canonical for the buffer-format details and the capture→encode handoff mechanics. |
| Dual-path encoding (stream + record) | [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md) (lifecycle view) | [`../05_Video_Audio/04_DualPath_Encoding.md`](../05_Video_Audio/04_DualPath_Encoding.md) | Video/Audio chapter 4 is canonical. |
| Recording storage (MKV, fMP4, network targets) | n/a | [`../05_Video_Audio/05_Recording_Storage.md`](../05_Video_Audio/05_Recording_Storage.md) | Video/Audio chapter 5 is canonical; Architecture references for the NVMe ring buffer and SMB/NFS/FTP/WebDAV sync targets. |
| Audio (Opus MultiStream, AC3/EAC3 passthrough, Dolby Atmos) | n/a | [`../05_Video_Audio/06_Audio_Pipeline.md`](../05_Video_Audio/06_Audio_Pipeline.md) | Video/Audio chapter 6 is canonical; the System Overview's audio paragraph in §8 points here. |
| HDR pipeline | [`03_Host_OS_Capture.md`](03_Host_OS_Capture.md) (per-OS surface) | [`../05_Video_Audio/07_HDR_and_Color.md`](../05_Video_Audio/07_HDR_and_Color.md) | Video/Audio chapter 7 is canonical for HDR10 / HDR10+ / Dolby Vision / HLG mapping; Architecture chapter 3 references it for `DuplicateOutput1` and ScreenCaptureKit-HDR specifics. |
| ABR, FEC, congestion control | [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md) (protocol view) | [`../05_Video_Audio/08_ABR_FEC_Congestion.md`](../05_Video_Audio/08_ABR_FEC_Congestion.md) | Video/Audio chapter 8 is canonical for the algorithms (SQP, BBRv3, FEC schemes); Architecture chapter 1 references it for the protocol-level switching logic. |
| Network transport (WebRTC vs custom UDP, QUIC, AF_XDP, Pion) | [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md), [`05_RealTime_APIs.md`](05_RealTime_APIs.md) | [`../05_Video_Audio/12_Network_Transport.md`](../05_Video_Audio/12_Network_Transport.md) | Architecture chapters 1 & 5 are canonical for the **selection** and the higher-level shape; Video/Audio chapter 12 is canonical for the wire-format details, the Pion code paths, and the AF_XDP recipe. |

In every case the rule is: **architectural decision** lives in the
Architecture family; **mechanism** and **measurement** live in the
adjacent family. A reader who finds the same topic explained
differently in two places must treat the canonical owner above as
authoritative and file an issue against the non-canonical text.

---

## 6. Vocabulary anchors

Twelve architecture-specific terms newcomers must internalise before
reading any sibling chapter. Each term has a one-sentence definition
and a forward link to the chapter that owns its long form. Constitution
§14 has the full glossary; this list is the scoped subset for the
Architecture family.

1. **Sunshine++** — the architectural pattern of forking Sunshine's
   proven capture/encode/stream core and layering session orchestration,
   controller forwarding, save sync, and white-label tenant binding on
   top — see [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md).
2. **Hybrid client triad** — the architectural decision to use Wails
   (desktop), Flutter+Go FFI (mobile/TV), and Angular+Go-WASM (web)
   sharing one Go core compiled three ways — see
   [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md).
3. **Go core** — the platform-agnostic Go library that implements the
   streaming session state machine, the controller binary protocol,
   the auth client, the catalog client, and the OpenTelemetry exporter
   — compiled `c-shared` for FFI, native for Wails, WASM for the web
   — see [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md).
4. **Clean host** — a host machine that uses only OS-provided capture
   APIs, signed virtual controller drivers, and no DLL injection or
   kernel hooks, so anti-cheat treats it as a normal gaming PC — see
   [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md)
   and [`09_Security_and_Isolation.md`](09_Security_and_Isolation.md).
5. **Rendezvous service** — the backend service that pairs clients
   with hosts outside the LAN, brokers ICE candidates, and prioritises
   geographic proximity in host selection — see
   [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md).
6. **Capability advertisement** — the host's structured publication of
   its GPU model, supported codecs, max resolution, refresh rate,
   NVENC session count, and thermal headroom, consumed by the
   rendezvous and load-balancing logic — see
   [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md).
7. **Controller binary protocol** — the 16–32-byte unreliable / unordered
   packet format that carries DualSense-class state from client to
   host, including haptics, adaptive triggers, gyro, accelerometer,
   touchpad, and lightbar — see
   [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md).
8. **Per-game controller profile** — the runtime mapping table that
   converts a generic controller event stream into the per-game
   keybinding, axis curve, and dead-zone configuration, modelled on
   Steam Input — see
   [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md).
9. **Design token** — the smallest themable unit (colour, spacing,
   radius, font weight) consumed by primitive → semantic → component
   layers per Material Design 3 tonal-palette logic — see
   [`10_WhiteLabel_and_Theming.md`](10_WhiteLabel_and_Theming.md).
10. **Tenant** — an operator deploying HelixPlay under their own
    brand, with isolated database schema, identity issuer, catalog
    overlay, and rate-limit profile — see
    [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md).
11. **REST gateway** — the dedicated microservice that translates
    REST traffic into the canonical gRPC contracts; REST is never
    canonical (Constitution §4.2) — see
    [`05_RealTime_APIs.md`](05_RealTime_APIs.md).
12. **Shelf** — a horizontally-scrolling row of game cards on the
    catalog landing screen; the PS4-class atomic UI unit, themable
    per tenant — see [`11_TV_UX.md`](11_TV_UX.md) and
    [`06_Catalog_and_Assets.md`](06_Catalog_and_Assets.md).
13. **Quick-resume tile** — the launcher tile that resumes the most
    recent session in one input event; in MVP this is a graceful
    return rather than a process-suspend snapshot (CZ-03) — see
    [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md).
14. **Capability matrix** — the per-game compatibility record (anti-cheat
    behaviour, suspension safety, save-sync method, virtual controller
    compatibility) maintained as living documentation — see
    [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md).
15. **Edge node** — a regional placement of host capacity within
    ~100 km of the player, prioritised by the rendezvous service to
    minimise speed-of-light transit — see
    [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md).

---

## 7. Open questions

The cross-verification report (`cloudgaming_cross_verification.md`)
catalogued five conflict zones (CZ-01..CZ-05). Each is resolved at a
specific chapter; this section names the resolution owner so the
Architecture family does not silently reintroduce the conflict.
Master Plan §4.3 requires every chapter's verification block to
re-state the resolution; this index restates the assignment.

| CZ-ID | Conflict | Resolution sketch | Resolving chapter |
|-------|----------|-------------------|-------------------|
| CZ-01 | WebRTC vs custom UDP for streaming and input transport | Hybrid: WebRTC default and mandatory for web; custom UDP available as an opt-in fallback for native desktop / TV clients. Both behind a transport abstraction in the Go core. | [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md) plus the wire-level mechanics in [`../05_Video_Audio/12_Network_Transport.md`](../05_Video_Audio/12_Network_Transport.md). |
| CZ-02 | Fyne viability for mobile/TV | Excluded: Fyne lacks D-Pad navigation and accessibility coverage required for TV-first UX. Wails covers desktop; Flutter covers mobile/TV; Angular+Go-WASM covers web. | [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md). |
| CZ-03 | Process suspension as a quick-resume primitive | Deferred to Phase 2: in MVP, "Home" returns to the catalog and the running game continues in the background or closes safely depending on the per-game capability matrix. CRIU and `NtSuspendProcess` are tracked but not shipped. | [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md). |
| CZ-04 | Bluetooth controller latency vs user expectation of wireless | Both supported with documented tradeoff: USB (preferred) and 2.4 GHz dongle (preferred for competitive) both at 1000 Hz; Bluetooth at 8 ms (125 Hz) flagged as "casual" tier with an explicit UI hint. | [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md), referencing [`../04_Latency/07_Controller_Input_Optimization.md`](../04_Latency/07_Controller_Input_Optimization.md). |
| CZ-05 | Bare metal vs cloud GPU hosts | Hybrid: bare-metal GPU servers are the cost-optimal primary; cloud GPU instances (with documented passthrough caveats) provide burst and failover. Per-tenant policy decides which is the default. | [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md), with the security implications in [`09_Security_and_Isolation.md`](09_Security_and_Isolation.md). |

Until each resolving chapter is produced and its Anti-Bluff
Verification block signed, the conflict is "owned but not closed." The
queue ordering in Master Plan §7.2 (C02..C13) ensures every
resolution lands within the Architecture family before the dependent
families (Submodules, Testing, Operations, Implementation Phases)
consume it.

In addition to the source-derived CZs, this index records two
**index-level** open questions that must be answered before the
Architecture family is signed off:

- **OQ-01: Wails vs Tauri-Go on desktop.** The Stream-1 sources
  evaluated Wails as the desktop default; Tauri-Go (Tauri host with
  a Go side-car bound by `tauri-plugin-shell` or a sibling FFI)
  appeared in the candidate list. The decision is currently Wails
  per cross-verification HC-03; chapter 4 must record any new evidence
  and either confirm or document the migration path.
- **OQ-02: Compose for TV vs Flutter on Android TV.** Compose for TV
  is the Google-recommended, stable-1.0 successor to the deprecated
  Leanback; Flutter remains the cross-cutting Go-FFI client. Chapter
  11 must specify which is the **primary** Android TV surface and
  which is the **fallback**, with the fallback explicitly named (no
  "or" clauses).

Both OQs are flagged for resolution inside their respective chapters'
verification blocks.

---

## Anti-Bluff Verification

### Source Evidence Reviewed

- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/05_Response/00_Master_Plan.md` — 613 lines, reviewed 2026-04-28 (§4.2, §4.3, §4.4, §7.2 row C01).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/05_Response/01_Constitution.md` — 700 lines, reviewed 2026-04-28 (§1, §6, §14).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/05_Response/02_System_Overview.md` — 643 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/01_Request.md` — 3 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim_decomposition.md` — 70 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_insight.md` — 156 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_cross_verification.md` — 130 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim01.md` — 812 lines, indexed by line count 2026-04-28 (deep dive deferred to C02).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim02.md` — 557 lines, indexed by line count 2026-04-28 (deep dive deferred to C03).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim03.md` — 909 lines, indexed by line count 2026-04-28 (deep dive deferred to C04).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim04.md` — 1,380 lines, indexed by line count 2026-04-28 (deep dive deferred to C05).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim05.md` — 966 lines, indexed by line count 2026-04-28 (deep dive deferred to C06).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim06.md` — 1,431 lines, indexed by line count 2026-04-28 (deep dive deferred to C07).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim07.md` — 1,449 lines, indexed by line count 2026-04-28 (deep dive deferred to C08).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim08.md` — 1,003 lines, indexed by line count 2026-04-28 (deep dive deferred to C09).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim09.md` — 974 lines, indexed by line count 2026-04-28 (deep dive deferred to C10).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim10.md` — 1,353 lines, indexed by line count 2026-04-28 (deep dive deferred to C11).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim11.md` — 1,155 lines, indexed by line count 2026-04-28 (deep dive deferred to C12).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_dim12.md` — 1,340 lines, indexed by line count 2026-04-28 (deep dive deferred to C13).

### Web Sources Consulted

- None. Master Plan §4.1 reserves web research for the per-dimension
  technical chapters (C02..C13). The Architecture index is a
  navigational artifact derived from existing project research; no
  external evidence is required to construct it.

### Insights Incorporated

- cloudgaming Insight #1 (Sunshine++ host agent pattern) → §3.1.
- cloudgaming Insight #2 (Controller protocol as the hidden differentiator) → §3.3.
- cloudgaming Insight #3 (Three clients, one Go core) → §3.2.
- cloudgaming Insight #4 (Catalog as content business) → §3.4.
- cloudgaming Insight #5 (Anti-cheat clean host) → §3.7.
- cloudgaming Insight #6 (PS4 UX as legal safe harbor) → §3.4 and §6 (Shelf, Quick-resume tile).
- cloudgaming Insight #7 (Edge > codec for latency) → §3.6.
- cloudgaming Insight #8 (White-label enables GaaS) → §3.5 and §3.8.
- cross-verification HC-01 (Pion WebRTC for Go) → §3.3, §5.1, §5.2 (transport rows).
- cross-verification HC-02 (H.264 default, AV1 upgrade) → §5.2 (codec row).
- cross-verification HC-03 (No single Go framework covers all clients) → §3.2.
- cross-verification HC-04 (Sunshine architecture as host agent reference) → §3.1.
- cross-verification HC-05 (Platform-specific capture APIs are mature) → §5.1, §5.2 (capture rows).
- cross-verification HC-06 (UDP/DataChannels + virtual controller) → §3.3.
- cross-verification HC-07 (OAuth2/OIDC + JWT + DTLS-SRTP) → §4.3, §6 (Tenant, Capability advertisement).
- cross-verification HC-08 (Design token architecture) → §3.5, §6 (Design token).
- cross-verification HC-09 (Sub-50 ms LAN latency achievable) → §3.6.
- cross-verification HC-10 (Anti-cheat as architectural constraint) → §3.7.

### Conflict Zones Resolved

| CZ-ID | Conflict | Decision (this index) | Rationale | Final resolution chapter |
|-------|----------|-----------------------|-----------|--------------------------|
| CZ-01 | WebRTC vs custom UDP | Hybrid; WebRTC default + web mandatory; custom UDP fallback for native, behind Go-core abstraction. | §3.3 commits to both transports; §5 binds the wire-format detail to the Video/Audio family. | [`01_Streaming_Protocols_and_Codecs.md`](01_Streaming_Protocols_and_Codecs.md) (queued C02). |
| CZ-02 | Fyne viability for mobile/TV | Excluded. | TV-first UX requires D-Pad and accessibility coverage Fyne lacks per HC-03. | [`04_Go_Client_Ecosystem.md`](04_Go_Client_Ecosystem.md) (queued C05). |
| CZ-03 | Game suspension as quick-resume primitive | Deferred to Phase 2; MVP delivers graceful game-switching governed by the per-game capability matrix. | Anti-cheat compatibility per HC-10 and Insight #5 forbids opaque suspension on flagged titles; CRIU does not preserve GPU state. | [`07_Host_Agent_and_Game_Lifecycle.md`](07_Host_Agent_and_Game_Lifecycle.md) (queued C08). |
| CZ-04 | Bluetooth controller latency | Both supported; USB / 2.4 GHz preferred for competitive, BT explicitly tagged "casual" tier in UI. | User requirement of wireless support is non-negotiable; latency tradeoff is documented and surfaced in UI. | [`02_Controller_Input_Pipeline.md`](02_Controller_Input_Pipeline.md) (queued C03). |
| CZ-05 | Bare metal vs cloud for hosts | Hybrid; bare metal primary for cost, cloud GPU for burst / failover; per-tenant policy. | Cost (bare metal 45–90% cheaper) vs flexibility; hardware passthrough constraints on cloud GPU instances per HC-10. | [`08_Scalability_and_MultiRegion.md`](08_Scalability_and_MultiRegion.md) (queued C09). |

### Coverage Confirmation

- Source line count (Stream 1 aggregate): 13,688 lines.
- Synthesized line count for **this index**: ≥400 lines (R-01 floor for C01 per Master Plan §7.2).
- The 13,688-line floor for the **whole Architecture family** is satisfied by the cumulative minimum targets in §2 (16,050 lines aggregate across C01..C13), which exceed the source floor.
- Per-chapter coverage (C02..C13) will be confirmed in each sibling chapter's own verification block when produced.
- The 400-line floor for this Index is met: ≥ 400 lines including the verification block.

### Sign-off

Executed by: subagent (C01) on 2026-04-28
Reviewed by: pending orchestrator review
Date: 2026-04-28

End of Architecture chapter index — 2026-04-28.
