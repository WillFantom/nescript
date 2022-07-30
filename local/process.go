package local

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/willfantom/nescript"
)

// Process represents a single instance of the script running or completed on
// the local device.
type LocalProcess struct {
	cmd         *exec.Cmd
	stdin       io.Writer
	stdoutBytes bytes.Buffer
	stderrBytes bytes.Buffer
}

func (p *LocalProcess) Kill() error {
	if err := p.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}
	return nil
}

func (p *LocalProcess) Signal(s os.Signal) error {
	if err := p.cmd.Process.Signal(s); err != nil {
		return fmt.Errorf("failed to send signal to process: %w", err)
	}
	return nil
}

func (p *LocalProcess) Exited() bool {
	if process, err := os.FindProcess(p.cmd.Process.Pid); err != nil || process == nil {
		return true
	}
	return false
}

func (p *LocalProcess) Write(input string) error {
	if p.Exited() {
		return fmt.Errorf("can not write to stdin, process has exited")
	}
	if _, err := io.WriteString(p.stdin, input); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	return nil
}

func (p *LocalProcess) Result() (*nescript.Result, error) {
	if err := p.cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("failed to wait for process: %w", err)
		}
	}
	result := nescript.Result{
		StdOut: string(p.stdoutBytes.String()),
		StdErr: string(p.stderrBytes.String()),
	}
	result.ExitCode = p.cmd.ProcessState.ExitCode()
	if err := p.cmd.Process.Release(); err != nil {
		return nil, fmt.Errorf("failed to release to process resources: %w", err)
	}
	return &result, nil
}

func (p *LocalProcess) Close() {
	// nothing to close
}
