package executive

import "time"

//Result represents the output of a completed script execution. This is only
//created by the `Result` function of a "Process" and thus the process must have
//exited. The time metrics included are from the process and do not incorporate
//the time used by executive.
type Result struct {
	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
	JSON     any    `json:"json"`
	ExitCode int    `json:"exitCode"`

	TotalTime  time.Duration `json:"executionTime"`
	SystemTime time.Duration
	UserTime   time.Duration
}
