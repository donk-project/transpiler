// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"fmt"
	"log"
	"strings"

	"github.com/golang/protobuf/proto"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
	"snowfrost.garden/donk/transpiler/scope"
	vsk "snowfrost.garden/vasker"
	cctpb "snowfrost.garden/vasker/cc_grammar"
)

type Transformer struct {
	coreNamespace string
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
	coreNamespace := strings.ToUpper(cn)
	coreNamespace = strings.ReplaceAll(cn, ".", "_")
	t := &Transformer{
		parser:        p,
		coreNamespace: coreNamespace,
		project:       &cctpb.Project{},
		includePrefix: ip,
	}

	t.lastFileId = 1
	t.scopeStack = scope.NewStack()

	return t
}

func (t Transformer) PushScope() {
	t.scopeStack.Push(t.curScope().MakeChild())
}

func (t Transformer) PushScopePath(p paths.Path) {
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

	root, ok := t.parser.TypesByPath[paths.New("/")]
	if ok {
		t.scan(paths.New("/"), root)
	}

	for path, typ := range t.parser.TypesByPath {
		if path.IsRoot() {
			continue
		}
		if !(t.HasEmittableVars(typ) || t.HasEmittableProcs(typ)) {
			continue
		}
		t.PushScopePath(path)
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

	var nsStr string = t.coreNamespace
	if t.isCoreGen() {
		nsStr = nsStr + "::api"
	}
	if !p.IsRoot() {
		nsStr = nsStr + "::" + typ.ResolvedPath().AsNamespace()
	}

	ns := proto.String(nsStr)

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
			if t.isCoreGen() {
				fc := vsk.FuncCall(vsk.Id("Unimplemented"))
				procStrName := strings.Replace(
					t.curScope().CurType.Path.FullyQualifiedString()+"/proc/"+dmproc.Name, "//", "/", -1)
				vsk.AddFuncArg(fc.GetFunctionCallExpression(), vsk.StringLiteralExpr(procStrName))
				funcDefn.BlockDefinition = &cctpb.BlockDefinition{}
				coYield := &cctpb.CoYield{
					Expression: vsk.ObjMember(vsk.StringIdExpr("ctxt"), fc),
				}
				funcDefn.BlockDefinition.Statements = append(funcDefn.BlockDefinition.Statements, &cctpb.Statement{
					Value: &cctpb.Statement_ExpressionStatement{
						&cctpb.Expression{
							Value: &cctpb.Expression_CoYield{coYield},
						},
					},
				})
			}
			nsDefn.FunctionDefinitions = append(nsDefn.FunctionDefinitions, funcDefn)
		}
	}

	if t.isCoreGen() && t.curScope().CurPath.Equals("/world") {
		procCount += 2
		broadcast := t.makeApiFuncDecl(BroadcastRedirectProcName)
		broadcastLog := t.makeApiFuncDecl(BroadcastLogRedirectProcName)

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
	if proc.GetCode().GetInvalid() {
		panic(fmt.Sprintf(
			"invalid code block found: \n%v\n\n", proto.MarshalTextString(proc.GetCode())))
	}
	// log.Printf("========================= Proc Definition =========================\n")
	// for _, prm := range proc.GetParameter() {
	// 	log.Printf("\n%v\n", proto.MarshalTextString(prm))
	// }
	funcDefn.BlockDefinition = t.walkBlock(proc.GetCode().GetPresent())

	funcDefn.BlockDefinition.Statements = append(funcDefn.BlockDefinition.Statements,
		&cctpb.Statement{
			Value: &cctpb.Statement_CoReturn{},
		})

	t.PopScope()
	return funcDefn
}

func (t Transformer) makeFuncDecl(p *parser.DMProc) *cctpb.FunctionDeclaration {
	fd := &cctpb.FunctionDeclaration{
		Name: proto.String(p.EmitName()),
		ReturnType: &cctpb.CppType{
			PType: cctpb.CppType_NONE.Enum(),
			Name:  proto.String("donk::running_proc"),
		},
	}
	t.curScope().AddDeclHeader("\"donk/core/procs.h\"")

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
	emptyStmt := &cctpb.Statement{}
	blockDef := &cctpb.BlockDefinition{}
	for _, stmt := range block.GetStatement() {
		walked := t.walkStatement(stmt)
		if !proto.Equal(walked, emptyStmt) {
			blockDef.Statements = append(blockDef.Statements, walked)
		}
	}
	return blockDef
}

func (t Transformer) declaredInPathOrParents(name string) bool {
	if t.curScope().HasField(name) {
		return true
	}

	p := t.curScope().CurPath.ParentPath()
	for !p.IsRoot() {
		if _, ok := t.parser.VarsByPath[p.Child(name)]; ok {
			return true
		}
		p = p.ParentPath()
	}

	return false
}

func (t Transformer) passedAsArg(name string) bool {
	return t.curScope().CurProc.HasArg(name)
}

func (t Transformer) supertypeNamespace(p *parser.DMProc) string {
	path := p.Type.Path
	for !path.IsRoot() {
		path = path.ParentPath()
		typ, ok := t.parser.TypesByPath[path]
		if ok {
			for _, proc := range typ.Procs {
				if proc.Name == p.Name {
					if proc.Type.Path.IsCoretype() {
						return "donk::api::" + proc.Type.Path.AsNamespace()
					} else if proc.Type.Path.ParentPath().IsCoretype() {
						return "donk::api::" + proc.Type.Path.ParentPath().AsNamespace()
					}

					return t.coreNamespace + "::" + proc.Type.ParentType().AsNamespace()
				}
			}
		}
	}
	panic("cannot find parent proc")
}

func (t Transformer) supertypeInclude(p *parser.DMProc) string {
	path := p.Type.Path
	for !path.IsRoot() {
		path = path.ParentPath()
		typ, ok := t.parser.TypesByPath[path]
		if ok {
			for _, proc := range typ.Procs {
				if proc.Name == p.Name {
					if proc.Type.Path.IsCoretype() {
						return "\"donk/api" + proc.Type.Path.FullyQualifiedString() + ".h\""
					} else if p.Type.Path.ParentPath().IsCoretype() {
						return "\"donk/api" + proc.Type.Path.ParentPath().FullyQualifiedString() + ".h\""
					}
					return "\"" + strings.TrimLeft(proc.Type.Path.ParentPath().FullyQualifiedString(), "/") + ".h\""
				}
			}
		}
	}
	panic("cannot find parent proc")
}
