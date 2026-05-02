package game

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Terminator handles graceful game shutdown with fallback to force kill.
type Terminator struct {
	gracefulTimeout time.Duration
}

// NewTerminator creates a terminator with the given graceful shutdown timeout.
func NewTerminator(gracefulTimeout time.Duration) *Terminator {
	if gracefulTimeout <= 0 {
		gracefulTimeout = 10 * time.Second
	}
	return &Terminator{gracefulTimeout: gracefulTimeout}
}

// Terminate attempts graceful shutdown then force-kills if needed.
func (t *Terminator) Terminate(proc *LaunchedProcess) error {
	if proc == nil || proc.Cmd == nil || proc.Cmd.Process == nil {
		return nil
	}

	// Step 1: Graceful signal
	if err := proc.Terminate(); err != nil {
		// If graceful signal fails, try force kill immediately
		return proc.Kill()
	}

	// Step 2: Wait for graceful exit with timeout
	ctx, cancel := context.WithTimeout(context.Background(), t.gracefulTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- proc.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		// Timeout: force kill
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("force kill failed after graceful timeout: %w", err)
		}
		return nil
	}
}

// TerminateByTitle finds a monitored process by title and terminates it.
func (t *Terminator) TerminateByTitle(monitor *Monitor, title string) error {
	monitor.mu.RLock()
	mp, ok := monitor.processes[title]
	monitor.mu.RUnlock()
	if !ok {
		return fmt.Errorf("game %s not found in monitor", title)
	}
	return t.Terminate(mp.LaunchedProcess)
}

// SendSignal sends a specific OS signal to the process.
func SendSignal(proc *LaunchedProcess, sig os.Signal) error {
	if proc == nil || proc.Cmd == nil || proc.Cmd.Process == nil {
		return fmt.Errorf("process not running")
	}
	return proc.Cmd.Process.Signal(sig)
}
