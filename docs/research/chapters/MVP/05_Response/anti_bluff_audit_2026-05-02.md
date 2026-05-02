# Anti-Bluff Compliance Audit Report

**Date:** 2026-05-02
**Auditor:** Automated scan + manual review
**Constitution Version:** v2.2.0
**Scope:** All 23 active submodules + root module

---

## Executive Summary

| Category | Findings | Severity |
|----------|----------|----------|
| Vacuous assertions (`assert.True(t, true)`) | 0 | — |
| Constructor-only tests | 2 | CRITICAL |
| Mocks in non-Unit tests | 9 files | BLOCKER |
| `panic("not implemented")` | 0 | — |
| Empty function bodies (non-fixture) | 0 | — |
| `TODO`/`FIXME` in own code | 2 | WARNING |
| `t.Skip` without containerization plan | 0 | — |
| Spec Constitution version outdated | 0 | HIGH |

**Overall Status:** 🔴 **VIOLATIONS FOUND** — merge blocked until remediation

---

## 1. Constructor-Only Tests (§1.1.1 — CRITICAL)

A constructor-only test calls a constructor and only verifies the returned object is non-nil or has expected field values, without exercising any observable behaviour.

### 1.1 `Challenges/Containers/pkg/ctop/ctop_test.go:29` — `TestNewCollectorWithExecutor`

```go
func TestNewCollectorWithExecutor(t *testing.T) {
	exec := &mockExecutor{}
	c := NewCollectorWithExecutor("docker", nil, exec)
	assert.NotNil(t, c)
	assert.Equal(t, "docker", c.runtime)
	assert.Equal(t, exec, c.executor)
}
```

**Violation:** Only verifies field values. No call to `Collect()`, `Stream()`, or any behaviour-exercising method.
**Fix:** Add a sub-test that calls `Collect()` and verifies container metrics are returned.

### 1.2 `Challenges/Containers/pkg/logging/logger_test.go:17` — `TestNewStdLogger`

```go
func TestNewStdLogger(t *testing.T) {
	// ... table-driven tests ...
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewStdLogger(tt.prefix)
			require.NotNil(t, logger)
			assert.Equal(t, tt.prefix, logger.prefix)
		})
	}
}
```

**Violation:** Only verifies `prefix` field. No log write/output capture verification.
**Fix:** Add a sub-test that writes a log line and verifies output contains the prefix and message.

---

## 2. Mocks in Non-Unit Tests (§6 / R-12 — BLOCKER)

Constitution §6 and R-12 mandate: **Only Unit tests may use mocks/stubs/hardcoded values.** All other 9 test types MUST exercise the real system.

### 2.1 Integration Tests with Mocks

| File | Mock | Violation |
|------|------|-----------|
| `Challenges/Containers/tests/integration/distribution_integration_test.go` | `TestDistributor_WithMockHosts` (NopLogger + missing scheduler) | Uses no-op logger; test name misleadingly says "MockHosts" but does not mock hosts. **Downgraded to WARNING** after review. |
| `Containers/tests/integration/distribution_integration_test.go` | Same as above | Same as above |
| `HelixQA/tests/integration/ocu_record_test.go` | `intMockSource` | Mock capture source in integration test. **BLOCKER.** |
| `HelixQA/tests/integration/agent_stack_integration_test.go` | UI-TARS + OmniParser mock servers | Entire integration test mocks external AI services. **BLOCKER.** |
| `Storage/tests/integration/recording_integration_test.go` | `MockObjectStoreIntegration` | Comment explicitly cites R-12 but then violates it. **BLOCKER.** |
| `Formatters/tests/integration/formatters_integration_test.go` | `mockFormatter` | Mock formatter implementations in integration test. **BLOCKER.** |

### 2.2 E2E Tests with Mocks

| File | Mock | Violation |
|------|------|-----------|
| `Streaming/tests/e2e/streaming_e2e_test.go` | `mockTransport` | Transport layer mocked in E2E test. **BLOCKER.** |
| `Security/tests/e2e/security_e2e_test.go` | `testScanner` | Scanner mocked in E2E test. **BLOCKER.** |

### 2.3 Benchmark Tests with Mocks

| File | Mock | Violation |
|------|------|-----------|
| `HelixQA/tests/benchmark/benchmark_test.go` | `mockBenchProvider` | LLM provider mocked in benchmark. Benchmark measures analyzer with fake LLM, not real behaviour. **BLOCKER.** |

---

## 3. TODO in Own Code (§1.1 — WARNING)

| File | Line | Context | Action |
|------|------|---------|--------|
| `Challenges/pkg/assertion/builtin.go:51` | 51 | `"TODO"` in forbidden-tokens list | **OK** — this is the anti-bluff scanner listing forbidden tokens. |
| `HelixQA/pkg/autonomous/structured_executor.go:577` | 577 | `// unfinished placeholder ("# TODO: Convert to executable ...")` | **WARNING** — placeholder step in executor. Tracked as known limitation. |

---

## 4. Submodule Constitution Propagation Gaps

| Submodule | CLAUDE.md | AGENTS.md | CONSTITUTION.md | Version |
|-----------|-----------|-----------|-----------------|---------|
| All 23 submodules | 40 lines | 45 lines | 6 lines | v2.2.0 ✅ |

---

## 5. Spec Version Mismatch

`specs/001-helixplay-system/spec.md` was updated: Constitution references changed from v2.0.0 to **v2.2.0**, Go version from 1.22+ to **1.26.2**.

---

## Remediation Plan

| Priority | Task | Owner | Deadline |
|----------|------|-------|----------|
| P0 | Update spec.md Constitution references to v2.2.0 + Go 1.26.2 | AI Agent | 2026-05-02 |
| P0 | Rewrite `TestNewCollectorWithExecutor` to exercise `Collect()` | Developer | 2026-05-03 |
| P0 | Rewrite `TestNewStdLogger` to exercise `Write()`/`Output()` | Developer | 2026-05-03 |
| P1 | Replace mocks in `HelixQA/tests/integration/` with real deps or move to Unit | Developer | 2026-05-05 |
| P1 | Replace `MockObjectStoreIntegration` in `Storage/tests/integration/` with MinIO container | Developer | 2026-05-05 |
| P1 | Replace `mockFormatter` in `Formatters/tests/integration/` with real formatter | Developer | 2026-05-05 |
| P1 | Replace `mockTransport` in `Streaming/tests/e2e/` with real WebRTC/QUIC | Developer | 2026-05-05 |
| P1 | Replace `testScanner` in `Security/tests/e2e/` with real scanner | Developer | 2026-05-05 |
| P2 | Add `BLUFF-VIOLATION: R-12` comments to all mock-in-non-unit files | Developer | 2026-05-02 |

---

## Anti-Bluff Verification

| Source | Lines | Insights |
|--------|-------|----------|
| Constitution v2.2.0 §1, §6, §17 | 1,156 | Forbidden patterns, test discipline, R-01..R-18 |
| `scripts/anti-bluff-scan.sh` | 245 | CI scan logic, detection heuristics |
| Test files across 23 submodules | ~550 | grep/awk manual review |
| Gap analysis (speckit explore) | 596 | spec.md gap identification |

**Conflict Resolution:** N/A — audit finds violations, no conflicting sources.

**Coverage Confirmation:** All 24 submodules scanned. `tools/opensource/` third-party code excluded.

**Sign-off:** This audit was generated automatically and MUST be reviewed by a human before merge.

---

*End of Audit Report*
