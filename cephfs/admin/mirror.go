package admin

import (
	ccom "github.com/ceph/go-ceph/common/commands"
	"github.com/ceph/go-ceph/internal/commands"
)

// SnapshotMirrorAdmin helps administer the snapshot mirroring features of
// cephfs. Snapshot mirroring is only available in ceph pacific and later.
type SnapshotMirrorAdmin struct {
	conn ccom.MgrCommander
}

// SnapshotMirror returns a new SnapshotMirrorAdmin to be used for the
// administration of snapshot mirroring features.
func (fsa *FSAdmin) SnapshotMirror() *SnapshotMirrorAdmin {
	return &SnapshotMirrorAdmin{conn: fsa.conn}
}

// Enable snapshot mirroring for the given file system.
//
// Similar To:
//  ceph fs snapshot mirror enable <fs_name>
func (sma *SnapshotMirrorAdmin) Enable(fsname string) error {
	m := map[string]string{
		"prefix":  "fs snapshot mirror enable",
		"fs_name": fsname,
		"format":  "json",
	}
	return commands.MarshalMgrCommand(sma.conn, m).NoStatus().EmptyBody().End()
}

// Disable snapshot mirroring for the given file system.
//
// Similar To:
//  ceph fs snapshot mirror disable <fs_name>
func (sma *SnapshotMirrorAdmin) Disable(fsname string) error {
	m := map[string]string{
		"prefix":  "fs snapshot mirror disable",
		"fs_name": fsname,
		"format":  "json",
	}
	return commands.MarshalMgrCommand(sma.conn, m).NoStatus().EmptyBody().End()
}

// Add a path in the file system to be mirrored.
//
// Similar To:
//  ceph fs snapshot mirror add <fs_name> <path>
func (sma *SnapshotMirrorAdmin) Add(fsname, path string) error {
	m := map[string]string{
		"prefix":  "fs snapshot mirror add",
		"fs_name": fsname,
		"path":    path,
		"format":  "json",
	}
	return commands.MarshalMgrCommand(sma.conn, m).NoStatus().EmptyBody().End()
}

// Remove a path in the file system from mirroring.
//
// Similar To:
//  ceph fs snapshot mirror remove <fs_name> <path>
func (sma *SnapshotMirrorAdmin) Remove(fsname, path string) error {
	m := map[string]string{
		"prefix":  "fs snapshot mirror remove",
		"fs_name": fsname,
		"path":    path,
		"format":  "json",
	}
	return commands.MarshalMgrCommand(sma.conn, m).NoStatus().EmptyBody().End()
}

type bootstrapTokenResponse struct {
	Token string `json:"token"`
}

// CreatePeerBootstrapToken returns a token that can be used to create
// a peering association between this site an another site.
//
// Similar To:
//  ceph fs snapshot mirror peer_bootstrap create <fs_name> <client_entity> <site-name>
func (sma *SnapshotMirrorAdmin) CreatePeerBootstrapToken(
	fsname, client, site string) (string, error) {
	m := map[string]string{
		"prefix":      "fs snapshot mirror peer_bootstrap create",
		"fs_name":     fsname,
		"client_name": client,
		"format":      "json",
	}
	if site != "" {
		m["site_name"] = site
	}
	var bt bootstrapTokenResponse
	err := commands.MarshalMgrCommand(sma.conn, m).NoStatus().Unmarshal(&bt).End()
	return bt.Token, err
}

// ImportPeerBoostrapToken creates an association between another site, one
// that has provided a token, with the current site.
//
// Similar To:
//  ceph fs snapshot mirror peer_bootstrap import <fs_name> <token>
func (sma *SnapshotMirrorAdmin) ImportPeerBoostrapToken(fsname, token string) error {
	m := map[string]string{
		"prefix":  "fs snapshot mirror peer_bootstrap import",
		"fs_name": fsname,
		"token":   token,
		"format":  "json",
	}
	return commands.MarshalMgrCommand(sma.conn, m).NoStatus().EmptyBody().End()
}
