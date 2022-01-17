//go:build !nautilus
// +build !nautilus

package admin

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ceph/go-ceph/internal/admintest"
	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

var (
	defaultPoolName = "rbd"
	testImageSize   = uint64(1 << 22)
	testImageOrder  = 22
	alreadyExists   = -0x11

	radosConnector = admintest.NewConnector()
)

func getConn(t *testing.T) *rados.Conn {
	return radosConnector.GetConn(t)
}

func getAdmin(t *testing.T) *RBDAdmin {
	return NewFromConn(radosConnector.Get(t))
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
