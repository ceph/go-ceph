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

func (suite *RadosTestSuite) TestSnapshotIDFunctions() {
	suite.SetupConnection()

	suite.T().Run("invalidIOCtx", func(t *testing.T) {
		ioctx := &IOContext{}
		_, err := ioctx.LookupSnap("")
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidIOContext)

		var snapID SnapID
		snapID = 22 // some random number
		_, err = ioctx.GetSnapName(snapID)
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidIOContext)

		_, err = ioctx.GetSnapStamp(snapID)
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidIOContext)
	})

	// Invalid args
	suite.T().Run("InvalidArgs", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)

		err = ioctx.CreateSnap("")
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, ioctx.RemoveSnap(""))
		}()

		// Again, this works!!
		_, err = ioctx.LookupSnap("")
		assert.NoError(t, err)

		// Non-existing Snap
		_, err = ioctx.LookupSnap("someSnapName")
		assert.Error(t, err)

		var snapID SnapID
		snapID = 22 // some random number
		_, err = ioctx.GetSnapName(snapID)
		assert.Error(t, err)

		_, err = ioctx.GetSnapStamp(snapID)
		assert.Error(t, err)
	})

	// Valid SnapID operations.
	suite.T().Run("ValidSnapIDOps", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)

		snapName := "mySnap"
		err = ioctx.CreateSnap(snapName)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, ioctx.RemoveSnap(snapName))
		}()

		snapID, err := ioctx.LookupSnap(snapName)
		assert.NoError(t, err)

		retName, err := ioctx.GetSnapName(snapID)
		assert.NoError(t, err)
		assert.Equal(t, snapName, retName)

		_, err = ioctx.GetSnapStamp(snapID)
		assert.NoError(t, err)
	})
}

func (suite *RadosTestSuite) TestListSnapshot() {
	suite.SetupConnection()
	ioctx, err := suite.conn.OpenIOContext(suite.pool)
	require.NoError(suite.T(), err)

	snapName := []string{"snap1", "snap2", "snap3"}
	err = ioctx.CreateSnap(snapName[0])
	assert.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), ioctx.RemoveSnap(snapName[0]))
	}()

	err = ioctx.CreateSnap(snapName[1])
	assert.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), ioctx.RemoveSnap(snapName[1]))
	}()

	err = ioctx.CreateSnap(snapName[2])
	assert.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), ioctx.RemoveSnap(snapName[2]))
	}()

	suite.T().Run("invalidIOContext", func(t *testing.T) {
		ioctx := &IOContext{}
		_, err := ioctx.ListSnaps()
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidIOContext)
	})

	snapList, err := ioctx.ListSnaps()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), snapList)

	listLen := len(snapList)

	suite.T().Run("NumberOfSnapshots", func(t *testing.T) {
		assert.Equal(t, 3, listLen)
	})

	suite.T().Run("MatchSnapNamesWithID", func(t *testing.T) {
		for _, id := range snapList[0 : listLen-1] {
			retName, err := ioctx.GetSnapName(id)
			assert.NoError(t, err)
			assert.NotNil(t, retName)
			assert.Contains(t, snapName, retName)
		}
	})
}
