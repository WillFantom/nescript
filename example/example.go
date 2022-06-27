package main

import (
	"fmt"

	"github.com/willfantom/executive"
)

var scriptText string = `#!/bin/bash
{{.Command}} "Hello, "$WHO"!"
echo "::set-output name=exampleNumber type=int::42"
`

func main() {
	template, err := executive.NewTemplate("example", scriptText)
	if err != nil {
		panic(err)
	}

	script, err := template.WithField("Command", "echo").Compile()
	if err != nil {
		panic(err)
	}

	// sets the env var "WHO" to world and executes the script locally. could also
	// be executed in a docker container or over ssh.
	process, err := script.WithEnv("WHO", "world").Execute()
	if err != nil {
		panic(err)
	}

	result, err := process.Result()
	if err != nil {
		panic(err)
	}

	if correct, err := result.CombinedOutput().Evaluate("exampleNumber == 42"); err == nil && correct {
		fmt.Println("the number in the output was 42, as expected")
	} else if err == nil && !correct {
		fmt.Println("the number in the output was not 42")
	} else {
		panic(err)
	}
}
