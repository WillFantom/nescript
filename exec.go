package nescript

// ExecFunc is used to execute a script, in-turn creating a process. If the
// script fails to be executed for any reason, this function can also error.
type ExecFunc func(Script) (Process, error)

// Exec will call the given ExecFunc to execute the script. Returned will be the
// process that is created as a result of execution. An error is returned if the
// script fails to execute for any reason.
func (s Script) Exec(executor ExecFunc) (Process, error) {
	return executor(s)
}

// CompileExec will "compile" the script using the given data and the golang
// template system, then calls the given ExecFunc. Returned will be the process
// that is created as a result of execution. An error is returned if the script
// fails to execute for any reason.
func (s Script) CompileExec(executor ExecFunc) (Process, error) {
	cs, err := s.Compile()
	if err != nil {
		return nil, err
	}
	return cs.Exec(executor)
}
