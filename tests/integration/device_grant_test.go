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

func TestDeviceAuthorizationGrantFlow(t *testing.T) {
	// Mock device authorization server (RFC 8628)
	deviceCode := "device-abc-123"
	userCode := "ABCD-EFGH"
	var tokenExchanged bool
	baseURL := ""

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/device_authorization", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"device_code":               deviceCode,
			"user_code":                 userCode,
			"verification_uri":          baseURL + "/activate",
			"verification_uri_complete": baseURL + "/activate?user_code=" + userCode,
			"expires_in":                600,
			"interval":                  1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			GrantType  string `json:"grant_type"`
			DeviceCode string `json:"device_code"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.GrantType != "urn:ietf:params:oauth:grant-type:device_code" {
			http.Error(w, `{"error":"unsupported_grant_type"}`, http.StatusBadRequest)
			return
		}

		mgr := jwt.NewManager(jwt.DefaultConfig("test-secret"))
		accessToken, _ := mgr.Create(map[string]interface{}{
			"sub":   "user-device-1",
			"scope": "streaming",
		})

		tokenExchanged = true
		resp := map[string]interface{}{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()
	baseURL = ts.URL

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Request device code
	deviceReq := map[string]string{
		"client_id": "helixplay-device",
		"scope":     "streaming",
	}
	body, _ := json.Marshal(deviceReq)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/oauth/device_authorization", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("device authorization request failed: %v", err)
	}
	defer resp.Body.Close()

	var deviceResp struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		t.Fatalf("failed to decode device response: %v", err)
	}
	if deviceResp.DeviceCode == "" {
		t.Fatal("expected device_code")
	}
	if deviceResp.UserCode == "" {
		t.Fatal("expected user_code")
	}
	if deviceResp.VerificationURI == "" {
		t.Fatal("expected verification_uri")
	}

	// Step 2: Poll for token (simulating user approval)
	tokenReq := map[string]string{
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		"device_code": deviceResp.DeviceCode,
		"client_id":   "helixplay-device",
	}
	body, _ = json.Marshal(tokenReq)
	tokenHTTPReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/oauth/token", strings.NewReader(string(body)))
	tokenHTTPReq.Header.Set("Content-Type", "application/json")

	tokenResp, err := client.Do(tokenHTTPReq)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}
	if tokenResult.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if !tokenExchanged {
		t.Fatal("expected token to be exchanged")
	}
}
