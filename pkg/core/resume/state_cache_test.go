package resume

import (
	"testing"
	"time"
)

func TestCacheSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir)

	state := State{
		UserID:        "user-1",
		GameID:        "game-1",
		GameTitle:     "Test Game",
		LastPosition:  120000,
		Timestamp:     time.Now().UTC(),
		ControllerType: "dualsense",
	}

	if err := c.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, ok := c.Load("user-1")
	if !ok {
		t.Fatal("expected state to be found")
	}
	if loaded.GameID != "game-1" {
		t.Errorf("expected game-1, got %s", loaded.GameID)
	}
	if loaded.LastPosition != 120000 {
		t.Errorf("expected position 120000, got %d", loaded.LastPosition)
	}
}

func TestCacheLastPlayed(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir)

	c.Save(State{UserID: "user-1", GameID: "game-a", Timestamp: time.Now().UTC()})

	gameID, ok := c.LastPlayed("user-1")
	if !ok {
		t.Fatal("expected last played game")
	}
	if gameID != "game-a" {
		t.Errorf("expected game-a, got %s", gameID)
	}
}

func TestCacheIsFresh(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir)

	c.Save(State{UserID: "user-1", GameID: "game-1", Timestamp: time.Now().UTC()})
	if !c.IsFresh("user-1", 5*time.Minute) {
		t.Error("expected state to be fresh")
	}

	c.Save(State{UserID: "user-1", GameID: "game-1", Timestamp: time.Now().UTC().Add(-10 * time.Minute)})
	if c.IsFresh("user-1", 5*time.Minute) {
		t.Error("expected state to be stale")
	}
}

func TestCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewCache(tmpDir)

	c.Save(State{UserID: "user-1", GameID: "game-1", Timestamp: time.Now().UTC()})
	if err := c.Clear("user-1"); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	_, ok := c.Load("user-1")
	if ok {
		t.Error("expected state to be cleared")
	}
}

func TestCacheMissing(t *testing.T) {
	c := NewCache("")
	_, ok := c.Load("unknown")
	if ok {
		t.Error("expected no state for unknown user")
	}
}
