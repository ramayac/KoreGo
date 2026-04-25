// Package wc implements the POSIX wc utility.
package wc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// WcResult is the --json output for a file.
type WcResult struct {
	Lines int `json:"lines"`
	Words int `json:"words"`
	Bytes int `json:"bytes"`
	Chars int `json:"chars"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "l", Long: "lines", Type: common.FlagBool},
		{Short: "w", Long: "words", Type: common.FlagBool},
		{Short: "c", Long: "bytes", Type: common.FlagBool},
		{Short: "m", Long: "chars", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Count reads from r and returns the counts.
func Count(r io.Reader) (WcResult, error) {
	var res WcResult
	buf := make([]byte, 32*1024)
	inWord := false

	for {
		n, err := r.Read(buf)
		if n > 0 {
			res.Bytes += n
			chunk := buf[:n]

			res.Lines += bytes.Count(chunk, []byte{'\n'})

			res.Chars += utf8.RuneCount(chunk) // Simplified, actual needs state between chunks

			for _, r := range string(chunk) {
				if unicode.IsSpace(r) {
					inWord = false
				} else if !inWord {
					inWord = true
					res.Words++
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

// Simple reliable counter using bufio.Scanner for words and lines, but it might be slower.
func CountScanner(r io.Reader) (WcResult, error) {
	var res WcResult
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes) // We process byte by byte for accurate counts

	inWord := false

	for scanner.Scan() {
		b := scanner.Bytes()[0]
		res.Bytes++
		if b == '\n' {
			res.Lines++
		}

		r := rune(b) // This is naïve for chars, assumes ASCII for the word splitting part. Proper implementation uses a better reader.
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			res.Words++
		}
	}

	// Real char count needs proper rune parsing
	// Since we need posix wc, let's use bufio.Reader
	return res, scanner.Err()
}

// Proper count implementation
func CountProper(r io.Reader) (WcResult, error) {
	var res WcResult
	reader := bufio.NewReader(r)
	inWord := false

	for {
		r, size, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return res, err
		}
		res.Bytes += size
		res.Chars++
		if r == '\n' {
			res.Lines++
		}
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			res.Words++
		}
	}
	return res, nil
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")

	showLines := flags.Has("l")
	showWords := flags.Has("w")
	showBytes := flags.Has("c")
	showChars := flags.Has("m")

	// Default POSIX behavior: if no flags are given, print lines, words, bytes
	if !showLines && !showWords && !showBytes && !showChars {
		showLines, showWords, showBytes = true, true, true
	}

	paths := flags.Positional
	if len(paths) == 0 {
		paths = append(paths, "-")
	}

	var total WcResult
	var jsonResults map[string]WcResult
	if jsonMode {
		jsonResults = make(map[string]WcResult)
	}

	exitCode := 0
	for _, p := range paths {
		var f io.Reader
		if p == "-" {
			f = os.Stdin
		} else {
			file, err := os.Open(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "wc: %s: %v\n", p, err)
				exitCode = 1
				continue
			}
			defer file.Close()
			f = file
		}

		res, err := CountProper(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wc: %s: %v\n", p, err)
			exitCode = 1
			continue
		}

		total.Lines += res.Lines
		total.Words += res.Words
		total.Bytes += res.Bytes
		total.Chars += res.Chars

		if jsonMode {
			jsonResults[p] = res
		} else {
			printCount(res, p, showLines, showWords, showBytes, showChars)
		}
	}

	if len(paths) > 1 {
		if jsonMode {
			jsonResults["total"] = total
		} else {
			printCount(total, "total", showLines, showWords, showBytes, showChars)
		}
	}

	if jsonMode {
		// Output json results
		// If single file, unwrap
		if len(paths) == 1 {
			common.Render("wc", jsonResults[paths[0]], true, out, func() {})
		} else {
			common.Render("wc", jsonResults, true, out, func() {})
		}
	}

	return exitCode
}

func printCount(res WcResult, name string, showLines, showWords, showBytes, showChars bool) {
	out := ""
	if showLines {
		out += fmt.Sprintf(" %7d", res.Lines)
	}
	if showWords {
		out += fmt.Sprintf(" %7d", res.Words)
	}
	if showChars {
		out += fmt.Sprintf(" %7d", res.Chars)
	} else if showBytes {
		out += fmt.Sprintf(" %7d", res.Bytes)
	}
	if name != "-" {
		out += fmt.Sprintf(" %s", name)
	}
	fmt.Println(out[1:]) // trim leading space
}

func init() {
	dispatch.Register(dispatch.Command{Name: "wc", Usage: "Print newline, word, and byte counts", Run: run})
}
