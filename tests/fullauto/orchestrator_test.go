package fullauto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestratorRunAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full automation in short mode")
	}

	o := NewOrchestrator()
	passed := o.RunAll(false)

	// Verify all 9 test types were executed.
	results := o.Results()
	require.Len(t, results, 9, "orchestrator should run 9 test types")

	// Verify each test type has a result.
	for _, tt := range []TestType{Unit, Integration, E2E, Security, Benchmark, Chaos, Stress, Smoke, Challenge} {
		r := o.ResultByType(tt)
		require.NotNil(t, r, "result for %s should not be nil", tt)
		assert.Equal(t, tt, r.Type)
		t.Logf("%s: passed=%v", tt, r.Passed)
	}

	t.Logf("Orchestrator completed. Overall passed: %v", passed)
}

func TestOrchestratorResultByType(t *testing.T) {
	o := NewOrchestrator()

	// Before running, result should be nil.
	require.Nil(t, o.ResultByType(Unit))

	// After running (even with no-op test types), results should exist.
	_ = o.RunAll(true)
	require.NotNil(t, o.ResultByType(Unit))
}

func TestOrchestratorSummary(t *testing.T) {
	o := NewOrchestrator()
	// Populate with dummy results.
	o.results = append(o.results, Result{Type: Unit, Passed: true})
	o.results = append(o.results, Result{Type: Smoke, Passed: false, Error: assert.AnError})

	// Summary should not panic and should reflect the results.
	o.Summary()
	assert.Len(t, o.Results(), 2)
}

func TestRunAntiBluffScan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping anti-bluff scan in short mode")
	}

	err := RunAntiBluffScan()
	assert.NoError(t, err, "anti-bluff scan should pass")
}
