package rados

import (
	"syscall"

	"github.com/stretchr/testify/assert"
)

// Each case used to panic with an index out of range on the empty slice.

func (suite *RadosTestSuite) TestWriteOpSetXattrEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	for _, value := range [][]byte{nil, {}} {
		oid := suite.GenObjectName()
		op := CreateWriteOp()
		op.Create(CreateIdempotent)
		op.SetXattr("empty", value)
		ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))
		op.Release()

		buf := make([]byte, 16)
		n, err := suite.ioctx.GetXattr(oid, "empty", buf)
		ta.NoError(err)
		ta.Equal(0, n)
	}
}

func (suite *RadosTestSuite) TestWriteOpWriteFullEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	for _, data := range [][]byte{nil, {}} {
		oid := suite.GenObjectName()
		op := CreateWriteOp()
		op.WriteFull(data)
		ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))
		op.Release()

		stat, err := suite.ioctx.Stat(oid)
		ta.NoError(err)
		ta.Equal(uint64(0), stat.Size)
	}

	// an empty WriteFull truncates an existing object
	oid := suite.GenObjectName()
	ta.NoError(suite.ioctx.Write(oid, []byte("data"), 0))
	op := CreateWriteOp()
	defer op.Release()
	op.WriteFull(nil)
	ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))
	stat, err := suite.ioctx.Stat(oid)
	ta.NoError(err)
	ta.Equal(uint64(0), stat.Size)
}

func (suite *RadosTestSuite) TestWriteOpWriteEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := suite.GenObjectName()
	ta.NoError(suite.ioctx.Write(oid, []byte("data"), 0))
	op := CreateWriteOp()
	defer op.Release()
	op.Write([]byte{}, 0)
	ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))

	buf := make([]byte, 16)
	n, err := suite.ioctx.Read(oid, buf, 0)
	ta.NoError(err)
	ta.Equal("data", string(buf[:n]))
}

func (suite *RadosTestSuite) TestWriteOpWriteSameEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := suite.GenObjectName()
	ta.NoError(suite.ioctx.Write(oid, []byte("data"), 0))

	// the OSD accepts a zero-length writesame as a no-op
	op := CreateWriteOp()
	op.WriteSame([]byte{}, 0, 0)
	ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))
	op.Release()

	// but rejects an empty pattern with a nonzero length
	op = CreateWriteOp()
	op.WriteSame(nil, 8, 0)
	ta.Error(op.Operate(suite.ioctx, oid, OperationNoFlag))
	op.Release()

	buf := make([]byte, 16)
	n, err := suite.ioctx.Read(oid, buf, 0)
	ta.NoError(err)
	ta.Equal("data", string(buf[:n]))
}

func (suite *RadosTestSuite) TestWriteOpCmpExtEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := suite.GenObjectName()
	ta.NoError(suite.ioctx.Write(oid, []byte("data"), 0))

	// comparing an empty extent always matches
	for _, data := range [][]byte{nil, {}} {
		op := CreateWriteOp()
		step := op.CmpExt(data, 0)
		ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))
		ta.Equal(0, step.Result)
		op.Release()
	}
}

func (suite *RadosTestSuite) TestReadOpReadEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := suite.GenObjectName()
	ta.NoError(suite.ioctx.Create(oid, CreateExclusive))

	for _, data := range [][]byte{nil, {}} {
		op := CreateReadOp()
		step := op.Read(0, data)
		ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))
		ta.Equal(0, step.Result)
		ta.Equal(int64(0), step.BytesRead)
		op.Release()
	}

	// A zero-length read asks the OSD for the whole object, which does
	// not fit in an empty buffer.
	ta.NoError(suite.ioctx.Write(oid, []byte("data"), 0))
	op := CreateReadOp()
	defer op.Release()
	step := op.Read(0, nil)
	ta.NoError(op.Operate(suite.ioctx, oid, OperationNoFlag))
	ta.Equal(-int(syscall.ERANGE), step.Result)
	ta.Equal(int64(0), step.BytesRead)
}
