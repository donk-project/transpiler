// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func printExpression(e *cctpb.Expression) string {
	switch x := e.Value.(type) {
	case *cctpb.Expression_AssignmentExpression:
		return printAssignmentExpression(e.GetAssignmentExpression())
	case *cctpb.Expression_UnaryExpression:
		return printUnaryExpression(e.GetUnaryExpression())
	case *cctpb.Expression_ArithmeticExpression:
		return printArithmeticExpression(e.GetArithmeticExpression())
	case *cctpb.Expression_ComparisonExpression:
		return printComparisonExpression(e.GetComparisonExpression())
	case *cctpb.Expression_MemberAccessExpression:
		return printMemberAccessExpression(e.GetMemberAccessExpression())
	case *cctpb.Expression_FunctionCallExpression:
		return printFunctionCallExpression(e.GetFunctionCallExpression())
	case *cctpb.Expression_LiteralExpression:
		return printLiteral(e.GetLiteralExpression())
	case *cctpb.Expression_IdentifierExpression:
		return printIdentifier(e.GetIdentifierExpression())
	case nil:
		panic("cannot print nil expression")
	default:
		panic(fmt.Errorf("expression has unexpected type %T", x))
	}
}

func printAssignmentExpression(e *cctpb.AssignmentExpression) string {
	return fmt.Sprintf(
		ASSIGNMENT_OP_FORMATTERS[e.GetOperator()],
		printExpression(e.GetLhs()),
		printExpression(e.GetRhs()))
}

func printUnaryExpression(e *cctpb.UnaryExpression) string {
	return fmt.Sprintf(UNARY_OP_FORMATTERS[e.GetOperator()],
		printExpression(e.GetOperand()))
}

func printArithmeticExpression(e *cctpb.ArithmeticExpression) string {
	return fmt.Sprintf(ARITHMETIC_OP_FORMATTERS[e.GetOperator()],
		printExpression(e.GetLhs()), printExpression(e.GetRhs()))
}

func printComparisonExpression(e *cctpb.ComparisonExpression) string {
	return fmt.Sprintf(COMPARISON_OP_FORMATTERS[e.GetOperator()],
		printExpression(e.GetLhs()), printExpression(e.GetRhs()))
}

func printMemberAccessExpression(e *cctpb.MemberAccessExpression) string {
	return fmt.Sprintf(MEMBER_ACCESS_EXPRESSION_OP_FORMATTERS[e.GetOperator()],
		printExpression(e.GetLhs()), printExpression(e.GetRhs()))
}

func printFunctionCallExpression(e *cctpb.FunctionCallExpression) string {
	var args []string
	for _, ex := range e.GetArguments() {
		switch ex.GetValue().(type) {
		case *cctpb.FunctionCallExpression_ExpressionArg_Expression:
			{
				args = append(args, printExpression(ex.GetExpression()))
			}
		case *cctpb.FunctionCallExpression_ExpressionArg_InitializerList:
			{
				iList := ex.GetInitializerList()
				var iListArgs []string
				for _, arg := range iList.GetArgs() {
					iListArgs = append(iListArgs, printExpression(arg))
				}
				args = append(args, "{"+strings.Join(iListArgs, ", ")+"}")
			}
		default:
			panic(fmt.Sprintf("cannot print unsupported function-call-expr arg %v", proto.MarshalTextString(ex)))
		}
	}
	return fmt.Sprintf("%v(%v)", printExpression(e.GetName()), strings.Join(args, ", "))
}

func printLiteral(e *cctpb.Literal) string {
	switch x := e.Value.(type) {
	case *cctpb.Literal_IntegerLiteral:
		return strconv.FormatInt(e.GetIntegerLiteral(), 10)
	case *cctpb.Literal_CharacterLiteral:
		return strconv.FormatInt(e.GetCharacterLiteral(), 10)
	case *cctpb.Literal_FloatLiteral:
		return strconv.FormatFloat(float64(e.GetFloatLiteral()), 'f', 0, 32)
	case *cctpb.Literal_StringLiteral:
		return fmt.Sprintf("\"%v\"", e.GetStringLiteral())
	case *cctpb.Literal_BooleanLiteral:
		return strconv.FormatBool(e.GetBooleanLiteral())
	case nil:
		panic("cannot print nil literal")
	default:
		panic(fmt.Errorf("literal has unexpected type %T", x))
	}
}

func printIdentifier(i *cctpb.Identifier) string {
	if i.Namespace != nil {
		return fmt.Sprintf("%v::%v", i.GetNamespace(), i.GetId())
	}
	return i.GetId()
}
