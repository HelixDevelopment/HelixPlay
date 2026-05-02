package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.DiscoveryMode != "mdns" {
		t.Errorf("expected default discovery_mode mdns, got %s", cfg.DiscoveryMode)
	}
	if cfg.MaxSessions != 1 {
		t.Errorf("expected default max_sessions 1, got %d", cfg.MaxSessions)
	}
	if cfg.InputPollHz != 1000 {
		t.Errorf("expected default input_poll_hz 1000, got %d", cfg.InputPollHz)
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}

	cfg.DiscoveryMode = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for invalid discovery_mode")
	}
	cfg.DiscoveryMode = "mdns"

	cfg.ServiceName = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for empty service_name")
	}
	cfg.ServiceName = "test"

	cfg.MaxSessions = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for max_sessions < 1")
	}
	cfg.MaxSessions = 1

	cfg.DefaultFPS = 300
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for default_fps > 240")
	}
	cfg.DefaultFPS = 60

	cfg.InputPollHz = 10
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for input_poll_hz < 60")
	}
}

func TestLoadFromTOML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "host-agent.toml")
	content := `discovery_mode = "rendezvous"
service_name = "TestHost"
max_sessions = 4
default_fps = 120
input_poll_hz = 2000
enable_dualsense = false
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DiscoveryMode != "rendezvous" {
		t.Errorf("expected discovery_mode rendezvous, got %s", cfg.DiscoveryMode)
	}
	if cfg.ServiceName != "TestHost" {
		t.Errorf("expected service_name TestHost, got %s", cfg.ServiceName)
	}
	if cfg.MaxSessions != 4 {
		t.Errorf("expected max_sessions 4, got %d", cfg.MaxSessions)
	}
	if cfg.DefaultFPS != 120 {
		t.Errorf("expected default_fps 120, got %d", cfg.DefaultFPS)
	}
	if cfg.InputPollHz != 2000 {
		t.Errorf("expected input_poll_hz 2000, got %d", cfg.InputPollHz)
	}
	if cfg.EnableDualSense != false {
		t.Error("expected enable_dualsense false")
	}
}

func TestLoadEnvOverlay(t *testing.T) {
	t.Setenv("HELIXPLAY_DISCOVERY_MODE", "both")
	t.Setenv("HELIXPLAY_MAX_SESSIONS", "8")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DiscoveryMode != "both" {
		t.Errorf("expected discovery_mode both from env, got %s", cfg.DiscoveryMode)
	}
	if cfg.MaxSessions != 8 {
		t.Errorf("expected max_sessions 8 from env, got %d", cfg.MaxSessions)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Load should fallback to defaults when file missing: %v", err)
	}
	if cfg.DiscoveryMode != "mdns" {
		t.Errorf("expected default discovery_mode, got %s", cfg.DiscoveryMode)
	}
}
