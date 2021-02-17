// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	_ "log"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
	"snowfrost.garden/donk/transpiler/scope"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func genericCtxtCall(n string) *cctpb.Expression {
	mae := &cctpb.MemberAccessExpression{
		Operator: cctpb.MemberAccessExpression_MEMBER_OF_OBJECT.Enum(),
	}

	ctxt := &cctpb.Identifier{
		Id: proto.String("ctxt"),
	}
	lhs := &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{ctxt},
	}

	src := &cctpb.Identifier{
		Id: proto.String(n),
	}
	srcCall := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{src},
		},
	}
	rhs := &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{srcCall},
	}

	mae.Lhs = lhs
	mae.Rhs = rhs

	expr := &cctpb.Expression{
		Value: &cctpb.Expression_MemberAccessExpression{mae},
	}
	return expr
}

func ctxtMakeCall(p paths.Path) *cctpb.Expression {
	mae := &cctpb.MemberAccessExpression{
		Operator: cctpb.MemberAccessExpression_MEMBER_OF_OBJECT.Enum(),
	}

	ctxt := &cctpb.Identifier{
		Id: proto.String("ctxt"),
	}
	lhs := &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{ctxt},
	}

	mk := &cctpb.Identifier{
		Id: proto.String("make"),
	}
	mkCall := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{mk},
		},
	}
	mkCall.Arguments = append(mkCall.Arguments,
		&cctpb.FunctionCallExpression_ExpressionArg{
			Value: &cctpb.FunctionCallExpression_ExpressionArg_Expression{
				&cctpb.Expression{
					Value: &cctpb.Expression_LiteralExpression{
						&cctpb.Literal{
							Value: &cctpb.Literal_StringLiteral{p.FullyQualifiedString()},
						},
					},
				},
			},
		})
	rhs := &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{mkCall},
	}

	mae.Lhs = lhs
	mae.Rhs = rhs

	expr := &cctpb.Expression{
		Value: &cctpb.Expression_MemberAccessExpression{mae},
	}
	return expr
}

func ctxtSrc() *cctpb.Expression {
	return genericCtxtCall("src")
}

func getObjVar(v string) *cctpb.Expression {
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{
				&cctpb.Identifier{Id: proto.String("v")},
			},
		},
	}
	l := &cctpb.Literal{
		Value: &cctpb.Literal_StringLiteral{v},
	}
	addFuncExprArg(fc, &cctpb.Expression{Value: &cctpb.Expression_LiteralExpression{l}})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func getObjFunc(f string) *cctpb.Expression {
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{
				&cctpb.Identifier{Id: proto.String("p")},
			},
		},
	}
	l := &cctpb.Literal{
		Value: &cctpb.Literal_StringLiteral{f},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{l},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func stdStringCtor(s string) *cctpb.Expression {
	i := &cctpb.Identifier{
		Namespace: proto.String("std"),
		Id:        proto.String("string"),
	}
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{&cctpb.Literal{
			Value: &cctpb.Literal_StringLiteral{s}}},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func resourceId(resource string) *cctpb.Expression {
	i := &cctpb.Identifier{
		Namespace: proto.String("donk"),
		Id:        proto.String("resource_t"),
	}
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{&cctpb.Literal{
			Value: &cctpb.Literal_StringLiteral{resource}}},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func pathExpression(p paths.Path) *cctpb.Expression {
	i := &cctpb.Identifier{
		Namespace: proto.String("donk"),
		Id:        proto.String("path_t"),
	}
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{&cctpb.Literal{
			Value: &cctpb.Literal_StringLiteral{p.FullyQualifiedString()}}},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func prefabExpression(p paths.Path) *cctpb.Expression {
	i := &cctpb.Identifier{
		Namespace: proto.String("donk"),
		Id:        proto.String("prefab_t"),
	}
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{&cctpb.Literal{
			Value: &cctpb.Literal_StringLiteral{p.FullyQualifiedString()}}},
	})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func ctxtUsr() *cctpb.Expression {
	return genericCtxtCall("usr")
}

func argParam() *cctpb.Expression {
	return &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{
			&cctpb.Identifier{
				Id: proto.String("args"),
			},
		},
	}
}

func coreProcCall(name string) *cctpb.Expression {
	core := genericCtxtCall("core")
	fc := core.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression()
	addFuncExprArg(fc, &cctpb.Expression{
		Value: &cctpb.Expression_IdentifierExpression{&cctpb.Identifier{
			Id: proto.String(fmt.Sprintf("\"%v\"", name)),
		}},
	})
	return core
}

func (t Transformer) isVarAssignFromPtr(e *cctpb.Expression) bool {
	if e.GetMemberAccessExpression() != nil {
		mae := e.GetMemberAccessExpression()
		if !proto.Equal(mae.GetLhs(), ctxtSrc()) && !proto.Equal(mae.GetLhs(), ctxtUsr()) {
			return false
		}

		if mae.GetRhs().GetFunctionCallExpression().GetName().GetIdentifierExpression().GetId() == "v" {
			return true
		}
	}

	if isRawIdentifier(e) {
		rId := rawIdentifier(e)
		if t.curScope().HasField(rId) && t.curScope().VarType(rId) == scope.VarTypeDMObject {
			return true
		}
	}

	return false
}

func (t Transformer) isFmtRedirected(e *cctpb.Expression) bool {
	if proto.Equal(e, ctxtSrc()) || proto.Equal(e, ctxtUsr()) {
		return true
	}

	lhs := e.GetMemberAccessExpression().GetLhs()
	if lhs != nil {
		if isRawIdentifier(lhs) {
			rId := rawIdentifier(lhs)
			if t.curScope().HasLocal(rId) && t.curScope().VarType(rId) == scope.VarTypeDMObject {
				return true
			}
			if rId == "args" {
				return true
			}
		}
	}

	return false
}

func wrapFuncCallExpr(e *cctpb.Expression) *cctpb.FunctionCallExpression_ExpressionArg {
	return &cctpb.FunctionCallExpression_ExpressionArg{
		Value: &cctpb.FunctionCallExpression_ExpressionArg_Expression{e},
	}
}

func addFuncExprArg(fce *cctpb.FunctionCallExpression, e *cctpb.Expression) {
	fce.Arguments = append(fce.Arguments, wrapFuncCallExpr(e))
}

func addFuncInitListArg(fce *cctpb.FunctionCallExpression, exprs ...*cctpb.Expression) {
	iList := &cctpb.InitializerList{}
	for _, expr := range exprs {
		iList.Args = append(iList.Args, expr)
	}
	fce.Arguments = append(fce.Arguments, &cctpb.FunctionCallExpression_ExpressionArg{
		Value: &cctpb.FunctionCallExpression_ExpressionArg_InitializerList{iList},
	})
}

func wrapDeclarationInStatment(decl *cctpb.Declaration) *cctpb.Statement {
	return &cctpb.Statement{
		Value: &cctpb.Statement_DeclarationStatement{decl},
	}
}

func (t *Transformer) declareInt(name string) *cctpb.Declaration {
	return declareVarWithTypeIdent(name, &cctpb.Identifier{
		Id: proto.String("int"),
	})
}

func declareVarWithTypeIdentAndInitializer(name string, ident *cctpb.Identifier, init *cctpb.Initializer) *cctpb.Declaration {
	ds := &cctpb.DeclarationSpecifier{
		Value: &cctpb.DeclarationSpecifier_TypeSpecifier{
			&cctpb.TypeSpecifier{
				Value: &cctpb.TypeSpecifier_SimpleTypeSpecifier{
					&cctpb.SimpleTypeSpecifier{
						Value: &cctpb.SimpleTypeSpecifier_DeclaredName{ident},
					},
				},
			},
		},
	}

	d := &cctpb.Declarator{
		Value: &cctpb.Declarator_DeclaredName{
			&cctpb.Identifier{
				Id: proto.String(name),
			},
		},
		Initializer: init,
	}

	sd := &cctpb.SimpleDeclaration{}
	sd.Specifiers = append(sd.Specifiers, ds)
	sd.Declarators = append(sd.Declarators, d)

	return &cctpb.Declaration{
		Value: &cctpb.Declaration_BlockDeclaration{
			&cctpb.BlockDeclaration{
				Value: &cctpb.BlockDeclaration_SimpleDeclaration{sd},
			},
		},
	}
}

func declareVarWithTypeIdent(name string, ident *cctpb.Identifier) *cctpb.Declaration {
	ds := &cctpb.DeclarationSpecifier{
		Value: &cctpb.DeclarationSpecifier_TypeSpecifier{
			&cctpb.TypeSpecifier{
				Value: &cctpb.TypeSpecifier_SimpleTypeSpecifier{
					&cctpb.SimpleTypeSpecifier{
						Value: &cctpb.SimpleTypeSpecifier_DeclaredName{ident},
					},
				},
			},
		},
	}

	d := &cctpb.Declarator{
		Value: &cctpb.Declarator_DeclaredName{
			&cctpb.Identifier{
				Id: proto.String(name),
			},
		},
	}

	sd := &cctpb.SimpleDeclaration{}
	sd.Specifiers = append(sd.Specifiers, ds)
	sd.Declarators = append(sd.Declarators, d)

	return &cctpb.Declaration{
		Value: &cctpb.Declaration_BlockDeclaration{
			&cctpb.BlockDeclaration{
				Value: &cctpb.BlockDeclaration_SimpleDeclaration{sd},
			},
		},
	}
}

func declareVar(name string) *cctpb.Declaration {
	return declareVarWithTypeIdent(name, &cctpb.Identifier{
		Namespace: proto.String("donk"),
		Id:        proto.String("var_t"),
	})
}

func declareAuto(name string) *cctpb.Declaration {
	return declareVarWithTypeIdent(name, &cctpb.Identifier{
		Id: proto.String("auto"),
	})
}

func isStringLiteral(e *cctpb.Expression) bool {
	switch e.GetValue().(type) {
	case *cctpb.Expression_LiteralExpression:
		switch e.GetLiteralExpression().GetValue().(type) {
		case *cctpb.Literal_StringLiteral:
			{
				return true
			}
		default:
			{
				return false
			}
		}
	default:
		{
			return false
		}
	}
}

func (t *Transformer) declareVarWithVal(name string, val *cctpb.Expression, varType scope.VarType) *cctpb.Statement {
	var ds *cctpb.DeclarationSpecifier
	// if varType == scope.VarTypeDMObject {
	ds = &cctpb.DeclarationSpecifier{
		Value: &cctpb.DeclarationSpecifier_TypeSpecifier{
			&cctpb.TypeSpecifier{
				Value: &cctpb.TypeSpecifier_SimpleTypeSpecifier{
					&cctpb.SimpleTypeSpecifier{
						Value: &cctpb.SimpleTypeSpecifier_DeclaredName{
							&cctpb.Identifier{
								Id: proto.String("auto"),
							},
						},
					},
				},
			},
		},
	}
	// } else {
	// 	ds = &cctpb.DeclarationSpecifier{
	// 		Value: &cctpb.DeclarationSpecifier_TypeSpecifier{
	// 			&cctpb.TypeSpecifier{
	// 				Value: &cctpb.TypeSpecifier_SimpleTypeSpecifier{
	// 					&cctpb.SimpleTypeSpecifier{
	// 						Value: &cctpb.SimpleTypeSpecifier_DeclaredName{
	// 							&cctpb.Identifier{
	// 								Namespace: proto.String("donk"),
	// 								Id:        proto.String("var_t"),
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	}
	// }
	if isStringLiteral(val) {
		t.curScope().AddDefnHeader("<string>")
		val = stdStringCtor(val.GetLiteralExpression().GetStringLiteral())
	}

	d := &cctpb.Declarator{
		Initializer: &cctpb.Initializer{
			Value: &cctpb.Initializer_CopyInitializer{
				&cctpb.CopyInitializer{
					Other: val,
				},
			},
		},
		Value: &cctpb.Declarator_DeclaredName{
			&cctpb.Identifier{
				Id: proto.String(name),
			},
		},
	}

	sd := &cctpb.SimpleDeclaration{}
	sd.Specifiers = append(sd.Specifiers, ds)
	sd.Declarators = append(sd.Declarators, d)

	return &cctpb.Statement{
		Value: &cctpb.Statement_DeclarationStatement{
			&cctpb.Declaration{
				Value: &cctpb.Declaration_BlockDeclaration{
					&cctpb.BlockDeclaration{
						Value: &cctpb.BlockDeclaration_SimpleDeclaration{sd},
					},
				},
			},
		},
	}
}

func isRawIdentifier(expr *cctpb.Expression) bool {
	switch expr.GetValue().(type) {
	case *cctpb.Expression_IdentifierExpression:
		{
			return expr.GetIdentifierExpression().GetNamespace() == "" &&
				expr.GetIdentifierExpression().GetId() != ""
		}
	default:
		{
			return false
		}
	}
}

func isRawLiteral(expr *cctpb.Expression) bool {
	switch expr.GetValue().(type) {
	case *cctpb.Expression_LiteralExpression:
		{
			return true
		}
	default:
		{
			return false
		}
	}
}

func isRawIdentifierAstpb(expr *astpb.Expression) bool {
	return expr.GetBase().GetTerm().Ident != nil
}

func rawIdentifierAstpb(expr *astpb.Expression) string {
	return expr.GetBase().GetTerm().GetIdent()
}

func rawIdentifier(expr *cctpb.Expression) string {
	if !isRawIdentifier(expr) {
		panic("asked for raw identifier from expression which is not one")
	}
	return expr.GetIdentifierExpression().GetId()
}

func isDeclaringNewDMObject(expr *astpb.Expression) bool {
	return expr.GetBase().GetTerm().GetNew().GetType().GetPrefab() != nil &&
		len(expr.GetBase().GetTerm().GetNew().GetType().GetPrefab().Path) > 0
}

func isRawInt(expr *astpb.Expression) bool {
	if expr.GetBase().GetTerm().IntT != nil {
		return true
	}
	if expr.GetBase().GetTerm().GetExpr() != nil {
		return isRawInt(expr.GetBase().GetTerm().GetExpr())
	}
	return false
}

func rawInt(expr *astpb.Expression) int32 {
	if expr.GetBase().GetTerm().IntT != nil {
		return expr.GetBase().GetTerm().GetIntT()
	}
	if expr.GetBase().GetTerm().GetExpr() != nil {
		return rawInt(expr.GetBase().GetTerm().GetExpr())
	}

	panic(fmt.Sprintf("asked for raw int of unsupported expression %v", proto.MarshalTextString(expr)))
}

func makeIntExpr(i int32) *cctpb.Expression {
	return &cctpb.Expression{
		Value: &cctpb.Expression_LiteralExpression{
			&cctpb.Literal{
				Value: &cctpb.Literal_IntegerLiteral{int64(i)},
			},
		},
	}
}

func makeApiFuncDecl(name string) *cctpb.FunctionDeclaration {
	fd := &cctpb.FunctionDeclaration{
		Name: proto.String(name),
		ReturnType: &cctpb.CppType{
			PType: cctpb.CppType_NONE.Enum(),
			Name:  proto.String("void"),
		},
	}

	fd.Arguments = append(fd.Arguments,
		&cctpb.FunctionArgument{
			Name: proto.String("ctxt"),
			CppType: &cctpb.CppType{
				PType: cctpb.CppType_REFERENCE.Enum(),
				Name:  proto.String("donk::proc_ctxt_t"),
			},
		})

	fd.Arguments = append(fd.Arguments,
		&cctpb.FunctionArgument{
			Name: proto.String("args"),
			CppType: &cctpb.CppType{
				PType: cctpb.CppType_REFERENCE.Enum(),
				Name:  proto.String("donk::proc_args_t"),
			},
		})
	return fd
}
