// Package head implements the POSIX head utility.
package head

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// HeadResult is the --json output for head.
type HeadResult struct {
	Lines     []string `json:"lines"`
	LineCount int      `json:"lineCount"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "n", Long: "lines", Type: common.FlagValue},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Run reads up to linesCount lines from r and writes to w.
// Returns the read lines for JSON mode.
func Run(r io.Reader, w io.Writer, linesCount int) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	count := 0

	for scanner.Scan() {
		if count >= linesCount {
			break
		}
		line := scanner.Text()
		fmt.Fprintln(w, line)
		lines = append(lines, line)
		count++
	}
	return lines, scanner.Err()
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "head: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	linesCount := 10
	if nStr := flags.Get("n"); nStr != "" {
		n, err := strconv.Atoi(nStr)
		if err != nil || n < 0 {
			fmt.Fprintf(os.Stderr, "head: illegal line count -- %s\n", nStr)
			return 2
		}
		linesCount = n
	}

	var readers []io.Reader
	if len(flags.Positional) == 0 || flags.Stdin {
		readers = append(readers, os.Stdin)
	}
	for _, path := range flags.Positional {
		if path == "-" {
			readers = append(readers, os.Stdin)
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "head: %s: %v\n", path, err)
			return 1
		}
		defer f.Close()
		readers = append(readers, f)
	}

	exitCode := 0
	var allLines []string
	for i, r := range readers {
		if len(readers) > 1 {
			name := "standard input"
			if len(flags.Positional) > 0 && flags.Positional[i] != "-" {
				name = flags.Positional[i]
			}
			header := fmt.Sprintf("==> %s <==", name)
			if !jsonMode {
				if i > 0 {
					fmt.Println()
				}
				fmt.Println(header)
			}
		}

		writer := os.Stdout
		if jsonMode {
			writer = os.NewFile(os.Stderr.Fd(), "/dev/null") // redirect to /dev/null temporarily, or rather just discard
		}

		var w io.Writer = writer
		if jsonMode {
			w = io.Discard
		}

		lines, err := Run(r, w, linesCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "head: %v\n", err)
			exitCode = 1
		}
		if jsonMode {
			allLines = append(allLines, lines...)
		}
	}

	if jsonMode {
		common.Render("head", HeadResult{Lines: allLines, LineCount: len(allLines)}, true, out, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "head", Usage: "Output the first part of files", Run: run})
}
