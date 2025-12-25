package protocol

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "simple error message",
			err:  &Error{Code: CodeInternalError, Message: "something went wrong"},
			want: "mcp: something went wrong (code: -32603)",
		},
		{
			name: "parse error",
			err:  &Error{Code: CodeParseError, Message: "invalid JSON"},
			want: "mcp: invalid JSON (code: -32700)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestError_Is(t *testing.T) {
	err1 := NewInternalError("test")
	err2 := NewInternalError("different message")
	err3 := NewInvalidParams("test")

	if !errors.Is(err1, err2) {
		t.Error("errors with same code should match with errors.Is")
	}

	if errors.Is(err1, err3) {
		t.Error("errors with different codes should not match with errors.Is")
	}
}

func TestNewParseError(t *testing.T) {
	err := NewParseError("invalid JSON")

	if err.Code != CodeParseError {
		t.Errorf("Code = %d, want %d", err.Code, CodeParseError)
	}
	if err.Message != "invalid JSON" {
		t.Errorf("Message = %q, want %q", err.Message, "invalid JSON")
	}
}

func TestNewInvalidRequest(t *testing.T) {
	err := NewInvalidRequest("missing method")

	if err.Code != CodeInvalidRequest {
		t.Errorf("Code = %d, want %d", err.Code, CodeInvalidRequest)
	}
}

func TestNewMethodNotFound(t *testing.T) {
	err := NewMethodNotFound("unknown/method")

	if err.Code != CodeMethodNotFound {
		t.Errorf("Code = %d, want %d", err.Code, CodeMethodNotFound)
	}
}

func TestNewInvalidParams(t *testing.T) {
	err := NewInvalidParams("missing required field")

	if err.Code != CodeInvalidParams {
		t.Errorf("Code = %d, want %d", err.Code, CodeInvalidParams)
	}
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("database connection failed")

	if err.Code != CodeInternalError {
		t.Errorf("Code = %d, want %d", err.Code, CodeInternalError)
	}
}

func TestNewNotFound(t *testing.T) {
	err := NewNotFound("tool not found")

	if err.Code != CodeNotFound {
		t.Errorf("Code = %d, want %d", err.Code, CodeNotFound)
	}
}

func TestNewUnauthorized(t *testing.T) {
	err := NewUnauthorized("invalid token")

	if err.Code != CodeUnauthorized {
		t.Errorf("Code = %d, want %d", err.Code, CodeUnauthorized)
	}
}

func TestError_WithData(t *testing.T) {
	data := map[string]string{"field": "query", "reason": "required"}
	err := NewInvalidParams("validation failed").WithData(data)

	if err.Data == nil {
		t.Fatal("Data should not be nil")
	}

	dataMap, ok := err.Data.(map[string]string)
	if !ok {
		t.Fatalf("Data type = %T, want map[string]string", err.Data)
	}

	if dataMap["field"] != "query" {
		t.Errorf("Data[field] = %q, want %q", dataMap["field"], "query")
	}
}
