package executive

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

// SSHProcess represents a single instance of the remotely executed (ssh)
// script. Note that scripts over SSH have some limitations not found on local
// processes. For example, SSHProcess does not support a WorkDir.
type SSHProcess struct {
	OriginExecutable Executable

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

func (p *SSHProcess) SigInt() error {
	if err := p.sshSession.Signal(ssh.SIGINT); err != nil {
		return fmt.Errorf("failed to send sigint to process: %w", err)
	}
	return nil
}

func (p *SSHProcess) Write(input string) error {
	if _, err := io.WriteString(p.stdin, input); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	return nil
}

func (p *SSHProcess) Result() (*Result, error) {
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
	result := Result{
		StdOut: string(p.stdoutBytes.String()),
		StdErr: string(p.stderrBytes.String()),
	}
	result.ExitCode = exitCode
	return &result, nil
}
