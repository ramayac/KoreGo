// Package cp implements the POSIX cp utility.
package cp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// CopyRecord records a single copy operation.
type CopyRecord struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// CpResult is the --json output.
type CpResult struct {
	Copied []CopyRecord `json:"copied"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "r", Long: "recursive", Type: common.FlagBool},
		{Short: "p", Long: "preserve", Type: common.FlagBool},
		{Short: "i", Long: "interactive", Type: common.FlagBool},
		{Short: "f", Long: "force", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

func copyFile(src, dst string, preserve bool) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	si, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, si.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if preserve {
		if err := os.Chtimes(dst, si.ModTime(), si.ModTime()); err != nil {
			return err
		}
	}
	return nil
}

// Run copies src paths to dst, returns the list of copy records.
func Run(srcs []string, dst string, recursive, preserve bool) (CpResult, error) {
	var result CpResult
	dstInfo, dstErr := os.Stat(dst)
	dstIsDir := dstErr == nil && dstInfo.IsDir()

	for _, src := range srcs {
		srcInfo, err := os.Lstat(src)
		if err != nil {
			return result, err
		}

		target := dst
		if dstIsDir {
			target = filepath.Join(dst, filepath.Base(src))
		}

		if srcInfo.IsDir() {
			if !recursive {
				return result, fmt.Errorf("cp: omitting directory %q", src)
			}
			if err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				rel, _ := filepath.Rel(src, path)
				dstPath := filepath.Join(target, rel)
				if d.IsDir() {
					return os.MkdirAll(dstPath, 0755)
				}
				if err := copyFile(path, dstPath, preserve); err != nil {
					return err
				}
				result.Copied = append(result.Copied, CopyRecord{From: path, To: dstPath})
				return nil
			}); err != nil {
				return result, err
			}
			continue
		}

		if err := copyFile(src, target, preserve); err != nil {
			return result, err
		}
		result.Copied = append(result.Copied, CopyRecord{From: src, To: target})
	}
	return result, nil
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cp: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	if len(flags.Positional) < 2 {
		fmt.Fprintln(os.Stderr, "cp: missing file operand")
		return 1
	}
	srcs := flags.Positional[:len(flags.Positional)-1]
	dst := flags.Positional[len(flags.Positional)-1]
	result, err := Run(srcs, dst, flags.Has("r"), flags.Has("p"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cp: %v\n", err)
		common.RenderError("cp", 1, "ECP", err.Error(), jsonMode)
		return 1
	}
	common.Render("cp", result, jsonMode, func() {})
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "cp", Usage: "Copy files and directories", Run: run})
}
