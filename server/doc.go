// Package server provides the core MCP server implementation.
//
// This package implements the server-side logic for MCP, including
// tool, resource, and prompt registration. Most users should use
// the higher-level mcp package instead of using this package directly.
//
// # Server
//
// The Server type manages tool, resource, and prompt registrations:
//
//	srv := server.New(server.Info{
//	    Name:    "my-server",
//	    Version: "1.0.0",
//	    Capabilities: server.Capabilities{
//	        Tools:     true,
//	        Resources: true,
//	        Prompts:   true,
//	    },
//	})
//
// # Tools
//
// Tools are registered using the fluent builder API:
//
//	type SearchInput struct {
//	    Query string `json:"query" jsonschema:"required"`
//	}
//
//	srv.Tool("search").
//	    Description("Search for items").
//	    Handler(func(ctx context.Context, input SearchInput) ([]string, error) {
//	        return []string{"result1", "result2"}, nil
//	    })
//
// # Resources
//
// Resources expose data via URI templates:
//
//	srv.Resource("file://{path}").
//	    Name("File").
//	    Description("Read file content").
//	    MimeType("text/plain").
//	    Handler(func(ctx context.Context, uri string, params map[string]string) (*ResourceContent, error) {
//	        return &ResourceContent{URI: uri, Text: "content"}, nil
//	    })
//
// # Prompts
//
// Prompts expose parameterized templates:
//
//	srv.Prompt("greet").
//	    Description("Generate a greeting").
//	    Argument("name", "Name to greet", true).
//	    Handler(func(ctx context.Context, args map[string]string) (*PromptResult, error) {
//	        return &PromptResult{
//	            Messages: []PromptMessage{{Role: "user", Content: TextContent{Type: "text", Text: "Hello, " + args["name"]}}},
//	        }, nil
//	    })
package server
