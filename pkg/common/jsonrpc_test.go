package common

import (
	"encoding/json"
	"testing"
)

func TestRPCRequestRoundTrip(t *testing.T) {
	req, err := NewRequest(1, "ls", map[string]string{"path": "/tmp"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got RPCRequest
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.JSONRPC != "2.0" {
		t.Errorf("jsonrpc: got %q, want 2.0", got.JSONRPC)
	}
	if got.Method != "ls" {
		t.Errorf("method: got %q, want ls", got.Method)
	}
}

func TestRPCResponseResultOnly(t *testing.T) {
	resp := NewResult(42, map[string]string{"output": "hello"})
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	// The "error" field must be absent when there is a result.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["error"]; ok {
		t.Error("error field should be absent when result is present")
	}
	if _, ok := raw["result"]; !ok {
		t.Error("result field should be present")
	}
}

func TestRPCResponseErrorOnly(t *testing.T) {
	resp := NewErrorResponse(1, ErrMethodNotFound, "method not found")
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["result"]; ok {
		t.Error("result field should be absent when error is present")
	}
	if _, ok := raw["error"]; !ok {
		t.Error("error field should be present")
	}
}

func TestRPCBatchParsing(t *testing.T) {
	batch := `[
		{"jsonrpc":"2.0","method":"echo","params":{"text":"hi"},"id":1},
		{"jsonrpc":"2.0","method":"ls","id":2}
	]`
	var reqs []RPCRequest
	if err := json.Unmarshal([]byte(batch), &reqs); err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 2 {
		t.Errorf("expected 2 requests, got %d", len(reqs))
	}
}

func TestRPCIDTypes(t *testing.T) {
	tests := []struct {
		raw string
	}{
		{`{"jsonrpc":"2.0","method":"m","id":1}`},
		{`{"jsonrpc":"2.0","method":"m","id":"str"}`},
		{`{"jsonrpc":"2.0","method":"m","id":null}`},
	}
	for _, tt := range tests {
		var req RPCRequest
		if err := json.Unmarshal([]byte(tt.raw), &req); err != nil {
			t.Errorf("unmarshal %s: %v", tt.raw, err)
		}
	}
}

func TestRPCErrorCodes(t *testing.T) {
	// Per JSON-RPC 2.0 spec, reserved error codes are in [-32768, -32000].
	codes := map[string]int{
		"ErrParse":          ErrParse,
		"ErrInvalidRequest": ErrInvalidRequest,
		"ErrMethodNotFound": ErrMethodNotFound,
		"ErrInvalidParams":  ErrInvalidParams,
		"ErrInternal":       ErrInternal,
	}
	for name, code := range codes {
		if code < -32768 || code > -32000 {
			t.Errorf("%s=%d outside standard JSON-RPC reserved range [-32768, -32000]", name, code)
		}
	}
	// Custom codes must be positive (application-defined).
	customCodes := map[string]int{
		"ErrPermission": ErrPermission,
		"ErrNotFound":   ErrNotFound,
		"ErrTimeout":    ErrTimeout,
	}
	for name, code := range customCodes {
		if code <= 0 {
			t.Errorf("custom code %s=%d should be positive", name, code)
		}
	}
}
