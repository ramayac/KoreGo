// Package echo implements the POSIX echo utility.
package echo

import (
	"fmt"
	"os"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// EchoResult is the structured result for --json mode.
type EchoResult struct {
	Text string `json:"text"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "n", Long: "no-newline", Type: common.FlagBool},
		{Short: "e", Long: "escape", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

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

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "echo: %v\n", err)
		return 2
	}

	jsonMode := flags.Has("json")
	noNewline := flags.Has("n")
	escape := flags.Has("e")

	result := Run(noNewline, escape, flags.Positional)

	common.Render("echo", result, jsonMode, func() {
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
