package executive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/crypto/ssh"
)

// Script is a "compiled" script template that can be executed. Unlike a
// Template, an executable has already been "compiled" (parsed through the
// golang templating engine).
type Script struct {
	OriginTemplate Template `json:"origin"`
	compiledScript string
	scriptPath     string

	env       []string
	args      []string
	workDir   string
	singleUse bool
}

var (
	// tempFilePatternRegExpString matches a string that is alphanumeric (with
	// hypens or underscores) and ends in an asterisk. This can then be used to
	// prefix files.
	tempFilePatternRegExpString = "^[-a-zA-Z0-9_]*.[*]$"
	tempFilePatternRegExp       *regexp.Regexp
	tempFilePattern             = "executive-*"

	singleUse = false
)

// NewScript takes a compiled Template to generate a file that can be executed,
// running the script. If none of the Template features are needed, this can be
// used directly with a nil originScript.
func NewScript(originScript *Template, compiled string) (*Script, error) {
	script := Script{
		OriginTemplate: *originScript,
		scriptPath:     "",
		compiledScript: compiled,
		env:            make([]string, 0),
		args:           make([]string, 0),
		workDir:        "",
		singleUse:      false,
	}
	scriptFile, err := os.CreateTemp("", tempFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for script: %w", err)
	}
	script.scriptPath = scriptFile.Name()
	if _, err := scriptFile.Write([]byte(compiled)); err != nil {
		return nil, fmt.Errorf("could write compiled script to temp file: %w", err)
	}
	if err := scriptFile.Chmod(0555); err != nil {
		return nil, fmt.Errorf("could not make temp script file script: %w", err)
	}
	return &script, nil
}

// NewScriptFromFile is similar to creating a template from file, however this
// assumes the template is already "compiled" and so does not need to be parsed
// by the template system.
func NewScriptFromFile(path string) (*Script, error) {
	t, err := NewTemplateFromFile("", path)
	if err != nil {
		return nil, fmt.Errorf("failed to create supporting template: %w", err)
	}
	return NewScript(t, t.Raw)
}

// NewScriptFromHTTP is similar to creating a template from a http endpoint,
// however this assumes the template is already "compiled" and so does not need
// to be parsed by the template system.
func NewScriptFromHTTP(link string) (*Script, error) {
	t, err := NewTemplateFromHTTP("", link)
	if err != nil {
		return nil, fmt.Errorf("failed to create supporting template: %w", err)
	}
	return NewScript(t, t.Raw)
}

// WithEnv sets an environment variable that should be set within the script's
// runtime env. Script is returned and not directly modified as to support
// function chains.
func (e Script) WithEnv(key, value string) Script {
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, fmt.Sprintf("%s=%s", key, value))
	return e
}

// WithArg adds an argument that will be provided to the script upon execution.
// Args are provided to the script in the the order they are provided. Script is
// returned and not directly modified as to support function chains.
func (e Script) WithArg(arg string) Script {
	if e.args == nil {
		e.args = make([]string, 0)
	}
	e.args = append(e.args, arg)
	return e
}

// WithOSEnv ensures that the script will be able to use the environment vars
// from the os env. Script is returned and not directly modified as to support
// function chains.
func (e Script) WithOSEnv() Script {
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, os.Environ()...)
	return e
}

// WithWorkDir sets the working directory for the context of the script. Script
// is returned and not directly modified as to support function chains.
func (e Script) WithWorkDir(path string) Script {
	e.workDir = path
	return e
}

// SingleUse sets the temporary file associated with the script to remove itself
// immediately after it starts execution. This can help keep things neat, but
// can also cause issues. Use carefully. Script is returned to facilitate
// function chinaing.
func (e Script) SingleUse() Script {
	e.singleUse = true
	return e
}

// SetEnv sets an environment variable that should be set within the script's
// runtime env.
func (e *Script) SetEnv(key, value string) error {
	if key == "" {
		return fmt.Errorf("key must not be empty")
	}
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, fmt.Sprintf("%s=%s", key, value))
	return nil
}

// SetArgs defines all the args that will be provided to the script upon
// execution. This overwrites any changes made via WithArg.
func (e *Script) SetArgs(args []string) error {
	if args == nil {
		return fmt.Errorf("arguments must not be empty")
	}
	for _, a := range args {
		if a == "" {
			return fmt.Errorf("argument values must not be empty")
		}
	}
	e.args = args
	return nil
}

// SetOSEnv ensures that the script will be able to use the environment vars
// from the os env.
func (e *Script) SetOSEnv() {
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, os.Environ()...)
}

// SetWorkDir sets the working directory for the context of the script. This
// does not check if the directory exists in order to be compatible with
// non-local execution.
func (e *Script) SetWorkDir(path string) error {
	// if info, err := os.Stat(path); err != nil {
	// 	return fmt.Errorf("failed to get directory info: %w", err)
	// } else if !info.IsDir() {
	// 	return fmt.Errorf("working directory must be a directory")
	// }
	e.workDir = path
	return nil
}

// SetSingleUse sets the temporary file associated with the script to remove
// itself immediately after it starts execution. This can help keep things neat,
// but can also cause issues. Use carefully. Script is returned to
// facilitate function chinaing.
func (e *Script) SetSingleUse(isSingleUse bool) {
	e.singleUse = isSingleUse
}

// Execute simply runs the "compiled" script with the env and workdir given. If
// the process returned is nil the process failed to start. If a single use
// script fails to be removed, the process is still returned along with the
// error.
func (e Script) Execute() (Process, error) {
	process := LocalProcess{
		OriginExecutable: e,
		cmd:              exec.Command(e.scriptPath),
	}
	if len(e.env) > 0 {
		process.cmd.Env = e.env
	}
	if e.workDir != "" {
		process.cmd.Dir = e.workDir
	}
	process.cmd.Stdout = &process.stdoutBytes
	process.cmd.Stderr = &process.stderrBytes
	if stdin, err := process.cmd.StdinPipe(); err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	} else {
		process.stdin = stdin
	}
	if err := process.cmd.Start(); err != nil || process.cmd.Process == nil {
		return nil, fmt.Errorf("process failed to start: %w", err)
	}
	if e.singleUse {
		if err := e.Remove(); err != nil {
			return &process, fmt.Errorf("failed to remove single use script file: %w", err)
		}
	}
	return &process, nil
}

// ExecuteOverSSH simply runs the "compiled" script with the env vars provided.
// To run over SSH, some configuration must be supplied, including the remote
// user, the target ssh server address (and port), and an SSH auth method (such
// as a password). If the process returned is nil the process failed to start.
// If a single use script fails to be removed, the process is still returned
// along with the error. Note that WorkDir is not supported for SSH.
func (e Script) ExecuteOverSSH(user, addr string, auth ssh.AuthMethod) (Process, error) {
	process := SSHProcess{
		OriginExecutable: e,
	}
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to make ssh connection: %w", err)
	}
	process.sshClient = client
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to make ssh session: %w", err)
	}
	process.sshSession = session
	for _, kv := range e.env {
		kvPair := strings.Split(kv, "=")
		if len(kvPair) == 2 {
			if err := session.Setenv(kvPair[0], kvPair[1]); err != nil {
				session.Close()
				client.Close()
				return nil, fmt.Errorf("failed to set env %s: %w", kv, err)
			}
		}
	}
	session.Stdout = &process.stdoutBytes
	session.Stderr = &process.stderrBytes
	if process.stdin, err = session.StdinPipe(); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to connect to stdin: %w", err)
	}
	if err := session.Start(e.compiledScript); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to start script: %w", err)
	}
	return &process, err
}

// ExecuteOverDocker simply runs the "compiled" script with the env vars and
// workdir provided. To run over Docker, some configuration must be supplied,
// including a docker client linked to the releveant docker host, the container
// ID that the script should be executed on, and a subprocess command to execute
// the script. In many cases, the subprocess command 'sh -c' will suffice, but
// others such as '/bin/bash -c' can also be used. The limitations here come
// from interrupts: docker processes will not support them.
func (e Script) ExecuteOverDocker(client *docker.Client, containerID string, subprocessCmd string) (Process, error) {
	process := DockerProcess{
		OriginExecutable: e,
		dockerClient:     client,
		complete:         make(chan error),
	}
	config := types.ExecConfig{
		Tty:          false,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Env:          e.env,
		Cmd:          append(strings.Fields(subprocessCmd), e.compiledScript),
	}
	if e.workDir != "" {
		config.WorkingDir = e.workDir
	}
	idResponse, err := client.ContainerExecCreate(context.Background(), containerID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker exec: %w", err)
	}
	execID := idResponse.ID
	process.commandID = execID
	if conn, err := client.ContainerExecAttach(context.Background(), execID, types.ExecStartCheck{}); err != nil {
		return nil, fmt.Errorf("failed to attach to docker exec: %w", err)
	} else {
		process.dockerConn = &conn
		go func() {
			_, err := stdcopy.StdCopy(&process.stdoutBytes, &process.stderrBytes, conn.Reader)
			process.complete <- err
		}()
	}
	if err := client.ContainerExecStart(context.Background(), execID, types.ExecStartCheck{}); err != nil {
		return nil, fmt.Errorf("failed to start docker exec: %w", err)
	}
	return &process, nil
}

// Remove simply deletes the temporary file that is generated upon creation of
// an Script. Doing so is not vital as the file is stored in a temporary dir,
// but it can help long running programs that frequently use executive.
func (e *Script) Remove() error {
	if !e.ScriptPathExists() {
		return nil
	}
	if err := os.Remove(e.scriptPath); err != nil {
		return fmt.Errorf("failed to remove script file: %w", err)
	}
	return nil
}

// ScriptPath simply returns the path containing the temporary file of the
// script.
func (e Script) ScriptPath() string {
	return e.scriptPath
}

// ScriptPathExists simply returns a bool representing the existence of the
// temporary script file (without discoling errors). THis can be used to
// determine if the executable has been removed. In the case where Stat returns
// an error that is not ErrNotExist, this assumes the file does exist as to
// allow for remove to operate more optimistically.
func (e Script) ScriptPathExists() bool {
	if fileInfo, err := os.Stat(e.scriptPath); errors.Is(err, os.ErrNotExist) || fileInfo.IsDir() {
		return false
	}
	return true
}

// SetTempFilePrefix defines the prefix pattern for the temporary files created
// when "compiling" scripts into "executables". Any prefix must be only
// alphanumeric, with hypens and underscores permitted. It also must end in a
// single asterisk.
func SetTempFilePrefix(prefixPattern string) error {
	if !tempFilePatternRegExp.MatchString(prefixPattern) {
		return fmt.Errorf("prefix pattern is not valid")
	}
	tempFilePattern = prefixPattern
	return nil
}

func init() {
	if compiledRegexp, err := regexp.Compile(tempFilePatternRegExpString); err != nil {
		panic(err)
	} else {
		tempFilePatternRegExp = compiledRegexp
	}
}
