// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
	"snowfrost.garden/donk/transpiler/scope"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

type Transformer struct {
	coreNamespace string
	coreParser    *parser.Parser
	scopeStack    *scope.Stack
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

func (t Transformer) curScope() *scope.ScopeCtxt {
	return t.scopeStack.LastScope()
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
	t.scopeStack = scope.NewStack()

	return t
}

func (t Transformer) PushScope() {
	t.scopeStack.Push(t.curScope().MakeChild())
}

func (t Transformer) PushScopePath(p *paths.Path) {
	t.scopeStack.Push(t.curScope().MakeChildPath(p))
}

func (t Transformer) PopScope() {
	if t.scopeStack.Len() == 0 {
		panic(fmt.Sprintf("tried to pop top-most scope context %v", t.curScope()))
	}
	t.scopeStack.Pop()
}

func (t Transformer) BeginTransform() {
	if t.walkPerformed {
		log.Panic("transformer has already walked graph, cannot reuse")
	}

	t.walkPerformed = true
	t.scopeStack.Push(scope.MakeRoot())

	root := t.parser.TypesByPath[*paths.New("/")]
	t.scan(*paths.New("/"), root)

	for path, typ := range t.parser.TypesByPath {
		if path.IsRoot() {
			continue
		}
		t.PushScopePath(&path)
		t.curScope().CurType = typ
		t.scan(path, typ)
		t.PopScope()
	}
}

func (t Transformer) makeVarRepresentation(v *parser.DMVar) *scope.VarInScope {
	vr := &scope.VarInScope{
		Name: v.Name,
	}

	return vr
}

func (t Transformer) scan(p paths.Path, typ *parser.DMType) {
	t.curScope().CurType = typ

	for _, dmvar := range t.curScope().CurType.Vars {
		vr := t.makeVarRepresentation(dmvar)
		if !t.curScope().HasParent() {
			vr.Scope = scope.VarScopeGlobal
		} else {
			vr.Scope = scope.VarScopeField
		}
		t.curScope().AddScopedVar(*vr)
	}

	for _, dmproc := range t.curScope().CurType.Procs {
		t.curScope().DeclaredProcs.Add(dmproc.Name)
	}

	t.buildDeclFile()
	t.buildDefnFile()

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

	if t.isCoreGen() && !t.curScope().CurPath.IsRoot() {
		nsDecl.ClassDeclarations = append(nsDecl.ClassDeclarations, t.buildCoretypeDecl())
	}

	procCount := 0
	for _, dmproc := range t.curScope().CurType.Procs {
		if t.shouldEmitProc(dmproc) {
			funcDecl := t.makeFuncDecl(dmproc)
			nsDecl.FunctionDeclarations = append(nsDecl.FunctionDeclarations, funcDecl)
			procCount++
			funcDefn := t.makeFuncDefn(dmproc)
			funcDefn.Declaration = funcDecl
			nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, funcDefn)
		}
	}

	if t.isCoreGen() && t.curScope().CurPath.Equals("/world") {
		procCount += 2
		broadcast := makeApiFuncDecl(BroadcastRedirectProcName)
		broadcastLog := makeApiFuncDecl(BroadcastLogRedirectProcName)

		nsDecl.FunctionDeclarations = append(nsDecl.FunctionDeclarations, broadcast)
		nsDecl.FunctionDeclarations = append(nsDecl.FunctionDeclarations, broadcastLog)
		nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, &cctpb.FunctionDefinition{
			Declaration: broadcast,
		})
		nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, &cctpb.FunctionDefinition{
			Declaration: broadcastLog,
		})
	}

	if procCount > 0 {
		t.curScope().AddDeclHeader("\"donk/core/procs.h\"")
	}

	if t.isCoreGen() && !t.curScope().CurPath.IsRoot() {
		nsDefn.Constructors = append(nsDefn.Constructors, t.generateCoreConstructor())
		registerFd := &cctpb.FunctionDeclaration{
			Name: proto.String("InternalCoreRegister"),
			ReturnType: &cctpb.CppType{
				PType: cctpb.CppType_NONE.Enum(),
				Name:  proto.String("void"),
			},
			MemberOf: &cctpb.Identifier{
				Id: proto.String(t.curScope().CurPath.Basename + "_coretype"),
			},
		}
		internalCoreReg := t.generateInternalCoreRegister(*ns)
		internalCoreReg.Declaration = registerFd
		nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, internalCoreReg)
	} else {
		t.curScope().AddDeclHeader("\"donk/core/iota.h\"")
		t.curScope().AddDefnHeader("\"donk/core/iota.h\"")
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

	t.curScope().AddDeclHeader("\"donk/core/procs.h\"")
	t.curScope().AddDefnHeader("\"donk/core/procs.h\"")

	t.curScope().CurDeclFile.NamespacedDeclarations = append(
		t.curScope().CurDeclFile.NamespacedDeclarations, nsDecl)

	t.curScope().CurDefnFile.NamespacedDefinitions = append(
		t.curScope().CurDefnFile.NamespacedDefinitions, nsDefn)

	t.project.DeclarationFiles = append(t.project.DeclarationFiles, t.curScope().CurDeclFile)
	t.project.DefinitionFiles = append(t.project.DefinitionFiles, t.curScope().CurDefnFile)

	for hdr := range t.curScope().CurDeclHeaders.Headers {
		t.curScope().CurDeclFile.Preamble.Headers = append(
			t.curScope().CurDeclFile.Preamble.Headers, hdr)
	}
	for hdr := range t.curScope().CurDefnHeaders.Headers {
		t.curScope().CurDefnFile.Preamble.Headers = append(
			t.curScope().CurDefnFile.Preamble.Headers, hdr)
	}
}

func (t Transformer) makeFuncDefn(p *parser.DMProc) *cctpb.FunctionDefinition {
	t.PushScope()
	t.curScope().CurProc = p

	proc := p.Proto.Value[len(p.Proto.Value)-1]
	funcDefn := &cctpb.FunctionDefinition{}
	funcDefn.BlockDefinition = t.walkBlock(proc.GetCode().GetPresent())

	t.PopScope()
	return funcDefn
}

func (t Transformer) makeFuncDecl(p *parser.DMProc) *cctpb.FunctionDeclaration {
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
	if t.curScope().HasField(name) {
		return true
	}

	p := t.curScope().CurPath.ParentPath()
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
	return t.curScope().CurProc.HasArg(name)
}
