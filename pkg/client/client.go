// Package client provides a Go client for the korego daemon.
package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Client is a JSON-RPC client.
type Client struct {
	socketPath string
	timeout    time.Duration
}

// Dial connects to the socket.
func Dial(socketPath string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Client{socketPath: socketPath, timeout: timeout}
}

// Call executes a method and unmarshals the result.
func (c *Client) Call(method string, params interface{}, result interface{}) error {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(c.timeout))

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	b, _ := json.Marshal(req)
	conn.Write(b)

	dec := json.NewDecoder(conn)
	var res map[string]interface{}
	if err := dec.Decode(&res); err != nil {
		return err
	}

	if res["error"] != nil {
		return fmt.Errorf("RPC Error: %v", res["error"])
	}

	if result != nil && res["result"] != nil {
		// re-marshal result to unmarshal into strongly typed result
		rb, _ := json.Marshal(res["result"])
		return json.Unmarshal(rb, result)
	}
	return nil
}
