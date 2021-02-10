// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"

	cctpb "snowfrost.garden/vasker/cc_grammar"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
)

type scopeCtxt struct {
	curDeclFile    *cctpb.DeclarationFile
	curDefnFile    *cctpb.DefinitionFile
	curClassDecl   *cctpb.ClassDeclaration
	curPath        *paths.Path
	curType        *parser.DMType
	curProc        *parser.DMProc
	declaredVars   []varRepresentation
	curDefnHeaders map[string]bool
	curDeclHeaders map[string]bool
	parentScope    *scopeCtxt
}

func (s *scopeCtxt) addDeclHeader(n string) {
	s.curDeclHeaders[n] = true
}

func (s *scopeCtxt) addDefnHeader(n string) {
	s.curDefnHeaders[n] = true
}

func (s *scopeCtxt) child() *scopeCtxt {
	child := s
	child.parentScope = s
	child.declaredVars = nil

	return child
}

func (s *scopeCtxt) hasVar(name string) bool {
	for _, v := range s.declaredVars {
		if v.name == name {
			return true
		}
	}
	return false
}

func (s *scopeCtxt) isRoot() bool {
	return s.curPath.IsRoot()
}

func (s *scopeCtxt) String() string {
	return fmt.Sprintf("ScopeContext<%v>", s.curPath)
}

func (s *scopeCtxt) childType(p *paths.Path, t *parser.DMType) *scopeCtxt {
	child := s
	child.parentScope = s
	child.declaredVars = nil
	child.curPath = p
	child.curType = t

	return child
}

func (s *scopeCtxt) parent() *scopeCtxt {
	return s.parentScope
}
