//go:build !(octopus || pacific || quincy || reef || squid)

package smb

// Intent indicates how a resource description should be processed.
type Intent string

const (
	// Present resources will be created or updated.
	Present = Intent("present")
	// Removed resources will be removed or ignored.
	Removed = Intent("removed")
)

// ResourceType values are used to identify the type of a resource.
type ResourceType string

const (
	// ClusterType resources represent SMB clusters.
	ClusterType = ResourceType("ceph.smb.cluster")
	// ShareType resources represent SMB shares.
	ShareType = ResourceType("ceph.smb.share")
	// JoinAuthType resources contain information used to join a domain.
	JoinAuthType = ResourceType("ceph.smb.join.auth")
	// UsersAndGroupsType resources contain data used to define users and groups.
	UsersAndGroupsType = ResourceType("ceph.smb.usersgroups")
	// TLSCredentialType resources contain data used to establish TLS
	// secured network connections.
	TLSCredentialType = ResourceType("ceph.smb.tls.credential")
)

// SourceType indicates how a Cluster resource refers to another resource it
// needs. Currently only ResourceSource is available.
type SourceType string

const (
	// ResourceSource indicates that another resource is being referenced.
	ResourceSource = SourceType("resource")
)

// ClusterAuthMode indicates how a Cluster should authenticate users.
type ClusterAuthMode string

const (
	// ActiveDirectoryAuth indicates a cluster will use an active directory domain.
	ActiveDirectoryAuth = ClusterAuthMode("active-directory")
	// UserAuth indicates a cluster will use locally defined users and groups.
	UserAuth = ClusterAuthMode("user")
)

// Clustering indicates how an abstract cluster should be managed.
type Clustering string

const (
	// DefaultClustering indicates SMB clustering should be enabled based on
	// the placement value.
	DefaultClustering = Clustering("default")
	// NeverClustering indicates SMB clustering should never be enabled.
	NeverClustering = Clustering("never")
	// AlwaysClustering indicates SMB clustering should always be enabled.
	AlwaysClustering = Clustering("always")
)

// AccessCategory determines if share login controls applies to a user
// or group.
type AccessCategory string

const (
	// UserAccess indicates a share login control applies to a user.
	UserAccess = AccessCategory("user")
	// GroupAccess indicates a share login control applies to a group.
	GroupAccess = AccessCategory("group")
)

// AccessMode determines what kind of access a share login control will
// grant.
type AccessMode string

const (
	// ReadAccess grants read-only access to a share.
	ReadAccess = AccessMode("read")
	// ReadWriteAccess grants read-write access to a share.
	ReadWriteAccess = AccessMode("read-write")
	// AdminAccess grants administrative access to a share.
	AdminAccess = AccessMode("admin")
	// NoneAccess denies access to a share.
	NoneAccess = AccessMode("none")
)

// CephFSProvider indicates what method will be used to bridge smb services to
// CephFS.
type CephFSProvider string

const (
	// SambaVFSProvider sets the default VFS based provider.
	SambaVFSProvider = CephFSProvider("samba-vfs")
	// SambaVFSNewProvider sets the new Ceph module VFS based provider.
	SambaVFSNewProvider = CephFSProvider("samba-vfs/new")
	// SambaVFSClassicProvider sets the older Ceph module VFS based provider.
	SambaVFSClassicProvider = CephFSProvider("samba-vfs/classic")
	// SambaVFSProxiedProvider sets the new Ceph module VFS based provider with CephFS proxy server support.
	SambaVFSProxiedProvider = CephFSProvider("samba-vfs/proxied")
)

// PasswordFilter allows password values to be hidden or obfuscated when
// sent to or fetched from the smb module.
type PasswordFilter string

const (
	// PasswordFilterUnset specifies no password filter.
	PasswordFilterUnset = PasswordFilter("")
	// PasswordFilterNone specifies no password filtering should be done.
	PasswordFilterNone = PasswordFilter("none")
	// PasswordFilterBase64 specifies passwords should be converted from/to
	// base64 encoding.
	PasswordFilterBase64 = PasswordFilter("base64")
	// PasswordFilterHidden specifies passwords should be replaced by opaque
	// placeholder values.
	PasswordFilterHidden = PasswordFilter("hidden")
)

// Service names particular network services provided by an ceph smb cluster.
type Service string

const (
	// SMBService represents the core smb network file system service.
	SMBService = Service("smb")
	// SMBMetricsService represents the prometheus style metrics service.
	SMBMetricsService = Service("smbmetrics")
	// CTDBService represents the ctdb service used to coordinate clusters.
	CTDBService = Service("ctdb")
	// RemoteControlService represents a cloud compatible remote control
	// service (based on gRPC).
	RemoteControlService = Service("remote-control")
)
