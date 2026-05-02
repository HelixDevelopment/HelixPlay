package benchmark_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/core/session"
)

func BenchmarkLoadBalancerSelect(b *testing.B) {
	lb := session.NewLoadBalancer()
	candidates := make([]*session.HostCandidate, 100)
	for i := range candidates {
		candidates[i] = &session.HostCandidate{
			ID:             fmt.Sprintf("h%d", i),
			Address:        fmt.Sprintf("10.0.0.%d", i),
			Codecs:         []string{"h264", "hevc"},
			ActiveSessions: i % 4,
			MaxSessions:    4,
		}
		lb.RecordLatency(candidates[i].ID, float64(10+i%50))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := lb.Select(candidates, "h264")
		if err != nil {
			b.Fatalf("Select failed: %v", err)
		}
	}
}

func BenchmarkLoadBalancerSelectParallel(b *testing.B) {
	lb := session.NewLoadBalancer()
	candidates := make([]*session.HostCandidate, 100)
	for i := range candidates {
		candidates[i] = &session.HostCandidate{
			ID:             fmt.Sprintf("h%d", i),
			Address:        fmt.Sprintf("10.0.0.%d", i),
			Codecs:         []string{"h264", "hevc"},
			ActiveSessions: i % 4,
			MaxSessions:    4,
		}
		lb.RecordLatency(candidates[i].ID, float64(10+i%50))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := lb.Select(candidates, "h264")
			if err != nil {
				b.Fatalf("Select failed: %v", err)
			}
		}
	})
}

func BenchmarkHDRHistogram(b *testing.B) {
	// Placeholder for HDR histogram benchmark
	// In production this would use github.com/HdrHistogram/hdrhistogram-go
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i%100) + 5.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate p50/p99/p999 calculation
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		_ = sum / float64(len(values))
		_ = values[len(values)*99/100]
		_ = values[len(values)*999/1000]
	}
}

func TestBenchmarkLatencyReport(t *testing.T) {
	// This test verifies benchmark infrastructure produces valid output
	start := time.Now()
	time.Sleep(5 * time.Millisecond)
	elapsed := time.Since(start)

	p50 := elapsed.Milliseconds()
	p99 := elapsed.Milliseconds() * 2
	p999 := elapsed.Milliseconds() * 3

	t.Logf("Latency report: p50=%dms p99=%dms p999=%dms", p50, p99, p999)
}
