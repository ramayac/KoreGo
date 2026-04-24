// Package tee implements the POSIX tee utility.
package tee

import (
	"fmt"
	"io"
	"os"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "a", Long: "append", Type: common.FlagBool},
	},
}

func run(args []string) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tee: %v\n", err)
		return 2
	}
	appendMode := flags.Has("a")

	var writers []io.Writer
	writers = append(writers, os.Stdout)

	var files []*os.File

	exitCode := 0
	for _, path := range flags.Positional {
		flags := os.O_WRONLY | os.O_CREATE
		if appendMode {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}
		f, err := os.OpenFile(path, flags, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tee: %s: %v\n", path, err)
			exitCode = 1
			continue
		}
		files = append(files, f)
		writers = append(writers, f)
	}

	multiWriter := io.MultiWriter(writers...)
	_, err = io.Copy(multiWriter, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tee: %v\n", err)
		exitCode = 1
	}

	for _, f := range files {
		f.Close()
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "tee", Usage: "Read from standard input and write to standard output and files", Run: run})
}
