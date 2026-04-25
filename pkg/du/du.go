package du

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "h", Long: "human-readable", Type: common.FlagBool},
		{Short: "s", Long: "summarize", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type DirInfo struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Files int    `json:"files"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "du: %v\n", err)
		return 1
	}

	roots := flags.Positional
	if len(roots) == 0 {
		roots = []string{"."}
	}

	var results []DirInfo
	exitCode := 0

	for _, root := range roots {
		var totalSize int64
		var count int
		
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "du: cannot read directory %q: %v\n", p, err)
				exitCode = 1
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			totalSize += info.Size()
			count++
			return nil
		})
		
		if err != nil {
			exitCode = 1
		}

		results = append(results, DirInfo{
			Path:  root,
			Size:  totalSize,
			Files: count,
		})
	}

	jsonMode := flags.Has("j")
	human := flags.Has("h")

	common.Render("du", results, jsonMode, out, func() {
		for _, r := range results {
			if human {
				fmt.Fprintf(out, "%dM\t%s\n", r.Size/1024/1024, r.Path)
			} else {
				fmt.Fprintf(out, "%d\t%s\n", r.Size/1024, r.Path)
			}
		}
	})

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "du", Usage: "Estimate file space usage", Run: run})
}
