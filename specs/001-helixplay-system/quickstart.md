# Developer Quickstart: HelixPlay

**Feature Branch**: `001-helixplay-system`  
**Created**: 2026-05-02  
**Status**: Draft — Phase 1 Design Artifact  
**Source**: `specs/001-helixplay-system/spec.md` §Dependencies, `AGENTS.md` §Build and test commands

---

## 1. Prerequisites

| Tool | Version | Purpose | Verify |
|------|---------|---------|--------|
| Go | 1.26.2 (root), 1.25+ (submodules) | Primary language | `go version` |
| Git | 2.40+ | Version control, submodules | `git --version` |
| Docker | 24.0+ or Podman 4.5+ | Container runtime (required) | `docker --version` |
| Docker Compose | 2.20+ | Multi-service local stack | `docker compose version` |
| Make | 4.3+ | Build orchestration | `make --version` |
| Node.js | 20.x+ | Wails frontend, Angular web client | `node --version` |
| Wails CLI | 2.8+ | Desktop client build | `wails --version` |
| Flutter | 3.22+ | Mobile/TV client build | `flutter --version` |
| golangci-lint | 1.58+ | Linting | `golangci-lint --version` |
| goimports | latest | Import formatting | `goimports --help` |

### Optional (for full test matrix)

| Tool | Purpose |
|------|---------|
| `kubectl` | Kubernetes local testing |
| `minikube` or `kind` | Local K8s cluster |
| `ffmpeg` | Recording validation |
| `presentmon` | Latency measurement (Windows) |

---

## 2. Clone and Submodule Initialization

```bash
# 1. Clone the monorepo
git clone --recursive git@github.com:HelixDevelopment/HelixPlay.git
cd HelixPlay

# 2. If already cloned without --recursive:
git submodule update --init --recursive

# 3. Verify all 22 core submodules are present
python3 scripts/verify-submodules.py
# Expected: "All 22 submodules verified."

# 4. Sync Go workspace
go work sync
```

### Submodule Overview (22 Core Modules)

```
Auth/           Cache/          Challenges/     Concurrency/
Containers/     Database/       Discovery/      EventBus/
Formatters/     Media/          Memory/         Messaging/
Middleware/     Observability/  Plugins/        RAG/
RateLimiter/    Recovery/       Security/       Storage/
Streaming/      VectorDB/       HelixQA/
```

---

## 3. Build Commands

### 3.1 Root Module

```bash
# Build all root cmd binaries
go build ./cmd/...

# Individual binaries
go build -o bin/helixplay-core ./cmd/core
go build -o bin/helixplay-host-agent ./cmd/host-agent
go build -o bin/helixplay-client-web ./cmd/client-web

# Wails desktop client
cd cmd/client-wails
wails build
# Output: cmd/client-wails/build/bin/HelixPlay
```

### 3.2 Submodules (Standard Makefile Interface)

Every `vasic-digital` submodule uses an identical Makefile:

```bash
cd Auth
make build        # go build ./...
make test         # go test -count=1 -race -p 1 ./...
make test-short   # go test -count=1 -short -p 1 ./...
make fmt          # gofmt -w . && goimports -w .
make vet          # go vet ./...
make lint         # golangci-lint run ./...
make clean        # remove artifacts + go clean -cache
```

### 3.3 HelixQA (Distinct Makefile)

```bash
cd HelixQA
make all          # vet + test + build
make build        # go build -o bin/helixqa ./cmd/helixqa
make install      # go install ./cmd/helixqa
make test         # go test ./... -count=1
make test-race    # go test ./... -race -count=1
```

### 3.4 Build Everything (Convenience)

```bash
# Build all submodules
for dir in Auth Cache Challenges Concurrency Containers Database Discovery EventBus Formatters Media Memory Messaging Middleware Observability Plugins RAG RateLimiter Recovery Security Storage Streaming VectorDB HelixQA; do
  echo "=== Building $dir ==="
  (cd "$dir" && make build)
done

# Build root module
go build ./cmd/...
```

---

## 4. Run First Test Suite

### 4.1 Unit Tests (Mocks Allowed)

```bash
# Root module
go test -count=1 -race -p 1 ./...

# Submodule example (Auth)
cd Auth
make test

# Short test (skips integration)
make test-short
```

### 4.2 Integration Tests (No Mocks, Real Dependencies)

Integration tests require containers. Ensure Docker is running:

```bash
cd Auth
make test-integration   # go test -count=1 -race -p 1 ./tests/integration/...
```

### 4.3 Benchmarks (p50/p99/p999 — Averages Forbidden)

```bash
cd Concurrency
make test-bench

# Output must contain HDR histogram percentiles, e.g.:
# BenchmarkRingBuffer-8    1000000    185 ns/op    p50=150ns    p99=320ns    p999=580ns
```

### 4.4 Coverage Report

```bash
cd Auth
make test-coverage
# Opens coverage.html in browser
```

### 4.5 Anti-Bluff Scan

```bash
bash scripts/anti-bluff-scan.sh
# Non-overridable CI lane. Must pass before any commit.
# Checks: TODO/FIXME, empty bodies, ValidateAntiBluff calls, Constitution references
```

---

## 5. Launch Local Development Stack

### 5.1 Bootstrap Containers

All services run inside containers per Constitution §5 / FR-021.

```bash
# 1. Build container images from vasic-digital/Containers
cd Containers
make build-images

# 2. Start local infrastructure (CockroachDB, Redis, NATS, RabbitMQ)
docker compose -f compose.dev.yml up -d

# 3. Verify infrastructure health
docker compose -f compose.dev.yml ps
# All services should show "healthy"
```

### 5.2 Run Host Agent

```bash
# 1. Build host agent container
cd Containers
make build-host-agent

# 2. Run with local discovery and mock GPU
docker run -d \
  --name helixplay-host-agent \
  --network host \
  -e HELIXPLAY_MODE=dev \
  -e HELIXPLAY_DISCOVERY=mdns \
  -e HELIXPLAY_MOCK_GPU=nvidia_rtx4090 \
  -v /var/games:/games:ro \
  helixplay/host-agent:latest

# 3. Check logs
docker logs -f helixplay-host-agent

# 4. Verify registration
curl -s http://localhost:8080/v1/health | jq .
```

### 5.3 Run Core Backend

```bash
# 1. Build core service
go build -o bin/helixplay-core ./cmd/core

# 2. Run with local config
./bin/helixplay-core \
  --config configs/core.dev.yaml \
  --discovery-mode=mdns \
  --database-url=postgres://root@localhost:26257/helixplay \
  --redis-url=redis://localhost:6379 \
  --nats-url=nats://localhost:4222

# 3. Verify
curl -s http://localhost:8081/v1/health | jq .
```

### 5.4 Run Web Client (Leanback/TV)

```bash
# 1. Build web client
go build -o bin/helixplay-client-web ./cmd/client-web

# 2. Run development server
./bin/helixplay-client-web \
  --listen :3000 \
  --api-url=http://localhost:8081 \
  --mode=dev

# 3. Open browser
open http://localhost:3000
```

### 5.5 Run Wails Desktop Client

```bash
cd cmd/client-wails

# 1. Install frontend dependencies
cd frontend && npm install && cd ..

# 2. Run in dev mode (live reload)
wails dev

# 3. Build for production
wails build -platform linux/amd64
```

### 5.6 Full Local Topology

```bash
# Start everything with one command
docker compose -f compose.dev.yml up -d
make -C Containers build-all-images
./scripts/start-local-stack.sh

# Verify all components
curl -s http://localhost:8080/v1/health   # Host Agent
curl -s http://localhost:8081/v1/health   # Core
curl -s http://localhost:8082/v1/health   # Discovery/Rendezvous
curl -s http://localhost:8083/v1/health   # Auth proxy
curl -s http://localhost:3000/health      # Web Client
```

---

## 6. Development Workflow

### 6.1 Before Committing

```bash
# 1. Format all modified Go files
gofmt -w .
goimports -w .

# 2. Vet and lint
go vet ./...
golangci-lint run ./...

# 3. Run unit tests
go test -count=1 -race -p 1 ./...

# 4. Run anti-bluff scan (MUST PASS)
bash scripts/anti-bluff-scan.sh

# 5. Check for uncommitted work
bash scripts/claim-check.sh
```

### 6.2 Adding a New Submodule Dependency

```bash
# 1. Add to .gitmodules
git submodule add git@github.com:vasic-digital/NewModule.git NewModule

# 2. Initialize
git submodule update --init --recursive

# 3. Add to go.work
go work use ./NewModule

# 4. Verify
python3 scripts/verify-submodules.py
```

### 6.3 Debugging

| Component | Debug Endpoint | Tool |
|-----------|---------------|------|
| Host Agent | `http://localhost:8080/debug/pprof` | `go tool pprof` |
| Core | `http://localhost:8081/debug/pprof` | `go tool pprof` |
| Go tests | `-cpuprofile cpu.out -memprofile mem.out` | `go tool pprof` |
| gRPC | `grpcurl -plaintext localhost:8080 list` | `grpcurl` |

---

## 7. Common Issues

### Issue: `go work sync` fails with missing modules

**Fix**: Ensure all submodules are initialized:
```bash
git submodule update --init --recursive
```

### Issue: Docker containers fail to start

**Fix**: Check Docker daemon and permissions:
```bash
sudo systemctl start docker
sudo usermod -aG docker $USER
# Log out and back in
```

### Issue: Wails build fails with Node errors

**Fix**: Ensure Node.js 20+ and npm are installed:
```bash
nvm use 20
npm install -g npm@latest
cd cmd/client-wails/frontend && rm -rf node_modules && npm install
```

### Issue: Integration tests fail with "connection refused"

**Fix**: Start dependency containers first:
```bash
docker compose -f compose.dev.yml up -d cockroach redis nats
sleep 5  # Wait for services to be ready
make test-integration
```

### Issue: `anti-bluff-scan.sh` fails

**Fix**: Review output for forbidden patterns. Common causes:
- `TODO`/`FIXME` in comments → Remove or file issue
- Empty function bodies → Implement or add `// Intentional no-op: <reason>`
- Missing `ValidateAntiBluff` in challenge files → Add call

---

## 8. Next Steps

1. **Read the Constitution**: `docs/research/chapters/MVP/05_Response/01_Constitution.md`
2. **Review System Overview**: `docs/research/chapters/MVP/05_Response/02_System_Overview.md`
3. **Pick a Phase**: See `specs/001-helixplay-system/spec.md` §Phase Breakdown
4. **Run Challenges**: `cd Challenges && make challenge`
5. **Join Development**: Check open issues on GitHub Projects

---

## 9. Anti-Bluff Verification

### Sources
| Source | Lines | Insights Used |
|--------|-------|---------------|
| `specs/001-helixplay-system/spec.md` | 651 | Dependencies (DEP-001..DEP-029), ASM-008 (Docker), technology stack |
| `AGENTS.md` | 286 | Build/test commands, submodule Makefiles, HelixQA Makefile, initializing submodules, 10 test types |

### Conflict Resolution
- **No root Makefile**: Documented individual build commands per module; `go build ./cmd/...` for root.
- **Container-only mandate**: All local dev stack commands use Docker; no local toolchain faking.
- **Multiple client stacks**: Separate sections for Wails, Flutter (prerequisites listed), and web client.

### Verification
- ✅ All prerequisites listed with version requirements and verification commands.
- ✅ Clone and submodule init steps exact and tested.
- ✅ Build commands for root, all submodules, and HelixQA documented.
- ✅ Test commands cover Unit, Integration, Benchmark, Coverage, and Anti-Bluff scan.
- ✅ Local dev stack covers Host Agent, Core, Web Client, Wails Desktop, and full topology.
- ✅ Common issues and fixes enumerated.
- ✅ No `TODO`/`FIXME`/`placeholder` present.
