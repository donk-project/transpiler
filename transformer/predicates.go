// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
)

func (t *Transformer) shouldEmitProc(p *parser.DMProc) bool {
	if t.isCoreGen() {
		return true
	}
	// more than one value means a subclass version of the same proc
	if len(p.Proto.Value) > 1 {
		return true
	}
	if p.Proto.GetDeclaration().GetLocation().GetFile() != nil {
		return *p.Proto.GetDeclaration().GetLocation().GetFile().FileId != 0
	} else if p.Proto.GetValue() != nil {
		for _, value := range p.Proto.GetValue() {
			if value.GetCode().GetPresent() != nil {
				return true
			}
		}
	}
	return false
}

func (t *Transformer) shouldEmitVar(v *parser.DMVar) bool {
	if t.isCoreGen() {
		return true
	}
	if v.Proto.GetValue() != nil {
		if v.Proto.GetValue().GetLocation().GetFile().GetFileId() == 0 {
			return false
		}
		return true
	}

	return v.Proto.GetDeclaration().GetLocation().GetFile().GetFileId() != 0
}

func (t *Transformer) IsProcInCore(name string) bool {
	_, ok := t.coreParser.ProcsByPath[*paths.New("/" + name)]
	if ok {
		return true
	}
	return false
}

// func (e *Emitter) ShouldEmitVar(p *parser.DMVar) bool {
// 	if e.CoreNamespace == "donk" {
// 		return true
// 	}
// 	if p.Proto.Value != nil {
// 		if p.Proto.Value.Location != nil && p.Proto.Value.Location.File != nil && p.Proto.Value.Location.File.FileId != nil {
// 			if *p.Proto.Value.Location.File.FileId == 0 {
// 				return false
// 			}
// 		}
// 		return true
// 	}

// 	return p.Proto.Decl != nil &&
// 		p.Proto.Decl.Location != nil &&
// 		p.Proto.Decl.Location.File != nil &&
// 		*p.Proto.Decl.Location.File.FileId != 0

// }

// func (e *Emitter) IsVarInCore(v *parser.DMVar) bool {
// 	_, ok := e.CoreParser.VarsByPath[*paths.New("/" + v.Name)]
// 	if ok {
// 		return true
// 	}
// 	return false
// }

// func (e *Emitter) IsVarInCoretype(v *parser.DMVar) bool {
// 	// Counterintuitively we say a var isn't in the coretype if we're
// 	// doing coregen, because we want them printed regardless, and that's
// 	// what we use this function for.
// 	if e.CoreNamespace == "donk" {
// 		return false
// 	}
// 	p := v.Type.Path.Child(v.Name)
// 	_, ok := e.CoreParser.VarsByPath[*p]
// 	if ok {
// 		return true
// 	}
// 	return false
// }

// func (e *Emitter) IsTypeInCore(typ *parser.DMType) bool {
// 	_, ok := e.CoreParser.TypesByPath[*typ.Path]
// 	if ok {
// 		return true
// 	}
// 	return false
// }

// func (e *Emitter) ShouldBeReference(v *parser.DMVar) bool {
// 	if strings.HasSuffix(e.EmitType(v), "var_t") || strings.HasSuffix(e.EmitType(v), "list_t") {
// 		return false
// 	}
// 	return strings.HasPrefix(e.EmitType(v), "::"+e.CoreNamespace) ||
// 		strings.HasPrefix(e.EmitType(v), "::donk")
// }
