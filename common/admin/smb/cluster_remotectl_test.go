//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalUnmarshalClusterWithRemoteControl(t *testing.T) {
	c1 := NewUserCluster("rcc")
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
	c1.RemoteControl = &RemoteControl{
		Cert:   &TLSCredentialSource{ResourceSource, tc1.TLSCredentialID},
		Key:    &TLSCredentialSource{ResourceSource, tc2.TLSCredentialID},
		CACert: &TLSCredentialSource{ResourceSource, tc3.TLSCredentialID},
	}

	rg := resourceGroup{Resources: []Resource{c1, ug, tc1, tc2, tc3}}
	assert.NoError(t, ValidateResources(rg.Resources))
	j, err := json.Marshal(rg)
	assert.NoError(t, err)

	rg = resourceGroup{}
	err = json.Unmarshal(j, &rg)
	assert.NoError(t, err)
	assert.Len(t, rg.Resources, 5)
	assert.Equal(t, rg.Resources[0].Type(), ClusterType)
	assert.Equal(t, rg.Resources[1].Type(), UsersAndGroupsType)
	assert.Equal(t, rg.Resources[2].Type(), TLSCredentialType)
	assert.Equal(t, rg.Resources[3].Type(), TLSCredentialType)
	assert.Equal(t, rg.Resources[4].Type(), TLSCredentialType)

	c2 := rg.Resources[0].(*Cluster)
	if assert.NotNil(t, c2.RemoteControl) {
		assert.EqualValues(t, c2.RemoteControl.Cert.Ref, tc1.TLSCredentialID)
		assert.EqualValues(t, c2.RemoteControl.Key.Ref, tc2.TLSCredentialID)
		assert.EqualValues(t, c2.RemoteControl.CACert.Ref, tc3.TLSCredentialID)
	}
}
