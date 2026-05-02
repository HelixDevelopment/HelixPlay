package backend

import (
	"net"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/input"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/streaming"
	streamingv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/streaming"
	"google.golang.org/grpc"
)

type mockStreamingServer struct {
	streamingv1.UnimplementedStreamingControlServer
}

func (m *mockStreamingServer) NegotiateStream(s streamingv1.StreamingControl_NegotiateStreamServer) error {
	// Echo back a Ready message to complete negotiation quickly.
	for {
		msg, err := s.Recv()
		if err != nil {
			return err
		}
		switch msg.Payload.(type) {
		case *streamingv1.NegotiateMessage_ClientHello:
			_ = s.Send(&streamingv1.NegotiateMessage{
				Payload: &streamingv1.NegotiateMessage_HostAdvertisement{
					HostAdvertisement: &streamingv1.HostAdvertisement{
						HostId: "mock-host",
						Codecs: []string{"h264"},
					},
				},
			})
		case *streamingv1.NegotiateMessage_ClientOffer:
			_ = s.Send(&streamingv1.NegotiateMessage{
				Payload: &streamingv1.NegotiateMessage_HostAnswer{
					HostAnswer: &streamingv1.HostAnswer{
						SelectedCodec:     "h264",
						SelectedTransport: "udp",
						Resolution:        &streamingv1.Resolution{Width: 1280, Height: 720},
						RefreshRate:       60,
					},
				},
			})
			_ = s.Send(&streamingv1.NegotiateMessage{
				Payload: &streamingv1.NegotiateMessage_Ready{
					Ready: &streamingv1.Ready{SessionId: "mock-session"},
				},
			})
		}
	}
}

func startMockStreamingServer(t *testing.T) string {
	t.Helper()
	srv := grpc.NewServer()
	streamingv1.RegisterStreamingControlServer(srv, &mockStreamingServer{})
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	go srv.Serve(lis)
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

func TestBackendStartStop(t *testing.T) {
	b := NewBackend()
	if err := b.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !b.IsRunning() {
		t.Error("Expected backend to be running")
	}
	b.Stop()
	if b.IsRunning() {
		t.Error("Expected backend to be stopped")
	}
}

func TestBackendConnectDisconnect(t *testing.T) {
	addr := startMockStreamingServer(t)
	b := NewBackend()
	if err := b.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer b.Stop()

	err := b.Connect(addr)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait briefly for negotiation to complete
	time.Sleep(200 * time.Millisecond)
	if b.StreamingState() != streaming.StateStreaming {
		t.Errorf("Expected state streaming, got %s", b.StreamingState())
	}
	b.Disconnect()
	if b.IsConnected() {
		t.Error("Expected not connected after disconnect")
	}
}

func TestBackendControllerInput(t *testing.T) {
	b := NewBackend()
	if err := b.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer b.Stop()

	st := input.ControllerState{
		ID:         0,
		Connected:  true,
		ButtonMask: 0x1000,
	}
	if err := b.UpdateControllerState(st); err != nil {
		t.Fatalf("UpdateControllerState failed: %v", err)
	}

	states := b.ControllerStates()
	if len(states) != 1 {
		t.Fatalf("Expected 1 controller state, got %d", len(states))
	}
	if states[0].ButtonMask != 0x1000 {
		t.Errorf("Expected ButtonMask 0x1000, got 0x%04x", states[0].ButtonMask)
	}
}

func TestBackendDiscoveryStartStop(t *testing.T) {
	b := NewBackend()
	if err := b.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer b.Stop()

	if err := b.StartDiscovery("", "tenant-1", ""); err != nil {
		t.Fatalf("StartDiscovery failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	hosts := b.DiscoveredHosts()
	_ = hosts

	b.StopDiscovery()
}
