# Frame Pacing & VRR

> **Source dimensions:**
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/01_Request.md`.
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_dim08.md` — 91 lines (primary per-dim source).
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md` — 2,199 lines (long-form synthesis — VRR + frame-pacing sections).
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_insight.md` — **Insight #3** (Asymmetric optimisation — HOST optimises input-to-render; CLIENT optimises receive-to-display — this chapter is **CLIENT-side**), **Insight #5** (Conservative Prediction Paradox — bears on frame-interpolation policy; default OFF).
> - `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_cross_verification.md` — **HC-03** (Reflex + Frame Warp ≈ 75% perceived-latency reduction — host-side; this chapter's frame-interpolation receive-side complement), **HC-09** (VRR < 1 ms display-side cost — binding 2024 baseline; 2026 evidence reaffirms).
> - Web research addendum: [`../99_Web_Research_Addenda/2026-04-29-frame-pacing-and-vrr.md`](../99_Web_Research_Addenda/2026-04-29-frame-pacing-and-vrr.md) — 610 lines, 120 distinct URLs across 9 clusters (§A NVIDIA G-Sync 2026 status, §B AMD FreeSync 2026 status, §C HDMI 2.1 VRR + Adaptive Sync, §D HDMI 2.1 ALLM, §E LFC + frame doubling, §F frame interpolation client-side — DLSS-G / FSR-G / XeSS-G, §G Adaptive VSync + FastSync + tear-free presentation, §H display timing — VSYNC + buffer-flip + scanout, §I 120/144/240/360 Hz feasibility 2026 + 4K240/4K480 OLED panels) plus §Z contradictions index Z1..Z8.
>
> **Source line floor for R-01 (per Master Plan §7.2 row C22):** 250 lines of body prose. **Achieved:** see Anti-Bluff Verification block.
>
> **Chapter targets:** R-01, R-02, R-03 (new public submodule `vasic-digital/helix-display`), R-04 (DRY — `r18.SafeExec` inherited from C08 §10; `host-integrity-scan` inherited from C08 §12.11), R-08, R-09, R-10, R-11, R-12, R-13, **R-18 §11.5 Operational Integrity** (`xrandr` / Windows display API / macOS `system_profiler` / `CGDisplayConfigure` all wrap through the inherited `r18.SafeExec` on Wails desktop tier; web/mobile/TV use OS-native APIs without subprocess).
>
> **Cross-links:**
> - Master Plan: [`../../00_Master_Plan.md`](../../00_Master_Plan.md). Constitution: [`../../01_Constitution.md`](../../01_Constitution.md) (§1.1 forbidden-pattern list; §6 quality — p50/p99/p999 ≥ 10 K samples; §11.5 R-18). System Overview: [`../../02_System_Overview.md`](../../02_System_Overview.md) (§9 latency budget — display-side floor cited).
> - Latency family index: [`00_Index.md`](00_Index.md).
> - Architecture-side overview: [`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md) (C13 §3 frame-time + §7 frame interpolation + §8 VRR/G-Sync/FreeSync/ALLM + §9 120/144/240Hz feasibility — this chapter is the client-side display-sync elaboration).
> - Sibling Latency chapters: [`04_GPU_Direct_and_Hardware_Pipelines.md`](04_GPU_Direct_and_Hardware_Pipelines.md) (C18 §5 — Reflex 2 / Anti-Lag 2 / XeLL host-side; this chapter §4.5 is the client-side complement), [`07_Controller_Input_Optimization.md`](07_Controller_Input_Optimization.md) (C21 §3 + §4 — Conservative Prediction Paradox cross-link from §4.5), [`10_Latency_Testing_and_Validation.md`](10_Latency_Testing_and_Validation.md) (C24 — display-side measurement harness; cross-link §8.5).
> - Sibling Architecture chapters: [`../03_Architecture/04_Go_Client_Ecosystem.md`](../03_Architecture/04_Go_Client_Ecosystem.md) (Wails / Flutter / Compose for TV — client-tier UI), [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) (§10 `r18.SafeExec` inheritance; §1 capability-schema delta cross-link from §6.2; §12.11 host-integrity-scan inheritance), [`../03_Architecture/08_Scalability_and_MultiRegion.md`](../03_Architecture/08_Scalability_and_MultiRegion.md) (§3 capability-based admission `display.vrr_supported` + `display.refresh_rates` predicates), [`../03_Architecture/11_TV_UX.md`](../03_Architecture/11_TV_UX.md) (C12 §6 — TV-side ALLM / HDMI-CEC primitives — this chapter elaborates display-sync timing with TV-side cross-link).
> - Operations / Testing / Phases queued. Notably: [`../09_Implementation_Phases/Phase_12_Latency_Tuning.md`](../09_Implementation_Phases/Phase_12_Latency_Tuning.md) is the primary implementation phase; [`../07_Testing/02_Benchmarking_Tests.md`](../07_Testing/02_Benchmarking_Tests.md) imports the §8.5 harness.
>
> **Status:** Draft v1 — section-stitched assembly (R1 recovery model). Awaiting operator review.
>
> **Last updated:** 2026-04-29.

This chapter is the **eighth deep chapter of the `04_Latency/`
family** — the **client-side display-synchronisation layer**
(Insight #3 asymmetric optimisation: HOST owns input-to-render;
CLIENT owns receive-to-display). It owns the full client-side
VRR posture (G-Sync / FreeSync / HDMI 2.1 VRR / ALLM), frame
pacing + tear-free presentation (full-screen exclusive vs
windowed, Adaptive VSync vs FastSync), client-side frame
interpolation (DLSS-G / FSR-G / XeSS-G — opt-in only per
Conservative Prediction Paradox + 1-frame latency penalty), and
display timing primitives (VSYNC + buffer-flip + scanout, OLED
vs LCD response, 4K240 / 4K480 OLED 2026 panels).

The chapter establishes that **VRR-in-streaming is HelixPlay's
codified differentiator** (cross-link C13 Z5): open-source
streaming clients still treat VRR as a gap; HelixPlay codifies
VRR end-to-end as an MVP differentiator. The client-tier UI
negotiates VRR over whatever cable the user has (HDMI 2.1 VRR,
DisplayPort Adaptive Sync, G-Sync Compatible), reports capability
to the host-agent, and the host-agent picks stream rate ≤
min(host render rate, display rate). LFC (Low Framerate
Compensation) absorbs transient frame-rate drops below the VRR
range; ALLM signals through HDMI-CEC to TV-class displays for
50-100 ms image-processing-bypass savings (cross-link C12 §6).

**HC-03 reaffirmed (host-side, cross-link C18 §5); HC-09
reaffirmed (display-side cost confirmed by HDMI Forum + VESA +
NVIDIA 2026 measurements)** per the addendum's eight
contradictions:

- **VRR range expansion** to 480/500 Hz top + 1 Hz bottom on
  2026 OLED panels (addendum Z1) — chapter §2.7 documents.
- **Frame Generation latency cost** is NOT zero (contradicts 2024
  marketing claims); 2026 measurements show ~16 ms penalty per
  frame-gen tier (addendum Z2) — chapter §4 documents the trade-
  off + Conservative Prediction Paradox enforcement.
- **Sunshine/Moonlight VRR partial** (addendum Z3) — open-source
  streaming clients still struggle with VRR; HelixPlay codifies
  end-to-end as the differentiator (cross-link C13 Z5).
- **OLED VRR brightness flicker** (addendum Z4) — flicker
  observed on some OLED panels at low VRR rates (< 60 Hz);
  HelixPlay's encode pipeline targets ≥ 60 fps with LFC as the
  safety net.
- **HDMI 2.2 Ultra96 96 Gbps** (addendum Z5) — HelixPlay marks
  HDMI 2.2 as V1 deferral; MVP scope is HDMI 2.1.
- **DisplayPort 2.1 UHBR20 panel availability** (addendum Z6) —
  4K480 + 8K120 panels with UHBR20; HelixPlay treats UHBR20 as
  V1 (capability-schema-future-proofed but not deployed).
- **G-Sync Pulsar variable-rate strobing** (addendum Z7) — 2026
  G-Sync feature reducing motion blur via frame-synchronous
  strobing; HelixPlay capability-advertises but doesn't gate.
- **Wayland/Gamescope multi-monitor VRR brokenness** (addendum
  Z8) — chapter §7 F12 documents; Steam OS Gamescope and KWin
  6.0+ are the known-good options.

The chapter introduces and resolves **eight addendum-defined
contradictions** (cite addendum §Z):

- **Z1** — VRR range expansion to 480/500 Hz top + 1 Hz bottom —
  chapter §2.7.
- **Z2** — Frame-gen latency cost ~16 ms (not zero) — chapter §4.
- **Z3** — Sunshine/Moonlight VRR partial (cross-link C13 Z5) —
  chapter §1.
- **Z4** — OLED VRR brightness flicker — chapter §2.6 LFC safety
  net + §7 F4 mitigation.
- **Z5** — HDMI 2.2 Ultra96 96 Gbps V1 deferral — chapter §1.2.
- **Z6** — DisplayPort 2.1 UHBR20 panel availability — chapter
  §5.4 + V1 deferral.
- **Z7** — G-Sync Pulsar variable-rate strobing — chapter §2.2.
- **Z8** — Wayland/Gamescope multi-monitor VRR brokenness —
  chapter §7 F12 + OQ-C22-04.

The chapter **inherits without re-implementing** (Constitution §2 DRY):

- The `r18.SafeExec` wrapper from [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) §10 — used in §6 Implementation contract for `xrandr` / `system_profiler` / Windows registry-read invocations on Wails desktop tier.
- The `host-integrity-scan` test from [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) §12.11 — inherited by §8.11 of this chapter (non-overridable per Constitution §11.5.4).
- The Reflex 2 / Anti-Lag 2 / XeLL host-side capability surface from C18 §5 — this chapter's §4.5 is the client-side complement.
- The TV-side ALLM / HDMI-CEC primitives from [`../03_Architecture/11_TV_UX.md`](../03_Architecture/11_TV_UX.md) §6 — this chapter cross-links from §2.5.
- All inherited CZs/OQs from prior chapters — not relitigated.

---

## Table of Contents

- [§1 Scope & non-scope](#1-scope--non-scope)
- [§2 VRR fundamentals (G-Sync / FreeSync / HDMI 2.1 / ALLM)](#2-vrr-fundamentals-g-sync--freesync--hdmi-21--allm)
- [§3 Frame pacing + tear-free presentation](#3-frame-pacing--tear-free-presentation)
- [§4 Client-side frame interpolation (DLSS-G / FSR-G / XeSS-G)](#4-client-side-frame-interpolation-dlss-g--fsr-g--xess-g)
- [§5 Display timing synchronisation](#5-display-timing-synchronisation)
- [§6 Implementation contract](#6-implementation-contract)
- [§7 Failure modes](#7-failure-modes)
- [§8 Test surface](#8-test-surface)
- [§9 Open questions](#9-open-questions)
- [§10 References](#10-references)
- [Anti-Bluff Verification](#anti-bluff-verification)

---

## 1. Scope & non-scope

### 1.1 Place in the Latency family

C22 — *Frame Pacing and VRR* — is the **eighth deep chapter of the
Latency family**, sitting after C21 (*Controller Input Optimization*)
and consolidating the **client-side display synchronisation layer** of
the HelixPlay pipeline. Where C18 (Reflex / Anti-Lag / XeLL) and the
related GPU-Direct chapter optimise everything from the controller
interrupt to the encoder bitstream on the **host**, this chapter
optimises everything from the decoder output to the photons that leave
the panel on the **client**. The split is not arbitrary. **Insight #3
(Asymmetric Optimisation)** in `latency_insight.md` is binding: "the
host optimises input-to-render (I2FS + FS2P), while the client
optimises receive-to-display (decode + VRR + scanout). A symmetric
optimisation strategy (same techniques on both sides) is suboptimal."
This chapter inherits Insight #3 verbatim and uses it to gate
admission of every technique, every metric, and every R-18 boundary
called out below.

### 1.2 Binding 2024 baseline + 2026 reaffirmations

Two High-Confidence findings bracket this chapter end-to-end:

- **HC-09** — *VRR (G-Sync / FreeSync) adds < 1 ms while eliminating
  tearing.* Source: NVIDIA G-SYNC technical documentation cited in
  `latency_dim08.md` §2 ("G-SYNC adds minimal latency — less than 1 ms
  compared to fixed refresh rate."). 2024 baseline; 2026 web research
  (C13's `2026-04-28-latency-engineering-overview.md` §C, including
  the NVIDIA G-SYNC 10K-sample p99 paper) reaffirms with two
  refinements: (a) the < 1 ms figure is a *mean*, not a p999, and the
  real value of VRR is in the tail, not the average; (b) the figure
  applies to monitors and panel-grade TVs but does NOT apply when the
  TV's image-processing pipeline is engaged (motion smoothing, HDR
  tone-mapping) — in which case ALLM is the prerequisite for HC-09 to
  hold.

- **HC-03** — *NVIDIA Reflex + Frame Warp delivers ~75 % perceived-
  latency reduction.* HC-03 is host-side (C18 §5 owns it). This
  chapter owns the **receive-side complement**: client-side frame
  interpolation (DLSS-G client-side, FSR-G, XeSS-G) plus jitter
  buffer plus VRR plus scanout alignment. C22 and C18 together
  realise the asymmetric playbook of Insight #3.

### 1.3 Cross-links to sibling chapters

- **C12 §6** (TV UX) elaborates the **TV-side** ALLM assertion +
  HDMI-CEC + overscan-safe area negotiation. C22 elaborates the
  **display-sync timing** of the same primitives — ALLM is the
  prerequisite for VRR's < 1 ms cost on consumer TVs, and §2.5
  below specifies the negotiation timing while C12 specifies the UI
  surface.
- **C13 §8 + §9** (Latency Engineering Overview) identifies VRR-in-
  streaming as the **Z5 differentiator** ("open-source streaming
  clients still treat VRR as a gap; HelixPlay codifies VRR end-to-
  end as MVP differentiator"). C22 is the implementation chapter that
  owns the Z5 deliverable: every VRR mechanism prescribed in C13 §8
  is specified in operational detail in this chapter.
- **C18 §5** (Reflex / Anti-Lag / XeLL) — host-side fork of the
  vendor-neutral input-to-render reduction. Out of scope here.
- **C21** (Controller Input Optimization) — input-side; out of scope.
- **C18 §2** (GPU-Direct) — host-side zero-copy; out of scope.

### 1.4 In scope

This chapter specifies, in operational detail: (a) **VRR primitives**
— G-Sync (Ultimate / module / Compatible), AMD FreeSync (Premium Pro
/ Premium / basic), HDMI 2.1 VRR, DisplayPort Adaptive Sync; (b)
**ALLM** — HDMI 2.1 Auto Low Latency Mode negotiation timing; (c)
**LFC** — Low Framerate Compensation behaviour at the VRR-range
floor; (d) **client-side frame interpolation** — DLSS-G, FSR-G,
XeSS-G applied receive-side to absorb network jitter; (e) **Adaptive
VSync vs FastSync** decision logic; (f) **tear-free presentation**
strategies including scanline sync; (g) **display timing** — the
VSYNC / buffer-flip / scanout chain at 60 / 120 / 144 / 240 / 360 Hz
and at 4K240 / 4K480 OLED; (h) **feasibility thresholds** for
streaming at high refresh rates.

### 1.5 Out of scope

Out of scope: **host-side Reflex 2 / Frame Warp** (C18 §5 owns the
host-side perceived-latency lever), **TV-side ALLM negotiation
primitives** (C12 §6 owns the negotiation surface; this chapter
specifies the timing), **input prediction** (C21 owns it), **GPU-
Direct host-side zero-copy** (C18 §2 owns it), and **encoder-side
frame pacing** (C13 §3 + C18 §4 own it).

### 1.6 R-18 (Operational Integrity) boundary

The Constitution §11.5 forbids any command that suspends, hibernates,
locks, or terminates the operator's host. This chapter introduces
client-side display-mode setting — `xrandr` on Linux, the Windows
display-settings API, `CGDisplayConfigure` on macOS — to negotiate
VRR ranges and set refresh rates. **All such subprocess invocations
on dedicated client hardware (the Wails desktop client) wrap through
`r18.SafeExec`**, inherited from C13 §10 and the host-agent chapter.
Web clients, mobile clients, and TV clients use OS-native APIs
without subprocess and are therefore R-18-trivial. The dedicated
client never invokes `xrandr --off`, `xrandr --auto` without a target
mode, or any DDC/CI command that could blank the operator's primary
display while a streaming session is live.

## 2. VRR fundamentals

### 2.1 What VRR is

Variable Refresh Rate is the technique by which a display
**dynamically changes its refresh interval** to match the GPU's frame
production cadence. Traditional fixed-refresh displays scan the panel
at a rigid 60 / 120 / 144 / 240 Hz cadence — every 16.67 / 8.33 / 6.94
/ 4.17 ms regardless of whether the GPU has produced a new frame. If
the GPU produces a frame mid-scanout, the display tears (one half of
the panel shows the new frame, the other half shows the old). The
classic mitigation, fixed VSYNC, **buffers the GPU output until the
next refresh boundary** — eliminating tearing but adding up to one
full refresh interval of latency, plus the input-latency penalty of
double or triple buffering.

VRR replaces this rigid cadence with a **per-frame variable scanout
interval**. The display waits up to a maximum interval (the panel's
VRR maximum) for the GPU to signal "frame ready"; on signal, it
begins scanout immediately. The result: zero tearing, zero added
buffer latency, and a hard cap on added latency at the inverse of the
panel's VRR minimum.

`latency_dim08.md` §2 cites the binding figure: **VRR adds < 1 ms of
latency compared to fixed refresh rate**, derived from the NVIDIA
G-SYNC technical documentation. This is HC-09. The 2026 reaffirmation
(C13 addendum §C4) clarifies that this figure is a **mean** and that
the real win is in the **tail** — fixed VSYNC's worst case is one
full refresh interval (16.67 ms at 60 Hz); VRR's worst case is < 1
ms.

### 2.2 NVIDIA G-Sync — three tiers (2026 status)

NVIDIA segments G-Sync into three tiers, distinguished by hardware
implementation and certification regime:

| Tier | Hardware | VRR range | HDR | Notes |
|---|---|---|---|---|
| G-Sync Ultimate | Dedicated G-Sync module + HDR1000+ panel | 1–360 Hz (some 4K240) | HDR1000+ certified | Highest tier; full G-Sync feature set including variable overdrive |
| G-Sync (legacy module) | Dedicated G-Sync module | 30–240 Hz typical | Not required | Original 2013-era hardware; still shipped on some pro gaming monitors |
| G-Sync Compatible | No module — VESA Adaptive Sync over DP / HDMI 2.1 | Per-panel (typically 48–240 Hz) | Not required | NVIDIA-certified subset of FreeSync / VESA panels |

`latency_dim08.md` §2 anchors the 30–240 Hz typical range. The 1–360
Hz and 4K240 figures for the Ultimate tier are 2026 reaffirmations
from the C13 addendum §C cluster (NVIDIA technology page snapshot
2026). HelixPlay's host-agent capability schema reports the G-Sync
tier in its hello-frame, but the **admission policy does not gate on
tier** — any VRR-capable display is acceptable, including G-Sync
Compatible.

### 2.3 AMD FreeSync — three tiers (2026 status)

AMD's tier structure mirrors NVIDIA's three-tier model with different
naming and certification thresholds:

| Tier | Required features | Refresh / response | HDR | Notes |
|---|---|---|---|---|
| FreeSync Premium Pro | LFC + HDR + < 4 ms response time | ≥ 120 Hz at FHD | Yes (HDR400+) | Renamed from FreeSync 2 HDR; highest tier |
| FreeSync Premium | LFC + ≥ 120 Hz at FHD | ≥ 120 Hz at FHD | Not required | Mid tier |
| FreeSync (basic) | VRR over DP / HDMI | Any | Not required | Entry tier |

LFC (Low Framerate Compensation, §2.6) is the gating feature for the
Premium and Premium Pro tiers. As with G-Sync, HelixPlay's capability
schema **advertises tier but does not gate on it** — any FreeSync tier
is acceptable for streaming; only the per-panel VRR range matters for
admission decisions.

### 2.4 HDMI 2.1 VRR + DisplayPort Adaptive Sync

HDMI 2.1 (the 2017 specification, deployed across consumer TVs from
2019 onwards) adds **VRR as a standardised industry feature**,
independent of the G-Sync / FreeSync vendor brands. The PS5, Xbox
Series X / S, and Switch 2 (April 2026) all implement HDMI 2.1 VRR;
all major TV brands (Samsung, LG, Sony, TCL, Hisense) ship HDMI 2.1
VRR on flagship models from 2020 onwards. HelixPlay's TV-tier client
negotiates VRR over HDMI 2.1 with no vendor-specific code path.

DisplayPort Adaptive Sync (DP 1.2a, 2014) is the equivalent over
DisplayPort and is the underlying transport for G-Sync Compatible
and most monitor-tier FreeSync implementations. The HelixPlay client
negotiates VRR over **whichever cable the user has** — the protocol
is identical from the client's perspective; only the EDID
fingerprint differs.

### 2.5 HDMI 2.1 ALLM — Auto Low Latency Mode

ALLM is an HDMI 2.1 feature in which the **source signals "low-
latency mode"** through the AVI InfoFrame and the display **disables
its image-processing pipeline** in response — motion smoothing, HDR
tone-mapping, edge enhancement, scaling, audio-visual sync delay,
and any other latency-adding processing is suppressed for the
duration of the low-latency frame.

The negotiation timeline:

1. **HDMI handshake** — at cable connect / mode-set, the source's
   EDID-read inspects the sink's VSDB for the ALLM-capable bit.
2. **AVI InfoFrame assertion** — when entering a low-latency
   workload (game-mode, streaming session), the source asserts the
   ALLM bit in the AVI InfoFrame on every video frame.
3. **HDMI-CEC supplementary signalling** — for displays that do not
   honour AVI-InfoFrame ALLM (common on older HDMI 2.0 + ALLM
   "compatible" panels), HelixPlay falls back to HDMI-CEC's
   `<Set System Audio Mode>` / `<Active Source>` sequence as
   specified in C12 §6.
4. **Mode-change propagation** — when the streaming session ends,
   the source de-asserts the ALLM bit and the panel returns to its
   non-game picture mode.

The win is enormous on consumer TVs: ALLM saves **50–100 ms** of
display-side processing on most TVs, where the default picture-mode
processing chain (motion smoothing alone is typically 30–80 ms) is
the **dominant latency contributor in the receive-to-display
budget**. For comparison, VRR's < 1 ms gain (HC-09) is a polish
item; ALLM's 50–100 ms gain is foundational. C12 §6.3 owns the TV-
side surface; this chapter §2.5 owns the timing of the AVI
InfoFrame + HDMI-CEC negotiation chain.

### 2.6 LFC — Low Framerate Compensation

VRR ranges have a minimum (typically 30 / 40 / 48 Hz depending on
panel). When the GPU's frame rate drops below this minimum, the
display reverts to a fixed mode and tearing or stutter returns. LFC
mitigates this by **duplicating frames inside the GPU driver** to
keep the effective rate inside the VRR range — at 25 fps with a
40 Hz VRR minimum, LFC duplicates every frame, presenting at 50 Hz
inside the VRR window.

LFC is **mandatory for FreeSync Premium / FreeSync Premium Pro
certification** and standard on all G-Sync Ultimate / G-Sync module
displays. The result is the elimination of the "stutter cliff" that
afflicts non-LFC VRR setups when frame rate drops below the VRR
minimum.

For HelixPlay, LFC is a **safety net**, not a primary mechanism. The
encode pipeline targets ≥ 60 fps end-to-end (C18 §4 owns the encode
floor); LFC absorbs only transient drops — typically network-
induced frame drops where the client's jitter buffer underflows for
1–3 frames.

### 2.7 VRR adoption status — 2026

The VRR ecosystem is mature in 2026:

- **TVs** — All major brands (Samsung, LG, Sony, TCL, Hisense) ship
  HDMI 2.1 VRR on flagship models. Mid-range models from 2023 onwards
  ship VRR as standard. Entry-level models from 2024 onwards
  increasingly include HDMI 2.1 VRR as a marketing feature.
- **Consoles** — PS5, Xbox Series X / S, Switch 2 (April 2026)
  support HDMI 2.1 VRR. Legacy consoles (PS4, Xbox One, Switch 1)
  lack support and fall back to fixed-rate VSync.
- **Monitors** — The vast majority of gaming-class monitors since
  2022 support either FreeSync, G-Sync Compatible, or both.
- **Mobile** — High-refresh-rate mobile panels (90 / 120 Hz) on
  flagship phones from 2021 onwards support per-frame variable
  refresh rate, although mobile VRR ranges are typically narrow
  (48–120 Hz) and are exposed via OS-level APIs rather than as a
  user-facing feature.

HelixPlay's client capability schema reports VRR-capable as a single
boolean per output, with the panel's VRR range and the relevant
G-Sync / FreeSync tier reported as a structured sub-object for
diagnostics. **Non-VRR fallback is fixed-rate VSync on the panel's
native refresh rate, NOT a session refusal** — VRR is a polish layer,
not an admission gate. This is the Z5 differentiator's operational
form: VRR is the default, fixed-rate is the fallback, and HelixPlay's
codification of this end-to-end is what distinguishes it from open-
source streaming clients (Moonlight on macOS, Sunshine without a
custom build) that still treat VRR as a gap.
## 3. Frame pacing + tear-free presentation

### 3.1 Frame pacing — what it is

Frame pacing is the discipline of delivering frames to the display at uniform
intervals so the human visual system perceives smooth motion. Even a render
pipeline that averages a comfortable 60 fps can produce a juddering, unstable
experience when frame intervals deviate from the nominal 16.67 ms. This is the
core finding of dim 08: a game running at 60 FPS with uneven inter-frame
intervals (10 ms, 20 ms, 10 ms) feels worse than 60 FPS with consistent
16.67 ms frames (latency_dim08.md §1, HIGH confidence). Two distinct artifact
classes emerge: **stutter**, where consecutive frames arrive at irregular
boundaries and motion appears to skip; and **judder**, where the same frame
is repeated across multiple display refreshes because the pipeline missed a
deadline. Frame pacing's job is to align frame *production* to display
*vertical-blank* boundaries — either by stalling production until the display
is ready (fixed-rate sync), by letting the display follow production (VRR), or
by inserting interpolated frames to maintain rhythm. HelixPlay treats frame
pacing as a first-class concern of the **client surface** (per Insight #3,
asymmetric optimisation — C13 §3.3 names this the "client-side frame pacing"
tier).

### 3.2 VSYNC — fixed-rate sync

Classic VSYNC is the oldest answer: the GPU waits for the display's VBLANK
signal before flipping the back-buffer to the front-buffer. Tearing is
eliminated because the swap happens during the blanking interval when no scan
is in progress. The cost is latency: a frame finished 1 ms after VBLANK must
wait the full refresh interval (up to 16.67 ms at 60 Hz, 8.33 ms at 120 Hz)
before it can be shown. For a cloud-gaming client where every millisecond
costs perceived responsiveness, this penalty is unacceptable as a *default*.
HelixPlay's MVP nonetheless ships VSYNC as the **fallback** path for displays
that lack VRR support and lack the GPU/driver combination required for the
adaptive variants below — a sub-tier of the client-display capability matrix.
Capability advertisement (cross-link C18 §5 / C13 §8) marks VSYNC-fallback
clients explicitly so the host encoder can reason about target cadence.

### 3.3 Triple-buffered VSync (FastSync, Enhanced Sync)

NVIDIA FastSync and AMD Enhanced Sync introduce a third buffer between the
back-buffer and the front-buffer. The GPU renders into the back-buffer
without blocking; the third buffer holds the most recent completed frame; at
each VBLANK the most recent buffer flips to the display and the older one is
discarded. The effect is tear-free presentation without the full VSYNC
latency penalty when the render rate exceeds the display rate. The cost is
VRAM (3× framebuffer footprint), driver complexity, and slightly increased
GPU load (because frames are sometimes rendered only to be discarded).
HelixPlay does not enable triple-buffered VSync by default on the streaming
client — there is no scenario in which the client renders faster than the
display rate (the host's encode rate is the upstream bottleneck) — but the
mode is exposed in the desktop client's advanced settings for parity with
local-game expectations.

### 3.4 Adaptive VSync (NVIDIA Adaptive)

NVIDIA's Adaptive VSync (driver-side, ≥ R304) is a hybrid that engages VSYNC
when the GPU's frame rate meets or exceeds the display rate (no tearing) and
disengages VSYNC when the frame rate falls below it (no stutter from missed
VBLANK deadlines). Tearing returns under load, but the perceptual trade is
favourable for many users. HelixPlay's client uses Adaptive VSync as the
**preferred fallback** when VRR is unavailable on the connected display —
i.e. the chosen mode for the bulk of laptop-integrated panels and TVs that
predate HDMI 2.1 VRR.

### 3.5 VRR — frame pacing solved

Variable Refresh Rate (G-Sync, FreeSync, HDMI 2.1 VRR — covered in §2) is
HelixPlay's **default** display mode whenever the client capability surface
includes it. The display scans out frames as they arrive from the GPU, with
no fixed VBLANK boundary to miss; tearing is eliminated because every frame
boundary is a refresh boundary. Cross-verification HC-09 confirms the cost is
< 1 ms versus a fixed-rate display (latency_cross_verification.md;
latency_dim08.md §2 — "G-SYNC adds minimal latency — less than 1ms compared
to fixed refresh rate", HIGH confidence). Combined with LFC (§2.3) for
sub-VRR-floor frame rates, VRR completely solves the frame-pacing problem
for HelixPlay's primary tier (desktop / laptop / TV with HDMI 2.1, plus
modern Android phones with adaptive sync). Cross-link §2 of this chapter for
the per-vendor VRR taxonomy.

### 3.6 Tear-free presentation — full-screen exclusive vs windowed

Independent of VSYNC/VRR, the **presentation surface** matters:

- **Full-screen exclusive (FSE)**: the GPU directly drives the display, the
  desktop compositor is bypassed, and the swap-chain is owned by the client
  process. Lowest latency; required for the < 1 ms VRR floor on Windows.
- **Windowed (composited)**: DWM (Windows), the Wayland compositor (Linux),
  or WindowServer (macOS) intercepts frames and re-composites them with the
  rest of the desktop. This adds typically one frame of latency (16.67 ms at
  60 Hz) and can disable VRR on some platforms.

HelixPlay's rule: the **desktop / laptop tier** of the client offers FSE mode
explicitly and selects it by default for full-screen sessions; **mobile, TV,
and web tiers** are inherently composited (the platform owns presentation)
and run with whatever the platform exposes. On Windows 11 22H2+, the
"Optimised for windowed games" mode (DirectFlip / independent flip) recovers
FSE-like latency in windowed apps when the swap-chain meets the criteria
(no overlapping windows, full-screen-of-monitor, matching format) — HelixPlay
hints the desktop client toward this configuration on capable Windows
builds. On Linux, the equivalent path is Wayland's `tearing-control-v1`
protocol (with KDE 6 / GNOME 47+ support landing through 2025) — HelixPlay
declares tearing-control intent on the Wayland surface so the compositor can
direct-scanout when geometry permits.

### 3.7 Frame-pacing budget for HelixPlay

The inter-frame budget at the display tier is set by the refresh rate and is
**hard** — the display will scan out at its rate regardless of what the
pipeline upstream does. HelixPlay's encode rate (host-side) and decode rate
(client-side) target the operator-configured stream rate; the display rate
may differ; LFC + VRR absorb the mismatch.

| Display rate | Inter-frame budget | HelixPlay tier (typical) |
|---|---|---|
| 60 Hz | 16.67 ms | TV (legacy), web fallback, low-end mobile |
| 120 Hz | 8.33 ms | TV (HDMI 2.1), gaming laptops, modern phones |
| 144 Hz | 6.94 ms | Desktop monitors (mainstream gaming) |
| 240 Hz | 4.17 ms | Desktop monitors (esports / competitive) |
| 360 Hz | 2.78 ms | Desktop monitors (top-tier 2024–2026) |

The host's encode pipeline produces frames at the host's render rate (which
may itself be VRR-driven on the host GPU); the network jitter buffer (§5.4
of C13) reshapes their arrival cadence; the client decoder emits decoded
frames; VRR scans them out as they arrive. The per-tier rule: **never assume
the client display rate matches the host render rate.** The mismatch is the
norm, and the absorption mechanism is the LFC-extended VRR window plus the
1–3-frame jitter buffer (latency_dim08.md §4 — "Jitter buffer of 1-3 frames
(16-50ms at 60Hz) absorbs network variability while keeping latency
minimal", MEDIUM confidence).

## 4. Client-side frame interpolation (DLSS-G / FSR-G / XeSS-G)

### 4.1 What client-side frame interpolation is

Client-side frame interpolation is a perceived-latency lever distinct from
host-side Frame Warp (Reflex 2 / Anti-Lag 2 / XeLL — covered in C18 §5).
The mechanism: display real frame N, then *generate* interpolated frame
N+0.5 (or N+0.33 / N+0.66 / N+0.75 for multi-frame variants) from an optical-
flow model running on the client GPU's tensor / matrix cores, then display
the generated frame, then display real frame N+1. The visible frame rate
doubles (or 4× with MFG); the underlying *real* frame cadence is unchanged.
The asymmetry is critical: interpolation increases visible smoothness but
**adds latency**, because frame N+0.5 cannot be shown until the model has
completed its prediction, which requires that frame N+1 already be available
on the GPU as a reference. The minimum cost is one frame of additional
latency (16.67 ms at 60 Hz, 8.33 ms at 120 Hz). Cross-link
latency_dim08.md §5 — frame interpolation in the practical-recommendations
table is annotated "+0ms" but that is the *steady-state* visible-frame-rate
view; the real cost is one buffered real-frame interval, which dim 08 §4
implicitly captures via the jitter-buffer discussion.

### 4.2 NVIDIA DLSS Frame Generation (DLSS-G) — 2026 status

DLSS Frame Generation shipped first in DLSS 3 (2022, RTX 40 series /
Lovelace). DLSS 3.5 added Ray Reconstruction; DLSS 4 (CES 2025) introduced
**Multi Frame Generation** on RTX 50 series (Blackwell), generating up to
three interpolated frames per real frame (4× MFG). DLSS 4.5 (2026)
extends MFG availability and reduces the per-frame compute cost on Blackwell
silicon. The latency cost remains approximately one frame at the source
cadence — the model's inference time is well under a millisecond on
Blackwell tensor cores, but the buffered-reference-frame requirement
dominates. HelixPlay's posture: **opt-in only**. The operator can enable
DLSS-G per tenant via the client capability flag; default is off. Rationale:
the Conservative Prediction Paradox (Insight #5) — interpolated frames are
*predictions* of intermediate game state, and incorrect predictions during
rapid scene changes produce visible artifacts that are more jarring than the
underlying frame-rate mismatch they were trying to mask. C13 §7 already
records this as MVP posture: HelixPlay does **not** perform client-side
frame interpolation by default in the streaming client — it relies on host-
side frame pacing (Reflex 2 + Anti-Lag 2 + XeLL surface — C18 §5), VRR
scanout alignment (§3.5 above), and the 1–3-frame jitter buffer.

### 4.3 AMD FSR Frame Generation (FSR-G) — 2026 status

AMD's FSR 3 (2023) introduced Frame Generation as a vendor-neutral 2×
generation path, working on RDNA 2 and newer AMD GPUs, on Lovelace+ NVIDIA
GPUs (where DLSS-G is also available), and on Intel Arc Alchemist+. FSR 4
(2024) and FSR 4.1 (2026) refine the optical-flow model and add AI-assisted
upscaling on RDNA 4. Latency cost: parity with DLSS-G — one buffered
real-frame interval. HelixPlay's posture matches DLSS-G: **opt-in only,
default off**, behind the same capability surface and the same Conservative
Prediction Paradox rationale. The vendor-neutrality is a deployment
advantage for the desktop client (no per-vendor branching needed) but does
not change the latency or artifact properties.

### 4.4 Intel XeSS Frame Generation (XeSS-G) — 2026 status

Intel released XeSS Frame Generation as part of XeSS 2 in late 2024, with
XeSS 3 (2026Q1) shipping alongside Battlemage Arc B-series. XeSS-G is 2×
generation; on Battlemage, XeSS-G runs on the XMX matrix cores at parity
latency cost with FSR-G and DLSS-G (≈ one buffered real-frame interval).
XeSS-G is paired with **XeLL** (Xe Low Latency) — Intel's Reflex / Anti-Lag
counterpart, covered in C18 §5 — which mitigates the latency cost when both
are enabled together. HelixPlay's posture: same opt-in default-off rule as
DLSS-G and FSR-G; the per-tenant capability flag distinguishes the three so
operators can match the underlying GPU vendor.

### 4.5 Conservative Prediction Paradox (cross-link Insight #5)

Frame interpolation is a special case of the broader prediction paradox
identified in latency_insight.md Insight #5 (HIGH confidence): predicting
*wrong* is worse than predicting *late*. The interpolated frame is a
prediction of game state at time N+0.5 derived from frames N and N+1.
When the prediction is correct (smooth camera pan, continuous analog motion)
the artifact is invisible and the perceived smoothness gain is real. When
the prediction is wrong (rapid scene change, occlusion-disocclusion at a
geometry boundary, particle effect appearing between reference frames) the
visible artifact is a "snap" or "warp" that draws the eye. Insight #5's
rule applies directly: predict only high-confidence motion, never predict
discrete events. Frame-interpolation models cannot make that distinction at
runtime — they predict everything, all the time.

HelixPlay's MVP default is therefore: client-side frame interpolation is
**OFF**. The streamed video flows at the host's native render rate, the
client's VRR display absorbs the cadence (§3.5), and the perceived-latency
budget is owned by the host-side Reflex 2 / Anti-Lag 2 / XeLL surface
(C18 §5) plus the host's frame-time pacing tier (latency_dim08.md §4 —
"Frame Warp on host (if NVIDIA) to reduce perceived input latency"; HC-03
in cross-verification). Operators may enable client-side interpolation per
tenant on a per-game basis when telemetry shows the game is dominated by
smooth-motion content (cinematic / open-world / driving) and not by
twitch-shooter content where the artifact penalty is highest. This
operator-question is tracked as **OQ-C22-02** in §10 of this chapter, with
the resolution criteria tied to bitstream-verifier artifact rates (C24)
and per-game perceived-latency telemetry from PresentMon 2.2 (C18 §5
addendum) and the LDAT/OSRTT methodology pipeline (C13 §11 addendum).
## 5. Display timing synchronisation

§§2–4 specified the **stream-side** half of HelixPlay's frame-pacing
contract: VRR negotiation between encoder and client, frame-pacing
interpolation that absorbs network jitter, and the client-side frame
generation surface (DLSS-G / FSR-G / XeSS-G) that fills cadence gaps
without inflating perceived latency. The chapter now closes on the
**panel-side** terminal stage — the physical journey of each pixel
from the client's framebuffer to the photons emitted by the panel
glass. C13 §3 reckons HelixPlay's display-side budget at **5–6 ms**
end-to-end on a contemporary OLED panel, and the contract in this
section is what makes that number reproducible across the Wails /
Flutter / Compose-for-TV client surface enumerated in C04. Cross-link
[`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md)
§3 + §8, and C18 §6 for the GPU-side pattern this section mirrors.

### 5.1 The display pipeline — three stages

Every panel HelixPlay drives shares the same three-stage internal
pipeline, regardless of vendor or panel technology:

- **Buffer flip** — the GPU swaps the front-buffer / back-buffer
  pair in its display engine. Under VSYNC the swap is gated to the
  next VBLANK boundary; under VRR the swap can occur at an
  arbitrary instant within the panel's supported refresh range
  (§5.4). The flip itself is sub-millisecond on modern GPUs (one
  PCIe descriptor write to the display controller's flip register).
- **Scanout** — the display controller reads the framebuffer line
  by line and pushes pixel data over the physical link (HDMI 2.1,
  DisplayPort 2.1, or eDP for laptop panels). At 4K60 a full
  scanout takes ~16.4 ms; at 4K120 it takes ~8.2 ms; at 4K240
  it takes ~4.1 ms. Scanout time scales **linearly with refresh
  rate**, which is why higher-Hz panels matter for end-to-end
  latency even when the source frame rate is identical.
- **Panel response** — the physical pixels change colour. On
  modern OLED panels the response time is **< 1 ms** (organic LED
  emission is effectively instantaneous). On LCD panels response
  time depends on the transition magnitude — black→white is ~4 ms,
  partial-grey transitions can stretch to 16 ms in worst-case
  panels. HelixPlay's display-side budget of 5–6 ms is comfortable
  on OLED but tight on LCD; C04 §6 tier-2 admission policy records
  the panel-class delta.

### 5.2 Tearing — what it is, why VRR fixes it

When the GPU flips the framebuffer in the **middle** of a scanout —
the display controller is reading line 540 when the buffer pointer
swaps from frame N to frame N+1 — the upper half of the panel shows
frame N and the lower half shows frame N+1, with a horizontal seam
at the swap line. This is **tearing**, and it is purely a frame-
pacing artefact: the GPU and display are running on independent
clocks, and without synchronisation any flip can land mid-scanout.

Two solutions: **VSYNC** waits for the next VBLANK before flipping
(zero tearing, but adds up to one full refresh interval of latency
when the GPU finishes early); **VRR** lets the **display** wait
for the GPU instead — the display extends its current scanout
until the GPU presents the next frame, then immediately starts a
new scanout. VRR adds **< 1 ms** of latency versus a fixed-rate
panel that is already aligned to the source cadence (panel-side
control logic + scanout-clamp negotiation), and eliminates tearing
entirely within the supported refresh range. HelixPlay defaults
to VRR-on whenever the client display advertises support; cross-
link §3 for the negotiation envelope and §6.2 for the capability
schema.

### 5.3 OLED vs LCD response time

OLED panels (every generation since LG WOLED 2019, Samsung QD-OLED
2022, JOLED inkjet-OLED 2024) emit light per-pixel via an organic
LED, with a response time floor of **0.1–0.5 ms** dominated by the
TFT backplane switching speed rather than the emitter. LCD panels
modulate a backlight via a liquid-crystal shutter that has a
mechanical relaxation time of **4–16 ms** depending on the
transition (worst case is a partial grey-to-grey transition through
the liquid-crystal hysteresis curve). HelixPlay's rule: when the
client capability schema reports `display.is_oled = true`, the
host-agent advertises the session into a **"minimal-latency"**
marketing tier that the operator white-label theming surface (C11)
can surface to the user. LCD-tier sessions are not refused, but
they receive a different perceived-latency expectation in the
client UI and skip the §5.4 high-refresh tier admission rule.

### 5.4 4K240 / 4K480 OLED panels (2026)

The 2026 panel market has moved decisively. **4K240 OLED** is
mainstream — LG OLED Evo G5 + C5 ship 4K240; Samsung S95F + S90F
ship 4K240; ASUS PG27UCDM and MSI MEG 32-inch OLED ship 4K240 in
the gaming-monitor segment. Every major TV chassis at CES 2026
demoed at minimum 4K240 capability. **4K480 OLED** appeared in
prototype form (Samsung Display + LG Display CES booths) but is
not yet shipping in consumer SKUs as of April 2026, and the
host-side encode pipeline (C18 §3 + C13 §3) targets a stable
4K120 ceiling. HelixPlay's posture: a 4K240 client is **over-
headroom** for the MVP host encode rate (only matters when the
host renders at 240+ fps and the encoder advertises matching
capability — a rare configuration in cloud-rendering scenarios).
The orchestrator records the client's max supported rate in the
capability schema and admits sessions at the lower of the two,
so a 4K240 panel paired with a 4K120 host stream simply caps at
120 Hz without complaint.

### 5.5 Refresh-rate negotiation — EDID + capability schema

Every modern panel exposes its supported video modes through
**EDID** (Extended Display Identification Data) — a 128-byte (or
256-byte with E-EDID extensions) blob the panel transmits over
the HDMI / DisplayPort / eDP DDC channel during link training.
The blob carries the panel manufacturer ID, supported refresh
rates, supported resolutions, VRR range (`Min Hz` / `Max Hz`),
HDMI 2.1 ALLM bit, and HDR metadata.

HelixPlay's bootstrap (§6.3) reads the EDID through the platform-
specific API (`xrandr` on X11; `wlr-randr` or `kscreen-doctor` on
Wayland; `EnumDisplayDevices` + `DEVMODE` on Windows;
`CGDisplayCopyAllDisplayModes` on macOS; native EDID-on-HDMI APIs
on Android TV / Tizen / webOS), parses the supported rate set,
and reports it to the host-agent over the existing capability
channel. The host-agent picks the stream rate as
`min(host_render_rate, client_display_rate)` and signals VRR
range so the encode pipeline (C18 §3.5) knows whether the panel
can stretch its scanout — relevant when network jitter pushes a
frame's arrival outside the nominal cadence.

Cross-link **OQ-C22-03** (per-tenant refresh-rate ceiling): the
operator-policy surface (C09 §3, C11 §7) admits a per-tenant
override that caps the negotiated rate independent of panel
capability, so an operator running a battery-conscious household
tier can constrain even a 4K240-capable client to 60 Hz to
preserve client-side power budget.

## 6. Implementation contract

### 6.1 Submodule boundaries (R-03)

Per **R-03** (every reusable component lives under
`vasic-digital`), the display-timing-synchronisation work
introduces one new public submodule:

- **`vasic-digital/helix-display`** — public Go module, MIT
  licensed, hosted on GitHub mirrored to GitLab + GitVerse +
  GitFlic. Exposes the following packages:
  - `display.RefreshRateNegotiator` — EDID parser + capability
    negotiator. Constructor reads the platform-specific EDID
    surface, enumerates supported modes (resolution / refresh
    rate / colour depth tuples), and exposes a
    `Negotiate(hostRate int) (int, error)` method that returns
    the chosen stream rate.
  - `display.VRRController` — enables / disables VRR on the
    target display. Linux: `xrandr --output ... --set "vrr_enable" 1`
    on X11; `wlr-randr --output ... --adaptive-sync enabled` on
    Wayland. Windows: DXGI swap-chain `SetFullscreenState` plus
    `IDXGIOutput6::CheckHardwareCompositionSupport`. macOS:
    `CGDisplaySetDisplayMode` with the `kCGDisplayShowDuplicateLowResolutionModes`
    flag for ProMotion. Android TV / Tizen / webOS: native
    media-framework call.
  - `display.ALLMSignaller` — emits the HDMI-CEC ALLM (Auto
    Low-Latency Mode) signal when the client capability schema
    reports HDMI 2.1 + ALLM support. Cross-link C12 §6 (TV UX —
    HDMI-CEC owner; this submodule is the latency-side caller
    of the C12 primitive).
  - `display.FrameInterpolator` — capability detection only for
    DLSS-G / FSR-G / XeSS-G. The actual frame generation is
    done by the vendor SDK; this submodule advertises whether
    the client GPU supports it so the host-agent can decide
    whether the orchestrator allows the operator-policy
    `frame_gen_enabled` toggle on this client.

`helix-display` **reuses** existing submodules (R-04 DRY):

- `vasic-digital/helix-r18-safeexec` (C08 §10 — single-source
  forbidden-command deny-list; **never** duplicated here).
- `vasic-digital/helix-capability-schema` (C09 §3 — admission
  policy schema versioned across the host-agent / orchestrator
  / client surface).

### 6.2 Capability schema delta

The **client** capability stanza (C09 §3) gains the following
fields under a new `display` section:

- `display.refresh_rates: []int` — sorted ascending, e.g.
  `[60, 120, 144, 240]`. Populated from EDID parsing at
  bootstrap.
- `display.vrr_supported: bool` — `true` when the panel
  advertises VRR via EDID extension block CTA-861-G or
  DisplayID 2.0 `AdaptiveSync` capability.
- `display.vrr_min_hz: int` — typically `30` on G-Sync /
  FreeSync Premium panels, `48` on entry-level FreeSync.
- `display.vrr_max_hz: int` — typically `144` on mid-tier
  gaming displays, `240` on 2026 OLED gaming displays.
- `display.allm_supported: bool` — `true` when the HDMI link
  is 2.1 (data-rate ≥ 18 Gbps via Forum-VRR is sufficient) and
  the panel exposes the ALLM bit in its CTA-861-G InfoFrame
  reply.
- `display.frame_gen_supported: bool` — `true` when the client
  GPU advertises DLSS-G (NVIDIA Lovelace+), FSR-G (AMD RDNA 3+),
  or XeSS-G (Intel Battlemage+).
- `display.is_oled: bool` — `true` when EDID manufacturer ID +
  product ID match the OLED-panel ID database baked into
  `helix-display` at build time.

Cross-link **C09 §3 admission policy**: the orchestrator uses
`display.refresh_rates` and `display.vrr_min_hz` /
`display.vrr_max_hz` to compute the admission envelope —
sessions whose host render rate falls outside the client VRR
range are admitted in fixed-VSYNC mode rather than refused, but
the operator-policy `tier=competitive` flag does refuse out-of-
range admissions.

### 6.3 Bootstrap sequence (client-side)

The client runs the following startup sequence:

1. **Read EDID** through the platform-specific API listed in
   §5.5; populate the capability schema stanza in §6.2.
2. **Signal capability set** to the host-agent over the
   existing Connect-Go RPC channel (cross-link C13 §8 + C09).
3. **Host-agent picks the stream rate** as
   `min(host_render_rate, client_display_rate)`, factoring in
   the per-tenant ceiling from §5.5 OQ-C22-03.
4. **Enable VRR mode** on the client side, platform-specific:
   - **Linux X11**: `xrandr --output <name> --mode <mode> --rate <hz>`
     wrapped through `r18.SafeExec`.
   - **Linux Wayland**: same effect via the compositor's
     protocol (`kde-output-management-v2` /
     `wlr-output-management-unstable-v1`); no shell-out needed.
   - **Windows**: `SetDisplayConfig` Win32 API + `DXGI`
     swap-chain `SetMaximumFrameLatency(1)`. No shell-out.
   - **macOS**: `CGDisplaySetDisplayMode` plus
     `CGDisplaySetStereoOperation` for ProMotion. No shell-out.
   - **Android TV / Tizen / webOS**: native API
     (`Display.setRefreshRate` / Tizen `display.setMode` /
     webOS `LSCallOneReply` to `com.webos.service.tv.display`).
5. **Signal ALLM** via HDMI-CEC if the panel advertises it
   (cross-link C12 §6 — the TV UX chapter owns the CEC
   transport; `display.ALLMSignaller` is the latency-side
   caller).
6. **Toggle frame-gen** capability if the operator policy
   permits — per-tenant gate from C09 §3, default off (per
   C13 §7 — frame generation increases perceived smoothness
   but adds 1 frame of buffering, so it is opt-in not default).

### 6.4 Go code

The `RefreshRateNegotiator` constructor and `Negotiate` method,
end-to-end. Real imports: `golang.org/x/sys/unix` for the EDID
sysfs read on Linux; `r18` for `xrandr` shells; the parser is
pure Go.

```go
package display

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"golang.org/x/sys/unix"
	r18 "github.com/vasic-digital/helix-r18-safeexec"
)

// Mode is one EDID-advertised display mode.
type Mode struct {
	Width, Height int
	RefreshHz     int
	IsPreferred   bool
}

// RefreshRateNegotiator parses EDID and picks the best stream rate.
type RefreshRateNegotiator struct {
	displayPath string // e.g. "/sys/class/drm/card0-HDMI-A-1"
	modes       []Mode
	vrrMinHz    int
	vrrMaxHz    int
}

// NewRefreshRateNegotiator reads the EDID blob from the kernel DRM
// sysfs interface and parses the supported mode set. The caller
// supplies the canonical sysfs path for the connected output; the
// constructor returns an error if the file is missing or if EDID
// parsing fails.
func NewRefreshRateNegotiator(displayPath string) (*RefreshRateNegotiator, error) {
	edidPath := displayPath + "/edid"
	fd, err := unix.Open(edidPath, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("display: open %s: %w", edidPath, err)
	}
	defer unix.Close(fd)
	blob, err := os.ReadFile(edidPath)
	if err != nil {
		return nil, fmt.Errorf("display: read EDID: %w", err)
	}
	if len(blob) < 128 {
		return nil, errors.New("display: EDID truncated")
	}
	modes, vmin, vmax, perr := parseEDID(blob)
	if perr != nil {
		return nil, fmt.Errorf("display: parse EDID: %w", perr)
	}
	sort.Slice(modes, func(i, j int) bool {
		return modes[i].RefreshHz < modes[j].RefreshHz
	})
	return &RefreshRateNegotiator{
		displayPath: displayPath,
		modes:       modes,
		vrrMinHz:    vmin,
		vrrMaxHz:    vmax,
	}, nil
}

// Negotiate picks the highest supported refresh rate that is
// less-than-or-equal to the host render rate, applies it via
// xrandr through r18.SafeExec, and returns the chosen rate.
func (r *RefreshRateNegotiator) Negotiate(ctx context.Context, hostRate int) (int, error) {
	if len(r.modes) == 0 {
		return 0, errors.New("display: no modes available")
	}
	chosen := r.modes[0].RefreshHz
	for _, m := range r.modes {
		if m.RefreshHz <= hostRate && m.RefreshHz > chosen {
			chosen = m.RefreshHz
		}
	}
	outputName := outputFromPath(r.displayPath)
	if err := r18.SafeExec(ctx, "xrandr",
		"--output", outputName,
		"--rate", fmt.Sprintf("%d", chosen)); err != nil {
		return 0, fmt.Errorf("display: xrandr apply: %w", err)
	}
	return chosen, nil
}
```

The constructor opens the EDID sysfs node with `O_CLOEXEC` so the
fd does not leak across the client's fork-exec for the player
window. `parseEDID` is defined elsewhere in the package and is
allocation-free on the cold path (it runs once per session).
`Negotiate` walks the parsed mode set, picks the highest mode at
or below the host rate, and applies it through `r18.SafeExec`
— never bypassing the wrapper. Every field is initialised by the
constructor; every method has a real body that exercises the
underlying EDID read or `xrandr` invocation.

### 6.5 R-18 enforcement

The Constitution §11.5 forbidden-command list is **not duplicated
here**; the package imports `r18.SafeExec` from C08 §10 and
inherits its single-source policy. The Latency family allow-list
extension specific to **C22** records the verbatim argv shapes
admitted by the wrapper for this chapter:

- `xrandr --output <name> --mode <mode> --rate <hz>` — allowed on
  dedicated client hosts to apply a negotiated mode at session
  bootstrap. The `<name>`, `<mode>`, `<hz>` arguments are
  substituted from the parsed EDID by the negotiator above and
  validated against a regex-allow-list before being passed to
  `r18.SafeExec`.
- `xrandr --listmonitors` — read-only capability discovery used
  by the bootstrap to enumerate connected outputs before the
  per-output EDID read.
- Windows: `reg query HKLM\SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}\<idx> /v RefreshRate`
  — read-only registry probe used to confirm the negotiated rate
  applied successfully.
- macOS: `system_profiler SPDisplaysDataType` — read-only
  capability discovery, used as the fallback when the
  `CGDisplay*` API surface is unavailable (e.g., headless boot
  before the WindowServer is up).

All four commands wrap through `r18.SafeExec`. **Nothing else** in
`helix-display` calls `os/exec` or `syscall.Exec` directly. Static
analysis enforces this via the per-package CI rule
`helix-r18-safeexec-required` (C08 §12.11 inheritance), and the
Challenge-tier test surface (C24 §5) verifies that the binary
contains no string matching the §11.5.1 forbidden-command set
through a build-time `go vet` plugin.
## 7. Failure modes

The frame-pacing + VRR pipeline owns three runtime populations that
can break: the **negotiation path** (capability-schema probe →
EDID parse → mode-list emission → VRR engagement on the client
display), the **steady-state scanout path** (host stream rate ↔
client refresh rate ↔ panel refresh range, plus ALLM and LFC), and
the **operator-exec path** (`r18.SafeExec`-mediated `xrandr`,
`system_profiler`, and Windows display-API calls when the dedicated
client mutates the display mode). Every canonical failure mode below
gives a Symptom (what the operator or end-user observes), a
Detection (how HelixPlay learns programmatically), a Mitigation
(the in-process recovery), and a Fallback (what happens when
mitigation does not restore the budget). The five-column table that
closes this section is the source of truth for the runbook
generator at `../03_Architecture/12_Latency_Engineering_Overview.md`
§13 and the alert-rule generation in `../08_Operations/04_
Observability_and_Events.md` (queued).

The fallback semantics across F1–F4 follow the same pattern the C15
chapter pinned: **fail closed at admission, degrade open at
runtime**. If the bootstrap primitive is unavailable (F1, F8, F10),
the session is refused with a structured admission failure that the
client surfaces as `ErrCapabilityMismatch` — never a silent fall-back
to fixed-rate VSync without telling the user. If a runtime invariant
breaks after admission (F3 tearing-despite-VRR, F4 LFC oscillation,
F11 multi-monitor confusion), the client re-negotiates the swap
chain, optionally falls back one VRR tier (G-Sync Ultimate → G-Sync
Compatible → vanilla Adaptive Sync → fixed VSync), and the operator
dashboard records the incident. The capability-schema check at
admission (cross-link `../03_Architecture/07_Host_Agent_and_Game_
Lifecycle.md` §2) is the single source of truth for what the **client
display** promises it can sustain; admission rejects sessions whose
tier requires capabilities the client has not advertised.

The **VRR-range mismatch** row (F2) is unique because it is
operator-policy-tunable: the per-tenant ceiling can be 60 Hz / 120 Hz
/ 144 Hz / 240 Hz, and the host caps the streamed rate at the
ceiling rather than at the panel maximum, so a 240 Hz panel paired
with a 144 Hz tenant ceiling is admitted at 144 Hz with no
artefacts. F3 (tearing-despite-VRR) is the chapter's nastiest mode
because it implies a swap-chain bug or a Wayland-compositor lag
bug — both addressed by re-init plus, in the worst case, a known-
good-compositor allow-list (cross-link OQ-C22-04). F5 (ALLM not
signalled) couples directly with C12 §6 — the TV-side surface owns
the HDMI-CEC retry loop; this chapter's role is to detect the missed
signal via the receive-to-display latency tax and to call the
TV-UX retry path, not to re-implement it.

The **frame-gen artefact** row (F6) is the receive-side analogue of
the host-side **Conservative Prediction Paradox** documented in C18:
DLSS-G / FSR-G / XeSS-G interpolate in-between frames to absorb
network jitter, but mis-prediction at scene cuts produces visible
double-image or warp artefacts. The mitigation is a per-frame SSIM /
VMAF check that cross-references the interpolated frame against the
last fully-decoded frame plus the last-known-good interpolation
fingerprint; a sustained drift below the SSIM threshold disables
frame-gen for the session and logs the disable as a tier-degradation
event. F7 (composite-mode fallback when FSE acquisition fails on
Windows) routes through the Windows 11 22H2+ "Optimized for windowed
games" mode, which preserves VRR + ALLM at a measured 1-frame extra
latency cost — acceptable per Insight #3 because the alternative is
no VRR at all on a windowed-game host.

The **`r18.SafeExec`-rejection row (F10)** is the chapter's R-18
compliance trip-wire on the dedicated client. Any privileged
subprocess call HelixPlay issues around the display — `xrandr
--output <out> --rate <hz>` for refresh-rate changes,
`system_profiler SPDisplaysDataType` for macOS EDID retrieval,
`xrandr --query` for capability-schema probing — goes through the
inherited wrapper from `../03_Architecture/07_Host_Agent_and_Game_
Lifecycle.md` §10. The wrapper rejects any argv shape that is not
on the allow-list (`xrandr --query`, `xrandr --output <out> --rate
<hz>`, `xrandr --output <out> --mode <wxh>`, `system_profiler
SPDisplaysDataType`), and any of the §11.5.1 forbidden patterns
(`xrandr --off`, `xrandr --auto` without target, DDC/CI blank
commands) is rejected unconditionally regardless of context. F10
fires whenever a developer attempts a new argv shape that the
wrapper has not yet been taught to recognise; the resolution is to
extend the allow-list with operator review, never to bypass.

Two long-tail rows (F11, F12) capture multi-display + Wayland-
compositor hazards. Multi-monitor VRR has historically been a NVIDIA
G-Sync limitation (a single G-Sync display per system); 2026 driver
support resolves it on both NVIDIA and AMD, but pre-2026 driver
fleets in the field still trip. The Wayland-compositor row tracks a
real population of compositors that drop frames or fail to honour
VRR mode-set requests — the mitigation is a known-good allow-list
seeded with KWin 6.0+ and GNOME Mutter 47+ at MVP launch.

| # | Failure mode | Symptom | Detection | Mitigation | Fallback |
|---|---|---|---|---|---|
| F1 | VRR not enabled despite display capable | Tearing on panning shots; fixed-rate VSync judder at 60 Hz on a 144 Hz monitor | Capability-schema runtime probe (`xrandr --props` parse) vs negotiated mode mismatch — declared `vrr_supported=true` but the active scanout reports `vrr_active=false` | Re-negotiate VRR via the client's swap-chain re-init path; emit `display.vrr_renegotiate{cause="probe-mismatch"}` | Log `display.vrr_degraded{cause="probe-mismatch"}` and continue with fixed VSync — admission already succeeded so the session does not bounce |
| F2 | VRR range mismatch (host renders at 200 fps; display max is 144 Hz) | Streamed rate exceeds panel rate; over-the-budget frames are dropped at the panel boundary | Rate-comparison check at session start (`stream.fps > panel.refresh_max`); per-frame counter at runtime detects sustained excess | Cap the streamed rate at the per-tenant ceiling or panel maximum (whichever is lower) via host-side encoder rate-clamp; emit `display.rate_clamped{ceiling=…}` | Alert `display.rate_mismatch_alert` if the clamp itself fails (e.g., encoder rejects the rate change); session bounces to a host that can match |
| F3 | Tearing visible despite VRR engaged | End-user reports tearing on UI transitions or scene cuts even with `vrr_active=true` | Client-side observability: frame-time variance histogram exceeds the C24-canonical 5% drift SLO over a 30-s window | Re-init the swap chain with explicit `VK_PRESENT_MODE_FIFO_RELAXED_KHR` (Vulkan) or DXGI tearing-flag re-set (D3D12); cross-link OQ-C22-04 (Wayland) | Fall back to fixed-rate VSync for the remainder of the session; log `display.tearing_persistent` and surface a "tier degraded" notice in the client UI |
| F4 | LFC kicks in repeatedly (frame rate dipping below 30 fps for sustained periods) | Stutter on dense scenes or under network jitter; LFC double-frames are visually visible at low framerates | LFC engagement counter exposed by the client GPU driver (`nvidia-smi --query-gpu=…` on NVIDIA; AMD equivalent on Adrenalin) — sustained > 5 % engagement over a 60-s window | Scale up host-side encode pipeline (the encoder is saturating, not the network); raise the encoder priority via `chrt -f 50 <pid>` through `r18.SafeExec` (allow-listed) | Alert `host.encode_saturated`; mark the session for host migration at the next convenient boundary |
| F5 | HDMI 2.1 ALLM not signalled (TV's image-processing pipeline still engaged) | Receive-to-display latency tax of 30–80 ms (motion smoothing, HDR tone-mapping); end-user reports feeling sluggish | TV-side image-processing observed via the receive-to-display histogram crossing the C12 §6 ALLM threshold | Re-send HDMI-CEC ALLM signal via the C12-owned retry path (this chapter calls `c12.ALLMRetry()` rather than re-implementing) | Log `display.allm_not_signalled`; the session continues with the latency tax, end-user is notified, alert fires for the operator |
| F6 | Frame-gen artefacts visible (DLSS-G / FSR-G / XeSS-G mis-prediction at scene cuts) | Visible double-image, warp, or judder at panning + cut transitions | Per-frame SSIM / VMAF check on candidate interpolated frames vs last fully-decoded frame; sustained drift below SSIM 0.92 over a 5-s window | Disable frame-gen for the current session; emit `display.framegen_disabled{cause="ssim-drift"}` | Log + alert; session continues at base framerate without interpolation — a tier-degradation event but not a session bounce |
| F7 | Composite mode fallback (FSE acquisition failed on Windows) | Cannot enter Fullscreen Exclusive; Windows refuses the mode-set request | FSE acquisition error returned by the D3D12 swap-chain `SetFullscreenState` call; capability schema field `fse_available=false` flips at runtime | Switch to Windows 11 22H2+ "Optimized for windowed games" mode — preserves VRR + ALLM at 1-frame extra latency | Composited mode (DWM-mediated, additional 1-frame budget); log `display.composite_fallback`; tier-degraded but session continues |
| F8 | DisplayPort / HDMI cable does not support VRR (older cable, certified for HDMI 2.0 not 2.1) | VRR negotiation never completes; client sees fixed-rate scanout despite both endpoints being capable | Capability-schema check fails at admission: declared `vrr_supported=true` on the display but `vrr_active=false` after attempted engagement | Alert the user via the client UI to upgrade to an HDMI 2.1 / DisplayPort 1.4a-certified cable; emit `display.cable_inadequate{required="HDMI 2.1"}` | Continue in fixed-rate VSync mode; the session admits but a tier-degraded marker rides with it |
| F9 | Refresh-rate change drops the connection (some monitors flicker or blank) | Display goes black for 2–5 s when the host requests a rate change; some monitors never come back | Timeout on rate-change `xrandr` call (>3 s wall-clock); the client polls for display reattachment | Revert to last-stable rate via `r18.SafeExec`-mediated `xrandr --output <out> --rate <last-stable>`; emit `display.rate_revert` | Stay at the default rate for the remainder of the session; log `display.rate_revert_persistent`; per-tenant ceiling drops by one tier |
| F10 | `r18.SafeExec` rejects an attempted `xrandr` call (allow-list mismatch — developer used `--off`, `--auto` without target, or some DDC/CI blank command) | Bootstrap or runtime mode-set fails; the structured error includes the rejected argv | The wrapper's regex check at the `os/exec` boundary returns `ErrHostDisruptiveCommand` (or `ErrForbiddenArgvShape`) with the offending argv | Fix the call site to use the allow-listed argv shape; allow-list extension requires operator review per Constitution §11.5.4 | **Blocking** — refuse refresh-rate change; non-overridable; the rule lives in the Constitution and bypass requires a §13 exception |
| F11 | Multi-monitor setup confuses VRR (G-Sync historically allowed only one display per system) | VRR engages on one display and not the other; tearing visible on the second display only | Capability-schema check: monitor count > 1 + VRR-on; the negotiation reports `vrr_active=true` for one output, `vrr_active=false` for the other | 2026 NVIDIA + AMD drivers support multi-display VRR; check driver version, prompt for upgrade if pre-2026 driver detected | Alert the user; force single-display VRR (the second display falls back to fixed-rate VSync); cross-link OQ-C22-01 |
| F12 | Wayland compositor does not respect VRR (some compositors lag or fail to honour mode-set) | Latency observability shows the receive-to-display tax that should be eliminated by VRR is not eliminated | Latency observability: receive-to-display p999 stays > 5 ms even after `vrr_active=true` is reported by the compositor | Switch to a known-good compositor (KWin 6.0+, GNOME Mutter 47+); the allow-list is published at session-bootstrap time and the client surfaces a "compositor not on allow-list" warning | Log + alert `display.compositor_not_allowlisted{compositor=…,version=…}`; cross-link OQ-C22-04 |

The table interlocks with the kill-switch hierarchy that C13 §13
establishes. The display-sync layer is **Layer 0-adjacent**: when the
swap chain is healthy, the latency budget is bounded; when it
degrades, the encoder + ABR layers pick up the slack until the
display primitive recovers. The display layer **never** escalates
to Layer 3 — Constitution §11.5 bars host-disruptive actions from
this stack, and F10 is the explicit SafeExec trip-wire that
enforces the bar.

## 8. Test surface

Every executable file in the display-sync submodule MUST be covered
by all ten test types listed in Constitution §6.1, plus the
inherited host-integrity-scan from C08 §12.11. The mock-allowed list
is **only Unit** (Constitution §6.2 / R-12); every other type drives
the real container topology with real EDID parsing on a real display
passthrough, real `xrandr` rate-change calls through `r18.SafeExec`,
real VRR engagement on the test fleet's G-Sync / FreeSync / HDMI 2.1
panels, and real producer/consumer streaming pairs across the
canonical bench host. The test surface below enumerates the binding
between each test type and the implementation contract enumerated in
§6 (the negotiator + VRR controller + frame-gen mediator). The full
per-type chapters live under `../07_Testing/` (queued).

### 8.1 Unit (mocks/stubs/hardcoded values permitted — R-12)

- `display.RefreshRateNegotiator` EDID-parsing test with mock EDID
  blobs covering: G-Sync Ultimate (NVIDIA range 1–360 Hz),
  FreeSync Premium Pro (AMD 48–240 Hz), HDMI 2.1 VRR (TV with ALLM
  +VRR — LG OLED-class), DisplayPort 1.4a Adaptive Sync (basic
  range 48–144 Hz), vanilla 60 Hz fallback (no VRR). Property-based
  testing via Go's `testing/quick` over arbitrary mode lists.
- `display.VRRController` enable/disable state-machine: assert the
  legal transitions (off → negotiating → active → degraded →
  off) and assert the illegal transitions (e.g., active →
  negotiating without going through degraded) return
  `ErrIllegalTransition` rather than corrupting state.
- Cache-line padding constant unit test for shared
  observability counters: assert `unsafe.Sizeof(display.PaddedCounter{})
  >= 64` on x86-64 build tags and `>= 128` on `arm64` build tags.
  Negative leg: remove the padding; assert the test fails — Constitution
  §6.3 mandates the negative leg.

Mock-allowed scope: **only the Unit lane** may use mocks/stubs/
hardcoded values. Cited under Constitution §6.1 / §6.2.

### 8.2 Integration

Real EDID read on a test container with display passthrough; verify
the mode list returned by `display.RefreshRateNegotiator.Probe()`
matches the expected mode list for the fixture display. The
container's `cap-add` list per Constitution §11.5.2 includes
`CAP_SYS_ADMIN` for display-passthrough — no `--privileged`,
no host-root mount.

Real `xrandr --output <out> --rate <hz>` invocation through
`r18.SafeExec`; verify the rate change is observed on the
passthrough display via a follow-up `xrandr --query` parse plus a
photodiode-style frame-time histogram that confirms the new cadence
within the per-tier SLO.

The lane also exercises the bootstrap protocol end-to-end: the
client probes the display, parses EDID, advertises the capability
schema, requests VRR engagement, and verifies the swap chain reports
`vrr_active=true` after the negotiation completes.

No mocks. Tests boot the full container topology via the Containers
submodule.

### 8.3 End-to-End (E2E)

Full host-agent + game + capture + encode + 4K120 stream + VRR-
capable client display on the canonical bench host (cross-link
`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md` §12.3).
Assertion: **no tearing across 1 hour of continuous gameplay** as
verified by a high-speed-camera fixture (240 fps minimum; 1000 fps
preferred per `latency_dim10.md` §1) capturing the panel and
post-processing the frames for tear-line detection. The 1-hour
duration catches drift that a 60-s smoke test misses.

Secondary assertion: the C24-canonical receive-to-display p999 ≤
5 ms holds across the full hour, with ≥ 10 K samples per
Constitution §6 / latency Insight #2.

No mocks — Constitution §6.2.

### 8.4 Security

- Verify `r18.SafeExec` rejects forbidden `xrandr` argv shapes
  (`xrandr --off`, `xrandr --auto` without target mode, DDC/CI blank
  commands) and forbidden `system_profiler` argv shapes
  (`system_profiler SPHardwareDataType` is not on the allow-list —
  only `SPDisplaysDataType` is). The fuzzer feeds the wrapper a
  corpus of allow-listed-but-mutated argvs; the wrapper MUST reject
  every mutation outside the allow-list with `ErrForbiddenArgvShape`.
  Cross-link C08 §12.4's deny-list bypass attempts.
- **Fuzz EDID blobs** with malformed data (truncated headers,
  invalid checksum bytes, oversized DTD blocks, malformed CTA-861
  extension blocks); assert `display.RefreshRateNegotiator` rejects
  every malformed input with a structured error and does NOT panic.
  The Go fuzzer (`go test -fuzz`) runs against the EDID parser as
  the canonical input surface.
- **auditd integration test** — boot the display-sync submodule's
  test container with an `auditd` rule auditing every `execve(2)`
  and every display-API ioctl call; run the full Ten-test-type
  matrix; grep the audit log for any §11.5.1 forbidden pattern.
  Zero matches is the gate.

No mocks. Tests use real attacker-pattern fuzz inputs and real
`auditd` boot-test instrumentation.

### 8.5 Benchmarking

`go test -bench` measuring p50 / p99 / p999 of:

- **Display-side scanout latency on OLED vs LCD test displays** at
  the canonical refresh-rate matrix (60 / 120 / 144 / 240 / 360 Hz);
  reports ≥ 10 K samples per Constitution §6 / latency Insight #2.
  The dim10 testing/validation file at
  `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_dim10.md`
  (see §5 "Statistical Rigor" — minimum sample size of 10 K
  measurements + p99/p999 over averages) is the canonical citation
  for the floor.
- **Frame-gen latency cost on DLSS-G / FSR-G / XeSS-G**: measure the
  end-to-end input-to-glass cost with frame-gen on vs off, on the
  bench host's RTX 4090 (DLSS-G), Radeon RX 7900 XTX (FSR-G), and
  Arc A770 (XeSS-G) test fixtures. The expected delta is ~1 frame
  at the host's native rate; regressions beyond 1.5 frames fail the
  build.
- Cross-link to C24 `10_Latency_Testing_and_Validation.md` for the
  canonical histogram pipeline and the Prometheus 3.x native-
  histogram exposition spec.

The benchmark report format follows the C24 canonical histogram
pipeline. Per-tier histogram archives are uploaded to the operator's
local artifact store (Constitution §3.3 local CI/CD); regressions
are detected by the benchmark-CI scan that fails the build if any
percentile crosses the per-tier budget by more than 5%. Average-only
benchmarks are merge blockers (Constitution §6.1 + latency Insight
#2).

### 8.6 Chaos

Synthetic fault injection at each layer:

- **Force monitor disconnect mid-stream** (HDMI hot-unplug via the
  bench host's relay-controlled HDMI switcher); assert the client
  reconnects within 500 ms wall-clock, assert VRR is re-enabled on
  reconnect (the F1 mitigation path engages automatically).
- **Force frame-rate drop below VRR range** by injecting synthetic
  CPU load on the consumer-side game-engine thread via `stress-ng
  --cpu N --cpu-load 90 --timeout 30s` (with explicit `--memory`
  cap per Constitution §11.5.3); assert LFC engages on the client
  display, assert no tearing is observed by the high-speed-camera
  fixture.
- **Cable degradation simulation** — physically swap from an HDMI
  2.1-certified cable to an HDMI 2.0-only cable mid-session via the
  bench host's HDMI switcher; assert F8's mitigation engages and
  the session continues at fixed-rate VSync.

Chaos lanes use the container topology from
`../07_Testing/07_Chaos.md` (queued).

### 8.7 Stress

Run the VRR-mode 4K120 stream for **24 h sustained** on a fixture
host with:

- The host-side encoder pinned to NUMA-local cores via
  `taskset -pc <numa-cpu-mask> <pid>` (gated by `r18.SafeExec`).
- The client running on a G-Sync Ultimate test fixture with the
  panel set to 144 Hz (within the VRR range 48–144 Hz).
- A continuous high-speed-camera fixture recording the panel
  output, with a tear-line detector running over the recorded video.

Assertions over the 24-hour run:

- **No rate-negotiation drift** — the VRR range advertised at hour
  0 matches the VRR range advertised at hour 24.
- **No LFC oscillation** — the LFC engagement counter does not
  oscillate (a sustained 0–5 % engagement is acceptable; a 0 → 100
  → 0 oscillation pattern fails the build).
- **No swap-chain leak** — the client's GPU memory baseline at hour
  24 matches the baseline at hour 0 within 1 %.
- **Drift-free p999** — the per-hour p999 receive-to-display
  histogram does not monotonically drift more than 5 % over the 24
  h (the C13 §14.7 drift SLO applies).

### 8.8 Smoke

Boot the client in a clean container with display passthrough; verify
the capability schema reports correct `refresh_rates`,
`vrr_supported`, `allm_supported`, `is_oled`, plus the derived
`vrr_range_min` / `vrr_range_max` fields. Total wall-clock ≤ 30 s.
Gates promotion (Constitution §6.1). Runs on every PR and every
container image build.

### 8.9 Full automation

All of §8.1–§8.8 run on every commit via the local container-driven
CI lane (Constitution §10 — local CI is the canonical gate).
Histograms are archived as native-histogram exports for trend
analysis; the `auditd` log + `strace -fe trace=execve` log from
§8.11 are archived alongside. No human input from clean checkout to
deployable artifact and back. Cross-link `../08_Operations/01_
Container_CI_CD.md` (queued) for the lane topology.

### 8.10 Challenges (production-like, full system up)

HelixQA dispatches a Challenges scenario where **4 concurrent client
sessions stream to 4 different displays**: a G-Sync Ultimate
monitor, a FreeSync Premium TV, an HDMI 2.1 VRR LG OLED, and a
vanilla 60 Hz fallback display. Assertion: per-display VRR
negotiation succeeds on each capable endpoint AND tear-free
presentation is sustained on all four for the duration of the run
(verified by the per-display high-speed-camera fixture). The
60 Hz fallback display admits at fixed-rate VSync without tearing
artefacts that would imply an inappropriate VRR engagement attempt.

Cross-link `../06_Submodules/04_HelixQA_Integration.md` (queued).

Failures stop the pipeline (Constitution §6.6); HelixQA findings are
normal P1/P2 work items mirrored on GitHub Projects + GitLab (R-17),
not advisory.

### 8.11 Inherited host-integrity-scan (R-18, non-overridable per Constitution §11.5.4)

Inherits VERBATIM from C08 §12.11: `strace -fe trace=execve` +
`auditd` boot test runs against the C22 implementation contract code
paths; asserts NO forbidden-command syscall (`reboot`, `kexec_load`,
`init_module`, `delete_module`, etc.) is invoked.

The inherited gate is:

```
zero matches across the entire log → PASS
≥ 1 match anywhere in the log     → FAIL (non-overridable)
```

Coverage extends to: the EDID-probe path (real `xrandr --query` /
`system_profiler SPDisplaysDataType` invocations), the VRR
engagement path, the rate-change path (`xrandr --output <out>
--rate <hz>`), the frame-gen mediator (DLSS-G / FSR-G / XeSS-G
toggle), and every operator-supplied script under
`vasic-digital/HelixPlayDisplay/scripts/`. The CI lane fails the
build on any §11.5.1 forbidden-pattern match.

## 9. Open questions

The questions below are tracked as `OQ-C22-NN` and feed back into the
master open-question register at `../00_Master_Plan.md` §10. They are
deliberately scoped to the display-sync layer and do not duplicate
host-side or input-side OQs (those live in C18 and C21 respectively).

- **OQ-C22-01** — *Multi-display split-screen.* Does HelixPlay
  support a "multi-monitor split-screen" mode for power users
  (driving two physical displays from a single session, each with
  its own VRR engagement), or is the MVP scoped to single-display
  only? The 2026 NVIDIA + AMD multi-display VRR support resolves
  the historical hardware constraint; the operator policy decision
  is whether to expose the feature in the MVP UI or defer to V1.
  Cross-link F11 mitigation path and the C13 Z5 differentiator
  scope.
- **OQ-C22-02** — *Frame-gen default.* Should HelixPlay enable
  DLSS-G / FSR-G / XeSS-G frame interpolation per-default, or
  per-operator-policy with default-off? Trade-off: visual fluidity
  vs the Conservative Prediction Paradox (C18) plus the 1-frame
  latency penalty plus the F6 SSIM-drift mitigation cost. The MVP
  position is default-off pending operator pilot data; V1 may flip
  the default after the SSIM threshold is empirically tuned.
- **OQ-C22-03** — *Per-tenant refresh-rate ceiling.* Does the
  operator-policy posture want to cap the streamed rate at 60 / 120
  / 144 / 240 Hz per-tenant, or always offer the panel maximum?
  The capacity-vs-quality trade-off is non-trivial: a 240 Hz tenant
  ceiling at scale costs significantly more host-side compute than
  a 144 Hz ceiling, but the 240 Hz tier is a marketing differentiator.
  The MVP exposes the ceiling as a per-tenant config; the default
  is to be decided by operator policy review.
- **OQ-C22-04** — *Wayland compositor allow-list.* Should HelixPlay
  maintain a known-good-compositor allow-list (KWin 6.0+, GNOME
  Mutter 47+, Hyprland latest stable) and refuse admission on
  unlisted compositors, or warn-and-admit? F12 mitigation currently
  warns-and-admits; a stricter posture (refuse admission) is more
  defensible for the < 1 ms VRR cost SLO but may surprise
  early-adopter Linux users. Cross-link F3 + F12.
- **OQ-C22-05** — *VRR over Web (Chrome / Firefox).* Web client VRR
  support in 2026 is partial: Chromium has experimental flags for
  VRR but no stable surface; Firefox lags further. Does HelixPlay
  capability-advertise to web clients differently (admit at
  fixed-rate VSync only, with a tier-degraded marker) or expose VRR
  behind an experimental flag? The MVP position is to admit web
  clients at fixed-rate VSync until Chromium ships a stable VRR
  surface.
- **OQ-C22-06** — *ALLM auto-retry policy.* When the C12 §6
  TV-side ALLM signal is lost mid-session (TV firmware update, HDMI
  re-handshake), should HelixPlay auto-retry the signal up to N
  times (current proposal: N=3 with exponential backoff) or fail
  the session immediately? The MVP position is N=3 with backoff;
  V1 may tune N based on TV-firmware-update telemetry.

Each OQ is tagged with a target-decision-date in the master register
and rolls forward into the Phase-12 (Latency Tuning) implementation
review (`../09_Implementation_Phases/Phase_12_Latency_Tuning.md`,
queued) where the operator review board ratifies the chosen posture
before code lands.

---

## 10. References

### Project artifacts

- Master Plan: [`../../00_Master_Plan.md`](../../00_Master_Plan.md). Constitution: [`../../01_Constitution.md`](../../01_Constitution.md) (§1.1 forbidden-pattern list; §6 quality — p50/p99/p999 ≥ 10 K samples; §11.5 R-18 enforcement). System Overview: [`../../02_System_Overview.md`](../../02_System_Overview.md) (§9 latency budget — display-side floor cited). Latency family index: [`00_Index.md`](00_Index.md). Sibling chapters cited above.

### Source research artifacts

- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_dim08.md` — 91 lines (primary).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md` — 2,199 lines (long-form synthesis).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_insight.md` — Insight #3 (Asymmetric optimisation) + Insight #5 (Conservative Prediction Paradox).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_cross_verification.md` — HC-03 + HC-09.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_dim10.md` — 128 lines (testing — §8.5 citation).

### Web research

[`../99_Web_Research_Addenda/2026-04-29-frame-pacing-and-vrr.md`](../99_Web_Research_Addenda/2026-04-29-frame-pacing-and-vrr.md) — 610 lines, 120 distinct URLs across 9 clusters + §Z contradictions index (Z1..Z8).

| Cluster | Topic | Cited in chapter |
|---------|-------|------------------|
| §A | NVIDIA G-Sync 2026 status (Z1, Z7) | §2.2 |
| §B | AMD FreeSync 2026 status | §2.3 |
| §C | HDMI 2.1 VRR + Adaptive Sync (Z5) | §2.4 |
| §D | HDMI 2.1 ALLM | §2.5 |
| §E | LFC + frame doubling (Z4) | §2.6 |
| §F | Frame interpolation client-side — DLSS-G / FSR-G / XeSS-G (Z2) | §4 |
| §G | Adaptive VSync + FastSync + tear-free presentation | §3 |
| §H | Display timing — VSYNC + buffer-flip + scanout | §5 |
| §I | 120/144/240/360 Hz feasibility 2026 + 4K240/4K480 OLED panels (Z6) | §5.4 |
| §Z | Contradictions index (Z1..Z8, including Z3 + Z8) | §1, §2, §4, §5, §7 |

### Sibling chapters (queued)

The sibling chapters cross-linked in the header are queued for synthesis under [Master Plan §7.2](../../00_Master_Plan.md#72-queued).

---

## Anti-Bluff Verification

> Per [Master Plan §4.3](../../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Source Evidence Reviewed

| Path | Lines (observed) | Reviewed by | Date | Used in §§ |
|------|-----------------:|-------------|------|------------|
| `04_Request.md` | 99 | A | 2026-04-29 | §1 |
| `01_base/01_Request.md` | 13 | A | 2026-04-29 | §1 |
| `02_latency/02_Response/Agent_results/research/latency_dim08.md` | 91 | A, B, C, D | 2026-04-29 | §§1–9 (primary) |
| `02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md` | 2,199 | A, B, C | 2026-04-29 | §§1–6 (long-form synthesis) |
| `02_latency/02_Response/Agent_results/research/latency_insight.md` | n/a | A, B | 2026-04-29 | §1 (Insight #3, #5) |
| `02_latency/02_Response/Agent_results/research/latency_cross_verification.md` | 101 | A, B | 2026-04-29 | §1 (HC-03, HC-09) |
| `02_latency/02_Response/Agent_results/research/latency_dim10.md` | 128 | D | 2026-04-29 | §8.5 (Benchmarking — ≥ 10 K samples) |
| `05_Response/00_Master_Plan.md` | post-Session-6 | A, B, C, D | 2026-04-29 | header / §6 / §9 |
| `05_Response/01_Constitution.md` | post §11.5 | A, B, C, D | 2026-04-29 | §§1, 2, 3, 4, 5, 6, 7, 8 (R-18 enforcement; §6 p50/p99/p999) |
| `05_Response/02_System_Overview.md` | 643 | A | 2026-04-29 | §1 (§9 budget — display-side floor) |
| `05_Response/04_Latency/00_Index.md` | 302 | A, B, C, D | 2026-04-29 | header voice alignment + cross-cutting trade-off matrix |
| `05_Response/04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md` | 1,749 | B | 2026-04-29 | §4.5 (Reflex 2 / Anti-Lag 2 / XeLL host-side complement) |
| `05_Response/04_Latency/07_Controller_Input_Optimization.md` | 1,127 | B | 2026-04-29 | §4.5 (Conservative Prediction Paradox) |
| `05_Response/03_Architecture/04_Go_Client_Ecosystem.md` | 3,336 | C | 2026-04-29 | §6.3 (client-tier UI bootstrap) |
| `05_Response/03_Architecture/07_Host_Agent_and_Game_Lifecycle.md` | 3,425 | C, D | 2026-04-29 | §6 (`r18.SafeExec` inheritance), §8.11 (host-integrity-scan inheritance) |
| `05_Response/03_Architecture/11_TV_UX.md` | 3,273 | A | 2026-04-29 | §2.5 (TV-side ALLM / HDMI-CEC cross-link) |
| `05_Response/03_Architecture/12_Latency_Engineering_Overview.md` | 3,816 | A | 2026-04-29 | §1 (C13 §3 + §7 + §8 + §9 cross-references) |

### Web Sources Consulted

The companion addendum
[`../99_Web_Research_Addenda/2026-04-29-frame-pacing-and-vrr.md`](../99_Web_Research_Addenda/2026-04-29-frame-pacing-and-vrr.md)
lists every URL with title and 2026-04-29 access date. **120 distinct URLs across 9 clusters + §Z.** Coverage shown in §10 References.

### Insights Incorporated

| Insight | Source file | Sections |
|---------|-------------|----------|
| latency Insight #3 — Asymmetric optimisation (CLIENT owns receive-to-display) | `latency_insight.md` | §1, §3.1 |
| latency Insight #5 — Conservative Prediction Paradox (frame-interpolation default OFF) | `latency_insight.md` | §1, §4.2, §4.5 |

### Conflict Zones Resolved

| CZ-ID | Description | Decision | Section |
|-------|-------------|----------|---------|
| HC-03 | Reflex + Frame Warp ≈ 75% perceived-latency reduction | **Reaffirmed (host-side cross-link C18 §5).** This chapter's §4.5 is the client-side complement | §1, §4.5 |
| HC-09 | VRR < 1 ms display-side cost | **Reaffirmed.** 2026 measurements (HDMI Forum + VESA + NVIDIA) confirm | §1, §2.1 |
| Z1 (NEW) | VRR range expansion (480/500 Hz top + 1 Hz bottom on 2026 OLED) | Documented; HelixPlay capability-advertises actual range | §2.7 |
| Z2 (NEW) | Frame Generation latency cost ~16 ms (NOT zero) | HelixPlay rule: frame-gen opt-in only; default OFF; Conservative Prediction Paradox + 1-frame penalty trade-off | §4 |
| Z3 (NEW) | Sunshine/Moonlight VRR partial | HelixPlay codifies VRR end-to-end as MVP differentiator (cross-link C13 Z5) | §1 |
| Z4 (NEW) | OLED VRR brightness flicker at < 60 Hz | LFC absorbs transient drops; encode pipeline targets ≥ 60 fps | §2.6 |
| Z5 (NEW) | HDMI 2.2 Ultra96 96 Gbps | V1 deferral; MVP scope is HDMI 2.1 | §1.2 |
| Z6 (NEW) | DisplayPort 2.1 UHBR20 panel availability | V1 deferral; capability-schema-future-proofed | §5.4 |
| Z7 (NEW) | G-Sync Pulsar variable-rate strobing | Capability-advertised, not gated | §2.2 |
| Z8 (NEW) | Wayland/Gamescope multi-monitor VRR brokenness | Known-good allow-list (KWin 6.0+, GNOME Mutter 47+, Steam OS Gamescope) — OQ-C22-04 | §7 F12 |
| Inherited (CZ-S1..CZ-S5, C07 Z-1..Z-7, C08 Z-1..Z-7, C09 Z, C10 Z-S1..Z-S5, C11 Z-1..Z-9, C12 Z-1..Z-5, C13 Z1..Z6, C15 Z-1..Z-7, C16 Z-1..Z-11, C17 Z-1..Z-9, C18 Z-1..Z-9, C19 Z-1..Z-13, C20 Z-01..Z-09, C21 Z-1..Z-9) | Owned by prior chapters | Not relitigated | header preamble |

### R-18 Operational Integrity Compliance

This chapter inherits R-18 enforcement from C08 (Constitution §2 DRY). Compliance verified:

- **Static — chapter prose**: §1.3 references R-18; §6.5 explicitly recaps the family-level allow-list extension specific to this chapter (`xrandr --output <name> --mode <mode> --rate <hz>`, `xrandr --listmonitors`, Windows `reg query HKLM\SYSTEM\CurrentControlSet\... /v RefreshRate`, macOS `system_profiler SPDisplaysDataType`).
- **Static — code in §6**: imports `r18 "github.com/vasic-digital/helix-r18-safeexec"` (origin C08 §10). The deny-list is **not duplicated** here — DRY.
- **Static — bootstrap subprocess invocations**: `xrandr` for refresh-rate change on Wails desktop tier; web/mobile/TV use OS-native APIs without subprocess. All wrap through `r18.SafeExec`.
- **Test — §8.11**: `host-integrity-scan` is **inherited** from C08 §12.11 verbatim (`strace -fe trace=execve` + `auditd` boot test). Non-overridable per Constitution §11.5.4.

The "self-referential mentions of forbidden patterns" (e.g. quoting `panic("not implemented")` in §6.4 to assert the code does NOT use it; quoting placeholder language in `R-02 (no bluffing, no placeholders)` Constitution citation; the `etc.` at the tail of the §8.11 forbidden-syscall enumeration which is verbatim-inherited from C08 §12.11) are explicitly permitted by Constitution §1.1 and Master Plan §5.2.3.

### Coverage Confirmation

| Metric | Value |
|--------|-------|
| Source per-dim primary (`latency_dim08.md`) | 91 lines |
| R-01 minimum (Master Plan §7.2 row C22) | 250 lines of body prose |
| Body prose actually synthesised | **1,285 lines** across §§1–9 (A 271 + B 226 + C 378 + D 410) |
| Coverage ratio vs minimum | 5.14× |
| Coverage ratio vs primary per-dim source | 14.12× |
| Forbidden-pattern scan (chapter prose) | clean (only legitimate self-referential mentions inside Constitution-cite text and the C08 §12.11-inherited `etc.` in §8.11) |
| Empty-section-body scan | clean |
| Tables | G-Sync vs FreeSync tier matrix in §2.2/§2.3; frame-pacing budget table in §3.7; failure-mode table 12 rows F1–F12 in §7; test-type matrix in §8 (Ten test types) |
| Section count | 10 normative sections (§§1–10) + this verification block |
| Go code blocks | §6.4 (~83 LOC across `display.NewRefreshRateNegotiator` constructor + `Negotiate` method + `Mode` type — real imports `golang.org/x/sys/unix` + `os` + `sort` + `context` + `errors` + `fmt` + `r18 "github.com/vasic-digital/helix-r18-safeexec"`; no stubs) |
| R-18 enforcement | inherited from C08 §10 (`r18.SafeExec` import — DRY) + §8.11 host-integrity-scan inheritance from C08 §12.11 |

### Sign-off

- Section A (§§1–2) executed by: subagent (C22 Group A) on 2026-04-29.
- Section B (§§3–4) executed by: subagent (C22 Group B) on 2026-04-29.
- Section C (§§5–6) executed by: subagent (C22 Group C) on 2026-04-29.
- Section D (§§7–9) executed by: subagent (C22 Group D) on 2026-04-29.
- Web research addendum compiled by: addendum subagent (C22) on 2026-04-29.
- Header, ToC, §10 References, and this Anti-Bluff Verification block stitched by: orchestrator (Claude) on 2026-04-29.
- Reviewed by: pending operator review.

End of `04_Latency/08_Frame_Pacing_and_VRR.md` — 2026-04-29.
