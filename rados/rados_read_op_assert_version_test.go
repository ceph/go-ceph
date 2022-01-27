//go:build ceph_preview
// +build ceph_preview

package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestReadOpAssertVersion() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	var (
		oid = "TestReadOpAssertVersion"
		err error
	)

	// Create an object.
	op1 := CreateWriteOp()
	defer op1.Release()
	op1.Create(CreateIdempotent)
	err = op1.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Retrieve last object version after writing.
	ver1, err := suite.ioctx.GetLastVersion()
	ta.NoError(err)

	// Read with version assert. It should succeed.
	op2 := CreateReadOp()
	defer op2.Release()
	op2.AssertVersion(ver1)
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Refresh the version.
	ver2, err := suite.ioctx.GetLastVersion()
	ta.NoError(err)

	// Read with version assert, but modify the version first.
	// It should fail.
	op3 := CreateReadOp()
	defer op3.Release()
	op3.AssertVersion(ver2 + 1)
	err = op3.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.Error(err)
}
