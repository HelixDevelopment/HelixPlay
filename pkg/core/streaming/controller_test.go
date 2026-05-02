package streaming

import (
	"context"
	"net"
	"testing"
	"time"

	streamingv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/streaming"
	"google.golang.org/grpc"
)

type mockStreamingServer struct {
	streamingv1.UnimplementedStreamingControlServer
	recvChan chan *streamingv1.NegotiateMessage
	sendChan chan *streamingv1.NegotiateMessage
}

func (m *mockStreamingServer) NegotiateStream(s streamingv1.StreamingControl_NegotiateStreamServer) error {
	go func() {
		for {
			msg, err := s.Recv()
			if err != nil {
				return
			}
			m.recvChan <- msg
		}
	}()

	for msg := range m.sendChan {
		if err := s.Send(msg); err != nil {
			return err
		}
	}
	return nil
}

func TestSessionControllerConnectAndNegotiate(t *testing.T) {
	recvCh := make(chan *streamingv1.NegotiateMessage, 4)
	sendCh := make(chan *streamingv1.NegotiateMessage, 4)

	srv := grpc.NewServer()
	mock := &mockStreamingServer{recvChan: recvCh, sendChan: sendCh}
	streamingv1.RegisterStreamingControlServer(srv, mock)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	go srv.Serve(lis)
	defer srv.Stop()

	cfg := Config{
		HostAddress:        lis.Addr().String(),
		ClientID:           "client-1",
		UserID:             "user-1",
		AuthToken:          "token",
		RequestedGameID:    "game-1",
		PreferredCodec:     "hevc",
		PreferredTransport: "quic",
		DisplayWidth:       1920,
		DisplayHeight:      1080,
		DisplayRefreshHz:   60,
	}

	sc := NewSessionController(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sc.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer sc.Disconnect()

	select {
	case msg := <-recvCh:
		if _, ok := msg.Payload.(*streamingv1.NegotiateMessage_ClientHello); !ok {
			t.Fatalf("Expected ClientHello, got %T", msg.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for ClientHello")
	}

	sendCh <- &streamingv1.NegotiateMessage{
		Payload: &streamingv1.NegotiateMessage_HostAdvertisement{
			HostAdvertisement: &streamingv1.HostAdvertisement{
				HostId: "host-1",
				Codecs: []string{"hevc", "h264"},
			},
		},
	}

	select {
	case msg := <-recvCh:
		if _, ok := msg.Payload.(*streamingv1.NegotiateMessage_ClientOffer); !ok {
			t.Fatalf("Expected ClientOffer, got %T", msg.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for ClientOffer")
	}

	sendCh <- &streamingv1.NegotiateMessage{
		Payload: &streamingv1.NegotiateMessage_HostAnswer{
			HostAnswer: &streamingv1.HostAnswer{
				SelectedCodec:     "hevc",
				SelectedTransport: "quic",
				Resolution:        &streamingv1.Resolution{Width: 1920, Height: 1080},
				RefreshRate:       60,
			},
		},
	}

	sendCh <- &streamingv1.NegotiateMessage{
		Payload: &streamingv1.NegotiateMessage_Ready{
			Ready: &streamingv1.Ready{SessionId: "session-1"},
		},
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sc.State() == StateStreaming {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if sc.State() != StateStreaming {
		t.Fatalf("Expected state streaming, got %s", sc.State())
	}
}
