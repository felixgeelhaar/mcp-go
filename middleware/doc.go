// Package middleware provides request/response middleware for MCP servers.
//
// Middleware follows the standard pattern where each middleware wraps the
// next handler in the chain, allowing pre- and post-processing of requests.
//
// # Basic Usage
//
// Create and compose middleware:
//
//	chain := middleware.Chain(
//	    middleware.Recover(),
//	    middleware.RequestID(),
//	    middleware.Logging(logger),
//	)
//	handler := chain(baseHandler)
//
// # Available Middleware
//
// The package provides several built-in middleware:
//
//   - Recover: Catches panics and converts them to internal errors
//   - RequestID: Injects unique request IDs into the context
//   - Timeout: Enforces request deadlines
//   - Logging: Logs request details and timing
//
// # Default Stacks
//
// Pre-configured middleware stacks are available for common use cases:
//
//	// Recover + RequestID + Logging
//	stack := middleware.DefaultStack(logger)
//
//	// Recover + RequestID + Timeout + Logging
//	stack := middleware.DefaultStackWithTimeout(logger, 30*time.Second)
//
// # Custom Middleware
//
// Implement custom middleware using the Middleware type:
//
//	func RateLimit(limit int) middleware.Middleware {
//	    return func(next middleware.HandlerFunc) middleware.HandlerFunc {
//	        return func(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
//	            // Rate limiting logic here
//	            return next(ctx, req)
//	        }
//	    }
//	}
package middleware
