package middleware

import "time"

// DefaultStack returns the recommended production middleware stack.
// This includes panic recovery, request ID injection, and logging.
func DefaultStack(logger Logger) []Middleware {
	return []Middleware{
		Recover(),
		RequestID(),
		Logging(logger),
	}
}

// DefaultStackWithTimeout returns the default stack with a timeout middleware.
func DefaultStackWithTimeout(logger Logger, timeout time.Duration) []Middleware {
	return []Middleware{
		Recover(),
		RequestID(),
		Timeout(timeout),
		Logging(logger),
	}
}
