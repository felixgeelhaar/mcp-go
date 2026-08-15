package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.klarlabs.de/mcp/protocol"
)

func TestSkills_DiscoverAndRead(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "git-workflow")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "---\nname: git-workflow\ndescription: Follow team Git conventions\n---\n\n# Git\n\nUse feature branches.\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte("# Notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := NewServer(ServerInfo{Name: "skills-demo", Version: "1"})
	if err := srv.Skill("git-workflow").FromDir(dir); err != nil {
		t.Fatalf("Skill.FromDir: %v", err)
	}

	res := discover(t, srv)
	caps := res["capabilities"].(map[string]any)
	ext := caps["extensions"].(map[string]any)
	if _, ok := ext[protocol.ExtensionSkills]; !ok {
		t.Fatalf("expected %s in extensions, got %v", protocol.ExtensionSkills, ext)
	}

	// initialize also advertises extensions (SEP-2640).
	handler := newRequestHandler(srv)
	initResp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodInitialize,
		Params: json.RawMessage(`{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"t","version":"1"}}`),
	})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	initCaps := initResp.Result.(map[string]any)["capabilities"].(map[string]any)
	if _, ok := initCaps["extensions"].(map[string]any)[protocol.ExtensionSkills]; !ok {
		t.Fatalf("initialize missing skills extension: %v", initCaps["extensions"])
	}

	listResp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`2`), Method: protocol.MethodResourcesList,
		Params: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("resources/list: %v", err)
	}
	resourcesRaw := listResp.Result.(map[string]any)["resources"]
	uris := map[string]bool{}
	switch resources := resourcesRaw.(type) {
	case []any:
		for _, r := range resources {
			uris[r.(map[string]any)["uri"].(string)] = true
		}
	case []map[string]any:
		for _, r := range resources {
			uris[r["uri"].(string)] = true
		}
	default:
		t.Fatalf("unexpected resources type %T", resourcesRaw)
	}
	for _, want := range []string{
		SkillIndexURI,
		"skill://git-workflow/SKILL.md",
		"skill://git-workflow/notes.md",
	} {
		if !uris[want] {
			t.Errorf("resources/list missing %s (have %v)", want, uris)
		}
	}

	readResp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`3`), Method: protocol.MethodResourcesRead,
		Params: mustJSON(t, map[string]any{"uri": "skill://git-workflow/SKILL.md"}),
	})
	if err != nil {
		t.Fatalf("resources/read: %v", err)
	}
	contentsRaw := readResp.Result.(map[string]any)["contents"]
	var text string
	switch contents := contentsRaw.(type) {
	case []any:
		text = contents[0].(map[string]any)["text"].(string)
	case []map[string]any:
		text = contents[0]["text"].(string)
	default:
		t.Fatalf("unexpected contents type %T", contentsRaw)
	}
	if !strings.Contains(text, "feature branches") {
		t.Errorf("SKILL.md text = %q", text)
	}

	idxResp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`4`), Method: protocol.MethodResourcesRead,
		Params: mustJSON(t, map[string]any{"uri": SkillIndexURI}),
	})
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	idxContents := idxResp.Result.(map[string]any)["contents"]
	var idxText string
	switch contents := idxContents.(type) {
	case []any:
		idxText = contents[0].(map[string]any)["text"].(string)
	case []map[string]any:
		idxText = contents[0]["text"].(string)
	default:
		t.Fatalf("unexpected index contents type %T", idxContents)
	}
	var idx SkillIndex
	if err := json.Unmarshal([]byte(idxText), &idx); err != nil {
		t.Fatal(err)
	}
	if len(idx.Skills) != 1 || idx.Skills[0].Name != "git-workflow" {
		t.Fatalf("index = %#v", idx)
	}
}

func TestSkills_NotAdvertisedWithoutRegistration(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	srv.Tool("t").Description("").Handler(func(_ struct{}) (string, error) { return "ok", nil })
	res := discover(t, srv)
	ext := res["capabilities"].(map[string]any)["extensions"].(map[string]any)
	if _, ok := ext[protocol.ExtensionSkills]; ok {
		t.Fatal("skills extension advertised without skills")
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
