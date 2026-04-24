// Package readlink implements the POSIX readlink utility.
package readlink

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// ReadlinkResult is the --json output.
type ReadlinkResult struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "f", Long: "canonicalize", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Run reads the symlink target for path.
func Run(path string, canonicalize bool) (ReadlinkResult, error) {
	if canonicalize {
		abs, err := filepath.EvalSymlinks(path)
		if err != nil {
			return ReadlinkResult{}, err
		}
		return ReadlinkResult{Path: path, Target: abs}, nil
	}
	target, err := os.Readlink(path)
	if err != nil {
		return ReadlinkResult{}, err
	}
	return ReadlinkResult{Path: path, Target: target}, nil
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "readlink: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	if len(flags.Positional) == 0 {
		fmt.Fprintln(os.Stderr, "readlink: missing operand")
		return 1
	}
	exitCode := 0
	for _, p := range flags.Positional {
		result, err := Run(p, flags.Has("f"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "readlink: %v\n", err)
			common.RenderError("readlink", 1, "EREADLINK", err.Error(), jsonMode)
			exitCode = 1
			continue
		}
		common.Render("readlink", result, jsonMode, func() {
			fmt.Println(result.Target)
		})
	}
	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "readlink", Usage: "Print resolved symbolic links", Run: run})
}
