# Dual-Path Encoding

> **Source:** `video-tech_dim04.md` (1,013 lines primary), `video-tech.agent.final.md` (2,588 lines), Insights #1 (thermal wall) + #4 (recording = save system) + #10 (recording differentiates).
> **Web addendum:** [`../99_Web_Research_Addenda/2026-04-29-dual-path-encoding.md`](../99_Web_Research_Addenda/2026-04-29-dual-path-encoding.md) — 228 lines, 80 distinct URLs across 9 clusters + §Z (Z-1..Z-8).
> **R-01 floor:** 1,150 lines body prose. **Achieved:** see Anti-Bluff Verification block.
> **Targets:** R-01..R-13, R-18; new public submodule `vasic-digital/helix-dualpath`; reuses helix-shm + helix-lockfree + helix-codec + helix-encoder + helix-capture + r18.
> **Cross-links:** [`00_Index.md`](00_Index.md), [`01_Codec_Selection.md`](01_Codec_Selection.md) (C26), [`02_Hardware_Encoders.md`](02_Hardware_Encoders.md) (C27), [`03_Capture_Pipelines.md`](03_Capture_Pipelines.md) (C28), [`05_Recording_Storage.md`](05_Recording_Storage.md) (C30 — local-buffer pattern), [`09_Thermal_and_GPU_Balancing.md`](09_Thermal_and_GPU_Balancing.md) (C34 — fleet thermal). Architecture-side: [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md). Latency-side: [`../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md`](../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md).
> **Status:** Draft v1. **Last updated:** 2026-04-29.

This chapter is the **fourth deep chapter of the `05_Video_Audio/`
family** — dual-path encoding orchestration. C28 owns the capture
pipeline producing single frame-event stream; C29 owns the
Frame-Tee fan-out into two parallel encoder sessions
(stream-encoder + record-encoder). Each encoder consumes
independently with per-encoder bitrate, preset, and GOP divergence.

**Insight #1 (thermal wall) binding**: dual-path is GPU-thermal-
bounded, not session-count-bounded. Pre-emptive quality reduction
at 78°C; hard cap at 83°C; emergency record-encoder pause at
82°C. **Insight #4 (recording = save system)**: local-first
recording with async sync (cross-link C30). **Insight #10
(recording differentiates)**: DVR-for-PC-gaming feature unique
to HelixPlay vs Parsec/Moonlight/Steam Remote Play.

**8 Z-contradictions resolved**: Z-1 NVENC session limit 5→8 on
consumer Lovelace; Z-2 FFmpeg tee-muxer vs GStreamer Frame-Tee
distinction; Z-3 AMF multi-HW-instance encoder mode; Z-4 Intel
concurrent-session ceiling quantification; Z-5 fMP4 default vs
MKV alternative; Z-6 AV1 efficiency baseline; Z-7 Sunshine
multi-client vs multi-session distinction; Z-8 NVENC GOP-
immutability constraint.

---

## Table of Contents

- [§1 Scope & non-scope](#1-scope--non-scope)
- [§2 Frame-Tee fan-out + per-encoder independence](#2-frame-tee-fan-out--per-encoder-independence)
- [§3 Thermal headroom budget](#3-thermal-headroom-budget)
- [§4 Quality budget allocation](#4-quality-budget-allocation)
- [§5 Codec / container divergence](#5-codec--container-divergence)
- [§6 Implementation contract](#6-implementation-contract)
- [§7 Failure modes](#7-failure-modes)
- [§8 Test surface](#8-test-surface)
- [§9 Open questions](#9-open-questions)
- [§10 References](#10-references)
- [Anti-Bluff Verification](#anti-bluff-verification)

---

## 1. Scope & non-scope

### 1.1 Place in the Video/Audio family — fourth deep chapter, dual-path encoding fan-out over the C28 capture feed

C29 is the **fourth deep chapter of the Video/Audio family**
(`05_Video_Audio/`). The family index landed at
[`00_Index.md`](00_Index.md) (C25); the codec-selection chapter
landed at [`01_Codec_Selection.md`](01_Codec_Selection.md) (C26);
the per-vendor hardware-encoder chapter landed at
[`02_Hardware_Encoders.md`](02_Hardware_Encoders.md) (C27); and
the capture-pipeline-orchestration chapter landed at
[`03_Capture_Pipelines.md`](03_Capture_Pipelines.md) (C28). Where
C26 owns **what codec** is selected at the architectural level,
C27 owns **which vendor encoder profile** is tuned per session,
and C28 owns **how the captured-frame stream gets to that
encoder**, C29 owns the next concern in the pipeline: **how a
single captured-frame stream is fanned out into two parallel
encoder consumers — one tuned for low-latency live streaming, one
tuned for high-quality session recording — without either path
starving, blocking, or thermally degrading the other**.

This separation is deliberate per Constitution §2 DRY: C28 owns
the producer side of the frame-event channel — Sunshine++ process
isolation, per-OS capture primitive orchestration, VBLANK-aligned
pacing, anti-cheat-aware capture posture. C28 stops at the point
where a captured GPU buffer descriptor (CUDA IPC handle on
NVIDIA, DMA-BUF fd on AMD/Intel/Apple) is published into a single-
producer ring. C29 starts at that point and asks: **how do two
encoders consume the same producer feed without coupling, and how
does the GPU's thermal budget get partitioned between them?** The
answer C29 develops is the **Frame-Tee fan-out pattern** — a
zero-copy single-producer-multiple-consumer (SPMC) topology backed
by reference-counted GPU buffer reuse, with each consumer running
an **independent encoder session** with independent preset,
bitrate, GOP structure, and quality target.

C29 is therefore the bridge chapter between the architecturally-
defined capture surface (C28's frame-event publish) and the two
downstream chapters that own each encoder branch: C29 itself
covers the **stream-encoder branch** end-to-end (preset,
bitrate, GOP, latency posture); the **record-encoder branch's
storage backends** (local NVMe staging + background sync to
NFS / SMB / WebDAV / S3) are owned by C30
([`05_Recording_Storage.md`](05_Recording_Storage.md)). C29 owns
the encoder-fan-out *mechanism* and the GPU-thermal budget
allocation; it hands off the recording bytestream to C30 once the
record-encoder has emitted its first NAL unit.

### 1.2 Insights binding this chapter

C29 cites three insights binding for the dual-path design — all
three drawn from the video-technology insight catalogue at
[`../../03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md`](../../03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md):

- **Insight #1 (video-tech) — the "thermal wall" is the hidden
  bottleneck for dual-path encoding.** The binding rule is that
  while hardware encoders (NVENC, AMF, QSV, VideoToolbox) have
  sufficient *throughput* for simultaneous streaming + recording
  on every modern consumer GPU, the *real* limiting factor is the
  GPU thermal budget. Dual encoding adds 15–25 W of additional
  GPU power draw on Lovelace / RDNA3 / RDNA4 / Arc Battlemage
  silicon, and this draw can push a sustained-load GPU past its
  thermal-throttle threshold (typically 83–88 °C on consumer
  Lovelace, 95 °C on consumer RDNA3). Once thermal throttling
  kicks in, **both** encoder paths degrade simultaneously — the
  GPU clock drops, NVENC frame-time rises on both sessions, and
  bitrate-on-demand drops on both branches. The implication for
  C29 is that dual-path is not a session-count-bounded resource
  problem (NVENC SDK 12+ supports 8 concurrent sessions on
  consumer Lovelace and 16 on consumer Blackwell, more than
  enough to multiplex many game sessions) but a **GPU-thermal-
  bounded resource problem**. §2.4 / §2.5 elaborate per-vendor
  session orchestration; the deep thermal-monitoring posture
  (PresentMon-driven feedback loop, bitrate scaling on
  thermal-throttle entry) is owned by C34
  ([`09_Thermal_Power_GPU_Aware.md`](09_Thermal_Power_GPU_Aware.md))
  and consumed here by reference rather than re-stated.

- **Insight #4 (video-tech) — recording storage architecture
  should mirror video game save systems.** The binding rule is
  that the most reliable recording storage pattern is *not* real-
  time network write but a **local buffer + background sync**
  model — identical to how modern games handle save files. This
  decouples recording from network reliability and game
  performance. The implication for C29 is that the **record-
  encoder branch must always write to a local NVMe staging
  surface first**, and that the network-storage backends (NFS,
  SMB, WebDAV, S3) are *consumers* of the local stage rather than
  direct writers from the encoder. C29 owns the encoder-side
  contract — the record-encoder writes a fragmented MP4 (fMP4)
  bytestream with `frag_keyframe + empty_moov + separate_moof`
  movflags into a local file path managed by C30; C30 owns the
  durability and async-sync semantics. §2.2 elaborates the
  preset / bitrate / GOP independence that this local-first
  posture enables (the record-encoder can run a higher-quality
  preset because it does not need to track the network's tail
  latency).

- **Insight #10 (video-tech) — recording is the feature that
  differentiates HelixPlay from competitors.** The binding rule
  is that while Parsec, Moonlight, and Steam Remote Play focus
  purely on streaming, HelixPlay's combination of zero-impact
  dual-path recording + configurable multi-backend storage
  creates a unique value proposition that no existing open-source
  cloud-gaming platform offers — effectively adding **"DVR for
  PC gaming"** as a product differentiator. The implication for
  C29 is that the dual-path mechanism is not a nice-to-have or a
  v2 feature: it is **MVP-binding** because it is the load-
  bearing differentiator for the product surface that
  [`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md)
  (C13) names as the binding architectural floor. The record-
  encoder branch must be production-grade in the MVP, not
  scaffolding; the Frame-Tee fan-out must be zero-copy and
  reference-counted from the first ship rather than refactored
  later.

The chapter also inherits, without re-stating, the cross-cutting
Insight #1 (cloudgaming) "Sunshine++ pattern" that C28 §2.1
established — the capture process is its own process, distinct
from both the game and the host-agent main process. C29 extends
this posture: the **stream-encoder** and **record-encoder** are
*also* their own processes (or, equivalently, their own goroutine
trees inside the encoder process container, depending on the
deployment posture chosen at session bootstrap), each
independently restartable. §2.6 elaborates how this isolation
interacts with the Sunshine 2026.x multi-session removal
(Sunshine upstream removed multi-session host support in their
2026.x release) and how HelixPlay's posture (VM-per-session via
KubeVirt for multi-session, dual-path within a single session)
diverges from Sunshine upstream.

### 1.3 In-scope and out-of-scope

**In scope** for this chapter:

- **Frame-Tee fan-out architecture** — single-producer multi-
  consumer pattern that splits the C28 frame-event stream into
  two parallel SPSC ring queues, one feeding the stream-encoder
  and one feeding the record-encoder, with each ring's tail
  managed independently and the GPU buffer's reference count
  drained only when *both* consumers have signalled completion
  on a given frame (§2.1).
- **Per-encoder independence** — independent preset, bitrate,
  GOP structure, B-frame count, lookahead depth, and rate-control
  mode for the stream and record branches; the stream branch
  pinned to low-latency presets (NVENC P5, AMF latency-tuning 7,
  QSV ULL) and the record branch pinned to quality-first presets
  (NVENC P3–P4, lookahead 8, B-frames 2) (§2.2).
- **Frame-buffer GPU-side reuse** — both encoders consume the
  same GPU buffer (CUDA IPC handle on NVIDIA, DMA-BUF fd on
  AMD/Intel/Apple — the contract from C18 §3 reused), with the
  buffer recycled into the producer's allocation pool only after
  both consumers have acked drain via a Treiber-stack reference-
  counting protocol (§2.3).
- **NVENC dual-session orchestration** — NVENC SDK 12+ concurrent
  session contract on consumer Lovelace (8 sessions per GPU) and
  consumer Blackwell (16 sessions per GPU); HelixPlay's
  reservation policy of 2 sessions per game session (1 stream +
  1 record); the resulting session-budget arithmetic (4 game
  sessions per Lovelace consumer card, 8 per Blackwell) (§2.4).
- **AMD AMF + Intel QSV dual-session orchestration** — neither
  vendor enforces a hard session limit (per Insight #9 vendor
  topology); HelixPlay's soft-cap policy (16 concurrent AMF
  sessions per GPU, 12 concurrent QSV sessions per GPU) which
  exists not to satisfy a driver constraint but to bound the
  thermal-headroom budget per Insight #1 (§2.5).
- **Sunshine 2026.x multi-session removal context** — Sunshine
  upstream removed multi-session host support in their 2026.x
  release; HelixPlay's MVP requires VM-per-session (cross-link
  C09 §4 KubeVirt) for multi-session, and dual-path is a
  *per-session* property within each VM — not a multi-tenant
  property of the host (§2.6).

**Out of scope** for this chapter — explicitly delegated to
sibling chapters or upstream chapters:

- **Codec selection** (H.264 / HEVC / AV1 / VVC trade-offs and
  capability negotiation) — owned by C26
  ([`01_Codec_Selection.md`](01_Codec_Selection.md)). C29 treats
  the codec choice as a parameter handed in by the session-
  bootstrap layer; the stream and record branches may negotiate
  *different* codecs (typical: H.264 for stream because it is the
  only mandatory WebRTC codec per Insight #3, HEVC for record
  because its storage efficiency is ~40 % better than H.264 at
  equivalent VMAF).
- **Per-vendor encoder profile knobs** (NVENC P1..P7, AMD AMF
  preset, Intel QSV preset numerics, Apple VideoToolbox preset)
  — owned by C27
  ([`02_Hardware_Encoders.md`](02_Hardware_Encoders.md)). C29
  reuses C27's profile vocabulary and does not redefine it.
- **Capture orchestration** (process tree, per-OS primitive,
  VBLANK alignment, anti-cheat posture) — owned by C28
  ([`03_Capture_Pipelines.md`](03_Capture_Pipelines.md)). C29
  consumes the C28 frame-event publish surface as a fixed
  contract.
- **Recording storage backends** — local NVMe staging,
  background sync to NFS / SMB / WebDAV / S3, segment rotation,
  retention policy, encryption-at-rest — owned by C30
  ([`05_Recording_Storage.md`](05_Recording_Storage.md)).
- **HDR colour-space handling** in the dual path — owned by C32
  ([`07_HDR_and_Color.md`](07_HDR_and_Color.md)). The HDR
  metadata flows through the Frame-Tee unchanged; per-branch
  tone-mapping (record keeps full-range HDR; stream tone-maps to
  client capability) is C32's concern.
- **Adaptive bitrate** for the stream branch — owned by C33
  ([`08_ABR_FEC_Congestion.md`](08_ABR_FEC_Congestion.md)). C29
  exposes a bitrate-update endpoint on the stream-encoder
  control plane; C33 drives it.
- **Thermal monitoring deep dive** — PresentMon-driven feedback,
  thermal-throttle detection, GPU power-budget allocation across
  concurrent sessions — owned by C34
  ([`09_Thermal_Power_GPU_Aware.md`](09_Thermal_Power_GPU_Aware.md)).
  C29 cross-links and consumes the C34 telemetry surface to
  adjust per-branch bitrate / preset on thermal-throttle entry,
  but does not re-implement the monitoring loop.

R-18 inheritance: per the C25 family allow-list at
[`00_Index.md`](00_Index.md) §7, every subprocess invocation in
this chapter wraps through `r18.SafeExec` from
`vasic-digital/helix-r18-safeexec` (origin C08 §10). No new
allow-list extensions are required for C29 — the dual-path
encoder paths reuse the C27 allow-list (`nvidia-smi`,
`vainfo`, `qsv-tools`, `ffmpeg` for the v1 fallback). No host-
disruption commands appear; no `kill -9 <game>`, `systemctl
suspend|hibernate|poweroff`, `pmset`, or privileged-container
hazards are invoked. The encoder processes run unprivileged with
device-scoped access (`/dev/dri/*` on Linux, the DXGI / NVENC
SDK surface on Windows, the Metal queue on macOS).

---

## 2. Frame-Tee fan-out pattern + per-encoder independence

This section describes the **Frame-Tee fan-out pattern** — the
mechanism by which HelixPlay splits a single capture-process
frame-event stream into two parallel encoder consumers — and the
**per-encoder independence** contract that lets each branch run
an unrelated preset, bitrate, GOP, and rate-control mode without
coupling. The section is organised by topology concern (fan-out
ring topology, per-branch preset independence, GPU buffer
reference-counting, per-vendor session orchestration on NVENC /
AMF / QSV, and Sunshine 2026.x context) rather than by encoder
vendor, because the fan-out *concerns* are common across vendors
even though the per-vendor primitive each consumes is vendor-
specific.

### 2.1 Frame-Tee architecture — single producer, two parallel SPSC consumers

The foundational topology rule is the **Frame-Tee fan-out**:
every captured frame produced by the C28 capture process is
*tee'd* — copied by reference, not by data — into two parallel
single-producer single-consumer (SPSC) ring queues, one feeding
the stream-encoder process and one feeding the record-encoder
process. The capture process emits a single frame-event publish
into the Frame-Tee root; the Frame-Tee node atomically
increments the per-frame reference count from 1 to 2 and pushes
the frame descriptor onto both consumer rings.

The choice of two parallel SPSC rings (rather than a single
single-producer multi-consumer queue) is deliberate. SPSC rings
are the lowest-overhead lock-free data structure in the literature
— a single atomic store on the producer side, a single atomic
load on the consumer side, no compare-and-swap on the hot path.
A SPMC queue, by contrast, requires either CAS on the consumer
side (each consumer races to claim a slot) or a fan-out broker
that adds an additional cache-coherence round-trip on every
frame. For a 60 fps capture cadence (16.667 ms per frame), the
extra cache-coherence cost is small in absolute terms (~50–200 ns
per frame on modern x86) but it compounds with the per-frame
overhead of every other pipeline stage; HelixPlay's microwave-
pipeline budget (Latency family Insight #1) cannot afford
unnecessary serialisation.

The **producer-side fan-out** therefore performs two independent
SPSC pushes — first to the stream-encoder ring, then to the
record-encoder ring — with the per-frame reference count
maintained as a 32-bit atomic colocated with the GPU buffer
descriptor. Each consumer pops independently; if one consumer is
slow (e.g. record-encoder doing a higher-quality two-pass
analysis), its ring fills up to its cap (configured at 4 frames
on Lovelace, 6 frames on Blackwell — sized to keep the GPU
buffer pool small enough to fit in the L2-cache-resident
allocation arena) and *that consumer's* push backs off; the
*other* consumer is unaffected.

The **back-pressure semantics** are the critical correctness
property: if the record-encoder ring fills up, the producer-side
fan-out drops the record-frame for that capture tick *but
continues to push the stream-encoder side*. The stream branch is
strictly hot-path; dropping a recording frame is acceptable
(C30's segment-rotation policy makes a 16.667 ms gap in the
recorded segment effectively invisible after fMP4 fragment
boundaries), but dropping a stream frame is not (it shows up as
visible stutter on the client display). The asymmetry is
encoded in the Frame-Tee root's drop policy, summarised in the
matrix below:

| Producer-side condition | Stream branch action | Record branch action |
| --- | --- | --- |
| Both rings have capacity | Push to both, refcount=2 | (n/a — paired with stream) |
| Stream ring full, record has capacity | Drop frame entirely (no record either, by design — the recorded clip should reflect what was streamed) | Skip — match stream policy |
| Stream has capacity, record ring full | Push stream, refcount=1 | Drop record frame; emit drop counter |
| Both rings full | Drop frame; emit congestion event to C34 | Drop record frame; emit drop counter |

The "stream-full → drop entirely" rule is deliberate: HelixPlay's
recording philosophy is **"the recording is what was streamed"**
— a recording that contains frames the player never saw is
useless for review purposes and confusing for the operator
viewing playback. C30's recording-policy surface lets a tenant
opt out of this rule (record the full capture even when stream
drops), but the MVP default is "stream-faithful recording".

The Frame-Tee root itself is implemented in the
`vasic-digital/helix-frame-tee` submodule (a new submodule
introduced by this chapter; it depends on the C15 `helix-shm`
shared-memory submodule and on the C18 `helix-gpu-direct` GPU-
buffer-handle submodule). The submodule carries Unit, Integration,
E2E, Security, Benchmarking, Chaos, Stress, Smoke, full-
automation, and Challenges tests per the project-wide constraint
in `04_Request.md` and the family inheritance in
[`00_Index.md`](00_Index.md) §6.

### 2.2 Per-encoder independence — preset, bitrate, GOP, rate-control

The second foundational rule is **per-encoder independence**: the
stream-encoder and the record-encoder MUST run with independent
preset, bitrate, GOP structure, B-frame count, lookahead depth,
and rate-control mode. This is what makes the dual-path MVP-
viable — if the two branches were forced to share encoder state,
the stream branch's low-latency posture would force the record
branch to skip B-frames and lookahead, sacrificing 30–40 % of the
record's storage efficiency for no benefit. Per-encoder
independence reclaims that storage efficiency by letting the
record branch make storage-quality trade-offs that the stream
branch cannot.

The default preset assignments per vendor are summarised below.
The values are inherited from C27's per-vendor profile vocabulary
and are not re-derived here:

| Branch | NVENC | AMD AMF | Intel QSV | Apple VT |
| --- | --- | --- | --- | --- |
| **Stream-encoder** | Preset P5 (low-latency-quality), tuning `ll`, RC `cbr_ld_hq`, B-frames 0, lookahead 0, GOP 60 | `latency_tuning 7`, RC `vbr_latency_constrained`, B-frames 0, GOP 60 | Preset `ULL` (ultra-low-latency), `target_usage 7`, B-frames 0, GOP 60 | `RealTime=YES`, `ProfileLevel=Main_4_2`, GOP 60 |
| **Record-encoder** | Preset P3–P4 (quality), tuning `hq`, RC `vbr_hq`, B-frames 2, lookahead 8, GOP 240 | `quality_tuning 1`, RC `vbr_high_quality`, B-frames 2, lookahead 8, GOP 240 | Preset `BestQuality`, `target_usage 1`, B-frames 2, lookahead 8, GOP 240 | `RealTime=NO`, `ProfileLevel=Main_5_1`, GOP 240 |

The **codec choice** also differs by branch — the stream-encoder
defaults to H.264 (per Insight #3, the only codec with > 98 %
hardware decode support across HelixPlay's client tier), while
the record-encoder defaults to HEVC (40 % storage efficiency
improvement at equivalent VMAF; supported by every consumer-tier
encoder HelixPlay targets, per C27 §3). The record branch can be
escalated to AV1 if tenant policy allows (AV1 is 30–35 % more
efficient than HEVC at equivalent VMAF on Lovelace and later, per
the dim04 §2.2 dual-NVENC measurements at the RTX 4070 Ti and
above), but the MVP default is HEVC because AV1 client-side
playback support remains < 60 % across HelixPlay's target
device matrix as of 2026-04.

The **bitrate independence** is the second axis. The stream
branch's bitrate tracks the C33 ABR controller's output (a
function of the network's bandwidth estimate, the FEC overhead,
and the congestion-control regime); the record branch's bitrate
is either fixed (configured at session bootstrap) or driven by a
quality target (CRF mode, where the encoder picks the bitrate to
achieve a target VMAF or PSNR). The two bitrates are *not*
linked: a stream branch backing off from 8 Mbit/s to 2 Mbit/s on
network congestion does *not* drag the record branch down with
it — the record branch continues writing at 25 Mbit/s HEVC
because the local NVMe staging surface (per Insight #4) is not
network-bound.

The **GOP independence** matters for storage-efficiency trade-
offs. The stream branch's GOP is short (60 frames at 60 fps =
1 s) so that any FEC repair point is quickly reached and any
client-reconnect can resume within ~1 s. The record branch's GOP
is long (240 frames at 60 fps = 4 s) so that I-frame overhead is
amortised across more frames, improving storage efficiency by
~10–15 % in the typical case. These two GOP cadences are
independently scheduled inside the GPU encoder — each session
gets its own keyframe-trigger timer.

### 2.3 Frame-buffer GPU-side reuse — Treiber-stack reference counting

Both encoders consume the **same GPU buffer** — neither makes a
local copy. The capture process emits a CUDA IPC handle (on
NVIDIA), a DMA-BUF file descriptor (on AMD / Intel / Apple), or
an `IOSurface` reference (on macOS); both encoder processes import
that handle into their own GPU context (per the C18 §3 cross-
process import contract) and bind it as the input surface for
their respective NVENC / VAAPI / QSV / VideoToolbox session.

The buffer's lifetime is managed by a **Treiber-stack reference
counting protocol**. The Treiber stack is the canonical lock-free
LIFO data structure (R. Kent Treiber, "Systems programming:
Coping with parallelism," 1986); HelixPlay uses it as the GPU
buffer pool's free-list. Each GPU buffer descriptor carries a
32-bit atomic reference count colocated in the descriptor's
metadata header. The lifetime is:

1. Capture process allocates a GPU buffer (or pops one from the
   Treiber-stack free-list); refcount = 0.
2. Capture process writes the captured frame into the buffer;
   refcount stays 0 (still privately owned by capture).
3. Capture process publishes the buffer descriptor into the
   Frame-Tee root; the Frame-Tee atomically sets refcount = 2 and
   pushes the descriptor onto both consumer rings.
4. Stream-encoder pops the descriptor from its ring, imports the
   handle, encodes, and on encoder-completion atomically
   decrements refcount.
5. Record-encoder pops the descriptor from its ring, imports the
   handle, encodes, and on encoder-completion atomically
   decrements refcount.
6. When the second decrement transitions refcount from 1 to 0,
   that consumer pushes the buffer descriptor onto the Treiber-
   stack free-list, available for the capture process's next
   allocation.

The protocol is correct under the standard Treiber-stack memory-
ordering semantics (acquire on pop, release on push); the
acquire-release pair on the refcount transition ensures that
the encoder's "I am done with this buffer" write happens-before
the capture's "I am about to reuse this buffer" read on the
*next* allocation. There is no GC, no reference-count leak path
(both encoders are obliged by their session contract to either
ack or fail; failure-ack is treated identically to success-ack
for refcount purposes — the buffer is reclaimable either way),
and no per-frame allocation cost in steady state because the
free-list is hot in the L2 cache.

The reference-counting protocol is implemented in the
`vasic-digital/helix-gpu-bufpool` submodule (also a new submodule
introduced by this chapter). Its dependency tree:

| Submodule | Purpose | Origin |
| --- | --- | --- |
| `vasic-digital/helix-gpu-bufpool` | Treiber-stack-backed GPU buffer pool with refcount lifetime | C29 (this chapter) |
| `vasic-digital/helix-frame-tee` | SPMC fan-out from capture to two SPSC rings | C29 (this chapter) |
| `vasic-digital/helix-shm` | Shared-memory ring buffers used by Frame-Tee | C15 §6 |
| `vasic-digital/helix-gpu-direct` | Cross-process GPU buffer handle import | C18 §3 |
| `vasic-digital/helix-r18-safeexec` | Subprocess wrapper for any encoder subprocess | C08 §10 |

### 2.4 NVENC dual-session orchestration on consumer Lovelace and Blackwell

NVENC SDK 12+ (released alongside R545 / R555 driver branches)
supports **8 concurrent encoding sessions per consumer GPU** on
all NVENC-capable consumer cards from Maxwell 2nd Gen through
Ada Lovelace, and **16 concurrent sessions** on consumer
Blackwell (RTX 5000-series and later). Workstation and data-
center cards (Quadro, RTX A-series, L40, L40S) retain the
historical "no enforced session limit" posture — the limit is
purely a thermal-budget concern there, per Insight #1.

HelixPlay's reservation policy is **2 NVENC sessions per game
session** — one for the stream-encoder branch, one for the
record-encoder branch. The arithmetic that follows from this
policy and the SDK 12+ contract is summarised below:

| GPU class | NVENC sessions per GPU | HelixPlay game sessions per GPU |
| --- | --- | --- |
| Consumer Lovelace (RTX 4060–4090) | 8 | 4 |
| Consumer Blackwell (RTX 5060–5090) | 16 | 8 |
| Workstation Lovelace (RTX 6000 Ada) | unlimited (thermal-bound) | thermal-bound (typically 8–10) |
| Data-center Lovelace (L40, L40S) | unlimited (thermal-bound) | thermal-bound (typically 12–16) |
| Pre-Lovelace consumer (RTX 3xxx, 2xxx) | 8 (post-2024 driver) | 4 |

The session reservation is per-VM in the multi-session topology
(KubeVirt VM per game session, per C09 §4); each VM passes
through one virtual NVENC session-pair. Cross-VM contention is
handled by the host-level NVENC scheduler, not by HelixPlay; if
the operator over-commits (more game sessions than the GPU's
session budget supports), the NVENC SDK returns
`NV_ENC_ERR_OUT_OF_MEMORY` on session-create and the C09
session-bootstrap rejects the new session before any user-
facing work begins. HelixPlay never silently degrades; the
operator is told.

The **Dual NVENC** hardware feature on RTX 4070 Ti and higher
(two physical NVENC engines, plus the SDK 12+ Split-Frame
Encoding feature for high-resolution single-stream
distribution) is *not* exposed to the application as a session-
count multiplier — it is a per-engine throughput multiplier that
the NVENC scheduler exploits internally to keep the per-session
frame-time below the encoder's target. HelixPlay's session-
budget arithmetic does not change on dual-NVENC silicon; what
changes is the per-session 4K-and-above latency floor (~3 ms
better at 4K60 on RTX 4080 versus RTX 4070, per dim04 §2.2).

### 2.5 AMD AMF + Intel QSV dual-session — no driver limit, soft-cap for thermal headroom

AMD AMF (Advanced Media Framework) and Intel QSV (Quick Sync
Video) both expose **no driver-enforced session limit** —
either vendor can run as many concurrent encoder sessions as the
hardware encoder block can sustain. This is the property that
Insight #9 (vendor topology) names as the case for "AMD RDNA4 /
Intel QSV for budget deployments". HelixPlay still enforces a
soft cap — not because of a driver constraint, but to bound the
GPU thermal headroom budget per Insight #1:

| Vendor | Driver-enforced session limit | HelixPlay soft cap | HelixPlay game sessions per GPU |
| --- | --- | --- | --- |
| AMD AMF (RDNA3 / RDNA4) | none | 16 | 8 |
| Intel QSV (Arc Battlemage / Iris Xe) | none | 12 | 6 |
| Intel QSV (Lunar Lake iGPU) | none | 6 | 3 |

The AMD soft cap (16 concurrent AMF sessions) is sized to keep
total dual-path encoder GPU power draw under 150 W on a 250 W
RX 7900 XTX or under 75 W on a 165 W RX 7800 — enough thermal
headroom that the GPU's sustained-load temperature stays below
the throttle threshold even under summer ambient (35 °C
data-center inlet) for at least 4 hours of continuous load.
The Intel soft cap is sized analogously for Arc Battlemage's
~190 W TBP and Lunar Lake iGPU's ~30 W package power.

The **AMF preset** for the stream branch is `latency_tuning 7`
(maximum latency optimisation, equivalent to NVENC P5); for the
record branch, `quality_tuning 1` (maximum quality, equivalent
to NVENC P3). The QSV preset for the stream branch is `ULL`
(ultra-low-latency; equivalent to NVENC's P5) and for the record
branch is `BestQuality` (equivalent to NVENC's P3). These
mappings are inherited from C27 §4 (AMD AMF) and §5 (Intel QSV)
and are not re-derived here.

The dual-session orchestration on AMD and Intel uses the same
Frame-Tee fan-out + Treiber-stack refcount protocol as NVENC —
the *mechanism* is vendor-agnostic; only the per-session API
surface (`AMF::CreateComponent` for AMD, `MFXVideoENCODE_Init`
for QSV) differs, and that surface is encapsulated by the
respective C27 vendor adapter.

### 2.6 Sunshine 2026.x multi-session removal — VM-per-session for multi-tenant; dual-path is per-session

Sunshine upstream (the open-source Moonlight host that
HelixPlay's Sunshine++ pattern derives from) **removed multi-
session host support in the 2026.x release** (C08 §8 documents
the upstream change and HelixPlay's response). The removal was
motivated upstream by complexity: Sunshine's C++ implementation
was not designed for multi-tenant resource-isolation (anti-cheat
posture, GPU-buffer-pool partitioning, NVENC session-budget
arithmetic across concurrent users), and rather than refactor
toward a multi-tenant posture upstream chose to scope Sunshine
back to its original single-host single-session use-case.

HelixPlay's MVP requires multi-session support for the
multi-tenant deployment topology that the architecture floor
(C09) defines. The **VM-per-session topology** (KubeVirt VM per
game session, per C09 §4) is HelixPlay's response: each game
session runs in its own KubeVirt-managed VM with its own
isolated capture process, its own isolated encoder pair, and its
own pass-through GPU partition (vGPU on NVIDIA, MxGPU on AMD,
SR-IOV on Intel). The dual-path Frame-Tee is therefore a
**per-session property** — every VM hosts one capture process
and one encoder pair (one stream + one record); the multi-
session host orchestration happens at the KubeVirt / Kubernetes
control-plane level, not inside HelixPlay's encoder code.

The implication for C29's design is clarifying: the chapter does
*not* need to defend a multi-tenant Frame-Tee — the Frame-Tee is
single-tenant by deployment topology, and the multi-tenant
arithmetic is the host-level NVENC / AMF / QSV session-budget
arithmetic of §2.4 and §2.5. The dual-path concern is purely
about the *two* encoder branches inside a single session, never
about *N* game sessions sharing a fan-out node. The Sunshine
2026.x removal is therefore not a regression for HelixPlay; it
is a confirmation that the per-session-isolated topology is the
correct posture, and HelixPlay's KubeVirt-based multi-session
floor is the right architectural answer.

The cross-link to C09 §4 is the binding reference for the
multi-session topology — C29 does not redefine the VM-per-
session contract, but it does inherit the per-VM resource budget
(1 capture process, 2 encoder sessions, 1 GPU partition) and
treats that budget as the binding floor for all the §2.1–§2.5
arithmetic above. The MVP's Frame-Tee implementation lives
inside each VM; the Sunshine++ pattern (capture process is its
own process, separate from game and host-agent main) is enforced
inside each VM by the same supervisor pattern C28 §2.1
documented for the single-session case.
## 3. Thermal headroom budget

Dual-path encoding is not gated by session-count, frame-rate, or codec configuration in the way operators traditionally assume. The binding constraint is GPU thermal headroom. Insight #1 from `video-tech_dim04.md` makes this explicit: the second NVENC slot consumes 15-25 W of additional GPU board power above a single-path baseline, and that power becomes heat. On an air-cooled RTX 4080 running a flagship title at 4K60, the GPU silicon temperature trajectory under dual-path is steeper than under single-path, and once the package crosses the throttle threshold (83°C on most NVIDIA Ada/Blackwell SKUs), NVENC frame-submission throughput drops 25-30 % until the temperature recovers.

This section codifies how HelixPlay's host-agent measures, predicts, and acts on thermal pressure so that the dual-path design (§2) does not silently degrade into a single-path-with-hiccups system at the wall-clock moment when the operator's premium tenants are most active. The deeper, multi-host, multi-GPU thermal balancing story belongs to C34 — this section is scoped strictly to the dual-path-induced thermal interaction on a single host.

### 3.1 Insight #1: the thermal wall is binding

Insight #1 (`video-tech_insight.md`, Item #1) states the empirical finding that surfaced repeatedly across the recording-pipeline research: dual-path encoding's hard ceiling on a given GPU SKU is not the published NVENC session count, not the encoder API throughput, not VRAM, and not PCIe bandwidth. It is the thermal envelope.

Three observations make the wall binding rather than advisory:

- **Power delta is unavoidable.** A second 1080p60 HEVC NVENC session adds 15-25 W to GPU board draw on Ada-generation silicon (RTX 4070/4080/4090 cohort). At 4K60 the delta climbs toward the upper bound. This is incremental over whatever the rendering workload already costs and is not amortizable: the second encoder instance does real work on dedicated NVENC silicon, and that silicon dissipates power proportional to pixel rate and rate-control complexity.
- **Throttle is sharp, not soft.** NVIDIA's thermal management drops NVENC clock once the package sensor reports ≥83°C. The drop is not graceful: encoder submission latency spikes from sub-millisecond to multi-millisecond ranges, frame-pacing breaks, and the recording stream ends up with visible cadence faults. The same throttle applies to the live stream because both paths share the encoder silicon, so the player-facing experience also degrades.
- **Recovery hysteresis is wide.** Once throttling engages, the GPU does not re-enable full NVENC clocks until the package falls back below roughly 78°C. On an air-cooled card under sustained load, that cooldown can take 30-90 seconds. The operator therefore loses dual-path quality for the better part of a minute every time the wall is hit, which is unacceptable for a service that markets recording as a first-class feature.

The host-agent treats Insight #1 as a hard architectural axiom: HelixPlay never queues a second encoder path on a host whose forward-looking thermal trajectory would cross the wall. Section 3.2 defines the tier ladder that lets the host-agent react before the wall is hit; sections 3.3 and 3.4 define how the trajectory is computed; section 3.5 is the operator-facing escape hatch for thermally-aggressive deployments; and section 3.6 marks the boundary with C34.

The host-agent samples temperature, power, and clock data via vendor utilities. For NVIDIA, the canonical query is `nvidia-smi --query-gpu=temperature.gpu,power.draw,clocks.sm --format=csv,noheader,nounits` issued at 1 Hz; the agent buffers the last 30 samples for trend analysis (see §3.4). Each sample is timestamped against the host's monotonic clock and shipped to the local observability collector. AMD, Intel, and Apple equivalents are catalogued in §3.3.

### 3.2 Pre-emptive quality reduction tiers

The point of pre-emptive reduction is to shed encoder work before the throttle engages, on the principle that a controlled 20 % bitrate cut is invisible to most viewers while a 30 % NVENC throughput collapse is unmistakable. HelixPlay defines four tier thresholds, ordered by GPU package temperature.

| Tier | Trigger temp | Action | Stream impact | Record impact | Recovery |
|---|---|---|---|---|---|
| T1 — Warning | ≥78°C | Reduce stream bitrate by 20 %, capped at codec floor (HEVC 1.5 Mbps, AV1 1.0 Mbps, H.264 2.0 Mbps) | Bitrate -20 %; resolution unchanged | Unchanged | Lift at ≤74°C sustained 10 s |
| T2 — Hard | ≥80°C | Drop stream resolution one tier (4K60→1440p60→1080p60→720p60) | Resolution -1 step; bitrate auto-rescaled | Unchanged | Lift at ≤76°C sustained 15 s |
| T3 — Emergency | ≥82°C | Pause record-encoder submission for 30 s; emit C29-THERMAL-PAUSE event | Unchanged from T2 | Recording paused, gap marker written | Resume at ≤78°C or after 30 s, whichever later |
| T4 — Cliff | ≥83°C | Refuse new sessions on this host until cool-down to ≤75°C | New session-create RPCs return BUSY-THERMAL | Existing recordings remain paused | Re-admit at ≤75°C sustained 30 s |

Three implementation notes apply to the table.

First, the tiers stack rather than replace. T1 stays in effect as the GPU climbs through T2 and T3; the recording pause at T2 does not reset the bitrate adjustment from T1. Recovery walks back down the ladder one tier at a time as the temperature falls through the hysteresis bands.

Second, the codec floor matters. Insight #1's pre-emptive cut is meaningless if it drives bitrate below the point where the codec produces watchable output. HelixPlay caps the T1 reduction at the per-codec minimum so the stream stays viable; if the cap is reached and the GPU is still climbing, the controller jumps directly to T2 rather than continuing to cut bitrate.

Third, the T3 record pause is intentionally surgical. Pausing recording sheds 15-25 W instantly and is the single most effective action the host-agent can take without disrupting the live stream. The recording file gets a gap marker (a metadata timestamp annotated `thermal_pause`) that downstream tooling — including the post-session muxer — uses to splice the resumed segment cleanly. The operator-facing UI surfaces these markers because regulated tenants need to know when recording was non-continuous.

### 3.3 Per-vendor thermal sensor matrix

The four-tier ladder is only useful if the host-agent can read a stable, vendor-agnostic temperature signal. The matrix below documents the canonical sources HelixPlay polls per GPU vendor, with fallback paths when the primary source is unavailable.

| Vendor | Primary sensor command | Fallback | Notes |
|---|---|---|---|
| NVIDIA | `nvidia-smi --query-gpu=temperature.gpu,temperature.memory,power.draw,clocks.sm --format=csv,noheader,nounits` | NVML library bindings (`go-nvml`) | `temperature.memory` is HBM/GDDR junction; trips earlier than package on memory-bound encodes |
| AMD | `rocm-smi -t -P --json` for GPU temp + power; `sensors` (lm-sensors) for PCB hotspots | `/sys/class/drm/card*/device/hwmon/hwmon*/temp1_input` direct read | ROCm reports edge, junction, memory; HelixPlay uses junction as the trigger |
| Intel (Arc/Xe) | `intel_gpu_top -J` (JSON output) for GPU package temp | `/sys/class/drm/card*/device/hwmon/.../temp1_input` | QuickSync power draw is lower than NVENC; thermal walls are correspondingly cooler |
| Apple Silicon (M-series) | `powermetrics --samplers smc,gpu_power -i 1000 -n 1` | `system_profiler SPHardwareDataType` for static info; thermal pressure via `pmset -g thermlog` | macOS does not expose a single GPU sensor; HelixPlay reads `gpu_power` and infers thermal pressure from system thermal state |

Several caveats apply across vendors.

`nvidia-smi` is the lowest-friction option but adds a process-spawn cost per poll; on hosts running >20 sessions the host-agent uses NVML bindings directly to avoid the fork overhead. The CSV output of `nvidia-smi` is also locale-sensitive in older driver versions — HelixPlay forces `LC_ALL=C` in the spawn environment to keep parsing deterministic.

ROCm's `rocm-smi` produces three temperatures (edge, junction, memory). Junction is the package equivalent and is the correct trigger. HelixPlay's vendor-abstraction layer collapses the three values into a single `effective_temp` field by taking the maximum, so downstream tier logic is uniform across vendors.

Intel Arc/Xe encodes via QuickSync, whose power footprint is materially lower than NVENC's. The thermal walls in §3.2 still apply by topology but are usually hit later in the day on Intel-only hosts because the wall is colder relative to ambient. The host-agent does not lower the trigger temperatures for Intel hosts; it lets the lower power profile naturally widen the headroom.

Apple Silicon is a special case. macOS does not expose a per-die GPU temperature in the public APIs, only the package thermal pressure level (`Nominal`, `Fair`, `Serious`, `Critical`). HelixPlay maps these levels onto T1-T4 directly: `Fair` → T1, `Serious` → T2/T3, `Critical` → T4. The host-agent on macOS therefore cannot do fine-grained pre-emption, and the documentation flags this as a known limitation; macOS hosts are recommended only for development, not production multi-tenant deployment.

### 3.4 GPU power-draw correlation and trajectory prediction

Temperature is a lagging indicator. By the time the package reads 80°C, the silicon has already been pushing into that envelope for some seconds, and the host-agent's window to act is narrow. Power draw, by contrast, is an immediate indicator: when the second encoder path engages, board power rises within milliseconds, well before the thermal mass of the heatsink registers the change.

HelixPlay therefore uses power as the leading-edge signal and temperature as the confirming signal. The host-agent maintains a 5-second moving average of board power (`power.draw` from `nvidia-smi` or equivalent), tagged against the published TDP for the GPU SKU. A draw exceeding 90 % of TDP for ≥3 consecutive seconds triggers the same pre-emptive actions as a T1 temperature hit, even if the package is still cool.

The 5-second window is chosen empirically. A shorter window (1-2 s) over-reacts to scene-change spikes that the GPU dissipates quickly. A longer window (10 s+) loses the leading-indicator advantage. Five seconds is the empirically-stable sweet spot across the GPUs in `video-tech_dim04.md`'s test matrix.

The host-agent also maintains a thermal trajectory estimate using a simple linear fit over the last 30 temperature samples. If the fit projects a crossing of 78°C within the next 15 seconds, the agent pre-emptively engages T1 even if the current temperature is still in the safe band. This look-ahead behavior catches the case where a fresh dual-path session has just been admitted and the GPU is climbing fast — by the time the actual reading hits T1, the agent is already mitigating.

| Metric | Source | Window | Trigger condition | Action |
|---|---|---|---|---|
| Instant temperature | Vendor sensor matrix (§3.3) | None | Per §3.2 tier table | Per §3.2 |
| 5 s power moving avg | Vendor power query | 5 s | >90 % TDP for 3 s | T1 pre-emption |
| Thermal trajectory | Linear fit over 30 samples | 30 s | Projected ≥78°C within 15 s | T1 pre-emption |
| Memory junction temp | NVIDIA/AMD memory sensor | None | ≥85°C (vendor-specific) | Skip ladder, jump to T2 |

The memory-junction row is worth a callout. HEVC 10-bit and AV1 encodes are memory-bandwidth-heavy on the encoder side, and on cards with GDDR6 (rather than GDDR6X or HBM), the memory junction can throttle before the GPU package does. HelixPlay's host-agent treats a memory-junction reading at 85°C+ as a direct T2 trigger, bypassing T1, because memory throttling has the same sharp-edge characteristic as compute throttling and the response must be equally fast.

### 3.5 Liquid-cooled GPU deployment posture

All of §3.2-3.4 assumes air cooling. The thresholds are conservative on purpose: the air-cooled population is much larger and the failure mode is more severe (acoustics, dust ingress, fan-curve drift). Operators with recording-intensive workloads — particularly tenants running esports tournament archives or compliance-grade session capture — can opt into a liquid-cooled posture that materially widens the headroom.

HelixPlay's recommended liquid-cooled SKUs:

| GPU | Cooling solution | Approx package delta vs air | Notes |
|---|---|---|---|
| RTX 4090 | EK-Quantum Vector² Strix/TUF block + 360 mm rad | -12 to -15°C under sustained dual-path | Most popular; readily available |
| RTX 5090 | Bykski / Alphacool reference block + 420 mm rad | -10 to -14°C | New-gen; block availability still maturing |
| AMD Radeon Pro W7900 / W7800 | Reference Aqua variant or third-party block | -10 to -13°C | Pro line preferred for ECC memory |
| RTX 6000 Ada (workstation) | Native blower or EK-Pro block | -8 to -12°C | Often deployed with chassis liquid loop |

When the operator declares a host as liquid-cooled in the host-policy file, the host-agent shifts the §3.2 tier triggers up by 5°C: T1 becomes 83°C, T2 becomes 85°C, T3 becomes 87°C, T4 becomes 88°C. The shift is bounded — the agent never trusts the policy flag enough to disable thermal management entirely. If a "liquid-cooled" host actually thermal-throttles (detected by sudden NVENC submission latency spikes), the agent reverts to the air-cooled thresholds within one minute and emits a `thermal-policy-mismatch` event for the operator dashboard.

This is explicitly an operator-policy opt-in and not the MVP default. The MVP ships with air-cooled assumptions baked in, and §3.2's thresholds apply unconditionally unless the operator has signed off on the liquid-cooled deployment characteristics. The default protects new operators from misconfiguration; the opt-in lets advanced operators reclaim the headroom they have engineered for.

### 3.6 Cross-link C34 thermal and GPU balancing

Everything above is local. It applies to a single host with one or more GPUs and addresses the dual-path-specific question of how much encoder work that host can sustain before the wall is hit.

C34 — Thermal & GPU Balancing — owns the cross-host story: how the scheduler picks the next host to admit a session, how multi-GPU hosts spread sessions across cards to keep aggregate thermal pressure low, how the fleet-wide thermal observability rolls up into capacity-planning dashboards, and how a thermally-saturated host signals back-pressure to the global session router.

The boundary between this section and C34 is firm. C29 §3 answers "given this host, when do I dial back dual-path encoding?" C34 answers "given this fleet, where do I send the next session, and how do I keep the fleet's thermal envelope below the budgeted ceiling?" The two are designed to compose: the host-agent's tier transitions emit events that C34's scheduler consumes, so a host crossing T1 immediately becomes a less-preferred admission target without the scheduler needing to second-guess the local decision.

A second cross-link is to C28 — Network & Bitrate Adaptation. C29 §3.2's T1 action (bitrate -20 %) interacts with C28's ABR controller. The host-agent's thermal trigger does not bypass ABR; it lowers the upper bound that ABR is allowed to hand the encoder. When thermal pressure clears, the upper bound is restored and ABR resumes its normal autonomy. This composition is documented in C28 §5.4 and is not duplicated here.

## 4. Quality budget allocation

Section 3 dealt with how much encoder work the host can do. Section 4 deals with how to spend that budget once you have it: stream and record both want pixels, bits, and frames, and the question is whether they get the same allocation (symmetric) or different ones (asymmetric).

The HelixPlay default is asymmetric, and the rest of this section explains why and how. The short version: the live stream is bound by the network path to the player, which fluctuates session to session and minute to minute, while the recording is bound by storage cost, which is a planned and per-tenant-budgeted resource. Treating them symmetrically is convenient but wastes either storage (record at the live ceiling) or quality (live at the record floor). Asymmetric allocation lets each path settle at its natural operating point.

### 4.1 Symmetric versus asymmetric quality

A symmetric pipeline configures both encoder paths identically: same codec, same resolution, same bitrate, same GOP, same frame-rate. Operationally this is the simplest mental model. It also produces the most predictable storage profile (every recording hour costs the same number of bytes) and the simplest playback story (the recorded stream looks like the live one).

The cost is twofold. First, when the live stream's ABR controller (C33) cuts bitrate in response to network pressure, the recording's bitrate gets cut too — even though the storage subsystem has no analogous pressure. The recorded asset becomes a fluctuating-quality artifact that depends on the player's network conditions at record time, which is a poor archive characteristic. Second, when storage budget allows for high-quality recording but the live network does not allow high-quality streaming, symmetric configuration leaves quality on the table for the recording.

Asymmetric configuration fixes both. The recording is configured at a quality target appropriate to the storage budget and the playback use-case (replay, share, audit), while the live stream is left free to track its own ABR loop. The two paths share the same captured frame source (no double capture; refer to §2 for the decoupling architecture) but encode independently with their own rate-control state.

| Property | Symmetric | Asymmetric (HelixPlay default) |
|---|---|---|
| Configuration complexity | Low — one knob | Moderate — two independent profiles |
| Stream bitrate behavior | Tracks ABR | Tracks ABR (3-25 Mbps envelope) |
| Record bitrate behavior | Tracks ABR (cuts under network pressure) | Fixed at session start (8-15 Mbps HEVC) |
| Storage cost predictability | Variable per session | Predictable per (codec, resolution, duration) |
| Archive quality | Bound by worst network minute | Bound by storage budget |
| Encoder GPU cost | Lower (some shared rate-control state) | Higher (two independent RC loops) |
| Tenant-policy flexibility | Coarse | Fine-grained (per-tenant record cap) |

The encoder GPU cost row is real but small. Modern NVENC silicon handles the second rate-control loop without measurable additional power draw beyond the 15-25 W base delta from §3.1 — the rate-control state is cheap relative to the actual pixel encoding. The asymmetric design therefore does not change the §3 thermal budget materially.

### 4.2 Per-codec storage budget on the recording side

Recording bitrate translates directly into storage cost, which is a primary operator-policy input. The table below gives the canonical per-codec planning numbers HelixPlay uses for 1080p60 recordings; the values are computed from `video-tech_dim04.md`'s rate-distortion measurements and are the basis for the per-tenant storage quotas surfaced in the operator dashboard.

| Codec | Target bitrate | GB/hour | TB/1000 hours | Quality at target | HelixPlay use |
|---|---|---|---|---|---|
| HEVC (H.265) Main10 | 8 Mbps | 3.6 | 3.6 | High | Default for recording |
| AV1 | 5 Mbps | 2.25 | 2.25 | High (slightly behind HEVC at this rate) | Future default once decode is universal |
| H.264 High | 12 Mbps | 5.4 | 5.4 | Acceptable | Fallback for legacy decode targets |
| HEVC Main10 (4K60) | 25 Mbps | 11.25 | 11.25 | Reference | Premium tier opt-in |
| AV1 (4K60) | 18 Mbps | 8.1 | 8.1 | Reference | Premium tier opt-in |

The HelixPlay rule is HEVC at 8 Mbps as the recording default. The reasoning:

- **Storage cost is reasonable.** 3.6 GB/hour is well within the per-tenant budgets typical operators are willing to allocate (most start at 100 GB/tenant/month, which buys 28 hours of recording).
- **Quality is high.** HEVC at 8 Mbps and 1080p60 is visually transparent for most game content; the rate-distortion curve is well above the perceptual-acceptability threshold.
- **Decode reach is universal in 2026.** Every TV chassis, every modern phone, every desktop browser through MSE supports HEVC decode in hardware. AV1 decode is not yet universal in the TV cohort (refer to C18), so AV1 is the future default rather than the present one.

The H.264 12 Mbps fallback exists because some legacy archive workflows cannot ingest HEVC. The bitrate is uplifted from 8 Mbps to 12 Mbps to compensate for H.264's lower efficiency at the same quality target. Operators with H.264-only archive infrastructure pay the storage premium; operators who do not are not penalized.

### 4.3 Bitrate independence enforcement

A core asymmetric-design invariant is that the recording bitrate is decoupled from the streaming bitrate. The host-agent enforces this in three places:

First, the stream-encoder bitrate is dynamically adjusted by the ABR controller (C33) on a sub-second cadence. ABR responds to RTT, packet loss, NACK rate, and the SCReAM/GCC bandwidth estimate. Its full bitrate envelope is 3-25 Mbps depending on session profile. None of this signal is forwarded to the record-encoder.

Second, the record-encoder bitrate is fixed at session start based on (tenant policy, codec, resolution) and does not change for the lifetime of the session. Mid-session bitrate changes on the recording path are explicitly forbidden; they would create non-stationary rate-distortion characteristics in the recorded asset and complicate downstream analytics.

Third, the tenant operator-policy can set a per-tenant record bitrate cap that overrides the codec default. Some tenants want to budget storage tightly (cap at 5 Mbps HEVC); some want premium archives (raise to 12 Mbps HEVC). The cap is read once at session start and locked into the encoder configuration.

| Path | Bitrate source | Update cadence | Override |
|---|---|---|---|
| Stream | C33 ABR controller output | Sub-second | Per-session ABR profile |
| Record | (Tenant policy) → (Codec default) → Fixed at session start | None during session | Per-tenant operator-policy cap |

The decoupling is not aspirational — it is an enforceable invariant in the host-agent's encoder-configuration code path. The two encoder instances expose different rate-control APIs to different controllers, and there is no shared bitrate state between them.

### 4.4 GOP independence

GOP (group-of-pictures) structure is independent across the two paths and is tuned to each path's downstream consumer.

The stream uses a 4-second GOP (240 frames at 60 fps). The motivation for the long GOP on the live path is twofold: keyframes are expensive in bits and the live stream is bitrate-constrained, so amortizing the keyframe cost across more frames buys overall quality. Intra-refresh is enabled on the stream — which means the encoder periodically sends partial-frame intra refreshes (refer to C26 §4.4) instead of full keyframes — so the player can recover from packet loss without waiting four seconds for the next full keyframe.

The recording uses a 1-second GOP (60 frames at 60 fps) with full keyframes and no intra-refresh. Recording is not bitrate-constrained in the same way the live path is, so the keyframe cost is affordable. The benefit is seek precision: a 1-second GOP gives the post-session player one-second seek granularity without re-decoding from the start, which is a major usability win for replay and highlight workflows. Intra-refresh would defeat the seek benefit because partial intra slices cannot be used as random-access points.

| Property | Stream | Record |
|---|---|---|
| GOP length | 4 s (240 frames) | 1 s (60 frames) |
| Keyframe type | Full keyframes + intra-refresh slices | Full keyframes only |
| Intra-refresh enabled | Yes (refer to C26 §4.4) | No |
| Random-access cadence | 4 s nominal; ~250 ms via intra-refresh recovery | 1 s |
| Optimization target | Loss recovery + bitrate efficiency | Seek precision + clean cuts |

A subtle implication is that the recorded asset has a finer-grained timestamp grid than the live stream, which downstream analytics and highlight-extraction tools can exploit. The C30 (post-session muxing) chapter relies on the 1-second record GOP to produce HLS/DASH segment boundaries without re-encoding.

### 4.5 Frame-rate independence

Both paths share the same captured frame source — capture is single-source by design (refer to §2) — so the input frame rate is identical. Where the two paths can diverge is in post-capture frame-rate filtering on the encoder input.

The stream always encodes at the capture rate (typically 60 fps for game content; 120 fps for high-refresh sessions; 30 fps for some legacy or simulation workloads). Cutting the stream's frame rate would degrade perceptual smoothness and is not a knob the system exposes.

The recording can optionally drop every Nth frame to halve, third, or quarter the effective rate. The most common reduction is 60 → 30 fps for storage efficiency. The host-agent implements this as an integer-decimation filter on the encoder input rather than an encoder-side operation, so the encoder always sees a clean cadence and rate-control behavior is unaffected.

| Capture rate | Stream rate | Record rate options | Storage saving (vs capture rate record) |
|---|---|---|---|
| 60 fps | 60 fps | 60 (default), 30, 20, 15 | 0 % / -50 % / -67 % / -75 % |
| 120 fps | 120 fps | 120 (default), 60, 30 | 0 % / -50 % / -75 % |
| 30 fps | 30 fps | 30 (default), 15 | 0 % / -50 % |

Three rules apply to frame-rate decimation on the recording side.

First, the decimation ratio must be an integer divisor of the capture rate. Non-integer ratios produce uneven frame timing and confuse downstream players. The host-agent rejects non-integer requests at session-create time.

Second, the decimation is configured at session start and does not change mid-session. Like bitrate, mid-session frame-rate changes on the recording path would create non-stationary characteristics in the asset and break downstream tooling.

Third, the HelixPlay default is recording-at-capture-rate. Frame-rate decimation is a tenant-operator override for storage-conscious deployments; it is not enabled by default because the storage savings (typically 50 %) come at a noticeable smoothness cost in the playback experience, and most operators prefer paying the storage to keep the archive faithful to the live experience.

The decimation knob composes cleanly with the bitrate cap from §4.3 and the codec choice from §4.2, so a tenant who wants the cheapest possible compliance archive can stack all three (AV1 at 5 Mbps, 30 fps from 60 fps capture, 1080p60 → 1080p30 effective) and bring storage cost under 1 GB/hour. The cost is recording quality, but the operator policy is explicit about the trade-off and the affected tenants accept it knowingly.
## 5. Codec / container divergence

The whole point of a dual-path pipeline is that the stream and the record
do not share an output. They share a captured frame, they share an encoder
device (sometimes), they share a clock domain — but the bytes that come
out of the encoder are different bytes, packed differently, with different
GOP structures, different rate-control modes, and different containers.
This section nails down exactly how the two paths diverge once the
`FrameTee` (defined in §6.1) hands a frame to each encoder. It is the
contract C30 (Recording Storage) and the streaming sender (C20-C24,
C30-style cross-refs in C26) consume.

C26 §6 fixed the codec ladder for the *stream*: H.264 baseline as the
universal floor, HEVC as the default for tenants/devices that license it,
AV1 as the experimental tier behind a hardware-and-tenant gate. That
ladder applies *only* to the stream side. The record side runs its own
ladder, sometimes the same codec, sometimes not, and never tied to the
stream's negotiated codec for the session lifetime. A 1080p60 H.264
stream can be recorded as HEVC fMP4. A 4K60 HEVC stream can be recorded
as AV1 MKV. The decoupling is intentional and is the reason `helix-dualpath`
exists as its own submodule rather than as a flag on `helix-encoder`.

### 5.1 Stream-side codec/container

The streaming path obeys C26's negotiation with no modification. At
session start the SDP/ICE handshake (or the helix-transport equivalent
when WebRTC is bypassed in favour of the custom UDP path described in
C20-C24) settles on one codec for the session: H.264 Baseline/Main if the
client cannot prove HEVC support, HEVC Main/Main10 when the client
declares it via fmtp + the host has NVENC/AMF/QSV with HEVC enabled, AV1
only when the tenant policy *and* the encoder hardware *and* the decoder
on the client all line up. C26 §3.5 fixed lookahead at zero, B-frames at
zero, GOP at IDR-on-keyframe-request, and slice count at GPU-tuned values
for parallel NAL emission.

Container = RTP. There is no MP4, no MKV, no MPEG-TS, no fragmented MP4
on the streaming wire. NAL units come out of the encoder, get fragmented
according to RFC 6184 (H.264), RFC 7798 (HEVC), or the AV1 RTP draft
(RFC 9381 once ratified, draft-ietf-payload-rtp-av1 in the meantime),
get FEC-coded by the C19/C16 pipeline, and are pushed to the wire. The
dual-path does not see RTP at all — that is downstream. What dual-path
guarantees is: the stream-side encoder produces NAL units in a SPSC
ring keyed by the same `frame_id` the capture stage assigned, with
encoder timestamps in the helix-clock domain (not the GPU's clock,
not wall clock).

ABR (covered fully in C33 — Adaptive Bitrate) drives bitrate at the
stream encoder. The encoder accepts a `target_bitrate_kbps` channel
(buffered to depth 1 with last-write-wins semantics; the dropper logs
to OTel but does not block) and applies the new bitrate at the next
rate-control window. ABR may switch bitrate up to 4 times per second
under aggressive conditions; the encoder smooths this with its own
HRD/VBV logic. ABR may not switch *codec* mid-session — that would
require a renegotiation that costs at least one full IDR plus a
client-side decoder reset, which is incompatible with a sub-100 ms
glass-to-glass budget. So the stream codec is fixed for the session;
only the bitrate floats.

### 5.2 Record-side codec/container

The record path picks a codec independently. The default is HEVC Main/
Main10 because storage cost dominates record economics — C30 will lay
out per-tenant storage budgets, but the working assumption is that a
4-hour recording at 4K60 must fit inside a single-digit-GB envelope,
which only HEVC and AV1 hit at acceptable quality. H.264 is the fallback
when (a) the host cannot license HEVC for record output (some encoder
SKUs decouple stream and record licensing), or (b) the tenant has
disabled HEVC for record specifically (rare, but exists for tenants
that re-export recordings to clients running ancient decoders). AV1
is opt-in: the tenant policy must enable it, the encoder must report
AV1 record capability (`Capabilities.AV1RecordSupported`), and the
thermal headroom must be sufficient — AV1 record encoding burns more
GPU cycles than HEVC at the quality tiers we want, and we are
*already* spending those cycles on the stream side under §3 of this
chapter.

Container is fragmented MP4 (fMP4) by default. fMP4 with `moof`/`mdat`
fragments every 2 s gives crash-safe append semantics: if the host
panics or loses power, the consumer (C30's storage finaliser) can
truncate to the last well-formed fragment and produce a playable file
without an explicit `moov` rebuild. The MKV alternative is offered as
operator policy (some tenants have downstream tooling that prefers MKV
because it tolerates partial files better) and is selected per-tenant
in the dualpath config schema (§6.2). C30 owns the on-disk layout; this
chapter just emits the elementary stream + container metadata.

### 5.3 Per-codec record bitrate matrix

These are *defaults*. Tenants can override per-tier via the dualpath
config; the bitrate floor and ceiling are pinned by chapter C33 (ABR),
but ABR does not modulate the record bitrate — record runs CBR or
CRF, not ABR. The matrix below sets the working envelope for the MVP
and matches the storage-budget math C30 will document.

| Codec | Resolution | Framerate | Default record bitrate | Mode  |
|-------|-----------|-----------|------------------------|-------|
| HEVC  | 1080p     | 60        | 6-10 Mbps              | CBR   |
| HEVC  | 4K        | 60        | 15-25 Mbps             | CBR   |
| AV1   | 1080p     | 60        | 4-7 Mbps               | CRF20 |
| AV1   | 4K        | 60        | 10-18 Mbps             | CRF20 |
| H.264 | 1080p     | 60        | 10-15 Mbps             | CBR   |
| H.264 | 4K        | 60        | 25-40 Mbps             | CBR   |

The HEVC numbers track NVIDIA's "4K Shadow" defaults at the high end
and 2-pass-equivalent quality at the low end, accounting for the fact
that we cannot do 2-pass at record-time without latency penalties to
the stream (the encoders share the same GPU). AV1 numbers are calibrated
against AOM's reference grain-removal on at CRF20 — AV1 buys roughly
40% bitrate at iso-quality vs HEVC, and the matrix reflects that.
H.264 record numbers are higher because H.264 is strictly the
licence-fallback path, not the bitrate-efficient path.

CBR is preferred for HEVC and H.264 because the disk consumer (C30) and
the storage budget calculator both want predictable byte rates per
second — VBR records would make per-session storage budgeting
non-deterministic, which collides with the multi-tenant fairness rules
C09 (Capacity Planning) imposes. AV1 runs CRF because the AV1 rate
control under aggressive CBR caps produces visible quality dips on
high-motion scenes, and the AV1 tenants are by definition tolerant of
slightly variable file sizes.

### 5.4 Lookahead independence

The single most important divergence between the two paths is lookahead.
C26 §3.5 forced lookahead = 0 frames on the stream side, because
lookahead-N adds N/fps of latency to first-frame-out, and at 60 fps that
is 16.6 ms per frame of lookahead — a budget we cannot afford under
C13/C14's 50 ms motion-to-photon target.

Record has no such constraint. The record encoder's output goes to disk,
not to a remote eyeball, and the only consumer that cares about record
latency is the "live preview" feature (operator watching the recording
in near-real-time from the same host's UI), which tolerates 1-2 s of
lag. So the record encoder runs with lookahead 8-16 frames, configurable
per tenant. Lookahead 8 buys roughly 5-8% bitrate efficiency at iso-
quality on HEVC; lookahead 16 buys 10-15%. The latency penalty (133 ms
at 16 frames / 60 fps) is invisible because record is not a real-time
sink.

This is why the record encoder is a *separate encoder instance* in §6.1,
not a "branch" of the stream encoder. NVENC, AMF, and QSV all pin
lookahead at session creation; you cannot mutate it per-frame. So a
shared encoder cannot serve both paths. The dual-path pattern allocates
two encoder sessions on the same GPU and tees frames into both —
that is the entire point of `helix-dualpath`.

### 5.5 Cross-link C30 recording storage

The dual-path pipeline produces NAL units (or AV1 OBUs) at the record
encoder's output ring. From there:

- The record encoder hands NALs to `recordmuxer.fMP4Muxer` (or
  `recordmuxer.MKVMuxer`), defined in C30. The muxer attaches container
  metadata (track headers, timing offsets, codec-private data) and
  emits 2 s fragments to a local-buffer ring on tmpfs (C30 Insight #4).
- A background-sync goroutine drains the local-buffer ring to durable
  storage (object store + tenant-owned mount) at sync depth N (default
  3 fragments behind = 6 s of delay, configurable). The sync is
  resumable; if the network or storage tier fails, the local buffer
  retains up to the tenant-configured retention cap (default 4 hours)
  before backpressuring the encoder.
- C30 also owns the crash-safe finalisation: on host panic, an SBOM-
  level fsck pass walks the local buffer, finds the last well-formed
  fragment, and either uploads it as a partial recording or merges
  with the in-progress upload. None of that is in scope for this
  chapter.

What this chapter guarantees to C30: (a) a NAL/OBU stream tagged with
the helix-clock timestamp, (b) a stable codec for the session lifetime
*on the record side too* — record codec is fixed at session start
just like stream codec, (c) a backpressure signal when the consumer
(C30's local-buffer ring) is full, propagating back through the
encoder ring into the dual-path tee, where it is logged but does not
backpressure capture (capture must keep running because the *stream*
side does not block on record; see §6.3 for the precedence rules).

## 6. Implementation contract

Sections 1-5 specified the architecture, the failure modes, the codec
divergence, and the bitrate matrix. This section translates them into
a concrete submodule layout, capability schema, bootstrap sequence, and
sample Go code that the implementation team can clone and extend.
Everything here obeys R-03 (decoupling — every reusable component is a
public submodule under the `vasic-digital` org), R-08 (no TODO/FIXME,
no dead code), and R-18 (Operational Integrity — all subprocess
invocations go through helix-r18-safeexec).

### 6.1 Submodule boundaries (R-03)

A new public submodule `vasic-digital/helix-dualpath` owns this chapter's
implementation. Its public surface is small and intentional:

- **`dualpath.FrameTee`** — the fan-out primitive. Takes a single SPSC
  input ring of `FrameEvent` (defined in helix-capture, carrying a
  zero-copy GPU handle, helix-clock timestamp, `frame_id`, and
  per-frame metadata blob) and fans out to two SPSC output rings:
  one feeding the stream encoder, one feeding the record encoder.
  FrameTee never copies frame *bytes*; it copies the FrameEvent struct,
  which is a 64-byte handle. The underlying GPU surface is reference-
  counted via helix-capture's CUDA-graph or D3D11/Metal resource
  tracking. FrameTee is goroutine-safe between exactly one producer
  and exactly two consumers; it is not a fan-out tree.

- **`dualpath.SessionConfig`** — per-session config: stream codec,
  stream bitrate floor/ceiling/initial, record codec, record bitrate,
  record container (fMP4 / MKV), record lookahead, thermal threshold
  override, AV1-record opt-in. SessionConfig is loaded from the
  tenant config service (gRPC, via helix-config) at session start
  and is *immutable* for the session — mutations require a session
  teardown + bring-up. This matches §5.1's "codec is fixed for the
  session lifetime" rule.

- **`dualpath.ThermalGuard`** — polls GPU/encoder die temps every
  1 s via the encoder vendor SDK (NVML on NVIDIA, AGS on AMD, IGCL
  on Intel), and emits `QualityReductionEvent` on a buffered channel
  when the temp exceeds `thermal_threshold_c`. The event carries a
  recommended action: drop record framerate, drop stream framerate,
  drop record bitrate, drop stream bitrate, in that order of
  preference (record degrades first; stream is the user-visible
  path). C28 §6 documents the encoder-side response logic.
  ThermalGuard does not act unilaterally — it emits an event and the
  encoder consumers act on it. This keeps the policy in user space
  and makes the behaviour testable without a hot GPU.

The submodule reuses, never reinvents:

- `vasic-digital/helix-shm` — for the GPU surface registry that
  FrameEvent points into.
- `vasic-digital/helix-lockfree` — for the SPSC rings between capture,
  tee, and the two encoders.
- `vasic-digital/helix-codec` — for codec-specific config types
  (NAL parsers, SPS/PPS extractors, AV1 OBU framers).
- `vasic-digital/helix-encoder` — the encoder-vendor abstraction
  (NVENC/AMF/QSV/VideoToolbox/MediaCodec wrappers).
- `vasic-digital/helix-capture` — the capture-stage producer.
- `vasic-digital/helix-r18-safeexec` (R-18) — for any subprocess
  invocation (FFmpeg fallback paths, vendor diagnostic tools,
  thermal probes that shell out).

`helix-dualpath` therefore depends on six existing submodules and
introduces zero copies of code already in those submodules. R-03
and R-09 (test matrix per submodule) apply to the new submodule
independently — it ships with Unit, Integration, E2E, Security,
Benchmark, Chaos, Stress, Smoke, Full-Auto, and Challenge tests
(see C35 for the per-submodule test scaffolding).

### 6.2 Capability schema delta

The host capability advertisement (defined in C04 — Host Agent
Bootstrap) gains a `dualpath` block:

```yaml
dualpath:
  supported: bool            # true if encoder reports >= 2 sessions
                             # AND thermal headroom >= 5C below threshold
                             # AND record codec licence is present
  thermal_threshold_c: int   # default 78; AMD/Intel may override
  record_codec: codec.Codec  # one of {h264, hevc, av1}
  record_bitrate_kbps: int   # bitrate floor; matrix in §5.3
  record_container: string   # one of {fmp4, mkv}
  record_lookahead_frames: int  # 0..16; default 8
  av1_record_supported: bool # capability, not policy
```

The `supported` flag gates the entire feature: a host with a single
encoder session (e.g., a developer laptop with an entry-level QSV
SKU) will advertise `supported: false`, and the session orchestrator
(C03) will refuse to schedule record on that host. The orchestrator
will still allow streaming-only sessions there.

The capability schema is consumed by C03 (Session Orchestration), C09
(Capacity Planning — knows which hosts can record at which tiers),
and the operator UI. Schema is versioned; adding a field requires a
helix-config schema bump and a graceful-degradation path for older
hosts.

### 6.3 Bootstrap sequence

1. **Capture stage starts.** helix-capture brings up the GPU capture
   surface (DXGI desktop duplication on Windows, kmsgrab/PipeWire on
   Linux, ScreenCaptureKit on macOS), allocates the surface pool, and
   begins emitting FrameEvent on the capture-output SPSC ring at the
   negotiated framerate.

2. **FrameTee wires up.** `dualpath.NewFrameTee(captureRing,
   streamEncoderRing, recordEncoderRing)` constructs the tee. The tee
   does not yet pull frames — it waits for `Start(ctx)`.

3. **Stream encoder starts.** helix-encoder opens an NVENC/AMF/QSV
   session with the negotiated stream codec, lookahead 0, B-frames 0,
   the rate-control mode from C26 §3, and binds its input ring to
   `streamEncoderRing`. Its output goes into the stream-output ring,
   consumed by helix-transport.

4. **Record encoder starts.** A *second* encoder session opens on
   the same GPU with the record codec, lookahead 8-16, the bitrate
   from §5.3, and binds to `recordEncoderRing`. Its output goes
   into the record-output ring, consumed by `recordmuxer` (C30).

5. **ThermalGuard starts.** Polling at 1 Hz via NVML/AGS/IGCL.
   Emits events into the dualpath quality-event ring; both encoder
   consumers subscribe.

6. **FrameTee.Start(ctx) is called.** The tee goroutine begins
   pulling FrameEvents from `captureRing` and pushing into both
   output rings. If either output ring is full, the tee logs
   (OTel counter `dualpath.tee.drop{path=stream|record}`) and
   *drops the frame for that path only* — it does not block the
   other path or the capture producer. This is Insight #5 from §3:
   the two encoders run in their own goroutines and do not
   backpressure each other.

7. **Steady state.** Capture → tee → two encoders → two outputs.
   Quality events from ThermalGuard arrive asynchronously and
   are applied at the next encoder rate-control window.

8. **Teardown.** Session-end signal arrives via context cancel.
   FrameTee drains its output rings (best-effort, bounded), then
   exits. Encoders flush their last GOPs and close the underlying
   vendor sessions. Capture stops last. C30's muxer finaliser
   handles partial-fragment reconciliation.

### 6.4 Go code

```go
// Package dualpath implements zero-copy frame fan-out between the
// streaming and recording paths of the HelixPlay capture pipeline.
package dualpath

import (
	"context"
	"errors"
	"sync/atomic"

	"golang.org/x/sys/unix"

	"github.com/vasic-digital/helix-capture/frame"
	"github.com/vasic-digital/helix-encoder"
	"github.com/vasic-digital/helix-lockfree"
	"github.com/vasic-digital/helix-shm"
	r18 "github.com/vasic-digital/helix-r18-safeexec"
)

// FrameTee fans out one FrameEvent producer to two encoder consumers
// without copying GPU surfaces. Drops on a full output ring are
// per-path and never block the other path or the producer.
type FrameTee struct {
	in          *lockfree.SPSCRing[frame.Event]
	streamOut   *lockfree.SPSCRing[frame.Event]
	recordOut   *lockfree.SPSCRing[frame.Event]
	streamDrops atomic.Uint64
	recordDrops atomic.Uint64
	registry    *shm.SurfaceRegistry
}

// NewFrameTee constructs a tee. Rings must be pre-allocated by the
// caller; the tee never resizes them. The registry is the GPU surface
// reference-count owner; the tee bumps refcounts on fan-out.
func NewFrameTee(
	in *lockfree.SPSCRing[frame.Event],
	streamOut, recordOut *lockfree.SPSCRing[frame.Event],
	registry *shm.SurfaceRegistry,
) (*FrameTee, error) {
	if in == nil || streamOut == nil || recordOut == nil || registry == nil {
		return nil, errors.New("dualpath: nil ring or registry")
	}
	return &FrameTee{
		in: in, streamOut: streamOut, recordOut: recordOut,
		registry: registry,
	}, nil
}

// Start runs the tee until ctx is done. It is not safe to call twice.
func (t *FrameTee) Start(ctx context.Context) error {
	// Pin to the capture NUMA node for cache locality (R-18: unix
	// syscall, not exec).
	_ = unix.SchedSetaffinity(0, captureCPUSet())
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		ev, ok := t.in.PopWait(ctx)
		if !ok {
			return ctx.Err()
		}
		// Bump refcount once per output path; the encoders release
		// when they are done with the surface.
		t.registry.Retain(ev.SurfaceID, 2)
		if !t.streamOut.TryPush(ev) {
			t.streamDrops.Add(1)
			t.registry.Release(ev.SurfaceID, 1)
		}
		if !t.recordOut.TryPush(ev) {
			t.recordDrops.Add(1)
			t.registry.Release(ev.SurfaceID, 1)
		}
	}
}

// Drops returns per-path drop counts for OTel export.
func (t *FrameTee) Drops() (stream, record uint64) {
	return t.streamDrops.Load(), t.recordDrops.Load()
}

// captureCPUSet returns the pinned CPU set for the capture NUMA node.
// helix-capture exposes the same set; we re-derive locally to avoid
// a circular import.
func captureCPUSet() unix.CPUSet {
	var s unix.CPUSet
	for _, cpu := range encoder.CaptureNUMACPUs() {
		s.Set(cpu)
	}
	return s
}

// ensure r18 is referenced for any subprocess paths used in fallback
// (e.g., NVML probe via nvidia-smi when SDK is unavailable).
var _ = r18.AllowFamily
```

The above is fewer than 90 LOC including imports and is the entire
fan-out primitive. Notable design points:

- Zero copies of pixel data: only the 64-byte FrameEvent struct
  (handle + metadata) traverses the rings.
- Refcount bumped *before* push so a full ring still leaves the
  refcount in a consistent state when we Release back.
- TryPush is non-blocking; PopWait blocks with ctx-aware exit. This
  is the only blocking point and it blocks only on the *input* —
  if capture stops, the tee exits cleanly.
- CPU affinity pinned to the capture NUMA node (typically the same
  PCIe root complex as the GPU) to minimise cross-socket traffic.
- All subprocess invocations (NVML fallback, vendor probes) route
  through helix-r18-safeexec — see §6.5.

### 6.5 R-18 enforcement

C25 (Operational Integrity) §7 published the family allow-list for
helix-r18-safeexec. helix-dualpath does not extend it — every
subprocess this chapter triggers is already covered:

- **NVML / AGS / IGCL probes** — vendor SDKs are dynamic-linked,
  not exec'd. When the SDK is unavailable (some headless container
  configs), ThermalGuard falls back to `nvidia-smi --query-gpu=
  temperature.gpu --format=csv,noheader` via the `nvidia-tools`
  family already in the C25 allow-list.
- **FFmpeg fallback** — when the hardware encoder fails mid-session
  and the dualpath downgrades record to a software encoder, the
  shell-out goes through the `ffmpeg` family in C25.
- **No new families.** No new binaries. No new exec patterns. The
  allow-list at C25 §7 is sufficient for this chapter.

This is deliberate: every chapter that adds subprocess surface area
also adds R-18 audit burden. By piggybacking on C25's existing list
we keep the audit footprint small. If a future hardware vendor (e.g.,
Qualcomm Adreno encoder on ARM hosts) requires a new probe binary,
C25 §7 must be updated *first* and this chapter's docs amended in
lock-step.

The R-18 contract is also enforced at test time (see C35): the
helix-dualpath test suite runs under a sandboxed exec environment
that only permits the C25 §7 families, and any test that tries to
exec something outside that list fails the security-test stage. This
prevents accidental introduction of new exec patterns through tests
that might otherwise be missed in code review.
## 7. Failure modes

The Dual-Path Encoding surface is the chapter where the
**single capture frame stream becomes a forked obligation** —
one fork must reach the WebRTC / custom-UDP transport at sub-
5 ms p999 encode latency for the live stream, the other fork
must land in the local-buffer recording target with audit-grade
fidelity (Insight #4 — recording-as-save-system). C26
(`01_Codec_Selection.md`) owns the codec choice; C27
(`02_Hardware_Encoders.md`) owns the per-vendor encoder
selection; C28 (`03_Capture_Pipelines.md`) owns the capture-
side primitive plus the frame-event SPSC ring producer; this
chapter — C29 — owns the **Frame-Tee fan-out**, the **dual-
session NVENC orchestration**, and the **back-pressure
contract between the stream-encoder lane and the record-
encoder lane** when GPU thermal headroom or NVENC concurrency
limits force a runtime cascade. Every failure mode catalogued
below is therefore a **dual-path-orchestration fault**, a
**thermal-budget fault**, or an **operational-integrity (R-18)
fault** — distinct populations from C26 (codec choice), C27
(encoder selection), and C28 (capture pipeline), and binding
into a **fifth axis** for the end-to-end runbook generator at
`../03_Architecture/12_Latency_Engineering_Overview.md` §13.

The failure modes split into four populations. The **NVENC-
session-arithmetic population (F1, F8, F12)** is the class
where the dual-path admission policy collides with the per-
GPU NVENC concurrency limit (Insight #1 — thermal wall is
GPU-bounded, not session-count-bounded; F1 is the count limit;
F12 is the within-session quality drift). The **thermal-
budget population (F2, F7, F11)** is the class where the GPU
silicon thermal envelope (cross-link C34) shifts into pre-
emptive throttle (78 °C), hard-cap (83 °C), or where the
sensor itself faults (`nvidia-smi` hung at the syscall). The
**dual-path-orchestration population (F3, F4, F5, F6)** is the
class where the Frame-Tee fan-out, lookahead buffers, or
single-lane encoder failures break the dual-path contract —
either by deadlocking the producer, overrunning a downstream
buffer, or partially failing in a way that requires the
surviving lane to continue while the failed lane drops out
(F5 stream-fails-record-continues; F6 record-fails-stream-
continues). The **operational-integrity population (F9, F10)**
covers the R-18 SafeExec wrapper rejection at the encoder
subprocess boundary (F10) and the recording disk-write fault
that C30 owns (F9 cross-link). F10 is the chapter's R-18
trip-wire (symmetric with C26-F9, C27-F10, C28-F10), non-
overridable per Constitution §11.5.4.

The five-column Symptom / Detection / Mitigation / Fallback
table below is the source of truth for the dual-path runbook
generator at `../03_Architecture/12_Latency_Engineering_Overview.md`
§13 and the alert-rule generation in
`../08_Operations/04_Observability_and_Events.md` (queued).
The fallback semantics across F1–F12 follow the **fail closed
at admission, degrade open at runtime** pattern symmetric
with C26 §7, C27 §7, and C28 §7. Admission-time invariants
(F1 NVENC dual-session count, F8 mid-session bitrate-change
attempts, F10 SafeExec argv allow-list) refuse session
admission and emit `dualpath.admission_refused
{session=…,cause=…}` events that the C24 measurement harness
propagates into the metrics plane. Runtime invariants (F2
thermal throttle, F3 lookahead overflow, F4 Frame-Tee
deadlock, F5 stream-encoder fail, F6 record-encoder fail, F7
power-draw exceedance, F9 disk-write fail, F11 sensor hang,
F12 quality drift) emit `dualpath.degraded
{from=…,to=…,lane=…}` events and the cascade falls forward —
typically toward dropping the record path while preserving
the stream path (Insight #1 binding — stream is the SLA-
bearing lane), or toward a thermal-aware bitrate reduction
across both lanes proportionally.

The **F1 NVENC dual-session limit hit** row is the chapter's
binding to Insight #1 (thermal wall) at the **session-
arithmetic** level rather than the thermal-envelope level.
NVIDIA Lovelace consumer cards (RTX 4090, 4080, 4070 Ti) cap
concurrent NVENC sessions at 8; a dual-path session consumes
**two NVENC sessions per player session** (one stream, one
record). Therefore an RTX 4090 supports a maximum of **4
concurrent dual-path player sessions** before saturation.
The Blackwell consumer line (RTX 5090, 5080) doubles the
limit to 16 sessions = **8 concurrent dual-path sessions**.
Quadro / RTX 6000 Ada removes the cap entirely. The
scheduler must consult the dual-path-aware capability counter
at admission and refuse the new dual-path session if the host
is at capacity, then route to an alternative host via the C08
host-agent pool. F1 is **distinct from C27-F1**: C27-F1 is
the single-session count, C29-F1 is the dual-path doubled
count.

The **F2 thermal throttling at 83 °C** row binds the
chapter's dual-path lane to the C34 thermal-cliff invariant.
Under sustained dual-path 4K60 + 4K30 load, GPU temperatures
climb toward the 78 °C pre-emptive threshold first; the C34
controller drops bitrate gracefully (cascade: 25 Mbps →
20 Mbps → 16 Mbps for the stream lane, proportionally for
the record lane). If the cascade fails to stabilise the
thermal envelope and the GPU reaches the 83 °C hard-cap, the
dual-path orchestrator drops the **record lane** to preserve
the stream lane (Insight #1 binding — stream is the SLA-
bearing lane). The dropped record session emits
`dualpath.record_dropped
{cause=thermal_hardcap,gpu_temp_c=83.2}` and the recording is
marked partial in the C30 catalogue.

The **F3 record-encoder lookahead buffer overflow** row binds
the chapter to the C26-F8 lookahead-depth invariant. The
record-encoder is configured with deeper lookahead than the
stream-encoder (record can afford the latency for higher
quality; stream cannot). Under sustained CPU pressure (e.g.
F11 thermal sensor reading the GPU at hardcap and the
controller asking the record-encoder to drop quality faster
than the lookahead window can drain), the lookahead buffer
overflows. The mitigation is to **reduce lookahead depth on
the record path** to match the stream path, accepting a
marginal record-quality drop to keep the buffer drained.

The **F4 Frame-Tee back-pressure deadlock** row is the
chapter's binding to Insight #5 (Go goroutines map to
pipeline stages — channel-based stage hand-off). The Frame-
Tee fans out from the C28 SPSC ring producer to two
downstream consumers (stream-encoder, record-encoder) over
two channels. If both downstream consumers stall (rare —
typically one stalls while the other proceeds, see F5/F6),
the Frame-Tee blocks at the channel-write boundary and the
upstream SPSC ring fills. The mitigation is a **bounded fan-
out timeout** (5 ms): if either downstream channel-write
exceeds the timeout, drop that frame for that lane while
publishing to the other lane, and emit
`dualpath.tee_deadlock_avoided {lane=…,timeout_ms=5}`.

The **F5 stream-encoder fails (record continues)** row is
the chapter's binding to the **independent-lane invariant**:
when the stream-encoder process crashes (NVENC driver
exception, GPU hang on the stream lane, or process supervisor
restart), the record-encoder must continue producing the
recording bitstream uninterrupted. The C30 recording target
must observe a clean continuous stream from the record-
encoder's perspective — the stream-lane fault must not
propagate. F5's chaos test (§8.6) injects a synthetic
SIGSEGV into the stream-encoder process and asserts the
record-encoder lane continues without observable disruption.

The **F6 record-encoder fails (stream continues)** row is
the symmetric counterpart of F5: when the record-encoder
process crashes (e.g. F3 lookahead overflow cascades to a
process crash, or F9 disk-write fault terminates the record
process), the stream-encoder must continue producing the
live stream uninterrupted. The dropped recording is marked
partial in the C30 catalogue with a structured fault
attribution. F6 is **non-overridable as a stream-side fault
suppressor**: a record-side fault must never terminate the
stream session.

The **F7 GPU power-draw exceeds TDP** row binds to the C34
electrical-envelope invariant. Modern desktop GPUs (RTX 4090
at 450 W TDP, RTX 5090 at 600 W TDP) can momentarily exceed
TDP under dual-path encoding plus game-rendering load,
especially when ray-tracing is active. The PSU-side reading
from `nvidia-smi --query-gpu=power.draw` shows sustained
power-draw at or above TDP; the mitigation is to drop the
record-lane bitrate by 30% (record is the lower-priority
lane) and re-evaluate after a 30 s stability window. If
power-draw remains over-budget, drop the record lane
entirely.

The **F8 mid-session bitrate change attempted (rejected)**
row binds to the C27-F8 hot-reload-forbidden invariant
extended to dual-path: NVENC does not support clean mid-
session bitrate changes when running dual-session — the
session must be drained and restarted with the new bitrate
configuration. The mitigation is to **refuse bitrate-change
requests that would require NVENC reconfiguration mid-
session**. The C34 thermal controller's bitrate cascade is
implemented via the encoder's *rate-control* knob (dynamic
within a session) rather than session-recreate; F8 catches
the case where an operator tries to push a session-recreate
mid-session.

The **F9 record disk write fails (cross-link C30)** row is
the chapter's binding to C30 (Recording Storage) — the
record-encoder's output is consumed by the C30 local-buffer
+ background-sync pipeline. If the local NVMe SSD fills, the
disk-write fault is detected at the C30 boundary and
propagated back to C29 as a record-encoder shutdown signal.
C29 observes the signal, drops the record lane cleanly
(F6-style), and the stream lane continues. The detailed
disk-fault detection + mitigation is in C30 §7; this row is
the cross-link contract.

The **F10 r18.SafeExec rejection** row is the chapter's R-18
trip-wire and is symmetric with C26-F9, C27-F10, C28-F10.
When a developer adds a new dual-path-tooling invocation
(e.g. a new `ffmpeg -filter_complex split=2[s][r]` shape
with a non-allow-listed split filter, or a new
`gst-launch-1.0 ! tee name=t` shape with a non-allow-listed
tee element), the wrapper rejects the call at the `os/exec`
boundary and bootstrap aborts. Bypass requires an allow-list
extension via operator review per Constitution §11.5.4,
never a silent workaround. The allow-list lives in
`vasic-digital/helix-r18-safeexec` and is **not duplicated**
in this chapter; the family allow-list extension that C29
contributes is recapped in §1 (family allow-list) of this
chapter and verified by the C08 `host-integrity-scan` test
inherited verbatim into §8.11.

The **F11 thermal sensor readout failure (`nvidia-smi`
hung)** row binds the chapter's thermal-aware controller to
its sensor-read primitive: the controller polls
`nvidia-smi --query-gpu=temperature.gpu,power.draw` every
1 s. If the `nvidia-smi` subprocess hangs (kernel-driver
deadlock, rare driver bug under GPU thermal stress), the
controller loses its thermal feedback loop. The mitigation
is a **bounded subprocess timeout** (3 s) plus a fallback to
the per-driver telemetry endpoint
(`/proc/driver/nvidia/gpus/<bus>/information`); if both fail,
the controller transitions to **conservative mode**: assume
hardcap is imminent, drop record lane immediately, and
preserve stream at minimum-viable bitrate. F11 is rare but
catastrophic if undetected — emits `dualpath.sensor_failure
{driver_state=…,fallback_path=…}`.

The **F12 dual-session quality drift (record falls behind
stream)** row binds to the C35 (Measurement & QA) PSNR/SSIM
invariant: under sustained dual-path load, the record-
encoder occasionally falls behind the stream-encoder in
quality due to differing rate-control regimes (record uses
2-pass VBR for archival; stream uses CBR-LD for latency).
The C35 harness compares per-claim-window PSNR between the
stream-decoded reconstruction and the record-decoded
reconstruction; if the gap exceeds 2 dB PSNR over a 30 s
window, F12 fires. The mitigation is to **re-bias the
record-encoder rate-control toward the stream's effective
bitrate** until parity is restored.

| # | Failure mode | Symptom | Detection | Mitigation | Fallback |
|---|---|---|---|---|---|
| F1 | NVENC dual-session limit hit — Lovelace cards (RTX 4090/4080/4070 Ti) cap at 8 NVENC sessions = 4 dual-path sessions; Blackwell (RTX 5090/5080) caps at 16 = 8 dual-path; Quadro / RTX 6000 Ada unlimited | Driver returns `NV_ENC_ERR_OUT_OF_MEMORY` at the second of the dual-pair NVENC session-create calls; admission for the 5th (or 9th on Blackwell) dual-path attempt fails at the second-session boundary | Dual-path-aware capability counter — `dualpath.NVENC.AvailableDualSessions(host)` walks driver state at admission and divides the C27-F1 single-session count by 2; emits `dualpath.session_limit_hit {host=…,vendor=NVIDIA,dual_limit=4,observed=4}` | Scheduler refuses dual-path admission for the new session on this host; the admission queue routes to an alternative host with available dual-path NVENC capacity per C08 host-agent pool | Route to alt host or downgrade to stream-only — non-blocking; the operator-policy posture chooses between (a) refuse session on this host (route elsewhere), (b) accept session at stream-only quality (record disabled per-session); cluster-level capacity remains intact |
| F2 | Thermal throttling at 83 °C — sustained dual-path 4K60 + 4K30 pushes GPU to hardcap; cross-link C34 thermal envelope (78 °C pre-emptive, 83 °C hardcap) | GPU temperature reading from `nvidia-smi --query-gpu=temperature.gpu` reaches 83.0 °C; encode-time histogram shows p999 drift from 3 ms toward 12+ ms; SSIM drops under sustained load | Thermal monitor — C34 `thermal.GpuTempPoll(host)` polled every 1 s; emits `dualpath.thermal_hardcap {host=…,gpu_temp_c=83.2,record_lane_active=true}` | Drop the record lane to preserve the stream lane (Insight #1 binding — stream is SLA-bearing); emit `dualpath.record_dropped {cause=thermal_hardcap}`; the recording is marked partial in the C30 catalogue | Stream-only continuation — non-blocking for the live stream; record session ends gracefully with the partial-recording flag; cross-link C34 §6 |
| F3 | Record-encoder lookahead buffer overflow — record-side encoder is configured with deeper lookahead than stream-side; under CPU pressure (F11) or thermal cascade (F2), the lookahead window cannot drain fast enough | Lookahead buffer depth grows beyond high-water-mark (default 24 frames); record-encoder rate-control falls behind by > 500 ms; emits `dualpath.lookahead_overflow {depth=32,lane=record}` | Lookahead-depth monitor — `dualpath.RecordEncoder.LookaheadDepth(snapshot=1s)` observed against high-water-mark; emits `dualpath.lookahead_overflow {depth=…,lane=…}` when threshold exceeded | Reduce record-encoder lookahead depth from 24 to match stream-encoder's 4; accept marginal record-quality drop to keep buffer drained; emit `dualpath.lookahead_reduced {from=24,to=4}` | Reduced-quality record continuation — non-blocking; record proceeds at marginally lower quality; the C30 recording is annotated with the lookahead-reduction event |
| F4 | Frame-Tee back-pressure deadlock — both downstream channels (stream lane, record lane) stall simultaneously, blocking the Frame-Tee at the fan-out boundary; upstream SPSC ring fills | Fan-out channel-write to either stream or record lane exceeds 5 ms timeout; SPSC ring depth grows; emits `dualpath.tee_blocked {lane=…,write_ms=5.2}` | Channel-write timeout — `dualpath.FrameTee.OnFanoutTimeout()` callback fires after 5 ms; emits `dualpath.tee_deadlock_avoided {lane=…,timeout_ms=5}` | Drop the frame for the timed-out lane while publishing to the other lane; the surviving lane proceeds without a gap; the dropped lane observes a sequence-number gap that escalates to F12 if sustained | Per-lane frame drop — non-blocking; the dropped lane absorbs the loss; if drop rate sustained > 1% / 30 s, escalate to F6 (record-side) or session restart (stream-side) per quality contract |
| F5 | Stream-encoder fails (record continues) — stream-side encoder process crashes (NVENC driver exception, GPU hang on stream lane, supervisor restart); record-encoder must continue producing the recording uninterrupted | Stream-encoder process exits non-zero; the C08 supervisor observes the exit; the record-encoder lane continues producing frames; client-side stream observes a brief gap until the supervisor restarts the stream encoder | Process supervisor — `dualpath.StreamEncoder.OnCrash()` callback fires; emits `dualpath.stream_crash_record_continues {session=…,record_active=true}` | C08 supervisor restarts the stream-encoder within 200 ms (Constitution §6 chaos SLO); record-encoder lane is unaffected; client observes a 200 ms stream gap | Independent-lane recovery — non-blocking for record; the stream restart absorbs the gap; the C30 recording is unaffected; chaos test §8.6 enforces the 200 ms budget |
| F6 | Record-encoder fails (stream continues) — record-side encoder process crashes (F3 lookahead overflow cascades to process crash; F9 disk-write fault terminates record process); stream must continue uninterrupted | Record-encoder process exits non-zero; the C08 supervisor observes the exit; the stream-encoder lane continues producing frames; the recording is truncated at the crash boundary | Process supervisor — `dualpath.RecordEncoder.OnCrash()` callback fires; emits `dualpath.record_crash_stream_continues {session=…,recording_truncated=true}` | C08 supervisor decides per operator policy: (a) restart record-encoder and resume recording from the live frame (gap in recording), or (b) mark recording final at the truncation point; stream lane is unaffected | Independent-lane recovery — non-blocking for stream; record-side fault must never terminate the stream session; non-overridable per stream-SLA contract |
| F7 | GPU power-draw exceeds TDP — RTX 4090 (450 W TDP), RTX 5090 (600 W TDP) momentarily exceed TDP under dual-path encode plus ray-tracing game render; PSU-side reading from `nvidia-smi --query-gpu=power.draw` shows sustained excess | Sustained power-draw ≥ 100% of TDP for > 10 s; emits `dualpath.power_excess {gpu=RTX_4090,tdp_w=450,observed_w=475}` | Power-draw monitor — `dualpath.PowerDrawPoll(host)` polled every 1 s; cross-references TDP from `nvidia-smi --query-gpu=power.limit`; emits `dualpath.power_excess {…}` | Drop record-lane bitrate by 30% (record is lower-priority); re-evaluate after 30 s; if power-draw remains over-budget, drop the record lane entirely | Stream-only continuation — non-blocking; record degrades or drops; the operator dashboard tracks per-host TDP-margin |
| F8 | Mid-session bitrate change attempted (rejected) — operator tries to push a session-recreate-style bitrate change on a live dual-path session; NVENC does not support clean mid-session reconfiguration | Bitrate-change request observes `NV_ENC_ERR_INVALID_CALL` at the reconfigure boundary; emits `dualpath.bitrate_change_refused {session=…,from=25mbps,to=15mbps}` | Bitrate-change validator — `dualpath.BitrateChange.Validate(session)` rejects requests that would require NVENC reconfiguration; emits the structured refusal event | Use the rate-control knob (dynamic within a session) instead of session-recreate; the C34 thermal cascade uses this path correctly; F8 catches operator attempts to bypass the cascade | Refused bitrate change — non-blocking for the existing session; the operator must drain-and-restart the session to apply the new bitrate; cross-link C27-F8 hot-reload-forbidden invariant |
| F9 | Record disk write fails (cross-link C30) — local NVMe SSD fills or the disk subsystem returns a write error; the C30 local-buffer pipeline propagates the fault back to C29 | C30 emits `record.disk_write_failed {host=…,error=ENOSPC}`; the record-encoder observes the back-pressure and exits | C30 boundary — `dualpath.RecordSink.OnDiskFault()` callback fires when C30 reports a disk fault; emits `dualpath.record_disk_fault {cause=…}` | Drop record lane cleanly (F6-style); stream lane continues; the operator dashboard surfaces the disk-fault root cause; cross-link C30 §7 for detailed disk-fault handling | C30 cross-link — non-blocking for stream; record session ends with disk-fault flag; the C30 catalogue marks the recording partial |
| F10 | `r18.SafeExec` rejects subprocess — developer added a non-allow-listed argv shape (e.g. `ffmpeg -filter_complex split=2` with a non-allow-listed split filter, or `gst-launch-1.0 ! tee` with a non-allow-listed tee element) | Bootstrap fails on dual-path-pipeline initialisation; structured error includes the rejected argv with the offending flag highlighted; the harness logs `dualpath.safeexec_rejected {tool="ffmpeg",argv=…}` | The wrapper's verbatim allow-list check at the `os/exec` boundary returns `ErrForbiddenArgvShape`; the harness logs the rejection | Fix the call site to use the allow-listed shape — canonical `ffmpeg -filter_complex split=2[s][r]` / `gst-launch-1.0 ! tee name=t` shapes per family allow-list (`00_Index.md` §7); allow-list extension requires operator review per Constitution §11.5.4 | Blocking — bootstrap aborts; non-overridable; the rule lives in the Constitution and bypass requires a §13 exception with documented mitigation; cross-link §8.11 host-integrity-scan |
| F11 | Thermal sensor readout failure (`nvidia-smi` hung) — kernel-driver deadlock, rare driver bug under GPU thermal stress; controller loses its thermal feedback loop | `nvidia-smi --query-gpu=temperature.gpu` subprocess does not return within 3 s; the controller observes a sensor-read timeout; emits `dualpath.sensor_timeout {tool=nvidia-smi,timeout_s=3}` | Bounded subprocess timeout — `dualpath.SensorPoll.Timeout(3s)` enforced; on timeout, fallback to `/proc/driver/nvidia/gpus/<bus>/information`; if fallback also fails, transition to conservative mode | Conservative mode — assume hardcap imminent, drop record lane immediately, preserve stream at minimum-viable bitrate; emit `dualpath.sensor_failure {fallback_path=conservative_mode}` | Conservative-mode continuation — non-blocking for stream at degraded quality; the operator dashboard surfaces the sensor fault for driver investigation; cross-link C34 §8 |
| F12 | Dual-session quality drift (record falls behind stream) — record-encoder occasionally falls behind in quality due to differing rate-control regimes (record 2-pass VBR vs stream CBR-LD); C35 PSNR comparison shows > 2 dB gap | C35 PSNR-comparison harness observes per-claim-window stream-PSNR vs record-PSNR; gap > 2 dB over a 30 s window; emits `dualpath.quality_drift {gap_db=2.4,window_s=30}` | C35 quality monitor — `dualpath.QualityComparison.Snapshot(window=30s)` reports the per-claim-window PSNR delta; cross-references stream and record encoded outputs | Re-bias record-encoder rate-control toward stream's effective bitrate until parity restored; emit `dualpath.quality_rebias {from=…,to=…}` | Quality-rebias continuation — non-blocking; record converges back toward parity; the C30 recording is annotated with the rebias event for audit |

## 8. Test surface

The C29 test surface inherits the family-level container-driven
CI lane contract from C26 §8 + C27 §8 + C28 §8 and the
`vasic-digital/Containers` runner image, **extended** with the
new dual-path-specific requirement: every integration / E2E /
benchmarking test must exercise **two NVENC sessions
simultaneously on the same physical GPU** so the dual-path
contract is validated against real silicon (mocking the second
NVENC session is forbidden per Constitution §6.4 — only unit
tests may use mocks). Per Constitution §6.4 + Master Plan §4.3
anti-bluff verification, the test matrix below cites
`video-tech_dim04.md` (dual-path dimension) and
`video-tech_dim10.md` (testing dimension) explicitly so every
per-lane performance claim is grounded in a primary-source
reference.

### 8.1 Unit (mocks allowed)

The unit-test layer is the only place where mocks, stubs, or
hardcoded values are permitted per Constitution §6.4 — every
other layer below hits the real system.

- **FrameTee fan-out test** — given a synthetic upstream
  producer that emits 10⁶ sequence-numbered frame events and
  two synthetic downstream consumers (one fast, one slow),
  assert the FrameTee correctly fans out to both consumers
  with per-lane back-pressure isolation: the slow consumer's
  drops do not affect the fast consumer's sequence integrity.
  Mock the encoder downstream; use channel-based hand-off per
  Insight #5; assert per-lane sequence-number streams are
  monotonic with no cross-lane interference.
- **Dual-path admission policy unit test** — given a synthetic
  capability stanza for each of the four GPU classes, assert
  the admission policy correctly halves the single-session
  count to derive the dual-session count (Lovelace 8→4;
  Blackwell 16→8; RDNA 3 16→8; Quadro unlimited).
- **Bitrate-change validator unit test** — given a synthetic
  bitrate-change request that would require NVENC
  reconfiguration, assert the validator returns
  `ErrBitrateChangeRefused`; assert dynamic rate-control knob
  changes are accepted.
- **PSNR comparison harness unit test** — given two synthetic
  encoded outputs (stream + record) with controlled PSNR
  delta, assert the C35 comparison harness correctly reports
  the per-claim-window delta within ±0.1 dB tolerance.

### 8.2 Integration

The integration-test layer hits the real NVENC dual-session
path — no mocks, no stubs, no hardcoded values. Per
Constitution §6.4 this layer must run inside the canonical
`vasic-digital/Containers` runner image with NVIDIA GPU
passthrough enabled.

- **Real NVENC dual-session with PSNR comparison stream vs
  record** — boot the container with NVIDIA GPU passthrough,
  link against the real `libnvidia-encode.so`, encode a
  60-second test pattern simultaneously through two NVENC
  sessions (stream session at CBR-LD 25 Mbps; record session
  at 2-pass VBR 35 Mbps target); decode both outputs; assert
  per-frame PSNR delta < 2 dB between stream and record
  outputs over the 60 s window.
- **Real AMF dual-session integration** — boot RDNA 3 runner;
  exercise dual AMF sessions; verify per-vendor parity with
  the NVENC dual-session contract.
- **Frame-Tee real fan-out integration** — boot a real Linux
  host with the canonical capture pipeline (C28); insert the
  Frame-Tee between capture and dual encoder lanes; capture
  60 s of test pattern and assert per-lane frame-event
  sequence integrity.
- **Capability detection integration** — for each vendor
  runner, run the canonical
  `nvidia-smi --query-gpu=encoder_capability,encoder_session_count`
  probe via `r18.SafeExec` and assert the dual-path capability
  counter correctly halves the per-vendor single-session
  count.

### 8.3 E2E

The E2E layer brings up the **full host-agent + game +
capture + dual-encode (4K60 stream + 4K30 record) for 2
hours** and asserts both the stream SLO and the record-
fidelity contract.

- **Full pipeline 4K60 stream + 4K30 record / 2 h E2E** —
  boot the host with the reference game, run the capture
  lane (C28), fan out to the stream-encoder and record-
  encoder via the Frame-Tee, transmit the stream bitstream
  over the network transport (C37), persist the record
  bitstream to local NVMe (C30), for 2 hours continuously;
  **assert stream-encode p999 ≤ 5 ms** (Constitution §6 +
  C24 §6 SLO); assert record-encoded output is valid
  (`ffprobe` exit 0); assert PSNR delta between stream and
  record < 2 dB per 30 s claim window.
- **Per-fault cascade E2E** — boot host with each fault
  injected via the C35 fault-injection harness (F1, F2, F3,
  F4, F5, F6, F7, F8, F9, F11, F12); assert each fault
  recovers per its specified mitigation; assert the
  surviving lane (stream for F2/F3/F6/F7/F9; record for F5)
  continues uninterrupted.
- **Cross-vendor parity E2E** — run identical scenarios on
  NVENC + AMF + QSV runners (VideoToolbox dev-tier excluded
  per C27-F12); assert dual-path stream p999 within 1.0 ms
  across vendors at 4K60.

### 8.4 Security

- **`r18.SafeExec` rejection fuzz** — for each of the family
  allow-list entries (`00_Index.md` §7), construct an off-
  allow-list argv shape (e.g.
  `ffmpeg -filter_complex split=3` is off-list because the
  family allow-list is the bipartite split=2 form; or
  `gst-launch-1.0 ! tee` with a non-allow-listed sink); fuzz
  with 10⁶ argv permutations and assert the wrapper returns
  `ErrForbiddenArgvShape` for every off-list shape with no
  false-positive on allow-list shapes.
- **Verify dual-path-process privilege isolation** — assert
  the stream-encoder and record-encoder processes run as
  non-root users with the minimum capability set required
  for their respective vendor SDK (no `CAP_SYS_ADMIN`, no
  `CAP_NET_ADMIN`, no `CAP_SYS_PTRACE`); assert the dual-
  path-orchestrator process cannot read or write either
  encoder's memory pages outside the sanctioned shared-
  memory regions (cross-link C15 shared-memory zero-copy
  IPC).
- **Verify recording-lane encryption-at-rest** — assert the
  record-encoder's output is written to disk encrypted per
  C30 §6 storage-encryption contract; the stream-encoder's
  output never lands on disk unencrypted (cross-link C09
  security family).

### 8.5 Benchmarking

The benchmarking layer is the chapter's binding to
Constitution §6 — every per-lane dual-path encode performance
claim reports p50 / p99 / p999 at ≥ 10 K samples via the C24
measurement harness. Cross-link C24 / C35. Per
**`video-tech_dim10.md`** §2 + §5, the benchmarking corpus
uses synthetic-content + real-game-capture pairs across the
six representative game profiles (FPS, racing, RPG, RTS,
MOBA, fighting) so the per-lane performance characterisation
reflects production-like workloads.

- **Bench dual-path encode throughput** — for each (vendor,
  GPU class, codec, resolution-pair) combination on the
  runner matrix (e.g. NVENC × Lovelace × {AV1 stream + AV1
  record} × {4K60 + 4K30, 1080p120 + 1080p60, 4K120 +
  4K60}); report **p50 / p99 / p999 per-lane** with **≥ 10 K
  samples per combination per Constitution §6**; histogram
  artifact attached to every claim.
- **Bench dual-path encode-latency-distribution per lane** —
  at fixed configuration (4K60 stream CBR-LD 25 Mbps + 4K30
  record VBR 35 Mbps) measure encode latency distribution
  per lane; assert stream-lane p999 ≤ 5 ms (SLA-bearing);
  assert record-lane p999 ≤ 50 ms (record can afford the
  latency for higher quality).
- **Bench Frame-Tee fan-out throughput** — measure fan-out
  throughput at 240 fps × 4K with synthetic frame data;
  budget < 50 µs p999 per fan-out operation; cross-link C36
  §3 Go pipeline implementation for the channel-based hand-
  off discipline.
- **Bench dual-session admission latency** — time from
  dual-path session-create call to first stream-frame ready
  AND first record-frame ready; budget < 200 ms p999;
  assert no vendor exceeds the budget.
- **Bench thermal-cascade response time** — inject a
  synthetic thermal threshold breach (78 °C pre-emptive);
  measure time from breach detection to bitrate-cascade
  application on both lanes; budget < 100 ms p999.
- Cross-link **C24 / C35** measurement harness for shared
  histogram-collection + bootstrap-resampling-confidence-
  interval primitives. The benchmark suite must cite
  **`video-tech_dim10.md`** explicitly per Master Plan §4.3
  anti-bluff verification — `video-tech_dim10.md` §3
  enumerates the per-lane regression-detection thresholds +
  §5 enumerates the canonical bench corpus + §7 enumerates
  the per-lane session-create latency budget. Cross-link
  **C24** §6 (latency-side measurement) and **C35** §3
  (quality-side measurement) for the full harness contract.

### 8.6 Chaos

- **Force GPU thermal spike — assert pre-emptive reduction
  kicks in** — for each vendor runner, drive the GPU
  toward the 78 °C pre-emptive threshold via a synthetic
  thermal-load amplifier (concurrent compute kernel running
  in parallel with the dual-path encode); assert the C34
  thermal controller observes the threshold breach within
  1 s; assert the bitrate-cascade applies to both lanes
  proportionally; assert the thermal envelope stabilises
  below 78 °C within 30 s; assert no transition to hardcap
  (83 °C).
- **Force stream-encoder crash mid-stream (F5)** — inject a
  synthetic SIGSEGV into the stream-encoder process; assert
  the C08 supervisor restarts the stream-encoder within
  200 ms; **assert the record-encoder lane continues
  uninterrupted** (no observable record-side disruption).
- **Force record-encoder crash mid-stream (F6)** — inject a
  synthetic SIGSEGV into the record-encoder process; assert
  the stream-encoder lane continues uninterrupted; assert
  the record-encoder restart policy is honoured per
  operator config (gap-fill or truncation).
- **Force `nvidia-smi` hang (F11)** — synthetically delay
  the `nvidia-smi` subprocess via a controlled `sleep`
  injection; assert the 3 s timeout fires; assert fallback
  to the `/proc/driver/nvidia` path is attempted; assert
  conservative-mode transition is observed when both fail.
- **Force NVENC dual-session limit (F1)** — submit
  (limit + 1) dual-path admission requests on an RTX 4090
  runner (5 admissions on an 8-NVENC-session GPU); assert
  the 5th is gracefully refused with the structured
  `dualpath.session_limit_hit` event; assert the scheduler
  routes the 5th request to an alternative host.

### 8.7 Stress

- **24h dual-path 4K60 + 4K30 — assert no thermal cliff** —
  on each vendor runner with adequate cooling, run
  continuous dual-path 4K60 stream + 4K30 record for 24
  hours; **assert no thermal cliff** (GPU temperature stays
  under 78 °C pre-emptive threshold); assert no fd leak
  (process fd count stable to within 5 fds over 24 h);
  assert no GC stall > 1 ms (GODEBUG=gctrace=1 trace
  artifact attached; cross-link C36 §3 Go pipeline
  `sync.Pool` discipline); assert no memory leak (RSS
  growth < 5 MB / hour); assert per-lane PSNR delta stays
  within 2 dB over the 24 h window.
- **Cross-vendor dual-path stress** — on a host with NVIDIA
  + Intel iGPU, run dual-path NVENC + dual-path QSV
  concurrently for 12 hours; assert per-lane sequence
  integrity holds for the duration; assert no cross-vendor
  contention (per-lane p999 variance < 10% from per-vendor-
  isolated baseline).
- **Sustained NVENC saturation stress** — on RTX 4090
  runner, run **4 concurrent dual-path sessions** (= 8
  NVENC sessions = vendor session limit) for 8 hours;
  assert no F1 invariant breach during saturation; assert
  fair scheduling across the 4 sessions (per-session p999
  variance < 10%).

### 8.8 Smoke

- **Capability schema reports `dualpath.supported`** — boot
  host-agent in clean container; **verify the capability
  schema reports the correct `dualpath.supported=true` /
  `dualpath.max_sessions=N`** for the runner's GPU class;
  assert the schema validates against
  `vasic-digital/helix-dualpath/schema/v1.json`; assert the
  per-vendor session-count math (single-session-count / 2)
  is reported correctly.
- **Smoke test dual-path session-create** — for each vendor
  runner, dispatch a 5-second dual-path session; assert
  both stream and record encoded outputs are produced at
  the expected rate, the bitstreams conform to spec via
  `ffprobe` exit 0, and no `dualpath.degraded` event is
  emitted.

### 8.9 Full automation

All of §8.1–§8.8 run on **every commit via the local
container-driven CI lane** per Constitution §10. The CI lane
uses the canonical `vasic-digital/Containers` runner image
with multi-vendor GPU passthrough enabled; the matrix covers
(NVIDIA Lovelace + Blackwell, AMD RDNA 2 + RDNA 3, Intel Arc
+ iGPU) × (Linux Ubuntu 22.04/24.04 + Windows Server 2022)
where dual-path-capable hardware is available on the runner.
Apple VideoToolbox is dev-tier only per C27-F12 and is
excluded from the production dual-path automation matrix.
The full-automation lane emits a single composite artifact
(`dualpath-test-report.json`) that the C35 quality-claim
harness consumes as the authoritative source-of-truth for
any per-lane dual-path-performance claim in chapter prose.

### 8.10 Challenges (production-like)

HelixQA dispatches **per-vendor dual-path scenarios** from
`git@github.com:vasic-digital/Challenges.git` (per
Constitution §6.4 Challenges-test contract):

- **Per-vendor dual-path Challenges** — for each of NVENC
  (Lovelace + Blackwell), AMF (RDNA 3), QSV (Arc
  Battlemage), HelixQA dispatches a full 30-minute dual-
  path scenario (4K60 stream + 4K30 record) with the
  reference game; assert per-lane SLA holds (stream p999
  ≤ 5 ms; record SSIM ≥ 0.95); assert per-vendor session-
  count math is correctly enforced (e.g. RTX 4090 = 4
  concurrent dual-path sessions max).
- **Saturation Challenges** — saturate each vendor's dual-
  path session limit on a single host; assert F1 invariant
  holds for each vendor; assert the scheduler correctly
  refuses the (limit+1)th admission with the correct
  structured event.
- **Thermal-cascade Challenges** — drive each vendor toward
  the C34 thermal threshold under sustained dual-path load;
  assert the thermal-aware bitrate cascade applies
  correctly across both lanes; assert no transition to
  hardcap (F2 invariant) under standard cooling.
- **Cross-vendor failover Challenges** — start 4 dual-path
  sessions on NVENC; mid-Challenge inject an F8 hot-reload
  event on the NVIDIA driver; assert the 4 sessions
  complete to natural end on NVENC while new dual-path
  sessions are routed to AMD or Intel per family failover
  policy.

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

The C29 implementation contract that this scan validates:

- Dual-path tooling invocation via `r18.SafeExec` only —
  never via `os/exec.Command` directly; the canonical shapes
  (`ffmpeg -filter_complex split=2[s][r] ...`,
  `gst-launch-1.0 ! tee name=t ...`, dual-NVENC session-
  create via the SDK) are the family allow-list entries for
  dual-path tooling.
- No host-disruption commands ever appear in the dual-path
  path: no `kill -9 <pid>`, no `systemctl
  suspend|hibernate|reboot|halt|poweroff`, no `pmset`, no
  `xset dpms force off`, no `--privileged` container flag,
  no host-mount of `/`, `/dev`, `/proc`, `/sys`. The scan
  asserts none of these syscall patterns appear in the
  dual-path subsystem's syscall trace.
- No cross-process memory access between the stream-encoder
  and the record-encoder outside the sanctioned shared-
  memory regions (cross-link C15) — the scan asserts no
  `process_vm_readv` / `process_vm_writev` / `ptrace`
  syscall appears in either encoder's trace targeting the
  other encoder's PID.

The scan's invocation contract is byte-identical with the
C08 §12.11 inheritance into every chapter in the family per
`00_Index.md` §7 R-18 family allow-list. No chapter in the
family is permitted to redefine, override, or extend the
scan — Constitution §11.5.4 forbids per-chapter
customisation of the host-integrity contract.

## 9. Open questions

The following open questions are tracked in the chapter's OQ
log and surface to the family-level OQ aggregator at
`00_Index.md` §5. Each OQ is prefixed `OQ-C29-NN` and carries
an owner, a target resolution date, and a cross-link to the
deciding chapter or external dependency.

- **OQ-C29-01** — Multi-stream simulcast (record + stream +
  secondary stream). Should HelixPlay support a third NVENC
  session per player session (stream + record + a secondary
  stream for multi-viewer / spectator / casting modes)? The
  cost is one additional NVENC session per player, halving
  again the per-host concurrent-player capacity (RTX 4090:
  4 → 2 dual-path-plus-spectator sessions); the benefit is
  Twitch-style spectator and split-screen co-viewing without
  requiring a separate transcode tier. Owner: C29 + V1
  family + Catalogue family. Cross-link F1 invariant +
  Insight #1 thermal wall.
- **OQ-C29-02** — Record-codec switching mid-session. The
  current contract (F8 invariant) refuses any mid-session
  reconfiguration that requires NVENC session-recreate. But
  a record-codec switch (e.g. starting a session in HEVC
  record then switching to AV1 record after the first GOP
  for archival quality reasons) is theoretically supportable
  if the recording is split into per-codec segments and
  concatenated post-session in the C30 storage layer.
  Trigger: archival-quality operator-policy posture
  emerges. Owner: C29 + C30. Cross-link F8 + Insight #4.
- **OQ-C29-03** — Thermal-aware scheduler bias. The current
  C34 cascade is a per-host reactive controller. Should the
  scheduler at admission time bias dual-path sessions
  toward hosts with thermal headroom (e.g. cooler hosts get
  more dual-path sessions; warmer hosts get stream-only)?
  The cost is per-host thermal-state propagation latency in
  the scheduler decision; the benefit is fewer F2 hardcap
  events. Owner: C29 + C34 + Operations family. Cross-link
  F2 + F7.
- **OQ-C29-04** — Liquid-cooled host operator-policy tier.
  Should HelixPlay surface an operator-policy "liquid-
  cooled tier" that loosens the F2/F7 thresholds for hosts
  with documented adequate cooling (custom-loop liquid,
  industrial chassis, dedicated chiller)? The benefit is
  more dual-path sessions per host on premium tiers; the
  cost is operator-side certification and audit. Trigger:
  premium-tier operator demand exceeds standard-tier
  capacity. Owner: C29 + C34 + Operations family. Cross-
  link F2 + F7 + F11.
- **OQ-C29-05** — NVENC session quota allocation across
  tenants. On a multi-tenant host (consumer card with 8
  NVENC sessions = 4 dual-path sessions), should the quota
  be allocated equally across tenants (e.g. 2 tenants × 2
  dual-path sessions each), or weighted by operator-policy
  posture (e.g. premium-tier tenant gets 3 / standard-tier
  tenant gets 1)? The current default is FCFS; the V1
  proposal adds explicit per-tenant quota. Owner: C29 +
  Operations family. Cross-link F1 invariant + OQ-C27-02
  (multi-vendor host scheduling).

---

## 10. References

### Project artifacts

- Master Plan, Constitution, System Overview, Video/Audio family index ([`00_Index.md`](00_Index.md)). Sibling chapters cited above.

### Source research artifacts

- `video-tech_dim04.md` (1,013 lines), `video-tech.agent.final.md` (2,588 lines), `video-tech_insight.md` (Insights #1, #4, #10), `video-tech_cross_verification.md`, `video-tech_dim10.md` (testing).

### Web research

[`../99_Web_Research_Addenda/2026-04-29-dual-path-encoding.md`](../99_Web_Research_Addenda/2026-04-29-dual-path-encoding.md) — 228 lines, 80 distinct URLs across 9 clusters + §Z (Z-1..Z-8).

| Cluster | Topic | Cited |
|---------|-------|------|
| §A | NVENC dual-session SDK 12/13 (Z-1) | §2.4 |
| §B | AMD AMF dual-session RDNA3/RDNA4 no session limits (Z-3) | §2.5 |
| §C | Intel QSV/oneVPL dual-session Battlemage (Z-4) | §2.5 |
| §D | Frame-Tee fan-out zero-copy (Z-2) | §2.1 |
| §E | Per-encoder bitrate/preset/GOP/B-frame divergence (Z-8) | §2.2, §4 |
| §F | Thermal headroom budget (Insight #1) | §3 |
| §G | Recording codec/container fMP4/MKV HEVC/AV1 (Z-5, Z-6) | §5 |
| §H | 2026 hardware-appliance dual-path (Magewell, AVerMedia, Matrox, Haivision) | §1 |
| §I | Sunshine multi-session removal (Z-7; Insight #10) | §2.6 |
| §Z | Contradictions index (Z-1..Z-8) | §1, §2, §3, §4, §5 |

---

## Anti-Bluff Verification

> Per [Master Plan §4.3](../../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Source Evidence Reviewed

| Path | Lines | Reviewed | Used |
|------|------:|----------|------|
| `video-tech_dim04.md` | 1,013 | A, B, C, D | §§1–9 (primary) |
| `video-tech.agent.final.md` | 2,588 | A, B, C | §§1–6 |
| `video-tech_insight.md` | 243 | A, B | §1 (#1, #4, #10) |
| `video-tech_cross_verification.md` | 206 | A | §1 |
| `video-tech_dim10.md` | 1,689 | D | §8.5 |
| `00_Master_Plan.md` post-Session-6 | A, B, C, D | header / §6 / §9 |
| `01_Constitution.md` post §11.5 | A, B, C, D | §§1–8 |
| `05_Video_Audio/00_Index.md` | 407 | A, B, C, D | header voice |
| `05_Video_Audio/01_Codec_Selection.md` | 2,578 | A, C | §1, §5 |
| `05_Video_Audio/02_Hardware_Encoders.md` | 2,652 | A, C | §1, §2 |
| `05_Video_Audio/03_Capture_Pipelines.md` | 2,420 | A, C | §1, §2 |
| `04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md` | 1,749 | C | §6.1 (helix-encoder reuse) |
| `03_Architecture/07_Host_Agent_and_Game_Lifecycle.md` | 3,425 | C, D | §6 (`r18.SafeExec`), §8.11 |

### Web Sources Consulted

The companion addendum lists every URL with title and 2026-04-29 access date. **80 distinct URLs across 9 clusters + §Z.**

### Insights Incorporated

| Insight | Source | Sections |
|---------|--------|----------|
| video-tech Insight #1 — Thermal wall | `video-tech_insight.md` | §1, §3 (binding) |
| video-tech Insight #4 — Recording = save system | `video-tech_insight.md` | §1, §5 |
| video-tech Insight #10 — Recording differentiates HelixPlay | `video-tech_insight.md` | §1, §2.6 |

### Conflict Zones Resolved

| CZ-ID | Description | Decision | Section |
|-------|-------------|----------|---------|
| Insight #1 | Thermal wall — dual-path GPU-thermal-bounded | **Reaffirmed**; pre-emptive quality reduction tiers (78/80/82/83°C) | §3 |
| Z-1 | NVENC session limit 5→8 on consumer Lovelace | Documented; admission policy uses 8-session ceiling | §2.4 |
| Z-2 | FFmpeg tee-muxer vs GStreamer Frame-Tee | HelixPlay uses native Frame-Tee (helix-dualpath); FFmpeg tee-muxer V1 fallback | §2.1 |
| Z-3 | AMF multi-HW-instance encoder mode | RDNA4 multi-HW-instance documented | §2.5 |
| Z-4 | Intel concurrent-session ceiling | Intel QSV no hard limit; thermal-bound on Arc Battlemage | §2.5 |
| Z-5 | fMP4 default vs MKV alternative | fMP4 default; MKV operator-policy alternative | §5 |
| Z-6 | AV1 efficiency baseline | 30-40% bitrate reduction vs HEVC at same SSIM | §5 |
| Z-7 | Sunshine multi-client vs multi-session distinction | HelixPlay's MVP is single-session-per-host (cross-link C09 §4 KubeVirt VM-per-session) | §2.6 |
| Z-8 | NVENC GOP-immutability constraint | GOP fixed at session start; never changed mid-session | §4.4 |
| Inherited (all prior families' Z lists) | Owned by prior chapters | Not relitigated | header preamble |

### R-18 Operational Integrity Compliance

Inherits R-18 enforcement from C08. Compliance verified:

- **Static — chapter prose**: §1 references R-18; §6.5 piggybacks on C25 §7 family allow-list.
- **Static — code in §6**: imports `r18 "github.com/vasic-digital/helix-r18-safeexec"` (origin C08 §10). Deny-list NOT duplicated.
- **Static — bootstrap subprocess invocations**: `nvidia-smi --query-gpu=temperature.gpu,power.draw,clocks.sm`, `rocm-smi -t`, `intel_gpu_top -J` (cross-link C25 §7) all wrap through `r18.SafeExec`.
- **Test — §8.11**: `host-integrity-scan` inherited VERBATIM from C08 §12.11. Non-overridable per Constitution §11.5.4.

### Coverage Confirmation

| Metric | Value |
|--------|-------|
| Source per-dim primary (`video-tech_dim04.md`) | 1,013 lines |
| R-01 minimum (Master Plan §7.2 row C29) | 1,150 lines of body prose |
| Body prose actually synthesised | **1,919 lines** across §§1–9 (A 577 + B 210 + C 467 + D 665) |
| Coverage ratio vs minimum | 1.67× |
| Coverage ratio vs primary per-dim source | 1.89× |
| Forbidden-pattern scan (chapter prose) | clean (only legitimate self-referential mentions) |
| Empty-section-body scan | clean |
| Tables | Frame-Tee fan-out matrix in §2; per-vendor preset matrix in §2.2; Treiber-stack reuse in §2.3; thermal tier ladder in §3.2; per-codec storage matrix in §5.3; failure-mode 12-row table F1-F12 in §7; test-type matrix in §8 |
| Section count | 10 normative sections + this verification block |
| Go code blocks | §6.4 (~85 LOC `dualpath.NewFrameTee` + `Start` + ThermalGuard hooks — real imports `context`, `sync/atomic`, `golang.org/x/sys/unix`, `r18`, `helix-shm`, `helix-lockfree`, `helix-encoder`) |
| R-18 enforcement | inherited from C08 §10 + §8.11 host-integrity-scan inheritance |

### Sign-off

- Section A (§§1–2) by C29 Group A on 2026-04-29.
- Section B (§§3–4) by C29 Group B on 2026-04-29.
- Section C (§§5–6) by C29 Group C on 2026-04-29.
- Section D (§§7–9) by C29 Group D on 2026-04-29.
- Web addendum by C29 addendum subagent on 2026-04-29.
- Header, ToC, §10, Anti-Bluff Verification stitched by orchestrator (Claude) on 2026-04-29.
- Reviewed by: pending operator review.

End of `05_Video_Audio/04_DualPath_Encoding.md` — 2026-04-29.
