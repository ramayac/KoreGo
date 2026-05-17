// expand: convert tabs to spaces
package expand

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ramayac/goposix/internal/dispatch"
	"github.com/ramayac/goposix/pkg/common"
)

type ExpandResult struct{ Lines []string `json:"lines"` }

var expSpec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "t", Long: "tabs", Type: common.FlagValue},
		{Short: "i", Long: "initial", Type: common.FlagBool},
		{Long: "json", Type: common.FlagBool},
	},
}

func expandLine(line string, tabWidth int, initialOnly bool) string {
	var result strings.Builder
	col := 0
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '\t' {
			spaces := tabWidth - (col % tabWidth)
			result.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		} else {
			result.WriteByte(ch)
			col++
			if initialOnly && ch != ' ' {
				// After first non-space, stop converting tabs.
				result.WriteString(line[i+1:])
				return result.String()
			}
		}
	}
	return result.String()
}

func expandRun(args []string, out, errOut io.Writer, stdin io.Reader) int {
	flags, err := common.ParseFlags(args, expSpec)
	if err != nil {
		fmt.Fprintf(errOut, "expand: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("json")
	initialOnly := flags.Has("i")

	tabWidth := 8
	if tw := flags.Get("t"); tw != "" {
		if v, e := strconv.Atoi(tw); e == nil && v > 0 {
			tabWidth = v
		}
	}

	var input []byte
	if len(flags.Positional) == 0 {
		input, _ = io.ReadAll(stdin)
	} else {
		for _, f := range flags.Positional {
			d, _ := os.ReadFile(f)
			input = append(input, d...)
		}
	}

	text := string(input)
	if len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}
	lines := strings.Split(text, "\n")
	var outLines []string
	for _, line := range lines {
		outLines = append(outLines, expandLine(line, tabWidth, initialOnly))
	}

	if jsonMode {
		common.Render("expand", ExpandResult{Lines: outLines}, true, out, func() {})
		return 0
	}

	fmt.Fprint(out, strings.Join(outLines, "\n"))
	if len(input) > 0 && input[len(input)-1] == '\n' {
		fmt.Fprint(out, "\n")
	}
	return 0
}

func run(args []string, out io.Writer) int { return expandRun(args, out, os.Stderr, os.Stdin) }
func init() {
	dispatch.Register(dispatch.Command{Name: "expand", Usage: "Convert tabs to spaces", Run: run})
}
