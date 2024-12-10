package striper

import (
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *StriperTestSuite) TestReadWrite() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestReadWrite"
	err = striper.Write(name, []byte("hello world"), 0)
	require.NoError(suite.T(), err)

	err = striper.Write(name, []byte("earthlings"), 6)
	require.NoError(suite.T(), err)

	buf := make([]byte, 32)
	size, err := striper.Read(name, buf, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 16, size)
	assert.Equal(suite.T(), "hello earthlings", string(buf[:size]))
}

func (suite *StriperTestSuite) TestReadWriteFull() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestReadWriteFull"
	err = striper.WriteFull(name, []byte("hello world"))
	require.NoError(suite.T(), err)

	err = striper.WriteFull(name, []byte("earthlings"))
	require.NoError(suite.T(), err)

	buf := make([]byte, 32)
	size, err := striper.Read(name, buf, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 10, size)
	assert.Equal(suite.T(), "earthlings", string(buf[:size]))
}

func (suite *StriperTestSuite) TestReadAppend() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestReadAppend"
	err = striper.Append(name, []byte("hello world"))
	require.NoError(suite.T(), err)

	err = striper.Append(name, []byte(" and all you earthlings"))
	require.NoError(suite.T(), err)

	buf := make([]byte, 64)
	size, err := striper.Read(name, buf, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 34, size)
	assert.Equal(suite.T(),
		"hello world and all you earthlings",
		string(buf[:size]))
}

func (suite *StriperTestSuite) TestRemove() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestRemove"
	err = striper.Append(name, []byte("goodbye world"))
	require.NoError(suite.T(), err)

	err = striper.Remove(name)
	require.NoError(suite.T(), err)

	buf := make([]byte, 32)
	_, err = striper.Read(name, buf, 0)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), -2, err.(radosStriperError).ErrorCode())
}

func (suite *StriperTestSuite) TestTruncate() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestTruncate"
	err = striper.Write(name, []byte("that's all folks"), 0)
	require.NoError(suite.T(), err)

	err = striper.Truncate(name, 10)
	require.NoError(suite.T(), err)

	buf := make([]byte, 32)
	size, err := striper.Read(name, buf, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 10, size)
	assert.Equal(suite.T(), "that's all", string(buf[:size]))
}

func (suite *StriperTestSuite) TestStat() {
	ut := time.Now().Unix()

	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestStat"
	err = striper.Write(name, []byte("that's all..."), 0)
	require.NoError(suite.T(), err)
	err = striper.Write(name, []byte("...folks"), 4096)
	require.NoError(suite.T(), err)

	ss, err := striper.Stat(name)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), uint64(4104), ss.Size)
	assert.Equal(suite.T(), uint64(4104), ss.Size)
	// the Seconds field of the timespec should be between the time we started
	// the test and two seconds after the time we started the test (as that
	// ought to be a pretty large window for this operation)
	assert.GreaterOrEqual(suite.T(), ss.ModTime.Sec, ut)
	assert.GreaterOrEqual(suite.T(), ut+2, ss.ModTime.Sec)
}

func (suite *StriperTestSuite) TestStatMissing() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	require.NoError(suite.T(), err)
	defer striper.Destroy()

	name := "TestStatMissing"
	_, err = striper.Stat(name)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), -2, err.(radosStriperError).ErrorCode())
}
