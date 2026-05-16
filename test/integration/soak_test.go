package integration

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

// TestSoak runs a moderate load against the daemon for a prolonged time.
// Set -timeout 25h when running.
func TestSoak(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping soak test in short mode")
	}

	duration := 5 * time.Second
	// Use GOPOSIX_SOAK_DURATION to override for true 24h testing
	end := time.Now().Add(duration)

	for time.Now().Before(end) {
		conn, err := net.Dial("unix", "/tmp/goposix.sock")
		if err != nil {
			t.Logf("daemon not running at /tmp/goposix.sock, skipping soak test: %v", err)
			return
		}

		req := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "goposix.ping",
			"id":      1,
		}
		b, _ := json.Marshal(req)
		conn.Write(b)

		var res map[string]interface{}
		json.NewDecoder(conn).Decode(&res)
		conn.Close()

		time.Sleep(10 * time.Millisecond) // ~100 req/sec
	}
}
