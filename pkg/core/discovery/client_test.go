package discovery

import (
	"context"
	"testing"
	"time"
)

func TestClientStartStop(t *testing.T) {
	c := NewClient("", "tenant-1", "us-east")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if c.cancel == nil {
		t.Error("Expected cancel func to be set")
	}
	c.Stop()
}

func TestHostMerge(t *testing.T) {
	c := NewClient("", "tenant-1", "us-east")
	c.hosts = make(map[string]Host)
	c.mdnsHosts = make(map[string]Host)
	c.rendezvousHosts = make(map[string]Host)

	h1 := Host{ID: "host-1", Name: "MDNS Host", LocalEndpoint: "192.168.1.10:50051", Status: "online"}
	c.mdnsHosts["host-1"] = h1
	c.mergeHost("host-1")

	hosts := c.Hosts()
	if len(hosts) != 1 {
		t.Fatalf("Expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Name != "MDNS Host" {
		t.Errorf("Expected Name 'MDNS Host', got %s", hosts[0].Name)
	}

	h2 := Host{ID: "host-1", Name: "", PublicEndpoint: "1.2.3.4:50051", Status: "streaming", ActiveSessions: 2}
	c.rendezvousHosts["host-1"] = h2
	c.mergeHost("host-1")

	hosts = c.Hosts()
	if len(hosts) != 1 {
		t.Fatalf("Expected 1 host after merge, got %d", len(hosts))
	}
	merged := hosts[0]
	if merged.PublicEndpoint != "1.2.3.4:50051" {
		t.Errorf("Expected PublicEndpoint from rendezvous, got %s", merged.PublicEndpoint)
	}
	if merged.LocalEndpoint != "192.168.1.10:50051" {
		t.Errorf("Expected LocalEndpoint from mDNS, got %s", merged.LocalEndpoint)
	}
	if merged.Status != "streaming" {
		t.Errorf("Expected Status from rendezvous, got %s", merged.Status)
	}
	if merged.ActiveSessions != 2 {
		t.Errorf("Expected ActiveSessions 2, got %d", merged.ActiveSessions)
	}
}

func TestWatch(t *testing.T) {
	c := NewClient("", "tenant-1", "us-east")
	c.hosts = make(map[string]Host)
	c.mdnsHosts = make(map[string]Host)
	c.rendezvousHosts = make(map[string]Host)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := c.Watch(ctx)

	c.setHost("host-1", Host{ID: "host-1", Name: "Test", Status: "online"})

	select {
	case ev := <-ch:
		if ev.Type != EventTypeAdded {
			t.Errorf("Expected added event, got %s", ev.Type)
		}
		if ev.Host.ID != "host-1" {
			t.Errorf("Expected host-1, got %s", ev.Host.ID)
		}
	case <-time.After(time.Second):
		t.Error("Timed out waiting for event")
	}
}
