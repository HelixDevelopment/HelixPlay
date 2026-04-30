# `helix-gpu-direct` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-gpu-direct`                                                                                                     |
| **Origin chapter:section**  | [C18 §3](../../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md) — *GPUDirect RDMA + p2p PCIe + zero-copy GPU memory* |
| **Public path (4 mirrors)** | `vasic-digital/helix-gpu-direct` on GitHub + GitLab + GitFlic + GitVerse                                                |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `github.com/NVIDIA/go-nvml`, custom CUDA driver-API cgo bindings, `golang.org/x/sys/unix`                              |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `gpu-direct-rdma-1.x` — builder `golang-builder-gpu`, runtime `distroless-cuda`                                     |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/06_4k_120hz_hdr_dolby_vision/scenarios/01_gpu_direct_rdma_under_4k120_session.scenario.yaml`            |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-gpu-direct` exposes **GPUDirect RDMA + peer-to-peer PCIe + zero-copy GPU memory access** for HelixPlay's encode-to-network pipeline. Origin: [C18 §3](../../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md). The submodule wraps NVIDIA's `cuMemAlloc` + `cuMemImportFromShareableHandle` plus the kernel-side `nvidia-peermem` driver path so an encoded frame leaves the GPU encoder's video memory and lands directly in an RDMA-capable NIC's transmit queue without traversing system RAM.

The submodule was introduced to consolidate GPU-direct integration logic that ranges across NVIDIA, AMD, and Intel ecosystems. Each vendor's API (CUDA / ROCm / Level Zero) requires different setup; the submodule's API normalises the common operations (allocate GPU memory, export a handle, import on the network side, RDMA-register).

---

## 2. Public API Surface

### 2.1 The `Buffer` type

```go
package gpudirect

// Buffer is a GPU-resident buffer with a kernel-visible RDMA-export handle.
type Buffer struct { /* ... */ }

func NewBuffer(size int, opts ...Option) (*Buffer, error)
func (b *Buffer) Close() error
func (b *Buffer) ExportHandle() (Handle, error)
func (b *Buffer) Size() int
func (b *Buffer) DeviceID() int
```

### 2.2 The `Handle` type (cross-process / cross-device)

```go
package gpudirect

// Handle is an opaque, transferable handle to a GPU buffer.
type Handle struct { /* ... */ }

func ImportHandle(h Handle) (*Buffer, error)
func (h Handle) Marshal() ([]byte, error)
func UnmarshalHandle(data []byte) (Handle, error)
```

### 2.3 RDMA registration

```go
package gpudirect

// RegisterRDMA pins the buffer for RDMA access via the supplied PD.
func (b *Buffer) RegisterRDMA(pd RDMAProtectionDomain) (*MemoryRegion, error)

type MemoryRegion struct { /* ... */ }
func (mr *MemoryRegion) Lkey() uint32
func (mr *MemoryRegion) Rkey() uint32
func (mr *MemoryRegion) Close() error
```

### 2.4 Vendor capability detection

```go
package gpudirect

func KernelCaps() (Capabilities, error)
type Capabilities struct {
    NVIDIAGPUDirect       bool   // nvidia-peermem loaded
    AMDROCmGPUDirect      bool   // amdgpu kernel module + ROCm runtime
    IntelLevelZero        bool   // Level Zero runtime
    P2PPCIeAvailable      bool   // BIOS exposes ACS / above-4G decoding off
    NumaTopology          string // NUMA placement string for NIC + GPU pairing
}
```

### 2.5 Option types

```go
package gpudirect

type Option func(*config)

func WithDevice(deviceID int) Option       // explicit GPU selection
func WithPinned() Option                    // pin pages to prevent migration
func WithNUMAPlacement(node int) Option     // align with the NIC's NUMA node
func WithRDMAVerbs() Option                 // bring up RDMA verbs interface
```

### 2.6 The `Pool` type for reusable buffers

```go
package gpudirect

type Pool struct { /* ... */ }
func NewPool(bufferSize int, capacity int, opts ...PoolOption) *Pool
func (p *Pool) Acquire(ctx context.Context) (*Buffer, error)
func (p *Pool) Release(buf *Buffer)
func (p *Pool) Stats() PoolStats
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`nvidia-smi -q` for diagnostics; `lspci -vvv` for PCIe topology).

### 3.2 External (Go)

- `github.com/NVIDIA/go-nvml` — NVML bindings for device enumeration + status.
- Custom cgo bindings to the CUDA driver-API (`libcuda.so`); pinned to CUDA ≥ 12.6 toolkit headers.
- `golang.org/x/sys/unix` — Linux syscall surface for `ioctl` against `/dev/nvidia0`.

### 3.3 External (system)

- NVIDIA driver ≥ 560.x with `nvidia-peermem` module loaded; or AMD ROCm ≥ 6.2; or Intel Level Zero ≥ 1.18.
- BIOS configured with PCIe peer-to-peer support: ACS disabled on the relevant root-port, "Above 4G Decoding" enabled, "Re-Size BAR" enabled where supported.
- RDMA-capable NIC: ConnectX-6/7 (Mellanox), Intel E810, or equivalent.

---

## 4. Container Build (S02 §3 lane: `gpu-direct-rdma-1.x`)

**Builder:** `golang-builder-gpu` (cgo + CUDA toolkit). **Runtime:** `distroless-cuda`. **Multi-arch:** `linux/amd64` only (arm64 GPU support deferred — most production GPUs are amd64-host-attached). **Hardening:** standard S02 §8.2 + `--device=/dev/nvidia0` + `--device=/dev/nvidia-uvm` + `--device=/dev/nvidiactl` + `--device=/dev/infiniband/uverbs0` (RDMA verbs). Per Constitution §11.5.3 + S02 §8.4, device passthrough is the explicit carve-out; `--privileged` remains denied.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Mock GPU buffer lifecycle + handle export / import.

### 5.2 Integration
Real GPU + real RDMA NIC; allocate buffer; register; verify RDMA send between two processes on the same host.

### 5.3 E2E
helix-gpu-direct → helix-encoder GPU buffer → helix-transport RDMA send.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: attempt cross-tenant buffer access (must fail RDMA verbs auth).

### 5.5 Benchmarking
RDMA send throughput ≥ 90 % of NIC line-rate (e.g. 90 Gb/s on a 100 GbE NIC); GPU-to-NIC latency ≤ 5 µs.

### 5.6 Chaos
PCIe-link error injection; verify graceful fallback to system-RAM-staged path.

### 5.7 Stress
24-hour 4K120 RDMA send loop; zero buffer leak; zero PCIe error escalation.

### 5.8 Smoke
30-second post-deploy: allocate 8 MiB GPU buffer; export handle; verify size.

### 5.9 Full Automation
§5.1–§5.8 in CI matrix.

### 5.10 Challenges
`06_4k_120hz_hdr_dolby_vision/01_gpu_direct_rdma_under_4k120_session.scenario` — full 4K120 HDR session with GPUDirect path active; verify frames go GPU → NIC without staging through system RAM (verified via `perf` + `nvprof`).

---

## 6. Challenges Entry-Point (S03 §4 row #10)

**Topology:** `06_4k_120hz_hdr_dolby_vision`. **Scenario:** `01_gpu_direct_rdma_under_4k120_session.scenario.yaml`. **Why this scenario.** GPUDirect's value prop is realised only at high frame rates where system-RAM staging would cost cycles; 4K120 with HDR is the canonical workload. **Baseline:** end-to-end frame latency ≤ 2 ms; system-RAM bandwidth utilisation ≤ 5 % of frame size (proves staging is bypassed).

---

## 7. R-18 Inheritance

`helix-gpu-direct` imports `helix-r18-safeexec` for boundary subprocess invocations (NVIDIA NVML / lspci diagnostics). Hot path (CUDA driver-API + ioctl on /dev/nvidia*) is pure cgo + syscall. Device-passthrough capability set is a documented carve-out (S02 §8.4); `--privileged` remains denied.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation tracks NVIDIA driver / ROCm / Level Zero release cadence; `v1.0.0` requires green CI on the operator's target driver version matrix.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                | Purpose                                                                |
|------------------------------------|---------------|-----------------------------|------------------------------------------------------------------------|
| `HELIX_GPUDIRECT_VENDOR`           | `auto`        | `nvidia` / `amd` / `intel` / `auto` | Force vendor or auto-detect.                                  |
| `HELIX_GPUDIRECT_DEVICE_ID`        | `0`           | int                         | GPU device index.                                                       |
| `HELIX_GPUDIRECT_NUMA_NODE`        | `auto`        | int / `auto`                | NUMA node alignment; auto picks the NIC's node.                       |
| `HELIX_GPUDIRECT_FALLBACK_ENABLED` | `true`        | bool                        | Fall back to system-RAM staging if GPUDirect unavailable.             |
| `HELIX_GPUDIRECT_PINNED_PAGES`     | `true`        | bool                        | Pin host pages used for staging (when fallback active).               |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Buffer allocation                     | 100 µs   | 250 µs  | 500 µs  | cuMemAlloc + ioctl for handle.                                    |
| RDMA register                         | 50 µs    | 120 µs  | 200 µs  | ibv_reg_mr against the GPU buffer.                                |
| GPU-to-NIC throughput (100 GbE)       | 90 Gbps  | —       | —       | Line-rate minus protocol overhead.                                |
| GPU-to-NIC frame latency (4K HDR)     | 1.5 ms   | 1.8 ms  | 2 ms    | End-to-end including RDMA wire transit.                          |
| Pool acquire (warm)                   | 2 µs     | 5 µs    | 10 µs   | LRU lookup; no allocation.                                        |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `gpudirect: ErrPeerMemNotLoaded`              | nvidia-peermem kernel module missing                           | Load via `modprobe nvidia-peermem` (operator-side; documented in runbook).|
| `gpudirect: ErrACSEnabled`                    | PCIe ACS blocking peer-to-peer                                 | Disable ACS on the relevant root-port via BIOS or boot-time kernel param. |
| `gpudirect: ErrVendorNotSupported`            | GPU vendor not in vendor list                                  | Verify GPU; HelixPlay supports NVIDIA + AMD + Intel.                      |
| `gpudirect: ErrFallback`                      | GPUDirect path failed; using system-RAM staging               | Investigate NIC + GPU NUMA placement; verify peer-mem module.             |
| `gpudirect: ErrRDMARegistrationFailed`        | NIC verbs reject the GPU memory region                         | Verify NIC supports peer-mem; check verbs library version.                |

### 9.4 Migration from staged GPU-to-network path

A consumer migrating from GPU encode → cudaMemcpy → system RAM → send():

1. Verify operator deployment supports GPUDirect (kernel module + BIOS config).
2. Replace `cudaMemcpy(host, device, size, DeviceToHost)` with `buf.ExportHandle()`.
3. Replace `send(socket, host, size)` with `mr := buf.RegisterRDMA(pd); rdmaSend(qp, mr.Lkey(), ...)`.
4. Remove the staging buffer; track GPU buffer lifecycle via `gpudirect.Pool`.
5. Add OTLP span around `RegisterRDMA` + RDMA send; metric for fallback rate (alert: > 0.1 %).

The migration is documented in `docs/migration-from-staged-gpu.md`. Be aware: GPUDirect's BIOS prerequisites are operationally significant — every operator host needs verification before this submodule produces value.

#### Operator pre-deployment checklist

Operators provisioning a HelixPlay host for GPUDirect MUST verify, before the host enters service:

1. **Kernel modules**: `lsmod | grep nvidia-peermem` returns a row.
2. **BIOS — Above 4G Decoding**: enabled (vendor-specific path; documented in operator runbook).
3. **BIOS — PCIe ACS**: disabled on the GPU + NIC root-port pair (otherwise peer-to-peer is blocked).
4. **NIC + GPU NUMA topology**: both devices on the same NUMA node (verify via `lstopo`).
5. **RDMA verbs**: `ibv_devices` lists the NIC; `ibv_devinfo` shows transport=IB or RoCE.

A future `helixctl gpudirect-check` command (OQ-gpudirect-C) will automate this checklist.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_gpudirect_alloc_total`                   | counter    | Buffer allocations, labelled `result`, `device_id`.                          |
| `helix_gpudirect_alloc_latency_seconds`         | histogram  | Allocation latency.                                                          |
| `helix_gpudirect_rdma_send_total`               | counter    | RDMA sends, labelled `result`, `nic_id`.                                     |
| `helix_gpudirect_fallback_total`                | counter    | Fallback to system-RAM staging events; alert if rate > 0.1 %.                |
| `helix_gpudirect_throughput_bytes`              | counter    | Cumulative bytes pushed via GPUDirect.                                       |
| `helix_gpudirect_pcie_errors_total`             | counter    | PCIe link errors observed; > 0 sustained is a hardware concern.             |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C18 §3](../../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md) — origin                | Origin chapter; full GPUDirect / RDMA API.                                           |
| [C27 §6](../../05_Video_Audio/02_Hardware_Encoders.md) — Hardware Encoders                 | Encoder allocates GPU buffers; output handle goes to gpu-direct for export.          |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport receives RDMA-registered memory regions and submits sends.                |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-gpudirect-A      | AMD ROCm GPUDirect — GA timeline + NIC compat matrix?                                                          | C18 §3 next revision                                |
| OQ-gpudirect-B      | Intel Level Zero — production-readiness for streaming workloads?                                              | C18 §3 next revision                                |
| OQ-gpudirect-C      | BIOS prerequisite checklist — should the submodule ship a `helixctl gpudirect-check` command?                  | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md`](../../04_Latency/04_GPU_Direct_and_Hardware_Pipelines.md) §3 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #10                                 |
| [`../02_Containers_Submodule.md`](../02_Containers_Submodule.md) §8.4 |   627 | 2026-04-30 | device-passthrough carve-out                   |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Buffer + Handle + RegisterRDMA paths).                            |
| Integration    | Real GPU + RDMA NIC; cross-process handle import/export.                                      |
| E2E            | encoder → gpu-direct → transport RDMA send roundtrip.                                          |
| Security       | govulncheck + Snyk + Trivy + cross-tenant access denial.                                      |
| Benchmarking   | RDMA send ≥ 90 % NIC line-rate; latency ≤ 2 ms p999.                                          |
| Chaos          | PCIe error injection; fallback path engages cleanly.                                          |
| Stress         | 24-hour 4K120 RDMA loop; zero leak.                                                           |
| Smoke          | 30-second 8 MiB allocation + handle export.                                                   |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `06_4k_120hz_hdr_dolby_vision/01_gpu_direct_rdma_under_4k120_session` baseline-parity.        |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-gpu-direct.md` — 2026-04-30.
