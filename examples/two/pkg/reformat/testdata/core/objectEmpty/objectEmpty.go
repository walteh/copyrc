// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/core/objectEmpty/objectEmpty.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

type ObjectEmpty struct {
	// Foo corresponds to the JSON schema field "foo".
	Foo ObjectEmptyFoo `json:"foo,omitempty" yaml:"foo,omitempty" mapstructure:"foo,omitempty"`
}

type ObjectEmptyFoo map[string]interface{}
