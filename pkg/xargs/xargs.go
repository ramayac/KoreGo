package xargs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ramayac/korego/internal/dispatch"
	"github.com/ramayac/korego/pkg/common"
)

var spec = common.FlagSpec{
	Defs: []common.FlagDef{
		{Short: "j", Long: "json", Type: common.FlagBool},
	},
}

type ExecResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exitCode"`
}

func run(args []string, out io.Writer) int {
	flags, err := common.ParseFlags(args, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "xargs: %v\n", err)
		return 1
	}

	baseCmd := "echo"
	if len(flags.Positional) > 0 {
		baseCmd = flags.Positional[0]
	}

	cmdArgs := []string{}
	if len(flags.Positional) > 1 {
		cmdArgs = flags.Positional[1:]
	}

	scanner := bufio.NewScanner(os.Stdin)
	var words []string
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	cmdArgs = append(cmdArgs, words...)

	cmd := exec.Command(baseCmd, cmdArgs...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
	}

	res := []ExecResult{
		{
			Command:  baseCmd,
			ExitCode: exitCode,
		},
	}

	if flags.Has("j") {
		common.Render("xargs", res, true, out, func() {})
	}

	return exitCode
}

func init() {
	dispatch.Register(dispatch.Command{Name: "xargs", Usage: "Build and execute command lines from standard input", Run: run})
}
