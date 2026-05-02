package lifecycle

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Registration holds the result of a successful host registration.
type Registration struct {
	HostID    string `json:"host_id"`
	AuthToken string `json:"auth_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Register sends a host registration request to the coordinator.
// It uses mTLS when cert/key paths are provided, otherwise falls back to
// bootstrap token authentication over HTTPS.
func Register(ctx context.Context, coordinatorURL, hardwareID, bootstrapToken string, certPath, keyPath, caPath string) (*Registration, error) {
	client, err := newHTTPClient(certPath, keyPath, caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	payload := map[string]string{
		"hardware_id":     hardwareID,
		"bootstrap_token": bootstrapToken,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration payload: %w", err)
	}

	url := coordinatorURL + "/api/v1/hosts/register"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var reg Registration
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	return &reg, nil
}

func newHTTPClient(certPath, keyPath, caPath string) (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load mTLS key pair: %w", err)
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}
