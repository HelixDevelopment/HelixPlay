package input

import (
	"testing"
)

func TestManagerUpdateAndGet(t *testing.T) {
	m := NewManager()
	st := ControllerState{
		ID:         0,
		Connected:  true,
		ButtonMask: 0x1000,
		LeftStickX: -1000,
	}
	m.UpdateState(st)

	got, ok := m.GetState(0)
	if !ok {
		t.Fatal("Expected state to exist")
	}
	if got.ButtonMask != 0x1000 {
		t.Errorf("Expected ButtonMask 0x1000, got 0x%04x", got.ButtonMask)
	}
	if got.LeftStickX != -1000 {
		t.Errorf("Expected LeftStickX -1000, got %d", got.LeftStickX)
	}
	if got.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestManagerList(t *testing.T) {
	m := NewManager()
	m.UpdateState(ControllerState{ID: 0, Connected: true})
	m.UpdateState(ControllerState{ID: 1, Connected: false})

	list := m.List()
	if len(list) != 2 {
		t.Fatalf("Expected 2 controllers, got %d", len(list))
	}
}

func TestManagerRemove(t *testing.T) {
	m := NewManager()
	m.UpdateState(ControllerState{ID: 0})
	m.Remove(0)

	if _, ok := m.GetState(0); ok {
		t.Error("Expected controller to be removed")
	}
}

func TestManagerAnyConnected(t *testing.T) {
	m := NewManager()
	if m.AnyConnected() {
		t.Error("Expected no connected controllers")
	}
	m.UpdateState(ControllerState{ID: 0, Connected: true})
	if !m.AnyConnected() {
		t.Error("Expected at least one connected controller")
	}
}
