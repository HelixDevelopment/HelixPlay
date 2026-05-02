package resume

import (
	"context"
	"sync"
	"time"
)

// Preconnector maintains warm connections to reduce cold-start latency.
type Preconnector struct {
	mu        sync.RWMutex
	addresses []string
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewPreconnector creates a preconnector for the given coordinator addresses.
func NewPreconnector(addresses []string) *Preconnector {
	return &Preconnector{
		addresses: addresses,
	}
}

// Start begins background connection warming.
func (p *Preconnector) Start(ctx context.Context) {
	cctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.done = make(chan struct{})
	go p.loop(cctx)
}

// Stop halts the preconnector.
func (p *Preconnector) Stop() {
	if p.cancel != nil {
		p.cancel()
		<-p.done
	}
}

// IsRunning reports whether the preconnector is active.
func (p *Preconnector) IsRunning() bool {
	select {
	case <-p.done:
		return false
	default:
		return p.cancel != nil
	}
}

func (p *Preconnector) loop(ctx context.Context) {
	defer close(p.done)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial warm-up
	p.warm()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.warm()
		}
	}
}

func (p *Preconnector) warm() {
	p.mu.RLock()
	addrs := make([]string, len(p.addresses))
	copy(addrs, p.addresses)
	p.mu.RUnlock()

	for _, addr := range addrs {
		_ = addr
		// Stub: in production this would attempt a TCP or QUIC handshake
	}
}
