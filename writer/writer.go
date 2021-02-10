// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"text/template"

	"os"
	"path/filepath"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	cctpb "snowfrost.garden/vasker/cc_grammar"
	"snowfrost.garden/donk/transpiler/paths"
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
	rfpath, err := bazel.Runfile("snowfrost/donk/transpiler/templates")
	if err != nil {
		panic(err)
	}
	pattern := filepath.Join(rfpath, "transformer_*.tmpl")
	writer.tmpl, err = template.New("").Funcs(template.FuncMap{
		"printCppType":        writer.printCppType,
		"joinArgs":            writer.joinArgs,
		"printStatement":      printStatement,
		"printBaseSpecifiers": printBaseSpecifiers,
		"printMemberSpecifier": printMemberSpecifier,

	}).ParseGlob(pattern)
	if err != nil {
		panic(err)
	}

	return &writer
}

func (w *Writer) WriteOutput() {
	abs, err := filepath.Abs(w.outputPath)
	if err != nil {
		panic(err)
	}

	for _, declFile := range w.project.GetDeclarationFiles() {
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
