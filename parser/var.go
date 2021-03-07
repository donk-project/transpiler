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

type DMVar struct {
	State *Parser
	Type  *DMType
	Name  string
	Proto *astpb.TypeVar
	Path  *paths.Path
}

type DMVars []*DMVar

func (slice DMVars) Len() int {
	return len(slice)
}

func (slice DMVars) Less(i, j int) bool {
	return slice[i].VarName() < slice[j].VarName()
}

func (slice DMVars) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (v DMVar) String() string {
	return fmt.Sprintf("Var<%v>", v.Path)
}

func (v *DMVar) VarName() string {
	return v.Path.Basename
}

func NewVar(s *Parser, t *DMType, n string, p *astpb.TypeVar) *DMVar {
	v := &DMVar{State: s, Type: t, Name: n, Proto: p}
	return v
}

func CachedVar(s *Parser, p *paths.Path) *DMVar {
	v := &DMVar{State: s, Name: p.Basename, Path: p}
	return v
}

func (v *DMVar) HasStaticValue() bool {
	if v.Proto.GetValue().GetExpression() != nil {
		return true
	}
	if v.Proto.GetValue().GetConstant() != nil {
		if v.Proto.GetValue().GetConstant().GetNull() != nil {
			return false
		}
		return true
	}
	return false
}

func (d *DMVar) DeclarationTypePath() *paths.Path {
	if d.Proto.GetDeclaration().GetVarType().GetTypePath() != nil {
		return paths.NewFromTreePath(d.Proto.Declaration.VarType.TypePath)
	}
	return nil
}

func (p *DMVar) PrettyProto() string {
	return proto.MarshalTextString(p.Proto)
}
