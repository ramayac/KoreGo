package find

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "j", Long: "json", Type: common.FlagBool},
		{Short: "n", Long: "name", Type: common.FlagValue},
		{Short: "t", Long: "type", Type: common.FlagValue},
	},
}

type FileInfo struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Size  int64  `json:"size"`
	Mtime string `json:"mtime"`
}

func run(args []string, out io.Writer) int {
	var parsedArgs []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") && len(a) > 2 && a != "-exec" {
			parsedArgs = append(parsedArgs, "-"+a)
		} else {
			parsedArgs = append(parsedArgs, a)
		}
	}

	flags, err := common.ParseFlags(parsedArgs, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "find: %v\n", err)
		return 1
	}

	root := "."
	if len(flags.Positional) > 0 {
		root = flags.Positional[0]
	}

	namePattern := flags.Get("n")
	typeFilter := flags.Get("t")

	var results []FileInfo

	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "find: %s: %v\n", p, err)
			return nil
		}

		if namePattern != "" {
			match, _ := filepath.Match(namePattern, filepath.Base(p))
			if !match {
				return nil
			}
		}

		tStr := "f"
		if d.IsDir() {
			tStr = "d"
		} else if d.Type()&fs.ModeSymlink != 0 {
			tStr = "l"
		}

		if typeFilter != "" && tStr != typeFilter {
			return nil
		}

		info, _ := d.Info()
		size := int64(0)
		mtime := ""
		if info != nil {
			size = info.Size()
			mtime = info.ModTime().Format(time.RFC3339)
		}

		results = append(results, FileInfo{
			Path:  p,
			Type:  tStr,
			Size:  size,
			Mtime: mtime,
		})

		return nil
	})

	if err != nil {
		return 1
	}

	jsonMode := flags.Has("j")

	common.Render("find", results, jsonMode, out, func() {
		for _, r := range results {
			fmt.Fprintln(out, r.Path)
		}
	})

	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "find", Usage: "Search for files in a directory hierarchy", Run: run})
}
