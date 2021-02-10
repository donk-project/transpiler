// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"strings"

	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) buildDefnFile() {
	t.curScope.curDefnFile = &cctpb.DefinitionFile{
		FileMetadata: &cctpb.FileMetadata{},
		Preamble:     &cctpb.Preamble{},
	}
	t.curScope.curDefnHeaders = make(map[string]bool)
	bName := strings.ToLower(
		strings.TrimPrefix(t.curScope.curPath.FullyQualifiedString(), "/"))
	t.curScope.curDefnFile.FileMetadata.Filename = proto.String(bName + ".cc")
	t.curScope.curDefnFile.BaseInclude = proto.String("\"" + t.includePrefix + "/" + bName + ".h\"")
	if t.curScope.curPath.IsRoot() {
		t.curScope.curDefnFile.BaseInclude = proto.String("\"" + t.includePrefix + "/" + bName + "root.h\"")
		t.curScope.curDefnFile.FileMetadata.Filename = proto.String(strings.ToLower(
			strings.TrimPrefix(t.curScope.curPath.ParentPath().FullyQualifiedString(), "/")) + "root.cc")
	}
	t.lastFileId++
	t.curScope.curDefnFile.FileMetadata.FileId = proto.Uint32(t.lastFileId)
	t.curScope.curDefnFile.FileMetadata.SourcePath = proto.String(t.curScope.curPath.FullyQualifiedString())
}
