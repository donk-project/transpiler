// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"io/ioutil"
	"log"

	"github.com/bazelbuild/rules_go/go/tools/bazel"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

type Transformer struct {
	coreNamespace string
	coreParser    *parser.Parser
	curScope      *scopeCtxt
	includePrefix string
	lastFileId    uint32
	parser        *parser.Parser
	project       *cctpb.Project
	walkPerformed bool
}

func (t Transformer) Project() *cctpb.Project {
	return t.project
}

func (t Transformer) CoreNamespace() string {
	return t.coreNamespace
}

func (t Transformer) isCoreGen() bool {
	return t.coreNamespace == "donk"
}

func New(p *parser.Parser, cn string, ip string) *Transformer {
	t := &Transformer{
		parser:        p,
		coreNamespace: cn,
		project:       &cctpb.Project{},
		includePrefix: ip,
	}

	binaryPbPath, err := bazel.Runfile("core.binarypb")
	if err != nil {
		panic(err)
	}
	in, err := ioutil.ReadFile(binaryPbPath)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	g := &astpb.Graph{}
	if err := proto.Unmarshal(in, g); err != nil {
		log.Fatalln("Failed to parse:", err)
	}
	t.coreParser = parser.NewParser(g)
	t.lastFileId = 1

	return t
}

func (t Transformer) BeginTransform() {
	if t.walkPerformed {
		log.Panic("transformer has already walked graph, cannot reuse")
	}
	t.walkPerformed = true
	t.curScope = &scopeCtxt{
		curPath: paths.New("/"),
	}

	for path, typ := range t.parser.TypesByPath {
		t.curScope = t.curScope.childType(&path, typ)
		t.scan(path)
		t.curScope = t.curScope.parent()
	}
}

func (t Transformer) makeVarRepresentation(v *parser.DMVar) *varRepresentation {
	vr := &varRepresentation{
		name: v.Name,
	}

	return vr
}

func (t Transformer) scan(p paths.Path) {
	t.buildDeclFile()
	t.buildDefnFile()

	for _, dmvar := range t.curScope.curType.Vars {
		t.curScope.declaredVars = append(
			t.curScope.declaredVars, *t.makeVarRepresentation(dmvar))
	}

	ns := proto.String(t.coreNamespace + "::" + p.AsNamespace())
	if p.IsRoot() {
		ns = proto.String(t.coreNamespace)
	}

	nsDecl := &cctpb.NamespacedDeclaration{
		Namespace: ns,
	}
	nsDefn := &cctpb.NamespacedDefinition{
		Namespace: ns,
	}

	if t.isCoreGen() && !t.curScope.curPath.IsRoot() {
		nsDecl.ClassDeclarations = append(nsDecl.ClassDeclarations, t.buildCoretypeDecl())
	}

	procCount := 0
	for _, dmproc := range t.curScope.curType.Procs {
		if t.shouldEmitProc(dmproc) {
			funcDecl := t.makeFuncDecl(dmproc)
			nsDecl.FunctionDeclarations = append(nsDecl.FunctionDeclarations, funcDecl)
			procCount++
			funcDefn := t.makeFuncDefn(dmproc)
			funcDefn.Declaration = funcDecl
			nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, funcDefn)
		}
	}

	if t.isCoreGen() && t.curScope.curPath.Equals("/world") {
		procCount += 1
		fd := &cctpb.FunctionDeclaration{
			Name: proto.String(BroadcastRedirectProcName),
			ReturnType: &cctpb.CppType{
				PType: cctpb.CppType_NONE.Enum(),
				Name:  proto.String("void"),
			},
		}

		fd.Arguments = append(fd.Arguments,
			&cctpb.FunctionArgument{
				Name: proto.String("ctxt"),
				CppType: &cctpb.CppType{
					PType: cctpb.CppType_REFERENCE.Enum(),
					Name:  proto.String("donk::proc_ctxt_t"),
				},
			})

		fd.Arguments = append(fd.Arguments,
			&cctpb.FunctionArgument{
				Name: proto.String("args"),
				CppType: &cctpb.CppType{
					PType: cctpb.CppType_REFERENCE.Enum(),
					Name:  proto.String("donk::proc_args_t"),
				},
			})

		nsDecl.FunctionDeclarations = append(nsDecl.FunctionDeclarations, fd)
		broadcastFunctionDefn := &cctpb.FunctionDefinition{
			Declaration: fd,
		}
		nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, broadcastFunctionDefn)
	}

	if procCount > 0 {
		t.curScope.addDeclHeader("\"donk/core/procs.h\"")
	}

	if t.isCoreGen() && !t.curScope.curPath.IsRoot() {
		nsDefn.Constructors = append(nsDefn.Constructors, t.generateCoreConstructor())
		registerFd := &cctpb.FunctionDeclaration{
			Name: proto.String("InternalCoreRegister"),
			ReturnType: &cctpb.CppType{
				PType: cctpb.CppType_NONE.Enum(),
				Name:  proto.String("void"),
			},
			MemberOf: &cctpb.Identifier{
				Id: proto.String(t.curScope.curPath.Basename + "_coretype"),
			},
		}
		internalCoreReg := t.generateInternalCoreRegister(*ns)
		internalCoreReg.Declaration = registerFd
		nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, internalCoreReg)
	} else {
		t.curScope.addDeclHeader("\"donk/core/iota.h\"")
		t.curScope.addDefnHeader("\"donk/core/iota.h\"")
		registerFd := &cctpb.FunctionDeclaration{
			Name: proto.String("Register"),
			ReturnType: &cctpb.CppType{
				PType: cctpb.CppType_NONE.Enum(),
				Name:  proto.String("void"),
			},
		}
		registerFd.Arguments = append(registerFd.Arguments,
			&cctpb.FunctionArgument{
				Name: proto.String("iota"),
				CppType: &cctpb.CppType{
					PType: cctpb.CppType_REFERENCE.Enum(),
					Name:  proto.String("donk::iota_t"),
				},
			})

		nsDecl.FunctionDeclarations = append(
			nsDecl.FunctionDeclarations, registerFd)
		registerFuncDefn := t.generateRegistrationFunction(*ns)
		registerFuncDefn.Declaration = registerFd
		nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, registerFuncDefn)
	}

	t.curScope.addDeclHeader("\"donk/core/procs.h\"")
	t.curScope.addDefnHeader("\"donk/core/procs.h\"")

	t.curScope.curDeclFile.NamespacedDeclarations = append(
		t.curScope.curDeclFile.NamespacedDeclarations, nsDecl)

	t.curScope.curDefnFile.NamespacedDefinitions = append(
		t.curScope.curDefnFile.NamespacedDefinitions, nsDefn)

	t.project.DeclarationFiles = append(t.project.DeclarationFiles, t.curScope.curDeclFile)
	t.project.DefinitionFiles = append(t.project.DefinitionFiles, t.curScope.curDefnFile)

	for hdr := range t.curScope.curDeclHeaders {
		t.curScope.curDeclFile.Preamble.Headers = append(
			t.curScope.curDeclFile.Preamble.Headers, hdr)
	}
	for hdr := range t.curScope.curDefnHeaders {
		t.curScope.curDefnFile.Preamble.Headers = append(
			t.curScope.curDefnFile.Preamble.Headers, hdr)
	}
}

func (t Transformer) makeFuncDefn(p *parser.DMProc) *cctpb.FunctionDefinition {
	t.curScope = t.curScope.child()
	t.curScope.curProc = p

	proc := p.Proto.Value[len(p.Proto.Value)-1]
	funcDefn := &cctpb.FunctionDefinition{}
	funcDefn.BlockDefinition = t.walkBlock(proc.GetCode().GetPresent())

	t.curScope = t.curScope.parentScope
	return funcDefn
}

func (t Transformer) makeFuncDecl(p *parser.DMProc) *cctpb.FunctionDeclaration {
	t.curScope = t.curScope.child()
	t.curScope.curProc = p

	fd := &cctpb.FunctionDeclaration{
		Name: proto.String(p.EmitName()),
		ReturnType: &cctpb.CppType{
			PType: cctpb.CppType_NONE.Enum(),
			Name:  proto.String("void"),
		},
	}

	fd.Arguments = append(fd.Arguments,
		&cctpb.FunctionArgument{
			Name: proto.String("ctxt"),
			CppType: &cctpb.CppType{
				PType: cctpb.CppType_REFERENCE.Enum(),
				Name:  proto.String("donk::proc_ctxt_t"),
			},
		})

	fd.Arguments = append(fd.Arguments,
		&cctpb.FunctionArgument{
			Name: proto.String("args"),
			CppType: &cctpb.CppType{
				PType: cctpb.CppType_REFERENCE.Enum(),
				Name:  proto.String("donk::proc_args_t"),
			},
		})

	t.curScope = t.curScope.parentScope
	return fd
}

func (t Transformer) walkBlock(block *astpb.Block) *cctpb.BlockDefinition {
	blockDef := &cctpb.BlockDefinition{}
	for _, stmt := range block.GetStatement() {
		blockDef.Statements = append(blockDef.Statements, t.walkStatement(stmt))
	}
	return blockDef
}

func (t Transformer) declaredInPathOrParents(name string) bool {
	if t.curScope.HasField(name) {
		return true
	}

	p := t.curScope.curPath.ParentPath()
	for !p.IsRoot() {
		if _, ok := t.parser.VarsByPath[*p.Child(name)]; ok {
			return true
		}
		if _, ok := t.coreParser.VarsByPath[*p.Child(name)]; ok {
			return true
		}
		p = p.ParentPath()
	}

	return false
}

func (t Transformer) passedAsArg(name string) bool {
	return t.curScope.curProc.HasArg(name)
}
