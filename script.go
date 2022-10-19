package nescript

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
)

// Script is some executable string, along with data to supplement its
// execution, such as template data and env vars. The string can contain go
// template handles. These are to replace script arguments, as the use of
// arguments can be complex on certain platforms where the script may be
// executed.
type Script struct {
	raw        string
	subcommand Subcommand
	*dynamicData
}

var (
	defaultSubcommand Subcommand = SCShell
)

// NewScript creates a script based on the raw executable string.
func NewScript(raw string) *Script {
	script := Script{
		raw:        raw,
		subcommand: defaultSubcommand,
		dynamicData: &dynamicData{
			data: make(map[string]any),
			env:  make([]string, 0),
		},
	}
	return &script
}

// NewScriptFromFile creates a Script from the string extracted from a given
// file. This can error if the file can not be read.
func NewScriptFromFile(path string) (*Script, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get script from file: %w", err)
	}
	return NewScript(string(fileBytes)), nil
}

// NewScriptFromHTTP creates a Script from the string extracted from a given
// URL. This can error if the contents of the remote resource can not be read.
func NewScriptFromHTTP(link string) (*Script, error) {
	scriptURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("could not parse given link as a url: %w", err)
	}
	if response, err := http.Get(scriptURL.String()); err != nil {
		return nil, fmt.Errorf("could not get script from url: %w", err)
	} else {
		defer response.Body.Close()
		if bodyBytes, err := io.ReadAll(response.Body); err != nil {
			return nil, fmt.Errorf("could not read the downloaded script: %w", err)
		} else {
			return NewScript(string(bodyBytes)), nil
		}
	}
}

// Raw returns the raw executable string as is. If the script contains template
// handlebars, they will be returned as provided, not compiled.
func (s Script) Raw() string {
	return s.raw
}

func (s Script) WithSubcommand(sc Subcommand) Script {
	s.subcommand = sc
	return s
}

// WithField adds a key/value to the map of template data to be used when
// compiling the script. If the key already exists, it is overwritten.
func (s Script) WithField(key string, value any) Script {
	s.addField(key, value)
	return s
}

// WithFields takes a map of fields that is merged with the current script data.
// If a key already exists in the script data, overwite must be set to true in
// order to replace it, otherwise that key/value is left untouched.
func (s Script) WithFields(fields map[string]any, overwrite bool) Script {
	s.addFields(fields, overwrite)
	return s
}

// WithEnv takes one or more environmental variables in KEY=VALUE format. These
// will be used when executing the script.
func (s Script) WithEnv(env ...string) Script {
	s.addEnv(env...)
	return s
}

// WithLocalOSEnv appends the environmental variables from the local system to
// the env var set currently held be the script.
func (s Script) WithLocalOSEnv() Script {
	s.addLocalOSEnv()
	return s
}

// Compile uses the go template engine and the provided data fields to compile
// the script. These in-turn act a more portable approach than command-line
// arguments.
func (s Script) Compile() (Script, error) {
	scriptTemplate, err := template.New("").Parse(s.raw)
	if err != nil {
		return s, fmt.Errorf("failed to parse the script: %w", err)
	}
	if s.data == nil {
		s.data = make(map[string]any)
	}
	compiledRaw := &bytes.Buffer{}
	if err := scriptTemplate.Execute(compiledRaw, s.data); err != nil {
		return s, fmt.Errorf("script template could not be compiled: %w", err)
	}
	s.raw = compiledRaw.String()
	s.data = make(map[string]any)
	return s, nil
}

// MustCompile compiles the script, however will panic if an error occurs.
func (s Script) MustCompile() Script {
	compiledScript, err := s.Compile()
	if err != nil {
		panic(err)
	}
	return compiledScript
}

// Cmd converts the script to a Cmd in the format [subcommand parts... , raw].
// For example, the raw script "ping 8.8.8.8" with a subcommand ["sh", "-c"],
// this would result in a Cmd equivalent to ["sh", "-c", "ping 8.8.8.8"]. This
// does not compile the script first or the command after, thus handlebar values
// will persist.
func (s Script) Cmd() Cmd {
	command := append(s.subcommand, s.raw)
	var cmd *Cmd
	if len(command) <= 0 {
		cmd = NewCmd("")
	} else if len(command) == 1 {
		cmd = NewCmd(command[0])
	} else {
		cmd = NewCmd(command[0], command[1:]...)
	}
	cmd.dynamicData = s.dynamicData
	cmd.formatter = defaultScriptFormatter
	return *cmd
}
