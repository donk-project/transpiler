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

	r := t.curScope().HasGlobalProc(name)
	return r
}
