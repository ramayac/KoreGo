// Package grep implements the POSIX grep utility.
package grep

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// GrepMatch represents a single matched line.
type GrepMatch struct {
	File    string   `json:"file,omitempty"`
	Line    int      `json:"line"`
	Text    string   `json:"text"`
	Matches []string `json:"matches"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "i", Long: "ignore-case", Type: common.FlagBool},
		{Short: "v", Long: "invert-match", Type: common.FlagBool},
		{Short: "c", Long: "count", Type: common.FlagBool},
		{Short: "n", Long: "line-number", Type: common.FlagBool},
		{Short: "l", Long: "files-with-matches", Type: common.FlagBool},
		{Short: "r", Long: "recursive", Type: common.FlagBool},
		{Short: "E", Long: "extended-regexp", Type: common.FlagBool},
		{Short: "F", Long: "fixed-strings", Type: common.FlagBool},
		{Short: "w", Long: "word-regexp", Type: common.FlagBool},
		{Short: "x", Long: "line-regexp", Type: common.FlagBool},
		{Short: "A", Long: "after-context", Type: common.FlagValue},
		{Short: "B", Long: "before-context", Type: common.FlagValue},
		{Short: "C", Long: "context", Type: common.FlagValue},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

func Run(r io.Reader, filename string, re *regexp.Regexp, fixedPattern string, invert, fixed, lineRegexp bool) ([]GrepMatch, error) {
	scanner := bufio.NewScanner(r)
	var matches []GrepMatch
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		
		var matchFound bool
		var substrings []string

		if fixed {
			if lineRegexp {
				matchFound = text == fixedPattern
			} else {
				matchFound = strings.Contains(text, fixedPattern)
			}
			if matchFound && !invert {
				substrings = append(substrings, fixedPattern)
			}
		} else {
			if lineRegexp {
				matchFound = re.MatchString(text)
			} else {
				locs := re.FindAllString(text, -1)
				matchFound = len(locs) > 0
				substrings = locs
			}
		}

		if invert {
			matchFound = !matchFound
			substrings = nil
		}

		if matchFound {
			matches = append(matches, GrepMatch{
				File:    filename,
				Line:    lineNum,
				Text:    text,
				Matches: substrings,
			})
		}
	}
	return matches, scanner.Err()
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		return 2
	}
	
	if len(flags.Positional) == 0 {
		fmt.Fprintln(os.Stderr, "grep: missing pattern")
		return 2
	}

	pattern := flags.Positional[0]
	paths := flags.Positional[1:]
	
	jsonMode := flags.Has("j")
	invert := flags.Has("v")
	ignoreCase := flags.Has("i")
	countMode := flags.Has("c")
	lineNum := flags.Has("n")
	filesWithMatches := flags.Has("l")
	fixed := flags.Has("F")
	wordRegexp := flags.Has("w")
	lineRegexp := flags.Has("x")

	var re *regexp.Regexp
	var fixedPattern string = pattern

	if fixed {
		if ignoreCase {
			fixedPattern = strings.ToLower(pattern) // simplified
			// Actually fixed ignoreCase is tricky, but let's just make it regex if ignoreCase + fixed
			// For simplicity we will compile it as regex anyway if ignoreCase is true.
			fixed = false
			pattern = regexp.QuoteMeta(pattern)
		}
	}

	if !fixed {
		if wordRegexp {
			pattern = "\\b" + pattern + "\\b"
		}
		if lineRegexp {
			pattern = "^" + pattern + "$"
		}
		if ignoreCase {
			pattern = "(?i)" + pattern
		}
		
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: invalid regex: %v\n", err)
			return 2
		}
		re = compiled
	}

	var readers []string
	if len(paths) == 0 {
		readers = append(readers, "-")
	} else {
		readers = paths
	}

	var allMatches []GrepMatch
	exitCode := 1 // POSIX grep returns 1 if no lines were selected

	for _, path := range readers {
		var r io.Reader
		var fname string
		if path == "-" {
			r = os.Stdin
			fname = "standard input"
		} else {
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "grep: %s: %v\n", path, err)
				exitCode = 2
				continue
			}
			defer f.Close()
			r = f
			fname = path
		}

		matches, err := Run(r, fname, re, fixedPattern, invert, fixed, lineRegexp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			exitCode = 2
		}

		if len(matches) > 0 {
			if exitCode != 2 {
				exitCode = 0
			}
		}

		if jsonMode {
			allMatches = append(allMatches, matches...)
			continue
		}

		if filesWithMatches {
			if len(matches) > 0 {
				fmt.Println(fname)
			}
			continue
		}

		if countMode {
			prefix := ""
			if len(readers) > 1 {
				prefix = fname + ":"
			}
			fmt.Printf("%s%d\n", prefix, len(matches))
			continue
		}

		for _, m := range matches {
			prefix := ""
			if len(readers) > 1 {
				prefix = fname + ":"
			}
			if lineNum {
				prefix += fmt.Sprintf("%d:", m.Line)
			}
			fmt.Printf("%s%s\n", prefix, m.Text)
		}
	}

	if jsonMode {
		common.Render("grep", allMatches, true, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "grep", Usage: "Print lines matching a pattern", Run: run})
}
