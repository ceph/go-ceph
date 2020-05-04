package implements

import (
	"strings"
)

type foundFlags int

const (
	isCalled     = foundFlags(1)
	isDocumented = foundFlags(2)
	isDeprecated = foundFlags(4)
)

// Inspector types collect the high-level results from C and Go
// code scans.
type Inspector struct {
	visitor *visitor

	expected          CFunctions
	found             map[string]foundFlags
	deprecatedMissing int
}

// SetExpected sets the expected C functions, asuming the supplied prefix.
func (ii *Inspector) SetExpected(prefix string, expected CFunctions) error {
	ii.expected = make([]CFunction, 0, len(expected))
	for _, cfunc := range expected {
		if strings.HasPrefix(cfunc.Name, prefix) {
			logger.Printf("C function \"%s\" has matching prefix", cfunc.Name)
			ii.expected = append(ii.expected, cfunc)
		}
	}
	_, err := ii.expected.ensure()
	return err
}

func (ii *Inspector) update() {
	ii.found = map[string]foundFlags{}
	ii.deprecatedMissing = 0
	for i := range ii.expected {
		n := ii.expected[i].Name
		if _, found := ii.visitor.callMap[n]; found {
			ii.found[n] |= isCalled
		}
		if _, found := ii.visitor.docMap[n]; found {
			ii.found[n] |= isDocumented
		}
		if ii.expected[i].isDeprecated() {
			if _, found := ii.found[n]; found {
				ii.found[n] |= isDeprecated
			} else {
				ii.deprecatedMissing++
			}
		}
	}
}

// NewInspector returns a newly created code inspector object.
func NewInspector() *Inspector {
	return &Inspector{
		visitor: newVisitor(),
	}
}
