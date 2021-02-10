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

func (t Transformer) walkExpression(e *astpb.Expression) *cctpb.Expression {
	if e == nil {
		panic("tried to walk empty expression")
	}

	switch {
	case e.GetBase() != nil:
		{
			if e.GetBase().GetTerm() != nil {
				if len(e.GetBase().GetFollow()) > 0 {
					term := t.walkTerm(e.GetBase().GetTerm())
					curFollow := &cctpb.MemberAccessExpression{}
					var follows []*cctpb.MemberAccessExpression
					for _, follow := range e.GetBase().GetFollow() {
						call := follow.GetCall()
						if call != nil {
							// TODO: Flesh out call transformation
						} else if follow.GetField() != nil {
							field := follow.GetField()
							if field.GetIndexKind() == astpb.IndexKind_INDEX_KIND_DOT {
								curFollow.Operator = cctpb.MemberAccessExpression_MEMBER_OF_OBJECT.Enum()
								if field.S != nil {
									if proto.Equal(term, ctxtSrc()) || proto.Equal(term, ctxtUsr()) {
										curFollow.Rhs = getObjVar(field.GetS())
									} else {
										curFollow.Rhs = &cctpb.Expression{
											Value: &cctpb.Expression_IdentifierExpression{
												&cctpb.Identifier{
													Id: field.S,
												},
											},
										}
									}
								}
							} else {
								panic(fmt.Sprintf("unsupported index kind %v", proto.MarshalTextString(follow)))
							}
							follows = append(follows, curFollow)
							curFollow = &cctpb.MemberAccessExpression{}
						}
					}

					var result *cctpb.MemberAccessExpression
					for len(follows) > 0 {
						var head *cctpb.MemberAccessExpression
						head, follows = follows[0], follows[1:]
						if result == nil {
							if proto.Equal(term, ctxtSrc()) || proto.Equal(term, ctxtUsr()) {
								head.Operator = cctpb.MemberAccessExpression_MEMBER_OF_POINTER.Enum()
							}
							head.Lhs = term
							result = head
						} else {
							head.Lhs = &cctpb.Expression{
								Value: &cctpb.Expression_MemberAccessExpression{result},
							}
							result = head
						}
					}

					switch term.Value.(type) {
					case *cctpb.Expression_IdentifierExpression:
						{
							expr := &cctpb.Expression{
								Value: &cctpb.Expression_MemberAccessExpression{result},
							}
							return expr
						}
					case *cctpb.Expression_MemberAccessExpression:
						{
							expr := &cctpb.Expression{
								Value: &cctpb.Expression_MemberAccessExpression{result},
							}
							return expr
						}
					default:
						panic(fmt.Sprintf("unsupported base term type: %v", proto.MarshalTextString(term)))
					}
				}

				return t.walkTerm(e.GetBase().GetTerm())
			}

			panic("cannot walk BaseExpression")
		}

	case e.GetBinaryOp() != nil:
		{
			if e.GetBinaryOp().Op == nil {
				panic("cannot walk binary op with no op type")
			}
			expr := t.BinaryOpToExpr(e.GetBinaryOp().GetOp())
			switch expr.Value.(type) {
			case *cctpb.Expression_ComparisonExpression:
				{
					expr.GetComparisonExpression().Lhs = t.walkExpression(e.GetBinaryOp().GetLhs())
					expr.GetComparisonExpression().Rhs = t.walkExpression(e.GetBinaryOp().GetRhs())
					return expr
				}
			case *cctpb.Expression_ArithmeticExpression:
				{
					lhs := t.walkExpression(e.GetBinaryOp().GetLhs())
					rhs := t.walkExpression(e.GetBinaryOp().GetRhs())

					isWorld := proto.Equal(lhs, genericCtxtCall("world"))
					isView := proto.Equal(lhs, coreProcCall("view"))
					isBitwiseLShift := expr.GetArithmeticExpression().GetOperator() == cctpb.ArithmeticExpression_BITWISE_LSHIFT

					if (isView || isWorld) && isBitwiseLShift {
						// rewrite ctxt.world() << foo --> ctxt.world()->p("Broadcast")
						mae := &cctpb.MemberAccessExpression{
							Operator: cctpb.MemberAccessExpression_MEMBER_OF_POINTER.Enum(),
							Lhs:      lhs,
						}
						p := &cctpb.FunctionCallExpression{
							Name: &cctpb.Expression{
								Value: &cctpb.Expression_IdentifierExpression{
									&cctpb.Identifier{
										Id: proto.String("p"),
									},
								},
							},
						}
						p.Arguments = append(p.Arguments,
							&cctpb.FunctionCallExpression_ExpressionArg{
								Value: &cctpb.FunctionCallExpression_ExpressionArg_Expression{
									&cctpb.Expression{
										Value: &cctpb.Expression_LiteralExpression{
											&cctpb.Literal{
												Value: &cctpb.Literal_StringLiteral{"Broadcast"},
											},
										},
									},
								},
							})

						p.Arguments = append(p.Arguments, &cctpb.FunctionCallExpression_ExpressionArg{
							Value: &cctpb.FunctionCallExpression_ExpressionArg_Expression{rhs},
						})
						mae.Rhs = &cctpb.Expression{
							Value: &cctpb.Expression_FunctionCallExpression{p}}
						return &cctpb.Expression{
							Value: &cctpb.Expression_MemberAccessExpression{mae},
						}
					}

					expr.GetArithmeticExpression().Lhs = lhs
					expr.GetArithmeticExpression().Rhs = rhs
					return expr
				}
			case *cctpb.Expression_LogicalExpression:
				{
					expr.GetLogicalExpression().Lhs = t.walkExpression(e.GetBinaryOp().GetLhs())
					expr.GetLogicalExpression().Rhs = t.walkExpression(e.GetBinaryOp().GetRhs())
					return expr
				}
			case *cctpb.Expression_FunctionCallExpression:
				{
					// in function calls just add lhs and rhs as arguments
					addFuncExprArg(expr.GetFunctionCallExpression(),
						t.walkExpression(e.GetBinaryOp().GetLhs()))
					addFuncExprArg(expr.GetFunctionCallExpression(),
						t.walkExpression(e.GetBinaryOp().GetRhs()))
					return expr
				}
			default:
				panic(fmt.Sprintf("unsupported expression for binary op %v",
					proto.MarshalTextString(expr)))
			}
		}

	case e.GetAssignOp() != nil:
		{
			operator := ConvertAssignOp(e.GetAssignOp().GetOp())
			lhs := t.walkExpression(e.GetAssignOp().GetLhs())
			rhs := t.walkExpression(e.GetAssignOp().GetRhs())
			if isVarAssignFromPtr(lhs) {
				assignExpr := &cctpb.AssignmentExpression{
					Operator: &operator,
					Rhs:      rhs,
				}
				deref := &cctpb.Expression{
					Value: &cctpb.Expression_UnaryExpression{
						&cctpb.UnaryExpression{
							Operator: cctpb.UnaryExpression_POINTER_INDIRECTION.Enum(),
							Operand:  lhs,
						},
					},
				}
				assignExpr.Lhs = deref
				return &cctpb.Expression{
					Value: &cctpb.Expression_AssignmentExpression{assignExpr},
				}
			}
			assignExpr := &cctpb.AssignmentExpression{
				Operator: &operator,
				Lhs:      lhs,
				Rhs:      rhs,
			}
			return &cctpb.Expression{
				Value: &cctpb.Expression_AssignmentExpression{assignExpr},
			}

		}

	default:
		panic(fmt.Sprintf("cannot walk unsupported expr %v", proto.MarshalTextString(e)))
	}
}
