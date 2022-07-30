package nescript

import "os"

// Process is a single instance of the script, either running or exited. A
// process can be used to control the script and extract results from a script
// that has completed its execution.
type Process interface {
	// Kill sends a SIGKILL to the running process. If this fails, for example if
	// the process is not running, this will return an error.
	Kill() error

	// Signal sends a signal (such as SIGINT) to the running process. If this
	// fails, for example if the process is not running, this will return an
	// error.
	Signal(os.Signal) error

	// Write sends a string to the process's STDIN. Note that the string is sent
	// as-is. Thus if the program is looking for a newline before the read is
	// complete, this must be included in the string provided. If the write fails,
	// an error is returned.
	Write(string) error

	// Result waits for a script to complete execution, then a result is returned.
	// If the script returns an unknown error, this will also error.
	Result() (*Result, error)

	// Close should be called on a process, freeing any resources used where
	// appropriate.
	Close()
}
