// Package touch implements the POSIX touch utility.
package touch

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "t", Long: "time", Type: common.FlagValue},
		{Short: "r", Long: "reference", Type: common.FlagValue},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// TouchResult is the --json output.
type TouchResult struct {
	Touched []string `json:"touched"`
}

// Run creates files or updates their timestamps.
func Run(paths []string, ts time.Time) (TouchResult, error) {
	var result TouchResult
	for _, p := range paths {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return result, err
		}
		f.Close()
		if err := os.Chtimes(p, ts, ts); err != nil {
			return result, err
		}
		result.Touched = append(result.Touched, p)
	}
	return result, nil
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "touch: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	ts := time.Now()

	if ref := flags.Get("r"); ref != "" {
		info, err := os.Stat(ref)
		if err != nil {
			fmt.Fprintf(os.Stderr, "touch: %v\n", err)
			return 1
		}
		ts = info.ModTime()
	} else if tStr := flags.Get("t"); tStr != "" {
		layouts := []string{"200601021504.05", "200601021504", "0601021504"}
		var parsed bool
		for _, layout := range layouts {
			t, err := time.ParseInLocation(layout, tStr, time.Local)
			if err == nil {
				ts = t
				parsed = true
				break
			}
		}
		if !parsed {
			fmt.Fprintf(os.Stderr, "touch: invalid date format: %q\n", tStr)
			return 1
		}
	}

	if len(flags.Positional) == 0 {
		fmt.Fprintln(os.Stderr, "touch: missing file operand")
		return 1
	}
	result, err := Run(flags.Positional, ts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "touch: %v\n", err)
		common.RenderError("touch", 1, "ETOUCH", err.Error(), jsonMode, out)
		return 1
	}
	common.Render("touch", result, jsonMode, out, func() {})
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "touch", Usage: "Change file timestamps or create files", Run: run})
}
