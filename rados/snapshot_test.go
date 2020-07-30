package rados

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *RadosTestSuite) TestCreateRemoveSnapshot() {
	suite.SetupConnection()

	suite.T().Run("invalidIOCtx", func(t *testing.T) {
		ioctx := &IOContext{}
		err := ioctx.CreateSnap("someSnap")
		assert.Error(t, err)
		err = ioctx.RemoveSnap("someSnap")
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidIOContext)
	})

	suite.T().Run("NewSnap", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)

		snapName := "mySnap"
		err = ioctx.CreateSnap(snapName)
		assert.NoError(t, err)
		err = ioctx.RemoveSnap(snapName)
		assert.NoError(t, err)
	})

	suite.T().Run("ExistingSnap", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)

		snapName := "mySnap"
		err = ioctx.CreateSnap(snapName)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, ioctx.RemoveSnap(snapName))
		}()
		err = ioctx.CreateSnap(snapName)
		assert.Error(t, err)
	})

	suite.T().Run("NonExistingSnap", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)

		err = ioctx.RemoveSnap("someSnapName")
		assert.Error(t, err)
	})

	// Strangely, this works!!
	suite.T().Run("EmptySnapNameString", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)

		err = ioctx.CreateSnap("")
		assert.NoError(t, err)

		err = ioctx.RemoveSnap("")
		assert.NoError(t, err)
	})
}
