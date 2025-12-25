package middleware

import (
	"context"
	"time"

	"github.com/felixgeelhaar/mcp-go/protocol"
)

// Timeout returns middleware that enforces a request deadline.
// If the handler does not complete within the specified duration,
// the context is cancelled and context.DeadlineExceeded is returned.
func Timeout(d time.Duration) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
			ctx, cancel := context.WithTimeout(ctx, d)
			defer cancel()
			return next(ctx, req)
		}
	}
}
