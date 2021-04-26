// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package paths

import (
	"fmt"
	"strings"

	astpb "snowfrost.garden/donk/proto/ast"
)

// NAMESPACE_RENAMES maps namespaces that might collide with C++ syntax
// with a more compatible version.
var NAMESPACE_RENAMES = map[string]string{
	"export": "export_",
}

type Path struct {
	Name     string
	Basename string
}

func (p Path) String() string {
	return fmt.Sprintf("`%v`", p.FullyQualifiedString())
}

func NewFromTreePath(tp *astpb.TreePath) Path {
	var s []string
	for _, x := range tp.S {
		s = append(s, x)
	}
	return New("/" + strings.Join(s, "/"))
}

func NewFromTypePaths(tps []*astpb.TypePath) Path {
	var s []string
	for _, tp := range tps {
		s = append(s, tp.GetS())
	}
	return New("/" + strings.Join(s, "/"))
}

func New(s string) Path {
	if s == "" {
		panic("Empty path")
	}
	var p Path
	p.Name = s
	if strings.HasPrefix(p.Name, "/area") {
		p.Name = "/datum/atom" + p.Name
	} else if strings.HasPrefix(p.Name, "/atom") {
		p.Name = "/datum" + p.Name
	} else if strings.HasPrefix(p.Name, "/mob") {
		p.Name = "/datum/atom/movable" + p.Name
	} else if strings.HasPrefix(p.Name, "/turf") {
		p.Name = "/datum/atom" + p.Name
	} else if strings.HasPrefix(p.Name, "/obj") {
		p.Name = "/datum/atom/movable" + p.Name
	}

	parts := strings.Split(p.Name, "/")
	p.Basename = parts[len(parts)-1]

	return p
}

func JoinIntoPath(c []string) Path {
	name := strings.Join(c, "/")
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	return New(name)
}

func (p Path) IsRoot() bool {
	return p.Name == "/"
}

func (p Path) Equals(s string) bool {
	return p.Name == s
}

func (p Path) FullyQualifiedString() string {
	return p.Name
}

func (p Path) ParentPath() Path {
	if p.Name == "/area" {
		return New("/atom")
	}
	if p.Name == "/atom" {
		return New("/datum")
	}
	if p.Name == "/mob" {
		return New("/atom/movable")
	}
	if p.Name == "/turf" {
		return New("/atom")
	}
	if p.Name == "/obj" {
		return New("/atom/movable")
	}
	split := strings.Split(p.Name, "/")

	return JoinIntoPath(split[:len(split)-1])
}

func (p Path) Child(child string) Path {
	return JoinIntoPath([]string{strings.TrimRight(p.Name, "/"), strings.TrimLeft(child, "/")})
}

func (p Path) Parts() []string {
	result := strings.Split(strings.TrimPrefix(p.FullyQualifiedString(), "/"), "/")
	return result
}

func (p Path) AsNamespace() string {
	var result []string
	if !p.IsRoot() {
		for _, part := range p.Parts() {
			replacement, ok := NAMESPACE_RENAMES[part]
			if ok {
				result = append(result, replacement)
			} else {
				result = append(result, part)
			}
		}
	}

	return strings.Join(result, "::")
}

func (p Path) IsCoretype() bool {
	switch p.FullyQualifiedString() {
	case "/datum/atom/movable/mob":
		return true
	case "/image":
		return true
	case "/database":
		return true
	case "/regex":
		return true
	case "/icon":
		return true
	case "/sound":
		return true
	case "/datum":
		return true
	case "/":
		return true
	case "/mutable_appearance":
		return true
	case "/exception":
		return true
	case "/datum/atom":
		return true
	case "/datum/atom/movable":
		return true
	case "/database/query":
		return true
	case "/matrix":
		return true
	case "/client":
		return true
	case "/list":
		return true
	case "/savefile":
		return true
	case "/world":
		return true
	case "/datum/atom/movable/obj":
		return true
	case "/datum/atom/turf":
		return true
	case "/datum/atom/area":
		return true
	}
	return false
}
