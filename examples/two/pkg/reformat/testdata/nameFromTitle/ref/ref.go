// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/nameFromTitle/ref/ref.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

type ExtRef struct {
	// MyThing corresponds to the JSON schema field "myThing".
	MyThing *Thing `json:"myThing,omitempty" yaml:"myThing,omitempty" mapstructure:"myThing,omitempty"`

	// MyThing2 corresponds to the JSON schema field "myThing2".
	MyThing2 *Thing `json:"myThing2,omitempty" yaml:"myThing2,omitempty" mapstructure:"myThing2,omitempty"`
}

type Thing struct {
	// Name corresponds to the JSON schema field "name".
	Name *string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`
}
