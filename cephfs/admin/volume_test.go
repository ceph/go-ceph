//go:build !luminous && !mimic
// +build !luminous,!mimic

package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListVolumes(t *testing.T) {
	fsa := getFSAdmin(t)

	vl, err := fsa.ListVolumes()
	assert.NoError(t, err)
	assert.Len(t, vl, 1)
	assert.Equal(t, "cephfs", vl[0])
}

func TestEnumerateVolumes(t *testing.T) {
	fsa := getFSAdmin(t)

	ve, err := fsa.EnumerateVolumes()
	assert.NoError(t, err)
	if assert.Len(t, ve, 1) {
		assert.Equal(t, "cephfs", ve[0].Name)
		assert.Equal(t, int64(1), ve[0].ID)
	}
}

// note: some of these dumps are simplified for testing purposes if we add
// general dump support these samples may need to be expanded upon.
var sampleDump1 = []byte(`
{
  "epoch": 5,
  "default_fscid": 1,
  "filesystems": [
    {
      "mdsmap": {
        "epoch": 5,
        "flags": 18,
        "ever_allowed_features": 0,
        "explicitly_allowed_features": 0,
        "created": "2020-08-31T18:37:34.657633+0000",
        "modified": "2020-08-31T18:37:36.700989+0000",
        "tableserver": 0,
        "root": 0,
        "session_timeout": 60,
        "session_autoclose": 300,
        "min_compat_client": "0 (unknown)",
        "max_file_size": 1099511627776,
        "last_failure": 0,
        "last_failure_osd_epoch": 0,
        "compat": {
          "compat": {},
          "ro_compat": {},
          "incompat": {
            "feature_1": "base v0.20",
            "feature_2": "client writeable ranges",
            "feature_3": "default file layouts on dirs",
            "feature_4": "dir inode in separate object",
            "feature_5": "mds uses versioned encoding",
            "feature_6": "dirfrag is stored in omap",
            "feature_8": "no anchor table",
            "feature_9": "file layout v2",
            "feature_10": "snaprealm v2"
          }
        },
        "max_mds": 1,
        "in": [
          0
        ],
        "up": {
          "mds_0": 4115
        },
        "failed": [],
        "damaged": [],
        "stopped": [],
        "info": {
          "gid_4115": {
            "gid": 4115,
            "name": "Z",
            "rank": 0,
            "incarnation": 4,
            "state": "up:active",
            "state_seq": 2,
            "addr": "127.0.0.1:6809/2568111595",
            "addrs": {
              "addrvec": [
                {
                  "type": "v1",
                  "addr": "127.0.0.1:6809",
                  "nonce": 2568111595
                }
              ]
            },
            "join_fscid": -1,
            "export_targets": [],
            "features": 4540138292836696000,
            "flags": 0
          }
        },
        "data_pools": [
          1
        ],
        "metadata_pool": 2,
        "enabled": true,
        "fs_name": "cephfs",
        "balancer": "",
        "standby_count_wanted": 0
      },
      "id": 1
    }
  ]
}
`)

var sampleDump2 = []byte(`
{
  "epoch": 5,
  "default_fscid": 1,
  "filesystems": [
    {
      "mdsmap": {
        "fs_name": "wiffleball",
        "standby_count_wanted": 0
      },
      "id": 1
    },
    {
      "mdsmap": {
        "fs_name": "beanbag",
        "standby_count_wanted": 0
      },
      "id": 2
    }
  ]
}
`)

func TestParseDumpToIdents(t *testing.T) {
	R := newResponse
	fakePrefix := dumpOkPrefix + " 5"
	t.Run("error", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(nil, "", errors.New("boop")))
		assert.Error(t, err)
		assert.Equal(t, "boop", err.Error())
		assert.Nil(t, idents)
	})
	t.Run("badStatus", func(t *testing.T) {
		_, err := parseDumpToIdents(R(sampleDump1, "unexpected!", nil))
		assert.Error(t, err)
	})
	t.Run("oneVolOk", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(sampleDump1, fakePrefix, nil))
		assert.NoError(t, err)
		if assert.Len(t, idents, 1) {
			assert.Equal(t, "cephfs", idents[0].Name)
			assert.Equal(t, int64(1), idents[0].ID)
		}
	})
	t.Run("twoVolOk", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(sampleDump2, fakePrefix, nil))
		assert.NoError(t, err)
		if assert.Len(t, idents, 2) {
			assert.Equal(t, "wiffleball", idents[0].Name)
			assert.Equal(t, int64(1), idents[0].ID)
			assert.Equal(t, "beanbag", idents[1].Name)
			assert.Equal(t, int64(2), idents[1].ID)
		}
	})
	t.Run("unexpectedStatus", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(sampleDump1, "slip-up", nil))
		assert.Error(t, err)
		assert.Nil(t, idents)
	})
}

func TestVolumeStatus(t *testing.T) {
	if serverVersion == cephNautilus {
		t.Skipf("can only execute on octopus/pacific servers")
	}
	fsa := getFSAdmin(t)

	vs, err := fsa.VolumeStatus("cephfs")
	assert.NoError(t, err)
	assert.Contains(t, vs.MDSVersion, "version")
}

func TestVolumeStatusInvalid(t *testing.T) {
	if serverVersion != cephNautilus {
		t.Skipf("can only excecute on nautilus servers")
	}
	fsa := getFSAdmin(t)

	vs, err := fsa.VolumeStatus("cephfs")
	assert.Error(t, err)
	assert.Nil(t, vs)
	var notImpl NotImplementedError
	assert.True(t, errors.As(err, &notImpl))
}

var sampleVolumeStatus1 = []byte(`
{
"clients": [{"clients": 1, "fs": "cephfs"}],
"mds_version": "ceph version 15.2.4 (7447c15c6ff58d7fce91843b705a268a1917325c) octopus (stable)",
"mdsmap": [{"dns": 76, "inos": 19, "name": "Z", "rank": 0, "rate": 0.0, "state": "active"}],
"pools": [{"avail": 1017799872, "id": 2, "name": "cephfs_metadata", "type": "metadata", "used": 2204126}, {"avail": 1017799872, "id": 1, "name": "cephfs_data", "type": "data", "used": 0}]
}
`)

var sampleVolumeStatusTextJunk = []byte(`cephfs - 2 clients
======
+------+--------+-----+---------------+-------+-------+
| Rank | State  | MDS |    Activity   |  dns  |  inos |
+------+--------+-----+---------------+-------+-------+
|  0   | active |  Z  | Reqs:   98 /s |  254  |  192  |
+------+--------+-----+---------------+-------+-------+
+-----------------+----------+-------+-------+
|       Pool      |   type   |  used | avail |
+-----------------+----------+-------+-------+
| cephfs_metadata | metadata | 62.1M |  910M |
|   cephfs_data   |   data   |    0  |  910M |
+-----------------+----------+-------+-------+
+-------------+
| Standby MDS |
+-------------+
+-------------+
MDS version: ceph version 14.2.11 (f7fdb2f52131f54b891a2ec99d8205561242cdaf) nautilus (stable)
`)

func TestParseVolumeStatus(t *testing.T) {
	R := newResponse
	t.Run("error", func(t *testing.T) {
		_, err := parseVolumeStatus(R(nil, "", errors.New("bonk")))
		assert.Error(t, err)
		assert.Equal(t, "bonk", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseVolumeStatus(R(nil, "unexpected!", nil))
		assert.Error(t, err)
	})
	t.Run("badJSON", func(t *testing.T) {
		_, err := parseVolumeStatus(R([]byte("_XxXxX"), "", nil))
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		s, err := parseVolumeStatus(R(sampleVolumeStatus1, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, s) {
			assert.Contains(t, s.MDSVersion, "ceph version 15.2.4")
			assert.Contains(t, s.MDSVersion, "octopus")
		}
	})
	t.Run("notJSONfromServer", func(t *testing.T) {
		_, err := parseVolumeStatus(R(sampleVolumeStatusTextJunk, "", nil))
		if assert.Error(t, err) {
			var notImpl NotImplementedError
			assert.True(t, errors.As(err, &notImpl))
		}
	})

}

var sampleFsLs1 = []byte(`
[
  {
    "name": "cephfs",
    "metadata_pool": "cephfs_metadata",
    "metadata_pool_id": 2,
    "data_pool_ids": [
      1
    ],
    "data_pools": [
      "cephfs_data"
    ]
  }
]
`)

var sampleFsLs2 = []byte(`
[
  {
    "name": "cephfs",
    "metadata_pool": "cephfs_metadata",
    "metadata_pool_id": 2,
    "data_pool_ids": [
      1
    ],
    "data_pools": [
      "cephfs_data"
    ]
  },
  {
    "name": "archivefs",
    "metadata_pool": "archivefs_metadata",
    "metadata_pool_id": 6,
    "data_pool_ids": [
      4,
      5
    ],
    "data_pools": [
      "archivefs_data1",
      "archivefs_data2"
    ]
  }
]
`)

func TestParseFsList(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		_, err := parseFsList(
			newResponse(nil, "", errors.New("eek")))
		assert.Error(t, err)
		assert.Equal(t, "eek", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseFsList(
			newResponse(nil, "oof", nil))
		assert.Error(t, err)
	})
	t.Run("badJSON", func(t *testing.T) {
		_, err := parseFsList(
			newResponse([]byte("______"), "", nil))
		assert.Error(t, err)
	})
	t.Run("ok1", func(t *testing.T) {
		l, err := parseFsList(
			newResponse(sampleFsLs1, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, l) && assert.Len(t, l, 1) {
			fs := l[0]
			assert.Equal(t, "cephfs", fs.Name)
			assert.Equal(t, "cephfs_metadata", fs.MetadataPool)
			assert.Equal(t, 2, fs.MetadataPoolID)
			assert.Len(t, fs.DataPools, 1)
			assert.Contains(t, fs.DataPools, "cephfs_data")
			assert.Len(t, fs.DataPoolIDs, 1)
			assert.Contains(t, fs.DataPoolIDs, 1)
		}
	})
	t.Run("ok2", func(t *testing.T) {
		l, err := parseFsList(
			newResponse(sampleFsLs2, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, l) && assert.Len(t, l, 2) {
			fs := l[0]
			assert.Equal(t, "cephfs", fs.Name)
			assert.Equal(t, "cephfs_metadata", fs.MetadataPool)
			assert.Equal(t, 2, fs.MetadataPoolID)
			assert.Len(t, fs.DataPools, 1)
			assert.Contains(t, fs.DataPools, "cephfs_data")
			assert.Len(t, fs.DataPoolIDs, 1)
			assert.Contains(t, fs.DataPoolIDs, 1)
			fs = l[1]
			assert.Equal(t, "archivefs", fs.Name)
			assert.Equal(t, "archivefs_metadata", fs.MetadataPool)
			assert.Equal(t, 6, fs.MetadataPoolID)
			assert.Len(t, fs.DataPools, 2)
			assert.Contains(t, fs.DataPools, "archivefs_data1")
			assert.Contains(t, fs.DataPools, "archivefs_data2")
			assert.Len(t, fs.DataPoolIDs, 2)
			assert.Contains(t, fs.DataPoolIDs, 4)
			assert.Contains(t, fs.DataPoolIDs, 5)
		}
	})
}

func TestListFileSystems(t *testing.T) {
	fsa := getFSAdmin(t)

	l, err := fsa.ListFileSystems()
	assert.NoError(t, err)
	if assert.Len(t, l, 1) {
		assert.Equal(t, "cephfs", l[0].Name)
		assert.Equal(t, "cephfs_metadata", l[0].MetadataPool)
		assert.Len(t, l[0].DataPools, 1)
		assert.Contains(t, l[0].DataPools, "cephfs_data")
	}
}
