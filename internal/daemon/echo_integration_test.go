package daemon

import (
	"encoding/json"
	"testing"

	// Register echo so dispatch.Lookup("echo") works.
	_ "github.com/ramayac/korego/pkg/echo"
)

func TestProcessRequest_EchoJSONMode(t *testing.T) {
	srv := &Server{sm: NewSessionManager(30)}

	// Simulate what the client sends for Echo(ctx, "hello world")
	params, err := json.Marshal(KoregoParams{Flags: nil, Text: "hello world"})
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	req := Request{
		JSONRPC: "2.0",
		Method:  "korego.echo",
		Params:  json.RawMessage(params),
		ID:      1,
	}

	resp := srv.processRequest(req)
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC error: %s (code %d)", resp.Error.Message, resp.Error.Code)
	}

	// The result should be a map (daemon envelope), not a raw string.
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result to be map[string]interface{}, got %T", resp.Result)
	}

	exitCode, ok := resultMap["exitCode"].(int)
	if !ok {
		t.Fatalf("expected exitCode (int) in result map, got %v", resultMap)
	}
	if exitCode != 0 {
		t.Errorf("expected exitCode 0, got %v", exitCode)
	}

	data, ok := resultMap["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map[string]interface{}, got %T: %v", resultMap["data"], resultMap)
	}

	text, ok := data["text"].(string)
	if !ok {
		t.Fatalf("expected data.text string, got %T: %v", data["text"], data)
	}
	if text != "hello world" {
		t.Errorf("expected 'hello world', got %q", text)
	}
}

func TestProcessRequest_EchoJSONModeWithFlags(t *testing.T) {
	srv := &Server{sm: NewSessionManager(30)}

	// Echo with -n flag: {"flags":["-n"], "text":"hello"}
	params, err := json.Marshal(KoregoParams{Flags: []string{"-n"}, Text: "hello"})
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	req := Request{
		JSONRPC: "2.0",
		Method:  "korego.echo",
		Params:  json.RawMessage(params),
		ID:      1,
	}

	resp := srv.processRequest(req)
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Error != nil {
		t.Fatalf("unexpected RPC error: %s (code %d)", resp.Error.Message, resp.Error.Code)
	}

	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result map, got %T", resp.Result)
	}

	data, ok := resultMap["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data map, got %T: %v", resultMap["data"], resultMap)
	}

	text, ok := data["text"].(string)
	if !ok {
		t.Fatalf("expected data.text string, got %T: %v", data["text"], data)
	}
	if text != "hello" {
		t.Errorf("expected 'hello', got %q", text)
	}
}
