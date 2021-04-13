// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) procCall(
	callee *cctpb.Expression,
	procName string,
	args []*cctpb.Expression,
	name string) *cctpb.Expression {
	fc := vsk.FuncCall(vsk.Id(name))
	vsk.AddFuncArg(fc.GetFunctionCallExpression(), callee)
	vsk.AddFuncArg(fc.GetFunctionCallExpression(), vsk.StringLiteralExpr(procName))
	vsk.AddFuncInitListArg(fc.GetFunctionCallExpression(), args...)
	ctxt := vsk.StringIdExpr("ctxt")
	mae := vsk.ObjMember(ctxt, fc)
	return mae
}
