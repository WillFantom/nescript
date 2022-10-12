package main

import (
	"fmt"

	"github.com/willfantom/nescript"
	"github.com/willfantom/nescript/local"
)

var scriptText string = `{{.Command}} "Hello, "$WHO"!"
echo "::set-output name=exampleNumber type=int::42"
`

func main() {

	script := nescript.NewScript(scriptText)
	process, err := script.WithField("Command", "echo").WithEnv("WHO=world").CompileExec(local.Executor("", nil))
	if err != nil {
		panic(err)
	}
	defer process.Close()
	result, err := process.Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(result.StdOut)

	if correct, err := result.CombinedOutput().Evaluate("exampleNumber == 42"); err == nil && correct {
		fmt.Println("the number in the output was 42, as expected")
	} else if err == nil && !correct {
		fmt.Println("the number in the output was not 42")
	} else {
		panic(err)
	}
}
