// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) buildCoretypeDecl() *cctpb.ClassDeclaration {
	clsDecl := &cctpb.ClassDeclaration{
		Name: proto.String(t.curScope.curPath.Basename + "_coretype"),
	}

	if t.curScope.curPath.ParentPath().IsRoot() {
		t.curScope.addDeclHeader("\"snowfrost/donk/core/iota.h\"")
		clsDecl.BaseSpecifiers = append(clsDecl.BaseSpecifiers,
			&cctpb.BaseSpecifier{
				AccessSpecifier: cctpb.AccessSpecifier_PUBLIC.Enum(),
				ClassOrDecltype: &cctpb.Identifier{
					Namespace: proto.String(t.coreNamespace),
					Id:        proto.String("iota_t"),
				},
			},
		)
	} else {
		t.curScope.addDeclHeader(
			fmt.Sprintf("\"snowfrost/donk/api%v.h\"", t.curScope.curPath.ParentPath().FullyQualifiedString()))
		clsDecl.BaseSpecifiers = append(clsDecl.BaseSpecifiers,
			&cctpb.BaseSpecifier{
				AccessSpecifier: cctpb.AccessSpecifier_PUBLIC.Enum(),
				ClassOrDecltype: &cctpb.Identifier{
					Namespace: proto.String(t.coreNamespace + "::" + t.curScope.curPath.ParentPath().AsNamespace()),
					Id:        proto.String(t.curScope.curPath.ParentPath().Basename + "_coretype"),
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
		Value: &cctpb.MemberSpecification_AccessSpecifier{
			*cctpb.AccessSpecifier_PROTECTED.Enum(),
		},
	})

	return clsDecl
}
