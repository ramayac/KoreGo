// Package stat implements the POSIX stat utility.
package stat

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// StatResult is the comprehensive --json output for a file.
type StatResult struct {
	Path   string    `json:"path"`
	Size   int64     `json:"size"`
	Mode   string    `json:"mode"`
	UID    uint32    `json:"uid"`
	GID    uint32    `json:"gid"`
	Atime  time.Time `json:"atime"`
	Mtime  time.Time `json:"mtime"`
	Ctime  time.Time `json:"ctime"`
	Inode  uint64    `json:"inode"`
	Links  uint64    `json:"links"`
	Blocks int64     `json:"blocks"`
	IsDir  bool      `json:"isDir"`
	IsLink bool      `json:"isLink"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

// Run returns a StatResult for the given path.
func Run(path string) (StatResult, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return StatResult{}, err
	}
	sr := StatResult{
		Path:   path,
		Size:   info.Size(),
		Mode:   info.Mode().String(),
		IsDir:  info.IsDir(),
		IsLink: info.Mode()&os.ModeSymlink != 0,
		Mtime:  info.ModTime(),
	}
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		sr.UID = sys.Uid
		sr.GID = sys.Gid
		sr.Inode = sys.Ino
		sr.Links = uint64(sys.Nlink)
		sr.Blocks = sys.Blocks
		sr.Atime = time.Unix(sys.Atim.Sec, sys.Atim.Nsec)
		sr.Ctime = time.Unix(sys.Ctim.Sec, sys.Ctim.Nsec)
	}
	return sr, nil
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stat: %v\n", err)
		return 2
	}
	jsonMode := flags.Has("j")
	if len(flags.Positional) == 0 {
		fmt.Fprintln(os.Stderr, "stat: missing file operand")
		return 1
	}
	exitCode := 0
	for _, p := range flags.Positional {
		result, err := Run(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stat: %v\n", err)
			common.RenderError("stat", 1, "ESTAT", err.Error(), jsonMode)
			exitCode = 1
			continue
		}
		common.Render("stat", result, jsonMode, func() {
			fmt.Printf("  File: %s\n", result.Path)
			fmt.Printf("  Size: %-15d Blocks: %-10d  %s\n", result.Size, result.Blocks, result.Mode)
			fmt.Printf("  Inode: %-14d Links: %d\n", result.Inode, result.Links)
			fmt.Printf("  Uid: %-4d  Gid: %-4d\n", result.UID, result.GID)
			fmt.Printf("  Access: %s\n", result.Atime.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Modify: %s\n", result.Mtime.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Change: %s\n", result.Ctime.Format("2006-01-02 15:04:05"))
		})
	}
	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "stat", Usage: "Display file status", Run: run})
}
