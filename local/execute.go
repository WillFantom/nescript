package local

import (
	"fmt"
	"os/exec"

	"github.com/willfantom/nescript"
)

var (
	defaultSubcommand = []string{"sh", "-c"}
)

// Executor returns an exec func that can execute a NEScript locally. A
// subcommand can be provided for the script (e.g. ["sh", "-c"]) or if nil, the
// default will be used. Also a working directory can be set, that if left empty
// will be set to the current working directory.
func Executor(workdir string, subcommand []string) nescript.ExecFunc {
	if subcommand == nil {
		subcommand = defaultSubcommand
	}
	return func(s nescript.Script) (nescript.Process, error) {
		command := append(subcommand, s.Raw())
		process := LocalProcess{
			cmd: exec.Command(command[0], command[1:]...),
		}
		process.cmd.Env = s.Env()
		process.cmd.Dir = workdir
		process.cmd.Stdout = &process.stdoutBytes
		process.cmd.Stderr = &process.stderrBytes
		if stdin, err := process.cmd.StdinPipe(); err != nil {
			return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		} else {
			process.stdin = stdin
		}
		if err := process.cmd.Start(); err != nil || process.cmd.Process == nil {
			return nil, fmt.Errorf("process failed to start: %w", err)
		}
		return &process, nil
	}
}
