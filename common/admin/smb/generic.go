//go:build !(pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ceph/go-ceph/internal/commands"
)

const (
	// ResourceTypeKey contains the key used for a resource's resource_type field.
	ResourceTypeKey = "resource_type"
	// IntentKey contains the key used for a resource's intent field.
	IntentKey = "intent"
)

func getResourceType(m map[string]any) ResourceType {
	return ResourceType(m[ResourceTypeKey].(string))
}

func hasResourceType(m map[string]any) bool {
	if v, ok := m[ResourceTypeKey]; ok {
		_, ok = v.(string)
		return ok
	}
	return false
}

func getStringValue(key string, m map[string]any) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GenericIdentityKind is a helper interface used to identify generic
// resource instances.
type GenericIdentityKind interface {
	Identity(map[string]any) ResourceRef
}

// TopIdentityKind contains metadata used to identify a top-level resource.
// A top-level resource has no parent resources and a single ID field when
// expressed in JSON or YAML.
type TopIdentityKind struct {
	IDKey string
}

// Identity extracts ID fields from the generic map and returns the
// resource reference for the TopIdentityKind.
func (t TopIdentityKind) Identity(m map[string]any) ResourceRef {
	return ResourceID{
		ResourceType: getResourceType(m),
		ID:           m[t.IDKey].(string),
	}
}

// ChildIdentityKind contains metadata used to identify a resource that
// has a child relationship to another resource. A Share in a Cluster for
// example. A child resource has two ID fields when expressed in JSON or YAML.
type ChildIdentityKind struct {
	ParentIDKey string
	IDKey       string
}

// Identity extracts ID fields from the generic map and returns the
// resource reference for the ChildIdentityKind.
func (c ChildIdentityKind) Identity(m map[string]any) ResourceRef {
	return ChildResourceID{
		ResourceType: getResourceType(m),
		ParentID:     m[c.ParentIDKey].(string),
		ID:           m[c.IDKey].(string),
	}
}

// GuessIdentityKind inspects the generic map and attempts to return the
// best GenericIdentityKind for the supplied data. If a good guess can not
// be made an error will be returned.
func GuessIdentityKind(m map[string]any) (GenericIdentityKind, error) {
	keys := []string{}
	for key := range m {
		if strings.HasSuffix(key, "_id") {
			keys = append(keys, key)
		}
	}
	// Because golang does not remember order in maps:
	// Reorder keys that are ids to be "parent keys" as currently the
	// only parent types are clusters. Unlike most of this module, this
	// is not future-proof but it is a *guess* as the function name says
	sort.Slice(keys, func(i, _ int) bool {
		if keys[i] == "cluster_id" {
			return true
		}
		return false
	})

	if len(keys) == 1 {
		return TopIdentityKind{IDKey: keys[0]}, nil
	}
	if len(keys) == 2 {
		return ChildIdentityKind{ParentIDKey: keys[0], IDKey: keys[1]}, nil
	}
	return nil, fmt.Errorf(
		"failed to guess identity kind (%d id keys found)", len(keys))
}

// GenericResource is a smb mgr module resource that can hold generic
// data or data that is not known to the concrete types implemented in
// this module. It can be used for working with new features, experimental
// forks, or other situations where the specific types are insufficient.
//
// Data is stored in the Values map and metadata used to identify the
// resource is provided by the IDKind field.
type GenericResource struct {
	Values map[string]any
	IDKind GenericIdentityKind
}

// Type returns a ResourceType value.
func (g *GenericResource) Type() ResourceType {
	return getResourceType(g.Values)
}

// Intent controls if a resource is to be created/updated or removed.
func (g *GenericResource) Intent() Intent {
	i := getStringValue(IntentKey, g.Values)
	if i == "" {
		return Present
	}
	return Intent(i)
}

// Identity returns a ResourceRef identifying this generic resource.
func (g *GenericResource) Identity() ResourceRef {
	return g.IDKind.Identity(g.Values)
}

// MarshalJSON supports marshalling a generic resource to JSON.
func (g *GenericResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Values)
}

// UnmarshalJSON supports unmarshalling a generic resource from JSON.
func (g *GenericResource) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &g.Values); err != nil {
		return err
	}

	var err error
	g.IDKind, err = GuessIdentityKind(g.Values)
	return err
}

// Validate returns an error describing an issue with the resource or
// nil if the object is valid.
func (g *GenericResource) Validate() error {
	if getStringValue(ResourceTypeKey, g.Values) == "" {
		return fmt.Errorf(
			"%w: %s", ErrUnknownResourceType, "resource_type not set")
	}
	intent := Intent(getStringValue(IntentKey, g.Values))
	if intent != Present && intent != Removed && intent != Intent("") {
		return fmt.Errorf("invalid intent %s", intent)
	}
	_, err := GuessIdentityKind(g.Values)
	return err
}

// Convert a generic resource to a specific resource. The function
// will return a Resource representing a concrete resource
// type such as *Share or *Cluster. If the resource type is unknown
// to this module or the conversion fails an error will be returned.
func (g *GenericResource) Convert() (Resource, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	var (
		r  Resource
		rt = g.Type()
	)
	if rnew, ok := resourceTypes[rt]; ok {
		r = rnew()
	} else {
		return nil, fmt.Errorf("%w: %s", ErrUnknownResourceType, rt)
	}
	j, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, &r)
	if err != nil {
		return nil, err
	}
	return r, err
}

// ToGeneric creates a new GenericResource from another resource object.
// Typically this will be some more concrete resource type such as
// *Share or *Cluster.
// One use of this function might be to fill in known fields using the
// well structured concrete resource type and then, once generic, extend
// it with fields not currently known to this module.
func ToGeneric(r Resource) (*GenericResource, error) {
	g := GenericResource{}
	j, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func showGenericUnmarshal(r commands.Response) ([]Resource, error) {
	g := struct{ Resources []GenericResource }{}
	if err := r.NoStatus().Unmarshal(&g).End(); err != nil {
		return nil, err
	}
	out := make([]Resource, len(g.Resources))
	for i := range g.Resources {
		out[i] = &g.Resources[i]
	}
	return out, nil
}

// SetGeneric enables or disables returning GenericResource objects from
// the Show function. If true all Resource objects returned from Show
// will be GenericResource instances.
func (opts *ShowOptions) SetGeneric(b bool) *ShowOptions {
	if b {
		opts.unmarshal = showGenericUnmarshal
	} else {
		opts.unmarshal = nil
	}
	return opts
}

// Generic returns true if the show options are set to return
// GenericResource instances from the Show function.
func (opts *ShowOptions) Generic() bool {
	return opts.unmarshal != nil
}
