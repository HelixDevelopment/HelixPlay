package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/catalog"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/discovery"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/input"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/streaming"
	catalogizerv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/catalog"
)

// Backend is the Wails desktop client backend.
type Backend struct {
	mu        sync.RWMutex
	running   bool
	connected bool
	address   string

	discoveryClient *discovery.Client
	catalogClient   *catalog.Client
	streamCtrl      *streaming.SessionController
	inputMgr        *input.Manager

	ctx    context.Context
	cancel context.CancelFunc
}

// NewBackend creates a new backend instance.
func NewBackend() *Backend {
	return &Backend{
		inputMgr: input.NewManager(),
	}
}

// Start initializes the backend.
func (b *Backend) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.running {
		return fmt.Errorf("already running")
	}
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.running = true
	return nil
}

// Stop tears down the backend.
func (b *Backend) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return
	}
	if b.streamCtrl != nil {
		b.streamCtrl.Disconnect()
		b.streamCtrl = nil
	}
	if b.discoveryClient != nil {
		b.discoveryClient.Stop()
		b.discoveryClient = nil
	}
	if b.catalogClient != nil {
		b.catalogClient.Close()
		b.catalogClient = nil
	}
	if b.cancel != nil {
		b.cancel()
	}
	b.running = false
	b.connected = false
}

// IsRunning reports whether the backend is active.
func (b *Backend) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// Connect establishes a connection to a host with default configuration.
func (b *Backend) Connect(address string) error {
	return b.ConnectWithConfig(address, streaming.Config{
		HostAddress: address,
		ClientID:    "wails-client",
	})
}

// ConnectWithConfig establishes a streaming session with full configuration.
func (b *Backend) ConnectWithConfig(address string, cfg streaming.Config) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return fmt.Errorf("backend not running")
	}
	b.address = address
	b.connected = true

	b.streamCtrl = streaming.NewSessionController(cfg)
	return b.streamCtrl.Connect(b.ctx)
}

// Disconnect ends the current session.
func (b *Backend) Disconnect() {
	b.mu.Lock()
	sc := b.streamCtrl
	b.streamCtrl = nil
	b.connected = false
	b.mu.Unlock()
	if sc != nil {
		sc.Disconnect()
	}
}

// IsConnected reports whether a connection has been established.
func (b *Backend) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// IsStreaming reports whether the session is actively streaming.
func (b *Backend) IsStreaming() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected && b.streamCtrl != nil && b.streamCtrl.State() == streaming.StateStreaming
}

// StreamingState returns the current streaming session state.
func (b *Backend) StreamingState() streaming.SessionState {
	b.mu.RLock()
	sc := b.streamCtrl
	b.mu.RUnlock()
	if sc == nil {
		return streaming.StateIdle
	}
	return sc.State()
}

// StartDiscovery begins host discovery using mDNS and rendezvous.
func (b *Backend) StartDiscovery(rendezvousAddr, tenantID, region string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return fmt.Errorf("backend not running")
	}
	if b.discoveryClient != nil {
		return fmt.Errorf("discovery already started")
	}
	b.discoveryClient = discovery.NewClient(rendezvousAddr, tenantID, region)
	return b.discoveryClient.Start(b.ctx)
}

// StopDiscovery halts host discovery.
func (b *Backend) StopDiscovery() {
	b.mu.Lock()
	dc := b.discoveryClient
	b.discoveryClient = nil
	b.mu.Unlock()
	if dc != nil {
		dc.Stop()
	}
}

// DiscoveredHosts returns the current list of discovered hosts.
func (b *Backend) DiscoveredHosts() []discovery.Host {
	b.mu.RLock()
	dc := b.discoveryClient
	b.mu.RUnlock()
	if dc == nil {
		return nil
	}
	return dc.Hosts()
}

// UpdateControllerState accepts controller input from the frontend or native capture.
func (b *Backend) UpdateControllerState(state input.ControllerState) error {
	b.mu.RLock()
	sc := b.streamCtrl
	b.mu.RUnlock()
	if sc != nil {
		return sc.SendControllerState(state)
	}
	b.inputMgr.UpdateState(state)
	return nil
}

// ControllerStates returns all known controller states.
func (b *Backend) ControllerStates() []input.ControllerState {
	return b.inputMgr.List()
}

// OpenCatalog dials the catalog service.
func (b *Backend) OpenCatalog(catalogAddr string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return fmt.Errorf("backend not running")
	}
	if b.catalogClient != nil {
		b.catalogClient.Close()
	}
	client, err := catalog.NewClient(catalogAddr)
	if err != nil {
		return err
	}
	b.catalogClient = client
	return nil
}

// CloseCatalog closes the catalog connection.
func (b *Backend) CloseCatalog() {
	b.mu.Lock()
	cc := b.catalogClient
	b.catalogClient = nil
	b.mu.Unlock()
	if cc != nil {
		cc.Close()
	}
}

// ListCatalogGames fetches games from the catalog.
func (b *Backend) ListCatalogGames(tenantID string, pageSize int32) (*catalogizerv1.ListGamesResponse, error) {
	b.mu.RLock()
	cc := b.catalogClient
	b.mu.RUnlock()
	if cc == nil {
		return nil, fmt.Errorf("catalog not open")
	}
	ctx, cancel := context.WithTimeout(b.ctx, 5*time.Second)
	defer cancel()
	return cc.ListGames(ctx, tenantID, pageSize, "")
}
