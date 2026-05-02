// Package transport provides network transport implementations for streaming.
package transport

// Transport is the common interface for all transport backends.
type Transport interface {
	Start() error
	Stop()
}
