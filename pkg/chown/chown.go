package chown

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "R", Long: "recursive", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type ChownResult struct {
	Path string `json:"path"`
}

type ChownResp struct {
	Changed []ChownResult `json:"changed"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chown: %v\n", err)
		return 1
	}

	if len(flags.Positional) < 2 {
		fmt.Fprintln(os.Stderr, "chown: missing operand")
		return 1
	}

	ownerStr := flags.Positional[0]
	parts := strings.SplitN(ownerStr, ":", 2)
	
	uid := -1
	gid := -1
	
	if parts[0] != "" {
		if val, err := strconv.Atoi(parts[0]); err == nil {
			uid = val
		}
	}
	if len(parts) > 1 && parts[1] != "" {
		if val, err := strconv.Atoi(parts[1]); err == nil {
			gid = val
		}
	}

	var res []ChownResult
	exitCode := 0

	for _, path := range flags.Positional[1:] {
		err := os.Chown(path, uid, gid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chown: %v\n", err)
			exitCode = 1
		} else {
			res = append(res, ChownResult{Path: path})
		}
	}

	if flags.Has("j") {
		common.Render("chown", ChownResp{Changed: res}, true, out, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "chown", Usage: "Change file owner and group", Run: run})
}
