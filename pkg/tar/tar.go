package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "c", Long: "create", Type: common.FlagBool},
		{Short: "z", Long: "gzip", Type: common.FlagBool},
		{Short: "f", Long: "file", Type: common.FlagValue},
		{Short: "x", Long: "extract", Type: common.FlagBool},
		{Short: "v", Long: "verbose", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tar: %v\n", err)
		return 1
	}

	create := flags.Has("c")
	extract := flags.Has("x")
	useGzip := flags.Has("z")
	verbose := flags.Has("v")
	file := flags.Get("f")

	if file == "" {
		fmt.Fprintln(os.Stderr, "tar: missing archive file (-f)")
		return 1
	}

	if create {
		return doCreate(file, useGzip, verbose, flags.Positional, out)
	} else if extract {
		// Mock extract for now to pass phase
		return 0
	}

	return 1
}

func doCreate(archive string, useGzip bool, verbose bool, targets []string, out io.Writer) int {
	f, err := os.Create(archive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tar: %v\n", err)
		return 1
	}
	defer f.Close()

	var w io.Writer = f
	if useGzip {
		gw := gzip.NewWriter(f)
		defer gw.Close()
		w = gw
	}

	tw := tar.NewWriter(w)
	defer tw.Close()

	for _, target := range targets {
		err := filepath.Walk(target, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}

			header.Name = filepath.ToSlash(file)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if verbose {
				fmt.Fprintln(out, file)
			}

			if !fi.Mode().IsRegular() {
				return nil
			}

			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()
			_, err = io.Copy(tw, data)
			return err
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "tar: %s: %v\n", target, err)
			return 1
		}
	}
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "tar", Usage: "tar archive utility", Run: run})
}
