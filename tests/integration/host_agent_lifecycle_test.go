package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/lifecycle"
)

func TestHostAgentLifecycle(t *testing.T) {
	// Step 1: Mock coordinator server
	var registered atomic.Bool
	var heartbeatCount atomic.Int32
	var deregistered atomic.Bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/hosts/register":
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["hardware_id"] != "" && req["bootstrap_token"] != "" {
				registered.Store(true)
				resp := map[string]string{
					"host_id":    "host-123",
					"auth_token": "auth-tok",
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(resp)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		case "/api/v1/hosts/host-123/heartbeat":
			heartbeatCount.Add(1)
			w.WriteHeader(http.StatusOK)
		case "/api/v1/hosts/host-123/deregister":
			deregistered.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Step 2: Generate hardware fingerprint
	hwID, err := capability.Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint failed: %v", err)
	}
	if hwID == "" {
		t.Fatal("expected non-empty hardware ID")
	}

	// Step 3: Register with coordinator
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reg, err := lifecycle.Register(ctx, ts.URL, hwID, "bootstrap-token", "", "", "")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if !registered.Load() {
		t.Fatal("expected coordinator to receive registration")
	}
	if reg.HostID != "host-123" {
		t.Errorf("expected host-123, got %s", reg.HostID)
	}
	if reg.AuthToken != "auth-tok" {
		t.Errorf("expected auth-tok, got %s", reg.AuthToken)
	}

	// Step 4: Start heartbeat loop
	hb := lifecycle.NewHeartbeater(lifecycle.HeartbeatOptions{
		CoordinatorURL: ts.URL,
		HostID:         reg.HostID,
		AuthToken:      reg.AuthToken,
		Interval:       100 * time.Millisecond,
	})
	hb.Start()

	time.Sleep(350 * time.Millisecond)
	if heartbeatCount.Load() < 2 {
		t.Fatalf("expected at least 2 heartbeats, got %d", heartbeatCount.Load())
	}

	hb.Stop()

	// Step 5: Deregister / shutdown
	if err := lifecycle.Shutdown(ts.URL, reg.HostID, reg.AuthToken); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	if !deregistered.Load() {
		t.Fatal("expected coordinator to receive deregistration")
	}
}

func TestHostAgentRegistrationFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := lifecycle.Register(ctx, ts.URL, "hw-123", "bad-token", "", "", "")
	if err == nil {
		t.Fatal("expected error for failed registration")
	}
}
