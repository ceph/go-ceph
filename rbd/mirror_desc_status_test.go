//go:build !nautilus && ceph_preview
// +build !nautilus,ceph_preview

// IMPORTANT - when removing ceph_preview from this file also delete
// rbd/mirror_stub_test.go as it will no longer serve a purpose.

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMirrorDescriptionJSON(t *testing.T) {
	cases := []struct {
		desc string
		ok   bool
	}{
		{
			desc: "",
			ok:   false,
		},
		{
			desc: "local image is primary",
			ok:   false,
		},
		{
			desc: "status not found",
			ok:   false,
		},
		{
			desc: "invalid {",
			ok:   false,
		},
		{
			desc: "} invalid {",
			ok:   false,
		},
		{
			desc: "phony, {}",
			ok:   true,
		},
		{
			desc: `replaying, {"bytes_per_second":0.0,"bytes_per_snapshot":0.0,"last_snapshot_bytes":0,"last_snapshot_sync_seconds":0,"remote_snapshot_timestamp":1678125999,"replay_state":"idle"}`,
			ok:   true,
		},
		{
			desc: "invalid-json, {:::...!}",
			ok:   false,
		},
	}
	for _, tcase := range cases {
		t.Run("testParse", func(t *testing.T) {
			var data map[string]interface{}
			s := SiteMirrorImageStatus{
				Description: tcase.desc,
			}
			err := s.UnmarshalDescriptionJSON(&data)
			if tcase.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMirrorDescriptionReplayStatus(t *testing.T) {
	cases := []struct {
		desc     string
		ok       bool
		expected MirrorDescriptionReplayStatus
	}{
		{
			desc: "",
			ok:   false,
		},
		{
			desc: "local image is primary",
			ok:   false,
		},
		{
			desc: "status not found",
			ok:   false,
		},
		{
			desc: "invalid {",
			ok:   false,
		},
		{
			desc: "} invalid {",
			ok:   false,
		},
		{
			desc: "phony, {}",
			ok:   true,
		},
		{
			desc: `replaying, {"bytes_per_second":0.0,"bytes_per_snapshot":0.0,"last_snapshot_bytes":0,"last_snapshot_sync_seconds":0,"remote_snapshot_timestamp":1678125999,"replay_state":"idle"}`,
			ok:   true,
			expected: MirrorDescriptionReplayStatus{
				ReplayState:             "idle",
				RemoteSnapshotTimestamp: 1678125999,
			},
		},
		{
			desc: `replaying, {"bytes_per_second":446028.87,"bytes_per_snapshot":559983.04,"last_snapshot_bytes":4030,"last_snapshot_sync_seconds":9087,"remote_snapshot_timestamp":1678125999,"replay_state":"syncing"}`,
			ok:   true,
			expected: MirrorDescriptionReplayStatus{
				ReplayState:             "syncing",
				RemoteSnapshotTimestamp: 1678125999,
				BytesPerSecond:          446028.87,
				BytesPerSnapshot:        559983.04,
				LastSnapshotSyncSeconds: 9087,
				LastSnapshotBytes:       4030,
			},
		},
		{
			desc: `something-or-other, {"bytes_per_second":446028.87,"bytes_per_snapshot":559983.04,"last_snapshot_bytes":4030,"last_snapshot_sync_seconds":9087,"remote_snapshot_timestamp":1678125999,"replay_state":"syncing","local_snapshot_timestamp":1674425567,"syncing_snapshot_timestamp":1674325567,"syncing_percent":31}`,
			ok:   true,
			expected: MirrorDescriptionReplayStatus{
				ReplayState:              "syncing",
				RemoteSnapshotTimestamp:  1678125999,
				LocalSnapshotTimestamp:   1674425567,
				SyncingSnapshotTimestamp: 1674325567,
				SyncingPercent:           31,
				BytesPerSecond:           446028.87,
				BytesPerSnapshot:         559983.04,
				LastSnapshotSyncSeconds:  9087,
				LastSnapshotBytes:        4030,
			},
		},
	}
	for _, tcase := range cases {
		t.Run("testCase", func(t *testing.T) {
			s := SiteMirrorImageStatus{
				Description: tcase.desc,
			}
			rs, err := s.DescriptionReplayStatus()
			if tcase.ok {
				assert.NoError(t, err)
				assert.EqualValues(t, tcase.expected, *rs)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// testDescriptionReplayStatus is a function that exists only to be compiled on
// ceph_preview builds so that we do not need to reimplement the bulk of the
// mirroring tests to check the functionality of our new preview funcs.
func testDescriptionReplayStatus(t *testing.T, smis SiteMirrorImageStatus) {
	t.Log("testing DescriptionReplayStatus")

	rsts, err := smis.DescriptionReplayStatus()
	if assert.NoError(t, err) {
		assert.Subset(t, []string{"idle", "syncing"}, []string{rsts.ReplayState})
		// timestamp is approx. a year in the past so unless your
		// clock is really messed up, this should pass
		assert.Greater(t, rsts.RemoteSnapshotTimestamp, int64(1646593940))
	}
}
