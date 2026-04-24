// Package basename implements the POSIX basename utility.
package basename

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// BasenameResult is the --json output.
type BasenameResult struct {
	Result string `json:"result"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Run returns the base name of path, optionally stripping suffix.
func Run(path, suffix string) BasenameResult {
	base := filepath.Base(path)
	if suffix != "" {
		base = strings.TrimSuffix(base, suffix)
	}
	return BasenameResult{Result: base}
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "basename: %v\n", err)
		return 2
	}
	if len(flags.Positional) == 0 {
		fmt.Fprintln(os.Stderr, "basename: missing operand")
		return 1
	}
	jsonMode := flags.Has("j")
	path := flags.Positional[0]
	suffix := ""
	if len(flags.Positional) >= 2 {
		suffix = flags.Positional[1]
	}
	result := Run(path, suffix)
	common.Render("basename", result, jsonMode, func() {
		fmt.Println(result.Result)
	})
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "basename", Usage: "Strip directory and suffix from filenames", Run: run})
}
