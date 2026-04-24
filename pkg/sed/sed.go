// Package sed implements a basic POSIX sed utility.
package sed

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

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "n", Long: "quiet", Type: common.FlagBool},
		{Short: "e", Long: "expression", Type: common.FlagValue},
		{Short: "i", Long: "in-place", Type: common.FlagBool},
	},
}

type substituteCmd struct {
	re      *regexp.Regexp
	repl    string
	global  bool
	print   bool
}

func parseExpr(expr string) (*substituteCmd, error) {
	// extremely basic s/pat/repl/flags
	if !strings.HasPrefix(expr, "s") {
		return nil, fmt.Errorf("only 's' command is supported currently")
	}
	if len(expr) < 4 {
		return nil, fmt.Errorf("invalid s command")
	}
	delim := expr[1:2]
	parts := strings.Split(expr[2:], delim)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid s command syntax")
	}
	pat := parts[0]
	repl := parts[1]
	flags := ""
	if len(parts) > 2 {
		flags = parts[2]
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}

	return &substituteCmd{
		re:     re,
		repl:   repl,
		global: strings.Contains(flags, "g"),
		print:  strings.Contains(flags, "p"),
	}, nil
}

func Run(r io.Reader, w io.Writer, cmd *substituteCmd, suppressDefault bool) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		
		matched := false
		var out string
		if cmd != nil {
			if cmd.re.MatchString(line) {
				matched = true
				if cmd.global {
					out = cmd.re.ReplaceAllString(line, cmd.repl)
				} else {
					// replace only first
					loc := cmd.re.FindStringIndex(line)
					if loc != nil {
						out = line[:loc[0]] + cmd.re.ReplaceAllString(line[loc[0]:loc[1]], cmd.repl) + line[loc[1]:]
					} else {
						out = line
					}
				}
			} else {
				out = line
			}
		} else {
			out = line
		}

		if !suppressDefault {
			fmt.Fprintln(w, out)
		}
		if matched && cmd.print {
			fmt.Fprintln(w, out)
		}
	}
	return scanner.Err()
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sed: %v\n", err)
		return 2
	}
	suppressDefault := flags.Has("n")
	inPlace := flags.Has("i")
	
	var expr string
	if e := flags.Get("e"); e != "" {
		expr = e
	} else if len(flags.Positional) > 0 {
		expr = flags.Positional[0]
		flags.Positional = flags.Positional[1:]
	} else {
		fmt.Fprintln(os.Stderr, "sed: missing command")
		return 1
	}

	cmd, err := parseExpr(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sed: %v\n", err)
		return 1
	}

	var readers []string
	if len(flags.Positional) == 0 {
		readers = append(readers, "-")
	} else {
		readers = flags.Positional
	}

	exitCode := 0
	for _, path := range readers {
		var r io.Reader
		if path == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sed: %s: %v\n", path, err)
				exitCode = 1
				continue
			}
			r = f
		}

		var w io.Writer = os.Stdout
		var tmpFile *os.File
		if inPlace && path != "-" {
			tmpPath := path + ".tmp"
			tmpFile, err = os.Create(tmpPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sed: %v\n", err)
				if rc, ok := r.(io.Closer); ok {
					rc.Close()
				}
				exitCode = 1
				continue
			}
			w = tmpFile
		}

		if err := Run(r, w, cmd, suppressDefault); err != nil {
			fmt.Fprintf(os.Stderr, "sed: %v\n", err)
			exitCode = 1
		}

		if rc, ok := r.(io.Closer); ok {
			rc.Close()
		}

		if inPlace && path != "-" && tmpFile != nil {
			tmpFile.Close()
			os.Rename(tmpFile.Name(), path)
		}
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "sed", Usage: "Stream editor for filtering and transforming text", Run: run})
}
