package gzip

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// GzipStat holds the result of compressing or decompressing a single file.
type GzipStat struct {
	File         string  `json:"file"`
	OriginalSize int64   `json:"originalSize"`
	NewSize      int64   `json:"newSize"`
	Ratio        float64 `json:"ratio"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "d", Long: "decompress", Type: common.FlagBool},
		{Short: "c", Long: "stdout", Type: common.FlagBool},
		{Short: "k", Long: "keep", Type: common.FlagBool},
		{Short: "f", Long: "force", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
		// Numeric compression levels -1 through -9 (boolean-like for presence detection).
		{Short: "1", Long: "", Type: common.FlagBool},
		{Short: "2", Long: "", Type: common.FlagBool},
		{Short: "3", Long: "", Type: common.FlagBool},
		{Short: "4", Long: "", Type: common.FlagBool},
		{Short: "5", Long: "", Type: common.FlagBool},
		{Short: "6", Long: "", Type: common.FlagBool},
		{Short: "7", Long: "", Type: common.FlagBool},
		{Short: "8", Long: "", Type: common.FlagBool},
		{Short: "9", Long: "", Type: common.FlagBool},
	},
}

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
	return execute(args, out, false)
}

func runGunzip(args []string, out io.Writer) int {
	return execute(args, out, true)
}

// getCompressionLevel returns the compression level from flags, or flate.DefaultCompression.
func getCompressionLevel(flags *common.ParseResult) int {
	// Check for numeric compression level flags (highest priority last).
	for i := 9; i >= 1; i-- {
		if flags.Has(fmt.Sprintf("%d", i)) {
			return i
		}
	}
	return flate.DefaultCompression
}

func execute(args []string, out io.Writer, forceDecompress bool) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
		return 1
	}

	isJSON := flags.Has("json")
	toStdout := flags.Has("c")
	keepOriginal := flags.Has("k") || toStdout
	force := flags.Has("f")
	decompress := forceDecompress || flags.Has("d")
	level := getCompressionLevel(flags)
	
	files := flags.Positional

	if len(files) == 0 {
		if decompress {
			gr, err := gzip.NewReader(os.Stdin)
			if err != nil {
				if !isJSON {
					fmt.Fprintf(os.Stderr, "gzip: stdin: %v\n", err)
				}
				return 1
			}
			io.Copy(out, gr)
			gr.Close()
		} else {
			gw, err := gzip.NewWriterLevel(out, level)
			if err != nil {
				if !isJSON {
					fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
				}
				return 1
			}
			io.Copy(gw, os.Stdin)
			gw.Close()
		}
		return 0
	}

	exitCode := 0
	var stats []GzipStat

	for _, file := range files {
		// Handle stdin/stdout via "-".
		if file == "-" {
			if decompress {
				gr, err := gzip.NewReader(os.Stdin)
				if err != nil {
					if !isJSON {
						fmt.Fprintf(os.Stderr, "gzip: stdin: %v\n", err)
					}
					return 1
				}
				io.Copy(out, gr)
				gr.Close()
			} else {
				gw, err := gzip.NewWriterLevel(out, level)
				if err != nil {
					if !isJSON {
						fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
					}
					return 1
				}
				io.Copy(gw, os.Stdin)
				gw.Close()
			}
			continue
		}

		inInfo, err := os.Stat(file)
		if err != nil {
			common.RenderError("gzip", 1, "STAT_FAIL", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
			}
			exitCode = 1
			continue
		}

		in, err := os.Open(file)
		if err != nil {
			common.RenderError("gzip", 1, "OPEN_FAIL", err.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
			}
			exitCode = 1
			continue
		}

		var targetWriter io.Writer
		var outFile *os.File
		var outName string

		if toStdout {
			targetWriter = out
		} else {
			if decompress {
				outName = strings.TrimSuffix(file, ".gz")
				if outName == file {
					outName = file + ".unpacked"
				}
			} else {
				outName = file + ".gz"
			}

			if !force {
				if _, err := os.Stat(outName); err == nil {
					if !isJSON {
						fmt.Fprintf(os.Stderr, "gzip: %s already exists; use -f to overwrite\n", outName)
					}
					in.Close()
					exitCode = 1
					continue
				}
			}

			outFile, err = os.OpenFile(outName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, inInfo.Mode())
			if err != nil {
				common.RenderError("gzip", 1, "CREATE_FAIL", err.Error(), isJSON, out)
				if !isJSON {
					fmt.Fprintf(os.Stderr, "gzip: %v\n", err)
				}
				in.Close()
				exitCode = 1
				continue
			}
			targetWriter = outFile
		}

		var processErr error
		if decompress {
			gr, err := gzip.NewReader(in)
			if err != nil {
				processErr = err
			} else {
				_, processErr = io.Copy(targetWriter, gr)
				gr.Close()
			}
		} else {
			gw, err := gzip.NewWriterLevel(targetWriter, level)
			if err != nil {
				processErr = err
			} else {
				_, processErr = io.Copy(gw, in)
				gw.Close()
			}
		}

		if outFile != nil {
			outFile.Close()
		}
		in.Close()

		if processErr != nil {
			common.RenderError("gzip", 1, "PROCESS_FAIL", processErr.Error(), isJSON, out)
			if !isJSON {
				fmt.Fprintf(os.Stderr, "gzip: %v\n", processErr)
			}
			if outFile != nil {
				os.Remove(outName) // remove incomplete output
			}
			exitCode = 1
			continue
		}

		var outSize int64
		if !toStdout && outFile != nil {
			if outInfo, err := os.Stat(outName); err == nil {
				outSize = outInfo.Size()
			}
			if !keepOriginal {
				os.Remove(file)
			}
		}

		inSize := inInfo.Size()
		var ratio float64
		if inSize > 0 && outSize > 0 {
			if decompress {
				ratio = float64(outSize) / float64(inSize)
			} else {
				ratio = float64(inSize) / float64(outSize)
			}
		}

		stats = append(stats, GzipStat{
			File:         file,
			OriginalSize: inSize,
			NewSize:      outSize,
			Ratio:        ratio,
		})
	}

	if isJSON {
		common.Render("gzip", stats, true, out, func() {})
	}

	return exitCode
}
