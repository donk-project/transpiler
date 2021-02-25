// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

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
