// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/core/date/date.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

import "encoding/json"
import "fmt"
import "github.com/atombender/go-jsonschema/pkg/types"

type Date struct {
	// MyObject corresponds to the JSON schema field "myObject".
	MyObject *DateMyObject `json:"myObject,omitempty" yaml:"myObject,omitempty" mapstructure:"myObject,omitempty"`
}

type DateMyObject struct {
	// MyDate corresponds to the JSON schema field "myDate".
	MyDate types.SerializableDate `json:"myDate" yaml:"myDate" mapstructure:"myDate"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *DateMyObject) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["myDate"]; raw != nil && !ok {
		return fmt.Errorf("field myDate in DateMyObject: required")
	}
	type Plain DateMyObject
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = DateMyObject(plain)
	return nil
}
