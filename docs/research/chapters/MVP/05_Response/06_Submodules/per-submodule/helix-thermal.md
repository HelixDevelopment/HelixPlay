# `helix-thermal` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-thermal`                                                                                                        |
| **Origin chapter:section**  | [C34 §6](../../05_Video_Audio/09_Thermal_and_GPU_Balancing.md) — *NVML/ADL/Level Zero + DVFS-aware quality scaling*    |
| **Public path (4 mirrors)** | `vasic-digital/helix-thermal` on GitHub + GitLab + GitFlic + GitVerse                                                  |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `github.com/NVIDIA/go-nvml`, AMD ADL cgo bindings, Intel Level Zero cgo                                                 |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `thermal-dvfs-1.x` — builder `golang-builder-gpu`, runtime `distroless-cuda`                                         |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/09_thermal_throttling_dvfs/scenarios/01_thermal_throttling_quality_scaling.scenario.yaml`               |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-thermal` is the **GPU thermal awareness + DVFS-aware quality scaling** primitive. Origin: [C34 §6](../../05_Video_Audio/09_Thermal_and_GPU_Balancing.md). The submodule polls per-vendor thermal telemetry (NVIDIA NVML / AMD ADL / Intel Level Zero), tracks GPU temperature + clock + utilisation, and provides quality-scaling recommendations to `helix-pipeline` so the encoder profile is throttled before the GPU hits its TJunction limit.

---

## 2. Public API Surface

### 2.1 The `Monitor` type

```go
package thermal

// Monitor polls GPU telemetry and emits scaling recommendations.
type Monitor struct { /* ... */ }

func NewMonitor(deviceID int, opts ...Option) (*Monitor, error)
func (m *Monitor) Start(ctx context.Context) error
func (m *Monitor) Stop() error
func (m *Monitor) Telemetry() Telemetry
func (m *Monitor) Recommend() Recommendation
```

### 2.2 Telemetry + recommendation

```go
package thermal

type Telemetry struct {
    TemperatureC      float64
    ClockMHz          int
    UtilisationPct    int
    PowerW            float64
    Throttling        ThrottleReason   // NoThrottle, ThermalSlow, ThermalCritical, PowerLimit
    LastUpdated       time.Time
}

type ThrottleReason int
const (
    NoThrottle ThrottleReason = iota
    ThermalSlow
    ThermalCritical
    PowerLimit
)

type Recommendation struct {
    ScalingFactor   float64    // 0.0–1.0 (1.0 = no scaling, 0.5 = half quality)
    Reason          string
    UpdatedAt       time.Time
}
```

### 2.3 Vendor capability detection

```go
package thermal

func DetectVendors() ([]Vendor, error)
type Vendor int
const (
    VendorNVIDIA Vendor = iota
    VendorAMD
    VendorIntel
)

type Capabilities struct {
    NVMLAvailable      bool
    ADLAvailable       bool
    LevelZeroAvailable bool
    PerCoreThermal     bool
}
```

### 2.4 Configuration options

```go
package thermal

type Option func(*config)

func WithPollIntervalMs(ms int) Option       // default 200
func WithCriticalTemperatureC(c float64) Option   // vendor-specific default
func WithSlowThresholdC(c float64) Option
```

### 2.5 The `Predictor` (optional)

```go
package thermal

// Predictor extrapolates near-future temperature based on the last
// few samples; useful for proactive scaling before throttling fires.
type Predictor struct { /* ... */ }
func NewPredictor(samples []Telemetry) *Predictor
func (p *Predictor) PredictAt(future time.Duration) float64
```

### 2.6 Scaling strategy

```go
package thermal

type Strategy int
const (
    StrategyHard   Strategy = iota   // hard cliff at SlowThreshold
    StrategySoft                     // gradual descent from SlowThreshold to Critical
    StrategyPredictive               // Predictor-based proactive descent
)
func (m *Monitor) SetStrategy(s Strategy)
```

### 2.7 Statistics

```go
package thermal

type ThermalStats struct {
    SamplesCollected uint64
    ThrottleEvents   uint64
    AverageTempC     float64
    PeakTempC        float64
    RecommendationsEmitted uint64
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`nvidia-smi -q --display=TEMPERATURE` for diagnostic backup).

### 3.2 External (Go)

- `github.com/NVIDIA/go-nvml` — NVML bindings.
- AMD ADL cgo (vendor-supplied SDK).
- Intel Level Zero cgo.

### 3.3 External (system)

- GPU driver (NVIDIA / AMD / Intel) at the host level.

---

## 4. Container Build (S02 §3 lane: `thermal-dvfs-1.x`)

**Builder:** `golang-builder-gpu`. **Runtime:** `distroless-cuda`. **Multi-arch:** `linux/amd64`. **Hardening:** standard S02 §8.2 + GPU device passthrough.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Recommendation algorithm correctness; Predictor extrapolation accuracy.

### 5.2 Integration
Real GPU + telemetry poll; verify NVML returns valid Telemetry.

### 5.3 E2E
helix-thermal → helix-pipeline → helix-encoder reconfigure.

### 5.4 Security
govulncheck + Snyk + Trivy.

### 5.5 Benchmarking
Telemetry poll p999 ≤ 10 ms; Recommend p999 ≤ 100 µs.

### 5.6 Chaos
Synthetic temperature ramp; verify scaling triggers at SlowThreshold + CriticalThreshold.

### 5.7 Stress
24-hour run with rolling temperature variation; zero leak; AverageTempC stable.

### 5.8 Smoke
30-second post-deploy: poll telemetry; verify TemperatureC > 0.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`09_thermal_throttling_dvfs/01_thermal_throttling_quality_scaling.scenario` — induce GPU thermal load, verify quality scales down before TJunction hit + restores after cooldown.

---

## 6. Challenges Entry-Point (S03 §4 row #26)

**Topology:** `09_thermal_throttling_dvfs`. **Scenario:** `01_thermal_throttling_quality_scaling.scenario.yaml`. **Why this scenario.** Thermal scaling under sustained load is the canonical thermal-management test; the scenario verifies the submodule's recommendations prevent unwanted hard throttle. **Baseline:** ScalingFactor descent matches recorded sequence; max temperature stays below TJunction by ≥ 5 °C margin.

---

## 7. R-18 Inheritance

`helix-thermal` imports `helix-r18-safeexec` for boundary subprocess invocations. Hot path is cgo to vendor SDKs.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on per-vendor SDK stability + cross-vendor scaling parity verification.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                | Purpose                                                                |
|------------------------------------|---------------|-----------------------------|------------------------------------------------------------------------|
| `HELIX_THERMAL_POLL_MS`            | `200`         | int [50, 5000]              | Telemetry poll interval.                                              |
| `HELIX_THERMAL_SLOW_C`             | `80`          | float [50, 100]             | Slow-throttle threshold.                                              |
| `HELIX_THERMAL_CRITICAL_C`         | `92`          | float [60, 110]             | Critical threshold; hard scaling cliff.                              |
| `HELIX_THERMAL_STRATEGY`           | `soft`        | `hard`/`soft`/`predictive`  | Scaling strategy.                                                     |
| `HELIX_THERMAL_DEVICE_ID`          | `0`           | int                         | GPU device index.                                                      |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Telemetry poll                        | 3 ms     | 7 ms    | 10 ms   | NVML / ADL / L0 round-trip.                                      |
| Recommend                             | 30 µs    | 70 µs   | 100 µs  | Pure-Go decision.                                                 |
| Predictor.PredictAt                   | 1 µs     | 2 µs    | 5 µs    | Linear regression on last N samples.                             |
| Memory per Monitor                    | 8 KiB    | —       | —       | Sample ring buffer + state.                                       |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `thermal: ErrNVMLInit`                        | NVML library unavailable                                       | Install NVIDIA-driver libnvidia-ml.so; verify container has access.       |
| `thermal: ErrSensorReadFailed`                | Vendor SDK returned an error reading the sensor                | Retry; if persistent, fall back to subprocess `nvidia-smi -q`.            |
| `thermal: ErrThresholdCrossed`                | Critical threshold reached but no scaling applied              | Audit Strategy; ensure consumer applies Recommendation.ScalingFactor.    |
| `thermal: ErrUnsupportedDevice`               | Vendor mismatch (e.g. configured NVIDIA, GPU is AMD)          | Auto-detect via `DetectVendors()`.                                       |

### 9.4 Migration from manual thermal monitoring

A consumer migrating from `nvidia-smi -q` polling:

1. Replace the subprocess invocation (R-18-friendly via SafeExec, but inefficient) with `mon, _ := thermal.NewMonitor(0)`.
2. Start in goroutine: `mon.Start(ctx)`.
3. Periodically read: `rec := mon.Recommend(); pipeline.ApplyScaling(rec.ScalingFactor)`.
4. Add OTLP span; metric for `ThermalStats.PeakTempC`.

The migration is documented in `docs/migration-from-nvidia-smi.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_thermal_temperature_celsius`             | gauge      | Current GPU temperature, labelled `device_id`, `vendor`.                    |
| `helix_thermal_clock_mhz`                       | gauge      | Current GPU clock.                                                           |
| `helix_thermal_throttle_events_total`           | counter    | Throttle events, labelled `reason`.                                          |
| `helix_thermal_scaling_factor`                  | gauge      | Current Recommendation.ScalingFactor.                                        |
| `helix_thermal_peak_temp_celsius`               | gauge      | All-time peak; reset on session boundary.                                    |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C34 §6](../../05_Video_Audio/09_Thermal_and_GPU_Balancing.md) — origin                    | Origin chapter; full Monitor + Telemetry + Recommendation API.                       |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline applies ScalingFactor to encoder profile.                                  |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-thermal-A        | Per-vendor thermal-zone availability — does the submodule expose unified zones?                                | C34 §6 next revision                                |
| OQ-thermal-B        | Predictive strategy default — operator opt-in or default-on?                                                   | C34 §6 next revision                                |
| OQ-thermal-C        | Mac M-series thermal API integration — required for v1.0.0?                                                   | `08_Operations/01_Container_CI_CD.md`              |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/09_Thermal_and_GPU_Balancing.md`](../../05_Video_Audio/09_Thermal_and_GPU_Balancing.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #26                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Recommend + Predictor + Strategy variants).                       |
| Integration    | Real GPU NVML telemetry poll round-trip.                                                      |
| E2E            | helix-thermal → helix-pipeline → helix-encoder reconfigure.                                   |
| Security       | govulncheck + Snyk + Trivy.                                                                   |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Synthetic temperature ramp; scaling triggers correctly.                                       |
| Stress         | 24-hour rolling variation; zero leak.                                                         |
| Smoke          | 30-second telemetry poll + temperature verification.                                          |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `09_thermal_throttling_dvfs/01_thermal_throttling_quality_scaling` baseline-parity.           |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-thermal.md` — 2026-04-30.
