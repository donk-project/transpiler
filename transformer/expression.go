// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	// "log"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/scope"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

const (
	BroadcastRedirectProcName    = "DONKAPI_Broadcast"
	BroadcastLogRedirectProcName = "DONKAPI_BroadcastLog"
)

func (t Transformer) walkExpression(e *astpb.Expression) *cctpb.Expression {
	if e == nil {
		panic("tried to walk empty expression")
	}
	// log.Printf("=========================== ASTPB EXPRESSION ========================\n%v\n", proto.MarshalTextString(e))

	switch {
	case e.GetBase() != nil:
		{
			if e.GetBase().GetTerm() != nil {
				if len(e.GetBase().GetFollow()) > 0 {
					term := t.walkTerm(e.GetBase().GetTerm())
					isVarAccess := false
					if isRawIdentifier(term) {
						rId := rawIdentifier(term)
						if t.curScope().HasField(rId) {
							isVarAccess = true
						} else if t.curScope().HasLocal(rId) {
							if t.curScope().VarType(rId) == scope.VarTypeDMObject {
								isVarAccess = true
							}
						}
					}
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
									if isVarAccess || proto.Equal(term, ctxtSrc()) || proto.Equal(term, ctxtUsr()) || proto.Equal(term, genericCtxtCall("world")) {
										curFollow.Operator = cctpb.MemberAccessExpression_MEMBER_OF_POINTER.Enum()
										if field.GetS() == "log" {
											curFollow.Rhs = getObjFunc(BroadcastLogRedirectProcName)
										} else {
											curFollow.Rhs = getObjVar(field.GetS())
										}
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
					lhs := t.walkExpression(e.GetBinaryOp().GetLhs())
					rhs := t.walkExpression(e.GetBinaryOp().GetRhs())

					expr.GetComparisonExpression().Lhs = lhs
					expr.GetComparisonExpression().Rhs = rhs
					return expr
				}
			case *cctpb.Expression_ArithmeticExpression:
				{
					lhs := t.walkExpression(e.GetBinaryOp().GetLhs())
					rhs := t.walkExpression(e.GetBinaryOp().GetRhs())

					wrld := genericCtxtCall("world")
					isWorld := proto.Equal(lhs, wrld)
					wrldLog := vsk.PtrMember(wrld, getObjFunc(BroadcastLogRedirectProcName))
					isWorldLog := proto.Equal(lhs, wrldLog)
					isBitwiseLShift := expr.GetArithmeticExpression().GetOperator() == cctpb.ArithmeticExpression_BITWISE_LSHIFT

					// TODO: Also world.log, or any list containing mobs, or any mob, or view/oview
					if (isWorld || isWorldLog) && isBitwiseLShift {
						pc := t.procCall(wrld, BroadcastRedirectProcName, []*cctpb.Expression{rhs}, "ChildProc")
						return &cctpb.Expression{
							Value: &cctpb.Expression_CoYield{
								&cctpb.CoYield{
									Expression: pc,
								},
							},
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
					vsk.AddFuncArg(expr.GetFunctionCallExpression(),
						t.walkExpression(e.GetBinaryOp().GetLhs()))
					vsk.AddFuncArg(expr.GetFunctionCallExpression(),
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
