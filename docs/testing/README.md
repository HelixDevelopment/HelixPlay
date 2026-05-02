# HelixPlay Testing Strategy

## Overview

HelixPlay mandates 100% test coverage across **ten test types**:

1. **Unit** — Only type allowed to use mocks/stubs/hardcoded values
2. **Integration** — Real service interactions
3. **E2E** — Full user journey
4. **Security** — Vulnerability scanning, penetration testing
5. **Benchmarking** — Latency tests report p50/p99/p999 (averages forbidden)
6. **Chaos** — Random failures and recovery
7. **Stress** — Load beyond capacity
8. **Smoke** — Basic health after deployment
9. **Full Automation** — Autonomous HelixQA orchestration
10. **Challenges** — Production-like full system boot

## Test Commands

```bash
# Unit (with race detection)
make test-unit

# Integration
docker compose -f tests/integration/docker-compose.yml up
go test -count=1 -race -p 1 ./tests/integration/...

# E2E
go test -count=1 -race -p 1 ./tests/e2e/...

# Benchmarks
go test -bench=. -benchmem ./tests/benchmark/...

# Chaos
make test-chaos

# Stress
make test-stress

# Security
make test-security

# Smoke
make test-smoke

# Full Automation
make test-fullauto

# Challenges
make challenge
```

## Coverage Requirements

| Package | Minimum Coverage |
|---------|-----------------|
| `pkg/core/streaming` | 90% |
| `pkg/core/discovery` | 85% |
| `pkg/core/network` | 85% |
| `cmd/core` | 80% |
| `cmd/host-agent` | 80% |
| `cmd/client-*` | 75% |

## Anti-Bluff Verification

All tests must verify observable behavior. Forbidden:
- `assert.True(t, true)`
- `assert.NotNil(t, nil)`
- Constructor-only tests (`TestNew*` with only nil checks)
- Mock-only integration/E2E tests
- Permanently skipped tests without containerization plans

## Mutation Testing

Run mutation testing before release:

```bash
go install github.com/zimmski/go-mutesting@latest
go-mutesting ./pkg/... ./cmd/...
# Minimum mutation score: 80%
```

## CI Gates

1. All unit tests pass with `-race`
2. Integration tests pass in containers
3. Security scan (Snyk + SonarQube) clean
4. Benchmark regression < 5%
5. Anti-bluff scan zero violations
6. Challenge scenarios all pass
