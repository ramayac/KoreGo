package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "gzip",
		Usage: "gzip [FILE]...",
		Run:   runGzip,
	})
	dispatch.Register(dispatch.Command{
		Name:  "gunzip",
		Usage: "gunzip [FILE]...",
		Run:   runGunzip,
	})
}

func runGzip(args []string, out io.Writer) int {
	spec := common.FlagSpec{
		Defs: []common.FlagDef{
			{Short: "d", Long: "decompress", Type: common.FlagBool},
			{Short: "j", Long: "json", Type: common.FlagBool},
		},
	}
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
		return 1
	}

	isJSON := flags.Has("json")
	files := flags.Positional

	if flags.Has("d") {
		return runGunzip(args, out)
	}

	if len(files) == 0 {
		gw := gzip.NewWriter(out)
		io.Copy(gw, os.Stdin)
		gw.Close()
		return 0
	}

	for _, file := range files {
		in, err := os.Open(file)
		if err != nil {
			common.RenderError("gzip", 1, "OPEN_FAIL", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
			}
			return 1
		}

		outName := file + ".gz"
		outFile, err := os.Create(outName)
		if err != nil {
			in.Close()
			common.RenderError("gzip", 1, "CREATE_FAIL", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
			}
			return 1
		}

		gw := gzip.NewWriter(outFile)
		io.Copy(gw, in)
		gw.Close()
		outFile.Close()
		in.Close()

		os.Remove(file)
	}

	common.Render("gzip", map[string]string{"status": "compressed successfully"}, isJSON, out, func() {})
	return 0
}

func runGunzip(args []string, out io.Writer) int {
	spec := common.FlagSpec{
		Defs: []common.FlagDef{
			{Short: "j", Long: "json", Type: common.FlagBool},
		},
	}
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gunzip: %v\n", err)
		return 1
	}

	isJSON := flags.Has("json")
	files := flags.Positional

	if len(files) == 0 {
		gr, err := gzip.NewReader(os.Stdin)
		if err == nil {
			io.Copy(out, gr)
			gr.Close()
		}
		return 0
	}

	for _, file := range files {
		in, err := os.Open(file)
		if err != nil {
			common.RenderError("gunzip", 1, "OPEN_FAIL", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "gunzip: %v\n", err)
			}
			return 1
		}

		outName := strings.TrimSuffix(file, ".gz")
		if outName == file {
			outName = file + ".unpacked"
		}

		outFile, err := os.Create(outName)
		if err != nil {
			in.Close()
			common.RenderError("gunzip", 1, "CREATE_FAIL", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "gunzip: %v\n", err)
			}
			return 1
		}

		gr, err := gzip.NewReader(in)
		if err == nil {
			io.Copy(outFile, gr)
			gr.Close()
		}
		outFile.Close()
		in.Close()

		os.Remove(file)
	}

	common.Render("gunzip", map[string]string{"status": "decompressed successfully"}, isJSON, out, func() {})
	return 0
}
