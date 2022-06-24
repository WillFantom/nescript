# Executive

[![Go Reference](https://pkg.go.dev/badge/github.com/willfantom/executive.svg)](https://pkg.go.dev/github.com/willfantom/executive) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/willfantom/executive?label=Latest%20Version&sort=semver&style=flat-square) 

[`os.exec`](https://pkg.go.dev/os/exec), but a little fancier! A simple wrapper providing extra functionality and cleaner usage when using it heavily.

> I have made this with a specific use case in mind, so the functionality added is of course biased towards that...

---

## Script

A script is considered to be a normal script file, however, it may also contain handlebar values that can be populated in the program. For example, a script compatible with this package is:

> 💡 As the content of the script is written to and executed from a file, the shebang is needed on the first line...

_`example.sh`_
```bash
#!/bin/bash
set -e

echo "{{.Title}} Example"
{{range .People}}    
echo -e "\tName: {{.Name.First}} {{.Name.Last}}"
echo -e "\tCharacter: {{.Character}}"
echo -e "\t-----"
{{end}}

echo "IMDB Score: "$IMDB
```

> ⚠️ Handlebar values that are not given a value by the program will simply be omitted from the script

For more on how to use golang templates, see [here](https://pkg.go.dev/text/template).

## Executable

An executable is simply a "compiled" version of the script (where compiled here is just saying it has been though the template system). It is worth nothing that the script is written to a temporary file, and to keep things neat, should be removed when the script is done with.

## Process

A process here is considered an executed/executing instance of the executable. This can be killed, strings can be written to the stdin, or `Result` can be called to wait for the script to exit and get the result.

## Result

A result just contains the core output values of the process (stdout, stderr, exit code) and the execution time of the process. After a script has completed execution, the std out is evaulated line-by-line. If a line starts with the prefix `::set-output name=<> (type=<>)::`, the remainer of the line is considered the value and stored in the result outputs.

### Evaulation

Any script can store outputs, allowing the reults of the script to be automatically parsed. To store an output, a script must print a line to StdOut starting with `::set-output name=example::`, where `example` can be replaced with any name for the specific output. The remainer of the line is then considered the output's value. 
- If the output should be interpreted as another type than string, this should be specified, for example:
  `::set-output name=example type=int::42`. 

- For more complex outputs, JSON can be specified, for example: 
  `::set-output name=example type=json::{"hello": "world:, "number": 42}`

To evaluate these on a scripts result, expressions can be given (using [expr](https://github.com/antonmedv/expr)). 
- For example an output of type `int` with the name `example` can be evaluated like so:
  `example > 41` 
- JSON evaluation works too, using the above JSON example, the following expressions would work: 
  `example.number > 41`.



---

## Example

If the script (`example.sh`) where to be used, a program could look like so:

_`example.go`_
```go
package main

import (
	"fmt"

	"github.com/willfantom/executive"
)

type Person struct {
	Name      *Name
	Character string
}

type Name struct {
	First string
	Last  string
}

const title = "Shrek Cast"

var people = []Person{
	{
		Name: &Name{
			First: "Mike",
			Last:  "Myers",
		},
		Character: "Shrek",
	},
	{
		Name: &Name{
			First: "Eddie",
			Last:  "Murphy",
		},
		Character: "Donkey",
	},
	{
		Name: &Name{
			First: "Cameron",
			Last:  "Diaz",
		},
		Character: "Fiona",
	},
}

func main() {

	script, err := executive.NewScriptFromFile("shrek", "./example.sh")
	if err != nil {
		panic(err)
	}
	executable, err := script.WithField("People", people).WithField("Title", title).Compile()
	if err != nil {
		panic(err)
	}
	process, err := executable.WithOSEnv().WithEnv("IMDB", "7.9").Execute()
	if err != nil {
		panic(err)
	}
	result, _ := process.Result()
	fmt.Printf("%+v\n", result)

}
```
