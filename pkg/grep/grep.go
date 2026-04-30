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
		{Short: "q", Long: "quiet", Type: common.FlagBool},
		{Short: "s", Long: "no-messages", Type: common.FlagBool},
		{Short: "f", Long: "file", Type: common.FlagValue},
		{Short: "e", Long: "regexp", Type: common.FlagValue},
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

func Run(r io.Reader, filename string, re *regexp.Regexp, fixedPatterns []string, invert, fixed, lineRegexp bool) ([]GrepMatch, error) {
	scanner := bufio.NewScanner(r)
	var matches []GrepMatch
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()

		var matchFound bool
		var substrings []string

		if fixed {
			for _, pat := range fixedPatterns {
				var found bool
				if lineRegexp {
					found = text == pat
				} else {
					found = strings.Contains(text, pat)
				}
				if found {
					matchFound = true
					if !invert {
						substrings = append(substrings, pat)
					}
				}
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

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		return 2
	}

	var pattern string
	var paths []string

	if ePat := flags.Get("e"); ePat != "" {
		pattern = ePat
		paths = flags.Positional
	} else if len(flags.Positional) > 0 {
		pattern = flags.Positional[0]
		paths = flags.Positional[1:]
	}

	quiet := flags.Has("q")
	suppressErrors := flags.Has("s")
	filePattern := flags.Get("f")

	if pattern == "" && filePattern == "" {
		if !suppressErrors {
			fmt.Fprintln(os.Stderr, "grep: missing pattern")
		}
		return 2
	}

	jsonMode := flags.Has("j")
	invert := flags.Has("v")
	ignoreCase := flags.Has("i")
	countMode := flags.Has("c")
	lineNum := flags.Has("n")
	filesWithMatches := flags.Has("l")
	fixed := flags.Has("F")
	wordRegexp := flags.Has("w")
	lineRegexp := flags.Has("x")

	var patterns []string
	if pattern != "" {
		patterns = append(patterns, strings.Split(pattern, "\n")...)
	}
	if filePattern != "" {
		b, err := os.ReadFile(filePattern)
		if err == nil {
			patterns = append(patterns, strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")...)
		} else if !suppressErrors {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			return 2
		}
	}

	var re *regexp.Regexp
	var fixedPatterns []string

	if fixed {
		if ignoreCase || wordRegexp {
			var reParts []string
			for _, p := range patterns {
				escaped := regexp.QuoteMeta(p)
				if wordRegexp {
					escaped = "\\b" + escaped + "\\b"
				}
				reParts = append(reParts, escaped)
			}
			pattern = strings.Join(reParts, "|")
			if ignoreCase {
				pattern = "(?i)" + pattern
			}
			fixed = false
		} else {
			fixedPatterns = patterns
		}
	}

	if !fixed {
		if len(patterns) > 0 && pattern == "" {
			pattern = strings.Join(patterns, "|")
		}
		if wordRegexp && !strings.Contains(pattern, "\\b") {
			pattern = "\\b(" + pattern + ")\\b"
		}
		if lineRegexp {
			pattern = "^(" + pattern + ")$"
		}
		if ignoreCase && !strings.HasPrefix(pattern, "(?i)") {
			pattern = "(?i)" + pattern
		}

		compiled, err := regexp.Compile(pattern)
		if err != nil {
			if !suppressErrors {
				fmt.Fprintf(os.Stderr, "grep: invalid regex: %v\n", err)
			}
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
				if !suppressErrors {
					fmt.Fprintf(os.Stderr, "grep: %s: %v\n", path, err)
				}
				exitCode = 2
				continue
			}
			defer f.Close()
			r = f
			fname = path
		}

		matches, err := Run(r, fname, re, fixedPatterns, invert, fixed, lineRegexp)
		if err != nil {
			if !suppressErrors {
				fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			}
			exitCode = 2
		}

		if len(matches) > 0 {
			if exitCode != 2 {
				exitCode = 0
			}
			if quiet {
				return 0
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
		common.Render("grep", allMatches, true, out, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "grep", Usage: "Print lines matching a pattern", Run: run})
}
