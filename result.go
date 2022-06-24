package executive

import (
	"fmt"
	"time"

	"github.com/antonmedv/expr"
)

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

func (r Result) Evaluate(expression string) (bool, error) {
	program, err := expr.Compile(expression, expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("expression compilation failed: %w", err)
	}
	output, err := expr.Run(program, r.JSON)
	if err != nil {
		return false, fmt.Errorf("expression evaluation failed: %w", err)
	}
	if o, ok := output.(bool); ok {
		return o, nil
	} else {
		return false, fmt.Errorf("expression output was non-boolean")
	}
}
