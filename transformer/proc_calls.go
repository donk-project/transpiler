// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	// "fmt"
	// "log"

	// "github.com/golang/protobuf/proto"
	// "snowfrost.garden/donk/transpiler/scope"

	// astpb "snowfrost.garden/donk/proto/ast"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

type InvokeType int

const (
	InvokeTypeAsyncProc InvokeType = iota
	InvokeTypeSyncProc
	InvokeTypeChildProc
)

var PROC_NAMES = map[InvokeType]string{
	InvokeTypeAsyncProc: "Proc",
	InvokeTypeSyncProc:  "SProc",
	InvokeTypeChildProc: "ChildProc",
}

func (t Transformer) procCall(
	callee *cctpb.Expression,
	procName string,
	args []*cctpb.Expression,
	invokeType InvokeType) *cctpb.Expression {
	fc := vsk.FuncCall(vsk.Id(PROC_NAMES[invokeType]))
	vsk.AddFuncArg(fc.GetFunctionCallExpression(), callee)
	vsk.AddFuncArg(fc.GetFunctionCallExpression(), vsk.StringLiteralExpr(procName))
	vsk.AddFuncInitListArg(fc.GetFunctionCallExpression(), args...)
	ctxt := vsk.StringIdExpr("ctxt")
	mae := vsk.ObjMember(ctxt, fc)
	return mae
}
