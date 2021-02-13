// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"
	"strings"

	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (w Writer) printMemberSpecifier(ms *cctpb.MemberSpecification) string {
	switch ms.Value.(type) {
	case *cctpb.MemberSpecification_AccessSpecifier:
		{
			return ACCESS_SPECIFIERS[ms.GetAccessSpecifier()] + ":"
		}
	case *cctpb.MemberSpecification_Destructor:
		{
			result := fmt.Sprintf("~%v()", printIdentifier(ms.GetDestructor().GetClassName()))
			if ms.GetDestructor().GetBlockDefinition() != nil {
				if len(ms.GetDestructor().GetBlockDefinition().GetStatements()) == 0 {
					result += " {}"
					return result
				}
				result += " {"
				for _, stmt := range ms.GetDestructor().GetBlockDefinition().GetStatements() {
					result += "\n" + printStatement(stmt)
				}
				result += "\n}\n"
			}
			return result
		}
	case *cctpb.MemberSpecification_FunctionDeclaration:
		{
			fd := ms.GetFunctionDeclaration()
			result := fmt.Sprintf("%v %v(%v)",
				w.printCppType(fd.GetReturnType()),
				fd.GetName(),
				w.joinArgs(fd.GetArguments()))
			if ms.GetFunctionDeclaration().GetVirtSpecifier() != nil {
				result += fmt.Sprintf(" %v", VIRT_SPECIFIERS[ms.GetFunctionDeclaration().GetVirtSpecifier().GetKeyword()])
			}
			return result + ";\n"
		}
	case *cctpb.MemberSpecification_Constructor:
		{
			c := ms.GetConstructor()
			result := fmt.Sprintf("%v(%v)",
				printIdentifier(c.GetClassName()),
				w.joinArgs(c.GetArguments()))

			return result + ";\n"
		}
	case *cctpb.MemberSpecification_MemberDeclarator:
		{
			md := ms.GetMemberDeclarator()
			result := ""
			switch md.GetValue().(type) {
			case *cctpb.MemberDeclarator_DeclaredName:
				{
					result = printIdentifier(md.GetDeclaredName())
				}
			}

			if md.GetClass() {
				result = "class " + result
			}
			if md.GetFriend() {
				result = "friend " + result
			}
			return result + ";\n"
		}
	}
	return ""
}

func (w Writer) printMemberInitializerList(mi []*cctpb.MemberInitializer) string {
	if len(mi) == 0 {
		return ""
	}
	var ms []string
	for _, m := range mi {
	var exprs []string
	for _, e := range m.GetExpressions() {
		exprs = append(exprs, printExpression(e))
	}
		ms = append(ms, fmt.Sprintf("%v(%v)",
			printIdentifier(m.GetMember()),
			strings.Join(exprs, ", ")))
	}
	return " : " + strings.Join(ms, ", ")
}

func (w Writer) printConstructor(c *cctpb.Constructor) string {
	cn := printIdentifier(c.GetClassName())
	args := w.joinArgs(c.GetArguments())
	mi := w.printMemberInitializerList(c.GetMemberInitializers())
	var stmts []string
	for _, stmt := range c.GetBlockDefinition().GetStatements() {
		stmts = append(stmts, printStatement(stmt))
	}

	result := fmt.Sprintf("\n%v(%v)%v {", cn, args, mi)
	for _, stmt := range stmts {
		result += "\n" + stmt
	}
	result += "\n}\n"
	return result
}
