package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestWriteOpSetAllocationHint() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpSetAllocationHint"
	data := []byte("write this")

	// Create an object and populate it with data.
	op1 := CreateWriteOp()
	defer op1.Release()
	op1.Create(CreateIdempotent)
	op1.SetAllocationHint(4096, 20, AllocHintCompressible|AllocHintLonglived)
	op1.WriteFull([]byte(data))
	err := op1.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	op2 := CreateWriteOp()
	defer op2.Release()
	op2.SetAllocationHint(4096, 200, AllocHintNoHint)
	op2.WriteSame([]byte(data), 200, uint64(len(data)))
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)
}
