# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **Source of truth for project rules:** [`docs/research/chapters/MVP/05_Response/01_Constitution.md`](docs/research/chapters/MVP/05_Response/01_Constitution.md).
> Where this file and the Constitution conflict, the Constitution wins. The
> Constitution codifies clauses **R-01..R-18** drawn from `04_Request.md`,
> plus **R-18 (Operational Integrity)** added 2026-04-28 after a session-
> disruption incident — no command, hook, container, CI lane, or agent
> prompt may suspend/hibernate/lock/terminate/crash the operator's host.
> See Constitution §11.5 for the forbidden-commands list and the container
> hazards inventory.
>
> **Synthesis programme master plan:** [`docs/research/chapters/MVP/05_Response/00_Master_Plan.md`](docs/research/chapters/MVP/05_Response/00_Master_Plan.md).
> All chapter work, line targets, dispatch templates, and the work queue
> live there.
>
> **System overview:** [`docs/research/chapters/MVP/05_Response/02_System_Overview.md`](docs/research/chapters/MVP/05_Response/02_System_Overview.md) is the navigation hub for the chapter family under `05_Response/`.
>
> **Non-Claude agents** (Codex, Cursor, Aider, etc.) read `AGENTS.md` at
> the repo root, which carries the same content tailored to those tools.
>
> **User Mandate 2026-04-30:** All submodules MUST respect DRY, KISS, and Top 10
> principles. Lazy initialization is the default (Constitution §5.2). Anti-bluff tests
> MUST guarantee real end-user usability — green tests without working features are a
> Constitution §1 violation. 100% coverage across all ten test types is mandatory.
> Challenges + HelixQA integration is mandatory. All submodules MUST contain
> Constitution, CLAUDE.md, AGENTS.md with these clauses baked in.
 
## Repository state

**This repo currently contains specifications and research only — no source code, no build system.** The MVP implementation has not started. Treat tasks here as documentation/specification work unless the user explicitly says they are kicking off implementation.

The directory layout reflects that:
- `docs/research/chapters/MVP/` — three research streams plus the master request:
  - `01_base/` — overall architecture (Go backend, Wails/Flutter/Angular clients, Sunshine-style host agent, WebRTC + custom UDP, NATS, CockroachDB).
  - `02_latency/` — zero-latency communication research (input→render→display budget, Reflex, BBR, FEC, DSCP, jitter buffer).
  - `03_video_technology/` — capture, codec (H.264/HEVC/AV1), encode, recording, audio (PCM, 5.1/7.1, AC3/Dolby) research.
  - `04_Request.md` — **authoritative MVP brief.** It supersedes the individual chapter requests where they conflict and lists the mandatory project-wide constraints. Read it before answering any planning question.

  Per-chapter request files (superseded by `04_Request.md` on conflict): `01_base/01_Request.md`, `02_latency/01_Request.md`, `03_video_technology/Request.md` — note the last one drops the `01_` prefix; don't normalize, the repo is the source of truth. Response folders also drift in casing: `01_base/02_response/` (lowercase) vs. `02_latency/02_Response/` and `03_video_technology/02_Response/` (capital R).
- `docs/research/chapters/V1/` — placeholder for the next phase (currently only an empty `04_monitoring_and_monitors/` folder). When V1 gets populated, mirror the MVP layout (`01_*/01_Request.md`, …, `04_Request.md`) so structure stays consistent.
- `Upstreams/*.sh` — one script per remote that exports `UPSTREAMABLE_REPOSITORY`. Used by external tooling to pick a target remote; not run as part of any build here.

The agent-generated research lives under each chapter's `02_response/` (or `02_Response/`) folder. The main long-form documents are:
- `MVP/01_base/02_response/cloudgaming.agent.final/cloudgaming.agent.final.md` (plus per-section `cloudgaming_secNN.md` files in `Research/`).
- `MVP/02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md` (plus `Agent_results/` mirror).
- `MVP/03_video_technology/02_Response/video-tech.agent.final/video-tech.agent.final.md` (plus `Agent_Results/` per-section files).

When the user asks you to merge / extend / cross-reference research, these are the files to dive into — not the `.docx`/`.pdf`/`.zip` siblings.

## Build / test / lint commands

There are none yet. No `package.json`, no `go.mod`, no `Makefile`, no test runner, no lint config. If a task requires running a command, that's a signal the user is starting implementation — and per `04_Request.md` the runtime must be containerised from day one, so don't fake a local toolchain to "make it work."

## Mandatory project constraints (from `04_Request.md`)

These are not aspirations; the user has explicitly marked them as required, and they must be carried into anything you produce (code, specs, agent prompts, sub-module CLAUDE.md/AGENTS.md):

- **Anti-bluff:** no `TODO` / `FIXME` placeholders, no dummy classes, no dead code, no skipping. Tests must actually exercise real behavior — green tests on broken features have been a real problem on this project before.
- **Decoupling:** every reusable component goes into its own public submodule under the `vasic-digital` GitHub/GitLab organization (Git submodules and/or Go modules). Reuse existing `vasic-digital` submodules instead of recreating them; extend them if features are missing. Containers are managed via `https://github.com/vasic-digital/Containers`.
- **Tests required for every submodule, 100% coverage, types:** Unit, Integration, E2E, Security, Benchmarking, Chaos, Stress, Smoke, Full automation, and **Challenges** (production-like, full system up — uses `git@github.com:vasic-digital/Challenges.git`). Only Unit tests may use mocks/stubs/hardcoded values; everything else hits the real system. Autonomous QA is `git@github.com:HelixDevelopment/HelixQA.git`.
- **Runtime:** every service, infra component, build, test, scan runs **inside containers**. CI/CD is local, container-driven. Service discovery on the LAN, dynamic port assignment, gRPC preferred (REST + middleware as separate microservices), HTTP/3 (QUIC/Cronet), Brotli compression, Redis/RabbitMQ where useful.
- **Concurrency:** non-blocking by default, lazy init over eager, semaphores/backpressure to prevent clogging, events + observability so state changes propagate in real time.
- **Quality gates:** SonarQube, Snyk, plus other heavy security/quality scans.
- **Submodules carry their own dependencies:** when you add a submodule, also pull in every submodule it depends on, and propagate these constraints into its `CLAUDE.md` / `AGENTS.md` if absent.
- **Tracking:** implementation phases and tickets must be mirrored on GitHub Projects and the GitLab equivalent via their CLIs; the end state is a clean board with nothing skipped or disabled.

When drafting plans or specs, structure them as fine-grained phases → tasks → subtasks; do not collapse detail to summarize.

## Git topology

Four remotes are configured and kept in sync; `origin` is split (fetch from GitHub, push to GitFlic). When the user says "push", confirm which mirror(s) they want unless it's obvious from context — pushing to `origin` only updates GitFlic.

```
github      git@github.com:HelixDevelopment/HelixPlay.git
gitlab      git@gitlab.com:helixdevelopment1/HelixPlay.git
gitverse    git@gitverse.ru:helixdevelopment/HelixPlay.git
gitflic     git@gitflic.ru:helixdevelopment/helixplay.git
origin      fetch=github, push=gitflic
```

## Claude Code configuration

`.claude/settings.json` registers a `Stop` hook that runs `bash scripts/claim-check.sh` (5 s timeout). **The `scripts/` directory does not exist yet**, so the hook will fail until it's created — flag this if the user is troubleshooting hook output, and don't silently work around it.

`.claude/settings.local.json` sets `defaultMode: bypassPermissions` for this repo. Tool calls won't prompt; the safety bar is your judgment, not the permission system. Be especially careful with destructive git operations (force push, reset --hard) and with the four configured remotes.

The same file contains a stray allow-rule referencing `…/HelixAgent/.git/index.lock` — that's a copy/paste residue from a sibling project, not deliberate policy here. Don't extend it; if you're cleaning up, drop that line.

## What "doing a task" usually looks like here

Until code lands, the typical task is one of:
1. **Synthesize across the three MVP research streams** into a single deeper spec (the user calls this the "ultimate documentation"). Do not drop content from any of the three sources; extend with web research where it adds value.
2. **Write or refine a section of a research chapter** under `docs/research/chapters/...`. Keep the existing per-section file naming (`cloudgaming_secNN.md`, `video-tech_secNN.md`, etc.).
3. **Plan submodule structure** under `vasic-digital` / `HelixDevelopment` — every plan needs the constraints above baked in, plus the test matrix and container story.

Generic "summarize this" or "trim this" responses are explicitly not what the user wants — `04_Request.md` forbids simplification of the spec.
