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
	w.WriteHeaders()
	w.WriteSources()
}

func (w *Writer) WriteHeaders() {
	abs, err := filepath.Abs(w.outputPath)
	if err != nil {
		panic(err)
	}

	funcMap := template.FuncMap{
		"printCppType":             printCppType,
		"joinArgs":                 w.joinArgs,
		"printStatement":           printStatement,
		"printBaseSpecifiers":      printBaseSpecifiers,
		"printMemberSpecifier":     w.printMemberSpecifier,
		"printFunctionDeclaration": w.printFunctionDeclaration,
		"printConstructor":         w.printConstructor,
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
		err = template.Must(template.New("").Funcs(funcMap).Parse(TRANSFORMER_DECLARATION_H)).Execute(x, declFile)
		if err != nil {
			panic(err)
		}
	}

}

func (w *Writer) WriteSources() {
	abs, err := filepath.Abs(w.outputPath)
	if err != nil {
		panic(err)
	}

	funcMap := template.FuncMap{
		"printCppType":             printCppType,
		"joinArgs":                 w.joinArgs,
		"printStatement":           printStatement,
		"printBaseSpecifiers":      printBaseSpecifiers,
		"printMemberSpecifier":     w.printMemberSpecifier,
		"printFunctionDeclaration": w.printFunctionDeclaration,
		"printConstructor":         w.printConstructor,
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
		err = template.Must(template.New("").Funcs(funcMap).Parse(TRANSFORMER_DEFINITION_CC)).Execute(y, defnFile)
		if err != nil {
			panic(err)
		}
	}
}
