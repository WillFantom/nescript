package nescript

import (
	"bytes"
	"fmt"
	"html/template"
)

type Cmd struct {
	command   string
	args      []string
	formatter Formatter
	*dynamicData
}

func NewCmd(command string, args ...string) *Cmd {
	if len(args) == 0 {
		args = make([]string, 0)
	}
	cmd := Cmd{
		command:   command,
		args:      args,
		formatter: defaultCmdFormatter,
		dynamicData: &dynamicData{
			data: make(map[string]any),
			env:  make([]string, 0),
		},
	}
	return &cmd
}

// Raw returns the command split by its arguments in its current state. If not
// compiled, handlebar values will still be present.
func (c Cmd) Raw() []string {
	return append([]string{c.command}, c.args...)
}

// WithArg adds an argument to the end of the current arguments slice associated
// with the command.
func (c Cmd) WithArg(arg string) Cmd {
	if c.args == nil {
		c.args = make([]string, 0)
	}
	c.args = append(c.args, arg)
	return c
}

// WithArgs appends multiple args to the end of the argument slice currently
// associated with the command.
func (c Cmd) WithArgs(args ...string) Cmd {
	for _, arg := range args {
		c = c.WithArg(arg)
	}
	return c
}

// WithField adds a key/value to the map of template data to be used when
// compiling the command. If the key already exists, it is overwritten.
func (c Cmd) WithField(key string, value any) Cmd {
	c.addField(key, value)
	return c
}

// WithFields takes a map of fields that is merged with the current command
// data. If a key already exists in the command data, overwite must be set to
// true in order to replace it, otherwise that key/value is left untouched.
func (c Cmd) WithFields(fields map[string]any, overwrite bool) Cmd {
	c.addFields(fields, overwrite)
	return c
}

// WithEnv takes one or more environmental variables in KEY=VALUE format. These
// will be used when executing the command. These will not be applied to the
// actual arguments of the command, but to any subprocess spawned by the
// command. This is different from the Env behavior of a Script.
func (c Cmd) WithEnv(env ...string) Cmd {
	c.addEnv(env...)
	return c
}

// WithLocalOSEnv appends the environmental variables from the local system to
// the env var set currently held be the command.
func (c Cmd) WithLocalOSEnv() Cmd {
	c.addLocalOSEnv()
	return c
}

func (c Cmd) WithFormatter(formatter Formatter) Cmd {
	c.formatter = formatter
	return c
}

// Compile uses the go template engine and the provided data fields to compile
// the command. These in-turn act a more portable approach than command-line
// arguments.
func (c Cmd) Compile() (Cmd, error) {
	compiledArgs := make([]string, len(c.args))
	for idx, a := range c.args {
		argTemplate, err := template.New("").Parse(a)
		if err != nil {
			return c, fmt.Errorf("failed to parse a command arg: %w", err)
		}
		if c.data == nil {
			c.data = make(map[string]any)
		}
		compiledArg := &bytes.Buffer{}
		if err := argTemplate.Execute(compiledArg, c.data); err != nil {
			return c, fmt.Errorf("cmd arg template could not be compiled: %w", err)
		}
		compiledArgs[idx] = compiledArg.String()
	}
	c.args = compiledArgs
	c.data = make(map[string]any)
	return c, nil
}

// MustCompile compiles the command, however will panic if an error occurred.
func (c Cmd) MustCompile() Cmd {
	compiledCmd, err := c.Compile()
	if err != nil {
		panic(err)
	}
	return compiledCmd
}
