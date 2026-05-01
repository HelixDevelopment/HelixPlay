# AGENTS.md

> **NON-NEGOTIABLE PRIME DIRECTIVE:**
> **"We had been in position that all tests do execute with success and
> all Challenges as well, but in reality the most of the features does
> not work and can't be used! This MUST NOT be the case and execution
> of tests and Challenges MUST guarantee the quality, the completion
> and full usability by end users of the product!"**
> This statement is the foundational requirement of this project. Any
> agent dispatch, any CI configuration, any code review that allows
> green tests on broken features is a violation and MUST be rejected.

For non-Claude agents (Codex, Cursor, Aider, Copilot, Cline, etc.) working in this repo. Claude Code reads `CLAUDE.md`.

> **Source of truth:** `docs/research/chapters/MVP/05_Response/01_Constitution.md` **v2.1.0**
> **Authoritative MVP brief:** `docs/research/chapters/MVP/04_Request.md` — read before any planning question.
> **Synthesis programme master plan:** `docs/research/chapters/MVP/05_Response/00_Master_Plan.md`
> **System overview:** `docs/research/chapters/MVP/05_Response/02_System_Overview.md`
>
> **Constitution v2.1.0 amendments (2026-05-01):**
> 1. Anti-bluff tests MUST guarantee real, end-user-usable behaviour. Forbidden:
>    `assert.True(t, true)`, `assert.NotNil(t, nil)`, constructor-only tests,
>    mock-only integration/E2E tests, and permanently skipped tests without
>    containerization plans.
> 2. **Usability evidence mandatory** per §6.7 — every feature needs HelixQA visual
>    assertion, manual recording, or Challenge scenario evidence.
> 3. **Automatic negative-leg fault injection** per §1.3 / §6.3 / §11.5.7 — CI
>    breaks each feature and verifies non-Unit tests fail. If none fail, merge
>    is blocked.
> 4. `ValidateAntiBluff` unconditional; `CHALLENGE_ANTIBLUFF_STRICT` removed.
>    All challenge implementations now call `RecordAction()`.
> 5. Container verifier `execCommand()` executes real commands; no more no-op.

---

## Project overview

HelixPlay is a cloud-gaming platform built as a Go-centric monorepo with Git submodules for reusable infrastructure. The project slogan is "Ultimate gaming experience!"

The repo has two layers:

1. **Root module** (`github.com/HelixDevelopment/HelixPlay`) — Go 1.26.2 — contains:
   - `cmd/client-wails/` — Wails desktop client (Go backend + frontend placeholder).
   - `cmd/client-web/` — Web client (leanback/TV experience).
   - `cmd/core/` — Core backend stubs (discovery, protocol).
   - `cmd/host-agent/` — Sunshine-style host agent with capture (Windows DX11, Darwin ScreenCaptureKit, Linux PipeWire), encoding (hardware, dual-path), codec negotiation, input (USB, DualSense, hotplug), and transport (UDP, QUIC, WebRTC).

2. **22 Git submodules** — independent Go modules vendored under the root. Most are `vasic-digital/*` reusable libraries; `HelixQA` is the autonomous QA framework under `HelixDevelopment`.

Key paths:
- `docs/research/chapters/MVP/` — three research streams (base, latency, video) plus canonical synthesized docs under `05_Response/`.
- `04_Request.md` — supersedes per-stream requests on conflict.
- `05_Response/` — canonical synthesized documentation (agent output target). Every `.md` file here must end with an `Anti-Bluff Verification` block (template in `05_Response/00_Master_Plan.md` §4.3).
- `specs/001-helixplay-system/spec.md` — feature specification.
- `scripts/` — project-wide automation (see below).
- `Upstreams/*.sh` — export `UPSTREAMABLE_REPOSITORY` for external tooling; not part of any build.

Per-stream request files drop the `01_` prefix inconsistently (`01_base/01_Request.md` vs `03_video_technology/Request.md`). Response folders drift in casing (`02_response/` vs `02_Response/`) — preserve as-is.

---

## Technology stack

- **Language:** Go (root: 1.26.2; submodules: 1.25+).
- **Desktop client:** Wails (Go backend, frontend package in `cmd/client-wails/frontend/package.json`).
- **Web client:** Plain Go HTTP handlers for TV/leanback mode.
- **Networking:** gRPC preferred; REST + middleware as separate microservices; HTTP/3 (QUIC/Cronet); WebRTC; UDP with custom policies; NATS for messaging.
- **Databases:** CockroachDB (primary), Redis, RabbitMQ where useful.
- **Compression:** Brotli (`github.com/andybalholm/brotli`).
- **Container runtimes:** Docker, Podman, Kubernetes — definitions live exclusively in `vasic-digital/Containers`.
- **Observability:** Prometheus client libraries in some submodules.
- **Auth:** JWT ( `golang-jwt/jwt/v5` ), OAuth/OAuth2 providers, API keys, access tokens — via `vasic-digital/Auth`.
- **QA:** Custom `helixqa` binary built from `HelixQA` submodule; uses OpenCV for visual verification.

---

## Code organization

### Root module layout
```
cmd/
  client-wails/   # Wails desktop entry point + backend package
  client-web/     # Web/TV leanback entry point
  core/           # Backend discovery & protocol stubs
  host-agent/     # Capture, encode, codec, input, transport, lifecycle, capability, game enumeration
```

### Submodule layout (consistent pattern)
Each submodule follows standard Go conventions:
```
<Submodule>/
  cmd/            # Executable entry points (if any)
  pkg/            # Public API / reusable packages
  internal/       # Private implementation
  docs/           # API reference, architecture diagrams (Mermaid)
  challenges/     # Challenge test scripts and Go challenge tests
  tests/          # Unit, integration, benchmark subdirectories
  Makefile        # Standard build/test/lint targets
  go.mod          # Independent Go module
  AGENTS.md       # Submodule-specific agent guidance
  CLAUDE.md       # Submodule-specific Claude guidance
  CONSTITUTION.md # Reference to HelixPlay Constitution
  ARCHITECTURE.md # High-level design
```

### Active submodules (from `.gitmodules`)
| Path | Organization | Purpose |
|------|--------------|---------|
| `Auth` | vasic-digital | JWT, OAuth, API key, middleware auth |
| `Cache` | vasic-digital | Caching abstractions |
| `Catalogizer` | vasic-digital | Asset catalogization |
| `Challenges` | vasic-digital | Full-stack challenge runner & bluff scanner |
| `Concurrency` | vasic-digital | Non-blocking concurrency primitives |
| `Containers` | vasic-digital | Container orchestration, health checks, lifecycle, service discovery |
| `Database` | vasic-digital | DB abstractions |
| `Discovery` | vasic-digital | LAN service discovery, dynamic ports |
| `EventBus` | vasic-digital | Event propagation |
| `Formatters` | vasic-digital | Output formatting |
| `Media` | vasic-digital | Media processing |
| `Memory` | vasic-digital | Memory management utilities |
| `Messaging` | vasic-digital | Message queue abstractions |
| `Middleware` | vasic-digital | HTTP/gRPC middleware |
| `Observability` | vasic-digital | Metrics, logging, tracing |
| `Plugins` | vasic-digital | Plugin system |
| `RAG` | vasic-digital | Retrieval-Augmented Generation support |
| `RateLimiter` | vasic-digital | Rate limiting |
| `Recovery` | vasic-digital | Fault recovery |
| `Security` | vasic-digital | Security utilities |
| `Storage` | vasic-digital | Storage abstractions |
| `Streaming` | vasic-digital | Streaming protocols |
| `VectorDB` | vasic-digital | Vector database support |
| `HelixQA` | HelixDevelopment | Autonomous QA orchestration framework |

---

## Build and test commands

### Root module
There is no root-level `Makefile` yet. Build submodules individually or use `go build ./cmd/...` from the root after ensuring submodules are initialized.

### Submodule Makefiles (standardized pattern)
Every `vasic-digital` submodule carries an identical `Makefile` interface:

```bash
make build              # go build ./...
make test               # go test -count=1 -race -p 1 ./...
make test-race          # same as test
make test-short         # go test -count=1 -short -p 1 ./...
make test-integration   # go test -count=1 -race -p 1 ./tests/integration/...
make test-bench         # go test -bench=. -benchmem ./tests/benchmark/...
make test-coverage      # coverage.out + coverage.html
make fmt                # gofmt -w . && goimports -w .
make vet                # go vet ./...
make lint               # golangci-lint run ./...
make clean              # remove coverage artifacts + go clean -cache
make challenge          # run challenges/scripts/<module>_challenge.sh (from parent project)
```

Environment variable: `GOMAXPROCS ?= 2` is set by default in submodule tests.

### HelixQA (distinct Makefile)
```bash
make all        # vet + test + build
make build      # go build -o bin/helixqa ./cmd/helixqa
make install    # go install ./cmd/helixqa
make test       # go test ./... -count=1
make test-race  # go test ./... -race -count=1
make test-cover # coverage report
make vet        # go vet ./...
make lint       # golangci-lint run ./...
make fmt        # gofmt -w .
make tidy       # go mod tidy
make clean      # rm -rf bin/ coverage.out coverage.html qa-results/ results/ evidence/
```

### Initializing submodules
```bash
git submodule update --init --recursive
```

---

## Testing strategy

The project mandates **100% coverage across all ten test types**:

1. **Unit** — only type allowed to use mocks/stubs/hardcoded values.
2. **Integration**
3. **E2E**
4. **Security**
5. **Benchmarking** — latency tests must report **p50/p99/p999**; averages are forbidden.
6. **Chaos**
7. **Stress**
8. **Smoke**
9. **Full Automation**
10. **Challenges** — production-like, full system booted from `vasic-digital/Challenges`.

Additional rules:
- Every non-Unit test must exercise the **real** system (no mocks).
- Autonomous QA lives in `HelixDevelopment/HelixQA`.
- Anti-bluff tests must guarantee real end-user usability — green tests without working features are a Constitution §1 violation.

---

## Code style guidelines

- **Formatting:** `gofmt` + `goimports` (run via `make fmt`).
- **Linting:** `golangci-lint` (run via `make lint`).
- **Vetting:** `go vet ./...` (run via `make vet`).
- **DRY, KISS, and Top 10 principles** are mandatory (User Mandate 2026-04-30).
- **Lazy initialization** is the default (Constitution §5.2).
- **Concurrency:** non-blocking by default; use semaphores/backpressure to prevent clogging.
- **No local toolchain faking:** every service, DB, build step, test runner, and scanner runs inside a container. Do not vendor `Dockerfile`s outside `vasic-digital/Containers`.

---

## Security considerations

- **R-18 Operational Integrity:** No command, hook, container entrypoint, or prompt may suspend, hibernate, lock, terminate, or crash the operator's host. See Constitution §11.5 for the forbidden-commands list and container hazards inventory.
- **Quality gates:** SonarQube, Snyk, plus other heavy security/quality scans are required.
- **Anti-bluff enforcement:** `scripts/anti-bluff-scan.sh` is a non-overridable CI lane that scans for forbidden patterns (TODO, FIXME, empty bodies, etc.) and verifies Constitution references in submodule configs.
- Force-push requires explicit authorization. `--no-verify` is forbidden.

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

When operator says "push", confirm which mirror — `origin` only updates GitFlic.

---

## Project scripts

Located in `scripts/`:

- **`anti-bluff-scan.sh`** — CI lane. Scans code for forbidden patterns (TODO, FIXME, empty function bodies, `panic("not implemented")`, etc.), verifies `ValidateAntiBluff` is called in the Challenges runner, checks for Anti-Bluff Verification blocks in `05_Response/`, and verifies Constitution references in submodule configs. Non-overridable; exits 1 on failure.
- **`claim-check.sh`** — Hook invoked by Claude Code on stop (5s timeout). Checks for uncommitted work or claims.
- **`propagate-constitution.sh`** — Propagates Constitution v2.0.0 preamble to all submodules by rewriting their `CLAUDE.md`, `AGENTS.md`, and `CONSTITUTION.md` files.
- **`verify-submodules.py`** — Validates that `.gitmodules` contains all 22 required submodules with correct paths and URLs.

---

## Critical constraints

These are mandatory project-wide rules, not suggestions:

- **Anti-bluff:** No `TODO`, `FIXME`, `XXX`, `placeholder`, empty function bodies, dead code, or tests that pass without exercising real behaviour. Details in Constitution §1. Explicitly forbidden: `assert.True(t, true)`, `assert.NotNil(t, nil)`, constructor-only tests (`TestNew*` with only nil checks), mock-only integration/E2E tests, and permanently skipped tests without containerization plans. Tests MUST confirm that all tested codebase really works as expected and can be used by end users. Usability evidence (HelixQA visual assertion, manual recording, or Challenge scenario) is mandatory per §6.7.
- **Containers only:** Every service, DB, build step, test runner, and scanner runs inside a container. Definitions live in `vasic-digital/Containers` — never vendor a `Dockerfile` outside that submodule. No faking a local toolchain.
- **Decoupling:** Reusable components live in **public** `vasic-digital` Git/Go submodules. Reuse before recreating.
- **Tests:** 100% coverage across **all ten** types: Unit, Integration, E2E, Security, Benchmarking, Chaos, Stress, Smoke, Full Automation, **Challenges**. Only Unit may use mocks. Latency tests report p50/p99/p999 — no averages. Challenges boot the full stack from `vasic-digital/Challenges`; QA lives in `HelixDevelopment/HelixQA`.
- **R-18 Operational Integrity:** No command, hook, container entrypoint, or prompt may suspend, hibernate, lock, terminate, or crash the operator's host. Forbidden-commands list in Constitution §11.5 — read before generating any Bash command.

---

## Agent harness configuration

- `.claude/settings.json` registers a `Stop` hook running `bash scripts/claim-check.sh` (5s timeout).
- `.claude/settings.local.json` sets `defaultMode: bypassPermissions`. Tool calls won't prompt; safety bar is judgment, not the permission system. Be careful with destructive git operations and the four remotes.

---

## Documentation work

Output goes under `05_Response/`. Every file must end with an `Anti-Bluff Verification` block (template in `05_Response/00_Master_Plan.md` §4.3) listing sources, URLs, insights, conflict resolutions, and line counts. Skipping this block is a Constitution §1 violation and a merge blocker.

Synthesis methodology: `05_Response/00_Master_Plan.md` §4. Do not simplify specs — `04_Request.md` forbids summarization.

---

End of `AGENTS.md`. Last updated 2026-05-01.
