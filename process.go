package executive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

const (
	setOutputPattern string = `::set-output name=([^\s][^::][a-zA-Z_-]+)(?:\stype=([a-zA-Z]+))?::(.*)`
)

var (
	setOutputRegex *regexp.Regexp
)

//Process represents a single instance of the script.
type Process struct {
	OriginExecutable Executable

	cmd         *exec.Cmd
	stdin       io.Writer
	stdoutBytes bytes.Buffer
	stderrBytes bytes.Buffer
}

func (p *Process) Kill() error {
	if !p.Exited() {
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	return nil
}

func (p *Process) SigInt() error {
	if !p.Exited() {
		if err := p.cmd.Process.Signal(syscall.SIGINT); err != nil {
			return fmt.Errorf("failed to send sigint to process: %w", err)
		}
	}
	return nil
}

func (p *Process) Exited() bool {
	if process, err := os.FindProcess(p.cmd.Process.Pid); err != nil || process == nil {
		return true
	}
	return false
}

func (p *Process) Write(input string) error {
	if p.Exited() {
		return fmt.Errorf("can not write to stdin, process has exited")
	}
	if _, err := io.WriteString(p.stdin, input); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	return nil
}

func (p *Process) Result() (*Result, error) {
	if err := p.cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("failed to wait for process: %w", err)
		}
	}
	result := Result{
		StdOut: string(p.stdoutBytes.String()),
		StdErr: string(p.stderrBytes.String()),
	}
	result.Outputs = extractOutputs(result.StdOut)
	result.ExitCode = p.cmd.ProcessState.ExitCode()
	result.SystemTime = p.cmd.ProcessState.SystemTime()
	result.UserTime = p.cmd.ProcessState.UserTime()
	result.TotalTime = result.SystemTime + result.UserTime
	if err := p.cmd.Process.Release(); err != nil {
		return nil, fmt.Errorf("failed to release to process resources: %w", err)
	}
	return &result, nil
}

func extractOutputs(fullString string) map[string]any {
	outputs := make(map[string]any)
	lines := strings.Split(fullString, "\n")
	for _, l := range lines {
		matches := setOutputRegex.FindAllStringSubmatch(l, -1)
		if len(matches) > 0 {
			match := matches[0]
			name := match[1]
			t := match[2]
			value := match[3]
			switch strings.ToLower(t) {
			case "json", "j":
				rawJSON := json.RawMessage(value)
				var jsonValue any
				json.Unmarshal(rawJSON, &jsonValue)
				outputs[name] = jsonValue
			case "int", "i":
				if v, err := strconv.Atoi(value); err == nil {
					outputs[name] = v
				}
			default:
				outputs[name] = value
			}
		}
	}
	return outputs
}

func init() {
	setOutputRegex = regexp.MustCompile(setOutputPattern)
}
