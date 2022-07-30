package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/willfantom/nescript"
)

type DockerExecutor struct {
	subCommand   []string
	dockerClient *docker.Client
	containerID  string
	workDir      string
}

var (
	defaultSubCmd = []string{"sh", "-c"}
)

func NewExecutor(client *docker.Client, containerID string) *DockerExecutor {
	return &DockerExecutor{
		subCommand:   defaultSubCmd,
		dockerClient: client,
		containerID:  containerID,
	}
}

func (de DockerExecutor) ExecFunc() (nescript.ExecFunc, error) {
	if len(de.subCommand) == 0 {
		return nil, fmt.Errorf("no sub-command for script execution was provided")
	}

	return func(s nescript.Script) (nescript.Process, error) {
		config := types.ExecConfig{
			Tty:          false,
			AttachStdin:  true,
			AttachStderr: true,
			AttachStdout: true,
			Env:          s.Env(),
			WorkingDir:   de.workDir,
			Cmd:          append(de.subCommand, s.Raw()),
		}
		idResponse, err := de.dockerClient.ContainerExecCreate(context.Background(), de.containerID, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create docker exec in container '%s': %w", de.containerID, err)
		}
		process := DockerProcess{
			dockerClient: de.dockerClient,
			commandID:    idResponse.ID,
			complete:     make(chan error),
		}
		if conn, err := de.dockerClient.ContainerExecAttach(context.Background(), process.commandID, types.ExecStartCheck{}); err != nil {
			return nil, fmt.Errorf("failed to attach to docker exec: %w", err)
		} else {
			process.dockerConn = &conn
			go func() {
				_, err := stdcopy.StdCopy(&process.stdoutBytes, &process.stderrBytes, conn.Reader)
				process.complete <- err
			}()
		}
		if err := de.dockerClient.ContainerExecStart(context.Background(), process.commandID, types.ExecStartCheck{}); err != nil {
			return nil, fmt.Errorf("failed to start docker exec: %w", err)
		}
		return &process, nil

	}, nil
}
