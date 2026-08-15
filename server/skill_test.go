package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkillDir(t *testing.T, root, name, description string, extras map[string]string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	b.WriteString("description: " + description + "\n")
	b.WriteString("---\n\n# " + name + "\n\nBody.\n")
	if err := os.WriteFile(filepath.Join(dir, SkillEntryPoint), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	for rel, body := range extras {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestSkill_FromDirRegistersResourcesAndIndex(t *testing.T) {
	root := t.TempDir()
	dir := writeSkillDir(t, root, "git-workflow", "Follow team Git conventions", map[string]string{
		"references/BRANCHING.md": "# Branching\n",
	})

	srv := New(Info{Name: "s", Version: "1"})
	if err := srv.Skill("git-workflow").FromDir(dir); err != nil {
		t.Fatalf("FromDir: %v", err)
	}
	if !srv.HasSkills() {
		t.Fatal("HasSkills = false")
	}

	if _, ok := srv.GetResource("skill://git-workflow/SKILL.md"); !ok {
		t.Fatal("missing SKILL.md resource")
	}
	if _, ok := srv.GetResource("skill://git-workflow/references/BRANCHING.md"); !ok {
		t.Fatal("missing supporting resource")
	}
	if _, ok := srv.GetResource(SkillIndexURI); !ok {
		t.Fatal("missing skill://index.json")
	}

	idx := srv.skillIndexSnapshot()
	if idx.Schema != SkillIndexSchema {
		t.Errorf("schema = %q", idx.Schema)
	}
	if len(idx.Skills) != 1 || idx.Skills[0].Name != "git-workflow" {
		t.Fatalf("index = %#v", idx.Skills)
	}
	if idx.Skills[0].Type != SkillIndexTypeSkillMD {
		t.Errorf("type = %q", idx.Skills[0].Type)
	}
	if idx.Skills[0].URL != "skill://git-workflow/SKILL.md" {
		t.Errorf("url = %q", idx.Skills[0].URL)
	}
}

func TestSkill_NestedPath(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "refunds")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "---\nname: refunds\ndescription: Process refunds\n---\n\n# Refunds\n"
	if err := os.WriteFile(filepath.Join(dir, SkillEntryPoint), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := New(Info{Name: "s", Version: "1"})
	if err := srv.Skill("acme/billing/refunds").FromDir(dir); err != nil {
		t.Fatalf("FromDir: %v", err)
	}
	if _, ok := srv.GetResource("skill://acme/billing/refunds/SKILL.md"); !ok {
		t.Fatal("missing nested skill URI")
	}
	idx := srv.skillIndexSnapshot()
	if idx.Skills[0].Name != "refunds" {
		t.Errorf("name = %q, want refunds", idx.Skills[0].Name)
	}
}

func TestSkill_NameMismatchRejected(t *testing.T) {
	root := t.TempDir()
	dir := writeSkillDir(t, root, "git-workflow", "desc", nil)
	srv := New(Info{Name: "s", Version: "1"})
	err := srv.Skill("other-name").FromDir(dir)
	if err == nil {
		t.Fatal("expected name mismatch error")
	}
}

func TestSkillsFromDir(t *testing.T) {
	root := t.TempDir()
	writeSkillDir(t, root, "alpha", "Alpha skill", nil)
	writeSkillDir(t, root, "beta", "Beta skill", nil)
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("ignore"), 0o644)

	srv := New(Info{Name: "s", Version: "1"})
	if err := srv.SkillsFromDir(root); err != nil {
		t.Fatalf("SkillsFromDir: %v", err)
	}
	idx := srv.skillIndexSnapshot()
	if len(idx.Skills) != 2 {
		t.Fatalf("got %d skills, want 2", len(idx.Skills))
	}
}

func TestParseSkillFrontmatter(t *testing.T) {
	fm, err := parseSkillFrontmatter("---\nname: demo\ndescription: Hello world\n---\n\n# Demo\n")
	if err != nil {
		t.Fatal(err)
	}
	if fm.Name != "demo" || fm.Description != "Hello world" {
		t.Fatalf("%+v", fm)
	}

	fm, err = parseSkillFrontmatter("---\nname: demo\ndescription: |\n  Line one\n  Line two\n---\n")
	if err != nil {
		t.Fatal(err)
	}
	if fm.Description != "Line one\nLine two" {
		t.Fatalf("multiline = %q", fm.Description)
	}
}

func TestSkill_ReadContent(t *testing.T) {
	root := t.TempDir()
	dir := writeSkillDir(t, root, "demo", "Demo skill", nil)
	srv := New(Info{Name: "s", Version: "1"})
	if err := srv.Skill("demo").FromDir(dir); err != nil {
		t.Fatal(err)
	}
	r, ok := srv.GetResource("skill://demo/SKILL.md")
	if !ok {
		t.Fatal("missing resource")
	}
	content, err := r.Read(t.Context(), "skill://demo/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if content.MimeType != SkillMarkdownMIME {
		t.Errorf("mime = %q", content.MimeType)
	}
	if !strings.Contains(content.Text, "name: demo") {
		t.Errorf("text = %q", content.Text)
	}

	idxRes, _ := srv.GetResource(SkillIndexURI)
	idxContent, err := idxRes.Read(t.Context(), SkillIndexURI)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(idxContent.Text, `"$schema"`) || !strings.Contains(idxContent.Text, "demo") {
		t.Errorf("index = %s", idxContent.Text)
	}
}
