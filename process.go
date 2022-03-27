package executive

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

//Process represents a single instance of the script.
type Process struct {
	OriginExecutable Executable

	cmd    *exec.Cmd
	stdin  io.Writer
	stdout io.Reader
	stderr io.Reader
}

func (p *Process) Kill() error {
	if !p.Exited() {
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	return nil
}

func (p *Process) SigInt() error {
	if !p.Exited() {
		if err := p.cmd.Process.Signal(syscall.SIGINT); err != nil {
			return fmt.Errorf("failed to send sigint to process: %w", err)
		}
	}
	return nil
}

func (p *Process) Exited() bool {
	if process, err := os.FindProcess(p.cmd.Process.Pid); err != nil || process == nil {
		return true
	}
	return false
}

func (p *Process) Write(input string) error {
	if p.Exited() {
		return fmt.Errorf("can not write to stdin, process has exited")
	}
	if _, err := io.WriteString(p.stdin, input); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	return nil
}

func (p *Process) ReadOut() io.Reader {
	return p.stdout
}

func (p *Process) ReadErr() io.Reader {
	return p.stderr
}

func (p *Process) Result() (*Result, error) {
	result := Result{}
	if out, err := io.ReadAll(p.stdout); err != nil {
		return nil, fmt.Errorf("failed to read std out: %w", err)
	} else {
		result.StdOut = string(out)
	}
	if stdErr, err := io.ReadAll(p.stderr); err != nil {
		return nil, fmt.Errorf("failed to read std err: %w", err)
	} else {
		result.StdErr = string(stdErr)
	}
	if err := p.cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("failed to wait for process: %w", err)
		}
	}
	result.ExitCode = p.cmd.ProcessState.ExitCode()
	result.SystemTime = p.cmd.ProcessState.SystemTime()
	result.UserTime = p.cmd.ProcessState.UserTime()
	result.TotalTime = result.SystemTime + result.UserTime
	if err := p.cmd.Process.Release(); err != nil {
		return nil, fmt.Errorf("failed to release to process resources: %w", err)
	}
	return &result, nil
}
