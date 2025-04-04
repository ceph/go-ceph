//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShareIdentity(t *testing.T) {
	share := NewShare("c1", "ss1")
	assert.Equal(t, share.Type(), ShareType)
	shid := share.Identity()
	assert.Equal(t, shid.Type(), ShareType)
	assert.Equal(t, shid.String(), "ceph.smb.share.c1.ss1")
	assert.Equal(t, share.Intent(), Present)
}

func TestShareValidate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := &Share{}
		assert.ErrorContains(t, s.Validate(), "Intent")
	})
	t.Run("missingID", func(t *testing.T) {
		s := &Share{IntentValue: Present}
		assert.ErrorContains(t, s.Validate(), "ClusterID")
	})
	t.Run("missingID2", func(t *testing.T) {
		s := &Share{IntentValue: Present, ClusterID: "c1"}
		assert.ErrorContains(t, s.Validate(), "ShareID")
	})
	t.Run("intentRemoved", func(t *testing.T) {
		s := &Share{IntentValue: Removed, ClusterID: "c1", ShareID: "s1"}
		assert.NoError(t, s.Validate())
	})

	// from this point on we're going to progressively add to
	// s1 until it is valid
	s1 := &Share{
		IntentValue: Present,
		ClusterID:   "c1",
		ShareID:     "ss1",
	}
	t.Run("missingCephFS", func(t *testing.T) {
		assert.ErrorContains(t, s1.Validate(), "CephFS")
	})
	s1.CephFS = &CephFSSource{}
	t.Run("missingCephFSVolume", func(t *testing.T) {
		assert.ErrorContains(t, s1.Validate(), "Volume")
	})
	s1.CephFS.Volume = "cephfs"
	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, s1.Validate())
	})

	s1.RestrictAccess = true
	s1.LoginControl = []ShareAccess{{}}
	t.Run("loginControlName", func(t *testing.T) {
		assert.ErrorContains(t, s1.Validate(), "Name")
	})
	s1.LoginControl = []ShareAccess{{Name: "abc", Category: AccessCategory("x")}}
	t.Run("loginControlCategory", func(t *testing.T) {
		assert.ErrorContains(t, s1.Validate(), "Category")
	})
	s1.LoginControl = []ShareAccess{{Name: "abc", Access: AccessMode("x")}}
	t.Run("loginControlAccess", func(t *testing.T) {
		assert.ErrorContains(t, s1.Validate(), "Access")
	})
	s1.LoginControl = []ShareAccess{{Name: "abc"}}
	t.Run("loginControlOk", func(t *testing.T) {
		assert.NoError(t, s1.Validate())
	})
}

func TestShareMarshalUnmarshal(t *testing.T) {
	share := NewShare("c1", "ss1").SetCephFS("cephfs", "g1", "v1", "/")
	j, err := json.Marshal(share)
	assert.NoError(t, err)

	share2 := &Share{}
	err = json.Unmarshal(j, share2)
	assert.NoError(t, err)
	assert.Equal(t, share.ClusterID, share2.ClusterID)
	assert.Equal(t, share.ShareID, share2.ShareID)
	assert.EqualValues(t, share.IntentValue, share2.IntentValue)
	assert.EqualValues(t, share.CephFS, share2.CephFS)
}
