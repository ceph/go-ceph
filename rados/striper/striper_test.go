package striper

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tsuite "github.com/stretchr/testify/suite"

	"github.com/ceph/go-ceph/rados"
)

type StriperTestSuite struct {
	tsuite.Suite

	pool string
	conn *rados.Conn
}

func (suite *StriperTestSuite) connect() *rados.Conn {
	conn, err := rados.NewConn()
	require.NoError(suite.T(), err)

	err = conn.ReadDefaultConfigFile()
	require.NoError(suite.T(), err)

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(conn *rados.Conn) {
		ch <- conn.Connect()
	}(conn)
	select {
	case err = <-ch:
	case <-timeout:
		err = fmt.Errorf("timed out waiting for connect")
	}
	require.NoError(suite.T(), err)
	return conn
}

func (suite *StriperTestSuite) defaultContext() *rados.IOContext {
	ioctx, err := suite.conn.OpenIOContext(suite.pool)
	require.NoError(suite.T(), err)
	return ioctx
}

func (suite *StriperTestSuite) SetupSuite() {
	suite.pool = fmt.Sprintf("striper%s", uuid.Must(uuid.NewV4()).String())
	conn := suite.connect()
	defer conn.Shutdown()
	require.NoError(suite.T(), conn.MakePool(suite.pool))
}

func (suite *StriperTestSuite) TearDownSuite() {
	conn := suite.connect()
	defer conn.Shutdown()
	assert.NoError(suite.T(), conn.DeletePool(suite.pool))
}

func (suite *StriperTestSuite) SetupTest() {
	suite.conn = suite.connect()
}

func (suite *StriperTestSuite) TearDownTest() {
	suite.conn.Shutdown()
	suite.conn = nil
}

func (suite *StriperTestSuite) TestNewStriper() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	assert.NoError(suite.T(), err)
	striper.Destroy()
}

func (suite *StriperTestSuite) TestStriperSetObjectLayout() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	striper, err := New(ioctx)
	assert.NoError(suite.T(), err)
	defer striper.Destroy()

	assert.NoError(suite.T(), striper.SetObjectLayoutStripeUnit(65536))
	assert.NoError(suite.T(), striper.SetObjectLayoutStripeCount(16))
	assert.NoError(suite.T(), striper.SetObjectLayoutObjectSize(8388608))
}

func (suite *StriperTestSuite) TestNewStriperWithLayout() {
	ioctx := suite.defaultContext()
	defer ioctx.Destroy()

	l := Layout{65536, 16, 8388608}
	striper, err := NewWithLayout(ioctx, l)
	assert.NoError(suite.T(), err)
	striper.Destroy()
}

func TestStriperTestSuite(t *testing.T) {
	tsuite.Run(t, new(StriperTestSuite))
}
