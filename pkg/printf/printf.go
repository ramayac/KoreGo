// Package printf implements the POSIX printf utility.
package printf

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// PrintfResult is the structured result for --json mode.
type PrintfResult struct {
	Output string `json:"output"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Format interprets a POSIX printf format string with arguments.
// It processes escape sequences and format specifiers, cycling through
// args when there are more specifiers than arguments (POSIX behaviour).
func Format(format string, args []string) string {
	format = processEscapes(format)

	var buf strings.Builder
	argIdx := 0

	i := 0
	for i < len(format) {
		if format[i] == '%' {
			if i+1 >= len(format) {
				buf.WriteByte('%')
				break
			}
			i++ // skip '%'

			// Literal %%
			if format[i] == '%' {
				buf.WriteByte('%')
				i++
				continue
			}

			// Parse optional flags: -, +, space, 0, #
			fmtStr := "%"
			for i < len(format) && strings.ContainsRune("-+ 0#", rune(format[i])) {
				fmtStr += string(format[i])
				i++
			}

			// Parse optional width
			for i < len(format) && format[i] >= '0' && format[i] <= '9' {
				fmtStr += string(format[i])
				i++
			}

			// Parse optional .precision
			if i < len(format) && format[i] == '.' {
				fmtStr += "."
				i++
				for i < len(format) && format[i] >= '0' && format[i] <= '9' {
					fmtStr += string(format[i])
					i++
				}
			}

			// Parse conversion specifier
			if i >= len(format) {
				buf.WriteString(fmtStr)
				break
			}
			spec := format[i]
			i++

			// Get current arg (cycle back to start if exhausted)
			arg := ""
			if len(args) > 0 {
				arg = args[argIdx%len(args)]
				argIdx++
			}

			switch spec {
			case 's':
				buf.WriteString(fmt.Sprintf(fmtStr+"s", arg))
			case 'd', 'i':
				n, _ := strconv.ParseInt(arg, 0, 64)
				buf.WriteString(fmt.Sprintf(fmtStr+"d", n))
			case 'o':
				n, _ := strconv.ParseInt(arg, 0, 64)
				buf.WriteString(fmt.Sprintf(fmtStr+"o", n))
			case 'x':
				n, _ := strconv.ParseInt(arg, 0, 64)
				buf.WriteString(fmt.Sprintf(fmtStr+"x", n))
			case 'X':
				n, _ := strconv.ParseInt(arg, 0, 64)
				buf.WriteString(fmt.Sprintf(fmtStr+"X", n))
			case 'f':
				f, _ := strconv.ParseFloat(arg, 64)
				buf.WriteString(fmt.Sprintf(fmtStr+"f", f))
			case 'e':
				f, _ := strconv.ParseFloat(arg, 64)
				buf.WriteString(fmt.Sprintf(fmtStr+"e", f))
			case 'g':
				f, _ := strconv.ParseFloat(arg, 64)
				buf.WriteString(fmt.Sprintf(fmtStr+"g", f))
			case 'c':
				if len(arg) > 0 {
					buf.WriteByte(arg[0])
				}
			default:
				buf.WriteString(fmtStr + string(spec))
			}
		} else {
			buf.WriteByte(format[i])
			i++
		}
	}

	// POSIX: If there are remaining args not consumed, re-process the format.
	// We implement the simpler "cycle" variant above which handles the common
	// case. For full spec compliance with multiple re-passes we'd loop here,
	// but cycling is the widely expected behaviour.

	return buf.String()
}

// processEscapes expands POSIX backslash escape sequences in a format string.
func processEscapes(s string) string {
	var buf strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			i++ // skip backslash
			switch s[i] {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			case '\\':
				buf.WriteByte('\\')
			case 'a':
				buf.WriteByte('\a')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 'v':
				buf.WriteByte('\v')
			case '0':
				// Octal escape: \0NNN (1-3 octal digits)
				end := i + 1
				for end < len(s) && end < i+4 && s[end] >= '0' && s[end] <= '7' {
					end++
				}
				if end > i+1 {
					val, err := strconv.ParseUint(s[i+1:end], 8, 8)
					if err == nil {
						buf.WriteByte(byte(val))
						i = end
						continue
					}
				}
				// Just \0 with no valid digits = NUL
				buf.WriteByte(0)
			default:
				// Unknown escape: preserve literally
				buf.WriteByte('\\')
				buf.WriteByte(s[i])
			}
			i++
		} else {
			buf.WriteByte(s[i])
			i++
		}
	}
	return buf.String()
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "printf: %v\n", err)
		return 1
	}

	jsonMode := flags.Has("json")
	restArgs := flags.Positional

	if len(restArgs) == 0 {
		common.RenderError("printf", 1, "MISSING_OPERAND", "missing operand", jsonMode, out)
		if !jsonMode {
			fmt.Fprintf(os.Stderr, "printf: missing operand\n")
		}
		return 1
	}

	format := restArgs[0]
	fmtArgs := restArgs[1:]

	output := Format(format, fmtArgs)

	common.Render("printf", PrintfResult{Output: output}, jsonMode, out, func() {
		fmt.Fprint(out, output)
	})
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "printf",
		Usage: "Format and print data",
		Run:   run,
	})
}
