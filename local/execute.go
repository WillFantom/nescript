package local

import (
	"fmt"

	"github.com/willfantom/nescript"
)

// Executor returns an exec func that can execute a NEScript locally. A working
// directory can optionally be specified, where if not, the current working
// directory of the application is used. This ExecFunc does not require that the
// cmd/script be converted to a string, so is Formatter agnostic.
func Executor(workdir string) nescript.ExecFunc {
	return func(c nescript.Cmd) (nescript.Process, error) {
		command, err := c.OSCmd()
		if err != nil {
			return nil, err
		}
		process := LocalProcess{
			cmd: command,
		}
		process.cmd.Env = c.Env()
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
