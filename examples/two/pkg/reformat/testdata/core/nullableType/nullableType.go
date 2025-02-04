// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/core/nullableType/nullableType.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

type BoolThing *bool

type FloatThing *float64

type IntegerThing *int

type NullableType struct {
	// MyInlineStringValue corresponds to the JSON schema field "MyInlineStringValue".
	MyInlineStringValue *string `json:"MyInlineStringValue,omitempty" yaml:"MyInlineStringValue,omitempty" mapstructure:"MyInlineStringValue,omitempty"`

	// MyStringValue corresponds to the JSON schema field "MyStringValue".
	MyStringValue StringThing `json:"MyStringValue,omitempty" yaml:"MyStringValue,omitempty" mapstructure:"MyStringValue,omitempty"`
}

type StringThing *string
