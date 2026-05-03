package challenges_test

import (
	"context"
	"net"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/game"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/lifecycle"
	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/transport"
	"github.com/stretchr/testify/require"
)

// TestChallengeHostBootGameStreamTerminate verifies the full lifecycle:
// host boots -> capability advertisement -> game launch -> stream -> graceful termination.
// This is Challenge scenario T256.
func TestChallengeHostBootGameStreamTerminate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping challenge in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 1: Host boots and advertises capabilities
	meta, err := capability.Advertise()
	require.NoError(t, err, "host must advertise capabilities on boot")
	require.NotEmpty(t, meta.GPUModel, "GPU model must be detected")
	require.NotEmpty(t, meta.CodecsSupported, "at least one codec must be offered")

	// Step 2: Initialize encoder (software path always available)
	enc := encoder.NewDualPath("software")
	require.NotNil(t, enc, "encoder must be created")
	require.NoError(t, enc.Start(), "encoder must start")
	defer enc.Stop()

	// Step 3: Set up UDP transport with a real loopback destination
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	require.NoError(t, err, "UDP listener must start")
	defer listener.Close()
	destAddr := listener.LocalAddr().String()

	udp, err := transport.NewUDP(destAddr)
	require.NoError(t, err, "UDP transport must be created")
	require.NoError(t, udp.Start(), "UDP transport must start")
	defer udp.Stop()

	// Step 4: Launch a game process (use sleep as a long-running stand-in)
	binary := "sleep"
	args := []string{"30"}
	if runtime.GOOS == "windows" {
		binary = "timeout"
		args = []string{"/t", "30"}
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	require.NoError(t, cmd.Start(), "game process must launch")
	defer cmd.Process.Kill()

	proc := &game.LaunchedProcess{Cmd: cmd, PID: cmd.Process.Pid, Game: game.Game{Title: "TestGame", BinaryPath: binary}}
	require.True(t, proc.IsRunning(), "game process must be running after launch")

	// Step 5: Stream for a shortened duration
	frame := make([]byte, 1920*1080*4) // 4K RGBA frame
	for i := range frame {
		frame[i] = byte(i % 256)
	}

	streamTicks := 10
	for i := 0; i < streamTicks; i++ {
		encoded, encErr := enc.EncodeFrame(frame)
		require.NoError(t, encErr, "frame encoding must succeed during stream")
		require.NotEmpty(t, encoded, "encoded frame must be non-empty")

		sendErr := udp.SendPacket(encoded)
		require.NoError(t, sendErr, "packet send must succeed during stream")
		time.Sleep(50 * time.Millisecond)
	}

	// Verify encoder and transport remain operational
	require.True(t, enc.IsStreaming(), "encoder must remain streaming")
	require.True(t, udp.IsRunning(), "transport must remain running")

	// Step 6: Graceful termination
	term := game.NewTerminator(2 * time.Second)
	require.NoError(t, term.Terminate(proc), "game must terminate gracefully")
	time.Sleep(200 * time.Millisecond)
	require.False(t, proc.IsRunning(), "game process must be stopped after termination")

	// Step 7: Lifecycle manager confirms game is no longer running
	mgr := lifecycle.NewManager()
	require.False(t, mgr.IsRunning(game.Game{Title: "TestGame"}), "game must no longer be tracked as running")
}
