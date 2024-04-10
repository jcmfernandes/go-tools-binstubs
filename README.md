# `go-tools-binstubs`

_Automatically generate binstubs using tools.go_

`go-tools-binstubs` is a nearly zero-configuration way to generate project
binstubs that run the correct version of a go command tracked in your go.mod
file.

1. Create a `tools.go` that looks like this, pointing to the executable
    program path for each dependency:

``` go
package main

import (
    _ "github.com/path/to/dep/cmd/something"
    _ "github.com/golang-migrate/migrate/v4/cmd/migrate"
)
```

2. Run `go run github.com/jcmfernandes/go-tools-binstubs` to create
   corresponding shell scripts in `bin/`. You can also add a `go:generate`
   comment to the top of tools.go instead.

3. If you need to import a tool that you don't want a binstub for, add an
   inline `binstub:"-"` comment after the import:

``` go
import (
    _ "github.com/path/to/dep/cmd/something" // binstub:"-"
)
```

4. If you need extra flags passed to `go run`, add an inline
   `binstub:foo,-tags,tools"` comment after the import:

``` go
import (
    _ "github.com/path/to/dep/cmd/something" // binstub:"foo,-tags,tools"
)
```

    Where `foo` is the name of the binstub and `-tags,tools` are the extra
    flags, that are joined with whitespaces.

## Acknowledgements

`go-tools-binstubs` is a fork of [binstubs](https://github.com/brasic/binstubs).
