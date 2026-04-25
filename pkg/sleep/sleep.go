package sleep

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sleep: %v\n", err)
		return 1
	}

	if len(flags.Positional) == 0 {
		fmt.Fprintln(os.Stderr, "sleep: missing operand")
		return 1
	}

	durStr := flags.Positional[0]
	var d time.Duration

	// Try Go duration first
	dur, err := time.ParseDuration(durStr)
	if err == nil {
		d = dur
	} else {
		// Try float seconds
		sec, err := strconv.ParseFloat(durStr, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sleep: invalid time interval %q\n", durStr)
			return 1
		}
		d = time.Duration(sec * float64(time.Second))
	}

	time.Sleep(d)
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "sleep", Usage: "Delay for a specified amount of time", Run: run})
}
