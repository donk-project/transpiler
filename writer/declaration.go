// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func printDeclaration(d *cctpb.Declaration) string {
	switch d.Value.(type) {
	case *cctpb.Declaration_BlockDeclaration:
		{
			return printBlockDeclaration(d.GetBlockDeclaration())
		}
	default:
		panic(fmt.Sprintf("cannot print unsupported declaration %v", proto.MarshalTextString(d)))
	}
}

func printBlockDeclaration(bd *cctpb.BlockDeclaration) string {
	switch bd.Value.(type) {
	case *cctpb.BlockDeclaration_SimpleDeclaration:
		{
			return printSimpleDeclaration(bd.GetSimpleDeclaration())
		}
	default:
		panic(fmt.Sprintf("cannot print unsupported block declaration %v", proto.MarshalTextString(bd)))
	}
}

func printSimpleDeclaration(sd *cctpb.SimpleDeclaration) string {
	var specs []string
	for _, s := range sd.GetSpecifiers() {
		specs = append(specs, printDeclarationSpecifier(s))
	}

	var decls []string
	for _, d := range sd.GetDeclarators() {
		decls = append(decls, printDeclarator(d))
	}

	return strings.Trim(strings.Join(specs, " ")+" "+strings.Join(decls, " "), " ")
}

func printDeclarationSpecifier(ds *cctpb.DeclarationSpecifier) string {
	switch ds.GetValue().(type) {
	case *cctpb.DeclarationSpecifier_TypeSpecifier:
		{
			return printTypeSpecifier(ds.GetTypeSpecifier())
		}
	default:
		panic(fmt.Sprintf("cannot print unsupported declaration specifier %v", proto.MarshalTextString(ds)))
	}
}

func printDeclarator(d *cctpb.Declarator) string {
	format := "%v"
	if d.GetInitializer() != nil {
		switch d.GetInitializer().Value.(type) {
		case *cctpb.Initializer_CopyInitializer:
			{
				format = "%v = " + printExpression(d.GetInitializer().GetCopyInitializer().GetOther())
			}
		default:
			panic(fmt.Sprintf("cannot print unsupported initializer %v", proto.MarshalTextString(d)))
		}
	}

	switch d.GetValue().(type) {
	case *cctpb.Declarator_DeclaredName:
		{
			return fmt.Sprintf(format, printIdentifier(d.GetDeclaredName()))
		}
	default:
		panic(fmt.Sprintf("cannot print unsupported declarator %v", proto.MarshalTextString(d)))
	}
}

func printTypeSpecifier(ts *cctpb.TypeSpecifier) string {
	switch ts.GetValue().(type) {
	case *cctpb.TypeSpecifier_SimpleTypeSpecifier:
		{
			return printSimpleTypeSpecifier(ts.GetSimpleTypeSpecifier())
		}
	default:
		panic(fmt.Sprintf("cannot print unsupported type specifier %v", proto.MarshalTextString(ts)))
	}
}

func printSimpleTypeSpecifier(sts *cctpb.SimpleTypeSpecifier) string {
	switch sts.GetValue().(type) {
	case *cctpb.SimpleTypeSpecifier_DeclaredName:
		{
			return printIdentifier(sts.GetDeclaredName())
		}
	default:
		panic(fmt.Sprintf("cannot print unsupported simple type specifier %v", proto.MarshalTextString(sts)))
	}
}
