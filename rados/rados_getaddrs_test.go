package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestGetAddrs() {
	suite.SetupConnection()

	addrs, err := suite.conn.GetAddrs()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), addrs, "rados_getaddrs")
}
