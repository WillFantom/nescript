package executive

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

// DockerProcess represents a single instance of the executed on a docker
// container. script. Note that scripts over Docker have some limitations not
// found on local processes. For example, DockerProcess does not support a
// signals.
type DockerProcess struct {
	OriginExecutable Executable

	dockerClient *docker.Client
	dockerConn   *types.HijackedResponse
	commandID    string
	stdoutBytes  bytes.Buffer
	stderrBytes  bytes.Buffer
	complete     chan error
}

func (p *DockerProcess) Kill() error {
	return fmt.Errorf("not availbale over docker")
}

func (p *DockerProcess) SigInt() error {
	return fmt.Errorf("not availbale over docker")

}

func (p *DockerProcess) Write(input string) error {
	if _, err := p.dockerConn.Conn.Write([]byte(input)); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	return nil
}

func (p *DockerProcess) Result() (*Result, error) {
	defer p.dockerConn.Close()
	err := <-p.complete
	if err != nil {
		return nil, fmt.Errorf("failed to wait for docker process: %w", err)
	}
	res, err := p.dockerClient.ContainerExecInspect(context.Background(), p.commandID)
	if err != nil {
		return nil, fmt.Errorf("could not determin exit code: %w", err)
	}
	result := Result{
		StdOut: string(p.stdoutBytes.String()),
		StdErr: string(p.stderrBytes.String()),
	}
	result.ExitCode = res.ExitCode
	return &result, nil
}
