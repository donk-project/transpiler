// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transpiler

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

type WriteMode int

const (
	WriteModeUnknown WriteMode = iota
	WriteModeSources
	WriteModeHeaders
	WriteModeAll
)

type Options struct {
	cpuProfile  string
	memProfile  string
	outputPath  string
	inputProto  string
	projectName string
	writeMode   WriteMode
}

func OptsFromFlags() Options {
	o := Options{}
	flag.StringVar(&o.cpuProfile, "cpuprofile", "", "write cpu profile to `file`")
	flag.StringVar(&o.memProfile, "memprofile", "", "write memory profile to `file`")
	flag.StringVar(&o.outputPath, "output_path", "", "Directory for generated output")
	flag.StringVar(&o.inputProto, "input_proto", "", "Location of input binarypb")
	// `dtpo` simply stands for Donk Transpiler Project Output.
	flag.StringVar(&o.projectName, "project_name", "dtpo", "Name of generated C++ project (filename friendly)")
	writeModeFlag := flag.String("write_mode", "all", "Write mode")
	flag.Parse()

	if o.outputPath == "" {
		panic("No --output_path specified")
	}
	if o.inputProto == "" {
		panic("No --input_proto specified")
	}
	if o.projectName == "" {
		panic("No --project_name specified")
	}
	if *writeModeFlag == "" {
		panic("No --write_mode specified")
	} else {
		if *writeModeFlag == "sources" {
			o.writeMode = WriteModeSources
		} else if *writeModeFlag == "headers" {
			o.writeMode = WriteModeHeaders
		} else if *writeModeFlag == "all" {
			o.writeMode = WriteModeAll
		} else {
			panic("--write_mode " + *writeModeFlag + " not valid")
		}
	}

	return o
}

func (t transpiler) Transpile() {
	if t.opts.cpuProfile != "" {
		f, err := os.Create(t.opts.cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	in, err := ioutil.ReadFile(t.opts.inputProto)
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

	tr := transformer.New(p, t.opts.projectName, t.opts.projectName)
	tr.BeginTransform()

	w := writer.New(tr.Project(), t.opts.outputPath)
	if t.opts.writeMode == WriteModeSources {
		w.WriteSources()
	} else if t.opts.writeMode == WriteModeHeaders {
		w.WriteHeaders()
	} else if t.opts.writeMode == WriteModeAll {
		w.WriteOutput()
	}

	ctxt := writer.NewRegistrarContext()
	ctxt.CoreNamespace = tr.CoreNamespace()
	ctxt.OutputPath = t.opts.outputPath
	ctxt.Parser = p
	ctxt.RootInclude = ""
	ctxt.TypeRegistrarInclude = "type_registrar.h"

	for pth, typ := range p.TypesByPath {
		if tr.HasEmittableVars(typ) || tr.HasEmittableProcs(typ) {
			ctxt.AffectedPaths = append(ctxt.AffectedPaths, typ.ResolvedPath())
			if pth.IsRoot() {
				ctxt.AddRegistration(pth.FullyQualifiedString(), tr.CoreNamespace()+typ.ResolvedPath().AsNamespace())
			} else {
				ctxt.AddRegistration(pth.FullyQualifiedString(), tr.CoreNamespace()+"::"+typ.ResolvedPath().AsNamespace())
			}
		}
	}

	if t.opts.writeMode == WriteModeSources {
		writer.WriteTypeRegistrarSources(ctxt)
	} else if t.opts.writeMode == WriteModeHeaders {
		writer.WriteTypeRegistrarHeaders(ctxt)
	} else if t.opts.writeMode == WriteModeAll {
		writer.WriteTypeRegistrar(ctxt)
	}

	if t.opts.memProfile != "" {
		f, err := os.Create(t.opts.memProfile)
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

type transpiler struct {
	opts Options
}

func New(opts Options) transpiler {
	t := transpiler{opts: opts}
	return t
}
