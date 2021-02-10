// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package parser

import (
	"sort"

	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
)

type Stats struct {
	ReferenceCount map[paths.Path]int
}

type Parser struct {
	Graph         *astpb.Graph
	TypesByPath   map[paths.Path]*DMType
	ProcsByPath   map[paths.Path]*DMProc
	VarsByPath    map[paths.Path]*DMVar
	FieldTable    map[paths.Path]bool
	FunctionTable map[paths.Path]bool
	Stats         Stats
}

func (p *Parser) ParseTypes(g *astpb.Graph, tbp *map[paths.Path]*DMType) {
	for _, t := range g.GetType() {
		if *t.Name == "" {
			(*tbp)[*paths.New("/")] = NewType(p, paths.New("/"), t)
			continue
		}
		iterpath := paths.New(*t.Path)
		p.Stats.ReferenceCount[*iterpath]++
		dmt := NewType(p, iterpath, t)
		(*tbp)[*iterpath] = dmt
		if !iterpath.ParentPath().IsRoot() {
			dmt.Dependencies = append(dmt.Dependencies, paths.New(iterpath.ParentPath().Name))
		}
	}
}

func (p *Parser) ParseVars(g *astpb.Graph, tbp *map[paths.Path]*DMType) {
	for path, dmType := range *tbp {
		protos := make(map[string]*astpb.TypeVar)
		for k, pv := range dmType.Proto.GetVars() {
			protos[k] = pv
			vv := NewVar(p, dmType, k, pv)
			vv.Path = path.Child(k)
			p.VarsByPath[*vv.Path] = vv
			if !vv.Type.Path.IsRoot() {
				p.Stats.ReferenceCount[*vv.Type.Path]++
			}
			dmType.Vars = append(dmType.Vars, vv)
		}
	}
}

func NewParser(graph *astpb.Graph) *Parser {
	var d Parser
	d.TypesByPath = make(map[paths.Path]*DMType)
	d.VarsByPath = make(map[paths.Path]*DMVar)
	d.ProcsByPath = make(map[paths.Path]*DMProc)
	d.Stats.ReferenceCount = make(map[paths.Path]int)
	d.Graph = graph

	d.ParseTypes(d.Graph, &d.TypesByPath)
	d.ParseVars(d.Graph, &d.TypesByPath)
	d.ParseProcs(d.Graph, &d.TypesByPath)

	for _, t := range d.TypesByPath {
		sort.Sort(t.Vars)
		sort.Slice(t.Procs, func(i, j int) bool {
			return t.Procs[i].Name < t.Procs[j].Name
		})
	}

	return &d
}
