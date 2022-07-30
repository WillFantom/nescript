# NEScript

[![Go Reference](https://pkg.go.dev/badge/github.com/willfantom/executive.svg)](https://pkg.go.dev/github.com/willfantom/executive) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/willfantom/executive?label=Latest%20Version&sort=semver&style=flat-square) 

<!-- TODO: Add NES repo link -->
Add automation to your network emulation workflows with [NES]() & NEScript 🚀

An [`os.exec`](https://pkg.go.dev/os/exec) wrapper providing extra functionality and cleaner usage when using it heavily.

Added features include:
 - Support for handlebar values in scripts ([go templates](https://pkg.go.dev/text/template))
 - Function chaining for cleaner code
 - Complex GitHub Actions style output parsing
 - Dynamic evaluation of output using expressions ([expr](https://github.com/antonmedv/expr))
 - Script execution on the local machine, ssh target, or docker container (plugin-friendly 🔌)

---

## Understanding Executive

Executive is divided into 5 core components:

 - **Script**: A script is somewhat self explantory. A script can either be created from a source (string, file, http), and can contain [template engine](https://pkg.go.dev/text/template) handlebars (awesome for loops, etc...). A script is not executed upon creation, instead further configuration can be set. When executing a script, a specific Executor should be specified (allowing for local & non-local execution).
 - **Executor**: A plugin that allows for scripts to be executed in many ways. Provided is a local executor (that just runs the script on the local machine), ssh executor (that executes the script on a remote SSH target), and a docker executor (for executing scripts on a docker container).
 - **Process**: A process is an executing or executed script instance. Calling for a `Result` from this will wait for execution to be complete. 
 - **Result**: A result is the output of an executed script, including the exit code, stdout and stderr.
 - **Output**: Output is key/value mapping of explicitly set outputs. This is done similarly to github actions, where outputs are picked up from stdout/stderr with a prefix similar to `::set-output name=example::...`. As these values can be typed (string, int, JSON), they can also be evaluated based on expressions.

---

## Key Features

### Templating

Using templates in scripts is easy, just have a handlebar value in the script named as desired. For example:
```bash
{{.Command}} "Hello, {{.Name}}"
```

Then to set the value to replace `{{.Name}}`, the following go code can be used:
```go
...
script, err := NewScript('{{.Command}} "Hello, {{.Name}}"').WithField("Command", "echo").WithField("Name", "world").Compile()
if err != nil {
	panic(err)
}
...
```

The templating system powering this supports other features too, such as loops when fields are slices of data etc...

> Shebangs (`#!/bin/bash` etc...) should not be used as these can be hard to use on certain executors. Instead, NEScript allows for a sub-command to be set, for example `sh -c`, where the script is provided as the last argument. This overall seems to be a more portable approach.

### Remote Execution

Scripts require an `Executor` to actually be executed. There are the 3 provided, but more can easily be created. Executors, such as SSH, can have required configuration parameters.

The limitations are that optional working directories for script execution are not available when executing over SSH, and signals such as SIGINT are not supported when using a remote Docker target.

> ⚠️ When using env vars over SSH, be sure to allow any (`*`) env var on the SSH server by setting the `AcceptEnv` option in `sshd`

### Output Handling & Evaluation

If specific output is desired to be able to evaluate a response to a script, this package allows for specific typed outputs to be set. If a line in StdOut or StdErr has a prefix similar to `::set-output name=example::`, the rest of the line is stored as an output value with the key being provided in the `name` field. For example, the output key/value `Hello/world` can be set like so if a script is executing via a shell such as bash:

```
echo ::set-output name=Hello::world
```

By default, all outputted keys and values are considered to be string types. Optionally, a value can be an int, or a JSON structure. For example:

- **INT**: 		`echo ::set-output name=example type=int::42`
- **JSON**: 	`echo ::set-output name=example type=json::{"sometext": "json", "anumber": 42}` 

Script results can then be programmatically evaluated to boolean values using expressions. For example, to ensure the number set in the above JSON example is 42, the expression could be given:

```go
...
if ok, err := result.CombinedOutput().Evaluate("example.anumber == 42"); err == nil && ok {
	fmt.Println("number is 42")
} else {
	panic(fmt.Errorf("number is not 42: %w", err))
}
...
```

