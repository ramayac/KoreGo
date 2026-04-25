// Package uniq implements the POSIX uniq utility.
package uniq

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// UniqItem is the output for json format
type UniqItem struct {
	Line  string `json:"line"`
	Count int    `json:"count"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "c", Long: "count", Type: common.FlagBool},
		{Short: "d", Long: "repeated", Type: common.FlagBool},
		{Short: "u", Long: "unique", Type: common.FlagBool},
		{Short: "i", Long: "ignore-case", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

func Run(r io.Reader, countMode, duplicatesOnly, uniqueOnly, ignoreCase bool) ([]UniqItem, error) {
	scanner := bufio.NewScanner(r)
	var items []UniqItem

	var prev string
	var prevOrig string
	count := 0
	first := true

	for scanner.Scan() {
		orig := scanner.Text()
		line := orig
		if ignoreCase {
			line = strings.ToLower(line)
		}

		if first {
			prev = line
			prevOrig = orig
			count = 1
			first = false
			continue
		}

		if line == prev {
			count++
		} else {
			if (!duplicatesOnly && !uniqueOnly) || (duplicatesOnly && count > 1) || (uniqueOnly && count == 1) {
				items = append(items, UniqItem{Line: prevOrig, Count: count})
			}
			prev = line
			prevOrig = orig
			count = 1
		}
	}

	if !first {
		if (!duplicatesOnly && !uniqueOnly) || (duplicatesOnly && count > 1) || (uniqueOnly && count == 1) {
			items = append(items, UniqItem{Line: prevOrig, Count: count})
		}
	}

	return items, scanner.Err()
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	countMode := flags.Has("c")
	duplicatesOnly := flags.Has("d")
	uniqueOnly := flags.Has("u")
	ignoreCase := flags.Has("i")

	var in io.Reader = os.Stdin
	if len(flags.Positional) > 0 && flags.Positional[0] != "-" {
		f, err := os.Open(flags.Positional[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
			return 1
		}
		defer f.Close()
		in = f
	}

	out = os.Stdout
	if len(flags.Positional) > 1 {
		f, err := os.Create(flags.Positional[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
			return 1
		}
		defer f.Close()
		out = f
	}

	items, err := Run(in, countMode, duplicatesOnly, uniqueOnly, ignoreCase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
		return 1
	}

	if jsonMode {
		common.Render("uniq", items, true, out, func() {})
	} else {
		for _, item := range items {
			if countMode {
				fmt.Fprintf(out, "%7d %s\n", item.Count, item.Line)
			} else {
				fmt.Fprintln(out, item.Line)
			}
		}
	}

	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "uniq", Usage: "Report or omit repeated lines", Run: run})
}
