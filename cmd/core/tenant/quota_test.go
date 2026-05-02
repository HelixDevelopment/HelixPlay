package tenant

import (
	"testing"

	"github.com/HelixDevelopment/HelixPlay/pkg/models"
)

func TestEnforcerCanStartSession(t *testing.T) {
	e := NewEnforcer()
	e.SetQuota("t1", Quota{MaxSessions: 2})

	if err := e.CanStartSession("t1"); err != nil {
		t.Fatalf("expected allowed: %v", err)
	}

	e.RecordSessionStart("t1")
	e.RecordSessionStart("t1")

	if err := e.CanStartSession("t1"); err == nil {
		t.Fatal("expected quota exceeded")
	}
}

func TestEnforcerNoQuota(t *testing.T) {
	e := NewEnforcer()
	if err := e.CanStartSession("unknown"); err == nil {
		t.Fatal("expected error for missing quota")
	}
}

func TestEnforcerSessionTracking(t *testing.T) {
	e := NewEnforcer()
	e.SetQuota("t1", Quota{MaxSessions: 5})

	e.RecordSessionStart("t1")
	e.RecordSessionStart("t1")
	if u := e.Usage("t1"); u.ActiveSessions != 2 {
		t.Fatalf("expected 2 active sessions, got %d", u.ActiveSessions)
	}

	e.RecordSessionEnd("t1")
	if u := e.Usage("t1"); u.ActiveSessions != 1 {
		t.Fatalf("expected 1 active session, got %d", u.ActiveSessions)
	}

	// Should not go negative
	e.RecordSessionEnd("t1")
	e.RecordSessionEnd("t1")
	if u := e.Usage("t1"); u.ActiveSessions != 0 {
		t.Fatalf("expected 0 active sessions, got %d", u.ActiveSessions)
	}
}

func TestEnforcerApplyDefaults(t *testing.T) {
	e := NewEnforcer()
	tenants := []*models.Tenant{
		{ID: "t1", ResourceQuotaMaxSessions: 10, ResourceQuotaStorageGB: 100},
		{ID: "t2", ResourceQuotaMaxSessions: 5, ResourceQuotaStorageGB: 50},
	}

	e.ApplyDefaults(tenants)

	q1, ok := e.GetQuota("t1")
	if !ok {
		t.Fatal("expected quota for t1")
	}
	if q1.MaxSessions != 10 {
		t.Errorf("expected 10 sessions, got %d", q1.MaxSessions)
	}
	if q1.MaxStorageGB != 100 {
		t.Errorf("expected 100 GB, got %d", q1.MaxStorageGB)
	}
}

func TestDefaultQuota(t *testing.T) {
	q := DefaultQuota()
	if q.MaxSessions <= 0 {
		t.Fatal("expected positive max sessions")
	}
	if q.MaxStorageGB <= 0 {
		t.Fatal("expected positive max storage")
	}
}
