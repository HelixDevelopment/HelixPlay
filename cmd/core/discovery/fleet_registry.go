package discovery

import (
	"fmt"
	"sync"
	"time"
)

// FleetHost holds registered host information with capabilities.
type FleetHost struct {
	ID              string
	Name            string
	Address         string
	Port            int
	Codecs          []string
	MaxResolution   string
	MaxFPS          int
	GPUs            []string
	ActiveSessions  int
	MaxSessions     int
	LastHeartbeat   time.Time
	Status          string
	Region          string
}

// Registry maintains the fleet of available hosts.
type Registry struct {
	mu    sync.RWMutex
	hosts map[string]*FleetHost
}

// NewRegistry creates an empty fleet registry.
func NewRegistry() *Registry {
	return &Registry{hosts: make(map[string]*FleetHost)}
}

// Register adds or updates a host in the fleet.
func (r *Registry) Register(h FleetHost) error {
	if h.ID == "" {
		return fmt.Errorf("host ID is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hosts[h.ID] = &h
	return nil
}

// Deregister removes a host from the fleet.
func (r *Registry) Deregister(hostID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hosts, hostID)
}

// Get returns a host by ID.
func (r *Registry) Get(hostID string) (*FleetHost, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hosts[hostID]
	return h, ok
}

// List returns all registered hosts.
func (r *Registry) List() []*FleetHost {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*FleetHost, 0, len(r.hosts))
	for _, h := range r.hosts {
		out = append(out, h)
	}
	return out
}

// ListAvailable returns hosts that have capacity for new sessions.
func (r *Registry) ListAvailable() []*FleetHost {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*FleetHost
	for _, h := range r.hosts {
		if h.Status == "online" && h.ActiveSessions < h.MaxSessions {
			out = append(out, h)
		}
	}
	return out
}

// ListByRegion returns hosts in the given region.
func (r *Registry) ListByRegion(region string) []*FleetHost {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*FleetHost
	for _, h := range r.hosts {
		if h.Region == region {
			out = append(out, h)
		}
	}
	return out
}

// Count returns the total number of registered hosts.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.hosts)
}

// UpdateHeartbeat updates the last heartbeat time for a host.
func (r *Registry) UpdateHeartbeat(hostID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.hosts[hostID]
	if !ok {
		return false
	}
	h.LastHeartbeat = time.Now().UTC()
	return true
}

// IncrementSessions atomically increments the active session count.
func (r *Registry) IncrementSessions(hostID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.hosts[hostID]
	if !ok || h.ActiveSessions >= h.MaxSessions {
		return false
	}
	h.ActiveSessions++
	return true
}

// DecrementSessions atomically decrements the active session count.
func (r *Registry) DecrementSessions(hostID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.hosts[hostID]
	if !ok || h.ActiveSessions <= 0 {
		return false
	}
	h.ActiveSessions--
	return true
}
