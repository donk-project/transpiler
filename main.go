// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/golang/protobuf/proto"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/transformer"
	"snowfrost.garden/donk/transpiler/writer"

	astpb "snowfrost.garden/donk/proto/ast"
)


var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	outputPath := flag.String("output_path", "", "Directory for generated output")
	inputProto := flag.String("input_proto", "", "Location of input binarypb")
  // Change this if you want the C++ representation of your Dreammaker project to
  // have a different root name. `tgcc` refers to 'tgstation' and '.cc', the C++
  // definition file extension.
	projectName := flag.String("project_name", "tgcc", "Name of generated C++ project (filename friendly)")
	flag.Parse()

	if *outputPath == "" {
		panic("No --output_path specified")
	}
	if *inputProto == "" {
		panic("No --input_proto specified")
	}
	if *projectName == "" {
		panic("No --project_name specified")
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	in, err := ioutil.ReadFile(*inputProto)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	g := &astpb.Graph{}

	if err := proto.Unmarshal(in, g); err != nil {
		log.Fatalln("Failed to parse:", err)
	}

	start := time.Now()
	p := parser.NewParser(g)
	duration := time.Since(start)
	log.Printf("%6.2fs parsing", duration.Seconds())

	t := transformer.New(p, *projectName, *projectName)
	t.BeginTransform()

	w := writer.New(t.Project(), *outputPath)
	w.WriteOutput()

	ctxt := writer.RegistrarContext{
		CoreNamespace:        t.CoreNamespace(),
		OutputPath:           *outputPath,
		Parser:               p,
		RootInclude:          *projectName,
		TypeRegistrarInclude: *projectName + "/type_registrar.h",
	}

	writer.WriteTypeRegistrar(ctxt)

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
