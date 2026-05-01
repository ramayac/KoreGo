// Package cut implements the POSIX cut utility.
package cut

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// CutLine represents the fields/characters extracted from a single line.
type CutLine struct {
	Fields []string `json:"fields"`
}

// CutResult is the --json output for cut.
type CutResult struct {
	Lines []CutLine `json:"lines"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "f", Long: "fields", Type: common.FlagValue},
		{Short: "d", Long: "delimiter", Type: common.FlagValue},
		{Short: "c", Long: "characters", Type: common.FlagValue},
		{Short: "b", Long: "bytes", Type: common.FlagValue},
		{Short: "n", Long: "no-split-chars", Type: common.FlagBool},
		{Short: "s", Long: "only-delimited", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// rangeSpec defines an inclusive range.
type rangeSpec struct {
	start, end int
}

// parseList parses a list like "1,3,5" or "1-5" or "1-" or "-5"
func parseList(list string) ([]rangeSpec, error) {
	var specs []rangeSpec
	parts := strings.Split(list, ",")
	for _, p := range parts {
		if strings.Contains(p, "-") {
			sp := strings.SplitN(p, "-", 2)
			start := 1
			if sp[0] != "" {
				v, err := strconv.Atoi(sp[0])
				if err != nil || v < 1 {
					return nil, fmt.Errorf("invalid byte, character or field list")
				}
				start = v
			}
			end := int(^uint(0) >> 1) // math.MaxInt
			if sp[1] != "" {
				v, err := strconv.Atoi(sp[1])
				// POSIX says start <= end is not strictly required to be error, but busybox does
				if err != nil || v < 1 {
					return nil, fmt.Errorf("invalid byte, character or field list")
				}
				end = v
			}
			specs = append(specs, rangeSpec{start, end})
		} else {
			v, err := strconv.Atoi(p)
			if err != nil || v < 1 {
				return nil, fmt.Errorf("invalid byte, character or field list")
			}
			specs = append(specs, rangeSpec{v, v})
		}
	}
	return specs, nil
}

func inRange(idx int, specs []rangeSpec) bool {
	for _, s := range specs {
		if idx >= s.start && idx <= s.end {
			return true
		}
	}
	return false
}

func Run(r io.Reader, fields, delimiter, chars, bytesList string, onlyDelimited bool) ([]CutLine, error) {
	scanner := bufio.NewScanner(r)
	var lines []CutLine

	delim := "\t"
	if delimiter != "" {
		delim = delimiter
	}

	var fieldSel []rangeSpec
	var err error
	if fields != "" {
		if fieldSel, err = parseList(fields); err != nil {
			return nil, err
		}
	}
	var charSel []rangeSpec
	if chars != "" {
		if charSel, err = parseList(chars); err != nil {
			return nil, err
		}
	}
	var byteSel []rangeSpec
	if bytesList != "" {
		if byteSel, err = parseList(bytesList); err != nil {
			return nil, err
		}
	}

	for scanner.Scan() {
		text := scanner.Text()
		var extracted []string

		if fields != "" {
			parts := strings.Split(text, delim)
			// If no delimiter found, POSIX cut by default prints the whole line
			// (unless -s is passed, but we'll ignore -s for simplicity unless required)
			if len(parts) == 1 && !strings.Contains(text, delim) {
				if !onlyDelimited {
					extracted = append(extracted, text)
				} else {
					continue // Suppress line
				}
			} else {
				var selected []string
				for i, p := range parts {
					if inRange(i+1, fieldSel) {
						selected = append(selected, p)
					}
				}
				extracted = append(extracted, strings.Join(selected, delim))
			}
		} else if chars != "" {
			runes := []rune(text)
			var sb strings.Builder
			for i, r := range runes {
				if inRange(i+1, charSel) {
					sb.WriteRune(r)
				}
			}
			extracted = append(extracted, sb.String())
		} else if bytesList != "" {
			var sb strings.Builder
			for i := 0; i < len(text); i++ {
				if inRange(i+1, byteSel) {
					sb.WriteByte(text[i])
				}
			}
			extracted = append(extracted, sb.String())
		}

		lines = append(lines, CutLine{Fields: extracted})
	}

	return lines, scanner.Err()
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cut: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	fields := flags.Get("f")
	delimiter := flags.Get("d")
	chars := flags.Get("c")
	bytesList := flags.Get("b")
	onlyDelimited := flags.Has("s")

	if fields == "" && chars == "" && bytesList == "" {
		fmt.Fprintf(os.Stderr, "cut: you must specify a list of bytes, characters, or fields\n")
		return 1
	}

	var readers []io.Reader
	if len(flags.Positional) == 0 {
		readers = append(readers, os.Stdin)
	} else {
		for _, path := range flags.Positional {
			if path == "-" {
				readers = append(readers, os.Stdin)
			} else {
				f, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "cut: %s: %v\n", path, err)
					return 1
				}
				defer f.Close()
				readers = append(readers, f)
			}
		}
	}

	var allLines []CutLine
	for _, r := range readers {
		lines, err := Run(r, fields, delimiter, chars, bytesList, onlyDelimited)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cut: %v\n", err)
			return 1
		}
		allLines = append(allLines, lines...)
	}

	if jsonMode {
		common.Render("cut", CutResult{Lines: allLines}, true, out, func() {})
	} else {
		for _, line := range allLines {
			if len(line.Fields) > 0 {
				fmt.Println(line.Fields[0])
			}
		}
	}

	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "cut", Usage: "Remove sections from each line of files", Run: run})
}
