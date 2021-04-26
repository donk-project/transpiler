package writer

// Dealing with Runfiles and Starlark sucks shit so all these templates are just
// going straight into the source files.
const (
	TRANSFORMER_DECLARATION_H = `// Generated by Donk Transpiler. Changes may be overwritten.
// Template:    transformer_declaration.h.tmpl
// Filename:    {{.GetFileMetadata.GetFilename}}

#ifndef {{.GetIncludeGuard}}
#define {{.GetIncludeGuard}}
{{range $hdr := .GetPreamble.GetHeaders}}
#include {{$hdr}}
{{- end}}

{{range $ns_decl := .GetNamespacedDeclarations -}}
namespace {{$ns_decl.GetNamespace}} {
{{range $cd := $ns_decl.GetClassDeclarations}}
class {{$cd.GetName}}
{{- if $cd.GetBaseSpecifiers }} : {{printBaseSpecifiers $cd.GetBaseSpecifiers -}}{{- end}} {
{{range $ms := $cd.GetMemberSpecifiers}}
{{printMemberSpecifier $ms}}
{{- end}}
};
{{- end}}

{{range $fd := $ns_decl.GetFunctionDeclarations}}
{{printFunctionDeclaration $fd}};
{{- end}}

}  // namespace {{$ns_decl.GetNamespace}}
{{- end}}

#endif  // {{.GetIncludeGuard}}
`

	TRANSFORMER_DEFINITION_CC = `// Generated by Donk Transpiler. Changes may be overwritten.
// Template:    transformer_definition_cc.tmpl
// Filename:    {{.GetFileMetadata.GetFilename}}
#include {{.GetBaseInclude}}

{{range $hdr := .GetPreamble.GetHeaders -}}
#include {{$hdr}}
{{end -}}

{{range $ns_defn := .GetNamespacedDefinitions }}
namespace {{$ns_defn.GetNamespace}} {

{{- range $c := $ns_defn.GetConstructors -}}
{{printConstructor $c}}
{{- end -}}

{{- range $fd := $ns_defn.GetFunctionDefinitions}}

{{printFunctionDeclaration $fd.GetDeclaration}} {
{{range $stmt := $fd.GetBlockDefinition.GetStatements}}  {{printStatement $stmt}}
{{end -}}
}
{{- end}}

}  // namespace {{$ns_defn.GetNamespace}}
{{- end}}

`

	TYPE_REGISTRAR_H = `#ifndef __SNOWFROST__{{$.CoreNamespace}}_TYPE_REGISTRAR_H__
#define __SNOWFROST__{{$.CoreNamespace}}_TYPE_REGISTRAR_H__

#include "donk/core/iota.h"
#include "donk/core/vars.h"
#include "donk/core/procs.h"

#include <functional>
#include <map>
#include <vector>
#include <string>

namespace {{$.CoreNamespace}} {

void RegisterAll(std::shared_ptr<std::map<donk::path_t, std::vector<std::function<void(donk::iota_t&)>>>> collector);

} // namespace {{$.CoreNamespace}}

#endif //  __SNOWFROST__{{$.CoreNamespace}}_TYPE_REGISTRAR_H__

// END - Generated from type_registrar.h.tmpl
`

	TYPE_REGISTRAR_CC = `// START - Generated from type_registrar.cc.tmpl
#include "{{$.TypeRegistrarInclude}}"

#include <map>
#include <string>
#include <vector>
#include <functional>

#include "spdlog/spdlog.h"

#include "donk/core/iota.h"
#include "donk/core/vars.h"
#include "donk/core/procs.h"

{{- range $h := $.TypeIncludes}}
#include "{{$h}}"
{{- end }}

namespace {{$.CoreNamespace}} {

void RegisterAll(std::shared_ptr<std::map<donk::path_t, std::vector<std::function<void(donk::iota_t&)>>>> collector) {
  {{- range $FullPath, $RegisterNs := $.Registrations }}
    (*collector)[donk::path_t("{{$FullPath}}")].push_back({{$RegisterNs}}::Register);
  {{- end }}
}

} // namespace {{$.CoreNamespace}}
// END - Generated from type_registrar.cc.tmpl

`
)
