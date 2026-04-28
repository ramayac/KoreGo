package diff

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "diff",
		Usage: "diff FILE1 FILE2",
		Run:   run,
	})
}

func run(args []string, out io.Writer) int {
	spec := common.FlagSpec{
		Defs: []common.FlagDef{
			{Short: "u", Long: "unified", Type: common.FlagBool},
			{Short: "j", Long: "json", Type: common.FlagBool},
		},
	}
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		return 2
	}

	isJSON := flags.Has("json")
	files := flags.Positional

	if len(files) != 2 {
		common.RenderError("diff", 2, "USAGE", "missing operand", isJSON, out)
		if !isJSON {
			fmt.Fprintf(os.Stderr, "diff: missing operand\n")
		}
		return 2
	}

	b1, err := os.ReadFile(files[0])
	if err != nil {
		common.RenderError("diff", 2, "IO", err.Error(), isJSON, out)
		if !isJSON {
			fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		}
		return 2
	}
	b2, err := os.ReadFile(files[1])
	if err != nil {
		common.RenderError("diff", 2, "IO", err.Error(), isJSON, out)
		if !isJSON {
			fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		}
		return 2
	}

	if bytes.Equal(b1, b2) {
		common.Render("diff", map[string]interface{}{"differ": false}, isJSON, out, func() {})
		return 0 // Files match, exit 0
	}

	// Naive output instead of full hunk algorithm
	msg := fmt.Sprintf("Files %s and %s differ\n", files[0], files[1])
	common.Render("diff", map[string]interface{}{"differ": true, "msg": msg}, isJSON, out, func() {
		fmt.Fprint(out, msg)
	})
	return 1 // Files differ
}
