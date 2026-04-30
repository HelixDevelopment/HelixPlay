# T02 — Unit Tests

> **Source dimensions:** [`00_Index.md`](00_Index.md), [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.1, §7 (Mock policy enforcement), §10 (Tooling lockstep); [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §5; [`../01_Constitution.md`](../01_Constitution.md) §6.2 (R-11 + R-12).
> **Source line count:** ≈ 1k+ lines from family inputs; chapter floor is 300 lines per Master Plan §7.2 row T02.
> **Chapter targets:** R-11 (≥ 95 % statement coverage), R-12 (mocks allowed only here).
> **Cross-links:** [`01_Test_Matrix.md`](01_Test_Matrix.md), [`03_Integration_Tests.md`](03_Integration_Tests.md), [`05_Security_Tests.md`](05_Security_Tests.md).
> **Status:** Draft v1.
> **Last updated:** 2026-04-30.

---

## 1. Role in the Master Test Matrix

Unit is **test-type 1 of 10** in the [T01 §2 grid](01_Test_Matrix.md#2-the-29--10--4-cell-grid-the-canonical-lookup). It is the **only** test type where Constitution §6.2 (R-12) permits mocks, stubs, and hardcoded values. Every of the 29 submodules owns its Unit row in-tree under `<submodule>/tests/unit/` (or co-located `*_test.go` per Go convention). The Unit row runs at every cadence (per-PR, nightly, canary, pre-release).

Unit's role is **API contract verification**. A Unit test exercises one logical unit (a function, a method, a struct's behaviour) against either its real dependencies or a mock thereof, and asserts the unit's contract holds: input → output, error pathways, side effects.

Unit is **not** sufficient. Past projects had green Unit + green Integration that hid broken end-user behaviour because Unit's mocks misrepresented the real dependency. R-13 makes this explicit: green Unit alone does not guarantee real behaviour. Unit + Challenges together do.

---

## 2. The Discipline — What Makes "Green" Green

A green Unit row means:

1. **Coverage ≥ 95 %** (statement coverage measured by `go test -coverprofile=coverage.out -covermode=atomic ./...`). The 5 % gap accommodates unreachable error branches; auditable per `<submodule>/.coverage-exemptions.yaml`.
2. **All assertions pass** with no `t.Skip`, no `t.SkipNow`, no skipped tests under any conditional except documented platform-specific `runtime.GOOS == "linux"` guards.
3. **No data races** under `go test -race ./...`.
4. **No timeouts** at the test level — every test completes within its declared `time.After` budget.
5. **No flake retries** — a test marked `flaky-tolerant` is treated as a defect, not a feature.

The first four are objective; the fifth is a **policy** the helix-no-flaky-tests lint enforces by rejecting any `t.Skip` or test-retry annotation in the codebase outside Constitution-§13-approved exceptions.

---

## 3. Tooling

Per [T01 §10](01_Test_Matrix.md#10-tooling-lockstep--pinned-versions-across-the-fleet):

- Go's `testing` package — primary.
- `github.com/stretchr/testify` — selected per submodule (assertion DSL).
- `github.com/onsi/gomega` — alternative; not project-mandated.
- `go-fuzz` (deprecated) → Go ≥ 1.18's stdlib `testing.F` fuzz harness — for fuzz-targeted tests within Unit.
- `go test -race` — race detector mandatory on every Unit invocation.
- `go test -coverprofile` — coverage measurement.

Mock libraries (allowed **only** under `tests/unit/`):

- `github.com/golang/mock/gomock` (deprecated; new code uses…)
- `github.com/stretchr/testify/mock` — preferred mock framework.
- Hand-rolled interfaces — preferred over framework when the mock surface is small (≤ 5 methods).

The `helix-mock-discipline` lint rejects any of these imports outside `tests/unit/` or non-test code.

---

## 4. Per-Submodule Expectations

Each of the 29 submodules' [S05 §11 Per-test-type coverage targets](../06_Submodules/per-submodule/) entries names "Unit" with the ≥ 95 % statement-coverage criterion. Specific submodule patterns:

- **`helix-r18-safeexec`** — Unit tests cover every deny-list entry's rejection + adjacent-string acceptance + Unicode-normalisation attacks + path-resolution attacks + argument-array attacks. The deny-list is the single source of truth, so its Unit coverage is mandatory at every release boundary.
- **`helix-codec`** — Unit tests verify Profile validation against libavutil's matrix; mock libavutil at the Unit layer to avoid the cgo dependency in fast-test runs.
- **`helix-pipeline`** — Unit tests cover Pipeline lifecycle state-machine + Event channel correctness; the 9-sibling integration is verified at the Integration layer (T03), not Unit.
- **`helix-allocator`** — Unit tests verify the analyser detects every documented forbidden pattern; the runtime guard's Strict-mode panic is observed via `recover()` in the test.

The remaining submodules follow the same shape — Unit verifies per-symbol contracts; Integration / E2E / Challenges verify cross-symbol + cross-system behaviour.

---

## 5. Anti-Pattern Catalogue

The following patterns are **forbidden** under R-12 + R-13 enforcement. The lint catches them; PR review escalates.

### 5.1 Test that always passes

```go
func TestSomething(t *testing.T) {
    if false {
        t.Fatal("this never fires")
    }
}
```

A test with no real assertion is a defect. The helix-empty-test-detector lint rejects functions whose body has no `t.Errorf`, `t.Fatalf`, or test-helper invocation.

### 5.2 Test that mocks the system under test

```go
func TestEncrypt(t *testing.T) {
    mock := NewEncryptMock()  // mocks the very function we're testing
    mock.On("Encrypt").Return([]byte{...})
    result, _ := mock.Encrypt(plaintext)
    assert.Equal(t, expected, result)
}
```

This tests the mock, not the system under test. PR review catches this; the static analyser flags the pattern when a mock's method-name matches the package's own exported function.

### 5.3 Mock that returns the input

```go
mock.On("Encrypt", mock.Anything).Return(func(plain []byte) []byte {
    return plain  // returns the input — a no-op encryption
})
```

Same defect — the mock pretends to be the function. Reviewer escalation.

### 5.4 Test that asserts on the mock's call count and nothing else

```go
func TestAuthFlow(t *testing.T) {
    mockAuth := NewMockAuth()
    flow := NewFlow(mockAuth)
    flow.Run()
    mockAuth.AssertCalled(t, "Authenticate")
}
```

Asserting that the mock was called says nothing about whether the system did the right thing. Add output assertions; the call-count assertion is a supplement, not a substitute.

### 5.5 Test marked `t.Skip` "until we figure it out"

A skipped test is a defect, not a placeholder. The helix-no-skip lint rejects any unconditional `t.Skip` outside Constitution §13 *Exceptions* approval.

### 5.6 Hardcoded sleep instead of synchronisation

```go
go pipeline.Run()
time.Sleep(100 * time.Millisecond)  // hope the goroutine started
assert.True(t, pipeline.Started())
```

The 100 ms is fragile; under CI load it fails. Use `<-startedCh` or `pipeline.WaitForStart(ctx)` instead.

### 5.7 Goroutine leak in a Unit test

A Unit test that spawns goroutines must clean them up. Use `t.Cleanup(func(){ cancel(); wg.Wait() })`. The `goleak` library (`go.uber.org/goleak`) is the go-to detector at test-package teardown.

---

## 6. CI Lane Invocation Pattern

The Unit row's invocation in `vasic-digital/Containers/ci-fragments/ten-test-types-matrix.yml`:

```yaml
- name: Unit tests
  if: matrix.test-type == 'unit'
  run: |
    go test -race -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out | tail -1 | awk '{print "Coverage:", $3}'
    coverage_pct=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage_pct < 95.0" | bc -l) )); then
        echo "FAIL: Unit coverage $coverage_pct% < 95% required"
        exit 1
    fi
- name: Mock-discipline lint
  if: matrix.test-type == 'unit'
  run: |
    helix-mock-discipline ./...
- name: No-skip lint
  if: matrix.test-type == 'unit'
  run: |
    helix-no-skip ./...
- name: Goleak verification
  if: matrix.test-type == 'unit'
  run: |
    GOFLAGS="-count=1" go test -tags=goleak ./...
```

The pattern is per-submodule via the workflow's `matrix.test-type` axis. A red exit on any sub-step blocks the merge.

---

## 7. Test Fixture Pattern — Recommended

Per [T01 §14c](01_Test_Matrix.md#14c-test-fixture-conventions):

```go
// helix-shm/page_test.go
func TestPage_NewPool(t *testing.T) {
    cases := []struct {
        name      string
        pageSize  int
        capacity  int
        wantErr   error
    }{
        {"valid", 4096, 64, nil},
        {"zero-capacity", 4096, 0, ErrInvalidCapacity},
        {"non-page-aligned", 4097, 64, ErrSizeNotPageAligned},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            t.Cleanup(func() {})
            pool, err := NewPool(tc.pageSize, tc.capacity)
            if !errors.Is(err, tc.wantErr) {
                t.Fatalf("NewPool: got %v, want %v", err, tc.wantErr)
            }
            if err == nil {
                pool.Close()
            }
        })
    }
}
```

Properties of the recommended pattern:

- Table-driven (named cases for readability).
- `t.Cleanup` for resource teardown.
- `errors.Is` for sentinel-error matching (compatible with wrapped errors).
- Subtests via `t.Run` (each case isolated; failure one doesn't block the rest).
- No mocks where real types suffice (`*Pool` here — there's no dependency to mock).

---

## 7a. Mandatory test-package boilerplate

Every submodule's `tests/unit/` ships an entry-point file that registers `goleak` + the per-submodule fixtures. The canonical shape:

```go
// helix-shm/tests/unit/main_test.go
package unit_test

import (
    "os"
    "testing"

    "go.uber.org/goleak"
)

func TestMain(m *testing.M) {
    code := m.Run()
    if code == 0 {
        // Only check leaks on success; on failure, the leak is not the
        // top-level signal (and may be the consequence of a panicked test).
        if err := goleak.Find(); err != nil {
            os.Stderr.WriteString("goleak: " + err.Error() + "\n")
            code = 1
        }
    }
    os.Exit(code)
}
```

The pattern is consumed by `helix-test-conventions-lint` (per [T01 §14c](01_Test_Matrix.md#14c-test-fixture-conventions)) — a `tests/unit/` directory missing this `TestMain` entry-point is a defect.

## 7b. Coverage-Exemption File Format

The per-submodule `<submodule>/.coverage-exemptions.yaml` declares the genuinely-unreachable code that the ≥ 95 % statement-coverage gate may skip. Schema:

```yaml
# Schema: vasic-digital/.github/coverage-exemptions-schema.json
exemptions:
  - file: internal/encoder/dispatch.go
    function: panicUnreachable
    reason: "Catch-all panic after exhaustive switch on Vendor enum. Adding a Vendor enum value triggers a compile-time review."
    expires: "2027-04-30"   # Annual review per Constitution §15.
  - file: internal/cgo_bindings.go
    function: cleanup
    reason: "C-side `defer` cleanup invoked from cgo cleanup-handler; not statement-coverable from Go side."
    expires: "2027-04-30"
```

Rules: every exemption requires a `reason` (audited at PR review) and an `expires` field (annual review per Constitution §15 *Amendment Procedure*). The `helix-coverage-exemption-audit` lint fails CI if any exemption is past its `expires` date.

## 8. Open Questions

| ID                  | Question                                                                                                       | Defer to                                            |
|---------------------|----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| OQ-T02-A            | Coverage exemption review — quarterly automated audit or manual per-release?                                  | T10 Full Automation chapter                         |
| OQ-T02-B            | Fuzzer adoption — every submodule mandatory or opt-in?                                                         | T05 Security chapter                                |
| OQ-T02-C            | Goleak default-on — every test package or opt-in via tag?                                                     | T10 Full Automation chapter                         |

---

## 9. References & Anti-Bluff Verification

### 9.1 Internal

- [`01_Test_Matrix.md`](01_Test_Matrix.md) §6.1, §7, §10, §14c.
- [`../06_Submodules/01_Submodule_Catalog.md`](../06_Submodules/01_Submodule_Catalog.md) §5.
- [`../01_Constitution.md`](../01_Constitution.md) §6.2.

### 9.2 External (web)

- Go testing package: https://pkg.go.dev/testing (accessed 2026-04-30).
- testify mock: https://pkg.go.dev/github.com/stretchr/testify/mock (accessed 2026-04-30).
- goleak: https://github.com/uber-go/goleak (accessed 2026-04-30).
- Go fuzzing (testing.F): https://go.dev/doc/security/fuzz/ (accessed 2026-04-30).

### 9.3 Anti-Bluff Verification

- Coverage: chapter exceeds the 300-line floor.
- Forbidden patterns: clean. Three §8 open questions named with deferred resolution chapters.
- The §5 anti-pattern catalogue codifies the failure modes that R-12 exists to prevent.

### 9.4 Sign-off

- Drafted by: orchestrator (Claude Opus 4.7) on 2026-04-30.
- Reviewed by: pending operator review.

End of `07_Testing/02_Unit_Tests.md` — 2026-04-30.
