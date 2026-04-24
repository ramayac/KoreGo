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
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// parseList parses a list like "1,3,5" or "1-5" into a map of selected 1-based indices.
func parseList(list string) map[int]bool {
	sel := make(map[int]bool)
	parts := strings.Split(list, ",")
	for _, p := range parts {
		if strings.Contains(p, "-") {
			sp := strings.SplitN(p, "-", 2)
			start, _ := strconv.Atoi(sp[0])
			end, _ := strconv.Atoi(sp[1])
			// naive range
			for i := start; i <= end; i++ {
				sel[i] = true
			}
		} else {
			if v, err := strconv.Atoi(p); err == nil {
				sel[v] = true
			}
		}
	}
	return sel
}

func Run(r io.Reader, fields, delimiter, chars, bytesList string) ([]CutLine, error) {
	scanner := bufio.NewScanner(r)
	var lines []CutLine

	delim := "\t"
	if delimiter != "" {
		delim = delimiter
	}

	var fieldSel map[int]bool
	if fields != "" {
		fieldSel = parseList(fields)
	}
	var charSel map[int]bool
	if chars != "" {
		charSel = parseList(chars)
	}
	var byteSel map[int]bool
	if bytesList != "" {
		byteSel = parseList(bytesList)
	}

	for scanner.Scan() {
		text := scanner.Text()
		var extracted []string

		if fields != "" {
			parts := strings.Split(text, delim)
			// If no delimiter found, POSIX cut by default prints the whole line
			// (unless -s is passed, but we'll ignore -s for simplicity unless required)
			if len(parts) == 1 && !strings.Contains(text, delim) {
				extracted = append(extracted, text)
			} else {
				var selected []string
				for i, p := range parts {
					if fieldSel[i+1] {
						selected = append(selected, p)
					}
				}
				extracted = append(extracted, strings.Join(selected, delim))
			}
		} else if chars != "" {
			runes := []rune(text)
			var sb strings.Builder
			for i, r := range runes {
				if charSel[i+1] {
					sb.WriteRune(r)
				}
			}
			extracted = append(extracted, sb.String())
		} else if bytesList != "" {
			var sb strings.Builder
			for i := 0; i < len(text); i++ {
				if byteSel[i+1] {
					sb.WriteByte(text[i])
				}
			}
			extracted = append(extracted, sb.String())
		}

		lines = append(lines, CutLine{Fields: extracted})
	}

	return lines, scanner.Err()
}

func run(args []string) int {
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
		lines, err := Run(r, fields, delimiter, chars, bytesList)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cut: %v\n", err)
			return 1
		}
		allLines = append(allLines, lines...)
	}

	if jsonMode {
		common.Render("cut", CutResult{Lines: allLines}, true, func() {})
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
