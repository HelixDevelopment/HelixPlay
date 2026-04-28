# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository state

**This repo currently contains specifications and research only — no source code, no build system.** The MVP implementation has not started. Treat tasks here as documentation/specification work unless the user explicitly says they are kicking off implementation.

The directory layout reflects that:
- `docs/research/chapters/MVP/` — three research streams plus the master request:
  - `01_base/` — overall architecture (Go backend, Wails/Flutter/Angular clients, Sunshine-style host agent, WebRTC + custom UDP, NATS, CockroachDB).
  - `02_latency/` — zero-latency communication research (input→render→display budget, Reflex, BBR, FEC, DSCP, jitter buffer).
  - `03_video_technology/` — capture, codec (H.264/HEVC/AV1), encode, recording, audio (PCM, 5.1/7.1, AC3/Dolby) research.
  - `04_Request.md` — **authoritative MVP brief.** It supersedes the individual chapter requests where they conflict and lists the mandatory project-wide constraints. Read it before answering any planning question.
- `docs/research/chapters/V1/` — placeholder for the next phase (currently only an empty `04_monitoring_and_monitors/` folder).
- `Upstreams/*.sh` — one script per remote that exports `UPSTREAMABLE_REPOSITORY`. Used by external tooling to pick a target remote; not run as part of any build here.

The agent-generated research lives under each chapter's `02_response/` (or `02_Response/`) folder. The main long-form documents are:
- `MVP/01_base/02_response/cloudgaming.agent.final/cloudgaming.agent.final.md` (plus per-section `cloudgaming_secNN.md` files in `Research/`).
- `MVP/02_latency/02_Response/ZeroLatency_Communication_CloudGaming.md` (plus `Agent_results/` mirror).
- `MVP/03_video_technology/02_Response/video-tech.agent.final/video-tech.agent.final.md` (plus `Agent_Results/` per-section files).

When the user asks you to merge / extend / cross-reference research, these are the files to dive into — not the `.docx`/`.pdf`/`.zip` siblings.

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

## What "doing a task" usually looks like here

Until code lands, the typical task is one of:
1. **Synthesize across the three MVP research streams** into a single deeper spec (the user calls this the "ultimate documentation"). Do not drop content from any of the three sources; extend with web research where it adds value.
2. **Write or refine a section of a research chapter** under `docs/research/chapters/...`. Keep the existing per-section file naming (`cloudgaming_secNN.md`, `video-tech_secNN.md`, etc.).
3. **Plan submodule structure** under `vasic-digital` / `HelixDevelopment` — every plan needs the constraints above baked in, plus the test matrix and container story.

Generic "summarize this" or "trim this" responses are explicitly not what the user wants — `04_Request.md` forbids simplification of the spec.
