// Package grep implements the POSIX grep utility.
package grep

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		{Short: "L", Long: "files-without-match", Type: common.FlagBool},
		{Short: "o", Long: "only-matching", Type: common.FlagBool},
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
		{Short: "a", Long: "text", Type: common.FlagBool},
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
		} else if re != nil {
			if lineRegexp {
				matchFound = re.MatchString(text)
			} else {
				locs := re.FindAllString(text, -1)
				for _, loc := range locs {
					if loc != "" {
						substrings = append(substrings, loc)
					}
				}
				matchFound = len(locs) > 0
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

	var paths []string
	var patterns []string
	ePats := flags.GetAll("e")
	fPats := flags.GetAll("f")

	quiet := flags.Has("q")
	suppressErrors := flags.Has("s")

	if len(ePats) > 0 || len(fPats) > 0 {
		paths = flags.Positional
	} else if len(flags.Positional) > 0 {
		ePats = append(ePats, flags.Positional[0])
		paths = flags.Positional[1:]
	}

	if len(ePats) == 0 && len(fPats) == 0 {
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

	for _, p := range ePats {
		patterns = append(patterns, strings.Split(p, "\n")...)
	}

	for _, fp := range fPats {
		var b []byte
		var err error
		if fp == "-" {
			b, err = io.ReadAll(os.Stdin)
		} else {
			b, err = os.ReadFile(fp)
		}
		
		if err == nil {
			if len(b) > 0 {
				patterns = append(patterns, strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")...)
			}
		} else if !suppressErrors {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			return 2
		}
	}

	var re *regexp.Regexp
	var fixedPatterns []string

	var pattern string
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

	if len(patterns) == 0 {
		// match nothing, re remains nil
	} else if !fixed {
		pattern = strings.Join(patterns, "|")
		if wordRegexp && !strings.Contains(pattern, "\\b") {
			if pattern == "^" || pattern == "$" {
				pattern = `a^`
			} else {
				pattern = "\\b(" + pattern + ")\\b"
			}
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
	var exitCode int = 1
	if len(paths) == 0 {
		readers = append(readers, "-")
	} else if flags.Has("r") {
		for _, root := range paths {
			stat, err := os.Stat(root)
			if err != nil {
				if !suppressErrors {
					fmt.Fprintf(os.Stderr, "grep: %s: %v\n", root, err)
				}
				exitCode = 2
				continue
			}
			if stat.IsDir() {
				walkRoot := root
				if info, err := os.Lstat(root); err == nil && info.Mode()&os.ModeSymlink != 0 {
					if !strings.HasSuffix(walkRoot, "/") {
						walkRoot += "/"
					}
				}
				filepath.Walk(walkRoot, func(p string, info os.FileInfo, err error) error {
					if err != nil { return nil }
					if !info.IsDir() {
						readers = append(readers, p)
					}
					return nil
				})
			} else {
				readers = append(readers, root)
			}
		}
	} else {
		readers = paths
	}

	var allMatches []GrepMatch
	// exitCode is already declared

	printPrefix := len(paths) > 1 || flags.Has("r") || len(readers) > 1

	for _, path := range readers {
		var r io.Reader
		var fname string
		if path == "-" {
			r = os.Stdin
			fname = "(standard input)"
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

		if flags.Has("L") {
			if len(matches) == 0 {
				fmt.Println(fname)
				if exitCode != 2 { exitCode = 0 }
				if quiet { return 0 }
			}
			continue
		}

		if len(matches) > 0 {
			if exitCode != 2 { exitCode = 0 }
			if quiet { return 0 }
		}

		if countMode {
			prefix := ""
			if printPrefix {
				prefix = fname + ":"
			}
			fmt.Printf("%s%d\n", prefix, len(matches))
			continue
		}

		for _, m := range matches {
			prefix := ""
			if printPrefix {
				prefix = fname + ":"
			}
			if lineNum {
				prefix += fmt.Sprintf("%d:", m.Line)
			}
			if flags.Has("o") {
				for _, sub := range m.Matches {
					fmt.Printf("%s%s\n", prefix, sub)
				}
			} else {
				fmt.Printf("%s%s\n", prefix, m.Text)
			}
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
