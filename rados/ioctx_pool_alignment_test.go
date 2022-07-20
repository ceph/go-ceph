//go:build ceph_preview
// +build ceph_preview

package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestAlignment() {
	suite.SetupConnection()

	_, err := suite.ioctx.Alignment()
	assert.NoError(suite.T(), err)
}
