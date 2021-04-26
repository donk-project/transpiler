package transformer

import (
	"log"
	"testing"

	"google.golang.org/protobuf/encoding/prototext"
	astpb "snowfrost.garden/donk/proto/ast"
	"snowfrost.garden/donk/transpiler/parser"
	"snowfrost.garden/donk/transpiler/paths"
	// "github.com/golang/protobuf/proto"
)

func TestTransformerCoreSubclass(t *testing.T) {
	astpbProtoBytes := []byte(`type: <
  name: "mob"
  path: "/mob"
  location: <
    file: <
      file_id: 2
    >
    line: 22
    column: 10
  >
  location_specificity: 2
	procs: <
	key: "Login"
	value: <
		value: <
			location: <
				file: <
					file_id: 0
				>
				line: 1
				column: 1
			>
			docs: <
			>
			code: <
				builtin: true
			>
		>
		value: <
			location: <
				file: <
					file_id: 2
				>
				line: 1
				column: 10
			>
			docs: <
			>
			code: <
				present: <
					statement: <
						expr: <
							base: <
								term: <
									parent_call: <
									>
								>
							>
						>
					>
					statement: <
						expr: <
							binary_op: <
								op: BINARYOP_LSHIFT
								lhs: <
									base: <
										term: <
											ident: "world"
										>
									>
								>
								rhs: <
									base: <
										term: <
											string_t: "Hello world!"
										>
									>
								>
							>
						>
					>
				>
			>
		>
		declaration: <
			location: <
				file: <
					file_id: 0
				>
				line: 1
				column: 1
			>
			kind: PROC_DECL_KIND_PROC
			id: <
				symbol_id: 364
			>
			is_private: false
			is_protected: false
		>
	>
>
>`)

	g := &astpb.Graph{}
	if err := prototext.Unmarshal(astpbProtoBytes, g); err != nil {
		log.Fatalln("Failed to parse: ", err)
	}
	p := parser.NewParser(g)
	tr := New(p, "test", "test")
	tr.BeginTransform()

	mob := p.TypesByPath[paths.New("/mob")]
	loginProc := mob.Proc("Login")
	ns := tr.supertypeNamespace(loginProc)
	want := "donk::api::datum::atom::movable::mob"
	if ns != want {
		t.Fatalf("supertypeParser() got %v wanted %v", ns, want)
	}

	include := tr.supertypeInclude(loginProc)
	want = "donk/api/datum/atom/movable/mob.h"
	if include != want {
		t.Fatalf("supertypeInclude() got %v wanted %v", include, want)
	}

}
