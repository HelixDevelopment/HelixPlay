package game

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Monitor watches launched game processes for health and crashes.
type Monitor struct {
	mu        sync.RWMutex
	processes map[string]*MonitoredProcess
	checkInterval time.Duration
	cancel    context.CancelFunc
	done      chan struct{}
}

// MonitoredProcess extends LaunchedProcess with monitoring metadata.
type MonitoredProcess struct {
	*LaunchedProcess
	LastHealthyAt time.Time
	CrashHandler  func(gp Game, exitErr error)
}

// NewMonitor creates a process monitor with the given check interval.
func NewMonitor(checkInterval time.Duration) *Monitor {
	if checkInterval <= 0 {
		checkInterval = 5 * time.Second
	}
	return &Monitor{
		processes:     make(map[string]*MonitoredProcess),
		checkInterval: checkInterval,
		done:          make(chan struct{}),
	}
}

// Start begins the monitoring loop.
func (m *Monitor) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	go m.loop(ctx)
}

// Stop halts the monitoring loop.
func (m *Monitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	<-m.done
}

// Add registers a process for monitoring.
func (m *Monitor) Add(mp *MonitoredProcess) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mp.LastHealthyAt = time.Now().UTC()
	m.processes[mp.Game.Title] = mp
}

// Remove unregisters a process from monitoring.
func (m *Monitor) Remove(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.processes, title)
}

// IsRunning reports whether a monitored game is still alive.
func (m *Monitor) IsRunning(title string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mp, ok := m.processes[title]
	if !ok {
		return false
	}
	return mp.LaunchedProcess.IsRunning()
}

// Count returns the number of actively monitored processes.
func (m *Monitor) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processes)
}

func (m *Monitor) loop(ctx context.Context) {
	defer close(m.done)
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.check()
		}
	}
}

func (m *Monitor) check() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for title, mp := range m.processes {
		if !mp.LaunchedProcess.IsRunning() {
			exitErr := fmt.Errorf("process exited")
			if mp.CrashHandler != nil {
				go mp.CrashHandler(mp.Game, exitErr)
			}
			delete(m.processes, title)
			continue
		}
		mp.LastHealthyAt = time.Now().UTC()
	}
}
