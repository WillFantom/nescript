package executive

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
)

type Output map[string]any

const (
	setOutputPattern string = `::set-output name=([^\s][^::][a-zA-Z_-]+)(?:\stype=([a-zA-Z]+))?::(.*)`
)

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

var (
	setOutputRegex *regexp.Regexp
)

func init() {
	setOutputRegex = regexp.MustCompile(setOutputPattern)
}
