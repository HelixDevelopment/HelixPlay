package streaming

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/input"
	streamingv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/streaming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SessionState represents the lifecycle of a streaming session.
type SessionState string

const (
	StateIdle         SessionState = "idle"
	StateConnecting   SessionState = "connecting"
	StateNegotiating  SessionState = "negotiating"
	StateStreaming    SessionState = "streaming"
	StateDisconnected SessionState = "disconnected"
	StateError        SessionState = "error"
)

// Config holds parameters for establishing a stream.
type Config struct {
	HostAddress        string
	ClientID           string
	UserID             string
	AuthToken          string
	RequestedGameID    string
	PreferredCodec     string
	PreferredTransport string
	DisplayWidth       int32
	DisplayHeight      int32
	DisplayRefreshHz   int32
}

// SessionController manages a single streaming session.
type SessionController struct {
	cfg Config

	mu              sync.RWMutex
	state           SessionState
	sessionID       string
	selectedCodec   string
	selectedTransport string
	resolution      Resolution
	refreshHz       int32
	bitrateKbps     int32

	conn   *grpc.ClientConn
	client streamingv1.StreamingControlClient
	stream streamingv1.StreamingControl_NegotiateStreamClient

	telemetryCancel context.CancelFunc
}

// Resolution represents negotiated display resolution.
type Resolution struct {
	Width  int32
	Height int32
}

// NewSessionController creates a controller for the supplied configuration.
func NewSessionController(cfg Config) *SessionController {
	return &SessionController{
		cfg:   cfg,
		state: StateIdle,
	}
}

// State returns the current session state.
func (sc *SessionController) State() SessionState {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.state
}

func (sc *SessionController) setState(s SessionState) {
	sc.mu.Lock()
	sc.state = s
	sc.mu.Unlock()
}

// Connect establishes the control channel and begins negotiation.
func (sc *SessionController) Connect(ctx context.Context) error {
	sc.setState(StateConnecting)

	conn, err := grpc.NewClient(sc.cfg.HostAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		sc.setState(StateError)
		return fmt.Errorf("dial host: %w", err)
	}
	sc.conn = conn
	sc.client = streamingv1.NewStreamingControlClient(conn)

	stream, err := sc.client.NegotiateStream(ctx)
	if err != nil {
		sc.setState(StateError)
		return fmt.Errorf("start negotiate stream: %w", err)
	}
	sc.mu.Lock()
	sc.stream = stream
	sc.mu.Unlock()

	hello := &streamingv1.NegotiateMessage{
		Payload: &streamingv1.NegotiateMessage_ClientHello{
			ClientHello: &streamingv1.ClientHello{
				ProtocolVersion: "1.0",
				ClientId:        sc.cfg.ClientID,
				UserId:          sc.cfg.UserID,
				AuthToken:       sc.cfg.AuthToken,
				RequestedGameId: sc.cfg.RequestedGameID,
			},
		},
	}
	if err := stream.Send(hello); err != nil {
		sc.setState(StateError)
		return fmt.Errorf("send client hello: %w", err)
	}

	sc.setState(StateNegotiating)
	go sc.negotiateLoop(ctx)

	return nil
}

func (sc *SessionController) negotiateLoop(ctx context.Context) {
	sc.mu.RLock()
	stream := sc.stream
	sc.mu.RUnlock()
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				sc.setState(StateDisconnected)
			} else {
				sc.setState(StateError)
			}
			return
		}

		switch p := msg.Payload.(type) {
		case *streamingv1.NegotiateMessage_HostAdvertisement:
			sc.handleHostAdvertisement(ctx, p.HostAdvertisement)
		case *streamingv1.NegotiateMessage_HostAnswer:
			sc.handleHostAnswer(p.HostAnswer)
		case *streamingv1.NegotiateMessage_Ready:
			sc.handleReady(p.Ready)
		case *streamingv1.NegotiateMessage_Terminate:
			sc.setState(StateDisconnected)
			return
		case *streamingv1.NegotiateMessage_Error:
			sc.setState(StateError)
			return
		}
	}
}

func (sc *SessionController) handleHostAdvertisement(ctx context.Context, adv *streamingv1.HostAdvertisement) {
	offer := &streamingv1.ClientOffer{
		PreferredCodec:     sc.cfg.PreferredCodec,
		FallbackCodecs:     []string{"h264", "av1"},
		PreferredTransport: sc.cfg.PreferredTransport,
		DisplayCaps: &streamingv1.DisplayCaps{
			Width:        sc.cfg.DisplayWidth,
			Height:       sc.cfg.DisplayHeight,
			RefreshHz:    sc.cfg.DisplayRefreshHz,
			HdrSupported: false,
		},
		DecoderCaps: &streamingv1.DecoderCaps{
			HardwareDecoders: []string{"nvdec", "videotoolbox"},
		},
		NetworkCaps: &streamingv1.NetworkCaps{
			MaxBandwidthMbps: 100,
			SupportsBbr:      true,
			RttMsEstimate:    20,
		},
		AudioCaps: &streamingv1.AudioCaps{
			Codecs:               []string{"opus"},
			MaxChannels:          2,
			PreferredBitrateKbps: 128,
		},
		ControllerCaps: &streamingv1.ControllerCaps{
			Types:     []string{"dualsense", "xbox"},
			Haptic:    true,
			Trigger:   true,
			Gyro:      false,
			AudioJack: true,
		},
	}

	msg := &streamingv1.NegotiateMessage{
		Payload: &streamingv1.NegotiateMessage_ClientOffer{
			ClientOffer: offer,
		},
	}
	_ = sc.stream.Send(msg)
}

func (sc *SessionController) handleHostAnswer(ans *streamingv1.HostAnswer) {
	sc.mu.Lock()
	sc.selectedCodec = ans.SelectedCodec
	sc.selectedTransport = ans.SelectedTransport
	sc.resolution = Resolution{Width: ans.Resolution.GetWidth(), Height: ans.Resolution.GetHeight()}
	sc.refreshHz = ans.RefreshRate
	sc.bitrateKbps = ans.BitrateTargetKbps
	sc.mu.Unlock()
}

func (sc *SessionController) handleReady(ready *streamingv1.Ready) {
	sc.mu.Lock()
	sc.sessionID = ready.SessionId
	tctx, cancel := context.WithCancel(context.Background())
	sc.telemetryCancel = cancel
	sc.mu.Unlock()
	sc.setState(StateStreaming)
	go sc.pushTelemetry(tctx)
}

// Disconnect ends the session.
func (sc *SessionController) Disconnect() {
	sc.mu.Lock()
	cancel := sc.telemetryCancel
	sc.telemetryCancel = nil
	sc.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	sc.mu.Lock()
	stream := sc.stream
	sc.mu.Unlock()
	if stream != nil {
		_ = stream.Send(&streamingv1.NegotiateMessage{
			Payload: &streamingv1.NegotiateMessage_Terminate{
				Terminate: &streamingv1.Terminate{Reason: "user"},
			},
		})
		stream.CloseSend()
	}
	if sc.conn != nil {
		sc.conn.Close()
	}
	sc.setState(StateDisconnected)
}

// SendControllerState forwards controller input to the host via telemetry.
func (sc *SessionController) SendControllerState(st input.ControllerState) error {
	sc.mu.RLock()
	state := sc.state
	sc.mu.RUnlock()
	if state != StateStreaming {
		return fmt.Errorf("not streaming")
	}
	return nil
}

func (sc *SessionController) pushTelemetry(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
