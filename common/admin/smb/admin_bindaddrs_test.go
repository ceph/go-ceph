//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"github.com/stretchr/testify/assert"
)

func (suite *SMBAdminSuite) TestClusterWithBindAddrs() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	cluster := NewUserCluster("caddrs1")
	ug := NewLinkedUsersAndGroups(cluster).SetValues(
		[]UserInfo{{"foobar", "L3tM31n"}},
		[]GroupInfo{{"clients"}},
	)
	cluster.BindAddrs = []BindAddress{
		NewBindAddress("192.168.2.12"),
		NewBindAddress("192.168.2.13"),
		NewBindAddress("192.168.2.14"),
		NewNetworkBindAddress("192.168.2.200/30"),
	}
	rgroup, err := sa.Apply([]Resource{cluster, ug}, nil)
	assert.NoError(t, err)
	assert.True(t, rgroup.Ok())

	res, err := sa.Show([]ResourceRef{cluster.Identity()}, nil)
	assert.NoError(t, err)
	if assert.Len(t, res, 1) {
		nc := res[0].(*Cluster)
		assert.Len(t, nc.BindAddrs, 4)
		assert.False(t, nc.BindAddrs[0].IsNetwork())
		assert.EqualValues(t, nc.BindAddrs[0].Address(), "192.168.2.12")
		assert.True(t, nc.BindAddrs[3].IsNetwork())
		assert.EqualValues(t, nc.BindAddrs[3].Network(), "192.168.2.200/30")
	}
}
