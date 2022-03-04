//go:build ceph_preview
// +build ceph_preview

package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestSetLocator() {
	suite.SetupConnection()

	// create normal object without locator - used later to test reset of locator
	testDataNoLocator := []byte("no locator")
	err := suite.ioctx.Write("default-locator", testDataNoLocator, 0)
	assert.NoError(suite.T(), err)

	// test create and read with different locator
	testDataLocator := []byte("test data")
	suite.ioctx.SetLocator("SomeOtherLocator")
	err = suite.ioctx.Write("different-locator", testDataLocator, 0)
	assert.NoError(suite.T(), err)

	_, err = suite.ioctx.Stat("different-locator")
	assert.NoError(suite.T(), err)

	bytesOut := make([]byte, len(testDataLocator))
	nOut, err := suite.ioctx.Read("different-locator", bytesOut, 0)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), nOut, len(testDataLocator))
	assert.Equal(suite.T(), testDataLocator, bytesOut)

	// test stat with wrong locator
	suite.ioctx.SetLocator("SomeWrongLocator")
	_, err = suite.ioctx.Stat("different-locator")
	assert.Error(suite.T(), err)
	_, err = suite.ioctx.Stat("default-locator")
	assert.Error(suite.T(), err)

	// test reset of locator and access to object without locator
	suite.ioctx.SetLocator("")
	_, err = suite.ioctx.Stat("default-locator")
	assert.NoError(suite.T(), err)

	bytesOut = make([]byte, len(testDataNoLocator))
	nOut, err = suite.ioctx.Read("default-locator", bytesOut, 0)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), nOut, len(testDataNoLocator))
	assert.Equal(suite.T(), testDataNoLocator, bytesOut)
}
