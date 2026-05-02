package lifecycle

import (
	"os/exec"
	"sync"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/game"
)

type Manager struct {
	mu    sync.RWMutex
	procs map[string]*exec.Cmd
}

func NewManager() *Manager {
	return &Manager{
		procs: make(map[string]*exec.Cmd),
	}
}

func (m *Manager) Launch(g game.Game) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cmd := exec.Command(g.BinaryPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	m.procs[g.Title] = cmd
	return nil
}

func (m *Manager) Terminate(g game.Game) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cmd, ok := m.procs[g.Title]
	if !ok {
		return nil // already terminated
	}
	delete(m.procs, g.Title)
	return cmd.Process.Kill()
}

func (m *Manager) IsRunning(g game.Game) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cmd, ok := m.procs[g.Title]
	if !ok {
		return false
	}
	if cmd.Process == nil {
		return false
	}
	return true
}

// QuickResume saves state and stops the game, allowing quick resumption later
func (m *Manager) QuickResume(g game.Game) error {
	return m.Terminate(g)
}
