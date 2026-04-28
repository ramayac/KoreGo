package expr

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "expr",
		Usage: "expr EXPRESSION",
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
		fmt.Fprintf(os.Stderr, "expr: %v\n", err)
		return 2
	}

	isJSON := flags.Has("json")
	restArgs := flags.Positional

	if len(restArgs) != 3 {
		common.RenderError("expr", 2, "USAGE", "naive expr only supports 'A OP B'", isJSON, out)
		if !isJSON {
			fmt.Fprintf(os.Stderr, "expr: naive expr only supports 'A OP B'\n")
		}
		return 2
	}

	a, errA := strconv.Atoi(restArgs[0])
	b, errB := strconv.Atoi(restArgs[2])
	op := restArgs[1]

	if errA != nil || errB != nil {
		common.RenderError("expr", 2, "TYPE", "non-integer argument", isJSON, out)
		if !isJSON {
			fmt.Fprintf(os.Stderr, "expr: non-integer argument\n")
		}
		return 2
	}

	var res int
	switch op {
	case "+":
		res = a + b
	case "-":
		res = a - b
	case "*":
		res = a * b
	case "/":
		if b == 0 {
			common.RenderError("expr", 2, "DIVZERO", "division by zero", isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "expr: division by zero\n")
			}
			return 2
		}
		res = a / b
	default:
		common.RenderError("expr", 2, "UNSUPPORTED", "unsupported operator", isJSON, out)
		if !isJSON {
			fmt.Fprintf(os.Stderr, "expr: unsupported operator\n")
		}
		return 2
	}

	common.Render("expr", map[string]int{"result": res}, isJSON, out, func() {
		fmt.Fprintf(out, "%d\n", res)
	})

	if res == 0 {
		return 1
	}
	return 0
}
