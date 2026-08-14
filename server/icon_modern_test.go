package server

import (
	"encoding/json"
	"reflect"
	"testing"

	"go.klarlabs.de/mcp/protocol"
)

// TestIcon_LegacyMarshal asserts that an Icon constructed the legacy
// (2025-11-25, SEP-973) way still serializes with the original uri/mimeType/size
// fields and does not emit the modern src/sizes/theme keys.
func TestIcon_LegacyMarshal(t *testing.T) {
	icon := Icon{URI: "https://example.com/icon.png", MimeType: "image/png", Size: 48}

	data, err := json.Marshal(icon)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"uri":      "https://example.com/icon.png",
		"mimeType": "image/png",
		"size":     float64(48),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy Icon JSON = %v, want %v", got, want)
	}
}

// TestIcon_ModernMarshal asserts that the modern fields serialize under their
// spec keys and that unset legacy scalars are omitted.
func TestIcon_ModernMarshal(t *testing.T) {
	icon := NewIcon("https://example.com/icon.svg").
		WithMimeType("image/svg+xml").
		WithSizes("48x48 96x96").
		WithTheme("dark")

	data, err := json.Marshal(icon)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"uri":      "", // legacy field retains its non-omitempty tag
		"mimeType": "image/svg+xml",
		"src":      "https://example.com/icon.svg",
		"sizes":    "48x48 96x96",
		"theme":    "dark",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("modern Icon JSON = %v, want %v", got, want)
	}
}

// TestIcon_ModernFieldsOmittedWhenEmpty asserts src/sizes/theme are dropped when
// unset so legacy-only icons never leak empty modern keys.
func TestIcon_ModernFieldsOmittedWhenEmpty(t *testing.T) {
	data, err := json.Marshal(Icon{URI: "data://x"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	for _, key := range []string{"src", "sizes", "theme"} {
		var got map[string]any
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if _, present := got[key]; present {
			t.Fatalf("expected %q to be omitted, got JSON %s", key, data)
		}
	}
}

// TestIcon_Normalize verifies the legacy<->modern field mapping in both
// directions and that already-set fields are never overwritten.
func TestIcon_Normalize(t *testing.T) {
	tests := []struct {
		name string
		in   Icon
		want Icon
	}{
		{
			name: "legacy fills modern",
			in:   Icon{URI: "https://x/i.png", Size: 48},
			want: Icon{URI: "https://x/i.png", Size: 48, Src: "https://x/i.png", Sizes: "48x48"},
		},
		{
			name: "modern fills legacy",
			in:   Icon{Src: "https://x/i.png", Sizes: "64x64"},
			want: Icon{URI: "https://x/i.png", Size: 64, Src: "https://x/i.png", Sizes: "64x64"},
		},
		{
			name: "does not overwrite existing fields",
			in:   Icon{URI: "legacy://a", Src: "modern://b", Size: 16, Sizes: "any"},
			want: Icon{URI: "legacy://a", Src: "modern://b", Size: 16, Sizes: "any"},
		},
		{
			name: "any size descriptor yields no legacy pixel size",
			in:   Icon{Src: "https://x/i.svg", Sizes: "any"},
			want: Icon{URI: "https://x/i.svg", Src: "https://x/i.svg", Sizes: "any"},
		},
		{
			name: "multi-descriptor uses first square size",
			in:   Icon{Src: "https://x/i.png", Sizes: "32x32 64x64"},
			want: Icon{URI: "https://x/i.png", Size: 32, Src: "https://x/i.png", Sizes: "32x32 64x64"},
		},
		{
			name: "empty icon is unchanged",
			in:   Icon{},
			want: Icon{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Normalize(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Normalize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestNewIcon confirms the constructor sets only the modern source.
func TestNewIcon(t *testing.T) {
	got := NewIcon("data://icon")
	want := Icon{Src: "data://icon"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NewIcon() = %+v, want %+v", got, want)
	}
}

func TestIcon_ForProtocol(t *testing.T) {
	icon := Icon{URI: "https://example.com/i.png", MimeType: "image/png", Size: 48}

	legacy, ok := icon.ForProtocol(protocol.MCPVersion).(map[string]any)
	if !ok {
		t.Fatalf("legacy ForProtocol type %T", icon.ForProtocol(protocol.MCPVersion))
	}
	if _, hasURI := legacy["uri"]; !hasURI {
		t.Errorf("legacy missing uri: %#v", legacy)
	}
	if _, hasSrc := legacy["src"]; hasSrc {
		t.Errorf("legacy must not include src: %#v", legacy)
	}

	modern, ok := icon.ForProtocol(protocol.ModernVersion).(map[string]any)
	if !ok {
		t.Fatalf("modern ForProtocol type %T", icon.ForProtocol(protocol.ModernVersion))
	}
	if _, hasSrc := modern["src"]; !hasSrc {
		t.Errorf("modern missing src: %#v", modern)
	}
	if _, hasURI := modern["uri"]; hasURI {
		t.Errorf("modern must not include uri: %#v", modern)
	}
}
