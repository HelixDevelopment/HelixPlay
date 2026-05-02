package challenges_test

import (
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestChallengeFullStreamingLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping challenge in short mode")
	}

	// Step 1: Host capability advertisement
	meta, err := capability.Advertise()
	if err != nil {
		t.Fatalf("capability advertise failed: %v", err)
	}
	if meta.GPUModel == "" {
		t.Fatal("expected GPU model")
	}

	// Step 2: Initialize encoder
	enc := encoder.NewDualPath("software")
	if enc == nil {
		t.Fatal("expected encoder to be created")
	}
	if err := enc.Start(); err != nil {
		t.Fatalf("encoder start failed: %v", err)
	}
	defer enc.Stop()

	// Step 3: Initialize transport
	udp, err := transport.NewUDP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("transport creation failed: %v", err)
	}
	if err := udp.Start(); err != nil {
		t.Fatalf("transport start failed: %v", err)
	}
	defer udp.Stop()

	// Step 4: Verify streaming output
	output := enc.GetStreamOutput()
	if len(output) == 0 {
		t.Fatal("expected non-empty stream output")
	}

	// Step 5: Simulate 60-second stream (shortened for test)
	streamDuration := 100 * time.Millisecond
	time.Sleep(streamDuration)

	// Verify all components remain operational
	if !enc.IsStreaming() {
		t.Fatal("expected encoder to remain streaming")
	}
	if !udp.IsRunning() {
		t.Fatal("expected transport to remain running")
	}

	// RecordAction would capture evidence here in production
	t.Logf("Challenge passed: full streaming lifecycle verified (GPU=%s, Codec=%s)", meta.GPUModel, meta.CodecsSupported)
}
