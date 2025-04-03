//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"fmt"
)

// CephFSSource defines parameters that connect an SMB Share to a path
// or subvolume in CephFS.
type CephFSSource struct {
	Volume         string         `json:"volume"`
	SubVolumeGroup string         `json:"subvolumegroup,omitempty"`
	SubVolume      string         `json:"subvolume,omitempty"`
	Path           string         `json:"path"`
	Provider       CephFSProvider `json:"provider,omitempty"`
}

// ShareAccess defines parameters that control the ability to log in to a share
// with particular access levels.
type ShareAccess struct {
	Name     string         `json:"name"`
	Category AccessCategory `json:"category"`
	Access   AccessMode     `json:"access"`
}

// Validate returns an error describing an issue with the share access object
// or nil if the object is valid.
func (sa *ShareAccess) Validate() error {
	if sa.Name == "" {
		return fmt.Errorf("missing Name")
	}
	switch sa.Category {
	case "", UserAccess, GroupAccess:
	default:
		return fmt.Errorf("invalid Category")
	}
	switch sa.Access {
	case "", ReadAccess, ReadWriteAccess, AdminAccess:
	default:
		return fmt.Errorf("invalid Access mode")
	}
	return nil
}

// Share is a resource representing SMB Shares that will be configured on
// the SMB servers hosted in the Ceph cluster.
type Share struct {
	IntentValue    Intent        `json:"intent"`
	ClusterID      string        `json:"cluster_id"`
	ShareID        string        `json:"share_id"`
	Name           string        `json:"name"`
	ReadOnly       bool          `json:"readonly"`
	Browseable     bool          `json:"browseable"`
	CephFS         *CephFSSource `json:"cephfs,omitempty"`
	RestrictAccess bool          `json:"restrict_access"`
	LoginControl   []ShareAccess `json:"login_control,omitempty"`
}

// Type returns a ResourceType value.
func (*Share) Type() ResourceType {
	return ShareType
}

// Intent controls if a resource is to be created/updated or removed.
func (share *Share) Intent() Intent {
	return share.IntentValue
}

// Identity returns a ResourceRef identifying this share resource.
func (share *Share) Identity() ResourceRef {
	return ChildResourceID{
		ResourceType: share.Type(),
		ParentID:     share.ClusterID,
		ID:           share.ShareID,
	}
}

// MarshalJSON supports marshalling a Share resource to JSON.
func (share *Share) MarshalJSON() ([]byte, error) {
	type vShare Share
	type wShare struct {
		ResourceType ResourceType `json:"resource_type"`
		vShare
	}
	return json.Marshal(wShare{
		ResourceType: share.Type(),
		vShare:       vShare(*share),
	})
}

// Validate returns an error describing an issue with the resource or nil if
// the resource object is valid.
func (share *Share) Validate() error {
	var minimal bool
	switch share.IntentValue {
	case Present:
	case Removed:
		minimal = true
	default:
		return fmt.Errorf("missing IntentValue")
	}
	if share.ClusterID == "" {
		return fmt.Errorf("missing ClusterID")
	}
	if share.ShareID == "" {
		return fmt.Errorf("missing ShareID")
	}
	if minimal {
		return nil // minimal checks have been done, return early
	}

	if share.CephFS == nil {
		return fmt.Errorf("missing CephFS configuration")
	}
	if share.CephFS.Volume == "" {
		return fmt.Errorf("missing CephFS Volume")
	}
	for _, sa := range share.LoginControl {
		if err := sa.Validate(); err != nil {
			return fmt.Errorf("invalid LoginControl value: %w", err)
		}
	}
	return nil
}

// SetCephFS modifies a Share resource's CephFS storage parameters.
func (share *Share) SetCephFS(
	volume, subvolumegroup, subvolume, path string) *Share {

	if share.CephFS == nil {
		share.CephFS = &CephFSSource{}
	}
	share.CephFS.Volume = volume
	share.CephFS.SubVolumeGroup = subvolumegroup
	share.CephFS.SubVolume = subvolume
	share.CephFS.Path = path
	return share
}

// NewShare returns a new Share resource object with default values.
func NewShare(clusterID, shareID string) *Share {
	return &Share{
		IntentValue: Present,
		ClusterID:   clusterID,
		ShareID:     shareID,
	}
}

// NewShareToRemove returns a new Share resource object with default values set
// to remove the share from management.
func NewShareToRemove(clusterID, shareID string) *Share {
	return &Share{
		IntentValue: Removed,
		ClusterID:   clusterID,
		ShareID:     shareID,
	}
}
