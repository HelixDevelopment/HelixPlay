package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/discovery"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/game"
)

func main() {
	var (
		configPath    = flag.String("config", "/etc/helixplay/host-agent.toml", "Path to config file")
		discoveryMode = flag.String("discovery", "mdns", "Discovery mode: mdns, rendezvous, or both")
	)
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("HelixPlay Host Agent starting", "version", "1.0.0", "config", *configPath)

	// Start discovery beacon
	beacon, err := discovery.NewBeacon("HelixPlay-Host", "_helixplay._tcp", 0)
	if err != nil {
		slog.Error("Failed to create discovery beacon", "error", err)
		os.Exit(1)
	}
	if err := beacon.Start(); err != nil {
		slog.Error("Failed to start discovery beacon", "error", err)
		os.Exit(1)
	}
	defer beacon.Stop()

	// Advertise capabilities
	caps, err := capability.Advertise()
	if err != nil {
		slog.Error("Failed to advertise capabilities", "error", err)
		os.Exit(1)
	}
	slog.Info("Capabilities advertised", "gpu", caps.GPUModel, "codecs", caps.CodecsSupported)

	// Enumerate installed games
	enumerator := game.NewEnumerator("all")
	games, err := enumerator.Enumerate()
	if err != nil {
		slog.Error("Failed to enumerate games", "error", err)
	} else {
		slog.Info("Games enumerated", "count", len(games))
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Host Agent running", "discovery", *discoveryMode)
	<-sigCh

	slog.Info("Host Agent shutting down gracefully")
}
