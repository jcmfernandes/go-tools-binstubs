package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Tool struct {
	Package string `yaml:"package"`
	// Defaults to `true`.
	AddToGoFile *bool  `yaml:"add_to_go_file"`
	Version     string `yaml:"version"`
	// Defaults to `true`.
	GenerateBinstub *bool `yaml:"generate_binstub"`
	// Defaults to the last component of the package.
	Binstub                      string   `yaml:"binstub"`
	GoRunModifiers               []string `yaml:"go_run_modifiers,flow"`
	OverrideGlobalGoRunModifiers bool     `yaml:"override_global_go_run_modifiers"`
	BinstubModifiers             []string `yaml:"binstub_modifiers,flow"`
}

func (t Tool) BinstubFilename() string {
	if len(t.Binstub) > 0 {
		return t.Binstub
	} else {
		return filepath.Base(t.Package)
	}
}

type Options struct {
	Package              string   `yaml:"package"`
	Tools                []*Tool  `yaml:"tools,flow"`
	GlobalGoRunModifiers []string `yaml:"global_go_run_modifiers,flow"`
	BuildTags            []string `yaml:"build_tags,flow"`

	// Defaults to `true`.
	OutputGoFile *bool `yaml:"output_go_file"`
	// Defaults to `./tools.go`.
	OutputGoFilePath string `yaml:"output_go_file_path"`
	// Defaults to `./bin`.
	OutputBinstubsDirectoryPath string `yaml:"output_binstubs_directory_path"`
}

const (
	goToolsBinstubsVersion = "v0.0.6"

	selfAbsPackage                      = "github.com/jcmfernandes/go-tools-binstubs"
	bashSourceAbsParentDirectoryVarName = "binstubAbsParentDirectory"
)

func (opts Options) Generate() error {
	if len(opts.Tools) == 0 {
		return nil
	}

	var t bool = true
	if opts.OutputGoFile == nil {
		opts.OutputGoFile = &t
	}
	if len(opts.OutputGoFilePath) == 0 {
		opts.OutputGoFilePath = "tools.go"
	}
	if len(opts.OutputBinstubsDirectoryPath) == 0 {
		opts.OutputBinstubsDirectoryPath = "bin"
	}

	for _, tool := range opts.Tools {
		if tool.AddToGoFile == nil {
			tool.AddToGoFile = &t
		}
		if tool.GenerateBinstub == nil {
			tool.GenerateBinstub = &t
		}
	}

	if *opts.OutputGoFile {
		if err := opts.generateToolsFile(); err != nil {
			return err
		}
	}
	if err := opts.generateBinstubs(); err != nil {
		return err
	}

	return nil
}

func (opts Options) generateToolsFile() error {
	if len(opts.Package) == 0 {
		return fmt.Errorf("missing a package for the tools file")
	}

	toolsFile, err := os.OpenFile(opts.OutputGoFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer toolsFile.Close()

	if len(opts.BuildTags) > 0 {
		fmt.Fprintf(toolsFile, "//go:build %s\n\n", strings.Join(opts.BuildTags, ","))
	}
	fmt.Fprintf(toolsFile, "// Code generated by \"%s\"; DO NOT EDIT.\n", selfAbsPackage)
	fmt.Fprintf(toolsFile, "\npackage %s\n", opts.Package)
	fmt.Fprintf(toolsFile, "\nimport (\n")
	for _, tool := range opts.Tools {
		if len(tool.Version) > 0 {
			continue
		}

		fmt.Fprintf(toolsFile, "\t_ \"%s\"\n", tool.Package)
	}
	fmt.Fprintf(toolsFile, ")\n")

	return nil
}

func (opts Options) generateBinstubs() error {
	if err := os.MkdirAll(opts.OutputBinstubsDirectoryPath, os.ModePerm); err != nil {
		return err
	}

	for _, tool := range opts.Tools {
		if !*tool.GenerateBinstub || len(tool.Package) == 0 {
			continue
		}

		binstubFilePath := filepath.Join(opts.OutputBinstubsDirectoryPath, tool.BinstubFilename())
		binstubFile, err := os.OpenFile(binstubFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0775)
		if err != nil {
			return err
		}
		defer binstubFile.Close()

		var goRunModifiers []string
		if tool.OverrideGlobalGoRunModifiers {
			goRunModifiers = append(goRunModifiers, tool.GoRunModifiers...)
		} else {
			goRunModifiers = append(goRunModifiers, opts.GlobalGoRunModifiers...)
			goRunModifiers = append(goRunModifiers, tool.GoRunModifiers...)
		}

		var goRunCommand = []string{"go run"}
		goRunCommand = append(goRunCommand, goRunModifiers...)
		if len(tool.Version) == 0 {
			goRunCommand = append(goRunCommand, tool.Package)
		} else {
			goRunCommand = append(goRunCommand, fmt.Sprintf("%s@%s", tool.Package, tool.Version))
		}
		goRunCommand = append(goRunCommand, tool.BinstubModifiers...)

		fmt.Fprintf(binstubFile, "#!/usr/bin/env bash\n")
		fmt.Fprintf(binstubFile, "# Code generated by \"%s\"; DO NOT EDIT.\n", selfAbsPackage)
		fmt.Fprintf(binstubFile, "\n%s=$( cd -- \"$( dirname -- \"${BASH_SOURCE[0]}\" )\" &> /dev/null && pwd )\n", bashSourceAbsParentDirectoryVarName)
		fmt.Fprintf(binstubFile, "\nexec %s \"$@\"\n", strings.Join(goRunCommand, " "))
	}

	return nil
}

var (
	input       = flag.String("input", "", "input file name")
	gentemplate = flag.String("gentemplate", "", "generate template YAML file name")
	version     = flag.Bool("version", false, "print the version")
)

func Usage() {
	flag.PrintDefaults()
}

func generateToolsFileAndBinstubs() error {
	yamlData, err := os.ReadFile(*input)
	if err != nil {
		return err
	}

	var opts Options
	if err := yaml.Unmarshal(yamlData, &opts); err != nil {
		return err
	}
	if err := opts.Generate(); err != nil {
		return err
	}

	return nil
}

func generateTemplate() error {
	var f bool
	var t bool = true
	opts := Options{
		Package:                     "tools",
		BuildTags:                   []string{"tools", "ignore"},
		OutputGoFile:                &t,
		OutputGoFilePath:            "tools.go",
		OutputBinstubsDirectoryPath: "bin",
		GlobalGoRunModifiers:        []string{"-x"},
		Tools: []*Tool{
			{
				Package:                      selfAbsPackage,
				GenerateBinstub:              &t,
				AddToGoFile:                  &f,
				Version:                      "v1.2.3",
				Binstub:                      "go-tools-binstubs",
				GoRunModifiers:               []string{"-work"},
				OverrideGlobalGoRunModifiers: false,
				BinstubModifiers:             []string{"-help"},
			},
		},
	}

	templateFile, err := os.OpenFile(*gentemplate, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer templateFile.Close()

	yamlData, err := yaml.Marshal(&opts)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(templateFile, "%s\n", string(yamlData))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	var err error
	var exitCode int
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		}
		os.Exit(exitCode)
	}()

	if *version {
		fmt.Println(goToolsBinstubsVersion)
		return
	}

	if len(*input) == 0 && len(*gentemplate) == 0 {
		flag.Usage()
		return
	} else if len(*input) > 0 && len(*gentemplate) > 0 {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "error: incompatible modifiers\n")
		exitCode = 1
		return
	}

	if len(*input) > 0 {
		err = generateToolsFileAndBinstubs()
		if err != nil {
			exitCode = 2
			return
		}
	} else if len(*gentemplate) > 0 {
		err = generateTemplate()
		if err != nil {
			exitCode = 2
			return
		}
	} else {
		panic(fmt.Errorf("this shouldn't happen"))
	}
}
