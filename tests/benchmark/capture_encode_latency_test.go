package benchmark_test

import (
	"sort"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
)

func BenchmarkSoftwareEncodeLatency(b *testing.B) {
	// Initialize real software encoder (always available, performs real work)
	enc := encoder.NewDualPath("software")
	if enc == nil {
		b.Fatal("expected software encoder to be created")
	}
	if err := enc.Start(); err != nil {
		b.Fatal(err)
	}
	defer enc.Stop()

	// Synthetic 1080p RGBA frame
	frame := make([]byte, 1920*1080*4)
	for i := range frame {
		frame[i] = byte(i % 256)
	}

	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Encode step: real software encoding with delta-RLE
		_, err := enc.EncodeFrame(frame)
		if err != nil {
			b.Fatalf("encode failed: %v", err)
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
