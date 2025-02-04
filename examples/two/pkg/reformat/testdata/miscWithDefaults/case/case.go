// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/miscWithDefaults/case/case.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

type Case struct {
	// CapitalCamelField corresponds to the JSON schema field "CapitalCamelField".
	CapitalCamelField *string `json:"CapitalCamelField,omitempty" yaml:"CapitalCamelField,omitempty" mapstructure:"CapitalCamelField,omitempty"`

	// UPPERCASEFIELD corresponds to the JSON schema field "UPPERCASEFIELD".
	UPPERCASEFIELD *string `json:"UPPERCASEFIELD,omitempty" yaml:"UPPERCASEFIELD,omitempty" mapstructure:"UPPERCASEFIELD,omitempty"`

	// CamelCase corresponds to the JSON schema field "camelCase".
	CamelCase *string `json:"camelCase,omitempty" yaml:"camelCase,omitempty" mapstructure:"camelCase,omitempty"`

	// Lowercase corresponds to the JSON schema field "lowercase".
	Lowercase *string `json:"lowercase,omitempty" yaml:"lowercase,omitempty" mapstructure:"lowercase,omitempty"`

	// SnakeMixedCase corresponds to the JSON schema field "snake_Mixed_Case".
	SnakeMixedCase *string `json:"snake_Mixed_Case,omitempty" yaml:"snake_Mixed_Case,omitempty" mapstructure:"snake_Mixed_Case,omitempty"`

	// SnakeCase corresponds to the JSON schema field "snake_case".
	SnakeCase *string `json:"snake_case,omitempty" yaml:"snake_case,omitempty" mapstructure:"snake_case,omitempty"`
}
