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
	incPrefix := t.includePrefix + "/"
	if incPrefix == "/" {
		incPrefix = ""
	}
	if t.curScope.curPath.IsRoot() {
		t.curScope.curDefnFile.BaseInclude = proto.String("\"" + incPrefix + "root.h\"")
		t.curScope.curDefnFile.FileMetadata.Filename = proto.String(strings.ToLower(
			strings.TrimPrefix(t.curScope.curPath.ParentPath().FullyQualifiedString(), "/")) + "root.cc")
	} else {
		t.curScope.curDefnFile.BaseInclude = proto.String("\"" + incPrefix + bName + ".h\"")
		t.curScope.curDefnFile.FileMetadata.Filename = proto.String(bName + ".cc")
	}

	t.lastFileId++

	t.curScope.curDefnFile.FileMetadata.FileId = proto.Uint32(t.lastFileId)
	t.curScope.curDefnFile.FileMetadata.SourcePath = proto.String(t.curScope.curPath.FullyQualifiedString())
}
