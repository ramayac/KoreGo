package sha256sum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type HashResult struct {
	File string `json:"file"`
	Hash string `json:"hash"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sha256sum: %v\n", err)
		return 1
	}

	var results []HashResult
	exitCode := 0

	files := flags.Positional
	if len(files) == 0 {
		// Read from stdin?
		// For simplicity we skip stdin hashing for now and just require files
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sha256sum: %s: %v\n", file, err)
			exitCode = 1
			continue
		}
		
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			fmt.Fprintf(os.Stderr, "sha256sum: %s: %v\n", file, err)
			exitCode = 1
			f.Close()
			continue
		}
		f.Close()

		hash := hex.EncodeToString(h.Sum(nil))
		results = append(results, HashResult{File: file, Hash: hash})
	}

	jsonMode := flags.Has("j")

	common.Render("sha256sum", results, jsonMode, out, func() {
		for _, r := range results {
			fmt.Fprintf(out, "%s  %s\n", r.Hash, r.File)
		}
	})

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "sha256sum", Usage: "Compute and check SHA256 message digest", Run: run})
}
