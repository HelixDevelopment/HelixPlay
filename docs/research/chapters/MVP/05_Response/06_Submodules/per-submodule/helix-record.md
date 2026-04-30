# `helix-record` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-record`                                                                                                         |
| **Origin chapter:section**  | [C30 §6](../../05_Video_Audio/05_Recording_Storage.md) — *fMP4 + MKV recording + S3/SMB/NFS sync*                     |
| **Public path (4 mirrors)** | `vasic-digital/helix-record` on GitHub + GitLab + GitFlic + GitVerse                                                   |
| **Direct deps (vasic-digital)** | `helix-r18-safeexec`, `helix-encoder`, `helix-dualpath`                                                            |
| **External Go deps**        | `github.com/Eyevinn/mp4ff` (fMP4 muxer), `github.com/yapingcat/gomedia/go-mpeg2` (TS), libavformat cgo for MKV          |
| **Licence (S01 §4.8)**      | MIT                                                                                                                     |
| **Container CI lane (S02 §3)** | `recording-mux-1.x` — builder `golang-builder-cgo`, runtime `distroless-cc`                                          |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/05_recording_emit_and_replay/scenarios/02_fmp4_mkv_emit_then_replay.scenario.yaml`                      |
| **HelixQA cadence (S04 §5)**| Per-PR + nightly + canary + pre-release                                                                                |
| **Topological depth (S01 §6.2)** | **3** (depends on helix-encoder + helix-dualpath)                                                                  |
| **R-04 duplication scan**   | Passed 2026-04-30; no collision.                                                                                        |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-record` is the **session recording + storage sync primitive**. Origin: [C30 §6](../../05_Video_Audio/05_Recording_Storage.md). It muxes the record-rung NAL output from `helix-dualpath` into fragmented MP4 (fMP4) for Web-friendly DASH delivery and into MKV for archival, then syncs the resulting files to S3-compatible / SMB / NFS storage backends per the operator's deployment.

The submodule was introduced because muxing logic spans format / storage / lifecycle management. Earlier projects re-implemented inline; consolidation gives a clean replay path.

---

## 2. Public API Surface

### 2.1 The `Recorder` type

```go
package record

// Recorder muxes incoming NAL units into a chosen container format
// and persists to disk + remote storage.
type Recorder struct { /* ... */ }

func NewRecorder(format Format, output string, opts ...Option) (*Recorder, error)
func (r *Recorder) Submit(nalUnits [][]byte, meta FrameMeta) error
func (r *Recorder) Finish() (Manifest, error)
func (r *Recorder) Close() error
```

### 2.2 Format + Manifest types

```go
package record

type Format int
const (
    FormatfMP4 Format = iota
    FormatMKV
    FormatTS
)

type Manifest struct {
    Path        string
    SizeBytes   int64
    DurationMS  int64
    SegmentURLs []string   // for fMP4 + DASH
    SHA256      string     // integrity hash
}
```

### 2.3 Storage backends

```go
package record

type Storage interface {
    Upload(localPath, remoteKey string) error
    Download(remoteKey, localPath string) error
}

func NewS3Storage(endpoint, bucket, accessKey, secretKey string) (Storage, error)
func NewSMBStorage(server, share, username, password string) (Storage, error)
func NewNFSStorage(server, exportPath string) (Storage, error)
func NewLocalFSStorage(rootPath string) (Storage, error)
```

### 2.4 The `Replay` reader

```go
package record

// Replay reads a previously-recorded session and emits the NAL
// units in original timing for replay use cases.
type Replay struct { /* ... */ }
func OpenReplay(path string, opts ...ReplayOption) (*Replay, error)
func (r *Replay) Read() (nalUnits [][]byte, meta FrameMeta, err error)
func (r *Replay) Seek(offset time.Duration) error
```

### 2.5 Configuration options

```go
package record

type Option func(*config)

func WithSegmentDuration(d time.Duration) Option
func WithStorage(s Storage) Option
func WithKeepLocal(d time.Duration) Option
func WithEncryption(key []byte) Option   // AES-GCM at rest
```

### 2.6 Statistics

```go
package record

type RecorderStats struct {
    BytesWritten   uint64
    SegmentsWritten uint64
    UploadFailures uint64
    AverageBitrate int
    StorageBackend string
}
```

### 2.7 Capability detection

```go
package record

type Capabilities struct {
    fMP4DASHCompatible bool
    MKVMuxAvailable    bool
    EncryptionAtRest   bool
}
```

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

- `helix-r18-safeexec` — for boundary subprocess invocations.
- `helix-encoder` — receives EncodedFrame.
- `helix-dualpath` — receives the record-rung output.

### 3.2 External (Go)

- `github.com/Eyevinn/mp4ff` — fMP4 muxer.
- `github.com/yapingcat/gomedia/go-mpeg2` — MPEG-TS muxer (legacy fallback).
- libavformat cgo — MKV muxer.

### 3.3 External (system)

- Storage backend reachable over network (S3 endpoint, SMB share, NFS export). For local-only operator deployments, `LocalFSStorage` is the default.

---

## 4. Container Build (S02 §3 lane: `recording-mux-1.x`)

**Builder:** `golang-builder-cgo`. **Runtime:** `distroless-cc`. **Multi-arch:** `linux/amd64` + `linux/arm64`.

---

## 5. Test Matrix (S01 §5: Ten / inline)

### 5.1 Unit
fMP4 / MKV mux byte-correctness against reference samples.

### 5.2 Integration
Real S3-compat MinIO container + Recorder upload round-trip.

### 5.3 E2E
helix-record → real fMP4 → DASH replay client + ffmpeg parse.

### 5.4 Security
govulncheck + Snyk + Trivy. AES-GCM round-trip; storage credential leakage check.

### 5.5 Benchmarking
Mux throughput ≥ 200 MB/s on amd64; segment finalisation ≤ 50 ms.

### 5.6 Chaos
S3 partition during upload; verify retry + local-buffer queue.

### 5.7 Stress
24-hour recording session; zero leak; storage upload backlog drains.

### 5.8 Smoke
30-second post-deploy: record 1-second clip; verify file exists + parses.

### 5.9 Full Automation
§5.1–§5.8.

### 5.10 Challenges
`05_recording_emit_and_replay/02_fmp4_mkv_emit_then_replay.scenario` — full session recorded → DASH replay; verify byte-exact playback parity vs original.

---

## 6. Challenges Entry-Point (S03 §4 row #22)

**Topology:** `05_recording_emit_and_replay`. **Scenario:** `02_fmp4_mkv_emit_then_replay.scenario.yaml`. **Why this scenario.** Recording is operator-revenue-critical (post-game replay, training); the scenario verifies end-to-end recording → replay parity. **Baseline:** SHA-256 integrity of recorded segments; DASH manifest valid; replay VMAF ≥ 99.0 vs original (lossless mux).

---

## 7. R-18 Inheritance

`helix-record` imports `helix-r18-safeexec` for boundary subprocess invocations (storage-backend health checks). Hot path is muxer + io.Writer; the storage upload is async via goroutine.

---

## 8. Release-Train Cadence (S01 §9)

`v0.x.y`. Graduation depends on per-storage-backend coverage; S3 + LocalFS reach `v1.0.0` first, SMB + NFS later.

---

## 9. Operational Surface

### 9.1 Configuration knobs

| Env var                              | Default                  | Range / type                        | Purpose                                                                |
|--------------------------------------|--------------------------|-------------------------------------|------------------------------------------------------------------------|
| `HELIX_RECORD_FORMAT`                | `fmp4`                   | `fmp4`/`mkv`/`ts`                   | Default container format.                                              |
| `HELIX_RECORD_SEGMENT_DURATION_S`    | `4`                      | int [1, 60]                         | DASH segment duration; shorter = lower replay latency.                 |
| `HELIX_RECORD_STORAGE_BACKEND`       | `local`                  | `local`/`s3`/`smb`/`nfs`            | Storage backend.                                                        |
| `HELIX_RECORD_KEEP_LOCAL_HOURS`      | `24`                     | int [0, 720]                        | Local copy retention.                                                  |
| `HELIX_RECORD_ENCRYPTION_KEY_FILE`   | `/run/secrets/rec-key`   | filesystem path                     | AES-GCM key for at-rest encryption (optional).                        |

### 9.2 Performance budget

| Metric                                | p50      | p99     | p999    | Notes                                                            |
|---------------------------------------|----------|---------|---------|------------------------------------------------------------------|
| Mux throughput                        | 250 MB/s | —       | —       | amd64; 4K bitrate is ~30 MB/s, plenty of headroom.              |
| Segment finalisation                  | 30 ms    | 45 ms   | 50 ms   | mp4ff finalize + SHA-256 + manifest update.                      |
| S3 upload (4 MB segment)              | 80 ms    | 200 ms  | 500 ms  | Same-region S3.                                                  |
| Replay seek                           | 50 ms    | 150 ms  | 300 ms  | fMP4 random-access.                                              |

### 9.3 Common errors and remediation

| Error                                         | Cause                                                          | Remediation                                                                |
|-----------------------------------------------|----------------------------------------------------------------|----------------------------------------------------------------------------|
| `record: ErrStorageUnreachable`               | S3 endpoint down                                                | Retry with exponential backoff; queue locally up to KEEP_LOCAL_HOURS.    |
| `record: ErrSegmentCorrupt`                   | Mux bug or disk corruption                                     | Re-record from last known IDR; alert P2.                                  |
| `record: ErrEncryptionKeyMissing`             | Env var set but key file absent                                | Fall back to unencrypted (with operator warning) or fail-closed.        |
| `record: ErrManifestInvalid`                  | DASH manifest fails validation                                 | Audit segment durations; verify mp4ff version compatibility.              |

### 9.4 Migration from FFmpeg-based recording

A consumer migrating from `ffmpeg -i ... -f segment ...`:

1. Replace the FFmpeg subprocess (R-18 forbids it) with `rec, _ := record.NewRecorder(record.FormatfMP4, ...)`.
2. Wire the record-rung NAL feed via `helix-dualpath`.
3. Configure storage: `record.WithStorage(record.NewS3Storage(...))`.
4. Add OTLP span around `Submit`; metric for `RecorderStats.UploadFailures`.

The migration is documented in `docs/migration-from-ffmpeg-recording.md`.

### 9.5 Observability metrics catalog

| Metric                                          | Type       | Description                                                                  |
|-------------------------------------------------|------------|------------------------------------------------------------------------------|
| `helix_record_bytes_written_total`              | counter    | Bytes muxed.                                                                 |
| `helix_record_segments_written_total`           | counter    | Segments finalised.                                                          |
| `helix_record_upload_failures_total`            | counter    | Storage upload failures, labelled `backend`, `error_kind`.                  |
| `helix_record_average_bitrate_bps`              | gauge      | Observed average bitrate.                                                    |
| `helix_record_local_buffer_bytes`               | gauge      | Local-disk upload-pending bytes; alert if > 1 GiB.                          |

### 9.6 Consumer matrix

| Consumer chapter:section                                                                  | Integration purpose                                                                  |
|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [C30 §6](../../05_Video_Audio/05_Recording_Storage.md) — origin                            | Origin chapter; full Recorder + Replay + Storage API.                                |
| [C36 §8](../../05_Video_Audio/11_Go_Pipeline_Implementation.md) — Pipeline                | Pipeline orchestrates record-rung → recorder lifecycle.                              |

---

## 10. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-record-A         | DASH-CMAF vs HLS — does the submodule support both?                                                            | C30 §6 next revision                                |
| OQ-record-B         | At-rest encryption KMS — bring-your-own-key vs vault-managed?                                                  | C30 §6 + C10 §6 next revision                       |
| OQ-record-C         | Replay scrubbing UX — submodule-side index or client-side?                                                     | `08_Operations/04_Observability_and_Events.md`     |

---

## 11. Anti-Bluff Verification

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../05_Video_Audio/05_Recording_Storage.md`](../../05_Video_Audio/05_Recording_Storage.md) §6 | (slice) | 2026-04-30 | origin chapter                          |
| [`helix-encoder.md`](helix-encoder.md), [`helix-dualpath.md`](helix-dualpath.md) | (this batch) | 2026-04-30 | direct deps              |

Forbidden patterns: clean. Three §10 open questions named with deferred resolution chapters.

### Per-test-type coverage targets (R-12)

| Test type      | Coverage target                                                                                |
|----------------|-----------------------------------------------------------------------------------------------|
| Unit           | ≥ 95 % statement coverage (mux + replay + storage abstraction).                              |
| Integration    | Real S3-compat MinIO + recorder upload round-trip.                                            |
| E2E            | helix-record → fMP4 → DASH replay parse with ffmpeg.                                          |
| Security       | govulncheck + Snyk + Trivy + AES-GCM round-trip + credential-leak check.                     |
| Benchmarking   | Mux ≥ 200 MB/s; finalisation ≤ 50 ms p999.                                                    |
| Chaos          | S3 partition; retry + local buffering correctness.                                            |
| Stress         | 24-hour recording; zero leak; backlog drains.                                                 |
| Smoke          | 30-second 1-second clip + parse verification.                                                 |
| Full Automation| §5.1–§5.8 in CI matrix on every PR; no row may be skipped.                                   |
| Challenges     | `05_recording_emit_and_replay/02_fmp4_mkv_emit_then_replay` baseline-parity.                  |

Sign-off: drafted by orchestrator (Claude Opus 4.7) on 2026-04-30. Pending operator review.

End of `06_Submodules/per-submodule/helix-record.md` — 2026-04-30.
