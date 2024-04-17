# `go-tools-binstubs`

_Automatically generate binstubs for the tools in your golang project_

`go-tools-binstubs` is a nearly zero-configuration way to generate project
binstubs that **run the correct version of a go command tracked in your go.mod
file**.

1. Create a `tools.yaml` file that looks like this, pointing to the executable
    program path for each dependency:

``` yaml
package: tools
global_go_run_modifiers:
  - '-x'
build_tags:
  - tools
  - ignore
output_go_file_path: tools.go
output_binstubs_directory_path: bin
tools:
  - package: github.com/jcmfernandes/go-tools-binstubs
    ignore: false
    binstub: go-tools-binstubs
    go_run_modifiers:
      - '-work'
    override_global_go_run_modifiers: false
    binstub_modifiers:
      - '-help'
```

Or run `go run github.com/jcmfernandes/go-tools-binstubs -gentemplate
tools.yaml` to quickly generate a YAML template.

2. Run `go run github.com/jcmfernandes/go-tools-binstubs -input tools.yaml` to create
   corresponding shell scripts in `output_binstubs_directory_path` (defaults to `bin`) and go file in `output_go_file_path` (defaults to `tools.go`).

## Using a separate `go.mod` file

It's a good practice to create a separate `go.mod` file for your project's tool. Let's say you have the following project setup:

```
.
├── go.mod
├── go.sum
├── internal
│   └── tools
│       ├── go.mod
│       ├── tools.yaml
│       └── go.sum
└── main.go
```

And you want it to look like the following:

```
.
├── bin
│   └── a-tool
├── go.mod
├── go.sum
├── internal
│   └── tools
│       ├── go.mod
│       ├── tools.yaml
│       ├── tools.go
│       └── go.sum
└── main.go
```

You'll have to change your binstubs `-modfile` to `./internal/tools/go.mod` to
make this work. Since you don't want to hardcode paths in your binstubs (you
want them to work on everyone's machine), every binstub ships with bash variable
`binstubAbsParentDirectory` that contains the runtime absolute path to the
binstubs parent directory. You can use it to change your binstubs `-modfile`:

``` yaml
package: tools
global_go_run_modifiers:
  - '-modfile=${binstubAbsParentDirectory}/../internal/tools/go.mod'
tools:
  - package: github.com/jcmfernandes/go-tools-binstubs
```

And _voilà_!

## Acknowledgements

`go-tools-binstubs` started as a fork of [binstubs](https://github.com/brasic/binstubs).
