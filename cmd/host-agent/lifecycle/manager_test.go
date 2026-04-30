package lifecycle_test

import (
	"testing"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/game"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/lifecycle"
)

func TestLaunchGame(t *testing.T) {
	mgr := lifecycle.NewManager()
	g := game.Game{Title: "Test Game", BinaryPath: "/bin/true"}
	err := mgr.Launch(g)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}
	if !mgr.IsRunning(g) {
		t.Error("Expected game to be running")
	}
}

func TestTerminateGame(t *testing.T) {
	mgr := lifecycle.NewManager()
	g := game.Game{Title: "Test Game", BinaryPath: "/bin/sleep"}
	mgr.Launch(g)
	err := mgr.Terminate(g)
	if err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}
	if mgr.IsRunning(g) {
		t.Error("Expected game to be stopped")
	}
}
