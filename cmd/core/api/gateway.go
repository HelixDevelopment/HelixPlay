package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/capability"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/auth"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/catalog"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/session"
	"github.com/HelixDevelopment/HelixPlay/cmd/core/tenant"
	catalogizerv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/catalog"
	"google.golang.org/grpc"
)

// Gateway provides unified gRPC and HTTP routing for core backend services.
type Gateway struct {
	grpcServer     *grpc.Server
	httpMux        *http.ServeMux
	sessionService *session.Service
	tenantService  *tenant.Service
	catalogProxy   *catalog.Proxy
	authMiddleware *auth.Middleware
	discovery      *capability.Discovery
	grpcAddress    string
	httpAddress    string
}

// Config holds gateway initialization parameters.
type Config struct {
	GRPCAddress    string
	HTTPAddress    string
	SessionService *session.Service
	TenantService  *tenant.Service
	CatalogProxy   *catalog.Proxy
	AuthMiddleware *auth.Middleware
	Discovery      *capability.Discovery
}

// NewGateway assembles a gateway from the provided configuration.
func NewGateway(cfg Config) *Gateway {
	return &Gateway{
		grpcServer:     grpc.NewServer(),
		httpMux:        http.NewServeMux(),
		sessionService: cfg.SessionService,
		tenantService:  cfg.TenantService,
		catalogProxy:   cfg.CatalogProxy,
		authMiddleware: cfg.AuthMiddleware,
		discovery:      cfg.Discovery,
		grpcAddress:    cfg.GRPCAddress,
		httpAddress:    cfg.HTTPAddress,
	}
}

// RegisterGRPCService registers a generic gRPC service implementation.
func (g *Gateway) RegisterGRPCService(desc *grpc.ServiceDesc, impl any) {
	g.grpcServer.RegisterService(desc, impl)
}

// RegisterCatalogServer registers the CatalogService gRPC server.
func (g *Gateway) RegisterCatalogServer(srv catalogizerv1.CatalogServiceServer) {
	catalogizerv1.RegisterCatalogServiceServer(g.grpcServer, srv)
}

// HandleHTTP registers an HTTP handler at the given pattern.
func (g *Gateway) HandleHTTP(pattern string, handler http.Handler) {
	g.httpMux.Handle(pattern, handler)
}

// HandleHTTPWithAuth registers an HTTP handler protected by JWT validation.
func (g *Gateway) HandleHTTPWithAuth(pattern string, handler http.Handler, roles ...string) {
	g.httpMux.Handle(pattern, g.authMiddleware.ValidateAndEnforce(roles...)(handler))
}

// Serve starts both the gRPC and HTTP listeners and blocks until the supplied
// context is cancelled.
func (g *Gateway) Serve(ctx context.Context) error {
	grpcListener, err := net.Listen("tcp", g.grpcAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC address %s: %w", g.grpcAddress, err)
	}

	httpListener, err := net.Listen("tcp", g.httpAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on HTTP address %s: %w", g.httpAddress, err)
	}

	errCh := make(chan error, 2)

	go func() {
		if err := g.grpcServer.Serve(grpcListener); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	httpServer := &http.Server{Handler: g.httpMux}
	go func() {
		if err := httpServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	go func() {
		<-ctx.Done()
		g.grpcServer.GracefulStop()
		_ = httpServer.Shutdown(context.Background())
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SessionService returns the registered session service.
func (g *Gateway) SessionService() *session.Service {
	return g.sessionService
}

// TenantService returns the registered tenant service.
func (g *Gateway) TenantService() *tenant.Service {
	return g.tenantService
}

// CatalogProxy returns the registered catalog proxy.
func (g *Gateway) CatalogProxy() *catalog.Proxy {
	return g.catalogProxy
}

// DiscoveryEndpoints returns discovered endpoints filtered by type.
func (g *Gateway) DiscoveryEndpoints(serviceType string) []capability.Endpoint {
	return g.discovery.DiscoverByType(serviceType)
}

// StripPrefix returns an HTTP handler that removes the given prefix before
// delegating to the provided handler.
func StripPrefix(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		if p == r.URL.Path {
			http.NotFound(w, r)
			return
		}
		r.URL.Path = p
		h.ServeHTTP(w, r)
	})
}
