package discovery

import (
	"context"
	"fmt"
	"sync"

	discoveryv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EventType describes the kind of discovery event.
type EventType string

const (
	EventTypeAdded   EventType = "added"
	EventTypeUpdated EventType = "updated"
	EventTypeRemoved EventType = "removed"
)

// Event represents a change in the discovered host set.
type Event struct {
	Type EventType
	Host Host
}

// Host aggregates information about a discovered HelixPlay host.
type Host struct {
	ID             string
	Name           string
	PublicEndpoint string
	LocalEndpoint  string
	OS             string
	Version        string
	Status         string
	ActiveSessions int32
	LastSeenAt     string
	Capabilities   CapabilitiesSnapshot
}

// CapabilitiesSnapshot mirrors the protocol capabilities.
type CapabilitiesSnapshot struct {
	Codecs        []string
	Transports    []string
	MaxResolution Resolution
	MaxRefreshHz  int32
	HDRFormats    []string
	GPUs          []GPUCapability
}

// Resolution represents display resolution.
type Resolution struct {
	Width  int32
	Height int32
}

// GPUCapability represents a GPU advertised by a host.
type GPUCapability struct {
	Index    int32
	Vendor   string
	Model    string
	Encoder  string
	Codecs   []string
	VRAMGB   int32
	ThermalC int32
}

// Client discovers HelixPlay hosts via mDNS and a rendezvous server.
type Client struct {
	rendezvousAddr string
	tenantID       string
	region         string

	mu              sync.RWMutex
	hosts           map[string]Host // keyed by host ID or endpoint
	cancel          context.CancelFunc
	done            chan struct{}
	mdnsHosts       map[string]Host
	rendezvousHosts map[string]Host

	rendezvousConn   *grpc.ClientConn
	rendezvousClient discoveryv1.RendezvousServiceClient

	watchers []chan Event
	watchMu  sync.RWMutex
}

// NewClient creates a discovery client.
func NewClient(rendezvousAddr, tenantID, region string) *Client {
	return &Client{
		rendezvousAddr:  rendezvousAddr,
		tenantID:        tenantID,
		region:          region,
		hosts:           make(map[string]Host),
		mdnsHosts:       make(map[string]Host),
		rendezvousHosts: make(map[string]Host),
	}
}

// Start begins background discovery.
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		return fmt.Errorf("already started")
	}

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})

	if c.rendezvousAddr != "" {
		conn, err := grpc.NewClient(c.rendezvousAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			cancel()
			return fmt.Errorf("rendezvous dial: %w", err)
		}
		c.rendezvousConn = conn
		c.rendezvousClient = discoveryv1.NewRendezvousServiceClient(conn)
	}

	go c.run(ctx)
	return nil
}

// Stop halts discovery.
func (c *Client) Stop() {
	c.mu.Lock()
	cancel := c.cancel
	c.cancel = nil
	if c.rendezvousConn != nil {
		c.rendezvousConn.Close()
		c.rendezvousConn = nil
	}
	c.mu.Unlock()

	if cancel != nil {
		cancel()
		<-c.done
	}
}

// Hosts returns a snapshot of currently discovered hosts.
func (c *Client) Hosts() []Host {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Host, 0, len(c.hosts))
	for _, h := range c.hosts {
		out = append(out, h)
	}
	return out
}

// Watch returns a channel that receives discovery events.
func (c *Client) Watch(ctx context.Context) <-chan Event {
	ch := make(chan Event, 8)
	c.watchMu.Lock()
	c.watchers = append(c.watchers, ch)
	c.watchMu.Unlock()

	go func() {
		<-ctx.Done()
		c.watchMu.Lock()
		for i, w := range c.watchers {
			if w == ch {
				c.watchers = append(c.watchers[:i], c.watchers[i+1:]...)
				break
			}
		}
		c.watchMu.Unlock()
		close(ch)
	}()
	return ch
}

func (c *Client) emit(ev Event) {
	c.watchMu.RLock()
	watchers := make([]chan Event, len(c.watchers))
	copy(watchers, c.watchers)
	c.watchMu.RUnlock()

	for _, w := range watchers {
		select {
		case w <- ev:
		default:
		}
	}
}

func (c *Client) setHost(key string, h Host) {
	c.mu.Lock()
	old, exists := c.hosts[key]
	c.hosts[key] = h
	c.mu.Unlock()

	if !exists {
		c.emit(Event{Type: EventTypeAdded, Host: h})
	} else if hostChanged(old, h) {
		c.emit(Event{Type: EventTypeUpdated, Host: h})
	}
}

func (c *Client) removeHost(key string) {
	c.mu.Lock()
	h, exists := c.hosts[key]
	if exists {
		delete(c.hosts, key)
	}
	c.mu.Unlock()

	if exists {
		c.emit(Event{Type: EventTypeRemoved, Host: h})
	}
}

func hostChanged(a, b Host) bool {
	return a.Status != b.Status || a.ActiveSessions != b.ActiveSessions || a.PublicEndpoint != b.PublicEndpoint
}

func (c *Client) run(ctx context.Context) {
	defer close(c.done)

	mdnsCtx, mdnsCancel := context.WithCancel(ctx)
	defer mdnsCancel()
	go c.runMDNS(mdnsCtx)

	if c.rendezvousClient != nil {
		rendezvousCtx, rendezvousCancel := context.WithCancel(ctx)
		defer rendezvousCancel()
		go c.runRendezvous(rendezvousCtx)
	}

	<-ctx.Done()
}
