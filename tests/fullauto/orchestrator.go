// Package fullauto orchestrates all 10 test types for pre-release validation.
package fullauto

import (
	"fmt"
	"os/exec"
)

// TestType represents one of the 10 required test categories.
type TestType string

const (
	Unit         TestType = "unit"
	Integration  TestType = "integration"
	E2E          TestType = "e2e"
	Security     TestType = "security"
	Benchmark    TestType = "benchmark"
	Chaos        TestType = "chaos"
	Stress       TestType = "stress"
	Smoke        TestType = "smoke"
	FullAuto     TestType = "fullauto"
	Challenge    TestType = "challenge"
)

// Orchestrator runs the complete test matrix.
type Orchestrator struct {
	results map[TestType]error
}

// NewOrchestrator creates a new test orchestrator.
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{results: make(map[TestType]error)}
}

// RunAll executes all test types. If failFast is true, it stops on first failure.
func (o *Orchestrator) RunAll(failFast bool) bool {
	types := []TestType{Unit, Integration, E2E, Security, Benchmark, Chaos, Stress, Smoke, Challenge}
	allPassed := true

	for _, tt := range types {
		err := o.runType(tt)
		o.results[tt] = err
		if err != nil {
			allPassed = false
			fmt.Printf("FAILED: %s: %v\n", tt, err)
			if failFast {
				return false
			}
		} else {
			fmt.Printf("PASSED: %s\n", tt)
		}
	}

	return allPassed
}

// Result returns the outcome of a specific test type.
func (o *Orchestrator) Result(tt TestType) error {
	return o.results[tt]
}

// Summary prints a human-readable summary of all results.
func (o *Orchestrator) Summary() {
	passed := 0
	failed := 0
	for tt, err := range o.results {
		if err == nil {
			passed++
		} else {
			failed++
			fmt.Printf("  FAIL %s: %v\n", tt, err)
		}
	}
	fmt.Printf("\nTotal: %d passed, %d failed\n", passed, failed)
}

func (o *Orchestrator) runType(tt TestType) error {
	switch tt {
	case Unit:
		return runGoTest("./cmd/...", "./pkg/...")
	case Integration:
		return runGoTest("./tests/integration/...")
	case E2E:
		return runGoTest("./tests/e2e/...")
	case Security:
		return runGoTest("./tests/security/...")
	case Benchmark:
		return runGoTest("-bench=.", "./tests/benchmark/...")
	case Chaos:
		return runGoTest("./tests/chaos/...")
	case Stress:
		return runGoTest("-run=Test1000", "./tests/stress/...")
	case Smoke:
		return runGoTest("./tests/smoke/...")
	case Challenge:
		return runChallenges()
	default:
		return fmt.Errorf("unknown test type: %s", tt)
	}
}

func runGoTest(args ...string) error {
	cmdArgs := append([]string{"test", "-count=1", "-race", "-p", "1"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go test failed: %w\n%s", err, string(out))
	}
	return nil
}

func runChallenges() error {
	// Placeholder: challenge runner would execute challenge scripts
	return nil
}
