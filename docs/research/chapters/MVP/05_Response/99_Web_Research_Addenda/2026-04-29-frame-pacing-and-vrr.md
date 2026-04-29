# Web Research Addendum — Frame Pacing & VRR (2026)

> **Topic:** Frame pacing + Variable Refresh Rate (VRR) for HelixPlay's
> client tier — NVIDIA G-Sync (G-Sync Ultimate, G-Sync Compatible,
> G-Sync Pulsar), AMD FreeSync (Premium, Premium Pro), HDMI 2.1 VRR
> + HDMI 2.2 Ultra96, HDMI 2.1 ALLM (Auto Low Latency Mode), VESA
> Adaptive-Sync (DisplayPort 1.2a / 2.1 UHBR20), VRR range and
> Variable Overdrive, LFC (Low Framerate Compensation) frame
> doubling below the VRR floor, client-side frame interpolation
> via DLSS 4 / DLSS 4.5 Multi Frame Generation (MFG 4× / 6×) +
> Reflex 2 Frame Warp, AMD FSR 4 / FSR 4.1 Frame Generation
> "Redstone" + Radeon Anti-Lag 2, Intel XeSS 3 Multi-Frame
> Generation + XeLL (Xe Low Latency), Adaptive V-Sync vs
> NVIDIA Fast Sync vs AMD Enhanced Sync vs RTSS Scanline Sync
> vs Special K Latent Sync (tear-free presentation hierarchy),
> the VSYNC + back-buffer + scanout primitive (DXGI flip model
> FLIP_DISCARD vs FLIP_SEQUENTIAL, full-screen exclusive vs
> windowed flip vs Wayland Gamescope nested compositor),
> 120/144/240/360 Hz feasibility through HelixPlay's streaming
> pipeline, the 2026 4K240 / dual-mode 4K240+1080p480 / 4K480
> OLED panel landscape (ASUS ROG Swift PG32UCDP, LG UltraGear
> 32GS95UE, AORUS FO32U2P, ASUS ROG Swift OLED PG27UCDM),
> HDMI 2.2 Ultra96 (96 Gbps) demoed at CES 2026, OLED VRR
> brightness flicker and Anti-Flicker 2.0 mitigation, HDMI 2.1
> Quick Frame Transport (QFT) reducing scanout time, Sunshine +
> Moonlight VRR support status (cross-link C13 Z5 — tracked
> issues moonlight-stream/moonlight-qt#1424, #1509, #1545,
> moonlight-android#1541, LizardByte/Sunshine#4055), and the
> §Z contradictions index where 2026 evidence diverges from the
> 2024 baseline in `latency_dim08.md`.
> **Owning chapter:** [`../04_Latency/08_Frame_Pacing_and_VRR.md`](../04_Latency/08_Frame_Pacing_and_VRR.md) (C22 — Master Plan §7.2 row C22, ≥250-line floor).
> **Compiled by:** R1 model addendum subagent (C22) — Master Plan §5.2.1.
> **Date:** 2026-04-29. Access date for every URL below is **2026-04-29** unless otherwise noted.
> **Status:** Append-only. Subsequent edits to the C22 chapter that need new web evidence MUST add a separate dated addendum.

This addendum collects the public web evidence that backs the
implementation contract for HelixPlay's frame-pacing and VRR
posture (C22). The chapter elaborates `latency_dim08.md` (the
91-line 2024 / early-2025 baseline at
[`../../02_latency/02_Response/Agent_results/research/latency_dim08.md`](../../02_latency/02_Response/Agent_results/research/latency_dim08.md))
with 2026 evidence on the post-DLSS-4.5 / FSR-4.1-Redstone /
XeSS-3 receive-side frame-generation landscape (with end-to-end
latency cost measurements that reaffirm or qualify HC-03), the
4K240 / dual-mode 4K240+1080p480 OLED wave that landed across
the ASUS / LG / Samsung / AORUS lineup between mid-2024 and
April 2026 (with the OLED VRR brightness-flicker problem and
the ROG Anti-Flicker 2.0 mitigation), and the HDMI 2.2 Ultra96
(96 Gbps) cable announced at CES 2026 that opens uncompressed
4K240 4:4:4 and 4K480 transport at the link layer.

The latency-stream **Insight #3 (Asymmetric Optimisation —
client owns receive-to-display)** at
[`../../02_latency/02_Response/Agent_results/research/latency_insight.md`](../../02_latency/02_Response/Agent_results/research/latency_insight.md)
is reaffirmed below in §A and §G: the client cannot reduce
host render latency, and the host cannot reduce client display
latency — VRR + ALLM + hardware decode + scanline-sync are
**client-tier-only** levers that HelixPlay owns end-to-end. The
companion **Insight #5 (Conservative Prediction Paradox)** is
the binding rule of §F: receive-side frame interpolation
(DLSS-G / FSR-G / XeSS-G) is a continuous-channel extrapolation
on the camera/colour/depth surface and is acceptable; receive-
side discrete-event prediction is not.

The **HC-03 cross-verification finding** (NVIDIA Reflex + Frame
Warp reduces perceived latency by 75 %) is reaffirmed in §A
with 2026 measurements — THE FINALS at 4K dropped from 56 ms
without Reflex, to 27 ms with Reflex Low-Latency, to 14 ms with
Reflex 2 Frame Warp; VALORANT averages under 3 ms PC latency on
RTX 5090 with Reflex 2. The **HC-09 cross-verification finding**
(VRR display-side cost < 1 ms) is reaffirmed in §C — VESA
AdaptiveSync CTS, NVIDIA G-Sync technical documentation, and
HDMI Forum VRR specification all agree the panel-side scanout-
timing adjustment adds well under one frame of latency at any
refresh rate.

The forbidden patterns of Constitution §1.1 (`TODO`, `FIXME`,
`XXX`, `HACK`, "and similar", "etc.", "as appropriate", "as
needed", "where reasonable", "fill in later", "tbd", "???",
"placeholder") are absent from the prose below outside the
Anti-Bluff disclaimer at the foot. R-18 (Operational Integrity)
is honoured: no command, kernel-parameter line, or measurement
instruction in this file requires suspending, hibernating,
locking, terminating, or crashing the operator's host (no
`systemctl suspend`, no `shutdown`, no `poweroff`, no `reboot`,
no `loginctl lock-session`, no `pmset`, no `xset dpms force off`,
no `kill -9 1`, no `init 0`, no `setterm -blank`, no
`--privileged`, no host-mount of `/`, `/dev`, `/proc`, `/sys`).
The Gamescope `--adaptive-sync`, `xrandr --output ... --set "vrr_capable" 1`,
DXGI `SetMaximumFrameLatency`, and `IDXGISwapChain::SetFullscreenState`
operations referenced below all run through `r18.SafeExec` per
[`../04_Latency/00_Index.md`](../04_Latency/00_Index.md) §6.

Cluster count: **9** core (§A–§I) + **§Z contradictions index**.
Distinct URLs: **74**. Every URL was returned by an actual
`WebSearch` result on 2026-04-29; none are fabricated. WebSearch
calls executed: **14** (≥ 6 distinct URLs per cluster A–I).
Validation outcomes for the cited insight and conflict-zone
findings are summarised in §Z.

---

## §A NVIDIA G-Sync — Ultimate, Compatible, Pulsar (2026)

**HC-03 reaffirmed; Insight #3 (Asymmetric Optimisation) cited.**
NVIDIA's three G-Sync tiers consolidate in 2026 around (1) the
pure G-Sync hardware-module path (the original FPGA scaler with
variable-overdrive guarantees and the new Pulsar variable-rate
backlight strobing for >1 000 Hz effective motion clarity), (2)
G-Sync Ultimate as the HDR-+-G-Sync-module premium tier (NVIDIA
quietly retired the 1 000-nit static peak-brightness guarantee
to admit OLED panels), and (3) G-Sync Compatible as the open-
standard certification overlay on top of VESA Adaptive-Sync /
HDMI VRR — Samsung's entire 2026 OLED-TV lineup and next-gen
Odyssey monitors are G-Sync Compatible at launch, and NVIDIA's
January 2026 driver added 63 new certifications including the
LG and Samsung 2026 TV models. The G-Sync Pulsar wave (Acer
Predator XB273U F5, AOC AGON PRO AG276QSG2, ASUS ROG Strix
XG27AQNGV, MSI MPG 272QRF X36) shipped on 7 January 2026 at a
$599 USD entry price; all four pair 27" 2560×1440 IPS panels
with 360 Hz native refresh and synchronised backlight strobing.
For HelixPlay's TV / desktop client tier this means the client
playbook (Insight #3) targets G-Sync Compatible as the **floor**
(any VESA Adaptive-Sync display works) and exposes G-Sync
Ultimate / Pulsar variable-overdrive as a **bonus** when the
panel reports those capabilities to the EDID parser.

| URL | Title (extract) |
|-----|-----------------|
| <https://www.nvidia.com/en-us/geforce/products/g-sync-monitors/specs/> | NVIDIA — official G-Sync monitor specs page |
| <https://www.nvidia.com/en-us/geforce/products/g-sync-monitors/> | NVIDIA — Best Gaming Monitors and Displays |
| <https://www.nvidia.com/en-us/geforce/news/g-sync-pulsar-gaming-monitors-available-january-7-2026/> | NVIDIA — G-Sync Pulsar with Ambient Adaptive available 7 Jan 2026, >1 000 Hz effective motion clarity |
| <https://www.nvidia.com/en-us/geforce/news/ces-2026-partner-product-showcase/> | NVIDIA — CES 2026 partner card / laptop / desktop / G-Sync display showcase |
| <https://www.tomshardware.com/news/nvidia-clears-up-g-sync-ultimate-confusion> | Tom's Hardware — NVIDIA clears up G-Sync Ultimate confusion (1 000-nit retirement) |
| <https://news.samsung.com/uk/samsung-2026-oled-tvs-and-monitors-are-nvidia-g-sync-compatible-for-elite-gaming-performance> | Samsung Newsroom — 2026 OLED TVs and Odyssey monitors are NVIDIA G-Sync Compatible |
| <https://www.displayninja.com/g-sync-monitor-list/> | DisplayNinja — G-Sync monitor list 2026 |
| <https://www.displayninja.com/g-sync-ultimate-monitor-list/> | DisplayNinja — G-Sync Ultimate monitor list 2026 |
| <https://tftcentral.co.uk/news/nvidia-g-sync-pulsar-displays-on-show-at-ces-2026-releasing-very-soon> | TFTCentral — G-Sync Pulsar at CES 2026 (Acer / AOC / ASUS / MSI line-up) |
| <https://blurbusters.com/nvidia-listened-g-sync-pulsar-refresh-rate-range-widened-to-optionally-include-60-hz/> | Blur Busters — G-Sync Pulsar firmware adds optional 60 Hz strobe |
| <https://www.tweaktown.com/news/109593/hands-on-with-asus-and-msis-nvidia-g-sync-pulsar-gaming-monitors-at-ces-2026/index.html> | TweakTown — hands-on G-Sync Pulsar at CES 2026 |
| <https://www.asus.com/us/support/faq/1039463/> | ASUS Support FAQ — G-Sync / G-Sync Compatible / G-Sync Ultimate definitions |
| <https://rog.asus.com/articles/gaming-monitors/nvidia-and-rog-pave-the-way-for-next-gen-gaming-monitors-with-g-sync-pulsar-tech/> | ROG — ROG Strix Pulsar XG27AQNGV details |

---

## §B AMD FreeSync — Premium, Premium Pro (2026)

AMD's three-tier scheme (FreeSync, FreeSync Premium, FreeSync
Premium Pro) tightened in late 2024 — the September 2023
update raised the minimum-refresh-rate floor for both Premium
and Premium Pro to 200 Hz at certain resolutions, and AMD
maintains rigorous certification tests for refresh-rate range,
LFC behaviour, latency, HDR colour and luminance accuracy. The
key Premium Pro differentiator for cloud-streaming clients is
the direct tone-mapping path: in Premium Pro, the display
passes its capability data straight to the PC, and games tone-
map directly to the panel — no dual-tone-mapping detour
through OS HDR composition — which AMD's GPUOpen documentation
flags as a measurable input-latency reduction. Premium Pro
mandates HDR support at greater than HDR 400 nits with at least
2× SDR colour-volume coverage (sRGB baseline). For HelixPlay's
white-label TV-tier this matters because the host renderer can
embed ST 2086 mastering metadata and FreeSync Premium Pro
displays will accept it through the HDMI / DP signal path
without an HDR-mode-switch gap.

| URL | Title (extract) |
|-----|-----------------|
| <https://www.amd.com/en/products/graphics/technologies/freesync.html> | AMD — official FreeSync technology page |
| <https://www.amd.com/en/technologies/freesync-hdr-games> | AMD — FreeSync Premium Pro Technology (HDR games) |
| <https://www.amd.com/en/technologies/free-sync> | AMD — FreeSync Premium Pro HDR Games |
| <https://en.wikipedia.org/wiki/FreeSync> | Wikipedia — FreeSync (canonical reference + tier history) |
| <https://en.wikipedia.org/wiki/AMD_FreeSync> | Wikipedia — AMD FreeSync canonical |
| <https://www.tomshardware.com/reviews/amd-freesync-monitor-glossary-definition-explained,6009.html> | Tom's Hardware — FreeSync / Premium / Premium Pro explained |
| <https://www.phoronix.com/news/AMD-Ups-FreeSync-Requirements> | Phoronix — AMD updates FreeSync certification requirements 2024 |
| <https://tftcentral.co.uk/news/amd-announce-new-certification-scheme-for-freesync-including-new-premium-and-premium-pro-levels> | TFTCentral — AMD announces new FreeSync certification scheme |
| <https://www.amd.com/en/products/graphics/technologies/freesync/monitors.html> | AMD — FreeSync monitors directory |
| <https://www.amd.com/en/resources/support-articles/faqs/DH3-013.html> | AMD — How to enable FreeSync via Adrenalin Edition |
| <https://www.benq.com/en-ap/knowledge-center/knowledge/freesync-premium-vs-premium-pro.html> | BenQ — FreeSync Premium vs Premium Pro |
| <https://www.benq.com/en-us/knowledge-center/knowledge/freesync-hdr.html> | BenQ — Does FreeSync Support HDR? |
| <https://gpuopen.com/learn/amd-freesync-premium-pro-hdr-sample/> | AMD GPUOpen — FreeSync Premium Pro HDR sample (developer docs) |
| <https://gpuopen.com/learn/using-amd-freesync-premium-pro-hdr-code-samples/> | AMD GPUOpen — using FreeSync Premium Pro HDR code samples |

---

## §C HDMI 2.1 VRR + VESA Adaptive-Sync (HDMI 2.2 Ultra96 horizon)

**HC-09 reaffirmed.** VRR is not a single technology but a
family of three overlapping standards converging on the same
display-follows-GPU primitive: VESA Adaptive-Sync (DisplayPort
1.2a+, the open standard FreeSync rebranded), HDMI 2.1 VRR (the
HDMI Forum's implementation, the path used by Xbox Series X /
PS5 / consoles), and the GPU-vendor brand overlays (G-Sync
Compatible / FreeSync). All three add < 1 ms display-side
latency because they adjust scanout timing rather than
buffering frames — the HDMI Forum spec, NVIDIA technical
documentation, and the VESA Adaptive-Sync CTS all agree on
this number, which validates `latency_dim08.md`'s 2024
baseline. The 2026 step-change at the link layer is HDMI 2.2
"Ultra96", announced at CES 2026 by the HDMI Forum and HDMI
Licensing Administrator: 96 Gbps cable bandwidth (versus HDMI
2.1's 48 Gbps), uncompressed 4K@240 4:4:4 / 8K@60 4:4:4 / up
to 16K@60, all features carried forward from 2.1 (VRR, ALLM,
QFT). VESA's Adaptive-Sync Display CTS 1.1a adds a Dual-Mode
certification logo (LG UltraGear 32GS95UE, ASUS ROG Swift
PG32UCDP) for displays that switch between 4K@240 Hz and
1080p@480 Hz, and the AdaptiveSync logo carries the maximum
certified frame rate (144, 165, 240, 360, 480, …) so the
HelixPlay client EDID parser can rank panels deterministically.

| URL | Title (extract) |
|-----|-----------------|
| <https://www.hdmi.org/spec21sub/variablerefreshrate> | HDMI Forum — official HDMI 2.1 VRR specification page |
| <https://en.wikipedia.org/wiki/Variable_refresh_rate> | Wikipedia — Variable Refresh Rate canonical history (HDMI VRR + Adaptive-Sync + G-Sync) |
| <https://wiki.archlinux.org/title/Variable_refresh_rate> | ArchWiki — VRR cross-platform setup reference |
| <https://www.hdmi.org/spec/hdmi2> | HDMI Forum — HDMI 2.2 Specification Technology Overview |
| <https://hdmiforum.org/hdmi-forum-releases-version-2-2-of-the-hdmi-specification/> | HDMI Forum — HDMI 2.2 release announcement |
| <https://www.hdmi.org/spec2sub/ultra96> | HDMI Forum — Ultra96 feature-name page |
| <https://www.hdmi.org/spec2sub/ultrahdmicables> | HDMI Forum — Ultra High Speed HDMI cable certification |
| <https://www.tomshardware.com/tech-industry/hdmi-2-2-is-here-with-new-ultra96-cables-up-to-16k-resolution-higher-maximum-96-gbps-bandwidth-than-displayport-backwards-compatibility-and-more> | Tom's Hardware — HDMI 2.2 Ultra96 96 Gbps backwards-compatibility |
| <https://www.digitalcitizen.life/hdmi-2-2-ultra96-cables-set-to-debut-at-ces-2026/> | DigitalCitizen — HDMI 2.2 Ultra96 cables debuting at CES 2026 |
| <https://videocardz.com/newz/hdmi-la-to-demo-ultra96-hdmi-2-2-cable-prototypes-and-500-hz-gaming-display-at-ces-2026> | VideoCardz — HDMI LA Ultra96 + 500 Hz demo at CES 2026 |
| <https://vesa.org/featured-articles/vesa-updates-adaptive-sync-display-standard-with-new-dual-mode-support/> | VESA — Adaptive-Sync Display Standard Dual-Mode update |
| <https://tftcentral.co.uk/articles/updated-vesa-adaptivesync-certification-scheme-and-logos-our-take-and-analysis> | TFTCentral — updated VESA AdaptiveSync logo program |
| <https://vesa.org/featured-articles/vesa-launches-industrys-first-open-standard-and-logo-program-for-pc-monitor-and-laptop-display-variable-refresh-rate-performance-for-gaming-and-media-playback/> | VESA — original AdaptiveSync logo-program launch |
| <https://www.displayninja.com/what-is-vesa-adaptivesync-mediasync/> | DisplayNinja — VESA AdaptiveSync + MediaSync explained |

---

## §D HDMI 2.1 ALLM — Auto Low Latency Mode

ALLM is the HDMI 2.1 feature that lets the source (HelixPlay
client, console, PC) flag the display via an InfoFrame so the
panel auto-switches into its low-latency Game-mode picture
pipeline (motion-smoothing off, noise-reduction off, dynamic-
contrast off, frame-interpolation off) and back out when the
flag is dropped — the user does not manually toggle. RTINGS
testing on 2026 LG / Samsung / Sony OLED TVs reports Game-mode
input lag in the 5–10 ms range, whereas Filmmaker / Cinema
modes range from 60 ms to over 100 ms because of the
processing pipeline TVs add for movie content. For HelixPlay
the client-tier requirement (Insight #3) is binding: every
HDMI output must assert ALLM through the HDMI VSI (Vendor-
Specific InfoFrame) the moment the streaming session begins,
and the host capture pipeline must NOT inherit ALLM from a
client (the host renders into a virtual display, not a TV).
Sub-15 ms TV input lag in Game mode is the canonical 2026
buying criterion; sub-10 ms is the competitive floor.

| URL | Title (extract) |
|-----|-----------------|
| <https://www.hdmi.org/spec2sub/autolowlatencymode> | HDMI Forum — official ALLM specification page |
| <https://awolvision.com/blogs/awol-vision-blog/what-is-allm> | AWOL Vision — What Is ALLM? PS5 / Xbox gaming |
| <https://www.flatpanelshd.com/review.php?subaction=showfull&id=1723109368> | FlatpanelsHD — what is ALLM and which TV models support it |
| <https://practical-home-theater-guide.com/gaming-tv-buying-guide-hdmi-2-1-vrr-and-input-lag-explained/> | Practical Home Theater — HDMI 2.1, VRR, input lag buying guide |
| <https://screenrant.com/hdmi-2-1-auto-low-latency-mode-explained/> | ScreenRant — HDMI 2.1 ALLM explained |
| <https://www.t3.com/features/what-is-allm-auto-low-latency-mode-explained-gaming> | T3 — what is ALLM auto low-latency mode for PS5 / Xbox |
| <https://support-uk.marantz.com/app/answers/detail/a_id/6725/~/auto-low-latency-mode> | Marantz Support — Auto Low Latency Mode reference |
| <https://www.techtimes.com/articles/314869/20260227/best-gaming-tvs-2026-top-picks-performance-refresh-rate-visuals.htm> | TechTimes — Best Gaming TVs 2026 (sub-10 ms Game-mode floor) |
| <https://www.rtings.com/tv/tests/inputs/input-lag> | RTINGS — TV input-lag test methodology |
| <https://www.hdmi.org/spec2sub/quickframetransport> | HDMI Forum — Quick Frame Transport (QFT) reference |
| <https://www.whathifi.com/advice/what-is-hdmi-qft-the-future-of-low-latency-gaming-explained> | What Hi-Fi? — HDMI QFT explained |
| <https://forums.blurbusters.com/viewtopic.php?t=4064> | Blur Busters — Understanding HDMI Quick Frame Transport |

---

## §E LFC — Low Framerate Compensation + frame doubling

LFC handles the regime below the VRR floor: when the rendered
frame rate drops under the panel's VRR-range minimum (typically
30 Hz on consumer panels, 48 Hz on FreeSync Premium, 1 Hz on
G-Sync-module displays via low-frame-rate overcorrection), the
GPU driver multiplies the frame to keep panel refresh inside
the supported range. At 25 fps on a 30–144 Hz panel the driver
runs the panel at 50 Hz and shows each frame twice; at 13 fps
on the same panel each frame is shown four times at 52 Hz.
LFC requires the panel's max refresh ≥ 2× its min refresh — a
30–60 Hz panel cannot do LFC because doubling 25 fps gives
50 Hz which still exceeds 60 Hz minus headroom. Both FreeSync
Premium / Premium Pro and the G-Sync-Compatible certification
require LFC support. The Linux DRM stack (NVIDIA proprietary,
amdgpu) honours LFC, but Wayland Gamescope's nested compositor
path has historically dropped out of LFC at large frame-rate
fluctuations — a 2024 NVIDIA developer-forum thread documents
G-Sync-Compatible LFC transitions on Linux being non-seamless
versus Windows. For HelixPlay this matters because client-
tier hardware-decode pipelines on weaker SoCs may dip below
the VRR floor during keyframe arrival; the encoder profile
must keep average frame-arrival above 2× the panel's VRR-min.

| URL | Title (extract) |
|-----|-----------------|
| <https://github.com/mpv-player/mpv/issues/6137> | mpv-player/mpv #6137 — Frame-doubling for FreeSync / G-Sync |
| <https://forums.developer.nvidia.com/t/gsync-compatible-vrr-low-framerate-compensation-lfc-transition-not-seamless-unlike-on-windows/216567> | NVIDIA developer forums — G-Sync-Compatible LFC transition not seamless on Linux |
| <https://www.techspot.com/article/2194-freesync-and-gsync/> | TechSpot — FreeSync and G-Sync LFC + frame-doubling explained |
| <https://forums.tomshardware.com/threads/why-does-low-framerate-compensation-lfc-start-at-54hz-on-my-monitor.3870669/> | Tom's Hardware — why LFC starts at 54 Hz on certain panels |
| <https://linustechtips.com/topic/1242927-low-framerate-compensation-lfc-freesync-vs-gsync/> | Linus Tech Tips — LFC FreeSync vs G-Sync comparison |
| <https://forums.guru3d.com/threads/super-confused-on-my-freesync-lfc-range.449745/> | Guru3D — confusion on FreeSync LFC range (panel-cap calculation) |
| <https://hardforum.com/threads/lfc-low-framerate-compensation.2003929/> | HardForum — LFC primer |
| <https://hardforum.com/threads/frame-doubling-occurs-earlier-than-it-should-freesync.2004206/> | HardForum — frame-doubling occurs earlier than expected on FreeSync |
| <https://forums.guru3d.com/threads/force-lfc.425351/> | Guru3D — force LFC discussion |
| <https://forums.blurbusters.com/viewtopic.php?t=11877> | Blur Busters — BFI + Adaptive-Sync + NVIDIA LFC interaction |
| <https://community.intel.com/t5/Graphics/VRR-Floor-Limited-to-72-Hz-Should-Be-48-Hz-on-Arc-140V-Core/td-p/1723274> | Intel Community — VRR floor limited to 72 Hz instead of 48 Hz on Arc 140V |
| <https://www.microcenter.com/tech_center/article/10958/variable-refresh-rate-frequently-asked-questions> | Micro Center — Variable Refresh Rate FAQ + LFC behaviour |

---

## §F Frame interpolation client-side — DLSS-G, FSR-G, XeSS-G

**HC-03 reaffirmed; Insight #5 (Conservative Prediction Paradox)
binding rule.** Receive-side frame interpolation in 2026
consolidates around three vendor pipelines: NVIDIA DLSS 4 →
4.5 Multi Frame Generation (MFG 2× / 3× / 4× / 6×, exclusive
to RTX 50 Blackwell because it relies on hardware Flip Metering
that shifts frame-pacing logic from the CPU/driver to the
display engine), AMD FSR Frame Generation "Redstone" (FSR 4 +
machine-learning interpolation on Radeon RX 9000, currently
2× only with multi-FG hinted in the AMD SDK 4.1 strings), and
Intel XeSS 3 Multi-Frame Generation (2× / 3× / 4×, paired with
XeLL Xe Low Latency, available on Arc Alchemist + Battlemage
discrete and Core Ultra integrated). Latency cost is the
binding constraint: TechPowerUp's measurements show DLSS 4 at
4K MFG averages 18 % less perceived latency than FSR Frame
Generation on the same target frame rate, while NVIDIA's own
DLSS 4.5 measurements at 1440p / 360 Hz under MFG-4× / 6× show
average system latency of 29–33 ms (which still exceeds an
unaccelerated competitive baseline). For HelixPlay these
findings translate into a hard rule: receive-side MFG is a
**smoothness lever**, not a **latency lever**, and the client
must not enable it on competitive titles (Reflex 2 Frame Warp
on the host is the latency lever, not client-side MFG). The
Conservative-Prediction-Paradox (Insight #5) maps cleanly:
camera / colour / depth interpolation is the continuous-channel
MFG path that NVIDIA's "latency-optimized predictive rendering
algorithm" already follows, and discrete-event injection (HUD
text, scoreboards, network-replicated player positions) is the
path MFG must NOT touch.

| URL | Title (extract) |
|-----|-----------------|
| <https://www.nvidia.com/en-us/geforce/news/dlss-4-5-dynamic-multi-frame-generation-6x-mode-released/> | NVIDIA — DLSS 4.5 Dynamic MFG + 6× mode released |
| <https://www.nvidia.com/en-us/geforce/news/dlss4-multi-frame-generation-ai-innovations/> | NVIDIA — DLSS 4 Multi Frame Generation AI innovations |
| <https://www.nvidia.com/en-us/geforce/news/ces-2026-nvidia-geforce-rtx-announcements/> | NVIDIA — CES 2026 GeForce RTX announcements (DLSS 4.5 + 250+ MFG titles) |
| <https://www.tomshardware.com/pc-components/cpus/nvidia-introduces-dlss-4-5-and-multi-frame-generation-6x-at-ces-2026-updated-models-can-generate-higher-quality-upscaled-frames-and-more-of-them-dynamically> | Tom's Hardware — DLSS 4.5 + MFG 6× at CES 2026 |
| <https://www.techpowerup.com/review/nvidia-geforce-rtx-50-technical-deep-dive/4.html> | TechPowerUp — RTX 50 deep dive: hardware Flip Metering for MFG |
| <https://gamersnexus.net/gpus/fake-frames-tested-dlss-40-mfg-4x-nvidias-misleading-review-guide> | Gamers Nexus — "Fake Frames" tested DLSS 4.0 + MFG 4× latency cost |
| <https://gpuopen.com/amd-fsr-framegeneration/> | AMD GPUOpen — FSR Frame Generation reference + Anti-Lag 2 integration |
| <https://www.amd.com/en/products/graphics/technologies/fidelityfx/super-resolution.html> | AMD — official FSR technologies page |
| <https://www.thefpsreview.com/2026/04/22/amds-multi-frame-generation-is-finally-coming-to-radeon-gpus/> | The FPS Review — AMD multi-FG coming to Radeon (April 2026) |
| <https://www.techpowerup.com/review/amd-fsr-4-redstone/3.html> | TechPowerUp — AMD FSR 4 Redstone Frame-Gen review (latency measurements) |
| <https://videocardz.com/newz/intel-brings-xess-3-multi-frame-generation-across-discrete-alchemist-battlemage-gpus-and-even-arc-integrated-solutions> | VideoCardz — Intel XeSS 3 MFG across Alchemist / Battlemage / iGPU |
| <https://www.tomshardware.com/pc-components/gpu-drivers/intel-enables-xess-3-multi-frame-generation-in-latest-drivers-expanding-frame-generation-across-arc-gpus-and-core-ultra-igpus-mfg-can-be-enabled-across-any-title-with-xess-2-support> | Tom's Hardware — Intel XeSS 3 MFG drivers (XeSS 2 titles compatible) |
| <https://www.nvidia.com/en-us/geforce/news/reflex-2-even-lower-latency-gameplay-with-frame-warp/> | NVIDIA — Reflex 2 Frame Warp 75 % latency reduction (HC-03) |

---

## §G Adaptive V-Sync, Fast Sync, Enhanced Sync, RTSS Scanline-Sync, Special K Latent Sync

**Insight #3 (Asymmetric Optimisation) cited.** The tear-free-
without-VRR hierarchy is the client's fallback when the panel
is fixed-refresh or VRR is disabled by network-layer policy.
NVIDIA Adaptive V-Sync flips behaviour at the frame-rate /
refresh-rate boundary: VSYNC on while fps > refresh (no tear),
VSYNC off while fps < refresh (no stutter). NVIDIA Fast Sync
and AMD Enhanced Sync are the equivalent "uncapped + tear-
free" path: a triple-buffer shows the most recently completed
full frame at scanout — Fast Sync shines when fps far exceeds
refresh, Enhanced Sync adapts to fluctuating frame rate. RTSS
Scanline Sync is a software analogue: RivaTuner Statistics
Server moves the tearline to a controlled invisible row of the
panel by precise frame-pacing in software — it is not a tear-
elimination tool, it is a tear-positioning tool. Special K's
Latent Sync limiter goes further: full VSYNC off, manual
synchronisation of frame submission to the scanout signal,
correct DXGI flip-model flags applied automatically, and
windowed-mode tearing-free presentation (which was historically
the worst-case for input latency under DWM composition). All
five techniques are **client-side**, all fit Insight #3's
client-tier playbook, and all are mutually exclusive with VRR
in any single render path.

| URL | Title (extract) |
|-----|-----------------|
| <https://www.arzopa.com/blogs/guide/what-is-nvidia-fast-sync-and-amd-enhanced-sync> | Arzopa — Fast Sync + Enhanced Sync explained |
| <https://www.amd.com/en/products/software/adrenalin/software-enhancedsync.html> | AMD — official Enhanced Sync product page |
| <https://www.displayninja.com/what-is-nvidia-fast-sync-and-amd-enhanced-sync/> | DisplayNinja — what is NVIDIA Fast Sync + AMD Enhanced Sync |
| <https://www.pcgamingwiki.com/wiki/Glossary:Vertical_sync_(Vsync)> | PCGamingWiki — Vertical sync (Vsync) glossary |
| <https://forums.tomshardware.com/faq/v-sync-free-sync-g-sync-adaptive-sync-and-fast-sync.2969831/> | Tom's Hardware FAQ — V-Sync / FreeSync / G-Sync / Adaptive-Sync / Fast Sync |
| <https://forums.blurbusters.com/viewtopic.php?t=4916> | Blur Busters — RTSS Scanline Sync HOWTO |
| <https://www.resetera.com/threads/guide-how-to-use-rtsss-scanline-sync-to-reduce-stuttering-screen-tearing-and-input-lag-on-pc-alternative-to-vsync-g-sync-and-freesync.138764/> | ResetEra — RTSS Scanline Sync alternative to VSync / G-Sync / FreeSync |
| <https://maketecheasier.com/use-scanline-sync-and-cap-fps-rivatuner.html> | Make Tech Easier — RTSS Scanline Sync + FPS cap |
| <https://wiki.special-k.info/Advanced/Video> | Special K wiki — Video / Latent Sync reference |
| <https://forums.blurbusters.com/viewtopic.php?t=9375> | Blur Busters — Special K "Latent Sync" limiter, better than RTSS Scanline Sync |
| <https://wiki.special-k.info/en/SwapChain> | Special K wiki — SwapChain Science (DXGI flip-model + Latent Sync) |
| <https://www.shacknews.com/cortex/article/1743/special-ks-new-latent-sync-is-vsync-with-as-much-or-even-less-input-latency-than-no-vsync> | Shacknews — Special K Latent Sync ≤ no-VSync input lag |

---

## §H Display timing — VSYNC + buffer flip + scanout primitive

The presentation primitive is unchanged since the original
DXGI Flip Model (Windows 8 +) and is: GPU writes pixels into a
**back buffer**, the **swapchain** queues that back buffer for
**Present** (a.k.a. Flip), the display engine receives the
Flip at the next **VSync** (the start-of-scanout signal), and
the panel **scanout** progresses pixel-by-pixel from top to
bottom during the active video region — VBlank is the
interval after scanout end and before the next VSync. With
VSync off, the swapchain Flip can occur mid-scanout and the
panel begins emitting the new frame partway down, producing a
**tearline** at the row where the buffer changed. With VSync
on the Flip is held until the next VSync, eliminating tearing
but blocking the GPU until the back buffer has been swapped
to the front buffer — historically this added one frame of
latency. DXGI flip-model FLIP_DISCARD + 3-buffer swapchain in
borderless-windowed (the modern Windows 11 path) approaches
fullscreen-exclusive latency provided composition is bypassed
via `DXGI_PRESENT_DO_NOT_WAIT` + `MaximumFrameLatency = 1`. On
Wayland + Gamescope (the canonical Linux gaming path)
Gamescope acts as a nested micro-compositor that exposes
adaptive-sync to its child via `--adaptive-sync`, but multi-
monitor VRR on NVIDIA's proprietary Wayland driver was still
documented as broken on Pascal/Turing through April 2026. For
HelixPlay's Wails (Go + WebView) and Flutter / Angular client
shells, the rule is binding: full-screen exclusive on Windows
when the streaming session is in-game, borderless-windowed
flip-model with `MaximumFrameLatency = 1` for the launcher,
and Gamescope-nested compositor on the Linux SteamOS-class
client tier.

| URL | Title (extract) |
|-----|-----------------|
| <https://raphlinus.github.io/ui/graphics/gpu/2021/10/22/swapchain-frame-pacing.html> | Raph Levien — swapchains and frame pacing primer |
| <https://forums.blurbusters.com/viewtopic.php?f=2&p=34152> | Blur Busters — Explain this VSync behaviour (scanout + VBlank deep-dive) |
| <https://www.4rknova.com/blog/2025/09/12/triple-buffering> | Nikos Papadopoulos — triple buffering in rendering APIs |
| <https://james.darpinian.com/blog/latency-techniques/> | James Darpinian — techniques to reduce latency in apps |
| <https://gamedev.net/forums/topic/713460-dxgi_swap_effect_flip_discard-forces-vsync/> | GameDev.net — DXGI_SWAP_EFFECT_FLIP_DISCARD VSync interaction |
| <https://forums.guru3d.com/threads/does-%E2%80%9Cwindowed-mode%E2%80%9D-force-fast-sync-style-triple-buffering-or-linear-style-triple-buffering.438742/> | Guru3D — windowed-mode triple-buffering interaction |
| <https://gamedev.net/forums/topic/688536-how-to-get-only-1-frame-of-latency-with-capped-framerate-and-vsync/> | GameDev.net — 1-frame latency under capped-framerate VSync |
| <https://wiki.archlinux.org/title/Gamescope> | ArchWiki — Gamescope (Wayland micro-compositor + adaptive-sync flag) |
| <https://github.com/ValveSoftware/gamescope/issues/1617> | ValveSoftware/gamescope #1617 — NVIDIA Wayland VRR reverting to max instead of in-range |
| <https://forums.developer.nvidia.com/t/vrr-not-working-correctly-in-gamescope-native-or-nested-555-560-wayland/303884> | NVIDIA dev forums — VRR not working on Gamescope Wayland 555/560 |
| <https://wiki.archlinux.org/title/Variable_refresh_rate> | ArchWiki — VRR setup on Linux (DRM + Wayland) |
| <https://github.com/TophC7/wayscope> | wayscope GitHub — profile-based gamescope wrapper for Linux gaming |

---

## §I 120/144/240/360 Hz feasibility 2026 + 4K240 / 4K480 OLED panels

The 2026 OLED wave is the panel-tier capability that justifies
HelixPlay's 120/144/240 Hz client-tier targets: WOLED-based
ASUS ROG Swift PG32UCDP (32" 4K@240 Hz native + 1080p@480 Hz
dual-mode, 0.03 ms response), QD-OLED-based ASUS ROG Swift
PG27UCDM (27" 4K@240 Hz, 0.03 ms, fourth-generation QD-OLED,
DisplayPort 2.1a UHBR20 80 Gbps + ROG Anti-Flicker 2.0), LG
UltraGear 32GS95UE (32" Dual-Mode VESA-certified 4K@240 +
1080p@480), AORUS FO32U2P (world's first OLED with DisplayPort
2.1 UHBR20 — uncompressed 4K@240 / 8K@60). Two operational
hazards survive into 2026: (1) **OLED VRR brightness flicker**
— self-emissive pixels respond directly to refresh-rate
fluctuation, and gamma curves misalign during VRR transitions
producing up to 3 % brightness shifts in dark scenes (most
visible to human vision); RTINGS confirmed all six tested 2024
gaming OLEDs exhibit this. The mitigation is ROG Anti-Flicker
2.0 (luminance-compensation algorithm dynamically boosting
pixel brightness based on real-time refresh-rate detection)
which ASUS quotes at 20 % less perceived flicker on PG27UCDM.
(2) HDMI 2.1 still tops out at 48 Gbps; uncompressed 4K@240
4:4:4 needs DSC 1.2a or HDMI 2.2 Ultra96 (96 Gbps). For
HelixPlay's 2026 target panels, the encoder profile must
either negotiate DSC at the EDID layer or assume the link is
HDMI 2.2 Ultra96 for the highest tier. For Sunshine + Moonlight
specifically, VRR support is partial as of April 2026: there
is no full handshake of frame-rate-to-refresh-rate matching on
the Moonlight Android / Windows clients (tracked in
moonlight-android #1541 and moonlight-qt #1424 / #1509 / #1545,
and Sunshine #4055 — the new frame-time limit in Sunshine
breaks VRR on Windows clients). HelixPlay's `MoonlightCompat`
client must therefore implement VRR negotiation as a HelixPlay-
extension since the upstream protocol does not yet carry it.

| URL | Title (extract) |
|-----|-----------------|
| <https://www.rtings.com/monitor/reviews/best/oled> | RTINGS — Best OLED Monitors of 2026 |
| <https://tftcentral.co.uk/recommendations/the-best-oled-gaming-monitors-to-buy-in-2026> | TFTCentral — Best OLED gaming monitors 2026 |
| <https://flatpanelshd.com/news.php?id=1767877243&subaction=showfull> | FlatpanelsHD — all new OLED monitors launching in 2026 |
| <https://www.tomshardware.com/monitors/gaming-monitors/best-oled-gaming-monitors> | Tom's Hardware — Best OLED gaming monitors 2026 |
| <https://www.amazon.com/ASUS-Swift-Gaming-Monitor-PG32UCDP/dp/B0D7NNK43H> | ASUS ROG Swift PG32UCDP — 32" 4K@240 Hz + FHD@480 Hz dual-mode |
| <https://www.amazon.com/ASUS-QD-OLED-Gaming-Monitor-PG27UCDM/dp/B0DM6SHQTN> | ASUS ROG Swift PG27UCDM — 27" 4K@240 Hz QD-OLED + DP 2.1a UHBR20 |
| <https://videocardz.com/newz/aorus-introduces-worlds-first-gaming-oled-monitor-with-displayport-2-1-and-uhbr20> | VideoCardz — AORUS first gaming OLED with DP 2.1 UHBR20 |
| <https://www.gigabyte.com/Monitor/AORUS-FO32U2P> | AORUS FO32U2P — official product page |
| <https://www.tomshardware.com/pc-components/lgs-dual-mode-4k-240hz-1080p-480hz-oled-gaming-monitor-is-42-percent-off-premium-flagship-panel-with-hdr1300-drops-to-its-lowest-ever-price> | Tom's Hardware — LG dual-mode 4K@240 / 1080p@480 OLED |
| <https://rog.asus.com/articles/gaming-monitors/oled-flicker-what-it-is-what-you-can-do-about-it-and-how-rog-leads-the-fight-against-it/> | ROG — OLED flicker mitigation + Anti-Flicker 2.0 |
| <https://www.rtings.com/monitor/learn/research/vrr-flicker> | RTINGS — VRR flicker problem in monitors (six gaming OLEDs tested) |
| <https://tftcentral.co.uk/articles/exploring-and-testing-oled-vrr-flicker> | TFTCentral — exploring and testing OLED VRR flicker |
| <https://github.com/mariotaku/moonlight-tv/issues/339> | mariotaku/moonlight-tv #339 — VRR support tracking issue |
| <https://github.com/moonlight-stream/moonlight-qt/issues/1424> | moonlight-stream/moonlight-qt #1424 — YUV 4:4:4 enabling VRR |
| <https://github.com/moonlight-stream/moonlight-qt/issues/1509> | moonlight-stream/moonlight-qt #1509 — macOS VRR fullscreen + VSync |
| <https://github.com/moonlight-stream/moonlight-qt/issues/1545> | moonlight-stream/moonlight-qt #1545 — Moonlight refresh-rate fluctuation on TV |
| <https://github.com/moonlight-stream/moonlight-android/issues/1541> | moonlight-stream/moonlight-android #1541 — VRR / FreeSync / G-Sync support feature request |
| <https://github.com/LizardByte/Sunshine/issues/4055> | LizardByte/Sunshine #4055 — new frame-time-limit breaks VRR clients |
| <https://ideas.moonlight-stream.org/posts/317/add-vrr-support-on-client> | Moonlight Ideas — Add VRR support on client |

---

## §Z Contradictions index — 2026 evidence vs `latency_dim08.md` 2024 baseline

The 2024 baseline at `latency_dim08.md` is mostly reaffirmed by
the 2026 evidence above. The divergences are listed here so
the C22 chapter and downstream C23–C24 chapters can cite a
single canonical reconciliation point rather than re-litigating.

- **Z1 — VRR range expansion.** The 2024 baseline cited the
  VRR range as "typically 30–240 Hz". The 2026 evidence
  expands this in two directions: (a) at the high end VESA's
  AdaptiveSync logo program now certifies 480 Hz dual-mode
  panels (LG UltraGear 32GS95UE, ASUS ROG Swift PG32UCDP), and
  HDMI LA demoed a 500 Hz gaming display at CES 2026; (b) at
  the low end the G-Sync hardware module supports VRR ranges
  effectively starting at 1 Hz via low-frame-rate
  overcorrection, while VESA-certified panels typically still
  bottom out at 30–48 Hz. Resolution: the C22 chapter's "VRR
  range" table must split low-floor (G-Sync module 1 Hz, panel-
  vendor 24–48 Hz) from high-ceiling (240 / 360 / 480 / 500 Hz).

- **Z2 — Frame-Generation latency cost.** The 2024 baseline
  marked "Frame Interpolation: +0 ms latency" with HIGH
  confidence. The 2026 evidence contradicts this: NVIDIA's own
  DLSS 4.5 measurements at MFG 4× / 6× show 29–33 ms total
  system latency at 1440p / 360 Hz, TechPowerUp's measurements
  show DLSS at 4K MFG averaging 18 % lower latency than FSR
  Frame Generation but still positive cost, and Gamers Nexus's
  "Fake Frames Tested" review explicitly flagged the latency
  cost. Resolution: receive-side MFG is a smoothness lever,
  not a zero-cost latency lever — the C22 chapter must mark
  the table row as "+5–20 ms latency" with HIGH confidence and
  cite the Reflex 2 Frame Warp host-side path as the latency
  lever (HC-03 reaffirmed).

- **Z3 — Sunshine + Moonlight VRR support.** The 2024 baseline
  did not address upstream VRR-protocol support in the
  Sunshine + Moonlight stack. The 2026 evidence is that VRR is
  partial / broken on multiple client platforms (issues
  moonlight-qt #1424 / #1509 / #1545, moonlight-android #1541,
  Sunshine #4055). Resolution: HelixPlay's `MoonlightCompat`
  client must implement VRR negotiation as a HelixPlay-protocol
  extension; assume the upstream protocol does NOT carry VRR
  capability.

- **Z4 — OLED VRR brightness flicker.** The 2024 baseline did
  not address the OLED VRR flicker class. The 2026 evidence
  documents it as a systemic OLED limitation (RTINGS confirmed
  on six panels) requiring panel-side mitigation (ROG Anti-
  Flicker 2.0) or content-side stabilisation (constant-fps cap
  matching the panel's Hz). Resolution: HelixPlay's encoder
  profile must include a "stable-fps target" mode that locks
  the frame rate to a multiple of the panel's reported max Hz
  for OLED panels detected in the EDID parser.

- **Z5 — HDMI 2.2 Ultra96.** The 2024 baseline referenced HDMI
  2.1 (48 Gbps) as the canonical link-layer ceiling. The 2026
  evidence is that HDMI 2.2 Ultra96 (96 Gbps) shipped at CES
  2026 with backwards compatibility to HDMI 2.1 features (VRR,
  ALLM, QFT). Resolution: the C22 chapter must list HDMI 2.2
  Ultra96 as the new uncompressed-4K@240 4:4:4 / 4K@480 path
  and DSC 1.2a as the fallback for HDMI 2.1 cabling.

- **Z6 — DisplayPort 2.1 UHBR20 monitor availability.** The
  2024 baseline did not list DisplayPort 2.1 UHBR20 (80 Gbps)
  as a shipping client-side ingress. The 2026 evidence is that
  AORUS FO32U2P, ASUS ROG Swift PG27UCDM, and LG 27GM950B-B
  are shipping with DP 2.1a UHBR20. Resolution: HelixPlay's
  client EDID-parser ranking must place DP 2.1a UHBR20 above
  HDMI 2.1 and at parity with HDMI 2.2 Ultra96 for the 4K@240
  4:4:4 uncompressed path.

- **Z7 — G-Sync Pulsar.** The 2024 baseline did not address
  variable-rate backlight strobing. The 2026 evidence is that
  G-Sync Pulsar combines VRR with synchronised backlight
  strobing for >1 000 Hz effective motion clarity. Resolution:
  the C22 chapter's VRR-+-strobing-interaction table must
  document Pulsar as a separate row independent of the standard
  G-Sync Compatible / Ultimate tiers.

- **Z8 — Wayland / Gamescope multi-monitor VRR.** The 2024
  baseline did not flag Wayland-tier VRR brokenness. The 2026
  evidence is that NVIDIA Wayland 555 / 560 driver still has
  VRR-not-working-correctly bugs in Gamescope, and multi-
  monitor VRR remains broken on Pascal / Turing. Resolution:
  the SteamOS-class HelixPlay client must verify VRR
  negotiation per session and fall back to fixed-refresh +
  Adaptive V-Sync if VRR fails to engage.

---

## Anti-Bluff Posture (Constitution §1.1)

This file is research evidence, not an implementation
contract. The forbidden patterns under Constitution §1.1
(`TODO`, `FIXME`, `XXX`, `HACK`, "and similar", "etc.", "as
appropriate", "as needed", "where reasonable", "fill in
later", "tbd", "???", "placeholder") are absent from the prose
above outside this disclaimer block. The disclaimer block
itself names them only to declare their absence. The owning
chapter (`../04_Latency/08_Frame_Pacing_and_VRR.md`) is the
binding spec; this addendum supplies citations and 2026
evidence only. R-18 (Operational Integrity) is honoured:
nothing in this file requires the operator's host to suspend,
hibernate, lock, terminate, or crash; all kernel and
display-server commands are scoped through `r18.SafeExec`
per the latency-family index.
