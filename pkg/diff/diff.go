// Package diff implements the POSIX diff utility.
package diff

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

// Hunk represents a section of difference between files.
type Hunk struct {
	OldStart int      `json:"oldStart"`
	OldLines int      `json:"oldLines"`
	NewStart int      `json:"newStart"`
	NewLines int      `json:"newLines"`
	Lines    []string `json:"lines"`
}

// DiffResult is the structured result for --json mode.
type DiffResult struct {
	Files  []string `json:"files"`
	Differ bool     `json:"differ"`
	Hunks  []Hunk   `json:"hunks"`
}

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "u", Long: "unified", Type: common.FlagBool},
		{Short: "U", Long: "unified-context", Type: common.FlagValue},
		{Short: "b", Long: "ignore-space-change", Type: common.FlagBool},
		{Short: "B", Long: "ignore-blank-lines", Type: common.FlagBool},
		{Short: "q", Long: "brief", Type: common.FlagBool},
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type diffOp int

const (
	opEq diffOp = iota
	opDel
	opIns
)

type diffItem struct {
	op   diffOp
	text string
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Diff computes the differences between two sequences of strings using an O(ND) algorithm.
func Diff(a, b []string) []diffItem {
	n := len(a)
	m := len(b)
	maxD := n + m

	if maxD == 0 {
		return nil
	}

	v := make([]int, 2*maxD+1)
	trace := make([][]int, 0, maxD+1)

	var d, k, x, y int

Outer:
	for d = 0; d <= maxD; d++ {
		for k = -d; k <= d; k += 2 {
			if k == -d || (k != d && v[maxD+k-1] < v[maxD+k+1]) {
				x = v[maxD+k+1]
			} else {
				x = v[maxD+k-1] + 1
			}
			y = x - k

			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[maxD+k] = x

			if x >= n && y >= m {
				tr := make([]int, 2*maxD+1)
				copy(tr, v)
				trace = append(trace, tr)
				break Outer
			}
		}
		tr := make([]int, 2*maxD+1)
		copy(tr, v)
		trace = append(trace, tr)
	}

	// Backtrack
	var script []diffItem
	x = n
	y = m

	for d > 0 {
		k = x - y
		var prevK int
		if k == -d || (k != d && trace[d-1][maxD+k-1] < trace[d-1][maxD+k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := trace[d-1][maxD+prevK]
		var startX int
		var isDel bool
		if prevK == k - 1 { 
			startX = prevX + 1
			isDel = true
		} else { 
			startX = prevX
			isDel = false
		}

		for x > startX {
			x--
			y--
			script = append(script, diffItem{op: opEq, text: a[x]})
		}
		
		if isDel {
			x--
			script = append(script, diffItem{op: opDel, text: a[x]})
		} else {
			y--
			script = append(script, diffItem{op: opIns, text: b[y]})
		}
		
		d--
	}

	for x > 0 && y > 0 {
		x--
		y--
		script = append(script, diffItem{op: opEq, text: a[x]})
	}

	// Reverse script
	for i, j := 0, len(script)-1; i < j; i, j = i+1, j-1 {
		script[i], script[j] = script[j], script[i]
	}

	return script
}

// BuildHunks builds unified diff hunks from a diff script with a given number of context lines.
func BuildHunks(script []diffItem, context int) []Hunk {
	var hunks []Hunk
	
	i := 0
	oldLine := 1
	newLine := 1

	for i < len(script) {
		// Skip initial equal lines
		for i < len(script) && script[i].op == opEq {
			oldLine++
			newLine++
			i++
		}
		if i >= len(script) {
			break
		}

		// Found a difference, backtrack for context
		startIdx := i - context
		if startIdx < 0 {
			startIdx = 0
		}
		
		// Find end of current difference block
		endIdx := i
		for endIdx < len(script) {
			nextDiff := endIdx
			// Skip diff ops
			for nextDiff < len(script) && script[nextDiff].op != opEq {
				nextDiff++
			}

			// Find next equal block
			eqCount := 0
			for nextDiff < len(script) && script[nextDiff].op == opEq {
				eqCount++
				nextDiff++
			}
			
			if nextDiff >= len(script) {
				if eqCount > context {
					endIdx = nextDiff - eqCount + context
				} else {
					endIdx = len(script)
				}
				break
			}
			
			if eqCount > 2*context {
				// The next diff is far enough away that we can end the current hunk
				endIdx = endIdx + context
				// Wait, if endIdx + context goes into the equal block, the end of the hunk is nextDiff minus the remaining eq ops
				// Actually, the equal block starts at the original `nextDiff` before the second loop.
				// Let's record the start of the equal block.
				eqStart := nextDiff - eqCount
				endIdx = eqStart + context
				if endIdx > len(script) {
					endIdx = len(script)
				}
				break
			}
			endIdx = nextDiff
		}

		// Ensure endIdx does not exceed script length
		if endIdx > len(script) {
			endIdx = len(script)
		}

		hunk := Hunk{
			OldStart: 1, // Will be computed
			NewStart: 1, // Will be computed
		}

		// Recompute line numbers for the hunk start
		hunkOld := 1
		hunkNew := 1
		for j := 0; j < startIdx; j++ {
			if script[j].op == opEq || script[j].op == opDel {
				hunkOld++
			}
			if script[j].op == opEq || script[j].op == opIns {
				hunkNew++
			}
		}

		hunk.OldStart = hunkOld
		hunk.NewStart = hunkNew

		for j := startIdx; j < endIdx; j++ {
			switch script[j].op {
			case opEq:
				hunk.Lines = append(hunk.Lines, " "+script[j].text)
				hunk.OldLines++
				hunk.NewLines++
			case opDel:
				hunk.Lines = append(hunk.Lines, "-"+script[j].text)
				hunk.OldLines++
			case opIns:
				hunk.Lines = append(hunk.Lines, "+"+script[j].text)
				hunk.NewLines++
			}
		}

		hunks = append(hunks, hunk)
		i = endIdx
		
		// Update oldLine and newLine to the end of the hunk
		oldLine = hunkOld + hunk.OldLines
		newLine = hunkNew + hunk.NewLines
	}

	return hunks
}

// GenerateDiff compares two file contents and produces hunks and a differ boolean.
func GenerateDiff(content1, content2 string, contextLines int) (bool, []Hunk) {
	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")
	
	// If the file does not end in a newline, strings.Split produces a trailing empty string,
	// but we might want to preserve exact lines. We'll leave it as is for simplicity.
	// If the string is empty, split returns [""] which is one line, except if it's really empty.
	if content1 == "" {
		lines1 = nil
	}
	if content2 == "" {
		lines2 = nil
	}

	script := Diff(lines1, lines2)
	differ := false
	for _, item := range script {
		if item.op != opEq {
			differ = true
			break
		}
	}

	hunks := BuildHunks(script, contextLines)
	return differ, hunks
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		return 2
	}

	jsonMode := flags.Has("json")
	files := flags.Positional
	
	contextLines := 3
	if flags.Has("U") {
		c, err := strconv.Atoi(flags.Get("U"))
		if err == nil && c >= 0 {
			contextLines = c
		}
	}

	if len(files) != 2 {
		common.RenderError("diff", 2, "USAGE", "missing operand", jsonMode, out)
		if !jsonMode {
			fmt.Fprintf(os.Stderr, "diff: missing operand\n")
		}
		return 2
	}

	b1, err := os.ReadFile(files[0])
	if err != nil {
		common.RenderError("diff", 2, "IO", err.Error(), jsonMode, out)
		if !jsonMode {
			fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		}
		return 2
	}
	b2, err := os.ReadFile(files[1])
	if err != nil {
		common.RenderError("diff", 2, "IO", err.Error(), jsonMode, out)
		if !jsonMode {
			fmt.Fprintf(os.Stderr, "diff: %v\n", err)
		}
		return 2
	}

	differ, hunks := GenerateDiff(string(b1), string(b2), contextLines)

	res := DiffResult{
		Files:  []string{files[0], files[1]},
		Differ: differ,
		Hunks:  hunks,
	}

	if differ {
		common.Render("diff", res, jsonMode, out, func() {
			if flags.Has("q") {
				fmt.Fprintf(out, "Files %s and %s differ\n", files[0], files[1])
			} else if flags.Has("u") || flags.Has("U") {
				fmt.Fprintf(out, "--- %s\n+++ %s\n", files[0], files[1])
				for _, h := range hunks {
					fmt.Fprintf(out, "@@ -%d,%d +%d,%d @@\n", h.OldStart, h.OldLines, h.NewStart, h.NewLines)
					for _, l := range h.Lines {
						fmt.Fprintln(out, l)
					}
				}
			} else {
				fmt.Fprintf(out, "Files %s and %s differ\n", files[0], files[1])
			}
		})
		return 1
	}

	common.Render("diff", res, jsonMode, out, func() {})
	return 0
}

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "diff",
		Usage: "Compare files line by line",
		Run:   run,
	})
}
