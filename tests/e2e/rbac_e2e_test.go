package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/core/auth"
	"digital.vasic.auth/pkg/jwt"
)

func TestRBACEnforcement(t *testing.T) {
	jwtMgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
	middleware := auth.NewMiddleware(jwtMgr)

	// Protected admin endpoint
	adminHandler := middleware.ValidateAndEnforce("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"admin":true}`))
	}))

	// Protected player endpoint
	playerHandler := middleware.ValidateAndEnforce("player", "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"player":true}`))
	}))

	// Generate tokens
	playerToken, _ := jwtMgr.Create(map[string]interface{}{
		"sub":   "user-player",
		"roles": []string{"player"},
	})
	adminToken, _ := jwtMgr.Create(map[string]interface{}{
		"sub":   "user-admin",
		"roles": []string{"admin"},
	})

	// Test 1: Player cannot access admin endpoint
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+playerToken)
	rec := httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for player accessing admin, got %d", rec.Code)
	}

	// Test 2: Admin can access admin endpoint
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", rec.Code)
	}

	// Test 3: Player can access player endpoint
	req = httptest.NewRequest(http.MethodGet, "/player", nil)
	req.Header.Set("Authorization", "Bearer "+playerToken)
	rec = httptest.NewRecorder()
	playerHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for player, got %d", rec.Code)
	}

	// Test 4: Admin can access player endpoint (admin has broader access)
	req = httptest.NewRequest(http.MethodGet, "/player", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	playerHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin on player endpoint, got %d", rec.Code)
	}

	// Test 5: No token gets 401
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec = httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rec.Code)
	}
}
