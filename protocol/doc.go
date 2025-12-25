// Package protocol defines the MCP JSON-RPC 2.0 message types and error codes.
//
// This package provides the low-level protocol structures used by mcp-go.
// Most users should use the higher-level mcp package instead.
//
// # Request and Response Types
//
// The package defines the core JSON-RPC 2.0 message types:
//
//	type Request struct {
//	    JSONRPC string          `json:"jsonrpc"`
//	    ID      json.RawMessage `json:"id,omitempty"`
//	    Method  string          `json:"method"`
//	    Params  json.RawMessage `json:"params,omitempty"`
//	}
//
//	type Response struct {
//	    JSONRPC string      `json:"jsonrpc"`
//	    ID      json.RawMessage `json:"id,omitempty"`
//	    Result  any         `json:"result,omitempty"`
//	    Error   *Error      `json:"error,omitempty"`
//	}
//
// # Error Codes
//
// Standard JSON-RPC 2.0 error codes are defined as constants:
//
//	CodeParseError     = -32700  // Invalid JSON
//	CodeInvalidRequest = -32600  // Invalid Request object
//	CodeMethodNotFound = -32601  // Method not found
//	CodeInvalidParams  = -32602  // Invalid method parameters
//	CodeInternalError  = -32603  // Internal server error
//
// Helper functions create properly formatted errors:
//
//	err := protocol.NewMethodNotFound("unknown/method")
//	err := protocol.NewInvalidParams("missing required field: name")
//
// # MCP Method Constants
//
// Standard MCP method names are defined as constants:
//
//	MethodInitialize    = "initialize"
//	MethodToolsList     = "tools/list"
//	MethodToolsCall     = "tools/call"
//	MethodResourcesList = "resources/list"
//	MethodResourcesRead = "resources/read"
//	MethodPromptsList   = "prompts/list"
//	MethodPromptsGet    = "prompts/get"
//	MethodPing          = "ping"
package protocol
