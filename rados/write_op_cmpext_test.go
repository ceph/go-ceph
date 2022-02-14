package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestWriteOpCmpExt() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpCmpExt"
	data := []byte("compare this")

	// Create an object and populate it with data.
	op1 := CreateWriteOp()
	defer op1.Release()
	op1.Create(CreateIdempotent)
	op1.WriteFull([]byte(data))
	err := op1.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Compare contents of the object. Should succeed.
	op2 := CreateWriteOp()
	defer op2.Release()
	cmpExtRes1 := op2.CmpExt(data, 0)
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)
	ta.Equal(cmpExtRes1.Result, int(0))

	// Compare contents of the object. Should fail.
	op3 := CreateWriteOp()
	defer op3.Release()
	cmpExtRes2 := op3.CmpExt([]byte("xxx"), 0)
	err = op3.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.Error(err)
	ta.NotEqual(cmpExtRes2.Result, int(0))
}
