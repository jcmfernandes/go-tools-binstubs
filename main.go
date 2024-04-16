package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/fatih/structtag"
)

var importRegex = regexp.MustCompile(`\s+_\s"([^"]+)"\s*/?/?\s*(.*)`)
var globalRegex = regexp.MustCompile(`\/{2}\s*(binstubsArgs:.*)`)

var binstubTemplate = template.Must(template.New("binstub").Parse(`#!/usr/bin/env bash
# Code generated by go-tools-binstubs. DO NOT EDIT.

binstubAbsFilePath=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

exec {{ .GoRunCommand }} {{ .Module }} "$@"
`))

// TemplateArgs are passed to binstubTpl.
type TemplateArgs struct {
	GoRunCommand string
	Module       string
}

// CommentOption is a parsed set of generator options specified via a postfix
// inline comment.
type CommentOption struct {
	Ignore bool
	Args   string
	Name   string
}

func parseComment(comment string) *CommentOption {
	tags, err := structtag.Parse(comment)
	if err != nil {
		panic("bad syntax for comment " + comment)
	}

	opt := CommentOption{}
	for _, tag := range tags.Tags() {
		if tag.Key != "binstub" {
			continue
		}
		if tag.Name == "-" && tag.Options == nil {
			opt.Ignore = true
			continue
		}

		opt.Name = tag.Name
		opt.Args = strings.Join(tag.Options, " ")
	}

	return &opt
}

func generateBinstub(module string, comment string) {
	options := parseComment(comment)
	if options.Ignore {
		return
	}
	binstubName := options.Name
	if len(binstubName) == 0 {
		binstubName = filepath.Base(module)
	}
	binstubPath := filepath.Join(binstubsDirectoryPath, binstubName)

	f, err := os.OpenFile(binstubPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	tplArgs := &TemplateArgs{Module: module, GoRunCommand: "go run"}
	if globalArgs != "" {
		tplArgs.GoRunCommand = fmt.Sprintf("%s %s", tplArgs.GoRunCommand, globalArgs)
	}
	if options.Args != "" {
		tplArgs.GoRunCommand = fmt.Sprintf("%s %s", tplArgs.GoRunCommand, options.Args)
	}
	err = binstubTemplate.Execute(f, tplArgs)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "Wrote binstub for %s to %s\n", module, binstubPath)
}

var toolsFilePath string
var binstubsDirectoryPath string
var globalArgs string

func main() {
	flag.StringVar(&toolsFilePath, "tools_file", "tools.go", "the path to tools.go")
	flag.StringVar(&binstubsDirectoryPath, "binstubs_dir", "bin", "that path to the binstubs directory")
	flag.Usage = func() { flag.PrintDefaults() }
	flag.Parse()

	f, err := os.Open(toolsFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0

	for scanner.Scan() {
		matches := globalRegex.FindStringSubmatch(scanner.Text())
		if len(matches) >= 2 {
			globalArgs = matches[1]
			tags, err := structtag.Parse(globalArgs)
			if err != nil {
				panic(err)
			}

			stag, err := tags.Get("binstubsArgs")
			if err != nil {
				panic(err)
			}
			globalArgs = fmt.Sprintf("%s %s", stag.Name, strings.Join(stag.Options, " "))
		}
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	scanner = bufio.NewScanner(f)

	if err := os.MkdirAll(binstubsDirectoryPath, os.ModePerm); err != nil {
		panic(err)
	}
	for scanner.Scan() {
		matches := importRegex.FindStringSubmatch(scanner.Text())
		if len(matches) >= 3 {
			count++
			generateBinstub(matches[1], matches[2])
		}
	}
	if count == 0 {
		fmt.Fprintf(os.Stderr, "Failed to generate binstubs: no imports found in %s\n", toolsFilePath)
		os.Exit(1)
	}
}
