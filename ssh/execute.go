package ssh

import (
	"fmt"
	"strings"

	"github.com/willfantom/nescript"
	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	subCommand []string
	target     string
	config     *ssh.ClientConfig
	sanitizer  func([]string, string) string
}

var (
	defaultSubCmd = []string{"sh", "-c"}
)

func NewExecutor(target string, c *ssh.ClientConfig) *SSHExecutor {
	return &SSHExecutor{
		subCommand: defaultSubCmd,
		target:     target,
		config:     c,
		sanitizer: func(sub []string, script string) string {
			return fmt.Sprintf("%s '%s'", strings.Join(sub, " "), script)
		},
	}
}

func (sshe SSHExecutor) ExecFunc() (nescript.ExecFunc, error) {
	if len(sshe.subCommand) == 0 {
		return nil, fmt.Errorf("no sub-command for script execution was provided")
	}

	return func(s nescript.Script) (nescript.Process, error) {
		process := SSHProcess{}
		sshClient, err := ssh.Dial("tcp", sshe.target, sshe.config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to ssh target '%s': %w", sshe.target, err)
		}
		process.sshClient = sshClient
		sshSession, err := sshClient.NewSession()
		if err != nil {
			sshClient.Close()
			return nil, fmt.Errorf("failed to create ssh session on targer '%s': %w", sshe.target, err)
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

		if err := sshSession.Start(sshe.sanitizer(sshe.subCommand, s.Raw())); err != nil {
			process.Close()
			return nil, fmt.Errorf("process failed to start: %w", err)
		}
		return &process, nil
	}, nil
}
