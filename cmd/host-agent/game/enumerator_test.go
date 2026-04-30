package game_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/game"
)

func TestEnumerateSteamGames(t *testing.T) {
    enu := game.NewEnumerator("steam")
    games, err := enu.Enumerate()
    if err != nil {
        t.Fatalf("Enumerate failed: %v", err)
    }
    if len(games) == 0 {
        t.Error("Expected at least one game")
    }
}
