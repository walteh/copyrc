// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/core/refToEnum/refToEnum.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

import "encoding/json"
import "fmt"
import "reflect"

type RefToEnum struct {
	// MyThing corresponds to the JSON schema field "myThing".
	MyThing *Thing `json:"myThing,omitempty" yaml:"myThing,omitempty" mapstructure:"myThing,omitempty"`
}

type Thing string

const ThingX Thing = "x"
const ThingY Thing = "y"

var enumValues_Thing = []interface{}{
	"x",
	"y",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Thing) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_Thing {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_Thing, v)
	}
	*j = Thing(v)
	return nil
}
