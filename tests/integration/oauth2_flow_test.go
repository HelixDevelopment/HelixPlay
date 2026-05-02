package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"digital.vasic.auth/pkg/jwt"
)

func TestOAuth2FullFlow(t *testing.T) {
	// Mock OAuth2 authorization server
	var authCode string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/authorize":
			// Authorization endpoint: redirect with code
			state := r.URL.Query().Get("state")
			authCode = "mock-auth-code-" + state
			redirectURI := r.URL.Query().Get("redirect_uri")
			http.Redirect(w, r, redirectURI+"?code="+authCode+"&state="+state, http.StatusFound)

		case "/oauth/token":
			// Token endpoint: exchange code for tokens
			var req struct {
				GrantType   string `json:"grant_type"`
				Code        string `json:"code"`
				RedirectURI string `json:"redirect_uri"`
				ClientID    string `json:"client_id"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			if req.GrantType != "authorization_code" {
				http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
				return
			}
			if !strings.HasPrefix(req.Code, "mock-auth-code-") {
				http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
				return
			}

			// Generate JWT access token
			mgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
			accessToken, _ := mgr.Create(map[string]interface{}{
				"sub":   "user-123",
				"email": "user@example.com",
				"roles": []string{"player"},
			})

			resp := map[string]interface{}{
				"access_token": accessToken,
				"token_type":   "Bearer",
				"expires_in":   3600,
				"refresh_token": "refresh-xyz",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)

		case "/oauth/userinfo":
			// UserInfo endpoint
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"error":"invalid_token"}`, http.StatusUnauthorized)
				return
			}
			resp := map[string]interface{}{
				"sub":   "user-123",
				"email": "user@example.com",
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Simulate authorization request
	state := "random-state-123"
	authURL := ts.URL + "/oauth/authorize?response_type=code&client_id=helixplay&redirect_uri=" + ts.URL + "/callback&state=" + state
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("authorize request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect, got %d", resp.StatusCode)
	}

	// Step 2: Extract code and exchange for token
	if authCode == "" {
		t.Fatal("expected auth code to be set")
	}

	tokenReq := map[string]string{
		"grant_type":   "authorization_code",
		"code":         authCode,
		"redirect_uri": ts.URL + "/callback",
		"client_id":    "helixplay",
	}
	body, _ := json.Marshal(tokenReq)
	tokenHTTPReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/oauth/token", strings.NewReader(string(body)))
	tokenHTTPReq.Header.Set("Content-Type", "application/json")
	tokenResp, err := client.Do(tokenHTTPReq)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	if tokenResult.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if tokenResult.TokenType != "Bearer" {
		t.Errorf("expected Bearer, got %s", tokenResult.TokenType)
	}

	// Step 3: Use access token to fetch user info
	userReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/oauth/userinfo", nil)
	userReq.Header.Set("Authorization", "Bearer "+tokenResult.AccessToken)
	userResp, err := client.Do(userReq)
	if err != nil {
		t.Fatalf("userinfo request failed: %v", err)
	}
	defer userResp.Body.Close()

	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		t.Fatalf("failed to decode userinfo: %v", err)
	}
	if userInfo.Sub != "user-123" {
		t.Errorf("expected user-123, got %s", userInfo.Sub)
	}
}
