package protocol

import (
	"encoding/json"
	"testing"
)

func TestRequest_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Request
		wantErr bool
	}{
		{
			name:  "valid request with params",
			input: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search"}}`,
			want: Request{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Method:  "tools/call",
				Params:  json.RawMessage(`{"name":"search"}`),
			},
		},
		{
			name:  "valid request without params",
			input: `{"jsonrpc":"2.0","id":"abc-123","method":"tools/list"}`,
			want: Request{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`"abc-123"`),
				Method:  "tools/list",
			},
		},
		{
			name:  "notification (no id)",
			input: `{"jsonrpc":"2.0","method":"notifications/cancelled"}`,
			want: Request{
				JSONRPC: "2.0",
				Method:  "notifications/cancelled",
			},
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Request
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.JSONRPC != tt.want.JSONRPC {
				t.Errorf("JSONRPC = %q, want %q", got.JSONRPC, tt.want.JSONRPC)
			}
			if got.Method != tt.want.Method {
				t.Errorf("Method = %q, want %q", got.Method, tt.want.Method)
			}
			if string(got.ID) != string(tt.want.ID) {
				t.Errorf("ID = %s, want %s", got.ID, tt.want.ID)
			}
			if string(got.Params) != string(tt.want.Params) {
				t.Errorf("Params = %s, want %s", got.Params, tt.want.Params)
			}
		})
	}
}

func TestRequest_IsNotification(t *testing.T) {
	tests := []struct {
		name string
		req  Request
		want bool
	}{
		{
			name: "request with id is not notification",
			req:  Request{ID: json.RawMessage(`1`)},
			want: false,
		},
		{
			name: "request without id is notification",
			req:  Request{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.IsNotification(); got != tt.want {
				t.Errorf("IsNotification() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		resp Response
		want string
	}{
		{
			name: "success response",
			resp: Response{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Result:  map[string]string{"status": "ok"},
			},
			want: `{"jsonrpc":"2.0","id":1,"result":{"status":"ok"}}`,
		},
		{
			name: "error response",
			resp: Response{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Error:   &Error{Code: CodeInternalError, Message: "failed"},
			},
			want: `{"jsonrpc":"2.0","id":1,"error":{"code":-32603,"message":"failed"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.resp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare as JSON (normalize whitespace)
			var gotJSON, wantJSON any
			if err := json.Unmarshal(got, &gotJSON); err != nil {
				t.Fatalf("failed to parse got JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.want), &wantJSON); err != nil {
				t.Fatalf("failed to parse want JSON: %v", err)
			}

			gotNorm, _ := json.Marshal(gotJSON)
			wantNorm, _ := json.Marshal(wantJSON)

			if string(gotNorm) != string(wantNorm) {
				t.Errorf("MarshalJSON() = %s, want %s", gotNorm, wantNorm)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	id := json.RawMessage(`42`)
	result := map[string]int{"count": 10}

	resp := NewResponse(id, result)

	if resp.JSONRPC != JSONRPCVersion {
		t.Errorf("JSONRPC = %q, want %q", resp.JSONRPC, JSONRPCVersion)
	}
	if string(resp.ID) != string(id) {
		t.Errorf("ID = %s, want %s", resp.ID, id)
	}
	if resp.Error != nil {
		t.Error("Error should be nil for success response")
	}
}

func TestNewErrorResponse(t *testing.T) {
	id := json.RawMessage(`42`)
	err := NewInternalError("something failed")

	resp := NewErrorResponse(id, err)

	if resp.JSONRPC != JSONRPCVersion {
		t.Errorf("JSONRPC = %q, want %q", resp.JSONRPC, JSONRPCVersion)
	}
	if resp.Result != nil {
		t.Error("Result should be nil for error response")
	}
	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if resp.Error.Code != CodeInternalError {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, CodeInternalError)
	}
}
