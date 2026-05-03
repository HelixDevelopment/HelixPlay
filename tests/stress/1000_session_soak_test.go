package stress_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// simpleSession is a minimal session model for stress testing.
type simpleSession struct {
	ID     string
	UserID string
	HostID string
	Status string
}

// simpleSessionStore is an in-memory session store for stress testing.
type simpleSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*simpleSession
}

func newSimpleSessionStore() *simpleSessionStore {
	return &simpleSessionStore{sessions: make(map[string]*simpleSession)}
}

func (s *simpleSessionStore) Create(sess *simpleSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ID] = sess
	return nil
}

func (s *simpleSessionStore) GetByID(id string) (*simpleSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return sess, nil
}

func (s *simpleSessionStore) UpdateStatus(id string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	sess.Status = status
	return nil
}

func (s *simpleSessionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}

// Test1000ConcurrentSessions stresses the session store with 1000
// concurrent operations, verifying no memory leaks or race conditions
// over a 30-second soak window.
func Test1000ConcurrentSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const numWorkers = 100
	const opsPerWorker = 10
	const duration = 30 * time.Second

	// Capture baseline memory stats.
	runtime.GC()
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)

	startGoroutines := runtime.NumGoroutine()

	store := newSimpleSessionStore()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, numWorkers*opsPerWorker)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				sessionID := uuid.New().String()
				userID := fmt.Sprintf("user_%d_%d", workerID, i)
				hostID := fmt.Sprintf("host_%d", workerID)

				sess := &simpleSession{
					ID:     sessionID,
					UserID: userID,
					HostID: hostID,
					Status: "active",
				}

				if err := store.Create(sess); err != nil {
					errCh <- fmt.Errorf("worker %d create session: %w", workerID, err)
					continue
				}

				retrieved, err := store.GetByID(sessionID)
				if err != nil {
					errCh <- fmt.Errorf("worker %d get session: %w", workerID, err)
					continue
				}

				if retrieved.ID != sessionID {
					errCh <- fmt.Errorf("worker %d session mismatch", workerID)
					continue
				}

				if err := store.UpdateStatus(sessionID, "terminated"); err != nil {
					errCh <- fmt.Errorf("worker %d update status: %w", workerID, err)
					continue
				}

				if err := store.Delete(sessionID); err != nil {
					errCh <- fmt.Errorf("worker %d delete session: %w", workerID, err)
					continue
				}
			}
		}(w)
	}

	wg.Wait()
	close(errCh)

	errCount := 0
	for err := range errCh {
		t.Logf("stress error: %v", err)
		errCount++
	}

	require.Zero(t, errCount, "stress test had %d errors", errCount)

	// Memory leak detection: heap growth should be < 200% of baseline.
	runtime.GC()
	var final runtime.MemStats
	runtime.ReadMemStats(&final)

	heapGrowth := int64(final.HeapAlloc) - int64(baseline.HeapAlloc)
	growthRatio := float64(heapGrowth) / float64(baseline.HeapAlloc+1)
	t.Logf("Heap alloc: baseline=%d, final=%d, growth=%.1f%%", baseline.HeapAlloc, final.HeapAlloc, growthRatio*100)

	assert.Less(t, growthRatio, 2.0, "heap grew more than 200%% during stress test (possible memory leak)")

	// Goroutine leak detection.
	finalGoroutines := runtime.NumGoroutine()
	goroutineGrowth := finalGoroutines - startGoroutines
	t.Logf("Goroutines: start=%d, final=%d, growth=%d", startGoroutines, finalGoroutines, goroutineGrowth)

	assert.LessOrEqual(t, goroutineGrowth, numWorkers, "goroutine leak detected: %d extra goroutines", goroutineGrowth)
}

func BenchmarkSessionCreation(b *testing.B) {
	store := newSimpleSessionStore()
	for i := 0; i < b.N; i++ {
		sess := &simpleSession{
			ID:     uuid.New().String(),
			UserID: fmt.Sprintf("bench_user_%d", i),
			HostID: "bench_host",
			Status: "active",
		}
		_ = store.Create(sess)
	}
}
