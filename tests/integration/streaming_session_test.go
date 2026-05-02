package integration_test

import (
	"testing"

	hostcap "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	corecap "github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/codec"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
)

func TestStreamingSessionCreateNegotiateTerminate(t *testing.T) {
	// Step 1: Host advertises capabilities (observable: non-empty metadata)
	meta, err := hostcap.Advertise()
	if err != nil {
		t.Fatalf("capability advertise failed: %v", err)
	}
	if meta.GPUModel == "" {
		t.Fatal("expected GPU model in advertised capabilities")
	}
	if len(meta.CodecsSupported) == 0 {
		t.Fatal("expected at least one advertised codec")
	}

	// Step 2: Core protocol negotiation between client and server
	negotiator := corecap.NewNegotiator()
	clientCaps := corecap.Capabilities{
		Codecs:           []string{"H.264", "HEVC"},
		Resolutions:      []string{"1080p", "4K"},
		MaxFPS:           60,
		Haptics:          true,
		AdaptiveTriggers: true,
	}
	serverCaps := corecap.Capabilities{
		Codecs:           meta.CodecsSupported,
		Resolutions:      []string{"720p", "1080p", "1440p", meta.MaxResolution},
		MaxFPS:           meta.RefreshRate,
		Haptics:          true,
		AdaptiveTriggers: true,
	}
	result := negotiator.Negotiate(clientCaps, serverCaps)
	if !result.Success {
		t.Fatalf("core negotiation failed: %s", result.Reason)
	}
	if result.AgreedResolution == "" {
		t.Fatal("expected agreed resolution to be non-empty")
	}
	if result.AgreedFPS <= 0 {
		t.Fatal("expected agreed FPS to be positive")
	}

	// Step 3: Codec-level negotiation
	codecNeg := codec.NewNegotiator()
	codecResult := codecNeg.Negotiate(clientCaps.Codecs, serverCaps.Codecs)
	if codecResult.Failed {
		t.Fatalf("codec negotiation failed: %s", codecResult.Reason)
	}
	if len(codecResult.Agreed) == 0 {
		t.Fatal("expected at least one agreed codec")
	}

	// Step 4: Create WebRTC transport session
	webrtc, err := transport.NewWebRTC("nvenc", codecResult.Agreed[0])
	if err != nil {
		t.Fatalf("failed to create WebRTC transport: %v", err)
	}
	if err := webrtc.Start(); err != nil {
		t.Fatalf("failed to start WebRTC transport: %v", err)
	}
	if !webrtc.IsConnected() {
		t.Fatal("expected WebRTC transport to be connected")
	}

	// Step 5: Start dual-path encoder
	enc := encoder.NewDualPath("nvenc")
	if enc == nil {
		t.Fatal("expected dual-path encoder to be created")
	}
	if err := enc.Start(); err != nil {
		t.Fatalf("failed to start encoder: %v", err)
	}
	if !enc.IsStreaming() {
		t.Fatal("expected encoder to be streaming")
	}
	if len(enc.GetStreamOutput()) == 0 {
		t.Fatal("expected non-empty stream output after encoder start")
	}

	// Step 6: Terminate session (observable: state transitions)
	webrtc.Stop()
	enc.Stop()

	if webrtc.IsConnected() {
		t.Error("expected WebRTC to be disconnected after termination")
	}
	if enc.IsStreaming() {
		t.Error("expected encoder streaming to be stopped after termination")
	}
	if enc.IsRecording() {
		t.Error("expected encoder recording to be stopped after termination")
	}
}
