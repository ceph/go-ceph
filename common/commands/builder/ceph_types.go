//go:build !(pacific || quincy) && ceph_preview

package builder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// CephArgumentType represents types that can be used to manage argument values
// for a Ceph command.
type CephArgumentType interface {
	TypeName() string
	Name() string
	Set(map[string]any, any) error
	Validate(map[string]any) error
}

// CephScalarArgumentType represents types that can be used to manage a single
// argument value for a Ceph command.
type CephScalarArgumentType interface {
	CephArgumentType
	Convert(v any) (any, error)
	Check(v any) error
}

// CephMultiArgumentType represents types that can be used to manage multiple
// (slices of) values for a single argument in a Ceph command.
type CephMultiArgumentType interface {
	Append(map[string]any, any) error
}

// The CephType... values are gleaned from the ceph code
// NOT all of these types are implemented in go-ceph.

// CephTypeX constants naming all currently known variable argument types.
const (
	CephTypeBool       = "CephBool"
	CephTypeChoices    = "CephChoices"
	CephTypeEntityAddr = "CephEntityAddr"
	CephTypeFilePath   = "CephFilepath"
	CephTypeFloat      = "CephFloat"
	CephTypeFragment   = "CephFragment"
	CephTypeInt        = "CephInt"
	CephTypeIPAddr     = "CephIPAddr"
	CephTypeName       = "CephName"
	CephTypeObjectName = "CephObjectname"
	CephTypeOSDName    = "CephOsdName"
	CephTypePGID       = "CephPgid"
	CephTypePoolName   = "CephPoolname"
	CephTypeSocketPath = "CephSocketpath"
	CephTypeString     = "CephString"
	CephTypeUUID       = "CephUUID"
)

/* Type: Ceph Choices */

// CephChoices arguments are basically strings constrained to certain
// allowed values.
type CephChoices struct {
	sv *SignatureVar
}

// TypeName returns the name of this ceph argument type.
func (*CephChoices) TypeName() string { return CephTypeChoices }

// Name returns the name of this ceph argument.
func (t *CephChoices) Name() string { return t.sv.Name }

// Choices returns the allowed values for this argument as map of strings to
// bools.
func (t *CephChoices) Choices() map[string]bool {
	m := map[string]bool{}
	for _, ch := range strings.Split(t.sv.Choices, "|") {
		m[ch] = true
	}
	return m
}

func (t *CephChoices) choose(s string) (string, error) {
	if !t.Choices()[s] {
		return "", fmt.Errorf("invalid choice: %s", s)
	}
	return s, nil
}

// Convert an any value into a valid underlying type for later serialization.
// Returns new type as any or error if conversion fails.
func (t *CephChoices) Convert(v any) (any, error) {
	switch vs := v.(type) {
	case string:
		return t.choose(vs)
	case fmt.Stringer:
		return t.choose(vs.String())
	}
	return "", fmt.Errorf("not a string: %v", v)
}

// Check that a given value meets requirements for this argument.
// Returns an error if this value fails the check.
func (t *CephChoices) Check(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("not a string: %v (at %s)", v, t.sv.Name)
	}
	if !t.Choices()[s] {
		return fmt.Errorf("invalid choice: %s (at %s)", s, t.sv.Name)
	}
	return nil
}

// Set the given value into the map ensuring that the value is of the correct
// underlying type. If the type is not valid returns an error.
func (t *CephChoices) Set(data map[string]any, v any) error {
	x, e := t.Convert(v)
	return save(t.sv, data, x, e)
}

// Validate the data map contains the necessary values and those values have
// the correct underlying type and value.
func (t *CephChoices) Validate(data map[string]any) error {
	return checkEntry(t.sv, t, data)
}

/* Type: Ceph String */

// CephString arguments represent arbitrary strings.
type CephString struct {
	sv *SignatureVar
}

// TypeName returns the name of this ceph argument type.
func (*CephString) TypeName() string { return CephTypeString }

// Name returns the name of this ceph argument.
func (t *CephString) Name() string { return t.sv.Name }

// Convert an any value into a valid underlying type for later serialization.
// Returns new type as any or error if conversion fails.
func (*CephString) Convert(v any) (any, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	if ss, ok := v.(fmt.Stringer); ok {
		return ss.String(), nil
	}
	return "", fmt.Errorf("not a string: %v", v)
}

// Check that a given value meets requirements for this argument.
// Returns an error if this value fails the check.
func (t *CephString) Check(v any) error {
	if _, ok := v.(string); !ok {
		return fmt.Errorf("not a string: %v (at %s)", v, t.sv.Name)
	}
	return nil
}

// Set the given value into the map ensuring that the value is of the correct
// underlying type. If the type is not valid returns an error.
func (t *CephString) Set(data map[string]any, v any) error {
	x, e := t.Convert(v)
	return save(t.sv, data, x, e)
}

// Validate the data map contains the necessary values and those values have
// the correct underlying type and value.
func (t *CephString) Validate(data map[string]any) error {
	return checkEntry(t.sv, t, data)
}

/* Type: Ceph Int */

// CephInt arguments represent integer valued arguments.
type CephInt struct {
	sv *SignatureVar
}

// TypeName returns the name of this ceph argument type.
func (*CephInt) TypeName() string { return CephTypeInt }

// Name returns the name of this ceph argument.
func (t *CephInt) Name() string { return t.sv.Name }

// Convert an any value into a valid underlying type for later serialization.
// Returns new type as any or error if conversion fails.
func (*CephInt) Convert(v any) (any, error) {
	switch vv := v.(type) {
	case int, int64, uint64, int32, uint32, int16, uint16, int8, uint8:
		return vv, nil
	case string:
		return strconv.ParseInt(vv, 10, 64)
	}
	return "", fmt.Errorf("not a CephInt: %v", v)
}

// Check that a given value meets requirements for this argument.
// Returns an error if this value fails the check.
func (t *CephInt) Check(v any) error {
	switch v.(type) {
	case int, int64, uint64, int32, uint32, int16, uint16, int8, uint8:
		return nil
	}
	return fmt.Errorf("not a CephInt: %v (at %s)", v, t.sv.Name)
}

// Set the given value into the map ensuring that the value is of the correct
// underlying type. If the type is not valid returns an error.
func (t *CephInt) Set(data map[string]any, v any) error {
	x, e := t.Convert(v)
	return save(t.sv, data, x, e)
}

// Validate the data map contains the necessary values and those values have
// the correct underlying type and value.
func (t *CephInt) Validate(data map[string]any) error {
	return checkEntry(t.sv, t, data)
}

/* Type: Ceph Float */

// CephFloat arguments represent floating-point valued arguments.
type CephFloat struct {
	sv *SignatureVar
}

// TypeName returns the name of this ceph argument type.
func (*CephFloat) TypeName() string { return CephTypeFloat }

// Name returns the name of this ceph argument.
func (t *CephFloat) Name() string { return t.sv.Name }

// Convert an any value into a valid underlying type for later serialization.
// Returns new type as any or error if conversion fails.
func (*CephFloat) Convert(v any) (any, error) {
	switch vv := v.(type) {
	case float64, float32:
		return vv, nil
	case string:
		return strconv.ParseFloat(vv, 64)
	}
	return "", fmt.Errorf("not a CephFloat: %v", v)
}

// Check that a given value meets requirements for this argument.
// Returns an error if this value fails the check.
func (t *CephFloat) Check(v any) error {
	switch v.(type) {
	case float64, float32:
		return nil
	}
	return fmt.Errorf("not a float: %v (at %s)", v, t.sv.Name)
}

// Set the given value into the map ensuring that the value is of the correct
// underlying type. If the type is not valid returns an error.
func (t *CephFloat) Set(data map[string]any, v any) error {
	x, e := t.Convert(v)
	return save(t.sv, data, x, e)
}

// Validate the data map contains the necessary values and those values have
// the correct underlying type and value.
func (t *CephFloat) Validate(data map[string]any) error {
	return checkEntry(t.sv, t, data)
}

/* Type: Ceph Bool */

// CephBool arguments represent boolean valued arguments.
type CephBool struct {
	sv *SignatureVar
}

// TypeName returns the name of this ceph argument type.
func (*CephBool) TypeName() string { return CephTypeBool }

// Name returns the name of this ceph argument.
func (t *CephBool) Name() string { return t.sv.Name }

// Convert an any value into a valid underlying type for later serialization.
// Returns new type as any or error if conversion fails.
func (*CephBool) Convert(v any) (any, error) {
	switch vv := v.(type) {
	case bool:
		return vv, nil
	case string:
		return strconv.ParseBool(vv)
	}
	return "", fmt.Errorf("not a CephBool: %v", v)
}

// Check that a given value meets requirements for this argument.
// Returns an error if this value fails the check.
func (t *CephBool) Check(v any) error {
	if _, ok := v.(bool); !ok {
		return fmt.Errorf("not a bool: %v (at %s)", v, t.sv.Name)
	}
	return nil
}

// Set the given value into the map ensuring that the value is of the correct
// underlying type. If the type is not valid returns an error.
func (t *CephBool) Set(data map[string]any, v any) error {
	x, e := t.Convert(v)
	return save(t.sv, data, x, e)
}

// Validate the data map contains the necessary values and those values have
// the correct underlying type and value.
func (t *CephBool) Validate(data map[string]any) error {
	return checkEntry(t.sv, t, data)
}

/* Type: Ceph Pool Name */

// CephPoolName arguments represent strings limited to naming ceph pools.
type CephPoolName struct {
	CephString
}

// TypeName returns the name of this ceph argument type.
func (*CephPoolName) TypeName() string { return CephTypePoolName }

/* Type: Ceph Object Name */

// CephObjectName arguments represent strings limited to naming ceph objects.
type CephObjectName struct {
	CephString
}

// TypeName returns the name of this ceph argument type.
func (*CephObjectName) TypeName() string { return CephTypeObjectName }

/* Type: Ceph OSD Name */

// CephOSDName arguments represent strings limited to naming ceph OSDs.
type CephOSDName struct {
	CephString
}

// TypeName returns the name of this ceph argument type.
func (*CephOSDName) TypeName() string { return CephTypeOSDName }

/* Type: Ceph PG ID */

// CephPGID arguments represent strings limited to identifying ceph PGs.
type CephPGID struct {
	CephString
}

// TypeName returns the name of this ceph argument type.
func (*CephPGID) TypeName() string { return CephTypePGID }

/* Type: Unknown */

// CephUnknownType is a placeholder type for other unknown or unimplemented
// types.
type CephUnknownType struct {
	sv *SignatureVar
}

// TypeName returns the name of this ceph argument type.
func (*CephUnknownType) TypeName() string { return "(Unknown)" }

// Name returns the name of this ceph argument.
func (t *CephUnknownType) Name() string { return t.sv.Name }

// Set the given value into the map ensuring that the value is of the correct
// underlying type. If the type is not valid returns an error.
func (t *CephUnknownType) Set(map[string]any, any) error {
	return fmt.Errorf("Set operation invalid: Unknown Type: %s", t.sv.Type)
}

// Validate the data map contains the necessary values and those values have
// the correct underlying type and value.
func (*CephUnknownType) Validate(map[string]any) error {
	return nil
}

// CephRepeatedArg is a special argument type that wraps a more basic
// (scalar) type allowing it to be repeated in the argument sequence.
type CephRepeatedArg struct {
	inner CephScalarArgumentType
	sv    *SignatureVar
}

// TypeName returns the name of this ceph argument type.
func (t *CephRepeatedArg) TypeName() string {
	return fmt.Sprintf("%s (Repeat: %s)", t.inner.TypeName(), t.sv.Repeat)
}

// Name returns the name of this ceph argument.
func (t *CephRepeatedArg) Name() string { return t.sv.Name }

// Set the given value into the map ensuring that the value is of the correct
// underlying type. If the type is not valid returns an error.
func (t *CephRepeatedArg) Set(data map[string]any, v any) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Slice {
		// input is a slice
		temp := map[string]any{
			t.sv.Name: []any{},
		}
		for i := 0; i < rval.Len(); i++ {
			if err := t.Append(temp, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
		data[t.sv.Name] = temp[t.sv.Name]
		return nil
	}
	// input is not a slice
	return t.Append(data, v)
}

// Append the given value onto a slice in the data map. Works like Set but
// assumes a single value of the desired type (string for CephString, int or
// string-with-int-value for CephInt).
func (t *CephRepeatedArg) Append(data map[string]any, v any) error {
	key := t.sv.Name
	var temp []any
	if _, ok := data[key]; !ok {
		temp = []any{}
	} else {
		temp = data[key].([]any)
	}
	vv, err := t.inner.Convert(v)
	if err != nil {
		return err
	}
	data[key] = append(temp, vv)
	return nil
}

// Validate the data map contains the necessary values and those values have
// the correct underlying type and value.
func (t *CephRepeatedArg) Validate(data map[string]any) error {
	if v, ok := data[t.sv.Name]; ok {
		if vv, ok := v.([]any); ok {
			for i := range vv {
				if err := t.inner.Check(vv[i]); err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("not a slice: %v (at %s)", v, t.sv.Name)
		}
		return nil
	}
	if t.sv.Required() {
		return fmt.Errorf("missing required arg: %s", t.sv.Name)
	}
	return nil
}

func save(
	sv *SignatureVar, data map[string]any, v any, e error) error {
	// ---
	if e != nil {
		return e
	}
	data[sv.Name] = v
	return nil
}

func checkEntry(
	sv *SignatureVar, t CephScalarArgumentType, data map[string]any) error {
	// ---
	v, ok := data[sv.Name]
	if !ok {
		if sv.Required() {
			return fmt.Errorf("missing required arg: %s", sv.Name)
		}
		return nil
	}
	return t.Check(v)
}
