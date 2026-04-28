package main

import (
	"fmt"
	"strings"
)

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
                // save the last trace!
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

	for i, j := 0, len(script)-1; i < j; i, j = i+1, j-1 {
		script[i], script[j] = script[j], script[i]
	}

	return script
}

func main() {
	script := Diff(strings.Split("a\nb\nc", "\n"), strings.Split("a\nx\nc", "\n"))
	for _, it := range script {
		var opStr string
		if it.op == opEq { opStr = " " }
		if it.op == opDel { opStr = "-" }
		if it.op == opIns { opStr = "+" }
		fmt.Printf("%s%s\n", opStr, it.text)
	}
}
