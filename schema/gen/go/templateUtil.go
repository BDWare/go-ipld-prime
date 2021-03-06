package gengo

import (
	"io"
	"strings"
	"text/template"

	ipld "github.com/ipld/go-ipld-prime"

	wish "github.com/warpfork/go-wish"
)

func doTemplate(tmplstr string, w io.Writer, data interface{}) {
	tmpl := template.Must(template.New("").
		Funcs(template.FuncMap{
			// 'ReprKindConst' returns the source-string for "ipld.ReprKind_{{Kind}}".
			"ReprKindConst": func(k ipld.ReprKind) string {
				return "ipld.ReprKind_" + k.String() // happens to be fairly trivial.
			},

			// 'Add' does what it says on the tin.
			"Add": func(a, b int) int {
				return a + b
			},

			// Title case.  Used to make a exported symbol.  Could be more efficient.
			"titlize": strings.Title,

			"mungeTypeNodeIdent":                mungeTypeNodeIdent,
			"mungeTypeNodeItrIdent":             mungeTypeNodeItrIdent,
			"mungeTypeNodebuilderIdent":         mungeTypeNodebuilderIdent,
			"mungeTypeNodeMapBuilderIdent":      mungeTypeNodeMapBuilderIdent,
			"mungeTypeNodeListBuilderIdent":     mungeTypeNodeListBuilderIdent,
			"mungeTypeReprNodeIdent":            mungeTypeReprNodeIdent,
			"mungeTypeReprNodeItrIdent":         mungeTypeReprNodeItrIdent,
			"mungeTypeReprNodebuilderIdent":     mungeTypeReprNodebuilderIdent,
			"mungeTypeReprNodeMapBuilderIdent":  mungeTypeReprNodeMapBuilderIdent,
			"mungeTypeReprNodeListBuilderIdent": mungeTypeReprNodeListBuilderIdent,

			"mungeNodebuilderConstructorIdent":     mungeNodebuilderConstructorIdent,
			"mungeReprNodebuilderConstructorIdent": mungeReprNodebuilderConstructorIdent,
		}).
		Parse(wish.Dedent(tmplstr)))
	if err := tmpl.Execute(w, data); err != nil {
		panic(err)
	}
}
