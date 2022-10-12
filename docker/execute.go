package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/willfantom/nescript"
)

var (
	defaultSubcommand = []string{"sh", "-c"}
)

func Executor(subcommand []string, client *docker.Client, workdir, containerID string) nescript.ExecFunc {
	if subcommand == nil {
		subcommand = defaultSubcommand
	}
	return func(s nescript.Script) (nescript.Process, error) {
		config := types.ExecConfig{
			Tty:          false,
			AttachStdin:  true,
			AttachStderr: true,
			AttachStdout: true,
			Env:          s.Env(),
			WorkingDir:   workdir,
			Cmd:          append(subcommand, s.Raw()),
		}
		idResponse, err := client.ContainerExecCreate(context.Background(), containerID, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create docker exec in container '%s': %w", containerID, err)
		}
		process := DockerProcess{
			dockerClient: client,
			commandID:    idResponse.ID,
			complete:     make(chan error),
		}
		if conn, err := client.ContainerExecAttach(context.Background(), process.commandID, types.ExecStartCheck{}); err != nil {
			return nil, fmt.Errorf("failed to attach to docker exec: %w", err)
		} else {
			process.dockerConn = &conn
			go func() {
				_, err := stdcopy.StdCopy(&process.stdoutBytes, &process.stderrBytes, conn.Reader)
				process.complete <- err
			}()
		}
		if err := client.ContainerExecStart(context.Background(), process.commandID, types.ExecStartCheck{}); err != nil {
			return nil, fmt.Errorf("failed to start docker exec: %w", err)
		}
		return &process, nil

	}
}
