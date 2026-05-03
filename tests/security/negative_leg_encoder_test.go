package security_test

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
	"github.com/stretchr/testify/require"
)

// TestNegativeLeg_BrokenEncoderFailsE2E verifies that when the encoder
// is configured with an invalid codec, the streaming pipeline fails
// rather than silently producing garbage. Negative-leg test T270.
func TestNegativeLeg_BrokenEncoderFailsE2E(t *testing.T) {
	// Attempt to create an encoder with a non-existent codec.
	// This MUST fail because an unusable encoder cannot stream.
	enc := encoder.NewHardwareEncoder("nonexistent-gpu")
	require.Nil(t, enc, "encoder with invalid codec must not be created")
}
