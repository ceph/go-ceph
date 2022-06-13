package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestWriteOpSetXattr() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	var (
		oid        = "TestWriteOpSetXattr"
		xattrName  = "attrname"
		xattrValue = []byte("attrvalue")
	)

	// Create an object and populate it with data.
	op1 := CreateWriteOp()
	defer op1.Release()
	op1.Create(CreateIdempotent)
	op1.SetXattr(xattrName, xattrValue)
	err := op1.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Read object's xattrs and compare.
	actualXattrs, err := suite.ioctx.ListXattrs(oid)
	ta.NoError(err)
	ta.Equal(map[string][]byte{xattrName: xattrValue}, actualXattrs)
}
