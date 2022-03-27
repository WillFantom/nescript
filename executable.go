package executive

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

//Executable is a "compiled" script that can be executed. Unlike a Script, an
//executable has already been "compiled" (parsed through the golang templating
//engine).
type Executable struct {
	OriginScript Script `json:"origin"`
	scriptPath   string

	env       []string
	args      []string
	workDir   string
	singleUse bool
}

var (
	//tempFilePatternRegExpString matches a string that is alphanumeric (with
	//hypens or underscores) and ends in an asterisk. This can then be used to
	//prefix files.
	tempFilePatternRegExpString = "^[-a-zA-Z0-9_]*.[*]$"
	tempFilePatternRegExp       *regexp.Regexp
	tempFilePattern             = "executive-*"

	singleUse = false
)

//NewExecutable takes a compiled Script to generate a file that can be executed,
//running the script. If none of the Script features are needed (such as
//templates), this can be used directly with a nil originScript.
func NewExecutable(originScript *Script, compiled string) (*Executable, error) {
	executable := Executable{
		OriginScript: *originScript,
		scriptPath:   "",
		env:          make([]string, 0),
		args:         make([]string, 0),
		workDir:      "",
		singleUse:    false,
	}
	scriptFile, err := os.CreateTemp("", tempFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for script: %w", err)
	}
	executable.scriptPath = scriptFile.Name()
	if _, err := scriptFile.Write([]byte(compiled)); err != nil {
		return nil, fmt.Errorf("could write compiled script to temp file: %w", err)
	}
	if err := scriptFile.Chmod(0555); err != nil {
		return nil, fmt.Errorf("could not make temp script file executable: %w", err)
	}
	return &executable, nil
}

//WithEnv sets an environment variable that should be set within the script's
//runtime env. Executable is returned and not directly modified as to support
//function chains.
func (e Executable) WithEnv(key, value string) Executable {
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, fmt.Sprintf("%s=%s", key, value))
	return e
}

//WithArg adds an argument that will be provided to the script upon execution.
//Args are provided to the script in the the order they are provided. Executable
//is returned and not directly modified as to support function chains.
func (e Executable) WithArg(arg string) Executable {
	if e.args == nil {
		e.args = make([]string, 0)
	}
	e.args = append(e.args, arg)
	return e
}

//WithOSEnv ensures that the script will be able to use the environment vars
//from the os env. Executable is returned and not directly modified as to
//support function chains.
func (e Executable) WithOSEnv() Executable {
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, os.Environ()...)
	return e
}

//WithWorkDir sets the working directory for the context of the script.
//Executable is returned and not directly modified as to support function
//chains.
func (e Executable) WithWorkDir(path string) Executable {
	e.workDir = path
	return e
}

//SingleUse sets the temporary file associated with the script to remove itself
//immediately after it starts execution. This can help keep things neat, but can
//also cause issues. Use carefully. Executable is returned to facilitate
//function chinaing.
func (e Executable) SingleUse() Executable {
	e.singleUse = true
	return e
}

//SetEnv sets an environment variable that should be set within the script's
//runtime env.
func (e *Executable) SetEnv(key, value string) error {
	if key == "" {
		return fmt.Errorf("key must not be empty")
	}
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, fmt.Sprintf("%s=%s", key, value))
	return nil
}

//SetArgs defines all the args that will be provided to the script upon
//execution. This overwrites any changes made via WithArg.
func (e *Executable) SetArgs(args []string) error {
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

//SetOSEnv ensures that the script will be able to use the environment vars
//from the os env.
func (e *Executable) SetOSEnv() {
	if e.env == nil {
		e.env = make([]string, 0)
	}
	e.env = append(e.env, os.Environ()...)
}

//SetWorkDir sets the working directory for the context of the script.
func (e *Executable) SetWorkDir(path string) error {
	if info, err := os.Stat(path); err != nil {
		return fmt.Errorf("failed to get directory info: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("working directory must be a directory")
	}
	e.workDir = path
	return nil
}

//SetSingleUse sets the temporary file associated with the script to remove itself
//immediately after it starts execution. This can help keep things neat, but can
//also cause issues. Use carefully. Executable is returned to facilitate
//function chinaing.
func (e *Executable) SetSingleUse(isSingleUse bool) {
	e.singleUse = isSingleUse
}

//Execute simply runs the "compiled" script with the env and workdir given. If
//the process returned is nil the process failed to start. If a single use
//script fails to be removed, the process is still returned along with the
//error.
func (e Executable) Execute() (*Process, error) {
	process := Process{
		OriginExecutable: e,
		cmd:              exec.Command(e.scriptPath),
	}
	if len(e.env) > 0 {
		process.cmd.Env = e.env
	}
	if e.workDir != "" {
		process.cmd.Dir = e.workDir
	}
	if stdout, err := process.cmd.StdoutPipe(); err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	} else {
		process.stdout = stdout
	}
	if stderr, err := process.cmd.StderrPipe(); err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	} else {
		process.stderr = stderr
	}
	if stdin, err := process.cmd.StdinPipe(); err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	} else {
		process.stdin = stdin
	}
	if err := process.cmd.Start(); err != nil || process.cmd.Process == nil {
		return nil, fmt.Errorf("process failed to start")
	}
	if e.singleUse {
		if err := e.Remove(); err != nil {
			return &process, fmt.Errorf("failed to remove single use script file: %w", err)
		}
	}
	return &process, nil
}

//Remove simply deletes the temporary file that is generated upon creation of an
//Executable. Doing so is not vital as the file is stored in a temporary dir,
//but it can help long running programs that frequently use executive.
func (e *Executable) Remove() error {
	if !e.ScriptPathExists() {
		return nil
	}
	if err := os.Remove(e.scriptPath); err != nil {
		return fmt.Errorf("failed to remove script file: %w", err)
	}
	return nil
}

//ScriptPath simply returns the path containing the temporary file of the
//script.
func (e Executable) ScriptPath() string {
	return e.scriptPath
}

//ScriptPathExists simply returns a bool representing the existence of the
//temporary script file (without discoling errors). THis can be used to
//determine if the executable has been removed. In the case where Stat returns
//an error that is not ErrNotExist, this assumes the file does exist as to allow
//for remove to operate more optimistically.
func (e Executable) ScriptPathExists() bool {
	if fileInfo, err := os.Stat(e.scriptPath); errors.Is(err, os.ErrNotExist) || fileInfo.IsDir() {
		return false
	}
	return true
}

//SetTempFilePrefix defines the prefix pattern for the temporary files created
//when "compiling" scripts into "executables". Any prefix must be only
//alphanumeric, with hypens and underscores permitted. It also must end in a
//single asterisk.
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