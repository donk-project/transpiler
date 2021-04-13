// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package parser

import (
	"fmt"
	// "log"
	"strings"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
)

type DMType struct {
	State          *Parser
	Path           *paths.Path
	Proto          *astpb.Type
	Dependencies   []*paths.Path
	AdditionalDeps []*paths.Path
	Procs          []*DMProc
	Vars           DMVars
	ForwardDecls   map[string]*paths.Path
}

func NewType(s *Parser, p *paths.Path, pb *astpb.Type) *DMType {
	t := &DMType{State: s, Path: p, Proto: pb}
	t.ForwardDecls = make(map[string]*paths.Path)
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

func (d *DMType) AllForwardDecls() map[string][]*paths.Path {
	x := make(map[string][]*paths.Path)

	for _, dep := range d.AdditionalDeps {
		if dep == d.Path {
			continue
		}
		if !strings.HasPrefix(dep.Name, "/list") {
			x[strings.TrimPrefix(dep.AsNamespace(), "::")] = append(x[dep.ParentPath().Name], dep)
		}
	}

	return x
}
