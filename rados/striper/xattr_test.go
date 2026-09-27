package striper

import (
	"github.com/stretchr/testify/require"
)

func (suite *StriperTestSuite) TestSetGetXattr() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestSetGetXattr"
	err = striper.Write(name, []byte("foo"), 0)
	require.NoError(suite.T(), err)

	xname := "foo.bar"
	err = striper.SetXattr(name, xname, []byte("razzmatazz"))
	require.NoError(suite.T(), err)

	buf := make([]byte, 32)
	size, err := striper.GetXattr(name, xname, buf)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 10, size)
	require.Equal(suite.T(), "razzmatazz", string(buf[:size]))

	size, err = striper.GetXattr(name, "nope.nope.nope", buf)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), -61, err.(radosStriperError).ErrorCode())
}

func (suite *StriperTestSuite) TestRmXattr() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestRmXattr"
	err = striper.Write(name, []byte("foo"), 0)
	require.NoError(suite.T(), err)

	xname := "foo.bar"
	err = striper.SetXattr(name, xname, []byte("razzmatazz"))
	require.NoError(suite.T(), err)

	buf := make([]byte, 32)
	size, err := striper.GetXattr(name, xname, buf)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 10, size)
	require.Equal(suite.T(), "razzmatazz", string(buf[:size]))

	err = striper.RmXattr(name, xname)
	require.NoError(suite.T(), err)

	size, err = striper.GetXattr(name, xname, buf)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), -61, err.(radosStriperError).ErrorCode())
}

func (suite *StriperTestSuite) TestListXattrs() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestListXattrs"
	err = striper.Write(name, []byte("foo"), 0)
	require.NoError(suite.T(), err)

	err = striper.SetXattr(name, "foo.bar", []byte("razzmatazz"))
	require.NoError(suite.T(), err)
	err = striper.SetXattr(name, "foo.baz", []byte("razzle"))
	require.NoError(suite.T(), err)
	err = striper.SetXattr(name, "foo.zap", []byte("dazzle"))
	require.NoError(suite.T(), err)

	xm, err := striper.ListXattrs(name)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), xm, 3)
	require.Equal(suite.T(), "razzmatazz", string(xm["foo.bar"]))
	require.Equal(suite.T(), "razzle", string(xm["foo.baz"]))
	require.Equal(suite.T(), "dazzle", string(xm["foo.zap"]))

	err = striper.RmXattr(name, "foo.bar")
	require.NoError(suite.T(), err)

	xm, err = striper.ListXattrs(name)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), xm, 2)
	require.Equal(suite.T(), "razzle", string(xm["foo.baz"]))
	require.Equal(suite.T(), "dazzle", string(xm["foo.zap"]))
}

func (suite *StriperTestSuite) TestXattrEmpty() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestXattrEmpty"
	err = striper.Write(name, []byte("foo"), 0)
	require.NoError(suite.T(), err)

	err = striper.SetXattr(name, "full", []byte("value"))
	require.NoError(suite.T(), err)

	// Each call used to panic with an index out of range on the empty slice.
	for _, data := range [][]byte{nil, {}} {
		err = striper.SetXattr(name, "empty", data)
		require.NoError(suite.T(), err)

		size, err := striper.GetXattr(name, "empty", data)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), 0, size)

		// a non-empty value does not fit in an empty buffer
		_, err = striper.GetXattr(name, "full", data)
		require.Error(suite.T(), err)
		require.Equal(suite.T(), -34, err.(radosStriperError).ErrorCode())
	}
}
