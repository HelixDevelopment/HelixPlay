// Discovery service entry point for HelixPlay.
// Provides rendezvous registration and lookup for host agents.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// HostRecord represents a registered host agent.
type HostRecord struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Version     string            `json:"version"`
	Codecs      []string          `json:"codecs"`
	Resolution  string            `json:"resolution"`
	LastSeen    time.Time         `json:"last_seen"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Registry holds in-memory host registrations with TTL-based expiration.
type Registry struct {
	mu        sync.RWMutex
	hosts     map[string]*HostRecord
	ttl       time.Duration
	listenAddr string
}

// NewRegistry creates a new host registry.
func NewRegistry(ttl time.Duration, listenAddr string) *Registry {
	return &Registry{
		hosts:      make(map[string]*HostRecord),
		ttl:        ttl,
		listenAddr: listenAddr,
	}
}

// Register adds or updates a host record.
func (r *Registry) Register(h *HostRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	h.LastSeen = time.Now().UTC()
	r.hosts[h.ID] = h
}

// Deregister removes a host record.
func (r *Registry) Deregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hosts, id)
}

// List returns all currently registered hosts.
func (r *Registry) List() []*HostRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*HostRecord
	now := time.Now().UTC()
	for _, h := range r.hosts {
		if now.Sub(h.LastSeen) <= r.ttl {
			out = append(out, h)
		}
	}
	return out
}

// Get returns a single host by ID.
func (r *Registry) Get(id string) (*HostRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hosts[id]
	if !ok {
		return nil, false
	}
	if time.Since(h.LastSeen) > r.ttl {
		return nil, false
	}
	return h, true
}

func main() {
	var (
		addr = flag.String("addr", ":8053", "HTTP listen address")
		ttl  = flag.Duration("ttl", 60*time.Second, "Host record TTL")
	)
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("HelixPlay Discovery starting", "addr", *addr, "ttl", *ttl)

	registry := NewRegistry(*ttl, *addr)

	mux := http.NewServeMux()

	// POST /register — host agents call this to register
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var h HostRecord
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		if h.ID == "" {
			http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
			return
		}
		registry.Register(&h)
		slog.Info("Host registered", "id", h.ID, "name", h.Name, "address", h.Address)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"registered"}`))
	})

	// POST /deregister — host agents call this on shutdown
	mux.HandleFunc("/deregister", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct{ ID string `json:"id"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		registry.Deregister(req.ID)
		slog.Info("Host deregistered", "id", req.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"deregistered"}`))
	})

	// GET /hosts — clients query this to discover available hosts
	mux.HandleFunc("/hosts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		hosts := registry.List()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"hosts": hosts})
	})

	// GET /health — health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	// Start cleanup goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cleanupLoop(ctx, registry, *ttl)

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("Shutting down discovery gracefully")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
		cancel()
	}()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		slog.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	slog.Info("Discovery listening", "addr", listener.Addr().String())
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}

	slog.Info("Discovery stopped")
}

func cleanupLoop(ctx context.Context, registry *Registry, ttl time.Duration) {
	ticker := time.NewTicker(ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// List implicitly filters expired records; nothing else needed for in-memory
			_ = registry.List()
		case <-ctx.Done():
			return
		}
	}
}
