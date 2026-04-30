# `helix-dualpath` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-dualpath`                                                                                                       |
| **Origin chapter:section**  | [C29 §6](../../05_Video_Audio/04_DualPath_Encoding.md) — *Dual-rung NAL feed (stream + record)*                       |
| **Public path (4 mirrors)** | `vasic-digital/helix-dualpath` on GitHub + GitLab + GitFlic + GitVerse                                                 |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-encoder`                                                                              |
| **External Go deps**        | None at runtime; cgo bindings to libavformat for container muxing validation                                            |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `dual-path-nal-1.x` — builder `golang-builder-gpu`, runtime `distroless-cuda`                                         |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/05_recording_emit_and_replay/scenarios/01_dual_path_nal_stream_plus_record.scenario.yaml`                |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **2** (depends on helix-encoder)                                                                                   |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-dualpath` performs the **dual-rung NAL feed split** for HelixPlay's stream + record pipelines. Origin: [C29 §6](../../05_Video_Audio/04_DualPath_Encoding.md). The submodule consumes a single `helix-encoder.EncodedFrame` and emits two parallel feeds: a low-latency stream feed (P-frames every 16 ms, IDR every 60 frames) and a recording-quality feed (more I-frames for seek-friendliness, longer GOP for compression).

The design avoids encoding twice (which would double the GPU cost). Instead, the encoder emits a single high-quality stream, and `helix-dualpath` re-orders + filters NAL units into two streams suitable for their respective downstream consumers (`helix-transport` for streaming, `helix-record` for recording).

---

## 2. Public API Surface

### 2.1 The `Splitter` type

```go
package dualpath

// Splitter is the per-session NAL re-orderer. The hot path is
// allocation-free; output NAL slices reference the input bytes.
type Splitter struct { /* ... */ }

func NewSplitter(streamSink, recordSink Sink, opts ...Option) *Splitter
func (s *Splitter) Submit(frame *encoder.EncodedFrame) error
func (s *Splitter) Close() error
```

### 2.2 The `Sink` interface

```go
package dualpath

type Sink interface {
    Write(rung Rung, nalUnits [][]byte, meta FrameMeta) error
}

type Rung int
const (
    RungStream Rung = iota   // low-latency feed
    RungRecord                // recording-quality feed
)
```

### 2.3 Configuration options

```go
package dualpath

type Option func(*config)

func WithStreamGOP(frames int) Option
func WithRecordGOP(frames int) Option
func WithStreamMaxSliceSize(bytes int) Option
func WithRecordKeyframeOnDemand() Option   // emit IDR on Sink.Write request
```

### 2.4 Statistics

```go
package dualpath

type SplitterStats struct {
    FramesSubmitted uint64
    StreamRungCount uint64
    RecordRungCount uint64
    DropsBackpressure uint64
    KeyframesEmitted uint64
}
```

### 2.5 Capability detection

```go
package dualpath

type Capabilities struct {
    SliceFiltering    bool   // ability to filter NAL units by slice type
    KeyframeOnDemand  bool   // record-side IDR injection
}
```

### 2.6 NAL-unit slice-type constants (per H.264/HEVC)

```go
package dualpath

// SliceType per the H.264 / HEVC specifications. The classification
// drives rung routing.
type SliceType int
const (
    SliceP   SliceType = iota   // P-slice; stream-rung primary
    SliceB                       // B-slice; record-rung
    SliceI                       // I-slice; both rungs
    SliceSI                      // SI-slice; both rungs
    SliceSP                      // SP-slice; rare
    SliceIDR                     // IDR; both rungs (resync)
    SliceSEI                     // Supplemental Enhancement Information
    SlicePPS                     // Picture Parameter Set
    SliceSPS                     // Sequence Parameter Set
    SliceVPS                     // Video Parameter Set (HEVC only)
)
```

### 2.7 Per-rung filter rules

```go
package dualpath

// FilterRule describes how the splitter routes a slice type to
// each rung. Defaults are documented in C29 §6 + C30 §6.
type FilterRule struct {
    SliceType  SliceType
    StreamRung bool
    RecordRung bool
}

// DefaultStreamFilter routes IDR + P + SEI + PPS/SPS/VPS to the
// stream rung; B-slices and SI/SP go only to record.
func DefaultStreamFilter() []FilterRule

// DefaultRecordFilter routes everything to the record rung; the
// record path archives the full coded video sequence.
func DefaultRecordFilter() []FilterRule
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations.
- `helix-encoder` — consumes EncodedFrame.

### 3.2 External (Go)

- libavformat cgo (validation only).

### 3.3 External (system)

None.

---

## 4. Container Build (S02 §3 lane: `dual-path-nal-1.x`)

**Builder:** `golang-builder-gpu` (depends on encoder lane image). **Runtime:** `distroless-cuda`. **Multi-arch:** `linux/amd64` + `linux/arm64`.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
NAL-unit re-ordering correctness; rung-classification logic.

### 5.2 Integration
helix-encoder real output + Splitter; verify NAL units appear on the correct rung.

### 5.3 E2E
helix-dualpath → helix-transport (stream rung) + helix-record (record rung).

### 5.4 Security
govulncheck + Snyk + Trivy; malformed-NAL fuzzer.

### 5.5 Benchmarking
Splitter.Submit p999 ≤ 5 µs; allocation-free hot path verified.

### 5.6 Chaos
Backpressure on one sink while the other drains; verify clean drop categorisation.

### 5.7 Stress
24-hour 4K120 dual-rung; zero leak; rung counts balanced.

### 5.8 Smoke
30-second post-deploy: submit one EncodedFrame; verify both sinks receive their rung.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`05_recording_emit_and_replay/01_dual_path_nal_stream_plus_record.scenario` — full session with both rungs active; verify recording is seekable + stream maintains low latency.

---

## 6. Challenges Entry-Point (S03 §4 row #21)

**Topology:** `05_recording_emit_and_replay`. **Scenario:** `01_dual_path_nal_stream_plus_record.scenario.yaml`. **Why this scenario.** The dual-rung split's value prop is realised when both feeds are live; the scenario verifies neither feed degrades the other. **Baseline:** stream rung p999 latency unchanged from single-rung baseline; record rung produces seek-friendly fMP4 with IDR every 30 s.

---

## 7. R-18 Inheritance

`helix-dualpath` imports `helix-r18-safeexec` for boundary subprocess invocations. Hot path is pure-Go NAL re-ordering.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on the C29 §6 API freeze + cross-rung parity verification.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type            | Purpose                                                                |
|------------------------------------|---------------|-------------------------|------------------------------------------------------------------------|
| `HELIX_DUALPATH_STREAM_GOP`        | `60`          | int [1, 240]            | Stream-rung GOP size.                                                  |
| `HELIX_DUALPATH_RECORD_GOP`        | `120`         | int [60, 600]           | Record-rung GOP size.                                                  |
| `HELIX_DUALPATH_RECORD_KF_ON_DEMAND` | `true`      | bool                    | Allow record-rung IDR injection on Sink request.                      |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Splitter.Submit                       | 1 µs     | 3 µs    | 5 µs    | Hot path; allocation-free.                                       |
| NAL-unit classification               | 100 ns   | 300 ns  | 500 ns  | Single-byte slice-type read.                                     |
| Per-frame memory                      | 0 bytes  | —       | —       | Slices reference input bytes.                                    |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `dualpath: ErrSinkBackpressure`               | One sink fell behind                                           | Investigate slow sink; the splitter drops to maintain the other rung.    |
| `dualpath: ErrInvalidNAL`                     | Encoder emitted unrecognisable NAL                              | Audit encoder version; report upstream if vendor-bug.                     |
| `dualpath: ErrKeyframeOnDemandUnsupported`    | Encoder doesn't expose keyframe injection                       | Disable WithRecordKeyframeOnDemand; rely on natural IDR cadence.         |

### 9.4 Migration from sequential stream-then-record

A consumer migrating from re-encoding the stream for recording:

1. Replace the second encoder with `dualpath.NewSplitter(streamSink, recordSink)`.
2. Update the encoder GOP to be record-friendly (longer); the stream-rung filter handles low-latency.
3. Remove the second GPU encoder allocation; recoup the GPU budget.
4. Add OTLP span around `Splitter.Submit`; metric for `SplitterStats.RungCount` ratio.

The migration is documented in `docs/migration-from-double-encode.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_dualpath_submitted_total`                | counter    | Frames submitted to splitter.                                                |
| `helix_dualpath_rung_writes_total`              | counter    | Per-rung writes, labelled `rung={stream, record}`.                           |
| `helix_dualpath_drops_backpressure_total`       | counter    | Drops due to sink backpressure.                                              |
| `helix_dualpath_keyframes_emitted_total`        | counter    | Keyframes emitted, labelled `rung`.                                          |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C29 §6](../../05_Video_Audio/04_DualPath_Encoding.md) — origin                            | Origin chapter; full Splitter / Sink API.                                            |
| [C30 §6](../../05_Video_Audio/05_Recording_Storage.md) — Recording                        | helix-record consumes the record-rung output.                                        |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates split between transport + record.                              |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport consumes the stream-rung output.                                           |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-dualpath-A       | Per-rung bitrate independence — possible without re-encoding?                                                  | C29 §6 next revision                                |
| OQ-dualpath-B       | AV1 OBU re-ordering — semantics differ from H.264/HEVC; needs separate code path?                              | C29 §6 next revision                                |
| OQ-dualpath-C       | Keyframe-on-demand for live-replay — does this submodule own the API?                                          | `08_Operations/04_Observability_and_Events.md`     |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/04_DualPath_Encoding.md`](../../05_Video_Audio/04_DualPath_Encoding.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-encoder.md`](helix-encoder.md) | (this batch) | 2026-04-30 | direct dependency                              |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (NAL classification + re-ordering).                                |
| Integration    | helix-encoder real output + Splitter.                                                         |
| E2E            | dualpath → transport (stream) + record (record) round-trip.                                   |
| Security       | govulncheck + Snyk + Trivy + malformed-NAL fuzzer.                                            |
| Benchmarking   | Submit ≤ 5 µs p999; zero allocations.                                                         |
| Chaos          | Per-sink backpressure; drop categorisation correct.                                           |
| Stress         | 24-hour 4K120 dual-rung; zero leak.                                                           |
| Smoke          | 30-second 1-frame split + both-sink verification.                                             |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `05_recording_emit_and_replay/01_dual_path_nal_stream_plus_record` baseline-parity.           |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-dualpath.md` — 2026-04-30.
