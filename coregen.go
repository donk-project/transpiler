// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/transformer"
	"snowfrost.garden/donk/transpiler/writer"
)

// coregen generates the skeleton of the Donk API: all of the files in donk/api.
// This means running it will clobber any custom implementations filled in there.
// It is only used to expose the API used by the interpreter.
func main() {
	outputPath := flag.String("output_path", "", "Directory for generated output")
	includePrefix := flag.String("include_prefix", "", "Prefix to append to include statements")
	flag.Parse()

	if *outputPath == "" {
		panic("No --output_path specified")
	}

	prefix := ""
	if *includePrefix != "" {
		prefix = strings.TrimRight(*includePrefix, "/")
	}

	binarypb_path, err := bazel.Runfile("core.binarypb")
	if err != nil {
		panic(err)
	}
	in, err := ioutil.ReadFile(binarypb_path)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	g := &astpb.Graph{}

	if err := proto.Unmarshal(in, g); err != nil {
		log.Fatalln("Failed to parse:", err)
	}

	p := parser.NewParser(g)
	t := transformer.New(p, "donk", prefix)
	t.BeginTransform()

	w := writer.New(t.Project(), *outputPath)
	w.WriteOutput()
}
