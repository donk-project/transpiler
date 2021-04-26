// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package main

import (
	"snowfrost.garden/donk/transpiler"
)

func main() {
	o := transpiler.OptsFromFlags()
	t := transpiler.New(o)
	t.Transpile()
}
