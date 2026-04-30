# AGENTS.md

For non-Claude agents (Codex, Cursor, Aider, Copilot, Cline, etc.) working in this repo. Claude Code reads `CLAUDE.md`.

> **Source of truth:** `docs/research/chapters/MVP/05_Response/01_Constitution.md`
> **Authoritative MVP brief:** `docs/research/chapters/MVP/04_Request.md` — read before any planning question.

---

## Repo state

**Specs and research only — no source code, no build system.** Treat all tasks as documentation/specification unless the operator explicitly says otherwise.

Key paths:
- `docs/research/chapters/MVP/` — three research streams (base, latency, video)
- `04_Request.md` — supersedes per-stream requests on conflict
- `05_Response/` — canonical synthesized documentation (your output target)
- `Upstreams/*.sh` — export `UPSTREAMABLE_REPOSITORY` for external tooling; not part of any build

Per-stream request files drop the `01_` prefix inconsistently (`01_base/01_Request.md` vs `03_video_technology/Request.md`). Response folders drift in casing (`02_response/` vs `02_Response/`) — preserve as-is.

---

## Git topology

Four remotes; `origin` is **split**: fetch from GitHub, push to GitFlic.

```
github      git@github.com:HelixDevelopment/HelixPlay.git
gitlab      git@gitlab.com:helixdevelopment1/HelixPlay.git
gitverse    git@gitverse.ru:helixdevelopment/HelixPlay.git
gitflic     git@gitflic.ru:helixdevelopment/helixplay.git
origin      fetch=github, push=gitflic
```

When operator says "push", confirm which mirror — `origin` only updates GitFlic. Force-push requires explicit authorization. `--no-verify` is forbidden.

---

## Critical constraints

These are mandatory project-wide rules, not suggestions:

- **Anti-bluff:** No `TODO`, `FIXME`, `XXX`, `placeholder`, empty function bodies, dead code, or tests that pass without exercising real behavior. Details in Constitution §1.
- **Containers only:** Every service, DB, build step, test runner, and scanner runs inside a container. Definitions live in `vasic-digital/Containers` — never vendor a `Dockerfile` outside that submodule. No faking a local toolchain.
- **Decoupling:** Reusable components live in **public** `vasic-digital` Git/Go submodules. Reuse before recreating.
- **Tests:** 100% coverage across **all ten** types: Unit, Integration, E2E, Security, Benchmarking, Chaos, Stress, Smoke, Full Automation, **Challenges**. Only Unit may use mocks. Latency tests report p50/p99/p999 — no averages. Challenges boot the full stack from `vasic-digital/Challenges`; QA lives in `HelixDevelopment/HelixQA`.
- **R-18 Operational Integrity:** No command, hook, container entrypoint, or prompt may suspend, hibernate, lock, terminate, or crash the operator's host. Forbidden-commands list in Constitution §11.5 — read before generating any Bash command.

---

## Agent harness configuration

- `.claude/settings.json` registers a `Stop` hook running `bash scripts/claim-check.sh` (5s timeout). **`scripts/` does not exist** — the hook fails until created.
- `.claude/settings.local.json` sets `defaultMode: bypassPermissions`. Tool calls won't prompt; safety bar is judgment, not the permission system. Be careful with destructive git operations and the four remotes.

---

## Documentation work

Output goes under `05_Response/`. Every file must end with an `Anti-Bluff Verification` block (template in `05_Response/00_Master_Plan.md` §4.3) listing sources, URLs, insights, conflict resolutions, and line counts. Skipping this block is a Constitution §1 violation and a merge blocker.

Synthesis methodology: `05_Response/00_Master_Plan.md` §4. Do not simplify specs — `04_Request.md` forbids summarization.

---

End of `AGENTS.md`. Last updated 2026-04-30.
