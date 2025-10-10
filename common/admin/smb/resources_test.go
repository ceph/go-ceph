//go:build !(octopus || pacific || quincy || reef || squid)

package smb

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var smbSample1 = `
{
  "resources": [
    {
      "resource_type": "ceph.smb.cluster",
      "cluster_id": "cluster1",
      "auth_mode": "active-directory",
      "domain_settings": {
        "realm": "DOMAIN1.SINK.TEST",
        "join_sources": [
          {"source_type": "resource", "ref": "join1-admin"}
        ]
      },
      "custom_dns": ["192.168.76.210"],
      "placement": {
        "count": 1,
        "label": "ilovesmb"
      }
    },
    {
      "resource_type": "ceph.smb.join.auth",
      "cluster_id": "join1-admin",
      "auth": {
        "username": "Administrator",
        "password": "Passw0rd"
      }
    },
    {
      "resource_type": "ceph.smb.share",
      "cluster_id": "cluster1",
      "share_id": "share1",
      "cephfs": {
        "volume": "cephfs",
        "path": "/"
      }
    },
    {
      "resource_type": "ceph.smb.share",
      "cluster_id": "cluster1",
      "share_id": "share2",
      "name": "Sub Volume",
      "cephfs": {
        "volume": "cephfs",
        "subvolumegroup": "g1",
        "subvolume": "sv1",
        "path": "/"
      }
    }
  ]
}
`

func TestUnmarshalResources(t *testing.T) {
	g := new(resourceGroup)
	b := []byte(smbSample1)
	assert.NoError(t, json.Unmarshal(b, g))
	assert.Len(t, g.Resources, 4)
	_, ok := g.Resources[0].(*Cluster)
	assert.True(t, ok)
	_, ok = g.Resources[2].(*Share)
	assert.True(t, ok)
	_, ok = g.Resources[3].(*Share)
	assert.True(t, ok)
}

var smbSampleBad1 = `
{
  "resources": [
    {
      "resource_type": "ceph.smb.cluster",
      "cluster_id": "cluster1",
      "auth_mode": "active-directory",
      "domain_settings": {
        "realm": "DOMAIN1.SINK.TEST",
        "join_sources": [
          {"source_type": "resource", "ref": "join1-admin"}
        ]
      },
      "custom_dns": ["192.168.76.210"],
      "placement": {
        "count": 1,
        "label": "ilovesmb"
      }
    },
    {
      "resource_type": "ceph.smb.fish",
      "fish_id": "flounder",
      "size": 100
    }
  ]
}
`

func TestResourcesUnmarshalError(t *testing.T) {
	g := resourceGroup{}
	err := json.Unmarshal([]byte(smbSampleBad1), &g)
	assert.Error(t, err)
}

func TestValidateResources(t *testing.T) {
	good := NewUsersAndGroupsToRemove("bread")
	bad1 := NewShare("zzz", "nofs")
	bad2 := &Cluster{}

	t.Run("allGood", func(t *testing.T) {
		assert.NoError(t, ValidateResources([]Resource{good}))
	})
	t.Run("oneBad", func(t *testing.T) {
		assert.ErrorContains(t, ValidateResources([]Resource{bad1}), "#0")
	})
	t.Run("twoBad", func(t *testing.T) {
		assert.ErrorContains(t, ValidateResources([]Resource{bad1, bad2}), "#0")
	})
	t.Run("goodBad", func(t *testing.T) {
		assert.ErrorContains(t, ValidateResources([]Resource{good, bad2}), "#1")
	})
}

func TestResourceRef(t *testing.T) {
	t.Run("resourceType", func(t *testing.T) {
		var ref ResourceRef = ClusterType
		assert.Equal(t, ref.Type(), ClusterType)
		assert.Equal(t, ref.String(), "ceph.smb.cluster")
	})
	t.Run("resourceIDCluster", func(t *testing.T) {
		var ref ResourceRef = ResourceID{ClusterType, "bob"}
		assert.Equal(t, ref.Type(), ClusterType)
		assert.Equal(t, ref.String(), "ceph.smb.cluster.bob")
	})
	t.Run("resourceIDShare", func(t *testing.T) {
		// Create a Share ResourceRef with only one ID - in string form this is
		// ceph.smb.share.bob (a cluster ID, but no share ID) in the show command
		// this would match all shares in cluster
		var ref ResourceRef = ResourceID{ShareType, "bob"}
		assert.Equal(t, ref.Type(), ShareType)
		assert.Equal(t, ref.String(), "ceph.smb.share.bob")
	})
	t.Run("childResourceID", func(t *testing.T) {
		// Create a Share ResourceRef with both the cluster and share id values.
		// This would identify a single share
		var ref ResourceRef = ChildResourceID{ShareType, "bob", "share1"}
		assert.Equal(t, ref.Type(), ShareType)
		assert.Equal(t, ref.String(), "ceph.smb.share.bob.share1")
	})
}

func TestRandName(t *testing.T) {
	t.Run("cluster", func(t *testing.T) {
		n := randName("cluster")
		assert.Len(t, n, 15)
		assert.Contains(t, n, "cluster")
	})
	t.Run("c", func(t *testing.T) {
		n := randName("c")
		assert.Len(t, n, 9)
		assert.Contains(t, n, "c")
	})
	t.Run("clustersthatmuster", func(t *testing.T) {
		n := randName("clustersthatmuster")
		assert.Len(t, n, 18)
		assert.Contains(t, n, "clustersth")
	})
}

func TestErrorMerge(t *testing.T) {
	var e error
	results := ResultGroup{
		Success: true,
		Results: []*Result{
			{
				success:  true,
				resource: NewClusterToRemove("blat"),
			},
		},
	}
	assert.NoError(t, errorPick(results, e))

	results.Results[0].success = false
	results.Results[0].message = "bad mood"
	results.Success = false
	assert.ErrorContains(t, errorPick(results, e), "bad mood")

	e = fmt.Errorf("bigbad")
	assert.ErrorContains(t, errorPick(results, e), "bigbad")
}

type phoneyConn struct {
	mgrCommand                func(buf [][]byte) ([]byte, string, error)
	mgrCommandWithInputBuffer func([][]byte, []byte) ([]byte, string, error)
	monCommand                func(buf []byte) ([]byte, string, error)
	monCommandWithInputBuffer func([]byte, []byte) ([]byte, string, error)
}

func (p *phoneyConn) MgrCommand(buf [][]byte) ([]byte, string, error) {
	return p.mgrCommand(buf)
}
func (p *phoneyConn) MgrCommandWithInputBuffer(cbuf [][]byte, dbuf []byte) ([]byte, string, error) {
	return p.mgrCommandWithInputBuffer(cbuf, dbuf)
}
func (p *phoneyConn) MonCommand(buf []byte) ([]byte, string, error) {
	return p.monCommand(buf)
}
func (p *phoneyConn) MonCommandWithInputBuffer(cbuf []byte, dbuf []byte) ([]byte, string, error) {
	return p.monCommandWithInputBuffer(cbuf, dbuf)
}

func TestPhonyApplyError(t *testing.T) {
	sa := NewFromConn(&phoneyConn{
		mgrCommandWithInputBuffer: func(_ [][]byte, _ []byte) ([]byte, string, error) {
			return []byte("ERROR!"), "FAIL", nil
		},
	})
	ja1 := NewJoinAuth("ja1").SetAuth("joiner", "xyz")
	_, err := sa.Apply([]Resource{ja1}, nil)
	assert.Error(t, err)
}

func TestPhonyShowError(t *testing.T) {
	sa := NewFromConn(&phoneyConn{
		mgrCommand: func(_ [][]byte) ([]byte, string, error) {
			return []byte("ERROR!"), "FAIL", nil
		},
	})
	_, err := sa.Show(nil, nil)
	assert.Error(t, err)
}

func TestPhonyShowJSONError(t *testing.T) {
	sa := NewFromConn(&phoneyConn{
		mgrCommand: func(_ [][]byte) ([]byte, string, error) {
			r := `
{"resources: [
  {
    "resource_type": "ceph.smb.rubber.duck",
    "color": "yellow",
    "squeak": true
  }
]}
`
			return []byte(r), "", nil
		},
	})
	_, err := sa.Show(nil, nil)
	assert.Error(t, err)
}

func TestPhonyShowJSONSingle(t *testing.T) {
	sa := NewFromConn(&phoneyConn{
		mgrCommand: func(_ [][]byte) ([]byte, string, error) {
			r := `
{
"resource_type": "ceph.smb.join.auth",
"auth_id": "adj1",
"auth": {"username": "bob", "password": "myBIGsecret"}
}
`
			return []byte(r), "", nil
		},
	})
	_, err := sa.Show(nil, nil)
	assert.NoError(t, err)
}
