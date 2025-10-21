//go:build !(octopus || pacific || quincy || reef || squid)

package smb

// Placement is passed to cephadm to determine where cluster services
// will be run.
type Placement map[string]any

// SimplePlacement returns a placement with common placement parameters - count
// and label - specified.
func SimplePlacement(count int, label string) Placement {
	p := Placement{}
	if count > 0 {
		p["count"] = count
	}
	if label != "" {
		p["label"] = label
	}
	return p
}
