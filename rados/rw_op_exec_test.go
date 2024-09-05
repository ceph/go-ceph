package rados

import (
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *RadosTestSuite) TestReadWriteOpExec() {
	suite.SetupConnection()

	oid, err := uuid.NewV4()
	assert.NoError(suite.T(), err)

	data := "cls data"

	wrOp := CreateWriteOp()
	defer wrOp.Release()
	wrOp.Exec("hello", "record_hello", []byte(data))
	assert.NoError(suite.T(), err)
	err = wrOp.Operate(suite.ioctx, oid.String(), OperationNoFlag)
	assert.NoError(suite.T(), err)

	rdOp := CreateReadOp()
	defer rdOp.Release()

	es := rdOp.Exec("hello", "replay", nil)
	result, err := es.Bytes()
	require.ErrorIs(suite.T(), err, ErrOperationIncomplete)
	require.Nil(suite.T(), result)

	err = rdOp.Operate(suite.ioctx, oid.String(), OperationNoFlag)
	assert.NoError(suite.T(), err)

	result, err = es.Bytes()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []byte("Hello, "+data+"!"), result)
}
