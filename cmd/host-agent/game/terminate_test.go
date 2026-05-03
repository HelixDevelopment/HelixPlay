package game

import (
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTerminatorGraceful(t *testing.T) {
	binary := "sleep"
	if runtime.GOOS == "windows" {
		binary = "timeout"
	}

	cmd := exec.Command(binary, "30")
	require.NoError(t, cmd.Start(), "sleep binary must be available for game terminate tests")

	proc := &LaunchedProcess{Cmd: cmd, PID: cmd.Process.Pid}
	term := NewTerminator(2 * time.Second)

	if err := term.Terminate(proc); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if proc.IsRunning() {
		t.Fatal("expected process to be terminated")
	}
}

func TestTerminatorForceKill(t *testing.T) {
	// Start a process that ignores SIGTERM
	cmd := exec.Command("sleep", "30")
	require.NoError(t, cmd.Start(), "sleep binary must be available for game terminate tests")

	proc := &LaunchedProcess{Cmd: cmd, PID: cmd.Process.Pid}
	// Very short graceful timeout to force kill path
	term := NewTerminator(1 * time.Millisecond)

	if err := term.Terminate(proc); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if proc.IsRunning() {
		t.Fatal("expected process to be force-killed")
	}
}

func TestTerminatorNilProcess(t *testing.T) {
	term := NewTerminator(time.Second)
	if err := term.Terminate(nil); err != nil {
		t.Fatalf("expected nil error for nil process, got %v", err)
	}
}

func TestSendSignal(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	require.NoError(t, cmd.Start(), "sleep binary must be available for game terminate tests")
	defer cmd.Process.Kill()

	proc := &LaunchedProcess{Cmd: cmd, PID: cmd.Process.Pid}
	if err := SendSignal(proc, syscall.Signal(0)); err != nil {
		t.Fatalf("SendSignal failed: %v", err)
	}
}
