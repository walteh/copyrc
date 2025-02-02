// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/core/time/time.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

import "encoding/json"
import "fmt"
import "github.com/atombender/go-jsonschema/pkg/types"

type Time struct {
	// MyObject corresponds to the JSON schema field "myObject".
	MyObject *TimeMyObject `json:"myObject,omitempty" yaml:"myObject,omitempty" mapstructure:"myObject,omitempty"`
}

type TimeMyObject struct {
	// MyTime corresponds to the JSON schema field "myTime".
	MyTime types.SerializableTime `json:"myTime" yaml:"myTime" mapstructure:"myTime"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *TimeMyObject) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["myTime"]; raw != nil && !ok {
		return fmt.Errorf("field myTime in TimeMyObject: required")
	}
	type Plain TimeMyObject
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = TimeMyObject(plain)
	return nil
}
