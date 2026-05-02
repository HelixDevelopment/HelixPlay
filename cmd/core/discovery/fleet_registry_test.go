package discovery

import (
	"testing"
)

func TestRegistryRegisterGet(t *testing.T) {
	r := NewRegistry()
	h := FleetHost{ID: "h1", Name: "Host1", Address: "10.0.0.1", Port: 50051, MaxSessions: 4}
	if err := r.Register(h); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got, ok := r.Get("h1")
	if !ok {
		t.Fatal("expected host to be found")
	}
	if got.Name != "Host1" {
		t.Errorf("expected Host1, got %s", got.Name)
	}
}

func TestRegistryRegisterEmptyID(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(FleetHost{}); err == nil {
		t.Fatal("expected error for empty host ID")
	}
}

func TestRegistryDeregister(t *testing.T) {
	r := NewRegistry()
	r.Register(FleetHost{ID: "h1", Name: "Host1", MaxSessions: 1})
	r.Deregister("h1")

	_, ok := r.Get("h1")
	if ok {
		t.Error("expected host to be removed")
	}
}

func TestRegistryListAvailable(t *testing.T) {
	r := NewRegistry()
	r.Register(FleetHost{ID: "h1", Status: "online", ActiveSessions: 0, MaxSessions: 2})
	r.Register(FleetHost{ID: "h2", Status: "online", ActiveSessions: 2, MaxSessions: 2})
	r.Register(FleetHost{ID: "h3", Status: "offline", ActiveSessions: 0, MaxSessions: 2})

	available := r.ListAvailable()
	if len(available) != 1 {
		t.Fatalf("expected 1 available host, got %d", len(available))
	}
	if available[0].ID != "h1" {
		t.Errorf("expected h1, got %s", available[0].ID)
	}
}

func TestRegistryListByRegion(t *testing.T) {
	r := NewRegistry()
	r.Register(FleetHost{ID: "h1", Region: "us-east"})
	r.Register(FleetHost{ID: "h2", Region: "eu-west"})
	r.Register(FleetHost{ID: "h3", Region: "us-east"})

	usHosts := r.ListByRegion("us-east")
	if len(usHosts) != 2 {
		t.Fatalf("expected 2 us-east hosts, got %d", len(usHosts))
	}
}

func TestRegistrySessionTracking(t *testing.T) {
	r := NewRegistry()
	r.Register(FleetHost{ID: "h1", ActiveSessions: 0, MaxSessions: 2})

	if !r.IncrementSessions("h1") {
		t.Fatal("expected increment to succeed")
	}
	h, _ := r.Get("h1")
	if h.ActiveSessions != 1 {
		t.Errorf("expected 1 session, got %d", h.ActiveSessions)
	}

	if !r.IncrementSessions("h1") {
		t.Fatal("expected second increment to succeed")
	}
	if r.IncrementSessions("h1") {
		t.Fatal("expected third increment to fail (quota exceeded)")
	}

	if !r.DecrementSessions("h1") {
		t.Fatal("expected decrement to succeed")
	}
	if h.ActiveSessions != 1 {
		t.Errorf("expected 1 session after decrement, got %d", h.ActiveSessions)
	}
}

func TestRegistryHeartbeat(t *testing.T) {
	r := NewRegistry()
	r.Register(FleetHost{ID: "h1"})

	if !r.UpdateHeartbeat("h1") {
		t.Fatal("expected heartbeat update to succeed")
	}
	h, _ := r.Get("h1")
	if h.LastHeartbeat.IsZero() {
		t.Error("expected heartbeat to be set")
	}

	if r.UpdateHeartbeat("unknown") {
		t.Error("expected heartbeat update to fail for unknown host")
	}
}
