// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	// "log"

	"github.com/golang/protobuf/proto"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/scope"

	astpb "snowfrost.garden/donk/proto/ast"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

var knownSynchronousProcs = map[string]bool{
	"/rand": true,
	// TODO: Keeping track of object types is going to have to be much more sophisticated.
	// It's easy to tell that a proc call is a global, or if the type is a known token
	// (world, usr). But random objects in scope will have to have their types deduced
	// before we can determine if the proc being called is known synchronous. i.e.
	// There is no way at current to put "/mob/procname" here and have it recognized
	// without more object type deduction.
}

func (t Transformer) walkStatement(s *astpb.Statement) *cctpb.Statement {
	stmt := &cctpb.Statement{}
	// log.Printf("=========================== ASTPB STATEMENT ========================\n%v\n", proto.MarshalTextString(s))
	if s == nil {
		return stmt
	}

	if t.isSleep(s) {
		return t.astpbSleepToCctpbSleep(s)
	}

	switch {
	case s.GetExpr() != nil:
		{
			stmt.Value = &cctpb.Statement_ExpressionStatement{t.walkExpression(s.GetExpr())}
			return stmt
		}
	case s.GetDel() != nil:
		{
			mae := genericCtxtCall("del")

			vsk.AddFuncArg(
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
				vr.Type = scope.VarTypeDMObject

				if s.GetVar().GetValue().GetBase().GetTerm().GetCall() != nil &&
					s.GetVar().GetValue().GetBase().GetTerm().GetCall().S != nil {
					fn := s.GetVar().GetValue().GetBase().GetTerm().GetCall().GetS()
					if t.IsProcInCore(fn) {
						var exprs []*cctpb.Expression
						for _, ex := range s.GetVar().GetValue().GetBase().GetTerm().GetCall().GetExpr() {
							exprs = append(exprs, t.walkExpression(ex))
						}
						pc := genericCtxtCall("Gproc")
						vsk.AddFuncArg(pc.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), vsk.StringLiteralExpr(fn))
						vsk.AddFuncInitListArg(pc.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), exprs...)
						t.curScope().AddScopedVar(vr)
						return t.declareVarWithVal(s.GetVar().GetName(), pc, vr.Type)
					} else {
						vr.Type = scope.VarTypeCalledProc
						var exprs []*cctpb.Expression
						for _, ex := range s.GetVar().GetValue().GetBase().GetTerm().GetCall().GetExpr() {
							exprs = append(exprs, t.walkExpression(ex))
						}
						pc := genericCtxtCall("ChildProc")
						vsk.AddFuncArg(pc.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), genericCtxtCall("Global"))
						vsk.AddFuncArg(pc.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), vsk.StringLiteralExpr(fn))
						vsk.AddFuncInitListArg(pc.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), exprs...)
						t.curScope().AddScopedVar(vr)
						return t.declareVarWithVal(s.GetVar().GetName(), &cctpb.Expression{
							Value: &cctpb.Expression_CoYield{
								&cctpb.CoYield{
									Expression: pc,
								},
							},
						}, vr.Type)
					}
				}
				t.curScope().AddScopedVar(vr)
				return t.declareVarWithVal(
					s.GetVar().GetName(),
					t.walkExpression(s.GetVar().GetValue()),
					vr.Type,
				)
			}

			// TODO: Pretty sure this is wrong, an auto decl with no value?
			return wrapDeclarationInStatment(vsk.DeclareAuto(s.GetVar().GetName()))
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

	case s.GetWhileS() != nil:
		{
			if s.GetWhileS().GetCondition() == nil {
				panic(fmt.Sprintf("no expected condition in while: %v", proto.MarshalTextString(s.GetDoWhile())))
			}
			wExpr := t.walkExpression(s.GetWhileS().GetCondition())
			var wStmts []*cctpb.Statement
			for _, wStmt := range s.GetWhileS().GetBlock().GetStatement() {
				wStmts = append(wStmts, t.walkStatement(wStmt))
			}
			stmt.Value = &cctpb.Statement_While{
				&cctpb.While{
					Condition: wExpr,
					BlockDefinition: &cctpb.BlockDefinition{
						Statements: wStmts,
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
				start := vsk.IntLiteralExpr(rawInt(s.GetForRange().GetStart()))
				end := vsk.IntLiteralExpr(rawInt(s.GetForRange().GetEnd()))
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
			rbf.Declaration = vsk.VarDecl(
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
					mae := &cctpb.Expression{
						Value: &cctpb.Expression_MemberAccessExpression{
							&cctpb.MemberAccessExpression{
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
							},
						},
					}
					rbf.RangeExpression = vsk.PtrIndirect(mae)
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
			t.curScope().ReturnFound = true
			fce := genericCtxtCall("SetResult")
			if s.GetReturnS().GetExpr() != nil {
				vsk.AddFuncArg(fce.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), t.walkExpression(s.GetReturnS().GetExpr()))
			}
			stmt.Value = &cctpb.Statement_ExpressionStatement{fce}
			return stmt
		}

	case s.GetIfS() != nil:
		{
			var allIfs []*cctpb.IfStatement
			for _, arm := range s.GetIfS().GetArm() {
				ifs := &cctpb.IfStatement{}
				ifs.Condition = t.walkExpression(arm.GetExpr())
				var trueStmts []*cctpb.Statement

				for _, tS := range arm.GetBlock().GetStatement() {
					trueStmts = append(trueStmts, t.walkStatement(tS))
				}

				ifs.StatementTrue = &cctpb.Statement{
					Value: &cctpb.Statement_CompoundStatement{
						&cctpb.CompoundStatement{
							Statements: trueStmts,
						},
					},
				}

				allIfs = append([]*cctpb.IfStatement{ifs}, allIfs...)
			}

			// This ugly garbage below is needed to turn DM ifs, which are represented
			// in AST as a repeated list of arms followed by one condition-less else,
			// into C++ grammar, in which ifs are pairs of statements, with the latter
			// capable of being an additional if statement.
			//
			// We add each else arm in reverse order to a slice and then respectively
			// assign each one to the false-arm of the one after it.
			var lastElse []*cctpb.Statement
			for _, fS := range s.GetIfS().GetElseArm().GetStatement() {
				lastElse = append(lastElse, t.walkStatement(fS))
			}

			if len(allIfs) == 1 {
				ifs := &cctpb.IfStatement{}
				if len(lastElse) > 0 {
					ifs.StatementFalse = &cctpb.Statement{
						Value: &cctpb.Statement_CompoundStatement{
							&cctpb.CompoundStatement{
								Statements: lastElse,
							},
						},
					}
				}
				stmt.Value = &cctpb.Statement_IfStatement{ifs}
				return stmt
			} else if len(allIfs) > 1 {
				var prior *cctpb.IfStatement
				last, allIfs := allIfs[0], allIfs[1:]
				for len(allIfs) > 0 {
					prior, allIfs = allIfs[0], allIfs[1:]
					prior.StatementFalse = &cctpb.Statement{
						Value: &cctpb.Statement_IfStatement{last},
					}
					last = prior
				}
				stmt.Value = &cctpb.Statement_IfStatement{last}
				return stmt
			}
		}
	case s.GetSetting() != nil:
		{
			if s.GetSetting().GetName() == "src" {
				if s.GetSetting().GetMode() == astpb.SettingMode_SETTING_MODE_IN {
					if s.GetSetting().GetValue().GetBase().GetTerm().GetCall().GetS() == "oview" {
						r := rawInt(s.GetSetting().GetValue().GetBase().GetTerm().GetCall().GetExpr()[0])
						t.curScope().CurProc.Flags.Range = int(r)
						t.curScope().CurProc.Flags.Access = parser.ProcAccessInOView
						return stmt
					}
				}
			} else if s.GetSetting().GetName() == "name" {
				if s.GetSetting().GetMode() == astpb.SettingMode_SETTING_MODE_ASSIGN {
					name := rawString(s.GetSetting().GetValue())
					t.curScope().CurProc.Flags.Name = name
					return stmt
				}
			}
			panic(fmt.Sprintf("cannot transform unsupported setting: %v", proto.MarshalTextString(s)))
		}
	case s.GetSpawn() != nil:
		{
			t.curScope().AddDeclHeader("\"donk/core/procs.h\"")

			spawnCall := vsk.ObjMember(vsk.StringIdExpr("ctxt"), vsk.FuncCall(vsk.Id("Spawn")))
			lambda := &cctpb.LambdaExpression{}

			vsk.AddFuncArg(spawnCall.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(), &cctpb.Expression{
				Value: &cctpb.Expression_LambdaExpression{lambda},
			})

			vsk.AddFuncArg(
				spawnCall.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression(),
				vsk.StringIdExpr("args"),
			)

			lambda.TrailingReturnType = &cctpb.CppType{
				PType: cctpb.CppType_NONE.Enum(),
				Name:  proto.String("donk::running_proc"),
			}

			lambda.Arguments = append(lambda.Arguments,
				&cctpb.FunctionArgument{
					Name: proto.String("ctxt"),
					CppType: &cctpb.CppType{
						PType: cctpb.CppType_REFERENCE.Enum(),
						Name:  proto.String("donk::proc_ctxt_t"),
					},
				})

			lambda.Arguments = append(lambda.Arguments,
				&cctpb.FunctionArgument{
					Name: proto.String("args"),
					CppType: &cctpb.CppType{
						PType: cctpb.CppType_REFERENCE.Enum(),
						Name:  proto.String("donk::proc_args_t"),
					},
				})

			lambda.Body = &cctpb.BlockDefinition{}
			t.PushScope()
			t.curScope().InSpawn = true
			for _, stmt := range s.GetSpawn().GetBlock().GetStatement() {
				lambda.Body.Statements = append(lambda.Body.Statements, t.walkStatement(stmt))
			}
			t.PopScope()

			return &cctpb.Statement{
				Value: &cctpb.Statement_ExpressionStatement{spawnCall},
			}
		}

	}
	panic(fmt.Sprintf("cannot walk unsupported statement %v", proto.MarshalTextString(s)))
}
