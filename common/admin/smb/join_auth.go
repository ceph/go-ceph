//go:build !(octopus || pacific || quincy || reef || squid)

package smb

import (
	"encoding/json"
	"fmt"
)

// JoinAuthValues contains the username and password an SMB server will
// use to join a domain.
type JoinAuthValues struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JoinAuth is a resource containing the parameters needed to join an SMB
// server to a domain.
type JoinAuth struct {
	IntentValue     Intent          `json:"intent"`
	AuthID          string          `json:"auth_id"`
	Auth            *JoinAuthValues `json:"auth,omitempty"`
	LinkedToCluster string          `json:"linked_to_cluster,omitempty"`
}

// Type returns a ResourceType value.
func (*JoinAuth) Type() ResourceType {
	return JoinAuthType
}

// Intent controls if a resource is to be created/updated or removed.
func (ja *JoinAuth) Intent() Intent {
	return ja.IntentValue
}

// Identity returns a ResourceRef identifying this joinauth resource.
func (ja *JoinAuth) Identity() ResourceRef {
	return ResourceID{
		ResourceType: ja.Type(),
		ID:           ja.AuthID,
	}
}

// Validate returns an error describing an issue with the resource or nil if
// the resource object is valid.
func (ja *JoinAuth) Validate() error {
	var minimal bool
	switch ja.IntentValue {
	case Present:
	case Removed:
		minimal = true
	default:
		return fmt.Errorf("missing intent")
	}
	if ja.AuthID == "" {
		return fmt.Errorf("missing AuthID")
	}
	if minimal {
		return nil // minimal checks have been done, return early
	}

	if ja.Auth == nil {
		return fmt.Errorf("missing Auth parameters")
	}
	if ja.Auth.Username == "" {
		return fmt.Errorf("missing Username")
	}
	if ja.Auth.Password == "" {
		return fmt.Errorf("missing Password")
	}
	return nil
}

// MarshalJSON supports marshalling a cluster to JSON.
func (ja *JoinAuth) MarshalJSON() ([]byte, error) {
	type vJoinAuth JoinAuth
	type wJoinAuth struct {
		ResourceType ResourceType `json:"resource_type"`
		vJoinAuth
	}
	return json.Marshal(wJoinAuth{
		ResourceType: ja.Type(),
		vJoinAuth:    vJoinAuth(*ja),
	})
}

// SetAuth modifies a JoinAuth's authentication values.
func (ja *JoinAuth) SetAuth(un, pw string) *JoinAuth {
	if ja.Auth == nil {
		ja.Auth = &JoinAuthValues{}
	}
	ja.Auth.Username = un
	ja.Auth.Password = pw
	return ja
}

// NewJoinAuth returns a new JoinAuth with default values.
func NewJoinAuth(authID string) *JoinAuth {
	return &JoinAuth{
		IntentValue: Present,
		AuthID:      authID,
		Auth:        &JoinAuthValues{},
	}
}

// NewLinkedJoinAuth returns a new JoinAuth with default values that link the
// resource to a particular cluster. Linked resources can only be used by the
// cluster they link to and are automatically deleted when the linked cluster
// is deleted.
func NewLinkedJoinAuth(cluster *Cluster) *JoinAuth {
	ja := NewJoinAuth(randName(cluster.ClusterID))
	ja.LinkedToCluster = cluster.ClusterID
	cluster.DomainSettings.JoinSources = append(
		cluster.DomainSettings.JoinSources,
		JoinAuthSource{SourceType: ResourceSource, Ref: ja.AuthID},
	)
	return ja
}

// NewJoinAuthToRemove returns a new JoinAuth with default values set to remove
// the join auth resource from management.
func NewJoinAuthToRemove(authID string) *JoinAuth {
	return &JoinAuth{
		IntentValue: Removed,
		AuthID:      authID,
	}
}
