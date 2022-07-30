package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/willfantom/nescript"
	"golang.org/x/crypto/ssh"
)

// Process represents a single instance of the script running or completed on
// the local device.
type SSHProcess struct {
	sshSession  *ssh.Session
	sshClient   *ssh.Client
	stdin       io.Writer
	stdoutBytes bytes.Buffer
	stderrBytes bytes.Buffer
}

func (p *SSHProcess) Kill() error {
	if err := p.sshSession.Signal(ssh.SIGKILL); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}
	return nil
}

func (p *SSHProcess) Signal(s os.Signal) error {
	if err := p.sshSession.Signal(ssh.Signal(s.String())); err != nil {
		return fmt.Errorf("failed to send signal to process: %w", err)
	}
	return nil
}

func (p *SSHProcess) Write(input string) error {
	if _, err := io.WriteString(p.stdin, input); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	return nil
}

func (p *SSHProcess) Result() (*nescript.Result, error) {
	defer p.sshSession.Close()
	defer p.sshClient.Close()
	exitCode := 0
	if err := p.sshSession.Wait(); err != nil {
		if eerr, ok := err.(*ssh.ExitError); !ok {
			return nil, fmt.Errorf("failed to wait for ssh process: %w", err)
		} else {
			exitCode = eerr.ExitStatus()
		}
	}
	result := nescript.Result{
		StdOut: string(p.stdoutBytes.String()),
		StdErr: string(p.stderrBytes.String()),
	}
	result.ExitCode = exitCode
	return &result, nil
}

func (p *SSHProcess) Close() {
	p.sshSession.Close()
	p.sshClient.Close()
}
