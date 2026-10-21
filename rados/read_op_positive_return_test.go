//go:build ceph_preview

package rados

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cls_hello's write_return_data returns 42, which reaches the client only
// with OperationReturnVec. The OSD also leaves 42 in the result of a later
// action that does not set its own, and returns the result of the last
// action as the result of the operation. Like C++ librados, the op and its
// steps must count all of these as success.
func (suite *RadosTestSuite) TestReadOpPositiveReturn() {
	suite.SetupConnection()

	suite.T().Run("getOmapValuesByKeys", func(t *testing.T) {
		ta := assert.New(t)
		oid := suite.GenObjectName()
		require.NoError(t, suite.ioctx.SetOmap(
			oid, map[string][]byte{"key": []byte("value")}))

		rop := CreateReadOp()
		defer rop.Release()
		es := rop.Exec("hello", "write_return_data", nil)
		kvStep := rop.GetOmapValuesByKeys([]string{"key"})
		ta.NoError(rop.Operate(suite.ioctx, oid, OperationReturnVec))

		out, err := es.Bytes()
		ta.NoError(err)
		ta.Equal([]byte("you might see this"), out)
		kv, err := kvStep.Next()
		ta.NoError(err)
		ta.Equal(&OmapKeyValue{Key: "key", Value: []byte("value")}, kv)
	})

	// without omap the OSD skips the omap iteration and leaves the
	// previous result in place
	suite.T().Run("getOmapValues", func(t *testing.T) {
		ta := assert.New(t)
		oid := suite.GenObjectName()

		rop := CreateReadOp()
		defer rop.Release()
		es := rop.Exec("hello", "write_return_data", nil)
		gos := rop.GetOmapValues("", "", 10)
		ta.NoError(rop.Operate(suite.ioctx, oid, OperationReturnVec))

		out, err := es.Bytes()
		ta.NoError(err)
		ta.Equal([]byte("you might see this"), out)
		kv, err := gos.Next()
		ta.NoError(err)
		ta.Nil(kv)
	})
}
