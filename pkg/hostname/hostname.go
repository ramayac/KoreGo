// Package hostname implements the POSIX hostname utility.
package hostname

import (
	"fmt"
	"os"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// HostnameResult is the structured result for --json mode.
type HostnameResult struct {
	Name string `json:"hostname"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Run returns the system hostname.
func Run() (HostnameResult, error) {
	name, err := os.Hostname()
	if err != nil {
		return HostnameResult{}, err
	}
	return HostnameResult{Name: name}, nil
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hostname: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("json")

	result, err := Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "hostname: %v\n", err)
		common.RenderError("hostname", 1, "EHOSTNAME", err.Error(), jsonMode)
		return 1
	}

	common.Render("hostname", result, jsonMode, func() {
		fmt.Println(result.Name)
	})
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "hostname",
		Usage: "Print or set the system hostname",
		Run:   run,
	})
}
