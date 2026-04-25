// Package truefalse implements the POSIX true and false utilities.
// They are combined here because both are trivially simple.
package truefalse

import (
	"io"

	"github.com/ramayac/korego/internal/dispatch"
)

func runTrue(_ []string, _ io.Writer) int  { return 0 }
func runFalse(_ []string, _ io.Writer) int { return 1 }

func init() {
	dispatch.Register(dispatch.Command{
		Name:  "true",
		Usage: "Return true (exit 0)",
		Run:   runTrue,
	})
	dispatch.Register(dispatch.Command{
		Name:  "false",
		Usage: "Return false (exit 1)",
		Run:   runFalse,
	})
}
