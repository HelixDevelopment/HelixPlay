package handlers

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/streaming"
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

func TestStreamingNegotiate(t *testing.T) {
	addr := startMockStreamingServer(t)
	h := &Streaming{}
	body, _ := json.Marshal(map[string]string{
		"host_address": addr,
		"client_id":    "web-1",
		"user_id":      "user-1",
		"auth_token":   "tok",
		"game_id":      "game-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/streaming/negotiate", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Negotiate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp["state"] != string(streaming.StateNegotiating) {
		t.Errorf("Expected state negotiating, got %s", resp["state"])
	}
}

func TestStreamingControllerUpdate(t *testing.T) {
	h := &Streaming{}
	body, _ := json.Marshal(map[string]any{
		"id":         0,
		"connected":  true,
		"buttonMask": 0x1000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/streaming/controller", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.ControllerUpdate(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", w.Code)
	}
}

func TestStreamingState(t *testing.T) {
	h := &Streaming{}
	req := httptest.NewRequest(http.MethodGet, "/api/streaming/state", nil)
	w := httptest.NewRecorder()

	h.State(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp["state"] != "idle" {
		t.Errorf("Expected idle state, got %s", resp["state"])
	}
}

func TestCatalogGamesMissingTenant(t *testing.T) {
	h := &Catalog{}
	req := httptest.NewRequest(http.MethodGet, "/api/catalog/games", nil)
	w := httptest.NewRecorder()

	h.Games(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}
}
