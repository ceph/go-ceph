//go:build !(pacific || quincy) && ceph_preview

package builder

import (
	"encoding/json"
	"fmt"
)

// CephTypeFunc signatures describe a function that can provide a
// CephArgumentType instance given a SignatureVar object. This is used to
// customize argument type look ups if needed.
type CephTypeFunc func(*SignatureVar) CephArgumentType

// Builder objects are used to construct command inputs that interact with Ceph
// APIs such as MgrCommand, MonCommand, and so forth.  This type provides a
// MarshalJSON method that will return JSON encoded bytes that can be passed as
// the first argument to these RADOS APIs.  The MarshalJSON uses the command
// description to produce argument types that validate the contents of the
// Values map.
// The Values map is public so that you can directly manipulate the contents in
// unplanned ways and customize what gets encoded in the final JSON.
// You can also customize the ceph types returned by setting an alternative
// GetType attribute. By default, this uses the BindArgumentType function but
// you can replace or re-use this function to return customized
// CephArgumentType values to fit your needs.
type Builder struct {
	Values      map[string]any
	Description Description
	GetType     CephTypeFunc
}

// NewBuilder returns a new command builder given a Description of ceph
// command.
func NewBuilder(d Description) *Builder {
	b := Builder{
		Values:      map[string]any{},
		Description: d,
		GetType:     BindArgumentType,
	}
	return b.Prepare()
}

// Prepare sets default values in the Values map. It is called
// automatically by NewBuilder. It can be used to reset values
// in the map if needed.
func (b *Builder) Prepare() *Builder {
	b.Values["prefix"] = b.Description.PrefixString()
	return b
}

// Arguments returns a slice of all the ceph argument types known
// to this builder.
func (b *Builder) Arguments() []CephArgumentType {
	out := []CephArgumentType{}
	for _, v := range b.Description.Variables() {
		out = append(out, b.GetType(v))
	}
	return out
}

// ArgumentsMap returns a map of argument names to the various ceph argument
// types known to this builder.
func (b *Builder) ArgumentsMap() map[string]CephArgumentType {
	m := map[string]CephArgumentType{}
	for _, argtype := range b.Arguments() {
		m[argtype.Name()] = argtype
	}
	return m
}

// Validate returns an error if the contents of the Values map do not match
// parameters defined by the ceph argument types known to this builder.
func (b *Builder) Validate() error {
	for _, t := range b.Arguments() {
		if err := t.Validate(b.Values); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) applyArgs(args []string) error {
	argtypes := b.Arguments()
	tlen := len(argtypes)
	for i, argval := range args {
		if i >= tlen {
			lat := argtypes[tlen-1]
			if alat, ok := lat.(CephMultiArgumentType); ok {
				if err := alat.Append(b.Values, argval); err != nil {
					return err
				}
			}
			continue
		}
		if err := argtypes[i].Set(b.Values, argval); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) applyNamedArgs(args map[string]string) error {
	m := b.ArgumentsMap()
	for k, argval := range args {
		cat, ok := m[k]
		if !ok {
			return fmt.Errorf("not found: %s", k)
		}
		if err := cat.Set(b.Values, argval); err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON returns the builder's Values map as JSON encoded bytes or an
// error if the Values don't validate or marshal to JSON.
func (b *Builder) MarshalJSON() ([]byte, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(b.Values)
}

// Apply takes string argument values, in either a slice (linear) or
// map (named) form and, using the known argument types, converts
// the values and stores the results in the builder's Values map.
//
// NB. This function doesn't handle repeat arguments (n:N) other than in
// the args slice and then only at the end of the slice. This function
// is meant to serve as a simple example for mapping argument values into
// a call to MonCommand/MgrCommand/etc. not implement everything the
// standard `ceph` command can do.
func (b *Builder) Apply(args []string, named map[string]string) error {
	if len(args) > 0 {
		if err := b.applyArgs(args); err != nil {
			return err
		}
	}
	if len(named) > 0 {
		if err := b.applyNamedArgs(named); err != nil {
			return err
		}
	}
	return nil
}

// BindArgumentType returns a CephArgumentType bound to the given
// SignatureVar.
func BindArgumentType(sv *SignatureVar) CephArgumentType {
	// go-ceph treats arguments with an "n" value as a special wrapper type
	// rather teach all types how to deal with repeats all over the simpler
	// types. We create an internal distinction between the regular types,
	// deemed "scalar" types and the one non-scalar Repeat type.
	if sv.Repeat == "N" {
		inner := getScalarArgumentType(sv)
		if st, ok := inner.(CephScalarArgumentType); ok {
			return &CephRepeatedArg{st, sv}
		}
		panic("inner type not a scalar type: " + inner.TypeName())
	}
	return getScalarArgumentType(sv)
}

func getScalarArgumentType(sv *SignatureVar) CephArgumentType {
	switch sv.Type {
	case CephTypeString:
		return &CephString{sv}
	case CephTypeChoices:
		return &CephChoices{sv}
	case CephTypeInt:
		return &CephInt{sv}
	case CephTypeFloat:
		return &CephFloat{sv}
	case CephTypeBool:
		return &CephBool{sv}
	case CephTypePoolName:
		return &CephPoolName{CephString{sv}}
	case CephTypeObjectName:
		return &CephObjectName{CephString{sv}}
	case CephTypeOSDName:
		return &CephOSDName{CephString{sv}}
	case CephTypePGID:
		return &CephPGID{CephString{sv}}
	}
	return &CephUnknownType{sv}
}
