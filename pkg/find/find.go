package find

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
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
		{Short: "d", Long: "maxdepth", Type: common.FlagValue},
	},
}

// getMaxDepth extracts the -maxdepth N value, returning -1 if not set (unlimited).
func getMaxDepth(flags *common.ParseResult) int {
	dStr := flags.Get("d")
	if dStr == "" {
		return -1
	}
	d, err := strconv.Atoi(dStr)
	if err != nil || d < 0 {
		return -1
	}
	return d
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
	maxDepth := getMaxDepth(flags)

	// Normalize root for depth counting.
	rootClean := filepath.Clean(root)

	var results []FileInfo

	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "find: %s: %v\n", p, err)
			return nil
		}

		// Compute depth relative to root.
		if maxDepth >= 0 {
			depth := 0
			if p != rootClean {
				rel := p
				if strings.HasPrefix(rel, rootClean+"/") || rel == rootClean {
					rel = rel[len(rootClean):]
				}
				rel = strings.TrimLeft(rel, "/")
				if rel != "" {
					depth = strings.Count(rel, string(filepath.Separator)) + 1
				}
			}
			if depth > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
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

		// Normalize path: when root is ".", WalkDir returns paths without "./"
		// (e.g., "file" instead of "./file"). Prepend "./" for BusyBox compat.
		outPath := p
		if rootClean == "." && p != "." && !strings.HasPrefix(p, ".") {
			outPath = "./" + p
		}

		results = append(results, FileInfo{
			Path:  outPath,
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
