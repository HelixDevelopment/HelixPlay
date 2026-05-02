package session

import (
	"context"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
	corecap "github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
)

func TestManagerCreate(t *testing.T) {
	m := NewManager(nil, nil)
	sess, err := m.Create("client-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sess.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	if sess.ClientID != "client-1" {
		t.Errorf("expected client-1, got %s", sess.ClientID)
	}
	if sess.Status != StatusConnecting {
		t.Errorf("expected status connecting, got %s", sess.Status)
	}
	if m.Count() != 1 {
		t.Fatalf("expected 1 session, got %d", m.Count())
	}
}

func TestManagerNegotiate(t *testing.T) {
	m := NewManager(nil, nil)
	sess, _ := m.Create("client-1")

	clientCaps := corecap.Capabilities{
		Codecs:      []string{"H.264"},
		Resolutions: []string{"1080p"},
		MaxFPS:      60,
	}

	if err := m.Negotiate(sess.ID, clientCaps); err != nil {
		t.Fatalf("Negotiate failed: %v", err)
	}

	sess.mu.RLock()
	status := sess.Status
	codec := sess.Codec
	res := sess.Resolution
	sess.mu.RUnlock()

	if status != StatusStreaming {
		t.Errorf("expected status streaming, got %s", status)
	}
	if codec == "" {
		t.Error("expected non-empty codec")
	}
	if res == "" {
		t.Error("expected non-empty resolution")
	}
}

func TestManagerNegotiateWrongState(t *testing.T) {
	m := NewManager(nil, nil)
	sess, _ := m.Create("client-1")

	clientCaps := corecap.Capabilities{
		Codecs:      []string{"H.264"},
		Resolutions: []string{"1080p"},
		MaxFPS:      60,
	}

	if err := m.Negotiate(sess.ID, clientCaps); err != nil {
		t.Fatalf("first negotiate failed: %v", err)
	}

	// Second negotiate should fail because status is now streaming
	if err := m.Negotiate(sess.ID, clientCaps); err == nil {
		t.Fatal("expected error for second negotiate")
	}
}

func TestManagerTerminate(t *testing.T) {
	m := NewManager(nil, nil)
	sess, _ := m.Create("client-1")

	if err := m.Terminate(sess.ID); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}

	if m.Count() != 0 {
		t.Fatalf("expected 0 sessions, got %d", m.Count())
	}

	if sess.IsActive() {
		t.Fatal("expected session to be inactive")
	}
	if sess.StatusValue() != StatusTerminated {
		t.Errorf("expected status terminated, got %s", sess.StatusValue())
	}
}

func TestManagerGet(t *testing.T) {
	m := NewManager(nil, nil)
	sess, _ := m.Create("client-1")

	got, ok := m.Get(sess.ID)
	if !ok {
		t.Fatal("expected session to be found")
	}
	if got.ID != sess.ID {
		t.Errorf("expected %s, got %s", sess.ID, got.ID)
	}

	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("expected session not to be found")
	}
}

func TestManagerList(t *testing.T) {
	m := NewManager(nil, nil)
	m.Create("client-1")
	m.Create("client-2")

	list := m.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(list))
	}
}

func TestManagerMonitor(t *testing.T) {
	m := NewManager(nil, nil)
	sess, _ := m.Create("client-1")

	clientCaps := corecap.Capabilities{
		Codecs:      []string{"H.264"},
		Resolutions: []string{"1080p"},
		MaxFPS:      60,
	}
	_ = m.Negotiate(sess.ID, clientCaps)

	// Artificially age the session
	sess.StartedAt = time.Now().UTC().Add(-2 * time.Hour)

	terminated, err := m.Monitor(context.Background(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Monitor failed: %v", err)
	}
	if terminated != 1 {
		t.Fatalf("expected 1 terminated session, got %d", terminated)
	}
	if m.Count() != 0 {
		t.Fatalf("expected 0 sessions after monitor, got %d", m.Count())
	}
}

func TestManagerMonitorNotIdle(t *testing.T) {
	m := NewManager(nil, nil)
	sess, _ := m.Create("client-1")

	clientCaps := corecap.Capabilities{
		Codecs:      []string{"H.264"},
		Resolutions: []string{"1080p"},
		MaxFPS:      60,
	}
	_ = m.Negotiate(sess.ID, clientCaps)

	terminated, err := m.Monitor(context.Background(), 1*time.Hour)
	if err != nil {
		t.Fatalf("Monitor failed: %v", err)
	}
	if terminated != 0 {
		t.Fatalf("expected 0 terminated sessions, got %d", terminated)
	}
}

func TestManagerTerminateUnknown(t *testing.T) {
	m := NewManager(nil, nil)
	if err := m.Terminate("nonexistent"); err == nil {
		t.Fatal("expected error for unknown session")
	}
}

func TestManagerWithEncoderAndTransport(t *testing.T) {
	enc := encoder.NewDualPath("software")
	if enc == nil {
		t.Fatal("expected encoder to be created")
	}
	udp, err := transport.NewUDP("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create UDP transport: %v", err)
	}

	m := NewManager(enc, udp)
	sess, err := m.Create("client-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sess == nil {
		t.Fatal("expected non-nil session")
	}
}
