// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"snowfrost.garden/donk/transpiler/parser"
)

type RegistrarContext struct {
	CoreNamespace        string
	OutputPath           string
	TypeRegistrarInclude string
	Parser               *parser.Parser
	RootInclude          string
}

func WriteTypeRegistrar(ctxt RegistrarContext) {
	rfpath, err := bazel.Runfile("snowfrost/donk/transpiler/templates")
	if err != nil {
		panic(err)
	}
	pattern := filepath.Join(rfpath, "type_registrar*.tmpl")
	funcMap := template.FuncMap{
		"StringsJoin":    strings.Join,
		"StringsToLower": strings.ToLower,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		panic(err)
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
	err = tmpl.ExecuteTemplate(file, "type_registrar.h.tmpl", ctxt)
	if err != nil {
		panic(err)
	}

	x, err := os.Create(abs + "/type_registrar.cc")
	defer x.Close()
	if err != nil {
		panic(err)
	}
	err = tmpl.ExecuteTemplate(x, "type_registrar.cc.tmpl", ctxt)
	if err != nil {
		panic(err)
	}

}
