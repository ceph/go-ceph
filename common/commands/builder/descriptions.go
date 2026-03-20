//go:build !(pacific || quincy) && ceph_preview

package builder

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

// SignatureVar describes variable arguments in a Ceph command description.
type SignatureVar struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Req     *bool  `json:"req"`
	Choices string `json:"strings"`
	Repeat  string `json:"n"`
}

// Required returns true if the variable is required.
func (sv SignatureVar) Required() bool {
	if sv.Req == nil {
		return true
	}
	return *(sv.Req)
}

// SignatureElement describes a single element in a Ceph command description.
// It can either be a static or fixed string (that will be part of the command
// prefix) or a variable input.
type SignatureElement struct {
	Static   string
	Variable *SignatureVar
}

// UnmarshalJSON decodes JSON into a SignatureElement.
func (se *SignatureElement) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &se.Static); err == nil {
		return nil
	}
	se.Variable = new(SignatureVar)
	return json.Unmarshal(data, &se.Variable)
}

// Description represents a single ceph command known to a process such
// as the Ceph MON, Ceph MGR, or so on.
type Description struct {
	Key    string
	Sig    []*SignatureElement `json:"sig"`
	Help   string              `json:"help"`
	Module string              `json:"module"`
	Perm   string              `json:"perm"`
	Flags  uint64              `json:"flags"`
}

// Prefix returns the static strings in the signature of a command as a slice
// of strings.
func (d Description) Prefix() []string {
	p := []string{}
	for _, elem := range d.Sig {
		if elem.Variable != nil {
			break
		}
		p = append(p, elem.Static)
	}
	return p
}

// PrefixString returns the static strings in the signature of a command as a
// single space separated string.
func (d Description) PrefixString() string {
	return strings.Join(d.Prefix(), " ")
}

// Variables returns the variable components in a ceph command signature in a
// slice.
func (d Description) Variables() []*SignatureVar {
	v := []*SignatureVar{}
	for _, elem := range d.Sig {
		if elem.Variable == nil {
			continue
		}
		v = append(v, elem.Variable)
	}
	return v
}

// CommandDescriptions is a wrapper type to encapsulate the ceph commands
// known to a particular process. Methods such as Match or Find can be
// used to narrow down commands to a set with matching prefix terms.
type CommandDescriptions struct {
	Entries []Description
}

// UnmarshalJSON decodes JSON data into a CommandDescriptions object.
func (cd *CommandDescriptions) UnmarshalJSON(data []byte) error {
	cdmap := map[string]json.RawMessage{}
	err := json.Unmarshal(data, &cdmap)
	if err != nil {
		return err
	}
	for _, key := range slices.Sorted(maps.Keys(cdmap)) {
		var desc Description
		err = json.Unmarshal(cdmap[key], &desc)
		if err != nil {
			return err
		}
		desc.Key = key
		cd.Entries = append(cd.Entries, desc)
	}
	return nil
}

func matchDescriptions(ds []Description, index int, term string) []Description {
	out := []Description{}
	for _, desc := range ds {
		s := desc.Sig[index]
		if s.Variable == nil && s.Static == term {
			out = append(out, desc)
		}
	}
	return out
}

// Match returns all command Descriptions that have full or partially matching
// prefix strings. For example, passing `[]string{"osd"}` will return a slice
// with all commands where the first prefix term is "osd".  Passing
// `[]string{"osd", "ls"}` will return a slice with all commands where the
// first two prefix terms are "osd" and "ls".
func (cd *CommandDescriptions) Match(terms []string) []Description {
	matches := cd.Entries
	for idx, term := range terms {
		matches = matchDescriptions(matches, idx, term)
	}
	return matches
}

// Find returns all command Descriptions that have full or partially matching
// prefix strings. This is like Match but uses variable arguments instead
// of a slice for convenience in code where you know what command you
// intend to call. For example: `matches = cd.Find("osd", "rm")`.
func (cd *CommandDescriptions) Find(n ...string) []Description {
	return cd.Match(n)
}
