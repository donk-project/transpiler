// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package parser

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
)

func (p *DMProc) String() string {
	return fmt.Sprintf("Proc<%v%v>", p.Type.Path.FullyQualifiedString(), p.Name)
}

type DMProc struct {
	State *Parser
	Type  *DMType
	Name  string
	Proto *astpb.TypeProc
	Decl  []string
}

func (p *DMProc) ArgNames() []string {
	var argNames []string
	for _, procVal := range p.Proto.GetValue() {
		for _, param := range procVal.GetParameter() {
			argNames = append(argNames, *param.Name)
		}
	}
	return argNames
}

func (p *DMProc) HasArg(s string) bool {
	for _, n := range p.ArgNames() {
		if s == n {
			return true
		}
	}
	return false
}

// EmitName deals with symbol collisions with C++, or collisions
// between namespaces (such as donk::icon) with core procs.
func (p *DMProc) EmitName() string {
	if p.Name == "new" {
		return "new_"
	}
	if p.ProcPath().FullyQualifiedString() == "/icon" {
		return "icon_"
	}
	if p.ProcPath().FullyQualifiedString() == "/matrix" {
		return "matrix_"
	}
	if p.ProcPath().FullyQualifiedString() == "/sound" {
		return "sound_"
	}
	if p.ProcPath().FullyQualifiedString() == "/regex" {
		return "regex_"
	}
	if p.ProcPath().FullyQualifiedString() == "/image" {
		return "image_"
	}
	if p.ProcPath().FullyQualifiedString() == "/list" {
		return "list_"
	}
	return p.Name
}

func (p *DMProc) PrettyProto() string {
	return proto.MarshalTextString(p.Proto)
}

func NewProc(s *Parser, t *DMType, n string, pb *astpb.TypeProc) *DMProc {
	p := &DMProc{State: s, Type: t, Name: n, Proto: pb}
	return p
}

func (d *DMProc) Block(idx int) *astpb.Block {
	return d.Proto.Value[idx].Code.Present
}

func (d *DMProc) ProcPath() *paths.Path {
	return d.Type.Path.Child(d.Name)
}

func (p *Parser) ParseProc(name string, dmType *DMType, pb *astpb.TypeProc) *DMProc {
	proc := NewProc(p, dmType, name, pb)
	return proc
}

func (p *Parser) ParseProcs(g *astpb.Graph, tbp *map[paths.Path]*DMType) {
	for _, dmType := range *tbp {
		for k, pb := range dmType.Proto.GetProcs() {
			if !dmType.IsProcRegistered(k) {
				proc := NewProc(p, dmType, k, pb)
				p.ProcsByPath[*proc.ProcPath()] = proc
				dmType.Procs = append(dmType.Procs, proc)
			}
		}
	}
}
