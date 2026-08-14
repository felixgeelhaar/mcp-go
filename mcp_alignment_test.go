package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"go.klarlabs.de/mcp/protocol"
	"go.klarlabs.de/mcp/server"
)

func TestResourcesList_ExcludesTemplates(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	srv.Resource("static://info").
		Name("info").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*ResourceContent, error) {
			return &ResourceContent{URI: uri, Text: "ok"}, nil
		})
	srv.Resource("users://{id}").
		Name("user").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*ResourceContent, error) {
			return &ResourceContent{URI: uri, Text: "u"}, nil
		})

	handler := newRequestHandler(srv)
	resp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodResourcesList, Params: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("resources/list: %v", err)
	}
	list := resp.Result.(map[string]any)["resources"].([]map[string]any)
	if len(list) != 1 {
		t.Fatalf("resources/list len = %d, want 1 (templates excluded), got %#v", len(list), list)
	}
	if list[0][fieldURI] != "static://info" {
		t.Errorf("uri = %v, want static://info", list[0][fieldURI])
	}

	tpl, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`2`), Method: protocol.MethodResourcesTemplatesList, Params: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("templates/list: %v", err)
	}
	templates := tpl.Result.(map[string]any)["resourceTemplates"].([]map[string]any)
	if len(templates) != 1 || templates[0]["uriTemplate"] != "users://{id}" {
		t.Errorf("templates = %#v, want users://{id}", templates)
	}
}

func TestToolsCall_UnknownToolIsInvalidParams(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	handler := newRequestHandler(srv)
	_, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodToolsCall,
		Params: mustParams(t, map[string]any{"name": "nope", "arguments": map[string]any{}}),
	})
	var mcpErr *protocol.Error
	if !errors.As(err, &mcpErr) || mcpErr.Code != protocol.CodeInvalidParams {
		t.Fatalf("unknown tool: got %v, want -32602", err)
	}
}

func TestResourcesRead_BlobOmitsEmptyText(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	srv.Resource("bin://x").
		Name("bin").
		Handler(func(_ context.Context, uri string, _ map[string]string) (*ResourceContent, error) {
			return &ResourceContent{URI: uri, MimeType: "application/octet-stream", Blob: "aGk="}, nil
		})
	handler := newRequestHandler(srv)
	resp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodResourcesRead,
		Params: mustParams(t, map[string]any{"uri": "bin://x"}),
	})
	if err != nil {
		t.Fatalf("resources/read: %v", err)
	}
	contents := resp.Result.(map[string]any)["contents"].([]map[string]any)
	if len(contents) != 1 {
		t.Fatalf("contents len = %d", len(contents))
	}
	if _, hasText := contents[0]["text"]; hasText {
		t.Errorf("blob-only content must not include text, got %#v", contents[0])
	}
	if contents[0]["blob"] != "aGk=" {
		t.Errorf("blob = %v", contents[0]["blob"])
	}
}

func TestToolsList_IconsMatchProtocolVersion(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	type in struct {
		X string `json:"x"`
	}
	srv.Tool("iconic").
		Description("has an icon").
		Icons(Icon{URI: "https://example.com/i.png", MimeType: "image/png", Size: 48}).
		Handler(func(in in) (string, error) { return in.X, nil })
	handler := newRequestHandler(srv)

	legacy, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodToolsList, Params: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("legacy tools/list: %v", err)
	}
	icons := legacy.Result.(map[string]any)["tools"].([]map[string]any)[0]["icons"].([]any)
	legacyIcon := icons[0].(map[string]any)
	if _, ok := legacyIcon["uri"]; !ok {
		t.Errorf("legacy icon missing uri: %#v", legacyIcon)
	}
	if _, ok := legacyIcon["src"]; ok {
		t.Errorf("legacy icon must not include src: %#v", legacyIcon)
	}

	modern, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`2`), Method: protocol.MethodToolsList,
		Params: modernParams(t, protocol.ModernVersion, nil),
	})
	if err != nil {
		t.Fatalf("modern tools/list: %v", err)
	}
	micons := modern.Result.(map[string]any)["tools"].([]map[string]any)[0]["icons"].([]any)
	modernIcon := micons[0].(map[string]any)
	if _, ok := modernIcon["src"]; !ok {
		t.Errorf("modern icon missing src: %#v", modernIcon)
	}
	if _, ok := modernIcon["uri"]; ok {
		t.Errorf("modern icon must not include uri: %#v", modernIcon)
	}
}

func TestList_CursorPagination(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	type in struct{}
	for i := 0; i < defaultListPageSize+1; i++ {
		name := fmt.Sprintf("t%03d", i)
		srv.Tool(name).Handler(func(_ in) (string, error) { return name, nil })
	}
	handler := newRequestHandler(srv)

	first, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodToolsList, Params: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	res := first.Result.(map[string]any)
	tools := res["tools"].([]map[string]any)
	if len(tools) != defaultListPageSize {
		t.Fatalf("first page len = %d, want %d", len(tools), defaultListPageSize)
	}
	next, _ := res["nextCursor"].(string)
	if next == "" {
		t.Fatal("expected nextCursor on a truncated list")
	}

	second, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`2`), Method: protocol.MethodToolsList,
		Params: mustParams(t, map[string]any{"cursor": next}),
	})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	page2 := second.Result.(map[string]any)["tools"].([]map[string]any)
	if len(page2) != 1 {
		t.Fatalf("second page len = %d, want 1", len(page2))
	}
	if _, ok := second.Result.(map[string]any)["nextCursor"]; ok {
		t.Error("last page must not include nextCursor")
	}

	_, err = handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`3`), Method: protocol.MethodToolsList,
		Params: mustParams(t, map[string]any{"cursor": "does-not-exist"}),
	})
	var mcpErr *protocol.Error
	if !errors.As(err, &mcpErr) || mcpErr.Code != protocol.CodeInvalidParams {
		t.Fatalf("invalid cursor: got %v, want -32602", err)
	}
}

func TestCompletion_ContextOnContext(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	var seen server.CompletionContext
	srv.PromptCompletion("greet").Handler(func(ctx context.Context, _ CompletionRef, _ CompletionArgument) (*CompletionResult, error) {
		seen = CompletionContextFromContext(ctx)
		return &CompletionResult{Values: []string{"ok"}}, nil
	})
	handler := newRequestHandler(srv)
	_, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodCompletionComplete,
		Params: mustParams(t, map[string]any{
			"ref":      map[string]any{"type": "ref/prompt", "name": "greet"},
			"argument": map[string]any{"name": "lang", "value": "p"},
			"context":  map[string]any{"arguments": map[string]string{"name": "Ada"}},
		}),
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if seen.Arguments["name"] != "Ada" {
		t.Errorf("completion context = %#v, want name=Ada", seen)
	}
}

func TestElicitationComplete_UnblocksWaiter(t *testing.T) {
	sess := NewSession("s", nil, nil)
	done := make(chan error, 1)
	go func() {
		done <- sess.WaitElicitationComplete(context.Background(), "elic-1")
	}()
	handler := newRequestHandler(NewServer(ServerInfo{Name: "s", Version: "1"}))
	ctx := ContextWithSession(context.Background(), sess)
	_, err := handler.HandleRequest(ctx, &protocol.Request{
		JSONRPC: "2.0", Method: protocol.MethodElicitationComplete,
		Params: mustParams(t, map[string]any{"elicitationId": "elic-1"}),
	})
	if err != nil {
		t.Fatalf("elicitation/complete: %v", err)
	}
	if waitErr := <-done; waitErr != nil {
		t.Fatalf("WaitElicitationComplete: %v", waitErr)
	}
}

func TestInitialize_ModernVersionFallsBackToInitializeEra(t *testing.T) {
	res := initResult(t, NewServer(ServerInfo{Name: "s", Version: "1"}), protocol.ModernVersion)
	if res["protocolVersion"] != protocol.MCPVersion {
		t.Errorf("initialize(%s) = %v, want %s", protocol.ModernVersion, res["protocolVersion"], protocol.MCPVersion)
	}
}
