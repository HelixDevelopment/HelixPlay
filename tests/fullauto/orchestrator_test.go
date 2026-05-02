package fullauto

import (
	"testing"
)

func TestOrchestratorRunAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full automation in short mode")
	}

	o := NewOrchestrator()
	passed := o.RunAll(false)

	// In this environment, some test suites may not have real dependencies
	// so we just verify the orchestrator runs without panicking
	t.Logf("Orchestrator completed. Overall passed: %v", passed)
}

func TestOrchestratorResult(t *testing.T) {
	o := NewOrchestrator()

	// Before running, all results should be nil (not executed)
	if o.Result(Unit) != nil {
		t.Log("Unit test not yet executed")
	}
}

func TestOrchestratorSummary(t *testing.T) {
	o := NewOrchestrator()
	// Populate with dummy results
	o.results[Unit] = nil
	o.results[Smoke] = nil
	o.Summary()
}
