package nescript

import "fmt"

// ExecFunc is used to execute a script, in-turn creating a process. If the
// script fails to be executed for any reason, this function can also error.
type ExecFunc func(Script) (Process, error)

// Executor is effectively a factory for ExecFunc. This allows for a complex
// ExecFunc to be generated based on a configuration. It is this that enables
// the execution-focused plugins.
type Executor interface {
	ExecFunc() (ExecFunc, error)
}

// Exec will call the given executor's ExecFunc, then execute the script using
// the ExecFunc. Returned will be the process that is created as a result of
// execution. An error is returned if the script fails to execute for any
// reason.
func (s Script) Exec(executor Executor) (Process, error) {
	if f, err := executor.ExecFunc(); err == nil && f != nil {
		return f(s)
	} else {
		return nil, fmt.Errorf("failed to get executor function: %w", err)
	}
}

// CompileExec will "compile" the script using the given data and the golang
// template system, then calls the given executor's ExecFunc, then execute the
// script using the ExecFunc. Returned will be the process that is created as a
// result of execution. An error is returned if the script fails to execute for
// any reason.
func (s Script) CompileExec(executor Executor) (Process, error) {
	cs, err := s.Compile()
	if err != nil {
		return nil, err
	}
	return cs.Exec(executor)
}
