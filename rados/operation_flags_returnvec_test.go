//go:build ceph_preview

package rados

import (
	"encoding/binary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodeObjVersion encodes a cls_version obj_version wrapped in the
// single-member struct that cls_version_set_op and cls_version_read_ret
// both are.
func encodeObjVersion(ver uint64, tag string) []byte {
	envelope := func(payload []byte) []byte {
		b := []byte{1, 1} // struct_v, struct_compat
		b = binary.LittleEndian.AppendUint32(b, uint32(len(payload)))
		return append(b, payload...)
	}
	objv := binary.LittleEndian.AppendUint64(nil, ver)
	objv = binary.LittleEndian.AppendUint32(objv, uint32(len(tag)))
	objv = append(objv, tag...)
	return envelope(envelope(objv))
}

func (suite *RadosTestSuite) TestOperationReturnVec() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := suite.GenObjectName()
	v42 := encodeObjVersion(42, "returnvec")
	v43 := encodeObjVersion(43, "returnvec")
	v44 := encodeObjVersion(44, "returnvec")

	wop := CreateWriteOp()
	defer wop.Release()
	wop.Exec("version", "set", v42)
	require.NoError(suite.T(), wop.Operate(suite.ioctx, oid, OperationNoFlag))

	// "set" writes, so the OSD runs the whole op as a write and discards
	// the output of "read" unless the op carries OperationReturnVec.
	// "read" sees the version from before the op.
	setAndRead := func(objv []byte, flags OperationFlags) []byte {
		rop := CreateReadOp()
		defer rop.Release()
		rop.Exec("version", "set", objv)
		readStep := rop.Exec("version", "read", nil)
		require.NoError(suite.T(), rop.Operate(suite.ioctx, oid, flags))
		out, err := readStep.Bytes()
		ta.NoError(err)
		return out
	}
	ta.Empty(setAndRead(v43, OperationNoFlag))
	ta.Equal(v43, setAndRead(v44, OperationReturnVec))

	rop := CreateReadOp()
	defer rop.Release()
	readStep := rop.Exec("version", "read", nil)
	require.NoError(suite.T(), rop.Operate(suite.ioctx, oid, OperationNoFlag))
	out, err := readStep.Bytes()
	ta.NoError(err)
	ta.Equal(v44, out)
}
