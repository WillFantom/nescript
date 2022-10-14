package nescript

import (
	"fmt"
	"os/exec"
)

// ExecFunc is used to execute a script, in-turn creating a process. If the
// script fails to be executed for any reason, this function can also error.
type ExecFunc func(Script) (Process, error)

// Exec will call the given ExecFunc to execute the script. Returned will be the
// process that is created as a result of execution. An error is returned if the
// script fails to execute for any reason.
func (s Script) Exec(executor ExecFunc) (Process, error) {
	return executor(s)
}

// CompileExec will "compile" the script using the given data and the golang
// template system, then calls the given ExecFunc. Returned will be the process
// that is created as a result of execution. An error is returned if the script
// fails to execute for any reason.
func (s Script) CompileExec(executor ExecFunc) (Process, error) {
	cs, err := s.Compile()
	if err != nil {
		return nil, err
	}
	return cs.Exec(executor)
}

// OSCmd attempts to convert the script to an os.exec package Cmd. For this, a
// subcommand must be provided. For example, if the subcommand ["sh", "-c"] was
// provided, and the compiled script was `echo 'Hello, world!'“, the resulting
// command would be equivalent to: sh -c "echo 'Hello, world!'". Env vars given
// to the script are also preserved in the resulting Cmd.
func (s Script) OSCmd(subcommand []string) (*exec.Cmd, error) {
	if len(subcommand) == 0 {
		return nil, fmt.Errorf("no subcommand must was provided for conversion to exec.Cmd as required")
	}
	commandSlice := append(subcommand, s.raw)
	command := &exec.Cmd{}
	if len(commandSlice) < 1 {
		command = exec.Command(commandSlice[0])
	} else {
		command = exec.Command(commandSlice[0], commandSlice[1:]...)
	}
	command.Env = s.env
	return command, nil
}
