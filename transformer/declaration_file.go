// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"snowfrost.garden/donk/transpiler/scope"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

func (t Transformer) buildDeclFile() {
	declFilename := strings.ToLower(
		strings.TrimPrefix(t.curScope().CurType.ResolvedPath().FullyQualifiedString(), "/")) + ".h"
	if t.curScope().CurPath.IsRoot() {
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
	t.curScope().CurDeclFile = declFile
	t.curScope().CurDeclHeaders = scope.NewHeaderCollection()
	t.curScope().CurDeclFile.FileMetadata = &cctpb.FileMetadata{
		FileId:     proto.Uint32(t.lastFileId),
		SourcePath: proto.String(t.curScope().CurType.ResolvedPath().FullyQualifiedString()),
		Filename:   proto.String(declFilename),
	}
	if t.curScope().CurPath.IsRoot() {
		t.curScope().CurDeclFile.FileMetadata.Filename = proto.String(strings.ToLower(
			strings.TrimPrefix(t.curScope().CurPath.ParentPath().FullyQualifiedString(), "/")) + "root.h")
	}
}
