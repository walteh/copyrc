// 📦 originally copied by copyrc
// 🔗 source: https://raw.githubusercontent.com/omissis/go-jsonschema/442a4c100c62a7d8543d1a7ab7052397057add86/tests/data/minSizedInts/exactReferences/exact.go
// 📝 license: MIT
// ℹ️ see .copyrc.lock for more details

// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package testdata

import "encoding/json"
import "fmt"

type Bound16 int16

type Bound32 int32

type Bound64 int64

type Bound8 int8

type Exact struct {
	// I16 corresponds to the JSON schema field "i16".
	I16 Bound16 `json:"i16" yaml:"i16" mapstructure:"i16"`

	// I32 corresponds to the JSON schema field "i32".
	I32 Bound32 `json:"i32" yaml:"i32" mapstructure:"i32"`

	// I64 corresponds to the JSON schema field "i64".
	I64 Bound64 `json:"i64" yaml:"i64" mapstructure:"i64"`

	// I8 corresponds to the JSON schema field "i8".
	I8 Bound8 `json:"i8" yaml:"i8" mapstructure:"i8"`

	// U16 corresponds to the JSON schema field "u16".
	U16 UBound16 `json:"u16" yaml:"u16" mapstructure:"u16"`

	// U32 corresponds to the JSON schema field "u32".
	U32 UBound32 `json:"u32" yaml:"u32" mapstructure:"u32"`

	// U64 corresponds to the JSON schema field "u64".
	U64 UBound64 `json:"u64" yaml:"u64" mapstructure:"u64"`

	// U8 corresponds to the JSON schema field "u8".
	U8 UBound8 `json:"u8" yaml:"u8" mapstructure:"u8"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Exact) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["i16"]; raw != nil && !ok {
		return fmt.Errorf("field i16 in Exact: required")
	}
	if _, ok := raw["i32"]; raw != nil && !ok {
		return fmt.Errorf("field i32 in Exact: required")
	}
	if _, ok := raw["i64"]; raw != nil && !ok {
		return fmt.Errorf("field i64 in Exact: required")
	}
	if _, ok := raw["i8"]; raw != nil && !ok {
		return fmt.Errorf("field i8 in Exact: required")
	}
	if _, ok := raw["u16"]; raw != nil && !ok {
		return fmt.Errorf("field u16 in Exact: required")
	}
	if _, ok := raw["u32"]; raw != nil && !ok {
		return fmt.Errorf("field u32 in Exact: required")
	}
	if _, ok := raw["u64"]; raw != nil && !ok {
		return fmt.Errorf("field u64 in Exact: required")
	}
	if _, ok := raw["u8"]; raw != nil && !ok {
		return fmt.Errorf("field u8 in Exact: required")
	}
	type Plain Exact
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Exact(plain)
	return nil
}

type UBound16 uint16

type UBound32 uint32

type UBound64 uint64

type UBound8 uint8
