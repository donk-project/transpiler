// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/scope"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) walkStatement(s *astpb.Statement) *cctpb.Statement {
	stmt := &cctpb.Statement{}
	// log.Printf("\n============ASTPB Statement:\n%v\n", proto.MarshalTextString(s))
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
				vr := scope.VarInScope{
					Name:  s.GetVar().GetName(),
					Scope: scope.VarScopeLocal,
				}
				if isRawInt(s.GetVar().GetValue()) {
					vr.Type = scope.VarTypeInt
				}
				t.curScope().AddScopedVar(vr)
				return t.declareVarWithVal(
					s.GetVar().GetName(),
					t.walkExpression(s.GetVar().GetValue()),
					vr.Type,
				)
			}

			return wrapDeclarationInStatment(declareAuto(s.GetVar().GetName()))
		}
	case s.GetDoWhile() != nil:
		{
			if s.GetDoWhile().GetCondition() == nil {
				panic(fmt.Sprintf("no expected condition in do-while: %v", proto.MarshalTextString(s.GetDoWhile())))
			}
			dwExpr := t.walkExpression(s.GetDoWhile().GetCondition())
			var dwStmts []*cctpb.Statement
			for _, dwStmt := range s.GetDoWhile().GetBlock().GetStatement() {
				dwStmts = append(dwStmts, t.walkStatement(dwStmt))
			}
			stmt.Value = &cctpb.Statement_DoWhile{
				&cctpb.DoWhile{
					Condition: dwExpr,
					BlockDefinition: &cctpb.BlockDefinition{
						Statements: dwStmts,
					},
				},
			}
			return stmt
		}
	case s.GetForRange() != nil:
		{
			if isRawInt(s.GetForRange().GetStart()) && isRawInt(s.GetForRange().GetEnd()) {
				t.PushScope()

				rbf := &cctpb.RangeBasedFor{}
				decl := t.declareInt(s.GetForRange().GetName())
				start := makeIntExpr(rawInt(s.GetForRange().GetStart()))
				end := makeIntExpr(rawInt(s.GetForRange().GetEnd()))
				rbf.Declaration = decl
				t.curScope().AddDefnHeader("\"donk/core/utils.h\"")
				intRangeF := &cctpb.Identifier{
					Namespace: proto.String("donk::util"),
					Id:        proto.String("int_range"),
				}
				intRangeFc := &cctpb.FunctionCallExpression{
					Name: &cctpb.Expression{
						Value: &cctpb.Expression_IdentifierExpression{intRangeF},
					},
				}
				intRangeFc.Arguments = append(intRangeFc.Arguments, &cctpb.FunctionCallExpression_ExpressionArg{
					Value: &cctpb.FunctionCallExpression_ExpressionArg_Expression{start},
				})
				intRangeFc.Arguments = append(intRangeFc.Arguments, &cctpb.FunctionCallExpression_ExpressionArg{
					Value: &cctpb.FunctionCallExpression_ExpressionArg_Expression{end},
				})
				rbf.RangeExpression = &cctpb.Expression{
					Value: &cctpb.Expression_FunctionCallExpression{intRangeFc},
				}

				vr := scope.VarInScope{
					Name: s.GetForRange().GetName(),
					Type: scope.VarTypeInt,
				}
				t.curScope().AddScopedVar(vr)

				var rbfStmts []*cctpb.Statement
				for _, rbfStmt := range s.GetForRange().GetBlock().GetStatement() {
					rbfStmts = append(rbfStmts, t.walkStatement(rbfStmt))
				}
				rbf.LoopDefinition = &cctpb.BlockDefinition{
					Statements: rbfStmts,
				}

				stmt.Value = &cctpb.Statement_RangeBasedFor{rbf}
				t.PopScope()

				return stmt
			}

			panic(fmt.Sprintf("given for-range syntax not supported yet: %v", proto.MarshalTextString(s.GetForRange())))
		}
	case s.GetForList() != nil:
		{
			rbf := &cctpb.RangeBasedFor{}
			rbf.Declaration = declareVarWithTypeIdent(
				s.GetForList().GetName(),
				&cctpb.Identifier{
					Id: proto.String("auto"),
				},
			)

			if s.GetForList().GetInList() != nil {
				expr := t.walkExpression(s.GetForList().GetInList())

				isVarIterator := false
				if isRawIdentifierAstpb(s.GetForList().GetInList()) {
					rId := rawIdentifierAstpb(s.GetForList().GetInList())
					if t.declaredInPathOrParents(rId) || t.curScope().HasGlobal(rId) {
						isVarIterator = true
					}
				}
				if isVarIterator {
					// Most likely a var_t pointer
					mae := &cctpb.MemberAccessExpression{
						Operator: cctpb.MemberAccessExpression_MEMBER_OF_POINTER.Enum(),
						Lhs:      expr,
						Rhs: &cctpb.Expression{
							Value: &cctpb.Expression_FunctionCallExpression{
								&cctpb.FunctionCallExpression{
									Name: &cctpb.Expression{
										Value: &cctpb.Expression_IdentifierExpression{
											&cctpb.Identifier{
												Id: proto.String("get_list"),
											},
										},
									},
								},
							},
						},
					}
					rbf.RangeExpression = &cctpb.Expression{
						Value: &cctpb.Expression_UnaryExpression{
							&cctpb.UnaryExpression{
								Operator: cctpb.UnaryExpression_POINTER_INDIRECTION.Enum(),
								Operand: &cctpb.Expression{
									Value: &cctpb.Expression_MemberAccessExpression{mae},
								},
							},
						},
					}

				} else {
					rbf.RangeExpression = expr
				}
			}

			t.PushScope()
			vr := scope.VarInScope{
				Name: s.GetForList().GetName(),
				Type: scope.VarTypeListIterator,
			}

			t.curScope().AddScopedVar(vr)

			var rbfStmts []*cctpb.Statement
			for _, rbfStmt := range s.GetForList().GetBlock().GetStatement() {
				rbfStmts = append(rbfStmts, t.walkStatement(rbfStmt))
			}
			rbf.LoopDefinition = &cctpb.BlockDefinition{
				Statements: rbfStmts,
			}

			t.PopScope()
			stmt.Value = &cctpb.Statement_RangeBasedFor{rbf}

			return stmt
		}
	case s.GetReturnS() != nil:
		{
			fce := genericCtxtCall("Result")
			if s.GetReturnS().GetExpr() != nil {
				addFuncExprArg(fce.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), t.walkExpression(s.GetReturnS().GetExpr()))
			}
			stmt.Value = &cctpb.Statement_ExpressionStatement{fce}
			return stmt
		}

	case s.GetIfS() != nil:
		{
			if len(s.GetIfS().GetArm()) == 1 {
				ifs := &cctpb.IfStatement{}
				arm := s.GetIfS().GetArm()[0]
				ifs.Condition = t.walkExpression(arm.GetExpr())
				var trueStmts []*cctpb.Statement
				var falseStmts []*cctpb.Statement

				for _, tS := range arm.GetBlock().GetStatement() {
					trueStmts = append(trueStmts, t.walkStatement(tS))
				}

				for _, fS := range s.GetIfS().GetElseArm().GetStatement() {
					falseStmts = append(falseStmts, t.walkStatement(fS))
				}

				ifs.StatementTrue = &cctpb.Statement{
					Value: &cctpb.Statement_CompoundStatement{
						&cctpb.CompoundStatement{
							Statements: trueStmts,
						},
					},
				}

				if len(falseStmts) > 0 {
					ifs.StatementFalse = &cctpb.Statement{
						Value: &cctpb.Statement_CompoundStatement{
							&cctpb.CompoundStatement{
								Statements: falseStmts,
							},
						},
					}

				}

				stmt.Value = &cctpb.Statement_IfStatement{ifs}
				return stmt
			}
			panic(fmt.Sprintf("multi-armed if not supported yet: %v", proto.MarshalTextString(s.GetIfS())))

		}

	}
	panic(fmt.Sprintf("cannot walk unsupported statement %v", proto.MarshalTextString(s)))
}
