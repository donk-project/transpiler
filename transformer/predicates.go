// Donk Project
// Copyright (c) 2021 Warriorstar Orion <orion@snowfrost.garden>
// SPDX-License-Identifier: MIT
package transformer

import (
	"snowfrost.garden/donk/transpiler/parser"
)

func (t *Transformer) shouldEmitProc(p *parser.DMProc) bool {
	if t.isCoreGen() {
		return true
	}
	// more than one value means a subclass version of the same proc
	if len(p.Proto.Value) > 1 {
		return true
	}
	if p.Proto.GetDeclaration().GetLocation().GetFile() != nil {
		return *p.Proto.GetDeclaration().GetLocation().GetFile().FileId != 0
	} else if p.Proto.GetValue() != nil {
		for _, value := range p.Proto.GetValue() {
			if value.GetCode().GetPresent() != nil {
				return true
			}
		}
	}
	return false
}

func (t *Transformer) shouldEmitVar(v *parser.DMVar) bool {
	if t.isCoreGen() {
		return true
	}
	if v.Proto.GetValue() != nil {
		if v.Proto.GetValue().GetLocation().GetFile().GetFileId() == 0 {
			return false
		}
		return true
	}

	return v.Proto.GetDeclaration().GetLocation().GetFile().GetFileId() != 0
}

func (t *Transformer) HasEmittableVars(typ *parser.DMType) bool {
	if len(typ.Vars) > 0 && t.isCoreGen() {
		return true
	}

	for _, v := range typ.Vars {
		if t.shouldEmitVar(v) {
			return true
		}
	}

	return false
}

func (t *Transformer) HasEmittableProcs(typ *parser.DMType) bool {
	if len(typ.Procs) > 0 && t.isCoreGen() {
		return true
	}

	for _, p := range typ.Procs {
		if t.shouldEmitProc(p) {
			return true
		}
	}

	return false
}

func (t *Transformer) IsProcInCore(name string) bool {
	// TODO: Tie this to known dreamchecker builtins
	// This was originally read from the core binarypb but building out the
	// Starlark rules was such a pain due to runfiles resolution it was taken
	// out of the critical path for transpilation
	if name == "abs" {
		return true
	}
	if name == "addtext" {
		return true
	}
	if name == "alert" {
		return true
	}
	if name == "animate" {
		return true
	}
	if name == "arccos" {
		return true
	}
	if name == "arcsin" {
		return true
	}
	if name == "arglist" {
		return true
	}
	if name == "ascii2text" {
		return true
	}
	if name == "block" {
		return true
	}
	if name == "bounds" {
		return true
	}
	if name == "bounds_dist" {
		return true
	}
	if name == "browse" {
		return true
	}
	if name == "browse_rsc" {
		return true
	}
	if name == "ckey" {
		return true
	}
	if name == "ckeyEx" {
		return true
	}
	if name == "cmptext" {
		return true
	}
	if name == "cmptextEx" {
		return true
	}
	if name == "copytext" {
		return true
	}
	if name == "cos" {
		return true
	}
	if name == "fcopy" {
		return true
	}
	if name == "fcopy_rsc" {
		return true
	}
	if name == "fdel" {
		return true
	}
	if name == "fexists" {
		return true
	}
	if name == "file" {
		return true
	}
	if name == "file2text" {
		return true
	}
	if name == "filter" {
		return true
	}
	if name == "findlasttext" {
		return true
	}
	if name == "findlasttextEx" {
		return true
	}
	if name == "findtext" {
		return true
	}
	if name == "findtextEx" {
		return true
	}
	if name == "flick" {
		return true
	}
	if name == "flist" {
		return true
	}
	if name == "ftp" {
		return true
	}
	if name == "get_dir" {
		return true
	}
	if name == "get_dist" {
		return true
	}
	if name == "get_step" {
		return true
	}
	if name == "get_step_away" {
		return true
	}
	if name == "get_step_rand" {
		return true
	}
	if name == "get_step_to" {
		return true
	}
	if name == "get_step_towards" {
		return true
	}
	if name == "gradient" {
		return true
	}
	if name == "hascall" {
		return true
	}
	if name == "hearers" {
		return true
	}
	if name == "html_decode" {
		return true
	}
	if name == "html_encode" {
		return true
	}
	if name == "icon" {
		return true
	}
	if name == "icon_states" {
		return true
	}
	if name == "image" {
		return true
	}
	if name == "initial" {
		return true
	}
	if name == "input" {
		return true
	}
	if name == "isarea" {
		return true
	}
	if name == "isfile" {
		return true
	}
	if name == "isicon" {
		return true
	}
	if name == "isloc" {
		return true
	}
	if name == "ismob" {
		return true
	}
	if name == "isnull" {
		return true
	}
	if name == "isnum" {
		return true
	}
	if name == "isobj" {
		return true
	}
	if name == "ispath" {
		return true
	}
	if name == "issaved" {
		return true
	}
	if name == "istext" {
		return true
	}
	if name == "isturf" {
		return true
	}
	if name == "istype" {
		return true
	}
	if name == "jointext" {
		return true
	}
	if name == "json_decode" {
		return true
	}
	if name == "json_encode" {
		return true
	}
	if name == "length" {
		return true
	}
	if name == "lentext" {
		return true
	}
	if name == "link" {
		return true
	}
	if name == "list" {
		return true
	}
	if name == "list2params" {
		return true
	}
	if name == "load_resource" {
		return true
	}
	if name == "locate" {
		return true
	}
	if name == "log" {
		return true
	}
	if name == "lowertext" {
		return true
	}
	if name == "matrix" {
		return true
	}
	if name == "max" {
		return true
	}
	if name == "md5" {
		return true
	}
	if name == "min" {
		return true
	}
	if name == "missile" {
		return true
	}
	if name == "newlist" {
		return true
	}
	if name == "nonspantext" {
		return true
	}
	if name == "num2text" {
		return true
	}
	if name == "obounds" {
		return true
	}
	if name == "ohearers" {
		return true
	}
	if name == "orange" {
		return true
	}
	if name == "output" {
		return true
	}
	if name == "oview" {
		return true
	}
	if name == "oviewers" {
		return true
	}
	if name == "params2list" {
		return true
	}
	if name == "pick" {
		return true
	}
	if name == "prob" {
		return true
	}
	if name == "rand" {
		return true
	}
	if name == "rand_seed" {
		return true
	}
	if name == "range" {
		return true
	}
	if name == "regex" {
		return true
	}
	if name == "REGEX_QUOTE" {
		return true
	}
	if name == "REGEX_QUOTE_REPLACEMENT" {
		return true
	}
	if name == "replacetext" {
		return true
	}
	if name == "replacetextEx" {
		return true
	}
	if name == "rgb" {
		return true
	}
	if name == "rgb2num" {
		return true
	}
	if name == "roll" {
		return true
	}
	if name == "round" {
		return true
	}
	if name == "run" {
		return true
	}
	if name == "shell" {
		return true
	}
	if name == "shutdown" {
		return true
	}
	if name == "sin" {
		return true
	}
	if name == "sleep" {
		return true
	}
	if name == "sorttext" {
		return true
	}
	if name == "sorttextEx" {
		return true
	}
	if name == "sound" {
		return true
	}
	if name == "spantext" {
		return true
	}
	if name == "splicetext" {
		return true
	}
	if name == "splittext" {
		return true
	}
	if name == "sqrt" {
		return true
	}
	if name == "startup" {
		return true
	}
	if name == "stat" {
		return true
	}
	if name == "statpanel" {
		return true
	}
	if name == "step" {
		return true
	}
	if name == "step_away" {
		return true
	}
	if name == "step_rand" {
		return true
	}
	if name == "step_to" {
		return true
	}
	if name == "step_towards" {
		return true
	}
	if name == "text" {
		return true
	}
	if name == "text2ascii" {
		return true
	}
	if name == "text2file" {
		return true
	}
	if name == "text2num" {
		return true
	}
	if name == "text2path" {
		return true
	}
	if name == "time2text" {
		return true
	}
	if name == "turn" {
		return true
	}
	if name == "typesof" {
		return true
	}
	if name == "uppertext" {
		return true
	}
	if name == "url_decode" {
		return true
	}
	if name == "url_encode" {
		return true
	}
	if name == "view" {
		return true
	}
	if name == "viewers" {
		return true
	}
	if name == "walk" {
		return true
	}
	if name == "walk_away" {
		return true
	}
	if name == "walk_rand" {
		return true
	}
	if name == "walk_to" {
		return true
	}
	if name == "walk_towards" {
		return true
	}
	if name == "winclone" {
		return true
	}
	if name == "winexists" {
		return true
	}
	if name == "winget" {
		return true
	}
	if name == "winset" {
		return true
	}
	if name == "winshow" {
		return true
	}
	if name == "CRASH" {
		return true
	}
	if name == "_dm_db_new_query" {
		return true
	}
	if name == "_dm_db_execute" {
		return true
	}
	if name == "_dm_db_next_row" {
		return true
	}
	if name == "_dm_db_rows_affected" {
		return true
	}
	if name == "_dm_db_row_count" {
		return true
	}
	if name == "_dm_db_error_msg" {
		return true
	}
	if name == "_dm_db_columns" {
		return true
	}
	if name == "_dm_db_close" {
		return true
	}
	if name == "_dm_db_new_con" {
		return true
	}
	if name == "_dm_db_connect" {
		return true
	}
	if name == "_dm_db_quote" {
		return true
	}
	if name == "_dm_db_is_connected" {
		return true
	}

	return false
}
