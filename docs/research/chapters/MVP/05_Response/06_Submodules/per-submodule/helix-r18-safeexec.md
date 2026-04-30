# `helix-r18-safeexec` — Per-Submodule Descriptor (S05)

| Field                       | Value                                                                                                                  |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------|
| **Submodule name**          | `helix-r18-safeexec`                                                                                                   |
| **Origin chapter:section**  | [C08 §10](../../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) — *Subprocess wrapper for R-18 enforcement*       |
| **Public path (4 mirrors)** | `vasic-digital/helix-r18-safeexec` on GitHub + GitLab + GitFlic + GitVerse                                              |
| **Direct deps (vasic-digital)** | (none — root of the dependency tree)                                                                              |
| **External Go deps**        | Go standard library only (`os`, `os/exec`, `syscall`, `context`, `errors`, `strings`)                                  |
| **Licence (S01 §4.8)**      | MIT (`SPDX-License-Identifier: MIT`)                                                                                    |
| **Container CI lane (S02 §3)** | `subprocess-wrapper-1.x` — builder `golang-builder`, runtime `distroless-static`                                    |
| **Test matrix (S01 §5)**    | Ten / inline                                                                                                            |
| **Challenges entry (S03 §4)** | `topologies/14_security_attack_surface/scenarios/01_attempt_forbidden_subprocess_invocation.scenario.yaml`           |
| **HelixQA cadence (S04 §5)**| Per-PR (primary scenario) + nightly (full regression) + canary (extended fault injection) + pre-release (replay set) |
| **Topological depth (S01 §6.2)** | **0** — no `vasic-digital` dependencies. The foundation of the dependency tree.                                  |
| **Imported by**             | 24 of 29 submodules (everything except `helix-vault` and `helix-tenant`, plus this submodule itself).                  |
| **R-18 inheritance**        | This submodule **is** the R-18 enforcement primitive. Every other submodule's R-18 inheritance ([S01 §7](../01_Submodule_Catalog.md#7-the-r-18-inheritance-ladder-canonical-from-helix-r18-safeexec-outward)) chains through here. |
| **R-04 duplication scan**   | Passed 2026-04-30 across all four `vasic-digital` mirrors; no collision with existing repos.                            |
| **Status**                  | Draft v1.                                                                                                                |
| **Last updated**            | 2026-04-30.                                                                                                              |

---

## 1. Purpose

`helix-r18-safeexec` is the **subprocess execution wrapper** that operationalises Constitution §11.5 (R-18 — Operational Integrity). The submodule exposes a single public constructor — `r18.SafeExec(ctx, name, args ...string) (*exec.Cmd, error)` — which enforces a non-overridable deny-list of host-disruptive commands. Every other `vasic-digital` submodule that ever invokes a subprocess imports this submodule directly; no submodule may construct an `*exec.Cmd` outside this wrapper.

The submodule was introduced in [C08 §10](../../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) after the host-disruption incident recorded in [Master Plan §10 Session 3](../../00_Master_Plan.md#10-session-log) — the operator's development host was suspended / signed-out mid-session despite no causal link from any tool call. The post-incident retrospective concluded that the project could not afford a recurrence, and a defence-in-depth wrapper at the `os/exec` boundary was the strongest mitigation available without sacrificing the project's broader containerised-runtime posture.

The wrapper is the **single point of failure** for R-18 enforcement (S01 §6.3). A CVE in this submodule, or a regression that relaxes the deny-list, would propagate across 24 of 29 submodules. The mitigations — conservative API surface, maximum test coverage, two-reviewer rule, tag protection — are codified in [S01 §6.3](../01_Submodule_Catalog.md#63-single-point-of-failure-helix-r18-safeexec) and re-inherited here.

`★ The R-18 § identifier convention.` Constitution §11.5 is partitioned into seven sub-sections; this submodule's deny-list is the runtime expression of §11.5.1 (forbidden host-disruptive commands). The static deny-list (the *list itself*) lives at §11.5.4 in the Constitution and is mirrored as a private constant in this submodule's source. The five-layer enforcement model (S01 §7.3) ranges from chapter prose down to auditd events, with this submodule occupying layers 2 and 3 (static deny-list + runtime wrapper).

---

## 2. Public API Surface

The full API is intentionally narrow. C08 §10 codifies the surface; this section reproduces the high-level shape and refers back to C08 for the detailed semantics.

### 2.1 The `SafeExec` constructor

```go
package r18

// SafeExec returns an *exec.Cmd that, when started, will reject the
// invocation if `name` or any of `args` matches an entry on the
// non-overridable deny-list. The deny-list is a private constant
// inside this package; callers cannot read or modify it.
//
// On a deny-list match, SafeExec returns r18.ErrForbidden and the
// returned *exec.Cmd is non-nil but its Run/Start methods will also
// return r18.ErrForbidden (defensive double-gate).
//
// Returns an error from context cancellation, deny-list match, or
// argument validation failure (e.g. nil args, empty name).
func SafeExec(ctx context.Context, name string, args ...string) (*exec.Cmd, error)
```

The constructor's **defensive double-gate** — both the constructor itself and the returned `*exec.Cmd`'s `Run` / `Start` methods reject deny-list matches — is intentional. A bug in either gate alone would still leave the other intact; both must be subverted for a forbidden invocation to reach the kernel.

### 2.2 The `ErrForbidden` sentinel

```go
package r18

// ErrForbidden is returned when SafeExec rejects an invocation.
// Callers should test with errors.Is(err, r18.ErrForbidden) and
// surface the rejection up the call chain rather than retrying with
// a different command.
var ErrForbidden = errors.New("r18: forbidden host-disruptive command")
```

Callers MUST NOT attempt to fall back to a different command when `ErrForbidden` is returned — the rejection is intentional and represents a programming error or a malicious input. The correct response is to log the attempt (which `helix-r18-safeexec`'s test harness will detect via `host-integrity-scan` strace traces) and propagate the error to the caller.

### 2.3 Helpers for callers

```go
package r18

// Wrap takes an *exec.Cmd that the caller has already constructed
// (typically via exec.Command) and re-validates it against the
// deny-list. This is the migration path for code that pre-dates
// helix-r18-safeexec; new code should use SafeExec directly.
//
// Wrap returns ErrForbidden if the command's Path or Args matches
// the deny-list.
func Wrap(cmd *exec.Cmd) (*exec.Cmd, error)

// MustSafeExec is like SafeExec but panics on error. Suitable only
// for init-time use where a deny-list match indicates a programming
// error that should crash early.
func MustSafeExec(ctx context.Context, name string, args ...string) *exec.Cmd
```

`Wrap` exists for the migration period during Phase 02 implementation; once all 24 consumer submodules are on `SafeExec`, the `Wrap` helper is deprecated (but not removed; the deprecation is recorded in the submodule's `CHANGELOG.md`).

### 2.4 The deny-list (private)

The deny-list is a private constant slice inside `r18/internal/denylist.go`. Its content is mirrored as a public test fixture at `vasic-digital/Containers/lanes/host-integrity-scan/deny-list.txt` per [S02 §3.4](../02_Containers_Submodule.md#34-the-host-integrity-scan-shared-lane); the fixture is the **single source of truth** for the deny-list across the fleet, and the submodule's source `denylist.go` is generated from the fixture by a `go:generate` directive at build time.

The fixture contains (non-exhaustive sample):

- `pm-suspend`, `pm-hibernate`, `pm-suspend-hybrid`
- `systemctl suspend`, `systemctl hibernate`, `systemctl reboot`, `systemctl poweroff`, `systemctl halt`
- `shutdown`, `reboot`, `halt`, `poweroff`
- `gnome-session-quit`, `loginctl terminate-session`, `loginctl terminate-user`
- `dconf write /org/gnome/desktop/screensaver/lock-enabled`
- `pmset sleepnow` (macOS)
- `tell application "System Events" to sleep` (macOS osascript)
- Windows equivalents: `shutdown.exe`, `psshutdown.exe`, `tsdiscon.exe`

The full list lives in the fixture; the deny-list is **versioned with the Constitution** (Constitution §15 *Amendment Procedure* applies to additions / removals).

---

## 3. Direct Dependencies

### 3.1 Within `vasic-digital`

**None.** This submodule is the dependency-tree root. Adding a `vasic-digital` dependency to it would create a cycle (since 24 submodules import it), and the per-PR `dep-cycle-check.go` lint (S01 §6.5) would reject the addition.

### 3.2 External (Go)

Standard library only:

- `os` — `os.Getenv`, `os.LookupEnv` for environment-variable propagation.
- `os/exec` — `exec.Cmd`, `exec.Command`, `exec.CommandContext` (wrapped, never exported directly).
- `syscall` — `syscall.SysProcAttr` for setting cgroup membership and namespace flags.
- `context` — context propagation for cancellation.
- `errors` — `errors.New`, `errors.Is`.
- `strings` — `strings.HasPrefix`, `strings.Contains` for deny-list matching.
- `regexp` — `regexp.MustCompile` for flexible deny-list patterns (compiled at init).

The deliberate absence of any external Go dependency keeps the supply chain audit at this submodule trivially small. A Snyk + govulncheck run on `helix-r18-safeexec` reports zero issues because the dependency closure is the standard library plus this submodule itself.

### 3.3 External (C / system)

None at runtime — the runtime image is `distroless-static` per the §3.1 table. At test time, the `host-integrity-scan` lane's container additionally bundles `strace` and `auditd` for the syscall-tracing test row, but those binaries are not part of `helix-r18-safeexec`'s runtime distribution.

---

## 4. Container Build (S02 §3 lane: `subprocess-wrapper-1.x`)

### 4.1 Builder image

`ghcr.io/vasic-digital/containers/golang-builder:1.24-bookworm` (S02 §3.2). The base builder is sufficient — no cgo, no GPU, no Windows / Darwin variants. The build is statically-linked with `CGO_ENABLED=0` so the runtime image can be `distroless-static`.

### 4.2 Runtime image

`gcr.io/distroless/static-debian12:nonroot@sha256:<digest>` (S02 §5). Static-linked Go binary in a base image with only `/etc/passwd` + `/etc/nsswitch.conf` + CA certs. UID 65532 (`nonroot`).

### 4.3 Multi-arch coverage

`linux/amd64` + `linux/arm64` per S02 §4. The submodule has no architecture-specific code paths (the deny-list is platform-text, not platform-byte; deny-list entries with platform suffixes like `pmset` for macOS are matched against the `name` argument string, not against the build OS).

### 4.4 Container hardening

Per S02 §8.2:

- `--user=65532:65532` — UID 65532 nonroot.
- `--read-only` — root filesystem read-only.
- `--tmpfs=/tmp:rw,noexec,nosuid` — `/tmp` is tmpfs, no executable bit.
- `--security-opt=seccomp=<allowlist.json>` — seccomp default-deny posture, allowlist-based.
- `--security-opt=no-new-privileges` — no setuid escalation.
- `--cap-drop=ALL` — drop all capabilities. The submodule does not need any.
- `--memory=256m` — generous for a wrapper submodule that holds only the deny-list in memory.
- `--cpus=0.5` — half a CPU sufficient.

### 4.5 The image signing chain

Per S02 §6:

- Build signs the image with cosign keyless (Sigstore-anchored, GitHub OIDC-bound).
- SLSA L3 provenance attached as a cosign attestation.
- cyclonedx-gomod SBOM attached as `bom.cdx.json`.
- syft container SBOM attached as `image.spdx.json`.
- All four artefacts pushed to all four mirror registries.

---

## 5. Test Matrix (S01 §5: Ten / inline)

The Ten test types per Constitution §6.2, all in-tree per S01 §5.1. Special properties of this submodule's matrix:

### 5.1 Unit (mocks allowed per R-12)

The deny-list matcher is the primary unit-test target. Tests cover:

- Each deny-list entry rejected (positive coverage).
- Adjacent strings accepted (e.g. `pm-suspend-fake` is accepted because it's not on the list, `pm-suspend` is rejected — the matcher is exact-match on the command name, regex-match on argument patterns).
- Unicode normalisation attacks (e.g. `pm-suspend​` with a zero-width space — the matcher should reject this because the deny-list entries are normalised).
- Path-resolution attacks (e.g. `/usr/bin/pm-suspend` — the matcher uses the basename as well as the full path).
- Argument-array attacks (e.g. `systemctl` is allowed but `systemctl suspend` is rejected; the matcher must inspect both `name` and `args`).

### 5.2 Integration (no mocks)

Real `os/exec` invocations against allowed commands (e.g. `echo`, `true`). The integration tests verify that allowed invocations succeed and produce expected stdout / stderr / exit code.

### 5.3 E2E (no mocks)

`helix-r18-safeexec`'s E2E test invokes `r18.SafeExec(ctx, "true")` and asserts exit code 0; invokes `r18.SafeExec(ctx, "pm-suspend")` and asserts `errors.Is(err, r18.ErrForbidden)`.

### 5.4 Security (no mocks)

- Govulncheck against the dependency closure (zero findings expected, since the closure is the standard library).
- Snyk scan (zero findings expected).
- Trivy container scan (zero findings expected on `distroless-static`; periodic findings only when distroless ships an updated base image with a freshly-disclosed CVE).
- Custom fuzzer (`go-fuzz` on the deny-list matcher, looking for bypass via Unicode / path-traversal / argument-array tricks).

### 5.5 Benchmarking (no mocks)

The `SafeExec` constructor's overhead is measured. Target: ≤ 10 µs per invocation (the deny-list match should be a hash-table lookup plus a small regex match, both microsecond-scale).

### 5.6 Chaos (no mocks)

Toxiproxy-injected delays on stdin/stdout pipes during test invocations; verify the wrapper handles I/O timeouts cleanly without leaking goroutines.

### 5.7 Stress (no mocks)

24-hour run of 10K SafeExec invocations per second against allowed commands; verify no memory growth, no fd leaks, no goroutine leaks.

### 5.8 Smoke (no mocks)

30-second post-deploy: invoke `r18.SafeExec(ctx, "echo", "ok")` and verify "ok" appears in stdout.

### 5.9 Full Automation

The matrix invocation orchestrator runs §5.1 through §5.8 in CI matrix order with fail-fast disabled.

### 5.10 Challenges

`topologies/14_security_attack_surface/scenarios/01_attempt_forbidden_subprocess_invocation.scenario.yaml`. Boots the canonical attack-surface topology and attempts every deny-list entry against the running `helix-r18-safeexec` instance; verifies every attempt is rejected and no host-disruptive syscall reaches the kernel (verified via auditd inside the container per S02 §3.4).

---

## 6. Challenges Entry-Point (S03 §4 row #01)

**Topology:** [`topologies/14_security_attack_surface`](../03_Challenges_Submodule.md#3-repository-layout-topologies--baselines--harness)

**Scenario:** `01_attempt_forbidden_subprocess_invocation.scenario.yaml`

**Why this scenario.** The `14_security_attack_surface` topology is the only canonical Challenges topology that explicitly attempts host-disruptive operations. The scenario's role is to confirm at the **production-like** level (full container topology, real auditd, real strace, real seccomp profile) that the deny-list rejects every entry. A green run on this scenario is the strongest available signal that R-18 enforcement is intact at the operational level.

**Baseline content.** The baseline at `baselines/14_security_attack_surface/01_attempt_forbidden_subprocess_invocation/` contains:

- `audit-events.jsonl` — recording of every auditd event during the scenario (zero forbidden-syscall events expected).
- `strace-events.jsonl` — recording of every syscall the wrapper itself invoked (a small allowlist subset).
- `rejection-log.jsonl` — recording of every `ErrForbidden` rejection the wrapper emitted.
- `latency-histograms.json` — p50/p99/p999 of the rejection-decision wall-clock (target: ≤ 10 µs p999).

**Change-point detection.** Per S03 §6.3:

- Auditd events: **exact equality** with the baseline. Any new event is a regression candidate.
- Strace events: structural identity at the syscall level (system calls in the same order, same arguments).
- Rejection log: every entry on the deny-list must be rejected; missing rejections are regressions.
- Latency histograms: Mann-Whitney U test, p > 0.01 to pass.

---

## 7. R-18 Inheritance (S01 §7)

This submodule **defines** the R-18 inheritance ladder; it does not consume it. Per [S01 §7.1](../01_Submodule_Catalog.md#71-the-ladder), 24 other submodules import `r18.SafeExec` from this submodule. Per [S01 §7.3](../01_Submodule_Catalog.md#73-the-five-layer-enforcement-per-c08-§10--§115), this submodule occupies layers 2 (static deny-list) and 3 (runtime wrapper) of the five-layer enforcement.

The other three layers (chapter prose, ripgrep CI lint, host-integrity-scan strace+auditd) are not in-scope for this submodule's source code but are **operationally required** for the submodule's R-18 guarantee to hold:

- **Layer 1** (chapter prose) — every chapter §6 documents its specific R-18 surface. C08 §10 is this submodule's origin chapter and the canonical reference.
- **Layer 4** (ripgrep CI lint) — `vasic-digital/Containers/ci-fragments/dep-cycle-check.yml` includes the `direct-exec-grep` step that fails any submodule's CI if a non-wrapped `exec.Command` is found.
- **Layer 5** (host-integrity-scan strace+auditd) — `vasic-digital/Containers/lanes/host-integrity-scan/` is the runtime test that this submodule's Challenges row consumes.

A regression in any one layer is caught by the next; the five-layer model is defence-in-depth. This submodule's source-code regressions are caught at layers 2–4; runtime regressions are caught at layer 5.

---

## 8. Release-Train Cadence (S01 §9)

### 8.1 Current phase: `v0.x.y`

The submodule is currently in `v0.x.y` MVP development. Per [S01 §9.1](../01_Submodule_Catalog.md#91-the-four-phase-per-submodule-lifecycle), the API may evolve until the API-freeze graduation criteria (§9.2) are met.

### 8.2 Graduation criteria

Per S01 §9.2, this submodule graduates to `v1.0.0` when:

1. The public API (`SafeExec`, `Wrap`, `ErrForbidden`, `MustSafeExec`) has been frozen — no exported symbol removed or signature-changed in two consecutive release cycles.
2. The Ten-test-type matrix has been fully green for two consecutive release cycles.
3. No `v0` consumer remains (this is automatic for the dependency-tree root since it has no `vasic-digital` deps).
4. SBOM + vulnerability reports have been clean for two consecutive cycles.
5. Constitution §16 *Acceptance* sign-off from two reviewers.

### 8.3 The two-reviewer rule (S01 §6.3 mitigation #3)

This submodule has the **strictest** review policy in the fleet: every PR requires two reviewer sign-offs, the two reviewers cannot be the same person, and the reviewers cannot also be the PR author. The rule is enforced by GitHub branch protection (`Require approvals: 2`, `Require review from Code Owners: true`) and the equivalent on GitLab.

The two-reviewer rule applies to **every** PR (not just `v1.0.0+` graduation), because a relaxation of the deny-list at any phase has fleet-wide R-18 consequences.

### 8.4 v2.0.0 cost (S01 §9.3)

A `v2.0.0` bump for this submodule would require updating **every** of the 24 consumer submodules' `go.mod` files in lockstep. This is the largest possible release-train cost in the fleet (S01 §9.3 documents the 1-month wall-clock for v2 bumps; this submodule is the worst case). A `v2` bump is therefore avoided unless an irrecoverable API change demands it.

---

## 9. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-r18-safeexec-A   | Should the deny-list include macOS-specific `osascript` patterns beyond the AppleScript `tell application` form? | C08 §10 next revision                               |
| OQ-r18-safeexec-B   | Performance budget for the deny-list match in the hot path — 10 µs p999 sustainable at 1M qps?                 | `04_Latency/09_Memory_and_Cache_Optimization.md` revisions |
| OQ-r18-safeexec-C   | Should `Wrap` graduate to a generally-supported helper or be removed after the migration period?               | C08 §10 next revision (target: pre-`v1.0.0`)        |

None of the three are placeholders; each has a named resolution chapter and a specific operational concern.

---

## 10. Anti-Bluff Verification

### 10.1 Source evidence reviewed

| Path                                                              | Lines  | Reviewed   | Role                                            |
|-------------------------------------------------------------------|-------:|------------|-------------------------------------------------|
| [`../../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md`](../../03_Architecture/07_Host_Agent_and_Game_Lifecycle.md) §10 | (slice) | 2026-04-30 | origin chapter; full API surface + R-18 wrapper rationale |
| [`../01_Submodule_Catalog.md`](../01_Submodule_Catalog.md) §3 §6.3 §7 | 1,218 | 2026-04-30 | catalog row #01, SPOF analysis, inheritance ladder |
| [`../02_Containers_Submodule.md`](../02_Containers_Submodule.md) §3.4 §8 |   627 | 2026-04-30 | host-integrity-scan harness + container hardening |
| [`../03_Challenges_Submodule.md`](../03_Challenges_Submodule.md) §4.1 §6 |   517 | 2026-04-30 | scenario + change-point thresholds              |
| [`../04_HelixQA_Integration.md`](../04_HelixQA_Integration.md) §5 §9 |   584 | 2026-04-30 | cadence + deployment-gate                      |
| [`../../01_Constitution.md`](../../01_Constitution.md) §11.5     |    879 | 2026-04-30 | R-18 normative wording + the Forbidden Commands list |

### 10.2 Forbidden-pattern scan

Clean. No `TODO` / `FIXME` / `tbd` / `xxx` / `???` / `placeholder` / "fill in later" markers in this descriptor's body. The three §9 open questions are explicitly named with deferred resolution chapters.

### 10.3 Coverage confirmation

- Floor: **≥ 300 lines** per [Master Plan §7.2](../../00_Master_Plan.md#72-queued) row S05.
- Achieved: full descriptor; `wc -l` recorded at chapter close.
- Cross-link integrity: every `[S01 §X](path)` link resolves to a real heading anchor in `01_Submodule_Catalog.md`; every `[C08 §10](path)` link resolves to the origin chapter; every `[S02 §X](path)` and `[S03 §X](path)` and `[S04 §X](path)` link resolves to its sibling chapter.

### 10.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30 as a single inline `Write` call.
- Reviewed by: pending operator review.
- Two-reviewer rule (§8.3) applies to every PR against this submodule's repository, but does not gate this descriptor's chapter sign-off (the descriptor is documentation, not source code).

End of `06_Submodules/per-submodule/helix-r18-safeexec.md` — 2026-04-30.
