package chgrp

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

type ChgrpResult struct {
	Path string `json:"path"`
}

type ChgrpResp struct {
	Changed []ChgrpResult `json:"changed"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chgrp: %v\n", err)
		return 1
	}

	if len(flags.Positional) < 2 {
		fmt.Fprintln(os.Stderr, "chgrp: missing operand")
		return 1
	}

	groupStr := flags.Positional[0]
	gid, err := strconv.Atoi(groupStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chgrp: invalid group: %s\n", groupStr)
		return 1
	}

	var res []ChgrpResult
	exitCode := 0

	for _, path := range flags.Positional[1:] {
		err := os.Chown(path, -1, gid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chgrp: %v\n", err)
			exitCode = 1
		} else {
			res = append(res, ChgrpResult{Path: path})
		}
	}

	if flags.Has("j") {
		common.Render("chgrp", ChgrpResp{Changed: res}, true, out, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "chgrp", Usage: "Change group ownership", Run: run})
}
