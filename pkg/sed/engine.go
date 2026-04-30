package sed

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type engineState struct {
	patSpace    string
	holdSpace   string
	lineNum     int
	isEOF       bool
	substituted bool
	skipRead    bool

	out          io.Writer
	suppress     bool
	files        []string
	fileIdx      int
	inPlace      bool
	currentFile  string
	tmpFile      *os.File
	outChanged   bool

	lastRegex    *regexp.Regexp

	pendingAppend []string
	pendingRead   []string

	addrState map[*Instruction]bool
	
	scanner *bufio.Scanner
}

func (e *engineState) printLine(s string) {
	fmt.Fprintln(e.out, s)
}

func (e *engineState) printLineRaw(s string) {
	fmt.Fprint(e.out, s)
}

func runEngine(insts []*Instruction, readers []string, suppress bool, inPlace bool, globalOut io.Writer) int {
	if len(readers) == 0 {
		readers = []string{"-"}
	}

	e := &engineState{
		files:     readers,
		suppress:  suppress,
		inPlace:   inPlace,
		addrState: make(map[*Instruction]bool),
		out:       globalOut,
	}

	exitCode := 0

	for i := 0; i < len(e.files); i++ {
		e.fileIdx = i
		path := e.files[i]
		e.currentFile = path

		var r io.Reader
		if path == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sed: %s: %v\n", path, err)
				exitCode = 1
				continue
			}
			r = f
			defer f.Close()
		}

		if e.inPlace && path != "-" {
			tmpPath := path + ".tmp"
			tmpFile, err := os.Create(tmpPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sed: %v\n", err)
				exitCode = 1
				continue
			}
			e.tmpFile = tmpFile
			e.out = tmpFile
		}

		e.scanner = bufio.NewScanner(r)
		
		var nextLine string
		var hasNext bool

		if e.scanner.Scan() {
			nextLine = e.scanner.Text()
			hasNext = true
		} else {
			// empty file
			if e.inPlace && e.tmpFile != nil {
				e.tmpFile.Close()
				os.Rename(e.tmpFile.Name(), path)
			}
			continue
		}

		for hasNext {
			if !e.skipRead {
				e.patSpace = nextLine
				e.lineNum++
			}
			e.skipRead = false

			if e.scanner.Scan() {
				nextLine = e.scanner.Text()
				hasNext = true
				e.isEOF = false
			} else {
				hasNext = false
				e.isEOF = true
			}

			// execute instructions
			if err := e.execBlock(insts); err != nil {
				if err.Error() == "quit" {
					if !e.suppress && !e.skipRead {
						e.printLine(e.patSpace)
					}
					return exitCode
				} else if err.Error() == "next" {
					// d or similar
				} else if err.Error() == "quit_no_print" {
					return exitCode
				} else {
					fmt.Fprintf(os.Stderr, "sed: %v\n", err)
					exitCode = 1
					break
				}
			} else {
				if !e.suppress {
					e.printLine(e.patSpace)
				}
			}
			
			// flush appends
			for _, text := range e.pendingAppend {
				e.printLine(text)
			}
			e.pendingAppend = nil
		}

		if e.inPlace && path != "-" && e.tmpFile != nil {
			e.tmpFile.Close()
			os.Rename(e.tmpFile.Name(), path)
		}
	}

	return exitCode
}

func (e *engineState) matchAddress(addr *Address) bool {
	if addr == nil {
		return false
	}
	if addr.Type == AddrLine {
		return e.lineNum == addr.Line
	}
	if addr.Type == AddrLast {
		return e.isEOF
	}
	if addr.Type == AddrRegexp {
		re := addr.Regexp
		if re == nil {
			re = e.lastRegex
		}
		if re != nil {
			e.lastRegex = re
			return re.MatchString(e.patSpace)
		}
	}
	return false
}

func (e *engineState) shouldRun(inst *Instruction) bool {
	if inst.Addr1 == nil {
		return true != inst.AddressInvert
	}
	if inst.Addr2 == nil {
		return e.matchAddress(inst.Addr1) != inst.AddressInvert
	}

	// Address ranges
	active := e.addrState[inst]
	if !active {
		if e.matchAddress(inst.Addr1) {
			active = true
			e.addrState[inst] = true
			// If addr2 matches right away (and it's not a +N address)
			// we don't deactivate it here because POSIX says address ranges are inclusive
			// but if it matches the SAME line, it closes immediately.
			// Actually POSIX says if addr2 matches on the SAME line, it still spans at least one line.
			if inst.Addr2.Type == AddrLine && inst.Addr2.Step > 0 {
				// GNU extension +N
				// we change Addr2 to be e.lineNum + step
				inst.Addr2 = &Address{Type: AddrLine, Line: e.lineNum + inst.Addr2.Step}
			}
		}
	}

	result := active

	if active {
		if e.matchAddress(inst.Addr2) {
			e.addrState[inst] = false // deactivate for next line
		}
	}

	return result != inst.AddressInvert
}

func (e *engineState) execBlock(insts []*Instruction) error {
	for i := 0; i < len(insts); i++ {
		inst := insts[i]
		if !e.shouldRun(inst) {
			continue
		}

		switch inst.Cmd {
		case '{':
			if err := e.execBlock(inst.Block); err != nil {
				return err
			}
		case 's':
			re := inst.Regexp
			if re == nil {
				re = e.lastRegex
			}
			if re != nil {
				e.lastRegex = re
				if re.MatchString(e.patSpace) {
					e.substituted = true
					
					// Replacement logic
					repl := inst.Repl
					
					if inst.Global {
						e.patSpace = re.ReplaceAllString(e.patSpace, repl)
					} else if inst.SubNum > 0 {
						// replace Nth occurence
						matches := re.FindAllStringIndex(e.patSpace, -1)
						if len(matches) >= inst.SubNum {
							loc := matches[inst.SubNum-1]
							replced := re.ReplaceAllString(e.patSpace[loc[0]:loc[1]], repl)
							e.patSpace = e.patSpace[:loc[0]] + replced + e.patSpace[loc[1]:]
						}
					} else {
						// replace first
						loc := re.FindStringIndex(e.patSpace)
						if loc != nil {
							replced := re.ReplaceAllString(e.patSpace[loc[0]:loc[1]], repl)
							e.patSpace = e.patSpace[:loc[0]] + replced + e.patSpace[loc[1]:]
						}
					}

					if inst.Print {
						e.printLine(e.patSpace)
					}
					if inst.WFile != "" {
						f, err := os.OpenFile(inst.WFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err == nil {
							fmt.Fprintln(f, e.patSpace)
							f.Close()
						}
					}
				}
			}
		case 'p':
			e.printLine(e.patSpace)
		case 'P':
			idx := strings.Index(e.patSpace, "\n")
			if idx >= 0 {
				e.printLineRaw(e.patSpace[:idx+1])
			} else {
				e.printLine(e.patSpace)
			}
		case 'd':
			return fmt.Errorf("next") // skip rest of commands and don't print
		case 'D':
			idx := strings.Index(e.patSpace, "\n")
			if idx >= 0 {
				e.patSpace = e.patSpace[idx+1:]
				// POSIX says if pattern space is not empty, start next cycle without reading a new line.
				// We can achieve this by rewinding to the start of the block and setting skipRead = true.
				e.skipRead = true
				i = -1 // restart block
				continue
			} else {
				return fmt.Errorf("next")
			}
		case 'n':
			// print current, read next
			if !e.suppress {
				e.printLine(e.patSpace)
			}
			if e.scanner.Scan() {
				e.patSpace = e.scanner.Text()
				e.lineNum++
			} else {
				return fmt.Errorf("quit_no_print")
			}
		case 'N':
			// append next line to pattern space
		case 'a':
			e.pendingAppend = append(e.pendingAppend, inst.Text)
		case 'i':
			e.printLine(inst.Text)
		case 'c':
			// delete pattern space, print text, skip rest of script
			e.printLine(inst.Text)
			return fmt.Errorf("next")
		case 'q':
			return fmt.Errorf("quit")
		case '=':
			fmt.Fprintln(e.out, e.lineNum)
		case 'h':
			e.holdSpace = e.patSpace
		case 'H':
			e.holdSpace += "\n" + e.patSpace
		case 'g':
			e.patSpace = e.holdSpace
		case 'G':
			e.patSpace += "\n" + e.holdSpace
		case 'x':
			e.patSpace, e.holdSpace = e.holdSpace, e.patSpace
		case 'b':
			if inst.Label == "" {
				return nil // jump to end of script, print pattern space
			}
			// need label resolution
		case 't':
			if e.substituted {
				e.substituted = false
				if inst.Label == "" {
					return nil // jump to end of script
				}
				// jump to label
			}
		case 'T':
			if !e.substituted {
				if inst.Label == "" {
					return nil // jump to end of script
				}
				// jump to label
			}
		}
	}
	return nil
}
