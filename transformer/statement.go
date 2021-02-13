// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) walkStatement(s *astpb.Statement) *cctpb.Statement {
	stmt := &cctpb.Statement{}
	if s == nil {
		return stmt
	}
	switch {
	case s.GetExpr() != nil:
		{
			stmt.Value = &cctpb.Statement_ExpressionStatement{t.walkExpression(s.GetExpr().GetExpr())}
			return stmt
		}
	case s.GetDel() != nil:
		{
			mae := genericCtxtCall("del")

			addFuncExprArg(
				mae.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(),
				t.walkExpression(s.GetDel().GetExpr()))

			stmt.Value = &cctpb.Statement_ExpressionStatement{mae}
			return stmt
		}
	case s.GetVar() != nil:
		{
			if s.GetVar().GetValue().GetBase() != nil {
				vr := varRepresentation{
					name:     s.GetVar().GetName(),
					varScope: VarScopeLocal,
				}
				if isDeclaringNewDMObject(s.GetVar().GetValue()) {
					vr.varType = VarTypeDMObject
				}
				t.curScope.AddScopedVar(vr)
				return t.declareVarWithVal(
					s.GetVar().GetName(),
					t.walkExpression(s.GetVar().GetValue()),
					vr.varType,
				)
			}
			return declareVar(s.GetVar().GetName())
		}
	}
	panic(fmt.Sprintf("cannot walk unsupported statement %v", proto.MarshalTextString(s)))
}
