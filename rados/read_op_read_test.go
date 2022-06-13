package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestReadOpRead() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	var (
		oid  = "TestReadOpRead"
		data = []byte("data to read")
		err  error
	)

	// Create an object and populate it with data.
	op1 := CreateWriteOp()
	defer op1.Release()
	op1.Create(CreateIdempotent)
	op1.WriteFull(data)
	err = op1.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Read the object's contents and compare them with expected data.
	readBuf := make([]byte, 64)
	op2 := CreateReadOp()
	defer op2.Release()
	readOpRes := op2.Read(0, readBuf)
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)
	ta.Equal(int(0), readOpRes.Result)
	ta.Equal(int64(len(data)), readOpRes.BytesRead)
	ta.Equal(data, readBuf[:readOpRes.BytesRead])
}
