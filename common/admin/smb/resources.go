//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/ceph/go-ceph/internal/commands"
)

// ErrUnknownResourceType indicates that JSON values contained a resource
// type value unknown to this library.
var ErrUnknownResourceType = errors.New("unknown resource type")

// Resource is an interface provided for working with abstract resource
// description structures in the Ceph smb module.
type Resource interface {
	// Type returns the ResourceType enum value for the resource.
	Type() ResourceType
	// Identity returns a resource reference for the resource.
	Identity() ResourceRef
	// Intent returns the intent value for the resource.
	Intent() Intent
	// Validate returns an error if the resource is not well-formed or
	// incomplete.
	Validate() error
}

type resourceGroup struct {
	Resources []Resource `json:"resources"`
}

func (g *resourceGroup) UnmarshalJSON(data []byte) error {
	var stub stubResource
	if err := json.Unmarshal(data, &stub); err == nil &&
		validResourceType(stub.ResourceType) {
		// single resource
		var entry resourceEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return err
		}
		g.Resources = []Resource{entry.r}
		return nil
	}

	var vg struct {
		Resources []resourceEntry `json:"resources"`
	}
	if err := json.Unmarshal(data, &vg); err != nil {
		return err
	}
	g.Resources = make([]Resource, len(vg.Resources))
	for i := range vg.Resources {
		g.Resources[i] = vg.Resources[i].r
	}
	return nil
}

type stubResource struct {
	ResourceType ResourceType `json:"resource_type"`
	IntentValue  Intent       `json:"intent"`
}

type resourceEntry struct {
	stubResource
	r Resource
}

func (e *resourceEntry) UnmarshalJSON(data []byte) error {
	stub := stubResource{}
	if err := json.Unmarshal(data, &stub); err != nil {
		return err
	}
	e.stubResource = stub

	switch stub.ResourceType {
	case ClusterType:
		e.r = new(Cluster)
	case ShareType:
		e.r = new(Share)
	case JoinAuthType:
		e.r = new(JoinAuth)
	case UsersAndGroupsType:
		e.r = new(UsersAndGroups)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownResourceType, stub.ResourceType)
	}

	return json.Unmarshal(data, e.r)
}

func validResourceType(v ResourceType) bool {
	switch v {
	case ClusterType, ShareType, JoinAuthType, UsersAndGroupsType:
		return true
	}
	return false
}

var rl = []rune("bcdfghjklmnpqrstvwxyz")

func randName(prefix string) string {
	// max len = 18; suffix len = 8
	suffix := make([]rune, 8)
	for i := range suffix {
		suffix[i] = rl[rand.Intn(len(rl))]
	}
	n := len(prefix)
	if n > 10 {
		n = 10
	}
	return prefix[:n] + string(suffix)
}

// ShowOptions controls optional behavior of the Show function.
type ShowOptions struct {
	// PasswordFilter can be used to filter/obfuscate password values
	// stored on the Ceph cluster.
	PasswordFilter PasswordFilter
}

// Show smb module resource descriptions stored on the Ceph cluster.
// If any values are provided in the refs slice, the function will query
// only resources matching those references. These may be all matching
// resources of a type or more specific references with IDs. The opts value
// can be nil for default behavior or supplied to customize the query results.
// Currently, ShowOptions can be used to filter password values.
//
// Similar To:
//
//	ceph smb show
func (a *Admin) Show(refs []ResourceRef, opts *ShowOptions) (
	[]Resource, error) {

	rnames := make([]string, len(refs))
	for i := range refs {
		rnames[i] = refs[i].String()
	}
	m := map[string]any{
		"prefix":         "smb show",
		"format":         "json",
		"resource_names": rnames,
		"results":        "full",
	}
	if opts != nil && opts.PasswordFilter != PasswordFilterUnset {
		m["password_filter"] = string(opts.PasswordFilter)
	}
	g := resourceGroup{}
	c := commands.MarshalMgrCommand(a.conn, m)
	if err := c.NoStatus().Unmarshal(&g).End(); err != nil {
		return nil, err
	}
	return g.Resources, nil
}

// ApplyOptions controls optional behavior of the Apply function.
type ApplyOptions struct {
	// PasswordFilter can be used to filter/obfuscate password values
	// sent to the Ceph cluster.
	PasswordFilter PasswordFilter
	// PasswordFilterOut can be used to filter/obfuscate password values
	// returned from the Ceph cluster.
	PasswordFilterOut PasswordFilter
}

// Apply changes to the resource descriptions stored on the Ceph cluster.
// Supply one or more Resource objects in the slice and these resources will
// be created, updated, or removed based on the Resource's parameters.
// An Intent() of Present will create or update a resource.
// An Intent() of Removed will remove a matching resource or be a no-op if nothing
// is matched.
// The opts value can be nil for default behavior or supplied to customize the
// way the command processes inputs and outputs. Currently, the password values
// supplied in the objects and returned in the result can be filtered depending
// on the fields in the ApplyOptions structure.
//
// Similar To:
//
//	ceph smb apply -i -
func (a *Admin) Apply(resources []Resource, opts *ApplyOptions) (
	ResultGroup, error) {

	rg := ResultGroup{}
	if err := ValidateResources(resources); err != nil {
		return rg, err
	}
	buf, err := json.Marshal(resourceGroup{Resources: resources})
	if err != nil {
		return rg, err
	}
	m := map[string]string{
		"prefix": "smb apply",
		"format": "json",
	}
	if opts != nil && opts.PasswordFilter != PasswordFilterUnset {
		m["password_filter"] = string(opts.PasswordFilter)
	}
	if opts != nil && opts.PasswordFilterOut != PasswordFilterUnset {
		m["password_filter_out"] = string(opts.PasswordFilterOut)
	}
	c := commands.MarshalMgrCommandWithBuffer(a.conn, m, buf)
	if err := c.NoStatus().Unmarshal(&rg).End(); err != nil {
		return rg, err
	}
	return rg, nil
}

// ValidateResources returns an error if any resource in the supplied slice
// is invalid. It returns nil if all resources are valid. The first invalid
// resource will be identified and described in the resulting error.
func ValidateResources(resources []Resource) error {
	for i, res := range resources {
		if err := res.Validate(); err != nil {
			return fmt.Errorf("Resource #%d: %s: %w", i, res.Identity(), err)
		}
	}
	return nil
}

// errorMerge is a helper function for combining a lower level error from
// the api with a resultGroup, treating the result group as an error if it
// is not successful.
func errorMerge(results ResultGroup, err error) error {
	if err != nil {
		return err
	}
	if !results.Ok() {
		return results
	}
	return nil
}

// RemoveCluster will remove a Cluster resource with a matching ID value
// from the Ceph cluster. This is a convenience function that creates a
// Cluster resource to remove and then applies it in one step.
func (a *Admin) RemoveCluster(clusterID string) error {
	rl := []Resource{NewClusterToRemove(clusterID)}
	return errorMerge(a.Apply(rl, nil))
}

// RemoveShare will remove a Share resource with matching ID values from
// the Ceph cluster. This is a convenience function that creates a Share
// resource to remove and then applies it in one step.
func (a *Admin) RemoveShare(clusterID, shareID string) error {
	rl := []Resource{NewShareToRemove(clusterID, shareID)}
	return errorMerge(a.Apply(rl, nil))
}

// RemoveJoinAuth will remove a JoinAuth resource with a matching ID value
// from the Ceph cluster. This is a convenience function that creates a
// JoinAuth resource to remove and then applies it in one step.
func (a *Admin) RemoveJoinAuth(authID string) error {
	rl := []Resource{NewJoinAuthToRemove(authID)}
	return errorMerge(a.Apply(rl, nil))
}

// RemoveUsersAndGroups will remove a UsersAndGroups resource with a matching
// ID value from the Ceph cluster. This is a convenience function that creates
// a UsersAndGroups resource to remove and then applies it in one step.
func (a *Admin) RemoveUsersAndGroups(ugID string) error {
	rl := []Resource{NewUsersAndGroupsToRemove(ugID)}
	return errorMerge(a.Apply(rl, nil))
}
