# Comprehensive Testing & Anti-Bluff Strategy

## HelixPlay Cloud Gaming Platform — Implementation Plan Section

**Document Version**: 1.0  
**Constitution Reference**: v2.1.0  
**Date**: 2026-05-02  
**Classification**: Mandatory — Non-Overrideable CI Lane  

---

## Table of Contents

- [Section 1: Testing Philosophy & Anti-Bluff Constitution](#section-1-testing-philosophy--anti-bluff-constitution)
- [Section 2: The Ten Test Types — Detailed Implementation](#section-2-the-ten-test-types--detailed-implementation)
- [Section 3: Test Matrix (29 Submodules x 10 Types)](#section-3-test-matrix-29-submodules--10-types)
- [Section 4: Anti-Bluff Infrastructure Implementation](#section-4-anti-bluff-infrastructure-implementation)
- [Section 5: HelixQA Autonomous QA Integration](#section-5-helixqa-autonomous-qa-integration)
- [Section 6: CI/CD Pipeline Design](#section-6-cicd-pipeline-design)
- [Section 7: Fixing Current Issues](#section-7-fixing-current-issues)
- [Section 8: Challenges Implementation Plan](#section-8-challenges-implementation-plan)

---

## Section 1: Testing Philosophy & Anti-Bluff Constitution

### 1.1 The Prime Directive

**Constitutional Text (Constitution v2.1.0, Clause 1 — Anti-Bluff Pledge)**:

> *"Green tests MUST guarantee real, end-user-usable behavior. The project has suffered from 'green tests on broken features' — tests that pass while the actual feature does not work. This MUST NEVER happen again. Every test that passes MUST correspond to observable, verifiable functionality that a real user can interact with."*

**Historical Context**: HelixPlay previously had 267+ vacuous tests — tests that passed but verified nothing meaningful. These included `assert.True(t, true)`, constructor-only tests that verified an object was created but never exercised its behavior, and mock-only integration tests where mocked dependencies returned hardcoded values. The result: all tests showed green while core streaming, input, and discovery features were non-functional.

**What This Means in Practice**:

1. **No test may assert a tautology** — `assert.True(t, true)`, `assert.Equal(t, 1, 1)`, `assert.Nil(t, nil)` are forbidden by constitutional law. Any such assertion detected in CI causes an immediate pipeline failure.

2. **Constructor-only tests are bluff tests** — A test that only verifies `require.NotNil(t, obj)` after calling `NewFoo()` proves the allocator works, not the feature. Every test must exercise at least one observable behavior of the constructed object.

3. **Mock-only integration tests are bluff tests** — Integration tests (Types 2–10) must use real dependencies. If an integration test mocks the database, the HTTP client, and the auth layer, it tests the mock wiring, not the integration.

4. **Passing tests must correlate to working features** — If a feature is broken for a real user, at least one test must fail. If all tests pass but the feature doesn't work, the test suite is bluffing and must be rebuilt.

5. **Usability evidence is mandatory** — Per Constitution §6.7, every feature must provide evidence of real usability: HelixQA visual assertion, manual screen recording, or Challenge scenario execution with anti-bluff validation.

### 1.2 Forbidden Patterns (Complete List)

#### 1.2.1 Forbidden Code Patterns

| Pattern | Severity | Detection Method | Example |
|---------|----------|-----------------|---------|
| Empty function body `{}` | **CRITICAL** | Static AST scan | `func (s *Server) Handle() {}` |
| `panic("not implemented")` stub | **CRITICAL** | Regex scan | `panic("not implemented")` |
| `return nil, fmt.Errorf("not implemented")` | **CRITICAL** | Regex scan | `return nil, errors.New("not implemented")` |
| `TODO` comment without issue reference | **WARNING** | Regex scan | `// TODO: fix this` |
| `FIXME` comment without issue reference | **WARNING** | Regex scan | `// FIXME: broken` |
| `XXX` or `HACK` comments | **WARNING** | Regex scan | `// HACK: workaround` |
| Standalone `tbd` or `TBD` | **WARNING** | Word-boundary scan | `status = tbd` |

**Enforcement**: `scripts/anti-bluff-scan.sh` Step 1 performs a full-source-tree regex scan for all patterns above. Matches are collected into a report. Any CRITICAL match fails the CI pipeline. WARNING matches are logged and tracked but do not fail the pipeline unless the count increases relative to the established baseline.

#### 1.2.2 Forbidden Test Patterns

| Pattern | Severity | Detection Method | Why It Is a Bluff |
|---------|----------|-----------------|-------------------|
| `assert.True(t, true)` | **BLOCKER** | Regex + AST | Asserts a tautology; always passes, proves nothing |
| `assert.Equal(t, true, true)` | **BLOCKER** | Regex + AST | Same as above with different syntax |
| `assert.Nil(t, nil)` | **BLOCKER** | Regex + AST | Asserts nil is nil; always passes |
| `assert.Equal(t, 1, 1)` | **BLOCKER** | Regex + AST | Literal compared to itself |
| Constructor-only `require.NotNil(t, obj)` | **CRITICAL** | Heuristic: test body has only construction + NotNil | Verifies allocation, not behavior |
| Mock-only integration test | **CRITICAL** | Heuristic: integration/e2e test directory uses mocks | Tests mock wiring, not real integration |
| No negative-leg test | **CRITICAL** | Challenge runner validation | Feature works for happy path but no test verifies failure detection |
| Empty test body `func TestX(t *testing.T) {}` | **BLOCKER** | AST scan | Test that passes without executing any code |
| Test with no assertions | **CRITICAL** | AST scan: count assert/require calls | Test executes code but never checks results |
| `assert.NoError(t, nil)` | **BLOCKER** | Regex scan | Asserts nil error is nil; tautology |

**Enforcement**: `scripts/anti-bluff-scan.sh` Step 5 performs the vacuous assertion scan. The bluff scanner self-test (`bluff_scanner_challenge.sh` Phase 1) verifies the scanner can detect hand-crafted bluff fixtures. If the scanner cannot find known bluff patterns, the pipeline fails — the test for bluff detection must itself not be a bluff.

#### 1.2.3 Forbidden Documentation Patterns

| Pattern | Severity | Detection Method |
|---------|----------|-----------------|
| Missing Anti-Bluff Verification section | **CRITICAL** | Documentation structure scan |
| Anti-Bluff section without actionable assertions | **WARNING** | Content heuristic |
| Placeholder text in verification section | **CRITICAL** | Regex: "TBD", "placeholder", "to be defined" |

**Enforcement**: `scripts/anti-bluff-scan.sh` Step 4 checks that all submodule documentation contains Anti-Bluff Verification sections. Missing sections are flagged and tracked.

#### 1.2.4 Enforcement Mechanisms

The anti-bluff enforcement operates at four levels:

1. **Local Prevention**: Claude Code `.claude/settings.json` Stop hook runs `claim-check.sh` before each commit, blocking R-18 forbidden commands.

2. **Pre-Commit Scan**: Developers can (and should) run `make anti-bluff` locally before pushing. This runs the scanner, anchor manifest validator, and mutation ratchet.

3. **CI Non-Overrideable Lane**: The `anti-bluff.yml` workflow (see Section 6.2) runs on every PR and cannot be bypassed. A failure here blocks merge regardless of other passing checks.

4. **Challenge Runner Gate**: The `ValidateAntiBluff()` function is called unconditionally (since v2.1.0, the `CHALLENGE_ANTIBLUFF_STRICT` toggle was removed) on every challenge result. Any challenge that reports Pass without evidence fails the challenge pipeline.

### 1.3 What is Required for Every Feature

#### 1.3.1 Observable Behavior Assertions

Every feature must have tests that verify **observable, externally-visible behavior**, not internal state. An observable behavior assertion checks something a user or external system could notice.

**Observable behaviors include**:
- An HTTP endpoint returns a specific status code and response body
- A database query returns the expected rows
- A file is created with expected content
- A gRPC stream delivers frames in the expected order
- A JWT token validates with the correct claims
- A game appears in the discovery list after enumeration
- A screen capture frame has the expected dimensions and pixel format

**Non-observable (internal state) assertions that are insufficient alone**:
- A private field was set to a specific value
- An internal counter incremented
- A mock was called with expected arguments

**Requirement**: At least 60% of assertions in any test file must verify observable behavior. The remaining 40% may verify internal state for diagnostic purposes.

#### 1.3.2 Negative-Leg Testing

For every feature, there must be a test that **deliberately breaks the feature and verifies the test suite detects the breakage**.

**Constitutional Text (§6.3)**:
> *"Automatic negative-leg fault injection: CI must break each feature and verify non-Unit tests fail."*

**Implementation**: The negative-leg fault injection system (detailed in Section 4.5) operates as follows:

1. For each feature under test, the CI pipeline creates a **mutated version** of the source code that introduces a deliberate defect (e.g., swaps `<` for `>`, removes a bounds check, changes a constant).

2. The full test suite (excluding Unit tests, which may test internals) runs against the mutated code.

3. **At least one test MUST fail**. If all tests pass on the mutated code, the test suite is bluffing — it doesn't actually verify the feature's behavior.

4. The mutation is reverted and the next feature is tested.

**Coverage Target**: 100% of production code features must have negative-leg verification.

#### 1.3.3 Usability Evidence Requirements

Per Constitution §6.7, every feature must provide one of three forms of usability evidence:

| Evidence Type | Required For | Method | Verification |
|---------------|-------------|--------|-------------|
| **HelixQA Visual Assertion** | UI-facing features | Automated screenshot/screen recording capture with OpenCV-based verification | HelixQA pipeline produces pass/fail with visual diff |
| **Manual Screen Recording** | Complex interactive features | Human records screen session following test script | Recording reviewed and attached to feature sign-off |
| **Challenge Scenario Execution** | All backend features | Challenges repository scenario runs with anti-bluff validation | `ValidateAntiBluff()` passes with `RecordedActions` and non-empty assertions |

#### 1.3.4 Coverage Requirements

| Metric | Target | Measurement Method | Enforcement |
|--------|--------|-------------------|-------------|
| **Line Coverage** | 100% | Union of all 10 test types | SonarQube gate fails if < 100% |
| **Branch Coverage** | 100% | Union of all 10 test types | SonarQube gate fails if < 100% |
| **Function Coverage** | 100% | Union of all 10 test types | SonarQube gate fails if < 100% |
| **Mutation Score** | >= 85% | go-mutesting | `mutation_ratchet_challenge.sh` fails if < 85% |
| **Anti-Bluff Pass Rate** | 100% | `ValidateAntiBluff()` | Unconditional; any bluff result fails pipeline |

**Important**: Coverage is measured across the **union** of all 10 test types, not per type. A line covered by an E2E test counts even if no Unit test covers it. However, lines covered only by Unit tests (especially mock-based Unit tests) are flagged for additional integration/E2E coverage.

---

## Section 2: The Ten Test Types — Detailed Implementation

### Overview

The Constitution §6 mandates exactly 10 test types. Each type has a specific purpose, specific mock rules, and specific CI integration. No feature is considered complete until all applicable test types are implemented and passing.

| # | Test Type | Mocks Permitted | Scope | Required Coverage |
|---|-----------|----------------|-------|-------------------|
| 1 | Unit | Yes (with restrictions) | Single function/method | 100% branches of isolated logic |
| 2 | Integration | **NO** | Cross-component real dependencies | All interaction paths |
| 3 | E2E | **NO** | Full system path, production-like topology | All user journeys |
| 4 | Security | **NO** | Fuzzing, SAST, DAST | All attack surfaces |
| 5 | Benchmarking | **NO** | p50/p99/p999 latency, throughput | All performance-critical paths |
| 6 | Chaos | Limited (chaos injection only) | Fault injection in production-like env | All failure modes |
| 7 | Stress | **NO** | Load beyond capacity, 24h profiles | All resource-limited paths |
| 8 | Smoke | **NO** | Post-deploy 30-second sanity | All critical endpoints |
| 9 | Full Automation | **NO** | Entire pipeline unattended | All of the above combined |
| 10 | Challenges | **NO** | Production-equivalent scenarios | All challenge bank scenarios |

### 2.1 Unit Tests

**Purpose**: Verify isolated logic of a single function, method, or struct in complete isolation from external dependencies.

**What Mocks Are Permitted**:
- **Interfaces defined in the same module**: Mock implementations of interfaces that the code under test depends on, where the mock is defined in a `*_test.go` file or `test/` subdirectory.
- **External dependencies with non-deterministic behavior**: Time (`time.Now`), random number generation, UUID generation may be mocked or controlled via dependency injection.
- **Network/disk I/O**: For pure business logic tests, network clients and filesystem operations may be mocked.

**What Mocks Are FORBIDDEN**:
- Mocking the thing you are testing (the System Under Test itself)
- Mocking every dependency so the test only verifies mock wiring
- Mocking database operations when testing database query builders
- Mocking HTTP handlers when testing HTTP middleware

**Required Coverage**: 100% branch coverage of isolated business logic. Lines that are impossible to reach (e.g., defensive checks for conditions that can't occur in practice) must be annotated with `// unreachable: <reason>` and reviewed in PR.

**CI Integration**:
```yaml
# Part of ci.yml — Unit test job
unit-tests:
  name: Unit Tests
  runs-on: ubuntu-latest
  container: golang:1.26
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Fix replace directives
      run: ./scripts/fix-replace.sh
    - name: Run unit tests
      run: go test -count=1 -race -p 1 ./...
    - name: Upload coverage
      uses: actions/upload-artifact@v4
      with:
        name: unit-coverage
        path: coverage.out
```

**Example Test Pattern** (legitimate Unit test):
```go
func TestRateLimiter_Allow(t *testing.T) {
    // Arrange: Create a real rate limiter with a test clock
    clock := testclock.New()
    rl := NewRateLimiter(10, time.Second, WithClock(clock))

    // Act: Consume all 10 tokens
    for i := 0; i < 10; i++ {
        assert.True(t, rl.Allow(), "request %d should be allowed", i)
    }

    // Assert: 11th request rejected
    assert.False(t, rl.Allow(), "11th request should be rejected")

    // Act: Advance clock by 1 second
    clock.Advance(time.Second)

    // Assert: Bucket refilled, request allowed again
    assert.True(t, rl.Allow(), "request after refill should be allowed")
}
```

**Anti-Bluff Verification Method**:
- Run `go-mutesting` on the package. If mutants survive, the Unit test is not actually verifying the logic.
- Negative leg: Introduce a deliberate bug (e.g., change `>=` to `>`) and verify the Unit test fails.

**Directory Structure**:
```
<module>/
  pkg/<package>/
    <file>.go
    <file>_test.go          # Unit tests (colocated)
  tests/unit/               # Alternative: centralized unit tests
    <package>_test.go
```

### 2.2 Integration Tests

**Purpose**: Verify that multiple real components work together correctly. Integration tests exercise the actual interaction paths between modules with real (not mocked) dependencies.

**What Mocks Are Permitted**: **NONE**. Integration tests must use real dependencies:
- Real databases (SQLite in-memory or test-container PostgreSQL)
- real HTTP servers (`httptest.NewServer` is acceptable — it is a real HTTP server, not a mock)
- Real message queues (test-container RabbitMQ/NATS)
- Real caches (test-container Redis or in-memory)
- Real filesystem operations (temporary directories)

**Required Coverage**: All cross-component interaction paths. Every pair of components that communicate must have at least one integration test verifying that communication.

**CI Integration**:
```yaml
integration-tests:
  name: Integration Tests
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:16
      env:
        POSTGRES_PASSWORD: test
        POSTGRES_DB: helixtest
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
      ports: ['25432:5432']
    redis:
      image: redis:7-alpine
      ports: ['26379:6379']
    nats:
      image: nats:2-alpine
      ports: ['4222:4222']
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Fix replace directives
      run: ./scripts/fix-replace.sh
    - name: Run integration tests
      run: go test -count=1 -race -p 1 ./tests/integration/...
      env:
        HELIX_TEST_DB_DSN: postgres://postgres:test@localhost:25432/helixtest?sslmode=disable
        HELIX_TEST_REDIS_ADDR: localhost:26379
        HELIX_TEST_NATS_URL: nats://localhost:4222
```

**Example Test Pattern**:
```go
func TestAuthStorage_Integration(t *testing.T) {
    // Arrange: Real database connection
    db, err := database.Connect(os.Getenv("HELIX_TEST_DB_DSN"))
    require.NoError(t, err)
    defer db.Close()

    // Arrange: Real auth manager using real DB
    auth := auth.NewManager(db, auth.WithJWTSecret("test-secret"))

    // Act: Register a user
    token, err := auth.Register(ctx, "test@example.com", "password123")
    require.NoError(t, err)
    require.NotEmpty(t, token)

    // Act: Validate the token
    claims, err := auth.ValidateToken(ctx, token)
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", claims.Subject)

    // Act: Login with same credentials
    token2, err := auth.Login(ctx, "test@example.com", "password123")
    require.NoError(t, err)
    assert.NotEmpty(t, token2)

    // Negative leg: Wrong password
    _, err = auth.Login(ctx, "test@example.com", "wrongpassword")
    assert.Error(t, err)
}
```

**Anti-Bluff Verification Method**:
- Temporarily break the integration point (e.g., change the table name in one component) and verify the test fails.
- Verify no mocks are used: `grep -r "mock\|Mock" tests/integration/` should return zero matches.

**Directory Structure**:
```
<module>/
  tests/integration/
    <component_pair>_test.go    # e.g., auth_storage_test.go
    <flow>_test.go              # e.g., user_registration_flow_test.go
```

### 2.3 E2E Tests

**Purpose**: Verify complete user journeys from the external entry point through the entire system, using a production-like topology.

**What Mocks Are Permitted**: **NONE**. E2E tests must exercise the full stack:
- Real compiled binaries running in containers
- Real network communication (localhost ports or Docker network)
- Real databases, caches, message queues
- Real client interactions (HTTP requests, gRPC calls, WebRTC streams)

**Required Coverage**: All primary user journeys. Every user story in the spec must have at least one corresponding E2E test.

**CI Integration**:
```yaml
e2e-tests:
  name: E2E Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Build all services
      run: docker compose -f tests/e2e/docker-compose.yml build
    - name: Run E2E test suite
      run: go test -count=1 -v -timeout 30m ./tests/e2e/...
    - name: Collect logs on failure
      if: failure()
      run: docker compose -f tests/e2e/docker-compose.yml logs > e2e-logs.txt
    - name: Upload E2E artifacts
      if: failure()
      uses: actions/upload-artifact@v4
      with:
        name: e2e-failure-artifacts
        path: |
          e2e-logs.txt
          tests/e2e/screenshots/
```

**Example Test Pattern** (Host Discovery E2E):
```go
func TestEndToEnd_HostDiscoveryAndConnect(t *testing.T) {
    // Arrange: Start core backend, host agent, and client in containers
    compose := e2e.MustUp(t, "docker-compose.e2e.yml")
    defer compose.Down()

    coreURL := compose.ServiceURL("core-backend")
    hostURL := compose.ServiceURL("host-agent")

    // Act: Host agent registers with core
    hostClient := host.NewClient(hostURL)
    caps, err := hostClient.AdvertiseCapabilities(ctx, host.Capabilities{
        GPU: "RTX 4090",
        Codecs: []string{"HEVC", "AV1"},
        Games: []string{"Elden Ring", "Cyberpunk 2077"},
    })
    require.NoError(t, err)
    assert.NotEmpty(t, caps.HostID)

    // Act: Client queries available hosts
    client := helixplay.NewClient(coreURL)
    hosts, err := client.DiscoverHosts(ctx)
    require.NoError(t, err)
    assert.Len(t, hosts, 1)
    assert.Equal(t, "RTX 4090", hosts[0].GPU)

    // Act: Client requests connection to host
    session, err := client.ConnectToHost(ctx, hosts[0].HostID)
    require.NoError(t, err)
    assert.NotEmpty(t, session.SessionID)
    assert.Equal(t, "ready", session.Status)

    // Negative leg: Connect to non-existent host
    _, err = client.ConnectToHost(ctx, "non-existent-host-id")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}
```

**Anti-Bluff Verification Method**:
- Stop one critical service container mid-test and verify the E2E test detects the failure.
- Verify the test exercises real network paths (not in-process calls).
- HelixQA visual assertion for UI-facing E2E flows.

**Directory Structure**:
```
<module>/
  tests/e2e/
    docker-compose.e2e.yml      # Full topology definition
    <journey>_test.go            # e.g., host_discovery_test.go
    <journey>_test.go            # e.g., game_stream_test.go
    fixtures/                    # Test data, configs, game stubs
```

### 2.4 Security Tests

**Purpose**: Identify vulnerabilities through automated security testing including fuzzing, static application security testing (SAST), and dynamic application security testing (DAST).

**What Mocks Are Permitted**: **NONE**. Security tests must target the real application to find real vulnerabilities.

**Required Coverage**: All attack surfaces:
- HTTP/gRPC endpoints (injection, authentication bypass, authorization bypass)
- Input validation (malformed JSON, oversized payloads, path traversal)
- Cryptographic operations (weak algorithms, key exposure)
- Container configurations (privileged mode, exposed secrets)
- Dependency vulnerabilities (known CVEs in transitive dependencies)

**CI Integration**:
```yaml
security-tests:
  name: Security Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run gofuzz targets
      run: go test -fuzz=FuzzAuth -fuzztime=60s ./tests/security/...
    - name: Run gofuzz targets (all)
      run: go test -fuzz=. -fuzztime=30s ./...
    - name: SAST — Semgrep
      uses: returntocorp/semgrep-action@v1
      with:
        config: >-
          p/security-audit
          p/owasp-top-ten
          p/cwe-top-25
          p/gosec
    - name: SAST — govulncheck
      run: govulncheck ./...
    - name: Container scan — Trivy
      run: trivy fs --exit-code 1 --severity HIGH,CRITICAL .
    - name: Secret scan — gitleaks
      run: gitleaks detect --source . --verbose
```

**Example Test Pattern**:
```go
func FuzzAuth_ValidateToken(f *testing.F) {
    // Seed corpus with valid and invalid tokens
    f.Add("eyJhbGciOiJIUzI1NiIs...")  // valid
    f.Add("invalid.token.here")
    f.Add("")
    f.Add("../../../etc/passwd")

    f.Fuzz(func(t *testing.T, token string) {
        // This should NEVER panic regardless of input
        claims, err := auth.ValidateToken(ctx, token)
        // We only care that it doesn't panic; error is expected for fuzz inputs
        _ = claims
        _ = err
    })
}

func TestSecurity_SQLInjection(t *testing.T) {
    db := testdb.New(t)
    defer db.Close()

    // Attempt SQL injection in user input
    maliciousInputs := []string{
        "'; DROP TABLE users; --",
        "1 OR 1=1",
        "' UNION SELECT * FROM passwords --",
    }

    for _, input := range maliciousInputs {
        // This should return an error or safe result, NEVER execute the injected SQL
        _, err := userRepo.FindByUsername(ctx, input)
        // Expect error or empty result, not a data breach
        assert.True(t, err != nil || len(users) == 0, "SQL injection possible with: %s", input)
    }
}
```

**Anti-Bluff Verification Method**:
- Introduce a known vulnerability (e.g., remove input sanitization) and verify the security test detects it.
- Verify fuzzing actually finds crashes: check that the fuzz corpus grows over time.

**Directory Structure**:
```
<module>/
  tests/security/
    fuzz_<target>_test.go       # Fuzzing targets
    <attack>_test.go            # Specific attack vectors
    owasp_top10_test.go         # OWASP Top 10 coverage
```

### 2.5 Benchmarking

**Purpose**: Measure and track performance characteristics of critical code paths. Benchmarks detect performance regressions and validate latency SLAs.

**What Mocks Are Permitted**: **NONE**. Benchmarks must measure real code with real data to produce meaningful results.

**Required Coverage**: All performance-critical paths:
- Stream encoding pipeline (frame encode latency)
- Database query paths (query execution time)
- Memory allocation hot paths (allocs/op)
- Network serialization/deserialization
- JWT signing/validation
- Game discovery and enumeration
- Storage upload/download throughput

**SLA Targets**:
| Metric | p50 Target | p99 Target | p999 Target |
|--------|-----------|-----------|-------------|
| Frame encode latency | 8ms | 16ms | 33ms |
| End-to-end stream latency (LAN) | 15ms | 30ms | 50ms |
| End-to-end stream latency (WAN) | 25ms | 50ms | 100ms |
| Game discovery response | 50ms | 200ms | 500ms |
| Auth token validation | 1ms | 5ms | 10ms |
| Storage upload (1MB chunk) | 100ms | 500ms | 1000ms |

**CI Integration**:
```yaml
benchmarks:
  name: Benchmarks
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run benchmarks
      run: go test -bench=. -benchmem -benchtime=5s ./tests/benchmark/... | tee benchmark.txt
    - name: Compare with baseline
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'go'
        output-file-path: benchmark.txt
        github-token: ${{ secrets.GITHUB_TOKEN }}
        alert-threshold: '150%'  # Fail if 50% slower than baseline
        comment-on-alert: true
        fail-on-alert: true
    - name: Upload benchmark results
      uses: actions/upload-artifact@v4
      with:
        name: benchmark-results
        path: benchmark.txt
```

**Example Test Pattern**:
```go
func BenchmarkEncoder_EncodeFrame(b *testing.B) {
    enc := encoder.New(encoder.Config{Codec: "HEVC", Bitrate: 20_000_000})
    frame := generateTestFrame(1920, 1080, pixel.RGBA)

    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        packet, err := enc.EncodeFrame(frame)
        if err != nil {
            b.Fatal(err)
        }
        if len(packet.Data) == 0 {
            b.Fatal("empty packet")
        }
    }
}

func BenchmarkAuth_ValidateToken(b *testing.B) {
    auth := auth.NewManager(db, auth.WithJWTSecret("benchmark-secret"))
    token, _ := auth.GenerateToken(ctx, "benchmark-user")

    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        _, err := auth.ValidateToken(ctx, token)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Anti-Bluff Verification Method**:
- Introduce an artificial slowdown (e.g., add `time.Sleep(10 * time.Millisecond)`) and verify the benchmark detects the regression.
- Verify benchmark results are actually stored and compared (not just generated and discarded).

**Directory Structure**:
```
<module>/
  tests/benchmark/
    <component>_bench_test.go    # e.g., encoder_bench_test.go
    benchmark_baseline.txt       # Committed baseline for comparison
```

### 2.6 Chaos Tests

**Purpose**: Verify system resilience by injecting faults into a running production-like deployment. Chaos tests prove the system degrades gracefully under failure conditions.

**What Mocks Are Permitted**: Only the chaos injection mechanisms themselves (e.g., the tool that kills a container is a test fixture; the container being killed is a real service).

**Required Coverage**: All failure modes:
- Container/pod crashes (kill random service)
- Network partitions (isolate service from its dependencies)
- Latency injection (add 500ms to all DB queries)
- Packet loss (drop 10% of UDP stream packets)
- Resource exhaustion (CPU throttling, memory pressure)
- DNS failure (unresolvable service names)
- Certificate expiry (invalid TLS certificates)

**CI Integration**:
```yaml
chaos-tests:
  name: Chaos Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Start full topology
      run: docker compose -f tests/chaos/docker-compose.yml up -d
    - name: Run chaos test suite
      run: go test -count=1 -v -timeout 60m ./tests/chaos/...
      env:
        CHAOS_DURATION: 5m
        CHAOS_INTERVAL: 30s
    - name: Collect chaos experiment results
      if: always()
      run: |
        docker compose -f tests/chaos/docker-compose.yml logs > chaos-logs.txt
        kubectl describe chaosresults > chaos-results.txt
    - name: Upload artifacts
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: chaos-artifacts
        path: |
          chaos-logs.txt
          chaos-results.txt
```

**Example Test Pattern**:
```go
func TestChaos_HostAgentCrash_StreamContinues(t *testing.T) {
    // Arrange: Start full streaming session
    topo := chaos.MustStartTopology(t, "streaming-topology.yml")
    defer topo.Teardown()

    client := topo.Client()
    hostID := topo.HostID(0)

    // Start streaming
    session, err := client.ConnectToHost(ctx, hostID)
    require.NoError(t, err)

    stream, err := client.StartStream(ctx, session.SessionID)
    require.NoError(t, err)

    // Verify stream is delivering frames
    frameCount := countFrames(stream, 5*time.Second)
    require.Greater(t, frameCount, 0, "stream should deliver frames")

    // Act: Kill the host agent container
    topo.KillContainer("host-agent")

    // Assert: Stream should detect disconnection within timeout
    err = waitForStreamError(stream, 10*time.Second)
    assert.NoError(t, err, "stream should detect host disconnection")

    // Assert: Client should be able to reconnect to another host
    topo.StartContainer("host-agent")  // Restart
    hostID2 := topo.HostID(1)
    session2, err := client.ConnectToHost(ctx, hostID2)
    assert.NoError(t, err, "should connect to alternate host")
    assert.NotNil(t, session2)
}
```

**Anti-Bluff Verification Method**:
- Run the chaos test without injecting chaos and verify it fails (or is skipped) — a chaos test that passes without chaos being injected is a bluff.
- Verify the fault injection actually occurred by checking container restart counts / network drop logs.

**Directory Structure**:
```
<module>/
  tests/chaos/
    <failure_mode>_test.go       # e.g., container_crash_test.go
    <failure_mode>_test.go       # e.g., network_partition_test.go
    docker-compose.chaos.yml     # Chaos topology
    chaos-experiments/           # Litmus/chaos-mesh experiment definitions
```

### 2.7 Stress Tests

**Purpose**: Verify system behavior under sustained load beyond normal capacity. Stress tests find resource leaks, deadlock conditions, and degradation patterns that only appear over time.

**What Mocks Are Permitted**: **NONE**. Stress tests must exercise real services under real load.

**Required Coverage**:
- 24-hour sustained load profile (memory leak detection)
- Connection saturation (10,000 concurrent connections)
- Request flood (10x normal request rate for 1 hour)
- Storage exhaustion (fill disk to 95%)
- Memory pressure (reduce available RAM by 50%)
- Database connection pool exhaustion
- Goroutine leak detection (goroutine count must not grow unboundedly)

**CI Integration**:
```yaml
stress-tests:
  name: Stress Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Build stress test services
      run: docker compose -f tests/stress/docker-compose.yml build
    - name: Run 4-hour stress profile
      run: go test -count=1 -v -timeout 240m ./tests/stress/...
      env:
        STRESS_DURATION: 4h
        STRESS_CONCURRENCY: 1000
    - name: Collect resource usage data
      if: always()
      run: |
        cat tests/stress/resource-usage.log > stress-resources.txt
        go tool pprof -top tests/stress/profile.pb.gz > stress-pprof.txt
    - name: Upload artifacts
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: stress-artifacts
        path: |
          stress-resources.txt
          stress-pprof.txt
```

**Example Test Pattern**:
```go
func TestStress_24HourStreaming(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping 24-hour stress test in short mode")
    }

    topo := stress.MustStartTopology(t, "streaming-topology.yml")
    defer topo.Teardown()

    client := topo.Client()
    hostID := topo.HostID(0)

    // Start streaming session
    session, err := client.ConnectToHost(ctx, hostID)
    require.NoError(t, err)

    stream, err := client.StartStream(ctx, session.SessionID)
    require.NoError(t, err)

    // Collect baseline metrics
    baselineGoroutines := runtime.NumGoroutine()
    baselineMem := getMemUsage()

    // Run stream for 24 hours, collecting frames
    duration := 24 * time.Hour
    if os.Getenv("CI") == "true" {
        duration = 4 * time.Hour  // Shorter in CI
    }

    ctx, cancel := context.WithTimeout(ctx, duration)
    defer cancel()

    framesReceived := int64(0)
    go func() {
        for {
            _, err := stream.RecvFrame()
            if err != nil {
                return
            }
            atomic.AddInt64(&framesReceived, 1)
        }
    }()

    <-ctx.Done()

    // Assert: Should have received frames throughout
    assert.Greater(t, atomic.LoadInt64(&framesReceived), int64(0), "should have received frames")

    // Assert: No goroutine leak
    finalGoroutines := runtime.NumGoroutine()
    assert.LessOrEqual(t, finalGoroutines, baselineGoroutines+10,
        "goroutine leak detected: baseline=%d, final=%d", baselineGoroutines, finalGoroutines)

    // Assert: Memory growth bounded
    finalMem := getMemUsage()
    memGrowth := float64(finalMem-baselineMem) / float64(baselineMem) * 100
    assert.LessOrEqual(t, memGrowth, 50.0,
        "memory grew by %.1f%% over %v", memGrowth, duration)
}
```

**Anti-Bluff Verification Method**:
- Verify the test actually ran for the full duration (check timestamps in logs).
- Introduce a goroutine leak (intentionally forget to close a goroutine) and verify the stress test detects it.

**Directory Structure**:
```
<module>/
  tests/stress/
    <scenario>_stress_test.go    # e.g., streaming_stress_test.go
    docker-compose.stress.yml    # Stress topology
    profiles/                    # Load profiles (normal, peak, extreme)
```

### 2.8 Smoke Tests

**Purpose**: Provide rapid post-deployment validation that the system is minimally functional. Smoke tests run in under 30 seconds and cover the most critical paths.

**What Mocks Are Permitted**: **NONE**. Smoke tests verify real deployed services.

**Required Coverage**: All critical endpoints:
- Health check endpoints return 200 OK
- Core backend responds to discovery queries
- Host agent beacon is active
- Database connectivity
- Auth service token generation
- Storage service read/write
- All container statuses are healthy

**CI Integration**:
```yaml
smoke-tests:
  name: Smoke Tests
  runs-on: ubuntu-latest
  needs: [deploy-staging]
  steps:
    - uses: actions/checkout@v4
    - name: Run smoke tests against staging
      run: go test -count=1 -v -timeout 60s ./tests/smoke/...
      env:
        HELIX_BASE_URL: https://staging.helixplay.dev
    - name: Notify on failure
      if: failure()
      uses: slack-action/notify@v1
      with:
        message: "SMOKE TEST FAILED on staging — deployment may be broken"
```

**Example Test Pattern**:
```go
func TestSmoke_HealthChecks(t *testing.T) {
    baseURL := os.Getenv("HELIX_BASE_URL")
    require.NotEmpty(t, baseURL, "HELIX_BASE_URL required")

    services := []struct {
        name   string
        path   string
        expect int
    }{
        {"core-backend", "/health", 200},
        {"host-agent", "/health", 200},
        {"auth-service", "/health", 200},
        {"storage-service", "/health", 200},
    }

    for _, svc := range services {
        t.Run(svc.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
            defer cancel()

            req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+svc.path, nil)
            resp, err := http.DefaultClient.Do(req)
            require.NoError(t, err, "%s is unreachable", svc.name)
            defer resp.Body.Close()

            assert.Equal(t, svc.expect, resp.StatusCode, "%s health check failed", svc.name)
        })
    }
}
```

**Anti-Bluff Verification Method**:
- Deploy a broken version (e.g., service that returns 500 on health check) and verify smoke tests fail.
- Measure actual execution time — smoke tests that take > 30 seconds indicate a problem.

**Directory Structure**:
```
<module>/
  tests/smoke/
    smoke_test.go               # All smoke tests (fast, consolidated)
```

### 2.9 Full Automation Tests

**Purpose**: Verify that the entire system can be deployed, configured, and operated without human intervention. Full automation tests are the ultimate integration test — they test the system as a whole, including deployment automation.

**What Mocks Are Permitted**: **NONE**. Full automation tests must use real infrastructure (even if that infrastructure is Docker Compose on CI runners).

**Required Coverage**:
- Full deployment from clean state completes without manual steps
- All services start and reach healthy state
- All 10 test types can be triggered and complete
- Configuration is automatically applied
- Monitoring and alerting are functional
- Rollback procedure works
- Backup and restore procedures work

**CI Integration**:
```yaml
full-automation:
  name: Full Automation
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Full automation test
      run: go test -count=1 -v -timeout 120m ./tests/fullauto/...
      env:
        FULLAUTO_MODE: ci
    - name: Upload full automation report
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: fullauto-report
        path: tests/fullauto/report/
```

**Example Test Pattern**:
```go
func TestFullAutomation_DeployAndOperate(t *testing.T) {
    // Arrange: Clean environment
    ctx := context.Background()
    deployer := fullauto.NewDeployer()

    // Act: Full deployment from clean state
    deployment, err := deployer.Deploy(ctx, fullauto.Config{
        Topology:   "minimal-3-node",
        Version:    "latest",
        AutoConfig: true,
    })
    require.NoError(t, err, "deployment should complete without human intervention")
    defer deployment.Teardown()

    // Assert: All services healthy
    services := deployment.ListServices()
    assert.GreaterOrEqual(t, len(services), 5, "should have at least 5 services")
    for _, svc := range services {
        assert.Equal(t, "healthy", svc.Status, "service %s should be healthy", svc.Name)
    }

    // Act: Trigger all 10 test types
    results := deployment.RunAllTests(ctx)
    assert.Equal(t, 10, results.TotalTypes, "all 10 test types should be executable")
    assert.Equal(t, 10, results.PassedTypes, "all 10 test types should pass")

    // Act: Simulate service failure and verify auto-recovery
    deployment.KillService("host-agent")
    err = deployment.WaitForRecovery(ctx, "host-agent", 2*time.Minute)
    assert.NoError(t, err, "host-agent should auto-recover")

    // Act: Backup and restore
    backup, err := deployment.Backup(ctx)
    require.NoError(t, err)

    deployment.Teardown()
    restored, err := deployer.Restore(ctx, backup)
    require.NoError(t, err, "restore should complete without human intervention")
    defer restored.Teardown()

    // Assert: Restored system functional
    restoredServices := restored.ListServices()
    assert.GreaterOrEqual(t, len(restoredServices), 5, "restored system should have all services")
}
```

**Anti-Bluff Verification Method**:
- Run the full automation test with a manual step required (e.g., a prompt that needs human input) and verify it times out and fails.
- Verify the deployment is truly clean (no pre-existing state, containers, or volumes).

**Directory Structure**:
```
<module>/
  tests/fullauto/
    deploy_test.go              # Deployment automation
    operate_test.go             # Operational procedures
    disaster_recovery_test.go   # Backup/restore
```

### 2.10 Challenges

**Purpose**: Execute structured, production-equivalent test scenarios defined in the Challenges repository. Challenges are the highest-fidelity test type — they verify complete feature behavior in an environment that mirrors production.

**What Mocks Are Permitted**: **NONE**. Challenges use real everything.

**Required Coverage**: All challenge bank scenarios:
- Game-specific challenges (each supported game)
- Full QA challenges (Android, Android TV, Desktop, Web)
- Navigation challenges (all UI flows)
- Recording challenges (stream capture, storage, playback)
- Admin challenges (user management, system configuration)
- Security challenges (auth, rate limiting, DDoS)
- Performance challenges (latency, throughput under load)
- Regression challenges (previously fixed bugs must stay fixed)

**CI Integration**:
```yaml
challenges:
  name: Challenges
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive
    - name: Build challenge runner
      run: cd vasic-digital/Challenges && go build -o challenge-runner ./cmd/userflow-runner
    - name: Run all challenges
      run: ./vasic-digital/Challenges/challenge-runner --all --report challenges-report.json
    - name: Upload challenge report
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: challenges-report
        path: challenges-report.json
    - name: Anti-bluff validation
      run: |
        if ! jq '.antiBluffValid' challenges-report.json | grep -q true; then
          echo "ANTI-BLUFF VALIDATION FAILED"
          exit 1
        fi
```

**Example Challenge Definition**:
```yaml
# challenges/banks/game-streaming.yaml
challenge:
  name: "HelixPlay Game Stream — Elden Ring"
  category: "game-streaming"
  priority: 1
  steps:
    - name: "Discover host"
      action: "api_call"
      endpoint: "/api/v1/discover"
      assert:
        - status: 200
        - jsonpath: "$.hosts[0].gpu" equals "RTX 4090"
    - name: "Connect to host"
      action: "api_call"
      endpoint: "/api/v1/connect"
      body: '{"host_id": "${hosts[0].id}"}'
      assert:
        - status: 201
        - jsonpath: "$.status" equals "ready"
    - name: "Launch game"
      action: "api_call"
      endpoint: "/api/v1/games/launch"
      body: '{"game": "Elden Ring", "session_id": "${session.id}"}'
      assert:
        - status: 200
        - jsonpath: "$.state" equals "running"
    - name: "Verify stream active"
      action: "verify_stream"
      timeout: 30s
      assert:
        - fps greater_than 30
        - latency_ms less_than 50
        - resolution equals "1920x1080"
  antiBluff:
    recordedActions: ["api_call", "verify_stream"]
    assertions: ["status", "jsonpath", "fps", "latency_ms", "resolution"]
```

**Anti-Bluff Verification Method**:
- `ValidateAntiBluff()` is called unconditionally on every challenge result.
- Challenge scripts (Section 8) verify compilation, functionality, and unit test coverage.
- The bluff scanner runs against challenge definitions to detect vacuous assertions.

**Directory Structure**:
```
vasic-digital/Challenges/
  banks/                        # Challenge bank definitions
    game-streaming.yaml
    full-qa-android.yaml
    full-qa-androidtv.yaml
    navigation.yaml
    recording.yaml
    admin.yaml
    security.yaml
    performance.yaml
    regression.yaml
  pkg/challenge/                # Challenge framework
  pkg/runner/                   # Execution engine
  scripts/                      # Challenge scripts (Section 8)
```

---

## Section 3: Test Matrix (29 Submodules x 10 Types)

### 3.1 Submodules Overview

The HelixPlay platform consists of **29 submodules** organized into three groups:

| Group | Count | Submodules |
|-------|-------|-----------|
| Core Infrastructure | 19 | Auth, Cache, Database, Discovery, EventBus, Formatters, Media, Memory, Messaging, Middleware, Observability, Plugins, RAG, RateLimiter, Recovery, Security, Storage, Streaming, VectorDB |
| Challenges | 1 | vasic-digital/Challenges |
| HelixQA | 1 | HelixDevelopment/HelixQA |
| Application Modules | 8 | client-wails, client-web, core-backend, host-agent/capability, host-agent/capture, host-agent/codec, host-agent/discovery, host-agent/encoder, host-agent/game, host-agent/input, host-agent/lifecycle, host-agent/transport |

*Note: The host-agent submodules are treated as a single unit for some test types but individually for Unit and Integration tests.*

### 3.2 Complete Test Matrix

| # | Submodule | Unit | Integ | E2E | Sec | Bench | Chaos | Stress | Smoke | FullAuto | Chall | Owner | Status |
|---|-----------|:----:|:-----:|:---:|:---:|:-----:|:-----:|:------:|:-----:|:--------:|:-----:|-------|--------|
| 1 | **Auth** | [x] | [x] | [x] | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Auth | Partial |
| 2 | **Cache** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Cache | Partial |
| 3 | **Database** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Database | Partial |
| 4 | **Discovery** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Discovery | Partial |
| 5 | **EventBus** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | EventBus | Partial |
| 6 | **Formatters** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Formatters | Minimal |
| 7 | **Media** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Media | Partial |
| 8 | **Memory** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Memory | Minimal |
| 9 | **Messaging** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Messaging | Partial |
| 10 | **Middleware** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Middleware | Partial |
| 11 | **Observability** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Observability | Partial |
| 12 | **Plugins** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Plugins | Minimal |
| 13 | **RAG** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | RAG | Partial |
| 14 | **RateLimiter** | [x] | [x] | [x] | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | RateLimiter | Partial |
| 15 | **Recovery** | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Recovery | Partial |
| 16 | **Security** | [x] | [x] | [x] | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Security | Partial |
| 17 | **Storage** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Storage | **Good** |
| 18 | **Streaming** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Streaming | Partial |
| 19 | **VectorDB** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | VectorDB | Minimal |
| 20 | **Challenges** | [x] | [x] | [x] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Challenges | **Good** |
| 21 | **HelixQA** | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [x] | HelixQA | Partial |
| 22 | **client-wails** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Client | Minimal |
| 23 | **client-web** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Client | Minimal |
| 24 | **core-backend** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | Core | Minimal |
| 25 | **host-agent/capability** | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [x] | Host | Partial |
| 26 | **host-agent/capture** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [x] | Host | Partial |
| 27 | **host-agent/codec** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Host | Partial |
| 28 | **host-agent/discovery** | [x] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [ ] | [x] | Host | Partial |
| 29 | **host-agent/encoder** | [x] | [x] | [ ] | [ ] | [x] | [ ] | [ ] | [ ] | [ ] | [ ] | Host | Partial |
| | **TOTAL** | **29** | **18** | **8** | **4** | **18** | **0** | **0** | **0** | **0** | **11** | | |

### 3.3 Priority for Missing Tests

**Priority 1 — Immediate (Sprint 1-2)**:
| Submodule | Test Type | Why Critical | Effort |
|-----------|-----------|-------------|--------|
| All 29 | Full Automation | No full automation exists; this is the most critical gap | 2 weeks |
| All 29 | Smoke | No smoke tests exist; every deploy is blind | 1 week |
| Streaming, Storage, Auth | Chaos | Most critical services need resilience validation | 2 weeks |
| Streaming, Storage, Auth | Stress | Verify no resource leaks under sustained load | 2 weeks |

**Priority 2 — High (Sprint 3-4)**:
| Submodule | Test Type | Why High | Effort |
|-----------|-----------|----------|--------|
| Cache, EventBus, Messaging | E2E | Core infrastructure needs E2E coverage | 1 week |
| All host-agent/* | E2E | Host agent is the core product; needs E2E | 2 weeks |
| client-wails, client-web | All types | Client has minimal test coverage | 3 weeks |
| Formatters, Memory, Plugins, VectorDB | Integration | No integration tests at all | 1 week |

**Priority 3 — Medium (Sprint 5-6)**:
| Submodule | Test Type | Why Medium | Effort |
|-----------|-----------|------------|--------|
| RAG, Observability | E2E | Important but not critical path | 1 week |
| Recovery | E2E, Chaos | Fault recovery needs chaos testing | 1 week |
| core-backend | All types | Core backend is stub; tests depend on implementation | 2 weeks |
| All | Security (fuzz) | Fuzzing coverage expansion | 2 weeks |

---


## Section 4: Anti-Bluff Infrastructure Implementation

### 4.1 ValidateAntiBluff() Validator

**Location**: `vasic-digital/Challenges/pkg/challenge/antibluff.go`  
**Purpose**: Unconditionally validate that a challenge result claiming `Status=Passed` actually has evidence of execution.

**Constitutional Basis (v2.1.0)**:
> *"`ValidateAntiBluff` gate is unconditional; `CHALLENGE_ANTIBLUFF_STRICT` removed."*

#### 4.1.1 Implementation Details

The `ValidateAntiBluff()` function is called on every `Result` object, regardless of test type or challenge category. It enforces three rules:

```go
// pkg/challenge/antibluff.go
package challenge

import "fmt"

var ErrBluffPass = fmt.Errorf("antibluff: result claims Pass but lacks execution evidence")

type Result struct {
    Status         Status      // Passed, Failed, Skipped, Error
    RecordedActions []Action   // What the runtime actually did
    Assertions     []Assertion // What was checked
    // ... other fields
}

type Action struct {
    Name      string
    Timestamp int64
    Details   map[string]interface{}
}

type Assertion struct {
    Name   string
    Passed bool
    Details map[string]interface{}
}

// ValidateAntiBluff enforces three rules on any Result with Status=Passed.
// This function is called UNCONDITIONALLY — no toggle, no bypass.
func ValidateAntiBluff(r *Result) error {
    // Rule 1: Non-Pass statuses are honest by definition
    if r.Status != StatusPassed {
        return nil // Failed, Skipped, Error — these can't be bluff passes
    }

    // Rule 2: RecordedActions must be non-empty
    // Proof the runtime actually executed something
    if len(r.RecordedActions) == 0 {
        return fmt.Errorf("%w: zero recorded actions for Passed result", ErrBluffPass)
    }

    // Rule 3: Assertions must be non-empty
    // At least one expectation was checked
    if len(r.Assertions) == 0 {
        return fmt.Errorf("%w: no assertions recorded for Passed result", ErrBluffPass)
    }

    // Rule 4: At least one assertion must have Passed=true
    // Something was positively confirmed
    hasPassingAssertion := false
    for _, a := range r.Assertions {
        if a.Passed {
            hasPassingAssertion = true
            break
        }
    }
    if !hasPassingAssertion {
        return fmt.Errorf("%w: all assertions failed but status is Passed", ErrBluffPass)
    }

    return nil
}
```

#### 4.1.2 Three Enforcement Rules

| Rule | Name | Purpose | Failure Mode |
|------|------|---------|-------------|
| **Rule 1** | Non-Pass Honesty | `Failed`, `Skipped`, `Error` statuses are not validated (they can't be bluff passes) | N/A |
| **Rule 2** | Action Evidence | `RecordedActions` must contain at least one entry proving the runtime executed code | Result claims Pass but did nothing |
| **Rule 3** | Assertion Evidence | `Assertions` must contain at least one entry, and at least one must have `Passed=true` | Result claims Pass but no assertions passed |

#### 4.1.3 Integration into Challenge Runner

The validator is integrated at three points in the challenge execution pipeline:

```go
// pkg/runner/runner.go — ExecuteChallenge
func (r *Runner) ExecuteChallenge(ctx context.Context, ch challenge.Challenge) (*challenge.Result, error) {
    // Execute the challenge
    result, err := ch.Execute(ctx)
    if err != nil {
        return nil, err
    }

    // UNCONDITIONAL anti-bluff validation
    if err := challenge.ValidateAntiBluff(result); err != nil {
        // Force status to Error if anti-bluff validation fails
        result.Status = challenge.StatusError
        result.ErrorDetails = err.Error()
        return result, err
    }

    return result, nil
}
```

**Integration Points**:
1. **Challenge Runner** (`pkg/runner/`): Every executed challenge is validated before the result is returned.
2. **Report Generation** (`pkg/report/`): Reports include an `antiBluffValid` boolean field that aggregates validation across all challenges.
3. **CI Pipeline**: The `challenges.yml` workflow checks `antiBluffValid` in the report JSON and fails if any challenge lacks evidence.

#### 4.1.4 Anti-Bluff Test Suite

The validator itself is tested with 9 test cases in `pkg/challenge/antibluff_test.go`:

| Test | Scenario | Expected |
|------|----------|----------|
| `TestValidate_PassWithEvidence` | Happy path: Pass with actions and passing assertion | `nil` (valid) |
| `TestValidate_PassWithZeroActions` | THE bluff pattern: claims Pass but did nothing | `ErrBluffPass` |
| `TestValidate_PassWithEmptyAssertions` | Metadata-only pattern: actions recorded but nothing checked | `ErrBluffPass` |
| `TestValidate_PassWithAllAssertionsFailing` | Most insidious bluff: all assertions failed but status=Pass | `ErrBluffPass` |
| `TestValidate_PassWithMixedAssertions` | Mixed pass/fail is OK if at least one passes | `nil` (valid) |
| `TestValidate_StatusFailedHonest` | Non-Pass statuses are honest by definition | `nil` (valid) |
| `TestValidate_StatusSkipped` | Skipped status doesn't need validation | `nil` (valid) |
| `TestRecordAction` | Action recording works correctly | Action recorded |
| `TestRecordAction_NilReceiver` | Defensive check against nil panic | No panic |

### 4.2 Anti-Bluff Scan Script

**Location**: `HelixPlay/scripts/anti-bluff-scan.sh`  
**Purpose**: Five-step comprehensive scan of the entire source tree for bluff patterns, forbidden code, and constitutional compliance.

#### 4.2.1 Five-Step Scan Process

```bash
#!/bin/bash
# scripts/anti-bluff-scan.sh

set -euo pipefail

REPORT_DIR=".anti-bluff-reports"
mkdir -p "$REPORT_DIR"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
REPORT_FILE="$REPORT_DIR/scan-${TIMESTAMP}.json"

STEP=0
FAILURES=0

echo "=== Anti-Bluff Scan v2.1.0 ==="
echo "Started: $TIMESTAMP"

# ───────────────────────────────────────────
# STEP 1: Forbidden Pattern Detection
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Forbidden pattern detection..."

FORBIDDEN_PATTERNS=(
    'panic\s*\(\s*"not implemented"\s*\)'
    'return\s+.*fmt\.Errorf\s*\(\s*"not implemented"'
    'return\s+.*errors\.New\s*\(\s*"not implemented"'
    '\bTODO\b.*[^#]'          # TODO without issue reference
    '\bFIXME\b.*[^#]'         # FIXME without issue reference
    '\bXXX\b'                 # XXX marker
    '\bHACK\b'                # HACK marker
    '\btbd\b'                 # standalone TBD
    '\{\s*\}'                 # empty function body
)

STEP1_ISSUES=()
for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.go" --include="*.md" . || true)
    if [[ -n "$matches" ]]; then
        STEP1_ISSUES+=("$matches")
    fi
done

STEP1_COUNT=${#STEP1_ISSUES[@]}
if [[ $STEP1_COUNT -gt 0 ]]; then
    echo "  FAIL: Found $STEP1_COUNT forbidden pattern(s)"
    printf '%s\n' "${STEP1_ISSUES[@]}"
    ((FAILURES++))
else
    echo "  PASS: No forbidden patterns found"
fi

# ───────────────────────────────────────────
# STEP 2: ValidateAntiBluff Verification
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] ValidateAntiBluff() verification..."

# Check that ValidateAntiBluff is called in all challenge/runner code
VALIDATE_CALLS=$(grep -r "ValidateAntiBluff" --include="*.go" . | wc -l)
if [[ $VALIDATE_CALLS -eq 0 ]]; then
    echo "  FAIL: No calls to ValidateAntiBluff found"
    ((FAILURES++))
else
    echo "  PASS: Found $VALIDATE_CALLS call(s) to ValidateAntiBluff"
fi

# Verify unconditional call (no CHALLENGE_ANTIBLUFF_STRICT toggle)
if grep -r "CHALLENGE_ANTIBLUFF_STRICT" --include="*.go" . > /dev/null 2>&1; then
    echo "  FAIL: CHALLENGE_ANTIBLUFF_STRICT toggle still present (must be removed per v2.1.0)"
    ((FAILURES++))
else
    echo "  PASS: No anti-bluff bypass toggles found"
fi

# ───────────────────────────────────────────
# STEP 3: Documentation Anti-Bluff Blocks
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Documentation anti-bluff verification..."

# Check that documentation files contain Anti-Bluff Verification sections
DOC_FILES=$(find . -name "*.md" -not -path "*/vendor/*" -not -path "*/.git/*")
MISSING_BLOCKS=0
for doc in $DOC_FILES; do
    if ! grep -q "Anti-Bluff Verification\|Anti Bluff\|antibluff" "$doc" 2>/dev/null; then
        # Only flag docs in submodule directories, not root-level docs
        if [[ "$doc" == *"/pkg/"* ]] || [[ "$doc" == *"submodule"* ]]; then
            MISSING_BLOCKS=$((MISSING_BLOCKS + 1))
        fi
    fi
done

if [[ $MISSING_BLOCKS -gt 0 ]]; then
    echo "  WARN: $MISSING_BLOCKS submodule doc(s) missing anti-bluff verification section"
else
    echo "  PASS: All submodule docs have anti-bluff verification sections"
fi

# ───────────────────────────────────────────
# STEP 4: Constitution Propagation
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Constitution propagation check..."

# Verify all submodules reference the Constitution
SUBMODULES=$(git submodule status | awk '{print $2}')
MISSING_CONSTITUTION=0
for submod in $SUBMODULES; do
    if [[ -d "$submod" ]]; then
        if ! grep -r "Constitution\|constitution" "$submod" --include="*.md" --include="*.go" > /dev/null 2>&1; then
            MISSING_CONSTITUTION=$((MISSING_CONSTITUTION + 1))
            echo "  WARN: $submod does not reference Constitution"
        fi
    fi
done

if [[ $MISSING_CONSTITUTION -eq 0 ]]; then
    echo "  PASS: All submodules reference Constitution"
fi

# ───────────────────────────────────────────
# STEP 5: Vacuous Assertion Detection
# ───────────────────────────────────────────
((STEP++))
echo "[Step $STEP/5] Vacuous assertion detection..."

VACUOUS_PATTERNS=(
    'assert\.True\s*\(\s*t\s*,\s*true\s*\)'
    'assert\.Equal\s*\(\s*t\s*,\s*true\s*,\s*true\s*\)'
    'assert\.Nil\s*\(\s*t\s*,\s*nil\s*\)'
    'assert\.Equal\s*\(\s*t\s*,\s*[^,]+\s*,\s*\2\s*\)'  # Same var compared to itself
    'assert\.NoError\s*\(\s*t\s*,\s*nil\s*\)'
)

STEP5_ISSUES=()
for pattern in "${VACUOUS_PATTERNS[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*_test.go" . || true)
    if [[ -n "$matches" ]]; then
        STEP5_ISSUES+=("$matches")
    fi
done

STEP5_COUNT=${#STEP5_ISSUES[@]}
if [[ $STEP5_COUNT -gt 0 ]]; then
    echo "  FAIL: Found $STEP5_COUNT vacuous assertion(s)"
    printf '%s\n' "${STEP5_ISSUES[@]}"
    ((FAILURES++))
else
    echo "  PASS: No vacuous assertions found"
fi

# ───────────────────────────────────────────
# SUMMARY
# ───────────────────────────────────────────
echo ""
echo "=== Scan Complete ==="
echo "Failures: $FAILURES"
echo "Report: $REPORT_FILE"

# Generate JSON report
cat > "$REPORT_FILE" <<EOF
{
  "timestamp": "$TIMESTAMP",
  "version": "2.1.0",
  "failures": $FAILURES,
  "steps": {
    "forbidden_patterns": { "status": "$([[ ${#STEP1_ISSUES[@]} -eq 0 ]] && echo "PASS" || echo "FAIL)", "count": ${#STEP1_ISSUES[@]} },
    "validate_antibluff": { "status": "$([[ $VALIDATE_CALLS -gt 0 ]] && echo "PASS" || echo "FAIL")", "calls_found": $VALIDATE_CALLS },
    "documentation_blocks": { "status": "PASS", "missing_count": $MISSING_BLOCKS },
    "constitution_propagation": { "status": "PASS", "missing_count": $MISSING_CONSTITUTION },
    "vacuous_assertions": { "status": "$([[ $STEP5_COUNT -eq 0 ]] && echo "PASS" || echo "FAIL")", "count": $STEP5_COUNT }
  }
}
EOF

exit $FAILURES
```

#### 4.2.2 CI Integration

The anti-bluff scan runs as a **non-overridable CI lane** — it cannot be bypassed, skipped, or configured away:

```yaml
# .github/workflows/anti-bluff.yml (see Section 6.2 for full file)
anti-bluff-scan:
  name: Anti-Bluff Scan
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run anti-bluff scan
      run: bash scripts/anti-bluff-scan.sh
    - name: Upload scan report
      uses: actions/upload-artifact@v4
      with:
        name: anti-bluff-report
        path: .anti-bluff-reports/
```

**Non-Overrideable Guarantee**:
- The workflow uses `on: [pull_request, push]` triggers with no path exclusions
- Branch protection rules require this check to pass before merge
- The check cannot be skipped via `[skip ci]` or commit message patterns
- Only repository administrators can override, and overrides are logged

### 4.3 Host-Integrity Scan

**Location**: `HelixPlay/scripts/claim-check.sh`  
**Constitutional Basis**: R-18 Operational Integrity (Constitution §11.5)  
**Purpose**: Prevent execution of commands that could damage the host system or disrupt CI infrastructure.

#### 4.3.1 R-18 Forbidden Commands List

The following commands and patterns are **forbidden at the highest severity level**. Their presence in any code, script, or configuration causes immediate CI failure:

| Category | Forbidden Patterns | Severity |
|----------|-------------------|----------|
| **Power Management** | `systemctl suspend`, `pm-suspend`, `rtcwake`, `poweroff`, `shutdown`, `halt`, `reboot`, `init 0`, `telinit 0` | **BLOCKER** |
| **Disk Destruction** | `mkfs\..*\s+/dev/`, `dd\s+.*if=.*of=/dev/`, `rm\s+-rf\s+/`, `rm\s+-rf\s+\*/`, `:(){ :|:& };:` (fork bomb) | **BLOCKER** |
| **Disk Partitioning** | `fdisk\s+/dev/`, `parted\s+/dev/`, `gdisk\s+/dev/` | **BLOCKER** |
| **Network Interference** | `iptables\s+-F`, `ip\s+link\s+set.*down`, `ifconfig.*down` | **CRITICAL** |
| **User/Permission Destruction** | `userdel\s+root`, `usermod\s+-L\s+root`, `chmod\s+-R\s+000\s+/` | **BLOCKER** |
| **Container Hazards** | `--privileged` with host path mounts, `/var/run/docker.sock` mounts without justification, `hostPID: true` without justification | **CRITICAL** |
| **Credential Exposure** | `password\s*=\s*["'][^"']+["']` in non-test code, `PRIVATE KEY` without encryption | **BLOCKER** |

#### 4.3.2 Container Hazard Detection

Special rules for container configurations:

```bash
# claim-check.sh — Container hazard detection

check_container_hazards() {
    local failures=0

    # Check for privileged containers with host mounts
    grep -r "privileged: true" --include="*.yml" --include="*.yaml" . | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if grep -A5 -B5 "privileged: true" "$file" | grep -q "volumes:\|volumeMounts:"; then
            echo "BLOCKER: $file has privileged container with volume mounts"
            ((failures++))
        fi
    done

    # Check for hostPID without security context
    grep -r "hostPID: true" --include="*.yml" --include="*.yaml" . | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if ! grep -A10 "hostPID: true" "$file" | grep -q "securityContext:\|runAsNonRoot"; then
            echo "CRITICAL: $file uses hostPID without security context"
            ((failures++))
        fi
    done

    # Check for docker socket mounts
    grep -r "/var/run/docker.sock" --include="*.yml" --include="*.yaml" --include="*.json" . | while read line; do
        echo "CRITICAL: Docker socket mount found: $line"
        ((failures++))
    done

    return $failures
}
```

#### 4.3.3 Operational Integrity Enforcement

The host-integrity scan runs at two points:

1. **Pre-Commit Hook**: `.claude/settings.json` Stop hook runs `claim-check.sh` before every commit from Claude Code
2. **CI Gate**: The `anti-bluff.yml` workflow includes a host-integrity sub-lane

```yaml
  host-integrity:
    name: Host Integrity (R-18)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run claim check
        run: bash scripts/claim-check.sh
      - name: Check container configurations
        run: bash scripts/claim-check.sh --container-scan
```

### 4.4 Mutation Testing

**Tool**: `go-mutesting`  
**Configuration**: `.go-mutesting.yml`  
**Purpose**: Verify that tests actually detect code changes by introducing artificial mutations and confirming tests fail.

#### 4.4.1 go-mutesting Configuration

```yaml
# .go-mutesting.yml
---
timeout: 60s  # Per-mutant timeout

# Mutators to apply
mutators:
  - branch/case       # Mutate case statements in switches
  - branch/if         # Mutate if conditions (negate, remove)
  - expression/remove # Remove sub-expressions
  - statement/remove  # Remove statements
  - numbers/incrementer  # Increment/decrement numeric constants
  - numbers/decrementer

# Exclusions (don't mutate these)
excludes:
  - "vendor/**"
  - "**/*.pb.go"           # Protobuf generated
  - "**/*_mock*.go"        # Mock files
  - "**/tests/**"          # Test code
  - "**/scripts/**"        # Scripts (not Go code but exclude anyway)
  - "**/antibluff*.go"     # Don't mutate the anti-bluff system itself
  - "cmd/**"               # CLI entry points (thin wrappers)

# Score threshold — CI fails if below this
score_threshold: 0.85  # 85% mutation score required
```

#### 4.4.2 Mutation Ratchet Challenge

The `mutation_ratchet_challenge.sh` implements a **ratchet pattern** — mutation score can only increase, never decrease:

```bash
#!/bin/bash
# challenges/scripts/mutation_ratchet_challenge.sh

set -euo pipefail

SCORE_FILE=".mutation-score"
THRESHOLD=85

# Run mutation testing
echo "Running mutation testing..."
go-mutesting --config=.go-mutesting.yml ./... > mutation-report.txt 2>&1

# Parse mutation score from report
SCORE=$(grep "Mutation score" mutation-report.txt | sed 's/.*: \([0-9.]*\)%.*/\1/')
SCORE_INT=${SCORE%.*}

echo "Current mutation score: ${SCORE}%"

# Check absolute threshold
if [[ $SCORE_INT -lt $THRESHOLD ]]; then
    echo "FAIL: Mutation score ${SCORE}% is below threshold ${THRESHOLD}%"
    exit 1
fi

# Check ratchet (score must not decrease)
if [[ -f "$SCORE_FILE" ]]; then
    PREVIOUS=$(cat "$SCORE_FILE")
    if [[ $(echo "$SCORE < $PREVIOUS" | bc -l) -eq 1 ]]; then
        echo "FAIL: Mutation score decreased from ${PREVIOUS}% to ${SCORE}%"
        echo "This is a RATCHET VIOLATION. Fix your tests."
        exit 1
    fi
fi

# Update stored score
echo "$SCORE" > "$SCORE_FILE"
echo "PASS: Mutation score ${SCORE}% meets threshold and ratchet"
```

#### 4.4.3 CI Integration

```yaml
mutation-testing:
  name: Mutation Testing
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Install go-mutesting
      run: go install github.com/zimmski/go-mutesting@latest
    - name: Run mutation tests
      run: bash vasic-digital/Challenges/scripts/mutation_ratchet_challenge.sh
    - name: Upload mutation report
      uses: actions/upload-artifact@v4
      with:
        name: mutation-report
        path: mutation-report.txt
    - name: Update stored score
      run: |
        git add .mutation-score
        git diff --cached --quiet || git commit -m "chore: update mutation score [ci skip]"
```

### 4.5 Negative-Leg Fault Injection

**Constitutional Basis**: §1.3 and §6.3 — *"CI must break each feature and verify non-Unit tests fail"*
**Purpose**: Automatically verify that the test suite actually catches bugs by deliberately introducing defects and confirming test failure.

#### 4.5.1 Automatic Feature Breakage

The negative-leg system introduces controlled defects into the codebase:

| Mutation Type | Description | Example |
|--------------|-------------|---------|
| **Comparison Swap** | Swap comparison operators | `<` becomes `>`, `==` becomes `!=` |
| **Boundary Off-by-One** | Adjust loop bounds by 1 | `i < n` becomes `i <= n` |
| **Return Value Corruption** | Change return values | `return true` becomes `return false` |
| **Error Path Removal** | Remove error checks | `if err != nil { return err }` is deleted |
| **Constant Mutation** | Change numeric constants | `timeout := 30 * time.Second` becomes `timeout := 0` |
| **Branch Removal** | Delete if/else branches | Remove the `else` block entirely |

#### 4.5.2 Verification That Tests Fail

```go
// tests/internal/negativeleg/negative_leg_test.go
package negativeleg

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// TestNegativeLeg_AllFeatures tests that the test suite catches
// deliberate bugs in every production feature.
func TestNegativeLeg_AllFeatures(t *testing.T) {
    if os.Getenv("RUN_NEGATIVE_LEG") != "true" {
        t.Skip("Set RUN_NEGATIVE_LEG=true to run negative-leg fault injection")
    }

    // List all packages with production code
    packages := getProductionPackages(t)

    for _, pkg := range packages {
        t.Run(pkg.Name, func(t *testing.T) {
            // For each mutation point in the package
            for _, mutation := range pkg.MutationPoints {
                t.Run(mutation.Description, func(t *testing.T) {
                    // Apply the mutation
                    backup := mutation.Apply()
                    defer backup.Restore()  // Always restore, even on panic

                    // Run non-Unit tests against the mutated code
                    cmd := exec.Command("go", "test",
                        "-count=1",
                        "-run", "^(TestIntegration|TestE2E|TestChallenge)",
                        "./...",
                    )
                    output, _ := cmd.CombinedOutput()

                    // At least one test MUST fail
                    if !strings.Contains(string(output), "FAIL") {
                        t.Errorf("NEGATIVE LEG FAILED: Mutation '%s' in %s was NOT caught by any test.\n"+
                            "This means the test suite is BLUFFING — it passes even when the feature is broken.\n"+
                            "Output:\n%s", mutation.Description, pkg.Name, output)
                    }
                })
            }
        })
    }
}
```

#### 4.5.3 CI Integration

```yaml
negative-leg:
  name: Negative Leg Fault Injection
  runs-on: ubuntu-latest
  # Run weekly (expensive) and on-demand
  on:
    schedule:
      - cron: '0 2 * * 0'  # Sundays at 2 AM
    workflow_dispatch:
  steps:
    - uses: actions/checkout@v4
    - name: Run negative-leg fault injection
      run: RUN_NEGATIVE_LEG=true go test -count=1 -v -timeout 120m ./tests/internal/negativeleg/...
    - name: Upload negative-leg report
      if: always()
      uses: actions/upload-artifact@v4
      with:
        name: negative-leg-report
        path: tests/internal/negativeleg/report/
```

---

## Section 5: HelixQA Autonomous QA Integration

### 5.1 Visual Assertion Pipeline

**Repository**: `HelixDevelopment/HelixQA`  
**Test Files**: 350 `_test.go` files  
**Purpose**: Automated visual verification of UI features using screenshot capture and OpenCV-based analysis.

#### 5.1.1 Screenshot/Screen Recording Capture

The HelixQA capture system records the application under test using platform-specific capture methods:

```go
// pkg/capture/capture.go — Core capture interface
type Capture interface {
    // CaptureFrame grabs a single frame from the display
    CaptureFrame(ctx context.Context) (*Frame, error)

    // StartRecording begins continuous screen recording
    StartRecording(ctx context.Context, outputPath string) error

    // StopRecording ends the recording
    StopRecording() (*Recording, error)
}

// Frame represents a captured screen frame
type Frame struct {
    Image      image.Image
    Timestamp  time.Time
    Dimensions image.Rectangle
}
```

**Platform-Specific Implementations**:
- **Linux/X11**: Uses `x11grab` (FFmpeg) or XShm for frame capture
- **Linux/Wayland**: Uses PipeWire screencast portal
- **macOS**: Uses `ScreenCaptureKit` framework via CGO
- **Windows**: Uses DXGI Desktop Duplication API
- **Android**: Uses `MediaProjection` API
- **Android TV**: Uses same `MediaProjection` with leanback UI detection

#### 5.1.2 OpenCV-Based Verification

The vision system uses OpenCV (via Go bindings) to analyze captured frames:

```go
// pkg/vision/verifier.go — OpenCV-based visual assertion
type VisualAssertion struct {
    Name        string
    Expected    *ExpectedVisual  // What we expect to see
    Tolerance   float64         // Match tolerance (0.0-1.0)
    ROI         image.Rectangle // Region of interest (optional)
    Timeout     time.Duration
}

type ExpectedVisual struct {
    TemplatePath string        // Reference image to match
    TextContent  string        // Expected text (OCR)
    ColorProfile ColorProfile  // Expected dominant colors
    ElementCount int           // Expected number of UI elements
}

// Verify performs the visual assertion against a captured frame
func (va *VisualAssertion) Verify(frame *capture.Frame) (bool, float64, error) {
    // 1. Template matching (if template provided)
    if va.Expected.TemplatePath != "" {
        matchScore := vision.TemplateMatch(frame.Image, va.Expected.TemplatePath, va.ROI)
        if matchScore < va.Tolerance {
            return false, matchScore, fmt.Errorf("template match failed: %.2f < %.2f",
                matchScore, va.Tolerance)
        }
    }

    // 2. Text recognition (if text expected)
    if va.Expected.TextContent != "" {
        detected := vision.RecognizeText(frame.Image, va.ROI)
        if !strings.Contains(detected, va.Expected.TextContent) {
            return false, 0, fmt.Errorf("text not found: expected '%s'", va.Expected.TextContent)
        }
    }

    // 3. Color analysis (if color profile expected)
    if va.Expected.ColorProfile != nil {
        profile := vision.ExtractColorProfile(frame.Image, va.ROI)
        score := va.Expected.ColorProfile.Compare(profile)
        if score < va.Tolerance {
            return false, score, fmt.Errorf("color profile mismatch: %.2f < %.2f", score, va.Tolerance)
        }
    }

    return true, 1.0, nil
}
```

#### 5.1.3 Production-Equivalent Topology Testing

HelixQA tests run against a **production-equivalent topology** — not mocks, not stubs, but the actual compiled application running in a realistic environment:

```go
// pkg/autonomous/coordinator.go — Test topology setup
func (c *Coordinator) setupProductionEquivalentTopology(ctx context.Context) (*Topology, error) {
    topo := &Topology{
        Services: map[string]*Service{
            "core-backend": {
                Image:    "helixplay/core:latest",
                Env:      productionEnv(),
                Networks: []string{"helix-backend"},
            },
            "host-agent": {
                Image:    "helixplay/host-agent:latest",
                Env:      productionEnv(),
                Networks: []string{"helix-backend", "helix-host"},
                Devices:  []string{"/dev/dri"},  // GPU passthrough
            },
            "client-web": {
                Image:    "helixplay/client-web:latest",
                Env:      productionEnv(),
                Networks: []string{"helix-backend"},
            },
        },
    }

    if err := topo.Start(ctx); err != nil {
        return nil, err
    }

    return topo, nil
}
```

### 5.2 Integration with Challenges

#### 5.2.1 Challenge Runner Coordination

HelixQA integrates with the Challenges repository at the runner level:

```go
// pkg/autonomous/challenge_adapter.go — Bridges HelixQA to Challenges framework

type QAChallengeAdapter struct {
    qaRunner    *helixqa.Runner
    challengeRunner *challenge.Runner
}

func (a *QAChallengeAdapter) ExecuteQAChallenge(ctx context.Context, def QAChallengeDefinition) (*challenge.Result, error) {
    // 1. Set up production-equivalent topology
    topo, err := a.qaRunner.SetupTopology(ctx, def.Topology)
    if err != nil {
        return nil, err
    }
    defer topo.Teardown()

    // 2. Execute the user flow
    flowResult, err := a.qaRunner.ExecuteFlow(ctx, def.UserFlow)
    if err != nil {
        return nil, err
    }

    // 3. Perform visual assertions
    visualResults, err := a.qaRunner.PerformVisualAssertions(ctx, def.VisualAssertions)
    if err != nil {
        return nil, err
    }

    // 4. Build challenge result with anti-bluff evidence
    result := &challenge.Result{
        Status: challenge.StatusPassed,
        RecordedActions: flowResult.Actions,
        Assertions: make([]challenge.Assertion, 0, len(visualResults)),
    }

    for _, vr := range visualResults {
        result.Assertions = append(result.Assertions, challenge.Assertion{
            Name:   vr.AssertionName,
            Passed: vr.Passed,
            Details: map[string]interface{}{
                "match_score": vr.MatchScore,
                "screenshot":  vr.ScreenshotPath,
            },
        })
    }

    // 5. Unconditional anti-bluff validation
    if err := challenge.ValidateAntiBluff(result); err != nil {
        result.Status = challenge.StatusError
        return result, err
    }

    return result, nil
}
```

#### 5.2.2 Evidence Collection

Every HelixQA challenge produces three forms of evidence:

1. **Screenshots**: Captured at each assertion point, stored as PNG artifacts
2. **Screen Recordings**: Full session recordings for manual review if needed
3. **Assertion Logs**: Structured logs with timestamps, match scores, and pass/fail status

#### 5.2.3 Pass/Fail Criteria

A HelixQA challenge passes only when:

1. All user flow steps execute without error
2. All visual assertions pass (match score >= tolerance)
3. `ValidateAntiBluff()` returns `nil` (has recorded actions AND at least one passing assertion)
4. No critical log errors during execution
5. Performance metrics within SLA (e.g., UI response time < 200ms)

---

## Section 6: CI/CD Pipeline Design

### 6.1 Current Gap Analysis

**CRITICAL FINDING**: Despite having 550+ test files, 8 challenge scripts, a mutation testing framework, and a constitutional mandate for comprehensive testing — **zero GitHub Actions workflows exist** in any repository.

#### 6.1.1 What Exists Now (Manual/Local Only)

| Component | Status | Runs Where | Gaps |
|-----------|--------|-----------|------|
| `go test ./...` | Local only | Developer machine | No visibility, no enforcement |
| `make anti-bluff` | Local only | Developer machine | Can be forgotten, no audit trail |
| `make challenge` | Local only | Developer machine | Same issues |
| `go-mutesting` | Local only | Developer machine | Same issues |
| `scripts/anti-bluff-scan.sh` | Local only | Developer machine | Same issues |
| Quality gates (SonarQube, Snyk) | Manual | External dashboards | Not tied to PR workflow |

#### 6.1.2 Impact of No CI/CD

1. **No visibility**: No one knows if tests are currently passing across all 29 submodules
2. **No enforcement**: Quality gates are voluntary; developers can (and do) skip them
3. **No audit trail**: There's no record of which quality checks ran on which commit
4. **Cannot detect regressions**: A commit that breaks tests can be merged undetected
5. **Blocks external contribution**: Contributors have no way to verify their changes
6. **Blocks automated deployment**: Deploying without verified quality gates violates the Constitution

#### 6.1.3 Dependency Resolution Issues

The `Makefile` explicitly notes:

> *"`vet` and `test` are intentionally NOT included in `qa-all` because several packages depend on missing replace-directives (`../Dependencies/HelixDevelopment/*`) that aren't present in a clean checkout."*

This means:
- `go test ./...` fails on a clean checkout
- `go vet ./...` fails on a clean checkout
- CI cannot run these commands without first fixing the dependency resolution

### 6.2 Required CI/CD Workflows

#### 6.2.1 `.github/workflows/ci.yml` — Full Pipeline

```yaml
name: CI — Full Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

# Prevent redundant runs
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  # ───────────────────────────────────────────
  # JOB 1: Pre-Flight Checks
  # ───────────────────────────────────────────
  preflight:
    name: Pre-Flight
    runs-on: ubuntu-latest
    outputs:
      go_version: "1.26.2"
      changed_modules: ${{ steps.changes.outputs.modules }}
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
          fetch-depth: 0  # Full history for change detection

      - name: Detect changed modules
        id: changes
        run: |
          CHANGED=$(git diff --name-only origin/${{ github.base_ref }} HEAD | \
            grep -E "^vasic-digital/|^cmd/|^pkg/|^tests/" | \
            cut -d/ -f2 | sort -u | jq -R -s -c 'split("\n")[:-1]')
          echo "modules=$CHANGED" >> $GITHUB_OUTPUT

      - name: Submodule integrity check
        run: python scripts/verify-submodules.py

  # ───────────────────────────────────────────
  # JOB 2: Build Matrix (Go versions)
  # ───────────────────────────────────────────
  build:
    name: Build (Go ${{ matrix.go }})
    runs-on: ubuntu-latest
    needs: [preflight]
    strategy:
      matrix:
        go: ['1.26.2']
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}

      - name: Fix replace directives
        run: bash scripts/fix-replace.sh

      - name: Download dependencies
        run: go mod download

      - name: Build all commands
        run: go build ./cmd/...

      - name: Build all submodules
        run: |
          for mod in vasic-digital/*/; do
            if [[ -f "$mod/go.mod" ]]; then
              echo "Building $mod..."
              (cd "$mod" && go build ./...)
            fi
          done

  # ───────────────────────────────────────────
  # JOB 3: Unit Tests (all submodules)
  # ───────────────────────────────────────────
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: [build]
    container: golang:1.26
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Fix replace directives
        run: bash scripts/fix-replace.sh

      - name: Run unit tests
        run: go test -count=1 -race -p 1 ./...
        timeout-minutes: 30

      - name: Run unit tests (submodules)
        run: |
          FAILURES=0
          for mod in vasic-digital/*/; do
            if [[ -f "$mod/go.mod" ]]; then
              echo "Testing $mod..."
              if ! (cd "$mod" && go test -count=1 -race -p 1 ./...); then
                FAILURES=$((FAILURES + 1))
              fi
            fi
          done
          if [[ $FAILURES -gt 0 ]]; then
            echo "$FAILURES submodule test suite(s) failed"
            exit 1
          fi
        timeout-minutes: 60

      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: |
            coverage.out
            coverage.html

  # ───────────────────────────────────────────
  # JOB 4: Integration Tests
  # ───────────────────────────────────────────
  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: [build]
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: helixtest
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports: ['25432:5432']
      redis:
        image: redis:7-alpine
        ports: ['26379:6379']
      nats:
        image: nats:2-alpine
        ports: ['4222:4222']
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Fix replace directives
        run: bash scripts/fix-replace.sh

      - name: Run integration tests
        run: go test -count=1 -race -p 1 -tags=integration ./tests/integration/...
        env:
          HELIX_TEST_DB_DSN: postgres://postgres:test@localhost:25432/helixtest?sslmode=disable
          HELIX_TEST_REDIS_ADDR: localhost:26379
          HELIX_TEST_NATS_URL: nats://localhost:4222
        timeout-minutes: 30

  # ───────────────────────────────────────────
  # JOB 5: E2E Tests
  # ───────────────────────────────────────────
  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest
    needs: [integration-tests]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build E2E topology
        run: docker compose -f tests/e2e/docker-compose.yml build

      - name: Run E2E tests
        run: go test -count=1 -v -timeout 30m ./tests/e2e/...
        timeout-minutes: 35

      - name: Collect logs on failure
        if: failure()
        run: docker compose -f tests/e2e/docker-compose.yml logs > e2e-logs.txt 2>&1

      - name: Upload E2E artifacts
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-failure-artifacts
          path: |
            e2e-logs.txt
            tests/e2e/screenshots/

  # ───────────────────────────────────────────
  # JOB 6: Security Scans
  # ───────────────────────────────────────────
  security:
    name: Security
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Run Semgrep SAST
        uses: semgrep/semgrep-action@v1
        with:
          config: >-
            p/security-audit
            p/owasp-top-ten
            p/cwe-top-25
            p/gosec

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          exit-code: '1'
          severity: 'HIGH,CRITICAL'

      - name: Run gitleaks secret detection
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  # ───────────────────────────────────────────
  # JOB 7: Code Quality
  # ───────────────────────────────────────────
  quality:
    name: Code Quality
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Run go vet
        run: |
          bash scripts/fix-replace.sh
          go vet ./...

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=10m

      - name: Run gofmt check
        run: |
          UNFORMATTED=$(gofmt -l .)
          if [[ -n "$UNFORMATTED" ]]; then
            echo "The following files need formatting:"
            echo "$UNFORMATTED"
            exit 1
          fi

      - name: SonarQube Scan
        uses: sonarqube-quality-gate-action@master
        env:
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
          SONAR_HOST_URL: ${{ secrets.SONAR_HOST_URL }}
        with:
          scanMetadataReportFile: .scannerwork/report-task.txt

  # ───────────────────────────────────────────
  # JOB 8: Benchmarks (no fail on regression, just report)
  # ───────────────────────────────────────────
  benchmarks:
    name: Benchmarks
    runs-on: ubuntu-latest
    needs: [unit-tests]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Run benchmarks
        run: go test -bench=. -benchmem -benchtime=5s ./tests/benchmark/... | tee benchmark.txt

      - name: Upload benchmark results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: benchmark.txt

      - name: Comment benchmark results
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: benchmark.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          comment-on-alert: true
          alert-threshold: '150%'
```

#### 6.2.2 `.github/workflows/anti-bluff.yml` — Non-Overridable Scan

```yaml
name: Anti-Bluff — Non-Overridable

# This workflow CANNOT be skipped via [skip ci] or any other mechanism.
# It is the constitutional guarantee against bluff testing.

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

# This workflow always runs, regardless of what files changed
# NO path exclusions — every change is subject to anti-bluff verification

jobs:
  # ───────────────────────────────────────────
  # JOB 1: Vacuous Assertion Scan
  # ───────────────────────────────────────────
  vacuous-scan:
    name: Vacuous Assertion Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Run anti-bluff scan (full)
        run: bash scripts/anti-bluff-scan.sh
        timeout-minutes: 10

      - name: Upload scan report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: anti-bluff-report
          path: .anti-bluff-reports/

  # ───────────────────────────────────────────
  # JOB 2: Host Integrity (R-18)
  # ───────────────────────────────────────────
  host-integrity:
    name: Host Integrity (R-18)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Run claim check
        run: bash scripts/claim-check.sh

      - name: Container hazard scan
        run: bash scripts/claim-check.sh --container-scan

      - name: Verify no forbidden commands in codebase
        run: |
          FORBIDDEN=(
            'systemctl\s+suspend'
            'pm-suspend'
            'shutdown\s+'
            'poweroff'
            'reboot\s+'
            'mkfs\.'
            'rm\s+-rf\s+/\s'
            ':\(\)\{\s*:\|\s*:\s*\&\s*\};\s*:'
          )
          FAILURES=0
          for pattern in "${FORBIDDEN[@]}"; do
            if grep -r -P "$pattern" --include="*.go" --include="*.sh" --include="*.yml" .; then
              echo "FORBIDDEN COMMAND FOUND: $pattern"
              FAILURES=$((FAILURES + 1))
            fi
          done
          exit $FAILURES

  # ───────────────────────────────────────────
  # JOB 3: Constitution Propagation
  # ───────────────────────────────────────────
  constitution-check:
    name: Constitution Propagation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Run constitution propagation check
        run: bash scripts/propagate-constitution.sh --verify-only

      - name: Verify Constitution v2.1.0 references
        run: |
          # Check that the Constitution is referenced in key files
          for file in CLAUDE.md AGENTS.md; do
            if [[ -f "$file" ]]; then
              if ! grep -q "Constitution" "$file"; then
                echo "FAIL: $file does not reference Constitution"
                exit 1
              fi
            fi
          done
          echo "Constitution references verified"
```

#### 6.2.3 `.github/workflows/challenges.yml` — Challenge Execution

```yaml
name: Challenges

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
  schedule:
    - cron: '0 3 * * *'  # Daily at 3 AM

jobs:
  # ───────────────────────────────────────────
  # JOB 1: Build Challenge Runner
  # ───────────────────────────────────────────
  build-runner:
    name: Build Challenge Runner
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Build challenge runner
        run: cd vasic-digital/Challenges && go build -o challenge-runner ./cmd/userflow-runner

      - name: Build all challenge scripts
        run: |
          cd vasic-digital/Challenges/scripts
          chmod +x *.sh

      - name: Upload runner artifact
        uses: actions/upload-artifact@v4
        with:
          name: challenge-runner
          path: vasic-digital/Challenges/challenge-runner

  # ───────────────────────────────────────────
  # JOB 2: Run All Challenge Scripts
  # ───────────────────────────────────────────
  challenge-scripts:
    name: Challenge Scripts
    runs-on: ubuntu-latest
    needs: [build-runner]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Download runner
        uses: actions/download-artifact@v4
        with:
          name: challenge-runner
          path: ./challenge-runner

      - name: Make runner executable
        run: chmod +x ./challenge-runner/challenge-runner

      - name: Run anchor manifest challenge
        run: bash vasic-digital/Challenges/scripts/anchor_manifest_challenge.sh

      - name: Run bluff scanner challenge
        run: bash vasic-digital/Challenges/scripts/bluff_scanner_challenge.sh

      - name: Run compile challenge
        run: bash vasic-digital/Challenges/scripts/challenges_compile_challenge.sh

      - name: Run functionality challenge
        run: bash vasic-digital/Challenges/scripts/challenges_functionality_challenge.sh

      - name: Run unit challenge
        run: bash vasic-digital/Challenges/scripts/challenges_unit_challenge.sh

      - name: Run host no-auto-suspend challenge
        run: bash vasic-digital/Challenges/scripts/host_no_auto_suspend_challenge.sh

      - name: Run mutation ratchet challenge
        run: bash vasic-digital/Challenges/scripts/mutation_ratchet_challenge.sh

      - name: Run no-suspend-calls challenge
        run: bash vasic-digital/Challenges/scripts/no_suspend_calls_challenge.sh

      - name: Collect challenge reports
        if: always()
        run: |
          mkdir -p challenge-reports
          cp -r vasic-digital/Challenges/scripts/reports/* challenge-reports/ 2>/dev/null || true

      - name: Upload challenge reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: challenge-reports
          path: challenge-reports/

  # ───────────────────────────────────────────
  # JOB 3: Run Challenge Banks
  # ───────────────────────────────────────────
  challenge-banks:
    name: Challenge Banks
    runs-on: ubuntu-latest
    needs: [build-runner]
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Download runner
        uses: actions/download-artifact@v4
        with:
          name: challenge-runner
          path: ./challenge-runner

      - name: Make runner executable
        run: chmod +x ./challenge-runner/challenge-runner

      - name: Run all challenge banks
        run: |
          ./challenge-runner/challenge-runner \
            --banks-dir vasic-digital/Challenges/banks \
            --all \
            --report challenge-bank-report.json
        timeout-minutes: 60

      - name: Verify anti-bluff validation in report
        run: |
          if ! jq '.antiBluffValid' challenge-bank-report.json | grep -q true; then
            echo "ANTI-BLUFF VALIDATION FAILED: Some challenges passed without evidence"
            exit 1
          fi

      - name: Upload bank report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: challenge-bank-report
          path: challenge-bank-report.json
```

#### 6.2.4 `.github/workflows/mutation.yml` — Mutation Testing

```yaml
name: Mutation Testing

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 4 * * 0'  # Weekly on Sunday at 4 AM

jobs:
  mutation:
    name: Mutation Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.26.2

      - name: Install go-mutesting
        run: go install github.com/zimmski/go-mutesting@latest

      - name: Run mutation ratchet challenge
        run: bash vasic-digital/Challenges/scripts/mutation_ratchet_challenge.sh
        timeout-minutes: 120

      - name: Upload mutation report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: mutation-report
          path: |
            mutation-report.txt
            .mutation-score
```

#### 6.2.5 Container-Based Job Definitions

All CI jobs should eventually run in containers matching the production environment:

```yaml
# Example container-based job
container-test:
  runs-on: ubuntu-latest
  container:
    image: golang:1.26.2-alpine
    options: --cap-add=NET_ADMIN  # For network chaos tests
  services:
    helix-core:
      image: helixplay/core:test
      env:
        HELIX_ENV: test
        HELIX_DB_DSN: postgres://test@postgres/helix
    helix-host:
      image: helixplay/host-agent:test
      devices:
        - /dev/dri:/dev/dri  # GPU passthrough for encode tests
```

#### 6.2.6 Matrix Builds Across Go Versions

```yaml
strategy:
  matrix:
    go: ['1.25', '1.26.2']
    os: ['ubuntu-latest', 'ubuntu-24.04']
    include:
      - go: '1.26.2'
        os: 'ubuntu-latest'
        primary: true
    exclude:
      - go: '1.25'
        os: 'ubuntu-24.04'  # Reduce matrix size
```

#### 6.2.7 Artifact Collection and Retention

```yaml
# Global artifact retention policy
# Stored in .github/artifact-retention.yml
artifact-retention:
  coverage-reports:
    retention-days: 90
    compress: true
  challenge-reports:
    retention-days: 180
    compress: true
  failure-artifacts:
    retention-days: 30
    compress: true
  benchmark-results:
    retention-days: 365
    compress: false  # For trend analysis
```

### 6.3 Quality Gates

#### 6.3.1 SonarQube Configuration

```properties
# sonar-project.properties
sonar.projectKey=HelixPlay
sonar.projectName=HelixPlay Cloud Gaming Platform
sonar.sources=.
sonar.exclusions=vendor/**,**/*.pb.go,**/tests/**,**/scripts/**
sonar.tests=.
sonar.test.inclusions=**/*_test.go,**/tests/**
sonar.go.coverage.reportPaths=coverage.out
sonar.coverage.exclusions=cmd/**,**/mocks/**

# Quality gate thresholds (Constitution §7)
sonar.qualitygate.wait=true
sonar.coverage.minimum=100
sonar.duplications.minimum=3
```

**Gate Requirements**:
| Metric | Threshold | Action on Fail |
|--------|-----------|----------------|
| Coverage | >= 100% | Block merge |
| Duplications | <= 3% | Block merge |
| Code Smells | 0 (critical/blocker) | Block merge |
| Bugs | 0 | Block merge |
| Vulnerabilities | 0 | Block merge |
| Security Hotspots | 0 (high) | Block merge |

#### 6.3.2 Snyk Dependency Scanning

```yaml
snyk:
  name: Snyk Security Scan
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run Snyk
      uses: snyk/actions/golang@master
      env:
        SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
      with:
        args: --severity-threshold=high --fail-on=upgradable
```

#### 6.3.3 Semgrep Rule Configuration

```yaml
semgrep:
  name: Semgrep SAST
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run Semgrep
      uses: semgrep/semgrep-action@v1
      with:
        config: >-
          p/security-audit
          p/owasp-top-ten
          p/cwe-top-25
          p/gosec
          p/golang
          p/trailofbits
```

#### 6.3.4 Trivy Container Scanning

```yaml
trivy:
  name: Trivy Container Scan
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Build container image
      run: docker build -t helixplay:test .
    - name: Scan image
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: helixplay:test
        format: 'sarif'
        output: 'trivy-results.sarif'
        severity: 'HIGH,CRITICAL'
    - name: Upload results
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: trivy-results.sarif
```

#### 6.3.5 gitleaks Secret Scanning

```yaml
gitleaks:
  name: Secret Detection
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0  # Full history for secret detection
    - uses: gitleaks/gitleaks-action@v2
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        GITLEAKS_ENABLE_COMMENTS: true
```

#### 6.3.6 govulncheck Go Vulnerability Scanning

```yaml
govulncheck:
  name: Go Vulnerability Check
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: 1.26.2
    - run: go install golang.org/x/vuln/cmd/govulncheck@latest
    - run: govulncheck ./...
```

#### 6.3.7 Coverage Thresholds

| Metric | Threshold | Enforcement |
|--------|-----------|-------------|
| Line Coverage | 100% | SonarQube gate + CI fail |
| Branch Coverage | 100% | SonarQube gate + CI fail |
| Function Coverage | 100% | SonarQube gate + CI fail |
| Package Coverage | 100% | CI fail if any package < 100% |
| Mutation Score | >= 85% | `mutation_ratchet_challenge.sh` |

#### 6.3.8 Benchmark Regression Detection

```yaml
benchmark-regression:
  name: Benchmark Regression
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: 1.26.2
    - name: Run benchmarks
      run: go test -bench=. -benchmem -count=5 ./tests/benchmark/... | tee benchmark.txt
    - name: Compare with baseline
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'go'
        output-file-path: benchmark.txt
        external-data-json-path: ./cache/benchmark-data.json
        github-token: ${{ secrets.GITHUB_TOKEN }}
        alert-threshold: '150%'
        comment-on-alert: true
        fail-on-alert: true
        auto-push: true
```

---

## Section 7: Fixing Current Issues

### 7.1 Dependency Resolution

#### 7.1.1 Root Cause

The `go.mod` in the main repository and several submodules contain `replace` directives that reference paths outside the repository:

```go
// Current problematic replace directive
replace github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory => ./vasic-digital/Memory

// Hypothetical problematic directive (from Makefile note)
replace github.com/HelixDevelopment/Dependencies/SomeModule => ../Dependencies/HelixDevelopment/SomeModule
```

The `../Dependencies/` path does not exist in a clean checkout because:
1. It is a sibling directory, not a submodule
2. It is not cloned by `git submodule update --init --recursive`
3. It may be a private repository or local development dependency

#### 7.1.2 Fix: Replace Directives

**Strategy 1**: Convert sibling dependencies to proper Git submodules:

```bash
# scripts/fix-replace.sh — Dependency resolution fix
#!/bin/bash
set -euo pipefail

# Check if Dependencies directory exists
if [[ ! -d "../Dependencies" ]]; then
    echo "Dependencies directory not found. Using submodule fallback."

    # For each broken replace directive, try to resolve via submodule
    for modfile in $(find . -name "go.mod" -not -path "*/vendor/*"); do
        dir=$(dirname "$modfile")
        # Replace ../Dependencies references with proper module paths
        sed -i 's|=>\s*\.\./Dependencies/HelixDevelopment/\(.*\)|=> ./vasic-digital/\1|g' "$modfile" 2>/dev/null || true
    done
fi

# Verify all replace directives resolve
for modfile in $(find . -name "go.mod" -not -path "*/vendor/*"); do
    dir=$(dirname "$modfile")
    (cd "$dir" && go mod verify 2>/dev/null) || {
        echo "WARNING: $modfile has unresolvable replace directives"
    }
done

echo "Replace directive fix complete"
```

**Strategy 2**: Use `go.work` workspace configuration:

```go
// go.work — Go workspace for multi-module development
go 1.26.2

use (
    .
    ./vasic-digital/Auth
    ./vasic-digital/Cache
    ./vasic-digital/Challenges
    ./vasic-digital/Database
    ./vasic-digital/Discovery
    ./vasic-digital/EventBus
    ./vasic-digital/Formatters
    ./vasic-digital/Media
    ./vasic-digital/Memory
    ./vasic-digital/Messaging
    ./vasic-digital/Middleware
    ./vasic-digital/Observability
    ./vasic-digital/Plugins
    ./vasic-digital/RAG
    ./vasic-digital/RateLimiter
    ./vasic-digital/Recovery
    ./vasic-digital/Security
    ./vasic-digital/Storage
    ./vasic-digital/Streaming
    ./vasic-digital/VectorDB
)

// External dependencies that must be cloned separately
// These are NOT in the workspace and must be fetched via go mod download
replace github.com/HelixDevelopment/HelixQA => ./HelixQA
```

#### 7.1.3 Ensure Clean Checkout Builds

```yaml
# CI verification step
clean-checkout-test:
  name: Clean Checkout Build
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        submodules: recursive

    - name: Fresh clone — no cached state
      run: |
        # Remove any cached module state
        rm -rf ~/go/pkg/mod/cache
        go clean -cache -modcache

    - name: Fix replace directives
      run: bash scripts/fix-replace.sh

    - name: Download all dependencies
      run: |
        go mod download
        go work sync 2>/dev/null || true

    - name: Build everything
      run: |
        go build ./cmd/...
        for mod in vasic-digital/*/; do
          if [[ -f "$mod/go.mod" ]]; then
            (cd "$mod" && go build ./...)
          fi
        done

    - name: Run vet
      run: go vet ./...

    - name: Run tests
      run: go test -count=1 ./...
```

#### 7.1.4 go.work Workspace Configuration

The `go.work` file should be:
1. **Committed to the repository** so all developers use the same workspace
2. **Used by CI** so CI builds match local builds
3. **Kept in sync** with `.gitmodules` so every submodule in git is in the workspace

```bash
# scripts/sync-workspace.sh — Keep go.work in sync with submodules
#!/bin/bash
set -euo pipefail

WORK_FILE="go.work"
SUBMODULES=$(git submodule status | awk '{print $2}')

echo "go 1.26.2" > "$WORK_FILE"
echo "" >> "$WORK_FILE"
echo "use (" >> "$WORK_FILE"
echo "    ." >> "$WORK_FILE"

for submod in $SUBMODULES; do
    # Only include Go modules
    if [[ -f "$submod/go.mod" ]]; then
        echo "    ./$submod" >> "$WORK_FILE"
    fi
done

echo ")" >> "$WORK_FILE"

echo "go.work synced with $(echo "$SUBMODULES" | wc -w) submodules"
```

### 7.2 Test Gaps

#### 7.2.1 Missing E2E Tests

The following submodules have **zero E2E tests** and need them urgently:

| Submodule | Why E2E is Critical | Suggested E2E Scenario | Effort |
|-----------|-------------------|----------------------|--------|
| Cache | Cache invalidation must work end-to-end | Write through cache, invalidate, verify miss | 2 days |
| EventBus | Event delivery is core to system operation | Publish events across services, verify delivery | 2 days |
| Messaging | Message queue reliability | Produce/consume messages, verify ordering, test failure recovery | 3 days |
| Recovery | Fault recovery must work in production topology | Inject fault, verify automatic recovery | 3 days |
| client-wails | Desktop client is primary user interface | Full user journey from discovery to gameplay | 5 days |
| core-backend | Core backend is the system backbone | Full request lifecycle through all backend services | 4 days |

#### 7.2.2 Missing Integration Between Submodules

Current tests are mostly **intra-module** (testing within a single submodule). The following **inter-module** integration paths are untested:

| Integration Path | Components Involved | Test Scenario | Priority |
|-----------------|-------------------|--------------|----------|
| Auth -> Database | Auth, Database | User registration writes to DB; login reads from DB | P1 |
| Auth -> Middleware | Auth, Middleware | Authenticated requests are properly authorized | P1 |
| Discovery -> Core | Discovery, Core Backend | Host discovery results are available via core API | P1 |
| Streaming -> Storage | Streaming, Storage | Stream recordings are persisted to storage | P1 |
| RateLimiter -> Middleware | RateLimiter, Middleware | Rate limiting is applied to HTTP routes | P2 |
| RAG -> VectorDB | RAG, VectorDB | Document embedding and retrieval | P2 |
| Media -> Streaming | Media, Streaming | Media transcoding in streaming pipeline | P2 |

#### 7.2.3 Constructor-Only Tests Needing Behavior Verification

The following test files were identified as having constructor-only patterns that need behavior verification added:

| Test File | Current Pattern | Needed Addition | Effort |
|-----------|---------------|-----------------|--------|
| `pkg/autonomous/real_executor_test.go` | `require.NotNil(exec)` | Execute a command on each platform executor, verify output | 2 days |
| Various submodule tests | `require.NoError(err)` after `NewX()` | Call at least one method on the constructed object | 1 day per submodule |

### 7.3 CI Migration Plan

#### 7.3.1 Step-by-Step Migration

**Phase 1: Foundation (Week 1)**

| Day | Task | Verification |
|-----|------|-------------|
| 1 | Create `.github/workflows/` directory in all 3 repos | Directory exists and is committed |
| 1 | Implement `fix-replace.sh` script | Clean checkout builds successfully |
| 2 | Create `go.work` workspace file | `go work sync` succeeds |
| 2 | Create `.github/workflows/ci.yml` — basic build only | Build job passes on PR |
| 3 | Add unit test job to `ci.yml` | Unit test job passes, reports coverage |
| 3 | Add `golangci-lint` and `go vet` jobs | Lint job passes |
| 4 | Create `.github/workflows/anti-bluff.yml` | Anti-bluff scan runs on every PR |
| 4 | Set branch protection rules (require anti-bluff) | Cannot merge without anti-bluff pass |
| 5 | Test full pipeline on a PR | All jobs pass, merge blocked on failure |

**Phase 2: Quality Gates (Week 2)**

| Day | Task | Verification |
|-----|------|-------------|
| 6 | Add security scan job (govulncheck, Semgrep, Trivy, gitleaks) | Security job passes, detects test vulnerability |
| 7 | Add SonarQube integration | Coverage appears in SonarQube dashboard |
| 8 | Add Snyk dependency scanning | Snyk report shows 0 high-severity issues |
| 9 | Add benchmark job | Benchmark results are collected and stored |
| 10 | Create `.github/workflows/mutation.yml` | Mutation ratchet runs weekly |

**Phase 3: Integration & E2E (Week 3)**

| Day | Task | Verification |
|-----|------|-------------|
| 11 | Add integration test job with test services | Integration tests pass against real DB/cache |
| 12 | Create E2E Docker Compose topology | `docker compose up` starts all services |
| 13 | Add E2E test job | E2E tests pass in CI |
| 14 | Create `.github/workflows/challenges.yml` | All 8 challenge scripts pass |
| 15 | Add challenge bank execution | Challenge bank report shows antiBluffValid=true |

**Phase 4: Advanced Testing (Week 4)**

| Day | Task | Verification |
|-----|------|-------------|
| 16 | Add chaos test job | Chaos tests detect simulated container failures |
| 17 | Add stress test job | 4-hour stress profile completes, no goroutine leaks |
| 18 | Add smoke test job (post-deploy) | Smoke tests verify staging after deployment |
| 19 | Add full automation job | Full deployment test completes without manual steps |
| 20 | Add negative-leg fault injection job | Fault injection detects missing test coverage |

#### 7.3.2 Priority Order for Workflow Implementation

```
1. anti-bluff.yml     (MUST be first — constitutional requirement)
2. ci.yml (basic)     (build + unit tests)
3. ci.yml (extended)  (+ security + quality gates)
4. mutation.yml       (weekly mutation testing)
5. challenges.yml     (challenge execution)
6. ci.yml (full)      (+ integration + E2E + chaos + stress)
```

#### 7.3.3 Testing Each Workflow Before Merge

Every workflow must pass a **self-test** before being merged:

1. **Workflow syntax validation**: `actionlint` or GitHub's workflow editor
2. **Dry run**: Use `act` (local GitHub Actions runner) to test the workflow locally
3. **Intentional failure test**: Introduce a known defect and verify the workflow catches it
4. **Intentional success test**: Fix the defect and verify the workflow passes
5. **Timeout test**: Verify jobs timeout correctly rather than hanging indefinitely

```bash
# Self-test script for workflows
#!/bin/bash
# scripts/test-workflow.sh

WORKFLOW=$1

echo "Testing workflow: $WORKFLOW"

# 1. Syntax validation
if ! command -v actionlint &> /dev/null; then
    go install github.com/rhysd/actionlint/cmd/actionlint@latest
fi
actionlint "$WORKFLOW"

# 2. Dry run with act (if available)
if command -v act &> /dev/null; then
    act -W "$WORKFLOW" --dry-run
fi

echo "Workflow $WORKFLOW passed self-test"
```

---

## Section 8: Challenges Implementation Plan

### 8.1 Challenge Scripts to Implement

All 8 challenge scripts live in `vasic-digital/Challenges/scripts/` and are executed via `make challenge` or the CI `challenges.yml` workflow.

#### 8.1.1 `anchor_manifest_challenge.sh` — Behavior-Anchor Manifest Validator

**Purpose**: Verify that every submodule has a behavior-anchor manifest — a YAML file that documents the observable behaviors the module guarantees and the tests that verify them.

```bash
#!/bin/bash
# challenges/scripts/anchor_manifest_challenge.sh

set -euo pipefail

MANIFEST_FILE=".behavior-anchors.yml"
FAILURES=0

# Check that manifest exists for every submodule
for submod in vasic-digital/*/; do
    if [[ ! -f "$submod/$MANIFEST_FILE" ]]; then
        echo "FAIL: $submod missing behavior-anchor manifest"
        ((FAILURES++))
        continue
    fi

    # Validate manifest structure
    if ! yq eval '.anchors' "$submod/$MANIFEST_FILE" > /dev/null 2>&1; then
        echo "FAIL: $submod manifest missing 'anchors' section"
        ((FAILURES++))
        continue
    fi

    # Verify each anchor has a test reference
    anchor_count=$(yq eval '.anchors | length' "$submod/$MANIFEST_FILE")
    for ((i=0; i<anchor_count; i++)); do
        anchor_name=$(yq eval ".anchors[$i].name" "$submod/$MANIFEST_FILE")
        test_ref=$(yq eval ".anchors[$i].verified_by_test" "$submod/$MANIFEST_FILE")

        if [[ -z "$test_ref" || "$test_ref" == "null" ]]; then
            echo "FAIL: Anchor '$anchor_name' in $submod has no test reference"
            ((FAILURES++))
        elif [[ ! -f "$submod/$test_ref" ]]; then
            echo "FAIL: Anchor '$anchor_name' references missing test: $test_ref"
            ((FAILURES++))
        fi
    done
done

if [[ $FAILURES -gt 0 ]]; then
    echo "Anchor manifest challenge FAILED: $FAILURES issue(s)"
    exit 1
fi

echo "Anchor manifest challenge PASSED"
```

#### 8.1.2 `bluff_scanner_challenge.sh` — Two-Phase Bluff Scanner

**Purpose**: Verify the bluff scanner itself works correctly (Phase 1: self-test), then run it against the full tree (Phase 2).

```bash
#!/bin/bash
# challenges/scripts/bluff_scanner_challenge.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCANNER="$SCRIPT_DIR/anti-bluff/bluff-scanner.sh"
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# ─── PHASE 1: Scanner Self-Test ───
echo "Phase 1: Scanner self-test..."

# Create hand-crafted bluff fixtures that the scanner MUST detect
cat > "$TEMP_DIR/bluff_fixture_test.go" <<'EOF'
package bluff

import "testing"

// This is a DELIBERATELY bluff test — scanner must detect it
func TestBluffFixture_TrueIsTrue(t *testing.T) {
    assert.True(t, true)           // Vacuous: always passes
    assert.Equal(t, true, true)    // Vacuous: same value
    assert.Nil(t, nil)             // Vacuous: nil is nil
}

func TestBluffFixture_Empty(t *testing.T) {
    // Empty body — passes without doing anything
}
EOF

# Run scanner against fixtures
SCAN_OUTPUT=$($SCANNER "$TEMP_DIR" 2>&1) || true

# Verify scanner detected all bluff patterns
DETECTIONS=0
if echo "$SCAN_OUTPUT" | grep -q "assert.True.*true"; then
    ((DETECTIONS++))
else
    echo "FAIL: Scanner did not detect assert.True(t, true)"
fi

if echo "$SCAN_OUTPUT" | grep -q "assert.Equal.*true.*true"; then
    ((DETECTIONS++))
else
    echo "FAIL: Scanner did not detect assert.Equal(t, true, true)"
fi

if echo "$SCAN_OUTPUT" | grep -q "assert.Nil.*nil"; then
    ((DETECTIONS++))
else
    echo "FAIL: Scanner did not detect assert.Nil(t, nil)"
fi

if [[ $DETECTIONS -lt 3 ]]; then
    echo "Phase 1 FAILED: Scanner self-test detected only $DETECTIONS/3 bluff patterns"
    echo "The bluff scanner itself is broken — this is a CRITICAL failure."
    exit 1
fi

echo "Phase 1 PASSED: Scanner correctly detected $DETECTIONS/3 bluff patterns"

# ─── PHASE 2: Full Tree Scan ───
echo "Phase 2: Full source tree scan..."

FULL_OUTPUT=$($SCANNER "$(git rev-parse --show-toplevel)" 2>&1) || {
    echo "Phase 2 FAILED: Full tree scan found bluff patterns"
    echo "$FULL_OUTPUT"
    exit 1
}

echo "Phase 2 PASSED: No bluff patterns in full tree"
echo "Bluff scanner challenge PASSED"
```

#### 8.1.3 `challenges_compile_challenge.sh` — Compilation Verification

**Purpose**: Verify that all challenges and their dependencies compile successfully.

```bash
#!/bin/bash
# challenges/scripts/challenges_compile_challenge.sh

set -euo pipefail

CHALLENGES_DIR="vasic-digital/Challenges"
FAILURES=0

echo "Challenges compile challenge..."

# Compile the main challenge module
echo "Compiling challenge framework..."
if ! (cd "$CHALLENGES_DIR" && go build ./...); then
    echo "FAIL: Challenge framework does not compile"
    ((FAILURES++))
fi

# Compile challenge runner
echo "Compiling challenge runner..."
if ! (cd "$CHALLENGES_DIR" && go build -o /tmp/challenge-runner ./cmd/userflow-runner); then
    echo "FAIL: Challenge runner does not compile"
    ((FAILURES++))
fi

# Compile all challenge bank definitions
echo "Verifying challenge bank YAML..."
for bank in "$CHALLENGES_DIR"/banks/*.yaml; do
    if [[ -f "$bank" ]]; then
        if ! yq eval '.' "$bank" > /dev/null 2>&1; then
            echo "FAIL: Invalid YAML in $bank"
            ((FAILURES++))
        fi
    fi
done

# Verify all referenced challenge handlers exist
echo "Verifying challenge handler references..."
for bank in "$CHALLENGES_DIR"/banks/*.yaml; do
    if [[ -f "$bank" ]]; then
        handlers=$(yq eval '.challenges[].handler' "$bank" 2>/dev/null || true)
        for handler in $handlers; do
            if [[ "$handler" != "null" && -n "$handler" ]]; then
                # Check handler is registered
                if ! grep -r "Register.*$handler" "$CHALLENGES_DIR/pkg/" > /dev/null 2>&1; then
                    echo "FAIL: Handler '$handler' from $bank not registered"
                    ((FAILURES++))
                fi
            fi
        done
    fi
done

if [[ $FAILURES -gt 0 ]]; then
    echo "Challenges compile challenge FAILED: $FAILURES issue(s)"
    exit 1
fi

echo "Challenges compile challenge PASSED"
```

#### 8.1.4 `challenges_functionality_challenge.sh` — Functional Test Challenge

**Purpose**: Execute a subset of challenges that verify core functionality of the platform.

```bash
#!/bin/bash
# challenges/scripts/challenges_functionality_challenge.sh

set -euo pipefail

CHALLENGES_DIR="vasic-digital/Challenges"
RUNNER="$CHALLENGES_DIR/challenge-runner"
REPORT_FILE="/tmp/functional-challenge-report.json"
FAILURES=0

echo "Challenges functionality challenge..."

# Build runner if needed
if [[ ! -x "$RUNNER" ]]; then
    (cd "$CHALLENGES_DIR" && go build -o challenge-runner ./cmd/userflow-runner)
fi

# Run functionality challenges
if ! "$RUNNER" \
    --banks-dir "$CHALLENGES_DIR/banks" \
    --categories "game-streaming,navigation,recording" \
    --report "$REPORT_FILE"; then
    echo "FAIL: Functionality challenges execution failed"
    ((FAILURES++))
fi

# Verify anti-bluff validation
if [[ -f "$REPORT_FILE" ]]; then
    passed=$(jq '.summary.passed // 0' "$REPORT_FILE")
    total=$(jq '.summary.total // 0' "$REPORT_FILE")
    anti_bluff_valid=$(jq '.antiBluffValid // false' "$REPORT_FILE")

    echo "Results: $passed/$total challenges passed"

    if [[ "$anti_bluff_valid" != "true" ]]; then
        echo "FAIL: Anti-bluff validation failed"
        ((FAILURES++))
    fi

    if [[ $passed -eq 0 ]]; then
        echo "FAIL: No challenges passed"
        ((FAILURES++))
    fi
else
    echo "FAIL: No challenge report generated"
    ((FAILURES++))
fi

if [[ $FAILURES -gt 0 ]]; then
    echo "Challenges functionality challenge FAILED"
    exit 1
fi

echo "Challenges functionality challenge PASSED"
```

#### 8.1.5 `challenges_unit_challenge.sh` — Unit Test Challenge

**Purpose**: Run all unit tests in the Challenges repository and verify they pass.

```bash
#!/bin/bash
# challenges/scripts/challenges_unit_challenge.sh

set -euo pipefail

CHALLENGES_DIR="vasic-digital/Challenges"
FAILURES=0

echo "Challenges unit challenge..."

# Run all unit tests with race detection
echo "Running Challenges unit tests..."
if ! (cd "$CHALLENGES_DIR" && go test -count=1 -race -p 1 ./...); then
    echo "FAIL: Challenges unit tests failed"
    ((FAILURES++))
fi

# Verify coverage
echo "Checking coverage..."
coverage=$(cd "$CHALLENGES_DIR" && go test -coverprofile=/tmp/challenges-coverage.out ./... 2>&1 | \
    grep -oP 'coverage: \K[0-9.]+' | tail -1)

echo "Challenges coverage: ${coverage}%"

# Coverage threshold: 80% for challenges framework
if (( $(echo "$coverage < 80" | bc -l) )); then
    echo "FAIL: Coverage $coverage% is below 80% threshold"
    ((FAILURES++))
fi

if [[ $FAILURES -gt 0 ]]; then
    echo "Challenges unit challenge FAILED"
    exit 1
fi

echo "Challenges unit challenge PASSED (coverage: ${coverage}%)"
```

#### 8.1.6 `host_no_auto_suspend_challenge.sh` — Host Integrity Verification

**Purpose**: Verify that the host system is configured to never auto-suspend, as this would disrupt streaming sessions.

```bash
#!/bin/bash
# challenges/scripts/host_no_auto_suspend_challenge.sh

set -euo pipefail

FAILURES=0

echo "Host no-auto-suspend challenge..."

# Check systemd sleep configuration
if [[ -f "/etc/systemd/sleep.conf" ]]; then
    if grep -q "SuspendMode\|HibernateMode" "/etc/systemd/sleep.conf" 2>/dev/null; then
        echo "FAIL: sleep.conf contains suspend/hibernate configuration"
        ((FAILURES++))
    fi
fi

# Check systemd sleep.conf.d/
if [[ -d "/etc/systemd/sleep.conf.d" ]]; then
    for conf in /etc/systemd/sleep.conf.d/*.conf; do
        if [[ -f "$conf" ]]; then
            if grep -q "SuspendMode\|HibernateMode\|AllowHibernation" "$conf" 2>/dev/null; then
                echo "FAIL: $conf contains suspend/hibernate configuration"
                ((FAILURES++))
            fi
        fi
    done
fi

# Check logind.conf
if [[ -f "/etc/systemd/logind.conf" ]]; then
    # These settings must be explicitly set to ignore
    if ! grep -q "HandleLidSwitch=ignore" "/etc/systemd/logind.conf" 2>/dev/null; then
        echo "FAIL: logind.conf does not set HandleLidSwitch=ignore"
        ((FAILURES++))
    fi
    if ! grep -q "HandleSuspendKey=ignore" "/etc/systemd/logind.conf" 2>/dev/null; then
        echo "FAIL: logind.conf does not set HandleSuspendKey=ignore"
        ((FAILURES++))
    fi
    if ! grep -q "HandleHibernateKey=ignore" "/etc/systemd/logind.conf" 2>/dev/null; then
        echo "FAIL: logind.conf does not set HandleHibernateKey=ignore"
        ((FAILURES++))
    fi
fi

# Check for any running suspend services
if systemctl is-active --quiet suspend.target hibernate.target hybrid-sleep.target 2>/dev/null; then
    echo "FAIL: A suspend/hibernate target is active"
    ((FAILURES++))
fi

# Verify no suspend in cron/systemd timers
if grep -r "suspend\|hibernate" /etc/cron.* /var/spool/cron/ 2>/dev/null; then
    echo "FAIL: Found suspend/hibernate in cron jobs"
    ((FAILURES++))
fi

if [[ $FAILURES -gt 0 ]]; then
    echo "Host no-auto-suspend challenge FAILED: $FAILURES issue(s)"
    echo "WARNING: Host may auto-suspend, which will disrupt streaming!"
    exit 1
fi

echo "Host no-auto-suspend challenge PASSED"
```

#### 8.1.7 `mutation_ratchet_challenge.sh` — Mutation Testing Ratchet

See Section 4.4.2 for the full implementation. This script:
1. Runs `go-mutesting` on the full codebase
2. Compares the score against the stored baseline
3. Fails if the score decreases (ratchet violation)
4. Fails if the score is below the 85% threshold
5. Updates the stored score on success

#### 8.1.8 `no_suspend_calls_challenge.sh` — No Suspend Calls Verification

**Purpose**: Verify that no code in the repository calls suspend, hibernate, or shutdown functions.

```bash
#!/bin/bash
# challenges/scripts/no_suspend_calls_challenge.sh

set -euo pipefail

FAILURES=0

echo "No-suspend-calls challenge..."

# Forbidden function calls in Go code
FORBIDDEN_GO=(
    'exec\.Command.*"systemctl".*"suspend"'
    'exec\.Command.*"systemctl".*"hibernate"'
    'exec\.Command.*"pm-suspend"'
    'exec\.Command.*"rtcwake"'
    'exec\.Command.*"shutdown"'
    'exec\.Command.*"poweroff"'
    'exec\.Command.*"reboot"'
    'syscall\.Reboot'
)

# Forbidden shell commands in scripts
FORBIDDEN_SHELL=(
    'systemctl\s+suspend'
    'systemctl\s+hibernate'
    'pm-suspend'
    'rtcwake'
    'shutdown\s+'
    'poweroff'
    'reboot\s+'
    'init\s+0'
)

# Scan Go code
for pattern in "${FORBIDDEN_GO[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.go" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found forbidden suspend call in Go code:"
        echo "$matches"
        ((FAILURES++))
    fi
done

# Scan shell scripts
for pattern in "${FORBIDDEN_SHELL[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.sh" --include="*.bash" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found forbidden suspend call in shell script:"
        echo "$matches"
        ((FAILURES++))
    fi
done

# Scan YAML files (for container commands)
for pattern in "${FORBIDDEN_SHELL[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.yml" --include="*.yaml" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found forbidden suspend call in YAML:"
        echo "$matches"
        ((FAILURES++))
    fi
done

# Check for imports of dangerous packages
DANGEROUS_IMPORTS=(
    'syscall.*REBOOT'
    'golang.org/x/sys/unix.*Reboot'
)
for pattern in "${DANGEROUS_IMPORTS[@]}"; do
    matches=$(grep -r -n -P "$pattern" --include="*.go" . || true)
    if [[ -n "$matches" ]]; then
        echo "FAIL: Found dangerous reboot import:"
        echo "$matches"
        ((FAILURES++))
    fi
done

if [[ $FAILURES -gt 0 ]]; then
    echo "No-suspend-calls challenge FAILED: $FAILURES forbidden call(s) found"
    exit 1
fi

echo "No-suspend-calls challenge PASSED: No suspend/hibernate/shutdown calls found"
```

### 8.2 Challenge Banks

The Challenges repository contains 50+ challenge bank files covering all aspects of the platform.

#### 8.2.1 Game-Specific Challenges

**File**: `vasic-digital/Challenges/banks/game-streaming.yaml`

```yaml
challenge:
  name: "HelixPlay Game Stream Validation"
  category: "game-streaming"
  description: "Verify game streaming works for each supported title"
  games:
    - name: "Elden Ring"
      platform: "steam"
      expected_fps: 60
      expected_latency_ms: 30
    - name: "Cyberpunk 2077"
      platform: "steam"
      expected_fps: 60
      expected_latency_ms: 30
    - name: " Baldur's Gate 3"
      platform: "steam"
      expected_fps: 60
      expected_latency_ms: 30
  steps:
    - name: "Launch game"
      action: "host.launch_game"
      timeout: 60s
    - name: "Verify stream"
      action: "client.verify_stream"
      assertions:
        - fps: ">= ${game.expected_fps}"
        - latency_ms: "<= ${game.expected_latency_ms}"
        - resolution: "1920x1080"
  antiBluff:
    recordedActions: ["host.launch_game", "client.verify_stream"]
    assertions: ["fps", "latency_ms", "resolution"]
```

#### 8.2.2 Full QA Challenges

**Android**: `vasic-digital/Challenges/banks/full-qa-android.yaml`
- Tests all UI flows on Android mobile client
- Includes touch gestures, app lifecycle, background/foreground
- Requires HelixQA visual assertion for each screen

**Android TV**: `vasic-digital/Challenges/banks/full-qa-androidtv.yaml`
- Tests all UI flows on Android TV (leanback) client
- Includes D-pad navigation, voice search, recommendation row
- Requires HelixQA visual assertion for each screen

```yaml
challenge:
  name: "Full QA — Android TV"
  category: "full-qa"
  platform: "androidtv"
  flows:
    - name: "Main Navigation"
      steps:
        - action: "navigate.home"
          visual_assert: "home_screen_visible"
        - action: "navigate.library"
          visual_assert: "library_screen_visible"
        - action: "navigate.settings"
          visual_assert: "settings_screen_visible"
    - name: "Game Launch"
      steps:
        - action: "navigate.library"
        - action: "select.game"
          params: {game: "Elden Ring"}
        - action: "click.play"
          visual_assert: "stream_active"
          timeout: 30s
  antiBluff:
    visual_assertions_required: true
    recordedActions: ["navigate.*", "select.*", "click.*"]
```

#### 8.2.3 Navigation Challenges

**File**: `vasic-digital/Challenges/banks/navigation.yaml`

Tests all navigation flows across all client platforms:
- Home -> Library -> Game Detail -> Play
- Home -> Settings -> Account -> Logout
- Home -> Search -> Results -> Game Detail
- In-stream: Pause -> Resume -> Quit
- In-stream: Switch input method (touch/gamepad/keyboard)

#### 8.2.4 Recording Challenges

**File**: `vasic-digital/Challenges/banks/recording.yaml`

Tests the recording feature (host-agent encoder dual-path):
- Start stream -> Verify recording starts automatically
- Stop stream -> Verify recording is sealed
- Verify recording file exists with expected size
- Verify recording can be played back
- Verify recording metadata (duration, resolution, codec)

---

## Appendix A: Quick Reference — Anti-Bluff Checklist

### For Every New Feature

- [ ] Unit tests with >= 100% branch coverage (mock isolation logic only)
- [ ] Integration tests with **real** dependencies (no mocks)
- [ ] E2E test covering the complete user journey
- [ ] Security test (fuzz target + specific attack vectors)
- [ ] Benchmark covering performance-critical paths
- [ ] Negative-leg test (break the feature, verify test fails)
- [ ] `ValidateAntiBluff()` evidence in challenge result
- [ ] Usability evidence (HelixQA visual assertion, recording, or challenge execution)
- [ ] Behavior-anchor manifest entry
- [ ] No forbidden patterns in code (scanner clean)

### For Every PR

- [ ] All 8 challenge scripts pass locally
- [ ] `make anti-bluff` passes
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] `golangci-lint` passes
- [ ] No new vacuous assertions introduced
- [ ] Coverage does not decrease
- [ ] Anti-bluff scan passes in CI (non-overridable)

### CI Workflow Status Dashboard

| Workflow | Status | Priority | Implemented |
|----------|--------|----------|-------------|
| `anti-bluff.yml` | **MANDATORY** | P0 | No |
| `ci.yml` (basic) | Required | P0 | No |
| `ci.yml` (extended) | Required | P1 | No |
| `challenges.yml` | Required | P1 | No |
| `mutation.yml` | Required | P2 | No |
| Full CI (all 10 types) | Required | P3 | No |

---

*Document generated as part of the HelixPlay Comprehensive Implementation Plan.  
Constitutional Reference: v2.1.0  
All testing requirements are mandatory per Constitution §6 and cannot be overridden.*
