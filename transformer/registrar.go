// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"

	"github.com/golang/protobuf/proto"
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
			mae := &cctpb.MemberAccessExpression{
				Operator: cctpb.MemberAccessExpression_MEMBER_OF_OBJECT.Enum(),
				Lhs: &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{
						&cctpb.Identifier{
							Id: proto.String("iota"),
						},
					},
				},
				Rhs: &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{
						&cctpb.Identifier{
							Id: proto.String("RegisterProc"),
						},
					},
				},
			}

			fce := &cctpb.FunctionCallExpression{
				Name: &cctpb.Expression{
					Value: &cctpb.Expression_MemberAccessExpression{mae},
				},
			}

			addFuncExprArg(fce, &cctpb.Expression{
				Value: &cctpb.Expression_LiteralExpression{
					&cctpb.Literal{
						// Using proc.Name here and proc.EmitName() below maps e.g.
						// the proc "icon" to "icon_", so we don't have to try and
						// guess what proc names should be mangled in calling code
						Value: &cctpb.Literal_StringLiteral{proc.Name},
					},
				},
			})

			addFuncExprArg(fce, &cctpb.Expression{
				Value: &cctpb.Expression_IdentifierExpression{
					&cctpb.Identifier{
						Namespace: proto.String(namespace),
						Id:        proto.String(proc.EmitName()),
					},
				},
			})

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
			mae := &cctpb.MemberAccessExpression{
				Operator: cctpb.MemberAccessExpression_MEMBER_OF_OBJECT.Enum(),
				Lhs: &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{
						&cctpb.Identifier{
							Id: proto.String("iota"),
						},
					},
				},
				Rhs: &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{
						&cctpb.Identifier{
							Id: proto.String("RegisterVar"),
						},
					},
				},
			}

			fce := &cctpb.FunctionCallExpression{
				Name: &cctpb.Expression{
					Value: &cctpb.Expression_MemberAccessExpression{mae},
				},
			}

			addFuncExprArg(fce, &cctpb.Expression{
				Value: &cctpb.Expression_LiteralExpression{
					&cctpb.Literal{
						Value: &cctpb.Literal_StringLiteral{v.VarName()},
					},
				},
			})

			if v.HasStaticValue() {
				simplDecl := &cctpb.SimpleDeclaration{}

				typ := &cctpb.Identifier{
					Namespace: proto.String("donk"),
					Id:        proto.String("var_t"),
				}
				id := &cctpb.Identifier{
					Id: proto.String(fmt.Sprintf("donk_value_var__%v", v.VarName())),
				}
				cInit := &cctpb.CopyInitializer{}
				if v.Proto.GetValue().GetExpression() != nil {
					term := v.Proto.GetValue().GetExpression().GetBase().GetTerm()
					if term.StringT != nil {
						t.curScope().AddDefnHeader("<string>")
						cInit.Other = stdStringCtor(term.GetStringT())
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

				addFuncExprArg(fce, &cctpb.Expression{
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
