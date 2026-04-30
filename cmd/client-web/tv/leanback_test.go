package tv_test

import (
    "testing"
    "github.com/HelixDevelopment/HelixPlay/cmd/client-web/tv"
)

func TestLeanbackStart(t *testing.T) {
    lb := tv.NewLeanback()
    if err := lb.Start(); err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    if !lb.IsActive() {
        t.Error("Expected Leanback UI to be active")
    }
}

func TestDPadNavigation(t *testing.T) {
    lb := tv.NewLeanback()
    lb.Start()
    
    // Simulate D-pad Up
    err := lb.HandleDPad("Up")
    if err != nil {
        t.Fatalf("HandleDPad Up failed: %v", err)
    }
    
    // Simulate D-pad Down
    err = lb.HandleDPad("Down")
    if err != nil {
        t.Fatalf("HandleDPad Down failed: %v", err)
    }
    
    // Simulate D-pad Select
    err = lb.HandleDPad("Select")
    if err != nil {
        t.Fatalf("HandleDPad Select failed: %v", err)
    }
}

func TestLeanbackStop(t *testing.T) {
    lb := tv.NewLeanback()
    lb.Start()
    lb.Stop()
    if lb.IsActive() {
        t.Error("Expected Leanback UI to be stopped")
    }
}

func TestFocusManagement(t *testing.T) {
    lb := tv.NewLeanback()
    lb.Start()
    
    focus := lb.GetFocusedItem()
    if focus == "" {
        t.Error("Expected a focused item")
    }
    
    lb.MoveFocus("Down")
    newFocus := lb.GetFocusedItem()
    if newFocus == focus {
        t.Error("Expected focus to change after D-pad navigation")
    }
}
