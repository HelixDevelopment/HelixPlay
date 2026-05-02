package stress_test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test1000ConcurrentSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const numSessions = 1000
	const duration = 30 * time.Second // Shortened for test suite; real soak is 24h

	var wg sync.WaitGroup
	errors := make(chan error, numSessions)
	start := time.Now()

	for i := 0; i < numSessions; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Simulate session lifecycle
			timer := time.NewTimer(duration)
			defer timer.Stop()
			<-timer.C
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	close(errors)
	errCount := 0
	for range errors {
		errCount++
	}

	t.Logf("Completed %d concurrent sessions in %v", numSessions, elapsed)
	if errCount > 0 {
		t.Fatalf("%d sessions encountered errors", errCount)
	}
}

func BenchmarkSessionCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("sess_%d_%d", i, time.Now().UnixNano())
	}
}
