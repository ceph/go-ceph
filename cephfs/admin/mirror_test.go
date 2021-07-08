package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sampleOldDaemonStatus1 and sampleOldDaemonStatus2 are examples of the json
// returned for daemon status before Ceph v16.2.5. It is not used by the tests
// as the code now only supports the format of v16.2.5 and later. This is
// retained for reference and the off chance that someone asks go-ceph to
// support the older format.

var sampleOldDaemonStatus1 = `
{"4157": {"1": {"name": "cephfs", "directory_count": 0, "peers": {}}}}
`

var sampleOldDaemonStatus2 = `
{
  "4154": {
    "1": {
      "name": "cephfs",
      "directory_count": 1,
      "peers": {
        "d284fccd-6110-4e94-843c-78ecda3aef38": {
          "remote": {"client_name": "client.mirror_remote", "cluster_name": "ceph_b", "fs_name": "cephfs"},
          "stats": {"failure_count": 1, "recovery_count": 0}
        }
      }
    }
  }
}
`

var sampleDaemonStatus1 = `
[
  {
    "daemon_id": 4115,
    "filesystems": [
      {
        "filesystem_id": 1,
        "name": "cephfs",
        "directory_count": 0,
        "peers": []
      }
    ]
  }
]
`

var sampleDaemonStatus2 = `
[
  {
    "daemon_id": 4143,
    "filesystems": [
      {
        "filesystem_id": 1,
        "name": "cephfs",
        "directory_count": 1,
        "peers": [
          {
            "uuid": "43c50942-9dba-4f66-8f9b-102378fa863e",
            "remote": {
              "client_name": "client.mirror_remote",
              "cluster_name": "ceph_b",
              "fs_name": "cephfs"
            },
            "stats": {
              "failure_count": 1,
              "recovery_count": 0
            }
          }
        ]
      }
    ]
  }
]
`

var samplePeerList1 = `
{
  "f138660d-7b22-4036-95ba-0fda727bff40": {
    "client_name": "client.mirror_remote",
    "site_name": "ceph_b",
    "fs_name": "cephfs",
    "mon_host": "test_ceph_b"
  }
}
`

func TestParseDaemonStatus(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		r := newResponse(nil, "", errors.New("snark"))
		_, err := parseDaemonStatus(r)
		assert.Error(t, err)
		assert.Equal(t, "snark", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		r := newResponse(nil, "oopsie", nil)
		_, err := parseDaemonStatus(r)
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		r := newResponse([]byte(sampleDaemonStatus1), "", nil)
		ds, err := parseDaemonStatus(r)
		assert.NoError(t, err)
		if assert.NotNil(t, ds) && assert.Len(t, ds, 1) {
			dsi := ds[0]
			assert.Equal(t, DaemonID(4115), dsi.DaemonID)
			if assert.Len(t, dsi.FileSystems, 1) {
				fs := dsi.FileSystems[0]
				assert.Len(t, fs.Peers, 0)
				assert.Equal(t, "cephfs", fs.Name)
				assert.Equal(t, int64(0), fs.DirectoryCount)
			}
		}
	})
	t.Run("ok2", func(t *testing.T) {
		r := newResponse([]byte(sampleDaemonStatus2), "", nil)
		ds, err := parseDaemonStatus(r)
		assert.NoError(t, err)
		if assert.NotNil(t, ds) && assert.Len(t, ds, 1) {
			dsi := ds[0]
			assert.Equal(t, DaemonID(4143), dsi.DaemonID)
			if assert.Len(t, dsi.FileSystems, 1) {
				fs := dsi.FileSystems[0]
				assert.Equal(t, "cephfs", fs.Name)
				assert.Equal(t, int64(1), fs.DirectoryCount)
				if assert.Len(t, fs.Peers, 1) {
					p := fs.Peers[0]
					assert.Equal(t, "ceph_b", p.Remote.ClusterName)
					assert.Equal(t, "cephfs", p.Remote.FSName)
					assert.Equal(t, uint64(1), p.Stats.FailureCount)
					assert.Equal(t, uint64(0), p.Stats.RecoveryCount)
				}
			}
		}
	})
}

func TestParsePeerList(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		r := newResponse(nil, "", errors.New("snark"))
		_, err := parsePeerList(r)
		assert.Error(t, err)
		assert.Equal(t, "snark", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		r := newResponse(nil, "oopsie", nil)
		_, err := parsePeerList(r)
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		r := newResponse([]byte(samplePeerList1), "", nil)
		plr, err := parsePeerList(r)
		assert.NoError(t, err)
		if assert.NotNil(t, plr) && assert.Len(t, plr, 1) {
			p, ok := plr[PeerUUID("f138660d-7b22-4036-95ba-0fda727bff40")]
			assert.True(t, ok)
			assert.Equal(t, "client.mirror_remote", p.ClientName)
			assert.Equal(t, "ceph_b", p.SiteName)
			assert.Equal(t, "cephfs", p.FSName)
			assert.Equal(t, "test_ceph_b", p.MonHost)
		}
	})
}
