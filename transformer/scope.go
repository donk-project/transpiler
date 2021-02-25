// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package scope

import (
	"fmt"

	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

type VarType int

const (
	VarTypeUnknown VarType = iota
	VarTypeInt
	VarTypeFloat
	VarTypeString
	VarTypeDMObject
	VarTypeResource
	VarTypePrefab
	VarTypeList
	VarTypeListIterator
)

func (v VarType) String() string {
	return [...]string{"UnknownType", "Int", "Float", "String", "DMObject", "Resource", "Prefab", "List", "ListIterator"}[v]
}

type VarScope int

const (
	VarScopeUnknown VarScope = iota
	VarScopeField
	VarScopeLocal
	VarScopeGlobal
)

func (v VarScope) String() string {
	return [...]string{"UnknownScope", "Field", "Local", "Global"}[v]
}

type VarInScope struct {
	Name         string
	defaultValue string
	curValue     string
	Type         VarType
	Scope        VarScope
}

type HeaderCollection struct {
	Headers map[string]bool
}

type DeclaredVars struct {
	Vars map[string]VarInScope
}

type DeclaredProcs struct {
	Procs map[string]bool
}

func (d *DeclaredProcs) Add(s string) {
	d.Procs[s] = true
}

func NewHeaderCollection() *HeaderCollection {
	hc := &HeaderCollection{}
	hc.Headers = make(map[string]bool)
	return hc
}

func NewDeclaredVars() *DeclaredVars {
	dv := &DeclaredVars{}
	dv.Vars = make(map[string]VarInScope)
	return dv
}

func NewDeclaredProcs() *DeclaredProcs {
	dv := &DeclaredProcs{}
	dv.Procs = make(map[string]bool)
	return dv
}

type ScopeCtxt struct {
	CurDeclFile    *cctpb.DeclarationFile
	CurDefnFile    *cctpb.DefinitionFile
	curClassDecl   *cctpb.ClassDeclaration
	CurPath        *paths.Path
	CurType        *parser.DMType
	CurProc        *parser.DMProc
	DeclaredVars   *DeclaredVars
	DeclaredProcs  *DeclaredProcs
	CurDefnHeaders *HeaderCollection
	CurDeclHeaders *HeaderCollection
	parentScope    *ScopeCtxt
	currentDepth   int
}

func (s *ScopeCtxt) AddScopedVar(vr VarInScope) {
	s.DeclaredVars.Vars[vr.Name] = vr
}

func (s *ScopeCtxt) AddDeclHeader(n string) {
	s.CurDeclHeaders.Headers[n] = true
}

func (s *ScopeCtxt) AddDefnHeader(n string) {
	s.CurDefnHeaders.Headers[n] = true
}

func (s *ScopeCtxt) VarType(name string) VarType {
	c := s
	for c.HasParent() {
		for n, v := range c.DeclaredVars.Vars {
			if n == name {
				return v.Type
			}
		}
		c = c.parentScope
	}
	panic(fmt.Sprintf("asked for VarType of undefined var %v", name))
}

func (s *ScopeCtxt) HasField(name string) bool {
	for n, v := range s.DeclaredVars.Vars {
		if n == name && v.Scope == VarScopeField {
			return true
		}
	}
	return false
}

func (s *ScopeCtxt) HasLocal(name string) bool {
	for n, v := range s.DeclaredVars.Vars {
		if n == name && v.Scope == VarScopeLocal {
			return true
		}
	}

	return false
}

func (s *ScopeCtxt) HasGlobal(name string) bool {
	r := false
	c := s
	for c.HasParent() {
		c = c.parentScope
	}
	for n, v := range c.DeclaredVars.Vars {
		if n == name && v.Scope == VarScopeGlobal {
			r = true
		}
	}

	return r
}

func (s *ScopeCtxt) HasGlobalProc(name string) bool {
	c := s
	for c.HasParent() {
		c = c.parentScope
	}
	for n := range c.DeclaredProcs.Procs {
		if n == name {
			return true
		}
	}

	return false
}

func (s *ScopeCtxt) HasParent() bool {
	return s.parentScope != nil
}

func (s *ScopeCtxt) String() string {
	return fmt.Sprintf("ScopeContext<%v depth=%v parent=%v proc=%v varCount=%v procCount=%v>",
		s.CurPath, s.currentDepth, s.parentScope != nil, s.CurProc, len(s.DeclaredVars.Vars), len(s.DeclaredProcs.Procs))
}

func MakeRoot() *ScopeCtxt {
	root := &ScopeCtxt{
		CurPath:        paths.New("/"),
		CurDefnHeaders: NewHeaderCollection(),
		CurDeclHeaders: NewHeaderCollection(),
		DeclaredVars:   NewDeclaredVars(),
		DeclaredProcs:  NewDeclaredProcs(),
	}

	return root
}

func (s *ScopeCtxt) MakeChildPath(p *paths.Path) *ScopeCtxt {
	child := &ScopeCtxt{
		CurDeclFile:    s.CurDeclFile,
		CurDefnFile:    s.CurDefnFile,
		curClassDecl:   s.curClassDecl,
		CurPath:        p,
		CurType:        s.CurType,
		CurProc:        s.CurProc,
		CurDefnHeaders: s.CurDefnHeaders,
		CurDeclHeaders: s.CurDeclHeaders,
		DeclaredVars:   NewDeclaredVars(),
		DeclaredProcs:  NewDeclaredProcs(),
		parentScope:    s,
		currentDepth:   s.currentDepth + 1,
	}

	return child
}

func (s *ScopeCtxt) MakeChild() *ScopeCtxt {
	child := &ScopeCtxt{
		CurDeclFile:    s.CurDeclFile,
		CurDefnFile:    s.CurDefnFile,
		curClassDecl:   s.curClassDecl,
		CurPath:        s.CurPath,
		CurType:        s.CurType,
		CurProc:        s.CurProc,
		CurDefnHeaders: s.CurDefnHeaders,
		CurDeclHeaders: s.CurDeclHeaders,
		DeclaredVars:   NewDeclaredVars(),
		DeclaredProcs:  NewDeclaredProcs(),
		parentScope:    s,
		currentDepth:   s.currentDepth + 1,
	}

	return child
}

type Stack struct {
	Scopes []*ScopeCtxt
}

func (s Stack) String() string {
	return fmt.Sprintf("<ScopeStack@%p curScope=%v/%v>", &s, s.LastScope(), s.Len())
}

func NewStack() *Stack {
	return &Stack{}
}

func (s *Stack) Len() int {
	return len(s.Scopes)
}

func (s *Stack) LastScope() *ScopeCtxt {
	if len(s.Scopes) == 0 {
		return nil
	}
	return s.Scopes[len(s.Scopes)-1]
}

func (s *Stack) Pop() *ScopeCtxt {
	if len(s.Scopes) == 0 {
		return nil
	}

	result := s.Scopes[len(s.Scopes)-1]
	s.Scopes = s.Scopes[0:len(s.Scopes)]
	return result
}

func (s *Stack) Push(value *ScopeCtxt) {
	s.Scopes = append(s.Scopes, value)
}
