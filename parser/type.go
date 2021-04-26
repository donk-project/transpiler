// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package parser

import (
	"fmt"
	// "log"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
)

type DMType struct {
	State          *Parser
	Path           paths.Path
	Proto          *astpb.Type
	Dependencies   []paths.Path
	AdditionalDeps []paths.Path
	Procs          []*DMProc
	Vars           DMVars
	ForwardDecls   map[string]paths.Path
}

func NewType(s *Parser, p paths.Path, pb *astpb.Type) *DMType {
	t := &DMType{State: s, Path: p, Proto: pb}
	t.ForwardDecls = make(map[string]paths.Path)
	return t
}

func (t *DMType) IsProcRegistered(name string) bool {
	for _, p := range t.Procs {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (t *DMType) Proc(s string) *DMProc {
	for _, p := range t.Procs {
		if p.Name == s {
			return p
		}
	}
	return nil
}

func (t *DMType) String() string {
	return fmt.Sprintf("Type<%v>", t.Path)
}

func (p *DMType) PrettyProto() string {
	return proto.MarshalTextString(p.Proto)
}

func (t *DMType) ParentType() paths.Path {
	for _, v := range t.Vars {
		if v.Name == "parent_type" {
			if v.Proto.GetValue().GetConstant().GetPrefab().GetPop().GetTreePath() != nil {
				return paths.NewFromTreePath(v.Proto.GetValue().GetConstant().GetPrefab().GetPop().GetTreePath())
			}
		}
	}
	return t.Path.ParentPath()
}

func (t *DMType) ResolvedPath() paths.Path {
	return t.ParentType().Child(t.Path.Basename)
}
