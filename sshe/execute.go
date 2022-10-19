package sshe

import (
	"fmt"
	"strings"

	"github.com/willfantom/nescript"
	"golang.org/x/crypto/ssh"
)

// Executor provides an ExecFunc that will start the script/cmd process on an
// SSH target. The target must be provided in the form of ip:port along with the
// ssh client config, specifing factors such as HostKeyAuth and the Auth method.
// As executing a command over SSH must be done by passing a single string, this
// ExecFunc will convert the given cmd/script to a string, thus this will use
// the formatter associated with the cmd/script.
func Executor(target string, config *ssh.ClientConfig) nescript.ExecFunc {
	return func(c nescript.Cmd) (nescript.Process, error) {
		process := SSHProcess{}
		sshClient, err := ssh.Dial("tcp", target, config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to ssh target '%s': %w", target, err)
		}
		process.sshClient = sshClient
		sshSession, err := sshClient.NewSession()
		if err != nil {
			sshClient.Close()
			return nil, fmt.Errorf("failed to create ssh session on target '%s': %w", target, err)
		}
		process.sshSession = sshSession
		for _, e := range c.Env() {
			envVar := strings.Split(e, "=")
			if len(envVar) != 2 {
				process.Close()
				return nil, fmt.Errorf("invalid env var '%s'", e)
			}
			if err := sshSession.Setenv(envVar[0], envVar[1]); err != nil {
				process.Close()
				return nil, fmt.Errorf("failed to set env var '%s': %w", e, err)
			}
		}
		sshSession.Stdout = &process.stdoutBytes
		sshSession.Stderr = &process.stderrBytes
		if stdin, err := sshSession.StdinPipe(); err != nil {
			return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		} else {
			process.stdin = stdin
		}
		if err := sshSession.Start(c.String()); err != nil {
			process.Close()
			return nil, fmt.Errorf("process failed to start: %w", err)
		}
		return &process, nil
	}
}
