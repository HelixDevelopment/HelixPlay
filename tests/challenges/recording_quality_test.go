package challenges_test

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/recording"
	"github.com/stretchr/testify/require"
)

// TestChallengeRecordingPlaybackQuality verifies that a recording session
// can be started and stopped, producing a valid recording entry. Challenge T259.
func TestChallengeRecordingPlaybackQuality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping challenge in short mode")
	}

	// Step 1: Initialize recording manager
	mgr := recording.NewManager(recording.Config{
		OutputDir:   "/tmp/helixplay_test_recordings",
		MaxSegments: 5,
	})
	require.NotNil(t, mgr, "recording manager must be created")

	// Step 2: Start a recording session
	sessionID := "sess-quality-test"
	rec, err := mgr.Start(sessionID, "mkv")
	require.NoError(t, err, "recording must start")
	require.NotNil(t, rec, "recording object must be returned")
	require.Equal(t, recording.StatusRecording, rec.Status, "recording must be active")
	require.NotEmpty(t, rec.ID, "recording must have an ID")

	// Step 3: Rotate a segment (simulates real DVR behavior)
	seg, err := mgr.RotateSegment(rec.ID)
	require.NoError(t, err, "segment rotation must succeed")
	require.NotNil(t, seg, "segment must be returned")
	require.NotEmpty(t, seg.Path, "segment must have a file path")

	// Step 4: Verify active count
	require.Equal(t, 1, mgr.ActiveCount(), "exactly one recording must be active")

	// Step 5: Stop recording and verify completion
	stopped, err := mgr.Stop(rec.ID)
	require.NoError(t, err, "recording must stop cleanly")
	require.NotNil(t, stopped, "stopped recording must be returned")
	require.Equal(t, recording.StatusComplete, stopped.Status, "recording must be complete after stop")

	// Step 6: Verify recording is retrievable
	got, ok := mgr.Get(rec.ID)
	require.True(t, ok, "recording must be retrievable after stop")
	require.Equal(t, rec.ID, got.ID, "retrieved recording ID must match")
}
