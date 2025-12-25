package protocol

// MCP protocol version.
const MCPVersion = "2024-11-05"

// MCP method names.
const (
	MethodInitialize    = "initialize"
	MethodInitialized   = "notifications/initialized"
	MethodToolsList     = "tools/list"
	MethodToolsCall     = "tools/call"
	MethodResourcesList = "resources/list"
	MethodResourcesRead = "resources/read"
	MethodPromptsList   = "prompts/list"
	MethodPromptsGet    = "prompts/get"
	MethodPing          = "ping"
)

// MCP notification methods.
const (
	MethodProgress = "notifications/progress"
)
