# `helix-abr` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-abr`                                                                                                            |
| **Origin chapter:section**  | [C33 §6](../../05_Video_Audio/08_ABR_FEC_Congestion.md) — *8-tier ABR ladder + GCC/SCReAM/SQP congestion control + FEC* |
| **Public path (4 mirrors)** | `vasic-digital/helix-abr` on GitHub + GitLab + GitFlic + GitVerse                                                      |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-network`                                                                              |
| **External Go deps**        | `github.com/pion/rtcp`, `github.com/klauspost/reedsolomon` (FEC)                                                        |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `abr-ladder-1.x` — builder `golang-builder`, runtime `distroless-static`                                            |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/12_burst_marketing_spike/scenarios/01_abr_ladder_descent_under_10x_burst.scenario.yaml`                 |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2** (depends on helix-network)                                                                                   |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-abr` implements **adaptive bitrate ladder descent / ascent** plus **congestion control** (Google Congestion Controller variant + SCReAM + experimental SQP) plus **forward error correction** (Reed-Solomon-based). Origin: [C33 §6](../../05_Video_Audio/08_ABR_FEC_Congestion.md). The submodule consumes `helix-network.L4SFeedback` plus jitter-buffer stats and decides:

1. Which ABR rung to encode at next.
2. How much FEC to add per packet.
3. Whether to descend / hold / ascend the ladder.

---

## 2. Public API Surface

### 2.1 The `Controller` type

```go
package abr

// Controller drives the ABR descent / ascent decision.
type Controller struct { /* ... */ }

func NewController(initialProfile codec.Profile, opts ...Option) *Controller
func (c *Controller) Update(rttMs int, lossRate, ceMarkRate float64) Decision
type Decision struct {
    Profile      codec.Profile  // next profile
    FECOverhead  float64        // 0.0–1.0 fraction
    RungChanged  bool
}
```

### 2.2 Congestion controller variants

```go
package abr

type CCVariant int
const (
    CCGoogleCongestionController CCVariant = iota   // GCC, default
    CCSCReAM
    CCSQP                                             // experimental
)

func NewCC(variant CCVariant, opts ...CCOption) CongestionController
type CongestionController interface {
    OnReceive(rttMs int, lossRate float64) (cwndBps int)
}
```

### 2.3 FEC encoder/decoder

```go
package abr

type FEC struct { /* ... */ }
func NewFEC(dataShards, parityShards int) (*FEC, error)
func (f *FEC) Encode(data []byte) (shards [][]byte, err error)
func (f *FEC) Decode(shards [][]byte) (data []byte, err error)
```

### 2.4 Configuration options

```go
package abr

type Option func(*config)

func WithCCVariant(v CCVariant) Option
func WithMinFECRatio(r float64) Option
func WithMaxFECRatio(r float64) Option
func WithLadderHysteresisMs(ms int) Option
```

### 2.5 Statistics

```go
package abr

type ABRStats struct {
    LadderTransitionsUp    uint64
    LadderTransitionsDown  uint64
    AverageBitrate         int
    FECOverheadAverage     float64
    CWNDBytesPerSecond     int
}
```

### 2.6 Capability detection

```go
package abr

type Capabilities struct {
    GCCAvailable    bool
    SCReAMAvailable bool
    SQPAvailable    bool
    FECMaxParity    int
}
```

### 2.7 The ladder reference

```go
package abr

// HelixDefaultLadder returns the C33 §6-mandated 8-tier ladder.
// Reuses helix-codec.CanonicalLadder underneath.
func HelixDefaultLadder() []codec.Profile
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations.
- `helix-network` — consumes L4SFeedback + jitter-buffer stats.

### 3.2 External (Go)

- `github.com/pion/rtcp` — RTCP feedback parsing.
- `github.com/klauspost/reedsolomon` — Reed-Solomon FEC.

### 3.3 External (system)

None.

---

## 4. Container Build (S02 §3 lane: `abr-ladder-1.x`)

**Builder:** `golang-builder`. **Runtime:** `distroless-static`. **Multi-arch:** `linux/amd64` + `linux/arm64`.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
GCC + SCReAM controller decision matrices; FEC encode/decode round-trips.

### 5.2 Integration
Real RTCP feedback ingest + Controller decision loop.

### 5.3 E2E
helix-abr → helix-codec.PickProfile → helix-encoder reconfigure.

### 5.4 Security
govulncheck + Snyk + Trivy.

### 5.5 Benchmarking
Controller.Update p999 ≤ 10 µs; FEC encode 1 MiB p999 ≤ 50 ms.

### 5.6 Chaos
Inject network impairment (Toxiproxy); verify ladder descends within 1-second budget.

### 5.7 Stress
24-hour ABR session with rolling impairment; zero leak; ladder transitions consistent.

### 5.8 Smoke
30-second post-deploy: feed RTCP probe; verify Controller.Update returns a decision.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`12_burst_marketing_spike/01_abr_ladder_descent_under_10x_burst.scenario` — 10× traffic burst over 5 minutes; verify ladder descends without freezing the stream.

---

## 6. Challenges Entry-Point (S03 §4 row #25)

**Topology:** `12_burst_marketing_spike`. **Scenario:** `01_abr_ladder_descent_under_10x_burst.scenario.yaml`. **Why this scenario.** ABR descent under burst is the canonical congestion test; the scenario verifies the descent doesn't freeze the stream. **Baseline:** ladder transitions match recorded sequence; stream maintains continuity (no >2-second gaps); FEC overhead within budget.

---

## 7. R-18 Inheritance

`helix-abr` imports `helix-r18-safeexec` for boundary subprocess invocations. Hot path is pure-Go decision logic.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on Controller + FEC stability across the canonical 8-tier ladder.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                            | Purpose                                                                |
|------------------------------------|---------------|----------------------------------------|------------------------------------------------------------------------|
| `HELIX_ABR_CC_VARIANT`             | `gcc`         | `gcc`/`scream`/`sqp`                   | Congestion-controller choice.                                          |
| `HELIX_ABR_MIN_FEC_RATIO`          | `0.05`        | float [0.0, 0.5]                       | Minimum FEC overhead.                                                  |
| `HELIX_ABR_MAX_FEC_RATIO`          | `0.30`        | float [0.0, 0.5]                       | Maximum FEC overhead under heavy loss.                                |
| `HELIX_ABR_HYSTERESIS_MS`          | `1000`        | int [100, 10000]                       | Hold time before re-ascending after a descent.                       |
| `HELIX_ABR_INITIAL_RUNG`           | `top`         | `top`/`mid`/`bottom`                   | Where to start the ladder.                                            |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Controller.Update                     | 4 µs     | 8 µs    | 10 µs   | Decision computation.                                             |
| FEC encode (RTP-shape)                | 30 µs    | 80 µs   | 100 µs  | k=10, n=15 typical.                                               |
| FEC decode (1 missing shard)          | 80 µs    | 200 µs  | 400 µs  | Reed-Solomon recovery.                                            |
| Ladder transition                     | 2 µs     | 4 µs    | 6 µs    | Profile lookup.                                                  |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `abr: ErrNoCapacity`                          | Network bandwidth dropped below the lowest rung's floor       | Surface to client; require network upgrade or session pause.              |
| `abr: ErrThrashing`                           | Ladder oscillates between rungs                                | Increase `HELIX_ABR_HYSTERESIS_MS`.                                       |
| `abr: ErrFECOverflow`                         | FEC overhead at maximum but loss continues                     | Audit network; may indicate a partition rather than congestion.           |
| `abr: ErrCWNDStuck`                           | Congestion controller stuck at minimum                         | Reset CC; investigate L4S CE-mark rate.                                   |

### 9.4 Migration from fixed-bitrate streaming

A consumer migrating from fixed encode bitrate:

1. Replace `encoder.SetBitrate(constant)` with `controller := abr.NewController(initialProfile)`.
2. On every RTCP feedback: `decision := controller.Update(rtt, loss, ceMarkRate); encoder.Reconfigure(decision.Profile)`.
3. Pre-encode FEC: `shards, _ := fec.Encode(packet)`.
4. Add OTLP span; metric for `ABRStats.LadderTransitionsDown`.

The migration is documented in `docs/migration-from-fixed-bitrate.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_abr_ladder_transitions_total`            | counter    | Up + down transitions, labelled `direction`.                                |
| `helix_abr_average_bitrate_bps`                 | gauge      | Currently-encoded bitrate.                                                   |
| `helix_abr_fec_overhead_ratio`                  | gauge      | Current FEC overhead.                                                        |
| `helix_abr_cwnd_bps`                            | gauge      | Congestion controller's CWND.                                                |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C33 §6](../../05_Video_Audio/08_ABR_FEC_Congestion.md) — origin                           | Origin chapter; full Controller + FEC + CC API.                                      |
| [C19 §6](../../04_Latency/05_UltraLowLatency_Network_Protocols.md) — Network              | helix-network.L4SFeedback consumed.                                                 |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates ABR decisions + encoder reconfig.                             |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport applies FEC to outgoing RTP.                                              |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-abr-A            | SQP — production-readiness vs experimental?                                                                    | C33 §6 next revision                                |
| OQ-abr-B            | FEC strategy — RS vs Raptor / RaptorQ for higher loss rates?                                                  | C33 §6 next revision                                |
| OQ-abr-C            | Per-tenant ABR profile customisation — operator-tunable hysteresis?                                            | C11 §6 + C33 §6 next revisions                      |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/08_ABR_FEC_Congestion.md`](../../05_Video_Audio/08_ABR_FEC_Congestion.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-network.md`](helix-network.md) | (this batch) | 2026-04-30 | direct dependency                              |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Controller + CC + FEC).                                            |
| Integration    | RTCP feedback ingest + Controller decision loop.                                              |
| E2E            | abr → codec.PickProfile → encoder reconfigure round-trip.                                     |
| Security       | govulncheck + Snyk + Trivy.                                                                   |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Toxiproxy impairment; ladder descends within 1 s.                                             |
| Stress         | 24-hour rolling impairment; consistent transitions.                                           |
| Smoke          | 30-second RTCP probe + decision verification.                                                 |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `12_burst_marketing_spike/01_abr_ladder_descent_under_10x_burst` baseline-parity.             |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-abr.md` — 2026-04-30.
