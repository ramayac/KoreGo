package sed

import (
	"fmt"
	"io"
	"os"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "n", Long: "quiet", Type: common.FlagBool},
		{Short: "e", Long: "expression", Type: common.FlagValue},
		{Short: "f", Long: "file", Type: common.FlagValue},
		{Short: "i", Long: "in-place", Type: common.FlagBool},
	},
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sed: %v\n", err)
		return 2
	}
	suppressDefault := flags.Has("n")
	inPlace := flags.Has("i")

	var expr string
	if es := flags.GetAll("e"); len(es) > 0 {
		expr = ""
		for i, e := range es {
			if i > 0 {
				expr += "\n"
			}
			expr += e
		}
	}
	if fs := flags.GetAll("f"); len(fs) > 0 {
		for _, f := range fs {
			b, err := os.ReadFile(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sed: %v\n", err)
				return 1
			}
			if expr != "" {
				expr += "\n"
			}
			expr += string(b)
		}
	}
	if expr == "" && len(flags.Positional) > 0 {
		expr = flags.Positional[0]
		flags.Positional = flags.Positional[1:]
	} else if expr == "" {
		// No expression and no file
		fmt.Fprintln(os.Stderr, "sed: missing command")
		return 1
	}

	insts, err := Parse(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sed: %v\n", err)
		return 1
	}

	return runEngine(insts, flags.Positional, suppressDefault, inPlace, out)
}

func init() {
	dispatch.Register(dispatch.Command{Name: "sed", Usage: "Stream editor for filtering and transforming text", Run: run})
}
