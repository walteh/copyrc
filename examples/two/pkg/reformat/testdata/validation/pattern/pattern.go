// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/validation/pattern/pattern.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

import "encoding/json"
import "fmt"
import "regexp"

type Pattern struct {
	// MyNullableString corresponds to the JSON schema field "myNullableString".
	MyNullableString *string `json:"myNullableString,omitempty" yaml:"myNullableString,omitempty" mapstructure:"myNullableString,omitempty"`

	// MyString corresponds to the JSON schema field "myString".
	MyString string `json:"myString" yaml:"myString" mapstructure:"myString"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Pattern) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["myString"]; raw != nil && !ok {
		return fmt.Errorf("field myString in Pattern: required")
	}
	type Plain Pattern
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if plain.MyNullableString != nil {
		if matched, _ := regexp.MatchString(`^0x[0-9a-f]{10}$`, string(*plain.MyNullableString)); !matched {
			return fmt.Errorf("field %s pattern match: must match %s", "MyNullableString", `^0x[0-9a-f]{10}$`)
		}
	}
	if matched, _ := regexp.MatchString(`^0x[0-9a-f]{10}\.$`, string(plain.MyString)); !matched {
		return fmt.Errorf("field %s pattern match: must match %s", "MyString", `^0x[0-9a-f]{10}\.$`)
	}
	*j = Pattern(plain)
	return nil
}
