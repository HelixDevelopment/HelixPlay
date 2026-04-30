# `helix-transport` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-transport`                                                                                                      |
| **Origin chapter:section**  | [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — *RTP/SRTP/ICE/QUIC + io_uring/XDP send-path + Pion adapter*  |
| **Public path (4 mirrors)** | `vasic-digital/helix-transport` on GitHub + GitLab + GitFlic + GitVerse                                                |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-iouring`, `helix-xdp`, `helix-network`, `helix-shm`, `helix-bench`                    |
| **External Go deps**        | `github.com/pion/webrtc/v4`, `github.com/pion/rtp`, `github.com/pion/srtp/v3`, `github.com/pion/ice/v4`, `github.com/quic-go/quic-go` |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `rtp-srtp-quic-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                          |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/10_chaos_network_partition/scenarios/01_quic_recovery_after_partition.scenario.yaml`                    |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **4** (deepest tier; imports 6 sibling submodules)                                                                |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-transport` is the **network transport** for HelixPlay's encoded media + control plane. Origin: [C37 §9](../../05_Video_Audio/12_Network_Transport.md). It exposes RTP / SRTP / ICE for the WebRTC-style path and QUIC for HelixPlay's custom-UDP-over-QUIC path; both paths leverage `helix-iouring` and `helix-xdp` for kernel-bypass send-side performance.

The submodule was introduced because transport logic spans Pion (WebRTC stack), quic-go, custom-UDP framing, and the kernel-bypass primitives. R-04 mandates one canonical landing.

---

## 2. Public API Surface

### 2.1 The `Transport` interface

```go
package transport

// Transport is the abstraction over WebRTC and custom-UDP.
type Transport interface {
    Connect(ctx context.Context, peer Peer) (Connection, error)
    Listen(addr string) (Listener, error)
    Close() error
}

func NewWebRTC(opts ...Option) (Transport, error)
func NewQUIC(opts ...Option) (Transport, error)
func NewCustomUDP(opts ...Option) (Transport, error)
func NewBest() Transport
```

### 2.2 The `Connection` and `Listener` types

```go
package transport

type Connection interface {
    SendVideo(rtp *rtp.Packet) error
    SendAudio(rtp *rtp.Packet) error
    SendDatagram(data []byte) error
    Close() error
    Stats() ConnectionStats
}

type Listener interface {
    Accept() (Connection, error)
    Close() error
}
```

### 2.3 The kernel-bypass send path

```go
package transport

// FastSend uses helix-iouring + helix-xdp for batched kernel-bypass
// send. Falls back to standard sendmsg on platforms without support.
func (c Connection) FastSend(packets []rtp.Packet) error
```

### 2.4 Configuration options

```go
package transport

type Option func(*config)

func WithDSCPMarking(marker *network.Marker) Option
func WithJitterBuffer(jb *network.JitterBuffer) Option
func WithFEC(fec *abr.FEC) Option
func WithKernelBypass(enabled bool) Option
func WithMaxPacketRate(pps int) Option
```

### 2.5 ICE + signalling helpers

```go
package transport

type ICEAgent struct { /* ... */ }
func NewICEAgent(stunServers, turnServers []string) (*ICEAgent, error)
func (a *ICEAgent) GatherCandidates(ctx context.Context) ([]ICECandidate, error)
```

### 2.6 Connection statistics

```go
package transport

type ConnectionStats struct {
    BytesSent       uint64
    PacketsSent     uint64
    BytesReceived   uint64
    PacketsReceived uint64
    RoundTripTimeMs int
    LossRate        float64
    KernelBypassUsed bool
}
```

### 2.7 Capability detection

```go
package transport

func DetectCapabilities() (Capabilities, error)
type Capabilities struct {
    WebRTCSupported    bool
    QUICSupported      bool
    KernelBypassSupported bool   // io_uring + AF_XDP available
    SRTPHardwareAccelerated bool
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

Per S01 §3.1's row #29: 6 sibling submodules:

- `helix-r18-safeexec`
- `helix-iouring` — for batched UDP sends.
- `helix-xdp` — for AF_XDP zero-copy.
- `helix-network` — for DSCP marking + jitter buffer + L4S.
- `helix-shm` — for zero-copy from encoder buffers.
- `helix-bench` — for transport-side latency measurement.

### 3.2 External (Go)

- `github.com/pion/webrtc/v4` — WebRTC stack.
- `github.com/pion/rtp` — RTP packet types.
- `github.com/pion/srtp/v3` — SRTP encryption.
- `github.com/pion/ice/v4` — ICE for NAT traversal.
- `github.com/quic-go/quic-go` — QUIC.

### 3.3 External (system)

- STUN / TURN server reachable for ICE; HelixPlay's deployment ships a Coturn instance per Constitution §4.

---

## 4. Container Build (S02 §3 lane: `rtp-srtp-quic-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + CAP_BPF + AF_XDP device passthrough (inherits helix-xdp's S02 §8.4 carve-out).

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
WebRTC + QUIC connection state machines; RTP packetisation correctness; SRTP encryption round-trips.

### 5.2 Integration
Real Pion webrtc.PeerConnection + ICE NAT traversal via Coturn.

### 5.3 E2E
helix-pipeline → helix-transport → real network → client decode.

### 5.4 Security
govulncheck + Snyk + Trivy; SRTP key-rotation correctness; malformed-RTP fuzzer.

### 5.5 Benchmarking
SendVideo p999 ≤ 200 µs (kernel-bypass) or ≤ 1 ms (standard); throughput ≥ 10 Gb/s with kernel-bypass.

### 5.6 Chaos
Toxiproxy partition; verify QUIC recovery within 1 s.

### 5.7 Stress
24-hour 4K120 stream; zero leak; LossRate < 0.5 %.

### 5.8 Smoke
30-second post-deploy: connect to localhost peer; send + receive a probe packet.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`10_chaos_network_partition/01_quic_recovery_after_partition.scenario` — full session with mid-session partition; verify QUIC's connection migration recovers without session loss.

---

## 6. Challenges Entry-Point (S03 §4 row #29)

**Topology:** `10_chaos_network_partition`. **Scenario:** `01_quic_recovery_after_partition.scenario.yaml`. **Why this scenario.** Partition recovery is the canonical transport robustness test; the scenario verifies QUIC's connection-migration prevents session loss. **Baseline:** session continues across partition; recovery within 1 second; no observable artefact at the user.

---

## 7. R-18 Inheritance

`helix-transport` imports `helix-r18-safeexec` for boundary subprocess invocations + transitively through helix-xdp's CAP_BPF carve-out. Hot path is cgo (Pion + quic-go) + io_uring/AF_XDP via the kernel-bypass send path.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on Pion v4 + quic-go ≥ v0.51 stability + per-mirror CI green for kernel-bypass scenarios.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                                              | Purpose                                                                |
|------------------------------------|---------------|----------------------------------------------------------|------------------------------------------------------------------------|
| `HELIX_TRANSPORT_TYPE`             | `auto`        | `webrtc`/`quic`/`custom-udp`/`auto`                      | Transport type.                                                        |
| `HELIX_TRANSPORT_KERNEL_BYPASS`    | `auto`        | `auto`/`true`/`false`                                    | Use io_uring + XDP when supported.                                     |
| `HELIX_TRANSPORT_STUN_SERVERS`     | `stun:stun.l.google.com:19302` | csv URLs                                       | STUN servers.                                                           |
| `HELIX_TRANSPORT_TURN_USERNAME`    | (operator)    | string                                                    | TURN credentials.                                                       |
| `HELIX_TRANSPORT_TURN_PASSWORD`    | (operator)    | string                                                    | TURN credentials.                                                       |
| `HELIX_TRANSPORT_MAX_PPS`          | `2000000`     | int [1000, 10000000]                                     | Packet-rate cap to prevent abuse.                                     |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| SendVideo (kernel-bypass)             | 100 µs   | 150 µs  | 200 µs  | io_uring + AF_XDP path.                                          |
| SendVideo (standard sendmsg)          | 400 µs   | 800 µs  | 1 ms    | Linux kernel stack.                                              |
| QUIC connection establish             | 100 ms   | 250 ms  | 500 ms  | Includes 0-RTT where available.                                  |
| WebRTC ICE candidate gather           | 1 s      | 2 s     | 4 s     | STUN + TURN exchanges.                                            |
| Throughput (kernel-bypass)            | 10 Gbps  | —       | —       | Line-rate dependent on NIC.                                      |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `transport: ErrICEFailed`                     | NAT traversal couldn't establish a path                        | Verify TURN server reachability; degrade to TCP fallback.                 |
| `transport: ErrSRTPKeyMismatch`               | DTLS-SRTP key derivation diverged                              | Restart connection; investigate clock skew.                                |
| `transport: ErrKernelBypassUnavailable`       | io_uring or XDP missing                                        | Falls back automatically; metric the fallback rate.                       |
| `transport: ErrPacketRateExceeded`            | Send rate > MAX_PPS cap                                        | Raise cap or backpressure encoder.                                         |
| `transport: ErrConnectionMigrationFailed`     | QUIC migration didn't pick up new path                         | Verify QUIC version + migration extension support.                         |

### 9.4 Migration from raw UDP send loops

A consumer migrating from `net.PacketConn.WriteTo`:

1. Replace with `t, _ := transport.NewBest()`.
2. Establish: `conn, _ := t.Connect(ctx, peer)`.
3. Send: `conn.SendVideo(rtp)`.
4. Subscribe to ICE / SRTP events.
5. Add OTLP span; metric for `ConnectionStats.LossRate`.

The migration is documented in `docs/migration-from-raw-udp.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_transport_bytes_sent_total`              | counter    | Bytes sent.                                                                  |
| `helix_transport_packets_sent_total`            | counter    | Packets sent.                                                                |
| `helix_transport_loss_rate`                     | gauge      | Observed RTP loss rate.                                                      |
| `helix_transport_rtt_ms`                        | gauge      | Round-trip time.                                                             |
| `helix_transport_kernel_bypass_active`          | gauge      | 1 if kernel-bypass active.                                                   |
| `helix_transport_ice_candidates_total`          | counter    | ICE candidates gathered, labelled `type={host, srflx, relay, prflx}`.       |
| `helix_transport_partitions_recovered_total`    | counter    | QUIC connection-migration recoveries.                                        |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — origin                            | Origin chapter; full Transport + Connection + ICE API.                               |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | helix-pipeline.Sink delegates to helix-transport.SendVideo/SendAudio.               |
| [C06 §6](../../03_Architecture/05_RealTime_APIs.md) — Real-Time APIs                       | Control-plane gRPC rides over the same transport (HTTP/3 path).                     |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-transport-A      | QUIC v2 (RFC 9369) — production-readiness vs QUIC v1?                                                          | C37 §9 next revision                                |
| OQ-transport-B      | TURN-over-TLS vs TURN-over-DTLS — operator deployment matrix?                                                  | C37 §9 next revision                                |
| OQ-transport-C      | Kernel-bypass adoption gate — automatic or operator-tunable?                                                   | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/12_Network_Transport.md`](../../05_Video_Audio/12_Network_Transport.md) §9 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §6 §7 | 1,218 | 2026-04-30 | catalog row #29; depth-4                |
| All 6 dependency descriptors                                                                                                                            | (this batch) | 2026-04-30 | direct deps                                    |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Transport + Connection state machines).                           |
| Integration    | Real Pion + ICE via Coturn; QUIC + WebRTC paths exercised.                                    |
| E2E            | Full pipeline → transport → client decode round-trip.                                         |
| Security       | govulncheck + Snyk + Trivy + SRTP key rotation + RTP fuzzer.                                  |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Toxiproxy partition; QUIC connection-migration recovers in 1 s.                              |
| Stress         | 24-hour 4K120 stream; LossRate < 0.5 %.                                                      |
| Smoke          | 30-second localhost connect + probe packet round-trip.                                        |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `10_chaos_network_partition/01_quic_recovery_after_partition` baseline-parity.                |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-transport.md` — 2026-04-30.

---

## 🎉 Catalog Complete

This descriptor is **row #29 of 29** in the canonical submodule catalog (per [S01 §3.1](../01_Submodule_Catalog.md#31-the-29-submodule-table)). The Submodules family is now fully populated: aggregation chapters S01–S04 + per-submodule descriptors for every submodule introduced by chapters C01–C37 §6 / §3 / §8 / §9.
