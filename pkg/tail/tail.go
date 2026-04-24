// Package tail implements the POSIX tail utility.
package tail

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// TailResult is the --json output for tail.
type TailResult struct {
	Lines     []string `json:"lines"`
	LineCount int      `json:"lineCount"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "n", Long: "lines", Type: common.FlagValue},
		{Short: "f", Long: "follow", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Run reads all lines and writes the last `linesCount` lines.
func Run(r io.Reader, w io.Writer, linesCount int) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > linesCount {
			lines = lines[1:]
		}
	}

	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	return lines, scanner.Err()
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tail: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	follow := flags.Has("f")
	linesCount := 10
	if nStr := flags.Get("n"); nStr != "" {
		n, err := strconv.Atoi(nStr)
		if err != nil || n < 0 {
			fmt.Fprintf(os.Stderr, "tail: illegal line count -- %s\n", nStr)
			return 2
		}
		linesCount = n
	}

	var readers []string
	if len(flags.Positional) == 0 {
		readers = append(readers, "-")
	} else {
		readers = flags.Positional
	}

	exitCode := 0
	var allLines []string
	for i, path := range readers {
		if len(readers) > 1 {
			name := path
			if path == "-" {
				name = "standard input"
			}
			header := fmt.Sprintf("==> %s <==", name)
			if !jsonMode {
				if i > 0 {
					fmt.Println()
				}
				fmt.Println(header)
			}
		}

		var f *os.File
		if path == "-" {
			f = os.Stdin
		} else {
			file, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "tail: %s: %v\n", path, err)
				exitCode = 1
				continue
			}
			f = file
		}

		writer := os.Stdout
		if jsonMode {
			writer = os.NewFile(uintptr(os.Stderr.Fd()), "/dev/null")
		}
		
		var w io.Writer = writer
		if jsonMode {
			w = io.Discard
		}

		lines, err := Run(f, w, linesCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tail: %v\n", err)
			exitCode = 1
		}
		if jsonMode {
			allLines = append(allLines, lines...)
		}

		if follow && path != "-" {
			// Naive polling follow for non-stdin
			for {
				time.Sleep(500 * time.Millisecond)
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					fmt.Fprintln(writer, scanner.Text())
				}
			}
		}

		if path != "-" {
			f.Close()
		}
	}

	if jsonMode {
		common.Render("tail", TailResult{Lines: allLines, LineCount: len(allLines)}, true, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "tail", Usage: "Output the last part of files", Run: run})
}
