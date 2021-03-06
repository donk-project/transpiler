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

func (t Transformer) buildCoretypeDecl() *cctpb.ClassDeclaration {
	clsDecl := &cctpb.ClassDeclaration{
		Name: proto.String(t.curScope().CurPath.Basename + "_coretype"),
	}

	if t.curScope().CurPath.ParentPath().IsRoot() {
		t.curScope().AddDeclHeader("\"donk/core/iota.h\"")
		clsDecl.BaseSpecifiers = append(clsDecl.BaseSpecifiers,
			&cctpb.BaseSpecifier{
				AccessSpecifier: cctpb.AccessSpecifier_PUBLIC.Enum(),
				ClassOrDecltype: &cctpb.Identifier{
					Namespace: proto.String(t.coreNamespace),
					Id:        proto.String("iota_t"),
				},
				Virtual: proto.Bool(true),
			},
		)
	} else {
		t.curScope().AddDeclHeader(
			fmt.Sprintf("\"donk/api%v.h\"", t.curScope().CurPath.ParentPath().FullyQualifiedString()))
		nsRoot := t.coreNamespace + "::"
		if t.isCoreGen() {
			nsRoot = nsRoot + "api::"
		}
		clsDecl.BaseSpecifiers = append(clsDecl.BaseSpecifiers,
			&cctpb.BaseSpecifier{
				AccessSpecifier: cctpb.AccessSpecifier_PUBLIC.Enum(),
				ClassOrDecltype: &cctpb.Identifier{
					Namespace: proto.String(nsRoot + t.curScope().CurPath.ParentPath().AsNamespace()),
					Id:        proto.String(t.curScope().CurPath.ParentPath().Basename + "_coretype"),
				},
			},
		)

	}

	clsDecl.MemberSpecifiers = append(clsDecl.MemberSpecifiers, &cctpb.MemberSpecification{
		Value: &cctpb.MemberSpecification_AccessSpecifier{
			*cctpb.AccessSpecifier_PUBLIC.Enum(),
		},
	})

	clsDecl.MemberSpecifiers = append(clsDecl.MemberSpecifiers, &cctpb.MemberSpecification{
		Value: &cctpb.MemberSpecification_Destructor{
			&cctpb.Destructor{
				ClassName: &cctpb.Identifier{
					Id: proto.String(clsDecl.GetName()),
				},
				BlockDefinition: &cctpb.BlockDefinition{},
			},
		},
	})

	clsDecl.MemberSpecifiers = append(clsDecl.MemberSpecifiers, &cctpb.MemberSpecification{
		Value: &cctpb.MemberSpecification_FunctionDeclaration{
			&cctpb.FunctionDeclaration{
				Name: proto.String("InternalCoreRegister"),
				ReturnType: &cctpb.CppType{
					PType: cctpb.CppType_NONE.Enum(),
					Name:  proto.String("void"),
				},
				VirtSpecifier: &cctpb.VirtSpecifier{
					Keyword: cctpb.VirtSpecifier_OVERRIDE.Enum(),
				},
			},
		},
	})

	clsDecl.MemberSpecifiers = append(clsDecl.MemberSpecifiers, &cctpb.MemberSpecification{
		Value: &cctpb.MemberSpecification_Constructor{
			&cctpb.Constructor{
				ClassName: &cctpb.Identifier{
					Id: proto.String(clsDecl.GetName()),
				},
				Arguments: []*cctpb.FunctionArgument{
					&cctpb.FunctionArgument{
						CppType: &cctpb.CppType{
							PType: cctpb.CppType_NONE.Enum(),
							Name:  proto.String("donk::path_t"),
						},
						Name: proto.String("path"),
					},
				},
			},
		},
	})

	clsDecl.MemberSpecifiers = append(clsDecl.MemberSpecifiers, &cctpb.MemberSpecification{
		Value: &cctpb.MemberSpecification_MemberDeclarator{
			&cctpb.MemberDeclarator{
				Value: &cctpb.MemberDeclarator_DeclaredName{
					&cctpb.Identifier{
						Namespace: proto.String("donk"),
						Id:        proto.String("iota_t"),
					},
				},
				Class:  proto.Bool(true),
				Friend: proto.Bool(true),
			},
		},
	})

	return clsDecl
}

func (t *Transformer) makeCoregenRegisterProcFCE(namespace string, apiName string, actualName string) *cctpb.Expression {
	fce := &cctpb.FunctionCallExpression{Name: vsk.IdExpr(vsk.Id("RegisterProc"))}

	// Using proc.Name here and proc.EmitName() below maps e.g.
	// the proc "icon" to "icon_", so we don't have to try and
	// guess what proc names should be mangled in calling code
	vsk.AddFuncArg(fce, vsk.StringLiteralExpr(apiName))
	vsk.AddFuncArg(fce, vsk.IdExpr(vsk.NsId(namespace, actualName)))

	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fce},
	}
}

func (t *Transformer) generateInternalCoreRegister(namespace string) *cctpb.FunctionDefinition {
	fnDefn := &cctpb.FunctionDefinition{}
	blockDefn := &cctpb.BlockDefinition{}
	var stmts []*cctpb.Statement

	procs := t.curScope().CurType.Procs
	vars := t.curScope().CurType.Vars

	for _, proc := range procs {
		if t.shouldEmitProc(proc) {
			stmts = append(stmts, &cctpb.Statement{
				Value: &cctpb.Statement_ExpressionStatement{
					t.makeCoregenRegisterProcFCE(namespace, proc.Name, proc.EmitName()),
				},
			})
		}
	}

	if t.isCoreGen() && t.curScope().CurPath.Equals("/world") {
		stmts = append(stmts, &cctpb.Statement{
			Value: &cctpb.Statement_ExpressionStatement{
				t.makeCoregenRegisterProcFCE(namespace, BroadcastRedirectProcName, BroadcastRedirectProcName),
			},
		})
		stmts = append(stmts, &cctpb.Statement{
			Value: &cctpb.Statement_ExpressionStatement{
				t.makeCoregenRegisterProcFCE(namespace, BroadcastLogRedirectProcName, BroadcastLogRedirectProcName),
			},
		})
	}

	for _, v := range vars {
		if t.shouldEmitVar(v) {
			fce := &cctpb.FunctionCallExpression{
				Name: &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{
						&cctpb.Identifier{
							Id: proto.String("RegisterVar"),
						},
					},
				},
			}

			vsk.AddFuncArg(fce, &cctpb.Expression{
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

func (t *Transformer) generateCoreConstructor() *cctpb.Constructor {
	className := t.curScope().CurPath.Basename + "_coretype"
	c := &cctpb.Constructor{
		ClassName: &cctpb.Identifier{
			Id: proto.String(fmt.Sprintf("%v::%v", className, className)),
		},
		BlockDefinition: &cctpb.BlockDefinition{},
	}

	c.Arguments = append(c.Arguments, &cctpb.FunctionArgument{
		CppType: &cctpb.CppType{
			PType: cctpb.CppType_NONE.Enum(),
			Name:  proto.String("donk::path_t"),
		},
		Name: proto.String("path"),
	})

	fce := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{
				&cctpb.Identifier{
					Id: proto.String("InternalCoreRegister"),
				},
			},
		},
	}

	c.BlockDefinition.Statements = append(c.BlockDefinition.Statements,
		&cctpb.Statement{
			Value: &cctpb.Statement_ExpressionStatement{
				&cctpb.Expression{
					Value: &cctpb.Expression_FunctionCallExpression{fce},
				},
			},
		},
	)

	mi := &cctpb.MemberInitializer{
		Member: &cctpb.Identifier{
			Namespace: proto.String("donk"),
			Id:        proto.String("iota_t"),
		},
	}

	mi.Expressions = append(mi.Expressions, &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{
			&cctpb.Identifier{
				Id: proto.String("path"),
			},
		},
	})
	c.MemberInitializers = append(c.MemberInitializers, mi)

	if !t.curScope().CurPath.ParentPath().IsRoot() {
		t.curScope().AddDefnHeader(
			fmt.Sprintf("\"donk/api%v.h\"", t.curScope().CurPath.ParentPath().FullyQualifiedString()))
		coreNs := t.coreNamespace
		if t.isCoreGen() {
			coreNs = coreNs + "::api"
		}
		mi := &cctpb.MemberInitializer{
			Member: &cctpb.Identifier{
				Namespace: proto.String(coreNs + "::" + t.curScope().CurPath.ParentPath().AsNamespace()),
				Id:        proto.String(t.curScope().CurPath.ParentPath().Basename + "_coretype"),
			},
		}

		mi.Expressions = append(mi.Expressions, &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{
				&cctpb.Identifier{
					Id: proto.String("path"),
				},
			},
		})
		c.MemberInitializers = append(c.MemberInitializers, mi)

	}

	t.curScope().AddDefnHeader("\"donk/core/path.h\"")
	t.curScope().AddDefnHeader("\"donk/core/iota.h\"")
	return c
}
