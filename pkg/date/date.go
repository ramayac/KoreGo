package date

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
		{Short: "u", Long: "utc", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type DateInfo struct {
	ISO      string `json:"iso"`
	Unix     int64  `json:"unix"`
	UTC      string `json:"utc"`
	Timezone string `json:"timezone"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "date: %v\n", err)
		return 1
	}

	now := time.Now()
	if flags.Has("u") {
		now = now.UTC()
	}

	jsonMode := flags.Has("j")
	
	zone, _ := now.Zone()
	info := DateInfo{
		ISO:      now.Format(time.RFC3339),
		Unix:     now.Unix(),
		UTC:      now.UTC().Format(time.RFC3339),
		Timezone: zone,
	}

	common.Render("date", info, jsonMode, out, func() {
		// format
		fmt.Fprintln(out, now.Format(time.UnixDate))
	})

	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "date", Usage: "Print or set the system date and time", Run: run})
}
