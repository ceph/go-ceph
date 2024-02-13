//go:build !(nautilus || octopus || pacific) && ceph_preview
// +build !nautilus,!octopus,!pacific,ceph_preview

package admin

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ceph/go-ceph/rados"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func radosConnect(t *testing.T) *rados.Conn {
	conn, err := rados.NewConn()
	require.NoError(t, err)
	err = conn.ReadDefaultConfigFile()
	require.NoError(t, err)
	waitForRadosConn(t, conn)
	return conn
}

func waitForRadosConn(t *testing.T, conn *rados.Conn) {
	var err error
	timeout := time.After(time.Second * 15)
	ch := make(chan error)
	go func(conn *rados.Conn) {
		ch <- conn.Connect()
	}(conn)
	select {
	case err = <-ch:
	case <-timeout:
		err = fmt.Errorf("timed out waiting for connect")
	}
	require.NoError(t, err)
}

func (suite *RadosGWTestSuite) TestGetInfo() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	suite.T().Run("test get rgw cluster/endpoint information", func(_ *testing.T) {
		info, err := co.GetInfo(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "rados", info.InfoSpec.StorageBackends[0].Name)
		conn := radosConnect(suite.T())
		fsid, err := conn.GetFSID()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), fsid, info.InfoSpec.StorageBackends[0].ClusterID)
		conn.Shutdown()
	})
}
