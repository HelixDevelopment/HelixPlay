package encoder_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/host-agent/encoder"
)

func TestDualPathStart(t *testing.T) {
    dp := encoder.NewDualPath("nvenc")
    if err := dp.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    if !dp.IsStreaming() {
        t.Error("Expected streaming to be active")
    }
    if !dp.IsRecording() {
        t.Error("Expected recording to be active")
    }
}

func TestDualPathStop(t *testing.T) {
    dp := encoder.NewDualPath("nvenc")
    dp.Start()
    dp.Stop()
    if dp.IsStreaming() {
        t.Error("Expected streaming to be stopped")
    }
    if dp.IsRecording() {
        t.Error("Expected recording to be stopped")
    }
}

func TestDualPathGetStreamOutput(t *testing.T) {
    dp := encoder.NewDualPath("nvenc")
    dp.Start()
    output := dp.GetStreamOutput()
    if len(output) == 0 {
        t.Error("Expected non-empty stream output")
    }
}

func TestDualPathGetRecordOutput(t *testing.T) {
    dp := encoder.NewDualPath("nvenc")
    dp.Start()
    output := dp.GetRecordOutput()
    if len(output) == 0 {
        t.Error("Expected non-empty record output")
    }
}
