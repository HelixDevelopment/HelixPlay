// Package fullauto orchestrates all 10 test types for pre-release validation.
package fullauto

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// TestType represents one of the 10 required test categories.
type TestType string

const (
	Unit        TestType = "unit"
	Integration TestType = "integration"
	E2E         TestType = "e2e"
	Security    TestType = "security"
	Benchmark   TestType = "benchmark"
	Chaos       TestType = "chaos"
	Stress      TestType = "stress"
	Smoke       TestType = "smoke"
	FullAuto    TestType = "fullauto"
	Challenge   TestType = "challenge"
)

// Result captures the outcome of a single test-type run.
type Result struct {
	Type   TestType
	Passed bool
	Error  error
	Output string
}

// Orchestrator runs the complete test matrix.
type Orchestrator struct {
	results []Result
}

// NewOrchestrator creates a new test orchestrator.
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{results: make([]Result, 0, 10)}
}

// RunAll executes all test types. If failFast is true, it stops on first failure.
// Per Constitution §1.3, CI negative-leg fault injection runs with failFast=false
// to collect the full failure surface.
func (o *Orchestrator) RunAll(failFast bool) bool {
	types := []TestType{Unit, Integration, E2E, Security, Benchmark, Chaos, Stress, Smoke, Challenge}
	allPassed := true

	for _, tt := range types {
		result := o.runType(tt)
		o.results = append(o.results, result)
		if !result.Passed {
			allPassed = false
			fmt.Printf("FAILED: %s: %v\n", tt, result.Error)
			if failFast {
				return false
			}
		} else {
			fmt.Printf("PASSED: %s\n", tt)
		}
	}

	return allPassed
}

// Results returns all collected results.
func (o *Orchestrator) Results() []Result {
	return o.results
}

// ResultByType returns the outcome of a specific test type.
func (o *Orchestrator) ResultByType(tt TestType) *Result {
	for i := range o.results {
		if o.results[i].Type == tt {
			return &o.results[i]
		}
	}
	return nil
}

// Summary prints a human-readable summary of all results.
func (o *Orchestrator) Summary() {
	passed := 0
	failed := 0
	for _, r := range o.results {
		if r.Passed {
			passed++
		} else {
			failed++
			fmt.Printf("  FAIL %s: %v\n", r.Type, r.Error)
		}
	}
	fmt.Printf("\nTotal: %d passed, %d failed\n", passed, failed)
}

func (o *Orchestrator) runType(tt TestType) Result {
	r := Result{Type: tt, Passed: true}
	switch tt {
	case Unit:
		r.Error = runGoTest("./cmd/...", "./pkg/...")
	case Integration:
		r.Error = runGoTest("./tests/integration/...")
	case E2E:
		r.Error = runGoTest("./tests/e2e/...")
	case Security:
		r.Error = runGoTest("./tests/security/...")
	case Benchmark:
		r.Error = runGoTest("-bench=.", "./tests/benchmark/...")
	case Chaos:
		r.Error = runGoTest("./tests/chaos/...")
	case Stress:
		r.Error = runGoTest("-run=Test1000", "./tests/stress/...")
	case Smoke:
		r.Error = runGoTest("./tests/smoke/...")
	case Challenge:
		r.Error = runChallenges()
	default:
		r.Error = fmt.Errorf("unknown test type: %s", tt)
	}
	if r.Error != nil {
		r.Passed = false
	}
	return r
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
	// Locate the Challenges submodule relative to the current working directory.
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("cannot determine caller path")
	}
	challengesDir := filepath.Join(filepath.Dir(callerFile), "..", "..", "Challenges")
	if _, err := os.Stat(challengesDir); os.IsNotExist(err) {
		return fmt.Errorf("Challenges submodule not found at %s", challengesDir)
	}

	cmd := exec.Command("go", "test", "-count=1", "-race", "-p", "1", "./...")
	cmd.Dir = challengesDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("challenge tests failed: %w\n%s", err, string(out))
	}

	// Also run anti-bluff scan.
	rootDir := filepath.Dir(challengesDir)
	scanScript := filepath.Join(rootDir, "scripts", "anti-bluff-scan.sh")
	if _, err := os.Stat(scanScript); err == nil {
		scanCmd := exec.Command("bash", scanScript)
		scanOut, scanErr := scanCmd.CombinedOutput()
		if scanErr != nil {
			return fmt.Errorf("anti-bluff scan failed: %w\n%s", scanErr, string(scanOut))
		}
	}

	return nil
}

// RunAntiBluffScan executes the anti-bluff scanner as a standalone check.
func RunAntiBluffScan() error {
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("cannot determine caller path")
	}
	rootDir := filepath.Dir(callerFile)
	for i := 0; i < 2; i++ {
		rootDir = filepath.Dir(rootDir)
	}

	scanScript := filepath.Join(rootDir, "scripts", "anti-bluff-scan.sh")
	if _, err := os.Stat(scanScript); os.IsNotExist(err) {
		return fmt.Errorf("anti-bluff scan script not found at %s", scanScript)
	}

	cmd := exec.Command("bash", scanScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("anti-bluff scan failed: %w\n%s", err, string(out))
	}
	if strings.Contains(string(out), "ERROR") {
		return fmt.Errorf("anti-bluff scan found violations:\n%s", string(out))
	}
	return nil
}
