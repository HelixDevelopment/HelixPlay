# `helix-xdp` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-xdp`                                                                                                            |
| **Origin chapter:section**  | [C16 §6](../../04_Latency/02_io_uring_and_Kernel_Bypass.md) — *XDP eBPF kernel-bypass send/redirect path*             |
| **Public path (4 mirrors)** | `vasic-digital/helix-xdp` on GitHub + GitLab + GitFlic + GitVerse                                                       |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `github.com/cilium/ebpf`, `github.com/asavie/xdp`, `golang.org/x/sys/unix`                                              |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `xdp-ebpf-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                              |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/07_kernel_bypass_send_path/scenarios/02_xdp_redirect_under_packet_burst.scenario.yaml`                  |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-xdp` is the **XDP eBPF kernel-bypass packet-redirect primitive** for HelixPlay's send path. It loads a tiny eBPF program at the network-driver hook point that redirects HelixPlay's UDP egress packets straight to a userspace AF_XDP socket, bypassing the kernel's qdisc + IP stack for a 5–10× throughput improvement on the 4K120 send rate. Origin: [C16 §6](../../04_Latency/02_io_uring_and_Kernel_Bypass.md).

The submodule was introduced to consolidate XDP/eBPF/AF_XDP plumbing — three separate libraries (cilium/ebpf, asavie/xdp, the kernel's bpftool) interact in non-obvious ways, and earlier projects re-implemented the integration repeatedly with subtle bugs. R-04 mandates one canonical landing.

---

## 2. Public API Surface

### 2.1 The `Redirect` type

```go
package xdp

// Redirect attaches an XDP eBPF program to a network interface
// that redirects matching packets to the supplied AF_XDP socket.
type Redirect struct { /* ... */ }

func NewRedirect(ifname string, opts ...Option) (*Redirect, error)
func (r *Redirect) Attach(socket *AFXDPSocket) error
func (r *Redirect) Detach() error
func (r *Redirect) Stats() RedirectStats
```

### 2.2 The `AFXDPSocket` type

```go
package xdp

// AFXDPSocket is a userspace zero-copy socket that receives XDP-redirected packets.
type AFXDPSocket struct { /* ... */ }

func NewAFXDPSocket(queueID int, opts ...SocketOption) (*AFXDPSocket, error)
func (s *AFXDPSocket) Send(buf []byte) (int, error)
func (s *AFXDPSocket) Recv(buf []byte) (int, error)
func (s *AFXDPSocket) Close() error
```

### 2.3 Capability detection

```go
package xdp

func KernelCaps() (Capabilities, error)
type Capabilities struct {
    XDPSupported       bool
    XDPGenericMode     bool   // SKB-mode fallback
    XDPNativeMode      bool   // driver-native mode (preferred)
    AFXDPZeroCopy      bool   // requires driver support (mlx5, ice, i40e)
}
```

### 2.4 The eBPF program contract

```go
package xdp

// The compiled eBPF object exports the following symbols. The Go
// loader (cilium/ebpf) verifies presence at attach time.
//
// Symbols expected in the .o file:
//   - "xdp_redirect_helix"  XDP program handle, type BPF_PROG_TYPE_XDP
//   - "helix_redirect_map"  BPF_MAP_TYPE_XSKMAP, max 64 queues
//   - "helix_stats"         BPF_MAP_TYPE_PERCPU_ARRAY, redirect counters
//   - "helix_filter"        BPF_MAP_TYPE_HASH, 5-tuple filter (optional)
type EBPFContract struct {
    ProgramName  string
    XSKMapName   string
    StatsMapName string
    FilterName   string  // empty if no filter
}

var DefaultContract = EBPFContract{
    ProgramName: "xdp_redirect_helix",
    XSKMapName:  "helix_redirect_map",
    StatsMapName: "helix_stats",
    FilterName:  "helix_filter",
}
```

### 2.5 Configuration options

```go
package xdp

type Option func(*config)

// WithMode forces a specific XDP mode (native, generic, hardware).
// Default is auto-detect.
func WithMode(mode Mode) Option

// WithZeroCopy enables AF_XDP zerocopy if the driver supports it.
func WithZeroCopy() Option

// WithFilter installs a 5-tuple filter so only matching packets are
// redirected; non-matching packets fall through to the kernel stack.
func WithFilter(rules []FilterRule) Option

type FilterRule struct {
    SrcIP   netip.Prefix
    DstIP   netip.Prefix
    SrcPort uint16
    DstPort uint16
    Proto   uint8         // unix.IPPROTO_UDP, unix.IPPROTO_TCP
}
```

### 2.6 Statistics

```go
package xdp

type RedirectStats struct {
    PacketsRedirected uint64
    PacketsDropped    uint64
    BytesRedirected   uint64
    QueueOverflows    uint64
    LastUpdated       time.Time
}

func (r *Redirect) Stats() RedirectStats
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (running `bpftool prog list`, `ip link set xdp` for diagnostic-only ops; the production-path bpf_load goes through `cilium/ebpf` directly).

### 3.2 External (Go)

- `github.com/cilium/ebpf` — high-quality, well-maintained Go eBPF binding. Pinned to ≥ v0.18.0 for `bpf2go` codegen stability.
- `github.com/asavie/xdp` — AF_XDP socket binding. Pinned to ≥ v0.3.4.
- `golang.org/x/sys/unix` — Linux syscall surface.

### 3.3 External (system)

- Linux kernel ≥ 5.4 (XDP) or ≥ 5.10 (AF_XDP zero-copy on more drivers).
- Driver supporting XDP-native + AF_XDP zerocopy for full performance: mlx5, ice, i40e, ixgbe, bnxt_en. Other drivers fall back to XDP-generic (slower but still functional).

---

## 4. Container Build (S02 §3 lane: `xdp-ebpf-1.x`)

**Builder:** `golang-builder-cgo` (cgo for eBPF program loading). **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + `--cap-add=SYS_ADMIN` + `--cap-add=NET_ADMIN` + `--cap-add=BPF` (per Constitution §11.5.3 — Section 8.4 codec exception extends only to GPU device passthrough; this submodule's BPF capability is the second carve-out for the kernel-bypass send path).

**Carve-out justification.** XDP program loading requires CAP_BPF since Linux 5.8 (formerly CAP_SYS_ADMIN). The carve-out is documented inline at C16 §6's R-18 surface analysis; it does NOT enable `--privileged` (which would also unlock `pm-suspend`). The deny-list at the command-level (S01 §7) remains in force.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
eBPF program loader unit tests with `cilium/ebpf`'s test fixtures.

### 5.2 Integration
Real veth pair; attach XDP program; verify packet redirection.

### 5.3 E2E
helix-xdp → AF_XDP socket → helix-transport send loop.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: attempt to load an unverified eBPF program (must fail kernel verifier).

### 5.5 Benchmarking
Packet redirect throughput ≥ 14 Mpps (line-rate 10 GbE on amd64); zero-copy receive ≥ 8 Mpps.

### 5.6 Chaos
Driver-mode transitions (XDP-native ↔ XDP-generic); verify graceful fallback.

### 5.7 Stress
24-hour run at 1 Mpps; verify no eBPF map drift, no fd leak.

### 5.8 Smoke
30-second post-deploy: attach XDP program to loopback; send + receive a probe packet.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`07_kernel_bypass_send_path/02_xdp_redirect_under_packet_burst.scenario` — burst rate (10× steady state for 10 s); verify no dropped packets at the XDP hook, only at the userspace consumer.

---

## 6. Challenges Entry-Point (S03 §4 row #08)

**Topology:** `07_kernel_bypass_send_path`. **Scenario:** `02_xdp_redirect_under_packet_burst.scenario.yaml`. **Why this scenario.** XDP's value prop is *not dropping under burst*; the scenario specifically targets that property. **Baseline:** zero kernel drops; userspace-side queue depth oscillates within p999 ≤ 16 KiB; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-xdp` imports `helix-r18-safeexec` for diagnostic-only subprocess invocations (`bpftool` for debug). Hot path (XDP attach, AF_XDP send/recv) is pure cgo + syscall. The CAP_BPF carve-out (§4) is documented as an R-18 §11.5.3 named exception rather than a relaxation.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the target deployment kernel matrix; the capability-detection fallback keeps the submodule usable across kernel-version drift.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_XDP_INTERFACE`              | `eth0`        | string                  | Network interface to attach XDP to.                                    |
| `HELIX_XDP_MODE`                   | `auto`        | `native` / `generic` / `auto` | Force a mode or auto-detect.                                       |
| `HELIX_XDP_AFXDP_QUEUE`            | `0`           | int                     | RX queue ID to bind the AF_XDP socket to.                             |
| `HELIX_XDP_ZEROCOPY`               | `auto`        | `auto` / `true` / `false`| Use AF_XDP zerocopy mode when driver supports it.                    |
| `HELIX_XDP_PROG_PATH`              | `/usr/lib/helix/xdp.o` | filesystem path | Path to the compiled eBPF program (set at build time).               |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| XDP redirect (per packet)             | 80 ns    | 200 ns  | 400 ns  | At driver hook; no skb allocation.                                |
| AF_XDP send                           | 200 ns   | 600 ns  | 1.5 µs  | Zero-copy from userspace; ring write only.                       |
| AF_XDP recv                           | 200 ns   | 600 ns  | 1.5 µs  | Symmetric to send.                                                |
| Throughput (10 GbE, amd64)            | 14 Mpps  | —       | —       | Line-rate; matches mlx5 + AF_XDP zerocopy datasheet.             |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                      | Remediation                                                                |
|-----------------------------------------------|------------------------------------------------------------|----------------------------------------------------------------------------|
| `xdp: ErrCapBPF`                              | Container lacks CAP_BPF capability                          | Add `--cap-add=BPF` per §4; verify with `capsh --print`.                   |
| `xdp: ErrModeNotSupported`                    | Driver does not support XDP-native                          | Set `HELIX_XDP_MODE=generic` (slower) or use a supported NIC.              |
| `xdp: ErrAFXDPZeroCopyDenied`                 | Driver does not support AF_XDP zero-copy                    | Set `HELIX_XDP_ZEROCOPY=false`; performance falls to copy-mode (~3× slower).|
| `xdp: ErrProgVerifierRejected`                | eBPF verifier rejected the program (post-kernel-upgrade?)  | Re-compile against current kernel headers; file an upstream issue.        |
| `xdp: ErrInterfaceBusy`                       | Another XDP program already attached                        | Detach the prior program (`ip link set <if> xdp off`); investigate.       |

### 9.4 Migration from kernel-stack UDP send

A consumer migrating from `net.PacketConn`-based UDP send:

1. Provision the host: ensure CAP_BPF + driver supports XDP.
2. At startup: load eBPF program via `xdp.NewRedirect(...)`; create AF_XDP socket via `xdp.NewAFXDPSocket(...)`.
3. Replace `conn.WriteTo(buf, addr)` with `socket.Send(buf)`; the eBPF program tags egress destinations.
4. Add CAP_BPF check at startup; surface clear error if missing.
5. Add OTLP span around `Send()`; metric for kernel-drop counter (must remain zero).

The migration is documented in `docs/migration-from-kernel-udp.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_xdp_redirect_total`                      | counter    | Packets redirected by the eBPF program, labelled `result={ok, drop}`.        |
| `helix_xdp_afxdp_send_total`                    | counter    | AF_XDP sends, labelled `result={ok, retry, err}`.                            |
| `helix_xdp_afxdp_recv_total`                    | counter    | AF_XDP receives.                                                             |
| `helix_xdp_kernel_drops_total`                  | counter    | Kernel-side drops at the XDP hook (alert: > 0 sustained).                   |
| `helix_xdp_mode`                                | gauge      | Current XDP mode (1=native, 2=generic, 0=disabled).                         |
| `helix_xdp_zerocopy_active`                     | gauge      | 1 if AF_XDP zerocopy active, 0 otherwise.                                   |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C16 §6](../../04_Latency/02_io_uring_and_Kernel_Bypass.md) — origin                       | Origin chapter; full XDP / AF_XDP API.                                               |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport's high-throughput send path uses AF_XDP zerocopy.                          |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-xdp-A            | Multi-queue AF_XDP — when does HelixPlay's typical workload benefit from queue scaling?                        | C16 §6 next revision                                |
| OQ-xdp-B            | XDP-native mode upstream availability — driver support matrix re-check cadence?                                | C16 §6 next revision                                |
| OQ-xdp-C            | CAP_BPF carve-out documentation — should it appear in Constitution §11.5.3 inventory explicitly?               | Constitution §15 *Amendment Procedure*              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/02_io_uring_and_Kernel_Bypass.md`](../../04_Latency/02_io_uring_and_Kernel_Bypass.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #08                                 |
| [`../02_Containers_Submodule.md`](../02_Containers_Submodule.md) §8.4 |   627 | 2026-04-30 | CAP_BPF carve-out documentation                |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Redirect lifecycle + AFXDPSocket).                                |
| Integration    | Real veth pair; attach + redirect + verify.                                                   |
| E2E            | helix-xdp → AF_XDP → helix-transport send roundtrip.                                          |
| Security       | govulncheck + Snyk + Trivy + verifier-rejection test.                                         |
| Benchmarking   | Throughput ≥ 14 Mpps line-rate; redirect p999 ≤ 400 ns.                                       |
| Chaos          | Mode-transition under load.                                                                   |
| Stress         | 24-hour 1 Mpps; zero leak.                                                                    |
| Smoke          | 30-second loopback redirect probe.                                                            |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `07_kernel_bypass_send_path/02_xdp_redirect_under_packet_burst` baseline-parity.              |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-xdp.md` — 2026-04-30.
