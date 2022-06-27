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

// Template is a named raw script that may contain template handlebars, along
// with the data it is to be "compiled" with.
type Template struct {
	Name string `json:"name"`
	Raw  string `json:"raw"`

	data map[string]any
}

// NewTemplate returns a Template from a raw string with no deafult data.
func NewTemplate(name, raw string) (*Template, error) {
	return &Template{
		Name: name,
		Raw:  raw,

		data: make(map[string]any),
	}, nil
}

// NewTemplateFromFile creates a new Template using the contents of a given file
// and with no default data. An error is returned if the file contents can not
// be extracted.
func NewTemplateFromFile(name, filepath string) (*Template, error) {
	if scriptRawBytes, err := ioutil.ReadFile(filepath); err != nil {
		return nil, fmt.Errorf("failed to read script template file from path: %w", err)
	} else {
		return NewTemplate(name, string(scriptRawBytes))
	}
}

// NewTemplateFromHTTP creates a new Template using the contents retrived from a
// given link and with no default data. An error is returned if the link's
// contents can not be extracted.
func NewTemplateFromHTTP(name, link string) (*Template, error) {
	scriptURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("could not parse given link as a url: %w", err)
	}
	if response, err := http.Get(scriptURL.String()); err != nil {
		return nil, fmt.Errorf("could not get script template from url: %w", err)
	} else {
		defer response.Body.Close()
		if bodyBytes, err := io.ReadAll(response.Body); err != nil {
			return nil, fmt.Errorf("could not read the downloaded script: %w", err)
		} else {
			return NewTemplate(name, string(bodyBytes))
		}
	}
}

// WithData provides a data structure to the script so that handlebar values in
// the raw script can be populated when "compiled". If any key exists already,
// the data is overwritten with no warning. Note that the Script is not modifed,
// instead a modifed Template is returned, better suited for function chaining.
func (s Template) WithData(data map[string]any) Template {
	if s.data == nil {
		s.data = make(map[string]any)
	}
	//ignore error to facilitate function chaining
	mergo.Merge(&s.data, data, mergo.WithOverride)
	return s
}

// SetData provides a data structure to the template so that handlebar values in
// the raw script can be populated when "compiled". This is merged with any
// existing data. If the data fails to be merged with any existing data, and
// error is returned.
func (s Template) SetData(data map[string]any) error {
	if s.data == nil {
		s.data = make(map[string]any)
	}
	if err := mergo.Merge(&s.data, data, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge data: %w", err)
	}
	return nil
}

// WithField adds a single entry to the map of data that a Template is
// "compiled" with. If the key exists already, the data is overwritten with no
// warning. Note that the Script is not modifed, instead a modifed Template is
// returned, better suited for function chaining.
func (s Template) WithField(key string, value any) Template {
	if s.data == nil {
		s.data = make(map[string]any)
	}
	if key != "" {
		s.data[key] = value
	}
	//ignore else to facilitate function chaining
	return s
}

// Data simply returns the current supplementary data that will be used when
// "compiling" the script.
func (s Template) Data() map[string]any {
	if s.data == nil {
		s.data = make(map[string]any)
	}
	return s.data
}

// Compile runs the raw script string though the template engine, providing any
// data specified via WithData or WithField to the parser. Provided no template
// errors arise, a new Script is generated.
func (s Template) Compile() (*Script, error) {
	scriptTemplate, err := template.New(s.Name).Parse(s.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the script")
	}
	if s.data == nil {
		s.data = make(map[string]any)
	}
	compiled := &bytes.Buffer{}
	if err := scriptTemplate.Execute(compiled, s.data); err != nil {
		return nil, fmt.Errorf("script template could not be compiled: %w", err)
	}
	return NewScript(&s, compiled.String())
}

// CompileDryRun runs the template script string though the template engine,
// providing any data specified via WithData or WithField to the parser.
// Provided no template errors arise, the "compiled" script is returned.
func (s Template) CompileDryRun() (string, error) {
	scriptTemplate, err := template.New(s.Name).Parse(s.Raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse the script")
	}
	if s.data == nil {
		s.data = make(map[string]any)
	}
	compiled := &bytes.Buffer{}
	if err := scriptTemplate.Execute(compiled, s.data); err != nil {
		return "", fmt.Errorf("script template could not be compiled: %w", err)
	}
	return compiled.String(), nil
}
