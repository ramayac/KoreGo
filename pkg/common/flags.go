// Package common provides shared utilities for korego utilities.
package common

import (
	"fmt"
	"strings"
)

// FlagType enumerates the kinds of values a flag can hold.
type FlagType int

const (
	FlagBool  FlagType = iota // -l, --all
	FlagValue                 // --key=value
)

// FlagDef describes a single accepted flag.
type FlagDef struct {
	Short string   // single character, e.g. "l"
	Long  string   // long name without --, e.g. "all"
	Type  FlagType // Bool or Value
}

// FlagSpec is the set of accepted flags for a command.
type FlagSpec struct {
	Defs []FlagDef
}

// FlagError is returned when flag parsing fails.
type FlagError struct {
	ExitCode int
	Msg      string
}

func (e *FlagError) Error() string { return e.Msg }

// ParseResult holds the parsed flags and positional arguments.
type ParseResult struct {
	// Bools tracks bool flags that were set, keyed by short OR long name.
	Bools map[string]bool
	// Values holds key=value style flags.
	Values map[string]string
	// Count tracks how many times a flag was repeated (for -vvv style).
	Count map[string]int
	// Positional holds non-flag arguments.
	Positional []string
	// Stdin is true if bare "-" was present.
	Stdin bool
}

func newParseResult() *ParseResult {
	return &ParseResult{
		Bools:  make(map[string]bool),
		Values: make(map[string]string),
		Count:  make(map[string]int),
	}
}

// Has returns true if the short or long flag name was set.
func (r *ParseResult) Has(name string) bool {
	return r.Bools[name]
}

// Get returns the value for a value-type flag.
func (r *ParseResult) Get(name string) string {
	return r.Values[name]
}

// ParseFlags parses args according to spec.
// Unknown flags return *FlagError with ExitCode 2.
func ParseFlags(args []string, spec FlagSpec) (*ParseResult, error) {
	// Build lookup maps for quick validation.
	shortDef := make(map[string]FlagDef)
	longDef := make(map[string]FlagDef)
	for _, d := range spec.Defs {
		if d.Short != "" {
			shortDef[d.Short] = d
		}
		if d.Long != "" {
			longDef[d.Long] = d
		}
	}

	res := newParseResult()
	i := 0
	for i < len(args) {
		arg := args[i]

		// End of flags marker.
		if arg == "--" {
			res.Positional = append(res.Positional, args[i+1:]...)
			break
		}

		// Stdin marker.
		if arg == "-" {
			res.Stdin = true
			res.Positional = append(res.Positional, "-")
			i++
			continue
		}

		// Long flag: --name or --name=value.
		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			var value string
			hasEq := false
			if idx := strings.IndexByte(name, '='); idx != -1 {
				value = name[idx+1:]
				name = name[:idx]
				hasEq = true
			}
			def, ok := longDef[name]
			if !ok {
				return nil, &FlagError{ExitCode: 2, Msg: fmt.Sprintf("unknown flag: --%s", name)}
			}
			// Canonicalise: store under both short and long.
			if def.Type == FlagValue {
				if !hasEq {
					// --key value form: consume next arg.
					if i+1 >= len(args) {
						return nil, &FlagError{ExitCode: 2, Msg: fmt.Sprintf("flag --%s requires a value", name)}
					}
					i++
					value = args[i]
				}
				if def.Short != "" {
					res.Values[def.Short] = value
				}
				res.Values[name] = value
			} else {
				if def.Short != "" {
					res.Bools[def.Short] = true
					res.Count[def.Short]++
				}
				res.Bools[name] = true
				res.Count[name]++
			}
			i++
			continue
		}

		// Short flag(s): -laR or -v (may repeat).
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			chars := arg[1:]
			for ci, ch := range chars {
				key := string(ch)
				def, ok := shortDef[key]
				if !ok {
					return nil, &FlagError{ExitCode: 2, Msg: fmt.Sprintf("unknown flag: -%s", key)}
				}
				if def.Type == FlagValue {
					// Remainder of the cluster is the value, or next arg.
					remainder := chars[ci+1:]
					if remainder != "" {
						res.Values[key] = remainder
						if def.Long != "" {
							res.Values[def.Long] = remainder
						}
					} else {
						if i+1 >= len(args) {
							return nil, &FlagError{ExitCode: 2, Msg: fmt.Sprintf("flag -%s requires a value", key)}
						}
						i++
						res.Values[key] = args[i]
						if def.Long != "" {
							res.Values[def.Long] = args[i]
						}
					}
					break // value flags always consume the rest of the cluster.
				}
				res.Bools[key] = true
				res.Count[key]++
				if def.Long != "" {
					res.Bools[def.Long] = true
					res.Count[def.Long]++
				}
			}
			i++
			continue
		}

		// Positional argument.
		res.Positional = append(res.Positional, arg)
		i++
	}

	return res, nil
}
