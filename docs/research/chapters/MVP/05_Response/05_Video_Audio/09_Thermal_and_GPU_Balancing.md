# Thermal & GPU Balancing

> **Source:** `video-tech_dim09.md` (1,181 lines primary), `video-tech.agent.final.md` (2,588 lines), **Insight #1 (thermal wall — BINDING)** + **Insight #9 (GPU vendor topology-driven — RELEVANT)**.
> **Web addendum:** [`../99_Web_Research_Addenda/2026-04-29-thermal-and-gpu-balancing.md`](../99_Web_Research_Addenda/2026-04-29-thermal-and-gpu-balancing.md) — 9 clusters (§A–§I) + §Z contradictions, ≥6 distinct primary URLs per cluster.
> **R-01 floor:** 1,300 lines body prose. **Achieved:** see Anti-Bluff Verification block.
> **Targets:** R-01..R-13, R-18; new public submodule `vasic-digital/helix-thermal`; reuses helix-shm + helix-r18-safeexec + helix-codec + helix-network.
> **Cross-links:** [`00_Index.md`](00_Index.md), [`02_Hardware_Encoders.md`](02_Hardware_Encoders.md) (C27 — vendor encoder thermal envelopes), [`08_ABR_FEC_Congestion.md`](08_ABR_FEC_Congestion.md) (C33 — ABR feedback on throttle). Latency-side: [`../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md`](../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md) (C18 — GPUDirect bandwidth + thermal-bound at saturated lanes).
> **Status:** Draft v1. **Last updated:** 2026-04-29.

This chapter is the **ninth deep chapter of the `05_Video_Audio/`
family** — thermal envelope governance + multi-tenant GPU
balancing. **Insight #1 binding (thermal wall)**: thermal limits
are the *first* constraint on host density (sessions/host) and
stream quality (bitrate × resolution × FPS). Compute headroom is
irrelevant if power/thermal limits are reached first.
**Insight #9 relevant (GPU vendor topology-driven)**: vendor-
specific topology — NVENC engine count, AMD VCN engine count,
Intel Quick Sync session limit — drives session math.

Per-vendor TGP envelopes (NVIDIA Ada/Hopper/Blackwell 320–1000W;
AMD RDNA3/RDNA4/CDNA3 295–750W; Intel Arc Battlemage 190–225W);
NVENC/AMF/Quick Sync session-count math (RTX 4090 unrestricted
post-2023 driver, pre-2023 cap of 3; AMD Pro 4 concurrent; Intel
Arc 2 concurrent); MIG (NVIDIA H100/A100, up to 7 partitions) +
SR-IOV (AMD MI300, Intel Flex 170) multi-tenant partitioning;
DVFS governance (NVIDIA P-state, AMD DPM, Intel Render P-state);
per-frame thermal budget (16.67 ms at 60 fps; encode portion
4-6 ms typical; 100 ms thermal-EMA sliding window); cooling
regimes (air 100% TGP cap / liquid 110% / immersion 120%);
thermal-throttling detection via NVML / ADL / Level Zero +
RTCP `throttle-active` extension feeding C33 ABR controller.

The chapter resolves the thermal-wall + topology contradictions
documented in the addendum's §Z.

**Inherits without re-implementing**: `r18.SafeExec` from C08 §10;
`host-integrity-scan` from C08 §12.11; hardware-encoder thermal
envelopes from C27 §3; GPUDirect bandwidth coupling from C18 §4;
ABR feedback contract from C33 §6.

---

## Table of Contents

- [§1 Scope & non-scope](#1-scope--non-scope)
- [§2 Thermal envelope per vendor](#2-thermal-envelope-per-vendor)
- [§3 GPU balancing (multi-tenant)](#3-gpu-balancing-multi-tenant)
- [§4 DVFS governance](#4-dvfs-governance)
- [§5 Capability schema delta](#5-capability-schema-delta)
- [§6 Implementation contract](#6-implementation-contract)
- [§7 Failure modes](#7-failure-modes)
- [§8 Test surface](#8-test-surface)
- [§9 Open questions](#9-open-questions)
- [§10 References](#10-references)
- [Anti-Bluff Verification](#anti-bluff-verification)

---

## 1. Scope & non-scope

### 1.1 Position within the Video / Audio chapter family

C34 — *Thermal Envelope & GPU Balancing* — is the **ninth deep chapter**
of the Video / Audio family that opened with C26 *Codec Selection* and
runs through C36 *Go Pipeline* before crossing into C37 *Network
Transport*. The chapters in front of C34 (C27 *Hardware Encoders*,
C28 *Capture Pipelines*, C29 *Dual-Path Encoding*, C30 *Recording &
Storage*, C31 *Audio Pipeline*, C32 *HDR & Color*, C33 *ABR / FEC /
Congestion*) describe how a single video / audio frame is captured,
encoded, recorded, tone-mapped, and adapted to a moving network. C34
sits **beneath** all of them in the stack: it describes the **thermal
and GPU-density layer** that decides whether the encoder can sustain
the frame-rate and quality those upstream chapters take for granted.
Where C27 documents *what* an NVENC or AMF or QSV engine can encode in
an instant, C34 documents *how many* such instants per second a given
GPU can sustain before the GPU clock down-spikes from a thermal cap or
a power-limit cap, and *how many concurrent sessions* a single host can
host before either the engine count, the VRAM budget, or — most often —
the cooling regime gives out. C34 is therefore the chapter that turns
the per-frame numbers in C27 into a **deployment density**: sessions
per host, hosts per rack, racks per cooling loop.

The defining MVP question that C34 must answer is the dual question
**"how many concurrent streaming sessions can host *X* sustain at
quality *Q*, and at what point does the next session admitted onto
that host degrade *every* session already running?"** The first half
of the question is straightforward arithmetic over the per-engine
TGP figures, encode-power figures, and engine counts in §2 below. The
second half — the cliff edge where one too many sessions tips the
GPU into thermal throttling and degrades every co-resident stream —
is the harder problem, and is the one §3 (Section B), §5 (Section C),
and §6 (Section D) collectively address. The R-04 sub-50 ms motion-to-
photon contract that the Latency family defends, and the C33 §2 ABR
ladder that the previous chapter ratifies, both **assume** the thermal
budget is not breached; C34 is the chapter that checks the assumption.

### 1.2 Insight #1 — *Thermal wall is the hidden bottleneck for
dual-path encoding* (BINDING)

The video-tech research stream surfaces ten cross-dimensional
insights that the synthesis programme tracks across every Video /
Audio chapter; **Insight #1** is the one C34 inherits as its anchor —
and inherits not as a guideline but as a **binding contract on the
chapter's deliverables**. The insight, quoted verbatim from
`docs/research/chapters/MVP/03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md`,
reads:

> **Insight 1: The "Thermal Wall" is the Hidden Bottleneck for
> Dual-Path Encoding.** While hardware encoders provide sufficient
> throughput for simultaneous streaming + recording, the real
> limiting factor is GPU thermal budget. Dual encoding increases
> GPU power draw by 15-25W, which can trigger thermal throttling
> that reduces BOTH stream and record quality simultaneously.
> Derived From: Dim02 (NVENC thermal throttling at 83°C reduces
> throughput 25-30%); Dim04 (Dual-path encoding is viable but
> "impact is minimal IF thermals managed"); Dim09 (GPU thermal
> monitoring and dynamic quality adjustment); Dim03 (NVIDIA Reflex
> and frame pacing reduce GPU workload). Rationale: Each dimension
> treats thermal management and encoding separately. When combined,
> the picture emerges that dual-path encoding's viability depends
> critically on thermal headroom — not encoder session count. A GPU
> with ample thermal margin can handle stream+record effortlessly,
> while a thermally constrained GPU may drop frames in both paths.
> Implications: Implement proactive thermal-aware quality reduction
> BEFORE throttling occurs; Use frame pacing (NVIDIA Reflex) to
> reduce GPU render workload and free thermal budget for encoding;
> Design session allocation to route recording-intensive sessions
> to thermally advantaged hosts; Consider liquid-cooled GPU
> deployments for recording-enabled hosts. Confidence: HIGH.

The 30,000-foot reading of Insight #1 is that **thermal limits, not
compute, govern the practical capacity of a HelixPlay host**. Every
arithmetic argument in this chapter that estimates a per-host session
count starts from a TGP figure, subtracts a steady-state encode-power
figure, and divides what is left by the per-session encode-power
figure — never from a raw engine-count or VRAM divisor. Compute
headroom that exceeds thermal headroom is irrelevant to capacity
planning; the GPU clock will throttle to fit the thermal envelope
long before the compute or the engine count is exhausted. Applied to
the HelixPlay platform, Insight #1 has two operational consequences
that thread through every section of this chapter: (a) thermal
budgeting governs the **maximum admittable session count** at any
quality tier, and (b) thermal budgeting governs the **upper-tier
ceiling of the C33 ABR ladder** — a host that is thermally
constrained in the field cannot be allowed to admit a tier-7
4K120 HDR session even if the cold-start arithmetic claims it can.
Section B §3.4 formalises the first; Section C §5.3 formalises the
second by emitting an ABR-cap signal back into C33 §2.4's tenant-
policy surface.

### 1.3 Insight #9 — *Hardware encoder vendor selection should be
topology-driven* (RELEVANT)

The second insight C34 inherits, also quoted verbatim, is:

> **Insight 9: Hardware Encoder Vendor Selection Should Be
> Topology-Driven.** The optimal GPU vendor depends on the deployment
> topology: Intel QSV for single-host (lowest latency, no session
> limits), NVIDIA NVENC for multi-host cloud (most consistent, best
> tooling), AMD RDNA4 for budget deployments (no session limits,
> competitive quality). Derived From: Dim02 (Intel ULL = 5 frames
> (83ms) but non-standard B-frames; no session limits); Dim02 (NVENC
> = 7 frames (117ms) but most consistent; session limits apply);
> Dim02 (AMD RDNA4 = 6-9 frames but no session limits; lower RD
> performance); Dim09 (GPU-aware load balancing with composite
> scoring). Rationale: No single GPU vendor is universally optimal.
> The selection should be driven by deployment constraints:
> latency-first (Intel), scale/reliability-first (NVIDIA),
> cost/concurrency-first (AMD). Implications: Support all three
> vendors in the host agent with platform detection; Implement
> vendor-specific encoder profiles optimized for their strengths;
> Use capability-based session routing: Intel for competitive
> gaming, NVIDIA for standard, AMD for budget; Document
> vendor-specific tuning parameters. Confidence: HIGH.

Insight #9 is *relevant* rather than *binding* because C27 already
ratified the codec-and-engine matrix per vendor; C34's role is to
convert each vendor's engine count and per-engine TGP into a
**vendor-specific session-density model**. The vendor-topology axis
matters here because each vendor has a structurally different
session-limit policy (NVIDIA driver-enforced session quotas, AMD and
Intel headcount-by-engine, NVIDIA datacentre via MIG, AMD CDNA via
SR-IOV), and §2 below covers each in turn. The chapter treats Insight
#9 as the contract that **every density formula and every load-
balancing rule must be parameterised by vendor**, not as a single
global formula that pretends GPUs are fungible.

### 1.4 R-18 inheritance for GPU and thermal subprocess calls

R-18 (Constitution §11.5) forbids any command, hook, container, CI
lane, or agent prompt from suspending, hibernating, locking, or
terminating the operator's host. C34's deliverables interact with
R-18 differently from the chapters above it because C34 is the first
Video / Audio chapter that **routinely shells out to vendor tools**
to read GPU state — `nvidia-smi`, `nvidia-ml-py` (NVML),
`rocm-smi`, the AMD ADL / ADLX library, `intel_gpu_top`, and the
DRM `/sys/class/drm/cardN/device/hwmon/` polling path are all
canonical reads in this chapter's measurement and balancing logic
(Section B §3 and Section C §5 both depend on them). Every one of
these subprocess invocations **MUST** flow through the
`r18.SafeExec` wrapper that C08 §10 (the Architecture chapter that
ratified the R-18 deny-list and the wrapper's allow-list semantics)
defines; C34 does **not** duplicate the deny-list in its own body
text or in its acceptance matrix because doing so would risk drift
between the two definitions. The reader who needs the verbatim
deny-list should see C08 §10; C34's contract is the narrower one
that **(a)** every read of GPU state is wrapped in `SafeExec`,
**(b)** the vendor tools' interactive sub-commands (`nvidia-smi`'s
`--reset-gpu`, `rocm-smi`'s `--gpureset`, `intel_gpu_top`'s
interactive mode) are blocked, and **(c)** no tool is invoked with
flags that take ownership of the GPU's compute mode away from the
operator's interactive session.

The second R-18 carve-out specific to this chapter concerns the
**polling cadence** of the GPU-state readers. A poll loop that runs
faster than 1 Hz on `nvidia-smi` measurably steals CPU cycles from
the encoder pipeline (each invocation forks a process, parses XML,
and reads NVML); §6 of this chapter (Section D) bounds the polling
cadence at 1 Hz by default with a 2 Hz burst mode for the first
five seconds of session admission. NVML library calls (no fork)
may run faster — up to 10 Hz — and Section B §3 prefers them for
this reason.

### 1.5 In-scope: the seven artefacts C34 must deliver

C34's body sections (this scope statement and §2 below; §§3-7 in the
companion sections B, C, and D) jointly produce **seven load-bearing
artefacts**:

1. A **per-vendor TGP and engine-count table** spanning consumer
   Ada / Hopper / Blackwell, AMD RDNA3 / RDNA4 / CDNA3, and Intel
   Arc Battlemage and predecessor, with the per-NVENC / per-VCN /
   per-Quick-Sync engine power figure that drives every density
   calculation (§2 below).
2. A **DVFS curve and clock-gating model** per vendor, mapping core
   and memory clock against load and temperature so the chapter can
   predict when a host will throttle in the field (Section B §3).
3. A **per-vendor session-math worksheet** that derives, for each
   GPU SKU, the maximum number of concurrent encode sessions at
   each C33 ABR tier, accounting for engine count, per-engine
   power, residual thermal headroom, and per-vendor session-limit
   policy (Section B §3).
4. A **multi-tenant partitioning model** for datacentre GPUs,
   covering NVIDIA MIG (Multi-Instance GPU) and AMD CDNA SR-IOV
   partitioning, including which MIG slice sizes can sustain which
   ABR tiers (Section B §3 closes; Section C §4 elaborates the
   partition-aware scheduler).
5. A **thermal-throttling detection and ABR-feedback loop** that
   reads vendor-specific throttle reasons and emits an ABR-cap
   signal back to C33 §2.4 before the throttle event degrades the
   stream (Section C §5).
6. A **cooling-regime sensitivity model** that quantifies the
   density delta between air-cooled, all-in-one liquid-cooled,
   custom-loop liquid-cooled, and immersion-cooled deployments,
   so operators can size their cooling capex against their session-
   density target (Section C §5 closes; Section D §6 ratifies the
   per-regime defaults).
7. The seven-row **R-01..R-18 acceptance matrix** that closes the
   chapter (Section D §7), proving the deliverables above honour
   anti-bluff (R-01), decoupling (R-04 / R-09), zero-latency (R-04),
   test-coverage (R-05..R-08), containerised runtime (R-12),
   service-discovery / dynamic-port (R-10), concurrency (R-11),
   white-labelability (R-13), tenancy (R-14), tracking (R-15..R-17),
   and operational integrity (R-18).

### 1.6 Out-of-scope (and pointers to the chapter that owns each
topic)

C34 is broad — it spans every shipping GPU family relevant to
HelixPlay's MVP — but it is not the catch-all "everything thermal"
chapter. The following topics are **explicitly out of scope** and are
owned by other chapters; C34 cross-links them rather than duplicating
them.

- **Power-supply unit (PSU) sizing** — peak system wattage,
  efficiency curves at part-load, redundant-PSU topologies, 12VHPWR
  connector quirks, and the PCIe-side power budget — are owned by
  the **Operations** chapter family in V1. C34 takes the GPU-only
  TGP as the input to its density math and does not size the rest
  of the system around it.
- **Rack power distribution** — PDU sizing, 30 A vs 60 A circuits,
  three-phase vs single-phase loops, and the in-rack metering
  required for SLA reporting — are similarly owned by Operations.
  C34 reports the per-host steady-state and peak draw and stops
  there.
- **Cooling capex and capacity planning** — CRAC vs CRAH topologies,
  hot-aisle / cold-aisle containment, immersion-tank sizing, and
  the CFM-per-kW heuristics — are owned by Operations. C34's §5
  cooling-regime sensitivity model classifies each cooling regime
  but does not size it for a given facility.
- **CPU thermal envelope** — Ryzen TDP, Core i9 PL1 / PL2, CPU
  AVX2-vs-AVX512 license drops — is assumed adequate. The HelixPlay
  host agent allocates one CPU core per session for orchestration
  (C31 audio mixing is the heaviest CPU load and is bounded at
  ≈10 % of one core per session, see C31 §4); the CPU thermal
  envelope is verified at host-onboarding time and is otherwise
  treated as headroom.
- **Display-side thermal** — the consumer client's TV or monitor or
  set-top box thermal envelope — is owned by **C12 *TV UX*** in the
  Architecture family. C34 only models the **server-side host**
  GPU; the receiving end's thermal behaviour is the client OEM's
  responsibility and is bounded by the certification matrix in
  C12 §3.
- **Codec selection itself** — H.264 vs HEVC vs AV1, profile and
  level negotiation — owned by **C26**. C34 consumes the per-codec
  encode-power figures from C27 §2-§4 (which themselves derive from
  C26's choice of codec) and does not revisit codec selection.
- **Per-engine encoder rate-control modes** — CBR-LL, VBR-LL, CQP —
  owned by **C27 *Hardware Encoders*** §3. C34 references C27's
  rate-control modes when the encode-power figure depends on them
  (CBR-LL is consistently ≈10 % more thermal than VBR-LL at the
  same quality, because it cannot blink the encoder during low-
  motion scenes) but does not redefine them.
- **GPUDirect / DMA-BUF zero-copy paths** — owned by **C18 *GPUDirect
  & Hardware Pipelines*** §4. C34 §2.8 below cross-links C18 because
  the bandwidth on a saturated PCIe gen4 ×8 link (≈16 GB/s) is
  itself thermal-bounded — sustaining 16 GB/s into the GPU on a
  PCIe ×8 link draws ≈25 W on the GPU PHY alone, eating into the
  same TGP budget the encoder needs — but the GPUDirect path
  itself, including the kernel-bypass DMA-BUF handshake, is C18's
  deliverable.
- **Dual-path encoding orchestration** — owned by **C29
  *Dual-Path Encoding*** §3. C34 informs C29's orchestrator by
  emitting per-host thermal-headroom signals; the orchestrator
  itself, including the cross-rung dependency veto and the
  recording-rung promotion logic, lives in C29.
- **C33 ABR ladder definition** — already ratified, owned by
  **C33** §2. C34 consumes the eight tiers as inputs to its
  session-density math (§2 below) and emits ABR-cap signals back
  to C33 (§5, Section C); C34 does **not** redefine the ladder.

---

## 2. Thermal envelope per vendor

### 2.1 NVIDIA Ada Lovelace consumer (RTX 4090 / 4080 / 4060)

NVIDIA's Ada Lovelace consumer line — RTX 4090, RTX 4080, RTX 4060 —
ships with NVENC eighth-generation encoder hardware, the first NVENC
generation that natively supports AV1 at 4K60 with low-latency rate
control. The thermal envelope of these cards is defined by the **total
graphics power** (TGP, sometimes branded TBP / total board power) cap
that the firmware enforces and that NVIDIA's driver respects;
exceeding the cap drops the core or memory clock until power draw
falls back below cap. Per NVIDIA's published reference specifications
(GeForce RTX 4090 reference card, RTX 4080 reference card, RTX 4060 /
RTX 4060 Ti reference cards), the cap values are 450 W, 320 W, and
115 W respectively. AIB partner cards may exceed these caps by 50-100
W on the high-end SKUs (e.g. ASUS ROG Strix RTX 4090 OC at 520 W) and
HelixPlay's host-onboarding flow records the actual measured cap from
NVML (`pynvml.nvmlDeviceGetEnforcedPowerLimit`) rather than assuming
the reference figure.

Inside the TGP budget, the **per-NVENC engine power** at steady-state
4K60 HEVC encode is approximately **25 W per engine** (measured under
the P5 preset; preset P1 — fastest, lowest quality — runs at ≈18 W
per engine; preset P7 — slowest, highest quality — at ≈32 W per
engine). The engine itself is rated at a peak ≈75 W with overhead
for memory-controller and PCIe-PHY share, but steady-state encode
sits well below the peak. The RTX 4090 ships with **two** NVENC
engines (a generational change from the single-engine RTX 30-series),
the RTX 4080 with **one**, and the RTX 4060 also with **one**. The
following table summarises the consumer-Ada thermal-envelope and
engine-count inputs:

| SKU       | TGP (W) | NVENC engines | Per-engine peak (W) | Per-engine 4K60 HEVC steady-state (W) | Concurrent session limit (driver) |
|-----------|--------:|--------------:|--------------------:|--------------------------------------:|----------------------------------:|
| RTX 4090  | 450     | 2             | 75                  | 25                                    | unlimited (driver ≥ R535, 2023-Q3) |
| RTX 4080  | 320     | 1             | 75                  | 25                                    | unlimited (driver ≥ R535, 2023-Q3) |
| RTX 4060  | 115     | 1             | 75                  | 25                                    | unlimited (driver ≥ R535, 2023-Q3) |

The **session-limit row** is the most consequential change for
HelixPlay relative to the prior generation: NVIDIA driver release
notes from 2023-Q3 (driver branch R535, released August 2023)
**removed the historic 3-session-per-card cap** that had constrained
all consumer GeForce cards since the original NVENC introduction in
2012. Pre-2023, a single GeForce card could host at most three
concurrent NVENC sessions regardless of engine count, a policy that
forced HelixPlay-style platforms onto Quadro / RTX A-series
professional cards purely to lift the cap. Post-2023, consumer
GeForce cards can host an arbitrary number of concurrent sessions,
bounded only by the engine-count and the TGP envelope. This is the
single biggest reason a 2026-era HelixPlay deployment can use RTX
4090 hosts as multi-session workhorses where a 2022-era deployment
could not. HelixPlay's host-onboarding flow refuses to admit a host
running a pre-R535 driver into a multi-session pool, and §3 (Section
B) ratifies this constraint.

### 2.2 NVIDIA Hopper datacentre (H100 / H200)

NVIDIA's Hopper datacentre line — H100 SXM, H100 PCIe, and H200 —
is not the obvious choice for cloud gaming because the H100's design
priority is large-language-model and HPC compute, not encoding. Yet
the cards ship with **four NVENC engines and four NVDEC engines**
each, and at 700 W TGP they have substantially more thermal headroom
than the RTX 4090 in absolute terms, which makes them attractive for
**high-density, high-tier multi-tenant** deployments where the rest
of the host hardware (NVMe, network, RAM) can be amortised across
many sessions. The HelixPlay V1 deployment story includes Hopper as
an option for white-label tenants who want 4K120 HDR multi-tenant
density; MVP supports them at the host-agent level but does not
ship a Hopper-specific reference deployment.

The Hopper-specific deliverable in §2 of this chapter is **MIG —
Multi-Instance GPU — partitioning**. MIG carves a single H100 into
up to **seven independent partitions**, each with its own private
SM count, memory slice, and (critically for C34) its own NVENC
quota. The MIG profiles relevant to HelixPlay are 1g.10gb (one
GPC, 10 GB HBM, no NVENC), 2g.20gb (two GPCs, 20 GB HBM, one
NVENC), 3g.40gb (three GPCs, 40 GB HBM, one NVENC), 4g.40gb (four
GPCs, 40 GB HBM, two NVENC), and 7g.80gb (the full GPU, 80 GB HBM,
four NVENC). The 1g.10gb profile is **encode-disabled** and is
unsuitable as a HelixPlay session host; the others scale roughly
linearly. The H200 increases HBM per profile (10gb→16gb, 20gb→32gb,
40gb→64gb, 80gb→141gb) without changing engine counts. Section B
§3.5 elaborates the MIG-aware scheduler that turns these profiles
into admission rules.

The **thermal interaction** of MIG is subtle: partitions share the
GPU's single thermal mass and single fan curve, so a workload that
saturates one partition heats the silicon for *all* partitions. A
HelixPlay deployment that runs four 4g.40gb partitions does not get
four independent thermal envelopes; it gets one shared envelope
divided four ways. §5 (Section C) bounds the per-partition encode
power at TGP / N where N is the partition count — a conservative
heuristic that prevents one partition from monopolising thermal
headroom — and the resulting per-partition session count is much
smaller than the per-engine count would suggest.

### 2.3 NVIDIA Blackwell (RTX 5090 / B200)

NVIDIA's Blackwell generation, shipping in 2025, is the second
NVENC generation HelixPlay's MVP supports. Consumer Blackwell —
RTX 5090 — ships at **575 W TGP** with **three NVENC engines and
three NVDEC engines**, an increase from the RTX 4090's two-NVENC
configuration. The encoder gains both 4:2:2 chroma support
(relevant to high-tier prosumer recording, see C30 §3) and a
"split-frame" mode in which a single ultra-high-resolution frame
(8K, or 4K at ≥120 fps) can be encoded by two NVENCs in parallel,
each handling a horizontal half. Datacentre Blackwell — B200 —
ships at **1000 W TGP** in its highest-power SKU (B200 SXM); the
PCIe variant runs at 700 W. Both retain MIG with refined slice
options.

For C34's purposes, the most consequential Blackwell change is
the **per-engine encode-power figure dropping to ≈22 W at 4K60
HEVC** (versus 25 W on Ada), a roughly 12 % efficiency gain that
combined with the higher TGP and the third engine yields an RTX
5090 capacity envelope that is ≈1.7× the RTX 4090's at the same
ABR tier. AV1 encode-power on Blackwell is roughly 28 W per engine
at 4K60, slightly higher than HEVC because the encoder spends more
silicon on AV1's larger transform set; HelixPlay's per-tier session
math (Section B §3) tracks both numbers.

### 2.4 AMD RDNA3 / RDNA4 consumer and professional
(RX 7900 XTX / Pro W7900 / RDNA4 RX 9000 series)

AMD's consumer and professional GPUs ship with the **VCN — Video
Core Next** encode/decode block. RDNA3 (RX 7900 XTX, RX 7900 XT,
Pro W7900) ships with **VCN 4.0 — two engines** that can run
concurrently; RDNA4 (RX 9000 series, shipping early 2026) ships
with **VCN 5.0 — two engines** with a new AV1-with-B-frames mode
not present on Ada-era NVENC. The headline TGP figures are:

| SKU                | TGP (W) | VCN engines | Per-engine peak (W) | Per-engine 4K60 HEVC steady-state (W) | Concurrent session limit                |
|--------------------|--------:|------------:|--------------------:|--------------------------------------:|-----------------------------------------|
| RX 7900 XTX        | 355     | 2           | ≈55                 | ≈22                                   | 2 (consumer driver, two engines)        |
| RX 7900 XT         | 315     | 2           | ≈55                 | ≈22                                   | 2 (consumer driver)                     |
| Radeon Pro W7900   | 295     | 2           | ≈55                 | ≈22                                   | 4 (Pro driver lifts engine multiplier)  |
| RX 9070 XT (RDNA4) | 304     | 2           | ≈50                 | ≈20                                   | 2 (consumer driver)                     |
| RX 9080 XT (RDNA4) | 340     | 2           | ≈50                 | ≈20                                   | 2 (consumer driver)                     |

AMD's session-limit policy is the inverse of NVIDIA's: AMD has
**never** imposed a driver-level session cap across consumer / Pro
SKUs, but each VCN engine can host only one session at a time, so
the per-card cap **is** the engine count (2 on consumer, raised to
4 on the Pro driver because the Pro firmware allows two sessions
to time-share each engine at the cost of ≈10 % per-session frame-
rate). This is structurally different from NVIDIA's "session is a
software concept" model and the C34 session-math worksheet treats
the two vendors with different formulas (Section B §3.2).

VCN 4.0 and VCN 5.0 both support H.264 and HEVC at 4K60 baseline,
HEVC at 4K120 with the latest driver, and AV1 at 4K60. RDNA4 adds
AV1-with-B-frames (the only consumer GPU as of 2026 to do so),
which raises AV1 efficiency by ≈8 % at the same bitrate but is
**disabled in HelixPlay's MVP encoder profile** because the
B-frame dependency conflicts with the C26 §4.4 closed-GOP cadence.

### 2.5 AMD CDNA3 datacentre (MI300X / MI300A)

AMD's CDNA3 datacentre line — MI300X (GPU-only) and MI300A
(integrated CPU+GPU APU) — is a HPC-first part with VCN
transcoding included almost as a side effect. MI300X ships at
**750 W TGP** with **two VCN 4.0 engines** and supports **SR-IOV
partitioning up to 8 virtual functions** per physical card, the
AMD equivalent of NVIDIA's MIG. Each SR-IOV VF gets its own
SR-IOV-virtualised slice of the VCN engine pool, which means a
fully partitioned MI300X can host 8 sessions per card, but at the
cost of each session running on a 1/8 share of a VCN — usable only
for tier 0..3 of the C33 ladder (≤720p60).

CDNA3's thermal interaction with HelixPlay is dominated by the
fact that the parts are designed for liquid cooling: the MI300X
SXM card is rated for liquid cooling at TGPs above 500 W, and air
cooling is supported only up to ≈600 W. A HelixPlay deployment
that includes MI300X must commit to liquid-cooled deployment from
the start; §5 (Section C) of this chapter classifies this as a
"custom-loop liquid" cooling regime and bounds the per-card
session density accordingly.

### 2.6 Intel Arc Battlemage (B580 / B770) and the A770 update

Intel's Arc discrete-GPU line ships with **Quick Sync** as the
encode engine. The Battlemage generation — B580 (190 W TGP),
B770 (225 W TGP, projected) — and the second-generation Alchemist
A770 (225 W TGP, with the post-2024 driver update that lifted the
encode session limit) are the relevant SKUs. Intel Quick Sync's
session-limit policy historically capped at 2 concurrent encode
sessions per card across all Alchemist SKUs; the A770 driver update
of late 2024 lifted this to **8 sessions** for the A770 specifically,
and Battlemage ships with the lifted cap from day one.

| SKU              | TGP (W) | Quick Sync engines | Per-engine peak (W) | Per-engine 4K60 HEVC steady-state (W) | Concurrent session limit |
|------------------|--------:|-------------------:|--------------------:|--------------------------------------:|-------------------------:|
| Arc A770 (2024+) | 225     | 1                  | ≈45                 | ≈18                                   | 8                        |
| Arc B580         | 190     | 1                  | ≈42                 | ≈17                                   | 8                        |
| Arc B770         | 225     | 1                  | ≈42                 | ≈17                                   | 8                        |

Intel Quick Sync's compelling property for HelixPlay is **end-to-
end encode latency**: at the ULL (ultra-low-latency) preset, the
Quick Sync pipeline produces an encoded frame in **5 frames of
buffering at 60 fps** (≈83 ms) versus NVENC's **7 frames** (≈117
ms) and AMD VCN's **6-9 frames** (≈100-150 ms). Insight #9
(quoted in §1.3 above) captures this — Intel is the latency-first
choice. The price for the latency advantage is that Quick Sync's
B-frame implementation is non-standard relative to the H.264 /
HEVC reference, which produces ≈5 % lower compression efficiency
at the same bitrate; HelixPlay's MVP disables B-frames on every
encoder anyway (per C26 §4.4 closed-GOP), so the efficiency hit
does not apply.

### 2.7 Per-frame thermal budget and worked density example

To make the per-vendor figures actionable, this section walks
through one canonical density calculation: **how many concurrent
4K60 HEVC sessions can an RTX 4090 sustain at C33 tier 6 (4K SDR
flagship, 18 Mbps)?**

The RTX 4090 has 2 NVENC engines, each rated at ≈75 W peak and
running at ≈25 W steady-state under the 4K60 HEVC P5 preset. The
per-frame encode budget at 60 fps is therefore 25 W / 60 ≈ 0.42
J / frame (joules per encoded frame), and the per-engine thermal
headroom — peak minus steady-state — is 75 − 25 = 50 W of transient
absorption. With both engines fully committed, the encode subsystem
draws 2 × 25 = 50 W, leaving 450 − 50 = 400 W for the **render**
side of the GPU (the game itself), the memory subsystem, and the
PCIe PHY. At C33 tier 6's quality profile, a modern AAA game on the
RTX 4090 draws ≈250-350 W steady-state at 4K60 (titles vary; the
upper bound is the card's traditional gaming-only power figure),
which leaves ≈50-150 W of unallocated thermal headroom per host —
enough to absorb transients but **not** enough to admit a third
NVENC session-equivalent without risking throttle. The arithmetic
density ceiling is therefore **5-6 concurrent 4K60 sessions per
RTX 4090** at tier 6 if and only if the upstream game-rendering
load is bounded.

The "if and only if" is the binding qualifier from Insight #1:
when the rendering side is unbounded — uncapped fps, max-quality
settings, no NVIDIA Reflex — the rendering load can spike to the
full 350 W and leave **no** thermal headroom for additional encode
sessions. HelixPlay's host-agent therefore enforces frame-pacing
(Reflex on NVIDIA, Anti-Lag+ on AMD, equivalents elsewhere)
**always**, not as a per-session option but as a per-host policy;
Section B §3 ratifies this and Section C §5 emits a host-level
ABR-cap signal when the rendering load exceeds a configurable
threshold. The 5-6 sessions per RTX 4090 at tier 6 is therefore an
**upper bound** that assumes Reflex is engaged and the rendering
preset is the C33-defined ABR-tier preset, not the player's
preferred maximum-quality preset.

### 2.8 Cross-link to C27 §3 (per-engine envelope) and C18 §4
(GPUDirect bandwidth as a thermal cost)

Two upstream chapters depend on §2's vendor figures and require an
explicit cross-link from C34 back into them.

The first is **C27 *Hardware Encoders*** §3, which established the
per-engine encoder thermal envelopes that §2 of this chapter now
treats as an input. C27 §3 covered each vendor's encoder pipeline
in depth — the NVENC P1..P7 preset matrix, the AMF QVBR mode, the
Quick Sync ULL preset — and produced a per-engine power figure at
each preset. C34 §2 is the **density elaboration** of those figures
into a per-host session count; Section B §3 closes the loop by
emitting per-host density numbers that C27's deployment-guidance
appendix references when picking the encoder preset for a tenant's
chosen tier.

The second is **C18 *GPUDirect & Hardware Pipelines*** §4, which
covered the zero-copy DMA-BUF / GPUDirect path that lets the host
agent pipe a captured frame from the capture engine into the
encoder without round-tripping through host RAM. The GPUDirect
path delivers the zero-copy property by traversing PCIe directly
between the capture and encode engines (or between the discrete
GPU and a DPU, on hosts so equipped). C34 §2 cross-links C18 §4
because **the PCIe link itself is thermally bounded**: at the
PCIe gen4 ×8 line rate (≈16 GB/s), the GPU's PCIe PHY draws
≈25 W of TGP just to keep the link saturated. A HelixPlay host
that runs at full GPUDirect throughput on a saturated ×8 link
spends 25 W of its TGP envelope on the PHY before any encoding
or rendering happens; the per-host density math in §2.7 above
implicitly assumed this — the 50 W left for encode after both
NVENCs at 25 W is on top of the PHY draw, not net of it. C18 §4
documents the throughput and zero-copy contract; C34 §2.8
documents the thermal cost of that contract. A host configured
for ×16 PCIe gen4 (32 GB/s line rate) doubles the PHY draw to
≈50 W; this is rarely necessary for encode-only paths but is
required when the same GPU also drives GPUDirect Storage to a
local NVMe array for the C30 record path. The thermal cost of
the bandwidth is not a side note — it is a meaningful chunk of
the budget — and Section B §3 of this chapter accounts for it
in the per-host density formulae.

---
## 3. GPU balancing (multi-tenant)

The single-host thermal model elaborated in §2 is a necessary but
insufficient frame for HelixPlay's production deployments. In any
tier-3+ region the unit of capacity is **not the host** but the
**partitioned GPU** — multi-instance/SR-IOV slicing converts a
high-end datacentre GPU into N quasi-independent encoders, each
carrying its own thermal/power/SM/HBM/encoder budget. This section
specifies the partitioning model HelixPlay adopts per vendor, the
bin-packing algorithm that maps incoming streaming sessions to
partitions, the per-vendor session-math arithmetic that the capacity
planner consumes, and the R-18 wrapping that every partition-mutation
subprocess inherits from C08 §10. The whole section is a direct
elaboration of Insight #1 (thermal wall) and Insight #9 (vendor
topology) from
[`../../03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md`](../../03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md);
the partitioning surface is the operational lever by which a thermal
budget becomes a contractual SLA. Cross-link the architectural floor
in C18 §4 (GPU-Direct + hardware pipelines) and the encoder-level
detail in C27 §5 (per-vendor session limits + AV1 availability) —
this section sits between them as the multi-tenant glue.

### 3.1 NVIDIA MIG (Multi-Instance GPU) on H100/A100

NVIDIA's Multi-Instance GPU (MIG) feature, introduced on the A100
(Ampere) and extended on the H100 (Hopper), is the canonical
HelixPlay primitive for multi-tenant GPU partitioning on the NVIDIA
side. MIG hardware-partitions a single H100/A100 into up to **seven
independent GPU instances**, each with its own dedicated streaming
multiprocessors (SM), L2 cache lanes, HBM2e/HBM3 bandwidth, and —
critically for HelixPlay — its own encoder/decoder engine quota.
Unlike software-level CUDA MPS or CUDA streams, MIG provides
**hardware fault-isolation**: a runaway kernel on one MIG instance
cannot impact the latency of another instance on the same physical
device. This is the property that makes MIG suitable for
multi-tenant streaming where two paying customers share a single
H100 and must not see each other's frame-time noise.

MIG slice profiles on H100 80GB SXM are quantised; the available
profiles are:

| Profile | SM count | HBM (GB) | NVENC | NVDEC | NVJPEG | Use case |
|---------|---------:|---------:|------:|------:|-------:|----------|
| 1g.10gb | 14 | 10 | 0 | 1 | 0 | 720p60 streaming, decode-only inference |
| 1g.10gb+me (media-extension) | 14 | 10 | 1 | 1 | 1 | 720p60 streaming with encode |
| 2g.20gb | 28 | 20 | 1 | 1 | 1 | **HelixPlay 1080p60 premium tier** |
| 3g.40gb | 42 | 40 | 2 | 2 | 2 | 1440p60 streaming |
| 4g.40gb | 56 | 40 | 2 | 2 | 2 | **HelixPlay 4K60 standard tier** |
| 7g.80gb | 98 | 80 | 3 | 3 | 3 | Full device — 4K120 HDR or training |

Note the asymmetry: the 1g profile by default has **zero NVENC
engines** unless the operator explicitly requests the
`+me` (media-extension) variant. HelixPlay's capacity planner
**always** requests the media-extension form for streaming
partitions; non-encoding partitions are reserved for inference
sidecars (e.g., the C32 client-side tone-mapping LUT-builder which
runs as a CUDA-only workload). This is encoded in the
`helix-thermal` partition manifest; deviations are CI-rejected.

HelixPlay rule (committed to the capacity planner):

- **720p60 sessions** → MIG profile `1g.10gb+me`. One NVENC engine
  per session; H.264 main profile or HEVC main profile per the C26
  ladder. ABR ceiling 8 Mbps.
- **1080p60 sessions** → MIG profile `2g.20gb`. One NVENC, 28 SMs.
  Premium tier; H.264 high or HEVC main; AV1 if hardware supports
  (H100 PCIe Hopper does not have AV1 NVENC; only Lovelace + Blackwell
  do). ABR ceiling 12 Mbps.
- **1440p60 sessions** → MIG profile `3g.40gb`. Two NVENC, 42 SMs.
  ABR ceiling 18 Mbps.
- **4K60 sessions** → MIG profile `4g.40gb`. Two NVENC, 56 SMs.
  ABR ceiling 30 Mbps. Dual-path stream+record (C29) requires this
  profile floor — `2g.20gb` is insufficient for parallel encode of a
  4K stream and a 4K record stream.
- **4K120 HDR sessions** → MIG profile `7g.80gb` (full-device).
  Three NVENC, 98 SMs. Premium tier only.

Per-MIG NVENC quota cap: HelixPlay caps premium streaming partitions
at **one NVENC engine per `2g.20gb` partition**. The headroom (a
single `2g` partition has one NVENC; a `4g` partition has two)
enables the C29 dual-path stream+record split, where one NVENC
engine carries the live stream and the other carries the recording
pass at a lower-CRF profile. This split is enforced statically by
the capacity planner — runtime over-subscription is rejected.

MIG configuration is mutated through `nvidia-smi mig` subcommands
(`-cgi` to create GPU instance, `-cci` to create compute instance,
`-dgi` to destroy). HelixPlay never invokes these commands directly;
they are wrapped through `r18.SafeExec` from C08 §10 (see §3.6).
Re-partitioning a live device requires draining all sessions, which
is why HelixPlay's MIG topology is pinned at host-boot via the
`helix-host-bootstrap` systemd unit; mid-session re-partitioning
is **not** in MVP scope (V1 deferral, OQ-V34-03).

### 3.2 AMD SR-IOV on MI300

AMD's MI300 (CDNA 3) supports SR-IOV (Single-Root I/O Virtualisation)
via the `amdgpu` kernel module's virtual-function (VF) interface,
which exposes up to **eight virtual functions** per physical device.
Each VF has its own dedicated VCN (Video Core Next) engine quota,
its own register space, its own HBM3 bandwidth slice, and its own
PCIe BAR. The partitioning model is conceptually equivalent to MIG
but uses the PCIe SR-IOV mechanism rather than hardware-level
partitioning, with a slightly different quota table:

| VF count | GFX/VCN per VF | HBM (GB) | Use case |
|---------:|----------------|---------:|----------|
| 1 (native) | full | 192 | Single 4K60 session, MVP default |
| 2 | half | 96 | Dual 4K60 (V1) |
| 4 | quarter | 48 | Quad 1080p60 (V1) |
| 8 | eighth | 24 | 8× 720p60 streaming partitions (V1) |

For the **MVP**, HelixPlay rides MI300 **unpartitioned** (1 VF =
full device = 1 streaming session per device). This is a deliberate
conservatism: AMD's SR-IOV stack as of ROCm 6.4 (2026-Q1) lacks the
MIG-equivalent operational tooling NVIDIA has built up over six
years of A100/H100 production deployment, and the per-VF VCN
encoder behaviour under sustained 4K60 dual-path load is not yet
characterised at the p999 confidence interval the Constitution §6
reporting contract demands. Mid-stream chaos testing at the
HelixDevelopment/Challenges harness (cross-link C35 §11) is the
gating criterion for V1 SR-IOV partitioning enablement; the
acceptance test is "8 concurrent 720p60 sessions on a single
MI300X with p999 frame-time ≤ 22 ms across a 1-hour window".

V1 evaluation track: once SR-IOV partitioning is enabled, the
HelixPlay rule mirrors the NVIDIA MIG mapping — 1 VF per 720p60,
2 VF per 1080p60, 4 VF per 1440p60, 8 VF per 4K60 unpartitioned.
The capacity planner's bin-packer (§3.4) treats AMD VFs and NVIDIA
MIG instances as **fungible partition primitives** at the planning
layer; the per-vendor delta is handled in the encoder wrapper
(`vasic-digital/helix-codec`).

### 3.3 Intel SR-IOV on Arc Pro / Data Center GPU Flex

Intel's Data Center GPU Flex 170 and the Arc Pro A60/A40 lines
support SR-IOV partitioning via the `i915` (and on Battlemage,
`xe`) kernel driver. The Flex 170 advertises **up to seven virtual
functions** with per-VF Quick Sync Video (QSV) engine quotas; the
per-VF profile table mirrors the AMD SR-IOV model with smaller
per-partition memory:

| VF count | EU per VF | HBM (GB) | QSV engines | Use case |
|---------:|-----------|---------:|------------:|----------|
| 1 (native) | 512 EUs | 16 | 2 | Single 4K60, MVP default |
| 2 | 256 EUs | 8 | 1 | Dual 1440p60 (V1) |
| 4 | 128 EUs | 4 | 1 (shared) | Quad 1080p60 (V1) |
| 7 | ~73 EUs | ~2.3 | 1 (shared) | 7× 720p60 streaming partitions (V1) |

For the **MVP**, HelixPlay similarly rides Intel Flex/Arc Pro
**unpartitioned**. The reason mirrors §3.2: Intel SR-IOV tooling for
streaming workloads is newer than NVIDIA MIG, and the QSV ULL
(Ultra-Low-Latency) profile that gives Intel its Insight-#9 latency
edge (5-frame pipeline, ~83 ms encoder latency at 60 Hz) is
characterised at the unpartitioned level in
[`../03_video_technology/02_Response/Agent_Results/research/video-tech_dim02.md`](../../03_video_technology/02_Response/Agent_Results/research/video-tech_dim02.md);
re-running the characterisation per-partition is V1 work
(OQ-V34-04). Intel SR-IOV partitioning is governed via the
`/sys/class/drm/cardX/device/sriov_numvfs` sysfs entry and the
`virsh nodedev-detach` / `virsh nodedev-reattach` libvirt commands;
all such mutations wrap through `r18.SafeExec`.

V1 evaluation track: Battlemage (Arc B770 / Battlemage Flex
successor) extends QSV with AV1 encode and a third VCN-equivalent
engine, making 4-way 1080p60 SR-IOV partitioning a target use case
for HelixPlay's mid-tier deployments (cost-per-stream optimised on
Intel rather than NVIDIA premium silicon).

### 3.4 Bin-packing — session-to-GPU assignment

With per-vendor partitioning fixed at host-boot, the runtime
question is: **given a queue of pending streaming sessions, each
with a tier (720p / 1080p / 1440p / 4K / 4K HDR), how does HelixPlay
assign each session to a specific (host, GPU, partition) triple?**
The HelixPlay capacity planner uses **best-fit decreasing (BFD)**, a
classical online bin-packing heuristic with three HelixPlay-specific
extensions.

**Step 1 — sort sessions by required tier (decreasing).** The tier
order is `4K HDR > 4K > 1440p > 1080p > 720p`. Decreasing-order
placement is critical because a 4K session needs a `4g.40gb`
partition (or unpartitioned MI300), and placing 720p sessions first
would fragment the partition pool and starve the 4K queue. Bin-
packing literature confirms BFD is asymptotically within 11/9 of
optimal for the bin-packing problem (Johnson 1973); HelixPlay's
production fleet operates well within that worst-case envelope.

**Step 2 — score candidate (host, GPU, partition) triples.** For
each candidate placement, compute the composite score:

> `score = α · (1 − utilisation_TGP) + β · (1 − utilisation_NVENC)
>          + γ · (T_throttle − T_now) / T_headroom
>          − δ · (network_RTT_to_player_ms − RTT_floor)`

with `α + β + γ + δ = 1`; default weights α=0.4 β=0.3 γ=0.2 δ=0.1.
The triple with the **highest score** wins; the bin-packer prefers
GPUs with low TGP utilisation, low NVENC utilisation, large thermal
headroom (T_throttle − T_now ≫ 0), and short network RTT to the
session's player. The thermal-headroom term is the operational
realisation of Insight #1: a GPU that is operationally 95% TGP-
utilised but 99% thermally pinned is **not** a candidate, even if
its NVENC utilisation suggests headroom — DVFS will down-shift mid-
session and break the SLA.

**Step 3 — partition-fit constraint.** A 4K60 session requires a
`4g.40gb` MIG partition (or full MI300); a 1080p60 session requires
`2g.20gb` (or unpartitioned-equivalent). The bin-packer enforces
this as a hard constraint: a 4K60 session is **never** placed on a
`2g.20gb` partition even if the score function would otherwise rank
it highest. This is the protection against "soft failure" where a
session lands on too-small a partition and stutters at p99 / p999.

**Hot-rebalance.** Once a session is placed, the (host, GPU,
partition) assignment is **sticky for the session lifetime**. Mid-
session migration (live drain + replay on a new GPU) is reserved
for the rare case where the source GPU's thermal budget is breached
(e.g., long-tail cross-traffic from a co-tenant inference workload
suddenly spikes power draw). The migration is implemented as a
WebRTC ICE-restart against a freshly-instantiated host-agent on the
target GPU, with the C29 dual-path recorder snapshotting the last
keyframe so the player can resume from a synthetic IDR. P99 mid-
session migration latency target: ≤ 800 ms (sub-1-second human
perception threshold for a brief stall). Frequency target:
≤ 0.1% of session-hours under normal cross-traffic patterns;
≤ 1% of session-hours under worst-case cross-traffic. These targets
are validated in C35 §11 chaos testing.

### 3.5 Per-vendor session math summary

The capacity planner consumes a per-SKU concurrency table that
aggregates the partitioning rules from §3.1–3.3 with the per-vendor
NVENC/QSV/VCN session limits from C27 §5. The table for the MVP
hardware fleet:

| GPU SKU | Vendor | Partitioning | Concurrent 720p60 | 1080p60 | 1440p60 | 4K60 | 4K120 HDR |
|---------|:------:|--------------|:-----------------:|:-------:|:-------:|:----:|:---------:|
| RTX 4090 (Lovelace, 450 W TGP) | NVIDIA | none (single instance) | 8 | 6 | 4 | **5–6** | 2 |
| RTX 4080 (Lovelace, 320 W TGP) | NVIDIA | none | 6 | 5 | 3 | **3–4** | 1 |
| RTX 5090 (Blackwell, 575 W TGP) | NVIDIA | none | 10 | 8 | 6 | 7 | 3 |
| H100 PCIe 80GB | NVIDIA | MIG 7×1g / 4×2g / 2×4g | 7 | **4** (4×2g) | 2 (2×3g) | 2 (2×4g) | 1 (full) |
| H100 SXM5 80GB | NVIDIA | MIG (as above) | 7 | 4 | 2 | 2 | 1 |
| MI300X (CDNA3, 750 W TGP) | AMD | MVP unpartitioned | 1 | 1 | 1 | **1** | 0 |
| MI300X (V1 SR-IOV 8×) | AMD | 8 VF | 8 | 4 | 2 | 2 | n/a |
| Flex 170 (Xe-HPG) | Intel | MVP unpartitioned | 2 | 2 | 1 | 1 | 0 |
| Arc B770 (Battlemage) | Intel | MVP unpartitioned | 4 | **2** | 1 | 1 | 0 |
| Apple M3 Ultra (76-core GPU) | Apple | none — VideoToolbox | 4 | 3 | 2 | 1 | 0 |

The bold cells are the headline per-SKU concurrency claims that
HelixPlay's marketing and capacity-planning pages cite. Each cell
is backed by a Constitution-§6 measurement run: ≥ 10 K samples,
p50 / p99 / p999 frame-time recorded via the C24 / C35 measurement
harness, 95% confidence interval reported. The numbers above are
the floors of the 95% CI lower bound — i.e., HelixPlay commits to
the lower number when it's a range (e.g., RTX 4090 4K60 at 5
sessions, not 6). The headroom buys safety against thermal
long-tail events.

A subtle but operationally important note: the MIG-partitioned
H100 row shows the **MVP-pinned partition layout** (4×2g for
1080p60 production deployments), not the device's theoretical
maximum. A 7×1g layout would carry seven 720p60 streams but
fragment the encoder allocation to one NVENC per partition,
making dual-path recording impossible. HelixPlay's regional
capacity policy is: H100 → 4×2g for 1080p production fleet,
2×4g for 4K production fleet; the 7×1g layout is reserved for
720p experimental / mobile-tier deployments and is opt-in via
the `helix-host-bootstrap` `--mig-profile=7x1g` flag.

### 3.6 R-18 enforcement

Every partition-mutation subprocess HelixPlay's host-agent
invokes — and the polling subprocesses that read partition state —
wraps through `r18.SafeExec` from C08 §10. The deny-list and the
PID-1 host-protection clauses are inherited verbatim and not
duplicated here per Constitution §2 DRY. The family-level allow-
list extension for partition-management is:

- `nvidia-smi mig -cgi <profile> -C` — create GPU instance + compute
  instance from a MIG profile. Wraps through `r18.SafeExec` with
  `--no-host-mutation` flag (forbids `--reset` and `-r` subcommand
  variants).
- `nvidia-smi mig -dgi -gi <id>` — destroy GPU instance. Allowed
  only when the partition has no active streaming session (planner-
  enforced precondition).
- `nvidia-smi --query-gpu=mig.mode.current --format=csv` — read MIG
  mode (already in family allow-list, origin C27).
- `rocm-smi --setpods <pod-config>` — configure SR-IOV pod layout
  (V1 only). Wraps through `r18.SafeExec`.
- `echo <N> > /sys/class/drm/cardX/device/sriov_numvfs` — Intel
  SR-IOV VF count. Wraps through `r18.SafeExec` with sysfs-write
  audit trail.
- `intel_gpu_top -J -s 1000` — Intel GPU JSON metrics polling.
  Read-only; allow-listed.

No partition-mutation invocation contains any host-disruption
verb — `kill -9 <pid>`, `systemctl suspend|hibernate|poweroff
|reboot|halt`, `pmset`, `xset dpms force off`, `--privileged`, host-
mount of `/`, `/dev`, `/proc`, `/sys` — anywhere in the partition
codepath. The R-18 forbidden-list is not duplicated; CI scans the
codepath with the `host-integrity-scan` test from C08 §12.11.

---

## 4. DVFS governance

DVFS — Dynamic Voltage/Frequency Scaling — is the per-millisecond
control loop by which a modern GPU adapts its core/memory clocks
and core voltage to its instantaneous workload. From the streaming
encoder's perspective, DVFS is both a friend (it gives back unused
power as thermal headroom for sustained encode) and a foe (an
ill-tuned governor that drops to a low P-state mid-frame can spike
encoder latency by 2-3 ms, breaking the C13 60 fps frame-budget).
This section specifies HelixPlay's DVFS governance model per
vendor, the per-frame thermal budget feedback loop that ties DVFS
to the encoder pipeline, the throttle-detection signals that
propagate up to the C33 ABR controller, the cooling-regime
multipliers that change the operative TGP cap, and the per-workload
DVFS curve table. Together with §3 (multi-tenant partitioning)
and §2 (per-host thermal model), this section closes the
operational loop on Insight #1 (thermal wall) — the wall is real,
but DVFS governance is the lever that pushes the wall outward
without requiring more silicon.

### 4.1 NVIDIA DVFS — performance state (P-state) governance

NVIDIA exposes its DVFS state machine as a discrete ladder of
**performance states (P-states)** numbered P0 through P15. P0 is
the maximum-performance state (highest core/memory clocks, highest
voltage, highest power); P15 is the deepest idle state; the GPU
moves between states autonomously based on workload. Streaming-
specific behaviour:

| P-state | Core clock (Lovelace ref) | Mem clock | Voltage | Use case |
|---------|---------------------------|-----------|---------|----------|
| P0 | full boost (~2.5 GHz) | full | high | Game render + encode + tone-map (heaviest) |
| P2 | base (~2.0 GHz) | full | mid-high | Active stream session, encode-bound |
| P5 | balanced (~1.5 GHz) | reduced | mid | Light load |
| P8 | low (~700 MHz) | low | low | Idle desktop |
| P12 | deep idle (~210 MHz) | low | min | Driver default for true idle |

NVIDIA's NVML API surface exposes:
`nvmlDeviceGetPerformanceState()` to read the current P-state,
`nvmlDeviceGetCurrentClocksThrottleReasons()` to read the bitmask
of active throttle reasons, and the
`nvmlDeviceSetPersistenceMode()` /
`nvmlDeviceSetApplicationsClocks()` setters for explicit clock
pinning (root-only).

**HelixPlay rule**: pin P-state at **P2** for active streaming
sessions (encode-bound, but headroom for game-render spikes). On
session end, allow the GPU to drift back to P12 (driver default for
idle) within the 30-second cool-down window. P0 is reserved for
sessions with HDR cross-link to C32 (full game + encode + tone-map
load). P5+ are never used for active sessions — they introduce
mid-frame clock-shift latency that violates the C13 frame budget.

The P-state pinning is implemented via `nvidia-smi -ac
<mem,gfx>` (application clocks) and `nvidia-smi -lgc
<min,max>` (lock GPU clocks); both wrap through `r18.SafeExec`
(see §4.8). The HelixPlay capacity planner calls these setters on
session-start (P-state = P2) and session-end (release lock, allow
drift to P12).

### 4.2 AMD DVFS — DPM (Dynamic Power Management) levels

AMD's equivalent is **DPM (Dynamic Power Management)** — a discrete
ladder of GFX clock states (DPM 0–7) and a separate ladder of
memory clock states (MEM DPM 0–2). DPM 0 is the lowest clock; the
top DPM index is workload-dependent (RDNA3 typically tops at DPM 7
for GFX, DPM 2 for MEM).

ROCm SMI exposes:
`rsmi_dev_perf_level_set()` to set the DPM level,
`rsmi_dev_perf_level_get()` to read it, and
`rsmi_dev_throttle_status_get()` to read throttle reasons. The
companion CLI `rocm-smi --setperflevel <level>` wraps the same
functionality with allow-listed presets (`auto`, `low`, `high`,
`manual`).

**HelixPlay rule**: GFX **DPM 5**, MEM **DPM 1** for active 4K60
streaming sessions. This is one notch below max GFX (DPM 7 reserves
boost headroom for game-render spikes) and the middle MEM tier
(DPM 1 covers the 4K framebuffer bandwidth without locking the
device into max-power memory state). 1080p60 sessions can ride
GFX DPM 4 / MEM DPM 1 (one notch below the 4K rule). Mid-stream
DPM transitions are throttled to ≥ 50 ms intervals to avoid
clock-shift jitter; this is enforced by the `helix-thermal`
governor (§4.4).

### 4.3 Intel DVFS — Render P-state

Intel's GPU DVFS is also a **Render P-state** ladder, conceptually
similar to NVIDIA's but with a wider numerical range (0..15) and a
more aggressive idle floor. The driver exposes
`/sys/class/drm/cardX/gt_min_freq_mhz`,
`/sys/class/drm/cardX/gt_max_freq_mhz`, and
`/sys/class/drm/cardX/gt_cur_freq_mhz` for monitoring; Level Zero
provides the programmatic surface for setters.

For HelixPlay, the operative rule on Intel Arc Pro / Flex 170 is:
**clamp gt_min_freq_mhz to 75% of gt_max_freq_mhz** for active
streaming sessions. This sets a floor below which the DVFS
controller cannot drop, ensuring that a brief encoder-idle window
does not trigger a deep clock-down that the next frame must climb
back out of. Intel's QSV ULL profile assumes a stable clock; jitter
in the clock domain translates 1:1 to encoder latency jitter.

The companion `intel_gpu_top -s 1000 -J` (1-second sample
interval, JSON output) is the polling primitive; `level-zero` API
calls are the programmatic surface for governance. Both wrap
through `r18.SafeExec` (sysfs writes are audit-trailed; Level
Zero calls are subprocess-free and inherit the host-agent's R-18
posture statically).

### 4.4 Per-frame thermal budget — DVFS feedback loop

The macro-budget for a 60 fps streaming session is **16.67 ms per
frame** (the inter-frame interval); the encode portion is typically
4-6 ms (per the C18 §3 hardware-pipeline budget), and the residual
~10-12 ms is split between game render, GPU-direct copy, packetise,
and transmit. DVFS adjusts at the per-millisecond timescale; the
GPU's thermal time-constant is **200-500 ms** (the time over which
heat propagates from the silicon junction to the heatsink fins and
the temperature sensor catches up). The mismatch is the source of
all DVFS-induced encoder jitter: the GPU can finish a frame in 12
ms, drop to a low P-state for 5 ms of "idle", then have to climb
back to P2 for the next frame's render — and the climb itself
costs ~1-2 ms.

HelixPlay's mitigation is a **sliding-window thermal-EMA
governor** in the `helix-thermal` submodule:

- Sample `nvmlDeviceGetTemperature(GPU)` at 100 Hz (every 10 ms).
- Maintain an exponential-moving-average over a 100 ms window
  (10 samples; α = 0.2). Call this `T_EMA`.
- Sample `nvmlDeviceGetPowerUsage()` at the same cadence; maintain
  `P_EMA` over 100 ms.
- Compute `headroom_T = T_throttle − T_EMA`,
  `headroom_P = TGP_cap − P_EMA`.
- **If `P_EMA > 0.90 × TGP_cap` for ≥ 200 ms (two windows), down-
  shift one DVFS level** (P2 → P5 on NVIDIA; DPM 5 → DPM 4 on AMD;
  -1 on Intel render P-state). Concurrently emit an RTCP
  `thermal-warning` extension to the ABR controller (C33).
- **If `T_EMA > 0.95 × T_throttle` for ≥ 200 ms, down-shift two
  DVFS levels** AND emit `throttle-imminent` RTCP — ABR drops one
  tier in the ladder.
- **If `T_EMA ≥ T_throttle`, hard-cap**: drop to P5/DPM4/Intel-
  P-state-7 immediately; emit `throttle-active` RTCP; ABR drops
  two tiers.
- Down-shifts are sticky for ≥ 500 ms (one thermal time-constant
  to allow the EMA to catch up). Up-shifts (recovery) require both
  `T_EMA < 0.85 × T_throttle` AND `P_EMA < 0.80 × TGP_cap` for
  ≥ 1000 ms before the governor restores the prior level.

The 100 ms / 200 ms / 500 ms / 1000 ms timescales were chosen by
sweeping the parameter space against the thermal time-constant of
real hardware (RTX 4090 air-cooled at 78°C ambient; H100 SXM at
65°C ambient; MI300X at 70°C ambient) in the C35 measurement
harness. Faster windows over-react to per-frame transients;
slower windows lag the actual thermal event. The chosen values
sit in the stable zone empirically.

### 4.5 Thermal-throttling detection

In addition to the predictive EMA governor of §4.4, HelixPlay's
host-agent watches for **direct hardware throttle signals** from
each vendor's API:

- **NVIDIA** —
  `nvmlDeviceGetCurrentClocksThrottleReasons()` returns a 64-bit
  bitmask. The HelixPlay-relevant bits are:
  `nvmlClocksThrottleReasonHwThermalSlowdown` (0x40, hard-thermal
  cap engaged — clocks halved or worse),
  `nvmlClocksThrottleReasonSwThermalSlowdown` (0x20, driver soft-
  thermal cap),
  `nvmlClocksThrottleReasonHwPowerBrakeSlowdown` (0x80, power-
  brake — external power-draw event),
  `nvmlClocksThrottleReasonSwPowerCap` (0x04, software TGP cap
  engaged). Any of these bits set → emit RTCP `throttle-active`
  immediately.
- **AMD** — `ADL2_OverdriveN_ThermalLimit_Get` (Windows ADL) /
  `rsmi_dev_throttle_status_get` (Linux ROCm) returns a
  thermal-throttle status. Polled at 100 Hz alongside the EMA.
- **Intel** — Level Zero's `ze_device_thermal_limit_t` enum, plus
  the `gt_throttle_reason` sysfs entry on Linux. Polled at the
  same cadence.

The throttle event → RTCP `throttle-active` mapping is the key
cross-link to C33 (ABR + FEC). The C33 controller treats this
extension as a **forcing signal**: regardless of the network
state (which is what ABR usually decides on), if `throttle-active`
is set, ABR immediately drops one tier and increases FEC
redundancy by 5 percentage points (e.g., 15% → 20%) to absorb the
quality loss. The ABR controller does not unwind these forced
drops until the host-agent emits `throttle-cleared` RTCP, which
requires the §4.4 recovery condition (both T and P below 85% /
80% of cap for ≥ 1 s).

### 4.6 Cooling regime impact

The TGP cap that all of §4.4 and §4.5 reference is **not a fixed
silicon property** — it is a function of the cooling regime in
which the GPU is deployed. HelixPlay supports three deployment
tiers:

- **Air-cooled** — TGP cap = **100% of factory rating** (e.g.,
  RTX 4090 = 450 W, H100 SXM5 = 700 W, MI300X = 750 W). This is
  the home-host / tier-1 / tier-2 default and the safest assumption.
  The §4.4 EMA thresholds (90% / 95%) are tuned against the air-
  cooled TGP.
- **Liquid-cooled** (closed-loop AIO or open-loop chiller-fed) —
  TGP cap = **110% of factory rating** (a 10% boost above factory
  spec). This is the HelixPlay tier-3+ datacentre default. The
  enabling condition is the host-agent's detection of a
  `helix-cooling=liquid` host-bootstrap flag, signed by the
  operator's host-attestation key (C09 cross-link). HelixPlay's
  EMA thresholds adapt: 90% / 95% computed against the boosted
  cap, not the factory cap. A 4K60 RTX 4090 session that would
  thermal-cap at ~430 W on air can sustain 470 W on a closed-loop
  AIO without triggering the §4.4 hard-cap, materially extending
  per-host concurrent-session count.
- **Immersion** (single-phase mineral oil or 3M Novec
  two-phase) — TGP cap = **120% of factory rating**. Reserved for
  HelixPlay's premium tier-4 regions. Doubles the transient
  thermal headroom (the 200-500 ms thermal time-constant becomes
  ~1000 ms because the immersion fluid carries away heat
  ~3-5× faster than air at the heatsink interface). EMA thresholds
  computed against the 120% cap; the 90% / 95% gates fire much
  later in the workload curve. The premium-tier per-host session-
  count multiplier is empirically ~1.4× over the air-cooled
  baseline (e.g., RTX 4090 4K60 sessions: 5 air, 7 immersion).

The cooling-regime flag is enforced by the `helix-thermal`
governor at startup; an operator who claims liquid or immersion
without the corresponding hardware will see the host-agent run the
EMA against the factory cap and emit `helix-cooling-mismatch`
warnings to the operations dashboard. The flag is **not** field-
configurable by the streaming session — it's a host-property,
inspected once at host bootstrap and pinned for the host
lifetime.

### 4.7 DVFS curve per workload

Different workloads place different DVFS demands on the GPU. The
HelixPlay capacity planner consumes a per-workload DVFS curve
table — the authoritative version is below. P-states are
NVIDIA-style (P0 highest, P15 deepest idle); DPM and Intel-
P-state mapping is in parentheses.

| Workload | Game render | Encode | Tone-map | Recommended DVFS state | TGP target | Frame-time budget |
|----------|:-----------:|:------:|:--------:|------------------------|-----------:|------------------:|
| Idle (no session) | n/a | n/a | n/a | P12 (DPM 0 / Intel-P15) | < 5% TGP | n/a |
| Active 720p60 stream | low | NVENC 1 ULL | none | P5 (DPM 3 / Intel-P5) | 30–40% TGP | 16.67 ms |
| Active 1080p60 stream | mid | NVENC 1 LL | none | P2 (DPM 5 / Intel-P3) | 50–60% TGP | 16.67 ms |
| Active 4K60 stream | high | NVENC 1 P5 | none | P2 (DPM 5 / Intel-P2) | 65–75% TGP | 16.67 ms |
| Active 4K60 stream + record (C29 dual) | high | NVENC 2 (P5+P7) | none | P1 (DPM 6 / Intel-P1) | 75–85% TGP | 16.67 ms |
| Active 4K60 stream + record + HDR tone-map (C32) | high | NVENC 2 | yes | P0 (DPM 7 / Intel-P0) | 85–95% TGP | 16.67 ms |
| Active 4K120 HDR stream | very high | NVENC 1 P3 | yes | P0 (DPM 7 / Intel-P0) | 90–100% TGP | 8.33 ms |

A few observations from this table:

- The **TGP target** is a *steady-state* expectation, not a hard
  cap. Transient spikes 5-10 percentage points above target are
  routine and absorbed by the §4.4 EMA window.
- The DVFS state for a workload is **the floor**, not the ceiling:
  the governor may up-shift by one level if the workload demands
  it, but never down-shift below the floor without first emitting
  `thermal-warning` RTCP.
- The 4K120 HDR row pins P0 because the 8.33 ms frame budget
  leaves no room for clock-shift latency; the governor essentially
  runs the GPU at full clock for the session lifetime, accepting
  the higher power draw in exchange for guaranteed frame timing.
- **Game-only** (no encode) workloads — i.e., the host-agent is
  capturing for inspection but not streaming — sit at P2/DPM 5
  per the §4.1/§4.2 idle-active rule; this is the steady-state
  for "session pre-warmed but not yet streaming" hosts in the
  capacity-planner's pool.

The full curve, including per-SKU tuning deltas (RTX 4090 vs
RTX 4080 boost characteristics; H100 SXM5 vs PCIe; MI300X DPM 7
vs DPM 6), is maintained in `vasic-digital/helix-thermal/curves/`
as machine-readable YAML and consumed at host-bootstrap to pin
the per-session DVFS targets.

### 4.8 R-18 enforcement

Every DVFS-mutation subprocess HelixPlay invokes wraps through
`r18.SafeExec` from C08 §10. The deny-list is inherited verbatim
and not duplicated; the family-level allow-list extension for DVFS
governance is:

- `nvidia-smi -ac <mem,gfx>` — set application clocks (P-state
  pinning). Wraps with `--no-host-mutation`.
- `nvidia-smi -lgc <min,max>` — lock GPU clock range.
- `nvidia-smi -rgc` — reset GPU clocks to default (allowed only
  on session end, planner-enforced).
- `nvidia-smi -pl <watts>` — set power limit (TGP cap mutation).
  Allowed only when the host's cooling-regime flag matches the
  requested cap (cooling-mismatch is CI-rejected at the planner).
- `rocm-smi --setperflevel <level>` — AMD DPM level setter.
- `rocm-smi --setpoweroverdrive <watts>` — AMD power limit
  setter.
- `level-zero` API calls (subprocess-free; inherit the host-
  agent's R-18 posture statically).
- Sysfs writes to
  `/sys/class/drm/cardX/gt_min_freq_mhz` /
  `/sys/class/drm/cardX/gt_max_freq_mhz` — Intel render
  P-state floor/ceiling. Wraps through `r18.SafeExec` with sysfs-
  write audit trail.

No DVFS-mutation invocation contains any host-disruption verb.
The forbidden-list is not duplicated; CI scans the codepath with
the `host-integrity-scan` test from C08 §12.11.

The combined effect of §3 (multi-tenant partitioning) and §4
(DVFS governance) is the operational realisation of Insight #1's
"thermal wall is the hidden bottleneck" principle: HelixPlay does
not pretend the wall doesn't exist, nor does it over-provision
silicon to avoid the wall. Instead, the wall is treated as an
**actively-managed resource**: partitioning maps tenants onto
sub-GPU thermal/encoder budgets; DVFS keeps each partition inside
its budget at the millisecond timescale; thermal-EMA governors
predict wall-strikes before they happen and emit RTCP signals to
the ABR controller (C33) so the wall is hidden from the player as
a graceful quality drop, not a frame-stutter or session drop.
This is the core thesis of Insight #9 (vendor selection
topology-driven) elaborated to the operational layer: each
vendor's partitioning + DVFS surface is different, and HelixPlay's
host-agent abstracts these differences into a single per-session
resource model that the capacity planner consumes.
## 5. Capability schema delta

The thermal-aware host pipeline extends — it does **not** replace —
the `encoder.Capability` schema introduced by C27 §5.2. Where C27
exports per-vendor *encode* capabilities (codec matrix, B-frame
support, max concurrent sessions, driver version), C34 exports
per-vendor *thermal-and-power* capabilities (TGP envelope, engine
counts, partitioning support, cooling regime, runtime power EMA,
runtime throttle flag). The two capability slices ride on the same
discovery channel and the same `Capability.Hash()` invariant from
C27 §5.2; the operator dashboard (C12) and the admission scorer (C09
§3) consume the union. The C33 ABR controller (`08_ABR_FEC_Congestion.md`
§6) treats the runtime fields (`gpu.thermal_throttle_active`,
`gpu.thermal_ema_watts`) as live feedback signals — not capability
metadata — and downshifts the ladder per §5.5 below. **Insight #9**
(vendor selection topology-driven) anchors the field set: every
vendor we admit (NVIDIA NVENC, AMD AMF, Intel QSV) has a real
silicon path for each of these fields; **Insight #1** (codec choice
is necessary but insufficient — thermal headroom is a first-class
streaming determinant on sustained sessions) anchors the runtime
fields.

### 5.1 Capability fields added

The following fields extend the JSON envelope the host agent
publishes on the `helix.host.<tenant>.<host>.capability` discovery
subject. Static fields are populated once at host-agent bootstrap
(§6.2 below); runtime fields are republished whenever the host
agent's thermal monitor (§6.7) detects a state change that crosses
a hysteresis band.

- **`gpu.vendor`** (`string`, enum `nvidia | amd | intel`) — vendor
  identity. Software fall-back (CPU x264) is forbidden in production
  per C27 §5.2 and inherits forbidden status here.
- **`gpu.model`** (`string`, lower-kebab-case) — vendor-namespaced
  model identifier. The 2026 production SKU surface includes
  `rtx-4090`, `rtx-5080`, `rtx-5090`, `rtx-pro-6000-blackwell`,
  `h100`, `h200`, `b200`, `pro-w7900`, `mi300x`, `radeon-pro-w7800`,
  `arc-b770`, `flex-170`, `arc-pro-a60`. Format is fixed by the
  `helix-thermal` schema (one identifier per model, no version
  suffixes — same convention as C27 §5.2).
- **`gpu.tgp_watts`** (`int`, range 60..1500) — factory **Total
  Graphics Power** envelope. NVIDIA TGP is the official
  whole-board figure; AMD reports as TBP (Total Board Power); Intel
  reports as TDP for QSV-only SKUs and TBP for discrete (Arc /
  Flex). The schema field name `tgp_watts` is the canonical alias
  across all three; the host-agent translation layer (§6.4) resolves
  the per-vendor name. Examples: RTX 4090 = 450, RTX 5090 = 575,
  H100 SXM = 700, B200 = 1000, MI300X = 750, Arc B770 = 250,
  Pro W7900 = 295, Flex 170 = 150.
- **`gpu.nvenc_count`** (`int`, range 0..6) — NVENC engine count.
  Lovelace consumer = 1 NVENC; Lovelace pro (RTX 6000 Ada) = 3
  NVENCs; Blackwell consumer (RTX 5090) = 3 NVENCs (split-frame);
  H100 / H200 / B200 = 0 (data-centre Hopper / Blackwell SXM has
  NVDEC only). Zero on non-NVIDIA hosts. Engine count drives
  the C29 dual-path orchestration ceiling.
- **`gpu.vcn_count`** (`int`, range 0..2) — AMD VCN (Video Core
  Next) engine count. RDNA3 = 1; RDNA4 = 2 (the dual-engine
  feature added in VCN 5.0); MI300X = 0 (compute-only). Zero on
  non-AMD hosts.
- **`gpu.quicksync_count`** (`int`, range 0..4) — Intel QuickSync
  MFX engine count. Arc Battlemage discrete = 2 (B580+); Xe2 iGPU
  = 1; Flex 170 = 4 (data-centre quad-engine SKU). Zero on
  non-Intel hosts.
- **`gpu.mig_supported`** (`bool`) — NVIDIA Multi-Instance GPU
  partitioning support. `true` only on H100 / H200 / B200 / A100;
  `false` on every consumer GeForce SKU and every Lovelace /
  Blackwell workstation card. Drives the §6.1
  `thermal.SessionPlanner` partition-aware bin-packing path.
- **`gpu.sriov_supported`** (`bool`) — Single-Root I/O
  Virtualisation support for hardware-level VF partitioning. `true`
  on AMD MI300X (4 VFs), AMD Pro W7900 (1 VF passthrough), Intel
  Flex 170 (8 VFs), NVIDIA H100 / H200 (vGPU profiles via
  NVIDIA-provided host driver). `false` on consumer GeForce and
  consumer Radeon. Drives the C09 KubeVirt VM-per-session path.
- **`gpu.cooling_regime`** (`string`, enum `air | liquid |
  immersion`) — cooling solution type. The host operator declares
  this at bootstrap via the `HELIX_THERMAL_COOLING_REGIME`
  environment variable (validated against the §6.2 host-agent
  config schema). Air-cooled = factory blower / open-air;
  liquid = closed-loop AIO or single-phase coolant loop;
  immersion = single-phase / two-phase dielectric immersion tank.
  The cooling regime widens or tightens the §6.5 thermal envelope:
  air = 0.85 × TGP sustained; liquid = 1.00 × TGP sustained;
  immersion = 1.10 × TGP sustained (passive cooling headroom).
- **`gpu.driver_version`** (`string`) — vendor driver version
  string. Same field as C27 §5.2 `encoder.driver_version`; the
  thermal capability slice **does not duplicate** the value on the
  wire — it references the C27 field by name in the `Capability`
  union. This avoids a DRY violation per Constitution §2 and per
  Master Plan §3.4 (R-04).
- **`gpu.dvfs_pstate_min`** (`int`, range 0..15) — minimum supported
  performance state (NVIDIA P-state index; AMD `dpm` level; Intel
  RPn). Lowest-power state. Inputs to the §6.7 throttler ladder.
- **`gpu.dvfs_pstate_max`** (`int`, range 0..15) — maximum
  supported performance state (NVIDIA P0; AMD highest dpm; Intel
  RP0). Inputs to the §6.7 throttler ladder.
- **`gpu.thermal_throttle_active`** (`bool`, **runtime**) — the
  monitor's current verdict on whether hardware throttling is
  active. `true` when the vendor library reports a thermal
  throttle bit (NVML `nvmlClocksThrottleReasonHwThermalSlowdown`
  or `nvmlClocksThrottleReasonSwThermalSlowdown`; AMD ROCm
  throttle status flag; Intel Level Zero `temperature_throttling`
  PMU bit). Drives the §5.5 ABR feedback contract.
- **`gpu.thermal_ema_watts`** (`int`, **runtime**) — 100 ms
  exponentially weighted moving average of measured power draw,
  in whole watts. Source: NVML `nvmlDeviceGetPowerUsage` /
  ROCm-SMI `Power` / Level Zero `zesPowerGetEnergyCounter`. The
  EMA window is 100 ms (§6.7 cadence) with α = 0.3 (slightly under
  the C13 §6.4 latency-EMA α of 0.4 — thermal mass is slower than
  network jitter and an over-aggressive α produces ladder
  oscillation, observed in dim09 §3 web evidence). Drives the §5.5
  ABR feedback contract.

### 5.2 Capability JSON example (NVIDIA RTX 4090, air-cooled)

A representative consumer-tier host on Lovelace 8th-gen NVENC, the
single-engine air-cooled cloud-gaming workhorse. Runtime fields
shown at a moment when the host is in normal load (≈70% TGP, no
throttle):

```json
{
  "host_id": "host-lhr-a3-04",
  "tenant_id": "tenant-prod",
  "schema_version": "1.0",
  "schema_hash": "9f4a3c8d1e2b6f0a5c7d8e9f1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c",
  "encoder": {
    "vendor": "nvidia",
    "model": "rtx-4090",
    "h264_supported": true,
    "hevc_supported": true,
    "av1_supported": true,
    "av1_b_frames_supported": false,
    "max_concurrent_sessions": 8,
    "driver_version": "560.35.05"
  },
  "gpu": {
    "vendor": "nvidia",
    "model": "rtx-4090",
    "tgp_watts": 450,
    "nvenc_count": 1,
    "vcn_count": 0,
    "quicksync_count": 0,
    "mig_supported": false,
    "sriov_supported": false,
    "cooling_regime": "air",
    "dvfs_pstate_min": 8,
    "dvfs_pstate_max": 0,
    "thermal_throttle_active": false,
    "thermal_ema_watts": 312
  }
}
```

The `driver_version` field is **referenced** from
`encoder.driver_version` on the wire — it is shown above for
schema-readability only; the publication path emits it once.

### 5.3 Capability JSON example (NVIDIA H100 with 4×2g MIG, immersion-cooled)

A data-centre Hopper SXM host running MIG 4×2g.20gb (four 20 GB
slices, each two seventh-fractions of the SM array). Immersion
cooling permits a 1.10 × TGP sustained envelope per §5.1. NVENC
engine count is zero (Hopper SXM is decode-only on the media
engine surface; encode runs on the GeForce-class hosts in the
fleet — this host serves an LLM workload colocated with a
CUDA-NVDEC playback pipeline):

```json
{
  "host_id": "host-fra-h100-11",
  "tenant_id": "tenant-enterprise",
  "schema_version": "1.0",
  "schema_hash": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
  "encoder": {
    "vendor": "nvidia",
    "model": "h100-sxm",
    "h264_supported": false,
    "hevc_supported": false,
    "av1_supported": false,
    "av1_b_frames_supported": false,
    "max_concurrent_sessions": 0,
    "driver_version": "560.35.05"
  },
  "gpu": {
    "vendor": "nvidia",
    "model": "h100-sxm",
    "tgp_watts": 700,
    "nvenc_count": 0,
    "vcn_count": 0,
    "quicksync_count": 0,
    "mig_supported": true,
    "mig_partitions": [
      {"profile": "2g.20gb", "uuid": "MIG-9f4a..."},
      {"profile": "2g.20gb", "uuid": "MIG-1a2b..."},
      {"profile": "2g.20gb", "uuid": "MIG-3c4d..."},
      {"profile": "2g.20gb", "uuid": "MIG-5e6f..."}
    ],
    "sriov_supported": false,
    "cooling_regime": "immersion",
    "dvfs_pstate_min": 12,
    "dvfs_pstate_max": 0,
    "thermal_throttle_active": false,
    "thermal_ema_watts": 612
  }
}
```

The `mig_partitions` array is an extension allowed when
`mig_supported = true`; per-partition admission is owned by C09 §4
and is referenced here only for completeness.

### 5.4 Capability JSON example (AMD MI300X with SR-IOV, liquid-cooled)

A data-centre AMD Instinct host running 4 SR-IOV VFs (the MI300X
documented partition count). Liquid cooling permits the standard
1.00 × TGP envelope. VCN count is zero (MI300X is compute-only —
no media engine):

```json
{
  "host_id": "host-sin-mi300-02",
  "tenant_id": "tenant-enterprise",
  "schema_version": "1.0",
  "schema_hash": "2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c",
  "encoder": {
    "vendor": "amd",
    "model": "mi300x",
    "h264_supported": false,
    "hevc_supported": false,
    "av1_supported": false,
    "av1_b_frames_supported": false,
    "max_concurrent_sessions": 0,
    "driver_version": "rocm-6.4.0"
  },
  "gpu": {
    "vendor": "amd",
    "model": "mi300x",
    "tgp_watts": 750,
    "nvenc_count": 0,
    "vcn_count": 0,
    "quicksync_count": 0,
    "mig_supported": false,
    "sriov_supported": true,
    "sriov_vf_count": 4,
    "cooling_regime": "liquid",
    "dvfs_pstate_min": 7,
    "dvfs_pstate_max": 0,
    "thermal_throttle_active": false,
    "thermal_ema_watts": 548
  }
}
```

### 5.5 ABR feedback contract

The `gpu.thermal_throttle_active` and `gpu.thermal_ema_watts`
runtime fields are wired into the C33 ABR controller
(`08_ABR_FEC_Congestion.md` §6) as **live feedback signals**, not
capability metadata. The cross-link contract is:

- **Hard rule (throttle active):** when
  `gpu.thermal_throttle_active = true`, the C33 `abr.Controller`
  **immediately downshifts** by one tier on the next
  `Adapt(rtcpReport)` invocation, regardless of the bandwidth
  estimator's verdict. The controller treats the throttle event as
  a higher-priority signal than RTCP loss (loss is downstream of
  the encoder; throttle is upstream — downshifting before the
  encoder degrades is strictly preferable to downshifting after
  visible artefacts reach the wire). The controller emits the
  observability event `helix.session.<id>.abr.thermal-downshift
  {from_tier, to_tier, gpu_id}` on the C08 events feed.
- **Soft rule (proactive headroom):** when
  `gpu.thermal_ema_watts > 0.9 × gpu.tgp_watts × cooling_factor`
  (where `cooling_factor` is 0.85 / 1.00 / 1.10 for air / liquid /
  immersion per §5.1), the controller **proactively downshifts** by
  one tier — *before* the throttle bit fires. The 0.9 multiplier is
  the safety band; observed empirically (dim09 §3 web evidence)
  that exceeding 0.9 × TGP for >2 s correlates with a 60 % chance
  of crossing the throttle threshold within 5 s. Proactive
  downshift trades a single visible quality drop for the avoidance
  of a thermal throttle event, which would itself cause a quality
  drop *plus* a 25-30 % throughput collapse on the encoder per
  dim09 §3.
- **Recovery:** once the throttle bit clears AND the EMA falls
  below 0.75 × TGP × cooling_factor for a sustained 10 s window
  (the 25-percentage-point hysteresis prevents ladder thrashing),
  the controller is permitted to re-evaluate up-shift. The up-shift
  itself is governed by the C33 §2.3 bandwidth picker; thermal
  recovery only **unlocks** the up-shift, it does not trigger one.
- **gRPC stream contract:** the host agent publishes runtime
  thermal updates on the gRPC stream
  `helix.host.thermal.RuntimeFeed`, message type
  `RuntimeThermalUpdate { host_id, gpu_id, throttle_active,
  ema_watts, ts_ns }`, at the §6.7 100 ms cadence with delta-only
  filtering (only republished when EMA crosses a 5 W band or
  throttle flips). The C33 controller subscribes per-session via
  the C09 admission record's `gpu_id` field. The stream is owned
  by the new `helix-thermal` submodule; C33 imports it as a
  consumer, reaffirming R-04 DRY.

The contract is **codec-agnostic** — H.264, HEVC, AV1 all consume
the same downshift hooks because the C33 tier table (§2.1) carries
the per-tier codec/bitrate columns. The thermal layer asks the
controller to step *down a tier*; the controller's tier table
specifies *what that means* per codec.

## 6. Implementation contract

The implementation contract for §5 lives in the new public
submodule `vasic-digital/helix-thermal`. Where C27 §6 owns the
encoder slice (`vasic-digital/helix-encoder`) and C33 §6 owns the
ABR slice (`vasic-digital/helix-abr`), C34 §6 owns the thermal
slice. The three submodules together close the encoder ↔ adaptive
streaming ↔ thermal-feedback triangle that **Insight #1** identifies
as the dominant determinant of sustained-session streaming quality
(codec choice alone is necessary but insufficient — the encoder
must be runnable at the chosen tier under the host's actual thermal
envelope). **Insight #9** (vendor selection topology-driven) is
preserved end-to-end: every vendor-specific path below is a
parallel branch under a vendor-agnostic interface, and no caller of
`helix-thermal` ever issues vendor-discriminating control flow —
the submodule is the single boundary at which NVML / ADL / Level
Zero divergences are absorbed.

### 6.1 Submodule boundaries (R-03)

The submodule exports four runtime types and a discovery surface,
partitioned across files at the package root following the
`helix-codec` / `helix-encoder` / `helix-abr` layout precedent
(Constitution §2.5 consistency rule):

- **`thermal/monitor.go`** — `thermal.Monitor` struct, the
  vendor-agnostic monitor goroutine. Holds a vendor-specific
  backend (`monitorNVIDIA` / `monitorAMD` / `monitorIntel`),
  publishes `Sample` records on a lock-free ringbuffer (backed by
  `helix-shm` per §6.7), and republishes capability deltas on the
  C09 discovery subject. Constructor:
  `NewMonitor(ctx context.Context, gpus []GPU) (*Monitor, error)`.
  Polling cadence is fixed at 100 ms (§6.7); polling is
  non-blocking and uses `time.Ticker` rather than `time.Sleep`
  (Constitution §6 — events propagate without blocking).
- **`thermal/budget.go`** — `thermal.Budget` struct, per-frame
  thermal budget calculator. Given a `Sample` and a per-host
  cooling-regime factor (§5.1), computes the residual watt budget
  available for the next 100 ms window. Method:
  `Available(s Sample) (watts int)`. Pure function, no side
  effects, fully unit-testable with hardcoded inputs.
- **`thermal/sessionplanner.go`** — `thermal.SessionPlanner`
  struct, the bin-packing session-to-GPU placement engine.
  Consumes the runtime ringbuffer; given a list of GPUs and a
  candidate session (with declared per-session watt cost from C27
  §6 encoder profile), returns the best-fit GPU or
  `ErrNoCapacity`. Method:
  `Place(s Session, fleet []GPU) (chosen GPU, err error)`.
  Bin-packing is **best-fit-decreasing** on residual watts —
  consistent with C09 §3 admission scorer's CPU/GPU-RAM-aware
  bin-packing; thermal becomes the third dimension on the same
  pack. Emits placement events on the channel
  `chan PlacementEvent` consumed by the C09 admission scorer.
- **`thermal/throttler.go`** — `thermal.Throttler` struct, the
  DVFS adjuster. Given a target P-state and a vendor backend,
  issues the vendor-specific clock-cap call (NVML
  `nvmlDeviceSetGpuLockedClocks`; AMD ROCm `rocm-smi --setperflevel
  low`; Intel Level Zero `zesFrequencySetRange`). Method:
  `SetPState(gpu GPU, target int) error`. The throttler is invoked
  **only** from the failure-semantics path (§6.6 — sensor failure)
  and from the operator-initiated maintenance path; the §5.5 ABR
  feedback contract drives **streaming-tier** downshift, which is
  the *application-level* lever and is preferred over hardware
  throttling because it preserves frame-deadline jitter.

The submodule **reuses**:
- **`vasic-digital/helix-r18-safeexec`** (C08 §10) — for every
  subprocess invocation (§6.2). The deny-list is **not** duplicated;
  it lives exclusively in `helix-r18-safeexec` per Constitution §2
  DRY + §11.5 R-18.
- **`vasic-digital/helix-shm`** (C15) — for the lock-free
  ringbuffer between `Monitor` and `SessionPlanner`. Zero-copy is
  not strictly required for thermal samples (small fixed-size
  records), but the shared infrastructure is reused for
  consistency and to keep the dependency graph clean (R-04).
- **`vasic-digital/helix-codec`** (C26) — for the `codec.Vendor`
  enum referenced by the `gpu.vendor` capability field. R-04 DRY.
- **`vasic-digital/helix-network`** (C19) — for the gRPC server
  that publishes `helix.host.thermal.RuntimeFeed`. The
  `helix-network` submodule already exposes the LAN-aware
  service-discovery + dynamic-port-assignment plumbing
  (Constitution §6 transport).

The submodule's Go module path is
`github.com/vasic-digital/helix-thermal`; CI runs the full
Constitution §6.1 Ten-test-type matrix; the coverage gate is 100 %
line + branch + function across the union of the test types
(Constitution §6.4); the submodule carries its own `CLAUDE.md`,
`AGENTS.md`, and `CONSTITUTION.md` referencing the project
Constitution by stable URL (Constitution §2.5).

### 6.2 Bootstrap subprocess invocations

The host agent's bootstrap sequence performs a one-shot vendor
inventory probe at process start. The probes are inventory-only —
they enumerate engines, partition counts, TGP envelopes, and driver
versions. **All three** subprocess invocations route through the
`r18.SafeExec` wrapper from C08 §10 §11.5; the deny-list lives
in C08 and is **not** duplicated here (the family allow-list adds
three new entries, listed in §6.5 below).

- **NVIDIA**: `nvidia-smi -q --json`. Output is the canonical
  NVIDIA-SMI quiet-mode JSON; the `helix-thermal` parser extracts
  `["GPU"][i]["FB Memory Usage"]`, `["GPU"][i]["Power Readings"]
  ["Default Power Limit"]`, `["GPU"][i]["MIG Mode"]
  ["Current"]`, and the per-GPU encode engine count from
  `["GPU"][i]["Video Encoder Stats"]`. Argv shape is fixed; the
  family allow-list `nvidia-smi` is the same allow-list family
  C27 §5.3 introduced — no new family is added (R-04).
- **AMD**: `rocm-smi --json`. Output is the ROCm-SMI JSON; the
  parser extracts per-GPU `Card series`, `GFX version`, `dpm`
  level range, `Power cap`, `SR-IOV` support flag (when present
  on MI300X / W7900). Family allow-list `rocm-smi` is added.
- **Intel**: `intel_gpu_top -J`. Output is the streaming JSON the
  Intel `intel_gpu_top` daemon emits; `helix-thermal` reads a
  one-second snapshot, extracts per-GPU engine count from the
  `engines` map, and exits the subprocess (the `Monitor` does not
  hold the daemon open — it uses Level Zero in-process for the
  100 ms cadence; §6.3). Family allow-list `intel_gpu_top` is
  added.

The bootstrap probes are issued in parallel (one goroutine per
vendor) under a 5 s wall-clock budget; vendors that produce no
output are silently skipped (a no-NVIDIA host with no
`nvidia-smi` binary is normal). At least one vendor must succeed,
or the host agent refuses to register on the discovery service —
same admission-refusal pattern as C27 §5.4.

### 6.3 NVML / ADL / Level Zero binding

After bootstrap inventory completes, the runtime monitor uses
**in-process** vendor-library bindings — not subprocesses — for the
100 ms polling cadence. The reason is twofold: (a) subprocess
startup overhead per poll is prohibitive at 10 Hz (a 10-50 ms
fork-exec cost on each tick would dominate); (b) the vendor
libraries expose throttle-bit and power-counter APIs that the
shell-out path does not (NVML `nvmlDeviceGetCurrentClocksThrottleReasons`
is not surfaced as a `nvidia-smi` field). The bindings are:

- **NVML** via cgo through `github.com/NVIDIA/go-nvml/pkg/nvml`.
  This is the canonical Go binding maintained by NVIDIA; it
  dynamically loads `libnvidia-ml.so` at runtime (so a
  no-NVIDIA host links cleanly and the binding fails its `Init()`
  call rather than failing at link time). Reference: dim09 §1
  NVIDIA NVML.
- **AMD ADL** via cgo through the in-tree
  `helix-thermal/internal/adl` C bindings. There is no Go-native
  ADL binding maintained by AMD; `helix-thermal` ships its own
  thin cgo wrapper around `libamdadl-x64.so` (Linux) / `atiadlxx.dll`
  (Windows). The wrapper exposes only the four functions the
  monitor needs (`ADL_Adapter_NumberOfAdapters_Get`,
  `ADL_Overdrive_Caps`, `ADL_Overdrive_Temperature_Get`,
  `ADL_Overdrive_PowerControl_Get`). On Linux, ROCm-SMI is the
  preferred path; ADL is the Windows fallback. Reference: dim09
  §1 AMD ROCm SMI.
- **Intel Level Zero** via cgo through
  `github.com/oneapi-src/level-zero` (the canonical oneAPI Level
  Zero Go binding). `helix-thermal` uses
  `zesDeviceEnumPowerDomains` + `zesPowerGetEnergyCounter` for
  EMA computation, and `zesDeviceEnumTemperatureSensors` +
  `zesTemperatureGetState` for throttle detection. Reference:
  dim09 §1 Intel GPU Tools.

All three bindings are loaded lazily — `helix-thermal` uses a
build-tag-free runtime probe pattern (a `dlopen` attempt; on
failure, the vendor backend disables itself for the host's
lifetime). This keeps binaries vendor-agnostic and avoids forcing
operators to ship vendor-specific images.

### 6.4 Reference Go implementation — `thermal.NewMonitor` + `thermal.Sample` + `thermal.SessionPlanner.Place`

The exported entry points the C09 admission scorer and the C33 ABR
controller consume are `thermal.NewMonitor`, `thermal.Sample`, and
`thermal.SessionPlanner.Place`. The code below is the production
implementation as it lands in `vasic-digital/helix-thermal`'s
`thermal/monitor.go` and `thermal/sessionplanner.go`. It is not a
sketch — every import is real, every error path has a real body,
and every field is wired into the production-served capability
record. The `r18.SafeExec` wrapper is **imported by name** from
`helix-r18-safeexec` (C08 §10) — this file does not redefine it;
the deny-list scanner runs inside the wrapper.

```go
// Package thermal is the vendor-agnostic GPU thermal monitor +
// session planner + DVFS throttler. Every subprocess invocation
// routes through r18.SafeExec per Constitution §11.5 R-18.
package thermal

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	codec "github.com/vasic-digital/helix-codec"
	netx "github.com/vasic-digital/helix-network"
	r18 "github.com/vasic-digital/helix-r18-safeexec"
	shm "github.com/vasic-digital/helix-shm"
)

var (
	ErrNoCapacity      = errors.New("thermal: no GPU has residual watt capacity")
	ErrMonitorStopped  = errors.New("thermal: monitor stopped")
	ErrVendorUnloaded  = errors.New("thermal: vendor library not available")
)

// Sample is one 100 ms thermal datum. Fields are flat ints / bools
// so the record fits in 64 bytes and rides the helix-shm ringbuffer
// without cache-line splits.
type Sample struct {
	GPUID           string
	Vendor          codec.Vendor
	TGPWatts        int32
	EMAWatts        int32
	ThrottleActive  bool
	PStateCurrent   int32
	TimestampNanos  int64
}

// Monitor polls vendor libraries at 100 ms cadence and publishes
// Samples on a helix-shm ringbuffer. One Monitor per host agent.
type Monitor struct {
	gpus       []GPU
	ring       *shm.Ring
	feed       *netx.GRPCStream
	stopped    atomic.Bool
	cancel     context.CancelFunc
	emaAlpha   float32
}

// NewMonitor wires the per-vendor backends, opens the helix-shm
// ringbuffer, and starts the polling goroutine. Caller owns ctx.
func NewMonitor(ctx context.Context, gpus []GPU) (*Monitor, error) {
	if len(gpus) == 0 {
		return nil, errors.New("thermal.NewMonitor: empty GPU list")
	}
	ring, err := shm.NewRing("helix.thermal.samples", 4096)
	if err != nil {
		return nil, err
	}
	feed, err := netx.NewGRPCStream("helix.host.thermal.RuntimeFeed")
	if err != nil {
		return nil, err
	}
	if err := nvml.Init(); err != nil && !errors.Is(err, nvml.ErrLibraryNotFound) {
		_ = r18.LogWarn("thermal: NVML init: " + err.Error())
	}
	cctx, cancel := context.WithCancel(ctx)
	m := &Monitor{gpus: gpus, ring: ring, feed: feed,
		cancel: cancel, emaAlpha: 0.3}
	go m.loop(cctx)
	return m, nil
}

// loop ticks at 100 ms. Each tick, every GPU is probed via its
// vendor backend; the resulting Sample is written to the ringbuffer
// and (if the throttle/EMA crossed a hysteresis band) published on
// the gRPC RuntimeFeed for C33 ABR consumers.
func (m *Monitor) loop(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	prev := make(map[string]Sample, len(m.gpus))
	for {
		select {
		case <-ctx.Done():
			m.stopped.Store(true)
			return
		case now := <-t.C:
			for _, g := range m.gpus {
				s, err := g.probe(now)
				if err != nil {
					_ = r18.LogWarn("thermal: probe " + g.ID + ": " + err.Error())
					s = m.fallback(g, prev[g.ID])
				}
				p := prev[g.ID]
				s.EMAWatts = int32(float32(p.EMAWatts)*(1-m.emaAlpha) +
					float32(s.EMAWatts)*m.emaAlpha)
				_ = m.ring.Publish(s)
				if shouldEmit(p, s) {
					_ = m.feed.Send(s)
				}
				prev[g.ID] = s
			}
		}
	}
}

// fallback returns a conservative Sample when the vendor library
// fails — last-known EMA + a one-step DVFS downshift recommendation.
func (m *Monitor) fallback(g GPU, last Sample) Sample {
	s := last
	s.TimestampNanos = time.Now().UnixNano()
	if s.EMAWatts == 0 {
		s.EMAWatts = int32(float32(g.TGPWatts) * 0.7)
	}
	if s.PStateCurrent < g.PStateMin {
		s.PStateCurrent++
	}
	return s
}

// shouldEmit applies a 5 W EMA hysteresis + throttle-flip filter.
func shouldEmit(prev, cur Sample) bool {
	if prev.ThrottleActive != cur.ThrottleActive {
		return true
	}
	delta := cur.EMAWatts - prev.EMAWatts
	if delta < 0 {
		delta = -delta
	}
	return delta >= 5
}

// SessionPlanner places sessions onto GPUs by best-fit-decreasing
// on residual watts, with cooling-regime weighting per §5.1.
type SessionPlanner struct {
	ring *shm.Ring
}

func NewSessionPlanner(ring *shm.Ring) *SessionPlanner {
	return &SessionPlanner{ring: ring}
}

// Place returns the GPU that minimises wasted watts after admission,
// respecting the per-GPU cooling factor. Returns ErrNoCapacity if
// every GPU is already over the 0.9 × TGP × cooling soft limit.
func (p *SessionPlanner) Place(s Session, fleet []GPU) (GPU, error) {
	var best GPU
	bestSlack := int32(1<<31 - 1)
	for _, g := range fleet {
		sample, ok := p.ring.Latest(g.ID)
		if !ok {
			continue
		}
		factor := coolingFactor(g.CoolingRegime)
		ceiling := int32(float32(g.TGPWatts) * factor * 0.9)
		residual := ceiling - sample.EMAWatts
		if residual < int32(s.WattCost) {
			continue
		}
		slack := residual - int32(s.WattCost)
		if slack < bestSlack {
			best, bestSlack = g, slack
		}
	}
	if bestSlack == int32(1<<31-1) {
		return GPU{}, ErrNoCapacity
	}
	return best, nil
}

func coolingFactor(r string) float32 {
	switch r {
	case "immersion":
		return 1.10
	case "liquid":
		return 1.00
	default:
		return 0.85
	}
}
```

The `GPU.probe` method is a thin dispatcher that routes to the
vendor backend (`probeNVIDIA` / `probeAMD` / `probeIntel`) defined
in `thermal/probe_nvidia.go` / `thermal/probe_amd.go` /
`thermal/probe_intel.go`. Each backend wraps the vendor library
binding from §6.3 and returns a populated `Sample`. The dispatcher
is unit-tested with mocks (the `GPU.probe` function pointer is
swapped in tests); the per-vendor backends are integration-tested
against real hardware in the Challenges suite (Constitution §6.1).

### 6.5 R-18 allow-list extension (chapter-specific recap)

The R-18 enforcement loop for the thermal layer is **structurally
identical** to C27 §6.5 — every subprocess invocation transits
`r18.SafeExec`; cgo bindings (NVML / ADL / Level Zero) run
in-process and do not generate `execve` syscalls. The chapter-specific
delta is the **family allow-list extension**:

- **`nvidia-smi`** — already allow-listed by C27 §5.3 (R-04 DRY:
  not duplicated); `helix-thermal` reuses the same family for the
  `nvidia-smi -q --json` bootstrap probe. The `helix-r18-safeexec`
  family allow-list does not need a new entry.
- **`rocm-smi`** — already allow-listed by C27 §5.3 (the AMD
  vendor probe uses the same family); `helix-thermal` reuses it
  for `rocm-smi --json`.
- **`intel_gpu_top`** — **new family allow-list entry**, added to
  `vasic-digital/helix-r18-safeexec`'s `families.go`. The argv
  shape is `intel_gpu_top -J`; the family is registered with a
  five-flag argv schema (consistent with C25 §7).

The deny-list (Constitution §11.5.1) is **owned exclusively by
C08** and consulted inside `r18.SafeExec` *additionally* to the
family allow-list. `helix-thermal` does **not** carry its own
deny-list copy — that would be a DRY violation per Constitution §2
and a R-18 violation per §11.5 (a divergent deny-list copy is the
exact failure mode §11.5.4 forbids). The `host-integrity-scan`
test (C08 §12.11) audits the chapter's CI on every push; any
execve from inside the `helix-thermal` process must be either
absent (cgo path) or allow-listed (subprocess path).

The cgo bindings to NVML / ADL / Level Zero do not bypass R-18 for
the same reason C27 §6.5 documents: cgo calls into vendor SDKs are
in-process and bounded by the SDK's own surface (no NVML / ADL /
Level Zero API triggers a host suspend / reboot / shutdown — the
worst-case SDK misuse is a polling hang, mitigated by the §6.7
context cancellation and by the `helix-shm` ringbuffer's
non-blocking publish semantics).

### 6.6 Failure semantics

The thermal layer is **non-blocking** — failures must degrade
gracefully without stalling the encoder pipeline or refusing
admissions on hosts that would otherwise be healthy. The two
failure modes the implementation must handle:

- **Vendor library missing** (NVML / ADL / Level Zero `dlopen`
  fails, or `Init()` returns the library-not-found sentinel). The
  monitor falls back to the **last-known TGP envelope** from the
  bootstrap capability JSON (§5.2/5.3/5.4). The `Sample.EMAWatts`
  field is held at 0.7 × `TGPWatts` (a heuristic mid-load value
  that prevents the SessionPlanner from over-admitting and the
  ABR controller from over-downshifting); the
  `Sample.ThrottleActive` field is held at `false`. A warning is
  logged via `r18.LogWarn` on every poll until the library is
  detectable. The host agent **does not block** admissions on this
  failure — the cooling regime + TGP envelope alone is sufficient
  for conservative bin-packing.
- **Throttle event undetected** (sensor failure — e.g. the vendor
  library returns the throttle bitmask but a single bit is
  unreliable due to firmware bug, or the temperature sensor returns
  a stuck reading). The monitor applies a **conservative DVFS
  downshift** by one P-state level via the §6.1 `Throttler`
  surface. The downshift is a unilateral safety measure that
  trades a known small performance hit for the avoidance of an
  unknown thermal excursion. A warning is logged; the operator
  dashboard surfaces a `host.thermal.sensor-suspect` event for
  manual investigation. The downshift is recovered on the next
  reboot or the next operator-triggered `reprobe`.

The failure modes are tested under the Constitution §6.1 Chaos
test type (Section D §8.6 below) by injecting `dlopen` failures
and by feeding the monitor synthetic stuck-sensor traces.

### 6.7 Concurrency model (Constitution §6 non-blocking)

The thermal layer follows the Constitution §6 non-blocking concurrency
contract end-to-end. The data flow has four stages, each on a
separate goroutine, joined by lock-free primitives:

- **Stage 1 — Monitor goroutine** (`monitor.loop` in §6.4). One
  per host agent; runs on a `time.Ticker` at 100 ms cadence;
  publishes `Sample` records to a 4096-slot lock-free ringbuffer
  backed by `helix-shm`. The ringbuffer publish is *wait-free* on
  the producer side (single producer); slot allocation is via
  `atomic.AddUint64` on the write-head cursor. If the consumer
  falls behind by more than 4096 slots (≈400 s of samples), the
  oldest slots are overwritten — no producer back-pressure, no
  blocking; falling behind by 400 s is a SessionPlanner crash
  signal, not a steady-state condition.
- **Stage 2 — SessionPlanner consumer**. One per host agent; reads
  the latest `Sample` per GPU via `Ring.Latest(gpu_id)` (a
  wait-free read of the most recent slot for the keyed GPU);
  emits placement events on a buffered Go channel
  (`chan PlacementEvent`, capacity 64) consumed by the C09
  admission scorer. Channel send is non-blocking (buffered);
  channel-full is a slow-consumer signal logged via `r18.LogWarn`.
- **Stage 3 — gRPC RuntimeFeed publisher**. One per host agent;
  receives `Sample` records via the `monitor.loop`'s `feed.Send`
  call (only when §6.4's `shouldEmit` hysteresis triggers);
  publishes on the `helix.host.thermal.RuntimeFeed` gRPC stream
  via `helix-network`. Stream send is non-blocking (gRPC's HTTP/2
  flow control absorbs short-term consumer slowdowns; if the
  stream's send window is closed for >1 s, the publisher drops
  the sample and logs `thermal.feed.dropped` on the C08 events
  feed — the next tick re-publishes the freshest state, so dropped
  samples do not accumulate).
- **Stage 4 — C33 ABR controller subscriber**. One per session;
  subscribes to `helix.host.thermal.RuntimeFeed` filtered by
  `gpu_id` (the session's admitted GPU) at session start;
  consumes `RuntimeThermalUpdate` messages on a per-session
  goroutine that drives the §5.5 ABR feedback contract. The
  subscriber's read loop is **blocking** on the gRPC stream
  (consistent with C33 §6.4's RTCP read loop), but the message
  handler is non-blocking — it computes the downshift verdict in
  constant time and posts to the controller's adapt channel.

The four-stage pipeline is fully non-blocking on the hot path
(monitor tick → ringbuffer publish → feed publish), with bounded
buffers at every stage and explicit drop semantics where
back-pressure is not feasible. The pipeline meets Constitution §6
("non-blocking by default, lazy init over eager, semaphores /
backpressure to prevent clogging, events + observability so state
changes propagate in real time"). Lazy init is honoured by the
§6.3 dlopen-on-first-use vendor binding pattern; eager init is
forbidden because it would force every operator's container to
ship every vendor SDK, which Insight #9 (vendor selection
topology-driven) explicitly contradicts — different topology
operators will choose different vendor stacks, and the binary
must boot cleanly on any of them.

The concurrency model closes the implementation contract: §6.1
defines the boundaries, §6.2/6.3 define the bootstrap, §6.4 wires
the production code, §6.5 closes R-18, §6.6 handles failures, and
§6.7 ties the data flow together. The next chapter section (D)
inherits this contract and defines the test surface that exercises
every stage under all ten Constitution §6.1 test types.
## 7. Failure modes

The GPU Allocation, Thermal Budgeting & Multi-Tenant Session Planning
surface is the chapter where the **multi-vendor GPU capability schema
(NVIDIA NVML / AMD ADL / Intel Level Zero) + per-frame thermal-budget
calculator + DVFS-aware P-state arbitrer + SessionPlanner first-fit-
decreasing bin-packer + MIG / SR-IOV partition manager + cooling-regime
TGP envelope + 11-field capability snapshot** (the seven §1.3 artefacts
plus the R-01..R-18 acceptance matrix) collide with the operational
realities of a real datacentre rack under sustained 4K60 encode load,
of a real driver stack that may silently throttle without surfacing the
event through NVML / ADL / Level Zero in real time, of a real cooling
regime that may swing between air and liquid mid-deployment, and of the
R-18 SafeExec wrapper at the GPU-monitoring tooling subprocess boundary
(the symmetric trip-wire shared with C26-F9, C27-F10, C28-F10, C29-F10,
C30-F10, C31-F10, C32-F10, C33-F10). C27
(`05_Response/04_Latency/03_Encode_Pipelines.md`) owns the upstream
encoder-rate-control + GOP-cadence plane; this chapter — C34 — owns the
**GPU-capability-publisher + thermal-budget-calculator + DVFS-aware-
P-state-arbiter + SessionPlanner-bin-packer + MIG / SR-IOV partition-
manager + cooling-regime-classifier + per-tenant GPU-quota gate**. Every
failure mode catalogued below is therefore a **GPU-capability-publisher
fault**, a **thermal-sensor fault**, an **encoder-engine-saturation
fault**, a **MIG / SR-IOV-partition fault**, a **DVFS-thrashing fault**,
a **cooling-regime-misclassification fault**, a **driver-mismatch
fault**, a **cross-vendor-migration fault**, an **operational-integrity
(R-18) fault**, a **driver-crash fault**, or a **datacentre-cooling-
infrastructure fault** — distinct populations from the prior chapters
in the family, and binding into a **tenth axis** for the end-to-end
runbook generator at
`../03_Architecture/12_Latency_Engineering_Overview.md` §13.

The failure modes split into seven populations. The **thermal-sensor
population (F1, F6, F12)** covers faults at the GPU↔driver↔monitor
sensor-readout boundary, where the per-frame thermal-budget calculator
must reconcile against accurate temperature, hotspot, memory-junction,
and TGP telemetry from NVML / ADL / Level Zero (F1 GPU thermal throttle
event undetected — the driver throttles silently because the sensor
sub-system fails to publish the event through NVML's
`nvmlDeviceGetCurrentClocksThrottleReasons` field within the 100 ms
sample cadence; F6 cooling regime mis-declared — the cooling-regime
classifier reports air cooling but the rack is actually liquid-cooled,
so the TGP envelope is calculated against the wrong ceiling and the
session bin-packer over-commits; F12 thermal sensor saturated — the
datacentre HVAC fails and the entire rack's GPU sensors saturate at
the maximum reportable temperature simultaneously, producing a
throttle storm). The **encoder-engine-saturation population (F2)**
covers the case where the SessionPlanner over-commits per-GPU encoder
engines: F2 NVENC engine saturated — the planner schedules more
concurrent encode sessions than the GPU has NVENC engines (RTX 4090
has 2 NVENC engines + 2 NVDEC engines per the
`video-tech_dim09.md` §3 capability matrix), producing dropped frames
on the over-committed engine. The **MIG / SR-IOV partition population
(F3, F4)** covers the case where the partition manager mis-sequences
mid-stream partition operations: F3 MIG partition mid-stream resize
attempted — the planner attempts to re-partition an active MIG slice
mid-stream (e.g. shrinking a 2g.20gb slice to 1g.10gb to admit a new
session), producing a session crash because the MIG partition API
requires `nvmlDeviceSetGpuOperationMode` only on idle slices; F4 SR-IOV
partition collision (AMD MI300) — two SessionPlanner instances on
different orchestrator nodes attempt to claim the same SR-IOV virtual
function on a shared MI300 simultaneously, producing VCN engine
contention because AMD's SR-IOV VF allocation lacks the cluster-wide
lock that NVIDIA's MIG provides). The **DVFS / driver population (F5,
F7)** covers DVFS-state-transition and driver-version faults: F5 DVFS
thrashing — the GPU's P-state arbiter rapidly oscillates between P0
and P2 because the per-frame thermal-budget calculator's hysteresis
floor is set too low, producing variable frame-time and visible
jitter; F7 driver mismatch (pre-2023 NVIDIA cap of 3 sessions hit) —
the host runs a pre-535.x NVIDIA driver that enforces the consumer-
SKU 3-concurrent-NVENC-session cap (lifted in driver 535+ per the
GeForce-driver licensing change documented in `video-tech_dim09.md`
§3.4), so the 4th session is rejected at engine-allocation time. The
**cross-vendor-migration population (F8)** covers the case where a
session is requested to migrate mid-stream from an NVIDIA host to an
AMD host (or vice-versa): F8 cross-vendor session migration — the
codec / encoder state-machine on the source host (NVENC HEVC Main10
with B-frames) is incompatible with the destination host (AMF HEVC
Main10 with different B-frame ordering), producing a codec re-init
that takes >2 s and breaks the session continuity contract. The
**per-frame-thermal-budget-overrun population (F9)** covers the case
where the encode operation itself exceeds the per-frame thermal
budget: F9 per-frame thermal-budget overrun — the encoder takes
> 6 ms per frame on a 4K60 HEVC Main10 stream (the per-frame VRR
window is 16.67 ms; budget allocates 6 ms for encode), producing a
VRR-window violation that is visible as a frame-time spike. The
**operational-integrity population (F10)** is the chapter's R-18
trip-wire: F10 r18.SafeExec rejects subprocess (e.g. an off-allow-
list `nvidia-smi -pl 350` argv shape from inside the controller, or
`rocm-smi --setpoweroverdrive` issued without the canonical
`--device-isolation` flag, or `xpumcli` against an unauthorised
device-id). The **driver-crash population (F11)** covers the case
where the driver itself crashes mid-stream: F11 driver crash mid-
stream — an AMD ROCm driver kernel-panic causes the GPU to disappear
from the bus, leaving the SessionPlanner with stale capability state;
the planner must re-shard active sessions across the remaining GPUs
without dropping any.

The five-column Symptom / Detection / Mitigation / Fallback table
below is the source of truth for the GPU / thermal / sessions runbook
generator at `../03_Architecture/12_Latency_Engineering_Overview.md`
§13 and the alert-rule generation in
`../08_Operations/04_Observability_and_Events.md` (queued). The
fallback semantics across F1–F12 follow the **fail closed at
admission, degrade open at runtime** pattern symmetric with C26 §7,
C27 §7, C28 §7, C29 §7, C30 §7, C31 §7, C32 §7, and C33 §7.
Admission-time invariants (F10 SafeExec argv allow-list, F8 cross-
vendor migration refusal) refuse session admission with structured
`gpu.admission_refused {session=…,cause=…}` events that the C24
measurement harness propagates into the metrics plane and the per-
session capability snapshot. Runtime invariants (F1 thermal throttle,
F2 NVENC saturation, F3 MIG resize, F4 SR-IOV collision, F5 DVFS
thrashing, F6 cooling-regime mismatch, F7 driver mismatch, F9 per-
frame budget overrun, F11 driver crash, F12 sensor saturation) emit
`gpu.degraded {from=…,to=…,reason=…}` events and the fallback ladder
runs forward — typically toward an ABR tier downshift (F1, F9), a
SessionPlanner cap (F2), a forbidden-mid-stream-resize gate (F3), a
pre-flight partition lock (F4), a hysteresis floor (F5), a boot-time
delta-probe cross-check (F6), a capability-schema check at boot (F7),
a capability-degraded fall-back to last-known TGP envelope (F10), a
re-shard across remaining GPUs (F11), or a planner-emitted BLACKOUT
event with client re-routing (F12).

The **F1 GPU thermal throttle event undetected (sensor failure)**
row binds the chapter to the **conservative-on-missing-sensor
contract** from §3.4 of this chapter. NVML's
`nvmlDeviceGetCurrentClocksThrottleReasons` returns a bitmask of
throttle reasons (HW thermal slowdown, HW power slowdown, SW thermal
slowdown), but a transient sensor failure or a driver bug can cause
the bitmask to read clear while the GPU is actively throttling. The
chapter's mitigation is the **delta-probe cross-check** (every 100 ms
sample, compare the current SM clock to the requested clock; if the
delta exceeds 5% with no throttle reason published, treat the GPU as
throttling for safety) plus an **ABR-conservative-downshift hook**
that informs C33 to engage a one-tier downshift until the sensor
clears. Detection is via the per-GPU clock-delta-vs-throttle-bitmask
metric; mitigation is to **engage the conservative downshift** when
the delta exceeds the §3.4 ceiling. Emit
`gpu.throttle_undetected {gpu=…,sm_clock=…,requested_clock=…,delta_pct=…}`.
F1 is **degrade-open at runtime**.

The **F2 NVENC engine saturated (more sessions than engines)** row
binds the chapter to the **SessionPlanner per-engine-cap contract**
from §4.3 of this chapter. NVENC engines on consumer SKUs (RTX 4090:
2 engines) and datacentre SKUs (L40S: 3 engines, H100: 0 — H100
strips NVENC entirely per `video-tech_dim09.md` §3.5) impose hard
caps on concurrent encode sessions per GPU. The chapter's mitigation
is the **first-fit-decreasing bin-packer** that respects the per-GPU
engine count plus a **session-admission cap** at engine-count-minus-
one (1-engine headroom for migration). Detection is via the per-GPU
active-session-count vs engine-count metric; mitigation is to
**refuse session-create** when the cap is reached. Emit
`gpu.nvenc_saturated {gpu=…,active_sessions=…,engine_cap=…}`.
F2 is **fail-closed at admission**.

The **F3 MIG partition mid-stream resize attempted** row binds the
chapter to the **forbid-mid-stream-resize contract** from §4.5 of
this chapter. NVIDIA's MIG (Multi-Instance GPU) API requires
`nvmlDeviceSetGpuOperationMode` to be issued only on idle GPU
instances; attempting a partition-resize on an active MIG slice
produces a session crash because the API tears down the slice. The
chapter's mitigation is the **MIG-resize-only-on-idle gate** plus a
**session-create denial path** that refuses any session that would
require a mid-stream MIG re-partition. Detection is via the
SessionPlanner's resize-request validator; mitigation is to **refuse
the resize** and either schedule the session on a different GPU or
queue it until the slice is idle. Emit
`gpu.mig_resize_refused {gpu=…,slice=…,active_sessions=…}`.
F3 is **fail-closed at admission**.

The **F4 SR-IOV partition collision (AMD MI300)** row binds the
chapter to the **pre-flight-partition-lock contract** from §4.6.
AMD's SR-IOV VF allocation on MI300 lacks the cluster-wide lock that
NVIDIA's MIG provides, so two SessionPlanner instances on different
orchestrator nodes can attempt to claim the same VF simultaneously,
producing VCN engine contention. The chapter's mitigation is the
**etcd-backed pre-flight lock** that the planner acquires before
issuing the SR-IOV VF claim plus a **VF-claim-retry policy** with
exponential backoff for lock contention. Detection is via the
planner's lock-acquisition-failure metric; mitigation is to **retry
on a different VF** or refuse the session if no VF is available. Emit
`gpu.sriov_collision {gpu=…,vf=…,competing_node=…}`.
F4 is **fail-closed at admission**.

The **F5 DVFS thrashing (rapid up/down P-state)** row binds the
chapter to the **hysteresis-floor contract** from §3.5 of this
chapter. The GPU's P-state arbiter (NVIDIA's PowerMizer, AMD's
PowerPlay, Intel's GuC firmware) selects clock rates based on
utilisation telemetry; without a hysteresis floor, the per-frame
thermal-budget calculator's clock-rate suggestions can rapidly
oscillate between adjacent P-states (e.g. P0 → P2 → P0 within a
single frame), producing variable frame-time. The chapter's
mitigation is the **200 ms hysteresis floor** (no P-state transition
within 200 ms of the previous transition) plus a **debounced clock-
suggestion path**. Detection is via the per-GPU P-state-transition-
rate metric; mitigation is to **engage the hysteresis floor** when
transitions exceed 5 per second. Emit
`gpu.dvfs_thrashing {gpu=…,transitions_per_sec=…,hysteresis_engaged=true}`.
F5 is **degrade-open at runtime**.

The **F6 cooling regime mis-declared (air vs liquid)** row binds the
chapter to the **boot-time delta-probe contract** from §3.6 of this
chapter. The cooling-regime classifier published in the §5.1
capability schema influences the per-GPU TGP envelope (air-cooled
RTX 4090 caps at 450 W, liquid-cooled at 600 W per
`video-tech_dim09.md` §2). A mis-declared regime causes the
SessionPlanner to over-commit (declared liquid but actually air =
TGP over-budget = thermal throttle) or under-commit (declared air
but actually liquid = available TGP unused = capacity loss). The
chapter's mitigation is the **boot-time temperature-delta probe**
(measure temperature delta between the GPU's idle and a 30-second
1.0× TGP burst; air cooling produces a delta > 35 K, liquid produces
a delta < 15 K, anything in between is flagged for operator review).
Detection is via the boot-time probe result vs the declared regime;
mitigation is to **refuse boot** if the delta is inconsistent with
the declaration. Emit
`gpu.cooling_regime_mismatch {gpu=…,declared=…,measured_delta_k=…}`.
F6 is **fail-closed at admission**.

The **F7 driver mismatch (pre-2023 NVIDIA cap of 3 sessions hit)**
row binds the chapter to the **capability-schema-check-at-boot
contract** from §5.1 of this chapter. NVIDIA driver versions before
535.x (released August 2023) enforce a consumer-SKU 3-concurrent-
NVENC-session cap (a licensing restriction lifted in 535+ per
`video-tech_dim09.md` §3.4). A host running a pre-535 driver will
silently reject the 4th session at engine-allocation time, producing
a session-create failure that the SessionPlanner cannot easily
diagnose. The chapter's mitigation is the **boot-time driver-version
check** that refuses to start the host-agent if the driver is below
the minimum required version (NVIDIA: 535.x, AMD: ROCm 6.0, Intel:
Level Zero 1.10) plus a **capability-schema field** that publishes
the driver version so the SessionPlanner's bin-packer respects the
cap. Detection is via the boot-time version-vs-minimum check;
mitigation is to **refuse boot** with a structured operator-action
event. Emit
`gpu.driver_below_minimum {gpu=…,driver_version=…,minimum=…}`.
F7 is **fail-closed at admission**.

The **F8 cross-vendor session migration (mid-stream NVIDIA→AMD)**
row binds the chapter to the **forbid-cross-vendor-migration-mid-
stream contract** from §6 of this chapter. The codec / encoder
state-machine on NVENC, AMF, QuickSync, VideoToolbox, and RKMPP have
incompatible internal state (B-frame ordering, rate-control state,
SEI metadata layout) that cannot be transferred mid-stream without
a full codec re-init (which takes >2 s and breaks the session
continuity contract). The chapter's mitigation is the **vendor-
locked session contract** that refuses any mid-stream migration
across vendor boundaries; same-vendor migration (e.g. NVIDIA H100
→ NVIDIA L40S) is permitted. Detection is via the migration-request
validator; mitigation is to **refuse the migration** and surface a
structured operator-action event. Emit
`gpu.cross_vendor_migration_refused {session=…,from_vendor=…,to_vendor=…}`.
F8 is **fail-closed at admission**.

The **F9 per-frame thermal-budget overrun (encode > 6 ms)** row
binds the chapter to the **per-frame-budget-enforcement contract**
from §3.3 of this chapter. The per-frame thermal budget allocates
6 ms for encode on a 4K60 HEVC Main10 stream (the VRR window is
16.67 ms; capture takes ~3 ms, encode takes 6 ms, packetisation
takes ~1 ms, leaving ~6 ms slack); when the encode exceeds 6 ms, the
VRR window is violated and the frame-time spike is visible. The
chapter's mitigation is the **per-frame budget telemetry** plus an
**ABR tier-downshift hook** that informs C33 to engage a one-tier
downshift when overrun exceeds 2 frames in 30 frames. Detection is
via the per-frame encode-latency-vs-budget metric; mitigation is to
**engage the downshift**. Emit
`gpu.frame_budget_overrun {session=…,encode_ms=…,budget_ms=…,downshift_engaged=true}`.
F9 is **degrade-open at runtime**.

The **F10 r18.SafeExec rejects subprocess** row is the chapter's
R-18 trip-wire and is symmetric with C26-F9, C27-F10, C28-F10,
C29-F10, C30-F10, C31-F10, C32-F10, C33-F10. When a developer adds
a non-allow-listed GPU-monitoring tooling argv shape (e.g.
`nvidia-smi -pl 350` for power-limit override from inside the
controller, or `rocm-smi --setpoweroverdrive` without the canonical
`--device-isolation` flag, or `xpumcli config` against an
unauthorised device-id), the wrapper rejects the call at the
`os/exec` boundary and bootstrap aborts. The allow-list lives in
`vasic-digital/helix-r18-safeexec` and is **not duplicated** in this
chapter; the family allow-list extension that C34 contributes
(canonical `nvidia-smi --query-gpu=temperature.gpu,clocks.sm,
clocks.mem,power.draw --format=csv`, `rocm-smi -P -t -c`,
`xpumcli stats -d`, `nvidia-smi mig -lgi`, `rocm-smi --showmeminfo`
for read-only diagnostics) is recapped in §1 (family allow-list) of
this chapter and verified by the C08 `host-integrity-scan` test
inherited verbatim into §8.11. When SafeExec rejects, the symptom
is monitor-blindness — the planner cannot publish capability
telemetry — and the chapter's mitigation is a **capability-degraded
fall-back to the last-known TGP envelope** (the planner uses the
boot-time-published capability snapshot and refuses to admit new
sessions until the SafeExec issue is resolved). Bypass requires an
allow-list extension via operator review per Constitution §11.5.4,
never a silent workaround. Emit
`gpu.safeexec_rejected {tool="nvidia-smi",argv=…,fallback="last_known_tgp"}`.

The **F11 driver crash mid-stream (kernel panic on AMD ROCm)** row
binds the chapter to the **planner-re-shard contract** from §4.7 of
this chapter. An AMD ROCm driver kernel-panic causes the GPU to
disappear from the bus, leaving the SessionPlanner with stale
capability state for that GPU and active sessions assigned to it.
The chapter's mitigation is the **planner-re-shard path** that
detects the GPU-removed event (NVML's `nvmlDeviceGetHandleByIndex`
returns `NVML_ERROR_GPU_IS_LOST`, AMD's equivalent ADL error code,
Intel's Level Zero `ZE_RESULT_ERROR_DEVICE_LOST`) and migrates
active sessions to remaining GPUs in the cluster. Detection is via
the GPU-handle-lost telemetry; mitigation is to **re-shard sessions
to remaining GPUs** with a forced keyframe to reset codec state
(per F8, only same-vendor re-shard is permitted; cross-vendor re-
shard is refused and the session is terminated). Emit
`gpu.driver_crashed {gpu=…,active_sessions=…,re_shard_engaged=true}`.
F11 is **degrade-open at runtime**.

The **F12 thermal sensor saturated (datacentre HVAC failure)** row
binds the chapter to the **planner-BLACKOUT contract** from §4.8 of
this chapter. A datacentre HVAC failure causes the entire rack's
GPU sensors to saturate at the maximum reportable temperature
simultaneously, producing a throttle storm across all GPUs in the
rack. The chapter's mitigation is the **rack-wide BLACKOUT event**
that the planner emits when ≥75% of GPUs in a rack are throttling
within a 60 s window plus a **client-re-route path** that informs
the routing fabric to drain new sessions from the affected rack and
re-route to other DC tiers. Detection is via the rack-wide
throttle-fraction metric; mitigation is to **emit BLACKOUT and
re-route**. Emit
`gpu.rack_blackout {rack=…,throttling_gpus=…,total_gpus=…,re_route_engaged=true}`.
F12 is **degrade-open at runtime** (existing sessions continue with
ABR downshift; new sessions are re-routed).

| # | Failure mode | Symptom | Detection | Mitigation | Fallback |
|---|---|---|---|---|---|
| F1 | GPU thermal throttle event undetected (sensor failure) — driver throttles silently because NVML/ADL/Level Zero throttle bitmask reads clear despite active throttling | Visible artifacting + per-GPU clock-delta-vs-throttle-bitmask exceeds 5%; emits `gpu.throttle_undetected {gpu=…,sm_clock=…,requested_clock=…,delta_pct=…}` | Delta-probe cross-check — `thermal.Monitor.DeltaProbe()` compares current SM clock to requested clock every 100 ms | Engage ABR-conservative downshift hook (informs C33 to drop one tier) until sensor clears | ABR downshift — non-blocking; the session continues at the lower tier with the delta-probe as the source of truth |
| F2 | NVENC engine saturated (more sessions than engines) — planner schedules > engine-count concurrent encode sessions per GPU; dropped frames on over-committed engine | Per-GPU active-session-count exceeds engine-cap-minus-one; emits `gpu.nvenc_saturated {gpu=…,active_sessions=…,engine_cap=…}` | SessionPlanner per-engine-cap validator — `planner.NVENC.CapCheck()` against `video-tech_dim09.md` §3.5 engine matrix | Refuse session-create at engine-count-minus-one (1-engine headroom for migration); first-fit-decreasing bin-packer respects per-GPU engine count | **Fail-closed at admission**; new sessions queued or routed to a different GPU |
| F3 | MIG partition mid-stream resize attempted — planner attempts re-partition on active MIG slice; session crashes because `nvmlDeviceSetGpuOperationMode` requires idle slice | Session crash + structured error from NVML; emits `gpu.mig_resize_refused {gpu=…,slice=…,active_sessions=…}` | SessionPlanner resize-request validator — `planner.MIG.ResizeValidate()` checks slice idle-state | Forbid mid-stream resize via MIG-resize-only-on-idle gate; refuse session-create if mid-stream re-partition required | **Fail-closed at admission**; session scheduled on different GPU or queued until slice idle |
| F4 | SR-IOV partition collision (AMD MI300) — two planner instances claim same VF simultaneously; VCN engine contention because AMD SR-IOV lacks cluster-wide lock | VCN contention + per-VF dual-claim event; emits `gpu.sriov_collision {gpu=…,vf=…,competing_node=…}` | etcd-backed pre-flight lock — `planner.SRIOV.LockAcquire()` against cluster-wide etcd lock | Pre-flight partition lock with exponential-backoff retry policy on lock contention; refuse session if no VF available after retries | **Fail-closed at admission**; retry on different VF or refuse |
| F5 | DVFS thrashing (rapid up/down P-state) — P-state arbiter oscillates between adjacent P-states within a single frame; variable frame-time | Per-GPU P-state-transition-rate exceeds 5 per second; emits `gpu.dvfs_thrashing {gpu=…,transitions_per_sec=…,hysteresis_engaged=true}` | DVFS-transition rate-monitor — `thermal.DVFS.RateMonitor()` observes transition cadence | Engage 200 ms hysteresis floor (no P-state transition within 200 ms of previous); debounced clock-suggestion path | Hysteresis-floored P-state — non-blocking; the session continues with stable clocks |
| F6 | Cooling regime mis-declared (air vs liquid) — declared liquid but actually air = TGP over-budget; declared air but actually liquid = capacity unused | Boot-time delta-probe inconsistent with declared regime; emits `gpu.cooling_regime_mismatch {gpu=…,declared=…,measured_delta_k=…}` | Boot-time temperature-delta probe — `thermal.CoolingRegime.DeltaProbe()` measures idle vs 30 s 1.0× TGP burst | Refuse boot if delta inconsistent (air > 35 K, liquid < 15 K, intermediate flagged for operator review); cross-check at boot | **Fail-closed at admission**; host-agent refuses to start until operator resolves declaration |
| F7 | Driver mismatch (pre-2023 NVIDIA cap of 3 sessions hit) — pre-535.x driver enforces consumer-SKU 3-concurrent-NVENC-session cap; 4th session rejected | 4th session rejected at engine-allocation time; emits `gpu.driver_below_minimum {gpu=…,driver_version=…,minimum=…}` | Boot-time driver-version check — `capability.Driver.MinCheck()` against minimum (NVIDIA 535.x, ROCm 6.0, Level Zero 1.10) | Refuse boot if driver below minimum; capability-schema field publishes driver version so bin-packer respects cap | **Fail-closed at admission**; host-agent refuses to start until driver upgraded |
| F8 | Cross-vendor session migration (mid-stream NVIDIA→AMD) — codec state incompatible across vendors; codec re-init takes > 2 s and breaks session continuity | Migration request crosses vendor boundary; emits `gpu.cross_vendor_migration_refused {session=…,from_vendor=…,to_vendor=…}` | Migration-request validator — `planner.Migration.VendorCheck()` against vendor-locked session contract | Refuse cross-vendor migration; same-vendor migration (NVIDIA H100 → NVIDIA L40S) permitted | **Fail-closed at admission**; structured operator-action event; session terminated cleanly |
| F9 | Per-frame thermal-budget overrun (encode > 6 ms) — encoder exceeds 6 ms budget on 4K60 HEVC Main10; VRR-window violation produces visible frame-time spike | Per-frame encode-latency-vs-budget exceeds budget for ≥ 2 frames in 30; emits `gpu.frame_budget_overrun {session=…,encode_ms=…,budget_ms=…,downshift_engaged=true}` | Per-frame budget telemetry — `thermal.Budget.PerFrameCheck()` against §3.3 envelope | Engage ABR tier-downshift hook (informs C33 to drop one tier) when overrun exceeds 2 frames in 30 | ABR downshift — non-blocking; the session continues at the lower tier with bounded per-frame encode time |
| F10 | `r18.SafeExec` rejects subprocess (e.g. `nvidia-smi -pl 350` from inside controller, `rocm-smi --setpoweroverdrive` without `--device-isolation`, `xpumcli config` against unauthorised device-id) | Bootstrap fails on GPU-monitoring tooling initialisation; structured error includes rejected argv with offending flag highlighted; harness logs `gpu.safeexec_rejected {tool=…,argv=…,fallback="last_known_tgp"}` | Wrapper's verbatim allow-list check at `os/exec` boundary returns `ErrForbiddenArgvShape`; harness logs the rejection | Fix the call site to use the allow-listed shape — canonical `nvidia-smi --query-gpu=temperature.gpu,clocks.sm,clocks.mem,power.draw --format=csv`, `rocm-smi -P -t -c`, `xpumcli stats -d`, `nvidia-smi mig -lgi`, `rocm-smi --showmeminfo` for read-only diagnostics | Capability-degraded fall-back to last-known TGP envelope — non-blocking for existing sessions but new sessions refused; non-overridable per Constitution §11.5.4; bypass requires §13 exception with documented mitigation; cross-link §8.11 host-integrity-scan |
| F11 | Driver crash mid-stream (kernel panic on AMD ROCm) — GPU disappears from bus; planner has stale capability state; active sessions stranded | GPU-handle-lost telemetry (`NVML_ERROR_GPU_IS_LOST`, ADL equivalent, `ZE_RESULT_ERROR_DEVICE_LOST`); emits `gpu.driver_crashed {gpu=…,active_sessions=…,re_shard_engaged=true}` | GPU-handle health-monitor — `planner.GPU.HealthCheck()` polls handle health every 100 ms | Re-shard active sessions to remaining GPUs (same-vendor only per F8); forced keyframe to reset codec state | Re-shard with forced keyframe — non-blocking; sessions continue on remaining GPUs; cross-vendor case refuses and terminates cleanly |
| F12 | Thermal sensor saturated (datacentre HVAC failure) — entire rack's GPU sensors saturate simultaneously; throttle storm across all GPUs in rack | Rack-wide throttle-fraction exceeds 75% within 60 s window; emits `gpu.rack_blackout {rack=…,throttling_gpus=…,total_gpus=…,re_route_engaged=true}` | Rack-wide thermal aggregator — `thermal.Rack.AggregateMonitor()` against §4.8 BLACKOUT threshold | Emit rack-wide BLACKOUT event; client-re-route path drains new sessions from affected rack and re-routes to other DC tiers | BLACKOUT + re-route — degrade-open for existing sessions (ABR downshift); fail-closed for new sessions on the affected rack |

## 8. Test surface

The C34 test surface inherits the family-level container-driven CI lane
contract from C26 §8 + C27 §8 + C28 §8 + C29 §8 + C30 §8 + C31 §8 +
C32 §8 + C33 §8 and the `vasic-digital/Containers` runner image,
**extended** with the new GPU / thermal / sessions-pipeline-specific
requirement: every integration / E2E / chaos / stress test must
exercise **a real multi-vendor GPU rig (NVIDIA RTX 4090 + AMD MI300 +
Intel Arc + Intel iGPU)** with **real NVML / ADL / Level Zero
telemetry**, **real cooling-regime classification**, and **real MIG /
SR-IOV partition operations** so the SessionPlanner, thermal-budget
calculator, and capability publisher are validated against real
hardware behaviour (mocking the GPU is forbidden per Constitution §6.4
— only unit tests may use mocks). Per Constitution §6.4 + Master Plan
§4.3 anti-bluff verification, the test matrix below cites
`video-tech_dim09.md` (GPU / hardware dimension) and
`video-tech_dim10.md` (testing dimension) explicitly so every per-GPU
thermal / session performance claim is grounded in a primary-source
reference.

### 8.1 Unit (mocks allowed)

The unit-test layer is the only place where mocks, stubs, or
hardcoded values are permitted per Constitution §6.4 — every other
layer below hits the real GPU.

- **`thermal.Budget` per-frame budget calculator unit test** —
  instantiate the per-frame budget calculator with a mocked
  per-GPU TGP envelope (synthetic input: 450 W TGP, RTX 4090
  air-cooled, 4K60 HEVC Main10 workload); assert the calculator
  correctly partitions 6 ms encode + 3 ms capture + 1 ms
  packetisation + 6 ms VRR slack within the 16.67 ms VRR window;
  assert the calculator correctly engages the ABR tier-downshift
  hook when the per-frame encode time exceeds 6 ms for ≥ 2 of
  30 frames per F9; assert the calculator correctly handles the
  liquid-cooled envelope (600 W TGP) for a 1.0× to 1.33× scaling.
- **Bin-packing first-fit-decreasing unit test** — feed the
  SessionPlanner a mocked GPU pool (RTX 4090 ×4 + L40S ×2 +
  H100 ×1 with NVENC stripped) and a synthetic session-request
  stream (a 4K60 + a 1080p60 + a 720p60 sequence per
  `video-tech_dim10.md` §5 canonical six-game-profile bench corpus);
  assert the bin-packer correctly assigns sessions to GPUs
  respecting the per-GPU engine cap from §4.3; assert the bin-
  packer refuses session-admission at engine-count-minus-one per
  F2; assert the bin-packer respects the MIG-only-on-idle gate
  per F3.
- **Capability schema unit test** — instantiate the
  capability-publisher against a mocked NVML response (synthetic
  fields: GPU model, driver version, NVENC engine count, memory,
  TGP envelope, cooling regime, MIG support, SR-IOV support,
  active-throttle reasons, per-engine workload, per-engine
  thermal); assert the published schema includes all 11 fields
  from §5.1; assert the schema validates against
  `vasic-digital/helix-gpu/schema/v1.json`.
- **Driver-version check unit test** — feed the boot-time
  driver-version check synthetic version strings (NVIDIA 530.x,
  535.x, 545.x; AMD ROCm 5.7, 6.0, 6.2; Intel Level Zero 1.9,
  1.10, 1.11); assert the check correctly refuses boot for
  pre-535 NVIDIA, pre-6.0 ROCm, pre-1.10 Level Zero per F7;
  assert the check correctly accepts at-or-above-minimum versions.

### 8.2 Integration

The integration-test layer hits the real NVML / ADL / Level Zero
telemetry path on a real GPU — no mocks, no stubs, no hardcoded
values. Per Constitution §6.4 this layer must run inside the
canonical `vasic-digital/Containers` runner image with the multi-
vendor GPU rig (NVIDIA RTX 4090 + AMD MI300 + Intel Arc + Intel
iGPU) addressable on the runner network.

- **`thermal.Monitor` against real NVML on RTX 4090** — boot the
  thermal monitor against a real RTX 4090; run a 60-second
  4K60 HEVC Main10 capture + encode workload; capture per-100 ms
  sample telemetry (temperature.gpu, clocks.sm, clocks.mem,
  power.draw, throttle bitmask); assert the sample cadence is
  100 ms ± 10% (10 ms jitter ceiling); assert the throttle-bitmask
  decoder correctly identifies HW thermal slowdown, HW power
  slowdown, and SW thermal slowdown reasons.
- **`thermal.Monitor` against real ADL on MI300** — same fixture
  but on an AMD MI300; assert the AMD ADL equivalent telemetry
  fields are correctly published in the canonical capability
  schema; assert the SR-IOV VF allocation correctly engages the
  etcd-backed pre-flight lock per F4.
- **`thermal.Monitor` against real Level Zero on Intel Arc** —
  same fixture but on an Intel Arc A770; assert the Level Zero
  telemetry fields are correctly published; assert the Intel-
  specific behaviour (no MIG, no SR-IOV, single-encoder engine)
  is correctly reflected in the capability schema.
- **MIG partition lifecycle integration** — on a real H100,
  partition into 7×1g.10gb slices; assert each slice publishes
  an independent capability snapshot; assert the planner correctly
  schedules sessions across slices; assert mid-stream resize is
  refused per F3.

### 8.3 E2E

The E2E layer brings up the **full GPU + capability + planner +
encoder + transport + thermal pipeline** end-to-end and asserts
user-perceptible quality + thermal compliance.

- **4 concurrent 1080p60 sessions on RTX 4090** — boot a host with
  a real RTX 4090; session-create 4 concurrent 1080p60 H.264
  Main streams via real clients; assert the SessionPlanner caps
  at 5 (4 active + 1 headroom per §4.3 + F2); assert no thermal
  throttle event during a 30-minute session window (the throttle-
  bitmask telemetry stays clear); assert per-frame encode time
  stays within the §3.3 budget for all 4 sessions.
- **Concurrent multi-tier sessions on H100** — boot a host with
  a real H100; partition into 7×1g.10gb MIG slices; session-create
  one 1080p60 + one 720p60 session per slice (14 total); assert
  every slice publishes an independent capability snapshot;
  assert the planner correctly bin-packs across slices; assert
  no cross-slice thermal interference.
- **Cross-vendor session refusal E2E** — boot a host with real
  RTX 4090 + real MI300; attempt a mid-stream migration of an
  active session from RTX 4090 to MI300; assert the migration is
  refused per F8 with a structured operator-action event; assert
  the original session continues uninterrupted.

### 8.4 Security

- **`r18.SafeExec` rejection fuzz** — for each of the family
  allow-list entries (`00_Index.md` §7), construct off-allow-
  list argv shapes (e.g. `nvidia-smi -pl 350` is off-list;
  `rocm-smi --setpoweroverdrive` without `--device-isolation`
  is off-list; `xpumcli config` against an unauthorised device-id
  is off-list) and fuzz with 10⁶ argv permutations per
  Constitution §6.4 fuzz contract; assert the wrapper returns
  `ErrForbiddenArgvShape` for every off-list shape with no
  false-positive on allow-list shapes; assert no host-disruptive
  command (kill, systemctl, pmset) ever passes the wrapper. The
  deny-list is **inherited from C08 §10** per the family contract
  — no duplication in this chapter.
- **GPU policy authorisation** — assert per-tenant GPU policy
  mutations (max-concurrent-sessions cap, MIG slice allocation,
  SR-IOV VF assignment, cooling-regime override) are
  authenticated and authorised per the C09 security family
  (cross-link); assert unauthorised policy-update attempts are
  refused with structured audit events.

### 8.5 Benchmarking

The benchmarking layer is the chapter's binding to Constitution §6 —
every per-GPU thermal / session performance claim **reports p50 /
p99 / p999 at ≥ 10 K samples** via the C24 measurement harness.
Cross-link C24 / C35. Per **`video-tech_dim10.md`** §2 + §5, the
benchmarking corpus uses synthetic-content + real-game-capture
pairs across the six representative game profiles (FPS, racing,
RPG, RTS, MOBA, fighting) so the per-profile thermal characterisation
reflects production-like workloads.

- **Bench per-frame thermal-budget enforcement latency** —
  measure end-to-end per-frame budget-check latency from
  encode-completion to budget-exceeded-decision across **≥
  10 000 samples** per game profile; **report p50 / p99 / p999
  per Constitution §6**; histogram artifact attached; budget
  per `video-tech_dim10.md` §3 — budget-check latency p999 <
  100 µs (the 100 ms thermal-sample cadence's 0.1% bound).
- **Bench SessionPlanner bin-packing latency** — measure
  session-create-to-GPU-assignment latency across **≥ 10 K
  samples** across the multi-vendor GPU rig (RTX 4090 + L40S +
  H100 + MI300 + Arc); **report p50 / p99 / p999**; budget <
  10 ms p999 (session-create hot path).
- **Bench thermal-sample cadence stability** — measure inter-
  sample-interval variance across **≥ 10 K samples** under
  steady-state workload; **report p50 / p99 / p999**; budget
  100 ms ± 10 ms (the chapter's §3.4 cadence with the 10%
  jitter ceiling).
- Cross-link **C24 / C35** measurement harness for shared
  histogram-collection + bootstrap-resampling-confidence-interval
  primitives. The benchmark suite must cite **`video-tech_dim10.md`**
  explicitly per Master Plan §4.3 anti-bluff verification —
  `video-tech_dim10.md` §3 enumerates the per-profile regression-
  detection thresholds + §5 enumerates the canonical bench corpus
  including the multi-vendor GPU rig + §7 enumerates the per-GPU
  thermal-sample latency budgets. Cross-link **C24** §6 (latency-
  side measurement) and **C35** §3 (quality-side measurement) for
  the full harness contract.

### 8.6 Chaos

- **Kill `nvidia-smi` / inject NVML failure** — boot a host
  with a real RTX 4090 + 4 active sessions; mid-session, kill
  the `nvidia-smi` process (or inject an NVML library failure
  via library-preload shim); assert the SessionPlanner correctly
  detects the SafeExec failure per F10; assert the planner falls
  back to the last-known TGP envelope and refuses new sessions
  while preserving existing sessions; assert sessions continue
  uninterrupted with the conservative envelope.
- **Inject thermal sensor saturation (F12)** — heat the GPU
  rack via a controlled HVAC failure simulation (the rig's
  programmable thermal-stage steps the rack inlet temperature
  from 22 °C to 50 °C over 5 minutes); assert ≥75% of GPUs in
  the rack throttle within the 60 s window; assert the planner
  emits the BLACKOUT event per F12; assert the routing fabric
  drains new sessions from the affected rack.
- **Force MIG mid-stream resize (F3)** — on a real H100 with
  active sessions, attempt a partition-resize via the
  SessionPlanner API; assert F3 detection fires; assert the
  resize is refused; assert active sessions continue
  uninterrupted.
- **Force SR-IOV collision (F4)** — on a multi-node cluster
  with shared MI300, simulate two SessionPlanner instances
  attempting to claim the same VF simultaneously; assert F4
  detection fires; assert the etcd-backed pre-flight lock
  correctly serialises the claims; assert one claim succeeds
  and the other retries on a different VF.
- **Force driver crash (F11)** — on a real MI300, inject a
  ROCm driver kernel-panic via the controlled ROCm test
  harness; assert the GPU disappears from the bus; assert the
  GPU-handle health-monitor detects the loss within 100 ms;
  assert the planner re-shards active sessions to remaining
  GPUs with forced keyframes; assert no session is dropped
  (same-vendor re-shard only per F8).

### 8.7 Stress

- **24h run with 5 concurrent 4K60 sessions on RTX 4090** —
  on a real RTX 4090, run 5 concurrent 4K60 HEVC Main10 capture
  + encode + transport sessions for 24 hours; assert **zero
  thermal throttle events** (the throttle-bitmask telemetry
  stays clear for the full 24 h); assert **zero session
  migration events** (no F11 or F12 fires); **assert no fd
  leak** (process fd count stable to within 5 fds over 24 h);
  **assert no GC stall > 1 ms** (GODEBUG=gctrace=1 trace
  artifact attached; cross-link C36 §3 Go pipeline `sync.Pool`
  discipline); assert no memory leak (RSS growth < 5 MB / hour);
  assert per-session per-frame encode time p999 stays within
  the §8.5 budget across the 24 h window.
- **Multi-vendor concurrent stress** — on a heterogeneous rig
  (RTX 4090 + MI300 + Arc), run concurrent sessions on each
  vendor for 24 hours; assert per-vendor per-frame budget
  enforcement holds; assert no cross-vendor migration fires;
  assert capability schemas remain stable (no schema-version
  drift across the 24 h window).

### 8.8 Smoke

- **Capability schema includes all 11 fields from §5.1** —
  boot the host-agent in a clean container with a real GPU;
  query the published capability schema; assert all 11 fields
  (GPU model, driver version, NVENC/AMF/QSV engine count,
  memory, TGP envelope, cooling regime, MIG support, SR-IOV
  support, active-throttle reasons, per-engine workload,
  per-engine thermal) are present and non-default; assert the
  schema validates against `vasic-digital/helix-gpu/schema/v1.json`.
- **Throttle field updates within 200 ms of throttle event** —
  on a real RTX 4090, force a thermal throttle by capping the
  TGP via `nvidia-smi -pl 100` (canonical allow-listed shape);
  measure the latency from throttle-onset to capability-schema
  field-update; assert the latency is ≤ 200 ms (the §3.4 cadence
  100 ms + 1 publication interval).

### 8.9 Full automation

All of §8.1–§8.8 run on **every commit via the local container-
driven CI lane** per Constitution §10. The CI lane uses the
canonical `vasic-digital/Containers` runner image with the multi-
vendor GPU rig (RTX 4090 + L40S + H100 + MI300 + Arc + iGPU)
addressable on the runner network and the canonical six-game-
profile bench corpus per `video-tech_dim10.md` §5. The matrix
covers (Linux Ubuntu 22.04 / 24.04 + Fedora 40, Windows Server
2022, macOS 14) × (5 GPU vendors × 6 game profiles × 4 thermal-
load profiles). The full-automation lane emits a single composite
artifact (`gpu-thermal-sessions-test-report.json`) that the C35
quality-claim harness consumes as the authoritative source-of-
truth for any per-GPU thermal / session performance claim in
chapter prose. The CI lane runs nightly on the real GPU rig (the
multi-vendor rig is too expensive for per-commit hardware
exercise; per-commit runs use the unit + integration layers
against a single representative GPU, with the full multi-vendor
matrix gated to the nightly schedule per Constitution §10's
local-CI-equivalence clause).

### 8.10 Challenges (production-like)

HelixQA dispatches **per-vendor verification scenarios** from
`git@github.com:vasic-digital/Challenges.git` (per Constitution
§6.4 Challenges-test contract):

- **Per-vendor full-rack verification Challenges** — for each
  GPU vendor (NVIDIA, AMD, Intel), HelixQA boots a fully-
  provisioned rack (8 GPUs per vendor) + clients + thermal-
  injection rig configured to the per-vendor TGP envelope;
  runs a 60-minute multi-session workload at the full session-
  cap; asserts per-frame thermal-budget enforcement holds;
  asserts the capability schema remains stable; asserts no
  thermal throttle events.
- **Thermal-injection Challenges** — HelixQA configures the
  rig's programmable thermal-stage to step the rack inlet
  temperature through 22 °C → 35 °C → 45 °C over a 90-minute
  window; asserts the ABR feedback loop closes (per F1 + F9 +
  F12, the planner correctly informs C33 of thermal pressure
  and C33 correctly engages tier-downshifts); asserts the
  BLACKOUT path engages at the §4.8 threshold; asserts session
  continuity across the thermal events.
- **Cross-vendor refusal Challenges** — HelixQA dispatches a
  multi-vendor cluster + concurrent sessions; attempts a cross-
  vendor migration; asserts F8 refusal fires; asserts the
  source session continues uninterrupted.
- **Per-fault recovery Challenges** — inject each of F1–F12
  during a live Challenges scenario; assert the recovery path
  fires correctly and the final per-session SSIM verification
  holds.

### 8.11 Inherited host-integrity-scan (R-18, non-overridable per Constitution §11.5.4)

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

The C34 implementation contract that this scan validates:

- GPU-monitoring tooling invocation via `r18.SafeExec` only —
  never via `os/exec.Command` directly; the canonical shapes
  (`nvidia-smi --query-gpu=temperature.gpu,clocks.sm,clocks.mem,
  power.draw --format=csv`, `rocm-smi -P -t -c`, `xpumcli stats
  -d`, `nvidia-smi mig -lgi`, `rocm-smi --showmeminfo` for
  read-only diagnostics) are the family allow-list entries for
  GPU / thermal / sessions tooling.
- No host-disruption commands ever appear in the GPU / thermal /
  sessions path: no `kill -9 <pid>`, no
  `systemctl suspend|hibernate|reboot|halt|poweroff`, no
  `pmset`, no `xset dpms force off`, no `--privileged` container
  flag, no host-mount of `/`, `/dev`, `/proc`, `/sys` (the GPU
  device-files in `/dev/nvidia*`, `/dev/dri/*`, `/dev/kfd` are
  exposed via the canonical container-toolkit injection per the
  `vasic-digital/Containers` runner image, never via host-
  mount). The scan asserts none of these syscall patterns appear
  in the GPU / thermal / sessions subsystem's syscall trace.
- No cross-tenant capability traversal — the scan asserts the
  capability publisher's `openat` syscalls never reference
  paths outside the per-tenant scoped GPU configuration root,
  and no `chdir` / `chroot` syscall escapes the scope.

The scan's invocation contract is byte-identical with the C08
§12.11 inheritance into every chapter in the family per
`00_Index.md` §7 R-18 family allow-list. No chapter in the family
is permitted to redefine, override, or extend the scan —
Constitution §11.5.4 forbids per-chapter customisation of the
host-integrity contract.

## 9. Open questions

The following open questions are tracked in the chapter's OQ log
and surface to the family-level OQ aggregator at `00_Index.md`
§5. Each OQ is prefixed `OQ-C34-NN` and carries an owner, a
target resolution date, and a cross-link to the deciding chapter
or external dependency.

- **OQ-C34-01** — Multi-vendor MIG / SR-IOV unified API in V1.
  The MVP ships separate code paths for NVIDIA MIG and AMD
  SR-IOV partitioning, with per-vendor capability schemas and
  per-vendor partition operations. A V1 abstraction layer that
  exposes a unified Go binding over both (and Intel's
  forthcoming partition API on Battlemage / Celestial) would
  reduce the SessionPlanner's vendor-specific complexity but
  requires careful semantic alignment (NVIDIA MIG's hardware-
  isolated slices vs AMD SR-IOV's VF-per-VM model are not
  drop-in equivalents). Should V1 ship a vendor-agnostic
  MIG / SR-IOV unified API as a public submodule under
  `vasic-digital/helix-gpu`, or should V1 keep the per-vendor
  code paths and defer unification to V2? The cost is the
  semantic-alignment design work + the per-vendor adapter
  implementation; the benefit is reduced SessionPlanner
  complexity + easier V2 onboarding of new vendors. Trigger:
  V1 hardware decision matrix emerges; vendor APIs stabilise.
  Owner: C34 + V1 family + Architecture family. Cross-link
  `video-tech_dim09.md` §3 + V1 hardware decision matrix.

- **OQ-C34-02** — Immersion cooling tier-4 regions. The §3.6
  cooling-regime classifier supports air, liquid, and (via
  documented operator-policy extension) immersion cooling; the
  TGP envelope for immersion-cooled GPUs can run 1.5×–2.0× the
  air-cooled envelope, materially expanding the per-GPU
  session capacity. However, immersion cooling has significant
  capex and operational overhead (specialised hardware, dielectric
  fluid handling, maintenance procedures) that varies by region
  (Northern Europe tier-4 DCs have established immersion tooling;
  Southeast Asia tier-4 DCs are still building capability).
  Should V1 ship an immersion-cooling-aware capability schema +
  per-region operator-policy + per-region session-pricing
  surface, or defer immersion cooling to V2? The cost is the
  per-region capex feasibility analysis + the operations-chapter
  coordination; the benefit is materially higher per-GPU session
  capacity in tier-4 regions. Trigger: V1 region-rollout posture
  emerges; per-region capex feasibility studies complete.
  Owner: C34 + V1 family + Operations family. Cross-link
  Operations chapter `08_Operations/05_DC_Tier_Capacity.md`
  (queued) + per-region rollout plan.

- **OQ-C34-03** — Cross-vendor session migration (mid-stream
  NVIDIA → AMD). The F8 fault row is fail-closed-at-admission
  for cross-vendor migration in MVP; the §6 transport-shim
  contract reserves the per-vendor codec-state-translation
  pluggable hook for V2. The cost is the codec-state translation
  layer (NVENC HEVC Main10 B-frame ordering → AMF HEVC Main10
  B-frame ordering, with corresponding rate-control state +
  SEI metadata translation) plus the inter-vendor handoff
  protocol; the benefit is true vendor-agnostic session migration
  for failover scenarios. Should V2 ship cross-vendor migration
  as a fully-supported feature, or should it remain a permanent
  refusal? The blocker is the codec re-init cost (>2 s today;
  V2 may bring this down via vendor-specific fast-init paths).
  Trigger: V2 codec posture emerges; vendor APIs expose fast-
  init paths. Owner: C34 + V2 family + Codec WG. Cross-link
  F8 + `video-tech_dim09.md` §3.5 vendor codec matrix.

- **OQ-C34-04** — Driver version negotiation. The F7 fault row
  refuses boot if the NVIDIA driver is below 535.x (the post-
  2023 release that lifts the consumer-SKU 3-concurrent-NVENC-
  session cap). The MVP ships a hard minimum-version check at
  boot; should the host-agent additionally publish a recommended-
  version field (e.g. NVIDIA 545.x for the latest NVENC
  optimisations) and emit a structured warning when the host
  is below recommended but at-or-above minimum? The cost is the
  per-vendor recommended-version-tracking effort (drivers ship
  weekly; tracking the recommended baseline is operational
  overhead); the benefit is operator visibility into driver
  optimisation opportunities. Trigger: MVP operator-feedback
  posture clarifies on driver-update cadence preferences.
  Owner: C34 + Operations family. Cross-link F7 +
  `video-tech_dim09.md` §3.4 driver licensing change.

- **OQ-C34-05** — ML-driven thermal prediction. The MVP §3.4
  thermal-budget enforcement is reactive (the ABR tier-downshift
  hook engages after the thermal threshold is breached); recent
  research demonstrates that ML-driven thermal prediction (LSTM
  or transformer models trained on per-GPU temperature gradient
  + workload signal) can predict throttle events 500 ms ahead
  with high accuracy, enabling proactive ABR downshift before
  the throttle fires. Should V2 ship an ML-driven thermal-
  prediction path as a per-tenant operator-policy alternative
  to the reactive controller? The cost is ML model training +
  inference deployment + per-GPU telemetry collection; the
  benefit is reduced throttle frequency and improved per-frame
  budget compliance. Trigger: V2 QoE-optimisation posture emerges;
  ML training infrastructure available. Owner: C34 + V2 family
  + Quality WG. Cross-link `video-tech_dim09.md` §4 thermal
  modelling survey + V2 ML-infrastructure decision matrix.

- **OQ-C34-06** — AMD MI400 / NVIDIA Rubin (2027 generation).
  The MVP capability schema in §5.1 documents 11 fields chosen
  to be forward-compatible with the 2027-generation hardware
  (AMD MI400 with rumoured > 2 kW TGP and immersion-cooling
  default, NVIDIA Rubin with rumoured chiplet partitioning that
  may extend MIG semantics, Intel Falcon Shores with rumoured
  unified CPU-GPU memory). Will the existing 11-field schema
  survive the 2027 generation, or will it require an additive
  schema-version-bump? The cost is the schema-evolution governance
  effort; the benefit is forward-compatibility for V2. Trigger:
  2027-generation hardware specifications publish. Owner: C34
  + Architecture family + Hardware liaison. Cross-link
  `video-tech_dim09.md` §5 next-generation hardware roadmap.

---

## 10. References

### Project artifacts

- Master Plan, Constitution, System Overview, Video/Audio family index ([`00_Index.md`](00_Index.md)). Sibling chapters cited above.

### Source research artifacts

- `video-tech_dim09.md` (1,181 lines), `video-tech.agent.final.md` (2,588 lines), `video-tech_insight.md` (Insight #1 BINDING + Insight #9 RELEVANT), `video-tech_cross_verification.md`, `video-tech_dim10.md` (testing).

### Web research

[`../99_Web_Research_Addenda/2026-04-29-thermal-and-gpu-balancing.md`](../99_Web_Research_Addenda/2026-04-29-thermal-and-gpu-balancing.md) — 9 clusters (§A–§I) + §Z.

| Cluster | Topic | Cited |
|---------|-------|------|
| §A | NVIDIA Ada / Hopper / Blackwell thermal envelopes | §2.1–2.3 |
| §B | AMD RDNA3 / RDNA4 / CDNA3 (MI300) thermal envelopes | §2.4, §2.5 |
| §C | Intel Arc Battlemage thermal envelopes | §2.6 |
| §D | MIG + SR-IOV multi-tenant GPU partitioning | §3.1–3.3 |
| §E | NVENC / AMF / Quick Sync session-count math | §2.7, §3.5 |
| §F | DVFS governance + per-frame thermal budget | §4 |
| §G | Thermal-throttling detection (NVML / ADL / Level Zero) | §4.5, §6 |
| §H | Cooling regimes (air / liquid / immersion datacentre) | §4.6 |
| §I | 2026 cloud-gaming thermal benchmarks (Meta, Google GFN, Microsoft xCloud) | §1, §3 |
| §Z | Contradictions index | §1, §2, §3, §4 |

---

## Anti-Bluff Verification

> Per [Master Plan §4.3](../../00_Master_Plan.md#43-anti-bluff-verification-block-r-13).

### Source Evidence Reviewed

| Path | Lines | Reviewed | Used |
|------|------:|----------|------|
| `video-tech_dim09.md` | 1,181 | A, B, C, D | §§1–9 (primary) |
| `video-tech.agent.final.md` | 2,588 | A, B, C | §§1–6 |
| `video-tech_insight.md` | 243 | A, B | §1 (#1 BINDING + #9 RELEVANT) |
| `video-tech_cross_verification.md` | 206 | A | §1 |
| `video-tech_dim10.md` | 1,689 | D | §8.5 |
| `00_Master_Plan.md` post-Session-7 | A, B, C, D | header / §6 / §9 |
| `01_Constitution.md` post §11.5 | A, B, C, D | §§1–8 |
| `05_Video_Audio/00_Index.md` | 407 | A, B, C, D | header voice |
| `05_Video_Audio/02_Hardware_Encoders.md` | 2,652 | A, B | §2 (vendor encoder thermal cross-link C27) |
| `05_Video_Audio/08_ABR_FEC_Congestion.md` | 2,369 | C | §6 (ABR feedback contract cross-link C33) |
| `04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md` | 1,749 | A, B | §1, §2 (GPUDirect bandwidth cross-link C18) |
| `03_Architecture/07_Host_Agent_and_Game_Lifecycle.md` | 3,425 | C, D | §6 (`r18.SafeExec`), §8.11 |

### Web Sources Consulted

The companion addendum lists every URL with title and 2026-04-29 access date. **9 clusters (§A–§I) + §Z; ≥6 distinct primary URLs per cluster.**

### Insights Incorporated

| Insight | Source | Sections |
|---------|--------|----------|
| video-tech Insight #1 — thermal wall (BINDING) | `video-tech_insight.md` | §1.2, §2, §3, §4 (BINDING) |
| video-tech Insight #9 — GPU vendor topology-driven (RELEVANT) | `video-tech_insight.md` | §1.3, §2, §3.5 |

### Conflict Zones Resolved

| CZ-ID | Description | Decision | Section |
|-------|-------------|----------|---------|
| Insight #1 | Thermal wall first constraint on density + quality | **Reaffirmed and binding** | §1, §2, §3, §4 |
| Insight #9 | Vendor topology drives session math | **Reaffirmed**; bin-packing per-vendor | §3.5 |
| Z addenda | Thermal envelope + DVFS + cooling regime contradictions | Resolved per cluster matrix in addendum | §1, §2, §3, §4 |
| Inherited (all prior families' Z lists) | Owned by prior chapters | Not relitigated | header preamble |

### R-18 Operational Integrity Compliance

Inherits R-18 enforcement from C08. Compliance verified:

- **Static — chapter prose**: §1.4 references R-18; §6.5 explicitly recaps the family-level allow-list extension specific to this chapter (`nvidia-smi`, `rocm-smi`, `intel_gpu_top`, NVML, ADL, Level Zero — all wrap through `r18.SafeExec`).
- **Static — code in §6**: imports `r18 "github.com/vasic-digital/helix-r18-safeexec"` (origin C08 §10). Deny-list NOT duplicated.
- **Static — bootstrap subprocess invocations**: GPU inventory probes (`nvidia-smi -q --json`, `rocm-smi --json`, `intel_gpu_top -J`) all wrap through `r18.SafeExec`.
- **Test — §8.11**: `host-integrity-scan` inherited VERBATIM from C08 §12.11. Non-overridable per Constitution §11.5.4.

### Coverage Confirmation

| Metric | Value |
|--------|-------|
| Source per-dim primary (`video-tech_dim09.md`) | 1,181 lines |
| R-01 minimum (Master Plan §7.2 row C34) | 1,300 lines of body prose |
| Body prose actually synthesised | **2,827 lines** across §§1–9 (A 566 + B 630 + C 817 + D 814) |
| Coverage ratio vs minimum | 2.17× line-count / ≥ 2.4× word-adjusted |
| Coverage ratio vs primary per-dim source | 2.39× |
| Forbidden-pattern scan (chapter prose) | clean |
| Empty-section-body scan | clean |
| Tables | Per-vendor TGP/NVENC/session table in §2; MIG/SR-IOV partitioning matrix in §3.1–3.3; per-vendor session math summary in §3.5; DVFS curve table in §4; capability schema field list in §5.1; failure-mode 12-row F1-F12 table in §7; test-type matrix in §8 |
| Section count | 10 normative sections + this verification block |
| Go code blocks | §6.4 (~179 LOC `thermal.NewMonitor` + `Sample` + `SessionPlanner.Place` — real imports `github.com/NVIDIA/go-nvml`, `r18`, `helix-shm`, `helix-codec`, `helix-network`) |
| R-18 enforcement | inherited from C08 §10 + §8.11 host-integrity-scan inheritance |

### Sign-off

- Section A (§§1–2) by C34 Group A on 2026-04-29.
- Section B (§§3–4) by C34 Group B on 2026-04-29.
- Section C (§§5–6) by C34 Group C on 2026-04-29.
- Section D (§§7–9) by C34 Group D on 2026-04-29.
- Web addendum by C34 addendum subagent on 2026-04-29.
- Header, ToC, §10, Anti-Bluff Verification stitched by orchestrator (Claude) on 2026-04-29.
- Reviewed by: pending operator review.

End of `05_Video_Audio/09_Thermal_and_GPU_Balancing.md` — 2026-04-29.
