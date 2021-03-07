// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) walkTerm(term *astpb.Term) *cctpb.Expression {
	e := &cctpb.Expression{}

	switch {
	case term.NullT != nil:
		{
			exp := &cctpb.Identifier{Id: proto.String("nullptr")}
			e.Value = &cctpb.Expression_IdentifierExpression{exp}
		}
	case term.Ident != nil:
		{
			id := term.GetIdent()
			exp := &cctpb.Identifier{
				Id: proto.String(id),
			}
			if term.GetIdent() == "world" {
				e = genericCtxtCall("world")
			} else if t.declaredInPathOrParents(id) {
				mae := &cctpb.MemberAccessExpression{
					Operator: cctpb.MemberAccessExpression_MEMBER_OF_POINTER.Enum(),
					Lhs:      ctxtSrc(),
					Rhs:      getObjVar(id),
				}
				e.Value = &cctpb.Expression_MemberAccessExpression{mae}
			} else if t.passedAsArg(id) {
				e = vsk.ObjMember(argParam(), getObjVar(id))
			} else if t.curScope().HasGlobal(id) {
				e = genericCtxtCall("cvar")
				fce := e.GetMemberAccessExpression().GetRhs().GetFunctionCallExpression()
				vsk.AddFuncArg(fce, &cctpb.Expression{
					Value: &cctpb.Expression_LiteralExpression{
						&cctpb.Literal{Value: &cctpb.Literal_StringLiteral{id}},
					},
				})
			} else if term.GetIdent() == "TRUE" {
				exp.Id = proto.String("true")
				e.Value = &cctpb.Expression_IdentifierExpression{exp}
			} else if term.GetIdent() == "FALSE" {
				exp.Id = proto.String("false")
				e.Value = &cctpb.Expression_IdentifierExpression{exp}
			} else if term.GetIdent() == "src" {
				// maybe shouldn't immediately do this if it's part of a chained call?
				e = ctxtSrc()
			} else if term.GetIdent() == "usr" {
				e = ctxtUsr()
			} else {
				e.Value = &cctpb.Expression_IdentifierExpression{exp}
			}
		}

	case term.FloatT != nil:
		{
			e.Value = &cctpb.Expression_LiteralExpression{&cctpb.Literal{
				Value: &cctpb.Literal_FloatLiteral{*term.FloatT},
			}}
		}

	case term.IntT != nil:
		{
			e.Value = &cctpb.Expression_LiteralExpression{&cctpb.Literal{
				Value: &cctpb.Literal_IntegerLiteral{int64(*term.IntT)},
			}}
		}

	case term.StringT != nil:
		{
			e = vsk.StringLiteralExpr(*term.StringT)
		}

	case term.Resource != nil:
		{
			t.curScope().AddDefnHeader("\"donk/core/vars.h\"")
			e = resourceId(*term.Resource)
		}

	case term.GetCall() != nil:
		{
			if term.GetCall().S == nil {
				panic(fmt.Sprintf("call with no identifier %v", proto.MarshalTextString(term.GetCall())))
			}
			if t.IsProcInCore(term.GetCall().GetS()) {
				fn := term.GetCall().GetS()
				rooted := "/" + fn
				_, ok := knownSynchronousProcs[rooted]
				invokeType := InvokeTypeAsyncProc
				if ok {
					invokeType = InvokeTypeSyncProc
				}
				var transformed []*cctpb.Expression
				for _, ex := range term.GetCall().GetExpr() {
					transformed = append(transformed, t.walkExpression(ex))
				}
				e = t.procCall(
					genericCtxtCall("Global"), fn, transformed, invokeType)
			} else {
				// TODO: This is definitely wrong
				fc := &cctpb.FunctionCallExpression{}
				fc.Name = &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{
						&cctpb.Identifier{
							Id: proto.String(term.GetCall().GetS()),
						},
					},
				}
				for _, ex := range term.GetCall().GetExpr() {
					vsk.AddFuncArg(fc, t.walkExpression(ex))
				}
				e.Value = &cctpb.Expression_FunctionCallExpression{fc}
			}
		}

	case term.GetPrefab() != nil:
		{
			e = t.walkPrefab(term.GetPrefab())
		}

	case term.GetInterpString() != nil:
		{
			is := term.GetInterpString()
			t.curScope().AddDefnHeader("\"fmt/format.h\"")
			id := &cctpb.Identifier{
				Namespace: proto.String("fmt"),
				Id:        proto.String("format"),
			}
			fce := &cctpb.FunctionCallExpression{
				Name: &cctpb.Expression{
					Value: &cctpb.Expression_IdentifierExpression{id},
				},
			}

			var stringPieces []string
			var formatArgs []*cctpb.Expression

			if is.S != nil && is.GetS() != "" {
				stringPieces = append(stringPieces, *is.S)
			}

			for _, collection := range is.GetCollections() {
				if collection.GetExpr() != nil {
					colExpr := t.walkExpression(collection.GetExpr())
					if !isRawLiteral(colExpr) {
						ue := &cctpb.UnaryExpression{
							Operator: cctpb.UnaryExpression_POINTER_INDIRECTION.Enum(),
							Operand:  colExpr,
						}
						formatArgs = append(formatArgs, &cctpb.Expression{
							Value: &cctpb.Expression_UnaryExpression{ue},
						})
					} else {
						formatArgs = append(formatArgs, colExpr)
					}
					stringPieces = append(stringPieces, "{}")
				}
				if collection.S != nil {
					stringPieces = append(stringPieces, collection.GetS())
				}
			}

			fs := strings.Join(stringPieces, "")
			fsExpr := &cctpb.Expression{
				Value: &cctpb.Expression_LiteralExpression{
					&cctpb.Literal{
						Value: &cctpb.Literal_StringLiteral{fs},
					},
				},
			}
			vsk.AddFuncArg(fce, fsExpr)

			for _, formatArg := range formatArgs {
				vsk.AddFuncArg(fce, formatArg)
			}

			e.Value = &cctpb.Expression_FunctionCallExpression{fce}
		}

	case term.GetLocate() != nil:
		{
			var transformed []*cctpb.Expression
			for _, arg := range term.GetLocate().GetArgs() {
				transformed = append(transformed, t.walkExpression(arg))
			}
			e = t.procCall(genericCtxtCall("Global"), "locate", transformed, InvokeTypeSyncProc)
		}

	case term.GetNew() != nil:
		{
			typ := term.GetNew().GetType()
			if typ.GetPrefab() != nil && len(typ.GetPrefab().Path) > 0 {
				p := paths.NewFromTypePaths(typ.GetPrefab().Path)
				e = ctxtMakeCall(*p)
			} else {
				e = genericCtxtCall("make")
			}
		}

	case term.GetExpr() != nil:
		{
			e = t.walkExpression(term.GetExpr())
		}

	case term.GetList() != nil:
		{
			fce := &cctpb.FunctionCallExpression{}
			fce.Name = &cctpb.Expression{
				Value: &cctpb.Expression_IdentifierExpression{
					&cctpb.Identifier{
						Namespace: proto.String("donk"),
						Id:        proto.String("assoc_list_t"),
					},
				},
			}

			iList := &cctpb.InitializerList{}

			for _, call := range term.GetList().GetCall() {
				iList.Args = append(iList.Args, t.walkExpression(call))
			}
			fce.Arguments = append(fce.Arguments,
				&cctpb.FunctionCallExpression_ExpressionArg{
					Value: &cctpb.FunctionCallExpression_ExpressionArg_InitializerList{iList},
				})

			e.Value = &cctpb.Expression_FunctionCallExpression{fce}
		}

	default:
		panic(fmt.Sprintf("cannot walk unsupported term %v", proto.MarshalTextString(term)))
	}

	return e
}
