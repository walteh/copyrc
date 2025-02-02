// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/pkg/generator/formatter.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

package reformat

import (
	"github.com/atombender/go-jsonschema/pkg/codegen"
)

type formatter interface {
	addImport(out *codegen.File)

	generate(declType codegen.TypeDecl, validators []validator) func(*codegen.Emitter)
	enumMarshal(declType codegen.TypeDecl) func(*codegen.Emitter)
	enumUnmarshal(
		declType codegen.TypeDecl,
		enumType codegen.Type,
		valueConstant *codegen.Var,
		wrapInStruct bool,
	) func(*codegen.Emitter)
}
