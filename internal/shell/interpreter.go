package shell

import (
	"bytes"
	"context"
	"strings"
	"time"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode uint8  `json:"exitCode"`
}

func Exec(script string, cwd string, env map[string]string) ExecResult {
	var stdout, stderr bytes.Buffer

	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(script), "")
	if err != nil {
		return ExecResult{Stderr: err.Error(), ExitCode: 127}
	}

	opts := []interp.RunnerOption{
		interp.StdIO(nil, &stdout, &stderr),
	}
	
	if cwd != "" {
		opts = append(opts, interp.Dir(cwd))
	}

	runner, err := interp.New(opts...)
	if err != nil {
		return ExecResult{Stderr: err.Error(), ExitCode: 127}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = runner.Run(ctx, prog)
	
	exitCode := uint8(0)
	if err != nil {
		if exit, ok := interp.IsExitStatus(err); ok {
			exitCode = exit
		} else {
			exitCode = 1
			stderr.WriteString(err.Error() + "\n")
		}
	}

	return ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}
