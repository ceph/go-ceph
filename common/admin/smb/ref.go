//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"fmt"
)

// ResourceRef provides a structured interface to refer to resources.
type ResourceRef interface {
	Type() ResourceType
	String() string
}

// ------------------------------------------------
// Give the ResourceType the methods needed to make it also
// a ResourceRef.

// Type returns a ResourceType value.
func (rt ResourceType) Type() ResourceType {
	// brought to you by the dept. of redundancy dept.
	return rt
}

// String returns a string value referring to a resource.
func (rt ResourceType) String() string {
	return string(rt)
}

// ------------------------------------------------

// ResourceID refers to a resource via its ResourceType value and a string ID.
type ResourceID struct {
	ResourceType ResourceType
	ID           string
}

// Type returns a ResourceType value.
func (r ResourceID) Type() ResourceType { return r.ResourceType }

// String returns a string value referring to a resource ID.
func (r ResourceID) String() string {
	return fmt.Sprintf("%s.%s", r.ResourceType, r.ID)
}

// ------------------------------------------------

// ChildResourceID refers to a resource via its ResourceType value, the ID of
// a parent resource (typically a cluster) and a string ID for the child.
type ChildResourceID struct {
	ResourceType ResourceType
	ParentID     string
	ID           string
}

// Type returns a ResourceType value.
func (c ChildResourceID) Type() ResourceType { return c.ResourceType }

// String returns a string value referring to a child resource ID.
func (c ChildResourceID) String() string {
	return fmt.Sprintf("%s.%s.%s", c.ResourceType, c.ParentID, c.ID)
}
