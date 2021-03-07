// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	// "log"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
	"snowfrost.garden/donk/transpiler/scope"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func genericCtxtCall(n string) *cctpb.Expression {
	return vsk.ObjMember(vsk.StringIdExpr("ctxt"), vsk.FuncCall(vsk.Id(n)))
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
				vsk.StringLiteralExpr(p.FullyQualifiedString()),
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
	vsk.AddFuncArg(fc, &cctpb.Expression{Value: &cctpb.Expression_LiteralExpression{l}})
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func getObjFunc(f string) *cctpb.Expression {
	fc := &cctpb.FunctionCallExpression{Name: vsk.IdExpr(vsk.Id("p"))}
	vsk.AddFuncArg(fc, vsk.StringLiteralExpr(f))
	return &cctpb.Expression{
		Value: &cctpb.Expression_FunctionCallExpression{fc},
	}
}

func declareVar(name string) *cctpb.Declaration {
	return vsk.VarDecl(name, vsk.NsId("donk", "var_t"))
}

func resourceId(resource string) *cctpb.Expression {
	i := vsk.NsId("donk", "resource_t")
	fc := &cctpb.FunctionCallExpression{
		Name: &cctpb.Expression{
			Value: &cctpb.Expression_IdentifierExpression{i},
		},
	}
	vsk.AddFuncArg(fc, vsk.StringLiteralExpr(resource))
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
	vsk.AddFuncArg(fc, &cctpb.Expression{
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
	vsk.AddFuncArg(fc, &cctpb.Expression{
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
	vsk.AddFuncArg(fc, &cctpb.Expression{
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
		if t.curScope().HasLocal(rId) && t.curScope().VarType(rId) == scope.VarTypeDMObject {
			return true
		}
	}

	return false
}

func (t Transformer) isDMObject(e *cctpb.Expression) bool {
	if isRawIdentifier(e) {
		rId := rawIdentifier(e)
		if t.curScope().HasField(rId) && t.curScope().VarType(rId) == scope.VarTypeDMObject {
			return true
		}
		if t.curScope().HasLocal(rId) && t.curScope().VarType(rId) == scope.VarTypeDMObject {
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

func wrapDeclarationInStatment(decl *cctpb.Declaration) *cctpb.Statement {
	return &cctpb.Statement{
		Value: &cctpb.Statement_DeclarationStatement{decl},
	}
}

func (t *Transformer) declareInt(name string) *cctpb.Declaration {
	return vsk.VarDecl(name, &cctpb.Identifier{
		Id: proto.String("int"),
	})
}

func VarDeclAndInitializer(name string, ident *cctpb.Identifier, init *cctpb.Initializer) *cctpb.Declaration {
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

func (t *Transformer) declareVarWithVal(name string, val *cctpb.Expression, varType scope.VarType) *cctpb.Statement {
	ds := &cctpb.DeclarationSpecifier{
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
	if isStringLiteral(val) {
		if varType == scope.VarTypeDMObject {
			t.curScope().AddDefnHeader("\"donk/core/vars.h\"")
			vstr := vsk.NsId("donk", "var_t::str")
			fc := vsk.FuncCall(vstr)
			vsk.AddFuncArg(fc.GetFunctionCallExpression(), val)
			val = fc
		} else {
			t.curScope().AddDefnHeader("<string>")
			val = vsk.StdStringCtor(val.GetLiteralExpression().GetStringLiteral())
		}
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

func isRawString(expr *astpb.Expression) bool {
	if expr.GetBase().GetTerm().StringT != nil {
		return true
	}
	if expr.GetBase().GetTerm().GetExpr() != nil {
		return isRawString(expr.GetBase().GetTerm().GetExpr())
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

func (t Transformer) isSleep(s *astpb.Statement) bool {
	call := s.GetExpr().GetBase().GetTerm().GetCall()
	if call == nil {
		return false
	}
	if call.GetS() != "sleep" {
		return false
	}
	return true
}

func (t Transformer) astpbSleepToCctpbSleep(s *astpb.Statement) *cctpb.Statement {
	ticks := rawInt(s.GetExpr().GetBase().GetTerm().GetCall().GetExpr()[0])
	fc := vsk.FuncCall(vsk.Id("sleep"))
	vsk.AddFuncArg(fc.GetFunctionCallExpression(), vsk.IntLiteralExpr(ticks))
	ctxt := vsk.StringIdExpr("ctxt")
	mae := vsk.ObjMember(ctxt, fc)
	return &cctpb.Statement{
		Value: &cctpb.Statement_CoYield{
			&cctpb.CoYield{
				Expr: mae,
			},
		},
	}
}

func (t Transformer) makeApiFuncDecl(name string) *cctpb.FunctionDeclaration {
	fd := &cctpb.FunctionDeclaration{
		Name: proto.String(name),
		ReturnType: &cctpb.CppType{
			PType: cctpb.CppType_NONE.Enum(),
			Name:  proto.String("donk::running_proc"),
		},
	}
	t.curScope().AddDeclHeader("\"cppcoro/generator.hpp\"")
	t.curScope().AddDeclHeader("\"donk/core/procs.h\"")

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
