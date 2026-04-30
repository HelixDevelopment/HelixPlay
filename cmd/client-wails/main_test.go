package main_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/client-wails/backend"
)

func TestWailsStart(t *testing.T) {
    b := backend.NewBackend()
    if err := b.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    if !b.IsRunning() {
        t.Error("Expected backend to be running")
    }
}

func TestWailsStop(t *testing.T) {
    b := backend.NewBackend()
    b.Start()
    b.Stop()
    if b.IsRunning() {
        t.Error("Expected backend to be stopped")
    }
}

func TestWailsConnect(t *testing.T) {
    b := backend.NewBackend()
    b.Start()
    err := b.Connect("localhost:50051")
    if err != nil {
        t.Fatalf("Connect failed: %v", err)
    }
    if !b.IsConnected() {
        t.Error("Expected to be connected")
    }
}
