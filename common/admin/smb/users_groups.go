//go:build !(octopus || pacific || quincy || reef || squid)

package smb

import (
	"encoding/json"
	"fmt"
)

// UserInfo defines a user account managed by an SMB server instance.
type UserInfo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// GroupInfo defines a group managed by an SMB server instance.
type GroupInfo struct {
	Name string `json:"name"`
}

// UsersAndGroupsValues contains user and group definitions managed by
// an SMB server instance.
type UsersAndGroupsValues struct {
	Users  []UserInfo  `json:"users,omitempty"`
	Groups []GroupInfo `json:"groups,omitempty"`
}

// UsersAndGroups is a resource containing user and group definitions that
// are managed by an SMB server instances that do not use active directory
// domains.
type UsersAndGroups struct {
	IntentValue     Intent                `json:"intent"`
	UsersGroupsID   string                `json:"users_groups_id"`
	Values          *UsersAndGroupsValues `json:"values,omitempty"`
	LinkedToCluster string                `json:"linked_to_cluster,omitempty"`
}

// Type returns a ResourceType value.
func (*UsersAndGroups) Type() ResourceType {
	return UsersAndGroupsType
}

// Intent controls if a resource is to be created/updated or removed.
func (ug *UsersAndGroups) Intent() Intent {
	return ug.IntentValue
}

// Identity returns a ResourceRef identifying this users and groups resource.
func (ug *UsersAndGroups) Identity() ResourceRef {
	return ResourceID{
		ResourceType: ug.Type(),
		ID:           ug.UsersGroupsID,
	}
}

// Validate returns an error describing an issue with the resource or nil if
// the resource object is valid.
func (ug *UsersAndGroups) Validate() error {
	var minimal bool
	switch ug.IntentValue {
	case Present:
	case Removed:
		minimal = true
	default:
		return fmt.Errorf("missing intent")
	}
	if ug.UsersGroupsID == "" {
		return fmt.Errorf("missing UsersGroupsID")
	}
	if minimal {
		return nil // minimal checks have been done, return early
	}

	if ug.Values == nil {
		return fmt.Errorf("missing Values parameter")
	}
	if len(ug.Values.Users) < 1 {
		return fmt.Errorf("no Users defined")
	}
	return nil
}

// MarshalJSON supports marshalling a UsersAndGroups resource to JSON.
func (ug *UsersAndGroups) MarshalJSON() ([]byte, error) {
	type vUsersAndGroups UsersAndGroups
	type wUsersAndGroups struct {
		ResourceType ResourceType `json:"resource_type"`
		vUsersAndGroups
	}
	return json.Marshal(wUsersAndGroups{
		ResourceType:    ug.Type(),
		vUsersAndGroups: vUsersAndGroups(*ug),
	})
}

// SetValues modifies a UsersAndGroups resource's users list and groups list.
func (ug *UsersAndGroups) SetValues(
	users []UserInfo, groups []GroupInfo) *UsersAndGroups {

	if ug.Values == nil {
		ug.Values = &UsersAndGroupsValues{}
	}
	ug.Values.Users = users
	ug.Values.Groups = groups
	return ug
}

// NewUsersAndGroups returns a new UsersAndGroups resource object with default
// values.
func NewUsersAndGroups(ugID string) *UsersAndGroups {
	return &UsersAndGroups{
		IntentValue:   Present,
		UsersGroupsID: ugID,
		Values:        &UsersAndGroupsValues{},
	}
}

// NewLinkedUsersAndGroups returns a new UsersAndGroups resource object with
// default values that link the resource to a particular cluster. Linked
// resources can only be used by the cluster they link to and are automatically
// deleted when the linked cluster is deleted.
func NewLinkedUsersAndGroups(cluster *Cluster) *UsersAndGroups {
	ug := NewUsersAndGroups(randName(cluster.ClusterID))
	ug.LinkedToCluster = cluster.ClusterID
	cluster.UserGroupSettings = append(
		cluster.UserGroupSettings,
		UserGroupSource{SourceType: ResourceSource, Ref: ug.UsersGroupsID},
	)
	return ug
}

// NewUsersAndGroupsToRemove returns a new UsersAndGroups resource object with
// default values set to remove the users and groups resource from management.
func NewUsersAndGroupsToRemove(ugID string) *UsersAndGroups {
	return &UsersAndGroups{
		IntentValue:   Removed,
		UsersGroupsID: ugID,
	}
}
