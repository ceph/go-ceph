//go:build !(octopus || pacific || quincy || reef || squid)

package smb

import (
	"encoding/json"
	"fmt"
)

// JoinAuthSource identifies a Join Auth resource that will be used
// as a source of authentication parameters to join a cluster to
// a domain.
type JoinAuthSource struct {
	SourceType SourceType `json:"source_type"`
	Ref        string     `json:"ref"`
}

// UserGroupSource identifies a Users and Groups resource that will be
// used as a source of user and group information on the SMB cluster.
type UserGroupSource struct {
	SourceType SourceType `json:"source_type"`
	Ref        string     `json:"ref"`
}

// DomainSettings are used to configure domain related settings for a
// cluster using active directory authentication.
type DomainSettings struct {
	// Realm identifies the AD/Kerberos Realm to use.
	Realm string `json:"realm"`
	// JoinSources should contain one or more JoinAuthSource that
	// the cluster may use to join a domain.
	JoinSources []JoinAuthSource `json:"join_sources"`
}

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

// PublicAddress used by a cluster with integrated Samba clustering enabled.
type PublicAddress struct {
	Address     string
	Destination []string
}

// Cluster configures an SMB Cluster resource that is managed within a
// Ceph cluster.
type Cluster struct {
	IntentValue       Intent            `json:"intent"`
	ClusterID         string            `json:"cluster_id"`
	AuthMode          ClusterAuthMode   `json:"auth_mode"`
	DomainSettings    *DomainSettings   `json:"domain_settings,omitempty"`
	UserGroupSettings []UserGroupSource `json:"user_group_settings,omitempty"`
	CustomDNS         []string          `json:"custom_dns,omitempty"`
	Placement         Placement         `json:"placement,omitempty"`
	Clustering        Clustering        `json:"clustering,omitempty"`
	PublicAddrs       []PublicAddress   `json:"public_addrs,omitempty"`
}

// Type returns a ResourceType value.
func (*Cluster) Type() ResourceType {
	return ClusterType
}

// Intent controls if a resource is to be created/updated or removed.
func (cluster *Cluster) Intent() Intent {
	return cluster.IntentValue
}

// Identity returns a ResourceRef identifying this cluster.
func (cluster *Cluster) Identity() ResourceRef {
	return ResourceID{
		ResourceType: cluster.Type(),
		ID:           cluster.ClusterID,
	}
}

// MarshalJSON supports marshalling a cluster to JSON.
func (cluster *Cluster) MarshalJSON() ([]byte, error) {
	type vCluster Cluster
	type wCluster struct {
		ResourceType ResourceType `json:"resource_type"`
		vCluster
	}
	return json.Marshal(wCluster{
		ResourceType: cluster.Type(),
		vCluster:     vCluster(*cluster),
	})
}

// Validate returns an error describing an issue with the resource or
// nil if the resource object is valid.
func (cluster *Cluster) Validate() error {
	var minimal bool
	switch cluster.IntentValue {
	case Present:
	case Removed:
		minimal = true
	default:
		return fmt.Errorf("missing intent")
	}
	if cluster.ClusterID == "" {
		return fmt.Errorf("missing ClusterID")
	}
	if minimal {
		return nil // minimal checks have been done, return early
	}

	switch cluster.AuthMode {
	case ActiveDirectoryAuth:
		if cluster.DomainSettings == nil {
			return fmt.Errorf(
				"missing domain settings for %v cluster", ActiveDirectoryAuth)
		}
	case UserAuth:
		if len(cluster.UserGroupSettings) == 0 {
			return fmt.Errorf(
				"missing user and group settings for %v cluster",
				UserAuth)
		}
	default:
		return fmt.Errorf("invalid AuthMode: %#v", cluster.AuthMode)
	}
	return nil
}

// SetPlacement modifies a cluster's placment settings.
func (cluster *Cluster) SetPlacement(p Placement) *Cluster {
	cluster.Placement = p
	return cluster
}

// NewUserCluster returns a new Cluster with default values set to
// create/update a cluster with local users-and-groups defined.
// In addition to the cluster name, this function accepts zero or more
// ID values naming ceph.smb.usersgroups resources.
func NewUserCluster(clusterID string, ids ...string) *Cluster {
	ugs := make([]UserGroupSource, len(ids))
	for i := range ids {
		ugs[i] = UserGroupSource{
			SourceType: ResourceSource,
			Ref:        ids[i],
		}
	}
	return &Cluster{
		IntentValue:       Present,
		ClusterID:         clusterID,
		AuthMode:          UserAuth,
		UserGroupSettings: ugs,
	}
}

// NewActiveDirectoryCluster returns a new Cluster with default values set to
// create/update a cluster with active directory authentication.
// In addition to the cluster name, this function accepts the name of the
// domain/realm and zero or more ID values naming ceph.smb.join.auth resources.
func NewActiveDirectoryCluster(
	clusterID string, realm string, ids ...string) *Cluster {

	jss := make([]JoinAuthSource, len(ids))
	for i := range ids {
		jss[i] = JoinAuthSource{
			SourceType: ResourceSource,
			Ref:        ids[i],
		}
	}
	return &Cluster{
		IntentValue: Present,
		ClusterID:   clusterID,
		AuthMode:    ActiveDirectoryAuth,
		DomainSettings: &DomainSettings{
			Realm:       realm,
			JoinSources: jss,
		},
	}
}

// NewClusterToRemove return a new Cluster with default values set to remove a
// cluster from management.
func NewClusterToRemove(clusterID string) *Cluster {
	return &Cluster{IntentValue: Removed, ClusterID: clusterID}
}
