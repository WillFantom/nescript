# `ExecFunc`: Docker 🐳

This allows for executing nescript Cmds and Scripts on Docker container targets denoted by their container ID. For this, an initialized docker client must also be provided.

There are some quirks when using the Docker `ExecFunc`:
 - Any subprocess spawned by a Cmd, or any Script executed will have access to the containers Env vars by default.
 - Metrics can not be obtained, specifically timing data as is possible with local script execution.

## Example

```go
containerID := "abcdef0123"
dockerClient, err := client.NewClientWithOpts(client.FromEnv)
if err != nil {
	panic(err)
}
dockerExecutor := docker.Executor(dockerClient, containerID, "")
```
