# `ExecFunc`: SSH 🧑‍💻

This allows for executing nescript Cmds and Scripts on remote SSH targets.

There are some quirks when using the SSH `ExecFunc`:
 - Env vars can only be used if the SSH server allows for it (e.g. by having a wildcard `AcceptEnv`).
 - Scripts and subprocess spawned from commands will have access to the systems Env vars by default.
 - Metrics can not be obtained, specifically timing data as is possible with local script execution.

## Example

```go
target := "10.0.0.1:22"
config := &ssh.ClientConfig{
	User: "root",
	Auth: []ssh.AuthMethod{
		ssh.Password("password"),
  },
  HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}
sshExecutor := sshe.Executor(target, config)
```
