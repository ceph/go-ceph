package rados

import (
	"github.com/stretchr/testify/assert"
)

// Each case used to panic with an index out of range on the empty slice.

func (suite *RadosTestSuite) TestWriteFullEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	for _, data := range [][]byte{nil, {}} {
		oid := suite.GenObjectName()
		ta.NoError(suite.ioctx.Write(oid, []byte("data"), 0))

		// an empty WriteFull truncates the object
		ta.NoError(suite.ioctx.WriteFull(oid, data))
		stat, err := suite.ioctx.Stat(oid)
		ta.NoError(err)
		ta.Equal(uint64(0), stat.Size)
	}
}

func (suite *RadosTestSuite) TestAppendEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := suite.GenObjectName()
	ta.NoError(suite.ioctx.Write(oid, []byte("data"), 0))

	for _, data := range [][]byte{nil, {}} {
		ta.NoError(suite.ioctx.Append(oid, data))
	}

	buf := make([]byte, 16)
	n, err := suite.ioctx.Read(oid, buf, 0)
	ta.NoError(err)
	ta.Equal("data", string(buf[:n]))
}

func (suite *RadosTestSuite) TestXattrEmpty() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := suite.GenObjectName()
	ta.NoError(suite.ioctx.Create(oid, CreateExclusive))
	ta.NoError(suite.ioctx.SetXattr(oid, "nil", nil))
	ta.NoError(suite.ioctx.SetXattr(oid, "empty", []byte{}))
	ta.NoError(suite.ioctx.SetXattr(oid, "full", []byte("value")))

	for _, data := range [][]byte{nil, {}} {
		n, err := suite.ioctx.GetXattr(oid, "nil", data)
		ta.NoError(err)
		ta.Equal(0, n)
		n, err = suite.ioctx.GetXattr(oid, "empty", data)
		ta.NoError(err)
		ta.Equal(0, n)

		// a non-empty value does not fit in an empty buffer
		_, err = suite.ioctx.GetXattr(oid, "full", data)
		ta.ErrorIs(err, errRange)
	}

	xattrs, err := suite.ioctx.ListXattrs(oid)
	ta.NoError(err)
	ta.Empty(xattrs["nil"])
	ta.Empty(xattrs["empty"])
	ta.Equal([]byte("value"), xattrs["full"])
}
