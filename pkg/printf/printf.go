package printf

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "printf",
		Usage: "printf FORMAT [ARGUMENT...]",
		Run:   run,
	})
}

func run(args []string, out io.Writer) int {
	spec := common.FlagSpec{
		Defs: []common.FlagDef{
			{Short: "j", Long: "json", Type: common.FlagBool},
		},
	}
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "printf: %v\n", err)
		return 1
	}

	isJSON := flags.Has("json")
	restArgs := flags.Positional

	if len(restArgs) == 0 {
		common.RenderError("printf", 1, "MISSING_OPERAND", "missing operand", isJSON, out)
		if !isJSON {
			fmt.Fprintf(os.Stderr, "printf: missing operand\n")
		}
		return 1
	}

	format := restArgs[0]
	format = strings.ReplaceAll(format, "\\n", "\n")
	format = strings.ReplaceAll(format, "\\t", "\t")
	format = strings.ReplaceAll(format, "\\r", "\r")

	formatArgs := make([]interface{}, len(restArgs)-1)
	for i, arg := range restArgs[1:] {
		formatArgs[i] = arg
	}

	output := ""
	if len(formatArgs) > 0 {
		output = fmt.Sprintf(format, formatArgs...)
	} else {
		output = format
	}

	common.Render("printf", map[string]string{"output": output}, isJSON, out, func() {
		fmt.Fprint(out, output)
	})
	return 0
}
