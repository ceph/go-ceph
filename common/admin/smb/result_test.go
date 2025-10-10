//go:build !(octopus || pacific || quincy || reef || squid)

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var resultSampleBad1 = `
{
  "resource": "nope",
  "fate": "present",
  "very": false
}
`

func TestResultInvalidJSON(t *testing.T) {
	result := &Result{}
	err := json.Unmarshal([]byte(resultSampleBad1), result)
	assert.Error(t, err)
}

var resultSample1 = `
{
  "resource": {
    "resource_type": "ceph.smb.share",
    "cluster_id": "bz1",
    "share_id": "bt4",
    "intent": "present",
    "name": "bt4",
    "readonly": false,
    "browseable": true,
    "cephfs": {
      "volume": "cephfs",
      "path": "/",
      "subvolumegroup": "g1",
      "subvolume": "t2",
      "provider": "samba-vfs"
    }
  },
  "state": "present",
  "success": true
}
`

func TestResultUnmarshal(t *testing.T) {
	result := &Result{}
	err := json.Unmarshal([]byte(resultSample1), result)
	assert.NoError(t, err)

	assert.True(t, result.Ok())
	assert.Equal(t, result.State(), "present")
	assert.Equal(t, result.Resource().Type(), ShareType)
	assert.Equal(t, result.Message(), "")
	assert.Equal(t, result.Error(), "")
}

var resultSample2 = `
{
  "resource": {
    "resource_type": "ceph.smb.share",
    "cluster_id": "nevah1",
    "share_id": "bt4",
    "intent": "present",
    "name": "bt4",
    "readonly": false,
    "browseable": true,
    "cephfs": {
      "volume": "cephfs",
      "path": "/",
      "subvolumegroup": "g1",
      "subvolume": "t2",
      "provider": "samba-vfs"
    }
  },
  "cluster_id": "nevah1",
  "msg": "no matching cluster id",
  "success": false
}
`

func TestResultUnmarshal2(t *testing.T) {
	result := &Result{}
	err := json.Unmarshal([]byte(resultSample2), result)
	assert.NoError(t, err)

	assert.False(t, result.Ok())
	assert.Equal(t, result.State(), "")
	assert.Equal(t, result.Resource().Type(), ShareType)
	assert.Equal(t, result.Message(), "no matching cluster id")

	var e error = result
	assert.ErrorContains(t, e, "no matching")
	if assert.Contains(t, result.Dump(), "cluster_id") {
		assert.Equal(t, result.Dump()["cluster_id"], "nevah1")
	}
}

var resultGroupSample1 = `
{
  "results": [
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "nevah1",
        "share_id": "bt4",
        "intent": "present",
        "name": "bt4",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/",
          "subvolumegroup": "g1",
          "subvolume": "t2",
          "provider": "samba-vfs"
        }
      },
      "cluster_id": "nevah1",
      "msg": "no matching cluster id",
      "success": false
    }
  ],
  "success": false
}
`

func TestResultGroupUnmashal(t *testing.T) {
	rgroup := ResultGroup{}
	err := json.Unmarshal([]byte(resultGroupSample1), &rgroup)
	assert.NoError(t, err)

	assert.False(t, rgroup.Ok())
	assert.Len(t, rgroup.Results, 1)
	assert.Len(t, rgroup.ErrorResults(), 1)

	var e error = rgroup
	assert.ErrorContains(t, e, "1 resource")
}

var resultGroupSample2 = `
{
  "results": [
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "bz1",
        "share_id": "bt4",
        "intent": "present",
        "name": "bt4",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/",
          "subvolumegroup": "g1",
          "subvolume": "t2",
          "provider": "samba-vfs"
        }
      },
      "state": "present",
      "success": true
    }
  ],
  "success": true
}
`

func TestResultGroupUnmashal2(t *testing.T) {
	rgroup := ResultGroup{}
	err := json.Unmarshal([]byte(resultGroupSample2), &rgroup)
	assert.NoError(t, err)

	assert.True(t, rgroup.Ok())
	assert.Len(t, rgroup.Results, 1)
	assert.Len(t, rgroup.ErrorResults(), 0)

	assert.Equal(t, rgroup.Error(), "")
}

var resultGroupSample3 = `
{
  "results": [
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "cluster1",
        "share_id": "share1",
        "intent": "present",
        "name": "Share One",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/",
          "subvolumegroup": "g1",
          "subvolume": "sv1",
          "provider": "samba-vfs"
        }
      },
      "state": "created",
      "success": true
    },
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "cluster1",
        "share_id": "share2",
        "intent": "present",
        "name": "Share Two",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/",
          "subvolumegroup": "g1",
          "subvolume": "sv2",
          "provider": "samba-vfs"
        }
      },
      "state": "created",
      "success": true
    },
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "cluster1",
        "share_id": "share3",
        "intent": "present",
        "name": "Share Three",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/",
          "subvolumegroup": "g1",
          "subvolume": "sv3",
          "provider": "samba-vfs"
        }
      },
      "state": "created",
      "success": true
    }
  ],
  "success": true
}
`

func TestResultGroupUnmashal3(t *testing.T) {
	rgroup := ResultGroup{}
	err := json.Unmarshal([]byte(resultGroupSample3), &rgroup)
	assert.NoError(t, err)

	assert.True(t, rgroup.Ok())
	assert.Len(t, rgroup.Results, 3)
	assert.Len(t, rgroup.ErrorResults(), 0)
	assert.EqualValues(t,
		rgroup.Results[0].Resource().Identity().String(),
		"ceph.smb.share.cluster1.share1")
	assert.EqualValues(t, rgroup.Results[0].State(), "created")

	assert.Equal(t, rgroup.Error(), "")
}

var resultGroupSample4 = `
{
  "results": [
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "cluster1",
        "share_id": "share1",
        "intent": "present",
        "name": "Share One",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/zim",
          "subvolumegroup": "g1",
          "subvolume": "sv1",
          "provider": "samba-vfs"
        }
      },
      "msg": "path is not a valid directory in volume",
      "success": false
    },
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "cluster1",
        "share_id": "share2",
        "intent": "present",
        "name": "Share Two",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/zam",
          "subvolumegroup": "g1",
          "subvolume": "sv2",
          "provider": "samba-vfs"
        }
      },
      "msg": "path is not a valid directory in volume",
      "success": false
    },
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "cluster1",
        "share_id": "share3",
        "intent": "present",
        "name": "Share Three",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/",
          "subvolumegroup": "g1",
          "subvolume": "sv3",
          "provider": "samba-vfs"
        }
      },
      "checked": true,
      "success": true
    }
  ],
  "success": false
}
`

func TestResultGroupUnmashal4(t *testing.T) {
	rgroup := ResultGroup{}
	err := json.Unmarshal([]byte(resultGroupSample4), &rgroup)
	assert.NoError(t, err)

	assert.False(t, rgroup.Ok())
	assert.Len(t, rgroup.Results, 3)
	assert.Len(t, rgroup.ErrorResults(), 2)
	tr := rgroup.ErrorResults()[0]
	assert.EqualValues(t,
		tr.Resource().Identity().String(),
		"ceph.smb.share.cluster1.share1")
	assert.EqualValues(t, tr.State(), "")
	assert.EqualValues(t, tr.Message(), "path is not a valid directory in volume")

	assert.Contains(t, rgroup.Error(), "2 resource errors:")
}
