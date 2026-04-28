package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// TarFileStat holds metadata for a single file in the archive for JSON output.
type TarFileStat struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Mode string `json:"mode"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "c", Long: "create", Type: common.FlagBool},
		{Short: "x", Long: "extract", Type: common.FlagBool},
		{Short: "t", Long: "list", Type: common.FlagBool},
		{Short: "z", Long: "gzip", Type: common.FlagBool},
		{Short: "v", Long: "verbose", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
		{Short: "f", Long: "file", Type: common.FlagValue},
		{Short: "C", Long: "directory", Type: common.FlagValue},
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
	list := flags.Has("t")
	useGzip := flags.Has("z")
	verbose := flags.Has("v")
	isJSON := flags.Has("json")
	file := flags.Get("f")
	dir := flags.Get("C")

	if file == "" && !create {
		// POSIX tar doesn't strictly require -f, but for our implementation it's required for now
		common.RenderError("tar", 1, "USAGE", "missing archive file (-f)", isJSON, out)
		if !isJSON {
			fmt.Fprintln(os.Stderr, "tar: missing archive file (-f)")
		}
		return 1
	}

	modeCount := 0
	if create {
		modeCount++
	}
	if extract {
		modeCount++
	}
	if list {
		modeCount++
	}
	if modeCount != 1 {
		common.RenderError("tar", 1, "USAGE", "must specify exactly one of -c, -x, or -t", isJSON, out)
		if !isJSON {
			fmt.Fprintln(os.Stderr, "tar: must specify exactly one of -c, -x, or -t")
		}
		return 1
	}

	var curDir string
	if dir != "" {
		curDir, err = os.Getwd()
		if err != nil {
			common.RenderError("tar", 1, "DIR", err.Error(), isJSON, out)
			return 1
		}
		if err := os.Chdir(dir); err != nil {
			common.RenderError("tar", 1, "DIR", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}
		defer os.Chdir(curDir)
	}

	if create {
		return doCreate(file, useGzip, verbose, isJSON, flags.Positional, out)
	} else if extract {
		return doExtract(file, useGzip, verbose, isJSON, out)
	} else if list {
		return doList(file, useGzip, verbose, isJSON, out)
	}

	return 1
}

func doCreate(archive string, useGzip, verbose, isJSON bool, targets []string, out io.Writer) int {
	var w io.Writer
	if archive == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(archive)
		if err != nil {
			common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}
		defer f.Close()
		w = f
	}

	if useGzip {
		gw := gzip.NewWriter(w)
		defer gw.Close()
		w = gw
	}

	tw := tar.NewWriter(w)
	defer tw.Close()

	var stats []TarFileStat

	for _, target := range targets {
		err := filepath.Walk(target, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Do not archive the output file itself
			if archive != "-" {
				absArchive, _ := filepath.Abs(archive)
				absFile, _ := filepath.Abs(file)
				if absArchive == absFile {
					return nil
				}
			}

			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}

			header.Name = filepath.ToSlash(file)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			stats = append(stats, TarFileStat{
				Name: header.Name,
				Size: header.Size,
				Mode: fi.Mode().String(),
			})

			if verbose && !isJSON {
				fmt.Fprintln(out, header.Name)
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
			common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %s: %v\n", target, err)
			}
			return 1
		}
	}

	if isJSON {
		common.Render("tar", stats, true, out, func() {})
	}
	return 0
}

func doExtract(archive string, useGzip, verbose, isJSON bool, out io.Writer) int {
	var r io.Reader
	if archive == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(archive)
		if err != nil {
			common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}
		defer f.Close()
		r = f
	}

	if useGzip {
		gr, err := gzip.NewReader(r)
		if err != nil {
			common.RenderError("tar", 1, "GZIP", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}
		defer gr.Close()
		r = gr
	}

	tr := tar.NewReader(r)
	var stats []TarFileStat

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}

		// Prevent zip slip
		target := filepath.Clean(header.Name)
		if strings.HasPrefix(target, "..") || strings.HasPrefix(target, "/") {
			continue // Skip unsafe paths
		}

		if verbose && !isJSON {
			fmt.Fprintln(out, header.Name)
		}

		stats = append(stats, TarFileStat{
			Name: header.Name,
			Size: header.Size,
			Mode: fmt.Sprintf("%04o", header.Mode),
		})

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
				if !isJSON {
					fmt.Fprintf(os.Stderr, "tar: %v\n", err)
				}
				return 1
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
				return 1
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
				if !isJSON {
					fmt.Fprintf(os.Stderr, "tar: %v\n", err)
				}
				return 1
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
				return 1
			}
			f.Close()
			// Restore timestamp
			os.Chtimes(target, header.AccessTime, header.ModTime)
		case tar.TypeSymlink:
			dir := filepath.Dir(target)
			os.MkdirAll(dir, 0755)
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Ignore symlink errors for now
			}
		}
	}

	if isJSON {
		common.Render("tar", stats, true, out, func() {})
	}
	return 0
}

func doList(archive string, useGzip, verbose, isJSON bool, out io.Writer) int {
	var r io.Reader
	if archive == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(archive)
		if err != nil {
			common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}
		defer f.Close()
		r = f
	}

	if useGzip {
		gr, err := gzip.NewReader(r)
		if err != nil {
			common.RenderError("tar", 1, "GZIP", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}
		defer gr.Close()
		r = gr
	}

	tr := tar.NewReader(r)
	var stats []TarFileStat

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			}
			return 1
		}

		stats = append(stats, TarFileStat{
			Name: header.Name,
			Size: header.Size,
			Mode: fmt.Sprintf("%04o", header.Mode),
		})

		if !isJSON {
			if verbose {
				fmt.Fprintf(out, "%s %d %s\n", os.FileMode(header.Mode).String(), header.Size, header.Name)
			} else {
				fmt.Fprintln(out, header.Name)
			}
		}
	}

	if isJSON {
		common.Render("tar", stats, true, out, func() {})
	}
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{Name: "tar", Usage: "tar archive utility", Run: run})
}
