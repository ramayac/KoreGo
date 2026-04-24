// Package common provides JSON-RPC 2.0 types for korego utilities.
// See https://www.jsonrpc.org/specification
package common

import "encoding/json"

// RPCRequest represents a JSON-RPC 2.0 request object.
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"` // must be "2.0"
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"` // string | int | null
}

// RPCResponse represents a JSON-RPC 2.0 response object.
// Exactly one of Result or Error must be non-nil (per spec).
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError is the error object inside an RPCResponse.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC 2.0 error codes.
const (
	ErrParse          = -32700 // Invalid JSON received
	ErrInvalidRequest = -32600 // JSON is not a valid request
	ErrMethodNotFound = -32601 // Method does not exist
	ErrInvalidParams  = -32602 // Invalid method parameters
	ErrInternal       = -32603 // Internal JSON-RPC error

	// Custom korego error codes (1000–1999 reserved).
	ErrPermission = 1001
	ErrNotFound   = 1002
	ErrTimeout    = 1003
)

// NewRequest is a convenience constructor for an RPC request.
func NewRequest(id interface{}, method string, params interface{}) (*RPCRequest, error) {
	var raw json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	return &RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  raw,
		ID:      id,
	}, nil
}

// NewResult builds a successful RPCResponse.
func NewResult(id interface{}, result interface{}) *RPCResponse {
	return &RPCResponse{JSONRPC: "2.0", ID: id, Result: result}
}

// NewError builds an error RPCResponse.
func NewErrorResponse(id interface{}, code int, message string) *RPCResponse {
	return &RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
}
