package lifecycle

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/hosts/register" {
			t.Errorf("expected /api/v1/hosts/register, got %s", r.URL.Path)
		}

		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req["hardware_id"] != "hw-123" {
			t.Errorf("expected hardware_id hw-123, got %s", req["hardware_id"])
		}

		resp := map[string]string{
			"host_id":    "host-abc",
			"auth_token": "tok-def",
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	reg, err := Register(context.Background(), ts.URL, "hw-123", "bootstrap-xxx", "", "", "")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if reg.HostID != "host-abc" {
		t.Errorf("expected host_id host-abc, got %s", reg.HostID)
	}
	if reg.AuthToken != "tok-def" {
		t.Errorf("expected auth_token tok-def, got %s", reg.AuthToken)
	}
}

func TestRegisterFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, err := Register(context.Background(), ts.URL, "hw-123", "bad-token", "", "", "")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestHeartbeaterStartStop(t *testing.T) {
	beatCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/hosts/h-1/heartbeat" {
			beatCount++
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hb := NewHeartbeater(HeartbeatOptions{
		CoordinatorURL: ts.URL,
		HostID:         "h-1",
		AuthToken:      "tok",
		Interval:       200 * time.Millisecond,
	})

	hb.Start()
	if !hb.IsRunning() {
		t.Fatal("expected heartbeater to be running")
	}

	time.Sleep(550 * time.Millisecond)
	hb.Stop()

	if hb.IsRunning() {
		t.Fatal("expected heartbeater to be stopped")
	}
	if beatCount < 2 {
		t.Fatalf("expected at least 2 heartbeats, got %d", beatCount)
	}
}

func TestShutdownSequence(t *testing.T) {
	var order []int

	seq := NewShutdownSequence()
	seq.Add("step1", func() error {
		order = append(order, 1)
		return nil
	})
	seq.Add("step2", func() error {
		order = append(order, 2)
		return nil
	})
	seq.Add("step3", func() error {
		order = append(order, 3)
		return nil
	})

	errs := seq.Execute()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}

	// Steps execute in LIFO order: 3, 2, 1
	if len(order) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(order))
	}
	if order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Fatalf("expected LIFO order [3,2,1], got %v", order)
	}
}

func TestShutdownSequenceWithError(t *testing.T) {
	seq := NewShutdownSequence()
	seq.Add("ok", func() error { return nil })
	seq.Add("fail", func() error { return context.Canceled })

	errs := seq.Execute()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestShutdownDeregister(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/hosts/h-1/deregister" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer tok" {
				t.Errorf("expected Bearer tok, got %s", auth)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	if err := Shutdown(ts.URL, "h-1", "tok"); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}
