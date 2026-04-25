package chmod

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "R", Long: "recursive", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type ChmodResult struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}

type ChmodResp struct {
	Changed []ChmodResult `json:"changed"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chmod: %v\n", err)
		return 1
	}

	if len(flags.Positional) < 2 {
		fmt.Fprintln(os.Stderr, "chmod: missing operand")
		return 1
	}

	modeStr := flags.Positional[0]
	modeNum, err := strconv.ParseUint(modeStr, 8, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chmod: invalid mode: %s\n", modeStr)
		return 1
	}
	mode := os.FileMode(modeNum)

	var res []ChmodResult
	exitCode := 0

	for _, path := range flags.Positional[1:] {
		err := os.Chmod(path, mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chmod: %v\n", err)
			exitCode = 1
		} else {
			res = append(res, ChmodResult{
				Path: path,
				Mode: fmt.Sprintf("%04o", mode),
			})
		}
	}

	if flags.Has("j") {
		common.Render("chmod", ChmodResp{Changed: res}, true, out, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "chmod", Usage: "Change file mode bits", Run: run})
}
