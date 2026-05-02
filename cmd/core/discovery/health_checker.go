package discovery

import (
	"context"
	"sync"
	"time"
)

// HealthChecker monitors host health based on heartbeat timeouts.
type HealthChecker struct {
	mu          sync.RWMutex
	registry    *Registry
	timeout     time.Duration
	unhealthy   map[string]time.Time
	cancel      context.CancelFunc
	done        chan struct{}
}

// NewHealthChecker creates a health checker for the given registry.
func NewHealthChecker(registry *Registry, timeout time.Duration) *HealthChecker {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HealthChecker{
		registry:  registry,
		timeout:   timeout,
		unhealthy: make(map[string]time.Time),
	}
}

// Start begins the background health check loop.
func (hc *HealthChecker) Start(ctx context.Context) {
	cctx, cancel := context.WithCancel(ctx)
	hc.cancel = cancel
	hc.done = make(chan struct{})
	go hc.loop(cctx)
}

// Stop halts the health checker.
func (hc *HealthChecker) Stop() {
	if hc.cancel != nil {
		hc.cancel()
		<-hc.done
	}
}

// IsHealthy reports whether a host is considered healthy.
func (hc *HealthChecker) IsHealthy(hostID string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	_, unhealthy := hc.unhealthy[hostID]
	return !unhealthy
}

// UnhealthyHosts returns the list of currently unhealthy host IDs.
func (hc *HealthChecker) UnhealthyHosts() []string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	out := make([]string, 0, len(hc.unhealthy))
	for id := range hc.unhealthy {
		out = append(out, id)
	}
	return out
}

func (hc *HealthChecker) loop(ctx context.Context) {
	defer close(hc.done)
	ticker := time.NewTicker(hc.timeout / 2)
	defer ticker.Stop()

	// Run initial check immediately
	hc.check()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.check()
		}
	}
}

func (hc *HealthChecker) check() {
	now := time.Now().UTC()
	hosts := hc.registry.List()

	hc.mu.Lock()
	defer hc.mu.Unlock()

	for _, h := range hosts {
		if h.LastHeartbeat.IsZero() {
			continue
		}
		if now.Sub(h.LastHeartbeat) > hc.timeout {
			hc.unhealthy[h.ID] = now
		} else {
			delete(hc.unhealthy, h.ID)
		}
	}
}
