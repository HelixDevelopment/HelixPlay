package challenges_test

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/game"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/lifecycle"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/resume"
	"github.com/stretchr/testify/require"
)

// TestChallengeQuickResumeFromIdle verifies that a session can resume
// from idle within 5 seconds. Challenge scenario T258.
func TestChallengeQuickResumeFromIdle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping challenge in short mode")
	}

	ctx := context.Background()

	// Step 1: Preconnect to reduce cold-start latency
	pc := resume.NewPreconnector([]string{"127.0.0.1:50051"})
	require.NotNil(t, pc, "preconnector must be created")
	ctx2, cancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel2()
	pc.Start(ctx2)
	require.True(t, pc.IsRunning(), "preconnector must be running after Start")
	defer pc.Stop()

	// Step 2: Launch a game process
	binary := "sleep"
	args := []string{"10"}
	if runtime.GOOS == "windows" {
		binary = "timeout"
		args = []string{"/t", "10"}
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	require.NoError(t, cmd.Start(), "game process must launch")
	defer cmd.Process.Kill()

	game := game.Game{Title: "QuickResumeTest", BinaryPath: binary}
	mgr := lifecycle.NewManager()
	require.NoError(t, mgr.Launch(game), "game must launch via lifecycle manager")
	require.True(t, mgr.IsRunning(game), "game must be running after launch")

	// Step 3: Simulate idle (pause) - very short for test
	time.Sleep(200 * time.Millisecond)

	// Step 4: Quick resume must complete within 5 seconds
	start := time.Now()
	require.NoError(t, mgr.QuickResume(game), "quick resume must succeed")
	elapsed := time.Since(start)

	require.Less(t, elapsed, 5*time.Second, "quick resume must complete within 5 seconds")
	require.True(t, mgr.IsRunning(game), "game must still be running after quick resume")
	require.NoError(t, mgr.Terminate(game), "game must terminate cleanly")
}
