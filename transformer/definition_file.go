// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"snowfrost.garden/donk/transpiler/scope"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) buildDefnFile() {
	t.curScope().CurDefnFile = &cctpb.DefinitionFile{
		FileMetadata: &cctpb.FileMetadata{},
		Preamble:     &cctpb.Preamble{},
	}
	t.curScope().CurDefnHeaders = scope.NewHeaderCollection()
	bName := strings.ToLower(
		strings.TrimPrefix(t.curScope().CurType.ResolvedPath().FullyQualifiedString(), "/"))
	incPrefix := ""
	if t.curScope().CurPath.IsRoot() {
		t.curScope().CurDefnFile.BaseInclude = proto.String("\"" + incPrefix + "root.h\"")
		t.curScope().CurDefnFile.FileMetadata.Filename = proto.String(strings.ToLower(
			strings.TrimPrefix(t.curScope().CurPath.ParentPath().FullyQualifiedString(), "/")) + "root.cc")
	} else {
		t.curScope().CurDefnFile.BaseInclude = proto.String("\"" + incPrefix + bName + ".h\"")
		t.curScope().CurDefnFile.FileMetadata.Filename = proto.String(bName + ".cc")
	}

	t.lastFileId++

	t.curScope().CurDefnFile.FileMetadata.FileId = proto.Uint32(t.lastFileId)
	t.curScope().CurDefnFile.FileMetadata.SourcePath = proto.String(t.curScope().CurType.ResolvedPath().FullyQualifiedString())
}
