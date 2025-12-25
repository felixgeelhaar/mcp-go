package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// ShutdownConfig configures graceful shutdown behavior.
type ShutdownConfig struct {
	// Timeout is the maximum time to wait for in-flight requests to complete.
	// Default: 30 seconds
	Timeout time.Duration

	// DrainDelay is the time to wait before starting to drain connections.
	// This allows load balancers to remove the server from the pool.
	// Default: 0 (no delay)
	DrainDelay time.Duration

	// OnShutdownStart is called when shutdown begins.
	OnShutdownStart func()

	// OnDrainStart is called when draining begins (after DrainDelay).
	OnDrainStart func()

	// OnShutdownComplete is called when shutdown is complete.
	OnShutdownComplete func(err error)
}

// DefaultShutdownConfig returns sensible defaults for shutdown configuration.
func DefaultShutdownConfig() ShutdownConfig {
	return ShutdownConfig{
		Timeout:    30 * time.Second,
		DrainDelay: 0,
	}
}

// ShutdownManager coordinates graceful shutdown with connection draining.
type ShutdownManager struct {
	config ShutdownConfig

	// State tracking
	draining  atomic.Bool
	inFlight  atomic.Int64
	doneCh    chan struct{}
	closeOnce sync.Once
}

// NewShutdownManager creates a new shutdown manager.
func NewShutdownManager(config ShutdownConfig) *ShutdownManager {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &ShutdownManager{
		config: config,
		doneCh: make(chan struct{}),
	}
}

// IsDraining returns true if the server is draining connections.
func (sm *ShutdownManager) IsDraining() bool {
	return sm.draining.Load()
}

// InFlightRequests returns the number of in-flight requests.
func (sm *ShutdownManager) InFlightRequests() int64 {
	return sm.inFlight.Load()
}

// TrackRequest increments the in-flight request counter.
// Returns false if the server is draining and new requests should be rejected.
func (sm *ShutdownManager) TrackRequest() bool {
	if sm.draining.Load() {
		return false
	}
	sm.inFlight.Add(1)
	return true
}

// CompleteRequest decrements the in-flight request counter.
func (sm *ShutdownManager) CompleteRequest() {
	sm.inFlight.Add(-1)
}

// Shutdown initiates graceful shutdown.
// It returns when all in-flight requests complete or timeout is reached.
func (sm *ShutdownManager) Shutdown(ctx context.Context) error {
	// Notify shutdown start
	if sm.config.OnShutdownStart != nil {
		sm.config.OnShutdownStart()
	}

	// Wait for drain delay if configured
	if sm.config.DrainDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sm.config.DrainDelay):
		}
	}

	// Start draining
	sm.draining.Store(true)
	if sm.config.OnDrainStart != nil {
		sm.config.OnDrainStart()
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, sm.config.Timeout)
	defer cancel()

	// Wait for in-flight requests to complete
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var shutdownErr error
	for {
		select {
		case <-timeoutCtx.Done():
			if sm.inFlight.Load() > 0 {
				shutdownErr = timeoutCtx.Err()
			}
			goto done
		case <-ticker.C:
			if sm.inFlight.Load() == 0 {
				goto done
			}
		}
	}

done:
	sm.closeOnce.Do(func() {
		close(sm.doneCh)
	})

	if sm.config.OnShutdownComplete != nil {
		sm.config.OnShutdownComplete(shutdownErr)
	}

	return shutdownErr
}

// Done returns a channel that is closed when shutdown is complete.
func (sm *ShutdownManager) Done() <-chan struct{} {
	return sm.doneCh
}

// WithShutdownTimeout sets the shutdown timeout for HTTP transport.
func WithShutdownTimeout(d time.Duration) HTTPOption {
	return func(h *HTTP) {
		h.shutdownTimeout = d
	}
}

// WithShutdownDrainDelay sets the drain delay for HTTP transport.
func WithShutdownDrainDelay(d time.Duration) HTTPOption {
	return func(h *HTTP) {
		h.drainDelay = d
	}
}
