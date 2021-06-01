// +build nautilus

package rados

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *RadosTestSuite) TestSetUnsetPoolFullTry() {
	suite.SetupConnection()
	suite.T().Run("invalidIOContext", func(t *testing.T) {
		ioctx := &IOContext{}
		err := ioctx.SetPoolFullTry()
		assert.Error(t, err)
		err = ioctx.UnsetPoolFullTry()
		assert.Error(t, err)
	})

	suite.T().Run("validIOContext", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)
		err = ioctx.SetPoolFullTry()
		assert.NoError(t, err)
		err = ioctx.UnsetPoolFullTry()
		assert.NoError(t, err)
	})
}
