// Package dispatch implements the multicall command registry for CoreGoLinux.
package dispatch

import (
	"fmt"
	"os"
	"sort"
)

// Command is a single registered utility.
type Command struct {
	Name  string
	Usage string
	Run   func(args []string) int // returns POSIX exit code
}

var registry = map[string]Command{}

// Register adds a command to the global registry.
// It panics if a command with the same name is already registered.
func Register(cmd Command) {
	if _, exists := registry[cmd.Name]; exists {
		panic(fmt.Sprintf("dispatch: duplicate command registration: %q", cmd.Name))
	}
	registry[cmd.Name] = cmd
}

// Lookup returns the named command and whether it was found.
func Lookup(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}

// ListAll returns all registered commands sorted by name.
func ListAll() []Command {
	out := make([]Command, 0, len(registry))
	for _, c := range registry {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// PrintHelp writes the help listing to stdout.
func PrintHelp(binName string) {
	fmt.Fprintf(os.Stdout, "Usage: %s <command> [args]\n\nCommands:\n", binName)
	for _, c := range ListAll() {
		fmt.Fprintf(os.Stdout, "  %-14s %s\n", c.Name, c.Usage)
	}
}
