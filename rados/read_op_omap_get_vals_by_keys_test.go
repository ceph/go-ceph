//go:build ceph_preview
// +build ceph_preview

package rados

import (
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestReadOpOmapGetValsByKeys() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	var (
		oid = "TestReadOpOmapGetValsByKeys"
		kvs = map[string][]byte{
			"k1": []byte("v1"),
			"k2": []byte("v2"),
		}
		err error
	)

	// Create an object and populate it with data.

	op1 := CreateWriteOp()
	defer op1.Release()
	op1.Create(CreateIdempotent)
	op1.SetOmap(kvs)
	err = op1.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// Retrieve objects omap key-value pairs by ks keys, but add
	// a few non-existingones which are expected to be silently skipped.

	var (
		actualKVs = make(map[string][]byte)
		ks        = []string{"k1", "k2", "xx", "yy"}
	)

	op2 := CreateReadOp()
	defer op2.Release()
	kvStep := op2.GetOmapValuesByKeys(ks)

	_, err = kvStep.Next()
	ta.Equal(ErrOperationIncomplete, err)

	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	for {
		kv, err := kvStep.Next()
		ta.NoError(err)

		if kv == nil {
			break
		}

		actualKVs[kv.Key] = kv.Value
	}

	ta.Equal(kvs, actualKVs)
}
