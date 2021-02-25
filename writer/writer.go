// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"snowfrost.garden/donk/transpiler/paths"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

type Writer struct {
	outputPath string
	project    *cctpb.Project
	tmpl       *template.Template
}

type LineOutput struct {
	Indent int
	Data   string
}

func New(project *cctpb.Project, outputPath string) *Writer {
	var writer Writer
	writer.outputPath = outputPath
	writer.project = project
	rfpath, err := bazel.Runfile("templates")
	if err != nil {
		panic(err)
	}
	pattern := filepath.Join(rfpath, "transformer_*.tmpl")
	writer.tmpl, err = template.New("").Funcs(template.FuncMap{
		"printCppType":             writer.printCppType,
		"joinArgs":                 writer.joinArgs,
		"printStatement":           printStatement,
		"printBaseSpecifiers":      printBaseSpecifiers,
		"printMemberSpecifier":     writer.printMemberSpecifier,
		"printFunctionDeclaration": writer.printFunctionDeclaration,
		"printConstructor":         writer.printConstructor,
	}).ParseGlob(pattern)
	if err != nil {
		panic(err)
	}

	return &writer
}

func makeSortHeaders(headers []string) func(a, b int) bool {
	return func(a, b int) bool {
		aS := headers[a]
		bS := headers[b]
		if strings.HasPrefix(aS, "<") && !strings.HasPrefix(bS, "<") {
			return true
		}
		if strings.HasPrefix(bS, "<") && !strings.HasPrefix(aS, "<") {
			return false
		}
		return aS < bS
	}
}

func (w *Writer) WriteOutput() {
	abs, err := filepath.Abs(w.outputPath)
	if err != nil {
		panic(err)
	}

	for _, declFile := range w.project.GetDeclarationFiles() {
		sort.Slice(declFile.GetPreamble().GetHeaders(),
			makeSortHeaders(declFile.GetPreamble().GetHeaders()))

		sp := paths.New(declFile.GetFileMetadata().GetSourcePath())
		dir := abs + sp.ParentPath().FullyQualifiedString()
		os.MkdirAll(dir, 0755)
		x, err := os.Create(abs + "/" + declFile.GetFileMetadata().GetFilename())
		defer x.Close()
		if err != nil {
			panic(err)
		}
		err = w.tmpl.ExecuteTemplate(x, "transformer_declaration.h.tmpl", declFile)
		if err != nil {
			panic(err)
		}
	}

	for _, defnFile := range w.project.GetDefinitionFiles() {
		sort.Slice(defnFile.GetPreamble().GetHeaders(),
			makeSortHeaders(defnFile.GetPreamble().GetHeaders()))
		sp := paths.New(defnFile.GetFileMetadata().GetSourcePath())
		dir := abs + sp.ParentPath().FullyQualifiedString()
		os.MkdirAll(dir, 0755)
		y, err := os.Create(abs + "/" + defnFile.GetFileMetadata().GetFilename())
		defer y.Close()
		if err != nil {
			panic(err)
		}
		err = w.tmpl.ExecuteTemplate(y, "transformer_definition.cc.tmpl", defnFile)
		if err != nil {
			panic(err)
		}
	}
}
