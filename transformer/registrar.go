// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	// "log"

	"github.com/golang/protobuf/proto"
	"snowfrost.garden/donk/transpiler/parser"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func ProcAccessToEnum(i parser.ProcAccess) *cctpb.Expression {
	switch i {
	case parser.ProcAccessUnknown:
		{
			panic("unknown procaccess should not be converted")
		}
	case parser.ProcAccessInView:
		{
			return vsk.IdExpr(vsk.NsId("donk", "proc_access::kInView"))
		}
	case parser.ProcAccessInOView:
		{
			return vsk.IdExpr(vsk.NsId("donk", "proc_access::kInOView"))
		}
	case parser.ProcAccessInUsrLoc:
		{
			return vsk.IdExpr(vsk.NsId("donk", "proc_access::kInUsrLoc"))
		}
	case parser.ProcAccessInUsr:
		{
			return vsk.IdExpr(vsk.NsId("donk", "proc_access::kInUsr"))
		}
	case parser.ProcAccessInWorld:
		{
			return vsk.IdExpr(vsk.NsId("donk", "proc_access::kInWorld"))
		}
	case parser.ProcAccessEqUsr:
		{
			return vsk.IdExpr(vsk.NsId("donk", "proc_access::kEqUsr"))
		}
	case parser.ProcAccessInGroup:
		{
			return vsk.IdExpr(vsk.NsId("donk", "proc_access::kInGroup"))
		}
	}
	panic("unknown proc access")
}

func (t *Transformer) generateRegistrationFunction(namespace string) *cctpb.FunctionDefinition {
	fnDefn := &cctpb.FunctionDefinition{}
	blockDefn := &cctpb.BlockDefinition{}
	var stmts []*cctpb.Statement

	procs := t.curScope().CurType.Procs
	vars := t.curScope().CurType.Vars

	for _, proc := range procs {
		if t.shouldEmitProc(proc) {
			mae := vsk.ObjMember(
				vsk.StringIdExpr("iota"),
				vsk.StringIdExpr("RegisterProc"))

			fce := &cctpb.FunctionCallExpression{Name: mae}

			// Using proc.Name here and proc.EmitName() below maps e.g.
			// the proc "icon" to "icon_", so we don't have to try and
			// guess what proc names should be mangled in calling code
			vsk.AddFuncArg(fce, vsk.StringLiteralExpr(proc.Name))
			vsk.AddFuncArg(fce, vsk.IdExpr(vsk.NsId(namespace, proc.EmitName())))

			stmts = append(stmts, &cctpb.Statement{
				Value: &cctpb.Statement_ExpressionStatement{
					&cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{fce},
					},
				},
			})

			if t.parser.DefaultProcFlags != proc.Flags {
				aps := vsk.ObjMember(
					vsk.StringIdExpr("iota"),
					vsk.StringIdExpr("ProcSettings"))
				apsC := &cctpb.Expression{
					Value: &cctpb.Expression_FunctionCallExpression{
						&cctpb.FunctionCallExpression{Name: aps},
					},
				}
				vsk.AddFuncArg(apsC.GetFunctionCallExpression(), vsk.StringLiteralExpr(proc.Name))

				pst := vsk.IdExpr(vsk.NsId("donk", "proc_settings_t"))
				pstFC := &cctpb.Expression{
					Value: &cctpb.Expression_FunctionCallExpression{
						&cctpb.FunctionCallExpression{Name: pst},
					},
				}

				var nFC *cctpb.FunctionCallExpression
				if proc.Flags.Name != t.parser.DefaultProcFlags.Name {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("name")}
					vsk.AddFuncArg(nFC, vsk.StringLiteralExpr(proc.Flags.Name))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Desc != t.parser.DefaultProcFlags.Desc {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("desc")}
					vsk.AddFuncArg(nFC, vsk.StringLiteralExpr(proc.Flags.Desc))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Category != t.parser.DefaultProcFlags.Category {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("category")}
					vsk.AddFuncArg(nFC, vsk.StringLiteralExpr(proc.Flags.Category))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Hidden != t.parser.DefaultProcFlags.Hidden {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("hidden")}
					vsk.AddFuncArg(nFC, vsk.Bool(proc.Flags.Hidden))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.PopupMenu != t.parser.DefaultProcFlags.PopupMenu {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("popup_menu")}
					vsk.AddFuncArg(nFC, vsk.Bool(proc.Flags.PopupMenu))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Instant != t.parser.DefaultProcFlags.Instant {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("instant")}
					vsk.AddFuncArg(nFC, vsk.Bool(proc.Flags.Instant))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Invisibility != t.parser.DefaultProcFlags.Invisibility {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("invisibility")}
					vsk.AddFuncArg(nFC, vsk.IntLiteralExpr(int32(proc.Flags.Invisibility)))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Access != t.parser.DefaultProcFlags.Access {
					if proc.Flags.Access == parser.ProcAccessUnknown {
						panic("proc access explicitly set to unknown")
					}

					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("access")}
					vsk.AddFuncArg(nFC, ProcAccessToEnum(proc.Flags.Access))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Range != t.parser.DefaultProcFlags.Range {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("range")}
					vsk.AddFuncArg(nFC, vsk.IntLiteralExpr(int32(proc.Flags.Range)))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.Background != t.parser.DefaultProcFlags.Background {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("background")}
					vsk.AddFuncArg(nFC, vsk.Bool(proc.Flags.Background))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}
				if proc.Flags.WaitFor != t.parser.DefaultProcFlags.WaitFor {
					nFC = &cctpb.FunctionCallExpression{Name: vsk.StringIdExpr("waitfor")}
					vsk.AddFuncArg(nFC, vsk.Bool(proc.Flags.WaitFor))
					pstFC = vsk.ObjMember(pstFC, &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{nFC},
					})
				}

				vsk.AddFuncArg(apsC.GetFunctionCallExpression(), pstFC)
				stmts = append(stmts, &cctpb.Statement{
					Value: &cctpb.Statement_ExpressionStatement{apsC},
				})

			}

			if !proc.Args.Empty() {
				for _, name := range proc.Args.Names() {
					aps := vsk.ObjMember(
						vsk.StringIdExpr("iota"),
						vsk.StringIdExpr("ProcInput"))
					apsC := &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{
							&cctpb.FunctionCallExpression{Name: aps},
						},
					}
					vsk.AddFuncArg(apsC.GetFunctionCallExpression(), vsk.StringLiteralExpr(proc.Name))

					pst := vsk.IdExpr(vsk.NsId("donk", "proc_input_t"))
					fce := &cctpb.FunctionCallExpression{Name:pst}
					vsk.AddFuncArg(fce, vsk.StringLiteralExpr(name))
					pstFC := &cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{fce},
					}
					vsk.AddFuncArg(apsC.GetFunctionCallExpression(), pstFC)
					stmts = append(stmts, &cctpb.Statement{
						Value: &cctpb.Statement_ExpressionStatement{apsC},
					})
				}

			}

		}
	}

	for _, v := range vars {
		if t.shouldEmitVar(v) {
			mae := vsk.ObjMember(
				vsk.StringIdExpr("iota"),
				vsk.StringIdExpr("RegisterVar"))

			fce := &cctpb.FunctionCallExpression{Name: mae}

			vsk.AddFuncArg(fce, vsk.StringLiteralExpr(v.VarName()))

			if v.HasStaticValue() {
				simplDecl := &cctpb.SimpleDeclaration{}

				typ := vsk.NsId("donk", "var_t")
				id := vsk.Id(fmt.Sprintf("donk_value_var__%v", v.VarName()))

				cInit := &cctpb.CopyInitializer{}
				if v.Proto.GetValue().GetExpression() != nil {
					term := v.Proto.GetValue().GetExpression().GetBase().GetTerm()
					if term.StringT != nil {
						t.curScope().AddDefnHeader("<string>")
						cInit.Other = vsk.StdStringCtor(term.GetStringT())
					} else {
						cInit.Other = t.walkExpression(v.Proto.GetValue().GetExpression())
					}
				} else if v.Proto.GetValue().GetConstant() != nil {
					cInit.Other = t.walkConstant(v.Proto.GetValue().GetConstant())
				} else {
					panic(fmt.Sprintf("cannot print unknown static value %v", proto.MarshalTextString(v.Proto)))
				}
				init := &cctpb.Initializer{
					Value: &cctpb.Initializer_CopyInitializer{cInit},
				}
				decl := &cctpb.Declarator{
					Initializer: init,
					Value:       &cctpb.Declarator_DeclaredName{id},
				}
				simplDecl.Declarators = append(simplDecl.Declarators, decl)

				sTSpec := &cctpb.SimpleTypeSpecifier{
					Value: &cctpb.SimpleTypeSpecifier_DeclaredName{typ},
				}
				tSpec := &cctpb.TypeSpecifier{
					Value: &cctpb.TypeSpecifier_SimpleTypeSpecifier{sTSpec},
				}
				declSpec := &cctpb.DeclarationSpecifier{
					Value: &cctpb.DeclarationSpecifier_TypeSpecifier{tSpec},
				}
				simplDecl.Specifiers = append(simplDecl.Specifiers, declSpec)

				blockDecl := &cctpb.BlockDeclaration{
					Value: &cctpb.BlockDeclaration_SimpleDeclaration{simplDecl},
				}
				declStmt := &cctpb.Declaration{
					Value: &cctpb.Declaration_BlockDeclaration{blockDecl},
				}

				stmt := &cctpb.Statement{
					Value: &cctpb.Statement_DeclarationStatement{declStmt},
				}

				stmts = append(stmts, stmt)

				vsk.AddFuncArg(fce, &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{id},
				})
			}

			stmts = append(stmts, &cctpb.Statement{
				Value: &cctpb.Statement_ExpressionStatement{
					&cctpb.Expression{
						Value: &cctpb.Expression_FunctionCallExpression{fce},
					},
				},
			})
		}

	}

	blockDefn.Statements = stmts
	fnDefn.BlockDefinition = blockDefn

	return fnDefn
}
