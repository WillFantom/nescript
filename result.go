package executive

import (
	"time"
)

//Result represents the output of a completed script execution. This is only
//created by the `Result` function of a "Process" and thus the process must have
//exited. The time metrics included are from the process and do not incorporate
//the time used by executive.
type Result struct {
	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`

	TotalTime  time.Duration `json:"executionTime"`
	SystemTime time.Duration
	UserTime   time.Duration
}

// Output parses the specified outputs from the script's stdOut (or stdErr if
// specified). This is returned as a map. Any field that is not correctly
// parsed, will simply be ignored.
func (r Result) Output(useErr bool) Output {
	if useErr {
		return NewOutput(r.StdErr)
	}
	return NewOutput(r.StdOut)
}

// Output parses the specified outputs from the script's stdOut and stdErr. In
// the event that stdOut and stdErr specify an output of the same name, the
// value from stdOut is preferred. This is returned as a map. Any field that is
// not correctly parsed, will simply be ignored.
func (r Result) CombinedOutput() Output {
	output := NewOutput(r.StdErr)
	for k, v := range NewOutput(r.StdOut) {
		output[k] = v
	}
	return output
}
