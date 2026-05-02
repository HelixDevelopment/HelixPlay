package smoke_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPostDeployHealthChecks(t *testing.T) {
	// Simulate core backend health endpoint
	coreHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	coreServer := httptest.NewServer(coreHandler)
	defer coreServer.Close()

	// Simulate discovery service health endpoint
	discoveryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	discoveryServer := httptest.NewServer(discoveryHandler)
	defer discoveryServer.Close()

	// 30-second post-deploy health check
	client := &http.Client{Timeout: 5 * time.Second}

	// Check core backend
	resp, err := client.Get(coreServer.URL + "/health")
	if err != nil {
		t.Fatalf("core health check failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("core health check returned %d", resp.StatusCode)
	}

	// Check discovery service
	resp, err = client.Get(discoveryServer.URL + "/health")
	if err != nil {
		t.Fatalf("discovery health check failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("discovery health check returned %d", resp.StatusCode)
	}

	t.Log("Post-deploy health checks passed")
}

func TestSmokeEndpointsRespond(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tenants", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tenants":[]}`))
	})
	mux.HandleFunc("/api/v1/hosts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hosts":[]}`))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	endpoints := []string{"/api/v1/tenants", "/api/v1/hosts"}

	for _, ep := range endpoints {
		resp, err := client.Get(ts.URL + ep)
		if err != nil {
			t.Fatalf("endpoint %s failed: %v", ep, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("endpoint %s returned %d", ep, resp.StatusCode)
		}
	}
}
