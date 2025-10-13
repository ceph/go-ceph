//go:build !(octopus || pacific || quincy || reef || squid)

package smb

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
