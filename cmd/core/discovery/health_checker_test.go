package discovery

import (
	"context"
	"testing"
	"time"
)

func TestHealthCheckerDetectsStaleHeartbeat(t *testing.T) {
	reg := NewRegistry()
	reg.Register(FleetHost{ID: "h1", LastHeartbeat: time.Now().UTC().Add(-2 * time.Minute)})
	reg.Register(FleetHost{ID: "h2", LastHeartbeat: time.Now().UTC()})

	hc := NewHealthChecker(reg, 30*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	hc.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	if hc.IsHealthy("h1") {
		t.Error("expected h1 to be unhealthy")
	}
	if !hc.IsHealthy("h2") {
		t.Error("expected h2 to be healthy")
	}

	unhealthy := hc.UnhealthyHosts()
	if len(unhealthy) != 1 || unhealthy[0] != "h1" {
		t.Fatalf("expected [h1] unhealthy, got %v", unhealthy)
	}

	hc.Stop()
	cancel()
}

func TestHealthCheckerStartStop(t *testing.T) {
	reg := NewRegistry()
	hc := NewHealthChecker(reg, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hc.Start(ctx)
	hc.Stop()
}

func TestHealthCheckerRecovers(t *testing.T) {
	reg := NewRegistry()
	reg.Register(FleetHost{ID: "h1", LastHeartbeat: time.Now().UTC().Add(-2 * time.Minute)})

	hc := NewHealthChecker(reg, 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	hc.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	if hc.IsHealthy("h1") {
		t.Error("expected h1 to be unhealthy initially")
	}

	// Update heartbeat
	reg.UpdateHeartbeat("h1")
	time.Sleep(60 * time.Millisecond)

	if !hc.IsHealthy("h1") {
		t.Error("expected h1 to be healthy after heartbeat update")
	}

	hc.Stop()
	cancel()
}
