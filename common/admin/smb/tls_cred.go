//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"fmt"
)

// TLSContent indicates the type of TLS file/data a resource contains.
type TLSContent string

const (
	// TLSCert indicates the resource contains a TLS certificate.
	TLSCert = TLSContent("cert")
	// TLSKey indicates the resource contains a TLS key.
	TLSKey = TLSContent("key")
	// TLSCACert indicates the resource contains a TLS CA certificate.
	TLSCACert = TLSContent("ca-cert")
)

// TLSCredential is a resource containing a TLS resource such as a
// certificate, key, or ca certificate.
// server to a domain.
type TLSCredential struct {
	IntentValue     Intent     `json:"intent"`
	TLSCredentialID string     `json:"tls_credential_id"`
	CredentialType  TLSContent `json:"credential_type"`
	Value           string     `json:"value,omitempty"`
	LinkedToCluster string     `json:"linked_to_cluster,omitempty"`
}

// Type returns a ResourceType value.
func (*TLSCredential) Type() ResourceType {
	return TLSCredentialType
}

// Intent controls if a resource is to be created/updated or removed.
func (tc *TLSCredential) Intent() Intent {
	return tc.IntentValue
}

// SetIntent updates the resource's intent value.
func (tc *TLSCredential) SetIntent(i Intent) {
	tc.IntentValue = i
}

// Identity returns a ResourceRef identifying this tls credential resource.
func (tc *TLSCredential) Identity() ResourceRef {
	return ResourceID{
		ResourceType: tc.Type(),
		ID:           tc.TLSCredentialID,
	}
}

// Validate returns an error describing an issue with the resource or nil if
// the resource object is valid.
func (tc *TLSCredential) Validate() error {
	var minimal bool
	switch tc.IntentValue {
	case Present:
	case Removed:
		minimal = true
	default:
		return fmt.Errorf("missing intent")
	}
	if tc.TLSCredentialID == "" {
		return fmt.Errorf("missing TLSCredentialID")
	}
	if minimal {
		return nil // minimal checks have been done, return early
	}

	switch tc.CredentialType {
	case TLSCert, TLSKey, TLSCACert:
	case TLSContent(""):
		return fmt.Errorf("missing CredentialType")
	default:
		return fmt.Errorf("invalid CredentialType")
	}
	if tc.Value == "" {
		return fmt.Errorf("missing Value")
	}
	return nil
}

// MarshalJSON supports marshalling a TLSCredential to JSON.
func (tc *TLSCredential) MarshalJSON() ([]byte, error) {
	type vTLSCredential TLSCredential
	type wTLSCredential struct {
		ResourceType ResourceType `json:"resource_type"`
		vTLSCredential
	}
	return json.Marshal(wTLSCredential{
		ResourceType:   tc.Type(),
		vTLSCredential: vTLSCredential(*tc),
	})
}

// Set modifies a TLSCredential's credential type and credential value.
func (tc *TLSCredential) Set(ctype TLSContent, value string) *TLSCredential {
	tc.CredentialType = ctype
	tc.Value = value
	return tc
}

// NewTLSCredential returns a new empty TLSCredential.
func NewTLSCredential(id string) *TLSCredential {
	return &TLSCredential{
		IntentValue:     Present,
		TLSCredentialID: id,
	}
}

// NewLinkedTLSCredential returns a new TLSCredentialID with default values
// that link the resource to a particular cluster. Linked resources can only
// be used by the cluster they link to and are automatically deleted when the
// linked cluster is deleted.
func NewLinkedTLSCredential(cluster *Cluster) *TLSCredential {
	tc := NewTLSCredential(randName(cluster.ClusterID))
	tc.LinkedToCluster = cluster.ClusterID
	return tc
}

// NewTLSCredentialToRemove returns a new TLSCredential with default values
// set to remove the credential resource from management.
func NewTLSCredentialToRemove(id string) *TLSCredential {
	return &TLSCredential{
		IntentValue:     Removed,
		TLSCredentialID: id,
	}
}

func init() {
	if resourceTypes == nil {
		resourceTypes = map[ResourceType]func() Resource{}
	}
	resourceTypes[TLSCredentialType] = func() Resource {
		return new(TLSCredential)
	}
}
