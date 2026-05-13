//go:build linux

package stat

import (
	"os"
	"syscall"
	"time"
)

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
