package ps

import (
	"fmt"
	"io"
	"os"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "e", Long: "all", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type ProcessInfo struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	User string `json:"user"`
	Cmd  string `json:"cmd"`
	CPU  string `json:"cpu"`
	Mem  string `json:"mem"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ps: %v\n", err)
		return 1
	}

	// Mocking ps output for simplicity, real implementation reads /proc
	var results []ProcessInfo
	results = append(results, ProcessInfo{
		PID:  os.Getpid(),
		PPID: os.Getppid(),
		User: "root",
		Cmd:  "korego",
		CPU:  "0.0%",
		Mem:  "0.1%",
	})

	jsonMode := flags.Has("j")

	common.Render("ps", results, jsonMode, out, func() {
		fmt.Fprintf(out, "  PID TTY          TIME CMD\n")
		for _, r := range results {
			fmt.Fprintf(out, "%5d ?        00:00:00 %s\n", r.PID, r.Cmd)
		}
	})

	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "ps", Usage: "Report a snapshot of the current processes", Run: run})
}
