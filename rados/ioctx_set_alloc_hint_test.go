package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestSetAllocationHint() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestSetAllocationHint"
	data := []byte("write this")

	ioctx, err := suite.conn.OpenIOContext(suite.pool)
	ta.NoError(err)
	err = ioctx.SetAllocationHint(oid, 4096, 20, AllocHintCompressible|AllocHintLonglived|AllocHintRandomRead|AllocHintSequentialWrite)
	ta.NoError(err)
	err = ioctx.WriteFull(oid, []byte(data))
	ta.NoError(err)
	err = ioctx.SetAllocationHint(oid, 128, 128, AllocHintShortlived|AllocHintAppendOnly)
	ta.NoError(err)
	err = ioctx.Append(oid, []byte(data))
	ta.NoError(err)
	err = ioctx.SetAllocationHint(oid, 20, 0, AllocHintNoHint)
	ta.NoError(err)
}
