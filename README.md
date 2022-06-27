# Executive

[![Go Reference](https://pkg.go.dev/badge/github.com/willfantom/executive.svg)](https://pkg.go.dev/github.com/willfantom/executive) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/willfantom/executive?label=Latest%20Version&sort=semver&style=flat-square) 

[`os.exec`](https://pkg.go.dev/os/exec), but a little fancier! A simple wrapper providing extra functionality and cleaner usage when using it heavily.

Added features include:
 - Support for handlebar values in scripts ([go templates](https://pkg.go.dev/text/template))
 - Function chaining for cleaner code
 - Complex output parsing
 - Dynamic evaluation of output using expressions ([expr](https://github.com/antonmedv/expr))
 - Script execution on the local machine, ssh target, or docker container

---

## Understanding Executive

Executive is divided into 5 core components:

 - **Template**: A template is a normal script file that can also contain handlebars for use with Go's [templating engine](https://pkg.go.dev/text/template). These can be more powerful than simple ENV vars (although this still allows for ENV vars to be set), as they are script language agnostic and support features such as loops.
 - **Script**: A script is somewhat self explantory. A script can either be created from a source (string, file, http), or be a 'compiled' template. A script is not executed upon creation. When executing a script, a remote target can be optionally specified.
 - **Process**: A process is an executing or executed script instance. Calling for a `Result` from this will wait for execution to be complete. 
 - **Result**: A result is the output of an executed script, including the exit code, stdout and stderr.
 - **Output**: Output is key/value mapping of explicitly set outputs. This is done similiarly to github actions, where outputs are picked up from stdout/stderr with a prefix similar to `::set-output name=example::...`. As these values can be typed (string, int, json), they can also be evaluated based on expressions.

---

## Key Features

### Templating

Using templates in scripts is easy, just have a handlebar value in the script named as desired. For example:
```bash
#!/bin/bash
{{.Command}} "Hello, {{.Name}}"
```

Then to set the value to replace `{{.Name}}`, the following go code can be used:
```go
...
script, err := myTemplate.WithField("Command", "echo").WithField("Name", "world").Compile()
if err != nil {
	panic(err)
}
...
```

The templating system powering this supports other features too, such as loops when fields are slices of data etc...

### Remote Execution

Scripts can be executed on your local machine when the `Execute` method is called, However, they can also be executed (with some limitations) on remote entities, such as SSH targets and Docker containers. 

- **SSH**: Replace `Execute()` with `ExecuteOverSSH("username", "10.0.0.1:22", ssh.Password("mysecurepassword"))`. Any ssh auth can be used, including key-based auth.
- **Docker**: Replace `Execute()` with `ExecuteOverDocker(cli, "containerid", "sh -c")`, where `cli` is a configured docker client and `"sh -c"` is the subprocess in which the script is passes as the next argument (a common alternative might be `/bin/bash -c`).

The limitations are that optional working directories for script execution are not available when executing over SSH, and signals such as SIGINT are not supported when using a remote Docker target.

> ⚠️ When using env vars over SSH, be sure to allow any (`*`) env var on the SSH server by setting the `AcceptEnv` option in `sshd`

### Output Handling & Evaluation

If specific output is desired to be able to evaluate a response to a script, this package allows for specific typed outputs to be set. If a line in StdOut or StdErr has a prefix similar to `::set-output name=example::`, the rest of the line is stored as an output value with the key being provided in the `name` field. For example, the output key/value `Hello/world` can be set like so if a script is executing via a shell such as bash:

```
echo ::set-output name=Hello::world
```

By default all outputted keys and values are considered to be string types. Optionally, a value can be an int, or a json structure. For example:

- **INT**: 		`echo ::set-output name=example type=int::42`
- **JSON**: 	`echo ::set-output name=example type=json::{"sometext": "json", "anumber": 42}` 

Script resuts can then be programatically evaluated to boolean values using expressions. For example, to ensure the number set in the above JSON example is 42, the expression could be given:

```go
...
if ok, err := result.CombinedOutput().Evaluate("example.anumber == 42"); err == nil && ok {
	fmt.Println("number is 42")
} else {
	panic(fmt.Errorf("number is not 42: %w", err))
}
...
```

