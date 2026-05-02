package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"digital.vasic.auth/pkg/jwt"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/auth"
)

func TestMiddleware_ValidTokenWithRole(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
	token, err := mgr.Create(map[string]interface{}{
		"sub":       "user-1",
		"email":     "user@example.com",
		"tenant_id": "tenant-1",
		"roles":     []string{"admin", "player"},
	})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	mw := auth.NewMiddleware(mgr)

	var captured *auth.UserIdentity
	handler := mw.ValidateAndEnforce("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = auth.UserIdentityFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if captured == nil {
		t.Fatal("expected user identity in context")
	}
	if captured.Subject != "user-1" {
		t.Errorf("expected subject user-1, got %s", captured.Subject)
	}
	if captured.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", captured.Email)
	}
	if captured.TenantID != "tenant-1" {
		t.Errorf("expected tenant_id tenant-1, got %s", captured.TenantID)
	}
	if len(captured.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(captured.Roles))
	}
}

func TestMiddleware_InsufficientRole(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
	token, _ := mgr.Create(map[string]interface{}{
		"sub":   "user-1",
		"roles": []string{"player"},
	})

	mw := auth.NewMiddleware(mgr)
	handler := mw.ValidateAndEnforce("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
	mw := auth.NewMiddleware(mgr)
	handler := mw.ValidateOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
	mw := auth.NewMiddleware(mgr)
	handler := mw.ValidateOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiddleware_ValidTokenNoRolesRequired(t *testing.T) {
	mgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
	token, _ := mgr.Create(map[string]interface{}{
		"sub": "user-1",
	})

	mw := auth.NewMiddleware(mgr)
	var captured *auth.UserIdentity
	handler := mw.ValidateOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = auth.UserIdentityFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if captured == nil || captured.Subject != "user-1" {
		t.Error("expected user identity to be injected")
	}
}

func TestMiddleware_WrongSecret(t *testing.T) {
	mgr1 := jwt.NewManager(jwt.DefaultConfig("secret-a"))
	mgr2 := jwt.NewManager(jwt.DefaultConfig("secret-b"))

	token, _ := mgr1.Create(map[string]interface{}{"sub": "user-1"})

	mw := auth.NewMiddleware(mgr2)
	handler := mw.ValidateOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong secret, got %d", rr.Code)
	}
}

func TestWithUserIdentity(t *testing.T) {
	identity := &auth.UserIdentity{Subject: "injected"}
	ctx := auth.WithUserIdentity(t.Context(), identity)

	retrieved := auth.UserIdentityFromContext(ctx)
	if retrieved == nil || retrieved.Subject != "injected" {
		t.Error("expected injected identity to be retrievable")
	}
}

func TestUserIdentityFromContext_Missing(t *testing.T) {
	retrieved := auth.UserIdentityFromContext(t.Context())
	if retrieved != nil {
		t.Error("expected nil when no identity is in context")
	}
}
