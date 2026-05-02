package integration_test

import (
	"testing"
	"time"

	"github.com/HelixDevelopment/HelixPlay/cmd/host-agent/input"
	"digital.vasic.memory/pkg/memfd"
)

func TestVirtualControllerInputEndToEnd(t *testing.T) {
	// Create a lock-free ring buffer for input events (real dependency, no mocks)
	rb := memfd.NewPSC(4096)
	defer rb.Close()

	// Connect virtual DualSense controller
	ds := input.NewDualSense()
	if err := ds.Connect(); err != nil {
		t.Fatalf("failed to connect DualSense: %v", err)
	}
	if !ds.IsConnected() {
		t.Fatal("expected DualSense to be connected")
	}

	// Configure haptics (observable state mutation)
	if err := ds.SetHaptics(0.5, 0.7); err != nil {
		t.Fatalf("failed to set haptics: %v", err)
	}

	// Configure adaptive trigger resistance
	if err := ds.SetTriggerResistance("L2", 0.8); err != nil {
		t.Fatalf("failed to set L2 trigger resistance: %v", err)
	}
	if got := ds.GetTriggerResistance("L2"); got != 0.8 {
		t.Errorf("expected L2 trigger resistance 0.8, got %v", got)
	}

	// Verify IMU read returns data when connected
	gyro, accel := ds.ReadIMU()
	if gyro == [3]float64{0, 0, 0} {
		t.Error("expected non-zero gyroscope data")
	}
	if accel == [3]float64{0, 0, 0} {
		t.Error("expected non-zero accelerometer data")
	}

	// Start USB poller feeding input events into the ring buffer
	poller := input.NewUSBPoller(rb, 1000)
	if err := poller.Start(); err != nil {
		t.Fatalf("failed to start USB poller: %v", err)
	}
	if !poller.IsPolling() {
		t.Fatal("expected USB poller to be polling")
	}

	// Allow the poller goroutine to write events into the ring buffer
	time.Sleep(50 * time.Millisecond)

	// Observable: read events back from the ring buffer
	buf := make([]byte, 64)
	n, err := rb.Read(buf)
	if err != nil {
		t.Fatalf("failed to read from ring buffer: %v", err)
	}
	if n == 0 {
		t.Fatal("expected non-zero bytes read from ring buffer")
	}

	// Stop the input pipeline
	poller.Stop()
	if poller.IsPolling() {
		t.Error("expected USB poller to be stopped")
	}

	// Disconnect controller
	ds.Disconnect()
	if ds.IsConnected() {
		t.Error("expected DualSense to be disconnected")
	}

	// Verify IMU returns zero after disconnect
	gyroPost, accelPost := ds.ReadIMU()
	if gyroPost != [3]float64{0, 0, 0} {
		t.Error("expected zero gyroscope data after disconnect")
	}
	if accelPost != [3]float64{0, 0, 0} {
		t.Error("expected zero accelerometer data after disconnect")
	}
}
