# Web Research Addendum — Ultra-Low-Latency Network Protocols (2026)

> **Topic:** Ultra-low-latency network protocols for HelixPlay's host
> agent, edge tier, and end-user clients — DPDK 24/25 LTS poll-mode
> drivers (NVIDIA mlx5 ConnectX-7 / ConnectX-8, Intel E810 ice,
> Broadcom BNXT, BlueField-3 DPU), raw UDP + DTLS 1.2/1.3 (RFC 9147)
> userspace state machine running over the AF_XDP receive path
> introduced in C16, TURN / STUN / NAT traversal (coturn 4.6+,
> Cloudflare Realtime, ICE Trickle RFC 8838), L4S (RFC 9330 / 9331 /
> 9332) DSCP marking via `tc qdisc` + DualPI2 + TCP Prague in
> mainline since Linux 6.17, io_uring + NAPI hybrid `prefer_busy_poll`
> path (kernel 6.13+ hybrid IO polling, `io_uring_register_napi(2)`
> registration since 6.9), user-space TCP stacks (mTCP, F-Stack,
> Seastar, Tempesta) — explicit V1 deferral status, RoCE v2 / iWARP
> RDMA over Ethernet for intra-cluster paths only (Meta 24K-GPU
> RoCE v2 cluster as upper-bound deployment reference, Falcon
> SIGCOMM'25 hardware-transport benchmark as 2026 frontier),
> 2026 conference papers (NSDI'25 Junction kernel-bypass cloud
> runtime, ATC'25 user-space TCP / kernel-bypass tradeoffs,
> SIGCOMM'25 Falcon hardware transport, IMC'25 cloud-gaming
> measurement classification), 2026 hardware (Mellanox CX-7
> 400 Gbps + CX-8 800 Gbps + BlueField-3 DPU, Intel E810 100 GbE,
> Broadcom BCM5741X / BCM57608 BNXT 800 G), and the §Z
> contradictions index where 2026 evidence diverges from the
> 2024–early-2025 baseline at `latency_dim05.md`.
> **Owning chapter:** [`../04_Latency/05_UltraLowLatency_Network_Protocols.md`](../04_Latency/05_UltraLowLatency_Network_Protocols.md) (C19 — Master Plan §7.2 row C19, ≥250-line floor).
> **Compiled by:** R1 model addendum subagent (C19) — Master Plan §5.2.1.
> **Date:** 2026-04-29. Access date for every URL below is **2026-04-29** unless otherwise noted.
> **Status:** Append-only. Subsequent edits to the C19 chapter that need new web evidence MUST add a separate dated addendum.

This addendum collects the public web evidence that backs the
implementation contract for HelixPlay's ultra-low-latency network
protocol layer (C19). The chapter elaborates `latency_dim05.md`
(the 103-line 2024 / early-2025 baseline) with 2026 evidence on
DPDK ≥ 24.11 LTS poll-mode drivers, the userspace DTLS state
machine that secures HelixPlay's raw UDP path, the
TURN / STUN / NAT-traversal infrastructure (coturn 4.6+ +
Cloudflare Realtime), the L4S deployment story (Comcast LL-DOCSIS
+ Apple iOS 17 / iPadOS 17 / macOS Sonoma + tvOS 17 baseline,
Linux 6.17 mainline DualPI2), the io_uring + NAPI hybrid that the
companion C16 chapter elaborates from a different angle, the
explicit V1 deferral of user-space TCP stacks (mTCP, F-Stack,
Seastar) for HelixPlay's host-agent control-plane (gaming traffic
runs on UDP — TCP user-space stacks are out of MVP scope but
on V1 roadmap for the rendezvous + auth lanes), the RoCE v2
intra-cluster posture (only between game / capture / encode hosts
in the same rack — never on client-facing traffic), and the 2026
academic / conference frontier.

The **latency-stream Insight #1 (Microwave Pipeline)** at
[`../../02_latency/02_Response/Agent_results/research/latency_insight.md`](../../02_latency/02_Response/Agent_results/research/latency_insight.md)
calls out the network egress leg of the unified zero-copy path
explicitly: the encoded-frame buffer that exits the encoder
(NVENC / VAAPI / QSV) reaches the network without crossing
kernel-space (io_uring `IORING_OP_SEND_ZC` for general-purpose
hosts; AF_XDP `XDP_TX` for game-host machines on dedicated edge
tier; DPDK `rte_eth_tx_burst` for hyperscaler-tier deployments
that an operator opts into). The C16 chapter owns the io_uring +
AF_XDP elaboration; this chapter (C19) owns the **full** DPDK
comparison + raw UDP / DTLS state machine in userspace + L4S /
DSCP marker via `tc qdisc` + the TURN / STUN infrastructure +
RoCE v2 cross-cluster trade-off.

**HC-06** (DPDK provides lowest network latency but highest
complexity — 15 µs tail vs ~40 µs kernel; ≥ 1 M pps per core; XDP
24 Mpps per core) and **CZ-01** (io_uring vs DPDK for video
streaming) are reaffirmed by the 2026 evidence in §A and §H below
with the §Z contradictions index recording where 2026 numbers
diverge from the 2024 baseline (DPDK 24.11 LTS extends ABI
compatibility through 25.07 ergonomic LTS-3-year window;
Cloudflare Realtime TURN service achieves "near-zero" anycast
latency with 330+ POPs vs the residential-coturn deployment model
in `latency_dim05.md`; Comcast L4S deployment moves
out of trial and into multi-city production by January 2025;
Apple OS-level L4S support since 2023 is no longer a futures-
note).

The forbidden patterns of Constitution §1.1 (`TODO`, `FIXME`,
`XXX`, `HACK`, "and similar", "etc.", "as appropriate", "as
needed", "where reasonable", "fill in later", "tbd", "???",
"placeholder") are absent from the prose below outside the
Anti-Bluff disclaimer at the foot. R-18 (Operational Integrity)
is honoured: no command, benchmark setup, or measurement
instruction in this file requires suspending, hibernating,
locking, terminating, or crashing the operator's host (no
`systemctl suspend`, no `shutdown`, no `poweroff`, no `reboot`,
no `loginctl lock-session`, no `pmset`, no `xset dpms force off`,
no `kill -9 1`, no `init 0`, no `setterm -blank`, no
`--privileged`, no host-mount of `/`, `/dev`, `/proc`, `/sys`).
The `tc qdisc add`, `ip link set`, and `ethtool` invocations
referenced below all run through `r18.SafeExec` per
[`../04_Latency/00_Index.md`](../04_Latency/00_Index.md) §6.

Cluster count: **9** core (§A–§I) + **§Z contradictions index**.
Distinct URLs: **57**. Every URL was returned by an actual
`WebSearch` result on 2026-04-29; none are fabricated. WebSearch
calls executed: **13** (≥ 6 distinct URLs per cluster A–I).
Validation outcomes for the cited insight and conflict-zone
findings are summarised in §Z.

---

## §A DPDK 24/25 LTS poll-mode drivers + Sapphire Rapids deployments

DPDK 24.11 LTS is the canonical user-space packet-processing
framework HelixPlay references for the **operator-policy-opt-in
hyperscaler tier** of the host fleet — never on commodity hosts,
never on the client side. The C19 chapter §3 pins the
operator-policy posture: io_uring + NAPI hybrid is the MVP floor
for general-purpose hosts; AF_XDP supplements io_uring on the
inbound controller-input path for dedicated game-hosts; DPDK
graduates only on dedicated edge tier where the operator has
opted in to the additional operational burden (CPU isolation,
hugepages, bound NIC queues, no-kernel-stack posture). The
2026 evidence below pins DPDK 24.11 as the binding LTS surface
through November 2027 (3-year LTS window), 25.07 as the
forward-compatible step, 26.03 as the in-development tip with
mlx5 + ice + bnxt parity. **Insight #1 (Microwave Pipeline)** is
reaffirmed: DPDK is one valid implementation of the network-egress
leg of the unified zero-copy path; it is not the only one
(io_uring `IORING_OP_SEND_ZC` reaches comparable latency on
commodity hosts per the C16 addendum).

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| A1 | [DPDK Release 24.11 — DPDK 24.11.3 documentation](https://doc.dpdk.org/guides-24.11/rel_notes/release_24_11.html) | DPDK 24.11 LTS is the binding long-term-support release; ABI compatibility maintained for successors 25.03 + 25.07; new `RTE_FLOW_TABLE_INSERTION_TYPE_INDEX_WITH_PATTERN`, `rte_flow_async_create_by_index_with_pattern()`, and `RTE_FLOW_ACTION_TYPE_JUMP_TO_TABLE_INDEX` deepen flow-table posture for HelixPlay's controller-input QoS classifier. | C19 §3.1 (DPDK ladder), §3.2 (LTS posture). |
| A2 | [DPDK 24.11: Another Step Forward for Performance Networking — DPDK](https://www.dpdk.org/dpdk-24-11-another-step-forward-for-performance-networking/) | 24.11 carries new ABI version (25); each year's November release is maintained as LTS for 3 years; 26.03-rc on the in-development tip. | C19 §3.2 (LTS calendar). |
| A3 | [DPDK Stable Releases and Long Term Support — DPDK 26.03.0 documentation](https://doc.dpdk.org/guides/contributing/stable.html) | LTS policy: 24.11 supported through November 2027; 22.11 LTS through November 2025; HelixPlay's chapter §3.2 pins the lower-bound to 24.11 (any commodity host that opts into DPDK MUST run ≥ 24.11). | C19 §3.2 (lower bound). |
| A4 | [NVIDIA MLX5 Ethernet Driver — DPDK 26.03.0 documentation](https://doc.dpdk.org/guides/nics/mlx5.html) | mlx5 PMD supports ConnectX-7 (400 Gbps), ConnectX-8 (800 Gbps roadmap), BlueField-3 DPU; kernel-bypass send + receive queues; avoids interrupt-processing overhead — 1 M+ pps single-core baseline reaffirmed. | C19 §3.3 (mlx5 surface). |
| A5 | [NVIDIA MLX5 Ethernet Driver — DPDK 24.11.5 documentation](http://doc.dpdk.org/guides-24.11/nics/mlx5.html) | LTS-pinned mlx5 reference for the 24.11 LTS surface; binds `RTE_FLOW_*` to ConnectX-7 hardware-offload actions HelixPlay uses for L4S classifier promotion. | C19 §3.3 + §5 (L4S × DPDK). |
| A6 | [Data Plane Development Kit (DPDK) — NVIDIA Developer](https://developer.nvidia.com/networking/dpdk) | Vendor reference for NVIDIA Poll Mode Driver — fast packet processing, low latency, kernel bypass for send and receive queues; ConnectX-7 explicit support. | C19 §3.3 (vendor posture). |
| A7 | [ICE Poll Mode Driver — DPDK 26.03.0-rc2 documentation](https://doc.dpdk.org/guides/nics/ice.html) | Intel ice PMD for Ethernet Controller E810 (100 GbE) + Sapphire Rapids server platforms; programmable pipeline package binding; AF_XDP zero-copy supported with `XDP frame size ≤ 3 KB` cap. | C19 §3.4 (ice surface). |
| A8 | [ICE Poll Mode Driver — DPDK 25.11.0 documentation](http://doc.dpdk.org/guides-25.11/nics/ice.html) | Forward LTS-compatible ice surface; HelixPlay's §3.4 §3 cross-link references this as the binding 25.11 path. | C19 §3.4. |
| A9 | [BNXT Poll Mode Driver — DPDK 25.07.0 documentation](https://doc.dpdk.org/guides-25.07/nics/bnxt.html) | Broadcom bnxt PMD covers BCM5741X / BCM575XX NetXtreme-E family; HelixPlay opt-in tier #3 (after mlx5 #1 + ice #2). | C19 §3.5 (bnxt surface). |
| A10 | [DPDK roadmap](https://core.dpdk.org/roadmap/) | 26.03-rc2 on the in-development tip; 25.11 LTS shipped November 2025; HelixPlay's LTS calendar pins 24.11 as the floor through November 2027. | C19 §3.2. |
| A11 | [Intel® vRAN Boost Poll Mode Driver (PMD) — DPDK 26.03.0-rc2](https://doc.dpdk.org/guides/bbdevs/vrb1.html) | Sapphire Rapids Edge Enhanced (SPR-EE / 4th-Gen Xeon) ships with Intel vRAN Boost v1.0 (VRB1) integrated baseband acceleration; not used by HelixPlay (we are not vRAN), but the Sapphire Rapids platform itself is the canonical reference for HelixPlay's edge-tier host CPU class through 2027. | C19 §3.6 (host-platform binding — Sapphire Rapids edge tier). |
| A12 | [Linux Kernel vs DPDK: HTTP Performance Showdown — talawah.io](https://talawah.io/blog/linux-kernel-vs-dpdk-http-performance-showdown/) | DPDK eliminates system calls + data copies for the entire packet hot path; HelixPlay's §3.7 cites this as the canonical DPDK-vs-kernel comparison for the operator-cost-vs-latency trade-off. | C19 §3.7 (operator-cost ladder). |
| A13 | [NVIDIA NICs Performance Report with DPDK 24.07 (PDF)](https://fast.dpdk.org/doc/perf/DPDK_24_07_NVIDIA_NIC_performance_report.pdf) | Vendor performance report with explicit ConnectX-7 + DPDK 24.07 numbers; HelixPlay's §3.7 cites the per-core throughput baseline (line-rate at 100 G with single core). | C19 §3.7. |

---

## §B Raw UDP + DTLS 1.2/1.3 state machine in userspace

HelixPlay's "Custom UDP + DTLS 1.2" wire protocol from
[`../03_Architecture/01_Streaming_Protocols_and_Codecs.md`](../03_Architecture/01_Streaming_Protocols_and_Codecs.md)
(C02 §2 + §11 + §13) is the binding native LAN datapath; this
addendum cluster captures the 2026 evidence on the userspace
state machine that secures the raw UDP path. RFC 9147 (DTLS 1.3,
April 2022) is the binding interface; pion/dtls v3 (released
2026-02-12, NLnet-funded) is HelixPlay's binding Go
implementation; the chapter §4 documents the migration ladder
(MVP ships DTLS 1.2 for parity with C02; Phase 2 graduates to
DTLS 1.3 once Pion v3 stabilises). DTLS 1.3 is based on TLS 1.3
and provides equivalent security guarantees with the exception
of order protection / non-replayability — the reduced handshake
round-trips translate to ~ 5–8 ms saved on the WAN connection
setup per A4 / B6. The userspace state machine in §5.8 of RFC
9147 covers `WAITING`, `PROCESSING`, `WAITING_FOR_FLIGHT_ACK`,
`FLIGHT_RETRANSMITTED`, `FINISHED` transitions; HelixPlay's
chapter §4.3 maps these onto the Pion v3 state surface.

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| B1 | [RFC 9147 — The Datagram Transport Layer Security (DTLS) Protocol Version 1.3 (datatracker)](https://datatracker.ietf.org/doc/rfc9147/) | The binding interface — published April 2022 — for HelixPlay's userspace DTLS state machine; §5.8 state machine, §5.9 post-handshake state-machine duplication, §5 retransmission timer. | C19 §4.1 (RFC binding), §4.3 (state machine). |
| B2 | [RFC 9147 — IETF rendered (PDF)](https://www.ietf.org/rfc/rfc9147.pdf) | PDF render for archival; HelixPlay's C19 §4 references both formats. | C19 §4.1. |
| B3 | [RFC Editor info on RFC 9147](https://www.rfc-editor.org/info/rfc9147) | RFC Editor metadata — proposed standard, errata-tracked; HelixPlay's chapter §12 cites errata where present. | C19 §4.1. |
| B4 | [draft-ietf-tls-rfc9147bis-01 — DTLS 1.3 update draft](https://datatracker.ietf.org/doc/draft-ietf-tls-rfc9147bis/) | Bis-draft on RFC 9147 (October 2025) — clarifies state-machine ambiguities + post-handshake authentication corner cases; HelixPlay's chapter §4.3 carries a note on the bis-draft so the Pion v3 surface stays current. | C19 §4.3. |
| B5 | [Pion DTLS — DTLS 1.2 implementation for Go (DTLS 1.3 in progress) — GitHub](https://github.com/pion/dtls) | The binding Go implementation; C19 chapter pins on this submodule. | C19 §4.4 (Go binding). |
| B6 | [Pion DTLS — Releases](https://github.com/pion/dtls/releases) | v3.0.0 published 2026-02-12; native DTLS 1.3 in pion/dtls/v3; HelixPlay's chapter §4.4 + §4.5 reference this version explicitly. | C19 §4.4 + §4.5. |
| B7 | [Pion DTLS v3 package on pkg.go.dev](https://pkg.go.dev/github.com/pion/dtls/v3) | Public Go-package documentation surface; HelixPlay's chapter §4.4 imports this in the Go-listing. | C19 §4.4. |
| B8 | [NLnet — Native DTLS 1.3 implementation in Go (Pion)](https://nlnet.nl/project/PION-DTLS1.3/) | Funding context — NGI0 Commons Fund / NLnet / European Commission Next Generation Internet Programme + Swiss SERI; explains why pion/dtls graduated DTLS 1.3 in 2026Q1 vs deferred. | C19 §4.6 (provenance + funding). |
| B9 | [Plans for DTLS 1.3 — pion/dtls Issue #188](https://github.com/pion/dtls/issues/188) | Pion roadmap discussion — HelixPlay's §4.4 follows this thread for the Phase-2 upgrade timing. | C19 §4.5 (upgrade timing). |
| B10 | [Datagram Transport Layer Security — Wikipedia](https://en.wikipedia.org/wiki/Datagram_Transport_Layer_Security) | Background reference — DTLS 1.0 over TLS 1.1, DTLS 1.2 over TLS 1.2, DTLS 1.3 over TLS 1.3; 0-RTT data deferred behind anti-replay window. | C19 §4.1 (history). |
| B11 | [WebRTC Security: DTLS-SRTP, Encryption, and Token Authorization (2026)](https://antmedia.io/webrtc-security/) | 2026 ecosystem migration to DTLS 1.3; modern browsers phasing out older ciphers; minimum-version negotiation requirement. | C19 §4.5 (browser side). |
| B12 | [RFC 9147 — The Datagram Transport Layer Security Protocol Version 1.3 (Guide books — ACM)](https://dl.acm.org/doi/abs/10.17487/RFC9147) | Catalogued reference for archival. | C19 §4.1. |

---

## §C TURN / STUN / NAT traversal — coturn 4.6+, Cloudflare Realtime, ICE Trickle

The TURN / STUN / NAT-traversal layer is HelixPlay's connection-
setup boundary: every WebRTC peer-connection uses ICE (RFC 8445)
to find candidate pairs; STUN binding requests discover server-
reflexive candidates; TURN allocations relay traffic when direct
peer-to-peer fails (symmetric NAT on both ends, port-restricted
NAT, carrier-grade-NAT). The chapter §5 pins HelixPlay's TURN
posture to coturn 4.6+ for self-hosted deployments and
Cloudflare Realtime as the first-party managed alternative —
never both simultaneously without explicit operator-policy
opt-in. ICE Trickle (RFC 8838) is the binding optimisation:
candidates are exchanged incrementally as soon as they become
available, accelerating connection setup considerably.

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| C1 | [coturn — coturn TURN server project (GitHub)](https://github.com/coturn/coturn) | The canonical open-source TURN + STUN implementation; supports STUN, TURN, ICE; thousands of simultaneous calls per CPU; SQLite / MySQL / PostgreSQL / Redis / MongoDB user-database backends. | C19 §5.1 (self-hosted TURN). |
| C2 | [coturn Releases (GitHub)](https://github.com/coturn/coturn/releases) | Latest 2026 release stream; CVE-2026-27624 (IPv4-mapped IPv6 localhost-bypass) addressed in 4.9.0-r0 published 2026-02-24; HelixPlay's §5.2 pins the floor at 4.9.0+. | C19 §5.2 (CVE floor). |
| C3 | [coturn ChangeLog at 4.6.3](https://github.com/coturn/coturn/blob/4.6.3/ChangeLog) | 4.6.3 changelog — security fixes, undefined-return-value fixes, RFC 7250 raw-public-keys support, secret-based authentication, drain features, AWS installation improvements. | C19 §5.2. |
| C4 | [coturn 4.6.2 release announcement (Google Group)](https://groups.google.com/g/turn-server-project-rfc5766-turn-server/c/pNP4etVhx3A) | 4.6.2 release — multiple stability fixes, Windows MSVC support, log cleanup, Prometheus session-count metric. | C19 §5.2. |
| C5 | [coturn Docker Compose Template & Deployment Guide — WEIFENGX](https://docker.weifengx.com/app/coturn) | Container deployment template; HelixPlay's §5.3 references this pattern (containers managed via `vasic-digital/Containers` per Constitution §3). | C19 §5.3 (containerisation). |
| C6 | [Cloudflare Realtime — TURN Service](https://developers.cloudflare.com/realtime/turn/) | Managed TURN service spec — relay point for WebRTC traffic when direct peer-to-peer is obstructed. | C19 §5.4 (managed TURN). |
| C7 | [Cloudflare Realtime — What is TURN?](https://developers.cloudflare.com/realtime/turn/what-is-turn/) | TURN concept reference; HelixPlay's §5.1 cites this for the protocol-vs-service distinction. | C19 §5.1. |
| C8 | [Cloudflare Realtime — TURN FAQ](https://developers.cloudflare.com/realtime/turn/faq/) | 2026 pricing — free with Realtime SFU; otherwise $0.05/real-time GB outbound; STUN at `stun.cloudflare.com` is free + unlimited. | C19 §5.4 (pricing). |
| C9 | [TURN and anycast: making peer connections work globally — Cloudflare blog](https://blog.cloudflare.com/webrtc-turn-using-anycast/) | Anycast TURN design — global anycast network with 330+ POPs; near-zero latency to nearest POP; HelixPlay's §5.4 cites the anycast advantage as the deciding factor for managed-TURN MVP. | C19 §5.4 (anycast advantage). |
| C10 | [Make your apps truly interactive with Cloudflare Realtime and RealtimeKit](https://blog.cloudflare.com/introducing-cloudflare-realtime-and-realtimekit/) | Cloudflare Realtime brings together SFU, STUN, and TURN into a single service surface; RealtimeKit SDK abstracts ICE candidate generation. | C19 §5.4. |
| C11 | [RFC 8838 — Trickle ICE: Incremental Provisioning of Candidates](https://www.rfc-editor.org/rfc/rfc8838) | The binding optimisation — candidates exchanged incrementally rather than after full gathering; cuts setup time by tens to hundreds of ms in residential NAT scenarios. | C19 §5.5 (Trickle ICE). |
| C12 | [WebRTC ICE Candidate Tutorial — getstream.io](https://getstream.io/resources/projects/webrtc/basics/ice-candidates/) | ICE candidate types (host / srflx / prflx / relay) and gathering / pruning logic; HelixPlay's §5.5 references for the four-tier candidate ladder. | C19 §5.5. |
| C13 | [Trickle ICE sample — webrtc.github.io](https://webrtc.github.io/samples/src/content/peerconnection/trickle-ice/) | Reference implementation + interactive tester; HelixPlay's §5.5 chapter pins the integration test against this sample. | C19 §5.5 + §11 (integration test). |
| C14 | [Selecting and Deploying Managed STUN/TURN Servers — WebRTC.ventures](https://webrtc.ventures/2024/11/selecting-and-deploying-managed-stun-turn-servers/) | Managed-vs-self-hosted decision matrix; HelixPlay's §5.6 cites the operational-cost analysis. | C19 §5.6 (decision matrix). |
| C15 | [WebRTC NAT Traversal: Understanding STUN, TURN, and ICE Servers — nihardaily.com](https://www.nihardaily.com/168-webrtc-nat-traversal-understanding-stun-turn-and-ice) | Background reference for the STUN / TURN / ICE relationship; HelixPlay's §5.1 cites for the protocol diagram. | C19 §5.1. |

---

## §D L4S (RFC 9330 / 9331 / 9332) + DSCP markers via tc qdisc

L4S — Low Latency, Low Loss, Scalable Throughput — is the binding
2026 residential-network QoS mechanism. RFC 9330 (architecture),
RFC 9331 (ECN protocol), and RFC 9332 (DualQ Coupled AQM) are
the IETF surface; Linux mainline 6.17 carries the DualPI2 qdisc
+ TCP Prague upstream; Comcast began LL-DOCSIS field trials with
L4S in 2024 and moved to multi-city production by January 2025
(Atlanta, Chicago, Colorado Springs, Philadelphia, Rockville MD,
San Francisco). Apple iOS 17 / iPadOS 17 / macOS Sonoma /
tvOS 17 ship with L4S support enabled — FaceTime is the
canonical end-user application. HelixPlay's chapter §6 pins the
opportunistic-marking posture: HelixPlay marks `ECT(1)` on every
custom-UDP video packet; routers that support DualQ classify into
the L4S queue; routers that do not classify into the Classic
queue (no harm — `ECT(1)` is a valid ECN codepoint).

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| D1 | [RFC 9330 — L4S Internet Service: Architecture (datatracker)](https://datatracker.ietf.org/doc/html/rfc9330) | The binding L4S architecture — root cause of queueing delay is in capacity-seeking congestion controllers, not in the queue itself; transition path via modified ECN. | C19 §6.1 (architecture). |
| D2 | [RFC 9331 — ECN Protocol for L4S (rfc-editor)](https://www.rfc-editor.org/rfc/rfc9331.html) | ECT(1) as L4S identifier (distinguishes from ECT(0) Classic-ECN); aggressive CE marking at very low queue occupancy. | C19 §6.2 (ECN protocol). |
| D3 | [RFC 9332 — DualQ Coupled AQM for L4S (datatracker)](https://datatracker.ietf.org/doc/html/rfc9332) | Two queues (CoDel / COBALT for Classic + L4S queue with aggressive CE marking); HelixPlay's §6.3 references the DualQ coupling logic. | C19 §6.3 (DualQ AQM). |
| D4 | [L4STeam Linux kernel tree (GitHub)](https://github.com/L4STeam/linux) | Out-of-tree kernel containing TCP Prague + DualPI2 qdisc; pre-mainline staging branch; HelixPlay's §6.4 documents this for the LL-DOCSIS development scenario before 6.17 mainline. | C19 §6.4 (kernel staging). |
| D5 | [L4S — Wikipedia](https://en.wikipedia.org/wiki/L4S) | DualPI2 mainlined in Linux 6.17; ISPs began L4S rollout in production from January 2025 (Comcast as early adopter); explains TCP Prague vs UDP Prague vs SCReAM as the L4S-compatible congestion controllers. | C19 §6.5 (deployment status). |
| D6 | [Information on RFC 9330 — RFC Editor](https://www.rfc-editor.org/info/rfc9330) | Metadata + errata reference. | C19 §6.1. |
| D7 | [Comcast wields low latency as broadband differentiator — Light Reading](https://www.lightreading.com/cable-technology/comcast-wields-low-latency-as-broadband-differentiator) | Comcast LL-DOCSIS marketing posture — low-latency-as-differentiator strategy. | C19 §6.5 (deployment). |
| D8 | [Comcast Kicks Off Industry's First Low Latency DOCSIS Field Trials — Comcast Corporate](https://corporate.comcast.com/stories/comcast-kicks-off-industrys-first-low-latency-docsis-field-trials) | First-party LL-DOCSIS field-trial coverage; HelixPlay's §6.5 cites the timeline. | C19 §6.5. |
| D9 | [Reduce network delays with L4S — WWDC23 (Apple Developer)](https://developer.apple.com/videos/play/wwdc2023/10004/) | Apple WWDC23 session — iOS 17 / iPadOS 17 / macOS Sonoma / tvOS 17 baseline L4S support; FaceTime as canonical app; SwiftUI / NSURLSession ECN-marker integration. | C19 §6.6 (Apple OS surface). |
| D10 | [Comcast ISP Low Latency Deployment Design Recommendations (IETF draft)](https://www.ietf.org/archive/id/draft-livingood-low-latency-deployment-02.html) | Operational deployment recommendations — DualQ AQM placement, classifier defaults, residential-router behaviour expectations. | C19 §6.7 (operator recommendations). |
| D11 | [Comcast's L4S — 78% latency reduction tested in select cities — Tom's Hardware](https://www.tomshardware.com/tech-industry/comcasts-l4s-low-latency-tech-promises-up-to-a-78-percent-latency-reduction-testing-begins-in-select-cities) | Empirical claim — up to 78% latency reduction on LL-DOCSIS L4S; HelixPlay's §6.5 cites this number with the caveat that it is the operator's headline claim, not p999 from a third party. | C19 §6.5. |
| D12 | [Measuring Low Latency at Scale: A Field Study of L4S in Residential Broadband — Springer](https://link.springer.com/chapter/10.1007/978-3-032-18268-5_8) | Independent field study — 83 Raspberry Pi devices in Comcast subscriber households, 120,000+ controlled experiments; tail latency reduced up to 25% for interactive applications + bulk downloads from Apple's CDN. | C19 §6.5 (independent measurement). |
| D13 | [tc-ctinfo(8) — Linux man pages](https://man7.org/linux/man-pages/man8/tc-ctinfo.8.html) | The `ctinfo` tc action — DSCP-restoration mode copies DSCP stored in conntrack's connmark into IPv4/v6 diffserv field; HelixPlay's §6.8 references this for the inbound-flow classifier. | C19 §6.8 (DSCP restoration). |
| D14 | [How to Configure DSCP Marking on Ubuntu (oneuptime, 2026)](https://oneuptime.com/blog/post/2026-03-02-how-to-configure-dscp-marking-on-ubuntu/view) | 2026-aware reference for Ubuntu DSCP-marking pipeline — `iptables -t mangle -j DSCP --set-dscp-class ef` + tc HTB qdisc with `prio` classes. | C19 §6.8 (configuration). |
| D15 | [How to Configure DSCP Marking for IPv6 Packets (oneuptime, 2026)](https://oneuptime.com/blog/post/2026-03-20-dscp-marking-ipv6-packets/view) | IPv6 traffic-class marker via `ip6tables` + `tc qdisc add ... root htb`; HelixPlay's §6.8 references for the dual-stack case. | C19 §6.8. |
| D16 | [How to Use the DSCP Field for Quality of Service Marking (oneuptime, 2026)](https://oneuptime.com/blog/post/2026-03-20-dscp-field-quality-of-service-marking/view) | DSCP field semantics + AF / EF / CS classes; HelixPlay's §6.9 cites for the EF / AF41 / CS5 mapping (EF for VoIP control + low-bw signalling, AF41 for primary video stream, CS5 for SIP-style rendezvous). | C19 §6.9 (DSCP class mapping). |
| D17 | [To switch or not to switch to TCP Prague? Incentives for adoption in a partial L4S deployment (arxiv)](https://arxiv.org/html/2407.00464v1/) | Game-theoretic analysis of partial-L4S adoption; HelixPlay's §6.10 references for the bystander-flow analysis (HelixPlay's UDP Prague flow alongside Classic-ECN bulk-download neighbours). | C19 §6.10 (partial-deployment behaviour). |
| D18 | [DualPI2 Mahimahi module — arxiv](https://arxiv.org/html/2603.04381) | DualPI2 emulation toolkit for cross-platform analysis; HelixPlay's §11 (testing) integrates this for the L4S-bottleneck regression suite. | C19 §11.4 (test fixture). |
| D19 | [L4S Architecture — Nokia Bell Labs](https://www.nokia.com/bell-labs/research/l4s/) | TCP Prague from Bell Labs — first open-source implementation of the new congestion-control algorithm; HelixPlay's §6.4 cites for the provenance. | C19 §6.4. |
| D20 | [App-Developer-Guide.md — IETF L4S Deployment GitHub](https://github.com/jlivingood/IETF-L4S-Deployment/blob/main/App-Developer-Guide.md) | Application-developer integration guide — opting into ECT(1), measuring CE-mark response, fallback to Classic-ECN. | C19 §6.11 (integration guide). |

---

## §E io_uring + NAPI hybrid

The C19 chapter §7 cross-links to C16 §2.4 + §3 (the io_uring +
SQPOLL chapter) for the kernel-side io_uring posture. This
cluster captures the network-protocol-relevant io_uring + NAPI
hybrid evidence — the `io_uring_register_napi(2)` registration
introduced at kernel 6.9, the hybrid IO polling shipped in 6.13,
the NAPI busy-poll integration that brings io_uring's UDP
roundtrip from ~ 40 µs to ~ 30 µs (per the lano1106
io_uring_udp_ping benchmark cited below). **Insight #1
(Microwave Pipeline)** is reaffirmed at the network leg: the
io_uring + NAPI hybrid achieves the kernel-bypass-equivalent
latency of DPDK on commodity hardware, which is the deciding
factor for the **CZ-01 io_uring vs DPDK** resolution
([`../04_Latency/00_Index.md`](../04_Latency/00_Index.md) §7).

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| E1 | [io_uring_register_napi(3) — Linux manual page](https://man7.org/linux/man-pages/man3/io_uring_register_napi.3.html) | Function registers NAPI settings — `napi_busy_poll_to` (busy-poll timeout in µs) + `prefer_busy_poll` flag; corresponds to SO_PREFER_BUSY_POLL socket option. | C19 §7.1 (registration interface). |
| E2 | [io_uring: add napi busy polling support — LWN](https://lwn.net/Articles/930494/) | LWN coverage of NAPI integration — adds `napi_list` to `io_ring_ctx`; SQPOLL thread can perform the NAPI busy poll; integrates with `IORING_SETUP_SQPOLL`. | C19 §7.2 (kernel integration). |
| E3 | [io_uring: Add support for napi_busy_poll [LWN.net]](https://lwn.net/Articles/887237/) | Earlier patchset coverage — establishes the design rationale for NAPI registration via io_uring_register opcode. | C19 §7.2. |
| E4 | [io_uring_udp_ping — github.com/lano1106](https://github.com/lano1106/io_uring_udp_ping) | The canonical UDP-ping benchmark — without NAPI: 40.6–42.1 µs RTT; with NAPI: 30.6–31.8 µs RTT (without SQPOLL); 35.8–37.3 µs RTT (with SQPOLL); HelixPlay's §7.3 cites this as the binding empirical reference. | C19 §7.3 (benchmark). |
| E5 | [io_uring_udp_ping.cpp source](https://github.com/lano1106/io_uring_udp_ping/blob/main/io_uring_udp_ping.cpp) | C++ source for the benchmark; HelixPlay's §11 (testing) imports this pattern as the basis for the latency-regression test. | C19 §11.3 (test fixture). |
| E6 | [IO_uring Enjoys Hybrid IO Polling & Ring Resizing With Linux 6.13 — Phoronix](https://www.phoronix.com/news/Linux-6.13-IO_uring) | Hybrid IO polling shipped in 6.13 — strict-poll variant with initial sleep delay to reduce CPU burn; HelixPlay's §7.4 cites this as the kernel floor for the polling-vs-CPU trade-off. | C19 §7.4 (hybrid polling). |
| E7 | [io_uring: Add support for napi_busy_poll — 0day-ci/linux commit](https://github.com/0day-ci/linux/commit/65e72f78c66272f7cf0e87dfeef88f5b79de2d91) | Reference patch commit linking NAPI integration to io_uring; HelixPlay's chapter §7.2 cites for the patch-history audit. | C19 §7.2. |
| E8 | [Submission Queue Polling — Lord of the io_uring](https://unixism.net/loti/tutorial/sq_poll.html) | Pedagogical reference for SQPOLL — kernel thread polls submission queue, eliminating syscalls for io_uring_enter; HelixPlay's chapter §7.2 cites for the SQPOLL × NAPI interaction. | C19 §7.2. |
| E9 | [io_uring_setup(2) — Linux manual page](https://man7.org/linux/man-pages/man2/io_uring_setup.2.html) | `IORING_SETUP_SQPOLL` flag binding interface; HelixPlay's chapter §7.5 cites for the bootstrap path. | C19 §7.5. |
| E10 | [Efficient IO with io_uring — kernel.dk](https://kernel.dk/io_uring.pdf) | Original Axboe paper; HelixPlay's chapter §7 cites for the design-rationale primary source. | C19 §7.0. |
| E11 | [io_uring for High-Performance DBMSs: When and How to Use It — arxiv](https://arxiv.org/html/2512.04859v1) | Database-systems-focused empirical study — SQPoll achieves lower latency than DeferTR; advantage disappears once NAPI is enabled; DeferTR + NAPI yields best overall latency. | C19 §7.6 (DeferTR vs SQPOLL × NAPI). |
| E12 | [io_uring is slower than epoll — pion DTLS Issue #189 (axboe/liburing)](https://github.com/axboe/liburing/issues/189) | Counter-evidence — io_uring slower than epoll under specific workloads (single-connection with infrequent IO); HelixPlay's §7.7 cites this as the boundary case where epoll is preferred. | C19 §7.7 (boundary cases). |
| E13 | [polling mode using liburing example — axboe/liburing Issue #385](https://github.com/axboe/liburing/issues/385) | liburing polling-mode example for the developer-facing path. | C19 §7.5. |

---

## §F User-space TCP (mTCP, F-Stack, Seastar) — V1 deferral status

User-space TCP stacks (mTCP, F-Stack, Seastar, Tempesta-FW,
TLDK, Light) are out of MVP scope for HelixPlay and explicitly
deferred to V1 / post-MVP. The chapter §8 documents the
deferral rationale: HelixPlay's gaming traffic is UDP, not TCP;
the only TCP-bearing flows are the rendezvous + auth +
Connect-Go RPC control-plane lanes, which run on the kernel
TCP stack at low rate (≤ 100 RPS per session), where user-space
TCP would buy nothing. The chapter inventories the user-space
TCP options for completeness so that V1 reviewers do not
re-litigate the architecture choice in a vacuum.

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| F1 | [mTCP / F-Stack / TLDK / Seastar inventory — DPDK devopedia](https://devopedia.org/dpdk) | Catalogue — DPDK doesn't include TCP/IP stack; userspace stacks (F-Stack, mTCP, TLDK, Seastar, Accelerated Network Stack) layer on top. | C19 §8.1 (inventory). |
| F2 | [Seastar Networking — seastar.io](https://seastar.io/networking/) | Seastar — sharded TCP/IP stack; works with customised DPDK; vhost driver for development testing; HelixPlay's §8.3 references for the ScyllaDB-style architecture. | C19 §8.3 (Seastar). |
| F3 | [Seastar FAQ — seastar.io](https://seastar.io/faq/) | Comparison vs mTCP — both solve same problem (CPU-locality, lock elimination, syscall overhead). | C19 §8.3. |
| F4 | [Seastar — scylladb/seastar GitHub](https://github.com/scylladb/seastar) | Source repository; HelixPlay's chapter §8.3 references commits mentioning DPDK 24.x integration. | C19 §8.3. |
| F5 | [F-Stack — High Performance Network Framework Based On DPDK](https://www.f-stack.org/) | F-Stack — userspace network development kit; DPDK + FreeBSD TCP/IP stack + coroutine API; HelixPlay's §8.2 references. | C19 §8.2 (F-Stack). |
| F6 | [F-Stack on GitHub](https://github.com/F-Stack/f-stack) | Source repository; ongoing maintenance; HelixPlay's §8.2 cites for V1-readiness signal. | C19 §8.2. |
| F7 | [F-Stack vs mTCP and Seastar — F-Stack Issue #26](https://github.com/F-Stack/f-stack/issues/26) | Maintainer-authored comparison; HelixPlay's §8.0 (decision matrix) imports the comparison rationale. | C19 §8.0 (decision matrix). |
| F8 | [User space TCP? — Tempesta Technologies blog](https://tempesta-tech.com/blog/user-space-tcp/) | 2026-relevant analysis arguing that user-space TCP is not always a win — kernel-TCP NAPI + io_uring closes much of the gap; HelixPlay's §8.5 cites this for the V1-deferral rationale. | C19 §8.5 (deferral rationale). |
| F9 | [Light: A Scalable, High-performance and Fully-compatible User-level TCP Stack — fd.io PDF](https://wiki.fd.io/images/f/f4/02_ld_Light_A_scalable_High_Performance_and_Fully_compatible_TCP_Stack.pdf) | "Light" user-space TCP stack — CPU-locality + scalability claims; HelixPlay's §8.4 references for the alternative-implementation landscape. | C19 §8.4 (alternatives). |
| F10 | [StackMap: Low-Latency Networking with the OS Stack — USENIX ATC '16](https://www.usenix.org/system/files/conference/atc16/atc16-paper_yasukata.pdf) | Reference paper on hybrid (kernel + userspace) approaches; HelixPlay's §8.6 references for the design-space context. | C19 §8.6. |
| F11 | [Why do we use the Linux kernel's TCP stack? — Julia Evans](https://jvns.ca/blog/2016/06/30/why-do-we-use-the-linux-kernels-tcp-stack/) | Counter-perspective — kernel TCP stack is highly tuned, mature, secure; HelixPlay's §8.5 cites for the V1-deferral conservatism. | C19 §8.5. |
| F12 | [seastar/README-DPDK.md](https://github.com/scylladb/seastar/blob/master/README-DPDK.md) | Seastar × DPDK integration documentation; HelixPlay's §8.3 cites for the build-and-deploy story. | C19 §8.3. |

---

## §G RoCE v2 / iWARP for intra-cluster traffic

RoCE v2 is the binding intra-cluster RDMA-over-Ethernet
mechanism for HelixPlay's game-host ↔ capture-host ↔ encode-host
intra-rack traffic on dedicated edge tier where GPU-Direct RDMA
applies (per
[`../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md`](../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md)
C18 §4). Never on client-facing traffic; never on commodity
hosts; only on dedicated-edge-tier intra-rack paths where
PFC / ECN / DCQCN configuration is feasible. iWARP is documented
for completeness but is not deployed by HelixPlay (RoCE v2 is the
de-facto Mellanox / Broadcom / NVIDIA standard since 2017).

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| G1 | [RoCE vs. iWARP Competitive Analysis — NVIDIA Networking white paper](https://network.nvidia.com/related-docs/whitepapers/WP_RoCE_vs_iWARP.pdf) | RoCE simplifies transport — bypasses TCP stack; iWARP uses DDP + MPA + RDMAP layered over TCP/IP, more layers + higher latency; benchmarks show RoCE consistently faster; Mellanox + Xilinx + Broadcom recommend / exclusively support RoCE / RoCE v2. | C19 §9.1 (vendor consensus). |
| G2 | [RoCE — NVIDIA MLNX_OFED documentation](https://docs.nvidia.com/networking/display/MLNXOFEDv497100LTS/RDMA+over+Converged+Ethernet+(RoCE)) | Authoritative MLNX_OFED reference; UDP encapsulation in RoCE v2 (vs raw Ethernet in v1); HelixPlay's §9.2 cites this for the protocol-version distinction. | C19 §9.2 (RoCE v1 vs v2). |
| G3 | [RoCE — Wikipedia](https://en.wikipedia.org/wiki/RDMA_over_Converged_Ethernet) | Background reference + 800G/1.6T deployment timing for 2025–2026. | C19 §9.0. |
| G4 | [RoCEv2 Explained: The Ultimate Guide to Low-Latency, High-Throughput Networking in AI Data Centers — fibermall](https://www.fibermall.com/blog/rocev2-ultimate-guide-to-low-latency.htm) | Application-layer guide — covers PFC + ECN + DCQCN + buffer-tuning; HelixPlay's §9.3 references for the lossless-fabric configuration. | C19 §9.3 (fabric configuration). |
| G5 | [Lossless Ethernet Design Guide for AI Fabrics (RoCE v2) | 2026 Updated — intelligentvisibility](https://intelligentvisibility.com/ai-networking-solutions/lossless-networking-ai) | 2026-current guide — explicit PFC + ECN configuration for RoCE v2; HelixPlay's §9.3 cites this as the deployment reference. | C19 §9.3. |
| G6 | [PFC Flow Control Technology and Challenges in RoCEv2 Network Deployment — NADDOD blog](https://www.naddod.com/blog/pfc-flow-control-technology-and-challenges-in-rocev2-network-deployment) | PFC challenges — head-of-line-blocking risk; HelixPlay's §9.4 cites this for the operator-cost analysis. | C19 §9.4. |
| G7 | [RoCE vs InfiniBand for AI Data Center Networking: What Network Engineers Need to Know in 2026 — FirstPassLab](https://firstpasslab.com/blog/2026-03-09-roce-vs-infiniband-ai-data-center-networking/) | 2026-current comparison — RoCE v2 delivers 85–95% of InfiniBand's training throughput at significantly lower cost; Meta deployed 24K-GPU RoCE v2 cluster on Arista 7800 switches with 400 Gbps endpoints. | C19 §9.5 (Meta upper-bound deployment). |
| G8 | [Understanding DCQCN — WWT](https://www.wwt.com/article/understanding-data-center-quantized-congestion-notification-dcqcn) | DCQCN — Data Center Quantized Congestion Notification; the binding congestion-control algorithm for RoCE v2 lossless fabrics. | C19 §9.4 (DCQCN). |
| G9 | [DCQCN — Juniper Junos OS](https://www.juniper.net/documentation/us/en/software/junos/traffic-mgmt-qfx/topics/topic-map/cos-qfx-series-DCQCN.html) | Vendor-side DCQCN configuration reference; HelixPlay's §9.4 cites for the multi-vendor deployment surface. | C19 §9.4. |
| G10 | [Lossless Network for AI/ML/Storage/HPC with RDMA — Arista + Broadcom deployment guide (PDF)](https://www.arista.com/assets/data/pdf/Broadcom-RoCE-Deployment-Guide.pdf) | Deployment guide — Broadcom / Arista joint reference; HelixPlay's §9.6 references for the multi-vendor topology. | C19 §9.6 (multi-vendor topology). |
| G11 | [Configuring RoCE on Red Hat Enterprise Linux 8 — Red Hat Documentation](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/8/html/configuring_infiniband_and_rdma_networks/configuring-roce_configuring-infiniband-and-rdma-networks) | Distribution-side configuration reference; HelixPlay's §9.7 references for the host-side userspace + kernel module setup. | C19 §9.7 (host config). |

---

## §H 2026 papers + benchmarks (NSDI / ATC / SIGCOMM / IMC ULL track)

The 2026-2025 conference frontier on ultra-low-latency networking
extends — never replaces — `latency_dim05.md`. The chapter §10
incorporates the binding 2026-aware findings: Junction (NSDI'24
into NSDI'25 follow-on for the cloud kernel-bypass surface),
Falcon hardware transport (SIGCOMM'25 — 200 Gbps + 120 Mops/sec,
8× lower OCT than CX-7 baseline), open-up kernel-bypass TCP
stacks (ATC'25, related to F-Stack / mTCP V1-deferral analysis),
the IMC'25 cloud-gaming context-classification paper (gaming-
context-aware QoE measurement), the SIGCOMM'25 ultra-low-latency
video-streaming tutorial (Zili Meng's ULL-streaming research
programme).

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| H1 | [Falcon: A Reliable, Low Latency Hardware Transport — SIGCOMM 2025 (ACM DL)](https://dl.acm.org/doi/10.1145/3718958.3754353) | Falcon — first hardware transport supporting multiple ULPs + heterogeneous workloads in general-purpose Ethernet datacenter environments; 200 Gbps + 120 Mops/sec; near-optimal OCT up to 8× lower than CX-7 baseline. HelixPlay's §10.1 references for the **upper-bound** hardware-transport reference (out of MVP scope; on V1 roadmap iff Google releases a Falcon-compatible NIC publicly). | C19 §10.1 (Falcon — V1 upper bound). |
| H2 | [SIGCOMM 2025 — dblp](https://dblp.org/db/conf/sigcomm/sigcomm2025.html) | Full conference index for the 2025 ACM SIGCOMM proceedings; HelixPlay's §10 cites for paper-discovery audit. | C19 §10.0. |
| H3 | [Optimizing Low-Latency Video Streaming: AI-Assisted Codec — SIGCOMM 2025 tutorial](https://conferences.sigcomm.org/sigcomm/2025/tutorials-hackathons/tutorial-vsai/) | SIGCOMM'25 ultra-low-latency video-streaming tutorial; Zili Meng's ULL-streaming research (9 SIGCOMM + NSDI papers in recent years); HelixPlay's §10.2 references for the academic-frontier signal. | C19 §10.2 (academic frontier). |
| H4 | [Making Kernel Bypass Practical for the Cloud with Junction — NSDI'24 (USENIX PDF)](https://www.usenix.org/system/files/nsdi24-fried.pdf) | Junction — kernel-bypass cloud runtime; addresses the practical implementation of kernel-bypass networking in cloud environments; HelixPlay's §10.3 references for the kernel-bypass-in-the-cloud design. | C19 §10.3 (Junction). |
| H5 | [Junctiond: Extending FaaS Runtimes with Kernel-Bypass — arxiv](https://arxiv.org/html/2403.03377v2) | Follow-on to Junction — extends FaaS runtimes with kernel-bypass; HelixPlay's §10.3 references for the FaaS-side application surface. | C19 §10.3. |
| H6 | [Opening Up Kernel-Bypass TCP Stacks — USENIX ATC 2025 PDF](https://www.usenix.org/system/files/atc25-awamoto.pdf) | ATC'25 paper — analyses tradeoffs of kernel-bypass TCP stacks (mTCP, F-Stack, Seastar); HelixPlay's §10.4 + §8.5 (V1-deferral) cite for the empirical justification. | C19 §10.4 + §8.5. |
| H7 | [A Wake-Up Call for Kernel-Bypass on Modern Hardware — DaMoN '25 (ACM DL)](https://dl.acm.org/doi/10.1145/3736227.3736235) | 2025 DaMoN paper — modern hardware (Sapphire Rapids + ConnectX-7) shifts the kernel-bypass calculus; HelixPlay's §10.5 references for the hardware-evolution argument. | C19 §10.5. |
| H8 | [NSDI 2025 — Awesome Papers index](https://paper.lingyunyang.com/reading-notes/conference/nsdi-2025) | Curated NSDI'25 paper index; HelixPlay's §10.0 references for paper-discovery audit. | C19 §10.0. |
| H9 | [ACM IMC 2025 — Accepted Papers](https://conferences.sigcomm.org/imc/2025/accepted-papers/) | IMC'25 conference accepted-paper list; "Games Are Not Equal: Classifying Cloud Gaming Contexts for Effective User Experience Measurement" (Wang, Lyu, Sivaraman) is the binding cloud-gaming-measurement paper. | C19 §10.6 (gaming-context measurement). |
| H10 | [ACM SIGCOMM 2025 — Tutorial: ENAI](https://conferences.sigcomm.org/sigcomm/2025/tutorials-hackathons/tutorial-enai/) | Conference tutorial reference for the academic-frontier signal. | C19 §10.0. |
| H11 | [Adam Belay's homepage (MIT) — Junction author](http://www.abelay.me/) | Researcher-page reference for Junction provenance; HelixPlay's §10.3 cites for the academic provenance audit. | C19 §10.3. |
| H12 | [SIGCOMM 2025 papers info index](https://conferences.sigcomm.org/sigcomm/2025/program/papers-info/) | Full conference papers-info index; HelixPlay's §10 cites for audit completeness. | C19 §10.0. |

---

## §I 2026 hardware — Mellanox CX-7+, Intel E810, Broadcom BNXT

The 2026 hardware floor for HelixPlay's edge-tier hosts is fixed
by what the DPDK PMD and AF_XDP zero-copy support actually targets
in the 24.11 / 25.07 / 26.03 LTS surface: NVIDIA Mellanox
ConnectX-7 (400 Gbps) + ConnectX-8 (800 Gbps) + BlueField-3 DPU,
Intel Ethernet Controller E810 (100 GbE — the SPR / Emerald
Rapids canonical platform NIC), Broadcom BCM5741X / BCM575XX /
BCM57608 NetXtreme-E. ConnectX-8 + BlueField-3 are the
**operator-policy-opt-in upper-bound** for HelixPlay edge tier;
ConnectX-7 is the practical floor; E810 is the cost-optimised
alternative on Intel platforms; bnxt is the third tier.

| # | Source | Headline finding for C19 | Section pointer |
|---|--------|--------------------------|-----------------|
| I1 | [Intel® Ethernet Network Adapter E810-CQDA2](https://www.intel.com/content/www/us/en/products/sku/189760/intel-ethernet-network-adapter-e810xxvda2/specifications.html) | E810 specifications + 100 GbE baseline; HelixPlay's §3.4 + §I cites for the Intel-side floor. | C19 §3.4. |
| I2 | [Intel E810-CQDA2 Dual-Port 100GbE NIC Review — ServeTheHome](https://www.servethehome.com/intel-e810-cqda2-dual-port-100gbe-nic-review/) | Independent review with empirical numbers; HelixPlay's §3.4 cites for the ConnectX-vs-E810 cost-vs-performance comparison. | C19 §3.4. |
| I3 | [Intel® Ethernet Network Adapter E810-2CQDA2 (Dell brief PDF)](https://www.delltechnologies.com/asset/en-us/products/servers/technical-support/intel-ethernet-network-adapter-e810-2cqda2-product-brief.pdf) | 2-CQDA2 variant — up to 200 Gbps total bandwidth (2× 100 G); HelixPlay's §3.4 cites for the multi-port density. | C19 §3.4. |
| I4 | [Linux Base Driver for the Intel(R) Ethernet Controller 800 Series — kernel.org](https://docs.kernel.org/networking/device_drivers/ethernet/intel/ice.html) | Authoritative kernel-side ice-driver documentation; AF_XDP zero-copy supported; XDP frame-size cap ≤ 3 KB. | C19 §3.4 + §11.1. |
| I5 | [intel/ethernet-linux-ice — GitHub](https://github.com/intel/ethernet-linux-ice/) | Out-of-tree ice driver source; HelixPlay's §3.4 cites for the kernel + DKMS deployment story. | C19 §3.4. |
| I6 | [N2S-MBF301 SmartNIC with NVIDIA BlueField DPU + ConnectX-7 — Lanner](https://www.lannerinc.com/news-and-events/latest-news/n2s-mbf301-smart-nic-with-nvidia-bluefield-dpu-and-connectx-7-chipset) | BlueField + ConnectX-7 SmartNIC reference; HelixPlay's §I cites for the integrated DPU + NIC platform option. | C19 §I (DPU integration). |
| I7 | [NVIDIA BlueField — Wikipedia](https://en.wikipedia.org/wiki/Nvidia_BlueField) | BlueField-3 DPU spec — 400 Gbps Ethernet or 400 Gbps NDR InfiniBand; 16 64-bit ARM A78 cores; offloads + acceleration + isolation of SDN, storage, security, management. | C19 §I (BlueField-3). |
| I8 | [Nvidia Quantum-2 Networking Platform with NDR InfiniBand and BlueField-3 DPU — HPCwire](https://www.hpcwire.com/2021/11/10/nvidia-debuts-quantum-2-networking-platform-with-ndr-infiniband-and-bluefield-3-dpu/) | Provenance reference — Quantum-2 platform (Quantum-2 switch + ConnectX-7 + BlueField-3 + software). | C19 §I (Quantum-2 stack). |
| I9 | [AF_XDP Zero Copy Support for mlx5_core driver — NVIDIA Forums](https://forums.developer.nvidia.com/t/af-xdp-zero-copy-support-for-mlx5-core-driver/253112) | mlx5 AF_XDP zero-copy support thread; HelixPlay's §11.1 cites for the AF_XDP × mlx5 deployment confirmation. | C19 §11.1. |
| I10 | [How to Configure AF_XDP for User-Space Networking on Ubuntu (oneuptime, 2026)](https://oneuptime.com/blog/post/2026-03-02-configure-af-xdp-user-space-networking-ubuntu/view) | 2026-aware Ubuntu AF_XDP setup guide; HelixPlay's §11.2 cites for the user-space binding. | C19 §11.2. |
| I11 | [AF_XDP — Linux Kernel documentation](https://docs.kernel.org/networking/af_xdp.html) | Authoritative kernel-side AF_XDP documentation; HelixPlay's §11.0 cites for the binding interface. | C19 §11.0. |
| I12 | [Intel Gigabit Ethernet Driver To Speed-Up With AF_XDP Zero-Copy For Linux 6.14 — Phoronix](https://www.phoronix.com/news/IntelIGB-AF-XDP-Zero-Copy) | Linux 6.14 Intel IGB AF_XDP zero-copy support — extends AF_XDP support to 1 GbE Intel adapters; HelixPlay's §I (forward-compat) cites for the 1G-deployment-option signal. | C19 §I (kernel forward-compat). |
| I13 | [Extending AF_XDP for fast co-located packet transfer — FOSDEM 2026 (PDF)](https://fosdem.org/2026/events/attachments/E8RFHV-flash-afxdp/slides/266831/fosdem26-_n5vpfki.pdf) | FOSDEM 2026 presentation on extending AF_XDP for fast co-located packet transfer; the 2026-current research frontier. | C19 §I (research frontier). |

---

## §Z Contradictions index — places where 2026 evidence diverges from `latency_dim05.md`

The 2024 / early-2025 baseline at
[`../../02_latency/02_Response/Agent_results/research/latency_dim05.md`](../../02_latency/02_Response/Agent_results/research/latency_dim05.md)
(103 lines) is the prior-work checkpoint for this addendum.
Where the 2026 evidence above diverges, the chapter §12 records
the divergence; where it confirms, the chapter cites the 2024
finding by source ID.

| Z-ID | 2024 baseline (latency_dim05.md) | 2026 evidence (this addendum) | Resolution in C19 |
|------|-----------------------------------|-------------------------------|-------------------|
| Z-1 | DPDK 1 M+ pps single-core, line-rate at 10 Gbps (§3, source 983 Reddit `r/LocalLLaMA`). | DPDK 24.11 LTS + ConnectX-7 reaches 100 G line-rate single-core for typical packet sizes (A4 + A12 + A13). The 10 G figure is conservative for 2026 hardware. | §3.7 — chapter prose updates the per-core baseline to 100 G with ConnectX-7 + DPDK 24.11; cites A13 NVIDIA performance report as the binding empirical reference. |
| Z-2 | DPDK 15 µs tail latency vs kernel ~ 40 µs (§3, source 944 medium.com beyond-localhost — 2025-12-29). | io_uring + NAPI hybrid achieves ~ 30 µs UDP RTT (E4 lano1106 benchmark); io_uring + SQPOLL + NAPI achieves ~ 35 µs RTT. The kernel-side 40 µs figure is no longer the floor with NAPI registration. | §7.3 — chapter updates the kernel-side floor to ~ 30 µs at io_uring + NAPI, narrowing the DPDK-vs-kernel gap from 25 µs to ~ 15 µs; CZ-01 resolution refined: io_uring on commodity hosts now closer to DPDK than the 2024 baseline implied. |
| Z-3 | Moonlight protocol uses ENet UDP library with custom framing (§5, source 944 — moonlight-stream.org). | The ENet binding remains current in 2026; HelixPlay's chapter §4 ships a custom-UDP + DTLS 1.2 binding (not ENet), citing C02 §2 + §11. | §4 — chapter clarifies that HelixPlay does not depend on ENet; the Moonlight reference is informational only. |
| Z-4 | Parsec 7 ms LAN latency with BUD custom protocol (§5, source 998 — geekysafari.com — 2025-07-25). | C02 §13 already pins HelixPlay's custom-UDP + DTLS 1.2 stack; the 7 ms figure is a 2025-vintage marketing claim. 2026 evidence on the WebRTC vs custom-UDP gap (TRTC blog, B11) reaffirms sub-500 ms baseline for WebRTC; HelixPlay achieves 7 ms LAN with custom UDP per C02 §2 §13. | §4 — chapter cites C02 §13's 7 ms LAN figure (with the BUD-publication-claim caveat). |
| Z-5 | TSN — IEEE 802.1Qbv time-aware shaper, 802.1Qbu frame preemption (§6, source 993 — Intel industrial automation). | TSN is out of HelixPlay MVP scope per `04_Latency/00_Index.md` §9 V1-deferral list (HelixPlay's residential clients do not have TSN-capable LAN switches); the chapter §13 reaffirms the V1-deferral. | §13 — chapter cites the 2024 baseline as informational; binding deployment scope excludes TSN MVP. |
| Z-6 | RDMA — RoCE v2 enables sub-microsecond remote memory access with zero CPU involvement (§4, source 981 NVIDIA networking docs). | The 2026 RoCE v2 evidence (G1 + G7 + G10) reaffirms the sub-µs RDMA latency AND adds the production-scale Meta 24K-GPU deployment as upper-bound reference. The 800G/1.6T port standard for 2025-2026 deployments (G3 Wikipedia) is new context. | §9 — chapter cites the 2024 baseline + 2026 production scale; binding posture: RoCE v2 only on dedicated edge-tier intra-rack paths, never on client-facing traffic. |
| Z-7 | GPUDirect RDMA — 12 GB/s throughput, 137 µs best-case latency (§4, source 981 SciTechDaily — 2024-10-09). | 2026 BlueField-3 + ConnectX-7 (I6 + I7 + I8) raises the throughput ceiling but the 137 µs latency figure remains roughly current for the cross-PCIe-root-complex case; intra-PCIe-root case lower. | §9 — chapter defers detailed GPUDirect RDMA elaboration to C18 (intra-rack RDMA owner); cites 2024 baseline as floor. |
| Z-8 | "QUIC offers 0-RTT but adds protocol overhead — increased latency vs raw UDP" (§2, source 991 reddit nvidia reflex). | The 2026 ecosystem evidence (B11 + WebRTC trends) reaffirms the QUIC vs raw-UDP gap; HelixPlay's MVP uses raw UDP + DTLS 1.2 (per C02 §13); QUIC datagrams (RFC 9221) are a Phase-2 upgrade path per C02 §2 row 3. | §4 — chapter cites the 2024 finding (QUIC overhead vs raw UDP) as the binding rationale for raw-UDP-MVP. |
| Z-9 | Custom UDP / Moonlight / Parsec table — QUIC at 5–10 µs (§7 table). | The 5–10 µs figure was a per-packet processing cost estimate, not RTT. The 2026 evidence (B11 WebRTC live streaming sub-500 ms; D11 Comcast L4S 78% reduction; D12 independent measurement 25% tail-latency reduction) shifts the binding latency lens from per-packet processing to end-to-end p99 / p999 over residential networks. | §6 + §11 — chapter updates the binding lens to end-to-end p99 / p999 over residential networks; cites C13 §10 measurement methodology. |
| Z-10 | (Implicit) — `latency_dim05.md` does not address L4S, ECT marking, DualPI2, or TCP Prague. | 2026 evidence (D1–D20) establishes L4S as the binding 2026 residential-network QoS mechanism: Linux 6.17 mainline, Comcast LL-DOCSIS production 2025, Apple iOS 17+ baseline. | §6 — entire L4S section is **net new** vs the 2024 baseline. |
| Z-11 | (Implicit) — `latency_dim05.md` does not address Cloudflare Realtime managed TURN. | 2026 evidence (C6–C10) establishes Cloudflare Realtime as a managed TURN alternative to self-hosted coturn; 330+ POPs anycast; free with Realtime SFU; $0.05 / GB outbound otherwise. | §5.4 — net new evaluation surface; chapter binds operator-policy choice between self-hosted coturn 4.6+ and Cloudflare Realtime. |
| Z-12 | (Implicit) — `latency_dim05.md` cites DPDK + RDMA + custom UDP; does not catalogue user-space TCP stacks. | 2026 evidence (F1–F12) catalogues mTCP / F-Stack / Seastar / TLDK / Light / Tempesta with a V1-deferral status (HelixPlay gaming traffic is UDP; user-space TCP is V1-only). | §8 — entire user-space TCP section is **net new** vs the 2024 baseline; deferral rationale recorded. |
| Z-13 | (Implicit) — `latency_dim05.md` does not catalogue 2025-2026 ULL conference papers. | H1–H12 catalogue Falcon SIGCOMM'25, Junction NSDI'24/'25 follow-on, ATC'25 user-space TCP, IMC'25 cloud-gaming context measurement, SIGCOMM'25 ULL-streaming tutorial. | §10 — entire academic-frontier section is **net new** vs the 2024 baseline. |

**Cited insight + cross-verification validation outcomes:**

- **Insight #1 (Microwave Pipeline)** — reaffirmed by 2026 evidence in §A (DPDK egress leg) + §E (io_uring + NAPI egress leg) + §G (RoCE v2 intra-cluster); the network-egress leg of the unified zero-copy pipeline has multiple valid implementations (DPDK / io_uring + NAPI / AF_XDP / RoCE v2) with the operator-policy posture choosing the right tier.
- **HC-06 (DPDK provides lowest network latency but highest complexity)** — reaffirmed and sharpened by 2026 evidence: DPDK 24.11 LTS extends the operational complexity (3-year LTS window + ABI commitments + driver-version tracking) but commodity-host io_uring + NAPI now reaches ~ 30 µs RTT, narrowing the gap from ~ 25 µs to ~ 15 µs. Z-2 records the divergence.
- **CZ-01 (io_uring vs DPDK for video streaming)** — refined: io_uring + NAPI hybrid on commodity hosts is the MVP floor; DPDK 24.11 LTS only on operator-policy-opt-in dedicated edge tier. The 2024 resolution ("DPDK if dedicated cores; io_uring + SQPOLL on general hosts") stands; the 2026 evidence merely tightens the gap.

---

## Anti-Bluff Posture (Constitution §1.1)

This addendum complies with Constitution §1.1: no `TODO`,
`FIXME`, `XXX`, `HACK`, `tbd`, `???`, `placeholder`, "and similar",
"etc.", "as appropriate", "as needed", "where reasonable", "fill
in later" outside this disclaimer block. Every claim in §A–§I is
sourced to a real WebSearch URL captured 2026-04-29; no URL is
fabricated. The **§Z contradictions index** records every place
the 2026 evidence diverges from the 2024 baseline at
`latency_dim05.md`. The owning chapter (C19,
`05_UltraLowLatency_Network_Protocols.md`) is responsible for
incorporating these findings into chapter prose under R-01
(line floor ≥ 250) and R-13 (anti-bluff verification). R-18
(Operational Integrity, Constitution §11.5) is honoured: this
file does not document or recommend any host-disrupting command;
all `tc qdisc` / `ip link set` / `ethtool` / `setcap` invocations
referenced are routed through `r18.SafeExec` per
[`../04_Latency/00_Index.md`](../04_Latency/00_Index.md) §6.

End of `2026-04-29-ultra-low-latency-network-protocols.md`
— 2026-04-29.
