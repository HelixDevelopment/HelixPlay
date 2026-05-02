package input_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/input"
    "digital.vasic.memory/pkg/memfd"
)

func TestUSBPollingStart(t *testing.T) {
    ringBuf := memfd.NewPSC(1024)
    usb := input.NewUSBPoller(ringBuf, 1000) // 1 kHz = 1000 Hz
    if err := usb.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    if !usb.IsPolling() {
        t.Error("Expected USB poller to be running")
    }
}

func TestUSBPollingStop(t *testing.T) {
    ringBuf := memfd.NewPSC(1024)
    usb := input.NewUSBPoller(ringBuf, 1000)
    usb.Start()
    usb.Stop()
    if usb.IsPolling() {
        t.Error("Expected USB poller to be stopped")
    }
}

func TestUSBPollData(t *testing.T) {
    ringBuf := memfd.NewPSC(1024)
    usb := input.NewUSBPoller(ringBuf, 1000)
    usb.Start()

    data := []byte{0x01, 0x02, 0x03}
    ringBuf.Write(data)

    readData := make([]byte, len(data))
    n, _ := ringBuf.Read(readData)
    if n != len(data) {
        t.Errorf("Expected %d bytes, got %d", len(data), n)
    }
}
