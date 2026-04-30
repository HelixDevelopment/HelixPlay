# `helix-network` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-network`                                                                                                        |
| **Origin chapter:section**  | [C19 §6](../../04_Latency/05_UltraLowLatency_Network_Protocols.md) — *DSCP / L4S / jitter buffer / pacing*            |
| **Public path (4 mirrors)** | `vasic-digital/helix-network` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `golang.org/x/sys/unix`, `github.com/pion/rtp`, `github.com/pion/rtcp`                                                  |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `dscp-l4s-jitter-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                       |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/08_controller_input_low_latency/scenarios/01_dscp_l4s_under_240hz_input.scenario.yaml`                  |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-network` provides **DSCP marking, L4S explicit-congestion-notification handling, jitter-buffer management, and packet pacing** primitives. Origin: [C19 §6](../../04_Latency/05_UltraLowLatency_Network_Protocols.md). The submodule is what makes HelixPlay's traffic distinguishable to ECN-aware queues on operator routers + switches; without it, HelixPlay's gameplay packets would mix with bulk traffic at every hop.

The submodule was introduced to consolidate DSCP/ECN/L4S/jitter-buffer logic that earlier projects re-implemented inline. R-04 mandates one canonical landing.

---

## 2. Public API Surface

### 2.1 The `Marker` type — DSCP + ECN

```go
package network

// Marker tags outgoing UDP packets with the DSCP code-point and
// (when L4S is enabled) the ECT(1) ECN code-point that downstream
// queues use to apply L4S marking.
type Marker struct { /* ... */ }

func NewMarker(dscp uint8, opts ...Option) *Marker
func (m *Marker) Apply(socket Socket) error
func (m *Marker) DSCP() uint8
func (m *Marker) ECN() ECNCodepoint  // ECT(0), ECT(1), Not-ECT, CE
```

DSCP code points used by HelixPlay (per C19 §6):
- **EF (46)** — Expedited Forwarding for the gameplay UDP stream.
- **AF41 (34)** — Assured Forwarding for video frames.
- **AF31 (26)** — Audio.
- **CS3 (24)** — Control plane.
- **BE (0)** — Bulk + telemetry.

### 2.2 The `JitterBuffer` type

```go
package network

// JitterBuffer absorbs receive-side jitter for time-sensitive
// streams. The buffer is an MPSC ring (helix-lockfree) with a
// playout-deadline schedule.
type JitterBuffer struct { /* ... */ }

func NewJitterBuffer(targetDelayMs int, opts ...Option) *JitterBuffer
func (jb *JitterBuffer) Push(pkt *rtp.Packet)
func (jb *JitterBuffer) Pop(deadline time.Time) (*rtp.Packet, bool)
func (jb *JitterBuffer) Stats() JitterStats
```

### 2.3 The `Pacer` type

```go
package network

// Pacer paces outbound packets to avoid microbursts that fill
// downstream queues; uses a leaky-bucket schedule.
type Pacer struct { /* ... */ }

func NewPacer(targetBitrate int, opts ...PacerOption) *Pacer
func (p *Pacer) Send(pkt *rtp.Packet) bool
func (p *Pacer) UpdateBitrate(bps int)
```

### 2.4 L4S ECN feedback parsing

```go
package network

// L4SFeedback decodes RTCP ECN-FB feedback and produces a
// congestion signal suitable for Pacer.UpdateBitrate or for a
// CWND-style controller.
type L4SFeedback struct { /* ... */ }

func ParseECNFB(rtcp []byte) (*L4SFeedback, error)
func (f *L4SFeedback) MarkRate() float64   // 0.0–1.0
```

### 2.5 Capability detection

```go
package network

func KernelCaps() (Capabilities, error)
type Capabilities struct {
    SO_TXTIME      bool   // TXTIME for hardware-pacing
    SO_BUSY_POLL   bool   // poll-mode receive
    L4SQDiscPresent bool  // tc qdisc dualpi2 / ll-tcp4 available
}
```

### 2.6 Configuration options

```go
package network

type Option func(*config)

func WithMarker(dscp uint8) Option
func WithECT1() Option           // request L4S ECT(1) marking
func WithSOTxTime() Option       // hardware pacing where supported
func WithBusyPoll(usec int) Option
func WithJitterBuffer(targetMs int) Option
```

### 2.7 The `JitterStats` and `PacerStats` types

```go
package network

type JitterStats struct {
    Pushed     uint64
    Popped     uint64
    Dropped    uint64       // packets too late to play
    Reordered  uint64
    DelayMS    float64      // EWMA of playout delay
}

type PacerStats struct {
    BytesSent      uint64
    PacketsSent    uint64
    QueueDepth     int
    EWMABitrateBps int
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`tc qdisc show` for ECN-feedback reachability check).

### 3.2 External (Go)

- `golang.org/x/sys/unix` — Linux syscall surface; `setsockopt(IP_TOS, ...)` for DSCP marking.
- `github.com/pion/rtp` — RTP packet types; pinned to ≥ v1.8.0.
- `github.com/pion/rtcp` — RTCP feedback parsing; pinned to ≥ v1.2.13.

### 3.3 External (system)

- Linux kernel ≥ 5.10 for `SO_TXTIME` (hardware pacing); otherwise software pacing.
- L4S queueing discipline (`dualpi2` or `ll-tcp4`) at the local NIC and at downstream routers — operator deployment concern documented in [C19 §6](../../04_Latency/05_UltraLowLatency_Network_Protocols.md).

---

## 4. Container Build (S02 §3 lane: `dscp-l4s-jitter-1.x`)

**Builder:** `golang-builder-cgo` (cgo for some unix.Setsockopt edge cases). **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + `--cap-add=NET_ADMIN` only when L4S qdisc setup is in scope (most deployments inherit the qdisc from the host, not the container).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
DSCP marker round-trip via `setsockopt`; jitter-buffer push/pop with synthetic timestamps; pacer rate-limit bookkeeping.

### 5.2 Integration
Real UDP loopback; verify TOS byte arrives at the receiver; ECN bits propagate.

### 5.3 E2E
helix-network → real network with L4S qdisc → jitter buffer → roundtrip.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: malformed RTP/RTCP fuzzer.

### 5.5 Benchmarking
DSCP marker setsockopt p999 ≤ 5 µs; jitter-buffer push p999 ≤ 200 ns; pacer schedule decision ≤ 1 µs.

### 5.6 Chaos
Toxiproxy-injected jitter (5 ms / 20 ms / 50 ms p99); verify jitter-buffer absorbs without late-packet drops.

### 5.7 Stress
24-hour 240 Hz packet stream; zero buffer leak; pacer sustains target bitrate.

### 5.8 Smoke
30-second post-deploy: send a marked packet to loopback; verify TOS bits.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`08_controller_input_low_latency/01_dscp_l4s_under_240hz_input.scenario` — 240 Hz controller input + DSCP marking + L4S markings preserved end-to-end.

---

## 6. Challenges Entry-Point (S03 §4 row #11)

**Topology:** `08_controller_input_low_latency`. **Scenario:** `01_dscp_l4s_under_240hz_input.scenario.yaml`. **Why this scenario.** DSCP marking + L4S round-trip is the baseline for HelixPlay's low-latency input path; the scenario verifies the marking survives every hop in the canonical topology. **Baseline:** TOS byte preserved at receiver; L4S ECT(1) preserved; CE marking rate ≤ 0.5 % under steady-state load; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-network` imports `helix-r18-safeexec` for boundary subprocess invocations (tc qdisc check). Hot path (setsockopt + jitter-buffer push/pop + pacer) is pure syscall + atomic.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on operator-side L4S qdisc deployment; the submodule's API can graduate independently of operator network infra.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type              | Purpose                                                                |
|------------------------------------|---------------|---------------------------|------------------------------------------------------------------------|
| `HELIX_NET_DSCP_GAMEPLAY`          | `46` (EF)     | int [0, 63]               | DSCP code-point for gameplay UDP.                                      |
| `HELIX_NET_DSCP_VIDEO`             | `34` (AF41)   | int [0, 63]               | Video DSCP.                                                             |
| `HELIX_NET_DSCP_AUDIO`             | `26` (AF31)   | int [0, 63]               | Audio DSCP.                                                             |
| `HELIX_NET_DSCP_CONTROL`           | `24` (CS3)    | int [0, 63]               | Control-plane DSCP.                                                    |
| `HELIX_NET_L4S_ENABLED`            | `auto`        | `auto`/`true`/`false`     | Use ECT(1) marking when kernel supports it.                            |
| `HELIX_NET_JITTER_TARGET_MS`       | `40`          | int [10, 200]             | Target jitter-buffer playout delay.                                    |
| `HELIX_NET_PACER_BITRATE_BPS`      | `25000000`    | int                       | Initial pacer bitrate (overridden by ABR ladder at runtime).           |
| `HELIX_NET_SO_TXTIME`              | `auto`        | `auto`/`true`/`false`     | Hardware pacing via SO_TXTIME (kernel ≥ 5.10).                         |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| `Marker.Apply()`                      | 2 µs     | 4 µs    | 5 µs    | Single setsockopt at socket creation.                            |
| `JitterBuffer.Push()`                 | 100 ns   | 150 ns  | 200 ns  | MPSC ring path.                                                  |
| `JitterBuffer.Pop()`                  | 200 ns   | 400 ns  | 600 ns  | Includes deadline check.                                         |
| `Pacer.Send()` decision               | 400 ns   | 700 ns  | 1 µs    | Leaky-bucket bookkeeping.                                        |
| `ParseECNFB()`                        | 800 ns   | 2 µs    | 4 µs    | RTCP feedback parse.                                             |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `network: ErrECT1NotSupported`                | Kernel < 5.x or qdisc lacks dualpi2                            | Disable `HELIX_NET_L4S_ENABLED`; expect higher buffer-bloat sensitivity.   |
| `network: ErrJitterBufferOverflow`            | Push rate exceeds Pop rate                                     | Increase target delay; investigate consumer slow-drain.                    |
| `network: ErrPacerStarved`                    | Pacer bitrate set too low for the workload                     | Raise via `UpdateBitrate`; ABR ladder should drive this automatically.    |
| `network: ErrTOSNotApplied`                   | Container lacks NET_ADMIN where required                       | Verify container privilege; some kernels accept TOS-marking without it.   |
| `network: ErrInvalidDSCP`                     | DSCP code-point > 63 or violates standard mapping              | Use one of the documented values (EF, AF*, CS*).                          |

### 9.4 Migration from kernel-default UDP

A consumer migrating from `net.ListenUDP` with default settings:

1. Wrap the socket with `network.NewMarker(dscp, ...).Apply(socket)`.
2. Insert a `JitterBuffer` between receive and consumer.
3. Insert a `Pacer` between producer and send.
4. Add OTLP spans around buffer + pacer; metric for `JitterStats.Dropped` (alert: > 0.1 %).
5. Coordinate with operator: confirm L4S qdisc presence on the host's egress interface.

The migration is documented in `docs/migration-from-default-udp.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_network_marker_apply_total`              | counter    | DSCP marker applications, labelled `dscp_class`.                             |
| `helix_network_jitter_pushed_total`             | counter    | Jitter-buffer pushes.                                                        |
| `helix_network_jitter_popped_total`             | counter    | Jitter-buffer pops.                                                          |
| `helix_network_jitter_dropped_total`            | counter    | Late-packet drops at the buffer; alert threshold > 0.1 %.                    |
| `helix_network_pacer_queue_depth`               | gauge      | Pacer queue occupancy.                                                       |
| `helix_network_pacer_bitrate_bps`               | gauge      | Pacer's current target bitrate.                                              |
| `helix_network_l4s_ce_rate`                     | gauge      | EWMA L4S CE marking rate from RTCP-FB.                                       |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C19 §6](../../04_Latency/05_UltraLowLatency_Network_Protocols.md) — origin                | Origin chapter; full DSCP / L4S / jitter / pacer API.                                |
| [C33 §6](../../05_Video_Audio/08_ABR_FEC_Congestion.md) — ABR                             | ABR ladder reads `pacer.Stats()` + `L4SFeedback.MarkRate()` for rung descent.        |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport applies the Marker before send; consumes JitterBuffer on receive.          |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-network-A        | DSCP code-point matrix — should HelixPlay define a tenant-overridable mapping?                                | C19 §6 next revision                                |
| OQ-network-B        | L4S qdisc deployment — how to detect operator-side dualpi2 vs ll-tcp4 vs none?                                | `08_Operations/04_Observability_and_Events.md`     |
| OQ-network-C        | Adaptive jitter-buffer — fixed target vs RTT-aware target?                                                     | C19 §6 next revision                                |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/05_UltraLowLatency_Network_Protocols.md`](../../04_Latency/05_UltraLowLatency_Network_Protocols.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #11                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Marker + JitterBuffer + Pacer + L4SFeedback parse).               |
| Integration    | Real UDP loopback; TOS byte + ECN bits propagate.                                             |
| E2E            | Real network with L4S qdisc; roundtrip preserved.                                             |
| Security       | govulncheck + Snyk + Trivy + RTP/RTCP fuzzer.                                                 |
| Benchmarking   | All §9.2 budgets met; pacer sustains target bitrate ±2 %.                                     |
| Chaos          | Toxiproxy jitter injection; buffer absorbs without drops.                                     |
| Stress         | 24-hour 240 Hz stream; zero leak.                                                             |
| Smoke          | 30-second TOS-marker verification at receiver.                                                |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `08_controller_input_low_latency/01_dscp_l4s_under_240hz_input` baseline-parity.              |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-network.md` — 2026-04-30.
