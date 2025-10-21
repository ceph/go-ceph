//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"github.com/stretchr/testify/assert"
)

func (suite *SMBAdminSuite) TestClusterWithCustomPorts() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	cluster := NewUserCluster("cports1")
	ug := NewLinkedUsersAndGroups(cluster).SetValues(
		[]UserInfo{{"foobar", "L3tM31n"}},
		[]GroupInfo{{"clients"}},
	)
	cluster.CustomPorts = CustomPortsMap{
		SMBService:        8675,
		SMBMetricsService: 3090,
		CTDBService:       2112,
	}
	rgroup, err := sa.Apply([]Resource{cluster, ug}, nil)
	assert.NoError(t, err)
	assert.True(t, rgroup.Ok())

	res, err := sa.Show([]ResourceRef{cluster.Identity()}, nil)
	assert.NoError(t, err)
	if assert.Len(t, res, 1) {
		nc := res[0].(*Cluster)
		assert.Len(t, nc.CustomPorts, 3)
		assert.EqualValues(t, nc.CustomPorts[SMBService], 8675)
		assert.EqualValues(t, nc.CustomPorts[SMBMetricsService], 3090)
		assert.EqualValues(t, nc.CustomPorts[CTDBService], 2112)
	}
}
