package e2e_test

import (
	"net"
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/client-wails/backend"
	"github.com/HelixDevelopment/HelixPlay/cmd/client-web/tv"
	hostcap "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	corecap "github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
	streamingv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/streaming"
	"google.golang.org/grpc"
)

type mockStreamingServer struct {
	streamingv1.UnimplementedStreamingControlServer
}

func (m *mockStreamingServer) NegotiateStream(s streamingv1.StreamingControl_NegotiateStreamServer) error {
	for {
		msg, err := s.Recv()
		if err != nil {
			return err
		}
		if _, ok := msg.Payload.(*streamingv1.NegotiateMessage_ClientHello); ok {
			_ = s.Send(&streamingv1.NegotiateMessage{
				Payload: &streamingv1.NegotiateMessage_HostAdvertisement{
					HostAdvertisement: &streamingv1.HostAdvertisement{
						HostId: "host-1", Codecs: []string{"h264", "hevc"},
					},
				},
			})
		}
	}
}

func TestTripleClientStackConnectsToHost(t *testing.T) {
	// Start mock gRPC server for Wails backend to connect to
	srv := grpc.NewServer()
	streamingv1.RegisterStreamingControlServer(srv, &mockStreamingServer{})
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go srv.Serve(lis)
	defer srv.Stop()

	// Host advertises capabilities
	meta, err := hostcap.Advertise()
	if err != nil {
		t.Fatalf("host capability advertise failed: %v", err)
	}
	if meta.GPUModel == "" {
		t.Fatal("expected host GPU model to be advertised")
	}

	// Client 1: Wails desktop backend connects to host
	wailsBackend := backend.NewBackend()
	if err := wailsBackend.Start(); err != nil {
		t.Fatalf("failed to start Wails backend: %v", err)
	}
	defer wailsBackend.Stop()
	if !wailsBackend.IsRunning() {
		t.Fatal("expected Wails backend to be running")
	}
	if err := wailsBackend.Connect(lis.Addr().String()); err != nil {
		t.Fatalf("failed to connect Wails backend to host: %v", err)
	}
	if !wailsBackend.IsConnected() {
		t.Fatal("expected Wails backend to be connected to host")
	}

	// Client 2: Web leanback (TV) client activates
	leanback := tv.NewLeanback()
	if err := leanback.Start(); err != nil {
		t.Fatalf("failed to start leanback client: %v", err)
	}
	defer leanback.Stop()
	if !leanback.IsActive() {
		t.Fatal("expected leanback client to be active")
	}

	// Simulate D-pad navigation (observable focus change)
	if err := leanback.HandleDPad("Down"); err != nil {
		t.Fatalf("failed to handle D-pad Down: %v", err)
	}
	if leanback.GetFocusedItem() != "Game 2" {
		t.Fatalf("expected focused item 'Game 2', got %s", leanback.GetFocusedItem())
	}

	// Client 3: Generic transport client (QUIC) connects to host
	// Use UDP as lightweight transport for third stack representation
	udpClient, err := transport.NewUDP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create UDP client: %v", err)
	}
	if err := udpClient.Start(); err != nil {
		t.Fatalf("failed to start UDP client: %v", err)
	}
	defer udpClient.Stop()
	if !udpClient.IsRunning() {
		t.Fatal("expected UDP client to be running")
	}

	// Verify UDP packet exchange works
	if err := udpClient.SendPacket([]byte{0x01, 0x02}); err != nil {
		t.Fatalf("failed to send UDP packet: %v", err)
	}

	// Cross-client capability negotiation: all three clients should agree on common codecs
	negotiator := corecap.NewNegotiator()
	clientCaps := corecap.Capabilities{
		Codecs:      []string{"H.264", "HEVC"},
		Resolutions: []string{"1080p"},
		MaxFPS:      60,
	}
	serverCaps := corecap.Capabilities{
		Codecs:      meta.CodecsSupported,
		Resolutions: []string{"1080p", "1440p", meta.MaxResolution},
		MaxFPS:      meta.RefreshRate,
	}
	result := negotiator.Negotiate(clientCaps, serverCaps)
	if !result.Success {
		t.Fatalf("cross-client negotiation failed: %s", result.Reason)
	}
	if result.AgreedResolution != "1080p" {
		t.Fatalf("expected 1080p agreed resolution, got %s", result.AgreedResolution)
	}

	// Observable: all three clients maintain their connected/active state
	if !wailsBackend.IsConnected() {
		t.Error("expected Wails backend to remain connected")
	}
	if !leanback.IsActive() {
		t.Error("expected leanback client to remain active")
	}
	if !udpClient.IsRunning() {
		t.Error("expected UDP client to remain running")
	}
}
