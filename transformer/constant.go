// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	"github.com/golang/protobuf/proto"

	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) walkConstant(c *astpb.Constant) *cctpb.Expression {
	switch {
	case c.StringConstant != nil:
		{
			t.curScope.addDefnHeader("<string>")
			return stdStringCtor(c.GetStringConstant())
		}
	case c.Resource != nil:
		{
			return resourceId(c.GetResource())
		}
	case c.Int != nil:
		{
			return &cctpb.Expression{
				Value: &cctpb.Expression_LiteralExpression{
					&cctpb.Literal{
						Value: &cctpb.Literal_IntegerLiteral{
							int64(c.GetInt()),
						},
					},
				},
			}
		}
	case c.GetPrefab() != nil:
		{
			t.curScope.addDefnHeader("\"donk/core/vars.h\"")
			pf := paths.NewFromTreePath(c.GetPrefab().GetPop().GetTreePath())
			return prefabExpression(*pf)
		}
	}

	panic(fmt.Sprintf("cannot walk unknown constant %v", proto.MarshalTextString(c)))
}
