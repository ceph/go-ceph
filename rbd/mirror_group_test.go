//go:build ceph_preview

package rbd

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupMirroring(t *testing.T) {
	mconfig := mirrorConfig()
	if mconfig == "" {
		t.Skip("no mirror config env var set")
	}

	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	// enable per-image mirroring for this pool
	err = SetMirrorMode(ioctx, MirrorModeImage)
	require.NoError(t, err)

	name := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	require.NoError(t, err)

	groupName := "group1"
	err = GroupCreate(ioctx, groupName)
	assert.NoError(t, err)

	err = GroupImageAdd(ioctx, groupName, ioctx, name)
	assert.NoError(t, err)

	token, err := CreateMirrorPeerBootstrapToken(ioctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(token), 4)

	conn2 := radosConnectConfig(t, mconfig)
	defer conn2.Shutdown()

	err = conn2.MakePool(poolName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn2.DeletePool(poolName))
	}()

	ioctx2, err := conn2.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx2.Destroy()
	}()

	err = SetMirrorMode(ioctx2, MirrorModeImage)
	require.NoError(t, err)

	err = ImportMirrorPeerBootstrapToken(
		ioctx2, MirrorPeerDirectionRxTx, token)
	assert.NoError(t, err)

	// enable mirroring
	err = MirrorGroupEnable(ioctx, groupName, ImageMirrorModeSnapshot)
	if errors.Is(err, ErrNotImplemented) {
		t.Skipf("MirrorGroupEnable is not supported: %v", err)
		return
	}
	assert.NoError(t, err)

	waitCounter := 30
	// wait for mirroring to be enabled
	for i := 0; i < waitCounter; i++ {
		resp, err := GetMirrorGroupInfo(ioctx, groupName)
		if errors.Is(err, ErrNotImplemented) {
			t.Skipf("GetMirrorGroupInfo is not supported: %v", err)
			return
		}
		assert.NoError(t, err)
		if resp.State.String() == "enabled" {
			break
		}
		if i == waitCounter-1 {
			assert.Fail(t, "mirror not enabled")
		}
		time.Sleep(2 * time.Second)
	}

	for i := 0; i < waitCounter; i++ {
		resp, err := GetGlobalMirrorGroupStatus(ioctx, groupName)
		if errors.Is(err, ErrNotImplemented) {
			t.Skipf("GetGlobalMirrorGroupStatus is not supported: %v", err)
			return
		}
		assert.NoError(t, err)
		if resp.SiteStatusesCount > 0 {
			break
		}
		if i == waitCounter-1 {
			assert.Fail(t, "site status count not updated in the mirror group global status")
		}
		time.Sleep(2 * time.Second)
	}

	// resync peer mirror group
	err = MirrorGroupResync(ioctx2, groupName)
	if errors.Is(err, ErrNotImplemented) {
		t.Skipf("MirrorGroupResync is not supported: %v", err)
		return
	}
	assert.NoError(t, err)

	// demote primary mirror group
	err = MirrorGroupDemote(ioctx, groupName)
	assert.NoError(t, err)

	// wait for peer mirror group to be Primary
	for i := 0; i < 30; i++ {
		resp, err := GetMirrorGroupInfo(ioctx2, groupName)
		if errors.Is(err, ErrNotImplemented) {
			t.Skipf("GetMirrorGroupInfo is not supported: %v", err)
			return
		}
		assert.NoError(t, err)
		if resp.Primary {
			break
		}
		if i == waitCounter-1 {
			assert.Fail(t, "mirror group on secondary site is not promoted")
		}
		time.Sleep(2 * time.Second)
	}

	// demote mirror group
	err = MirrorGroupPromote(ioctx, groupName, true)
	if errors.Is(err, ErrNotImplemented) {
		t.Skipf("MirrorGroupPromote is not supported: %v", err)
		return
	}
	assert.NoError(t, err)

	// wait for mirror group to be promoted
	for i := 0; i < waitCounter; i++ {
		resp, err := GetMirrorGroupInfo(ioctx, groupName)
		if errors.Is(err, ErrNotImplemented) {
			t.Skipf("GetMirrorGroupInfo is not supported: %v", err)
			return
		}
		assert.NoError(t, err)
		if resp.Primary {
			break
		}
		if i == waitCounter-1 {
			assert.Fail(t, "mirror group on Primary site is not promoted")
		}
		time.Sleep(2 * time.Second)
	}
}
