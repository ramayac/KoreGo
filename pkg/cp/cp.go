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

// SymlinkMode controls how symbolic links are handled.
type SymlinkMode int

const (
	// SymlinkPreserve: never follow, copy symlinks as symlinks (-P / -d)
	SymlinkPreserve SymlinkMode = iota
	// SymlinkFollow: always dereference symlinks (-L)
	SymlinkFollow
	// SymlinkFollowArgs: dereference command-line arguments only (-H)
	SymlinkFollowArgs
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "r", Long: "recursive", Type: common.FlagBool},
		{Short: "R", Long: "recursive-R", Type: common.FlagBool},
		{Short: "p", Long: "preserve", Type: common.FlagBool},
		{Short: "i", Long: "interactive", Type: common.FlagBool},
		{Short: "f", Long: "force", Type: common.FlagBool},
		{Short: "d", Long: "no-dereference", Type: common.FlagBool},
		{Short: "P", Long: "no-dereference-p", Type: common.FlagBool},
		{Short: "L", Long: "dereference", Type: common.FlagBool},
		{Short: "H", Long: "dereference-command-line", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// copySingle copies src to dst, respecting the symlink mode.
// isArg indicates this is a command-line argument (relevant for SymlinkFollowArgs).
func copySingle(src, dst string, mode SymlinkMode, isArg bool, preserve bool, recursive bool, result *CpResult) error {
	// Determine whether to follow the symlink for this node
	followThisLink := mode == SymlinkFollow || (mode == SymlinkFollowArgs && isArg)

	var si os.FileInfo
	var err error
	if followThisLink {
		si, err = os.Stat(src)
	} else {
		si, err = os.Lstat(src)
	}
	if err != nil {
		return err
	}

	if si.IsDir() {
		if !recursive {
			return fmt.Errorf("omitting directory '%s'", src)
		}
		return copyDir(src, dst, mode, preserve, result)
	}

	// src is a symlink (Lstat says so) and we should not follow it
	lsi, lerr := os.Lstat(src)
	if lerr == nil && lsi.Mode()&os.ModeSymlink != 0 && !followThisLink {
		link, err := os.Readlink(src)
		if err != nil {
			return err
		}
		// Remove destination symlink if it exists
		os.Remove(dst)
		if err := os.Symlink(link, dst); err != nil {
			return err
		}
		result.Copied = append(result.Copied, CopyRecord{From: src, To: dst})
		return nil
	}

	// Regular file
	if err := copyRegularFile(src, dst, si, preserve); err != nil {
		return err
	}
	result.Copied = append(result.Copied, CopyRecord{From: src, To: dst})
	return nil
}

func copyRegularFile(src, dst string, si os.FileInfo, preserve bool) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

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

func copyDir(src, dst string, mode SymlinkMode, preserve bool, result *CpResult) error {
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, si.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		var entryInfo os.FileInfo
		if mode == SymlinkFollow {
			entryInfo, err = os.Stat(srcPath)
		} else {
			entryInfo, err = os.Lstat(srcPath)
		}
		if err != nil {
			return err
		}

		if entryInfo.IsDir() {
			if err := copyDir(srcPath, dstPath, mode, preserve, result); err != nil {
				return err
			}
		} else if entryInfo.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			os.Remove(dstPath)
			if err := os.Symlink(link, dstPath); err != nil {
				return err
			}
			result.Copied = append(result.Copied, CopyRecord{From: srcPath, To: dstPath})
		} else {
			if err := copyRegularFile(srcPath, dstPath, entryInfo, preserve); err != nil {
				return err
			}
			result.Copied = append(result.Copied, CopyRecord{From: srcPath, To: dstPath})
		}
	}
	return nil
}

// Run copies src paths to dst.
func Run(srcs []string, dst string, recursive bool, preserve bool, mode SymlinkMode) (CpResult, error) {
	var result CpResult
	dstInfo, dstErr := os.Stat(dst)
	dstIsDir := dstErr == nil && dstInfo.IsDir()

	for i, src := range srcs {
		target := dst
		if dstIsDir {
			target = filepath.Join(dst, filepath.Base(src))
		}
		if err := copySingle(src, target, mode, i == 0 || len(srcs) == 1, preserve, recursive, &result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func run(args []string, out io.Writer) int {
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

	recursive := flags.Has("r") || flags.Has("R")

	// Determine symlink mode; flag precedence: -L > -H > -P/-d > default
	var mode SymlinkMode
	if flags.Has("L") {
		mode = SymlinkFollow
	} else if flags.Has("H") {
		mode = SymlinkFollowArgs
	} else if flags.Has("P") || flags.Has("d") {
		mode = SymlinkPreserve
	} else {
		// Default: when -R/-r is used, preserve symlinks; otherwise follow
		if recursive {
			mode = SymlinkPreserve
		} else {
			mode = SymlinkFollow
		}
	}

	exitCode := 0
	var allCopied CpResult
	for _, src := range srcs {
		dstTarget := dst
		dstInfo, dstErr := os.Stat(dst)
		if dstErr == nil && dstInfo.IsDir() {
			dstTarget = filepath.Join(dst, filepath.Base(src))
		}
		// All source operands are command-line arguments (-H dereferences all of them)
		isArg := true
		var result CpResult
		if err := copySingle(src, dstTarget, mode, isArg, flags.Has("p"), recursive, &result); err != nil {
			fmt.Fprintf(os.Stderr, "cp: %v\n", err)
			exitCode = 1
		}
		allCopied.Copied = append(allCopied.Copied, result.Copied...)
	}

	common.Render("cp", allCopied, jsonMode, out, func() {})
	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "cp", Usage: "Copy files and directories", Run: run})
}
