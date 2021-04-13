// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
)

type RegistrarContext struct {
	CoreNamespace        string
	OutputPath           string
	TypeRegistrarInclude string
	Parser               *parser.Parser
	RootInclude          string
	TypeIncludes         []string
	AffectedPaths        []paths.Path
}

// TODO: Migrate TypeRegistrar templates over to ordinary codegen
func WriteTypeRegistrar(ctxt RegistrarContext) {
	WriteTypeRegistrarHeaders(ctxt)
	WriteTypeRegistrarSources(ctxt)
}

func WriteTypeRegistrarSources(ctxt RegistrarContext) {
	funcMap := template.FuncMap{
		"StringsJoin":    strings.Join,
		"StringsToLower": strings.ToLower,
	}

	abs, err := filepath.Abs(ctxt.OutputPath)
	if err != nil {
		panic(err)
	}

	ctxt.TypeIncludes = append(ctxt.TypeIncludes, "root.h")
	for _, path := range ctxt.AffectedPaths {
		if !path.IsRoot() {
			inc := fmt.Sprintf("%v.h",
				strings.ToLower(path.FullyQualifiedString()))
			ctxt.TypeIncludes = append(ctxt.TypeIncludes, strings.TrimLeft(inc, "/"))
		}
	}

	sort.Slice(ctxt.TypeIncludes, makeSortHeaders(ctxt.TypeIncludes))

	x, err := os.Create(abs + "/type_registrar.cc")
	defer x.Close()
	if err != nil {
		panic(err)
	}
	err = template.Must(template.New("").Funcs(funcMap).Parse(TYPE_REGISTRAR_CC)).Execute(x, ctxt)
	if err != nil {
		panic(err)
	}

}

func WriteTypeRegistrarHeaders(ctxt RegistrarContext) {
	funcMap := template.FuncMap{
		"StringsJoin":    strings.Join,
		"StringsToLower": strings.ToLower,
	}

	abs, err := filepath.Abs(ctxt.OutputPath)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(abs + "/type_registrar.h")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	err = template.Must(template.New("").Funcs(funcMap).Parse(TYPE_REGISTRAR_H)).Execute(file, ctxt)
	if err != nil {
		panic(err)
	}

}
