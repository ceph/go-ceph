//go:build ceph_main && ceph_preview

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
