//go:build ceph_preview

package rados

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cls_hello's write_return_data sets the xattr foo and returns 42. With
// OperationReturnVec the OSD returns 42 for the committed write instead of
// zeroing it, and the write op must count it as success.
func (suite *RadosTestSuite) TestWriteOpPositiveReturn() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	writeReturnData := func(flags OperationFlags) {
		oid := suite.GenObjectName()
		wop := CreateWriteOp()
		defer wop.Release()
		wop.Exec("hello", "write_return_data", nil)
		require.NoError(suite.T(), wop.Operate(suite.ioctx, oid, flags))

		buf := make([]byte, 8)
		n, err := suite.ioctx.GetXattr(oid, "foo", buf)
		ta.NoError(err)
		ta.Equal([]byte("bar"), buf[:n])
	}
	writeReturnData(OperationNoFlag)
	writeReturnData(OperationReturnVec)
}
