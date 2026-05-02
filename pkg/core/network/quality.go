// Package network estimates network quality metrics for adaptive bitrate streaming.
package network

import (
	"math"
	"sync"
	"time"
)

// Sample holds a single RTT measurement.
type Sample struct {
	RTTMs      float64
	BitrateKbps float64
	LossPct    float64
	Timestamp  time.Time
}

// Estimator tracks bandwidth, RTT, jitter, and packet loss over a sliding window.
type Estimator struct {
	mu       sync.RWMutex
	window   []Sample
	capacity int
}

// NewEstimator creates an estimator with the given sample window capacity.
func NewEstimator(capacity int) *Estimator {
	if capacity <= 0 {
		capacity = 60
	}
	return &Estimator{capacity: capacity}
}

// Record adds a new measurement sample.
func (e *Estimator) Record(rttMs, bitrateKbps, lossPct float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.window = append(e.window, Sample{
		RTTMs:       rttMs,
		BitrateKbps: bitrateKbps,
		LossPct:     lossPct,
		Timestamp:   time.Now().UTC(),
	})

	if len(e.window) > e.capacity {
		e.window = e.window[len(e.window)-e.capacity:]
	}
}

// RTT returns the median RTT in milliseconds.
func (e *Estimator) RTT() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.window) == 0 {
		return 0
	}
	return percentile(e.rtts(), 0.5)
}

// Jitter returns the standard deviation of RTT samples.
func (e *Estimator) Jitter() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.window) < 2 {
		return 0
	}
	return stdDev(e.rtts())
}

// Bandwidth returns the estimated available bandwidth in kbps (p90 of observed bitrates).
func (e *Estimator) Bandwidth() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.window) == 0 {
		return 0
	}
	return percentile(e.bitrates(), 0.9)
}

// Loss returns the average packet loss percentage.
func (e *Estimator) Loss() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.window) == 0 {
		return 0
	}
	var sum float64
	for _, s := range e.window {
		sum += s.LossPct
	}
	return sum / float64(len(e.window))
}

// Stable reports whether the network appears stable (low jitter and loss).
func (e *Estimator) Stable() bool {
	e.mu.RLock()
	n := len(e.window)
	e.mu.RUnlock()
	if n < 2 {
		return false
	}
	return e.Jitter() < 10 && e.Loss() < 1.0
}

// Recommendation returns a simple quality grade: "good", "fair", or "poor".
func (e *Estimator) Recommendation() string {
	rtt := e.RTT()
	loss := e.Loss()
	if rtt < 30 && loss < 0.5 {
		return "good"
	}
	if rtt < 80 && loss < 2.0 {
		return "fair"
	}
	return "poor"
}

func (e *Estimator) rtts() []float64 {
	out := make([]float64, len(e.window))
	for i, s := range e.window {
		out[i] = s.RTTMs
	}
	return out
}

func (e *Estimator) bitrates() []float64 {
	out := make([]float64, len(e.window))
	for i, s := range e.window {
		out[i] = s.BitrateKbps
	}
	return out
}

func percentile(vals []float64, p float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	// Simple sort-based percentile using nearest-rank method
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	// Use ceiling for nearest-rank: rank = ceil(p * N)
	rank := int(math.Ceil(p * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

func stdDev(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(len(vals))
	var sqDiff float64
	for _, v := range vals {
		d := v - mean
		sqDiff += d * d
	}
	return math.Sqrt(sqDiff / float64(len(vals)))
}
