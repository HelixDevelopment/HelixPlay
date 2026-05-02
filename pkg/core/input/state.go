package input

import (
	"sync"
	"time"
)

// ControllerState holds the normalized state of a single gamepad.
type ControllerState struct {
	ID            int
	Connected     bool
	ButtonMask    uint32
	LeftTrigger   uint8
	RightTrigger  uint8
	LeftStickX    int16
	LeftStickY    int16
	RightStickX   int16
	RightStickY   int16
	DPadX         int8 // -1, 0, 1
	DPadY         int8 // -1, 0, 1
	Timestamp     time.Time
}

// Manager maintains a thread-safe registry of controller states.
type Manager struct {
	mu          sync.RWMutex
	controllers map[int]ControllerState
}

// NewManager creates a new input manager.
func NewManager() *Manager {
	return &Manager{
		controllers: make(map[int]ControllerState),
	}
}

// UpdateState applies a new state snapshot for a controller.
func (m *Manager) UpdateState(state ControllerState) {
	state.Timestamp = time.Now().UTC()
	m.mu.Lock()
	m.controllers[state.ID] = state
	m.mu.Unlock()
}

// GetState returns the latest known state for a controller.
func (m *Manager) GetState(id int) (ControllerState, bool) {
	m.mu.RLock()
	s, ok := m.controllers[id]
	m.mu.RUnlock()
	return s, ok
}

// List returns a snapshot of all known controllers.
func (m *Manager) List() []ControllerState {
	m.mu.RLock()
	out := make([]ControllerState, 0, len(m.controllers))
	for _, s := range m.controllers {
		out = append(out, s)
	}
	m.mu.RUnlock()
	return out
}

// Remove deletes a controller from the registry.
func (m *Manager) Remove(id int) {
	m.mu.Lock()
	delete(m.controllers, id)
	m.mu.Unlock()
}

// Count returns the number of tracked controllers.
func (m *Manager) Count() int {
	m.mu.RLock()
	n := len(m.controllers)
	m.mu.RUnlock()
	return n
}

// AnyConnected returns true if at least one controller is connected.
func (m *Manager) AnyConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.controllers {
		if s.Connected {
			return true
		}
	}
	return false
}
