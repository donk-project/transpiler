// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package writer

import (
	"fmt"
	"strings"

	cctpb "snowfrost.garden/vasker/cc_grammar"
)

var UNARY_OP_FORMATTERS = map[cctpb.UnaryExpression_Operator]string{
	cctpb.UnaryExpression_ADDRESS_OF:          "&%v",
	cctpb.UnaryExpression_POINTER_INDIRECTION: "*(%v)",
}

var ARITHMETIC_OP_FORMATTERS = map[cctpb.ArithmeticExpression_Operator]string{
	cctpb.ArithmeticExpression_ADDITION:       "%v + %v",
	cctpb.ArithmeticExpression_SUBTRACTION:    "%v - %v",
	cctpb.ArithmeticExpression_MULTIPLICATION: "%v * %v",
	cctpb.ArithmeticExpression_DIVISION:       "%v / %v",
	cctpb.ArithmeticExpression_MODULO:         "%v %% %v",
	cctpb.ArithmeticExpression_BITWISE_AND:    "%v & %v",
	cctpb.ArithmeticExpression_BITWISE_OR:     "%v | %v",
	cctpb.ArithmeticExpression_BITWISE_XOR:    "%v ^ %v",
	cctpb.ArithmeticExpression_BITWISE_LSHIFT: "%v << %v",
	cctpb.ArithmeticExpression_BITWISE_RSHIFT: "%v >> %v",
}

var ASSIGNMENT_OP_FORMATTERS = map[cctpb.AssignmentExpression_Operator]string{
	cctpb.AssignmentExpression_SIMPLE:         "%v = %v",
	cctpb.AssignmentExpression_ADDITION:       "%v += %v",
	cctpb.AssignmentExpression_SUBTRACTION:    "%v -= %v",
	cctpb.AssignmentExpression_MULTIPLICATION: "%v *= %v",
	cctpb.AssignmentExpression_DIVISION:       "%v /= %v",
	cctpb.AssignmentExpression_MODULO:         "%v %%= %v",
	cctpb.AssignmentExpression_BITWISE_AND:    "%v &= %v",
	cctpb.AssignmentExpression_BITWISE_XOR:    "%v ^= %v",
	cctpb.AssignmentExpression_BITWISE_OR:     "%v |= %v",
	cctpb.AssignmentExpression_BITWISE_LSHIFT: "%v <<= %v",
	cctpb.AssignmentExpression_BITWISE_RSHIFT: "%v >>= %v",
}

var COMPARISON_OP_FORMATTERS = map[cctpb.ComparisonExpression_Operator]string{
	cctpb.ComparisonExpression_EQUAL_TO:                 "%v == %v",
	cctpb.ComparisonExpression_NOT_EQUAL_TO:             "%v != %v",
	cctpb.ComparisonExpression_LESS_THAN:                "%v < %v",
	cctpb.ComparisonExpression_GREATER_THAN:             "%v > %v",
	cctpb.ComparisonExpression_LESS_THAN_OR_EQUAL_TO:    "%v <= %v",
	cctpb.ComparisonExpression_GREATER_THAN_OR_EQUAL_TO: "%v >= %v",
}

var LOGICAL_OP_FORMATTERS = map[cctpb.LogicalExpression_Operator]string{
	cctpb.LogicalExpression_LOGICAL_AND: "%v && %v",
	cctpb.LogicalExpression_LOGICAL_OR:  "%v || %v",
}

var MEMBER_ACCESS_EXPRESSION_OP_FORMATTERS = map[cctpb.MemberAccessExpression_Operator]string{
	cctpb.MemberAccessExpression_SUBSCRIPT:         "%v[%v]",
	cctpb.MemberAccessExpression_MEMBER_OF_OBJECT:  "%v.%v",
	cctpb.MemberAccessExpression_MEMBER_OF_POINTER: "%v->%v",
}

var ACCESS_SPECIFIERS = map[cctpb.AccessSpecifier]string{
	cctpb.AccessSpecifier_PRIVATE:   "private",
	cctpb.AccessSpecifier_PUBLIC:    "public",
	cctpb.AccessSpecifier_PROTECTED: "protected",
}

func (w Writer) printCppType(t *cctpb.CppType) string {
	switch t.GetPType() {
	case cctpb.CppType_NONE:
		return t.GetName()
	case cctpb.CppType_RAW_POINTER:
		return fmt.Sprintf("%v*", t.GetName())
	case cctpb.CppType_REFERENCE:
		return fmt.Sprintf("%v&", t.GetName())
	case cctpb.CppType_UNIQUE_PTR:
		return fmt.Sprint("std::unique_ptr<%v>", t.GetName())
	case cctpb.CppType_SHARED_PTR:
		return fmt.Sprint("std::shared_ptr<%v>", t.GetName())
	}
	panic("unable to print cpptype")
}

func (w Writer) joinArgs(fas []*cctpb.FunctionArgument) string {
	var result []string
	for _, fa := range fas {
		result = append(result, w.printFuncArg(fa))
	}
	return strings.Join(result, ", ")
}

func (w Writer) printFuncArg(fa *cctpb.FunctionArgument) string {
	return fmt.Sprintf("%v %v", w.printCppType(fa.GetCppType()), fa.GetName())
}

func printBaseSpecifiers(bss []*cctpb.BaseSpecifier) string {
	var specifiers []string
	for _, bs := range bss {
		specifiers = append(specifiers, fmt.Sprintf("%v %v",
			ACCESS_SPECIFIERS[bs.GetAccessSpecifier()],
			printIdentifier(bs.GetClassOrDecltype())))
	}
	return strings.Join(specifiers, ", ")
}
