package implements

import (
	"encoding/json"
	"io"
	"sort"
)

// JSONReport is a type that implements the Report interface and generates
// structured JSON.
type JSONReport struct {
	o    ReportOptions
	dest io.Writer

	data jrOut
}

type jrFlags foundFlags

func (f jrFlags) MarshalJSON() ([]byte, error) {
	o := []string{}
	flags := foundFlags(f)

	if flags&isCalled == isCalled {
		o = append(o, "called")
	}
	if flags&isDocumented == isDocumented {
		o = append(o, "documented")
	}
	if flags&isDeprecated == isDeprecated {
		o = append(o, "deprecated")
	}
	return json.Marshal(o)
}

type jrFunction struct {
	CName string   `json:"c_name"`
	Flags jrFlags  `json:"flags,omitempty"`
	Refs  []string `json:"referenced_by,omitempty"`
}

type jrPackage struct {
	Name    string `json:"name"`
	Summary struct {
		Total      int `json:"total"`
		Found      int `json:"found"`
		Missing    int `json:"missing"`
		Deprecated int `json:"deprecated"`
	} `json:"summary"`
	Found   []jrFunction `json:"found,omitempty"`
	Missing []jrFunction `json:"missing,omitempty"`
}

type jrOut map[string]jrPackage

// NewJSONReport creates a new json report. The JSON will be written to
// the supplied dest when Done is called.
func NewJSONReport(o ReportOptions, dest io.Writer) *JSONReport {
	return &JSONReport{o, dest, jrOut{}}
}

// Report will update the JSON report with the current code inspector's state.
func (r *JSONReport) Report(name string, ii *Inspector) error {
	ii.update()

	jp := jrPackage{}
	total := len(ii.expected)
	found := len(ii.found)
	jp.Summary.Total = total
	jp.Summary.Found = found
	jp.Summary.Missing = total - found - ii.deprecatedMissing
	jp.Summary.Deprecated = ii.deprecatedMissing
	jp.Name = name

	if r.o.List {
		collectFuncs(&jp, ii)
	}
	r.data[name] = jp
	return nil
}

func collectFuncs(jp *jrPackage, ii *Inspector) {
	sort.Sort(ii.expected)
	for _, cf := range ii.expected {
		if flags, ok := ii.found[cf.Name]; ok {
			refm := map[string]bool{}
			if n := ii.visitor.callMap[cf.Name]; n != "" {
				refm[n] = true
			}
			if n := ii.visitor.docMap[cf.Name]; n != "" {
				refm[n] = true
			}
			jp.Found = append(jp.Found,
				jrFunction{cf.Name, jrFlags(flags), mkeys(refm)})
		}
	}
	for _, cf := range ii.expected {
		if _, ok := ii.found[cf.Name]; ok {
			continue
		}
		var flags jrFlags
		if cf.isDeprecated() {
			flags |= jrFlags(isDeprecated)
		}
		jp.Missing = append(jp.Missing,
			jrFunction{cf.Name, flags, []string{}})
	}
}

func mkeys(m map[string]bool) []string {
	o := make([]string, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	return o
}

// Done completes the JSON report and writes the JSON to the output.
func (r *JSONReport) Done() error {
	enc := json.NewEncoder(r.dest)
	enc.SetIndent("", "  ")
	return enc.Encode(r.data)
}
