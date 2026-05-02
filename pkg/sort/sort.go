// Package sort implements the POSIX sort utility.
package sort

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
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
		{Short: "o", Long: "output", Type: common.FlagValue},
		{Short: "s", Long: "stable", Type: common.FlagBool},
		{Short: "z", Long: "zero-terminated", Type: common.FlagBool},
		{Short: "h", Long: "human-numeric-sort", Type: common.FlagBool},
		{Short: "M", Long: "month-sort", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type keySpec struct {
	startField int
	startChar  int
	endField   int
	endChar    int
	numeric    bool
	reverse    bool
	month      bool
	human      bool
}

type lineItem struct {
	original string
	keys     []string
	numVals  []float64
	validNum []bool
}

var monthOrder = map[string]int{
	"jan": 1, "january": 1, "feb": 2, "february": 2,
	"mar": 3, "march": 3, "apr": 4, "april": 4,
	"may": 5, "jun": 6, "june": 6,
	"jul": 7, "july": 7, "aug": 8, "august": 8,
	"sep": 9, "september": 9, "oct": 10, "october": 10,
	"nov": 11, "november": 11, "dec": 12, "december": 12,
}

func parseHuman(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	suffix := s[len(s)-1]
	mult := 1.0
	switch suffix {
	case 'K', 'k':
		mult = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		mult = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		mult = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	case 'T', 't':
		mult = 1024 * 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		// Try parsing without stripping suffix (numeric suffix).
		return 0, false
	}
	return v * mult, true
}

func parseMonth(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	lower := strings.ToLower(s)
	if m, ok := monthOrder[lower]; ok {
		return m, true
	}
	if len(lower) >= 3 {
		if m, ok := monthOrder[lower[:3]]; ok {
			return m, true
		}
	}
	return 0, false
}

func extractKey(line string, ks keySpec, delimiter string) string {
	var parts []string
	if delimiter != "" {
		parts = strings.Split(line, delimiter)
	} else {
		parts = strings.Fields(line)
	}
	if ks.startField <= 0 {
		return line
	}
	var keyParts []string
	end := ks.endField
	if end <= 0 || end < ks.startField {
		end = ks.startField
	}
	for f := ks.startField; f <= end && f <= len(parts); f++ {
		keyParts = append(keyParts, parts[f-1])
	}
	key := strings.Join(keyParts, delimiter)
	if delimiter == "" {
		key = strings.Join(keyParts, " ")
	}
	if ks.startChar > 0 && len(key) > 0 {
		runes := []rune(key)
		start := ks.startChar - 1
		if start >= len(runes) {
			return ""
		}
		endChar := len(runes)
		if ks.endChar > 0 && ks.endChar <= len(runes) {
			endChar = ks.endChar
		}
		key = string(runes[start:endChar])
	}
	return key
}

func parseKeySpec(kStr string, globalNumeric, globalReverse, globalMonth, globalHuman bool) []keySpec {
	var specs []keySpec
	ks := keySpec{}
	rest := kStr
	// start field[.char]
	if idx := strings.IndexAny(rest, "nrMhb"); idx >= 0 {
		ks.startField, ks.startChar = parseFieldChar(rest[:idx])
		rest = rest[idx:]
	} else if idx := strings.IndexByte(rest, ','); idx >= 0 {
		ks.startField, ks.startChar = parseFieldChar(rest[:idx])
		rest = rest[idx+1:]
		// end field
		if idx2 := strings.IndexAny(rest, "nrMhb"); idx2 >= 0 {
			ks.endField, ks.endChar = parseFieldChar(rest[:idx2])
			rest = rest[idx2:]
		} else {
			ks.endField, ks.endChar = parseFieldChar(rest)
			rest = ""
		}
	} else {
		ks.startField, ks.startChar = parseFieldChar(rest)
		rest = ""
	}
	for _, ch := range rest {
		switch ch {
		case 'n':
			ks.numeric = true
		case 'r':
			ks.reverse = true
		case 'M':
			ks.month = true
		case 'h':
			ks.human = true
		}
	}
	if globalNumeric && !ks.numeric && !ks.month && !ks.human {
		ks.numeric = true
	}
	if globalReverse {
		ks.reverse = !ks.reverse
	}
	if globalMonth {
		ks.month = true
	}
	if globalHuman {
		ks.human = true
	}
	specs = append(specs, ks)
	return specs
}

func parseFieldChar(s string) (field, char int) {
	field = 1
	if s == "" {
		return
	}
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		if f, err := strconv.Atoi(s[:dot]); err == nil {
			field = f
		}
		if c, err := strconv.Atoi(s[dot+1:]); err == nil {
			char = c
		}
	} else {
		if f, err := strconv.Atoi(s); err == nil {
			field = f
		}
	}
	return
}

func parseLines(r io.Reader, keySpecs []keySpec, delimiter string, zeroTerminated bool) ([]lineItem, error) {
	var items []lineItem
	scanner := bufio.NewScanner(r)
	if zeroTerminated {
		scanner.Split(scanNUL)
	}
	for scanner.Scan() {
		text := scanner.Text()
		item := lineItem{original: text}
		if len(keySpecs) == 0 {
			item.keys = []string{text}
			item.numVals = make([]float64, 1)
			item.validNum = make([]bool, 1)
		} else {
			item.keys = make([]string, len(keySpecs))
			item.numVals = make([]float64, len(keySpecs))
			item.validNum = make([]bool, len(keySpecs))
			for i, ks := range keySpecs {
				key := extractKey(text, ks, delimiter)
				item.keys[i] = key
				if ks.numeric {
					if v, err := strconv.ParseFloat(strings.TrimSpace(key), 64); err == nil {
						item.numVals[i] = v
						item.validNum[i] = true
					}
				} else if ks.human {
					if v, ok := parseHuman(key); ok {
						item.numVals[i] = v
						item.validNum[i] = true
					}
				} else if ks.month {
					if v, ok := parseMonth(key); ok {
						item.numVals[i] = float64(v)
						item.validNum[i] = true
					}
				}
			}
		}
		items = append(items, item)
	}
	return items, scanner.Err()
}

func scanNUL(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func Run(items []lineItem, keySpecs []keySpec, reverse, numeric, unique, month, human bool) []string {
	slices.SortStableFunc(items, func(a, b lineItem) int {
		ksList := keySpecs
		if len(ksList) == 0 {
			ksList = []keySpec{{numeric: numeric, reverse: reverse, month: month, human: human}}
		}
		for ki := range ksList {
			if ki >= len(a.keys) || ki >= len(b.keys) {
				break
			}
			ks := ksList[ki]
			ak := a.keys[ki]
			bk := b.keys[ki]
			var less bool
			tie := false
			if (ks.human || ks.month || ks.numeric) && a.validNum[ki] && b.validNum[ki] {
				less = a.numVals[ki] < b.numVals[ki]
				tie = a.numVals[ki] == b.numVals[ki]
			} else if ks.numeric {
				if a.validNum[ki] && !b.validNum[ki] {
					less = true
				} else if !a.validNum[ki] && b.validNum[ki] {
					less = false
				} else {
					less = ak < bk
					tie = ak == bk
				}
			} else {
				less = ak < bk
				tie = ak == bk
			}
			if tie {
				continue
			}
			if ks.reverse {
				if less {
					return 1
				}
				return -1
			}
			if less {
				return -1
			}
			return 1
		}
		if a.original < b.original {
			if reverse {
				return 1
			}
			return -1
		}
		if a.original > b.original {
			if reverse {
				return -1
			}
			return 1
		}
		return 0
	})

	var result []string
	if unique {
		var last string
		for i, item := range items {
			key := item.original
			if len(item.keys) > 0 {
				key = item.keys[0]
			}
			if i == 0 || key != last {
				result = append(result, item.original)
				last = key
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
	outputFile := flags.Get("o")
	zeroTerm := flags.Has("z")
	human := flags.Has("h")
	month := flags.Has("M")

	var keySpecs []keySpec
	for _, kStr := range flags.GetAll("k") {
		specs := parseKeySpec(kStr, numeric, reverse, month, human)
		keySpecs = append(keySpecs, specs...)
	}
	if len(keySpecs) == 0 && (numeric || month || human) {
		keySpecs = []keySpec{{numeric: numeric, reverse: reverse, month: month, human: human}}
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
		items, err := parseLines(r, keySpecs, delimiter, zeroTerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sort: %v\n", err)
			return 1
		}
		allItems = append(allItems, items...)
	}

	sortedLines := Run(allItems, keySpecs, reverse, numeric, unique, month, human)

	var w io.Writer = out
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sort: %s: %v\n", outputFile, err)
			return 1
		}
		defer f.Close()
		w = f
	}

	if jsonMode {
		common.Render("sort", SortResult{Lines: sortedLines, Count: len(sortedLines)}, true, out, func() {})
	} else {
		for i, line := range sortedLines {
			if i > 0 {
				if zeroTerm {
					w.Write([]byte{0})
				} else {
					fmt.Fprintln(w)
				}
			}
			fmt.Fprint(w, line)
		}
		if !zeroTerm && len(sortedLines) > 0 {
			fmt.Fprintln(w)
		}
	}
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "sort", Usage: "Sort lines of text files", Run: run})
}
