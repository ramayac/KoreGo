package common

import (
	"encoding/json"
	"testing"
)

// FuzzJSONRPC ensures that the JSON unmarshaler doesn't panic on arbitrary inputs
// simulating malicious RPC payloads.
func FuzzJSONRPC(f *testing.F) {
	// Seed corpus
	f.Add([]byte(`{"jsonrpc": "2.0", "method": "test", "id": 1}`))
	f.Add([]byte(`{"jsonrpc": "2.0", "method": "korego.ping", "params": {"path": "/etc/shadow"}}`))
	f.Add([]byte(`[`))
	f.Add([]byte(`{"jsonrpc": "2.0"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var req map[string]interface{}
		// Just ensure it doesn't panic
		json.Unmarshal(data, &req)

		var env JSONEnvelope
		json.Unmarshal(data, &env)
	})
}
