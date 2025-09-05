//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

// CustomPortsMap is used to configure a cluster with custom ports for
// a specified service type.
type CustomPortsMap map[Service]int

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
	// CustomPorts allows the customization of network port binding
	// by virtual service name [PREVIEW].
	CustomPorts CustomPortsMap `json:"custom_ports,omitempty"`
}

// Validate returns an error describing an issue with the resource or
// nil if the resource object is valid.
func (cluster *Cluster) Validate() error {
	return cluster.coreValidate()
}

// PREVIEW Field Group tracking
// Group 1:
//   CustomPorts
