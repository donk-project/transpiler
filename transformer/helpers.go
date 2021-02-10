// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
	"snowfrost.garden/donk/transpiler/paths"
)

func genericCtxtCall(n string) *cctpb.Expression {
	mae := &cctpb.MemberAccessExpression{
		Operator: cctpb.MemberAccessExpression_MEMBER_OF_OBJECT.Enum(),
	}

	ctxt := &cctpb.Identifier{
		Id: proto.String("ctxt"),
	}
	lhs := &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{ctxt},
	}

	src := &cctpb.Identifier{
		Id: proto.String(n),
	}
	srcCall := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{src},
		},
	}
	rhs := &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{srcCall},
	}

	mae.Lhs = lhs
	mae.Rhs = rhs

	expr := &cctpb.Expression{
		Value: &cctpb.Expression_MemberAccessExpression{mae},
	}
	return expr
}

func ctxtSrc() *cctpb.Expression {
	return genericCtxtCall("src")
}

func getObjVar(v string) *cctpb.Expression {
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{
				&cctpb.Identifier{Id: proto.String("v")},
			},
		},
	}
	l := &cctpb.Literal{
		Value: &cctpb.Literal_StringLiteral{v},
	}
	addFuncExprArg(fc, &cctpb.Expression{Value: &cctpb.Expression_LiteralExpression{l}})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func stdStringCtor(s string) *cctpb.Expression {
	i := &cctpb.Identifier{
		Namespace: proto.String("std"),
		Id:        proto.String("string"),
	}
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{&cctpb.Literal{
			Value: &cctpb.Literal_StringLiteral{s}}},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func resourceId(resource string) *cctpb.Expression {
	i := &cctpb.Identifier{
		Namespace: proto.String("donk"),
		Id:        proto.String("resource_t"),
	}
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{&cctpb.Literal{
			Value: &cctpb.Literal_StringLiteral{resource}}},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func pathExpression(p paths.Path) *cctpb.Expression {
	i := &cctpb.Identifier{
		Namespace: proto.String("donk"),
		Id:        proto.String("path_t"),
	}
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{&cctpb.Literal{
			Value: &cctpb.Literal_StringLiteral{p.FullyQualifiedString()}}},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func ctxtUsr() *cctpb.Expression {
	return genericCtxtCall("usr")
}

func argParam() *cctpb.Expression {
	return &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{
			&cctpb.Identifier{
				Id: proto.String("args"),
			},
		},
	}
}

func coreProcCall(name string) *cctpb.Expression {
	core := genericCtxtCall("core")
	fc := core.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression()
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{&cctpb.Identifier{
			Id: proto.String(fmt.Sprintf("\"%v\"", name)),
		}},
	})
	return core
}

func isVarAssignFromPtr(e *cctpb.Expression) bool {
	// log.Printf("isVarAssignFromPtr: %v", proto.MarshalTextString(e))
	if e.GetMemberAccessExpression() == nil {
		return false
	}

	mae := e.GetMemberAccessExpression()
	if !proto.Equal(mae.GetLhs(), ctxtSrc()) && !proto.Equal(mae.GetLhs(), ctxtUsr()) {
		return false
	}

	if mae.GetRhs().GetFunctionCallExpression().GetName().GetIdentifierExpression().GetId() == "v" {
		return true
	}

	return false
}

func isFmtRedirected(e *cctpb.Expression) bool {
	// log.Printf("isFmtRedirected: %v", proto.MarshalTextString(e))
	if proto.Equal(e, ctxtSrc()) || proto.Equal(e, ctxtUsr()) {
		return true
	}
	// just checks for "args", should probably check for "args#v"
	if e.GetMemberAccessExpression().GetLhs().GetIdentifierExpression().GetId() == "args" {
		return true
	}
	return false
}

func wrapFuncCallExpr(e *cctpb.Expression) *cctpb.FunctionCallExpression_ExpressionArg {
	return &cctpb.FunctionCallExpression_ExpressionArg{
		Value: &cctpb.FunctionCallExpression_ExpressionArg_Expression{e},
	}
}

func addFuncExprArg(fce *cctpb.FunctionCallExpression, e *cctpb.Expression) {
	fce.Arguments = append(fce.Arguments, wrapFuncCallExpr(e))
}

func addFuncInitListArg(fce *cctpb.FunctionCallExpression, exprs ...*cctpb.Expression) {
	iList := &cctpb.InitializerList{}
	for _, expr := range exprs {
		iList.Args = append(iList.Args, expr)
	}
	fce.Arguments = append(fce.Arguments, &cctpb.FunctionCallExpression_ExpressionArg{
		Value: &cctpb.FunctionCallExpression_ExpressionArg_InitializerList{iList},
	})
}
