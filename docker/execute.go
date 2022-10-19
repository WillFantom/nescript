package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/willfantom/nescript"
)

// Executor provides an ExecFunc that will start the script/cmd process in the
// docker container with the given container ID. An initialized docker client
// must also be passed for communication with the relevant docker engine.
// Optionally, a WorkDir may be set, setting the precess working directory (path
// should be in the context of the container's file system). This ExecFunc does
// not require that the cmd/script be converted to a string, so is Formatter
// agnostic.
func Executor(client *docker.Client, containerID, workdir string) nescript.ExecFunc {
	return func(c nescript.Cmd) (nescript.Process, error) {
		config := types.ExecConfig{
			Tty:          false,
			AttachStdin:  true,
			AttachStderr: true,
			AttachStdout: true,
			Env:          c.Env(),
			WorkingDir:   workdir,
			Cmd:          c.Raw(),
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
