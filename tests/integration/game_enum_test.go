package integration_test

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/game"
)

func TestGameEnumerationAcrossStores(t *testing.T) {
	stores := []string{"steam", "epic", "gog", "ubisoft", "battlenet", "origin", "microsoftstore"}

	var totalGames int
	for _, store := range stores {
		t.Run(store, func(t *testing.T) {
			enu := game.NewEnumerator(store)
			games, err := enu.Enumerate()
			if err != nil {
				t.Fatalf("enumerate %s failed: %v", store, err)
			}

			// Verify observable properties of enumerated games
			for _, g := range games {
				if g.Title == "" {
					t.Error("expected game title to be non-empty")
				}
				if g.Store != store {
					t.Errorf("expected store %q, got %q", store, g.Store)
				}
				if g.BinaryPath == "" {
					t.Error("expected binary path to be non-empty")
				}
			}

			totalGames += len(games)
		})
	}

	// Steam stub returns at least one game; total must be > 0
	if totalGames == 0 {
		t.Fatal("expected at least one game enumerated across all stores")
	}
}
