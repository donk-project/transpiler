// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func printStatement(s *cctpb.Statement) string {
	// log.Printf("statement: %v", proto.MarshalTextString(s))
	switch x := s.Value.(type) {
	case *cctpb.Statement_ExpressionStatement:
		return printExpression(s.GetExpressionStatement())
	case *cctpb.Statement_CompoundStatement:
		return printCompoundStatement(s.GetCompoundStatement())
	case *cctpb.Statement_IfStatement:
		return printIfStatement(s.GetIfStatement())
	case *cctpb.Statement_DeclarationStatement:
		return printDeclaration(s.GetDeclarationStatement())
	case nil:
		panic(fmt.Sprintf("Hey: %v", proto.MarshalTextString(s)))
	default:
		panic(fmt.Errorf("expression has unexpected type %T", x))
	}
}

func printCompoundStatement(s *cctpb.CompoundStatement) string {
	return ""
}

func printIfStatement(s *cctpb.IfStatement) string {
	return ""
}
