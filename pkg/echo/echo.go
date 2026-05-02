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
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			c := s[i]
			switch c {
			case 'n': sb.WriteByte('\n')
			case 't': sb.WriteByte('\t')
			case 'r': sb.WriteByte('\r')
			case '\\': sb.WriteByte('\\')
			case 'a': sb.WriteByte('\a')
			case 'b': sb.WriteByte('\b')
			case 'v': sb.WriteByte('\v')
			case '0':
				// octal
				if i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '7' {
					oct := 0
					j := i + 1
					for ; j < len(s) && j < i+4 && s[j] >= '0' && s[j] <= '7'; j++ {
						oct = oct*8 + int(s[j]-'0')
					}
					sb.WriteByte(byte(oct))
					i = j - 1
				} else {
					sb.WriteByte(0)
				}
			case 'x':
				// hex
				if i+1 < len(s) {
					hex := 0
					j := i + 1
					for ; j < len(s) && j < i+3; j++ {
						h := s[j]
						if h >= '0' && h <= '9' {
							hex = hex*16 + int(h-'0')
						} else if h >= 'a' && h <= 'f' {
							hex = hex*16 + int(h-'a'+10)
						} else if h >= 'A' && h <= 'F' {
							hex = hex*16 + int(h-'A'+10)
						} else {
							break
						}
					}
					if j > i+1 {
						sb.WriteByte(byte(hex))
						i = j - 1
					} else {
						sb.WriteByte('\\')
						sb.WriteByte('x')
					}
				} else {
					sb.WriteByte('\\')
					sb.WriteByte('x')
				}
			default:
				sb.WriteByte('\\')
				sb.WriteByte(c)
			}
		} else {
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}

func run(args []string, out io.Writer) int {
	jsonMode := false
	noNewline := false
	escape := false
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-j" || arg == "--json" {
			jsonMode = true
			continue
		}
		if arg == "-n" {
			noNewline = true
		} else if arg == "-e" {
			escape = true
		} else if arg == "-ne" || arg == "-en" {
			noNewline = true
			escape = true
		} else {
			positional = append(positional, arg)
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
