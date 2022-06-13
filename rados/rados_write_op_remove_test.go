package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestWriteOpRemove() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	var (
		oid = "TestWriteOpRemove"
		err error
	)

	// Create an object.
	op1 := CreateWriteOp()
	defer op1.Release()
	op1.Create(CreateIdempotent)
	err = op1.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Try to stat() it before removal. It should succeed.
	_, err = suite.ioctx.Stat(oid)
	ta.NoError(err)

	// Try to remove it. It should succeed.
	op2 := CreateWriteOp()
	defer op2.Release()
	op2.Remove()
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Try to stat() it. It should fail with ENOENT.
	_, err = suite.ioctx.Stat(oid)
	ta.Error(err)
	ta.Equal(ErrNotFound, err)
}
