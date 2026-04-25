// Package sort implements the POSIX sort utility.
package sort

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// SortResult is the --json output for sort.
type SortResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "r", Long: "reverse", Type: common.FlagBool},
		{Short: "n", Long: "numeric-sort", Type: common.FlagBool},
		{Short: "u", Long: "unique", Type: common.FlagBool},
		{Short: "k", Long: "key", Type: common.FlagValue},
		{Short: "t", Long: "field-separator", Type: common.FlagValue},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type lineItem struct {
	original string
	key      string
	numKey   float64
	validNum bool
}

func parseLines(r io.Reader, keyField int, delimiter string, numeric bool) ([]lineItem, error) {
	scanner := bufio.NewScanner(r)
	var items []lineItem

	for scanner.Scan() {
		text := scanner.Text()
		item := lineItem{original: text, key: text}

		if keyField > 0 {
			var parts []string
			if delimiter != "" {
				parts = strings.Split(text, delimiter)
			} else {
				parts = strings.Fields(text)
			}

			if keyField <= len(parts) {
				item.key = parts[keyField-1]
			} else {
				item.key = ""
			}
		}

		if numeric {
			// Extract numeric prefix
			numStr := strings.TrimSpace(item.key)
			if numStr != "" {
				if v, err := strconv.ParseFloat(numStr, 64); err == nil {
					item.numKey = v
					item.validNum = true
				}
			}
		}

		items = append(items, item)
	}

	return items, scanner.Err()
}

func Run(items []lineItem, reverse, numeric, unique bool) []string {
	sort.SliceStable(items, func(i, j int) bool {
		less := false
		if numeric {
			if items[i].validNum && items[j].validNum {
				less = items[i].numKey < items[j].numKey
			} else if items[i].validNum {
				less = false // Numbers sort before non-numbers usually, wait POSIX actually treats non-numeric as 0? Let's just compare strings if both aren't valid, or just simple
				// Let's do simple: valid num < non-valid num? Actually posix says 0.
				less = items[i].numKey < 0 // 0 if not valid
			} else if items[j].validNum {
				less = 0 < items[j].numKey
			} else {
				less = items[i].key < items[j].key
			}
		} else {
			less = items[i].key < items[j].key
		}

		if reverse {
			return !less
		}
		return less
	})

	var result []string
	if unique {
		var last string
		for i, item := range items {
			if i == 0 || item.key != last {
				result = append(result, item.original)
				last = item.key
			}
		}
	} else {
		for _, item := range items {
			result = append(result, item.original)
		}
	}

	return result
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort: %v\n", err)
		return 2
	}

	jsonMode := flags.Has("j")
	reverse := flags.Has("r")
	numeric := flags.Has("n")
	unique := flags.Has("u")
	delimiter := flags.Get("t")

	keyField := 0
	if kStr := flags.Get("k"); kStr != "" {
		// Just simplistic parsing, ignores 1,2 complex ranges for now
		parts := strings.Split(kStr, ",")
		if k, err := strconv.Atoi(parts[0]); err == nil && k > 0 {
			keyField = k
		}
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
					fmt.Fprintf(os.Stderr, "sort: %s: %v\n", path, err)
					return 1
				}
				defer f.Close()
				readers = append(readers, f)
			}
		}
	}

	var allItems []lineItem
	for _, r := range readers {
		items, err := parseLines(r, keyField, delimiter, numeric)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sort: %v\n", err)
			return 1
		}
		allItems = append(allItems, items...)
	}

	sortedLines := Run(allItems, reverse, numeric, unique)

	if jsonMode {
		common.Render("sort", SortResult{Lines: sortedLines, Count: len(sortedLines)}, true, out, func() {})
	} else {
		for _, line := range sortedLines {
			fmt.Println(line)
		}
	}

	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "sort", Usage: "Sort lines of text files", Run: run})
}
