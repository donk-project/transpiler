// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"

	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func printMemberSpecifier(ms *cctpb.MemberSpecification) string {
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
	}
	return ""
}
