package nescript

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
)

// Output is a key/value map of data, where the value can be any type. This
// should be generated from NewOutput or from the helper methods of a result.
type Output map[string]any

var (
	// setOutputRegex determines the allowed syntax for a line outputting data.
	// E.g. ::set-output name=hello type=string::world!
	setOutputRegex *regexp.Regexp = regexp.MustCompile(`::set-output name=([^\s][^::][a-zA-Z_-]+)(?:\stype=([a-zA-Z]+))?::(.*)`)
)

// NewOutput creates an Output from a given input string (such as stdout). It
// will type cast select types if a type is given in the set-output message (or
// a string if not).
func NewOutput(source string) Output {
	outputs := make(map[string]any)
	lines := strings.Split(source, "\n")
	for _, l := range lines {
		matches := setOutputRegex.FindAllStringSubmatch(l, -1)
		if len(matches) > 0 {
			match := matches[0]
			name := match[1]
			t := match[2]
			value := match[3]
			switch strings.ToLower(t) {
			case "json", "j":
				rawJSON := json.RawMessage(value)
				var jsonValue any
				json.Unmarshal(rawJSON, &jsonValue)
				outputs[name] = jsonValue
			case "int", "i":
				if v, err := strconv.Atoi(value); err == nil {
					outputs[name] = v
				}
			default:
				outputs[name] = value
			}
		}
	}
	return outputs
}

// Evaluate takes an expr expression and uses the output data given to evaluate
// to a boolean. This will error is the expression can not be evaulation with
// the given data, or the output would not be a boolean.
func (o Output) Evaluate(expression string) (bool, error) {
	program, err := expr.Compile(expression, expr.Env(o), expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("expression compilation failed: %w", err)
	}
	output, err := expr.Run(program, o)
	if err != nil {
		return false, fmt.Errorf("expression evaluation failed: %w", err)
	}
	if o, ok := output.(bool); ok {
		return o, nil
	} else {
		return false, fmt.Errorf("expression output was non-boolean")
	}
}
