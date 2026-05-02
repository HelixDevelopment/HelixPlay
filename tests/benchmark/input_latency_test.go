package benchmark_test

import (
	"sort"
	"testing"
	"time"

	"digital.vasic.memory/pkg/memfd"
)

func BenchmarkControllerInputLatency(b *testing.B) {
	// Create a real lock-free ring buffer for input events
	rb := memfd.NewPSC(4096)
	defer rb.Close()

	// Warm-up: ensure ring buffer is ready
	if _, err := rb.Write([]byte{0x00}); err != nil {
		b.Fatalf("warm-up write failed: %v", err)
	}
	warmBuf := make([]byte, 1)
	if _, err := rb.Read(warmBuf); err != nil {
		b.Fatalf("warm-up read failed: %v", err)
	}

	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Simulate controller input event written to ring buffer
		event := []byte{0x01, byte(i % 256)}
		if _, err := rb.Write(event); err != nil {
			b.Fatalf("write failed: %v", err)
		}

		// Consumer reads the event back (observable end-to-end latency)
		readBuf := make([]byte, len(event))
		if _, err := rb.Read(readBuf); err != nil {
			b.Fatalf("read failed: %v", err)
		}

		latencies = append(latencies, time.Since(start))
	}
	b.StopTimer()

	if len(latencies) == 0 {
		b.Fatal("expected at least one latency sample")
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p50 := latencies[len(latencies)*50/100]
	p99 := latencies[len(latencies)*99/100]
	p999 := latencies[len(latencies)*999/1000]

	b.ReportMetric(float64(p50.Nanoseconds())/1e6, "p50_ms")
	b.ReportMetric(float64(p99.Nanoseconds())/1e6, "p99_ms")
	b.ReportMetric(float64(p999.Nanoseconds())/1e6, "p999_ms")
}
