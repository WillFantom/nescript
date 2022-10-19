package nescript

import (
	"os/exec"
)

// ExecFunc is used to execute a script, in-turn creating a process. If the
// script fails to be executed for any reason, this function can also error.
type ExecFunc func(Cmd) (Process, error)

// Subcommand is the actual command to be executed by a given ExecFunc. The
// remaining script is passed as the final argument to the subcommand, for
// example a script to be interpreted/executed by the shell ["sh", "-c"] would
// be the appropriate subcommand.
type Subcommand []string

var (
	SCBinary  Subcommand = nil
	SCShell   Subcommand = []string{"sh", "-c"}
	SCBash    Subcommand = []string{"bash", "-c"}
	SCAsh     Subcommand = []string{"ash", "-c"}
	SCPython  Subcommand = []string{"python", "-c"}
	SCPython3 Subcommand = []string{"python3", "-c"}
)

// Exec will call the given ExecFunc to execute the script. Returned will be the
// process that is created as a result of execution. An error is returned if the
// script fails to execute for any reason.
func (c Cmd) Exec(executor ExecFunc) (Process, error) {
	return executor(c)
}

// CompileExec will "compile" the script using the given data and the golang
// template system, then calls the given ExecFunc. Returned will be the process
// that is created as a result of execution. An error is returned if the script
// fails to execute for any reason.
func (c Cmd) CompileExec(executor ExecFunc) (Process, error) {
	cs, err := c.Compile()
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
func (c Cmd) OSCmd() (*exec.Cmd, error) {
	commandSlice := c.Raw()
	var command *exec.Cmd
	if len(commandSlice) <= 0 {
		command = exec.Command("")
	} else if len(commandSlice) == 1 {
		command = exec.Command(commandSlice[0])
	} else {
		command = exec.Command(commandSlice[0], commandSlice[1:]...)
	}
	command.Env = c.env
	return command, nil
}

// String uses the command's formatter to convert the raw command (a main
// executable path and a slice of arguments) into a single string.
func (c Cmd) String() string {
	return c.formatter(c.Raw())
}
