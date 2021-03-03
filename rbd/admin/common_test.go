// +build !nautilus

package admin

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ccom "github.com/ceph/go-ceph/common/commands"
	"github.com/ceph/go-ceph/internal/commands"
	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

var (
	defaultPoolName = "rbd"
	testImageSize   = uint64(1 << 22)
	testImageOrder  = 22
	alreadyExists   = -0x11

	cachedRadosConn *rados.Conn
	cachedRBDAdmin  *RBDAdmin
	debugTrace      bool
)

func init() {
	dt := os.Getenv("GO_CEPH_TEST_DEBUG_TRACE")
	if ok, err := strconv.ParseBool(dt); ok && err == nil {
		debugTrace = true
	}
}

func getConn(t *testing.T) *rados.Conn {
	if cachedRadosConn != nil {
		return cachedRadosConn
	}

	conn, err := rados.NewConn()
	require.NoError(t, err)
	require.NotNil(t, conn)
	err = conn.ReadDefaultConfigFile()
	require.NoError(t, err)

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(conn *rados.Conn) {
		ch <- conn.Connect()
	}(conn)
	select {
	case err = <-ch:
	case <-timeout:
		err = errors.New("timed out waiting for connect")
	}
	require.NoError(t, err)

	cachedRadosConn = conn
	return cachedRadosConn
}

func getAdmin(t *testing.T) *RBDAdmin {
	if cachedRBDAdmin != nil {
		return cachedRBDAdmin
	}

	var c ccom.RadosCommander = getConn(t)
	if debugTrace {
		c = commands.NewTraceCommander(c)
	}
	cachedRBDAdmin := NewFromConn(c)
	require.NotNil(t, cachedRBDAdmin)
	// We sleep briefly before returning in order to ensure we have a mgr map
	// before we start executing the tests.
	time.Sleep(50 * time.Millisecond)
	return cachedRBDAdmin
}

func ensureDefaultPool(t *testing.T) {
	t.Helper()
	conn := getConn(t)
	err := conn.MakePool(defaultPoolName)
	if err == nil {
		t.Logf("created pool: %s", defaultPoolName)
		ioctx, err := conn.OpenIOContext(defaultPoolName)
		require.NoError(t, err)
		defer ioctx.Destroy()
		// initialize rbd for the pool. if this is not done all mirror
		// schedules can not be used on the pool or images in the pool
		err = rbd.PoolInit(ioctx, false)
		require.NoError(t, err)
		// enable per image mirroring for the new pool
		err = rbd.SetMirrorMode(ioctx, rbd.MirrorModeImage)
		require.NoError(t, err)
		return
	}
	ec, ok := err.(interface{ ErrorCode() int })
	if ok && ec.ErrorCode() == alreadyExists {
		t.Logf("pool already exists: %s", defaultPoolName)
	} else {
		t.Logf("failed to create pool: %s", defaultPoolName)
		t.FailNow()
	}
}
