# HelixPlay Developer Onboarding

## Prerequisites

- Go 1.26.2+
- Protocol Buffers compiler (protoc) 25.5+
- protoc-gen-go and protoc-gen-go-grpc plugins
- Docker or Podman (for integration tests)
- Git with submodule support

## Repository Setup

```bash
# Clone with all submodules
git clone --recurse-submodules git@github.com:HelixDevelopment/HelixPlay.git
cd HelixPlay

# Verify submodule integrity
make verify-submodules

# Generate protobuf code
make proto

# Build all binaries
make build
```

## IDE Configuration

### VS Code

Install extensions:
- Go (golang.go)
- Protocol Buffer (zxh404.vscode-proto3)

Settings (`settings.json`):
```json
{
  "gopls": {
    "build.experimentalWorkspaceModule": true
  },
  "go.toolsManagement.autoUpdate": false
}
```

### GoLand / IntelliJ

1. Open root directory as project
2. Enable Go modules integration
3. Set GOROOT to Go 1.26.2
4. Mark `pkg/` and `cmd/` as source roots

## Running Tests

```bash
# Fast unit tests
make test-unit

# Full test suite (takes ~10 minutes)
make test

# Specific package
go test -race -v ./pkg/core/streaming/...

# With coverage
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

## Code Style

- Format: `gofmt` + `goimports` (run via `make fmt`)
- Lint: `golangci-lint` (run via `make lint`)
- Vet: `go vet ./...` (run via `make vet`)

## Submitting Changes

1. Create feature branch: `git checkout -b feature/description`
2. Write tests first (TDD preferred)
3. Ensure all tests pass: `make test`
4. Run anti-bluff scan: `make anti-bluff`
5. Commit with descriptive message
6. Push to GitFlic (origin): `git push origin feature/description`

## Troubleshooting

### "found packages main and core"

Caused by Go files with `package main` in non-cmd directories. Check:
- No `main()` functions in `pkg/` or `internal/`
- `cmd/core/` only has `main.go` with `package main`

### Submodule conflicts

```bash
# Re-sync submodules
git submodule update --init --recursive
go work sync
```

### Protobuf generation fails

```bash
# Reinstall plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
