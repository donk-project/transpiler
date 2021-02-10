// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) buildDeclFile() {
	declFilename := strings.ToLower(
		strings.TrimPrefix(t.curScope.curPath.FullyQualifiedString(), "/")) + ".h"
	if t.curScope.curPath.IsRoot() {
		declFilename = "root.h"
	}
	ifdef := fmt.Sprintf("__%v_%v__",
		strings.ToUpper(t.includePrefix),
		strings.ToUpper(declFilename))
	ifdef = strings.ReplaceAll(ifdef, "/", "_")
	ifdef = strings.ReplaceAll(ifdef, ".", "_")
	declFile := &cctpb.DeclarationFile{
		IncludeGuard: proto.String(ifdef),
		Preamble:     &cctpb.Preamble{},
	}
	t.lastFileId++
	t.curScope.curDeclFile = declFile
	t.curScope.curDeclHeaders = make(map[string]bool)

	t.curScope.curDeclFile.FileMetadata = &cctpb.FileMetadata{
		FileId:     proto.Uint32(t.lastFileId),
		SourcePath: proto.String(t.curScope.curPath.FullyQualifiedString()),
		Filename:   proto.String(declFilename),
	}
	if t.curScope.curPath.IsRoot() {
		t.curScope.curDeclFile.FileMetadata.Filename = proto.String(strings.ToLower(
			strings.TrimPrefix(t.curScope.curPath.ParentPath().FullyQualifiedString(), "/")) + "root.h")
	}

}
