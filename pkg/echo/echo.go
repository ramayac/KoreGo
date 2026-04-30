// Package echo implements the POSIX echo utility.
package echo

import (
	"fmt"
	"io"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// EchoResult is the structured result for --json mode.
type EchoResult struct {
	Text string `json:"text"`
}

// Echo has manual flag parsing so no spec is used

// Run is the library function: given flags and words, return EchoResult.
func Run(noNewline, escape bool, words []string) EchoResult {
	text := strings.Join(words, " ")
	if escape {
		text = processEscapes(text)
	}
	return EchoResult{Text: text}
}

// processEscapes expands \n, \t, \\ etc. like echo -e.
func processEscapes(s string) string {
	replacer := strings.NewReplacer(
		`\n`, "\n",
		`\t`, "\t",
		`\r`, "\r",
		`\\`, "\\",
		`\a`, "\a",
		`\b`, "\b",
		`\v`, "\v",
	)
	return replacer.Replace(s)
}

func run(args []string, out io.Writer) int {
	jsonMode := false
	noNewline := false
	escape := false
	var positional []string

	for i, arg := range args {
		if arg == "-n" {
			noNewline = true
		} else if arg == "-e" {
			escape = true
		} else if arg == "-ne" || arg == "-en" {
			noNewline = true
			escape = true
		} else if arg == "-j" || arg == "--json" {
			jsonMode = true
		} else {
			positional = args[i:]
			break
		}
	}

	result := Run(noNewline, escape, positional)

	common.Render("echo", result, jsonMode, out, func() {
		if noNewline {
			fmt.Print(result.Text)
		} else {
			fmt.Println(result.Text)
		}
	})
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "echo",
		Usage: "Display a line of text",
		Run:   run,
	})
}
