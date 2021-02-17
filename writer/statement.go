// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func printStatement(s *cctpb.Statement) string {
	// fmt.Printf("==============================\nCCTPB Statement:\n %v==============================\n", proto.MarshalTextString(s))
	switch s.Value.(type) {
	case *cctpb.Statement_ExpressionStatement:
		return printExpression(s.GetExpressionStatement()) + ";"
	case *cctpb.Statement_DeclarationStatement:
		return printDeclaration(s.GetDeclarationStatement()) + ";"
	case *cctpb.Statement_DoWhile:
		return printDoWhile(s.GetDoWhile())
	case *cctpb.Statement_RangeBasedFor:
		return printRangeBasedFor(s.GetRangeBasedFor())
	case *cctpb.Statement_ReturnStatement:
		return printReturnStatement(s.GetReturnStatement())
	case *cctpb.Statement_IfStatement:
		return printIfStatement(s.GetIfStatement())
	case *cctpb.Statement_CompoundStatement:
		return printCompoundStatement(s.GetCompoundStatement())
	case nil:
		panic("nil statment")
	default:
		panic(fmt.Sprintf("cannot print unsupported statement %v", proto.MarshalTextString(s)))
	}
}

func printReturnStatement(r *cctpb.ReturnStatement) string {
	switch r.Value.(type) {
	case *cctpb.ReturnStatement_Expression:
		return fmt.Sprintf("return %v", printExpression(r.GetExpression()))
	}
	panic(fmt.Sprintf("cannot print unsupported return %v", proto.MarshalTextString(r)))
}

func printCompoundStatement(s *cctpb.CompoundStatement) string {
	var stmts []string
	for _, stmt := range s.GetStatements() {
		stmts = append(stmts, printStatement(stmt))
	}
	return strings.Join(stmts, "\n")
}

func printIfStatement(s *cctpb.IfStatement) string {
	expr := printExpression(s.GetCondition())
	sT := printStatement(s.GetStatementTrue())

	if s.GetStatementFalse() != nil {
		sF := printStatement(s.GetStatementFalse())
		return fmt.Sprintf("if (%v) {\n%v\n} else {\n%v\n}\n", expr, sT, sF)
	}

	return fmt.Sprintf("if (%v) {\n%v\n}", expr, sT)
}

func printDoWhile(s *cctpb.DoWhile) string {
	var stmts []string
	for _, stmt := range s.GetBlockDefinition().GetStatements() {
		stmts = append(stmts, printStatement(stmt))
	}

	return fmt.Sprintf("do {\n%v\n} while (%v);\n",
		strings.Join(stmts, "\n"),
		printExpression(s.GetCondition()))
}

func printRangeBasedFor(rbf *cctpb.RangeBasedFor) string {
	var stmts []string
	for _, stmt := range rbf.GetLoopDefinition().GetStatements() {
		stmts = append(stmts, printStatement(stmt))
	}
	decl := printDeclaration(rbf.GetDeclaration())
	rangeExpr := printExpression(rbf.GetRangeExpression())

	return fmt.Sprintf("for (%v : %v) {\n%v\n}\n", decl, rangeExpr, strings.Join(stmts, "\n"))
}
