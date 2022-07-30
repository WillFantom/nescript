package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/willfantom/nescript"
)

type DockerProcess struct {
	dockerClient *docker.Client
	dockerConn   *types.HijackedResponse
	commandID    string
	stdoutBytes  bytes.Buffer
	stderrBytes  bytes.Buffer
	complete     chan error
}

func (p *DockerProcess) Kill() error {
	return fmt.Errorf("can not kill docker exec process")
}

func (p *DockerProcess) Signal(s os.Signal) error {
	return fmt.Errorf("can not signal docker exec process")
}

func (p *DockerProcess) Write(input string) error {
	if _, err := p.dockerConn.Conn.Write([]byte(input)); err != nil {
		return fmt.Errorf("failed to write to container exec stdin: %w", err)
	}
	return nil
}

func (p *DockerProcess) Result() (*nescript.Result, error) {
	defer p.Close()
	err := <-p.complete
	if err != nil {
		return nil, fmt.Errorf("failed to wait for docker process: %w", err)
	}
	res, err := p.dockerClient.ContainerExecInspect(context.Background(), p.commandID)
	if err != nil {
		return nil, fmt.Errorf("could not determine exit code: %w", err)
	}
	result := nescript.Result{
		StdOut: string(p.stdoutBytes.String()),
		StdErr: string(p.stderrBytes.String()),
	}
	result.ExitCode = res.ExitCode
	return &result, nil
}

func (p *DockerProcess) Close() {
	p.dockerConn.Close()
}
