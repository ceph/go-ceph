//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"github.com/stretchr/testify/assert"
)

func (suite *SMBAdminSuite) TestClusterWithRemoteControl() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	cluster := NewUserCluster("caddrs1")
	ug := NewLinkedUsersAndGroups(cluster).SetValues(
		[]UserInfo{{"foobar", "L3tM31n"}},
		[]GroupInfo{{"clients"}},
	)
	tc1 := NewTLSCredential("tc1").Set(TLSCert, "INVALIDCERTDATA")
	tc2 := NewTLSCredential("tc2").Set(TLSKey, "INVALIDCERTDATA")
	tc3 := NewTLSCredential("tc3").Set(TLSCACert, "INVALIDCERTDATA")
	cluster.RemoteControl = &RemoteControl{
		Cert:   &TLSCredentialSource{ResourceSource, "tc1"},
		Key:    &TLSCredentialSource{ResourceSource, "tc2"},
		CACert: &TLSCredentialSource{ResourceSource, "tc3"},
	}

	rgroup, err := sa.Apply([]Resource{cluster, ug, tc1, tc2, tc3}, nil)
	assert.NoError(t, err)
	assert.True(t, rgroup.Ok())

	res, err := sa.Show([]ResourceRef{cluster.Identity()}, nil)
	assert.NoError(t, err)
	if assert.Len(t, res, 1) {
		nc := res[0].(*Cluster)
		if assert.NotNil(t, nc.RemoteControl) {
			assert.EqualValues(t, nc.RemoteControl.Cert.Ref, "tc1")
			assert.EqualValues(t, nc.RemoteControl.Key.Ref, "tc2")
			assert.EqualValues(t, nc.RemoteControl.CACert.Ref, "tc3")
		}
	}
}
