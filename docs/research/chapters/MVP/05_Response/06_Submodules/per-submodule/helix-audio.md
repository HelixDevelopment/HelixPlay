# `helix-audio` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-audio`                                                                                                          |
| **Origin chapter:section**  | [C31 §6](../../05_Video_Audio/06_Audio_Pipeline.md) — *Opus MultiStream + eARC + ALLM + 5.1/7.1/Atmos*                |
| **Public path (4 mirrors)** | `vasic-digital/helix-audio` on GitHub + GitLab + GitFlic + GitVerse                                                    |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`                                                                                                |
| **External Go deps**        | `gopkg.in/hraban/opus.v2` (Opus codec), `golang.org/x/sys/unix` (ALSA / CoreAudio shims)                                |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `audio-opus-eARC-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                       |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/06_4k_120hz_hdr_dolby_vision/scenarios/05_atmos_eARC_under_full_journey.scenario.yaml`                  |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **1**                                                                                                              |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-audio` is HelixPlay's **audio capture + encode + transport primitive**. Origin: [C31 §6](../../05_Video_Audio/06_Audio_Pipeline.md). It captures the host's audio stream (PulseAudio / PipeWire on Linux, WASAPI on Windows, CoreAudio on macOS), encodes via Opus MultiStream for stereo + 5.1 + 7.1 + Atmos channel layouts, signals eARC + ALLM through the HDMI InfoFrame, and forwards encoded packets to `helix-transport`.

The submodule was introduced because audio paths drift across OS APIs and channel layouts in subtle ways (PulseAudio's channel-map vs WASAPI's vs CoreAudio's; Opus MultiStream's coupled-stream encoding requires correct mapping). Consolidation gives one canonical landing.

---

## 2. Public API Surface

### 2.1 The `Capturer` interface

```go
package audio

type Capturer interface {
    Start(layout ChannelLayout, sink Sink) error
    Stop() error
    Capabilities() Capabilities
}

func NewPulseAudioCapturer() (Capturer, error)
func NewPipeWireCapturer() (Capturer, error)
func NewWASAPICapturer() (Capturer, error)
func NewCoreAudioCapturer() (Capturer, error)
func NewBest() (Capturer, error)
```

### 2.2 The `Encoder` type

```go
package audio

// Encoder wraps Opus MultiStream for HelixPlay's channel-layout matrix.
type Encoder struct { /* ... */ }
func NewEncoder(layout ChannelLayout, bitrate int) (*Encoder, error)
func (e *Encoder) Encode(pcm []int16, sampleRate int) ([]byte, error)
func (e *Encoder) Close() error
```

### 2.3 Channel layouts

```go
package audio

type ChannelLayout int
const (
    LayoutMono     ChannelLayout = iota
    LayoutStereo
    Layout5_1                // 5.1 surround
    Layout7_1                // 7.1 surround
    LayoutAtmos              // Atmos object-based, encoded as 7.1.4 (12 channels)
)
```

### 2.4 eARC + ALLM signalling

```go
package audio

// SignalEARC writes the HDMI InfoFrame to enable eARC for compressed
// audio passthrough (Atmos, DTS:X). Requires HDMI 2.1.
func SignalEARC(monitor MonitorID) error

// SignalALLM enables Auto Low-Latency Mode (also signals through the
// InfoFrame); pairs with helix-display.EnableALLM.
func SignalALLM(monitor MonitorID) error
```

### 2.5 Capability detection

```go
package audio

func DetectCapabilities() (Capabilities, error)
type Capabilities struct {
    APIs            []API   // Pulse, PipeWire, WASAPI, CoreAudio
    OpusMultiStream bool
    AtmosSupported  bool
    DTSXSupported   bool
    eARCSupported   bool
    LowLatencyMode  bool
}
```

### 2.6 Configuration options

```go
package audio

type Option func(*config)

func WithBitrate(bps int) Option
func WithFrameDuration(ms int) Option       // 2.5/5/10/20/40/60 — Opus's options
func WithComplexity(level int) Option       // 0-10; default 8
func WithPacketLossExpected(percent int) Option
```

### 2.7 Statistics

```go
package audio

type AudioStats struct {
    BytesEncoded   uint64
    PacketsEmitted uint64
    DropsLate      uint64
    DropsBuffer    uint64
    AverageRMS     float32   // for sanity (silence detection)
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations (`pactl list` / `pw-cli` / `pacmd` for diagnostics).

### 3.2 External (Go)

- `gopkg.in/hraban/opus.v2` — Opus codec; pinned to ≥ v2.0.4.
- `golang.org/x/sys/unix` — ALSA shim.

### 3.3 External (system)

- PulseAudio / PipeWire on Linux; WASAPI on Windows; CoreAudio on macOS.
- HDMI 2.1 link for eARC + ALLM.

---

## 4. Container Build (S02 §3 lane: `audio-opus-eARC-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`. **Hardening:** standard S02 §8.2 + `--volume /run/user/1000/pulse:/run/pulse` for PulseAudio socket access.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
Channel-layout mapping correctness; Opus encode/decode round-trip.

### 5.2 Integration
Real PulseAudio + Opus encode + decode loop.

### 5.3 E2E
helix-audio → helix-transport → client decode → speaker.

### 5.4 Security
govulncheck + Snyk + Trivy. Custom: malformed-PCM fuzzer.

### 5.5 Benchmarking
Encode 7.1 @ 256 kbps p999 ≤ 5 ms / 20 ms-frame; encode latency budget ≤ 10 ms end-to-end.

### 5.6 Chaos
Inject packet loss; verify Opus FEC recovers within 1-frame budget.

### 5.7 Stress
24-hour 7.1 capture + encode; zero leak; AverageRMS within ±3 dB of baseline.

### 5.8 Smoke
30-second post-deploy: capture 1 second of silence; verify Opus packets emerge.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`06_4k_120hz_hdr_dolby_vision/05_atmos_eARC_under_full_journey.scenario` — full session with 7.1.4 Atmos audio; verify eARC pass-through to receiver, audio-video sync within 40 ms.

---

## 6. Challenges Entry-Point (S03 §4 row #23)

**Topology:** `06_4k_120hz_hdr_dolby_vision`. **Scenario:** `05_atmos_eARC_under_full_journey.scenario.yaml`. **Why this scenario.** Atmos + eARC is the most demanding audio path (object-based audio + HDMI 2.1 InfoFrame signalling); the scenario verifies end-to-end audio fidelity. **Baseline:** ViSQOL audio quality ≥ 4.5 vs original; A/V sync deviation ≤ 40 ms p999; OTLP trace topology stable.

---

## 7. R-18 Inheritance

`helix-audio` imports `helix-r18-safeexec` for boundary subprocess invocations. Hot path is cgo to OS audio API + Opus encode.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation tracks per-OS audio API stability + HDMI 2.1 ecosystem maturity.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                            | Default       | Range / type                                      | Purpose                                                                |
|------------------------------------|---------------|---------------------------------------------------|------------------------------------------------------------------------|
| `HELIX_AUDIO_API`                  | `auto`        | `pulse`/`pipewire`/`wasapi`/`coreaudio`/`auto`    | OS audio API choice.                                                   |
| `HELIX_AUDIO_LAYOUT`               | `5.1`         | `mono`/`stereo`/`5.1`/`7.1`/`atmos`               | Default channel layout.                                                |
| `HELIX_AUDIO_BITRATE_BPS`          | `256000`      | int [32000, 510000]                               | Opus target bitrate.                                                    |
| `HELIX_AUDIO_FRAME_MS`             | `10`          | int (Opus options)                                | Opus frame duration; trade-off latency vs efficiency.                  |
| `HELIX_AUDIO_eARC_ENABLED`         | `auto`        | `auto`/`true`/`false`                             | eARC compressed-passthrough.                                          |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Capture frame                         | 1 ms     | 2 ms    | 3 ms    | One PulseAudio buffer.                                            |
| Opus encode 7.1 @ 256 kbps            | 3 ms     | 4 ms    | 5 ms    | Per 20 ms frame.                                                  |
| Encode 5.1 @ 192 kbps                 | 2 ms     | 3 ms    | 4 ms    | Per 20 ms frame.                                                  |
| eARC InfoFrame                        | 50 ms    | 100 ms  | 200 ms  | One-time at session start.                                        |
| Memory per session                    | 2 MiB    | —       | —       | Opus encoder state + buffers.                                    |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `audio: ErrAPIUnavailable`                    | Audio daemon not running                                       | Start PulseAudio/PipeWire; document operator dependency.                  |
| `audio: ErrLayoutUnsupported`                 | OS API rejects the requested channel layout                   | Fall back to 5.1; metric the fallback rate.                              |
| `audio: ErreARCNotSignalled`                  | HDMI InfoFrame transmission failed                             | Verify HDMI 2.1 cable + receiver firmware.                                |
| `audio: ErrEncoderInit`                       | Opus init failed (invalid layout / sample rate combination)   | Audit channel-map; ensure sample rate is 48 kHz.                          |
| `audio: ErrUnderrun`                          | Capture rate slower than expected                              | Increase capture buffer; verify host CPU not starved.                     |

### 9.4 Migration from FFmpeg-based audio

A consumer migrating from `ffmpeg -f pulse -c:a libopus ...`:

1. Replace the FFmpeg subprocess (R-18 forbids long-running ffmpeg) with `cap, _ := audio.NewBest()` + `enc, _ := audio.NewEncoder(...)`.
2. Wire to `helix-transport` for delivery.
3. Add OTLP span around `Encode`; metric for `AudioStats.AverageRMS` (alert: silent stream detection).

The migration is documented in `docs/migration-from-ffmpeg-audio.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_audio_bytes_encoded_total`               | counter    | Bytes encoded.                                                               |
| `helix_audio_packets_emitted_total`             | counter    | Encoded packets emitted.                                                     |
| `helix_audio_drops_late_total`                  | counter    | Late drops.                                                                  |
| `helix_audio_drops_buffer_total`                | counter    | Buffer-overrun drops.                                                        |
| `helix_audio_avg_rms_db`                        | gauge      | RMS dB for silence detection.                                                |
| `helix_audio_eARC_active`                       | gauge      | 1 if eARC active, 0 otherwise.                                              |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C31 §6](../../05_Video_Audio/06_Audio_Pipeline.md) — origin                               | Origin chapter; full Capturer + Encoder + eARC API.                                  |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates audio + video sync.                                            |
| [C37 §9](../../05_Video_Audio/12_Network_Transport.md) — Transport                       | Transport carries Opus packets alongside video.                                      |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-audio-A          | DTS:X support — submodule-side or dedicated codec submodule?                                                   | C31 §6 next revision                                |
| OQ-audio-B          | Atmos object-based encoding — full 7.1.4 channel-bed or with extension data?                                  | C31 §6 next revision                                |
| OQ-audio-C          | macOS HDMI eARC reliability — Apple Silicon-specific quirks?                                                   | `08_Operations/04_Observability_and_Events.md`     |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/06_Audio_Pipeline.md`](../../05_Video_Audio/06_Audio_Pipeline.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §7  | 1,218 | 2026-04-30 | catalog row #23                                 |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (Capturer + Encoder + ChannelLayout mapping).                      |
| Integration    | Real PulseAudio + Opus encode/decode loop.                                                    |
| E2E            | Full audio → transport → client decode → speaker.                                             |
| Security       | govulncheck + Snyk + Trivy + PCM fuzzer.                                                     |
| Benchmarking   | All §9.2 budgets met.                                                                          |
| Chaos          | Packet-loss injection; Opus FEC recovers in 1 frame.                                          |
| Stress         | 24-hour 7.1 capture; zero leak.                                                               |
| Smoke          | 30-second silence capture + Opus packet verification.                                         |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `06_4k_120hz_hdr_dolby_vision/05_atmos_eARC_under_full_journey` baseline-parity.              |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-audio.md` — 2026-04-30.
