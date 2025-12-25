package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/felixgeelhaar/mcp-go/protocol"
)

// Stdio implements MCP transport over stdin/stdout.
type Stdio struct {
	in     io.Reader
	out    io.Writer
	errOut io.Writer

	mu sync.Mutex
}

// StdioOption configures a Stdio transport.
type StdioOption func(*Stdio)

// WithStdin sets a custom stdin reader.
func WithStdin(r io.Reader) StdioOption {
	return func(s *Stdio) {
		s.in = r
	}
}

// WithStdout sets a custom stdout writer.
func WithStdout(w io.Writer) StdioOption {
	return func(s *Stdio) {
		s.out = w
	}
}

// WithStderr sets a custom stderr writer.
func WithStderr(w io.Writer) StdioOption {
	return func(s *Stdio) {
		s.errOut = w
	}
}

// NewStdio creates a new stdio transport.
func NewStdio(opts ...StdioOption) *Stdio {
	s := &Stdio{
		in:     os.Stdin,
		out:    os.Stdout,
		errOut: os.Stderr,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Addr returns the transport address.
func (s *Stdio) Addr() string {
	return "stdio"
}

// Serve starts processing requests from stdin.
func (s *Stdio) Serve(ctx context.Context, handler Handler) error {
	scanner := bufio.NewScanner(s.in)

	// Channel for scanner results
	lines := make(chan string)
	scanErr := make(chan error, 1)

	go func() {
		for scanner.Scan() {
			select {
			case lines <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			scanErr <- err
		}
		close(lines)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-scanErr:
			return err
		case line, ok := <-lines:
			if !ok {
				return nil // EOF
			}
			s.handleLine(ctx, handler, line)
		}
	}
}

func (s *Stdio) handleLine(ctx context.Context, handler Handler, line string) {
	// Parse request
	var req protocol.Request
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		// Send parse error
		resp := protocol.NewErrorResponse(nil, protocol.NewParseError(err.Error()))
		s.writeResponse(resp)
		return
	}

	// Handle request
	resp, err := handler.HandleRequest(ctx, &req)

	// For notifications, don't send response
	if req.IsNotification() {
		return
	}

	// Handle handler errors
	if err != nil {
		if mcpErr, ok := err.(*protocol.Error); ok {
			resp = protocol.NewErrorResponse(req.ID, mcpErr)
		} else {
			resp = protocol.NewErrorResponse(req.ID, protocol.NewInternalError(err.Error()))
		}
	}

	if resp != nil {
		s.writeResponse(resp)
	}
}

func (s *Stdio) writeResponse(resp *protocol.Response) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}

	s.out.Write(data)
	s.out.Write([]byte("\n"))
}
