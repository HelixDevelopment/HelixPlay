package game

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Launcher handles spawning game processes and attaching capture.
type Launcher struct {
	// CaptureAttachFn is called after the game process starts to bind
	// the screen capture pipeline. If nil, no capture is attached.
	CaptureAttachFn func(pid int) error
}

// NewLauncher creates a game launcher with optional capture attachment.
func NewLauncher(attachFn func(pid int) error) *Launcher {
	return &Launcher{CaptureAttachFn: attachFn}
}

// Launch starts a game process and optionally attaches capture.
func (l *Launcher) Launch(ctx context.Context, g Game) (*LaunchedProcess, error) {
	if g.BinaryPath == "" {
		return nil, fmt.Errorf("game binary path is empty")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, g.BinaryPath)
	} else {
		cmd = exec.CommandContext(ctx, g.BinaryPath)
	}

	// Set working directory to the game's directory
	if g.InstallDir != "" {
		cmd.Dir = g.InstallDir
	}

	// Redirect output to avoid blocking on pipes
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start game %s: %w", g.Title, err)
	}

	proc := &LaunchedProcess{
		Game:      g,
		Cmd:       cmd,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now().UTC(),
	}

	if l.CaptureAttachFn != nil {
		if err := l.CaptureAttachFn(proc.PID); err != nil {
			_ = proc.Terminate()
			return nil, fmt.Errorf("failed to attach capture: %w", err)
		}
	}

	return proc, nil
}

// LaunchedProcess represents a running game process.
type LaunchedProcess struct {
	Game      Game
	Cmd       *exec.Cmd
	PID       int
	StartedAt time.Time
	mu        sync.Mutex
	waitOnce  sync.Once
	waitErr   error
}

// IsRunning reports whether the process is still alive.
func (p *LaunchedProcess) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Cmd == nil || p.Cmd.Process == nil {
		return false
	}
	return p.Cmd.Process.Signal(syscall.Signal(0)) == nil
}

// Terminate sends SIGTERM (or equivalent) and waits for exit.
func (p *LaunchedProcess) Terminate() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Cmd == nil || p.Cmd.Process == nil {
		return nil
	}
	return p.Cmd.Process.Signal(os.Interrupt)
}

// Kill forcefully terminates the process.
func (p *LaunchedProcess) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Cmd == nil || p.Cmd.Process == nil {
		return nil
	}
	return p.Cmd.Process.Kill()
}

// Wait blocks until the process exits.
func (p *LaunchedProcess) Wait() error {
	if p.Cmd == nil {
		return nil
	}
	p.waitOnce.Do(func() { p.waitErr = p.Cmd.Wait() })
	return p.waitErr
}
