package resume

import (
	"context"
	"testing"
	"time"
)

func TestPreconnectorStartStop(t *testing.T) {
	p := NewPreconnector([]string{"localhost:8080"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Start(ctx)
	if !p.IsRunning() {
		t.Fatal("expected preconnector to be running")
	}

	p.Stop()
	if p.IsRunning() {
		t.Fatal("expected preconnector to be stopped")
	}
}

func TestPreconnectorMultipleAddresses(t *testing.T) {
	p := NewPreconnector([]string{"host1:8080", "host2:8080"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	p.Stop()
}
