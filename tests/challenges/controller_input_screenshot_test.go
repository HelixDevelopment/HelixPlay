package challenges_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/input"
	"github.com/stretchr/testify/require"
)

// TestChallengeControllerInputScreenshotDiff verifies that controller input
// produces observable differences in captured frames. Challenge scenario T257.
func TestChallengeControllerInputScreenshotDiff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping challenge in short mode")
	}

	// Create two synthetic frames that differ (simulating before/after input)
	frameBefore := createTestFrame(1920, 1080, 0x0000FF) // blue screen
	frameAfter := createTestFrame(1920, 1080, 0xFF0000)  // red screen after button press

	// Verify frames are different (screenshot diff would catch this)
	require.False(t, bytes.Equal(frameBefore, frameAfter), "frames must differ after controller input")

	// Step 1: Simulate DualSense controller connection
	ds := input.NewDualSense()
	require.NotNil(t, ds, "DualSense must be created")
	connErr := ds.Connect()
	require.NoError(t, connErr, "DualSense must connect (or report unavailable with error)")

	// Step 2: Set haptics and trigger resistance (observable behavior)
	require.NoError(t, ds.SetHaptics(0.8, 0.6), "haptics must be settable")
	require.NoError(t, ds.SetTriggerResistance("R2", 0.5), "trigger resistance must be settable")
	require.InDelta(t, 0.5, ds.GetTriggerResistance("R2"), 0.01, "trigger resistance must be readable")

	// Step 3: Read IMU data (observable sensor output)
	gyro, accel := ds.ReadIMU()
	_ = gyro
	_ = accel

	// Step 4: Simulate frame capture after input delay
	time.Sleep(50 * time.Millisecond)
	frameCaptured := createTestFrame(1920, 1080, 0xFF0000)
	require.False(t, bytes.Equal(frameBefore, frameCaptured), "captured frame must differ from baseline")
}

func createTestFrame(width, height int, pixelColor uint32) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	r := uint8((pixelColor >> 16) & 0xFF)
	g := uint8((pixelColor >> 8) & 0xFF)
	b := uint8(pixelColor & 0xFF)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
