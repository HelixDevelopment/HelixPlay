// Package resume provides quick resume state caching for instant game relaunch.
package resume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// State holds the last played game and position for quick resume.
type State struct {
	UserID        string    `json:"user_id"`
	GameID        string    `json:"game_id"`
	GameTitle     string    `json:"game_title"`
	LastPosition  int64     `json:"last_position_ms"`
	Timestamp     time.Time `json:"timestamp"`
	ControllerType string   `json:"controller_type"`
}

// Cache persists resume state to disk.
type Cache struct {
	mu      sync.RWMutex
	states  map[string]State // keyed by user ID
	dataDir string
}

// NewCache creates a resume state cache.
func NewCache(dataDir string) *Cache {
	return &Cache{
		states:  make(map[string]State),
		dataDir: dataDir,
	}
}

// Save stores the resume state for a user.
func (c *Cache) Save(state State) error {
	c.mu.Lock()
	c.states[state.UserID] = state
	c.mu.Unlock()
	return c.persist(state.UserID)
}

// Load retrieves the resume state for a user.
func (c *Cache) Load(userID string) (State, bool) {
	c.mu.RLock()
	state, ok := c.states[userID]
	c.mu.RUnlock()
	return state, ok
}

// Clear removes the resume state for a user.
func (c *Cache) Clear(userID string) error {
	c.mu.Lock()
	delete(c.states, userID)
	c.mu.Unlock()
	return os.Remove(c.filePath(userID))
}

// LastPlayed returns the game ID of the most recently played game.
func (c *Cache) LastPlayed(userID string) (string, bool) {
	state, ok := c.Load(userID)
	if !ok {
		return "", false
	}
	return state.GameID, true
}

// IsFresh reports whether the cached state is within the freshness window.
func (c *Cache) IsFresh(userID string, maxAge time.Duration) bool {
	state, ok := c.Load(userID)
	if !ok {
		return false
	}
	return time.Since(state.Timestamp) <= maxAge
}

func (c *Cache) persist(userID string) error {
	if c.dataDir == "" {
		return nil
	}
	if err := os.MkdirAll(c.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}

	c.mu.RLock()
	state := c.states[userID]
	c.mu.RUnlock()

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	path := c.filePath(userID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return nil
}

func (c *Cache) filePath(userID string) string {
	return filepath.Join(c.dataDir, userID+".json")
}
