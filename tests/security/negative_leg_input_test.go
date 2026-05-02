package security_test

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/input"
)

func TestNegativeLegInputFailure(t *testing.T) {
	// This test verifies that controller input failures are detectable.
	mgr := input.NewManager()

	// Positive control: normal controller state
	mgr.UpdateState(input.ControllerState{
		ID:         0,
		Connected:  true,
		ButtonMask: 0x1000,
	})

	if !mgr.AnyConnected() {
		t.Fatal("expected at least one connected controller (positive control)")
	}

	// Negative leg: disconnect all controllers
	mgr.UpdateState(input.ControllerState{ID: 0, Connected: false})

	if mgr.AnyConnected() {
		t.Fatal("expected no connected controllers after disconnect")
	}

	// Verify state retrieval fails gracefully
	_, ok := mgr.GetState(99)
	if ok {
		t.Fatal("expected GetState to return false for unknown controller")
	}

	t.Log("Negative leg input test: fault injection path verified")
}
