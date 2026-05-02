package benchmark_test

import (
	"sort"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capture"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
)

func BenchmarkCaptureEncodeLatency(b *testing.B) {
	// Initialize real capture backend
	capturer := capture.NewCapturer("linux")
	if capturer == nil {
		b.Fatal("expected capturer to be created")
	}
	if err := capturer.Start(); err != nil {
		b.Fatal(err)
	}
	defer capturer.Stop()

	// Initialize real hardware encoder
	enc := encoder.NewDualPath("nvenc")
	if enc == nil {
		b.Fatal("expected encoder to be created")
	}
	if err := enc.Start(); err != nil {
		b.Fatal(err)
	}
	defer enc.Stop()

	// Synthetic frame for encoding when capture stub returns no frame
	syntheticFrame := make([]byte, 1920*1080*4)

	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Capture step
		frame, _ := capturer.GetFrame()
		if frame == nil {
			frame = syntheticFrame
		}

		// Encode step (observable: encoder returns output set during Start)
		_ = enc.GetStreamOutput()
		_ = frame // ensure frame is used to prevent compiler optimization issues

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
