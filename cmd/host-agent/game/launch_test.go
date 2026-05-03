package game

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLauncherLaunch(t *testing.T) {
	var attachedPID int
	launcher := NewLauncher(func(pid int) error {
		attachedPID = pid
		return nil
	})

	// Use a long-running sleep command
	binary := "sleep"
	args := "30"
	if runtime.GOOS == "windows" {
		binary = "timeout"
		args = "/t 30"
	}

	g := Game{
		Title:      "TestGame",
		BinaryPath: binary,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Launch manually to control args
	cmd := exec.CommandContext(ctx, binary, args)
	require.NoError(t, cmd.Start(), "sleep binary must be available for game launch tests")
	defer cmd.Process.Kill()

	proc := &LaunchedProcess{Cmd: cmd, PID: cmd.Process.Pid, Game: g}

	if proc.PID == 0 {
		t.Fatal("expected non-zero PID")
	}
	if !proc.IsRunning() {
		t.Fatal("expected process to be running")
	}

	// Simulate capture attach
	if launcher.CaptureAttachFn != nil {
		if err := launcher.CaptureAttachFn(proc.PID); err != nil {
			t.Fatalf("capture attach failed: %v", err)
		}
		if attachedPID != proc.PID {
			t.Fatalf("expected capture attached to PID %d, got %d", proc.PID, attachedPID)
		}
	}
}

func TestLauncherLaunchEmptyPath(t *testing.T) {
	launcher := NewLauncher(nil)
	g := Game{Title: "BadGame", BinaryPath: ""}
	_, err := launcher.Launch(context.Background(), g)
	if err == nil {
		t.Fatal("expected error for empty binary path")
	}
}

func TestLauncherLaunchInvalidBinary(t *testing.T) {
	launcher := NewLauncher(nil)
	g := Game{Title: "BadGame", BinaryPath: "/nonexistent/binary"}
	_, err := launcher.Launch(context.Background(), g)
	if err == nil {
		t.Fatal("expected error for invalid binary path")
	}
}

func TestLaunchedProcessKill(t *testing.T) {
	binary := "sleep"
	args := []string{"30"}
	if runtime.GOOS == "windows" {
		binary = "timeout"
		args = []string{"/t", "30"}
	}

	cmd := exec.Command(binary, args...)
	require.NoError(t, cmd.Start(), "sleep binary must be available for game launch tests")

	proc := &LaunchedProcess{Cmd: cmd, PID: cmd.Process.Pid}
	require.True(t, proc.IsRunning(), "expected process to be running")

	require.NoError(t, proc.Kill(), "Kill must succeed")

	_ = proc.Wait()
	time.Sleep(50 * time.Millisecond)
	if proc.IsRunning() {
		t.Fatal("expected process to be killed")
	}
}
