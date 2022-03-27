package executive

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/imdario/mergo"
)

//Script is a named raw script that may contain template handlebars, along with
//the data it is to be "compiled" with.
type Script struct {
	Name string `json:"name"`
	Raw  string `json:"raw"`

	data map[string]interface{}
}

//NewScript returns a Script from a raw string with no deafult data.
func NewScript(name, raw string) (*Script, error) {
	return &Script{
		Name: name,
		Raw:  raw,

		data: make(map[string]interface{}),
	}, nil
}

//NewScriptFromFile creates a new Script using the contents of a given file and
//with no default data. An error is returned if the file contents can not be
//extracted.
func NewScriptFromFile(name, filepath string) (*Script, error) {
	if scriptRawBytes, err := ioutil.ReadFile(filepath); err != nil {
		return nil, fmt.Errorf("failed to read script file from path: %w", err)
	} else {
		return NewScript(name, string(scriptRawBytes))
	}
}

//NewScriptFromHTTP creates a new Script using the contents retrived from a
//given link and with no default data. An error is returned if the link's
//contents can not be extracted.
func NewScriptFromHTTP(name, link string) (*Script, error) {
	scriptURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("could not parse given link as a url: %w", err)
	}
	if response, err := http.Get(scriptURL.Path); err != nil {
		return nil, fmt.Errorf("could not get script from url: %w", err)
	} else {
		defer response.Body.Close()
		if bodyBytes, err := io.ReadAll(response.Body); err != nil {
			return nil, fmt.Errorf("could not read the downloaded script: %w", err)
		} else {
			return NewScript(name, string(bodyBytes))
		}
	}
}

//WithData provides a data structure to the script so that handlebar values in
//the raw script can be populated when "compiled". If any key exists already,
//the data is overwritten with no warning. Note that the Script is not modifed,
//instead a modifed Script is returned, better suited for function chaining.
func (s Script) WithData(data map[string]interface{}) Script {
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	//ignore error to facilitate function chaining
	mergo.Merge(&s.data, data, mergo.WithOverride)
	return s
}

//SetData provides a data structure to the script so that handlebar values in
//the raw script can be populated when "compiled". This is merged with any
//existing data. If the data fails to be merged with any existing data, and
//error is returned.
func (s Script) SetData(data map[string]interface{}) error {
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	if err := mergo.Merge(&s.data, data, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge data: %w", err)
	}
	return nil
}

//WithField adds a single entry to the map of data that a Script is "compiled"
//with. If the key exists already, the data is overwritten with no warning. Note
//that the Script is not modifed, instead a modifed Script is returned, better
//suited for function chaining.
func (s Script) WithField(key string, value interface{}) Script {
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	if key != "" {
		s.data[key] = value
	}
	//ignore else to facilitate function chaining
	return s
}

//Data simply returns the current supplementary data that will be used when
//"compiling" the script.
func (s Script) Data() map[string]interface{} {
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	return s.data
}

//Compile runs the raw script string though the template engine, providing any
//data specified via WithData or WithField to the parser. Provided no templaint
//errors arise, a new Executable is generated.
func (s Script) Compile() (*Executable, error) {
	scriptTemplate, err := template.New(s.Name).Parse(s.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the script")
	}
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	compiled := &bytes.Buffer{}
	if err := scriptTemplate.Execute(compiled, s.data); err != nil {
		return nil, fmt.Errorf("script template could not be compiled: %w", err)
	}
	return NewExecutable(&s, compiled.String())
}

//CompileDryRun runs the raw script string though the template engine, providing
//any data specified via WithData or WithField to the parser. Provided no
//templaint errors arise, the "compiled" script is returned.
func (s Script) CompileDryRun() (string, error) {
	scriptTemplate, err := template.New(s.Name).Parse(s.Raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse the script")
	}
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	compiled := &bytes.Buffer{}
	if err := scriptTemplate.Execute(compiled, s.data); err != nil {
		return "", fmt.Errorf("script template could not be compiled: %w", err)
	}
	return compiled.String(), nil
}
