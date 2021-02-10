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

func arithmeticOpFromBinaryOp(op astpb.BinaryOp) cctpb.ArithmeticExpression_Operator {
	switch {
	case op == astpb.BinaryOp_BINARYOP_ADD:
		return cctpb.ArithmeticExpression_ADDITION
	case op == astpb.BinaryOp_BINARYOP_SUB:
		return cctpb.ArithmeticExpression_SUBTRACTION
	case op == astpb.BinaryOp_BINARYOP_MUL:
		return cctpb.ArithmeticExpression_MULTIPLICATION
	case op == astpb.BinaryOp_BINARYOP_DIV:
		return cctpb.ArithmeticExpression_DIVISION
	case op == astpb.BinaryOp_BINARYOP_MOD:
		return cctpb.ArithmeticExpression_MODULO
	case op == astpb.BinaryOp_BINARYOP_BITAND:
		return cctpb.ArithmeticExpression_BITWISE_AND
	case op == astpb.BinaryOp_BINARYOP_BITOR:
		return cctpb.ArithmeticExpression_BITWISE_OR
	case op == astpb.BinaryOp_BINARYOP_BITXOR:
		return cctpb.ArithmeticExpression_BITWISE_XOR
	case op == astpb.BinaryOp_BINARYOP_LSHIFT:
		return cctpb.ArithmeticExpression_BITWISE_LSHIFT
	case op == astpb.BinaryOp_BINARYOP_RSHIFT:
		return cctpb.ArithmeticExpression_BITWISE_RSHIFT
	default:
		panic(fmt.Sprintf("no arithmetic op matching binary op %v", op))
	}
}

func comparisonOpFromBinaryOp(op astpb.BinaryOp) cctpb.ComparisonExpression_Operator {
	switch {
	case op == astpb.BinaryOp_BINARYOP_EQ:
		return cctpb.ComparisonExpression_EQUAL_TO
	case op == astpb.BinaryOp_BINARYOP_NOTEQ:
		return cctpb.ComparisonExpression_NOT_EQUAL_TO
	case op == astpb.BinaryOp_BINARYOP_LESS:
		return cctpb.ComparisonExpression_LESS_THAN
	case op == astpb.BinaryOp_BINARYOP_GREATER:
		return cctpb.ComparisonExpression_GREATER_THAN
	case op == astpb.BinaryOp_BINARYOP_LESSEQ:
		return cctpb.ComparisonExpression_LESS_THAN_OR_EQUAL_TO
	case op == astpb.BinaryOp_BINARYOP_GREATEREQ:
		return cctpb.ComparisonExpression_GREATER_THAN_OR_EQUAL_TO
	default:
		panic(fmt.Sprintf("no comparison op matching binary op %v", op))
	}
}

func (t *Transformer) BinaryOpToExpr(op astpb.BinaryOp) *cctpb.Expression {
	switch {
	case op == astpb.BinaryOp_BINARYOP_POW:
		{
			// No native power call so this become's <cmath>'s std::pow
			t.curScope.addDefnHeader("<cmath>")
			id := &cctpb.Identifier{
				Namespace: proto.String("std"),
				Id:        proto.String("pow"),
			}
			fce := &cctpb.FunctionCallExpression{
				Name: &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{id},
				},
			}
			return &cctpb.Expression{
				Value: &cctpb.Expression_FunctionCallExpression{fce},
			}
		}

	case op == astpb.BinaryOp_BINARYOP_ADD:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_SUB:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_MUL:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_DIV:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_MOD:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_BITAND:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_BITOR:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_BITXOR:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_LSHIFT:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_RSHIFT:
		{
			aOp := arithmeticOpFromBinaryOp(op)
			ae := &cctpb.ArithmeticExpression{
				Operator: aOp.Enum(),
			}
			return &cctpb.Expression{
				Value: &cctpb.Expression_ArithmeticExpression{ae},
			}
		}

	case op == astpb.BinaryOp_BINARYOP_EQ:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_NOTEQ:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_LESS:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_GREATER:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_LESSEQ:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_GREATEREQ:
		{
			cOp := comparisonOpFromBinaryOp(op)
			ce := &cctpb.ComparisonExpression{
				Operator: cOp.Enum(),
			}
			return &cctpb.Expression{
				Value: &cctpb.Expression_ComparisonExpression{ce},
			}
		}

	case op == astpb.BinaryOp_BINARYOP_EQUIV:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_NOTEQUIV:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_AND:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_OR:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_IN:
		fallthrough
	case op == astpb.BinaryOp_BINARYOP_TO:
		fallthrough
	default:
		panic(fmt.Sprintf("cannot walk unsupported binary op %v", op))
	}
}

func ConvertAssignOp(op astpb.AssignOp) cctpb.AssignmentExpression_Operator {
	if op == astpb.AssignOp_ASSIGNOP_ASSIGN {
		return cctpb.AssignmentExpression_SIMPLE
	} else if op == astpb.AssignOp_ASSIGNOP_ADD_ASSIGN {
		return cctpb.AssignmentExpression_ADDITION
	} else if op == astpb.AssignOp_ASSIGNOP_SUB_ASSIGN {
		return cctpb.AssignmentExpression_SUBTRACTION
	} else if op == astpb.AssignOp_ASSIGNOP_MUL_ASSIGN {
		return cctpb.AssignmentExpression_MULTIPLICATION
	} else if op == astpb.AssignOp_ASSIGNOP_DIV_ASSIGN {
		return cctpb.AssignmentExpression_DIVISION
	} else if op == astpb.AssignOp_ASSIGNOP_MOD_ASSIGN {
		return cctpb.AssignmentExpression_MODULO
	} else if op == astpb.AssignOp_ASSIGNOP_BIT_AND_ASSIGN {
		return cctpb.AssignmentExpression_BITWISE_AND
	} else if op == astpb.AssignOp_ASSIGNOP_BIT_OR_ASSIGN {
		return cctpb.AssignmentExpression_BITWISE_OR
	} else if op == astpb.AssignOp_ASSIGNOP_BIT_XOR_ASSIGN {
		return cctpb.AssignmentExpression_BITWISE_XOR
	} else if op == astpb.AssignOp_ASSIGNOP_L_SHIFT_ASSIGN {
		return cctpb.AssignmentExpression_BITWISE_LSHIFT
	} else if op == astpb.AssignOp_ASSIGNOP_R_SHIFT_ASSIGN {
		return cctpb.AssignmentExpression_BITWISE_RSHIFT
	}

	panic(fmt.Sprintf("cannot convert unknown assignop %v", op))
}
