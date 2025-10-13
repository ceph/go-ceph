//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTLSCredentialIdentity(t *testing.T) {
	tc := NewTLSCredential("cert1")
	assert.Equal(t, tc.Type(), TLSCredentialType)
	tcid := tc.Identity()
	assert.Equal(t, tcid.Type(), TLSCredentialType)
	assert.Equal(t, tcid.String(), "ceph.smb.tls.credential.cert1")
	assert.Equal(t, tc.Intent(), Present)
}

func TestTLSCredentialValidate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		tc := &TLSCredential{}
		assert.ErrorContains(t, tc.Validate(), "intent")
	})
	t.Run("missingID", func(t *testing.T) {
		tc := &TLSCredential{IntentValue: Present}
		assert.ErrorContains(t, tc.Validate(), "TLSCredentialID")
	})
	t.Run("intentRemoved", func(t *testing.T) {
		tc := &TLSCredential{IntentValue: Removed, TLSCredentialID: "tc1"}
		assert.NoError(t, tc.Validate())
	})

	// from this point on we're going to progressively add to
	// tc1 until it is valid
	tc1 := &TLSCredential{
		IntentValue:     Present,
		TLSCredentialID: "tc1",
	}
	t.Run("missingCredType", func(t *testing.T) {
		assert.ErrorContains(t, tc1.Validate(), "missing CredentialType")
	})
	tc1.CredentialType = TLSContent("yyyy")
	t.Run("invalidCredType", func(t *testing.T) {
		assert.ErrorContains(t, tc1.Validate(), "invalid CredentialType")
	})
	tc1.CredentialType = TLSCert
	t.Run("missingValue", func(t *testing.T) {
		assert.ErrorContains(t, tc1.Validate(), "missing Value")
	})
	// not a valid cert, but we only check for some string in this lib.
	// let the smb mgr module validate the string contents as needed.
	tc1.Value = "asdfasdfasdfasdfasfasf"
	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, tc1.Validate())
	})
}

func TestMarshalTLSCredential(t *testing.T) {
	tc := NewTLSCredential("tc1")
	_, err := json.Marshal(tc)
	assert.NoError(t, err)
}

func TestMarshalUnmarshalTLSCredential(t *testing.T) {
	tc1 := NewTLSCredential("tc1").Set(TLSCert, "INVALIDCERTDATA")
	assert.NoError(t, tc1.Validate())
	j, err := json.Marshal(tc1)
	assert.NoError(t, err)
	tc2 := &TLSCredential{}
	err = json.Unmarshal(j, tc2)
	assert.NoError(t, err)
	assert.NoError(t, tc2.Validate())
	assert.EqualValues(t, tc1.TLSCredentialID, tc2.TLSCredentialID)
	assert.EqualValues(t, tc1.IntentValue, tc2.IntentValue)
	assert.EqualValues(t, tc1.CredentialType, tc2.CredentialType)
	assert.EqualValues(t, tc1.Value, tc2.Value)
}

func TestMarshalUnmarshalLinkedTLSCredential(t *testing.T) {
	c1 := NewUserCluster("c1")
	ug := NewLinkedUsersAndGroups(c1).SetValues(
		[]UserInfo{
			{"bob", "M4r13y"},
			{"billy", "p14n0m4N"},
		},
		[]GroupInfo{{"clients"}},
	)
	tc1 := NewLinkedTLSCredential(c1).Set(TLSCert, "INVALIDCERTDATA")
	tc2 := NewLinkedTLSCredential(c1).Set(TLSKey, "INVALIDCERTDATA")
	tc3 := NewLinkedTLSCredential(c1).Set(TLSCACert, "INVALIDCERTDATA")

	rg1 := resourceGroup{Resources: []Resource{c1, ug, tc1, tc2, tc3}}
	assert.NoError(t, ValidateResources(rg1.Resources))
	j, err := json.Marshal(rg1)
	assert.NoError(t, err)

	rg2 := resourceGroup{}
	err = json.Unmarshal(j, &rg2)
	assert.NoError(t, err)
	assert.Len(t, rg2.Resources, 5)
	assert.Equal(t, rg2.Resources[0].Type(), ClusterType)
	assert.Equal(t, rg2.Resources[1].Type(), UsersAndGroupsType)
	assert.Equal(t, rg2.Resources[2].Type(), TLSCredentialType)
	assert.Equal(t, rg2.Resources[3].Type(), TLSCredentialType)
	assert.Equal(t, rg2.Resources[4].Type(), TLSCredentialType)
}
