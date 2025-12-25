# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**mcp-go** is a Go framework for building Model Context Protocol (MCP) servers. The goal is to provide Gin-like developer experience for MCP, enabling Go developers to expose tools, resources, and prompts with strong typing, middleware support, and production-ready defaults.

## Build Commands

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test -v ./server/...

# Run specific test
go test -run TestServer_Tool ./server

# Build example
go build ./examples/basic

# Format code
gofmt -w .

# Lint (when golangci-lint is installed)
golangci-lint run
```

## Architecture

```
mcp-go/
├── mcp.go              # Public API facade - main entry point
├── mcp_test.go         # Integration tests for the public API
├── go.mod              # Module definition
│
├── protocol/           # MCP protocol layer (JSON-RPC 2.0)
│   ├── errors.go       # MCP error types and constructors
│   ├── messages.go     # Request/Response types
│   └── constants.go    # Protocol version and method names
│
├── server/             # Core server implementation
│   ├── server.go       # Server aggregate root
│   ├── handler.go      # HandlerFunc and Middleware types
│   └── tool.go         # Tool and ToolBuilder
│
├── schema/             # JSON Schema generation
│   └── schema.go       # Struct to JSON Schema reflection
│
├── transport/          # Transport implementations
│   ├── transport.go    # Transport interface
│   └── stdio.go        # stdio transport for CLI tools
│
└── examples/
    └── basic/          # Basic example server
```

## Key Patterns

### Typed Handlers
Handlers accept typed structs and return typed results:
```go
type SearchInput struct {
    Query string `json:"query" jsonschema:"required"`
}

srv.Tool("search").
    Description("Search for items").
    Handler(func(input SearchInput) ([]Result, error) {
        return results, nil
    })
```

### Context Support
Handlers can optionally receive context:
```go
srv.Tool("fetch").Handler(func(ctx context.Context, input Input) (Result, error) {
    // Use ctx for cancellation, deadlines, etc.
})
```

### Middleware Chain
Gin-style middleware wrapping:
```go
type Middleware func(next HandlerFunc) HandlerFunc
```

## TDD Workflow

Follow Red-Green-Refactor:
1. Write failing test (`test: add failing test for X`)
2. Implement minimal code to pass (`feat: implement X`)
3. Refactor if needed (`refactor: clean up X`)

Use table-driven tests:
```go
func TestX(t *testing.T) {
    tests := []struct {
        name string
        input any
        want any
        wantErr bool
    }{...}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {...})
    }
}
```

## Coverage Targets

| Package | Current | Target |
|---------|---------|--------|
| protocol | 93.8% | 95%+ |
| server | 86.4% | 90%+ |
| schema | 85.4% | 90%+ |
| transport | 86.0% | 85%+ |

## MCP Methods Implemented

- `initialize` - Server initialization handshake
- `tools/list` - List available tools
- `tools/call` - Execute a tool
- `ping` - Health check

## Roadmap (See docs/tdd.md)

**Phase 1 (Complete):**
- [x] Server core with info/capabilities
- [x] Tool registration with builder pattern
- [x] Typed handler validation
- [x] Basic JSON Schema generation
- [x] stdio transport

**Phase 2 (Next):**
- [ ] Resource registration with URI templates
- [ ] Prompt registration
- [ ] HTTP + SSE transport

**Phase 3:**
- [ ] Built-in middleware (recover, requestid, timeout, logging)
- [ ] Middleware chain execution
