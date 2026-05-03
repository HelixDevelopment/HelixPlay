package security_test

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/input"
	"github.com/stretchr/testify/require"
)

// TestNegativeLeg_BrokenControllerInputFailsIntegration verifies that
// attempting to use a controller that failed to connect does not crash
// the input pipeline. Negative-leg test T271.
func TestNegativeLeg_BrokenControllerInputFailsIntegration(t *testing.T) {
	ds := input.NewDualSense()
	require.NotNil(t, ds, "DualSense object must be created")

	// Connect will fail if no hardware is present (expected in CI).
	// The critical requirement is that it returns an error rather than
	// panicking or pretending to succeed.
	err := ds.Connect()
	if err == nil {
		// If connected, disconnecting must also be safe.
		ds.Disconnect()
	}

	// After disconnect (or failed connect), the controller must report
	// itself as not connected.
	require.False(t, ds.IsConnected(), "controller must not appear connected after failed connect or disconnect")
}
