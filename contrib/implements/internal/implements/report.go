package implements

import (
	"fmt"
	"sort"
	"strings"
)

// ReportOptions is a common set of options for reports.
type ReportOptions struct {
	List     bool
	Annotate bool
}

// Reporter is a common interface to report on the "implements"
// analysis.
type Reporter interface {
	// Report reports on the given (sub)package with the given inspector's
	// contents.
	Report(string, *Inspector) error
	// Done flushes any buffered state between calls to Report.
	Done() error
}

// TextReport implements a streaming plain-text output report.
type TextReport struct {
	o ReportOptions
}

// NewTextReport creates a new TextReport.
func NewTextReport(o ReportOptions) *TextReport {
	return &TextReport{o}
}

// Report on the given code inspector.
func (r *TextReport) Report(name string, ii *Inspector) error {
	o := r.o
	packageLabel := strings.ToUpper(name)

	ii.update()

	found := len(ii.found)
	total := len(ii.expected)
	fmt.Printf(
		"%s functions covered: %d/%d : %v%%\n",
		packageLabel,
		found,
		total,
		(100*found)/total)
	missing := total - found - ii.deprecatedMissing
	fmt.Printf(
		"%s functions missing: %d/%d : %v%%\n",
		packageLabel,
		missing,
		total,
		(100*missing)/total)
	fmt.Printf(
		"  (note missing count does not include deprecated functions in ceph)\n")

	if !o.List {
		return nil
	}
	sort.Sort(ii.expected)
	for _, cf := range ii.expected {
		if flags, ok := ii.found[cf.Name]; ok {
			tags := ""
			if o.Annotate {
				if flags&isCalled == isCalled {
					tags += " called"
				}
				if flags&isDocumented == isDocumented {
					tags += " documented"
				}
				if flags&isDeprecated == isDeprecated {
					tags += " deprecated"
				}
				if tags != "" {
					tags = " (" + strings.TrimSpace(tags) + ")"
				}
			}
			fmt.Printf("  Covered: %s%s\n", cf.Name, tags)
		}
	}
	for _, cf := range ii.expected {
		if _, ok := ii.found[cf.Name]; ok {
			continue
		}
		d := ""
		if o.Annotate && cf.isDeprecated() {
			d = " (deprecated)"
		}
		fmt.Printf("  Missing: %s%s\n", cf.Name, d)
	}
	return nil
}

// Done updating report with inspectors.
func (*TextReport) Done() error { return nil }
