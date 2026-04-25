// korego is the korego multicall binary.
// It dispatches to registered commands by argv[0] (symlink mode) or argv[1]
// (subcommand mode).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ramayac/korego/internal/dispatch"

	// Import all utilities to trigger their init() registrations.
	_ "github.com/ramayac/korego/pkg/basename"
	_ "github.com/ramayac/korego/pkg/cat"
	_ "github.com/ramayac/korego/pkg/chgrp"
	_ "github.com/ramayac/korego/pkg/chmod"
	_ "github.com/ramayac/korego/pkg/chown"
	_ "github.com/ramayac/korego/pkg/cp"
	_ "github.com/ramayac/korego/pkg/cut"
	_ "github.com/ramayac/korego/pkg/daemon"
	_ "github.com/ramayac/korego/pkg/date"
	_ "github.com/ramayac/korego/pkg/df"
	_ "github.com/ramayac/korego/pkg/dirname"
	_ "github.com/ramayac/korego/pkg/du"
	_ "github.com/ramayac/korego/pkg/echo"
	_ "github.com/ramayac/korego/pkg/env"
	_ "github.com/ramayac/korego/pkg/find"
	_ "github.com/ramayac/korego/pkg/grep"
	_ "github.com/ramayac/korego/pkg/head"
	_ "github.com/ramayac/korego/pkg/hostname"
	_ "github.com/ramayac/korego/pkg/id"
	_ "github.com/ramayac/korego/pkg/kill"
	_ "github.com/ramayac/korego/pkg/ln"
	_ "github.com/ramayac/korego/pkg/ls"
	_ "github.com/ramayac/korego/pkg/mkdir"
	_ "github.com/ramayac/korego/pkg/mv"
	_ "github.com/ramayac/korego/pkg/printenv"
	_ "github.com/ramayac/korego/pkg/ps"
	_ "github.com/ramayac/korego/pkg/pwd"
	_ "github.com/ramayac/korego/pkg/readlink"
	_ "github.com/ramayac/korego/pkg/rm"
	_ "github.com/ramayac/korego/pkg/rmdir"
	_ "github.com/ramayac/korego/pkg/sed"
	_ "github.com/ramayac/korego/pkg/sha256sum"
	_ "github.com/ramayac/korego/pkg/sleep"
	_ "github.com/ramayac/korego/pkg/sort"
	_ "github.com/ramayac/korego/pkg/stat"
	_ "github.com/ramayac/korego/pkg/tail"
	_ "github.com/ramayac/korego/pkg/tar"
	_ "github.com/ramayac/korego/pkg/tee"
	_ "github.com/ramayac/korego/pkg/touch"
	_ "github.com/ramayac/korego/pkg/tr"
	_ "github.com/ramayac/korego/pkg/truefalse"
	_ "github.com/ramayac/korego/pkg/uname"
	_ "github.com/ramayac/korego/pkg/uniq"
	_ "github.com/ramayac/korego/pkg/wc"
	_ "github.com/ramayac/korego/pkg/whoami"
	_ "github.com/ramayac/korego/pkg/xargs"
	_ "github.com/ramayac/korego/pkg/yes"
)

// Version is set by -ldflags at build time.
var Version = "0.1.0"

func main() {
	name := filepath.Base(os.Args[0])

	// Subcommand dispatch: `korego <cmd> [args]`
	if name == "korego" {
		if len(os.Args) < 2 {
			dispatch.PrintHelp("korego")
			os.Exit(0)
		}
		switch os.Args[1] {
		case "--help", "-h":
			dispatch.PrintHelp("korego")
			os.Exit(0)
		case "--version":
			fmt.Println("korego version", Version)
			os.Exit(0)
		case "--list-commands":
			dispatch.ListCommands()
			os.Exit(0)
		}
		name = os.Args[1]
		os.Args = os.Args[1:] // shift so cmd sees os.Args[0] = name
	}

	cmd, ok := dispatch.Lookup(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "korego: unknown command: %s\n", name)
		os.Exit(127)
	}

	exitCode := cmd.Run(os.Args[1:], os.Stdout)
	os.Exit(exitCode)
}
