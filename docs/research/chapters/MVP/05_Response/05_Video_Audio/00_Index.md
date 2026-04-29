# Video & Audio — Chapter Index

> **Family scope.** This is the index for HelixPlay's
> `05_Video_Audio/` chapter family — the deep, codec-/encoder-/
> capture-/recording-/HDR-/audio-/transport-engineering elaboration
> of the video and audio surfaces that the Architecture chapter
> family (C01..C13) names as the binding architectural floor.
> Where C01 (Streaming Protocols & Codecs) is the *what* of codec
> selection at the architectural level, and C03 (Host OS Capture)
> is the *what* of capture primitives, this family is the *how* —
> per-codec encoder profile tuning, dual-path stream+record
> orchestration, recording storage backends, multi-channel audio
> passthrough, HDR tone-mapping, ABR + FEC + congestion control,
> thermal- and GPU-aware load balancing, end-to-end measurement,
> the Go-based pipeline implementation, and the network-transport
> ladder that ties it all together.
>
> **Source stream:** Stream 3 (`docs/research/chapters/MVP/03_video_technology/`)
> — 12 video-technology dimensions plus a long-form synthesis at
> [`docs/research/chapters/MVP/03_video_technology/02_Response/video-tech.agent.final/video-tech.agent.final.md`](../../03_video_technology/02_Response/video-tech.agent.final/video-tech.agent.final.md).
>
> **Status:** index landed; C26..C37 (12 deep chapters) queued
> under R1 section-stitched dispatch (Master Plan §5).
>
> **Last updated:** 2026-04-29.

---

## 1. Why this family exists

The Architecture family ends with C13
[`../03_Architecture/12_Latency_Engineering_Overview.md`](../03_Architecture/12_Latency_Engineering_Overview.md)
— a 3,816-line architectural overview that names video + audio as
the *single largest contributor to overall latency budget after
network*. Stream 3 (`03_video_technology/`) decomposes this into
12 dimensions covering codec selection through Go pipeline
implementation. The chapters indexed here (C26..C37) are the
deep elaborations.

Every chapter in this family inherits, without re-implementing
(Constitution §2 DRY):

- The `r18.SafeExec` wrapper from
  [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md)
  §10 — used for any video-tooling subprocess invocation
  (`ffmpeg`, `gst-launch-1.0`, `nvidia-smi --query-gpu=...`,
  `vainfo`, `qsv-tools`, `obs-studio --headless`, etc.).
- The `host-integrity-scan` test from C08 §12.11 — non-overridable
  per Constitution §11.5.4.
- The Constitution-§6 mandate that **every video / audio quality
  + latency claim reports p50 / p99 / p999 at ≥ 10 K samples**
  via the C24 measurement-harness from the Latency family.
- The PresentMon 2.2 wrapper introduced in
  [`../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md)
  §10.6 + reused by C24 — chapters in this family consume the
  per-frame trace stream produced there, never re-implement.

The family is the canonical home for the **ten video-technology
insights** documented in
[`../../03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md`](../../03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md):

| # | Insight | Primary chapter | Cross-cited |
|---|---------|-----------------|-------------|
| 1 | **Thermal wall** — dual-path encoding is GPU-thermal-bounded, not session-count-bounded | C34 (`09_Thermal_and_GPU_Balancing.md`) | C27, C28, C29 |
| 2 | **Audio passthrough constrained** — multi-channel audio is binary (full or stereo); validate end-to-end | C31 (`06_Audio_Pipeline.md`) | C32, C37 |
| 3 | **H.264 codec sweet spot paradox** — H.264 strategically optimal despite technical inferiority (98%+ decode coverage; mandatory WebRTC) | C26 (`01_Codec_Selection.md`) | C27, C33 |
| 4 | **Recording = save system** — local buffer + background sync mirrors save-file pattern | C30 (`05_Recording_Storage.md`) | C29 |
| 5 | **Go goroutines map to pipeline stages** — channel-based stage hand-off; `sync.Pool` for `[]byte` frames | C36 (`11_Go_Pipeline_Implementation.md`) | C28, C30 |
| 6 | **Display pipeline is largest unaddressed latency** — 30–100 ms TV processing; ALLM + game-mode mitigates | C32 (`07_HDR_and_Color.md`) | C13 §8 + C22 |
| 7 | **SQP + custom UDP next-gen** — sub-10 ms LAN; TCP-friendly without WebRTC overhead | C37 (`12_Network_Transport.md`) | C19 |
| 8 | **VVC hardware gap; AV1 correct bet** — VVC complexity 8–10× H.264; no real-time path before 2028 | C26 + C27 | C33 |
| 9 | **GPU vendor selection topology-driven** — Intel ULL latency-first; NVIDIA scale; AMD budget | C27 (`02_Hardware_Encoders.md`) | C34 |
| 10 | **Recording differentiates** — DVR-for-PC-gaming feature no competitor offers | C30 | C28 (Catalog) |

The family elaborates the **high-confidence cross-verified
findings** from
[`../../03_video_technology/02_Response/Agent_Results/research/video-tech_cross_verification.md`](../../03_video_technology/02_Response/Agent_Results/research/video-tech_cross_verification.md)
across the chapter map.

---

## 2. Chapter map

The chapter map mirrors the dimension decomposition in
`video-tech.agent.final.md`:

| Chapter | File | Source dim | Line floor (R-01) | Status |
|---------|------|------------|------------------:|:------:|
| C25 | [`00_Index.md`](00_Index.md) (this file) | overview | 400 | landed |
| C26 | `01_Codec_Selection.md` | video dim01 | 1,250 | queued |
| C27 | `02_Hardware_Encoders.md` | video dim02 | 1,050 | queued |
| C28 | `03_Capture_Pipelines.md` | video dim03 | 1,150 | queued |
| C29 | `04_DualPath_Encoding.md` | video dim04 | 1,150 | queued |
| C30 | `05_Recording_Storage.md` | video dim05 | 1,350 | queued |
| C31 | `06_Audio_Pipeline.md` | video dim06 | 1,250 | queued |
| C32 | `07_HDR_and_Color.md` | video dim07 | 1,150 | queued |
| C33 | `08_ABR_FEC_Congestion.md` | video dim08 | 1,450 | queued |
| C34 | `09_Thermal_and_GPU_Balancing.md` | video dim09 | 1,300 | queued |
| C35 | `10_Measurement_and_QA.md` | video dim10 | 1,800 | queued |
| C36 | `11_Go_Pipeline_Implementation.md` | video dim11 | 1,600 | queued |
| C37 | `12_Network_Transport.md` | video dim12 | 1,750 | queued |

Per-dim source line counts (verified at this index landing):
dim01 = 1,151; dim02 = 934; dim03 = 1,009; dim04 = 1,013;
dim05 = 1,234; dim06 = 1,141; dim07 = 1,058; dim08 = 1,329;
dim09 = 1,181; dim10 = 1,689; dim11 = 1,466; dim12 = 1,593.
Cross-verification: 206 lines. Insight: 243 lines. Total source
material: ~14,800 lines.

---

## 3. Cross-references to Architecture chapters

The Video/Audio family elaborates surfaces named at the
architectural level by the Architecture family:

| Architecture chapter | Maps to | Content scope |
|----------------------|---------|---------------|
| C01 (`01_Streaming_Protocols_and_Codecs.md`) | C26 + C27 | Codec selection (architectural) → per-codec encoder tuning (this family) |
| C03 (`03_Host_OS_Capture.md`) | C28 | Capture primitives (DXGI / DMA-BUF / IOSurface) → capture pipeline orchestration |
| C04 (`04_Go_Client_Ecosystem.md`) | C36 | Per-platform UI (architectural) → Go pipeline implementation (this family) |
| C09 (`09_Security_and_Isolation.md`) | C30 | mTLS / DTLS for secure streaming → recording-storage encryption + audit |
| C13 (`12_Latency_Engineering_Overview.md`) §3 | C29 + C33 | Frame-time + frame pacing (architectural) → dual-path encode + ABR/FEC tuning |
| C13 §11 | C37 | Bandwidth requirements (architectural) → network-transport ladder |
| C18 (`04_GPU_Direct_and_Hardware_Pipelines.md`) | C27 | NVENC + AMF + QSV vendor capability (architectural) → encoder profile per-codec |
| C22 (`08_Frame_Pacing_and_VRR.md`) | C32 | Display-side VRR (architectural) → HDR tone-mapping client-side |
| C24 (`10_Latency_Testing_and_Validation.md`) | C35 | Latency measurement harness → video-quality measurement (VMAF / SSIM / PSNR) |

---

## 4. Cross-cutting trade-off matrix

The video + audio surface is dominated by trade-offs that cut
across multiple chapters. Each trade-off is owned by a specific
chapter; downstream chapters cite the resolution by reference.

| Trade-off | Owning chapter | Decision summary |
|-----------|----------------|------------------|
| H.264 vs HEVC vs AV1 codec selection | C26 | H.264 baseline (universal); HEVC standard (90%+ decode); AV1 premium (~25% decode 2026, ~75% by 2028); VVC out of MVP scope (Insight #8). |
| GPU vendor selection (NVENC vs AMF vs QSV) | C27 | Topology-driven (Insight #9): Intel QSV for latency-first single-host; NVIDIA NVENC for scale + tooling; AMD AMF for budget + concurrency. |
| Capture primitive (DXGI vs DMA-BUF vs IOSurface) | C28 (cross-link C03) | Per-platform: Windows = DXGI Desktop Duplication; Linux = DMA-BUF (Wayland screencopy / NVFBC); macOS = IOSurface (dev-only). |
| Single-encode vs dual-path stream+record | C29 | Dual-path enabled by default on hosts with thermal headroom (Insight #1); thermal-aware quality reduction kicks in at 78°C, hard-cap at 83°C. |
| Recording storage backend (local vs NFS vs SMB vs WebDAV) | C30 | Local NVMe SSD always (Insight #4); background sync to NFS / SMB / WebDAV / S3-compatible per operator policy. |
| Multi-channel audio passthrough (stereo vs 5.1 vs 7.1 vs Atmos) | C31 | End-to-end capability validation at session bootstrap (Insight #2); fallback ladder: Atmos → 7.1 → 5.1 → stereo. Opus MultiStream for WebRTC up to 8 channels. |
| HDR tone-mapping (host-side vs client-side) | C32 | Client-side preferred (display-aware); host-side fallback when client lacks capability. PQ + HLG + Dolby Vision metadata in RTP extensions. |
| ABR ladder design + FEC redundancy schedule | C33 | 8-tier ABR ladder: 240p → 4K120 HDR; FlexFEC RFC 8627 with 10-25% redundancy at network impairment threshold. |
| Stream + record GPU thermal budget | C34 | Insight #1 reaffirmed: dual encoding adds 15-25 W GPU power; HelixPlay's host-agent monitors GPU temp + dynamically reduces quality before throttling. |
| VMAF vs SSIM vs PSNR for QA | C35 | VMAF as primary (Netflix-developed; correlates best with MOS); SSIM as cross-check; PSNR for legacy comparison. |
| Pipeline concurrency model (goroutines vs threads) | C36 | Insight #5 binding: goroutine-per-stage with `sync.Pool` for `[]byte` frame buffers; ring-buffer channels (capacity 1-3) for back-pressure. |
| WebRTC vs custom UDP transport | C37 | Insight #7: dual transport — WebRTC for WAN/browser; custom UDP + DTLS 1.2 + SQP for LAN/native (sub-10 ms target). |

This matrix is intentionally narrow — broader trade-offs that
span multiple stream families (e.g. WebRTC vs custom UDP — C01;
NATS vs RabbitMQ — C06; YugabyteDB vs CockroachDB — C09; Compose
for TV vs Leanback — C12) are owned by the Architecture family
and not relitigated here.

---

## 5. Initial open questions for the family

Each chapter carries its own OQ list. The family-level questions
below are open at this index landing and will be resolved by the
chapter that owns each topic:

- **OQ-V00-01** — When does AV1 become the primary codec? (Hardware-decode-adoption-curve gated; current estimate 2027–2028.) Owned by C26.
- **OQ-V00-02** — Should HelixPlay support Vulkan Video encode for V1, or stay vendor-specific (NVENC / AMF / QSV) indefinitely? Owned by C27 (cross-link C18 §4.6 V1 deferral).
- **OQ-V00-03** — Recording: per-tenant storage budget enforcement — does the operator-policy posture want hard quotas (refuse new sessions if storage full) or graceful degradation (drop oldest recording)? Owned by C30.
- **OQ-V00-04** — Atmos over WebRTC — Opus MultiStream caps at 8 channels; does HelixPlay need a custom Opus extension for true Atmos object-audio (12+ channels)? Owned by C31.
- **OQ-V00-05** — Dolby Vision metadata licensing — does HelixPlay need a per-tenant Dolby license, or is HDR10/HLG sufficient for MVP? Owned by C32.
- **OQ-V00-06** — Custom UDP transport: should HelixPlay open-source the BUD-style protocol under `vasic-digital`, or keep proprietary? Owned by C37.

---

## 6. Relationship to V1 / post-MVP

The MVP scope of this family is bounded by what HelixPlay needs to
**land a quality-gated encoding + recording reference deployment
by Phase 13** (see [`../09_Implementation_Phases/Phase_13_Video_Audio.md`](../09_Implementation_Phases/Phase_13_Video_Audio.md)).
Topics deferred to V1 / post-MVP are flagged here:

- **VVC (H.266) encoding** — Insight #8 confirms hardware gap; defer to 2028+ when real-time hardware encode lands.
- **Custom GPU compute codecs** (Parsec-style intra-only) — performance-niche deferral; AV1 + HEVC cover 99% of MVP use cases.
- **Atmos object-audio over WebRTC** — capped at 8 channels in MVP; OQ-V00-04 tracks if true object-audio becomes critical.
- **Dolby Vision dynamic metadata licensing** — HDR10 + HLG + HDR10+ is MVP; Dolby Vision is V1 if a tenant requests it (per-tenant licence).
- **8K @ 60 / 120 fps streaming** — H.265 hardware encode at 8K60 viable on RTX 5090+ but consumer bandwidth uncommon; V1 deferral.

---

## 7. R-18 Operational Integrity inheritance

Every chapter in this family inherits R-18 enforcement from C08
(Constitution §2 DRY):

- **Static — chapter prose**: §1 of each chapter references R-18; §11.5 forbidden-command list is recapped where needed (chapters with subprocess-heavy contracts: C27 — `nvidia-smi`/`vainfo`/`qsv-tools`; C28 — `ffmpeg` capture invocations; C30 — recording subprocess; C34 — `nvidia-smi --query-gpu=temperature.gpu` polling; C35 — VMAF / SSIM / PSNR tooling).
- **Static — code**: imports `r18.SafeExec` from `vasic-digital/helix-r18-safeexec` (origin C08 §10). The deny-list is **not duplicated** anywhere in this family.
- **Test — `host-integrity-scan`**: inherited from C08 §12.11 verbatim into every chapter's §12 Test surface.

Family-level R-18 allow-list extension (each command wraps
through `r18.SafeExec`):

- `ffmpeg <argv>` — encoding / transcoding / capture orchestration (C27, C28, C29, C30).
- `gst-launch-1.0 <pipeline>` — GStreamer pipeline invocation (C28, C29).
- `nvidia-smi --query-gpu=<fields>` — NVIDIA GPU state polling (C27, C34).
- `vainfo` — VAAPI capability detection (C27, C28).
- `qsv-tools <args>` — Intel QuickSync capability detection (C27).
- `rocm-smi -i` — AMD ROCm GPU state polling (C27, C34).
- `obs-studio --headless --startrecording <args>` — OBS-headless recording fallback (C30, V1 only).
- `vmaf <args>` / `ffprobe -i` — video quality measurement (C35).
- `presentmon -session_name <name> -captureall` — already in family allow-list (origin C24).

No host-disruption commands appear anywhere in this family —
`ffmpeg` is allow-listed for encoding only; `kill -9 <pid>`,
`systemctl suspend|hibernate|poweroff|reboot|halt`, `pmset`,
`xset dpms force off`, `--privileged`, host-mount of `/`, `/dev`,
`/proc`, `/sys` never appear in any subprocess invocation in
this family.

---

## 8. Forward-links to other families

- **Latency family** — [`../04_Latency/10_Latency_Testing_and_Validation.md`](../04_Latency/10_Latency_Testing_and_Validation.md) (C24) hosts the canonical measurement harness; this family's C35 video-quality measurement extends it with VMAF / SSIM / PSNR.
- **Operations** — [`../08_Operations/04_Observability_and_Events.md`](../08_Operations/04_Observability_and_Events.md) consumes encoder + recording metrics emitted by chapters in this family.
- **Testing** — [`../07_Testing/02_Benchmarking_Tests.md`](../07_Testing/02_Benchmarking_Tests.md) imports the C35 measurement harness as the canonical Benchmarking-test surface for video-quality claims.
- **Implementation Phases** — [`../09_Implementation_Phases/Phase_13_Video_Audio.md`](../09_Implementation_Phases/Phase_13_Video_Audio.md) is the primary implementation phase for the family.
- **Submodules + Containers** — every chapter's §6.1 Submodule boundaries adds to the public submodule inventory under `vasic-digital/`. Expected new public submodules from this family: `vasic-digital/helix-codec` (codec abstraction over NVENC/AMF/QSV), `vasic-digital/helix-capture` (DXGI/DMA-BUF/IOSurface façade), `vasic-digital/helix-record` (dual-path recording with local-buffer + background-sync), `vasic-digital/helix-audio` (Opus MultiStream + multi-channel pipeline), `vasic-digital/helix-hdr` (PQ/HLG/HDR10/HDR10+/Dolby Vision), `vasic-digital/helix-abr` (adaptive bitrate ladder + FlexFEC), `vasic-digital/helix-thermal` (GPU thermal-aware quality controller), `vasic-digital/helix-vqa` (VMAF/SSIM/PSNR harness), `vasic-digital/helix-pipeline` (Go pipeline orchestration), `vasic-digital/helix-transport` (custom UDP + DTLS 1.2 + SQP). All ten reuse `vasic-digital/helix-r18-safeexec` (DRY).

---

## 9. Cross-stream insights cited inside this family

For traceability, the cross-stream insights from prior families that
this family's chapters cite are catalogued here (each chapter's
header preamble names the specific insights it consumes; the listing
is for orchestrator-level navigation):

| Cross-stream insight | Origin family | Cited by Video/Audio chapters |
|----------------------|---------------|-------------------------------|
| **Insight #1 (cloudgaming)** — Sunshine++ pattern | Architecture | C28 (Sunshine-style capture process), C29 (Sunshine multi-session removal context), C30 (Sunshine recording fork) |
| **Insight #5 (cloudgaming)** — Anti-cheat clean host | Architecture | C28 (capture process under anti-cheat constraints) |
| **Insight #7 (cloudgaming)** — Edge > Codec for latency | Architecture | C26, C27, C33, C37 (codec choices are necessary but insufficient without edge proximity) |
| **Insight #1 (latency)** — Microwave Pipeline | Latency | C28, C29, C36 (capture→encode→network is the pipeline this family elaborates) |
| **Insight #2 (latency)** — p999 only metric | Latency | C35 (binding for video-quality + latency reporting) |
| **Insight #3 (latency)** — Asymmetric optimisation | Latency | C32 (host HDR vs client tone-mapping) |
| **Insight #4 (latency)** — Allocation-free hot path | Latency | C36 (`sync.Pool` for `[]byte` frame buffers) |
| **Insight #5 (latency)** — Conservative Prediction Paradox | Latency | C33 (ABR prediction conservatism — over-bitrate over under-bitrate) |

---

## 10. Codec ladder + bandwidth budgets at a glance

For orchestrator-level navigation, the codec ladder + bandwidth
budgets that the family elaborates in C26 + C27 + C33 are
summarised here:

| Codec | Container | Hardware encode 2026 | Bandwidth @ 1080p60 | Bandwidth @ 4K60 | Decode coverage 2026 |
|-------|-----------|----------------------|---------------------:|-----------------:|----------------------:|
| H.264 | MP4 / MKV | NVENC 8th gen + AMF + QSV (universal) | 6–10 Mbps | 25–40 Mbps | 98%+ (mandatory WebRTC) |
| HEVC | MP4 (HEVC profile) | NVENC + AMF + QSV (since Skylake / Polaris) | 4–7 Mbps | 15–25 Mbps | ~90% (Apple, modern Android, Windows + macOS native) |
| AV1 | MP4 (AV1 profile) / WebM | NVENC 8th-gen Lovelace + AMD RDNA3+ + Intel Arc Battlemage + Apple M3+ Pro/Max | 3–5 Mbps | 10–18 Mbps | ~25% (Chrome 100+, Firefox 113+, modern HW; **NOT** universal) |
| VVC | MP4 (VVC profile) | **NONE** (real-time hardware encode not viable < 2028) | n/a | n/a | < 5% (no browser support) |

HelixPlay's MVP codec ladder is **H.264 (universal) + HEVC
(standard tier) + AV1 (premium tier where client supports)**.
VVC is V1 deferral only.

For audio, the channel-config ladder is:

| Audio config | Codec / format | Channels | WebRTC support | Native HelixPlay support |
|--------------|----------------|----------|:--------------:|:------------------------:|
| Stereo PCM | Opus / PCM | 2 | yes | yes (universal) |
| 5.1 multi-channel | Opus MultiStream | 6 | yes (Opus MS) | yes |
| 7.1 multi-channel | Opus MultiStream | 8 | yes (Opus MS) | yes |
| Atmos (object-audio) | Dolby Digital Plus + JOC / E-AC3 | 12+ | NO native | passthrough only (capability-validated end-to-end per Insight #2) |

Insight #2 binding: HelixPlay's session bootstrap validates the
**entire audio chain** (game audio API → OS stack → capture →
encode → transmit → decode → AV receiver) before negotiating
non-stereo modes. ANY link in the chain that lacks the capability
forces fallback to the next-lower tier.

---

## 11. Recording storage backend matrix (Insight #4 elaboration)

Per Insight #4, recording follows a **local-buffer + background-
sync** pattern. The supported backend matrix is:

| Backend | Local-buffer staging | Background sync target | Use case |
|---------|---------------------|-------------------------|----------|
| Local NVMe SSD | yes (canonical primary) | n/a | Always; primary recording target |
| SMB / CIFS | yes (NVMe stage) | SMB v3 / 3.1.1 share | Home NAS (Synology, QNAP, TrueNAS) |
| NFS v4 | yes (NVMe stage) | NFS v4.1 mount | Linux-tier home / enterprise NAS |
| WebDAV | yes (NVMe stage) | WebDAV PUT (HTTPS) | Nextcloud, Owncloud, Apache mod_dav |
| FTP / FTPS | yes (NVMe stage) | FTP STOR (FTPS preferred) | Legacy enterprise |
| S3-compatible | yes (NVMe stage) | s3 PutObject (multi-part) | Backblaze B2, MinIO, Wasabi, AWS S3 |
| iCloud Drive | yes (NVMe stage) | macOS-tier only (CloudKit) | macOS dev-tier; not production |

C30 is the canonical chapter; this matrix is for orchestrator
navigation only.

---

## 12. Per-chapter topic preview

For orchestrator-level navigation, a short preview of each
chapter's scope is given here. Full prose is in the chapter file
itself (queued under R1 dispatch).

### C26 — Codec Selection (`01_Codec_Selection.md`, dim01, ≥1,250 floor)
H.264 / HEVC / AV1 / VVC selection logic. Insight #3 (H.264
sweet spot) + Insight #8 (AV1 the correct near-term bet; VVC
deferred to 2028+). Per-codec encoder profiles, Slice-based
parallelism, B-frame configuration, GOP structure, intra-refresh
patterns. Capability-negotiation between host and client.

### C27 — Hardware Encoders (`02_Hardware_Encoders.md`, dim02, ≥1,050 floor)
NVENC 8th-gen Lovelace + 9th-gen Blackwell + AMD AMF on RDNA3+RDNA4 +
Intel QSV on Arc Battlemage + Apple VideoToolbox. Insight #9 (vendor
selection topology-driven). Per-vendor latency profiles, session
limits, AV1 hardware availability, B-frame support.

### C28 — Capture Pipelines (`03_Capture_Pipelines.md`, dim03, ≥1,150 floor)
Capture-side pipeline orchestration over the C03-owned primitives
(DXGI / DMA-BUF / IOSurface). Sunshine++ pattern (Insight #1
cloudgaming). Capture-process isolation, frame-event publishing.

### C29 — Dual-Path Encoding (`04_DualPath_Encoding.md`, dim04, ≥1,150 floor)
Stream + record dual-path encoding. Insight #1 (thermal wall);
GPU thermal headroom budget. NVENC dual-session orchestration.
Frame-Tee → stream-encoder + record-encoder; back-pressure handling.

### C30 — Recording Storage (`05_Recording_Storage.md`, dim05, ≥1,350 floor)
Local-buffer + background-sync pattern (Insight #4). Recording-
backend matrix per §11. Storage encryption, audit, retention. fMP4
+ MKV crash-safe containers. Instant-replay circular buffer.

### C31 — Audio Pipeline (`06_Audio_Pipeline.md`, dim06, ≥1,250 floor)
Multi-channel passthrough (Insight #2). Stereo / 5.1 / 7.1 / Atmos
fallback ladder. Opus MultiStream up to 8 channels. eARC + SPDIF
+ HDMI passthrough. Channel-order mismatch (Windows vs Dolby).

### C32 — HDR & Color (`07_HDR_and_Color.md`, dim07, ≥1,150 floor)
HDR10 + HLG + HDR10+ + Dolby Vision metadata. PQ vs HLG transfer
functions. RTP extension carriage. Client-side tone-mapping
preferred (Insight #3 asymmetric — host vs client). Display-
metadata negotiation.

### C33 — ABR + FEC + Congestion (`08_ABR_FEC_Congestion.md`, dim08, ≥1,450 floor)
8-tier adaptive-bitrate ladder (240p → 4K120 HDR). FlexFEC RFC
8627 with 10–25% redundancy. SQP congestion control (cross-link
C19 §3 + Insight #7). GCC vs SCReAM vs SQP comparison.

### C34 — Thermal & GPU Balancing (`09_Thermal_and_GPU_Balancing.md`, dim09, ≥1,300 floor)
Insight #1 (thermal wall) elaborated. GPU temp / power / clock
polling via `nvidia-smi --query-gpu`. Pre-emptive quality
reduction at 78°C; hard cap at 83°C. Multi-GPU host load-balancer
with composite scoring.

### C35 — Measurement & QA (`10_Measurement_and_QA.md`, dim10, ≥1,800 floor)
VMAF (Netflix) + SSIM + PSNR. Cross-link C24 measurement harness
for latency side. Per-frame quality logging, regression-detection
gate in CI. MOS correlation. Reference-vs-distorted comparison
with synthetic streams.

### C36 — Go Pipeline Implementation (`11_Go_Pipeline_Implementation.md`, dim11, ≥1,600 floor)
Insight #5 — goroutine-per-pipeline-stage with `sync.Pool` for
`[]byte` frame buffers. Ring-buffer channels (capacity 1–3) for
back-pressure. Producer-consumer pattern across capture →
encode → packetize → transmit stages.

### C37 — Network Transport (`12_Network_Transport.md`, dim12, ≥1,750 floor)
WebRTC for WAN/browser; custom UDP + DTLS 1.2 + SQP (Parsec BUD-
style) for LAN/native (Insight #7). Cross-link C19 + C01. RTP
sequence-number management, FIR/PLI re-keyframe negotiation,
NACK loss recovery. RTCP report intervals.

---

## 13. Anti-Bluff posture (R-13) for this family

Every chapter in this family carries an `## Anti-Bluff Verification`
block per Master Plan §4.3. The blocks list:

- The exact source files reviewed (path + line count + reviewer + date + sections used).
- The web-research addendum URLs with cluster table.
- The video-tech insights and HCs incorporated.
- The conflict zones resolved (own + inherited).
- R-18 compliance evidence (chapter prose + code + tests).
- Coverage confirmation (line counts vs floors, forbidden-pattern scan, table inventory, code-block inventory).
- Sign-off rows per section subagent + orchestrator.

Forbidden-pattern scan: `TODO`, `FIXME`, `XXX`, `HACK`, `tbd`,
`???`, `placeholder`, "and similar", "etc.", "as appropriate",
"as needed", "where reasonable", "fill in later". Self-referential
mentions inside Constitution-cite text are explicitly permitted by
Constitution §1.1 and Master Plan §5.2.3.

The Constitution-§6 reporting contract (p50 / p99 / p999 with
≥ 10 K samples + 95% CI) applies to **every video-quality and
latency claim** in this family, enforced via the C24 measurement
harness (cross-link).

---

End of `05_Video_Audio/00_Index.md` — 2026-04-29.
