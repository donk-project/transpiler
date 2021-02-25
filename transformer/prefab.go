// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"strings"

	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t *Transformer) walkPrefab(p *astpb.Prefab) *cctpb.Expression {
	if p == nil || p.GetPath() == nil {
		panic("cannot walk nil prefab")
	}
	var parts []string
	for idx, tp := range p.GetPath() {
		if idx == 0 {
			if !strings.HasPrefix(*tp.S, "/") {
				parts = append(parts, "/"+*tp.S)
				continue
			}
		}

		parts = append(parts, *tp.S)
	}
	fqp := paths.New(strings.Join(parts, "/"))
	return pathExpression(*fqp)
}
