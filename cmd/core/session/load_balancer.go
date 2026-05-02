package session

import (
	"fmt"
	"math"
	"sync"
)

// LoadBalancer selects hosts for new sessions based on latency and GPU affinity.
type LoadBalancer struct {
	mu        sync.RWMutex
	latencies map[string]float64 // host ID -> RTT ms
}

// NewLoadBalancer creates a load balancer.
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{latencies: make(map[string]float64)}
}

// RecordLatency stores the observed RTT for a host.
func (lb *LoadBalancer) RecordLatency(hostID string, rttMs float64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.latencies[hostID] = rttMs
}

// Select returns the best host from the available pool.
// It prefers hosts with lowest latency that support the requested codec.
func (lb *LoadBalancer) Select(available []*HostCandidate, preferredCodec string) (*HostCandidate, error) {
	if len(available) == 0 {
		return nil, fmt.Errorf("no available hosts")
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var best *HostCandidate
	bestScore := math.MaxFloat64

	for _, h := range available {
		// Filter by codec support
		if preferredCodec != "" && !hasCodec(h.Codecs, preferredCodec) {
			continue
		}

		// Score = latency + penalty for active sessions
		rtt := lb.latencies[h.ID]
		if rtt == 0 {
			rtt = 50 // Default assumption if no measurement
		}
		sessionPenalty := float64(h.ActiveSessions) * 10
		score := rtt + sessionPenalty

		if score < bestScore {
			bestScore = score
			best = h
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no host supports codec %s", preferredCodec)
	}
	return best, nil
}

// HostCandidate holds host info needed for load balancing.
type HostCandidate struct {
	ID              string
	Address         string
	Port            int
	Codecs          []string
	ActiveSessions  int
	MaxSessions     int
	GPUModel        string
}

func hasCodec(codecs []string, target string) bool {
	for _, c := range codecs {
		if c == target {
			return true
		}
	}
	return false
}
