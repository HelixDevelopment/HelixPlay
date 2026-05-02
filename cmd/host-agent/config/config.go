// Package config loads and validates host-agent configuration from TOML files
// and environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all host-agent runtime configuration.
type Config struct {
	// Discovery settings
	DiscoveryMode string `toml:"discovery_mode" env:"HELIXPLAY_DISCOVERY_MODE"`
	ServiceName   string `toml:"service_name" env:"HELIXPLAY_SERVICE_NAME"`
	ServiceType   string `toml:"service_type" env:"HELIXPLAY_SERVICE_TYPE"`

	// Network settings
	GRPCAddress    string `toml:"grpc_address" env:"HELIXPLAY_GRPC_ADDRESS"`
	HTTPAddress    string `toml:"http_address" env:"HELIXPLAY_HTTP_ADDRESS"`
	DiscoveryPort  int    `toml:"discovery_port" env:"HELIXPLAY_DISCOVERY_PORT"`
	QUICPort       int    `toml:"quic_port" env:"HELIXPLAY_QUIC_PORT"`

	// Coordinator settings
	CoordinatorURL   string `toml:"coordinator_url" env:"HELIXPLAY_COORDINATOR_URL"`
	BootstrapToken   string `toml:"bootstrap_token" env:"HELIXPLAY_BOOTSTRAP_TOKEN"`
	MTLSCertPath     string `toml:"mtls_cert_path" env:"HELIXPLAY_MTLS_CERT_PATH"`
	MTLSKeyPath      string `toml:"mtls_key_path" env:"HELIXPLAY_MTLS_KEY_PATH"`
	MTLSCAPath       string `toml:"mtls_ca_path" env:"HELIXPLAY_MTLS_CA_PATH"`

	// Game store settings
	SteamRoot      string `toml:"steam_root" env:"HELIXPLAY_STEAM_ROOT"`
	EpicRoot       string `toml:"epic_root" env:"HELIXPLAY_EPIC_ROOT"`
	GOGRoot        string `toml:"gog_root" env:"HELIXPLAY_GOG_ROOT"`
	StandalonePaths []string `toml:"standalone_paths" env:"HELIXPLAY_STANDALONE_PATHS"`

	// Hardware settings
	GPUVendor      string `toml:"gpu_vendor" env:"HELIXPLAY_GPU_VENDOR"`
	CaptureBackend string `toml:"capture_backend" env:"HELIXPLAY_CAPTURE_BACKEND"`
	EncoderBackend string `toml:"encoder_backend" env:"HELIXPLAY_ENCODER_BACKEND"`

	// Session settings
	MaxSessions       int    `toml:"max_sessions" env:"HELIXPLAY_MAX_SESSIONS"`
	DefaultCodec      string `toml:"default_codec" env:"HELIXPLAY_DEFAULT_CODEC"`
	DefaultResolution string `toml:"default_resolution" env:"HELIXPLAY_DEFAULT_RESOLUTION"`
	DefaultFPS        int    `toml:"default_fps" env:"HELIXPLAY_DEFAULT_FPS"`

	// Input settings
	InputPollHz     int  `toml:"input_poll_hz" env:"HELIXPLAY_INPUT_POLL_HZ"`
	EnableDualSense bool `toml:"enable_dualsense" env:"HELIXPLAY_ENABLE_DUALSENSE"`
	EnableHotplug   bool `toml:"enable_hotplug" env:"HELIXPLAY_ENABLE_HOTPLUG"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		DiscoveryMode:     "mdns",
		ServiceName:       "HelixPlay-Host",
		ServiceType:       "_helixplay._tcp",
		GRPCAddress:       ":50051",
		HTTPAddress:       ":8080",
		DiscoveryPort:     0,
		QUICPort:          0,
		CoordinatorURL:    "http://localhost:8080",
		SteamRoot:         "",
		EpicRoot:          "",
		GOGRoot:           "",
		StandalonePaths:   []string{},
		GPUVendor:         "auto",
		CaptureBackend:    "auto",
		EncoderBackend:    "auto",
		MaxSessions:       1,
		DefaultCodec:      "H.264",
		DefaultResolution: "1080p",
		DefaultFPS:        60,
		InputPollHz:       1000,
		EnableDualSense:   true,
		EnableHotplug:     true,
	}
}

// Load reads configuration from a TOML file and overlays environment variables.
// If path is empty or the file does not exist, only defaults + env vars are used.
func Load(path string) (*Config, error) {
	cfg := Default()

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if err := loadTOML(path, cfg); err != nil {
				return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
			}
		}
	}

	overlayEnv(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required fields are present and valid.
func (c *Config) Validate() error {
	if c.DiscoveryMode != "mdns" && c.DiscoveryMode != "rendezvous" && c.DiscoveryMode != "both" && c.DiscoveryMode != "none" {
		return fmt.Errorf("invalid discovery_mode: %q (must be mdns, rendezvous, both, or none)", c.DiscoveryMode)
	}
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if c.MaxSessions < 1 {
		return fmt.Errorf("max_sessions must be >= 1")
	}
	if c.DefaultFPS < 1 || c.DefaultFPS > 240 {
		return fmt.Errorf("default_fps must be between 1 and 240")
	}
	if c.InputPollHz < 60 || c.InputPollHz > 8000 {
		return fmt.Errorf("input_poll_hz must be between 60 and 8000")
	}
	return nil
}

// loadTOML parses a minimal TOML-like key=value config file.
// This is a lightweight parser sufficient for flat host-agent configs.
func loadTOML(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)

		switch key {
		case "discovery_mode":
			cfg.DiscoveryMode = val
		case "service_name":
			cfg.ServiceName = val
		case "service_type":
			cfg.ServiceType = val
		case "grpc_address":
			cfg.GRPCAddress = val
		case "http_address":
			cfg.HTTPAddress = val
		case "discovery_port":
			cfg.DiscoveryPort, _ = strconv.Atoi(val)
		case "quic_port":
			cfg.QUICPort, _ = strconv.Atoi(val)
		case "coordinator_url":
			cfg.CoordinatorURL = val
		case "bootstrap_token":
			cfg.BootstrapToken = val
		case "mtls_cert_path":
			cfg.MTLSCertPath = val
		case "mtls_key_path":
			cfg.MTLSKeyPath = val
		case "mtls_ca_path":
			cfg.MTLSCAPath = val
		case "steam_root":
			cfg.SteamRoot = val
		case "epic_root":
			cfg.EpicRoot = val
		case "gog_root":
			cfg.GOGRoot = val
		case "gpu_vendor":
			cfg.GPUVendor = val
		case "capture_backend":
			cfg.CaptureBackend = val
		case "encoder_backend":
			cfg.EncoderBackend = val
		case "max_sessions":
			cfg.MaxSessions, _ = strconv.Atoi(val)
		case "default_codec":
			cfg.DefaultCodec = val
		case "default_resolution":
			cfg.DefaultResolution = val
		case "default_fps":
			cfg.DefaultFPS, _ = strconv.Atoi(val)
		case "input_poll_hz":
			cfg.InputPollHz, _ = strconv.Atoi(val)
		case "enable_dualsense":
			cfg.EnableDualSense, _ = strconv.ParseBool(val)
		case "enable_hotplug":
			cfg.EnableHotplug, _ = strconv.ParseBool(val)
		}
	}
	return nil
}

// overlayEnv applies environment variables on top of the current config.
func overlayEnv(cfg *Config) {
	if v := os.Getenv("HELIXPLAY_DISCOVERY_MODE"); v != "" {
		cfg.DiscoveryMode = v
	}
	if v := os.Getenv("HELIXPLAY_SERVICE_NAME"); v != "" {
		cfg.ServiceName = v
	}
	if v := os.Getenv("HELIXPLAY_SERVICE_TYPE"); v != "" {
		cfg.ServiceType = v
	}
	if v := os.Getenv("HELIXPLAY_GRPC_ADDRESS"); v != "" {
		cfg.GRPCAddress = v
	}
	if v := os.Getenv("HELIXPLAY_HTTP_ADDRESS"); v != "" {
		cfg.HTTPAddress = v
	}
	if v := os.Getenv("HELIXPLAY_DISCOVERY_PORT"); v != "" {
		cfg.DiscoveryPort, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("HELIXPLAY_QUIC_PORT"); v != "" {
		cfg.QUICPort, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("HELIXPLAY_COORDINATOR_URL"); v != "" {
		cfg.CoordinatorURL = v
	}
	if v := os.Getenv("HELIXPLAY_BOOTSTRAP_TOKEN"); v != "" {
		cfg.BootstrapToken = v
	}
	if v := os.Getenv("HELIXPLAY_MTLS_CERT_PATH"); v != "" {
		cfg.MTLSCertPath = v
	}
	if v := os.Getenv("HELIXPLAY_MTLS_KEY_PATH"); v != "" {
		cfg.MTLSKeyPath = v
	}
	if v := os.Getenv("HELIXPLAY_MTLS_CA_PATH"); v != "" {
		cfg.MTLSCAPath = v
	}
	if v := os.Getenv("HELIXPLAY_STEAM_ROOT"); v != "" {
		cfg.SteamRoot = v
	}
	if v := os.Getenv("HELIXPLAY_GPU_VENDOR"); v != "" {
		cfg.GPUVendor = v
	}
	if v := os.Getenv("HELIXPLAY_CAPTURE_BACKEND"); v != "" {
		cfg.CaptureBackend = v
	}
	if v := os.Getenv("HELIXPLAY_ENCODER_BACKEND"); v != "" {
		cfg.EncoderBackend = v
	}
	if v := os.Getenv("HELIXPLAY_MAX_SESSIONS"); v != "" {
		cfg.MaxSessions, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("HELIXPLAY_DEFAULT_CODEC"); v != "" {
		cfg.DefaultCodec = v
	}
	if v := os.Getenv("HELIXPLAY_DEFAULT_RESOLUTION"); v != "" {
		cfg.DefaultResolution = v
	}
	if v := os.Getenv("HELIXPLAY_DEFAULT_FPS"); v != "" {
		cfg.DefaultFPS, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("HELIXPLAY_INPUT_POLL_HZ"); v != "" {
		cfg.InputPollHz, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("HELIXPLAY_ENABLE_DUALSENSE"); v != "" {
		cfg.EnableDualSense, _ = strconv.ParseBool(v)
	}
	if v := os.Getenv("HELIXPLAY_ENABLE_HOTPLUG"); v != "" {
		cfg.EnableHotplug, _ = strconv.ParseBool(v)
	}
}
