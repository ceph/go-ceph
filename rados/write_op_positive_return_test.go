package rados

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The hello class's write_return_data method sets an xattr and returns 42.
// RETURNVEC preserves that positive overall result.
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

	// A lua script that changes nothing leaves the transaction empty, so
	// its 42 reaches the client without RETURNVEC.
	oid := suite.GenObjectName()
	wop := CreateWriteOp()
	defer wop.Release()
	wop.Exec("lua", "eval_json", []byte(luaReturn42))
	ta.NoError(wop.Operate(suite.ioctx, oid, OperationNoFlag))
	_, err := suite.ioctx.Stat(oid)
	ta.ErrorIs(err, ErrNotFound)
}
