// Package transport provides MCP transport implementations.
//
// This package implements the communication layer for MCP servers,
// supporting multiple transport protocols.
//
// # Stdio Transport
//
// The stdio transport communicates via stdin/stdout, suitable for
// local tools and CLI integrations:
//
//	t := transport.NewStdio()
//	err := t.Serve(ctx, handler)
//
// # HTTP Transport
//
// The HTTP transport provides an HTTP server with Server-Sent Events (SSE)
// support for real-time communication:
//
//	t := transport.NewHTTP(":8080",
//	    transport.WithReadTimeout(30*time.Second),
//	    transport.WithWriteTimeout(30*time.Second),
//	)
//	err := t.Serve(ctx, handler)
//
// The HTTP transport exposes the following endpoints:
//   - POST /mcp - Handle JSON-RPC requests
//   - GET /sse - Establish SSE connection
//   - GET /health - Health check endpoint
//
// # Handler Interface
//
// All transports expect a Handler that processes requests:
//
//	type Handler interface {
//	    HandleRequest(ctx context.Context, req *protocol.Request) (*protocol.Response, error)
//	}
//
// # Usage with mcp Package
//
// Most users should use the mcp package's convenience functions:
//
//	mcp.ServeStdio(ctx, srv)
//	mcp.ServeHTTP(ctx, srv, ":8080")
package transport
