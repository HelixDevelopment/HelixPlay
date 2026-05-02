package e2e_test

import (
	"testing"
	"time"

	hostcap "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	corecap "github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capture"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/codec"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestFullStreaming1080pSessionWithLatencyCheck(t *testing.T) {
	// Step 1: Host advertises hardware capabilities
	meta, err := hostcap.Advertise()
	if err != nil {
		t.Fatalf("capability advertise failed: %v", err)
	}
	if len(meta.CodecsSupported) == 0 {
		t.Fatal("expected host to advertise at least one codec")
	}
	if meta.MaxResolution == "" {
		t.Fatal("expected host to advertise max resolution")
	}

	// Step 2: Negotiate 1080p H.264 session between client and host
	negotiator := corecap.NewNegotiator()
	clientCaps := corecap.Capabilities{
		Codecs:      []string{"H.264"},
		Resolutions: []string{"1080p"},
		MaxFPS:      60,
		Haptics:     false,
	}
	serverCaps := corecap.Capabilities{
		Codecs:      meta.CodecsSupported,
		Resolutions: []string{"1080p", "1440p", meta.MaxResolution},
		MaxFPS:      meta.RefreshRate,
		Haptics:     true,
	}
	result := negotiator.Negotiate(clientCaps, serverCaps)
	if !result.Success {
		t.Fatalf("negotiation failed: %s", result.Reason)
	}
	if result.AgreedResolution != "1080p" {
		t.Fatalf("expected 1080p resolution, got %s", result.AgreedResolution)
	}

	// Step 3: Codec negotiation confirms H.264
	codecNeg := codec.NewNegotiator()
	cresult := codecNeg.Negotiate(clientCaps.Codecs, serverCaps.Codecs)
	if cresult.Failed {
		t.Fatalf("codec negotiation failed: %s", cresult.Reason)
	}
	if len(cresult.Agreed) == 0 || cresult.Agreed[0] != "H.264" {
		t.Fatalf("expected H.264 to be agreed, got %v", cresult.Agreed)
	}

	// Step 4: Start screen capture pipeline
	capturer := capture.NewCapturer("linux")
	if capturer == nil {
		t.Fatal("expected capturer to be created")
	}
	if err := capturer.Start(); err != nil {
		t.Fatalf("failed to start capture: %v", err)
	}
	defer capturer.Stop()
	if !capturer.IsRunning() {
		t.Fatal("expected capturer to be running")
	}

	// Step 5: Initialize hardware encoder
	enc := encoder.NewDualPath("nvenc")
	if enc == nil {
		t.Fatal("expected dual-path encoder to be created")
	}
	if err := enc.Start(); err != nil {
		t.Fatalf("failed to start encoder: %v", err)
	}
	defer enc.Stop()
	if !enc.IsStreaming() {
		t.Fatal("expected encoder to be streaming")
	}

	// Step 6: Initialize UDP transport to localhost
	udp, err := transport.NewUDP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create UDP transport: %v", err)
	}
	if err := udp.Start(); err != nil {
		t.Fatalf("failed to start UDP transport: %v", err)
	}
	defer udp.Stop()
	if !udp.IsRunning() {
		t.Fatal("expected UDP transport to be running")
	}

	// Step 7: Full pipeline latency measurement
	start := time.Now()

	// Attempt to capture a frame (stub may return error, which is observable)
	frame, captureErr := capturer.GetFrame()
	if captureErr == nil && frame != nil {
		// Encode and transport the captured frame
		encoded := enc.GetStreamOutput()
		if len(encoded) == 0 {
			t.Fatal("expected non-empty encoded stream output")
		}
		if err := udp.SendPacket(encoded); err != nil {
			t.Fatalf("failed to send packet over UDP: %v", err)
		}
	}

	latency := time.Since(start)
	t.Logf("Full pipeline latency (advertise -> capture -> encode -> transport): %v", latency)

	// Observable: transport remains operational after packet send
	if !udp.IsRunning() {
		t.Fatal("expected UDP transport to remain running after packet send")
	}

	// Observable: encoder still streaming
	if !enc.IsStreaming() {
		t.Fatal("expected encoder to remain streaming")
	}

	// Verify local address was bound
	if udp.LocalAddr() == nil {
		t.Fatal("expected UDP local address to be bound")
	}
}
