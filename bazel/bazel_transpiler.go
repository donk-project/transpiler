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
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/transformer"
	"snowfrost.garden/donk/transpiler/writer"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

type WriteMode int

const (
	WriteModeUnknown WriteMode = iota
	WriteModeSources
	WriteModeHeaders
)

func main() {
	outputPath := flag.String("output_path", "", "Directory for generated output")
	inputProto := flag.String("input_proto", "", "Location of input binarypb")
	// Change this if you want the C++ representation of your Dreammaker project
	// to have a different root name. `dtpo` is a vague allusion to 'tgstation'
	// and '.cc', the C++ definition file extension.
	projectName := flag.String("project_name", "dtpo", "Name of generated C++ project (filename friendly)")
	writeModeFlag := flag.String("write_mode", "", "Write mode")
	writeMode := WriteModeUnknown

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
	if *writeModeFlag == "" {
		panic("No --write_mode specified")
	} else {
		if *writeModeFlag == "sources" {
			writeMode = WriteModeSources
		} else if *writeModeFlag == "headers" {
			writeMode = WriteModeHeaders
		} else {
			panic("--write_mode " + *writeModeFlag + " not valid")
		}
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
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

	t := transformer.New(p, "dtpo", *projectName)
	t.BeginTransform()

	w := writer.New(t.Project(), *outputPath)
	if writeMode == WriteModeSources {
		w.WriteSources()
	} else if writeMode == WriteModeHeaders {
		w.WriteHeaders()
	}

	ctxt := writer.RegistrarContext{
		CoreNamespace:        t.CoreNamespace(),
		OutputPath:           *outputPath,
		Parser:               p,
		RootInclude:          "",
		TypeRegistrarInclude: "type_registrar.h",
		AffectedPaths:        t.AffectedPaths(),
	}

	if writeMode == WriteModeSources {
		writer.WriteTypeRegistrarSources(ctxt)
	} else if writeMode == WriteModeHeaders {
		writer.WriteTypeRegistrarHeaders(ctxt)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}