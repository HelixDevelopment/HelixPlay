package game

import (
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMonitorStartStop(t *testing.T) {
	m := NewMonitor(100 * time.Millisecond)
	m.Start()
	if m.Count() != 0 {
		t.Fatalf("expected 0 monitored processes, got %d", m.Count())
	}
	m.Stop()
}

func TestMonitorDetectsExit(t *testing.T) {
	binary := "sleep"
	args := []string{"0.2"}
	if runtime.GOOS == "windows" {
		binary = "timeout"
		args = []string{"/t", "0"}
	}

	cmd := exec.Command(binary, args...)
	require.NoError(t, cmd.Start(), "sleep binary must be available for game monitor tests")

	proc := &LaunchedProcess{Cmd: cmd, PID: cmd.Process.Pid, Game: Game{Title: "ShortGame"}}
	// Reap process in background so Signal(0) returns false promptly after exit
	go func() { _ = proc.Wait() }()

	m := NewMonitor(50 * time.Millisecond)
	m.Start()
	defer m.Stop()

	mp := &MonitoredProcess{
		LaunchedProcess: proc,
		CrashHandler: func(gp Game, exitErr error) {
			// crash handler invoked
		},
	}
	m.Add(mp)

	if !m.IsRunning("ShortGame") {
		t.Fatal("expected game to be running initially")
	}

	// Wait for process to exit and monitor to detect it
	time.Sleep(400 * time.Millisecond)

	if m.IsRunning("ShortGame") {
		t.Fatal("expected game to no longer be running")
	}
	if m.Count() != 0 {
		t.Fatalf("expected 0 monitored processes after exit, got %d", m.Count())
	}
}

func TestMonitorAddRemove(t *testing.T) {
	m := NewMonitor(time.Second)
	m.Start()
	defer m.Stop()

	mp := &MonitoredProcess{
		LaunchedProcess: &LaunchedProcess{Game: Game{Title: "TestGame"}},
	}
	m.Add(mp)
	if m.Count() != 1 {
		t.Fatalf("expected 1 process, got %d", m.Count())
	}

	m.Remove("TestGame")
	if m.Count() != 0 {
		t.Fatalf("expected 0 processes after remove, got %d", m.Count())
	}
}
