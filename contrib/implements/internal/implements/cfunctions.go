package implements

import (
	"fmt"
	"strings"
)

// CFunction represents a function in C code.
type CFunction struct {
	Name string `xml:"name,attr"`
	Attr string `xml:"attributes,attr"`
}

// isDeprecated will return true if the C function is marked deprecated
// via attributes.
func (cf CFunction) isDeprecated() bool {
	return strings.Contains(cf.Attr, "deprecated")
}

// CFunctions is a sortable slice of CFunction.
type CFunctions []CFunction

func (cfs CFunctions) Len() int           { return len(cfs) }
func (cfs CFunctions) Swap(i, j int)      { cfs[i], cfs[j] = cfs[j], cfs[i] }
func (cfs CFunctions) Less(i, j int) bool { return cfs[i].Name < cfs[j].Name }

func (cfs CFunctions) ensure() (CFunctions, error) {
	if len(cfs) < 1 {
		return nil, fmt.Errorf("found %d c functions", len(cfs))
	}
	return cfs, nil
}
