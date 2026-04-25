// Package daemon implements the persistent JSON-RPC 2.0 server.
package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/internal/shell"
	"github.com/ramayac/korego/pkg/common"
)

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// Error represents a JSON-RPC 2.0 error object.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WorkerPool provides a bounded concurrency pool.
type WorkerPool struct {
	sem chan struct{}
}

// NewWorkerPool creates a new pool.
func NewWorkerPool(size int) *WorkerPool {
	return &WorkerPool{sem: make(chan struct{}, size)}
}

// Submit queues a function to run.
func (wp *WorkerPool) Submit(ctx context.Context, fn func()) error {
	select {
	case wp.sem <- struct{}{}:
		go func() {
			defer func() { <-wp.sem }()
			fn()
		}()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Server is the daemon server.
type Server struct {
	socketPath string
	listener   net.Listener
	pool       *WorkerPool
	uptime     time.Time
	
	sm *SessionManager

	activeWorkers int32
	totalRequests int64
}

// NewServer creates a new daemon server.
func NewServer(socketPath string, workers int) *Server {
	return &Server{
		socketPath: socketPath,
		pool:       NewWorkerPool(workers),
		uptime:     time.Now(),
		sm:         NewSessionManager(30 * time.Minute),
	}
}

// Start begins listening on the unix socket.
func (s *Server) Start() error {
	// Remove stale socket
	os.Remove(s.socketPath)

	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}
	s.listener = l

	// Set permissions (0660)
	if err := os.Chmod(s.socketPath, 0660); err != nil {
		l.Close()
		return err
	}

	go s.acceptLoop()
	return nil
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	// Wait for workers? We can add waitgroup later if needed.
	os.Remove(s.socketPath)
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// If listener closed, stop accepting
			select {
			case <-time.After(10 * time.Millisecond):
				if s.listener == nil {
					return
				}
			}
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	// Parse incoming JSON stream. It can be a single object or array.
	dec := json.NewDecoder(conn)
	for {
		// Read raw message to check if it's array or object
		var raw json.RawMessage
		err := dec.Decode(&raw)
		if err != nil {
			if err != io.EOF {
				s.writeError(conn, nil, -32700, "Parse error")
			}
			return
		}

		if len(raw) > 0 && raw[0] == '[' {
			// Batch request
			var reqs []Request
			if err := json.Unmarshal(raw, &reqs); err != nil {
				s.writeError(conn, nil, -32600, "Invalid Request")
				continue
			}
			
			if len(reqs) == 0 {
				s.writeError(conn, nil, -32600, "Invalid Request")
				continue
			}

			s.handleBatch(conn, reqs)
		} else {
			// Single request
			var req Request
			if err := json.Unmarshal(raw, &req); err != nil {
				s.writeError(conn, nil, -32600, "Invalid Request")
				continue
			}
			s.handleSingleAsync(conn, req)
		}
	}
}

func (s *Server) handleBatch(conn net.Conn, reqs []Request) {
	var wg sync.WaitGroup
	results := make([]*Response, len(reqs))

	for i, req := range reqs {
		wg.Add(1)
		idx := i
		r := req
		err := s.pool.Submit(context.Background(), func() {
			defer wg.Done()
			res := s.processRequest(r)
			if res != nil {
				results[idx] = res
			}
		})
		if err != nil {
			wg.Done()
			results[idx] = &Response{
				JSONRPC: "2.0",
				ID:      r.ID,
				Error:   &Error{Code: -32000, Message: "Server busy"},
			}
		}
	}
	wg.Wait()

	// Remove nil responses (notifications)
	var final []Response
	for _, res := range results {
		if res != nil {
			final = append(final, *res)
		}
	}

	if len(final) > 0 {
		enc := json.NewEncoder(conn)
		enc.Encode(final)
	}
}

func (s *Server) handleSingleAsync(conn net.Conn, req Request) {
	err := s.pool.Submit(context.Background(), func() {
		res := s.processRequest(req)
		if res != nil {
			enc := json.NewEncoder(conn)
			enc.Encode(res)
		}
	})
	if err != nil && req.ID != nil {
		s.writeError(conn, req.ID, -32000, "Server busy")
	}
}

func (s *Server) writeError(conn net.Conn, id interface{}, code int, msg string) {
	enc := json.NewEncoder(conn)
	enc.Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &Error{Code: code, Message: msg},
	})
}

// KoregoParams is used to parse the standard parameters.
type KoregoParams struct {
	Flags     []string `json:"flags"`
	Path      string   `json:"path"`
	Text      string   `json:"text"`
	SessionId string   `json:"sessionId"`
}

func (s *Server) processRequest(req Request) *Response {
	atomic.AddInt64(&s.totalRequests, 1)
	atomic.AddInt32(&s.activeWorkers, 1)
	defer atomic.AddInt32(&s.activeWorkers, -1)

	if req.JSONRPC != "2.0" {
		if req.ID != nil {
			return &Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32600, Message: "Invalid Request"}}
		}
		return nil
	}

	if req.Method == "korego.ping" {
		if req.ID == nil {
			return nil
		}
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"pong":    true,
				"uptime":  time.Since(s.uptime).String(),
				"version": common.Version,
				"workers": map[string]interface{}{
					"active": atomic.LoadInt32(&s.activeWorkers),
				},
			},
		}
	}

	if req.Method == "korego.session.create" {
		if req.ID == nil { return nil }
		s := s.sm.Create()
		return &Response{JSONRPC: "2.0", ID: req.ID, Result: s}
	}

	if req.Method == "korego.session.setCwd" {
		if req.ID == nil { return nil }
		var p struct {
			SessionId string `json:"sessionId"`
			Path      string `json:"path"`
		}
		if err := json.Unmarshal(req.Params, &p); err == nil {
			if s.sm.SetCwd(p.SessionId, p.Path) {
				return &Response{JSONRPC: "2.0", ID: req.ID, Result: true}
			}
		}
		return &Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32602, Message: "Invalid session"}}
	}

	if req.Method == "korego.shell.exec" {
		if req.ID == nil { return nil }
		var p struct {
			SessionId string `json:"sessionId"`
			Script    string `json:"script"`
		}
		if err := json.Unmarshal(req.Params, &p); err == nil {
			cwd := "/"
			var env map[string]string
			if p.SessionId != "" {
				if session, ok := s.sm.Get(p.SessionId); ok {
					cwd = session.CWD
					env = session.Env
				}
			}
			res := shell.Exec(p.Script, cwd, env)
			return &Response{JSONRPC: "2.0", ID: req.ID, Result: res}
		}
		return &Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32602, Message: "Invalid params"}}
	}

	if !strings.HasPrefix(req.Method, "korego.") {
		if req.ID != nil {
			return &Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32601, Message: "Method not found"}}
		}
		return nil
	}

	cmdName := strings.TrimPrefix(req.Method, "korego.")
	cmd, ok := dispatch.Lookup(cmdName)
	if !ok {
		if req.ID != nil {
			return &Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32601, Message: "Method not found"}}
		}
		return nil
	}

	// Build args
	var args []string
	var session *Session

	if len(req.Params) > 0 {
		var p KoregoParams
		var dynMap map[string]interface{}
		
		if err := json.Unmarshal(req.Params, &p); err == nil {
			if p.SessionId != "" {
				session, _ = s.sm.Get(p.SessionId)
			}
			args = append(args, p.Flags...)
			if cmdName == "echo" && p.Text != "" {
				args = append(args, p.Text)
			}
			if p.Path != "" {
				args = append(args, p.Path)
			}
		} else if err := json.Unmarshal(req.Params, &dynMap); err == nil {
		}
	}

	// Very rudimentary CWD injection if session is present
	// We prepend it if it's the `path` param and it's relative
	if session != nil && session.CWD != "" && session.CWD != "/" {
		for i, arg := range args {
			if !strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "/") {
				// Naive path resolution
				args[i] = session.CWD + "/" + arg
			}
		}
	}
	
	args = append(args, "--json") // Force JSON mode

	var buf bytes.Buffer
	
	// Execute the command
	cmd.Run(args, &buf)

	// We intercept the output which should be a JSONEnvelope.
	// But `buf` might contain multiple lines or other things if the utility misbehaves.
	// We only want the JSON output to embed in our result.
	var env common.JSONEnvelope
	
	// Try parsing it
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		// It might be raw ndjson or something else.
		// For simplicity, just return it as data if it's valid JSON, else string.
		if req.ID != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  buf.String(),
			}
		}
		return nil
	}

	if req.ID == nil {
		return nil
	}

	// If there's an envelope error, map it
	if env.Error != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32000,
				Message: env.Error.Message,
				Data:    env.Error.Code,
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  env.Data, // Extract just the data
	}
}

// RunDaemon sets up signal handling and runs until terminated.
func RunDaemon(socketPath string, workers int) error {
	server := NewServer(socketPath, workers)
	if err := server.Start(); err != nil {
		return err
	}

	// PID file
	pidFile := "/var/run/korego.pid"
	if socketPath != "/var/run/korego.sock" {
		pidFile = socketPath + ".pid" // mostly for testing in /tmp
	}
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c // wait for signal

	// Graceful shutdown
	server.Stop()
	os.Remove(pidFile)
	return nil
}
