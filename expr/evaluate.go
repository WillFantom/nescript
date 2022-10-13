package expr

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/willfantom/nescript"
)

// EvalFunc returns a evaluation function that will use the expr package to
// evaluate a given expression using the data provided from the output.
func EvalFunc() nescript.EvalFunc {
	return func(o nescript.Output, e string) (bool, error) {
		program, err := expr.Compile(e, expr.Env(o), expr.AsBool())
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
}
