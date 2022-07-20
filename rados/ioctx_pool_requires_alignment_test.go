//go:build ceph_preview
// +build ceph_preview

package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestRequiresAlignment() {
	suite.SetupConnection()

	_, err := suite.ioctx.RequiresAlignment()
	assert.NoError(suite.T(), err)
}
