//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

// TLSCredentialSource identifies a TLS Credential resource that will be
// used as a source for TLS-based connection security for a service
// used for-or-by the smb cluster.
type TLSCredentialSource struct {
	SourceType SourceType `json:"source_type"`
	Ref        string     `json:"ref"`
}
