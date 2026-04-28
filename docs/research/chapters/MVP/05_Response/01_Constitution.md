# HelixPlay Project Constitution

> **Slogan:** "Ultimate gaming experience!"
>
> **Authority:** This document is the single source of truth for the
> non-negotiable rules of the HelixPlay project. It is derived from
> `docs/research/chapters/MVP/04_Request.md` and supersedes any guidance
> that contradicts the clauses below.
>
> **Audience:** Every human contributor, every AI agent (Claude Code,
> Codex, Cursor, Aider, etc.), every CI runner, every reviewer.
>
> **Propagation:** Every submodule under the `vasic-digital` and
> `HelixDevelopment` organisations that participates in HelixPlay MUST
> reference this Constitution from its own `CLAUDE.md`, `AGENTS.md`, and
> `CONSTITUTION.md` (creating those files if absent). Reference is by
> stable URL, not by copy-paste, so updates propagate automatically.
>
> **Stable URL (after first push):**
> `https://github.com/HelixDevelopment/HelixPlay/blob/main/docs/research/chapters/MVP/05_Response/01_Constitution.md`

---

## 0. Preamble

The HelixPlay project has, in prior iterations on sister codebases
(see CLAUDE.md, "We had been in position that all tests do execute with
success and all Challenges as well, but in reality the most of the
features does not work…"), suffered from **green tests on broken
features**. This Constitution exists to make that failure mode
structurally impossible.

The Constitution is short, normative, and uncompromising. Where the
phrase **MUST**, **MUST NOT**, **SHALL**, or **SHALL NOT** appears, it is
used in the RFC 2119 sense. There are no soft preferences in this
document. Soft preferences live in chapter-specific documents under
`05_Response/`.

---

## 1. The Anti-Bluff Pledge (R-02, R-13)

### 1.1 What is forbidden

The following artifacts MUST NOT exist anywhere in the HelixPlay
codebase, documentation, configuration, or scripting:

- `TODO`, `FIXME`, `XXX`, `HACK`, `tbd`, `???`, `placeholder`,
  "implement later", "fill in details", "to be defined" comments or
  identifiers.
- Empty function bodies. `pass`, `panic("not implemented")`,
  `throw new NotImplementedException()`, `return null` as a stand-in,
  `// stub` and equivalents.
- Dead code. Unused exports, unreferenced files, commented-out blocks
  longer than two lines, `_unused` parameters that aren't load-bearing.
- Dummy / placeholder classes whose only purpose is to satisfy a
  signature. If a class exists, it MUST do real work.
- Tests that pass without exercising the system. Tests whose failure
  mode is "the test framework crashed" or "the assertion was vacuously
  true." Tests whose green result does not imply that an end user can
  use the feature.
- Phrases used to dodge specifying behaviour: "and similar", "etc."
  (in normative text), "as appropriate", "as needed", "where reasonable".
  These are allowed in prose only when they describe past evidence,
  never future behaviour.
- Configuration keys mentioned without their defaults, ranges, units,
  and effect.
- Tables with empty cells. Use `N/A` only with a footnote justifying
  why the cell is genuinely not applicable.
- Documentation paragraphs whose claims are not supported by either
  source code, source research artifacts, or a cited URL/RFC/paper.

### 1.2 What is required

For every feature, every fix, every refactor, every documentation page:

1. The change does what its description says it does, and that statement
   is verifiable by an external observer who has not been told the
   answer.
2. There exists at least one **non-Unit** test (see §6) that exercises
   the change end-to-end against the real production-equivalent system.
3. The Anti-Bluff Verification block (defined in
   `05_Response/00_Master_Plan.md` §4.3) is filled in for documentation
   changes.
4. The change is reviewed by at least one party other than the author,
   and that party explicitly confirms they ran the verifying test.

### 1.3 Enforcement

A pre-merge CI lane (`anti-bluff-scan`, defined in
`05_Response/08_Operations/01_Container_CI_CD.md`) runs:

- A `ripgrep` pass for the forbidden tokens listed in §1.1.
- A coverage delta check confirming the change is covered by at least
  one non-Unit test type.
- A "documentation change must include verification block" check on any
  `05_Response/` modification.

The lane's failure is non-overridable. Bypass requires a documented
exception via §13.

---

## 2. Decoupling & Submodule Discipline (R-03, R-04, R-15)

### 2.1 Reusability bar

Every component that could plausibly be useful in a project other than
HelixPlay MUST be packaged as a public submodule under
`https://github.com/vasic-digital` (and mirrored on GitLab where the
organisation has presence). The bar for "plausibly useful" is
deliberately low — when in doubt, extract.

### 2.2 Reuse first

Before authoring a new submodule, the contributor MUST search the
existing `vasic-digital` inventory (and the `HelixDevelopment` inventory
where applicable). If a submodule already covers the need, it MUST be
reused. If it covers the need partially, it MUST be **extended**, with a
PR to its own repository, not forked or duplicated.

The submodule catalog at `05_Response/06_Submodules/01_Submodule_Catalog.md`
is the canonical inventory for this project. Any contribution that
introduces a new submodule MUST update the catalog in the same change
set.

### 2.3 Recursive dependency capture

When a submodule is added to HelixPlay, **all of its own dependency
submodules** are also added to HelixPlay (R-15). The `.gitmodules`
graph MUST be transitively complete — no submodule may rely on a sibling
that is not present in the workspace.

### 2.4 Public visibility

All HelixPlay submodules under `vasic-digital` MUST be public. Private
or internal-only repositories are not permitted as direct dependencies
of HelixPlay. Closed-source third-party SDKs are permitted only under
a documented exception (§13) and only when wrapped by a public
submodule that abstracts the dependency.

### 2.5 Constitutional propagation

Every submodule MUST contain:

- A `CLAUDE.md` that references this Constitution by stable URL.
- An `AGENTS.md` with equivalent content for non-Claude agents.
- A `CONSTITUTION.md` that mirrors the relevant subset of these
  clauses (the entire Constitution is acceptable; subsetting requires
  documenting which clauses are inherited).

The CI lane on each submodule MUST verify the references on every push.

---

## 3. Containerised Runtime (R-05, R-06)

### 3.1 Universal containerisation

Every executable artifact in HelixPlay MUST run inside a container.
This includes, without exception:

- Application services (backend microservices, host agent, clients
  where containerisable).
- Infrastructure components: databases (CockroachDB, Redis, NATS),
  message brokers (RabbitMQ), reverse proxies, observability stacks.
- Build steps. The "build host" is a build container.
- Test runners. Each test type runs in its own container topology
  defined in `05_Response/07_Testing/`.
- Static analysers, security scanners, formatters, linters.
- Local CI/CD runners.

### 3.2 The `Containers` submodule

`https://github.com/vasic-digital/Containers` is the **only** location
in which container definitions live. HelixPlay MUST NOT vendor its own
`Dockerfile` outside that submodule. If a needed image is missing,
contribute it to the `Containers` submodule.

### 3.3 Local-only CI/CD

CI/CD pipelines MUST be runnable end-to-end on a developer workstation
(LAN-only, no SaaS dependencies). Cloud CI providers are permitted
only as **mirrors** of the local pipeline; they cannot be the canonical
gate. Phase 01 (`05_Response/09_Implementation_Phases/Phase_01_Containers_and_CI.md`)
specifies the local-runner topology in detail.

### 3.4 Reproducibility

Every container image MUST be reproducible: same source revision +
same base image digest = same output digest. Image tags MUST be
content-addressed (digest-pinned) in production deployment manifests;
floating tags (`latest`, `stable`, plain version like `:1.2`) are
permitted only inside the `Containers` submodule's build matrix.

### 3.5 Networking & ports

Service-to-service communication MUST use service discovery on the
LAN (mDNS or a registry service — exact choice in
`05_Response/08_Operations/03_Service_Discovery_and_Ports.md`). Ports
MUST be assigned dynamically; static `8080`-style hard-coding is
forbidden outside developer convenience scripts.

---

## 4. Communication Stack (R-07, R-08)

### 4.1 Default protocol: gRPC

Internal service-to-service communication defaults to **gRPC over
HTTP/3 (QUIC)**, with Protocol Buffers schemas living in a public
submodule under `vasic-digital` so other projects can consume them.

### 4.2 REST as a microservice

Where REST is required (browser ergonomics, third-party integrations),
it MUST be implemented as a **dedicated REST gateway microservice**
that translates to/from the canonical gRPC services. REST is never the
canonical contract.

### 4.3 Real-time fan-out

For server-to-client real-time updates the order of preference is:

1. **WebRTC DataChannels** for game-input and stream-control payloads
   (per the Streaming chapter `05_Response/03_Architecture/01_Streaming_Protocols_and_Codecs.md`).
2. **gRPC server-streaming over HTTP/3** for typed, low-fan-out events.
3. **Server-Sent Events (SSE)** for browser-native one-way streams.
4. **WebSocket** only when none of the above suffice.

### 4.4 Eventing & queues (R-08)

Heavy use of events is mandatory. The default event bus is **NATS /
NATS JetStream**; **Redis** is used for caches and pub/sub of ephemeral
state; **RabbitMQ** is used where strict per-message ack semantics or
plug-in integrations are required. Selection criteria are documented
per use case in `05_Response/03_Architecture/05_RealTime_APIs.md`.

### 4.5 Compression

**Brotli** is the default response compression for HTTP. gzip is
permitted as a fallback only for clients that do not negotiate Brotli.
Static assets are pre-compressed with both at build time.

### 4.6 Cronet on mobile

Mobile clients MUST use **Cronet** (or an equivalent QUIC client
library) to ensure HTTP/3 end-to-end. Plain `okhttp`/`URLSession`
without HTTP/3 is forbidden in production paths.

---

## 5. Concurrency Posture (R-09)

### 5.1 Non-blocking by default

All network I/O, all disk I/O, all IPC MUST be non-blocking.
Synchronous-blocking calls are permitted only inside well-bounded
helper functions called from goroutine workers (or platform-equivalent
async contexts).

### 5.2 Lazy initialisation

Resources are constructed on first use, not at startup, unless
eager construction is required for a documented reason (e.g.
warming a JIT, pre-allocating memory pools per Insight #4 of the
latency stream). Justifications live next to the call site.

### 5.3 Backpressure & semaphores

Every producer/consumer pair MUST have an explicit bounded buffer.
Unbounded channels, queues, or work pools are forbidden. Where
backpressure cannot reasonably propagate to the producer, an explicit
**drop policy** MUST be documented and a metric MUST count drops.

### 5.4 Allocation discipline (Latency Insight #4)

On the streaming hot path (controller → render → encode → network
on the host; receive → decode → display on the client), there MUST
be **zero dynamic allocations** after warmup. Pre-allocated pools,
ring buffers, and object reuse are mandatory.

### 5.5 Cache-line awareness

Shared mutable state MUST be cache-line padded (64 bytes on x86-64,
128 bytes on the ARM platforms documented to need it). False sharing
detected by `perf c2c` in benchmark runs is treated as a Sev-2 bug.

---

## 6. Testing Discipline (R-11, R-12, R-13)

### 6.1 The Ten

Every submodule and every executable file of HelixPlay MUST be covered
by **all ten** of the following test types. None may be omitted:

1. **Unit** — fast, isolated, mocks/stubs/hardcoded values **permitted**.
2. **Integration** — multiple components, real dependencies, no mocks
   except for systems genuinely outside the project's control (and
   even those should use a recorded-replay layer where possible).
3. **End-to-End (E2E)** — full system path, real production-like
   topology, no mocks.
4. **Security** — fuzzing, dependency scanning (Snyk), SAST (SonarQube,
   Semgrep, etc.), DAST against deployed containers.
5. **Benchmarking** — performance regressions are CI-blocking.
   Latency benchmarks report **p50, p99, p999** (per Latency Insight #2);
   averages alone are insufficient.
6. **Chaos** — fault injection (kill containers, drop packets, flip
   bits in a controlled set, NUMA pinning errors, IRQ storms).
7. **Stress** — load up to and beyond design capacity to characterise
   failure modes.
8. **Smoke** — minimum viable check that runs in seconds and gates
   every promotion.
9. **Full Automation** — the entire pipeline runs without human input
   from clean checkout to deployable artifact and back, on a schedule.
10. **Challenges** — production-equivalent scenarios driven through
    the real, fully booted system. Source: `git@github.com:vasic-digital/Challenges.git`.

### 6.2 Mocks are confined to Unit

The phrase "mocks/stubs/hardcoded values" means **anything that is not
the production implementation**. They are permitted **only** in Unit
tests. Every other test type MUST drive the real binary path with
real (or production-equivalent) infrastructure. Violations are merge
blockers.

### 6.3 Anti-Bluff Tests

A test is **anti-bluff** when its green result implies that an end user
can use the feature it covers. To qualify:

- The test asserts on **observable behaviour** (HTTP responses,
  database state, rendered frames, audio output, log lines reaching
  the observability backend) — not on internal call counts to mocks.
- The test runs against the same image artifact that ships to
  production.
- The assertion includes a **negative leg**: removing the feature
  must cause the test to fail. (This is the structural defence
  against vacuous tests.)
- The test's reason for failure, when it does fail, must be
  diagnostic — a log fragment or screenshot lands in the artifacts
  bundle.

### 6.4 Coverage gate

Every submodule's CI MUST enforce **100% line, branch, and function
coverage across the union of test types** — not unit-only coverage.
The gate is non-overridable.

### 6.5 HelixQA integration

The autonomous QA system at `git@github.com:HelixDevelopment/HelixQA.git`
is integrated into the Challenges pipeline. HelixQA runs unattended
against the production-equivalent topology and files findings as
issues mirrored on GitHub Projects + GitLab (R-17). HelixQA's findings
are normal P1/P2 work items, not advisory.

### 6.6 The Challenges pattern

Challenges (R-14) are integrated as in HelixAgent and Catalogizer:
the Challenges submodule sits at the project root and provides a
`make challenge` (or container equivalent) that boots the full
production stack and runs end-to-end scenarios. Failures stop the
pipeline.

---

## 7. Quality Gates (R-10)

### 7.1 Mandatory scanners

Every submodule's CI runs at minimum:

- **SonarQube** — code quality, smells, hotspots, coverage validation.
- **Snyk** — dependency vulnerabilities, license compliance, IaC scanning.
- **Semgrep** — pattern-based static analysis (rules in a public
  `vasic-digital/semgrep-rules` submodule).
- **Trivy / Grype** — container image vulnerability scanning.
- **gitleaks / trufflehog** — secret scanning, on every commit.
- **govulncheck** — Go-specific vulnerability scanning, where Go is in use.

Additional scanners SHOULD be added per language ecosystem (e.g.
`bandit` for Python helpers, `cargo-audit` for any Rust components).
The full canonical list lives in `05_Response/08_Operations/02_Quality_Gates_SonarQube_Snyk.md`.

### 7.2 Severity gating

- **Critical** / **High** findings: merge blocker, no exceptions
  except documented §13 exceptions with a fixed expiry date.
- **Medium**: must have a tracked ticket within 7 days.
- **Low** / **Informational**: tracked, addressed at convenience.

### 7.3 Local equivalence

All scanners MUST run locally inside the standard toolchain
container (R-06). Cloud-only scanners are not acceptable as the
canonical gate.

---

## 8. Tracking (R-16, R-17)

### 8.1 Both platforms

Every phase, task, and subtask MUST exist as an issue on both
**GitHub Projects** (`HelixDevelopment/HelixPlay`) and the
**GitLab equivalent** (`gitlab.com/helixdevelopment1/HelixPlay`),
with the issues mutually cross-linked.

### 8.2 ID convention

`[Pxx.Tyy.Szz]` — phase, task, subtask. Phase IDs map 1:1 to files
under `05_Response/09_Implementation_Phases/`.

### 8.3 No skipping, no disabling

Per R-16, a clean board at end-of-phase requires every ticket
**resolved** (closed with evidence) — never skipped, simplified,
omitted, or marked "won't fix" without a §13 exception. Re-scoping
mid-flight requires a documented decision, a successor ticket
linked from the original, and a Constitution-aware review.

### 8.4 Closure evidence

A ticket may be closed only when its evidence link points to:

- The merged PR(s).
- The chapter's Anti-Bluff Verification block (for documentation work).
- The CI run that proves the relevant tests passed.
- Where applicable, the HelixQA Challenges run that exercised the
  feature in production-equivalent topology.

### 8.5 Audit log

`05_Response/00_Master_Plan.md` §10 (Session Log) is append-only and
is the human-readable audit trail. Tickets are the machine-readable
trail.

---

## 9. Source Control & Git Topology

### 9.1 The four remotes

HelixPlay's primary repository is mirrored across four remotes:

```
github      git@github.com:HelixDevelopment/HelixPlay.git
gitlab      git@gitlab.com:helixdevelopment1/HelixPlay.git
gitverse    git@gitverse.ru:helixdevelopment/HelixPlay.git
gitflic     git@gitflic.ru:helixdevelopment/helixplay.git
origin      fetch=github, push=gitflic
```

### 9.2 Push policy

- A "release" push updates **all four** remotes. The exact command
  sequence is documented in
  `05_Response/08_Operations/06_Git_Topology_and_Push_Policy.md`.
- A "wip" push to `origin` updates **only GitFlic** (because of the
  split fetch/push configuration). Contributors MUST confirm with
  the operator which remote(s) are intended unless the context is
  unambiguous.
- Force-push to any remote requires explicit operator authorisation
  per push (no blanket grants).
- `--no-verify` is forbidden. Pre-commit and pre-push hooks failing
  means the underlying issue is fixed first.

### 9.3 Branching

- `main` is the default branch on all four remotes.
- Feature work happens on `feat/<short-name>` branches.
- Documentation work happens on `docs/<short-name>` branches.
- Branch protection MUST require the full CI matrix to be green
  before merge.

### 9.4 Commits

- Commit messages follow Conventional Commits.
- Every commit message links to the corresponding `[Pxx.Tyy.Szz]`
  ticket where applicable.
- Co-author trailers are mandatory when AI agents materially
  contributed (per repo-root CLAUDE.md guidance).

---

## 10. Observability (R-08)

### 10.1 Three pillars + one

- **Logs** — structured JSON, shipped to a central aggregator.
- **Metrics** — Prometheus-compatible exposition, scraped on a
  per-service basis.
- **Traces** — OpenTelemetry, propagated through gRPC headers.
- **Events** — domain events on the NATS bus (R-08), consumable by
  any subscriber. State changes propagate in real time per the brief.

### 10.2 Hot-path budget

On the streaming hot path, observability MUST be **sample-based**
(e.g. 0.1% trace sampling) to avoid contaminating the latency budget.
The sample rate is documented per service in
`05_Response/08_Operations/04_Observability_and_Events.md`.

### 10.3 Mandatory metrics

Every service exposes at minimum:

- p50, p99, p999 latency for every public RPC (Latency Insight #2).
- Error rate by class (4xx, 5xx, transport).
- Saturation indicators per the USE method.
- Dropped-message counts for any backpressure-applied queue (§5.3).

---

## 11. Security & Privacy

### 11.1 Defence in depth

- All RPCs use **mTLS** between services.
- All client-server traffic uses TLS 1.3 (or DTLS 1.2/1.3 for
  WebRTC media).
- Secrets are managed by an external secret store (Vault or equivalent)
  and injected at container boot. No secrets in environment variables
  baked into images, no secrets in repositories.
- All input is validated at the boundary; no trust is granted to
  internal callers.

### 11.2 Authentication & authorisation

- OAuth2 / OIDC for human users, with **Device Authorization Grant
  (RFC 8628)** on input-constrained devices (TVs, consoles).
- Short-lived JWT access tokens with refresh-token rotation.
- mTLS-based service identity for service-to-service calls.

### 11.3 Anti-cheat compatibility

The host agent MUST use **only OS-provided capture APIs** (DXGI DDA,
ScreenCaptureKit, KMS/PipeWire). Hook-based capture is forbidden.
Virtual controller drivers MUST be properly signed (and ideally
WHQL-certified on Windows). The "clean host" model from cloudgaming
Insight #5 is mandatory.

### 11.4 Privacy

- Player input is treated as personal data. Logging at the input layer
  is restricted by default and requires consented telemetry.
- 4K capture frames are never persisted by default; recording is
  opt-in and routed to user-controlled storage (per video-tech
  Insight #4).

### 11.5 Operational Integrity (R-18) — no host disruption

> **Origin:** added 2026-04-28 after a Session-2/3 incident in which
> the operator's development host was suspended / signed out mid-
> session, terminating an active orchestrator and its in-flight
> subagents. A read-only investigation (recorded in
> `00_Master_Plan.md` §10 Session 3 row) found **no causal link**
> from any tool call we issued to the host disruption — no
> Bash command we invoked has the OS privilege to suspend, hibernate,
> or terminate a login session, and no Docker / Podman container was
> instantiated during the session — but the absence of an explicit
> rule made the audit slower than it should have been. This clause
> closes that gap.

#### 11.5.1 Forbidden commands (host-disruptive)

The following commands and equivalents MUST NEVER appear in
HelixPlay code, scripts, hooks, container entrypoints, CI lanes, or
agent prompts. The list is non-exhaustive — the **principle** is
"no command that disrupts the operator's active host".

- **Power state**: `systemctl suspend`, `systemctl hibernate`,
  `systemctl hybrid-sleep`, `systemctl suspend-then-hibernate`,
  `systemctl poweroff`, `systemctl reboot`, `systemctl halt`,
  `pm-suspend`, `pm-hibernate`, `pm-suspend-hybrid`, `shutdown` (any
  flag), `reboot`, `halt`, `init 0`, `init 6`, `rtcwake`, `s2ram`,
  `s2disk`.
- **Session termination**: `loginctl terminate-session`,
  `loginctl kill-session`, `loginctl lock-session`,
  `loginctl terminate-user`, `loginctl kill-user`, `pkill -KILL -u
  $USER`, `killall -u $USER`, `gnome-session-quit`, `qdbus
  org.kde.ksmserver`, `pkill -HUP $XSESSION`.
- **DBus power-management calls**: `dbus-send` /
  `gdbus call` / `qdbus` to any of `org.freedesktop.login1.Manager.Suspend`,
  `org.freedesktop.login1.Manager.Hibernate`,
  `org.freedesktop.login1.Manager.PowerOff`,
  `org.freedesktop.login1.Manager.Reboot`,
  `org.freedesktop.login1.Manager.LockSessions`,
  `org.freedesktop.login1.Manager.TerminateSession`,
  `org.freedesktop.login1.Manager.TerminateUser`,
  `org.freedesktop.ScreenSaver.Lock`,
  `org.gnome.SessionManager.Logout`,
  `org.gnome.SessionManager.Shutdown`,
  `org.gnome.SessionManager.Reboot`.
- **Display manipulation that masquerades as suspend**: `xset dpms
  force off`, `xset dpms force standby`, `xset s activate`,
  `wlopm --off '*'`, `swaymsg "output * power off"`.
- **Memory / swap manipulation that can freeze the host**: `swapoff
  -a`, `mkswap` on the active swap, large `dd if=/dev/zero of=/...`
  writes, fork bombs (`:(){ :|:& };:`), `stress-ng` without resource
  caps, `memhog` without caps.
- **Filesystem / device sabotage**: `dd if=/dev/zero of=/dev/sd*`,
  `mkfs.*` on a non-loopback device, `wipefs -a` on the host's disks,
  unmounting `/`, `/home`, `/run`, `/proc`, `/sys`, `/dev`.
- **Kernel module manipulation that can wedge the host**: `rmmod` on
  storage / display / input drivers, `modprobe -r` on same.
- **Init / service manager wedging**: `systemctl daemon-reexec` (when
  the active session depends on systemd-managed services), `kill -KILL 1`,
  `kill -KILL -1`.

#### 11.5.2 Containers — forbidden and guarded

Even though HelixPlay's Containers submodule is the canonical home
for every image, the **container client** itself is also constrained:

- **NEVER invoke** `docker system prune --all --force` or `podman
  system prune --all --force --volumes` on the operator's host
  outside a documented `make clean-slate` target that the operator
  invokes deliberately. Image / volume nuking is destructive and
  can dwarf the operator's reasonable expectations. If a "clean-
  slate destroy and rebuild" is required, the action is gated by an
  explicit operator confirmation (e.g. an interactive prompt or a
  Make target with `clean-slate` in the name) — never a default.
- **NEVER mount the host's `/`, `/home`, `/run`, `/proc`, `/sys`, or
  `/dev` into a container** unless the design explicitly requires
  it (e.g. host-agent capture container needs `/dev/dri/card0`,
  `/dev/uinput`, or `/dev/input/event*` — these are fine if scoped
  to the specific device file). Mounting parent directories is a
  privilege escalation vector and is forbidden.
- **NEVER run a container with `--privileged`** unless the design
  explicitly requires it (with §13 exception); prefer `--cap-add` /
  `--cap-drop` granularity, `--device` for specific devices, and
  user namespaces.
- **NEVER set host networking (`--network host`)** unless the design
  explicitly requires it (with §13 exception); prefer bridge / macvlan
  with documented port mappings.
- **NEVER bind to a privileged port (≤1024)** without the operator
  understanding the implication — prefer `>1024` with a reverse-proxy
  in front.

#### 11.5.3 Container-runtime hazards

Both Docker and Podman are themselves capable of triggering host-
side instability under operator misuse. Documented hazards we
acknowledge and avoid:

- **OOM cascade**: an unbounded `docker run` with `--memory=unlimited`
  on a host without swap can trigger the OOM killer to victimize the
  display server or the user session. **Mandatory mitigation**:
  every container declares `--memory` and `--memory-swap` limits.
- **Disk-fill via logs**: container stdout/stderr streamed to
  `json-file` with no rotation can fill `/var/lib/docker` and freeze
  the host. **Mandatory mitigation**: every container uses
  `--log-driver=local` (or equivalent) with size + count limits, OR
  forwards logs to the observability collector immediately.
- **Privileged volume mounts breaking the boot path**: mounting host
  `/etc/systemd` or `/boot` read-write inside a container that then
  edits these directories can break the next boot. **Mandatory
  mitigation**: never mount these paths; if you must, only `:ro`.
- **CPU starvation by busy containers without `--cpus`**: long-
  running CPU-bound containers without `--cpus` can starve the
  display server's render thread, leading to a perceived "host
  freeze" indistinguishable from suspend. **Mandatory mitigation**:
  every container declares `--cpus`.
- **Mount-namespace sealing under cgroups v2**: certain container
  runtimes on cgroups v2 hosts can leak mount namespaces if killed
  ungracefully, leaving stale mount points that block subsequent
  shutdowns. **Mandatory mitigation**: container teardown uses
  `docker stop` / `podman stop` (SIGTERM with timeout), never
  `kill -9` of the runtime process.

These hazards are documented further in the future
`05_Response/06_Submodules/02_Containers_Submodule.md` chapter and
in the `vasic-digital/Containers` submodule's own CONSTITUTION.md.

#### 11.5.4 Hooks and CI

`.claude/settings.json`, `.claude/settings.local.json`, every CI
script, every git hook, every container entrypoint, every test
runner MUST be reviewed against §11.5.1. The pre-merge
`anti-bluff-scan` CI lane (defined in
`05_Response/08_Operations/01_Container_CI_CD.md`) extends with a
**`host-integrity-scan`** sub-lane that `ripgrep`s for the patterns
in §11.5.1 across the entire repository, the rendered prompts of
all agent dispatches, and the merged container manifests. The lane
is non-overridable; bypass requires §13 exception with a documented
mitigation.

#### 11.5.5 Subagent and tool-budget isolation

When the orchestrator dispatches subagents (Master Plan §5 R1 model),
each subagent inherits the same forbidden-commands list. Subagents
do **not** receive elevated privileges; their tool budgets are
identical to the orchestrator's. The operator's host MUST NOT be
disrupted by any subagent's activity. If a subagent's prompt is
detected to attempt any §11.5.1 pattern, the dispatch is aborted
and the offending prompt template is rewritten before re-dispatch.

#### 11.5.6 Recovery posture

If the operator's session **is** disrupted (suspend, sign-out, host
crash) for reasons outside our control, recovery posture is:

1. **No automatic re-execution** of in-flight subagents on resume.
   The operator is the source of truth for whether to resume.
2. **Investigation first**: read-only `git status`, `git submodule
   status`, `.claude/settings*.json`, container daemon state — to
   establish whether anything we issued is implicated.
3. **Honest finding**: if the investigation finds no causal link,
   say so. Do not fabricate a cause to seem helpful (Constitution §1).
4. **Document the incident** in `00_Master_Plan.md` §10 Session Log
   with a new row, including: timestamps, observed symptoms,
   investigation steps, conclusion, and any new clause added to this
   Constitution as a result.
5. **Resume only after** the operator has confirmed the investigation
   conclusion and the recovery action.

#### 11.5.7 Anti-bluff testing reinforcement (cross-link §1, §6.3)

Reaffirmed because the operator emphasised it after the Session-3
incident: **green tests must guarantee real, end-user-usable
behaviour**. A test that passes without exercising the system is a
Constitution §1 violation regardless of how convenient the green
result is. Every submodule's CI MUST include a "negative leg" test
per Constitution §6.3 — removing the feature must cause the test to
fail. Constitution §6.3 already mandates this; this sub-clause
ensures the rule is propagated to every submodule's CONSTITUTION.md
when those submodules come into existence.

---

## 12. Documentation Discipline

### 12.1 Living documents

Documents under `05_Response/` are **living**. They are updated as the
system changes. A documentation drift longer than 30 days is a Sev-3
defect.

### 12.2 No simplification

Per R-01, simplification is forbidden. Documents may be reorganised
for clarity but the total information content MUST monotonically
increase, never decrease, across revisions. Removed content moves to
a dated archive entry, not the bin.

### 12.3 Cross-linking

Every section that touches another section's territory MUST link to
it bidirectionally. A reader can navigate from any leaf back to the
Master Plan in ≤3 clicks.

### 12.4 Diagrams

Diagrams referenced from chapters MUST live in version-controlled
source form (Mermaid, Graphviz, PlantUML, or original Figma exports
with the source `.fig` committed). Pre-rendered PNGs are kept as
secondary artifacts; the source is the canonical artifact.

---

## 13. Exceptions

### 13.1 When an exception is allowed

A rule in this Constitution may be temporarily set aside only when:

1. Compliance is **technically impossible** for a documented reason.
2. Compliance would actively harm the project's mission (e.g. a
   third-party SaaS dependency cannot be containerised without
   contractual breach).
3. A fixed expiry date is set, by which compliance is restored.

### 13.2 Exception process

- The exception is filed as an issue on **both** GitHub Projects and
  GitLab, with title `[EXCEPTION] <clause-id> <short-name>`.
- The body documents: the clause, the reason, the impact, the expiry
  date, and the compensating controls.
- The exception requires approval from at least two reviewers, one of
  whom is **not** the author.
- The exception is referenced in the affected file via a comment of
  the form `// CONSTITUTION-EXCEPTION: <issue-id>`.
- On the expiry date, the exception lapses automatically and CI starts
  failing again unless the exception is re-approved.

### 13.3 No silent exceptions

A pattern in the codebase that violates the Constitution and is **not**
linked to an active exception is a defect, regardless of how long it
has existed.

---

## 14. Definitions

For unambiguous interpretation:

- **Hot path** — the code path traversed for every controller event
  (host) or every received frame (client). All other paths are
  cold paths.
- **Production-equivalent** — runs the same container image, with the
  same configuration shape (different secrets/sizing permitted), as
  the eventual production deployment.
- **Anti-bluff** — see §1.
- **Submodule** — a Git submodule living under `vasic-digital` or
  `HelixDevelopment`, public, with its own Constitution reference.
- **The Ten** — the ten test types listed in §6.1.
- **R-NN** — a clause from `04_Request.md` reproduced in
  `05_Response/00_Master_Plan.md` §1.

---

## 15. Amendment Procedure

This Constitution may be amended only by:

1. Drafting the change in a feature branch.
2. Updating the Master Plan's Session Log to reflect the change.
3. Securing approval from the project operator (the human who owns
   the four remotes).
4. Merging the amendment in a single PR, then propagating the change
   to every submodule's Constitution reference.

Amendments **never relax** R-01..R-17 unless the operator has explicitly
revised `04_Request.md`. Tightening (adding new clauses, narrowing
existing ones) is allowed without operator approval but still requires
the procedure above.

---

## 16. Acceptance

By contributing to HelixPlay (code, documentation, configuration,
tickets, reviews), every contributor — human or agent — accepts this
Constitution. There is no opt-out. There are no exceptions to §1
(Anti-Bluff). All other clauses may be excepted only via §13.

---

## Anti-Bluff Verification

### Source Evidence Reviewed
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/04_Request.md` — 99 lines, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/CLAUDE.md` — full file, reviewed 2026-04-28.
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/01_base/02_response/Research/research/cloudgaming_insight.md` — 156 lines, reviewed 2026-04-28 (Insight #5 anti-cheat clean host informs §11.3).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/02_latency/02_Response/Agent_results/research/latency_insight.md` — 100 lines, reviewed 2026-04-28 (Insights #2, #4 inform §5, §10.3).
- `/run/media/milosvasic/DATA4TB/Projects/HelixPlay/docs/research/chapters/MVP/03_video_technology/02_Response/Agent_Results/research/video-tech_insight.md` — 243 lines, reviewed 2026-04-28 (Insight #4 informs §11.4).

### Web Sources Consulted
- None for Constitution v1. Web research is reserved for the technical
  chapters where it adds substantive evidence; the Constitution is
  derived solely from the operator's `04_Request.md` and the project's
  own research artifacts.

### Insights Incorporated
- cloudgaming_insight #5 (Anti-Cheat Clean Host) → §11.3.
- latency_insight #2 (p999 only metric) → §6.1, §10.3.
- latency_insight #4 (Allocation-free hot path) → §5.4.
- video-tech_insight #4 (Recording mirrors save patterns) → §11.4.

### Conflict Zones Resolved
| CZ-ID | Conflict | Decision | Rationale |
|-------|----------|----------|-----------|
| n/a   | None at Constitution level | n/a | Conflict zones are technology-specific and resolved in the chapter that owns the decision. |

### Coverage Confirmation
- The Constitution's purpose is normative, not encyclopedic. R-01's
  line-floor applies to the synthesis programme as a whole, not to
  this single document.

### Sign-off
Executed by: Claude (orchestrator session 1)
Reviewed by: pending operator review
Date: 2026-04-28

End of Constitution v1 — 2026-04-28.
