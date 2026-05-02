// Core backend entry point for HelixPlay.
// Wires together the API gateway, session service, tenant service,
// auth middleware, catalog proxy, and discovery.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/api"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/auth"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/catalog"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/session"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/tenant"
	memrepo "github.com/HelixDevelopment/HelixPlay/cmd/core/internal/repository"
	"digital.vasic.auth/pkg/jwt"
)

func main() {
	var (
		grpcAddr = flag.String("grpc", ":50051", "gRPC listen address")
		httpAddr = flag.String("http", ":8080", "HTTP listen address")
		jwtSecret = flag.String("jwt-secret", os.Getenv("HELIXPLAY_JWT_SECRET"), "JWT signing secret")
	)
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if *jwtSecret == "" {
		*jwtSecret = "dev-secret-change-in-production"
		slog.Warn("Using default JWT secret — set HELIXPLAY_JWT_SECRET in production")
	}

	slog.Info("HelixPlay Core starting", "grpc", *grpcAddr, "http", *httpAddr)

	// Repositories (in-memory stubs until persistent DB is wired)
	sessionRepo := memrepo.NewMemorySessionRepository()
	tenantRepo := memrepo.NewMemoryTenantRepository()
	hostRepo := memrepo.NewMemoryHostRepository()

	// Discovery
	discoverySvc := capability.NewDiscovery("host-agent", "127.0.0.1")

	// Services
	sessionSvc := session.NewService(sessionRepo, hostRepo, discoverySvc)
	tenantSvc := tenant.NewService(tenantRepo)

	// Auth
	jwtCfg := jwt.DefaultConfig(*jwtSecret)
	jwtMgr := jwt.NewManager(jwtCfg)
	authMiddleware := auth.NewMiddleware(jwtMgr)

	// Catalog proxy (stub — connects to Catalogizer via gRPC when available)
	catalogProxy := catalog.NewProxy(nil) // nil client until catalogizer is reachable

	// Gateway
	gateway := api.NewGateway(api.Config{
		GRPCAddress:    *grpcAddr,
		HTTPAddress:    *httpAddr,
		SessionService: sessionSvc,
		TenantService:  tenantSvc,
		CatalogProxy:   catalogProxy,
		AuthMiddleware: authMiddleware,
		Discovery:      discoverySvc,
	})

	// Health check endpoint
	gateway.HandleHTTP("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	// Protected tenant endpoints
	gateway.HandleHTTPWithAuth("/api/v1/tenants", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tenants":[]}`))
	}), "admin")

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("Shutting down core gracefully")
		cancel()
	}()

	// Start session monitoring loop
	go monitorSessions(ctx, sessionSvc)

	if err := gateway.Serve(ctx); err != nil && err != context.Canceled {
		slog.Error("Gateway error", "error", err)
		os.Exit(1)
	}

	slog.Info("Core stopped")
}

func monitorSessions(ctx context.Context, svc *session.Service) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			terminated, err := svc.MonitorSessions(ctx, 5*time.Minute)
			if err != nil {
				slog.Error("Session monitor error", "error", err)
			} else if terminated > 0 {
				slog.Info("Terminated idle sessions", "count", terminated)
			}
		case <-ctx.Done():
			return
		}
	}
}
