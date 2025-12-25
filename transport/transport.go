// Package transport provides MCP transport implementations.
package transport

import (
	"context"

	"github.com/felixgeelhaar/mcp-go/protocol"
)

// Handler processes incoming MCP requests.
type Handler interface {
	HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error)
}

// HandlerFunc is an adapter to allow ordinary functions as handlers.
type HandlerFunc func(ctx context.Context, req *protocol.Request) (*protocol.Response, error)

// HandleRequest calls f(ctx, req).
func (f HandlerFunc) HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return f(ctx, req)
}

// Transport defines the communication layer interface.
type Transport interface {
	// Serve starts the transport, blocking until ctx is cancelled or an error occurs.
	Serve(ctx context.Context, handler Handler) error

	// Addr returns the transport's address description.
	Addr() string
}
