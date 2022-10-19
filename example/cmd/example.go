package main

import (
	"fmt"

	"github.com/willfantom/nescript"
	"github.com/willfantom/nescript/expr"
	"github.com/willfantom/nescript/local"
)

var (
	executor   nescript.ExecFunc = local.Executor("")
	evaluator  nescript.EvalFunc = expr.EvalFunc()
	expression string            = "nice == 69"
	command    *nescript.Cmd     = nescript.NewCmd("echo", "{{.Greeting}}, world!\n::set-output name=nice type=int::69")
)

func main() {
	process, err := command.
		WithField("Greeting", "Hello").
		CompileExec(executor)
	if err != nil {
		panic(err)
	}
	defer process.Close()

	result, err := process.Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("- Script StdOut:\n%s\n", result.StdOut)
	if success, err := result.Output(false).Evaluate(evaluator, expression); err != nil {
		panic(err)
	} else if success {
		fmt.Println("- The expression held true")
	} else {
		fmt.Println("- Something went wrong... The expression was not true")
	}
}
