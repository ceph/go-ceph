//go:build !(pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericResourceUnmarshal(t *testing.T) {
	g := struct{ Resources []GenericResource }{}
	b := []byte(smbSample1)
	assert.NoError(t, json.Unmarshal(b, &g))
	assert.Len(t, g.Resources, 4)

	t.Run("r0", func(t *testing.T) {
		r0 := g.Resources[0]
		assert.NoError(t, r0.Validate())
		assert.Equal(t, r0.Type(), ClusterType)
		assert.Equal(t, r0.Intent(), Present)
		ref := r0.Identity()
		assert.Equal(t, ref.Type(), ClusterType)
		assert.Equal(t, ref.String(), "ceph.smb.cluster.cluster1")
	})

	t.Run("r1", func(t *testing.T) {
		r1 := g.Resources[1]
		assert.NoError(t, r1.Validate())
		assert.Equal(t, r1.Type(), JoinAuthType)
		assert.Equal(t, r1.Intent(), Present)
		ref := r1.Identity()
		assert.Equal(t, ref.Type(), JoinAuthType)
		assert.Equal(t, ref.String(), "ceph.smb.join.auth.join1-admin")
	})

	t.Run("r2", func(t *testing.T) {
		r2 := g.Resources[2]
		assert.NoError(t, r2.Validate())
		assert.Equal(t, r2.Type(), ShareType)
		assert.Equal(t, r2.Intent(), Present)
		ref := r2.Identity()
		assert.Equal(t, ref.Type(), ShareType)
		assert.Equal(t, ref.String(), "ceph.smb.share.cluster1.share1")
	})

	t.Run("r3", func(t *testing.T) {
		r3 := g.Resources[3]
		assert.NoError(t, r3.Validate())
		assert.Equal(t, r3.Type(), ShareType)
		assert.Equal(t, r3.Intent(), Present)
		ref := r3.Identity()
		assert.Equal(t, ref.Type(), ShareType)
		assert.Equal(t, ref.String(), "ceph.smb.share.cluster1.share2")

		assert.Equal(t, r3.Values["name"].(string), "Sub Volume")
		assert.Equal(t, r3.Values["cephfs"].(map[string]any)["subvolume"].(string), "sv1")
	})
}

func TestGenericConversionToShare(t *testing.T) {
	share := NewShare("c1", "ss1").SetCephFS("cephfs", "g1", "v1", "/")
	g, err := ToGeneric(share)
	assert.NoError(t, err)
	assert.NotNil(t, g)

	assert.Equal(t, g.Type(), ShareType)
	assert.Equal(t, g.Identity().String(), "ceph.smb.share.c1.ss1")

	k := map[string]string{
		"scope": "mem",
		"name":  "bob",
	}
	g.Values["cephfs"].(map[string]any)["fscrypt_key"] = k
	j, err := json.Marshal(g)
	assert.NoError(t, err)
	assert.Contains(t, string(j), "mem")
	assert.Contains(t, string(j), "bob")
}

var jg1 = `
{
  "browseable": false,
  "cephfs": {
    "fscrypt_key": {
      "name": "bob",
      "scope": "mem"
    },
    "path": "g",
    "subvolume": "v1",
    "subvolumegroup": "g1",
    "volume": "cephfs"
  },
  "cluster_id": "c1",
  "intent": "present",
  "name": "",
  "readonly": false,
  "resource_type": "ceph.smb.share",
  "restrict_access": false,
  "share_id": "ss1"
}
`

func TestGenericConvert(t *testing.T) {
	g := &GenericResource{}
	assert.NoError(t, json.Unmarshal([]byte(jg1), g))
	assert.NotNil(t, g)
	assert.Equal(t, g.Type(), ShareType)
	assert.Equal(t, g.Values["cluster_id"].(string), "c1")
	assert.Equal(t, g.Values["share_id"].(string), "ss1")

	sr, err := g.Convert()
	assert.NoError(t, err)
	assert.Equal(t, sr.Type(), ShareType)
	assert.Equal(t, sr.(*Share).ClusterID, "c1")
	assert.Equal(t, sr.(*Share).ShareID, "ss1")
}

func TestGenericConvertErrors(t *testing.T) {
	g := &GenericResource{
		Values: map[string]any{},
	}
	_, err := g.Convert()
	assert.Error(t, err)

	g.Values["resource_type"] = "ceph.smb.floop"
	_, err = g.Convert()
	assert.Error(t, err)

	g.Values["resource_type"] = "ceph.smb.share"
	g.Values["cluster_id"] = "foo"
	g.Values["share_id"] = "bar"
	_, err = g.Convert()
	assert.NoError(t, err)
}

func TestGenericValidate(t *testing.T) {
	g := &GenericResource{
		Values: map[string]any{},
	}
	err := g.Validate()
	assert.ErrorContains(t, err, "resource_type")

	g.Values["resource_type"] = "ceph.smb.floop"
	err = g.Validate()
	assert.ErrorContains(t, err, "identity")

	g.Values["intent"] = "money"
	err = g.Validate()
	assert.ErrorContains(t, err, "intent")

	g.Values["intent"] = "present"
	g.Values["floop_id"] = "gambol"
	err = g.Validate()
	assert.NoError(t, err)
}

func TestGenericResourcesShow(t *testing.T) {
	sa := NewFromConn(&phoneyConn{
		mgrCommand: func(_ [][]byte) ([]byte, string, error) {
			return []byte(smbSample1), "", nil
		},
	})
	opts := &ShowOptions{}
	res, err := sa.Show(nil, opts.SetGeneric(true))
	assert.NoError(t, err)
	assert.Len(t, res, 4)
	assert.Equal(t, res[0].Type(), ClusterType)
	assert.Equal(t, res[1].Type(), JoinAuthType)
	assert.Equal(t, res[2].Type(), ShareType)
	assert.Equal(t, res[3].Type(), ShareType)
	g3 := res[3].(*GenericResource)
	assert.Equal(t,
		g3.Values["cephfs"].(map[string]any)["subvolume"].(string),
		"sv1")
}

// invent a bogus subsection that could exist in the future
// but is unknown by the library
var jg2 = `
{
  "resources": [
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
      },
      "bogusmogus": {
        "enabled": true,
        "weight": 7,
        "uuid": "0fca87c9-2572-4762-b868-9f06ff0791ee"
      }
    }
  ]
}
`

func TestGenericResourcesShowWithExtraData(t *testing.T) {
	sa := NewFromConn(&phoneyConn{
		mgrCommand: func(_ [][]byte) ([]byte, string, error) {
			return []byte(jg2), "", nil
		},
	})
	opts := &ShowOptions{}
	res, err := sa.Show(nil, opts.SetGeneric(true))
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, res[0].Type(), ShareType)
	g := res[0].(*GenericResource)
	assert.Equal(t,
		g.Values["cephfs"].(map[string]any)["subvolume"].(string),
		"sv1")
	assert.EqualValues(t,
		g.Values["bogusmogus"].(map[string]any)["enabled"].(bool),
		true)
	assert.EqualValues(t,
		g.Values["bogusmogus"].(map[string]any)["weight"].(float64),
		7)
	assert.EqualValues(t,
		g.Values["bogusmogus"].(map[string]any)["uuid"].(string),
		"0fca87c9-2572-4762-b868-9f06ff0791ee")

	// convert to known type but loses extra data
	r, err := g.Convert()
	assert.NoError(t, err)
	assert.Equal(t, r.Type(), ShareType)
	share := r.(*Share)
	assert.Equal(t, share.ShareID, "share2")
	assert.Equal(t, share.CephFS.SubVolume, "sv1")
	assert.Equal(t, share.CephFS.Path, "/")
}
