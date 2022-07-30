package local

import (
	"fmt"
	"os/exec"

	"github.com/willfantom/nescript"
)

type LocalExecutor struct {
	subCommand []string
	workdir    string
}

var (
	defaultSubCmd = []string{"sh", "-c"}
)

func NewExecutor() *LocalExecutor {
	return &LocalExecutor{
		subCommand: defaultSubCmd,
		workdir:    "",
	}
}

func (le LocalExecutor) ExecFunc() (nescript.ExecFunc, error) {
	if len(le.subCommand) == 0 {
		return nil, fmt.Errorf("no sub-command for script execution was provided")
	}

	return func(s nescript.Script) (nescript.Process, error) {
		command := append(le.subCommand, s.Raw())
		process := LocalProcess{
			cmd: exec.Command(command[0], command[1:]...),
		}
		process.cmd.Env = s.Env()
		process.cmd.Dir = le.workdir
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
	}, nil
}
