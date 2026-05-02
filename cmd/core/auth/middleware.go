package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"digital.vasic.auth/pkg/jwt"
)

// contextKey is an unexported type used to avoid key collisions in context values.
type contextKey int

const (
	// userContextKey stores the authenticated user identity in the request context.
	userContextKey contextKey = iota
)

// UserIdentity holds the claims extracted from a validated JWT.
type UserIdentity struct {
	Subject  string
	Roles    []string
	TenantID string
	Email    string
}

// Middleware provides JWT validation and RBAC enforcement.
type Middleware struct {
	jwtManager *jwt.Manager
}

// NewMiddleware creates an auth middleware backed by the given JWT manager.
func NewMiddleware(jwtManager *jwt.Manager) *Middleware {
	return &Middleware{jwtManager: jwtManager}
}

// ValidateAndEnforce returns an HTTP handler that validates the Authorization
// header and enforces that the caller has at least one of the required roles.
func (m *Middleware) ValidateAndEnforce(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := extractBearer(r.Header.Get("Authorization"))
			if !ok {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			token, err := m.jwtManager.Validate(tokenString)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
				return
			}

			identity := identityFromClaims(token.Claims)
			if len(requiredRoles) > 0 && !hasAnyRole(identity.Roles, requiredRoles) {
				http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
				return
			}

			ctx := WithUserIdentity(r.Context(), identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ValidateOnly returns an HTTP handler that only validates the JWT without
// enforcing roles.
func (m *Middleware) ValidateOnly(next http.Handler) http.Handler {
	return m.ValidateAndEnforce()(next)
}

// UserIdentityFromContext retrieves the user identity from the request context.
// Returns nil if no identity was injected.
func UserIdentityFromContext(ctx context.Context) *UserIdentity {
	v, ok := ctx.Value(userContextKey).(*UserIdentity)
	if !ok {
		return nil
	}
	return v
}

// WithUserIdentity injects a user identity into a context.
func WithUserIdentity(ctx context.Context, identity *UserIdentity) context.Context {
	return context.WithValue(ctx, userContextKey, identity)
}

func extractBearer(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func identityFromClaims(claims map[string]interface{}) *UserIdentity {
	id := &UserIdentity{}

	if v, ok := claims["sub"].(string); ok {
		id.Subject = v
	}
	if v, ok := claims["email"].(string); ok {
		id.Email = v
	}
	if v, ok := claims["tenant_id"].(string); ok {
		id.TenantID = v
	}
	if roles, ok := claims["roles"].([]interface{}); ok {
		for _, r := range roles {
			if s, ok := r.(string); ok {
				id.Roles = append(id.Roles, s)
			}
		}
	}
	return id
}

func hasAnyRole(have, want []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, r := range have {
		set[r] = struct{}{}
	}
	for _, r := range want {
		if _, ok := set[r]; ok {
			return true
		}
	}
	return false
}
