package tar

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
		{Short: "O", Long: "to-stdout", Type: common.FlagBool},
		{Short: "overwrite", Long: "overwrite", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
		{Short: "f", Long: "file", Type: common.FlagValue},
		{Short: "C", Long: "directory", Type: common.FlagValue},
		{Short: "X", Long: "exclude-from", Type: common.FlagValue},
	},
}

// preprocessOldStyleFlags expands traditional tar flag bundles like "xvf" → "-x -v -f".
func preprocessOldStyleFlags(args []string) []string {
	if len(args) == 0 {
		return args
	}
	first := args[0]
	if first == "" || first[0] == '-' {
		return args
	}

	// Check if first arg is a bundle of valid tar single-char flags.
	validChars := map[byte]bool{
		'c': true, 'x': true, 't': true, 'r': true, 'u': true,
		'z': true, 'v': true, 'O': true, 'j': true, 'J': true,
		'f': true, 'C': true, 'X': true,
	}
	isOldStyle := true
	hasModeChar := false
	for i := 0; i < len(first); i++ {
		if !validChars[first[i]] {
			isOldStyle = false
			break
		}
		if first[i] == 'c' || first[i] == 'x' || first[i] == 't' || first[i] == 'r' || first[i] == 'u' {
			hasModeChar = true
		}
	}
	if !isOldStyle || !hasModeChar {
		return args
	}

	var expanded []string
	rest := args[1:]
	for i := 0; i < len(first); i++ {
		ch := first[i]
		expanded = append(expanded, "-"+string(ch))
		if (ch == 'f' || ch == 'C' || ch == 'X') && len(rest) > 0 {
			expanded = append(expanded, rest[0])
			rest = rest[1:]
		}
	}
	expanded = append(expanded, rest...)
	return expanded
}

func run(args []string, out io.Writer) int {
	// Preprocess old-style tar flags (e.g. "xvf" → "-x -v -f").
	args = preprocessOldStyleFlags(args)

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
	toStdout := flags.Has("O")
	overwrite := flags.Has("overwrite")
	isJSON := flags.Has("json")
	file := flags.Get("f")
	dir := flags.Get("C")

	// Resolve -X exclude files: collect all values from repeatable -X flags.
	var excludePatterns []string
	for _, xf := range flags.GetAll("X") {
		if xf != "" {
			patterns, err := readExcludeFile(xf)
			if err != nil {
				common.RenderError("tar", 1, "IO", fmt.Sprintf("%s: %v", xf, err), isJSON, out)
				if !isJSON {
					fmt.Fprintf(os.Stderr, "tar: %s: %v\n", xf, err)
				}
				return 1
			}
			excludePatterns = append(excludePatterns, patterns...)
		}
	}

	if file == "" && !create {
		// No -f specified: default to stdin for extract/list.
		// Also check if there's a positional "-" meaning stdin.
		file = "-"
		for _, p := range flags.Positional {
			if p == "-" {
				file = "-"
				break
			}
		}
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

	// Resolve archive path before chdir (relative paths break after chdir).
	if file != "" && file != "-" {
		if abs, err := filepath.Abs(file); err == nil {
			file = abs
		}
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
		return doExtract(file, useGzip, verbose, toStdout, overwrite, isJSON, excludePatterns, flags.Positional, out)
	} else if list {
		return doList(file, useGzip, verbose, isJSON, excludePatterns, out)
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
		// Resolve /../ and /./ in target path for member name normalization.
		cleanTarget := filepath.Clean(target)
		strippedPrefix := ""
		if cleanTarget != target && !strings.HasPrefix(target, "/") && !strings.HasPrefix(cleanTarget, "..") {
			// Find the common suffix between target and cleanTarget.
			// WalkDir uses target as root, files are target/file.
			// Member names should be cleanTarget/file.
			// The stripped prefix is the difference.
			strippedPrefix = strings.TrimSuffix(target, cleanTarget)
			if strippedPrefix == target {
				strippedPrefix = target[:len(target)-len(cleanTarget)]
			}
			// Ensure stripped prefix ends with / for the message.
			if !strings.HasSuffix(strippedPrefix, "/") && strippedPrefix != "" {
				strippedPrefix += "/"
			}
		}

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

			// Build member name: strip the prefix path.
			memberName := filepath.ToSlash(file)
			if strippedPrefix != "" {
				if strings.HasPrefix(memberName, target) {
					// Replace the target prefix with the clean target.
					memberName = filepath.ToSlash(cleanTarget + memberName[len(target):])
				}
			}
			header.Name = memberName

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

		// Emit message about stripped prefix.
		if strippedPrefix != "" && !isJSON {
			fmt.Fprintf(os.Stderr, "tar: removing leading '%s' from member names\n", strippedPrefix)
		}
	}

	if isJSON {
		common.Render("tar", stats, true, out, func() {})
	}
	return 0
}

// readExcludeFile reads a list of exclude patterns from a file, one per line.
func readExcludeFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var patterns []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			patterns = append(patterns, line)
		}
	}
	return patterns, sc.Err()
}

// isExcluded checks if name matches any exclude pattern (path prefix match for dirs).
func isExcluded(name string, excludePatterns []string) bool {
	for _, p := range excludePatterns {
		if p == name {
			return true
		}
		// Check if name is under an excluded directory.
		if strings.HasPrefix(name, p+"/") {
			return true
		}
	}
	return false
}

// buildIncludeSet from positional args for extract/include mode.
func buildIncludeSet(positional []string) map[string]bool {
	if len(positional) == 0 {
		return nil
	}
	set := make(map[string]bool)
	for _, p := range positional {
		// Normalize: strip leading ./ and trailing /
		p = strings.TrimPrefix(p, "./")
		p = strings.TrimSuffix(p, "/")
		set[p] = true
	}
	return set
}

func doExtract(archive string, useGzip, verbose, toStdout, overwrite, isJSON bool, excludePatterns, positional []string, out io.Writer) int {
	includeSet := buildIncludeSet(positional)
	hasIncludeList := includeSet != nil
	matchedAny := false

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

	// Peek at input to detect empty streams (0 bytes = not a tarball).
	br := bufio.NewReader(r)
	if _, err := br.Peek(1); err == io.EOF {
		common.RenderError("tar", 1, "IO", "short read", isJSON, out)
		if !isJSON {
			fmt.Fprintln(os.Stderr, "tar: short read")
		}
		return 1
	}

	tr := tar.NewReader(br)
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

		// Skip excluded files.
		if isExcluded(header.Name, excludePatterns) {
			continue
		}

		// If an include list is provided, only extract matching entries.
		if includeSet != nil && !includeSet[header.Name] {
			continue
		}
		matchedAny = true

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
			if toStdout {
				continue
			}
			// Create dir with writable perms first (so files can be extracted into it),
			// actual perms will be applied after extraction.
			if err := os.MkdirAll(target, os.FileMode(header.Mode)|0300); err != nil {
				common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
				if !isJSON {
					fmt.Fprintf(os.Stderr, "tar: %v\n", err)
				}
				return 1
			}
			// Record for later chmod.
			defer os.Chmod(target, os.FileMode(header.Mode))
		case tar.TypeReg:
			if toStdout {
				if _, err := io.Copy(out, tr); err != nil {
					common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
					return 1
				}
				continue
			}
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				common.RenderError("tar", 1, "IO", err.Error(), isJSON, out)
				return 1
			}
			var flag int
			if overwrite {
				flag = os.O_WRONLY | os.O_TRUNC
			} else {
				flag = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
			}
			f, err := os.OpenFile(target, flag, os.FileMode(header.Mode))
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
			if toStdout {
				continue
			}
			dir := filepath.Dir(target)
			os.MkdirAll(dir, 0755)
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Ignore symlink errors for now
			}
		}
	}

	// If include list was provided but no files matched, return error.
	if hasIncludeList && !matchedAny {
		common.RenderError("tar", 1, "NOT_FOUND", "file not found in archive", isJSON, out)
		if !isJSON {
			for _, p := range positional {
				fmt.Fprintf(os.Stderr, "tar: %s: Not found in archive\n", p)
			}
		}
		return 1
	}

	if isJSON {
		common.Render("tar", stats, true, out, func() {})
	}
	return 0
}

// localTime returns t adjusted for the TZ environment variable.
// POSIX TZ convention: "UTC-2" means 2 hours EAST of UTC (UTC+2).
func localTime(t time.Time) time.Time {
	tz := os.Getenv("TZ")
	if tz == "" {
		return t
	}
	// Handle POSIX "UTC±N" format.
	if strings.HasPrefix(tz, "UTC") {
		rest := tz[3:]
		if rest == "" {
			return t
		}
		sign := 1
		start := 0
		if rest[0] == '-' {
			sign = 1 // POSIX: - means east of UTC
			start = 1
		} else if rest[0] == '+' {
			sign = -1 // POSIX: + means west of UTC
			start = 1
		}
		hours, err := strconv.Atoi(rest[start:])
		if err != nil {
			return t
		}
		offset := time.Duration(sign*hours) * time.Hour
		loc := time.FixedZone(tz, int(offset.Seconds()))
		return t.In(loc)
	}
	return t
}

// formatTarMode returns a BusyBox-style mode string like "drwxr-xr-x" from the tar header.
func formatTarMode(header *tar.Header) string {
	var tc byte
	switch header.Typeflag {
	case tar.TypeDir:
		tc = 'd'
	case tar.TypeSymlink:
		tc = 'l'
	case tar.TypeLink:
		tc = 'h'
	case tar.TypeChar:
		tc = 'c'
	case tar.TypeBlock:
		tc = 'b'
	case tar.TypeFifo:
		tc = 'p'
	default:
		tc = '-'
	}
	mode := int64(header.Mode) & 0777
	rwx := [9]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-'}
	if mode&0400 != 0 {
		rwx[0] = 'r'
	}
	if mode&0200 != 0 {
		rwx[1] = 'w'
	}
	if mode&0100 != 0 {
		rwx[2] = 'x'
	}
	if mode&0040 != 0 {
		rwx[3] = 'r'
	}
	if mode&0020 != 0 {
		rwx[4] = 'w'
	}
	if mode&0010 != 0 {
		rwx[5] = 'x'
	}
	if mode&0004 != 0 {
		rwx[6] = 'r'
	}
	if mode&0002 != 0 {
		rwx[7] = 'w'
	}
	if mode&0001 != 0 {
		rwx[8] = 'x'
	}
	return string(tc) + string(rwx[:])
}

func doList(archive string, useGzip, verbose, isJSON bool, excludePatterns []string, out io.Writer) int {
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

	// Peek at input to detect empty streams (0 bytes = not a tarball).
	br := bufio.NewReader(r)
	if _, err := br.Peek(1); err == io.EOF {
		common.RenderError("tar", 1, "IO", "short read", isJSON, out)
		if !isJSON {
			fmt.Fprintln(os.Stderr, "tar: short read")
		}
		return 1
	}

	tr := tar.NewReader(br)
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

		// Skip excluded files.
		if isExcluded(header.Name, excludePatterns) {
			continue
		}

		stats = append(stats, TarFileStat{
			Name: header.Name,
			Size: header.Size,
			Mode: fmt.Sprintf("%04o", header.Mode),
		})

		if !isJSON {
			if verbose {
				// BusyBox tar tvf format:
				//   %s %s/%s%10d %04d-%02d-%02d %02d:%02d:%02d %s[ -> linkname]
				mode := formatTarMode(header)
				size := header.Size
				if header.Typeflag == tar.TypeSymlink {
					size = 0
				}
				t := localTime(header.ModTime)
				line := fmt.Sprintf("%s %s/%s%10d %04d-%02d-%02d %02d:%02d:%02d %s",
					mode,
					header.Uname, header.Gname,
					size,
					t.Year(), t.Month(), t.Day(),
					t.Hour(), t.Minute(), t.Second(),
					header.Name,
				)
				if header.Typeflag == tar.TypeSymlink {
					line += " -> " + header.Linkname
				}
				fmt.Fprintln(out, line)
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
