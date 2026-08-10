package server

import (
	"encoding/json"
	"strings"
	"testing"
)

// A StructuredResult with no structured payload must OMIT structuredContent,
// not emit null.
//
// The MCP spec makes structuredContent optional and requires it to match the
// tool's outputSchema when present. `null` matches no object schema, so a
// strict client rejects the entire result during validation — before the
// agent sees the text content that explains what happened.
//
// The practical damage is worst on error results, which is exactly when the
// text matters: a handler returning IsError with a helpful message had that
// message discarded and replaced by a schema-validation failure, so the
// caller could not tell "the operation failed" from "the response could not
// be encoded". Reported downstream as felixgeelhaar/roady#92, where a
// mutating tool surfaced as malformed and left the caller unable to know
// whether the write had applied.
func TestStructuredResultOmitsAbsentStructuredContent(t *testing.T) {
	tests := []struct {
		name       string
		result     StructuredResult
		wantOmit   bool
		wantSubstr string
	}{
		{
			name: "error result with no structured payload",
			result: StructuredResult{
				IsError: true,
				Content: []Content{{Type: "text", Text: "Invalid status."}},
			},
			wantOmit: true,
		},
		{
			name: "text-only success",
			result: StructuredResult{
				Content: []Content{{Type: "text", Text: "ok"}},
			},
			wantOmit: true,
		},
		{
			// An explicitly empty object is a real payload — a tool whose
			// schema permits {} must still be able to say so.
			name: "empty but non-nil map is preserved",
			result: StructuredResult{
				Content:           []Content{{Type: "text", Text: "ok"}},
				StructuredContent: map[string]any{},
			},
			wantOmit:   false,
			wantSubstr: `"structuredContent":{}`,
		},
		{
			name: "populated payload is preserved",
			result: StructuredResult{
				StructuredContent: map[string]any{"applied": true},
			},
			wantOmit:   false,
			wantSubstr: `"applied":true`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.result)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			got := string(b)

			if strings.Contains(got, `"structuredContent":null`) {
				t.Errorf("emitted explicit null structuredContent: %s", got)
			}
			if tc.wantOmit && strings.Contains(got, "structuredContent") {
				t.Errorf("structuredContent should be omitted entirely: %s", got)
			}
			if tc.wantSubstr != "" && !strings.Contains(got, tc.wantSubstr) {
				t.Errorf("want %q in %s", tc.wantSubstr, got)
			}
		})
	}
}
