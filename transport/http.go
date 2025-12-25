package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/felixgeelhaar/mcp-go/protocol"
)

// HTTP implements an HTTP transport with SSE support for MCP.
type HTTP struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration

	mu         sync.RWMutex
	listenAddr string
	server     *http.Server

	// SSE clients
	sseClients   map[string]chan []byte
	sseClientsMu sync.RWMutex
}

// HTTPOption configures the HTTP transport.
type HTTPOption func(*HTTP)

// WithReadTimeout sets the read timeout for HTTP requests.
func WithReadTimeout(d time.Duration) HTTPOption {
	return func(h *HTTP) {
		h.readTimeout = d
	}
}

// WithWriteTimeout sets the write timeout for HTTP responses.
func WithWriteTimeout(d time.Duration) HTTPOption {
	return func(h *HTTP) {
		h.writeTimeout = d
	}
}

// NewHTTP creates a new HTTP transport.
func NewHTTP(addr string, opts ...HTTPOption) *HTTP {
	h := &HTTP{
		addr:         addr,
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		sseClients:   make(map[string]chan []byte),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Addr returns the configured address.
func (h *HTTP) Addr() string {
	return h.addr
}

// ListenAddr returns the actual address the server is listening on.
func (h *HTTP) ListenAddr() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.listenAddr
}

// Serve starts the HTTP server and handles requests.
func (h *HTTP) Serve(ctx context.Context, handler Handler) error {
	httpHandler := h.createHandler(handler)

	listener, err := net.Listen("tcp", h.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	h.mu.Lock()
	h.listenAddr = listener.Addr().String()
	h.server = &http.Server{
		Handler:      httpHandler,
		ReadTimeout:  h.readTimeout,
		WriteTimeout: h.writeTimeout,
	}
	h.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		if err := h.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// createHandler creates the HTTP handler for MCP requests.
func (h *HTTP) createHandler(handler Handler) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// SSE endpoint for server-to-client messages
	mux.HandleFunc("/mcp/sse", func(w http.ResponseWriter, r *http.Request) {
		h.handleSSE(w, r)
	})

	// Main MCP endpoint
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		h.handleMCP(w, r, handler)
	})

	return mux
}

// handleMCP handles JSON-RPC requests over HTTP.
func (h *HTTP) handleMCP(w http.ResponseWriter, r *http.Request, handler Handler) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req protocol.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := protocol.NewErrorResponse(nil, protocol.NewParseError("Invalid JSON"))
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp, err := handler.HandleRequest(r.Context(), &req)
	if err != nil {
		resp = protocol.NewErrorResponse(req.ID, protocol.NewInternalError(err.Error()))
	}

	if resp != nil {
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// handleSSE handles Server-Sent Events connections.
func (h *HTTP) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	messageCh := make(chan []byte, 10)

	h.sseClientsMu.Lock()
	h.sseClients[clientID] = messageCh
	h.sseClientsMu.Unlock()

	defer func() {
		h.sseClientsMu.Lock()
		delete(h.sseClients, clientID)
		close(messageCh)
		h.sseClientsMu.Unlock()
	}()

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"clientId\":\"%s\"}\n\n", clientID)
	flusher.Flush()

	// Keep connection open and send messages
	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-messageCh:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

// Broadcast sends a message to all connected SSE clients.
func (h *HTTP) Broadcast(data []byte) {
	h.sseClientsMu.RLock()
	defer h.sseClientsMu.RUnlock()

	for _, ch := range h.sseClients {
		select {
		case ch <- data:
		default:
			// Skip if channel is full
		}
	}
}

// SendTo sends a message to a specific SSE client.
func (h *HTTP) SendTo(clientID string, data []byte) bool {
	h.sseClientsMu.RLock()
	defer h.sseClientsMu.RUnlock()

	if ch, ok := h.sseClients[clientID]; ok {
		select {
		case ch <- data:
			return true
		default:
			return false
		}
	}
	return false
}
