package ssh

import (
	"fmt"
	"strings"

	"github.com/willfantom/nescript"
	"golang.org/x/crypto/ssh"
)

var (
	defaultSubcommand = []string{"sh", "-c"}
)

func Executor(target string, config *ssh.ClientConfig, subcommand []string) nescript.ExecFunc {
	if subcommand == nil {
		subcommand = defaultSubcommand
	}
	return func(s nescript.Script) (nescript.Process, error) {
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
		for _, e := range s.Env() {
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
		if err := sshSession.Start(fmt.Sprintf("%s '%s'", strings.Join(subcommand, " "), s.Raw())); err != nil {
			process.Close()
			return nil, fmt.Errorf("process failed to start: %w", err)
		}
		return &process, nil
	}
}
