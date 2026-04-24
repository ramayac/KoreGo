// Package cat implements the POSIX cat utility.
package cat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// CatResult is the --json output for cat (used for small files).
type CatResult struct {
	Lines     []string `json:"lines"`
	LineCount int      `json:"lineCount"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "n", Long: "number", Type: common.FlagBool},
		{Short: "b", Long: "number-nonblank", Type: common.FlagBool},
		{Short: "s", Long: "squeeze-blank", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Run reads from r and writes processed lines to w.
// Returns lines for JSON mode.
func Run(r io.Reader, w io.Writer, numberAll, numberNonBlank, squeezeBlank bool) ([]string, error) {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	prevBlank := false
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		isBlank := strings.TrimSpace(line) == ""

		if squeezeBlank && isBlank && prevBlank {
			continue
		}
		prevBlank = isBlank

		var prefix string
		if numberAll {
			lineNum++
			prefix = fmt.Sprintf("%6d\t", lineNum)
		} else if numberNonBlank && !isBlank {
			lineNum++
			prefix = fmt.Sprintf("%6d\t", lineNum)
		}

		out := prefix + line
		fmt.Fprintln(w, out)
		lines = append(lines, out)
	}
	return lines, scanner.Err()
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cat: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	numberAll := flags.Has("n")
	numberNonBlank := flags.Has("b")
	squeezeBlank := flags.Has("s")

	// Collect readers: files or stdin.
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
			fmt.Fprintf(os.Stderr, "cat: %s: %v\n", path, err)
			return 1
		}
		defer f.Close()
		readers = append(readers, f)
	}

	if jsonMode {
		var allLines []string
		for _, r := range readers {
			lines, err := Run(r, io.Discard, numberAll, numberNonBlank, squeezeBlank)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cat: %v\n", err)
				return 1
			}
			allLines = append(allLines, lines...)
		}
		common.Render("cat", CatResult{Lines: allLines, LineCount: len(allLines)}, true, func() {})
		return 0
	}

	for _, r := range readers {
		if _, err := Run(r, os.Stdout, numberAll, numberNonBlank, squeezeBlank); err != nil {
			fmt.Fprintf(os.Stderr, "cat: %v\n", err)
			return 1
		}
	}
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "cat",
		Usage: "Concatenate and print files",
		Run:   run,
	})
}
