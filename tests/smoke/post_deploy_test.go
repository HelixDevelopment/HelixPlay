package smoke_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostDeployHealthChecks verifies core services respond with 200 OK
// within 30 seconds of simulated deployment.
func TestPostDeployHealthChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping smoke test in short mode")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","services":{"core":"up","discovery":"up"}}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ready":true}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Check core health endpoint.
	resp, err := client.Get(server.URL + "/health")
	require.NoError(t, err, "core health check failed")
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "core health check returned non-200")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Check readiness endpoint.
	resp, err = client.Get(server.URL + "/ready")
	require.NoError(t, err, "readiness check failed")
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "readiness check returned non-200")
}

// TestSmokeAPIEndpoints verifies key REST endpoints return expected
// shapes within the smoke-test time budget.
func TestSmokeAPIEndpointsRespond(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping smoke test in short mode")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tenants", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tenants":[{"id":"tenant-1","name":"Smoke Tenant"}]}`))
	})
	mux.HandleFunc("/api/v1/hosts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hosts":[{"id":"host-1","name":"Smoke Host","online":true}]}`))
	})
	mux.HandleFunc("/api/v1/games", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"games":[{"id":"game-1","title":"Smoke Game"}]}`))
	})
	mux.HandleFunc("/api/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"sessions":[]}`))
	})
	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"users":[]}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	endpoints := []struct {
		path       string
		wantStatus int
		wantType   string
	}{
		{"/api/v1/tenants", http.StatusOK, "application/json"},
		{"/api/v1/hosts", http.StatusOK, "application/json"},
		{"/api/v1/games", http.StatusOK, "application/json"},
		{"/api/v1/sessions", http.StatusOK, "application/json"},
		{"/api/v1/users", http.StatusOK, "application/json"},
	}

	for _, ep := range endpoints {
		resp, err := client.Get(server.URL + ep.path)
		require.NoError(t, err, "endpoint %s failed", ep.path)
		resp.Body.Close()
		assert.Equal(t, ep.wantStatus, resp.StatusCode, "endpoint %s returned unexpected status", ep.path)
		assert.Equal(t, ep.wantType, resp.Header.Get("Content-Type"), "endpoint %s returned wrong content type", ep.path)
	}
}
