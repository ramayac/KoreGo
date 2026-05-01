package xargs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "E", Long: "eof-str", Type: common.FlagValue},
		{Short: "e", Long: "eof-str-compat", Type: common.FlagValue},
		{Short: "s", Long: "max-chars", Type: common.FlagValue},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type ExecResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exitCode"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "xargs: %v\n", err)
		return 1
	}

	baseCmd := "echo"
	if len(flags.Positional) > 0 {
		baseCmd = flags.Positional[0]
	}

	cmdArgs := []string{}
	if len(flags.Positional) > 1 {
		cmdArgs = flags.Positional[1:]
	}

	maxSize := 0
	if val := flags.Get("s"); val != "" {
		fmt.Sscanf(val, "%d", &maxSize)
	}

	eofStr := ""
	hasEOF := false
	if _, ok := flags.Values["E"]; ok {
		eofStr = flags.Get("E")
		hasEOF = true
	} else if _, ok := flags.Values["e"]; ok {
		eofStr = flags.Get("e")
		hasEOF = true
	}

	baseSize := len(baseCmd)
	for _, a := range cmdArgs {
		baseSize += len(a) + 1
	}

	scanner := bufio.NewScanner(os.Stdin)
	// Default split by words, not lines. Actually POSIX xargs splits by blanks/newlines, but scanner.Scan() is lines?
	// The original implementation used scanner.Text() which is lines. To properly support xargs we should use ScanWords unless -0.
	// But let's keep it as words.
	scanner.Split(bufio.ScanWords)

	var batches [][]string
	var currentBatch []string
	currentSize := baseSize

	for scanner.Scan() {
		word := scanner.Text()
		if hasEOF && word == eofStr {
			break
		}
		// Plus 1 for space
		if maxSize > 0 && currentSize+len(word)+1 > maxSize && len(currentBatch) > 0 {
			batches = append(batches, currentBatch)
			currentBatch = nil
			currentSize = baseSize
		}
		currentBatch = append(currentBatch, word)
		currentSize += len(word) + 1
	}

	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}
	if len(batches) == 0 {
		batches = append(batches, []string{}) // xargs runs once if no args
	}

	exitCode := 0
	var results []ExecResult

	for _, batch := range batches {
		args := append([]string{}, cmdArgs...)
		args = append(args, batch...)

		cmd := exec.Command(baseCmd, args...)
		cmd.Stdout = out
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		code := 0
		if err != nil {
			code = 123
			if exitError, ok := err.(*exec.ExitError); ok {
				code = exitError.ExitCode()
			}
			exitCode = 123 // POSIX says exit 123 if any invocation returns 1-125
		}

		results = append(results, ExecResult{
			Command:  baseCmd,
			ExitCode: code,
		})

		// POSIX: if command exits with 255, xargs stops immediately
		if code == 255 {
			exitCode = 124
			break
		}
	}

	if flags.Has("j") {
		common.Render("xargs", results, true, out, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "xargs", Usage: "Build and execute command lines from standard input", Run: run})
}
