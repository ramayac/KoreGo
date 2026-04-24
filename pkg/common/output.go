// Package common provides the standard JSON output envelope for CoreGoLinux utilities.
package common

import (
	"encoding/json"
	"os"
)

// Version is the CoreGoLinux release version, injected at build time with -ldflags.
var Version = "0.1.0"

// JSONEnvelope is the standard top-level response format for every utility when
// run with --json.  All five keys are always present.
type JSONEnvelope struct {
	Command  string      `json:"command"`
	Version  string      `json:"version"`
	ExitCode int         `json:"exitCode"`
	Data     interface{} `json:"data"`
	Error    *ErrorInfo  `json:"error"`
}

// ErrorInfo carries a POSIX-style error code and a human-readable message.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Render writes output to stdout.  If jsonMode is true it marshals a
// JSONEnvelope; otherwise it calls textFn for the traditional text output.
func Render(cmdName string, data interface{}, jsonMode bool, textFn func()) {
	if jsonMode {
		env := JSONEnvelope{
			Command:  cmdName,
			Version:  Version,
			ExitCode: 0,
			Data:     data,
			Error:    nil,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(env)
	} else {
		textFn()
	}
}

// RenderError writes a JSON error envelope to stdout and returns the exit code.
// If jsonMode is false it is a no-op (caller should print its own error to stderr).
func RenderError(cmdName string, exitCode int, errCode, message string, jsonMode bool) {
	if !jsonMode {
		return
	}
	env := JSONEnvelope{
		Command:  cmdName,
		Version:  Version,
		ExitCode: exitCode,
		Data:     nil,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: message,
		},
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(env)
}
