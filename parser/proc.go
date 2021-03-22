// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package parser

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/paths"
)

type ProcAccess int

const (
	ProcAccessUnknown = iota
	ProcAccessInView
	ProcAccessInOView
	ProcAccessInUsrLoc
	ProcAccessInUsr
	ProcAccessInWorld
	ProcAccessEqUsr
	ProcAccessInGroup
)

type ArgInputSchema struct {
	Text        bool
	Password    bool
	Message     bool
	CommandText bool
	Num         bool
	Icon        bool
	Sound       bool
	File        bool
	Key         bool
	Color       bool
	Null        bool

	Area     bool
	Turf     bool
	Obj      bool
	Mob      bool
	Anything bool
}

type ProcFlags struct {
	Name         string
	Desc         string
	Category     string
	Hidden       bool
	PopupMenu    bool
	Instant      bool
	Invisibility int
	Access       ProcAccess
	Range        int
	Background   bool
	WaitFor      bool
}

type ProcArg struct {
	Name        string
	InputSchema ArgInputSchema
	InListExpr  *astpb.Expression
}

type ProcArgs struct {
	args map[string]*ProcArg
}

func NewProcArgs() ProcArgs {
	pa := ProcArgs{}
	pa.args = make(map[string]*ProcArg)
	return pa
}

func (pa ProcArgs) Add(name string, inputSchema ArgInputSchema, inListExpr *astpb.Expression) {
	newPa := &ProcArg{Name: name, InputSchema: inputSchema, InListExpr: inListExpr}
	pa.args[name] = newPa
}

func MakeDefaultProcFlags() ProcFlags {
	flags := ProcFlags{
		PopupMenu: true,
		WaitFor:   true,
	}
	return flags
}

type DMProc struct {
	State *Parser
	Type  *DMType
	Name  string
	Proto *astpb.TypeProc
	Flags ProcFlags
	Args  ProcArgs
}

func NewProc(s *Parser, t *DMType, n string, pb *astpb.TypeProc) *DMProc {
	p := &DMProc{State: s, Type: t, Name: n, Proto: pb, Flags: MakeDefaultProcFlags()}
	p.Args = NewProcArgs()
	selectedProcDef := p.Proto.Value[len(p.Proto.Value)-1]
	for _, prm := range selectedProcDef.GetParameter() {
		name := prm.GetName()
		schema := ArgInputSchema{}
		if prm.GetInputType() != nil {
			for _, kn := range prm.GetInputType().GetKeyName() {
				if kn == astpb.InputTypeKey_INPUT_TYPE_TEXT {
					schema.Text = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_PASSWORD {
					schema.Password = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_MESSAGE {
					schema.Message = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_COMMAND_TEXT {
					schema.CommandText = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_NUM {
					schema.Num = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_ICON {
					schema.Icon = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_SOUND {
					schema.Sound = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_FILE {
					schema.File = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_KEY {
					schema.Key = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_COLOR {
					schema.Color = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_NULL {
					schema.Null = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_AREA {
					schema.Area = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_TURF {
					schema.Turf = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_OBJ {
					schema.Obj = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_MOB {
					schema.Mob = true
				} else if kn == astpb.InputTypeKey_INPUT_TYPE_ANYTHING {
					schema.Anything = true
				}
			}
		}
		inListExpr := prm.GetInList()
		p.Args.Add(name, schema, inListExpr)
	}
	return p
}

func (p *DMProc) String() string {
	return fmt.Sprintf("Proc<%v%v>", p.Type.Path.FullyQualifiedString(), p.Name)
}

func (p *DMProc) ArgNames() []string {
	var argNames []string
	for _, procVal := range p.Proto.GetValue() {
		for _, param := range procVal.GetParameter() {
			argNames = append(argNames, *param.Name)
		}
	}
	return argNames
}

func (p *DMProc) HasArg(s string) bool {
	for _, n := range p.ArgNames() {
		if s == n {
			return true
		}
	}
	return false
}

// EmitName deals with symbol collisions with C++, or collisions between
// namespaces (such as donk::icon) with core procs. This only affects
// source-level symbols; procs and vars are registered with their original names
// in the C++ iotas.
func (p *DMProc) EmitName() string {
	if p.Name == "new" {
		return "new_"
	}
	if p.ProcPath().FullyQualifiedString() == "/icon" {
		return "icon_"
	}
	if p.ProcPath().FullyQualifiedString() == "/matrix" {
		return "matrix_"
	}
	if p.ProcPath().FullyQualifiedString() == "/sound" {
		return "sound_"
	}
	if p.ProcPath().FullyQualifiedString() == "/regex" {
		return "regex_"
	}
	if p.ProcPath().FullyQualifiedString() == "/image" {
		return "image_"
	}
	if p.ProcPath().FullyQualifiedString() == "/list" {
		return "list_"
	}
	return p.Name
}

func (p *DMProc) PrettyProto() string {
	return proto.MarshalTextString(p.Proto)
}

func (d *DMProc) Block(idx int) *astpb.Block {
	return d.Proto.Value[idx].Code.Present
}

func (d *DMProc) ProcPath() *paths.Path {
	return d.Type.Path.Child(d.Name)
}
