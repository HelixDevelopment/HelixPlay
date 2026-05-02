package main_test

import (
	"net"
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/client-wails/backend"
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

func TestWailsStart(t *testing.T) {
	b := backend.NewBackend()
	if err := b.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !b.IsRunning() {
		t.Error("Expected backend to be running")
	}
}

func TestWailsStop(t *testing.T) {
	b := backend.NewBackend()
	b.Start()
	b.Stop()
	if b.IsRunning() {
		t.Error("Expected backend to be stopped")
	}
}

func TestWailsConnect(t *testing.T) {
	addr := startMockStreamingServer(t)
	b := backend.NewBackend()
	b.Start()
	err := b.Connect(addr)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	if !b.IsConnected() {
		t.Error("Expected to be connected")
	}
}
