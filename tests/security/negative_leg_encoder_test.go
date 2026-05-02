package security_test

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
)

func TestNegativeLegEncoderFailure(t *testing.T) {
	// This test verifies that the encoder performs real work under normal conditions
	// and properly fails after being stopped (negative leg).

	enc := encoder.NewDualPath("software")
	if enc == nil {
		t.Fatal("expected encoder to be created")
	}

	if err := enc.Start(); err != nil {
		t.Fatalf("encoder start failed: %v", err)
	}
	defer enc.Stop()

	// Positive control: verify encoder performs real transformation
	frame := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	encoded, err := enc.EncodeFrame(frame)
	if err != nil {
		t.Fatalf("positive control encode failed: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("expected encoded output (positive control)")
	}

	// Negative leg: simulate encoder failure by stopping it
	enc.Stop()
	if enc.IsStreaming() {
		t.Fatal("expected encoder to be stopped after Stop()")
	}

	// Verify EncodeFrame fails after stop
	_, err = enc.EncodeFrame(frame)
	if err == nil {
		t.Fatal("expected EncodeFrame to fail after encoder stopped")
	}

	t.Log("Negative leg encoder test: fault injection path verified")
}
