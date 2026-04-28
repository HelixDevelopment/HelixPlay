# AGENTS.md

This file provides guidance to non-Claude AI coding agents (Codex / Cursor / Aider / Continue / Copilot Chat / Cline / etc.) when working with code in this repository. Claude Code reads `CLAUDE.md` instead, but the rules below are identical — they originate in the project's Constitution.

> **Source of truth:** [`docs/research/chapters/MVP/05_Response/01_Constitution.md`](docs/research/chapters/MVP/05_Response/01_Constitution.md).
> Where this file and the Constitution conflict, the Constitution wins.
> The Constitution codifies clauses **R-01..R-18** drawn from
> [`docs/research/chapters/MVP/04_Request.md`](docs/research/chapters/MVP/04_Request.md).
>
> **R-18 (Operational Integrity)** was added 2026-04-28 after a session-
> disruption incident on the operator's development host. It is
> **mandatory and non-negotiable** for every agent — including non-Claude
> agents reading this file. No command, no hook, no container entrypoint,
> no CI lane, no prompt may suspend, hibernate, lock, terminate, or crash
> the operator's active host. The forbidden-commands list and the
> container-runtime hazards inventory live in Constitution §11.5; read
> them in full before generating any Bash command or container manifest.
>
> **Master plan for the documentation programme:**
> [`docs/research/chapters/MVP/05_Response/00_Master_Plan.md`](docs/research/chapters/MVP/05_Response/00_Master_Plan.md).
>
> **System overview:**
> [`docs/research/chapters/MVP/05_Response/02_System_Overview.md`](docs/research/chapters/MVP/05_Response/02_System_Overview.md).

---

## 1. Repository state

This repo currently contains **specifications and research only** — no
source code, no build system. Treat tasks here as documentation /
specification work unless the operator explicitly says they are kicking
off implementation.

The structure under `docs/research/chapters/MVP/`:

- `01_base/` — overall architecture stream (Go backend, Wails / Flutter /
  Angular clients, Sunshine-style host agent, WebRTC + custom UDP, NATS,
  CockroachDB).
- `02_latency/` — zero-latency communication stream (input → render →
  display budget, Reflex, BBR, FEC, DSCP, jitter buffer).
- `03_video_technology/` — capture, codec (H.264/HEVC/AV1), encode,
  recording, audio (PCM, 5.1/7.1, AC3/Dolby) stream.
- `04_Request.md` — **authoritative MVP brief**. Read this before
  answering any planning question. It supersedes the per-stream request
  files where they conflict.
- `05_Response/` — the synthesised "ultimate documentation" deliverables.
  This is the canonical artifact your work feeds into.

Per-stream request files (superseded by `04_Request.md` on conflict):
`01_base/01_Request.md`, `02_latency/01_Request.md`,
`03_video_technology/Request.md` — the last drops the `01_` prefix; do
not normalise. Response folders also drift in casing
(`01_base/02_response/` vs `02_latency/02_Response/` vs
`03_video_technology/02_Response/`); preserve.

The `docs/research/chapters/V1/` directory is a placeholder for the
next phase. When V1 is populated, mirror the MVP layout.

The agent-generated long-form research lives at:

- `MVP/01_base/02_response/cloudgaming.agent.final/cloudgaming.agent.final.md`.
- `MVP/02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md`.
- `MVP/03_video_technology/02_Response/video-tech.agent.final/video-tech.agent.final.md`.

When asked to merge / extend / cross-reference research, dive into these
files (and the per-dim siblings under `Research/research/` /
`Agent_results/research/` / `Agent_Results/research/`) — not the
`.docx` / `.pdf` / `.zip` siblings.

## 2. Build / test / lint

There are none yet. No `package.json`, no `go.mod`, no `Makefile`, no
test runner, no lint configuration. If a task requires running a
command, that is a signal the operator is starting implementation —
and per `04_Request.md` the runtime must be containerised from day one
(Constitution §3), so do not fake a local toolchain to "make it work."

## 3. Mandatory project constraints

The Constitution at
[`docs/research/chapters/MVP/05_Response/01_Constitution.md`](docs/research/chapters/MVP/05_Response/01_Constitution.md)
is the canonical statement. The headlines are:

- **Anti-bluff (Constitution §1).** No `TODO`, `FIXME`, `XXX`, `HACK`,
  `tbd`, `???`, `placeholder`, "implement later", "fill in details"
  artifacts. No empty function bodies (`pass`, `panic("not implemented")`,
  `throw new NotImplementedException()`, `return null` as stand-in).
  No dead code, no dummy / placeholder classes. Tests that pass without
  exercising real behaviour are forbidden. Past incidents on sister
  codebases had green tests on broken features; this MUST NOT recur.
- **Decoupling (Constitution §2).** Reusable components live in **public**
  Git/Go submodules under the `vasic-digital` GitHub & GitLab
  organisations. Reuse existing `vasic-digital` submodules instead of
  recreating; extend them when features are missing.
- **Containers (Constitution §3).** Every service, infra component
  (databases included), build step, test runner, scanner runs inside a
  container. The container definitions live in
  `https://github.com/vasic-digital/Containers` — never vendor a
  `Dockerfile` outside that submodule. CI/CD is local-only and
  container-driven.
- **Stack (Constitution §4).** gRPC over HTTP/3 (QUIC) preferred; REST
  is a separate gateway microservice. NATS / Redis / RabbitMQ used
  where they replace ad-hoc plumbing. Brotli compression. Cronet on
  mobile.
- **Concurrency (Constitution §5).** Non-blocking by default; lazy
  initialisation preferred over eager; bounded queues / semaphores for
  backpressure; allocation-free hot paths; cache-line awareness.
- **Tests (Constitution §6).** Every submodule and every executable
  file is covered by **all ten** test types: Unit, Integration, E2E,
  Security, Benchmarking, Chaos, Stress, Smoke, Full Automation,
  **Challenges**. Only Unit tests may use mocks/stubs/hardcoded values
  (R-12). Latency tests report p50, p99, p999 — averages are
  insufficient. Coverage gate is 100% across the union of test types.
  HelixQA (`git@github.com:HelixDevelopment/HelixQA.git`) is integrated
  into the Challenges pipeline. Challenges
  (`git@github.com:vasic-digital/Challenges.git`) sit at the project
  root and boot the full production stack end-to-end.
- **Quality gates (Constitution §7).** SonarQube, Snyk, Semgrep, Trivy
  / Grype, gitleaks / trufflehog, govulncheck. Critical / High findings
  are merge blockers.
- **Tracking (Constitution §8).** Every phase, task, and subtask is
  mirrored on **GitHub Projects** AND the **GitLab equivalent**, with
  cross-links. ID convention `[Pxx.Tyy.Szz]`. No skipping, no disabling,
  no silent re-scoping.

When drafting plans or specs, structure them as fine-grained
**phases → tasks → subtasks** (R-16). Do not collapse detail to
summarise. The "synthesis programme" methodology (master plan §4) is
the canonical approach — every chapter ends with an
`Anti-Bluff Verification` block listing source paths, web sources, and
conflict-zone resolutions.

## 4. Git topology

Four remotes are configured and kept in sync; `origin` is split (fetch
from GitHub, push to GitFlic). Confirm which mirror(s) the operator
wants when they say "push" unless context is unambiguous — pushing to
`origin` only updates GitFlic.

```
github      git@github.com:HelixDevelopment/HelixPlay.git
gitlab      git@gitlab.com:helixdevelopment1/HelixPlay.git
gitverse    git@gitverse.ru:helixdevelopment/HelixPlay.git
gitflic     git@gitflic.ru:helixdevelopment/helixplay.git
origin      fetch=github, push=gitflic
```

Force-push to any remote requires explicit operator authorisation per
push. `--no-verify` is forbidden. Pre-commit and pre-push hooks failing
means the underlying issue gets fixed first.

## 5. Configuration of the agent harness

This file is ignored by Claude Code (which reads `CLAUDE.md` directly)
but is consulted by other agents. The repo also contains:

- `.claude/settings.json` — Claude Code hooks. Currently registers a
  `Stop` hook running `bash scripts/claim-check.sh` (5 s timeout).
  **The `scripts/` directory does not yet exist**; the hook fails until
  it is created. If you are troubleshooting hook output, flag this
  rather than silently work around it.
- `.claude/settings.local.json` — sets `defaultMode: bypassPermissions`
  for this repo. Tool calls won't prompt; the safety bar is judgement,
  not the permission system. Be careful with destructive git
  operations and with the four configured remotes.
- `Upstreams/*.sh` — one script per remote that exports
  `UPSTREAMABLE_REPOSITORY`. Used by external tooling to pick a target
  remote; not part of any build here.

## 6. What "doing a task" usually looks like

Until code lands, the typical task is one of:

1. **Synthesise across the three MVP research streams** into the
   "ultimate documentation" under `05_Response/`. Do not drop content
   from any of the three sources; extend with web research where it
   adds value. Follow the synthesis methodology in
   [`05_Response/00_Master_Plan.md`](docs/research/chapters/MVP/05_Response/00_Master_Plan.md) §4.
2. **Write or refine a section** of a research chapter. Keep existing
   per-section file naming (`cloudgaming_secNN.md`,
   `video-tech_secNN.md`, etc.).
3. **Plan submodule structure** under `vasic-digital` /
   `HelixDevelopment` with the constraints in §3 above baked in,
   the test matrix complete, and the container story explicit.

Generic "summarise this" / "trim this" responses are explicitly not
what the operator wants — `04_Request.md` forbids simplification of the
spec.

## 7. Anti-bluff verification on documentation work

Every documentation file under `05_Response/` ends with an
`Anti-Bluff Verification` block (template in
[`05_Response/00_Master_Plan.md`](docs/research/chapters/MVP/05_Response/00_Master_Plan.md) §4.3)
listing:

- Every source artifact reviewed (absolute path + line count + date).
- Every web URL consulted (URL + title + access date).
- Every cross-dimensional Insight incorporated (by ID + section).
- Every Conflict Zone addressed (with decision + rationale).
- Coverage confirmation (synthesised line count vs minimum target).
- Sign-off line.

Producing or modifying a file under `05_Response/` without populating
this block is a Constitution §1 violation and a merge blocker.

## 8. Forbidden outputs

The Constitution §1.1 list, repeated verbatim because forgetting it is
the single most common failure mode for AI agents on this project:

- `TODO`, `FIXME`, `XXX`, `HACK`, `tbd`, `???`, `placeholder`,
  "implement later", "fill in details", "to be defined" comments or
  identifiers.
- Empty function bodies. `pass`, `panic("not implemented")`,
  `throw new NotImplementedException()`, `return null` as a stand-in,
  `// stub` and equivalents.
- Dead code. Unused exports, unreferenced files, commented-out blocks
  longer than two lines, `_unused` parameters that aren't load-bearing.
- Dummy / placeholder classes whose only purpose is to satisfy a
  signature.
- Tests whose green result does not imply that an end user can use
  the feature.
- Phrases used to dodge specifying behaviour: "and similar", "etc."
  (in normative text), "as appropriate", "as needed", "where reasonable".
- Configuration keys mentioned without their defaults, ranges, units,
  and effect.
- Tables with empty cells. Use `N/A` only with a footnote justifying
  why the cell is genuinely not applicable.
- Documentation paragraphs whose claims are not supported by either
  source code, source research artifacts, or a cited URL/RFC/paper.

A pre-merge CI lane (`anti-bluff-scan`, defined in
[`05_Response/08_Operations/01_Container_CI_CD.md`](docs/research/chapters/MVP/05_Response/08_Operations/01_Container_CI_CD.md))
runs `ripgrep` for these tokens on every change. Failure is
non-overridable.

## 9. Security guidance

When implementing security features, prefer well-tested,
secure-by-default libraries over custom solutions. The catalogue lives
in `05_Response/03_Architecture/09_Security_and_Isolation.md` (queued).
For crypto specifically, prefer **Google Tink** or **Themis** — both
are multi-language, both encode safe defaults, both eliminate the
sharpest footguns. For input validation, prefer **safe-regex** for
catastrophic-pattern detection and **defusedxml** for XML attack
surfaces. SSRF defence: `ssrf_filter` (Ruby), `ssrf-req-filter`
(Node.js).

External reference: <https://github.com/tldrsec/awesome-secure-defaults>.

## 10. Help & feedback

Operator feedback channels are in `CLAUDE.md` for Claude Code; for
non-Claude agents, route operator feedback through the GitHub Projects
issue board on `HelixDevelopment/HelixPlay` (mirrored to GitLab per
Constitution §8).

End of `AGENTS.md`. Last updated 2026-04-28.
